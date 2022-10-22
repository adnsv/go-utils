package filesystem

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type OverwriteStatus = int

const (
	StatusErr = OverwriteStatus(iota)
	Creating
	Unchanged
	Overwriting
)

// WriteFileEntry is an entry in WriteFileSet.
type WriteFileEntry struct {
	Descr    string
	FilePath string
	Payload  *bytes.Buffer
	Perm     os.FileMode
	Backup   BackupNameGenerator

	status OverwriteStatus
	err    error
}

func NewWriteFileEntry(descr string, fn string, payload *bytes.Buffer) *WriteFileEntry {
	return &WriteFileEntry{
		Descr:    descr,
		FilePath: fn,
		Payload:  payload,
	}
}

func (en *WriteFileEntry) Status() OverwriteStatus {
	return en.status
}

func (en *WriteFileEntry) LastError() error {
	return en.err
}

var errMissingPayload = errors.New("missing file payload")

func (en *WriteFileEntry) UpdateStatus() {
	en.status = StatusErr
	en.err = nil

	if en.Payload == nil {
		en.err = errMissingPayload
		return
	}

	exists, err := CheckFileExists(en.FilePath)
	if err != nil {
		en.status = StatusErr
		en.err = err
		return
	}
	if !exists {
		en.status = Creating
		return
	}

	match, err := FileContentMatch(en.FilePath, en.Payload.Bytes())
	if err != nil {
		en.status = StatusErr
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
	Entries []*WriteFileEntry
}

// Add adds new entry into the set.
func (v *WriteFileset) Add(descr string, fn string, payload *bytes.Buffer) {
	v.Entries = append(v.Entries, &WriteFileEntry{
		Descr:    descr,
		FilePath: fn,
		Payload:  payload,
	})
}

// AddWithBackup adds new entry into the set.
func (v *WriteFileset) AddWithBackup(descr string, fn string, backup BackupNameGenerator, payload *bytes.Buffer) {
	v.Entries = append(v.Entries, &WriteFileEntry{
		Descr:    descr,
		FilePath: fn,
		Payload:  payload,
		Backup:   backup,
	})
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
		if en.err != nil || en.status == StatusErr {
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
func (v WriteFileset) Count(status OverwriteStatus) int {
	n := 0
	for _, en := range v.Entries {
		if en.status == status {
			n++
		}
	}
	return n
}

func (v WriteFileset) CountActionable() int {
	n := 0
	for _, en := range v.Entries {
		if en.status == Creating || en.status == Overwriting {
			n++
		}
	}
	return n
}

func (v WriteFileset) WriteOut(feedback WriteFeedbackProc) error {
	if err := v.Errors(); err != nil {
		return err
	}
	for _, en := range v.Entries {
		if en.status == Creating || en.status == Overwriting {
			opts := WriteOptions{
				Perm:       en.Perm,
				Backup:     en.Backup,
				OnFeedback: feedback,
			}
			en.err = WriteFile(en.FilePath, en.Payload.Bytes(), &opts)
		}
	}
	return v.Errors()
}
