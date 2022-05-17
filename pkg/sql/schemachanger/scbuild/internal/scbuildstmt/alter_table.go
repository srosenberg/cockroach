package scbuildstmt

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

var supportedAlterTableStatements = map[reflect.Type]supportedStatement{
	reflect.TypeOf((*tree.AlterTableAddColumn)(nil)): {alterTableAddColumn, false},
}

func init() {

	for statementType, statementEntry := range supportedAlterTableStatements {
		callBackType := reflect.TypeOf(statementEntry.fn)
		if callBackType.Kind() != reflect.Func {
			panic(errors.AssertionFailedf("%v entry for statement is "+
				"not a function", statementType))
		}
		if callBackType.NumIn() != 4 ||
			!callBackType.In(0).Implements(reflect.TypeOf((*BuildCtx)(nil)).Elem()) ||
			callBackType.In(1) != reflect.TypeOf((*tree.TableName)(nil)) ||
			callBackType.In(2) != reflect.TypeOf((*scpb.Table)(nil)) ||
			callBackType.In(3) != statementType {
			panic(errors.AssertionFailedf("%v entry for alter table statement "+
				"does not have a valid signature got %v", statementType, callBackType))
		}
	}
}

func AlterTable(b BuildCtx, n *tree.AlterTable) {
	__antithesis_instrumentation__.Notify(579723)

	n.HoistAddColumnConstraints()

	for _, cmd := range n.Cmds {
		__antithesis_instrumentation__.Notify(579727)
		info, ok := supportedAlterTableStatements[reflect.TypeOf(cmd)]
		if !ok {
			__antithesis_instrumentation__.Notify(579729)
			panic(scerrors.NotImplementedError(cmd))
		} else {
			__antithesis_instrumentation__.Notify(579730)
		}
		__antithesis_instrumentation__.Notify(579728)

		if !info.IsFullySupported(b.EvalCtx().SessionData().NewSchemaChangerMode) {
			__antithesis_instrumentation__.Notify(579731)
			panic(scerrors.NotImplementedError(cmd))
		} else {
			__antithesis_instrumentation__.Notify(579732)
		}
	}
	__antithesis_instrumentation__.Notify(579724)
	tn := n.Table.ToTableName()
	elts := b.ResolveTable(n.Table, ResolveParams{
		IsExistenceOptional: n.IfExists,
		RequiredPrivilege:   privilege.CREATE,
	})
	_, target, tbl := scpb.FindTable(elts)
	if tbl == nil {
		__antithesis_instrumentation__.Notify(579733)
		b.MarkNameAsNonExistent(&tn)
		return
	} else {
		__antithesis_instrumentation__.Notify(579734)
	}
	__antithesis_instrumentation__.Notify(579725)
	if target != scpb.ToPublic {
		__antithesis_instrumentation__.Notify(579735)
		panic(pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
			"table %q is being dropped, try again later", n.Table.Object()))
	} else {
		__antithesis_instrumentation__.Notify(579736)
	}
	__antithesis_instrumentation__.Notify(579726)
	tn.ObjectNamePrefix = b.NamePrefix(tbl)
	b.SetUnresolvedNameAnnotation(n.Table, &tn)
	b.IncrementSchemaChangeAlterCounter("table")
	for _, cmd := range n.Cmds {
		__antithesis_instrumentation__.Notify(579737)
		info := supportedAlterTableStatements[reflect.TypeOf(cmd)]

		fn := reflect.ValueOf(info.fn)
		fn.Call([]reflect.Value{
			reflect.ValueOf(b),
			reflect.ValueOf(&tn),
			reflect.ValueOf(tbl),
			reflect.ValueOf(cmd),
		})
		b.IncrementSubWorkID()
	}
}
