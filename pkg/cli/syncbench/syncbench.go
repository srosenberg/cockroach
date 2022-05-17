package syncbench

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/codahale/hdrhistogram"
)

var numOps uint64
var numBytes uint64

const (
	minLatency = 100 * time.Microsecond
	maxLatency = 10 * time.Second
)

func clampLatency(d, min, max time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(34603)
	if d < min {
		__antithesis_instrumentation__.Notify(34606)
		return min
	} else {
		__antithesis_instrumentation__.Notify(34607)
	}
	__antithesis_instrumentation__.Notify(34604)
	if d > max {
		__antithesis_instrumentation__.Notify(34608)
		return max
	} else {
		__antithesis_instrumentation__.Notify(34609)
	}
	__antithesis_instrumentation__.Notify(34605)
	return d
}

type worker struct {
	db      storage.Engine
	latency struct {
		syncutil.Mutex
		*hdrhistogram.WindowedHistogram
	}
	logOnly bool
}

func newWorker(db storage.Engine) *worker {
	__antithesis_instrumentation__.Notify(34610)
	w := &worker{db: db}
	w.latency.WindowedHistogram = hdrhistogram.NewWindowed(1,
		minLatency.Nanoseconds(), maxLatency.Nanoseconds(), 1)
	return w
}

func (w *worker) run(wg *sync.WaitGroup) {
	__antithesis_instrumentation__.Notify(34611)
	defer wg.Done()

	ctx := context.Background()
	rand := rand.New(rand.NewSource(timeutil.Now().UnixNano()))
	var buf []byte

	randBlock := func(min, max int) []byte {
		__antithesis_instrumentation__.Notify(34613)
		data := make([]byte, rand.Intn(max-min)+min)
		for i := range data {
			__antithesis_instrumentation__.Notify(34615)
			data[i] = byte(rand.Int() & 0xff)
		}
		__antithesis_instrumentation__.Notify(34614)
		return data
	}
	__antithesis_instrumentation__.Notify(34612)

	for {
		__antithesis_instrumentation__.Notify(34616)
		start := timeutil.Now()
		b := w.db.NewBatch()
		if w.logOnly {
			__antithesis_instrumentation__.Notify(34620)
			block := randBlock(300, 400)
			if err := b.LogData(block); err != nil {
				__antithesis_instrumentation__.Notify(34621)
				log.Fatalf(ctx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(34622)
			}
		} else {
			__antithesis_instrumentation__.Notify(34623)
			for j := 0; j < 5; j++ {
				__antithesis_instrumentation__.Notify(34624)
				block := randBlock(60, 80)
				key := encoding.EncodeUint32Ascending(buf, rand.Uint32())
				if err := b.PutUnversioned(key, block); err != nil {
					__antithesis_instrumentation__.Notify(34626)
					log.Fatalf(ctx, "%v", err)
				} else {
					__antithesis_instrumentation__.Notify(34627)
				}
				__antithesis_instrumentation__.Notify(34625)
				buf = key[:0]
			}
		}
		__antithesis_instrumentation__.Notify(34617)
		bytes := uint64(b.Len())
		if err := b.Commit(true); err != nil {
			__antithesis_instrumentation__.Notify(34628)
			log.Fatalf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(34629)
		}
		__antithesis_instrumentation__.Notify(34618)
		atomic.AddUint64(&numOps, 1)
		atomic.AddUint64(&numBytes, bytes)
		elapsed := clampLatency(timeutil.Since(start), minLatency, maxLatency)
		w.latency.Lock()
		if err := w.latency.Current.RecordValue(elapsed.Nanoseconds()); err != nil {
			__antithesis_instrumentation__.Notify(34630)
			log.Fatalf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(34631)
		}
		__antithesis_instrumentation__.Notify(34619)
		w.latency.Unlock()
	}
}

type Options struct {
	Dir         string
	Concurrency int
	Duration    time.Duration
	LogOnly     bool
}

func Run(opts Options) error {
	__antithesis_instrumentation__.Notify(34632)

	_, err := os.Stat(opts.Dir)
	if err == nil {
		__antithesis_instrumentation__.Notify(34639)
		return errors.Errorf("error: supplied path '%s' must not exist", opts.Dir)
	} else {
		__antithesis_instrumentation__.Notify(34640)
	}
	__antithesis_instrumentation__.Notify(34633)

	defer func() {
		__antithesis_instrumentation__.Notify(34641)
		_ = os.RemoveAll(opts.Dir)
	}()
	__antithesis_instrumentation__.Notify(34634)

	fmt.Printf("writing to %s\n", opts.Dir)

	db, err := storage.Open(
		context.Background(),
		storage.Filesystem(opts.Dir),
		storage.CacheSize(0),
		storage.Settings(cluster.MakeTestingClusterSettings()))
	if err != nil {
		__antithesis_instrumentation__.Notify(34642)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34643)
	}
	__antithesis_instrumentation__.Notify(34635)

	workers := make([]*worker, opts.Concurrency)

	var wg sync.WaitGroup
	for i := range workers {
		__antithesis_instrumentation__.Notify(34644)
		wg.Add(1)
		workers[i] = newWorker(db)
		workers[i].logOnly = opts.LogOnly
		go workers[i].run(&wg)
	}
	__antithesis_instrumentation__.Notify(34636)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	done := make(chan os.Signal, 3)
	signal.Notify(done, os.Interrupt)

	go func() {
		__antithesis_instrumentation__.Notify(34645)
		wg.Wait()
		done <- sysutil.Signal(0)
	}()
	__antithesis_instrumentation__.Notify(34637)

	if opts.Duration > 0 {
		__antithesis_instrumentation__.Notify(34646)
		go func() {
			__antithesis_instrumentation__.Notify(34647)
			time.Sleep(opts.Duration)
			done <- sysutil.Signal(0)
		}()
	} else {
		__antithesis_instrumentation__.Notify(34648)
	}
	__antithesis_instrumentation__.Notify(34638)

	start := timeutil.Now()
	lastNow := start
	var lastOps uint64
	var lastBytes uint64

	for i := 0; ; i++ {
		__antithesis_instrumentation__.Notify(34649)
		select {
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(34650)
			var h *hdrhistogram.Histogram
			for _, w := range workers {
				__antithesis_instrumentation__.Notify(34654)
				w.latency.Lock()
				m := w.latency.Merge()
				w.latency.Rotate()
				w.latency.Unlock()
				if h == nil {
					__antithesis_instrumentation__.Notify(34655)
					h = m
				} else {
					__antithesis_instrumentation__.Notify(34656)
					h.Merge(m)
				}
			}
			__antithesis_instrumentation__.Notify(34651)

			p50 := h.ValueAtQuantile(50)
			p95 := h.ValueAtQuantile(95)
			p99 := h.ValueAtQuantile(99)
			pMax := h.ValueAtQuantile(100)

			now := timeutil.Now()
			elapsed := now.Sub(lastNow)
			ops := atomic.LoadUint64(&numOps)
			bytes := atomic.LoadUint64(&numBytes)

			if i%20 == 0 {
				__antithesis_instrumentation__.Notify(34657)
				fmt.Println("_elapsed____ops/sec___mb/sec__p50(ms)__p95(ms)__p99(ms)_pMax(ms)")
			} else {
				__antithesis_instrumentation__.Notify(34658)
			}
			__antithesis_instrumentation__.Notify(34652)
			fmt.Printf("%8s %10.1f %8.1f %8.1f %8.1f %8.1f %8.1f\n",
				time.Duration(timeutil.Since(start).Seconds()+0.5)*time.Second,
				float64(ops-lastOps)/elapsed.Seconds(),
				float64(bytes-lastBytes)/(1024.0*1024.0)/elapsed.Seconds(),
				time.Duration(p50).Seconds()*1000,
				time.Duration(p95).Seconds()*1000,
				time.Duration(p99).Seconds()*1000,
				time.Duration(pMax).Seconds()*1000,
			)
			lastNow = now
			lastOps = ops
			lastBytes = bytes

		case <-done:
			__antithesis_instrumentation__.Notify(34653)
			return nil
		}
	}
}
