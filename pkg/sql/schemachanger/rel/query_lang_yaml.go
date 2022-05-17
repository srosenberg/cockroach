package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

func valueForYAML(v interface{}) interface{} {
	__antithesis_instrumentation__.Notify(578926)
	switch v := v.(type) {
	case fmt.Stringer:
		__antithesis_instrumentation__.Notify(578927)
		return v.String()
	default:
		__antithesis_instrumentation__.Notify(578928)
		return v
	}
}

func (v valueExpr) encoded() interface{} {
	__antithesis_instrumentation__.Notify(578929)
	return valueForYAML(v.value)
}

func (a anyExpr) encoded() interface{} {
	__antithesis_instrumentation__.Notify(578930)
	ret := make([]interface{}, 0, len(a))
	for _, v := range a {
		__antithesis_instrumentation__.Notify(578932)
		ret = append(ret, valueForYAML(v))
	}
	__antithesis_instrumentation__.Notify(578931)
	return ret
}

func (v Var) encoded() interface{} {
	__antithesis_instrumentation__.Notify(578933)
	return "$" + string(v)
}

func (e *eqDecl) MarshalYAML() (interface{}, error) {
	__antithesis_instrumentation__.Notify(578934)
	return clauseStr("$"+string(e.v), e.expr)
}

func exprToString(e expr) (string, error) {
	__antithesis_instrumentation__.Notify(578935)
	var expr yaml.Node
	if err := expr.Encode(e.encoded()); err != nil {
		__antithesis_instrumentation__.Notify(578938)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(578939)
	}
	__antithesis_instrumentation__.Notify(578936)
	expr.Style = yaml.FlowStyle
	out, err := yaml.Marshal(&expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(578940)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(578941)
	}
	__antithesis_instrumentation__.Notify(578937)
	return strings.TrimSpace(string(out)), err
}

func (f *tripleDecl) MarshalYAML() (interface{}, error) {
	__antithesis_instrumentation__.Notify(578942)
	return clauseStr(fmt.Sprintf("$%s[%s]", f.entity, f.attribute), f.value)
}

func clauseStr(lhs string, rhs expr) (string, error) {
	__antithesis_instrumentation__.Notify(578943)
	rhsStr, err := exprToString(rhs)
	if err != nil {
		__antithesis_instrumentation__.Notify(578946)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(578947)
	}
	__antithesis_instrumentation__.Notify(578944)
	op := "="
	if _, isAny := rhs.(anyExpr); isAny {
		__antithesis_instrumentation__.Notify(578948)
		op = "IN"
	} else {
		__antithesis_instrumentation__.Notify(578949)
	}
	__antithesis_instrumentation__.Notify(578945)
	return fmt.Sprintf("%s %s %s", lhs, op, rhsStr), nil
}

func (f filterDecl) MarshalYAML() (interface{}, error) {
	__antithesis_instrumentation__.Notify(578950)
	var buf strings.Builder
	buf.WriteString(f.name)
	buf.WriteString("(")
	ft := reflect.TypeOf(f.predicateFunc)
	for i := 0; i < ft.NumIn(); i++ {
		__antithesis_instrumentation__.Notify(578953)
		if i > 0 {
			__antithesis_instrumentation__.Notify(578955)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(578956)
		}
		__antithesis_instrumentation__.Notify(578954)
		buf.WriteString(ft.In(i).String())
	}
	__antithesis_instrumentation__.Notify(578951)
	buf.WriteString(")(")
	for i, v := range f.vars {
		__antithesis_instrumentation__.Notify(578957)
		if i > 0 {
			__antithesis_instrumentation__.Notify(578959)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(578960)
		}
		__antithesis_instrumentation__.Notify(578958)
		buf.WriteString("$")
		buf.WriteString(string(v))
	}
	__antithesis_instrumentation__.Notify(578952)
	buf.WriteString(")")
	return buf.String(), nil
}
