// Package ptreconcile provides logic to reconcile protected timestamp records
// with state associated with their metadata.
package ptreconcile

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var ReconcileInterval = settings.RegisterDurationSetting(
	settings.TenantReadOnly,
	"kv.protectedts.reconciliation.interval",
	"the frequency for reconciling jobs with protected timestamp records",
	5*time.Minute,
	settings.NonNegativeDuration,
).WithPublic()

type StatusFunc func(
	ctx context.Context, txn *kv.Txn, meta []byte,
) (shouldRemove bool, _ error)

type StatusFuncs map[string]StatusFunc

type Reconciler struct {
	settings    *cluster.Settings
	db          *kv.DB
	pts         protectedts.Storage
	metrics     Metrics
	statusFuncs StatusFuncs
}

func New(
	st *cluster.Settings, db *kv.DB, storage protectedts.Storage, statusFuncs StatusFuncs,
) *Reconciler {
	__antithesis_instrumentation__.Notify(111750)
	return &Reconciler{
		settings:    st,
		db:          db,
		pts:         storage,
		metrics:     makeMetrics(),
		statusFuncs: statusFuncs,
	}
}

func (r *Reconciler) StartReconciler(ctx context.Context, stopper *stop.Stopper) error {
	__antithesis_instrumentation__.Notify(111751)
	return stopper.RunAsyncTask(ctx, "protectedts-reconciliation", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(111752)
		r.run(ctx, stopper)
	})
}

func (r *Reconciler) Metrics() *Metrics {
	__antithesis_instrumentation__.Notify(111753)
	return &r.metrics
}

func (r *Reconciler) run(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(111754)
	reconcileIntervalChanged := make(chan struct{}, 1)
	ReconcileInterval.SetOnChange(&r.settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(111757)
		select {
		case reconcileIntervalChanged <- struct{}{}:
			__antithesis_instrumentation__.Notify(111758)
		default:
			__antithesis_instrumentation__.Notify(111759)
		}
	})
	__antithesis_instrumentation__.Notify(111755)
	lastReconciled := time.Time{}
	getInterval := func() time.Duration {
		__antithesis_instrumentation__.Notify(111760)
		interval := ReconcileInterval.Get(&r.settings.SV)
		const jitterFrac = .1
		return time.Duration(float64(interval) * (1 + (rand.Float64()-.5)*jitterFrac))
	}
	__antithesis_instrumentation__.Notify(111756)
	timer := timeutil.NewTimer()
	for {
		__antithesis_instrumentation__.Notify(111761)
		timer.Reset(timeutil.Until(lastReconciled.Add(getInterval())))
		select {
		case <-timer.C:
			__antithesis_instrumentation__.Notify(111762)
			timer.Read = true
			r.reconcile(ctx)
			lastReconciled = timeutil.Now()
		case <-reconcileIntervalChanged:
			__antithesis_instrumentation__.Notify(111763)

		case <-stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(111764)
			return
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(111765)
			return
		}
	}
}

func (r *Reconciler) reconcile(ctx context.Context) {
	__antithesis_instrumentation__.Notify(111766)

	var state ptpb.State
	if err := r.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(111769)
		var err error
		state, err = r.pts.GetState(ctx, txn)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(111770)
		r.metrics.ReconciliationErrors.Inc(1)
		log.Errorf(ctx, "failed to load protected timestamp records: %+v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(111771)
	}
	__antithesis_instrumentation__.Notify(111767)
	for _, rec := range state.Records {
		__antithesis_instrumentation__.Notify(111772)
		task, ok := r.statusFuncs[rec.MetaType]
		if !ok {
			__antithesis_instrumentation__.Notify(111774)

			continue
		} else {
			__antithesis_instrumentation__.Notify(111775)
		}
		__antithesis_instrumentation__.Notify(111773)
		var didRemove bool
		if err := r.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) (err error) {
			__antithesis_instrumentation__.Notify(111776)
			didRemove = false
			shouldRemove, err := task(ctx, txn, rec.Meta)
			if err != nil {
				__antithesis_instrumentation__.Notify(111780)
				return err
			} else {
				__antithesis_instrumentation__.Notify(111781)
			}
			__antithesis_instrumentation__.Notify(111777)
			if !shouldRemove {
				__antithesis_instrumentation__.Notify(111782)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(111783)
			}
			__antithesis_instrumentation__.Notify(111778)
			err = r.pts.Release(ctx, txn, rec.ID.GetUUID())
			if err != nil && func() bool {
				__antithesis_instrumentation__.Notify(111784)
				return !errors.Is(err, protectedts.ErrNotExists) == true
			}() == true {
				__antithesis_instrumentation__.Notify(111785)
				return err
			} else {
				__antithesis_instrumentation__.Notify(111786)
			}
			__antithesis_instrumentation__.Notify(111779)
			didRemove = true
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(111787)
			r.metrics.ReconciliationErrors.Inc(1)
			log.Errorf(ctx, "failed to reconcile protected timestamp with id %s: %v",
				rec.ID.String(), err)
		} else {
			__antithesis_instrumentation__.Notify(111788)
			r.metrics.RecordsProcessed.Inc(1)
			if didRemove {
				__antithesis_instrumentation__.Notify(111789)
				r.metrics.RecordsRemoved.Inc(1)
			} else {
				__antithesis_instrumentation__.Notify(111790)
			}
		}
	}
	__antithesis_instrumentation__.Notify(111768)
	r.metrics.ReconcilationRuns.Inc(1)
}
