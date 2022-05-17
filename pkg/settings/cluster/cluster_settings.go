package cluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type Settings struct {
	SV settings.Values

	Manual atomic.Value

	ExternalIODir string

	cpuProfiling int32

	Version clusterversion.Handle

	Cache sync.Map

	OverridesInformer OverridesInformer
}

type OverridesInformer interface {
	IsOverridden(settingName string) bool
}

func TelemetryOptOut() bool {
	__antithesis_instrumentation__.Notify(239584)
	return envutil.EnvOrDefaultBool("COCKROACH_SKIP_ENABLING_DIAGNOSTIC_REPORTING", false)
}

var NoSettings *Settings

type CPUProfileType int32

const (
	CPUProfileNone CPUProfileType = iota

	CPUProfileDefault

	CPUProfileWithLabels
)

func (s *Settings) CPUProfileType() CPUProfileType {
	__antithesis_instrumentation__.Notify(239585)
	return CPUProfileType(atomic.LoadInt32(&s.cpuProfiling))
}

func (s *Settings) SetCPUProfiling(to CPUProfileType) error {
	__antithesis_instrumentation__.Notify(239586)
	if to == CPUProfileNone {
		__antithesis_instrumentation__.Notify(239589)
		atomic.StoreInt32(&s.cpuProfiling, int32(CPUProfileNone))
	} else {
		__antithesis_instrumentation__.Notify(239590)
		if !atomic.CompareAndSwapInt32(&s.cpuProfiling, int32(CPUProfileNone), int32(to)) {
			__antithesis_instrumentation__.Notify(239591)
			return errors.New("a CPU profile is already in process, try again later")
		} else {
			__antithesis_instrumentation__.Notify(239592)
		}
	}
	__antithesis_instrumentation__.Notify(239587)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(239593)
		log.Infof(context.Background(), "active CPU profile type set to: %d", to)
	} else {
		__antithesis_instrumentation__.Notify(239594)
	}
	__antithesis_instrumentation__.Notify(239588)
	return nil
}

func (s *Settings) MakeUpdater() settings.Updater {
	__antithesis_instrumentation__.Notify(239595)
	if isManual, ok := s.Manual.Load().(bool); ok && func() bool {
		__antithesis_instrumentation__.Notify(239597)
		return isManual == true
	}() == true {
		__antithesis_instrumentation__.Notify(239598)
		return &settings.NoopUpdater{}
	} else {
		__antithesis_instrumentation__.Notify(239599)
	}
	__antithesis_instrumentation__.Notify(239596)
	return settings.NewUpdater(&s.SV)
}

func MakeClusterSettings() *Settings {
	__antithesis_instrumentation__.Notify(239600)
	s := &Settings{}

	sv := &s.SV
	s.Version = clusterversion.MakeVersionHandle(&s.SV)
	sv.Init(context.TODO(), s.Version)
	return s
}

func MakeTestingClusterSettings() *Settings {
	__antithesis_instrumentation__.Notify(239601)
	return MakeTestingClusterSettingsWithVersions(
		clusterversion.TestingBinaryVersion, clusterversion.TestingBinaryVersion, true)
}

func MakeTestingClusterSettingsWithVersions(
	binaryVersion, binaryMinSupportedVersion roachpb.Version, initializeVersion bool,
) *Settings {
	__antithesis_instrumentation__.Notify(239602)
	s := &Settings{}

	sv := &s.SV
	s.Version = clusterversion.MakeVersionHandleWithOverride(
		&s.SV, binaryVersion, binaryMinSupportedVersion)
	sv.Init(context.TODO(), s.Version)

	if initializeVersion {
		__antithesis_instrumentation__.Notify(239604)

		if err := clusterversion.Initialize(context.TODO(), binaryVersion, &s.SV); err != nil {
			__antithesis_instrumentation__.Notify(239605)
			log.Fatalf(context.TODO(), "unable to initialize version: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(239606)
		}
	} else {
		__antithesis_instrumentation__.Notify(239607)
	}
	__antithesis_instrumentation__.Notify(239603)
	return s
}
