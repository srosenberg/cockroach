package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net"
	"time"

	"github.com/cockroachdb/cockroach/pkg/blobs"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/server/diagnostics"
)

type TestingKnobs struct {
	DisableAutomaticVersionUpgrade chan struct{}

	DefaultZoneConfigOverride *zonepb.ZoneConfig

	DefaultSystemZoneConfigOverride *zonepb.ZoneConfig

	SignalAfterGettingRPCAddress chan struct{}

	PauseAfterGettingRPCAddress chan struct{}

	ContextTestingKnobs rpc.ContextTestingKnobs

	DiagnosticsTestingKnobs diagnostics.TestingKnobs

	RPCListener net.Listener

	BinaryVersionOverride roachpb.Version

	OnDecommissionedCallback func(livenesspb.Liveness)

	StickyEngineRegistry StickyInMemEnginesRegistry

	ClockSource func() int64

	ImportTimeseriesFile string

	ImportTimeseriesMappingFile string

	DrainSleepFn func(time.Duration)

	BlobClientFactory blobs.BlobClientFactory
}

func (*TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(239004) }
