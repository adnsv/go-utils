package unpack

import (
	"bytes"
	"compress/zlib"
	"io"
	"os"
)

// ZToFile unpacks zlib-compressed stream into a file
func ZToFile(src []byte, filename string, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := zlib.NewReader(bytes.NewReader(src))
	if err != nil {
		return err
	}
	defer r.Close()

	_, err = io.Copy(f, r)
	defer f.Close()

	return err
}
