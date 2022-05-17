package ledger

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"math/rand"
)

type balance struct{}

var _ ledgerTx = balance{}

func (bal balance) run(config *ledger, db *gosql.DB, rng *rand.Rand) (interface{}, error) {
	__antithesis_instrumentation__.Notify(694580)
	cID := config.randCustomer(rng)

	_, err := getBalance(db, config, cID, config.historicalBalance)
	return nil, err
}
