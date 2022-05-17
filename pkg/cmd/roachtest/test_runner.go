package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"archive/zip"
	"context"
	"fmt"
	"html"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/internal/issues"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/internal/team"
	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/petermattis/goid"
)

var errTestsFailed = fmt.Errorf("some tests failed")
var errClusterProvisioningFailed = fmt.Errorf("some clusters could not be created")

type testRunner struct {
	stopper *stop.Stopper

	buildVersion version.Version

	config struct {
		skipClusterValidationOnAttach bool

		skipClusterStopOnAttach bool
		skipClusterWipeOnAttach bool

		disableIssue bool
	}

	status struct {
		syncutil.Mutex
		running map[*testImpl]struct{}
		pass    map[*testImpl]struct{}
		fail    map[*testImpl]struct{}
		skip    map[*testImpl]struct{}
	}

	cr *clusterRegistry

	workersMu struct {
		syncutil.Mutex

		workers map[string]*workerStatus
	}

	work *workPool

	completedTestsMu struct {
		syncutil.Mutex

		completed []completedTestInfo
	}

	numClusterErrs int32
}

func newTestRunner(
	cr *clusterRegistry, stopper *stop.Stopper, buildVersion version.Version,
) *testRunner {
	__antithesis_instrumentation__.Notify(44784)
	r := &testRunner{
		stopper:      stopper,
		cr:           cr,
		buildVersion: buildVersion,
	}
	r.config.skipClusterWipeOnAttach = !clusterWipe
	r.config.disableIssue = disableIssue
	r.workersMu.workers = make(map[string]*workerStatus)
	return r
}

type clustersOpt struct {
	typ clusterType

	clusterName string

	clusterID string

	user string

	cpuQuota int

	keepClustersOnTestFailure bool
}

func (c clustersOpt) validate() error {
	__antithesis_instrumentation__.Notify(44785)
	if c.typ == localCluster {
		__antithesis_instrumentation__.Notify(44787)
		if c.clusterName != "" {
			__antithesis_instrumentation__.Notify(44789)
			return errors.New("clusterName cannot be set when typ=localCluster")
		} else {
			__antithesis_instrumentation__.Notify(44790)
		}
		__antithesis_instrumentation__.Notify(44788)
		if c.clusterID != "" {
			__antithesis_instrumentation__.Notify(44791)
			return errors.New("clusterID cannot be set when typ=localCluster")
		} else {
			__antithesis_instrumentation__.Notify(44792)
		}
	} else {
		__antithesis_instrumentation__.Notify(44793)
	}
	__antithesis_instrumentation__.Notify(44786)
	return nil
}

type testOpts struct {
	versionsBinaryOverride map[string]string
}

func (r *testRunner) Run(
	ctx context.Context,
	tests []registry.TestSpec,
	count int,
	parallelism int,
	clustersOpt clustersOpt,
	topt testOpts,
	lopt loggingOpt,
	clusterAllocator clusterAllocatorFn,
) error {
	__antithesis_instrumentation__.Notify(44794)

	if len(tests) == 0 {
		__antithesis_instrumentation__.Notify(44806)
		return fmt.Errorf("no test matched filters")
	} else {
		__antithesis_instrumentation__.Notify(44807)
	}
	__antithesis_instrumentation__.Notify(44795)

	hasDevLicense := config.CockroachDevLicense != ""
	for _, t := range tests {
		__antithesis_instrumentation__.Notify(44808)
		if t.RequiresLicense && func() bool {
			__antithesis_instrumentation__.Notify(44809)
			return !hasDevLicense == true
		}() == true {
			__antithesis_instrumentation__.Notify(44810)
			return fmt.Errorf("test %q requires an enterprise license, set COCKROACH_DEV_LICENSE", t.Name)
		} else {
			__antithesis_instrumentation__.Notify(44811)
		}
	}
	__antithesis_instrumentation__.Notify(44796)

	if err := clustersOpt.validate(); err != nil {
		__antithesis_instrumentation__.Notify(44812)
		return err
	} else {
		__antithesis_instrumentation__.Notify(44813)
	}
	__antithesis_instrumentation__.Notify(44797)
	if parallelism != 1 {
		__antithesis_instrumentation__.Notify(44814)
		if clustersOpt.clusterName != "" {
			__antithesis_instrumentation__.Notify(44816)
			return fmt.Errorf("--cluster incompatible with --parallelism. Use --parallelism=1")
		} else {
			__antithesis_instrumentation__.Notify(44817)
		}
		__antithesis_instrumentation__.Notify(44815)
		if clustersOpt.typ == localCluster {
			__antithesis_instrumentation__.Notify(44818)
			return fmt.Errorf("--local incompatible with --parallelism. Use --parallelism=1")
		} else {
			__antithesis_instrumentation__.Notify(44819)
		}
	} else {
		__antithesis_instrumentation__.Notify(44820)
	}
	__antithesis_instrumentation__.Notify(44798)

	if name := clustersOpt.clusterName; name != "" {
		__antithesis_instrumentation__.Notify(44821)

		spec := tests[0].Cluster
		spec.Lifetime = 0
		for i := 1; i < len(tests); i++ {
			__antithesis_instrumentation__.Notify(44822)
			spec2 := tests[i].Cluster
			spec2.Lifetime = 0
			if spec != spec2 {
				__antithesis_instrumentation__.Notify(44823)
				return errors.Errorf("cluster specified but found tests "+
					"with incompatible specs: %s (%s) - %s (%s)",
					tests[0].Name, spec, tests[i].Name, spec2,
				)
			} else {
				__antithesis_instrumentation__.Notify(44824)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(44825)
	}
	__antithesis_instrumentation__.Notify(44799)
	if clusterAllocator == nil {
		__antithesis_instrumentation__.Notify(44826)
		clusterAllocator = defaultClusterAllocator(r, clustersOpt, lopt)
	} else {
		__antithesis_instrumentation__.Notify(44827)
	}
	__antithesis_instrumentation__.Notify(44800)

	rand.Seed(timeutil.Now().UnixNano())

	n := len(tests)
	if n*count < parallelism {
		__antithesis_instrumentation__.Notify(44828)

		parallelism = n * count
	} else {
		__antithesis_instrumentation__.Notify(44829)
	}
	__antithesis_instrumentation__.Notify(44801)

	r.status.running = make(map[*testImpl]struct{})
	r.status.pass = make(map[*testImpl]struct{})
	r.status.fail = make(map[*testImpl]struct{})
	r.status.skip = make(map[*testImpl]struct{})

	r.work = newWorkPool(tests, count)
	errs := &workerErrors{}

	qp := quotapool.NewIntPool("cloud cpu", uint64(clustersOpt.cpuQuota))
	l := lopt.l
	var wg sync.WaitGroup

	for i := 0; i < parallelism; i++ {
		__antithesis_instrumentation__.Notify(44830)
		i := i
		wg.Add(1)
		if err := r.stopper.RunAsyncTask(ctx, "worker", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(44831)
			defer wg.Done()

			err := r.runWorker(
				ctx, fmt.Sprintf("w%d", i), r.work, qp,
				r.stopper.ShouldQuiesce(),
				clustersOpt.keepClustersOnTestFailure,
				lopt.artifactsDir, lopt.literalArtifactsDir, lopt.tee, lopt.stdout,
				clusterAllocator,
				topt,
				l,
			)

			if err != nil {
				__antithesis_instrumentation__.Notify(44832)

				msg := fmt.Sprintf("Worker %d returned with error. Quiescing. Error: %v", i, err)
				shout(ctx, l, lopt.stdout, msg)
				errs.AddErr(err)

				wg.Add(1)
				go func() {
					__antithesis_instrumentation__.Notify(44834)
					defer wg.Done()
					r.stopper.Stop(ctx)
				}()
				__antithesis_instrumentation__.Notify(44833)

				if qp != nil {
					__antithesis_instrumentation__.Notify(44835)
					qp.Close(msg)
				} else {
					__antithesis_instrumentation__.Notify(44836)
				}
			} else {
				__antithesis_instrumentation__.Notify(44837)
			}
		}); err != nil {
			__antithesis_instrumentation__.Notify(44838)
			wg.Done()
		} else {
			__antithesis_instrumentation__.Notify(44839)
		}
	}
	__antithesis_instrumentation__.Notify(44802)

	wg.Wait()
	r.cr.destroyAllClusters(ctx, l)

	if errs.Err() != nil {
		__antithesis_instrumentation__.Notify(44840)
		shout(ctx, l, lopt.stdout, "FAIL (err: %s)", errs.Err())
		return errs.Err()
	} else {
		__antithesis_instrumentation__.Notify(44841)
	}
	__antithesis_instrumentation__.Notify(44803)
	passFailLine := r.generateReport()
	shout(ctx, l, lopt.stdout, passFailLine)

	if r.numClusterErrs > 0 {
		__antithesis_instrumentation__.Notify(44842)
		shout(ctx, l, lopt.stdout, "%d clusters could not be created", r.numClusterErrs)
		return errClusterProvisioningFailed
	} else {
		__antithesis_instrumentation__.Notify(44843)
	}
	__antithesis_instrumentation__.Notify(44804)

	if len(r.status.fail) > 0 {
		__antithesis_instrumentation__.Notify(44844)
		return errTestsFailed
	} else {
		__antithesis_instrumentation__.Notify(44845)
	}
	__antithesis_instrumentation__.Notify(44805)
	return nil
}

func numConcurrentClusterCreations() int {
	__antithesis_instrumentation__.Notify(44846)
	var res int
	if cloud == "aws" {
		__antithesis_instrumentation__.Notify(44848)

		res = 1
	} else {
		__antithesis_instrumentation__.Notify(44849)
		res = 1000
	}
	__antithesis_instrumentation__.Notify(44847)
	return res
}

func defaultClusterAllocator(
	r *testRunner, clustersOpt clustersOpt, lopt loggingOpt,
) clusterAllocatorFn {
	__antithesis_instrumentation__.Notify(44850)
	clusterFactory := newClusterFactory(
		clustersOpt.user, clustersOpt.clusterID, lopt.artifactsDir, r.cr, numConcurrentClusterCreations())

	allocateCluster := func(
		ctx context.Context,
		t registry.TestSpec,
		alloc *quotapool.IntAlloc,
		artifactsDir string,
		wStatus *workerStatus,
	) (*clusterImpl, error) {
		__antithesis_instrumentation__.Notify(44852)
		wStatus.SetStatus("creating cluster")
		defer wStatus.SetStatus("")

		existingClusterName := clustersOpt.clusterName
		if existingClusterName != "" {
			__antithesis_instrumentation__.Notify(44854)

			logPath := filepath.Join(artifactsDir, runnerLogsDir, "cluster-create", existingClusterName+".log")
			clusterL, err := logger.RootLogger(logPath, lopt.tee)
			if err != nil {
				__antithesis_instrumentation__.Notify(44856)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(44857)
			}
			__antithesis_instrumentation__.Notify(44855)
			defer clusterL.Close()
			opt := attachOpt{
				skipValidation: r.config.skipClusterValidationOnAttach,
				skipStop:       r.config.skipClusterStopOnAttach,
				skipWipe:       r.config.skipClusterWipeOnAttach,
			}
			lopt.l.PrintfCtx(ctx, "Attaching to existing cluster %s for test %s", existingClusterName, t.Name)
			return attachToExistingCluster(ctx, existingClusterName, clusterL, t.Cluster, opt, r.cr)
		} else {
			__antithesis_instrumentation__.Notify(44858)
		}
		__antithesis_instrumentation__.Notify(44853)
		lopt.l.PrintfCtx(ctx, "Creating new cluster for test %s: %s", t.Name, t.Cluster)

		cfg := clusterConfig{
			spec:         t.Cluster,
			artifactsDir: artifactsDir,
			username:     clustersOpt.user,
			localCluster: clustersOpt.typ == localCluster,
			alloc:        alloc,
		}
		return clusterFactory.newCluster(ctx, cfg, wStatus.SetStatus, lopt.tee)
	}
	__antithesis_instrumentation__.Notify(44851)
	return allocateCluster
}

type clusterAllocatorFn func(
	ctx context.Context,
	t registry.TestSpec,
	alloc *quotapool.IntAlloc,
	artifactsDir string,
	wStatus *workerStatus,
) (*clusterImpl, error)

func (r *testRunner) runWorker(
	ctx context.Context,
	name string,
	work *workPool,
	qp *quotapool.IntPool,
	interrupt <-chan struct{},
	debug bool,
	artifactsRootDir string,
	literalArtifactsDir string,
	teeOpt logger.TeeOptType,
	stdout io.Writer,
	allocateCluster clusterAllocatorFn,
	topt testOpts,
	l *logger.Logger,
) error {
	__antithesis_instrumentation__.Notify(44859)
	ctx = logtags.AddTag(ctx, name, nil)
	wStatus := r.addWorker(ctx, name)
	defer func() {
		__antithesis_instrumentation__.Notify(44862)
		r.removeWorker(ctx, name)
	}()
	__antithesis_instrumentation__.Notify(44860)

	var c *clusterImpl

	defer func() {
		__antithesis_instrumentation__.Notify(44863)
		wStatus.SetTest(nil, testToRunRes{noWork: true})
		wStatus.SetStatus("worker done")
		wStatus.SetCluster(nil)

		if c == nil {
			__antithesis_instrumentation__.Notify(44865)
			l.PrintfCtx(ctx, "Worker exiting; no cluster to destroy.")
			return
		} else {
			__antithesis_instrumentation__.Notify(44866)
		}
		__antithesis_instrumentation__.Notify(44864)
		doDestroy := ctx.Err() == nil
		if doDestroy {
			__antithesis_instrumentation__.Notify(44867)
			l.PrintfCtx(ctx, "Worker exiting; destroying cluster.")
			c.Destroy(context.Background(), closeLogger, l)
		} else {
			__antithesis_instrumentation__.Notify(44868)
			l.PrintfCtx(ctx, "Worker exiting with canceled ctx. Not destroying cluster.")
		}
	}()
	__antithesis_instrumentation__.Notify(44861)

	for {
		__antithesis_instrumentation__.Notify(44869)
		select {
		case <-interrupt:
			__antithesis_instrumentation__.Notify(44881)
			l.ErrorfCtx(ctx, "worker detected interruption")
			return errors.Errorf("interrupted")
		default:
			__antithesis_instrumentation__.Notify(44882)
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(44883)

				return errors.Wrap(ctx.Err(), "worker ctx done")
			} else {
				__antithesis_instrumentation__.Notify(44884)
			}
		}
		__antithesis_instrumentation__.Notify(44870)

		if c != nil {
			__antithesis_instrumentation__.Notify(44885)
			if _, ok := c.spec.ReusePolicy.(spec.ReusePolicyNone); ok {
				__antithesis_instrumentation__.Notify(44886)
				wStatus.SetStatus("destroying cluster")

				c.Destroy(context.Background(), closeLogger, l)
				c = nil
			} else {
				__antithesis_instrumentation__.Notify(44887)
			}
		} else {
			__antithesis_instrumentation__.Notify(44888)
		}
		__antithesis_instrumentation__.Notify(44871)
		var testToRun testToRunRes
		var err error

		wStatus.SetTest(nil, testToRunRes{})
		wStatus.SetStatus("getting work")
		testToRun, err = r.getWork(
			ctx, work, qp, c, interrupt, l,
			getWorkCallbacks{
				onDestroy: func() {
					__antithesis_instrumentation__.Notify(44889)
					wStatus.SetCluster(nil)
				},
			})
		__antithesis_instrumentation__.Notify(44872)
		if err != nil {
			__antithesis_instrumentation__.Notify(44890)

			return err
		} else {
			__antithesis_instrumentation__.Notify(44891)
		}
		__antithesis_instrumentation__.Notify(44873)
		if testToRun.noWork {
			__antithesis_instrumentation__.Notify(44892)
			shout(ctx, l, stdout, "no work remaining; runWorker is bailing out...")
			return nil
		} else {
			__antithesis_instrumentation__.Notify(44893)
		}
		__antithesis_instrumentation__.Notify(44874)

		if c != nil && func() bool {
			__antithesis_instrumentation__.Notify(44894)
			return testToRun.canReuseCluster == true
		}() == true {
			__antithesis_instrumentation__.Notify(44895)
			err = func() error {
				__antithesis_instrumentation__.Notify(44897)
				l.PrintfCtx(ctx, "Using existing cluster: %s. Wiping", c.name)
				if err := c.WipeE(ctx, l); err != nil {
					__antithesis_instrumentation__.Notify(44901)
					return err
				} else {
					__antithesis_instrumentation__.Notify(44902)
				}
				__antithesis_instrumentation__.Notify(44898)
				if err := c.RunE(ctx, c.All(), "rm -rf "+perfArtifactsDir); err != nil {
					__antithesis_instrumentation__.Notify(44903)
					return errors.Wrapf(err, "failed to remove perf artifacts dir")
				} else {
					__antithesis_instrumentation__.Notify(44904)
				}
				__antithesis_instrumentation__.Notify(44899)
				if c.localCertsDir != "" {
					__antithesis_instrumentation__.Notify(44905)
					if err := os.RemoveAll(c.localCertsDir); err != nil {
						__antithesis_instrumentation__.Notify(44907)
						return errors.Wrapf(err,
							"failed to remove local certs in %s", c.localCertsDir)
					} else {
						__antithesis_instrumentation__.Notify(44908)
					}
					__antithesis_instrumentation__.Notify(44906)
					c.localCertsDir = ""
				} else {
					__antithesis_instrumentation__.Notify(44909)
				}
				__antithesis_instrumentation__.Notify(44900)

				c.spec = testToRun.spec.Cluster
				return nil
			}()
			__antithesis_instrumentation__.Notify(44896)
			if err != nil {
				__antithesis_instrumentation__.Notify(44910)

				shout(ctx, l, stdout, "Unable to reuse cluster: %s due to: %s. Will attempt to create a fresh one",
					c.Name(), err)
				atomic.AddInt32(&r.numClusterErrs, 1)

				testToRun.canReuseCluster = false
			} else {
				__antithesis_instrumentation__.Notify(44911)
			}
		} else {
			__antithesis_instrumentation__.Notify(44912)
		}
		__antithesis_instrumentation__.Notify(44875)
		var clusterCreateErr error

		if !testToRun.canReuseCluster {
			__antithesis_instrumentation__.Notify(44913)

			wStatus.SetTest(nil, testToRun)
			wStatus.SetStatus("creating cluster")
			c, clusterCreateErr = allocateCluster(ctx, testToRun.spec, testToRun.alloc, artifactsRootDir, wStatus)
			if clusterCreateErr != nil {
				__antithesis_instrumentation__.Notify(44914)
				atomic.AddInt32(&r.numClusterErrs, 1)
				shout(ctx, l, stdout, "Unable to create (or reuse) cluster for test %s due to: %s.",
					testToRun.spec.Name, clusterCreateErr)
			} else {
				__antithesis_instrumentation__.Notify(44915)
			}
		} else {
			__antithesis_instrumentation__.Notify(44916)
		}
		__antithesis_instrumentation__.Notify(44876)

		logPath := ""
		var artifactsDir string
		var artifactsSpec string
		if artifactsRootDir != "" {
			__antithesis_instrumentation__.Notify(44917)
			escapedTestName := teamCityNameEscape(testToRun.spec.Name)
			runSuffix := "run_" + strconv.Itoa(testToRun.runNum)

			artifactsDir = filepath.Join(filepath.Join(artifactsRootDir, escapedTestName), runSuffix)
			logPath = filepath.Join(artifactsDir, "test.log")

			artifactsSpec = fmt.Sprintf("%s/%s/** => %s/%s", filepath.Join(literalArtifactsDir, escapedTestName), runSuffix, escapedTestName, runSuffix)
		} else {
			__antithesis_instrumentation__.Notify(44918)
		}
		__antithesis_instrumentation__.Notify(44877)
		testL, err := logger.RootLogger(logPath, teeOpt)
		if err != nil {
			__antithesis_instrumentation__.Notify(44919)
			return err
		} else {
			__antithesis_instrumentation__.Notify(44920)
		}
		__antithesis_instrumentation__.Notify(44878)
		t := &testImpl{
			spec:                   &testToRun.spec,
			cockroach:              cockroach,
			deprecatedWorkload:     workload,
			buildVersion:           r.buildVersion,
			artifactsDir:           artifactsDir,
			artifactsSpec:          artifactsSpec,
			l:                      testL,
			versionsBinaryOverride: topt.versionsBinaryOverride,
		}

		l.PrintfCtx(ctx, "starting test: %s:%d", testToRun.spec.Name, testToRun.runNum)

		if clusterCreateErr != nil {
			__antithesis_instrumentation__.Notify(44921)

			oldName := t.spec.Name
			oldOwner := t.spec.Owner

			t.printAndFail(0, clusterCreateErr)
			issueOutput := "test %s was skipped due to %s"
			issueOutput = fmt.Sprintf(issueOutput, oldName, t.FailureMsg())

			t.spec.Name = "cluster_creation"
			t.spec.Owner = registry.OwnerDevInf
			r.maybePostGithubIssue(ctx, l, t, stdout, issueOutput)

			t.spec.Name = oldName
			t.spec.Owner = oldOwner
		} else {
			__antithesis_instrumentation__.Notify(44922)

			c.status("running test")
			c.setTest(t)

			encAtRest := encrypt.asBool()
			if encrypt.String() == "random" && func() bool {
				__antithesis_instrumentation__.Notify(44924)
				return !t.Spec().(*registry.TestSpec).EncryptAtRandom == true
			}() == true {
				__antithesis_instrumentation__.Notify(44925)

				encAtRest = false
			} else {
				__antithesis_instrumentation__.Notify(44926)
			}
			__antithesis_instrumentation__.Notify(44923)
			c.encAtRest = encAtRest

			wStatus.SetCluster(c)
			wStatus.SetTest(t, testToRun)
			wStatus.SetStatus("running test")

			err = r.runTest(ctx, t, testToRun.runNum, testToRun.runCount, c, stdout, testL)
		}
		__antithesis_instrumentation__.Notify(44879)

		if err != nil {
			__antithesis_instrumentation__.Notify(44927)
			shout(ctx, l, stdout, "test returned error: %s: %s", t.Name(), err)

			if !t.Failed() {
				__antithesis_instrumentation__.Notify(44928)
				t.printAndFail(0, err)
			} else {
				__antithesis_instrumentation__.Notify(44929)
			}
		} else {
			__antithesis_instrumentation__.Notify(44930)
			msg := "test passed: %s (run %d)"
			if t.Failed() {
				__antithesis_instrumentation__.Notify(44932)
				msg = "test failed: %s (run %d)"
			} else {
				__antithesis_instrumentation__.Notify(44933)
			}
			__antithesis_instrumentation__.Notify(44931)
			msg = fmt.Sprintf(msg, t.Name(), testToRun.runNum)
			l.PrintfCtx(ctx, msg)
		}
		__antithesis_instrumentation__.Notify(44880)
		testL.Close()
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(44934)
			return t.Failed() == true
		}() == true {
			__antithesis_instrumentation__.Notify(44935)
			failureMsg := fmt.Sprintf("%s (%d) - ", testToRun.spec.Name, testToRun.runNum)
			if err != nil {
				__antithesis_instrumentation__.Notify(44938)
				failureMsg += fmt.Sprintf("%+v", err)
			} else {
				__antithesis_instrumentation__.Notify(44939)
				failureMsg += t.FailureMsg()
			}
			__antithesis_instrumentation__.Notify(44936)
			if c != nil {
				__antithesis_instrumentation__.Notify(44940)
				if debug {
					__antithesis_instrumentation__.Notify(44941)

					c.Save(ctx, failureMsg, l)

					c = nil
				} else {
					__antithesis_instrumentation__.Notify(44942)

					l.PrintfCtx(ctx, "destroying cluster %s because: %s", c, failureMsg)
					c.Destroy(context.Background(), closeLogger, l)
					c = nil
				}
			} else {
				__antithesis_instrumentation__.Notify(44943)
			}
			__antithesis_instrumentation__.Notify(44937)
			if err != nil {
				__antithesis_instrumentation__.Notify(44944)

				return err
			} else {
				__antithesis_instrumentation__.Notify(44945)
			}
		} else {
			__antithesis_instrumentation__.Notify(44946)

			getPerfArtifacts(ctx, l, c, t)
		}
	}
}

func getPerfArtifacts(ctx context.Context, l *logger.Logger, c *clusterImpl, t test.Test) {
	__antithesis_instrumentation__.Notify(44947)
	g := ctxgroup.WithContext(ctx)
	fetchNode := func(node int) func(context.Context) error {
		__antithesis_instrumentation__.Notify(44950)
		return func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(44951)
			testCmd := `'PERF_ARTIFACTS="` + perfArtifactsDir + `"
if [[ -d "${PERF_ARTIFACTS}" ]]; then
    echo true
elif [[ -e "${PERF_ARTIFACTS}" ]]; then
    ls -la "${PERF_ARTIFACTS}"
    exit 1
else
    echo false
fi'`
			result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(node), "bash", "-c", testCmd)
			if err != nil {
				__antithesis_instrumentation__.Notify(44953)
				return errors.Wrapf(err, "failed to check for perf artifacts")
			} else {
				__antithesis_instrumentation__.Notify(44954)
			}
			__antithesis_instrumentation__.Notify(44952)
			out := strings.TrimSpace(result.Stdout)
			switch out {
			case "true":
				__antithesis_instrumentation__.Notify(44955)
				dst := fmt.Sprintf("%s/%d.%s", t.ArtifactsDir(), node, perfArtifactsDir)
				return c.Get(ctx, l, perfArtifactsDir, dst, c.Node(node))
			case "false":
				__antithesis_instrumentation__.Notify(44956)
				l.PrintfCtx(ctx, "no perf artifacts exist on node %v", c.Node(node))
				return nil
			default:
				__antithesis_instrumentation__.Notify(44957)
				return errors.Errorf("unexpected output when checking for perf artifacts: %s", out)
			}
		}
	}
	__antithesis_instrumentation__.Notify(44948)
	for _, i := range c.All() {
		__antithesis_instrumentation__.Notify(44958)
		g.GoCtx(fetchNode(i))
	}
	__antithesis_instrumentation__.Notify(44949)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(44959)
		l.PrintfCtx(ctx, "failed to get perf artifacts: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(44960)
	}
}

func allStacks() []byte {
	__antithesis_instrumentation__.Notify(44961)

	b := make([]byte, 5*(1<<20))
	return b[:runtime.Stack(b, true)]
}

func (r *testRunner) runTest(
	ctx context.Context,
	t *testImpl,
	runNum int,
	runCount int,
	c *clusterImpl,
	stdout io.Writer,
	l *logger.Logger,
) error {
	__antithesis_instrumentation__.Notify(44962)
	if t.Spec().(*registry.TestSpec).Skip != "" {
		__antithesis_instrumentation__.Notify(44972)
		return fmt.Errorf("can't run skipped test: %s: %s", t.Name(), t.Spec().(*registry.TestSpec).Skip)
	} else {
		__antithesis_instrumentation__.Notify(44973)
	}
	__antithesis_instrumentation__.Notify(44963)

	runID := t.Name()
	if runCount > 1 {
		__antithesis_instrumentation__.Notify(44974)
		runID += fmt.Sprintf("#%d", runNum)
	} else {
		__antithesis_instrumentation__.Notify(44975)
	}
	__antithesis_instrumentation__.Notify(44964)
	if teamCity {
		__antithesis_instrumentation__.Notify(44976)
		shout(ctx, l, stdout, "##teamcity[testStarted name='%s' flowId='%s']", t.Name(), runID)
	} else {
		__antithesis_instrumentation__.Notify(44977)
		shout(ctx, l, stdout, "=== RUN   %s", runID)
	}
	__antithesis_instrumentation__.Notify(44965)

	r.status.Lock()
	r.status.running[t] = struct{}{}
	r.status.Unlock()

	t.runner = callerName()
	t.runnerID = goid.Get()

	defer func() {
		__antithesis_instrumentation__.Notify(44978)
		t.end = timeutil.Now()

		if err := recover(); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(44983)
			return err != errTestFatal == true
		}() == true {
			__antithesis_instrumentation__.Notify(44984)
			t.mu.Lock()
			t.mu.failed = true
			t.mu.output = append(t.mu.output, t.decorate(0, fmt.Sprint(err))...)
			t.mu.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(44985)
		}
		__antithesis_instrumentation__.Notify(44979)

		t.mu.Lock()
		t.mu.done = true
		t.mu.Unlock()

		durationStr := fmt.Sprintf("%.2fs", t.duration().Seconds())
		if t.Failed() {
			__antithesis_instrumentation__.Notify(44986)
			t.mu.Lock()
			output := fmt.Sprintf("test artifacts and logs in: %s\n", t.ArtifactsDir()) + string(t.mu.output)
			t.mu.Unlock()

			if teamCity {
				__antithesis_instrumentation__.Notify(44988)
				shout(ctx, l, stdout, "##teamcity[testFailed name='%s' details='%s' flowId='%s']",
					t.Name(), teamCityEscape(output), runID)
			} else {
				__antithesis_instrumentation__.Notify(44989)
			}
			__antithesis_instrumentation__.Notify(44987)

			shout(ctx, l, stdout, "--- FAIL: %s (%s)\n%s", runID, durationStr, output)

			r.maybePostGithubIssue(ctx, l, t, stdout, output)
		} else {
			__antithesis_instrumentation__.Notify(44990)
			shout(ctx, l, stdout, "--- PASS: %s (%s)", runID, durationStr)

		}
		__antithesis_instrumentation__.Notify(44980)

		if teamCity {
			__antithesis_instrumentation__.Notify(44991)
			shout(ctx, l, stdout, "##teamcity[testFinished name='%s' flowId='%s']", t.Name(), runID)

			if err := zipArtifacts(t.ArtifactsDir()); err != nil {
				__antithesis_instrumentation__.Notify(44993)
				l.Printf("unable to zip artifacts: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(44994)
			}
			__antithesis_instrumentation__.Notify(44992)

			if t.artifactsSpec != "" {
				__antithesis_instrumentation__.Notify(44995)

				shout(ctx, l, stdout, "##teamcity[publishArtifacts '%s']", t.artifactsSpec)
			} else {
				__antithesis_instrumentation__.Notify(44996)
			}
		} else {
			__antithesis_instrumentation__.Notify(44997)
		}
		__antithesis_instrumentation__.Notify(44981)

		r.recordTestFinish(completedTestInfo{
			test:    t.Name(),
			run:     runNum,
			start:   t.start,
			end:     t.end,
			pass:    !t.Failed(),
			failure: t.FailureMsg(),
		})
		r.status.Lock()
		delete(r.status.running, t)

		if t.Spec().(*registry.TestSpec).Run != nil {
			__antithesis_instrumentation__.Notify(44998)
			if t.Failed() {
				__antithesis_instrumentation__.Notify(44999)
				r.status.fail[t] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(45000)
				if t.Spec().(*registry.TestSpec).Skip == "" {
					__antithesis_instrumentation__.Notify(45001)
					r.status.pass[t] = struct{}{}
				} else {
					__antithesis_instrumentation__.Notify(45002)
					r.status.skip[t] = struct{}{}
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(45003)
		}
		__antithesis_instrumentation__.Notify(44982)
		r.status.Unlock()
	}()
	__antithesis_instrumentation__.Notify(44966)

	t.start = timeutil.Now()

	timeout := 10 * time.Hour
	if d := t.Spec().(*registry.TestSpec).Timeout; d != 0 {
		__antithesis_instrumentation__.Notify(45004)
		timeout = d
	} else {
		__antithesis_instrumentation__.Notify(45005)
	}
	__antithesis_instrumentation__.Notify(44967)

	minExp := timeutil.Now().Add(timeout + time.Hour)
	if c.expiration.Before(minExp) {
		__antithesis_instrumentation__.Notify(45006)
		extend := minExp.Sub(c.expiration)
		l.PrintfCtx(ctx, "cluster needs to survive until %s, but has expiration: %s. Extending.",
			minExp, c.expiration)
		if err := c.Extend(ctx, extend, l); err != nil {
			__antithesis_instrumentation__.Notify(45007)
			return errors.Wrapf(err, "failed to extend cluster: %s", c.name)
		} else {
			__antithesis_instrumentation__.Notify(45008)
		}
	} else {
		__antithesis_instrumentation__.Notify(45009)
	}
	__antithesis_instrumentation__.Notify(44968)

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	t.mu.Lock()

	t.mu.cancel = cancel
	t.mu.Unlock()

	testReturnedCh := make(chan struct{})
	go func() {
		__antithesis_instrumentation__.Notify(45010)
		defer close(testReturnedCh)

		defer func() {
			__antithesis_instrumentation__.Notify(45012)

			if r := recover(); r != nil && func() bool {
				__antithesis_instrumentation__.Notify(45013)
				return r != errTestFatal == true
			}() == true {
				__antithesis_instrumentation__.Notify(45014)

				t.Errorf("test panicked: %v", r)
			} else {
				__antithesis_instrumentation__.Notify(45015)
			}
		}()
		__antithesis_instrumentation__.Notify(45011)

		t.Spec().(*registry.TestSpec).Run(runCtx, t, c)
	}()
	__antithesis_instrumentation__.Notify(44969)

	var timedOut bool

	select {
	case <-testReturnedCh:
		__antithesis_instrumentation__.Notify(45016)
		s := "success"
		if t.Failed() {
			__antithesis_instrumentation__.Notify(45019)
			s = "failure"
		} else {
			__antithesis_instrumentation__.Notify(45020)
		}
		__antithesis_instrumentation__.Notify(45017)
		t.L().Printf("tearing down after %s; see teardown.log", s)
	case <-time.After(timeout):
		__antithesis_instrumentation__.Notify(45018)

		t.L().Printf("test timed out after %s; check __stacks.log and CRDB logs for goroutine dumps", timeout)
		timedOut = true
	}
	__antithesis_instrumentation__.Notify(44970)

	teardownL, err := c.l.ChildLogger("teardown", logger.QuietStderr, logger.QuietStdout)
	if err != nil {
		__antithesis_instrumentation__.Notify(45021)
		return err
	} else {
		__antithesis_instrumentation__.Notify(45022)
	}
	__antithesis_instrumentation__.Notify(44971)
	l, c.l = teardownL, teardownL
	t.ReplaceL(teardownL)

	return r.teardownTest(ctx, t, c, timedOut)
}

func (r *testRunner) teardownTest(
	ctx context.Context, t *testImpl, c *clusterImpl, timedOut bool,
) error {
	__antithesis_instrumentation__.Notify(45023)

	artifactsCollectedCh := make(chan struct{})
	_ = r.stopper.RunAsyncTask(ctx, "collect-artifacts", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(45027)

		defer close(artifactsCollectedCh)
		if timedOut {
			__antithesis_instrumentation__.Notify(45029)

			const stacksFile = "__stacks"
			if cl, err := t.L().ChildLogger(stacksFile, logger.QuietStderr, logger.QuietStdout); err == nil {
				__antithesis_instrumentation__.Notify(45031)
				sl := allStacks()
				if c.Spec().NodeCount == 0 {
					__antithesis_instrumentation__.Notify(45033)
					sl = []byte("<elided during unit test>")
				} else {
					__antithesis_instrumentation__.Notify(45034)
				}
				__antithesis_instrumentation__.Notify(45032)
				cl.PrintfCtx(ctx, "all stacks:\n\n%s\n", sl)
				t.L().PrintfCtx(ctx, "dumped stacks to %s", stacksFile)
			} else {
				__antithesis_instrumentation__.Notify(45035)
			}
			__antithesis_instrumentation__.Notify(45030)

			args := option.DefaultStopOpts()
			args.RoachprodOpts.Sig = 3
			err := c.StopE(ctx, t.L(), args, c.All())
			t.L().PrintfCtx(ctx, "asked CRDB nodes to dump stacks; check their main (DEV) logs: %v", err)

			if c.Spec().NodeCount > 0 {
				__antithesis_instrumentation__.Notify(45036)
				time.Sleep(3 * time.Second)
			} else {
				__antithesis_instrumentation__.Notify(45037)
			}
		} else {
			__antithesis_instrumentation__.Notify(45038)
		}
		__antithesis_instrumentation__.Notify(45028)

		c.assertNoDeadNode(ctx, t)

		c.FailOnReplicaDivergence(ctx, t)

		if timedOut || func() bool {
			__antithesis_instrumentation__.Notify(45039)
			return t.Failed() == true
		}() == true {
			__antithesis_instrumentation__.Notify(45040)
			r.collectClusterArtifacts(ctx, c, t)
		} else {
			__antithesis_instrumentation__.Notify(45041)
		}
	})
	__antithesis_instrumentation__.Notify(45024)

	const artifactsCollectionTimeout = time.Hour
	select {
	case <-artifactsCollectedCh:
		__antithesis_instrumentation__.Notify(45042)
	case <-time.After(artifactsCollectionTimeout):
		__antithesis_instrumentation__.Notify(45043)

		t.L().Printf("giving up on artifacts collection after %s", artifactsCollectionTimeout)
	}
	__antithesis_instrumentation__.Notify(45025)

	if timedOut {
		__antithesis_instrumentation__.Notify(45044)

		_ = c.StopE(ctx, t.L(), option.DefaultStopOpts(), c.All())

		t.Errorf("test timed out (%s)", t.Spec().(*registry.TestSpec).Timeout)
	} else {
		__antithesis_instrumentation__.Notify(45045)
	}
	__antithesis_instrumentation__.Notify(45026)
	return nil
}

func (r *testRunner) shouldPostGithubIssue(t test.Test) bool {
	__antithesis_instrumentation__.Notify(45046)
	opts := issues.DefaultOptionsFromEnv()
	return !r.config.disableIssue && func() bool {
		__antithesis_instrumentation__.Notify(45047)
		return opts.CanPost() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(45048)
		return opts.IsReleaseBranch() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(45049)
		return t.Spec().(*registry.TestSpec).Run != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(45050)
		return t.Spec().(*registry.TestSpec).Cluster.NodeCount > 0 == true
	}() == true
}

func (r *testRunner) maybePostGithubIssue(
	ctx context.Context, l *logger.Logger, t test.Test, stdout io.Writer, output string,
) {
	__antithesis_instrumentation__.Notify(45051)
	if !r.shouldPostGithubIssue(t) {
		__antithesis_instrumentation__.Notify(45058)
		return
	} else {
		__antithesis_instrumentation__.Notify(45059)
	}
	__antithesis_instrumentation__.Notify(45052)

	teams, err := team.DefaultLoadTeams()
	if err != nil {
		__antithesis_instrumentation__.Notify(45060)
		t.Fatalf("could not load teams: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(45061)
	}
	__antithesis_instrumentation__.Notify(45053)

	var mention []string
	var projColID int
	if sl, ok := teams.GetAliasesForPurpose(ownerToAlias(t.Spec().(*registry.TestSpec).Owner), team.PurposeRoachtest); ok {
		__antithesis_instrumentation__.Notify(45062)
		for _, alias := range sl {
			__antithesis_instrumentation__.Notify(45064)
			mention = append(mention, "@"+string(alias))
		}
		__antithesis_instrumentation__.Notify(45063)
		projColID = teams[sl[0]].TriageColumnID
	} else {
		__antithesis_instrumentation__.Notify(45065)
	}
	__antithesis_instrumentation__.Notify(45054)

	branch := os.Getenv("TC_BUILD_BRANCH")
	if branch == "" {
		__antithesis_instrumentation__.Notify(45066)
		branch = "<unknown branch>"
	} else {
		__antithesis_instrumentation__.Notify(45067)
	}
	__antithesis_instrumentation__.Notify(45055)

	msg := fmt.Sprintf("The test failed on branch=%s, cloud=%s:\n%s",
		branch, t.Spec().(*registry.TestSpec).Cluster.Cloud, output)
	artifacts := fmt.Sprintf("/%s", t.Name())

	labels := []string{"O-roachtest"}
	if !t.Spec().(*registry.TestSpec).NonReleaseBlocker {
		__antithesis_instrumentation__.Notify(45068)
		labels = append(labels, "release-blocker")
	} else {
		__antithesis_instrumentation__.Notify(45069)
	}
	__antithesis_instrumentation__.Notify(45056)

	req := issues.PostRequest{
		MentionOnCreate: mention,
		ProjectColumnID: projColID,
		PackageName:     "roachtest",
		TestName:        t.Name(),
		Message:         msg,
		Artifacts:       artifacts,
		ExtraLabels:     labels,
		HelpCommand: func(renderer *issues.Renderer) {
			__antithesis_instrumentation__.Notify(45070)
			issues.HelpCommandAsLink(
				"roachtest README",
				"https://github.com/cockroachdb/cockroach/blob/master/pkg/cmd/roachtest/README.md",
			)(renderer)
			issues.HelpCommandAsLink(
				"How To Investigate (internal)",
				"https://cockroachlabs.atlassian.net/l/c/SSSBr8c7",
			)(renderer)
		},
	}
	__antithesis_instrumentation__.Notify(45057)
	if err := issues.Post(
		context.Background(),
		issues.UnitTestFormatter,
		req,
	); err != nil {
		__antithesis_instrumentation__.Notify(45071)
		shout(ctx, l, stdout, "failed to post issue: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(45072)
	}
}

func (r *testRunner) collectClusterArtifacts(ctx context.Context, c *clusterImpl, t test.Test) {
	__antithesis_instrumentation__.Notify(45073)

	t.L().PrintfCtx(ctx, "collecting cluster logs")

	if err := saveDiskUsageToLogsDir(ctx, c); err != nil {
		__antithesis_instrumentation__.Notify(45081)
		t.L().Printf("failed to fetch disk uage summary: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(45082)
	}
	__antithesis_instrumentation__.Notify(45074)
	if err := c.FetchLogs(ctx, t); err != nil {
		__antithesis_instrumentation__.Notify(45083)
		t.L().Printf("failed to download logs: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(45084)
	}
	__antithesis_instrumentation__.Notify(45075)
	if err := c.FetchDmesg(ctx, t); err != nil {
		__antithesis_instrumentation__.Notify(45085)
		t.L().Printf("failed to fetch dmesg: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(45086)
	}
	__antithesis_instrumentation__.Notify(45076)
	if err := c.FetchJournalctl(ctx, t); err != nil {
		__antithesis_instrumentation__.Notify(45087)
		t.L().Printf("failed to fetch journalctl: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(45088)
	}
	__antithesis_instrumentation__.Notify(45077)
	if err := c.FetchCores(ctx, t); err != nil {
		__antithesis_instrumentation__.Notify(45089)
		t.L().Printf("failed to fetch cores: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(45090)
	}
	__antithesis_instrumentation__.Notify(45078)
	if err := c.CopyRoachprodState(ctx); err != nil {
		__antithesis_instrumentation__.Notify(45091)
		t.L().Printf("failed to copy roachprod state: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(45092)
	}
	__antithesis_instrumentation__.Notify(45079)
	if err := c.FetchTimeseriesData(ctx, t); err != nil {
		__antithesis_instrumentation__.Notify(45093)
		t.L().Printf("failed to fetch timeseries data: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(45094)
	}
	__antithesis_instrumentation__.Notify(45080)
	if err := c.FetchDebugZip(ctx, t); err != nil {
		__antithesis_instrumentation__.Notify(45095)
		t.L().Printf("failed to collect zip: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(45096)
	}
}

func callerName() string {
	__antithesis_instrumentation__.Notify(45097)

	var pc [2]uintptr
	n := runtime.Callers(2, pc[:])
	if n == 0 {
		__antithesis_instrumentation__.Notify(45099)
		panic("zero callers found")
	} else {
		__antithesis_instrumentation__.Notify(45100)
	}
	__antithesis_instrumentation__.Notify(45098)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}

func (r *testRunner) generateReport() string {
	__antithesis_instrumentation__.Notify(45101)
	r.status.Lock()
	defer r.status.Unlock()
	postSlackReport(r.status.pass, r.status.fail, r.status.skip)

	fails := len(r.status.fail)
	var msg string
	if fails > 0 {
		__antithesis_instrumentation__.Notify(45103)
		msg = fmt.Sprintf("FAIL (%d fails)\n", fails)
	} else {
		__antithesis_instrumentation__.Notify(45104)
		msg = "PASS"
	}
	__antithesis_instrumentation__.Notify(45102)
	return msg
}

type getWorkCallbacks struct {
	onDestroy func()
}

func (r *testRunner) getWork(
	ctx context.Context,
	work *workPool,
	qp *quotapool.IntPool,
	c *clusterImpl,
	interrupt <-chan struct{},
	l *logger.Logger,
	callbacks getWorkCallbacks,
) (testToRunRes, error) {
	__antithesis_instrumentation__.Notify(45105)

	select {
	case <-interrupt:
		__antithesis_instrumentation__.Notify(45109)
		return testToRunRes{}, fmt.Errorf("interrupted")
	default:
		__antithesis_instrumentation__.Notify(45110)
	}
	__antithesis_instrumentation__.Notify(45106)

	testToRun, err := work.getTestToRun(ctx, c, qp, r.cr, callbacks.onDestroy, l)
	if err != nil {
		__antithesis_instrumentation__.Notify(45111)
		return testToRunRes{}, err
	} else {
		__antithesis_instrumentation__.Notify(45112)
	}
	__antithesis_instrumentation__.Notify(45107)
	if !testToRun.noWork {
		__antithesis_instrumentation__.Notify(45113)
		l.PrintfCtx(ctx, "Selected test: %s run: %d.", testToRun.spec.Name, testToRun.runNum)
	} else {
		__antithesis_instrumentation__.Notify(45114)

		return testToRun, nil
	}
	__antithesis_instrumentation__.Notify(45108)
	return testToRun, nil
}

func (r *testRunner) addWorker(ctx context.Context, name string) *workerStatus {
	__antithesis_instrumentation__.Notify(45115)
	r.workersMu.Lock()
	defer r.workersMu.Unlock()
	w := &workerStatus{name: name}
	if _, ok := r.workersMu.workers[name]; ok {
		__antithesis_instrumentation__.Notify(45117)
		log.Fatalf(ctx, "worker %q already exists", name)
	} else {
		__antithesis_instrumentation__.Notify(45118)
	}
	__antithesis_instrumentation__.Notify(45116)
	r.workersMu.workers[name] = w
	return w
}

func (r *testRunner) removeWorker(ctx context.Context, name string) {
	__antithesis_instrumentation__.Notify(45119)
	r.workersMu.Lock()
	delete(r.workersMu.workers, name)
	r.workersMu.Unlock()
}

func (r *testRunner) runHTTPServer(httpPort int, stdout io.Writer) error {
	__antithesis_instrumentation__.Notify(45120)
	http.HandleFunc("/", r.serveHTTP)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", httpPort))
	if err != nil {
		__antithesis_instrumentation__.Notify(45123)
		return err
	} else {
		__antithesis_instrumentation__.Notify(45124)
	}
	__antithesis_instrumentation__.Notify(45121)
	httpPort = listener.Addr().(*net.TCPAddr).Port
	go func() {
		__antithesis_instrumentation__.Notify(45125)
		if err := http.Serve(listener, nil); err != nil {
			__antithesis_instrumentation__.Notify(45126)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(45127)
		}
	}()
	__antithesis_instrumentation__.Notify(45122)
	fmt.Fprintf(stdout, "HTTP server listening on all network interfaces, port %d.\n", httpPort)
	return nil
}

func (r *testRunner) serveHTTP(wr http.ResponseWriter, req *http.Request) {
	__antithesis_instrumentation__.Notify(45128)
	fmt.Fprintf(wr, "<html><body>")
	fmt.Fprintf(wr, "<a href='debug/pprof'>pprof</a>")
	fmt.Fprintf(wr, "<p>")

	fmt.Fprintf(wr, "<h2>Workers:</h2>")
	fmt.Fprintf(wr, `<table border='1'>
	<tr><th>Worker</th>
	<th>Worker Status</th>
	<th>Test</th>
	<th>Cluster</th>
	<th>Cluster reused</th>
	<th>Test Status</th>
	</tr>`)
	r.workersMu.Lock()
	workers := make([]*workerStatus, len(r.workersMu.workers))
	i := 0
	for _, w := range r.workersMu.workers {
		__antithesis_instrumentation__.Notify(45135)
		workers[i] = w
		i++
	}
	__antithesis_instrumentation__.Notify(45129)
	r.workersMu.Unlock()
	sort.Slice(workers, func(i int, j int) bool {
		__antithesis_instrumentation__.Notify(45136)
		l := workers[i]
		r := workers[j]
		return strings.Compare(l.name, r.name) < 0
	})
	__antithesis_instrumentation__.Notify(45130)
	for _, w := range workers {
		__antithesis_instrumentation__.Notify(45137)
		var testName string
		ttr := w.TestToRun()
		clusterReused := ""
		if ttr.noWork {
			__antithesis_instrumentation__.Notify(45141)
			testName = "done"
		} else {
			__antithesis_instrumentation__.Notify(45142)
			if ttr.spec.Name == "" {
				__antithesis_instrumentation__.Notify(45143)
				testName = "N/A"
			} else {
				__antithesis_instrumentation__.Notify(45144)
				testName = fmt.Sprintf("%s (run %d)", ttr.spec.Name, ttr.runNum)
				if ttr.canReuseCluster {
					__antithesis_instrumentation__.Notify(45145)
					clusterReused = "yes"
				} else {
					__antithesis_instrumentation__.Notify(45146)
					clusterReused = "no"
				}
			}
		}
		__antithesis_instrumentation__.Notify(45138)
		var clusterName, clusterAdminUIAddr string
		if w.Cluster() != nil {
			__antithesis_instrumentation__.Notify(45147)
			clusterName = w.Cluster().name
			adminUIAddrs, err := w.Cluster().ExternalAdminUIAddr(req.Context(), w.Cluster().l, w.Cluster().Node(1))
			if err == nil {
				__antithesis_instrumentation__.Notify(45148)
				clusterAdminUIAddr = adminUIAddrs[0]
			} else {
				__antithesis_instrumentation__.Notify(45149)
			}
		} else {
			__antithesis_instrumentation__.Notify(45150)
		}
		__antithesis_instrumentation__.Notify(45139)
		t := w.Test()
		testStatus := "N/A"
		if t != nil {
			__antithesis_instrumentation__.Notify(45151)
			testStatus = t.GetStatus()
		} else {
			__antithesis_instrumentation__.Notify(45152)
		}
		__antithesis_instrumentation__.Notify(45140)
		fmt.Fprintf(wr, "<tr><td>%s</td><td>%s</td><td>%s</td><td><a href='//%s'>%s</a></td><td>%s</td><td>%s</td></tr>\n",
			w.name, w.Status(), testName, clusterAdminUIAddr, clusterName, clusterReused, testStatus)
	}
	__antithesis_instrumentation__.Notify(45131)
	fmt.Fprintf(wr, "</table>")

	fmt.Fprintf(wr, "<p>")
	fmt.Fprintf(wr, "<h2>Finished tests:</h2>")
	fmt.Fprintf(wr, `<table border='1'>
	<tr><th>Test</th>
	<th>Status</th>
	<th>Duration</th>
	</tr>`)
	for _, t := range r.getCompletedTests() {
		__antithesis_instrumentation__.Notify(45153)
		name := fmt.Sprintf("%s (run %d)", t.test, t.run)
		status := "PASS"
		if !t.pass {
			__antithesis_instrumentation__.Notify(45155)
			status = "FAIL " + strings.ReplaceAll(html.EscapeString(t.failure), "\n", "<br>")
		} else {
			__antithesis_instrumentation__.Notify(45156)
		}
		__antithesis_instrumentation__.Notify(45154)
		duration := fmt.Sprintf("%s (%s - %s)", t.end.Sub(t.start), t.start, t.end)
		fmt.Fprintf(wr, "<tr><td>%s</td><td>%s</td><td>%s</td><tr/>", name, status, duration)
	}
	__antithesis_instrumentation__.Notify(45132)
	fmt.Fprintf(wr, "</table>")

	fmt.Fprintf(wr, "<p>")
	fmt.Fprintf(wr, "<h2>Clusters left alive for further debugging "+
		"(if --debug was specified):</h2>")
	fmt.Fprintf(wr, `<table border='1'>
	<tr><th>Cluster</th>
	<th>Test</th>
	</tr>`)
	for _, c := range r.cr.savedClusters() {
		__antithesis_instrumentation__.Notify(45157)
		fmt.Fprintf(wr, "<tr><td>%s</td><td>%s</td><tr/>", c.name, c.savedMsg)
	}
	__antithesis_instrumentation__.Notify(45133)
	fmt.Fprintf(wr, "</table>")

	fmt.Fprintf(wr, "<p>")
	fmt.Fprintf(wr, "<h2>Tests left:</h2>")
	fmt.Fprintf(wr, `<table border='1'>
	<tr><th>Test</th>
	<th>Runs</th>
	</tr>`)
	for _, t := range r.work.workRemaining() {
		__antithesis_instrumentation__.Notify(45158)
		fmt.Fprintf(wr, "<tr><td>%s</td><td>%d</td><tr/>", t.spec.Name, t.count)
	}
	__antithesis_instrumentation__.Notify(45134)

	fmt.Fprintf(wr, "</body></html>")
}

func (r *testRunner) recordTestFinish(info completedTestInfo) {
	__antithesis_instrumentation__.Notify(45159)
	r.completedTestsMu.Lock()
	r.completedTestsMu.completed = append(r.completedTestsMu.completed, info)
	r.completedTestsMu.Unlock()
}

func (r *testRunner) getCompletedTests() []completedTestInfo {
	__antithesis_instrumentation__.Notify(45160)
	r.completedTestsMu.Lock()
	defer r.completedTestsMu.Unlock()
	res := make([]completedTestInfo, len(r.completedTestsMu.completed))
	copy(res, r.completedTestsMu.completed)
	return res
}

type completedTestInfo struct {
	test    string
	run     int
	start   time.Time
	end     time.Time
	pass    bool
	failure string
}

type workerErrors struct {
	mu struct {
		syncutil.Mutex
		errs []error
	}
}

func (we *workerErrors) AddErr(err error) {
	__antithesis_instrumentation__.Notify(45161)
	we.mu.Lock()
	defer we.mu.Unlock()
	we.mu.errs = append(we.mu.errs, err)
}

func (we *workerErrors) Err() error {
	__antithesis_instrumentation__.Notify(45162)
	we.mu.Lock()
	defer we.mu.Unlock()
	if len(we.mu.errs) == 0 {
		__antithesis_instrumentation__.Notify(45164)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(45165)
	}
	__antithesis_instrumentation__.Notify(45163)

	return we.mu.errs[0]
}

func zipArtifacts(path string) error {
	__antithesis_instrumentation__.Notify(45166)
	f, err := os.Create(filepath.Join(path, "artifacts.zip"))
	if err != nil {
		__antithesis_instrumentation__.Notify(45173)
		return err
	} else {
		__antithesis_instrumentation__.Notify(45174)
	}
	__antithesis_instrumentation__.Notify(45167)
	defer f.Close()
	z := zip.NewWriter(f)
	rel := func(targetpath string) string {
		__antithesis_instrumentation__.Notify(45175)
		relpath, err := filepath.Rel(path, targetpath)
		if err != nil {
			__antithesis_instrumentation__.Notify(45177)
			return targetpath
		} else {
			__antithesis_instrumentation__.Notify(45178)
		}
		__antithesis_instrumentation__.Notify(45176)
		return relpath
	}
	__antithesis_instrumentation__.Notify(45168)

	walk := func(visitor func(string, os.FileInfo) error) error {
		__antithesis_instrumentation__.Notify(45179)
		return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			__antithesis_instrumentation__.Notify(45180)
			if err != nil {
				__antithesis_instrumentation__.Notify(45183)
				return err
			} else {
				__antithesis_instrumentation__.Notify(45184)
			}
			__antithesis_instrumentation__.Notify(45181)
			if !info.IsDir() && func() bool {
				__antithesis_instrumentation__.Notify(45185)
				return strings.HasSuffix(path, ".zip") == true
			}() == true {
				__antithesis_instrumentation__.Notify(45186)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(45187)
			}
			__antithesis_instrumentation__.Notify(45182)
			return visitor(path, info)
		})
	}
	__antithesis_instrumentation__.Notify(45169)

	if err := walk(func(path string, info os.FileInfo) error {
		__antithesis_instrumentation__.Notify(45188)
		if info.IsDir() {
			__antithesis_instrumentation__.Notify(45193)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(45194)
		}
		__antithesis_instrumentation__.Notify(45189)
		w, err := z.Create(rel(path))
		if err != nil {
			__antithesis_instrumentation__.Notify(45195)
			return err
		} else {
			__antithesis_instrumentation__.Notify(45196)
		}
		__antithesis_instrumentation__.Notify(45190)
		r, err := os.Open(path)
		if err != nil {
			__antithesis_instrumentation__.Notify(45197)
			return err
		} else {
			__antithesis_instrumentation__.Notify(45198)
		}
		__antithesis_instrumentation__.Notify(45191)
		defer r.Close()
		if _, err := io.Copy(w, r); err != nil {
			__antithesis_instrumentation__.Notify(45199)
			return err
		} else {
			__antithesis_instrumentation__.Notify(45200)
		}
		__antithesis_instrumentation__.Notify(45192)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(45201)
		return err
	} else {
		__antithesis_instrumentation__.Notify(45202)
	}
	__antithesis_instrumentation__.Notify(45170)
	if err := z.Close(); err != nil {
		__antithesis_instrumentation__.Notify(45203)
		return err
	} else {
		__antithesis_instrumentation__.Notify(45204)
	}
	__antithesis_instrumentation__.Notify(45171)
	if err := f.Sync(); err != nil {
		__antithesis_instrumentation__.Notify(45205)
		return err
	} else {
		__antithesis_instrumentation__.Notify(45206)
	}
	__antithesis_instrumentation__.Notify(45172)

	root := path
	return walk(func(path string, info os.FileInfo) error {
		__antithesis_instrumentation__.Notify(45207)
		if path == root {
			__antithesis_instrumentation__.Notify(45211)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(45212)
		}
		__antithesis_instrumentation__.Notify(45208)
		if err := os.RemoveAll(path); err != nil {
			__antithesis_instrumentation__.Notify(45213)
			return err
		} else {
			__antithesis_instrumentation__.Notify(45214)
		}
		__antithesis_instrumentation__.Notify(45209)
		if info.IsDir() {
			__antithesis_instrumentation__.Notify(45215)
			return filepath.SkipDir
		} else {
			__antithesis_instrumentation__.Notify(45216)
		}
		__antithesis_instrumentation__.Notify(45210)
		return nil
	})
}
