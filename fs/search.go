package fs

import (
	"os"
	"path/filepath"
	"regexp"
)

// type aliases to make the code more self-descriptive
type filename = string
type dirname = string
type symlink = string

// SearchFilesInDir returns the list of file names within the specified
// directory that matches the regular expression. Returned are the basenames
// relative to the specified dir. This is a non-recursive search.
func SearchFilesInDir(dir dirname, re *regexp.Regexp) []filename {
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
		if !fi.IsDir() && re.MatchString(fi.Name()) {
			ret = append(ret, fi.Name())
		}
	}
	return ret
}

// SearchSubdirsInDir returns the list of subdir names within the specified
// directory that matches the regular expression. Returned are the basenames
// relative to the specified dir. This is a non-recursive search.
func SearchSubdirsInDir(dir dirname, re *regexp.Regexp) []dirname {
	d, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer d.Close()
	finfos, err := d.Readdir(-1)
	if err != nil {
		return nil
	}

	ret := []dirname{}
	for _, fi := range finfos {
		if fi.IsDir() && re.MatchString(fi.Name()) {
			ret = append(ret, fi.Name())
		}
	}
	return ret
}

// SearchFilesAndSymlinks scans the provided set of directories and returns
// absolute filenames that match the search criteria specified within a regular
// expression. While searching, symlinks are resolved. Returns a map of
// real absolute file paths and symlinks pointing to those files.
func SearchFilesAndSymlinks(dirs []string, re *regexp.Regexp) map[filename][]symlink {
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
			fn := fi.Name()
			if !fi.IsDir() && re.MatchString(fn) {
				path, err := filepath.Abs(filepath.Join(dir, fn))
				if err == nil {
					if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
						real, err := filepath.EvalSymlinks(path)
						if err == nil {
							real, err = filepath.Abs(real)
							if err == nil {
								ll := files[real]
								files[real] = appendIfUnique(ll, path)
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
