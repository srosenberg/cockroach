//go:build windows
// +build windows

package democluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func useUnixSocketsInDemo() bool {
	__antithesis_instrumentation__.Notify(32473)

	return false
}
