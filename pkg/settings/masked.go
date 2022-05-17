package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type MaskedSetting struct {
	setting NonMaskedSetting
}

var _ Setting = &MaskedSetting{}

func (s *MaskedSetting) UnderlyingSetting() NonMaskedSetting {
	__antithesis_instrumentation__.Notify(240027)
	return s.setting
}

func (s *MaskedSetting) String(sv *Values) string {
	__antithesis_instrumentation__.Notify(240028)

	if st, ok := s.UnderlyingSetting().(*StringSetting); ok && func() bool {
		__antithesis_instrumentation__.Notify(240030)
		return st.String(sv) == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(240031)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(240032)
	}
	__antithesis_instrumentation__.Notify(240029)
	return "<redacted>"
}

func (s *MaskedSetting) Visibility() Visibility {
	__antithesis_instrumentation__.Notify(240033)
	return s.setting.Visibility()
}

func (s *MaskedSetting) Key() string {
	__antithesis_instrumentation__.Notify(240034)
	return s.setting.Key()
}

func (s *MaskedSetting) Description() string {
	__antithesis_instrumentation__.Notify(240035)
	return s.setting.Description()
}

func (s *MaskedSetting) Typ() string {
	__antithesis_instrumentation__.Notify(240036)
	return s.setting.Typ()
}

func (s *MaskedSetting) Class() Class {
	__antithesis_instrumentation__.Notify(240037)
	return s.setting.Class()
}

func TestingIsReportable(s Setting) bool {
	__antithesis_instrumentation__.Notify(240038)
	if _, ok := s.(*MaskedSetting); ok {
		__antithesis_instrumentation__.Notify(240041)
		return false
	} else {
		__antithesis_instrumentation__.Notify(240042)
	}
	__antithesis_instrumentation__.Notify(240039)
	if e, ok := s.(internalSetting); ok {
		__antithesis_instrumentation__.Notify(240043)
		return e.isReportable()
	} else {
		__antithesis_instrumentation__.Notify(240044)
	}
	__antithesis_instrumentation__.Notify(240040)
	return true
}
