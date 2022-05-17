package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"strconv"

	"github.com/cockroachdb/errors"
)

type FloatSetting struct {
	common
	defaultValue float64
	validateFn   func(float64) error
}

var _ internalSetting = &FloatSetting{}

func (f *FloatSetting) Get(sv *Values) float64 {
	__antithesis_instrumentation__.Notify(239922)
	return math.Float64frombits(uint64(sv.getInt64(f.slot)))
}

func (f *FloatSetting) String(sv *Values) string {
	__antithesis_instrumentation__.Notify(239923)
	return EncodeFloat(f.Get(sv))
}

func (f *FloatSetting) Encoded(sv *Values) string {
	__antithesis_instrumentation__.Notify(239924)
	return f.String(sv)
}

func (f *FloatSetting) EncodedDefault() string {
	__antithesis_instrumentation__.Notify(239925)
	return EncodeFloat(f.defaultValue)
}

func (f *FloatSetting) DecodeToString(encoded string) (string, error) {
	__antithesis_instrumentation__.Notify(239926)
	fv, err := f.DecodeValue(encoded)
	if err != nil {
		__antithesis_instrumentation__.Notify(239928)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(239929)
	}
	__antithesis_instrumentation__.Notify(239927)
	return EncodeFloat(fv), nil
}

func (f *FloatSetting) DecodeValue(encoded string) (float64, error) {
	__antithesis_instrumentation__.Notify(239930)
	return strconv.ParseFloat(encoded, 64)
}

func (*FloatSetting) Typ() string {
	__antithesis_instrumentation__.Notify(239931)
	return "f"
}

func (f *FloatSetting) Default() float64 {
	__antithesis_instrumentation__.Notify(239932)
	return f.defaultValue
}

var _ = (*FloatSetting).Default

func (f *FloatSetting) Override(ctx context.Context, sv *Values, v float64) {
	__antithesis_instrumentation__.Notify(239933)
	if err := f.set(ctx, sv, v); err != nil {
		__antithesis_instrumentation__.Notify(239935)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(239936)
	}
	__antithesis_instrumentation__.Notify(239934)
	sv.setDefaultOverride(f.slot, v)
}

func (f *FloatSetting) Validate(v float64) error {
	__antithesis_instrumentation__.Notify(239937)
	if f.validateFn != nil {
		__antithesis_instrumentation__.Notify(239939)
		if err := f.validateFn(v); err != nil {
			__antithesis_instrumentation__.Notify(239940)
			return err
		} else {
			__antithesis_instrumentation__.Notify(239941)
		}
	} else {
		__antithesis_instrumentation__.Notify(239942)
	}
	__antithesis_instrumentation__.Notify(239938)
	return nil
}

func (f *FloatSetting) set(ctx context.Context, sv *Values, v float64) error {
	__antithesis_instrumentation__.Notify(239943)
	if err := f.Validate(v); err != nil {
		__antithesis_instrumentation__.Notify(239945)
		return err
	} else {
		__antithesis_instrumentation__.Notify(239946)
	}
	__antithesis_instrumentation__.Notify(239944)
	sv.setInt64(ctx, f.slot, int64(math.Float64bits(v)))
	return nil
}

func (f *FloatSetting) setToDefault(ctx context.Context, sv *Values) {
	__antithesis_instrumentation__.Notify(239947)

	if val := sv.getDefaultOverride(f.slot); val != nil {
		__antithesis_instrumentation__.Notify(239949)

		_ = f.set(ctx, sv, val.(float64))
		return
	} else {
		__antithesis_instrumentation__.Notify(239950)
	}
	__antithesis_instrumentation__.Notify(239948)
	if err := f.set(ctx, sv, f.defaultValue); err != nil {
		__antithesis_instrumentation__.Notify(239951)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(239952)
	}
}

func (f *FloatSetting) WithPublic() *FloatSetting {
	__antithesis_instrumentation__.Notify(239953)
	f.SetVisibility(Public)
	return f
}

func RegisterFloatSetting(
	class Class, key, desc string, defaultValue float64, validateFns ...func(float64) error,
) *FloatSetting {
	__antithesis_instrumentation__.Notify(239954)
	var validateFn func(float64) error
	if len(validateFns) > 0 {
		__antithesis_instrumentation__.Notify(239957)
		validateFn = func(v float64) error {
			__antithesis_instrumentation__.Notify(239958)
			for _, fn := range validateFns {
				__antithesis_instrumentation__.Notify(239960)
				if err := fn(v); err != nil {
					__antithesis_instrumentation__.Notify(239961)
					return errors.Wrapf(err, "invalid value for %s", key)
				} else {
					__antithesis_instrumentation__.Notify(239962)
				}
			}
			__antithesis_instrumentation__.Notify(239959)
			return nil
		}
	} else {
		__antithesis_instrumentation__.Notify(239963)
	}
	__antithesis_instrumentation__.Notify(239955)

	if validateFn != nil {
		__antithesis_instrumentation__.Notify(239964)
		if err := validateFn(defaultValue); err != nil {
			__antithesis_instrumentation__.Notify(239965)
			panic(errors.Wrap(err, "invalid default"))
		} else {
			__antithesis_instrumentation__.Notify(239966)
		}
	} else {
		__antithesis_instrumentation__.Notify(239967)
	}
	__antithesis_instrumentation__.Notify(239956)
	setting := &FloatSetting{
		defaultValue: defaultValue,
		validateFn:   validateFn,
	}
	register(class, key, desc, setting)
	return setting
}

func NonNegativeFloat(v float64) error {
	__antithesis_instrumentation__.Notify(239968)
	if v < 0 {
		__antithesis_instrumentation__.Notify(239970)
		return errors.Errorf("cannot set to a negative value: %f", v)
	} else {
		__antithesis_instrumentation__.Notify(239971)
	}
	__antithesis_instrumentation__.Notify(239969)
	return nil
}

func PositiveFloat(v float64) error {
	__antithesis_instrumentation__.Notify(239972)
	if v <= 0 {
		__antithesis_instrumentation__.Notify(239974)
		return errors.Errorf("cannot set to a non-positive value: %f", v)
	} else {
		__antithesis_instrumentation__.Notify(239975)
	}
	__antithesis_instrumentation__.Notify(239973)
	return nil
}
