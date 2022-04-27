package model

import (
	"github.com/vela-security/public/assert"
	"path/filepath"
	"strings"
	"time"
)

type RockThirds []*RockThird

type RockThird struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
	Hash string `json:"hash"`
}

func (t RockThird) Archived() bool {
	ext := filepath.Ext(t.Name)
	return strings.EqualFold(ext, ".zip")
}

// Map 将 slice 形式转为 map 形式，key-ID
func (t RockThirds) Map() map[string]*RockThird {
	hm := make(map[string]*RockThird, len(t))
	for _, third := range t {
		hm[third.ID] = third
	}
	return hm
}

type RockTasks []*assert.Task

// TaskRunner 内部运行状态
type TaskRunner struct {
	Name   string `json:"name"`   // 内部服务名字
	Type   string `json:"type"`   // 类型
	Status string `json:"status"` // 状态
}

// RockTask 配置运行
type RockTask struct {
	Name    string        `json:"name"`    // 配置名称
	Link    string        `json:"link"`    // 外链
	Status  string        `json:"status"`  // 运行状态
	Hash    string        `json:"hash"`    // hash
	From    string        `json:"from"`    // 来源
	Uptime  time.Time     `json:"uptime"`  // 启动时间
	Failed  bool          `json:"failed"`  // 是否失败
	Cause   string        `json:"cause"`   // 如果发生失败，失败的原因
	Runners []*TaskRunner `json:"runners"` // task 内部服务
}
