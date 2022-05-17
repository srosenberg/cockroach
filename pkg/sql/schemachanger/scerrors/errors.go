package scerrors

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type notImplementedError struct {
	n      tree.NodeFormatter
	detail string
}

var _ error = (*notImplementedError)(nil)

func HasNotImplemented(err error) bool {
	__antithesis_instrumentation__.Notify(581303)
	return errors.HasType(err, (*notImplementedError)(nil))
}

func (e *notImplementedError) Error() string {
	__antithesis_instrumentation__.Notify(581304)
	var buf strings.Builder
	fmt.Fprintf(&buf, "%T not implemented in the new schema changer", e.n)
	if e.detail != "" {
		__antithesis_instrumentation__.Notify(581306)
		fmt.Fprintf(&buf, ": %s", e.detail)
	} else {
		__antithesis_instrumentation__.Notify(581307)
	}
	__antithesis_instrumentation__.Notify(581305)
	return buf.String()
}

func NotImplementedError(n tree.NodeFormatter) error {
	__antithesis_instrumentation__.Notify(581308)
	return &notImplementedError{n: n}
}

func NotImplementedErrorf(n tree.NodeFormatter, fmtstr string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(581309)
	return &notImplementedError{n: n, detail: fmt.Sprintf(fmtstr, args...)}
}

type concurrentSchemaChangeError struct {
	descID descpb.ID
}

func (e *concurrentSchemaChangeError) ClientVisibleRetryError() {
	__antithesis_instrumentation__.Notify(581310)
}

func (e *concurrentSchemaChangeError) Error() string {
	__antithesis_instrumentation__.Notify(581311)
	return fmt.Sprintf("descriptor %d is undergoing another schema change", e.descID)
}

func ConcurrentSchemaChangeDescID(err error) descpb.ID {
	__antithesis_instrumentation__.Notify(581312)
	cscErr := (*concurrentSchemaChangeError)(nil)
	if !errors.As(err, &cscErr) {
		__antithesis_instrumentation__.Notify(581314)
		return descpb.InvalidID
	} else {
		__antithesis_instrumentation__.Notify(581315)
	}
	__antithesis_instrumentation__.Notify(581313)
	return cscErr.descID
}

func ConcurrentSchemaChangeError(desc catalog.Descriptor) error {
	__antithesis_instrumentation__.Notify(581316)
	return &concurrentSchemaChangeError{descID: desc.GetID()}
}
