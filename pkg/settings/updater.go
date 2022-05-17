package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
)

func EncodeDuration(d time.Duration) string {
	__antithesis_instrumentation__.Notify(240132)
	return d.String()
}

func EncodeBool(b bool) string {
	__antithesis_instrumentation__.Notify(240133)
	return strconv.FormatBool(b)
}

func EncodeInt(i int64) string {
	__antithesis_instrumentation__.Notify(240134)
	return strconv.FormatInt(i, 10)
}

func EncodeFloat(f float64) string {
	__antithesis_instrumentation__.Notify(240135)
	return strconv.FormatFloat(f, 'G', -1, 64)
}

type updater struct {
	sv *Values
	m  map[string]struct{}
}

type Updater interface {
	Set(ctx context.Context, key string, value EncodedValue) error
	ResetRemaining(ctx context.Context)
}

type NoopUpdater struct{}

func (u NoopUpdater) Set(ctx context.Context, key string, value EncodedValue) error {
	__antithesis_instrumentation__.Notify(240136)
	return nil
}

func (u NoopUpdater) ResetRemaining(context.Context) { __antithesis_instrumentation__.Notify(240137) }

func NewUpdater(sv *Values) Updater {
	__antithesis_instrumentation__.Notify(240138)
	return updater{
		m:  make(map[string]struct{}, len(registry)),
		sv: sv,
	}
}

func (u updater) Set(ctx context.Context, key string, value EncodedValue) error {
	__antithesis_instrumentation__.Notify(240139)
	d, ok := registry[key]
	if !ok {
		__antithesis_instrumentation__.Notify(240143)
		if _, ok := retiredSettings[key]; ok {
			__antithesis_instrumentation__.Notify(240145)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(240146)
		}
		__antithesis_instrumentation__.Notify(240144)

		return errors.Errorf("unknown setting '%s'", key)
	} else {
		__antithesis_instrumentation__.Notify(240147)
	}
	__antithesis_instrumentation__.Notify(240140)

	u.m[key] = struct{}{}

	if expected := d.Typ(); value.Type != expected {
		__antithesis_instrumentation__.Notify(240148)
		return errors.Errorf("setting '%s' defined as type %s, not %s", key, expected, value.Type)
	} else {
		__antithesis_instrumentation__.Notify(240149)
	}
	__antithesis_instrumentation__.Notify(240141)

	switch setting := d.(type) {
	case *StringSetting:
		__antithesis_instrumentation__.Notify(240150)
		return setting.set(ctx, u.sv, value.Value)
	case *BoolSetting:
		__antithesis_instrumentation__.Notify(240151)
		b, err := setting.DecodeValue(value.Value)
		if err != nil {
			__antithesis_instrumentation__.Notify(240162)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240163)
		}
		__antithesis_instrumentation__.Notify(240152)
		setting.set(ctx, u.sv, b)
		return nil
	case numericSetting:
		__antithesis_instrumentation__.Notify(240153)
		i, err := setting.DecodeValue(value.Value)
		if err != nil {
			__antithesis_instrumentation__.Notify(240164)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240165)
		}
		__antithesis_instrumentation__.Notify(240154)
		return setting.set(ctx, u.sv, i)
	case *FloatSetting:
		__antithesis_instrumentation__.Notify(240155)
		f, err := setting.DecodeValue(value.Value)
		if err != nil {
			__antithesis_instrumentation__.Notify(240166)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240167)
		}
		__antithesis_instrumentation__.Notify(240156)
		return setting.set(ctx, u.sv, f)
	case *DurationSetting:
		__antithesis_instrumentation__.Notify(240157)
		d, err := setting.DecodeValue(value.Value)
		if err != nil {
			__antithesis_instrumentation__.Notify(240168)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240169)
		}
		__antithesis_instrumentation__.Notify(240158)
		return setting.set(ctx, u.sv, d)
	case *DurationSettingWithExplicitUnit:
		__antithesis_instrumentation__.Notify(240159)
		d, err := setting.DecodeValue(value.Value)
		if err != nil {
			__antithesis_instrumentation__.Notify(240170)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240171)
		}
		__antithesis_instrumentation__.Notify(240160)
		return setting.set(ctx, u.sv, d)
	case *VersionSetting:
		__antithesis_instrumentation__.Notify(240161)

		return nil
	}
	__antithesis_instrumentation__.Notify(240142)
	return nil
}

func (u updater) ResetRemaining(ctx context.Context) {
	__antithesis_instrumentation__.Notify(240172)
	for k, v := range registry {
		__antithesis_instrumentation__.Notify(240173)
		if u.sv.NonSystemTenant() && func() bool {
			__antithesis_instrumentation__.Notify(240175)
			return v.Class() == SystemOnly == true
		}() == true {
			__antithesis_instrumentation__.Notify(240176)

			continue
		} else {
			__antithesis_instrumentation__.Notify(240177)
		}
		__antithesis_instrumentation__.Notify(240174)
		if _, ok := u.m[k]; !ok {
			__antithesis_instrumentation__.Notify(240178)
			v.setToDefault(ctx, u.sv)
		} else {
			__antithesis_instrumentation__.Notify(240179)
		}
	}
}
