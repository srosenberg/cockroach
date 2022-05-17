package pgerror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errorspb"
	"github.com/gogo/protobuf/proto"
)

func WithConstraintName(err error, constraint string) error {
	__antithesis_instrumentation__.Notify(560261)
	if err == nil {
		__antithesis_instrumentation__.Notify(560263)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(560264)
	}
	__antithesis_instrumentation__.Notify(560262)

	return &withConstraintName{cause: err, constraint: constraint}
}

func GetConstraintName(err error) string {
	__antithesis_instrumentation__.Notify(560265)
	if c := (*withConstraintName)(nil); errors.As(err, &c) {
		__antithesis_instrumentation__.Notify(560267)
		return c.constraint
	} else {
		__antithesis_instrumentation__.Notify(560268)
	}
	__antithesis_instrumentation__.Notify(560266)
	return ""
}

type withConstraintName struct {
	cause      error
	constraint string
}

var _ error = (*withConstraintName)(nil)
var _ errors.SafeDetailer = (*withConstraintName)(nil)
var _ fmt.Formatter = (*withConstraintName)(nil)
var _ errors.SafeFormatter = (*withConstraintName)(nil)

func (w *withConstraintName) Error() string {
	__antithesis_instrumentation__.Notify(560269)
	return w.cause.Error()
}
func (w *withConstraintName) Cause() error {
	__antithesis_instrumentation__.Notify(560270)
	return w.cause
}
func (w *withConstraintName) Unwrap() error {
	__antithesis_instrumentation__.Notify(560271)
	return w.cause
}
func (w *withConstraintName) SafeDetails() []string {
	__antithesis_instrumentation__.Notify(560272)

	return nil
}

func (w *withConstraintName) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(560273)
	errors.FormatError(w, s, verb)
}

func (w *withConstraintName) SafeFormatError(p errors.Printer) (next error) {
	__antithesis_instrumentation__.Notify(560274)
	if p.Detail() {
		__antithesis_instrumentation__.Notify(560276)
		p.Printf("constraint name: %s", w.constraint)
	} else {
		__antithesis_instrumentation__.Notify(560277)
	}
	__antithesis_instrumentation__.Notify(560275)
	return w.cause
}

func encodeWithConstraintName(_ context.Context, err error) (string, []string, proto.Message) {
	__antithesis_instrumentation__.Notify(560278)
	w := err.(*withConstraintName)
	return "", nil, &errorspb.StringPayload{Msg: w.constraint}
}

func decodeWithConstraintName(
	_ context.Context, cause error, _ string, _ []string, payload proto.Message,
) error {
	__antithesis_instrumentation__.Notify(560279)
	m, ok := payload.(*errorspb.StringPayload)
	if !ok {
		__antithesis_instrumentation__.Notify(560281)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(560282)
	}
	__antithesis_instrumentation__.Notify(560280)
	return &withConstraintName{cause: cause, constraint: m.Msg}
}

func init() {
	key := errors.GetTypeKey((*withConstraintName)(nil))
	errors.RegisterWrapperEncoder(key, encodeWithConstraintName)
	errors.RegisterWrapperDecoder(key, decodeWithConstraintName)
}
