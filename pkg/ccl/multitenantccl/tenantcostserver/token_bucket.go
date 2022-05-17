package tenantcostserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func (s *instance) TokenBucketRequest(
	ctx context.Context, tenantID roachpb.TenantID, in *roachpb.TokenBucketRequest,
) *roachpb.TokenBucketResponse {
	__antithesis_instrumentation__.Notify(20383)
	if tenantID == roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(20388)
		return &roachpb.TokenBucketResponse{
			Error: errors.EncodeError(ctx, errors.New("token bucket request for system tenant")),
		}
	} else {
		__antithesis_instrumentation__.Notify(20389)
	}
	__antithesis_instrumentation__.Notify(20384)
	instanceID := base.SQLInstanceID(in.InstanceID)
	if instanceID < 1 {
		__antithesis_instrumentation__.Notify(20390)
		return &roachpb.TokenBucketResponse{
			Error: errors.EncodeError(ctx, errors.Errorf("invalid instance ID %d", instanceID)),
		}
	} else {
		__antithesis_instrumentation__.Notify(20391)
	}
	__antithesis_instrumentation__.Notify(20385)
	if in.RequestedRU < 0 {
		__antithesis_instrumentation__.Notify(20392)
		return &roachpb.TokenBucketResponse{
			Error: errors.EncodeError(ctx, errors.Errorf("negative requested RUs")),
		}
	} else {
		__antithesis_instrumentation__.Notify(20393)
	}
	__antithesis_instrumentation__.Notify(20386)

	metrics := s.metrics.getTenantMetrics(tenantID)

	metrics.mutex.Lock()
	defer metrics.mutex.Unlock()

	result := &roachpb.TokenBucketResponse{}
	var consumption roachpb.TenantConsumption
	if err := s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(20394)
		*result = roachpb.TokenBucketResponse{}

		h := makeSysTableHelper(ctx, s.executor, txn, tenantID)
		tenant, instance, err := h.readTenantAndInstanceState(instanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(20403)
			return err
		} else {
			__antithesis_instrumentation__.Notify(20404)
		}
		__antithesis_instrumentation__.Notify(20395)

		if !tenant.Present {
			__antithesis_instrumentation__.Notify(20405)

			if err := s.checkTenantID(ctx, txn, tenantID); err != nil {
				__antithesis_instrumentation__.Notify(20406)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20407)
			}
		} else {
			__antithesis_instrumentation__.Notify(20408)
		}
		__antithesis_instrumentation__.Notify(20396)
		now := s.timeSource.Now()
		tenant.update(now)

		if !instance.Present {
			__antithesis_instrumentation__.Notify(20409)
			if err := h.accomodateNewInstance(&tenant, &instance); err != nil {
				__antithesis_instrumentation__.Notify(20410)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20411)
			}
		} else {
			__antithesis_instrumentation__.Notify(20412)
		}
		__antithesis_instrumentation__.Notify(20397)
		if string(instance.Lease) != string(in.InstanceLease) {
			__antithesis_instrumentation__.Notify(20413)

			instance.Seq = 0
			instance.Lease = tree.DBytes(in.InstanceLease)
		} else {
			__antithesis_instrumentation__.Notify(20414)
		}
		__antithesis_instrumentation__.Notify(20398)

		if in.NextLiveInstanceID != 0 {
			__antithesis_instrumentation__.Notify(20415)
			if err := s.handleNextLiveInstanceID(
				&h, &tenant, &instance, base.SQLInstanceID(in.NextLiveInstanceID),
			); err != nil {
				__antithesis_instrumentation__.Notify(20416)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20417)
			}
		} else {
			__antithesis_instrumentation__.Notify(20418)
		}
		__antithesis_instrumentation__.Notify(20399)

		if instance.Seq == 0 || func() bool {
			__antithesis_instrumentation__.Notify(20419)
			return instance.Seq < in.SeqNum == true
		}() == true {
			__antithesis_instrumentation__.Notify(20420)
			tenant.Consumption.Add(&in.ConsumptionSinceLastRequest)
			instance.Seq = in.SeqNum
		} else {
			__antithesis_instrumentation__.Notify(20421)
		}
		__antithesis_instrumentation__.Notify(20400)

		*result = tenant.Bucket.Request(ctx, in)

		instance.LastUpdate.Time = now
		if err := h.updateTenantAndInstanceState(tenant, instance); err != nil {
			__antithesis_instrumentation__.Notify(20422)
			return err
		} else {
			__antithesis_instrumentation__.Notify(20423)
		}
		__antithesis_instrumentation__.Notify(20401)

		if err := h.maybeCheckInvariants(); err != nil {
			__antithesis_instrumentation__.Notify(20424)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(20425)
		}
		__antithesis_instrumentation__.Notify(20402)
		consumption = tenant.Consumption
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(20426)
		return &roachpb.TokenBucketResponse{
			Error: errors.EncodeError(ctx, err),
		}
	} else {
		__antithesis_instrumentation__.Notify(20427)
	}
	__antithesis_instrumentation__.Notify(20387)

	metrics.totalRU.Update(consumption.RU)
	metrics.totalReadRequests.Update(int64(consumption.ReadRequests))
	metrics.totalReadBytes.Update(int64(consumption.ReadBytes))
	metrics.totalWriteRequests.Update(int64(consumption.WriteRequests))
	metrics.totalWriteBytes.Update(int64(consumption.WriteBytes))
	metrics.totalSQLPodsCPUSeconds.Update(consumption.SQLPodsCPUSeconds)
	metrics.totalPGWireEgressBytes.Update(int64(consumption.PGWireEgressBytes))
	return result
}

func (s *instance) handleNextLiveInstanceID(
	h *sysTableHelper,
	tenant *tenantState,
	instance *instanceState,
	nextLiveInstanceID base.SQLInstanceID,
) error {
	__antithesis_instrumentation__.Notify(20428)

	expected := instance.NextInstance
	if expected == 0 {
		__antithesis_instrumentation__.Notify(20432)
		expected = tenant.FirstInstance
	} else {
		__antithesis_instrumentation__.Notify(20433)
	}
	__antithesis_instrumentation__.Notify(20429)
	if nextLiveInstanceID == expected {
		__antithesis_instrumentation__.Notify(20434)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(20435)
	}
	__antithesis_instrumentation__.Notify(20430)
	var err error
	cutoff := s.timeSource.Now().Add(-instanceInactivity.Get(&s.settings.SV))
	if nextLiveInstanceID > instance.ID {
		__antithesis_instrumentation__.Notify(20436)

		if instance.NextInstance != 0 && func() bool {
			__antithesis_instrumentation__.Notify(20437)
			return instance.NextInstance < nextLiveInstanceID == true
		}() == true {
			__antithesis_instrumentation__.Notify(20438)

			instance.NextInstance, err = h.maybeCleanupStaleInstances(
				cutoff, instance.NextInstance, nextLiveInstanceID,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(20439)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20440)
			}
		} else {
			__antithesis_instrumentation__.Notify(20441)
		}
	} else {
		__antithesis_instrumentation__.Notify(20442)

		if tenant.FirstInstance < nextLiveInstanceID {
			__antithesis_instrumentation__.Notify(20444)

			tenant.FirstInstance, err = h.maybeCleanupStaleInstances(
				cutoff, tenant.FirstInstance, nextLiveInstanceID,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(20445)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20446)
			}
		} else {
			__antithesis_instrumentation__.Notify(20447)
		}
		__antithesis_instrumentation__.Notify(20443)
		if instance.NextInstance != 0 {
			__antithesis_instrumentation__.Notify(20448)

			instance.NextInstance, err = h.maybeCleanupStaleInstances(
				cutoff, instance.NextInstance, -1,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(20449)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20450)
			}
		} else {
			__antithesis_instrumentation__.Notify(20451)
		}
	}
	__antithesis_instrumentation__.Notify(20431)
	return nil
}
