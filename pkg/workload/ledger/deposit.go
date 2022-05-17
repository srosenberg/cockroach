package ledger

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"math/rand"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
)

type deposit struct{}

var _ ledgerTx = deposit{}

func (bal deposit) run(config *ledger, db *gosql.DB, rng *rand.Rand) (interface{}, error) {
	__antithesis_instrumentation__.Notify(694581)
	cID := randInt(rng, 0, config.customers-1)

	err := crdb.ExecuteTx(
		context.Background(),
		db,
		nil,
		func(tx *gosql.Tx) error {
			__antithesis_instrumentation__.Notify(694583)
			c, err := getBalance(tx, config, cID, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(694588)
				return err
			} else {
				__antithesis_instrumentation__.Notify(694589)
			}
			__antithesis_instrumentation__.Notify(694584)

			amount := randAmount(rng)
			c.balance += amount
			c.sequence++

			if err := updateBalance(tx, config, c); err != nil {
				__antithesis_instrumentation__.Notify(694590)
				return err
			} else {
				__antithesis_instrumentation__.Notify(694591)
			}
			__antithesis_instrumentation__.Notify(694585)

			if err := getSession(tx, config, rng); err != nil {
				__antithesis_instrumentation__.Notify(694592)
				return err
			} else {
				__antithesis_instrumentation__.Notify(694593)
			}
			__antithesis_instrumentation__.Notify(694586)

			tID, err := insertTransaction(tx, config, rng, c.identifier)
			if err != nil {
				__antithesis_instrumentation__.Notify(694594)
				return err
			} else {
				__antithesis_instrumentation__.Notify(694595)
			}
			__antithesis_instrumentation__.Notify(694587)

			cIDs := [2]int{
				cID,
				randInt(rng, 0, config.customers-1),
			}
			return insertEntries(tx, config, rng, cIDs, tID)
		})
	__antithesis_instrumentation__.Notify(694582)
	return nil, err
}
