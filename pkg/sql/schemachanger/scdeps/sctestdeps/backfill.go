package sctestdeps

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
)

type backfillTrackerDeps interface {
	Catalog() scexec.Catalog
	Codec() keys.SQLCodec
}

type testBackfillTracker struct {
	deps backfillTrackerDeps
}

func (s *TestState) BackfillProgressTracker() scexec.BackfillTracker {
	__antithesis_instrumentation__.Notify(580767)
	return s.backfillTracker
}

var _ scexec.BackfillTracker = (*testBackfillTracker)(nil)

func (s *testBackfillTracker) GetBackfillProgress(
	ctx context.Context, b scexec.Backfill,
) (scexec.BackfillProgress, error) {
	__antithesis_instrumentation__.Notify(580768)
	return scexec.BackfillProgress{Backfill: b}, nil
}

func (s *testBackfillTracker) SetBackfillProgress(
	ctx context.Context, progress scexec.BackfillProgress,
) error {
	__antithesis_instrumentation__.Notify(580769)
	return nil
}

func (s *testBackfillTracker) FlushCheckpoint(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(580770)
	return nil
}

func (s *testBackfillTracker) FlushFractionCompleted(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(580771)
	return nil
}

func (s *TestState) StartPeriodicFlush(
	ctx context.Context,
) (close func(context.Context) error, _ error) {
	__antithesis_instrumentation__.Notify(580772)
	return func(ctx context.Context) error { __antithesis_instrumentation__.Notify(580773); return nil }, nil
}

type testBackfiller struct {
	s *TestState
}

var _ scexec.Backfiller = (*testBackfiller)(nil)

func (s *testBackfiller) BackfillIndex(
	_ context.Context,
	progress scexec.BackfillProgress,
	_ scexec.BackfillProgressWriter,
	tbl catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(580774)
	s.s.LogSideEffectf(
		"backfill indexes %v from index #%d in table #%d",
		progress.DestIndexIDs, progress.SourceIndexID, tbl.GetID(),
	)
	return nil
}

func (s *testBackfiller) MaybePrepareDestIndexesForBackfill(
	ctx context.Context, progress scexec.BackfillProgress, descriptor catalog.TableDescriptor,
) (scexec.BackfillProgress, error) {
	__antithesis_instrumentation__.Notify(580775)
	return progress, nil
}

var _ scexec.IndexSpanSplitter = (*indexSpanSplitter)(nil)

type indexSpanSplitter struct{}

func (s *indexSpanSplitter) MaybeSplitIndexSpans(
	_ context.Context, _ catalog.TableDescriptor, _ catalog.Index,
) error {
	__antithesis_instrumentation__.Notify(580776)
	return nil
}
