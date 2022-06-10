package dispatch

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// thirdFile 3rd 文件
type thirdFile struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Name     string `json:"name"`
	Hash     string `json:"hash"`
	Filepath string `json:"filepath"`
	Extract  string `json:"extract"`
}

type thirdFiles []*thirdFile

// sumMD5 计算本地文件的 md5
func (f thirdFile) sumMD5() string {
	file, err := os.Open(f.Filepath)
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
