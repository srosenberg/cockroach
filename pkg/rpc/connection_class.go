package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type ConnectionClass int8

const (
	DefaultClass ConnectionClass = iota

	SystemClass

	RangefeedClass

	NumConnectionClasses int = iota
)

var connectionClassName = map[ConnectionClass]string{
	DefaultClass:   "default",
	SystemClass:    "system",
	RangefeedClass: "rangefeed",
}

func (c ConnectionClass) String() string {
	__antithesis_instrumentation__.Notify(184285)
	return connectionClassName[c]
}

func (ConnectionClass) SafeValue() { __antithesis_instrumentation__.Notify(184286) }

var systemClassKeyPrefixes = []roachpb.RKey{
	roachpb.RKey(keys.Meta1Prefix),
	roachpb.RKey(keys.NodeLivenessPrefix),
}

func ConnectionClassForKey(key roachpb.RKey) ConnectionClass {
	__antithesis_instrumentation__.Notify(184287)

	if len(key) == 0 {
		__antithesis_instrumentation__.Notify(184290)
		return SystemClass
	} else {
		__antithesis_instrumentation__.Notify(184291)
	}
	__antithesis_instrumentation__.Notify(184288)
	for _, prefix := range systemClassKeyPrefixes {
		__antithesis_instrumentation__.Notify(184292)
		if bytes.HasPrefix(key, prefix) {
			__antithesis_instrumentation__.Notify(184293)
			return SystemClass
		} else {
			__antithesis_instrumentation__.Notify(184294)
		}
	}
	__antithesis_instrumentation__.Notify(184289)
	return DefaultClass
}
