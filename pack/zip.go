package pack

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func ToFlatZip(filename string, files ...string) error {

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	zipper := zip.NewWriter(out)
	defer zipper.Close()

	for _, srcname := range files {
		file, err := os.Open(srcname)
		if err != nil {
			return err
		}
		defer file.Close()

		// Get the file information
		info, err := file.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.Base(srcname)
		header.Method = zip.Deflate

		writer, err := zipper.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}
	}
	return nil
}
