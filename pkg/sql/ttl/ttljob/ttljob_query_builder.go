package ttljob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/errors"
)

type selectQueryBuilder struct {
	tableID         descpb.ID
	pkColumns       []string
	selectOpName    string
	startPK, endPK  tree.Datums
	selectBatchSize int
	aost            tree.DTimestampTZ

	isFirst bool

	cachedQuery string

	cachedArgs []interface{}

	pkColumnNamesSQL string

	endPKColumnNamesSQL string
}

func makeSelectQueryBuilder(
	tableID descpb.ID,
	cutoff time.Time,
	pkColumns []string,
	relationName string,
	startPK, endPK tree.Datums,
	aost tree.DTimestampTZ,
	selectBatchSize int,
) selectQueryBuilder {
	__antithesis_instrumentation__.Notify(628784)

	cachedArgs := make([]interface{}, 0, 1+len(pkColumns)*2)
	cachedArgs = append(cachedArgs, cutoff)
	for _, d := range endPK {
		__antithesis_instrumentation__.Notify(628787)
		cachedArgs = append(cachedArgs, d)
	}
	__antithesis_instrumentation__.Notify(628785)
	for _, d := range startPK {
		__antithesis_instrumentation__.Notify(628788)
		cachedArgs = append(cachedArgs, d)
	}
	__antithesis_instrumentation__.Notify(628786)

	return selectQueryBuilder{
		tableID:         tableID,
		pkColumns:       pkColumns,
		selectOpName:    fmt.Sprintf("ttl select %s", relationName),
		startPK:         startPK,
		endPK:           endPK,
		aost:            aost,
		selectBatchSize: selectBatchSize,

		cachedArgs:          cachedArgs,
		isFirst:             true,
		pkColumnNamesSQL:    makeColumnNamesSQL(pkColumns),
		endPKColumnNamesSQL: makeColumnNamesSQL(pkColumns[:len(endPK)]),
	}
}

func (b *selectQueryBuilder) buildQuery() string {
	__antithesis_instrumentation__.Notify(628789)

	var endFilterClause string
	if len(b.endPK) > 0 {
		__antithesis_instrumentation__.Notify(628792)
		endFilterClause = fmt.Sprintf(" AND (%s) < (", b.endPKColumnNamesSQL)
		for i := range b.endPK {
			__antithesis_instrumentation__.Notify(628794)
			if i > 0 {
				__antithesis_instrumentation__.Notify(628796)
				endFilterClause += ", "
			} else {
				__antithesis_instrumentation__.Notify(628797)
			}
			__antithesis_instrumentation__.Notify(628795)
			endFilterClause += fmt.Sprintf("$%d", i+2)
		}
		__antithesis_instrumentation__.Notify(628793)
		endFilterClause += ")"
	} else {
		__antithesis_instrumentation__.Notify(628798)
	}
	__antithesis_instrumentation__.Notify(628790)

	var filterClause string
	if !b.isFirst {
		__antithesis_instrumentation__.Notify(628799)

		filterClause = fmt.Sprintf(" AND (%s) > (", b.pkColumnNamesSQL)
		for i := range b.pkColumns {
			__antithesis_instrumentation__.Notify(628801)
			if i > 0 {
				__antithesis_instrumentation__.Notify(628803)
				filterClause += ", "
			} else {
				__antithesis_instrumentation__.Notify(628804)
			}
			__antithesis_instrumentation__.Notify(628802)

			filterClause += fmt.Sprintf("$%d", 2+len(b.endPK)+i)
		}
		__antithesis_instrumentation__.Notify(628800)
		filterClause += ")"
	} else {
		__antithesis_instrumentation__.Notify(628805)
		if len(b.startPK) > 0 {
			__antithesis_instrumentation__.Notify(628806)

			filterClause = fmt.Sprintf(" AND (%s) >= (", makeColumnNamesSQL(b.pkColumns[:len(b.startPK)]))
			for i := range b.startPK {
				__antithesis_instrumentation__.Notify(628808)
				if i > 0 {
					__antithesis_instrumentation__.Notify(628810)
					filterClause += ", "
				} else {
					__antithesis_instrumentation__.Notify(628811)
				}
				__antithesis_instrumentation__.Notify(628809)

				filterClause += fmt.Sprintf("$%d", 2+len(b.endPK)+i)
			}
			__antithesis_instrumentation__.Notify(628807)
			filterClause += ")"
		} else {
			__antithesis_instrumentation__.Notify(628812)
		}
	}
	__antithesis_instrumentation__.Notify(628791)

	return fmt.Sprintf(
		`SELECT %[1]s FROM [%[2]d AS tbl_name]
AS OF SYSTEM TIME %[3]s
WHERE crdb_internal_expiration <= $1%[4]s%[5]s
ORDER BY %[1]s
LIMIT %[6]d`,
		b.pkColumnNamesSQL,
		b.tableID,
		b.aost.String(),
		filterClause,
		endFilterClause,
		b.selectBatchSize,
	)
}

func (b *selectQueryBuilder) nextQuery() (string, []interface{}) {
	__antithesis_instrumentation__.Notify(628813)
	if b.isFirst {
		__antithesis_instrumentation__.Notify(628816)
		q := b.buildQuery()
		b.isFirst = false
		return q, b.cachedArgs
	} else {
		__antithesis_instrumentation__.Notify(628817)
	}
	__antithesis_instrumentation__.Notify(628814)

	if b.cachedQuery == "" {
		__antithesis_instrumentation__.Notify(628818)
		b.cachedQuery = b.buildQuery()
	} else {
		__antithesis_instrumentation__.Notify(628819)
	}
	__antithesis_instrumentation__.Notify(628815)
	return b.cachedQuery, b.cachedArgs
}

func (b *selectQueryBuilder) run(
	ctx context.Context, ie *sql.InternalExecutor,
) ([]tree.Datums, error) {
	__antithesis_instrumentation__.Notify(628820)
	q, args := b.nextQuery()

	qosLevel := sessiondatapb.TTLLow
	ret, err := ie.QueryBufferedEx(
		ctx,
		b.selectOpName,
		nil,
		sessiondata.InternalExecutorOverride{
			User:             security.RootUserName(),
			QualityOfService: &qosLevel,
		},
		q,
		args...,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(628823)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(628824)
	}
	__antithesis_instrumentation__.Notify(628821)
	if err := b.moveCursor(ret); err != nil {
		__antithesis_instrumentation__.Notify(628825)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(628826)
	}
	__antithesis_instrumentation__.Notify(628822)
	return ret, nil
}

func (b *selectQueryBuilder) moveCursor(rows []tree.Datums) error {
	__antithesis_instrumentation__.Notify(628827)

	if len(rows) > 0 {
		__antithesis_instrumentation__.Notify(628829)
		lastRow := rows[len(rows)-1]
		b.cachedArgs = b.cachedArgs[:1+len(b.endPK)]
		if len(lastRow) != len(b.pkColumns) {
			__antithesis_instrumentation__.Notify(628831)
			return errors.AssertionFailedf("expected %d columns for last row, got %d", len(b.pkColumns), len(lastRow))
		} else {
			__antithesis_instrumentation__.Notify(628832)
		}
		__antithesis_instrumentation__.Notify(628830)
		for _, d := range lastRow {
			__antithesis_instrumentation__.Notify(628833)
			b.cachedArgs = append(b.cachedArgs, d)
		}
	} else {
		__antithesis_instrumentation__.Notify(628834)
	}
	__antithesis_instrumentation__.Notify(628828)
	return nil
}

type deleteQueryBuilder struct {
	tableID         descpb.ID
	pkColumns       []string
	deleteBatchSize int
	deleteOpName    string

	cachedQuery string

	cachedArgs []interface{}
}

func makeDeleteQueryBuilder(
	tableID descpb.ID, cutoff time.Time, pkColumns []string, relationName string, deleteBatchSize int,
) deleteQueryBuilder {
	__antithesis_instrumentation__.Notify(628835)
	cachedArgs := make([]interface{}, 0, 1+len(pkColumns)*deleteBatchSize)
	cachedArgs = append(cachedArgs, cutoff)

	return deleteQueryBuilder{
		tableID:         tableID,
		pkColumns:       pkColumns,
		deleteBatchSize: deleteBatchSize,
		deleteOpName:    fmt.Sprintf("ttl delete %s", relationName),

		cachedArgs: cachedArgs,
	}
}

func (b *deleteQueryBuilder) buildQuery(numRows int) string {
	__antithesis_instrumentation__.Notify(628836)
	columnNamesSQL := makeColumnNamesSQL(b.pkColumns)
	var placeholderStr string
	for i := 0; i < numRows; i++ {
		__antithesis_instrumentation__.Notify(628838)
		if i > 0 {
			__antithesis_instrumentation__.Notify(628841)
			placeholderStr += ", "
		} else {
			__antithesis_instrumentation__.Notify(628842)
		}
		__antithesis_instrumentation__.Notify(628839)
		placeholderStr += "("
		for j := 0; j < len(b.pkColumns); j++ {
			__antithesis_instrumentation__.Notify(628843)
			if j > 0 {
				__antithesis_instrumentation__.Notify(628845)
				placeholderStr += ", "
			} else {
				__antithesis_instrumentation__.Notify(628846)
			}
			__antithesis_instrumentation__.Notify(628844)
			placeholderStr += fmt.Sprintf("$%d", 2+i*len(b.pkColumns)+j)
		}
		__antithesis_instrumentation__.Notify(628840)
		placeholderStr += ")"
	}
	__antithesis_instrumentation__.Notify(628837)

	return fmt.Sprintf(
		`DELETE FROM [%d AS tbl_name] WHERE crdb_internal_expiration <= $1 AND (%s) IN (%s)`,
		b.tableID,
		columnNamesSQL,
		placeholderStr,
	)
}

func (b *deleteQueryBuilder) buildQueryAndArgs(rows []tree.Datums) (string, []interface{}) {
	__antithesis_instrumentation__.Notify(628847)
	var q string
	if len(rows) == b.deleteBatchSize {
		__antithesis_instrumentation__.Notify(628850)
		if b.cachedQuery == "" {
			__antithesis_instrumentation__.Notify(628852)
			b.cachedQuery = b.buildQuery(len(rows))
		} else {
			__antithesis_instrumentation__.Notify(628853)
		}
		__antithesis_instrumentation__.Notify(628851)
		q = b.cachedQuery
	} else {
		__antithesis_instrumentation__.Notify(628854)
		q = b.buildQuery(len(rows))
	}
	__antithesis_instrumentation__.Notify(628848)
	deleteArgs := b.cachedArgs[:1]
	for _, row := range rows {
		__antithesis_instrumentation__.Notify(628855)
		for _, col := range row {
			__antithesis_instrumentation__.Notify(628856)
			deleteArgs = append(deleteArgs, col)
		}
	}
	__antithesis_instrumentation__.Notify(628849)
	return q, deleteArgs
}

func (b *deleteQueryBuilder) run(
	ctx context.Context, ie *sql.InternalExecutor, txn *kv.Txn, rows []tree.Datums,
) error {
	__antithesis_instrumentation__.Notify(628857)
	q, deleteArgs := b.buildQueryAndArgs(rows)
	qosLevel := sessiondatapb.TTLLow
	_, err := ie.ExecEx(
		ctx,
		b.deleteOpName,
		txn,
		sessiondata.InternalExecutorOverride{
			User:             security.RootUserName(),
			QualityOfService: &qosLevel,
		},
		q,
		deleteArgs...,
	)
	return err
}

func makeColumnNamesSQL(columns []string) string {
	__antithesis_instrumentation__.Notify(628858)
	var b bytes.Buffer
	for i, pkColumn := range columns {
		__antithesis_instrumentation__.Notify(628860)
		if i > 0 {
			__antithesis_instrumentation__.Notify(628862)
			b.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(628863)
		}
		__antithesis_instrumentation__.Notify(628861)
		lexbase.EncodeRestrictedSQLIdent(&b, pkColumn, lexbase.EncNoFlags)
	}
	__antithesis_instrumentation__.Notify(628859)
	return b.String()
}
