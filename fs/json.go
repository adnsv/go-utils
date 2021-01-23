package fs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// JSONErrDetail ammends an error returned from json.Unmarshal with
// line:position info.
func JSONErrDetail(input string, err error) error {
	switch v := err.(type) {
	case *json.UnmarshalTypeError:
		line, pos, lcErr := LineAndCharacter(input, int(v.Offset))
		if lcErr != nil {
			return err
		}
		return fmt.Errorf("at line %d:%d %s", line+1, pos, err)
	case *json.SyntaxError:
		line, pos, lcErr := LineAndCharacter(input, int(v.Offset))
		if lcErr != nil {
			return err
		}
		return fmt.Errorf("at line %d:%d %s", line+1, pos, err)
	default:
		return err
	}
}

// WriteJSON marshals v to JSON and writes it to the
// specified file with 0666 permissions.
func WriteJSON(fn string, v interface{}) error {
	buf, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}
	return WriteFileIfChanged(fn, buf)
}

func ReadJSON(fn string, v interface{}) error {
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buf, v)
	if err != nil {
		return fmt.Errorf("failed to parse '%s':\n%s", fn, JSONErrDetail(string(buf), err))
	}
	return nil
}
