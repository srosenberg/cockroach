//go:build !windows && !freebsd && !dragonfly
// +build !windows,!freebsd,!dragonfly

package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

// #include <sys/types.h>
// #include <sys/sysctl.h>
import "C"

func setRlimitNoFile(limits *rlimit) error {
	__antithesis_instrumentation__.Notify(195440)
	return unix.Setrlimit(unix.RLIMIT_NOFILE, (*unix.Rlimit)(limits))
}

func getRlimitNoFile(limits *rlimit) error {
	__antithesis_instrumentation__.Notify(195441)
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, (*unix.Rlimit)(limits)); err != nil {
		__antithesis_instrumentation__.Notify(195447)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195448)
	}
	__antithesis_instrumentation__.Notify(195442)

	sysctlMaxFiles, err := getSysctlMaxFiles()
	if err != nil {
		__antithesis_instrumentation__.Notify(195449)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195450)
	}
	__antithesis_instrumentation__.Notify(195443)
	if limits.Max > sysctlMaxFiles {
		__antithesis_instrumentation__.Notify(195451)
		limits.Max = sysctlMaxFiles
	} else {
		__antithesis_instrumentation__.Notify(195452)
	}
	__antithesis_instrumentation__.Notify(195444)
	sysctlMaxFilesPerProc, err := getSysctlMaxFilesPerProc()
	if err != nil {
		__antithesis_instrumentation__.Notify(195453)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195454)
	}
	__antithesis_instrumentation__.Notify(195445)
	if limits.Max > sysctlMaxFilesPerProc {
		__antithesis_instrumentation__.Notify(195455)
		limits.Max = sysctlMaxFilesPerProc
	} else {
		__antithesis_instrumentation__.Notify(195456)
	}
	__antithesis_instrumentation__.Notify(195446)
	return nil
}

func getSysctlMaxFiles() (uint64, error) {
	__antithesis_instrumentation__.Notify(195457)
	return getSysctl(C.CTL_KERN, C.KERN_MAXFILES)
}

func getSysctlMaxFilesPerProc() (uint64, error) {
	__antithesis_instrumentation__.Notify(195458)
	return getSysctl(C.CTL_KERN, C.KERN_MAXFILESPERPROC)
}

func getSysctl(x, y C.int) (uint64, error) {
	__antithesis_instrumentation__.Notify(195459)
	var out int32
	outLen := C.size_t(unsafe.Sizeof(out))
	sysctlMib := [...]C.int{x, y}
	r, errno := C.sysctl(&sysctlMib[0], C.u_int(len(sysctlMib)), unsafe.Pointer(&out), &outLen,
		nil, 0)
	if r != 0 {
		__antithesis_instrumentation__.Notify(195461)
		return 0, errno
	} else {
		__antithesis_instrumentation__.Notify(195462)
	}
	__antithesis_instrumentation__.Notify(195460)
	return uint64(out), nil
}
