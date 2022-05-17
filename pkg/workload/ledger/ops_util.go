package ledger

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"database/sql/driver"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"golang.org/x/sync/syncmap"
)

type querier interface {
	Exec(query string, args ...interface{}) (gosql.Result, error)
	Query(query string, args ...interface{}) (*gosql.Rows, error)
	QueryRow(query string, args ...interface{}) *gosql.Row
}

var sqlParamRE = regexp.MustCompile(`\$(\d+)`)
var replacedSQLParams syncmap.Map

func replaceSQLParams(s string) string {
	__antithesis_instrumentation__.Notify(694648)

	if res, ok := replacedSQLParams.Load(s); ok {
		__antithesis_instrumentation__.Notify(694650)
		return res.(string)
	} else {
		__antithesis_instrumentation__.Notify(694651)
	}
	__antithesis_instrumentation__.Notify(694649)

	res := sqlParamRE.ReplaceAllString(s, "'%[${1}]v'")
	replacedSQLParams.Store(s, res)
	return res
}

func maybeInlineStmtArgs(
	config *ledger, query string, args ...interface{},
) (string, []interface{}) {
	__antithesis_instrumentation__.Notify(694652)
	if !config.inlineArgs {
		__antithesis_instrumentation__.Notify(694655)
		return query, args
	} else {
		__antithesis_instrumentation__.Notify(694656)
	}
	__antithesis_instrumentation__.Notify(694653)
	queryFmt := replaceSQLParams(query)
	for i, arg := range args {
		__antithesis_instrumentation__.Notify(694657)
		if v, ok := arg.(driver.Valuer); ok {
			__antithesis_instrumentation__.Notify(694658)
			val, err := v.Value()
			if err != nil {
				__antithesis_instrumentation__.Notify(694660)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(694661)
			}
			__antithesis_instrumentation__.Notify(694659)
			args[i] = val
		} else {
			__antithesis_instrumentation__.Notify(694662)
		}
	}
	__antithesis_instrumentation__.Notify(694654)
	return strings.Replace(
			strings.Replace(
				fmt.Sprintf(queryFmt, args...),
				" UTC", "", -1),
			`'<nil>'`, `NULL`, -1),
		nil
}

type customer struct {
	id               int
	identifier       string
	name             gosql.NullString
	currencyCode     string
	isSystemCustomer bool
	isActive         bool
	created          time.Time
	balance          float64
	creditLimit      gosql.NullFloat64
	sequence         int
}

func getBalance(q querier, config *ledger, id int, historical bool) (customer, error) {
	__antithesis_instrumentation__.Notify(694663)
	aostSpec := ""
	if historical {
		__antithesis_instrumentation__.Notify(694667)
		aostSpec = " AS OF SYSTEM TIME '-10s'"
	} else {
		__antithesis_instrumentation__.Notify(694668)
	}
	__antithesis_instrumentation__.Notify(694664)
	stmt, args := maybeInlineStmtArgs(config, `
		SELECT
			id,
			identifier,
			"name",
			currency_code,
			is_system_customer,
			is_active,
			created,
			balance,
			credit_limit,
			sequence_number
		FROM customer`+
		aostSpec+`
		WHERE id = $1 AND IS_ACTIVE = true`,
		id,
	)
	rows, err := q.Query(stmt, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(694669)
		return customer{}, err
	} else {
		__antithesis_instrumentation__.Notify(694670)
	}
	__antithesis_instrumentation__.Notify(694665)
	defer rows.Close()

	var c customer
	for rows.Next() {
		__antithesis_instrumentation__.Notify(694671)
		if err := rows.Scan(
			&c.id,
			&c.identifier,
			&c.name,
			&c.currencyCode,
			&c.isSystemCustomer,
			&c.isActive,
			&c.created,
			&c.balance,
			&c.creditLimit,
			&c.sequence,
		); err != nil {
			__antithesis_instrumentation__.Notify(694672)
			return customer{}, err
		} else {
			__antithesis_instrumentation__.Notify(694673)
		}
	}
	__antithesis_instrumentation__.Notify(694666)
	return c, rows.Err()
}

func updateBalance(q querier, config *ledger, c customer) error {
	__antithesis_instrumentation__.Notify(694674)
	stmt, args := maybeInlineStmtArgs(config, `
		UPDATE customer SET
			balance         = $1,
			credit_limit    = $2,
			is_active       = $3,
			name            = $4,
			sequence_number = $5
		WHERE id = $6`,
		c.balance, c.creditLimit, c.isActive, c.name, c.sequence, c.id,
	)
	_, err := q.Exec(stmt, args...)
	return err
}

func insertTransaction(q querier, config *ledger, rng *rand.Rand, username string) (string, error) {
	__antithesis_instrumentation__.Notify(694675)
	tID := randPaymentID(rng)

	stmt, args := maybeInlineStmtArgs(config, `
		INSERT INTO transaction (
			tcomment, context, response, reversed_by, created_ts, 
			transaction_type_reference, username, external_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		nil, randContext(rng), randResponse(rng), nil,
		timeutil.Now(), txnTypeReference, username, tID,
	)
	_, err := q.Exec(stmt, args...)
	return tID, err
}

func insertEntries(q querier, config *ledger, rng *rand.Rand, cIDs [2]int, tID string) error {
	__antithesis_instrumentation__.Notify(694676)
	amount1 := randAmount(rng)
	sysAmount := 88.433571
	ts := timeutil.Now()

	stmt, args := maybeInlineStmtArgs(config, `
		INSERT INTO entry (
			amount, system_amount, created_ts, transaction_id, customer_id, money_type
		) VALUES
			($1 , $2 , $3 , $4 , $5 , $6 ),
			($7 , $8 , $9 , $10, $11, $12)`,
		amount1, sysAmount, ts, tID, cIDs[0], cashMoneyType,
		-amount1, -sysAmount, ts, tID, cIDs[1], cashMoneyType,
	)
	_, err := q.Exec(stmt, args...)
	return err
}

func getSession(q querier, config *ledger, rng *rand.Rand) error {
	__antithesis_instrumentation__.Notify(694677)
	stmt, args := maybeInlineStmtArgs(config, `
		SELECT
			session_id,
			expiry_timestamp,
			data,
			last_update
		FROM session
		WHERE session_id >= $1
		LIMIT 1`,
		randSessionID(rng),
	)
	rows, err := q.Query(stmt, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(694680)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694681)
	}
	__antithesis_instrumentation__.Notify(694678)
	defer rows.Close()

	for rows.Next() {
		__antithesis_instrumentation__.Notify(694682)

	}
	__antithesis_instrumentation__.Notify(694679)
	return rows.Err()
}

func insertSession(q querier, config *ledger, rng *rand.Rand) error {
	__antithesis_instrumentation__.Notify(694683)
	stmt, args := maybeInlineStmtArgs(config, `
		INSERT INTO session (
			data, expiry_timestamp, last_update, session_id
		) VALUES ($1, $2, $3, $4)`,
		randSessionData(rng), randTimestamp(rng), timeutil.Now(), randSessionID(rng),
	)
	_, err := q.Exec(stmt, args...)
	return err
}
