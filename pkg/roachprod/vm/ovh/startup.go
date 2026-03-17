// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package ovh

import (
	"bytes"

	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
)

// startupArgs specifies template arguments for the setup template.
type startupArgs struct {
	vm.StartupArgs
}

const startupTemplate = `#!/usr/bin/env bash

# Script for setting up an OVH bare metal machine for roachprod use.

{{ template "head_utils" . }}
{{ template "apt_packages" . }}

# Provider specific disk logic
function detect_disks() {
	# OVH bare metal disk detection — find NVMe and SATA drives, skip the boot disk.
	local disks=()
	local boot_disk
	boot_disk=$(lsblk -no PKNAME $(findmnt -n -o SOURCE /) 2>/dev/null || echo "")

	for d in /dev/nvme*n1 /dev/sd?; do
		[ -e "$d" ] || continue

		# Skip the boot disk.
		local base
		base=$(basename "$d")
		if [ "$base" = "$boot_disk" ]; then
			echo "Disk ${d} is the boot disk, skipping..." >&2
			continue
		fi

		mounted="no"

{{ if eq .Filesystem "zfs" }}
		if (zpool list -v -P | grep -q ${d}) || (mount | grep -q ${d}); then
			mounted="yes"
		fi
{{ else }}
		if mount | grep -q ${d}; then
			mounted="yes"
		fi
{{ end }}

		if [ "$mounted" = "no" ]; then
			disks+=("${d}")
			echo "Disk ${d} is not mounted, need to mount..." >&2
		else
			echo "Disk ${d} is already mounted, skipping..." >&2
		fi
	done

	# Return disks array by printing each element (only if non-empty)
	if [ "${#disks[@]}" -gt 0 ]; then
		printf '%s\n' "${disks[@]}"
	fi
}

# Common disk setup logic that calls the above detect_disks function
{{ template "setup_disks_utils" . }}

{{ template "ulimits" . }}
{{ template "tcpdump" . }}
{{ template "keepalives" . }}
{{ template "cron_utils" . }}
{{ template "chrony_utils" . }}
{{ template "timers_services_utils" . }}
{{ template "core_dumps_utils" . }}
{{ template "hostname_utils" . }}
{{ template "fips_utils" . }}
{{ template "ssh_utils" . }}
{{ template "node_exporter" . }}
{{ template "ebpf_exporter" . }}

sudo touch {{ .OSInitializedFile }}
`

// startupScript returns the startup script for the given arguments.
func (p *Provider) startupScript(args startupArgs) (string, error) {
	data := bytes.NewBuffer(nil)

	err := vm.GenerateStartupScript(data, startupTemplate, args)
	if err != nil {
		return "", err
	}

	return data.String(), nil
}
