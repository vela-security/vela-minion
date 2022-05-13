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

func (vt *velaThird) saveDB() {
	thirds := vt.thirds()
	_ = vt.store.Store(vt.storeKey, thirds, 0)
}

func (vt *velaThird) loadDB() {
	data, err := vt.store.Get(vt.storeKey)
	if err != nil {
		return
	}

	thirds, ok := data.(model.VelaThirds)
	if !ok {
		return
	}

	hm := make(map[string]*model.VelaThird, 32)
	for _, third := range thirds {
		// 开始校验文件
		hash := vt.md5(filepath.Join(third.Path, third.Name))
		if hash == third.Hash {
			hm[third.ID] = third
		}
	}

	vt.mutex.Lock()
	defer vt.mutex.Unlock()
	vt.files = hm
}

func (*velaThird) md5(path string) string {
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

func (vt *velaThird) encodeFunc(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (vt *velaThird) decodeFunc(data []byte) (interface{}, error) {
	var res model.VelaThirds
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}
