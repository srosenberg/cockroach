package loqrecoverypb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/keysutil"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
)

type RecoveryKey roachpb.RKey

func (r RecoveryKey) MarshalYAML() (interface{}, error) {
	__antithesis_instrumentation__.Notify(108874)
	return roachpb.RKey(r).String(), nil
}

func (r *RecoveryKey) UnmarshalYAML(fn func(interface{}) error) error {
	__antithesis_instrumentation__.Notify(108875)
	var pretty string
	if err := fn(&pretty); err != nil {
		__antithesis_instrumentation__.Notify(108878)
		return err
	} else {
		__antithesis_instrumentation__.Notify(108879)
	}
	__antithesis_instrumentation__.Notify(108876)
	scanner := keysutil.MakePrettyScanner(nil)
	key, err := scanner.Scan(pretty)
	if err != nil {
		__antithesis_instrumentation__.Notify(108880)
		return errors.Wrapf(err, "failed to parse key %s", pretty)
	} else {
		__antithesis_instrumentation__.Notify(108881)
	}
	__antithesis_instrumentation__.Notify(108877)
	*r = RecoveryKey(key)
	return nil
}

func (r RecoveryKey) AsRKey() roachpb.RKey {
	__antithesis_instrumentation__.Notify(108882)
	return roachpb.RKey(r)
}

func (m ReplicaUpdate) String() string {
	__antithesis_instrumentation__.Notify(108883)
	return proto.CompactTextString(&m)
}

func (m ReplicaUpdate) NodeID() roachpb.NodeID {
	__antithesis_instrumentation__.Notify(108884)
	return m.NewReplica.NodeID
}

func (m ReplicaUpdate) StoreID() roachpb.StoreID {
	__antithesis_instrumentation__.Notify(108885)
	return m.NewReplica.StoreID
}

func (m *ReplicaInfo) Replica() (roachpb.ReplicaDescriptor, error) {
	__antithesis_instrumentation__.Notify(108886)
	if d, ok := m.Desc.GetReplicaDescriptor(m.StoreID); ok {
		__antithesis_instrumentation__.Notify(108888)
		return d, nil
	} else {
		__antithesis_instrumentation__.Notify(108889)
	}
	__antithesis_instrumentation__.Notify(108887)
	return roachpb.ReplicaDescriptor{}, errors.Errorf(
		"invalid replica info: its own store s%d is not present in descriptor replicas %s",
		m.StoreID, m.Desc)
}

func (m *ReplicaRecoveryRecord) AsStructuredLog() eventpb.DebugRecoverReplica {
	__antithesis_instrumentation__.Notify(108890)
	return eventpb.DebugRecoverReplica{
		CommonEventDetails: eventpb.CommonEventDetails{
			Timestamp: m.Timestamp,
		},
		CommonDebugEventDetails: eventpb.CommonDebugEventDetails{
			NodeID: int32(m.NewReplica.NodeID),
		},
		RangeID:           int64(m.RangeID),
		StoreID:           int64(m.NewReplica.StoreID),
		SurvivorReplicaID: int32(m.OldReplicaID),
		UpdatedReplicaID:  int32(m.NewReplica.ReplicaID),
		StartKey:          m.StartKey.AsRKey().String(),
		EndKey:            m.EndKey.AsRKey().String(),
	}
}
