package spanconfigptsreader

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

type adapter struct {
	cache        protectedts.Cache
	kvSubscriber spanconfig.KVSubscriber
}

var _ spanconfig.ProtectedTSReader = &adapter{}

func NewAdapter(
	cache protectedts.Cache, kvSubscriber spanconfig.KVSubscriber,
) spanconfig.ProtectedTSReader {
	__antithesis_instrumentation__.Notify(240768)
	return &adapter{
		cache:        cache,
		kvSubscriber: kvSubscriber,
	}
}

func (a *adapter) GetProtectionTimestamps(
	ctx context.Context, sp roachpb.Span,
) (protectionTimestamps []hlc.Timestamp, asOf hlc.Timestamp, err error) {
	__antithesis_instrumentation__.Notify(240769)
	cacheTimestamps, cacheFreshness, err := a.cache.GetProtectionTimestamps(ctx, sp)
	if err != nil {
		__antithesis_instrumentation__.Notify(240772)
		return nil, hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(240773)
	}
	__antithesis_instrumentation__.Notify(240770)
	subscriberTimestamps, subscriberFreshness, err := a.kvSubscriber.GetProtectionTimestamps(ctx, sp)
	if err != nil {
		__antithesis_instrumentation__.Notify(240774)
		return nil, hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(240775)
	}
	__antithesis_instrumentation__.Notify(240771)

	subscriberFreshness.Backward(cacheFreshness)
	return append(subscriberTimestamps, cacheTimestamps...), subscriberFreshness, nil
}

func TestingRefreshPTSState(
	ctx context.Context,
	t *testing.T,
	protectedTSReader spanconfig.ProtectedTSReader,
	asOf hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(240776)
	a, ok := protectedTSReader.(*adapter)
	if !ok {
		__antithesis_instrumentation__.Notify(240779)
		return errors.AssertionFailedf("could not convert protectedTSReader to adapter")
	} else {
		__antithesis_instrumentation__.Notify(240780)
	}
	__antithesis_instrumentation__.Notify(240777)

	require.NoError(t, a.cache.Refresh(ctx, asOf))

	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(240781)
		_, fresh, err := a.GetProtectionTimestamps(ctx, keys.EverythingSpan)
		if err != nil {
			__antithesis_instrumentation__.Notify(240784)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240785)
		}
		__antithesis_instrumentation__.Notify(240782)
		if fresh.Less(asOf) {
			__antithesis_instrumentation__.Notify(240786)
			return errors.AssertionFailedf("KVSubscriber fresh as of %s; not caught up to %s", fresh, asOf)
		} else {
			__antithesis_instrumentation__.Notify(240787)
		}
		__antithesis_instrumentation__.Notify(240783)
		return nil
	})
	__antithesis_instrumentation__.Notify(240778)
	return nil
}
