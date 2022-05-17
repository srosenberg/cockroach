package spanconfigtestcluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobsprotectedts"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigreconciler"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigtestutils"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/distsql"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/stretchr/testify/require"
)

type Tenant struct {
	serverutils.TestTenantInterface

	t          *testing.T
	db         *sqlutils.SQLRunner
	reconciler *spanconfigreconciler.Reconciler
	recorder   *spanconfigtestutils.KVAccessorRecorder
	cleanup    func()

	mu struct {
		syncutil.Mutex
		lastCheckpoint, tsAfterLastSQLChange hlc.Timestamp
	}
}

func (s *Tenant) ExecCfg() sql.ExecutorConfig {
	__antithesis_instrumentation__.Notify(241762)
	return s.ExecutorConfig().(sql.ExecutorConfig)
}

func (s *Tenant) ProtectedTimestampProvider() protectedts.Provider {
	__antithesis_instrumentation__.Notify(241763)
	return s.DistSQLServer().(*distsql.ServerImpl).ServerConfig.ProtectedTimestampProvider
}

func (s *Tenant) JobsRegistry() *jobs.Registry {
	__antithesis_instrumentation__.Notify(241764)
	return s.JobRegistry().(*jobs.Registry)
}

func (s *Tenant) updateTimestampAfterLastSQLChange() {
	__antithesis_instrumentation__.Notify(241765)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.tsAfterLastSQLChange = s.Clock().Now()
}

func (s *Tenant) Exec(query string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(241766)
	s.db.Exec(s.t, query, args...)
	s.updateTimestampAfterLastSQLChange()
}

func (s *Tenant) ExecWithErr(query string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(241767)
	_, err := s.db.DB.ExecContext(context.Background(), query, args...)
	s.updateTimestampAfterLastSQLChange()
	return err
}

func (s *Tenant) TimestampAfterLastSQLChange() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(241768)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.tsAfterLastSQLChange
}

func (s *Tenant) RecordCheckpoint() {
	__antithesis_instrumentation__.Notify(241769)
	ts := s.Reconciler().Checkpoint()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.lastCheckpoint = ts
}

func (s *Tenant) LastCheckpoint() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(241770)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.lastCheckpoint
}

func (s *Tenant) Query(query string, args ...interface{}) *gosql.Rows {
	__antithesis_instrumentation__.Notify(241771)
	return s.db.Query(s.t, query, args...)
}

func (s *Tenant) QueryRow(query string, args ...interface{}) *sqlutils.Row {
	__antithesis_instrumentation__.Notify(241772)
	return s.db.QueryRow(s.t, query, args...)
}

func (s *Tenant) Reconciler() spanconfig.Reconciler {
	__antithesis_instrumentation__.Notify(241773)
	return s.reconciler
}

func (s *Tenant) KVAccessorRecorder() *spanconfigtestutils.KVAccessorRecorder {
	__antithesis_instrumentation__.Notify(241774)
	return s.recorder
}

func (s *Tenant) WithMutableTableDescriptor(
	ctx context.Context, dbName string, tbName string, f func(*tabledesc.Mutable),
) {
	__antithesis_instrumentation__.Notify(241775)
	execCfg := s.ExecutorConfig().(sql.ExecutorConfig)
	require.NoError(s.t, sql.DescsTxn(ctx, &execCfg, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(241776)
		_, desc, err := descsCol.GetMutableTableByName(
			ctx,
			txn,
			tree.NewTableNameWithSchema(tree.Name(dbName), "public", tree.Name(tbName)),
			tree.ObjectLookupFlags{
				CommonLookupFlags: tree.CommonLookupFlags{
					Required:       true,
					IncludeOffline: true,
				},
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(241778)
			return err
		} else {
			__antithesis_instrumentation__.Notify(241779)
		}
		__antithesis_instrumentation__.Notify(241777)
		f(desc)
		return descsCol.WriteDesc(ctx, false, desc, txn)
	}))
}

var descLookupFlags = tree.CommonLookupFlags{
	IncludeDropped: true,
	IncludeOffline: true,
	AvoidLeased:    true,
}

func (s *Tenant) LookupTableDescriptorByID(
	ctx context.Context, id descpb.ID,
) (desc catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(241780)
	execCfg := s.ExecutorConfig().(sql.ExecutorConfig)
	require.NoError(s.t, sql.DescsTxn(ctx, &execCfg, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(241782)
		var err error
		desc, err = descsCol.GetImmutableTableByID(ctx, txn, id,
			tree.ObjectLookupFlags{
				CommonLookupFlags: descLookupFlags,
			},
		)
		return err
	}))
	__antithesis_instrumentation__.Notify(241781)
	return desc
}

func (s *Tenant) LookupTableByName(
	ctx context.Context, dbName string, tbName string,
) (desc catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(241783)
	execCfg := s.ExecutorConfig().(sql.ExecutorConfig)
	require.NoError(s.t, sql.DescsTxn(ctx, &execCfg, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(241785)
		var err error
		_, desc, err = descsCol.GetImmutableTableByName(ctx, txn,
			tree.NewTableNameWithSchema(tree.Name(dbName), "public", tree.Name(tbName)),
			tree.ObjectLookupFlags{
				CommonLookupFlags: descLookupFlags,
			},
		)
		return err
	}))
	__antithesis_instrumentation__.Notify(241784)
	return desc
}

func (s *Tenant) LookupDatabaseByName(
	ctx context.Context, dbName string,
) (desc catalog.DatabaseDescriptor) {
	__antithesis_instrumentation__.Notify(241786)
	execCfg := s.ExecutorConfig().(sql.ExecutorConfig)
	require.NoError(s.t, sql.DescsTxn(ctx, &execCfg,
		func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
			__antithesis_instrumentation__.Notify(241788)
			var err error
			desc, err = descsCol.GetImmutableDatabaseByName(ctx, txn, dbName,
				tree.DatabaseLookupFlags{
					Required:       true,
					IncludeOffline: true,
					AvoidLeased:    true,
				},
			)
			return err
		}))
	__antithesis_instrumentation__.Notify(241787)
	return desc
}

func (s *Tenant) MakeProtectedTimestampRecordAndProtect(
	ctx context.Context, recordID string, protectTS int, target *ptpb.Target,
) {
	__antithesis_instrumentation__.Notify(241789)
	jobID := s.JobsRegistry().MakeJobID()
	require.NoError(s.t, s.ExecCfg().DB.Txn(ctx,
		func(ctx context.Context, txn *kv.Txn) (err error) {
			__antithesis_instrumentation__.Notify(241791)
			require.Len(s.t, recordID, 1,
				"datadriven test only supports single character record IDs")
			recID, err := uuid.FromBytes([]byte(strings.Repeat(recordID, 16)))
			require.NoError(s.t, err)
			rec := jobsprotectedts.MakeRecord(recID, int64(jobID),
				hlc.Timestamp{WallTime: int64(protectTS)}, nil,
				jobsprotectedts.Jobs, target)
			return s.ProtectedTimestampProvider().Protect(ctx, txn, rec)
		}))
	__antithesis_instrumentation__.Notify(241790)
	s.updateTimestampAfterLastSQLChange()
}

func (s *Tenant) ReleaseProtectedTimestampRecord(ctx context.Context, recordID string) {
	__antithesis_instrumentation__.Notify(241792)
	require.NoError(s.t, s.ExecCfg().DB.Txn(ctx,
		func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(241794)
			require.Len(s.t, recordID, 1,
				"datadriven test only supports single character record IDs")
			recID, err := uuid.FromBytes([]byte(strings.Repeat(recordID, 16)))
			require.NoError(s.t, err)
			return s.ProtectedTimestampProvider().Release(ctx, txn, recID)
		}))
	__antithesis_instrumentation__.Notify(241793)
	s.updateTimestampAfterLastSQLChange()
}
