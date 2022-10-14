package binpack

import (
	"bytes"
	"compress/zlib"
	"os"
)

type Source struct {
	Name         string
	Content      []byte
	Compressed   bool
	OriginalSize int
}

func FromFile(fn string, name string, compress bool) (*Source, error) {
	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	ret := &Source{Name: name, OriginalSize: len(data)}
	if compress {
		b := bytes.Buffer{}
		w, err := zlib.NewWriterLevel(&b, zlib.BestCompression)
		if err != nil {
			return nil, err
		}
		_, err = w.Write(data)
		w.Close()
		if err != nil {
			return nil, err
		}
		data = b.Bytes()
		ret.Compressed = true
	}
	ret.Content = data
	return ret, nil
}
