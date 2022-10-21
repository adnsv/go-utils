package filesystem

import (
	"bytes"
	"io"
	"os"
)

func skipData(w io.Reader, data []byte) (bool, error) {
	const cap = 1024
	var tmp [cap]byte
	rem := len(data)
	for rem > 0 {
		n := rem
		if n > cap {
			n = cap
		}
		t := tmp[0:n]
		got, err := w.Read(t)
		if err != nil || got != n || !bytes.Equal(t, data[:n]) {
			return false, err
		}
		data = data[n:]
		rem -= n
	}
	return true, nil
}

func matchData(f *os.File, data []byte) (bool, error) {
	info, err := f.Stat()
	if err != nil {
		return false, err
	}
	if info.Size() != int64(len(data)) {
		return false, nil
	}
	return skipData(f, data)
}

// FileContentMatch checks if the file has content that matches the specified
// data.
func FileContentMatch(fn string, data []byte) (bool, error) {
	f, err := os.Open(fn)
	if err != nil {
		return false, err
	}
	match, err := matchData(f, data)
	if e := f.Close(); err == nil {
		err = e
	}
	return match, err
}
