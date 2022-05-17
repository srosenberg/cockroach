package balancer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/tenant"

func selectTenantPod(rand float32, pods []*tenant.Pod) *tenant.Pod {
	__antithesis_instrumentation__.Notify(21107)
	if len(pods) == 0 {
		__antithesis_instrumentation__.Notify(21112)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(21113)
	}
	__antithesis_instrumentation__.Notify(21108)

	if len(pods) == 1 {
		__antithesis_instrumentation__.Notify(21114)
		return pods[0]
	} else {
		__antithesis_instrumentation__.Notify(21115)
	}
	__antithesis_instrumentation__.Notify(21109)

	totalLoad := float32(0)
	for _, pod := range pods {
		__antithesis_instrumentation__.Notify(21116)
		totalLoad += 1 - pod.Load
	}
	__antithesis_instrumentation__.Notify(21110)

	totalLoad *= rand

	for _, pod := range pods {
		__antithesis_instrumentation__.Notify(21117)
		totalLoad -= 1 - pod.Load
		if totalLoad < 0 {
			__antithesis_instrumentation__.Notify(21118)
			return pod
		} else {
			__antithesis_instrumentation__.Notify(21119)
		}
	}
	__antithesis_instrumentation__.Notify(21111)

	return pods[len(pods)-1]
}
