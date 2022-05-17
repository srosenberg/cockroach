package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
)

func NewAmbiguousResultErrorf(format string, args ...interface{}) *AmbiguousResultError {
	__antithesis_instrumentation__.Notify(129011)
	return NewAmbiguousResultError(errors.NewWithDepthf(1, format, args...))
}

func NewAmbiguousResultError(err error) *AmbiguousResultError {
	__antithesis_instrumentation__.Notify(129012)
	return &AmbiguousResultError{
		EncodedError:      errors.EncodeError(context.Background(), err),
		DeprecatedMessage: err.Error(),
	}
}

var _ errors.SafeFormatter = (*AmbiguousResultError)(nil)
var _ fmt.Formatter = (*AmbiguousResultError)(nil)
var _ errors.Wrapper = func() errors.Wrapper {
	__antithesis_instrumentation__.Notify(129013)
	aErr := (*AmbiguousResultError)(nil)
	typeKey := errors.GetTypeKey(aErr)
	errors.RegisterWrapperEncoder(typeKey, func(ctx context.Context, err error) (msgPrefix string, safeDetails []string, payload proto.Message) {
		__antithesis_instrumentation__.Notify(129016)
		errors.As(err, &payload)
		return "", nil, payload
	})
	__antithesis_instrumentation__.Notify(129014)
	errors.RegisterWrapperDecoder(typeKey, func(ctx context.Context, cause error, msgPrefix string, safeDetails []string, payload proto.Message) error {
		__antithesis_instrumentation__.Notify(129017)
		return payload.(*AmbiguousResultError)
	})
	__antithesis_instrumentation__.Notify(129015)

	return aErr
}()

func (e *AmbiguousResultError) SafeFormatError(p errors.Printer) error {
	__antithesis_instrumentation__.Notify(129018)
	p.Printf("result is ambiguous: %s", e.unwrapOrDefault())
	return nil
}

func (e *AmbiguousResultError) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(129019)
	errors.FormatError(e, s, verb)
}

func (e *AmbiguousResultError) Error() string {
	__antithesis_instrumentation__.Notify(129020)
	return fmt.Sprint(e)
}

func (e *AmbiguousResultError) Unwrap() error {
	__antithesis_instrumentation__.Notify(129021)
	if e.EncodedError.Error == nil {
		__antithesis_instrumentation__.Notify(129023)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(129024)
	}
	__antithesis_instrumentation__.Notify(129022)
	return errors.DecodeError(context.Background(), e.EncodedError)
}

func (e *AmbiguousResultError) unwrapOrDefault() error {
	__antithesis_instrumentation__.Notify(129025)
	cause := e.Unwrap()
	if cause == nil {
		__antithesis_instrumentation__.Notify(129027)
		return errors.New("unknown cause")
	} else {
		__antithesis_instrumentation__.Notify(129028)
	}
	__antithesis_instrumentation__.Notify(129026)
	return cause
}

func (e *AmbiguousResultError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(129029)
	return fmt.Sprintf("result is ambiguous: %v", e.unwrapOrDefault())
}

func (e *AmbiguousResultError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(129030)
	return AmbiguousResultErrType
}
