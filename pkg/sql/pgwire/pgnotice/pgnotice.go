package pgnotice

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
)

type Notice error

func Newf(format string, args ...interface{}) Notice {
	__antithesis_instrumentation__.Notify(560901)
	err := errors.NewWithDepthf(1, format, args...)
	err = pgerror.WithCandidateCode(err, pgcode.SuccessfulCompletion)
	err = pgerror.WithSeverity(err, "NOTICE")
	return Notice(err)
}

func NewWithSeverityf(severity string, format string, args ...interface{}) Notice {
	__antithesis_instrumentation__.Notify(560902)
	err := errors.NewWithDepthf(1, format, args...)
	err = pgerror.WithCandidateCode(err, pgcode.SuccessfulCompletion)
	err = pgerror.WithSeverity(err, severity)
	return Notice(err)
}
