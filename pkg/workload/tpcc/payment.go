package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/rand"
)

type paymentData struct {
	dID  int
	cID  int
	cDID int
	cWID int

	wStreet1 string
	wStreet2 string
	wCity    string
	wState   string
	wZip     string

	dStreet1 string
	dStreet2 string
	dCity    string
	dState   string
	dZip     string

	cFirst     string
	cMiddle    string
	cLast      string
	cStreet1   string
	cStreet2   string
	cCity      string
	cState     string
	cZip       string
	cPhone     string
	cSince     time.Time
	cCredit    string
	cCreditLim float64
	cDiscount  float64
	cBalance   float64
	cData      string

	hAmount float64
	hDate   time.Time
}

type payment struct {
	config *tpcc
	mcp    *workload.MultiConnPool
	sr     workload.SQLRunner

	updateWarehouse   workload.StmtHandle
	updateDistrict    workload.StmtHandle
	selectByLastName  workload.StmtHandle
	updateWithPayment workload.StmtHandle
	insertHistory     workload.StmtHandle

	a bufalloc.ByteAllocator
}

var _ tpccTx = &payment{}

func createPayment(ctx context.Context, config *tpcc, mcp *workload.MultiConnPool) (tpccTx, error) {
	__antithesis_instrumentation__.Notify(698189)
	p := &payment{
		config: config,
		mcp:    mcp,
	}

	p.updateWarehouse = p.sr.Define(`
		UPDATE warehouse
		SET w_ytd = w_ytd + $1
		WHERE w_id = $2
		RETURNING w_name, w_street_1, w_street_2, w_city, w_state, w_zip`,
	)

	p.updateDistrict = p.sr.Define(`
		UPDATE district
		SET d_ytd = d_ytd + $1
		WHERE d_w_id = $2 AND d_id = $3
		RETURNING d_name, d_street_1, d_street_2, d_city, d_state, d_zip`,
	)

	p.selectByLastName = p.sr.Define(`
		SELECT c_id
		FROM customer
		WHERE c_w_id = $1 AND c_d_id = $2 AND c_last = $3
		ORDER BY c_first ASC`,
	)

	p.updateWithPayment = p.sr.Define(`
		UPDATE customer
		SET (c_balance, c_ytd_payment, c_payment_cnt, c_data) =
		(c_balance - ($1:::float)::decimal, c_ytd_payment + ($1:::float)::decimal, c_payment_cnt + 1,
			 case c_credit when 'BC' then
			 left(c_id::text || c_d_id::text || c_w_id::text || ($5:::int)::text || ($6:::int)::text || ($1:::float)::text || c_data, 500)
			 else c_data end)
		WHERE c_w_id = $2 AND c_d_id = $3 AND c_id = $4
		RETURNING c_first, c_middle, c_last, c_street_1, c_street_2,
					c_city, c_state, c_zip, c_phone, c_since, c_credit,
					c_credit_lim, c_discount, c_balance, case c_credit when 'BC' then left(c_data, 200) else '' end`,
	)

	p.insertHistory = p.sr.Define(`
		INSERT INTO history (h_c_id, h_c_d_id, h_c_w_id, h_d_id, h_w_id, h_amount, h_date, h_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
	)

	if err := p.sr.Init(ctx, "payment", mcp, config.connFlags); err != nil {
		__antithesis_instrumentation__.Notify(698191)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(698192)
	}
	__antithesis_instrumentation__.Notify(698190)

	return p, nil
}

func (p *payment) run(ctx context.Context, wID int) (interface{}, error) {
	__antithesis_instrumentation__.Notify(698193)
	atomic.AddUint64(&p.config.auditor.paymentTransactions, 1)

	rng := rand.New(rand.NewSource(uint64(timeutil.Now().UnixNano())))

	d := paymentData{
		dID: rng.Intn(10) + 1,

		hAmount: float64(randInt(rng, 100, 500000)) / float64(100.0),
		hDate:   timeutil.Now(),
	}

	if p.config.localWarehouses || func() bool {
		__antithesis_instrumentation__.Notify(698197)
		return rng.Intn(100) < 85 == true
	}() == true {
		__antithesis_instrumentation__.Notify(698198)
		d.cWID = wID
		d.cDID = d.dID
	} else {
		__antithesis_instrumentation__.Notify(698199)
		d.cWID = p.config.wPart.randActive(rng)

		for d.cWID == wID && func() bool {
			__antithesis_instrumentation__.Notify(698201)
			return p.config.activeWarehouses > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(698202)
			d.cWID = p.config.wPart.randActive(rng)
		}
		__antithesis_instrumentation__.Notify(698200)
		p.config.auditor.Lock()
		p.config.auditor.paymentRemoteWarehouseFreq[d.cWID]++
		p.config.auditor.Unlock()
		d.cDID = rng.Intn(10) + 1
	}
	__antithesis_instrumentation__.Notify(698194)

	if rng.Intn(100) < 60 {
		__antithesis_instrumentation__.Notify(698203)
		d.cLast = string(p.config.randCLast(rng, &p.a))
		atomic.AddUint64(&p.config.auditor.paymentsByLastName, 1)
	} else {
		__antithesis_instrumentation__.Notify(698204)
		d.cID = p.config.randCustomerID(rng)
	}
	__antithesis_instrumentation__.Notify(698195)

	if err := crdbpgx.ExecuteTx(
		ctx, p.mcp.Get(), p.config.txOpts,
		func(tx pgx.Tx) error {
			__antithesis_instrumentation__.Notify(698205)
			var wName, dName string

			if err := p.updateWarehouse.QueryRowTx(
				ctx, tx, d.hAmount, wID,
			).Scan(&wName, &d.wStreet1, &d.wStreet2, &d.wCity, &d.wState, &d.wZip); err != nil {
				__antithesis_instrumentation__.Notify(698210)
				return err
			} else {
				__antithesis_instrumentation__.Notify(698211)
			}
			__antithesis_instrumentation__.Notify(698206)

			if err := p.updateDistrict.QueryRowTx(
				ctx, tx, d.hAmount, wID, d.dID,
			).Scan(&dName, &d.dStreet1, &d.dStreet2, &d.dCity, &d.dState, &d.dZip); err != nil {
				__antithesis_instrumentation__.Notify(698212)
				return err
			} else {
				__antithesis_instrumentation__.Notify(698213)
			}
			__antithesis_instrumentation__.Notify(698207)

			if d.cID == 0 {
				__antithesis_instrumentation__.Notify(698214)

				rows, err := p.selectByLastName.QueryTx(ctx, tx, wID, d.dID, d.cLast)
				if err != nil {
					__antithesis_instrumentation__.Notify(698218)
					return errors.Wrap(err, "select by last name fail")
				} else {
					__antithesis_instrumentation__.Notify(698219)
				}
				__antithesis_instrumentation__.Notify(698215)
				customers := make([]int, 0, 1)
				for rows.Next() {
					__antithesis_instrumentation__.Notify(698220)
					var cID int
					err = rows.Scan(&cID)
					if err != nil {
						__antithesis_instrumentation__.Notify(698222)
						rows.Close()
						return err
					} else {
						__antithesis_instrumentation__.Notify(698223)
					}
					__antithesis_instrumentation__.Notify(698221)
					customers = append(customers, cID)
				}
				__antithesis_instrumentation__.Notify(698216)
				if err := rows.Err(); err != nil {
					__antithesis_instrumentation__.Notify(698224)
					return err
				} else {
					__antithesis_instrumentation__.Notify(698225)
				}
				__antithesis_instrumentation__.Notify(698217)
				rows.Close()
				cIdx := (len(customers) - 1) / 2
				d.cID = customers[cIdx]
			} else {
				__antithesis_instrumentation__.Notify(698226)
			}
			__antithesis_instrumentation__.Notify(698208)

			if err := p.updateWithPayment.QueryRowTx(
				ctx, tx, d.hAmount, d.cWID, d.cDID, d.cID, d.dID, wID,
			).Scan(&d.cFirst, &d.cMiddle, &d.cLast, &d.cStreet1, &d.cStreet2,
				&d.cCity, &d.cState, &d.cZip, &d.cPhone, &d.cSince, &d.cCredit,
				&d.cCreditLim, &d.cDiscount, &d.cBalance, &d.cData,
			); err != nil {
				__antithesis_instrumentation__.Notify(698227)
				return errors.Wrap(err, "select by customer idfail")
			} else {
				__antithesis_instrumentation__.Notify(698228)
			}
			__antithesis_instrumentation__.Notify(698209)

			hData := fmt.Sprintf("%s    %s", wName, dName)

			_, err := p.insertHistory.ExecTx(
				ctx, tx,
				d.cID, d.cDID, d.cWID, d.dID, wID, d.hAmount, d.hDate.Format("2006-01-02 15:04:05"), hData,
			)
			return err
		}); err != nil {
		__antithesis_instrumentation__.Notify(698229)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(698230)
	}
	__antithesis_instrumentation__.Notify(698196)
	return d, nil
}
