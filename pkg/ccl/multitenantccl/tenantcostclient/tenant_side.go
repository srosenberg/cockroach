package tenantcostclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvtenant"
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/multitenant/tenantcostmodel"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errorspb"
)

var TargetPeriodSetting = settings.RegisterDurationSetting(
	settings.TenantReadOnly,
	"tenant_cost_control_period",
	"target duration between token bucket requests from tenants (requires restart)",
	10*time.Second,
	checkDurationInRange(5*time.Second, 120*time.Second),
)

var CPUUsageAllowance = settings.RegisterDurationSetting(
	settings.TenantReadOnly,
	"tenant_cpu_usage_allowance",
	"this much CPU usage per second is considered background usage and "+
		"doesn't contribute to consumption; for example, if it is set to 10ms, "+
		"that corresponds to 1% of a CPU",
	10*time.Millisecond,
	checkDurationInRange(0, 1000*time.Millisecond),
)

func checkDurationInRange(min, max time.Duration) func(v time.Duration) error {
	__antithesis_instrumentation__.Notify(20005)
	return func(v time.Duration) error {
		__antithesis_instrumentation__.Notify(20006)
		if v < min || func() bool {
			__antithesis_instrumentation__.Notify(20008)
			return v > max == true
		}() == true {
			__antithesis_instrumentation__.Notify(20009)
			return errors.Errorf("value %s out of range (%s, %s)", v, min, max)
		} else {
			__antithesis_instrumentation__.Notify(20010)
		}
		__antithesis_instrumentation__.Notify(20007)
		return nil
	}
}

const mainLoopUpdateInterval = 1 * time.Second

const movingAvgFactor = 0.5

const notifyFraction = 0.1

const anticipation = time.Second

const consumptionReportingThreshold = 100

const extendedReportingPeriodFactor = 4

const bufferRUs = 5000

func newTenantSideCostController(
	st *cluster.Settings,
	tenantID roachpb.TenantID,
	provider kvtenant.TokenBucketProvider,
	timeSource timeutil.TimeSource,
	testInstr TestInstrumentation,
) (multitenant.TenantSideCostController, error) {
	__antithesis_instrumentation__.Notify(20011)
	if tenantID == roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(20013)
		return nil, errors.AssertionFailedf("cost controller can't be used for system tenant")
	} else {
		__antithesis_instrumentation__.Notify(20014)
	}
	__antithesis_instrumentation__.Notify(20012)
	c := &tenantSideCostController{
		timeSource:      timeSource,
		testInstr:       testInstr,
		settings:        st,
		tenantID:        tenantID,
		provider:        provider,
		responseChan:    make(chan *roachpb.TokenBucketResponse, 1),
		lowRUNotifyChan: make(chan struct{}, 1),
	}
	c.limiter.Init(timeSource, testInstr, c.lowRUNotifyChan)

	c.costCfg = tenantcostmodel.ConfigFromSettings(&st.SV)
	return c, nil
}

func NewTenantSideCostController(
	st *cluster.Settings, tenantID roachpb.TenantID, provider kvtenant.TokenBucketProvider,
) (multitenant.TenantSideCostController, error) {
	__antithesis_instrumentation__.Notify(20015)
	return newTenantSideCostController(
		st, tenantID, provider,
		timeutil.DefaultTimeSource{},
		nil,
	)
}

func TestingTenantSideCostController(
	st *cluster.Settings,
	tenantID roachpb.TenantID,
	provider kvtenant.TokenBucketProvider,
	timeSource timeutil.TimeSource,
	testInstr TestInstrumentation,
) (multitenant.TenantSideCostController, error) {
	__antithesis_instrumentation__.Notify(20016)
	return newTenantSideCostController(st, tenantID, provider, timeSource, testInstr)
}

func init() {
	server.NewTenantSideCostController = NewTenantSideCostController
}

type tenantSideCostController struct {
	timeSource           timeutil.TimeSource
	testInstr            TestInstrumentation
	settings             *cluster.Settings
	costCfg              tenantcostmodel.Config
	tenantID             roachpb.TenantID
	provider             kvtenant.TokenBucketProvider
	limiter              limiter
	stopper              *stop.Stopper
	instanceID           base.SQLInstanceID
	sessionID            sqlliveness.SessionID
	externalUsageFn      multitenant.ExternalUsageFn
	nextLiveInstanceIDFn multitenant.NextLiveInstanceIDFn

	mu struct {
		syncutil.Mutex

		consumption roachpb.TenantConsumption
	}

	lowRUNotifyChan chan struct{}

	responseChan chan *roachpb.TokenBucketResponse

	run struct {
		now time.Time

		externalUsage multitenant.ExternalUsage

		consumption roachpb.TenantConsumption

		requestSeqNum int64

		targetPeriod time.Duration

		initialRequestCompleted bool

		requestInProgress bool

		requestNeedsRetry bool

		lastRequestTime         time.Time
		lastReportedConsumption roachpb.TenantConsumption

		lastDeadline time.Time
		lastRate     float64

		setupNotificationTimer     timeutil.TimerI
		setupNotificationCh        <-chan time.Time
		setupNotificationThreshold tenantcostmodel.RU

		fallbackRate float64

		fallbackRateStart time.Time

		avgRUPerSec float64

		avgRUPerSecLastRU float64
	}
}

var _ multitenant.TenantSideCostController = (*tenantSideCostController)(nil)

func (c *tenantSideCostController) Start(
	ctx context.Context,
	stopper *stop.Stopper,
	instanceID base.SQLInstanceID,
	sessionID sqlliveness.SessionID,
	externalUsageFn multitenant.ExternalUsageFn,
	nextLiveInstanceIDFn multitenant.NextLiveInstanceIDFn,
) error {
	__antithesis_instrumentation__.Notify(20017)
	if instanceID == 0 {
		__antithesis_instrumentation__.Notify(20020)
		return errors.New("invalid SQLInstanceID")
	} else {
		__antithesis_instrumentation__.Notify(20021)
	}
	__antithesis_instrumentation__.Notify(20018)
	if sessionID == "" {
		__antithesis_instrumentation__.Notify(20022)
		return errors.New("invalid sqlliveness.SessionID")
	} else {
		__antithesis_instrumentation__.Notify(20023)
	}
	__antithesis_instrumentation__.Notify(20019)
	c.stopper = stopper
	c.instanceID = instanceID
	c.sessionID = sessionID
	c.externalUsageFn = externalUsageFn
	c.nextLiveInstanceIDFn = nextLiveInstanceIDFn
	return stopper.RunAsyncTask(ctx, "cost-controller", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20024)
		c.mainLoop(ctx)
	})
}

func (c *tenantSideCostController) initRunState(ctx context.Context) {
	__antithesis_instrumentation__.Notify(20025)
	c.run.targetPeriod = TargetPeriodSetting.Get(&c.settings.SV)

	now := c.timeSource.Now()
	c.run.now = now
	c.run.externalUsage = c.externalUsageFn(ctx)
	c.run.lastRequestTime = now
	c.run.avgRUPerSec = initialRUs / c.run.targetPeriod.Seconds()
	c.run.requestSeqNum = 1
}

func (c *tenantSideCostController) updateRunState(ctx context.Context) {
	__antithesis_instrumentation__.Notify(20026)
	c.run.targetPeriod = TargetPeriodSetting.Get(&c.settings.SV)

	newTime := c.timeSource.Now()
	newExternalUsage := c.externalUsageFn(ctx)

	deltaCPU := newExternalUsage.CPUSecs - c.run.externalUsage.CPUSecs

	if deltaTime := newTime.Sub(c.run.now); deltaTime > 0 {
		__antithesis_instrumentation__.Notify(20030)
		deltaCPU -= CPUUsageAllowance.Get(&c.settings.SV).Seconds() * deltaTime.Seconds()
	} else {
		__antithesis_instrumentation__.Notify(20031)
	}
	__antithesis_instrumentation__.Notify(20027)
	if deltaCPU < 0 {
		__antithesis_instrumentation__.Notify(20032)
		deltaCPU = 0
	} else {
		__antithesis_instrumentation__.Notify(20033)
	}
	__antithesis_instrumentation__.Notify(20028)
	ru := deltaCPU * float64(c.costCfg.PodCPUSecond)

	var deltaPGWireEgressBytes uint64
	if newExternalUsage.PGWireEgressBytes > c.run.externalUsage.PGWireEgressBytes {
		__antithesis_instrumentation__.Notify(20034)
		deltaPGWireEgressBytes = newExternalUsage.PGWireEgressBytes - c.run.externalUsage.PGWireEgressBytes
		ru += float64(deltaPGWireEgressBytes) * float64(c.costCfg.PGWireEgressByte)
	} else {
		__antithesis_instrumentation__.Notify(20035)
	}
	__antithesis_instrumentation__.Notify(20029)

	c.mu.Lock()
	c.mu.consumption.SQLPodsCPUSeconds += deltaCPU
	c.mu.consumption.PGWireEgressBytes += deltaPGWireEgressBytes
	c.mu.consumption.RU += ru
	newConsumption := c.mu.consumption
	c.mu.Unlock()

	c.run.now = newTime
	c.run.externalUsage = newExternalUsage
	c.run.consumption = newConsumption

	c.limiter.RemoveTokens(newTime, tenantcostmodel.RU(ru))
}

func (c *tenantSideCostController) updateAvgRUPerSec() {
	__antithesis_instrumentation__.Notify(20036)
	delta := c.run.consumption.RU - c.run.avgRUPerSecLastRU
	c.run.avgRUPerSec = movingAvgFactor*c.run.avgRUPerSec + (1-movingAvgFactor)*delta
	c.run.avgRUPerSecLastRU = c.run.consumption.RU
}

func (c *tenantSideCostController) shouldReportConsumption() bool {
	__antithesis_instrumentation__.Notify(20037)
	if c.run.requestInProgress {
		__antithesis_instrumentation__.Notify(20040)
		return false
	} else {
		__antithesis_instrumentation__.Notify(20041)
	}
	__antithesis_instrumentation__.Notify(20038)

	timeSinceLastRequest := c.run.now.Sub(c.run.lastRequestTime)
	if timeSinceLastRequest >= c.run.targetPeriod {
		__antithesis_instrumentation__.Notify(20042)
		consumptionToReport := c.run.consumption.RU - c.run.lastReportedConsumption.RU
		if consumptionToReport >= consumptionReportingThreshold {
			__antithesis_instrumentation__.Notify(20044)
			return true
		} else {
			__antithesis_instrumentation__.Notify(20045)
		}
		__antithesis_instrumentation__.Notify(20043)
		if timeSinceLastRequest >= extendedReportingPeriodFactor*c.run.targetPeriod {
			__antithesis_instrumentation__.Notify(20046)
			return true
		} else {
			__antithesis_instrumentation__.Notify(20047)
		}
	} else {
		__antithesis_instrumentation__.Notify(20048)
	}
	__antithesis_instrumentation__.Notify(20039)

	return false
}

func (c *tenantSideCostController) sendTokenBucketRequest(ctx context.Context) {
	__antithesis_instrumentation__.Notify(20049)
	deltaConsumption := c.run.consumption
	deltaConsumption.Sub(&c.run.lastReportedConsumption)

	var requested float64

	if !c.run.initialRequestCompleted {
		__antithesis_instrumentation__.Notify(20052)
		requested = initialRUs
	} else {
		__antithesis_instrumentation__.Notify(20053)

		requested = c.run.avgRUPerSec*c.run.targetPeriod.Seconds() + bufferRUs

		requested -= float64(c.limiter.AvailableTokens(c.run.now))
		if requested < 0 {
			__antithesis_instrumentation__.Notify(20054)

			requested = 0
		} else {
			__antithesis_instrumentation__.Notify(20055)
		}
	}
	__antithesis_instrumentation__.Notify(20050)

	req := roachpb.TokenBucketRequest{
		TenantID:                    c.tenantID.ToUint64(),
		InstanceID:                  uint32(c.instanceID),
		InstanceLease:               c.sessionID.UnsafeBytes(),
		NextLiveInstanceID:          uint32(c.nextLiveInstanceIDFn(ctx)),
		SeqNum:                      c.run.requestSeqNum,
		ConsumptionSinceLastRequest: deltaConsumption,
		RequestedRU:                 requested,
		TargetRequestPeriod:         c.run.targetPeriod,
	}
	c.run.requestSeqNum++

	c.run.lastRequestTime = c.run.now

	c.run.lastReportedConsumption = c.run.consumption
	c.run.requestInProgress = true

	ctx, _ = c.stopper.WithCancelOnQuiesce(ctx)
	err := c.stopper.RunAsyncTask(ctx, "token-bucket-request", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20056)
		if log.ExpensiveLogEnabled(ctx, 1) {
			__antithesis_instrumentation__.Notify(20059)
			log.Infof(ctx, "issuing TokenBucket: %s\n", req.String())
		} else {
			__antithesis_instrumentation__.Notify(20060)
		}
		__antithesis_instrumentation__.Notify(20057)
		resp, err := c.provider.TokenBucket(ctx, &req)
		if err != nil {
			__antithesis_instrumentation__.Notify(20061)

			if !errors.Is(err, context.Canceled) {
				__antithesis_instrumentation__.Notify(20063)
				log.Warningf(ctx, "TokenBucket RPC error: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(20064)
			}
			__antithesis_instrumentation__.Notify(20062)
			resp = nil
		} else {
			__antithesis_instrumentation__.Notify(20065)
			if (resp.Error != errorspb.EncodedError{}) {
				__antithesis_instrumentation__.Notify(20066)

				err := errors.DecodeError(ctx, resp.Error)
				log.Warningf(ctx, "TokenBucket error: %v", err)
				resp = nil
			} else {
				__antithesis_instrumentation__.Notify(20067)
			}
		}
		__antithesis_instrumentation__.Notify(20058)
		c.responseChan <- resp
	})
	__antithesis_instrumentation__.Notify(20051)
	if err != nil {
		__antithesis_instrumentation__.Notify(20068)

		c.responseChan <- nil
	} else {
		__antithesis_instrumentation__.Notify(20069)
	}
}

func (c *tenantSideCostController) handleTokenBucketResponse(
	ctx context.Context, resp *roachpb.TokenBucketResponse,
) {
	__antithesis_instrumentation__.Notify(20070)
	if log.ExpensiveLogEnabled(ctx, 1) {
		__antithesis_instrumentation__.Notify(20078)
		log.Infof(
			ctx, "TokenBucket response: %g RUs over %s (fallback rate %g)",
			resp.GrantedRU, resp.TrickleDuration, resp.FallbackRate,
		)
	} else {
		__antithesis_instrumentation__.Notify(20079)
	}
	__antithesis_instrumentation__.Notify(20071)
	c.run.fallbackRate = resp.FallbackRate

	if !c.run.initialRequestCompleted {
		__antithesis_instrumentation__.Notify(20080)
		c.run.initialRequestCompleted = true

		c.limiter.RemoveTokens(c.run.now, initialRUs)
	} else {
		__antithesis_instrumentation__.Notify(20081)
	}
	__antithesis_instrumentation__.Notify(20072)

	granted := resp.GrantedRU
	if granted == 0 {
		__antithesis_instrumentation__.Notify(20082)

		if !c.run.fallbackRateStart.IsZero() {
			__antithesis_instrumentation__.Notify(20084)
			c.sendTokenBucketRequest(ctx)
		} else {
			__antithesis_instrumentation__.Notify(20085)
		}
		__antithesis_instrumentation__.Notify(20083)
		return
	} else {
		__antithesis_instrumentation__.Notify(20086)
	}
	__antithesis_instrumentation__.Notify(20073)

	c.run.fallbackRateStart = time.Time{}

	if !c.run.lastDeadline.IsZero() {
		__antithesis_instrumentation__.Notify(20087)

		if since := c.run.lastDeadline.Sub(c.run.now); since > 0 {
			__antithesis_instrumentation__.Notify(20088)
			granted += c.run.lastRate * since.Seconds()
		} else {
			__antithesis_instrumentation__.Notify(20089)
		}
	} else {
		__antithesis_instrumentation__.Notify(20090)
	}
	__antithesis_instrumentation__.Notify(20074)

	if c.run.setupNotificationTimer != nil {
		__antithesis_instrumentation__.Notify(20091)
		c.run.setupNotificationTimer.Stop()
		c.run.setupNotificationTimer = nil
		c.run.setupNotificationCh = nil
	} else {
		__antithesis_instrumentation__.Notify(20092)
	}
	__antithesis_instrumentation__.Notify(20075)

	notifyThreshold := tenantcostmodel.RU(granted * notifyFraction)
	if notifyThreshold < bufferRUs {
		__antithesis_instrumentation__.Notify(20093)
		notifyThreshold = bufferRUs
	} else {
		__antithesis_instrumentation__.Notify(20094)
	}
	__antithesis_instrumentation__.Notify(20076)
	var cfg tokenBucketReconfigureArgs
	if resp.TrickleDuration == 0 {
		__antithesis_instrumentation__.Notify(20095)

		cfg.NewTokens = tenantcostmodel.RU(granted)

		cfg.NewRate = 0
		cfg.NotifyThreshold = notifyThreshold

		c.run.lastDeadline = time.Time{}
	} else {
		__antithesis_instrumentation__.Notify(20096)

		deadline := c.run.now.Add(resp.TrickleDuration)

		cfg.NewRate = tenantcostmodel.RU(granted / resp.TrickleDuration.Seconds())

		timerDuration := resp.TrickleDuration - anticipation
		if timerDuration <= 0 {
			__antithesis_instrumentation__.Notify(20098)
			timerDuration = (resp.TrickleDuration + 1) / 2
		} else {
			__antithesis_instrumentation__.Notify(20099)
		}
		__antithesis_instrumentation__.Notify(20097)

		c.run.setupNotificationTimer = c.timeSource.NewTimer()
		c.run.setupNotificationTimer.Reset(timerDuration)
		c.run.setupNotificationCh = c.run.setupNotificationTimer.Ch()
		c.run.setupNotificationThreshold = notifyThreshold

		c.run.lastDeadline = deadline
	}
	__antithesis_instrumentation__.Notify(20077)
	c.run.lastRate = float64(cfg.NewRate)
	c.limiter.Reconfigure(c.run.now, cfg)
}

func (c *tenantSideCostController) mainLoop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(20100)
	interval := mainLoopUpdateInterval

	if targetPeriod := TargetPeriodSetting.Get(&c.settings.SV); targetPeriod < interval {
		__antithesis_instrumentation__.Notify(20102)
		interval = targetPeriod
	} else {
		__antithesis_instrumentation__.Notify(20103)
	}
	__antithesis_instrumentation__.Notify(20101)
	ticker := c.timeSource.NewTicker(interval)
	defer ticker.Stop()
	tickerCh := ticker.Ch()

	c.initRunState(ctx)
	c.sendTokenBucketRequest(ctx)

	for {
		__antithesis_instrumentation__.Notify(20104)
		select {
		case <-tickerCh:
			__antithesis_instrumentation__.Notify(20105)
			c.updateRunState(ctx)
			c.updateAvgRUPerSec()

			if !c.run.fallbackRateStart.IsZero() && func() bool {
				__antithesis_instrumentation__.Notify(20113)
				return !c.run.now.Before(c.run.fallbackRateStart) == true
			}() == true {
				__antithesis_instrumentation__.Notify(20114)
				log.Infof(ctx, "switching to fallback rate %.10g", c.run.fallbackRate)
				c.limiter.Reconfigure(c.run.now, tokenBucketReconfigureArgs{
					NewRate: tenantcostmodel.RU(c.run.fallbackRate),
				})
				c.run.fallbackRateStart = time.Time{}
			} else {
				__antithesis_instrumentation__.Notify(20115)
			}
			__antithesis_instrumentation__.Notify(20106)
			if c.run.requestNeedsRetry || func() bool {
				__antithesis_instrumentation__.Notify(20116)
				return c.shouldReportConsumption() == true
			}() == true {
				__antithesis_instrumentation__.Notify(20117)
				c.run.requestNeedsRetry = false
				c.sendTokenBucketRequest(ctx)
			} else {
				__antithesis_instrumentation__.Notify(20118)
			}
			__antithesis_instrumentation__.Notify(20107)
			if c.testInstr != nil {
				__antithesis_instrumentation__.Notify(20119)
				c.testInstr.Event(c.run.now, TickProcessed)
			} else {
				__antithesis_instrumentation__.Notify(20120)
			}

		case resp := <-c.responseChan:
			__antithesis_instrumentation__.Notify(20108)
			c.run.requestInProgress = false
			if resp != nil {
				__antithesis_instrumentation__.Notify(20121)
				c.updateRunState(ctx)
				c.handleTokenBucketResponse(ctx, resp)
				if c.testInstr != nil {
					__antithesis_instrumentation__.Notify(20122)
					c.testInstr.Event(c.run.now, TokenBucketResponseProcessed)
				} else {
					__antithesis_instrumentation__.Notify(20123)
				}
			} else {
				__antithesis_instrumentation__.Notify(20124)

				c.run.requestNeedsRetry = true
			}

		case <-c.run.setupNotificationCh:
			__antithesis_instrumentation__.Notify(20109)
			c.run.setupNotificationTimer = nil
			c.run.setupNotificationCh = nil

			c.updateRunState(ctx)
			c.limiter.SetupNotification(c.run.now, c.run.setupNotificationThreshold)

		case <-c.lowRUNotifyChan:
			__antithesis_instrumentation__.Notify(20110)
			c.updateRunState(ctx)
			c.run.fallbackRateStart = c.run.now.Add(anticipation)

			if !c.run.requestInProgress {
				__antithesis_instrumentation__.Notify(20125)
				c.sendTokenBucketRequest(ctx)
			} else {
				__antithesis_instrumentation__.Notify(20126)
			}
			__antithesis_instrumentation__.Notify(20111)
			if c.testInstr != nil {
				__antithesis_instrumentation__.Notify(20127)
				c.testInstr.Event(c.run.now, LowRUNotification)
			} else {
				__antithesis_instrumentation__.Notify(20128)
			}

		case <-c.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(20112)
			c.limiter.Close()

			return
		}
	}
}

func (c *tenantSideCostController) OnRequestWait(
	ctx context.Context, info tenantcostmodel.RequestInfo,
) error {
	__antithesis_instrumentation__.Notify(20129)
	if multitenant.HasTenantCostControlExemption(ctx) {
		__antithesis_instrumentation__.Notify(20131)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(20132)
	}
	__antithesis_instrumentation__.Notify(20130)

	return c.limiter.Wait(ctx, c.costCfg.RequestCost(info))
}

func (c *tenantSideCostController) OnResponse(
	ctx context.Context, req tenantcostmodel.RequestInfo, resp tenantcostmodel.ResponseInfo,
) {
	__antithesis_instrumentation__.Notify(20133)
	if multitenant.HasTenantCostControlExemption(ctx) {
		__antithesis_instrumentation__.Notify(20136)
		return
	} else {
		__antithesis_instrumentation__.Notify(20137)
	}
	__antithesis_instrumentation__.Notify(20134)
	if resp.ReadBytes() > 0 {
		__antithesis_instrumentation__.Notify(20138)
		c.limiter.RemoveTokens(c.timeSource.Now(), c.costCfg.ResponseCost(resp))
	} else {
		__antithesis_instrumentation__.Notify(20139)
	}
	__antithesis_instrumentation__.Notify(20135)

	c.mu.Lock()
	defer c.mu.Unlock()

	if isWrite, writeBytes := req.IsWrite(); isWrite {
		__antithesis_instrumentation__.Notify(20140)
		c.mu.consumption.WriteRequests++
		c.mu.consumption.WriteBytes += uint64(writeBytes)
		c.mu.consumption.RU += float64(c.costCfg.KVWriteCost(writeBytes))
	} else {
		__antithesis_instrumentation__.Notify(20141)
		c.mu.consumption.ReadRequests++
		readBytes := resp.ReadBytes()
		c.mu.consumption.ReadBytes += uint64(readBytes)
		c.mu.consumption.RU += float64(c.costCfg.KVReadCost(readBytes))
	}
}
