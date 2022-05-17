package poison

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

func NewPoisonedError(span roachpb.Span, ts hlc.Timestamp) *PoisonedError {
	__antithesis_instrumentation__.Notify(100821)
	return &PoisonedError{Span: span, Timestamp: ts}
}

var _ errors.SafeFormatter = (*PoisonedError)(nil)
var _ fmt.Formatter = (*PoisonedError)(nil)

func (e *PoisonedError) SafeFormatError(p errors.Printer) error {
	__antithesis_instrumentation__.Notify(100822)
	p.Printf("encountered poisoned latch %s@%s", e.Span, e.Timestamp)
	return nil
}

func (e *PoisonedError) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(100823)
	errors.FormatError(e, s, verb)
}

func (e *PoisonedError) Error() string {
	__antithesis_instrumentation__.Notify(100824)
	return fmt.Sprint(e)
}
