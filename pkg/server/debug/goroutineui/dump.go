package goroutineui

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"io"
	"io/ioutil"
	"runtime"
	"sort"
	"strings"

	"github.com/maruel/panicparse/v2/stack"
)

func stacks() []byte {
	__antithesis_instrumentation__.Notify(190209)

	var trace []byte
	for n := 1 << 20; n <= (1 << 29); n *= 2 {
		__antithesis_instrumentation__.Notify(190211)
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, true)
		if nbytes < len(trace) {
			__antithesis_instrumentation__.Notify(190212)
			return trace[:nbytes]
		} else {
			__antithesis_instrumentation__.Notify(190213)
		}
	}
	__antithesis_instrumentation__.Notify(190210)
	return trace
}

type Dump struct {
	agg *stack.Aggregated
	err error
}

func NewDump() Dump {
	__antithesis_instrumentation__.Notify(190214)
	return newDumpFromBytes(stacks(), stack.DefaultOpts())
}

func newDumpFromBytes(b []byte, opts *stack.Opts) Dump {
	__antithesis_instrumentation__.Notify(190215)
	s, _, err := stack.ScanSnapshot(bytes.NewBuffer(b), ioutil.Discard, opts)
	if err != io.EOF {
		__antithesis_instrumentation__.Notify(190217)
		return Dump{err: err}
	} else {
		__antithesis_instrumentation__.Notify(190218)
	}
	__antithesis_instrumentation__.Notify(190216)
	return Dump{agg: s.Aggregate(stack.AnyValue)}
}

func (d Dump) SortCountDesc() {
	__antithesis_instrumentation__.Notify(190219)
	if d.err != nil {
		__antithesis_instrumentation__.Notify(190221)
		return
	} else {
		__antithesis_instrumentation__.Notify(190222)
	}
	__antithesis_instrumentation__.Notify(190220)
	sort.Slice(d.agg.Buckets, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(190223)
		a, b := d.agg.Buckets[i], d.agg.Buckets[j]
		return len(a.IDs) > len(b.IDs)
	})
}

func (d Dump) SortWaitDesc() {
	__antithesis_instrumentation__.Notify(190224)
	if d.err != nil {
		__antithesis_instrumentation__.Notify(190226)
		return
	} else {
		__antithesis_instrumentation__.Notify(190227)
	}
	__antithesis_instrumentation__.Notify(190225)
	sort.Slice(d.agg.Buckets, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(190228)
		a, b := d.agg.Buckets[i], d.agg.Buckets[j]
		return a.SleepMax > b.SleepMax
	})
}

func (d Dump) HTML(w io.Writer) error {
	__antithesis_instrumentation__.Notify(190229)
	if d.err != nil {
		__antithesis_instrumentation__.Notify(190231)
		return d.err
	} else {
		__antithesis_instrumentation__.Notify(190232)
	}
	__antithesis_instrumentation__.Notify(190230)
	return d.agg.ToHTML(w, "")
}

func (d Dump) HTMLString() string {
	__antithesis_instrumentation__.Notify(190233)
	var w strings.Builder
	if err := d.HTML(&w); err != nil {
		__antithesis_instrumentation__.Notify(190235)
		return err.Error()
	} else {
		__antithesis_instrumentation__.Notify(190236)
	}
	__antithesis_instrumentation__.Notify(190234)
	return w.String()
}
