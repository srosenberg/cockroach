package colflow

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/errors"
)

type panicInjector struct {
	colexecop.OneInputNode
	colexecop.InitHelper
	rng *rand.Rand
}

var _ colexecop.Operator = &panicInjector{}

const (
	initPanicInjectionProbability = 0.001
	nextPanicInjectionProbability = 0.00001
)

func newPanicInjector(input colexecop.Operator) colexecop.Operator {
	__antithesis_instrumentation__.Notify(456168)
	rng, _ := randutil.NewTestRand()
	return &panicInjector{
		OneInputNode: colexecop.OneInputNode{Input: input},
		rng:          rng,
	}
}

func (i *panicInjector) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456169)
	if !i.InitHelper.Init(ctx) {
		__antithesis_instrumentation__.Notify(456172)
		return
	} else {
		__antithesis_instrumentation__.Notify(456173)
	}
	__antithesis_instrumentation__.Notify(456170)
	if i.rng.Float64() < initPanicInjectionProbability {
		__antithesis_instrumentation__.Notify(456174)
		log.Info(i.Ctx, "injecting panic in Init")
		colexecerror.ExpectedError(errors.New("injected panic in Init"))
	} else {
		__antithesis_instrumentation__.Notify(456175)
	}
	__antithesis_instrumentation__.Notify(456171)
	i.Input.Init(i.Ctx)
}

func (i *panicInjector) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(456176)
	if i.rng.Float64() < nextPanicInjectionProbability {
		__antithesis_instrumentation__.Notify(456178)
		log.Info(i.Ctx, "injecting panic in Next")
		colexecerror.ExpectedError(errors.New("injected panic in Next"))
	} else {
		__antithesis_instrumentation__.Notify(456179)
	}
	__antithesis_instrumentation__.Notify(456177)
	return i.Input.Next()
}
