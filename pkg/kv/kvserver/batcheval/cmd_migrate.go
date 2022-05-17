package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/errors"
)

func init() {
	RegisterReadWriteCommand(roachpb.Migrate, declareKeysMigrate, Migrate)
}

func declareKeysMigrate(
	rs ImmutableRangeState,
	_ *roachpb.Header,
	_ roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97044)

	latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.RangeVersionKey(rs.GetRangeID())})
	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeDescriptorKey(rs.GetStartKey())})
}

var migrationRegistry = make(map[roachpb.Version]migration)

type migration func(context.Context, storage.ReadWriter, CommandArgs) (result.Result, error)

func init() {
	_ = registerMigration
	registerMigration(
		clusterversion.AddRaftAppliedIndexTermMigration, addRaftAppliedIndexTermMigration)
}

func registerMigration(key clusterversion.Key, migration migration) {
	__antithesis_instrumentation__.Notify(97045)
	migrationRegistry[clusterversion.ByKey(key)] = migration
}

func Migrate(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, _ roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97046)
	args := cArgs.Args.(*roachpb.MigrateRequest)
	migrationVersion := args.Version

	fn, ok := migrationRegistry[migrationVersion]
	if !ok {
		__antithesis_instrumentation__.Notify(97051)
		return result.Result{}, errors.AssertionFailedf("migration for %s not found", migrationVersion)
	} else {
		__antithesis_instrumentation__.Notify(97052)
	}
	__antithesis_instrumentation__.Notify(97047)
	pd, err := fn(ctx, readWriter, cArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(97053)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97054)
	}
	__antithesis_instrumentation__.Notify(97048)

	if err := MakeStateLoader(cArgs.EvalCtx).SetVersion(
		ctx, readWriter, cArgs.Stats, &migrationVersion,
	); err != nil {
		__antithesis_instrumentation__.Notify(97055)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97056)
	}
	__antithesis_instrumentation__.Notify(97049)
	if pd.Replicated.State == nil {
		__antithesis_instrumentation__.Notify(97057)
		pd.Replicated.State = &kvserverpb.ReplicaState{}
	} else {
		__antithesis_instrumentation__.Notify(97058)
	}
	__antithesis_instrumentation__.Notify(97050)

	pd.Replicated.State.Version = &migrationVersion
	return pd, nil
}

func addRaftAppliedIndexTermMigration(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97059)
	return result.Result{
		Replicated: kvserverpb.ReplicatedEvalResult{
			State: &kvserverpb.ReplicaState{

				RaftAppliedIndexTerm: stateloader.RaftLogTermSignalForAddRaftAppliedIndexTermMigration,
			},
		},
	}, nil
}

func TestingRegisterMigrationInterceptor(version roachpb.Version, fn func()) (unregister func()) {
	__antithesis_instrumentation__.Notify(97060)
	if _, ok := migrationRegistry[version]; ok {
		__antithesis_instrumentation__.Notify(97063)
		panic("doubly registering migration")
	} else {
		__antithesis_instrumentation__.Notify(97064)
	}
	__antithesis_instrumentation__.Notify(97061)
	migrationRegistry[version] = func(context.Context, storage.ReadWriter, CommandArgs) (result.Result, error) {
		__antithesis_instrumentation__.Notify(97065)
		fn()
		return result.Result{}, nil
	}
	__antithesis_instrumentation__.Notify(97062)
	return func() { __antithesis_instrumentation__.Notify(97066); delete(migrationRegistry, version) }
}
