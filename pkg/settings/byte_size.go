package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/errors"
)

type ByteSizeSetting struct {
	IntSetting
}

var _ numericSetting = &ByteSizeSetting{}

func (*ByteSizeSetting) Typ() string {
	__antithesis_instrumentation__.Notify(239563)
	return "z"
}

func (b *ByteSizeSetting) String(sv *Values) string {
	__antithesis_instrumentation__.Notify(239564)
	return string(humanizeutil.IBytes(b.Get(sv)))
}

func (b *ByteSizeSetting) DecodeToString(encoded string) (string, error) {
	__antithesis_instrumentation__.Notify(239565)
	iv, err := b.DecodeValue(encoded)
	if err != nil {
		__antithesis_instrumentation__.Notify(239567)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(239568)
	}
	__antithesis_instrumentation__.Notify(239566)
	return string(humanizeutil.IBytes(iv)), nil
}

func (b *ByteSizeSetting) WithPublic() *ByteSizeSetting {
	__antithesis_instrumentation__.Notify(239569)
	b.SetVisibility(Public)
	return b
}

func RegisterByteSizeSetting(
	class Class, key, desc string, defaultValue int64, validateFns ...func(int64) error,
) *ByteSizeSetting {
	__antithesis_instrumentation__.Notify(239570)

	var validateFn func(int64) error
	if len(validateFns) > 0 {
		__antithesis_instrumentation__.Notify(239573)
		validateFn = func(v int64) error {
			__antithesis_instrumentation__.Notify(239574)
			for _, fn := range validateFns {
				__antithesis_instrumentation__.Notify(239576)
				if err := fn(v); err != nil {
					__antithesis_instrumentation__.Notify(239577)
					return errors.Wrapf(err, "invalid value for %s", key)
				} else {
					__antithesis_instrumentation__.Notify(239578)
				}
			}
			__antithesis_instrumentation__.Notify(239575)
			return nil
		}
	} else {
		__antithesis_instrumentation__.Notify(239579)
	}
	__antithesis_instrumentation__.Notify(239571)

	if validateFn != nil {
		__antithesis_instrumentation__.Notify(239580)
		if err := validateFn(defaultValue); err != nil {
			__antithesis_instrumentation__.Notify(239581)
			panic(errors.Wrap(err, "invalid default"))
		} else {
			__antithesis_instrumentation__.Notify(239582)
		}
	} else {
		__antithesis_instrumentation__.Notify(239583)
	}
	__antithesis_instrumentation__.Notify(239572)
	setting := &ByteSizeSetting{IntSetting{
		defaultValue: defaultValue,
		validateFn:   validateFn,
	}}
	register(class, key, desc, setting)
	return setting
}
