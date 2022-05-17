package opgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func newLogEventOp(e scpb.Element, md targetsWithElementMap) *scop.LogEvent {
	__antithesis_instrumentation__.Notify(593994)
	idx, ok := md.elementToTarget[e]
	if !ok {
		__antithesis_instrumentation__.Notify(593996)
		panic(errors.AssertionFailedf(
			"could not find element %s in target state", screl.ElementString(e),
		))
	} else {
		__antithesis_instrumentation__.Notify(593997)
	}
	__antithesis_instrumentation__.Notify(593995)
	t := md.Targets[idx]
	return &scop.LogEvent{
		TargetMetadata: *protoutil.Clone(&t.Metadata).(*scpb.TargetMetadata),
		Authorization:  *protoutil.Clone(&md.Authorization).(*scpb.Authorization),
		Statement:      md.Statements[t.Metadata.StatementID].RedactedStatement,
		StatementTag:   md.Statements[t.Metadata.StatementID].StatementTag,
		Element:        *protoutil.Clone(&t.ElementProto).(*scpb.ElementProto),
		TargetStatus:   t.TargetStatus,
	}
}

func statementForDropJob(e scpb.Element, md targetsWithElementMap) scop.StatementForDropJob {
	__antithesis_instrumentation__.Notify(593998)
	stmtID := md.Targets[md.elementToTarget[e]].Metadata.StatementID
	return scop.StatementForDropJob{

		Statement: redact.RedactableString(
			md.Statements[stmtID].RedactedStatement,
		).StripMarkers(),
		StatementID: stmtID,
		Rollback:    md.InRollback,
	}
}

type targetsWithElementMap struct {
	scpb.TargetState
	elementToTarget map[scpb.Element]int
	InRollback      bool
}

func makeTargetsWithElementMap(cs scpb.CurrentState) targetsWithElementMap {
	__antithesis_instrumentation__.Notify(593999)
	md := targetsWithElementMap{
		InRollback:      cs.InRollback,
		TargetState:     cs.TargetState,
		elementToTarget: make(map[scpb.Element]int),
	}
	for i := range cs.Targets {
		__antithesis_instrumentation__.Notify(594001)
		e := cs.Targets[i].Element()
		if prev, exists := md.elementToTarget[e]; exists {
			__antithesis_instrumentation__.Notify(594003)
			panic(errors.AssertionFailedf(
				"duplicate targets for %s: %v and %v", screl.ElementString(e),
				cs.Targets[i].TargetStatus, cs.Targets[prev].TargetStatus,
			))
		} else {
			__antithesis_instrumentation__.Notify(594004)
		}
		__antithesis_instrumentation__.Notify(594002)
		md.elementToTarget[e] = i
	}
	__antithesis_instrumentation__.Notify(594000)
	return md
}

type opsFunc func(element scpb.Element, md targetsWithElementMap) []scop.Op

func makeOpsFunc(el scpb.Element, fns []interface{}) (opsFunc, error) {
	__antithesis_instrumentation__.Notify(594005)
	var funcValues []reflect.Value
	for _, fn := range fns {
		__antithesis_instrumentation__.Notify(594007)
		if err := checkOpFunc(el, fn); err != nil {
			__antithesis_instrumentation__.Notify(594009)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(594010)
		}
		__antithesis_instrumentation__.Notify(594008)
		funcValues = append(funcValues, reflect.ValueOf(fn))
	}
	__antithesis_instrumentation__.Notify(594006)
	return func(element scpb.Element, md targetsWithElementMap) []scop.Op {
		__antithesis_instrumentation__.Notify(594011)
		ret := make([]scop.Op, 0, len(funcValues))
		in := []reflect.Value{reflect.ValueOf(element)}
		inWithMeta := []reflect.Value{reflect.ValueOf(element), reflect.ValueOf(md)}
		for _, fn := range funcValues {
			__antithesis_instrumentation__.Notify(594013)
			var out []reflect.Value
			if fn.Type().NumIn() == 1 {
				__antithesis_instrumentation__.Notify(594015)
				out = fn.Call(in)
			} else {
				__antithesis_instrumentation__.Notify(594016)
				out = fn.Call(inWithMeta)
			}
			__antithesis_instrumentation__.Notify(594014)
			if !out[0].IsNil() {
				__antithesis_instrumentation__.Notify(594017)
				ret = append(ret, out[0].Interface().(scop.Op))
			} else {
				__antithesis_instrumentation__.Notify(594018)
			}
		}
		__antithesis_instrumentation__.Notify(594012)
		return ret
	}, nil
}

var opType = reflect.TypeOf((*scop.Op)(nil)).Elem()

func checkOpFunc(el scpb.Element, fn interface{}) error {
	__antithesis_instrumentation__.Notify(594019)
	fnV := reflect.ValueOf(fn)
	fnT := fnV.Type()
	if fnT.Kind() != reflect.Func {
		__antithesis_instrumentation__.Notify(594023)
		return errors.Errorf(
			"%v is a %s, expected %s", fnT, fnT.Kind(), reflect.Func,
		)
	} else {
		__antithesis_instrumentation__.Notify(594024)
	}
	__antithesis_instrumentation__.Notify(594020)
	elType := reflect.TypeOf(el)
	if !(fnT.NumIn() == 1 && func() bool {
		__antithesis_instrumentation__.Notify(594025)
		return fnT.In(0) == elType == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(594026)
		return !(fnT.NumIn() == 2 && func() bool {
			__antithesis_instrumentation__.Notify(594027)
			return fnT.In(0) == elType == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(594028)
			return fnT.In(1) == reflect.TypeOf(targetsWithElementMap{}) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(594029)
		return errors.Errorf(
			"expected %v to be a func with one argument of type %s", fnT, elType,
		)
	} else {
		__antithesis_instrumentation__.Notify(594030)
	}
	__antithesis_instrumentation__.Notify(594021)
	if fnT.NumOut() != 1 || func() bool {
		__antithesis_instrumentation__.Notify(594031)
		return !fnT.Out(0).Implements(opType) == true
	}() == true {
		__antithesis_instrumentation__.Notify(594032)
		return errors.Errorf(
			"expected %v to be a func with one return value of type %s", fnT, opType,
		)
	} else {
		__antithesis_instrumentation__.Notify(594033)
	}
	__antithesis_instrumentation__.Notify(594022)
	return nil
}
