package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/rand"
)

type orderStatusData struct {
	dID        int
	cID        int
	cFirst     string
	cMiddle    string
	cLast      string
	cBalance   float64
	oID        int
	oEntryD    time.Time
	oCarrierID pgtype.Int8

	items []orderItem
}

type customerData struct {
	cID      int
	cBalance float64
	cFirst   string
	cMiddle  string
}

type orderStatus struct {
	config *tpcc
	mcp    *workload.MultiConnPool
	sr     workload.SQLRunner

	selectByCustID   workload.StmtHandle
	selectByLastName workload.StmtHandle
	selectOrder      workload.StmtHandle
	selectItems      workload.StmtHandle

	a bufalloc.ByteAllocator
}

var _ tpccTx = &orderStatus{}

func createOrderStatus(
	ctx context.Context, config *tpcc, mcp *workload.MultiConnPool,
) (tpccTx, error) {
	__antithesis_instrumentation__.Notify(697961)
	o := &orderStatus{
		config: config,
		mcp:    mcp,
	}

	o.selectByCustID = o.sr.Define(`
		SELECT c_balance, c_first, c_middle, c_last
		FROM customer
		WHERE c_w_id = $1 AND c_d_id = $2 AND c_id = $3`,
	)

	o.selectByLastName = o.sr.Define(`
		SELECT c_id, c_balance, c_first, c_middle
		FROM customer
		WHERE c_w_id = $1 AND c_d_id = $2 AND c_last = $3
		ORDER BY c_first ASC`,
	)

	o.selectOrder = o.sr.Define(`
		SELECT o_id, o_entry_d, o_carrier_id
		FROM "order"
		WHERE o_w_id = $1 AND o_d_id = $2 AND o_c_id = $3
		ORDER BY o_id DESC
		LIMIT 1`,
	)

	o.selectItems = o.sr.Define(`
		SELECT ol_i_id, ol_supply_w_id, ol_quantity, ol_amount, ol_delivery_d
		FROM order_line
		WHERE ol_w_id = $1 AND ol_d_id = $2 AND ol_o_id = $3`,
	)

	if err := o.sr.Init(ctx, "order-status", mcp, config.connFlags); err != nil {
		__antithesis_instrumentation__.Notify(697963)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(697964)
	}
	__antithesis_instrumentation__.Notify(697962)

	return o, nil
}

func (o *orderStatus) run(ctx context.Context, wID int) (interface{}, error) {
	__antithesis_instrumentation__.Notify(697965)
	atomic.AddUint64(&o.config.auditor.orderStatusTransactions, 1)

	rng := rand.New(rand.NewSource(uint64(timeutil.Now().UnixNano())))

	d := orderStatusData{
		dID: rng.Intn(10) + 1,
	}

	if rng.Intn(100) < 60 {
		__antithesis_instrumentation__.Notify(697968)
		d.cLast = string(o.config.randCLast(rng, &o.a))
		atomic.AddUint64(&o.config.auditor.orderStatusByLastName, 1)
	} else {
		__antithesis_instrumentation__.Notify(697969)
		d.cID = o.config.randCustomerID(rng)
	}
	__antithesis_instrumentation__.Notify(697966)

	if err := crdbpgx.ExecuteTx(
		ctx, o.mcp.Get(), o.config.txOpts,
		func(tx pgx.Tx) error {
			__antithesis_instrumentation__.Notify(697970)

			if d.cID != 0 {
				__antithesis_instrumentation__.Notify(697975)

				if err := o.selectByCustID.QueryRowTx(
					ctx, tx, wID, d.dID, d.cID,
				).Scan(&d.cBalance, &d.cFirst, &d.cMiddle, &d.cLast); err != nil {
					__antithesis_instrumentation__.Notify(697976)
					return errors.Wrap(err, "select by customer idfail")
				} else {
					__antithesis_instrumentation__.Notify(697977)
				}
			} else {
				__antithesis_instrumentation__.Notify(697978)

				rows, err := o.selectByLastName.QueryTx(ctx, tx, wID, d.dID, d.cLast)
				if err != nil {
					__antithesis_instrumentation__.Notify(697983)
					return errors.Wrap(err, "select by last name fail")
				} else {
					__antithesis_instrumentation__.Notify(697984)
				}
				__antithesis_instrumentation__.Notify(697979)
				customers := make([]customerData, 0, 1)
				for rows.Next() {
					__antithesis_instrumentation__.Notify(697985)
					c := customerData{}
					err = rows.Scan(&c.cID, &c.cBalance, &c.cFirst, &c.cMiddle)
					if err != nil {
						__antithesis_instrumentation__.Notify(697987)
						rows.Close()
						return err
					} else {
						__antithesis_instrumentation__.Notify(697988)
					}
					__antithesis_instrumentation__.Notify(697986)
					customers = append(customers, c)
				}
				__antithesis_instrumentation__.Notify(697980)
				if err := rows.Err(); err != nil {
					__antithesis_instrumentation__.Notify(697989)
					return err
				} else {
					__antithesis_instrumentation__.Notify(697990)
				}
				__antithesis_instrumentation__.Notify(697981)
				rows.Close()
				if len(customers) == 0 {
					__antithesis_instrumentation__.Notify(697991)
					return errors.New("found no customers matching query orderStatus.selectByLastName")
				} else {
					__antithesis_instrumentation__.Notify(697992)
				}
				__antithesis_instrumentation__.Notify(697982)
				cIdx := (len(customers) - 1) / 2
				c := customers[cIdx]
				d.cID = c.cID
				d.cBalance = c.cBalance
				d.cFirst = c.cFirst
				d.cMiddle = c.cMiddle
			}
			__antithesis_instrumentation__.Notify(697971)

			if err := o.selectOrder.QueryRowTx(
				ctx, tx, wID, d.dID, d.cID,
			).Scan(&d.oID, &d.oEntryD, &d.oCarrierID); err != nil {
				__antithesis_instrumentation__.Notify(697993)
				return errors.Wrap(err, "select order fail")
			} else {
				__antithesis_instrumentation__.Notify(697994)
			}
			__antithesis_instrumentation__.Notify(697972)

			rows, err := o.selectItems.QueryTx(ctx, tx, wID, d.dID, d.oID)
			if err != nil {
				__antithesis_instrumentation__.Notify(697995)
				return errors.Wrap(err, "select items fail")
			} else {
				__antithesis_instrumentation__.Notify(697996)
			}
			__antithesis_instrumentation__.Notify(697973)
			defer rows.Close()

			d.items = make([]orderItem, 0, 10)
			for rows.Next() {
				__antithesis_instrumentation__.Notify(697997)
				item := orderItem{}
				if err := rows.Scan(&item.olIID, &item.olSupplyWID, &item.olQuantity, &item.olAmount, &item.olDeliveryD); err != nil {
					__antithesis_instrumentation__.Notify(697999)
					return err
				} else {
					__antithesis_instrumentation__.Notify(698000)
				}
				__antithesis_instrumentation__.Notify(697998)
				d.items = append(d.items, item)
			}
			__antithesis_instrumentation__.Notify(697974)
			return rows.Err()
		}); err != nil {
		__antithesis_instrumentation__.Notify(698001)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(698002)
	}
	__antithesis_instrumentation__.Notify(697967)
	return d, nil
}
