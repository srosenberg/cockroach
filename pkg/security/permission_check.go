package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
	"github.com/cockroachdb/errors"
)

func checkFilePermissions(processGID int, fullKeyPath string, fileACL sysutil.ACLInfo) error {
	__antithesis_instrumentation__.Notify(186984)

	if fileACL.IsOwnedByUID(uint64(0)) && func() bool {
		__antithesis_instrumentation__.Notify(186987)
		return fileACL.IsOwnedByGID(uint64(processGID)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(186988)

		if sysutil.ExceedsPermissions(fileACL.Mode(), maxGroupKeyPermissions) {
			__antithesis_instrumentation__.Notify(186990)
			return errors.Errorf("key file %s has permissions %s, exceeds %s",
				fullKeyPath, fileACL.Mode(), maxGroupKeyPermissions)

		} else {
			__antithesis_instrumentation__.Notify(186991)
		}
		__antithesis_instrumentation__.Notify(186989)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(186992)
	}
	__antithesis_instrumentation__.Notify(186985)

	if sysutil.ExceedsPermissions(fileACL.Mode(), maxKeyPermissions) {
		__antithesis_instrumentation__.Notify(186993)
		return errors.Errorf("key file %s has permissions %s, exceeds %s",
			fullKeyPath, fileACL.Mode(), maxKeyPermissions)
	} else {
		__antithesis_instrumentation__.Notify(186994)
	}
	__antithesis_instrumentation__.Notify(186986)

	return nil
}
