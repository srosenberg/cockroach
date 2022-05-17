package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type workPool struct {
	count int
	mu    struct {
		syncutil.Mutex

		tests []testWithCount
	}
}

func newWorkPool(tests []registry.TestSpec, count int) *workPool {
	__antithesis_instrumentation__.Notify(52585)
	p := &workPool{count: count}
	for _, spec := range tests {
		__antithesis_instrumentation__.Notify(52587)
		p.mu.tests = append(p.mu.tests, testWithCount{spec: spec, count: count})
	}
	__antithesis_instrumentation__.Notify(52586)
	return p
}

type testToRunRes struct {
	noWork bool

	spec registry.TestSpec

	runCount int

	runNum int

	canReuseCluster bool

	alloc *quotapool.IntAlloc
}

func (p *workPool) workRemaining() []testWithCount {
	__antithesis_instrumentation__.Notify(52588)
	p.mu.Lock()
	defer p.mu.Unlock()
	res := make([]testWithCount, len(p.mu.tests))
	copy(res, p.mu.tests)
	return res
}

func (p *workPool) getTestToRun(
	ctx context.Context,
	c *clusterImpl,
	qp *quotapool.IntPool,
	cr *clusterRegistry,
	onDestroy func(),
	l *logger.Logger,
) (testToRunRes, error) {
	__antithesis_instrumentation__.Notify(52589)

	if c != nil {
		__antithesis_instrumentation__.Notify(52591)
		ttr := p.selectTestForCluster(ctx, c.spec, cr)
		if ttr.noWork {
			__antithesis_instrumentation__.Notify(52592)

			l.PrintfCtx(ctx,
				"No tests that can reuse cluster %s found (or there are no further tests to run). "+
					"Destroying.", c)
			c.Destroy(ctx, closeLogger, l)
			onDestroy()
		} else {
			__antithesis_instrumentation__.Notify(52593)
			return ttr, nil
		}
	} else {
		__antithesis_instrumentation__.Notify(52594)
	}
	__antithesis_instrumentation__.Notify(52590)

	return p.selectTest(ctx, qp)
}

func (p *workPool) selectTestForCluster(
	ctx context.Context, s spec.ClusterSpec, cr *clusterRegistry,
) testToRunRes {
	__antithesis_instrumentation__.Notify(52595)
	p.mu.Lock()
	defer p.mu.Unlock()
	testsWithCounts := p.findCompatibleTestsLocked(s)

	if len(testsWithCounts) == 0 {
		__antithesis_instrumentation__.Notify(52599)
		return testToRunRes{noWork: true}
	} else {
		__antithesis_instrumentation__.Notify(52600)
	}
	__antithesis_instrumentation__.Notify(52596)

	tag := ""
	if p, ok := s.ReusePolicy.(spec.ReusePolicyTagged); ok {
		__antithesis_instrumentation__.Notify(52601)
		tag = p.Tag
	} else {
		__antithesis_instrumentation__.Notify(52602)
	}
	__antithesis_instrumentation__.Notify(52597)

	candidateScore := 0
	var candidate testWithCount
	for _, tc := range testsWithCounts {
		__antithesis_instrumentation__.Notify(52603)
		score := scoreTestAgainstCluster(tc, tag, cr)
		if score > candidateScore {
			__antithesis_instrumentation__.Notify(52604)
			candidateScore = score
			candidate = tc
		} else {
			__antithesis_instrumentation__.Notify(52605)
		}
	}
	__antithesis_instrumentation__.Notify(52598)

	p.decTestLocked(ctx, candidate.spec.Name)

	runNum := p.count - candidate.count + 1
	return testToRunRes{
		spec:            candidate.spec,
		runCount:        p.count,
		runNum:          runNum,
		canReuseCluster: true,
	}
}

func (p *workPool) selectTest(ctx context.Context, qp *quotapool.IntPool) (testToRunRes, error) {
	__antithesis_instrumentation__.Notify(52606)
	var ttr testToRunRes
	alloc, err := qp.AcquireFunc(ctx, func(ctx context.Context, pi quotapool.PoolInfo) (uint64, error) {
		__antithesis_instrumentation__.Notify(52609)
		p.mu.Lock()
		defer p.mu.Unlock()

		if len(p.mu.tests) == 0 {
			__antithesis_instrumentation__.Notify(52613)
			ttr = testToRunRes{
				noWork: true,
			}
			return 0, nil
		} else {
			__antithesis_instrumentation__.Notify(52614)
		}
		__antithesis_instrumentation__.Notify(52610)

		candidateIdx := -1
		candidateCount := 0
		smallestTest := math.MaxInt64
		for i, t := range p.mu.tests {
			__antithesis_instrumentation__.Notify(52615)
			cpu := t.spec.Cluster.NodeCount * t.spec.Cluster.CPUs
			if cpu < smallestTest {
				__antithesis_instrumentation__.Notify(52618)
				smallestTest = cpu
			} else {
				__antithesis_instrumentation__.Notify(52619)
			}
			__antithesis_instrumentation__.Notify(52616)
			if uint64(cpu) > pi.Available {
				__antithesis_instrumentation__.Notify(52620)
				continue
			} else {
				__antithesis_instrumentation__.Notify(52621)
			}
			__antithesis_instrumentation__.Notify(52617)
			if t.count > candidateCount {
				__antithesis_instrumentation__.Notify(52622)
				candidateIdx = i
				candidateCount = t.count
			} else {
				__antithesis_instrumentation__.Notify(52623)
			}
		}
		__antithesis_instrumentation__.Notify(52611)

		if candidateIdx == -1 {
			__antithesis_instrumentation__.Notify(52624)
			if uint64(smallestTest) > pi.Capacity {
				__antithesis_instrumentation__.Notify(52626)
				return 0, fmt.Errorf("not enough CPU quota to run any of the remaining tests")
			} else {
				__antithesis_instrumentation__.Notify(52627)
			}
			__antithesis_instrumentation__.Notify(52625)

			return 0, quotapool.ErrNotEnoughQuota
		} else {
			__antithesis_instrumentation__.Notify(52628)
		}
		__antithesis_instrumentation__.Notify(52612)

		tc := p.mu.tests[candidateIdx]
		runNum := p.count - tc.count + 1
		p.decTestLocked(ctx, tc.spec.Name)
		ttr = testToRunRes{
			spec:            tc.spec,
			runCount:        p.count,
			runNum:          runNum,
			canReuseCluster: false,
		}
		cpu := tc.spec.Cluster.NodeCount * tc.spec.Cluster.CPUs
		return uint64(cpu), nil
	})
	__antithesis_instrumentation__.Notify(52607)
	if err != nil {
		__antithesis_instrumentation__.Notify(52629)
		return testToRunRes{}, err
	} else {
		__antithesis_instrumentation__.Notify(52630)
	}
	__antithesis_instrumentation__.Notify(52608)
	ttr.alloc = alloc
	return ttr, nil
}

func scoreTestAgainstCluster(tc testWithCount, tag string, cr *clusterRegistry) int {
	__antithesis_instrumentation__.Notify(52631)
	t := tc.spec
	testPolicy := t.Cluster.ReusePolicy
	if tag != "" && func() bool {
		__antithesis_instrumentation__.Notify(52634)
		return testPolicy != (spec.ReusePolicyTagged{Tag: tag}) == true
	}() == true {
		__antithesis_instrumentation__.Notify(52635)
		log.Fatalf(context.TODO(),
			"incompatible test and cluster. Cluster tag: %s. Test policy: %+v",
			tag, t.Cluster.ReusePolicy)
	} else {
		__antithesis_instrumentation__.Notify(52636)
	}
	__antithesis_instrumentation__.Notify(52632)
	score := 0
	if _, ok := testPolicy.(spec.ReusePolicyAny); ok {
		__antithesis_instrumentation__.Notify(52637)
		score = 1000000
	} else {
		__antithesis_instrumentation__.Notify(52638)
		if _, ok := testPolicy.(spec.ReusePolicyTagged); ok {
			__antithesis_instrumentation__.Notify(52639)
			score = 500000
			if tag == "" {
				__antithesis_instrumentation__.Notify(52640)

				score -= 1000 * cr.countForTag(tag)
			} else {
				__antithesis_instrumentation__.Notify(52641)
			}
		} else {
			__antithesis_instrumentation__.Notify(52642)
			score = 0
		}
	}
	__antithesis_instrumentation__.Notify(52633)

	score += tc.count

	return score
}

func (p *workPool) findCompatibleTestsLocked(clusterSpec spec.ClusterSpec) []testWithCount {
	__antithesis_instrumentation__.Notify(52643)
	if _, ok := clusterSpec.ReusePolicy.(spec.ReusePolicyNone); ok {
		__antithesis_instrumentation__.Notify(52646)
		panic("can't search for tests compatible with a ReuseNone policy")
	} else {
		__antithesis_instrumentation__.Notify(52647)
	}
	__antithesis_instrumentation__.Notify(52644)
	var tests []testWithCount
	for _, tc := range p.mu.tests {
		__antithesis_instrumentation__.Notify(52648)
		if spec.ClustersCompatible(clusterSpec, tc.spec.Cluster) {
			__antithesis_instrumentation__.Notify(52649)
			tests = append(tests, tc)
		} else {
			__antithesis_instrumentation__.Notify(52650)
		}
	}
	__antithesis_instrumentation__.Notify(52645)
	return tests
}

func (p *workPool) decTestLocked(ctx context.Context, name string) {
	__antithesis_instrumentation__.Notify(52651)
	idx := -1
	for idx = range p.mu.tests {
		__antithesis_instrumentation__.Notify(52654)
		if p.mu.tests[idx].spec.Name == name {
			__antithesis_instrumentation__.Notify(52655)
			break
		} else {
			__antithesis_instrumentation__.Notify(52656)
		}
	}
	__antithesis_instrumentation__.Notify(52652)
	if idx == -1 {
		__antithesis_instrumentation__.Notify(52657)
		log.Fatalf(ctx, "failed to find test: %s", name)
	} else {
		__antithesis_instrumentation__.Notify(52658)
	}
	__antithesis_instrumentation__.Notify(52653)
	tc := &p.mu.tests[idx]
	tc.count--
	if tc.count == 0 {
		__antithesis_instrumentation__.Notify(52659)

		p.mu.tests = append(p.mu.tests[:idx], p.mu.tests[idx+1:]...)
	} else {
		__antithesis_instrumentation__.Notify(52660)
	}
}
