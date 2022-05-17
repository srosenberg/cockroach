package scbuildstmt

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/errors"
)

type supportedStatement struct {
	fn             interface{}
	fullySupported bool
}

func (s supportedStatement) IsFullySupported(mode sessiondatapb.NewSchemaChangerMode) bool {
	__antithesis_instrumentation__.Notify(580239)

	if mode == sessiondatapb.UseNewSchemaChangerUnsafeAlways || func() bool {
		__antithesis_instrumentation__.Notify(580241)
		return mode == sessiondatapb.UseNewSchemaChangerUnsafe == true
	}() == true {
		__antithesis_instrumentation__.Notify(580242)
		return true
	} else {
		__antithesis_instrumentation__.Notify(580243)
	}
	__antithesis_instrumentation__.Notify(580240)
	return s.fullySupported
}

var supportedStatements = map[reflect.Type]supportedStatement{

	reflect.TypeOf((*tree.AlterTable)(nil)):   {AlterTable, true},
	reflect.TypeOf((*tree.CreateIndex)(nil)):  {CreateIndex, false},
	reflect.TypeOf((*tree.DropDatabase)(nil)): {DropDatabase, true},
	reflect.TypeOf((*tree.DropSchema)(nil)):   {DropSchema, true},
	reflect.TypeOf((*tree.DropSequence)(nil)): {DropSequence, true},
	reflect.TypeOf((*tree.DropTable)(nil)):    {DropTable, true},
	reflect.TypeOf((*tree.DropType)(nil)):     {DropType, true},
	reflect.TypeOf((*tree.DropView)(nil)):     {DropView, true},
}

func init() {

	for statementType, statementEntry := range supportedStatements {
		callBackType := reflect.TypeOf(statementEntry.fn)
		if callBackType.Kind() != reflect.Func {
			panic(errors.AssertionFailedf("%v entry for statement is "+
				"not a function", statementType))
		}
		if callBackType.NumIn() != 2 ||
			!callBackType.In(0).Implements(reflect.TypeOf((*BuildCtx)(nil)).Elem()) ||
			callBackType.In(1) != statementType {
			panic(errors.AssertionFailedf("%v entry for statement is "+
				"does not have a valid signature got %v", statementType, callBackType))
		}
	}
}

func Process(b BuildCtx, n tree.Statement) {
	__antithesis_instrumentation__.Notify(580244)

	info, ok := supportedStatements[reflect.TypeOf(n)]
	if !ok {
		__antithesis_instrumentation__.Notify(580248)
		panic(scerrors.NotImplementedError(n))
	} else {
		__antithesis_instrumentation__.Notify(580249)
	}
	__antithesis_instrumentation__.Notify(580245)

	if !info.IsFullySupported(b.EvalCtx().SessionData().NewSchemaChangerMode) {
		__antithesis_instrumentation__.Notify(580250)
		panic(scerrors.NotImplementedError(n))
	} else {
		__antithesis_instrumentation__.Notify(580251)
	}
	__antithesis_instrumentation__.Notify(580246)

	fn := reflect.ValueOf(info.fn)
	in := []reflect.Value{reflect.ValueOf(b), reflect.ValueOf(n)}

	err := b.CheckFeature(b, tree.GetSchemaFeatureNameFromStmt(n))
	if err != nil {
		__antithesis_instrumentation__.Notify(580252)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580253)
	}
	__antithesis_instrumentation__.Notify(580247)
	fn.Call(in)
}
