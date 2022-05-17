package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type statementWeight struct {
	weight int
	elem   statement
}

func newWeightedStatementSampler(weights []statementWeight, seed int64) *statementSampler {
	__antithesis_instrumentation__.Notify(69372)
	sum := 0
	for _, w := range weights {
		__antithesis_instrumentation__.Notify(69376)
		if w.weight < 1 {
			__antithesis_instrumentation__.Notify(69378)
			panic("expected weight >= 1")
		} else {
			__antithesis_instrumentation__.Notify(69379)
		}
		__antithesis_instrumentation__.Notify(69377)
		sum += w.weight
	}
	__antithesis_instrumentation__.Notify(69373)
	if sum == 0 {
		__antithesis_instrumentation__.Notify(69380)
		panic("expected weights")
	} else {
		__antithesis_instrumentation__.Notify(69381)
	}
	__antithesis_instrumentation__.Notify(69374)
	samples := make([]statement, sum)
	pos := 0
	for _, w := range weights {
		__antithesis_instrumentation__.Notify(69382)
		for count := 0; count < w.weight; count++ {
			__antithesis_instrumentation__.Notify(69383)
			samples[pos] = w.elem
			pos++
		}
	}
	__antithesis_instrumentation__.Notify(69375)
	return &statementSampler{
		rnd:     rand.New(rand.NewSource(seed)),
		samples: samples,
	}
}

type statementSampler struct {
	mu      syncutil.Mutex
	rnd     *rand.Rand
	samples []statement
}

func (w *statementSampler) Next() statement {
	__antithesis_instrumentation__.Notify(69384)
	w.mu.Lock()
	v := w.samples[w.rnd.Intn(len(w.samples))]
	w.mu.Unlock()
	return v
}

type tableExprWeight struct {
	weight int
	elem   tableExpr
}

func newWeightedTableExprSampler(weights []tableExprWeight, seed int64) *tableExprSampler {
	__antithesis_instrumentation__.Notify(69385)
	sum := 0
	for _, w := range weights {
		__antithesis_instrumentation__.Notify(69389)
		if w.weight < 1 {
			__antithesis_instrumentation__.Notify(69391)
			panic("expected weight >= 1")
		} else {
			__antithesis_instrumentation__.Notify(69392)
		}
		__antithesis_instrumentation__.Notify(69390)
		sum += w.weight
	}
	__antithesis_instrumentation__.Notify(69386)
	if sum == 0 {
		__antithesis_instrumentation__.Notify(69393)
		panic("expected weights")
	} else {
		__antithesis_instrumentation__.Notify(69394)
	}
	__antithesis_instrumentation__.Notify(69387)
	samples := make([]tableExpr, sum)
	pos := 0
	for _, w := range weights {
		__antithesis_instrumentation__.Notify(69395)
		for count := 0; count < w.weight; count++ {
			__antithesis_instrumentation__.Notify(69396)
			samples[pos] = w.elem
			pos++
		}
	}
	__antithesis_instrumentation__.Notify(69388)
	return &tableExprSampler{
		rnd:     rand.New(rand.NewSource(seed)),
		samples: samples,
	}
}

type tableExprSampler struct {
	mu      syncutil.Mutex
	rnd     *rand.Rand
	samples []tableExpr
}

func (w *tableExprSampler) Next() tableExpr {
	__antithesis_instrumentation__.Notify(69397)
	w.mu.Lock()
	v := w.samples[w.rnd.Intn(len(w.samples))]
	w.mu.Unlock()
	return v
}

type selectStatementWeight struct {
	weight int
	elem   selectStatement
}

func newWeightedSelectStatementSampler(
	weights []selectStatementWeight, seed int64,
) *selectStatementSampler {
	__antithesis_instrumentation__.Notify(69398)
	sum := 0
	for _, w := range weights {
		__antithesis_instrumentation__.Notify(69402)
		if w.weight < 1 {
			__antithesis_instrumentation__.Notify(69404)
			panic("expected weight >= 1")
		} else {
			__antithesis_instrumentation__.Notify(69405)
		}
		__antithesis_instrumentation__.Notify(69403)
		sum += w.weight
	}
	__antithesis_instrumentation__.Notify(69399)
	if sum == 0 {
		__antithesis_instrumentation__.Notify(69406)
		panic("expected weights")
	} else {
		__antithesis_instrumentation__.Notify(69407)
	}
	__antithesis_instrumentation__.Notify(69400)
	samples := make([]selectStatement, sum)
	pos := 0
	for _, w := range weights {
		__antithesis_instrumentation__.Notify(69408)
		for count := 0; count < w.weight; count++ {
			__antithesis_instrumentation__.Notify(69409)
			samples[pos] = w.elem
			pos++
		}
	}
	__antithesis_instrumentation__.Notify(69401)
	return &selectStatementSampler{
		rnd:     rand.New(rand.NewSource(seed)),
		samples: samples,
	}
}

type selectStatementSampler struct {
	mu      syncutil.Mutex
	rnd     *rand.Rand
	samples []selectStatement
}

func (w *selectStatementSampler) Next() selectStatement {
	__antithesis_instrumentation__.Notify(69410)
	w.mu.Lock()
	v := w.samples[w.rnd.Intn(len(w.samples))]
	w.mu.Unlock()
	return v
}

type scalarExprWeight struct {
	weight int
	elem   scalarExpr
}

func newWeightedScalarExprSampler(weights []scalarExprWeight, seed int64) *scalarExprSampler {
	__antithesis_instrumentation__.Notify(69411)
	sum := 0
	for _, w := range weights {
		__antithesis_instrumentation__.Notify(69415)
		if w.weight < 1 {
			__antithesis_instrumentation__.Notify(69417)
			panic("expected weight >= 1")
		} else {
			__antithesis_instrumentation__.Notify(69418)
		}
		__antithesis_instrumentation__.Notify(69416)
		sum += w.weight
	}
	__antithesis_instrumentation__.Notify(69412)
	if sum == 0 {
		__antithesis_instrumentation__.Notify(69419)
		panic("expected weights")
	} else {
		__antithesis_instrumentation__.Notify(69420)
	}
	__antithesis_instrumentation__.Notify(69413)
	samples := make([]scalarExpr, sum)
	pos := 0
	for _, w := range weights {
		__antithesis_instrumentation__.Notify(69421)
		for count := 0; count < w.weight; count++ {
			__antithesis_instrumentation__.Notify(69422)
			samples[pos] = w.elem
			pos++
		}
	}
	__antithesis_instrumentation__.Notify(69414)
	return &scalarExprSampler{
		rnd:     rand.New(rand.NewSource(seed)),
		samples: samples,
	}
}

type scalarExprSampler struct {
	mu      syncutil.Mutex
	rnd     *rand.Rand
	samples []scalarExpr
}

func (w *scalarExprSampler) Next() scalarExpr {
	__antithesis_instrumentation__.Notify(69423)
	w.mu.Lock()
	v := w.samples[w.rnd.Intn(len(w.samples))]
	w.mu.Unlock()
	return v
}
