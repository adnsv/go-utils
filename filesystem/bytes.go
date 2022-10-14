package filesystem

import (
	"fmt"
	"math"
)

func lk(n float64) float64 {
	return math.Log(n) / math.Log(1e3)
}

var units = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

func ByteSizeStr(sz uint64) string {
	if sz < 10 {
		return fmt.Sprintf("%dB", sz)
	}

	e := math.Floor(lk(float64(sz)))
	u := units[int(e)]
	val := float64(sz) / math.Pow(1e3, math.Floor(e))
	f := "%.0f"
	if val < 10 {
		f = "%.1f"
	}

	return fmt.Sprintf(f+"%s", val, u)
}
