// Package kvtenant provides utilities required by SQL-only tenant processes in
// order to interact with the key-value layer.
package kvtenant

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net"

	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/settingswatcher"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
)

type Connector interface {
	Start(context.Context) error

	kvcoord.NodeDescStore

	rangecache.RangeDescriptorDB

	serverpb.RegionsServer

	serverpb.TenantStatusServer

	TokenBucketProvider

	spanconfig.KVAccessor

	settingswatcher.OverridesMonitor

	config.SystemConfigProvider
}

type TokenBucketProvider interface {
	TokenBucket(
		ctx context.Context, in *roachpb.TokenBucketRequest,
	) (*roachpb.TokenBucketResponse, error)
}

type ConnectorConfig struct {
	TenantID          roachpb.TenantID
	AmbientCtx        log.AmbientContext
	RPCContext        *rpc.Context
	RPCRetryOptions   retry.Options
	DefaultZoneConfig *zonepb.ZoneConfig
}

type ConnectorFactory interface {
	NewConnector(cfg ConnectorConfig, addrs []string) (Connector, error)
}

var Factory ConnectorFactory = requiresCCLBinaryFactory{}

type requiresCCLBinaryFactory struct{}

func (requiresCCLBinaryFactory) NewConnector(_ ConnectorConfig, _ []string) (Connector, error) {
	__antithesis_instrumentation__.Notify(89251)
	return nil, errors.Errorf(`tenant connector requires a CCL binary`)
}

func AddressResolver(s kvcoord.NodeDescStore) nodedialer.AddressResolver {
	__antithesis_instrumentation__.Notify(89252)
	return func(nodeID roachpb.NodeID) (net.Addr, error) {
		__antithesis_instrumentation__.Notify(89253)
		nd, err := s.GetNodeDescriptor(nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(89255)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(89256)
		}
		__antithesis_instrumentation__.Notify(89254)
		return &nd.Address, nil
	}
}

var GossipSubscriptionSystemConfigMask = config.MakeSystemConfigMask(

	config.MakeZoneKey(keys.SystemSQLCodec, keys.RootNamespaceID),
	config.MakeZoneKey(keys.SystemSQLCodec, keys.TenantsRangesID),
)
