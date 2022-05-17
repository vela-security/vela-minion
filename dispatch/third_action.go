package dispatch

import (
	"net/http"
	"os"

	"github.com/vela-security/vela-minion/internal/archive"
	"github.com/vela-security/vela-minion/tunnel"
)

type thirdAction int8

const (
	taCreate thirdAction = iota + 1
	taMove
	taDelete
	taUpdate
)

type thirdDiffs []*thirdDiff

type thirdDiff struct {
	Action      thirdAction // 文件比对后要执行的动作
	ID          string      // 文件 ID
	Hash        string      // 文件 hash
	NewPath     string      // 新文件路径
	NewName     string      // 新文件名字
	NewFilepath string      // 新文件路径 + 文件名
	NewExtract  string      // 新文件的解压路径, 不是可解压文件值为空
	OldFilepath string      // 旧文件路径 + 文件名
	OldExtract  string      // 旧文件的解压路径, 不是可解压文件值为空
}

func (td thirdDiff) create(cli *tunnel.Client) (*thirdFile, error) {
	// 删除旧文件与旧目录
	_ = os.Remove(td.OldFilepath)
	if td.OldExtract != "" {
		_ = os.RemoveAll(td.OldExtract)
	}

	query := "id=" + td.ID
	res := cli.HTTP(http.MethodGet, "/v1/third/sync", query, nil, nil)
	hash, err := res.SaveFile(td.NewFilepath)
	if err != nil {
		return nil, err
	}
	if td.NewExtract != "" {
		if err = archive.Unzip(td.NewFilepath, td.NewExtract); err != nil {
			return nil, err
		}
	}

	tf := &thirdFile{ID: td.ID, Path: td.NewPath, Name: td.NewName, Hash: hash,
		Filepath: td.NewFilepath, Extract: td.NewExtract}

	return tf, nil
}

func (td thirdDiff) move() (*thirdFile, error) {
	if err := os.Rename(td.OldFilepath, td.NewFilepath); err != nil {
		return nil, err
	}
	if td.OldExtract != "" && td.NewExtract != "" {
		if err := os.Rename(td.OldExtract, td.NewExtract); err != nil {
			return nil, err
		}
	}

	tf := &thirdFile{ID: td.ID, Path: td.NewPath, Name: td.NewName, Hash: td.Hash,
		Filepath: td.NewFilepath, Extract: td.OldExtract}

	return tf, nil
}

func (td thirdDiff) update(cli *tunnel.Client) (*thirdFile, error) {
	// 删除旧文件与旧目录
	_ = os.Remove(td.OldFilepath)
	if td.OldExtract != "" {
		_ = os.RemoveAll(td.OldExtract)
	}

	query := "id=" + td.ID
	res := cli.HTTP(http.MethodGet, "/v1/third/sync", query, nil, nil)
	hash, err := res.SaveFile(td.NewFilepath)
	if err != nil {
		return nil, err
	}
	if td.NewExtract != "" {
		if err = archive.Unzip(td.NewFilepath, td.NewExtract); err != nil {
			return nil, err
		}
	}

	tf := &thirdFile{ID: td.ID, Path: td.NewPath, Name: td.NewName, Hash: hash,
		Filepath: td.NewFilepath, Extract: td.NewExtract}

	return tf, nil
}

func (td thirdDiff) delete() error {
	if err := os.Remove(td.OldFilepath); err != nil || td.OldExtract == "" {
		return err
	}

	return os.RemoveAll(td.OldExtract)
}
