package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type RowCounter struct {
	roachpb.BulkOpSummary
	prev roachpb.Key
}

func (r *RowCounter) Count(key roachpb.Key) error {
	__antithesis_instrumentation__.Notify(643668)

	row, err := keys.EnsureSafeSplitKey(key)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(643675)
		return len(key) == len(row) == true
	}() == true {
		__antithesis_instrumentation__.Notify(643676)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(643677)
	}
	__antithesis_instrumentation__.Notify(643669)

	if bytes.Equal(row, r.prev) {
		__antithesis_instrumentation__.Notify(643678)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(643679)
	}
	__antithesis_instrumentation__.Notify(643670)

	r.prev = append(r.prev[:0], row...)

	rem, _, err := keys.DecodeTenantPrefix(row)
	if err != nil {
		__antithesis_instrumentation__.Notify(643680)
		return err
	} else {
		__antithesis_instrumentation__.Notify(643681)
	}
	__antithesis_instrumentation__.Notify(643671)
	_, tableID, indexID, err := keys.DecodeTableIDIndexID(rem)
	if err != nil {
		__antithesis_instrumentation__.Notify(643682)
		return err
	} else {
		__antithesis_instrumentation__.Notify(643683)
	}
	__antithesis_instrumentation__.Notify(643672)

	if r.EntryCounts == nil {
		__antithesis_instrumentation__.Notify(643684)
		r.EntryCounts = make(map[uint64]int64)
	} else {
		__antithesis_instrumentation__.Notify(643685)
	}
	__antithesis_instrumentation__.Notify(643673)
	r.EntryCounts[roachpb.BulkOpSummaryID(uint64(tableID), uint64(indexID))]++

	if indexID == 1 {
		__antithesis_instrumentation__.Notify(643686)
		r.DeprecatedRows++
	} else {
		__antithesis_instrumentation__.Notify(643687)
		r.DeprecatedIndexEntries++
	}
	__antithesis_instrumentation__.Notify(643674)

	return nil
}
