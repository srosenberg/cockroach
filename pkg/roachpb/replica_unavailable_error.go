package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	context "context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
)

func NewReplicaUnavailableError(
	cause error, desc *RangeDescriptor, replDesc ReplicaDescriptor,
) error {
	__antithesis_instrumentation__.Notify(176877)
	return &ReplicaUnavailableError{
		Desc:    *desc,
		Replica: replDesc,
		Cause:   errors.EncodeError(context.Background(), cause),
	}
}

var _ errors.SafeFormatter = (*ReplicaUnavailableError)(nil)
var _ fmt.Formatter = (*ReplicaUnavailableError)(nil)
var _ errors.Wrapper = (*ReplicaUnavailableError)(nil)

func (e *ReplicaUnavailableError) SafeFormatError(p errors.Printer) error {
	__antithesis_instrumentation__.Notify(176878)
	p.Printf("replica unavailable: %s unable to serve request to %s: %s", e.Replica, e.Desc, e.Unwrap())
	return nil
}

func (e *ReplicaUnavailableError) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(176879)
	errors.FormatError(e, s, verb)
}

func (e *ReplicaUnavailableError) Error() string {
	__antithesis_instrumentation__.Notify(176880)
	return fmt.Sprint(e)
}

func (e *ReplicaUnavailableError) Unwrap() error {
	__antithesis_instrumentation__.Notify(176881)
	return errors.DecodeError(context.Background(), e.Cause)
}

func init() {
	encode := func(ctx context.Context, err error) (msgPrefix string, safeDetails []string, payload proto.Message) {
		errors.As(err, &payload)
		return "", nil, payload
	}
	decode := func(ctx context.Context, cause error, msgPrefix string, safeDetails []string, payload proto.Message) error {
		return payload.(*ReplicaUnavailableError)
	}
	typeName := errors.GetTypeKey((*ReplicaUnavailableError)(nil))
	errors.RegisterWrapperEncoder(typeName, encode)
	errors.RegisterWrapperDecoder(typeName, decode)
}
