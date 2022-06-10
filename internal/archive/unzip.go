package archive

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// Unzip 解压文件
func Unzip(from, dest string) error {
	rc, err := zip.OpenReader(from)
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	files := rc.File
	for _, file := range files {
		if err = extract(dest, file); err != nil {
			break
		}
	}

	return err
}

// extract 将文件提取出来
func extract(dir string, file *zip.File) error {
	info := file.FileInfo()
	full := filepath.Join(dir, file.Name)
	if info.IsDir() {
		return os.MkdirAll(full, info.Mode())
	}
	_ = os.MkdirAll(filepath.Dir(full), 0644)

	df, err := os.OpenFile(full, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = df.Close() }()

	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	_, err = io.Copy(df, rc)

	return err
}
