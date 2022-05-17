package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

var ErrClusterInitialized = fmt.Errorf("cluster has already been initialized")

var ErrIncompatibleBinaryVersion = fmt.Errorf("binary is incompatible with the cluster attempted to join")

type initServer struct {
	log.AmbientContext

	config initServerCfg

	mu struct {
		syncutil.Mutex

		bootstrapped bool

		rejectErr error
	}

	inspectedDiskState *initState

	bootstrapReqCh chan *initState
}

func (s *initServer) NeedsBootstrap() bool {
	__antithesis_instrumentation__.Notify(193862)
	s.mu.Lock()
	defer s.mu.Unlock()

	return !s.mu.bootstrapped
}

func newInitServer(
	actx log.AmbientContext, inspectedDiskState *initState, config initServerCfg,
) *initServer {
	__antithesis_instrumentation__.Notify(193863)
	initServer := &initServer{
		AmbientContext:     actx,
		bootstrapReqCh:     make(chan *initState, 1),
		config:             config,
		inspectedDiskState: inspectedDiskState,
	}

	if inspectedDiskState.bootstrapped() {
		__antithesis_instrumentation__.Notify(193865)
		initServer.mu.bootstrapped = true
	} else {
		__antithesis_instrumentation__.Notify(193866)
	}
	__antithesis_instrumentation__.Notify(193864)
	return initServer
}

type initState struct {
	nodeID               roachpb.NodeID
	clusterID            uuid.UUID
	clusterVersion       clusterversion.ClusterVersion
	initializedEngines   []storage.Engine
	uninitializedEngines []storage.Engine
	initialSettingsKVs   []roachpb.KeyValue
}

func (i *initState) bootstrapped() bool {
	__antithesis_instrumentation__.Notify(193867)
	return len(i.initializedEngines) > 0
}

func (i *initState) validate() error {
	__antithesis_instrumentation__.Notify(193868)
	if (i.clusterID == uuid.UUID{}) {
		__antithesis_instrumentation__.Notify(193871)
		return errors.New("missing cluster ID")
	} else {
		__antithesis_instrumentation__.Notify(193872)
	}
	__antithesis_instrumentation__.Notify(193869)
	if i.nodeID == 0 {
		__antithesis_instrumentation__.Notify(193873)
		return errors.New("missing node ID")
	} else {
		__antithesis_instrumentation__.Notify(193874)
	}
	__antithesis_instrumentation__.Notify(193870)
	return nil
}

type joinResult struct {
	state *initState
	err   error
}

func (s *initServer) ServeAndWait(
	ctx context.Context, stopper *stop.Stopper, sv *settings.Values,
) (state *initState, initialBoot bool, err error) {
	__antithesis_instrumentation__.Notify(193875)

	if s.inspectedDiskState.bootstrapped() {
		__antithesis_instrumentation__.Notify(193879)
		return s.inspectedDiskState, false, nil
	} else {
		__antithesis_instrumentation__.Notify(193880)
	}
	__antithesis_instrumentation__.Notify(193876)

	log.Info(ctx, "no stores initialized")
	log.Info(ctx, "awaiting `cockroach init` or join with an already initialized node")

	var joinCh chan joinResult
	var cancelJoin = func() { __antithesis_instrumentation__.Notify(193881) }
	__antithesis_instrumentation__.Notify(193877)
	var wg sync.WaitGroup

	if len(s.config.bootstrapAddresses) == 0 {
		__antithesis_instrumentation__.Notify(193882)

	} else {
		__antithesis_instrumentation__.Notify(193883)
		joinCh = make(chan joinResult, 1)
		wg.Add(1)

		var joinCtx context.Context
		joinCtx, cancelJoin = context.WithCancel(ctx)
		defer cancelJoin()

		err := stopper.RunAsyncTask(joinCtx, "init server: join loop",
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(193885)
				defer wg.Done()

				state, err := s.startJoinLoop(ctx, stopper)
				joinCh <- joinResult{state: state, err: err}
			})
		__antithesis_instrumentation__.Notify(193884)
		if err != nil {
			__antithesis_instrumentation__.Notify(193886)
			wg.Done()
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(193887)
		}
	}
	__antithesis_instrumentation__.Notify(193878)

	for {
		__antithesis_instrumentation__.Notify(193888)
		select {
		case state := <-s.bootstrapReqCh:
			__antithesis_instrumentation__.Notify(193889)

			cancelJoin()
			wg.Wait()

			if err := clusterversion.Initialize(ctx, state.clusterVersion.Version, sv); err != nil {
				__antithesis_instrumentation__.Notify(193894)
				return nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(193895)
			}
			__antithesis_instrumentation__.Notify(193890)

			log.Infof(ctx, "cluster %s has been created", state.clusterID)
			log.Infof(ctx, "allocated node ID: n%d (for self)", state.nodeID)
			log.Infof(ctx, "active cluster version: %s", state.clusterVersion)

			return state, true, nil
		case result := <-joinCh:
			__antithesis_instrumentation__.Notify(193891)

			wg.Wait()

			if err := result.err; err != nil {
				__antithesis_instrumentation__.Notify(193896)
				if errors.Is(err, ErrIncompatibleBinaryVersion) {
					__antithesis_instrumentation__.Notify(193898)
					return nil, false, err
				} else {
					__antithesis_instrumentation__.Notify(193899)
				}
				__antithesis_instrumentation__.Notify(193897)

				return nil, false, errors.NewAssertionErrorWithWrappedErrf(err, "unexpected error")
			} else {
				__antithesis_instrumentation__.Notify(193900)
			}
			__antithesis_instrumentation__.Notify(193892)

			state := result.state

			log.Infof(ctx, "joined cluster %s through join rpc", state.clusterID)
			log.Infof(ctx, "received node ID: %d", state.nodeID)
			log.Infof(ctx, "received cluster version: %s", state.clusterVersion)

			return state, true, nil
		case <-stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(193893)
			return nil, false, stop.ErrUnavailable
		}
	}
}

var errInternalBootstrapError = errors.New("unable to bootstrap due to internal error")

func (s *initServer) Bootstrap(
	ctx context.Context, _ *serverpb.BootstrapRequest,
) (*serverpb.BootstrapResponse, error) {
	__antithesis_instrumentation__.Notify(193901)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.mu.bootstrapped {
		__antithesis_instrumentation__.Notify(193905)
		return nil, ErrClusterInitialized
	} else {
		__antithesis_instrumentation__.Notify(193906)
	}
	__antithesis_instrumentation__.Notify(193902)

	if s.mu.rejectErr != nil {
		__antithesis_instrumentation__.Notify(193907)
		return nil, s.mu.rejectErr
	} else {
		__antithesis_instrumentation__.Notify(193908)
	}
	__antithesis_instrumentation__.Notify(193903)

	state, err := s.tryBootstrap(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(193909)
		log.Errorf(ctx, "bootstrap: %v", err)
		s.mu.rejectErr = errInternalBootstrapError
		return nil, s.mu.rejectErr
	} else {
		__antithesis_instrumentation__.Notify(193910)
	}
	__antithesis_instrumentation__.Notify(193904)

	s.mu.bootstrapped = true
	s.bootstrapReqCh <- state

	return &serverpb.BootstrapResponse{}, nil
}

func (s *initServer) startJoinLoop(ctx context.Context, stopper *stop.Stopper) (*initState, error) {
	__antithesis_instrumentation__.Notify(193911)
	if len(s.config.bootstrapAddresses) == 0 {
		__antithesis_instrumentation__.Notify(193915)
		return nil, errors.AssertionFailedf("expected to find at least one bootstrap address, found none")
	} else {
		__antithesis_instrumentation__.Notify(193916)
	}
	__antithesis_instrumentation__.Notify(193912)

	for _, addr := range s.config.bootstrapAddresses {
		__antithesis_instrumentation__.Notify(193917)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(193922)
			return nil, context.Canceled
		case <-stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(193923)
			return nil, stop.ErrUnavailable
		default:
			__antithesis_instrumentation__.Notify(193924)
		}
		__antithesis_instrumentation__.Notify(193918)

		resp, err := s.attemptJoinTo(ctx, addr.String())
		if errors.Is(err, ErrIncompatibleBinaryVersion) {
			__antithesis_instrumentation__.Notify(193925)

			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(193926)
		}
		__antithesis_instrumentation__.Notify(193919)
		if err != nil {
			__antithesis_instrumentation__.Notify(193927)

			if IsWaitingForInit(err) {
				__antithesis_instrumentation__.Notify(193929)
				log.Infof(ctx, "%s is itself waiting for init, will retry", addr)
			} else {
				__antithesis_instrumentation__.Notify(193930)
				log.Warningf(ctx, "outgoing join rpc to %s unsuccessful: %v", addr, err.Error())
			}
			__antithesis_instrumentation__.Notify(193928)
			continue
		} else {
			__antithesis_instrumentation__.Notify(193931)
		}
		__antithesis_instrumentation__.Notify(193920)

		state, err := s.initializeFirstStoreAfterJoin(ctx, resp)
		if err != nil {
			__antithesis_instrumentation__.Notify(193932)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(193933)
		}
		__antithesis_instrumentation__.Notify(193921)

		s.mu.Lock()
		s.mu.bootstrapped = true
		s.mu.Unlock()

		return state, nil
	}
	__antithesis_instrumentation__.Notify(193913)

	const joinRPCBackoff = time.Second
	var tickChan <-chan time.Time
	{
		__antithesis_instrumentation__.Notify(193934)
		ticker := time.NewTicker(joinRPCBackoff)
		tickChan = ticker.C
		defer ticker.Stop()
	}
	__antithesis_instrumentation__.Notify(193914)

	for idx := 0; ; idx = (idx + 1) % len(s.config.bootstrapAddresses) {
		__antithesis_instrumentation__.Notify(193935)
		addr := s.config.bootstrapAddresses[idx].String()
		select {
		case <-tickChan:
			__antithesis_instrumentation__.Notify(193936)
			resp, err := s.attemptJoinTo(ctx, addr)
			if errors.Is(err, ErrIncompatibleBinaryVersion) {
				__antithesis_instrumentation__.Notify(193942)

				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(193943)
			}
			__antithesis_instrumentation__.Notify(193937)
			if err != nil {
				__antithesis_instrumentation__.Notify(193944)

				if IsWaitingForInit(err) {
					__antithesis_instrumentation__.Notify(193946)
					log.Infof(ctx, "%s is itself waiting for init, will retry", addr)
				} else {
					__antithesis_instrumentation__.Notify(193947)
					log.Warningf(ctx, "outgoing join rpc to %s unsuccessful: %v", addr, err.Error())
				}
				__antithesis_instrumentation__.Notify(193945)
				continue
			} else {
				__antithesis_instrumentation__.Notify(193948)
			}
			__antithesis_instrumentation__.Notify(193938)

			state, err := s.initializeFirstStoreAfterJoin(ctx, resp)
			if err != nil {
				__antithesis_instrumentation__.Notify(193949)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(193950)
			}
			__antithesis_instrumentation__.Notify(193939)

			s.mu.Lock()
			s.mu.bootstrapped = true
			s.mu.Unlock()

			return state, nil
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(193940)
			return nil, context.Canceled
		case <-stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(193941)
			return nil, stop.ErrUnavailable
		}
	}
}

func (s *initServer) attemptJoinTo(
	ctx context.Context, addr string,
) (*roachpb.JoinNodeResponse, error) {
	__antithesis_instrumentation__.Notify(193951)
	conn, err := grpc.DialContext(ctx, addr, s.config.dialOpts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(193955)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193956)
	}
	__antithesis_instrumentation__.Notify(193952)

	defer func() {
		__antithesis_instrumentation__.Notify(193957)
		_ = conn.Close()
	}()
	__antithesis_instrumentation__.Notify(193953)

	binaryVersion := s.config.binaryVersion
	req := &roachpb.JoinNodeRequest{
		BinaryVersion: &binaryVersion,
	}

	initClient := roachpb.NewInternalClient(conn)
	resp, err := initClient.Join(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(193958)

		status, ok := grpcstatus.FromError(errors.UnwrapAll(err))
		if !ok {
			__antithesis_instrumentation__.Notify(193961)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(193962)
		}
		__antithesis_instrumentation__.Notify(193959)

		if status.Code() == codes.PermissionDenied {
			__antithesis_instrumentation__.Notify(193963)
			log.Infof(ctx, "%s is running a version higher than our binary version %s", addr, req.BinaryVersion.String())
			return nil, ErrIncompatibleBinaryVersion
		} else {
			__antithesis_instrumentation__.Notify(193964)
		}
		__antithesis_instrumentation__.Notify(193960)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193965)
	}
	__antithesis_instrumentation__.Notify(193954)

	return resp, nil
}

func (s *initServer) tryBootstrap(ctx context.Context) (*initState, error) {
	__antithesis_instrumentation__.Notify(193966)

	if err := assertEnginesEmpty(s.inspectedDiskState.uninitializedEngines); err != nil {
		__antithesis_instrumentation__.Notify(193969)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193970)
	}
	__antithesis_instrumentation__.Notify(193967)

	cv := clusterversion.ClusterVersion{Version: s.config.binaryVersion}
	if err := kvserver.WriteClusterVersionToEngines(ctx, s.inspectedDiskState.uninitializedEngines, cv); err != nil {
		__antithesis_instrumentation__.Notify(193971)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193972)
	}
	__antithesis_instrumentation__.Notify(193968)

	return bootstrapCluster(ctx, s.inspectedDiskState.uninitializedEngines, s.config)
}

func (s *initServer) DiskClusterVersion() clusterversion.ClusterVersion {
	__antithesis_instrumentation__.Notify(193973)
	return s.inspectedDiskState.clusterVersion
}

func (s *initServer) initializeFirstStoreAfterJoin(
	ctx context.Context, resp *roachpb.JoinNodeResponse,
) (*initState, error) {
	__antithesis_instrumentation__.Notify(193974)

	if err := assertEnginesEmpty(s.inspectedDiskState.uninitializedEngines); err != nil {
		__antithesis_instrumentation__.Notify(193979)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193980)
	}
	__antithesis_instrumentation__.Notify(193975)

	firstEngine := s.inspectedDiskState.uninitializedEngines[0]
	clusterVersion := clusterversion.ClusterVersion{Version: *resp.ActiveVersion}
	if err := kvserver.WriteClusterVersion(ctx, firstEngine, clusterVersion); err != nil {
		__antithesis_instrumentation__.Notify(193981)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193982)
	}
	__antithesis_instrumentation__.Notify(193976)

	sIdent, err := resp.CreateStoreIdent()
	if err != nil {
		__antithesis_instrumentation__.Notify(193983)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193984)
	}
	__antithesis_instrumentation__.Notify(193977)
	if err := kvserver.InitEngine(ctx, firstEngine, sIdent); err != nil {
		__antithesis_instrumentation__.Notify(193985)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193986)
	}
	__antithesis_instrumentation__.Notify(193978)

	return inspectEngines(
		ctx, s.inspectedDiskState.uninitializedEngines,
		s.config.binaryVersion, s.config.binaryMinSupportedVersion,
	)
}

func assertEnginesEmpty(engines []storage.Engine) error {
	__antithesis_instrumentation__.Notify(193987)
	storeClusterVersionKey := keys.StoreClusterVersionKey()

	for _, engine := range engines {
		__antithesis_instrumentation__.Notify(193989)
		err := func() error {
			__antithesis_instrumentation__.Notify(193991)
			iter := engine.NewEngineIterator(storage.IterOptions{
				UpperBound: roachpb.KeyMax,
			})
			defer iter.Close()

			valid, err := iter.SeekEngineKeyGE(storage.EngineKey{Key: roachpb.KeyMin})
			for ; valid && func() bool {
				__antithesis_instrumentation__.Notify(193993)
				return err == nil == true
			}() == true; valid, err = iter.NextEngineKey() {
				__antithesis_instrumentation__.Notify(193994)
				k, err := iter.UnsafeEngineKey()
				if err != nil {
					__antithesis_instrumentation__.Notify(193997)
					return err
				} else {
					__antithesis_instrumentation__.Notify(193998)
				}
				__antithesis_instrumentation__.Notify(193995)

				if storeClusterVersionKey.Equal(k.Key) {
					__antithesis_instrumentation__.Notify(193999)
					continue
				} else {
					__antithesis_instrumentation__.Notify(194000)
				}
				__antithesis_instrumentation__.Notify(193996)
				return errors.New("engine is not empty")
			}
			__antithesis_instrumentation__.Notify(193992)
			return err
		}()
		__antithesis_instrumentation__.Notify(193990)
		if err != nil {
			__antithesis_instrumentation__.Notify(194001)
			return err
		} else {
			__antithesis_instrumentation__.Notify(194002)
		}
	}
	__antithesis_instrumentation__.Notify(193988)
	return nil
}

type initServerCfg struct {
	advertiseAddr             string
	binaryMinSupportedVersion roachpb.Version
	binaryVersion             roachpb.Version
	defaultSystemZoneConfig   zonepb.ZoneConfig
	defaultZoneConfig         zonepb.ZoneConfig

	dialOpts []grpc.DialOption

	bootstrapAddresses []util.UnresolvedAddr

	testingKnobs base.TestingKnobs
}

func newInitServerConfig(
	ctx context.Context, cfg Config, dialOpts []grpc.DialOption,
) initServerCfg {
	__antithesis_instrumentation__.Notify(194003)
	binaryVersion := cfg.Settings.Version.BinaryVersion()
	if knobs := cfg.TestingKnobs.Server; knobs != nil {
		__antithesis_instrumentation__.Notify(194006)
		if ov := knobs.(*TestingKnobs).BinaryVersionOverride; ov != (roachpb.Version{}) {
			__antithesis_instrumentation__.Notify(194007)
			binaryVersion = ov
		} else {
			__antithesis_instrumentation__.Notify(194008)
		}
	} else {
		__antithesis_instrumentation__.Notify(194009)
	}
	__antithesis_instrumentation__.Notify(194004)
	binaryMinSupportedVersion := cfg.Settings.Version.BinaryMinSupportedVersion()
	if binaryVersion.Less(binaryMinSupportedVersion) {
		__antithesis_instrumentation__.Notify(194010)
		log.Fatalf(ctx, "binary version (%s) less than min supported version (%s)",
			binaryVersion, binaryMinSupportedVersion)
	} else {
		__antithesis_instrumentation__.Notify(194011)
	}
	__antithesis_instrumentation__.Notify(194005)

	bootstrapAddresses := cfg.FilterGossipBootstrapAddresses(context.Background())
	return initServerCfg{
		advertiseAddr:             cfg.AdvertiseAddr,
		binaryMinSupportedVersion: binaryMinSupportedVersion,
		binaryVersion:             binaryVersion,
		defaultSystemZoneConfig:   cfg.DefaultSystemZoneConfig,
		defaultZoneConfig:         cfg.DefaultZoneConfig,
		dialOpts:                  dialOpts,
		bootstrapAddresses:        bootstrapAddresses,
		testingKnobs:              cfg.TestingKnobs,
	}
}

func inspectEngines(
	ctx context.Context,
	engines []storage.Engine,
	binaryVersion, binaryMinSupportedVersion roachpb.Version,
) (*initState, error) {
	__antithesis_instrumentation__.Notify(194012)
	var clusterID uuid.UUID
	var nodeID roachpb.NodeID
	var initializedEngines, uninitializedEngines []storage.Engine
	var initialSettingsKVs []roachpb.KeyValue

	for _, eng := range engines {
		__antithesis_instrumentation__.Notify(194015)

		if len(initialSettingsKVs) == 0 {
			__antithesis_instrumentation__.Notify(194021)
			var err error
			initialSettingsKVs, err = loadCachedSettingsKVs(ctx, eng)
			if err != nil {
				__antithesis_instrumentation__.Notify(194022)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(194023)
			}
		} else {
			__antithesis_instrumentation__.Notify(194024)
		}
		__antithesis_instrumentation__.Notify(194016)

		storeIdent, err := kvserver.ReadStoreIdent(ctx, eng)
		if errors.HasType(err, (*kvserver.NotBootstrappedError)(nil)) {
			__antithesis_instrumentation__.Notify(194025)
			uninitializedEngines = append(uninitializedEngines, eng)
			continue
		} else {
			__antithesis_instrumentation__.Notify(194026)
			if err != nil {
				__antithesis_instrumentation__.Notify(194027)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(194028)
			}
		}
		__antithesis_instrumentation__.Notify(194017)

		if clusterID != uuid.Nil && func() bool {
			__antithesis_instrumentation__.Notify(194029)
			return clusterID != storeIdent.ClusterID == true
		}() == true {
			__antithesis_instrumentation__.Notify(194030)
			return nil, errors.Errorf("conflicting store ClusterIDs: %s, %s", storeIdent.ClusterID, clusterID)
		} else {
			__antithesis_instrumentation__.Notify(194031)
		}
		__antithesis_instrumentation__.Notify(194018)
		clusterID = storeIdent.ClusterID

		if storeIdent.StoreID == 0 || func() bool {
			__antithesis_instrumentation__.Notify(194032)
			return storeIdent.NodeID == 0 == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(194033)
			return storeIdent.ClusterID == uuid.Nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(194034)
			return nil, errors.Errorf("partially initialized store: %+v", storeIdent)
		} else {
			__antithesis_instrumentation__.Notify(194035)
		}
		__antithesis_instrumentation__.Notify(194019)

		if nodeID != 0 && func() bool {
			__antithesis_instrumentation__.Notify(194036)
			return nodeID != storeIdent.NodeID == true
		}() == true {
			__antithesis_instrumentation__.Notify(194037)
			return nil, errors.Errorf("conflicting store NodeIDs: %s, %s", storeIdent.NodeID, nodeID)
		} else {
			__antithesis_instrumentation__.Notify(194038)
		}
		__antithesis_instrumentation__.Notify(194020)
		nodeID = storeIdent.NodeID

		initializedEngines = append(initializedEngines, eng)
	}
	__antithesis_instrumentation__.Notify(194013)
	clusterVersion, err := kvserver.SynthesizeClusterVersionFromEngines(
		ctx, initializedEngines, binaryVersion, binaryMinSupportedVersion,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(194039)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194040)
	}
	__antithesis_instrumentation__.Notify(194014)

	state := &initState{
		clusterID:            clusterID,
		nodeID:               nodeID,
		initializedEngines:   initializedEngines,
		uninitializedEngines: uninitializedEngines,
		clusterVersion:       clusterVersion,
		initialSettingsKVs:   initialSettingsKVs,
	}
	return state, nil
}
