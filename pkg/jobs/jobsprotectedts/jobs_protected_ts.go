package jobsprotectedts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptreconcile"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type MetaType int

const (
	Jobs MetaType = iota

	Schedules
)

var metaTypes = map[MetaType]string{Jobs: "jobs", Schedules: "schedules"}

func GetMetaType(metaType MetaType) string {
	__antithesis_instrumentation__.Notify(84128)
	return metaTypes[metaType]
}

func MakeStatusFunc(
	jr *jobs.Registry, ie sqlutil.InternalExecutor, metaType MetaType,
) ptreconcile.StatusFunc {
	__antithesis_instrumentation__.Notify(84129)
	switch metaType {
	case Jobs:
		__antithesis_instrumentation__.Notify(84131)
		return func(ctx context.Context, txn *kv.Txn, meta []byte) (shouldRemove bool, _ error) {
			__antithesis_instrumentation__.Notify(84134)
			jobID, err := decodeID(meta)
			if err != nil {
				__antithesis_instrumentation__.Notify(84138)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(84139)
			}
			__antithesis_instrumentation__.Notify(84135)
			j, err := jr.LoadJobWithTxn(ctx, jobspb.JobID(jobID), txn)
			if jobs.HasJobNotFoundError(err) {
				__antithesis_instrumentation__.Notify(84140)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(84141)
			}
			__antithesis_instrumentation__.Notify(84136)
			if err != nil {
				__antithesis_instrumentation__.Notify(84142)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(84143)
			}
			__antithesis_instrumentation__.Notify(84137)
			isTerminal := j.CheckTerminalStatus(ctx, txn)
			return isTerminal, nil
		}
	case Schedules:
		__antithesis_instrumentation__.Notify(84132)
		return func(ctx context.Context, txn *kv.Txn, meta []byte) (shouldRemove bool, _ error) {
			__antithesis_instrumentation__.Notify(84144)
			scheduleID, err := decodeID(meta)
			if err != nil {
				__antithesis_instrumentation__.Notify(84147)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(84148)
			}
			__antithesis_instrumentation__.Notify(84145)
			_, err = jobs.LoadScheduledJob(ctx, scheduledjobs.ProdJobSchedulerEnv, scheduleID, ie, txn)
			if jobs.HasScheduledJobNotFoundError(err) {
				__antithesis_instrumentation__.Notify(84149)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(84150)
			}
			__antithesis_instrumentation__.Notify(84146)
			return false, err
		}
	default:
		__antithesis_instrumentation__.Notify(84133)
	}
	__antithesis_instrumentation__.Notify(84130)
	return nil
}

func MakeRecord(
	recordID uuid.UUID,
	metaID int64,
	tsToProtect hlc.Timestamp,
	deprecatedSpans []roachpb.Span,
	metaType MetaType,
	target *ptpb.Target,
) *ptpb.Record {
	__antithesis_instrumentation__.Notify(84151)
	return &ptpb.Record{
		ID:              recordID.GetBytesMut(),
		Timestamp:       tsToProtect,
		Mode:            ptpb.PROTECT_AFTER,
		MetaType:        metaTypes[metaType],
		Meta:            encodeID(metaID),
		DeprecatedSpans: deprecatedSpans,
		Target:          target,
	}
}

func encodeID(id int64) []byte {
	__antithesis_instrumentation__.Notify(84152)
	return []byte(strconv.FormatInt(id, 10))
}

func decodeID(meta []byte) (id int64, err error) {
	__antithesis_instrumentation__.Notify(84153)
	id, err = strconv.ParseInt(string(meta), 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(84155)
		return 0, errors.Wrapf(err, "failed to interpret meta %q as bytes", meta)
	} else {
		__antithesis_instrumentation__.Notify(84156)
	}
	__antithesis_instrumentation__.Notify(84154)
	return id, err
}
