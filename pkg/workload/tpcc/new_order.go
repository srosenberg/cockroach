package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq"
	"golang.org/x/exp/rand"
)

type orderItem struct {
	olSupplyWID  int
	olIID        int
	olNumber     int
	iName        string
	olQuantity   int
	brandGeneric string
	iPrice       float64
	olAmount     float64
	olDeliveryD  pq.NullTime

	remoteWarehouse bool
}

type newOrderData struct {
	wID         int
	dID         int
	cID         int
	oID         int
	oOlCnt      int
	cLast       string
	cCredit     string
	cDiscount   float64
	wTax        float64
	dTax        float64
	oEntryD     time.Time
	totalAmount float64

	items []orderItem
}

var errSimulated = errors.New("simulated user error")

type newOrder struct {
	config *tpcc
	mcp    *workload.MultiConnPool
	sr     workload.SQLRunner

	updateDistrict     workload.StmtHandle
	selectWarehouseTax workload.StmtHandle
	selectCustomerInfo workload.StmtHandle
	insertOrder        workload.StmtHandle
	insertNewOrder     workload.StmtHandle
}

var _ tpccTx = &newOrder{}

func createNewOrder(
	ctx context.Context, config *tpcc, mcp *workload.MultiConnPool,
) (tpccTx, error) {
	__antithesis_instrumentation__.Notify(697852)
	n := &newOrder{
		config: config,
		mcp:    mcp,
	}

	n.updateDistrict = n.sr.Define(`
		UPDATE district
		SET d_next_o_id = d_next_o_id + 1
		WHERE d_w_id = $1 AND d_id = $2
		RETURNING d_tax, d_next_o_id`,
	)

	n.selectWarehouseTax = n.sr.Define(`
		SELECT w_tax FROM warehouse WHERE w_id = $1`,
	)

	n.selectCustomerInfo = n.sr.Define(`
		SELECT c_discount, c_last, c_credit
		FROM customer
		WHERE c_w_id = $1 AND c_d_id = $2 AND c_id = $3`,
	)

	n.insertOrder = n.sr.Define(`
		INSERT INTO "order" (o_id, o_d_id, o_w_id, o_c_id, o_entry_d, o_ol_cnt, o_all_local)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
	)

	n.insertNewOrder = n.sr.Define(`
		INSERT INTO new_order (no_o_id, no_d_id, no_w_id)
		VALUES ($1, $2, $3)`,
	)

	if err := n.sr.Init(ctx, "new-order", mcp, config.connFlags); err != nil {
		__antithesis_instrumentation__.Notify(697854)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(697855)
	}
	__antithesis_instrumentation__.Notify(697853)

	return n, nil
}

func (n *newOrder) run(ctx context.Context, wID int) (interface{}, error) {
	__antithesis_instrumentation__.Notify(697856)
	atomic.AddUint64(&n.config.auditor.newOrderTransactions, 1)

	rng := rand.New(rand.NewSource(uint64(timeutil.Now().UnixNano())))

	d := newOrderData{
		wID:    wID,
		dID:    int(randInt(rng, 1, 10)),
		cID:    n.config.randCustomerID(rng),
		oOlCnt: int(randInt(rng, 5, 15)),
	}
	d.items = make([]orderItem, d.oOlCnt)

	n.config.auditor.Lock()
	n.config.auditor.orderLinesFreq[d.oOlCnt]++
	n.config.auditor.Unlock()
	atomic.AddUint64(&n.config.auditor.totalOrderLines, uint64(d.oOlCnt))

	itemIDs := make(map[int]struct{})

	rollback := rng.Intn(100) == 0

	allLocal := 1
	for i := 0; i < d.oOlCnt; i++ {
		__antithesis_instrumentation__.Notify(697861)
		item := orderItem{
			olNumber: i + 1,

			olQuantity: rng.Intn(10) + 1,
		}

		if rollback && func() bool {
			__antithesis_instrumentation__.Notify(697865)
			return i == d.oOlCnt-1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(697866)
			item.olIID = -1
		} else {
			__antithesis_instrumentation__.Notify(697867)

			for {
				__antithesis_instrumentation__.Notify(697868)
				item.olIID = n.config.randItemID(rng)
				if _, ok := itemIDs[item.olIID]; !ok {
					__antithesis_instrumentation__.Notify(697869)
					itemIDs[item.olIID] = struct{}{}
					break
				} else {
					__antithesis_instrumentation__.Notify(697870)
				}
			}
		}
		__antithesis_instrumentation__.Notify(697862)

		if n.config.localWarehouses {
			__antithesis_instrumentation__.Notify(697871)
			item.remoteWarehouse = false
		} else {
			__antithesis_instrumentation__.Notify(697872)
			item.remoteWarehouse = rng.Intn(100) == 0
		}
		__antithesis_instrumentation__.Notify(697863)
		item.olSupplyWID = wID
		if item.remoteWarehouse && func() bool {
			__antithesis_instrumentation__.Notify(697873)
			return n.config.activeWarehouses > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(697874)
			allLocal = 0

			item.olSupplyWID = n.config.wPart.randActive(rng)
			for item.olSupplyWID == wID {
				__antithesis_instrumentation__.Notify(697876)
				item.olSupplyWID = n.config.wPart.randActive(rng)
			}
			__antithesis_instrumentation__.Notify(697875)
			n.config.auditor.Lock()
			n.config.auditor.orderLineRemoteWarehouseFreq[item.olSupplyWID]++
			n.config.auditor.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(697877)
			item.olSupplyWID = wID
		}
		__antithesis_instrumentation__.Notify(697864)
		d.items[i] = item
	}
	__antithesis_instrumentation__.Notify(697857)

	sort.Slice(d.items, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(697878)
		return d.items[i].olIID < d.items[j].olIID
	})
	__antithesis_instrumentation__.Notify(697858)

	d.oEntryD = timeutil.Now()

	err := crdbpgx.ExecuteTx(
		ctx, n.mcp.Get(), n.config.txOpts,
		func(tx pgx.Tx) error {
			__antithesis_instrumentation__.Notify(697879)

			var dNextOID int
			if err := n.updateDistrict.QueryRowTx(
				ctx, tx, d.wID, d.dID,
			).Scan(&d.dTax, &dNextOID); err != nil {
				__antithesis_instrumentation__.Notify(697898)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697899)
			}
			__antithesis_instrumentation__.Notify(697880)
			d.oID = dNextOID - 1

			if err := n.selectWarehouseTax.QueryRowTx(
				ctx, tx, wID,
			).Scan(&d.wTax); err != nil {
				__antithesis_instrumentation__.Notify(697900)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697901)
			}
			__antithesis_instrumentation__.Notify(697881)

			if err := n.selectCustomerInfo.QueryRowTx(
				ctx, tx, d.wID, d.dID, d.cID,
			).Scan(&d.cDiscount, &d.cLast, &d.cCredit); err != nil {
				__antithesis_instrumentation__.Notify(697902)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697903)
			}
			__antithesis_instrumentation__.Notify(697882)

			itemIDs := make([]string, d.oOlCnt)
			for i, item := range d.items {
				__antithesis_instrumentation__.Notify(697904)
				itemIDs[i] = fmt.Sprint(item.olIID)
			}
			__antithesis_instrumentation__.Notify(697883)
			rows, err := tx.Query(
				ctx,
				fmt.Sprintf(`
					SELECT i_price, i_name, i_data
					FROM item
					WHERE i_id IN (%[1]s)
					ORDER BY i_id`,
					strings.Join(itemIDs, ", "),
				),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(697905)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697906)
			}
			__antithesis_instrumentation__.Notify(697884)
			iDatas := make([]string, d.oOlCnt)
			for i := range d.items {
				__antithesis_instrumentation__.Notify(697907)
				item := &d.items[i]
				iData := &iDatas[i]

				if !rows.Next() {
					__antithesis_instrumentation__.Notify(697909)
					if err := rows.Err(); err != nil {
						__antithesis_instrumentation__.Notify(697912)
						return err
					} else {
						__antithesis_instrumentation__.Notify(697913)
					}
					__antithesis_instrumentation__.Notify(697910)
					if rollback {
						__antithesis_instrumentation__.Notify(697914)

						atomic.AddUint64(&n.config.auditor.newOrderRollbacks, 1)
						return errSimulated
					} else {
						__antithesis_instrumentation__.Notify(697915)
					}
					__antithesis_instrumentation__.Notify(697911)
					return errors.New("missing item row")
				} else {
					__antithesis_instrumentation__.Notify(697916)
				}
				__antithesis_instrumentation__.Notify(697908)

				err = rows.Scan(&item.iPrice, &item.iName, iData)
				if err != nil {
					__antithesis_instrumentation__.Notify(697917)
					rows.Close()
					return err
				} else {
					__antithesis_instrumentation__.Notify(697918)
				}
			}
			__antithesis_instrumentation__.Notify(697885)
			if rows.Next() {
				__antithesis_instrumentation__.Notify(697919)
				return errors.New("extra item row")
			} else {
				__antithesis_instrumentation__.Notify(697920)
			}
			__antithesis_instrumentation__.Notify(697886)
			if err := rows.Err(); err != nil {
				__antithesis_instrumentation__.Notify(697921)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697922)
			}
			__antithesis_instrumentation__.Notify(697887)
			rows.Close()

			stockIDs := make([]string, d.oOlCnt)
			for i, item := range d.items {
				__antithesis_instrumentation__.Notify(697923)
				stockIDs[i] = fmt.Sprintf("(%d, %d)", item.olIID, item.olSupplyWID)
			}
			__antithesis_instrumentation__.Notify(697888)
			rows, err = tx.Query(
				ctx,
				fmt.Sprintf(`
					SELECT s_quantity, s_ytd, s_order_cnt, s_remote_cnt, s_data, s_dist_%02[1]d
					FROM stock
					WHERE (s_i_id, s_w_id) IN (%[2]s)
					ORDER BY s_i_id`,
					d.dID, strings.Join(stockIDs, ", "),
				),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(697924)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697925)
			}
			__antithesis_instrumentation__.Notify(697889)
			distInfos := make([]string, d.oOlCnt)
			sQuantityUpdateCases := make([]string, d.oOlCnt)
			sYtdUpdateCases := make([]string, d.oOlCnt)
			sOrderCntUpdateCases := make([]string, d.oOlCnt)
			sRemoteCntUpdateCases := make([]string, d.oOlCnt)
			for i := range d.items {
				__antithesis_instrumentation__.Notify(697926)
				item := &d.items[i]

				if !rows.Next() {
					__antithesis_instrumentation__.Notify(697932)
					if err := rows.Err(); err != nil {
						__antithesis_instrumentation__.Notify(697934)
						return err
					} else {
						__antithesis_instrumentation__.Notify(697935)
					}
					__antithesis_instrumentation__.Notify(697933)
					return errors.New("missing stock row")
				} else {
					__antithesis_instrumentation__.Notify(697936)
				}
				__antithesis_instrumentation__.Notify(697927)

				var sQuantity, sYtd, sOrderCnt, sRemoteCnt int
				var sData string
				err = rows.Scan(&sQuantity, &sYtd, &sOrderCnt, &sRemoteCnt, &sData, &distInfos[i])
				if err != nil {
					__antithesis_instrumentation__.Notify(697937)
					rows.Close()
					return err
				} else {
					__antithesis_instrumentation__.Notify(697938)
				}
				__antithesis_instrumentation__.Notify(697928)

				if strings.Contains(sData, originalString) && func() bool {
					__antithesis_instrumentation__.Notify(697939)
					return strings.Contains(iDatas[i], originalString) == true
				}() == true {
					__antithesis_instrumentation__.Notify(697940)
					item.brandGeneric = "B"
				} else {
					__antithesis_instrumentation__.Notify(697941)
					item.brandGeneric = "G"
				}
				__antithesis_instrumentation__.Notify(697929)

				newSQuantity := sQuantity - item.olQuantity
				if sQuantity < item.olQuantity+10 {
					__antithesis_instrumentation__.Notify(697942)
					newSQuantity += 91
				} else {
					__antithesis_instrumentation__.Notify(697943)
				}
				__antithesis_instrumentation__.Notify(697930)

				newSRemoteCnt := sRemoteCnt
				if item.remoteWarehouse {
					__antithesis_instrumentation__.Notify(697944)
					newSRemoteCnt++
				} else {
					__antithesis_instrumentation__.Notify(697945)
				}
				__antithesis_instrumentation__.Notify(697931)

				sQuantityUpdateCases[i] = fmt.Sprintf("WHEN %s THEN %d", stockIDs[i], newSQuantity)
				sYtdUpdateCases[i] = fmt.Sprintf("WHEN %s THEN %d", stockIDs[i], sYtd+item.olQuantity)
				sOrderCntUpdateCases[i] = fmt.Sprintf("WHEN %s THEN %d", stockIDs[i], sOrderCnt+1)
				sRemoteCntUpdateCases[i] = fmt.Sprintf("WHEN %s THEN %d", stockIDs[i], newSRemoteCnt)
			}
			__antithesis_instrumentation__.Notify(697890)
			if rows.Next() {
				__antithesis_instrumentation__.Notify(697946)
				return errors.New("extra stock row")
			} else {
				__antithesis_instrumentation__.Notify(697947)
			}
			__antithesis_instrumentation__.Notify(697891)
			if err := rows.Err(); err != nil {
				__antithesis_instrumentation__.Notify(697948)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697949)
			}
			__antithesis_instrumentation__.Notify(697892)
			rows.Close()

			if _, err := n.insertOrder.ExecTx(
				ctx, tx,
				d.oID, d.dID, d.wID, d.cID, d.oEntryD.Format("2006-01-02 15:04:05"), d.oOlCnt, allLocal,
			); err != nil {
				__antithesis_instrumentation__.Notify(697950)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697951)
			}
			__antithesis_instrumentation__.Notify(697893)
			if _, err := n.insertNewOrder.ExecTx(
				ctx, tx, d.oID, d.dID, d.wID,
			); err != nil {
				__antithesis_instrumentation__.Notify(697952)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697953)
			}
			__antithesis_instrumentation__.Notify(697894)

			if _, err := tx.Exec(
				ctx,
				fmt.Sprintf(`
					UPDATE stock
					SET
						s_quantity = CASE (s_i_id, s_w_id) %[1]s ELSE crdb_internal.force_error('', 'unknown case') END,
						s_ytd = CASE (s_i_id, s_w_id) %[2]s END,
						s_order_cnt = CASE (s_i_id, s_w_id) %[3]s END,
						s_remote_cnt = CASE (s_i_id, s_w_id) %[4]s END
					WHERE (s_i_id, s_w_id) IN (%[5]s)`,
					strings.Join(sQuantityUpdateCases, " "),
					strings.Join(sYtdUpdateCases, " "),
					strings.Join(sOrderCntUpdateCases, " "),
					strings.Join(sRemoteCntUpdateCases, " "),
					strings.Join(stockIDs, ", "),
				),
			); err != nil {
				__antithesis_instrumentation__.Notify(697954)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697955)
			}
			__antithesis_instrumentation__.Notify(697895)

			olValsStrings := make([]string, d.oOlCnt)
			for i := range d.items {
				__antithesis_instrumentation__.Notify(697956)
				item := &d.items[i]
				item.olAmount = float64(item.olQuantity) * item.iPrice
				d.totalAmount += item.olAmount

				olValsStrings[i] = fmt.Sprintf("(%d,%d,%d,%d,%d,%d,%d,%f,'%s')",
					d.oID,
					d.dID,
					d.wID,
					item.olNumber,
					item.olIID,
					item.olSupplyWID,
					item.olQuantity,
					item.olAmount,
					distInfos[i],
				)
			}
			__antithesis_instrumentation__.Notify(697896)
			if _, err := tx.Exec(
				ctx,
				fmt.Sprintf(`
					INSERT INTO order_line(ol_o_id, ol_d_id, ol_w_id, ol_number, ol_i_id, ol_supply_w_id, ol_quantity, ol_amount, ol_dist_info)
					VALUES %s`,
					strings.Join(olValsStrings, ", "),
				),
			); err != nil {
				__antithesis_instrumentation__.Notify(697957)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697958)
			}
			__antithesis_instrumentation__.Notify(697897)

			d.totalAmount *= (1 - d.cDiscount) * (1 + d.wTax + d.dTax)

			return nil
		})
	__antithesis_instrumentation__.Notify(697859)
	if errors.Is(err, errSimulated) {
		__antithesis_instrumentation__.Notify(697959)
		return d, nil
	} else {
		__antithesis_instrumentation__.Notify(697960)
	}
	__antithesis_instrumentation__.Notify(697860)
	return d, err
}
