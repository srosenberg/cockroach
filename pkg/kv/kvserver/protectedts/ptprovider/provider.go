// Package ptprovider encapsulates the concrete implementation of the
// protectedts.Provider.
package ptprovider

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptcache"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptreconcile"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptstorage"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

type Config struct {
	Settings             *cluster.Settings
	DB                   *kv.DB
	Stores               *kvserver.Stores
	ReconcileStatusFuncs ptreconcile.StatusFuncs
	InternalExecutor     sqlutil.InternalExecutor
	Knobs                *protectedts.TestingKnobs
}

type Provider struct {
	protectedts.Storage
	protectedts.Cache
	protectedts.Reconciler
	metric.Struct
}

func New(cfg Config) (protectedts.Provider, error) {
	__antithesis_instrumentation__.Notify(111734)
	if err := validateConfig(cfg); err != nil {
		__antithesis_instrumentation__.Notify(111736)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(111737)
	}
	__antithesis_instrumentation__.Notify(111735)
	storage := ptstorage.New(cfg.Settings, cfg.InternalExecutor, cfg.Knobs)
	reconciler := ptreconcile.New(cfg.Settings, cfg.DB, storage, cfg.ReconcileStatusFuncs)
	cache := ptcache.New(ptcache.Config{
		DB:       cfg.DB,
		Storage:  storage,
		Settings: cfg.Settings,
	})

	return &Provider{
		Storage:    storage,
		Cache:      cache,
		Reconciler: reconciler,
		Struct:     reconciler.Metrics(),
	}, nil
}

func validateConfig(cfg Config) error {
	__antithesis_instrumentation__.Notify(111738)
	switch {
	case cfg.Settings == nil:
		__antithesis_instrumentation__.Notify(111739)
		return errors.Errorf("invalid nil Settings")
	case cfg.DB == nil:
		__antithesis_instrumentation__.Notify(111740)
		return errors.Errorf("invalid nil DB")
	case cfg.InternalExecutor == nil:
		__antithesis_instrumentation__.Notify(111741)
		return errors.Errorf("invalid nil InternalExecutor")
	default:
		__antithesis_instrumentation__.Notify(111742)
		return nil
	}
}

func (p *Provider) Start(ctx context.Context, stopper *stop.Stopper) error {
	__antithesis_instrumentation__.Notify(111743)
	if cache, ok := p.Cache.(*ptcache.Cache); ok {
		__antithesis_instrumentation__.Notify(111745)
		return cache.Start(ctx, stopper)
	} else {
		__antithesis_instrumentation__.Notify(111746)
	}
	__antithesis_instrumentation__.Notify(111744)
	return nil
}

func (p *Provider) Metrics() metric.Struct {
	__antithesis_instrumentation__.Notify(111747)
	return p.Struct
}
