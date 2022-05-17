package sessiondatapb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/errors"
)

type ExperimentalDistSQLPlanningMode int64

const (
	ExperimentalDistSQLPlanningOff ExperimentalDistSQLPlanningMode = iota

	ExperimentalDistSQLPlanningOn

	ExperimentalDistSQLPlanningAlways
)

func (m ExperimentalDistSQLPlanningMode) String() string {
	__antithesis_instrumentation__.Notify(617978)
	switch m {
	case ExperimentalDistSQLPlanningOff:
		__antithesis_instrumentation__.Notify(617979)
		return "off"
	case ExperimentalDistSQLPlanningOn:
		__antithesis_instrumentation__.Notify(617980)
		return "on"
	case ExperimentalDistSQLPlanningAlways:
		__antithesis_instrumentation__.Notify(617981)
		return "always"
	default:
		__antithesis_instrumentation__.Notify(617982)
		return fmt.Sprintf("invalid (%d)", m)
	}
}

func ExperimentalDistSQLPlanningModeFromString(val string) (ExperimentalDistSQLPlanningMode, bool) {
	__antithesis_instrumentation__.Notify(617983)
	var m ExperimentalDistSQLPlanningMode
	switch strings.ToUpper(val) {
	case "OFF":
		__antithesis_instrumentation__.Notify(617985)
		m = ExperimentalDistSQLPlanningOff
	case "ON":
		__antithesis_instrumentation__.Notify(617986)
		m = ExperimentalDistSQLPlanningOn
	case "ALWAYS":
		__antithesis_instrumentation__.Notify(617987)
		m = ExperimentalDistSQLPlanningAlways
	default:
		__antithesis_instrumentation__.Notify(617988)
		return 0, false
	}
	__antithesis_instrumentation__.Notify(617984)
	return m, true
}

type DistSQLExecMode int64

const (
	DistSQLOff DistSQLExecMode = iota

	DistSQLAuto

	DistSQLOn

	DistSQLAlways
)

func (m DistSQLExecMode) String() string {
	__antithesis_instrumentation__.Notify(617989)
	switch m {
	case DistSQLOff:
		__antithesis_instrumentation__.Notify(617990)
		return "off"
	case DistSQLAuto:
		__antithesis_instrumentation__.Notify(617991)
		return "auto"
	case DistSQLOn:
		__antithesis_instrumentation__.Notify(617992)
		return "on"
	case DistSQLAlways:
		__antithesis_instrumentation__.Notify(617993)
		return "always"
	default:
		__antithesis_instrumentation__.Notify(617994)
		return fmt.Sprintf("invalid (%d)", m)
	}
}

func DistSQLExecModeFromString(val string) (_ DistSQLExecMode, ok bool) {
	__antithesis_instrumentation__.Notify(617995)
	switch strings.ToUpper(val) {
	case "OFF":
		__antithesis_instrumentation__.Notify(617996)
		return DistSQLOff, true
	case "AUTO":
		__antithesis_instrumentation__.Notify(617997)
		return DistSQLAuto, true
	case "ON":
		__antithesis_instrumentation__.Notify(617998)
		return DistSQLOn, true
	case "ALWAYS":
		__antithesis_instrumentation__.Notify(617999)
		return DistSQLAlways, true
	default:
		__antithesis_instrumentation__.Notify(618000)
		return 0, false
	}
}

type SerialNormalizationMode int64

const (
	SerialUsesRowID SerialNormalizationMode = 0

	SerialUsesVirtualSequences SerialNormalizationMode = 1

	SerialUsesSQLSequences SerialNormalizationMode = 2

	SerialUsesCachedSQLSequences SerialNormalizationMode = 3

	SerialUsesUnorderedRowID SerialNormalizationMode = 4
)

func (m SerialNormalizationMode) String() string {
	__antithesis_instrumentation__.Notify(618001)
	switch m {
	case SerialUsesRowID:
		__antithesis_instrumentation__.Notify(618002)
		return "rowid"
	case SerialUsesUnorderedRowID:
		__antithesis_instrumentation__.Notify(618003)
		return "unordered_rowid"
	case SerialUsesVirtualSequences:
		__antithesis_instrumentation__.Notify(618004)
		return "virtual_sequence"
	case SerialUsesSQLSequences:
		__antithesis_instrumentation__.Notify(618005)
		return "sql_sequence"
	case SerialUsesCachedSQLSequences:
		__antithesis_instrumentation__.Notify(618006)
		return "sql_sequence_cached"
	default:
		__antithesis_instrumentation__.Notify(618007)
		return fmt.Sprintf("invalid (%d)", m)
	}
}

func SerialNormalizationModeFromString(val string) (_ SerialNormalizationMode, ok bool) {
	__antithesis_instrumentation__.Notify(618008)
	switch strings.ToUpper(val) {
	case "ROWID":
		__antithesis_instrumentation__.Notify(618009)
		return SerialUsesRowID, true
	case "UNORDERED_ROWID":
		__antithesis_instrumentation__.Notify(618010)
		return SerialUsesUnorderedRowID, true
	case "VIRTUAL_SEQUENCE":
		__antithesis_instrumentation__.Notify(618011)
		return SerialUsesVirtualSequences, true
	case "SQL_SEQUENCE":
		__antithesis_instrumentation__.Notify(618012)
		return SerialUsesSQLSequences, true
	case "SQL_SEQUENCE_CACHED":
		__antithesis_instrumentation__.Notify(618013)
		return SerialUsesCachedSQLSequences, true
	default:
		__antithesis_instrumentation__.Notify(618014)
		return 0, false
	}
}

type NewSchemaChangerMode int64

const (
	UseNewSchemaChangerOff NewSchemaChangerMode = iota

	UseNewSchemaChangerOn

	UseNewSchemaChangerUnsafe

	UseNewSchemaChangerUnsafeAlways
)

func (m NewSchemaChangerMode) String() string {
	__antithesis_instrumentation__.Notify(618015)
	switch m {
	case UseNewSchemaChangerOff:
		__antithesis_instrumentation__.Notify(618016)
		return "off"
	case UseNewSchemaChangerOn:
		__antithesis_instrumentation__.Notify(618017)
		return "on"
	case UseNewSchemaChangerUnsafe:
		__antithesis_instrumentation__.Notify(618018)
		return "unsafe"
	case UseNewSchemaChangerUnsafeAlways:
		__antithesis_instrumentation__.Notify(618019)
		return "unsafe_always"
	default:
		__antithesis_instrumentation__.Notify(618020)
		return fmt.Sprintf("invalid (%d)", m)
	}
}

func NewSchemaChangerModeFromString(val string) (_ NewSchemaChangerMode, ok bool) {
	__antithesis_instrumentation__.Notify(618021)
	switch strings.ToUpper(val) {
	case "OFF":
		__antithesis_instrumentation__.Notify(618022)
		return UseNewSchemaChangerOff, true
	case "ON":
		__antithesis_instrumentation__.Notify(618023)
		return UseNewSchemaChangerOn, true
	case "UNSAFE":
		__antithesis_instrumentation__.Notify(618024)
		return UseNewSchemaChangerUnsafe, true
	case "UNSAFE_ALWAYS":
		__antithesis_instrumentation__.Notify(618025)
		return UseNewSchemaChangerUnsafeAlways, true
	default:
		__antithesis_instrumentation__.Notify(618026)
		return 0, false
	}
}

type QoSLevel admission.WorkPriority

const (
	SystemLow = QoSLevel(admission.LowPri)

	TTLStatsLow = QoSLevel(admission.TTLLowPri)

	TTLLow = QoSLevel(admission.TTLLowPri)

	UserLow = QoSLevel(admission.UserLowPri)

	Normal = QoSLevel(admission.NormalPri)

	UserHigh = QoSLevel(admission.UserHighPri)

	Locking = QoSLevel(admission.LockingPri)

	SystemHigh = QoSLevel(admission.HighPri)
)

const (
	NormalName = "regular"

	UserHighName = "critical"

	UserLowName = "background"

	SystemHighName = "maximum"

	SystemLowName = "minimum"

	TTLLowName = "ttl_low"

	LockingName = "locking"
)

var qosLevelsDict = map[QoSLevel]string{
	SystemLow:  SystemLowName,
	TTLLow:     TTLLowName,
	UserLow:    UserLowName,
	Normal:     NormalName,
	UserHigh:   UserHighName,
	Locking:    LockingName,
	SystemHigh: SystemHighName,
}

func ParseQoSLevelFromString(val string) (_ QoSLevel, ok bool) {
	__antithesis_instrumentation__.Notify(618027)
	switch strings.ToUpper(val) {
	case strings.ToUpper(UserHighName):
		__antithesis_instrumentation__.Notify(618028)
		return UserHigh, true
	case strings.ToUpper(UserLowName):
		__antithesis_instrumentation__.Notify(618029)
		return UserLow, true
	case strings.ToUpper(NormalName):
		__antithesis_instrumentation__.Notify(618030)
		return Normal, true
	default:
		__antithesis_instrumentation__.Notify(618031)
		return 0, false
	}
}

func (e QoSLevel) String() string {
	__antithesis_instrumentation__.Notify(618032)
	if name, ok := qosLevelsDict[e]; ok {
		__antithesis_instrumentation__.Notify(618034)
		return name
	} else {
		__antithesis_instrumentation__.Notify(618035)
	}
	__antithesis_instrumentation__.Notify(618033)
	return fmt.Sprintf("%d", int(e))
}

func ToQoSLevelString(value int32) string {
	__antithesis_instrumentation__.Notify(618036)
	if value > int32(SystemHigh) || func() bool {
		__antithesis_instrumentation__.Notify(618038)
		return value < int32(SystemLow) == true
	}() == true {
		__antithesis_instrumentation__.Notify(618039)
		return fmt.Sprintf("%d", value)
	} else {
		__antithesis_instrumentation__.Notify(618040)
	}
	__antithesis_instrumentation__.Notify(618037)
	qosLevel := QoSLevel(value)
	return qosLevel.String()
}

func (e QoSLevel) Validate() QoSLevel {
	__antithesis_instrumentation__.Notify(618041)
	switch e {
	case Normal, UserHigh, UserLow:
		__antithesis_instrumentation__.Notify(618042)
		return e
	default:
		__antithesis_instrumentation__.Notify(618043)
		panic(errors.AssertionFailedf("use of illegal user QoSLevel: %s", e.String()))
	}
}

func (e QoSLevel) ValidateInternal() QoSLevel {
	__antithesis_instrumentation__.Notify(618044)
	if _, ok := qosLevelsDict[e]; ok {
		__antithesis_instrumentation__.Notify(618046)
		return e
	} else {
		__antithesis_instrumentation__.Notify(618047)
	}
	__antithesis_instrumentation__.Notify(618045)
	panic(errors.AssertionFailedf("use of illegal internal QoSLevel: %s", e.String()))
}
