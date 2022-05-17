package testutilsccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltestutils"
	"github.com/cockroachdb/cockroach/pkg/sql/tests"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/stretchr/testify/require"
)

type AlterPrimaryKeyCorrectZoneConfigIntermediateZoneConfig struct {
	ShowConfigStatement string
	ExpectedTarget      string
	ExpectedSQL         string
}

type AlterPrimaryKeyCorrectZoneConfigTestCase struct {
	Desc                            string
	SetupQuery                      string
	AlterQuery                      string
	ExpectedIntermediateZoneConfigs []AlterPrimaryKeyCorrectZoneConfigIntermediateZoneConfig
}

func AlterPrimaryKeyCorrectZoneConfigTest(
	t *testing.T, createDBStatement string, testCases []AlterPrimaryKeyCorrectZoneConfigTestCase,
) {
	__antithesis_instrumentation__.Notify(27134)
	chunkSize := int64(100)
	maxValue := 4000

	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(27136)

		maxValue = 200
		chunkSize = 5
	} else {
		__antithesis_instrumentation__.Notify(27137)
	}
	__antithesis_instrumentation__.Notify(27135)

	ctx := context.Background()

	for _, tc := range testCases {
		__antithesis_instrumentation__.Notify(27138)
		t.Run(tc.Desc, func(t *testing.T) {
			__antithesis_instrumentation__.Notify(27139)
			var db *gosql.DB
			params, _ := tests.CreateTestServerParams()
			params.Locality.Tiers = []roachpb.Tier{
				{Key: "region", Value: "ajstorm-1"},
			}

			runCheck := false
			params.Knobs = base.TestingKnobs{
				SQLSchemaChanger: &sql.SchemaChangerTestingKnobs{
					BackfillChunkSize: chunkSize,
				},
				DistSQL: &execinfra.TestingKnobs{
					RunBeforeBackfillChunk: func(sp roachpb.Span) error {
						__antithesis_instrumentation__.Notify(27142)
						if runCheck {
							__antithesis_instrumentation__.Notify(27144)
							for _, subTC := range tc.ExpectedIntermediateZoneConfigs {
								__antithesis_instrumentation__.Notify(27146)
								t.Run(subTC.ShowConfigStatement, func(t *testing.T) {
									__antithesis_instrumentation__.Notify(27147)
									var target, sql string
									require.NoError(
										t,
										db.QueryRow(subTC.ShowConfigStatement).Scan(&target, &sql),
									)
									require.Equal(t, subTC.ExpectedTarget, target)
									require.Equal(t, subTC.ExpectedSQL, sql)
								})
							}
							__antithesis_instrumentation__.Notify(27145)
							runCheck = false
						} else {
							__antithesis_instrumentation__.Notify(27148)
						}
						__antithesis_instrumentation__.Notify(27143)
						return nil
					},
				},

				JobsTestingKnobs: jobs.NewTestingKnobsWithShortIntervals(),
			}
			__antithesis_instrumentation__.Notify(27140)
			s, sqlDB, _ := serverutils.StartServer(t, params)
			db = sqlDB
			defer s.Stopper().Stop(ctx)

			if _, err := sqlDB.Exec(fmt.Sprintf(`
%s;
USE t;
%s
`, createDBStatement, tc.SetupQuery)); err != nil {
				__antithesis_instrumentation__.Notify(27149)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(27150)
			}
			__antithesis_instrumentation__.Notify(27141)

			require.NoError(t, sqltestutils.BulkInsertIntoTable(sqlDB, maxValue))

			runCheck = true
			_, err := sqlDB.Exec(tc.AlterQuery)
			require.NoError(t, err)
		})
	}

}
