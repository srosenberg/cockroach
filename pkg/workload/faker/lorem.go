package faker

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"
	"unicode"

	"golang.org/x/exp/rand"
)

type loremFaker struct {
	words *weightedEntries
}

func (f *loremFaker) Words(rng *rand.Rand, num int) []string {
	__antithesis_instrumentation__.Notify(694097)
	w := make([]string, num)
	for i := range w {
		__antithesis_instrumentation__.Notify(694099)
		w[i] = f.words.Rand(rng).(string)
	}
	__antithesis_instrumentation__.Notify(694098)
	return w
}

func (f *loremFaker) Sentences(rng *rand.Rand, num int) []string {
	__antithesis_instrumentation__.Notify(694100)
	s := make([]string, num)
	for i := range s {
		__antithesis_instrumentation__.Notify(694102)
		var b strings.Builder
		numWords := randInt(rng, 4, 8)
		for j := 0; j < numWords; j++ {
			__antithesis_instrumentation__.Notify(694104)
			word := f.words.Rand(rng).(string)
			if j == 0 {
				__antithesis_instrumentation__.Notify(694106)
				word = firstToUpper(word)
			} else {
				__antithesis_instrumentation__.Notify(694107)
			}
			__antithesis_instrumentation__.Notify(694105)
			b.WriteString(word)
			if j == numWords-1 {
				__antithesis_instrumentation__.Notify(694108)
				b.WriteString(`.`)
			} else {
				__antithesis_instrumentation__.Notify(694109)
				b.WriteString(` `)
			}
		}
		__antithesis_instrumentation__.Notify(694103)
		s[i] = b.String()
	}
	__antithesis_instrumentation__.Notify(694101)
	return s
}

func firstToUpper(s string) string {
	__antithesis_instrumentation__.Notify(694110)
	isFirst := true
	return strings.Map(func(r rune) rune {
		__antithesis_instrumentation__.Notify(694111)
		if isFirst {
			__antithesis_instrumentation__.Notify(694113)
			isFirst = false
			return unicode.ToUpper(r)
		} else {
			__antithesis_instrumentation__.Notify(694114)
		}
		__antithesis_instrumentation__.Notify(694112)
		return r
	}, s)
}

func (f *loremFaker) Paragraph(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694115)
	return strings.Join(f.Sentences(rng, randInt(rng, 1, 5)), ` `)
}

func newLoremFaker() loremFaker {
	__antithesis_instrumentation__.Notify(694116)
	f := loremFaker{}
	f.words = words()
	return f
}
