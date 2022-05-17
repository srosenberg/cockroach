package roachprod

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/roachprod/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/aws"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/azure"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/gce"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/local"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"golang.org/x/sys/unix"
)

func verifyClusterName(l *logger.Logger, clusterName, username string) error {
	__antithesis_instrumentation__.Notify(181906)
	if clusterName == "" {
		__antithesis_instrumentation__.Notify(181915)
		return fmt.Errorf("cluster name cannot be blank")
	} else {
		__antithesis_instrumentation__.Notify(181916)
	}
	__antithesis_instrumentation__.Notify(181907)

	alphaNum, err := regexp.Compile(`^[a-zA-Z0-9\-]+$`)
	if err != nil {
		__antithesis_instrumentation__.Notify(181917)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181918)
	}
	__antithesis_instrumentation__.Notify(181908)
	if !alphaNum.MatchString(clusterName) {
		__antithesis_instrumentation__.Notify(181919)
		return errors.Errorf("cluster name must match %s", alphaNum.String())
	} else {
		__antithesis_instrumentation__.Notify(181920)
	}
	__antithesis_instrumentation__.Notify(181909)

	if config.IsLocalClusterName(clusterName) {
		__antithesis_instrumentation__.Notify(181921)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(181922)
	}
	__antithesis_instrumentation__.Notify(181910)

	var accounts []string
	if len(username) > 0 {
		__antithesis_instrumentation__.Notify(181923)
		accounts = []string{username}
	} else {
		__antithesis_instrumentation__.Notify(181924)
		seenAccounts := map[string]bool{}
		active, err := vm.FindActiveAccounts()
		if err != nil {
			__antithesis_instrumentation__.Notify(181926)
			return err
		} else {
			__antithesis_instrumentation__.Notify(181927)
		}
		__antithesis_instrumentation__.Notify(181925)
		for _, account := range active {
			__antithesis_instrumentation__.Notify(181928)
			if !seenAccounts[account] {
				__antithesis_instrumentation__.Notify(181929)
				seenAccounts[account] = true
				cleanAccount := vm.DNSSafeAccount(account)
				if cleanAccount != account {
					__antithesis_instrumentation__.Notify(181931)
					log.Infof(context.TODO(), "WARN: using `%s' as username instead of `%s'", cleanAccount, account)
				} else {
					__antithesis_instrumentation__.Notify(181932)
				}
				__antithesis_instrumentation__.Notify(181930)
				accounts = append(accounts, cleanAccount)
			} else {
				__antithesis_instrumentation__.Notify(181933)
			}
		}
	}
	__antithesis_instrumentation__.Notify(181911)

	for _, account := range accounts {
		__antithesis_instrumentation__.Notify(181934)
		if strings.HasPrefix(clusterName, account+"-") && func() bool {
			__antithesis_instrumentation__.Notify(181935)
			return len(clusterName) > len(account)+1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(181936)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(181937)
		}
	}
	__antithesis_instrumentation__.Notify(181912)

	var suffix string
	if i := strings.Index(clusterName, "-"); i != -1 {
		__antithesis_instrumentation__.Notify(181938)

		suffix = clusterName[i+1:]
	} else {
		__antithesis_instrumentation__.Notify(181939)

		suffix = clusterName
	}
	__antithesis_instrumentation__.Notify(181913)

	var suggestions []string
	for _, account := range accounts {
		__antithesis_instrumentation__.Notify(181940)
		suggestions = append(suggestions, fmt.Sprintf("%s-%s", account, suffix))
	}
	__antithesis_instrumentation__.Notify(181914)
	return fmt.Errorf("malformed cluster name %s, did you mean one of %s",
		clusterName, suggestions)
}

func sortedClusters() []string {
	__antithesis_instrumentation__.Notify(181941)
	var r []string
	syncedClusters.mu.Lock()
	defer syncedClusters.mu.Unlock()
	for n := range syncedClusters.clusters {
		__antithesis_instrumentation__.Notify(181943)
		r = append(r, n)
	}
	__antithesis_instrumentation__.Notify(181942)
	sort.Strings(r)
	return r
}

func newCluster(
	l *logger.Logger, name string, opts ...install.ClusterSettingOption,
) (*install.SyncedCluster, error) {
	__antithesis_instrumentation__.Notify(181944)
	clusterSettings := install.MakeClusterSettings(opts...)
	nodeSelector := "all"
	{
		__antithesis_instrumentation__.Notify(181949)
		parts := strings.Split(name, ":")
		switch len(parts) {
		case 2:
			__antithesis_instrumentation__.Notify(181950)
			nodeSelector = parts[1]
			fallthrough
		case 1:
			__antithesis_instrumentation__.Notify(181951)
			name = parts[0]
		case 0:
			__antithesis_instrumentation__.Notify(181952)
			return nil, fmt.Errorf("no cluster specified")
		default:
			__antithesis_instrumentation__.Notify(181953)
			return nil, fmt.Errorf("invalid cluster name: %s", name)
		}
	}
	__antithesis_instrumentation__.Notify(181945)

	metadata, ok := readSyncedClusters(name)
	if !ok {
		__antithesis_instrumentation__.Notify(181954)
		err := errors.Newf(`unknown cluster: %s`, name)
		err = errors.WithHintf(err, "\nAvailable clusters:\n  %s\n", strings.Join(sortedClusters(), "\n  "))
		err = errors.WithHint(err, `Use "roachprod sync" to update the list of available clusters.`)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181955)
	}
	__antithesis_instrumentation__.Notify(181946)

	if clusterSettings.DebugDir == "" {
		__antithesis_instrumentation__.Notify(181956)
		clusterSettings.DebugDir = os.ExpandEnv(config.DefaultDebugDir)
	} else {
		__antithesis_instrumentation__.Notify(181957)
	}
	__antithesis_instrumentation__.Notify(181947)

	c, err := install.NewSyncedCluster(metadata, nodeSelector, clusterSettings)
	if err != nil {
		__antithesis_instrumentation__.Notify(181958)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181959)
	}
	__antithesis_instrumentation__.Notify(181948)

	return c, nil
}

func userClusterNameRegexp(l *logger.Logger) (*regexp.Regexp, error) {
	__antithesis_instrumentation__.Notify(181960)

	seenAccounts := map[string]bool{}
	accounts, err := vm.FindActiveAccounts()
	if err != nil {
		__antithesis_instrumentation__.Notify(181963)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181964)
	}
	__antithesis_instrumentation__.Notify(181961)
	pattern := ""
	for _, account := range accounts {
		__antithesis_instrumentation__.Notify(181965)
		if !seenAccounts[account] {
			__antithesis_instrumentation__.Notify(181966)
			seenAccounts[account] = true
			if len(pattern) > 0 {
				__antithesis_instrumentation__.Notify(181968)
				pattern += "|"
			} else {
				__antithesis_instrumentation__.Notify(181969)
			}
			__antithesis_instrumentation__.Notify(181967)
			pattern += fmt.Sprintf("(^%s-)", regexp.QuoteMeta(account))
		} else {
			__antithesis_instrumentation__.Notify(181970)
		}
	}
	__antithesis_instrumentation__.Notify(181962)
	return regexp.Compile(pattern)
}

func Version(l *logger.Logger) string {
	__antithesis_instrumentation__.Notify(181971)
	info := build.GetInfo()
	return info.Long()
}

func CachedClusters(l *logger.Logger, fn func(clusterName string, numVMs int)) {
	__antithesis_instrumentation__.Notify(181972)
	for _, name := range sortedClusters() {
		__antithesis_instrumentation__.Notify(181973)
		c, ok := readSyncedClusters(name)
		if !ok {
			__antithesis_instrumentation__.Notify(181975)
			return
		} else {
			__antithesis_instrumentation__.Notify(181976)
		}
		__antithesis_instrumentation__.Notify(181974)
		fn(c.Name, len(c.VMs))
	}
}

func acquireFilesystemLock() (unlockFn func(), _ error) {
	__antithesis_instrumentation__.Notify(181977)
	lockFile := os.ExpandEnv("$HOME/.roachprod/LOCK")
	f, err := os.Create(lockFile)
	if err != nil {
		__antithesis_instrumentation__.Notify(181980)
		return nil, errors.Wrapf(err, "creating lock file %q", lockFile)
	} else {
		__antithesis_instrumentation__.Notify(181981)
	}
	__antithesis_instrumentation__.Notify(181978)
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		__antithesis_instrumentation__.Notify(181982)
		f.Close()
		return nil, errors.Wrap(err, "acquiring lock on %q")
	} else {
		__antithesis_instrumentation__.Notify(181983)
	}
	__antithesis_instrumentation__.Notify(181979)
	return func() {
		__antithesis_instrumentation__.Notify(181984)
		f.Close()
	}, nil
}

func Sync(l *logger.Logger) (*cloud.Cloud, error) {
	__antithesis_instrumentation__.Notify(181985)
	if !config.Quiet {
		__antithesis_instrumentation__.Notify(181995)
		l.Printf("Syncing...")
	} else {
		__antithesis_instrumentation__.Notify(181996)
	}
	__antithesis_instrumentation__.Notify(181986)
	unlock, err := acquireFilesystemLock()
	if err != nil {
		__antithesis_instrumentation__.Notify(181997)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181998)
	}
	__antithesis_instrumentation__.Notify(181987)
	defer unlock()

	cld, err := cloud.ListCloud(l)
	if err != nil {
		__antithesis_instrumentation__.Notify(181999)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182000)
	}
	__antithesis_instrumentation__.Notify(181988)
	if err := syncClustersCache(cld); err != nil {
		__antithesis_instrumentation__.Notify(182001)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182002)
	}
	__antithesis_instrumentation__.Notify(181989)

	var vms vm.List
	for _, c := range cld.Clusters {
		__antithesis_instrumentation__.Notify(182003)
		vms = append(vms, c.VMs...)
	}
	__antithesis_instrumentation__.Notify(181990)

	refreshDNS := true

	if p := vm.Providers[gce.ProviderName]; !p.Active() {
		__antithesis_instrumentation__.Notify(182004)
		refreshDNS = false
	} else {
		__antithesis_instrumentation__.Notify(182005)
		var defaultProjectFound bool
		for _, prj := range p.(*gce.Provider).GetProjects() {
			__antithesis_instrumentation__.Notify(182007)
			if prj == gce.DefaultProject() {
				__antithesis_instrumentation__.Notify(182008)
				defaultProjectFound = true
				break
			} else {
				__antithesis_instrumentation__.Notify(182009)
			}
		}
		__antithesis_instrumentation__.Notify(182006)
		if !defaultProjectFound {
			__antithesis_instrumentation__.Notify(182010)
			refreshDNS = false
		} else {
			__antithesis_instrumentation__.Notify(182011)
		}
	}
	__antithesis_instrumentation__.Notify(181991)
	if !vm.Providers[aws.ProviderName].Active() {
		__antithesis_instrumentation__.Notify(182012)
		refreshDNS = false
	} else {
		__antithesis_instrumentation__.Notify(182013)
	}
	__antithesis_instrumentation__.Notify(181992)

	if refreshDNS {
		__antithesis_instrumentation__.Notify(182014)
		if !config.Quiet {
			__antithesis_instrumentation__.Notify(182016)
			l.Printf("Refreshing DNS entries...")
		} else {
			__antithesis_instrumentation__.Notify(182017)
		}
		__antithesis_instrumentation__.Notify(182015)
		if err := gce.SyncDNS(l, vms); err != nil {
			__antithesis_instrumentation__.Notify(182018)
			fmt.Fprintf(l.Stderr, "failed to update %s DNS: %v", gce.Subdomain, err)
		} else {
			__antithesis_instrumentation__.Notify(182019)
		}
	} else {
		__antithesis_instrumentation__.Notify(182020)
		if !config.Quiet {
			__antithesis_instrumentation__.Notify(182021)
			l.Printf("Not refreshing DNS entries. We did not have all the VMs.")
		} else {
			__antithesis_instrumentation__.Notify(182022)
		}
	}
	__antithesis_instrumentation__.Notify(181993)

	if err := vm.ProvidersSequential(vm.AllProviderNames(), func(p vm.Provider) error {
		__antithesis_instrumentation__.Notify(182023)
		return p.CleanSSH()
	}); err != nil {
		__antithesis_instrumentation__.Notify(182024)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182025)
	}
	__antithesis_instrumentation__.Notify(181994)

	return cld, nil
}

func List(l *logger.Logger, listMine bool, clusterNamePattern string) (cloud.Cloud, error) {
	__antithesis_instrumentation__.Notify(182026)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182030)
		return cloud.Cloud{}, err
	} else {
		__antithesis_instrumentation__.Notify(182031)
	}
	__antithesis_instrumentation__.Notify(182027)
	listPattern := regexp.MustCompile(".*")
	if clusterNamePattern == "" {
		__antithesis_instrumentation__.Notify(182032)
		if listMine {
			__antithesis_instrumentation__.Notify(182033)
			var err error
			listPattern, err = userClusterNameRegexp(l)
			if err != nil {
				__antithesis_instrumentation__.Notify(182034)
				return cloud.Cloud{}, err
			} else {
				__antithesis_instrumentation__.Notify(182035)
			}
		} else {
			__antithesis_instrumentation__.Notify(182036)
		}
	} else {
		__antithesis_instrumentation__.Notify(182037)
		if listMine {
			__antithesis_instrumentation__.Notify(182039)
			return cloud.Cloud{}, errors.New("'mine' option cannot be combined with 'pattern'")
		} else {
			__antithesis_instrumentation__.Notify(182040)
		}
		__antithesis_instrumentation__.Notify(182038)
		var err error
		listPattern, err = regexp.Compile(clusterNamePattern)
		if err != nil {
			__antithesis_instrumentation__.Notify(182041)
			return cloud.Cloud{}, errors.Wrapf(err, "could not compile regex pattern: %s", clusterNamePattern)
		} else {
			__antithesis_instrumentation__.Notify(182042)
		}
	}
	__antithesis_instrumentation__.Notify(182028)

	cld, err := Sync(l)
	if err != nil {
		__antithesis_instrumentation__.Notify(182043)
		return cloud.Cloud{}, err
	} else {
		__antithesis_instrumentation__.Notify(182044)
	}
	__antithesis_instrumentation__.Notify(182029)

	filteredClusters := cld.Clusters.FilterByName(listPattern)
	filteredCloud := cloud.Cloud{
		Clusters:     filteredClusters,
		BadInstances: cld.BadInstances,
	}
	return filteredCloud, nil
}

func Run(
	ctx context.Context,
	l *logger.Logger,
	clusterName, SSHOptions, processTag string,
	secure bool,
	stdout, stderr io.Writer,
	cmdArray []string,
) error {
	__antithesis_instrumentation__.Notify(182045)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182050)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182051)
	}
	__antithesis_instrumentation__.Notify(182046)
	c, err := newCluster(l, clusterName, install.SecureOption(secure), install.TagOption(processTag))
	if err != nil {
		__antithesis_instrumentation__.Notify(182052)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182053)
	}
	__antithesis_instrumentation__.Notify(182047)

	if len(cmdArray) == 0 {
		__antithesis_instrumentation__.Notify(182054)
		return c.SSH(ctx, l, strings.Split(SSHOptions, " "), cmdArray)
	} else {
		__antithesis_instrumentation__.Notify(182055)
	}
	__antithesis_instrumentation__.Notify(182048)

	cmd := strings.TrimSpace(strings.Join(cmdArray, " "))
	title := cmd
	if len(title) > 30 {
		__antithesis_instrumentation__.Notify(182056)
		title = title[:27] + "..."
	} else {
		__antithesis_instrumentation__.Notify(182057)
	}
	__antithesis_instrumentation__.Notify(182049)
	return c.Run(ctx, l, stdout, stderr, c.Nodes, title, cmd)
}

func RunWithDetails(
	ctx context.Context,
	l *logger.Logger,
	clusterName, SSHOptions, processTag string,
	secure bool,
	cmdArray []string,
) ([]install.RunResultDetails, error) {
	__antithesis_instrumentation__.Notify(182058)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182063)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182064)
	}
	__antithesis_instrumentation__.Notify(182059)
	c, err := newCluster(l, clusterName, install.SecureOption(secure), install.TagOption(processTag))
	if err != nil {
		__antithesis_instrumentation__.Notify(182065)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182066)
	}
	__antithesis_instrumentation__.Notify(182060)

	if len(cmdArray) == 0 {
		__antithesis_instrumentation__.Notify(182067)
		return nil, c.SSH(ctx, l, strings.Split(SSHOptions, " "), cmdArray)
	} else {
		__antithesis_instrumentation__.Notify(182068)
	}
	__antithesis_instrumentation__.Notify(182061)

	cmd := strings.TrimSpace(strings.Join(cmdArray, " "))
	title := cmd
	if len(title) > 30 {
		__antithesis_instrumentation__.Notify(182069)
		title = title[:27] + "..."
	} else {
		__antithesis_instrumentation__.Notify(182070)
	}
	__antithesis_instrumentation__.Notify(182062)
	return c.RunWithDetails(ctx, l, c.Nodes, title, cmd)
}

func SQL(
	ctx context.Context, l *logger.Logger, clusterName string, secure bool, cmdArray []string,
) error {
	__antithesis_instrumentation__.Notify(182071)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182074)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182075)
	}
	__antithesis_instrumentation__.Notify(182072)
	c, err := newCluster(l, clusterName, install.SecureOption(secure))
	if err != nil {
		__antithesis_instrumentation__.Notify(182076)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182077)
	}
	__antithesis_instrumentation__.Notify(182073)
	return c.SQL(ctx, l, cmdArray)
}

func IP(
	ctx context.Context, l *logger.Logger, clusterName string, external bool,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(182078)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182082)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182083)
	}
	__antithesis_instrumentation__.Notify(182079)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182084)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182085)
	}
	__antithesis_instrumentation__.Notify(182080)

	nodes := c.TargetNodes()
	ips := make([]string, len(nodes))

	if external {
		__antithesis_instrumentation__.Notify(182086)
		for i := 0; i < len(nodes); i++ {
			__antithesis_instrumentation__.Notify(182087)
			ips[i] = c.VMs[nodes[i]-1].PublicIP
		}
	} else {
		__antithesis_instrumentation__.Notify(182088)
		var err error
		if err := c.Parallel(l, "", len(nodes), 0, func(i int) ([]byte, error) {
			__antithesis_instrumentation__.Notify(182089)
			ips[i], err = c.GetInternalIP(ctx, nodes[i])
			return nil, err
		}); err != nil {
			__antithesis_instrumentation__.Notify(182090)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(182091)
		}
	}
	__antithesis_instrumentation__.Notify(182081)
	return ips, nil
}

func Status(ctx context.Context, l *logger.Logger, clusterName, processTag string) error {
	__antithesis_instrumentation__.Notify(182092)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182095)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182096)
	}
	__antithesis_instrumentation__.Notify(182093)
	c, err := newCluster(l, clusterName, install.TagOption(processTag))
	if err != nil {
		__antithesis_instrumentation__.Notify(182097)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182098)
	}
	__antithesis_instrumentation__.Notify(182094)
	return c.Status(ctx, l)
}

func Stage(
	ctx context.Context,
	l *logger.Logger,
	clusterName string,
	stageOS, stageDir, applicationName, version string,
) error {
	__antithesis_instrumentation__.Notify(182099)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182104)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182105)
	}
	__antithesis_instrumentation__.Notify(182100)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182106)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182107)
	}
	__antithesis_instrumentation__.Notify(182101)

	os := "linux"
	if stageOS != "" {
		__antithesis_instrumentation__.Notify(182108)
		os = stageOS
	} else {
		__antithesis_instrumentation__.Notify(182109)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(182110)
			os = runtime.GOOS
		} else {
			__antithesis_instrumentation__.Notify(182111)
		}
	}
	__antithesis_instrumentation__.Notify(182102)

	dir := "."
	if stageDir != "" {
		__antithesis_instrumentation__.Notify(182112)
		dir = stageDir
	} else {
		__antithesis_instrumentation__.Notify(182113)
	}
	__antithesis_instrumentation__.Notify(182103)

	return install.StageApplication(ctx, l, c, applicationName, version, os, dir)
}

func Reset(l *logger.Logger, clusterName string) error {
	__antithesis_instrumentation__.Notify(182114)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182119)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182120)
	}
	__antithesis_instrumentation__.Notify(182115)

	if config.IsLocalClusterName(clusterName) {
		__antithesis_instrumentation__.Notify(182121)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(182122)
	}
	__antithesis_instrumentation__.Notify(182116)

	cld, err := cloud.ListCloud(l)
	if err != nil {
		__antithesis_instrumentation__.Notify(182123)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182124)
	}
	__antithesis_instrumentation__.Notify(182117)
	c, ok := cld.Clusters[clusterName]
	if !ok {
		__antithesis_instrumentation__.Notify(182125)
		return errors.New("cluster not found")
	} else {
		__antithesis_instrumentation__.Notify(182126)
	}
	__antithesis_instrumentation__.Notify(182118)

	return vm.FanOut(c.VMs, func(p vm.Provider, vms vm.List) error {
		__antithesis_instrumentation__.Notify(182127)
		return p.Reset(vms)
	})
}

func SetupSSH(ctx context.Context, l *logger.Logger, clusterName string) error {
	__antithesis_instrumentation__.Notify(182128)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182141)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182142)
	}
	__antithesis_instrumentation__.Notify(182129)
	cld, err := Sync(l)
	if err != nil {
		__antithesis_instrumentation__.Notify(182143)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182144)
	}
	__antithesis_instrumentation__.Notify(182130)

	cloudCluster, ok := cld.Clusters[clusterName]
	if !ok {
		__antithesis_instrumentation__.Notify(182145)
		return fmt.Errorf("could not find %s in list of cluster", clusterName)
	} else {
		__antithesis_instrumentation__.Notify(182146)
	}
	__antithesis_instrumentation__.Notify(182131)

	zones := make(map[string][]string, len(cloudCluster.VMs))
	for _, vm := range cloudCluster.VMs {
		__antithesis_instrumentation__.Notify(182147)
		zones[vm.Provider] = append(zones[vm.Provider], vm.Zone)
	}
	__antithesis_instrumentation__.Notify(182132)
	providers := make([]string, 0)
	for provider := range zones {
		__antithesis_instrumentation__.Notify(182148)
		providers = append(providers, provider)
	}
	__antithesis_instrumentation__.Notify(182133)

	if err := vm.ProvidersSequential(providers, func(p vm.Provider) error {
		__antithesis_instrumentation__.Notify(182149)
		return p.ConfigSSH(zones[p.Name()])
	}); err != nil {
		__antithesis_instrumentation__.Notify(182150)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182151)
	}
	__antithesis_instrumentation__.Notify(182134)

	cloudCluster.PrintDetails(l)

	for _, v := range cloudCluster.VMs {
		__antithesis_instrumentation__.Notify(182152)
		cmd := exec.Command("ssh-keygen", "-R", v.PublicIP)
		out, err := cmd.CombinedOutput()
		if err != nil {
			__antithesis_instrumentation__.Notify(182153)
			log.Infof(context.TODO(), "could not clear ssh key for hostname %s:\n%s", v.PublicIP, string(out))
		} else {
			__antithesis_instrumentation__.Notify(182154)
		}

	}
	__antithesis_instrumentation__.Notify(182135)

	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182155)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182156)
	}
	__antithesis_instrumentation__.Notify(182136)

	installCluster, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182157)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182158)
	}
	__antithesis_instrumentation__.Notify(182137)

	for i := range installCluster.VMs {
		__antithesis_instrumentation__.Notify(182159)
		if cloudCluster.VMs[i].Provider == gce.ProviderName {
			__antithesis_instrumentation__.Notify(182160)
			installCluster.VMs[i].RemoteUser = config.OSUser.Username
		} else {
			__antithesis_instrumentation__.Notify(182161)
		}
	}
	__antithesis_instrumentation__.Notify(182138)
	if err := installCluster.Wait(ctx, l); err != nil {
		__antithesis_instrumentation__.Notify(182162)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182163)
	}
	__antithesis_instrumentation__.Notify(182139)

	installCluster.AuthorizedKeys, err = gce.GetUserAuthorizedKeys()
	if err != nil {
		__antithesis_instrumentation__.Notify(182164)
		return errors.Wrap(err, "failed to retrieve authorized keys from gcloud")
	} else {
		__antithesis_instrumentation__.Notify(182165)
	}
	__antithesis_instrumentation__.Notify(182140)
	return installCluster.SetupSSH(ctx, l)
}

func Extend(l *logger.Logger, clusterName string, lifetime time.Duration) error {
	__antithesis_instrumentation__.Notify(182166)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182173)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182174)
	}
	__antithesis_instrumentation__.Notify(182167)
	cld, err := cloud.ListCloud(l)
	if err != nil {
		__antithesis_instrumentation__.Notify(182175)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182176)
	}
	__antithesis_instrumentation__.Notify(182168)

	c, ok := cld.Clusters[clusterName]
	if !ok {
		__antithesis_instrumentation__.Notify(182177)
		return fmt.Errorf("cluster %s does not exist", clusterName)
	} else {
		__antithesis_instrumentation__.Notify(182178)
	}
	__antithesis_instrumentation__.Notify(182169)

	if err := cloud.ExtendCluster(c, lifetime); err != nil {
		__antithesis_instrumentation__.Notify(182179)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182180)
	}
	__antithesis_instrumentation__.Notify(182170)

	cld, err = cloud.ListCloud(l)
	if err != nil {
		__antithesis_instrumentation__.Notify(182181)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182182)
	}
	__antithesis_instrumentation__.Notify(182171)

	c, ok = cld.Clusters[clusterName]
	if !ok {
		__antithesis_instrumentation__.Notify(182183)
		return fmt.Errorf("cluster %s does not exist", clusterName)
	} else {
		__antithesis_instrumentation__.Notify(182184)
	}
	__antithesis_instrumentation__.Notify(182172)

	c.PrintDetails(l)
	return nil
}

func DefaultStartOpts() install.StartOpts {
	__antithesis_instrumentation__.Notify(182185)
	return install.StartOpts{
		Sequential:      true,
		EncryptedStores: false,
		NumFilesLimit:   config.DefaultNumFilesLimit,
		SkipInit:        false,
		StoreCount:      1,
		TenantID:        2,
	}
}

func Start(
	ctx context.Context,
	l *logger.Logger,
	clusterName string,
	startOpts install.StartOpts,
	clusterSettingsOpts ...install.ClusterSettingOption,
) error {
	__antithesis_instrumentation__.Notify(182186)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182189)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182190)
	}
	__antithesis_instrumentation__.Notify(182187)
	c, err := newCluster(l, clusterName, clusterSettingsOpts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(182191)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182192)
	}
	__antithesis_instrumentation__.Notify(182188)
	return c.Start(ctx, l, startOpts)
}

func Monitor(
	ctx context.Context, l *logger.Logger, clusterName string, opts install.MonitorOpts,
) (chan install.NodeMonitorInfo, error) {
	__antithesis_instrumentation__.Notify(182193)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182195)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182196)
	}
	__antithesis_instrumentation__.Notify(182194)
	return c.Monitor(ctx, opts), nil
}

type StopOpts struct {
	ProcessTag string
	Sig        int
	Wait       bool
}

func DefaultStopOpts() StopOpts {
	__antithesis_instrumentation__.Notify(182197)
	return StopOpts{
		ProcessTag: "",
		Sig:        9,
		Wait:       false,
	}
}

func Stop(ctx context.Context, l *logger.Logger, clusterName string, opts StopOpts) error {
	__antithesis_instrumentation__.Notify(182198)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182201)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182202)
	}
	__antithesis_instrumentation__.Notify(182199)
	c, err := newCluster(l, clusterName, install.TagOption(opts.ProcessTag))
	if err != nil {
		__antithesis_instrumentation__.Notify(182203)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182204)
	}
	__antithesis_instrumentation__.Notify(182200)
	return c.Stop(ctx, l, opts.Sig, opts.Wait)
}

func Init(ctx context.Context, l *logger.Logger, clusterName string) error {
	__antithesis_instrumentation__.Notify(182205)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182208)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182209)
	}
	__antithesis_instrumentation__.Notify(182206)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182210)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182211)
	}
	__antithesis_instrumentation__.Notify(182207)
	return c.Init(ctx, l)
}

func Wipe(ctx context.Context, l *logger.Logger, clusterName string, preserveCerts bool) error {
	__antithesis_instrumentation__.Notify(182212)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182215)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182216)
	}
	__antithesis_instrumentation__.Notify(182213)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182217)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182218)
	}
	__antithesis_instrumentation__.Notify(182214)
	return c.Wipe(ctx, l, preserveCerts)
}

func Reformat(ctx context.Context, l *logger.Logger, clusterName string, fs string) error {
	__antithesis_instrumentation__.Notify(182219)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182224)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182225)
	}
	__antithesis_instrumentation__.Notify(182220)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182226)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182227)
	}
	__antithesis_instrumentation__.Notify(182221)

	var fsCmd string
	switch fs {
	case vm.Zfs:
		__antithesis_instrumentation__.Notify(182228)
		if err := install.Install(ctx, l, c, []string{vm.Zfs}); err != nil {
			__antithesis_instrumentation__.Notify(182232)
			return err
		} else {
			__antithesis_instrumentation__.Notify(182233)
		}
		__antithesis_instrumentation__.Notify(182229)
		fsCmd = `sudo zpool create -f data1 -m /mnt/data1 /dev/sdb`
	case vm.Ext4:
		__antithesis_instrumentation__.Notify(182230)
		fsCmd = `sudo mkfs.ext4 -F /dev/sdb && sudo mount -o defaults /dev/sdb /mnt/data1`
	default:
		__antithesis_instrumentation__.Notify(182231)
		return fmt.Errorf("unknown filesystem %q", fs)
	}
	__antithesis_instrumentation__.Notify(182222)

	err = c.Run(ctx, l, os.Stdout, os.Stderr, c.Nodes, "reformatting", fmt.Sprintf(`
set -euo pipefail
if sudo zpool list -Ho name 2>/dev/null | grep ^data1$; then
sudo zpool destroy -f data1
fi
if mountpoint -q /mnt/data1; then
sudo umount -f /mnt/data1
fi
%s
sudo chmod 777 /mnt/data1
`, fsCmd))
	if err != nil {
		__antithesis_instrumentation__.Notify(182234)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182235)
	}
	__antithesis_instrumentation__.Notify(182223)
	return nil
}

func Install(ctx context.Context, l *logger.Logger, clusterName string, software []string) error {
	__antithesis_instrumentation__.Notify(182236)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182239)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182240)
	}
	__antithesis_instrumentation__.Notify(182237)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182241)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182242)
	}
	__antithesis_instrumentation__.Notify(182238)
	return install.Install(ctx, l, c, software)
}

func Download(
	ctx context.Context, l *logger.Logger, clusterName string, src, sha, dest string,
) error {
	__antithesis_instrumentation__.Notify(182243)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182246)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182247)
	}
	__antithesis_instrumentation__.Notify(182244)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182248)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182249)
	}
	__antithesis_instrumentation__.Notify(182245)
	return install.Download(ctx, l, c, src, sha, dest)
}

func DistributeCerts(ctx context.Context, l *logger.Logger, clusterName string) error {
	__antithesis_instrumentation__.Notify(182250)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182253)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182254)
	}
	__antithesis_instrumentation__.Notify(182251)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182255)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182256)
	}
	__antithesis_instrumentation__.Notify(182252)
	return c.DistributeCerts(ctx, l)
}

func Put(
	ctx context.Context, l *logger.Logger, clusterName, src, dest string, useTreeDist bool,
) error {
	__antithesis_instrumentation__.Notify(182257)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182260)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182261)
	}
	__antithesis_instrumentation__.Notify(182258)
	c, err := newCluster(l, clusterName, install.UseTreeDistOption(useTreeDist))
	if err != nil {
		__antithesis_instrumentation__.Notify(182262)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182263)
	}
	__antithesis_instrumentation__.Notify(182259)
	return c.Put(ctx, l, src, dest)
}

func Get(l *logger.Logger, clusterName, src, dest string) error {
	__antithesis_instrumentation__.Notify(182264)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182267)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182268)
	}
	__antithesis_instrumentation__.Notify(182265)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182269)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182270)
	}
	__antithesis_instrumentation__.Notify(182266)
	return c.Get(l, src, dest)
}

func PgURL(
	ctx context.Context, l *logger.Logger, clusterName, certsDir string, external, secure bool,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(182271)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182277)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182278)
	}
	__antithesis_instrumentation__.Notify(182272)
	c, err := newCluster(l, clusterName, install.SecureOption(secure), install.PGUrlCertsDirOption(certsDir))
	if err != nil {
		__antithesis_instrumentation__.Notify(182279)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182280)
	}
	__antithesis_instrumentation__.Notify(182273)
	nodes := c.TargetNodes()
	ips := make([]string, len(nodes))

	if external {
		__antithesis_instrumentation__.Notify(182281)
		for i := 0; i < len(nodes); i++ {
			__antithesis_instrumentation__.Notify(182282)
			ips[i] = c.VMs[nodes[i]-1].PublicIP
		}
	} else {
		__antithesis_instrumentation__.Notify(182283)
		var err error
		if err := c.Parallel(l, "", len(nodes), 0, func(i int) ([]byte, error) {
			__antithesis_instrumentation__.Notify(182284)
			ips[i], err = c.GetInternalIP(ctx, nodes[i])
			return nil, err
		}); err != nil {
			__antithesis_instrumentation__.Notify(182285)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(182286)
		}
	}
	__antithesis_instrumentation__.Notify(182274)

	var urls []string
	for i, ip := range ips {
		__antithesis_instrumentation__.Notify(182287)
		if ip == "" {
			__antithesis_instrumentation__.Notify(182289)
			return nil, errors.Errorf("empty ip: %v", ips)
		} else {
			__antithesis_instrumentation__.Notify(182290)
		}
		__antithesis_instrumentation__.Notify(182288)
		urls = append(urls, c.NodeURL(ip, c.NodePort(nodes[i])))
	}
	__antithesis_instrumentation__.Notify(182275)
	if len(urls) != len(nodes) {
		__antithesis_instrumentation__.Notify(182291)
		return nil, errors.Errorf("have nodes %v, but urls %v from ips %v", nodes, urls, ips)
	} else {
		__antithesis_instrumentation__.Notify(182292)
	}
	__antithesis_instrumentation__.Notify(182276)
	return urls, nil
}

func AdminURL(
	l *logger.Logger, clusterName, path string, usePublicIPs, openInBrowser, secure bool,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(182293)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182297)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182298)
	}
	__antithesis_instrumentation__.Notify(182294)
	c, err := newCluster(l, clusterName, install.SecureOption(secure))
	if err != nil {
		__antithesis_instrumentation__.Notify(182299)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182300)
	}
	__antithesis_instrumentation__.Notify(182295)

	var urls []string
	for i, node := range c.TargetNodes() {
		__antithesis_instrumentation__.Notify(182301)
		host := vm.Name(c.Name, int(node)) + "." + gce.Subdomain

		if i == 0 && func() bool {
			__antithesis_instrumentation__.Notify(182306)
			return !usePublicIPs == true
		}() == true {
			__antithesis_instrumentation__.Notify(182307)
			if _, err := net.LookupHost(host); err != nil {
				__antithesis_instrumentation__.Notify(182308)
				fmt.Fprintf(l.Stderr, "no valid DNS (yet?). might need to re-run `sync`?\n")
				usePublicIPs = true
			} else {
				__antithesis_instrumentation__.Notify(182309)
			}
		} else {
			__antithesis_instrumentation__.Notify(182310)
		}
		__antithesis_instrumentation__.Notify(182302)

		if usePublicIPs {
			__antithesis_instrumentation__.Notify(182311)
			host = c.VMs[node-1].PublicIP
		} else {
			__antithesis_instrumentation__.Notify(182312)
		}
		__antithesis_instrumentation__.Notify(182303)
		port := c.NodeUIPort(node)
		scheme := "http"
		if c.Secure {
			__antithesis_instrumentation__.Notify(182313)
			scheme = "https"
		} else {
			__antithesis_instrumentation__.Notify(182314)
		}
		__antithesis_instrumentation__.Notify(182304)
		if !strings.HasPrefix(path, "/") {
			__antithesis_instrumentation__.Notify(182315)
			path = "/" + path
		} else {
			__antithesis_instrumentation__.Notify(182316)
		}
		__antithesis_instrumentation__.Notify(182305)
		url := fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path)
		if openInBrowser {
			__antithesis_instrumentation__.Notify(182317)
			if err := exec.Command("python", "-m", "webbrowser", url).Run(); err != nil {
				__antithesis_instrumentation__.Notify(182318)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(182319)
			}
		} else {
			__antithesis_instrumentation__.Notify(182320)
			urls = append(urls, url)
		}
	}
	__antithesis_instrumentation__.Notify(182296)
	return urls, nil
}

type PprofOpts struct {
	Heap         bool
	Open         bool
	StartingPort int
	Duration     time.Duration
}

func Pprof(l *logger.Logger, clusterName string, opts PprofOpts) error {
	__antithesis_instrumentation__.Notify(182321)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182330)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182331)
	}
	__antithesis_instrumentation__.Notify(182322)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182332)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182333)
	}
	__antithesis_instrumentation__.Notify(182323)

	var profType string
	var description string
	if opts.Heap {
		__antithesis_instrumentation__.Notify(182334)
		description = "capturing heap profile"
		profType = "heap"
	} else {
		__antithesis_instrumentation__.Notify(182335)
		description = "capturing CPU profile"
		profType = "profile"
	}
	__antithesis_instrumentation__.Notify(182324)

	outputFiles := []string{}
	mu := &syncutil.Mutex{}
	pprofPath := fmt.Sprintf("debug/pprof/%s?seconds=%d", profType, int(opts.Duration.Seconds()))

	minTimeout := 30 * time.Second
	timeout := 2 * opts.Duration
	if timeout < minTimeout {
		__antithesis_instrumentation__.Notify(182336)
		timeout = minTimeout
	} else {
		__antithesis_instrumentation__.Notify(182337)
	}
	__antithesis_instrumentation__.Notify(182325)

	httpClient := httputil.NewClientWithTimeout(timeout)
	startTime := timeutil.Now().Unix()
	nodes := c.TargetNodes()
	failed, err := c.ParallelE(l, description, len(nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(182338)
		node := nodes[i]
		host := c.Host(node)
		port := c.NodeUIPort(node)
		scheme := "http"
		if c.Secure {
			__antithesis_instrumentation__.Notify(182348)
			scheme = "https"
		} else {
			__antithesis_instrumentation__.Notify(182349)
		}
		__antithesis_instrumentation__.Notify(182339)
		outputFile := fmt.Sprintf("pprof-%s-%d-%s-%04d.out", profType, startTime, c.Name, node)
		outputDir := filepath.Dir(outputFile)
		file, err := ioutil.TempFile(outputDir, ".pprof")
		if err != nil {
			__antithesis_instrumentation__.Notify(182350)
			return nil, errors.Wrap(err, "create tmpfile for pprof download")
		} else {
			__antithesis_instrumentation__.Notify(182351)
		}
		__antithesis_instrumentation__.Notify(182340)

		defer func() {
			__antithesis_instrumentation__.Notify(182352)
			err := file.Close()
			if err != nil && func() bool {
				__antithesis_instrumentation__.Notify(182354)
				return !errors.Is(err, oserror.ErrClosed) == true
			}() == true {
				__antithesis_instrumentation__.Notify(182355)
				fmt.Fprintf(l.Stderr, "warning: could not close temporary file")
			} else {
				__antithesis_instrumentation__.Notify(182356)
			}
			__antithesis_instrumentation__.Notify(182353)
			err = os.Remove(file.Name())
			if err != nil && func() bool {
				__antithesis_instrumentation__.Notify(182357)
				return !oserror.IsNotExist(err) == true
			}() == true {
				__antithesis_instrumentation__.Notify(182358)
				fmt.Fprintf(l.Stderr, "warning: could not remove temporary file")
			} else {
				__antithesis_instrumentation__.Notify(182359)
			}
		}()
		__antithesis_instrumentation__.Notify(182341)

		pprofURL := fmt.Sprintf("%s://%s:%d/%s", scheme, host, port, pprofPath)
		resp, err := httpClient.Get(context.Background(), pprofURL)
		if err != nil {
			__antithesis_instrumentation__.Notify(182360)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(182361)
		}
		__antithesis_instrumentation__.Notify(182342)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			__antithesis_instrumentation__.Notify(182362)
			return nil, errors.Newf("unexpected status from pprof endpoint: %s", resp.Status)
		} else {
			__antithesis_instrumentation__.Notify(182363)
		}
		__antithesis_instrumentation__.Notify(182343)

		if _, err := io.Copy(file, resp.Body); err != nil {
			__antithesis_instrumentation__.Notify(182364)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(182365)
		}
		__antithesis_instrumentation__.Notify(182344)
		if err := file.Sync(); err != nil {
			__antithesis_instrumentation__.Notify(182366)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(182367)
		}
		__antithesis_instrumentation__.Notify(182345)
		if err := file.Close(); err != nil {
			__antithesis_instrumentation__.Notify(182368)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(182369)
		}
		__antithesis_instrumentation__.Notify(182346)
		if err := os.Rename(file.Name(), outputFile); err != nil {
			__antithesis_instrumentation__.Notify(182370)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(182371)
		}
		__antithesis_instrumentation__.Notify(182347)

		mu.Lock()
		outputFiles = append(outputFiles, outputFile)
		mu.Unlock()
		return nil, nil
	})
	__antithesis_instrumentation__.Notify(182326)

	for _, s := range outputFiles {
		__antithesis_instrumentation__.Notify(182372)
		l.Printf("Created %s", s)
	}
	__antithesis_instrumentation__.Notify(182327)

	if err != nil {
		__antithesis_instrumentation__.Notify(182373)
		sort.Slice(failed, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(182376)
			return failed[i].Index < failed[j].Index
		})
		__antithesis_instrumentation__.Notify(182374)
		for _, f := range failed {
			__antithesis_instrumentation__.Notify(182377)
			fmt.Fprintf(l.Stderr, "%d: %+v: %s\n", f.Index, f.Err, f.Out)
		}
		__antithesis_instrumentation__.Notify(182375)
		exit.WithCode(exit.UnspecifiedError())
	} else {
		__antithesis_instrumentation__.Notify(182378)
	}
	__antithesis_instrumentation__.Notify(182328)

	if opts.Open {
		__antithesis_instrumentation__.Notify(182379)
		waitCommands := []*exec.Cmd{}
		for i, file := range outputFiles {
			__antithesis_instrumentation__.Notify(182381)
			port := opts.StartingPort + i
			cmd := exec.Command("go", "tool", "pprof",
				"-http", fmt.Sprintf(":%d", port),
				file)
			waitCommands = append(waitCommands, cmd)
			if err := cmd.Start(); err != nil {
				__antithesis_instrumentation__.Notify(182382)
				return err
			} else {
				__antithesis_instrumentation__.Notify(182383)
			}
		}
		__antithesis_instrumentation__.Notify(182380)

		for _, cmd := range waitCommands {
			__antithesis_instrumentation__.Notify(182384)
			err := cmd.Wait()
			if err != nil {
				__antithesis_instrumentation__.Notify(182385)
				return err
			} else {
				__antithesis_instrumentation__.Notify(182386)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(182387)
	}
	__antithesis_instrumentation__.Notify(182329)
	return nil
}

func Destroy(
	l *logger.Logger, destroyAllMine bool, destroyAllLocal bool, clusterNames ...string,
) error {
	__antithesis_instrumentation__.Notify(182388)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182392)
		return errors.Wrap(err, "problem loading clusters")
	} else {
		__antithesis_instrumentation__.Notify(182393)
	}
	__antithesis_instrumentation__.Notify(182389)

	var cld *cloud.Cloud

	switch {
	case destroyAllMine:
		__antithesis_instrumentation__.Notify(182394)
		if len(clusterNames) != 0 {
			__antithesis_instrumentation__.Notify(182402)
			return errors.New("--all-mine cannot be combined with cluster names")
		} else {
			__antithesis_instrumentation__.Notify(182403)
		}
		__antithesis_instrumentation__.Notify(182395)
		if destroyAllLocal {
			__antithesis_instrumentation__.Notify(182404)
			return errors.New("--all-mine cannot be combined with --all-local")
		} else {
			__antithesis_instrumentation__.Notify(182405)
		}
		__antithesis_instrumentation__.Notify(182396)
		destroyPattern, err := userClusterNameRegexp(l)
		if err != nil {
			__antithesis_instrumentation__.Notify(182406)
			return err
		} else {
			__antithesis_instrumentation__.Notify(182407)
		}
		__antithesis_instrumentation__.Notify(182397)
		cld, err = cloud.ListCloud(l)
		if err != nil {
			__antithesis_instrumentation__.Notify(182408)
			return err
		} else {
			__antithesis_instrumentation__.Notify(182409)
		}
		__antithesis_instrumentation__.Notify(182398)
		clusters := cld.Clusters.FilterByName(destroyPattern)
		clusterNames = clusters.Names()

	case destroyAllLocal:
		__antithesis_instrumentation__.Notify(182399)
		if len(clusterNames) != 0 {
			__antithesis_instrumentation__.Notify(182410)
			return errors.New("--all-local cannot be combined with cluster names")
		} else {
			__antithesis_instrumentation__.Notify(182411)
		}
		__antithesis_instrumentation__.Notify(182400)

		clusterNames = local.Clusters()

	default:
		__antithesis_instrumentation__.Notify(182401)
		if len(clusterNames) == 0 {
			__antithesis_instrumentation__.Notify(182412)
			return errors.New("no cluster name provided")
		} else {
			__antithesis_instrumentation__.Notify(182413)
		}
	}
	__antithesis_instrumentation__.Notify(182390)

	if err := ctxgroup.GroupWorkers(
		context.TODO(),
		len(clusterNames),
		func(ctx context.Context, idx int) error {
			__antithesis_instrumentation__.Notify(182414)
			name := clusterNames[idx]
			if config.IsLocalClusterName(name) {
				__antithesis_instrumentation__.Notify(182417)
				return destroyLocalCluster(ctx, l, name)
			} else {
				__antithesis_instrumentation__.Notify(182418)
			}
			__antithesis_instrumentation__.Notify(182415)
			if cld == nil {
				__antithesis_instrumentation__.Notify(182419)
				var err error
				cld, err = cloud.ListCloud(l)
				if err != nil {
					__antithesis_instrumentation__.Notify(182420)
					return err
				} else {
					__antithesis_instrumentation__.Notify(182421)
				}
			} else {
				__antithesis_instrumentation__.Notify(182422)
			}
			__antithesis_instrumentation__.Notify(182416)
			return destroyCluster(cld, l, name)
		}); err != nil {
		__antithesis_instrumentation__.Notify(182423)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182424)
	}
	__antithesis_instrumentation__.Notify(182391)
	l.Printf("OK")
	return nil
}

func destroyCluster(cld *cloud.Cloud, l *logger.Logger, clusterName string) error {
	__antithesis_instrumentation__.Notify(182425)
	c, ok := cld.Clusters[clusterName]
	if !ok {
		__antithesis_instrumentation__.Notify(182427)
		return fmt.Errorf("cluster %s does not exist", clusterName)
	} else {
		__antithesis_instrumentation__.Notify(182428)
	}
	__antithesis_instrumentation__.Notify(182426)
	l.Printf("Destroying cluster %s with %d nodes", clusterName, len(c.VMs))
	return cloud.DestroyCluster(c)
}

func destroyLocalCluster(ctx context.Context, l *logger.Logger, clusterName string) error {
	__antithesis_instrumentation__.Notify(182429)
	if _, ok := readSyncedClusters(clusterName); !ok {
		__antithesis_instrumentation__.Notify(182433)
		return fmt.Errorf("cluster %s does not exist", clusterName)
	} else {
		__antithesis_instrumentation__.Notify(182434)
	}
	__antithesis_instrumentation__.Notify(182430)

	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182435)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182436)
	}
	__antithesis_instrumentation__.Notify(182431)
	if err := c.Wipe(ctx, l, false); err != nil {
		__antithesis_instrumentation__.Notify(182437)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182438)
	}
	__antithesis_instrumentation__.Notify(182432)
	return local.DeleteCluster(l, clusterName)
}

type ClusterAlreadyExistsError struct {
	name string
}

func (e *ClusterAlreadyExistsError) Error() string {
	__antithesis_instrumentation__.Notify(182439)
	return fmt.Sprintf("cluster %s already exists", e.name)
}

func cleanupFailedCreate(l *logger.Logger, clusterName string) error {
	__antithesis_instrumentation__.Notify(182440)
	cld, err := cloud.ListCloud(l)
	if err != nil {
		__antithesis_instrumentation__.Notify(182443)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182444)
	}
	__antithesis_instrumentation__.Notify(182441)
	c, ok := cld.Clusters[clusterName]
	if !ok {
		__antithesis_instrumentation__.Notify(182445)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(182446)
	}
	__antithesis_instrumentation__.Notify(182442)
	return cloud.DestroyCluster(c)
}

func Create(
	ctx context.Context,
	l *logger.Logger,
	username string,
	numNodes int,
	createVMOpts vm.CreateOpts,
	providerOptsContainer vm.ProviderOptionsContainer,
) (retErr error) {
	__antithesis_instrumentation__.Notify(182447)
	if numNodes <= 0 || func() bool {
		__antithesis_instrumentation__.Notify(182456)
		return numNodes >= 1000 == true
	}() == true {
		__antithesis_instrumentation__.Notify(182457)

		return fmt.Errorf("number of nodes must be in [1..999]")
	} else {
		__antithesis_instrumentation__.Notify(182458)
	}
	__antithesis_instrumentation__.Notify(182448)
	clusterName := createVMOpts.ClusterName
	if err := verifyClusterName(l, clusterName, username); err != nil {
		__antithesis_instrumentation__.Notify(182459)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182460)
	}
	__antithesis_instrumentation__.Notify(182449)

	isLocal := config.IsLocalClusterName(clusterName)
	if isLocal {
		__antithesis_instrumentation__.Notify(182461)

		unlockFn, err := acquireFilesystemLock()
		if err != nil {
			__antithesis_instrumentation__.Notify(182463)
			return err
		} else {
			__antithesis_instrumentation__.Notify(182464)
		}
		__antithesis_instrumentation__.Notify(182462)
		defer unlockFn()
	} else {
		__antithesis_instrumentation__.Notify(182465)
	}
	__antithesis_instrumentation__.Notify(182450)

	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182466)
		return errors.Wrap(err, "problem loading clusters")
	} else {
		__antithesis_instrumentation__.Notify(182467)
	}
	__antithesis_instrumentation__.Notify(182451)

	if !isLocal {
		__antithesis_instrumentation__.Notify(182468)
		cld, err := cloud.ListCloud(l)
		if err != nil {
			__antithesis_instrumentation__.Notify(182471)
			return err
		} else {
			__antithesis_instrumentation__.Notify(182472)
		}
		__antithesis_instrumentation__.Notify(182469)
		if _, ok := cld.Clusters[clusterName]; ok {
			__antithesis_instrumentation__.Notify(182473)
			return &ClusterAlreadyExistsError{name: clusterName}
		} else {
			__antithesis_instrumentation__.Notify(182474)
		}
		__antithesis_instrumentation__.Notify(182470)

		defer func() {
			__antithesis_instrumentation__.Notify(182475)
			if retErr == nil {
				__antithesis_instrumentation__.Notify(182477)
				return
			} else {
				__antithesis_instrumentation__.Notify(182478)
			}
			__antithesis_instrumentation__.Notify(182476)
			fmt.Fprintf(l.Stderr, "Cleaning up partially-created cluster (prev err: %s)\n", retErr)
			if err := cleanupFailedCreate(l, clusterName); err != nil {
				__antithesis_instrumentation__.Notify(182479)
				fmt.Fprintf(l.Stderr, "Error while cleaning up partially-created cluster: %s\n", err)
			} else {
				__antithesis_instrumentation__.Notify(182480)
				fmt.Fprintf(l.Stderr, "Cleaning up OK\n")
			}
		}()
	} else {
		__antithesis_instrumentation__.Notify(182481)
		if _, ok := readSyncedClusters(clusterName); ok {
			__antithesis_instrumentation__.Notify(182483)
			return &ClusterAlreadyExistsError{name: clusterName}
		} else {
			__antithesis_instrumentation__.Notify(182484)
		}
		__antithesis_instrumentation__.Notify(182482)

		createVMOpts.VMProviders = []string{local.ProviderName}
	}
	__antithesis_instrumentation__.Notify(182452)

	if createVMOpts.SSDOpts.FileSystem == vm.Zfs {
		__antithesis_instrumentation__.Notify(182485)
		for _, provider := range createVMOpts.VMProviders {
			__antithesis_instrumentation__.Notify(182486)
			if provider != gce.ProviderName {
				__antithesis_instrumentation__.Notify(182487)
				return fmt.Errorf(
					"creating a node with --filesystem=zfs is currently only supported on gce",
				)
			} else {
				__antithesis_instrumentation__.Notify(182488)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(182489)
	}
	__antithesis_instrumentation__.Notify(182453)

	l.Printf("Creating cluster %s with %d nodes", clusterName, numNodes)
	if createErr := cloud.CreateCluster(l, numNodes, createVMOpts, providerOptsContainer); createErr != nil {
		__antithesis_instrumentation__.Notify(182490)
		return createErr
	} else {
		__antithesis_instrumentation__.Notify(182491)
	}
	__antithesis_instrumentation__.Notify(182454)

	if config.IsLocalClusterName(clusterName) {
		__antithesis_instrumentation__.Notify(182492)

		return LoadClusters()
	} else {
		__antithesis_instrumentation__.Notify(182493)
	}
	__antithesis_instrumentation__.Notify(182455)
	return SetupSSH(ctx, l, clusterName)
}

func GC(l *logger.Logger, dryrun bool) error {
	__antithesis_instrumentation__.Notify(182494)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182497)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182498)
	}
	__antithesis_instrumentation__.Notify(182495)
	cld, err := cloud.ListCloud(l)
	if err == nil {
		__antithesis_instrumentation__.Notify(182499)

		err = cloud.GCClusters(l, cld, dryrun)
	} else {
		__antithesis_instrumentation__.Notify(182500)
	}
	__antithesis_instrumentation__.Notify(182496)
	otherErr := cloud.GCAWSKeyPairs(dryrun)
	return errors.CombineErrors(err, otherErr)
}

type LogsOpts struct {
	Dir, Filter, ProgramFilter string
	Interval                   time.Duration
	From, To                   time.Time
	Out                        io.Writer
}

func Logs(l *logger.Logger, clusterName, dest, username string, logsOpts LogsOpts) error {
	__antithesis_instrumentation__.Notify(182501)
	if err := LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(182504)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182505)
	}
	__antithesis_instrumentation__.Notify(182502)
	c, err := newCluster(l, clusterName)
	if err != nil {
		__antithesis_instrumentation__.Notify(182506)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182507)
	}
	__antithesis_instrumentation__.Notify(182503)
	return c.Logs(
		logsOpts.Dir, dest, username, logsOpts.Filter, logsOpts.ProgramFilter,
		logsOpts.Interval, logsOpts.From, logsOpts.To, logsOpts.Out,
	)
}

func StageURL(l *logger.Logger, applicationName, version, stageOS string) ([]*url.URL, error) {
	__antithesis_instrumentation__.Notify(182508)
	os := runtime.GOOS
	if stageOS != "" {
		__antithesis_instrumentation__.Notify(182511)
		os = stageOS
	} else {
		__antithesis_instrumentation__.Notify(182512)
	}
	__antithesis_instrumentation__.Notify(182509)
	urls, err := install.URLsForApplication(applicationName, version, os)
	if err != nil {
		__antithesis_instrumentation__.Notify(182513)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(182514)
	}
	__antithesis_instrumentation__.Notify(182510)
	return urls, nil
}

func InitProviders() map[string]string {
	__antithesis_instrumentation__.Notify(182515)
	providersState := make(map[string]string)

	if err := aws.Init(); err != nil {
		__antithesis_instrumentation__.Notify(182520)
		providersState[aws.ProviderName] = "Inactive - " + err.Error()
	} else {
		__antithesis_instrumentation__.Notify(182521)
		providersState[aws.ProviderName] = "Active"
	}
	__antithesis_instrumentation__.Notify(182516)

	if err := gce.Init(); err != nil {
		__antithesis_instrumentation__.Notify(182522)
		providersState[gce.ProviderName] = "Inactive - " + err.Error()
	} else {
		__antithesis_instrumentation__.Notify(182523)
		providersState[gce.ProviderName] = "Active"
	}
	__antithesis_instrumentation__.Notify(182517)

	if err := azure.Init(); err != nil {
		__antithesis_instrumentation__.Notify(182524)
		providersState[azure.ProviderName] = "Inactive - " + err.Error()
	} else {
		__antithesis_instrumentation__.Notify(182525)
		providersState[azure.ProviderName] = "Active"
	}
	__antithesis_instrumentation__.Notify(182518)

	if err := local.Init(localVMStorage{}); err != nil {
		__antithesis_instrumentation__.Notify(182526)
		providersState[local.ProviderName] = "Inactive - " + err.Error()
	} else {
		__antithesis_instrumentation__.Notify(182527)
		providersState[local.ProviderName] = "Active"
	}
	__antithesis_instrumentation__.Notify(182519)

	return providersState
}
