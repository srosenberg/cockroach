package tenant

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type tenantEntry struct {
	TenantID roachpb.TenantID

	ClusterName string

	RefreshDelay time.Duration

	initialized sync.Once

	initError error

	pods struct {
		syncutil.Mutex
		pods []*Pod
	}

	calls struct {
		syncutil.Mutex

		lastRefresh time.Time
	}
}

func (e *tenantEntry) Initialize(ctx context.Context, client DirectoryClient) error {
	__antithesis_instrumentation__.Notify(23084)

	e.initialized.Do(func() {
		__antithesis_instrumentation__.Notify(23086)
		tenantResp, err := client.GetTenant(ctx, &GetTenantRequest{TenantID: e.TenantID.ToUint64()})
		if err != nil {
			__antithesis_instrumentation__.Notify(23088)
			e.initError = err
			return
		} else {
			__antithesis_instrumentation__.Notify(23089)
		}
		__antithesis_instrumentation__.Notify(23087)

		e.ClusterName = tenantResp.ClusterName
	})
	__antithesis_instrumentation__.Notify(23085)

	return e.initError
}

func (e *tenantEntry) RefreshPods(ctx context.Context, client DirectoryClient) error {
	__antithesis_instrumentation__.Notify(23090)

	e.calls.Lock()
	defer e.calls.Unlock()

	if !e.canRefreshLocked() {
		__antithesis_instrumentation__.Notify(23092)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(23093)
	}
	__antithesis_instrumentation__.Notify(23091)

	log.Infof(ctx, "refreshing tenant %d pods", e.TenantID)

	_, err := e.fetchPodsLocked(ctx, client)
	return err
}

func (e *tenantEntry) AddPod(pod *Pod) bool {
	__antithesis_instrumentation__.Notify(23094)
	e.pods.Lock()
	defer e.pods.Unlock()

	for i, existing := range e.pods.pods {
		__antithesis_instrumentation__.Notify(23096)
		if existing.Addr == pod.Addr {
			__antithesis_instrumentation__.Notify(23097)

			pods := e.pods.pods
			e.pods.pods = make([]*Pod, len(pods))
			copy(e.pods.pods, pods)
			e.pods.pods[i] = pod
			return false
		} else {
			__antithesis_instrumentation__.Notify(23098)
		}
	}
	__antithesis_instrumentation__.Notify(23095)

	e.pods.pods = append(e.pods.pods, pod)
	return true
}

func (e *tenantEntry) UpdatePod(pod *Pod) bool {
	__antithesis_instrumentation__.Notify(23099)
	e.pods.Lock()
	defer e.pods.Unlock()

	for i, existing := range e.pods.pods {
		__antithesis_instrumentation__.Notify(23101)
		if existing.Addr == pod.Addr {
			__antithesis_instrumentation__.Notify(23102)

			pods := e.pods.pods
			e.pods.pods = make([]*Pod, len(pods))
			copy(e.pods.pods, pods)
			e.pods.pods[i] = pod
			return true
		} else {
			__antithesis_instrumentation__.Notify(23103)
		}
	}
	__antithesis_instrumentation__.Notify(23100)

	return false
}

func (e *tenantEntry) RemovePodByAddr(addr string) bool {
	__antithesis_instrumentation__.Notify(23104)
	e.pods.Lock()
	defer e.pods.Unlock()

	for i, existing := range e.pods.pods {
		__antithesis_instrumentation__.Notify(23106)
		if existing.Addr == addr {
			__antithesis_instrumentation__.Notify(23107)
			copy(e.pods.pods[i:], e.pods.pods[i+1:])
			e.pods.pods = e.pods.pods[:len(e.pods.pods)-1]
			return true
		} else {
			__antithesis_instrumentation__.Notify(23108)
		}
	}
	__antithesis_instrumentation__.Notify(23105)
	return false
}

func (e *tenantEntry) GetPods() []*Pod {
	__antithesis_instrumentation__.Notify(23109)
	e.pods.Lock()
	defer e.pods.Unlock()
	return e.pods.pods
}

func (e *tenantEntry) EnsureTenantPod(
	ctx context.Context, client DirectoryClient, errorIfNoPods bool,
) (pods []*Pod, err error) {
	__antithesis_instrumentation__.Notify(23110)
	const retryDelay = 100 * time.Millisecond

	e.calls.Lock()
	defer e.calls.Unlock()

	pods = e.GetPods()
	if len(pods) != 0 {
		__antithesis_instrumentation__.Notify(23113)
		return pods, nil
	} else {
		__antithesis_instrumentation__.Notify(23114)
	}
	__antithesis_instrumentation__.Notify(23111)

	for {
		__antithesis_instrumentation__.Notify(23115)

		if err = ctx.Err(); err != nil {
			__antithesis_instrumentation__.Notify(23121)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(23122)
		}
		__antithesis_instrumentation__.Notify(23116)

		_, err = client.EnsurePod(ctx, &EnsurePodRequest{e.TenantID.ToUint64()})
		if err != nil {
			__antithesis_instrumentation__.Notify(23123)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(23124)
		}
		__antithesis_instrumentation__.Notify(23117)

		pods, err = e.fetchPodsLocked(ctx, client)
		if err != nil {
			__antithesis_instrumentation__.Notify(23125)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(23126)
		}
		__antithesis_instrumentation__.Notify(23118)
		if len(pods) != 0 {
			__antithesis_instrumentation__.Notify(23127)
			log.Infof(ctx, "resumed tenant %d", e.TenantID)
			break
		} else {
			__antithesis_instrumentation__.Notify(23128)
		}
		__antithesis_instrumentation__.Notify(23119)

		if errorIfNoPods {
			__antithesis_instrumentation__.Notify(23129)
			return nil, fmt.Errorf("no pods available for tenant %s", e.TenantID)
		} else {
			__antithesis_instrumentation__.Notify(23130)
		}
		__antithesis_instrumentation__.Notify(23120)
		sleepContext(ctx, retryDelay)
	}
	__antithesis_instrumentation__.Notify(23112)

	return pods, nil
}

func (e *tenantEntry) fetchPodsLocked(
	ctx context.Context, client DirectoryClient,
) (tenantPods []*Pod, err error) {
	__antithesis_instrumentation__.Notify(23131)

	list, err := client.ListPods(ctx, &ListPodsRequest{e.TenantID.ToUint64()})
	if err != nil {
		__antithesis_instrumentation__.Notify(23134)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23135)
	}
	__antithesis_instrumentation__.Notify(23132)

	e.pods.Lock()
	defer e.pods.Unlock()
	e.pods.pods = list.Pods

	if len(e.pods.pods) != 0 {
		__antithesis_instrumentation__.Notify(23136)
		log.Infof(ctx, "fetched IP addresses: %v", e.pods.pods)
	} else {
		__antithesis_instrumentation__.Notify(23137)
	}
	__antithesis_instrumentation__.Notify(23133)

	return e.pods.pods, nil
}

func (e *tenantEntry) canRefreshLocked() bool {
	__antithesis_instrumentation__.Notify(23138)
	now := timeutil.Now()
	if now.Sub(e.calls.lastRefresh) < e.RefreshDelay {
		__antithesis_instrumentation__.Notify(23140)
		return false
	} else {
		__antithesis_instrumentation__.Notify(23141)
	}
	__antithesis_instrumentation__.Notify(23139)
	e.calls.lastRefresh = now
	return true
}
