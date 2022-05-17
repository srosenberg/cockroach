package aws

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"strings"
	"text/template"

	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

const awsStartupScriptTemplate = `#!/usr/bin/env bash
# Script for setting up a AWS machine for roachprod use.

set -x
sudo apt-get update
sudo apt-get install -qy --no-install-recommends mdadm

mount_opts="defaults"
{{if .ExtraMountOpts}}mount_opts="${mount_opts},{{.ExtraMountOpts}}"{{end}}

use_multiple_disks='{{if .UseMultipleDisks}}true{{end}}'

disks=()
mount_prefix="/mnt/data"

# On different machine types, the drives are either called nvme... or xvdd.
for d in $(ls /dev/nvme?n1 /dev/xvdd); do
  if ! mount | grep ${d}; then
    disks+=("${d}")
    echo "Disk ${d} not mounted, need to mount..."
  else
    echo "Disk ${d} already mounted, skipping..."
  fi
done


if [ "${#disks[@]}" -eq "0" ]; then
  mountpoint="${mount_prefix}1"
  echo "No disks mounted, creating ${mountpoint}"
  mkdir -p ${mountpoint}
  chmod 777 ${mountpoint}
elif [ "${#disks[@]}" -eq "1" ] || [ -n "$use_multiple_disks" ]; then
  disknum=1
  for disk in "${disks[@]}"
  do
    mountpoint="${mount_prefix}${disknum}"
    disknum=$((disknum + 1 ))
    echo "Mounting ${disk} at ${mountpoint}"
    mkdir -p ${mountpoint}
    mkfs.ext4 -F ${disk}
    mount -o ${mount_opts} ${disk} ${mountpoint}
    chmod 777 ${mountpoint}
    echo "${disk} ${mountpoint} ext4 ${mount_opts} 1 1" | tee -a /etc/fstab
  done
else
  mountpoint="${mount_prefix}1"
  echo "${#disks[@]} disks mounted, creating ${mountpoint} using RAID 0"
  mkdir -p ${mountpoint}
  raiddisk="/dev/md0"
  mdadm --create ${raiddisk} --level=0 --raid-devices=${#disks[@]} "${disks[@]}"
  mkfs.ext4 -F ${raiddisk}
  mount -o ${mount_opts} ${raiddisk} ${mountpoint}
  chmod 777 ${mountpoint}
  echo "${raiddisk} ${mountpoint} ext4 ${mount_opts} 1 1" | tee -a /etc/fstab
fi

sudo apt-get install -qy chrony

# Override the chrony config. In particular,
# log aggressively when clock is adjusted (0.01s)
# and exclusively use a single time server.
sudo cat <<EOF > /etc/chrony/chrony.conf
keyfile /etc/chrony/chrony.keys
commandkey 1
driftfile /var/lib/chrony/chrony.drift
log tracking measurements statistics
logdir /var/log/chrony
maxupdateskew 100.0
dumponexit
dumpdir /var/lib/chrony
logchange 0.01
hwclockfile /etc/adjtime
rtcsync
server 169.254.169.123 prefer iburst
makestep 0.1 3
EOF

sudo /etc/init.d/chrony restart
sudo chronyc -a waitsync 30 0.01 | sudo tee -a /root/chrony.log

# sshguard can prevent frequent ssh connections to the same host. Disable it.
sudo service sshguard stop
# increase the number of concurrent unauthenticated connections to the sshd
# daemon. See https://en.wikibooks.org/wiki/OpenSSH/Cookbook/Load_Balancing.
# By default, only 10 unauthenticated connections are permitted before sshd
# starts randomly dropping connections.
sudo sh -c 'echo "MaxStartups 64:30:128" >> /etc/ssh/sshd_config'
# Crank up the logging for issues such as:
# https://github.com/cockroachdb/cockroach/issues/36929
sudo sed -i'' 's/LogLevel.*$/LogLevel DEBUG3/' /etc/ssh/sshd_config
sudo service sshd restart
# increase the default maximum number of open file descriptors for
# root and non-root users. Load generators running a lot of concurrent
# workers bump into this often.
sudo sh -c 'echo "root - nofile 1048576\n* - nofile 1048576" > /etc/security/limits.d/10-roachprod-nofiles.conf'

# Enable core dumps
cat <<EOF > /etc/security/limits.d/core_unlimited.conf
* soft core unlimited
* hard core unlimited
root soft core unlimited
root hard core unlimited
EOF

mkdir -p /mnt/data1/cores
chmod a+w /mnt/data1/cores
CORE_PATTERN="/mnt/data1/cores/core.%e.%p.%h.%t"
echo "$CORE_PATTERN" > /proc/sys/kernel/core_pattern
sed -i'~' 's/enabled=1/enabled=0/' /etc/default/apport
sed -i'~' '/.*kernel\\.core_pattern.*/c\\' /etc/sysctl.conf
echo "kernel.core_pattern=$CORE_PATTERN" >> /etc/sysctl.conf

sysctl --system  # reload sysctl settings

sudo touch /mnt/data1/.roachprod-initialized
`

func writeStartupScript(extraMountOpts string, useMultiple bool) (string, error) {
	__antithesis_instrumentation__.Notify(183033)
	type tmplParams struct {
		ExtraMountOpts   string
		UseMultipleDisks bool
	}

	args := tmplParams{ExtraMountOpts: extraMountOpts, UseMultipleDisks: useMultiple}

	tmpfile, err := ioutil.TempFile("", "aws-startup-script")
	if err != nil {
		__antithesis_instrumentation__.Notify(183036)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183037)
	}
	__antithesis_instrumentation__.Notify(183034)
	defer tmpfile.Close()

	t := template.Must(template.New("start").Parse(awsStartupScriptTemplate))
	if err := t.Execute(tmpfile, args); err != nil {
		__antithesis_instrumentation__.Notify(183038)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183039)
	}
	__antithesis_instrumentation__.Notify(183035)
	return tmpfile.Name(), nil
}

func (p *Provider) runCommand(args []string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(183040)

	if p.Profile != "" {
		__antithesis_instrumentation__.Notify(183043)
		args = append(args[:len(args):len(args)], "--profile", p.Profile)
	} else {
		__antithesis_instrumentation__.Notify(183044)
	}
	__antithesis_instrumentation__.Notify(183041)
	var stderrBuf bytes.Buffer
	cmd := exec.Command("aws", args...)
	cmd.Stderr = &stderrBuf
	output, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(183045)
		if exitErr := (*exec.ExitError)(nil); errors.As(err, &exitErr) {
			__antithesis_instrumentation__.Notify(183047)
			log.Infof(context.Background(), "%s", string(exitErr.Stderr))
		} else {
			__antithesis_instrumentation__.Notify(183048)
		}
		__antithesis_instrumentation__.Notify(183046)
		return nil, errors.Wrapf(err, "failed to run: aws %s: stderr: %v",
			strings.Join(args, " "), stderrBuf.String())
	} else {
		__antithesis_instrumentation__.Notify(183049)
	}
	__antithesis_instrumentation__.Notify(183042)
	return output, nil
}

func (p *Provider) runJSONCommand(args []string, parsed interface{}) error {
	__antithesis_instrumentation__.Notify(183050)

	args = append(args[:len(args):len(args)], "--output", "json")
	rawJSON, err := p.runCommand(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(183053)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183054)
	}
	__antithesis_instrumentation__.Notify(183051)
	if err := json.Unmarshal(rawJSON, &parsed); err != nil {
		__antithesis_instrumentation__.Notify(183055)
		return errors.Wrapf(err, "failed to parse json %s", rawJSON)
	} else {
		__antithesis_instrumentation__.Notify(183056)
	}
	__antithesis_instrumentation__.Notify(183052)

	return nil
}

func regionMap(vms vm.List) (map[string]vm.List, error) {
	__antithesis_instrumentation__.Notify(183057)

	byRegion := make(map[string]vm.List)
	for _, m := range vms {
		__antithesis_instrumentation__.Notify(183059)
		region, err := zoneToRegion(m.Zone)
		if err != nil {
			__antithesis_instrumentation__.Notify(183061)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(183062)
		}
		__antithesis_instrumentation__.Notify(183060)
		byRegion[region] = append(byRegion[region], m)
	}
	__antithesis_instrumentation__.Notify(183058)
	return byRegion, nil
}

func zoneToRegion(zone string) (string, error) {
	__antithesis_instrumentation__.Notify(183063)
	return zone[0 : len(zone)-1], nil
}
