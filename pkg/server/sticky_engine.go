package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble/vfs"
)

type stickyInMemEngine struct {
	id string

	closed bool

	storage.Engine

	fs vfs.FS
}

type StickyEngineRegistryConfigOption func(cfg *stickyEngineRegistryConfig)

var ReplaceEngines StickyEngineRegistryConfigOption = func(cfg *stickyEngineRegistryConfig) {
	__antithesis_instrumentation__.Notify(238063)
	cfg.replaceEngines = true
}

type StickyInMemEnginesRegistry interface {
	GetOrCreateStickyInMemEngine(ctx context.Context, cfg *Config, spec base.StoreSpec) (storage.Engine, error)

	GetUnderlyingFS(spec base.StoreSpec) (vfs.FS, error)

	CloseAllStickyInMemEngines()
}

var _ storage.Engine = &stickyInMemEngine{}

func (e *stickyInMemEngine) Close() {
	__antithesis_instrumentation__.Notify(238064)
	e.closed = true
}

func (e *stickyInMemEngine) Closed() bool {
	__antithesis_instrumentation__.Notify(238065)
	return e.closed
}

type stickyInMemEnginesRegistryImpl struct {
	entries map[string]*stickyInMemEngine
	mu      syncutil.Mutex
	cfg     stickyEngineRegistryConfig
}

func NewStickyInMemEnginesRegistry(
	opts ...StickyEngineRegistryConfigOption,
) StickyInMemEnginesRegistry {
	__antithesis_instrumentation__.Notify(238066)
	var cfg stickyEngineRegistryConfig
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(238068)
		opt(&cfg)
	}
	__antithesis_instrumentation__.Notify(238067)
	return &stickyInMemEnginesRegistryImpl{
		entries: map[string]*stickyInMemEngine{},
		cfg:     cfg,
	}
}

func (registry *stickyInMemEnginesRegistryImpl) GetOrCreateStickyInMemEngine(
	ctx context.Context, cfg *Config, spec base.StoreSpec,
) (storage.Engine, error) {
	__antithesis_instrumentation__.Notify(238069)
	registry.mu.Lock()
	defer registry.mu.Unlock()

	var fs vfs.FS
	if engine, ok := registry.entries[spec.StickyInMemoryEngineID]; ok {
		__antithesis_instrumentation__.Notify(238071)
		if !engine.closed {
			__antithesis_instrumentation__.Notify(238074)
			return nil, errors.Errorf("sticky engine %s has not been closed", spec.StickyInMemoryEngineID)
		} else {
			__antithesis_instrumentation__.Notify(238075)
		}
		__antithesis_instrumentation__.Notify(238072)
		if !registry.cfg.replaceEngines {
			__antithesis_instrumentation__.Notify(238076)
			log.Infof(ctx, "re-using sticky in-mem engine %s", spec.StickyInMemoryEngineID)
			engine.closed = false
			return engine, nil
		} else {
			__antithesis_instrumentation__.Notify(238077)
		}
		__antithesis_instrumentation__.Notify(238073)
		fs = engine.fs
		registry.deleteEngine(spec.StickyInMemoryEngineID)
	} else {
		__antithesis_instrumentation__.Notify(238078)
		fs = vfs.NewMem()
	}
	__antithesis_instrumentation__.Notify(238070)
	options := []storage.ConfigOption{
		storage.Attributes(spec.Attributes),
		storage.CacheSize(cfg.CacheSize),
		storage.MaxSize(spec.Size.InBytes),
		storage.EncryptionAtRest(spec.EncryptionOptions),
		storage.ForStickyEngineTesting,
	}

	log.Infof(ctx, "creating new sticky in-mem engine %s", spec.StickyInMemoryEngineID)
	engine := storage.InMemFromFS(ctx, fs, "", options...)

	engineEntry := &stickyInMemEngine{
		id:     spec.StickyInMemoryEngineID,
		closed: false,

		Engine: engine,
		fs:     fs,
	}
	registry.entries[spec.StickyInMemoryEngineID] = engineEntry
	return engineEntry, nil
}

func (registry *stickyInMemEnginesRegistryImpl) GetUnderlyingFS(
	spec base.StoreSpec,
) (vfs.FS, error) {
	__antithesis_instrumentation__.Notify(238079)
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if engine, ok := registry.entries[spec.StickyInMemoryEngineID]; ok {
		__antithesis_instrumentation__.Notify(238081)
		return engine.fs, nil
	} else {
		__antithesis_instrumentation__.Notify(238082)
	}
	__antithesis_instrumentation__.Notify(238080)
	return nil, errors.Errorf("engine '%s' was not created", spec.StickyInMemoryEngineID)
}

func (registry *stickyInMemEnginesRegistryImpl) CloseAllStickyInMemEngines() {
	__antithesis_instrumentation__.Notify(238083)
	registry.mu.Lock()
	defer registry.mu.Unlock()

	for id := range registry.entries {
		__antithesis_instrumentation__.Notify(238084)
		registry.deleteEngine(id)
	}
}

func (registry *stickyInMemEnginesRegistryImpl) deleteEngine(id string) {
	__antithesis_instrumentation__.Notify(238085)
	engine, ok := registry.entries[id]
	if !ok {
		__antithesis_instrumentation__.Notify(238087)
		return
	} else {
		__antithesis_instrumentation__.Notify(238088)
	}
	__antithesis_instrumentation__.Notify(238086)
	engine.closed = true
	engine.Engine.Close()
	delete(registry.entries, id)
}

type stickyEngineRegistryConfig struct {
	replaceEngines bool
}
