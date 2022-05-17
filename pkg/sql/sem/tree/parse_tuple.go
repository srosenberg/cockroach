package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

var enclosingRecordError = pgerror.Newf(pgcode.InvalidTextRepresentation, "record must be enclosed in ( and )")
var extraTextRecordError = pgerror.Newf(pgcode.InvalidTextRepresentation, "extra text after closing right paren")
var malformedRecordError = pgerror.Newf(pgcode.InvalidTextRepresentation, "malformed record literal")
var unsupportedRecordError = pgerror.Newf(pgcode.FeatureNotSupported, "cannot parse anonymous record type")

var isTupleControlChar = func(ch byte) bool {
	__antithesis_instrumentation__.Notify(611566)
	return ch == '(' || func() bool {
		__antithesis_instrumentation__.Notify(611567)
		return ch == ')' == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(611568)
		return ch == ',' == true
	}() == true
}

var isTupleElementChar = func(r rune) bool {
	__antithesis_instrumentation__.Notify(611569)
	return r != '(' && func() bool {
		__antithesis_instrumentation__.Notify(611570)
		return r != ')' == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(611571)
		return r != ',' == true
	}() == true
}

func (p *tupleParseState) gobbleString() (out string, err error) {
	__antithesis_instrumentation__.Notify(611572)
	isTerminatingChar := func(inQuote bool, ch byte) bool {
		__antithesis_instrumentation__.Notify(611577)
		if inQuote {
			__antithesis_instrumentation__.Notify(611579)
			return isQuoteChar(ch)
		} else {
			__antithesis_instrumentation__.Notify(611580)
		}
		__antithesis_instrumentation__.Notify(611578)
		return isTupleControlChar(ch)
	}
	__antithesis_instrumentation__.Notify(611573)
	var result bytes.Buffer
	start := 0
	i := 0
	inQuote := false
	for i < len(p.s) && func() bool {
		__antithesis_instrumentation__.Notify(611581)
		return (!isTerminatingChar(inQuote, p.s[i]) || func() bool {
			__antithesis_instrumentation__.Notify(611582)
			return inQuote == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(611583)

		if i < len(p.s) && func() bool {
			__antithesis_instrumentation__.Notify(611584)
			return p.s[i] == '\\' == true
		}() == true {
			__antithesis_instrumentation__.Notify(611585)
			result.WriteString(p.s[start:i])
			i++
			if i < len(p.s) {
				__antithesis_instrumentation__.Notify(611587)
				result.WriteByte(p.s[i])
				i++
			} else {
				__antithesis_instrumentation__.Notify(611588)
			}
			__antithesis_instrumentation__.Notify(611586)
			start = i
		} else {
			__antithesis_instrumentation__.Notify(611589)
			if i < len(p.s) && func() bool {
				__antithesis_instrumentation__.Notify(611590)
				return p.s[i] == '"' == true
			}() == true {
				__antithesis_instrumentation__.Notify(611591)
				result.WriteString(p.s[start:i])
				i++
				if inQuote && func() bool {
					__antithesis_instrumentation__.Notify(611593)
					return i < len(p.s) == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(611594)
					return p.s[i] == '"' == true
				}() == true {
					__antithesis_instrumentation__.Notify(611595)

					result.WriteByte(p.s[i])
					i++
				} else {
					__antithesis_instrumentation__.Notify(611596)

					inQuote = !inQuote
				}
				__antithesis_instrumentation__.Notify(611592)

				start = i
			} else {
				__antithesis_instrumentation__.Notify(611597)
				i++
			}
		}
	}
	__antithesis_instrumentation__.Notify(611574)
	if i >= len(p.s) {
		__antithesis_instrumentation__.Notify(611598)
		return "", malformedRecordError
	} else {
		__antithesis_instrumentation__.Notify(611599)
	}
	__antithesis_instrumentation__.Notify(611575)
	if inQuote {
		__antithesis_instrumentation__.Notify(611600)
		return "", malformedRecordError
	} else {
		__antithesis_instrumentation__.Notify(611601)
	}
	__antithesis_instrumentation__.Notify(611576)
	result.WriteString(p.s[start:i])
	p.s = p.s[i:]
	return result.String(), nil
}

type tupleParseState struct {
	s                string
	tupleIdx         int
	ctx              ParseTimeContext
	dependsOnContext bool
	result           *DTuple
	t                *types.T
}

func (p *tupleParseState) advance() {
	__antithesis_instrumentation__.Notify(611602)
	_, l := utf8.DecodeRuneInString(p.s)
	p.s = p.s[l:]
}

func (p *tupleParseState) eatWhitespace() {
	__antithesis_instrumentation__.Notify(611603)
	for unicode.IsSpace(p.peek()) {
		__antithesis_instrumentation__.Notify(611604)
		p.advance()
	}
}

func (p *tupleParseState) peek() rune {
	__antithesis_instrumentation__.Notify(611605)
	r, _ := utf8.DecodeRuneInString(p.s)
	return r
}

func (p *tupleParseState) eof() bool {
	__antithesis_instrumentation__.Notify(611606)
	return len(p.s) == 0
}

func (p *tupleParseState) parseString() (string, error) {
	__antithesis_instrumentation__.Notify(611607)
	out, err := p.gobbleString()
	if err != nil {
		__antithesis_instrumentation__.Notify(611609)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(611610)
	}
	__antithesis_instrumentation__.Notify(611608)

	return out, nil
}

func (p *tupleParseState) parseElement() error {
	__antithesis_instrumentation__.Notify(611611)
	if p.tupleIdx >= len(p.t.TupleContents()) {
		__antithesis_instrumentation__.Notify(611616)
		return errors.WithDetail(malformedRecordError, "Too many columns.")
	} else {
		__antithesis_instrumentation__.Notify(611617)
	}
	__antithesis_instrumentation__.Notify(611612)
	var next string
	var err error
	r := p.peek()
	switch r {
	case ')', ',':
		__antithesis_instrumentation__.Notify(611618)

		p.result.D[p.tupleIdx] = DNull
		p.tupleIdx++
		return nil
	default:
		__antithesis_instrumentation__.Notify(611619)
		if !isTupleElementChar(r) {
			__antithesis_instrumentation__.Notify(611621)
			return malformedRecordError
		} else {
			__antithesis_instrumentation__.Notify(611622)
		}
		__antithesis_instrumentation__.Notify(611620)
		next, err = p.parseString()
		if err != nil {
			__antithesis_instrumentation__.Notify(611623)
			return err
		} else {
			__antithesis_instrumentation__.Notify(611624)
		}
	}
	__antithesis_instrumentation__.Notify(611613)

	d, dependsOnContext, err := ParseAndRequireString(
		p.t.TupleContents()[p.tupleIdx],
		next,
		p.ctx,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(611625)
		return err
	} else {
		__antithesis_instrumentation__.Notify(611626)
	}
	__antithesis_instrumentation__.Notify(611614)
	if dependsOnContext {
		__antithesis_instrumentation__.Notify(611627)
		p.dependsOnContext = true
	} else {
		__antithesis_instrumentation__.Notify(611628)
	}
	__antithesis_instrumentation__.Notify(611615)
	p.result.D[p.tupleIdx] = d
	p.tupleIdx++
	return nil
}

func ParseDTupleFromString(
	ctx ParseTimeContext, s string, t *types.T,
) (_ *DTuple, dependsOnContext bool, _ error) {
	__antithesis_instrumentation__.Notify(611629)
	ret, dependsOnContext, err := doParseDTupleFromString(ctx, s, t)
	if err != nil {
		__antithesis_instrumentation__.Notify(611631)
		return ret, false, MakeParseError(s, t, err)
	} else {
		__antithesis_instrumentation__.Notify(611632)
	}
	__antithesis_instrumentation__.Notify(611630)
	return ret, dependsOnContext, nil
}

func doParseDTupleFromString(
	ctx ParseTimeContext, s string, t *types.T,
) (_ *DTuple, dependsOnContext bool, _ error) {
	__antithesis_instrumentation__.Notify(611633)
	if t.TupleContents() == nil {
		__antithesis_instrumentation__.Notify(611642)
		return nil, false, errors.AssertionFailedf("not a tuple type %s (%T)", t, t)
	} else {
		__antithesis_instrumentation__.Notify(611643)
	}
	__antithesis_instrumentation__.Notify(611634)
	if t == types.AnyTuple {
		__antithesis_instrumentation__.Notify(611644)
		return nil, false, unsupportedRecordError
	} else {
		__antithesis_instrumentation__.Notify(611645)
	}
	__antithesis_instrumentation__.Notify(611635)
	parser := tupleParseState{
		s:      s,
		ctx:    ctx,
		result: NewDTupleWithLen(t, len(t.TupleContents())),
		t:      t,
	}

	parser.eatWhitespace()
	if parser.peek() != '(' {
		__antithesis_instrumentation__.Notify(611646)
		return nil, false, enclosingRecordError
	} else {
		__antithesis_instrumentation__.Notify(611647)
	}
	__antithesis_instrumentation__.Notify(611636)
	parser.advance()
	if parser.peek() != ')' || func() bool {
		__antithesis_instrumentation__.Notify(611648)
		return len(t.TupleContents()) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(611649)
		if err := parser.parseElement(); err != nil {
			__antithesis_instrumentation__.Notify(611651)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(611652)
		}
		__antithesis_instrumentation__.Notify(611650)
		parser.eatWhitespace()
		for parser.peek() == ',' {
			__antithesis_instrumentation__.Notify(611653)
			parser.advance()
			if err := parser.parseElement(); err != nil {
				__antithesis_instrumentation__.Notify(611654)
				return nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(611655)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(611656)
	}
	__antithesis_instrumentation__.Notify(611637)
	parser.eatWhitespace()
	if parser.eof() {
		__antithesis_instrumentation__.Notify(611657)
		return nil, false, enclosingRecordError
	} else {
		__antithesis_instrumentation__.Notify(611658)
	}
	__antithesis_instrumentation__.Notify(611638)
	if parser.peek() != ')' {
		__antithesis_instrumentation__.Notify(611659)
		return nil, false, malformedRecordError
	} else {
		__antithesis_instrumentation__.Notify(611660)
	}
	__antithesis_instrumentation__.Notify(611639)
	if parser.tupleIdx < len(parser.t.TupleContents()) {
		__antithesis_instrumentation__.Notify(611661)
		return nil, false, errors.WithDetail(malformedRecordError, "Too few columns.")
	} else {
		__antithesis_instrumentation__.Notify(611662)
	}
	__antithesis_instrumentation__.Notify(611640)
	parser.advance()
	parser.eatWhitespace()
	if !parser.eof() {
		__antithesis_instrumentation__.Notify(611663)
		return nil, false, extraTextRecordError
	} else {
		__antithesis_instrumentation__.Notify(611664)
	}
	__antithesis_instrumentation__.Notify(611641)

	return parser.result, parser.dependsOnContext, nil
}
