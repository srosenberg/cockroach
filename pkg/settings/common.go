package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
)

type common struct {
	class         Class
	key           string
	description   string
	visibility    Visibility
	slot          slotIdx
	nonReportable bool
	retired       bool
}

type slotIdx int32

func (c *common) init(class Class, key string, description string, slot slotIdx) {
	c.class = class
	c.key = key
	c.description = description
	if slot < 0 {
		panic(fmt.Sprintf("Invalid slot index %d", slot))
	}
	if slot >= MaxSettings {
		panic("too many settings; increase MaxSettings")
	}
	c.slot = slot
}

func (c common) Class() Class {
	__antithesis_instrumentation__.Notify(239608)
	return c.class
}

func (c common) Key() string {
	__antithesis_instrumentation__.Notify(239609)
	return c.key
}

func (c common) Description() string {
	__antithesis_instrumentation__.Notify(239610)
	return c.description
}

func (c common) Visibility() Visibility {
	__antithesis_instrumentation__.Notify(239611)
	return c.visibility
}

func (c common) isReportable() bool {
	__antithesis_instrumentation__.Notify(239612)
	return !c.nonReportable
}

func (c *common) isRetired() bool {
	__antithesis_instrumentation__.Notify(239613)
	return c.retired
}

func (c *common) ErrorHint() (bool, string) {
	__antithesis_instrumentation__.Notify(239614)
	return false, ""
}

func (c *common) SetReportable(reportable bool) {
	__antithesis_instrumentation__.Notify(239615)
	c.nonReportable = !reportable
}

func (c *common) SetVisibility(v Visibility) {
	__antithesis_instrumentation__.Notify(239616)
	c.visibility = v
}

func (c *common) SetRetired() {
	__antithesis_instrumentation__.Notify(239617)
	c.description = "do not use - " + c.description
	c.retired = true
}

func (c *common) SetOnChange(sv *Values, fn func(ctx context.Context)) {
	__antithesis_instrumentation__.Notify(239618)
	sv.setOnChange(c.slot, fn)
}

type internalSetting interface {
	NonMaskedSetting

	init(class Class, key, description string, slot slotIdx)
	isRetired() bool
	setToDefault(ctx context.Context, sv *Values)

	isReportable() bool
}

type numericSetting interface {
	internalSetting
	DecodeValue(value string) (int64, error)
	set(ctx context.Context, sv *Values, value int64) error
}
