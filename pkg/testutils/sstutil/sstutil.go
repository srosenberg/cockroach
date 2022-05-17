package sstutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/stretchr/testify/require"
)

func MakeSST(t *testing.T, st *cluster.Settings, kvs []KV) ([]byte, roachpb.Key, roachpb.Key) {
	__antithesis_instrumentation__.Notify(646407)
	t.Helper()

	sstFile := &storage.MemFile{}
	writer := storage.MakeIngestionSSTWriter(context.Background(), st, sstFile)
	defer writer.Close()

	start, end := keys.MaxKey, keys.MinKey
	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(646409)
		if kv.Key().Compare(start) < 0 {
			__antithesis_instrumentation__.Notify(646412)
			start = kv.Key()
		} else {
			__antithesis_instrumentation__.Notify(646413)
		}
		__antithesis_instrumentation__.Notify(646410)
		if kv.Key().Compare(end) > 0 {
			__antithesis_instrumentation__.Notify(646414)
			end = kv.Key()
		} else {
			__antithesis_instrumentation__.Notify(646415)
		}
		__antithesis_instrumentation__.Notify(646411)
		if kv.Timestamp().IsEmpty() {
			__antithesis_instrumentation__.Notify(646416)
			meta := &enginepb.MVCCMetadata{RawBytes: kv.ValueBytes()}
			metaBytes, err := protoutil.Marshal(meta)
			require.NoError(t, err)
			require.NoError(t, writer.PutUnversioned(kv.Key(), metaBytes))
		} else {
			__antithesis_instrumentation__.Notify(646417)
			require.NoError(t, writer.PutMVCC(kv.MVCCKey(), kv.ValueBytes()))
		}
	}
	__antithesis_instrumentation__.Notify(646408)
	require.NoError(t, writer.Finish())
	writer.Close()

	return sstFile.Data(), start, end.Next()
}

func ScanSST(t *testing.T, sst []byte) []KV {
	__antithesis_instrumentation__.Notify(646418)
	t.Helper()

	var kvs []KV
	iter, err := storage.NewMemSSTIterator(sst, true)
	require.NoError(t, err)
	defer iter.Close()

	iter.SeekGE(storage.MVCCKey{Key: keys.MinKey})
	for {
		__antithesis_instrumentation__.Notify(646420)
		ok, err := iter.Valid()
		require.NoError(t, err)
		if !ok {
			__antithesis_instrumentation__.Notify(646422)
			break
		} else {
			__antithesis_instrumentation__.Notify(646423)
		}
		__antithesis_instrumentation__.Notify(646421)

		k := iter.UnsafeKey()
		v := roachpb.Value{RawBytes: iter.UnsafeValue()}
		value, err := v.GetBytes()
		require.NoError(t, err)
		kvs = append(kvs, KV{
			KeyString:     string(k.Key),
			WallTimestamp: k.Timestamp.WallTime,
			ValueString:   string(value),
		})
		iter.Next()
	}
	__antithesis_instrumentation__.Notify(646419)
	return kvs
}

func ComputeStats(t *testing.T, sst []byte) *enginepb.MVCCStats {
	__antithesis_instrumentation__.Notify(646424)
	t.Helper()

	iter, err := storage.NewMemSSTIterator(sst, true)
	require.NoError(t, err)
	defer iter.Close()

	stats, err := storage.ComputeStatsForRange(iter, keys.MinKey, keys.MaxKey, 0)
	require.NoError(t, err)
	return &stats
}
