package scdeps

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

func NewBuilderDependencies(
	clusterID uuid.UUID,
	codec keys.SQLCodec,
	txn *kv.Txn,
	descsCollection *descs.Collection,
	schemaResolver resolver.SchemaResolver,
	authAccessor scbuild.AuthorizationAccessor,
	astFormatter scbuild.AstFormatter,
	featureChecker scbuild.FeatureChecker,
	sessionData *sessiondata.SessionData,
	settings *cluster.Settings,
	statements []string,
) scbuild.Dependencies {
	__antithesis_instrumentation__.Notify(580561)
	return &buildDeps{
		clusterID:       clusterID,
		codec:           codec,
		txn:             txn,
		descsCollection: descsCollection,
		schemaResolver:  schemaResolver,
		authAccessor:    authAccessor,
		sessionData:     sessionData,
		settings:        settings,
		statements:      statements,
		astFormatter:    astFormatter,
		featureChecker:  featureChecker,
	}
}

type buildDeps struct {
	clusterID       uuid.UUID
	codec           keys.SQLCodec
	txn             *kv.Txn
	descsCollection *descs.Collection
	schemaResolver  resolver.SchemaResolver
	authAccessor    scbuild.AuthorizationAccessor
	sessionData     *sessiondata.SessionData
	settings        *cluster.Settings
	statements      []string
	astFormatter    scbuild.AstFormatter
	featureChecker  scbuild.FeatureChecker
}

var _ scbuild.CatalogReader = (*buildDeps)(nil)

func (d *buildDeps) MayResolveDatabase(
	ctx context.Context, name tree.Name,
) catalog.DatabaseDescriptor {
	__antithesis_instrumentation__.Notify(580562)
	db, err := d.descsCollection.GetImmutableDatabaseByName(ctx, d.txn, string(name), tree.DatabaseLookupFlags{
		AvoidLeased: true,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(580564)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580565)
	}
	__antithesis_instrumentation__.Notify(580563)
	return db
}

func (d *buildDeps) MayResolveSchema(
	ctx context.Context, name tree.ObjectNamePrefix,
) (catalog.DatabaseDescriptor, catalog.SchemaDescriptor) {
	__antithesis_instrumentation__.Notify(580566)
	if !name.ExplicitCatalog {
		__antithesis_instrumentation__.Notify(580569)
		name.CatalogName = tree.Name(d.schemaResolver.CurrentDatabase())
	} else {
		__antithesis_instrumentation__.Notify(580570)
	}
	__antithesis_instrumentation__.Notify(580567)
	db := d.MayResolveDatabase(ctx, name.CatalogName)
	schema, err := d.descsCollection.GetSchemaByName(ctx, d.txn, db, name.Schema(), tree.SchemaLookupFlags{
		AvoidLeased: true,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(580571)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580572)
	}
	__antithesis_instrumentation__.Notify(580568)
	return db, schema
}

func (d *buildDeps) MayResolveTable(
	ctx context.Context, name tree.UnresolvedObjectName,
) (catalog.ResolvedObjectPrefix, catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(580573)
	desc, prefix, err := resolver.ResolveExistingObject(ctx, d.schemaResolver, &name, tree.ObjectLookupFlags{
		CommonLookupFlags: tree.CommonLookupFlags{
			AvoidLeased: true,
		},
		DesiredObjectKind: tree.TableObject,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(580576)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580577)
	}
	__antithesis_instrumentation__.Notify(580574)
	if desc == nil {
		__antithesis_instrumentation__.Notify(580578)
		return prefix, nil
	} else {
		__antithesis_instrumentation__.Notify(580579)
	}
	__antithesis_instrumentation__.Notify(580575)
	return prefix, desc.(catalog.TableDescriptor)
}

func (d *buildDeps) MayResolveType(
	ctx context.Context, name tree.UnresolvedObjectName,
) (catalog.ResolvedObjectPrefix, catalog.TypeDescriptor) {
	__antithesis_instrumentation__.Notify(580580)
	desc, prefix, err := resolver.ResolveExistingObject(ctx, d.schemaResolver, &name, tree.ObjectLookupFlags{
		CommonLookupFlags: tree.CommonLookupFlags{
			AvoidLeased: true,
		},
		DesiredObjectKind: tree.TypeObject,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(580583)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580584)
	}
	__antithesis_instrumentation__.Notify(580581)
	if desc == nil {
		__antithesis_instrumentation__.Notify(580585)
		return prefix, nil
	} else {
		__antithesis_instrumentation__.Notify(580586)
	}
	__antithesis_instrumentation__.Notify(580582)
	return prefix, desc.(catalog.TypeDescriptor)
}

func (d *buildDeps) ReadObjectNamesAndIDs(
	ctx context.Context, db catalog.DatabaseDescriptor, schema catalog.SchemaDescriptor,
) (tree.TableNames, descpb.IDs) {
	__antithesis_instrumentation__.Notify(580587)
	names, ids, err := d.descsCollection.GetObjectNamesAndIDs(ctx, d.txn, db, schema.GetName(), tree.DatabaseListFlags{
		CommonLookupFlags: tree.CommonLookupFlags{
			Required:       true,
			RequireMutable: false,
			AvoidLeased:    true,
			IncludeOffline: true,
			IncludeDropped: true,
		},
		ExplicitPrefix: true,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(580589)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580590)
	}
	__antithesis_instrumentation__.Notify(580588)
	return names, ids
}

func (d *buildDeps) ResolveType(
	ctx context.Context, name *tree.UnresolvedObjectName,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(580591)
	return d.schemaResolver.ResolveType(ctx, name)
}

func (d *buildDeps) ResolveTypeByOID(ctx context.Context, oid oid.Oid) (*types.T, error) {
	__antithesis_instrumentation__.Notify(580592)
	return d.schemaResolver.ResolveTypeByOID(ctx, oid)
}

func (d *buildDeps) GetQualifiedTableNameByID(
	ctx context.Context, id int64, requiredType tree.RequiredTableKind,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(580593)
	return d.schemaResolver.GetQualifiedTableNameByID(ctx, id, requiredType)
}

func (d *buildDeps) CurrentDatabase() string {
	__antithesis_instrumentation__.Notify(580594)
	return d.schemaResolver.CurrentDatabase()
}

func (d *buildDeps) MustReadDescriptor(ctx context.Context, id descpb.ID) catalog.Descriptor {
	__antithesis_instrumentation__.Notify(580595)
	flags := tree.CommonLookupFlags{
		Required:       true,
		RequireMutable: false,
		AvoidLeased:    true,
		IncludeOffline: true,
		IncludeDropped: true,
	}
	desc, err := d.descsCollection.GetImmutableDescriptorByID(ctx, d.txn, id, flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(580597)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580598)
	}
	__antithesis_instrumentation__.Notify(580596)
	return desc
}

func (d *buildDeps) MustGetSchemasForDatabase(
	ctx context.Context, database catalog.DatabaseDescriptor,
) map[descpb.ID]string {
	__antithesis_instrumentation__.Notify(580599)
	schemas, err := d.descsCollection.GetSchemasForDatabase(ctx, d.txn, database)
	if err != nil {
		__antithesis_instrumentation__.Notify(580601)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580602)
	}
	__antithesis_instrumentation__.Notify(580600)
	return schemas
}

var CreatePartitioningCCL scbuild.CreatePartitioningCCLCallback

var _ scbuild.Dependencies = (*buildDeps)(nil)

func (d *buildDeps) AuthorizationAccessor() scbuild.AuthorizationAccessor {
	__antithesis_instrumentation__.Notify(580603)
	return d.authAccessor
}

func (d *buildDeps) CatalogReader() scbuild.CatalogReader {
	__antithesis_instrumentation__.Notify(580604)
	return d
}

func (d *buildDeps) ClusterID() uuid.UUID {
	__antithesis_instrumentation__.Notify(580605)
	return d.clusterID
}

func (d *buildDeps) Codec() keys.SQLCodec {
	__antithesis_instrumentation__.Notify(580606)
	return d.codec
}

func (d *buildDeps) SessionData() *sessiondata.SessionData {
	__antithesis_instrumentation__.Notify(580607)
	return d.sessionData
}

func (d *buildDeps) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(580608)
	return d.settings
}

func (d *buildDeps) Statements() []string {
	__antithesis_instrumentation__.Notify(580609)
	return d.statements
}

func (d *buildDeps) AstFormatter() scbuild.AstFormatter {
	__antithesis_instrumentation__.Notify(580610)
	return d.astFormatter
}

func (d *buildDeps) FeatureChecker() scbuild.FeatureChecker {
	__antithesis_instrumentation__.Notify(580611)
	return d.featureChecker
}

func (d *buildDeps) IndexPartitioningCCLCallback() scbuild.CreatePartitioningCCLCallback {
	__antithesis_instrumentation__.Notify(580612)
	if CreatePartitioningCCL == nil {
		__antithesis_instrumentation__.Notify(580614)
		return func(
			_ context.Context,
			_ *cluster.Settings,
			_ *tree.EvalContext,
			_ func(tree.Name) (catalog.Column, error),
			_ int,
			_ []string,
			_ *tree.PartitionBy,
			_ []tree.Name,
			_ bool,
		) ([]catalog.Column, catpb.PartitioningDescriptor, error) {
			__antithesis_instrumentation__.Notify(580615)
			return nil, catpb.PartitioningDescriptor{}, sqlerrors.NewCCLRequiredError(errors.New(
				"creating or manipulating partitions requires a CCL binary"))
		}
	} else {
		__antithesis_instrumentation__.Notify(580616)
	}
	__antithesis_instrumentation__.Notify(580613)
	return CreatePartitioningCCL
}

func (d *buildDeps) IncrementSchemaChangeAlterCounter(counterType string, extra ...string) {
	__antithesis_instrumentation__.Notify(580617)
	var maybeExtra string
	if len(extra) > 0 {
		__antithesis_instrumentation__.Notify(580619)
		maybeExtra = extra[0]
	} else {
		__antithesis_instrumentation__.Notify(580620)
	}
	__antithesis_instrumentation__.Notify(580618)
	telemetry.Inc(sqltelemetry.SchemaChangeAlterCounterWithExtra(counterType, maybeExtra))
}

func (d *buildDeps) IncrementSchemaChangeDropCounter(counterType string) {
	__antithesis_instrumentation__.Notify(580621)
	telemetry.Inc(sqltelemetry.SchemaChangeDropCounter(counterType))
}

func (d *buildDeps) IncrementUserDefinedSchemaCounter(
	counterType sqltelemetry.UserDefinedSchemaTelemetryType,
) {
	__antithesis_instrumentation__.Notify(580622)
	sqltelemetry.IncrementUserDefinedSchemaCounter(counterType)
}

func (d *buildDeps) IncrementEnumCounter(counterType sqltelemetry.EnumTelemetryType) {
	__antithesis_instrumentation__.Notify(580623)
	sqltelemetry.IncrementEnumCounter(counterType)
}
