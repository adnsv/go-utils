package ansi

import "strconv"

type Style int

const (
	Reset         = "\x1b[0m"
	Bold          = "\x1b[1m"
	Dim           = "\x1b[2m"
	Italic        = "\x1b[3m"
	Underline     = "\x1b[4m"
	Blinking      = "\x1b[5m"
	Inverse       = "\x1b[7m"
	Hidden        = "\x1b[8m"
	StrikeThrough = "\x1b[9m"
)

func Foreground(v ColorIndex) string {
	return "\x1b[38;5;" + strconv.Itoa(int(v)) + "m"
}
func Background(v ColorIndex) string {
	return "\x1b[48;5;" + strconv.Itoa(int(v)) + "m"
}
func ForegroundBlack() string {
	return Foreground(Black)
}
func BackgroundBlack() string {
	return Background(Black)
}
func ForegroundWhite() string {
	return Foreground(White)
}
func BackgroundWhite() string {
	return Background(White)
}
func ForegroundGray(l byte) string {
	return Foreground(Gray(l))
}
func BackgroundGray(l byte) string {
	return Background(Gray(l))
}
func ForegroundRGB(r, g, b byte) string {
	return Foreground(RGB(r, g, b))
}
func BackgroundRGB(r, g, b byte) string {
	return Background(RGB(r, g, b))
}
