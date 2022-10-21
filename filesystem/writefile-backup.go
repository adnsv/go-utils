package filesystem

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"
)

type BackupNameGenerator = func(original_fn string, n int) string

var errFileBackupFailed = errors.New("file backup failed")

// BackupNameNumeric produces a backup name generator that injects a numbered
// suffix into the original filename:
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
// suffix into the original filename:
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
