package treebin

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

type BinaryOperator struct {
	Symbol BinaryOperatorSymbol

	IsExplicitOperator bool
}

func MakeBinaryOperator(symbol BinaryOperatorSymbol) BinaryOperator {
	__antithesis_instrumentation__.Notify(614561)
	return BinaryOperator{Symbol: symbol}
}

func (o BinaryOperator) String() string {
	__antithesis_instrumentation__.Notify(614562)
	if o.IsExplicitOperator {
		__antithesis_instrumentation__.Notify(614564)
		return fmt.Sprintf("OPERATOR(%s)", o.Symbol.String())
	} else {
		__antithesis_instrumentation__.Notify(614565)
	}
	__antithesis_instrumentation__.Notify(614563)
	return o.Symbol.String()
}

func (BinaryOperator) Operator() { __antithesis_instrumentation__.Notify(614566) }

type BinaryOperatorSymbol uint8

const (
	Bitand BinaryOperatorSymbol = iota
	Bitor
	Bitxor
	Plus
	Minus
	Mult
	Div
	FloorDiv
	Mod
	Pow
	Concat
	LShift
	RShift
	JSONFetchVal
	JSONFetchText
	JSONFetchValPath
	JSONFetchTextPath

	NumBinaryOperatorSymbols
)

var _ = NumBinaryOperatorSymbols

var binaryOpName = [...]string{
	Bitand:            "&",
	Bitor:             "|",
	Bitxor:            "#",
	Plus:              "+",
	Minus:             "-",
	Mult:              "*",
	Div:               "/",
	FloorDiv:          "//",
	Mod:               "%",
	Pow:               "^",
	Concat:            "||",
	LShift:            "<<",
	RShift:            ">>",
	JSONFetchVal:      "->",
	JSONFetchText:     "->>",
	JSONFetchValPath:  "#>",
	JSONFetchTextPath: "#>>",
}

func (i BinaryOperatorSymbol) IsPadded() bool {
	__antithesis_instrumentation__.Notify(614567)
	return !(i == JSONFetchVal || func() bool {
		__antithesis_instrumentation__.Notify(614568)
		return i == JSONFetchText == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(614569)
		return i == JSONFetchValPath == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(614570)
		return i == JSONFetchTextPath == true
	}() == true)
}

func (i BinaryOperatorSymbol) String() string {
	__antithesis_instrumentation__.Notify(614571)
	if i > BinaryOperatorSymbol(len(binaryOpName)-1) {
		__antithesis_instrumentation__.Notify(614573)
		return fmt.Sprintf("BinaryOp(%d)", i)
	} else {
		__antithesis_instrumentation__.Notify(614574)
	}
	__antithesis_instrumentation__.Notify(614572)
	return binaryOpName[i]
}

func BinaryOpName(op BinaryOperatorSymbol) string {
	__antithesis_instrumentation__.Notify(614575)
	if int(op) >= len(binaryOpName) || func() bool {
		__antithesis_instrumentation__.Notify(614577)
		return binaryOpName[op] == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(614578)
		panic(errors.AssertionFailedf("missing name for operator %q", op.String()))
	} else {
		__antithesis_instrumentation__.Notify(614579)
	}
	__antithesis_instrumentation__.Notify(614576)
	return binaryOpName[op]
}
