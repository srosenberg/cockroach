package reports

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config"
)

type ZoneKey struct {
	ZoneID config.ObjectID

	SubzoneID base.SubzoneID
}

const NoSubzone base.SubzoneID = 0

func MakeZoneKey(zoneID config.ObjectID, subzoneID base.SubzoneID) ZoneKey {
	__antithesis_instrumentation__.Notify(122135)
	return ZoneKey{
		ZoneID:    zoneID,
		SubzoneID: subzoneID,
	}
}

func (k ZoneKey) String() string {
	__antithesis_instrumentation__.Notify(122136)
	return fmt.Sprintf("%d,%d", k.ZoneID, k.SubzoneID)
}

func (k ZoneKey) Less(other ZoneKey) bool {
	__antithesis_instrumentation__.Notify(122137)
	if k.ZoneID < other.ZoneID {
		__antithesis_instrumentation__.Notify(122140)
		return true
	} else {
		__antithesis_instrumentation__.Notify(122141)
	}
	__antithesis_instrumentation__.Notify(122138)
	if k.ZoneID > other.ZoneID {
		__antithesis_instrumentation__.Notify(122142)
		return false
	} else {
		__antithesis_instrumentation__.Notify(122143)
	}
	__antithesis_instrumentation__.Notify(122139)
	return k.SubzoneID < other.SubzoneID
}
