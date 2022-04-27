package dispatch

import (
	"github.com/vela-security/public/assert"
	"sync"
	"time"

	"github.com/vela-security/vela-minion/model"
	"github.com/vela-security/vela-minion/rockcli"
)

type normalReq struct {
	Data any `json:"data"`
}

type rockThird struct {
	xEnv     assert.Environment
	store    assert.Bucket
	storeKey string
	mutex    sync.RWMutex
	files    map[string]*model.RockThird
}

func newRockThird(env assert.Environment) *rockThird {
	storeKey := "third"
	store := env.Bucket(storeKey)

	files := make(map[string]*model.RockThird, 32)
	rt := &rockThird{xEnv: env, storeKey: storeKey, store: store, files: files}

	env.Mime(model.RockThirds{}, rt.encodeFunc, rt.decodeFunc)
	rt.loadDB()

	return rt
}

func (rt *rockThird) sync(cli *rockcli.Client) {
	var retry int
	var success bool
	for !success && retry < 5 {
		retry++

		thirds, err := rt.postThirds(cli)
		if err != nil {
			rt.xEnv.Errorf("上报三方文件错误: %v", err)
			time.Sleep(time.Second)
			continue
		}

		diffs := rt.compare(thirds)
		if len(diffs) == 0 {
			success = true
			break
		}

		rt.xEnv.Infof("正在处理三方文件差异")
		rt.process(cli, diffs)

		time.Sleep(time.Second)
	}

	rt.saveDB()
}

func (rt *rockThird) thirds() model.RockThirds {
	rt.mutex.RLock()
	defer rt.mutex.RUnlock()

	ret := make(model.RockThirds, 0, len(rt.files))
	for _, f := range rt.files {
		ret = append(ret, &model.RockThird{ID: f.ID, Path: f.Path, Name: f.Name, Hash: f.Hash})
	}

	return ret
}

func (rt *rockThird) postThirds(cli *rockcli.Client) (model.RockThirds, error) {
	data := rt.thirds()
	req := &dataReq{Data: data}

	var res struct {
		Data model.RockThirds `json:"data"`
	}
	if err := cli.PostJSON("/v1/third/sync", req, &res); err != nil {
		return nil, err
	}

	return res.Data, nil
}
