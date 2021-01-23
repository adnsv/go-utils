package unpack

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UnzipToDir(src string, dst string, opts Options) error {
	if opts.CollapseRoot != "" {
		if opts.CollapseRoot[len(opts.CollapseRoot)-1] != '/' {
			opts.CollapseRoot += "/"
		}
	}
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for f := range r.File {
		fn := r.File[f].Name
		info := r.File[f].FileInfo()
		if opts.FilterAllow != nil {
			allow := opts.FilterAllow(fn)
			if !allow {
				continue
			}
		}
		if opts.CollapseRoot != "" {
			// remove first level
			if fn == opts.CollapseRoot {
				if info.IsDir() {
					continue
				} else {
					return fmt.Errorf("failed to collapse root for file \"%s\"", fn)
				}
			}
			if strings.HasPrefix(fn, opts.CollapseRoot) {
				fn = fn[len(opts.CollapseRoot):]
			} else {
				return fmt.Errorf("failed to collapse root for path \"%s\"", fn)
			}
		}
		dstpath := filepath.Join(dst, fn)
		if !strings.HasPrefix(dstpath, dst) {
			return fmt.Errorf("illegal path expansion for item \"%s\": resulting path is \"%s\"", r.File[f].Name, dstpath)
		}
		if info.IsDir() {
			if err := os.MkdirAll(dstpath, os.ModePerm); err != nil {
				return err
			}
		} else {
			rc, err := r.File[f].Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			of, err := os.OpenFile(dstpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
			if err != nil {
				return err
			}
			defer of.Close()
			if _, err = io.Copy(of, rc); err != nil {
				return err
			}
			of.Close()
			rc.Close()
		}
	}
	return nil
}
