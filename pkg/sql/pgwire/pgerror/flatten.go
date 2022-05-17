package pgerror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/errors"
)

func Flatten(err error) *Error {
	__antithesis_instrumentation__.Notify(560782)
	if err == nil {
		__antithesis_instrumentation__.Notify(560787)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(560788)
	}
	__antithesis_instrumentation__.Notify(560783)
	resErr := &Error{
		Code:           GetPGCode(err).String(),
		Message:        err.Error(),
		Severity:       GetSeverity(err),
		ConstraintName: GetConstraintName(err),
	}

	if file, line, fn, ok := errors.GetOneLineSource(err); ok {
		__antithesis_instrumentation__.Notify(560789)
		resErr.Source = &Error_Source{File: file, Line: int32(line), Function: fn}
	} else {
		__antithesis_instrumentation__.Notify(560790)
	}
	__antithesis_instrumentation__.Notify(560784)

	if resErr.Code == pgcode.SerializationFailure.String() {
		__antithesis_instrumentation__.Notify(560791)
		err = withSerializationFailureHints(err)
	} else {
		__antithesis_instrumentation__.Notify(560792)
	}
	__antithesis_instrumentation__.Notify(560785)

	resErr.Hint = errors.FlattenHints(err)
	resErr.Detail = errors.FlattenDetails(err)

	switch resErr.Code {
	case pgcode.Internal.String():
		__antithesis_instrumentation__.Notify(560793)

		if !strings.HasPrefix(resErr.Message, InternalErrorPrefix) {
			__antithesis_instrumentation__.Notify(560797)
			resErr.Message = InternalErrorPrefix + ": " + resErr.Message
		} else {
			__antithesis_instrumentation__.Notify(560798)
		}
		__antithesis_instrumentation__.Notify(560794)

		resErr.Detail += getInnerMostStackTraceAsDetail(err)

	case pgcode.SerializationFailure.String():
		__antithesis_instrumentation__.Notify(560795)

		if !strings.HasPrefix(resErr.Message, TxnRetryMsgPrefix) {
			__antithesis_instrumentation__.Notify(560799)
			resErr.Message = TxnRetryMsgPrefix + ": " + resErr.Message
		} else {
			__antithesis_instrumentation__.Notify(560800)
		}
	default:
		__antithesis_instrumentation__.Notify(560796)
	}
	__antithesis_instrumentation__.Notify(560786)

	return resErr
}

var serializationFailureReasonRegexp = regexp.MustCompile(
	`((?:ABORT_|RETRY_)[A-Z_]*|ReadWithinUncertaintyInterval)`,
)

func withSerializationFailureHints(err error) error {
	__antithesis_instrumentation__.Notify(560801)
	url := docs.URL("transaction-retry-error-reference.html")
	if match := serializationFailureReasonRegexp.FindStringSubmatch(err.Error()); len(match) >= 2 {
		__antithesis_instrumentation__.Notify(560803)
		url += "#" + strings.ToLower(match[1])
	} else {
		__antithesis_instrumentation__.Notify(560804)
	}
	__antithesis_instrumentation__.Notify(560802)
	return errors.WithIssueLink(err, errors.IssueLink{IssueURL: url})
}

func getInnerMostStackTraceAsDetail(err error) string {
	__antithesis_instrumentation__.Notify(560805)
	if c := errors.UnwrapOnce(err); c != nil {
		__antithesis_instrumentation__.Notify(560808)
		s := getInnerMostStackTraceAsDetail(c)
		if s != "" {
			__antithesis_instrumentation__.Notify(560809)
			return s
		} else {
			__antithesis_instrumentation__.Notify(560810)
		}
	} else {
		__antithesis_instrumentation__.Notify(560811)
	}
	__antithesis_instrumentation__.Notify(560806)

	if st := errors.GetReportableStackTrace(err); st != nil {
		__antithesis_instrumentation__.Notify(560812)
		var t bytes.Buffer
		t.WriteString("stack trace:\n")
		for i := len(st.Frames) - 1; i >= 0; i-- {
			__antithesis_instrumentation__.Notify(560814)
			f := st.Frames[i]
			fmt.Fprintf(&t, "%s:%d: %s()\n", f.Filename, f.Lineno, f.Function)
		}
		__antithesis_instrumentation__.Notify(560813)
		return t.String()
	} else {
		__antithesis_instrumentation__.Notify(560815)
	}
	__antithesis_instrumentation__.Notify(560807)
	return ""
}

const InternalErrorPrefix = "internal error"

const TxnRetryMsgPrefix = "restart transaction"

func GetPGCode(err error) pgcode.Code {
	__antithesis_instrumentation__.Notify(560816)
	return GetPGCodeInternal(err, ComputeDefaultCode)
}
