package scrun

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan"

type TestingKnobs struct {
	BeforeStage func(p scplan.Plan, stageIdx int) error

	AfterStage func(p scplan.Plan, stageIdx int) error

	BeforeWaitingForConcurrentSchemaChanges func(stmts []string)

	OnPostCommitError func(p scplan.Plan, stageIdx int, err error) error
}

func (*TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(595191) }
