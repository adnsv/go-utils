package download

import (
	"io"
	"net/http"
	"os"
)

// FromHTTP downloads content from the specified url
func FromHTTP(dst io.Writer, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(dst, resp.Body)

	return err
}

// FileFromHTTP downloads content from the specified url into a file
func FileFromHTTP(fn string, url string) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	return FromHTTP(f, url)
}

// StatHTTP indicates whether the specified URL responds with http.StatusOK
func StatHTTP(url string) bool {
	resp, e := http.Head(url)
	avail := e == nil && resp.StatusCode == http.StatusOK
	resp.Body.Close()
	return avail
}
