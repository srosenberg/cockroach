package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
)

func CheckEndTxnTrigger(args kvserverbase.FilterArgs) *roachpb.Error {
	__antithesis_instrumentation__.Notify(628359)
	req, ok := args.Req.(*roachpb.EndTxnRequest)
	if !ok {
		__antithesis_instrumentation__.Notify(628364)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(628365)
	}
	__antithesis_instrumentation__.Notify(628360)

	if !req.Commit {
		__antithesis_instrumentation__.Notify(628366)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(628367)
	}
	__antithesis_instrumentation__.Notify(628361)

	modifiedSpanTrigger := req.InternalCommitTrigger.GetModifiedSpanTrigger()
	modifiedSystemConfigSpan := modifiedSpanTrigger != nil && func() bool {
		__antithesis_instrumentation__.Notify(628368)
		return modifiedSpanTrigger.SystemConfigSpan == true
	}() == true

	var hasSystemKey bool
	for _, span := range req.LockSpans {
		__antithesis_instrumentation__.Notify(628369)
		if bytes.Compare(span.Key, keys.SystemConfigSpan.Key) >= 0 && func() bool {
			__antithesis_instrumentation__.Notify(628370)
			return bytes.Compare(span.Key, keys.SystemConfigSpan.EndKey) < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(628371)
			hasSystemKey = true
			break
		} else {
			__antithesis_instrumentation__.Notify(628372)
		}
	}
	__antithesis_instrumentation__.Notify(628362)

	if hasSystemKey && func() bool {
		__antithesis_instrumentation__.Notify(628373)
		return !(clusterversion.ClusterVersion{Version: args.Version}).
			IsActive(clusterversion.DisableSystemConfigGossipTrigger) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(628374)
		return !modifiedSystemConfigSpan == true
	}() == true {
		__antithesis_instrumentation__.Notify(628375)
		return roachpb.NewError(errors.Errorf("EndTxn hasSystemKey=%t, but hasSystemConfigTrigger=%t",
			hasSystemKey, modifiedSystemConfigSpan))
	} else {
		__antithesis_instrumentation__.Notify(628376)
	}
	__antithesis_instrumentation__.Notify(628363)

	return nil
}
