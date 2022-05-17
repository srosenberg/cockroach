package issues

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"html"
	"strings"
	"unicode/utf8"
)

type IssueFormatter struct {
	Title func(TemplateData) string
	Body  func(*Renderer, TemplateData) error
}

type Renderer struct {
	buf strings.Builder
}

func (r *Renderer) printf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(41225)
	fmt.Fprintf(&r.buf, format, args...)
}

func (r *Renderer) esc(in string, chars string, with rune) string {
	__antithesis_instrumentation__.Notify(41226)
	for {
		__antithesis_instrumentation__.Notify(41227)
		r, n := utf8.DecodeRuneInString(chars)
		if r == utf8.RuneError {
			__antithesis_instrumentation__.Notify(41229)
			return in
		} else {
			__antithesis_instrumentation__.Notify(41230)
		}
		__antithesis_instrumentation__.Notify(41228)
		chars = chars[n:]
		s := string(r)
		in = strings.Replace(in, s, string(with)+s, -1)
	}
}

func (r *Renderer) nl() {
	__antithesis_instrumentation__.Notify(41231)
	if n := r.buf.Len(); n > 0 && func() bool {
		__antithesis_instrumentation__.Notify(41233)
		return r.buf.String()[n-1] == '\n' == true
	}() == true {
		__antithesis_instrumentation__.Notify(41234)
		return
	} else {
		__antithesis_instrumentation__.Notify(41235)
	}
	__antithesis_instrumentation__.Notify(41232)
	r.buf.WriteByte('\n')
}

func (r *Renderer) A(title, href string) {
	__antithesis_instrumentation__.Notify(41236)
	r.printf("[")
	r.Escaped(r.esc(title, "[]()", '\\'))
	r.printf("]")
	r.printf("(")
	r.printf("%s", r.esc(href, "[]()", '\\'))
	r.printf(")")
}

func (r *Renderer) P(inner func()) {
	__antithesis_instrumentation__.Notify(41237)
	r.HTML("p", inner)
}

func (r *Renderer) Escaped(txt string) {
	__antithesis_instrumentation__.Notify(41238)
	r.printf("%s", html.EscapeString(txt))
}

func (r *Renderer) CodeBlock(typ string, txt string) {
	__antithesis_instrumentation__.Notify(41239)
	r.nl()

	r.printf("\n```%s\n", r.esc(typ, "`", '`'))
	r.printf("%s", r.esc(txt, "`", '`'))
	r.nl()
	r.printf("%s", "```")
	r.nl()
}

func (r *Renderer) HTML(tag string, inner func()) {
	__antithesis_instrumentation__.Notify(41240)
	r.printf("<%s>", tag)
	inner()
	r.printf("</%s>", tag)
	r.nl()
}

func (r *Renderer) Collapsed(title string, inner func()) {
	__antithesis_instrumentation__.Notify(41241)
	r.HTML("details", func() {
		__antithesis_instrumentation__.Notify(41243)
		r.HTML("summary", func() {
			__antithesis_instrumentation__.Notify(41245)
			r.Escaped(title)
		})
		__antithesis_instrumentation__.Notify(41244)
		r.nl()
		r.P(func() {
			__antithesis_instrumentation__.Notify(41246)
			r.nl()
			inner()
		})
	})
	__antithesis_instrumentation__.Notify(41242)
	r.nl()
}

func (r *Renderer) String() string {
	__antithesis_instrumentation__.Notify(41247)
	return r.buf.String()
}
