package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/codahale/hdrhistogram"
)

const SpecWarehouseFactor = 12.86

const DeckWarehouseFactor = 12.605

const PassingEfficiency = 85.0

var passing90ThPercentile = map[string]time.Duration{
	"newOrder":    5 * time.Second,
	"payment":     5 * time.Second,
	"orderStatus": 5 * time.Second,
	"delivery":    5 * time.Second,
	"stockLevel":  20 * time.Second,
}

type Result struct {
	ActiveWarehouses int

	Cumulative map[string]*hdrhistogram.Histogram

	Elapsed time.Duration

	WarehouseFactor float64
}

func MergeResults(results ...*Result) *Result {
	__antithesis_instrumentation__.Notify(698253)
	if len(results) == 0 {
		__antithesis_instrumentation__.Notify(698256)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(698257)
	}
	__antithesis_instrumentation__.Notify(698254)
	base := &Result{
		ActiveWarehouses: results[0].ActiveWarehouses,
		Cumulative: make(map[string]*hdrhistogram.Histogram,
			len(results[0].Cumulative)),
		WarehouseFactor: results[0].WarehouseFactor,
	}
	for _, r := range results {
		__antithesis_instrumentation__.Notify(698258)
		if r.Elapsed > base.Elapsed {
			__antithesis_instrumentation__.Notify(698261)
			base.Elapsed = r.Elapsed
		} else {
			__antithesis_instrumentation__.Notify(698262)
		}
		__antithesis_instrumentation__.Notify(698259)
		if r.ActiveWarehouses != base.ActiveWarehouses {
			__antithesis_instrumentation__.Notify(698263)
			panic(errors.Errorf("cannot merge histograms with different "+
				"ActiveWarehouses values: got both %v and %v",
				r.ActiveWarehouses, base.ActiveWarehouses))
		} else {
			__antithesis_instrumentation__.Notify(698264)
		}
		__antithesis_instrumentation__.Notify(698260)
		for q, h := range r.Cumulative {
			__antithesis_instrumentation__.Notify(698265)
			if cur, exists := base.Cumulative[q]; exists {
				__antithesis_instrumentation__.Notify(698266)
				cur.Merge(h)
			} else {
				__antithesis_instrumentation__.Notify(698267)
				base.Cumulative[q] = histogram.Copy(h)
			}
		}
	}
	__antithesis_instrumentation__.Notify(698255)
	return base
}

func NewResult(
	activeWarehouses int,
	warehouseFactor float64,
	elapsed time.Duration,
	cumulative map[string]*hdrhistogram.Histogram,
) *Result {
	__antithesis_instrumentation__.Notify(698268)
	return &Result{
		ActiveWarehouses: activeWarehouses,
		WarehouseFactor:  warehouseFactor,
		Elapsed:          elapsed,
		Cumulative:       cumulative,
	}
}

func NewResultWithSnapshots(
	activeWarehouses int, warehouseFactor float64, snapshots map[string][]histogram.SnapshotTick,
) *Result {
	__antithesis_instrumentation__.Notify(698269)
	var start time.Time
	var end time.Time
	ret := make(map[string]*hdrhistogram.Histogram, len(snapshots))
	for n, snaps := range snapshots {
		__antithesis_instrumentation__.Notify(698271)
		var cur *hdrhistogram.Histogram
		for _, s := range snaps {
			__antithesis_instrumentation__.Notify(698273)
			h := hdrhistogram.Import(s.Hist)
			if cur == nil {
				__antithesis_instrumentation__.Notify(698276)
				cur = h
			} else {
				__antithesis_instrumentation__.Notify(698277)
				cur.Merge(h)
			}
			__antithesis_instrumentation__.Notify(698274)
			if start.IsZero() || func() bool {
				__antithesis_instrumentation__.Notify(698278)
				return s.Now.Before(start) == true
			}() == true {
				__antithesis_instrumentation__.Notify(698279)
				start = s.Now
			} else {
				__antithesis_instrumentation__.Notify(698280)
			}
			__antithesis_instrumentation__.Notify(698275)
			if sEnd := s.Now.Add(s.Elapsed); end.IsZero() || func() bool {
				__antithesis_instrumentation__.Notify(698281)
				return sEnd.After(end) == true
			}() == true {
				__antithesis_instrumentation__.Notify(698282)
				end = sEnd
			} else {
				__antithesis_instrumentation__.Notify(698283)
			}
		}
		__antithesis_instrumentation__.Notify(698272)
		ret[n] = cur
	}
	__antithesis_instrumentation__.Notify(698270)
	return NewResult(activeWarehouses, warehouseFactor, end.Sub(start), ret)
}

func (r *Result) TpmC() float64 {
	__antithesis_instrumentation__.Notify(698284)
	no := r.Cumulative["newOrder"]
	if no == nil {
		__antithesis_instrumentation__.Notify(698286)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(698287)
	}
	__antithesis_instrumentation__.Notify(698285)
	return float64(no.TotalCount()) / (r.Elapsed.Seconds() / 60)
}

func (r *Result) Efficiency() float64 {
	__antithesis_instrumentation__.Notify(698288)
	tpmC := r.TpmC()
	warehouseFactor := r.WarehouseFactor
	if warehouseFactor == 0 {
		__antithesis_instrumentation__.Notify(698290)
		warehouseFactor = DeckWarehouseFactor
	} else {
		__antithesis_instrumentation__.Notify(698291)
	}
	__antithesis_instrumentation__.Notify(698289)
	return (100 * tpmC) / (warehouseFactor * float64(r.ActiveWarehouses))
}

func (r *Result) FailureError() error {
	__antithesis_instrumentation__.Notify(698292)
	if _, newOrderExists := r.Cumulative["newOrder"]; !newOrderExists {
		__antithesis_instrumentation__.Notify(698296)
		return errors.Errorf("no newOrder data exists")
	} else {
		__antithesis_instrumentation__.Notify(698297)
	}
	__antithesis_instrumentation__.Notify(698293)

	var err error
	if eff := r.Efficiency(); eff < PassingEfficiency {
		__antithesis_instrumentation__.Notify(698298)
		err = errors.CombineErrors(err,
			errors.Errorf("efficiency value of %v is below passing threshold of %v",
				eff, PassingEfficiency))
	} else {
		__antithesis_instrumentation__.Notify(698299)
	}
	__antithesis_instrumentation__.Notify(698294)
	for query, max90th := range passing90ThPercentile {
		__antithesis_instrumentation__.Notify(698300)
		h, exists := r.Cumulative[query]
		if !exists {
			__antithesis_instrumentation__.Notify(698302)
			return errors.Errorf("no %v data exists", query)
		} else {
			__antithesis_instrumentation__.Notify(698303)
		}
		__antithesis_instrumentation__.Notify(698301)
		if v := time.Duration(h.ValueAtQuantile(90)); v > max90th {
			__antithesis_instrumentation__.Notify(698304)
			err = errors.CombineErrors(err,
				errors.Errorf("90th percentile latency for %v at %v exceeds passing threshold of %v",
					query, v, max90th))
		} else {
			__antithesis_instrumentation__.Notify(698305)
		}
	}
	__antithesis_instrumentation__.Notify(698295)
	return err
}
