package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type tripleDecl struct {
	entity    Var
	attribute Attr
	value     expr
}

func (f *tripleDecl) clause() { __antithesis_instrumentation__.Notify(578901) }

var _ Clause = (*tripleDecl)(nil)

type eqDecl struct {
	v    Var
	expr expr
}

func (e *eqDecl) clause() { __antithesis_instrumentation__.Notify(578902) }

type and []Clause

func (a and) clause() { __antithesis_instrumentation__.Notify(578903) }

type filterDecl struct {
	name          string
	vars          []Var
	predicateFunc interface{}
}

func (f filterDecl) clause() { __antithesis_instrumentation__.Notify(578904) }
