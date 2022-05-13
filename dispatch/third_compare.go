package dispatch

import (
	"path/filepath"

	"github.com/vela-security/vela-minion/model"
)

type thirdAction int8

const (
	taCreate thirdAction = iota + 1
	taMove
	taDelete
	taUpdate
)

type actionThird struct {
	model.VelaThird
	Action                          thirdAction
	LocalHash, LocalPath, LocalName string
}

// compare 比较差异
func (vt *velaThird) compare(recs model.VelaThirds) []*actionThird {

	res := make([]*actionThird, 0, 16)
	locals := vt.thirds().Map() // 获取当前已经加载好的三方文件

	for _, rec := range recs {
		id := rec.ID
		local := locals[id]

		mrt := model.VelaThird{ID: rec.ID, Path: rec.Path, Name: rec.Name, Hash: rec.Hash}
		at := &actionThird{VelaThird: mrt}

		if local == nil {
			at.Action = taCreate
			res = append(res, at)
			continue
		}

		delete(locals, id)

		at.LocalPath, at.LocalName, at.LocalHash = local.Path, local.Name, local.Hash
		recPath := filepath.Join(rec.Path, rec.Name)
		localPath := filepath.Join(local.Path, local.Name)
		if recPath == localPath && rec.Hash == local.Hash { // hash 与 路径都一样，说明没有任何修改不做处理
			continue
		}
		if rec.Hash == local.Hash && recPath != localPath {
			at.Action = taMove
			res = append(res, at)
		} else {
			at.Action = taUpdate
			res = append(res, at)
		}
	}

	for _, local := range locals {
		at := &actionThird{VelaThird: model.VelaThird{ID: local.ID},
			Action: taDelete, LocalHash: local.Hash, LocalPath: local.Path, LocalName: local.Name}
		res = append(res, at)
	}

	return res
}
