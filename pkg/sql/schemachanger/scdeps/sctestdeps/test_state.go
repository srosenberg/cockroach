package sctestdeps

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scrun"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/redact"
)

type TestState struct {
	catalog, synthetic      nstree.MutableCatalog
	currentDatabase         string
	phase                   scop.Phase
	sessionData             sessiondata.SessionData
	statements              []string
	testingKnobs            *scrun.TestingKnobs
	jobs                    []jobs.Record
	createdJobsInCurrentTxn []jobspb.JobID
	jobCounter              int
	txnCounter              int
	sideEffectLogBuffer     strings.Builder

	backfiller        scexec.Backfiller
	indexSpanSplitter scexec.IndexSpanSplitter
	backfillTracker   scexec.BackfillTracker

	approximateTimestamp time.Time
}

func NewTestDependencies(options ...Option) *TestState {
	__antithesis_instrumentation__.Notify(581233)
	var s TestState
	for _, o := range defaultOptions {
		__antithesis_instrumentation__.Notify(581236)
		o.apply(&s)
	}
	__antithesis_instrumentation__.Notify(581234)
	for _, o := range options {
		__antithesis_instrumentation__.Notify(581237)
		o.apply(&s)
	}
	__antithesis_instrumentation__.Notify(581235)
	return &s
}

func (s *TestState) LogSideEffectf(fmtstr string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(581238)
	s.sideEffectLogBuffer.WriteString(fmt.Sprintf(fmtstr, args...))
	s.sideEffectLogBuffer.WriteRune('\n')
}

func (s *TestState) SideEffectLog() string {
	__antithesis_instrumentation__.Notify(581239)
	return s.sideEffectLogBuffer.String()
}

func (s *TestState) WithTxn(fn func(s *TestState)) {
	__antithesis_instrumentation__.Notify(581240)
	s.txnCounter++
	defer s.synthetic.Clear()
	defer func() {
		__antithesis_instrumentation__.Notify(581242)
		if len(s.createdJobsInCurrentTxn) == 0 {
			__antithesis_instrumentation__.Notify(581244)
			return
		} else {
			__antithesis_instrumentation__.Notify(581245)
		}
		__antithesis_instrumentation__.Notify(581243)
		s.LogSideEffectf(
			"notified job registry to adopt jobs: %v", s.createdJobsInCurrentTxn,
		)
		s.createdJobsInCurrentTxn = nil
	}()
	__antithesis_instrumentation__.Notify(581241)
	defer s.LogSideEffectf("commit transaction #%d", s.txnCounter)
	s.LogSideEffectf("begin transaction #%d", s.txnCounter)
	fn(s)
}

func (s *TestState) IncrementPhase() {
	__antithesis_instrumentation__.Notify(581246)
	s.phase++
}

func (s *TestState) JobRecord(jobID jobspb.JobID) *jobs.Record {
	__antithesis_instrumentation__.Notify(581247)
	idx := int(jobID) - 1
	if idx < 0 || func() bool {
		__antithesis_instrumentation__.Notify(581249)
		return idx >= len(s.jobs) == true
	}() == true {
		__antithesis_instrumentation__.Notify(581250)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(581251)
	}
	__antithesis_instrumentation__.Notify(581248)
	return &s.jobs[idx]
}

func (s *TestState) FormatAstAsRedactableString(
	statement tree.Statement, ann *tree.Annotations,
) redact.RedactableString {
	__antithesis_instrumentation__.Notify(581252)

	f := tree.NewFmtCtx(
		tree.FmtAlwaysQualifyTableNames|tree.FmtMarkRedactionNode,
		tree.FmtAnnotations(ann))
	f.FormatNode(statement)
	formattedRedactableStatementString := f.CloseAndGetString()
	return redact.RedactableString(formattedRedactableStatementString)
}

func (s *TestState) AstFormatter() scbuild.AstFormatter {
	__antithesis_instrumentation__.Notify(581253)
	return s
}

func (s *TestState) CheckFeature(ctx context.Context, featureName tree.SchemaFeatureName) error {
	__antithesis_instrumentation__.Notify(581254)
	s.LogSideEffectf("checking for feature: %s", featureName)
	return nil
}

func (s *TestState) FeatureChecker() scbuild.FeatureChecker {
	__antithesis_instrumentation__.Notify(581255)
	return s
}
