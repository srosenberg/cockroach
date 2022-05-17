package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/pebble/vfs"

type swappableFS struct {
	vfs.FS
}

func (s *swappableFS) set(fs vfs.FS) {
	__antithesis_instrumentation__.Notify(34602)
	s.FS = fs
}
