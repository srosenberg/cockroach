package lexbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/util/stringencoding"
)

type EncodeFlags int

func (f EncodeFlags) HasFlags(subset EncodeFlags) bool {
	__antithesis_instrumentation__.Notify(499556)
	return f&subset == subset
}

const (
	EncNoFlags EncodeFlags = 0

	EncBareStrings EncodeFlags = 1 << iota

	EncBareIdentifiers

	EncFirstFreeFlagBit
)

func EncodeRestrictedSQLIdent(buf *bytes.Buffer, s string, flags EncodeFlags) {
	__antithesis_instrumentation__.Notify(499557)
	if flags.HasFlags(EncBareIdentifiers) || func() bool {
		__antithesis_instrumentation__.Notify(499559)
		return (!isReservedKeyword(s) && func() bool {
			__antithesis_instrumentation__.Notify(499560)
			return IsBareIdentifier(s) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(499561)
		buf.WriteString(s)
		return
	} else {
		__antithesis_instrumentation__.Notify(499562)
	}
	__antithesis_instrumentation__.Notify(499558)
	EncodeEscapedSQLIdent(buf, s)
}

func EncodeUnrestrictedSQLIdent(buf *bytes.Buffer, s string, flags EncodeFlags) {
	__antithesis_instrumentation__.Notify(499563)
	if flags.HasFlags(EncBareIdentifiers) || func() bool {
		__antithesis_instrumentation__.Notify(499565)
		return IsBareIdentifier(s) == true
	}() == true {
		__antithesis_instrumentation__.Notify(499566)
		buf.WriteString(s)
		return
	} else {
		__antithesis_instrumentation__.Notify(499567)
	}
	__antithesis_instrumentation__.Notify(499564)
	EncodeEscapedSQLIdent(buf, s)
}

func EncodeEscapedSQLIdent(buf *bytes.Buffer, s string) {
	__antithesis_instrumentation__.Notify(499568)
	buf.WriteByte('"')
	start := 0
	for i, n := 0, len(s); i < n; i++ {
		__antithesis_instrumentation__.Notify(499571)
		ch := s[i]

		if ch == '"' {
			__antithesis_instrumentation__.Notify(499572)
			if start != i {
				__antithesis_instrumentation__.Notify(499574)
				buf.WriteString(s[start:i])
			} else {
				__antithesis_instrumentation__.Notify(499575)
			}
			__antithesis_instrumentation__.Notify(499573)
			start = i + 1
			buf.WriteByte(ch)
			buf.WriteByte(ch)
		} else {
			__antithesis_instrumentation__.Notify(499576)
		}
	}
	__antithesis_instrumentation__.Notify(499569)
	if start < len(s) {
		__antithesis_instrumentation__.Notify(499577)
		buf.WriteString(s[start:])
	} else {
		__antithesis_instrumentation__.Notify(499578)
	}
	__antithesis_instrumentation__.Notify(499570)
	buf.WriteByte('"')
}

var mustQuoteMap = map[byte]bool{
	' ': true,
	',': true,
	'{': true,
	'}': true,
}

func EncodeSQLString(buf *bytes.Buffer, in string) {
	__antithesis_instrumentation__.Notify(499579)
	EncodeSQLStringWithFlags(buf, in, EncNoFlags)
}

func EscapeSQLString(in string) string {
	__antithesis_instrumentation__.Notify(499580)
	var buf bytes.Buffer
	EncodeSQLString(&buf, in)
	return buf.String()
}

func EncodeSQLStringWithFlags(buf *bytes.Buffer, in string, flags EncodeFlags) {
	__antithesis_instrumentation__.Notify(499581)

	start := 0
	escapedString := false
	bareStrings := flags.HasFlags(EncBareStrings)

	for i, r := range in {
		__antithesis_instrumentation__.Notify(499585)
		if i < start {
			__antithesis_instrumentation__.Notify(499590)
			continue
		} else {
			__antithesis_instrumentation__.Notify(499591)
		}
		__antithesis_instrumentation__.Notify(499586)
		ch := byte(r)
		if r >= 0x20 && func() bool {
			__antithesis_instrumentation__.Notify(499592)
			return r < 0x7F == true
		}() == true {
			__antithesis_instrumentation__.Notify(499593)
			if mustQuoteMap[ch] {
				__antithesis_instrumentation__.Notify(499595)

				bareStrings = false
			} else {
				__antithesis_instrumentation__.Notify(499596)
			}
			__antithesis_instrumentation__.Notify(499594)
			if !stringencoding.NeedEscape(ch) && func() bool {
				__antithesis_instrumentation__.Notify(499597)
				return ch != '\'' == true
			}() == true {
				__antithesis_instrumentation__.Notify(499598)
				continue
			} else {
				__antithesis_instrumentation__.Notify(499599)
			}
		} else {
			__antithesis_instrumentation__.Notify(499600)
		}
		__antithesis_instrumentation__.Notify(499587)

		if !escapedString {
			__antithesis_instrumentation__.Notify(499601)
			buf.WriteString("e'")
			escapedString = true
		} else {
			__antithesis_instrumentation__.Notify(499602)
		}
		__antithesis_instrumentation__.Notify(499588)
		buf.WriteString(in[start:i])
		ln := utf8.RuneLen(r)
		if ln < 0 {
			__antithesis_instrumentation__.Notify(499603)
			start = i + 1
		} else {
			__antithesis_instrumentation__.Notify(499604)
			start = i + ln
		}
		__antithesis_instrumentation__.Notify(499589)
		stringencoding.EncodeEscapedChar(buf, in, r, ch, i, '\'')
	}
	__antithesis_instrumentation__.Notify(499582)

	quote := !escapedString && func() bool {
		__antithesis_instrumentation__.Notify(499605)
		return !bareStrings == true
	}() == true
	if quote {
		__antithesis_instrumentation__.Notify(499606)
		buf.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(499607)
	}
	__antithesis_instrumentation__.Notify(499583)
	if start < len(in) {
		__antithesis_instrumentation__.Notify(499608)
		buf.WriteString(in[start:])
	} else {
		__antithesis_instrumentation__.Notify(499609)
	}
	__antithesis_instrumentation__.Notify(499584)
	if escapedString || func() bool {
		__antithesis_instrumentation__.Notify(499610)
		return quote == true
	}() == true {
		__antithesis_instrumentation__.Notify(499611)
		buf.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(499612)
	}
}

func EncodeSQLBytes(buf *bytes.Buffer, in string) {
	__antithesis_instrumentation__.Notify(499613)
	buf.WriteString("b'")
	EncodeSQLBytesInner(buf, in)
	buf.WriteByte('\'')
}

func EncodeSQLBytesInner(buf *bytes.Buffer, in string) {
	__antithesis_instrumentation__.Notify(499614)
	start := 0

	for i, n := 0, len(in); i < n; i++ {
		__antithesis_instrumentation__.Notify(499616)
		ch := in[i]
		if encodedChar := stringencoding.EncodeMap[ch]; encodedChar != stringencoding.DontEscape {
			__antithesis_instrumentation__.Notify(499617)
			buf.WriteString(in[start:i])
			buf.WriteByte('\\')
			buf.WriteByte(encodedChar)
			start = i + 1
		} else {
			__antithesis_instrumentation__.Notify(499618)
			if ch == '\'' {
				__antithesis_instrumentation__.Notify(499619)

				buf.WriteString(in[start:i])
				buf.WriteByte('\\')
				buf.WriteByte(ch)
				start = i + 1
			} else {
				__antithesis_instrumentation__.Notify(499620)
				if ch < 0x20 || func() bool {
					__antithesis_instrumentation__.Notify(499621)
					return ch >= 0x7F == true
				}() == true {
					__antithesis_instrumentation__.Notify(499622)
					buf.WriteString(in[start:i])

					buf.Write(stringencoding.HexMap[ch])
					start = i + 1
				} else {
					__antithesis_instrumentation__.Notify(499623)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(499615)
	buf.WriteString(in[start:])
}
