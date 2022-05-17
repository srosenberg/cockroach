package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/cockroach/pkg/workload/tpcc"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

var tpccChecksMeta = workload.Meta{
	Name:        `tpcc-checks`,
	Description: `tpcc-checks runs the TPC-C consistency checks as a workload.`,
	Details: `It is primarily intended as a tool to create an overload scenario.
An --as-of flag is exposed to prevent the work from interfering with a
foreground TPC-C workload`,
	Version: `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(698637)
		g := &tpccChecks{}
		g.flags.FlagSet = pflag.NewFlagSet(`tpcc`, pflag.ContinueOnError)
		g.flags.Meta = map[string]workload.FlagMeta{
			`db`:          {RuntimeOnly: true},
			`concurrency`: {RuntimeOnly: true},
			`as-of`:       {RuntimeOnly: true},
		}
		g.flags.IntVar(&g.concurrency, `concurrency`, 1,
			`Number of concurrent workers. Defaults to 1.`,
		)
		g.flags.StringVar(&g.asOfSystemTime, "as-of", "",
			"Timestamp at which the query should be run."+
				" If non-empty the provided value will be used as the expression in an"+
				" AS OF SYSTEM TIME CLAUSE for all checks.")
		checkNames := func() (checkNames []string) {
			__antithesis_instrumentation__.Notify(698640)
			for _, c := range tpcc.AllChecks() {
				__antithesis_instrumentation__.Notify(698642)
				checkNames = append(checkNames, c.Name)
			}
			__antithesis_instrumentation__.Notify(698641)
			return checkNames
		}()
		__antithesis_instrumentation__.Notify(698638)
		g.flags.StringSliceVar(&g.checks, "checks", checkNames,
			"Name of checks to be run.")
		g.connFlags = workload.NewConnFlags(&g.flags)
		{
			__antithesis_instrumentation__.Notify(698643)
			dbOverrideFlag := g.flags.Lookup(`db`)
			dbOverrideFlag.DefValue = `tpcc`
			if err := dbOverrideFlag.Value.Set(`tpcc`); err != nil {
				__antithesis_instrumentation__.Notify(698644)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(698645)
			}
		}
		__antithesis_instrumentation__.Notify(698639)
		return g
	},
}

func (w *tpccChecks) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(698646)
	return w.flags
}

func init() {
	workload.Register(tpccChecksMeta)
}

type tpccChecks struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	asOfSystemTime string
	checks         []string
	concurrency    int
}

func (*tpccChecks) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(698647)
	return nil
}

func (*tpccChecks) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(698648)
	return tpccChecksMeta
}

func (w *tpccChecks) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(698649)
	sqlDatabase, err := workload.SanitizeUrls(w, w.flags.Lookup("db").Value.String(), urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(698655)
		return workload.QueryLoad{}, errors.Wrapf(err, "could not sanitize urls %v", urls)
	} else {
		__antithesis_instrumentation__.Notify(698656)
	}
	__antithesis_instrumentation__.Notify(698650)
	dbs := make([]*gosql.DB, len(urls))
	for i, url := range urls {
		__antithesis_instrumentation__.Notify(698657)
		dbs[i], err = gosql.Open(`cockroach`, url)
		if err != nil {
			__antithesis_instrumentation__.Notify(698659)
			return workload.QueryLoad{}, errors.Wrapf(err, "failed to dial %s", url)
		} else {
			__antithesis_instrumentation__.Notify(698660)
		}
		__antithesis_instrumentation__.Notify(698658)

		dbs[i].SetMaxOpenConns(3 * w.concurrency)
		dbs[i].SetMaxIdleConns(3 * w.concurrency)
	}
	__antithesis_instrumentation__.Notify(698651)
	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	ql.WorkerFns = make([]func(context.Context) error, w.concurrency)
	checks, err := filterChecks(tpcc.AllChecks(), w.checks)
	if err != nil {
		__antithesis_instrumentation__.Notify(698661)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(698662)
	}
	__antithesis_instrumentation__.Notify(698652)
	for i := range ql.WorkerFns {
		__antithesis_instrumentation__.Notify(698663)
		worker := newCheckWorker(dbs, checks, reg.GetHandle(), w.asOfSystemTime)
		ql.WorkerFns[i] = worker.run
	}
	__antithesis_instrumentation__.Notify(698653)

	for _, c := range checks {
		__antithesis_instrumentation__.Notify(698664)
		reg.GetHandle().Get(c.Name)
	}
	__antithesis_instrumentation__.Notify(698654)
	return ql, nil
}

type checkWorker struct {
	dbs            []*gosql.DB
	checks         []tpcc.Check
	histograms     *histogram.Histograms
	asOfSystemTime string
	dbPerm         []int
	checkPerm      []int
	i              int
}

func newCheckWorker(
	dbs []*gosql.DB, checks []tpcc.Check, histograms *histogram.Histograms, asOfSystemTime string,
) *checkWorker {
	__antithesis_instrumentation__.Notify(698665)
	return &checkWorker{
		dbs:            dbs,
		checks:         checks,
		histograms:     histograms,
		asOfSystemTime: asOfSystemTime,
		dbPerm:         rand.Perm(len(dbs)),
		checkPerm:      rand.Perm(len(checks)),
	}
}

func (w *checkWorker) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(698666)
	defer func() { __antithesis_instrumentation__.Notify(698669); w.i++ }()
	__antithesis_instrumentation__.Notify(698667)
	c := w.checks[w.checkPerm[w.i%len(w.checks)]]
	db := w.dbs[w.dbPerm[w.i%len(w.dbs)]]
	start := timeutil.Now()
	if err := c.Fn(db, w.asOfSystemTime); err != nil {
		__antithesis_instrumentation__.Notify(698670)
		return errors.Wrapf(err, "failed check %s", c.Name)
	} else {
		__antithesis_instrumentation__.Notify(698671)
	}
	__antithesis_instrumentation__.Notify(698668)
	w.histograms.Get(c.Name).Record(timeutil.Since(start))
	return nil
}

func filterChecks(checks []tpcc.Check, toRun []string) ([]tpcc.Check, error) {
	__antithesis_instrumentation__.Notify(698672)
	toRunSet := make(map[string]struct{}, len(toRun))
	for _, s := range toRun {
		__antithesis_instrumentation__.Notify(698676)
		toRunSet[s] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(698673)
	filtered := checks[:0]
	for _, c := range checks {
		__antithesis_instrumentation__.Notify(698677)
		if _, exists := toRunSet[c.Name]; exists {
			__antithesis_instrumentation__.Notify(698678)
			filtered = append(filtered, c)
			delete(toRunSet, c.Name)
		} else {
			__antithesis_instrumentation__.Notify(698679)
		}
	}
	__antithesis_instrumentation__.Notify(698674)
	if len(toRunSet) > 0 {
		__antithesis_instrumentation__.Notify(698680)
		return nil, fmt.Errorf("cannot run checks %v which do not exist", toRun)
	} else {
		__antithesis_instrumentation__.Notify(698681)
	}
	__antithesis_instrumentation__.Notify(698675)
	return filtered, nil
}
