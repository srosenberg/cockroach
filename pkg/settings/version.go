package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
)

type VersionSetting struct {
	impl VersionSettingImpl
	common
}

var _ Setting = &VersionSetting{}

type VersionSettingImpl interface {
	Decode(val []byte) (ClusterVersionImpl, error)

	ValidateVersionUpgrade(ctx context.Context, sv *Values, oldV, newV []byte) error

	ValidateBinaryVersions(ctx context.Context, sv *Values, newV []byte) error

	SettingsListDefault() string
}

type ClusterVersionImpl interface {
	ClusterVersionImpl()

	fmt.Stringer
}

func MakeVersionSetting(impl VersionSettingImpl) VersionSetting {
	__antithesis_instrumentation__.Notify(240220)
	return VersionSetting{impl: impl}
}

func (v *VersionSetting) Decode(val []byte) (ClusterVersionImpl, error) {
	__antithesis_instrumentation__.Notify(240221)
	return v.impl.Decode(val)
}

func (v *VersionSetting) Validate(ctx context.Context, sv *Values, oldV, newV []byte) error {
	__antithesis_instrumentation__.Notify(240222)
	return v.impl.ValidateVersionUpgrade(ctx, sv, oldV, newV)
}

func (v *VersionSetting) SettingsListDefault() string {
	__antithesis_instrumentation__.Notify(240223)
	return v.impl.SettingsListDefault()
}

func (*VersionSetting) Typ() string {
	__antithesis_instrumentation__.Notify(240224)

	return "m"
}

func (v *VersionSetting) String(sv *Values) string {
	__antithesis_instrumentation__.Notify(240225)
	encV := []byte(v.Get(sv))
	if encV == nil {
		__antithesis_instrumentation__.Notify(240228)
		panic("unexpected nil value")
	} else {
		__antithesis_instrumentation__.Notify(240229)
	}
	__antithesis_instrumentation__.Notify(240226)
	cv, err := v.impl.Decode(encV)
	if err != nil {
		__antithesis_instrumentation__.Notify(240230)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(240231)
	}
	__antithesis_instrumentation__.Notify(240227)
	return cv.String()
}

func (v *VersionSetting) Encoded(sv *Values) string {
	__antithesis_instrumentation__.Notify(240232)
	return v.Get(sv)
}

func (v *VersionSetting) EncodedDefault() string {
	__antithesis_instrumentation__.Notify(240233)
	return encodedDefaultVersion
}

const encodedDefaultVersion = "unsupported"

func (v *VersionSetting) DecodeToString(encoded string) (string, error) {
	__antithesis_instrumentation__.Notify(240234)
	if encoded == encodedDefaultVersion {
		__antithesis_instrumentation__.Notify(240237)
		return encodedDefaultVersion, nil
	} else {
		__antithesis_instrumentation__.Notify(240238)
	}
	__antithesis_instrumentation__.Notify(240235)
	cv, err := v.impl.Decode([]byte(encoded))
	if err != nil {
		__antithesis_instrumentation__.Notify(240239)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(240240)
	}
	__antithesis_instrumentation__.Notify(240236)
	return cv.String(), nil
}

func (v *VersionSetting) Get(sv *Values) string {
	__antithesis_instrumentation__.Notify(240241)
	encV := v.GetInternal(sv)
	if encV == nil {
		__antithesis_instrumentation__.Notify(240243)
		panic(fmt.Sprintf("missing value for version setting in slot %d", v.slot))
	} else {
		__antithesis_instrumentation__.Notify(240244)
	}
	__antithesis_instrumentation__.Notify(240242)
	return string(encV.([]byte))
}

func (v *VersionSetting) GetInternal(sv *Values) interface{} {
	__antithesis_instrumentation__.Notify(240245)
	return sv.getGeneric(v.slot)
}

func (v *VersionSetting) SetInternal(ctx context.Context, sv *Values, newVal interface{}) {
	__antithesis_instrumentation__.Notify(240246)
	sv.setGeneric(ctx, v.slot, newVal)
}

func (v *VersionSetting) setToDefault(ctx context.Context, sv *Values) {
	__antithesis_instrumentation__.Notify(240247)
}

func RegisterVersionSetting(class Class, key, desc string, setting *VersionSetting) {
	__antithesis_instrumentation__.Notify(240248)
	register(class, key, desc, setting)
}
