package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/errors"
)

type intervalLexer struct {
	str    string
	offset int
	err    error
}

func (l *intervalLexer) consumeNum() (int64, bool, float64) {
	__antithesis_instrumentation__.Notify(610107)
	if l.err != nil {
		__antithesis_instrumentation__.Notify(610113)
		return 0, false, 0
	} else {
		__antithesis_instrumentation__.Notify(610114)
	}
	__antithesis_instrumentation__.Notify(610108)

	offset := l.offset

	neg := false
	if l.offset < len(l.str) && func() bool {
		__antithesis_instrumentation__.Notify(610115)
		return l.str[l.offset] == '-' == true
	}() == true {
		__antithesis_instrumentation__.Notify(610116)

		neg = true
	} else {
		__antithesis_instrumentation__.Notify(610117)
	}
	__antithesis_instrumentation__.Notify(610109)

	intPart := l.consumeInt()

	var decPart float64
	hasDecimal := false
	if l.offset < len(l.str) && func() bool {
		__antithesis_instrumentation__.Notify(610118)
		return l.str[l.offset] == '.' == true
	}() == true {
		__antithesis_instrumentation__.Notify(610119)
		hasDecimal = true
		start := l.offset

		l.offset++
		for ; l.offset < len(l.str) && func() bool {
			__antithesis_instrumentation__.Notify(610122)
			return l.str[l.offset] >= '0' == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(610123)
			return l.str[l.offset] <= '9' == true
		}() == true; l.offset++ {
			__antithesis_instrumentation__.Notify(610124)
		}
		__antithesis_instrumentation__.Notify(610120)

		value, err := strconv.ParseFloat(l.str[start:l.offset], 64)
		if err != nil {
			__antithesis_instrumentation__.Notify(610125)
			l.err = pgerror.Wrap(
				err, pgcode.InvalidDatetimeFormat, "interval")
			return 0, false, 0
		} else {
			__antithesis_instrumentation__.Notify(610126)
		}
		__antithesis_instrumentation__.Notify(610121)
		decPart = value
	} else {
		__antithesis_instrumentation__.Notify(610127)
	}
	__antithesis_instrumentation__.Notify(610110)

	if offset == l.offset {
		__antithesis_instrumentation__.Notify(610128)
		l.err = pgerror.Newf(
			pgcode.InvalidDatetimeFormat, "interval: missing number at position %d: %q", offset, l.str)
		return 0, false, 0
	} else {
		__antithesis_instrumentation__.Notify(610129)
	}
	__antithesis_instrumentation__.Notify(610111)

	if neg {
		__antithesis_instrumentation__.Notify(610130)
		decPart = -decPart
	} else {
		__antithesis_instrumentation__.Notify(610131)
	}
	__antithesis_instrumentation__.Notify(610112)
	return intPart, hasDecimal, decPart
}

func (l *intervalLexer) consumeInt() int64 {
	__antithesis_instrumentation__.Notify(610132)
	if l.err != nil {
		__antithesis_instrumentation__.Notify(610139)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(610140)
	}
	__antithesis_instrumentation__.Notify(610133)

	start := l.offset

	if l.offset < len(l.str) && func() bool {
		__antithesis_instrumentation__.Notify(610141)
		return (l.str[l.offset] == '-' || func() bool {
			__antithesis_instrumentation__.Notify(610142)
			return l.str[l.offset] == '+' == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(610143)
		l.offset++
	} else {
		__antithesis_instrumentation__.Notify(610144)
	}
	__antithesis_instrumentation__.Notify(610134)
	for ; l.offset < len(l.str) && func() bool {
		__antithesis_instrumentation__.Notify(610145)
		return l.str[l.offset] >= '0' == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(610146)
		return l.str[l.offset] <= '9' == true
	}() == true; l.offset++ {
		__antithesis_instrumentation__.Notify(610147)
	}
	__antithesis_instrumentation__.Notify(610135)

	if start == l.offset && func() bool {
		__antithesis_instrumentation__.Notify(610148)
		return len(l.str) > (l.offset + 1) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(610149)
		return l.str[l.offset] == '.' == true
	}() == true {
		__antithesis_instrumentation__.Notify(610150)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(610151)
	}
	__antithesis_instrumentation__.Notify(610136)

	x, err := strconv.ParseInt(l.str[start:l.offset], 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(610152)
		l.err = pgerror.Wrap(
			err, pgcode.InvalidDatetimeFormat, "interval")
		return 0
	} else {
		__antithesis_instrumentation__.Notify(610153)
	}
	__antithesis_instrumentation__.Notify(610137)
	if start == l.offset {
		__antithesis_instrumentation__.Notify(610154)
		l.err = pgerror.Newf(
			pgcode.InvalidDatetimeFormat, "interval: missing number at position %d: %q", start, l.str)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(610155)
	}
	__antithesis_instrumentation__.Notify(610138)
	return x
}

func (l *intervalLexer) consumeUnit(skipCharacter byte) string {
	__antithesis_instrumentation__.Notify(610156)
	if l.err != nil {
		__antithesis_instrumentation__.Notify(610160)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(610161)
	}
	__antithesis_instrumentation__.Notify(610157)

	offset := l.offset
	for ; l.offset < len(l.str); l.offset++ {
		__antithesis_instrumentation__.Notify(610162)
		if (l.str[l.offset] >= '0' && func() bool {
			__antithesis_instrumentation__.Notify(610163)
			return l.str[l.offset] <= '9' == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(610164)
			return l.str[l.offset] == skipCharacter == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(610165)
			return l.str[l.offset] == '-' == true
		}() == true {
			__antithesis_instrumentation__.Notify(610166)
			break
		} else {
			__antithesis_instrumentation__.Notify(610167)
		}
	}
	__antithesis_instrumentation__.Notify(610158)

	if offset == l.offset {
		__antithesis_instrumentation__.Notify(610168)
		l.err = pgerror.Newf(
			pgcode.InvalidDatetimeFormat, "interval: missing unit at position %d: %q", offset, l.str)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(610169)
	}
	__antithesis_instrumentation__.Notify(610159)
	return l.str[offset:l.offset]
}

func (l *intervalLexer) consumeSpaces() {
	__antithesis_instrumentation__.Notify(610170)
	if l.err != nil {
		__antithesis_instrumentation__.Notify(610172)
		return
	} else {
		__antithesis_instrumentation__.Notify(610173)
	}
	__antithesis_instrumentation__.Notify(610171)
	for ; l.offset < len(l.str) && func() bool {
		__antithesis_instrumentation__.Notify(610174)
		return l.str[l.offset] == ' ' == true
	}() == true; l.offset++ {
		__antithesis_instrumentation__.Notify(610175)
	}
}

var isoDateUnitMap = map[string]duration.Duration{
	"D": duration.MakeDuration(0, 1, 0),
	"W": duration.MakeDuration(0, 7, 0),
	"M": duration.MakeDuration(0, 0, 1),
	"Y": duration.MakeDuration(0, 0, 12),
}

var isoTimeUnitMap = map[string]duration.Duration{
	"S": duration.MakeDuration(time.Second.Nanoseconds(), 0, 0),
	"M": duration.MakeDuration(time.Minute.Nanoseconds(), 0, 0),
	"H": duration.MakeDuration(time.Hour.Nanoseconds(), 0, 0),
}

const errInvalidSQLDuration = "invalid input syntax for type interval %s"

type parsedIndex uint8

const (
	nothingParsed parsedIndex = iota
	hmsParsed
	dayParsed
	yearMonthParsed
)

func newInvalidSQLDurationError(s string) error {
	__antithesis_instrumentation__.Notify(610176)
	return pgerror.Newf(pgcode.InvalidDatetimeFormat, errInvalidSQLDuration, s)
}

func sqlStdToDuration(s string, itm types.IntervalTypeMetadata) (duration.Duration, error) {
	__antithesis_instrumentation__.Notify(610177)
	var d duration.Duration
	parts := strings.Fields(s)
	if len(parts) > 3 || func() bool {
		__antithesis_instrumentation__.Notify(610180)
		return len(parts) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(610181)
		return d, newInvalidSQLDurationError(s)
	} else {
		__antithesis_instrumentation__.Notify(610182)
	}
	__antithesis_instrumentation__.Notify(610178)

	parsedIdx := nothingParsed

	floatParsed := false

	for i := len(parts) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(610183)

		part := parts[i]

		consumeNeg := func(str string) (newStr string, mult int64, ok bool) {
			__antithesis_instrumentation__.Notify(610186)
			neg := false

			if str != "" {
				__antithesis_instrumentation__.Notify(610191)
				c := str[0]
				if c == '-' || func() bool {
					__antithesis_instrumentation__.Notify(610192)
					return c == '+' == true
				}() == true {
					__antithesis_instrumentation__.Notify(610193)
					neg = c == '-'
					str = str[1:]
				} else {
					__antithesis_instrumentation__.Notify(610194)
				}
			} else {
				__antithesis_instrumentation__.Notify(610195)
			}
			__antithesis_instrumentation__.Notify(610187)
			if len(str) == 0 {
				__antithesis_instrumentation__.Notify(610196)
				return str, 0, false
			} else {
				__antithesis_instrumentation__.Notify(610197)
			}
			__antithesis_instrumentation__.Notify(610188)
			if str[0] == '-' || func() bool {
				__antithesis_instrumentation__.Notify(610198)
				return str[0] == '+' == true
			}() == true {
				__antithesis_instrumentation__.Notify(610199)
				return str, 0, false
			} else {
				__antithesis_instrumentation__.Notify(610200)
			}
			__antithesis_instrumentation__.Notify(610189)

			mult = 1
			if neg {
				__antithesis_instrumentation__.Notify(610201)
				mult = -1
			} else {
				__antithesis_instrumentation__.Notify(610202)
			}
			__antithesis_instrumentation__.Notify(610190)
			return str, mult, true
		}
		__antithesis_instrumentation__.Notify(610184)

		var mult int64
		var ok bool
		if part, mult, ok = consumeNeg(part); !ok {
			__antithesis_instrumentation__.Notify(610203)
			return d, newInvalidSQLDurationError(s)
		} else {
			__antithesis_instrumentation__.Notify(610204)
		}
		__antithesis_instrumentation__.Notify(610185)

		if strings.ContainsRune(part, ':') {
			__antithesis_instrumentation__.Notify(610205)

			if parsedIdx != nothingParsed {
				__antithesis_instrumentation__.Notify(610211)
				return d, newInvalidSQLDurationError(s)
			} else {
				__antithesis_instrumentation__.Notify(610212)
			}
			__antithesis_instrumentation__.Notify(610206)
			parsedIdx = hmsParsed

			hms := strings.Split(part, ":")

			firstComponentIsFloat := strings.Contains(hms[0], ".")
			if firstComponentIsFloat || func() bool {
				__antithesis_instrumentation__.Notify(610213)
				return hms[0] == "" == true
			}() == true {
				__antithesis_instrumentation__.Notify(610214)

				if mult != 1 || func() bool {
					__antithesis_instrumentation__.Notify(610217)
					return len(hms) == 1 == true
				}() == true {
					__antithesis_instrumentation__.Notify(610218)
					return d, newInvalidSQLDurationError(s)
				} else {
					__antithesis_instrumentation__.Notify(610219)
				}
				__antithesis_instrumentation__.Notify(610215)
				if firstComponentIsFloat {
					__antithesis_instrumentation__.Notify(610220)
					days, err := strconv.ParseFloat(hms[0], 64)
					if err != nil {
						__antithesis_instrumentation__.Notify(610222)
						return d, newInvalidSQLDurationError(s)
					} else {
						__antithesis_instrumentation__.Notify(610223)
					}
					__antithesis_instrumentation__.Notify(610221)
					d = d.Add(duration.MakeDuration(0, 1, 0).MulFloat(days))
				} else {
					__antithesis_instrumentation__.Notify(610224)
				}
				__antithesis_instrumentation__.Notify(610216)

				hms = hms[1:]
				if hms[0], mult, ok = consumeNeg(hms[0]); !ok {
					__antithesis_instrumentation__.Notify(610225)
					return d, newInvalidSQLDurationError(s)
				} else {
					__antithesis_instrumentation__.Notify(610226)
				}
			} else {
				__antithesis_instrumentation__.Notify(610227)
			}
			__antithesis_instrumentation__.Notify(610207)

			for i := 0; i < len(hms); i++ {
				__antithesis_instrumentation__.Notify(610228)
				if hms[i] == "" {
					__antithesis_instrumentation__.Notify(610229)
					hms[i] = "0"
				} else {
					__antithesis_instrumentation__.Notify(610230)
				}
			}
			__antithesis_instrumentation__.Notify(610208)

			var hours, mins int64
			var secs float64

			switch len(hms) {
			case 2:
				__antithesis_instrumentation__.Notify(610231)

				var err error
				if strings.Contains(hms[1], ".") || func() bool {
					__antithesis_instrumentation__.Notify(610236)
					return itm.DurationField.IsMinuteToSecond() == true
				}() == true {
					__antithesis_instrumentation__.Notify(610237)
					if mins, err = strconv.ParseInt(hms[0], 10, 64); err != nil {
						__antithesis_instrumentation__.Notify(610239)
						return d, newInvalidSQLDurationError(s)
					} else {
						__antithesis_instrumentation__.Notify(610240)
					}
					__antithesis_instrumentation__.Notify(610238)
					if secs, err = strconv.ParseFloat(hms[1], 64); err != nil {
						__antithesis_instrumentation__.Notify(610241)
						return d, newInvalidSQLDurationError(s)
					} else {
						__antithesis_instrumentation__.Notify(610242)
					}
				} else {
					__antithesis_instrumentation__.Notify(610243)
					if hours, err = strconv.ParseInt(hms[0], 10, 64); err != nil {
						__antithesis_instrumentation__.Notify(610245)
						return d, newInvalidSQLDurationError(s)
					} else {
						__antithesis_instrumentation__.Notify(610246)
					}
					__antithesis_instrumentation__.Notify(610244)
					if mins, err = strconv.ParseInt(hms[1], 10, 64); err != nil {
						__antithesis_instrumentation__.Notify(610247)
						return d, newInvalidSQLDurationError(s)
					} else {
						__antithesis_instrumentation__.Notify(610248)
					}
				}
			case 3:
				__antithesis_instrumentation__.Notify(610232)
				var err error
				if hours, err = strconv.ParseInt(hms[0], 10, 64); err != nil {
					__antithesis_instrumentation__.Notify(610249)
					return d, newInvalidSQLDurationError(s)
				} else {
					__antithesis_instrumentation__.Notify(610250)
				}
				__antithesis_instrumentation__.Notify(610233)
				if mins, err = strconv.ParseInt(hms[1], 10, 64); err != nil {
					__antithesis_instrumentation__.Notify(610251)
					return d, newInvalidSQLDurationError(s)
				} else {
					__antithesis_instrumentation__.Notify(610252)
				}
				__antithesis_instrumentation__.Notify(610234)
				if secs, err = strconv.ParseFloat(hms[2], 64); err != nil {
					__antithesis_instrumentation__.Notify(610253)
					return d, newInvalidSQLDurationError(s)
				} else {
					__antithesis_instrumentation__.Notify(610254)
				}
			default:
				__antithesis_instrumentation__.Notify(610235)
				return d, newInvalidSQLDurationError(s)
			}
			__antithesis_instrumentation__.Notify(610209)

			if hours < 0 || func() bool {
				__antithesis_instrumentation__.Notify(610255)
				return mins < 0 == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(610256)
				return secs < 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(610257)
				return d, newInvalidSQLDurationError(s)
			} else {
				__antithesis_instrumentation__.Notify(610258)
			}
			__antithesis_instrumentation__.Notify(610210)

			d = d.Add(duration.MakeDuration(time.Hour.Nanoseconds(), 0, 0).Mul(mult * hours))
			d = d.Add(duration.MakeDuration(time.Minute.Nanoseconds(), 0, 0).Mul(mult * mins))
			d = d.Add(duration.MakeDuration(time.Second.Nanoseconds(), 0, 0).MulFloat(float64(mult) * secs))
		} else {
			__antithesis_instrumentation__.Notify(610259)
			if strings.ContainsRune(part, '-') {
				__antithesis_instrumentation__.Notify(610260)

				if parsedIdx >= yearMonthParsed {
					__antithesis_instrumentation__.Notify(610264)
					return d, newInvalidSQLDurationError(s)
				} else {
					__antithesis_instrumentation__.Notify(610265)
				}
				__antithesis_instrumentation__.Notify(610261)
				parsedIdx = yearMonthParsed

				yms := strings.Split(part, "-")
				if len(yms) != 2 {
					__antithesis_instrumentation__.Notify(610266)
					return d, newInvalidSQLDurationError(s)
				} else {
					__antithesis_instrumentation__.Notify(610267)
				}
				__antithesis_instrumentation__.Notify(610262)
				year, errYear := strconv.Atoi(yms[0])
				var month int
				var errMonth error
				if yms[1] != "" {
					__antithesis_instrumentation__.Notify(610268)

					month, errMonth = strconv.Atoi(yms[1])
				} else {
					__antithesis_instrumentation__.Notify(610269)
				}
				__antithesis_instrumentation__.Notify(610263)
				if errYear == nil && func() bool {
					__antithesis_instrumentation__.Notify(610270)
					return errMonth == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(610271)
					delta := duration.MakeDuration(0, 0, 1).Mul(int64(year)*12 + int64(month))
					if mult < 0 {
						__antithesis_instrumentation__.Notify(610272)
						d = d.Sub(delta)
					} else {
						__antithesis_instrumentation__.Notify(610273)
						d = d.Add(delta)
					}
				} else {
					__antithesis_instrumentation__.Notify(610274)
					return d, newInvalidSQLDurationError(s)
				}
			} else {
				__antithesis_instrumentation__.Notify(610275)
				if value, err := strconv.ParseFloat(part, 64); err == nil {
					__antithesis_instrumentation__.Notify(610276)

					if floatParsed && func() bool {
						__antithesis_instrumentation__.Notify(610278)
						return !itm.DurationField.IsDayToHour() == true
					}() == true {
						__antithesis_instrumentation__.Notify(610279)
						return d, newInvalidSQLDurationError(s)
					} else {
						__antithesis_instrumentation__.Notify(610280)
					}
					__antithesis_instrumentation__.Notify(610277)
					floatParsed = true
					if parsedIdx == nothingParsed {
						__antithesis_instrumentation__.Notify(610281)

						switch itm.DurationField.DurationType {
						case types.IntervalDurationType_YEAR:
							__antithesis_instrumentation__.Notify(610283)
							d = d.Add(duration.MakeDuration(0, 0, 12).MulFloat(value * float64(mult)))
						case types.IntervalDurationType_MONTH:
							__antithesis_instrumentation__.Notify(610284)
							d = d.Add(duration.MakeDuration(0, 0, 1).MulFloat(value * float64(mult)))
						case types.IntervalDurationType_DAY:
							__antithesis_instrumentation__.Notify(610285)
							d = d.Add(duration.MakeDuration(0, 1, 0).MulFloat(value * float64(mult)))
						case types.IntervalDurationType_HOUR:
							__antithesis_instrumentation__.Notify(610286)
							d = d.Add(duration.MakeDuration(time.Hour.Nanoseconds(), 0, 0).MulFloat(value * float64(mult)))
						case types.IntervalDurationType_MINUTE:
							__antithesis_instrumentation__.Notify(610287)
							d = d.Add(duration.MakeDuration(time.Minute.Nanoseconds(), 0, 0).MulFloat(value * float64(mult)))
						case types.IntervalDurationType_SECOND, types.IntervalDurationType_UNSET:
							__antithesis_instrumentation__.Notify(610288)
							d = d.Add(duration.MakeDuration(time.Second.Nanoseconds(), 0, 0).MulFloat(value * float64(mult)))
						case types.IntervalDurationType_MILLISECOND:
							__antithesis_instrumentation__.Notify(610289)
							d = d.Add(duration.MakeDuration(time.Millisecond.Nanoseconds(), 0, 0).MulFloat(value * float64(mult)))
						default:
							__antithesis_instrumentation__.Notify(610290)
							return d, errors.AssertionFailedf("unhandled DurationField constant %#v", itm.DurationField)
						}
						__antithesis_instrumentation__.Notify(610282)
						parsedIdx = hmsParsed
					} else {
						__antithesis_instrumentation__.Notify(610291)
						if parsedIdx == hmsParsed {
							__antithesis_instrumentation__.Notify(610292)

							delta := duration.MakeDuration(0, 1, 0).MulFloat(value)
							if mult < 0 {
								__antithesis_instrumentation__.Notify(610294)
								d = d.Sub(delta)
							} else {
								__antithesis_instrumentation__.Notify(610295)
								d = d.Add(delta)
							}
							__antithesis_instrumentation__.Notify(610293)
							parsedIdx = dayParsed
						} else {
							__antithesis_instrumentation__.Notify(610296)
							return d, newInvalidSQLDurationError(s)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(610297)
					return d, newInvalidSQLDurationError(s)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(610179)
	return d, nil
}

func iso8601ToDuration(s string) (duration.Duration, error) {
	__antithesis_instrumentation__.Notify(610298)
	var d duration.Duration
	if len(s) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(610301)
		return s[0] != 'P' == true
	}() == true {
		__antithesis_instrumentation__.Notify(610302)
		return d, newInvalidSQLDurationError(s)
	} else {
		__antithesis_instrumentation__.Notify(610303)
	}
	__antithesis_instrumentation__.Notify(610299)

	l := intervalLexer{str: s, offset: 1, err: nil}
	unitMap := isoDateUnitMap

	for l.offset < len(s) {
		__antithesis_instrumentation__.Notify(610304)

		if s[l.offset] == 'T' {
			__antithesis_instrumentation__.Notify(610307)
			unitMap = isoTimeUnitMap
			l.offset++
		} else {
			__antithesis_instrumentation__.Notify(610308)
		}
		__antithesis_instrumentation__.Notify(610305)

		v, hasDecimal, vp := l.consumeNum()
		u := l.consumeUnit('T')
		if l.err != nil {
			__antithesis_instrumentation__.Notify(610309)
			return d, l.err
		} else {
			__antithesis_instrumentation__.Notify(610310)
		}
		__antithesis_instrumentation__.Notify(610306)

		if unit, ok := unitMap[u]; ok {
			__antithesis_instrumentation__.Notify(610311)
			d = d.Add(unit.Mul(v))
			if hasDecimal {
				__antithesis_instrumentation__.Notify(610312)
				var err error
				d, err = addFrac(d, unit, vp)
				if err != nil {
					__antithesis_instrumentation__.Notify(610313)
					return d, err
				} else {
					__antithesis_instrumentation__.Notify(610314)
				}
			} else {
				__antithesis_instrumentation__.Notify(610315)
			}
		} else {
			__antithesis_instrumentation__.Notify(610316)
			return d, pgerror.Newf(
				pgcode.InvalidDatetimeFormat,
				"interval: unknown unit %s in ISO-8601 duration %s", u, s)
		}
	}
	__antithesis_instrumentation__.Notify(610300)

	return d, nil
}

var unitMap = func(
	units map[string]duration.Duration,
	aliases map[string][]string,
) map[string]duration.Duration {
	__antithesis_instrumentation__.Notify(610317)
	for a, alist := range aliases {
		__antithesis_instrumentation__.Notify(610319)

		units[a+"s"] = units[a]
		for _, alias := range alist {
			__antithesis_instrumentation__.Notify(610320)

			units[alias] = units[a]
		}
	}
	__antithesis_instrumentation__.Notify(610318)
	return units
}(map[string]duration.Duration{

	"microsecond": duration.MakeDuration(time.Microsecond.Nanoseconds(), 0, 0),
	"millisecond": duration.MakeDuration(time.Millisecond.Nanoseconds(), 0, 0),
	"second":      duration.MakeDuration(time.Second.Nanoseconds(), 0, 0),
	"minute":      duration.MakeDuration(time.Minute.Nanoseconds(), 0, 0),
	"hour":        duration.MakeDuration(time.Hour.Nanoseconds(), 0, 0),
	"day":         duration.MakeDuration(0, 1, 0),
	"week":        duration.MakeDuration(0, 7, 0),
	"month":       duration.MakeDuration(0, 0, 1),
	"year":        duration.MakeDuration(0, 0, 12),
}, map[string][]string{

	"microsecond": {"us", "µs", "μs", "usec", "usecs", "usecond", "useconds"},
	"millisecond": {"ms", "msec", "msecs", "msecond", "mseconds"},
	"second":      {"s", "sec", "secs"},
	"minute":      {"m", "min", "mins"},
	"hour":        {"h", "hr", "hrs"},
	"day":         {"d"},
	"week":        {"w"},
	"month":       {"mon", "mons"},
	"year":        {"y", "yr", "yrs"},
})

func parseDuration(
	style duration.IntervalStyle, s string, itm types.IntervalTypeMetadata,
) (duration.Duration, error) {
	__antithesis_instrumentation__.Notify(610321)
	var d duration.Duration
	l := intervalLexer{str: s, offset: 0, err: nil}
	l.consumeSpaces()

	if l.offset == len(l.str) {
		__antithesis_instrumentation__.Notify(610326)
		return d, pgerror.Newf(
			pgcode.InvalidDatetimeFormat, "interval: invalid input syntax: %q", l.str)
	} else {
		__antithesis_instrumentation__.Notify(610327)
	}
	__antithesis_instrumentation__.Notify(610322)

	isSQLStandardNegative :=
		style == duration.IntervalStyle_SQL_STANDARD && func() bool {
			__antithesis_instrumentation__.Notify(610328)
			return (l.offset + 1) < len(l.str) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(610329)
			return l.str[l.offset] == '-' == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(610330)
			return !strings.ContainsAny(l.str[l.offset+1:], "+-") == true
		}() == true
	if isSQLStandardNegative {
		__antithesis_instrumentation__.Notify(610331)
		l.offset++
	} else {
		__antithesis_instrumentation__.Notify(610332)
	}
	__antithesis_instrumentation__.Notify(610323)

	for l.offset != len(l.str) {
		__antithesis_instrumentation__.Notify(610333)

		sign := l.str[l.offset] == '-'

		v, hasDecimal, vp := l.consumeNum()
		l.consumeSpaces()

		if l.offset < len(l.str) && func() bool {
			__antithesis_instrumentation__.Notify(610338)
			return l.str[l.offset] == ':' == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(610339)
			return !hasDecimal == true
		}() == true {
			__antithesis_instrumentation__.Notify(610340)

			delta, err := l.parseShortDuration(v, sign, itm)
			if err != nil {
				__antithesis_instrumentation__.Notify(610342)
				return d, err
			} else {
				__antithesis_instrumentation__.Notify(610343)
			}
			__antithesis_instrumentation__.Notify(610341)
			d = d.Add(delta)
			continue
		} else {
			__antithesis_instrumentation__.Notify(610344)
		}
		__antithesis_instrumentation__.Notify(610334)

		u := l.consumeUnit(' ')
		l.consumeSpaces()
		if unit, ok := unitMap[strings.ToLower(u)]; ok {
			__antithesis_instrumentation__.Notify(610345)

			d = d.Add(unit.Mul(v))
			if hasDecimal {
				__antithesis_instrumentation__.Notify(610347)
				var err error
				d, err = addFrac(d, unit, vp)
				if err != nil {
					__antithesis_instrumentation__.Notify(610348)
					return d, err
				} else {
					__antithesis_instrumentation__.Notify(610349)
				}
			} else {
				__antithesis_instrumentation__.Notify(610350)
			}
			__antithesis_instrumentation__.Notify(610346)
			continue
		} else {
			__antithesis_instrumentation__.Notify(610351)
		}
		__antithesis_instrumentation__.Notify(610335)

		if l.err != nil {
			__antithesis_instrumentation__.Notify(610352)
			return d, l.err
		} else {
			__antithesis_instrumentation__.Notify(610353)
		}
		__antithesis_instrumentation__.Notify(610336)
		if u != "" {
			__antithesis_instrumentation__.Notify(610354)
			return d, pgerror.Newf(
				pgcode.InvalidDatetimeFormat, "interval: unknown unit %q in duration %q", u, s)
		} else {
			__antithesis_instrumentation__.Notify(610355)
		}
		__antithesis_instrumentation__.Notify(610337)
		return d, pgerror.Newf(
			pgcode.InvalidDatetimeFormat, "interval: missing unit at position %d: %q", l.offset, s)
	}
	__antithesis_instrumentation__.Notify(610324)
	if isSQLStandardNegative {
		__antithesis_instrumentation__.Notify(610356)
		return duration.MakeDuration(
			-d.Nanos(),
			-d.Days,
			-d.Months,
		), l.err
	} else {
		__antithesis_instrumentation__.Notify(610357)
	}
	__antithesis_instrumentation__.Notify(610325)
	return d, l.err
}

func (l *intervalLexer) parseShortDuration(
	h int64, hasSign bool, itm types.IntervalTypeMetadata,
) (duration.Duration, error) {
	__antithesis_instrumentation__.Notify(610358)
	sign := int64(1)
	if hasSign {
		__antithesis_instrumentation__.Notify(610365)
		sign = -1
	} else {
		__antithesis_instrumentation__.Notify(610366)
	}
	__antithesis_instrumentation__.Notify(610359)

	if l.str[l.offset] != ':' {
		__antithesis_instrumentation__.Notify(610367)
		return duration.Duration{}, pgerror.Newf(
			pgcode.InvalidDatetimeFormat, "interval: invalid format %s", l.str[l.offset:])
	} else {
		__antithesis_instrumentation__.Notify(610368)
	}
	__antithesis_instrumentation__.Notify(610360)
	l.offset++

	m, hasDecimal, mp := l.consumeNum()

	if m < 0 {
		__antithesis_instrumentation__.Notify(610369)
		return duration.Duration{}, pgerror.Newf(
			pgcode.InvalidDatetimeFormat, "interval: invalid format: %s", l.str)
	} else {
		__antithesis_instrumentation__.Notify(610370)
	}
	__antithesis_instrumentation__.Notify(610361)

	if hasDecimal {
		__antithesis_instrumentation__.Notify(610371)
		l.consumeSpaces()
		return duration.MakeDuration(
			h*time.Minute.Nanoseconds()+
				sign*(m*time.Second.Nanoseconds()+
					floatToNanos(mp)),
			0,
			0,
		), nil
	} else {
		__antithesis_instrumentation__.Notify(610372)
	}
	__antithesis_instrumentation__.Notify(610362)

	var s int64
	var sp float64
	hasSecondsComponent := false
	if l.offset != len(l.str) && func() bool {
		__antithesis_instrumentation__.Notify(610373)
		return l.str[l.offset] == ':' == true
	}() == true {
		__antithesis_instrumentation__.Notify(610374)
		hasSecondsComponent = true

		l.offset++
		s, _, sp = l.consumeNum()
		if s < 0 {
			__antithesis_instrumentation__.Notify(610375)
			return duration.Duration{}, pgerror.Newf(
				pgcode.InvalidDatetimeFormat, "interval: invalid format: %s", l.str)
		} else {
			__antithesis_instrumentation__.Notify(610376)
		}
	} else {
		__antithesis_instrumentation__.Notify(610377)
	}
	__antithesis_instrumentation__.Notify(610363)

	l.consumeSpaces()

	if !hasSecondsComponent && func() bool {
		__antithesis_instrumentation__.Notify(610378)
		return itm.DurationField.IsMinuteToSecond() == true
	}() == true {
		__antithesis_instrumentation__.Notify(610379)
		return duration.MakeDuration(
			h*time.Minute.Nanoseconds()+sign*(m*time.Second.Nanoseconds()),
			0,
			0,
		), nil
	} else {
		__antithesis_instrumentation__.Notify(610380)
	}
	__antithesis_instrumentation__.Notify(610364)
	return duration.MakeDuration(
		h*time.Hour.Nanoseconds()+
			sign*(m*time.Minute.Nanoseconds()+
				int64(mp*float64(time.Minute.Nanoseconds()))+
				s*time.Second.Nanoseconds()+
				floatToNanos(sp)),
		0,
		0,
	), nil
}

func addFrac(d duration.Duration, unit duration.Duration, f float64) (duration.Duration, error) {
	__antithesis_instrumentation__.Notify(610381)
	if unit.Months > 0 {
		__antithesis_instrumentation__.Notify(610383)
		f = f * float64(unit.Months)
		d.Months += int64(f)
		switch unit.Months {
		case 1:
			__antithesis_instrumentation__.Notify(610384)
			f = math.Mod(f, 1) * 30
			d.Days += int64(f)
			f = math.Mod(f, 1) * 24
			d.SetNanos(d.Nanos() + int64(float64(time.Hour.Nanoseconds())*f))
		case 12:
			__antithesis_instrumentation__.Notify(610385)

		default:
			__antithesis_instrumentation__.Notify(610386)
			return duration.Duration{}, errors.AssertionFailedf("unhandled unit type %v", unit)
		}
	} else {
		__antithesis_instrumentation__.Notify(610387)
		if unit.Days > 0 {
			__antithesis_instrumentation__.Notify(610388)
			f = f * float64(unit.Days)
			d.Days += int64(f)
			f = math.Mod(f, 1) * 24
			d.SetNanos(d.Nanos() + int64(float64(time.Hour.Nanoseconds())*f))
		} else {
			__antithesis_instrumentation__.Notify(610389)
			d.SetNanos(d.Nanos() + int64(float64(unit.Nanos())*f))
		}
	}
	__antithesis_instrumentation__.Notify(610382)
	return d, nil
}

func floatToNanos(f float64) int64 {
	__antithesis_instrumentation__.Notify(610390)
	return int64(math.Round(f * float64(time.Second.Nanoseconds())))
}
