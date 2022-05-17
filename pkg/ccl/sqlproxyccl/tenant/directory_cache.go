package tenant

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DirectoryCache interface {
	LookupTenantPods(ctx context.Context, tenantID roachpb.TenantID, clusterName string) ([]*Pod, error)

	TryLookupTenantPods(ctx context.Context, tenantID roachpb.TenantID) ([]*Pod, error)

	ReportFailure(ctx context.Context, tenantID roachpb.TenantID, addr string) error
}

type dirOptions struct {
	deterministic bool
	refreshDelay  time.Duration
	podWatcher    chan *Pod
}

type DirOption func(opts *dirOptions)

func RefreshDelay(delay time.Duration) func(opts *dirOptions) {
	__antithesis_instrumentation__.Notify(22954)
	return func(opts *dirOptions) {
		__antithesis_instrumentation__.Notify(22955)
		opts.refreshDelay = delay
	}
}

func PodWatcher(podWatcher chan *Pod) func(opts *dirOptions) {
	__antithesis_instrumentation__.Notify(22956)
	return func(opts *dirOptions) {
		__antithesis_instrumentation__.Notify(22957)
		opts.podWatcher = podWatcher
	}
}

type directoryCache struct {
	client DirectoryClient

	stopper *stop.Stopper

	options dirOptions

	mut struct {
		syncutil.Mutex

		tenants map[roachpb.TenantID]*tenantEntry
	}
}

var _ DirectoryCache = &directoryCache{}

func NewDirectoryCache(
	ctx context.Context, stopper *stop.Stopper, client DirectoryClient, opts ...DirOption,
) (DirectoryCache, error) {
	__antithesis_instrumentation__.Notify(22958)
	dir := &directoryCache{client: client, stopper: stopper}

	dir.mut.tenants = make(map[roachpb.TenantID]*tenantEntry)
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(22962)
		opt(&dir.options)
	}
	__antithesis_instrumentation__.Notify(22959)
	if dir.options.refreshDelay == 0 {
		__antithesis_instrumentation__.Notify(22963)

		dir.options.refreshDelay = 100 * time.Millisecond
	} else {
		__antithesis_instrumentation__.Notify(22964)
	}
	__antithesis_instrumentation__.Notify(22960)

	if err := dir.watchPods(ctx, stopper); err != nil {
		__antithesis_instrumentation__.Notify(22965)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(22966)
	}
	__antithesis_instrumentation__.Notify(22961)

	return dir, nil
}

func (d *directoryCache) LookupTenantPods(
	ctx context.Context, tenantID roachpb.TenantID, clusterName string,
) ([]*Pod, error) {
	__antithesis_instrumentation__.Notify(22967)

	entry, err := d.getEntry(ctx, tenantID, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(22972)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(22973)
	}
	__antithesis_instrumentation__.Notify(22968)

	if clusterName != "" && func() bool {
		__antithesis_instrumentation__.Notify(22974)
		return entry.ClusterName != "" == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(22975)
		return clusterName != entry.ClusterName == true
	}() == true {
		__antithesis_instrumentation__.Notify(22976)

		log.Errorf(ctx, "cluster name %s doesn't match expected %s", clusterName, entry.ClusterName)
		return nil, status.Errorf(codes.NotFound,
			"cluster name %s doesn't match expected %s", clusterName, entry.ClusterName)
	} else {
		__antithesis_instrumentation__.Notify(22977)
	}
	__antithesis_instrumentation__.Notify(22969)

	ctx, _ = d.stopper.WithCancelOnQuiesce(ctx)
	tenantPods := entry.GetPods()

	hasRunningPod := false
	for _, pod := range tenantPods {
		__antithesis_instrumentation__.Notify(22978)
		if pod.State == RUNNING {
			__antithesis_instrumentation__.Notify(22979)
			hasRunningPod = true
			break
		} else {
			__antithesis_instrumentation__.Notify(22980)
		}
	}
	__antithesis_instrumentation__.Notify(22970)
	if !hasRunningPod {
		__antithesis_instrumentation__.Notify(22981)

		var err error
		if tenantPods, err = entry.EnsureTenantPod(ctx, d.client, d.options.deterministic); err != nil {
			__antithesis_instrumentation__.Notify(22982)
			if status.Code(err) == codes.NotFound {
				__antithesis_instrumentation__.Notify(22984)
				d.deleteEntry(entry)
			} else {
				__antithesis_instrumentation__.Notify(22985)
			}
			__antithesis_instrumentation__.Notify(22983)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(22986)
		}
	} else {
		__antithesis_instrumentation__.Notify(22987)
	}
	__antithesis_instrumentation__.Notify(22971)
	return tenantPods, nil
}

func (d *directoryCache) TryLookupTenantPods(
	ctx context.Context, tenantID roachpb.TenantID,
) ([]*Pod, error) {
	__antithesis_instrumentation__.Notify(22988)

	entry, err := d.getEntry(ctx, tenantID, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(22991)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(22992)
	}
	__antithesis_instrumentation__.Notify(22989)

	if entry == nil {
		__antithesis_instrumentation__.Notify(22993)
		return nil, status.Errorf(
			codes.NotFound, "tenant %d not in directory cache", tenantID.ToUint64())
	} else {
		__antithesis_instrumentation__.Notify(22994)
	}
	__antithesis_instrumentation__.Notify(22990)

	return entry.GetPods(), nil
}

func (d *directoryCache) ReportFailure(
	ctx context.Context, tenantID roachpb.TenantID, addr string,
) error {
	__antithesis_instrumentation__.Notify(22995)
	entry, err := d.getEntry(ctx, tenantID, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(22997)
		return err
	} else {
		__antithesis_instrumentation__.Notify(22998)
		if entry == nil {
			__antithesis_instrumentation__.Notify(22999)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(23000)
		}
	}
	__antithesis_instrumentation__.Notify(22996)

	return entry.RefreshPods(ctx, d.client)
}

func (d *directoryCache) getEntry(
	ctx context.Context, tenantID roachpb.TenantID, allowCreate bool,
) (*tenantEntry, error) {
	__antithesis_instrumentation__.Notify(23001)
	entry := func() *tenantEntry {
		__antithesis_instrumentation__.Notify(23005)

		d.mut.Lock()
		defer d.mut.Unlock()

		entry, ok := d.mut.tenants[tenantID]
		if ok {
			__antithesis_instrumentation__.Notify(23008)

			return entry
		} else {
			__antithesis_instrumentation__.Notify(23009)
		}
		__antithesis_instrumentation__.Notify(23006)

		if !allowCreate {
			__antithesis_instrumentation__.Notify(23010)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(23011)
		}
		__antithesis_instrumentation__.Notify(23007)

		log.Infof(ctx, "creating directory entry for tenant %d", tenantID)
		entry = &tenantEntry{TenantID: tenantID, RefreshDelay: d.options.refreshDelay}
		d.mut.tenants[tenantID] = entry
		return entry
	}()
	__antithesis_instrumentation__.Notify(23002)

	if entry == nil {
		__antithesis_instrumentation__.Notify(23012)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(23013)
	}
	__antithesis_instrumentation__.Notify(23003)

	err := entry.Initialize(ctx, d.client)
	if err != nil {
		__antithesis_instrumentation__.Notify(23014)

		if d.deleteEntry(entry) {
			__antithesis_instrumentation__.Notify(23016)
			log.Infof(ctx, "error initializing tenant %d: %v", tenantID, err)
		} else {
			__antithesis_instrumentation__.Notify(23017)
		}
		__antithesis_instrumentation__.Notify(23015)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23018)
	}
	__antithesis_instrumentation__.Notify(23004)

	return entry, nil
}

func (d *directoryCache) deleteEntry(entry *tenantEntry) bool {
	__antithesis_instrumentation__.Notify(23019)

	d.mut.Lock()
	defer d.mut.Unlock()

	existing, ok := d.mut.tenants[entry.TenantID]
	if ok && func() bool {
		__antithesis_instrumentation__.Notify(23021)
		return entry == existing == true
	}() == true {
		__antithesis_instrumentation__.Notify(23022)
		delete(d.mut.tenants, entry.TenantID)
		return true
	} else {
		__antithesis_instrumentation__.Notify(23023)
	}
	__antithesis_instrumentation__.Notify(23020)

	return false
}

func (d *directoryCache) watchPods(ctx context.Context, stopper *stop.Stopper) error {
	__antithesis_instrumentation__.Notify(23024)
	req := WatchPodsRequest{}

	var waitInit sync.WaitGroup
	waitInit.Add(1)

	err := stopper.RunAsyncTask(ctx, "watch-pods-client", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(23027)
		var client Directory_WatchPodsClient
		var err error
		firstRun := true
		ctx, _ = stopper.WithCancelOnQuiesce(ctx)

		watchPodsErr := log.Every(10 * time.Second)
		recvErr := log.Every(10 * time.Second)

		for {
			__antithesis_instrumentation__.Notify(23028)
			if client == nil {
				__antithesis_instrumentation__.Notify(23032)
				client, err = d.client.WatchPods(ctx, &req)
				if firstRun {
					__antithesis_instrumentation__.Notify(23034)
					waitInit.Done()
					firstRun = false
				} else {
					__antithesis_instrumentation__.Notify(23035)
				}
				__antithesis_instrumentation__.Notify(23033)
				if err != nil {
					__antithesis_instrumentation__.Notify(23036)
					if grpcutil.IsContextCanceled(err) {
						__antithesis_instrumentation__.Notify(23039)
						break
					} else {
						__antithesis_instrumentation__.Notify(23040)
					}
					__antithesis_instrumentation__.Notify(23037)
					if watchPodsErr.ShouldLog() {
						__antithesis_instrumentation__.Notify(23041)
						log.Errorf(ctx, "err creating new watch pod client: %s", err)
					} else {
						__antithesis_instrumentation__.Notify(23042)
					}
					__antithesis_instrumentation__.Notify(23038)
					sleepContext(ctx, time.Second)
					continue
				} else {
					__antithesis_instrumentation__.Notify(23043)
					log.Info(ctx, "established watch on pods")
				}
			} else {
				__antithesis_instrumentation__.Notify(23044)
			}
			__antithesis_instrumentation__.Notify(23029)

			resp, err := client.Recv()
			if err != nil {
				__antithesis_instrumentation__.Notify(23045)
				if grpcutil.IsContextCanceled(err) {
					__antithesis_instrumentation__.Notify(23049)
					break
				} else {
					__antithesis_instrumentation__.Notify(23050)
				}
				__antithesis_instrumentation__.Notify(23046)
				if recvErr.ShouldLog() {
					__antithesis_instrumentation__.Notify(23051)
					log.Errorf(ctx, "err receiving stream events: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(23052)
				}
				__antithesis_instrumentation__.Notify(23047)

				if err != io.EOF {
					__antithesis_instrumentation__.Notify(23053)
					time.Sleep(time.Second)
				} else {
					__antithesis_instrumentation__.Notify(23054)
				}
				__antithesis_instrumentation__.Notify(23048)
				client = nil
				continue
			} else {
				__antithesis_instrumentation__.Notify(23055)
			}
			__antithesis_instrumentation__.Notify(23030)

			if d.options.podWatcher != nil {
				__antithesis_instrumentation__.Notify(23056)
				select {
				case d.options.podWatcher <- resp.Pod:
					__antithesis_instrumentation__.Notify(23057)
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(23058)
					return
				}
			} else {
				__antithesis_instrumentation__.Notify(23059)
			}
			__antithesis_instrumentation__.Notify(23031)

			d.updateTenantEntry(ctx, resp.Pod)
		}
	})
	__antithesis_instrumentation__.Notify(23025)
	if err != nil {
		__antithesis_instrumentation__.Notify(23060)
		return err
	} else {
		__antithesis_instrumentation__.Notify(23061)
	}
	__antithesis_instrumentation__.Notify(23026)

	waitInit.Wait()
	return err
}

func (d *directoryCache) updateTenantEntry(ctx context.Context, pod *Pod) {
	__antithesis_instrumentation__.Notify(23062)
	if pod.Addr == "" {
		__antithesis_instrumentation__.Notify(23065)

		return
	} else {
		__antithesis_instrumentation__.Notify(23066)
	}
	__antithesis_instrumentation__.Notify(23063)

	entry, err := d.getEntry(ctx, roachpb.MakeTenantID(pod.TenantID), true)
	if err != nil {
		__antithesis_instrumentation__.Notify(23067)
		if !grpcutil.IsContextCanceled(err) {
			__antithesis_instrumentation__.Notify(23069)

			log.Errorf(ctx, "ignoring error getting entry for tenant %d: %v", pod.TenantID, err)
		} else {
			__antithesis_instrumentation__.Notify(23070)
		}
		__antithesis_instrumentation__.Notify(23068)
		return
	} else {
		__antithesis_instrumentation__.Notify(23071)
	}
	__antithesis_instrumentation__.Notify(23064)

	switch pod.State {
	case RUNNING, DRAINING:
		__antithesis_instrumentation__.Notify(23072)

		if entry.AddPod(pod) {
			__antithesis_instrumentation__.Notify(23075)
			log.Infof(ctx, "added IP address %s with load %.3f for tenant %d", pod.Addr, pod.Load, pod.TenantID)
		} else {
			__antithesis_instrumentation__.Notify(23076)
			log.Infof(ctx, "updated IP address %s with load %.3f for tenant %d", pod.Addr, pod.Load, pod.TenantID)
		}

	case UNKNOWN:
		__antithesis_instrumentation__.Notify(23073)
		if entry.UpdatePod(pod) {
			__antithesis_instrumentation__.Notify(23077)
			log.Infof(ctx, "updated IP address %s with load %.3f for tenant %d", pod.Addr, pod.Load, pod.TenantID)
		} else {
			__antithesis_instrumentation__.Notify(23078)
		}
	default:
		__antithesis_instrumentation__.Notify(23074)

		if entry.RemovePodByAddr(pod.Addr) {
			__antithesis_instrumentation__.Notify(23079)
			log.Infof(ctx, "deleted IP address %s for tenant %d", pod.Addr, pod.TenantID)
		} else {
			__antithesis_instrumentation__.Notify(23080)
		}
	}
}

func sleepContext(ctx context.Context, delay time.Duration) {
	__antithesis_instrumentation__.Notify(23081)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(23082)
	case <-time.After(delay):
		__antithesis_instrumentation__.Notify(23083)
	}
}
