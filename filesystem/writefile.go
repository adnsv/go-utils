package filesystem

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
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
// data
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

type WriteFeedback int

const (
	FeedbackWriteBegin = WriteFeedback(iota)
	FeedbackWriteSucceded
	FeedbackWriteFailed
	FeedbackWriteSkipped

	FeedbackBackupBegin
	FeedbackBackupSucceded
	FeedbackBackupFailed
	FeedbackBackupRestoreFailed
)

type WriteFeedbackProc = func(fb WriteFeedback, fn string)
type BackupNameGenerator = func(original_fn string, n int) string

// WriteOptions provides detailed configuration
type WriteOptions struct {
	Perm                     os.FileMode         // file writing permissions, defaults to 0666 if unspecidied (perm == 0)
	OverwriteMatchingContent bool                // backup and overwrite, even if content matches
	Backup                   BackupNameGenerator // backup filename generator, no backup by default
	OnFeedback               WriteFeedbackProc   // use this if logging or user feedback is required
}

var errFileBackupFailed = errors.New("file backup failed")

// WriteFile writes data to the named file with configurable behavior and
// feedback
func WriteFile(fn string, buf []byte, opts *WriteOptions) error {

	if opts == nil {
		return os.WriteFile(fn, buf, 0666)
	}

	// effective permissions
	perm := opts.Perm
	if perm == 0 {
		perm = 0666
	}

	perform_write := func() error {
		if opts.OnFeedback != nil {
			opts.OnFeedback(FeedbackWriteBegin, fn)
		}
		err := os.WriteFile(fn, buf, perm)
		if opts.OnFeedback != nil {
			if err == nil {
				opts.OnFeedback(FeedbackWriteSucceded, fn)
			} else {
				opts.OnFeedback(FeedbackWriteFailed, fn)
			}
		}
		return err
	}

	if !opts.OverwriteMatchingContent {
		match, err := FileContentMatch(fn, buf)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return perform_write()
			} else {
				return err
			}
		}
		if match {
			if opts.OnFeedback != nil {
				opts.OnFeedback(FeedbackWriteSkipped, fn)
			}
			return nil
		}
	}

	if opts.Backup == nil {
		return perform_write()
	}

	backup_attempt := 1
	backup_fn := opts.Backup(fn, backup_attempt)
	if backup_fn == "" {
		return errFileBackupFailed
	}

	for FileExists(backup_fn) {
		backup_attempt++
		backup_fn = opts.Backup(fn, backup_attempt)
		if backup_fn == "" {
			return errFileBackupFailed
		}
	}

	if opts.OnFeedback != nil {
		opts.OnFeedback(FeedbackBackupBegin, backup_fn)
	}
	if err := os.Rename(fn, backup_fn); err != nil {
		if opts.OnFeedback != nil {
			opts.OnFeedback(FeedbackBackupFailed, backup_fn)
		}
		return err
	}
	if opts.OnFeedback != nil {
		opts.OnFeedback(FeedbackBackupSucceded, backup_fn)
	}

	err := perform_write()
	if err != nil {
		// try to restore backup
		restore_err := os.Rename(backup_fn, fn)
		if restore_err != nil && opts.OnFeedback != nil {
			opts.OnFeedback(FeedbackBackupRestoreFailed, backup_fn)
		}
		return err
	}

	return nil
}

// BackupNameNumeric produces a backup name generator that injects a numbered
// suffix into the original filename
//
//   - first backup: file<suffix>.ext
//   - n-th backup: file<suffix>_n.ext
//   - stops after max_attempts if generated filenames are already claimed by existing files
func BackupNameNumeric(suffix string, max_attempts int) BackupNameGenerator {
	return func(original_fn string, n int) string {
		if n > max_attempts {
			return ""
		}
		ext := filepath.Ext(original_fn)
		without_ext := original_fn[:len(original_fn)-len(ext)]
		if n <= 1 {
			return fmt.Sprintf("%s%s%s", without_ext, suffix, ext)
		} else {
			return fmt.Sprintf("%s%s_%d%s", without_ext, suffix, n, ext)
		}
	}
}

// BackupNameTimestamp produces a backup name generator that injects a timestamp
// suffix into the original filename
//
//   - naming pattern: file<suffix>timestamp.ext
//   - uses time.Now().Format(timestamp_format)
//   - stops after max_attempts if generated filenames are already claimed by existing files
func BackupNameTimestamp(suffix string, timestamp_format string, max_attempts int) BackupNameGenerator {
	return func(original_fn string, n int) string {
		if n > max_attempts {
			return ""
		}
		ext := filepath.Ext(original_fn)
		without_ext := original_fn[:len(original_fn)-len(ext)]
		ts := time.Now().Format(timestamp_format)
		return fmt.Sprintf("%s%s%s%s", without_ext, suffix, ts, ext)
	}
}

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
