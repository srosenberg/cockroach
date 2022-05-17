package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"io"
	"time"

	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
)

type outputFormat interface {
	rampDone()

	outputError(err error)

	outputTick(startElapsed time.Duration, t histogram.Tick)

	outputTotal(startElapsed time.Duration, t histogram.Tick)

	outputResult(startElapsed time.Duration, t histogram.Tick)
}

type textFormatter struct {
	i      int
	numErr int
}

func (f *textFormatter) rampDone() {
	__antithesis_instrumentation__.Notify(693670)
	f.i = 0
}

func (f *textFormatter) outputError(_ error) {
	__antithesis_instrumentation__.Notify(693671)
	f.numErr++
}

func (f *textFormatter) outputTick(startElapsed time.Duration, t histogram.Tick) {
	__antithesis_instrumentation__.Notify(693672)
	if f.i%20 == 0 {
		__antithesis_instrumentation__.Notify(693674)
		fmt.Println("_elapsed___errors__ops/sec(inst)___ops/sec(cum)__p50(ms)__p95(ms)__p99(ms)_pMax(ms)")
	} else {
		__antithesis_instrumentation__.Notify(693675)
	}
	__antithesis_instrumentation__.Notify(693673)
	f.i++
	fmt.Printf("%7.1fs %8d %14.1f %14.1f %8.1f %8.1f %8.1f %8.1f %s\n",
		startElapsed.Seconds(),
		f.numErr,
		float64(t.Hist.TotalCount())/t.Elapsed.Seconds(),
		float64(t.Cumulative.TotalCount())/startElapsed.Seconds(),
		time.Duration(t.Hist.ValueAtQuantile(50)).Seconds()*1000,
		time.Duration(t.Hist.ValueAtQuantile(95)).Seconds()*1000,
		time.Duration(t.Hist.ValueAtQuantile(99)).Seconds()*1000,
		time.Duration(t.Hist.ValueAtQuantile(100)).Seconds()*1000,
		t.Name,
	)
}

const totalHeader = "\n_elapsed___errors_____ops(total)___ops/sec(cum)__avg(ms)__p50(ms)__p95(ms)__p99(ms)_pMax(ms)"

func (f *textFormatter) outputTotal(startElapsed time.Duration, t histogram.Tick) {
	__antithesis_instrumentation__.Notify(693676)
	f.outputFinal(startElapsed, t, "__total")
}

func (f *textFormatter) outputResult(startElapsed time.Duration, t histogram.Tick) {
	__antithesis_instrumentation__.Notify(693677)
	f.outputFinal(startElapsed, t, "__result")
}

func (f *textFormatter) outputFinal(
	startElapsed time.Duration, t histogram.Tick, titleSuffix string,
) {
	__antithesis_instrumentation__.Notify(693678)
	fmt.Println(totalHeader + titleSuffix)
	if t.Cumulative == nil {
		__antithesis_instrumentation__.Notify(693681)
		return
	} else {
		__antithesis_instrumentation__.Notify(693682)
	}
	__antithesis_instrumentation__.Notify(693679)
	if t.Cumulative.TotalCount() == 0 {
		__antithesis_instrumentation__.Notify(693683)
		return
	} else {
		__antithesis_instrumentation__.Notify(693684)
	}
	__antithesis_instrumentation__.Notify(693680)
	fmt.Printf("%7.1fs %8d %14d %14.1f %8.1f %8.1f %8.1f %8.1f %8.1f  %s\n",
		startElapsed.Seconds(),
		f.numErr,
		t.Cumulative.TotalCount(),
		float64(t.Cumulative.TotalCount())/startElapsed.Seconds(),
		time.Duration(t.Cumulative.Mean()).Seconds()*1000,
		time.Duration(t.Cumulative.ValueAtQuantile(50)).Seconds()*1000,
		time.Duration(t.Cumulative.ValueAtQuantile(95)).Seconds()*1000,
		time.Duration(t.Cumulative.ValueAtQuantile(99)).Seconds()*1000,
		time.Duration(t.Cumulative.ValueAtQuantile(100)).Seconds()*1000,
		t.Name,
	)
}

type jsonFormatter struct {
	w      io.Writer
	numErr int
}

func (f *jsonFormatter) rampDone() { __antithesis_instrumentation__.Notify(693685) }

func (f *jsonFormatter) outputError(_ error) {
	__antithesis_instrumentation__.Notify(693686)
	f.numErr++
}

func (f *jsonFormatter) outputTick(startElapsed time.Duration, t histogram.Tick) {
	__antithesis_instrumentation__.Notify(693687)

	fmt.Fprintf(f.w, `{"time":"%s",`+
		`"errs":%d,`+
		`"avgt":%.1f,`+
		`"avgl":%.1f,`+
		`"p50l":%.1f,`+
		`"p95l":%.1f,`+
		`"p99l":%.1f,`+
		`"maxl":%.1f,`+
		`"type":"%s"`+
		"}\n",
		t.Now.UTC().Format(time.RFC3339Nano),
		f.numErr,
		float64(t.Hist.TotalCount())/t.Elapsed.Seconds(),
		float64(t.Cumulative.TotalCount())/startElapsed.Seconds(),
		time.Duration(t.Hist.ValueAtQuantile(50)).Seconds()*1000,
		time.Duration(t.Hist.ValueAtQuantile(95)).Seconds()*1000,
		time.Duration(t.Hist.ValueAtQuantile(99)).Seconds()*1000,
		time.Duration(t.Hist.ValueAtQuantile(100)).Seconds()*1000,
		t.Name,
	)
}

func (f *jsonFormatter) outputTotal(startElapsed time.Duration, t histogram.Tick) {
	__antithesis_instrumentation__.Notify(693688)
}

func (f *jsonFormatter) outputResult(startElapsed time.Duration, t histogram.Tick) {
	__antithesis_instrumentation__.Notify(693689)
}
