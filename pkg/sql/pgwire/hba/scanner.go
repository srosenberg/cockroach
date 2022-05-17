package hba

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"regexp"
	"strings"

	"github.com/cockroachdb/errors"
)

type rule struct {
	re string

	fn func(l *lex) (foundToken bool, err error)
}

type lex struct {
	String

	comma bool

	lexed string
}

var rules = []struct {
	r  rule
	rg *regexp.Regexp
}{
	{r: rule{`[ \t\r,]*`, func(l *lex) (bool, error) { __antithesis_instrumentation__.Notify(560048); return false, nil }}},
	{r: rule{`#.*$`, func(l *lex) (bool, error) { __antithesis_instrumentation__.Notify(560049); return false, nil }}},
	{r: rule{`[^[:cntrl:] ",]+,?`, func(l *lex) (bool, error) {
		__antithesis_instrumentation__.Notify(560050)
		l.checkComma()
		l.Value = l.lexed
		return true, nil
	}}},
	{r: rule{`"[^[:cntrl:]"]*",?`, func(l *lex) (bool, error) {
		__antithesis_instrumentation__.Notify(560051)
		l.checkComma()
		l.stripQuotes()
		l.Value = l.lexed
		return true, nil
	}}},
	{r: rule{`"[^"]*$`, func(l *lex) (bool, error) {
		__antithesis_instrumentation__.Notify(560052)
		return false, errors.New("unterminated quoted string")
	}}},
	{r: rule{`"[^"]*"`, func(l *lex) (bool, error) {
		__antithesis_instrumentation__.Notify(560053)
		return false, errors.New("invalid characters in quoted string")
	}}},
	{r: rule{`.`, func(l *lex) (bool, error) {
		__antithesis_instrumentation__.Notify(560054)
		return false, errors.Newf("unsupported character: %q", l.lexed)
	}}},
}

func (l *lex) checkComma() {
	__antithesis_instrumentation__.Notify(560055)
	l.comma = l.lexed[len(l.lexed)-1] == ','
	if l.comma {
		__antithesis_instrumentation__.Notify(560056)
		l.lexed = l.lexed[:len(l.lexed)-1]
	} else {
		__antithesis_instrumentation__.Notify(560057)
	}
}

func (l *lex) stripQuotes() {
	__antithesis_instrumentation__.Notify(560058)
	l.Quoted = true
	l.lexed = l.lexed[1 : len(l.lexed)-1]
}

func init() {
	for i := range rules {
		rules[i].rg = regexp.MustCompile("^" + rules[i].r.re)
	}
}

func nextToken(buf string) (remaining string, tok String, trailingComma bool, err error) {
	__antithesis_instrumentation__.Notify(560059)
	remaining = buf
	var l lex
outer:
	for remaining != "" {
		__antithesis_instrumentation__.Notify(560061)
		l = lex{}
	inner:
		for _, rule := range rules {
			__antithesis_instrumentation__.Notify(560062)
			l.lexed = rule.rg.FindString(remaining)
			remaining = remaining[len(l.lexed):]
			if l.lexed != "" {
				__antithesis_instrumentation__.Notify(560063)
				var foundToken bool
				foundToken, err = rule.r.fn(&l)
				if foundToken || func() bool {
					__antithesis_instrumentation__.Notify(560065)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(560066)
					break outer
				} else {
					__antithesis_instrumentation__.Notify(560067)
				}
				__antithesis_instrumentation__.Notify(560064)
				break inner
			} else {
				__antithesis_instrumentation__.Notify(560068)
			}
		}
	}
	__antithesis_instrumentation__.Notify(560060)
	return remaining, l.String, l.comma, err
}

func nextFieldExpand(buf string) (remaining string, field []String, err error) {
	__antithesis_instrumentation__.Notify(560069)
	remaining = buf
	for {
		__antithesis_instrumentation__.Notify(560071)
		var trailingComma bool
		var tok String
		remaining, tok, trailingComma, err = nextToken(remaining)
		if tok.Empty() || func() bool {
			__antithesis_instrumentation__.Notify(560073)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(560074)
			return
		} else {
			__antithesis_instrumentation__.Notify(560075)
		}
		__antithesis_instrumentation__.Notify(560072)
		field = append(field, tok)
		if !trailingComma {
			__antithesis_instrumentation__.Notify(560076)
			break
		} else {
			__antithesis_instrumentation__.Notify(560077)
		}
	}
	__antithesis_instrumentation__.Notify(560070)
	return
}

func tokenize(input string) (res scannedInput, err error) {
	__antithesis_instrumentation__.Notify(560078)
	inputLines := strings.Split(input, "\n")

	for lineIdx, lineS := range inputLines {
		__antithesis_instrumentation__.Notify(560080)
		var currentLine hbaLine
		currentLine.input = strings.TrimSpace(lineS)
		for remaining := lineS; remaining != ""; {
			__antithesis_instrumentation__.Notify(560082)
			var currentField []String
			remaining, currentField, err = nextFieldExpand(remaining)
			if err != nil {
				__antithesis_instrumentation__.Notify(560084)
				return res, errors.Wrapf(err, "line %d", lineIdx+1)
			} else {
				__antithesis_instrumentation__.Notify(560085)
			}
			__antithesis_instrumentation__.Notify(560083)
			if len(currentField) > 0 {
				__antithesis_instrumentation__.Notify(560086)
				currentLine.tokens = append(currentLine.tokens, currentField)
			} else {
				__antithesis_instrumentation__.Notify(560087)
			}
		}
		__antithesis_instrumentation__.Notify(560081)
		if len(currentLine.tokens) > 0 {
			__antithesis_instrumentation__.Notify(560088)
			res.lines = append(res.lines, currentLine)
			res.linenos = append(res.linenos, lineIdx+1)
		} else {
			__antithesis_instrumentation__.Notify(560089)
		}
	}
	__antithesis_instrumentation__.Notify(560079)
	return res, err
}
