package histogram

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/codahale/hdrhistogram"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	PrometheusNamespace = "workload"

	MockWorkloadName = "mock"

	sigFigs    = 1
	minLatency = 100 * time.Microsecond
)

type NamedHistogram struct {
	name                string
	prometheusHistogram prometheus.Histogram
	mu                  struct {
		syncutil.Mutex
		current *hdrhistogram.Histogram
	}
}

func (w *Registry) newNamedHistogramLocked(name string) *NamedHistogram {
	__antithesis_instrumentation__.Notify(694152)
	hist := &NamedHistogram{
		name:                name,
		prometheusHistogram: w.getPrometheusHistogramLocked(name),
	}
	hist.mu.current = w.newHistogram()
	return hist
}

func (w *NamedHistogram) Record(elapsed time.Duration) {
	__antithesis_instrumentation__.Notify(694153)
	w.prometheusHistogram.Observe(float64(elapsed.Nanoseconds()) / float64(time.Second))
	maxLatency := time.Duration(w.mu.current.HighestTrackableValue())
	if elapsed < minLatency {
		__antithesis_instrumentation__.Notify(694155)
		elapsed = minLatency
	} else {
		__antithesis_instrumentation__.Notify(694156)
		if elapsed > maxLatency {
			__antithesis_instrumentation__.Notify(694157)
			elapsed = maxLatency
		} else {
			__antithesis_instrumentation__.Notify(694158)
		}
	}
	__antithesis_instrumentation__.Notify(694154)

	w.mu.Lock()
	err := w.mu.current.RecordValue(elapsed.Nanoseconds())
	w.mu.Unlock()

	if err != nil {
		__antithesis_instrumentation__.Notify(694159)

		panic(fmt.Sprintf(`%s: recording value: %s`, w.name, err))
	} else {
		__antithesis_instrumentation__.Notify(694160)
	}
}

func (w *NamedHistogram) tick(
	newHistogram *hdrhistogram.Histogram, fn func(h *hdrhistogram.Histogram),
) {
	__antithesis_instrumentation__.Notify(694161)
	w.mu.Lock()
	h := w.mu.current
	w.mu.current = newHistogram
	w.mu.Unlock()
	fn(h)
}

type Registry struct {
	workloadName string
	promReg      *prometheus.Registry
	mu           struct {
		syncutil.Mutex

		registered map[string][]*NamedHistogram

		prometheusHistograms map[string]prometheus.Histogram
	}

	start         time.Time
	cumulative    map[string]*hdrhistogram.Histogram
	prevTick      map[string]time.Time
	histogramPool *sync.Pool
}

func NewRegistry(maxLat time.Duration, workloadName string) *Registry {
	__antithesis_instrumentation__.Notify(694162)
	r := &Registry{
		workloadName: workloadName,
		start:        timeutil.Now(),
		cumulative:   make(map[string]*hdrhistogram.Histogram),
		prevTick:     make(map[string]time.Time),
		promReg:      prometheus.NewRegistry(),
		histogramPool: &sync.Pool{
			New: func() interface{} {
				__antithesis_instrumentation__.Notify(694164)
				return hdrhistogram.New(minLatency.Nanoseconds(), maxLat.Nanoseconds(), sigFigs)
			},
		},
	}
	__antithesis_instrumentation__.Notify(694163)
	r.mu.registered = make(map[string][]*NamedHistogram)
	r.mu.prometheusHistograms = make(map[string]prometheus.Histogram)
	return r
}

func (w *Registry) Registerer() prometheus.Registerer {
	__antithesis_instrumentation__.Notify(694165)
	return w.promReg
}

func (w *Registry) Gatherer() prometheus.Gatherer {
	__antithesis_instrumentation__.Notify(694166)
	return w.promReg
}

func (w *Registry) newHistogram() *hdrhistogram.Histogram {
	__antithesis_instrumentation__.Notify(694167)
	h := w.histogramPool.Get().(*hdrhistogram.Histogram)
	return h
}

var invalidPrometheusMetricRe = regexp.MustCompile(`[^a-zA-Z0-9:_]`)

func cleanPrometheusName(name string) string {
	__antithesis_instrumentation__.Notify(694168)
	return invalidPrometheusMetricRe.ReplaceAllString(name, "_")
}

func makePrometheusLatencyHistogramBuckets() []float64 {
	__antithesis_instrumentation__.Notify(694169)

	return prometheus.ExponentialBuckets(0.0005, 1.1, 150)
}

func (w *Registry) getPrometheusHistogramLocked(name string) prometheus.Histogram {
	__antithesis_instrumentation__.Notify(694170)
	ph, ok := w.mu.prometheusHistograms[name]

	if !ok {
		__antithesis_instrumentation__.Notify(694172)

		promName := cleanPrometheusName(name) + "_duration_seconds"
		ph = promauto.With(w.promReg).NewHistogram(prometheus.HistogramOpts{
			Namespace: PrometheusNamespace,
			Subsystem: cleanPrometheusName(w.workloadName),
			Name:      promName,
			Buckets:   makePrometheusLatencyHistogramBuckets(),
		})
		w.mu.prometheusHistograms[name] = ph
	} else {
		__antithesis_instrumentation__.Notify(694173)
	}
	__antithesis_instrumentation__.Notify(694171)

	return ph
}

func (w *Registry) GetHandle() *Histograms {
	__antithesis_instrumentation__.Notify(694174)
	hists := &Histograms{
		reg: w,
	}
	hists.mu.hists = make(map[string]*NamedHistogram)
	return hists
}

func (w *Registry) Tick(fn func(Tick)) {
	__antithesis_instrumentation__.Notify(694175)
	merged := make(map[string]*hdrhistogram.Histogram)
	var names []string
	var wg sync.WaitGroup

	w.mu.Lock()
	for name, nameRegistered := range w.mu.registered {
		__antithesis_instrumentation__.Notify(694177)
		wg.Add(1)
		registered := append([]*NamedHistogram(nil), nameRegistered...)
		merged[name] = w.newHistogram()
		names = append(names, name)
		go func(registered []*NamedHistogram, merged *hdrhistogram.Histogram) {
			__antithesis_instrumentation__.Notify(694178)
			for _, hist := range registered {
				__antithesis_instrumentation__.Notify(694180)
				hist.tick(w.newHistogram(), func(h *hdrhistogram.Histogram) {
					__antithesis_instrumentation__.Notify(694181)
					merged.Merge(h)
					h.Reset()
					w.histogramPool.Put(h)
				})
			}
			__antithesis_instrumentation__.Notify(694179)
			wg.Done()
		}(registered, merged[name])
	}
	__antithesis_instrumentation__.Notify(694176)
	w.mu.Unlock()

	wg.Wait()

	now := timeutil.Now()
	sort.Strings(names)
	for _, name := range names {
		__antithesis_instrumentation__.Notify(694182)
		mergedHist := merged[name]
		if _, ok := w.cumulative[name]; !ok {
			__antithesis_instrumentation__.Notify(694185)
			w.cumulative[name] = w.newHistogram()
		} else {
			__antithesis_instrumentation__.Notify(694186)
		}
		__antithesis_instrumentation__.Notify(694183)
		w.cumulative[name].Merge(mergedHist)

		prevTick, ok := w.prevTick[name]
		if !ok {
			__antithesis_instrumentation__.Notify(694187)
			prevTick = w.start
		} else {
			__antithesis_instrumentation__.Notify(694188)
		}
		__antithesis_instrumentation__.Notify(694184)
		w.prevTick[name] = now
		fn(Tick{
			Name:       name,
			Hist:       mergedHist,
			Cumulative: w.cumulative[name],
			Elapsed:    now.Sub(prevTick),
			Now:        now,
		})
		mergedHist.Reset()
		w.histogramPool.Put(mergedHist)
	}
}

type Histograms struct {
	reg *Registry
	mu  struct {
		syncutil.RWMutex
		hists map[string]*NamedHistogram
	}
}

func (w *Histograms) Get(name string) *NamedHistogram {
	__antithesis_instrumentation__.Notify(694189)

	w.mu.RLock()
	hist, ok := w.mu.hists[name]
	if ok {
		__antithesis_instrumentation__.Notify(694192)
		w.mu.RUnlock()
		return hist
	} else {
		__antithesis_instrumentation__.Notify(694193)
	}
	__antithesis_instrumentation__.Notify(694190)
	w.mu.RUnlock()

	w.mu.Lock()
	defer w.mu.Unlock()
	w.reg.mu.Lock()
	defer w.reg.mu.Unlock()

	hist, ok = w.mu.hists[name]
	if !ok {
		__antithesis_instrumentation__.Notify(694194)
		hist = w.reg.newNamedHistogramLocked(name)
		w.mu.hists[name] = hist
		w.reg.mu.registered[name] = append(w.reg.mu.registered[name], hist)
	} else {
		__antithesis_instrumentation__.Notify(694195)
	}
	__antithesis_instrumentation__.Notify(694191)

	return hist
}

func Copy(h *hdrhistogram.Histogram) *hdrhistogram.Histogram {
	__antithesis_instrumentation__.Notify(694196)
	dup := hdrhistogram.New(h.LowestTrackableValue(), h.HighestTrackableValue(),
		int(h.SignificantFigures()))
	dup.Merge(h)
	return dup
}

type Tick struct {
	Name string

	Hist *hdrhistogram.Histogram

	Cumulative *hdrhistogram.Histogram

	Elapsed time.Duration

	Now time.Time
}

func (t Tick) Snapshot() SnapshotTick {
	__antithesis_instrumentation__.Notify(694197)
	return SnapshotTick{
		Name:    t.Name,
		Elapsed: t.Elapsed,
		Now:     t.Now,
		Hist:    t.Hist.Export(),
	}
}

type SnapshotTick struct {
	Name    string
	Hist    *hdrhistogram.Snapshot
	Elapsed time.Duration
	Now     time.Time
}

func DecodeSnapshots(path string) (map[string][]SnapshotTick, error) {
	__antithesis_instrumentation__.Notify(694198)
	f, err := os.Open(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(694202)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(694203)
	}
	__antithesis_instrumentation__.Notify(694199)
	defer func() { __antithesis_instrumentation__.Notify(694204); _ = f.Close() }()
	__antithesis_instrumentation__.Notify(694200)
	dec := json.NewDecoder(f)
	ret := make(map[string][]SnapshotTick)
	for {
		__antithesis_instrumentation__.Notify(694205)
		var tick SnapshotTick
		if err := dec.Decode(&tick); err == io.EOF {
			__antithesis_instrumentation__.Notify(694207)
			break
		} else {
			__antithesis_instrumentation__.Notify(694208)
			if err != nil {
				__antithesis_instrumentation__.Notify(694209)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(694210)
			}
		}
		__antithesis_instrumentation__.Notify(694206)
		ret[tick.Name] = append(ret[tick.Name], tick)
	}
	__antithesis_instrumentation__.Notify(694201)
	return ret, nil
}
