package gcjob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var (
	MaxSQLGCInterval = 5 * time.Minute
)

func SetSmallMaxGCIntervalForTest() func() {
	__antithesis_instrumentation__.Notify(492198)
	oldInterval := MaxSQLGCInterval
	MaxSQLGCInterval = 500 * time.Millisecond
	return func() {
		__antithesis_instrumentation__.Notify(492199)
		MaxSQLGCInterval = oldInterval
	}
}

type schemaChangeGCResumer struct {
	jobID jobspb.JobID
}

func performGC(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	details *jobspb.SchemaChangeGCDetails,
	progress *jobspb.SchemaChangeGCProgress,
) error {
	__antithesis_instrumentation__.Notify(492200)
	if details.Tenant != nil {
		__antithesis_instrumentation__.Notify(492203)
		return errors.Wrapf(
			gcTenant(ctx, execCfg, details.Tenant.ID, progress),
			"attempting to GC tenant %+v", details.Tenant,
		)
	} else {
		__antithesis_instrumentation__.Notify(492204)
	}
	__antithesis_instrumentation__.Notify(492201)
	if details.Indexes != nil {
		__antithesis_instrumentation__.Notify(492205)
		return errors.Wrap(gcIndexes(ctx, execCfg, details.ParentID, progress), "attempting to GC indexes")
	} else {
		__antithesis_instrumentation__.Notify(492206)
		if details.Tables != nil {
			__antithesis_instrumentation__.Notify(492207)
			if err := gcTables(ctx, execCfg, progress); err != nil {
				__antithesis_instrumentation__.Notify(492209)
				return errors.Wrap(err, "attempting to GC tables")
			} else {
				__antithesis_instrumentation__.Notify(492210)
			}
			__antithesis_instrumentation__.Notify(492208)

			if details.ParentID != descpb.InvalidID && func() bool {
				__antithesis_instrumentation__.Notify(492211)
				return isDoneGC(progress) == true
			}() == true {
				__antithesis_instrumentation__.Notify(492212)
				if err := deleteDatabaseZoneConfig(
					ctx,
					execCfg.DB,
					execCfg.Codec,
					execCfg.Settings,
					details.ParentID,
				); err != nil {
					__antithesis_instrumentation__.Notify(492213)
					return errors.Wrap(err, "deleting database zone config")
				} else {
					__antithesis_instrumentation__.Notify(492214)
				}
			} else {
				__antithesis_instrumentation__.Notify(492215)
			}
		} else {
			__antithesis_instrumentation__.Notify(492216)
		}
	}
	__antithesis_instrumentation__.Notify(492202)
	return nil
}

func unsplitRangesForTables(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	droppedTables []jobspb.SchemaChangeGCDetails_DroppedID,
) error {
	__antithesis_instrumentation__.Notify(492217)
	if !execCfg.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(492220)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(492221)
	}
	__antithesis_instrumentation__.Notify(492218)

	for _, droppedTable := range droppedTables {
		__antithesis_instrumentation__.Notify(492222)
		startKey := execCfg.Codec.TablePrefix(uint32(droppedTable.ID))
		span := roachpb.Span{
			Key:    startKey,
			EndKey: startKey.PrefixEnd(),
		}
		if err := sql.UnsplitRangesInSpan(ctx, execCfg.DB, span); err != nil {
			__antithesis_instrumentation__.Notify(492223)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492224)
		}
	}
	__antithesis_instrumentation__.Notify(492219)

	return nil
}

func unsplitRangesForIndexes(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	indexes []jobspb.SchemaChangeGCDetails_DroppedIndex,
	parentTableID descpb.ID,
) error {
	__antithesis_instrumentation__.Notify(492225)
	if !execCfg.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(492228)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(492229)
	}
	__antithesis_instrumentation__.Notify(492226)

	for _, idx := range indexes {
		__antithesis_instrumentation__.Notify(492230)
		startKey := execCfg.Codec.IndexPrefix(uint32(parentTableID), uint32(idx.IndexID))
		idxSpan := roachpb.Span{
			Key:    startKey,
			EndKey: startKey.PrefixEnd(),
		}

		if err := sql.UnsplitRangesInSpan(ctx, execCfg.DB, idxSpan); err != nil {
			__antithesis_instrumentation__.Notify(492231)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492232)
		}
	}
	__antithesis_instrumentation__.Notify(492227)

	return nil
}

func maybeUnsplitRanges(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	jobID jobspb.JobID,
	details *jobspb.SchemaChangeGCDetails,
	progress *jobspb.SchemaChangeGCProgress,
) error {
	__antithesis_instrumentation__.Notify(492233)
	if progress.RangesUnsplitDone {
		__antithesis_instrumentation__.Notify(492237)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(492238)
	}
	__antithesis_instrumentation__.Notify(492234)

	if len(details.Indexes) > 0 {
		__antithesis_instrumentation__.Notify(492239)
		if err := unsplitRangesForIndexes(ctx, execCfg, details.Indexes, details.ParentID); err != nil {
			__antithesis_instrumentation__.Notify(492240)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492241)
		}
	} else {
		__antithesis_instrumentation__.Notify(492242)
	}
	__antithesis_instrumentation__.Notify(492235)

	if len(details.Tables) > 0 {
		__antithesis_instrumentation__.Notify(492243)
		if err := unsplitRangesForTables(ctx, execCfg, details.Tables); err != nil {
			__antithesis_instrumentation__.Notify(492244)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492245)
		}
	} else {
		__antithesis_instrumentation__.Notify(492246)
	}
	__antithesis_instrumentation__.Notify(492236)

	progress.RangesUnsplitDone = true
	persistProgress(ctx, execCfg, jobID, progress, runningStatusGC(progress))

	return nil
}

func (r schemaChangeGCResumer) Resume(ctx context.Context, execCtx interface{}) (err error) {
	__antithesis_instrumentation__.Notify(492247)
	defer func() {
		__antithesis_instrumentation__.Notify(492252)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(492253)
			return !r.isPermanentGCError(err) == true
		}() == true {
			__antithesis_instrumentation__.Notify(492254)
			err = jobs.MarkAsRetryJobError(err)
		} else {
			__antithesis_instrumentation__.Notify(492255)
		}
	}()
	__antithesis_instrumentation__.Notify(492248)
	p := execCtx.(sql.JobExecContext)

	execCfg := p.ExecCfg()
	if fn := execCfg.GCJobTestingKnobs.RunBeforeResume; fn != nil {
		__antithesis_instrumentation__.Notify(492256)
		if err := fn(r.jobID); err != nil {
			__antithesis_instrumentation__.Notify(492257)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492258)
		}
	} else {
		__antithesis_instrumentation__.Notify(492259)
	}
	__antithesis_instrumentation__.Notify(492249)
	details, progress, err := initDetailsAndProgress(ctx, execCfg, r.jobID)
	if err != nil {
		__antithesis_instrumentation__.Notify(492260)
		return err
	} else {
		__antithesis_instrumentation__.Notify(492261)
	}
	__antithesis_instrumentation__.Notify(492250)

	if err := maybeUnsplitRanges(ctx, execCfg, r.jobID, details, progress); err != nil {
		__antithesis_instrumentation__.Notify(492262)
		return err
	} else {
		__antithesis_instrumentation__.Notify(492263)
	}
	__antithesis_instrumentation__.Notify(492251)

	tableDropTimes, indexDropTimes := getDropTimes(details)

	timer := timeutil.NewTimer()
	defer timer.Stop()
	timer.Reset(0)
	gossipUpdateC, cleanup := execCfg.GCJobNotifier.AddNotifyee(ctx)
	defer cleanup()
	for {
		__antithesis_instrumentation__.Notify(492264)
		select {
		case <-gossipUpdateC:
			__antithesis_instrumentation__.Notify(492270)
			if log.V(2) {
				__antithesis_instrumentation__.Notify(492273)
				log.Info(ctx, "received a new system config")
			} else {
				__antithesis_instrumentation__.Notify(492274)
			}
		case <-timer.C:
			__antithesis_instrumentation__.Notify(492271)
			timer.Read = true
			if log.V(2) {
				__antithesis_instrumentation__.Notify(492275)
				log.Info(ctx, "SchemaChangeGC timer triggered")
			} else {
				__antithesis_instrumentation__.Notify(492276)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(492272)
			return ctx.Err()
		}
		__antithesis_instrumentation__.Notify(492265)

		var expired bool
		earliestDeadline := timeutil.Unix(0, math.MaxInt64)
		if details.Tenant == nil {
			__antithesis_instrumentation__.Notify(492277)
			remainingTables := getAllTablesWaitingForGC(details, progress)
			expired, earliestDeadline = refreshTables(
				ctx, execCfg, remainingTables, tableDropTimes, indexDropTimes, r.jobID, progress,
			)
		} else {
			__antithesis_instrumentation__.Notify(492278)
			expired, earliestDeadline, err = refreshTenant(ctx, execCfg, details.Tenant.DropTime, details, progress)
			if err != nil {
				__antithesis_instrumentation__.Notify(492279)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492280)
			}
		}
		__antithesis_instrumentation__.Notify(492266)
		timerDuration := time.Until(earliestDeadline)

		if expired {
			__antithesis_instrumentation__.Notify(492281)

			persistProgress(ctx, execCfg, r.jobID, progress, runningStatusGC(progress))
			if fn := execCfg.GCJobTestingKnobs.RunBeforePerformGC; fn != nil {
				__antithesis_instrumentation__.Notify(492284)
				if err := fn(r.jobID); err != nil {
					__antithesis_instrumentation__.Notify(492285)
					return err
				} else {
					__antithesis_instrumentation__.Notify(492286)
				}
			} else {
				__antithesis_instrumentation__.Notify(492287)
			}
			__antithesis_instrumentation__.Notify(492282)
			if err := performGC(ctx, execCfg, details, progress); err != nil {
				__antithesis_instrumentation__.Notify(492288)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492289)
			}
			__antithesis_instrumentation__.Notify(492283)
			persistProgress(ctx, execCfg, r.jobID, progress, sql.RunningStatusWaitingGC)

			timerDuration = 0
		} else {
			__antithesis_instrumentation__.Notify(492290)
		}
		__antithesis_instrumentation__.Notify(492267)

		if isDoneGC(progress) {
			__antithesis_instrumentation__.Notify(492291)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(492292)
		}
		__antithesis_instrumentation__.Notify(492268)

		if timerDuration > MaxSQLGCInterval {
			__antithesis_instrumentation__.Notify(492293)
			timerDuration = MaxSQLGCInterval
		} else {
			__antithesis_instrumentation__.Notify(492294)
		}
		__antithesis_instrumentation__.Notify(492269)
		timer.Reset(timerDuration)
	}
}

func (r schemaChangeGCResumer) OnFailOrCancel(context.Context, interface{}) error {
	__antithesis_instrumentation__.Notify(492295)
	return nil
}

func (r *schemaChangeGCResumer) isPermanentGCError(err error) bool {
	__antithesis_instrumentation__.Notify(492296)

	return sql.IsPermanentSchemaChangeError(err)
}

func init() {
	createResumerFn := func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
		return &schemaChangeGCResumer{
			jobID: job.ID(),
		}
	}
	jobs.RegisterConstructor(jobspb.TypeSchemaChangeGC, createResumerFn)
}
