package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/persistedsqlstats/sqlstatsutil"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func getTimeFromSeconds(seconds int64) *time.Time {
	__antithesis_instrumentation__.Notify(189722)
	if seconds != 0 {
		__antithesis_instrumentation__.Notify(189724)
		t := timeutil.Unix(seconds, 0)
		return &t
	} else {
		__antithesis_instrumentation__.Notify(189725)
	}
	__antithesis_instrumentation__.Notify(189723)
	return nil
}

func (s *statusServer) CombinedStatementStats(
	ctx context.Context, req *serverpb.CombinedStatementsStatsRequest,
) (*serverpb.StatementsResponse, error) {
	__antithesis_instrumentation__.Notify(189726)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if err := s.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(189728)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(189729)
	}
	__antithesis_instrumentation__.Notify(189727)

	return getCombinedStatementStats(
		ctx,
		req,
		s.sqlServer.pgServer.SQLServer.GetSQLStatsProvider(),
		s.internalExecutor,
		s.st,
		s.sqlServer.execCfg.SQLStatsTestingKnobs)
}

func getCombinedStatementStats(
	ctx context.Context,
	req *serverpb.CombinedStatementsStatsRequest,
	statsProvider sqlstats.Provider,
	ie *sql.InternalExecutor,
	settings *cluster.Settings,
	testingKnobs *sqlstats.TestingKnobs,
) (*serverpb.StatementsResponse, error) {
	__antithesis_instrumentation__.Notify(189730)
	startTime := getTimeFromSeconds(req.Start)
	endTime := getTimeFromSeconds(req.End)
	limit := SQLStatsResponseMax.Get(&settings.SV)
	whereClause, orderAndLimit, args := getCombinedStatementsQueryClausesAndArgs(startTime, endTime, limit, testingKnobs)
	statements, err := collectCombinedStatements(ctx, ie, whereClause, args, orderAndLimit)
	if err != nil {
		__antithesis_instrumentation__.Notify(189733)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189734)
	}
	__antithesis_instrumentation__.Notify(189731)

	transactions, err := collectCombinedTransactions(ctx, ie, whereClause, args, orderAndLimit)
	if err != nil {
		__antithesis_instrumentation__.Notify(189735)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189736)
	}
	__antithesis_instrumentation__.Notify(189732)

	response := &serverpb.StatementsResponse{
		Statements:            statements,
		Transactions:          transactions,
		LastReset:             statsProvider.GetLastReset(),
		InternalAppNamePrefix: catconstants.InternalAppNamePrefix,
	}

	return response, nil
}

func getCombinedStatementsQueryClausesAndArgs(
	start, end *time.Time, limit int64, testingKnobs *sqlstats.TestingKnobs,
) (whereClause string, orderAndLimitClause string, args []interface{}) {
	__antithesis_instrumentation__.Notify(189737)
	var buffer strings.Builder
	buffer.WriteString(testingKnobs.GetAOSTClause())

	buffer.WriteString(fmt.Sprintf(" WHERE app_name NOT LIKE '%s%%'", catconstants.InternalAppNamePrefix))

	if start != nil {
		__antithesis_instrumentation__.Notify(189740)
		buffer.WriteString(" AND aggregated_ts >= $1")
		args = append(args, *start)
	} else {
		__antithesis_instrumentation__.Notify(189741)
	}
	__antithesis_instrumentation__.Notify(189738)

	if end != nil {
		__antithesis_instrumentation__.Notify(189742)
		args = append(args, *end)
		buffer.WriteString(fmt.Sprintf(" AND aggregated_ts <= $%d", len(args)))
	} else {
		__antithesis_instrumentation__.Notify(189743)
	}
	__antithesis_instrumentation__.Notify(189739)
	args = append(args, limit)
	orderAndLimitClause = fmt.Sprintf(` ORDER BY aggregated_ts DESC LIMIT $%d`, len(args))

	return buffer.String(), orderAndLimitClause, args
}

func collectCombinedStatements(
	ctx context.Context,
	ie *sql.InternalExecutor,
	whereClause string,
	args []interface{},
	orderAndLimit string,
) ([]serverpb.StatementsResponse_CollectedStatementStatistics, error) {
	__antithesis_instrumentation__.Notify(189744)

	query := fmt.Sprintf(
		`SELECT
				fingerprint_id,
				transaction_fingerprint_id,
				app_name,
				max(aggregated_ts) as aggregated_ts,
				metadata,
				crdb_internal.merge_statement_stats(array_agg(statistics)) AS statistics,
				max(sampled_plan) AS sampled_plan,
				aggregation_interval
		FROM crdb_internal.statement_statistics %s
		GROUP BY
				fingerprint_id,
				transaction_fingerprint_id,
				app_name,
				metadata,
				aggregation_interval
		%s`, whereClause, orderAndLimit)

	const expectedNumDatums = 8

	it, err := ie.QueryIteratorEx(ctx, "combined-stmts-by-interval", nil,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		}, query, args...)

	if err != nil {
		__antithesis_instrumentation__.Notify(189749)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189750)
	}
	__antithesis_instrumentation__.Notify(189745)

	defer func() {
		__antithesis_instrumentation__.Notify(189751)
		closeErr := it.Close()
		if closeErr != nil {
			__antithesis_instrumentation__.Notify(189752)
			err = errors.CombineErrors(err, closeErr)
		} else {
			__antithesis_instrumentation__.Notify(189753)
		}
	}()
	__antithesis_instrumentation__.Notify(189746)

	var statements []serverpb.StatementsResponse_CollectedStatementStatistics
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(189754)
		var row tree.Datums
		if row = it.Cur(); row == nil {
			__antithesis_instrumentation__.Notify(189762)
			return nil, errors.New("unexpected null row")
		} else {
			__antithesis_instrumentation__.Notify(189763)
		}
		__antithesis_instrumentation__.Notify(189755)

		if row.Len() != expectedNumDatums {
			__antithesis_instrumentation__.Notify(189764)
			return nil, errors.Newf("expected %d columns, received %d", expectedNumDatums)
		} else {
			__antithesis_instrumentation__.Notify(189765)
		}
		__antithesis_instrumentation__.Notify(189756)

		var statementFingerprintID uint64
		if statementFingerprintID, err = sqlstatsutil.DatumToUint64(row[0]); err != nil {
			__antithesis_instrumentation__.Notify(189766)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189767)
		}
		__antithesis_instrumentation__.Notify(189757)

		var transactionFingerprintID uint64
		if transactionFingerprintID, err = sqlstatsutil.DatumToUint64(row[1]); err != nil {
			__antithesis_instrumentation__.Notify(189768)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189769)
		}
		__antithesis_instrumentation__.Notify(189758)

		app := string(tree.MustBeDString(row[2]))
		aggregatedTs := tree.MustBeDTimestampTZ(row[3]).Time

		var metadata roachpb.CollectedStatementStatistics
		metadataJSON := tree.MustBeDJSON(row[4]).JSON
		if err = sqlstatsutil.DecodeStmtStatsMetadataJSON(metadataJSON, &metadata); err != nil {
			__antithesis_instrumentation__.Notify(189770)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189771)
		}
		__antithesis_instrumentation__.Notify(189759)

		metadata.Key.App = app
		metadata.Key.TransactionFingerprintID =
			roachpb.TransactionFingerprintID(transactionFingerprintID)

		statsJSON := tree.MustBeDJSON(row[5]).JSON
		if err = sqlstatsutil.DecodeStmtStatsStatisticsJSON(statsJSON, &metadata.Stats); err != nil {
			__antithesis_instrumentation__.Notify(189772)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189773)
		}
		__antithesis_instrumentation__.Notify(189760)

		planJSON := tree.MustBeDJSON(row[6]).JSON
		plan, err := sqlstatsutil.JSONToExplainTreePlanNode(planJSON)
		if err != nil {
			__antithesis_instrumentation__.Notify(189774)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189775)
		}
		__antithesis_instrumentation__.Notify(189761)
		metadata.Stats.SensitiveInfo.MostRecentPlanDescription = *plan

		aggInterval := tree.MustBeDInterval(row[7]).Duration

		stmt := serverpb.StatementsResponse_CollectedStatementStatistics{
			Key: serverpb.StatementsResponse_ExtendedStatementStatisticsKey{
				KeyData:             metadata.Key,
				AggregatedTs:        aggregatedTs,
				AggregationInterval: time.Duration(aggInterval.Nanos()),
			},
			ID:    roachpb.StmtFingerprintID(statementFingerprintID),
			Stats: metadata.Stats,
		}

		statements = append(statements, stmt)

	}
	__antithesis_instrumentation__.Notify(189747)

	if err != nil {
		__antithesis_instrumentation__.Notify(189776)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189777)
	}
	__antithesis_instrumentation__.Notify(189748)

	return statements, nil
}

func collectCombinedTransactions(
	ctx context.Context,
	ie *sql.InternalExecutor,
	whereClause string,
	args []interface{},
	orderAndLimit string,
) ([]serverpb.StatementsResponse_ExtendedCollectedTransactionStatistics, error) {
	__antithesis_instrumentation__.Notify(189778)

	query := fmt.Sprintf(
		`SELECT
				app_name,
				max(aggregated_ts) as aggregated_ts,
				fingerprint_id,
				metadata,
				crdb_internal.merge_transaction_stats(array_agg(statistics)) AS statistics,
				aggregation_interval
			FROM crdb_internal.transaction_statistics %s
			GROUP BY
				app_name,
				fingerprint_id,
				metadata,
				aggregation_interval
			%s`, whereClause, orderAndLimit)

	const expectedNumDatums = 6

	it, err := ie.QueryIteratorEx(ctx, "combined-txns-by-interval", nil,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		}, query, args...)

	if err != nil {
		__antithesis_instrumentation__.Notify(189783)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189784)
	}
	__antithesis_instrumentation__.Notify(189779)

	defer func() {
		__antithesis_instrumentation__.Notify(189785)
		closeErr := it.Close()
		if closeErr != nil {
			__antithesis_instrumentation__.Notify(189786)
			err = errors.CombineErrors(err, closeErr)
		} else {
			__antithesis_instrumentation__.Notify(189787)
		}
	}()
	__antithesis_instrumentation__.Notify(189780)

	var transactions []serverpb.StatementsResponse_ExtendedCollectedTransactionStatistics
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(189788)
		var row tree.Datums
		if row = it.Cur(); row == nil {
			__antithesis_instrumentation__.Notify(189794)
			return nil, errors.New("unexpected null row")
		} else {
			__antithesis_instrumentation__.Notify(189795)
		}
		__antithesis_instrumentation__.Notify(189789)

		if row.Len() != expectedNumDatums {
			__antithesis_instrumentation__.Notify(189796)
			return nil, errors.Newf("expected %d columns, received %d", expectedNumDatums, row.Len())
		} else {
			__antithesis_instrumentation__.Notify(189797)
		}
		__antithesis_instrumentation__.Notify(189790)

		app := string(tree.MustBeDString(row[0]))
		aggregatedTs := tree.MustBeDTimestampTZ(row[1]).Time
		fingerprintID, err := sqlstatsutil.DatumToUint64(row[2])
		if err != nil {
			__antithesis_instrumentation__.Notify(189798)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189799)
		}
		__antithesis_instrumentation__.Notify(189791)

		var metadata roachpb.CollectedTransactionStatistics
		metadataJSON := tree.MustBeDJSON(row[3]).JSON
		if err = sqlstatsutil.DecodeTxnStatsMetadataJSON(metadataJSON, &metadata); err != nil {
			__antithesis_instrumentation__.Notify(189800)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189801)
		}
		__antithesis_instrumentation__.Notify(189792)

		statsJSON := tree.MustBeDJSON(row[4]).JSON
		if err = sqlstatsutil.DecodeTxnStatsStatisticsJSON(statsJSON, &metadata.Stats); err != nil {
			__antithesis_instrumentation__.Notify(189802)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189803)
		}
		__antithesis_instrumentation__.Notify(189793)

		aggInterval := tree.MustBeDInterval(row[5]).Duration

		txnStats := serverpb.StatementsResponse_ExtendedCollectedTransactionStatistics{
			StatsData: roachpb.CollectedTransactionStatistics{
				StatementFingerprintIDs:  metadata.StatementFingerprintIDs,
				App:                      app,
				Stats:                    metadata.Stats,
				AggregatedTs:             aggregatedTs,
				AggregationInterval:      time.Duration(aggInterval.Nanos()),
				TransactionFingerprintID: roachpb.TransactionFingerprintID(fingerprintID),
			},
		}

		transactions = append(transactions, txnStats)
	}
	__antithesis_instrumentation__.Notify(189781)

	if err != nil {
		__antithesis_instrumentation__.Notify(189804)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189805)
	}
	__antithesis_instrumentation__.Notify(189782)

	return transactions, nil
}

func (s *statusServer) StatementDetails(
	ctx context.Context, req *serverpb.StatementDetailsRequest,
) (*serverpb.StatementDetailsResponse, error) {
	__antithesis_instrumentation__.Notify(189806)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if err := s.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(189808)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(189809)
	}
	__antithesis_instrumentation__.Notify(189807)

	return getStatementDetails(
		ctx,
		req,
		s.internalExecutor,
		s.st,
		s.sqlServer.execCfg.SQLStatsTestingKnobs)
}

func getStatementDetails(
	ctx context.Context,
	req *serverpb.StatementDetailsRequest,
	ie *sql.InternalExecutor,
	settings *cluster.Settings,
	testingKnobs *sqlstats.TestingKnobs,
) (*serverpb.StatementDetailsResponse, error) {
	__antithesis_instrumentation__.Notify(189810)
	limit := SQLStatsResponseMax.Get(&settings.SV)
	whereClause, args, err := getStatementDetailsQueryClausesAndArgs(req, testingKnobs)
	if err != nil {
		__antithesis_instrumentation__.Notify(189816)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189817)
	}
	__antithesis_instrumentation__.Notify(189811)

	statementTotal, err := getTotalStatementDetails(ctx, ie, whereClause, args)
	if err != nil {
		__antithesis_instrumentation__.Notify(189818)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189819)
	}
	__antithesis_instrumentation__.Notify(189812)
	statementStatisticsPerAggregatedTs, err := getStatementDetailsPerAggregatedTs(ctx, ie, whereClause, args, limit)
	if err != nil {
		__antithesis_instrumentation__.Notify(189820)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189821)
	}
	__antithesis_instrumentation__.Notify(189813)
	statementStatisticsPerPlanHash, err := getStatementDetailsPerPlanHash(ctx, ie, whereClause, args, limit)
	if err != nil {
		__antithesis_instrumentation__.Notify(189822)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189823)
	}
	__antithesis_instrumentation__.Notify(189814)

	statementTotal.Metadata.DistSQLCount = 0
	statementTotal.Metadata.FailedCount = 0
	statementTotal.Metadata.FullScanCount = 0
	statementTotal.Metadata.VecCount = 0
	statementTotal.Metadata.TotalCount = 0
	for _, planStats := range statementStatisticsPerPlanHash {
		__antithesis_instrumentation__.Notify(189824)
		statementTotal.Metadata.DistSQLCount += planStats.Metadata.DistSQLCount
		statementTotal.Metadata.FailedCount += planStats.Metadata.FailedCount
		statementTotal.Metadata.FullScanCount += planStats.Metadata.FullScanCount
		statementTotal.Metadata.VecCount += planStats.Metadata.VecCount
		statementTotal.Metadata.TotalCount += planStats.Metadata.TotalCount
	}
	__antithesis_instrumentation__.Notify(189815)

	response := &serverpb.StatementDetailsResponse{
		Statement:                          statementTotal,
		StatementStatisticsPerAggregatedTs: statementStatisticsPerAggregatedTs,
		StatementStatisticsPerPlanHash:     statementStatisticsPerPlanHash,
		InternalAppNamePrefix:              catconstants.InternalAppNamePrefix,
	}

	return response, nil
}

func getStatementDetailsQueryClausesAndArgs(
	req *serverpb.StatementDetailsRequest, testingKnobs *sqlstats.TestingKnobs,
) (whereClause string, args []interface{}, err error) {
	__antithesis_instrumentation__.Notify(189825)
	var buffer strings.Builder
	buffer.WriteString(testingKnobs.GetAOSTClause())

	fingerprintID, err := strconv.ParseUint(req.FingerprintId, 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(189830)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(189831)
	}
	__antithesis_instrumentation__.Notify(189826)
	args = append(args, strconv.FormatUint(fingerprintID, 16))
	buffer.WriteString(fmt.Sprintf(" WHERE encode(fingerprint_id, 'hex') = $%d", len(args)))

	buffer.WriteString(fmt.Sprintf(" AND app_name NOT LIKE '%s%%'", catconstants.InternalAppNamePrefix))

	if len(req.AppNames) > 0 {
		__antithesis_instrumentation__.Notify(189832)
		if !(len(req.AppNames) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(189833)
			return req.AppNames[0] == "" == true
		}() == true) {
			__antithesis_instrumentation__.Notify(189834)
			buffer.WriteString(" AND (")
			for i, app := range req.AppNames {
				__antithesis_instrumentation__.Notify(189836)
				if i != 0 {
					__antithesis_instrumentation__.Notify(189837)
					args = append(args, app)
					buffer.WriteString(fmt.Sprintf(" OR app_name = $%d", len(args)))
				} else {
					__antithesis_instrumentation__.Notify(189838)
					args = append(args, app)
					buffer.WriteString(fmt.Sprintf(" app_name = $%d", len(args)))
				}
			}
			__antithesis_instrumentation__.Notify(189835)
			buffer.WriteString(" )")
		} else {
			__antithesis_instrumentation__.Notify(189839)
		}
	} else {
		__antithesis_instrumentation__.Notify(189840)
	}
	__antithesis_instrumentation__.Notify(189827)

	start := getTimeFromSeconds(req.Start)
	if start != nil {
		__antithesis_instrumentation__.Notify(189841)
		args = append(args, *start)
		buffer.WriteString(fmt.Sprintf(" AND aggregated_ts >= $%d", len(args)))
	} else {
		__antithesis_instrumentation__.Notify(189842)
	}
	__antithesis_instrumentation__.Notify(189828)
	end := getTimeFromSeconds(req.End)
	if end != nil {
		__antithesis_instrumentation__.Notify(189843)
		args = append(args, *end)
		buffer.WriteString(fmt.Sprintf(" AND aggregated_ts <= $%d", len(args)))
	} else {
		__antithesis_instrumentation__.Notify(189844)
	}
	__antithesis_instrumentation__.Notify(189829)
	whereClause = buffer.String()

	return whereClause, args, nil
}

func getTotalStatementDetails(
	ctx context.Context, ie *sql.InternalExecutor, whereClause string, args []interface{},
) (serverpb.StatementDetailsResponse_CollectedStatementSummary, error) {
	__antithesis_instrumentation__.Notify(189845)
	query := fmt.Sprintf(
		`SELECT
				crdb_internal.merge_stats_metadata(array_agg(metadata)) AS metadata,
				aggregation_interval,
				array_agg(app_name) as app_names,
				crdb_internal.merge_statement_stats(array_agg(statistics)) AS statistics,
				max(sampled_plan) as sampled_plan
		FROM crdb_internal.statement_statistics %s
		GROUP BY
				aggregation_interval
		LIMIT 1`, whereClause)

	const expectedNumDatums = 5
	var statement serverpb.StatementDetailsResponse_CollectedStatementSummary

	row, err := ie.QueryRowEx(ctx, "combined-stmts-details-total", nil,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		}, query, args...)

	if err != nil {
		__antithesis_instrumentation__.Notify(189854)
		return statement, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189855)
	}
	__antithesis_instrumentation__.Notify(189846)
	if len(row) == 0 {
		__antithesis_instrumentation__.Notify(189856)
		return statement, nil
	} else {
		__antithesis_instrumentation__.Notify(189857)
	}
	__antithesis_instrumentation__.Notify(189847)
	if row.Len() != expectedNumDatums {
		__antithesis_instrumentation__.Notify(189858)
		return statement, serverError(ctx, errors.Newf("expected %d columns, received %d", expectedNumDatums))
	} else {
		__antithesis_instrumentation__.Notify(189859)
	}
	__antithesis_instrumentation__.Notify(189848)

	var statistics roachpb.CollectedStatementStatistics
	var aggregatedMetadata roachpb.AggregatedStatementMetadata
	metadataJSON := tree.MustBeDJSON(row[0]).JSON

	if err = sqlstatsutil.DecodeAggregatedMetadataJSON(metadataJSON, &aggregatedMetadata); err != nil {
		__antithesis_instrumentation__.Notify(189860)
		return statement, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189861)
	}
	__antithesis_instrumentation__.Notify(189849)

	aggInterval := tree.MustBeDInterval(row[1]).Duration

	apps := tree.MustBeDArray(row[2])
	var appNames []string
	for _, s := range apps.Array {
		__antithesis_instrumentation__.Notify(189862)
		appNames = util.CombineUniqueString(appNames, []string{string(tree.MustBeDString(s))})
	}
	__antithesis_instrumentation__.Notify(189850)
	aggregatedMetadata.AppNames = appNames

	statsJSON := tree.MustBeDJSON(row[3]).JSON
	if err = sqlstatsutil.DecodeStmtStatsStatisticsJSON(statsJSON, &statistics.Stats); err != nil {
		__antithesis_instrumentation__.Notify(189863)
		return statement, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189864)
	}
	__antithesis_instrumentation__.Notify(189851)

	planJSON := tree.MustBeDJSON(row[4]).JSON
	plan, err := sqlstatsutil.JSONToExplainTreePlanNode(planJSON)
	if err != nil {
		__antithesis_instrumentation__.Notify(189865)
		return statement, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189866)
	}
	__antithesis_instrumentation__.Notify(189852)
	statistics.Stats.SensitiveInfo.MostRecentPlanDescription = *plan

	args = []interface{}{}
	args = append(args, aggregatedMetadata.Query)
	query = fmt.Sprintf(
		`SELECT prettify_statement($1, %d, %d, %d)`,
		tree.ConsoleLineWidth, tree.PrettyAlignAndDeindent, tree.UpperCase)
	row, err = ie.QueryRowEx(ctx, "combined-stmts-details-format-query", nil,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		}, query, args...)

	if err != nil {
		__antithesis_instrumentation__.Notify(189867)
		return statement, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189868)
	}
	__antithesis_instrumentation__.Notify(189853)
	aggregatedMetadata.FormattedQuery = string(tree.MustBeDString(row[0]))

	statement = serverpb.StatementDetailsResponse_CollectedStatementSummary{
		Metadata:            aggregatedMetadata,
		AggregationInterval: time.Duration(aggInterval.Nanos()),
		Stats:               statistics.Stats,
	}

	return statement, nil
}

func getStatementDetailsPerAggregatedTs(
	ctx context.Context,
	ie *sql.InternalExecutor,
	whereClause string,
	args []interface{},
	limit int64,
) ([]serverpb.StatementDetailsResponse_CollectedStatementGroupedByAggregatedTs, error) {
	__antithesis_instrumentation__.Notify(189869)
	query := fmt.Sprintf(
		`SELECT
				aggregated_ts,
				crdb_internal.merge_stats_metadata(array_agg(metadata)) AS metadata,
				crdb_internal.merge_statement_stats(array_agg(statistics)) AS statistics,
				max(sampled_plan) as sampled_plan,
				aggregation_interval
		FROM crdb_internal.statement_statistics %s
		GROUP BY
				aggregated_ts,
				aggregation_interval
		LIMIT $%d`, whereClause, len(args)+1)

	args = append(args, limit)
	const expectedNumDatums = 5

	it, err := ie.QueryIteratorEx(ctx, "combined-stmts-details-by-aggregated-timestamp", nil,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		}, query, args...)

	if err != nil {
		__antithesis_instrumentation__.Notify(189874)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189875)
	}
	__antithesis_instrumentation__.Notify(189870)

	defer func() {
		__antithesis_instrumentation__.Notify(189876)
		closeErr := it.Close()
		if closeErr != nil {
			__antithesis_instrumentation__.Notify(189877)
			err = errors.CombineErrors(err, closeErr)
		} else {
			__antithesis_instrumentation__.Notify(189878)
		}
	}()
	__antithesis_instrumentation__.Notify(189871)

	var statements []serverpb.StatementDetailsResponse_CollectedStatementGroupedByAggregatedTs
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(189879)
		var row tree.Datums
		if row = it.Cur(); row == nil {
			__antithesis_instrumentation__.Notify(189885)
			return nil, errors.New("unexpected null row")
		} else {
			__antithesis_instrumentation__.Notify(189886)
		}
		__antithesis_instrumentation__.Notify(189880)

		if row.Len() != expectedNumDatums {
			__antithesis_instrumentation__.Notify(189887)
			return nil, errors.Newf("expected %d columns, received %d", expectedNumDatums)
		} else {
			__antithesis_instrumentation__.Notify(189888)
		}
		__antithesis_instrumentation__.Notify(189881)

		aggregatedTs := tree.MustBeDTimestampTZ(row[0]).Time

		var metadata roachpb.CollectedStatementStatistics
		var aggregatedMetadata roachpb.AggregatedStatementMetadata
		metadataJSON := tree.MustBeDJSON(row[1]).JSON
		if err = sqlstatsutil.DecodeAggregatedMetadataJSON(metadataJSON, &aggregatedMetadata); err != nil {
			__antithesis_instrumentation__.Notify(189889)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189890)
		}
		__antithesis_instrumentation__.Notify(189882)

		statsJSON := tree.MustBeDJSON(row[2]).JSON
		if err = sqlstatsutil.DecodeStmtStatsStatisticsJSON(statsJSON, &metadata.Stats); err != nil {
			__antithesis_instrumentation__.Notify(189891)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189892)
		}
		__antithesis_instrumentation__.Notify(189883)

		planJSON := tree.MustBeDJSON(row[3]).JSON
		plan, err := sqlstatsutil.JSONToExplainTreePlanNode(planJSON)
		if err != nil {
			__antithesis_instrumentation__.Notify(189893)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189894)
		}
		__antithesis_instrumentation__.Notify(189884)
		metadata.Stats.SensitiveInfo.MostRecentPlanDescription = *plan

		aggInterval := tree.MustBeDInterval(row[4]).Duration

		stmt := serverpb.StatementDetailsResponse_CollectedStatementGroupedByAggregatedTs{
			AggregatedTs:        aggregatedTs,
			AggregationInterval: time.Duration(aggInterval.Nanos()),
			Stats:               metadata.Stats,
			Metadata:            aggregatedMetadata,
		}

		statements = append(statements, stmt)
	}
	__antithesis_instrumentation__.Notify(189872)
	if err != nil {
		__antithesis_instrumentation__.Notify(189895)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189896)
	}
	__antithesis_instrumentation__.Notify(189873)

	return statements, nil
}

func getExplainPlanFromGist(ctx context.Context, ie *sql.InternalExecutor, planGist string) string {
	__antithesis_instrumentation__.Notify(189897)
	planError := "Error collecting Explain Plan."
	var args []interface{}

	query := `SELECT crdb_internal.decode_plan_gist($1)`
	args = append(args, planGist)

	it, err := ie.QueryIteratorEx(ctx, "combined-stmts-details-get-explain-plan", nil,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		}, query, args...)

	if err != nil {
		__antithesis_instrumentation__.Notify(189901)
		return planError
	} else {
		__antithesis_instrumentation__.Notify(189902)
	}
	__antithesis_instrumentation__.Notify(189898)

	var explainPlan []string
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(189903)
		var row tree.Datums
		if row = it.Cur(); row == nil {
			__antithesis_instrumentation__.Notify(189905)
			return planError
		} else {
			__antithesis_instrumentation__.Notify(189906)
		}
		__antithesis_instrumentation__.Notify(189904)
		explainPlanLine := string(tree.MustBeDString(row[0]))
		explainPlan = append(explainPlan, explainPlanLine)
	}
	__antithesis_instrumentation__.Notify(189899)
	if err != nil {
		__antithesis_instrumentation__.Notify(189907)
		return planError
	} else {
		__antithesis_instrumentation__.Notify(189908)
	}
	__antithesis_instrumentation__.Notify(189900)

	return strings.Join(explainPlan, "\n")
}

func getStatementDetailsPerPlanHash(
	ctx context.Context,
	ie *sql.InternalExecutor,
	whereClause string,
	args []interface{},
	limit int64,
) ([]serverpb.StatementDetailsResponse_CollectedStatementGroupedByPlanHash, error) {
	__antithesis_instrumentation__.Notify(189909)
	query := fmt.Sprintf(
		`SELECT
				plan_hash,
				(statistics -> 'statistics' -> 'planGists'->>0) as plan_gist,
				crdb_internal.merge_stats_metadata(array_agg(metadata)) AS metadata,
				crdb_internal.merge_statement_stats(array_agg(statistics)) AS statistics,
				max(sampled_plan) as sampled_plan,
				aggregation_interval
		FROM crdb_internal.statement_statistics %s
		GROUP BY
				plan_hash,
				plan_gist,
				aggregation_interval
		LIMIT $%d`, whereClause, len(args)+1)

	args = append(args, limit)
	const expectedNumDatums = 6

	it, err := ie.QueryIteratorEx(ctx, "combined-stmts-details-by-plan-hash", nil,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		}, query, args...)

	if err != nil {
		__antithesis_instrumentation__.Notify(189914)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189915)
	}
	__antithesis_instrumentation__.Notify(189910)

	defer func() {
		__antithesis_instrumentation__.Notify(189916)
		closeErr := it.Close()
		if closeErr != nil {
			__antithesis_instrumentation__.Notify(189917)
			err = errors.CombineErrors(err, closeErr)
		} else {
			__antithesis_instrumentation__.Notify(189918)
		}
	}()
	__antithesis_instrumentation__.Notify(189911)

	var statements []serverpb.StatementDetailsResponse_CollectedStatementGroupedByPlanHash
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(189919)
		var row tree.Datums
		if row = it.Cur(); row == nil {
			__antithesis_instrumentation__.Notify(189930)
			return nil, errors.New("unexpected null row")
		} else {
			__antithesis_instrumentation__.Notify(189931)
		}
		__antithesis_instrumentation__.Notify(189920)

		if row.Len() != expectedNumDatums {
			__antithesis_instrumentation__.Notify(189932)
			return nil, errors.Newf("expected %d columns, received %d", expectedNumDatums)
		} else {
			__antithesis_instrumentation__.Notify(189933)
		}
		__antithesis_instrumentation__.Notify(189921)

		var planHash uint64
		if planHash, err = sqlstatsutil.DatumToUint64(row[0]); err != nil {
			__antithesis_instrumentation__.Notify(189934)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189935)
		}
		__antithesis_instrumentation__.Notify(189922)
		planGist := string(tree.MustBeDString(row[1]))
		explainPlan := getExplainPlanFromGist(ctx, ie, planGist)

		var metadata roachpb.CollectedStatementStatistics
		var aggregatedMetadata roachpb.AggregatedStatementMetadata
		metadataJSON := tree.MustBeDJSON(row[2]).JSON
		if err = sqlstatsutil.DecodeAggregatedMetadataJSON(metadataJSON, &aggregatedMetadata); err != nil {
			__antithesis_instrumentation__.Notify(189936)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189937)
		}
		__antithesis_instrumentation__.Notify(189923)

		statsJSON := tree.MustBeDJSON(row[3]).JSON
		if err = sqlstatsutil.DecodeStmtStatsStatisticsJSON(statsJSON, &metadata.Stats); err != nil {
			__antithesis_instrumentation__.Notify(189938)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189939)
		}
		__antithesis_instrumentation__.Notify(189924)

		planJSON := tree.MustBeDJSON(row[4]).JSON
		plan, err := sqlstatsutil.JSONToExplainTreePlanNode(planJSON)
		if err != nil {
			__antithesis_instrumentation__.Notify(189940)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(189941)
		}
		__antithesis_instrumentation__.Notify(189925)
		metadata.Stats.SensitiveInfo.MostRecentPlanDescription = *plan
		aggInterval := tree.MustBeDInterval(row[5]).Duration

		if aggregatedMetadata.DistSQLCount > 0 {
			__antithesis_instrumentation__.Notify(189942)
			aggregatedMetadata.DistSQLCount = metadata.Stats.Count
		} else {
			__antithesis_instrumentation__.Notify(189943)
		}
		__antithesis_instrumentation__.Notify(189926)
		if aggregatedMetadata.FailedCount > 0 {
			__antithesis_instrumentation__.Notify(189944)
			aggregatedMetadata.FailedCount = metadata.Stats.Count
		} else {
			__antithesis_instrumentation__.Notify(189945)
		}
		__antithesis_instrumentation__.Notify(189927)
		if aggregatedMetadata.FullScanCount > 0 {
			__antithesis_instrumentation__.Notify(189946)
			aggregatedMetadata.FullScanCount = metadata.Stats.Count
		} else {
			__antithesis_instrumentation__.Notify(189947)
		}
		__antithesis_instrumentation__.Notify(189928)
		if aggregatedMetadata.VecCount > 0 {
			__antithesis_instrumentation__.Notify(189948)
			aggregatedMetadata.VecCount = metadata.Stats.Count
		} else {
			__antithesis_instrumentation__.Notify(189949)
		}
		__antithesis_instrumentation__.Notify(189929)
		aggregatedMetadata.TotalCount = metadata.Stats.Count

		stmt := serverpb.StatementDetailsResponse_CollectedStatementGroupedByPlanHash{
			AggregationInterval: time.Duration(aggInterval.Nanos()),
			ExplainPlan:         explainPlan,
			PlanHash:            planHash,
			Stats:               metadata.Stats,
			Metadata:            aggregatedMetadata,
		}

		statements = append(statements, stmt)
	}
	__antithesis_instrumentation__.Notify(189912)
	if err != nil {
		__antithesis_instrumentation__.Notify(189950)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189951)
	}
	__antithesis_instrumentation__.Notify(189913)

	return statements, nil
}
