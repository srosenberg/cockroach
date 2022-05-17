package base

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type SubzoneID uint32

func (id SubzoneID) ToSubzoneIndex() int32 {
	__antithesis_instrumentation__.Notify(1769)
	return int32(id) - 1
}

func SubzoneIDFromIndex(idx int) SubzoneID {
	__antithesis_instrumentation__.Notify(1770)
	return SubzoneID(idx + 1)
}
