package pgerror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
)

var _ error = (*Error)(nil)
var _ errors.ErrorHinter = (*Error)(nil)
var _ errors.ErrorDetailer = (*Error)(nil)
var _ fmt.Formatter = (*Error)(nil)

func (pg *Error) Error() string { __antithesis_instrumentation__.Notify(560283); return pg.Message }

func (pg *Error) ErrorHint() string { __antithesis_instrumentation__.Notify(560284); return pg.Hint }

func (pg *Error) ErrorDetail() string {
	__antithesis_instrumentation__.Notify(560285)
	return pg.Detail
}

func FullError(err error) string {
	__antithesis_instrumentation__.Notify(560286)
	var errString string
	if pqErr := (*pq.Error)(nil); errors.As(err, &pqErr) {
		__antithesis_instrumentation__.Notify(560288)
		errString = formatMsgHintDetail("pq", pqErr.Message, pqErr.Hint, pqErr.Detail)
	} else {
		__antithesis_instrumentation__.Notify(560289)
		pg := Flatten(err)
		errString = formatMsgHintDetail(pg.Severity, err.Error(), pg.Hint, pg.Detail)
	}
	__antithesis_instrumentation__.Notify(560287)
	return errString
}

func formatMsgHintDetail(prefix, msg, hint, detail string) string {
	__antithesis_instrumentation__.Notify(560290)
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteString(": ")
	b.WriteString(msg)
	if hint != "" {
		__antithesis_instrumentation__.Notify(560293)
		b.WriteString("\nHINT: ")
		b.WriteString(hint)
	} else {
		__antithesis_instrumentation__.Notify(560294)
	}
	__antithesis_instrumentation__.Notify(560291)
	if detail != "" {
		__antithesis_instrumentation__.Notify(560295)
		b.WriteString("\nDETAIL: ")
		b.WriteString(detail)
	} else {
		__antithesis_instrumentation__.Notify(560296)
	}
	__antithesis_instrumentation__.Notify(560292)
	return b.String()
}

func NewWithDepthf(depth int, code pgcode.Code, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(560297)
	err := errors.NewWithDepthf(1+depth, format, args...)
	err = WithCandidateCode(err, code)
	return err
}

func New(code pgcode.Code, msg string) error {
	__antithesis_instrumentation__.Notify(560298)
	err := errors.NewWithDepth(1, msg)
	err = WithCandidateCode(err, code)
	return err
}

func Newf(code pgcode.Code, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(560299)
	err := errors.NewWithDepthf(1, format, args...)
	err = WithCandidateCode(err, code)
	return err
}

func DangerousStatementf(format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(560300)
	err := errors.Newf(format, args...)
	err = errors.WithMessage(err, "rejected (sql_safe_updates = true)")
	err = WithCandidateCode(err, pgcode.Warning)
	return err
}

func WrongNumberOfPreparedStatements(n int) error {
	__antithesis_instrumentation__.Notify(560301)
	err := errors.NewWithDepthf(1, "prepared statement had %d statements, expected 1", errors.Safe(n))
	err = WithCandidateCode(err, pgcode.InvalidPreparedStatementDefinition)
	return err
}

var _ fmt.Formatter = &Error{}

func (pg *Error) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(560302)
	switch {
	case verb == 'v' && func() bool {
		__antithesis_instrumentation__.Notify(560310)
		return s.Flag('+') == true
	}() == true:
		__antithesis_instrumentation__.Notify(560303)

		if pg.Source != nil {
			__antithesis_instrumentation__.Notify(560311)
			fmt.Fprintf(s, "%s:%d in %s(): ", pg.Source.File, pg.Source.Line, pg.Source.Function)
		} else {
			__antithesis_instrumentation__.Notify(560312)
		}
		__antithesis_instrumentation__.Notify(560304)
		fmt.Fprintf(s, "(%s) %s", pg.Code, pg.Message)
		return
	case verb == 'v' && func() bool {
		__antithesis_instrumentation__.Notify(560313)
		return s.Flag('#') == true
	}() == true:
		__antithesis_instrumentation__.Notify(560305)

		fmt.Fprintf(s, "(%s) %s", pg.Code, pg.Message)
	case verb == 'v':
		__antithesis_instrumentation__.Notify(560306)
		fallthrough
	case verb == 's':
		__antithesis_instrumentation__.Notify(560307)
		fmt.Fprintf(s, "%s", pg.Message)
	case verb == 'q':
		__antithesis_instrumentation__.Notify(560308)
		fmt.Fprintf(s, "%q", pg.Message)
	default:
		__antithesis_instrumentation__.Notify(560309)
	}
}

var _ errors.SafeFormatter = (*Error)(nil)

func (pg *Error) SafeFormatError(s errors.Printer) (next error) {
	__antithesis_instrumentation__.Notify(560314)
	s.Print(pg.Message)
	if s.Detail() {
		__antithesis_instrumentation__.Notify(560316)
		if pg.Source != nil {
			__antithesis_instrumentation__.Notify(560318)
			s.Printf("Source: %s:%d in %s()",
				errors.Safe(pg.Source.File), errors.Safe(pg.Source.Line), errors.Safe(pg.Source.Function))
		} else {
			__antithesis_instrumentation__.Notify(560319)
		}
		__antithesis_instrumentation__.Notify(560317)
		s.Printf("SQLSTATE ", errors.Safe(pg.Code))
	} else {
		__antithesis_instrumentation__.Notify(560320)
	}
	__antithesis_instrumentation__.Notify(560315)
	return nil
}

func IsSQLRetryableError(err error) bool {
	__antithesis_instrumentation__.Notify(560321)

	errString := FullError(err)
	matched, merr := regexp.MatchString(
		"(no inbound stream connection|connection reset by peer|connection refused|failed to send RPC|rpc error: code = Unavailable|EOF|result is ambiguous)",
		errString)
	if merr != nil {
		__antithesis_instrumentation__.Notify(560323)
		return false
	} else {
		__antithesis_instrumentation__.Notify(560324)
	}
	__antithesis_instrumentation__.Notify(560322)
	return matched
}
