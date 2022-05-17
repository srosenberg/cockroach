package flowinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

var errNoInboundStreamConnection = errors.New("no inbound stream connection")

func IsNoInboundStreamConnectionError(err error) bool {
	__antithesis_instrumentation__.Notify(491620)
	return errors.Is(err, errNoInboundStreamConnection)
}

var SettingFlowStreamTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.distsql.flow_stream_timeout",
	"amount of time incoming streams wait for a flow to be set up before erroring out",
	10*time.Second,
	settings.NonNegativeDuration,
)

const expectedConnectionTime time.Duration = 500 * time.Millisecond

type InboundStreamInfo struct {
	mu struct {
		syncutil.Mutex
		connected bool

		canceled bool

		finished bool
	}

	receiver InboundStreamHandler

	onFinish func()
}

func NewInboundStreamInfo(
	receiver InboundStreamHandler, waitGroup *sync.WaitGroup,
) *InboundStreamInfo {
	__antithesis_instrumentation__.Notify(491621)
	return &InboundStreamInfo{
		receiver: receiver,

		onFinish: waitGroup.Done,
	}
}

func (s *InboundStreamInfo) connect(
	flowID execinfrapb.FlowID, streamID execinfrapb.StreamID,
) error {
	__antithesis_instrumentation__.Notify(491622)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mu.connected {
		__antithesis_instrumentation__.Notify(491625)
		return errors.Errorf("flow %s: inbound stream %d already connected", flowID, streamID)
	} else {
		__antithesis_instrumentation__.Notify(491626)
	}
	__antithesis_instrumentation__.Notify(491623)
	if s.mu.canceled {
		__antithesis_instrumentation__.Notify(491627)
		return errors.Errorf("flow %s: inbound stream %d came too late", flowID, streamID)
	} else {
		__antithesis_instrumentation__.Notify(491628)
	}
	__antithesis_instrumentation__.Notify(491624)
	s.mu.connected = true
	return nil
}

func (s *InboundStreamInfo) disconnect() {
	__antithesis_instrumentation__.Notify(491629)
	s.mu.Lock()
	s.mu.connected = false
	s.mu.Unlock()
}

func (s *InboundStreamInfo) finishLocked() {
	__antithesis_instrumentation__.Notify(491630)
	if !s.mu.connected && func() bool {
		__antithesis_instrumentation__.Notify(491633)
		return !s.mu.canceled == true
	}() == true {
		__antithesis_instrumentation__.Notify(491634)
		panic("finishing inbound stream that didn't connect or time out")
	} else {
		__antithesis_instrumentation__.Notify(491635)
	}
	__antithesis_instrumentation__.Notify(491631)
	if s.mu.finished {
		__antithesis_instrumentation__.Notify(491636)
		panic("double finish")
	} else {
		__antithesis_instrumentation__.Notify(491637)
	}
	__antithesis_instrumentation__.Notify(491632)

	s.mu.finished = true
	s.onFinish()
}

func (s *InboundStreamInfo) cancelIfNotConnected() InboundStreamHandler {
	__antithesis_instrumentation__.Notify(491638)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mu.connected || func() bool {
		__antithesis_instrumentation__.Notify(491640)
		return s.mu.finished == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(491641)
		return s.mu.canceled == true
	}() == true {
		__antithesis_instrumentation__.Notify(491642)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(491643)
	}
	__antithesis_instrumentation__.Notify(491639)
	s.mu.canceled = true
	s.finishLocked()
	return s.receiver
}

func (s *InboundStreamInfo) finish() {
	__antithesis_instrumentation__.Notify(491644)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.finishLocked()
}

type flowEntry struct {
	waitCh chan struct{}

	refCount int

	flow *FlowBase

	inboundStreams map[execinfrapb.StreamID]*InboundStreamInfo

	streamTimer *time.Timer
}

type FlowRegistry struct {
	syncutil.Mutex

	flows map[execinfrapb.FlowID]*flowEntry

	draining bool

	flowDone *sync.Cond

	testingRunBeforeDrainSleep func()
}

func NewFlowRegistry() *FlowRegistry {
	__antithesis_instrumentation__.Notify(491645)
	fr := &FlowRegistry{flows: make(map[execinfrapb.FlowID]*flowEntry)}
	fr.flowDone = sync.NewCond(fr)
	return fr
}

func (fr *FlowRegistry) getEntryLocked(id execinfrapb.FlowID) *flowEntry {
	__antithesis_instrumentation__.Notify(491646)
	entry, ok := fr.flows[id]
	if !ok {
		__antithesis_instrumentation__.Notify(491648)
		entry = &flowEntry{}
		fr.flows[id] = entry
	} else {
		__antithesis_instrumentation__.Notify(491649)
	}
	__antithesis_instrumentation__.Notify(491647)
	return entry
}

func (fr *FlowRegistry) releaseEntryLocked(id execinfrapb.FlowID) {
	__antithesis_instrumentation__.Notify(491650)
	entry := fr.flows[id]
	if entry.refCount > 1 {
		__antithesis_instrumentation__.Notify(491651)
		entry.refCount--
	} else {
		__antithesis_instrumentation__.Notify(491652)
		if entry.refCount != 1 {
			__antithesis_instrumentation__.Notify(491654)
			panic(errors.AssertionFailedf("invalid refCount: %d", entry.refCount))
		} else {
			__antithesis_instrumentation__.Notify(491655)
		}
		__antithesis_instrumentation__.Notify(491653)
		delete(fr.flows, id)
		fr.flowDone.Signal()
	}
}

type flowRetryableError struct {
	cause error
}

func (e *flowRetryableError) Error() string {
	__antithesis_instrumentation__.Notify(491656)
	return fmt.Sprintf("flow retryable error: %+v", e.cause)
}

func IsFlowRetryableError(e error) bool {
	__antithesis_instrumentation__.Notify(491657)
	return errors.HasType(e, (*flowRetryableError)(nil))
}

func (fr *FlowRegistry) RegisterFlow(
	ctx context.Context,
	id execinfrapb.FlowID,
	f *FlowBase,
	inboundStreams map[execinfrapb.StreamID]*InboundStreamInfo,
	timeout time.Duration,
) (retErr error) {
	__antithesis_instrumentation__.Notify(491658)
	fr.Lock()
	defer fr.Unlock()
	defer func() {
		__antithesis_instrumentation__.Notify(491665)
		if retErr != nil {
			__antithesis_instrumentation__.Notify(491666)
			for _, stream := range inboundStreams {
				__antithesis_instrumentation__.Notify(491667)
				stream.onFinish()
			}
		} else {
			__antithesis_instrumentation__.Notify(491668)
		}
	}()
	__antithesis_instrumentation__.Notify(491659)

	draining := fr.draining
	if f.Cfg != nil {
		__antithesis_instrumentation__.Notify(491669)
		if knobs, ok := f.Cfg.TestingKnobs.Flowinfra.(*TestingKnobs); ok && func() bool {
			__antithesis_instrumentation__.Notify(491670)
			return knobs != nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(491671)
			return knobs.FlowRegistryDraining != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(491672)
			draining = knobs.FlowRegistryDraining()
		} else {
			__antithesis_instrumentation__.Notify(491673)
		}
	} else {
		__antithesis_instrumentation__.Notify(491674)
	}
	__antithesis_instrumentation__.Notify(491660)

	if draining {
		__antithesis_instrumentation__.Notify(491675)
		return &flowRetryableError{cause: errors.Errorf(
			"could not register flowID %d because the registry is draining",
			id,
		)}
	} else {
		__antithesis_instrumentation__.Notify(491676)
	}
	__antithesis_instrumentation__.Notify(491661)
	entry := fr.getEntryLocked(id)
	if entry.flow != nil {
		__antithesis_instrumentation__.Notify(491677)
		return errors.Errorf(
			"flow already registered: flowID: %s.\n"+
				"Current flow: %+v\nExisting flow: %+v",
			f.spec.FlowID, f.spec, entry.flow.spec)
	} else {
		__antithesis_instrumentation__.Notify(491678)
	}
	__antithesis_instrumentation__.Notify(491662)

	entry.refCount++
	entry.flow = f
	entry.inboundStreams = inboundStreams

	if entry.waitCh != nil {
		__antithesis_instrumentation__.Notify(491679)
		close(entry.waitCh)
	} else {
		__antithesis_instrumentation__.Notify(491680)
	}
	__antithesis_instrumentation__.Notify(491663)

	if len(inboundStreams) > 0 {
		__antithesis_instrumentation__.Notify(491681)

		entry.streamTimer = time.AfterFunc(timeout, func() {
			__antithesis_instrumentation__.Notify(491682)

			numTimedOutReceivers := fr.cancelPendingStreams(id, errNoInboundStreamConnection)
			if numTimedOutReceivers != 0 {
				__antithesis_instrumentation__.Notify(491683)

				timeoutCtx := tracing.ContextWithSpan(ctx, nil)
				log.Errorf(
					timeoutCtx,
					"flow id:%s : %d inbound streams timed out after %s; propagated error throughout flow",
					id,
					numTimedOutReceivers,
					timeout,
				)
			} else {
				__antithesis_instrumentation__.Notify(491684)
			}
		})
	} else {
		__antithesis_instrumentation__.Notify(491685)
	}
	__antithesis_instrumentation__.Notify(491664)
	return nil
}

func (fr *FlowRegistry) cancelPendingStreams(
	id execinfrapb.FlowID, pendingReceiverErr error,
) (numTimedOutReceivers int) {
	__antithesis_instrumentation__.Notify(491686)
	var inboundStreams map[execinfrapb.StreamID]*InboundStreamInfo
	fr.Lock()
	entry := fr.flows[id]
	if entry != nil && func() bool {
		__antithesis_instrumentation__.Notify(491690)
		return entry.flow != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(491691)

		inboundStreams = entry.inboundStreams
	} else {
		__antithesis_instrumentation__.Notify(491692)
	}
	__antithesis_instrumentation__.Notify(491687)
	fr.Unlock()
	if inboundStreams == nil {
		__antithesis_instrumentation__.Notify(491693)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(491694)
	}
	__antithesis_instrumentation__.Notify(491688)
	for _, is := range inboundStreams {
		__antithesis_instrumentation__.Notify(491695)

		pendingReceiver := is.cancelIfNotConnected()
		if pendingReceiver != nil {
			__antithesis_instrumentation__.Notify(491696)
			numTimedOutReceivers++
			go func(receiver InboundStreamHandler) {
				__antithesis_instrumentation__.Notify(491697)
				receiver.Timeout(pendingReceiverErr)
			}(pendingReceiver)
		} else {
			__antithesis_instrumentation__.Notify(491698)
		}
	}
	__antithesis_instrumentation__.Notify(491689)
	return numTimedOutReceivers
}

func (fr *FlowRegistry) UnregisterFlow(id execinfrapb.FlowID) {
	__antithesis_instrumentation__.Notify(491699)
	fr.Lock()
	entry := fr.flows[id]
	if entry.streamTimer != nil {
		__antithesis_instrumentation__.Notify(491701)
		entry.streamTimer.Stop()
		entry.streamTimer = nil
	} else {
		__antithesis_instrumentation__.Notify(491702)
	}
	__antithesis_instrumentation__.Notify(491700)
	fr.releaseEntryLocked(id)
	fr.Unlock()
}

func (fr *FlowRegistry) waitForFlow(
	ctx context.Context, id execinfrapb.FlowID, timeout time.Duration,
) *FlowBase {
	__antithesis_instrumentation__.Notify(491703)
	fr.Lock()
	defer fr.Unlock()
	entry := fr.getEntryLocked(id)
	if entry.flow != nil {
		__antithesis_instrumentation__.Notify(491707)

		return entry.flow
	} else {
		__antithesis_instrumentation__.Notify(491708)
	}
	__antithesis_instrumentation__.Notify(491704)

	waitCh := entry.waitCh
	if waitCh == nil {
		__antithesis_instrumentation__.Notify(491709)
		waitCh = make(chan struct{})
		entry.waitCh = waitCh
	} else {
		__antithesis_instrumentation__.Notify(491710)
	}
	__antithesis_instrumentation__.Notify(491705)
	entry.refCount++
	fr.Unlock()

	select {
	case <-waitCh:
		__antithesis_instrumentation__.Notify(491711)
	case <-time.After(timeout):
		__antithesis_instrumentation__.Notify(491712)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(491713)
	}
	__antithesis_instrumentation__.Notify(491706)

	fr.Lock()
	fr.releaseEntryLocked(id)
	return entry.flow
}

func (fr *FlowRegistry) Drain(
	flowDrainWait time.Duration,
	minFlowDrainWait time.Duration,
	reporter func(int, redact.SafeString),
) {
	__antithesis_instrumentation__.Notify(491714)
	allFlowsDone := make(chan struct{}, 1)
	start := timeutil.Now()
	stopWaiting := false

	sleep := func(t time.Duration) {
		__antithesis_instrumentation__.Notify(491722)
		if fr.testingRunBeforeDrainSleep != nil {
			__antithesis_instrumentation__.Notify(491724)
			fr.testingRunBeforeDrainSleep()
		} else {
			__antithesis_instrumentation__.Notify(491725)
		}
		__antithesis_instrumentation__.Notify(491723)
		time.Sleep(t)
	}
	__antithesis_instrumentation__.Notify(491715)

	defer func() {
		__antithesis_instrumentation__.Notify(491726)

		fr.Lock()
		fr.draining = true
		if len(fr.flows) > 0 {
			__antithesis_instrumentation__.Notify(491728)
			fr.Unlock()
			time.Sleep(expectedConnectionTime)
			fr.Lock()
		} else {
			__antithesis_instrumentation__.Notify(491729)
		}
		__antithesis_instrumentation__.Notify(491727)
		fr.Unlock()
	}()
	__antithesis_instrumentation__.Notify(491716)

	fr.Lock()
	if len(fr.flows) == 0 {
		__antithesis_instrumentation__.Notify(491730)
		fr.Unlock()
		sleep(minFlowDrainWait)
		fr.Lock()

		if len(fr.flows) == 0 {
			__antithesis_instrumentation__.Notify(491731)
			fr.Unlock()
			return
		} else {
			__antithesis_instrumentation__.Notify(491732)
		}
	} else {
		__antithesis_instrumentation__.Notify(491733)
	}
	__antithesis_instrumentation__.Notify(491717)
	if reporter != nil {
		__antithesis_instrumentation__.Notify(491734)

		reporter(len(fr.flows), "distSQL execution flows")
	} else {
		__antithesis_instrumentation__.Notify(491735)
	}
	__antithesis_instrumentation__.Notify(491718)

	go func() {
		__antithesis_instrumentation__.Notify(491736)
		select {
		case <-time.After(flowDrainWait):
			__antithesis_instrumentation__.Notify(491737)
			fr.Lock()
			stopWaiting = true
			fr.flowDone.Signal()
			fr.Unlock()
		case <-allFlowsDone:
			__antithesis_instrumentation__.Notify(491738)
		}
	}()
	__antithesis_instrumentation__.Notify(491719)

	for !(stopWaiting || func() bool {
		__antithesis_instrumentation__.Notify(491739)
		return len(fr.flows) == 0 == true
	}() == true) {
		__antithesis_instrumentation__.Notify(491740)
		fr.flowDone.Wait()
	}
	__antithesis_instrumentation__.Notify(491720)
	fr.Unlock()

	waitTime := timeutil.Since(start)
	if waitTime < minFlowDrainWait {
		__antithesis_instrumentation__.Notify(491741)
		sleep(minFlowDrainWait - waitTime)
		fr.Lock()
		for !(stopWaiting || func() bool {
			__antithesis_instrumentation__.Notify(491743)
			return len(fr.flows) == 0 == true
		}() == true) {
			__antithesis_instrumentation__.Notify(491744)
			fr.flowDone.Wait()
		}
		__antithesis_instrumentation__.Notify(491742)
		fr.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(491745)
	}
	__antithesis_instrumentation__.Notify(491721)

	allFlowsDone <- struct{}{}
}

func (fr *FlowRegistry) Undrain() {
	__antithesis_instrumentation__.Notify(491746)
	fr.Lock()
	fr.draining = false
	fr.Unlock()
}

func (fr *FlowRegistry) ConnectInboundStream(
	ctx context.Context,
	flowID execinfrapb.FlowID,
	streamID execinfrapb.StreamID,
	stream execinfrapb.DistSQL_FlowStreamServer,
	timeout time.Duration,
) (*FlowBase, InboundStreamHandler, func(), error) {
	__antithesis_instrumentation__.Notify(491747)
	fr.Lock()
	entry := fr.getEntryLocked(flowID)
	flow := entry.flow
	fr.Unlock()
	if flow == nil {
		__antithesis_instrumentation__.Notify(491752)

		deadline := timeutil.Now().Add(timeout)
		if err := stream.Send(&execinfrapb.ConsumerSignal{
			Handshake: &execinfrapb.ConsumerHandshake{
				ConsumerScheduled:        false,
				ConsumerScheduleDeadline: &deadline,
				Version:                  execinfra.Version,
				MinAcceptedVersion:       execinfra.MinAcceptedVersion,
			},
		}); err != nil {
			__antithesis_instrumentation__.Notify(491754)

			return nil, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(491755)
		}
		__antithesis_instrumentation__.Notify(491753)
		flow = fr.waitForFlow(ctx, flowID, timeout)
		if flow == nil {
			__antithesis_instrumentation__.Notify(491756)
			return nil, nil, nil, errors.Errorf("flow %s not found", flowID)
		} else {
			__antithesis_instrumentation__.Notify(491757)
		}
	} else {
		__antithesis_instrumentation__.Notify(491758)
	}
	__antithesis_instrumentation__.Notify(491748)

	s, ok := entry.inboundStreams[streamID]
	if !ok {
		__antithesis_instrumentation__.Notify(491759)
		return nil, nil, nil, errors.Errorf("flow %s: no inbound stream %d", flowID, streamID)
	} else {
		__antithesis_instrumentation__.Notify(491760)
	}
	__antithesis_instrumentation__.Notify(491749)

	if err := s.connect(flowID, streamID); err != nil {
		__antithesis_instrumentation__.Notify(491761)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(491762)
	}
	__antithesis_instrumentation__.Notify(491750)

	if err := stream.Send(&execinfrapb.ConsumerSignal{
		Handshake: &execinfrapb.ConsumerHandshake{
			ConsumerScheduled:  true,
			Version:            execinfra.Version,
			MinAcceptedVersion: execinfra.MinAcceptedVersion,
		},
	}); err != nil {
		__antithesis_instrumentation__.Notify(491763)
		s.disconnect()
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(491764)
	}
	__antithesis_instrumentation__.Notify(491751)

	return flow, s.receiver, s.finish, nil
}
