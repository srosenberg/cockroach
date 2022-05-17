package roachprod

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachprod/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/local"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
)

type syncedClustersWithMutex struct {
	clusters cloud.Clusters
	mu       syncutil.Mutex
}

var syncedClusters syncedClustersWithMutex

func readSyncedClusters(key string) (*cloud.Cluster, bool) {
	__antithesis_instrumentation__.Notify(180314)
	syncedClusters.mu.Lock()
	defer syncedClusters.mu.Unlock()
	if cluster, ok := syncedClusters.clusters[key]; ok {
		__antithesis_instrumentation__.Notify(180316)
		return cluster, ok
	} else {
		__antithesis_instrumentation__.Notify(180317)
	}
	__antithesis_instrumentation__.Notify(180315)
	return nil, false
}

func InitDirs() error {
	__antithesis_instrumentation__.Notify(180318)
	cd := os.ExpandEnv(config.ClustersDir)
	if err := os.MkdirAll(cd, 0755); err != nil {
		__antithesis_instrumentation__.Notify(180320)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180321)
	}
	__antithesis_instrumentation__.Notify(180319)
	return os.MkdirAll(os.ExpandEnv(config.DefaultDebugDir), 0755)
}

func saveCluster(c *cloud.Cluster) error {
	__antithesis_instrumentation__.Notify(180322)
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c); err != nil {
		__antithesis_instrumentation__.Notify(180327)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180328)
	}
	__antithesis_instrumentation__.Notify(180323)

	filename := clusterFilename(c.Name)

	tmpFile, err := os.CreateTemp(os.ExpandEnv(config.ClustersDir), c.Name)
	if err != nil {
		__antithesis_instrumentation__.Notify(180329)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180330)
	}
	__antithesis_instrumentation__.Notify(180324)

	_, err = tmpFile.Write(b.Bytes())
	err = errors.CombineErrors(err, tmpFile.Sync())
	err = errors.CombineErrors(err, tmpFile.Close())
	if err == nil {
		__antithesis_instrumentation__.Notify(180331)
		err = os.Rename(tmpFile.Name(), filename)
	} else {
		__antithesis_instrumentation__.Notify(180332)
	}
	__antithesis_instrumentation__.Notify(180325)
	if err != nil {
		__antithesis_instrumentation__.Notify(180333)
		_ = os.Remove(tmpFile.Name())
		return err
	} else {
		__antithesis_instrumentation__.Notify(180334)
	}
	__antithesis_instrumentation__.Notify(180326)
	return nil
}

func loadCluster(name string) (*cloud.Cluster, error) {
	__antithesis_instrumentation__.Notify(180335)
	filename := clusterFilename(name)
	data, err := os.ReadFile(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(180339)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(180340)
	}
	__antithesis_instrumentation__.Notify(180336)
	c := &cloud.Cluster{}
	if err := json.Unmarshal(data, c); err != nil {
		__antithesis_instrumentation__.Notify(180341)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(180342)
	}
	__antithesis_instrumentation__.Notify(180337)
	if c.Name != name {
		__antithesis_instrumentation__.Notify(180343)
		return nil, errors.Errorf("name mismatch (%s vs %s)", name, c.Name)
	} else {
		__antithesis_instrumentation__.Notify(180344)
	}
	__antithesis_instrumentation__.Notify(180338)
	return c, nil
}

func shouldIgnoreCluster(c *cloud.Cluster) bool {
	__antithesis_instrumentation__.Notify(180345)
	for i := range c.VMs {
		__antithesis_instrumentation__.Notify(180347)
		provider, ok := vm.Providers[c.VMs[i].Provider]
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(180348)
			return !provider.ProjectActive(c.VMs[i].Project) == true
		}() == true {
			__antithesis_instrumentation__.Notify(180349)
			return true
		} else {
			__antithesis_instrumentation__.Notify(180350)
		}
	}
	__antithesis_instrumentation__.Notify(180346)
	return false
}

func LoadClusters() error {
	__antithesis_instrumentation__.Notify(180351)
	syncedClusters.mu.Lock()
	defer syncedClusters.mu.Unlock()
	syncedClusters.clusters = make(cloud.Clusters)

	clusterNames, err := listClustersInCache()
	if err != nil {
		__antithesis_instrumentation__.Notify(180354)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180355)
	}
	__antithesis_instrumentation__.Notify(180352)

	for _, name := range clusterNames {
		__antithesis_instrumentation__.Notify(180356)
		c, err := loadCluster(name)
		if err != nil {
			__antithesis_instrumentation__.Notify(180360)
			if oserror.IsNotExist(err) {
				__antithesis_instrumentation__.Notify(180362)

				continue
			} else {
				__antithesis_instrumentation__.Notify(180363)
			}
			__antithesis_instrumentation__.Notify(180361)
			return errors.Wrapf(err, "could not load info for cluster %s", name)
		} else {
			__antithesis_instrumentation__.Notify(180364)
		}
		__antithesis_instrumentation__.Notify(180357)

		if len(c.VMs) == 0 {
			__antithesis_instrumentation__.Notify(180365)
			return errors.Errorf("found no VMs in %s", clusterFilename(name))
		} else {
			__antithesis_instrumentation__.Notify(180366)
		}
		__antithesis_instrumentation__.Notify(180358)
		if shouldIgnoreCluster(c) {
			__antithesis_instrumentation__.Notify(180367)
			continue
		} else {
			__antithesis_instrumentation__.Notify(180368)
		}
		__antithesis_instrumentation__.Notify(180359)

		syncedClusters.clusters[c.Name] = c

		if config.IsLocalClusterName(c.Name) {
			__antithesis_instrumentation__.Notify(180369)

			local.AddCluster(c)
		} else {
			__antithesis_instrumentation__.Notify(180370)
		}
	}
	__antithesis_instrumentation__.Notify(180353)

	return nil
}

func syncClustersCache(cloud *cloud.Cloud) error {
	__antithesis_instrumentation__.Notify(180371)

	for _, c := range cloud.Clusters {
		__antithesis_instrumentation__.Notify(180375)
		if err := saveCluster(c); err != nil {
			__antithesis_instrumentation__.Notify(180376)
			return err
		} else {
			__antithesis_instrumentation__.Notify(180377)
		}
	}
	__antithesis_instrumentation__.Notify(180372)

	clusterNames, err := listClustersInCache()
	if err != nil {
		__antithesis_instrumentation__.Notify(180378)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180379)
	}
	__antithesis_instrumentation__.Notify(180373)
	for _, name := range clusterNames {
		__antithesis_instrumentation__.Notify(180380)
		if _, ok := cloud.Clusters[name]; !ok {
			__antithesis_instrumentation__.Notify(180381)

			c, err := loadCluster(name)
			if err != nil {
				__antithesis_instrumentation__.Notify(180383)
				return err
			} else {
				__antithesis_instrumentation__.Notify(180384)
			}
			__antithesis_instrumentation__.Notify(180382)
			if !shouldIgnoreCluster(c) {
				__antithesis_instrumentation__.Notify(180385)
				filename := clusterFilename(name)
				if err := os.Remove(filename); err != nil {
					__antithesis_instrumentation__.Notify(180386)
					log.Infof(context.Background(), "failed to remove file %s", filename)
				} else {
					__antithesis_instrumentation__.Notify(180387)
				}
			} else {
				__antithesis_instrumentation__.Notify(180388)
			}
		} else {
			__antithesis_instrumentation__.Notify(180389)
		}
	}
	__antithesis_instrumentation__.Notify(180374)

	return nil
}

func clusterFilename(name string) string {
	__antithesis_instrumentation__.Notify(180390)
	cd := os.ExpandEnv(config.ClustersDir)
	return path.Join(cd, name+".json")
}

func listClustersInCache() ([]string, error) {
	__antithesis_instrumentation__.Notify(180391)
	var result []string
	cd := os.ExpandEnv(config.ClustersDir)
	files, err := os.ReadDir(cd)
	if err != nil {
		__antithesis_instrumentation__.Notify(180394)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(180395)
	}
	__antithesis_instrumentation__.Notify(180392)
	for _, file := range files {
		__antithesis_instrumentation__.Notify(180396)
		if !file.Type().IsRegular() {
			__antithesis_instrumentation__.Notify(180399)
			continue
		} else {
			__antithesis_instrumentation__.Notify(180400)
		}
		__antithesis_instrumentation__.Notify(180397)
		if !strings.HasSuffix(file.Name(), ".json") {
			__antithesis_instrumentation__.Notify(180401)
			continue
		} else {
			__antithesis_instrumentation__.Notify(180402)
		}
		__antithesis_instrumentation__.Notify(180398)
		result = append(result, strings.TrimSuffix(file.Name(), ".json"))
	}
	__antithesis_instrumentation__.Notify(180393)
	return result, nil
}

type localVMStorage struct{}

var _ local.VMStorage = localVMStorage{}

func (localVMStorage) SaveCluster(cluster *cloud.Cluster) error {
	__antithesis_instrumentation__.Notify(180403)
	return saveCluster(cluster)
}

func (localVMStorage) DeleteCluster(name string) error {
	__antithesis_instrumentation__.Notify(180404)
	return os.Remove(clusterFilename(name))
}
