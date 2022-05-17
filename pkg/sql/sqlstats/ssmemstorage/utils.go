package ssmemstorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type stmtList []stmtKey

func (s stmtList) Len() int {
	__antithesis_instrumentation__.Notify(625761)
	return len(s)
}
func (s stmtList) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(625762)
	s[i], s[j] = s[j], s[i]
}
func (s stmtList) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(625763)
	cmp := strings.Compare(s[i].anonymizedStmt, s[j].anonymizedStmt)
	if cmp == -1 {
		__antithesis_instrumentation__.Notify(625766)
		return true
	} else {
		__antithesis_instrumentation__.Notify(625767)
	}
	__antithesis_instrumentation__.Notify(625764)

	if cmp == 1 {
		__antithesis_instrumentation__.Notify(625768)
		return false
	} else {
		__antithesis_instrumentation__.Notify(625769)
	}
	__antithesis_instrumentation__.Notify(625765)
	return s[i].transactionFingerprintID < s[j].transactionFingerprintID
}

type txnList []roachpb.TransactionFingerprintID

func (t txnList) Len() int {
	__antithesis_instrumentation__.Notify(625770)
	return len(t)
}

func (t txnList) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(625771)
	t[i], t[j] = t[j], t[i]
}

func (t txnList) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(625772)
	return t[i] < t[j]
}
