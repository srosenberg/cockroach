package spec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/aws"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/azure"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/gce"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type fileSystemType int

const (
	Zfs fileSystemType = 1
)

type ClusterSpec struct {
	Cloud        string
	InstanceType string
	NodeCount    int

	CPUs           int
	SSDs           int
	RAID0          bool
	VolumeSize     int
	PreferLocalSSD bool
	Zones          string
	Geo            bool
	Lifetime       time.Duration
	ReusePolicy    clusterReusePolicy

	FileSystem fileSystemType

	RandomlyUseZfs bool
}

func MakeClusterSpec(cloud string, instanceType string, nodeCount int, opts ...Option) ClusterSpec {
	__antithesis_instrumentation__.Notify(44444)
	spec := ClusterSpec{Cloud: cloud, InstanceType: instanceType, NodeCount: nodeCount}
	defaultOpts := []Option{CPU(4), nodeLifetimeOption(12 * time.Hour), ReuseAny()}
	for _, o := range append(defaultOpts, opts...) {
		__antithesis_instrumentation__.Notify(44446)
		o.apply(&spec)
	}
	__antithesis_instrumentation__.Notify(44445)
	return spec
}

func ClustersCompatible(s1, s2 ClusterSpec) bool {
	__antithesis_instrumentation__.Notify(44447)
	s1.Lifetime = 0
	s2.Lifetime = 0
	return s1 == s2
}

func (s ClusterSpec) String() string {
	__antithesis_instrumentation__.Notify(44448)
	str := fmt.Sprintf("n%dcpu%d", s.NodeCount, s.CPUs)
	if s.Geo {
		__antithesis_instrumentation__.Notify(44450)
		str += "-Geo"
	} else {
		__antithesis_instrumentation__.Notify(44451)
	}
	__antithesis_instrumentation__.Notify(44449)
	return str
}

func awsMachineSupportsSSD(machineType string) bool {
	__antithesis_instrumentation__.Notify(44452)
	typeAndSize := strings.Split(machineType, ".")
	if len(typeAndSize) == 2 {
		__antithesis_instrumentation__.Notify(44454)

		return strings.HasPrefix(typeAndSize[0], "i3") || func() bool {
			__antithesis_instrumentation__.Notify(44455)
			return strings.HasSuffix(typeAndSize[0], "d") == true
		}() == true
	} else {
		__antithesis_instrumentation__.Notify(44456)
	}
	__antithesis_instrumentation__.Notify(44453)
	return false
}

func getAWSOpts(machineType string, zones []string, localSSD bool) vm.ProviderOpts {
	__antithesis_instrumentation__.Notify(44457)
	opts := aws.DefaultProviderOpts()
	if localSSD {
		__antithesis_instrumentation__.Notify(44460)
		opts.SSDMachineType = machineType
	} else {
		__antithesis_instrumentation__.Notify(44461)
		opts.MachineType = machineType
	}
	__antithesis_instrumentation__.Notify(44458)
	if len(zones) != 0 {
		__antithesis_instrumentation__.Notify(44462)
		opts.CreateZones = zones
	} else {
		__antithesis_instrumentation__.Notify(44463)
	}
	__antithesis_instrumentation__.Notify(44459)
	return opts
}

func getGCEOpts(
	machineType string, zones []string, volumeSize, localSSDCount int, localSSD bool, RAID0 bool,
) vm.ProviderOpts {
	__antithesis_instrumentation__.Notify(44464)
	opts := gce.DefaultProviderOpts()
	opts.MachineType = machineType
	if volumeSize != 0 {
		__antithesis_instrumentation__.Notify(44468)
		opts.PDVolumeSize = volumeSize
	} else {
		__antithesis_instrumentation__.Notify(44469)
	}
	__antithesis_instrumentation__.Notify(44465)
	if len(zones) != 0 {
		__antithesis_instrumentation__.Notify(44470)
		opts.Zones = zones
	} else {
		__antithesis_instrumentation__.Notify(44471)
	}
	__antithesis_instrumentation__.Notify(44466)
	opts.SSDCount = localSSDCount
	if localSSD && func() bool {
		__antithesis_instrumentation__.Notify(44472)
		return localSSDCount > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(44473)

		opts.UseMultipleDisks = !RAID0
	} else {
		__antithesis_instrumentation__.Notify(44474)
	}
	__antithesis_instrumentation__.Notify(44467)
	return opts
}

func getAzureOpts(machineType string, zones []string) vm.ProviderOpts {
	__antithesis_instrumentation__.Notify(44475)
	opts := azure.DefaultProviderOpts()
	opts.MachineType = machineType
	if len(zones) != 0 {
		__antithesis_instrumentation__.Notify(44477)
		opts.Locations = zones
	} else {
		__antithesis_instrumentation__.Notify(44478)
	}
	__antithesis_instrumentation__.Notify(44476)
	return opts
}

func (s *ClusterSpec) RoachprodOpts(
	clusterName string, useIOBarrier bool,
) (vm.CreateOpts, vm.ProviderOpts, error) {
	__antithesis_instrumentation__.Notify(44479)

	createVMOpts := vm.DefaultCreateOpts()
	createVMOpts.ClusterName = clusterName
	if s.Lifetime != 0 {
		__antithesis_instrumentation__.Notify(44487)
		createVMOpts.Lifetime = s.Lifetime
	} else {
		__antithesis_instrumentation__.Notify(44488)
	}
	__antithesis_instrumentation__.Notify(44480)
	switch s.Cloud {
	case Local:
		__antithesis_instrumentation__.Notify(44489)
		createVMOpts.VMProviders = []string{s.Cloud}

		return createVMOpts, nil, nil
	case AWS, GCE, Azure:
		__antithesis_instrumentation__.Notify(44490)
		createVMOpts.VMProviders = []string{s.Cloud}
	default:
		__antithesis_instrumentation__.Notify(44491)
		return vm.CreateOpts{}, nil, errors.Errorf("unsupported cloud %v", s.Cloud)
	}
	__antithesis_instrumentation__.Notify(44481)

	if s.Cloud != GCE {
		__antithesis_instrumentation__.Notify(44492)
		if s.VolumeSize != 0 {
			__antithesis_instrumentation__.Notify(44494)
			return vm.CreateOpts{}, nil, errors.Errorf("specifying volume size is not yet supported on %s", s.Cloud)
		} else {
			__antithesis_instrumentation__.Notify(44495)
		}
		__antithesis_instrumentation__.Notify(44493)
		if s.SSDs != 0 {
			__antithesis_instrumentation__.Notify(44496)
			return vm.CreateOpts{}, nil, errors.Errorf("specifying SSD count is not yet supported on %s", s.Cloud)
		} else {
			__antithesis_instrumentation__.Notify(44497)
		}
	} else {
		__antithesis_instrumentation__.Notify(44498)
	}
	__antithesis_instrumentation__.Notify(44482)

	createVMOpts.GeoDistributed = s.Geo
	machineType := s.InstanceType
	ssdCount := s.SSDs
	if s.CPUs != 0 {
		__antithesis_instrumentation__.Notify(44499)

		if len(machineType) == 0 {
			__antithesis_instrumentation__.Notify(44501)

			switch s.Cloud {
			case AWS:
				__antithesis_instrumentation__.Notify(44502)
				machineType = AWSMachineType(s.CPUs)
			case GCE:
				__antithesis_instrumentation__.Notify(44503)
				machineType = GCEMachineType(s.CPUs)
			case Azure:
				__antithesis_instrumentation__.Notify(44504)
				machineType = AzureMachineType(s.CPUs)
			default:
				__antithesis_instrumentation__.Notify(44505)
			}
		} else {
			__antithesis_instrumentation__.Notify(44506)
		}
		__antithesis_instrumentation__.Notify(44500)

		if s.PreferLocalSSD && func() bool {
			__antithesis_instrumentation__.Notify(44507)
			return s.VolumeSize == 0 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(44508)
			return (s.Cloud != AWS || func() bool {
				__antithesis_instrumentation__.Notify(44509)
				return awsMachineSupportsSSD(machineType) == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(44510)

			if ssdCount == 0 {
				__antithesis_instrumentation__.Notify(44512)
				ssdCount = 1
			} else {
				__antithesis_instrumentation__.Notify(44513)
			}
			__antithesis_instrumentation__.Notify(44511)
			createVMOpts.SSDOpts.UseLocalSSD = true
			createVMOpts.SSDOpts.NoExt4Barrier = !useIOBarrier
		} else {
			__antithesis_instrumentation__.Notify(44514)
			createVMOpts.SSDOpts.UseLocalSSD = false
		}
	} else {
		__antithesis_instrumentation__.Notify(44515)
	}
	__antithesis_instrumentation__.Notify(44483)

	if s.FileSystem == Zfs {
		__antithesis_instrumentation__.Notify(44516)
		if s.Cloud != GCE {
			__antithesis_instrumentation__.Notify(44518)
			return vm.CreateOpts{}, nil, errors.Errorf(
				"node creation with zfs file system not yet supported on %s", s.Cloud,
			)
		} else {
			__antithesis_instrumentation__.Notify(44519)
		}
		__antithesis_instrumentation__.Notify(44517)
		createVMOpts.SSDOpts.FileSystem = vm.Zfs
	} else {
		__antithesis_instrumentation__.Notify(44520)
		if s.RandomlyUseZfs && func() bool {
			__antithesis_instrumentation__.Notify(44521)
			return s.Cloud == GCE == true
		}() == true {
			__antithesis_instrumentation__.Notify(44522)
			rng, _ := randutil.NewPseudoRand()
			if rng.Float64() <= 0.2 {
				__antithesis_instrumentation__.Notify(44523)
				createVMOpts.SSDOpts.FileSystem = vm.Zfs
			} else {
				__antithesis_instrumentation__.Notify(44524)
			}
		} else {
			__antithesis_instrumentation__.Notify(44525)
		}
	}
	__antithesis_instrumentation__.Notify(44484)
	var zones []string
	if s.Zones != "" {
		__antithesis_instrumentation__.Notify(44526)
		zones = strings.Split(s.Zones, ",")
		if !s.Geo {
			__antithesis_instrumentation__.Notify(44527)
			zones = zones[:1]
		} else {
			__antithesis_instrumentation__.Notify(44528)
		}
	} else {
		__antithesis_instrumentation__.Notify(44529)
	}
	__antithesis_instrumentation__.Notify(44485)

	var providerOpts vm.ProviderOpts
	switch s.Cloud {
	case AWS:
		__antithesis_instrumentation__.Notify(44530)
		providerOpts = getAWSOpts(machineType, zones, createVMOpts.SSDOpts.UseLocalSSD)
	case GCE:
		__antithesis_instrumentation__.Notify(44531)
		providerOpts = getGCEOpts(machineType, zones, s.VolumeSize, ssdCount, createVMOpts.SSDOpts.UseLocalSSD, s.RAID0)
	case Azure:
		__antithesis_instrumentation__.Notify(44532)
		providerOpts = getAzureOpts(machineType, zones)
	default:
		__antithesis_instrumentation__.Notify(44533)
	}
	__antithesis_instrumentation__.Notify(44486)

	return createVMOpts, providerOpts, nil
}

func (s *ClusterSpec) Expiration() time.Time {
	__antithesis_instrumentation__.Notify(44534)
	l := s.Lifetime
	if l == 0 {
		__antithesis_instrumentation__.Notify(44536)
		l = 12 * time.Hour
	} else {
		__antithesis_instrumentation__.Notify(44537)
	}
	__antithesis_instrumentation__.Notify(44535)
	return timeutil.Now().Add(l)
}
