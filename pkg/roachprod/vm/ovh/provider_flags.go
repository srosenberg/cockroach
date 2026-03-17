// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package ovh

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/spf13/pflag"
)

// ProviderOpts provides user-configurable, OVH-specific create options.
type ProviderOpts struct {
	// MachineType is the OVH flavor name (e.g. "b3-64" for bare metal).
	MachineType string

	// Image is the OS image name (e.g. "Ubuntu 22.04").
	Image string

	// CreateZones stores the list of OVH regions for cluster creation.
	CreateZones []string
}

// DefaultProviderOpts returns the default OVH provider options.
func DefaultProviderOpts() *ProviderOpts {
	return &ProviderOpts{
		MachineType: defaultMachineType,
		Image:       defaultImage,
	}
}

// ConfigureCreateFlags is part of the vm.ProviderOpts interface.
func (o *ProviderOpts) ConfigureCreateFlags(flags *pflag.FlagSet) {
	flags.StringVar(
		&o.MachineType,
		ProviderName+"-machine-type",
		o.MachineType,
		"OVH flavor name (see OVH console for available flavors)",
	)

	flags.StringVar(
		&o.Image,
		ProviderName+"-image",
		o.Image,
		"OVH OS image name",
	)

	flags.StringSliceVar(
		&o.CreateZones,
		ProviderName+"-zones",
		o.CreateZones,
		fmt.Sprintf(
			`OVH regions to use for cluster creation.
If zones are formatted as AZ:N where N is an integer, the zone will be repeated N times.
If > 1 zone specified, the cluster will be spread out evenly by zone regardless of geo.
(default [%s])`,
			strings.Join(defaultZones, ","),
		),
	)
}

// ConfigureProviderFlags is part of the vm.Provider interface.
func (p *Provider) ConfigureProviderFlags(flags *pflag.FlagSet, _ vm.MultipleProjectsOption) {}

// ConfigureClusterCleanupFlags is part of the vm.Provider interface.
func (p *Provider) ConfigureClusterCleanupFlags(flags *pflag.FlagSet) {}

// CreateProviderOpts is part of the vm.Provider interface.
func (p *Provider) CreateProviderOpts() vm.ProviderOpts {
	return DefaultProviderOpts()
}
