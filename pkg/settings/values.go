package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

const MaxSettings = 511

type Values struct {
	container valuesContainer

	nonSystemTenant bool

	defaultOverridesMu struct {
		syncutil.Mutex

		defaultOverrides map[slotIdx]interface{}
	}

	changeMu struct {
		syncutil.Mutex

		onChange [MaxSettings][]func(ctx context.Context)
	}

	opaque interface{}
}

const numSlots = MaxSettings + 1

type valuesContainer struct {
	intVals     [numSlots]int64
	genericVals [numSlots]atomic.Value

	forbidden [numSlots]bool
}

func (c *valuesContainer) setGenericVal(slot slotIdx, newVal interface{}) {
	__antithesis_instrumentation__.Notify(240180)
	if !c.checkForbidden(slot) {
		__antithesis_instrumentation__.Notify(240182)
		return
	} else {
		__antithesis_instrumentation__.Notify(240183)
	}
	__antithesis_instrumentation__.Notify(240181)
	c.genericVals[slot].Store(newVal)
}

func (c *valuesContainer) setInt64Val(slot slotIdx, newVal int64) (changed bool) {
	__antithesis_instrumentation__.Notify(240184)
	if !c.checkForbidden(slot) {
		__antithesis_instrumentation__.Notify(240186)
		return false
	} else {
		__antithesis_instrumentation__.Notify(240187)
	}
	__antithesis_instrumentation__.Notify(240185)
	return atomic.SwapInt64(&c.intVals[slot], newVal) != newVal
}

func (c *valuesContainer) getInt64(slot slotIdx) int64 {
	__antithesis_instrumentation__.Notify(240188)
	c.checkForbidden(slot)
	return atomic.LoadInt64(&c.intVals[slot])
}

func (c *valuesContainer) getGeneric(slot slotIdx) interface{} {
	__antithesis_instrumentation__.Notify(240189)
	c.checkForbidden(slot)
	return c.genericVals[slot].Load()
}

func (c *valuesContainer) checkForbidden(slot slotIdx) bool {
	__antithesis_instrumentation__.Notify(240190)
	if c.forbidden[slot] {
		__antithesis_instrumentation__.Notify(240192)
		if buildutil.CrdbTestBuild {
			__antithesis_instrumentation__.Notify(240194)
			panic(errors.AssertionFailedf("attempted to set forbidden setting %s", slotTable[slot].Key()))
		} else {
			__antithesis_instrumentation__.Notify(240195)
		}
		__antithesis_instrumentation__.Notify(240193)
		return false
	} else {
		__antithesis_instrumentation__.Notify(240196)
	}
	__antithesis_instrumentation__.Notify(240191)
	return true
}

type testOpaqueType struct{}

var TestOpaque interface{} = testOpaqueType{}

func (sv *Values) Init(ctx context.Context, opaque interface{}) {
	__antithesis_instrumentation__.Notify(240197)
	sv.opaque = opaque
	for _, s := range registry {
		__antithesis_instrumentation__.Notify(240198)
		s.setToDefault(ctx, sv)
	}
}

func (sv *Values) SetNonSystemTenant() {
	__antithesis_instrumentation__.Notify(240199)
	sv.nonSystemTenant = true
	for slot, setting := range slotTable {
		__antithesis_instrumentation__.Notify(240200)
		if setting != nil && func() bool {
			__antithesis_instrumentation__.Notify(240201)
			return setting.Class() == SystemOnly == true
		}() == true {
			__antithesis_instrumentation__.Notify(240202)
			sv.container.forbidden[slot] = true
		} else {
			__antithesis_instrumentation__.Notify(240203)
		}
	}
}

func (sv *Values) NonSystemTenant() bool {
	__antithesis_instrumentation__.Notify(240204)
	return sv.nonSystemTenant
}

func (sv *Values) Opaque() interface{} {
	__antithesis_instrumentation__.Notify(240205)
	return sv.opaque
}

func (sv *Values) settingChanged(ctx context.Context, slot slotIdx) {
	__antithesis_instrumentation__.Notify(240206)
	sv.changeMu.Lock()
	funcs := sv.changeMu.onChange[slot]
	sv.changeMu.Unlock()
	for _, fn := range funcs {
		__antithesis_instrumentation__.Notify(240207)
		fn(ctx)
	}
}

func (sv *Values) setInt64(ctx context.Context, slot slotIdx, newVal int64) {
	__antithesis_instrumentation__.Notify(240208)
	if sv.container.setInt64Val(slot, newVal) {
		__antithesis_instrumentation__.Notify(240209)
		sv.settingChanged(ctx, slot)
	} else {
		__antithesis_instrumentation__.Notify(240210)
	}
}

func (sv *Values) setDefaultOverride(slot slotIdx, newVal interface{}) {
	__antithesis_instrumentation__.Notify(240211)
	sv.defaultOverridesMu.Lock()
	defer sv.defaultOverridesMu.Unlock()
	if sv.defaultOverridesMu.defaultOverrides == nil {
		__antithesis_instrumentation__.Notify(240213)
		sv.defaultOverridesMu.defaultOverrides = make(map[slotIdx]interface{})
	} else {
		__antithesis_instrumentation__.Notify(240214)
	}
	__antithesis_instrumentation__.Notify(240212)
	sv.defaultOverridesMu.defaultOverrides[slot] = newVal
}

func (sv *Values) getDefaultOverride(slot slotIdx) interface{} {
	__antithesis_instrumentation__.Notify(240215)
	sv.defaultOverridesMu.Lock()
	defer sv.defaultOverridesMu.Unlock()
	return sv.defaultOverridesMu.defaultOverrides[slot]
}

func (sv *Values) setGeneric(ctx context.Context, slot slotIdx, newVal interface{}) {
	__antithesis_instrumentation__.Notify(240216)
	sv.container.setGenericVal(slot, newVal)
	sv.settingChanged(ctx, slot)
}

func (sv *Values) getInt64(slot slotIdx) int64 {
	__antithesis_instrumentation__.Notify(240217)
	return sv.container.getInt64(slot)
}

func (sv *Values) getGeneric(slot slotIdx) interface{} {
	__antithesis_instrumentation__.Notify(240218)
	return sv.container.getGeneric(slot)
}

func (sv *Values) setOnChange(slot slotIdx, fn func(ctx context.Context)) {
	__antithesis_instrumentation__.Notify(240219)
	sv.changeMu.Lock()
	sv.changeMu.onChange[slot] = append(sv.changeMu.onChange[slot], fn)
	sv.changeMu.Unlock()
}
