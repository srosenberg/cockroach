package ledger

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"math/rand"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
)

type reversal struct{}

var _ ledgerTx = reversal{}

func (reversal) run(config *ledger, db *gosql.DB, rng *rand.Rand) (interface{}, error) {
	__antithesis_instrumentation__.Notify(694705)
	err := crdb.ExecuteTx(
		context.Background(),
		db,
		nil,
		func(tx *gosql.Tx) error {
			__antithesis_instrumentation__.Notify(694707)

			return nil
		})
	__antithesis_instrumentation__.Notify(694706)
	return nil, err
}
