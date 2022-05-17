package colexecutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
)

type CancelChecker struct {
	colexecop.OneInputNode
	colexecop.InitHelper
	colexecop.NonExplainable

	callsSinceLastCheck uint32
}

var _ colexecop.Operator = &CancelChecker{}

func NewCancelChecker(op colexecop.Operator) *CancelChecker {
	__antithesis_instrumentation__.Notify(431713)
	return &CancelChecker{OneInputNode: colexecop.NewOneInputNode(op)}
}

func (c *CancelChecker) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(431714)
	if !c.InitHelper.Init(ctx) {
		__antithesis_instrumentation__.Notify(431716)
		return
	} else {
		__antithesis_instrumentation__.Notify(431717)
	}
	__antithesis_instrumentation__.Notify(431715)
	if c.Input != nil {
		__antithesis_instrumentation__.Notify(431718)

		c.Input.Init(c.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(431719)
	}
}

func (c *CancelChecker) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431720)
	c.CheckEveryCall()
	return c.Input.Next()
}

const cancelCheckInterval = 1024

func (c *CancelChecker) Check() {
	__antithesis_instrumentation__.Notify(431721)
	if c.callsSinceLastCheck%cancelCheckInterval == 0 {
		__antithesis_instrumentation__.Notify(431723)
		c.CheckEveryCall()
	} else {
		__antithesis_instrumentation__.Notify(431724)
	}
	__antithesis_instrumentation__.Notify(431722)

	c.callsSinceLastCheck++
}

func (c *CancelChecker) CheckEveryCall() {
	__antithesis_instrumentation__.Notify(431725)
	select {
	case <-c.Ctx.Done():
		__antithesis_instrumentation__.Notify(431726)
		colexecerror.ExpectedError(cancelchecker.QueryCanceledError)
	default:
		__antithesis_instrumentation__.Notify(431727)
	}
}
