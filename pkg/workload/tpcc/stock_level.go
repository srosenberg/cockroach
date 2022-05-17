package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/rand"
)

type stockLevelData struct {
	dID       int
	threshold int
	lowStock  int
}

type stockLevel struct {
	config *tpcc
	mcp    *workload.MultiConnPool
	sr     workload.SQLRunner

	selectDNextOID    workload.StmtHandle
	countRecentlySold workload.StmtHandle
}

var _ tpccTx = &stockLevel{}

func createStockLevel(
	ctx context.Context, config *tpcc, mcp *workload.MultiConnPool,
) (tpccTx, error) {
	__antithesis_instrumentation__.Notify(698306)
	s := &stockLevel{
		config: config,
		mcp:    mcp,
	}

	s.selectDNextOID = s.sr.Define(`
		SELECT d_next_o_id
		FROM district
		WHERE d_w_id = $1 AND d_id = $2`,
	)

	s.countRecentlySold = s.sr.Define(`
		SELECT count(*) FROM (
			SELECT DISTINCT s_i_id
			FROM order_line
			JOIN stock
			ON s_i_id=ol_i_id AND s_w_id=ol_w_id
			WHERE ol_w_id = $1
				AND ol_d_id = $2
				AND ol_o_id BETWEEN $3 - 20 AND $3 - 1
				AND s_quantity < $4
		)`,
	)

	if err := s.sr.Init(ctx, "stock-level", mcp, config.connFlags); err != nil {
		__antithesis_instrumentation__.Notify(698308)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(698309)
	}
	__antithesis_instrumentation__.Notify(698307)

	return s, nil
}

func (s *stockLevel) run(ctx context.Context, wID int) (interface{}, error) {
	__antithesis_instrumentation__.Notify(698310)
	rng := rand.New(rand.NewSource(uint64(timeutil.Now().UnixNano())))

	d := stockLevelData{
		threshold: int(randInt(rng, 10, 20)),
		dID:       rng.Intn(10) + 1,
	}

	if err := crdbpgx.ExecuteTx(
		ctx, s.mcp.Get(), s.config.txOpts,
		func(tx pgx.Tx) error {
			__antithesis_instrumentation__.Notify(698312)
			var dNextOID int
			if err := s.selectDNextOID.QueryRowTx(
				ctx, tx, wID, d.dID,
			).Scan(&dNextOID); err != nil {
				__antithesis_instrumentation__.Notify(698314)
				return err
			} else {
				__antithesis_instrumentation__.Notify(698315)
			}
			__antithesis_instrumentation__.Notify(698313)

			return s.countRecentlySold.QueryRowTx(
				ctx, tx, wID, d.dID, dNextOID, d.threshold,
			).Scan(&d.lowStock)
		}); err != nil {
		__antithesis_instrumentation__.Notify(698316)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(698317)
	}
	__antithesis_instrumentation__.Notify(698311)
	return d, nil
}
