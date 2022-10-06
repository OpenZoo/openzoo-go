//go:build wasm

package main

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"syscall/js"
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

type vfsReadableFile struct {
	*bytes.Reader
}

func (v vfsReadableFile) Close() error {
	return nil
}

type dummyDirEntry struct {
	name string
}

func (d dummyDirEntry) Name() string {
	return d.name
}

func (d dummyDirEntry) IsDir() bool {
	return false
}

func (d dummyDirEntry) Type() fs.FileMode {
	return 0
}

func (d dummyDirEntry) Info() (fs.FileInfo, error) {
	return nil, errors.New("FIXME")
}

func VfsOpen(name string) (VfsReadableFile, error) {
	v := js.Global().Get("ozg_vfsRead").Invoke(name)
	if !v.Truthy() {
		return nil, errors.New("file not found")
	}

	b := make([]byte, v.Length())
	js.CopyBytesToGo(b, v)

	return vfsReadableFile{bytes.NewReader(b)}, nil
}

func VfsCreate(name string) (VfsWritableFile, error) {
	return nil, errors.New("FIXME")
}

func VfsReadDir(root string) ([]fs.DirEntry, error) {
	v := js.Global().Get("ozg_vfsList").Invoke(root)
	b := make([]fs.DirEntry, v.Length())
	for i := 0; i < v.Length(); i++ {
		b[i] = dummyDirEntry{v.Index(i).String()}
	}
	return b, nil
}
