package unpack

type Options struct {
	CollapseRoot string
	FilterAllow  func(fn string) bool
}
