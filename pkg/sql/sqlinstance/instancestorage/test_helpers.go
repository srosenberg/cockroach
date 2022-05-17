// Package instancestorage provides a mock implementation
// of instance storage for testing purposes.
package instancestorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type FakeStorage struct {
	mu struct {
		syncutil.Mutex
		instances     map[base.SQLInstanceID]sqlinstance.InstanceInfo
		instanceIDCtr base.SQLInstanceID
		started       bool
	}
}

func NewFakeStorage() *FakeStorage {
	__antithesis_instrumentation__.Notify(624155)
	f := &FakeStorage{}
	f.mu.instances = make(map[base.SQLInstanceID]sqlinstance.InstanceInfo)
	f.mu.instanceIDCtr = base.SQLInstanceID(1)
	return f
}

func (f *FakeStorage) CreateInstance(
	_ context.Context, sessionID sqlliveness.SessionID, _ hlc.Timestamp, addr string,
) (base.SQLInstanceID, error) {
	__antithesis_instrumentation__.Notify(624156)
	f.mu.Lock()
	defer f.mu.Unlock()
	i := sqlinstance.InstanceInfo{
		InstanceID:   f.mu.instanceIDCtr,
		InstanceAddr: addr,
		SessionID:    sessionID,
	}
	f.mu.instances[f.mu.instanceIDCtr] = i
	f.mu.instanceIDCtr++
	return i.InstanceID, nil
}

func (f *FakeStorage) ReleaseInstanceID(_ context.Context, id base.SQLInstanceID) error {
	__antithesis_instrumentation__.Notify(624157)
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.mu.instances, id)
	return nil
}

func (s *Storage) GetInstanceDataForTest(
	ctx context.Context, instanceID base.SQLInstanceID,
) (sqlinstance.InstanceInfo, error) {
	__antithesis_instrumentation__.Notify(624158)
	i, err := s.getInstanceData(ctx, instanceID)
	if err != nil {
		__antithesis_instrumentation__.Notify(624160)
		return sqlinstance.InstanceInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(624161)
	}
	__antithesis_instrumentation__.Notify(624159)
	instanceInfo := sqlinstance.InstanceInfo{
		InstanceID:   i.instanceID,
		InstanceAddr: i.addr,
		SessionID:    i.sessionID,
	}
	return instanceInfo, nil
}

func (s *Storage) GetAllInstancesDataForTest(
	ctx context.Context,
) (instances []sqlinstance.InstanceInfo, _ error) {
	__antithesis_instrumentation__.Notify(624162)
	rows, err := s.getAllInstancesData(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(624165)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624166)
	}
	__antithesis_instrumentation__.Notify(624163)
	for _, instance := range rows {
		__antithesis_instrumentation__.Notify(624167)
		instanceInfo := sqlinstance.InstanceInfo{
			InstanceID:   instance.instanceID,
			InstanceAddr: instance.addr,
			SessionID:    instance.sessionID,
		}
		instances = append(instances, instanceInfo)
	}
	__antithesis_instrumentation__.Notify(624164)
	return instances, nil
}
