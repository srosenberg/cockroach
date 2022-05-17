package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io"
	"os"
	"path/filepath"

	"github.com/cockroachdb/pebble/vfs"
)

type absoluteFS struct {
	fs vfs.FS
}

var _ vfs.FS = &absoluteFS{}

func (fs *absoluteFS) Create(name string) (vfs.File, error) {
	__antithesis_instrumentation__.Notify(27910)
	name, err := filepath.Abs(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(27912)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(27913)
	}
	__antithesis_instrumentation__.Notify(27911)
	return fs.fs.Create(name)
}

func (fs *absoluteFS) Link(oldname, newname string) error {
	__antithesis_instrumentation__.Notify(27914)
	return wrapWithAbsolute(fs.fs.Link, oldname, newname)
}

func (fs *absoluteFS) Open(name string, opts ...vfs.OpenOption) (vfs.File, error) {
	__antithesis_instrumentation__.Notify(27915)
	name, err := filepath.Abs(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(27917)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(27918)
	}
	__antithesis_instrumentation__.Notify(27916)
	return fs.fs.Open(name, opts...)
}

func (fs *absoluteFS) OpenDir(name string) (vfs.File, error) {
	__antithesis_instrumentation__.Notify(27919)
	return fs.fs.OpenDir(name)
}

func (fs *absoluteFS) Remove(name string) error {
	__antithesis_instrumentation__.Notify(27920)
	name, err := filepath.Abs(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(27922)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27923)
	}
	__antithesis_instrumentation__.Notify(27921)
	return fs.fs.Remove(name)
}

func (fs *absoluteFS) RemoveAll(name string) error {
	__antithesis_instrumentation__.Notify(27924)
	return fs.fs.RemoveAll(name)
}

func (fs *absoluteFS) Rename(oldname, newname string) error {
	__antithesis_instrumentation__.Notify(27925)
	return wrapWithAbsolute(fs.fs.Rename, oldname, newname)
}

func (fs *absoluteFS) ReuseForWrite(oldname, newname string) (vfs.File, error) {
	__antithesis_instrumentation__.Notify(27926)
	oldname, err := filepath.Abs(oldname)
	if err != nil {
		__antithesis_instrumentation__.Notify(27929)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(27930)
	}
	__antithesis_instrumentation__.Notify(27927)
	newname, err = filepath.Abs(newname)
	if err != nil {
		__antithesis_instrumentation__.Notify(27931)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(27932)
	}
	__antithesis_instrumentation__.Notify(27928)
	return fs.fs.ReuseForWrite(oldname, newname)
}

func (fs *absoluteFS) MkdirAll(dir string, perm os.FileMode) error {
	__antithesis_instrumentation__.Notify(27933)
	return fs.fs.MkdirAll(dir, perm)
}

func (fs *absoluteFS) Lock(name string) (io.Closer, error) {
	__antithesis_instrumentation__.Notify(27934)
	return fs.fs.Lock(name)
}

func (fs *absoluteFS) List(dir string) ([]string, error) {
	__antithesis_instrumentation__.Notify(27935)
	return fs.fs.List(dir)
}

func (fs *absoluteFS) Stat(name string) (os.FileInfo, error) {
	__antithesis_instrumentation__.Notify(27936)
	return fs.fs.Stat(name)
}

func (fs *absoluteFS) PathBase(path string) string {
	__antithesis_instrumentation__.Notify(27937)
	return fs.fs.PathBase(path)
}

func (fs *absoluteFS) PathJoin(elem ...string) string {
	__antithesis_instrumentation__.Notify(27938)
	return fs.fs.PathJoin(elem...)
}

func (fs *absoluteFS) PathDir(path string) string {
	__antithesis_instrumentation__.Notify(27939)
	return fs.fs.PathDir(path)
}

func (fs *absoluteFS) GetDiskUsage(path string) (vfs.DiskUsage, error) {
	__antithesis_instrumentation__.Notify(27940)
	return fs.fs.GetDiskUsage(path)
}

func wrapWithAbsolute(fn func(string, string) error, oldname, newname string) error {
	__antithesis_instrumentation__.Notify(27941)
	oldname, err := filepath.Abs(oldname)
	if err != nil {
		__antithesis_instrumentation__.Notify(27944)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27945)
	}
	__antithesis_instrumentation__.Notify(27942)
	newname, err = filepath.Abs(newname)
	if err != nil {
		__antithesis_instrumentation__.Notify(27946)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27947)
	}
	__antithesis_instrumentation__.Notify(27943)
	return fn(oldname, newname)
}
