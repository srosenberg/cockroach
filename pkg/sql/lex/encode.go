package lex

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	"golang.org/x/text/language"
)

func NormalizeLocaleName(s string) string {
	__antithesis_instrumentation__.Notify(499444)
	b := bytes.NewBuffer(make([]byte, 0, len(s)))
	EncodeLocaleName(b, s)
	return b.String()
}

func EncodeLocaleName(buf *bytes.Buffer, s string) {
	__antithesis_instrumentation__.Notify(499445)

	if normalized, err := language.Parse(s); err == nil {
		__antithesis_instrumentation__.Notify(499447)
		s = normalized.String()
	} else {
		__antithesis_instrumentation__.Notify(499448)
	}
	__antithesis_instrumentation__.Notify(499446)
	for i, n := 0, len(s); i < n; i++ {
		__antithesis_instrumentation__.Notify(499449)
		ch := s[i]
		if ch == '-' {
			__antithesis_instrumentation__.Notify(499450)
			buf.WriteByte('_')
		} else {
			__antithesis_instrumentation__.Notify(499451)
			buf.WriteByte(ch)
		}
	}
}

func LocaleNamesAreEqual(a, b string) bool {
	__antithesis_instrumentation__.Notify(499452)
	if a == b {
		__antithesis_instrumentation__.Notify(499456)
		return true
	} else {
		__antithesis_instrumentation__.Notify(499457)
	}
	__antithesis_instrumentation__.Notify(499453)
	if len(a) != len(b) {
		__antithesis_instrumentation__.Notify(499458)
		return false
	} else {
		__antithesis_instrumentation__.Notify(499459)
	}
	__antithesis_instrumentation__.Notify(499454)
	for i, n := 0, len(a); i < n; i++ {
		__antithesis_instrumentation__.Notify(499460)
		ai, bi := a[i], b[i]
		if ai == bi {
			__antithesis_instrumentation__.Notify(499464)
			continue
		} else {
			__antithesis_instrumentation__.Notify(499465)
		}
		__antithesis_instrumentation__.Notify(499461)
		if ai == '-' && func() bool {
			__antithesis_instrumentation__.Notify(499466)
			return bi == '_' == true
		}() == true {
			__antithesis_instrumentation__.Notify(499467)
			continue
		} else {
			__antithesis_instrumentation__.Notify(499468)
		}
		__antithesis_instrumentation__.Notify(499462)
		if ai == '_' && func() bool {
			__antithesis_instrumentation__.Notify(499469)
			return bi == '-' == true
		}() == true {
			__antithesis_instrumentation__.Notify(499470)
			continue
		} else {
			__antithesis_instrumentation__.Notify(499471)
		}
		__antithesis_instrumentation__.Notify(499463)
		if unicode.ToLower(rune(ai)) != unicode.ToLower(rune(bi)) {
			__antithesis_instrumentation__.Notify(499472)
			return false
		} else {
			__antithesis_instrumentation__.Notify(499473)
		}
	}
	__antithesis_instrumentation__.Notify(499455)
	return true
}

func EncodeByteArrayToRawBytes(data string, be BytesEncodeFormat, skipHexPrefix bool) string {
	__antithesis_instrumentation__.Notify(499474)
	switch be {
	case BytesEncodeHex:
		__antithesis_instrumentation__.Notify(499475)
		head := 2
		if skipHexPrefix {
			__antithesis_instrumentation__.Notify(499482)
			head = 0
		} else {
			__antithesis_instrumentation__.Notify(499483)
		}
		__antithesis_instrumentation__.Notify(499476)
		res := make([]byte, head+hex.EncodedLen(len(data)))
		if !skipHexPrefix {
			__antithesis_instrumentation__.Notify(499484)
			res[0] = '\\'
			res[1] = 'x'
		} else {
			__antithesis_instrumentation__.Notify(499485)
		}
		__antithesis_instrumentation__.Notify(499477)
		hex.Encode(res[head:], []byte(data))
		return string(res)

	case BytesEncodeEscape:
		__antithesis_instrumentation__.Notify(499478)

		res := make([]byte, 0, len(data))
		for _, c := range []byte(data) {
			__antithesis_instrumentation__.Notify(499486)
			if c == '\\' {
				__antithesis_instrumentation__.Notify(499487)
				res = append(res, '\\', '\\')
			} else {
				__antithesis_instrumentation__.Notify(499488)
				if c < 32 || func() bool {
					__antithesis_instrumentation__.Notify(499489)
					return c >= 127 == true
				}() == true {
					__antithesis_instrumentation__.Notify(499490)

					res = append(res, '\\', '0'+(c>>6), '0'+((c>>3)&7), '0'+(c&7))
				} else {
					__antithesis_instrumentation__.Notify(499491)
					res = append(res, c)
				}
			}
		}
		__antithesis_instrumentation__.Notify(499479)
		return string(res)

	case BytesEncodeBase64:
		__antithesis_instrumentation__.Notify(499480)
		return base64.StdEncoding.EncodeToString([]byte(data))

	default:
		__antithesis_instrumentation__.Notify(499481)
		panic(errors.AssertionFailedf("unhandled format: %s", be))
	}
}

func DecodeRawBytesToByteArray(data string, be BytesEncodeFormat) ([]byte, error) {
	__antithesis_instrumentation__.Notify(499492)
	switch be {
	case BytesEncodeHex:
		__antithesis_instrumentation__.Notify(499493)
		return hex.DecodeString(data)

	case BytesEncodeEscape:
		__antithesis_instrumentation__.Notify(499494)

		res := make([]byte, 0, len(data))
		for i := 0; i < len(data); i++ {
			__antithesis_instrumentation__.Notify(499498)
			ch := data[i]
			if ch != '\\' {
				__antithesis_instrumentation__.Notify(499504)
				res = append(res, ch)
				continue
			} else {
				__antithesis_instrumentation__.Notify(499505)
			}
			__antithesis_instrumentation__.Notify(499499)
			if i >= len(data)-1 {
				__antithesis_instrumentation__.Notify(499506)
				return nil, pgerror.New(pgcode.InvalidEscapeSequence,
					"bytea encoded value ends with escape character")
			} else {
				__antithesis_instrumentation__.Notify(499507)
			}
			__antithesis_instrumentation__.Notify(499500)
			if data[i+1] == '\\' {
				__antithesis_instrumentation__.Notify(499508)
				res = append(res, '\\')
				i++
				continue
			} else {
				__antithesis_instrumentation__.Notify(499509)
			}
			__antithesis_instrumentation__.Notify(499501)
			if i+3 >= len(data) {
				__antithesis_instrumentation__.Notify(499510)
				return nil, pgerror.New(pgcode.InvalidEscapeSequence,
					"bytea encoded value ends with incomplete escape sequence")
			} else {
				__antithesis_instrumentation__.Notify(499511)
			}
			__antithesis_instrumentation__.Notify(499502)
			b := byte(0)
			for j := 1; j <= 3; j++ {
				__antithesis_instrumentation__.Notify(499512)
				octDigit := data[i+j]
				if octDigit < '0' || func() bool {
					__antithesis_instrumentation__.Notify(499514)
					return octDigit > '7' == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(499515)
					return (j == 1 && func() bool {
						__antithesis_instrumentation__.Notify(499516)
						return octDigit > '3' == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(499517)
					return nil, pgerror.New(pgcode.InvalidEscapeSequence,
						"invalid bytea escape sequence")
				} else {
					__antithesis_instrumentation__.Notify(499518)
				}
				__antithesis_instrumentation__.Notify(499513)
				b = (b << 3) | (octDigit - '0')
			}
			__antithesis_instrumentation__.Notify(499503)
			res = append(res, b)
			i += 3
		}
		__antithesis_instrumentation__.Notify(499495)
		return res, nil

	case BytesEncodeBase64:
		__antithesis_instrumentation__.Notify(499496)
		return base64.StdEncoding.DecodeString(data)

	default:
		__antithesis_instrumentation__.Notify(499497)
		return nil, errors.AssertionFailedf("unhandled format: %s", be)
	}
}

func DecodeRawBytesToByteArrayAuto(data []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(499519)
	if len(data) >= 2 && func() bool {
		__antithesis_instrumentation__.Notify(499521)
		return data[0] == '\\' == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(499522)
		return (data[1] == 'x' || func() bool {
			__antithesis_instrumentation__.Notify(499523)
			return data[1] == 'X' == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(499524)
		return DecodeRawBytesToByteArray(string(data[2:]), BytesEncodeHex)
	} else {
		__antithesis_instrumentation__.Notify(499525)
	}
	__antithesis_instrumentation__.Notify(499520)
	return DecodeRawBytesToByteArray(string(data), BytesEncodeEscape)
}

func (f BytesEncodeFormat) String() string {
	__antithesis_instrumentation__.Notify(499526)
	switch f {
	case BytesEncodeHex:
		__antithesis_instrumentation__.Notify(499527)
		return "hex"
	case BytesEncodeEscape:
		__antithesis_instrumentation__.Notify(499528)
		return "escape"
	case BytesEncodeBase64:
		__antithesis_instrumentation__.Notify(499529)
		return "base64"
	default:
		__antithesis_instrumentation__.Notify(499530)
		return fmt.Sprintf("invalid (%d)", f)
	}
}

func BytesEncodeFormatFromString(val string) (_ BytesEncodeFormat, ok bool) {
	__antithesis_instrumentation__.Notify(499531)
	switch strings.ToUpper(val) {
	case "HEX":
		__antithesis_instrumentation__.Notify(499532)
		return BytesEncodeHex, true
	case "ESCAPE":
		__antithesis_instrumentation__.Notify(499533)
		return BytesEncodeEscape, true
	case "BASE64":
		__antithesis_instrumentation__.Notify(499534)
		return BytesEncodeBase64, true
	default:
		__antithesis_instrumentation__.Notify(499535)
		return -1, false
	}
}
