package cloud

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
)

type Cloud struct {
	Clusters Clusters `json:"clusters"`

	BadInstances vm.List `json:"bad_instances"`
}

func (c *Cloud) BadInstanceErrors() map[error]vm.List {
	__antithesis_instrumentation__.Notify(179953)
	ret := map[error]vm.List{}

	for _, vm := range c.BadInstances {
		__antithesis_instrumentation__.Notify(179956)
		for _, err := range vm.Errors {
			__antithesis_instrumentation__.Notify(179957)
			ret[err] = append(ret[err], vm)
		}
	}
	__antithesis_instrumentation__.Notify(179954)

	for _, v := range ret {
		__antithesis_instrumentation__.Notify(179958)
		sort.Sort(v)
	}
	__antithesis_instrumentation__.Notify(179955)

	return ret
}

type Clusters map[string]*Cluster

func (c Clusters) Names() []string {
	__antithesis_instrumentation__.Notify(179959)
	result := make([]string, 0, len(c))
	for n := range c {
		__antithesis_instrumentation__.Notify(179961)
		result = append(result, n)
	}
	__antithesis_instrumentation__.Notify(179960)
	sort.Strings(result)
	return result
}

func (c Clusters) FilterByName(pattern *regexp.Regexp) Clusters {
	__antithesis_instrumentation__.Notify(179962)
	result := make(Clusters)
	for name, cluster := range c {
		__antithesis_instrumentation__.Notify(179964)
		if pattern.MatchString(name) {
			__antithesis_instrumentation__.Notify(179965)
			result[name] = cluster
		} else {
			__antithesis_instrumentation__.Notify(179966)
		}
	}
	__antithesis_instrumentation__.Notify(179963)
	return result
}

type Cluster struct {
	Name string `json:"name"`
	User string `json:"user"`

	CreatedAt time.Time     `json:"created_at"`
	Lifetime  time.Duration `json:"lifetime"`
	VMs       vm.List       `json:"vms"`
}

func (c *Cluster) Clouds() []string {
	__antithesis_instrumentation__.Notify(179967)
	present := make(map[string]bool)
	for _, m := range c.VMs {
		__antithesis_instrumentation__.Notify(179970)
		p := m.Provider
		if m.Project != "" {
			__antithesis_instrumentation__.Notify(179972)
			p = fmt.Sprintf("%s:%s", m.Provider, m.Project)
		} else {
			__antithesis_instrumentation__.Notify(179973)
		}
		__antithesis_instrumentation__.Notify(179971)
		present[p] = true
	}
	__antithesis_instrumentation__.Notify(179968)

	var ret []string
	for provider := range present {
		__antithesis_instrumentation__.Notify(179974)
		ret = append(ret, provider)
	}
	__antithesis_instrumentation__.Notify(179969)
	sort.Strings(ret)
	return ret
}

func (c *Cluster) ExpiresAt() time.Time {
	__antithesis_instrumentation__.Notify(179975)
	return c.CreatedAt.Add(c.Lifetime)
}

func (c *Cluster) GCAt() time.Time {
	__antithesis_instrumentation__.Notify(179976)

	return c.ExpiresAt().Add(time.Hour - 1).Truncate(time.Hour)
}

func (c *Cluster) LifetimeRemaining() time.Duration {
	__antithesis_instrumentation__.Notify(179977)
	return time.Until(c.GCAt())
}

func (c *Cluster) String() string {
	__antithesis_instrumentation__.Notify(179978)
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s: %d", c.Name, len(c.VMs))
	if !c.IsLocal() {
		__antithesis_instrumentation__.Notify(179980)
		fmt.Fprintf(&buf, " (%s)", c.LifetimeRemaining().Round(time.Second))
	} else {
		__antithesis_instrumentation__.Notify(179981)
	}
	__antithesis_instrumentation__.Notify(179979)
	return buf.String()
}

func (c *Cluster) PrintDetails(logger *logger.Logger) {
	__antithesis_instrumentation__.Notify(179982)
	logger.Printf("%s: %s ", c.Name, c.Clouds())
	if !c.IsLocal() {
		__antithesis_instrumentation__.Notify(179984)
		l := c.LifetimeRemaining().Round(time.Second)
		if l <= 0 {
			__antithesis_instrumentation__.Notify(179985)
			logger.Printf("expired %s ago", -l)
		} else {
			__antithesis_instrumentation__.Notify(179986)
			logger.Printf("%s remaining", l)
		}
	} else {
		__antithesis_instrumentation__.Notify(179987)
		logger.Printf("(no expiration)")
	}
	__antithesis_instrumentation__.Notify(179983)
	for _, vm := range c.VMs {
		__antithesis_instrumentation__.Notify(179988)
		logger.Printf("  %s\t%s\t%s\t%s", vm.Name, vm.DNS, vm.PrivateIP, vm.PublicIP)
	}
}

func (c *Cluster) IsLocal() bool {
	__antithesis_instrumentation__.Notify(179989)
	return config.IsLocalClusterName(c.Name)
}

const vmNameFormat = "user-<clusterid>-<nodeid>"

func namesFromVM(v vm.VM) (userName string, clusterName string, _ error) {
	__antithesis_instrumentation__.Notify(179990)
	if v.IsLocal() {
		__antithesis_instrumentation__.Notify(179993)
		return config.Local, v.LocalClusterName, nil
	} else {
		__antithesis_instrumentation__.Notify(179994)
	}
	__antithesis_instrumentation__.Notify(179991)
	name := v.Name
	parts := strings.Split(name, "-")
	if len(parts) < 3 {
		__antithesis_instrumentation__.Notify(179995)
		return "", "", fmt.Errorf("expected VM name in the form %s, got %s", vmNameFormat, name)
	} else {
		__antithesis_instrumentation__.Notify(179996)
	}
	__antithesis_instrumentation__.Notify(179992)
	return parts[0], strings.Join(parts[:len(parts)-1], "-"), nil
}

func ListCloud(l *logger.Logger) (*Cloud, error) {
	__antithesis_instrumentation__.Notify(179997)
	cloud := &Cloud{
		Clusters: make(Clusters),
	}

	providerNames := vm.AllProviderNames()
	providerVMs := make([]vm.List, len(providerNames))
	var g errgroup.Group
	for i, providerName := range providerNames {
		__antithesis_instrumentation__.Notify(180002)

		index := i
		provider := vm.Providers[providerName]
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(180003)
			var err error
			providerVMs[index], err = provider.List(l)
			return err
		})
	}
	__antithesis_instrumentation__.Notify(179998)

	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(180004)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(180005)
	}
	__antithesis_instrumentation__.Notify(179999)

	for _, vms := range providerVMs {
		__antithesis_instrumentation__.Notify(180006)
		for _, v := range vms {
			__antithesis_instrumentation__.Notify(180007)

			userName, clusterName, err := namesFromVM(v)
			if err != nil {
				__antithesis_instrumentation__.Notify(180012)
				v.Errors = append(v.Errors, vm.ErrInvalidName)
			} else {
				__antithesis_instrumentation__.Notify(180013)
			}
			__antithesis_instrumentation__.Notify(180008)

			if len(v.Errors) > 0 {
				__antithesis_instrumentation__.Notify(180014)
				cloud.BadInstances = append(cloud.BadInstances, v)
				continue
			} else {
				__antithesis_instrumentation__.Notify(180015)
			}
			__antithesis_instrumentation__.Notify(180009)

			if _, ok := cloud.Clusters[clusterName]; !ok {
				__antithesis_instrumentation__.Notify(180016)
				cloud.Clusters[clusterName] = &Cluster{
					Name:      clusterName,
					User:      userName,
					CreatedAt: v.CreatedAt,
					Lifetime:  v.Lifetime,
					VMs:       nil,
				}
			} else {
				__antithesis_instrumentation__.Notify(180017)
			}
			__antithesis_instrumentation__.Notify(180010)

			c := cloud.Clusters[clusterName]
			c.VMs = append(c.VMs, v)
			if v.CreatedAt.Before(c.CreatedAt) {
				__antithesis_instrumentation__.Notify(180018)
				c.CreatedAt = v.CreatedAt
			} else {
				__antithesis_instrumentation__.Notify(180019)
			}
			__antithesis_instrumentation__.Notify(180011)
			if v.Lifetime < c.Lifetime {
				__antithesis_instrumentation__.Notify(180020)
				c.Lifetime = v.Lifetime
			} else {
				__antithesis_instrumentation__.Notify(180021)
			}
		}
	}
	__antithesis_instrumentation__.Notify(180000)

	for _, c := range cloud.Clusters {
		__antithesis_instrumentation__.Notify(180022)
		sort.Sort(c.VMs)
	}
	__antithesis_instrumentation__.Notify(180001)

	return cloud, nil
}

func CreateCluster(
	l *logger.Logger,
	nodes int,
	opts vm.CreateOpts,
	providerOptsContainer vm.ProviderOptionsContainer,
) error {
	__antithesis_instrumentation__.Notify(180023)
	providerCount := len(opts.VMProviders)
	if providerCount == 0 {
		__antithesis_instrumentation__.Notify(180026)
		return errors.New("no VMProviders configured")
	} else {
		__antithesis_instrumentation__.Notify(180027)
	}
	__antithesis_instrumentation__.Notify(180024)

	vmLocations := map[string][]string{}
	for i, p := 1, 0; i <= nodes; i++ {
		__antithesis_instrumentation__.Notify(180028)
		pName := opts.VMProviders[p]
		vmName := vm.Name(opts.ClusterName, i)
		vmLocations[pName] = append(vmLocations[pName], vmName)

		p = (p + 1) % providerCount
	}
	__antithesis_instrumentation__.Notify(180025)

	return vm.ProvidersParallel(opts.VMProviders, func(p vm.Provider) error {
		__antithesis_instrumentation__.Notify(180029)
		return p.Create(l, vmLocations[p.Name()], opts, providerOptsContainer[p.Name()])
	})
}

func DestroyCluster(c *Cluster) error {
	__antithesis_instrumentation__.Notify(180030)
	return vm.FanOut(c.VMs, func(p vm.Provider, vms vm.List) error {
		__antithesis_instrumentation__.Notify(180031)

		if x, ok := p.(vm.DeleteCluster); ok {
			__antithesis_instrumentation__.Notify(180033)
			return x.DeleteCluster(c.Name)
		} else {
			__antithesis_instrumentation__.Notify(180034)
		}
		__antithesis_instrumentation__.Notify(180032)
		return p.Delete(vms)
	})
}

func ExtendCluster(c *Cluster, extension time.Duration) error {
	__antithesis_instrumentation__.Notify(180035)

	newLifetime := (c.Lifetime + extension).Round(time.Second)
	return vm.FanOut(c.VMs, func(p vm.Provider, vms vm.List) error {
		__antithesis_instrumentation__.Notify(180036)
		return p.Extend(vms, newLifetime)
	})
}
