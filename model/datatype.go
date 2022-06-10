package model

import (
	"github.com/vela-security/vela-public/assert"
	"path/filepath"
	"strings"
)

type VelaThirds []*VelaThird
type VelaTasks []*assert.Task

type VelaThird struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
	Hash string `json:"hash"`
}

func (t VelaThird) Archived() bool {
	ext := filepath.Ext(t.Name)
	return strings.EqualFold(ext, ".zip")
}

// Map 将 slice 形式转为 map 形式，key-ID
func (t VelaThirds) Map() map[string]*VelaThird {
	hm := make(map[string]*VelaThird, len(t))
	for _, third := range t {
		hm[third.ID] = third
	}
	return hm
}
