package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "bytes"

func MakeKey(keys ...[]byte) []byte {
	__antithesis_instrumentation__.Notify(644380)
	return bytes.Join(keys, nil)
}
