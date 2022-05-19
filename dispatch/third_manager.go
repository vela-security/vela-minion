package dispatch

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vela-security/vela-minion/model"
	"github.com/vela-security/vela-minion/tunnel"
	"github.com/vela-security/vela-public/assert"
)

type thirdManager struct {
	env    assert.Environment
	bkt    assert.Bucket
	bktKey string
	mutex  sync.RWMutex
	files  map[string]*thirdFile
}

func newThirdManager(env assert.Environment) *thirdManager {
	bktKey := "third"
	bkt := env.Bucket(bktKey)
	files := make(map[string]*thirdFile, 8)
	tm := &thirdManager{env: env, bktKey: bktKey, bkt: bkt, files: files}
	env.Mime(thirdFiles{}, tm.encodeFunc, tm.decodeFunc)

	tm.readBucket()

	return tm
}

func (tm *thirdManager) sync(cli *tunnel.Client) {
	var retry int
	var success bool
	for !success && retry < 5 {
		retry++

		thirds, err := tm.postThirds(cli)
		if err != nil {
			tm.env.Errorf("上报三方文件错误: %v", err)
			time.Sleep(time.Second)
			continue
		}

		diffs := tm.compare(thirds)
		if len(diffs) == 0 {
			success = true
			break
		}

		tm.env.Infof("正在处理三方文件差异")
		tm.process(cli, diffs)

		time.Sleep(time.Second)
	}

	// 同步完成后将最新的 3rd 信息持久化，
	tm.saveBucket()
}

// postThirds
func (tm *thirdManager) postThirds(cli *tunnel.Client) (model.VelaThirds, error) {
	data := tm.velaThirds()
	req := &struct {
		Data interface{} `json:"data"`
	}{
		Data: data,
	}

	var res struct {
		Data model.VelaThirds `json:"data"`
	}
	if err := cli.PostJSON("/v1/third/sync", req, &res); err != nil {
		return nil, err
	}

	return res.Data, nil
}

func (tm *thirdManager) velaThirds() model.VelaThirds {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	ret := make(model.VelaThirds, 0, len(tm.files))
	for _, file := range tm.files {
		vt := &model.VelaThird{ID: file.ID, Path: file.Path, Name: file.Name, Hash: file.Hash}
		ret = append(ret, vt)
	}

	return ret
}

func (tm *thirdManager) thirdMap() map[string]*thirdFile {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	hm := make(map[string]*thirdFile, len(tm.files))
	for _, file := range tm.files {
		tf := &thirdFile{
			ID:       file.ID,
			Path:     file.Path,
			Name:     file.Name,
			Hash:     file.Hash,
			Filepath: file.Filepath,
			Extract:  file.Extract,
		}
		hm[file.ID] = tf
	}

	return hm
}

func (tm *thirdManager) saveBucket() {
	files := tm.thirdMap()
	ret := make(thirdFiles, 0, len(files))
	for _, f := range files {
		ret = append(ret, f)
	}

	if err := tm.bkt.Store(tm.bktKey, ret, 0); err != nil {
		tm.env.Warnf("读取bucket中的3rd错误: %v", err)
	}
}

// readBucket 读取并校验 bucket 中存储的 3rd 文件信息
func (tm *thirdManager) readBucket() {
	dat, err := tm.bkt.Get(tm.bktKey)
	if err != nil {
		tm.env.Warnf("读取bucket的third信息错误: %v", err)
		return
	}

	thirds, ok := dat.(thirdFiles)
	if !ok {
		return
	}

	hm := make(map[string]*thirdFile, 32)
	for _, third := range thirds {
		hash := third.sumMD5()
		if hash == third.Hash {
			hm[third.ID] = third
		}
	}

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.files = hm
}

func (*thirdManager) encodeFunc(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (*thirdManager) decodeFunc(data []byte) (interface{}, error) {
	var res thirdFiles
	err := json.Unmarshal(data, &res)
	return res, err
}

func (tm *thirdManager) compare(recs model.VelaThirds) thirdDiffs {
	hm := tm.thirdMap()
	ret := make(thirdDiffs, 0, 16)

	for _, rec := range recs {
		id := rec.ID
		file := hm[id]

		newFilePath := filepath.Join(rec.Path, rec.Name)
		td := &thirdDiff{
			ID:          rec.ID,
			Hash:        rec.Hash,
			NewPath:     rec.Path,
			NewName:     rec.Name,
			NewFilepath: newFilePath,
			NewExtract:  tm.ExtractPath(newFilePath),
		}
		// 本地不存在为新增
		if file == nil {
			td.Action = taCreate
			ret = append(ret, td)
			continue
		}

		delete(hm, id)
		td.OldFilepath, td.OldExtract = file.Filepath, file.Extract

		if newFilePath == file.Filepath && rec.Hash == file.Hash { // hash 与 路径都一样，说明没有任何修改不做处理
			continue
		}

		if rec.Hash == file.Hash && newFilePath == file.Filepath {
			td.Action = taMove
		} else {
			td.Action = taUpdate
		}

		ret = append(ret, td)
	}

	for _, file := range hm {
		td := &thirdDiff{
			Action:      taDelete,
			ID:          file.ID,
			Hash:        file.Hash,
			OldFilepath: file.Filepath,
			OldExtract:  file.Extract,
		}
		ret = append(ret, td)
	}

	return ret
}

func (*thirdManager) ExtractPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".zip" {
		return ""
	}

	dir := path[0 : len(path)-len(ext)] // 同名文件夹: /a/b/data.zip --> /a/b/data

	return dir
}

func (tm *thirdManager) process(cli *tunnel.Client, diffs thirdDiffs) {
	for _, diff := range diffs {
		switch diff.Action {
		case taCreate:
			tm.env.Infof("3rd 新增: %s", diff.NewFilepath)
			if file, err := diff.create(cli); err != nil {
				tm.env.Warnf("3rd 创建错误: %s, %v", diff.NewFilepath, err)
			} else {
				tm.mutex.Lock()
				tm.files[diff.ID] = file
				tm.mutex.Unlock()
			}
		case taMove:
			tm.env.Infof("3rd 移动: %s -> %s", diff.OldFilepath, diff.NewFilepath)
			if file, err := diff.move(); err != nil {
				tm.env.Warnf("3rd 移动错误: %s -> %s, %v", diff.OldFilepath, diff.NewFilepath, err)
			} else {
				tm.mutex.Lock()
				tm.files[diff.ID] = file
				tm.mutex.Unlock()
			}
		case taUpdate:
			tm.env.Infof("3rd 更新: %s -> %s", diff.OldFilepath, diff.NewFilepath)
			if file, err := diff.update(cli); err != nil {
				tm.env.Warnf("3rd 更新错误: %s -> %s, %v", diff.OldFilepath, diff.NewFilepath, err)
			} else {
				tm.mutex.Lock()
				tm.files[diff.ID] = file
				tm.mutex.Unlock()
			}
		case taDelete:
			tm.env.Infof("3rd 删除: %s", diff.OldFilepath)
			if err := diff.delete(); err != nil {
				tm.env.Warnf("3rd 删除错误: %s, %v", diff.OldFilepath, err)
			}
			tm.mutex.Lock()
			delete(tm.files, diff.ID)
			tm.mutex.Unlock()
		}
	}
}
