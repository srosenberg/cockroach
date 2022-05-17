// Package reporoot contains utilities to determine the repository root.
package reporoot

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors/oserror"
)

func Get() string {
	__antithesis_instrumentation__.Notify(68502)
	return GetFor(".", ".git")
}

func GetFor(path string, checkFor string) string {
	__antithesis_instrumentation__.Notify(68503)
	path, err := filepath.Abs(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(68505)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(68506)
	}
	__antithesis_instrumentation__.Notify(68504)
	for {
		__antithesis_instrumentation__.Notify(68507)
		s, err := os.Stat(filepath.Join(path, checkFor))
		if err != nil {
			__antithesis_instrumentation__.Notify(68509)
			if !oserror.IsNotExist(err) {
				__antithesis_instrumentation__.Notify(68510)
				return ""
			} else {
				__antithesis_instrumentation__.Notify(68511)
			}
		} else {
			__antithesis_instrumentation__.Notify(68512)
			if s != nil {
				__antithesis_instrumentation__.Notify(68513)
				return path
			} else {
				__antithesis_instrumentation__.Notify(68514)
			}
		}
		__antithesis_instrumentation__.Notify(68508)
		path = filepath.Dir(path)
		if path == "/" {
			__antithesis_instrumentation__.Notify(68515)
			return ""
		} else {
			__antithesis_instrumentation__.Notify(68516)
		}
	}
}
