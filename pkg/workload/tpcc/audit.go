package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"math"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

const (
	minSignificantTransactions = 10000
)

type auditor struct {
	syncutil.Mutex

	warehouses int

	newOrderTransactions    uint64
	newOrderRollbacks       uint64
	paymentTransactions     uint64
	orderStatusTransactions uint64
	deliveryTransactions    uint64

	orderLinesFreq map[int]uint64

	totalOrderLines uint64

	orderLineRemoteWarehouseFreq map[int]uint64

	paymentRemoteWarehouseFreq map[int]uint64

	paymentsByLastName    uint64
	orderStatusByLastName uint64

	skippedDelivieries uint64
}

func newAuditor(warehouses int) *auditor {
	__antithesis_instrumentation__.Notify(697440)
	return &auditor{
		warehouses:                   warehouses,
		orderLinesFreq:               make(map[int]uint64),
		orderLineRemoteWarehouseFreq: make(map[int]uint64),
		paymentRemoteWarehouseFreq:   make(map[int]uint64),
	}
}

type auditResult struct {
	status      string
	description string
}

var passResult = auditResult{status: "PASS"}

func newFailResult(format string, args ...interface{}) auditResult {
	__antithesis_instrumentation__.Notify(697441)
	return auditResult{"FAIL", fmt.Sprintf(format, args...)}
}

func newSkipResult(format string, args ...interface{}) auditResult {
	__antithesis_instrumentation__.Notify(697442)
	return auditResult{"SKIP", fmt.Sprintf(format, args...)}
}

func (a *auditor) runChecks(localWarehouses bool) {
	__antithesis_instrumentation__.Notify(697443)
	type check struct {
		name string
		f    func(a *auditor) auditResult
	}
	checks := []check{
		{"9.2.1.7", check9217},
		{"9.2.2.5.1", check92251},
		{"9.2.2.5.2", check92252},
		{"9.2.2.5.5", check92255},
		{"9.2.2.5.6", check92256},
	}

	if !localWarehouses {
		__antithesis_instrumentation__.Notify(697445)
		checks = append(checks,
			check{"9.2.2.5.3", check92253},
			check{"9.2.2.5.4", check92254},
		)
	} else {
		__antithesis_instrumentation__.Notify(697446)
	}
	__antithesis_instrumentation__.Notify(697444)

	for _, check := range checks {
		__antithesis_instrumentation__.Notify(697447)
		result := check.f(a)
		msg := fmt.Sprintf("Audit check %s: %s", check.name, result.status)
		if result.description == "" {
			__antithesis_instrumentation__.Notify(697448)
			fmt.Println(msg)
		} else {
			__antithesis_instrumentation__.Notify(697449)
			fmt.Println(msg + ": " + result.description)
		}
	}
}

func check9217(a *auditor) auditResult {
	__antithesis_instrumentation__.Notify(697450)

	a.Lock()
	defer a.Unlock()

	if a.deliveryTransactions < minSignificantTransactions {
		__antithesis_instrumentation__.Notify(697454)
		return newSkipResult("not enough delivery transactions to be statistically significant")
	} else {
		__antithesis_instrumentation__.Notify(697455)
	}
	__antithesis_instrumentation__.Notify(697451)

	var threshold uint64
	if a.deliveryTransactions > 100 {
		__antithesis_instrumentation__.Notify(697456)
		threshold = a.deliveryTransactions / 100
	} else {
		__antithesis_instrumentation__.Notify(697457)
		threshold = 1
	}
	__antithesis_instrumentation__.Notify(697452)
	if a.skippedDelivieries > threshold {
		__antithesis_instrumentation__.Notify(697458)
		return newFailResult(
			"expected no more than %d skipped deliveries, got %d", threshold, a.skippedDelivieries)
	} else {
		__antithesis_instrumentation__.Notify(697459)
	}
	__antithesis_instrumentation__.Notify(697453)
	return passResult
}

func check92251(a *auditor) auditResult {
	__antithesis_instrumentation__.Notify(697460)

	orders := atomic.LoadUint64(&a.newOrderTransactions)
	if orders < minSignificantTransactions {
		__antithesis_instrumentation__.Notify(697463)
		return newSkipResult("not enough orders to be statistically significant")
	} else {
		__antithesis_instrumentation__.Notify(697464)
	}
	__antithesis_instrumentation__.Notify(697461)
	rollbacks := atomic.LoadUint64(&a.newOrderRollbacks)
	rollbackPct := 100 * float64(rollbacks) / float64(orders)
	if rollbackPct < 0.9 || func() bool {
		__antithesis_instrumentation__.Notify(697465)
		return rollbackPct > 1.1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(697466)
		return newFailResult(
			"new order rollback percent %.1f is not between allowed bounds [0.9, 1.1]", rollbackPct)
	} else {
		__antithesis_instrumentation__.Notify(697467)
	}
	__antithesis_instrumentation__.Notify(697462)
	return passResult
}

func check92252(a *auditor) auditResult {
	__antithesis_instrumentation__.Notify(697468)

	a.Lock()
	defer a.Unlock()

	if a.newOrderTransactions < minSignificantTransactions {
		__antithesis_instrumentation__.Notify(697472)
		return newSkipResult("not enough orders to be statistically significant")
	} else {
		__antithesis_instrumentation__.Notify(697473)
	}
	__antithesis_instrumentation__.Notify(697469)

	avg := float64(a.totalOrderLines) / float64(a.newOrderTransactions)
	if avg < 9.5 || func() bool {
		__antithesis_instrumentation__.Notify(697474)
		return avg > 10.5 == true
	}() == true {
		__antithesis_instrumentation__.Notify(697475)
		return newFailResult(
			"average order-lines count %.1f is not between allowed bounds [9.5, 10.5]", avg)
	} else {
		__antithesis_instrumentation__.Notify(697476)
	}
	__antithesis_instrumentation__.Notify(697470)

	expectedPct := 100.0 / 11
	tolerance := 1.0
	for i := 5; i <= 15; i++ {
		__antithesis_instrumentation__.Notify(697477)
		freq := a.orderLinesFreq[i]
		pct := 100 * float64(freq) / float64(a.newOrderTransactions)
		if math.Abs(expectedPct-pct) > tolerance {
			__antithesis_instrumentation__.Notify(697478)
			return newFailResult(
				"order-lines count should be uniformly distributed from 5 to 15, but it was %d for %.1f "+
					"percent of orders", i, pct)
		} else {
			__antithesis_instrumentation__.Notify(697479)
		}
	}
	__antithesis_instrumentation__.Notify(697471)
	return passResult
}

func check92253(a *auditor) auditResult {
	__antithesis_instrumentation__.Notify(697480)

	a.Lock()
	defer a.Unlock()

	if a.warehouses == 1 {
		__antithesis_instrumentation__.Notify(697487)

		return passResult
	} else {
		__antithesis_instrumentation__.Notify(697488)
	}
	__antithesis_instrumentation__.Notify(697481)
	if a.newOrderTransactions < minSignificantTransactions {
		__antithesis_instrumentation__.Notify(697489)
		return newSkipResult("not enough orders to be statistically significant")
	} else {
		__antithesis_instrumentation__.Notify(697490)
	}
	__antithesis_instrumentation__.Notify(697482)

	var remoteOrderLines uint64
	for _, freq := range a.orderLineRemoteWarehouseFreq {
		__antithesis_instrumentation__.Notify(697491)
		remoteOrderLines += freq
	}
	__antithesis_instrumentation__.Notify(697483)
	remotePct := 100 * float64(remoteOrderLines) / float64(a.totalOrderLines)
	if remotePct < 0.95 || func() bool {
		__antithesis_instrumentation__.Notify(697492)
		return remotePct > 1.05 == true
	}() == true {
		__antithesis_instrumentation__.Notify(697493)
		return newFailResult(
			"remote order-line percent %.1f is not between allowed bounds [0.95, 1.05]", remotePct)
	} else {
		__antithesis_instrumentation__.Notify(697494)
	}
	__antithesis_instrumentation__.Notify(697484)

	if remoteOrderLines < 15*uint64(a.warehouses) {
		__antithesis_instrumentation__.Notify(697495)
		return newSkipResult("insufficient data for remote warehouse distribution check")
	} else {
		__antithesis_instrumentation__.Notify(697496)
	}
	__antithesis_instrumentation__.Notify(697485)
	for i := 0; i < a.warehouses; i++ {
		__antithesis_instrumentation__.Notify(697497)
		if _, ok := a.orderLineRemoteWarehouseFreq[i]; !ok {
			__antithesis_instrumentation__.Notify(697498)
			return newFailResult("no remote order-lines for warehouses %d", i)
		} else {
			__antithesis_instrumentation__.Notify(697499)
		}
	}
	__antithesis_instrumentation__.Notify(697486)

	return passResult
}

func check92254(a *auditor) auditResult {
	__antithesis_instrumentation__.Notify(697500)

	a.Lock()
	defer a.Unlock()

	if a.warehouses == 1 {
		__antithesis_instrumentation__.Notify(697507)

		return passResult
	} else {
		__antithesis_instrumentation__.Notify(697508)
	}
	__antithesis_instrumentation__.Notify(697501)
	if a.paymentTransactions < minSignificantTransactions {
		__antithesis_instrumentation__.Notify(697509)
		return newSkipResult("not enough payments to be statistically significant")
	} else {
		__antithesis_instrumentation__.Notify(697510)
	}
	__antithesis_instrumentation__.Notify(697502)

	var remotePayments uint64
	for _, freq := range a.paymentRemoteWarehouseFreq {
		__antithesis_instrumentation__.Notify(697511)
		remotePayments += freq
	}
	__antithesis_instrumentation__.Notify(697503)
	remotePct := 100 * float64(remotePayments) / float64(a.paymentTransactions)
	if remotePct < 14 || func() bool {
		__antithesis_instrumentation__.Notify(697512)
		return remotePct > 16 == true
	}() == true {
		__antithesis_instrumentation__.Notify(697513)
		return newFailResult(
			"remote payment percent %.1f is not between allowed bounds [14, 16]", remotePct)
	} else {
		__antithesis_instrumentation__.Notify(697514)
	}
	__antithesis_instrumentation__.Notify(697504)

	if remotePayments < 15*uint64(a.warehouses) {
		__antithesis_instrumentation__.Notify(697515)
		return newSkipResult("insufficient data for remote warehouse distribution check")
	} else {
		__antithesis_instrumentation__.Notify(697516)
	}
	__antithesis_instrumentation__.Notify(697505)
	for i := 0; i < a.warehouses; i++ {
		__antithesis_instrumentation__.Notify(697517)
		if _, ok := a.paymentRemoteWarehouseFreq[i]; !ok {
			__antithesis_instrumentation__.Notify(697518)
			return newFailResult("no remote payments for warehouses %d", i)
		} else {
			__antithesis_instrumentation__.Notify(697519)
		}
	}
	__antithesis_instrumentation__.Notify(697506)

	return passResult
}

func check92255(a *auditor) auditResult {
	__antithesis_instrumentation__.Notify(697520)

	a.Lock()
	defer a.Unlock()

	if a.paymentTransactions < minSignificantTransactions {
		__antithesis_instrumentation__.Notify(697523)
		return newSkipResult("not enough payments to be statistically significant")
	} else {
		__antithesis_instrumentation__.Notify(697524)
	}
	__antithesis_instrumentation__.Notify(697521)
	lastNamePct := 100 * float64(a.paymentsByLastName) / float64(a.paymentTransactions)
	if lastNamePct < 57 || func() bool {
		__antithesis_instrumentation__.Notify(697525)
		return lastNamePct > 63 == true
	}() == true {
		__antithesis_instrumentation__.Notify(697526)
		return newFailResult(
			"percent of customer selections by last name in payment transactions %.1f is not between "+
				"allowed bounds [57, 63]", lastNamePct)
	} else {
		__antithesis_instrumentation__.Notify(697527)
	}
	__antithesis_instrumentation__.Notify(697522)

	return passResult
}

func check92256(a *auditor) auditResult {
	__antithesis_instrumentation__.Notify(697528)

	a.Lock()
	defer a.Unlock()

	if a.orderStatusTransactions < minSignificantTransactions {
		__antithesis_instrumentation__.Notify(697531)
		return newSkipResult("not enough order status transactions to be statistically significant")
	} else {
		__antithesis_instrumentation__.Notify(697532)
	}
	__antithesis_instrumentation__.Notify(697529)
	lastNamePct := 100 * float64(a.orderStatusByLastName) / float64(a.orderStatusTransactions)
	if lastNamePct < 57 || func() bool {
		__antithesis_instrumentation__.Notify(697533)
		return lastNamePct > 63 == true
	}() == true {
		__antithesis_instrumentation__.Notify(697534)
		return newFailResult(
			"percent of customer selections by last name in order status transactions %.1f is not "+
				"between allowed bounds [57, 63]", lastNamePct)
	} else {
		__antithesis_instrumentation__.Notify(697535)
	}
	__antithesis_instrumentation__.Notify(697530)

	return passResult
}
