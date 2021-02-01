package fs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Error Codes
var (
	ErrFileNotDir  = errors.New("a file is located at the expected directory location")
	ErrDirNotFile  = errors.New("a directory is located at the expected file location")
	ErrDirNotExist = errors.New("directory does not exist")
	ErrDirNotEmpty = errors.New("directory is not empty")
)

func FileExists(path ...string) bool {
	stat, err := os.Stat(filepath.Join(path...))
	return err == nil && !stat.IsDir()
}

func CheckFileExists(path ...string) (bool, error) {
	stat, err := os.Stat(filepath.Join(path...))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	} else if stat.IsDir() {
		return false, ErrDirNotFile
	}
	return true, nil
}

func DirExists(path ...string) bool {
	stat, err := os.Stat(filepath.Join(path...))
	return err == nil && stat.IsDir()
}

func CheckDirExists(path ...string) (bool, error) {
	stat, err := os.Stat(filepath.Join(path...))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	} else if !stat.IsDir() {
		return false, ErrFileNotDir
	}
	return true, nil
}

func ValidateFileExists(path ...string) error {
	exists, err := CheckFileExists(path...)
	if err != nil {
		return err
	}
	if !exists {
		return os.ErrNotExist
	}
	return nil
}

func ValidateDirExists(path ...string) error {
	exists, err := CheckDirExists(path...)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDirNotExist
	}
	return nil
}

func ValidateEmptyDirExists(path ...string) error {
	f, err := os.Open(filepath.Join(path...))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return nil
	}
	return ErrDirNotEmpty
}
