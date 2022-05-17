package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
)

type DurationSetting struct {
	common
	defaultValue time.Duration
	validateFn   func(time.Duration) error
}

type DurationSettingWithExplicitUnit struct {
	DurationSetting
}

var _ internalSetting = &DurationSetting{}

var _ internalSetting = &DurationSettingWithExplicitUnit{}

func (d *DurationSettingWithExplicitUnit) ErrorHint() (bool, string) {
	__antithesis_instrumentation__.Notify(239619)
	return true, "try using an interval value with explicit units, e.g 500ms or 1h2m"
}

func (d *DurationSetting) Get(sv *Values) time.Duration {
	__antithesis_instrumentation__.Notify(239620)
	return time.Duration(sv.getInt64(d.slot))
}

func (d *DurationSetting) String(sv *Values) string {
	__antithesis_instrumentation__.Notify(239621)
	return EncodeDuration(d.Get(sv))
}

func (d *DurationSetting) Encoded(sv *Values) string {
	__antithesis_instrumentation__.Notify(239622)
	return d.String(sv)
}

func (d *DurationSetting) EncodedDefault() string {
	__antithesis_instrumentation__.Notify(239623)
	return EncodeDuration(d.defaultValue)
}

func (d *DurationSetting) DecodeToString(encoded string) (string, error) {
	__antithesis_instrumentation__.Notify(239624)
	v, err := d.DecodeValue(encoded)
	if err != nil {
		__antithesis_instrumentation__.Notify(239626)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(239627)
	}
	__antithesis_instrumentation__.Notify(239625)
	return EncodeDuration(v), nil
}

func (d *DurationSetting) DecodeValue(encoded string) (time.Duration, error) {
	__antithesis_instrumentation__.Notify(239628)
	return time.ParseDuration(encoded)
}

func (*DurationSetting) Typ() string {
	__antithesis_instrumentation__.Notify(239629)
	return "d"
}

func (d *DurationSetting) Default() time.Duration {
	__antithesis_instrumentation__.Notify(239630)
	return d.defaultValue
}

var _ = (*DurationSetting).Default

func (d *DurationSetting) Validate(v time.Duration) error {
	__antithesis_instrumentation__.Notify(239631)
	if d.validateFn != nil {
		__antithesis_instrumentation__.Notify(239633)
		if err := d.validateFn(v); err != nil {
			__antithesis_instrumentation__.Notify(239634)
			return err
		} else {
			__antithesis_instrumentation__.Notify(239635)
		}
	} else {
		__antithesis_instrumentation__.Notify(239636)
	}
	__antithesis_instrumentation__.Notify(239632)
	return nil
}

func (d *DurationSetting) Override(ctx context.Context, sv *Values, v time.Duration) {
	__antithesis_instrumentation__.Notify(239637)
	sv.setInt64(ctx, d.slot, int64(v))
	sv.setDefaultOverride(d.slot, v)
}

func (d *DurationSetting) set(ctx context.Context, sv *Values, v time.Duration) error {
	__antithesis_instrumentation__.Notify(239638)
	if err := d.Validate(v); err != nil {
		__antithesis_instrumentation__.Notify(239640)
		return err
	} else {
		__antithesis_instrumentation__.Notify(239641)
	}
	__antithesis_instrumentation__.Notify(239639)
	sv.setInt64(ctx, d.slot, int64(v))
	return nil
}

func (d *DurationSetting) setToDefault(ctx context.Context, sv *Values) {
	__antithesis_instrumentation__.Notify(239642)

	if val := sv.getDefaultOverride(d.slot); val != nil {
		__antithesis_instrumentation__.Notify(239644)

		_ = d.set(ctx, sv, val.(time.Duration))
		return
	} else {
		__antithesis_instrumentation__.Notify(239645)
	}
	__antithesis_instrumentation__.Notify(239643)
	if err := d.set(ctx, sv, d.defaultValue); err != nil {
		__antithesis_instrumentation__.Notify(239646)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(239647)
	}
}

func (d *DurationSetting) WithPublic() *DurationSetting {
	__antithesis_instrumentation__.Notify(239648)
	d.SetVisibility(Public)
	return d
}

func RegisterDurationSetting(
	class Class,
	key, desc string,
	defaultValue time.Duration,
	validateFns ...func(time.Duration) error,
) *DurationSetting {
	__antithesis_instrumentation__.Notify(239649)
	var validateFn func(time.Duration) error
	if len(validateFns) > 0 {
		__antithesis_instrumentation__.Notify(239652)
		validateFn = func(v time.Duration) error {
			__antithesis_instrumentation__.Notify(239653)
			for _, fn := range validateFns {
				__antithesis_instrumentation__.Notify(239655)
				if err := fn(v); err != nil {
					__antithesis_instrumentation__.Notify(239656)
					return errors.Wrapf(err, "invalid value for %s", key)
				} else {
					__antithesis_instrumentation__.Notify(239657)
				}
			}
			__antithesis_instrumentation__.Notify(239654)
			return nil
		}
	} else {
		__antithesis_instrumentation__.Notify(239658)
	}
	__antithesis_instrumentation__.Notify(239650)

	if validateFn != nil {
		__antithesis_instrumentation__.Notify(239659)
		if err := validateFn(defaultValue); err != nil {
			__antithesis_instrumentation__.Notify(239660)
			panic(errors.Wrap(err, "invalid default"))
		} else {
			__antithesis_instrumentation__.Notify(239661)
		}
	} else {
		__antithesis_instrumentation__.Notify(239662)
	}
	__antithesis_instrumentation__.Notify(239651)
	setting := &DurationSetting{
		defaultValue: defaultValue,
		validateFn:   validateFn,
	}
	register(class, key, desc, setting)
	return setting
}

func RegisterPublicDurationSettingWithExplicitUnit(
	class Class, key, desc string, defaultValue time.Duration, validateFn func(time.Duration) error,
) *DurationSettingWithExplicitUnit {
	__antithesis_instrumentation__.Notify(239663)
	var fn func(time.Duration) error

	if validateFn != nil {
		__antithesis_instrumentation__.Notify(239665)
		fn = func(v time.Duration) error {
			__antithesis_instrumentation__.Notify(239666)
			return errors.Wrapf(validateFn(v), "invalid value for %s", key)
		}
	} else {
		__antithesis_instrumentation__.Notify(239667)
	}
	__antithesis_instrumentation__.Notify(239664)

	setting := &DurationSettingWithExplicitUnit{
		DurationSetting{
			defaultValue: defaultValue,
			validateFn:   fn,
		},
	}
	setting.SetVisibility(Public)
	register(class, key, desc, setting)
	return setting
}

func NonNegativeDuration(v time.Duration) error {
	__antithesis_instrumentation__.Notify(239668)
	if v < 0 {
		__antithesis_instrumentation__.Notify(239670)
		return errors.Errorf("cannot be set to a negative duration: %s", v)
	} else {
		__antithesis_instrumentation__.Notify(239671)
	}
	__antithesis_instrumentation__.Notify(239669)
	return nil
}

func NonNegativeDurationWithMaximum(maxValue time.Duration) func(time.Duration) error {
	__antithesis_instrumentation__.Notify(239672)
	return func(v time.Duration) error {
		__antithesis_instrumentation__.Notify(239673)
		if v < 0 {
			__antithesis_instrumentation__.Notify(239676)
			return errors.Errorf("cannot be set to a negative duration: %s", v)
		} else {
			__antithesis_instrumentation__.Notify(239677)
		}
		__antithesis_instrumentation__.Notify(239674)
		if v > maxValue {
			__antithesis_instrumentation__.Notify(239678)
			return errors.Errorf("cannot be set to a value larger than %s", maxValue)
		} else {
			__antithesis_instrumentation__.Notify(239679)
		}
		__antithesis_instrumentation__.Notify(239675)
		return nil
	}
}

func PositiveDuration(v time.Duration) error {
	__antithesis_instrumentation__.Notify(239680)
	if v <= 0 {
		__antithesis_instrumentation__.Notify(239682)
		return errors.Errorf("cannot be set to a non-positive duration: %s", v)
	} else {
		__antithesis_instrumentation__.Notify(239683)
	}
	__antithesis_instrumentation__.Notify(239681)
	return nil
}
