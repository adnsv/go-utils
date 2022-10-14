package filesystem

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Error Codes
var (
	ErrFileNotDir        = errors.New("a file is located at the expected directory location")
	ErrDirNotFile        = errors.New("a directory is located at the expected file location")
	ErrFileDoesNotExist  = os.ErrNotExist
	ErrDirDoesNotExist   = errors.New("directory does not exist")
	ErrDirIsNotEmpty     = errors.New("directory is not empty")
	ErrPathDoesNotExist  = errors.New("path does not exist")
	ErrFileExists        = os.ErrExist
	ErrDirExists         = errors.New("directory already exists")
	ErrPathIsNotAbsolute = errors.New("path is not absolute")
	ErrPathIsAbsolute    = errors.New("path is absolute")
)

// FileExists returns true if a file exists at the specified location.
// If the path points to a directory, this function returns false.
func FileExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}

// DirExists returns true if a directory exists at the specified location.
// If the path points to a file, this function returns false.
func DirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

func CheckFileExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	} else if stat.IsDir() {
		return false, ErrDirNotFile
	}
	return true, nil
}

func CheckDirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	} else if !stat.IsDir() {
		return false, ErrFileNotDir
	}
	return true, nil
}

func ValidateFileExists(path string) error {
	exists, err := CheckFileExists(path)
	if err != nil {
		return err
	}
	if !exists {
		return os.ErrNotExist
	}
	return nil
}

func ValidateNoFileExists(path string) error {
	exists, err := CheckFileExists(path)
	if err != nil {
		return err
	}
	if exists {
		return ErrFileExists
	}
	return nil
}

func ValidateDirExists(path string) error {
	exists, err := CheckDirExists(path)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDirDoesNotExist
	}
	return nil
}

func ValidateNoDirExists(path string) error {
	exists, err := CheckDirExists(path)
	if err != nil {
		return err
	}
	if exists {
		return ErrDirExists
	}
	return nil
}

func ValidateEmptyDirExists(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return nil
	}
	return ErrDirIsNotEmpty
}

func ValidatePathExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrPathDoesNotExist
		}
	}
	return err
}

func ValidateNoPathExists(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if stat.IsDir() {
		return ErrDirExists
	}
	return ErrFileExists
}

func ValidatePathIsAbsolute(path string) error {
	if filepath.IsAbs(path) {
		return nil
	}
	return ErrPathIsNotAbsolute
}

func ValidatePathIsNotAbsolute(path string) error {
	if !filepath.IsAbs(path) {
		return nil
	}
	return ErrPathIsAbsolute
}

// ResolvesToSameFile returns true if the two paths resolve to the
// same actual file. Follows symlinks.
func ResolvesToSameFile(pathA, pathB string) bool {
	if pathA == pathB {
		return true
	}
	sa, err := os.Stat(pathA)
	if err != nil {
		return false
	}
	if sa.Mode()&os.ModeSymlink == os.ModeSymlink {
		pathA, err = filepath.EvalSymlinks(pathA)
		if err != nil {
			return false
		}
		sa, err = os.Stat(pathA)
		if err != nil {
			return false
		}
	}

	sb, err := os.Stat(pathB)
	if err != nil {
		return false
	}
	if sb.Mode()&os.ModeSymlink == os.ModeSymlink {
		pathB, err = filepath.EvalSymlinks(pathB)
		if err != nil {
			return false
		}
		sb, err = os.Stat(pathB)
		if err != nil {
			return false
		}
	}

	return os.SameFile(sa, sb)
}

// NormalizePathsToSlash normalizes a list of file paths:
// - removes empty paths
// - converts separators to slashes
// - removes duplicates
// - sorts lexicographically
func NormalizePathsToSlash(paths []string) []string {
	m := make(map[string]struct{}, len(paths))
	for _, s := range paths {
		if s == "" {
			continue
		}
		m[filepath.ToSlash(s)] = struct{}{}
	}
	ret := make([]string, 0, len(m))
	for f := range m {
		ret = append(ret, f)
	}
	return sort.StringSlice(ret)
}
