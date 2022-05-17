// Package slprovider exposes an implementation of the sqlliveness.Provider
// interface.
package slprovider

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness/slinstance"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness/slstorage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
)

func New(
	ambientCtx log.AmbientContext,
	stopper *stop.Stopper,
	clock *hlc.Clock,
	db *kv.DB,
	codec keys.SQLCodec,
	settings *cluster.Settings,
	testingKnobs *sqlliveness.TestingKnobs,
) sqlliveness.Provider {
	__antithesis_instrumentation__.Notify(624263)
	storage := slstorage.NewStorage(ambientCtx, stopper, clock, db, codec, settings)
	instance := slinstance.NewSQLInstance(stopper, clock, storage, settings, testingKnobs)
	return &provider{
		Storage:  storage,
		Instance: instance,
	}
}

func (p *provider) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(624264)
	p.Storage.Start(ctx)
	p.Instance.Start(ctx)
}

func (p *provider) Metrics() metric.Struct {
	__antithesis_instrumentation__.Notify(624265)
	return p.Storage.Metrics()
}

type provider struct {
	*slstorage.Storage
	*slinstance.Instance
}
