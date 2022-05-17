package clierror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
)

func OutputError(w io.Writer, err error, showSeverity, verbose bool) {
	__antithesis_instrumentation__.Notify(28293)
	f := formattedError{err: err, showSeverity: showSeverity, verbose: verbose}
	fmt.Fprintln(w, f.Error())
}

func NewFormattedError(err error, showSeverity, verbose bool) error {
	__antithesis_instrumentation__.Notify(28294)

	if f := (*formattedError)(nil); errors.As(err, &f) {
		__antithesis_instrumentation__.Notify(28296)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28297)
	}
	__antithesis_instrumentation__.Notify(28295)
	return &formattedError{err: err, showSeverity: showSeverity, verbose: verbose}
}

type formattedError struct {
	err                   error
	showSeverity, verbose bool
}

func (f *formattedError) Unwrap() error {
	__antithesis_instrumentation__.Notify(28298)
	return f.err
}

func (f *formattedError) Error() string {
	__antithesis_instrumentation__.Notify(28299)

	var other *formattedError
	if errors.As(f.err, &other) {
		__antithesis_instrumentation__.Notify(28308)
		return other.Error()
	} else {
		__antithesis_instrumentation__.Notify(28309)
	}
	__antithesis_instrumentation__.Notify(28300)
	var buf strings.Builder

	severity := "ERROR"

	var message, hint, detail, location, constraintName string
	var code pgcode.Code
	if pqErr := (*pq.Error)(nil); errors.As(f.err, &pqErr) {
		__antithesis_instrumentation__.Notify(28310)
		if pqErr.Severity != "" {
			__antithesis_instrumentation__.Notify(28312)
			severity = pqErr.Severity
		} else {
			__antithesis_instrumentation__.Notify(28313)
		}
		__antithesis_instrumentation__.Notify(28311)
		constraintName = pqErr.Constraint
		message = pqErr.Message
		code = pgcode.MakeCode(string(pqErr.Code))
		hint, detail = pqErr.Hint, pqErr.Detail
		location = formatLocation(pqErr.File, pqErr.Line, pqErr.Routine)
	} else {
		__antithesis_instrumentation__.Notify(28314)
		message = f.err.Error()
		code = pgerror.GetPGCode(f.err)

		hint = errors.FlattenHints(f.err)
		detail = errors.FlattenDetails(f.err)
		if file, line, fn, ok := errors.GetOneLineSource(f.err); ok {
			__antithesis_instrumentation__.Notify(28315)
			location = formatLocation(file, strconv.FormatInt(int64(line), 10), fn)
		} else {
			__antithesis_instrumentation__.Notify(28316)
		}
	}
	__antithesis_instrumentation__.Notify(28301)

	if f.showSeverity && func() bool {
		__antithesis_instrumentation__.Notify(28317)
		return severity != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(28318)
		fmt.Fprintf(&buf, "%s: ", severity)
	} else {
		__antithesis_instrumentation__.Notify(28319)
	}
	__antithesis_instrumentation__.Notify(28302)
	fmt.Fprintln(&buf, message)

	if severity != "NOTICE" && func() bool {
		__antithesis_instrumentation__.Notify(28320)
		return code.String() != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(28321)

		if code == pgcode.Uncategorized && func() bool {
			__antithesis_instrumentation__.Notify(28322)
			return !f.verbose == true
		}() == true {
			__antithesis_instrumentation__.Notify(28323)

		} else {
			__antithesis_instrumentation__.Notify(28324)
			fmt.Fprintln(&buf, "SQLSTATE:", code)
		}
	} else {
		__antithesis_instrumentation__.Notify(28325)
	}
	__antithesis_instrumentation__.Notify(28303)

	if detail != "" {
		__antithesis_instrumentation__.Notify(28326)
		fmt.Fprintln(&buf, "DETAIL:", detail)
	} else {
		__antithesis_instrumentation__.Notify(28327)
	}
	__antithesis_instrumentation__.Notify(28304)
	if constraintName != "" {
		__antithesis_instrumentation__.Notify(28328)
		fmt.Fprintln(&buf, "CONSTRAINT:", constraintName)
	} else {
		__antithesis_instrumentation__.Notify(28329)
	}
	__antithesis_instrumentation__.Notify(28305)
	if hint != "" {
		__antithesis_instrumentation__.Notify(28330)
		fmt.Fprintln(&buf, "HINT:", hint)
	} else {
		__antithesis_instrumentation__.Notify(28331)
	}
	__antithesis_instrumentation__.Notify(28306)
	if f.verbose && func() bool {
		__antithesis_instrumentation__.Notify(28332)
		return location != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(28333)
		fmt.Fprintln(&buf, "LOCATION:", location)
	} else {
		__antithesis_instrumentation__.Notify(28334)
	}
	__antithesis_instrumentation__.Notify(28307)

	return strings.TrimRight(buf.String(), "\n")
}

func formatLocation(file, line, fn string) string {
	__antithesis_instrumentation__.Notify(28335)
	var res strings.Builder
	res.WriteString(fn)
	if file != "" || func() bool {
		__antithesis_instrumentation__.Notify(28337)
		return line != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(28338)
		if fn != "" {
			__antithesis_instrumentation__.Notify(28341)
			res.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(28342)
		}
		__antithesis_instrumentation__.Notify(28339)
		if file == "" {
			__antithesis_instrumentation__.Notify(28343)
			res.WriteString("<unknown>")
		} else {
			__antithesis_instrumentation__.Notify(28344)
			res.WriteString(file)
		}
		__antithesis_instrumentation__.Notify(28340)
		if line != "" {
			__antithesis_instrumentation__.Notify(28345)
			res.WriteByte(':')
			res.WriteString(line)
		} else {
			__antithesis_instrumentation__.Notify(28346)
		}
	} else {
		__antithesis_instrumentation__.Notify(28347)
	}
	__antithesis_instrumentation__.Notify(28336)
	return res.String()
}
