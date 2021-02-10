package fs

import (
	"path/filepath"
	"strings"
)

func normalizePath(s string) string {
	if s == "" {
		return s
	}

	// no spaces? return as-is
	if i := strings.IndexByte(s, ' '); i == -1 {
		return s
	}

	if filepath.ListSeparator == ':' {
		// posix
		if strings.Index(s, "\\ ") >= 0 || strings.IndexAny(s, "'\"") >= 0 {
			return s
		}
		return strings.ReplaceAll(s, " ", "\\ ")
	} else {
		// windows
		if strings.IndexAny(s, "^`\"") >= 0 {
			return s
		}
		return "\"" + s + "\""
	}

}

// JoinPathList joins multiple paths into a string with OS-specific path
// separator. This is an opposite of the GOLANG's filepath.SplitList() function.
func JoinPathList(paths ...string) string {
	tt := make([]string, 0, len(paths))
	for _, p := range paths {
		tt = append(tt, normalizePath(p))
	}
	return strings.Join(tt, string(filepath.ListSeparator))
}
