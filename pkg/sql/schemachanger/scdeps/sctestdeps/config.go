package sctestdeps

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scrun"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

type Option interface {
	apply(*TestState)
}

type optionFunc func(*TestState)

func (o optionFunc) apply(state *TestState) { __antithesis_instrumentation__.Notify(580777); o(state) }

var _ Option = (optionFunc)(nil)

func WithNamespace(c nstree.Catalog) Option {
	__antithesis_instrumentation__.Notify(580778)
	return optionFunc(func(state *TestState) {
		__antithesis_instrumentation__.Notify(580779)
		_ = c.ForEachNamespaceEntry(func(e catalog.NameEntry) error {
			__antithesis_instrumentation__.Notify(580780)
			state.catalog.UpsertNamespaceEntry(e, e.GetID())
			return nil
		})
	})
}

func WithDescriptors(c nstree.Catalog) Option {
	__antithesis_instrumentation__.Notify(580781)
	return optionFunc(func(state *TestState) {
		__antithesis_instrumentation__.Notify(580782)
		_ = c.ForEachDescriptorEntry(func(desc catalog.Descriptor) error {
			__antithesis_instrumentation__.Notify(580783)
			if table, isTable := desc.(catalog.TableDescriptor); isTable {
				__antithesis_instrumentation__.Notify(580785)
				mut := table.NewBuilder().BuildExistingMutable().(*tabledesc.Mutable)
				for _, idx := range mut.AllIndexes() {
					__antithesis_instrumentation__.Notify(580787)
					if !idx.CreatedAt().IsZero() {
						__antithesis_instrumentation__.Notify(580788)
						idx.IndexDesc().CreatedAtNanos = defaultOverriddenCreatedAt.UnixNano()
					} else {
						__antithesis_instrumentation__.Notify(580789)
					}
				}
				__antithesis_instrumentation__.Notify(580786)
				desc = mut.ImmutableCopy()
			} else {
				__antithesis_instrumentation__.Notify(580790)
			}
			__antithesis_instrumentation__.Notify(580784)
			state.catalog.UpsertDescriptorEntry(desc)
			return nil
		})
	})
}

func WithSessionData(sessionData sessiondata.SessionData) Option {
	__antithesis_instrumentation__.Notify(580791)
	return optionFunc(func(state *TestState) {
		__antithesis_instrumentation__.Notify(580792)
		state.sessionData = sessionData
	})
}

func WithTestingKnobs(testingKnobs *scrun.TestingKnobs) Option {
	__antithesis_instrumentation__.Notify(580793)
	return optionFunc(func(state *TestState) {
		__antithesis_instrumentation__.Notify(580794)
		state.testingKnobs = testingKnobs
	})
}

func WithStatements(statements ...string) Option {
	__antithesis_instrumentation__.Notify(580795)
	return optionFunc(func(state *TestState) {
		__antithesis_instrumentation__.Notify(580796)
		state.statements = statements
	})
}

func WithCurrentDatabase(db string) Option {
	__antithesis_instrumentation__.Notify(580797)
	return optionFunc(func(state *TestState) {
		__antithesis_instrumentation__.Notify(580798)
		state.currentDatabase = db
	})
}

func WithBackfillTracker(backfillTracker scexec.BackfillTracker) Option {
	__antithesis_instrumentation__.Notify(580799)
	return optionFunc(func(state *TestState) {
		__antithesis_instrumentation__.Notify(580800)
		state.backfillTracker = backfillTracker
	})
}

func WithBackfiller(backfiller scexec.Backfiller) Option {
	__antithesis_instrumentation__.Notify(580801)
	return optionFunc(func(state *TestState) {
		__antithesis_instrumentation__.Notify(580802)
		state.backfiller = backfiller
	})
}

var (
	defaultOverriddenCreatedAt = time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)

	defaultCreatedAt = defaultOverriddenCreatedAt.Add(time.Hour)
)

var defaultOptions = []Option{
	optionFunc(func(state *TestState) {
		__antithesis_instrumentation__.Notify(580803)
		state.backfillTracker = &testBackfillTracker{deps: state}
		state.backfiller = &testBackfiller{s: state}
		state.indexSpanSplitter = &indexSpanSplitter{}
		state.approximateTimestamp = defaultCreatedAt
	}),
}
