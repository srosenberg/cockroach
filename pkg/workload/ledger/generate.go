package ledger

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/binary"
	"hash"
	"math"
	"math/rand"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/uint128"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

const (
	numTxnsPerCustomer    = 2
	numEntriesPerTxn      = 4
	numEntriesPerCustomer = numTxnsPerCustomer * numEntriesPerTxn

	paymentIDPrefix  = "payment:"
	txnTypeReference = 400
	cashMoneyType    = "C"
)

var ledgerCustomerTypes = []*types.T{
	types.Int,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bool,
	types.Bool,
	types.Bytes,
	types.Int,
	types.Int,
	types.Int,
}

func (w *ledger) ledgerCustomerInitialRow(rowIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694596)
	rng := w.rngPool.Get().(*rand.Rand)
	defer w.rngPool.Put(rng)
	rng.Seed(w.seed + int64(rowIdx))

	return []interface{}{
		rowIdx,
		strconv.Itoa(rowIdx),
		nil,
		randCurrencyCode(rng),
		true,
		true,
		randTimestamp(rng),
		0,
		nil,
		-1,
	}
}

func (w *ledger) ledgerCustomerSplitRow(splitIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694597)
	return []interface{}{
		(splitIdx + 1) * (w.customers / w.splits),
	}
}

var ledgerTransactionColTypes = []*types.T{
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Int,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
	types.Bytes,
}

func (w *ledger) ledgerTransactionInitialRow(rowIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694598)
	rng := w.rngPool.Get().(*rand.Rand)
	defer w.rngPool.Put(rng)
	rng.Seed(w.seed + int64(rowIdx))

	h := w.hashPool.Get().(hash.Hash64)
	defer w.hashPool.Put(h)
	defer h.Reset()

	return []interface{}{
		w.ledgerStablePaymentID(rowIdx),
		nil,
		randContext(rng),
		txnTypeReference,
		randUsername(rng),
		randTimestamp(rng),
		randTimestamp(rng),
		nil,
		randResponse(rng),
	}
}

func (w *ledger) ledgerTransactionSplitRow(splitIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694599)
	rng := rand.New(rand.NewSource(w.seed + int64(splitIdx)))
	u := uuid.FromUint128(uint128.FromInts(rng.Uint64(), rng.Uint64()))
	return []interface{}{
		paymentIDPrefix + u.String(),
	}
}

func (w *ledger) ledgerEntryInitialRow(rowIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694600)
	rng := w.rngPool.Get().(*rand.Rand)
	defer w.rngPool.Put(rng)
	rng.Seed(w.seed + int64(rowIdx))

	debit := rowIdx%2 == 0

	var amount float64
	if debit {
		__antithesis_instrumentation__.Notify(694603)
		amount = -float64(rowIdx) / 100
	} else {
		__antithesis_instrumentation__.Notify(694604)
		amount = float64(rowIdx-1) / 100
	}
	__antithesis_instrumentation__.Notify(694601)

	systemAmount := 88.122259
	if debit {
		__antithesis_instrumentation__.Notify(694605)
		systemAmount *= -1
	} else {
		__antithesis_instrumentation__.Notify(694606)
	}
	__antithesis_instrumentation__.Notify(694602)

	cRowIdx := rowIdx / numEntriesPerCustomer

	tRowIdx := rowIdx / numEntriesPerTxn
	tID := w.ledgerStablePaymentID(tRowIdx)

	return []interface{}{
		rng.Int(),
		amount,
		cRowIdx,
		tID,
		systemAmount,
		randTimestamp(rng),
		cashMoneyType,
	}
}

func (w *ledger) ledgerEntrySplitRow(splitIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694607)
	return []interface{}{
		(splitIdx + 1) * (int(math.MaxInt64) / w.splits),
	}
}

func (w *ledger) ledgerSessionInitialRow(rowIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694608)
	rng := w.rngPool.Get().(*rand.Rand)
	defer w.rngPool.Put(rng)
	rng.Seed(w.seed + int64(rowIdx))

	return []interface{}{
		randSessionID(rng),
		randTimestamp(rng),
		randSessionData(rng),
		randTimestamp(rng),
	}
}

func (w *ledger) ledgerSessionSplitRow(splitIdx int) []interface{} {
	__antithesis_instrumentation__.Notify(694609)
	rng := rand.New(rand.NewSource(w.seed + int64(splitIdx)))
	return []interface{}{
		randSessionID(rng),
	}
}

func (w *ledger) ledgerStablePaymentID(tRowIdx int) string {
	__antithesis_instrumentation__.Notify(694610)
	h := w.hashPool.Get().(hash.Hash64)
	defer w.hashPool.Put(h)
	defer h.Reset()

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(tRowIdx))
	if _, err := h.Write(b); err != nil {
		__antithesis_instrumentation__.Notify(694613)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(694614)
	}
	__antithesis_instrumentation__.Notify(694611)
	hi := h.Sum64()
	if _, err := h.Write(b); err != nil {
		__antithesis_instrumentation__.Notify(694615)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(694616)
	}
	__antithesis_instrumentation__.Notify(694612)
	low := h.Sum64()

	u := uuid.FromUint128(uint128.FromInts(hi, low))
	return paymentIDPrefix + u.String()
}
