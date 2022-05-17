package azure

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"text/template"
)

type azureStartupArgs struct {
	RemoteUser      string
	AttachedDiskLun *int
}

const azureStartupTemplate = `#!/bin/bash

# Script for setting up a Azure machine for roachprod use.
set -xe
mount_opts="defaults"

{{if .AttachedDiskLun}}
# Setup network attached storage
devices=("/dev/disk/azure/scsi1/lun{{.AttachedDiskLun}}")
{{else}}
# Setup local storage.
devices=($(realpath -qe /dev/disk/by-id/nvme-* | sort -u))
{{end}}

if (( ${#devices[@]} == 0 ));
then
  # Use /mnt directly.
  echo "No attached or NVME disks found, creating /mnt/data1"
  mkdir -p /mnt/data1
  chown {{.RemoteUser}} /mnt/data1
else
  for d in "${!devices[@]}"; do
    disk=${devices[$d]}
    mount="/data$((d+1))"
    sudo mkdir -p "${mount}"
    sudo mkfs.ext4 -F "${disk}"
    sudo mount -o "${mount_opts}" "${disk}" "${mount}"
    echo "${disk} ${mount} ext4 ${mount_opts} 1 1" | sudo tee -a /etc/fstab
    ln -s "${mount}" "/mnt/$(basename $mount)"
  done
  chown {{.RemoteUser}} /data*
fi

# increase the number of concurrent unauthenticated connections to the sshd
# daemon. See https://en.wikibooks.org/wiki/OpenSSH/Cookbook/Load_Balancing.
# By default, only 10 unauthenticated connections are permitted before sshd
# starts randomly dropping connections.
sh -c 'echo "MaxStartups 64:30:128" >> /etc/ssh/sshd_config'
# Crank up the logging for issues such as:
# https://github.com/cockroachdb/cockroach/issues/36929
sed -i'' 's/LogLevel.*$/LogLevel DEBUG3/' /etc/ssh/sshd_config
service sshd restart
# increase the default maximum number of open file descriptors for
# root and non-root users. Load generators running a lot of concurrent
# workers bump into this often.
sh -c 'echo "root - nofile 1048576\n* - nofile 1048576" > /etc/security/limits.d/10-roachprod-nofiles.conf'

# Send TCP keepalives every minute since GCE will terminate idle connections
# after 10m. Note that keepalives still need to be requested by the application
# with the SO_KEEPALIVE socket option.
cat <<EOF > /etc/sysctl.d/99-roachprod-tcp-keepalive.conf
net.ipv4.tcp_keepalive_time=60
net.ipv4.tcp_keepalive_intvl=60
net.ipv4.tcp_keepalive_probes=5
EOF

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
touch /mnt/data1/.roachprod-initialized
`

func evalStartupTemplate(args azureStartupArgs) (string, error) {
	__antithesis_instrumentation__.Notify(183552)
	cloudInit := bytes.NewBuffer(nil)
	encoder := base64.NewEncoder(base64.StdEncoding, cloudInit)
	gz := gzip.NewWriter(encoder)
	t := template.Must(template.New("start").Parse(azureStartupTemplate))
	if err := t.Execute(gz, args); err != nil {
		__antithesis_instrumentation__.Notify(183556)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183557)
	}
	__antithesis_instrumentation__.Notify(183553)
	if err := gz.Close(); err != nil {
		__antithesis_instrumentation__.Notify(183558)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183559)
	}
	__antithesis_instrumentation__.Notify(183554)
	if err := encoder.Close(); err != nil {
		__antithesis_instrumentation__.Notify(183560)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183561)
	}
	__antithesis_instrumentation__.Notify(183555)
	return cloudInit.String(), nil
}
