package dispatch

import (
	"github.com/vela-security/vela-public/assert"
	"github.com/vela-security/vela-minion/tunnel"
	"sync"
	"time"

	"github.com/vela-security/vela-minion/model"
)

type normalReq struct {
	Data any `json:"data"`
}

type velaThird struct {
	xEnv     assert.Environment
	store    assert.Bucket
	storeKey string
	mutex    sync.RWMutex
	files    map[string]*model.VelaThird
}

func newRockThird(env assert.Environment) *velaThird {
	storeKey := "third"
	store := env.Bucket(storeKey)

	files := make(map[string]*model.VelaThird, 32)
	vt := &velaThird{xEnv: env, storeKey: storeKey, store: store, files: files}

	env.Mime(model.VelaThirds{}, vt.encodeFunc, vt.decodeFunc)
	vt.loadDB()

	return vt
}

func (vt *velaThird) sync(cli *tunnel.Client) {
	var retry int
	var success bool
	for !success && retry < 5 {
		retry++

		thirds, err := vt.postThirds(cli)
		if err != nil {
			vt.xEnv.Errorf("上报三方文件错误: %v", err)
			time.Sleep(time.Second)
			continue
		}

		diffs := vt.compare(thirds)
		if len(diffs) == 0 {
			success = true
			break
		}

		vt.xEnv.Infof("正在处理三方文件差异")
		vt.process(cli, diffs)

		time.Sleep(time.Second)
	}

	vt.saveDB()
}

func (vt *velaThird) thirds() model.VelaThirds {
	vt.mutex.RLock()
	defer vt.mutex.RUnlock()

	ret := make(model.VelaThirds, 0, len(vt.files))
	for _, f := range vt.files {
		ret = append(ret, &model.VelaThird{ID: f.ID, Path: f.Path, Name: f.Name, Hash: f.Hash})
	}

	return ret
}

func (vt *velaThird) postThirds(cli *tunnel.Client) (model.VelaThirds, error) {
	data := vt.thirds()
	req := &dataReq{Data: data}

	var res struct {
		Data model.VelaThirds `json:"data"`
	}
	if err := cli.PostJSON("/v1/third/sync", req, &res); err != nil {
		return nil, err
	}

	return res.Data, nil
}
