package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/faketreeeval"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

const (
	RunningStatusWaitingGC jobs.RunningStatus = "waiting for GC TTL"

	RunningStatusDeleteOnly jobs.RunningStatus = "waiting in DELETE-ONLY"

	RunningStatusDeleteAndWriteOnly jobs.RunningStatus = "waiting in DELETE-AND-WRITE_ONLY"

	RunningStatusMerging jobs.RunningStatus = "waiting in MERGING"

	RunningStatusBackfill jobs.RunningStatus = "populating schema"

	RunningStatusValidation jobs.RunningStatus = "validating schema"
)

type SchemaChanger struct {
	descID            descpb.ID
	mutationID        descpb.MutationID
	droppedDatabaseID descpb.ID
	sqlInstanceID     base.SQLInstanceID
	db                *kv.DB
	leaseMgr          *lease.Manager

	metrics *SchemaChangerMetrics

	testingKnobs   *SchemaChangerTestingKnobs
	distSQLPlanner *DistSQLPlanner
	jobRegistry    *jobs.Registry

	job *jobs.Job

	rangeDescriptorCache *rangecache.RangeCache
	clock                *hlc.Clock
	settings             *cluster.Settings
	execCfg              *ExecutorConfig
	ieFactory            sqlutil.SessionBoundInternalExecutorFactory

	mvccCompliantAddIndex bool
}

func NewSchemaChangerForTesting(
	tableID descpb.ID,
	mutationID descpb.MutationID,
	sqlInstanceID base.SQLInstanceID,
	db *kv.DB,
	leaseMgr *lease.Manager,
	jobRegistry *jobs.Registry,
	execCfg *ExecutorConfig,
	settings *cluster.Settings,
) SchemaChanger {
	__antithesis_instrumentation__.Notify(577065)
	return SchemaChanger{
		descID:        tableID,
		mutationID:    mutationID,
		sqlInstanceID: sqlInstanceID,
		db:            db,
		leaseMgr:      leaseMgr,
		jobRegistry:   jobRegistry,
		settings:      settings,
		execCfg:       execCfg,

		ieFactory: func(
			ctx context.Context, sd *sessiondata.SessionData,
		) sqlutil.InternalExecutor {
			__antithesis_instrumentation__.Notify(577066)
			return execCfg.InternalExecutor
		},
		metrics:        NewSchemaChangerMetrics(),
		clock:          db.Clock(),
		distSQLPlanner: execCfg.DistSQLPlanner,
		testingKnobs:   &SchemaChangerTestingKnobs{},
	}
}

func IsConstraintError(err error) bool {
	__antithesis_instrumentation__.Notify(577067)
	pgCode := pgerror.GetPGCode(err)
	return pgCode == pgcode.CheckViolation || func() bool {
		__antithesis_instrumentation__.Notify(577068)
		return pgCode == pgcode.UniqueViolation == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(577069)
		return pgCode == pgcode.ForeignKeyViolation == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(577070)
		return pgCode == pgcode.NotNullViolation == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(577071)
		return pgCode == pgcode.IntegrityConstraintViolation == true
	}() == true
}

func IsPermanentSchemaChangeError(err error) bool {
	__antithesis_instrumentation__.Notify(577072)
	if err == nil {
		__antithesis_instrumentation__.Notify(577081)
		return false
	} else {
		__antithesis_instrumentation__.Notify(577082)
	}
	__antithesis_instrumentation__.Notify(577073)

	if grpcutil.IsClosedConnection(err) {
		__antithesis_instrumentation__.Notify(577083)
		return false
	} else {
		__antithesis_instrumentation__.Notify(577084)
	}
	__antithesis_instrumentation__.Notify(577074)

	if errors.HasType(err, (*roachpb.BatchTimestampBeforeGCError)(nil)) {
		__antithesis_instrumentation__.Notify(577085)
		return false
	} else {
		__antithesis_instrumentation__.Notify(577086)
	}
	__antithesis_instrumentation__.Notify(577075)

	if hlc.IsUntrustworthyRemoteWallTimeError(err) {
		__antithesis_instrumentation__.Notify(577087)
		return false
	} else {
		__antithesis_instrumentation__.Notify(577088)
	}
	__antithesis_instrumentation__.Notify(577076)

	if pgerror.IsSQLRetryableError(err) {
		__antithesis_instrumentation__.Notify(577089)
		return false
	} else {
		__antithesis_instrumentation__.Notify(577090)
	}
	__antithesis_instrumentation__.Notify(577077)

	if errors.IsAny(err,
		context.Canceled,
		context.DeadlineExceeded,
		errSchemaChangeNotFirstInLine,
		errTableVersionMismatchSentinel,
	) {
		__antithesis_instrumentation__.Notify(577091)
		return false
	} else {
		__antithesis_instrumentation__.Notify(577092)
	}
	__antithesis_instrumentation__.Notify(577078)

	if flowinfra.IsFlowRetryableError(err) {
		__antithesis_instrumentation__.Notify(577093)
		return false
	} else {
		__antithesis_instrumentation__.Notify(577094)
	}
	__antithesis_instrumentation__.Notify(577079)

	switch pgerror.GetPGCode(err) {
	case pgcode.SerializationFailure, pgcode.InternalConnectionFailure:
		__antithesis_instrumentation__.Notify(577095)
		return false

	case pgcode.Internal, pgcode.RangeUnavailable:
		__antithesis_instrumentation__.Notify(577096)
		if strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
			__antithesis_instrumentation__.Notify(577098)
			return false
		} else {
			__antithesis_instrumentation__.Notify(577099)
		}
	default:
		__antithesis_instrumentation__.Notify(577097)
	}
	__antithesis_instrumentation__.Notify(577080)

	return true
}

var errSchemaChangeNotFirstInLine = errors.Newf("schema change not first in line")

type errTableVersionMismatch struct {
	version  descpb.DescriptorVersion
	expected descpb.DescriptorVersion
}

var errTableVersionMismatchSentinel = errTableVersionMismatch{}

func makeErrTableVersionMismatch(version, expected descpb.DescriptorVersion) error {
	__antithesis_instrumentation__.Notify(577100)
	return errors.Mark(errors.WithStack(errTableVersionMismatch{
		version:  version,
		expected: expected,
	}), errTableVersionMismatchSentinel)
}

func (e errTableVersionMismatch) Error() string {
	__antithesis_instrumentation__.Notify(577101)
	return fmt.Sprintf("table version mismatch: %d, expected: %d", e.version, e.expected)
}

func (sc *SchemaChanger) refreshMaterializedView(
	ctx context.Context, table catalog.TableDescriptor, refresh catalog.MaterializedViewRefresh,
) error {
	__antithesis_instrumentation__.Notify(577102)

	if !refresh.ShouldBackfill() {
		__antithesis_instrumentation__.Notify(577104)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(577105)
	}
	__antithesis_instrumentation__.Notify(577103)

	tableToRefresh := refresh.TableWithNewIndexes(table)
	return sc.backfillQueryIntoTable(ctx, tableToRefresh, table.GetViewQuery(), refresh.AsOf(), "refreshView")
}

func (sc *SchemaChanger) backfillQueryIntoTable(
	ctx context.Context, table catalog.TableDescriptor, query string, ts hlc.Timestamp, desc string,
) error {
	__antithesis_instrumentation__.Notify(577106)
	if fn := sc.testingKnobs.RunBeforeQueryBackfill; fn != nil {
		__antithesis_instrumentation__.Notify(577108)
		if err := fn(); err != nil {
			__antithesis_instrumentation__.Notify(577109)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577110)
		}
	} else {
		__antithesis_instrumentation__.Notify(577111)
	}
	__antithesis_instrumentation__.Notify(577107)

	return sc.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(577112)
		if err := txn.SetFixedTimestamp(ctx, ts); err != nil {
			__antithesis_instrumentation__.Notify(577119)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577120)
		}
		__antithesis_instrumentation__.Notify(577113)

		p, cleanup := NewInternalPlanner(
			desc,
			txn,
			security.RootUserName(),
			&MemoryMetrics{},
			sc.execCfg,
			sessiondatapb.SessionData{},
		)

		defer cleanup()
		localPlanner := p.(*planner)
		stmt, err := parser.ParseOne(query)
		if err != nil {
			__antithesis_instrumentation__.Notify(577121)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577122)
		}
		__antithesis_instrumentation__.Notify(577114)

		localPlanner.stmt = makeStatement(stmt, ClusterWideID{})
		localPlanner.optPlanningCtx.init(localPlanner)

		localPlanner.runWithOptions(resolveFlags{skipCache: true}, func() {
			__antithesis_instrumentation__.Notify(577123)
			err = localPlanner.makeOptimizerPlan(ctx)
		})
		__antithesis_instrumentation__.Notify(577115)

		if err != nil {
			__antithesis_instrumentation__.Notify(577124)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577125)
		}
		__antithesis_instrumentation__.Notify(577116)
		defer localPlanner.curPlan.close(ctx)

		res := roachpb.BulkOpSummary{}
		rw := NewCallbackResultWriter(func(ctx context.Context, row tree.Datums) error {
			__antithesis_instrumentation__.Notify(577126)

			var counts roachpb.BulkOpSummary
			if err := protoutil.Unmarshal([]byte(*row[0].(*tree.DBytes)), &counts); err != nil {
				__antithesis_instrumentation__.Notify(577128)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577129)
			}
			__antithesis_instrumentation__.Notify(577127)
			res.Add(counts)
			return nil
		})
		__antithesis_instrumentation__.Notify(577117)
		recv := MakeDistSQLReceiver(
			ctx,
			rw,
			tree.Rows,
			sc.execCfg.RangeDescriptorCache,
			txn,
			sc.clock,

			&SessionTracing{},
			sc.execCfg.ContentionRegistry,
			nil,
		)
		defer recv.Release()

		var planAndRunErr error
		localPlanner.runWithOptions(resolveFlags{skipCache: true}, func() {
			__antithesis_instrumentation__.Notify(577130)

			if len(localPlanner.curPlan.subqueryPlans) != 0 {
				__antithesis_instrumentation__.Notify(577132)

				subqueryResultMemAcc := localPlanner.EvalContext().Mon.MakeBoundAccount()
				defer subqueryResultMemAcc.Close(ctx)
				if !sc.distSQLPlanner.PlanAndRunSubqueries(
					ctx, localPlanner, localPlanner.ExtendedEvalContextCopy,
					localPlanner.curPlan.subqueryPlans, recv, &subqueryResultMemAcc,
				) {
					__antithesis_instrumentation__.Notify(577133)
					if planAndRunErr = rw.Err(); planAndRunErr != nil {
						__antithesis_instrumentation__.Notify(577134)
						return
					} else {
						__antithesis_instrumentation__.Notify(577135)
					}
				} else {
					__antithesis_instrumentation__.Notify(577136)
				}
			} else {
				__antithesis_instrumentation__.Notify(577137)
			}
			__antithesis_instrumentation__.Notify(577131)

			isLocal := !getPlanDistribution(
				ctx, localPlanner, localPlanner.execCfg.NodeID,
				localPlanner.extendedEvalCtx.SessionData().DistSQLMode,
				localPlanner.curPlan.main,
			).WillDistribute()
			out := execinfrapb.ProcessorCoreUnion{BulkRowWriter: &execinfrapb.BulkRowWriterSpec{
				Table: *table.TableDesc(),
			}}

			PlanAndRunCTAS(ctx, sc.distSQLPlanner, localPlanner,
				txn, isLocal, localPlanner.curPlan.main, out, recv)
			if planAndRunErr = rw.Err(); planAndRunErr != nil {
				__antithesis_instrumentation__.Notify(577138)
				return
			} else {
				__antithesis_instrumentation__.Notify(577139)
			}
		})
		__antithesis_instrumentation__.Notify(577118)

		return planAndRunErr
	})
}

func (sc *SchemaChanger) maybeBackfillCreateTableAs(
	ctx context.Context, table catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(577140)
	if !(table.Adding() && func() bool {
		__antithesis_instrumentation__.Notify(577142)
		return table.IsAs() == true
	}() == true) {
		__antithesis_instrumentation__.Notify(577143)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(577144)
	}
	__antithesis_instrumentation__.Notify(577141)
	log.Infof(ctx, "starting backfill for CREATE TABLE AS with query %q", table.GetCreateQuery())

	return sc.backfillQueryIntoTable(ctx, table, table.GetCreateQuery(), table.GetCreateAsOfTime(), "ctasBackfill")
}

func (sc *SchemaChanger) maybeUpdateScheduledJobsForRowLevelTTL(
	ctx context.Context, tableDesc catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(577145)

	if tableDesc.Dropped() && func() bool {
		__antithesis_instrumentation__.Notify(577147)
		return tableDesc.GetRowLevelTTL() != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(577148)
		if err := sc.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(577149)
			scheduleID := tableDesc.GetRowLevelTTL().ScheduleID
			if scheduleID > 0 {
				__antithesis_instrumentation__.Notify(577151)
				log.Infof(ctx, "dropping TTL schedule %d", scheduleID)
				return DeleteSchedule(ctx, sc.execCfg, txn, scheduleID)
			} else {
				__antithesis_instrumentation__.Notify(577152)
			}
			__antithesis_instrumentation__.Notify(577150)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(577153)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577154)
		}
	} else {
		__antithesis_instrumentation__.Notify(577155)
	}
	__antithesis_instrumentation__.Notify(577146)
	return nil
}

func (sc *SchemaChanger) maybeBackfillMaterializedView(
	ctx context.Context, table catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(577156)
	if !(table.Adding() && func() bool {
		__antithesis_instrumentation__.Notify(577158)
		return table.MaterializedView() == true
	}() == true) {
		__antithesis_instrumentation__.Notify(577159)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(577160)
	}
	__antithesis_instrumentation__.Notify(577157)
	log.Infof(ctx, "starting backfill for CREATE MATERIALIZED VIEW with query %q", table.GetViewQuery())

	return sc.backfillQueryIntoTable(ctx, table, table.GetViewQuery(), table.GetCreateAsOfTime(), "materializedViewBackfill")
}

func (sc *SchemaChanger) maybeMakeAddTablePublic(
	ctx context.Context, table catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(577161)
	if !table.Adding() {
		__antithesis_instrumentation__.Notify(577163)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(577164)
	}
	__antithesis_instrumentation__.Notify(577162)
	log.Info(ctx, "making table public")

	return sc.txn(ctx, func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
		__antithesis_instrumentation__.Notify(577165)
		mut, err := descsCol.GetMutableTableVersionByID(ctx, table.GetID(), txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(577168)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577169)
		}
		__antithesis_instrumentation__.Notify(577166)
		if !mut.Adding() {
			__antithesis_instrumentation__.Notify(577170)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(577171)
		}
		__antithesis_instrumentation__.Notify(577167)
		mut.State = descpb.DescriptorState_PUBLIC
		return descsCol.WriteDesc(ctx, true, mut, txn)
	})
}

func (sc *SchemaChanger) ignoreRevertedDropIndex(
	ctx context.Context, table catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(577172)
	if !table.IsPhysicalTable() {
		__antithesis_instrumentation__.Notify(577174)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(577175)
	}
	__antithesis_instrumentation__.Notify(577173)
	return sc.txn(ctx, func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
		__antithesis_instrumentation__.Notify(577176)
		mut, err := descsCol.GetMutableTableVersionByID(ctx, table.GetID(), txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(577180)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577181)
		}
		__antithesis_instrumentation__.Notify(577177)
		mutationsModified := false
		for _, m := range mut.AllMutations() {
			__antithesis_instrumentation__.Notify(577182)
			if m.MutationID() != sc.mutationID {
				__antithesis_instrumentation__.Notify(577185)
				break
			} else {
				__antithesis_instrumentation__.Notify(577186)
			}
			__antithesis_instrumentation__.Notify(577183)

			if !m.IsRollback() || func() bool {
				__antithesis_instrumentation__.Notify(577187)
				return !m.Adding() == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(577188)
				return m.AsIndex() == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(577189)
				continue
			} else {
				__antithesis_instrumentation__.Notify(577190)
			}
			__antithesis_instrumentation__.Notify(577184)
			log.Warningf(ctx, "ignoring rollback of index drop; index %q will be dropped", m.AsIndex().GetName())
			mut.Mutations[m.MutationOrdinal()].Direction = descpb.DescriptorMutation_DROP
			mutationsModified = true
		}
		__antithesis_instrumentation__.Notify(577178)
		if mutationsModified {
			__antithesis_instrumentation__.Notify(577191)
			return descsCol.WriteDesc(ctx, true, mut, txn)
		} else {
			__antithesis_instrumentation__.Notify(577192)
		}
		__antithesis_instrumentation__.Notify(577179)
		return nil
	})
}

func drainNamesForDescriptor(
	ctx context.Context,
	descID descpb.ID,
	cf *descs.CollectionFactory,
	db *kv.DB,
	ie sqlutil.InternalExecutor,
	codec keys.SQLCodec,
	beforeDrainNames func(),
) error {
	__antithesis_instrumentation__.Notify(577193)
	log.Info(ctx, "draining previous names")

	run := func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
		__antithesis_instrumentation__.Notify(577195)
		if beforeDrainNames != nil {
			__antithesis_instrumentation__.Notify(577202)
			beforeDrainNames()
		} else {
			__antithesis_instrumentation__.Notify(577203)
		}
		__antithesis_instrumentation__.Notify(577196)

		mutDesc, err := descsCol.GetMutableDescriptorByID(ctx, txn, descID)
		if err != nil {
			__antithesis_instrumentation__.Notify(577204)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577205)
		}
		__antithesis_instrumentation__.Notify(577197)
		namesToReclaim := mutDesc.GetDrainingNames()
		if len(namesToReclaim) == 0 {
			__antithesis_instrumentation__.Notify(577206)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(577207)
		}
		__antithesis_instrumentation__.Notify(577198)
		b := txn.NewBatch()
		mutDesc.SetDrainingNames(nil)

		for _, drain := range namesToReclaim {
			__antithesis_instrumentation__.Notify(577208)
			b.Del(catalogkeys.EncodeNameKey(codec, drain))
		}
		__antithesis_instrumentation__.Notify(577199)

		if _, isSchema := mutDesc.(catalog.SchemaDescriptor); isSchema {
			__antithesis_instrumentation__.Notify(577209)
			mutDB, err := descsCol.GetMutableDescriptorByID(ctx, txn, mutDesc.GetParentID())
			if err != nil {
				__antithesis_instrumentation__.Notify(577212)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577213)
			}
			__antithesis_instrumentation__.Notify(577210)
			db := mutDB.(*dbdesc.Mutable)
			for _, name := range namesToReclaim {
				__antithesis_instrumentation__.Notify(577214)
				delete(db.Schemas, name.Name)
			}
			__antithesis_instrumentation__.Notify(577211)
			if err := descsCol.WriteDescToBatch(
				ctx, false, db, b,
			); err != nil {
				__antithesis_instrumentation__.Notify(577215)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577216)
			}
		} else {
			__antithesis_instrumentation__.Notify(577217)
		}
		__antithesis_instrumentation__.Notify(577200)
		if err := descsCol.WriteDescToBatch(
			ctx, false, mutDesc, b,
		); err != nil {
			__antithesis_instrumentation__.Notify(577218)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577219)
		}
		__antithesis_instrumentation__.Notify(577201)
		return txn.Run(ctx, b)
	}
	__antithesis_instrumentation__.Notify(577194)
	return cf.Txn(ctx, ie, db, run)
}

func startGCJob(
	ctx context.Context,
	db *kv.DB,
	jobRegistry *jobs.Registry,
	username security.SQLUsername,
	schemaChangeDescription string,
	details jobspb.SchemaChangeGCDetails,
) error {
	__antithesis_instrumentation__.Notify(577220)
	jobRecord := CreateGCJobRecord(schemaChangeDescription, username, details)
	jobID := jobRegistry.MakeJobID()
	if err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(577222)
		_, err := jobRegistry.CreateJobWithTxn(ctx, jobRecord, jobID, txn)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(577223)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577224)
	}
	__antithesis_instrumentation__.Notify(577221)
	log.Infof(ctx, "created GC job %d", jobID)
	jobRegistry.NotifyToResume(ctx, jobID)
	return nil
}

func (sc *SchemaChanger) execLogTags() *logtags.Buffer {
	__antithesis_instrumentation__.Notify(577225)
	buf := &logtags.Buffer{}
	buf = buf.Add("scExec", nil)

	buf = buf.Add("id", sc.descID)
	if sc.mutationID != descpb.InvalidMutationID {
		__antithesis_instrumentation__.Notify(577228)
		buf = buf.Add("mutation", sc.mutationID)
	} else {
		__antithesis_instrumentation__.Notify(577229)
	}
	__antithesis_instrumentation__.Notify(577226)
	if sc.droppedDatabaseID != descpb.InvalidID {
		__antithesis_instrumentation__.Notify(577230)
		buf = buf.Add("db", sc.droppedDatabaseID)
	} else {
		__antithesis_instrumentation__.Notify(577231)
	}
	__antithesis_instrumentation__.Notify(577227)
	return buf
}

func (sc *SchemaChanger) notFirstInLine(ctx context.Context, desc catalog.Descriptor) error {
	__antithesis_instrumentation__.Notify(577232)
	if tableDesc, ok := desc.(catalog.TableDescriptor); ok {
		__antithesis_instrumentation__.Notify(577234)

		for i, mutation := range tableDesc.AllMutations() {
			__antithesis_instrumentation__.Notify(577235)
			if mutation.MutationID() == sc.mutationID {
				__antithesis_instrumentation__.Notify(577236)
				if i != 0 {
					__antithesis_instrumentation__.Notify(577238)
					log.Infof(ctx,
						"schema change on %q (v%d): another change is still in progress",
						desc.GetName(), desc.GetVersion(),
					)
					return errSchemaChangeNotFirstInLine
				} else {
					__antithesis_instrumentation__.Notify(577239)
				}
				__antithesis_instrumentation__.Notify(577237)
				break
			} else {
				__antithesis_instrumentation__.Notify(577240)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(577241)
	}
	__antithesis_instrumentation__.Notify(577233)
	return nil
}

func (sc *SchemaChanger) getTargetDescriptor(ctx context.Context) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(577242)

	var desc catalog.Descriptor
	if err := sc.txn(ctx, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) (err error) {
		__antithesis_instrumentation__.Notify(577244)
		flags := tree.CommonLookupFlags{
			AvoidLeased:    true,
			Required:       true,
			IncludeOffline: true,
			IncludeDropped: true,
		}
		desc, err = descriptors.GetImmutableDescriptorByID(ctx, txn, sc.descID, flags)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(577245)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(577246)
	}
	__antithesis_instrumentation__.Notify(577243)
	return desc, nil
}

func (sc *SchemaChanger) checkForMVCCCompliantAddIndexMutations(
	ctx context.Context, desc catalog.Descriptor,
) error {
	__antithesis_instrumentation__.Notify(577247)
	tableDesc, ok := desc.(catalog.TableDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(577251)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(577252)
	}
	__antithesis_instrumentation__.Notify(577248)

	nonTempAddingIndexes := 0
	tempIndexes := 0

	for _, m := range tableDesc.AllMutations() {
		__antithesis_instrumentation__.Notify(577253)
		if m.MutationID() != sc.mutationID {
			__antithesis_instrumentation__.Notify(577256)
			break
		} else {
			__antithesis_instrumentation__.Notify(577257)
		}
		__antithesis_instrumentation__.Notify(577254)

		idx := m.AsIndex()
		if idx == nil {
			__antithesis_instrumentation__.Notify(577258)
			continue
		} else {
			__antithesis_instrumentation__.Notify(577259)
		}
		__antithesis_instrumentation__.Notify(577255)

		if idx.IsTemporaryIndexForBackfill() {
			__antithesis_instrumentation__.Notify(577260)
			tempIndexes++
		} else {
			__antithesis_instrumentation__.Notify(577261)
			if m.Adding() {
				__antithesis_instrumentation__.Notify(577262)
				nonTempAddingIndexes++
			} else {
				__antithesis_instrumentation__.Notify(577263)
			}
		}
	}
	__antithesis_instrumentation__.Notify(577249)

	if tempIndexes > 0 {
		__antithesis_instrumentation__.Notify(577264)
		sc.mvccCompliantAddIndex = true

		if tempIndexes != nonTempAddingIndexes {
			__antithesis_instrumentation__.Notify(577266)
			return errors.Newf("expected %d temporary indexes, but found %d; schema change may have been constructed during cluster version upgrade",
				tempIndexes,
				nonTempAddingIndexes)
		} else {
			__antithesis_instrumentation__.Notify(577267)
		}
		__antithesis_instrumentation__.Notify(577265)

		settings := sc.execCfg.Settings
		mvccCompliantBackfillSupported := settings.Version.IsActive(ctx, clusterversion.MVCCIndexBackfiller) && func() bool {
			__antithesis_instrumentation__.Notify(577268)
			return tabledesc.UseMVCCCompliantIndexCreation.Get(&settings.SV) == true
		}() == true
		if !mvccCompliantBackfillSupported {
			__antithesis_instrumentation__.Notify(577269)
			return errors.Newf("schema change requires MVCC-compliant backfiller, but MVCC-compliant backfiller is not supported")
		} else {
			__antithesis_instrumentation__.Notify(577270)
		}
	} else {
		__antithesis_instrumentation__.Notify(577271)
	}
	__antithesis_instrumentation__.Notify(577250)
	return nil
}

func (sc *SchemaChanger) exec(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(577272)
	sc.metrics.RunningSchemaChanges.Inc(1)
	defer sc.metrics.RunningSchemaChanges.Dec(1)

	ctx = logtags.AddTags(ctx, sc.execLogTags())

	desc, err := sc.getTargetDescriptor(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(577289)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577290)
	}
	__antithesis_instrumentation__.Notify(577273)

	if err := sc.notFirstInLine(ctx, desc); err != nil {
		__antithesis_instrumentation__.Notify(577291)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577292)
	}
	__antithesis_instrumentation__.Notify(577274)

	if err := sc.checkForMVCCCompliantAddIndexMutations(ctx, desc); err != nil {
		__antithesis_instrumentation__.Notify(577293)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577294)
	}
	__antithesis_instrumentation__.Notify(577275)

	log.Infof(ctx,
		"schema change on %q (v%d) starting execution...",
		desc.GetName(), desc.GetVersion(),
	)

	if len(desc.GetDrainingNames()) > 0 {
		__antithesis_instrumentation__.Notify(577295)
		if err := drainNamesForDescriptor(
			ctx, desc.GetID(), sc.execCfg.CollectionFactory, sc.db, sc.execCfg.InternalExecutor,
			sc.execCfg.Codec, sc.testingKnobs.OldNamesDrainedNotification,
		); err != nil {
			__antithesis_instrumentation__.Notify(577296)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577297)
		}
	} else {
		__antithesis_instrumentation__.Notify(577298)
	}
	__antithesis_instrumentation__.Notify(577276)

	waitToUpdateLeases := func(refreshStats bool) error {
		__antithesis_instrumentation__.Notify(577299)
		latestDesc, err := WaitToUpdateLeases(ctx, sc.leaseMgr, sc.descID)
		if err != nil {
			__antithesis_instrumentation__.Notify(577302)
			if errors.Is(err, catalog.ErrDescriptorNotFound) {
				__antithesis_instrumentation__.Notify(577304)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577305)
			}
			__antithesis_instrumentation__.Notify(577303)
			log.Warningf(ctx, "waiting to update leases: %+v", err)

			sqltelemetry.RecordError(ctx, err, &sc.settings.SV)
		} else {
			__antithesis_instrumentation__.Notify(577306)
		}
		__antithesis_instrumentation__.Notify(577300)

		if refreshStats {
			__antithesis_instrumentation__.Notify(577307)
			sc.refreshStats(latestDesc)
		} else {
			__antithesis_instrumentation__.Notify(577308)
		}
		__antithesis_instrumentation__.Notify(577301)
		return nil
	}
	__antithesis_instrumentation__.Notify(577277)

	tableDesc, ok := desc.(catalog.TableDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(577309)

		if err := waitToUpdateLeases(false); err != nil {
			__antithesis_instrumentation__.Notify(577312)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577313)
		}
		__antithesis_instrumentation__.Notify(577310)

		switch desc.(type) {
		case catalog.SchemaDescriptor, catalog.DatabaseDescriptor:
			__antithesis_instrumentation__.Notify(577314)
			if desc.Dropped() {
				__antithesis_instrumentation__.Notify(577315)
				if err := sc.execCfg.DB.Del(ctx, catalogkeys.MakeDescMetadataKey(sc.execCfg.Codec, desc.GetID())); err != nil {
					__antithesis_instrumentation__.Notify(577316)
					return err
				} else {
					__antithesis_instrumentation__.Notify(577317)
				}
			} else {
				__antithesis_instrumentation__.Notify(577318)
			}
		}
		__antithesis_instrumentation__.Notify(577311)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(577319)
	}
	__antithesis_instrumentation__.Notify(577278)

	if tableDesc.Dropped() && func() bool {
		__antithesis_instrumentation__.Notify(577320)
		return sc.droppedDatabaseID == descpb.InvalidID == true
	}() == true {
		__antithesis_instrumentation__.Notify(577321)
		if tableDesc.IsPhysicalTable() {
			__antithesis_instrumentation__.Notify(577322)

			dropTime := timeutil.Now().UnixNano()
			if tableDesc.GetDropTime() > 0 {
				__antithesis_instrumentation__.Notify(577324)
				dropTime = tableDesc.GetDropTime()
			} else {
				__antithesis_instrumentation__.Notify(577325)
			}
			__antithesis_instrumentation__.Notify(577323)
			gcDetails := jobspb.SchemaChangeGCDetails{
				Tables: []jobspb.SchemaChangeGCDetails_DroppedID{
					{
						ID:       tableDesc.GetID(),
						DropTime: dropTime,
					},
				},
			}
			if err := startGCJob(
				ctx, sc.db, sc.jobRegistry, sc.job.Payload().UsernameProto.Decode(), sc.job.Payload().Description, gcDetails,
			); err != nil {
				__antithesis_instrumentation__.Notify(577326)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577327)
			}
		} else {
			__antithesis_instrumentation__.Notify(577328)

			if err := DeleteTableDescAndZoneConfig(
				ctx, sc.db, sc.settings, sc.execCfg.Codec, tableDesc,
			); err != nil {
				__antithesis_instrumentation__.Notify(577329)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577330)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(577331)
	}
	__antithesis_instrumentation__.Notify(577279)

	if err := sc.ignoreRevertedDropIndex(ctx, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(577332)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577333)
	}
	__antithesis_instrumentation__.Notify(577280)

	if err := sc.maybeBackfillCreateTableAs(ctx, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(577334)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577335)
	}
	__antithesis_instrumentation__.Notify(577281)

	if err := sc.maybeBackfillMaterializedView(ctx, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(577336)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577337)
	}
	__antithesis_instrumentation__.Notify(577282)

	if err := sc.maybeMakeAddTablePublic(ctx, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(577338)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577339)
	}
	__antithesis_instrumentation__.Notify(577283)

	if err := sc.maybeUpdateScheduledJobsForRowLevelTTL(ctx, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(577340)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577341)
	}
	__antithesis_instrumentation__.Notify(577284)

	if sc.mutationID == descpb.InvalidMutationID {
		__antithesis_instrumentation__.Notify(577342)

		isCreateTableAs := tableDesc.Adding() && func() bool {
			__antithesis_instrumentation__.Notify(577343)
			return tableDesc.IsAs() == true
		}() == true
		return waitToUpdateLeases(isCreateTableAs)
	} else {
		__antithesis_instrumentation__.Notify(577344)
	}
	__antithesis_instrumentation__.Notify(577285)

	if err := sc.initJobRunningStatus(ctx); err != nil {
		__antithesis_instrumentation__.Notify(577345)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(577347)
			log.Infof(ctx, "failed to update job status: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(577348)
		}
		__antithesis_instrumentation__.Notify(577346)

		sqltelemetry.RecordError(ctx, err, &sc.settings.SV)
		if jobs.IsPauseSelfError(err) {
			__antithesis_instrumentation__.Notify(577349)

			return err
		} else {
			__antithesis_instrumentation__.Notify(577350)
		}
	} else {
		__antithesis_instrumentation__.Notify(577351)
	}
	__antithesis_instrumentation__.Notify(577286)

	if err := sc.runStateMachineAndBackfill(ctx); err != nil {
		__antithesis_instrumentation__.Notify(577352)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577353)
	}
	__antithesis_instrumentation__.Notify(577287)

	defer func() {
		__antithesis_instrumentation__.Notify(577354)
		if err := waitToUpdateLeases(err == nil); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(577355)
			return !errors.Is(err, catalog.ErrDescriptorNotFound) == true
		}() == true {
			__antithesis_instrumentation__.Notify(577356)

			log.Warningf(ctx, "unexpected error while waiting for leases to update: %+v", err)

			sqltelemetry.RecordError(ctx, err, &sc.settings.SV)
		} else {
			__antithesis_instrumentation__.Notify(577357)
		}
	}()
	__antithesis_instrumentation__.Notify(577288)

	return err
}

func (sc *SchemaChanger) handlePermanentSchemaChangeError(
	ctx context.Context, err error, evalCtx *extendedEvalContext,
) error {
	__antithesis_instrumentation__.Notify(577358)

	{
		__antithesis_instrumentation__.Notify(577363)

		desc, descErr := sc.getTargetDescriptor(ctx)
		if descErr != nil {
			__antithesis_instrumentation__.Notify(577366)
			return descErr
		} else {
			__antithesis_instrumentation__.Notify(577367)
		}
		__antithesis_instrumentation__.Notify(577364)

		if _, ok := desc.(catalog.TableDescriptor); !ok {
			__antithesis_instrumentation__.Notify(577368)
			return jobs.MarkAsPermanentJobError(errors.Newf("schema change jobs on databases and schemas cannot be reverted"))
		} else {
			__antithesis_instrumentation__.Notify(577369)
		}
		__antithesis_instrumentation__.Notify(577365)

		if err := sc.notFirstInLine(ctx, desc); err != nil {
			__antithesis_instrumentation__.Notify(577370)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577371)
		}
	}
	__antithesis_instrumentation__.Notify(577359)

	if rollbackErr := sc.rollbackSchemaChange(ctx, err); rollbackErr != nil {
		__antithesis_instrumentation__.Notify(577372)

		secondary := errors.Wrap(err, "original error when rolling back mutations")
		sqltelemetry.RecordError(ctx, secondary, &sc.settings.SV)
		return errors.WithSecondaryError(rollbackErr, secondary)
	} else {
		__antithesis_instrumentation__.Notify(577373)
	}
	__antithesis_instrumentation__.Notify(577360)

	waitToUpdateLeases := func(refreshStats bool) error {
		__antithesis_instrumentation__.Notify(577374)
		desc, err := WaitToUpdateLeases(ctx, sc.leaseMgr, sc.descID)
		if err != nil {
			__antithesis_instrumentation__.Notify(577377)
			if errors.Is(err, catalog.ErrDescriptorNotFound) {
				__antithesis_instrumentation__.Notify(577379)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577380)
			}
			__antithesis_instrumentation__.Notify(577378)
			log.Warningf(ctx, "waiting to update leases: %+v", err)

			sqltelemetry.RecordError(ctx, err, &sc.settings.SV)
		} else {
			__antithesis_instrumentation__.Notify(577381)
		}
		__antithesis_instrumentation__.Notify(577375)

		if refreshStats {
			__antithesis_instrumentation__.Notify(577382)
			sc.refreshStats(desc)
		} else {
			__antithesis_instrumentation__.Notify(577383)
		}
		__antithesis_instrumentation__.Notify(577376)
		return nil
	}
	__antithesis_instrumentation__.Notify(577361)

	defer func() {
		__antithesis_instrumentation__.Notify(577384)
		if err := waitToUpdateLeases(false); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(577385)
			return !errors.Is(err, catalog.ErrDescriptorNotFound) == true
		}() == true {
			__antithesis_instrumentation__.Notify(577386)

			log.Warningf(ctx, "unexpected error while waiting for leases to update: %+v", err)

			sqltelemetry.RecordError(ctx, err, &sc.settings.SV)
		} else {
			__antithesis_instrumentation__.Notify(577387)
		}
	}()
	__antithesis_instrumentation__.Notify(577362)

	return nil
}

func (sc *SchemaChanger) initJobRunningStatus(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(577388)
	return sc.txn(ctx, func(ctx context.Context, txn *kv.Txn, descriptors *descs.Collection) error {
		__antithesis_instrumentation__.Notify(577389)
		flags := tree.ObjectLookupFlagsWithRequired()
		flags.AvoidLeased = true
		desc, err := descriptors.GetImmutableTableByID(ctx, txn, sc.descID, flags)
		if err != nil {
			__antithesis_instrumentation__.Notify(577393)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577394)
		}
		__antithesis_instrumentation__.Notify(577390)

		var runStatus jobs.RunningStatus
		for _, mutation := range desc.AllMutations() {
			__antithesis_instrumentation__.Notify(577395)
			if mutation.MutationID() != sc.mutationID {
				__antithesis_instrumentation__.Notify(577397)

				break
			} else {
				__antithesis_instrumentation__.Notify(577398)
			}
			__antithesis_instrumentation__.Notify(577396)

			if mutation.Adding() && func() bool {
				__antithesis_instrumentation__.Notify(577399)
				return mutation.DeleteOnly() == true
			}() == true {
				__antithesis_instrumentation__.Notify(577400)
				runStatus = RunningStatusDeleteOnly
			} else {
				__antithesis_instrumentation__.Notify(577401)
				if mutation.Dropped() && func() bool {
					__antithesis_instrumentation__.Notify(577402)
					return mutation.WriteAndDeleteOnly() == true
				}() == true {
					__antithesis_instrumentation__.Notify(577403)
					runStatus = RunningStatusDeleteAndWriteOnly
				} else {
					__antithesis_instrumentation__.Notify(577404)
				}
			}
		}
		__antithesis_instrumentation__.Notify(577391)
		if runStatus != "" && func() bool {
			__antithesis_instrumentation__.Notify(577405)
			return !desc.Dropped() == true
		}() == true {
			__antithesis_instrumentation__.Notify(577406)
			if err := sc.job.RunningStatus(
				ctx, txn, func(ctx context.Context, details jobspb.Details) (jobs.RunningStatus, error) {
					__antithesis_instrumentation__.Notify(577407)
					return runStatus, nil
				}); err != nil {
				__antithesis_instrumentation__.Notify(577408)
				return errors.Wrapf(err, "failed to update job status")
			} else {
				__antithesis_instrumentation__.Notify(577409)
			}
		} else {
			__antithesis_instrumentation__.Notify(577410)
		}
		__antithesis_instrumentation__.Notify(577392)
		return nil
	})
}

func (sc *SchemaChanger) rollbackSchemaChange(ctx context.Context, err error) error {
	__antithesis_instrumentation__.Notify(577411)
	log.Warningf(ctx, "reversing schema change %d due to irrecoverable error: %s", sc.job.ID(), err)
	if errReverse := sc.maybeReverseMutations(ctx, err); errReverse != nil {
		__antithesis_instrumentation__.Notify(577416)
		return errReverse
	} else {
		__antithesis_instrumentation__.Notify(577417)
	}
	__antithesis_instrumentation__.Notify(577412)

	if fn := sc.testingKnobs.RunAfterMutationReversal; fn != nil {
		__antithesis_instrumentation__.Notify(577418)
		if err := fn(sc.job.ID()); err != nil {
			__antithesis_instrumentation__.Notify(577419)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577420)
		}
	} else {
		__antithesis_instrumentation__.Notify(577421)
	}
	__antithesis_instrumentation__.Notify(577413)

	if err := sc.runStateMachineAndBackfill(ctx); err != nil {
		__antithesis_instrumentation__.Notify(577422)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577423)
	}
	__antithesis_instrumentation__.Notify(577414)

	gcJobID := sc.jobRegistry.MakeJobID()
	if err := sc.txn(ctx, func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
		__antithesis_instrumentation__.Notify(577424)
		scTable, err := descsCol.GetMutableTableVersionByID(ctx, sc.descID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(577429)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577430)
		}
		__antithesis_instrumentation__.Notify(577425)
		if !scTable.Adding() {
			__antithesis_instrumentation__.Notify(577431)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(577432)
		}
		__antithesis_instrumentation__.Notify(577426)

		b := txn.NewBatch()
		scTable.SetDropped()
		scTable.DropTime = timeutil.Now().UnixNano()
		if err := descsCol.WriteDescToBatch(ctx, false, scTable, b); err != nil {
			__antithesis_instrumentation__.Notify(577433)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577434)
		}
		__antithesis_instrumentation__.Notify(577427)
		b.Del(catalogkeys.EncodeNameKey(sc.execCfg.Codec, scTable))

		jobRecord := CreateGCJobRecord(
			"ROLLBACK OF "+sc.job.Payload().Description,
			sc.job.Payload().UsernameProto.Decode(),
			jobspb.SchemaChangeGCDetails{
				Tables: []jobspb.SchemaChangeGCDetails_DroppedID{
					{
						ID:       scTable.GetID(),
						DropTime: timeutil.Now().UnixNano(),
					},
				},
			},
		)
		if _, err := sc.jobRegistry.CreateJobWithTxn(ctx, jobRecord, gcJobID, txn); err != nil {
			__antithesis_instrumentation__.Notify(577435)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577436)
		}
		__antithesis_instrumentation__.Notify(577428)
		return txn.Run(ctx, b)
	}); err != nil {
		__antithesis_instrumentation__.Notify(577437)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577438)
	}
	__antithesis_instrumentation__.Notify(577415)
	log.Infof(ctx, "starting GC job %d", gcJobID)
	sc.jobRegistry.NotifyToResume(ctx, gcJobID)
	return nil
}

func (sc *SchemaChanger) RunStateMachineBeforeBackfill(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(577439)
	log.Info(ctx, "stepping through state machine")

	var runStatus jobs.RunningStatus
	if err := sc.txn(ctx, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(577441)
		tbl, err := descsCol.GetMutableTableVersionByID(ctx, sc.descID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(577448)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577449)
		}
		__antithesis_instrumentation__.Notify(577442)
		_, dbDesc, err := descsCol.GetImmutableDatabaseByID(
			ctx,
			txn,
			tbl.GetParentID(),
			tree.DatabaseLookupFlags{
				Required:    true,
				AvoidLeased: true,
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(577450)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577451)
		}
		__antithesis_instrumentation__.Notify(577443)
		runStatus = ""

		for _, m := range tbl.AllMutations() {
			__antithesis_instrumentation__.Notify(577452)
			if m.MutationID() != sc.mutationID {
				__antithesis_instrumentation__.Notify(577455)

				break
			} else {
				__antithesis_instrumentation__.Notify(577456)
			}
			__antithesis_instrumentation__.Notify(577453)
			if m.Adding() {
				__antithesis_instrumentation__.Notify(577457)
				if m.DeleteOnly() {
					__antithesis_instrumentation__.Notify(577458)

					tbl.Mutations[m.MutationOrdinal()].State = descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY
					runStatus = RunningStatusDeleteAndWriteOnly
				} else {
					__antithesis_instrumentation__.Notify(577459)
				}

			} else {
				__antithesis_instrumentation__.Notify(577460)
				if m.Dropped() {
					__antithesis_instrumentation__.Notify(577461)
					if m.WriteAndDeleteOnly() || func() bool {
						__antithesis_instrumentation__.Notify(577462)
						return m.Merging() == true
					}() == true {
						__antithesis_instrumentation__.Notify(577463)
						tbl.Mutations[m.MutationOrdinal()].State = descpb.DescriptorMutation_DELETE_ONLY
						runStatus = RunningStatusDeleteOnly
					} else {
						__antithesis_instrumentation__.Notify(577464)
					}

				} else {
					__antithesis_instrumentation__.Notify(577465)
				}
			}
			__antithesis_instrumentation__.Notify(577454)

			if err := sc.applyZoneConfigChangeForMutation(
				ctx,
				txn,
				dbDesc,
				tbl,
				m,
				false,
				descsCol,
			); err != nil {
				__antithesis_instrumentation__.Notify(577466)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577467)
			}
		}
		__antithesis_instrumentation__.Notify(577444)
		if doNothing := runStatus == "" || func() bool {
			__antithesis_instrumentation__.Notify(577468)
			return tbl.Dropped() == true
		}() == true; doNothing {
			__antithesis_instrumentation__.Notify(577469)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(577470)
		}
		__antithesis_instrumentation__.Notify(577445)
		if err := descsCol.WriteDesc(
			ctx, true, tbl, txn,
		); err != nil {
			__antithesis_instrumentation__.Notify(577471)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577472)
		}
		__antithesis_instrumentation__.Notify(577446)
		if sc.job != nil {
			__antithesis_instrumentation__.Notify(577473)
			if err := sc.job.RunningStatus(ctx, txn, func(
				ctx context.Context, details jobspb.Details,
			) (jobs.RunningStatus, error) {
				__antithesis_instrumentation__.Notify(577474)
				return runStatus, nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(577475)
				return errors.Wrap(err, "failed to update job status")
			} else {
				__antithesis_instrumentation__.Notify(577476)
			}
		} else {
			__antithesis_instrumentation__.Notify(577477)
		}
		__antithesis_instrumentation__.Notify(577447)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(577478)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577479)
	}
	__antithesis_instrumentation__.Notify(577440)

	log.Info(ctx, "finished stepping through state machine")
	return nil
}

func (sc *SchemaChanger) RunStateMachineAfterIndexBackfill(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(577480)

	log.Info(ctx, "stepping through state machine after index backfill")
	if err := sc.stepStateMachineAfterIndexBackfill(ctx); err != nil {
		__antithesis_instrumentation__.Notify(577483)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577484)
	}
	__antithesis_instrumentation__.Notify(577481)
	if err := sc.stepStateMachineAfterIndexBackfill(ctx); err != nil {
		__antithesis_instrumentation__.Notify(577485)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577486)
	}
	__antithesis_instrumentation__.Notify(577482)
	log.Info(ctx, "finished stepping through state machine")
	return nil
}

func (sc *SchemaChanger) stepStateMachineAfterIndexBackfill(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(577487)
	log.Info(ctx, "stepping through state machine")

	var runStatus jobs.RunningStatus
	if err := sc.txn(ctx, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(577489)
		tbl, err := descsCol.GetMutableTableVersionByID(ctx, sc.descID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(577495)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577496)
		}
		__antithesis_instrumentation__.Notify(577490)
		runStatus = ""
		for _, m := range tbl.AllMutations() {
			__antithesis_instrumentation__.Notify(577497)
			if m.MutationID() != sc.mutationID {
				__antithesis_instrumentation__.Notify(577500)

				break
			} else {
				__antithesis_instrumentation__.Notify(577501)
			}
			__antithesis_instrumentation__.Notify(577498)
			idx := m.AsIndex()
			if idx == nil {
				__antithesis_instrumentation__.Notify(577502)

				continue
			} else {
				__antithesis_instrumentation__.Notify(577503)
			}
			__antithesis_instrumentation__.Notify(577499)

			if m.Adding() {
				__antithesis_instrumentation__.Notify(577504)
				if m.Backfilling() {
					__antithesis_instrumentation__.Notify(577505)
					tbl.Mutations[m.MutationOrdinal()].State = descpb.DescriptorMutation_DELETE_ONLY
					runStatus = RunningStatusDeleteOnly
				} else {
					__antithesis_instrumentation__.Notify(577506)
					if m.DeleteOnly() {
						__antithesis_instrumentation__.Notify(577507)
						tbl.Mutations[m.MutationOrdinal()].State = descpb.DescriptorMutation_MERGING
						runStatus = RunningStatusMerging
					} else {
						__antithesis_instrumentation__.Notify(577508)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(577509)
			}
		}
		__antithesis_instrumentation__.Notify(577491)
		if runStatus == "" || func() bool {
			__antithesis_instrumentation__.Notify(577510)
			return tbl.Dropped() == true
		}() == true {
			__antithesis_instrumentation__.Notify(577511)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(577512)
		}
		__antithesis_instrumentation__.Notify(577492)
		if err := descsCol.WriteDesc(
			ctx, true, tbl, txn,
		); err != nil {
			__antithesis_instrumentation__.Notify(577513)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577514)
		}
		__antithesis_instrumentation__.Notify(577493)
		if sc.job != nil {
			__antithesis_instrumentation__.Notify(577515)
			if err := sc.job.RunningStatus(ctx, txn, func(
				ctx context.Context, details jobspb.Details,
			) (jobs.RunningStatus, error) {
				__antithesis_instrumentation__.Notify(577516)
				return runStatus, nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(577517)
				return errors.Wrap(err, "failed to update job status")
			} else {
				__antithesis_instrumentation__.Notify(577518)
			}
		} else {
			__antithesis_instrumentation__.Notify(577519)
		}
		__antithesis_instrumentation__.Notify(577494)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(577520)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577521)
	}
	__antithesis_instrumentation__.Notify(577488)
	return nil
}

func (sc *SchemaChanger) createTemporaryIndexGCJob(
	ctx context.Context, indexID descpb.IndexID, txn *kv.Txn, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(577522)
	minimumDropTime := int64(1)
	return sc.createIndexGCJobWithDropTime(ctx, indexID, txn, jobDesc, minimumDropTime)
}

func (sc *SchemaChanger) createIndexGCJob(
	ctx context.Context, indexID descpb.IndexID, txn *kv.Txn, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(577523)
	dropTime := timeutil.Now().UnixNano()
	return sc.createIndexGCJobWithDropTime(ctx, indexID, txn, jobDesc, dropTime)
}

func (sc *SchemaChanger) createIndexGCJobWithDropTime(
	ctx context.Context, indexID descpb.IndexID, txn *kv.Txn, jobDesc string, dropTime int64,
) error {
	__antithesis_instrumentation__.Notify(577524)
	indexGCDetails := jobspb.SchemaChangeGCDetails{
		Indexes: []jobspb.SchemaChangeGCDetails_DroppedIndex{
			{
				IndexID:  indexID,
				DropTime: dropTime,
			},
		},
		ParentID: sc.descID,
	}

	gcJobRecord := CreateGCJobRecord(jobDesc, sc.job.Payload().UsernameProto.Decode(), indexGCDetails)
	jobID := sc.jobRegistry.MakeJobID()
	if _, err := sc.jobRegistry.CreateJobWithTxn(ctx, gcJobRecord, jobID, txn); err != nil {
		__antithesis_instrumentation__.Notify(577526)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577527)
	}
	__antithesis_instrumentation__.Notify(577525)
	log.Infof(ctx, "created index GC job %d", jobID)
	sc.jobRegistry.NotifyToResume(ctx, jobID)
	return nil
}

func WaitToUpdateLeases(
	ctx context.Context, leaseMgr *lease.Manager, descID descpb.ID,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(577528)

	retryOpts := retry.Options{
		InitialBackoff: 5 * time.Millisecond,
		MaxBackoff:     time.Second,
		Multiplier:     1.5,
	}
	start := timeutil.Now()
	log.Infof(ctx, "waiting for a single version...")
	desc, err := leaseMgr.WaitForOneVersion(ctx, descID, retryOpts)
	var version descpb.DescriptorVersion
	if desc != nil {
		__antithesis_instrumentation__.Notify(577530)
		version = desc.GetVersion()
	} else {
		__antithesis_instrumentation__.Notify(577531)
	}
	__antithesis_instrumentation__.Notify(577529)
	log.Infof(ctx, "waiting for a single version... done (at v %d), took %v", version, timeutil.Since(start))
	return desc, err
}

func (sc *SchemaChanger) done(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(577532)

	type commentToDelete struct {
		id          int64
		subID       int64
		commentType keys.CommentType
	}
	type commentToSwap struct {
		id          int64
		oldSubID    int64
		newSubID    int64
		commentType keys.CommentType
	}
	var commentsToDelete []commentToDelete
	var commentsToSwap []commentToSwap

	var didUpdate bool
	var depMutationJobs []jobspb.JobID
	var otherJobIDs []jobspb.JobID
	err := sc.execCfg.CollectionFactory.Txn(ctx, sc.execCfg.InternalExecutor, sc.db, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(577536)
		depMutationJobs = depMutationJobs[:0]
		otherJobIDs = otherJobIDs[:0]
		var err error
		scTable, err := descsCol.GetMutableTableVersionByID(ctx, sc.descID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(577553)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577554)
		}
		__antithesis_instrumentation__.Notify(577537)

		_, dbDesc, err := descsCol.GetImmutableDatabaseByID(
			ctx,
			txn,
			scTable.GetParentID(),
			tree.DatabaseLookupFlags{
				Required:    true,
				AvoidLeased: true,
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(577555)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577556)
		}
		__antithesis_instrumentation__.Notify(577538)

		collectReferencedTypeIDs := func() (catalog.DescriptorIDSet, error) {
			__antithesis_instrumentation__.Notify(577557)
			typeLookupFn := func(id descpb.ID) (catalog.TypeDescriptor, error) {
				__antithesis_instrumentation__.Notify(577559)
				desc, err := descsCol.GetImmutableTypeByID(ctx, txn, id, tree.ObjectLookupFlags{})
				if err != nil {
					__antithesis_instrumentation__.Notify(577561)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(577562)
				}
				__antithesis_instrumentation__.Notify(577560)
				return desc, nil
			}
			__antithesis_instrumentation__.Notify(577558)
			ids, _, err := scTable.GetAllReferencedTypeIDs(dbDesc, typeLookupFn)
			return catalog.MakeDescriptorIDSet(ids...), err
		}
		__antithesis_instrumentation__.Notify(577539)
		referencedTypeIDs, err := collectReferencedTypeIDs()

		collectReferencedSequenceIDs := func() map[descpb.ID]descpb.ColumnIDs {
			__antithesis_instrumentation__.Notify(577563)
			m := make(map[descpb.ID]descpb.ColumnIDs)
			for _, col := range scTable.AllColumns() {
				__antithesis_instrumentation__.Notify(577565)
				for i := 0; i < col.NumUsesSequences(); i++ {
					__antithesis_instrumentation__.Notify(577566)
					id := col.GetUsesSequenceID(i)
					m[id] = append(m[id], col.GetID())
				}
			}
			__antithesis_instrumentation__.Notify(577564)
			return m
		}
		__antithesis_instrumentation__.Notify(577540)
		referencedSequenceIDs := collectReferencedSequenceIDs()

		if err != nil {
			__antithesis_instrumentation__.Notify(577567)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577568)
		}
		__antithesis_instrumentation__.Notify(577541)
		b := txn.NewBatch()
		const kvTrace = true

		var i int
		var isRollback bool
		for _, m := range scTable.AllMutations() {
			__antithesis_instrumentation__.Notify(577569)
			if m.MutationID() != sc.mutationID {
				__antithesis_instrumentation__.Notify(577580)

				break
			} else {
				__antithesis_instrumentation__.Notify(577581)
			}
			__antithesis_instrumentation__.Notify(577570)
			isRollback = m.IsRollback()
			if idx := m.AsIndex(); m.Dropped() && func() bool {
				__antithesis_instrumentation__.Notify(577582)
				return idx != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(577583)
				description := sc.job.Payload().Description
				if isRollback {
					__antithesis_instrumentation__.Notify(577585)
					description = "ROLLBACK of " + description
				} else {
					__antithesis_instrumentation__.Notify(577586)
				}
				__antithesis_instrumentation__.Notify(577584)
				if idx.IsTemporaryIndexForBackfill() {
					__antithesis_instrumentation__.Notify(577587)
					if err := sc.createTemporaryIndexGCJob(ctx, idx.GetID(), txn, "temporary index used during index backfill"); err != nil {
						__antithesis_instrumentation__.Notify(577588)
						return err
					} else {
						__antithesis_instrumentation__.Notify(577589)
					}
				} else {
					__antithesis_instrumentation__.Notify(577590)
					if err := sc.createIndexGCJob(ctx, idx.GetID(), txn, description); err != nil {
						__antithesis_instrumentation__.Notify(577591)
						return err
					} else {
						__antithesis_instrumentation__.Notify(577592)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(577593)
			}
			__antithesis_instrumentation__.Notify(577571)
			if constraint := m.AsConstraint(); constraint != nil && func() bool {
				__antithesis_instrumentation__.Notify(577594)
				return constraint.Adding() == true
			}() == true {
				__antithesis_instrumentation__.Notify(577595)
				if constraint.IsForeignKey() && func() bool {
					__antithesis_instrumentation__.Notify(577596)
					return constraint.ForeignKey().Validity == descpb.ConstraintValidity_Unvalidated == true
				}() == true {
					__antithesis_instrumentation__.Notify(577597)

					backrefTable, err := descsCol.GetMutableTableVersionByID(ctx, constraint.ForeignKey().ReferencedTableID, txn)
					if err != nil {
						__antithesis_instrumentation__.Notify(577599)
						return err
					} else {
						__antithesis_instrumentation__.Notify(577600)
					}
					__antithesis_instrumentation__.Notify(577598)
					backrefTable.InboundFKs = append(backrefTable.InboundFKs, constraint.ForeignKey())
					if backrefTable != scTable {
						__antithesis_instrumentation__.Notify(577601)
						if err := descsCol.WriteDescToBatch(ctx, kvTrace, backrefTable, b); err != nil {
							__antithesis_instrumentation__.Notify(577602)
							return err
						} else {
							__antithesis_instrumentation__.Notify(577603)
						}
					} else {
						__antithesis_instrumentation__.Notify(577604)
					}
				} else {
					__antithesis_instrumentation__.Notify(577605)
				}
			} else {
				__antithesis_instrumentation__.Notify(577606)
			}
			__antithesis_instrumentation__.Notify(577572)

			if err := sc.applyZoneConfigChangeForMutation(
				ctx,
				txn,
				dbDesc,
				scTable,
				m,
				true,
				descsCol,
			); err != nil {
				__antithesis_instrumentation__.Notify(577607)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577608)
			}
			__antithesis_instrumentation__.Notify(577573)

			if refresh := m.AsMaterializedViewRefresh(); refresh != nil {
				__antithesis_instrumentation__.Notify(577609)
				if fn := sc.testingKnobs.RunBeforeMaterializedViewRefreshCommit; fn != nil {
					__antithesis_instrumentation__.Notify(577611)
					if err := fn(); err != nil {
						__antithesis_instrumentation__.Notify(577612)
						return err
					} else {
						__antithesis_instrumentation__.Notify(577613)
					}
				} else {
					__antithesis_instrumentation__.Notify(577614)
				}
				__antithesis_instrumentation__.Notify(577610)

				if m.Adding() {
					__antithesis_instrumentation__.Notify(577615)
					desc := fmt.Sprintf("REFRESH MATERIALIZED VIEW %q cleanup", scTable.Name)
					for _, idx := range scTable.ActiveIndexes() {
						__antithesis_instrumentation__.Notify(577616)
						if err := sc.createIndexGCJob(ctx, idx.GetID(), txn, desc); err != nil {
							__antithesis_instrumentation__.Notify(577617)
							return err
						} else {
							__antithesis_instrumentation__.Notify(577618)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(577619)
					if m.Dropped() {
						__antithesis_instrumentation__.Notify(577620)

						desc := fmt.Sprintf("ROLLBACK OF REFRESH MATERIALIZED VIEW %q", scTable.Name)
						err = refresh.ForEachIndexID(func(id descpb.IndexID) error {
							__antithesis_instrumentation__.Notify(577622)
							return sc.createIndexGCJob(ctx, id, txn, desc)
						})
						__antithesis_instrumentation__.Notify(577621)
						if err != nil {
							__antithesis_instrumentation__.Notify(577623)
							return err
						} else {
							__antithesis_instrumentation__.Notify(577624)
						}
					} else {
						__antithesis_instrumentation__.Notify(577625)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(577626)
			}
			__antithesis_instrumentation__.Notify(577574)

			if pkSwap := m.AsPrimaryKeySwap(); pkSwap != nil {
				__antithesis_instrumentation__.Notify(577627)
				id := pkSwap.PrimaryKeySwapDesc().OldPrimaryIndexId
				commentsToDelete = append(commentsToDelete,
					commentToDelete{
						id:          int64(scTable.GetID()),
						subID:       int64(id),
						commentType: keys.IndexCommentType,
					})
				for i := range pkSwap.PrimaryKeySwapDesc().OldIndexes {
					__antithesis_instrumentation__.Notify(577628)

					if pkSwap.PrimaryKeySwapDesc().OldIndexes[i] == id {
						__antithesis_instrumentation__.Notify(577630)
						continue
					} else {
						__antithesis_instrumentation__.Notify(577631)
					}
					__antithesis_instrumentation__.Notify(577629)

					commentsToSwap = append(commentsToSwap,
						commentToSwap{
							id:          int64(scTable.GetID()),
							oldSubID:    int64(pkSwap.PrimaryKeySwapDesc().OldIndexes[i]),
							newSubID:    int64(pkSwap.PrimaryKeySwapDesc().NewIndexes[i]),
							commentType: keys.IndexCommentType,
						},
					)
				}
			} else {
				__antithesis_instrumentation__.Notify(577632)
			}
			__antithesis_instrumentation__.Notify(577575)

			if err := scTable.MakeMutationComplete(scTable.Mutations[m.MutationOrdinal()]); err != nil {
				__antithesis_instrumentation__.Notify(577633)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577634)
			}
			__antithesis_instrumentation__.Notify(577576)

			if modify := m.AsModifyRowLevelTTL(); modify != nil {
				__antithesis_instrumentation__.Notify(577635)
				if fn := sc.testingKnobs.RunBeforeModifyRowLevelTTL; fn != nil {
					__antithesis_instrumentation__.Notify(577637)
					if err := fn(); err != nil {
						__antithesis_instrumentation__.Notify(577638)
						return err
					} else {
						__antithesis_instrumentation__.Notify(577639)
					}
				} else {
					__antithesis_instrumentation__.Notify(577640)
				}
				__antithesis_instrumentation__.Notify(577636)
				if m.Adding() {
					__antithesis_instrumentation__.Notify(577641)
					scTable.RowLevelTTL = modify.RowLevelTTL()
					shouldCreateScheduledJob := scTable.RowLevelTTL.ScheduleID == 0

					if scTable.RowLevelTTL.ScheduleID != 0 {
						__antithesis_instrumentation__.Notify(577643)
						_, err := jobs.LoadScheduledJob(
							ctx,
							JobSchedulerEnv(sc.execCfg),
							scTable.RowLevelTTL.ScheduleID,
							sc.execCfg.InternalExecutor,
							txn,
						)
						if err != nil {
							__antithesis_instrumentation__.Notify(577644)
							if !jobs.HasScheduledJobNotFoundError(err) {
								__antithesis_instrumentation__.Notify(577646)
								return errors.Wrapf(err, "unknown error fetching existing job for row level TTL in schema changer")
							} else {
								__antithesis_instrumentation__.Notify(577647)
							}
							__antithesis_instrumentation__.Notify(577645)
							shouldCreateScheduledJob = true
						} else {
							__antithesis_instrumentation__.Notify(577648)
						}
					} else {
						__antithesis_instrumentation__.Notify(577649)
					}
					__antithesis_instrumentation__.Notify(577642)

					if shouldCreateScheduledJob {
						__antithesis_instrumentation__.Notify(577650)
						j, err := CreateRowLevelTTLScheduledJob(
							ctx,
							sc.execCfg,
							txn,
							getOwnerOfDesc(scTable),
							scTable.GetID(),
							modify.RowLevelTTL(),
						)
						if err != nil {
							__antithesis_instrumentation__.Notify(577652)
							return err
						} else {
							__antithesis_instrumentation__.Notify(577653)
						}
						__antithesis_instrumentation__.Notify(577651)
						scTable.RowLevelTTL.ScheduleID = j.ScheduleID()
					} else {
						__antithesis_instrumentation__.Notify(577654)
					}
				} else {
					__antithesis_instrumentation__.Notify(577655)
					if m.Dropped() {
						__antithesis_instrumentation__.Notify(577656)
						if ttl := scTable.RowLevelTTL; ttl != nil {
							__antithesis_instrumentation__.Notify(577658)
							if err := DeleteSchedule(ctx, sc.execCfg, txn, ttl.ScheduleID); err != nil {
								__antithesis_instrumentation__.Notify(577659)
								return err
							} else {
								__antithesis_instrumentation__.Notify(577660)
							}
						} else {
							__antithesis_instrumentation__.Notify(577661)
						}
						__antithesis_instrumentation__.Notify(577657)
						scTable.RowLevelTTL = nil
					} else {
						__antithesis_instrumentation__.Notify(577662)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(577663)
			}
			__antithesis_instrumentation__.Notify(577577)

			if pkSwap := m.AsPrimaryKeySwap(); pkSwap != nil {
				__antithesis_instrumentation__.Notify(577664)
				if fn := sc.testingKnobs.RunBeforePrimaryKeySwap; fn != nil {
					__antithesis_instrumentation__.Notify(577668)
					fn()
				} else {
					__antithesis_instrumentation__.Notify(577669)
				}
				__antithesis_instrumentation__.Notify(577665)

				if pkSwap.HasLocalityConfig() {
					__antithesis_instrumentation__.Notify(577670)
					lcSwap := pkSwap.LocalityConfigSwap()
					localityConfigToSwapTo := lcSwap.NewLocalityConfig
					if m.Adding() {
						__antithesis_instrumentation__.Notify(577673)

						if !scTable.LocalityConfig.Equal(lcSwap.OldLocalityConfig) {
							__antithesis_instrumentation__.Notify(577675)
							return errors.AssertionFailedf(
								"expected locality on table to match old locality\ngot: %s\nwant %s",
								scTable.LocalityConfig,
								lcSwap.OldLocalityConfig,
							)
						} else {
							__antithesis_instrumentation__.Notify(577676)
						}
						__antithesis_instrumentation__.Notify(577674)

						if colID := lcSwap.NewRegionalByRowColumnID; colID != nil {
							__antithesis_instrumentation__.Notify(577677)
							col, err := scTable.FindColumnWithID(*colID)
							if err != nil {
								__antithesis_instrumentation__.Notify(577679)
								return err
							} else {
								__antithesis_instrumentation__.Notify(577680)
							}
							__antithesis_instrumentation__.Notify(577678)
							col.ColumnDesc().DefaultExpr = lcSwap.NewRegionalByRowColumnDefaultExpr
						} else {
							__antithesis_instrumentation__.Notify(577681)
						}

					} else {
						__antithesis_instrumentation__.Notify(577682)

						localityConfigToSwapTo = lcSwap.OldLocalityConfig
					}
					__antithesis_instrumentation__.Notify(577671)

					if err := setNewLocalityConfig(
						ctx, scTable, txn, b, localityConfigToSwapTo, kvTrace, descsCol); err != nil {
						__antithesis_instrumentation__.Notify(577683)
						return err
					} else {
						__antithesis_instrumentation__.Notify(577684)
					}
					__antithesis_instrumentation__.Notify(577672)
					switch localityConfigToSwapTo.Locality.(type) {
					case *catpb.LocalityConfig_RegionalByTable_,
						*catpb.LocalityConfig_Global_:
						__antithesis_instrumentation__.Notify(577685)
						scTable.PartitionAllBy = false
					case *catpb.LocalityConfig_RegionalByRow_:
						__antithesis_instrumentation__.Notify(577686)
						scTable.PartitionAllBy = true
					default:
						__antithesis_instrumentation__.Notify(577687)
						return errors.AssertionFailedf(
							"unknown locality on PK swap: %T",
							localityConfigToSwapTo,
						)
					}
				} else {
					__antithesis_instrumentation__.Notify(577688)
				}
				__antithesis_instrumentation__.Notify(577666)

				jobID, err := sc.queueCleanupJob(ctx, scTable, txn)
				if err != nil {
					__antithesis_instrumentation__.Notify(577689)
					return err
				} else {
					__antithesis_instrumentation__.Notify(577690)
				}
				__antithesis_instrumentation__.Notify(577667)
				if jobID > 0 {
					__antithesis_instrumentation__.Notify(577691)
					depMutationJobs = append(depMutationJobs, jobID)
				} else {
					__antithesis_instrumentation__.Notify(577692)
				}
			} else {
				__antithesis_instrumentation__.Notify(577693)
			}
			__antithesis_instrumentation__.Notify(577578)

			if m.AsComputedColumnSwap() != nil {
				__antithesis_instrumentation__.Notify(577694)
				if fn := sc.testingKnobs.RunBeforeComputedColumnSwap; fn != nil {
					__antithesis_instrumentation__.Notify(577697)
					fn()
				} else {
					__antithesis_instrumentation__.Notify(577698)
				}
				__antithesis_instrumentation__.Notify(577695)

				jobID, err := sc.queueCleanupJob(ctx, scTable, txn)
				if err != nil {
					__antithesis_instrumentation__.Notify(577699)
					return err
				} else {
					__antithesis_instrumentation__.Notify(577700)
				}
				__antithesis_instrumentation__.Notify(577696)
				if jobID > 0 {
					__antithesis_instrumentation__.Notify(577701)
					depMutationJobs = append(depMutationJobs, jobID)
				} else {
					__antithesis_instrumentation__.Notify(577702)
				}
			} else {
				__antithesis_instrumentation__.Notify(577703)
			}
			__antithesis_instrumentation__.Notify(577579)
			didUpdate = true
			i++
		}
		__antithesis_instrumentation__.Notify(577542)
		if didUpdate = i > 0; !didUpdate {
			__antithesis_instrumentation__.Notify(577704)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(577705)
		}
		__antithesis_instrumentation__.Notify(577543)
		committedMutations := scTable.AllMutations()[:i]

		scTable.Mutations = scTable.Mutations[i:]

		existingDepMutationJobs, err := sc.getDependentMutationsJobs(ctx, scTable, committedMutations)
		if err != nil {
			__antithesis_instrumentation__.Notify(577706)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577707)
		}
		__antithesis_instrumentation__.Notify(577544)
		depMutationJobs = append(depMutationJobs, existingDepMutationJobs...)

		for i, g := range scTable.MutationJobs {
			__antithesis_instrumentation__.Notify(577708)
			if g.MutationID == sc.mutationID {
				__antithesis_instrumentation__.Notify(577709)

				scTable.MutationJobs = append(scTable.MutationJobs[:i], scTable.MutationJobs[i+1:]...)
				break
			} else {
				__antithesis_instrumentation__.Notify(577710)
			}
		}
		__antithesis_instrumentation__.Notify(577545)

		if !scTable.Dropped() {
			__antithesis_instrumentation__.Notify(577711)
			newReferencedTypeIDs, err := collectReferencedTypeIDs()
			if err != nil {
				__antithesis_instrumentation__.Notify(577715)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577716)
			}
			__antithesis_instrumentation__.Notify(577712)
			update := make(map[descpb.ID]bool, newReferencedTypeIDs.Len()+referencedTypeIDs.Len())
			newReferencedTypeIDs.ForEach(func(id descpb.ID) {
				__antithesis_instrumentation__.Notify(577717)
				if !referencedTypeIDs.Contains(id) {
					__antithesis_instrumentation__.Notify(577718)

					update[id] = true
				} else {
					__antithesis_instrumentation__.Notify(577719)
				}
			})
			__antithesis_instrumentation__.Notify(577713)
			referencedTypeIDs.ForEach(func(id descpb.ID) {
				__antithesis_instrumentation__.Notify(577720)
				if !newReferencedTypeIDs.Contains(id) {
					__antithesis_instrumentation__.Notify(577721)

					update[id] = false
				} else {
					__antithesis_instrumentation__.Notify(577722)
				}
			})
			__antithesis_instrumentation__.Notify(577714)

			for id, isAddition := range update {
				__antithesis_instrumentation__.Notify(577723)
				typ, err := descsCol.GetMutableTypeVersionByID(ctx, txn, id)
				if err != nil {
					__antithesis_instrumentation__.Notify(577726)
					return err
				} else {
					__antithesis_instrumentation__.Notify(577727)
				}
				__antithesis_instrumentation__.Notify(577724)
				if isAddition {
					__antithesis_instrumentation__.Notify(577728)
					typ.AddReferencingDescriptorID(scTable.ID)
				} else {
					__antithesis_instrumentation__.Notify(577729)
					typ.RemoveReferencingDescriptorID(scTable.ID)
				}
				__antithesis_instrumentation__.Notify(577725)
				if err := descsCol.WriteDescToBatch(ctx, kvTrace, typ, b); err != nil {
					__antithesis_instrumentation__.Notify(577730)
					return err
				} else {
					__antithesis_instrumentation__.Notify(577731)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(577732)
		}
		__antithesis_instrumentation__.Notify(577546)

		if !scTable.Dropped() {
			__antithesis_instrumentation__.Notify(577733)
			newReferencedSequenceIDs := collectReferencedSequenceIDs()
			update := make(map[descpb.ID]catalog.TableColSet, len(newReferencedSequenceIDs)+len(referencedSequenceIDs))
			for id := range referencedSequenceIDs {
				__antithesis_instrumentation__.Notify(577736)
				if _, found := newReferencedSequenceIDs[id]; !found {
					__antithesis_instrumentation__.Notify(577737)

					update[id] = catalog.TableColSet{}
				} else {
					__antithesis_instrumentation__.Notify(577738)
				}
			}
			__antithesis_instrumentation__.Notify(577734)
			for id, newColIDs := range newReferencedSequenceIDs {
				__antithesis_instrumentation__.Notify(577739)
				newColIDSet := catalog.MakeTableColSet(newColIDs...)
				var oldColIDSet catalog.TableColSet
				if oldColIDs, found := referencedSequenceIDs[id]; found {
					__antithesis_instrumentation__.Notify(577741)
					oldColIDSet = catalog.MakeTableColSet(oldColIDs...)
				} else {
					__antithesis_instrumentation__.Notify(577742)
				}
				__antithesis_instrumentation__.Notify(577740)
				union := catalog.MakeTableColSet(newColIDs...)
				union.UnionWith(oldColIDSet)
				if union.Len() != oldColIDSet.Len() || func() bool {
					__antithesis_instrumentation__.Notify(577743)
					return union.Len() != newColIDSet.Len() == true
				}() == true {
					__antithesis_instrumentation__.Notify(577744)

					update[id] = newColIDSet
				} else {
					__antithesis_instrumentation__.Notify(577745)
				}
			}
			__antithesis_instrumentation__.Notify(577735)

			for id, colIDSet := range update {
				__antithesis_instrumentation__.Notify(577746)
				tbl, err := descsCol.GetMutableTableVersionByID(ctx, id, txn)
				if err != nil {
					__antithesis_instrumentation__.Notify(577748)
					return err
				} else {
					__antithesis_instrumentation__.Notify(577749)
				}
				__antithesis_instrumentation__.Notify(577747)
				tbl.UpdateColumnsDependedOnBy(scTable.ID, colIDSet)
				if err := descsCol.WriteDescToBatch(ctx, kvTrace, tbl, b); err != nil {
					__antithesis_instrumentation__.Notify(577750)
					return err
				} else {
					__antithesis_instrumentation__.Notify(577751)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(577752)
		}
		__antithesis_instrumentation__.Notify(577547)

		metaDataUpdater := sc.execCfg.DescMetadaUpdaterFactory.NewMetadataUpdater(
			ctx,
			txn,
			NewFakeSessionData(&sc.settings.SV))
		for _, comment := range commentsToDelete {
			__antithesis_instrumentation__.Notify(577753)
			err := metaDataUpdater.DeleteDescriptorComment(
				comment.id,
				comment.subID,
				comment.commentType)
			if err != nil {
				__antithesis_instrumentation__.Notify(577754)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577755)
			}
		}
		__antithesis_instrumentation__.Notify(577548)
		for _, comment := range commentsToSwap {
			__antithesis_instrumentation__.Notify(577756)
			err := metaDataUpdater.SwapDescriptorSubComment(
				comment.id,
				comment.oldSubID,
				comment.newSubID,
				comment.commentType,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(577757)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577758)
			}
		}
		__antithesis_instrumentation__.Notify(577549)

		if err := descsCol.WriteDescToBatch(ctx, kvTrace, scTable, b); err != nil {
			__antithesis_instrumentation__.Notify(577759)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577760)
		}
		__antithesis_instrumentation__.Notify(577550)
		if err := txn.Run(ctx, b); err != nil {
			__antithesis_instrumentation__.Notify(577761)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577762)
		}
		__antithesis_instrumentation__.Notify(577551)

		var info eventpb.EventPayload
		if isRollback {
			__antithesis_instrumentation__.Notify(577763)
			info = &eventpb.FinishSchemaChangeRollback{}
		} else {
			__antithesis_instrumentation__.Notify(577764)
			info = &eventpb.FinishSchemaChange{}
		}
		__antithesis_instrumentation__.Notify(577552)

		return logEventInternalForSchemaChanges(
			ctx, sc.execCfg, txn,
			sc.sqlInstanceID,
			sc.descID,
			sc.mutationID,
			info)
	})
	__antithesis_instrumentation__.Notify(577533)
	if err != nil {
		__antithesis_instrumentation__.Notify(577765)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577766)
	}
	__antithesis_instrumentation__.Notify(577534)

	if err := sc.jobRegistry.Run(ctx, sc.execCfg.InternalExecutor, depMutationJobs); err != nil {
		__antithesis_instrumentation__.Notify(577767)
		return errors.Wrap(err, "A dependent transaction failed for this schema change")
	} else {
		__antithesis_instrumentation__.Notify(577768)
	}
	__antithesis_instrumentation__.Notify(577535)

	return nil
}

func maybeUpdateZoneConfigsForPKChange(
	ctx context.Context,
	txn *kv.Txn,
	execCfg *ExecutorConfig,
	table *tabledesc.Mutable,
	swapInfo *descpb.PrimaryKeySwap,
) error {
	__antithesis_instrumentation__.Notify(577769)
	zone, err := getZoneConfigRaw(ctx, txn, execCfg.Codec, execCfg.Settings, table.ID)
	if err != nil {
		__antithesis_instrumentation__.Notify(577776)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577777)
	}
	__antithesis_instrumentation__.Notify(577770)

	if zone == nil {
		__antithesis_instrumentation__.Notify(577778)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(577779)
	}
	__antithesis_instrumentation__.Notify(577771)

	oldIdxToNewIdx := make(map[descpb.IndexID]descpb.IndexID)
	for i, oldID := range swapInfo.OldIndexes {
		__antithesis_instrumentation__.Notify(577780)
		oldIdxToNewIdx[oldID] = swapInfo.NewIndexes[i]
	}
	__antithesis_instrumentation__.Notify(577772)

	if table.IsLocalityRegionalByRow() {
		__antithesis_instrumentation__.Notify(577781)
		oldIdxToNewIdx[swapInfo.OldPrimaryIndexId] = swapInfo.NewPrimaryIndexId
	} else {
		__antithesis_instrumentation__.Notify(577782)
	}
	__antithesis_instrumentation__.Notify(577773)

	for oldIdx, newIdx := range oldIdxToNewIdx {
		__antithesis_instrumentation__.Notify(577783)
		for i := range zone.Subzones {
			__antithesis_instrumentation__.Notify(577784)
			subzone := &zone.Subzones[i]
			if subzone.IndexID == uint32(oldIdx) {
				__antithesis_instrumentation__.Notify(577785)

				subzoneCopy := *subzone
				subzoneCopy.IndexID = uint32(newIdx)
				zone.SetSubzone(subzoneCopy)
			} else {
				__antithesis_instrumentation__.Notify(577786)
			}
		}
	}
	__antithesis_instrumentation__.Notify(577774)

	_, err = writeZoneConfig(ctx, txn, table.ID, table, zone, execCfg, false)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(577787)
		return !sqlerrors.IsCCLRequiredError(err) == true
	}() == true {
		__antithesis_instrumentation__.Notify(577788)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577789)
	}
	__antithesis_instrumentation__.Notify(577775)

	return nil
}

func (sc *SchemaChanger) runStateMachineAndBackfill(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(577790)
	if fn := sc.testingKnobs.RunBeforePublishWriteAndDelete; fn != nil {
		__antithesis_instrumentation__.Notify(577795)
		fn()
	} else {
		__antithesis_instrumentation__.Notify(577796)
	}
	__antithesis_instrumentation__.Notify(577791)

	if err := sc.preSplitHashShardedIndexRanges(ctx); err != nil {
		__antithesis_instrumentation__.Notify(577797)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577798)
	}
	__antithesis_instrumentation__.Notify(577792)

	if err := sc.RunStateMachineBeforeBackfill(ctx); err != nil {
		__antithesis_instrumentation__.Notify(577799)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577800)
	}
	__antithesis_instrumentation__.Notify(577793)

	if err := sc.runBackfill(ctx); err != nil {
		__antithesis_instrumentation__.Notify(577801)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577802)
	}
	__antithesis_instrumentation__.Notify(577794)

	log.Info(ctx, "marking schema change as complete")
	return sc.done(ctx)
}

func (sc *SchemaChanger) refreshStats(desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(577803)

	if tableDesc, ok := desc.(catalog.TableDescriptor); ok {
		__antithesis_instrumentation__.Notify(577804)
		sc.execCfg.StatsRefresher.NotifyMutation(tableDesc, math.MaxInt32)
	} else {
		__antithesis_instrumentation__.Notify(577805)
	}
}

func (sc *SchemaChanger) maybeReverseMutations(ctx context.Context, causingError error) error {
	__antithesis_instrumentation__.Notify(577806)
	if fn := sc.testingKnobs.RunBeforeMutationReversal; fn != nil {
		__antithesis_instrumentation__.Notify(577811)
		if err := fn(sc.job.ID()); err != nil {
			__antithesis_instrumentation__.Notify(577812)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577813)
		}
	} else {
		__antithesis_instrumentation__.Notify(577814)
	}
	__antithesis_instrumentation__.Notify(577807)

	if sc.mutationID == descpb.InvalidMutationID {
		__antithesis_instrumentation__.Notify(577815)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(577816)
	}
	__antithesis_instrumentation__.Notify(577808)

	alreadyReversed := false
	const kvTrace = true
	err := sc.txn(ctx, func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
		__antithesis_instrumentation__.Notify(577817)
		scTable, err := descsCol.GetMutableTableVersionByID(ctx, sc.descID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(577827)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577828)
		}
		__antithesis_instrumentation__.Notify(577818)

		if len(scTable.Mutations) == 0 {
			__antithesis_instrumentation__.Notify(577829)
			alreadyReversed = true
		} else {
			__antithesis_instrumentation__.Notify(577830)
			if scTable.Mutations[0].MutationID != sc.mutationID {
				__antithesis_instrumentation__.Notify(577831)
				var found bool
				for i := range scTable.Mutations {
					__antithesis_instrumentation__.Notify(577833)
					if found = scTable.Mutations[i].MutationID == sc.mutationID; found {
						__antithesis_instrumentation__.Notify(577834)
						break
					} else {
						__antithesis_instrumentation__.Notify(577835)
					}
				}
				__antithesis_instrumentation__.Notify(577832)
				if alreadyReversed = !found; !alreadyReversed {
					__antithesis_instrumentation__.Notify(577836)
					return errors.AssertionFailedf("expected mutation %d to be the"+
						" first mutation when reverted, found %d in descriptor %d",
						sc.mutationID, scTable.Mutations[0].MutationID, scTable.ID)
				} else {
					__antithesis_instrumentation__.Notify(577837)
				}
			} else {
				__antithesis_instrumentation__.Notify(577838)
				if scTable.Mutations[0].Rollback {
					__antithesis_instrumentation__.Notify(577839)
					alreadyReversed = true
				} else {
					__antithesis_instrumentation__.Notify(577840)
				}
			}
		}
		__antithesis_instrumentation__.Notify(577819)

		if alreadyReversed {
			__antithesis_instrumentation__.Notify(577841)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(577842)
		}
		__antithesis_instrumentation__.Notify(577820)

		var droppedMutations map[descpb.MutationID]struct{}

		columns := make(map[string]struct{})
		b := txn.NewBatch()
		for _, m := range scTable.AllMutations() {
			__antithesis_instrumentation__.Notify(577843)
			if m.MutationID() != sc.mutationID {
				__antithesis_instrumentation__.Notify(577849)
				break
			} else {
				__antithesis_instrumentation__.Notify(577850)
			}
			__antithesis_instrumentation__.Notify(577844)

			if m.IsRollback() {
				__antithesis_instrumentation__.Notify(577851)

				return errors.AssertionFailedf("mutation already rolled back: %v", scTable.Mutations[m.MutationOrdinal()])
			} else {
				__antithesis_instrumentation__.Notify(577852)
			}
			__antithesis_instrumentation__.Notify(577845)

			if discarded, _ := isCurrentMutationDiscarded(scTable, m, m.MutationOrdinal()+1); discarded {
				__antithesis_instrumentation__.Notify(577853)
				continue
			} else {
				__antithesis_instrumentation__.Notify(577854)
			}
			__antithesis_instrumentation__.Notify(577846)

			if idx := m.AsIndex(); idx != nil && func() bool {
				__antithesis_instrumentation__.Notify(577855)
				return idx.IsTemporaryIndexForBackfill() == true
			}() == true {
				__antithesis_instrumentation__.Notify(577856)
				scTable.Mutations[m.MutationOrdinal()].State = descpb.DescriptorMutation_DELETE_ONLY
				scTable.Mutations[m.MutationOrdinal()].Direction = descpb.DescriptorMutation_DROP
			} else {
				__antithesis_instrumentation__.Notify(577857)
				log.Warningf(ctx, "reverse schema change mutation: %+v", scTable.Mutations[m.MutationOrdinal()])
				scTable.Mutations[m.MutationOrdinal()], columns = sc.reverseMutation(scTable.Mutations[m.MutationOrdinal()], false, columns)
			}
			__antithesis_instrumentation__.Notify(577847)

			if constraint := m.AsConstraint(); constraint != nil && func() bool {
				__antithesis_instrumentation__.Notify(577858)
				return constraint.Adding() == true
			}() == true {
				__antithesis_instrumentation__.Notify(577859)
				log.Warningf(ctx, "dropping constraint %+v", constraint.ConstraintToUpdateDesc())
				if err := sc.maybeDropValidatingConstraint(ctx, scTable, constraint); err != nil {
					__antithesis_instrumentation__.Notify(577861)
					return err
				} else {
					__antithesis_instrumentation__.Notify(577862)
				}
				__antithesis_instrumentation__.Notify(577860)

				if constraint.IsForeignKey() {
					__antithesis_instrumentation__.Notify(577863)
					backrefTable, err := descsCol.GetMutableTableVersionByID(ctx, constraint.ForeignKey().ReferencedTableID, txn)
					if err != nil {
						__antithesis_instrumentation__.Notify(577866)
						return err
					} else {
						__antithesis_instrumentation__.Notify(577867)
					}
					__antithesis_instrumentation__.Notify(577864)
					if err := removeFKBackReferenceFromTable(backrefTable, constraint.GetName(), scTable); err != nil {
						__antithesis_instrumentation__.Notify(577868)

						log.Infof(ctx,
							"error attempting to remove backreference %s during rollback: %s", constraint.GetName(), err)
					} else {
						__antithesis_instrumentation__.Notify(577869)
					}
					__antithesis_instrumentation__.Notify(577865)
					if err := descsCol.WriteDescToBatch(ctx, kvTrace, backrefTable, b); err != nil {
						__antithesis_instrumentation__.Notify(577870)
						return err
					} else {
						__antithesis_instrumentation__.Notify(577871)
					}
				} else {
					__antithesis_instrumentation__.Notify(577872)
				}
			} else {
				__antithesis_instrumentation__.Notify(577873)
			}
			__antithesis_instrumentation__.Notify(577848)
			scTable.Mutations[m.MutationOrdinal()].Rollback = true
		}
		__antithesis_instrumentation__.Notify(577821)

		if len(columns) > 0 {
			__antithesis_instrumentation__.Notify(577874)
			var err error
			droppedMutations, err = sc.deleteIndexMutationsWithReversedColumns(ctx, scTable, columns)
			if err != nil {
				__antithesis_instrumentation__.Notify(577875)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577876)
			}
		} else {
			__antithesis_instrumentation__.Notify(577877)
		}
		__antithesis_instrumentation__.Notify(577822)

		if err := descsCol.WriteDescToBatch(ctx, kvTrace, scTable, b); err != nil {
			__antithesis_instrumentation__.Notify(577878)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577879)
		}
		__antithesis_instrumentation__.Notify(577823)
		if err := txn.Run(ctx, b); err != nil {
			__antithesis_instrumentation__.Notify(577880)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577881)
		}
		__antithesis_instrumentation__.Notify(577824)

		tableDesc := scTable.ImmutableCopy().(catalog.TableDescriptor)

		err = sc.updateJobForRollback(ctx, txn, tableDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(577882)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577883)
		}
		__antithesis_instrumentation__.Notify(577825)

		for m := range droppedMutations {
			__antithesis_instrumentation__.Notify(577884)
			jobID, err := getJobIDForMutationWithDescriptor(ctx, tableDesc, m)
			if err != nil {
				__antithesis_instrumentation__.Notify(577886)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577887)
			}
			__antithesis_instrumentation__.Notify(577885)
			if err := sc.jobRegistry.Failed(ctx, txn, jobID, causingError); err != nil {
				__antithesis_instrumentation__.Notify(577888)
				return err
			} else {
				__antithesis_instrumentation__.Notify(577889)
			}
		}
		__antithesis_instrumentation__.Notify(577826)

		return logEventInternalForSchemaChanges(
			ctx, sc.execCfg, txn,
			sc.sqlInstanceID,
			sc.descID,
			sc.mutationID,
			&eventpb.ReverseSchemaChange{
				Error:    fmt.Sprintf("%+v", causingError),
				SQLSTATE: pgerror.GetPGCode(causingError).String(),
			})
	})
	__antithesis_instrumentation__.Notify(577809)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(577890)
		return alreadyReversed == true
	}() == true {
		__antithesis_instrumentation__.Notify(577891)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577892)
	}
	__antithesis_instrumentation__.Notify(577810)
	return nil
}

func (sc *SchemaChanger) updateJobForRollback(
	ctx context.Context, txn *kv.Txn, tableDesc catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(577893)

	span := tableDesc.PrimaryIndexSpan(sc.execCfg.Codec)
	var spanList []jobspb.ResumeSpanList
	for _, m := range tableDesc.AllMutations() {
		__antithesis_instrumentation__.Notify(577897)
		if m.MutationID() == sc.mutationID {
			__antithesis_instrumentation__.Notify(577898)
			spanList = append(spanList,
				jobspb.ResumeSpanList{
					ResumeSpans: []roachpb.Span{span},
				},
			)
		} else {
			__antithesis_instrumentation__.Notify(577899)
		}
	}
	__antithesis_instrumentation__.Notify(577894)
	oldDetails := sc.job.Details().(jobspb.SchemaChangeDetails)
	if err := sc.job.SetDetails(
		ctx, txn, jobspb.SchemaChangeDetails{
			DescID:          sc.descID,
			TableMutationID: sc.mutationID,
			ResumeSpanList:  spanList,
			FormatVersion:   oldDetails.FormatVersion,
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(577900)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577901)
	}
	__antithesis_instrumentation__.Notify(577895)
	if err := sc.job.SetProgress(ctx, txn, jobspb.SchemaChangeProgress{}); err != nil {
		__antithesis_instrumentation__.Notify(577902)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577903)
	}
	__antithesis_instrumentation__.Notify(577896)

	return nil
}

func (sc *SchemaChanger) maybeDropValidatingConstraint(
	ctx context.Context, desc *tabledesc.Mutable, constraint catalog.ConstraintToUpdate,
) error {
	__antithesis_instrumentation__.Notify(577904)
	if constraint.IsCheck() || func() bool {
		__antithesis_instrumentation__.Notify(577906)
		return constraint.IsNotNull() == true
	}() == true {
		__antithesis_instrumentation__.Notify(577907)
		if constraint.Check().Validity == descpb.ConstraintValidity_Unvalidated {
			__antithesis_instrumentation__.Notify(577910)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(577911)
		}
		__antithesis_instrumentation__.Notify(577908)
		for j, c := range desc.Checks {
			__antithesis_instrumentation__.Notify(577912)
			if c.Name == constraint.Check().Name {
				__antithesis_instrumentation__.Notify(577913)
				desc.Checks = append(desc.Checks[:j], desc.Checks[j+1:]...)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(577914)
			}
		}
		__antithesis_instrumentation__.Notify(577909)
		log.Infof(
			ctx,
			"attempted to drop constraint %s, but it hadn't been added to the table descriptor yet",
			constraint.Check().Name,
		)
	} else {
		__antithesis_instrumentation__.Notify(577915)
		if constraint.IsForeignKey() {
			__antithesis_instrumentation__.Notify(577916)
			for i, fk := range desc.OutboundFKs {
				__antithesis_instrumentation__.Notify(577918)
				if fk.Name == constraint.ForeignKey().Name {
					__antithesis_instrumentation__.Notify(577919)
					desc.OutboundFKs = append(desc.OutboundFKs[:i], desc.OutboundFKs[i+1:]...)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(577920)
				}
			}
			__antithesis_instrumentation__.Notify(577917)
			log.Infof(
				ctx,
				"attempted to drop constraint %s, but it hadn't been added to the table descriptor yet",
				constraint.ForeignKey().Name,
			)
		} else {
			__antithesis_instrumentation__.Notify(577921)
			if constraint.IsUniqueWithoutIndex() {
				__antithesis_instrumentation__.Notify(577922)
				if constraint.UniqueWithoutIndex().Validity == descpb.ConstraintValidity_Unvalidated {
					__antithesis_instrumentation__.Notify(577925)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(577926)
				}
				__antithesis_instrumentation__.Notify(577923)
				for j, c := range desc.UniqueWithoutIndexConstraints {
					__antithesis_instrumentation__.Notify(577927)
					if c.Name == constraint.UniqueWithoutIndex().Name {
						__antithesis_instrumentation__.Notify(577928)
						desc.UniqueWithoutIndexConstraints = append(
							desc.UniqueWithoutIndexConstraints[:j], desc.UniqueWithoutIndexConstraints[j+1:]...,
						)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(577929)
					}
				}
				__antithesis_instrumentation__.Notify(577924)
				log.Infof(
					ctx,
					"attempted to drop constraint %s, but it hadn't been added to the table descriptor yet",
					constraint.UniqueWithoutIndex().Name,
				)
			} else {
				__antithesis_instrumentation__.Notify(577930)
				return errors.AssertionFailedf("unsupported constraint type: %d", constraint.ConstraintToUpdateDesc().ConstraintType)
			}
		}
	}
	__antithesis_instrumentation__.Notify(577905)
	return nil
}

func (sc *SchemaChanger) startingStateForAddIndexMutations() descpb.DescriptorMutation_State {
	__antithesis_instrumentation__.Notify(577931)
	if sc.mvccCompliantAddIndex {
		__antithesis_instrumentation__.Notify(577933)
		return descpb.DescriptorMutation_BACKFILLING
	} else {
		__antithesis_instrumentation__.Notify(577934)
	}
	__antithesis_instrumentation__.Notify(577932)
	return descpb.DescriptorMutation_DELETE_ONLY
}

func (sc *SchemaChanger) deleteIndexMutationsWithReversedColumns(
	ctx context.Context, desc *tabledesc.Mutable, columns map[string]struct{},
) (map[descpb.MutationID]struct{}, error) {
	__antithesis_instrumentation__.Notify(577935)
	dropMutations := make(map[descpb.MutationID]struct{})

	for {
		__antithesis_instrumentation__.Notify(577937)
		start := len(dropMutations)
		for _, mutation := range desc.Mutations {
			__antithesis_instrumentation__.Notify(577941)
			if mutation.MutationID != sc.mutationID {
				__antithesis_instrumentation__.Notify(577942)
				if idx := mutation.GetIndex(); idx != nil {
					__antithesis_instrumentation__.Notify(577943)
					for _, name := range idx.KeyColumnNames {
						__antithesis_instrumentation__.Notify(577944)
						if _, ok := columns[name]; ok {
							__antithesis_instrumentation__.Notify(577945)

							if mutation.Direction != descpb.DescriptorMutation_ADD || func() bool {
								__antithesis_instrumentation__.Notify(577947)
								return mutation.State != sc.startingStateForAddIndexMutations() == true
							}() == true {
								__antithesis_instrumentation__.Notify(577948)
								panic(errors.AssertionFailedf("mutation in bad state: %+v", mutation))
							} else {
								__antithesis_instrumentation__.Notify(577949)
							}
							__antithesis_instrumentation__.Notify(577946)
							log.Warningf(ctx, "drop schema change mutation: %+v", mutation)
							dropMutations[mutation.MutationID] = struct{}{}
							break
						} else {
							__antithesis_instrumentation__.Notify(577950)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(577951)
				}
			} else {
				__antithesis_instrumentation__.Notify(577952)
			}
		}
		__antithesis_instrumentation__.Notify(577938)

		if len(dropMutations) == start {
			__antithesis_instrumentation__.Notify(577953)

			break
		} else {
			__antithesis_instrumentation__.Notify(577954)
		}
		__antithesis_instrumentation__.Notify(577939)

		newMutations := make([]descpb.DescriptorMutation, 0, len(desc.Mutations))
		for _, m := range desc.AllMutations() {
			__antithesis_instrumentation__.Notify(577955)
			mutation := desc.Mutations[m.MutationOrdinal()]
			if _, ok := dropMutations[m.MutationID()]; ok {
				__antithesis_instrumentation__.Notify(577956)

				mutation, columns = sc.reverseMutation(mutation, true, columns)

				if err := desc.MakeMutationComplete(mutation); err != nil {
					__antithesis_instrumentation__.Notify(577957)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(577958)
				}
			} else {
				__antithesis_instrumentation__.Notify(577959)
				newMutations = append(newMutations, mutation)
			}
		}
		__antithesis_instrumentation__.Notify(577940)

		desc.Mutations = newMutations
	}
	__antithesis_instrumentation__.Notify(577936)
	return dropMutations, nil
}

func (sc *SchemaChanger) reverseMutation(
	mutation descpb.DescriptorMutation, notStarted bool, columns map[string]struct{},
) (descpb.DescriptorMutation, map[string]struct{}) {
	__antithesis_instrumentation__.Notify(577960)
	switch mutation.Direction {
	case descpb.DescriptorMutation_ADD:
		__antithesis_instrumentation__.Notify(577962)
		mutation.Direction = descpb.DescriptorMutation_DROP

		if col := mutation.GetColumn(); col != nil {
			__antithesis_instrumentation__.Notify(577968)
			columns[col.Name] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(577969)
		}
		__antithesis_instrumentation__.Notify(577963)

		if pkSwap, computedColumnsSwap, refresh :=
			mutation.GetPrimaryKeySwap(),
			mutation.GetComputedColumnSwap(),
			mutation.GetMaterializedViewRefresh(); pkSwap != nil || func() bool {
			__antithesis_instrumentation__.Notify(577970)
			return computedColumnsSwap != nil == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(577971)
			return refresh != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(577972)
			return mutation, columns
		} else {
			__antithesis_instrumentation__.Notify(577973)
		}
		__antithesis_instrumentation__.Notify(577964)

		if notStarted {
			__antithesis_instrumentation__.Notify(577974)
			startingState := descpb.DescriptorMutation_DELETE_ONLY
			if idx := mutation.GetIndex(); idx != nil {
				__antithesis_instrumentation__.Notify(577976)
				startingState = sc.startingStateForAddIndexMutations()
			} else {
				__antithesis_instrumentation__.Notify(577977)
			}
			__antithesis_instrumentation__.Notify(577975)
			if mutation.State != startingState {
				__antithesis_instrumentation__.Notify(577978)
				panic(errors.AssertionFailedf("mutation in bad state: %+v", mutation))
			} else {
				__antithesis_instrumentation__.Notify(577979)
			}
		} else {
			__antithesis_instrumentation__.Notify(577980)
		}

	case descpb.DescriptorMutation_DROP:
		__antithesis_instrumentation__.Notify(577965)

		if mutation.GetIndex() != nil {
			__antithesis_instrumentation__.Notify(577981)
			return mutation, columns
		} else {
			__antithesis_instrumentation__.Notify(577982)
		}
		__antithesis_instrumentation__.Notify(577966)

		mutation.Direction = descpb.DescriptorMutation_ADD
		if notStarted && func() bool {
			__antithesis_instrumentation__.Notify(577983)
			return mutation.State != descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY == true
		}() == true {
			__antithesis_instrumentation__.Notify(577984)
			panic(errors.AssertionFailedf("mutation in bad state: %+v", mutation))
		} else {
			__antithesis_instrumentation__.Notify(577985)
		}
	default:
		__antithesis_instrumentation__.Notify(577967)
	}
	__antithesis_instrumentation__.Notify(577961)
	return mutation, columns
}

func CreateGCJobRecord(
	originalDescription string, username security.SQLUsername, details jobspb.SchemaChangeGCDetails,
) jobs.Record {
	__antithesis_instrumentation__.Notify(577986)
	descriptorIDs := make([]descpb.ID, 0)
	if len(details.Indexes) > 0 {
		__antithesis_instrumentation__.Notify(577988)
		if len(descriptorIDs) == 0 {
			__antithesis_instrumentation__.Notify(577989)
			descriptorIDs = []descpb.ID{details.ParentID}
		} else {
			__antithesis_instrumentation__.Notify(577990)
		}
	} else {
		__antithesis_instrumentation__.Notify(577991)
		for _, table := range details.Tables {
			__antithesis_instrumentation__.Notify(577992)
			descriptorIDs = append(descriptorIDs, table.ID)
		}
	}
	__antithesis_instrumentation__.Notify(577987)
	return jobs.Record{
		Description:   fmt.Sprintf("GC for %s", originalDescription),
		Username:      username,
		DescriptorIDs: descriptorIDs,
		Details:       details,
		Progress:      jobspb.SchemaChangeGCProgress{},
		RunningStatus: RunningStatusWaitingGC,
		NonCancelable: true,
	}
}

type GCJobTestingKnobs struct {
	RunBeforeResume    func(jobID jobspb.JobID) error
	RunBeforePerformGC func(jobID jobspb.JobID) error

	RunAfterIsProtectedCheck func(jobID jobspb.JobID, isProtected bool)
}

func (*GCJobTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(577993) }

type SchemaChangerTestingKnobs struct {
	SchemaChangeJobNoOp func() bool

	RunBeforePublishWriteAndDelete func()

	RunBeforeBackfill func() error

	RunAfterBackfill func(jobID jobspb.JobID) error

	RunBeforeQueryBackfill func() error

	RunBeforeIndexBackfill func()

	RunAfterIndexBackfill func()

	RunBeforeTempIndexMerge func()

	RunAfterTempIndexMerge func()

	RunBeforeMaterializedViewRefreshCommit func() error

	RunBeforePrimaryKeySwap func()

	RunBeforeComputedColumnSwap func()

	RunBeforeIndexValidation func() error

	RunBeforeConstraintValidation func(constraints []catalog.ConstraintToUpdate) error

	RunBeforeMutationReversal func(jobID jobspb.JobID) error

	RunAfterMutationReversal func(jobID jobspb.JobID) error

	RunBeforeOnFailOrCancel func(jobID jobspb.JobID) error

	RunAfterOnFailOrCancel func(jobID jobspb.JobID) error

	RunBeforeResume func(jobID jobspb.JobID) error

	RunBeforeDescTxn func(jobID jobspb.JobID) error

	OldNamesDrainedNotification func()

	WriteCheckpointInterval time.Duration

	BackfillChunkSize int64

	AlwaysUpdateIndexBackfillDetails bool

	AlwaysUpdateIndexBackfillProgress bool

	TwoVersionLeaseViolation func()

	RunBeforeHashShardedIndexRangePreSplit func(tbl *tabledesc.Mutable, kbDB *kv.DB, codec keys.SQLCodec) error

	RunAfterHashShardedIndexRangePreSplit func(tbl *tabledesc.Mutable, kbDB *kv.DB, codec keys.SQLCodec) error

	RunBeforeModifyRowLevelTTL func() error
}

func (*SchemaChangerTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(577994) }

func (sc *SchemaChanger) txn(
	ctx context.Context, f func(context.Context, *kv.Txn, *descs.Collection) error,
) error {
	__antithesis_instrumentation__.Notify(577995)
	if fn := sc.testingKnobs.RunBeforeDescTxn; fn != nil {
		__antithesis_instrumentation__.Notify(577997)
		if err := fn(sc.job.ID()); err != nil {
			__antithesis_instrumentation__.Notify(577998)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577999)
		}
	} else {
		__antithesis_instrumentation__.Notify(578000)
	}
	__antithesis_instrumentation__.Notify(577996)

	return sc.execCfg.CollectionFactory.Txn(ctx, sc.execCfg.InternalExecutor, sc.db, f)
}

func createSchemaChangeEvalCtx(
	ctx context.Context, execCfg *ExecutorConfig, ts hlc.Timestamp, descriptors *descs.Collection,
) extendedEvalContext {
	__antithesis_instrumentation__.Notify(578001)

	sd := NewFakeSessionData(execCfg.SV())

	evalCtx := extendedEvalContext{

		Tracing: &SessionTracing{},
		ExecCfg: execCfg,
		Descs:   descriptors,
		EvalContext: tree.EvalContext{
			SessionDataStack: sessiondata.NewStack(sd),

			Context:            ctx,
			Planner:            &faketreeeval.DummyEvalPlanner{},
			PrivilegedAccessor: &faketreeeval.DummyPrivilegedAccessor{},
			SessionAccessor:    &faketreeeval.DummySessionAccessor{},
			ClientNoticeSender: &faketreeeval.DummyClientNoticeSender{},
			Sequence:           &faketreeeval.DummySequenceOperators{},
			Tenant:             &faketreeeval.DummyTenantOperator{},
			Regions:            &faketreeeval.DummyRegionOperator{},
			Settings:           execCfg.Settings,
			TestingKnobs:       execCfg.EvalContextTestingKnobs,
			ClusterID:          execCfg.LogicalClusterID(),
			ClusterName:        execCfg.RPCContext.ClusterName(),
			NodeID:             execCfg.NodeID,
			Codec:              execCfg.Codec,
			Locality:           execCfg.Locality,
			Tracer:             execCfg.AmbientCtx.Tracer,
		},
	}

	evalCtx.SetTxnTimestamp(timeutil.Unix(0, ts.WallTime))
	evalCtx.SetStmtTimestamp(timeutil.Unix(0, ts.WallTime))

	return evalCtx
}

func NewFakeSessionData(sv *settings.Values) *sessiondata.SessionData {
	__antithesis_instrumentation__.Notify(578002)
	sd := &sessiondata.SessionData{
		SessionData: sessiondatapb.SessionData{

			Database:      "",
			UserProto:     security.NodeUserName().EncodeProto(),
			VectorizeMode: sessiondatapb.VectorizeExecMode(VectorizeClusterMode.Get(sv)),
			Internal:      true,
		},
		LocalOnlySessionData: sessiondatapb.LocalOnlySessionData{
			DistSQLMode: sessiondatapb.DistSQLExecMode(DistSQLClusterExecMode.Get(sv)),
		},
		SearchPath:    sessiondata.DefaultSearchPathForUser(security.NodeUserName()),
		SequenceState: sessiondata.NewSequenceState(),
		Location:      time.UTC,
	}

	return sd
}

type schemaChangeResumer struct {
	job *jobs.Job
}

func (r schemaChangeResumer) Resume(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(578003)
	p := execCtx.(JobExecContext)
	details := r.job.Details().(jobspb.SchemaChangeDetails)
	if p.ExecCfg().SchemaChangerTestingKnobs.SchemaChangeJobNoOp != nil && func() bool {
		__antithesis_instrumentation__.Notify(578013)
		return p.ExecCfg().SchemaChangerTestingKnobs.SchemaChangeJobNoOp() == true
	}() == true {
		__antithesis_instrumentation__.Notify(578014)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(578015)
	}
	__antithesis_instrumentation__.Notify(578004)
	if fn := p.ExecCfg().SchemaChangerTestingKnobs.RunBeforeResume; fn != nil {
		__antithesis_instrumentation__.Notify(578016)
		if err := fn(r.job.ID()); err != nil {
			__antithesis_instrumentation__.Notify(578017)
			return err
		} else {
			__antithesis_instrumentation__.Notify(578018)
		}
	} else {
		__antithesis_instrumentation__.Notify(578019)
	}
	__antithesis_instrumentation__.Notify(578005)
	execSchemaChange := func(descID descpb.ID, mutationID descpb.MutationID, droppedDatabaseID descpb.ID) error {
		__antithesis_instrumentation__.Notify(578020)
		sc := SchemaChanger{
			descID:               descID,
			mutationID:           mutationID,
			droppedDatabaseID:    droppedDatabaseID,
			sqlInstanceID:        p.ExecCfg().NodeID.SQLInstanceID(),
			db:                   p.ExecCfg().DB,
			leaseMgr:             p.ExecCfg().LeaseManager,
			testingKnobs:         p.ExecCfg().SchemaChangerTestingKnobs,
			distSQLPlanner:       p.DistSQLPlanner(),
			jobRegistry:          p.ExecCfg().JobRegistry,
			job:                  r.job,
			rangeDescriptorCache: p.ExecCfg().RangeDescriptorCache,
			clock:                p.ExecCfg().Clock,
			settings:             p.ExecCfg().Settings,
			execCfg:              p.ExecCfg(),
			ieFactory: func(ctx context.Context, sd *sessiondata.SessionData) sqlutil.InternalExecutor {
				__antithesis_instrumentation__.Notify(578023)
				return r.job.MakeSessionBoundInternalExecutor(ctx, sd)
			},
			metrics: p.ExecCfg().SchemaChangerMetrics,
		}
		__antithesis_instrumentation__.Notify(578021)
		opts := retry.Options{
			InitialBackoff: 20 * time.Millisecond,
			MaxBackoff:     20 * time.Second,
			Multiplier:     1.5,
		}

		var scErr error
		for r := retry.StartWithCtx(ctx, opts); r.Next(); {
			__antithesis_instrumentation__.Notify(578024)

			if err := p.ExecCfg().JobRegistry.CheckPausepoint("schemachanger.before.exec"); err != nil {
				__antithesis_instrumentation__.Notify(578026)
				return err
			} else {
				__antithesis_instrumentation__.Notify(578027)
			}
			__antithesis_instrumentation__.Notify(578025)
			scErr = sc.exec(ctx)
			switch {
			case scErr == nil:
				__antithesis_instrumentation__.Notify(578028)
				sc.metrics.Successes.Inc(1)
				return nil
			case errors.Is(scErr, catalog.ErrDescriptorNotFound):
				__antithesis_instrumentation__.Notify(578029)

				log.Infof(
					ctx,
					"descriptor %d not found for schema change processing mutation %d;"+
						"assuming it was dropped, and exiting",
					descID, mutationID,
				)
				return nil
			case !IsPermanentSchemaChangeError(scErr):
				__antithesis_instrumentation__.Notify(578030)

				log.Warningf(ctx, "error while running schema change, retrying: %v", scErr)
				sc.metrics.RetryErrors.Inc(1)
				if IsConstraintError(scErr) {
					__antithesis_instrumentation__.Notify(578034)
					telemetry.Inc(sc.metrics.ConstraintErrors)
				} else {
					__antithesis_instrumentation__.Notify(578035)
					telemetry.Inc(sc.metrics.UncategorizedErrors)
				}
			default:
				__antithesis_instrumentation__.Notify(578031)
				if ctx.Err() == nil {
					__antithesis_instrumentation__.Notify(578036)
					sc.metrics.PermanentErrors.Inc(1)
				} else {
					__antithesis_instrumentation__.Notify(578037)
				}
				__antithesis_instrumentation__.Notify(578032)
				if IsConstraintError(scErr) {
					__antithesis_instrumentation__.Notify(578038)
					telemetry.Inc(sc.metrics.ConstraintErrors)
				} else {
					__antithesis_instrumentation__.Notify(578039)
					telemetry.Inc(sc.metrics.UncategorizedErrors)
				}
				__antithesis_instrumentation__.Notify(578033)

				return scErr
			}

		}
		__antithesis_instrumentation__.Notify(578022)

		return scErr
	}
	__antithesis_instrumentation__.Notify(578006)

	for i := range details.DroppedTypes {
		__antithesis_instrumentation__.Notify(578040)
		ts := &typeSchemaChanger{
			typeID:  details.DroppedTypes[i],
			execCfg: p.ExecCfg(),
		}
		if err := ts.execWithRetry(ctx); err != nil {
			__antithesis_instrumentation__.Notify(578041)
			return err
		} else {
			__antithesis_instrumentation__.Notify(578042)
		}
	}
	__antithesis_instrumentation__.Notify(578007)

	for i := range details.DroppedTables {
		__antithesis_instrumentation__.Notify(578043)
		droppedTable := &details.DroppedTables[i]
		if err := execSchemaChange(droppedTable.ID, descpb.InvalidMutationID, details.DroppedDatabaseID); err != nil {
			__antithesis_instrumentation__.Notify(578044)
			return err
		} else {
			__antithesis_instrumentation__.Notify(578045)
		}
	}
	__antithesis_instrumentation__.Notify(578008)

	for _, id := range details.DroppedSchemas {
		__antithesis_instrumentation__.Notify(578046)
		if err := execSchemaChange(id, descpb.InvalidMutationID, descpb.InvalidID); err != nil {
			__antithesis_instrumentation__.Notify(578047)
			return err
		} else {
			__antithesis_instrumentation__.Notify(578048)
		}
	}
	__antithesis_instrumentation__.Notify(578009)

	if details.FormatVersion >= jobspb.DatabaseJobFormatVersion {
		__antithesis_instrumentation__.Notify(578049)
		if dbID := details.DroppedDatabaseID; dbID != descpb.InvalidID {
			__antithesis_instrumentation__.Notify(578050)
			if err := execSchemaChange(dbID, descpb.InvalidMutationID, descpb.InvalidID); err != nil {
				__antithesis_instrumentation__.Notify(578052)
				return err
			} else {
				__antithesis_instrumentation__.Notify(578053)
			}
			__antithesis_instrumentation__.Notify(578051)

			if len(details.DroppedTables) == 0 {
				__antithesis_instrumentation__.Notify(578054)
				zoneKeyPrefix := config.MakeZoneKeyPrefix(p.ExecCfg().Codec, dbID)
				if p.ExtendedEvalContext().Tracing.KVTracingEnabled() {
					__antithesis_instrumentation__.Notify(578056)
					log.VEventf(ctx, 2, "DelRange %s", zoneKeyPrefix)
				} else {
					__antithesis_instrumentation__.Notify(578057)
				}
				__antithesis_instrumentation__.Notify(578055)

				if _, err := p.ExecCfg().DB.DelRange(ctx, zoneKeyPrefix, zoneKeyPrefix.PrefixEnd(), false); err != nil {
					__antithesis_instrumentation__.Notify(578058)
					return err
				} else {
					__antithesis_instrumentation__.Notify(578059)
				}
			} else {
				__antithesis_instrumentation__.Notify(578060)
			}
		} else {
			__antithesis_instrumentation__.Notify(578061)
		}
	} else {
		__antithesis_instrumentation__.Notify(578062)
	}
	__antithesis_instrumentation__.Notify(578010)

	if len(details.DroppedTables) > 0 {
		__antithesis_instrumentation__.Notify(578063)
		dropTime := timeutil.Now().UnixNano()
		tablesToGC := make([]jobspb.SchemaChangeGCDetails_DroppedID, len(details.DroppedTables))
		for i, table := range details.DroppedTables {
			__antithesis_instrumentation__.Notify(578065)
			tablesToGC[i] = jobspb.SchemaChangeGCDetails_DroppedID{ID: table.ID, DropTime: dropTime}
		}
		__antithesis_instrumentation__.Notify(578064)
		multiTableGCDetails := jobspb.SchemaChangeGCDetails{
			Tables:   tablesToGC,
			ParentID: details.DroppedDatabaseID,
		}

		if err := startGCJob(
			ctx,
			p.ExecCfg().DB,
			p.ExecCfg().JobRegistry,
			r.job.Payload().UsernameProto.Decode(),
			r.job.Payload().Description,
			multiTableGCDetails,
		); err != nil {
			__antithesis_instrumentation__.Notify(578066)
			return err
		} else {
			__antithesis_instrumentation__.Notify(578067)
		}
	} else {
		__antithesis_instrumentation__.Notify(578068)
	}
	__antithesis_instrumentation__.Notify(578011)

	if details.DescID != descpb.InvalidID {
		__antithesis_instrumentation__.Notify(578069)
		return execSchemaChange(details.DescID, details.TableMutationID, details.DroppedDatabaseID)
	} else {
		__antithesis_instrumentation__.Notify(578070)
	}
	__antithesis_instrumentation__.Notify(578012)
	return nil
}

func (r schemaChangeResumer) OnFailOrCancel(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(578071)
	p := execCtx.(JobExecContext)
	details := r.job.Details().(jobspb.SchemaChangeDetails)

	if fn := p.ExecCfg().SchemaChangerTestingKnobs.RunBeforeOnFailOrCancel; fn != nil {
		__antithesis_instrumentation__.Notify(578078)
		if err := fn(r.job.ID()); err != nil {
			__antithesis_instrumentation__.Notify(578079)
			return err
		} else {
			__antithesis_instrumentation__.Notify(578080)
		}
	} else {
		__antithesis_instrumentation__.Notify(578081)
	}
	__antithesis_instrumentation__.Notify(578072)

	if details.DescID == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(578082)
		return errors.Newf("schema change jobs on databases and schemas cannot be reverted")
	} else {
		__antithesis_instrumentation__.Notify(578083)
	}
	__antithesis_instrumentation__.Notify(578073)
	sc := SchemaChanger{
		descID:               details.DescID,
		mutationID:           details.TableMutationID,
		sqlInstanceID:        p.ExecCfg().NodeID.SQLInstanceID(),
		db:                   p.ExecCfg().DB,
		leaseMgr:             p.ExecCfg().LeaseManager,
		testingKnobs:         p.ExecCfg().SchemaChangerTestingKnobs,
		distSQLPlanner:       p.DistSQLPlanner(),
		jobRegistry:          p.ExecCfg().JobRegistry,
		job:                  r.job,
		rangeDescriptorCache: p.ExecCfg().RangeDescriptorCache,
		clock:                p.ExecCfg().Clock,
		settings:             p.ExecCfg().Settings,
		execCfg:              p.ExecCfg(),
		ieFactory: func(ctx context.Context, sd *sessiondata.SessionData) sqlutil.InternalExecutor {
			__antithesis_instrumentation__.Notify(578084)
			return r.job.MakeSessionBoundInternalExecutor(ctx, sd)
		},
	}
	__antithesis_instrumentation__.Notify(578074)

	if r.job.Payload().FinalResumeError == nil {
		__antithesis_instrumentation__.Notify(578085)
		return errors.AssertionFailedf("job failed but had no recorded error")
	} else {
		__antithesis_instrumentation__.Notify(578086)
	}
	__antithesis_instrumentation__.Notify(578075)
	scErr := errors.DecodeError(ctx, *r.job.Payload().FinalResumeError)

	if rollbackErr := sc.handlePermanentSchemaChangeError(ctx, scErr, p.ExtendedEvalContext()); rollbackErr != nil {
		__antithesis_instrumentation__.Notify(578087)
		switch {
		case errors.Is(rollbackErr, catalog.ErrDescriptorNotFound):
			__antithesis_instrumentation__.Notify(578088)

			log.Infof(
				ctx,
				"descriptor %d not found for rollback of schema change processing mutation %d;"+
					"assuming it was dropped, and exiting",
				details.DescID, details.TableMutationID,
			)
		case ctx.Err() != nil:
			__antithesis_instrumentation__.Notify(578089)

			return rollbackErr
		case !IsPermanentSchemaChangeError(rollbackErr):
			__antithesis_instrumentation__.Notify(578090)

			return jobs.MarkAsRetryJobError(rollbackErr)
		default:
			__antithesis_instrumentation__.Notify(578091)

			return rollbackErr
		}
	} else {
		__antithesis_instrumentation__.Notify(578092)
	}
	__antithesis_instrumentation__.Notify(578076)

	if fn := sc.testingKnobs.RunAfterOnFailOrCancel; fn != nil {
		__antithesis_instrumentation__.Notify(578093)
		if err := fn(r.job.ID()); err != nil {
			__antithesis_instrumentation__.Notify(578094)
			return err
		} else {
			__antithesis_instrumentation__.Notify(578095)
		}
	} else {
		__antithesis_instrumentation__.Notify(578096)
	}
	__antithesis_instrumentation__.Notify(578077)
	return nil
}

func init() {
	createResumerFn := func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
		return &schemaChangeResumer{job: job}
	}
	jobs.RegisterConstructor(jobspb.TypeSchemaChange, createResumerFn)
}

func (sc *SchemaChanger) queueCleanupJob(
	ctx context.Context, scDesc *tabledesc.Mutable, txn *kv.Txn,
) (jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(578097)

	mutationID := scDesc.ClusterVersion().NextMutationID
	span := scDesc.PrimaryIndexSpan(sc.execCfg.Codec)
	var spanList []jobspb.ResumeSpanList
	for j := len(scDesc.ClusterVersion().Mutations); j < len(scDesc.Mutations); j++ {
		__antithesis_instrumentation__.Notify(578100)
		spanList = append(spanList,
			jobspb.ResumeSpanList{
				ResumeSpans: roachpb.Spans{span},
			},
		)
	}
	__antithesis_instrumentation__.Notify(578098)

	var jobID jobspb.JobID
	if len(spanList) > 0 {
		__antithesis_instrumentation__.Notify(578101)
		jobRecord := jobs.Record{
			Description:   fmt.Sprintf("CLEANUP JOB for '%s'", sc.job.Payload().Description),
			Username:      sc.job.Payload().UsernameProto.Decode(),
			DescriptorIDs: descpb.IDs{scDesc.GetID()},
			Details: jobspb.SchemaChangeDetails{
				DescID:          sc.descID,
				TableMutationID: mutationID,
				ResumeSpanList:  spanList,

				FormatVersion: jobspb.DatabaseJobFormatVersion,
			},
			Progress:      jobspb.SchemaChangeProgress{},
			NonCancelable: true,
		}
		jobID = sc.jobRegistry.MakeJobID()
		if _, err := sc.jobRegistry.CreateJobWithTxn(ctx, jobRecord, jobID, txn); err != nil {
			__antithesis_instrumentation__.Notify(578103)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(578104)
		}
		__antithesis_instrumentation__.Notify(578102)
		log.Infof(ctx, "created job %d to drop previous columns and indexes", jobID)
		scDesc.MutationJobs = append(scDesc.MutationJobs, descpb.TableDescriptor_MutationJob{
			MutationID: mutationID,
			JobID:      jobID,
		})
	} else {
		__antithesis_instrumentation__.Notify(578105)
	}
	__antithesis_instrumentation__.Notify(578099)
	return jobID, nil
}

func (sc *SchemaChanger) applyZoneConfigChangeForMutation(
	ctx context.Context,
	txn *kv.Txn,
	dbDesc catalog.DatabaseDescriptor,
	tableDesc *tabledesc.Mutable,
	mutation catalog.Mutation,
	isDone bool,
	descsCol *descs.Collection,
) error {
	__antithesis_instrumentation__.Notify(578106)
	if pkSwap := mutation.AsPrimaryKeySwap(); pkSwap != nil {
		__antithesis_instrumentation__.Notify(578108)
		if pkSwap.HasLocalityConfig() {
			__antithesis_instrumentation__.Notify(578110)

			opts := make([]applyZoneConfigForMultiRegionTableOption, 0, 3)
			lcSwap := pkSwap.LocalityConfigSwap()

			if mutation.Adding() {
				__antithesis_instrumentation__.Notify(578113)

				if isDone {
					__antithesis_instrumentation__.Notify(578115)

					if lcSwap.OldLocalityConfig.GetRegionalByRow() != nil {
						__antithesis_instrumentation__.Notify(578117)
						oldIndexIDs := make([]descpb.IndexID, 0, pkSwap.NumOldIndexes())
						_ = pkSwap.ForEachOldIndexIDs(func(id descpb.IndexID) error {
							__antithesis_instrumentation__.Notify(578119)
							oldIndexIDs = append(oldIndexIDs, id)
							return nil
						})
						__antithesis_instrumentation__.Notify(578118)
						opts = append(opts, dropZoneConfigsForMultiRegionIndexes(oldIndexIDs...))
					} else {
						__antithesis_instrumentation__.Notify(578120)
					}
					__antithesis_instrumentation__.Notify(578116)

					opts = append(
						opts,
						applyZoneConfigForMultiRegionTableOptionTableNewConfig(
							lcSwap.NewLocalityConfig,
						),
					)
				} else {
					__antithesis_instrumentation__.Notify(578121)
				}
				__antithesis_instrumentation__.Notify(578114)
				switch lcSwap.NewLocalityConfig.Locality.(type) {
				case *catpb.LocalityConfig_Global_,
					*catpb.LocalityConfig_RegionalByTable_:
					__antithesis_instrumentation__.Notify(578122)
				case *catpb.LocalityConfig_RegionalByRow_:
					__antithesis_instrumentation__.Notify(578123)

					newIndexIDs := make([]descpb.IndexID, 0, pkSwap.NumNewIndexes())
					_ = pkSwap.ForEachNewIndexIDs(func(id descpb.IndexID) error {
						__antithesis_instrumentation__.Notify(578126)
						newIndexIDs = append(newIndexIDs, id)
						return nil
					})
					__antithesis_instrumentation__.Notify(578124)
					opts = append(opts, applyZoneConfigForMultiRegionTableOptionNewIndexes(newIndexIDs...))
				default:
					__antithesis_instrumentation__.Notify(578125)
					return errors.AssertionFailedf(
						"unknown locality on PK swap: %T",
						lcSwap.NewLocalityConfig.Locality,
					)
				}
			} else {
				__antithesis_instrumentation__.Notify(578127)

				opts = append(
					opts,
					applyZoneConfigForMultiRegionTableOptionTableNewConfig(
						lcSwap.OldLocalityConfig,
					),
				)
			}
			__antithesis_instrumentation__.Notify(578111)

			regionConfig, err := SynthesizeRegionConfig(ctx, txn, dbDesc.GetID(), descsCol)
			if err != nil {
				__antithesis_instrumentation__.Notify(578128)
				return err
			} else {
				__antithesis_instrumentation__.Notify(578129)
			}
			__antithesis_instrumentation__.Notify(578112)
			if err := ApplyZoneConfigForMultiRegionTable(
				ctx,
				txn,
				sc.execCfg,
				regionConfig,
				tableDesc,
				opts...,
			); err != nil {
				__antithesis_instrumentation__.Notify(578130)
				return err
			} else {
				__antithesis_instrumentation__.Notify(578131)
			}
		} else {
			__antithesis_instrumentation__.Notify(578132)
		}
		__antithesis_instrumentation__.Notify(578109)

		return maybeUpdateZoneConfigsForPKChange(
			ctx, txn, sc.execCfg, tableDesc, pkSwap.PrimaryKeySwapDesc(),
		)
	} else {
		__antithesis_instrumentation__.Notify(578133)
	}
	__antithesis_instrumentation__.Notify(578107)
	return nil
}

func DeleteTableDescAndZoneConfig(
	ctx context.Context,
	db *kv.DB,
	settings *cluster.Settings,
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(578134)
	log.Infof(ctx, "removing table descriptor and zone config for table %d", tableDesc.GetID())
	return db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(578135)
		if !settings.Version.IsActive(
			ctx, clusterversion.DisableSystemConfigGossipTrigger,
		) {
			__antithesis_instrumentation__.Notify(578138)
			if err := txn.DeprecatedSetSystemConfigTrigger(codec.ForSystemTenant()); err != nil {
				__antithesis_instrumentation__.Notify(578139)
				return err
			} else {
				__antithesis_instrumentation__.Notify(578140)
			}
		} else {
			__antithesis_instrumentation__.Notify(578141)
		}
		__antithesis_instrumentation__.Notify(578136)
		b := &kv.Batch{}

		descKey := catalogkeys.MakeDescMetadataKey(codec, tableDesc.GetID())
		b.Del(descKey)

		if codec.ForSystemTenant() {
			__antithesis_instrumentation__.Notify(578142)
			zoneKeyPrefix := config.MakeZoneKeyPrefix(codec, tableDesc.GetID())
			b.DelRange(zoneKeyPrefix, zoneKeyPrefix.PrefixEnd(), false)
		} else {
			__antithesis_instrumentation__.Notify(578143)
		}
		__antithesis_instrumentation__.Notify(578137)
		return txn.Run(ctx, b)
	})
}

func (sc *SchemaChanger) getDependentMutationsJobs(
	ctx context.Context, tableDesc *tabledesc.Mutable, mutations []catalog.Mutation,
) ([]jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(578144)
	dependentJobs := make([]jobspb.JobID, 0, len(tableDesc.MutationJobs))
	for _, m := range mutations {
		__antithesis_instrumentation__.Notify(578146)

		discarded, dependentID := isCurrentMutationDiscarded(tableDesc, m, 0)
		if discarded {
			__antithesis_instrumentation__.Notify(578147)
			jobID, err := getJobIDForMutationWithDescriptor(ctx, tableDesc, dependentID)
			if err != nil {
				__antithesis_instrumentation__.Notify(578149)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(578150)
			}
			__antithesis_instrumentation__.Notify(578148)
			dependentJobs = append(dependentJobs, jobID)
		} else {
			__antithesis_instrumentation__.Notify(578151)
		}
	}
	__antithesis_instrumentation__.Notify(578145)
	return dependentJobs, nil
}

func (sc *SchemaChanger) shouldSplitAndScatter(
	tableDesc *tabledesc.Mutable, m catalog.Mutation, idx catalog.Index,
) bool {
	__antithesis_instrumentation__.Notify(578152)
	if idx == nil {
		__antithesis_instrumentation__.Notify(578155)
		return false
	} else {
		__antithesis_instrumentation__.Notify(578156)
	}
	__antithesis_instrumentation__.Notify(578153)

	if m.Adding() && func() bool {
		__antithesis_instrumentation__.Notify(578157)
		return idx.IsSharded() == true
	}() == true {
		__antithesis_instrumentation__.Notify(578158)
		if sc.mvccCompliantAddIndex {
			__antithesis_instrumentation__.Notify(578160)
			return m.Backfilling() || func() bool {
				__antithesis_instrumentation__.Notify(578161)
				return (idx.IsTemporaryIndexForBackfill() && func() bool {
					__antithesis_instrumentation__.Notify(578162)
					return m.DeleteOnly() == true
				}() == true) == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(578163)
		}
		__antithesis_instrumentation__.Notify(578159)
		return m.DeleteOnly()
	} else {
		__antithesis_instrumentation__.Notify(578164)
	}
	__antithesis_instrumentation__.Notify(578154)
	return false

}

func (sc *SchemaChanger) preSplitHashShardedIndexRanges(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(578165)
	if err := sc.txn(ctx, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(578167)
		hour := hlc.Timestamp{WallTime: timeutil.Now().Add(time.Hour).UnixNano()}
		tableDesc, err := descsCol.GetMutableTableByID(
			ctx, txn, sc.descID,
			tree.ObjectLookupFlags{
				CommonLookupFlags: tree.CommonLookupFlags{
					IncludeOffline: true,
					IncludeDropped: true,
				},
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(578172)
			return err
		} else {
			__antithesis_instrumentation__.Notify(578173)
		}
		__antithesis_instrumentation__.Notify(578168)

		if fn := sc.testingKnobs.RunBeforeHashShardedIndexRangePreSplit; fn != nil {
			__antithesis_instrumentation__.Notify(578174)
			if err := fn(tableDesc, sc.db, sc.execCfg.Codec); err != nil {
				__antithesis_instrumentation__.Notify(578175)
				return err
			} else {
				__antithesis_instrumentation__.Notify(578176)
			}
		} else {
			__antithesis_instrumentation__.Notify(578177)
		}
		__antithesis_instrumentation__.Notify(578169)

		for _, m := range tableDesc.AllMutations() {
			__antithesis_instrumentation__.Notify(578178)
			if m.MutationID() != sc.mutationID {
				__antithesis_instrumentation__.Notify(578180)

				break
			} else {
				__antithesis_instrumentation__.Notify(578181)
			}
			__antithesis_instrumentation__.Notify(578179)

			if idx := m.AsIndex(); sc.shouldSplitAndScatter(tableDesc, m, idx) {
				__antithesis_instrumentation__.Notify(578182)

				var partitionKeyPrefixes []roachpb.Key
				partitioning := idx.GetPartitioning()
				if err := partitioning.ForEachList(
					func(name string, values [][]byte, subPartitioning catalog.Partitioning) error {
						__antithesis_instrumentation__.Notify(578184)
						for _, tupleBytes := range values {
							__antithesis_instrumentation__.Notify(578186)
							_, key, err := rowenc.DecodePartitionTuple(
								&tree.DatumAlloc{},
								sc.execCfg.Codec,
								tableDesc,
								idx,
								partitioning,
								tupleBytes,
								tree.Datums{},
							)
							if err != nil {
								__antithesis_instrumentation__.Notify(578188)
								return err
							} else {
								__antithesis_instrumentation__.Notify(578189)
							}
							__antithesis_instrumentation__.Notify(578187)
							partitionKeyPrefixes = append(partitionKeyPrefixes, key)
						}
						__antithesis_instrumentation__.Notify(578185)
						return nil
					},
				); err != nil {
					__antithesis_instrumentation__.Notify(578190)
					return err
				} else {
					__antithesis_instrumentation__.Notify(578191)
				}
				__antithesis_instrumentation__.Notify(578183)

				splitAtShards := calculateSplitAtShards(maxHashShardedIndexRangePreSplit.Get(&sc.settings.SV), idx.GetSharded().ShardBuckets)
				if len(partitionKeyPrefixes) == 0 {
					__antithesis_instrumentation__.Notify(578192)

					for _, shard := range splitAtShards {
						__antithesis_instrumentation__.Notify(578193)
						keyPrefix := sc.execCfg.Codec.IndexPrefix(uint32(tableDesc.GetID()), uint32(idx.GetID()))
						splitKey := encoding.EncodeVarintAscending(keyPrefix, shard)
						if err := splitAndScatter(ctx, sc.db, splitKey, hour); err != nil {
							__antithesis_instrumentation__.Notify(578194)
							return err
						} else {
							__antithesis_instrumentation__.Notify(578195)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(578196)

					for _, partPrefix := range partitionKeyPrefixes {
						__antithesis_instrumentation__.Notify(578197)
						for _, shard := range splitAtShards {
							__antithesis_instrumentation__.Notify(578198)
							splitKey := encoding.EncodeVarintAscending(partPrefix, shard)
							if err := splitAndScatter(ctx, sc.db, splitKey, hour); err != nil {
								__antithesis_instrumentation__.Notify(578199)
								return err
							} else {
								__antithesis_instrumentation__.Notify(578200)
							}
						}
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(578201)
			}
		}
		__antithesis_instrumentation__.Notify(578170)

		if fn := sc.testingKnobs.RunAfterHashShardedIndexRangePreSplit; fn != nil {
			__antithesis_instrumentation__.Notify(578202)
			if err := fn(tableDesc, sc.db, sc.execCfg.Codec); err != nil {
				__antithesis_instrumentation__.Notify(578203)
				return err
			} else {
				__antithesis_instrumentation__.Notify(578204)
			}
		} else {
			__antithesis_instrumentation__.Notify(578205)
		}
		__antithesis_instrumentation__.Notify(578171)

		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(578206)
		return err
	} else {
		__antithesis_instrumentation__.Notify(578207)
	}
	__antithesis_instrumentation__.Notify(578166)

	return nil
}

func splitAndScatter(
	ctx context.Context, db *kv.DB, key roachpb.Key, expirationTime hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(578208)
	if err := db.AdminSplit(ctx, key, expirationTime); err != nil {
		__antithesis_instrumentation__.Notify(578210)
		return err
	} else {
		__antithesis_instrumentation__.Notify(578211)
	}
	__antithesis_instrumentation__.Notify(578209)
	_, err := db.AdminScatter(ctx, key, 0)
	return err
}

func calculateSplitAtShards(maxSplit int64, shardBucketCount int32) []int64 {
	__antithesis_instrumentation__.Notify(578212)
	splitCount := int(math.Min(float64(maxSplit), float64(shardBucketCount)))
	step := float64(shardBucketCount) / float64(splitCount)
	splitAtShards := make([]int64, splitCount)
	for i := 0; i < splitCount; i++ {
		__antithesis_instrumentation__.Notify(578214)
		splitAtShards[i] = int64(math.Floor(float64(i) * step))
	}
	__antithesis_instrumentation__.Notify(578213)
	return splitAtShards
}

func isCurrentMutationDiscarded(
	tableDesc catalog.TableDescriptor, currentMutation catalog.Mutation, nextMutationIdx int,
) (bool, descpb.MutationID) {
	__antithesis_instrumentation__.Notify(578215)
	if nextMutationIdx+1 > len(tableDesc.AllMutations()) {
		__antithesis_instrumentation__.Notify(578220)
		return false, descpb.InvalidMutationID
	} else {
		__antithesis_instrumentation__.Notify(578221)
	}
	__antithesis_instrumentation__.Notify(578216)

	if currentMutation.Dropped() {
		__antithesis_instrumentation__.Notify(578222)
		return false, descpb.InvalidMutationID
	} else {
		__antithesis_instrumentation__.Notify(578223)
	}
	__antithesis_instrumentation__.Notify(578217)

	colToCheck := make([]descpb.ColumnID, 0, 1)

	if constraint := currentMutation.AsConstraint(); constraint != nil {
		__antithesis_instrumentation__.Notify(578224)
		if constraint.IsNotNull() {
			__antithesis_instrumentation__.Notify(578225)
			colToCheck = append(colToCheck, constraint.NotNullColumnID())
		} else {
			__antithesis_instrumentation__.Notify(578226)
			if constraint.IsCheck() {
				__antithesis_instrumentation__.Notify(578227)
				colToCheck = constraint.Check().ColumnIDs
			} else {
				__antithesis_instrumentation__.Notify(578228)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(578229)
	}
	__antithesis_instrumentation__.Notify(578218)

	for _, m := range tableDesc.AllMutations()[nextMutationIdx:] {
		__antithesis_instrumentation__.Notify(578230)
		col := m.AsColumn()
		if col != nil && func() bool {
			__antithesis_instrumentation__.Notify(578231)
			return col.Dropped() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(578232)
			return !col.IsRollback() == true
		}() == true {
			__antithesis_instrumentation__.Notify(578233)

			for _, id := range colToCheck {
				__antithesis_instrumentation__.Notify(578234)
				if col.GetID() == id {
					__antithesis_instrumentation__.Notify(578235)
					return true, m.MutationID()
				} else {
					__antithesis_instrumentation__.Notify(578236)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(578237)
		}
	}
	__antithesis_instrumentation__.Notify(578219)

	return false, descpb.InvalidMutationID
}
