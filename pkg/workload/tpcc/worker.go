package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/exp/rand"
)

const (
	NumWorkersPerWarehouse = 10
	numConnsPerWarehouse   = 2
)

type tpccTx interface {
	run(ctx context.Context, wID int) (interface{}, error)
}

type createTxFn func(ctx context.Context, config *tpcc, mcp *workload.MultiConnPool) (tpccTx, error)

type txInfo struct {
	name        string
	constructor createTxFn
	keyingTime  int
	thinkTime   float64
	weight      int
}

var allTxs = [...]txInfo{
	{
		name:        "newOrder",
		constructor: createNewOrder,
		keyingTime:  18,
		thinkTime:   12,
	},
	{
		name:        "payment",
		constructor: createPayment,
		keyingTime:  3,
		thinkTime:   12,
	},
	{
		name:        "orderStatus",
		constructor: createOrderStatus,
		keyingTime:  2,
		thinkTime:   10,
	},
	{
		name:        "delivery",
		constructor: createDelivery,
		keyingTime:  2,
		thinkTime:   5,
	},
	{
		name:        "stockLevel",
		constructor: createStockLevel,
		keyingTime:  2,
		thinkTime:   5,
	},
}

type txCounter struct {
	success, error prometheus.Counter
}

type txCounters map[string]txCounter

func setupTPCCMetrics(reg prometheus.Registerer) txCounters {
	__antithesis_instrumentation__.Notify(698597)
	m := txCounters{}
	f := promauto.With(reg)
	for _, tx := range allTxs {
		__antithesis_instrumentation__.Notify(698599)
		m[tx.name] = txCounter{
			success: f.NewCounter(
				prometheus.CounterOpts{
					Namespace: histogram.PrometheusNamespace,
					Subsystem: tpccMeta.Name,
					Name:      fmt.Sprintf("%s_success_total", tx.name),
					Help:      fmt.Sprintf("The total number of successful %s transactions.", tx.name),
				},
			),
			error: f.NewCounter(
				prometheus.CounterOpts{
					Namespace: histogram.PrometheusNamespace,
					Subsystem: tpccMeta.Name,
					Name:      fmt.Sprintf("%s_error_total", tx.name),
					Help:      fmt.Sprintf("The total number of error %s transactions.", tx.name),
				}),
		}
	}
	__antithesis_instrumentation__.Notify(698598)
	return m
}

func initializeMix(config *tpcc) error {
	__antithesis_instrumentation__.Notify(698600)
	config.txInfos = append([]txInfo(nil), allTxs[0:]...)
	nameToTx := make(map[string]int)
	for i, tx := range config.txInfos {
		__antithesis_instrumentation__.Notify(698604)
		nameToTx[tx.name] = i
	}
	__antithesis_instrumentation__.Notify(698601)

	items := strings.Split(config.mix, `,`)
	totalWeight := 0
	for _, item := range items {
		__antithesis_instrumentation__.Notify(698605)
		kv := strings.Split(item, `=`)
		if len(kv) != 2 {
			__antithesis_instrumentation__.Notify(698609)
			return errors.Errorf(`Invalid mix %s: %s is not a k=v pair`, config.mix, item)
		} else {
			__antithesis_instrumentation__.Notify(698610)
		}
		__antithesis_instrumentation__.Notify(698606)
		txName, weightStr := kv[0], kv[1]

		weight, err := strconv.Atoi(weightStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(698611)
			return errors.Errorf(
				`Invalid percentage mix %s: %s is not an integer`, config.mix, weightStr)
		} else {
			__antithesis_instrumentation__.Notify(698612)
		}
		__antithesis_instrumentation__.Notify(698607)

		i, ok := nameToTx[txName]
		if !ok {
			__antithesis_instrumentation__.Notify(698613)
			return errors.Errorf(
				`Invalid percentage mix %s: no such transaction %s`, config.mix, txName)
		} else {
			__antithesis_instrumentation__.Notify(698614)
		}
		__antithesis_instrumentation__.Notify(698608)

		config.txInfos[i].weight = weight
		totalWeight += weight
	}
	__antithesis_instrumentation__.Notify(698602)

	config.deck = make([]int, 0, totalWeight)
	for i, t := range config.txInfos {
		__antithesis_instrumentation__.Notify(698615)
		for j := 0; j < t.weight; j++ {
			__antithesis_instrumentation__.Notify(698616)
			config.deck = append(config.deck, i)
		}
	}
	__antithesis_instrumentation__.Notify(698603)

	return nil
}

type worker struct {
	config *tpcc

	txs       []tpccTx
	hists     *histogram.Histograms
	warehouse int

	deckPerm []int
	permIdx  int

	counters txCounters
}

func newWorker(
	ctx context.Context,
	config *tpcc,
	mcp *workload.MultiConnPool,
	hists *histogram.Histograms,
	counters txCounters,
	warehouse int,
) (*worker, error) {
	__antithesis_instrumentation__.Notify(698617)
	w := &worker{
		config:    config,
		txs:       make([]tpccTx, len(config.txInfos)),
		hists:     hists,
		warehouse: warehouse,
		deckPerm:  append([]int(nil), config.deck...),
		permIdx:   len(config.deck),
		counters:  counters,
	}
	for i := range w.txs {
		__antithesis_instrumentation__.Notify(698619)
		var err error
		w.txs[i], err = config.txInfos[i].constructor(ctx, config, mcp)
		if err != nil {
			__antithesis_instrumentation__.Notify(698620)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(698621)
		}
	}
	__antithesis_instrumentation__.Notify(698618)
	return w, nil
}

func (w *worker) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(698622)

	if w.permIdx == len(w.deckPerm) {
		__antithesis_instrumentation__.Notify(698627)
		rand.Shuffle(len(w.deckPerm), func(i, j int) {
			__antithesis_instrumentation__.Notify(698629)
			w.deckPerm[i], w.deckPerm[j] = w.deckPerm[j], w.deckPerm[i]
		})
		__antithesis_instrumentation__.Notify(698628)
		w.permIdx = 0
	} else {
		__antithesis_instrumentation__.Notify(698630)
	}
	__antithesis_instrumentation__.Notify(698623)

	opIdx := w.deckPerm[w.permIdx]
	txInfo := &w.config.txInfos[opIdx]
	tx := w.txs[opIdx]
	w.permIdx++

	warehouseID := w.warehouse

	time.Sleep(time.Duration(float64(txInfo.keyingTime) * float64(time.Second) * w.config.waitFraction))

	start := timeutil.Now()
	if _, err := tx.run(context.Background(), warehouseID); err != nil {
		__antithesis_instrumentation__.Notify(698631)
		w.counters[txInfo.name].error.Inc()
		return errors.Wrapf(err, "error in %s", txInfo.name)
	} else {
		__antithesis_instrumentation__.Notify(698632)
	}
	__antithesis_instrumentation__.Notify(698624)
	if ctx.Err() == nil {
		__antithesis_instrumentation__.Notify(698633)
		elapsed := timeutil.Since(start)

		w.hists.Get(txInfo.name).Record(elapsed)
	} else {
		__antithesis_instrumentation__.Notify(698634)
	}
	__antithesis_instrumentation__.Notify(698625)
	w.counters[txInfo.name].success.Inc()

	thinkTime := -math.Log(rand.Float64()) * txInfo.thinkTime
	if thinkTime > (txInfo.thinkTime * 10) {
		__antithesis_instrumentation__.Notify(698635)
		thinkTime = txInfo.thinkTime * 10
	} else {
		__antithesis_instrumentation__.Notify(698636)
	}
	__antithesis_instrumentation__.Notify(698626)
	time.Sleep(time.Duration(thinkTime * float64(time.Second) * w.config.waitFraction))
	return ctx.Err()
}
