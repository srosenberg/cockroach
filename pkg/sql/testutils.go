package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

func CreateTestTableDescriptor(
	ctx context.Context, parentID, id descpb.ID, schema string, privileges *catpb.PrivilegeDescriptor,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(628380)
	st := cluster.MakeTestingClusterSettings()
	stmt, err := parser.ParseOne(schema)
	if err != nil {
		__antithesis_instrumentation__.Notify(628382)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(628383)
	}
	__antithesis_instrumentation__.Notify(628381)
	semaCtx := tree.MakeSemaContext()
	evalCtx := tree.MakeTestingEvalContext(st)
	switch n := stmt.AST.(type) {
	case *tree.CreateTable:
		__antithesis_instrumentation__.Notify(628384)
		db := dbdesc.NewInitial(parentID, "test", security.RootUserName())
		desc, err := NewTableDesc(
			ctx,
			nil,
			nil,
			st,
			n,
			db,
			schemadesc.GetPublicSchema(),
			id,
			nil,
			hlc.Timestamp{},
			privileges,
			nil,
			&semaCtx,
			&evalCtx,
			&sessiondata.SessionData{
				LocalOnlySessionData: sessiondatapb.LocalOnlySessionData{
					EnableUniqueWithoutIndexConstraints: true,
				},
			},
			tree.PersistencePermanent,
		)
		return desc, err
	case *tree.CreateSequence:
		__antithesis_instrumentation__.Notify(628385)
		desc, err := NewSequenceTableDesc(
			ctx,
			nil,
			st,
			n.Name.Table(),
			n.Options,
			parentID, keys.PublicSchemaID, id,
			hlc.Timestamp{},
			privileges,
			tree.PersistencePermanent,
			false,
		)
		return desc, err
	default:
		__antithesis_instrumentation__.Notify(628386)
		return nil, errors.Errorf("unexpected AST %T", stmt.AST)
	}
}

type StmtBufReader struct {
	buf *StmtBuf
}

func MakeStmtBufReader(buf *StmtBuf) StmtBufReader {
	__antithesis_instrumentation__.Notify(628387)
	return StmtBufReader{buf: buf}
}

func (r StmtBufReader) CurCmd() (Command, error) {
	__antithesis_instrumentation__.Notify(628388)
	cmd, _, err := r.buf.CurCmd()
	return cmd, err
}

func (r *StmtBufReader) AdvanceOne() {
	__antithesis_instrumentation__.Notify(628389)
	r.buf.AdvanceOne()
}

func (dsp *DistSQLPlanner) Exec(
	ctx context.Context, localPlanner interface{}, sql string, distribute bool,
) error {
	__antithesis_instrumentation__.Notify(628390)
	stmt, err := parser.ParseOne(sql)
	if err != nil {
		__antithesis_instrumentation__.Notify(628395)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628396)
	}
	__antithesis_instrumentation__.Notify(628391)
	p := localPlanner.(*planner)
	p.stmt = makeStatement(stmt, ClusterWideID{})
	if err := p.makeOptimizerPlan(ctx); err != nil {
		__antithesis_instrumentation__.Notify(628397)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628398)
	}
	__antithesis_instrumentation__.Notify(628392)
	rw := NewCallbackResultWriter(func(ctx context.Context, row tree.Datums) error {
		__antithesis_instrumentation__.Notify(628399)
		return nil
	})
	__antithesis_instrumentation__.Notify(628393)
	execCfg := p.ExecCfg()
	recv := MakeDistSQLReceiver(
		ctx,
		rw,
		stmt.AST.StatementReturnType(),
		execCfg.RangeDescriptorCache,
		p.txn,
		execCfg.Clock,
		p.ExtendedEvalContext().Tracing,
		execCfg.ContentionRegistry,
		nil,
	)
	defer recv.Release()

	distributionType := DistributionType(DistributionTypeNone)
	if distribute {
		__antithesis_instrumentation__.Notify(628400)
		distributionType = DistributionTypeSystemTenantOnly
	} else {
		__antithesis_instrumentation__.Notify(628401)
	}
	__antithesis_instrumentation__.Notify(628394)
	evalCtx := p.ExtendedEvalContext()
	planCtx := execCfg.DistSQLPlanner.NewPlanningCtx(ctx, evalCtx, p, p.txn,
		distributionType)
	planCtx.stmtType = recv.stmtType

	dsp.PlanAndRun(ctx, evalCtx, planCtx, p.txn, p.curPlan.main, recv)()
	return rw.Err()
}
