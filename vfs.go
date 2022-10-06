//go:build !wasm

package main

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type VfsReadableFile interface {
	io.Reader
	io.Seeker
	io.Closer
}

type VfsWritableFile interface {
	io.Writer
	io.Seeker
	io.Closer
}

func pathFindCaseInsensitiveFile(s string) string {
	absS, err := filepath.Abs(s)
	if err != nil {
		return s
	}
	if runtime.GOOS == "windows" {
		// Windows uses case-insensitive file access by default.
		return s
	}
	parentS := filepath.Dir(absS)
	filenameS := filepath.Base(s)
	if err != nil {
		return s
	}
	filepath.WalkDir(parentS, func(path string, d fs.DirEntry, err error) error {
		if err == nil {
			filenameCandS := filepath.Base(path)
			if strings.EqualFold(filenameCandS, filenameS) {
				filenameS = filenameCandS
			}
		}
		return err
	})
	return filepath.Join(parentS, filenameS)
}

func VfsOpen(name string) (VfsReadableFile, error) {
	return os.Open(pathFindCaseInsensitiveFile(name))
}

func VfsCreate(name string) (VfsWritableFile, error) {
	return os.Create(pathFindCaseInsensitiveFile(name))
}

func VfsReadDir(root string) ([]fs.DirEntry, error) {
	return os.ReadDir(root)
}
