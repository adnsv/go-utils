package sourcecode

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCalcAnchor(t *testing.T) {
	tests := []struct {
		buf  string
		want Anchor
	}{
		{"", Anchor{0, 0, 0}},
		{"\uFEFF", Anchor{0, 0, 3}}, // bom
		{"a", Anchor{0, 0, 1}},
		{"abc", Anchor{0, 0, 3}},
		{"\n", Anchor{1, 1, 1}},
		{"\r", Anchor{1, 1, 1}},
		{"\r\n", Anchor{1, 2, 2}},
		{"\n\r", Anchor{2, 2, 2}},
		{"\n\n", Anchor{2, 2, 2}},
		{"\r\r", Anchor{2, 2, 2}},
		{"a\n", Anchor{1, 2, 2}},
		{"a\r", Anchor{1, 2, 2}},
		{"a\r\n", Anchor{1, 3, 3}},
		{"a\n\r", Anchor{2, 3, 3}},
		{"a\n\n", Anchor{2, 3, 3}},
		{"a\r\r", Anchor{2, 3, 3}},
		{"a\nb", Anchor{1, 2, 3}},
		{"a\rb", Anchor{1, 2, 3}},
		{"a\r\nb", Anchor{1, 3, 4}},
		{"a\n\rb", Anchor{2, 3, 4}},
		{"a\n\nb", Anchor{2, 3, 4}},
		{"a\r\rb", Anchor{2, 3, 4}},
		{"a\rc\nb", Anchor{2, 4, 5}},
		{"a\nc\rb", Anchor{2, 4, 5}},
		{"a\nc\nb", Anchor{2, 4, 5}},
		{"a\rc\rb", Anchor{2, 4, 5}},
		{"ф", Anchor{0, 0, 2}}, // codeunits: 2, codepoints: 1
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.buf), func(t *testing.T) {
			if got := CalcAnchor(tt.buf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalcAnchor(%q) = %v, want %v", tt.buf, got, tt.want)
			}
		})
	}
}

func TestLocationAt(t *testing.T) {
	tests := []struct {
		name string
		buf  string
		want Location
	}{
		{"empty", "", Location{1, 1}},
		{"bom", "\uFEFF", Location{1, 2}},
		{"simple", "a", Location{1, 2}},
		{"simple", "abc", Location{1, 4}},
		{"eol", "\n", Location{2, 1}},
		{"eol", "\r\n", Location{2, 1}},
		{"eol2", "\n\r", Location{3, 1}},
		{"eol2", "\n\n", Location{3, 1}},
		{"eol2", "\r\r", Location{3, 1}},
		{"eol3", "\r \n\t\n", Location{4, 1}},
		{"eol3", "\n \r\t\r", Location{4, 1}},
		{"eol3", "\n \n\t\r", Location{4, 1}},
		{"eol3", "\r \r\t\n", Location{4, 1}},
		{"abc-n", "abc\n", Location{2, 1}},
		{"abc-n-def-n", "abc\ndef\n", Location{3, 1}},
		{"abc-n-def-n-xyz", "abc\ndef\nxyz", Location{3, 4}},
		{"tab", "\t", Location{1, 2}}, // tabs are treated as 1 columns
		{"tab-1", "\ta", Location{1, 3}},
		{"tab-2", "\tф", Location{1, 3}}, // counting runes, not codepoints
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := CalcAnchor(tt.buf)
			if got := LocationAt(tt.buf, a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LocationAt() = %v, want %v", got, tt.want)
			}
		})
	}
}
