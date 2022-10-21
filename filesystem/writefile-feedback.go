package filesystem

import (
	"fmt"
	"io"
)

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

// MakeWriteFileFeedback is a feedback example (use with os.Stdout or os.Stderr).
func MakeWriteFileFeedback(w io.Writer) WriteFeedbackProc {
	return func(fb WriteFeedback, fn string) {
		switch fb {
		case FeedbackWriteBegin:
			fmt.Fprintf(w, "writing %s ... ", fn)
		case FeedbackWriteFailed:
			fmt.Fprintf(w, "FAILED\n")
		case FeedbackWriteSucceded:
			fmt.Fprintf(w, "SUCCEEDED\n")
		case FeedbackWriteSkipped:
			fmt.Fprintf(w, "skipping %s: UNCHANGED\n", fn)

		case FeedbackBackupBegin:
			fmt.Fprintf(w, "backing up original to %s ... ", fn)
		case FeedbackBackupFailed:
			fmt.Fprintf(w, "FAILED\n")
		case FeedbackBackupSucceded:
			fmt.Fprintf(w, "SUCCEEDED\n")
		case FeedbackBackupRestoreFailed:
			fmt.Fprintf(w, "CRITICAL: failed to restore from backup\n")
			fmt.Fprintf(w, "CRITICAL: backed up content still available in %s\n", fn)
		}
	}
}
