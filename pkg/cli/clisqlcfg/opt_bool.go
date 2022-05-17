package clisqlcfg

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

type OptBool struct {
	isDef bool
	v     bool
}

func (o OptBool) Get() (isDef, val bool) {
	__antithesis_instrumentation__.Notify(28516)
	return o.isDef, o.v
}

func (o OptBool) String() string {
	__antithesis_instrumentation__.Notify(28517)
	if !o.isDef {
		__antithesis_instrumentation__.Notify(28519)
		return "<unspecified>"
	} else {
		__antithesis_instrumentation__.Notify(28520)
	}
	__antithesis_instrumentation__.Notify(28518)
	return strconv.FormatBool(o.v)
}

func (o OptBool) Type() string { __antithesis_instrumentation__.Notify(28521); return "bool" }

func (o *OptBool) Set(v string) error {
	__antithesis_instrumentation__.Notify(28522)
	b, err := strconv.ParseBool(v)
	if err != nil {
		__antithesis_instrumentation__.Notify(28524)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28525)
	}
	__antithesis_instrumentation__.Notify(28523)
	o.isDef = true
	o.v = b
	return nil
}
