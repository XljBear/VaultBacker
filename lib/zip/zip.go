package zip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Zip(srcDir string, zipFileName string) (err error) {

	_ = os.RemoveAll(zipFileName)

	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return
	}
	defer func() { _ = zipFile.Close() }()

	archive := zip.NewWriter(zipFile)
	defer func() { _ = archive.Close() }()

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, _ error) error {

		if path == srcDir {
			return nil
		}

		header, _ := zip.FileInfoHeader(info)

		header.Name = strings.TrimPrefix(path, srcDir+`/`)

		if info.IsDir() {
			header.Name += `/`
		} else {
			header.Method = zip.Deflate
		}

		writer, _ := archive.CreateHeader(header)
		if !info.IsDir() {
			file, _ := os.Open(path)
			defer func() { _ = file.Close() }()
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}
