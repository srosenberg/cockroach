package vm

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

const (
	TagCluster = "cluster"

	TagCreated = "created"

	TagLifetime = "lifetime"

	TagRoachprod = "roachprod"
)

func GetDefaultLabelMap(opts CreateOpts) map[string]string {
	__antithesis_instrumentation__.Notify(183876)
	return map[string]string{
		TagCluster:   opts.ClusterName,
		TagLifetime:  opts.Lifetime.String(),
		TagRoachprod: "true",
	}
}

type VM struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`

	Errors   []error           `json:"errors"`
	Lifetime time.Duration     `json:"lifetime"`
	Labels   map[string]string `json:"labels"`

	DNS string `json:"dns"`

	Provider string `json:"provider"`

	ProviderID string `json:"provider_id"`
	PrivateIP  string `json:"private_ip"`
	PublicIP   string `json:"public_ip"`

	RemoteUser string `json:"remote_user"`

	VPC         string `json:"vpc"`
	MachineType string `json:"machine_type"`
	Zone        string `json:"zone"`

	Project string `json:"project"`

	SQLPort int `json:"sql_port"`

	AdminUIPort int `json:"adminui_port"`

	LocalClusterName string `json:"local_cluster_name,omitempty"`
}

func Name(cluster string, idx int) string {
	__antithesis_instrumentation__.Notify(183877)
	return fmt.Sprintf("%s-%0.4d", cluster, idx)
}

var (
	ErrBadNetwork   = errors.New("could not determine network information")
	ErrInvalidName  = errors.New("invalid VM name")
	ErrNoExpiration = errors.New("could not determine expiration")
)

var regionRE = regexp.MustCompile(`(.*[^-])-?[a-z]$`)

func (vm *VM) IsLocal() bool {
	__antithesis_instrumentation__.Notify(183878)
	return vm.Zone == config.Local
}

func (vm *VM) Locality() (string, error) {
	__antithesis_instrumentation__.Notify(183879)
	var region string
	if vm.IsLocal() {
		__antithesis_instrumentation__.Notify(183881)
		region = vm.Zone
	} else {
		__antithesis_instrumentation__.Notify(183882)
		if match := regionRE.FindStringSubmatch(vm.Zone); len(match) == 2 {
			__antithesis_instrumentation__.Notify(183883)
			region = match[1]
		} else {
			__antithesis_instrumentation__.Notify(183884)
			return "", errors.Newf("unable to parse region from zone %q", vm.Zone)
		}
	}
	__antithesis_instrumentation__.Notify(183880)
	return fmt.Sprintf("cloud=%s,region=%s,zone=%s", vm.Provider, region, vm.Zone), nil
}

func (vm VM) ZoneEntry() (string, error) {
	__antithesis_instrumentation__.Notify(183885)
	if len(vm.Name) >= 60 {
		__antithesis_instrumentation__.Notify(183888)
		return "", errors.Errorf("Name too long: %s", vm.Name)
	} else {
		__antithesis_instrumentation__.Notify(183889)
	}
	__antithesis_instrumentation__.Notify(183886)
	if vm.PublicIP == "" {
		__antithesis_instrumentation__.Notify(183890)
		return "", errors.Errorf("Missing IP address: %s", vm.Name)
	} else {
		__antithesis_instrumentation__.Notify(183891)
	}
	__antithesis_instrumentation__.Notify(183887)

	return fmt.Sprintf("%s 60 IN A %s\n", vm.Name, vm.PublicIP), nil
}

type List []VM

func (vl List) Len() int { __antithesis_instrumentation__.Notify(183892); return len(vl) }
func (vl List) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(183893)
	vl[i], vl[j] = vl[j], vl[i]
}
func (vl List) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(183894)
	return vl[i].Name < vl[j].Name
}

func (vl List) Names() []string {
	__antithesis_instrumentation__.Notify(183895)
	ret := make([]string, len(vl))
	for i, vm := range vl {
		__antithesis_instrumentation__.Notify(183897)
		ret[i] = vm.Name
	}
	__antithesis_instrumentation__.Notify(183896)
	return ret
}

func (vl List) ProviderIDs() []string {
	__antithesis_instrumentation__.Notify(183898)
	ret := make([]string, len(vl))
	for i, vm := range vl {
		__antithesis_instrumentation__.Notify(183900)
		ret[i] = vm.ProviderID
	}
	__antithesis_instrumentation__.Notify(183899)
	return ret
}

const (
	Zfs = "zfs"

	Ext4 = "ext4"
)

type CreateOpts struct {
	ClusterName  string
	Lifetime     time.Duration
	CustomLabels map[string]string

	GeoDistributed bool
	VMProviders    []string
	SSDOpts        struct {
		UseLocalSSD bool

		NoExt4Barrier bool

		FileSystem string
	}
	OsVolumeSize int
}

func DefaultCreateOpts() CreateOpts {
	__antithesis_instrumentation__.Notify(183901)
	defaultCreateOpts := CreateOpts{
		ClusterName:    "",
		Lifetime:       12 * time.Hour,
		GeoDistributed: false,
		VMProviders:    []string{},
		OsVolumeSize:   10,
	}
	defaultCreateOpts.SSDOpts.UseLocalSSD = true
	defaultCreateOpts.SSDOpts.NoExt4Barrier = true
	defaultCreateOpts.SSDOpts.FileSystem = Ext4

	return defaultCreateOpts
}

type MultipleProjectsOption bool

const (
	SingleProject MultipleProjectsOption = false

	AcceptMultipleProjects = true
)

type ProviderOpts interface {
	ConfigureCreateFlags(*pflag.FlagSet)

	ConfigureClusterFlags(*pflag.FlagSet, MultipleProjectsOption)
}

type Provider interface {
	CreateProviderOpts() ProviderOpts
	CleanSSH() error

	ConfigSSH(zones []string) error
	Create(l *logger.Logger, names []string, opts CreateOpts, providerOpts ProviderOpts) error
	Reset(vms List) error
	Delete(vms List) error
	Extend(vms List, lifetime time.Duration) error

	FindActiveAccount() (string, error)
	List(l *logger.Logger) (List, error)

	Name() string

	Active() bool

	ProjectActive(project string) bool
}

type DeleteCluster interface {
	DeleteCluster(name string) error
}

var Providers = map[string]Provider{}

type ProviderOptionsContainer map[string]ProviderOpts

func CreateProviderOptionsContainer() ProviderOptionsContainer {
	__antithesis_instrumentation__.Notify(183902)
	container := make(ProviderOptionsContainer)
	for providerName, providerInstance := range Providers {
		__antithesis_instrumentation__.Notify(183904)
		container[providerName] = providerInstance.CreateProviderOpts()
	}
	__antithesis_instrumentation__.Notify(183903)
	return container
}

func (container ProviderOptionsContainer) SetProviderOpts(
	providerName string, providerOpts ProviderOpts,
) {
	__antithesis_instrumentation__.Notify(183905)
	container[providerName] = providerOpts
}

func AllProviderNames() []string {
	__antithesis_instrumentation__.Notify(183906)
	var ret []string
	for name := range Providers {
		__antithesis_instrumentation__.Notify(183908)
		ret = append(ret, name)
	}
	__antithesis_instrumentation__.Notify(183907)
	return ret
}

func FanOut(list List, action func(Provider, List) error) error {
	__antithesis_instrumentation__.Notify(183909)
	var m = map[string]List{}
	for _, vm := range list {
		__antithesis_instrumentation__.Notify(183912)
		m[vm.Provider] = append(m[vm.Provider], vm)
	}
	__antithesis_instrumentation__.Notify(183910)

	var g errgroup.Group
	for name, vms := range m {
		__antithesis_instrumentation__.Notify(183913)

		n := name
		v := vms
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(183914)
			p, ok := Providers[n]
			if !ok {
				__antithesis_instrumentation__.Notify(183916)
				return errors.Errorf("unknown provider name: %s", n)
			} else {
				__antithesis_instrumentation__.Notify(183917)
			}
			__antithesis_instrumentation__.Notify(183915)
			return action(p, v)
		})
	}
	__antithesis_instrumentation__.Notify(183911)

	return g.Wait()
}

var cachedActiveAccounts map[string]string

func FindActiveAccounts() (map[string]string, error) {
	__antithesis_instrumentation__.Notify(183918)
	source := cachedActiveAccounts

	if source == nil {
		__antithesis_instrumentation__.Notify(183921)

		source = map[string]string{}
		err := ProvidersSequential(AllProviderNames(), func(p Provider) error {
			__antithesis_instrumentation__.Notify(183924)
			account, err := p.FindActiveAccount()
			if err != nil {
				__antithesis_instrumentation__.Notify(183927)
				return err
			} else {
				__antithesis_instrumentation__.Notify(183928)
			}
			__antithesis_instrumentation__.Notify(183925)
			if len(account) > 0 {
				__antithesis_instrumentation__.Notify(183929)
				source[p.Name()] = account
			} else {
				__antithesis_instrumentation__.Notify(183930)
			}
			__antithesis_instrumentation__.Notify(183926)
			return nil
		})
		__antithesis_instrumentation__.Notify(183922)
		if err != nil {
			__antithesis_instrumentation__.Notify(183931)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(183932)
		}
		__antithesis_instrumentation__.Notify(183923)
		cachedActiveAccounts = source
	} else {
		__antithesis_instrumentation__.Notify(183933)
	}
	__antithesis_instrumentation__.Notify(183919)

	ret := make(map[string]string, len(source))
	for k, v := range source {
		__antithesis_instrumentation__.Notify(183934)
		ret[k] = v
	}
	__antithesis_instrumentation__.Notify(183920)

	return ret, nil
}

func ForProvider(named string, action func(Provider) error) error {
	__antithesis_instrumentation__.Notify(183935)
	p, ok := Providers[named]
	if !ok {
		__antithesis_instrumentation__.Notify(183938)
		return errors.Errorf("unknown vm provider: %s", named)
	} else {
		__antithesis_instrumentation__.Notify(183939)
	}
	__antithesis_instrumentation__.Notify(183936)
	if err := action(p); err != nil {
		__antithesis_instrumentation__.Notify(183940)
		return errors.Wrapf(err, "in provider: %s", named)
	} else {
		__antithesis_instrumentation__.Notify(183941)
	}
	__antithesis_instrumentation__.Notify(183937)
	return nil
}

func ProvidersParallel(named []string, action func(Provider) error) error {
	__antithesis_instrumentation__.Notify(183942)
	var g errgroup.Group
	for _, name := range named {
		__antithesis_instrumentation__.Notify(183944)

		n := name
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(183945)
			return ForProvider(n, action)
		})
	}
	__antithesis_instrumentation__.Notify(183943)
	return g.Wait()
}

func ProvidersSequential(named []string, action func(Provider) error) error {
	__antithesis_instrumentation__.Notify(183946)
	for _, name := range named {
		__antithesis_instrumentation__.Notify(183948)
		if err := ForProvider(name, action); err != nil {
			__antithesis_instrumentation__.Notify(183949)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183950)
		}
	}
	__antithesis_instrumentation__.Notify(183947)
	return nil
}

func ZonePlacement(numZones, numNodes int) (nodeZones []int) {
	__antithesis_instrumentation__.Notify(183951)
	if numZones < 1 {
		__antithesis_instrumentation__.Notify(183955)
		panic("expected 1 or more zones")
	} else {
		__antithesis_instrumentation__.Notify(183956)
	}
	__antithesis_instrumentation__.Notify(183952)
	numPerZone := numNodes / numZones
	if numPerZone < 1 {
		__antithesis_instrumentation__.Notify(183957)
		numPerZone = 1
	} else {
		__antithesis_instrumentation__.Notify(183958)
	}
	__antithesis_instrumentation__.Notify(183953)
	extraStartIndex := numPerZone * numZones
	nodeZones = make([]int, numNodes)
	for i := 0; i < numNodes; i++ {
		__antithesis_instrumentation__.Notify(183959)
		nodeZones[i] = i / numPerZone
		if i >= extraStartIndex {
			__antithesis_instrumentation__.Notify(183960)
			nodeZones[i] = i % numZones
		} else {
			__antithesis_instrumentation__.Notify(183961)
		}
	}
	__antithesis_instrumentation__.Notify(183954)
	return nodeZones
}

func ExpandZonesFlag(zoneFlag []string) (zones []string, err error) {
	__antithesis_instrumentation__.Notify(183962)
	for _, zone := range zoneFlag {
		__antithesis_instrumentation__.Notify(183964)
		colonIdx := strings.Index(zone, ":")
		if colonIdx == -1 {
			__antithesis_instrumentation__.Notify(183967)
			zones = append(zones, zone)
			continue
		} else {
			__antithesis_instrumentation__.Notify(183968)
		}
		__antithesis_instrumentation__.Notify(183965)
		n, err := strconv.Atoi(zone[colonIdx+1:])
		if err != nil {
			__antithesis_instrumentation__.Notify(183969)
			return zones, errors.Wrapf(err, "failed to parse %q", zone)
		} else {
			__antithesis_instrumentation__.Notify(183970)
		}
		__antithesis_instrumentation__.Notify(183966)
		for i := 0; i < n; i++ {
			__antithesis_instrumentation__.Notify(183971)
			zones = append(zones, zone[:colonIdx])
		}
	}
	__antithesis_instrumentation__.Notify(183963)
	return zones, nil
}

func DNSSafeAccount(account string) string {
	__antithesis_instrumentation__.Notify(183972)
	safe := func(r rune) rune {
		__antithesis_instrumentation__.Notify(183974)
		switch {
		case r >= 'a' && func() bool {
			__antithesis_instrumentation__.Notify(183978)
			return r <= 'z' == true
		}() == true:
			__antithesis_instrumentation__.Notify(183975)
			return r
		case r >= 'A' && func() bool {
			__antithesis_instrumentation__.Notify(183979)
			return r <= 'Z' == true
		}() == true:
			__antithesis_instrumentation__.Notify(183976)
			return unicode.ToLower(r)
		default:
			__antithesis_instrumentation__.Notify(183977)

			return -1
		}
	}
	__antithesis_instrumentation__.Notify(183973)
	return strings.Map(safe, account)
}
