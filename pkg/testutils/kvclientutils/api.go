package kvclientutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/roachpb"

func StrToCPutExistingValue(s string) []byte {
	__antithesis_instrumentation__.Notify(644429)
	var v roachpb.Value
	v.SetBytes([]byte(s))
	return v.TagAndDataBytes()
}
