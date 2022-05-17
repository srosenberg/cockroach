package txnrecovery

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
	"github.com/cockroachdb/errors"
)

type Manager interface {
	ResolveIndeterminateCommit(
		context.Context, *roachpb.IndeterminateCommitError,
	) (*roachpb.Transaction, error)

	Metrics() Metrics
}

const (
	defaultTaskLimit = 1024

	defaultBatchSize = 128
)

type manager struct {
	log.AmbientContext

	clock   *hlc.Clock
	db      *kv.DB
	stopper *stop.Stopper
	metrics Metrics
	txns    singleflight.Group
	sem     chan struct{}
}

func NewManager(ac log.AmbientContext, clock *hlc.Clock, db *kv.DB, stopper *stop.Stopper) Manager {
	__antithesis_instrumentation__.Notify(127224)
	ac.AddLogTag("txn-recovery", nil)
	return &manager{
		AmbientContext: ac,
		clock:          clock,
		db:             db,
		stopper:        stopper,
		metrics:        makeMetrics(),
		sem:            make(chan struct{}, defaultTaskLimit),
	}
}

func (m *manager) ResolveIndeterminateCommit(
	ctx context.Context, ice *roachpb.IndeterminateCommitError,
) (*roachpb.Transaction, error) {
	__antithesis_instrumentation__.Notify(127225)
	txn := &ice.StagingTxn
	if txn.Status != roachpb.STAGING {
		__antithesis_instrumentation__.Notify(127228)
		return nil, errors.Errorf("IndeterminateCommitError with non-STAGING transaction: %v", txn)
	} else {
		__antithesis_instrumentation__.Notify(127229)
	}
	__antithesis_instrumentation__.Notify(127226)

	log.VEventf(ctx, 2, "recovering txn %s from indeterminate commit", txn.ID.Short())
	resC, _ := m.txns.DoChan(txn.ID.String(), func() (interface{}, error) {
		__antithesis_instrumentation__.Notify(127230)
		return m.resolveIndeterminateCommitForTxn(txn)
	})
	__antithesis_instrumentation__.Notify(127227)

	select {
	case res := <-resC:
		__antithesis_instrumentation__.Notify(127231)
		if res.Err != nil {
			__antithesis_instrumentation__.Notify(127234)
			log.VEventf(ctx, 2, "recovery error: %v", res.Err)
			return nil, res.Err
		} else {
			__antithesis_instrumentation__.Notify(127235)
		}
		__antithesis_instrumentation__.Notify(127232)
		txn := res.Val.(*roachpb.Transaction)
		log.VEventf(ctx, 2, "recovered txn %s with status: %s", txn.ID.Short(), txn.Status)
		return txn, nil
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(127233)
		return nil, errors.Wrap(ctx.Err(), "abandoned indeterminate commit recovery")
	}
}

func (m *manager) resolveIndeterminateCommitForTxn(
	txn *roachpb.Transaction,
) (resTxn *roachpb.Transaction, resErr error) {
	__antithesis_instrumentation__.Notify(127236)

	onComplete := m.updateMetrics()
	defer func() { __antithesis_instrumentation__.Notify(127239); onComplete(resTxn, resErr) }()
	__antithesis_instrumentation__.Notify(127237)

	ctx := m.AnnotateCtx(context.Background())

	resErr = m.stopper.RunTaskWithErr(ctx,
		"recovery.manager: resolving indeterminate commit",
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(127240)

			select {
			case m.sem <- struct{}{}:
				__antithesis_instrumentation__.Notify(127244)
				defer func() { __antithesis_instrumentation__.Notify(127246); <-m.sem }()
			case <-m.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(127245)
				return stop.ErrUnavailable
			}
			__antithesis_instrumentation__.Notify(127241)

			preventedIntent, changedTxn, err := m.resolveIndeterminateCommitForTxnProbe(ctx, txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(127247)
				return err
			} else {
				__antithesis_instrumentation__.Notify(127248)
			}
			__antithesis_instrumentation__.Notify(127242)
			if changedTxn != nil {
				__antithesis_instrumentation__.Notify(127249)
				resTxn = changedTxn
				return nil
			} else {
				__antithesis_instrumentation__.Notify(127250)
			}
			__antithesis_instrumentation__.Notify(127243)

			resTxn, err = m.resolveIndeterminateCommitForTxnRecover(ctx, txn, preventedIntent)
			return err
		},
	)
	__antithesis_instrumentation__.Notify(127238)
	return resTxn, resErr
}

func (m *manager) resolveIndeterminateCommitForTxnProbe(
	ctx context.Context, txn *roachpb.Transaction,
) (preventedIntent bool, changedTxn *roachpb.Transaction, err error) {
	__antithesis_instrumentation__.Notify(127251)

	queryTxnReq := roachpb.QueryTxnRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: txn.Key,
		},
		Txn:           txn.TxnMeta,
		WaitForUpdate: false,
	}

	queryIntentReqs := make([]roachpb.QueryIntentRequest, 0, len(txn.InFlightWrites))
	for _, w := range txn.InFlightWrites {
		__antithesis_instrumentation__.Notify(127255)
		meta := txn.TxnMeta
		meta.Sequence = w.Sequence
		queryIntentReqs = append(queryIntentReqs, roachpb.QueryIntentRequest{
			RequestHeader: roachpb.RequestHeader{
				Key: w.Key,
			},
			Txn: meta,
		})
	}
	__antithesis_instrumentation__.Notify(127252)

	sort.Slice(queryIntentReqs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(127256)
		return queryIntentReqs[i].Header().Key.Compare(queryIntentReqs[j].Header().Key) < 0
	})
	__antithesis_instrumentation__.Notify(127253)

	for len(queryIntentReqs) > 0 {
		__antithesis_instrumentation__.Notify(127257)
		var b kv.Batch
		b.Header.Timestamp = m.batchTimestamp(txn)
		b.AddRawRequest(&queryTxnReq)
		for i := 0; i < defaultBatchSize && func() bool {
			__antithesis_instrumentation__.Notify(127261)
			return len(queryIntentReqs) > 0 == true
		}() == true; i++ {
			__antithesis_instrumentation__.Notify(127262)
			b.AddRawRequest(&queryIntentReqs[0])
			queryIntentReqs = queryIntentReqs[1:]
		}
		__antithesis_instrumentation__.Notify(127258)

		if err := m.db.Run(ctx, &b); err != nil {
			__antithesis_instrumentation__.Notify(127263)

			return false, nil, err
		} else {
			__antithesis_instrumentation__.Notify(127264)
		}
		__antithesis_instrumentation__.Notify(127259)

		resps := b.RawResponse().Responses
		queryTxnResp := resps[0].GetInner().(*roachpb.QueryTxnResponse)
		queriedTxn := &queryTxnResp.QueriedTxn
		if queriedTxn.Status.IsFinalized() || func() bool {
			__antithesis_instrumentation__.Notify(127265)
			return txn.Epoch < queriedTxn.Epoch == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(127266)
			return txn.WriteTimestamp.Less(queriedTxn.WriteTimestamp) == true
		}() == true {
			__antithesis_instrumentation__.Notify(127267)

			return false, queriedTxn, nil
		} else {
			__antithesis_instrumentation__.Notify(127268)
		}
		__antithesis_instrumentation__.Notify(127260)

		for _, ru := range resps[1:] {
			__antithesis_instrumentation__.Notify(127269)
			queryIntentResp := ru.GetInner().(*roachpb.QueryIntentResponse)
			if !queryIntentResp.FoundIntent {
				__antithesis_instrumentation__.Notify(127270)
				return true, nil, nil
			} else {
				__antithesis_instrumentation__.Notify(127271)
			}
		}
	}
	__antithesis_instrumentation__.Notify(127254)
	return false, nil, nil
}

func (m *manager) resolveIndeterminateCommitForTxnRecover(
	ctx context.Context, txn *roachpb.Transaction, preventedIntent bool,
) (*roachpb.Transaction, error) {
	__antithesis_instrumentation__.Notify(127272)
	var b kv.Batch
	b.Header.Timestamp = m.batchTimestamp(txn)
	b.AddRawRequest(&roachpb.RecoverTxnRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: txn.Key,
		},
		Txn:                 txn.TxnMeta,
		ImplicitlyCommitted: !preventedIntent,
	})

	if err := m.db.Run(ctx, &b); err != nil {
		__antithesis_instrumentation__.Notify(127274)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(127275)
	}
	__antithesis_instrumentation__.Notify(127273)

	resps := b.RawResponse().Responses
	recTxnResp := resps[0].GetInner().(*roachpb.RecoverTxnResponse)
	return &recTxnResp.RecoveredTxn, nil
}

func (m *manager) batchTimestamp(txn *roachpb.Transaction) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127276)
	now := m.clock.Now()
	now.Forward(txn.WriteTimestamp)
	return now
}

func (m *manager) Metrics() Metrics {
	__antithesis_instrumentation__.Notify(127277)
	return m.metrics
}

func (m *manager) updateMetrics() func(*roachpb.Transaction, error) {
	__antithesis_instrumentation__.Notify(127278)
	m.metrics.AttemptsPending.Inc(1)
	m.metrics.Attempts.Inc(1)
	return func(txn *roachpb.Transaction, err error) {
		__antithesis_instrumentation__.Notify(127279)
		m.metrics.AttemptsPending.Dec(1)
		if err != nil {
			__antithesis_instrumentation__.Notify(127280)
			m.metrics.Failures.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(127281)
			switch txn.Status {
			case roachpb.COMMITTED:
				__antithesis_instrumentation__.Notify(127282)
				m.metrics.SuccessesAsCommitted.Inc(1)
			case roachpb.ABORTED:
				__antithesis_instrumentation__.Notify(127283)
				m.metrics.SuccessesAsAborted.Inc(1)
			case roachpb.PENDING, roachpb.STAGING:
				__antithesis_instrumentation__.Notify(127284)
				m.metrics.SuccessesAsPending.Inc(1)
			default:
				__antithesis_instrumentation__.Notify(127285)
				panic("unexpected")
			}
		}
	}
}
