package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"golang.org/x/exp/rand"
)

const (
	numItems                 = 100000
	numDistrictsPerWarehouse = 10
	numStockPerWarehouse     = 100000
	numCustomersPerDistrict  = 3000
	numCustomersPerWarehouse = numCustomersPerDistrict * numDistrictsPerWarehouse
	numHistoryPerWarehouse   = numCustomersPerWarehouse
	numOrdersPerDistrict     = numCustomersPerDistrict
	numOrdersPerWarehouse    = numOrdersPerDistrict * numDistrictsPerWarehouse
	numNewOrdersPerDistrict  = 900
	numNewOrdersPerWarehouse = numNewOrdersPerDistrict * numDistrictsPerWarehouse
	minOrderLinesPerOrder    = 5
	maxOrderLinesPerOrder    = 15

	originalString = "ORIGINAL"
	wYtd           = 300000.00
	ytd            = 30000.00
	nextOrderID    = 3001
	creditLimit    = 50000.00
	balance        = -10.00
	ytdPayment     = 10.00
	paymentCount   = 1
	deliveryCount  = 0
)

var (
	middleName = []byte(`OE`)
	goodCredit = []byte("GC")
	badCredit  = []byte("BC")
)

const (
	numWarehousesPerRange = 10
	numItemsPerRange      = 100
)

type generateLocals struct {
	rng       tpccRand
	uuidAlloc uuid.UUID
}

var itemTypes = []*types.T{
	types.Int,
	types.Int,
	types.Bytes,
	types.Float,
	types.Bytes,
}

func (w *tpcc) tpccItemInitialRowBatch(rowIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator) {
	__antithesis_instrumentation__.Notify(697806)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	l.rng.Seed(w.seed + uint64(rowIdx))
	ao := aCharsOffset(l.rng.Intn(len(aCharsAlphabet)))

	iID := rowIdx + 1

	cb.Reset(itemTypes, 1, coldata.StandardColumnFactory)
	cb.ColVec(0).Int64()[0] = int64(iID)
	cb.ColVec(1).Int64()[0] = randInt(l.rng.Rand, 1, 10000)
	cb.ColVec(2).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 14, 24))
	cb.ColVec(3).Float64()[0] = float64(randInt(l.rng.Rand, 100, 10000)) / float64(100)
	cb.ColVec(4).Bytes().Set(0, randOriginalStringInitialDataOnly(&l.rng, &ao, a))
}

func (w *tpcc) tpccItemStats() []workload.JSONStatistic {
	__antithesis_instrumentation__.Notify(697807)
	rowCount := uint64(numItems)

	return []workload.JSONStatistic{
		workload.MakeStat([]string{"i_id"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"i_im_id"}, rowCount, workload.DistinctCount(rowCount, 10000), 0),
		workload.MakeStat([]string{"i_name"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"i_price"}, rowCount, workload.DistinctCount(rowCount, 9901), 0),
		workload.MakeStat([]string{"i_data"}, rowCount, rowCount, 0),
	}
}

var warehouseTypes = []*types.T{
	types.Int,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Float,
	types.Float,
}

func (w *tpcc) tpccWarehouseInitialRowBatch(
	rowIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(697808)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	l.rng.Seed(w.seed + uint64(rowIdx))
	no := numbersOffset(l.rng.Intn(len(numbersAlphabet)))
	lo := lettersOffset(l.rng.Intn(len(lettersAlphabet)))

	wID := rowIdx

	cb.Reset(warehouseTypes, 1, coldata.StandardColumnFactory)
	cb.ColVec(0).Int64()[0] = int64(wID)
	cb.ColVec(1).Bytes().Set(0, []byte(strconv.FormatInt(randInt(l.rng.Rand, 6, 10), 10)))
	cb.ColVec(2).Bytes().Set(0, []byte(strconv.FormatInt(randInt(l.rng.Rand, 10, 20), 10)))
	cb.ColVec(3).Bytes().Set(0, []byte(strconv.FormatInt(randInt(l.rng.Rand, 10, 20), 10)))
	cb.ColVec(4).Bytes().Set(0, []byte(strconv.FormatInt(randInt(l.rng.Rand, 10, 20), 10)))
	cb.ColVec(5).Bytes().Set(0, randStateInitialDataOnly(&l.rng, &lo, a))
	cb.ColVec(6).Bytes().Set(0, randZipInitialDataOnly(&l.rng, &no, a))
	cb.ColVec(7).Float64()[0] = randTax(l.rng.Rand)
	cb.ColVec(8).Float64()[0] = wYtd
}

func (w *tpcc) tpccWarehouseStats() []workload.JSONStatistic {
	__antithesis_instrumentation__.Notify(697809)
	rowCount := uint64(w.warehouses)
	return []workload.JSONStatistic{
		workload.MakeStat([]string{"w_id"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"w_name"}, rowCount, workload.DistinctCount(rowCount, 5), 0),
		workload.MakeStat([]string{"w_street_1"}, rowCount, workload.DistinctCount(rowCount, 11), 0),
		workload.MakeStat([]string{"w_street_2"}, rowCount, workload.DistinctCount(rowCount, 11), 0),
		workload.MakeStat([]string{"w_city"}, rowCount, workload.DistinctCount(rowCount, 11), 0),

		workload.MakeStat([]string{"w_state"}, rowCount, workload.DistinctCount(rowCount, 26*26), 0),

		workload.MakeStat([]string{"w_zip"}, rowCount, workload.DistinctCount(rowCount, 9999), 0),
		workload.MakeStat([]string{"w_tax"}, rowCount, workload.DistinctCount(rowCount, 2001), 0),
		workload.MakeStat([]string{"w_ytd"}, rowCount, 1, 0),
	}
}

var stockTypes = []*types.T{
	types.Int,
	types.Int,
	types.Int,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Int,
	types.Int,
	types.Int,
	types.Bytes,
}

func (w *tpcc) tpccStockInitialRowBatch(rowIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator) {
	__antithesis_instrumentation__.Notify(697810)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	l.rng.Seed(w.seed + uint64(rowIdx))
	ao := aCharsOffset(l.rng.Intn(len(aCharsAlphabet)))

	sID := (rowIdx % numStockPerWarehouse) + 1
	wID := (rowIdx / numStockPerWarehouse)

	cb.Reset(stockTypes, 1, coldata.StandardColumnFactory)
	cb.ColVec(0).Int64()[0] = int64(sID)
	cb.ColVec(1).Int64()[0] = int64(wID)
	cb.ColVec(2).Int64()[0] = randInt(l.rng.Rand, 10, 100)
	cb.ColVec(3).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 24, 24))
	cb.ColVec(4).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 24, 24))
	cb.ColVec(5).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 24, 24))
	cb.ColVec(6).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 24, 24))
	cb.ColVec(7).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 24, 24))
	cb.ColVec(8).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 24, 24))
	cb.ColVec(9).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 24, 24))
	cb.ColVec(10).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 24, 24))
	cb.ColVec(11).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 24, 24))
	cb.ColVec(12).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 24, 24))
	cb.ColVec(13).Int64()[0] = 0
	cb.ColVec(14).Int64()[0] = 0
	cb.ColVec(15).Int64()[0] = 0
	cb.ColVec(16).Bytes().Set(0, randOriginalStringInitialDataOnly(&l.rng, &ao, a))
}

func (w *tpcc) tpccStockStats() []workload.JSONStatistic {
	__antithesis_instrumentation__.Notify(697811)
	rowCount := uint64(w.warehouses * numStockPerWarehouse)

	return []workload.JSONStatistic{
		workload.MakeStat([]string{"s_i_id"}, rowCount, numStockPerWarehouse, 0),
		workload.MakeStat([]string{"s_w_id"}, rowCount, uint64(w.warehouses), 0),
		workload.MakeStat([]string{"s_quantity"}, rowCount, workload.DistinctCount(rowCount, 91), 0),
		workload.MakeStat([]string{"s_dist_01"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"s_dist_02"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"s_dist_03"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"s_dist_04"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"s_dist_05"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"s_dist_06"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"s_dist_07"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"s_dist_08"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"s_dist_09"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"s_dist_10"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"s_ytd"}, rowCount, 1, 0),
		workload.MakeStat([]string{"s_order_cnt"}, rowCount, 1, 0),
		workload.MakeStat([]string{"s_remote_cnt"}, rowCount, 1, 0),
		workload.MakeStat([]string{"s_data"}, rowCount, rowCount, 0),
	}
}

var districtTypes = []*types.T{
	types.Int,
	types.Int,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Float,
	types.Float,
	types.Int,
}

func (w *tpcc) tpccDistrictInitialRowBatch(
	rowIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(697812)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	l.rng.Seed(w.seed + uint64(rowIdx))
	ao := aCharsOffset(l.rng.Intn(len(aCharsAlphabet)))
	no := numbersOffset(l.rng.Intn(len(numbersAlphabet)))
	lo := lettersOffset(l.rng.Intn(len(lettersAlphabet)))

	dID := (rowIdx % numDistrictsPerWarehouse) + 1
	wID := (rowIdx / numDistrictsPerWarehouse)

	cb.Reset(districtTypes, 1, coldata.StandardColumnFactory)
	cb.ColVec(0).Int64()[0] = int64(dID)
	cb.ColVec(1).Int64()[0] = int64(wID)
	cb.ColVec(2).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 6, 10))
	cb.ColVec(3).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 10, 20))
	cb.ColVec(4).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 10, 20))
	cb.ColVec(5).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 10, 20))
	cb.ColVec(6).Bytes().Set(0, randStateInitialDataOnly(&l.rng, &lo, a))
	cb.ColVec(7).Bytes().Set(0, randZipInitialDataOnly(&l.rng, &no, a))
	cb.ColVec(8).Float64()[0] = randTax(l.rng.Rand)
	cb.ColVec(9).Float64()[0] = ytd
	cb.ColVec(10).Int64()[0] = nextOrderID
}

func (w *tpcc) tpccDistrictStats() []workload.JSONStatistic {
	__antithesis_instrumentation__.Notify(697813)
	rowCount := uint64(w.warehouses * numDistrictsPerWarehouse)

	return []workload.JSONStatistic{
		workload.MakeStat([]string{"d_id"}, rowCount, numDistrictsPerWarehouse, 0),
		workload.MakeStat([]string{"d_w_id"}, rowCount, uint64(w.warehouses), 0),
		workload.MakeStat([]string{"d_name"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"d_street_1"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"d_street_2"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"d_city"}, rowCount, rowCount, 0),

		workload.MakeStat([]string{"d_state"}, rowCount, workload.DistinctCount(rowCount, 26*26), 0),

		workload.MakeStat([]string{"d_zip"}, rowCount, workload.DistinctCount(rowCount, 9999), 0),
		workload.MakeStat([]string{"d_tax"}, rowCount, workload.DistinctCount(rowCount, 2001), 0),
		workload.MakeStat([]string{"d_ytd"}, rowCount, 1, 0),
		workload.MakeStat([]string{"d_next_o_id"}, rowCount, 1, 0),
	}
}

var customerTypes = []*types.T{
	types.Int,
	types.Int,
	types.Int,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Float,
	types.Float,
	types.Float,
	types.Float,
	types.Int,
	types.Int,
	types.Bytes,
}

func (w *tpcc) tpccCustomerInitialRowBatch(
	rowIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(697814)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	l.rng.Seed(w.seed + uint64(rowIdx))
	ao := aCharsOffset(l.rng.Intn(len(aCharsAlphabet)))
	no := numbersOffset(l.rng.Intn(len(numbersAlphabet)))
	lo := lettersOffset(l.rng.Intn(len(lettersAlphabet)))

	cID := (rowIdx % numCustomersPerDistrict) + 1
	dID := ((rowIdx / numCustomersPerDistrict) % numDistrictsPerWarehouse) + 1
	wID := (rowIdx / numCustomersPerWarehouse)

	credit := goodCredit
	if l.rng.Intn(9) == 0 {
		__antithesis_instrumentation__.Notify(697817)

		credit = badCredit
	} else {
		__antithesis_instrumentation__.Notify(697818)
	}
	__antithesis_instrumentation__.Notify(697815)
	var lastName []byte

	if cID <= 1000 {
		__antithesis_instrumentation__.Notify(697819)
		lastName = randCLastSyllables(cID-1, a)
	} else {
		__antithesis_instrumentation__.Notify(697820)
		lastName = w.randCLast(l.rng.Rand, a)
	}
	__antithesis_instrumentation__.Notify(697816)

	cb.Reset(customerTypes, 1, coldata.StandardColumnFactory)
	cb.ColVec(0).Int64()[0] = int64(cID)
	cb.ColVec(1).Int64()[0] = int64(dID)
	cb.ColVec(2).Int64()[0] = int64(wID)
	cb.ColVec(3).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 8, 16))
	cb.ColVec(4).Bytes().Set(0, middleName)
	cb.ColVec(5).Bytes().Set(0, lastName)
	cb.ColVec(6).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 10, 20))
	cb.ColVec(7).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 10, 20))
	cb.ColVec(8).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 10, 20))
	cb.ColVec(9).Bytes().Set(0, randStateInitialDataOnly(&l.rng, &lo, a))
	cb.ColVec(10).Bytes().Set(0, randZipInitialDataOnly(&l.rng, &no, a))
	cb.ColVec(11).Bytes().Set(0, randNStringInitialDataOnly(&l.rng, &no, a, 16, 16))
	cb.ColVec(12).Bytes().Set(0, w.nowString)
	cb.ColVec(13).Bytes().Set(0, credit)
	cb.ColVec(14).Float64()[0] = creditLimit
	cb.ColVec(15).Float64()[0] = float64(randInt(l.rng.Rand, 0, 5000)) / float64(10000.0)
	cb.ColVec(16).Float64()[0] = balance
	cb.ColVec(17).Float64()[0] = ytdPayment
	cb.ColVec(18).Int64()[0] = paymentCount
	cb.ColVec(19).Int64()[0] = deliveryCount
	cb.ColVec(20).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 300, 500))
}

func (w *tpcc) tpccCustomerStats() []workload.JSONStatistic {
	__antithesis_instrumentation__.Notify(697821)
	rowCount := uint64(w.warehouses * numCustomersPerWarehouse)

	return []workload.JSONStatistic{
		workload.MakeStat([]string{"c_id"}, rowCount, numCustomersPerDistrict, 0),
		workload.MakeStat([]string{"c_d_id"}, rowCount, numDistrictsPerWarehouse, 0),
		workload.MakeStat([]string{"c_w_id"}, rowCount, uint64(w.warehouses), 0),
		workload.MakeStat([]string{"c_first"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"c_middle"}, rowCount, 1, 0),

		workload.MakeStat([]string{"c_last"}, rowCount, workload.DistinctCount(rowCount, 10*10*10), 0),
		workload.MakeStat([]string{"c_street_1"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"c_street_2"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"c_city"}, rowCount, rowCount, 0),

		workload.MakeStat([]string{"c_state"}, rowCount, workload.DistinctCount(rowCount, 26*26), 0),

		workload.MakeStat([]string{"c_zip"}, rowCount, workload.DistinctCount(rowCount, 9999), 0),
		workload.MakeStat([]string{"c_phone"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"c_since"}, rowCount, 1, 0),

		workload.MakeStat([]string{"c_credit"}, rowCount, 2, 0),
		workload.MakeStat([]string{"c_credit_lim"}, rowCount, 1, 0),
		workload.MakeStat([]string{"c_discount"}, rowCount, workload.DistinctCount(rowCount, 5001), 0),
		workload.MakeStat([]string{"c_balance"}, rowCount, 1, 0),
		workload.MakeStat([]string{"c_ytd_payment"}, rowCount, 1, 0),
		workload.MakeStat([]string{"c_payment_cnt"}, rowCount, 1, 0),
		workload.MakeStat([]string{"c_delivery_cnt"}, rowCount, 1, 0),
		workload.MakeStat([]string{"c_data"}, rowCount, rowCount, 0),
	}
}

var historyTypes = []*types.T{
	types.Bytes,
	types.Int,
	types.Int,
	types.Int,
	types.Int,
	types.Int,
	types.Bytes,
	types.Float,
	types.Bytes,
}

func (w *tpcc) tpccHistoryInitialRowBatch(rowIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator) {
	__antithesis_instrumentation__.Notify(697822)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	l.rng.Seed(w.seed + uint64(rowIdx))
	ao := aCharsOffset(l.rng.Intn(len(aCharsAlphabet)))

	historyRowCount := numHistoryPerWarehouse * w.warehouses
	l.uuidAlloc.DeterministicV4(uint64(rowIdx), uint64(historyRowCount))
	var rowID []byte
	*a, rowID = a.Alloc(36, 0)
	l.uuidAlloc.StringBytes(rowID)

	cID := (rowIdx % numCustomersPerDistrict) + 1
	dID := ((rowIdx / numCustomersPerDistrict) % numDistrictsPerWarehouse) + 1
	wID := (rowIdx / numCustomersPerWarehouse)

	cb.Reset(historyTypes, 1, coldata.StandardColumnFactory)
	cb.ColVec(0).Bytes().Set(0, rowID)
	cb.ColVec(1).Int64()[0] = int64(cID)
	cb.ColVec(2).Int64()[0] = int64(dID)
	cb.ColVec(3).Int64()[0] = int64(wID)
	cb.ColVec(4).Int64()[0] = int64(dID)
	cb.ColVec(5).Int64()[0] = int64(wID)
	cb.ColVec(6).Bytes().Set(0, w.nowString)
	cb.ColVec(7).Float64()[0] = 10.00
	cb.ColVec(8).Bytes().Set(0, randAStringInitialDataOnly(&l.rng, &ao, a, 12, 24))
}

func (w *tpcc) tpccHistoryStats() []workload.JSONStatistic {
	__antithesis_instrumentation__.Notify(697823)
	rowCount := uint64(w.warehouses * numCustomersPerWarehouse)
	return []workload.JSONStatistic{
		workload.MakeStat([]string{"rowid"}, rowCount, rowCount, 0),
		workload.MakeStat([]string{"h_c_id"}, rowCount, numCustomersPerDistrict, 0),
		workload.MakeStat([]string{"h_c_d_id"}, rowCount, numDistrictsPerWarehouse, 0),
		workload.MakeStat([]string{"h_c_w_id"}, rowCount, uint64(w.warehouses), 0),
		workload.MakeStat([]string{"h_d_id"}, rowCount, numDistrictsPerWarehouse, 0),
		workload.MakeStat([]string{"h_w_id"}, rowCount, uint64(w.warehouses), 0),
		workload.MakeStat([]string{"h_date"}, rowCount, 1, 0),
		workload.MakeStat([]string{"h_amount"}, rowCount, 1, 0),

		workload.MakeStat([]string{"h_data"}, rowCount, rowCount, 0),
	}
}

var orderTypes = []*types.T{
	types.Int,
	types.Int,
	types.Int,
	types.Int,
	types.Bytes,
	types.Int,
	types.Int,
	types.Int,
}

func (w *tpcc) tpccOrderInitialRowBatch(rowIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator) {
	__antithesis_instrumentation__.Notify(697824)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)

	l.rng.Seed(w.seed + uint64(rowIdx))
	numOrderLines := randInt(l.rng.Rand, minOrderLinesPerOrder, maxOrderLinesPerOrder)

	oID := (rowIdx % numOrdersPerDistrict) + 1
	dID := ((rowIdx / numOrdersPerDistrict) % numDistrictsPerWarehouse) + 1
	wID := (rowIdx / numOrdersPerWarehouse)

	var cID int
	{
		__antithesis_instrumentation__.Notify(697828)

		w.randomCIDsCache.Lock()
		if w.randomCIDsCache.values == nil {
			__antithesis_instrumentation__.Notify(697831)
			w.randomCIDsCache.values = make([][]int, numDistrictsPerWarehouse*w.warehouses+1)
		} else {
			__antithesis_instrumentation__.Notify(697832)
		}
		__antithesis_instrumentation__.Notify(697829)
		if w.randomCIDsCache.values[dID] == nil {
			__antithesis_instrumentation__.Notify(697833)

			w.randomCIDsCache.values[dID] = make([]int, numCustomersPerDistrict)
			for i, cID := range rand.New(rand.NewSource(uint64(dID))).Perm(numCustomersPerDistrict) {
				__antithesis_instrumentation__.Notify(697834)
				w.randomCIDsCache.values[dID][i] = cID + 1
			}
		} else {
			__antithesis_instrumentation__.Notify(697835)
		}
		__antithesis_instrumentation__.Notify(697830)
		cID = w.randomCIDsCache.values[dID][oID-1]
		w.randomCIDsCache.Unlock()
	}
	__antithesis_instrumentation__.Notify(697825)

	var carrierSet bool
	var carrierID int64
	if oID < 2101 {
		__antithesis_instrumentation__.Notify(697836)
		carrierSet = true
		carrierID = randInt(l.rng.Rand, 1, 10)
	} else {
		__antithesis_instrumentation__.Notify(697837)
	}
	__antithesis_instrumentation__.Notify(697826)

	cb.Reset(orderTypes, 1, coldata.StandardColumnFactory)
	cb.ColVec(0).Int64()[0] = int64(oID)
	cb.ColVec(1).Int64()[0] = int64(dID)
	cb.ColVec(2).Int64()[0] = int64(wID)
	cb.ColVec(3).Int64()[0] = int64(cID)
	cb.ColVec(4).Bytes().Set(0, w.nowString)
	cb.ColVec(5).Nulls().UnsetNulls()
	if carrierSet {
		__antithesis_instrumentation__.Notify(697838)
		cb.ColVec(5).Int64()[0] = carrierID
	} else {
		__antithesis_instrumentation__.Notify(697839)
		cb.ColVec(5).Nulls().SetNull(0)
		cb.ColVec(5).Int64()[0] = 0
	}
	__antithesis_instrumentation__.Notify(697827)
	cb.ColVec(6).Int64()[0] = numOrderLines
	cb.ColVec(7).Int64()[0] = 1
}

func (w *tpcc) tpccOrderStats() []workload.JSONStatistic {
	__antithesis_instrumentation__.Notify(697840)
	rowCount := uint64(w.warehouses * numOrdersPerWarehouse)
	return []workload.JSONStatistic{
		workload.MakeStat([]string{"o_id"}, rowCount, numOrdersPerDistrict, 0),
		workload.MakeStat([]string{"o_d_id"}, rowCount, numDistrictsPerWarehouse, 0),
		workload.MakeStat([]string{"o_w_id"}, rowCount, uint64(w.warehouses), 0),
		workload.MakeStat([]string{"o_c_id"}, rowCount, numCustomersPerDistrict, 0),
		workload.MakeStat([]string{"o_entry_d"}, rowCount, 1, 0),

		workload.MakeStat([]string{"o_carrier_id"}, rowCount, workload.DistinctCount(rowCount, 10),
			uint64(w.warehouses)*numDistrictsPerWarehouse*(numOrdersPerDistrict-2100),
		),
		workload.MakeStat([]string{"o_ol_cnt"}, rowCount, workload.DistinctCount(rowCount, 11), 0),
		workload.MakeStat([]string{"o_all_local"}, rowCount, 1, 0),
	}
}

var newOrderTypes = []*types.T{
	types.Int,
	types.Int,
	types.Int,
}

func (w *tpcc) tpccNewOrderInitialRowBatch(
	rowIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(697841)

	const firstNewOrderOffset = numOrdersPerDistrict - numNewOrdersPerDistrict
	oID := (rowIdx % numNewOrdersPerDistrict) + firstNewOrderOffset + 1
	dID := ((rowIdx / numNewOrdersPerDistrict) % numDistrictsPerWarehouse) + 1
	wID := (rowIdx / numNewOrdersPerWarehouse)

	cb.Reset(newOrderTypes, 1, coldata.StandardColumnFactory)
	cb.ColVec(0).Int64()[0] = int64(oID)
	cb.ColVec(1).Int64()[0] = int64(dID)
	cb.ColVec(2).Int64()[0] = int64(wID)
}

func (w *tpcc) tpccNewOrderStats() []workload.JSONStatistic {
	__antithesis_instrumentation__.Notify(697842)
	rowCount := uint64(w.warehouses * numNewOrdersPerWarehouse)
	return []workload.JSONStatistic{
		workload.MakeStat([]string{"no_o_id"}, rowCount, numNewOrdersPerDistrict, 0),
		workload.MakeStat([]string{"no_d_id"}, rowCount, numDistrictsPerWarehouse, 0),
		workload.MakeStat([]string{"no_w_id"}, rowCount, uint64(w.warehouses), 0),
	}
}

var orderLineTypes = []*types.T{
	types.Int,
	types.Int,
	types.Int,
	types.Int,
	types.Int,
	types.Int,
	types.Bytes,
	types.Int,
	types.Float,
	types.Bytes,
}

func (w *tpcc) tpccOrderLineInitialRowBatch(
	orderRowIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(697843)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)

	l.rng.Seed(w.seed + uint64(orderRowIdx))
	numOrderLines := int(randInt(l.rng.Rand, minOrderLinesPerOrder, maxOrderLinesPerOrder))

	oID := (orderRowIdx % numOrdersPerDistrict) + 1
	dID := ((orderRowIdx / numOrdersPerDistrict) % numDistrictsPerWarehouse) + 1
	wID := (orderRowIdx / numOrdersPerWarehouse)

	ao := aCharsOffset(l.rng.Intn(len(aCharsAlphabet)))
	cb.Reset(orderLineTypes, numOrderLines, coldata.StandardColumnFactory)
	olOIDCol := cb.ColVec(0).Int64()
	olDIDCol := cb.ColVec(1).Int64()
	olWIDCol := cb.ColVec(2).Int64()
	olNumberCol := cb.ColVec(3).Int64()
	olIIDCol := cb.ColVec(4).Int64()
	olSupplyWIDCol := cb.ColVec(5).Int64()
	olDeliveryD := cb.ColVec(6)
	olDeliveryD.Nulls().UnsetNulls()
	olDeliveryDCol := olDeliveryD.Bytes()
	olQuantityCol := cb.ColVec(7).Int64()
	olAmountCol := cb.ColVec(8).Float64()
	olDistInfoCol := cb.ColVec(9).Bytes()

	olDeliveryDCol.Reset()
	olDistInfoCol.Reset()
	for rowIdx := 0; rowIdx < numOrderLines; rowIdx++ {
		__antithesis_instrumentation__.Notify(697844)
		olNumber := rowIdx + 1

		var amount float64
		var deliveryDSet bool
		var deliveryD []byte
		if oID < 2101 {
			__antithesis_instrumentation__.Notify(697847)
			amount = 0
			deliveryDSet = true
			deliveryD = w.nowString
		} else {
			__antithesis_instrumentation__.Notify(697848)
			amount = float64(randInt(l.rng.Rand, 1, 999999)) / 100.0
		}
		__antithesis_instrumentation__.Notify(697845)

		olOIDCol[rowIdx] = int64(oID)
		olDIDCol[rowIdx] = int64(dID)
		olWIDCol[rowIdx] = int64(wID)
		olNumberCol[rowIdx] = int64(olNumber)
		olIIDCol[rowIdx] = randInt(l.rng.Rand, 1, 100000)
		olSupplyWIDCol[rowIdx] = int64(wID)
		if deliveryDSet {
			__antithesis_instrumentation__.Notify(697849)
			olDeliveryDCol.Set(rowIdx, deliveryD)
		} else {
			__antithesis_instrumentation__.Notify(697850)
			olDeliveryD.Nulls().SetNull(rowIdx)
			olDeliveryDCol.Set(rowIdx, nil)
		}
		__antithesis_instrumentation__.Notify(697846)
		olQuantityCol[rowIdx] = 5
		olAmountCol[rowIdx] = amount
		olDistInfoCol.Set(rowIdx, randAStringInitialDataOnly(&l.rng, &ao, a, 24, 24))
	}
}

func (w *tpcc) tpccOrderLineStats() []workload.JSONStatistic {
	__antithesis_instrumentation__.Notify(697851)
	averageOrderLines := float64(maxOrderLinesPerOrder+minOrderLinesPerOrder) / 2
	rowCount := uint64(int64(float64(w.warehouses) * numOrdersPerWarehouse * averageOrderLines))
	deliveryIncomplete := uint64(int64(
		float64(w.warehouses) * numDistrictsPerWarehouse * (numOrdersPerDistrict - 2100) * averageOrderLines,
	))
	return []workload.JSONStatistic{
		workload.MakeStat([]string{"ol_o_id"}, rowCount, numOrdersPerDistrict, 0),
		workload.MakeStat([]string{"ol_d_id"}, rowCount, numDistrictsPerWarehouse, 0),
		workload.MakeStat([]string{"ol_w_id"}, rowCount, uint64(w.warehouses), 0),
		workload.MakeStat([]string{"ol_number"}, rowCount, maxOrderLinesPerOrder, 0),
		workload.MakeStat([]string{"ol_i_id"}, rowCount, workload.DistinctCount(rowCount, 100000), 0),
		workload.MakeStat([]string{"ol_supply_w_id"}, rowCount, uint64(w.warehouses), 0),
		workload.MakeStat([]string{"ol_delivery_d"}, rowCount, 1, deliveryIncomplete),
		workload.MakeStat([]string{"ol_quantity"}, rowCount, 1, 0),

		workload.MakeStat(
			[]string{"ol_amount"}, rowCount, workload.DistinctCount(deliveryIncomplete, 999999)+1, 0,
		),

		workload.MakeStat([]string{"ol_dist_info"}, rowCount, rowCount, 0),
	}
}
