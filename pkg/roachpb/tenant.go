package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"strconv"
)

var SystemTenantID = MakeTenantID(1)

var MinTenantID = MakeTenantID(2)

var MaxTenantID = MakeTenantID(math.MaxUint64)

func MakeTenantID(id uint64) TenantID {
	__antithesis_instrumentation__.Notify(179892)
	checkValid(id)
	return TenantID{id}
}

func (t TenantID) ToUint64() uint64 {
	__antithesis_instrumentation__.Notify(179893)
	checkValid(t.InternalValue)
	return t.InternalValue
}

func (t TenantID) String() string {
	__antithesis_instrumentation__.Notify(179894)
	switch t {
	case TenantID{}:
		__antithesis_instrumentation__.Notify(179895)
		return "invalid"
	case SystemTenantID:
		__antithesis_instrumentation__.Notify(179896)
		return "system"
	default:
		__antithesis_instrumentation__.Notify(179897)
		return strconv.FormatUint(t.InternalValue, 10)
	}
}

func (t TenantID) SafeValue() { __antithesis_instrumentation__.Notify(179898) }

func checkValid(id uint64) {
	__antithesis_instrumentation__.Notify(179899)
	if id == 0 {
		__antithesis_instrumentation__.Notify(179900)
		panic("invalid tenant ID 0")
	} else {
		__antithesis_instrumentation__.Notify(179901)
	}
}

func (t TenantID) IsSet() bool {
	__antithesis_instrumentation__.Notify(179902)
	return t.InternalValue != 0
}

func (t TenantID) IsSystem() bool {
	__antithesis_instrumentation__.Notify(179903)
	return IsSystemTenantID(t.InternalValue)
}

func IsSystemTenantID(id uint64) bool {
	__antithesis_instrumentation__.Notify(179904)
	return id == SystemTenantID.ToUint64()
}

type tenantKey struct{}

func NewContextForTenant(ctx context.Context, tenID TenantID) context.Context {
	__antithesis_instrumentation__.Notify(179905)
	return context.WithValue(ctx, tenantKey{}, tenID)
}

func TenantFromContext(ctx context.Context) (tenID TenantID, ok bool) {
	__antithesis_instrumentation__.Notify(179906)
	tenID, ok = ctx.Value(tenantKey{}).(TenantID)
	return
}

var _ = TenantFromContext
