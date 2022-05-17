package scrub

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

const (
	MissingIndexEntryError = "missing_index_entry"

	DanglingIndexReferenceError = "dangling_index_reference"

	PhysicalError = "physical_error"

	IndexKeyDecodingError = "index_key_decoding_error"

	IndexValueDecodingError = "index_value_decoding_error"

	SecondaryIndexKeyExtraValueDecodingError = "secondary_index_key_extra_value_decoding_error"

	UnexpectedNullValueError = "null_value_error"

	CheckConstraintViolation = "check_constraint_violation"

	ForeignKeyConstraintViolation = "foreign_key_violation"
)

type Error struct {
	Code       string
	underlying error
}

func (s *Error) Error() string {
	__antithesis_instrumentation__.Notify(595452)
	return fmt.Sprintf("%s: %+v", s.Code, s.underlying)
}

func (s *Error) Cause() error {
	__antithesis_instrumentation__.Notify(595453)
	return s.underlying
}

func (s *Error) Format(st fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(595454)
	errors.FormatError(s, st, verb)
}

func WrapError(code string, err error) *Error {
	__antithesis_instrumentation__.Notify(595455)
	return &Error{
		Code:       code,
		underlying: err,
	}
}

func IsScrubError(err error) bool {
	__antithesis_instrumentation__.Notify(595456)
	return errors.HasType(err, (*Error)(nil))
}

func UnwrapScrubError(err error) error {
	__antithesis_instrumentation__.Notify(595457)
	var e *Error
	if errors.As(err, &e) {
		__antithesis_instrumentation__.Notify(595459)
		return e.underlying
	} else {
		__antithesis_instrumentation__.Notify(595460)
	}
	__antithesis_instrumentation__.Notify(595458)
	return err
}
