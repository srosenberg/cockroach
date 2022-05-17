//go:build freebsd || dragonfly
// +build freebsd dragonfly

package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"golang.org/x/sys/unix"
)

func setRlimitNoFile(limits *rlimit) error {
	__antithesis_instrumentation__.Notify(195429)
	rLimit := unix.Rlimit{Cur: int64(limits.Cur), Max: int64(limits.Max)}
	return unix.Setrlimit(unix.RLIMIT_NOFILE, &rLimit)
}

func getRlimitNoFile(limits *rlimit) error {
	__antithesis_instrumentation__.Notify(195430)
	var rLimit unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &rLimit); err != nil {
		__antithesis_instrumentation__.Notify(195434)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195435)
	}
	__antithesis_instrumentation__.Notify(195431)

	if rLimit.Cur == -1 {
		__antithesis_instrumentation__.Notify(195436)
		limits.Cur = math.MaxUint64
	} else {
		__antithesis_instrumentation__.Notify(195437)
		limits.Cur = uint64(rLimit.Cur)
	}
	__antithesis_instrumentation__.Notify(195432)
	if rLimit.Max == -1 {
		__antithesis_instrumentation__.Notify(195438)
		limits.Max = math.MaxUint64
	} else {
		__antithesis_instrumentation__.Notify(195439)
		limits.Max = uint64(rLimit.Max)
	}
	__antithesis_instrumentation__.Notify(195433)
	return nil
}
