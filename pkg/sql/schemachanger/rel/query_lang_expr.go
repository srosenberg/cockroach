package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type expr interface {
	expr()

	encoded() interface{}
}

type Var string

func (Var) expr() { __antithesis_instrumentation__.Notify(578923) }

type valueExpr struct {
	value interface{}
}

func (v valueExpr) expr() { __antithesis_instrumentation__.Notify(578924) }

type anyExpr []interface{}

func (a anyExpr) expr() { __antithesis_instrumentation__.Notify(578925) }
