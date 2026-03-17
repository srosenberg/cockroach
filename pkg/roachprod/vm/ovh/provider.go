// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package ovh

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/flagstub"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/keypairs"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	"golang.org/x/sync/errgroup"
)

const (
	ProviderName = "ovh"

	defaultMachineType = "b3-64"
	defaultImage       = "Ubuntu 22.04"
	defaultRemoteUser  = "ubuntu"
	defaultCPUArch     = vm.ArchAMD64

	// OVH public network name.
	publicNetworkName = "Ext-Net"

	// createTimeout is the maximum time to wait for a server to reach ACTIVE
	// status. Bare metal provisioning can take 5–15 minutes.
	createTimeout = 10 * time.Minute

	// pollInterval is how often we check server status during creation.
	pollInterval = 10 * time.Second
)

var (
	defaultZones = []string{"GRA11"}

	ErrMissingAuth = fmt.Errorf(
		"OS_AUTH_URL environment variable not set; OVH provider requires OpenStack credentials",
	)
)

// providerInstance is the global instance of the OVH provider used by the
// roachprod CLI.
var providerInstance = &Provider{}

// Init initializes the OVH provider instance for the roachprod CLI.
func Init() (err error) {
	if os.Getenv("OS_AUTH_URL") == "" {
		vm.Providers[ProviderName] = flagstub.New(&Provider{}, ErrMissingAuth.Error())
		return nil
	}

	providerInstance, err = NewProvider()
	if err != nil {
		fmt.Printf("failed to create OVH provider: %v\n", err)
		vm.Providers[ProviderName] = flagstub.New(&Provider{}, err.Error())
		return nil
	}

	vm.Providers[ProviderName] = providerInstance
	return nil
}

// Provider implements the vm.Provider interface for OVH Metal Instances via
// the OpenStack Compute API.
type Provider struct {
	computeClient *gophercloud.ServiceClient
	imageClient   *gophercloud.ServiceClient
	region        string
	projectID     string
}

// NewProvider authenticates via OpenStack env vars and creates a compute
// service client.
func NewProvider() (*Provider, error) {
	ctx := context.Background()

	authOpts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, errors.Wrap(err, "reading OpenStack auth from environment")
	}

	providerClient, err := openstack.AuthenticatedClient(ctx, authOpts)
	if err != nil {
		return nil, errors.Wrap(err, "authenticating with OpenStack")
	}

	region := os.Getenv("OS_REGION_NAME")
	eo := gophercloud.EndpointOpts{Region: region}

	computeClient, err := openstack.NewComputeV2(providerClient, eo)
	if err != nil {
		return nil, errors.Wrap(err, "creating compute service client")
	}

	imageClient, err := openstack.NewImageV2(providerClient, eo)
	if err != nil {
		return nil, errors.Wrap(err, "creating image service client")
	}

	return &Provider{
		computeClient: computeClient,
		imageClient:   imageClient,
		region:        region,
		projectID:     authOpts.TenantID,
	}, nil
}

// Name is part of the vm.Provider interface.
func (p *Provider) Name() string { return ProviderName }

// String returns a human-readable identifier for the provider.
func (p *Provider) String() string {
	return fmt.Sprintf("%s-%s", ProviderName, p.region)
}

// Active is part of the vm.Provider interface.
func (p *Provider) Active() bool { return true }

// IsCentralizedProvider returns true because OVH is a remote provider.
func (p *Provider) IsCentralizedProvider() bool { return true }

// ProjectActive is part of the vm.Provider interface.
func (p *Provider) ProjectActive(project string) bool {
	return project == ""
}

// Create provisions OVH servers, waits for ACTIVE status, and returns vm.List.
func (p *Provider) Create(
	l *logger.Logger, names []string, opts vm.CreateOpts, vmProviderOpts vm.ProviderOpts,
) (vm.List, error) {
	providerOpts := vmProviderOpts.(*ProviderOpts)

	// Resolve zones for each VM.
	zones, err := p.computeZones(providerOpts, opts.GeoDistributed, len(names))
	if err != nil {
		return nil, errors.Wrap(err, "computing zones")
	}

	// Ensure SSH key is uploaded.
	if err := p.ConfigSSH(l, zones); err != nil {
		return nil, errors.Wrap(err, "configuring SSH")
	}

	ctx := context.Background()

	// Resolve image and flavor IDs.
	imageID, err := p.getImageID(ctx, providerOpts.Image)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving image %q", providerOpts.Image)
	}

	flavorID, err := p.getFlavorID(ctx, providerOpts.MachineType)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving flavor %q", providerOpts.MachineType)
	}

	keyName, err := p.sshKeyName()
	if err != nil {
		return nil, errors.Wrap(err, "computing SSH key name")
	}

	// Generate startup script.
	startupScript, err := p.startupScript(startupArgs{
		StartupArgs: vm.DefaultStartupArgs(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "generating startup script")
	}
	userData := []byte(startupScript)

	// Build metadata labels.
	labels := vm.GetDefaultLabelMap(opts)
	for k, v := range opts.CustomLabels {
		labels[k] = v
	}

	// Create servers in parallel.
	var mu syncutil.Mutex
	var g errgroup.Group
	vms := make(vm.List, len(names))

	for i, vmName := range names {
		g.Go(func() error {
			createOpts := keypairs.CreateOptsExt{
				CreateOptsBuilder: servers.CreateOpts{
					Name:      vmName,
					ImageRef:  imageID,
					FlavorRef: flavorID,
					UserData:  userData,
					Metadata:  labels,
				},
				KeyName: keyName,
			}

			server, err := servers.Create(ctx, p.computeClient, createOpts, nil).Extract()
			if err != nil {
				return errors.Wrapf(err, "creating server %s", vmName)
			}

			l.Printf("created server %s (id=%s), waiting for ACTIVE...", vmName, server.ID)

			// Poll until the server reaches ACTIVE status.
			server, err = p.waitForActive(ctx, server.ID)
			if err != nil {
				return errors.Wrapf(err, "waiting for server %s to become ACTIVE", vmName)
			}

			v := p.serverToVM(server, zones[i], nil)

			mu.Lock()
			vms[i] = v
			mu.Unlock()

			l.Printf("server %s is ACTIVE (ip=%s)", vmName, v.PublicIP)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return vms, nil
}

// List queries OVH for all roachprod-tagged servers.
func (p *Provider) List(
	ctx context.Context, l *logger.Logger, opts vm.ListOptions,
) (vm.List, error) {
	allPages, err := servers.List(p.computeClient, nil).AllPages(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "listing servers")
	}

	allServers, err := servers.ExtractServers(allPages)
	if err != nil {
		return nil, errors.Wrap(err, "extracting servers")
	}

	// Build a flavor ID-to-name map so we can display human-readable machine
	// types instead of UUIDs.
	flavorNames, err := p.getFlavorNames(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "resolving flavor names")
	}

	var ret vm.List
	for i := range allServers {
		s := &allServers[i]
		if s.Metadata[vm.TagRoachprod] != "true" {
			continue
		}
		ret = append(ret, p.serverToVM(s, p.region, flavorNames))
	}

	return ret, nil
}

// Delete destroys the specified VMs.
func (p *Provider) Delete(l *logger.Logger, vms vm.List) error {
	ctx := context.Background()
	var g errgroup.Group

	for _, v := range vms {
		g.Go(func() error {
			result := servers.Delete(ctx, p.computeClient, v.ProviderID)
			if err := result.ExtractErr(); err != nil {
				return errors.Wrapf(err, "deleting server %s (id=%s)", v.Name, v.ProviderID)
			}
			l.Printf("deleted server %s (id=%s)", v.Name, v.ProviderID)
			return nil
		})
	}

	return g.Wait()
}

// AddLabels updates server metadata with the given labels.
func (p *Provider) AddLabels(l *logger.Logger, vms vm.List, labels map[string]string) error {
	ctx := context.Background()
	var g errgroup.Group

	for _, v := range vms {
		g.Go(func() error {
			opts := servers.MetadataOpts(labels)
			_, err := servers.UpdateMetadata(ctx, p.computeClient, v.ProviderID, opts).Extract()
			if err != nil {
				return errors.Wrapf(err, "updating metadata for server %s", v.Name)
			}
			return nil
		})
	}

	return g.Wait()
}

// RemoveLabels is part of the vm.Provider interface.
func (p *Provider) RemoveLabels(l *logger.Logger, vms vm.List, labels []string) error {
	ctx := context.Background()
	var g errgroup.Group

	for _, v := range vms {
		g.Go(func() error {
			for _, key := range labels {
				result := servers.DeleteMetadatum(ctx, p.computeClient, v.ProviderID, key)
				if err := result.ExtractErr(); err != nil {
					// Ignore errors for keys that don't exist.
					l.Printf("WARN: failed to remove label %q from server %s: %v", key, v.Name, err)
				}
			}
			return nil
		})
	}

	return g.Wait()
}

// Extend is part of the vm.Provider interface.
func (p *Provider) Extend(l *logger.Logger, vms vm.List, lifetime time.Duration) error {
	return p.AddLabels(l, vms, map[string]string{
		vm.TagLifetime: lifetime.String(),
	})
}

// ConfigSSH uploads the user's SSH public key as an OpenStack keypair if not
// already present.
func (p *Provider) ConfigSSH(l *logger.Logger, zones []string) error {
	ctx := context.Background()

	keyName, err := p.sshKeyName()
	if err != nil {
		return err
	}

	// Check if the keypair already exists.
	_, err = keypairs.Get(ctx, p.computeClient, keyName, nil).Extract()
	if err == nil {
		return nil // already exists
	}

	pubKey, err := config.SSHPublicKey()
	if err != nil {
		return errors.Wrap(err, "reading SSH public key")
	}

	_, err = keypairs.Create(ctx, p.computeClient, keypairs.CreateOpts{
		Name:      keyName,
		PublicKey: strings.TrimSpace(pubKey),
	}).Extract()
	if err != nil {
		return errors.Wrapf(err, "uploading SSH keypair %q", keyName)
	}

	l.Printf("uploaded SSH keypair %q to OVH", keyName)
	return nil
}

// FindActiveAccount is part of the vm.Provider interface.
func (p *Provider) FindActiveAccount(l *logger.Logger) (string, error) {
	username := os.Getenv("OS_USERNAME")
	if username == "" {
		return "", errors.New("OS_USERNAME not set")
	}
	return username, nil
}

//
// Stub/no-op methods
//

// CleanSSH is part of the vm.Provider interface.
func (p *Provider) CleanSSH(l *logger.Logger) error { return nil }

// Reset is part of the vm.Provider interface.
func (p *Provider) Reset(l *logger.Logger, vms vm.List) error { return nil }

// Grow is part of the vm.Provider interface.
func (p *Provider) Grow(
	l *logger.Logger, vms vm.List, clusterName string, names []string,
) (vm.List, error) {
	return nil, vm.UnimplementedError
}

// Shrink is part of the vm.Provider interface.
func (p *Provider) Shrink(l *logger.Logger, vmsToRemove vm.List, clusterName string) error {
	return vm.UnimplementedError
}

// SupportsSpotVMs is part of the vm.Provider interface.
func (p *Provider) SupportsSpotVMs() bool { return false }

// GetPreemptedSpotVMs is part of the vm.Provider interface.
func (p *Provider) GetPreemptedSpotVMs(
	l *logger.Logger, vms vm.List, since time.Time,
) ([]vm.PreemptedVM, error) {
	return nil, nil
}

// GetHostErrorVMs is part of the vm.Provider interface.
func (p *Provider) GetHostErrorVMs(
	l *logger.Logger, vms vm.List, since time.Time,
) ([]string, error) {
	return nil, nil
}

// GetLiveMigrationVMs is part of the vm.Provider interface.
func (p *Provider) GetLiveMigrationVMs(
	l *logger.Logger, vms vm.List, since time.Time,
) ([]string, error) {
	return nil, nil
}

// GetVMSpecs is part of the vm.Provider interface.
func (p *Provider) GetVMSpecs(
	l *logger.Logger, vms vm.List,
) (map[string]map[string]interface{}, error) {
	return nil, nil
}

// CreateVolume is part of the vm.Provider interface.
func (p *Provider) CreateVolume(l *logger.Logger, vco vm.VolumeCreateOpts) (vm.Volume, error) {
	return vm.Volume{}, vm.UnimplementedError
}

// ListVolumes is part of the vm.Provider interface.
func (p *Provider) ListVolumes(l *logger.Logger, v *vm.VM) ([]vm.Volume, error) {
	return nil, vm.UnimplementedError
}

// DeleteVolume is part of the vm.Provider interface.
func (p *Provider) DeleteVolume(l *logger.Logger, volume vm.Volume, v *vm.VM) error {
	return vm.UnimplementedError
}

// AttachVolume is part of the vm.Provider interface.
func (p *Provider) AttachVolume(l *logger.Logger, volume vm.Volume, v *vm.VM) (string, error) {
	return "", vm.UnimplementedError
}

// CreateVolumeSnapshot is part of the vm.Provider interface.
func (p *Provider) CreateVolumeSnapshot(
	l *logger.Logger, volume vm.Volume, vsco vm.VolumeSnapshotCreateOpts,
) (vm.VolumeSnapshot, error) {
	return vm.VolumeSnapshot{}, vm.UnimplementedError
}

// ListVolumeSnapshots is part of the vm.Provider interface.
func (p *Provider) ListVolumeSnapshots(
	l *logger.Logger, vslo vm.VolumeSnapshotListOpts,
) ([]vm.VolumeSnapshot, error) {
	return nil, vm.UnimplementedError
}

// DeleteVolumeSnapshots is part of the vm.Provider interface.
func (p *Provider) DeleteVolumeSnapshots(l *logger.Logger, snapshots ...vm.VolumeSnapshot) error {
	return vm.UnimplementedError
}

// CreateLoadBalancer is part of the vm.Provider interface.
func (p *Provider) CreateLoadBalancer(l *logger.Logger, vms vm.List, port int) error {
	return vm.UnimplementedError
}

// DeleteLoadBalancer is part of the vm.Provider interface.
func (p *Provider) DeleteLoadBalancer(l *logger.Logger, vms vm.List, port int) error {
	return vm.UnimplementedError
}

// ListLoadBalancers is part of the vm.Provider interface.
func (p *Provider) ListLoadBalancers(l *logger.Logger, vms vm.List) ([]vm.ServiceAddress, error) {
	return nil, nil
}

//
// Private helpers
//

// sshKeyName returns a deterministic name for the user's SSH keypair:
// <OS_USERNAME>-<sha256(pubkey)[:12]>.
func (p *Provider) sshKeyName() (string, error) {
	pubKey, err := config.SSHPublicKey()
	if err != nil {
		return "", err
	}
	h := sha256.Sum256([]byte(strings.TrimSpace(pubKey)))
	hash := fmt.Sprintf("%x", h)[:12]

	username := os.Getenv("OS_USERNAME")
	if username == "" {
		username = "unknown"
	}
	return fmt.Sprintf("%s-%s", username, hash), nil
}

// getImageID looks up an image UUID by name using the Glance Image Service.
func (p *Provider) getImageID(ctx context.Context, name string) (string, error) {
	listOpts := images.ListOpts{Name: name}
	allPages, err := images.List(p.imageClient, listOpts).AllPages(ctx)
	if err != nil {
		return "", errors.Wrap(err, "listing images")
	}

	allImages, err := images.ExtractImages(allPages)
	if err != nil {
		return "", errors.Wrap(err, "extracting images")
	}

	for _, img := range allImages {
		if img.Name == name {
			return img.ID, nil
		}
	}

	return "", errors.Newf("image %q not found", name)
}

// getFlavorID looks up a flavor UUID by name.
func (p *Provider) getFlavorID(ctx context.Context, name string) (string, error) {
	allPages, err := flavors.ListDetail(p.computeClient, nil).AllPages(ctx)
	if err != nil {
		return "", errors.Wrap(err, "listing flavors")
	}

	allFlavors, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		return "", errors.Wrap(err, "extracting flavors")
	}

	for _, f := range allFlavors {
		if f.Name == name {
			return f.ID, nil
		}
	}

	return "", errors.Newf("flavor %q not found", name)
}

// waitForActive polls until the server reaches ACTIVE status or times out.
func (p *Provider) waitForActive(ctx context.Context, serverID string) (*servers.Server, error) {
	deadline := timeutil.Now().Add(createTimeout)
	for {
		server, err := servers.Get(ctx, p.computeClient, serverID).Extract()
		if err != nil {
			return nil, errors.Wrapf(err, "getting server %s", serverID)
		}

		switch server.Status {
		case "ACTIVE":
			return server, nil
		case "ERROR":
			msg := "unknown error"
			if server.Fault.Message != "" {
				msg = server.Fault.Message
			}
			return nil, errors.Newf("server %s entered ERROR state: %s", serverID, msg)
		}

		if timeutil.Now().After(deadline) {
			return nil, errors.Newf(
				"server %s did not reach ACTIVE within %v (current: %s)",
				serverID, createTimeout, server.Status,
			)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

// serverToVM converts an OpenStack Server to a roachprod vm.VM.
// flavorNames maps flavor IDs to human-readable names; it may be nil, in which
// case the flavor ID is used as a fallback.
func (p *Provider) serverToVM(s *servers.Server, zone string, flavorNames map[string]string) vm.VM {
	publicIP := extractPublicIP(s.Addresses)

	var lifetime time.Duration
	if lt, ok := s.Metadata[vm.TagLifetime]; ok {
		lifetime, _ = time.ParseDuration(lt)
	}

	return vm.VM{
		Name:              s.Name,
		CreatedAt:         s.Created,
		Lifetime:          lifetime,
		Labels:            s.Metadata,
		Provider:          ProviderName,
		ProviderID:        s.ID,
		ProviderAccountID: p.projectID,
		PublicIP:          publicIP,
		PrivateIP:         publicIP, // use public IP as fallback
		RemoteUser:        defaultRemoteUser,
		VPC:               fmt.Sprintf("%s-%s", ProviderName, p.region),
		MachineType:       flavorName(s.Flavor, flavorNames),
		CPUArch:           defaultCPUArch,
		Zone:              zone,
	}
}

// flavorName resolves the human-readable flavor name from the Server.Flavor
// map. It first checks the "original_name" key (available in newer API
// microversions), then looks up the flavor ID in the provided name map, and
// falls back to the raw flavor ID.
func flavorName(flavor map[string]interface{}, namesByID map[string]string) string {
	if name, ok := flavor["original_name"].(string); ok && name != "" {
		return name
	}
	id, _ := flavor["id"].(string)
	if id != "" && namesByID != nil {
		if name, ok := namesByID[id]; ok {
			return name
		}
	}
	return id
}

// getFlavorNames returns a map from flavor ID to flavor name.
func (p *Provider) getFlavorNames(ctx context.Context) (map[string]string, error) {
	allPages, err := flavors.ListDetail(p.computeClient, nil).AllPages(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "listing flavors")
	}

	allFlavors, err := flavors.ExtractFlavors(allPages)
	if err != nil {
		return nil, errors.Wrap(err, "extracting flavors")
	}

	names := make(map[string]string, len(allFlavors))
	for _, f := range allFlavors {
		names[f.ID] = f.Name
	}
	return names, nil
}

// extractPublicIP parses the IPv4 address from the server's Addresses map.
// OVH typically uses "Ext-Net" as the public network name.
func extractPublicIP(addresses map[string]interface{}) string {
	for _, netName := range []string{publicNetworkName, "public"} {
		network, ok := addresses[netName]
		if !ok {
			continue
		}
		addrs, ok := network.([]interface{})
		if !ok {
			continue
		}
		for _, a := range addrs {
			addr, ok := a.(map[string]interface{})
			if !ok {
				continue
			}
			version, _ := addr["version"].(float64)
			ip, _ := addr["addr"].(string)
			if version == 4 && ip != "" {
				return ip
			}
		}
	}
	return ""
}

// computeZones resolves the zones for each VM based on provider options.
func (p *Provider) computeZones(
	providerOpts *ProviderOpts, geoDistributed bool, numNodes int,
) ([]string, error) {
	zoneFlag := providerOpts.CreateZones
	if len(zoneFlag) == 0 {
		zoneFlag = defaultZones
	}

	expandedZones, err := vm.ExpandZonesFlag(zoneFlag)
	if err != nil {
		return nil, err
	}

	// The OVH provider currently supports a single region, so zones map 1:1
	// to the region. Distribute nodes across the supplied zones.
	placement := vm.ZonePlacement(len(expandedZones), numNodes)
	zones := make([]string, numNodes)
	for i, zi := range placement {
		zones[i] = expandedZones[zi]
	}

	return zones, nil
}
