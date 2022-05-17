package scanner

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"go/constant"
	"go/token"
	"strconv"
	"strings"
	"unicode/utf8"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
)

const eof = -1
const errUnterminated = "unterminated string"
const errInvalidUTF8 = "invalid UTF-8 byte sequence"
const errInvalidHexNumeric = "invalid hexadecimal numeric literal"
const singleQuote = '\''
const identQuote = '"'

var NewNumValFn = func(constant.Value, string, bool) interface{} {
	__antithesis_instrumentation__.Notify(576406)
	return struct{}{}
}

var NewPlaceholderFn = func(string) (interface{}, error) {
	__antithesis_instrumentation__.Notify(576407)
	return struct{}{}, nil
}

type ScanSymType interface {
	ID() int32
	SetID(int32)
	Pos() int32
	SetPos(int32)
	Str() string
	SetStr(string)
	UnionVal() interface{}
	SetUnionVal(interface{})
}

type Scanner struct {
	in            string
	pos           int
	bytesPrealloc []byte
}

func (s *Scanner) In() string {
	__antithesis_instrumentation__.Notify(576408)
	return s.in
}

func (s *Scanner) Pos() int {
	__antithesis_instrumentation__.Notify(576409)
	return s.pos
}

func (s *Scanner) Init(str string) {
	__antithesis_instrumentation__.Notify(576410)
	s.in = str
	s.pos = 0

	s.bytesPrealloc = make([]byte, len(str))
}

func (s *Scanner) Cleanup() {
	__antithesis_instrumentation__.Notify(576411)
	s.bytesPrealloc = nil
}

func (s *Scanner) allocBytes(length int) []byte {
	__antithesis_instrumentation__.Notify(576412)
	if len(s.bytesPrealloc) >= length {
		__antithesis_instrumentation__.Notify(576414)
		res := s.bytesPrealloc[:length:length]
		s.bytesPrealloc = s.bytesPrealloc[length:]
		return res
	} else {
		__antithesis_instrumentation__.Notify(576415)
	}
	__antithesis_instrumentation__.Notify(576413)
	return make([]byte, length)
}

func (s *Scanner) buffer() []byte {
	__antithesis_instrumentation__.Notify(576416)
	buf := s.bytesPrealloc[:0]
	s.bytesPrealloc = nil
	return buf
}

func (s *Scanner) returnBuffer(buf []byte) {
	__antithesis_instrumentation__.Notify(576417)
	if len(buf) < cap(buf) {
		__antithesis_instrumentation__.Notify(576418)
		s.bytesPrealloc = buf[len(buf):]
	} else {
		__antithesis_instrumentation__.Notify(576419)
	}
}

func (s *Scanner) finishString(buf []byte) string {
	__antithesis_instrumentation__.Notify(576420)
	str := *(*string)(unsafe.Pointer(&buf))
	s.returnBuffer(buf)
	return str
}

func (s *Scanner) Scan(lval ScanSymType) {
	__antithesis_instrumentation__.Notify(576421)
	lval.SetID(0)
	lval.SetPos(int32(s.pos))
	lval.SetStr("EOF")

	if _, ok := s.skipWhitespace(lval, true); !ok {
		__antithesis_instrumentation__.Notify(576424)
		return
	} else {
		__antithesis_instrumentation__.Notify(576425)
	}
	__antithesis_instrumentation__.Notify(576422)

	ch := s.next()
	if ch == eof {
		__antithesis_instrumentation__.Notify(576426)
		lval.SetPos(int32(s.pos))
		return
	} else {
		__antithesis_instrumentation__.Notify(576427)
	}
	__antithesis_instrumentation__.Notify(576423)

	lval.SetID(int32(ch))
	lval.SetPos(int32(s.pos - 1))
	lval.SetStr(s.in[lval.Pos():s.pos])

	switch ch {
	case '$':
		__antithesis_instrumentation__.Notify(576428)

		if lexbase.IsDigit(s.peek()) {
			__antithesis_instrumentation__.Notify(576471)
			s.scanPlaceholder(lval)
			return
		} else {
			__antithesis_instrumentation__.Notify(576472)
			if s.scanDollarQuotedString(lval) {
				__antithesis_instrumentation__.Notify(576473)
				lval.SetID(lexbase.SCONST)
				return
			} else {
				__antithesis_instrumentation__.Notify(576474)
			}
		}
		__antithesis_instrumentation__.Notify(576429)
		return

	case identQuote:
		__antithesis_instrumentation__.Notify(576430)

		if s.scanString(lval, identQuote, false, true) {
			__antithesis_instrumentation__.Notify(576475)
			lval.SetID(lexbase.IDENT)
		} else {
			__antithesis_instrumentation__.Notify(576476)
		}
		__antithesis_instrumentation__.Notify(576431)
		return

	case singleQuote:
		__antithesis_instrumentation__.Notify(576432)

		if s.scanString(lval, ch, false, true) {
			__antithesis_instrumentation__.Notify(576477)
			lval.SetID(lexbase.SCONST)
		} else {
			__antithesis_instrumentation__.Notify(576478)
		}
		__antithesis_instrumentation__.Notify(576433)
		return

	case 'b':
		__antithesis_instrumentation__.Notify(576434)

		if s.peek() == singleQuote {
			__antithesis_instrumentation__.Notify(576479)

			s.pos++
			if s.scanString(lval, singleQuote, true, false) {
				__antithesis_instrumentation__.Notify(576481)
				lval.SetID(lexbase.BCONST)
			} else {
				__antithesis_instrumentation__.Notify(576482)
			}
			__antithesis_instrumentation__.Notify(576480)
			return
		} else {
			__antithesis_instrumentation__.Notify(576483)
		}
		__antithesis_instrumentation__.Notify(576435)
		s.scanIdent(lval)
		return

	case 'r', 'R':
		__antithesis_instrumentation__.Notify(576436)
		s.scanIdent(lval)
		return

	case 'e', 'E':
		__antithesis_instrumentation__.Notify(576437)

		if s.peek() == singleQuote {
			__antithesis_instrumentation__.Notify(576484)

			s.pos++
			if s.scanString(lval, singleQuote, true, true) {
				__antithesis_instrumentation__.Notify(576486)
				lval.SetID(lexbase.SCONST)
			} else {
				__antithesis_instrumentation__.Notify(576487)
			}
			__antithesis_instrumentation__.Notify(576485)
			return
		} else {
			__antithesis_instrumentation__.Notify(576488)
		}
		__antithesis_instrumentation__.Notify(576438)
		s.scanIdent(lval)
		return

	case 'B':
		__antithesis_instrumentation__.Notify(576439)

		if s.peek() == singleQuote {
			__antithesis_instrumentation__.Notify(576489)

			s.pos++
			s.scanBitString(lval, singleQuote)
			return
		} else {
			__antithesis_instrumentation__.Notify(576490)
		}
		__antithesis_instrumentation__.Notify(576440)
		s.scanIdent(lval)
		return

	case 'x', 'X':
		__antithesis_instrumentation__.Notify(576441)

		if s.peek() == singleQuote {
			__antithesis_instrumentation__.Notify(576491)

			s.pos++
			s.scanHexString(lval, singleQuote)
			return
		} else {
			__antithesis_instrumentation__.Notify(576492)
		}
		__antithesis_instrumentation__.Notify(576442)
		s.scanIdent(lval)
		return

	case '.':
		__antithesis_instrumentation__.Notify(576443)
		switch t := s.peek(); {
		case t == '.':
			__antithesis_instrumentation__.Notify(576493)
			s.pos++
			lval.SetID(lexbase.DOT_DOT)
			return
		case lexbase.IsDigit(t):
			__antithesis_instrumentation__.Notify(576494)
			s.scanNumber(lval, ch)
			return
		default:
			__antithesis_instrumentation__.Notify(576495)
		}
		__antithesis_instrumentation__.Notify(576444)
		return

	case '!':
		__antithesis_instrumentation__.Notify(576445)
		switch s.peek() {
		case '=':
			__antithesis_instrumentation__.Notify(576496)
			s.pos++
			lval.SetID(lexbase.NOT_EQUALS)
			return
		case '~':
			__antithesis_instrumentation__.Notify(576497)
			s.pos++
			switch s.peek() {
			case '*':
				__antithesis_instrumentation__.Notify(576500)
				s.pos++
				lval.SetID(lexbase.NOT_REGIMATCH)
				return
			default:
				__antithesis_instrumentation__.Notify(576501)
			}
			__antithesis_instrumentation__.Notify(576498)
			lval.SetID(lexbase.NOT_REGMATCH)
			return
		default:
			__antithesis_instrumentation__.Notify(576499)
		}
		__antithesis_instrumentation__.Notify(576446)
		return

	case '?':
		__antithesis_instrumentation__.Notify(576447)
		switch s.peek() {
		case '?':
			__antithesis_instrumentation__.Notify(576502)
			s.pos++
			lval.SetID(lexbase.HELPTOKEN)
			return
		case '|':
			__antithesis_instrumentation__.Notify(576503)
			s.pos++
			lval.SetID(lexbase.JSON_SOME_EXISTS)
			return
		case '&':
			__antithesis_instrumentation__.Notify(576504)
			s.pos++
			lval.SetID(lexbase.JSON_ALL_EXISTS)
			return
		default:
			__antithesis_instrumentation__.Notify(576505)
		}
		__antithesis_instrumentation__.Notify(576448)
		return

	case '<':
		__antithesis_instrumentation__.Notify(576449)
		switch s.peek() {
		case '<':
			__antithesis_instrumentation__.Notify(576506)
			s.pos++
			switch s.peek() {
			case '=':
				__antithesis_instrumentation__.Notify(576512)
				s.pos++
				lval.SetID(lexbase.INET_CONTAINED_BY_OR_EQUALS)
				return
			default:
				__antithesis_instrumentation__.Notify(576513)
			}
			__antithesis_instrumentation__.Notify(576507)
			lval.SetID(lexbase.LSHIFT)
			return
		case '>':
			__antithesis_instrumentation__.Notify(576508)
			s.pos++
			lval.SetID(lexbase.NOT_EQUALS)
			return
		case '=':
			__antithesis_instrumentation__.Notify(576509)
			s.pos++
			lval.SetID(lexbase.LESS_EQUALS)
			return
		case '@':
			__antithesis_instrumentation__.Notify(576510)
			s.pos++
			lval.SetID(lexbase.CONTAINED_BY)
			return
		default:
			__antithesis_instrumentation__.Notify(576511)
		}
		__antithesis_instrumentation__.Notify(576450)
		return

	case '>':
		__antithesis_instrumentation__.Notify(576451)
		switch s.peek() {
		case '>':
			__antithesis_instrumentation__.Notify(576514)
			s.pos++
			switch s.peek() {
			case '=':
				__antithesis_instrumentation__.Notify(576518)
				s.pos++
				lval.SetID(lexbase.INET_CONTAINS_OR_EQUALS)
				return
			default:
				__antithesis_instrumentation__.Notify(576519)
			}
			__antithesis_instrumentation__.Notify(576515)
			lval.SetID(lexbase.RSHIFT)
			return
		case '=':
			__antithesis_instrumentation__.Notify(576516)
			s.pos++
			lval.SetID(lexbase.GREATER_EQUALS)
			return
		default:
			__antithesis_instrumentation__.Notify(576517)
		}
		__antithesis_instrumentation__.Notify(576452)
		return

	case ':':
		__antithesis_instrumentation__.Notify(576453)
		switch s.peek() {
		case ':':
			__antithesis_instrumentation__.Notify(576520)
			if s.peekN(1) == ':' {
				__antithesis_instrumentation__.Notify(576523)

				s.pos += 2
				lval.SetID(lexbase.TYPEANNOTATE)
				return
			} else {
				__antithesis_instrumentation__.Notify(576524)
			}
			__antithesis_instrumentation__.Notify(576521)
			s.pos++
			lval.SetID(lexbase.TYPECAST)
			return
		default:
			__antithesis_instrumentation__.Notify(576522)
		}
		__antithesis_instrumentation__.Notify(576454)
		return

	case '|':
		__antithesis_instrumentation__.Notify(576455)
		switch s.peek() {
		case '|':
			__antithesis_instrumentation__.Notify(576525)
			s.pos++
			switch s.peek() {
			case '/':
				__antithesis_instrumentation__.Notify(576529)
				s.pos++
				lval.SetID(lexbase.CBRT)
				return
			default:
				__antithesis_instrumentation__.Notify(576530)
			}
			__antithesis_instrumentation__.Notify(576526)
			lval.SetID(lexbase.CONCAT)
			return
		case '/':
			__antithesis_instrumentation__.Notify(576527)
			s.pos++
			lval.SetID(lexbase.SQRT)
			return
		default:
			__antithesis_instrumentation__.Notify(576528)
		}
		__antithesis_instrumentation__.Notify(576456)
		return

	case '/':
		__antithesis_instrumentation__.Notify(576457)
		switch s.peek() {
		case '/':
			__antithesis_instrumentation__.Notify(576531)
			s.pos++
			lval.SetID(lexbase.FLOORDIV)
			return
		default:
			__antithesis_instrumentation__.Notify(576532)
		}
		__antithesis_instrumentation__.Notify(576458)
		return

	case '~':
		__antithesis_instrumentation__.Notify(576459)
		switch s.peek() {
		case '*':
			__antithesis_instrumentation__.Notify(576533)
			s.pos++
			lval.SetID(lexbase.REGIMATCH)
			return
		default:
			__antithesis_instrumentation__.Notify(576534)
		}
		__antithesis_instrumentation__.Notify(576460)
		return

	case '@':
		__antithesis_instrumentation__.Notify(576461)
		switch s.peek() {
		case '>':
			__antithesis_instrumentation__.Notify(576535)
			s.pos++
			lval.SetID(lexbase.CONTAINS)
			return
		default:
			__antithesis_instrumentation__.Notify(576536)
		}
		__antithesis_instrumentation__.Notify(576462)
		return

	case '&':
		__antithesis_instrumentation__.Notify(576463)
		switch s.peek() {
		case '&':
			__antithesis_instrumentation__.Notify(576537)
			s.pos++
			lval.SetID(lexbase.AND_AND)
			return
		default:
			__antithesis_instrumentation__.Notify(576538)
		}
		__antithesis_instrumentation__.Notify(576464)
		return

	case '-':
		__antithesis_instrumentation__.Notify(576465)
		switch s.peek() {
		case '>':
			__antithesis_instrumentation__.Notify(576539)
			if s.peekN(1) == '>' {
				__antithesis_instrumentation__.Notify(576542)

				s.pos += 2
				lval.SetID(lexbase.FETCHTEXT)
				return
			} else {
				__antithesis_instrumentation__.Notify(576543)
			}
			__antithesis_instrumentation__.Notify(576540)
			s.pos++
			lval.SetID(lexbase.FETCHVAL)
			return
		default:
			__antithesis_instrumentation__.Notify(576541)
		}
		__antithesis_instrumentation__.Notify(576466)
		return

	case '#':
		__antithesis_instrumentation__.Notify(576467)
		switch s.peek() {
		case '>':
			__antithesis_instrumentation__.Notify(576544)
			if s.peekN(1) == '>' {
				__antithesis_instrumentation__.Notify(576548)

				s.pos += 2
				lval.SetID(lexbase.FETCHTEXT_PATH)
				return
			} else {
				__antithesis_instrumentation__.Notify(576549)
			}
			__antithesis_instrumentation__.Notify(576545)
			s.pos++
			lval.SetID(lexbase.FETCHVAL_PATH)
			return
		case '-':
			__antithesis_instrumentation__.Notify(576546)
			s.pos++
			lval.SetID(lexbase.REMOVE_PATH)
			return
		default:
			__antithesis_instrumentation__.Notify(576547)
		}
		__antithesis_instrumentation__.Notify(576468)
		return

	default:
		__antithesis_instrumentation__.Notify(576469)
		if lexbase.IsDigit(ch) {
			__antithesis_instrumentation__.Notify(576550)
			s.scanNumber(lval, ch)
			return
		} else {
			__antithesis_instrumentation__.Notify(576551)
		}
		__antithesis_instrumentation__.Notify(576470)
		if lexbase.IsIdentStart(ch) {
			__antithesis_instrumentation__.Notify(576552)
			s.scanIdent(lval)
			return
		} else {
			__antithesis_instrumentation__.Notify(576553)
		}
	}

}

func (s *Scanner) peek() int {
	__antithesis_instrumentation__.Notify(576554)
	if s.pos >= len(s.in) {
		__antithesis_instrumentation__.Notify(576556)
		return eof
	} else {
		__antithesis_instrumentation__.Notify(576557)
	}
	__antithesis_instrumentation__.Notify(576555)
	return int(s.in[s.pos])
}

func (s *Scanner) peekN(n int) int {
	__antithesis_instrumentation__.Notify(576558)
	pos := s.pos + n
	if pos >= len(s.in) {
		__antithesis_instrumentation__.Notify(576560)
		return eof
	} else {
		__antithesis_instrumentation__.Notify(576561)
	}
	__antithesis_instrumentation__.Notify(576559)
	return int(s.in[pos])
}

func (s *Scanner) next() int {
	__antithesis_instrumentation__.Notify(576562)
	ch := s.peek()
	if ch != eof {
		__antithesis_instrumentation__.Notify(576564)
		s.pos++
	} else {
		__antithesis_instrumentation__.Notify(576565)
	}
	__antithesis_instrumentation__.Notify(576563)
	return ch
}

func (s *Scanner) skipWhitespace(lval ScanSymType, allowComments bool) (newline, ok bool) {
	__antithesis_instrumentation__.Notify(576566)
	newline = false
	for {
		__antithesis_instrumentation__.Notify(576568)
		ch := s.peek()
		if ch == '\n' {
			__antithesis_instrumentation__.Notify(576572)
			s.pos++
			newline = true
			continue
		} else {
			__antithesis_instrumentation__.Notify(576573)
		}
		__antithesis_instrumentation__.Notify(576569)
		if ch == ' ' || func() bool {
			__antithesis_instrumentation__.Notify(576574)
			return ch == '\t' == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(576575)
			return ch == '\r' == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(576576)
			return ch == '\f' == true
		}() == true {
			__antithesis_instrumentation__.Notify(576577)
			s.pos++
			continue
		} else {
			__antithesis_instrumentation__.Notify(576578)
		}
		__antithesis_instrumentation__.Notify(576570)
		if allowComments {
			__antithesis_instrumentation__.Notify(576579)
			if present, cok := s.ScanComment(lval); !cok {
				__antithesis_instrumentation__.Notify(576580)
				return false, false
			} else {
				__antithesis_instrumentation__.Notify(576581)
				if present {
					__antithesis_instrumentation__.Notify(576582)
					continue
				} else {
					__antithesis_instrumentation__.Notify(576583)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(576584)
		}
		__antithesis_instrumentation__.Notify(576571)
		break
	}
	__antithesis_instrumentation__.Notify(576567)
	return newline, true
}

func (s *Scanner) ScanComment(lval ScanSymType) (present, ok bool) {
	__antithesis_instrumentation__.Notify(576585)
	start := s.pos
	ch := s.peek()

	if ch == '/' {
		__antithesis_instrumentation__.Notify(576588)
		s.pos++
		if s.peek() != '*' {
			__antithesis_instrumentation__.Notify(576590)
			s.pos--
			return false, true
		} else {
			__antithesis_instrumentation__.Notify(576591)
		}
		__antithesis_instrumentation__.Notify(576589)
		s.pos++
		depth := 1
		for {
			__antithesis_instrumentation__.Notify(576592)
			switch s.next() {
			case '*':
				__antithesis_instrumentation__.Notify(576593)
				if s.peek() == '/' {
					__antithesis_instrumentation__.Notify(576597)
					s.pos++
					depth--
					if depth == 0 {
						__antithesis_instrumentation__.Notify(576599)
						return true, true
					} else {
						__antithesis_instrumentation__.Notify(576600)
					}
					__antithesis_instrumentation__.Notify(576598)
					continue
				} else {
					__antithesis_instrumentation__.Notify(576601)
				}

			case '/':
				__antithesis_instrumentation__.Notify(576594)
				if s.peek() == '*' {
					__antithesis_instrumentation__.Notify(576602)
					s.pos++
					depth++
					continue
				} else {
					__antithesis_instrumentation__.Notify(576603)
				}

			case eof:
				__antithesis_instrumentation__.Notify(576595)
				lval.SetID(lexbase.ERROR)
				lval.SetPos(int32(start))
				lval.SetStr("unterminated comment")
				return false, false
			default:
				__antithesis_instrumentation__.Notify(576596)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(576604)
	}
	__antithesis_instrumentation__.Notify(576586)

	if ch == '-' {
		__antithesis_instrumentation__.Notify(576605)
		s.pos++
		if s.peek() != '-' {
			__antithesis_instrumentation__.Notify(576607)
			s.pos--
			return false, true
		} else {
			__antithesis_instrumentation__.Notify(576608)
		}
		__antithesis_instrumentation__.Notify(576606)
		for {
			__antithesis_instrumentation__.Notify(576609)
			switch s.next() {
			case eof, '\n':
				__antithesis_instrumentation__.Notify(576610)
				return true, true
			default:
				__antithesis_instrumentation__.Notify(576611)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(576612)
	}
	__antithesis_instrumentation__.Notify(576587)

	return false, true
}

func (s *Scanner) scanIdent(lval ScanSymType) {
	__antithesis_instrumentation__.Notify(576613)
	s.pos--
	start := s.pos
	isASCII := true
	isLower := true

	for {
		__antithesis_instrumentation__.Notify(576617)
		ch := s.peek()

		if ch >= utf8.RuneSelf {
			__antithesis_instrumentation__.Notify(576620)
			isASCII = false
		} else {
			__antithesis_instrumentation__.Notify(576621)
			if ch >= 'A' && func() bool {
				__antithesis_instrumentation__.Notify(576622)
				return ch <= 'Z' == true
			}() == true {
				__antithesis_instrumentation__.Notify(576623)
				isLower = false
			} else {
				__antithesis_instrumentation__.Notify(576624)
			}
		}
		__antithesis_instrumentation__.Notify(576618)

		if !lexbase.IsIdentMiddle(ch) {
			__antithesis_instrumentation__.Notify(576625)
			break
		} else {
			__antithesis_instrumentation__.Notify(576626)
		}
		__antithesis_instrumentation__.Notify(576619)

		s.pos++
	}
	__antithesis_instrumentation__.Notify(576614)

	if isLower && func() bool {
		__antithesis_instrumentation__.Notify(576627)
		return isASCII == true
	}() == true {
		__antithesis_instrumentation__.Notify(576628)

		lval.SetStr(s.in[start:s.pos])
	} else {
		__antithesis_instrumentation__.Notify(576629)
		if isASCII {
			__antithesis_instrumentation__.Notify(576630)

			b := s.allocBytes(s.pos - start)
			_ = b[s.pos-start-1]
			for i, c := range s.in[start:s.pos] {
				__antithesis_instrumentation__.Notify(576632)
				if c >= 'A' && func() bool {
					__antithesis_instrumentation__.Notify(576634)
					return c <= 'Z' == true
				}() == true {
					__antithesis_instrumentation__.Notify(576635)
					c += 'a' - 'A'
				} else {
					__antithesis_instrumentation__.Notify(576636)
				}
				__antithesis_instrumentation__.Notify(576633)
				b[i] = byte(c)
			}
			__antithesis_instrumentation__.Notify(576631)
			lval.SetStr(*(*string)(unsafe.Pointer(&b)))
		} else {
			__antithesis_instrumentation__.Notify(576637)

			lval.SetStr(lexbase.NormalizeName(s.in[start:s.pos]))
		}
	}
	__antithesis_instrumentation__.Notify(576615)

	isExperimental := false
	kw := lval.Str()
	switch {
	case strings.HasPrefix(lval.Str(), "experimental_"):
		__antithesis_instrumentation__.Notify(576638)
		kw = lval.Str()[13:]
		isExperimental = true
	case strings.HasPrefix(lval.Str(), "testing_"):
		__antithesis_instrumentation__.Notify(576639)
		kw = lval.Str()[8:]
		isExperimental = true
	default:
		__antithesis_instrumentation__.Notify(576640)
	}
	__antithesis_instrumentation__.Notify(576616)
	lval.SetID(lexbase.GetKeywordID(kw))
	if lval.ID() != lexbase.IDENT {
		__antithesis_instrumentation__.Notify(576641)
		if isExperimental {
			__antithesis_instrumentation__.Notify(576642)
			if _, ok := lexbase.AllowedExperimental[kw]; !ok {
				__antithesis_instrumentation__.Notify(576643)

				lval.SetID(lexbase.GetKeywordID(lval.Str()))
			} else {
				__antithesis_instrumentation__.Notify(576644)

				lval.SetStr(kw)
			}
		} else {
			__antithesis_instrumentation__.Notify(576645)
		}
	} else {
		__antithesis_instrumentation__.Notify(576646)

		lval.SetID(lexbase.GetKeywordID(lval.Str()))
	}
}

func (s *Scanner) scanNumber(lval ScanSymType, ch int) {
	__antithesis_instrumentation__.Notify(576647)
	start := s.pos - 1
	isHex := false
	hasDecimal := ch == '.'
	hasExponent := false

	for {
		__antithesis_instrumentation__.Notify(576649)
		ch := s.peek()
		if (isHex && func() bool {
			__antithesis_instrumentation__.Notify(576655)
			return lexbase.IsHexDigit(ch) == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(576656)
			return lexbase.IsDigit(ch) == true
		}() == true {
			__antithesis_instrumentation__.Notify(576657)
			s.pos++
			continue
		} else {
			__antithesis_instrumentation__.Notify(576658)
		}
		__antithesis_instrumentation__.Notify(576650)
		if ch == 'x' || func() bool {
			__antithesis_instrumentation__.Notify(576659)
			return ch == 'X' == true
		}() == true {
			__antithesis_instrumentation__.Notify(576660)
			if isHex || func() bool {
				__antithesis_instrumentation__.Notify(576662)
				return s.in[start] != '0' == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(576663)
				return s.pos != start+1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(576664)
				lval.SetID(lexbase.ERROR)
				lval.SetStr(errInvalidHexNumeric)
				return
			} else {
				__antithesis_instrumentation__.Notify(576665)
			}
			__antithesis_instrumentation__.Notify(576661)
			s.pos++
			isHex = true
			continue
		} else {
			__antithesis_instrumentation__.Notify(576666)
		}
		__antithesis_instrumentation__.Notify(576651)
		if isHex {
			__antithesis_instrumentation__.Notify(576667)
			break
		} else {
			__antithesis_instrumentation__.Notify(576668)
		}
		__antithesis_instrumentation__.Notify(576652)
		if ch == '.' {
			__antithesis_instrumentation__.Notify(576669)
			if hasDecimal || func() bool {
				__antithesis_instrumentation__.Notify(576672)
				return hasExponent == true
			}() == true {
				__antithesis_instrumentation__.Notify(576673)
				break
			} else {
				__antithesis_instrumentation__.Notify(576674)
			}
			__antithesis_instrumentation__.Notify(576670)
			s.pos++
			if s.peek() == '.' {
				__antithesis_instrumentation__.Notify(576675)

				s.pos--
				break
			} else {
				__antithesis_instrumentation__.Notify(576676)
			}
			__antithesis_instrumentation__.Notify(576671)
			hasDecimal = true
			continue
		} else {
			__antithesis_instrumentation__.Notify(576677)
		}
		__antithesis_instrumentation__.Notify(576653)
		if ch == 'e' || func() bool {
			__antithesis_instrumentation__.Notify(576678)
			return ch == 'E' == true
		}() == true {
			__antithesis_instrumentation__.Notify(576679)
			if hasExponent {
				__antithesis_instrumentation__.Notify(576683)
				break
			} else {
				__antithesis_instrumentation__.Notify(576684)
			}
			__antithesis_instrumentation__.Notify(576680)
			hasExponent = true
			s.pos++
			ch = s.peek()
			if ch == '-' || func() bool {
				__antithesis_instrumentation__.Notify(576685)
				return ch == '+' == true
			}() == true {
				__antithesis_instrumentation__.Notify(576686)
				s.pos++
			} else {
				__antithesis_instrumentation__.Notify(576687)
			}
			__antithesis_instrumentation__.Notify(576681)
			ch = s.peek()
			if !lexbase.IsDigit(ch) {
				__antithesis_instrumentation__.Notify(576688)
				lval.SetID(lexbase.ERROR)
				lval.SetStr("invalid floating point literal")
				return
			} else {
				__antithesis_instrumentation__.Notify(576689)
			}
			__antithesis_instrumentation__.Notify(576682)
			continue
		} else {
			__antithesis_instrumentation__.Notify(576690)
		}
		__antithesis_instrumentation__.Notify(576654)
		break
	}
	__antithesis_instrumentation__.Notify(576648)

	lval.SetStr(s.in[start:s.pos])
	if hasDecimal || func() bool {
		__antithesis_instrumentation__.Notify(576691)
		return hasExponent == true
	}() == true {
		__antithesis_instrumentation__.Notify(576692)
		lval.SetID(lexbase.FCONST)
		floatConst := constant.MakeFromLiteral(lval.Str(), token.FLOAT, 0)
		if floatConst.Kind() == constant.Unknown {
			__antithesis_instrumentation__.Notify(576694)
			lval.SetID(lexbase.ERROR)
			lval.SetStr(fmt.Sprintf("could not make constant float from literal %q", lval.Str()))
			return
		} else {
			__antithesis_instrumentation__.Notify(576695)
		}
		__antithesis_instrumentation__.Notify(576693)
		lval.SetUnionVal(NewNumValFn(floatConst, lval.Str(), false))
	} else {
		__antithesis_instrumentation__.Notify(576696)
		if isHex && func() bool {
			__antithesis_instrumentation__.Notify(576700)
			return s.pos == start+2 == true
		}() == true {
			__antithesis_instrumentation__.Notify(576701)
			lval.SetID(lexbase.ERROR)
			lval.SetStr(errInvalidHexNumeric)
			return
		} else {
			__antithesis_instrumentation__.Notify(576702)
		}
		__antithesis_instrumentation__.Notify(576697)

		if !isHex {
			__antithesis_instrumentation__.Notify(576703)
			for len(lval.Str()) > 1 && func() bool {
				__antithesis_instrumentation__.Notify(576704)
				return lval.Str()[0] == '0' == true
			}() == true {
				__antithesis_instrumentation__.Notify(576705)
				lval.SetStr(lval.Str()[1:])
			}
		} else {
			__antithesis_instrumentation__.Notify(576706)
		}
		__antithesis_instrumentation__.Notify(576698)

		lval.SetID(lexbase.ICONST)
		intConst := constant.MakeFromLiteral(lval.Str(), token.INT, 0)
		if intConst.Kind() == constant.Unknown {
			__antithesis_instrumentation__.Notify(576707)
			lval.SetID(lexbase.ERROR)
			lval.SetStr(fmt.Sprintf("could not make constant int from literal %q", lval.Str()))
			return
		} else {
			__antithesis_instrumentation__.Notify(576708)
		}
		__antithesis_instrumentation__.Notify(576699)
		lval.SetUnionVal(NewNumValFn(intConst, lval.Str(), false))
	}
}

func (s *Scanner) scanPlaceholder(lval ScanSymType) {
	__antithesis_instrumentation__.Notify(576709)
	start := s.pos
	for lexbase.IsDigit(s.peek()) {
		__antithesis_instrumentation__.Notify(576712)
		s.pos++
	}
	__antithesis_instrumentation__.Notify(576710)
	lval.SetStr(s.in[start:s.pos])

	placeholder, err := NewPlaceholderFn(lval.Str())
	if err != nil {
		__antithesis_instrumentation__.Notify(576713)
		lval.SetID(lexbase.ERROR)
		lval.SetStr(err.Error())
		return
	} else {
		__antithesis_instrumentation__.Notify(576714)
	}
	__antithesis_instrumentation__.Notify(576711)
	lval.SetID(lexbase.PLACEHOLDER)
	lval.SetUnionVal(placeholder)
}

func (s *Scanner) scanHexString(lval ScanSymType, ch int) bool {
	__antithesis_instrumentation__.Notify(576715)
	buf := s.buffer()

	var curbyte byte
	bytep := 0
	const errInvalidBytesLiteral = "invalid hexadecimal bytes literal"
outer:
	for {
		__antithesis_instrumentation__.Notify(576718)
		b := s.next()
		switch b {
		case ch:
			__antithesis_instrumentation__.Notify(576720)
			newline, ok := s.skipWhitespace(lval, false)
			if !ok {
				__antithesis_instrumentation__.Notify(576727)
				return false
			} else {
				__antithesis_instrumentation__.Notify(576728)
			}
			__antithesis_instrumentation__.Notify(576721)

			if s.peek() == ch && func() bool {
				__antithesis_instrumentation__.Notify(576729)
				return newline == true
			}() == true {
				__antithesis_instrumentation__.Notify(576730)
				s.pos++
				continue
			} else {
				__antithesis_instrumentation__.Notify(576731)
			}
			__antithesis_instrumentation__.Notify(576722)
			break outer

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			__antithesis_instrumentation__.Notify(576723)
			curbyte = (curbyte << 4) | byte(b-'0')
		case 'a', 'b', 'c', 'd', 'e', 'f':
			__antithesis_instrumentation__.Notify(576724)
			curbyte = (curbyte << 4) | byte(b-'a'+10)
		case 'A', 'B', 'C', 'D', 'E', 'F':
			__antithesis_instrumentation__.Notify(576725)
			curbyte = (curbyte << 4) | byte(b-'A'+10)
		default:
			__antithesis_instrumentation__.Notify(576726)
			lval.SetID(lexbase.ERROR)
			lval.SetStr(errInvalidBytesLiteral)
			return false
		}
		__antithesis_instrumentation__.Notify(576719)
		bytep++

		if bytep > 1 {
			__antithesis_instrumentation__.Notify(576732)
			buf = append(buf, curbyte)
			bytep = 0
			curbyte = 0
		} else {
			__antithesis_instrumentation__.Notify(576733)
		}
	}
	__antithesis_instrumentation__.Notify(576716)

	if bytep != 0 {
		__antithesis_instrumentation__.Notify(576734)
		lval.SetID(lexbase.ERROR)
		lval.SetStr(errInvalidBytesLiteral)
		return false
	} else {
		__antithesis_instrumentation__.Notify(576735)
	}
	__antithesis_instrumentation__.Notify(576717)

	lval.SetID(lexbase.BCONST)
	lval.SetStr(s.finishString(buf))
	return true
}

func (s *Scanner) scanBitString(lval ScanSymType, ch int) bool {
	__antithesis_instrumentation__.Notify(576736)
	buf := s.buffer()
outer:
	for {
		__antithesis_instrumentation__.Notify(576738)
		b := s.next()
		switch b {
		case ch:
			__antithesis_instrumentation__.Notify(576739)
			newline, ok := s.skipWhitespace(lval, false)
			if !ok {
				__antithesis_instrumentation__.Notify(576744)
				return false
			} else {
				__antithesis_instrumentation__.Notify(576745)
			}
			__antithesis_instrumentation__.Notify(576740)

			if s.peek() == ch && func() bool {
				__antithesis_instrumentation__.Notify(576746)
				return newline == true
			}() == true {
				__antithesis_instrumentation__.Notify(576747)
				s.pos++
				continue
			} else {
				__antithesis_instrumentation__.Notify(576748)
			}
			__antithesis_instrumentation__.Notify(576741)
			break outer

		case '0', '1':
			__antithesis_instrumentation__.Notify(576742)
			buf = append(buf, byte(b))
		default:
			__antithesis_instrumentation__.Notify(576743)
			lval.SetID(lexbase.ERROR)
			lval.SetStr(fmt.Sprintf(`"%c" is not a valid binary digit`, rune(b)))
			return false
		}
	}
	__antithesis_instrumentation__.Notify(576737)

	lval.SetID(lexbase.BITCONST)
	lval.SetStr(s.finishString(buf))
	return true
}

func (s *Scanner) scanString(lval ScanSymType, ch int, allowEscapes, requireUTF8 bool) bool {
	__antithesis_instrumentation__.Notify(576749)
	buf := s.buffer()
	var runeTmp [utf8.UTFMax]byte
	start := s.pos
outer:
	for {
		__antithesis_instrumentation__.Notify(576753)
		switch s.next() {
		case ch:
			__antithesis_instrumentation__.Notify(576754)
			buf = append(buf, s.in[start:s.pos-1]...)
			if s.peek() == ch {
				__antithesis_instrumentation__.Notify(576761)

				start = s.pos
				s.pos++
				continue
			} else {
				__antithesis_instrumentation__.Notify(576762)
			}
			__antithesis_instrumentation__.Notify(576755)

			newline, ok := s.skipWhitespace(lval, false)
			if !ok {
				__antithesis_instrumentation__.Notify(576763)
				return false
			} else {
				__antithesis_instrumentation__.Notify(576764)
			}
			__antithesis_instrumentation__.Notify(576756)

			if s.peek() == ch && func() bool {
				__antithesis_instrumentation__.Notify(576765)
				return newline == true
			}() == true {
				__antithesis_instrumentation__.Notify(576766)
				s.pos++
				start = s.pos
				continue
			} else {
				__antithesis_instrumentation__.Notify(576767)
			}
			__antithesis_instrumentation__.Notify(576757)
			break outer

		case '\\':
			__antithesis_instrumentation__.Notify(576758)
			t := s.peek()

			if allowEscapes {
				__antithesis_instrumentation__.Notify(576768)
				buf = append(buf, s.in[start:s.pos-1]...)
				if t == ch {
					__antithesis_instrumentation__.Notify(576771)
					start = s.pos
					s.pos++
					continue
				} else {
					__antithesis_instrumentation__.Notify(576772)
				}
				__antithesis_instrumentation__.Notify(576769)

				switch t {
				case 'a', 'b', 'f', 'n', 'r', 't', 'v', 'x', 'X', 'u', 'U', '\\',
					'0', '1', '2', '3', '4', '5', '6', '7':
					__antithesis_instrumentation__.Notify(576773)
					var tmp string
					if t == 'X' && func() bool {
						__antithesis_instrumentation__.Notify(576778)
						return len(s.in[s.pos:]) >= 3 == true
					}() == true {
						__antithesis_instrumentation__.Notify(576779)

						tmp = "\\x" + s.in[s.pos+1:s.pos+3]
					} else {
						__antithesis_instrumentation__.Notify(576780)
						tmp = s.in[s.pos-1:]
					}
					__antithesis_instrumentation__.Notify(576774)
					v, multibyte, tail, err := strconv.UnquoteChar(tmp, byte(ch))
					if err != nil {
						__antithesis_instrumentation__.Notify(576781)
						lval.SetID(lexbase.ERROR)
						lval.SetStr(err.Error())
						return false
					} else {
						__antithesis_instrumentation__.Notify(576782)
					}
					__antithesis_instrumentation__.Notify(576775)
					if v < utf8.RuneSelf || func() bool {
						__antithesis_instrumentation__.Notify(576783)
						return !multibyte == true
					}() == true {
						__antithesis_instrumentation__.Notify(576784)
						buf = append(buf, byte(v))
					} else {
						__antithesis_instrumentation__.Notify(576785)
						n := utf8.EncodeRune(runeTmp[:], v)
						buf = append(buf, runeTmp[:n]...)
					}
					__antithesis_instrumentation__.Notify(576776)
					s.pos += len(tmp) - len(tail) - 1
					start = s.pos
					continue
				default:
					__antithesis_instrumentation__.Notify(576777)
				}
				__antithesis_instrumentation__.Notify(576770)

				start = s.pos
			} else {
				__antithesis_instrumentation__.Notify(576786)
			}

		case eof:
			__antithesis_instrumentation__.Notify(576759)
			lval.SetID(lexbase.ERROR)
			lval.SetStr(errUnterminated)
			return false
		default:
			__antithesis_instrumentation__.Notify(576760)
		}
	}
	__antithesis_instrumentation__.Notify(576750)

	if requireUTF8 && func() bool {
		__antithesis_instrumentation__.Notify(576787)
		return !utf8.Valid(buf) == true
	}() == true {
		__antithesis_instrumentation__.Notify(576788)
		lval.SetID(lexbase.ERROR)
		lval.SetStr(errInvalidUTF8)
		return false
	} else {
		__antithesis_instrumentation__.Notify(576789)
	}
	__antithesis_instrumentation__.Notify(576751)

	if ch == identQuote {
		__antithesis_instrumentation__.Notify(576790)
		lval.SetStr(lexbase.NormalizeString(s.finishString(buf)))
	} else {
		__antithesis_instrumentation__.Notify(576791)
		lval.SetStr(s.finishString(buf))
	}
	__antithesis_instrumentation__.Notify(576752)
	return true
}

func (s *Scanner) scanDollarQuotedString(lval ScanSymType) bool {
	__antithesis_instrumentation__.Notify(576792)
	buf := s.buffer()
	start := s.pos

	foundStartTag := false
	possibleEndTag := false
	startTagIndex := -1
	var startTag string

outer:
	for {
		__antithesis_instrumentation__.Notify(576795)
		ch := s.peek()
		switch ch {
		case '$':
			__antithesis_instrumentation__.Notify(576796)
			s.pos++
			if foundStartTag {
				__antithesis_instrumentation__.Notify(576801)
				if possibleEndTag {
					__antithesis_instrumentation__.Notify(576802)
					if len(startTag) == startTagIndex {
						__antithesis_instrumentation__.Notify(576803)

						buf = append(buf, s.in[start+len(startTag)+1:s.pos-len(startTag)-2]...)
						break outer
					} else {
						__antithesis_instrumentation__.Notify(576804)

						startTagIndex = 0
					}
				} else {
					__antithesis_instrumentation__.Notify(576805)
					possibleEndTag = true
					startTagIndex = 0
				}
			} else {
				__antithesis_instrumentation__.Notify(576806)
				startTag = s.in[start : s.pos-1]
				foundStartTag = true
			}

		case eof:
			__antithesis_instrumentation__.Notify(576797)
			if foundStartTag {
				__antithesis_instrumentation__.Notify(576807)

				lval.SetID(lexbase.ERROR)
				lval.SetStr(errUnterminated)
			} else {
				__antithesis_instrumentation__.Notify(576808)

				s.pos = start
			}
			__antithesis_instrumentation__.Notify(576798)
			return false

		default:
			__antithesis_instrumentation__.Notify(576799)

			if !foundStartTag && func() bool {
				__antithesis_instrumentation__.Notify(576809)
				return !lexbase.IsIdentStart(ch) == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(576810)
				return !lexbase.IsDigit(ch) == true
			}() == true {
				__antithesis_instrumentation__.Notify(576811)
				return false
			} else {
				__antithesis_instrumentation__.Notify(576812)
			}
			__antithesis_instrumentation__.Notify(576800)
			s.pos++
			if possibleEndTag {
				__antithesis_instrumentation__.Notify(576813)

				if startTagIndex >= len(startTag) || func() bool {
					__antithesis_instrumentation__.Notify(576814)
					return ch != int(startTag[startTagIndex]) == true
				}() == true {
					__antithesis_instrumentation__.Notify(576815)

					possibleEndTag = false
					startTagIndex = -1
				} else {
					__antithesis_instrumentation__.Notify(576816)
					startTagIndex++
				}
			} else {
				__antithesis_instrumentation__.Notify(576817)
			}
		}
	}
	__antithesis_instrumentation__.Notify(576793)

	if !utf8.Valid(buf) {
		__antithesis_instrumentation__.Notify(576818)
		lval.SetID(lexbase.ERROR)
		lval.SetStr(errInvalidUTF8)
		return false
	} else {
		__antithesis_instrumentation__.Notify(576819)
	}
	__antithesis_instrumentation__.Notify(576794)

	lval.SetStr(s.finishString(buf))
	return true
}

func HasMultipleStatements(sql string) (multipleStmt bool, err error) {
	__antithesis_instrumentation__.Notify(576820)
	var s Scanner
	var lval fakeSym
	s.Init(sql)
	count := 0
	for {
		__antithesis_instrumentation__.Notify(576822)
		done, hasToks, err := s.scanOne(&lval)
		if err != nil {
			__antithesis_instrumentation__.Notify(576825)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(576826)
		}
		__antithesis_instrumentation__.Notify(576823)
		if hasToks {
			__antithesis_instrumentation__.Notify(576827)
			count++
		} else {
			__antithesis_instrumentation__.Notify(576828)
		}
		__antithesis_instrumentation__.Notify(576824)
		if done || func() bool {
			__antithesis_instrumentation__.Notify(576829)
			return count > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(576830)
			break
		} else {
			__antithesis_instrumentation__.Notify(576831)
		}
	}
	__antithesis_instrumentation__.Notify(576821)
	return count > 1, nil
}

func (s *Scanner) scanOne(lval *fakeSym) (done, hasToks bool, err error) {
	__antithesis_instrumentation__.Notify(576832)

	for {
		__antithesis_instrumentation__.Notify(576834)
		s.Scan(lval)
		if lval.id == 0 {
			__antithesis_instrumentation__.Notify(576836)
			return true, false, nil
		} else {
			__antithesis_instrumentation__.Notify(576837)
		}
		__antithesis_instrumentation__.Notify(576835)
		if lval.id != ';' {
			__antithesis_instrumentation__.Notify(576838)
			break
		} else {
			__antithesis_instrumentation__.Notify(576839)
		}
	}
	__antithesis_instrumentation__.Notify(576833)

	for {
		__antithesis_instrumentation__.Notify(576840)
		if lval.id == lexbase.ERROR {
			__antithesis_instrumentation__.Notify(576842)
			return true, true, fmt.Errorf("scan error: %s", lval.s)
		} else {
			__antithesis_instrumentation__.Notify(576843)
		}
		__antithesis_instrumentation__.Notify(576841)
		s.Scan(lval)
		if lval.id == 0 || func() bool {
			__antithesis_instrumentation__.Notify(576844)
			return lval.id == ';' == true
		}() == true {
			__antithesis_instrumentation__.Notify(576845)
			return (lval.id == 0), true, nil
		} else {
			__antithesis_instrumentation__.Notify(576846)
		}
	}
}

func LastLexicalToken(sql string) (lastTok int, ok bool) {
	__antithesis_instrumentation__.Notify(576847)
	var s Scanner
	var lval fakeSym
	s.Init(sql)
	for {
		__antithesis_instrumentation__.Notify(576848)
		last := lval.ID()
		s.Scan(&lval)
		if lval.ID() == 0 {
			__antithesis_instrumentation__.Notify(576849)
			return int(last), last != 0
		} else {
			__antithesis_instrumentation__.Notify(576850)
		}
	}
}

func FirstLexicalToken(sql string) (tok int) {
	__antithesis_instrumentation__.Notify(576851)
	var s Scanner
	var lval fakeSym
	s.Init(sql)
	s.Scan(&lval)
	id := lval.ID()
	return int(id)
}

type fakeSym struct {
	id  int32
	pos int32
	s   string
}

var _ ScanSymType = (*fakeSym)(nil)

func (s fakeSym) ID() int32                 { __antithesis_instrumentation__.Notify(576852); return s.id }
func (s *fakeSym) SetID(id int32)           { __antithesis_instrumentation__.Notify(576853); s.id = id }
func (s fakeSym) Pos() int32                { __antithesis_instrumentation__.Notify(576854); return s.pos }
func (s *fakeSym) SetPos(p int32)           { __antithesis_instrumentation__.Notify(576855); s.pos = p }
func (s fakeSym) Str() string               { __antithesis_instrumentation__.Notify(576856); return s.s }
func (s *fakeSym) SetStr(v string)          { __antithesis_instrumentation__.Notify(576857); s.s = v }
func (s fakeSym) UnionVal() interface{}     { __antithesis_instrumentation__.Notify(576858); return nil }
func (s fakeSym) SetUnionVal(v interface{}) { __antithesis_instrumentation__.Notify(576859) }
