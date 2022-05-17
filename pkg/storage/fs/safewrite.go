package fs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"io"

	"github.com/cockroachdb/pebble/vfs"
)

const tempFileExtension = ".crdbtmp"

func SafeWriteToFile(fs vfs.FS, dir string, filename string, b []byte) error {
	__antithesis_instrumentation__.Notify(639117)

	tempName := filename + tempFileExtension
	f, err := fs.Create(tempName)
	if err != nil {
		__antithesis_instrumentation__.Notify(639124)
		return err
	} else {
		__antithesis_instrumentation__.Notify(639125)
	}
	__antithesis_instrumentation__.Notify(639118)
	bReader := bytes.NewReader(b)
	if _, err = io.Copy(f, bReader); err != nil {
		__antithesis_instrumentation__.Notify(639126)
		f.Close()
		return err
	} else {
		__antithesis_instrumentation__.Notify(639127)
	}
	__antithesis_instrumentation__.Notify(639119)
	if err = f.Sync(); err != nil {
		__antithesis_instrumentation__.Notify(639128)
		f.Close()
		return err
	} else {
		__antithesis_instrumentation__.Notify(639129)
	}
	__antithesis_instrumentation__.Notify(639120)
	if err = f.Close(); err != nil {
		__antithesis_instrumentation__.Notify(639130)
		return err
	} else {
		__antithesis_instrumentation__.Notify(639131)
	}
	__antithesis_instrumentation__.Notify(639121)
	if err = fs.Rename(tempName, filename); err != nil {
		__antithesis_instrumentation__.Notify(639132)
		return err
	} else {
		__antithesis_instrumentation__.Notify(639133)
	}
	__antithesis_instrumentation__.Notify(639122)
	fdir, err := fs.OpenDir(dir)
	if err != nil {
		__antithesis_instrumentation__.Notify(639134)
		return err
	} else {
		__antithesis_instrumentation__.Notify(639135)
	}
	__antithesis_instrumentation__.Notify(639123)
	defer fdir.Close()
	return fdir.Sync()
}
