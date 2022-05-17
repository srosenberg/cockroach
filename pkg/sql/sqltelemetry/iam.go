package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

const (
	Role = "role"

	User = "user"

	AlterRole = "alter"

	CreateRole = "create"

	OnDatabase = "on_database"

	OnSchema = "on_schema"

	OnTable = "on_table"

	OnType = "on_type"

	OnAllTablesInSchema = "on_all_tables_in_schemas"

	iamRoles = "iam.roles"
)

func IncIAMOptionCounter(opName string, option string) {
	__antithesis_instrumentation__.Notify(625784)
	telemetry.Inc(telemetry.GetCounter(
		fmt.Sprintf("%s.%s.%s", iamRoles, opName, option)))
}

func IncIAMCreateCounter(typ string) {
	__antithesis_instrumentation__.Notify(625785)
	telemetry.Inc(telemetry.GetCounter(
		fmt.Sprintf("%s.%s.%s", iamRoles, "create", typ)))
}

func IncIAMAlterCounter(typ string) {
	__antithesis_instrumentation__.Notify(625786)
	telemetry.Inc(telemetry.GetCounter(
		fmt.Sprintf("%s.%s.%s", iamRoles, "alter", typ)))
}

func IncIAMDropCounter(typ string) {
	__antithesis_instrumentation__.Notify(625787)
	telemetry.Inc(telemetry.GetCounter(
		fmt.Sprintf("%s.%s.%s", iamRoles, "drop", typ)))
}

func IncIAMGrantCounter(withAdmin bool) {
	__antithesis_instrumentation__.Notify(625788)
	var s string
	if withAdmin {
		__antithesis_instrumentation__.Notify(625790)
		s = fmt.Sprintf("%s.%s.with_admin", iamRoles, "grant")
	} else {
		__antithesis_instrumentation__.Notify(625791)
		s = fmt.Sprintf("%s.%s", iamRoles, "grant")
	}
	__antithesis_instrumentation__.Notify(625789)
	telemetry.Inc(telemetry.GetCounter(s))
}

func IncIAMRevokeCounter(withAdmin bool) {
	__antithesis_instrumentation__.Notify(625792)
	var s string
	if withAdmin {
		__antithesis_instrumentation__.Notify(625794)
		s = fmt.Sprintf("%s.%s.with_admin", iamRoles, "revoke")
	} else {
		__antithesis_instrumentation__.Notify(625795)
		s = fmt.Sprintf("%s.%s", iamRoles, "revoke")
	}
	__antithesis_instrumentation__.Notify(625793)
	telemetry.Inc(telemetry.GetCounter(s))
}

func IncIAMGrantPrivilegesCounter(on string) {
	__antithesis_instrumentation__.Notify(625796)
	telemetry.Inc(telemetry.GetCounter(
		fmt.Sprintf("%s.%s.%s.%s", iamRoles, "grant", "privileges", on)))
}

func IncIAMRevokePrivilegesCounter(on string) {
	__antithesis_instrumentation__.Notify(625797)
	telemetry.Inc(telemetry.GetCounter(
		fmt.Sprintf("%s.%s.%s.%s", iamRoles, "revoke", "privileges", on)))
}

var TurnConnAuditingOnUseCounter = telemetry.GetCounterOnce("auditing.connection.enabled")

var TurnConnAuditingOffUseCounter = telemetry.GetCounterOnce("auditing.connection.disabled")

var TurnAuthAuditingOnUseCounter = telemetry.GetCounterOnce("auditing.authentication.enabled")

var TurnAuthAuditingOffUseCounter = telemetry.GetCounterOnce("auditing.authentication.disabled")
