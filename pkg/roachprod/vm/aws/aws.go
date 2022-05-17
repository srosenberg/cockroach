package aws

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/flagstub"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

const ProviderName = "aws"

var providerInstance = &Provider{}

func Init() error {
	__antithesis_instrumentation__.Notify(182552)

	const unsupportedAwsCliVersionPrefix = "aws-cli/1."
	const unimplemented = "please install the AWS CLI utilities version 2+ " +
		"(https://docs.aws.amazon.com/cli/latest/userguide/installing.html)"
	const noCredentials = "missing AWS credentials, expected ~/.aws/credentials file or AWS_ACCESS_KEY_ID env var"

	configVal := awsConfigValue{awsConfig: *defaultConfig}
	providerInstance.Config = &configVal.awsConfig
	providerInstance.IAMProfile = "roachprod-testing"

	haveRequiredVersion := func() bool {
		__antithesis_instrumentation__.Notify(182557)
		cmd := exec.Command("aws", "--version")
		output, err := cmd.Output()
		if err != nil {
			__antithesis_instrumentation__.Notify(182560)
			return false
		} else {
			__antithesis_instrumentation__.Notify(182561)
		}
		__antithesis_instrumentation__.Notify(182558)
		if strings.HasPrefix(string(output), unsupportedAwsCliVersionPrefix) {
			__antithesis_instrumentation__.Notify(182562)
			return false
		} else {
			__antithesis_instrumentation__.Notify(182563)
		}
		__antithesis_instrumentation__.Notify(182559)
		return true
	}
	__antithesis_instrumentation__.Notify(182553)
	if !haveRequiredVersion() {
		__antithesis_instrumentation__.Notify(182564)
		vm.Providers[ProviderName] = flagstub.New(&Provider{}, unimplemented)
		return errors.New("doesn't have the required version")
	} else {
		__antithesis_instrumentation__.Notify(182565)
	}
	__antithesis_instrumentation__.Notify(182554)

	haveCredentials := func() bool {
		__antithesis_instrumentation__.Notify(182566)
		const credFile = "${HOME}/.aws/credentials"
		if _, err := os.Stat(os.ExpandEnv(credFile)); err == nil {
			__antithesis_instrumentation__.Notify(182569)
			return true
		} else {
			__antithesis_instrumentation__.Notify(182570)
		}
		__antithesis_instrumentation__.Notify(182567)
		if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
			__antithesis_instrumentation__.Notify(182571)
			return true
		} else {
			__antithesis_instrumentation__.Notify(182572)
		}
		__antithesis_instrumentation__.Notify(182568)
		return false
	}
	__antithesis_instrumentation__.Notify(182555)
	if !haveCredentials() {
		__antithesis_instrumentation__.Notify(182573)
		vm.Providers[ProviderName] = flagstub.New(&Provider{}, noCredentials)
		return errors.New("missing/invalid credentials")
	} else {
		__antithesis_instrumentation__.Notify(182574)
	}
	__antithesis_instrumentation__.Notify(182556)
	vm.Providers[ProviderName] = providerInstance
	return nil
}

type ebsDisk struct {
	VolumeType          string `json:"VolumeType"`
	VolumeSize          int    `json:"VolumeSize"`
	IOPs                int    `json:"Iops,omitempty"`
	Throughput          int    `json:"Throughput,omitempty"`
	DeleteOnTermination bool   `json:"DeleteOnTermination"`
}

type ebsVolume struct {
	DeviceName string  `json:"DeviceName"`
	Disk       ebsDisk `json:"Ebs"`
}

const ebsDefaultVolumeSizeGB = 500
const defaultEBSVolumeType = "gp3"

func (d *ebsDisk) Set(s string) error {
	__antithesis_instrumentation__.Notify(182575)
	if err := json.Unmarshal([]byte(s), &d); err != nil {
		__antithesis_instrumentation__.Notify(182579)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182580)
	}
	__antithesis_instrumentation__.Notify(182576)

	d.DeleteOnTermination = true

	if d.VolumeSize == 0 {
		__antithesis_instrumentation__.Notify(182581)
		d.VolumeSize = ebsDefaultVolumeSizeGB
	} else {
		__antithesis_instrumentation__.Notify(182582)
	}
	__antithesis_instrumentation__.Notify(182577)

	switch strings.ToLower(d.VolumeType) {
	case "gp2":
		__antithesis_instrumentation__.Notify(182583)

	case "gp3":
		__antithesis_instrumentation__.Notify(182584)
		if d.IOPs > 16000 {
			__antithesis_instrumentation__.Notify(182589)
			return errors.AssertionFailedf("Iops required for gp3 disk: [3000, 16000]")
		} else {
			__antithesis_instrumentation__.Notify(182590)
		}
		__antithesis_instrumentation__.Notify(182585)
		if d.IOPs == 0 {
			__antithesis_instrumentation__.Notify(182591)

			d.IOPs = 3000
		} else {
			__antithesis_instrumentation__.Notify(182592)
		}
		__antithesis_instrumentation__.Notify(182586)
		if d.Throughput == 0 {
			__antithesis_instrumentation__.Notify(182593)

			d.Throughput = 125
		} else {
			__antithesis_instrumentation__.Notify(182594)
		}
	case "io1", "io2":
		__antithesis_instrumentation__.Notify(182587)
		if d.IOPs == 0 {
			__antithesis_instrumentation__.Notify(182595)
			return errors.AssertionFailedf("Iops required for %s disk", d.VolumeType)
		} else {
			__antithesis_instrumentation__.Notify(182596)
		}
	default:
		__antithesis_instrumentation__.Notify(182588)
		return errors.Errorf("Unknown EBS volume type %s", d.VolumeType)
	}
	__antithesis_instrumentation__.Notify(182578)
	return nil
}

func (d *ebsDisk) Type() string {
	__antithesis_instrumentation__.Notify(182597)
	return "JSON"
}

func (d *ebsDisk) String() string {
	__antithesis_instrumentation__.Notify(182598)
	return "EBSDisk"
}

type ebsVolumeList []*ebsVolume

func (vl *ebsVolumeList) newVolume() *ebsVolume {
	__antithesis_instrumentation__.Notify(182599)
	return &ebsVolume{
		DeviceName: fmt.Sprintf("/dev/sd%c", 'd'+len(*vl)),
	}
}

func (vl *ebsVolumeList) Set(s string) error {
	__antithesis_instrumentation__.Notify(182600)
	v := vl.newVolume()
	if err := v.Disk.Set(s); err != nil {
		__antithesis_instrumentation__.Notify(182602)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182603)
	}
	__antithesis_instrumentation__.Notify(182601)
	*vl = append(*vl, v)
	return nil
}

func (vl *ebsVolumeList) Type() string {
	__antithesis_instrumentation__.Notify(182604)
	return "JSON"
}

func (vl *ebsVolumeList) String() string {
	__antithesis_instrumentation__.Notify(182605)
	return "EBSVolumeList"
}

func DefaultProviderOpts() *ProviderOpts {
	__antithesis_instrumentation__.Notify(182606)
	defaultEBSVolumeValue := ebsVolume{}
	defaultEBSVolumeValue.Disk.VolumeSize = ebsDefaultVolumeSizeGB
	defaultEBSVolumeValue.Disk.VolumeType = defaultEBSVolumeType
	return &ProviderOpts{
		MachineType:      "m5.xlarge",
		SSDMachineType:   "m5d.xlarge",
		RemoteUserName:   "ubuntu",
		DefaultEBSVolume: defaultEBSVolumeValue,
		CreateRateLimit:  2,
	}
}

func (p *Provider) CreateProviderOpts() vm.ProviderOpts {
	__antithesis_instrumentation__.Notify(182607)
	return DefaultProviderOpts()
}

type ProviderOpts struct {
	MachineType      string
	SSDMachineType   string
	CPUOptions       string
	RemoteUserName   string
	DefaultEBSVolume ebsVolume
	EBSVolumes       ebsVolumeList
	UseMultipleDisks bool

	ImageAMI string

	CreateZones []string

	CreateRateLimit float64
}

type Provider struct {
	Profile string

	Config *awsConfig

	IAMProfile string
}

const (
	defaultSSDMachineType = "m5d.xlarge"
	defaultMachineType    = "m5.xlarge"
)

var defaultConfig = func() (cfg *awsConfig) {
	__antithesis_instrumentation__.Notify(182608)
	cfg = new(awsConfig)
	if err := json.Unmarshal(MustAsset("config.json"), cfg); err != nil {
		__antithesis_instrumentation__.Notify(182610)
		panic(errors.Wrap(err, "failed to embedded configuration"))
	} else {
		__antithesis_instrumentation__.Notify(182611)
	}
	__antithesis_instrumentation__.Notify(182609)
	return cfg
}()

var defaultCreateZones = []string{
	"us-east-2b",
	"us-west-2b",
	"eu-west-2b",
}

func (o *ProviderOpts) ConfigureCreateFlags(flags *pflag.FlagSet) {
	__antithesis_instrumentation__.Notify(182612)

	flags.StringVar(&o.MachineType, ProviderName+"-machine-type", o.MachineType,
		"Machine type (see https://aws.amazon.com/ec2/instance-types/)")

	flags.StringVar(&o.SSDMachineType, ProviderName+"-machine-type-ssd", o.SSDMachineType,
		"Machine type for --local-ssd (see https://aws.amazon.com/ec2/instance-types/)")

	flags.StringVar(&o.CPUOptions, ProviderName+"-cpu-options", o.CPUOptions,
		"Options to specify number of cores and threads per core (see https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-optimize-cpu.html#instance-specify-cpu-options)")

	flags.StringVar(&o.RemoteUserName, ProviderName+"-user",
		o.RemoteUserName, "Name of the remote user to SSH as")

	flags.StringVar(&o.DefaultEBSVolume.Disk.VolumeType, ProviderName+"-ebs-volume-type",
		o.DefaultEBSVolume.Disk.VolumeType, "Type of the EBS volume, only used if local-ssd=false")
	flags.IntVar(&o.DefaultEBSVolume.Disk.VolumeSize, ProviderName+"-ebs-volume-size",
		o.DefaultEBSVolume.Disk.VolumeSize, "Size in GB of EBS volume, only used if local-ssd=false")
	flags.IntVar(&o.DefaultEBSVolume.Disk.IOPs, ProviderName+"-ebs-iops",
		o.DefaultEBSVolume.Disk.IOPs, "Number of IOPs to provision for supported disk types (io1, io2, gp3)")
	flags.IntVar(&o.DefaultEBSVolume.Disk.Throughput, ProviderName+"-ebs-throughput",
		o.DefaultEBSVolume.Disk.Throughput, "Additional throughput to provision, in MiB/s")

	flags.VarP(&o.EBSVolumes, ProviderName+"-ebs-volume", "",
		`Additional EBS disk to attached, repeated for extra disks; specified as JSON: {"VolumeType":"io2","VolumeSize":213,"Iops":321}`)

	flags.StringSliceVar(&o.CreateZones, ProviderName+"-zones", o.CreateZones,
		fmt.Sprintf("aws availability zones to use for cluster creation. If zones are formatted\n"+
			"as AZ:N where N is an integer, the zone will be repeated N times. If > 1\n"+
			"zone specified, the cluster will be spread out evenly by zone regardless\n"+
			"of geo (default [%s])", strings.Join(defaultCreateZones, ",")))
	flags.StringVar(&o.ImageAMI, ProviderName+"-image-ami",
		o.ImageAMI, "Override image AMI to use.  See https://awscli.amazonaws.com/v2/documentation/api/latest/reference/ec2/describe-images.html")
	flags.BoolVar(&o.UseMultipleDisks, ProviderName+"-enable-multiple-stores",
		false, "Enable the use of multiple stores by creating one store directory per disk. "+
			"Default is to raid0 stripe all disks. "+
			"See repeating --"+ProviderName+"-ebs-volume for adding extra volumes.")
	flags.Float64Var(&o.CreateRateLimit, ProviderName+"-create-rate-limit", o.CreateRateLimit, "aws"+
		" rate limit (per second) for instance creation. This is used to avoid hitting the request"+
		" limits from aws, which can vary based on the region, and the size of the cluster being"+
		" created. Try lowering this limit when hitting 'Request limit exceeded' errors.")
	flags.StringVar(&providerInstance.IAMProfile, ProviderName+"-	iam-profile", providerInstance.IAMProfile,
		"the IAM instance profile to associate with created VMs if non-empty")

}

func (o *ProviderOpts) ConfigureClusterFlags(flags *pflag.FlagSet, _ vm.MultipleProjectsOption) {
	__antithesis_instrumentation__.Notify(182613)
	flags.StringVar(&providerInstance.Profile, ProviderName+"-profile", providerInstance.Profile,
		"Profile to manage cluster in")
	configFlagVal := awsConfigValue{awsConfig: *defaultConfig}
	providerInstance.Config = &configFlagVal.awsConfig
	flags.Var(&configFlagVal, ProviderName+"-config",
		"Path to json for aws configuration, defaults to predefined configuration")
}

func (p *Provider) CleanSSH() error {
	__antithesis_instrumentation__.Notify(182614)
	return nil
}

func (p *Provider) ConfigSSH(zones []string) error {
	__antithesis_instrumentation__.Notify(182615)
	keyName, err := p.sshKeyName()
	if err != nil {
		__antithesis_instrumentation__.Notify(182619)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182620)
	}
	__antithesis_instrumentation__.Notify(182616)

	regions, err := p.allRegions(zones)
	if err != nil {
		__antithesis_instrumentation__.Notify(182621)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182622)
	}
	__antithesis_instrumentation__.Notify(182617)

	var g errgroup.Group
	for _, r := range regions {
		__antithesis_instrumentation__.Notify(182623)

		region := r
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(182624)
			exists, err := p.sshKeyExists(keyName, region)
			if err != nil {
				__antithesis_instrumentation__.Notify(182627)
				return err
			} else {
				__antithesis_instrumentation__.Notify(182628)
			}
			__antithesis_instrumentation__.Notify(182625)
			if !exists {
				__antithesis_instrumentation__.Notify(182629)
				err = p.sshKeyImport(keyName, region)
				if err != nil {
					__antithesis_instrumentation__.Notify(182631)
					return err
				} else {
					__antithesis_instrumentation__.Notify(182632)
				}
				__antithesis_instrumentation__.Notify(182630)
				log.Infof(context.Background(), "imported %s as %s in region %s",
					sshPublicKeyFile, keyName, region)
			} else {
				__antithesis_instrumentation__.Notify(182633)
			}
			__antithesis_instrumentation__.Notify(182626)
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(182618)

	return g.Wait()
}

func (p *Provider) Create(
	l *logger.Logger, names []string, opts vm.CreateOpts, vmProviderOpts vm.ProviderOpts,
) error {
	__antithesis_instrumentation__.Notify(182634)
	providerOpts := vmProviderOpts.(*ProviderOpts)
	expandedZones, err := vm.ExpandZonesFlag(providerOpts.CreateZones)
	if err != nil {
		__antithesis_instrumentation__.Notify(182643)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182644)
	}
	__antithesis_instrumentation__.Notify(182635)

	useDefaultZones := len(expandedZones) == 0
	if useDefaultZones {
		__antithesis_instrumentation__.Notify(182645)
		expandedZones = defaultCreateZones
	} else {
		__antithesis_instrumentation__.Notify(182646)
	}
	__antithesis_instrumentation__.Notify(182636)

	if err := p.ConfigSSH(expandedZones); err != nil {
		__antithesis_instrumentation__.Notify(182647)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182648)
	}
	__antithesis_instrumentation__.Notify(182637)

	regions, err := p.allRegions(expandedZones)
	if err != nil {
		__antithesis_instrumentation__.Notify(182649)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182650)
	}
	__antithesis_instrumentation__.Notify(182638)
	if len(regions) < 1 {
		__antithesis_instrumentation__.Notify(182651)
		return errors.Errorf("Please specify a valid region.")
	} else {
		__antithesis_instrumentation__.Notify(182652)
	}
	__antithesis_instrumentation__.Notify(182639)

	var zones []string
	if !opts.GeoDistributed && func() bool {
		__antithesis_instrumentation__.Notify(182653)
		return (useDefaultZones || func() bool {
			__antithesis_instrumentation__.Notify(182654)
			return len(expandedZones) == 1 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(182655)

		regionZones, err := p.regionZones(regions[0], expandedZones)
		if err != nil {
			__antithesis_instrumentation__.Notify(182657)
			return err
		} else {
			__antithesis_instrumentation__.Notify(182658)
		}
		__antithesis_instrumentation__.Notify(182656)

		zone := regionZones[rand.Intn(len(regionZones))]
		for range names {
			__antithesis_instrumentation__.Notify(182659)
			zones = append(zones, zone)
		}
	} else {
		__antithesis_instrumentation__.Notify(182660)

		nodeZones := vm.ZonePlacement(len(expandedZones), len(names))
		zones = make([]string, len(nodeZones))
		for i, z := range nodeZones {
			__antithesis_instrumentation__.Notify(182661)
			zones[i] = expandedZones[z]
		}
	}
	__antithesis_instrumentation__.Notify(182640)

	var g errgroup.Group
	limiter := rate.NewLimiter(rate.Limit(providerOpts.CreateRateLimit), 2)
	for i := range names {
		__antithesis_instrumentation__.Notify(182662)
		capName := names[i]
		placement := zones[i]
		res := limiter.Reserve()
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(182663)
			time.Sleep(res.Delay())
			return p.runInstance(capName, placement, opts, providerOpts)
		})
	}
	__antithesis_instrumentation__.Notify(182641)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(182664)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182665)
	}
	__antithesis_instrumentation__.Notify(182642)
	return p.waitForIPs(l, names, regions, providerOpts)
}

func (p *Provider) waitForIPs(
	l *logger.Logger, names []string, regions []string, opts *ProviderOpts,
) error {
	__antithesis_instrumentation__.Notify(182666)
	waitForIPRetry := retry.Start(retry.Options{
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     500 * time.Millisecond,
		MaxRetries:     120,
	})
	makeNameSet := func() map[string]struct{} {
		__antithesis_instrumentation__.Notify(182669)
		m := make(map[string]struct{}, len(names))
		for _, n := range names {
			__antithesis_instrumentation__.Notify(182671)
			m[n] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(182670)
		return m
	}
	__antithesis_instrumentation__.Notify(182667)
	for waitForIPRetry.Next() {
		__antithesis_instrumentation__.Notify(182672)
		vms, err := p.listRegions(l, regions, *opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(182675)
			return err
		} else {
			__antithesis_instrumentation__.Notify(182676)
		}
		__antithesis_instrumentation__.Notify(182673)
		nameSet := makeNameSet()
		for _, vm := range vms {
			__antithesis_instrumentation__.Notify(182677)
			if vm.PublicIP != "" && func() bool {
				__antithesis_instrumentation__.Notify(182678)
				return vm.PrivateIP != "" == true
			}() == true {
				__antithesis_instrumentation__.Notify(182679)
				delete(nameSet, vm.Name)
			} else {
				__antithesis_instrumentation__.Notify(182680)
			}
		}
		__antithesis_instrumentation__.Notify(182674)
		if len(nameSet) == 0 {
			__antithesis_instrumentation__.Notify(182681)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(182682)
		}
	}
	__antithesis_instrumentation__.Notify(182668)
	return fmt.Errorf("failed to retrieve IPs for all vms")
}

func (p *Provider) Delete(vms vm.List) error {
	__antithesis_instrumentation__.Notify(182683)
	byRegion, err := regionMap(vms)
	if err != nil {
		__antithesis_instrumentation__.Notify(182686)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182687)
	}
	__antithesis_instrumentation__.Notify(182684)
	g := errgroup.Group{}
	for region, list := range byRegion {
		__antithesis_instrumentation__.Notify(182688)
		args := []string{
			"ec2", "terminate-instances",
			"--region", region,
			"--instance-ids",
		}
		args = append(args, list.ProviderIDs()...)
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(182689)
			var data struct {
				TerminatingInstances []struct {
					InstanceID string `json:"InstanceId"`
				}
			}
			_ = data.TerminatingInstances
			if len(data.TerminatingInstances) > 0 {
				__antithesis_instrumentation__.Notify(182691)
				_ = data.TerminatingInstances[0].InstanceID
			} else {
				__antithesis_instrumentation__.Notify(182692)
			}
			__antithesis_instrumentation__.Notify(182690)
			return p.runJSONCommand(args, &data)
		})
	}
	__antithesis_instrumentation__.Notify(182685)
	return g.Wait()
}

func (p *Provider) Reset(vms vm.List) error {
	__antithesis_instrumentation__.Notify(182693)
	return nil
}

func (p *Provider) Extend(vms vm.List, lifetime time.Duration) error {
	__antithesis_instrumentation__.Notify(182694)
	byRegion, err := regionMap(vms)
	if err != nil {
		__antithesis_instrumentation__.Notify(182697)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182698)
	}
	__antithesis_instrumentation__.Notify(182695)
	g := errgroup.Group{}
	for region, list := range byRegion {
		__antithesis_instrumentation__.Notify(182699)

		args := []string{
			"ec2", "create-tags",
			"--region", region,
			"--tags", "Key=Lifetime,Value=" + lifetime.String(),
			"--resources",
		}
		args = append(args, list.ProviderIDs()...)

		g.Go(func() error {
			__antithesis_instrumentation__.Notify(182700)
			_, err := p.runCommand(args)
			return err
		})
	}
	__antithesis_instrumentation__.Notify(182696)
	return g.Wait()
}

var cachedActiveAccount string

func (p *Provider) FindActiveAccount() (string, error) {
	__antithesis_instrumentation__.Notify(182701)
	if len(cachedActiveAccount) > 0 {
		__antithesis_instrumentation__.Notify(182704)
		return cachedActiveAccount, nil
	} else {
		__antithesis_instrumentation__.Notify(182705)
	}
	__antithesis_instrumentation__.Notify(182702)
	var account string
	var err error
	if p.Profile == "" {
		__antithesis_instrumentation__.Notify(182706)
		account, err = p.iamGetUser()
		if err != nil {
			__antithesis_instrumentation__.Notify(182707)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(182708)
		}
	} else {
		__antithesis_instrumentation__.Notify(182709)
		account, err = p.stsGetCallerIdentity()
		if err != nil {
			__antithesis_instrumentation__.Notify(182710)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(182711)
		}
	}
	__antithesis_instrumentation__.Notify(182703)
	cachedActiveAccount = account
	return cachedActiveAccount, nil
}

func (p *Provider) iamGetUser() (string, error) {
	__antithesis_instrumentation__.Notify(182712)
	var userInfo struct {
		User struct {
			UserName string
		}
	}
	args := []string{"iam", "get-user"}
	err := p.runJSONCommand(args, &userInfo)
	if err != nil {
		__antithesis_instrumentation__.Notify(182715)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(182716)
	}
	__antithesis_instrumentation__.Notify(182713)
	if userInfo.User.UserName == "" {
		__antithesis_instrumentation__.Notify(182717)
		return "", errors.Errorf("username not configured. run 'aws iam get-user'")
	} else {
		__antithesis_instrumentation__.Notify(182718)
	}
	__antithesis_instrumentation__.Notify(182714)
	return userInfo.User.UserName, nil
}

func (p *Provider) stsGetCallerIdentity() (string, error) {
	__antithesis_instrumentation__.Notify(182719)
	var userInfo struct {
		Arn string
	}
	args := []string{"sts", "get-caller-identity"}
	err := p.runJSONCommand(args, &userInfo)
	if err != nil {
		__antithesis_instrumentation__.Notify(182722)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(182723)
	}
	__antithesis_instrumentation__.Notify(182720)
	s := strings.Split(userInfo.Arn, "/")
	if len(s) < 2 {
		__antithesis_instrumentation__.Notify(182724)
		return "", errors.Errorf("Could not parse caller identity ARN '%s'", userInfo.Arn)
	} else {
		__antithesis_instrumentation__.Notify(182725)
	}
	__antithesis_instrumentation__.Notify(182721)
	return s[1], nil
}

func (p *Provider) List(l *logger.Logger) (vm.List, error) {
	__antithesis_instrumentation__.Notify(182726)
	regions, err := p.allRegions(p.Config.availabilityZoneNames())
	if err != nil {
		__antithesis_instrumentation__.Notify(182728)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182729)
	}
	__antithesis_instrumentation__.Notify(182727)
	defaultOpts := p.CreateProviderOpts().(*ProviderOpts)
	return p.listRegions(l, regions, *defaultOpts)
}

func (p *Provider) listRegions(
	l *logger.Logger, regions []string, opts ProviderOpts,
) (vm.List, error) {
	__antithesis_instrumentation__.Notify(182730)
	var ret vm.List
	var mux syncutil.Mutex
	var g errgroup.Group

	for _, r := range regions {
		__antithesis_instrumentation__.Notify(182733)

		region := r
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(182734)
			vms, err := p.listRegion(region, opts)
			if err != nil {
				__antithesis_instrumentation__.Notify(182736)
				l.Printf("Failed to list AWS VMs in region: %s\n%v\n", region, err)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(182737)
			}
			__antithesis_instrumentation__.Notify(182735)
			mux.Lock()
			ret = append(ret, vms...)
			mux.Unlock()
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(182731)

	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(182738)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182739)
	}
	__antithesis_instrumentation__.Notify(182732)

	return ret, nil
}

func (p *Provider) Name() string {
	__antithesis_instrumentation__.Notify(182740)
	return ProviderName
}

func (p *Provider) allRegions(zones []string) (regions []string, err error) {
	__antithesis_instrumentation__.Notify(182741)
	byName := make(map[string]struct{})
	for _, z := range zones {
		__antithesis_instrumentation__.Notify(182743)
		az := p.Config.getAvailabilityZone(z)
		if az == nil {
			__antithesis_instrumentation__.Notify(182745)
			return nil, fmt.Errorf("unknown availability zone %v, please provide a "+
				"correct value or update your config accordingly", z)
		} else {
			__antithesis_instrumentation__.Notify(182746)
		}
		__antithesis_instrumentation__.Notify(182744)
		if _, have := byName[az.region.Name]; !have {
			__antithesis_instrumentation__.Notify(182747)
			byName[az.region.Name] = struct{}{}
			regions = append(regions, az.region.Name)
		} else {
			__antithesis_instrumentation__.Notify(182748)
		}
	}
	__antithesis_instrumentation__.Notify(182742)
	return regions, nil
}

func (p *Provider) regionZones(region string, allZones []string) (zones []string, _ error) {
	__antithesis_instrumentation__.Notify(182749)
	r := p.Config.getRegion(region)
	if r == nil {
		__antithesis_instrumentation__.Notify(182752)
		return nil, fmt.Errorf("region %s not found", region)
	} else {
		__antithesis_instrumentation__.Notify(182753)
	}
	__antithesis_instrumentation__.Notify(182750)
	for _, z := range allZones {
		__antithesis_instrumentation__.Notify(182754)
		for _, az := range r.AvailabilityZones {
			__antithesis_instrumentation__.Notify(182755)
			if az.name == z {
				__antithesis_instrumentation__.Notify(182756)
				zones = append(zones, z)
				break
			} else {
				__antithesis_instrumentation__.Notify(182757)
			}
		}
	}
	__antithesis_instrumentation__.Notify(182751)
	return zones, nil
}

func (p *Provider) listRegion(region string, opts ProviderOpts) (vm.List, error) {
	__antithesis_instrumentation__.Notify(182758)
	var data struct {
		Reservations []struct {
			Instances []struct {
				InstanceID string `json:"InstanceId"`
				LaunchTime string
				Placement  struct {
					AvailabilityZone string
				}
				PrivateDNSName   string `json:"PrivateDnsName"`
				PrivateIPAddress string `json:"PrivateIpAddress"`
				PublicDNSName    string `json:"PublicDnsName"`
				PublicIPAddress  string `json:"PublicIpAddress"`
				State            struct {
					Code int
					Name string
				}
				Tags []struct {
					Key   string
					Value string
				}
				VpcID        string `json:"VpcId"`
				InstanceType string
			}
		}
	}
	args := []string{
		"ec2", "describe-instances",
		"--region", region,
	}
	err := p.runJSONCommand(args, &data)
	if err != nil {
		__antithesis_instrumentation__.Notify(182761)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182762)
	}
	__antithesis_instrumentation__.Notify(182759)

	var ret vm.List
	for _, res := range data.Reservations {
		__antithesis_instrumentation__.Notify(182763)
	in:
		for _, in := range res.Instances {
			__antithesis_instrumentation__.Notify(182764)

			if in.State.Name != "pending" && func() bool {
				__antithesis_instrumentation__.Notify(182770)
				return in.State.Name != "running" == true
			}() == true {
				__antithesis_instrumentation__.Notify(182771)
				continue in
			} else {
				__antithesis_instrumentation__.Notify(182772)
			}
			__antithesis_instrumentation__.Notify(182765)
			_ = in.PublicDNSName
			_ = in.State.Code

			tagMap := make(map[string]string, len(in.Tags))
			for _, entry := range in.Tags {
				__antithesis_instrumentation__.Notify(182773)
				tagMap[entry.Key] = entry.Value
			}
			__antithesis_instrumentation__.Notify(182766)

			if tagMap["Roachprod"] != "true" {
				__antithesis_instrumentation__.Notify(182774)
				continue in
			} else {
				__antithesis_instrumentation__.Notify(182775)
			}
			__antithesis_instrumentation__.Notify(182767)

			var errs []error
			createdAt, err := time.Parse(time.RFC3339, in.LaunchTime)
			if err != nil {
				__antithesis_instrumentation__.Notify(182776)
				errs = append(errs, vm.ErrNoExpiration)
			} else {
				__antithesis_instrumentation__.Notify(182777)
			}
			__antithesis_instrumentation__.Notify(182768)

			var lifetime time.Duration
			if lifeText, ok := tagMap["Lifetime"]; ok {
				__antithesis_instrumentation__.Notify(182778)
				lifetime, err = time.ParseDuration(lifeText)
				if err != nil {
					__antithesis_instrumentation__.Notify(182779)
					errs = append(errs, err)
				} else {
					__antithesis_instrumentation__.Notify(182780)
				}
			} else {
				__antithesis_instrumentation__.Notify(182781)
				errs = append(errs, vm.ErrNoExpiration)
			}
			__antithesis_instrumentation__.Notify(182769)

			m := vm.VM{
				CreatedAt:   createdAt,
				DNS:         in.PrivateDNSName,
				Name:        tagMap["Name"],
				Errors:      errs,
				Lifetime:    lifetime,
				Labels:      tagMap,
				PrivateIP:   in.PrivateIPAddress,
				Provider:    ProviderName,
				ProviderID:  in.InstanceID,
				PublicIP:    in.PublicIPAddress,
				RemoteUser:  opts.RemoteUserName,
				VPC:         in.VpcID,
				MachineType: in.InstanceType,
				Zone:        in.Placement.AvailabilityZone,
				SQLPort:     config.DefaultSQLPort,
				AdminUIPort: config.DefaultAdminUIPort,
			}
			ret = append(ret, m)
		}
	}
	__antithesis_instrumentation__.Notify(182760)

	return ret, nil
}

func (p *Provider) runInstance(
	name string, zone string, opts vm.CreateOpts, providerOpts *ProviderOpts,
) error {
	__antithesis_instrumentation__.Notify(182782)

	if opts.SSDOpts.UseLocalSSD && func() bool {
		__antithesis_instrumentation__.Notify(182800)
		return providerOpts.MachineType != defaultMachineType == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(182801)
		return providerOpts.SSDMachineType == defaultSSDMachineType == true
	}() == true {
		__antithesis_instrumentation__.Notify(182802)
		return errors.Errorf("use the --aws-machine-type-ssd flag to set the " +
			"machine type when --local-ssd=true")
	} else {
		__antithesis_instrumentation__.Notify(182803)
		if !opts.SSDOpts.UseLocalSSD && func() bool {
			__antithesis_instrumentation__.Notify(182804)
			return providerOpts.MachineType == defaultMachineType == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(182805)
			return providerOpts.SSDMachineType != defaultSSDMachineType == true
		}() == true {
			__antithesis_instrumentation__.Notify(182806)
			return errors.Errorf("use the --aws-machine-type flag to set the " +
				"machine type when --local-ssd=false")
		} else {
			__antithesis_instrumentation__.Notify(182807)
		}
	}
	__antithesis_instrumentation__.Notify(182783)

	az, ok := p.Config.azByName[zone]
	if !ok {
		__antithesis_instrumentation__.Notify(182808)
		return fmt.Errorf("no region in %v corresponds to availability zone %v",
			p.Config.regionNames(), zone)
	} else {
		__antithesis_instrumentation__.Notify(182809)
	}
	__antithesis_instrumentation__.Notify(182784)

	keyName, err := p.sshKeyName()
	if err != nil {
		__antithesis_instrumentation__.Notify(182810)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182811)
	}
	__antithesis_instrumentation__.Notify(182785)

	var machineType string
	if opts.SSDOpts.UseLocalSSD {
		__antithesis_instrumentation__.Notify(182812)
		machineType = providerOpts.SSDMachineType
	} else {
		__antithesis_instrumentation__.Notify(182813)
		machineType = providerOpts.MachineType
	}
	__antithesis_instrumentation__.Notify(182786)

	cpuOptions := providerOpts.CPUOptions

	m := vm.GetDefaultLabelMap(opts)
	m[vm.TagCreated] = timeutil.Now().Format(time.RFC3339)
	m["Name"] = name
	var awsLabelsNameMap = map[string]string{
		vm.TagCluster:   "Cluster",
		vm.TagCreated:   "Created",
		vm.TagLifetime:  "Lifetime",
		vm.TagRoachprod: "Roachprod",
	}

	var sb strings.Builder
	sb.WriteString("ResourceType=instance,Tags=[")
	for key, value := range opts.CustomLabels {
		__antithesis_instrumentation__.Notify(182814)
		_, ok := m[strings.ToLower(key)]
		if ok {
			__antithesis_instrumentation__.Notify(182816)
			return fmt.Errorf("duplicate label name defined: %s", key)
		} else {
			__antithesis_instrumentation__.Notify(182817)
		}
		__antithesis_instrumentation__.Notify(182815)
		fmt.Fprintf(&sb, "{Key=%s,Value=%s},", key, value)
	}
	__antithesis_instrumentation__.Notify(182787)
	for key, value := range m {
		__antithesis_instrumentation__.Notify(182818)
		if n, ok := awsLabelsNameMap[key]; ok {
			__antithesis_instrumentation__.Notify(182820)
			key = n
		} else {
			__antithesis_instrumentation__.Notify(182821)
		}
		__antithesis_instrumentation__.Notify(182819)
		fmt.Fprintf(&sb, "{Key=%s,Value=%s},", key, value)
	}
	__antithesis_instrumentation__.Notify(182788)
	s := sb.String()
	tagSpecs := fmt.Sprintf("%s]", s[:len(s)-1])

	var data struct {
		Instances []struct {
			InstanceID string `json:"InstanceId"`
		}
	}
	_ = data.Instances
	if len(data.Instances) > 0 {
		__antithesis_instrumentation__.Notify(182822)
		_ = data.Instances[0].InstanceID
	} else {
		__antithesis_instrumentation__.Notify(182823)
	}
	__antithesis_instrumentation__.Notify(182789)

	extraMountOpts := ""

	if opts.SSDOpts.UseLocalSSD {
		__antithesis_instrumentation__.Notify(182824)
		if opts.SSDOpts.NoExt4Barrier {
			__antithesis_instrumentation__.Notify(182825)
			extraMountOpts = "nobarrier"
		} else {
			__antithesis_instrumentation__.Notify(182826)
		}
	} else {
		__antithesis_instrumentation__.Notify(182827)
	}
	__antithesis_instrumentation__.Notify(182790)
	filename, err := writeStartupScript(extraMountOpts, providerOpts.UseMultipleDisks)
	if err != nil {
		__antithesis_instrumentation__.Notify(182828)
		return errors.Wrapf(err, "could not write AWS startup script to temp file")
	} else {
		__antithesis_instrumentation__.Notify(182829)
	}
	__antithesis_instrumentation__.Notify(182791)
	defer func() {
		__antithesis_instrumentation__.Notify(182830)
		_ = os.Remove(filename)
	}()
	__antithesis_instrumentation__.Notify(182792)

	withFlagOverride := func(cfg string, fl *string) string {
		__antithesis_instrumentation__.Notify(182831)
		if *fl == "" {
			__antithesis_instrumentation__.Notify(182833)
			return cfg
		} else {
			__antithesis_instrumentation__.Notify(182834)
		}
		__antithesis_instrumentation__.Notify(182832)
		return *fl
	}
	__antithesis_instrumentation__.Notify(182793)

	args := []string{
		"ec2", "run-instances",
		"--associate-public-ip-address",
		"--count", "1",
		"--instance-type", machineType,
		"--image-id", withFlagOverride(az.region.AMI, &providerOpts.ImageAMI),
		"--key-name", keyName,
		"--region", az.region.Name,
		"--security-group-ids", az.region.SecurityGroup,
		"--subnet-id", az.subnetID,
		"--tag-specifications", tagSpecs,
		"--user-data", "file://" + filename,
	}

	if cpuOptions != "" {
		__antithesis_instrumentation__.Notify(182835)
		args = append(args, "--cpu-options", cpuOptions)
	} else {
		__antithesis_instrumentation__.Notify(182836)
	}
	__antithesis_instrumentation__.Notify(182794)

	if p.IAMProfile != "" {
		__antithesis_instrumentation__.Notify(182837)
		args = append(args, "--iam-instance-profile", "Name="+p.IAMProfile)
	} else {
		__antithesis_instrumentation__.Notify(182838)
	}
	__antithesis_instrumentation__.Notify(182795)

	ebsVolumes := providerOpts.EBSVolumes

	if !opts.SSDOpts.UseLocalSSD {
		__antithesis_instrumentation__.Notify(182839)
		if len(ebsVolumes) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(182841)
			return providerOpts.DefaultEBSVolume.Disk.VolumeType == "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(182842)
			providerOpts.DefaultEBSVolume.Disk.VolumeType = defaultEBSVolumeType
			providerOpts.DefaultEBSVolume.Disk.DeleteOnTermination = true
		} else {
			__antithesis_instrumentation__.Notify(182843)
		}
		__antithesis_instrumentation__.Notify(182840)

		if providerOpts.DefaultEBSVolume.Disk.VolumeType != "" {
			__antithesis_instrumentation__.Notify(182844)

			v := ebsVolumes.newVolume()
			v.Disk = providerOpts.DefaultEBSVolume.Disk
			v.Disk.DeleteOnTermination = true
			ebsVolumes = append(ebsVolumes, v)
		} else {
			__antithesis_instrumentation__.Notify(182845)
		}
	} else {
		__antithesis_instrumentation__.Notify(182846)
	}
	__antithesis_instrumentation__.Notify(182796)

	osDiskVolume := &ebsVolume{
		DeviceName: "/dev/sda1",
		Disk: ebsDisk{
			VolumeType:          defaultEBSVolumeType,
			VolumeSize:          opts.OsVolumeSize,
			DeleteOnTermination: true,
		},
	}
	ebsVolumes = append(ebsVolumes, osDiskVolume)

	mapping, err := json.Marshal(ebsVolumes)
	if err != nil {
		__antithesis_instrumentation__.Notify(182847)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182848)
	}
	__antithesis_instrumentation__.Notify(182797)

	deviceMapping, err := ioutil.TempFile("", "aws-block-device-mapping")
	if err != nil {
		__antithesis_instrumentation__.Notify(182849)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182850)
	}
	__antithesis_instrumentation__.Notify(182798)
	defer deviceMapping.Close()
	if _, err := deviceMapping.Write(mapping); err != nil {
		__antithesis_instrumentation__.Notify(182851)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182852)
	}
	__antithesis_instrumentation__.Notify(182799)
	args = append(args,
		"--block-device-mapping",
		"file://"+deviceMapping.Name(),
	)
	return p.runJSONCommand(args, &data)
}

func (p *Provider) Active() bool {
	__antithesis_instrumentation__.Notify(182853)
	return true
}

func (p *Provider) ProjectActive(project string) bool {
	__antithesis_instrumentation__.Notify(182854)
	return project == ""
}
