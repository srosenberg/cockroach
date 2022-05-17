package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/opt"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/optbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type opaqueMetadata struct {
	info    string
	plan    planNode
	columns colinfo.ResultColumns
}

var _ opt.OpaqueMetadata = &opaqueMetadata{}

func (o *opaqueMetadata) ImplementsOpaqueMetadata() { __antithesis_instrumentation__.Notify(501863) }
func (o *opaqueMetadata) String() string {
	__antithesis_instrumentation__.Notify(501864)
	return o.info
}
func (o *opaqueMetadata) Columns() colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(501865)
	return o.columns
}

func buildOpaque(
	ctx context.Context, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, stmt tree.Statement,
) (opt.OpaqueMetadata, error) {
	__antithesis_instrumentation__.Notify(501866)
	p := evalCtx.Planner.(*planner)

	scalarProps := &semaCtx.Properties
	defer scalarProps.Restore(*scalarProps)
	scalarProps.Require(stmt.StatementTag(), tree.RejectSubqueries)

	var plan planNode
	if tree.CanModifySchema(stmt) {
		__antithesis_instrumentation__.Notify(501870)
		if err := p.checkNoConflictingCursors(stmt); err != nil {
			__antithesis_instrumentation__.Notify(501872)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(501873)
		}
		__antithesis_instrumentation__.Notify(501871)

		if evalCtx.Settings.Version.IsActive(ctx, clusterversion.EnableDeclarativeSchemaChanger) {
			__antithesis_instrumentation__.Notify(501874)
			scPlan, usePlan, err := p.SchemaChange(ctx, stmt)
			if err != nil {
				__antithesis_instrumentation__.Notify(501876)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(501877)
			}
			__antithesis_instrumentation__.Notify(501875)
			if usePlan {
				__antithesis_instrumentation__.Notify(501878)
				plan = scPlan
			} else {
				__antithesis_instrumentation__.Notify(501879)
			}
		} else {
			__antithesis_instrumentation__.Notify(501880)
		}
	} else {
		__antithesis_instrumentation__.Notify(501881)
	}
	__antithesis_instrumentation__.Notify(501867)
	if plan == nil {
		__antithesis_instrumentation__.Notify(501882)
		var err error
		plan, err = planOpaque(ctx, p, stmt)
		if err != nil {
			__antithesis_instrumentation__.Notify(501883)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(501884)
		}
	} else {
		__antithesis_instrumentation__.Notify(501885)
	}
	__antithesis_instrumentation__.Notify(501868)
	if plan == nil {
		__antithesis_instrumentation__.Notify(501886)
		return nil, errors.AssertionFailedf("planNode cannot be nil for %T", stmt)
	} else {
		__antithesis_instrumentation__.Notify(501887)
	}
	__antithesis_instrumentation__.Notify(501869)
	res := &opaqueMetadata{
		info:    stmt.StatementTag(),
		plan:    plan,
		columns: planColumns(plan),
	}
	return res, nil
}

func planOpaque(ctx context.Context, p *planner, stmt tree.Statement) (planNode, error) {
	__antithesis_instrumentation__.Notify(501888)
	switch n := stmt.(type) {
	case *tree.AlterDatabaseOwner:
		__antithesis_instrumentation__.Notify(501889)
		return p.AlterDatabaseOwner(ctx, n)
	case *tree.AlterDatabaseAddRegion:
		__antithesis_instrumentation__.Notify(501890)
		return p.AlterDatabaseAddRegion(ctx, n)
	case *tree.AlterDatabaseDropRegion:
		__antithesis_instrumentation__.Notify(501891)
		return p.AlterDatabaseDropRegion(ctx, n)
	case *tree.AlterDatabasePrimaryRegion:
		__antithesis_instrumentation__.Notify(501892)
		return p.AlterDatabasePrimaryRegion(ctx, n)
	case *tree.AlterDatabasePlacement:
		__antithesis_instrumentation__.Notify(501893)
		return p.AlterDatabasePlacement(ctx, n)
	case *tree.AlterDatabaseSurvivalGoal:
		__antithesis_instrumentation__.Notify(501894)
		return p.AlterDatabaseSurvivalGoal(ctx, n)
	case *tree.AlterDatabaseAddSuperRegion:
		__antithesis_instrumentation__.Notify(501895)
		return p.AlterDatabaseAddSuperRegion(ctx, n)
	case *tree.AlterDatabaseDropSuperRegion:
		__antithesis_instrumentation__.Notify(501896)
		return p.AlterDatabaseDropSuperRegion(ctx, n)
	case *tree.AlterDatabaseAlterSuperRegion:
		__antithesis_instrumentation__.Notify(501897)
		return p.AlterDatabaseAlterSuperRegion(ctx, n)
	case *tree.AlterDefaultPrivileges:
		__antithesis_instrumentation__.Notify(501898)
		return p.alterDefaultPrivileges(ctx, n)
	case *tree.AlterIndex:
		__antithesis_instrumentation__.Notify(501899)
		return p.AlterIndex(ctx, n)
	case *tree.AlterSchema:
		__antithesis_instrumentation__.Notify(501900)
		return p.AlterSchema(ctx, n)
	case *tree.AlterTable:
		__antithesis_instrumentation__.Notify(501901)
		return p.AlterTable(ctx, n)
	case *tree.AlterTableLocality:
		__antithesis_instrumentation__.Notify(501902)
		return p.AlterTableLocality(ctx, n)
	case *tree.AlterTableOwner:
		__antithesis_instrumentation__.Notify(501903)
		return p.AlterTableOwner(ctx, n)
	case *tree.AlterTableSetSchema:
		__antithesis_instrumentation__.Notify(501904)
		return p.AlterTableSetSchema(ctx, n)
	case *tree.AlterTenantSetClusterSetting:
		__antithesis_instrumentation__.Notify(501905)
		return p.AlterTenantSetClusterSetting(ctx, n)
	case *tree.AlterType:
		__antithesis_instrumentation__.Notify(501906)
		return p.AlterType(ctx, n)
	case *tree.AlterRole:
		__antithesis_instrumentation__.Notify(501907)
		return p.AlterRole(ctx, n)
	case *tree.AlterRoleSet:
		__antithesis_instrumentation__.Notify(501908)
		return p.AlterRoleSet(ctx, n)
	case *tree.AlterSequence:
		__antithesis_instrumentation__.Notify(501909)
		return p.AlterSequence(ctx, n)
	case *tree.CloseCursor:
		__antithesis_instrumentation__.Notify(501910)
		return p.CloseCursor(ctx, n)
	case *tree.CommentOnColumn:
		__antithesis_instrumentation__.Notify(501911)
		return p.CommentOnColumn(ctx, n)
	case *tree.CommentOnConstraint:
		__antithesis_instrumentation__.Notify(501912)
		return p.CommentOnConstraint(ctx, n)
	case *tree.CommentOnDatabase:
		__antithesis_instrumentation__.Notify(501913)
		return p.CommentOnDatabase(ctx, n)
	case *tree.CommentOnSchema:
		__antithesis_instrumentation__.Notify(501914)
		return p.CommentOnSchema(ctx, n)
	case *tree.CommentOnIndex:
		__antithesis_instrumentation__.Notify(501915)
		return p.CommentOnIndex(ctx, n)
	case *tree.CommentOnTable:
		__antithesis_instrumentation__.Notify(501916)
		return p.CommentOnTable(ctx, n)
	case *tree.CreateDatabase:
		__antithesis_instrumentation__.Notify(501917)
		return p.CreateDatabase(ctx, n)
	case *tree.CreateIndex:
		__antithesis_instrumentation__.Notify(501918)
		return p.CreateIndex(ctx, n)
	case *tree.CreateSchema:
		__antithesis_instrumentation__.Notify(501919)
		return p.CreateSchema(ctx, n)
	case *tree.CreateType:
		__antithesis_instrumentation__.Notify(501920)
		return p.CreateType(ctx, n)
	case *tree.CreateRole:
		__antithesis_instrumentation__.Notify(501921)
		return p.CreateRole(ctx, n)
	case *tree.CreateSequence:
		__antithesis_instrumentation__.Notify(501922)
		return p.CreateSequence(ctx, n)
	case *tree.CreateExtension:
		__antithesis_instrumentation__.Notify(501923)
		return p.CreateExtension(ctx, n)
	case *tree.Deallocate:
		__antithesis_instrumentation__.Notify(501924)
		return p.Deallocate(ctx, n)
	case *tree.DeclareCursor:
		__antithesis_instrumentation__.Notify(501925)
		return p.DeclareCursor(ctx, n)
	case *tree.Discard:
		__antithesis_instrumentation__.Notify(501926)
		return p.Discard(ctx, n)
	case *tree.DropDatabase:
		__antithesis_instrumentation__.Notify(501927)
		return p.DropDatabase(ctx, n)
	case *tree.DropIndex:
		__antithesis_instrumentation__.Notify(501928)
		return p.DropIndex(ctx, n)
	case *tree.DropOwnedBy:
		__antithesis_instrumentation__.Notify(501929)
		return p.DropOwnedBy(ctx)
	case *tree.DropRole:
		__antithesis_instrumentation__.Notify(501930)
		return p.DropRole(ctx, n)
	case *tree.DropSchema:
		__antithesis_instrumentation__.Notify(501931)
		return p.DropSchema(ctx, n)
	case *tree.DropSequence:
		__antithesis_instrumentation__.Notify(501932)
		return p.DropSequence(ctx, n)
	case *tree.DropTable:
		__antithesis_instrumentation__.Notify(501933)
		return p.DropTable(ctx, n)
	case *tree.DropType:
		__antithesis_instrumentation__.Notify(501934)
		return p.DropType(ctx, n)
	case *tree.DropView:
		__antithesis_instrumentation__.Notify(501935)
		return p.DropView(ctx, n)
	case *tree.FetchCursor:
		__antithesis_instrumentation__.Notify(501936)
		return p.FetchCursor(ctx, &n.CursorStmt, false)
	case *tree.Grant:
		__antithesis_instrumentation__.Notify(501937)
		return p.Grant(ctx, n)
	case *tree.GrantRole:
		__antithesis_instrumentation__.Notify(501938)
		return p.GrantRole(ctx, n)
	case *tree.MoveCursor:
		__antithesis_instrumentation__.Notify(501939)
		return p.FetchCursor(ctx, &n.CursorStmt, true)
	case *tree.ReassignOwnedBy:
		__antithesis_instrumentation__.Notify(501940)
		return p.ReassignOwnedBy(ctx, n)
	case *tree.RefreshMaterializedView:
		__antithesis_instrumentation__.Notify(501941)
		return p.RefreshMaterializedView(ctx, n)
	case *tree.RenameColumn:
		__antithesis_instrumentation__.Notify(501942)
		return p.RenameColumn(ctx, n)
	case *tree.RenameDatabase:
		__antithesis_instrumentation__.Notify(501943)
		return p.RenameDatabase(ctx, n)
	case *tree.ReparentDatabase:
		__antithesis_instrumentation__.Notify(501944)
		return p.ReparentDatabase(ctx, n)
	case *tree.RenameIndex:
		__antithesis_instrumentation__.Notify(501945)
		return p.RenameIndex(ctx, n)
	case *tree.RenameTable:
		__antithesis_instrumentation__.Notify(501946)
		return p.RenameTable(ctx, n)
	case *tree.Revoke:
		__antithesis_instrumentation__.Notify(501947)
		return p.Revoke(ctx, n)
	case *tree.RevokeRole:
		__antithesis_instrumentation__.Notify(501948)
		return p.RevokeRole(ctx, n)
	case *tree.Scatter:
		__antithesis_instrumentation__.Notify(501949)
		return p.Scatter(ctx, n)
	case *tree.Scrub:
		__antithesis_instrumentation__.Notify(501950)
		return p.Scrub(ctx, n)
	case *tree.SetClusterSetting:
		__antithesis_instrumentation__.Notify(501951)
		return p.SetClusterSetting(ctx, n)
	case *tree.SetZoneConfig:
		__antithesis_instrumentation__.Notify(501952)
		return p.SetZoneConfig(ctx, n)
	case *tree.SetVar:
		__antithesis_instrumentation__.Notify(501953)
		return p.SetVar(ctx, n)
	case *tree.SetTransaction:
		__antithesis_instrumentation__.Notify(501954)
		return p.SetTransaction(ctx, n)
	case *tree.SetSessionAuthorizationDefault:
		__antithesis_instrumentation__.Notify(501955)
		return p.SetSessionAuthorizationDefault()
	case *tree.SetSessionCharacteristics:
		__antithesis_instrumentation__.Notify(501956)
		return p.SetSessionCharacteristics(n)
	case *tree.ShowClusterSetting:
		__antithesis_instrumentation__.Notify(501957)
		return p.ShowClusterSetting(ctx, n)
	case *tree.ShowTenantClusterSetting:
		__antithesis_instrumentation__.Notify(501958)
		return p.ShowTenantClusterSetting(ctx, n)
	case *tree.ShowCreateSchedules:
		__antithesis_instrumentation__.Notify(501959)
		return p.ShowCreateSchedule(ctx, n)
	case *tree.ShowHistogram:
		__antithesis_instrumentation__.Notify(501960)
		return p.ShowHistogram(ctx, n)
	case *tree.ShowTableStats:
		__antithesis_instrumentation__.Notify(501961)
		return p.ShowTableStats(ctx, n)
	case *tree.ShowTraceForSession:
		__antithesis_instrumentation__.Notify(501962)
		return p.ShowTrace(ctx, n)
	case *tree.ShowVar:
		__antithesis_instrumentation__.Notify(501963)
		return p.ShowVar(ctx, n)
	case *tree.ShowZoneConfig:
		__antithesis_instrumentation__.Notify(501964)
		return p.ShowZoneConfig(ctx, n)
	case *tree.ShowFingerprints:
		__antithesis_instrumentation__.Notify(501965)
		return p.ShowFingerprints(ctx, n)
	case *tree.Truncate:
		__antithesis_instrumentation__.Notify(501966)
		return p.Truncate(ctx, n)
	case tree.CCLOnlyStatement:
		__antithesis_instrumentation__.Notify(501967)
		plan, err := p.maybePlanHook(ctx, stmt)
		if plan == nil && func() bool {
			__antithesis_instrumentation__.Notify(501970)
			return err == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(501971)
			return nil, pgerror.Newf(pgcode.CCLRequired,
				"a CCL binary is required to use this statement type: %T", stmt)
		} else {
			__antithesis_instrumentation__.Notify(501972)
		}
		__antithesis_instrumentation__.Notify(501968)
		return plan, err
	default:
		__antithesis_instrumentation__.Notify(501969)
		return nil, errors.AssertionFailedf("unknown opaque statement %T", stmt)
	}
}

func init() {
	for _, stmt := range []tree.Statement{
		&tree.AlterChangefeed{},
		&tree.AlterDatabaseAddRegion{},
		&tree.AlterDatabaseDropRegion{},
		&tree.AlterDatabaseOwner{},
		&tree.AlterDatabasePrimaryRegion{},
		&tree.AlterDatabasePlacement{},
		&tree.AlterDatabaseSurvivalGoal{},
		&tree.AlterDatabaseAddSuperRegion{},
		&tree.AlterDatabaseDropSuperRegion{},
		&tree.AlterDatabaseAlterSuperRegion{},
		&tree.AlterDefaultPrivileges{},
		&tree.AlterIndex{},
		&tree.AlterSchema{},
		&tree.AlterTable{},
		&tree.AlterTableLocality{},
		&tree.AlterTableOwner{},
		&tree.AlterTableSetSchema{},
		&tree.AlterTenantSetClusterSetting{},
		&tree.AlterType{},
		&tree.AlterSequence{},
		&tree.AlterRole{},
		&tree.AlterRoleSet{},
		&tree.CloseCursor{},
		&tree.CommentOnColumn{},
		&tree.CommentOnDatabase{},
		&tree.CommentOnSchema{},
		&tree.CommentOnIndex{},
		&tree.CommentOnConstraint{},
		&tree.CommentOnTable{},
		&tree.CreateDatabase{},
		&tree.CreateExtension{},
		&tree.CreateIndex{},
		&tree.CreateSchema{},
		&tree.CreateSequence{},
		&tree.CreateType{},
		&tree.CreateRole{},
		&tree.Deallocate{},
		&tree.DeclareCursor{},
		&tree.Discard{},
		&tree.DropDatabase{},
		&tree.DropIndex{},
		&tree.DropOwnedBy{},
		&tree.DropRole{},
		&tree.DropSchema{},
		&tree.DropSequence{},
		&tree.DropTable{},
		&tree.DropType{},
		&tree.DropView{},
		&tree.FetchCursor{},
		&tree.Grant{},
		&tree.GrantRole{},
		&tree.MoveCursor{},
		&tree.ReassignOwnedBy{},
		&tree.RefreshMaterializedView{},
		&tree.RenameColumn{},
		&tree.RenameDatabase{},
		&tree.RenameIndex{},
		&tree.RenameTable{},
		&tree.ReparentDatabase{},
		&tree.Revoke{},
		&tree.RevokeRole{},
		&tree.Scatter{},
		&tree.Scrub{},
		&tree.SetClusterSetting{},
		&tree.SetZoneConfig{},
		&tree.SetVar{},
		&tree.SetTransaction{},
		&tree.SetSessionAuthorizationDefault{},
		&tree.SetSessionCharacteristics{},
		&tree.ShowClusterSetting{},
		&tree.ShowTenantClusterSetting{},
		&tree.ShowCreateSchedules{},
		&tree.ShowHistogram{},
		&tree.ShowTableStats{},
		&tree.ShowTraceForSession{},
		&tree.ShowZoneConfig{},
		&tree.ShowFingerprints{},
		&tree.ShowVar{},
		&tree.Truncate{},

		&tree.AlterBackup{},
		&tree.Backup{},
		&tree.ShowBackup{},
		&tree.Restore{},
		&tree.CreateChangefeed{},
		&tree.Import{},
		&tree.ScheduledBackup{},
		&tree.StreamIngestion{},
		&tree.ReplicationStream{},
	} {
		typ := optbuilder.OpaqueReadOnly
		if tree.CanModifySchema(stmt) {
			typ = optbuilder.OpaqueDDL
		} else if tree.CanWriteData(stmt) {
			typ = optbuilder.OpaqueMutation
		}
		optbuilder.RegisterOpaque(reflect.TypeOf(stmt), typ, buildOpaque)
	}
}
