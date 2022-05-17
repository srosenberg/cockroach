package ledger

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"math/rand"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
)

type withdrawal struct{}

var _ ledgerTx = withdrawal{}

func (withdrawal) run(config *ledger, db *gosql.DB, rng *rand.Rand) (interface{}, error) {
	__antithesis_instrumentation__.Notify(694708)
	cID := randInt(rng, 0, config.customers-1)

	err := crdb.ExecuteTx(
		context.Background(),
		db,
		nil,
		func(tx *gosql.Tx) error {
			__antithesis_instrumentation__.Notify(694710)
			c, err := getBalance(tx, config, cID, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(694716)
				return err
			} else {
				__antithesis_instrumentation__.Notify(694717)
			}
			__antithesis_instrumentation__.Notify(694711)

			amount := randAmount(rng)
			c.balance -= amount
			if c.balance < 0 {
				__antithesis_instrumentation__.Notify(694718)
				c.balance = 0
			} else {
				__antithesis_instrumentation__.Notify(694719)
			}
			__antithesis_instrumentation__.Notify(694712)
			c.sequence++

			if err := updateBalance(tx, config, c); err != nil {
				__antithesis_instrumentation__.Notify(694720)
				return err
			} else {
				__antithesis_instrumentation__.Notify(694721)
			}
			__antithesis_instrumentation__.Notify(694713)

			tID, err := insertTransaction(tx, config, rng, c.identifier)
			if err != nil {
				__antithesis_instrumentation__.Notify(694722)
				return err
			} else {
				__antithesis_instrumentation__.Notify(694723)
			}
			__antithesis_instrumentation__.Notify(694714)

			cIDs := [2]int{
				cID,
				randInt(rng, 0, config.customers-1),
			}
			if err := insertEntries(tx, config, rng, cIDs, tID); err != nil {
				__antithesis_instrumentation__.Notify(694724)
				return err
			} else {
				__antithesis_instrumentation__.Notify(694725)
			}
			__antithesis_instrumentation__.Notify(694715)

			return insertSession(tx, config, rng)
		})
	__antithesis_instrumentation__.Notify(694709)
	return nil, err
}
