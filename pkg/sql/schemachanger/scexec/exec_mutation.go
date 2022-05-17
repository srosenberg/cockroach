package scexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec/scmutationexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func executeDescriptorMutationOps(ctx context.Context, deps Dependencies, ops []scop.Op) error {
	__antithesis_instrumentation__.Notify(581426)

	mvs := newMutationVisitorState(deps.Catalog())
	v := scmutationexec.NewMutationVisitor(mvs, deps.Catalog(), deps.Catalog(), deps.Clock())
	for _, op := range ops {
		__antithesis_instrumentation__.Notify(581431)
		if err := op.(scop.MutationOp).Visit(ctx, v); err != nil {
			__antithesis_instrumentation__.Notify(581432)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581433)
		}
	}
	__antithesis_instrumentation__.Notify(581427)

	dbZoneConfigsToDelete, gcJobRecords := mvs.gcJobs.makeRecords(
		deps.TransactionalJobRegistry().MakeJobID,
	)
	if err := performBatchedCatalogWrites(
		ctx,
		mvs.descriptorsToDelete,
		dbZoneConfigsToDelete,
		mvs.checkedOutDescriptors,
		mvs.drainedNames,
		deps.Catalog(),
	); err != nil {
		__antithesis_instrumentation__.Notify(581434)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581435)
	}
	__antithesis_instrumentation__.Notify(581428)
	if err := logEvents(ctx, mvs, deps.EventLogger()); err != nil {
		__antithesis_instrumentation__.Notify(581436)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581437)
	}
	__antithesis_instrumentation__.Notify(581429)
	if err := updateDescriptorMetadata(
		ctx, mvs, deps.DescriptorMetadataUpdater(ctx),
	); err != nil {
		__antithesis_instrumentation__.Notify(581438)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581439)
	}
	__antithesis_instrumentation__.Notify(581430)
	return manageJobs(
		ctx,
		gcJobRecords,
		mvs.schemaChangerJob,
		mvs.schemaChangerJobUpdates,
		deps.TransactionalJobRegistry(),
	)
}

func performBatchedCatalogWrites(
	ctx context.Context,
	descriptorsToDelete catalog.DescriptorIDSet,
	dbZoneConfigsToDelete catalog.DescriptorIDSet,
	checkedOutDescriptors nstree.Map,
	drainedNames map[descpb.ID][]descpb.NameInfo,
	cat Catalog,
) error {
	__antithesis_instrumentation__.Notify(581440)
	b := cat.NewCatalogChangeBatcher()
	descriptorsToDelete.ForEach(func(id descpb.ID) {
		__antithesis_instrumentation__.Notify(581446)
		checkedOutDescriptors.Remove(id)
	})
	__antithesis_instrumentation__.Notify(581441)
	err := checkedOutDescriptors.IterateByID(func(entry catalog.NameEntry) error {
		__antithesis_instrumentation__.Notify(581447)
		return b.CreateOrUpdateDescriptor(ctx, entry.(catalog.MutableDescriptor))
	})
	__antithesis_instrumentation__.Notify(581442)
	if err != nil {
		__antithesis_instrumentation__.Notify(581448)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581449)
	}
	__antithesis_instrumentation__.Notify(581443)
	for _, id := range descriptorsToDelete.Ordered() {
		__antithesis_instrumentation__.Notify(581450)
		if err := b.DeleteDescriptor(ctx, id); err != nil {
			__antithesis_instrumentation__.Notify(581451)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581452)
		}
	}
	__antithesis_instrumentation__.Notify(581444)
	for id, drainedNames := range drainedNames {
		__antithesis_instrumentation__.Notify(581453)
		for _, name := range drainedNames {
			__antithesis_instrumentation__.Notify(581454)
			if err := b.DeleteName(ctx, name, id); err != nil {
				__antithesis_instrumentation__.Notify(581455)
				return err
			} else {
				__antithesis_instrumentation__.Notify(581456)
			}
		}
	}

	{
		__antithesis_instrumentation__.Notify(581457)
		var err error
		dbZoneConfigsToDelete.ForEach(func(id descpb.ID) {
			__antithesis_instrumentation__.Notify(581459)
			if err == nil {
				__antithesis_instrumentation__.Notify(581460)
				err = b.DeleteZoneConfig(ctx, id)
			} else {
				__antithesis_instrumentation__.Notify(581461)
			}
		})
		__antithesis_instrumentation__.Notify(581458)
		if err != nil {
			__antithesis_instrumentation__.Notify(581462)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581463)
		}
	}
	__antithesis_instrumentation__.Notify(581445)

	return b.ValidateAndRun(ctx)
}

func logEvents(ctx context.Context, mvs *mutationVisitorState, el EventLogger) error {
	__antithesis_instrumentation__.Notify(581464)
	statementIDs := make([]uint32, 0, len(mvs.eventsByStatement))
	for statementID := range mvs.eventsByStatement {
		__antithesis_instrumentation__.Notify(581468)
		statementIDs = append(statementIDs, statementID)
	}
	__antithesis_instrumentation__.Notify(581465)
	sort.Slice(statementIDs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(581469)
		return statementIDs[i] < statementIDs[j]
	})
	__antithesis_instrumentation__.Notify(581466)
	for _, statementID := range statementIDs {
		__antithesis_instrumentation__.Notify(581470)
		entries := eventLogEntriesForStatement(mvs.eventsByStatement[statementID])
		for _, e := range entries {
			__antithesis_instrumentation__.Notify(581471)

			if err := el.LogEvent(ctx, e.id, e.details, e.event); err != nil {
				__antithesis_instrumentation__.Notify(581472)
				return err
			} else {
				__antithesis_instrumentation__.Notify(581473)
			}
		}
	}
	__antithesis_instrumentation__.Notify(581467)
	return nil
}

func eventLogEntriesForStatement(statementEvents []eventPayload) (logEntries []eventPayload) {
	__antithesis_instrumentation__.Notify(581474)

	var dependentEvents = make(map[uint32][]eventPayload)
	var sourceEvents = make(map[uint32]eventPayload)

	for _, event := range statementEvents {
		__antithesis_instrumentation__.Notify(581479)
		dependentEvents[event.SubWorkID] = append(dependentEvents[event.SubWorkID], event)
	}
	__antithesis_instrumentation__.Notify(581475)

	orderedSubWorkID := make([]uint32, 0, len(dependentEvents))
	for subWorkID := range dependentEvents {
		__antithesis_instrumentation__.Notify(581480)
		elems := dependentEvents[subWorkID]
		sort.SliceStable(elems, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(581482)
			return elems[i].SourceElementID < elems[j].SourceElementID
		})
		__antithesis_instrumentation__.Notify(581481)
		sourceEvents[subWorkID] = elems[0]
		dependentEvents[subWorkID] = elems[1:]
		orderedSubWorkID = append(orderedSubWorkID, subWorkID)
	}
	__antithesis_instrumentation__.Notify(581476)

	sort.SliceStable(orderedSubWorkID, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(581483)
		return orderedSubWorkID[i] < orderedSubWorkID[j]
	})
	__antithesis_instrumentation__.Notify(581477)

	for _, subWorkID := range orderedSubWorkID {
		__antithesis_instrumentation__.Notify(581484)

		collectDependentViewNames := false
		collectDependentTables := false
		collectDependentSequences := false
		sourceEvent := sourceEvents[subWorkID]
		switch sourceEvent.event.(type) {
		case *eventpb.DropDatabase:
			__antithesis_instrumentation__.Notify(581488)

			collectDependentViewNames = true
			collectDependentTables = true
			collectDependentSequences = true
		case *eventpb.DropView, *eventpb.DropTable:
			__antithesis_instrumentation__.Notify(581489)

			collectDependentViewNames = true
		}
		__antithesis_instrumentation__.Notify(581485)
		var dependentObjects []string
		for _, dependentEvent := range dependentEvents[subWorkID] {
			__antithesis_instrumentation__.Notify(581490)
			switch ev := dependentEvent.event.(type) {
			case *eventpb.DropSequence:
				__antithesis_instrumentation__.Notify(581491)
				if collectDependentSequences {
					__antithesis_instrumentation__.Notify(581494)
					dependentObjects = append(dependentObjects, ev.SequenceName)
				} else {
					__antithesis_instrumentation__.Notify(581495)
				}
			case *eventpb.DropTable:
				__antithesis_instrumentation__.Notify(581492)
				if collectDependentTables {
					__antithesis_instrumentation__.Notify(581496)
					dependentObjects = append(dependentObjects, ev.TableName)
				} else {
					__antithesis_instrumentation__.Notify(581497)
				}
			case *eventpb.DropView:
				__antithesis_instrumentation__.Notify(581493)
				if collectDependentViewNames {
					__antithesis_instrumentation__.Notify(581498)
					dependentObjects = append(dependentObjects, ev.ViewName)
				} else {
					__antithesis_instrumentation__.Notify(581499)
				}
			}
		}
		__antithesis_instrumentation__.Notify(581486)

		switch ev := sourceEvent.event.(type) {
		case *eventpb.DropTable:
			__antithesis_instrumentation__.Notify(581500)
			ev.CascadeDroppedViews = dependentObjects
		case *eventpb.DropView:
			__antithesis_instrumentation__.Notify(581501)
			ev.CascadeDroppedViews = dependentObjects
		case *eventpb.DropDatabase:
			__antithesis_instrumentation__.Notify(581502)
			ev.DroppedSchemaObjects = dependentObjects
		}
		__antithesis_instrumentation__.Notify(581487)

		logEntries = append(logEntries, sourceEvent)
	}
	__antithesis_instrumentation__.Notify(581478)
	return logEntries
}

func updateDescriptorMetadata(
	ctx context.Context, mvs *mutationVisitorState, m DescriptorMetadataUpdater,
) error {
	__antithesis_instrumentation__.Notify(581503)
	for _, comment := range mvs.commentsToUpdate {
		__antithesis_instrumentation__.Notify(581509)
		if len(comment.comment) > 0 {
			__antithesis_instrumentation__.Notify(581510)
			if err := m.UpsertDescriptorComment(
				comment.id, comment.subID, comment.commentType, comment.comment); err != nil {
				__antithesis_instrumentation__.Notify(581511)
				return err
			} else {
				__antithesis_instrumentation__.Notify(581512)
			}
		} else {
			__antithesis_instrumentation__.Notify(581513)
			if err := m.DeleteDescriptorComment(
				comment.id, comment.subID, comment.commentType); err != nil {
				__antithesis_instrumentation__.Notify(581514)
				return err
			} else {
				__antithesis_instrumentation__.Notify(581515)
			}
		}
	}
	__antithesis_instrumentation__.Notify(581504)
	for _, comment := range mvs.constraintCommentsToUpdate {
		__antithesis_instrumentation__.Notify(581516)
		if len(comment.comment) > 0 {
			__antithesis_instrumentation__.Notify(581517)
			if err := m.UpsertConstraintComment(
				comment.tblID, comment.constraintID, comment.comment); err != nil {
				__antithesis_instrumentation__.Notify(581518)
				return err
			} else {
				__antithesis_instrumentation__.Notify(581519)
			}
		} else {
			__antithesis_instrumentation__.Notify(581520)
			if err := m.DeleteConstraintComment(
				comment.tblID, comment.constraintID); err != nil {
				__antithesis_instrumentation__.Notify(581521)
				return err
			} else {
				__antithesis_instrumentation__.Notify(581522)
			}
		}
	}
	__antithesis_instrumentation__.Notify(581505)
	if !mvs.tableCommentsToDelete.Empty() {
		__antithesis_instrumentation__.Notify(581523)
		if err := m.DeleteAllCommentsForTables(mvs.tableCommentsToDelete); err != nil {
			__antithesis_instrumentation__.Notify(581524)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581525)
		}
	} else {
		__antithesis_instrumentation__.Notify(581526)
	}
	__antithesis_instrumentation__.Notify(581506)
	for _, dbRoleSetting := range mvs.databaseRoleSettingsToDelete {
		__antithesis_instrumentation__.Notify(581527)
		err := m.DeleteDatabaseRoleSettings(ctx, dbRoleSetting.dbID)
		if err != nil {
			__antithesis_instrumentation__.Notify(581528)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581529)
		}
	}
	__antithesis_instrumentation__.Notify(581507)
	for _, scheduleID := range mvs.scheduleIDsToDelete {
		__antithesis_instrumentation__.Notify(581530)
		if err := m.DeleteSchedule(ctx, scheduleID); err != nil {
			__antithesis_instrumentation__.Notify(581531)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581532)
		}
	}
	__antithesis_instrumentation__.Notify(581508)
	return nil
}

func manageJobs(
	ctx context.Context,
	gcJobs []jobs.Record,
	scJob *jobs.Record,
	scJobUpdates map[jobspb.JobID]schemaChangerJobUpdate,
	jr TransactionalJobRegistry,
) error {
	__antithesis_instrumentation__.Notify(581533)

	for _, j := range gcJobs {
		__antithesis_instrumentation__.Notify(581537)
		if err := jr.CreateJob(ctx, j); err != nil {
			__antithesis_instrumentation__.Notify(581538)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581539)
		}
	}
	__antithesis_instrumentation__.Notify(581534)
	if scJob != nil {
		__antithesis_instrumentation__.Notify(581540)
		if err := jr.CreateJob(ctx, *scJob); err != nil {
			__antithesis_instrumentation__.Notify(581541)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581542)
		}
	} else {
		__antithesis_instrumentation__.Notify(581543)
	}
	__antithesis_instrumentation__.Notify(581535)
	for id, update := range scJobUpdates {
		__antithesis_instrumentation__.Notify(581544)
		if err := jr.UpdateSchemaChangeJob(ctx, id, func(
			md jobs.JobMetadata, updateProgress func(*jobspb.Progress), setNonCancelable func(),
		) error {
			__antithesis_instrumentation__.Notify(581545)
			progress := *md.Progress
			progress.RunningStatus = update.runningStatus
			updateProgress(&progress)
			if !md.Payload.Noncancelable && func() bool {
				__antithesis_instrumentation__.Notify(581547)
				return update.isNonCancelable == true
			}() == true {
				__antithesis_instrumentation__.Notify(581548)
				setNonCancelable()
			} else {
				__antithesis_instrumentation__.Notify(581549)
			}
			__antithesis_instrumentation__.Notify(581546)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(581550)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581551)
		}
	}
	__antithesis_instrumentation__.Notify(581536)
	return nil
}

type mutationVisitorState struct {
	c                            Catalog
	checkedOutDescriptors        nstree.Map
	drainedNames                 map[descpb.ID][]descpb.NameInfo
	descriptorsToDelete          catalog.DescriptorIDSet
	commentsToUpdate             []commentToUpdate
	tableCommentsToDelete        catalog.DescriptorIDSet
	constraintCommentsToUpdate   []constraintCommentToUpdate
	databaseRoleSettingsToDelete []databaseRoleSettingToDelete
	schemaChangerJob             *jobs.Record
	schemaChangerJobUpdates      map[jobspb.JobID]schemaChangerJobUpdate
	eventsByStatement            map[uint32][]eventPayload
	scheduleIDsToDelete          []int64

	gcJobs
}

type constraintCommentToUpdate struct {
	tblID        catid.DescID
	constraintID descpb.ConstraintID
	comment      string
}

type commentToUpdate struct {
	id          int64
	subID       int64
	commentType keys.CommentType
	comment     string
}

type databaseRoleSettingToDelete struct {
	dbID catid.DescID
}

type eventPayload struct {
	id descpb.ID
	scpb.TargetMetadata

	details eventpb.CommonSQLEventDetails
	event   eventpb.EventPayload
}

type schemaChangerJobUpdate struct {
	isNonCancelable bool
	runningStatus   string
}

func (mvs *mutationVisitorState) UpdateSchemaChangerJob(
	jobID jobspb.JobID, isNonCancelable bool, runningStatus string,
) error {
	__antithesis_instrumentation__.Notify(581552)
	if mvs.schemaChangerJobUpdates == nil {
		__antithesis_instrumentation__.Notify(581554)
		mvs.schemaChangerJobUpdates = make(map[jobspb.JobID]schemaChangerJobUpdate)
	} else {
		__antithesis_instrumentation__.Notify(581555)
		if _, exists := mvs.schemaChangerJobUpdates[jobID]; exists {
			__antithesis_instrumentation__.Notify(581556)
			return errors.AssertionFailedf("cannot update job %d more than once", jobID)
		} else {
			__antithesis_instrumentation__.Notify(581557)
		}
	}
	__antithesis_instrumentation__.Notify(581553)
	mvs.schemaChangerJobUpdates[jobID] = schemaChangerJobUpdate{
		isNonCancelable: isNonCancelable,
		runningStatus:   runningStatus,
	}
	return nil
}

func newMutationVisitorState(c Catalog) *mutationVisitorState {
	__antithesis_instrumentation__.Notify(581558)
	return &mutationVisitorState{
		c:                 c,
		drainedNames:      make(map[descpb.ID][]descpb.NameInfo),
		eventsByStatement: make(map[uint32][]eventPayload),
	}
}

var _ scmutationexec.MutationVisitorStateUpdater = (*mutationVisitorState)(nil)

func (mvs *mutationVisitorState) GetDescriptor(
	ctx context.Context, id descpb.ID,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(581559)
	if entry := mvs.checkedOutDescriptors.GetByID(id); entry != nil {
		__antithesis_instrumentation__.Notify(581562)
		return entry.(catalog.Descriptor), nil
	} else {
		__antithesis_instrumentation__.Notify(581563)
	}
	__antithesis_instrumentation__.Notify(581560)
	descs, err := mvs.c.MustReadImmutableDescriptors(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(581564)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581565)
	}
	__antithesis_instrumentation__.Notify(581561)
	return descs[0], nil
}

func (mvs *mutationVisitorState) CheckOutDescriptor(
	ctx context.Context, id descpb.ID,
) (catalog.MutableDescriptor, error) {
	__antithesis_instrumentation__.Notify(581566)
	entry := mvs.checkedOutDescriptors.GetByID(id)
	if entry != nil {
		__antithesis_instrumentation__.Notify(581569)
		return entry.(catalog.MutableDescriptor), nil
	} else {
		__antithesis_instrumentation__.Notify(581570)
	}
	__antithesis_instrumentation__.Notify(581567)
	mut, err := mvs.c.MustReadMutableDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(581571)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581572)
	}
	__antithesis_instrumentation__.Notify(581568)
	mut.MaybeIncrementVersion()
	mvs.checkedOutDescriptors.Upsert(mut)
	return mut, nil
}

func (mvs *mutationVisitorState) MaybeCheckedOutDescriptor(id descpb.ID) catalog.Descriptor {
	__antithesis_instrumentation__.Notify(581573)
	entry := mvs.checkedOutDescriptors.GetByID(id)
	if entry == nil {
		__antithesis_instrumentation__.Notify(581575)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(581576)
	}
	__antithesis_instrumentation__.Notify(581574)
	return entry.(catalog.Descriptor)
}

func (mvs *mutationVisitorState) DeleteDescriptor(id descpb.ID) {
	__antithesis_instrumentation__.Notify(581577)
	mvs.descriptorsToDelete.Add(id)
}

func (mvs *mutationVisitorState) DeleteAllTableComments(id descpb.ID) {
	__antithesis_instrumentation__.Notify(581578)
	mvs.tableCommentsToDelete.Add(id)
}

func (mvs *mutationVisitorState) DeleteComment(
	id descpb.ID, subID int, commentType keys.CommentType,
) {
	__antithesis_instrumentation__.Notify(581579)
	mvs.commentsToUpdate = append(mvs.commentsToUpdate,
		commentToUpdate{
			id:          int64(id),
			subID:       int64(subID),
			commentType: commentType,
		})
}

func (mvs *mutationVisitorState) DeleteConstraintComment(
	ctx context.Context, tblID descpb.ID, constraintID descpb.ConstraintID,
) error {
	__antithesis_instrumentation__.Notify(581580)
	mvs.constraintCommentsToUpdate = append(mvs.constraintCommentsToUpdate,
		constraintCommentToUpdate{
			tblID:        tblID,
			constraintID: constraintID,
		})
	return nil
}

func (mvs *mutationVisitorState) DeleteDatabaseRoleSettings(
	ctx context.Context, dbID descpb.ID,
) error {
	__antithesis_instrumentation__.Notify(581581)
	mvs.databaseRoleSettingsToDelete = append(mvs.databaseRoleSettingsToDelete,
		databaseRoleSettingToDelete{
			dbID: dbID,
		})
	return nil
}

func (mvs *mutationVisitorState) DeleteSchedule(scheduleID int64) {
	__antithesis_instrumentation__.Notify(581582)
	mvs.scheduleIDsToDelete = append(mvs.scheduleIDsToDelete, scheduleID)
}

func (mvs *mutationVisitorState) AddDrainedName(id descpb.ID, nameInfo descpb.NameInfo) {
	__antithesis_instrumentation__.Notify(581583)
	mvs.drainedNames[id] = append(mvs.drainedNames[id], nameInfo)
}

func (mvs *mutationVisitorState) AddNewSchemaChangerJob(
	jobID jobspb.JobID,
	stmts []scpb.Statement,
	isNonCancelable bool,
	auth scpb.Authorization,
	descriptorIDs descpb.IDs,
	runningStatus string,
) error {
	__antithesis_instrumentation__.Notify(581584)
	if mvs.schemaChangerJob != nil {
		__antithesis_instrumentation__.Notify(581586)
		return errors.AssertionFailedf("cannot create more than one new schema change job")
	} else {
		__antithesis_instrumentation__.Notify(581587)
	}
	__antithesis_instrumentation__.Notify(581585)
	mvs.schemaChangerJob = MakeDeclarativeSchemaChangeJobRecord(
		jobID,
		stmts,
		isNonCancelable,
		auth,
		descriptorIDs,
		runningStatus,
	)
	return nil
}

func MakeDeclarativeSchemaChangeJobRecord(
	jobID jobspb.JobID,
	stmts []scpb.Statement,
	isNonCancelable bool,
	auth scpb.Authorization,
	descriptorIDs descpb.IDs,
	runningStatus string,
) *jobs.Record {
	__antithesis_instrumentation__.Notify(581588)
	stmtStrs := make([]string, len(stmts))
	for i, stmt := range stmts {
		__antithesis_instrumentation__.Notify(581590)

		stmtStrs[i] = redact.RedactableString(stmt.RedactedStatement).StripMarkers()
	}
	__antithesis_instrumentation__.Notify(581589)

	description := strings.Join(stmtStrs, "; ")
	rec := &jobs.Record{
		JobID:         jobID,
		Description:   description,
		Statements:    stmtStrs,
		Username:      security.MakeSQLUsernameFromPreNormalizedString(auth.UserName),
		DescriptorIDs: descriptorIDs,
		Details:       jobspb.NewSchemaChangeDetails{},
		Progress:      jobspb.NewSchemaChangeProgress{},
		RunningStatus: jobs.RunningStatus(runningStatus),
		NonCancelable: isNonCancelable,
	}
	return rec
}

func (mvs *mutationVisitorState) EnqueueEvent(
	id descpb.ID,
	metadata scpb.TargetMetadata,
	details eventpb.CommonSQLEventDetails,
	event eventpb.EventPayload,
) error {
	__antithesis_instrumentation__.Notify(581591)
	mvs.eventsByStatement[metadata.StatementID] = append(
		mvs.eventsByStatement[metadata.StatementID],
		eventPayload{
			id:             id,
			event:          event,
			TargetMetadata: metadata,
			details:        details,
		},
	)
	return nil
}
