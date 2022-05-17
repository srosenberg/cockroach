//go:build !windows && !freebsd && !dragonfly && !darwin
// +build !windows,!freebsd,!dragonfly,!darwin

package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "golang.org/x/sys/unix"

func setRlimitNoFile(limits *rlimit) error {
	__antithesis_instrumentation__.Notify(195463)
	return unix.Setrlimit(unix.RLIMIT_NOFILE, (*unix.Rlimit)(limits))
}

func getRlimitNoFile(limits *rlimit) error {
	__antithesis_instrumentation__.Notify(195464)
	return unix.Getrlimit(unix.RLIMIT_NOFILE, (*unix.Rlimit)(limits))
}
