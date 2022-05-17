package parser

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/scanner"
)

func makeScanner(str string) scanner.Scanner {
	__antithesis_instrumentation__.Notify(552598)
	var s scanner.Scanner
	s.Init(str)
	return s
}

func SplitFirstStatement(sql string) (pos int, ok bool) {
	__antithesis_instrumentation__.Notify(552599)
	s := makeScanner(sql)
	var lval = &sqlSymType{}
	for {
		__antithesis_instrumentation__.Notify(552600)
		s.Scan(lval)
		switch lval.ID() {
		case 0, lexbase.ERROR:
			__antithesis_instrumentation__.Notify(552601)
			return 0, false
		case ';':
			__antithesis_instrumentation__.Notify(552602)
			return s.Pos(), true
		default:
			__antithesis_instrumentation__.Notify(552603)
		}
	}
}

func Tokens(sql string) (tokens []TokenString, ok bool) {
	__antithesis_instrumentation__.Notify(552604)
	s := makeScanner(sql)
	for {
		__antithesis_instrumentation__.Notify(552606)
		var lval = &sqlSymType{}
		s.Scan(lval)
		if lval.ID() == lexbase.ERROR {
			__antithesis_instrumentation__.Notify(552609)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(552610)
		}
		__antithesis_instrumentation__.Notify(552607)
		if lval.ID() == 0 {
			__antithesis_instrumentation__.Notify(552611)
			break
		} else {
			__antithesis_instrumentation__.Notify(552612)
		}
		__antithesis_instrumentation__.Notify(552608)
		tokens = append(tokens, TokenString{TokenID: lval.ID(), Str: lval.Str()})
	}
	__antithesis_instrumentation__.Notify(552605)
	return tokens, true
}

func TokensIgnoreErrors(sql string) (tokens []TokenString) {
	__antithesis_instrumentation__.Notify(552613)
	s := makeScanner(sql)
	for {
		__antithesis_instrumentation__.Notify(552615)
		var lval = &sqlSymType{}
		s.Scan(lval)
		if lval.ID() == 0 {
			__antithesis_instrumentation__.Notify(552617)
			break
		} else {
			__antithesis_instrumentation__.Notify(552618)
		}
		__antithesis_instrumentation__.Notify(552616)
		tokens = append(tokens, TokenString{TokenID: lval.ID(), Str: lval.Str()})
	}
	__antithesis_instrumentation__.Notify(552614)
	return tokens
}

type TokenString struct {
	TokenID int32
	Str     string
}
