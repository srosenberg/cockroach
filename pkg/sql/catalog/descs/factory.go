package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/hydratedtables"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
)

type CollectionFactory struct {
	settings           *cluster.Settings
	codec              keys.SQLCodec
	leaseMgr           *lease.Manager
	virtualSchemas     catalog.VirtualSchemas
	hydratedTables     *hydratedtables.Cache
	systemDatabase     *systemDatabaseNamespaceCache
	spanConfigSplitter spanconfig.Splitter
	spanConfigLimiter  spanconfig.Limiter
}

func NewCollectionFactory(
	settings *cluster.Settings,
	leaseMgr *lease.Manager,
	virtualSchemas catalog.VirtualSchemas,
	hydratedTables *hydratedtables.Cache,
	spanConfigSplitter spanconfig.Splitter,
	spanConfigLimiter spanconfig.Limiter,
) *CollectionFactory {
	__antithesis_instrumentation__.Notify(264500)
	return &CollectionFactory{
		settings:           settings,
		codec:              leaseMgr.Codec(),
		leaseMgr:           leaseMgr,
		virtualSchemas:     virtualSchemas,
		hydratedTables:     hydratedTables,
		systemDatabase:     newSystemDatabaseNamespaceCache(leaseMgr.Codec()),
		spanConfigSplitter: spanConfigSplitter,
		spanConfigLimiter:  spanConfigLimiter,
	}
}

func NewBareBonesCollectionFactory(
	settings *cluster.Settings, codec keys.SQLCodec,
) *CollectionFactory {
	__antithesis_instrumentation__.Notify(264501)
	return &CollectionFactory{
		settings: settings,
		codec:    codec,
	}
}

func (cf *CollectionFactory) MakeCollection(
	ctx context.Context, temporarySchemaProvider TemporarySchemaProvider,
) Collection {
	__antithesis_instrumentation__.Notify(264502)
	return makeCollection(ctx, cf.leaseMgr, cf.settings, cf.codec, cf.hydratedTables, cf.systemDatabase,
		cf.virtualSchemas, temporarySchemaProvider)
}

func (cf *CollectionFactory) NewCollection(
	ctx context.Context, temporarySchemaProvider TemporarySchemaProvider,
) *Collection {
	__antithesis_instrumentation__.Notify(264503)
	c := cf.MakeCollection(ctx, temporarySchemaProvider)
	return &c
}
