package sourcecode

import (
	"encoding/json"
	"math"
)

func JsonErrorOffset64(err error) int64 {
	switch v := err.(type) {
	case *json.UnmarshalTypeError:
		return v.Offset
	case *json.SyntaxError:
		return v.Offset
	default:
		return -1
	}
}

func MakeJsonLocationError(buf []byte, err error) error {
	offset := JsonErrorOffset64(err)
	if offset < 0 || offset > math.MaxInt {
		return err
	}
	a := CalcAnchor(buf[:int(offset)])
	return NewLocationError(LocationAt(buf, a), err)
}

func MakeJsonFileLocationError(filename string, buf []byte, err error) error {
	offset := JsonErrorOffset64(err)
	if offset < 0 || offset > math.MaxInt {
		return err
	}
	a := CalcAnchor(buf[:int(offset)])
	fl := FileLocation{filename, LocationAt(buf, a)}
	return NewFileLocationError(fl, err)
}
