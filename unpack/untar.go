package unpack

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UntarToDir(src string, dst string, opts Options) error {
	dst, err := filepath.Abs(filepath.Clean(dst))
	if err != nil {
		return err
	}

	if opts.CollapseRoot != "" {
		if opts.CollapseRoot[len(opts.CollapseRoot)-1] != '/' {
			opts.CollapseRoot += "/"
		}
	}

	ext := filepath.Ext(src)

	f, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	var r *tar.Reader

	if ext == ".bz2" {
		z := bzip2.NewReader(f)
		r = tar.NewReader(z)
	} else if ext == ".gz" {
		z, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer z.Close()
		r = tar.NewReader(z)
	} else if ext == ".tar" {
		r = tar.NewReader(f)
	} else {
		return fmt.Errorf("unsupported file extension %s", ext)
	}

	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		name := header.Name
		info := header.FileInfo()

		if opts.FilterAllow != nil {
			allow := opts.FilterAllow(name)
			if !allow {
				continue
			}
		}

		if opts.CollapseRoot != "" {
			// remove first level
			if name == opts.CollapseRoot {
				if info.IsDir() {
					continue
				} else {
					return fmt.Errorf("failed to collapse root for file %s", name)
				}
			}
			if strings.HasPrefix(name, opts.CollapseRoot) {
				name = name[len(opts.CollapseRoot):]
			} else {
				return fmt.Errorf("failed to collapse root for path %s", name)
			}
		}

		dstpath := filepath.Join(dst, name)
		if info.IsDir() {
			if err := os.MkdirAll(dstpath, os.ModePerm); err != nil {
				return err
			}
		} else {
			of, err := os.OpenFile(dstpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
			if err != nil {
				return err
			}
			defer of.Close()
			if _, err = io.Copy(of, r); err != nil {
				return err
			}
		}
	}
	return nil
}
