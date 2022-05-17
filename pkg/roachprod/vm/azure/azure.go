package azure

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/flagstub"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
)

const (
	ProviderName = "azure"
	remoteUser   = "ubuntu"
	tagComment   = "comment"
	tagSubnet    = "subnetPrefix"
)

var providerInstance = &Provider{}

func Init() error {
	__antithesis_instrumentation__.Notify(183097)
	const cliErr = "please install the Azure CLI utilities " +
		"(https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)"
	const authErr = "please use `az login` to login to Azure"

	providerInstance = New()
	providerInstance.OperationTimeout = 10 * time.Minute
	providerInstance.SyncDelete = false
	if _, err := exec.LookPath("az"); err != nil {
		__antithesis_instrumentation__.Notify(183100)
		vm.Providers[ProviderName] = flagstub.New(&Provider{}, cliErr)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183101)
	}
	__antithesis_instrumentation__.Notify(183098)
	if _, err := providerInstance.getAuthToken(); err != nil {
		__antithesis_instrumentation__.Notify(183102)
		vm.Providers[ProviderName] = flagstub.New(&Provider{}, authErr)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183103)
	}
	__antithesis_instrumentation__.Notify(183099)
	vm.Providers[ProviderName] = providerInstance
	return nil
}

type Provider struct {
	OperationTimeout time.Duration

	SyncDelete bool

	mu struct {
		syncutil.Mutex

		authorizer     autorest.Authorizer
		subscription   subscriptions.Subscription
		resourceGroups map[string]resources.Group
		subnets        map[string]network.Subnet
		securityGroups map[string]network.SecurityGroup
	}
}

func New() *Provider {
	__antithesis_instrumentation__.Notify(183104)
	p := &Provider{}
	p.mu.resourceGroups = make(map[string]resources.Group)
	p.mu.securityGroups = make(map[string]network.SecurityGroup)
	p.mu.subnets = make(map[string]network.Subnet)
	return p
}

func (p *Provider) Active() bool {
	__antithesis_instrumentation__.Notify(183105)
	return true
}

func (p *Provider) ProjectActive(project string) bool {
	__antithesis_instrumentation__.Notify(183106)
	return project == ""
}

func (p *Provider) CleanSSH() error {
	__antithesis_instrumentation__.Notify(183107)
	return nil
}

func (p *Provider) ConfigSSH(zones []string) error {
	__antithesis_instrumentation__.Notify(183108)

	return nil
}

func getAzureDefaultLabelMap(opts vm.CreateOpts) map[string]string {
	__antithesis_instrumentation__.Notify(183109)
	m := vm.GetDefaultLabelMap(opts)
	m[vm.TagCreated] = timeutil.Now().Format(time.RFC3339)
	return m
}

func (p *Provider) Create(
	l *logger.Logger, names []string, opts vm.CreateOpts, vmProviderOpts vm.ProviderOpts,
) error {
	__antithesis_instrumentation__.Notify(183110)
	providerOpts := vmProviderOpts.(*ProviderOpts)

	var sshKey string
	sshFile := os.ExpandEnv("${HOME}/.ssh/id_rsa.pub")
	if _, err := os.Stat(sshFile); err == nil {
		__antithesis_instrumentation__.Notify(183120)
		if bytes, err := ioutil.ReadFile(sshFile); err == nil {
			__antithesis_instrumentation__.Notify(183121)
			sshKey = string(bytes)
		} else {
			__antithesis_instrumentation__.Notify(183122)
			return errors.Wrapf(err, "could not read SSH public key file")
		}
	} else {
		__antithesis_instrumentation__.Notify(183123)
		return errors.Wrapf(err, "could not find SSH public key file")
	}
	__antithesis_instrumentation__.Notify(183111)

	m := getAzureDefaultLabelMap(opts)
	clusterTags := make(map[string]*string)
	for key, value := range opts.CustomLabels {
		__antithesis_instrumentation__.Notify(183124)
		_, ok := m[strings.ToLower(key)]
		if ok {
			__antithesis_instrumentation__.Notify(183126)
			return fmt.Errorf("duplicate label name defined: %s", key)
		} else {
			__antithesis_instrumentation__.Notify(183127)
		}
		__antithesis_instrumentation__.Notify(183125)

		clusterTags[key] = to.StringPtr(value)
	}
	__antithesis_instrumentation__.Notify(183112)
	for key, value := range m {
		__antithesis_instrumentation__.Notify(183128)
		clusterTags[key] = to.StringPtr(value)
	}
	__antithesis_instrumentation__.Notify(183113)

	getClusterResourceGroupName := func(location string) string {
		__antithesis_instrumentation__.Notify(183129)
		return fmt.Sprintf("%s-%s", opts.ClusterName, location)
	}
	__antithesis_instrumentation__.Notify(183114)

	ctx, cancel := context.WithTimeout(context.Background(), p.OperationTimeout)
	defer cancel()

	if len(providerOpts.Locations) == 0 {
		__antithesis_instrumentation__.Notify(183130)
		if opts.GeoDistributed {
			__antithesis_instrumentation__.Notify(183131)
			providerOpts.Locations = defaultLocations
		} else {
			__antithesis_instrumentation__.Notify(183132)
			providerOpts.Locations = []string{defaultLocations[0]}
		}
	} else {
		__antithesis_instrumentation__.Notify(183133)
	}
	__antithesis_instrumentation__.Notify(183115)

	if len(providerOpts.Zone) == 0 {
		__antithesis_instrumentation__.Notify(183134)
		providerOpts.Zone = defaultZone
	} else {
		__antithesis_instrumentation__.Notify(183135)
	}
	__antithesis_instrumentation__.Notify(183116)

	if _, err := p.createVNets(ctx, providerOpts.Locations, *providerOpts); err != nil {
		__antithesis_instrumentation__.Notify(183136)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183137)
	}
	__antithesis_instrumentation__.Notify(183117)

	nodeLocations := vm.ZonePlacement(len(providerOpts.Locations), len(names))

	nodesByLocIdx := make(map[int][]int, len(providerOpts.Locations))
	for nodeIdx, locIdx := range nodeLocations {
		__antithesis_instrumentation__.Notify(183138)
		nodesByLocIdx[locIdx] = append(nodesByLocIdx[locIdx], nodeIdx)
	}
	__antithesis_instrumentation__.Notify(183118)

	errs, _ := errgroup.WithContext(ctx)
	for locIdx, nodes := range nodesByLocIdx {
		__antithesis_instrumentation__.Notify(183139)

		locIdx := locIdx
		nodes := nodes
		errs.Go(func() error {
			__antithesis_instrumentation__.Notify(183140)
			location := providerOpts.Locations[locIdx]

			group, err := p.getOrCreateResourceGroup(
				ctx, getClusterResourceGroupName(location), location, clusterTags)
			if err != nil {
				__antithesis_instrumentation__.Notify(183144)
				return err
			} else {
				__antithesis_instrumentation__.Notify(183145)
			}
			__antithesis_instrumentation__.Notify(183141)

			p.mu.Lock()
			subnet, ok := p.mu.subnets[location]
			p.mu.Unlock()
			if !ok {
				__antithesis_instrumentation__.Notify(183146)
				return errors.Errorf("missing subnet for location %q", location)
			} else {
				__antithesis_instrumentation__.Notify(183147)
			}
			__antithesis_instrumentation__.Notify(183142)

			for _, nodeIdx := range nodes {
				__antithesis_instrumentation__.Notify(183148)
				name := names[nodeIdx]
				errs.Go(func() error {
					__antithesis_instrumentation__.Notify(183149)
					_, err := p.createVM(ctx, group, subnet, name, sshKey, opts, *providerOpts)
					err = errors.Wrapf(err, "creating VM %s", name)
					if err == nil {
						__antithesis_instrumentation__.Notify(183151)
						log.Infof(context.Background(), "created VM %s", name)
					} else {
						__antithesis_instrumentation__.Notify(183152)
					}
					__antithesis_instrumentation__.Notify(183150)
					return err
				})
			}
			__antithesis_instrumentation__.Notify(183143)
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(183119)
	return errs.Wait()
}

func (p *Provider) Delete(vms vm.List) error {
	__antithesis_instrumentation__.Notify(183153)
	ctx, cancel := context.WithTimeout(context.Background(), p.OperationTimeout)
	defer cancel()

	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183159)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183160)
	}
	__antithesis_instrumentation__.Notify(183154)
	client := compute.NewVirtualMachinesClient(*sub.ID)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183161)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183162)
	}
	__antithesis_instrumentation__.Notify(183155)

	var futures []compute.VirtualMachinesDeleteFuture
	for _, vm := range vms {
		__antithesis_instrumentation__.Notify(183163)
		parts, err := parseAzureID(vm.ProviderID)
		if err != nil {
			__antithesis_instrumentation__.Notify(183166)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183167)
		}
		__antithesis_instrumentation__.Notify(183164)
		future, err := client.Delete(ctx, parts.resourceGroup, parts.resourceName, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(183168)
			return errors.Wrapf(err, "could not delete %s", vm.ProviderID)
		} else {
			__antithesis_instrumentation__.Notify(183169)
		}
		__antithesis_instrumentation__.Notify(183165)
		futures = append(futures, future)
	}
	__antithesis_instrumentation__.Notify(183156)

	if !p.SyncDelete {
		__antithesis_instrumentation__.Notify(183170)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(183171)
	}
	__antithesis_instrumentation__.Notify(183157)

	for _, future := range futures {
		__antithesis_instrumentation__.Notify(183172)
		if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
			__antithesis_instrumentation__.Notify(183174)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183175)
		}
		__antithesis_instrumentation__.Notify(183173)
		if _, err := future.Result(client); err != nil {
			__antithesis_instrumentation__.Notify(183176)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183177)
		}
	}
	__antithesis_instrumentation__.Notify(183158)
	return nil
}

func (p *Provider) Reset(vms vm.List) error {
	__antithesis_instrumentation__.Notify(183178)
	return nil
}

func (p *Provider) DeleteCluster(name string) error {
	__antithesis_instrumentation__.Notify(183179)
	ctx, cancel := context.WithTimeout(context.Background(), p.OperationTimeout)
	defer cancel()

	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183186)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183187)
	}
	__antithesis_instrumentation__.Notify(183180)
	client := resources.NewGroupsClient(*sub.SubscriptionID)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183188)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183189)
	}
	__antithesis_instrumentation__.Notify(183181)

	filter := fmt.Sprintf("tagName eq '%s' and tagValue eq '%s'", vm.TagCluster, name)
	it, err := client.ListComplete(ctx, filter, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(183190)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183191)
	}
	__antithesis_instrumentation__.Notify(183182)

	var futures []resources.GroupsDeleteFuture
	for it.NotDone() {
		__antithesis_instrumentation__.Notify(183192)
		group := it.Value()

		future, err := client.Delete(ctx, *group.Name)
		if err != nil {
			__antithesis_instrumentation__.Notify(183194)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183195)
		}
		__antithesis_instrumentation__.Notify(183193)
		log.Infof(context.Background(), "marked Azure resource group %s for deletion\n", *group.Name)
		futures = append(futures, future)

		if err := it.NextWithContext(ctx); err != nil {
			__antithesis_instrumentation__.Notify(183196)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183197)
		}
	}
	__antithesis_instrumentation__.Notify(183183)

	if !p.SyncDelete {
		__antithesis_instrumentation__.Notify(183198)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(183199)
	}
	__antithesis_instrumentation__.Notify(183184)

	for _, future := range futures {
		__antithesis_instrumentation__.Notify(183200)
		if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
			__antithesis_instrumentation__.Notify(183202)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183203)
		}
		__antithesis_instrumentation__.Notify(183201)
		if _, err := future.Result(client); err != nil {
			__antithesis_instrumentation__.Notify(183204)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183205)
		}
	}
	__antithesis_instrumentation__.Notify(183185)
	return nil
}

func (p *Provider) Extend(vms vm.List, lifetime time.Duration) error {
	__antithesis_instrumentation__.Notify(183206)
	ctx, cancel := context.WithTimeout(context.Background(), p.OperationTimeout)
	defer cancel()

	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183211)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183212)
	}
	__antithesis_instrumentation__.Notify(183207)
	client := compute.NewVirtualMachinesClient(*sub.ID)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183213)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183214)
	}
	__antithesis_instrumentation__.Notify(183208)

	futures := make([]compute.VirtualMachinesUpdateFuture, len(vms))
	for idx, m := range vms {
		__antithesis_instrumentation__.Notify(183215)
		vmParts, err := parseAzureID(m.ProviderID)
		if err != nil {
			__antithesis_instrumentation__.Notify(183217)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183218)
		}
		__antithesis_instrumentation__.Notify(183216)
		update := compute.VirtualMachineUpdate{
			Tags: map[string]*string{
				vm.TagLifetime: to.StringPtr(lifetime.String()),
			},
		}
		futures[idx], err = client.Update(ctx, vmParts.resourceGroup, vmParts.resourceName, update)
		if err != nil {
			__antithesis_instrumentation__.Notify(183219)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183220)
		}
	}
	__antithesis_instrumentation__.Notify(183209)

	for _, future := range futures {
		__antithesis_instrumentation__.Notify(183221)
		if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
			__antithesis_instrumentation__.Notify(183223)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183224)
		}
		__antithesis_instrumentation__.Notify(183222)
		if _, err := future.Result(client); err != nil {
			__antithesis_instrumentation__.Notify(183225)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183226)
		}
	}
	__antithesis_instrumentation__.Notify(183210)
	return nil
}

func (p *Provider) FindActiveAccount() (string, error) {
	__antithesis_instrumentation__.Notify(183227)

	token, err := p.getAuthToken()
	if err != nil {
		__antithesis_instrumentation__.Notify(183231)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183232)
	}
	__antithesis_instrumentation__.Notify(183228)

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		__antithesis_instrumentation__.Notify(183233)
		return "", errors.Errorf("unexpected number of segments; expected 3, had %d", len(parts))
	} else {
		__antithesis_instrumentation__.Notify(183234)
	}
	__antithesis_instrumentation__.Notify(183229)

	a := base64.NewDecoder(base64.RawStdEncoding, strings.NewReader(parts[1]))
	var data struct {
		Username string `json:"upn"`
	}
	if err := json.NewDecoder(a).Decode(&data); err != nil {
		__antithesis_instrumentation__.Notify(183235)
		return "", errors.Wrapf(err, "could not decode JWT claims segment")
	} else {
		__antithesis_instrumentation__.Notify(183236)
	}
	__antithesis_instrumentation__.Notify(183230)

	return data.Username[:strings.Index(data.Username, "@")], nil
}

func (p *Provider) List(l *logger.Logger) (vm.List, error) {
	__antithesis_instrumentation__.Notify(183237)
	ctx, cancel := context.WithTimeout(context.Background(), p.OperationTimeout)
	defer cancel()

	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183242)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(183243)
	}
	__antithesis_instrumentation__.Notify(183238)

	client := compute.NewVirtualMachinesClient(*sub.SubscriptionID)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183244)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(183245)
	}
	__antithesis_instrumentation__.Notify(183239)

	it, err := client.ListAllComplete(ctx, "false")
	if err != nil {
		__antithesis_instrumentation__.Notify(183246)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(183247)
	}
	__antithesis_instrumentation__.Notify(183240)

	var ret vm.List
	for it.NotDone() {
		__antithesis_instrumentation__.Notify(183248)
		found := it.Value()
		if _, ok := found.Tags[vm.TagRoachprod]; !ok {
			__antithesis_instrumentation__.Notify(183255)
			if err := it.NextWithContext(ctx); err != nil {
				__antithesis_instrumentation__.Notify(183257)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(183258)
			}
			__antithesis_instrumentation__.Notify(183256)
			continue
		} else {
			__antithesis_instrumentation__.Notify(183259)
		}
		__antithesis_instrumentation__.Notify(183249)

		tags := make(map[string]string)
		for key, value := range found.Tags {
			__antithesis_instrumentation__.Notify(183260)
			tags[key] = *value
		}
		__antithesis_instrumentation__.Notify(183250)

		m := vm.VM{
			Name:        *found.Name,
			Labels:      tags,
			Provider:    ProviderName,
			ProviderID:  *found.ID,
			RemoteUser:  remoteUser,
			VPC:         "global",
			MachineType: string(found.HardwareProfile.VMSize),

			Zone:        *found.Location + "z",
			SQLPort:     config.DefaultSQLPort,
			AdminUIPort: config.DefaultAdminUIPort,
		}

		if createdPtr := found.Tags[vm.TagCreated]; createdPtr == nil {
			__antithesis_instrumentation__.Notify(183261)
			m.Errors = append(m.Errors, vm.ErrNoExpiration)
		} else {
			__antithesis_instrumentation__.Notify(183262)
			if parsed, err := time.Parse(time.RFC3339, *createdPtr); err == nil {
				__antithesis_instrumentation__.Notify(183263)
				m.CreatedAt = parsed
			} else {
				__antithesis_instrumentation__.Notify(183264)
				m.Errors = append(m.Errors, vm.ErrNoExpiration)
			}
		}
		__antithesis_instrumentation__.Notify(183251)

		if lifetimePtr := found.Tags[vm.TagLifetime]; lifetimePtr == nil {
			__antithesis_instrumentation__.Notify(183265)
			m.Errors = append(m.Errors, vm.ErrNoExpiration)
		} else {
			__antithesis_instrumentation__.Notify(183266)
			if parsed, err := time.ParseDuration(*lifetimePtr); err == nil {
				__antithesis_instrumentation__.Notify(183267)
				m.Lifetime = parsed
			} else {
				__antithesis_instrumentation__.Notify(183268)
				m.Errors = append(m.Errors, vm.ErrNoExpiration)
			}
		}
		__antithesis_instrumentation__.Notify(183252)

		nicID, err := parseAzureID(*(*found.NetworkProfile.NetworkInterfaces)[0].ID)
		if err != nil {
			__antithesis_instrumentation__.Notify(183269)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(183270)
		}
		__antithesis_instrumentation__.Notify(183253)
		if err := p.fillNetworkDetails(ctx, &m, nicID); errors.Is(err, vm.ErrBadNetwork) {
			__antithesis_instrumentation__.Notify(183271)
			m.Errors = append(m.Errors, err)
		} else {
			__antithesis_instrumentation__.Notify(183272)
			if err != nil {
				__antithesis_instrumentation__.Notify(183273)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(183274)
			}
		}
		__antithesis_instrumentation__.Notify(183254)

		ret = append(ret, m)

		if err := it.NextWithContext(ctx); err != nil {
			__antithesis_instrumentation__.Notify(183275)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(183276)
		}

	}
	__antithesis_instrumentation__.Notify(183241)

	return ret, nil
}

func (p *Provider) Name() string {
	__antithesis_instrumentation__.Notify(183277)
	return ProviderName
}

func (p *Provider) createVM(
	ctx context.Context,
	group resources.Group,
	subnet network.Subnet,
	name, sshKey string,
	opts vm.CreateOpts,
	providerOpts ProviderOpts,
) (vm compute.VirtualMachine, err error) {
	__antithesis_instrumentation__.Notify(183278)
	startupArgs := azureStartupArgs{RemoteUser: remoteUser}
	if !opts.SSDOpts.UseLocalSSD {
		__antithesis_instrumentation__.Notify(183291)

		lun := 42
		startupArgs.AttachedDiskLun = &lun
	} else {
		__antithesis_instrumentation__.Notify(183292)
	}
	__antithesis_instrumentation__.Notify(183279)
	startupScript, err := evalStartupTemplate(startupArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(183293)
		return vm, err
	} else {
		__antithesis_instrumentation__.Notify(183294)
	}
	__antithesis_instrumentation__.Notify(183280)
	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183295)
		return
	} else {
		__antithesis_instrumentation__.Notify(183296)
	}
	__antithesis_instrumentation__.Notify(183281)

	client := compute.NewVirtualMachinesClient(*sub.SubscriptionID)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183297)
		return
	} else {
		__antithesis_instrumentation__.Notify(183298)
	}
	__antithesis_instrumentation__.Notify(183282)

	ip, err := p.createIP(ctx, group, name, providerOpts)
	if err != nil {
		__antithesis_instrumentation__.Notify(183299)
		return
	} else {
		__antithesis_instrumentation__.Notify(183300)
	}
	__antithesis_instrumentation__.Notify(183283)
	nic, err := p.createNIC(ctx, group, ip, subnet)
	if err != nil {
		__antithesis_instrumentation__.Notify(183301)
		return
	} else {
		__antithesis_instrumentation__.Notify(183302)
	}
	__antithesis_instrumentation__.Notify(183284)

	tags := make(map[string]*string)
	for key, value := range opts.CustomLabels {
		__antithesis_instrumentation__.Notify(183303)
		tags[key] = to.StringPtr(value)
	}
	__antithesis_instrumentation__.Notify(183285)
	m := getAzureDefaultLabelMap(opts)
	for key, value := range m {
		__antithesis_instrumentation__.Notify(183304)
		tags[key] = to.StringPtr(value)
	}
	__antithesis_instrumentation__.Notify(183286)

	osVolumeSize := int32(opts.OsVolumeSize)
	if osVolumeSize < 32 {
		__antithesis_instrumentation__.Notify(183305)
		log.Info(context.Background(), "WARNING: increasing the OS volume size to minimally allowed 32GB")
		osVolumeSize = 32
	} else {
		__antithesis_instrumentation__.Notify(183306)
	}
	__antithesis_instrumentation__.Notify(183287)

	vm = compute.VirtualMachine{
		Location: group.Location,
		Zones:    to.StringSlicePtr([]string{providerOpts.Zone}),
		Tags:     tags,
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			HardwareProfile: &compute.HardwareProfile{
				VMSize: compute.VirtualMachineSizeTypes(providerOpts.MachineType),
			},
			StorageProfile: &compute.StorageProfile{

				ImageReference: &compute.ImageReference{
					Publisher: to.StringPtr("Canonical"),
					Offer:     to.StringPtr("0001-com-ubuntu-server-focal"),
					Sku:       to.StringPtr("20_04-lts"),
					Version:   to.StringPtr("20.04.202109080"),
				},
				OsDisk: &compute.OSDisk{
					CreateOption: compute.DiskCreateOptionTypesFromImage,
					ManagedDisk: &compute.ManagedDiskParameters{
						StorageAccountType: compute.StorageAccountTypesStandardSSDLRS,
					},
					DiskSizeGB: to.Int32Ptr(osVolumeSize),
				},
			},
			OsProfile: &compute.OSProfile{
				ComputerName:  to.StringPtr(name),
				AdminUsername: to.StringPtr(remoteUser),

				CustomData: to.StringPtr(startupScript),
				LinuxConfiguration: &compute.LinuxConfiguration{
					SSH: &compute.SSHConfiguration{
						PublicKeys: &[]compute.SSHPublicKey{
							{
								Path:    to.StringPtr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", remoteUser)),
								KeyData: to.StringPtr(sshKey),
							},
						},
					},
				},
			},
			NetworkProfile: &compute.NetworkProfile{
				NetworkInterfaces: &[]compute.NetworkInterfaceReference{
					{
						ID: nic.ID,
						NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
							Primary: to.BoolPtr(true),
						},
					},
				},
			},
		},
	}
	if !opts.SSDOpts.UseLocalSSD {
		__antithesis_instrumentation__.Notify(183307)
		caching := compute.CachingTypesNone

		switch providerOpts.DiskCaching {
		case "read-only":
			__antithesis_instrumentation__.Notify(183310)
			caching = compute.CachingTypesReadOnly
		case "read-write":
			__antithesis_instrumentation__.Notify(183311)
			caching = compute.CachingTypesReadWrite
		case "none":
			__antithesis_instrumentation__.Notify(183312)
			caching = compute.CachingTypesNone
		default:
			__antithesis_instrumentation__.Notify(183313)
			err = errors.Newf("unsupported caching behavior: %s", providerOpts.DiskCaching)
			return
		}
		__antithesis_instrumentation__.Notify(183308)
		dataDisks := []compute.DataDisk{
			{
				DiskSizeGB: to.Int32Ptr(providerOpts.NetworkDiskSize),
				Caching:    caching,
				Lun:        to.Int32Ptr(42),
			},
		}

		switch providerOpts.NetworkDiskType {
		case "ultra-disk":
			__antithesis_instrumentation__.Notify(183314)
			var ultraDisk compute.Disk
			ultraDisk, err = p.createUltraDisk(ctx, group, name+"-ultra-disk", providerOpts)
			if err != nil {
				__antithesis_instrumentation__.Notify(183318)
				return
			} else {
				__antithesis_instrumentation__.Notify(183319)
			}
			__antithesis_instrumentation__.Notify(183315)

			dataDisks[0].CreateOption = compute.DiskCreateOptionTypesAttach
			dataDisks[0].Name = ultraDisk.Name
			dataDisks[0].ManagedDisk = &compute.ManagedDiskParameters{
				StorageAccountType: compute.StorageAccountTypesUltraSSDLRS,
				ID:                 ultraDisk.ID,
			}

			vm.AdditionalCapabilities = &compute.AdditionalCapabilities{
				UltraSSDEnabled: to.BoolPtr(true),
			}
		case "premium-disk":
			__antithesis_instrumentation__.Notify(183316)

			dataDisks[0].CreateOption = compute.DiskCreateOptionTypesEmpty
			dataDisks[0].ManagedDisk = &compute.ManagedDiskParameters{
				StorageAccountType: compute.StorageAccountTypesPremiumLRS,
			}
		default:
			__antithesis_instrumentation__.Notify(183317)
			err = errors.Newf("unsupported network disk type: %s", providerOpts.NetworkDiskType)
			return
		}
		__antithesis_instrumentation__.Notify(183309)

		vm.StorageProfile.DataDisks = &dataDisks
	} else {
		__antithesis_instrumentation__.Notify(183320)
	}
	__antithesis_instrumentation__.Notify(183288)
	future, err := client.CreateOrUpdate(ctx, *group.Name, name, vm)
	if err != nil {
		__antithesis_instrumentation__.Notify(183321)
		return
	} else {
		__antithesis_instrumentation__.Notify(183322)
	}
	__antithesis_instrumentation__.Notify(183289)
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		__antithesis_instrumentation__.Notify(183323)
		return
	} else {
		__antithesis_instrumentation__.Notify(183324)
	}
	__antithesis_instrumentation__.Notify(183290)
	return future.Result(client)
}

func (p *Provider) createNIC(
	ctx context.Context, group resources.Group, ip network.PublicIPAddress, subnet network.Subnet,
) (iface network.Interface, err error) {
	__antithesis_instrumentation__.Notify(183325)
	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183331)
		return
	} else {
		__antithesis_instrumentation__.Notify(183332)
	}
	__antithesis_instrumentation__.Notify(183326)
	client := network.NewInterfacesClient(*sub.SubscriptionID)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183333)
		return
	} else {
		__antithesis_instrumentation__.Notify(183334)
	}
	__antithesis_instrumentation__.Notify(183327)

	p.mu.Lock()
	sg := p.mu.securityGroups[p.getVnetNetworkSecurityGroupName(*group.Location)]
	p.mu.Unlock()

	future, err := client.CreateOrUpdate(ctx, *group.Name, *ip.Name, network.Interface{
		Name:     ip.Name,
		Location: group.Location,
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				{
					Name: to.StringPtr("ipConfig"),
					InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
						Subnet:                    &subnet,
						PrivateIPAllocationMethod: network.IPAllocationMethodDynamic,
						PublicIPAddress:           &ip,
					},
				},
			},
			NetworkSecurityGroup:        &sg,
			EnableAcceleratedNetworking: to.BoolPtr(true),
			Primary:                     to.BoolPtr(true),
		},
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(183335)
		return
	} else {
		__antithesis_instrumentation__.Notify(183336)
	}
	__antithesis_instrumentation__.Notify(183328)
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		__antithesis_instrumentation__.Notify(183337)
		return
	} else {
		__antithesis_instrumentation__.Notify(183338)
	}
	__antithesis_instrumentation__.Notify(183329)
	iface, err = future.Result(client)
	if err == nil {
		__antithesis_instrumentation__.Notify(183339)
		log.Infof(context.Background(), "created NIC %s %s", *iface.Name, *(*iface.IPConfigurations)[0].PrivateIPAddress)
	} else {
		__antithesis_instrumentation__.Notify(183340)
	}
	__antithesis_instrumentation__.Notify(183330)
	return
}

func (p *Provider) getOrCreateNetworkSecurityGroup(
	ctx context.Context, name string, resourceGroup resources.Group,
) (network.SecurityGroup, error) {
	__antithesis_instrumentation__.Notify(183341)
	p.mu.Lock()
	group, ok := p.mu.securityGroups[name]
	p.mu.Unlock()
	if ok {
		__antithesis_instrumentation__.Notify(183350)
		return group, nil
	} else {
		__antithesis_instrumentation__.Notify(183351)
	}
	__antithesis_instrumentation__.Notify(183342)

	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183352)
		return network.SecurityGroup{}, err
	} else {
		__antithesis_instrumentation__.Notify(183353)
	}
	__antithesis_instrumentation__.Notify(183343)
	client := network.NewSecurityGroupsClient(*sub.SubscriptionID)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183354)
		return network.SecurityGroup{}, err
	} else {
		__antithesis_instrumentation__.Notify(183355)
	}
	__antithesis_instrumentation__.Notify(183344)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183356)
		return network.SecurityGroup{}, err
	} else {
		__antithesis_instrumentation__.Notify(183357)
	}
	__antithesis_instrumentation__.Notify(183345)

	cacheAndReturn := func(group network.SecurityGroup) (network.SecurityGroup, error) {
		__antithesis_instrumentation__.Notify(183358)
		p.mu.Lock()
		p.mu.securityGroups[name] = group
		p.mu.Unlock()
		return group, nil
	}
	__antithesis_instrumentation__.Notify(183346)

	future, err := client.CreateOrUpdate(ctx, *resourceGroup.Name, name, network.SecurityGroup{
		SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
			SecurityRules: &[]network.SecurityRule{
				{
					Name: to.StringPtr("SSH_Inbound"),
					SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
						Priority:                 to.Int32Ptr(300),
						Protocol:                 network.SecurityRuleProtocolTCP,
						Access:                   network.SecurityRuleAccessAllow,
						Direction:                network.SecurityRuleDirectionInbound,
						SourceAddressPrefix:      to.StringPtr("*"),
						SourcePortRange:          to.StringPtr("*"),
						DestinationAddressPrefix: to.StringPtr("*"),
						DestinationPortRange:     to.StringPtr("22"),
					},
				},
				{
					Name: to.StringPtr("SSH_Outbound"),
					SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
						Priority:                 to.Int32Ptr(301),
						Protocol:                 network.SecurityRuleProtocolTCP,
						Access:                   network.SecurityRuleAccessAllow,
						Direction:                network.SecurityRuleDirectionOutbound,
						SourceAddressPrefix:      to.StringPtr("*"),
						SourcePortRange:          to.StringPtr("*"),
						DestinationAddressPrefix: to.StringPtr("*"),
						DestinationPortRange:     to.StringPtr("*"),
					},
				},
				{
					Name: to.StringPtr("HTTP_Inbound"),
					SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
						Priority:                 to.Int32Ptr(320),
						Protocol:                 network.SecurityRuleProtocolTCP,
						Access:                   network.SecurityRuleAccessAllow,
						Direction:                network.SecurityRuleDirectionInbound,
						SourceAddressPrefix:      to.StringPtr("*"),
						SourcePortRange:          to.StringPtr("*"),
						DestinationAddressPrefix: to.StringPtr("*"),
						DestinationPortRange:     to.StringPtr("80"),
					},
				},
				{
					Name: to.StringPtr("HTTP_Outbound"),
					SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
						Priority:                 to.Int32Ptr(321),
						Protocol:                 network.SecurityRuleProtocolTCP,
						Access:                   network.SecurityRuleAccessAllow,
						Direction:                network.SecurityRuleDirectionOutbound,
						SourceAddressPrefix:      to.StringPtr("*"),
						SourcePortRange:          to.StringPtr("*"),
						DestinationAddressPrefix: to.StringPtr("*"),
						DestinationPortRange:     to.StringPtr("*"),
					},
				},
				{
					Name: to.StringPtr("HTTPS_Inbound"),
					SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
						Priority:                 to.Int32Ptr(340),
						Protocol:                 network.SecurityRuleProtocolTCP,
						Access:                   network.SecurityRuleAccessAllow,
						Direction:                network.SecurityRuleDirectionInbound,
						SourceAddressPrefix:      to.StringPtr("*"),
						SourcePortRange:          to.StringPtr("*"),
						DestinationAddressPrefix: to.StringPtr("*"),
						DestinationPortRange:     to.StringPtr("443"),
					},
				},
				{
					Name: to.StringPtr("HTTPS_Outbound"),
					SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
						Priority:                 to.Int32Ptr(341),
						Protocol:                 network.SecurityRuleProtocolTCP,
						Access:                   network.SecurityRuleAccessAllow,
						Direction:                network.SecurityRuleDirectionOutbound,
						SourceAddressPrefix:      to.StringPtr("*"),
						SourcePortRange:          to.StringPtr("*"),
						DestinationAddressPrefix: to.StringPtr("*"),
						DestinationPortRange:     to.StringPtr("*"),
					},
				},
				{
					Name: to.StringPtr("CockroachPG_Inbound"),
					SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
						Priority:                 to.Int32Ptr(342),
						Protocol:                 network.SecurityRuleProtocolTCP,
						Access:                   network.SecurityRuleAccessAllow,
						Direction:                network.SecurityRuleDirectionInbound,
						SourceAddressPrefix:      to.StringPtr("*"),
						SourcePortRange:          to.StringPtr("*"),
						DestinationAddressPrefix: to.StringPtr("*"),
						DestinationPortRange:     to.StringPtr("26257"),
					},
				},
				{
					Name: to.StringPtr("CockroachAdmin_Inbound"),
					SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
						Priority:                 to.Int32Ptr(343),
						Protocol:                 network.SecurityRuleProtocolTCP,
						Access:                   network.SecurityRuleAccessAllow,
						Direction:                network.SecurityRuleDirectionInbound,
						SourceAddressPrefix:      to.StringPtr("*"),
						SourcePortRange:          to.StringPtr("*"),
						DestinationAddressPrefix: to.StringPtr("*"),
						DestinationPortRange:     to.StringPtr("26258"),
					},
				},
			},
		},
		Location: resourceGroup.Location,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(183359)
		return network.SecurityGroup{}, err
	} else {
		__antithesis_instrumentation__.Notify(183360)
	}
	__antithesis_instrumentation__.Notify(183347)
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		__antithesis_instrumentation__.Notify(183361)
		return network.SecurityGroup{}, err
	} else {
		__antithesis_instrumentation__.Notify(183362)
	}
	__antithesis_instrumentation__.Notify(183348)
	securityGroup, err := future.Result(client)
	if err != nil {
		__antithesis_instrumentation__.Notify(183363)
		return network.SecurityGroup{}, err
	} else {
		__antithesis_instrumentation__.Notify(183364)
	}
	__antithesis_instrumentation__.Notify(183349)

	return cacheAndReturn(securityGroup)
}

func (p *Provider) getVnetNetworkSecurityGroupName(location string) string {
	__antithesis_instrumentation__.Notify(183365)
	return fmt.Sprintf("roachprod-vnets-nsg-%s", location)
}

func (p *Provider) createVNets(
	ctx context.Context, locations []string, providerOpts ProviderOpts,
) (map[string]network.VirtualNetwork, error) {
	__antithesis_instrumentation__.Notify(183366)
	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183375)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(183376)
	}
	__antithesis_instrumentation__.Notify(183367)

	groupsClient := resources.NewGroupsClient(*sub.SubscriptionID)

	vnetResourceGroupTags := make(map[string]*string)
	vnetResourceGroupTags[tagComment] = to.StringPtr("DO NOT DELETE: Used by all roachprod clusters")
	vnetResourceGroupTags[vm.TagRoachprod] = to.StringPtr("true")

	vnetResourceGroupName := func(location string) string {
		__antithesis_instrumentation__.Notify(183377)
		return fmt.Sprintf("roachprod-vnets-%s", location)
	}
	__antithesis_instrumentation__.Notify(183368)

	setVNetSubnetPrefix := func(group resources.Group, subnet int) (resources.Group, error) {
		__antithesis_instrumentation__.Notify(183378)
		return groupsClient.Update(ctx, *group.Name, resources.GroupPatchable{
			Tags: map[string]*string{
				tagSubnet: to.StringPtr(strconv.Itoa(subnet)),
			},
		})
	}
	__antithesis_instrumentation__.Notify(183369)

	for _, location := range locations {
		__antithesis_instrumentation__.Notify(183379)
		group, err := p.getOrCreateResourceGroup(ctx, vnetResourceGroupName(location), location, vnetResourceGroupTags)
		if err != nil {
			__antithesis_instrumentation__.Notify(183381)
			return nil, errors.Wrapf(err, "resource group for location %q", location)
		} else {
			__antithesis_instrumentation__.Notify(183382)
		}
		__antithesis_instrumentation__.Notify(183380)
		_, err = p.getOrCreateNetworkSecurityGroup(ctx, p.getVnetNetworkSecurityGroupName(location), group)
		if err != nil {
			__antithesis_instrumentation__.Notify(183383)
			return nil, errors.Wrapf(err, "nsg for location %q", location)
		} else {
			__antithesis_instrumentation__.Notify(183384)
		}
	}
	__antithesis_instrumentation__.Notify(183370)

	prefixesByLocation := make(map[string]int)
	activePrefixes := make(map[int]bool)

	nextAvailablePrefix := func() int {
		__antithesis_instrumentation__.Notify(183385)
		prefix := 1
		for activePrefixes[prefix] {
			__antithesis_instrumentation__.Notify(183387)
			prefix++
		}
		__antithesis_instrumentation__.Notify(183386)
		activePrefixes[prefix] = true
		return prefix
	}
	__antithesis_instrumentation__.Notify(183371)
	newSubnetsCreated := false

	for _, location := range providerOpts.Locations {
		__antithesis_instrumentation__.Notify(183388)
		p.mu.Lock()
		group := p.mu.resourceGroups[vnetResourceGroupName(location)]
		p.mu.Unlock()

		if prefixString := group.Tags[tagSubnet]; prefixString != nil {
			__antithesis_instrumentation__.Notify(183389)
			prefix, err := strconv.Atoi(*prefixString)
			if err != nil {
				__antithesis_instrumentation__.Notify(183391)
				return nil, errors.Wrapf(err, "for location %q", location)
			} else {
				__antithesis_instrumentation__.Notify(183392)
			}
			__antithesis_instrumentation__.Notify(183390)
			activePrefixes[prefix] = true
			prefixesByLocation[location] = prefix
		} else {
			__antithesis_instrumentation__.Notify(183393)

			newSubnetsCreated = true
			prefix := nextAvailablePrefix()
			prefixesByLocation[location] = prefix
			p.mu.Lock()
			group := p.mu.resourceGroups[vnetResourceGroupName(location)]
			p.mu.Unlock()
			group, err = setVNetSubnetPrefix(group, prefix)
			if err != nil {
				__antithesis_instrumentation__.Notify(183395)
				return nil, errors.Wrapf(err, "for location %q", location)
			} else {
				__antithesis_instrumentation__.Notify(183396)
			}
			__antithesis_instrumentation__.Notify(183394)

			p.mu.Lock()
			p.mu.resourceGroups[vnetResourceGroupName(location)] = group
			p.mu.Unlock()
		}
	}
	__antithesis_instrumentation__.Notify(183372)

	ret := make(map[string]network.VirtualNetwork)
	vnets := make([]network.VirtualNetwork, len(ret))
	for location, prefix := range prefixesByLocation {
		__antithesis_instrumentation__.Notify(183397)
		p.mu.Lock()
		resourceGroup := p.mu.resourceGroups[vnetResourceGroupName(location)]
		networkSecurityGroup := p.mu.securityGroups[p.getVnetNetworkSecurityGroupName(location)]
		p.mu.Unlock()
		if vnet, _, err := p.createVNet(ctx, resourceGroup, networkSecurityGroup, prefix, providerOpts); err == nil {
			__antithesis_instrumentation__.Notify(183398)
			ret[location] = vnet
			vnets = append(vnets, vnet)
		} else {
			__antithesis_instrumentation__.Notify(183399)
			return nil, errors.Wrapf(err, "for location %q", location)
		}
	}
	__antithesis_instrumentation__.Notify(183373)

	if newSubnetsCreated {
		__antithesis_instrumentation__.Notify(183400)
		return ret, p.createVNetPeerings(ctx, vnets)
	} else {
		__antithesis_instrumentation__.Notify(183401)
	}
	__antithesis_instrumentation__.Notify(183374)
	return ret, nil
}

func (p *Provider) createVNet(
	ctx context.Context,
	resourceGroup resources.Group,
	securityGroup network.SecurityGroup,
	prefix int,
	providerOpts ProviderOpts,
) (vnet network.VirtualNetwork, subnet network.Subnet, err error) {
	__antithesis_instrumentation__.Notify(183402)
	vnetName := providerOpts.VnetName

	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183408)
		return
	} else {
		__antithesis_instrumentation__.Notify(183409)
	}
	__antithesis_instrumentation__.Notify(183403)
	client := network.NewVirtualNetworksClient(*sub.SubscriptionID)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183410)
		return
	} else {
		__antithesis_instrumentation__.Notify(183411)
	}
	__antithesis_instrumentation__.Notify(183404)
	vnet = network.VirtualNetwork{
		Name:     to.StringPtr(vnetName),
		Location: resourceGroup.Location,
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{
				AddressPrefixes: &[]string{fmt.Sprintf("10.%d.0.0/16", prefix)},
			},
			Subnets: &[]network.Subnet{
				{
					Name: resourceGroup.Name,
					SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
						AddressPrefix:        to.StringPtr(fmt.Sprintf("10.%d.0.0/18", prefix)),
						NetworkSecurityGroup: &securityGroup,
					},
				},
			},
		},
	}
	future, err := client.CreateOrUpdate(ctx, *resourceGroup.Name, *resourceGroup.Name, vnet)
	if err != nil {
		__antithesis_instrumentation__.Notify(183412)
		err = errors.Wrapf(err, "creating Azure VNet %q in %q", vnetName, *resourceGroup.Name)
		return
	} else {
		__antithesis_instrumentation__.Notify(183413)
	}
	__antithesis_instrumentation__.Notify(183405)
	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		__antithesis_instrumentation__.Notify(183414)
		err = errors.Wrapf(err, "creating Azure VNet %q in %q", vnetName, *resourceGroup.Name)
		return
	} else {
		__antithesis_instrumentation__.Notify(183415)
	}
	__antithesis_instrumentation__.Notify(183406)
	vnet, err = future.Result(client)
	err = errors.Wrapf(err, "creating Azure VNet %q in %q", vnetName, *resourceGroup.Name)
	if err != nil {
		__antithesis_instrumentation__.Notify(183416)
		return
	} else {
		__antithesis_instrumentation__.Notify(183417)
	}
	__antithesis_instrumentation__.Notify(183407)
	subnet = (*vnet.Subnets)[0]
	p.mu.Lock()
	p.mu.subnets[*resourceGroup.Location] = subnet
	p.mu.Unlock()
	log.Infof(context.Background(), "created Azure VNet %q in %q with prefix %d", vnetName, *resourceGroup.Name, prefix)
	return
}

func (p *Provider) createVNetPeerings(ctx context.Context, vnets []network.VirtualNetwork) error {
	__antithesis_instrumentation__.Notify(183418)
	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183423)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183424)
	}
	__antithesis_instrumentation__.Notify(183419)
	client := network.NewVirtualNetworkPeeringsClient(*sub.SubscriptionID)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183425)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183426)
	}
	__antithesis_instrumentation__.Notify(183420)

	futures := make(map[string]network.VirtualNetworkPeeringsCreateOrUpdateFuture)
	for _, outer := range vnets {
		__antithesis_instrumentation__.Notify(183427)
		for _, inner := range vnets {
			__antithesis_instrumentation__.Notify(183428)
			if *outer.ID == *inner.ID {
				__antithesis_instrumentation__.Notify(183432)
				continue
			} else {
				__antithesis_instrumentation__.Notify(183433)
			}
			__antithesis_instrumentation__.Notify(183429)

			linkName := fmt.Sprintf("%s-%s", *outer.Name, *inner.Name)
			peering := network.VirtualNetworkPeering{
				Name: to.StringPtr(linkName),
				VirtualNetworkPeeringPropertiesFormat: &network.VirtualNetworkPeeringPropertiesFormat{
					AllowForwardedTraffic:     to.BoolPtr(true),
					AllowVirtualNetworkAccess: to.BoolPtr(true),
					RemoteAddressSpace:        inner.AddressSpace,
					RemoteVirtualNetwork: &network.SubResource{
						ID: inner.ID,
					},
				},
			}

			outerParts, err := parseAzureID(*outer.ID)
			if err != nil {
				__antithesis_instrumentation__.Notify(183434)
				return err
			} else {
				__antithesis_instrumentation__.Notify(183435)
			}
			__antithesis_instrumentation__.Notify(183430)

			future, err := client.CreateOrUpdate(ctx, outerParts.resourceGroup, *outer.Name, linkName, peering, network.SyncRemoteAddressSpaceTrue)
			if err != nil {
				__antithesis_instrumentation__.Notify(183436)
				return errors.Wrapf(err, "creating vnet peering %s", linkName)
			} else {
				__antithesis_instrumentation__.Notify(183437)
			}
			__antithesis_instrumentation__.Notify(183431)
			futures[linkName] = future
		}
	}
	__antithesis_instrumentation__.Notify(183421)

	for name, future := range futures {
		__antithesis_instrumentation__.Notify(183438)
		if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
			__antithesis_instrumentation__.Notify(183441)
			return errors.Wrapf(err, "creating vnet peering %s", name)
		} else {
			__antithesis_instrumentation__.Notify(183442)
		}
		__antithesis_instrumentation__.Notify(183439)
		peering, err := future.Result(client)
		if err != nil {
			__antithesis_instrumentation__.Notify(183443)
			return errors.Wrapf(err, "creating vnet peering %s", name)
		} else {
			__antithesis_instrumentation__.Notify(183444)
		}
		__antithesis_instrumentation__.Notify(183440)
		log.Infof(context.Background(), "created vnet peering %s", *peering.Name)
	}
	__antithesis_instrumentation__.Notify(183422)

	return nil
}

func (p *Provider) createIP(
	ctx context.Context, group resources.Group, name string, providerOpts ProviderOpts,
) (ip network.PublicIPAddress, err error) {
	__antithesis_instrumentation__.Notify(183445)
	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183451)
		return
	} else {
		__antithesis_instrumentation__.Notify(183452)
	}
	__antithesis_instrumentation__.Notify(183446)
	ipc := network.NewPublicIPAddressesClient(*sub.SubscriptionID)
	if ipc.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183453)
		return
	} else {
		__antithesis_instrumentation__.Notify(183454)
	}
	__antithesis_instrumentation__.Notify(183447)
	future, err := ipc.CreateOrUpdate(ctx, *group.Name, name,
		network.PublicIPAddress{
			Name: to.StringPtr(name),
			Sku: &network.PublicIPAddressSku{
				Name: network.PublicIPAddressSkuNameStandard,
			},
			Location: group.Location,
			Zones:    to.StringSlicePtr([]string{providerOpts.Zone}),
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAddressVersion:   network.IPVersionIPv4,
				PublicIPAllocationMethod: network.IPAllocationMethodStatic,
			},
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(183455)
		err = errors.Wrapf(err, "creating IP %s", name)
		return
	} else {
		__antithesis_instrumentation__.Notify(183456)
	}
	__antithesis_instrumentation__.Notify(183448)
	err = future.WaitForCompletionRef(ctx, ipc.Client)
	if err != nil {
		__antithesis_instrumentation__.Notify(183457)
		err = errors.Wrapf(err, "creating IP %s", name)
		return
	} else {
		__antithesis_instrumentation__.Notify(183458)
	}
	__antithesis_instrumentation__.Notify(183449)
	if ip, err = future.Result(ipc); err == nil {
		__antithesis_instrumentation__.Notify(183459)
		log.Infof(context.Background(), "created Azure IP %s", *ip.Name)
	} else {
		__antithesis_instrumentation__.Notify(183460)
		err = errors.Wrapf(err, "creating IP %s", name)
	}
	__antithesis_instrumentation__.Notify(183450)

	return
}

func (p *Provider) fillNetworkDetails(ctx context.Context, m *vm.VM, nicID azureID) error {
	__antithesis_instrumentation__.Notify(183461)
	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183471)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183472)
	}
	__antithesis_instrumentation__.Notify(183462)

	nicClient := network.NewInterfacesClient(*sub.SubscriptionID)
	if nicClient.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183473)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183474)
	}
	__antithesis_instrumentation__.Notify(183463)

	ipClient := network.NewPublicIPAddressesClient(*sub.SubscriptionID)
	ipClient.Authorizer = nicClient.Authorizer

	iface, err := nicClient.Get(ctx, nicID.resourceGroup, nicID.resourceName, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(183475)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183476)
	}
	__antithesis_instrumentation__.Notify(183464)
	if iface.IPConfigurations == nil {
		__antithesis_instrumentation__.Notify(183477)
		return vm.ErrBadNetwork
	} else {
		__antithesis_instrumentation__.Notify(183478)
	}
	__antithesis_instrumentation__.Notify(183465)
	cfg := (*iface.IPConfigurations)[0]
	if cfg.PrivateIPAddress == nil {
		__antithesis_instrumentation__.Notify(183479)
		return vm.ErrBadNetwork
	} else {
		__antithesis_instrumentation__.Notify(183480)
	}
	__antithesis_instrumentation__.Notify(183466)
	m.PrivateIP = *cfg.PrivateIPAddress
	m.DNS = m.PrivateIP
	if cfg.PublicIPAddress == nil || func() bool {
		__antithesis_instrumentation__.Notify(183481)
		return cfg.PublicIPAddress.ID == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(183482)
		return vm.ErrBadNetwork
	} else {
		__antithesis_instrumentation__.Notify(183483)
	}
	__antithesis_instrumentation__.Notify(183467)
	ipID, err := parseAzureID(*cfg.PublicIPAddress.ID)
	if err != nil {
		__antithesis_instrumentation__.Notify(183484)
		return vm.ErrBadNetwork
	} else {
		__antithesis_instrumentation__.Notify(183485)
	}
	__antithesis_instrumentation__.Notify(183468)

	ip, err := ipClient.Get(ctx, ipID.resourceGroup, ipID.resourceName, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(183486)
		return vm.ErrBadNetwork
	} else {
		__antithesis_instrumentation__.Notify(183487)
	}
	__antithesis_instrumentation__.Notify(183469)
	if ip.IPAddress == nil {
		__antithesis_instrumentation__.Notify(183488)
		return vm.ErrBadNetwork
	} else {
		__antithesis_instrumentation__.Notify(183489)
	}
	__antithesis_instrumentation__.Notify(183470)
	m.PublicIP = *ip.IPAddress

	return nil
}

func (p *Provider) getOrCreateResourceGroup(
	ctx context.Context, name string, location string, tags map[string]*string,
) (resources.Group, error) {
	__antithesis_instrumentation__.Notify(183490)

	p.mu.Lock()
	group, ok := p.mu.resourceGroups[name]
	p.mu.Unlock()
	if ok {
		__antithesis_instrumentation__.Notify(183498)
		return group, nil
	} else {
		__antithesis_instrumentation__.Notify(183499)
	}
	__antithesis_instrumentation__.Notify(183491)

	cacheAndReturn := func(group resources.Group) (resources.Group, error) {
		__antithesis_instrumentation__.Notify(183500)
		p.mu.Lock()
		p.mu.resourceGroups[name] = group
		p.mu.Unlock()
		return group, nil
	}
	__antithesis_instrumentation__.Notify(183492)

	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183501)
		return resources.Group{}, err
	} else {
		__antithesis_instrumentation__.Notify(183502)
	}
	__antithesis_instrumentation__.Notify(183493)

	client := resources.NewGroupsClient(*sub.SubscriptionID)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183503)
		return resources.Group{}, err
	} else {
		__antithesis_instrumentation__.Notify(183504)
	}
	__antithesis_instrumentation__.Notify(183494)

	group, err = client.Get(ctx, name)
	if err == nil {
		__antithesis_instrumentation__.Notify(183505)
		return cacheAndReturn(group)
	} else {
		__antithesis_instrumentation__.Notify(183506)
	}
	__antithesis_instrumentation__.Notify(183495)
	var detail autorest.DetailedError
	if errors.As(err, &detail) {
		__antithesis_instrumentation__.Notify(183507)

		if code, ok := detail.StatusCode.(int); ok && func() bool {
			__antithesis_instrumentation__.Notify(183508)
			return code != 404 == true
		}() == true {
			__antithesis_instrumentation__.Notify(183509)
			return resources.Group{}, err
		} else {
			__antithesis_instrumentation__.Notify(183510)
		}
	} else {
		__antithesis_instrumentation__.Notify(183511)
	}
	__antithesis_instrumentation__.Notify(183496)

	group, err = client.CreateOrUpdate(ctx, name,
		resources.Group{
			Location: to.StringPtr(location),
			Tags:     tags,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(183512)
		return resources.Group{}, err
	} else {
		__antithesis_instrumentation__.Notify(183513)
	}
	__antithesis_instrumentation__.Notify(183497)
	return cacheAndReturn(group)
}

func (p *Provider) createUltraDisk(
	ctx context.Context, group resources.Group, name string, providerOpts ProviderOpts,
) (compute.Disk, error) {
	__antithesis_instrumentation__.Notify(183514)
	sub, err := p.getSubscription(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183520)
		return compute.Disk{}, err
	} else {
		__antithesis_instrumentation__.Notify(183521)
	}
	__antithesis_instrumentation__.Notify(183515)

	client := compute.NewDisksClient(*sub.SubscriptionID)
	if client.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183522)
		return compute.Disk{}, err
	} else {
		__antithesis_instrumentation__.Notify(183523)
	}
	__antithesis_instrumentation__.Notify(183516)

	future, err := client.CreateOrUpdate(ctx, *group.Name, name,
		compute.Disk{
			Zones:    to.StringSlicePtr([]string{providerOpts.Zone}),
			Location: group.Location,
			Sku: &compute.DiskSku{
				Name: compute.DiskStorageAccountTypesUltraSSDLRS,
			},
			DiskProperties: &compute.DiskProperties{
				CreationData: &compute.CreationData{
					CreateOption: compute.DiskCreateOptionEmpty,
				},
				DiskSizeGB:        to.Int32Ptr(providerOpts.NetworkDiskSize),
				DiskIOPSReadWrite: to.Int64Ptr(providerOpts.UltraDiskIOPS),
			},
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(183524)
		return compute.Disk{}, err
	} else {
		__antithesis_instrumentation__.Notify(183525)
	}
	__antithesis_instrumentation__.Notify(183517)
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		__antithesis_instrumentation__.Notify(183526)
		return compute.Disk{}, err
	} else {
		__antithesis_instrumentation__.Notify(183527)
	}
	__antithesis_instrumentation__.Notify(183518)
	disk, err := future.Result(client)
	if err != nil {
		__antithesis_instrumentation__.Notify(183528)
		return compute.Disk{}, err
	} else {
		__antithesis_instrumentation__.Notify(183529)
	}
	__antithesis_instrumentation__.Notify(183519)
	log.Infof(context.Background(), "created ultra-disk: %s\n", *disk.Name)
	return disk, err
}

func (p *Provider) getSubscription(
	ctx context.Context,
) (sub subscriptions.Subscription, err error) {
	__antithesis_instrumentation__.Notify(183530)
	p.mu.Lock()
	sub = p.mu.subscription
	p.mu.Unlock()

	if sub.SubscriptionID != nil {
		__antithesis_instrumentation__.Notify(183534)
		return
	} else {
		__antithesis_instrumentation__.Notify(183535)
	}
	__antithesis_instrumentation__.Notify(183531)

	sc := subscriptions.NewClient()
	if sc.Authorizer, err = p.getAuthorizer(); err != nil {
		__antithesis_instrumentation__.Notify(183536)
		return
	} else {
		__antithesis_instrumentation__.Notify(183537)
	}
	__antithesis_instrumentation__.Notify(183532)

	page, err := sc.List(ctx)
	if err == nil {
		__antithesis_instrumentation__.Notify(183538)
		if len(page.Values()) == 0 {
			__antithesis_instrumentation__.Notify(183540)
			err = errors.New("did not find Azure subscription")
			return sub, err
		} else {
			__antithesis_instrumentation__.Notify(183541)
		}
		__antithesis_instrumentation__.Notify(183539)
		sub = page.Values()[0]

		p.mu.Lock()
		p.mu.subscription = page.Values()[0]
		p.mu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(183542)
	}
	__antithesis_instrumentation__.Notify(183533)
	return
}
