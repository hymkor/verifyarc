package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var flagCurdir = flag.String("C", ".", "set the current directory")

func compare(r1, r2 io.Reader) (bool, error) {
	const UNIT = 64 * 1024
	var buffer1 [UNIT]byte
	var buffer2 [UNIT]byte

	for {
		n1, err1 := io.ReadFull(r1, buffer1[:])
		n2, err2 := io.ReadFull(r2, buffer2[:])

		if err1 != nil {
			if err1 == io.EOF || err1 == io.ErrUnexpectedEOF {
				if err2 == io.EOF || err2 == io.ErrUnexpectedEOF {
					return n1 == n2 && bytes.Equal(buffer1[:n1], buffer2[:n2]), nil
				}
				return false, err2
			}
			return false, err1
		}
		if err2 != nil {
			return false, err2
		}
		if n1 != n2 || !bytes.Equal(buffer1[:n1], buffer2[:n2]) {
			return false, nil
		}
	}
}

type void = struct{}

func verify(root string, enum func() (string, io.ReadCloser, error)) error {
	touch := make(map[string]void)
	touchedDir := make(map[string]void)

	for {
		filename, r1, err := enum()
		if err != nil {
			return err
		}
		if filename == "" || r1 == nil {
			break
		}
		touch[filepath.ToSlash(filename)] = void{}
		localPath := filepath.Join(root, filename)
		r2, err := os.Open(localPath)
		if err != nil {
			r1.Close()
			return fmt.Errorf("%s: os.Open: %w", filename, err)
		}
		same, err := compare(r1, r2)
		r1.Close()
		r2.Close()
		if err != nil {
			return err
		}
		if !same {
			return fmt.Errorf("ARCHIVE: [DIFFER] %s", filename)
		}
		fmt.Println("ARCHIVE: [OK]", filename)
		localDir := filepath.Dir(localPath)
		touchedDir[localDir] = void{}
	}
	for tdir := range touchedDir {
		err := filepath.WalkDir(tdir, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}
			if _path, err := filepath.Rel(root, path); err == nil {
				path = _path
			}
			if _, ok := touch[filepath.ToSlash(path)]; ok {
				fmt.Println("FILESYS: [OK]", path)
			} else {
				fmt.Println("FILESYS: [NOT FOUND]", path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func verifyZip(zipName string, dir string) error {
	zr, err := zip.OpenReader(zipName)
	if err != nil {
		return err
	}
	defer zr.Close()

	index := 0
	return verify(dir, func() (string, io.ReadCloser, error) {
		if index >= len(zr.File) {
			return "", nil, nil
		}
		f := zr.File[index]
		index++
		rc, err := f.Open()
		return f.Name, rc, err
	})
}

func verifyTar(tarName string, dir string) error {
	var in io.Reader
	if tarName == "" || tarName == "-" {
		in = os.Stdin
		defer io.Copy(io.Discard, os.Stdin)
	} else {
		_in, err := os.Open(tarName)
		if err != nil {
			return err
		}
		defer _in.Close()
		in = _in
	}
	tr := tar.NewReader(in)
	return verify(dir, func() (string, io.ReadCloser, error) {
		for {
			header, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					return "", nil, nil
				}
				return "", nil, err
			}
			if n := header.Name; len(n) <= 0 || n[len(n)-1] != '/' {
				return header.Name, io.NopCloser(tr), nil
			}
		}
	})
}

func verifyDir(targetDir string, dir string) error {
	var atLater []string
	entry, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}
	return verify(dir, func() (string, io.ReadCloser, error) {
		for {
			for len(entry) <= 0 {
				if len(atLater) <= 0 {
					return "", nil, nil
				}
				targetDir = atLater[len(atLater)-1]
				atLater = atLater[:len(atLater)-1]
				entry, err = os.ReadDir(targetDir)
				if err != nil {
					return "", nil, err
				}
			}
			p := entry[len(entry)-1]
			entry = entry[:len(entry)-1]
			fullpath := filepath.Join(targetDir, p.Name())
			if !p.IsDir() {
				fd, err := os.Open(fullpath)
				return fullpath, fd, err
			}
			if name := p.Name(); name != "." && name != ".." {
				atLater = append(atLater, fullpath)
			}
		}
	})
}

func isDir(s string) bool {
	if len(s) <= 0 {
		return false
	}
	if last := s[len(s)-1]; last == '/' || last == '\\' {
		return true
	}
	stat, err := os.Stat(s)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

var version string = "snapshot"

func mains(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%s %s-%s-%s by %s\n",
			os.Args[0],
			version, runtime.GOOS, runtime.GOARCH, runtime.Version())
	}
	if strings.EqualFold(filepath.Ext(args[0]), ".zip") {
		return verifyZip(args[0], *flagCurdir)
	}
	if isDir(args[0]) {
		return verifyDir(args[0], *flagCurdir)
	}
	return verifyTar(args[0], *flagCurdir)
}

func main() {
	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
