package treecmp

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

type ComparisonOperator struct {
	Symbol ComparisonOperatorSymbol

	IsExplicitOperator bool
}

func MakeComparisonOperator(symbol ComparisonOperatorSymbol) ComparisonOperator {
	__antithesis_instrumentation__.Notify(614580)
	return ComparisonOperator{Symbol: symbol}
}

func (o ComparisonOperator) String() string {
	__antithesis_instrumentation__.Notify(614581)
	if o.IsExplicitOperator {
		__antithesis_instrumentation__.Notify(614583)
		return fmt.Sprintf("OPERATOR(%s)", o.Symbol.String())
	} else {
		__antithesis_instrumentation__.Notify(614584)
	}
	__antithesis_instrumentation__.Notify(614582)
	return o.Symbol.String()
}

func (ComparisonOperator) Operator() { __antithesis_instrumentation__.Notify(614585) }

type ComparisonOperatorSymbol int

const (
	EQ ComparisonOperatorSymbol = iota
	LT
	GT
	LE
	GE
	NE
	In
	NotIn
	Like
	NotLike
	ILike
	NotILike
	SimilarTo
	NotSimilarTo
	RegMatch
	NotRegMatch
	RegIMatch
	NotRegIMatch
	IsDistinctFrom
	IsNotDistinctFrom
	Contains
	ContainedBy
	JSONExists
	JSONSomeExists
	JSONAllExists
	Overlaps

	Any
	Some
	All

	NumComparisonOperatorSymbols
)

var _ = NumComparisonOperatorSymbols

var comparisonOpName = [...]string{
	EQ:           "=",
	LT:           "<",
	GT:           ">",
	LE:           "<=",
	GE:           ">=",
	NE:           "!=",
	In:           "IN",
	NotIn:        "NOT IN",
	Like:         "LIKE",
	NotLike:      "NOT LIKE",
	ILike:        "ILIKE",
	NotILike:     "NOT ILIKE",
	SimilarTo:    "SIMILAR TO",
	NotSimilarTo: "NOT SIMILAR TO",

	RegMatch:          "~",
	NotRegMatch:       "!~",
	RegIMatch:         "~*",
	NotRegIMatch:      "!~*",
	IsDistinctFrom:    "IS DISTINCT FROM",
	IsNotDistinctFrom: "IS NOT DISTINCT FROM",
	Contains:          "@>",
	ContainedBy:       "<@",
	JSONExists:        "?",
	JSONSomeExists:    "?|",
	JSONAllExists:     "?&",
	Overlaps:          "&&",
	Any:               "ANY",
	Some:              "SOME",
	All:               "ALL",
}

func (i ComparisonOperatorSymbol) String() string {
	__antithesis_instrumentation__.Notify(614586)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(614588)
		return i > ComparisonOperatorSymbol(len(comparisonOpName)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(614589)
		return fmt.Sprintf("ComparisonOp(%d)", i)
	} else {
		__antithesis_instrumentation__.Notify(614590)
	}
	__antithesis_instrumentation__.Notify(614587)
	return comparisonOpName[i]
}

func (i ComparisonOperatorSymbol) HasSubOperator() bool {
	__antithesis_instrumentation__.Notify(614591)
	switch i {
	case Any:
		__antithesis_instrumentation__.Notify(614593)
	case Some:
		__antithesis_instrumentation__.Notify(614594)
	case All:
		__antithesis_instrumentation__.Notify(614595)
	default:
		__antithesis_instrumentation__.Notify(614596)
		return false
	}
	__antithesis_instrumentation__.Notify(614592)
	return true
}

func ComparisonOpName(op ComparisonOperatorSymbol) string {
	__antithesis_instrumentation__.Notify(614597)
	if int(op) >= len(comparisonOpName) || func() bool {
		__antithesis_instrumentation__.Notify(614599)
		return comparisonOpName[op] == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(614600)
		panic(errors.AssertionFailedf("missing name for operator %q", op.String()))
	} else {
		__antithesis_instrumentation__.Notify(614601)
	}
	__antithesis_instrumentation__.Notify(614598)
	return comparisonOpName[op]
}
