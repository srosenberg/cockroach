package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uint128"
)

type ClusterWideID struct {
	uint128.Uint128
}

func GenerateClusterWideID(timestamp hlc.Timestamp, instID base.SQLInstanceID) ClusterWideID {
	__antithesis_instrumentation__.Notify(272609)
	loInt := (uint64)(instID)
	loInt = loInt | ((uint64)(timestamp.Logical) << 32)

	return ClusterWideID{Uint128: uint128.FromInts((uint64)(timestamp.WallTime), loInt)}
}

func StringToClusterWideID(s string) (ClusterWideID, error) {
	__antithesis_instrumentation__.Notify(272610)
	id, err := uint128.FromString(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(272612)
		return ClusterWideID{}, err
	} else {
		__antithesis_instrumentation__.Notify(272613)
	}
	__antithesis_instrumentation__.Notify(272611)
	return ClusterWideID{Uint128: id}, nil
}

func BytesToClusterWideID(b []byte) ClusterWideID {
	__antithesis_instrumentation__.Notify(272614)
	id := uint128.FromBytes(b)
	return ClusterWideID{Uint128: id}
}

func (id ClusterWideID) GetNodeID() int32 {
	__antithesis_instrumentation__.Notify(272615)
	return int32(0xFFFFFFFF & id.Lo)
}
