package pgerror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
)

const DefaultSeverity = "ERROR"

func WithSeverity(err error, severity string) error {
	__antithesis_instrumentation__.Notify(560853)
	if err == nil {
		__antithesis_instrumentation__.Notify(560855)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(560856)
	}
	__antithesis_instrumentation__.Notify(560854)

	return &withSeverity{cause: err, severity: severity}
}

func GetSeverity(err error) string {
	__antithesis_instrumentation__.Notify(560857)
	if c := (*withSeverity)(nil); errors.As(err, &c) {
		__antithesis_instrumentation__.Notify(560859)
		return c.severity
	} else {
		__antithesis_instrumentation__.Notify(560860)
	}
	__antithesis_instrumentation__.Notify(560858)
	return DefaultSeverity
}

type withSeverity struct {
	cause    error
	severity string
}

var _ error = (*withSeverity)(nil)
var _ errors.SafeDetailer = (*withSeverity)(nil)
var _ fmt.Formatter = (*withSeverity)(nil)
var _ errors.Formatter = (*withSeverity)(nil)

func (w *withSeverity) Error() string {
	__antithesis_instrumentation__.Notify(560861)
	return w.cause.Error()
}
func (w *withSeverity) Cause() error  { __antithesis_instrumentation__.Notify(560862); return w.cause }
func (w *withSeverity) Unwrap() error { __antithesis_instrumentation__.Notify(560863); return w.cause }
func (w *withSeverity) SafeDetails() []string {
	__antithesis_instrumentation__.Notify(560864)
	return []string{w.severity}
}

func (w *withSeverity) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(560865)
	errors.FormatError(w, s, verb)
}

func (w *withSeverity) FormatError(p errors.Printer) (next error) {
	__antithesis_instrumentation__.Notify(560866)
	if p.Detail() {
		__antithesis_instrumentation__.Notify(560868)
		p.Printf("severity: %s", w.severity)
	} else {
		__antithesis_instrumentation__.Notify(560869)
	}
	__antithesis_instrumentation__.Notify(560867)
	return w.cause
}

func decodeWithSeverity(
	_ context.Context, cause error, _ string, details []string, _ proto.Message,
) error {
	__antithesis_instrumentation__.Notify(560870)
	severity := DefaultSeverity
	if len(details) > 0 {
		__antithesis_instrumentation__.Notify(560872)
		severity = details[0]
	} else {
		__antithesis_instrumentation__.Notify(560873)
	}
	__antithesis_instrumentation__.Notify(560871)
	return &withSeverity{cause: cause, severity: severity}
}

func init() {
	errors.RegisterWrapperDecoder(errors.GetTypeKey((*withSeverity)(nil)), decodeWithSeverity)
}
