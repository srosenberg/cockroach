package yacc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type item struct {
	typ itemType
	pos Pos
	val string
}

func (i item) String() string {
	__antithesis_instrumentation__.Notify(68584)
	switch {
	case i.typ == itemEOF:
		__antithesis_instrumentation__.Notify(68586)
		return "EOF"
	case i.typ == itemError:
		__antithesis_instrumentation__.Notify(68587)
		return i.val
	case len(i.val) > 10:
		__antithesis_instrumentation__.Notify(68588)
		return fmt.Sprintf("%.10q...", i.val)
	default:
		__antithesis_instrumentation__.Notify(68589)
	}
	__antithesis_instrumentation__.Notify(68585)
	return fmt.Sprintf("%q", i.val)
}

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemComment
	itemPct
	itemDoublePct
	itemIdent
	itemColon
	itemLiteral
	itemExpr
	itemPipe
	itemNL
)

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	name    string
	input   string
	state   stateFn
	pos     Pos
	start   Pos
	width   Pos
	lastPos Pos
	items   chan item
}

func (l *lexer) next() rune {
	__antithesis_instrumentation__.Notify(68590)
	if int(l.pos) >= len(l.input) {
		__antithesis_instrumentation__.Notify(68592)
		l.width = 0
		return eof
	} else {
		__antithesis_instrumentation__.Notify(68593)
	}
	__antithesis_instrumentation__.Notify(68591)
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

func (l *lexer) peek() rune {
	__antithesis_instrumentation__.Notify(68594)
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	__antithesis_instrumentation__.Notify(68595)
	l.pos -= l.width
}

func (l *lexer) emit(t itemType) {
	__antithesis_instrumentation__.Notify(68596)
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) ignore() {
	__antithesis_instrumentation__.Notify(68597)
	l.start = l.pos
}

func (l *lexer) lineNumber() int {
	__antithesis_instrumentation__.Notify(68598)
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	__antithesis_instrumentation__.Notify(68599)
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

func (l *lexer) nextItem() item {
	__antithesis_instrumentation__.Notify(68600)
	i := <-l.items
	l.lastPos = i.pos
	return i
}

func lex(name, input string) *lexer {
	__antithesis_instrumentation__.Notify(68601)
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	__antithesis_instrumentation__.Notify(68602)
	for l.state = lexStart; l.state != nil; {
		__antithesis_instrumentation__.Notify(68603)
		l.state = l.state(l)
	}
}

func lexStart(l *lexer) stateFn {
	__antithesis_instrumentation__.Notify(68604)
Loop:
	for {
		__antithesis_instrumentation__.Notify(68606)
		switch r := l.next(); {
		case r == '/':
			__antithesis_instrumentation__.Notify(68607)
			return lexComment
		case r == '%':
			__antithesis_instrumentation__.Notify(68608)
			return lexPct
		case r == '\n':
			__antithesis_instrumentation__.Notify(68609)
			l.emit(itemNL)
		case r == ':':
			__antithesis_instrumentation__.Notify(68610)
			l.emit(itemColon)
		case r == '|':
			__antithesis_instrumentation__.Notify(68611)
			l.emit(itemPipe)
		case r == '{':
			__antithesis_instrumentation__.Notify(68612)
			return lexExpr
		case isSpace(r):
			__antithesis_instrumentation__.Notify(68613)
			l.ignore()
		case isIdent(r):
			__antithesis_instrumentation__.Notify(68614)
			return lexIdent
		case r == '\'':
			__antithesis_instrumentation__.Notify(68615)
			return lexLiteral
		case r == eof:
			__antithesis_instrumentation__.Notify(68616)
			l.emit(itemEOF)
			break Loop
		default:
			__antithesis_instrumentation__.Notify(68617)
			return l.errorf("invalid character: %v", string(r))
		}
	}
	__antithesis_instrumentation__.Notify(68605)
	return nil
}

func lexLiteral(l *lexer) stateFn {
	__antithesis_instrumentation__.Notify(68618)
	for {
		__antithesis_instrumentation__.Notify(68619)
		switch l.next() {
		case '\'':
			__antithesis_instrumentation__.Notify(68620)
			l.emit(itemLiteral)
			return lexStart
		default:
			__antithesis_instrumentation__.Notify(68621)
		}
	}
}

func lexExpr(l *lexer) stateFn {
	__antithesis_instrumentation__.Notify(68622)
	ct := 1
	for {
		__antithesis_instrumentation__.Notify(68623)
		switch l.next() {
		case '{':
			__antithesis_instrumentation__.Notify(68624)
			ct++
		case '}':
			__antithesis_instrumentation__.Notify(68625)
			ct--
			if ct == 0 {
				__antithesis_instrumentation__.Notify(68627)
				l.emit(itemExpr)
				return lexStart
			} else {
				__antithesis_instrumentation__.Notify(68628)
			}
		default:
			__antithesis_instrumentation__.Notify(68626)
		}
	}
}

func lexComment(l *lexer) stateFn {
	__antithesis_instrumentation__.Notify(68629)
	switch r := l.next(); r {
	case '/':
		__antithesis_instrumentation__.Notify(68630)
		for {
			__antithesis_instrumentation__.Notify(68633)
			switch l.next() {
			case '\n':
				__antithesis_instrumentation__.Notify(68634)
				l.backup()
				l.emit(itemComment)
				return lexStart
			default:
				__antithesis_instrumentation__.Notify(68635)
			}
		}
	case '*':
		__antithesis_instrumentation__.Notify(68631)
		for {
			__antithesis_instrumentation__.Notify(68636)
			switch l.next() {
			case '*':
				__antithesis_instrumentation__.Notify(68637)
				if l.peek() == '/' {
					__antithesis_instrumentation__.Notify(68639)
					l.next()
					l.emit(itemComment)
					return lexStart
				} else {
					__antithesis_instrumentation__.Notify(68640)
				}
			default:
				__antithesis_instrumentation__.Notify(68638)
			}
		}
	default:
		__antithesis_instrumentation__.Notify(68632)
		return l.errorf("expected comment: %c", r)
	}
}

func lexPct(l *lexer) stateFn {
	__antithesis_instrumentation__.Notify(68641)
	switch l.next() {
	case '%':
		__antithesis_instrumentation__.Notify(68642)
		l.emit(itemDoublePct)
		return lexStart
	case '{':
		__antithesis_instrumentation__.Notify(68643)
		for {
			__antithesis_instrumentation__.Notify(68647)
			switch l.next() {
			case '%':
				__antithesis_instrumentation__.Notify(68648)
				if l.peek() == '}' {
					__antithesis_instrumentation__.Notify(68650)
					l.next()
					l.emit(itemPct)
					return lexStart
				} else {
					__antithesis_instrumentation__.Notify(68651)
				}
			default:
				__antithesis_instrumentation__.Notify(68649)
			}
		}
	case 'p':
		__antithesis_instrumentation__.Notify(68644)
		if l.next() != 'r' || func() bool {
			__antithesis_instrumentation__.Notify(68652)
			return l.next() != 'e' == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(68653)
			return l.next() != 'c' == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(68654)
			return l.next() != ' ' == true
		}() == true {
			__antithesis_instrumentation__.Notify(68655)
			l.errorf("expected %%prec")
		} else {
			__antithesis_instrumentation__.Notify(68656)
		}
		__antithesis_instrumentation__.Notify(68645)
		for {
			__antithesis_instrumentation__.Notify(68657)
			switch r := l.next(); {
			case isIdent(r):
				__antithesis_instrumentation__.Notify(68658)

			default:
				__antithesis_instrumentation__.Notify(68659)
				l.backup()
				l.emit(itemPct)
				return lexStart
			}
		}
	default:
		__antithesis_instrumentation__.Notify(68646)
		ct := 0
		for {
			__antithesis_instrumentation__.Notify(68660)
			switch l.next() {
			case ' ':
				__antithesis_instrumentation__.Notify(68661)
			case '{':
				__antithesis_instrumentation__.Notify(68662)
				ct++
			case '}':
				__antithesis_instrumentation__.Notify(68663)
				ct--
				if ct == 0 {
					__antithesis_instrumentation__.Notify(68666)
					l.emit(itemPct)
					return lexStart
				} else {
					__antithesis_instrumentation__.Notify(68667)
				}
			case '\n':
				__antithesis_instrumentation__.Notify(68664)
				if ct == 0 {
					__antithesis_instrumentation__.Notify(68668)
					l.backup()
					l.emit(itemPct)
					return lexStart
				} else {
					__antithesis_instrumentation__.Notify(68669)
				}
			default:
				__antithesis_instrumentation__.Notify(68665)
			}
		}
	}
}

func lexIdent(l *lexer) stateFn {
	__antithesis_instrumentation__.Notify(68670)
	for {
		__antithesis_instrumentation__.Notify(68671)
		switch r := l.next(); {
		case isIdent(r):
			__antithesis_instrumentation__.Notify(68672)

		default:
			__antithesis_instrumentation__.Notify(68673)
			l.backup()
			l.emit(itemIdent)
			return lexStart
		}
	}
}

func isSpace(r rune) bool {
	__antithesis_instrumentation__.Notify(68674)
	return r == ' ' || func() bool {
		__antithesis_instrumentation__.Notify(68675)
		return r == '\t' == true
	}() == true
}

func isIdent(r rune) bool {
	__antithesis_instrumentation__.Notify(68676)
	return r == '_' || func() bool {
		__antithesis_instrumentation__.Notify(68677)
		return unicode.IsLetter(r) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(68678)
		return unicode.IsDigit(r) == true
	}() == true
}
