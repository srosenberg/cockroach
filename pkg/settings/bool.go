package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"
)

type BoolSetting struct {
	common
	defaultValue bool
}

var _ internalSetting = &BoolSetting{}

func (b *BoolSetting) Get(sv *Values) bool {
	__antithesis_instrumentation__.Notify(239541)
	return sv.getInt64(b.slot) != 0
}

func (b *BoolSetting) String(sv *Values) string {
	__antithesis_instrumentation__.Notify(239542)
	return EncodeBool(b.Get(sv))
}

func (b *BoolSetting) Encoded(sv *Values) string {
	__antithesis_instrumentation__.Notify(239543)
	return b.String(sv)
}

func (b *BoolSetting) EncodedDefault() string {
	__antithesis_instrumentation__.Notify(239544)
	return EncodeBool(b.defaultValue)
}

func (b *BoolSetting) DecodeToString(encoded string) (string, error) {
	__antithesis_instrumentation__.Notify(239545)
	bv, err := b.DecodeValue(encoded)
	if err != nil {
		__antithesis_instrumentation__.Notify(239547)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(239548)
	}
	__antithesis_instrumentation__.Notify(239546)
	return EncodeBool(bv), nil
}

func (b *BoolSetting) DecodeValue(encoded string) (bool, error) {
	__antithesis_instrumentation__.Notify(239549)
	return strconv.ParseBool(encoded)
}

func (*BoolSetting) Typ() string {
	__antithesis_instrumentation__.Notify(239550)
	return "b"
}

func (b *BoolSetting) Default() bool {
	__antithesis_instrumentation__.Notify(239551)
	return b.defaultValue
}

var _ = (*BoolSetting).Default

func (b *BoolSetting) Override(ctx context.Context, sv *Values, v bool) {
	__antithesis_instrumentation__.Notify(239552)
	b.set(ctx, sv, v)
	sv.setDefaultOverride(b.slot, v)
}

func (b *BoolSetting) set(ctx context.Context, sv *Values, v bool) {
	__antithesis_instrumentation__.Notify(239553)
	vInt := int64(0)
	if v {
		__antithesis_instrumentation__.Notify(239555)
		vInt = 1
	} else {
		__antithesis_instrumentation__.Notify(239556)
	}
	__antithesis_instrumentation__.Notify(239554)
	sv.setInt64(ctx, b.slot, vInt)
}

func (b *BoolSetting) setToDefault(ctx context.Context, sv *Values) {
	__antithesis_instrumentation__.Notify(239557)

	if val := sv.getDefaultOverride(b.slot); val != nil {
		__antithesis_instrumentation__.Notify(239559)
		b.set(ctx, sv, val.(bool))
		return
	} else {
		__antithesis_instrumentation__.Notify(239560)
	}
	__antithesis_instrumentation__.Notify(239558)
	b.set(ctx, sv, b.defaultValue)
}

func (b *BoolSetting) WithPublic() *BoolSetting {
	__antithesis_instrumentation__.Notify(239561)
	b.SetVisibility(Public)
	return b
}

func RegisterBoolSetting(class Class, key, desc string, defaultValue bool) *BoolSetting {
	__antithesis_instrumentation__.Notify(239562)
	setting := &BoolSetting{defaultValue: defaultValue}
	register(class, key, desc, setting)
	return setting
}
