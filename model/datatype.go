package model

import (
	"github.com/vela-security/public/assert"
	"path/filepath"
	"strings"
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
