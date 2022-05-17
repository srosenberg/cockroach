package scbuild

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild/internal/scbuildstmt"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func Build(
	ctx context.Context, dependencies Dependencies, initial scpb.CurrentState, n tree.Statement,
) (_ scpb.CurrentState, err error) {
	__antithesis_instrumentation__.Notify(579306)
	start := timeutil.Now()
	defer func() {
		__antithesis_instrumentation__.Notify(579312)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(579314)
			return !log.ExpensiveLogEnabled(ctx, 2) == true
		}() == true {
			__antithesis_instrumentation__.Notify(579315)
			return
		} else {
			__antithesis_instrumentation__.Notify(579316)
		}
		__antithesis_instrumentation__.Notify(579313)
		log.Infof(ctx, "build for %s took %v", n.StatementTag(), timeutil.Since(start))
	}()
	__antithesis_instrumentation__.Notify(579307)
	initial = initial.DeepCopy()
	bs := newBuilderState(ctx, dependencies, initial)
	els := newEventLogState(dependencies, initial, n)

	an, err := newAstAnnotator(n)
	if err != nil {
		__antithesis_instrumentation__.Notify(579317)
		return scpb.CurrentState{}, err
	} else {
		__antithesis_instrumentation__.Notify(579318)
	}
	__antithesis_instrumentation__.Notify(579308)
	b := buildCtx{
		Context:              ctx,
		Dependencies:         dependencies,
		BuilderState:         bs,
		EventLogState:        els,
		TreeAnnotator:        an,
		SchemaFeatureChecker: dependencies.FeatureChecker(),
	}
	defer func() {
		__antithesis_instrumentation__.Notify(579319)
		if recErr := recover(); recErr != nil {
			__antithesis_instrumentation__.Notify(579320)
			if errObj, ok := recErr.(error); ok {
				__antithesis_instrumentation__.Notify(579321)
				err = errObj
			} else {
				__antithesis_instrumentation__.Notify(579322)
				err = errors.Errorf("unexpected error encountered while building schema change plan %s", recErr)
			}
		} else {
			__antithesis_instrumentation__.Notify(579323)
		}
	}()
	__antithesis_instrumentation__.Notify(579309)
	scbuildstmt.Process(b, an.GetStatement())
	an.ValidateAnnotations()
	els.statements[len(els.statements)-1].RedactedStatement =
		string(els.astFormatter.FormatAstAsRedactableString(an.GetStatement(), &an.annotation))
	ts := scpb.TargetState{
		Targets:       make([]scpb.Target, 0, len(bs.output)),
		Statements:    els.statements,
		Authorization: els.authorization,
	}
	current := make([]scpb.Status, 0, len(bs.output))
	for _, e := range bs.output {
		__antithesis_instrumentation__.Notify(579324)
		if e.metadata.Size() == 0 {
			__antithesis_instrumentation__.Notify(579326)

			continue
		} else {
			__antithesis_instrumentation__.Notify(579327)
		}
		__antithesis_instrumentation__.Notify(579325)
		ts.Targets = append(ts.Targets, scpb.MakeTarget(e.target, e.element, &e.metadata))
		current = append(current, e.current)
	}
	__antithesis_instrumentation__.Notify(579310)

	descSet := screl.AllTargetDescIDs(ts)
	descSet.ForEach(func(id descpb.ID) {
		__antithesis_instrumentation__.Notify(579328)
		bs.ensureDescriptor(id)
		desc := bs.descCache[id].desc
		if desc.HasConcurrentSchemaChanges() {
			__antithesis_instrumentation__.Notify(579329)
			panic(scerrors.ConcurrentSchemaChangeError(desc))
		} else {
			__antithesis_instrumentation__.Notify(579330)
		}
	})
	__antithesis_instrumentation__.Notify(579311)
	return scpb.CurrentState{TargetState: ts, Current: current}, nil
}

type (
	// FeatureChecker contains operations for checking if a schema change
	// feature is allowed by the database administrator.
	FeatureChecker = scbuildstmt.SchemaFeatureChecker
)

type elementState struct {
	element  scpb.Element
	current  scpb.Status
	target   scpb.TargetStatus
	metadata scpb.TargetMetadata
}

type builderState struct {
	ctx             context.Context
	clusterSettings *cluster.Settings
	evalCtx         *tree.EvalContext
	semaCtx         *tree.SemaContext
	cr              CatalogReader
	auth            AuthorizationAccessor
	createPartCCL   CreatePartitioningCCLCallback
	hasAdmin        bool

	output []elementState

	descCache   map[catid.DescID]*cachedDesc
	tempSchemas map[catid.DescID]catalog.SchemaDescriptor
}

type cachedDesc struct {
	desc         catalog.Descriptor
	prefix       tree.ObjectNamePrefix
	backrefs     catalog.DescriptorIDSet
	ers          *elementResultSet
	privileges   map[privilege.Kind]error
	hasOwnership bool

	elementIndexMap map[string]int
}

func newBuilderState(ctx context.Context, d Dependencies, initial scpb.CurrentState) *builderState {
	__antithesis_instrumentation__.Notify(579331)
	bs := builderState{
		ctx:             ctx,
		clusterSettings: d.ClusterSettings(),
		evalCtx:         newEvalCtx(ctx, d),
		semaCtx:         newSemaCtx(d),
		cr:              d.CatalogReader(),
		auth:            d.AuthorizationAccessor(),
		createPartCCL:   d.IndexPartitioningCCLCallback(),
		output:          make([]elementState, 0, len(initial.Current)),
		descCache:       make(map[catid.DescID]*cachedDesc),
		tempSchemas:     make(map[catid.DescID]catalog.SchemaDescriptor),
	}
	var err error
	bs.hasAdmin, err = bs.auth.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(579335)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579336)
	}
	__antithesis_instrumentation__.Notify(579332)
	for _, t := range initial.TargetState.Targets {
		__antithesis_instrumentation__.Notify(579337)
		bs.ensureDescriptor(screl.GetDescID(t.Element()))
	}
	__antithesis_instrumentation__.Notify(579333)
	for i, t := range initial.TargetState.Targets {
		__antithesis_instrumentation__.Notify(579338)
		bs.Ensure(initial.Current[i], scpb.AsTargetStatus(t.TargetStatus), t.Element(), t.Metadata)
	}
	__antithesis_instrumentation__.Notify(579334)
	return &bs
}

type eventLogState struct {
	statements []scpb.Statement

	authorization scpb.Authorization

	statementMetaData scpb.TargetMetadata

	sourceElementID *scpb.SourceElementID

	astFormatter AstFormatter
}

func newEventLogState(d Dependencies, initial scpb.CurrentState, n tree.Statement) *eventLogState {
	__antithesis_instrumentation__.Notify(579339)
	stmts := initial.Statements
	els := eventLogState{
		statements: append(stmts, scpb.Statement{
			Statement:    n.String(),
			StatementTag: n.StatementTag(),
		}),
		authorization: scpb.Authorization{
			AppName:  d.SessionData().ApplicationName,
			UserName: d.SessionData().SessionUser().Normalized(),
		},
		sourceElementID: new(scpb.SourceElementID),
		statementMetaData: scpb.TargetMetadata{
			StatementID:     uint32(len(stmts)),
			SubWorkID:       1,
			SourceElementID: 1,
		},
		astFormatter: d.AstFormatter(),
	}
	*els.sourceElementID = 1
	return &els
}

type buildCtx struct {
	context.Context
	Dependencies
	scbuildstmt.BuilderState
	scbuildstmt.EventLogState
	scbuildstmt.TreeAnnotator
	scbuildstmt.SchemaFeatureChecker
}

var _ scbuildstmt.BuildCtx = buildCtx{}

func (b buildCtx) Add(element scpb.Element) {
	__antithesis_instrumentation__.Notify(579340)
	b.Ensure(scpb.Status_UNKNOWN, scpb.ToPublic, element, b.TargetMetadata())
}

func (b buildCtx) Drop(element scpb.Element) {
	__antithesis_instrumentation__.Notify(579341)
	b.Ensure(scpb.Status_UNKNOWN, scpb.ToAbsent, element, b.TargetMetadata())
}

func (b buildCtx) WithNewSourceElementID() scbuildstmt.BuildCtx {
	__antithesis_instrumentation__.Notify(579342)
	return buildCtx{
		Context:       b.Context,
		Dependencies:  b.Dependencies,
		BuilderState:  b.BuilderState,
		TreeAnnotator: b.TreeAnnotator,
		EventLogState: b.EventLogStateWithNewSourceElementID(),
	}
}
