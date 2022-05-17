package install

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
)

var installCmds = map[string]string{
	"cassandra": `
echo "deb http://www.apache.org/dist/cassandra/debian 311x main" | \
	sudo tee -a /etc/apt/sources.list.d/cassandra.sources.list;
curl https://www.apache.org/dist/cassandra/KEYS | sudo apt-key add -;
sudo apt-get update;
sudo apt-get install -y cassandra;
sudo service cassandra stop;
`,

	"charybdefs": `
  thrift_dir="/opt/thrift"

  if [ ! -f "/usr/bin/thrift" ]; then
	sudo apt-get update;
	sudo apt-get install -qy automake bison flex g++ git libboost-all-dev libevent-dev libssl-dev libtool make pkg-config python-setuptools libglib2.0-dev python2 python-six

    sudo mkdir -p "${thrift_dir}"
    sudo chmod 777 "${thrift_dir}"
    cd "${thrift_dir}"
    curl "https://archive.apache.org/dist/thrift/0.13.0/thrift-0.13.0.tar.gz" | sudo tar xvz --strip-components 1
    sudo ./configure --prefix=/usr
    sudo make -j$(nproc)
    sudo make install
    (cd "${thrift_dir}/lib/py" && sudo python2 setup.py install)
  fi

  charybde_dir="/opt/charybdefs"
  nemesis_path="${charybde_dir}/charybdefs-nemesis"

  if [ ! -f "${nemesis_path}" ]; then
    sudo apt-get install -qy build-essential cmake libfuse-dev fuse
    sudo rm -rf "${charybde_dir}" "${nemesis_path}" /usr/local/bin/charybdefs{,-nemesis}
    sudo mkdir -p "${charybde_dir}"
    sudo chmod 777 "${charybde_dir}"
    # TODO(bilal): Change URL back to scylladb/charybdefs once https://github.com/scylladb/charybdefs/pull/28 is merged.
    git clone --depth 1 "https://github.com/itsbilal/charybdefs.git" "${charybde_dir}"

    cd "${charybde_dir}"
    thrift -r --gen cpp server.thrift
    cmake CMakeLists.txt
    make -j$(nproc)

    sudo modprobe fuse
    sudo ln -s "${charybde_dir}/charybdefs" /usr/local/bin/charybdefs
    cat > "${nemesis_path}" <<EOF
#!/bin/bash
cd /opt/charybdefs/cookbook
./recipes "\$@"
EOF
    chmod +x "${nemesis_path}"
	sudo ln -s "${nemesis_path}" /usr/local/bin/charybdefs-nemesis
fi
`,

	"confluent": `
sudo apt-get update;
sudo apt-get install -y default-jdk-headless;
curl https://packages.confluent.io/archive/5.0/confluent-oss-5.0.0-2.11.tar.gz | sudo tar -C /usr/local -xz;
sudo ln -s /usr/local/confluent-5.0.0 /usr/local/confluent;
`,

	"docker": `
sudo apt-get update;
sudo apt-get install  -y \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common;
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -;
sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable";

sudo apt-get update;
sudo apt-get install  -y docker-ce;
`,

	"gcc": `
sudo apt-get update;
sudo apt-get install -y gcc;
`,

	"go": `
sudo apt-get update;
sudo apt-get install -y graphviz rlwrap;

curl https://dl.google.com/go/go1.12.linux-amd64.tar.gz | sudo tar -C /usr/local -xz;
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh > /dev/null;
sudo chmod +x /etc/profile.d/go.sh;
`,

	"haproxy": `
sudo apt-get update;
sudo apt-get install -y haproxy;
`,

	"ntp": `
sudo apt-get update;
sudo apt-get install -y \
  ntp \
  ntpdate;
`,

	"sysbench": `
curl -s https://packagecloud.io/install/repositories/akopytov/sysbench/script.deb.sh | sudo bash;
sudo apt-get update;
sudo apt-get install -y sysbench;
`,

	"tools": `
sudo apt-get update;
sudo apt-get install -y \
  fio \
  iftop \
  iotop \
  sysstat \
  linux-tools-common \
  linux-tools-4.10.0-35-generic \
  linux-cloud-tools-4.10.0-35-generic;
`,

	"zfs": `
sudo apt-get update;
sudo apt-get install -y \
  zfsutils-linux;
`,
}

func SortedCmds() []string {
	__antithesis_instrumentation__.Notify(181593)
	cmds := make([]string, 0, len(installCmds))
	for cmd := range installCmds {
		__antithesis_instrumentation__.Notify(181595)
		cmds = append(cmds, cmd)
	}
	__antithesis_instrumentation__.Notify(181594)
	sort.Strings(cmds)
	return cmds
}

func Install(ctx context.Context, l *logger.Logger, c *SyncedCluster, args []string) error {
	__antithesis_instrumentation__.Notify(181596)
	do := func(title, cmd string) error {
		__antithesis_instrumentation__.Notify(181599)
		var buf bytes.Buffer
		err := c.Run(ctx, l, &buf, &buf, c.Nodes, "installing "+title, cmd)
		if err != nil {
			__antithesis_instrumentation__.Notify(181601)
			l.Printf(buf.String())
		} else {
			__antithesis_instrumentation__.Notify(181602)
		}
		__antithesis_instrumentation__.Notify(181600)
		return err
	}
	__antithesis_instrumentation__.Notify(181597)

	for _, arg := range args {
		__antithesis_instrumentation__.Notify(181603)
		cmd, ok := installCmds[arg]
		if !ok {
			__antithesis_instrumentation__.Notify(181605)
			return fmt.Errorf("unknown tool %q", arg)
		} else {
			__antithesis_instrumentation__.Notify(181606)
		}
		__antithesis_instrumentation__.Notify(181604)

		cmd = "set -exuo pipefail;" + cmd
		if err := do(arg, cmd); err != nil {
			__antithesis_instrumentation__.Notify(181607)
			return err
		} else {
			__antithesis_instrumentation__.Notify(181608)
		}
	}
	__antithesis_instrumentation__.Notify(181598)
	return nil
}
