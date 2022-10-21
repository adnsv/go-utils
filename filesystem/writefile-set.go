package filesystem

import "os"

type OverwriteStatus = int

const (
	UnknownStatus = OverwriteStatus(iota)
	WritingNew
	UnchangedContent
	Overwriting
)

// WriteFileEntry is an entry in WriteFileSet.
type WriteFileEntry struct {
	Descr    string
	FilePath string
	Buffer   []byte
	Perm     os.FileMode
	Backup   BackupNameGenerator

	status OverwriteStatus
	err    error
}

func (en *WriteFileEntry) Status() OverwriteStatus {
	return en.status
}

func (en *WriteFileEntry) LastError() error {
	return en.err
}

func (en *WriteFileEntry) UpdateStatus() {
	en.status = UnknownStatus
	en.err = nil

	exists, err := CheckFileExists(en.FilePath)
	if err != nil {
		en.status = UnknownStatus
		en.err = err
		return
	}
	if !exists {
		en.status = WritingNew
		return
	}

	match, err := FileContentMatch(en.FilePath, en.Buffer)
	if err != nil {
		en.status = UnknownStatus
		en.err = err
		return
	}

	if match {
		en.status = UnchangedContent
	} else {
		en.status = Overwriting
	}
}

// WriteFileSet bundles multiple pending file write operations together.
type WriteFileSet []*WriteFileEntry

// Add adds new entry into the set.
func (v *WriteFileSet) Add(descr string, fn string, buf []byte) {
	*v = append(*v, &WriteFileEntry{
		Descr:    descr,
		FilePath: fn,
		Buffer:   buf,
	})
}

// AddWithBackup adds new entry into the set.
func (v *WriteFileSet) AddWithBackup(descr string, fn string, backup BackupNameGenerator, buf []byte) {
	*v = append(*v, &WriteFileEntry{
		Descr:    descr,
		FilePath: fn,
		Buffer:   buf,
		Backup:   backup,
	})
}

// UpdateStatus updates pending overwrite status for all entries in the set.
func (v WriteFileSet) UpdateStatus() {
	for _, en := range v {
		en.UpdateStatus()
	}
}

func (v WriteFileSet) NeedsOverwriteConfirmation() bool {
	for _, en := range v {
		if en.status == Overwriting {
			return true
		}
	}
	return false
}

func (v WriteFileSet) HasAnythingToDo() bool {
	for _, en := range v {
		if en.status == WritingNew || en.status == Overwriting {
			return true
		}
	}
	return false
}

func (v WriteFileSet) WriteOut(feedback WriteFeedbackProc) error {
	for _, en := range v {
		opts := WriteOptions{
			Perm:       en.Perm,
			Backup:     en.Backup,
			OnFeedback: feedback,
		}
		en.err = WriteFile(en.FilePath, en.Buffer, &opts)
	}
	return nil
}
