package pgerror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
)

type withCandidateCode struct {
	cause error
	code  string
}

var _ error = (*withCandidateCode)(nil)
var _ errors.SafeDetailer = (*withCandidateCode)(nil)
var _ fmt.Formatter = (*withCandidateCode)(nil)
var _ errors.SafeFormatter = (*withCandidateCode)(nil)

func (w *withCandidateCode) Error() string {
	__antithesis_instrumentation__.Notify(560874)
	return w.cause.Error()
}
func (w *withCandidateCode) Cause() error {
	__antithesis_instrumentation__.Notify(560875)
	return w.cause
}
func (w *withCandidateCode) Unwrap() error {
	__antithesis_instrumentation__.Notify(560876)
	return w.cause
}
func (w *withCandidateCode) SafeDetails() []string {
	__antithesis_instrumentation__.Notify(560877)
	return []string{w.code}
}

func (w *withCandidateCode) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(560878)
	errors.FormatError(w, s, verb)
}

func (w *withCandidateCode) SafeFormatError(p errors.Printer) (next error) {
	__antithesis_instrumentation__.Notify(560879)
	if p.Detail() {
		__antithesis_instrumentation__.Notify(560881)
		p.Printf("candidate pg code: %s", errors.Safe(w.code))
	} else {
		__antithesis_instrumentation__.Notify(560882)
	}
	__antithesis_instrumentation__.Notify(560880)
	return w.cause
}

func decodeWithCandidateCode(
	_ context.Context, cause error, _ string, details []string, _ proto.Message,
) error {
	__antithesis_instrumentation__.Notify(560883)
	code := pgcode.Uncategorized.String()
	if len(details) > 0 {
		__antithesis_instrumentation__.Notify(560885)
		code = details[0]
	} else {
		__antithesis_instrumentation__.Notify(560886)
	}
	__antithesis_instrumentation__.Notify(560884)
	return &withCandidateCode{cause: cause, code: code}
}

func init() {
	errors.RegisterWrapperDecoder(errors.GetTypeKey((*withCandidateCode)(nil)), decodeWithCandidateCode)
}
