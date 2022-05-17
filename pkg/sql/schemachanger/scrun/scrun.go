package scrun

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func RunStatementPhase(
	ctx context.Context, knobs *TestingKnobs, deps scexec.Dependencies, state scpb.CurrentState,
) (scpb.CurrentState, jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(595105)
	return runTransactionPhase(ctx, knobs, deps, state, scop.StatementPhase)
}

func RunPreCommitPhase(
	ctx context.Context, knobs *TestingKnobs, deps scexec.Dependencies, state scpb.CurrentState,
) (scpb.CurrentState, jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(595106)
	return runTransactionPhase(ctx, knobs, deps, state, scop.PreCommitPhase)
}

func runTransactionPhase(
	ctx context.Context,
	knobs *TestingKnobs,
	deps scexec.Dependencies,
	state scpb.CurrentState,
	phase scop.Phase,
) (scpb.CurrentState, jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(595107)
	if len(state.Current) == 0 {
		__antithesis_instrumentation__.Notify(595112)
		return scpb.CurrentState{}, jobspb.InvalidJobID, nil
	} else {
		__antithesis_instrumentation__.Notify(595113)
	}
	__antithesis_instrumentation__.Notify(595108)
	sc, err := scplan.MakePlan(state, scplan.Params{
		ExecutionPhase:             phase,
		SchemaChangerJobIDSupplier: deps.TransactionalJobRegistry().SchemaChangerJobID,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(595114)
		return scpb.CurrentState{}, jobspb.InvalidJobID, err
	} else {
		__antithesis_instrumentation__.Notify(595115)
	}
	__antithesis_instrumentation__.Notify(595109)
	after := state.Current
	if len(after) == 0 {
		__antithesis_instrumentation__.Notify(595116)
		return scpb.CurrentState{}, jobspb.InvalidJobID, nil
	} else {
		__antithesis_instrumentation__.Notify(595117)
	}
	__antithesis_instrumentation__.Notify(595110)
	stages := sc.StagesForCurrentPhase()
	for i := range stages {
		__antithesis_instrumentation__.Notify(595118)
		if err := executeStage(ctx, knobs, deps, sc, i, stages[i]); err != nil {
			__antithesis_instrumentation__.Notify(595120)
			return scpb.CurrentState{}, jobspb.InvalidJobID, err
		} else {
			__antithesis_instrumentation__.Notify(595121)
		}
		__antithesis_instrumentation__.Notify(595119)
		after = stages[i].After
	}
	__antithesis_instrumentation__.Notify(595111)
	return scpb.CurrentState{TargetState: state.TargetState, Current: after}, sc.JobID, nil
}

func RunSchemaChangesInJob(
	ctx context.Context,
	knobs *TestingKnobs,
	settings *cluster.Settings,
	deps JobRunDependencies,
	jobID jobspb.JobID,
	descriptorIDs []descpb.ID,
	rollback bool,
) error {
	__antithesis_instrumentation__.Notify(595122)
	state, err := makeState(ctx, deps, descriptorIDs, rollback)
	if err != nil {
		__antithesis_instrumentation__.Notify(595127)
		return errors.Wrapf(err, "failed to construct state for job %d", jobID)
	} else {
		__antithesis_instrumentation__.Notify(595128)
	}
	__antithesis_instrumentation__.Notify(595123)
	sc, err := scplan.MakePlan(state, scplan.Params{
		ExecutionPhase:             scop.PostCommitPhase,
		SchemaChangerJobIDSupplier: func() jobspb.JobID { __antithesis_instrumentation__.Notify(595129); return jobID },
	})
	__antithesis_instrumentation__.Notify(595124)
	if err != nil {
		__antithesis_instrumentation__.Notify(595130)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595131)
	}
	__antithesis_instrumentation__.Notify(595125)

	for i := range sc.Stages {
		__antithesis_instrumentation__.Notify(595132)

		if err := deps.WithTxnInJob(ctx, func(ctx context.Context, td scexec.Dependencies) error {
			__antithesis_instrumentation__.Notify(595133)
			if err := td.TransactionalJobRegistry().CheckPausepoint(
				pausepointName(state, i),
			); err != nil {
				__antithesis_instrumentation__.Notify(595135)
				return err
			} else {
				__antithesis_instrumentation__.Notify(595136)
			}
			__antithesis_instrumentation__.Notify(595134)
			return executeStage(ctx, knobs, td, sc, i, sc.Stages[i])
		}); err != nil {
			__antithesis_instrumentation__.Notify(595137)
			if knobs.OnPostCommitError != nil {
				__antithesis_instrumentation__.Notify(595139)
				return knobs.OnPostCommitError(sc, i, err)
			} else {
				__antithesis_instrumentation__.Notify(595140)
			}
			__antithesis_instrumentation__.Notify(595138)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595141)
		}
	}
	__antithesis_instrumentation__.Notify(595126)
	return nil
}

func pausepointName(state scpb.CurrentState, i int) string {
	__antithesis_instrumentation__.Notify(595142)
	return fmt.Sprintf(
		"schemachanger.%s.%s.%d",
		state.Authorization.UserName, state.Authorization.AppName, i,
	)
}

func executeStage(
	ctx context.Context,
	knobs *TestingKnobs,
	deps scexec.Dependencies,
	p scplan.Plan,
	stageIdx int,
	stage scplan.Stage,
) (err error) {
	__antithesis_instrumentation__.Notify(595143)
	if knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(595148)
		return knobs.BeforeStage != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(595149)
		if err := knobs.BeforeStage(p, stageIdx); err != nil {
			__antithesis_instrumentation__.Notify(595150)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595151)
		}
	} else {
		__antithesis_instrumentation__.Notify(595152)
	}
	__antithesis_instrumentation__.Notify(595144)

	log.Infof(ctx, "executing %s (rollback=%v)", stage, p.InRollback)
	start := timeutil.Now()
	defer func() {
		__antithesis_instrumentation__.Notify(595153)
		if log.ExpensiveLogEnabled(ctx, 2) {
			__antithesis_instrumentation__.Notify(595154)
			log.Infof(ctx, "executing %s (rollback=%v) took %v: err = %v",
				stage, p.InRollback, timeutil.Since(start), err)
		} else {
			__antithesis_instrumentation__.Notify(595155)
		}
	}()
	__antithesis_instrumentation__.Notify(595145)
	if err := scexec.ExecuteStage(ctx, deps, stage.Ops()); err != nil {
		__antithesis_instrumentation__.Notify(595156)

		if !errors.HasType(err, (*roachpb.TransactionRetryWithProtoRefreshError)(nil)) && func() bool {
			__antithesis_instrumentation__.Notify(595158)
			return !errors.Is(err, context.Canceled) == true
		}() == true {
			__antithesis_instrumentation__.Notify(595159)
			err = p.DecorateErrorWithPlanDetails(err)
		} else {
			__antithesis_instrumentation__.Notify(595160)
		}
		__antithesis_instrumentation__.Notify(595157)
		return errors.Wrapf(err, "error executing %s", stage)
	} else {
		__antithesis_instrumentation__.Notify(595161)
	}
	__antithesis_instrumentation__.Notify(595146)
	if knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(595162)
		return knobs.AfterStage != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(595163)
		if err := knobs.AfterStage(p, stageIdx); err != nil {
			__antithesis_instrumentation__.Notify(595164)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595165)
		}
	} else {
		__antithesis_instrumentation__.Notify(595166)
	}
	__antithesis_instrumentation__.Notify(595147)

	return nil
}

func makeState(
	ctx context.Context, deps JobRunDependencies, descriptorIDs []descpb.ID, rollback bool,
) (scpb.CurrentState, error) {
	__antithesis_instrumentation__.Notify(595167)
	var descriptorStates []*scpb.DescriptorState
	if err := deps.WithTxnInJob(ctx, func(ctx context.Context, txnDeps scexec.Dependencies) error {
		__antithesis_instrumentation__.Notify(595172)
		descriptorStates = nil

		descs, err := txnDeps.Catalog().MustReadImmutableDescriptors(ctx, descriptorIDs...)
		if err != nil {
			__antithesis_instrumentation__.Notify(595175)

			return err
		} else {
			__antithesis_instrumentation__.Notify(595176)
		}
		__antithesis_instrumentation__.Notify(595173)
		for _, desc := range descs {
			__antithesis_instrumentation__.Notify(595177)

			cs := desc.GetDeclarativeSchemaChangerState()
			if cs == nil {
				__antithesis_instrumentation__.Notify(595179)
				return errors.Errorf(
					"descriptor %q (%d) does not contain schema changer state", desc.GetName(), desc.GetID(),
				)
			} else {
				__antithesis_instrumentation__.Notify(595180)
			}
			__antithesis_instrumentation__.Notify(595178)
			descriptorStates = append(descriptorStates, cs)
		}
		__antithesis_instrumentation__.Notify(595174)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(595181)
		return scpb.CurrentState{}, err
	} else {
		__antithesis_instrumentation__.Notify(595182)
	}
	__antithesis_instrumentation__.Notify(595168)
	state, err := scpb.MakeCurrentStateFromDescriptors(descriptorStates)
	if err != nil {
		__antithesis_instrumentation__.Notify(595183)
		return scpb.CurrentState{}, err
	} else {
		__antithesis_instrumentation__.Notify(595184)
	}
	__antithesis_instrumentation__.Notify(595169)
	if !rollback && func() bool {
		__antithesis_instrumentation__.Notify(595185)
		return state.InRollback == true
	}() == true {
		__antithesis_instrumentation__.Notify(595186)

		return scpb.CurrentState{}, jobs.MarkAsPermanentJobError(errors.Errorf(
			"job in running state but schema change in rollback, " +
				"returning an error to restart in the reverting state"))
	} else {
		__antithesis_instrumentation__.Notify(595187)
	}
	__antithesis_instrumentation__.Notify(595170)
	if rollback && func() bool {
		__antithesis_instrumentation__.Notify(595188)
		return !state.InRollback == true
	}() == true {
		__antithesis_instrumentation__.Notify(595189)
		state.Rollback()
	} else {
		__antithesis_instrumentation__.Notify(595190)
	}
	__antithesis_instrumentation__.Notify(595171)
	return state, nil
}
