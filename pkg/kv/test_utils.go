package kv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/kv/kvbase"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

func OnlyFollowerReads(rec tracing.Recording) bool {
	__antithesis_instrumentation__.Notify(127721)
	foundFollowerRead := false
	for _, sp := range rec {
		__antithesis_instrumentation__.Notify(127723)
		if sp.Operation == "/cockroach.roachpb.Internal/Batch" && func() bool {
			__antithesis_instrumentation__.Notify(127724)
			return sp.Tags["span.kind"] == "server" == true
		}() == true {
			__antithesis_instrumentation__.Notify(127725)
			if tracing.LogsContainMsg(sp, kvbase.FollowerReadServingMsg) {
				__antithesis_instrumentation__.Notify(127726)
				foundFollowerRead = true
			} else {
				__antithesis_instrumentation__.Notify(127727)
				return false
			}
		} else {
			__antithesis_instrumentation__.Notify(127728)
		}
	}
	__antithesis_instrumentation__.Notify(127722)
	return foundFollowerRead
}

func IsExpectedRelocateError(err error) bool {
	__antithesis_instrumentation__.Notify(127729)
	allowlist := []string{
		"descriptor changed",
		"unable to remove replica .* which is not present",
		"unable to add replica .* which is already present",
		"none of the remaining voters .* are legal additions",
		"received invalid ChangeReplicasTrigger .* to remove self",
		"raft group deleted",
		"snapshot failed",
		"breaker open",
		"unable to select removal target",
		"cannot up-replicate to .*; missing gossiped StoreDescriptor",
		"remote couldn't accept .* snapshot",
		"cannot add placeholder",
		"removing leaseholder not allowed since it isn't the Raft leader",
		"could not find a better lease transfer target for",
	}
	pattern := "(" + strings.Join(allowlist, "|") + ")"
	return testutils.IsError(err, pattern)
}
