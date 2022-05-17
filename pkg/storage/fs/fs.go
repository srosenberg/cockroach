package fs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io"
	"os"
)

type File interface {
	io.ReadWriteCloser
	io.ReaderAt
	Sync() error
}

type FS interface {
	Create(name string) (File, error)

	CreateWithSync(name string, bytesPerSync int) (File, error)

	Link(oldname, newname string) error

	Open(name string) (File, error)

	OpenDir(name string) (File, error)

	Remove(name string) error

	Rename(oldname, newname string) error

	MkdirAll(name string) error

	RemoveAll(dir string) error

	List(name string) ([]string, error)

	Stat(name string) (os.FileInfo, error)
}

func WriteFile(fs FS, filename string, data []byte) error {
	__antithesis_instrumentation__.Notify(639110)
	f, err := fs.Create(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(639113)
		return err
	} else {
		__antithesis_instrumentation__.Notify(639114)
	}
	__antithesis_instrumentation__.Notify(639111)
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		__antithesis_instrumentation__.Notify(639115)
		err = err1
	} else {
		__antithesis_instrumentation__.Notify(639116)
	}
	__antithesis_instrumentation__.Notify(639112)
	return err
}
