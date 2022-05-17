// Package clusterversion defines the interfaces to interact with cluster/binary
// versions in order accommodate backward incompatible behaviors. It handles the
// feature gates and so must maintain a fairly lightweight set of dependencies.
// The migration sub-package handles advancing a cluster from one version to
// a later one.
//
// Ideally, every code change in a database would be backward compatible, but
// this is not always possible. Some features, fixes, or cleanups need to
// introduce a backward incompatibility and others are dramatically simplified by
// it. This package provides a way to do this safely with (hopefully) minimal
// disruption. It works as follows:
//
// - Each node in the cluster is running a binary that was released at some
//   version ("binary version"). We allow for rolling upgrades, so two nodes in
//   the cluster may be running different binary versions. All nodes in a given
//   cluster must be within 1 major release of each other (i.e. to upgrade two
//   major releases, the cluster must first be rolled onto X+1 and then to X+2).
// - Separate from the build versions of the binaries, the cluster itself has a
//   logical "active cluster version", the version all the binaries are
//   currently operating at. This is used for two related things: first as a
//   promise from the user that they'll never downgrade any nodes in the cluster
//   to a binary below some "minimum supported version", and second, to unlock
//   features that are not backwards compatible (which is now safe given that
//   the old binary will never be used).
// - Each binary can operate within a "range of supported versions". When a
// 	 cluster is initialized, the binary doing the initialization uses the upper
//	 end of its supported range as the initial "active cluster version". Each
//	 node that joins this cluster then must be compatible with this cluster
//	 version.
package clusterversion

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/redact"
)

func Initialize(ctx context.Context, ver roachpb.Version, sv *settings.Values) error {
	__antithesis_instrumentation__.Notify(37182)
	return version.initialize(ctx, ver, sv)
}

type Handle interface {
	ActiveVersion(context.Context) ClusterVersion

	ActiveVersionOrEmpty(context.Context) ClusterVersion

	IsActive(context.Context, Key) bool

	BinaryVersion() roachpb.Version

	BinaryMinSupportedVersion() roachpb.Version

	SetActiveVersion(context.Context, ClusterVersion) error

	SetOnChange(fn func(ctx context.Context, newVersion ClusterVersion))
}

type handleImpl struct {
	setting *clusterVersionSetting

	sv *settings.Values

	binaryVersion             roachpb.Version
	binaryMinSupportedVersion roachpb.Version
}

var _ Handle = (*handleImpl)(nil)

func MakeVersionHandle(sv *settings.Values) Handle {
	__antithesis_instrumentation__.Notify(37183)
	return MakeVersionHandleWithOverride(sv, binaryVersion, binaryMinSupportedVersion)
}

func MakeVersionHandleWithOverride(
	sv *settings.Values, binaryVersion, binaryMinSupportedVersion roachpb.Version,
) Handle {
	__antithesis_instrumentation__.Notify(37184)
	return newHandleImpl(version, sv, binaryVersion, binaryMinSupportedVersion)
}

func newHandleImpl(
	setting *clusterVersionSetting,
	sv *settings.Values,
	binaryVersion, binaryMinSupportedVersion roachpb.Version,
) Handle {
	__antithesis_instrumentation__.Notify(37185)
	return &handleImpl{
		setting:                   setting,
		sv:                        sv,
		binaryVersion:             binaryVersion,
		binaryMinSupportedVersion: binaryMinSupportedVersion,
	}
}

func (v *handleImpl) ActiveVersion(ctx context.Context) ClusterVersion {
	__antithesis_instrumentation__.Notify(37186)
	return v.setting.activeVersion(ctx, v.sv)
}

func (v *handleImpl) ActiveVersionOrEmpty(ctx context.Context) ClusterVersion {
	__antithesis_instrumentation__.Notify(37187)
	return v.setting.activeVersionOrEmpty(ctx, v.sv)
}

func (v *handleImpl) SetActiveVersion(ctx context.Context, cv ClusterVersion) error {
	__antithesis_instrumentation__.Notify(37188)

	if err := v.setting.validateBinaryVersions(cv.Version, v.sv); err != nil {
		__antithesis_instrumentation__.Notify(37191)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37192)
	}
	__antithesis_instrumentation__.Notify(37189)

	encoded, err := protoutil.Marshal(&cv)
	if err != nil {
		__antithesis_instrumentation__.Notify(37193)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37194)
	}
	__antithesis_instrumentation__.Notify(37190)

	v.setting.SetInternal(ctx, v.sv, encoded)
	return nil
}

func (v *handleImpl) SetOnChange(fn func(ctx context.Context, newVersion ClusterVersion)) {
	__antithesis_instrumentation__.Notify(37195)
	v.setting.SetOnChange(v.sv, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(37196)
		fn(ctx, v.ActiveVersion(ctx))
	})
}

func (v *handleImpl) IsActive(ctx context.Context, key Key) bool {
	__antithesis_instrumentation__.Notify(37197)
	return v.setting.isActive(ctx, v.sv, key)
}

func (v *handleImpl) BinaryVersion() roachpb.Version {
	__antithesis_instrumentation__.Notify(37198)
	return v.binaryVersion
}

func (v *handleImpl) BinaryMinSupportedVersion() roachpb.Version {
	__antithesis_instrumentation__.Notify(37199)
	return v.binaryMinSupportedVersion
}

func (cv ClusterVersion) IsActiveVersion(v roachpb.Version) bool {
	__antithesis_instrumentation__.Notify(37200)
	return !cv.Less(v)
}

func (cv ClusterVersion) IsActive(versionKey Key) bool {
	__antithesis_instrumentation__.Notify(37201)
	v := ByKey(versionKey)
	return cv.IsActiveVersion(v)
}

func (cv ClusterVersion) String() string {
	__antithesis_instrumentation__.Notify(37202)
	return redact.StringWithoutMarkers(cv)
}

func (cv ClusterVersion) SafeFormat(p redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(37203)
	p.Print(cv.Version)
}

func (cv ClusterVersion) PrettyPrint() string {
	__antithesis_instrumentation__.Notify(37204)

	fenceVersion := !cv.Version.LessEq(roachpb.Version{Major: 20, Minor: 2}) && func() bool {
		__antithesis_instrumentation__.Notify(37206)
		return (cv.Internal % 2) == 1 == true
	}() == true
	if !fenceVersion {
		__antithesis_instrumentation__.Notify(37207)
		return cv.String()
	} else {
		__antithesis_instrumentation__.Notify(37208)
	}
	__antithesis_instrumentation__.Notify(37205)
	return redact.Sprintf("%s%s", cv.String(), "(fence)").StripMarkers()
}

func (cv ClusterVersion) ClusterVersionImpl() { __antithesis_instrumentation__.Notify(37209) }

var _ settings.ClusterVersionImpl = ClusterVersion{}

func EncodingFromVersionStr(v string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(37210)
	newV, err := roachpb.ParseVersion(v)
	if err != nil {
		__antithesis_instrumentation__.Notify(37212)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(37213)
	}
	__antithesis_instrumentation__.Notify(37211)
	newCV := ClusterVersion{Version: newV}
	return protoutil.Marshal(&newCV)
}
