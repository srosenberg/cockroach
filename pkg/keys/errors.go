package keys

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type InvalidRangeMetaKeyError struct {
	Msg string
	Key roachpb.Key
}

func NewInvalidRangeMetaKeyError(msg string, k []byte) *InvalidRangeMetaKeyError {
	__antithesis_instrumentation__.Notify(85217)
	return &InvalidRangeMetaKeyError{Msg: msg, Key: k}
}

func (i *InvalidRangeMetaKeyError) Error() string {
	__antithesis_instrumentation__.Notify(85218)
	return fmt.Sprintf("%q is not valid range metadata key: %s", string(i.Key), i.Msg)
}
