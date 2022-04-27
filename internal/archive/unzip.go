package archive

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// Unzip 解压文件到相同目录下同名文件夹中
func Unzip(path string) error {
	rc, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	ext := filepath.Ext(path)
	dir := path[0 : len(path)-len(ext)] // 同名文件夹: /a/b/data.zip --> /a/b/data
	_ = os.RemoveAll(dir)
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	files := rc.File
	for _, file := range files {
		if err = extract(dir, file); err != nil {
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
