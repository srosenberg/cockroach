package gcjob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func markTableGCed(
	ctx context.Context, tableID descpb.ID, progress *jobspb.SchemaChangeGCProgress,
) {
	__antithesis_instrumentation__.Notify(492297)
	for i := range progress.Tables {
		__antithesis_instrumentation__.Notify(492298)
		tableProgress := &progress.Tables[i]
		if tableProgress.ID == tableID {
			__antithesis_instrumentation__.Notify(492299)
			tableProgress.Status = jobspb.SchemaChangeGCProgress_DELETED
			if log.V(2) {
				__antithesis_instrumentation__.Notify(492300)
				log.Infof(ctx, "determined table %d is GC'd", tableID)
			} else {
				__antithesis_instrumentation__.Notify(492301)
			}
		} else {
			__antithesis_instrumentation__.Notify(492302)
		}
	}
}

func markIndexGCed(
	ctx context.Context,
	garbageCollectedIndexID descpb.IndexID,
	progress *jobspb.SchemaChangeGCProgress,
) {
	__antithesis_instrumentation__.Notify(492303)

	for i := range progress.Indexes {
		__antithesis_instrumentation__.Notify(492304)
		indexToUpdate := &progress.Indexes[i]
		if indexToUpdate.IndexID == garbageCollectedIndexID {
			__antithesis_instrumentation__.Notify(492305)
			indexToUpdate.Status = jobspb.SchemaChangeGCProgress_DELETED
			log.Infof(ctx, "marked index %d as GC'd", garbageCollectedIndexID)
		} else {
			__antithesis_instrumentation__.Notify(492306)
		}
	}
}

func initDetailsAndProgress(
	ctx context.Context, execCfg *sql.ExecutorConfig, jobID jobspb.JobID,
) (*jobspb.SchemaChangeGCDetails, *jobspb.SchemaChangeGCProgress, error) {
	__antithesis_instrumentation__.Notify(492307)
	var details jobspb.SchemaChangeGCDetails
	var progress *jobspb.SchemaChangeGCProgress
	var job *jobs.Job
	if err := execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(492311)
		var err error
		job, err = execCfg.JobRegistry.LoadJobWithTxn(ctx, jobID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(492313)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492314)
		}
		__antithesis_instrumentation__.Notify(492312)
		details = job.Details().(jobspb.SchemaChangeGCDetails)
		jobProgress := job.Progress()
		progress = jobProgress.GetSchemaChangeGC()
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(492315)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(492316)
	}
	__antithesis_instrumentation__.Notify(492308)
	if err := validateDetails(&details); err != nil {
		__antithesis_instrumentation__.Notify(492317)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(492318)
	}
	__antithesis_instrumentation__.Notify(492309)
	if err := initializeProgress(ctx, execCfg, jobID, &details, progress); err != nil {
		__antithesis_instrumentation__.Notify(492319)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(492320)
	}
	__antithesis_instrumentation__.Notify(492310)
	return &details, progress, nil
}

func initializeProgress(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	jobID jobspb.JobID,
	details *jobspb.SchemaChangeGCDetails,
	progress *jobspb.SchemaChangeGCProgress,
) error {
	__antithesis_instrumentation__.Notify(492321)
	var update bool
	if details.Tenant != nil && func() bool {
		__antithesis_instrumentation__.Notify(492324)
		return progress.Tenant == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(492325)
		progress.Tenant = &jobspb.SchemaChangeGCProgress_TenantProgress{
			Status: jobspb.SchemaChangeGCProgress_WAITING_FOR_GC,
		}
		update = true
	} else {
		__antithesis_instrumentation__.Notify(492326)
		if len(progress.Tables) != len(details.Tables) || func() bool {
			__antithesis_instrumentation__.Notify(492327)
			return len(progress.Indexes) != len(details.Indexes) == true
		}() == true {
			__antithesis_instrumentation__.Notify(492328)
			update = true
			for _, table := range details.Tables {
				__antithesis_instrumentation__.Notify(492330)
				progress.Tables = append(progress.Tables, jobspb.SchemaChangeGCProgress_TableProgress{ID: table.ID})
			}
			__antithesis_instrumentation__.Notify(492329)
			for _, index := range details.Indexes {
				__antithesis_instrumentation__.Notify(492331)
				progress.Indexes = append(progress.Indexes, jobspb.SchemaChangeGCProgress_IndexProgress{IndexID: index.IndexID})
			}
		} else {
			__antithesis_instrumentation__.Notify(492332)
		}
	}
	__antithesis_instrumentation__.Notify(492322)

	if update {
		__antithesis_instrumentation__.Notify(492333)
		if err := execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(492334)
			job, err := execCfg.JobRegistry.LoadJobWithTxn(ctx, jobID, txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(492336)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492337)
			}
			__antithesis_instrumentation__.Notify(492335)
			return job.SetProgress(ctx, txn, *progress)
		}); err != nil {
			__antithesis_instrumentation__.Notify(492338)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492339)
		}
	} else {
		__antithesis_instrumentation__.Notify(492340)
	}
	__antithesis_instrumentation__.Notify(492323)
	return nil
}

func isDoneGC(progress *jobspb.SchemaChangeGCProgress) bool {
	__antithesis_instrumentation__.Notify(492341)
	for _, index := range progress.Indexes {
		__antithesis_instrumentation__.Notify(492345)
		if index.Status != jobspb.SchemaChangeGCProgress_DELETED {
			__antithesis_instrumentation__.Notify(492346)
			return false
		} else {
			__antithesis_instrumentation__.Notify(492347)
		}
	}
	__antithesis_instrumentation__.Notify(492342)
	for _, table := range progress.Tables {
		__antithesis_instrumentation__.Notify(492348)
		if table.Status != jobspb.SchemaChangeGCProgress_DELETED {
			__antithesis_instrumentation__.Notify(492349)
			return false
		} else {
			__antithesis_instrumentation__.Notify(492350)
		}
	}
	__antithesis_instrumentation__.Notify(492343)
	if progress.Tenant != nil && func() bool {
		__antithesis_instrumentation__.Notify(492351)
		return progress.Tenant.Status != jobspb.SchemaChangeGCProgress_DELETED == true
	}() == true {
		__antithesis_instrumentation__.Notify(492352)
		return false
	} else {
		__antithesis_instrumentation__.Notify(492353)
	}
	__antithesis_instrumentation__.Notify(492344)

	return true
}

func runningStatusGC(progress *jobspb.SchemaChangeGCProgress) jobs.RunningStatus {
	__antithesis_instrumentation__.Notify(492354)
	tableIDs := make([]string, 0, len(progress.Tables))
	indexIDs := make([]string, 0, len(progress.Indexes))
	for _, table := range progress.Tables {
		__antithesis_instrumentation__.Notify(492360)
		if table.Status == jobspb.SchemaChangeGCProgress_DELETING {
			__antithesis_instrumentation__.Notify(492361)
			tableIDs = append(tableIDs, strconv.Itoa(int(table.ID)))
		} else {
			__antithesis_instrumentation__.Notify(492362)
		}
	}
	__antithesis_instrumentation__.Notify(492355)
	for _, index := range progress.Indexes {
		__antithesis_instrumentation__.Notify(492363)
		if index.Status == jobspb.SchemaChangeGCProgress_DELETING {
			__antithesis_instrumentation__.Notify(492364)
			indexIDs = append(indexIDs, strconv.Itoa(int(index.IndexID)))
		} else {
			__antithesis_instrumentation__.Notify(492365)
		}
	}
	__antithesis_instrumentation__.Notify(492356)

	var b strings.Builder
	b.WriteString("performing garbage collection on")
	var flag bool
	if progress.Tenant != nil && func() bool {
		__antithesis_instrumentation__.Notify(492366)
		return progress.Tenant.Status == jobspb.SchemaChangeGCProgress_DELETING == true
	}() == true {
		__antithesis_instrumentation__.Notify(492367)
		b.WriteString(" tenant")
		flag = true
	} else {
		__antithesis_instrumentation__.Notify(492368)
	}
	__antithesis_instrumentation__.Notify(492357)

	for _, s := range []struct {
		ids      []string
		singular string
		plural   string
	}{
		{tableIDs, "table", "tables"},
		{indexIDs, "index", "indexes"},
	} {
		__antithesis_instrumentation__.Notify(492369)
		if len(s.ids) == 0 {
			__antithesis_instrumentation__.Notify(492373)
			continue
		} else {
			__antithesis_instrumentation__.Notify(492374)
		}
		__antithesis_instrumentation__.Notify(492370)
		if flag {
			__antithesis_instrumentation__.Notify(492375)
			b.WriteRune(';')
		} else {
			__antithesis_instrumentation__.Notify(492376)
		}
		__antithesis_instrumentation__.Notify(492371)
		b.WriteRune(' ')
		switch len(s.ids) {
		case 1:
			__antithesis_instrumentation__.Notify(492377)

			b.WriteString(s.singular)
			b.WriteRune(' ')
			b.WriteString(s.ids[0])
		case 2, 3, 4, 5:
			__antithesis_instrumentation__.Notify(492378)

			b.WriteString(s.plural)
			b.WriteRune(' ')
			b.WriteString(strings.Join(s.ids, ", "))
		default:
			__antithesis_instrumentation__.Notify(492379)

			b.WriteString(strconv.Itoa(len(s.ids)))
			b.WriteRune(' ')
			b.WriteString(s.plural)
		}
		__antithesis_instrumentation__.Notify(492372)
		flag = true
	}
	__antithesis_instrumentation__.Notify(492358)

	if !flag {
		__antithesis_instrumentation__.Notify(492380)

		return sql.RunningStatusWaitingGC
	} else {
		__antithesis_instrumentation__.Notify(492381)
	}
	__antithesis_instrumentation__.Notify(492359)
	return jobs.RunningStatus(b.String())
}

func getAllTablesWaitingForGC(
	details *jobspb.SchemaChangeGCDetails, progress *jobspb.SchemaChangeGCProgress,
) []descpb.ID {
	__antithesis_instrumentation__.Notify(492382)
	allRemainingTableIDs := make([]descpb.ID, 0, len(progress.Tables))
	if len(details.Indexes) > 0 {
		__antithesis_instrumentation__.Notify(492385)
		allRemainingTableIDs = append(allRemainingTableIDs, details.ParentID)
	} else {
		__antithesis_instrumentation__.Notify(492386)
	}
	__antithesis_instrumentation__.Notify(492383)
	for _, table := range progress.Tables {
		__antithesis_instrumentation__.Notify(492387)
		if table.Status != jobspb.SchemaChangeGCProgress_DELETED {
			__antithesis_instrumentation__.Notify(492388)
			allRemainingTableIDs = append(allRemainingTableIDs, table.ID)
		} else {
			__antithesis_instrumentation__.Notify(492389)
		}
	}
	__antithesis_instrumentation__.Notify(492384)

	return allRemainingTableIDs
}

func validateDetails(details *jobspb.SchemaChangeGCDetails) error {
	__antithesis_instrumentation__.Notify(492390)
	if details.Tenant != nil && func() bool {
		__antithesis_instrumentation__.Notify(492393)
		return (len(details.Tables) > 0 || func() bool {
			__antithesis_instrumentation__.Notify(492394)
			return len(details.Indexes) > 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(492395)
		return errors.AssertionFailedf(
			"Either field Tenant is set or any of Tables or Indexes: %+v", *details,
		)
	} else {
		__antithesis_instrumentation__.Notify(492396)
	}
	__antithesis_instrumentation__.Notify(492391)
	if len(details.Indexes) > 0 {
		__antithesis_instrumentation__.Notify(492397)
		if details.ParentID == descpb.InvalidID {
			__antithesis_instrumentation__.Notify(492398)
			return errors.Errorf("must provide a parentID when dropping an index")
		} else {
			__antithesis_instrumentation__.Notify(492399)
		}
	} else {
		__antithesis_instrumentation__.Notify(492400)
	}
	__antithesis_instrumentation__.Notify(492392)
	return nil
}

func persistProgress(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	jobID jobspb.JobID,
	progress *jobspb.SchemaChangeGCProgress,
	runningStatus jobs.RunningStatus,
) {
	__antithesis_instrumentation__.Notify(492401)
	if err := execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(492402)
		job, err := execCfg.JobRegistry.LoadJobWithTxn(ctx, jobID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(492407)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492408)
		}
		__antithesis_instrumentation__.Notify(492403)
		if err := job.SetProgress(ctx, txn, *progress); err != nil {
			__antithesis_instrumentation__.Notify(492409)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492410)
		}
		__antithesis_instrumentation__.Notify(492404)
		log.Infof(ctx, "updated progress payload: %+v", progress)
		err = job.RunningStatus(ctx, txn, func(_ context.Context, _ jobspb.Details) (jobs.RunningStatus, error) {
			__antithesis_instrumentation__.Notify(492411)
			return runningStatus, nil
		})
		__antithesis_instrumentation__.Notify(492405)
		if err != nil {
			__antithesis_instrumentation__.Notify(492412)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492413)
		}
		__antithesis_instrumentation__.Notify(492406)
		log.Infof(ctx, "updated running status: %+v", runningStatus)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(492414)
		log.Warningf(ctx, "failed to update job's progress payload or running status err: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(492415)
	}
}

func getDropTimes(
	details *jobspb.SchemaChangeGCDetails,
) (map[descpb.ID]int64, map[descpb.IndexID]int64) {
	__antithesis_instrumentation__.Notify(492416)
	tableDropTimes := make(map[descpb.ID]int64)
	for _, table := range details.Tables {
		__antithesis_instrumentation__.Notify(492419)
		tableDropTimes[table.ID] = table.DropTime
	}
	__antithesis_instrumentation__.Notify(492417)
	indexDropTimes := make(map[descpb.IndexID]int64)
	for _, index := range details.Indexes {
		__antithesis_instrumentation__.Notify(492420)
		indexDropTimes[index.IndexID] = index.DropTime
	}
	__antithesis_instrumentation__.Notify(492418)
	return tableDropTimes, indexDropTimes
}
