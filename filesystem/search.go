package filesystem

import (
	"os"
	"path/filepath"
)

// type aliases to make the code more self-descriptive
type filename = string
type dirname = string
type symlink = string

// SearchDir returns the list of paths within the specified directory that pass
// through the 'accept' callback. This is a non-recursive search.
func SearchDir(dir dirname, accept func(os.FileInfo) bool) []filename {
	d, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer d.Close()
	finfos, err := d.Readdir(-1)
	if err != nil {
		return nil
	}

	ret := []filename{}
	for _, fi := range finfos {
		if accept(fi) {
			ret = append(ret, fi.Name())
		}
	}
	return ret
}

// SearchFilesAndSymlinks scans the provided set of directories and returns
// absolute filenames that pass through a functional 'accept' filter.
//
// While searching, symlinks are resolved. For symlinks, the 'accept' is called
// twice: first on a symlink itself, then on its target.
//
// Returns a map of real absolute file paths and symlinks pointing to those
// files.
//
func SearchFilesAndSymlinks(dirs []string, accept func(os.FileInfo) bool) map[filename][]symlink {
	files := map[filename][]symlink{}
	for _, dir := range dirs {
		d, err := os.Open(dir)
		if err != nil {
			continue
		}
		defer d.Close()
		finfos, err := d.Readdir(-1)
		if err != nil {
			continue
		}

		for _, fi := range finfos {
			if !fi.IsDir() && accept(fi) {
				fn := fi.Name()
				path, err := filepath.Abs(filepath.Join(dir, fn))
				if err == nil {
					if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
						real, err := filepath.EvalSymlinks(path)
						if err == nil {
							rfi, err := os.Stat(real)
							if err == nil && !rfi.IsDir() && accept(rfi) {
								real, err = filepath.Abs(real)
								if err == nil {
									ll := files[real]
									files[real] = appendIfUnique(ll, path)
								}
							}
						}
					} else {
						_, ok := files[path]
						if !ok {
							files[path] = []string{}
						}
					}
				}
			}
		}
	}
	return files
}

func appendIfUnique(ss []string, s string) []string {
	for _, it := range ss {
		if it == s {
			return ss
		}
	}
	return append(ss, s)
}
