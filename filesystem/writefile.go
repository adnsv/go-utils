package filesystem

import (
	"errors"
	"io/fs"
	"os"
)

// WriteOptions provides detailed configuration for tuning WriteFile behavior.
type WriteOptions struct {
	Perm                     os.FileMode         // file writing permissions, defaults to 0666 if unspecidied (perm == 0)
	OverwriteMatchingContent bool                // backup and overwrite, even if content matches
	Backup                   BackupNameGenerator // backup filename generator, no backup by default
	OnFeedback               WriteFeedbackProc   // use this if logging or user feedback is required
}

// WriteFile writes data to the named file with configurable behavior and
// feedback.
func WriteFile(fn string, buf []byte, opts *WriteOptions) error {
	_, err := WriteFileEx(fn, buf, opts)
	return err
}

type WriteFileStatus = int

const (
	StatErr = WriteFileStatus(iota)

	Unchanged
	Creating
	Overwriting

	Failed
	Skipped
	Succeeded
)

// WriteFileEx writes data to the named file with configurable behavior and
// feedback, provides detailed status.
func WriteFileEx(fn string, buf []byte, opts *WriteOptions) (status WriteFileStatus, err error) {
	if opts == nil {
		err = os.WriteFile(fn, buf, 0666)
		if err == nil {
			status = Succeeded
		} else {
			status = Failed
		}
		return
	}

	// effective permissions
	perm := opts.Perm
	if perm == 0 {
		perm = 0666
	}

	perform_write := func() {
		if opts.OnFeedback != nil {
			opts.OnFeedback(FeedbackWriteBegin, fn)
		}
		err = os.WriteFile(fn, buf, perm)
		if err == nil {
			status = Succeeded
			if opts.OnFeedback != nil {
				opts.OnFeedback(FeedbackWriteSucceded, fn)
			}
		} else {
			status = Failed
			if opts.OnFeedback != nil {
				opts.OnFeedback(FeedbackWriteFailed, fn)
			}
		}
	}

	if !opts.OverwriteMatchingContent {
		var match bool
		match, err = FileContentMatch(fn, buf)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				// creating new
				perform_write()
				return
			} else {
				// errored while matching content
				status = Failed
				if opts.OnFeedback != nil {
					opts.OnFeedback(FeedbackWriteBegin, fn)
					opts.OnFeedback(FeedbackWriteFailed, fn)
				}
				return status, err
			}
		}
		if match {
			status = Skipped
			if opts.OnFeedback != nil {
				opts.OnFeedback(FeedbackWriteSkipped, fn)
			}
			return status, nil
		}
	}

	if opts.Backup == nil {
		perform_write()
		return
	}

	backup_attempt := 1
	backup_fn := opts.Backup(fn, backup_attempt)
	if backup_fn == "" {
		status = Failed
		err = errFileBackupFailed
		return
	}

	for FileExists(backup_fn) {
		backup_attempt++
		backup_fn = opts.Backup(fn, backup_attempt)
		if backup_fn == "" {
			status = Failed
			err = errFileBackupFailed
			return
		}
	}

	if opts.OnFeedback != nil {
		opts.OnFeedback(FeedbackBackupBegin, backup_fn)
	}
	if err = os.Rename(fn, backup_fn); err != nil {
		status = Failed
		if opts.OnFeedback != nil {
			opts.OnFeedback(FeedbackBackupFailed, backup_fn)
		}
		return
	}
	if opts.OnFeedback != nil {
		opts.OnFeedback(FeedbackBackupSucceded, backup_fn)
	}

	perform_write()
	if err != nil {
		// try to restore backup
		restore_err := os.Rename(backup_fn, fn)
		if restore_err != nil && opts.OnFeedback != nil {
			opts.OnFeedback(FeedbackBackupRestoreFailed, backup_fn)
		}
	}
	return
}
