package local

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachprod/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

const ProviderName = config.Local

func VMDir(clusterName string, nodeIdx int) string {
	__antithesis_instrumentation__.Notify(183822)
	if nodeIdx < 1 {
		__antithesis_instrumentation__.Notify(183825)
		panic("invalid nodeIdx")
	} else {
		__antithesis_instrumentation__.Notify(183826)
	}
	__antithesis_instrumentation__.Notify(183823)
	localDir := os.ExpandEnv("${HOME}/local")
	if clusterName == config.Local {
		__antithesis_instrumentation__.Notify(183827)
		return filepath.Join(localDir, fmt.Sprintf("%d", nodeIdx))
	} else {
		__antithesis_instrumentation__.Notify(183828)
	}
	__antithesis_instrumentation__.Notify(183824)
	name := strings.TrimPrefix(clusterName, config.Local+"-")
	return filepath.Join(localDir, fmt.Sprintf("%s-%d", name, nodeIdx))
}

func Init(storage VMStorage) error {
	__antithesis_instrumentation__.Notify(183829)
	vm.Providers[ProviderName] = &Provider{
		clusters: make(cloud.Clusters),
		storage:  storage,
	}
	return nil
}

func AddCluster(cluster *cloud.Cluster) {
	__antithesis_instrumentation__.Notify(183830)
	p := vm.Providers[ProviderName].(*Provider)
	p.clusters[cluster.Name] = cluster
}

func DeleteCluster(l *logger.Logger, name string) error {
	__antithesis_instrumentation__.Notify(183831)
	p := vm.Providers[ProviderName].(*Provider)
	c := p.clusters[name]
	if c == nil {
		__antithesis_instrumentation__.Notify(183835)
		return fmt.Errorf("local cluster %s does not exist", name)
	} else {
		__antithesis_instrumentation__.Notify(183836)
	}
	__antithesis_instrumentation__.Notify(183832)
	l.Printf("Deleting local cluster %s\n", name)

	for i := range c.VMs {
		__antithesis_instrumentation__.Notify(183837)
		if err := os.RemoveAll(VMDir(c.Name, i+1)); err != nil {
			__antithesis_instrumentation__.Notify(183838)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183839)
		}
	}
	__antithesis_instrumentation__.Notify(183833)

	if err := p.storage.DeleteCluster(name); err != nil {
		__antithesis_instrumentation__.Notify(183840)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183841)
	}
	__antithesis_instrumentation__.Notify(183834)

	delete(p.clusters, name)
	return nil
}

func Clusters() []string {
	__antithesis_instrumentation__.Notify(183842)
	p := vm.Providers[ProviderName].(*Provider)
	return p.clusters.Names()
}

type VMStorage interface {
	SaveCluster(cluster *cloud.Cluster) error

	DeleteCluster(name string) error
}

type Provider struct {
	clusters cloud.Clusters

	storage VMStorage
}

type providerOpts struct{}

func (o *providerOpts) ConfigureCreateFlags(flags *pflag.FlagSet) {
	__antithesis_instrumentation__.Notify(183843)
}

func (o *providerOpts) ConfigureClusterFlags(*pflag.FlagSet, vm.MultipleProjectsOption) {
	__antithesis_instrumentation__.Notify(183844)
}

func (p *Provider) CleanSSH() error {
	__antithesis_instrumentation__.Notify(183845)
	return nil
}

func (p *Provider) ConfigSSH(zones []string) error {
	__antithesis_instrumentation__.Notify(183846)
	return nil
}

func (p *Provider) Create(
	l *logger.Logger, names []string, opts vm.CreateOpts, unusedProviderOpts vm.ProviderOpts,
) error {
	__antithesis_instrumentation__.Notify(183847)
	now := timeutil.Now()
	c := &cloud.Cluster{
		Name:      opts.ClusterName,
		CreatedAt: now,
		Lifetime:  time.Hour,
		VMs:       make(vm.List, len(names)),
	}

	if !config.IsLocalClusterName(c.Name) {
		__antithesis_instrumentation__.Notify(183853)
		return errors.Errorf("'%s' is not a valid local cluster name", c.Name)
	} else {
		__antithesis_instrumentation__.Notify(183854)
	}
	__antithesis_instrumentation__.Notify(183848)

	var portsTaken util.FastIntSet
	for _, c := range p.clusters {
		__antithesis_instrumentation__.Notify(183855)
		for i := range c.VMs {
			__antithesis_instrumentation__.Notify(183856)
			portsTaken.Add(c.VMs[i].SQLPort)
			portsTaken.Add(c.VMs[i].AdminUIPort)
		}
	}
	__antithesis_instrumentation__.Notify(183849)
	sqlPort := config.DefaultSQLPort
	adminUIPort := config.DefaultAdminUIPort

	getPort := func(port *int) int {
		__antithesis_instrumentation__.Notify(183857)
		for portsTaken.Contains(*port) {
			__antithesis_instrumentation__.Notify(183859)
			(*port)++
		}
		__antithesis_instrumentation__.Notify(183858)
		result := *port
		portsTaken.Add(result)
		(*port)++
		return result
	}
	__antithesis_instrumentation__.Notify(183850)

	for i := range names {
		__antithesis_instrumentation__.Notify(183860)
		c.VMs[i] = vm.VM{
			Name:             "localhost",
			CreatedAt:        now,
			Lifetime:         time.Hour,
			PrivateIP:        "127.0.0.1",
			Provider:         ProviderName,
			ProviderID:       ProviderName,
			PublicIP:         "127.0.0.1",
			RemoteUser:       config.OSUser.Username,
			VPC:              ProviderName,
			MachineType:      ProviderName,
			Zone:             ProviderName,
			SQLPort:          getPort(&sqlPort),
			AdminUIPort:      getPort(&adminUIPort),
			LocalClusterName: c.Name,
		}

		err := os.MkdirAll(VMDir(c.Name, i+1), 0755)
		if err != nil {
			__antithesis_instrumentation__.Notify(183861)
			return err
		} else {
			__antithesis_instrumentation__.Notify(183862)
		}
	}
	__antithesis_instrumentation__.Notify(183851)
	if err := p.storage.SaveCluster(c); err != nil {
		__antithesis_instrumentation__.Notify(183863)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183864)
	}
	__antithesis_instrumentation__.Notify(183852)
	p.clusters[c.Name] = c
	return nil
}

func (p *Provider) Delete(vms vm.List) error {
	__antithesis_instrumentation__.Notify(183865)
	panic("DeleteCluster should be used")
}

func (p *Provider) Reset(vms vm.List) error {
	__antithesis_instrumentation__.Notify(183866)
	return nil
}

func (p *Provider) Extend(vms vm.List, lifetime time.Duration) error {
	__antithesis_instrumentation__.Notify(183867)
	return errors.New("local clusters have unlimited lifetime")
}

func (p *Provider) FindActiveAccount() (string, error) {
	__antithesis_instrumentation__.Notify(183868)
	return "", nil
}

func (p *Provider) CreateProviderOpts() vm.ProviderOpts {
	__antithesis_instrumentation__.Notify(183869)
	return &providerOpts{}
}

func (p *Provider) List(l *logger.Logger) (vm.List, error) {
	__antithesis_instrumentation__.Notify(183870)
	var result vm.List
	for _, clusterName := range p.clusters.Names() {
		__antithesis_instrumentation__.Notify(183872)
		c := p.clusters[clusterName]
		result = append(result, c.VMs...)
	}
	__antithesis_instrumentation__.Notify(183871)
	return result, nil
}

func (p *Provider) Name() string {
	__antithesis_instrumentation__.Notify(183873)
	return ProviderName
}

func (p *Provider) Active() bool {
	__antithesis_instrumentation__.Notify(183874)
	return true
}

func (p *Provider) ProjectActive(project string) bool {
	__antithesis_instrumentation__.Notify(183875)
	return project == ""
}
