package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	fmt "fmt"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobsprotectedts"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	roachpb "github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	pbtypes "github.com/gogo/protobuf/types"
)

func maybeUpdateSchedulePTSRecord(
	ctx context.Context,
	exec *sql.ExecutorConfig,
	backupDetails jobspb.BackupDetails,
	id jobspb.JobID,
) error {
	__antithesis_instrumentation__.Notify(12720)
	env := scheduledjobs.ProdJobSchedulerEnv
	if knobs, ok := exec.DistSQLSrv.TestingKnobs.JobsTestingKnobs.(*jobs.TestingKnobs); ok {
		__antithesis_instrumentation__.Notify(12722)
		if knobs.JobSchedulerEnv != nil {
			__antithesis_instrumentation__.Notify(12723)
			env = knobs.JobSchedulerEnv
		} else {
			__antithesis_instrumentation__.Notify(12724)
		}
	} else {
		__antithesis_instrumentation__.Notify(12725)
	}
	__antithesis_instrumentation__.Notify(12721)

	return exec.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(12726)

		datums, err := exec.InternalExecutor.QueryRowEx(
			ctx,
			"lookup-schedule-info",
			txn,
			sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
			fmt.Sprintf(
				"SELECT created_by_id FROM %s WHERE id=$1 AND created_by_type=$2",
				env.SystemJobsTableName()),
			id, jobs.CreatedByScheduledJobs)

		if err != nil {
			__antithesis_instrumentation__.Notify(12733)
			return errors.Wrap(err, "schedule info lookup")
		} else {
			__antithesis_instrumentation__.Notify(12734)
		}
		__antithesis_instrumentation__.Notify(12727)
		if datums == nil {
			__antithesis_instrumentation__.Notify(12735)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(12736)
		}
		__antithesis_instrumentation__.Notify(12728)

		scheduleID := int64(tree.MustBeDInt(datums[0]))
		_, args, err := getScheduledBackupExecutionArgsFromSchedule(ctx, env, txn,
			exec.InternalExecutor, scheduleID)
		if err != nil {
			__antithesis_instrumentation__.Notify(12737)
			return errors.Wrap(err, "load scheduled job")
		} else {
			__antithesis_instrumentation__.Notify(12738)
		}
		__antithesis_instrumentation__.Notify(12729)

		if !args.ChainProtectedTimestampRecords {
			__antithesis_instrumentation__.Notify(12739)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(12740)
		}
		__antithesis_instrumentation__.Notify(12730)

		if backupDetails.SchedulePTSChainingRecord == nil {
			__antithesis_instrumentation__.Notify(12741)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(12742)
		}
		__antithesis_instrumentation__.Notify(12731)

		switch args.BackupType {
		case ScheduledBackupExecutionArgs_INCREMENTAL:
			__antithesis_instrumentation__.Notify(12743)
			if backupDetails.SchedulePTSChainingRecord.Action != jobspb.SchedulePTSChainingRecord_UPDATE {
				__antithesis_instrumentation__.Notify(12748)
				return errors.AssertionFailedf("incremental backup has unexpected chaining action %d on"+
					" backup job details", backupDetails.SchedulePTSChainingRecord.Action)
			} else {
				__antithesis_instrumentation__.Notify(12749)
			}
			__antithesis_instrumentation__.Notify(12744)
			if err := manageIncrementalBackupPTSChaining(ctx,
				backupDetails.SchedulePTSChainingRecord.ProtectedTimestampRecord,
				backupDetails.EndTime, exec, txn, scheduleID); err != nil {
				__antithesis_instrumentation__.Notify(12750)
				return errors.Wrap(err, "failed to manage chaining of pts record during a inc backup")
			} else {
				__antithesis_instrumentation__.Notify(12751)
			}
		case ScheduledBackupExecutionArgs_FULL:
			__antithesis_instrumentation__.Notify(12745)
			if backupDetails.SchedulePTSChainingRecord.Action != jobspb.SchedulePTSChainingRecord_RELEASE {
				__antithesis_instrumentation__.Notify(12752)
				return errors.AssertionFailedf("full backup has unexpected chaining action %d on"+
					" backup job details", backupDetails.SchedulePTSChainingRecord.Action)
			} else {
				__antithesis_instrumentation__.Notify(12753)
			}
			__antithesis_instrumentation__.Notify(12746)
			if err := manageFullBackupPTSChaining(ctx, env, txn, backupDetails, exec, args); err != nil {
				__antithesis_instrumentation__.Notify(12754)
				return errors.Wrap(err, "failed to manage chaining of pts record during a full backup")
			} else {
				__antithesis_instrumentation__.Notify(12755)
			}
		default:
			__antithesis_instrumentation__.Notify(12747)
		}
		__antithesis_instrumentation__.Notify(12732)
		return nil
	})
}

func manageFullBackupPTSChaining(
	ctx context.Context,
	env scheduledjobs.JobSchedulerEnv,
	txn *kv.Txn,
	backupDetails jobspb.BackupDetails,
	exec *sql.ExecutorConfig,
	args *ScheduledBackupExecutionArgs,
) error {
	__antithesis_instrumentation__.Notify(12756)

	incSj, incArgs, err := getScheduledBackupExecutionArgsFromSchedule(ctx, env, txn,
		exec.InternalExecutor, args.DependentScheduleID)
	if err != nil {
		__antithesis_instrumentation__.Notify(12763)
		if jobs.HasScheduledJobNotFoundError(err) {
			__antithesis_instrumentation__.Notify(12765)
			log.Warningf(ctx, "could not find dependent schedule with id %d",
				args.DependentScheduleID)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(12766)
		}
		__antithesis_instrumentation__.Notify(12764)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12767)
	}
	__antithesis_instrumentation__.Notify(12757)

	targetToProtect, deprecatedSpansToProtect, err := getTargetProtectedByBackup(ctx, backupDetails, txn, exec)
	if err != nil {
		__antithesis_instrumentation__.Notify(12768)
		return errors.Wrap(err, "getting spans to protect")
	} else {
		__antithesis_instrumentation__.Notify(12769)
	}
	__antithesis_instrumentation__.Notify(12758)

	if targetToProtect != nil {
		__antithesis_instrumentation__.Notify(12770)
		targetToProtect.IgnoreIfExcludedFromBackup = true
	} else {
		__antithesis_instrumentation__.Notify(12771)
	}
	__antithesis_instrumentation__.Notify(12759)

	ptsRecord, err := protectTimestampRecordForSchedule(ctx, targetToProtect, deprecatedSpansToProtect,
		backupDetails.EndTime, incSj.ScheduleID(), exec, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(12772)
		return errors.Wrap(err, "protect and verify pts record for schedule")
	} else {
		__antithesis_instrumentation__.Notify(12773)
	}
	__antithesis_instrumentation__.Notify(12760)

	if err := releaseProtectedTimestamp(ctx, txn, exec.ProtectedTimestampProvider,
		backupDetails.SchedulePTSChainingRecord.ProtectedTimestampRecord); err != nil {
		__antithesis_instrumentation__.Notify(12774)
		return errors.Wrap(err, "release pts record for schedule")
	} else {
		__antithesis_instrumentation__.Notify(12775)
	}
	__antithesis_instrumentation__.Notify(12761)

	incArgs.ProtectedTimestampRecord = &ptsRecord
	any, err := pbtypes.MarshalAny(incArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(12776)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12777)
	}
	__antithesis_instrumentation__.Notify(12762)
	incSj.SetExecutionDetails(incSj.ExecutorType(), jobspb.ExecutionArguments{Args: any})
	return incSj.Update(ctx, exec.InternalExecutor, txn)
}

func manageIncrementalBackupPTSChaining(
	ctx context.Context,
	ptsRecordID *uuid.UUID,
	tsToProtect hlc.Timestamp,
	exec *sql.ExecutorConfig,
	txn *kv.Txn,
	scheduleID int64,
) error {
	__antithesis_instrumentation__.Notify(12778)
	if ptsRecordID == nil {
		__antithesis_instrumentation__.Notify(12781)
		return errors.Newf("unexpected nil pts record id on incremental schedule %d", scheduleID)
	} else {
		__antithesis_instrumentation__.Notify(12782)
	}
	__antithesis_instrumentation__.Notify(12779)
	err := exec.ProtectedTimestampProvider.UpdateTimestamp(ctx, txn, *ptsRecordID,
		tsToProtect)

	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(12783)
		return errors.Is(err, protectedts.ErrNotExists) == true
	}() == true {
		__antithesis_instrumentation__.Notify(12784)
		log.Warningf(ctx, "failed to update timestamp record %d since it does not exist", ptsRecordID)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(12785)
	}
	__antithesis_instrumentation__.Notify(12780)
	return err
}

func getTargetProtectedByBackup(
	ctx context.Context, backupDetails jobspb.BackupDetails, txn *kv.Txn, exec *sql.ExecutorConfig,
) (target *ptpb.Target, deprecatedSpans []roachpb.Span, err error) {
	__antithesis_instrumentation__.Notify(12786)
	if backupDetails.ProtectedTimestampRecord == nil {
		__antithesis_instrumentation__.Notify(12789)
		return nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(12790)
	}
	__antithesis_instrumentation__.Notify(12787)

	ptsRecord, err := exec.ProtectedTimestampProvider.GetRecord(ctx, txn,
		*backupDetails.ProtectedTimestampRecord)
	if err != nil {
		__antithesis_instrumentation__.Notify(12791)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(12792)
	}
	__antithesis_instrumentation__.Notify(12788)

	return ptsRecord.Target, ptsRecord.DeprecatedSpans, nil
}

func protectTimestampRecordForSchedule(
	ctx context.Context,
	targetToProtect *ptpb.Target,
	deprecatedSpansToProtect roachpb.Spans,
	tsToProtect hlc.Timestamp,
	scheduleID int64,
	exec *sql.ExecutorConfig,
	txn *kv.Txn,
) (uuid.UUID, error) {
	__antithesis_instrumentation__.Notify(12793)
	protectedtsID := uuid.MakeV4()
	rec := jobsprotectedts.MakeRecord(protectedtsID, scheduleID, tsToProtect, deprecatedSpansToProtect,
		jobsprotectedts.Schedules, targetToProtect)
	return protectedtsID, exec.ProtectedTimestampProvider.Protect(ctx, txn, rec)
}
