package lease

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

type StorageTestingKnobs struct {
	LeaseReleasedEvent func(id descpb.ID, version descpb.DescriptorVersion, err error)

	LeaseAcquiredEvent func(desc catalog.Descriptor, err error)

	LeaseAcquireResultBlockEvent func(leaseBlockType AcquireBlockType, id descpb.ID)

	RemoveOnceDereferenced bool
}

func (*StorageTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(266681) }

var _ base.ModuleTestingKnobs = &StorageTestingKnobs{}

type ManagerTestingKnobs struct {
	TestingDescriptorRefreshedEvent func(descriptor *descpb.Descriptor)

	TestingDescriptorUpdateEvent func(descriptor *descpb.Descriptor) error

	DisableDeleteOrphanedLeases bool

	VersionPollIntervalForRangefeeds time.Duration

	LeaseStoreTestingKnobs StorageTestingKnobs
}

var _ base.ModuleTestingKnobs = &ManagerTestingKnobs{}

func (*ManagerTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(266682) }

func (m *Manager) TestingAcquireAndAssertMinVersion(
	ctx context.Context, timestamp hlc.Timestamp, id descpb.ID, minVersion descpb.DescriptorVersion,
) (LeasedDescriptor, error) {
	__antithesis_instrumentation__.Notify(266683)
	t := m.findDescriptorState(id, true)
	if err := ensureVersion(ctx, id, minVersion, m); err != nil {
		__antithesis_instrumentation__.Notify(266686)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(266687)
	}
	__antithesis_instrumentation__.Notify(266684)
	desc, _, err := t.findForTimestamp(ctx, timestamp)
	if err != nil {
		__antithesis_instrumentation__.Notify(266688)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(266689)
	}
	__antithesis_instrumentation__.Notify(266685)
	return desc, nil
}

func (m *Manager) TestingOutstandingLeasesGauge() *metric.Gauge {
	__antithesis_instrumentation__.Notify(266690)
	return m.storage.outstandingLeases
}
