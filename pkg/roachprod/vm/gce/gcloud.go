package gce

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/flagstub"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

const (
	defaultProject = "cockroach-ephemeral"

	ProviderName = "gce"
)

var providerInstance = &Provider{}

func DefaultProject() string {
	__antithesis_instrumentation__.Notify(183575)
	return defaultProject
}

var projectsWithGC = []string{defaultProject, "andrei-jepsen"}

func Init() error {
	__antithesis_instrumentation__.Notify(183576)
	providerInstance.Projects = []string{defaultProject}
	projectFromEnv := os.Getenv("GCE_PROJECT")
	if projectFromEnv != "" {
		__antithesis_instrumentation__.Notify(183579)
		providerInstance.Projects = []string{projectFromEnv}
	} else {
		__antithesis_instrumentation__.Notify(183580)
	}
	__antithesis_instrumentation__.Notify(183577)
	providerInstance.ServiceAccount = os.Getenv("GCE_SERVICE_ACCOUNT")
	if _, err := exec.LookPath("gcloud"); err != nil {
		__antithesis_instrumentation__.Notify(183581)
		vm.Providers[ProviderName] = flagstub.New(&Provider{}, "please install the gcloud CLI utilities "+
			"(https://cloud.google.com/sdk/downloads)")
		return errors.New("gcloud not found")
	} else {
		__antithesis_instrumentation__.Notify(183582)
	}
	__antithesis_instrumentation__.Notify(183578)
	vm.Providers[ProviderName] = providerInstance
	return nil
}

func runJSONCommand(args []string, parsed interface{}) error {
	__antithesis_instrumentation__.Notify(183583)
	cmd := exec.Command("gcloud", args...)

	rawJSON, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(183586)
		var stderr []byte
		if exitErr := (*exec.ExitError)(nil); errors.As(err, &exitErr) {
			__antithesis_instrumentation__.Notify(183588)
			stderr = exitErr.Stderr
		} else {
			__antithesis_instrumentation__.Notify(183589)
		}
		__antithesis_instrumentation__.Notify(183587)

		if matched, _ := regexp.Match(`.*Unknown zone`, stderr); !matched {
			__antithesis_instrumentation__.Notify(183590)
			return errors.Wrapf(err, "failed to run: gcloud %s\nstdout: %s\nstderr: %s\n",
				strings.Join(args, " "), bytes.TrimSpace(rawJSON), bytes.TrimSpace(stderr))
		} else {
			__antithesis_instrumentation__.Notify(183591)
		}
	} else {
		__antithesis_instrumentation__.Notify(183592)
	}
	__antithesis_instrumentation__.Notify(183584)

	if err := json.Unmarshal(rawJSON, &parsed); err != nil {
		__antithesis_instrumentation__.Notify(183593)
		return errors.Wrapf(err, "failed to parse json %s", rawJSON)
	} else {
		__antithesis_instrumentation__.Notify(183594)
	}
	__antithesis_instrumentation__.Notify(183585)

	return nil
}

type jsonVM struct {
	Name              string
	Labels            map[string]string
	CreationTimestamp time.Time
	NetworkInterfaces []struct {
		Network       string
		NetworkIP     string
		AccessConfigs []struct {
			Name  string
			NatIP string
		}
	}
	MachineType string
	Zone        string
}

func (jsonVM *jsonVM) toVM(project string, opts *ProviderOpts) (ret *vm.VM) {
	__antithesis_instrumentation__.Notify(183595)
	var vmErrors []error
	var err error

	var lifetime time.Duration
	if lifetimeStr, ok := jsonVM.Labels["lifetime"]; ok {
		__antithesis_instrumentation__.Notify(183600)
		if lifetime, err = time.ParseDuration(lifetimeStr); err != nil {
			__antithesis_instrumentation__.Notify(183601)
			vmErrors = append(vmErrors, vm.ErrNoExpiration)
		} else {
			__antithesis_instrumentation__.Notify(183602)
		}
	} else {
		__antithesis_instrumentation__.Notify(183603)
		vmErrors = append(vmErrors, vm.ErrNoExpiration)
	}
	__antithesis_instrumentation__.Notify(183596)

	lastComponent := func(url string) string {
		__antithesis_instrumentation__.Notify(183604)
		s := strings.Split(url, "/")
		return s[len(s)-1]
	}
	__antithesis_instrumentation__.Notify(183597)

	var publicIP, privateIP, vpc string
	if len(jsonVM.NetworkInterfaces) == 0 {
		__antithesis_instrumentation__.Notify(183605)
		vmErrors = append(vmErrors, vm.ErrBadNetwork)
	} else {
		__antithesis_instrumentation__.Notify(183606)
		privateIP = jsonVM.NetworkInterfaces[0].NetworkIP
		if len(jsonVM.NetworkInterfaces[0].AccessConfigs) == 0 {
			__antithesis_instrumentation__.Notify(183607)
			vmErrors = append(vmErrors, vm.ErrBadNetwork)
		} else {
			__antithesis_instrumentation__.Notify(183608)
			_ = jsonVM.NetworkInterfaces[0].AccessConfigs[0].Name
			publicIP = jsonVM.NetworkInterfaces[0].AccessConfigs[0].NatIP
			vpc = lastComponent(jsonVM.NetworkInterfaces[0].Network)
		}
	}
	__antithesis_instrumentation__.Notify(183598)

	machineType := lastComponent(jsonVM.MachineType)
	zone := lastComponent(jsonVM.Zone)
	remoteUser := config.SharedUser
	if !opts.useSharedUser {
		__antithesis_instrumentation__.Notify(183609)

		remoteUser = config.OSUser.Username
	} else {
		__antithesis_instrumentation__.Notify(183610)
	}
	__antithesis_instrumentation__.Notify(183599)
	return &vm.VM{
		Name:        jsonVM.Name,
		CreatedAt:   jsonVM.CreationTimestamp,
		Errors:      vmErrors,
		DNS:         fmt.Sprintf("%s.%s.%s", jsonVM.Name, zone, project),
		Lifetime:    lifetime,
		Labels:      jsonVM.Labels,
		PrivateIP:   privateIP,
		Provider:    ProviderName,
		ProviderID:  jsonVM.Name,
		PublicIP:    publicIP,
		RemoteUser:  remoteUser,
		VPC:         vpc,
		MachineType: machineType,
		Zone:        zone,
		Project:     project,
		SQLPort:     config.DefaultSQLPort,
		AdminUIPort: config.DefaultAdminUIPort,
	}
}

type jsonAuth struct {
	Account string
	Status  string
}

func DefaultProviderOpts() *ProviderOpts {
	__antithesis_instrumentation__.Notify(183611)
	return &ProviderOpts{

		MachineType:    "n1-standard-4",
		MinCPUPlatform: "",
		Zones:          nil,
		Image:          "ubuntu-2004-focal-v20210603",
		SSDCount:       1,
		PDVolumeType:   "pd-ssd",
		PDVolumeSize:   500,
		useSharedUser:  true,
		preemptible:    false,
	}
}

func (p *Provider) CreateProviderOpts() vm.ProviderOpts {
	__antithesis_instrumentation__.Notify(183612)
	return DefaultProviderOpts()
}

type ProviderOpts struct {
	MachineType      string
	MinCPUPlatform   string
	Zones            []string
	Image            string
	SSDCount         int
	PDVolumeType     string
	PDVolumeSize     int
	UseMultipleDisks bool

	useSharedUser bool

	preemptible bool
}

type Provider struct {
	Projects       []string
	ServiceAccount string
}

type ProjectsVal struct {
	AcceptMultipleProjects bool
}

var defaultZones = []string{
	"us-east1-b",
	"us-west1-b",
	"europe-west2-b",
}

func (v ProjectsVal) Set(projects string) error {
	__antithesis_instrumentation__.Notify(183613)
	if projects == "" {
		__antithesis_instrumentation__.Notify(183616)
		return fmt.Errorf("empty GCE project")
	} else {
		__antithesis_instrumentation__.Notify(183617)
	}
	__antithesis_instrumentation__.Notify(183614)
	prj := strings.Split(projects, ",")
	if !v.AcceptMultipleProjects && func() bool {
		__antithesis_instrumentation__.Notify(183618)
		return len(prj) > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(183619)
		return fmt.Errorf("multiple GCE projects not supported for command")
	} else {
		__antithesis_instrumentation__.Notify(183620)
	}
	__antithesis_instrumentation__.Notify(183615)
	providerInstance.Projects = prj
	return nil
}

func (v ProjectsVal) Type() string {
	__antithesis_instrumentation__.Notify(183621)
	if v.AcceptMultipleProjects {
		__antithesis_instrumentation__.Notify(183623)
		return "comma-separated list of GCE projects"
	} else {
		__antithesis_instrumentation__.Notify(183624)
	}
	__antithesis_instrumentation__.Notify(183622)
	return "GCE project name"
}

func (v ProjectsVal) String() string {
	__antithesis_instrumentation__.Notify(183625)
	return strings.Join(providerInstance.Projects, ",")
}

func (p *Provider) GetProject() string {
	__antithesis_instrumentation__.Notify(183626)
	if len(p.Projects) > 1 {
		__antithesis_instrumentation__.Notify(183628)
		panic(fmt.Sprintf(
			"multiple projects not supported (%d specified)", len(p.Projects)))
	} else {
		__antithesis_instrumentation__.Notify(183629)
	}
	__antithesis_instrumentation__.Notify(183627)
	return p.Projects[0]
}

func (p *Provider) GetProjects() []string {
	__antithesis_instrumentation__.Notify(183630)
	return p.Projects
}

func (o *ProviderOpts) ConfigureCreateFlags(flags *pflag.FlagSet) {
	__antithesis_instrumentation__.Notify(183631)
	flags.StringVar(&o.MachineType, "machine-type", "n1-standard-4", "DEPRECATED")
	_ = flags.MarkDeprecated("machine-type", "use "+ProviderName+"-machine-type instead")
	flags.StringSliceVar(&o.Zones, "zones", nil, "DEPRECATED")
	_ = flags.MarkDeprecated("zones", "use "+ProviderName+"-zones instead")

	flags.StringVar(&providerInstance.ServiceAccount, ProviderName+"-service-account",
		providerInstance.ServiceAccount, "Service account to use")

	flags.StringVar(&o.MachineType, ProviderName+"-machine-type", "n1-standard-4",
		"Machine type (see https://cloud.google.com/compute/docs/machine-types)")
	flags.StringVar(&o.MinCPUPlatform, ProviderName+"-min-cpu-platform", "",
		"Minimum CPU platform (see https://cloud.google.com/compute/docs/instances/specify-min-cpu-platform)")
	flags.StringVar(&o.Image, ProviderName+"-image", "ubuntu-2004-focal-v20210603",
		"Image to use to create the vm, "+
			"use `gcloud compute images list --filter=\"family=ubuntu-2004-lts\"` to list available images")

	flags.IntVar(&o.SSDCount, ProviderName+"-local-ssd-count", 1,
		"Number of local SSDs to create, only used if local-ssd=true")
	flags.StringVar(&o.PDVolumeType, ProviderName+"-pd-volume-type", "pd-ssd",
		"Type of the persistent disk volume, only used if local-ssd=false")
	flags.IntVar(&o.PDVolumeSize, ProviderName+"-pd-volume-size", 500,
		"Size in GB of persistent disk volume, only used if local-ssd=false")
	flags.BoolVar(&o.UseMultipleDisks, ProviderName+"-enable-multiple-stores",
		false, "Enable the use of multiple stores by creating one store directory per disk. "+
			"Default is to raid0 stripe all disks.")

	flags.StringSliceVar(&o.Zones, ProviderName+"-zones", nil,
		fmt.Sprintf("Zones for cluster. If zones are formatted as AZ:N where N is an integer, the zone\n"+
			"will be repeated N times. If > 1 zone specified, nodes will be geo-distributed\n"+
			"regardless of geo (default [%s])",
			strings.Join(defaultZones, ",")))
	flags.BoolVar(&o.preemptible, ProviderName+"-preemptible", false, "use preemptible GCE instances")
}

func (o *ProviderOpts) ConfigureClusterFlags(flags *pflag.FlagSet, opt vm.MultipleProjectsOption) {
	__antithesis_instrumentation__.Notify(183632)
	var usage string
	if opt == vm.SingleProject {
		__antithesis_instrumentation__.Notify(183634)
		usage = "GCE project to manage"
	} else {
		__antithesis_instrumentation__.Notify(183635)
		usage = "List of GCE projects to manage"
	}
	__antithesis_instrumentation__.Notify(183633)

	flags.Var(
		ProjectsVal{
			AcceptMultipleProjects: opt == vm.AcceptMultipleProjects,
		},
		ProviderName+"-project",
		usage)

	flags.BoolVar(&o.useSharedUser,
		ProviderName+"-use-shared-user", true,
		fmt.Sprintf("use the shared user %q for ssh rather than your user %q",
			config.SharedUser, config.OSUser.Username))
}

func (p *Provider) CleanSSH() error {
	__antithesis_instrumentation__.Notify(183636)
	for _, prj := range p.GetProjects() {
		__antithesis_instrumentation__.Notify(183638)
		args := []string{"compute", "config-ssh", "--project", prj, "--quiet", "--remove"}
		cmd := exec.Command("gcloud", args...)

		output, err := cmd.CombinedOutput()
		if err != nil {
			__antithesis_instrumentation__.Notify(183639)
			return errors.Wrapf(err, "Command: gcloud %s\nOutput: %s", args, output)
		} else {
			__antithesis_instrumentation__.Notify(183640)
		}
	}
	__antithesis_instrumentation__.Notify(183637)
	return nil
}

func (p *Provider) ConfigSSH(zones []string) error {
	__antithesis_instrumentation__.Notify(183641)

	for _, prj := range p.GetProjects() {
		__antithesis_instrumentation__.Notify(183643)
		args := []string{"compute", "config-ssh", "--project", prj, "--quiet"}
		cmd := exec.Command("gcloud", args...)

		output, err := cmd.CombinedOutput()
		if err != nil {
			__antithesis_instrumentation__.Notify(183644)
			return errors.Wrapf(err, "Command: gcloud %s\nOutput: %s", args, output)
		} else {
			__antithesis_instrumentation__.Notify(183645)
		}
	}
	__antithesis_instrumentation__.Notify(183642)
	return nil
}

func (p *Provider) Create(
	l *logger.Logger, names []string, opts vm.CreateOpts, vmProviderOpts vm.ProviderOpts,
) error {
	__antithesis_instrumentation__.Notify(183646)
	providerOpts := vmProviderOpts.(*ProviderOpts)
	project := p.GetProject()
	var gcJob bool
	for _, prj := range projectsWithGC {
		__antithesis_instrumentation__.Notify(183662)
		if prj == p.GetProject() {
			__antithesis_instrumentation__.Notify(183663)
			gcJob = true
			break
		} else {
			__antithesis_instrumentation__.Notify(183664)
		}
	}
	__antithesis_instrumentation__.Notify(183647)
	if !gcJob {
		__antithesis_instrumentation__.Notify(183665)
		l.Printf("WARNING: --lifetime functionality requires "+
			"`roachprod gc --gce-project=%s` cronjob", project)
	} else {
		__antithesis_instrumentation__.Notify(183666)
	}
	__antithesis_instrumentation__.Notify(183648)

	zones, err := vm.ExpandZonesFlag(providerOpts.Zones)
	if err != nil {
		__antithesis_instrumentation__.Notify(183667)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183668)
	}
	__antithesis_instrumentation__.Notify(183649)
	if len(zones) == 0 {
		__antithesis_instrumentation__.Notify(183669)
		if opts.GeoDistributed {
			__antithesis_instrumentation__.Notify(183670)
			zones = defaultZones
		} else {
			__antithesis_instrumentation__.Notify(183671)
			zones = []string{defaultZones[0]}
		}
	} else {
		__antithesis_instrumentation__.Notify(183672)
	}
	__antithesis_instrumentation__.Notify(183650)

	args := []string{
		"compute", "instances", "create",
		"--subnet", "default",
		"--maintenance-policy", "MIGRATE",
		"--scopes", "default,storage-rw",
		"--image", providerOpts.Image,
		"--image-project", "ubuntu-os-cloud",
		"--boot-disk-type", "pd-ssd",
	}

	if project == defaultProject && func() bool {
		__antithesis_instrumentation__.Notify(183673)
		return p.ServiceAccount == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(183674)
		p.ServiceAccount = "21965078311-compute@developer.gserviceaccount.com"

	} else {
		__antithesis_instrumentation__.Notify(183675)
	}
	__antithesis_instrumentation__.Notify(183651)
	if p.ServiceAccount != "" {
		__antithesis_instrumentation__.Notify(183676)
		args = append(args, "--service-account", p.ServiceAccount)
	} else {
		__antithesis_instrumentation__.Notify(183677)
	}
	__antithesis_instrumentation__.Notify(183652)

	if providerOpts.preemptible {
		__antithesis_instrumentation__.Notify(183678)

		if opts.Lifetime > time.Hour*24 {
			__antithesis_instrumentation__.Notify(183680)
			return errors.New("lifetime cannot be longer than 24 hours for preemptible instances")
		} else {
			__antithesis_instrumentation__.Notify(183681)
		}
		__antithesis_instrumentation__.Notify(183679)
		args = append(args, "--preemptible")

		args = append(args, "--maintenance-policy=terminate")
		args = append(args, "--no-restart-on-failure")
	} else {
		__antithesis_instrumentation__.Notify(183682)
	}
	__antithesis_instrumentation__.Notify(183653)

	extraMountOpts := ""

	if opts.SSDOpts.UseLocalSSD {
		__antithesis_instrumentation__.Notify(183683)

		n2MachineTypes := regexp.MustCompile("^[cn]2-.+-16")
		if n2MachineTypes.MatchString(providerOpts.MachineType) && func() bool {
			__antithesis_instrumentation__.Notify(183686)
			return providerOpts.SSDCount == 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(183687)
			fmt.Fprint(os.Stderr, "WARNING: SSD count must be at least 2 for n2 and c2 machine types with 16vCPU. Setting --gce-local-ssd-count to 2.\n")
			providerOpts.SSDCount = 2
		} else {
			__antithesis_instrumentation__.Notify(183688)
		}
		__antithesis_instrumentation__.Notify(183684)
		for i := 0; i < providerOpts.SSDCount; i++ {
			__antithesis_instrumentation__.Notify(183689)
			args = append(args, "--local-ssd", "interface=NVME")
		}
		__antithesis_instrumentation__.Notify(183685)
		if opts.SSDOpts.NoExt4Barrier {
			__antithesis_instrumentation__.Notify(183690)
			extraMountOpts = "nobarrier"
		} else {
			__antithesis_instrumentation__.Notify(183691)
		}
	} else {
		__antithesis_instrumentation__.Notify(183692)
		pdProps := []string{
			fmt.Sprintf("type=%s", providerOpts.PDVolumeType),
			fmt.Sprintf("size=%dGB", providerOpts.PDVolumeSize),
			"auto-delete=yes",
		}
		args = append(args, "--create-disk", strings.Join(pdProps, ","))

		extraMountOpts = "discard"
	}
	__antithesis_instrumentation__.Notify(183654)

	filename, err := writeStartupScript(extraMountOpts, opts.SSDOpts.FileSystem, providerOpts.UseMultipleDisks)
	if err != nil {
		__antithesis_instrumentation__.Notify(183693)
		return errors.Wrapf(err, "could not write GCE startup script to temp file")
	} else {
		__antithesis_instrumentation__.Notify(183694)
	}
	__antithesis_instrumentation__.Notify(183655)
	defer func() {
		__antithesis_instrumentation__.Notify(183695)
		_ = os.Remove(filename)
	}()
	__antithesis_instrumentation__.Notify(183656)

	args = append(args, "--machine-type", providerOpts.MachineType)
	if providerOpts.MinCPUPlatform != "" {
		__antithesis_instrumentation__.Notify(183696)
		args = append(args, "--min-cpu-platform", providerOpts.MinCPUPlatform)
	} else {
		__antithesis_instrumentation__.Notify(183697)
	}
	__antithesis_instrumentation__.Notify(183657)

	m := vm.GetDefaultLabelMap(opts)

	time := timeutil.Now().Format(time.RFC3339)
	time = strings.ToLower(strings.ReplaceAll(time, ":", "_"))
	m[vm.TagCreated] = time

	var sb strings.Builder
	for key, value := range opts.CustomLabels {
		__antithesis_instrumentation__.Notify(183698)
		_, ok := m[key]
		if ok {
			__antithesis_instrumentation__.Notify(183700)
			return fmt.Errorf("duplicate label name defined: %s", key)
		} else {
			__antithesis_instrumentation__.Notify(183701)
		}
		__antithesis_instrumentation__.Notify(183699)
		fmt.Fprintf(&sb, "%s=%s,", key, value)
	}
	__antithesis_instrumentation__.Notify(183658)
	for key, value := range m {
		__antithesis_instrumentation__.Notify(183702)
		fmt.Fprintf(&sb, "%s=%s,", key, value)
	}
	__antithesis_instrumentation__.Notify(183659)
	s := sb.String()
	args = append(args, "--labels", s[:len(s)-1])

	args = append(args, "--metadata-from-file", fmt.Sprintf("startup-script=%s", filename))
	args = append(args, "--project", project)
	args = append(args, fmt.Sprintf("--boot-disk-size=%dGB", opts.OsVolumeSize))
	var g errgroup.Group

	nodeZones := vm.ZonePlacement(len(zones), len(names))
	zoneHostNames := make([][]string, len(zones))
	for i, name := range names {
		__antithesis_instrumentation__.Notify(183703)
		zone := nodeZones[i]
		zoneHostNames[zone] = append(zoneHostNames[zone], name)
	}
	__antithesis_instrumentation__.Notify(183660)
	for i, zoneHosts := range zoneHostNames {
		__antithesis_instrumentation__.Notify(183704)
		argsWithZone := append(args[:len(args):len(args)], "--zone", zones[i])
		argsWithZone = append(argsWithZone, zoneHosts...)
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(183705)
			cmd := exec.Command("gcloud", argsWithZone...)

			output, err := cmd.CombinedOutput()
			if err != nil {
				__antithesis_instrumentation__.Notify(183707)
				return errors.Wrapf(err, "Command: gcloud %s\nOutput: %s", args, output)
			} else {
				__antithesis_instrumentation__.Notify(183708)
			}
			__antithesis_instrumentation__.Notify(183706)
			return nil
		})

	}
	__antithesis_instrumentation__.Notify(183661)
	return g.Wait()
}

func (p *Provider) Delete(vms vm.List) error {
	__antithesis_instrumentation__.Notify(183709)

	projectZoneMap := make(map[string]map[string][]string)
	for _, v := range vms {
		__antithesis_instrumentation__.Notify(183712)
		if v.Provider != ProviderName {
			__antithesis_instrumentation__.Notify(183715)
			return errors.Errorf("%s received VM instance from %s", ProviderName, v.Provider)
		} else {
			__antithesis_instrumentation__.Notify(183716)
		}
		__antithesis_instrumentation__.Notify(183713)
		if projectZoneMap[v.Project] == nil {
			__antithesis_instrumentation__.Notify(183717)
			projectZoneMap[v.Project] = make(map[string][]string)
		} else {
			__antithesis_instrumentation__.Notify(183718)
		}
		__antithesis_instrumentation__.Notify(183714)

		projectZoneMap[v.Project][v.Zone] = append(projectZoneMap[v.Project][v.Zone], v.Name)
	}
	__antithesis_instrumentation__.Notify(183710)

	var g errgroup.Group
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	for project, zoneMap := range projectZoneMap {
		__antithesis_instrumentation__.Notify(183719)
		for zone, names := range zoneMap {
			__antithesis_instrumentation__.Notify(183720)
			args := []string{
				"compute", "instances", "delete",
				"--delete-disks", "all",
			}

			args = append(args, "--project", project)
			args = append(args, "--zone", zone)
			args = append(args, names...)

			g.Go(func() error {
				__antithesis_instrumentation__.Notify(183721)
				cmd := exec.CommandContext(ctx, "gcloud", args...)

				output, err := cmd.CombinedOutput()
				if err != nil {
					__antithesis_instrumentation__.Notify(183723)
					return errors.Wrapf(err, "Command: gcloud %s\nOutput: %s", args, output)
				} else {
					__antithesis_instrumentation__.Notify(183724)
				}
				__antithesis_instrumentation__.Notify(183722)
				return nil
			})
		}
	}
	__antithesis_instrumentation__.Notify(183711)

	return g.Wait()
}

func (p *Provider) Reset(vms vm.List) error {
	__antithesis_instrumentation__.Notify(183725)

	projectZoneMap := make(map[string]map[string][]string)
	for _, v := range vms {
		__antithesis_instrumentation__.Notify(183728)
		if v.Provider != ProviderName {
			__antithesis_instrumentation__.Notify(183731)
			return errors.Errorf("%s received VM instance from %s", ProviderName, v.Provider)
		} else {
			__antithesis_instrumentation__.Notify(183732)
		}
		__antithesis_instrumentation__.Notify(183729)
		if projectZoneMap[v.Project] == nil {
			__antithesis_instrumentation__.Notify(183733)
			projectZoneMap[v.Project] = make(map[string][]string)
		} else {
			__antithesis_instrumentation__.Notify(183734)
		}
		__antithesis_instrumentation__.Notify(183730)

		projectZoneMap[v.Project][v.Zone] = append(projectZoneMap[v.Project][v.Zone], v.Name)
	}
	__antithesis_instrumentation__.Notify(183726)

	var g errgroup.Group
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	for project, zoneMap := range projectZoneMap {
		__antithesis_instrumentation__.Notify(183735)
		for zone, names := range zoneMap {
			__antithesis_instrumentation__.Notify(183736)
			args := []string{
				"compute", "instances", "reset",
			}

			args = append(args, "--project", project)
			args = append(args, "--zone", zone)
			args = append(args, names...)

			g.Go(func() error {
				__antithesis_instrumentation__.Notify(183737)
				cmd := exec.CommandContext(ctx, "gcloud", args...)

				output, err := cmd.CombinedOutput()
				if err != nil {
					__antithesis_instrumentation__.Notify(183739)
					return errors.Wrapf(err, "Command: gcloud %s\nOutput: %s", args, output)
				} else {
					__antithesis_instrumentation__.Notify(183740)
				}
				__antithesis_instrumentation__.Notify(183738)
				return nil
			})
		}
	}
	__antithesis_instrumentation__.Notify(183727)

	return g.Wait()
}

func (p *Provider) Extend(vms vm.List, lifetime time.Duration) error {
	__antithesis_instrumentation__.Notify(183741)

	for _, v := range vms {
		__antithesis_instrumentation__.Notify(183743)
		args := []string{"compute", "instances", "add-labels"}

		args = append(args, "--project", v.Project)
		args = append(args, "--zone", v.Zone)
		args = append(args, "--labels", fmt.Sprintf("lifetime=%s", lifetime))
		args = append(args, v.Name)

		cmd := exec.Command("gcloud", args...)

		output, err := cmd.CombinedOutput()
		if err != nil {
			__antithesis_instrumentation__.Notify(183744)
			return errors.Wrapf(err, "Command: gcloud %s\nOutput: %s", args, output)
		} else {
			__antithesis_instrumentation__.Notify(183745)
		}
	}
	__antithesis_instrumentation__.Notify(183742)
	return nil
}

func (p *Provider) FindActiveAccount() (string, error) {
	__antithesis_instrumentation__.Notify(183746)
	args := []string{"auth", "list", "--format", "json", "--filter", "status~ACTIVE"}

	accounts := make([]jsonAuth, 0)
	if err := runJSONCommand(args, &accounts); err != nil {
		__antithesis_instrumentation__.Notify(183750)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183751)
	}
	__antithesis_instrumentation__.Notify(183747)

	if len(accounts) != 1 {
		__antithesis_instrumentation__.Notify(183752)
		return "", fmt.Errorf("no active accounts found, please configure gcloud")
	} else {
		__antithesis_instrumentation__.Notify(183753)
	}
	__antithesis_instrumentation__.Notify(183748)

	if !strings.HasSuffix(accounts[0].Account, config.EmailDomain) {
		__antithesis_instrumentation__.Notify(183754)
		return "", fmt.Errorf("active account %q does not belong to domain %s",
			accounts[0].Account, config.EmailDomain)
	} else {
		__antithesis_instrumentation__.Notify(183755)
	}
	__antithesis_instrumentation__.Notify(183749)
	_ = accounts[0].Status

	username := strings.Split(accounts[0].Account, "@")[0]
	return username, nil
}

func (p *Provider) List(l *logger.Logger) (vm.List, error) {
	__antithesis_instrumentation__.Notify(183756)
	var vms vm.List
	for _, prj := range p.GetProjects() {
		__antithesis_instrumentation__.Notify(183758)
		args := []string{"compute", "instances", "list", "--project", prj, "--format", "json"}

		jsonVMS := make([]jsonVM, 0)
		if err := runJSONCommand(args, &jsonVMS); err != nil {
			__antithesis_instrumentation__.Notify(183760)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(183761)
		}
		__antithesis_instrumentation__.Notify(183759)

		for _, jsonVM := range jsonVMS {
			__antithesis_instrumentation__.Notify(183762)
			defaultOpts := p.CreateProviderOpts().(*ProviderOpts)
			vms = append(vms, *jsonVM.toVM(prj, defaultOpts))
		}
	}
	__antithesis_instrumentation__.Notify(183757)

	return vms, nil
}

func (p *Provider) Name() string {
	__antithesis_instrumentation__.Notify(183763)
	return ProviderName
}

func (p *Provider) Active() bool {
	__antithesis_instrumentation__.Notify(183764)
	return true
}

func (p *Provider) ProjectActive(project string) bool {
	__antithesis_instrumentation__.Notify(183765)
	for _, p := range p.GetProjects() {
		__antithesis_instrumentation__.Notify(183767)
		if p == project {
			__antithesis_instrumentation__.Notify(183768)
			return true
		} else {
			__antithesis_instrumentation__.Notify(183769)
		}
	}
	__antithesis_instrumentation__.Notify(183766)
	return false
}
