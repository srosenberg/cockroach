package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"

	"github.com/cockroachdb/errors"
)

type IntSetting struct {
	common
	defaultValue int64
	validateFn   func(int64) error
}

var _ numericSetting = &IntSetting{}

func (i *IntSetting) Get(sv *Values) int64 {
	__antithesis_instrumentation__.Notify(239976)
	return sv.container.getInt64(i.slot)
}

func (i *IntSetting) String(sv *Values) string {
	__antithesis_instrumentation__.Notify(239977)
	return EncodeInt(i.Get(sv))
}

func (i *IntSetting) Encoded(sv *Values) string {
	__antithesis_instrumentation__.Notify(239978)
	return i.String(sv)
}

func (i *IntSetting) EncodedDefault() string {
	__antithesis_instrumentation__.Notify(239979)
	return EncodeInt(i.defaultValue)
}

func (i *IntSetting) DecodeToString(encoded string) (string, error) {
	__antithesis_instrumentation__.Notify(239980)
	iv, err := i.DecodeValue(encoded)
	if err != nil {
		__antithesis_instrumentation__.Notify(239982)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(239983)
	}
	__antithesis_instrumentation__.Notify(239981)
	return EncodeInt(iv), nil
}

func (i *IntSetting) DecodeValue(value string) (int64, error) {
	__antithesis_instrumentation__.Notify(239984)
	return strconv.ParseInt(value, 10, 64)
}

func (*IntSetting) Typ() string {
	__antithesis_instrumentation__.Notify(239985)
	return "i"
}

func (i *IntSetting) Default() int64 {
	__antithesis_instrumentation__.Notify(239986)
	return i.defaultValue
}

var _ = (*IntSetting).Default

func (i *IntSetting) Validate(v int64) error {
	__antithesis_instrumentation__.Notify(239987)
	if i.validateFn != nil {
		__antithesis_instrumentation__.Notify(239989)
		if err := i.validateFn(v); err != nil {
			__antithesis_instrumentation__.Notify(239990)
			return err
		} else {
			__antithesis_instrumentation__.Notify(239991)
		}
	} else {
		__antithesis_instrumentation__.Notify(239992)
	}
	__antithesis_instrumentation__.Notify(239988)
	return nil
}

func (i *IntSetting) Override(ctx context.Context, sv *Values, v int64) {
	__antithesis_instrumentation__.Notify(239993)
	sv.setInt64(ctx, i.slot, v)
	sv.setDefaultOverride(i.slot, v)
}

func (i *IntSetting) set(ctx context.Context, sv *Values, v int64) error {
	__antithesis_instrumentation__.Notify(239994)
	if err := i.Validate(v); err != nil {
		__antithesis_instrumentation__.Notify(239996)
		return err
	} else {
		__antithesis_instrumentation__.Notify(239997)
	}
	__antithesis_instrumentation__.Notify(239995)
	sv.setInt64(ctx, i.slot, v)
	return nil
}

func (i *IntSetting) setToDefault(ctx context.Context, sv *Values) {
	__antithesis_instrumentation__.Notify(239998)

	if val := sv.getDefaultOverride(i.slot); val != nil {
		__antithesis_instrumentation__.Notify(240000)

		_ = i.set(ctx, sv, val.(int64))
		return
	} else {
		__antithesis_instrumentation__.Notify(240001)
	}
	__antithesis_instrumentation__.Notify(239999)
	if err := i.set(ctx, sv, i.defaultValue); err != nil {
		__antithesis_instrumentation__.Notify(240002)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(240003)
	}
}

func RegisterIntSetting(
	class Class, key, desc string, defaultValue int64, validateFns ...func(int64) error,
) *IntSetting {
	__antithesis_instrumentation__.Notify(240004)
	var composed func(int64) error
	if len(validateFns) > 0 {
		__antithesis_instrumentation__.Notify(240007)
		composed = func(v int64) error {
			__antithesis_instrumentation__.Notify(240008)
			for _, validateFn := range validateFns {
				__antithesis_instrumentation__.Notify(240010)
				if err := validateFn(v); err != nil {
					__antithesis_instrumentation__.Notify(240011)
					return errors.Wrapf(err, "invalid value for %s", key)
				} else {
					__antithesis_instrumentation__.Notify(240012)
				}
			}
			__antithesis_instrumentation__.Notify(240009)
			return nil
		}
	} else {
		__antithesis_instrumentation__.Notify(240013)
	}
	__antithesis_instrumentation__.Notify(240005)
	if composed != nil {
		__antithesis_instrumentation__.Notify(240014)
		if err := composed(defaultValue); err != nil {
			__antithesis_instrumentation__.Notify(240015)
			panic(errors.Wrap(err, "invalid default"))
		} else {
			__antithesis_instrumentation__.Notify(240016)
		}
	} else {
		__antithesis_instrumentation__.Notify(240017)
	}
	__antithesis_instrumentation__.Notify(240006)
	setting := &IntSetting{
		defaultValue: defaultValue,
		validateFn:   composed,
	}
	register(class, key, desc, setting)
	return setting
}

func (i *IntSetting) WithPublic() *IntSetting {
	__antithesis_instrumentation__.Notify(240018)
	i.SetVisibility(Public)
	return i
}

func PositiveInt(v int64) error {
	__antithesis_instrumentation__.Notify(240019)
	if v < 1 {
		__antithesis_instrumentation__.Notify(240021)
		return errors.Errorf("cannot be set to a non-positive value: %d", v)
	} else {
		__antithesis_instrumentation__.Notify(240022)
	}
	__antithesis_instrumentation__.Notify(240020)
	return nil
}

func NonNegativeInt(v int64) error {
	__antithesis_instrumentation__.Notify(240023)
	if v < 0 {
		__antithesis_instrumentation__.Notify(240025)
		return errors.Errorf("cannot be set to a negative value: %d", v)
	} else {
		__antithesis_instrumentation__.Notify(240026)
	}
	__antithesis_instrumentation__.Notify(240024)
	return nil
}
