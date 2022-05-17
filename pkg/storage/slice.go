package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func nonZeroingMakeByteSlice(len int) []byte {
	__antithesis_instrumentation__.Notify(643688)
	ptr := mallocgc(uintptr(len), nil, false)
	return (*[MaxArrayLen]byte)(ptr)[:len:len]
}
