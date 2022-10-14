package filesystem

import (
	"encoding/json"
	"os"

	"github.com/adnsv/go-utils/sourcecode"
)

// ReadJSONFile is a version of json.Unmarshal that reads content from a file
// and ammends returned errors with location information (file:col:row)
func ReadJSONFile(fn string, v interface{}) error {
	buf, err := os.ReadFile(fn)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buf, v)
	if err != nil {
		return sourcecode.MakeJsonFileLocationError(fn, buf, err)
	} else {
		return nil
	}
}
