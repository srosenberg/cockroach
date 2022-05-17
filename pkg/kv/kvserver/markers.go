package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
)

var errMarkSnapshotError = errors.New("snapshot failed")

func isSnapshotError(err error) bool {
	__antithesis_instrumentation__.Notify(110106)
	return errors.Is(err, errMarkSnapshotError)
}

var errMarkCanRetryReplicationChangeWithUpdatedDesc = errors.New("should retry with updated descriptor")

func IsRetriableReplicationChangeError(err error) bool {
	__antithesis_instrumentation__.Notify(110107)
	return errors.Is(err, errMarkCanRetryReplicationChangeWithUpdatedDesc) || func() bool {
		__antithesis_instrumentation__.Notify(110108)
		return isSnapshotError(err) == true
	}() == true
}

const (
	descChangedErrorFmt              = "descriptor changed: [expected] %s != [actual] %s"
	descChangedRangeReplacedErrorFmt = "descriptor changed: [expected] %s != [actual] %s  (range replaced)"
	descChangedRangeSubsumedErrorFmt = "descriptor changed: [expected] %s != [actual] nil (range subsumed)"
)

func newDescChangedError(desc, actualDesc *roachpb.RangeDescriptor) error {
	__antithesis_instrumentation__.Notify(110109)
	var err error
	if actualDesc == nil {
		__antithesis_instrumentation__.Notify(110111)
		err = errors.Newf(descChangedRangeSubsumedErrorFmt, desc)
	} else {
		__antithesis_instrumentation__.Notify(110112)
		if desc.RangeID != actualDesc.RangeID {
			__antithesis_instrumentation__.Notify(110113)
			err = errors.Newf(descChangedRangeReplacedErrorFmt, desc, actualDesc)
		} else {
			__antithesis_instrumentation__.Notify(110114)
			err = errors.Newf(descChangedErrorFmt, desc, actualDesc)
		}
	}
	__antithesis_instrumentation__.Notify(110110)
	return errors.Mark(err, errMarkCanRetryReplicationChangeWithUpdatedDesc)
}

func wrapDescChangedError(err error, desc, actualDesc *roachpb.RangeDescriptor) error {
	__antithesis_instrumentation__.Notify(110115)
	var wrapped error
	if actualDesc == nil {
		__antithesis_instrumentation__.Notify(110117)
		wrapped = errors.Wrapf(err, descChangedRangeSubsumedErrorFmt, desc)
	} else {
		__antithesis_instrumentation__.Notify(110118)
		if desc.RangeID != actualDesc.RangeID {
			__antithesis_instrumentation__.Notify(110119)
			wrapped = errors.Wrapf(err, descChangedRangeReplacedErrorFmt, desc, actualDesc)
		} else {
			__antithesis_instrumentation__.Notify(110120)
			wrapped = errors.Wrapf(err, descChangedErrorFmt, desc, actualDesc)
		}
	}
	__antithesis_instrumentation__.Notify(110116)
	return errors.Mark(wrapped, errMarkCanRetryReplicationChangeWithUpdatedDesc)
}

var errMarkInvalidReplicationChange = errors.New("invalid replication change")

func IsIllegalReplicationChangeError(err error) bool {
	__antithesis_instrumentation__.Notify(110121)
	return errors.Is(err, errMarkInvalidReplicationChange)
}
