// Package kvtenantccl provides utilities required by SQL-only tenant processes
// in order to interact with the key-value layer.
package kvtenantccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"math/rand"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvtenant"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errorspb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	kvtenant.Factory = connectorFactory{}
}

type Connector struct {
	log.AmbientContext

	tenantID        roachpb.TenantID
	rpcContext      *rpc.Context
	rpcRetryOptions retry.Options
	rpcDialTimeout  time.Duration
	rpcDial         singleflight.Group
	defaultZoneCfg  *zonepb.ZoneConfig
	addrs           []string

	mu struct {
		syncutil.RWMutex
		client               *client
		nodeDescs            map[roachpb.NodeID]*roachpb.NodeDescriptor
		systemConfig         *config.SystemConfig
		systemConfigChannels map[chan<- struct{}]struct{}
	}

	settingsMu struct {
		syncutil.Mutex

		allTenantOverrides map[string]settings.EncodedValue
		specificOverrides  map[string]settings.EncodedValue

		notifyCh chan struct{}
	}
}

type client struct {
	roachpb.InternalClient
	serverpb.StatusClient
}

var _ kvcoord.NodeDescStore = (*Connector)(nil)

var _ rangecache.RangeDescriptorDB = (*Connector)(nil)

var _ config.SystemConfigProvider = (*Connector)(nil)

var _ serverpb.RegionsServer = (*Connector)(nil)

var _ serverpb.TenantStatusServer = (*Connector)(nil)

var _ spanconfig.KVAccessor = (*Connector)(nil)

func NewConnector(cfg kvtenant.ConnectorConfig, addrs []string) *Connector {
	__antithesis_instrumentation__.Notify(19641)
	cfg.AmbientCtx.AddLogTag("tenant-connector", nil)
	if cfg.TenantID.IsSystem() {
		__antithesis_instrumentation__.Notify(19643)
		panic("TenantID not set")
	} else {
		__antithesis_instrumentation__.Notify(19644)
	}
	__antithesis_instrumentation__.Notify(19642)
	c := &Connector{
		tenantID:        cfg.TenantID,
		AmbientContext:  cfg.AmbientCtx,
		rpcContext:      cfg.RPCContext,
		rpcRetryOptions: cfg.RPCRetryOptions,
		defaultZoneCfg:  cfg.DefaultZoneConfig,
		addrs:           addrs,
	}

	c.mu.nodeDescs = make(map[roachpb.NodeID]*roachpb.NodeDescriptor)
	c.mu.systemConfigChannels = make(map[chan<- struct{}]struct{})
	c.settingsMu.allTenantOverrides = make(map[string]settings.EncodedValue)
	c.settingsMu.specificOverrides = make(map[string]settings.EncodedValue)
	return c
}

type connectorFactory struct{}

func (connectorFactory) NewConnector(
	cfg kvtenant.ConnectorConfig, addrs []string,
) (kvtenant.Connector, error) {
	__antithesis_instrumentation__.Notify(19645)
	return NewConnector(cfg, addrs), nil
}

func (c *Connector) Start(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(19646)
	gossipStartupCh := make(chan struct{})
	settingsStartupCh := make(chan struct{})
	bgCtx := c.AnnotateCtx(context.Background())

	if err := c.rpcContext.Stopper.RunAsyncTask(bgCtx, "connector-gossip", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(19650)
		ctx = c.AnnotateCtx(ctx)
		ctx, cancel := c.rpcContext.Stopper.WithCancelOnQuiesce(ctx)
		defer cancel()
		c.runGossipSubscription(ctx, gossipStartupCh)
	}); err != nil {
		__antithesis_instrumentation__.Notify(19651)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19652)
	}
	__antithesis_instrumentation__.Notify(19647)

	if err := c.rpcContext.Stopper.RunAsyncTask(bgCtx, "connector-settings", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(19653)
		ctx = c.AnnotateCtx(ctx)
		ctx, cancel := c.rpcContext.Stopper.WithCancelOnQuiesce(ctx)
		defer cancel()
		c.runTenantSettingsSubscription(ctx, settingsStartupCh)
	}); err != nil {
		__antithesis_instrumentation__.Notify(19654)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19655)
	}
	__antithesis_instrumentation__.Notify(19648)

	for gossipStartupCh != nil || func() bool {
		__antithesis_instrumentation__.Notify(19656)
		return settingsStartupCh != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(19657)
		select {
		case <-gossipStartupCh:
			__antithesis_instrumentation__.Notify(19658)
			log.Infof(ctx, "kv connector gossip subscription started")
			gossipStartupCh = nil
		case <-settingsStartupCh:
			__antithesis_instrumentation__.Notify(19659)
			log.Infof(ctx, "kv connector tenant settings started")
			settingsStartupCh = nil
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(19660)
			return ctx.Err()
		}
	}
	__antithesis_instrumentation__.Notify(19649)
	return nil
}

func (c *Connector) runGossipSubscription(ctx context.Context, startupCh chan struct{}) {
	__antithesis_instrumentation__.Notify(19661)
	for ctx.Err() == nil {
		__antithesis_instrumentation__.Notify(19662)
		client, err := c.getClient(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(19665)
			continue
		} else {
			__antithesis_instrumentation__.Notify(19666)
		}
		__antithesis_instrumentation__.Notify(19663)
		stream, err := client.GossipSubscription(ctx, &roachpb.GossipSubscriptionRequest{
			Patterns: gossipSubsPatterns,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(19667)
			log.Warningf(ctx, "error issuing GossipSubscription RPC: %v", err)
			c.tryForgetClient(ctx, client)
			continue
		} else {
			__antithesis_instrumentation__.Notify(19668)
		}
		__antithesis_instrumentation__.Notify(19664)
		for {
			__antithesis_instrumentation__.Notify(19669)
			e, err := stream.Recv()
			if err != nil {
				__antithesis_instrumentation__.Notify(19673)
				if err == io.EOF {
					__antithesis_instrumentation__.Notify(19675)
					break
				} else {
					__antithesis_instrumentation__.Notify(19676)
				}
				__antithesis_instrumentation__.Notify(19674)

				log.Warningf(ctx, "error consuming GossipSubscription RPC: %v", err)
				c.tryForgetClient(ctx, client)
				break
			} else {
				__antithesis_instrumentation__.Notify(19677)
			}
			__antithesis_instrumentation__.Notify(19670)
			if e.Error != nil {
				__antithesis_instrumentation__.Notify(19678)

				log.Errorf(ctx, "error consuming GossipSubscription RPC: %v", e.Error)
				continue
			} else {
				__antithesis_instrumentation__.Notify(19679)
			}
			__antithesis_instrumentation__.Notify(19671)
			handler, ok := gossipSubsHandlers[e.PatternMatched]
			if !ok {
				__antithesis_instrumentation__.Notify(19680)
				log.Errorf(ctx, "unknown GossipSubscription pattern: %q", e.PatternMatched)
				continue
			} else {
				__antithesis_instrumentation__.Notify(19681)
			}
			__antithesis_instrumentation__.Notify(19672)
			handler(c, ctx, e.Key, e.Content)

			if startupCh != nil && func() bool {
				__antithesis_instrumentation__.Notify(19682)
				return e.PatternMatched == gossip.KeyClusterID == true
			}() == true {
				__antithesis_instrumentation__.Notify(19683)
				close(startupCh)
				startupCh = nil
			} else {
				__antithesis_instrumentation__.Notify(19684)
			}
		}
	}
}

var gossipSubsHandlers = map[string]func(*Connector, context.Context, string, roachpb.Value){

	gossip.KeyClusterID: (*Connector).updateClusterID,

	gossip.MakePrefixPattern(gossip.KeyNodeIDPrefix): (*Connector).updateNodeAddress,

	gossip.KeyDeprecatedSystemConfig: (*Connector).updateSystemConfig,
}

var gossipSubsPatterns = func() []string {
	__antithesis_instrumentation__.Notify(19685)
	patterns := make([]string, 0, len(gossipSubsHandlers))
	for pattern := range gossipSubsHandlers {
		__antithesis_instrumentation__.Notify(19687)
		patterns = append(patterns, pattern)
	}
	__antithesis_instrumentation__.Notify(19686)
	sort.Strings(patterns)
	return patterns
}()

func (c *Connector) updateClusterID(ctx context.Context, key string, content roachpb.Value) {
	__antithesis_instrumentation__.Notify(19688)
	bytes, err := content.GetBytes()
	if err != nil {
		__antithesis_instrumentation__.Notify(19691)
		log.Errorf(ctx, "invalid ClusterID value: %v", content.RawBytes)
		return
	} else {
		__antithesis_instrumentation__.Notify(19692)
	}
	__antithesis_instrumentation__.Notify(19689)
	clusterID, err := uuid.FromBytes(bytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(19693)
		log.Errorf(ctx, "invalid ClusterID value: %v", content.RawBytes)
		return
	} else {
		__antithesis_instrumentation__.Notify(19694)
	}
	__antithesis_instrumentation__.Notify(19690)
	c.rpcContext.StorageClusterID.Set(ctx, clusterID)
}

func (c *Connector) updateNodeAddress(ctx context.Context, key string, content roachpb.Value) {
	__antithesis_instrumentation__.Notify(19695)
	desc := new(roachpb.NodeDescriptor)
	if err := content.GetProto(desc); err != nil {
		__antithesis_instrumentation__.Notify(19697)
		log.Errorf(ctx, "could not unmarshal node descriptor: %v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(19698)
	}
	__antithesis_instrumentation__.Notify(19696)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.mu.nodeDescs[desc.NodeID] = desc
}

func (c *Connector) GetNodeDescriptor(nodeID roachpb.NodeID) (*roachpb.NodeDescriptor, error) {
	__antithesis_instrumentation__.Notify(19699)
	c.mu.RLock()
	defer c.mu.RUnlock()
	desc, ok := c.mu.nodeDescs[nodeID]
	if !ok {
		__antithesis_instrumentation__.Notify(19701)
		return nil, errors.Errorf("unable to look up descriptor for n%d", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(19702)
	}
	__antithesis_instrumentation__.Notify(19700)
	return desc, nil
}

func (c *Connector) updateSystemConfig(ctx context.Context, key string, content roachpb.Value) {
	__antithesis_instrumentation__.Notify(19703)
	cfg := config.NewSystemConfig(c.defaultZoneCfg)
	if err := content.GetProto(&cfg.SystemConfigEntries); err != nil {
		__antithesis_instrumentation__.Notify(19705)
		log.Errorf(ctx, "could not unmarshal system config: %v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(19706)
	}
	__antithesis_instrumentation__.Notify(19704)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.mu.systemConfig = cfg
	for c := range c.mu.systemConfigChannels {
		__antithesis_instrumentation__.Notify(19707)
		select {
		case c <- struct{}{}:
			__antithesis_instrumentation__.Notify(19708)
		default:
			__antithesis_instrumentation__.Notify(19709)
		}
	}
}

func (c *Connector) GetSystemConfig() *config.SystemConfig {
	__antithesis_instrumentation__.Notify(19710)

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mu.systemConfig
}

func (c *Connector) RegisterSystemConfigChannel() (_ <-chan struct{}, unregister func()) {
	__antithesis_instrumentation__.Notify(19711)

	ch := make(chan struct{}, 1)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.mu.systemConfigChannels[ch] = struct{}{}

	if c.mu.systemConfig != nil {
		__antithesis_instrumentation__.Notify(19713)
		ch <- struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(19714)
	}
	__antithesis_instrumentation__.Notify(19712)
	return ch, func() {
		__antithesis_instrumentation__.Notify(19715)
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.mu.systemConfigChannels, ch)
	}
}

func (c *Connector) RangeLookup(
	ctx context.Context, key roachpb.RKey, useReverseScan bool,
) ([]roachpb.RangeDescriptor, []roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(19716)

	ctx = c.AnnotateCtx(ctx)
	for ctx.Err() == nil {
		__antithesis_instrumentation__.Notify(19718)
		client, err := c.getClient(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(19722)
			continue
		} else {
			__antithesis_instrumentation__.Notify(19723)
		}
		__antithesis_instrumentation__.Notify(19719)
		resp, err := client.RangeLookup(ctx, &roachpb.RangeLookupRequest{
			Key: key,

			ReadConsistency: roachpb.READ_UNCOMMITTED,

			PrefetchNum:     0,
			PrefetchReverse: useReverseScan,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(19724)
			log.Warningf(ctx, "error issuing RangeLookup RPC: %v", err)
			if grpcutil.IsAuthError(err) {
				__antithesis_instrumentation__.Notify(19726)

				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(19727)
			}
			__antithesis_instrumentation__.Notify(19725)

			c.tryForgetClient(ctx, client)
			continue
		} else {
			__antithesis_instrumentation__.Notify(19728)
		}
		__antithesis_instrumentation__.Notify(19720)
		if resp.Error != nil {
			__antithesis_instrumentation__.Notify(19729)

			return nil, nil, resp.Error.GoError()
		} else {
			__antithesis_instrumentation__.Notify(19730)
		}
		__antithesis_instrumentation__.Notify(19721)
		return resp.Descriptors, resp.PrefetchedDescriptors, nil
	}
	__antithesis_instrumentation__.Notify(19717)
	return nil, nil, ctx.Err()
}

func (c *Connector) Regions(
	ctx context.Context, req *serverpb.RegionsRequest,
) (resp *serverpb.RegionsResponse, _ error) {
	__antithesis_instrumentation__.Notify(19731)
	if err := c.withClient(ctx, func(ctx context.Context, c *client) error {
		__antithesis_instrumentation__.Notify(19733)
		var err error
		resp, err = c.Regions(ctx, req)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(19734)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(19735)
	}
	__antithesis_instrumentation__.Notify(19732)

	return resp, nil
}

func (c *Connector) TenantRanges(
	ctx context.Context, req *serverpb.TenantRangesRequest,
) (resp *serverpb.TenantRangesResponse, _ error) {
	__antithesis_instrumentation__.Notify(19736)
	if err := c.withClient(ctx, func(ctx context.Context, c *client) error {
		__antithesis_instrumentation__.Notify(19738)
		var err error
		resp, err = c.TenantRanges(ctx, req)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(19739)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(19740)
	}
	__antithesis_instrumentation__.Notify(19737)

	return resp, nil
}

func (c *Connector) FirstRange() (*roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(19741)
	return nil, status.Error(codes.Unauthenticated, "kvtenant.Proxy does not have access to FirstRange")
}

func (c *Connector) TokenBucket(
	ctx context.Context, in *roachpb.TokenBucketRequest,
) (*roachpb.TokenBucketResponse, error) {
	__antithesis_instrumentation__.Notify(19742)

	ctx = c.AnnotateCtx(ctx)
	for ctx.Err() == nil {
		__antithesis_instrumentation__.Notify(19744)
		client, err := c.getClient(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(19748)
			continue
		} else {
			__antithesis_instrumentation__.Notify(19749)
		}
		__antithesis_instrumentation__.Notify(19745)
		resp, err := client.TokenBucket(ctx, in)
		if err != nil {
			__antithesis_instrumentation__.Notify(19750)
			log.Warningf(ctx, "error issuing TokenBucket RPC: %v", err)
			if grpcutil.IsAuthError(err) {
				__antithesis_instrumentation__.Notify(19752)

				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(19753)
			}
			__antithesis_instrumentation__.Notify(19751)

			c.tryForgetClient(ctx, client)
			continue
		} else {
			__antithesis_instrumentation__.Notify(19754)
		}
		__antithesis_instrumentation__.Notify(19746)
		if resp.Error != (errorspb.EncodedError{}) {
			__antithesis_instrumentation__.Notify(19755)

			return nil, errors.DecodeError(ctx, resp.Error)
		} else {
			__antithesis_instrumentation__.Notify(19756)
		}
		__antithesis_instrumentation__.Notify(19747)
		return resp, nil
	}
	__antithesis_instrumentation__.Notify(19743)
	return nil, ctx.Err()
}

func (c *Connector) GetSpanConfigRecords(
	ctx context.Context, targets []spanconfig.Target,
) (records []spanconfig.Record, _ error) {
	__antithesis_instrumentation__.Notify(19757)
	if err := c.withClient(ctx, func(ctx context.Context, c *client) error {
		__antithesis_instrumentation__.Notify(19759)
		resp, err := c.GetSpanConfigs(ctx, &roachpb.GetSpanConfigsRequest{
			Targets: spanconfig.TargetsToProtos(targets),
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(19762)
			return err
		} else {
			__antithesis_instrumentation__.Notify(19763)
		}
		__antithesis_instrumentation__.Notify(19760)

		records, err = spanconfig.EntriesToRecords(resp.SpanConfigEntries)
		if err != nil {
			__antithesis_instrumentation__.Notify(19764)
			return err
		} else {
			__antithesis_instrumentation__.Notify(19765)
		}
		__antithesis_instrumentation__.Notify(19761)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(19766)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(19767)
	}
	__antithesis_instrumentation__.Notify(19758)
	return records, nil
}

func (c *Connector) UpdateSpanConfigRecords(
	ctx context.Context,
	toDelete []spanconfig.Target,
	toUpsert []spanconfig.Record,
	minCommitTS, maxCommitTS hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(19768)
	return c.withClient(ctx, func(ctx context.Context, c *client) error {
		__antithesis_instrumentation__.Notify(19769)
		resp, err := c.UpdateSpanConfigs(ctx, &roachpb.UpdateSpanConfigsRequest{
			ToDelete:           spanconfig.TargetsToProtos(toDelete),
			ToUpsert:           spanconfig.RecordsToEntries(toUpsert),
			MinCommitTimestamp: minCommitTS,
			MaxCommitTimestamp: maxCommitTS,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(19772)
			return err
		} else {
			__antithesis_instrumentation__.Notify(19773)
		}
		__antithesis_instrumentation__.Notify(19770)
		if resp.Error.IsSet() {
			__antithesis_instrumentation__.Notify(19774)

			return errors.DecodeError(ctx, resp.Error)
		} else {
			__antithesis_instrumentation__.Notify(19775)
		}
		__antithesis_instrumentation__.Notify(19771)
		return nil
	})
}

func (c *Connector) GetAllSystemSpanConfigsThatApply(
	ctx context.Context, id roachpb.TenantID,
) ([]roachpb.SpanConfig, error) {
	__antithesis_instrumentation__.Notify(19776)
	var spanConfigs []roachpb.SpanConfig
	if err := c.withClient(ctx, func(ctx context.Context, c *client) error {
		__antithesis_instrumentation__.Notify(19778)
		var err error
		resp, err := c.GetAllSystemSpanConfigsThatApply(
			ctx, &roachpb.GetAllSystemSpanConfigsThatApplyRequest{
				TenantID: id,
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(19780)
			return err
		} else {
			__antithesis_instrumentation__.Notify(19781)
		}
		__antithesis_instrumentation__.Notify(19779)

		spanConfigs = resp.SpanConfigs
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(19782)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(19783)
	}
	__antithesis_instrumentation__.Notify(19777)
	return spanConfigs, nil
}

func (c *Connector) WithTxn(context.Context, *kv.Txn) spanconfig.KVAccessor {
	__antithesis_instrumentation__.Notify(19784)
	panic("not applicable")
}

func (c *Connector) withClient(
	ctx context.Context, f func(ctx context.Context, c *client) error,
) error {
	__antithesis_instrumentation__.Notify(19785)
	ctx = c.AnnotateCtx(ctx)
	for ctx.Err() == nil {
		__antithesis_instrumentation__.Notify(19787)
		c, err := c.getClient(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(19789)
			continue
		} else {
			__antithesis_instrumentation__.Notify(19790)
		}
		__antithesis_instrumentation__.Notify(19788)
		return f(ctx, c)
	}
	__antithesis_instrumentation__.Notify(19786)
	return ctx.Err()
}

func (c *Connector) getClient(ctx context.Context) (*client, error) {
	__antithesis_instrumentation__.Notify(19791)
	c.mu.RLock()
	if client := c.mu.client; client != nil {
		__antithesis_instrumentation__.Notify(19794)
		c.mu.RUnlock()
		return client, nil
	} else {
		__antithesis_instrumentation__.Notify(19795)
	}
	__antithesis_instrumentation__.Notify(19792)
	ch, _ := c.rpcDial.DoChan("dial", func() (interface{}, error) {
		__antithesis_instrumentation__.Notify(19796)
		dialCtx := c.AnnotateCtx(context.Background())
		dialCtx, cancel := c.rpcContext.Stopper.WithCancelOnQuiesce(dialCtx)
		defer cancel()
		var client *client
		err := c.rpcContext.Stopper.RunTaskWithErr(dialCtx, "kvtenant.Connector: dial",
			func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(19799)
				var err error
				client, err = c.dialAddrs(ctx)
				return err
			})
		__antithesis_instrumentation__.Notify(19797)
		if err != nil {
			__antithesis_instrumentation__.Notify(19800)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(19801)
		}
		__antithesis_instrumentation__.Notify(19798)
		c.mu.Lock()
		defer c.mu.Unlock()
		c.mu.client = client
		return client, nil
	})
	__antithesis_instrumentation__.Notify(19793)
	c.mu.RUnlock()

	select {
	case res := <-ch:
		__antithesis_instrumentation__.Notify(19802)
		if res.Err != nil {
			__antithesis_instrumentation__.Notify(19805)
			return nil, res.Err
		} else {
			__antithesis_instrumentation__.Notify(19806)
		}
		__antithesis_instrumentation__.Notify(19803)
		return res.Val.(*client), nil
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(19804)
		return nil, ctx.Err()
	}
}

func (c *Connector) dialAddrs(ctx context.Context) (*client, error) {
	__antithesis_instrumentation__.Notify(19807)
	for r := retry.StartWithCtx(ctx, c.rpcRetryOptions); r.Next(); {
		__antithesis_instrumentation__.Notify(19809)

		for _, i := range rand.Perm(len(c.addrs)) {
			__antithesis_instrumentation__.Notify(19810)
			addr := c.addrs[i]
			conn, err := c.dialAddr(ctx, addr)
			if err != nil {
				__antithesis_instrumentation__.Notify(19812)
				log.Warningf(ctx, "error dialing tenant KV address %s: %v", addr, err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(19813)
			}
			__antithesis_instrumentation__.Notify(19811)
			return &client{
				InternalClient: roachpb.NewInternalClient(conn),
				StatusClient:   serverpb.NewStatusClient(conn),
			}, nil
		}
	}
	__antithesis_instrumentation__.Notify(19808)
	return nil, ctx.Err()
}

func (c *Connector) dialAddr(ctx context.Context, addr string) (conn *grpc.ClientConn, err error) {
	__antithesis_instrumentation__.Notify(19814)
	if c.rpcDialTimeout == 0 {
		__antithesis_instrumentation__.Notify(19817)
		return c.rpcContext.GRPCUnvalidatedDial(addr).Connect(ctx)
	} else {
		__antithesis_instrumentation__.Notify(19818)
	}
	__antithesis_instrumentation__.Notify(19815)
	err = contextutil.RunWithTimeout(ctx, "dial addr", c.rpcDialTimeout, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(19819)
		conn, err = c.rpcContext.GRPCUnvalidatedDial(addr).Connect(ctx)
		return err
	})
	__antithesis_instrumentation__.Notify(19816)
	return conn, err
}

func (c *Connector) tryForgetClient(ctx context.Context, client roachpb.InternalClient) {
	__antithesis_instrumentation__.Notify(19820)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(19822)

		return
	} else {
		__antithesis_instrumentation__.Notify(19823)
	}
	__antithesis_instrumentation__.Notify(19821)

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mu.client == client {
		__antithesis_instrumentation__.Notify(19824)
		c.mu.client = nil
	} else {
		__antithesis_instrumentation__.Notify(19825)
	}
}
