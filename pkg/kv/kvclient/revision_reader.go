package kvclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type VersionedValues struct {
	Key    roachpb.Key
	Values []roachpb.Value
}

func GetAllRevisions(
	ctx context.Context, db *kv.DB, startKey, endKey roachpb.Key, startTime, endTime hlc.Timestamp,
) ([]VersionedValues, error) {
	__antithesis_instrumentation__.Notify(90020)

	header := roachpb.Header{Timestamp: endTime}
	req := &roachpb.ExportRequest{
		RequestHeader: roachpb.RequestHeader{Key: startKey, EndKey: endKey},
		StartTime:     startTime,
		MVCCFilter:    roachpb.MVCCFilter_All,
		ReturnSST:     true,
	}
	resp, pErr := kv.SendWrappedWith(ctx, db.NonTransactionalSender(), header, req)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(90023)
		return nil, pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(90024)
	}
	__antithesis_instrumentation__.Notify(90021)

	var res []VersionedValues
	for _, file := range resp.(*roachpb.ExportResponse).Files {
		__antithesis_instrumentation__.Notify(90025)
		iter, err := storage.NewMemSSTIterator(file.SST, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(90027)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(90028)
		}
		__antithesis_instrumentation__.Notify(90026)
		defer iter.Close()
		iter.SeekGE(storage.MVCCKey{Key: startKey})

		for ; ; iter.Next() {
			__antithesis_instrumentation__.Notify(90029)
			if valid, err := iter.Valid(); !valid || func() bool {
				__antithesis_instrumentation__.Notify(90032)
				return err != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(90033)
				if err != nil {
					__antithesis_instrumentation__.Notify(90035)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(90036)
				}
				__antithesis_instrumentation__.Notify(90034)
				break
			} else {
				__antithesis_instrumentation__.Notify(90037)
				if iter.UnsafeKey().Key.Compare(endKey) >= 0 {
					__antithesis_instrumentation__.Notify(90038)
					break
				} else {
					__antithesis_instrumentation__.Notify(90039)
				}
			}
			__antithesis_instrumentation__.Notify(90030)
			key := iter.UnsafeKey()
			keyCopy := make([]byte, len(key.Key))
			copy(keyCopy, key.Key)
			key.Key = keyCopy
			value := make([]byte, len(iter.UnsafeValue()))
			copy(value, iter.UnsafeValue())
			if len(res) == 0 || func() bool {
				__antithesis_instrumentation__.Notify(90040)
				return !res[len(res)-1].Key.Equal(key.Key) == true
			}() == true {
				__antithesis_instrumentation__.Notify(90041)
				res = append(res, VersionedValues{Key: key.Key})
			} else {
				__antithesis_instrumentation__.Notify(90042)
			}
			__antithesis_instrumentation__.Notify(90031)
			res[len(res)-1].Values = append(res[len(res)-1].Values, roachpb.Value{Timestamp: key.Timestamp, RawBytes: value})
		}
	}
	__antithesis_instrumentation__.Notify(90022)
	return res, nil
}
