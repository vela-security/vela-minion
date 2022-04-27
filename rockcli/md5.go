package rockcli

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
)

type md5Writer struct {
	h hash.Hash
}

func (m md5Writer) Write(p []byte) (n int, err error) {
	return m.h.Write(p)
}

func (m md5Writer) Sum() string {
	sum := m.h.Sum(nil)
	return hex.EncodeToString(sum)
}

func newMD5Writer() *md5Writer {
	return &md5Writer{h: md5.New()}
}
