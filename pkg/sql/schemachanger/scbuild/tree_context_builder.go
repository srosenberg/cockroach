package scbuild

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/sql/faketreeeval"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild/internal/scbuildstmt"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

var _ scbuildstmt.TreeContextBuilder = buildCtx{}

func (b buildCtx) SemaCtx() *tree.SemaContext {
	__antithesis_instrumentation__.Notify(580254)
	return newSemaCtx(b.Dependencies)
}

func newSemaCtx(d Dependencies) *tree.SemaContext {
	__antithesis_instrumentation__.Notify(580255)
	semaCtx := tree.MakeSemaContext()
	semaCtx.Annotations = nil
	semaCtx.SearchPath = d.SessionData().SearchPath
	if d.ClusterSettings().Version.IsActive(context.Background(), clusterversion.IncrementalBackupSubdir) {
		__antithesis_instrumentation__.Notify(580257)
		semaCtx.IntervalStyleEnabled = true
		semaCtx.DateStyleEnabled = true
	} else {
		__antithesis_instrumentation__.Notify(580258)
		semaCtx.IntervalStyleEnabled = d.SessionData().IntervalStyleEnabled
		semaCtx.DateStyleEnabled = d.SessionData().DateStyleEnabled
	}
	__antithesis_instrumentation__.Notify(580256)
	semaCtx.TypeResolver = d.CatalogReader()
	semaCtx.TableNameResolver = d.CatalogReader()
	semaCtx.DateStyle = d.SessionData().GetDateStyle()
	semaCtx.IntervalStyle = d.SessionData().GetIntervalStyle()
	return &semaCtx
}

func (b buildCtx) EvalCtx() *tree.EvalContext {
	__antithesis_instrumentation__.Notify(580259)
	return newEvalCtx(b.Context, b.Dependencies)
}

func newEvalCtx(ctx context.Context, d Dependencies) *tree.EvalContext {
	__antithesis_instrumentation__.Notify(580260)
	return &tree.EvalContext{
		ClusterID:          d.ClusterID(),
		SessionDataStack:   sessiondata.NewStack(d.SessionData()),
		Context:            ctx,
		Planner:            &faketreeeval.DummyEvalPlanner{},
		PrivilegedAccessor: &faketreeeval.DummyPrivilegedAccessor{},
		SessionAccessor:    &faketreeeval.DummySessionAccessor{},
		ClientNoticeSender: &faketreeeval.DummyClientNoticeSender{},
		Sequence:           &faketreeeval.DummySequenceOperators{},
		Tenant:             &faketreeeval.DummyTenantOperator{},
		Regions:            &faketreeeval.DummyRegionOperator{},
		Settings:           d.ClusterSettings(),
		Codec:              d.Codec(),
	}
}
