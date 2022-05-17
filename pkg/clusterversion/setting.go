package clusterversion

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

const KeyVersionSetting = "version"

var version = registerClusterVersionSetting()

type clusterVersionSetting struct {
	settings.VersionSetting
}

var _ settings.VersionSettingImpl = &clusterVersionSetting{}

func registerClusterVersionSetting() *clusterVersionSetting {
	__antithesis_instrumentation__.Notify(37270)
	s := &clusterVersionSetting{}
	s.VersionSetting = settings.MakeVersionSetting(s)
	settings.RegisterVersionSetting(
		settings.TenantWritable,
		KeyVersionSetting,
		"set the active cluster version in the format '<major>.<minor>'",
		&s.VersionSetting)
	s.SetVisibility(settings.Public)
	s.SetReportable(true)
	return s
}

func (cv *clusterVersionSetting) initialize(
	ctx context.Context, version roachpb.Version, sv *settings.Values,
) error {
	__antithesis_instrumentation__.Notify(37271)
	if ver := cv.activeVersionOrEmpty(ctx, sv); ver != (ClusterVersion{}) {
		__antithesis_instrumentation__.Notify(37275)

		if version.Less(ver.Version) {
			__antithesis_instrumentation__.Notify(37277)
			return errors.AssertionFailedf("cannot initialize version to %s because already set to: %s",
				version, ver)
		} else {
			__antithesis_instrumentation__.Notify(37278)
		}
		__antithesis_instrumentation__.Notify(37276)
		if version == ver.Version {
			__antithesis_instrumentation__.Notify(37279)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(37280)
		}

	} else {
		__antithesis_instrumentation__.Notify(37281)
	}
	__antithesis_instrumentation__.Notify(37272)
	if err := cv.validateBinaryVersions(version, sv); err != nil {
		__antithesis_instrumentation__.Notify(37282)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37283)
	}
	__antithesis_instrumentation__.Notify(37273)

	newV := ClusterVersion{Version: version}
	encoded, err := protoutil.Marshal(&newV)
	if err != nil {
		__antithesis_instrumentation__.Notify(37284)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37285)
	}
	__antithesis_instrumentation__.Notify(37274)
	cv.SetInternal(ctx, sv, encoded)
	return nil
}

func (cv *clusterVersionSetting) activeVersion(
	ctx context.Context, sv *settings.Values,
) ClusterVersion {
	__antithesis_instrumentation__.Notify(37286)
	ver := cv.activeVersionOrEmpty(ctx, sv)
	if ver == (ClusterVersion{}) {
		__antithesis_instrumentation__.Notify(37288)
		log.Fatalf(ctx, "version not initialized")
	} else {
		__antithesis_instrumentation__.Notify(37289)
	}
	__antithesis_instrumentation__.Notify(37287)
	return ver
}

func (cv *clusterVersionSetting) activeVersionOrEmpty(
	ctx context.Context, sv *settings.Values,
) ClusterVersion {
	__antithesis_instrumentation__.Notify(37290)
	encoded := cv.GetInternal(sv)
	if encoded == nil {
		__antithesis_instrumentation__.Notify(37293)
		return ClusterVersion{}
	} else {
		__antithesis_instrumentation__.Notify(37294)
	}
	__antithesis_instrumentation__.Notify(37291)
	var curVer ClusterVersion
	if err := protoutil.Unmarshal(encoded.([]byte), &curVer); err != nil {
		__antithesis_instrumentation__.Notify(37295)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(37296)
	}
	__antithesis_instrumentation__.Notify(37292)
	return curVer
}

func (cv *clusterVersionSetting) isActive(
	ctx context.Context, sv *settings.Values, versionKey Key,
) bool {
	__antithesis_instrumentation__.Notify(37297)
	return cv.activeVersion(ctx, sv).IsActive(versionKey)
}

func (cv *clusterVersionSetting) Decode(val []byte) (settings.ClusterVersionImpl, error) {
	__antithesis_instrumentation__.Notify(37298)
	var clusterVersion ClusterVersion
	if err := protoutil.Unmarshal(val, &clusterVersion); err != nil {
		__antithesis_instrumentation__.Notify(37300)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(37301)
	}
	__antithesis_instrumentation__.Notify(37299)
	return clusterVersion, nil
}

func (cv *clusterVersionSetting) ValidateVersionUpgrade(
	_ context.Context, sv *settings.Values, curRawProto, newRawProto []byte,
) error {
	__antithesis_instrumentation__.Notify(37302)
	var newCV ClusterVersion
	if err := protoutil.Unmarshal(newRawProto, &newCV); err != nil {
		__antithesis_instrumentation__.Notify(37308)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37309)
	}
	__antithesis_instrumentation__.Notify(37303)

	if err := cv.validateBinaryVersions(newCV.Version, sv); err != nil {
		__antithesis_instrumentation__.Notify(37310)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37311)
	}
	__antithesis_instrumentation__.Notify(37304)

	var oldCV ClusterVersion
	if err := protoutil.Unmarshal(curRawProto, &oldCV); err != nil {
		__antithesis_instrumentation__.Notify(37312)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37313)
	}
	__antithesis_instrumentation__.Notify(37305)

	if newCV.Version.Less(oldCV.Version) {
		__antithesis_instrumentation__.Notify(37314)
		return errors.Errorf(
			"versions cannot be downgraded (attempting to downgrade from %s to %s)",
			oldCV.Version, newCV.Version)
	} else {
		__antithesis_instrumentation__.Notify(37315)
	}
	__antithesis_instrumentation__.Notify(37306)

	if downgrade := preserveDowngradeVersion.Get(sv); downgrade != "" {
		__antithesis_instrumentation__.Notify(37316)
		return errors.Errorf(
			"cannot upgrade to %s: cluster.preserve_downgrade_option is set to %s",
			newCV.Version, downgrade)
	} else {
		__antithesis_instrumentation__.Notify(37317)
	}
	__antithesis_instrumentation__.Notify(37307)

	return nil
}

func (cv *clusterVersionSetting) ValidateBinaryVersions(
	ctx context.Context, sv *settings.Values, rawProto []byte,
) (retErr error) {
	__antithesis_instrumentation__.Notify(37318)
	defer func() {
		__antithesis_instrumentation__.Notify(37321)

		if retErr != nil {
			__antithesis_instrumentation__.Notify(37322)
			log.Fatalf(ctx, "failed to validate version upgrade: %s", retErr)
		} else {
			__antithesis_instrumentation__.Notify(37323)
		}
	}()
	__antithesis_instrumentation__.Notify(37319)

	var ver ClusterVersion
	if err := protoutil.Unmarshal(rawProto, &ver); err != nil {
		__antithesis_instrumentation__.Notify(37324)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37325)
	}
	__antithesis_instrumentation__.Notify(37320)
	return cv.validateBinaryVersions(ver.Version, sv)
}

func (cv *clusterVersionSetting) SettingsListDefault() string {
	__antithesis_instrumentation__.Notify(37326)
	return binaryVersion.String()
}

func (cv *clusterVersionSetting) validateBinaryVersions(
	ver roachpb.Version, sv *settings.Values,
) error {
	__antithesis_instrumentation__.Notify(37327)
	vh := sv.Opaque().(Handle)
	if vh.BinaryMinSupportedVersion() == (roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(37331)
		panic("BinaryMinSupportedVersion not set")
	} else {
		__antithesis_instrumentation__.Notify(37332)
	}
	__antithesis_instrumentation__.Notify(37328)
	if vh.BinaryVersion().Less(ver) {
		__antithesis_instrumentation__.Notify(37333)

		return errors.Errorf("cannot upgrade to %s: node running %s",
			ver, vh.BinaryVersion())
	} else {
		__antithesis_instrumentation__.Notify(37334)
	}
	__antithesis_instrumentation__.Notify(37329)
	if ver.Less(vh.BinaryMinSupportedVersion()) {
		__antithesis_instrumentation__.Notify(37335)
		return errors.Errorf("node at %s cannot run %s (minimum version is %s)",
			vh.BinaryVersion(), ver, vh.BinaryMinSupportedVersion())
	} else {
		__antithesis_instrumentation__.Notify(37336)
	}
	__antithesis_instrumentation__.Notify(37330)
	return nil
}

var preserveDowngradeVersion = registerPreserveDowngradeVersionSetting()

func registerPreserveDowngradeVersionSetting() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(37337)
	s := settings.RegisterValidatedStringSetting(
		settings.TenantWritable,
		"cluster.preserve_downgrade_option",
		"disable (automatic or manual) cluster version upgrade from the specified version until reset",
		"",
		func(sv *settings.Values, s string) error {
			__antithesis_instrumentation__.Notify(37339)
			if sv == nil || func() bool {
				__antithesis_instrumentation__.Notify(37343)
				return s == "" == true
			}() == true {
				__antithesis_instrumentation__.Notify(37344)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(37345)
			}
			__antithesis_instrumentation__.Notify(37340)
			clusterVersion := version.activeVersion(context.TODO(), sv).Version
			downgradeVersion, err := roachpb.ParseVersion(s)
			if err != nil {
				__antithesis_instrumentation__.Notify(37346)
				return err
			} else {
				__antithesis_instrumentation__.Notify(37347)
			}
			__antithesis_instrumentation__.Notify(37341)

			if downgradeVersion != clusterVersion {
				__antithesis_instrumentation__.Notify(37348)
				return errors.Errorf(
					"cannot set cluster.preserve_downgrade_option to %s (cluster version is %s)",
					s, clusterVersion)
			} else {
				__antithesis_instrumentation__.Notify(37349)
			}
			__antithesis_instrumentation__.Notify(37342)
			return nil
		},
	)
	__antithesis_instrumentation__.Notify(37338)
	s.SetReportable(true)
	s.SetVisibility(settings.Public)
	return s
}
