package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
)

func IsError(err error, re string) bool {
	__antithesis_instrumentation__.Notify(644201)
	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(644205)
		return re == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(644206)
		return true
	} else {
		__antithesis_instrumentation__.Notify(644207)
	}
	__antithesis_instrumentation__.Notify(644202)
	if err == nil || func() bool {
		__antithesis_instrumentation__.Notify(644208)
		return re == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(644209)
		return false
	} else {
		__antithesis_instrumentation__.Notify(644210)
	}
	__antithesis_instrumentation__.Notify(644203)
	errString := pgerror.FullError(err)
	matched, merr := regexp.MatchString(re, errString)
	if merr != nil {
		__antithesis_instrumentation__.Notify(644211)
		return false
	} else {
		__antithesis_instrumentation__.Notify(644212)
	}
	__antithesis_instrumentation__.Notify(644204)
	return matched
}

func IsPError(pErr *roachpb.Error, re string) bool {
	__antithesis_instrumentation__.Notify(644213)
	if pErr == nil && func() bool {
		__antithesis_instrumentation__.Notify(644217)
		return re == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(644218)
		return true
	} else {
		__antithesis_instrumentation__.Notify(644219)
	}
	__antithesis_instrumentation__.Notify(644214)
	if pErr == nil || func() bool {
		__antithesis_instrumentation__.Notify(644220)
		return re == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(644221)
		return false
	} else {
		__antithesis_instrumentation__.Notify(644222)
	}
	__antithesis_instrumentation__.Notify(644215)
	matched, merr := regexp.MatchString(re, pErr.String())
	if merr != nil {
		__antithesis_instrumentation__.Notify(644223)
		return false
	} else {
		__antithesis_instrumentation__.Notify(644224)
	}
	__antithesis_instrumentation__.Notify(644216)
	return matched
}
