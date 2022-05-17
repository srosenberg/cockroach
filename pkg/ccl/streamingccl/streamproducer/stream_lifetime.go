package streamproducer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streampb"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobsprotectedts"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/streaming"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

func completeStreamIngestion(
	evalCtx *tree.EvalContext,
	txn *kv.Txn,
	streamID streaming.StreamID,
	cutoverTimestamp hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(27061)

	const jobsQuery = `SELECT progress FROM system.jobs WHERE id=$1 FOR UPDATE`
	row, err := evalCtx.Planner.QueryRowEx(evalCtx.Context,
		"get-stream-ingestion-job-metadata",
		txn, sessiondata.NodeUserSessionDataOverride, jobsQuery, streamID)
	if err != nil {
		__antithesis_instrumentation__.Notify(27069)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27070)
	}
	__antithesis_instrumentation__.Notify(27062)

	if row == nil {
		__antithesis_instrumentation__.Notify(27071)
		return errors.Newf("job %d: not found in system.jobs table", streamID)
	} else {
		__antithesis_instrumentation__.Notify(27072)
	}
	__antithesis_instrumentation__.Notify(27063)

	progress, err := jobs.UnmarshalProgress(row[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(27073)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27074)
	}
	__antithesis_instrumentation__.Notify(27064)
	var sp *jobspb.Progress_StreamIngest
	var ok bool
	if sp, ok = progress.GetDetails().(*jobspb.Progress_StreamIngest); !ok {
		__antithesis_instrumentation__.Notify(27075)
		return errors.Newf("job %d: not of expected type StreamIngest", streamID)
	} else {
		__antithesis_instrumentation__.Notify(27076)
	}
	__antithesis_instrumentation__.Notify(27065)

	hw := progress.GetHighWater()
	if hw == nil || func() bool {
		__antithesis_instrumentation__.Notify(27077)
		return hw.Less(cutoverTimestamp) == true
	}() == true {
		__antithesis_instrumentation__.Notify(27078)
		var highWaterTimestamp hlc.Timestamp
		if hw != nil {
			__antithesis_instrumentation__.Notify(27080)
			highWaterTimestamp = *hw
		} else {
			__antithesis_instrumentation__.Notify(27081)
		}
		__antithesis_instrumentation__.Notify(27079)
		return errors.Newf("cannot cutover to a timestamp %s that is after the latest resolved time"+
			" %s for job %d", cutoverTimestamp.String(), highWaterTimestamp.String(), streamID)
	} else {
		__antithesis_instrumentation__.Notify(27082)
	}
	__antithesis_instrumentation__.Notify(27066)

	if !sp.StreamIngest.CutoverTime.IsEmpty() {
		__antithesis_instrumentation__.Notify(27083)
		return errors.Newf("cutover timestamp already set to %s, "+
			"job %d is in the process of cutting over", sp.StreamIngest.CutoverTime.String(), streamID)
	} else {
		__antithesis_instrumentation__.Notify(27084)
	}
	__antithesis_instrumentation__.Notify(27067)

	sp.StreamIngest.CutoverTime = cutoverTimestamp
	progress.ModifiedMicros = timeutil.ToUnixMicros(txn.ReadTimestamp().GoTime())
	progressBytes, err := protoutil.Marshal(progress)
	if err != nil {
		__antithesis_instrumentation__.Notify(27085)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27086)
	}
	__antithesis_instrumentation__.Notify(27068)
	updateJobQuery := `UPDATE system.jobs SET progress=$1 WHERE id=$2`
	_, err = evalCtx.Planner.QueryRowEx(evalCtx.Context,
		"set-stream-ingestion-job-metadata", txn,
		sessiondata.NodeUserSessionDataOverride, updateJobQuery, progressBytes, streamID)
	return err
}

func startReplicationStreamJob(
	evalCtx *tree.EvalContext, txn *kv.Txn, tenantID uint64,
) (streaming.StreamID, error) {
	__antithesis_instrumentation__.Notify(27087)
	execConfig := evalCtx.Planner.ExecutorConfig().(*sql.ExecutorConfig)
	hasAdminRole, err := evalCtx.SessionAccessor.HasAdminRole(evalCtx.Ctx())

	if err != nil {
		__antithesis_instrumentation__.Notify(27092)
		return streaming.InvalidStreamID, err
	} else {
		__antithesis_instrumentation__.Notify(27093)
	}
	__antithesis_instrumentation__.Notify(27088)

	if !hasAdminRole {
		__antithesis_instrumentation__.Notify(27094)
		return streaming.InvalidStreamID, errors.New("admin role required to start stream replication jobs")
	} else {
		__antithesis_instrumentation__.Notify(27095)
	}
	__antithesis_instrumentation__.Notify(27089)

	registry := execConfig.JobRegistry
	timeout := streamingccl.StreamReplicationJobLivenessTimeout.Get(&evalCtx.Settings.SV)
	ptsID := uuid.MakeV4()
	jr := makeProducerJobRecord(registry, tenantID, timeout, evalCtx.SessionData().User(), ptsID)
	if _, err := registry.CreateAdoptableJobWithTxn(evalCtx.Ctx(), jr, jr.JobID, txn); err != nil {
		__antithesis_instrumentation__.Notify(27096)
		return streaming.InvalidStreamID, err
	} else {
		__antithesis_instrumentation__.Notify(27097)
	}
	__antithesis_instrumentation__.Notify(27090)

	ptp := execConfig.ProtectedTimestampProvider
	statementTime := hlc.Timestamp{
		WallTime: evalCtx.GetStmtTimestamp().UnixNano(),
	}

	deprecatedSpansToProtect := roachpb.Spans{*makeTenantSpan(tenantID)}
	targetToProtect := ptpb.MakeTenantsTarget([]roachpb.TenantID{roachpb.MakeTenantID(tenantID)})

	pts := jobsprotectedts.MakeRecord(ptsID, int64(jr.JobID), statementTime,
		deprecatedSpansToProtect, jobsprotectedts.Jobs, targetToProtect)

	if err := ptp.Protect(evalCtx.Ctx(), txn, pts); err != nil {
		__antithesis_instrumentation__.Notify(27098)
		return streaming.InvalidStreamID, err
	} else {
		__antithesis_instrumentation__.Notify(27099)
	}
	__antithesis_instrumentation__.Notify(27091)
	return streaming.StreamID(jr.JobID), nil
}

func updateReplicationStreamProgress(
	ctx context.Context,
	expiration time.Time,
	ptsProvider protectedts.Provider,
	registry *jobs.Registry,
	streamID streaming.StreamID,
	ts hlc.Timestamp,
	txn *kv.Txn,
) (status streampb.StreamReplicationStatus, err error) {
	__antithesis_instrumentation__.Notify(27100)
	const useReadLock = false
	err = registry.UpdateJobWithTxn(ctx, jobspb.JobID(streamID), txn, useReadLock,
		func(txn *kv.Txn, md jobs.JobMetadata, ju *jobs.JobUpdater) error {
			__antithesis_instrumentation__.Notify(27103)
			if md.Status == jobs.StatusRunning {
				__antithesis_instrumentation__.Notify(27110)
				status.StreamStatus = streampb.StreamReplicationStatus_STREAM_ACTIVE
			} else {
				__antithesis_instrumentation__.Notify(27111)
				if md.Status == jobs.StatusPaused {
					__antithesis_instrumentation__.Notify(27112)
					status.StreamStatus = streampb.StreamReplicationStatus_STREAM_PAUSED
				} else {
					__antithesis_instrumentation__.Notify(27113)
					if md.Status.Terminal() {
						__antithesis_instrumentation__.Notify(27114)
						status.StreamStatus = streampb.StreamReplicationStatus_STREAM_INACTIVE
					} else {
						__antithesis_instrumentation__.Notify(27115)
						status.StreamStatus = streampb.StreamReplicationStatus_UNKNOWN_STREAM_STATUS_RETRY
					}
				}
			}
			__antithesis_instrumentation__.Notify(27104)

			if status.StreamStatus != streampb.StreamReplicationStatus_STREAM_ACTIVE && func() bool {
				__antithesis_instrumentation__.Notify(27116)
				return status.StreamStatus != streampb.StreamReplicationStatus_STREAM_PAUSED == true
			}() == true {
				__antithesis_instrumentation__.Notify(27117)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(27118)
			}
			__antithesis_instrumentation__.Notify(27105)

			ptsID := *md.Payload.GetStreamReplication().ProtectedTimestampRecord
			ptsRecord, err := ptsProvider.GetRecord(ctx, txn, ptsID)
			if err != nil {
				__antithesis_instrumentation__.Notify(27119)
				return err
			} else {
				__antithesis_instrumentation__.Notify(27120)
			}
			__antithesis_instrumentation__.Notify(27106)
			status.ProtectedTimestamp = &ptsRecord.Timestamp
			if status.StreamStatus != streampb.StreamReplicationStatus_STREAM_ACTIVE {
				__antithesis_instrumentation__.Notify(27121)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(27122)
			}
			__antithesis_instrumentation__.Notify(27107)

			if shouldUpdatePTS := ptsRecord.Timestamp.Less(ts); shouldUpdatePTS {
				__antithesis_instrumentation__.Notify(27123)
				if err = ptsProvider.UpdateTimestamp(ctx, txn, ptsID, ts); err != nil {
					__antithesis_instrumentation__.Notify(27125)
					return err
				} else {
					__antithesis_instrumentation__.Notify(27126)
				}
				__antithesis_instrumentation__.Notify(27124)
				status.ProtectedTimestamp = &ts
			} else {
				__antithesis_instrumentation__.Notify(27127)
			}
			__antithesis_instrumentation__.Notify(27108)

			if p := md.Progress; expiration.After(p.GetStreamReplication().Expiration) {
				__antithesis_instrumentation__.Notify(27128)
				p.GetStreamReplication().Expiration = expiration
				ju.UpdateProgress(p)
			} else {
				__antithesis_instrumentation__.Notify(27129)
			}
			__antithesis_instrumentation__.Notify(27109)
			return nil
		})
	__antithesis_instrumentation__.Notify(27101)

	if jobs.HasJobNotFoundError(err) || func() bool {
		__antithesis_instrumentation__.Notify(27130)
		return testutils.IsError(err, "not found in system.jobs table") == true
	}() == true {
		__antithesis_instrumentation__.Notify(27131)
		status.StreamStatus = streampb.StreamReplicationStatus_STREAM_INACTIVE
		err = nil
	} else {
		__antithesis_instrumentation__.Notify(27132)
	}
	__antithesis_instrumentation__.Notify(27102)

	return status, err
}

func heartbeatReplicationStream(
	evalCtx *tree.EvalContext, streamID streaming.StreamID, frontier hlc.Timestamp, txn *kv.Txn,
) (streampb.StreamReplicationStatus, error) {
	__antithesis_instrumentation__.Notify(27133)

	execConfig := evalCtx.Planner.ExecutorConfig().(*sql.ExecutorConfig)
	timeout := streamingccl.StreamReplicationJobLivenessTimeout.Get(&evalCtx.Settings.SV)
	expirationTime := timeutil.Now().Add(timeout)

	return updateReplicationStreamProgress(evalCtx.Ctx(),
		expirationTime, execConfig.ProtectedTimestampProvider, execConfig.JobRegistry, streamID, frontier, txn)
}
