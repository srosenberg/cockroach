package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
)

var enclosingError = pgerror.Newf(pgcode.InvalidTextRepresentation, "array must be enclosed in { and }")
var extraTextError = pgerror.Newf(pgcode.InvalidTextRepresentation, "extra text after closing right brace")
var nestedArraysNotSupportedError = unimplemented.NewWithIssueDetail(32552, "strcast", "nested arrays not supported")
var malformedError = pgerror.Newf(pgcode.InvalidTextRepresentation, "malformed array")

var isQuoteChar = func(ch byte) bool {
	__antithesis_instrumentation__.Notify(611442)
	return ch == '"'
}

var isControlChar = func(ch byte) bool {
	__antithesis_instrumentation__.Notify(611443)
	return ch == '{' || func() bool {
		__antithesis_instrumentation__.Notify(611444)
		return ch == '}' == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(611445)
		return ch == ',' == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(611446)
		return ch == '"' == true
	}() == true
}

var isElementChar = func(r rune) bool {
	__antithesis_instrumentation__.Notify(611447)
	return r != '{' && func() bool {
		__antithesis_instrumentation__.Notify(611448)
		return r != '}' == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(611449)
		return r != ',' == true
	}() == true
}

func (p *parseState) gobbleString(isTerminatingChar func(ch byte) bool) (out string, err error) {
	__antithesis_instrumentation__.Notify(611450)
	var result bytes.Buffer
	start := 0
	i := 0
	for i < len(p.s) && func() bool {
		__antithesis_instrumentation__.Notify(611453)
		return !isTerminatingChar(p.s[i]) == true
	}() == true {
		__antithesis_instrumentation__.Notify(611454)

		if i < len(p.s) && func() bool {
			__antithesis_instrumentation__.Notify(611455)
			return p.s[i] == '\\' == true
		}() == true {
			__antithesis_instrumentation__.Notify(611456)
			result.WriteString(p.s[start:i])
			i++
			if i < len(p.s) {
				__antithesis_instrumentation__.Notify(611458)
				result.WriteByte(p.s[i])
				i++
			} else {
				__antithesis_instrumentation__.Notify(611459)
			}
			__antithesis_instrumentation__.Notify(611457)
			start = i
		} else {
			__antithesis_instrumentation__.Notify(611460)
			i++
		}
	}
	__antithesis_instrumentation__.Notify(611451)
	if i >= len(p.s) {
		__antithesis_instrumentation__.Notify(611461)
		return "", malformedError
	} else {
		__antithesis_instrumentation__.Notify(611462)
	}
	__antithesis_instrumentation__.Notify(611452)
	result.WriteString(p.s[start:i])
	p.s = p.s[i:]
	return result.String(), nil
}

type parseState struct {
	s                string
	ctx              ParseTimeContext
	dependsOnContext bool
	result           *DArray
	t                *types.T
}

func (p *parseState) advance() {
	__antithesis_instrumentation__.Notify(611463)
	_, l := utf8.DecodeRuneInString(p.s)
	p.s = p.s[l:]
}

func (p *parseState) eatWhitespace() {
	__antithesis_instrumentation__.Notify(611464)
	for unicode.IsSpace(p.peek()) {
		__antithesis_instrumentation__.Notify(611465)
		p.advance()
	}
}

func (p *parseState) peek() rune {
	__antithesis_instrumentation__.Notify(611466)
	r, _ := utf8.DecodeRuneInString(p.s)
	return r
}

func (p *parseState) eof() bool {
	__antithesis_instrumentation__.Notify(611467)
	return len(p.s) == 0
}

func (p *parseState) parseQuotedString() (string, error) {
	__antithesis_instrumentation__.Notify(611468)
	return p.gobbleString(isQuoteChar)
}

func (p *parseState) parseUnquotedString() (string, error) {
	__antithesis_instrumentation__.Notify(611469)
	out, err := p.gobbleString(isControlChar)
	if err != nil {
		__antithesis_instrumentation__.Notify(611471)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(611472)
	}
	__antithesis_instrumentation__.Notify(611470)
	return strings.TrimSpace(out), nil
}

func (p *parseState) parseElement() error {
	__antithesis_instrumentation__.Notify(611473)
	var next string
	var err error
	r := p.peek()
	switch r {
	case '{':
		__antithesis_instrumentation__.Notify(611477)
		return nestedArraysNotSupportedError
	case '"':
		__antithesis_instrumentation__.Notify(611478)
		p.advance()
		next, err = p.parseQuotedString()
		if err != nil {
			__antithesis_instrumentation__.Notify(611483)
			return err
		} else {
			__antithesis_instrumentation__.Notify(611484)
		}
		__antithesis_instrumentation__.Notify(611479)
		p.advance()
	default:
		__antithesis_instrumentation__.Notify(611480)
		if !isElementChar(r) {
			__antithesis_instrumentation__.Notify(611485)
			return malformedError
		} else {
			__antithesis_instrumentation__.Notify(611486)
		}
		__antithesis_instrumentation__.Notify(611481)
		next, err = p.parseUnquotedString()
		if err != nil {
			__antithesis_instrumentation__.Notify(611487)
			return err
		} else {
			__antithesis_instrumentation__.Notify(611488)
		}
		__antithesis_instrumentation__.Notify(611482)
		if strings.EqualFold(next, "null") {
			__antithesis_instrumentation__.Notify(611489)
			return p.result.Append(DNull)
		} else {
			__antithesis_instrumentation__.Notify(611490)
		}
	}
	__antithesis_instrumentation__.Notify(611474)

	d, dependsOnContext, err := ParseAndRequireString(p.t, next, p.ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(611491)
		return err
	} else {
		__antithesis_instrumentation__.Notify(611492)
	}
	__antithesis_instrumentation__.Notify(611475)
	if dependsOnContext {
		__antithesis_instrumentation__.Notify(611493)
		p.dependsOnContext = true
	} else {
		__antithesis_instrumentation__.Notify(611494)
	}
	__antithesis_instrumentation__.Notify(611476)
	return p.result.Append(d)
}

func ParseDArrayFromString(
	ctx ParseTimeContext, s string, t *types.T,
) (_ *DArray, dependsOnContext bool, _ error) {
	__antithesis_instrumentation__.Notify(611495)
	ret, dependsOnContext, err := doParseDArrayFromString(ctx, s, t)
	if err != nil {
		__antithesis_instrumentation__.Notify(611497)
		return ret, false, MakeParseError(s, types.MakeArray(t), err)
	} else {
		__antithesis_instrumentation__.Notify(611498)
	}
	__antithesis_instrumentation__.Notify(611496)
	return ret, dependsOnContext, nil
}

func doParseDArrayFromString(
	ctx ParseTimeContext, s string, t *types.T,
) (_ *DArray, dependsOnContext bool, _ error) {
	__antithesis_instrumentation__.Notify(611499)
	parser := parseState{
		s:      s,
		ctx:    ctx,
		result: NewDArray(t),
		t:      t,
	}

	parser.eatWhitespace()
	if parser.peek() != '{' {
		__antithesis_instrumentation__.Notify(611505)
		return nil, false, enclosingError
	} else {
		__antithesis_instrumentation__.Notify(611506)
	}
	__antithesis_instrumentation__.Notify(611500)
	parser.advance()
	parser.eatWhitespace()
	if parser.peek() != '}' {
		__antithesis_instrumentation__.Notify(611507)
		if err := parser.parseElement(); err != nil {
			__antithesis_instrumentation__.Notify(611509)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(611510)
		}
		__antithesis_instrumentation__.Notify(611508)
		parser.eatWhitespace()
		for parser.peek() == ',' {
			__antithesis_instrumentation__.Notify(611511)
			parser.advance()
			parser.eatWhitespace()
			if err := parser.parseElement(); err != nil {
				__antithesis_instrumentation__.Notify(611512)
				return nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(611513)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(611514)
	}
	__antithesis_instrumentation__.Notify(611501)
	parser.eatWhitespace()
	if parser.eof() {
		__antithesis_instrumentation__.Notify(611515)
		return nil, false, enclosingError
	} else {
		__antithesis_instrumentation__.Notify(611516)
	}
	__antithesis_instrumentation__.Notify(611502)
	if parser.peek() != '}' {
		__antithesis_instrumentation__.Notify(611517)
		return nil, false, malformedError
	} else {
		__antithesis_instrumentation__.Notify(611518)
	}
	__antithesis_instrumentation__.Notify(611503)
	parser.advance()
	parser.eatWhitespace()
	if !parser.eof() {
		__antithesis_instrumentation__.Notify(611519)
		return nil, false, extraTextError
	} else {
		__antithesis_instrumentation__.Notify(611520)
	}
	__antithesis_instrumentation__.Notify(611504)

	return parser.result, parser.dependsOnContext, nil
}
