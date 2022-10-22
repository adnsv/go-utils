package filesystem

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/exp/slices"
)

// WriteFileEntry is an entry in WriteFileSet.
type WriteFileEntry struct {
	Descr    string
	FilePath string
	Payload  *bytes.Buffer
	Perm     os.FileMode
	Backup   BackupNameGenerator
	Tag      string

	status WriteFileStatus
	err    error
}

func NewWriteFileEntry(descr string, fn string, payload *bytes.Buffer) *WriteFileEntry {
	return &WriteFileEntry{
		Descr:    descr,
		FilePath: fn,
		Payload:  payload,
	}
}

func (en *WriteFileEntry) Status() WriteFileStatus {
	return en.status
}

func (en *WriteFileEntry) LastError() error {
	return en.err
}

var errMissingFileBuffer = errors.New("missing file buffer")
var errEmptyFilePath = errors.New("empty filepath")

func (en *WriteFileEntry) UpdateStatus() {
	en.status = StatErr
	en.err = nil

	if en.FilePath == "" {
		en.err = errEmptyFilePath
		return
	}
	if en.Payload == nil {
		en.err = errMissingFileBuffer
		return
	}

	exists, err := CheckFileExists(en.FilePath)
	if err != nil {
		en.status = StatErr
		en.err = err
		return
	}
	if !exists {
		en.status = Creating
		return
	}

	match, err := FileContentMatch(en.FilePath, en.Payload.Bytes())
	if err != nil {
		en.status = StatErr
		en.err = err
		return
	}

	if match {
		en.status = Unchanged
	} else {
		en.status = Overwriting
	}
}

// WriteFileset bundles multiple pending file write operations together.
type WriteFileset struct {
	Entries    []*WriteFileEntry
	OnFeedback WriteFeedbackProc
}

// Add adds new entry into the set.
func (v *WriteFileset) Add(descr string, fn string, payload *bytes.Buffer) *WriteFileEntry {
	en := &WriteFileEntry{
		Descr:    descr,
		FilePath: fn,
		Payload:  payload,
	}
	v.Entries = append(v.Entries, en)
	return en
}

type FileEntriesWithErrors []*WriteFileEntry

var errUnknownFilesetStatus = errors.New("unknown fileset status")

func (e FileEntriesWithErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	en := e[0]
	if en.err == nil {
		return errUnknownFilesetStatus.Error()
	}
	s := en.Descr
	if s == "" {
		s = "file: " + filepath.Base(en.FilePath)
	}
	if len(e) > 1 {
		return fmt.Sprintf("error in %s (+%d more): %s", s, len(e)-1, en.err)
	} else {
		return fmt.Sprintf("error in %s: %s", s, en.err)
	}
}

func (v WriteFileset) Errors() error {
	var r FileEntriesWithErrors
	for _, en := range v.Entries {
		if en.err != nil || en.status == StatErr {
			r = append(r, en)
		}
	}
	if len(r) > 0 {
		return r
	} else {
		return nil
	}
}

// UpdateStatus updates pending overwrite status for all entries in the set.
func (v WriteFileset) UpdateStatus() error {
	for _, en := range v.Entries {
		en.UpdateStatus()
	}
	return v.Errors()
}

// Count counts the entries with have the specified status value.
func (v WriteFileset) Count(status WriteFileStatus) int {
	n := 0
	for _, en := range v.Entries {
		if en.status == status {
			n++
		}
	}
	return n
}

// CountPending counts the entries that are still pending for write operation.
func (v WriteFileset) CountPending() int {
	n := 0
	for _, en := range v.Entries {
		if en.status == Creating || en.status == Overwriting {
			n++
		}
	}
	return n
}

// WriteTagged writes out pending entries that have matching tags.
func (v WriteFileset) WriteTagged(tags ...string) error {
	if err := v.Errors(); err != nil {
		return err
	}
	for _, en := range v.Entries {
		if (en.status == Creating || en.status == Overwriting) && slices.Contains(tags, en.Tag) {
			opts := WriteOptions{
				Perm:       en.Perm,
				Backup:     en.Backup,
				OnFeedback: v.OnFeedback,
			}
			en.status, en.err = WriteFileEx(en.FilePath, en.Payload.Bytes(), &opts)
		}
	}
	return v.Errors()
}

// WriteTagged writes out all pending entries.
func (v WriteFileset) WriteOut() error {
	if err := v.Errors(); err != nil {
		return err
	}
	for _, en := range v.Entries {
		if en.status == Creating || en.status == Overwriting {
			opts := WriteOptions{
				Perm:       en.Perm,
				Backup:     en.Backup,
				OnFeedback: v.OnFeedback,
			}
			en.status, en.err = WriteFileEx(en.FilePath, en.Payload.Bytes(), &opts)
		}
	}
	return v.Errors()
}
