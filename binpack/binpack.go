package binpack

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
)

type Source struct {
	Name       string
	Content    []byte
	Compressed bool
}

func FromFile(fn string, name string, compress bool) (*Source, error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
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
	}
	return &Source{Name: name, Content: data, Compressed: compress}, nil
}
