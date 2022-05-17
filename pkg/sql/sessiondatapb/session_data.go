package sessiondatapb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
)

func (c DataConversionConfig) GetFloatPrec() int {
	__antithesis_instrumentation__.Notify(619867)

	if c.ExtraFloatDigits >= 3 {
		__antithesis_instrumentation__.Notify(619870)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(619871)
	}
	__antithesis_instrumentation__.Notify(619868)

	const StdDoubleDigits = 15

	nDigits := StdDoubleDigits + c.ExtraFloatDigits
	if nDigits < 1 {
		__antithesis_instrumentation__.Notify(619872)

		nDigits = 1
	} else {
		__antithesis_instrumentation__.Notify(619873)
	}
	__antithesis_instrumentation__.Notify(619869)
	return int(nDigits)
}

func (m VectorizeExecMode) String() string {
	__antithesis_instrumentation__.Notify(619874)
	switch m {
	case VectorizeOn, VectorizeUnset:
		__antithesis_instrumentation__.Notify(619875)
		return "on"
	case VectorizeExperimentalAlways:
		__antithesis_instrumentation__.Notify(619876)
		return "experimental_always"
	case VectorizeOff:
		__antithesis_instrumentation__.Notify(619877)
		return "off"
	default:
		__antithesis_instrumentation__.Notify(619878)
		return fmt.Sprintf("invalid (%d)", m)
	}
}

func VectorizeExecModeFromString(val string) (VectorizeExecMode, bool) {
	__antithesis_instrumentation__.Notify(619879)
	var m VectorizeExecMode
	switch strings.ToUpper(val) {
	case "ON":
		__antithesis_instrumentation__.Notify(619881)
		m = VectorizeOn
	case "EXPERIMENTAL_ALWAYS":
		__antithesis_instrumentation__.Notify(619882)
		m = VectorizeExperimentalAlways
	case "OFF":
		__antithesis_instrumentation__.Notify(619883)
		m = VectorizeOff
	default:
		__antithesis_instrumentation__.Notify(619884)
		return 0, false
	}
	__antithesis_instrumentation__.Notify(619880)
	return m, true
}

func (s *SessionData) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(619885)
	return s.UserProto.Decode()
}
