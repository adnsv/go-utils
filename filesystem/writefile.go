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
