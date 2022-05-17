package yacc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Pos int

type ProductionNode struct {
	Pos
	Name        string
	Expressions []*ExpressionNode
}

func newProduction(pos Pos, name string) *ProductionNode {
	__antithesis_instrumentation__.Notify(68679)
	return &ProductionNode{Pos: pos, Name: name}
}

type ExpressionNode struct {
	Pos
	Items   []Item
	Command string
}

func newExpression(pos Pos) *ExpressionNode {
	__antithesis_instrumentation__.Notify(68680)
	return &ExpressionNode{Pos: pos}
}

type Item struct {
	Value string
	Typ   ItemTyp
}

type ItemTyp int

const (
	TypToken ItemTyp = iota

	TypLiteral
)
