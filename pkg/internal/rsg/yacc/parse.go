// Package yacc parses .y files.
package yacc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"runtime"

	"github.com/cockroachdb/errors"
)

type Tree struct {
	Name        string
	Productions []*ProductionNode
	text        string

	lex       *lexer
	token     [2]item
	peekCount int
}

func Parse(name, text string) (t *Tree, err error) {
	__antithesis_instrumentation__.Notify(68681)
	t = New(name)
	t.text = text
	err = t.Parse(text)
	return
}

func (t *Tree) next() item {
	__antithesis_instrumentation__.Notify(68682)
	if t.peekCount > 0 {
		__antithesis_instrumentation__.Notify(68684)
		t.peekCount--
	} else {
		__antithesis_instrumentation__.Notify(68685)
		t.token[0] = t.lex.nextItem()
	}
	__antithesis_instrumentation__.Notify(68683)
	return t.token[t.peekCount]
}

func (t *Tree) backup() {
	__antithesis_instrumentation__.Notify(68686)
	t.peekCount++
}

func (t *Tree) peek() item {
	__antithesis_instrumentation__.Notify(68687)
	if t.peekCount > 0 {
		__antithesis_instrumentation__.Notify(68689)
		return t.token[t.peekCount-1]
	} else {
		__antithesis_instrumentation__.Notify(68690)
	}
	__antithesis_instrumentation__.Notify(68688)
	t.peekCount = 1
	t.token[0] = t.lex.nextItem()
	return t.token[0]
}

func New(name string) *Tree {
	__antithesis_instrumentation__.Notify(68691)
	return &Tree{
		Name: name,
	}
}

func (t *Tree) errorf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(68692)
	err := fmt.Errorf(format, args...)
	err = errors.Wrapf(err, "parse: %s:%d", t.Name, t.lex.lineNumber())
	panic(err)
}

func (t *Tree) expect(expected itemType, context string) item {
	__antithesis_instrumentation__.Notify(68693)
	token := t.next()
	if token.typ != expected {
		__antithesis_instrumentation__.Notify(68695)
		t.unexpected(token, context)
	} else {
		__antithesis_instrumentation__.Notify(68696)
	}
	__antithesis_instrumentation__.Notify(68694)
	return token
}

func (t *Tree) unexpected(token item, context string) {
	__antithesis_instrumentation__.Notify(68697)
	t.errorf("unexpected %s in %s", token, context)
}

func (t *Tree) recover(errp *error) {
	__antithesis_instrumentation__.Notify(68698)
	if e := recover(); e != nil {
		__antithesis_instrumentation__.Notify(68699)
		if _, ok := e.(runtime.Error); ok {
			__antithesis_instrumentation__.Notify(68702)
			panic(e)
		} else {
			__antithesis_instrumentation__.Notify(68703)
		}
		__antithesis_instrumentation__.Notify(68700)
		if t != nil {
			__antithesis_instrumentation__.Notify(68704)
			t.stopParse()
		} else {
			__antithesis_instrumentation__.Notify(68705)
		}
		__antithesis_instrumentation__.Notify(68701)
		*errp = e.(error)
	} else {
		__antithesis_instrumentation__.Notify(68706)
	}
}

func (t *Tree) startParse(lex *lexer) {
	__antithesis_instrumentation__.Notify(68707)
	t.lex = lex
}

func (t *Tree) stopParse() {
	__antithesis_instrumentation__.Notify(68708)
	t.lex = nil
}

func (t *Tree) Parse(text string) (err error) {
	__antithesis_instrumentation__.Notify(68709)
	defer t.recover(&err)
	t.startParse(lex(t.Name, text))
	t.text = text
	t.parse()
	t.stopParse()
	return nil
}

func (t *Tree) parse() {
	__antithesis_instrumentation__.Notify(68710)
	for {
		__antithesis_instrumentation__.Notify(68711)
		switch token := t.next(); token.typ {
		case itemIdent:
			__antithesis_instrumentation__.Notify(68712)
			p := newProduction(token.pos, token.val)
			t.parseProduction(p)
			t.Productions = append(t.Productions, p)
		case itemEOF:
			__antithesis_instrumentation__.Notify(68713)
			return
		default:
			__antithesis_instrumentation__.Notify(68714)
		}
	}
}

func (t *Tree) parseProduction(p *ProductionNode) {
	__antithesis_instrumentation__.Notify(68715)
	const context = "production"
	t.expect(itemColon, context)
	if t.peek().typ == itemNL {
		__antithesis_instrumentation__.Notify(68717)
		t.next()
	} else {
		__antithesis_instrumentation__.Notify(68718)
	}
	__antithesis_instrumentation__.Notify(68716)
	expectExpr := true
	for {
		__antithesis_instrumentation__.Notify(68719)
		token := t.next()
		switch token.typ {
		case itemComment, itemNL:
			__antithesis_instrumentation__.Notify(68720)

		case itemPipe:
			__antithesis_instrumentation__.Notify(68721)
			if expectExpr {
				__antithesis_instrumentation__.Notify(68725)
				t.unexpected(token, context)
			} else {
				__antithesis_instrumentation__.Notify(68726)
			}
			__antithesis_instrumentation__.Notify(68722)
			expectExpr = true
		default:
			__antithesis_instrumentation__.Notify(68723)
			t.backup()
			if !expectExpr {
				__antithesis_instrumentation__.Notify(68727)
				return
			} else {
				__antithesis_instrumentation__.Notify(68728)
			}
			__antithesis_instrumentation__.Notify(68724)
			e := newExpression(token.pos)
			t.parseExpression(e)
			p.Expressions = append(p.Expressions, e)
			expectExpr = false
		}
	}
}

func (t *Tree) parseExpression(e *ExpressionNode) {
	__antithesis_instrumentation__.Notify(68729)
	const context = "expression"
	for {
		__antithesis_instrumentation__.Notify(68730)
		switch token := t.next(); token.typ {
		case itemNL:
			__antithesis_instrumentation__.Notify(68731)
			peek := t.peek().typ
			if peek == itemPipe || func() bool {
				__antithesis_instrumentation__.Notify(68738)
				return peek == itemNL == true
			}() == true {
				__antithesis_instrumentation__.Notify(68739)
				return
			} else {
				__antithesis_instrumentation__.Notify(68740)
			}
		case itemIdent:
			__antithesis_instrumentation__.Notify(68732)
			e.Items = append(e.Items, Item{token.val, TypToken})
		case itemLiteral:
			__antithesis_instrumentation__.Notify(68733)
			e.Items = append(e.Items, Item{token.val, TypLiteral})
		case itemExpr:
			__antithesis_instrumentation__.Notify(68734)
			e.Command = token.val
			if t.peek().typ == itemNL {
				__antithesis_instrumentation__.Notify(68741)
				t.next()
			} else {
				__antithesis_instrumentation__.Notify(68742)
			}
			__antithesis_instrumentation__.Notify(68735)
			return
		case itemPct, itemComment:
			__antithesis_instrumentation__.Notify(68736)

		default:
			__antithesis_instrumentation__.Notify(68737)
			t.unexpected(token, context)
		}
	}
}
