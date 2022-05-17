package pgerror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/errors"
)

func Wrapf(err error, code pgcode.Code, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(560887)
	return WrapWithDepthf(1, err, code, format, args...)
}

func WrapWithDepthf(
	depth int, err error, code pgcode.Code, format string, args ...interface{},
) error {
	__antithesis_instrumentation__.Notify(560888)
	err = errors.WrapWithDepthf(1+depth, err, format, args...)
	err = WithCandidateCode(err, code)
	return err
}

func Wrap(err error, code pgcode.Code, msg string) error {
	__antithesis_instrumentation__.Notify(560889)
	if msg == "" {
		__antithesis_instrumentation__.Notify(560891)
		return WithCandidateCode(err, code)
	} else {
		__antithesis_instrumentation__.Notify(560892)
	}
	__antithesis_instrumentation__.Notify(560890)
	return WrapWithDepthf(1, err, code, "%s", msg)
}
