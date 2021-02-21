package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/adnsv/go-utils/pack"
)

const packageDir = "./cmd/rtag"
const deployDir = "./.deploy"
const tmpDir = "./.tmp"

func main() {
	fmt.Println("staging command line utilities for common targets")
	if _, err := os.Stat(packageDir); os.IsNotExist(err) {
		log.Fatal("please run the staging with CWD at repo root")
	} else if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(deployDir, 0755); err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		log.Fatal(err)
	}

	type oa struct {
		os   string
		arch string
	}

	for _, oa := range []oa{
		{"windows", "amd64"},
		{"windows", "386"},
		{"linux", "amd64"},
		{"linux", "386"},
		{"linux", "arm"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"freebsd", "amd64"},
	} {
		if err := build(oa.os, oa.arch); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("done")
	fmt.Println("see output in", deployDir)
}

func build(osname, archname string) error {
	exe := "rtag"
	if osname == "windows" {
		exe = "rtag.exe"
	}
	fmt.Printf("- building rtag-%s-%s\n", osname, archname)
	o := filepath.Join(tmpDir, exe)
	cmd := exec.Command("go", "build", "-o", o, packageDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GOOS="+osname, "GOARCH="+archname)
	if err := cmd.Run(); err != nil {
		return err
	}

	if archname == "386" {
		archname = "x86"
	} else if archname == "amd64" {
		archname = "x64"
	}

	zfn := filepath.Join(deployDir, fmt.Sprintf("rtag-%s-%s.zip", osname, archname))
	if err := pack.ToFlatZip(zfn, o); err != nil {
		return err
	}
	return nil
}
