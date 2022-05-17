package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strings"

type SchemaFeatureName string

func GetSchemaFeatureNameFromStmt(stmt Statement) SchemaFeatureName {
	__antithesis_instrumentation__.Notify(613035)
	statementTag := stmt.StatementTag()
	statementInfo := strings.Split(statementTag, " ")

	if len(statementInfo) >= 2 {
		__antithesis_instrumentation__.Notify(613037)
		return SchemaFeatureName(statementInfo[0] + " " + statementInfo[1])
	} else {
		__antithesis_instrumentation__.Notify(613038)
	}
	__antithesis_instrumentation__.Notify(613036)
	return SchemaFeatureName(statementInfo[0])
}
