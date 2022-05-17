package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/rand"
)

type delivery struct {
	config *tpcc
	mcp    *workload.MultiConnPool
	sr     workload.SQLRunner

	selectNewOrder workload.StmtHandle
	sumAmount      workload.StmtHandle
}

var _ tpccTx = &delivery{}

func createDelivery(
	ctx context.Context, config *tpcc, mcp *workload.MultiConnPool,
) (tpccTx, error) {
	__antithesis_instrumentation__.Notify(697754)
	del := &delivery{
		config: config,
		mcp:    mcp,
	}

	del.selectNewOrder = del.sr.Define(`
		SELECT no_o_id
		FROM new_order
		WHERE no_w_id = $1 AND no_d_id = $2
		ORDER BY no_o_id ASC
		LIMIT 1`,
	)

	del.sumAmount = del.sr.Define(`
		SELECT sum(ol_amount) FROM order_line
		WHERE ol_w_id = $1 AND ol_d_id = $2 AND ol_o_id = $3`,
	)

	if err := del.sr.Init(ctx, "delivery", mcp, config.connFlags); err != nil {
		__antithesis_instrumentation__.Notify(697756)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(697757)
	}
	__antithesis_instrumentation__.Notify(697755)

	return del, nil
}

func (del *delivery) run(ctx context.Context, wID int) (interface{}, error) {
	__antithesis_instrumentation__.Notify(697758)
	atomic.AddUint64(&del.config.auditor.deliveryTransactions, 1)

	rng := rand.New(rand.NewSource(uint64(timeutil.Now().UnixNano())))

	oCarrierID := rng.Intn(10) + 1
	olDeliveryD := timeutil.Now()

	err := crdbpgx.ExecuteTx(
		ctx, del.mcp.Get(), del.config.txOpts,
		func(tx pgx.Tx) error {
			__antithesis_instrumentation__.Notify(697760)

			dIDoIDPairs := make(map[int]int)
			dIDolTotalPairs := make(map[int]float64)
			for dID := 1; dID <= 10; dID++ {
				__antithesis_instrumentation__.Notify(697768)
				var oID int
				if err := del.selectNewOrder.QueryRowTx(ctx, tx, wID, dID).Scan(&oID); err != nil {
					__antithesis_instrumentation__.Notify(697771)

					if !errors.Is(err, gosql.ErrNoRows) {
						__antithesis_instrumentation__.Notify(697773)
						atomic.AddUint64(&del.config.auditor.skippedDelivieries, 1)
						return err
					} else {
						__antithesis_instrumentation__.Notify(697774)
					}
					__antithesis_instrumentation__.Notify(697772)
					continue
				} else {
					__antithesis_instrumentation__.Notify(697775)
				}
				__antithesis_instrumentation__.Notify(697769)
				dIDoIDPairs[dID] = oID

				var olTotal float64
				if err := del.sumAmount.QueryRowTx(
					ctx, tx, wID, dID, oID,
				).Scan(&olTotal); err != nil {
					__antithesis_instrumentation__.Notify(697776)
					return err
				} else {
					__antithesis_instrumentation__.Notify(697777)
				}
				__antithesis_instrumentation__.Notify(697770)
				dIDolTotalPairs[dID] = olTotal
			}
			__antithesis_instrumentation__.Notify(697761)
			dIDoIDPairsStr := makeInTuples(dIDoIDPairs)

			rows, err := tx.Query(
				ctx,
				fmt.Sprintf(`
					UPDATE "order"
					SET o_carrier_id = %d
					WHERE o_w_id = %d AND (o_d_id, o_id) IN (%s)
					RETURNING o_d_id, o_c_id`,
					oCarrierID, wID, dIDoIDPairsStr,
				),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(697778)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697779)
			}
			__antithesis_instrumentation__.Notify(697762)
			dIDcIDPairs := make(map[int]int)
			for rows.Next() {
				__antithesis_instrumentation__.Notify(697780)
				var dID, oCID int
				if err := rows.Scan(&dID, &oCID); err != nil {
					__antithesis_instrumentation__.Notify(697782)
					rows.Close()
					return err
				} else {
					__antithesis_instrumentation__.Notify(697783)
				}
				__antithesis_instrumentation__.Notify(697781)
				dIDcIDPairs[dID] = oCID
			}
			__antithesis_instrumentation__.Notify(697763)
			if err := rows.Err(); err != nil {
				__antithesis_instrumentation__.Notify(697784)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697785)
			}
			__antithesis_instrumentation__.Notify(697764)
			rows.Close()

			if err := checkSameKeys(dIDoIDPairs, dIDcIDPairs); err != nil {
				__antithesis_instrumentation__.Notify(697786)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697787)
			}
			__antithesis_instrumentation__.Notify(697765)
			dIDcIDPairsStr := makeInTuples(dIDcIDPairs)
			dIDToOlTotalStr := makeWhereCases(dIDolTotalPairs)

			if _, err := tx.Exec(
				ctx,
				fmt.Sprintf(`
					UPDATE customer
					SET c_delivery_cnt = c_delivery_cnt + 1,
						c_balance = c_balance + CASE c_d_id %s END
					WHERE c_w_id = %d AND (c_d_id, c_id) IN (%s)`,
					dIDToOlTotalStr, wID, dIDcIDPairsStr,
				),
			); err != nil {
				__antithesis_instrumentation__.Notify(697788)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697789)
			}
			__antithesis_instrumentation__.Notify(697766)
			if _, err := tx.Exec(
				ctx,
				fmt.Sprintf(`
					DELETE FROM new_order
					WHERE no_w_id = %d AND (no_d_id, no_o_id) IN (%s)`,
					wID, dIDoIDPairsStr,
				),
			); err != nil {
				__antithesis_instrumentation__.Notify(697790)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697791)
			}
			__antithesis_instrumentation__.Notify(697767)

			_, err = tx.Exec(
				ctx,
				fmt.Sprintf(`
					UPDATE order_line
					SET ol_delivery_d = '%s'
					WHERE ol_w_id = %d AND (ol_d_id, ol_o_id) IN (%s)`,
					olDeliveryD.Format("2006-01-02 15:04:05"), wID, dIDoIDPairsStr,
				),
			)
			return err
		})
	__antithesis_instrumentation__.Notify(697759)
	return nil, err
}

func makeInTuples(pairs map[int]int) string {
	__antithesis_instrumentation__.Notify(697792)
	tupleStrs := make([]string, 0, len(pairs))
	for k, v := range pairs {
		__antithesis_instrumentation__.Notify(697794)
		tupleStrs = append(tupleStrs, fmt.Sprintf("(%d, %d)", k, v))
	}
	__antithesis_instrumentation__.Notify(697793)
	return strings.Join(tupleStrs, ", ")
}

func makeWhereCases(cases map[int]float64) string {
	__antithesis_instrumentation__.Notify(697795)
	casesStrs := make([]string, 0, len(cases))
	for k, v := range cases {
		__antithesis_instrumentation__.Notify(697797)
		casesStrs = append(casesStrs, fmt.Sprintf("WHEN %d THEN %f", k, v))
	}
	__antithesis_instrumentation__.Notify(697796)
	return strings.Join(casesStrs, " ")
}

func checkSameKeys(a, b map[int]int) error {
	__antithesis_instrumentation__.Notify(697798)
	if len(a) != len(b) {
		__antithesis_instrumentation__.Notify(697801)
		return errors.Errorf("different number of keys")
	} else {
		__antithesis_instrumentation__.Notify(697802)
	}
	__antithesis_instrumentation__.Notify(697799)
	for k := range a {
		__antithesis_instrumentation__.Notify(697803)
		if _, ok := b[k]; !ok {
			__antithesis_instrumentation__.Notify(697804)
			return errors.Errorf("missing key %v", k)
		} else {
			__antithesis_instrumentation__.Notify(697805)
		}
	}
	__antithesis_instrumentation__.Notify(697800)
	return nil
}
