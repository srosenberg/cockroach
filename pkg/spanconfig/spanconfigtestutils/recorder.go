package spanconfigtestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type KVAccessorRecorder struct {
	underlying spanconfig.KVAccessor

	mu struct {
		syncutil.Mutex
		mutations  []mutation
		batchCount int
	}
}

var _ spanconfig.KVAccessor = &KVAccessorRecorder{}

func NewKVAccessorRecorder(underlying spanconfig.KVAccessor) *KVAccessorRecorder {
	__antithesis_instrumentation__.Notify(241705)
	return &KVAccessorRecorder{
		underlying: underlying,
	}
}

type mutation struct {
	update   spanconfig.Update
	batchIdx int
}

func (r *KVAccessorRecorder) GetSpanConfigRecords(
	ctx context.Context, targets []spanconfig.Target,
) ([]spanconfig.Record, error) {
	__antithesis_instrumentation__.Notify(241706)
	return r.underlying.GetSpanConfigRecords(ctx, targets)
}

func (r *KVAccessorRecorder) UpdateSpanConfigRecords(
	ctx context.Context,
	toDelete []spanconfig.Target,
	toUpsert []spanconfig.Record,
	minCommitTS, maxCommitTS hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(241707)
	if err := r.underlying.UpdateSpanConfigRecords(
		ctx, toDelete, toUpsert, minCommitTS, maxCommitTS,
	); err != nil {
		__antithesis_instrumentation__.Notify(241711)
		return err
	} else {
		__antithesis_instrumentation__.Notify(241712)
	}
	__antithesis_instrumentation__.Notify(241708)

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, d := range toDelete {
		__antithesis_instrumentation__.Notify(241713)
		del, err := spanconfig.Deletion(d)
		if err != nil {
			__antithesis_instrumentation__.Notify(241715)
			return err
		} else {
			__antithesis_instrumentation__.Notify(241716)
		}
		__antithesis_instrumentation__.Notify(241714)
		r.mu.mutations = append(r.mu.mutations, mutation{
			update:   del,
			batchIdx: r.mu.batchCount,
		})
	}
	__antithesis_instrumentation__.Notify(241709)
	for _, u := range toUpsert {
		__antithesis_instrumentation__.Notify(241717)
		r.mu.mutations = append(r.mu.mutations, mutation{
			update:   spanconfig.Update(u),
			batchIdx: r.mu.batchCount,
		})
	}
	__antithesis_instrumentation__.Notify(241710)
	r.mu.batchCount++
	return nil
}

func (r *KVAccessorRecorder) GetAllSystemSpanConfigsThatApply(
	ctx context.Context, id roachpb.TenantID,
) ([]roachpb.SpanConfig, error) {
	__antithesis_instrumentation__.Notify(241718)
	return r.underlying.GetAllSystemSpanConfigsThatApply(ctx, id)
}

func (r *KVAccessorRecorder) WithTxn(context.Context, *kv.Txn) spanconfig.KVAccessor {
	__antithesis_instrumentation__.Notify(241719)
	panic("unimplemented")
}

func (r *KVAccessorRecorder) Recording(clear bool) string {
	__antithesis_instrumentation__.Notify(241720)
	r.mu.Lock()
	defer r.mu.Unlock()

	sort.Slice(r.mu.mutations, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(241724)
		mi, mj := r.mu.mutations[i], r.mu.mutations[j]
		if mi.batchIdx != mj.batchIdx {
			__antithesis_instrumentation__.Notify(241727)
			return mi.batchIdx < mj.batchIdx
		} else {
			__antithesis_instrumentation__.Notify(241728)
		}
		__antithesis_instrumentation__.Notify(241725)
		if !mi.update.GetTarget().Equal(mj.update.GetTarget()) {
			__antithesis_instrumentation__.Notify(241729)
			return mi.update.GetTarget().Less(mj.update.GetTarget())
		} else {
			__antithesis_instrumentation__.Notify(241730)
		}
		__antithesis_instrumentation__.Notify(241726)

		return mi.update.Deletion()
	})
	__antithesis_instrumentation__.Notify(241721)

	var output strings.Builder
	for _, m := range r.mu.mutations {
		__antithesis_instrumentation__.Notify(241731)
		if m.update.Deletion() {
			__antithesis_instrumentation__.Notify(241732)
			output.WriteString(fmt.Sprintf("delete %s\n", m.update.GetTarget()))
		} else {
			__antithesis_instrumentation__.Notify(241733)
			switch {
			case m.update.GetTarget().IsSpanTarget():
				__antithesis_instrumentation__.Notify(241734)
				output.WriteString(fmt.Sprintf("upsert %-35s %s\n", m.update.GetTarget(),
					PrintSpanConfigDiffedAgainstDefaults(m.update.GetConfig())))
			case m.update.GetTarget().IsSystemTarget():
				__antithesis_instrumentation__.Notify(241735)
				output.WriteString(fmt.Sprintf("upsert %-35s %s\n", m.update.GetTarget(),
					PrintSystemSpanConfigDiffedAgainstDefault(m.update.GetConfig())))
			default:
				__antithesis_instrumentation__.Notify(241736)
				panic("unsupported target type")
			}
		}
	}
	__antithesis_instrumentation__.Notify(241722)

	if clear {
		__antithesis_instrumentation__.Notify(241737)
		r.mu.mutations = r.mu.mutations[:0]
		r.mu.batchCount = 0
	} else {
		__antithesis_instrumentation__.Notify(241738)
	}
	__antithesis_instrumentation__.Notify(241723)

	return output.String()
}
