package spanconfig

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
)

type mismatchedDescriptorTypesError struct {
	d1, d2 catalog.DescriptorType
}

func NewMismatchedDescriptorTypesError(d1, d2 catalog.DescriptorType) error {
	__antithesis_instrumentation__.Notify(240249)
	return mismatchedDescriptorTypesError{d1, d2}
}

func IsMismatchedDescriptorTypesError(err error) bool {
	__antithesis_instrumentation__.Notify(240250)
	return errors.HasType(err, mismatchedDescriptorTypesError{})
}

func (e mismatchedDescriptorTypesError) Error() string {
	__antithesis_instrumentation__.Notify(240251)
	return fmt.Sprintf("mismatched descriptor types (%s, %s) for the same id", e.d1, e.d2)
}

type commitTimestampOutOfBoundsError struct{}

func NewCommitTimestampOutOfBoundsError() error {
	__antithesis_instrumentation__.Notify(240252)
	return commitTimestampOutOfBoundsError{}
}

func IsCommitTimestampOutOfBoundsError(err error) bool {
	__antithesis_instrumentation__.Notify(240253)
	return errors.Is(err, commitTimestampOutOfBoundsError{})
}

func (e commitTimestampOutOfBoundsError) Error() string {
	__antithesis_instrumentation__.Notify(240254)
	return "lease expired"
}

func decodeRetryableLeaseExpiredError(context.Context, string, []string, proto.Message) error {
	__antithesis_instrumentation__.Notify(240255)
	return NewCommitTimestampOutOfBoundsError()
}

func init() {
	errors.RegisterLeafDecoder(
		errors.GetTypeKey((*commitTimestampOutOfBoundsError)(nil)), decodeRetryableLeaseExpiredError,
	)
}
