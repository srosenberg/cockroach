package types

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"

	"github.com/gogo/protobuf/jsonpb"
)

func (t *T) MarshalJSONPB(marshaler *jsonpb.Marshaler) ([]byte, error) {
	__antithesis_instrumentation__.Notify(631204)
	temp := *t
	if err := temp.downgradeType(); err != nil {
		__antithesis_instrumentation__.Notify(631207)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(631208)
	}
	__antithesis_instrumentation__.Notify(631205)
	var buf bytes.Buffer
	if err := marshaler.Marshal(&buf, &temp.InternalType); err != nil {
		__antithesis_instrumentation__.Notify(631209)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(631210)
	}
	__antithesis_instrumentation__.Notify(631206)
	return buf.Bytes(), nil
}

func (t *T) UnmarshalJSONPB(unmarshaler *jsonpb.Unmarshaler, data []byte) error {
	__antithesis_instrumentation__.Notify(631211)
	if err := unmarshaler.Unmarshal(bytes.NewReader(data), &t.InternalType); err != nil {
		__antithesis_instrumentation__.Notify(631213)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631214)
	}
	__antithesis_instrumentation__.Notify(631212)
	return t.upgradeType()
}
