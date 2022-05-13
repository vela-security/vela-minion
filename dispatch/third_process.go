package dispatch

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/vela-security/vela-minion/model"
	"github.com/vela-security/vela-minion/tunnel"
)

func (vt *velaThird) process(cli *tunnel.Client, diffs []*actionThird) {
	for _, diff := range diffs {
		switch diff.Action {
		case taCreate: // 下载
			err := vt.create(cli, diff)
			vt.xEnv.Infof("下载三方文件: %s, 是否有错: %v", diff.Name, err)
		case taDelete:
			err := vt.remove(diff)
			vt.xEnv.Infof("删除三方文件: %s, 是否有错: %v", diff.Name, err)
		case taMove:
			err := vt.move(diff)
			vt.xEnv.Infof("移动三方文件: %s -> %s, 是否有错: %v", diff.LocalName, diff.Name, err)
		case taUpdate:
			err := vt.update(cli, diff)
			vt.xEnv.Infof("更新三方文件: %s, 是否有错: %v", diff.Name, err)
		}
	}
}

func (vt *velaThird) create(cli *tunnel.Client, diff *actionThird) error {
	_ = os.Remove(filepath.Join(diff.LocalPath, diff.LocalName))
	query := "id=" + diff.ID
	res := cli.HTTP(http.MethodGet, "/v1/third/sync", query, nil, nil)
	hash, err := res.SaveFile(filepath.Join(diff.Path, diff.Name))
	if err != nil {
		return err
	}

	vt.mutex.Lock()
	defer vt.mutex.Unlock()

	vt.files[diff.ID] = &model.VelaThird{ID: diff.ID, Path: diff.Path, Name: diff.Name, Hash: hash}

	return nil
}

func (vt *velaThird) update(cli *tunnel.Client, diff *actionThird) error {
	_ = os.Remove(filepath.Join(diff.LocalPath, diff.LocalName))
	query := "id=" + diff.ID
	res := cli.HTTP(http.MethodGet, "/v1/third/sync", query, nil, nil)
	hash, err := res.SaveFile(filepath.Join(diff.Path, diff.Name))
	if err != nil {
		return err
	}

	vt.mutex.Lock()
	defer vt.mutex.Unlock()

	vt.files[diff.ID] = &model.VelaThird{ID: diff.ID, Path: diff.Path, Name: diff.Name, Hash: hash}

	return nil
}

func (vt *velaThird) move(diff *actionThird) error {
	_ = os.MkdirAll(diff.Path, os.ModePerm)
	oldPath := filepath.Join(diff.LocalPath, diff.LocalName)
	_ = os.Remove(oldPath)
	newPath := filepath.Join(diff.Path, diff.Name)
	_ = os.Rename(oldPath, newPath)

	vt.mutex.Lock()
	defer vt.mutex.Unlock()

	vt.files[diff.ID] = &model.VelaThird{ID: diff.ID, Path: diff.Path, Name: diff.Name, Hash: diff.Hash}

	return nil
}

func (vt *velaThird) remove(diff *actionThird) error {
	vt.mutex.Lock()
	delete(vt.files, diff.ID)
	vt.mutex.Unlock()
	_ = os.Remove(filepath.Join(diff.LocalPath, diff.LocalName))

	return nil
}
