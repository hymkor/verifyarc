package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ignoreEOF(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}

func compare(r1, r2 io.Reader) (bool, error) {
	br1 := bufio.NewReader(r1)
	br2 := bufio.NewReader(r2)
	for {
		c1, err1 := br1.ReadByte()
		c2, err2 := br2.ReadByte()
		if err1 != nil {
			if err1 == io.EOF {
				if err2 == io.EOF {
					return true, nil
				}
				return false, nil
			}
			return false, err1
		}
		if err2 != nil {
			return false, ignoreEOF(err2)
		}
		if c1 != c2 {
			return false, nil
		}
	}
}

type void = struct{}

func verify(enum func() (string, io.ReadCloser, error), fs1 fs.FS) error {
	touch := make(map[string]void)

	for {
		filename, r1, err := enum()
		if err != nil {
			return err
		}
		if filename == "" || r1 == nil {
			break
		}
		touch[filepath.ToSlash(filename)] = void{}
		r2, err := fs1.Open(filename)
		if err != nil {
			r1.Close()
			return fmt.Errorf("%s: %w", filename, err)
		}
		same, err := compare(r1, r2)
		r1.Close()
		r2.Close()
		if err != nil {
			return err
		}
		if !same {
			return fmt.Errorf("%s differs", filename)
		}
		fmt.Println("ARCHIVE: [OK]", filename)
	}
	return fs.WalkDir(fs1, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if _, ok := touch[filepath.ToSlash(path)]; !ok {
			return fmt.Errorf("%s not found in archive", path)
		}
		fmt.Println("FILESYS: [OK]", path)
		return nil
	})
}

func verifyZip(zipName string, dir string) error {
	zr, err := zip.OpenReader(zipName)
	if err != nil {
		return err
	}
	defer zr.Close()

	index := 0
	return verify(func() (string, io.ReadCloser, error) {
		if index >= len(zr.File) {
			return "", nil, nil
		}
		f := zr.File[index]
		index++
		rc, err := f.Open()
		return f.Name, rc, err
	}, os.DirFS("."))
}

func mains(args []string) error {
	if len(args) < 1 {
		return errors.New("too few arguments")
	}
	if strings.EqualFold(filepath.Ext(args[0]), ".zip") {
		return verifyZip(args[0], ".")
	}
	return fmt.Errorf("%s: unsupported filetype", args[0])
}

func main() {
	if err := mains(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
