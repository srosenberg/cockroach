package validate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

const invalidSchemaChangerStatePrefix = "invalid schema changer state"

func validateSchemaChangerState(d catalog.Descriptor, vea catalog.ValidationErrorAccumulator) {
	__antithesis_instrumentation__.Notify(265819)
	scs := d.GetDeclarativeSchemaChangerState()
	if scs == nil {
		__antithesis_instrumentation__.Notify(265828)
		return
	} else {
		__antithesis_instrumentation__.Notify(265829)
	}
	__antithesis_instrumentation__.Notify(265820)
	report := func(err error) {
		__antithesis_instrumentation__.Notify(265830)
		vea.Report(errors.WrapWithDepth(
			1, err, invalidSchemaChangerStatePrefix,
		))
	}
	__antithesis_instrumentation__.Notify(265821)

	if scs.JobID == 0 {
		__antithesis_instrumentation__.Notify(265831)
		report(errors.New("empty job ID"))
	} else {
		__antithesis_instrumentation__.Notify(265832)
	}
	__antithesis_instrumentation__.Notify(265822)

	for i, t := range scs.Targets {
		__antithesis_instrumentation__.Notify(265833)
		if got := screl.GetDescID(t.Element()); got != d.GetID() {
			__antithesis_instrumentation__.Notify(265835)
			report(errors.Errorf("target %d corresponds to descriptor %d != %d",
				i, got, d.GetID()))
		} else {
			__antithesis_instrumentation__.Notify(265836)
		}
		__antithesis_instrumentation__.Notify(265834)

		switch t.TargetStatus {
		case scpb.Status_PUBLIC, scpb.Status_ABSENT, scpb.Status_TRANSIENT:
			__antithesis_instrumentation__.Notify(265837)

		default:
			__antithesis_instrumentation__.Notify(265838)
			report(errors.Errorf("target %d is targeting an invalid status %s",
				i, t.TargetStatus))
		}
	}

	{
		__antithesis_instrumentation__.Notify(265839)
		var haveProblem bool
		if nt, ntr := len(scs.Targets), len(scs.TargetRanks); nt != ntr {
			__antithesis_instrumentation__.Notify(265842)
			haveProblem = true
			report(errors.Errorf("number mismatch between Targets and TargetRanks: %d != %d",
				nt, ntr))
		} else {
			__antithesis_instrumentation__.Notify(265843)
		}
		__antithesis_instrumentation__.Notify(265840)
		if nt, ns := len(scs.Targets), len(scs.CurrentStatuses); nt != ns {
			__antithesis_instrumentation__.Notify(265844)
			haveProblem = true
			report(errors.Errorf("number mismatch between Targets and CurentStatuses: %d != %d",
				nt, ns))
		} else {
			__antithesis_instrumentation__.Notify(265845)
		}
		__antithesis_instrumentation__.Notify(265841)

		if haveProblem {
			__antithesis_instrumentation__.Notify(265846)
			return
		} else {
			__antithesis_instrumentation__.Notify(265847)
		}
	}
	__antithesis_instrumentation__.Notify(265823)

	ranksToTarget := map[uint32]*scpb.Target{}
	{
		__antithesis_instrumentation__.Notify(265848)
		var duplicates util.FastIntSet
		for i, r := range scs.TargetRanks {
			__antithesis_instrumentation__.Notify(265850)
			if _, exists := ranksToTarget[r]; exists {
				__antithesis_instrumentation__.Notify(265851)
				duplicates.Add(int(r))
			} else {
				__antithesis_instrumentation__.Notify(265852)
				ranksToTarget[r] = &scs.Targets[i]
			}
		}
		__antithesis_instrumentation__.Notify(265849)
		if duplicates.Len() > 0 {
			__antithesis_instrumentation__.Notify(265853)
			report(errors.Errorf("TargetRanks contains duplicate entries %v",
				redact.SafeString(fmt.Sprint(duplicates.Ordered()))))
		} else {
			__antithesis_instrumentation__.Notify(265854)
		}
	}
	__antithesis_instrumentation__.Notify(265824)

	if !sort.SliceIsSorted(scs.RelevantStatements, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(265855)
		return scs.RelevantStatements[i].StatementRank < scs.RelevantStatements[j].StatementRank
	}) {
		__antithesis_instrumentation__.Notify(265856)
		report(errors.New("RelevantStatements are not sorted"))
	} else {
		__antithesis_instrumentation__.Notify(265857)
	}
	__antithesis_instrumentation__.Notify(265825)

	statementsExpected := map[uint32]*util.FastIntSet{}
	for i := range scs.Targets {
		__antithesis_instrumentation__.Notify(265858)
		t := &scs.Targets[i]
		exp, ok := statementsExpected[t.Metadata.StatementID]
		if !ok {
			__antithesis_instrumentation__.Notify(265860)
			exp = &util.FastIntSet{}
			statementsExpected[t.Metadata.StatementID] = exp
		} else {
			__antithesis_instrumentation__.Notify(265861)
		}
		__antithesis_instrumentation__.Notify(265859)
		exp.Add(int(scs.TargetRanks[i]))
	}
	__antithesis_instrumentation__.Notify(265826)
	var statementRanks util.FastIntSet
	for _, s := range scs.RelevantStatements {
		__antithesis_instrumentation__.Notify(265862)
		statementRanks.Add(int(s.StatementRank))
		if _, ok := statementsExpected[s.StatementRank]; !ok {
			__antithesis_instrumentation__.Notify(265863)
			report(errors.Errorf("unexpected statement %d (%s)",
				s.StatementRank, s.Statement.Statement))
		} else {
			__antithesis_instrumentation__.Notify(265864)
		}
	}
	__antithesis_instrumentation__.Notify(265827)

	if statementRanks.Len() != len(scs.RelevantStatements) {
		__antithesis_instrumentation__.Notify(265865)
		report(errors.Errorf("duplicates exist in RelevantStatements"))
	} else {
		__antithesis_instrumentation__.Notify(265866)
	}

	{
		__antithesis_instrumentation__.Notify(265867)
		var expected util.FastIntSet
		stmts := statementRanks.Copy()
		for rank := range statementsExpected {
			__antithesis_instrumentation__.Notify(265870)
			expected.Add(int(rank))
		}
		__antithesis_instrumentation__.Notify(265868)
		stmts.ForEach(func(i int) { __antithesis_instrumentation__.Notify(265871); expected.Remove(i) })
		__antithesis_instrumentation__.Notify(265869)

		expected.ForEach(func(stmtRank int) {
			__antithesis_instrumentation__.Notify(265872)
			expectedTargetRanks := statementsExpected[uint32(stmtRank)]
			var ranks, elementStrs []string
			expectedTargetRanks.ForEach(func(targetRank int) {
				__antithesis_instrumentation__.Notify(265874)
				ranks = append(ranks, fmt.Sprint(targetRank))
				elementStrs = append(elementStrs,
					screl.ElementString(ranksToTarget[uint32(targetRank)].Element()))
			})
			__antithesis_instrumentation__.Notify(265873)
			report(errors.Errorf("missing statement for targets (%s) / (%s)",
				redact.SafeString(strings.Join(ranks, ", ")),
				strings.Join(elementStrs, ", "),
			))
		})
	}

}
