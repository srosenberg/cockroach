package tpch

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"golang.org/x/exp/rand"
)

var regionNames = [...]string{`AFRICA`, `AMERICA`, `ASIA`, `EUROPE`, `MIDDLE EAST`}
var nations = [...]struct {
	name      string
	regionKey int
}{
	{name: `ALGERIA`, regionKey: 0},
	{name: `ARGENTINA`, regionKey: 1},
	{name: `BRAZIL`, regionKey: 1},
	{name: `CANADA`, regionKey: 1},
	{name: `EGYPT`, regionKey: 4},
	{name: `ETHIOPIA`, regionKey: 0},
	{name: `FRANCE`, regionKey: 3},
	{name: `GERMANY`, regionKey: 3},
	{name: `INDIA`, regionKey: 2},
	{name: `INDONESIA`, regionKey: 2},
	{name: `IRAN`, regionKey: 4},
	{name: `IRAQ`, regionKey: 4},
	{name: `JAPAN`, regionKey: 2},
	{name: `JORDAN`, regionKey: 4},
	{name: `KENYA`, regionKey: 0},
	{name: `MOROCCO`, regionKey: 0},
	{name: `MOZAMBIQUE`, regionKey: 0},
	{name: `PERU`, regionKey: 1},
	{name: `CHINA`, regionKey: 2},
	{name: `ROMANIA`, regionKey: 3},
	{name: `SAUDI ARABIA`, regionKey: 4},
	{name: `VIETNAM`, regionKey: 2},
	{name: `RUSSIA`, regionKey: 3},
	{name: `UNITED KINGDOM`, regionKey: 3},
	{name: `UNITED STATES`, regionKey: 1},
}

var regionTypes = []*types.T{
	types.Int2,
	types.Bytes,
	types.Bytes,
}

func (w *tpch) tpchRegionInitialRowBatch(
	batchIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(698747)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	rng := l.rng
	rng.Seed(w.seed + uint64(batchIdx))

	regionKey := batchIdx
	cb.Reset(regionTypes, 1, coldata.StandardColumnFactory)
	cb.ColVec(0).Int16()[0] = int16(regionKey)
	cb.ColVec(1).Bytes().Set(0, []byte(regionNames[regionKey]))
	cb.ColVec(2).Bytes().Set(0, w.textPool.randString(rng, 31, 115))
}

var nationTypes = []*types.T{
	types.Int2,
	types.Bytes,
	types.Int2,
	types.Bytes,
}

func (w *tpch) tpchNationInitialRowBatch(
	batchIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(698748)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	rng := l.rng
	rng.Seed(w.seed + uint64(batchIdx))

	nationKey := batchIdx
	nation := nations[nationKey]
	cb.Reset(nationTypes, 1, coldata.StandardColumnFactory)
	cb.ColVec(0).Int16()[0] = int16(nationKey)
	cb.ColVec(1).Bytes().Set(0, []byte(nation.name))
	cb.ColVec(2).Int16()[0] = int16(nation.regionKey)
	cb.ColVec(3).Bytes().Set(0, w.textPool.randString(rng, 31, 115))
}

var supplierTypes = []*types.T{
	types.Int,
	types.Bytes,
	types.Bytes,
	types.Int2,
	types.Bytes,
	types.Float,
	types.Bytes,
}

func (w *tpch) tpchSupplierInitialRowBatch(
	batchIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(698749)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	rng := l.rng
	rng.Seed(w.seed + uint64(batchIdx))

	suppKey := int64(batchIdx) + 1
	nationKey := int16(randInt(rng, 0, 24))
	cb.Reset(supplierTypes, 1, coldata.StandardColumnFactory)
	cb.ColVec(0).Int64()[0] = suppKey
	cb.ColVec(1).Bytes().Set(0, supplierName(a, suppKey))
	cb.ColVec(2).Bytes().Set(0, randVString(rng, a, 10, 40))
	cb.ColVec(3).Int16()[0] = nationKey
	cb.ColVec(4).Bytes().Set(0, randPhone(rng, a, nationKey))
	cb.ColVec(5).Float64()[0] = float64(randFloat(rng, -99999, 999999, 100))

	cb.ColVec(6).Bytes().Set(0, w.textPool.randString(rng, 25, 100))
}

var partTypes = []*types.T{
	types.Int,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Int2,
	types.Bytes,
	types.Float,
	types.Bytes,
}

func makeRetailPriceFromPartKey(partKey int) float32 {
	__antithesis_instrumentation__.Notify(698750)
	return float32(90000+((partKey/10)%20001)+100*(partKey%1000)) / 100
}

func (w *tpch) tpchPartInitialRowBatch(batchIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator) {
	__antithesis_instrumentation__.Notify(698751)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	rng := l.rng
	rng.Seed(w.seed + uint64(batchIdx))

	partKey := batchIdx + 1
	cb.Reset(partTypes, 1, coldata.StandardColumnFactory)

	cb.ColVec(0).Int64()[0] = int64(partKey)

	cb.ColVec(1).Bytes().Set(0, randPartName(rng, l.namePerm, a))
	m, mfgr := randMfgr(rng, a)

	cb.ColVec(2).Bytes().Set(0, mfgr)

	cb.ColVec(3).Bytes().Set(0, randBrand(rng, a, m))

	cb.ColVec(4).Bytes().Set(0, randType(rng, a))

	cb.ColVec(5).Int16()[0] = int16(randInt(rng, 1, 50))

	cb.ColVec(6).Bytes().Set(0, randContainer(rng, a))

	cb.ColVec(7).Float64()[0] = float64(makeRetailPriceFromPartKey(partKey))

	cb.ColVec(8).Bytes().Set(0, w.textPool.randString(rng, 5, 22))
}

var partSuppTypes = []*types.T{
	types.Int,
	types.Int,
	types.Int2,
	types.Float,
	types.Bytes,
}

func (w *tpch) tpchPartSuppInitialRowBatch(
	batchIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(698752)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	rng := l.rng
	rng.Seed(w.seed + uint64(batchIdx))

	partKey := batchIdx + 1
	cb.Reset(partSuppTypes, numPartSuppPerPart, coldata.StandardColumnFactory)

	partKeyCol := cb.ColVec(0).Int64()
	suppKeyCol := cb.ColVec(1).Int64()
	availQtyCol := cb.ColVec(2).Int16()
	supplyCostCol := cb.ColVec(3).Float64()
	commentCol := cb.ColVec(4).Bytes()

	for i := 0; i < numPartSuppPerPart; i++ {
		__antithesis_instrumentation__.Notify(698753)

		partKeyCol[i] = int64(partKey)

		s := w.scaleFactor * 10000
		suppKeyCol[i] = int64((partKey+(i*((s/numPartSuppPerPart)+(partKey-1)/s)))%s + 1)

		availQtyCol[i] = int16(randInt(rng, 1, 9999))

		supplyCostCol[i] = float64(randFloat(rng, 1, 1000, 100))

		commentCol.Set(i, w.textPool.randString(rng, 49, 198))
	}
}

var customerTypes = []*types.T{
	types.Int,
	types.Bytes,
	types.Bytes,
	types.Int2,
	types.Bytes,
	types.Float,
	types.Bytes,
	types.Bytes,
}

func (w *tpch) tpchCustomerInitialRowBatch(
	batchIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(698754)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	rng := l.rng
	rng.Seed(w.seed + uint64(batchIdx))

	custKey := int64(batchIdx) + 1
	cb.Reset(customerTypes, 1, coldata.StandardColumnFactory)

	cb.ColVec(0).Int64()[0] = custKey

	cb.ColVec(1).Bytes().Set(0, customerName(a, custKey))

	cb.ColVec(2).Bytes().Set(0, randVString(rng, a, 10, 40))
	nationKey := int16(randInt(rng, 0, 24))

	cb.ColVec(3).Int16()[0] = nationKey

	cb.ColVec(4).Bytes().Set(0, randPhone(rng, a, nationKey))

	cb.ColVec(5).Float64()[0] = float64(randFloat(rng, -99999, 999999, 100))

	cb.ColVec(6).Bytes().Set(0, randSegment(rng))

	cb.ColVec(7).Bytes().Set(0, w.textPool.randString(rng, 29, 116))
}

const sparseBits = 2
const sparseKeep = 3

var (
	startDate   = time.Date(1992, 1, 1, 0, 0, 0, 0, time.UTC)
	currentDate = time.Date(1995, 6, 17, 0, 0, 0, 0, time.UTC)
	endDate     = time.Date(1998, 12, 31, 0, 0, 0, 0, time.UTC)

	startDateDays   int64
	currentDateDays int64
	endDateDays     int64
)

func init() {
	var d pgdate.Date
	var err error
	d, err = pgdate.MakeDateFromTime(startDate)
	if err != nil {
		panic(err)
	}
	startDateDays = d.UnixEpochDaysWithOrig()
	d, err = pgdate.MakeDateFromTime(currentDate)
	if err != nil {
		panic(err)
	}
	currentDateDays = d.UnixEpochDaysWithOrig()
	d, err = pgdate.MakeDateFromTime(endDate)
	if err != nil {
		panic(err)
	}
	endDateDays = d.UnixEpochDaysWithOrig()
}

func getOrderKey(orderIdx int) int {
	__antithesis_instrumentation__.Notify(698755)

	i := orderIdx + 1

	lowBits := i & ((1 << sparseKeep) - 1)
	i = i >> sparseKeep
	i = i << sparseBits
	i = i << sparseKeep
	i += lowBits
	return i
}

type orderSharedRandomData struct {
	nOrders    int
	orderDate  int
	partKeys   []int
	shipDates  []int64
	quantities []float32
	discount   []float32
	tax        []float32

	allO bool
	allF bool
}

var ordersTypes = []*types.T{
	types.Int,
	types.Int,
	types.Bytes,
	types.Float,
	types.Int,
	types.Bytes,
	types.Bytes,
	types.Int2,
	types.Bytes,
}

func populateSharedData(rng *rand.Rand, seed uint64, sf int, data *orderSharedRandomData) {
	__antithesis_instrumentation__.Notify(698756)

	rng.Seed(seed)

	data.nOrders = randInt(rng, 1, 7)
	data.orderDate = randInt(rng, int(startDateDays), int(endDateDays-151))
	data.partKeys = data.partKeys[:data.nOrders]
	data.shipDates = data.shipDates[:data.nOrders]
	data.quantities = data.quantities[:data.nOrders]
	data.discount = data.discount[:data.nOrders]
	data.tax = data.tax[:data.nOrders]

	data.allF = true
	data.allO = true

	for i := 0; i < data.nOrders; i++ {
		__antithesis_instrumentation__.Notify(698757)
		shipDate := int64(data.orderDate + randInt(rng, 1, 121))
		data.shipDates[i] = shipDate
		if shipDate > currentDateDays {
			__antithesis_instrumentation__.Notify(698759)
			data.allF = false
		} else {
			__antithesis_instrumentation__.Notify(698760)
			data.allO = false
		}
		__antithesis_instrumentation__.Notify(698758)
		data.partKeys[i] = randInt(rng, 1, sf*numPartPerSF)
		data.quantities[i] = float32(randInt(rng, 1, 50))
		data.discount[i] = randFloat(rng, 0, 10, 100)
		data.tax[i] = randFloat(rng, 0, 8, 100)
	}
}

func (w *tpch) tpchOrdersInitialRowBatch(
	batchIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(698761)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	rng := l.rng

	cb.Reset(ordersTypes, numOrderPerCustomer, coldata.StandardColumnFactory)

	orderKeyCol := cb.ColVec(0).Int64()
	custKeyCol := cb.ColVec(1).Int64()
	orderStatusCol := cb.ColVec(2).Bytes()
	totalPriceCol := cb.ColVec(3).Float64()
	orderDateCol := cb.ColVec(4).Int64()
	orderPriorityCol := cb.ColVec(5).Bytes()
	clerkCol := cb.ColVec(6).Bytes()
	shipPriorityCol := cb.ColVec(7).Int16()
	commentCol := cb.ColVec(8).Bytes()

	orderStartIdx := numOrderPerCustomer * batchIdx
	for i := 0; i < numOrderPerCustomer; i++ {
		__antithesis_instrumentation__.Notify(698762)
		populateSharedData(rng, w.seed+uint64(orderStartIdx+i), w.scaleFactor, l.orderData)

		orderKeyCol[i] = int64(getOrderKey(orderStartIdx + i))

		numCust := w.scaleFactor * numCustomerPerSF
		custKeyCol[i] = int64((randInt(rng, 1, w.scaleFactor*(numCustomerPerSF/3))*3 + rng.Intn(2) + 1) % numCust)

		if l.orderData.allF {
			__antithesis_instrumentation__.Notify(698765)
			orderStatusCol.Set(i, []byte("F"))
		} else {
			__antithesis_instrumentation__.Notify(698766)
			if l.orderData.allO {
				__antithesis_instrumentation__.Notify(698767)
				orderStatusCol.Set(i, []byte("O"))
			} else {
				__antithesis_instrumentation__.Notify(698768)
				orderStatusCol.Set(i, []byte("P"))
			}
		}
		__antithesis_instrumentation__.Notify(698763)
		totalPrice := float32(0)
		for j := 0; j < l.orderData.nOrders; j++ {
			__antithesis_instrumentation__.Notify(698769)
			ep := l.orderData.quantities[j] * makeRetailPriceFromPartKey(l.orderData.partKeys[j])

			totalPrice += float32(ep * (1 + l.orderData.tax[j]) * (1 - l.orderData.discount[j]))
		}
		__antithesis_instrumentation__.Notify(698764)

		totalPriceCol[i] = float64(totalPrice)

		orderDateCol[i] = int64(l.orderData.orderDate)

		orderPriorityCol.Set(i, randPriority(rng))

		clerkCol.Set(i, randClerk(rng, a, w.scaleFactor))

		shipPriorityCol[i] = 0

		commentCol.Set(i, w.textPool.randString(rng, 19, 78))
	}
}

var lineItemTypes = []*types.T{
	types.Int,
	types.Int,
	types.Int,
	types.Int2,
	types.Float,
	types.Float,
	types.Float,
	types.Float,
	types.Bytes,
	types.Bytes,
	types.Int,
	types.Int,
	types.Int,
	types.Bytes,
	types.Bytes,
	types.Bytes,
}

func (w *tpch) tpchLineItemInitialRowBatch(
	batchIdx int, cb coldata.Batch, a *bufalloc.ByteAllocator,
) {
	__antithesis_instrumentation__.Notify(698770)
	l := w.localsPool.Get().(*generateLocals)
	defer w.localsPool.Put(l)
	rng := l.rng

	cb.Reset(lineItemTypes, numOrderPerCustomer*7, coldata.StandardColumnFactory)

	orderKeyCol := cb.ColVec(0).Int64()
	partKeyCol := cb.ColVec(1).Int64()
	suppKeyCol := cb.ColVec(2).Int64()
	lineNumberCol := cb.ColVec(3).Int16()
	quantityCol := cb.ColVec(4).Float64()
	extendedPriceCol := cb.ColVec(5).Float64()
	discountCol := cb.ColVec(6).Float64()
	taxCol := cb.ColVec(7).Float64()
	returnFlagCol := cb.ColVec(8).Bytes()
	lineStatusCol := cb.ColVec(9).Bytes()
	shipDateCol := cb.ColVec(10).Int64()
	commitDateCol := cb.ColVec(11).Int64()
	receiptDateCol := cb.ColVec(12).Int64()
	shipInstructCol := cb.ColVec(13).Bytes()
	shipModeCol := cb.ColVec(14).Bytes()
	commentCol := cb.ColVec(15).Bytes()

	orderStartIdx := numOrderPerCustomer * batchIdx
	s := w.scaleFactor * 10000
	offset := 0
	for i := 0; i < numOrderPerCustomer; i++ {
		__antithesis_instrumentation__.Notify(698772)
		populateSharedData(rng, w.seed+uint64(orderStartIdx+i), w.scaleFactor, l.orderData)

		orderKey := int64(getOrderKey(orderStartIdx + i))
		for j := 0; j < l.orderData.nOrders; j++ {
			__antithesis_instrumentation__.Notify(698774)
			idx := offset + j

			orderKeyCol[idx] = orderKey
			partKey := l.orderData.partKeys[j]

			partKeyCol[idx] = int64(partKey)

			suppKey := (partKey+(randInt(rng, 0, 3)*((s/4)+(partKey-1)/s)))%s + 1
			suppKeyCol[idx] = int64(suppKey)

			lineNumberCol[idx] = int16(j)

			quantityCol[idx] = float64(l.orderData.quantities[j])

			extendedPriceCol[idx] = float64(l.orderData.quantities[j] * makeRetailPriceFromPartKey(partKey))

			discountCol[idx] = float64(l.orderData.discount[j])

			taxCol[idx] = float64(l.orderData.tax[j])

			shipDate := l.orderData.shipDates[j]
			shipDateCol[idx] = shipDate

			receiptDate := shipDate + int64(randInt(rng, 1, 30))
			receiptDateCol[idx] = receiptDate

			commitDateCol[idx] = int64(l.orderData.orderDate + randInt(rng, 30, 90))

			if receiptDate < currentDateDays {
				__antithesis_instrumentation__.Notify(698777)
				if rng.Intn(2) == 0 {
					__antithesis_instrumentation__.Notify(698778)
					returnFlagCol.Set(idx, []byte("R"))
				} else {
					__antithesis_instrumentation__.Notify(698779)
					returnFlagCol.Set(idx, []byte("A"))
				}
			} else {
				__antithesis_instrumentation__.Notify(698780)
				returnFlagCol.Set(idx, []byte("N"))
			}
			__antithesis_instrumentation__.Notify(698775)

			if shipDate > currentDateDays {
				__antithesis_instrumentation__.Notify(698781)
				lineStatusCol.Set(idx, []byte("O"))
			} else {
				__antithesis_instrumentation__.Notify(698782)
				lineStatusCol.Set(idx, []byte("F"))
			}
			__antithesis_instrumentation__.Notify(698776)

			shipInstructCol.Set(idx, randInstruction(rng))

			shipModeCol.Set(idx, randMode(rng))

			commentCol.Set(idx, w.textPool.randString(rng, 10, 43))
		}
		__antithesis_instrumentation__.Notify(698773)
		offset += l.orderData.nOrders
	}
	__antithesis_instrumentation__.Notify(698771)
	cb.SetLength(offset)
}
