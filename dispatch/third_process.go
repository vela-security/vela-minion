package dispatch

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/vela-security/vela-minion/model"
	"github.com/vela-security/vela-minion/rockcli"
)

func (rt *rockThird) process(cli *rockcli.Client, diffs []*actionThird) {
	for _, diff := range diffs {
		switch diff.Action {
		case taCreate: // 下载
			err := rt.create(cli, diff)
			rt.xEnv.Infof("下载三方文件: %s, 是否有错: %v", diff.Name, err)
		case taDelete:
			err := rt.remove(diff)
			rt.xEnv.Infof("删除三方文件: %s, 是否有错: %v", diff.Name, err)
		case taMove:
			err := rt.move(diff)
			rt.xEnv.Infof("移动三方文件: %s -> %s, 是否有错: %v", diff.LocalName, diff.Name, err)
		case taUpdate:
			err := rt.update(cli, diff)
			rt.xEnv.Infof("更新三方文件: %s, 是否有错: %v", diff.Name, err)
		}
	}
}

func (rt *rockThird) create(cli *rockcli.Client, diff *actionThird) error {
	_ = os.Remove(filepath.Join(diff.LocalPath, diff.LocalName))
	query := "id=" + diff.ID
	res := cli.HTTP(http.MethodGet, "/v1/third/sync", query, nil, nil)
	hash, err := res.SaveFile(filepath.Join(diff.Path, diff.Name))
	if err != nil {
		return err
	}

	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	rt.files[diff.ID] = &model.RockThird{ID: diff.ID, Path: diff.Path, Name: diff.Name, Hash: hash}

	return nil
}

func (rt *rockThird) update(cli *rockcli.Client, diff *actionThird) error {
	_ = os.Remove(filepath.Join(diff.LocalPath, diff.LocalName))
	query := "id=" + diff.ID
	res := cli.HTTP(http.MethodGet, "/v1/third/sync", query, nil, nil)
	hash, err := res.SaveFile(filepath.Join(diff.Path, diff.Name))
	if err != nil {
		return err
	}

	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	rt.files[diff.ID] = &model.RockThird{ID: diff.ID, Path: diff.Path, Name: diff.Name, Hash: hash}

	return nil
}

func (rt *rockThird) move(diff *actionThird) error {
	_ = os.MkdirAll(diff.Path, os.ModePerm)
	oldPath := filepath.Join(diff.LocalPath, diff.LocalName)
	_ = os.Remove(oldPath)
	newPath := filepath.Join(diff.Path, diff.Name)
	_ = os.Rename(oldPath, newPath)

	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	rt.files[diff.ID] = &model.RockThird{ID: diff.ID, Path: diff.Path, Name: diff.Name, Hash: diff.Hash}

	return nil
}

func (rt *rockThird) remove(diff *actionThird) error {
	rt.mutex.Lock()
	delete(rt.files, diff.ID)
	rt.mutex.Unlock()
	_ = os.Remove(filepath.Join(diff.LocalPath, diff.LocalName))

	return nil
}
