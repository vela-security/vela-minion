package dispatch

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/vela-security/vela-minion/model"
)

func (rt *rockThird) saveDB() {
	thirds := rt.thirds()
	_ = rt.store.Store(rt.storeKey, thirds, 0)
}

func (rt *rockThird) loadDB() {
	data, err := rt.store.Get(rt.storeKey)
	if err != nil {
		return
	}

	thirds, ok := data.(model.RockThirds)
	if !ok {
		return
	}

	hm := make(map[string]*model.RockThird, 32)
	for _, third := range thirds {
		// 开始校验文件
		hash := rt.md5(filepath.Join(third.Path, third.Name))
		if hash == third.Hash {
			hm[third.ID] = third
		}
	}

	rt.mutex.Lock()
	defer rt.mutex.Unlock()
	rt.files = hm
}

func (*rockThird) md5(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	h := md5.New()
	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return ""
		}
		h.Write(buf[:n])
	}

	sum := h.Sum(nil)

	return hex.EncodeToString(sum)
}

func (rt *rockThird) encodeFunc(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (rt *rockThird) decodeFunc(data []byte) (any, error) {
	var res model.RockThirds
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}
