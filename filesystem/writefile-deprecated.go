package filesystem

import (
	"io"
	"io/fs"
	"os"
)

// WriteFileIfChanged writes data to a file.
//
//   - Does not overwrite if the file already has the specified content
//   - Uses 0666 permission if overwriting is neccessary
//
// Deprecated: Use WriteFile with WriteOptions instead
func WriteFileIfChanged(fn string, data []byte) (err error) {
	perm := fs.FileMode(0666)

	if !FileExists(fn) {
		return os.WriteFile(fn, data, perm)
	}

	var f *os.File
	f, err = os.OpenFile(fn, os.O_RDWR, perm)
	if err != nil {
		return
	}
	defer func() {
		if e := f.Close(); err == nil {
			err = e
		}
	}()

	match, err := matchData(f, data)
	if err != nil || match {
		return
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return
	}
	_, err = f.Write(data)
	if err != nil {
		return
	}
	err = f.Truncate(int64(len(data)))
	return
}
