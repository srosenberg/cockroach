package sqlutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
)

type ScrubResult struct {
	ErrorType  string
	Database   string
	Table      string
	PrimaryKey string
	Timestamp  time.Time
	Repaired   bool
	Details    string
}

func GetScrubResultRows(rows *gosql.Rows) (results []ScrubResult, err error) {
	__antithesis_instrumentation__.Notify(646232)
	defer rows.Close()

	var unused *string
	for rows.Next() {
		__antithesis_instrumentation__.Notify(646235)
		result := ScrubResult{}
		if err := rows.Scan(

			&unused,
			&result.ErrorType,
			&result.Database,
			&result.Table,
			&result.PrimaryKey,
			&result.Timestamp,
			&result.Repaired,
			&result.Details,
		); err != nil {
			__antithesis_instrumentation__.Notify(646237)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(646238)
		}
		__antithesis_instrumentation__.Notify(646236)
		results = append(results, result)
	}
	__antithesis_instrumentation__.Notify(646233)

	if rows.Err() != nil {
		__antithesis_instrumentation__.Notify(646239)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(646240)
	}
	__antithesis_instrumentation__.Notify(646234)

	return results, nil
}

func RunScrub(sqlDB *gosql.DB, database string, table string) error {
	__antithesis_instrumentation__.Notify(646241)
	return RunScrubWithOptions(sqlDB, database, table, "")
}

func RunScrubWithOptions(sqlDB *gosql.DB, database string, table string, options string) error {
	__antithesis_instrumentation__.Notify(646242)
	rows, err := sqlDB.Query(fmt.Sprintf(`EXPERIMENTAL SCRUB TABLE %s.%s %s`,
		database, table, options))
	if err != nil {
		__antithesis_instrumentation__.Notify(646246)
		return err
	} else {
		__antithesis_instrumentation__.Notify(646247)
	}
	__antithesis_instrumentation__.Notify(646243)

	results, err := GetScrubResultRows(rows)
	if err != nil {
		__antithesis_instrumentation__.Notify(646248)
		return err
	} else {
		__antithesis_instrumentation__.Notify(646249)
	}
	__antithesis_instrumentation__.Notify(646244)

	if len(results) > 0 {
		__antithesis_instrumentation__.Notify(646250)
		return errors.Errorf("expected no scrub results instead got: %#v", results)
	} else {
		__antithesis_instrumentation__.Notify(646251)
	}
	__antithesis_instrumentation__.Notify(646245)
	return nil
}
