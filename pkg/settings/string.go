package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/errors"
)

type StringSetting struct {
	defaultValue string
	validateFn   func(*Values, string) error
	common
}

var _ internalSetting = &StringSetting{}

func (s *StringSetting) String(sv *Values) string {
	__antithesis_instrumentation__.Notify(240097)
	return s.Get(sv)
}

func (s *StringSetting) Encoded(sv *Values) string {
	__antithesis_instrumentation__.Notify(240098)
	return s.String(sv)
}

func (s *StringSetting) EncodedDefault() string {
	__antithesis_instrumentation__.Notify(240099)
	return s.defaultValue
}

func (s *StringSetting) DecodeToString(encoded string) (string, error) {
	__antithesis_instrumentation__.Notify(240100)
	return encoded, nil
}

func (*StringSetting) Typ() string {
	__antithesis_instrumentation__.Notify(240101)
	return "s"
}

func (s *StringSetting) Default() string {
	__antithesis_instrumentation__.Notify(240102)
	return s.defaultValue
}

var _ = (*StringSetting).Default

func (s *StringSetting) Get(sv *Values) string {
	__antithesis_instrumentation__.Notify(240103)
	loaded := sv.getGeneric(s.slot)
	if loaded == nil {
		__antithesis_instrumentation__.Notify(240105)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(240106)
	}
	__antithesis_instrumentation__.Notify(240104)
	return loaded.(string)
}

func (s *StringSetting) Validate(sv *Values, v string) error {
	__antithesis_instrumentation__.Notify(240107)
	if s.validateFn != nil {
		__antithesis_instrumentation__.Notify(240109)
		if err := s.validateFn(sv, v); err != nil {
			__antithesis_instrumentation__.Notify(240110)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240111)
		}
	} else {
		__antithesis_instrumentation__.Notify(240112)
	}
	__antithesis_instrumentation__.Notify(240108)
	return nil
}

func (s *StringSetting) Override(ctx context.Context, sv *Values, v string) {
	__antithesis_instrumentation__.Notify(240113)
	_ = s.set(ctx, sv, v)
}

func (s *StringSetting) set(ctx context.Context, sv *Values, v string) error {
	__antithesis_instrumentation__.Notify(240114)
	if err := s.Validate(sv, v); err != nil {
		__antithesis_instrumentation__.Notify(240117)
		return err
	} else {
		__antithesis_instrumentation__.Notify(240118)
	}
	__antithesis_instrumentation__.Notify(240115)
	if s.Get(sv) != v {
		__antithesis_instrumentation__.Notify(240119)
		sv.setGeneric(ctx, s.slot, v)
	} else {
		__antithesis_instrumentation__.Notify(240120)
	}
	__antithesis_instrumentation__.Notify(240116)
	return nil
}

func (s *StringSetting) setToDefault(ctx context.Context, sv *Values) {
	__antithesis_instrumentation__.Notify(240121)
	if err := s.set(ctx, sv, s.defaultValue); err != nil {
		__antithesis_instrumentation__.Notify(240122)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(240123)
	}
}

func (s *StringSetting) WithPublic() *StringSetting {
	__antithesis_instrumentation__.Notify(240124)
	s.SetVisibility(Public)
	return s
}

func RegisterStringSetting(class Class, key, desc string, defaultValue string) *StringSetting {
	__antithesis_instrumentation__.Notify(240125)
	return RegisterValidatedStringSetting(class, key, desc, defaultValue, nil)
}

func RegisterValidatedStringSetting(
	class Class, key, desc string, defaultValue string, validateFn func(*Values, string) error,
) *StringSetting {
	__antithesis_instrumentation__.Notify(240126)
	if validateFn != nil {
		__antithesis_instrumentation__.Notify(240128)
		if err := validateFn(nil, defaultValue); err != nil {
			__antithesis_instrumentation__.Notify(240129)
			panic(errors.Wrap(err, "invalid default"))
		} else {
			__antithesis_instrumentation__.Notify(240130)
		}
	} else {
		__antithesis_instrumentation__.Notify(240131)
	}
	__antithesis_instrumentation__.Notify(240127)
	setting := &StringSetting{
		defaultValue: defaultValue,
		validateFn:   validateFn,
	}

	setting.SetReportable(false)
	register(class, key, desc, setting)
	return setting
}
