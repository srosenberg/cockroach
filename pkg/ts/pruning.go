package ts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

var (
	firstTSRKey = roachpb.RKey(keys.TimeseriesPrefix)
	lastTSRKey  = firstTSRKey.PrefixEnd()
)

type timeSeriesResolutionInfo struct {
	Name       string
	Resolution Resolution
}

func (tsdb *DB) findTimeSeries(
	snapshot storage.Reader, startKey, endKey roachpb.RKey, now hlc.Timestamp,
) ([]timeSeriesResolutionInfo, error) {
	__antithesis_instrumentation__.Notify(648097)
	var results []timeSeriesResolutionInfo

	start := storage.MakeMVCCMetadataKey(startKey.AsRawKey())
	next := storage.MakeMVCCMetadataKey(keys.TimeseriesPrefix)
	if next.Less(start) {
		__antithesis_instrumentation__.Notify(648101)
		next = start
	} else {
		__antithesis_instrumentation__.Notify(648102)
	}
	__antithesis_instrumentation__.Notify(648098)

	end := storage.MakeMVCCMetadataKey(endKey.AsRawKey())
	lastTS := storage.MakeMVCCMetadataKey(keys.TimeseriesPrefix.PrefixEnd())
	if lastTS.Less(end) {
		__antithesis_instrumentation__.Notify(648103)
		end = lastTS
	} else {
		__antithesis_instrumentation__.Notify(648104)
	}
	__antithesis_instrumentation__.Notify(648099)

	thresholds := tsdb.computeThresholds(now.WallTime)

	iter := snapshot.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{UpperBound: endKey.AsRawKey()})
	defer iter.Close()

	for iter.SeekGE(next); ; iter.SeekGE(next) {
		__antithesis_instrumentation__.Notify(648105)
		if ok, err := iter.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(648109)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(648110)
			if !ok || func() bool {
				__antithesis_instrumentation__.Notify(648111)
				return !iter.UnsafeKey().Less(end) == true
			}() == true {
				__antithesis_instrumentation__.Notify(648112)
				break
			} else {
				__antithesis_instrumentation__.Notify(648113)
			}
		}
		__antithesis_instrumentation__.Notify(648106)
		foundKey := iter.Key().Key

		name, _, res, tsNanos, err := DecodeDataKey(foundKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(648114)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(648115)
		}
		__antithesis_instrumentation__.Notify(648107)

		if threshold, ok := thresholds[res]; !ok || func() bool {
			__antithesis_instrumentation__.Notify(648116)
			return threshold > tsNanos == true
		}() == true {
			__antithesis_instrumentation__.Notify(648117)
			results = append(results, timeSeriesResolutionInfo{
				Name:       name,
				Resolution: res,
			})
		} else {
			__antithesis_instrumentation__.Notify(648118)
		}
		__antithesis_instrumentation__.Notify(648108)

		next = storage.MakeMVCCMetadataKey(makeDataKeySeriesPrefix(name, res).PrefixEnd())
	}
	__antithesis_instrumentation__.Notify(648100)

	return results, nil
}

func (tsdb *DB) pruneTimeSeries(
	ctx context.Context, db *kv.DB, timeSeriesList []timeSeriesResolutionInfo, now hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(648119)
	thresholds := tsdb.computeThresholds(now.WallTime)

	b := &kv.Batch{}
	for _, timeSeries := range timeSeriesList {
		__antithesis_instrumentation__.Notify(648121)

		start := makeDataKeySeriesPrefix(timeSeries.Name, timeSeries.Resolution)

		var end roachpb.Key
		threshold, ok := thresholds[timeSeries.Resolution]
		if ok {
			__antithesis_instrumentation__.Notify(648123)
			end = MakeDataKey(timeSeries.Name, "", timeSeries.Resolution, threshold)
		} else {
			__antithesis_instrumentation__.Notify(648124)
			end = start.PrefixEnd()
		}
		__antithesis_instrumentation__.Notify(648122)

		b.AddRawRequest(&roachpb.DeleteRangeRequest{
			RequestHeader: roachpb.RequestHeader{
				Key:    start,
				EndKey: end,
			},
			Inline: true,
		})
	}
	__antithesis_instrumentation__.Notify(648120)

	return db.Run(ctx, b)
}
