package spanconfigkvaccessor

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

var (
	NoopKVAccessor = dummyKVAccessor{error: nil}

	DisabledKVAccessor = dummyKVAccessor{error: errors.New("span configs disabled")}
)

type dummyKVAccessor struct {
	error error
}

var _ spanconfig.KVAccessor = &dummyKVAccessor{}

func (k dummyKVAccessor) GetSpanConfigRecords(
	context.Context, []spanconfig.Target,
) ([]spanconfig.Record, error) {
	__antithesis_instrumentation__.Notify(240382)
	return nil, k.error
}

func (k dummyKVAccessor) UpdateSpanConfigRecords(
	context.Context, []spanconfig.Target, []spanconfig.Record, hlc.Timestamp, hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(240383)
	return k.error
}

func (k dummyKVAccessor) GetAllSystemSpanConfigsThatApply(
	context.Context, roachpb.TenantID,
) ([]roachpb.SpanConfig, error) {
	__antithesis_instrumentation__.Notify(240384)
	return nil, k.error
}

func (k dummyKVAccessor) WithTxn(context.Context, *kv.Txn) spanconfig.KVAccessor {
	__antithesis_instrumentation__.Notify(240385)
	return k
}
