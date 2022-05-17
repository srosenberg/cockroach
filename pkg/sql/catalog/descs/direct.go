package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/catkv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func (tc *Collection) Direct() Direct {
	__antithesis_instrumentation__.Notify(264412)
	return &tc.direct
}

type Direct interface {
	GetCatalogUnvalidated(
		ctx context.Context, txn *kv.Txn,
	) (nstree.Catalog, error)

	MustGetDatabaseDescByID(
		ctx context.Context, txn *kv.Txn, id descpb.ID,
	) (catalog.DatabaseDescriptor, error)

	MustGetSchemaDescByID(
		ctx context.Context, txn *kv.Txn, id descpb.ID,
	) (catalog.SchemaDescriptor, error)

	MustGetTypeDescByID(
		ctx context.Context, txn *kv.Txn, id descpb.ID,
	) (catalog.TypeDescriptor, error)

	MustGetTableDescByID(
		ctx context.Context, txn *kv.Txn, id descpb.ID,
	) (catalog.TableDescriptor, error)

	GetSchemaDescriptorsFromIDs(
		ctx context.Context, txn *kv.Txn, ids []descpb.ID,
	) ([]catalog.SchemaDescriptor, error)

	ResolveSchemaID(
		ctx context.Context, txn *kv.Txn, dbID descpb.ID, scName string,
	) (descpb.ID, error)

	GetDescriptorCollidingWithObject(
		ctx context.Context, txn *kv.Txn, parentID descpb.ID, parentSchemaID descpb.ID, name string,
	) (catalog.Descriptor, error)

	CheckObjectCollision(
		ctx context.Context,
		txn *kv.Txn,
		parentID descpb.ID,
		parentSchemaID descpb.ID,
		name tree.ObjectName,
	) error

	LookupDatabaseID(
		ctx context.Context, txn *kv.Txn, dbName string,
	) (descpb.ID, error)

	LookupSchemaID(
		ctx context.Context, txn *kv.Txn, dbID descpb.ID, schemaName string,
	) (descpb.ID, error)

	LookupObjectID(
		ctx context.Context, txn *kv.Txn, dbID descpb.ID, schemaID descpb.ID, objectName string,
	) (descpb.ID, error)

	WriteNewDescToBatch(
		ctx context.Context, kvTrace bool, b *kv.Batch, desc catalog.Descriptor,
	) error
}

type direct struct {
	settings *cluster.Settings
	codec    keys.SQLCodec
	version  clusterversion.ClusterVersion
}

func makeDirect(ctx context.Context, codec keys.SQLCodec, s *cluster.Settings) direct {
	__antithesis_instrumentation__.Notify(264413)
	return direct{
		settings: s,
		codec:    codec,
		version:  s.Version.ActiveVersion(ctx),
	}
}

func (d *direct) GetCatalogUnvalidated(ctx context.Context, txn *kv.Txn) (nstree.Catalog, error) {
	__antithesis_instrumentation__.Notify(264414)
	return catkv.GetCatalogUnvalidated(ctx, d.codec, txn)
}

func (d *direct) MustGetDatabaseDescByID(
	ctx context.Context, txn *kv.Txn, id descpb.ID,
) (catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(264415)
	desc, err := catkv.MustGetDescriptorByID(ctx, d.version, d.codec, txn, nil, id, catalog.Database)
	if err != nil {
		__antithesis_instrumentation__.Notify(264417)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264418)
	}
	__antithesis_instrumentation__.Notify(264416)
	return desc.(catalog.DatabaseDescriptor), nil
}

func (d *direct) MustGetSchemaDescByID(
	ctx context.Context, txn *kv.Txn, id descpb.ID,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(264419)
	desc, err := catkv.MustGetDescriptorByID(ctx, d.version, d.codec, txn, nil, id, catalog.Schema)
	if err != nil {
		__antithesis_instrumentation__.Notify(264421)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264422)
	}
	__antithesis_instrumentation__.Notify(264420)
	return desc.(catalog.SchemaDescriptor), nil
}

func (d *direct) MustGetTableDescByID(
	ctx context.Context, txn *kv.Txn, id descpb.ID,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(264423)
	desc, err := catkv.MustGetDescriptorByID(ctx, d.version, d.codec, txn, nil, id, catalog.Table)
	if err != nil {
		__antithesis_instrumentation__.Notify(264425)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264426)
	}
	__antithesis_instrumentation__.Notify(264424)
	return desc.(catalog.TableDescriptor), nil
}

func (d *direct) MustGetTypeDescByID(
	ctx context.Context, txn *kv.Txn, id descpb.ID,
) (catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(264427)
	desc, err := catkv.MustGetDescriptorByID(ctx, d.version, d.codec, txn, nil, id, catalog.Type)
	if err != nil {
		__antithesis_instrumentation__.Notify(264429)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264430)
	}
	__antithesis_instrumentation__.Notify(264428)
	return desc.(catalog.TypeDescriptor), nil
}

func (d *direct) GetSchemaDescriptorsFromIDs(
	ctx context.Context, txn *kv.Txn, ids []descpb.ID,
) ([]catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(264431)
	descs, err := catkv.MustGetDescriptorsByID(ctx, d.version, d.codec, txn, nil, ids, catalog.Schema)
	if err != nil {
		__antithesis_instrumentation__.Notify(264434)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264435)
	}
	__antithesis_instrumentation__.Notify(264432)
	ret := make([]catalog.SchemaDescriptor, len(descs))
	for i, desc := range descs {
		__antithesis_instrumentation__.Notify(264436)
		ret[i] = desc.(catalog.SchemaDescriptor)
	}
	__antithesis_instrumentation__.Notify(264433)
	return ret, nil
}

func (d *direct) ResolveSchemaID(
	ctx context.Context, txn *kv.Txn, dbID descpb.ID, scName string,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(264437)
	if !d.version.IsActive(clusterversion.PublicSchemasWithDescriptors) {
		__antithesis_instrumentation__.Notify(264439)

		if scName == tree.PublicSchema {
			__antithesis_instrumentation__.Notify(264440)
			return keys.PublicSchemaID, nil
		} else {
			__antithesis_instrumentation__.Notify(264441)
		}
	} else {
		__antithesis_instrumentation__.Notify(264442)
	}
	__antithesis_instrumentation__.Notify(264438)
	return catkv.LookupID(ctx, txn, d.codec, dbID, keys.RootNamespaceID, scName)
}

func (d *direct) GetDescriptorCollidingWithObject(
	ctx context.Context, txn *kv.Txn, parentID descpb.ID, parentSchemaID descpb.ID, name string,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(264443)
	id, err := catkv.LookupID(ctx, txn, d.codec, parentID, parentSchemaID, name)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(264447)
		return id == descpb.InvalidID == true
	}() == true {
		__antithesis_instrumentation__.Notify(264448)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264449)
	}
	__antithesis_instrumentation__.Notify(264444)

	desc, err := catkv.MaybeGetDescriptorByID(ctx, d.version, d.codec, txn, nil, id, catalog.Any)
	if desc == nil && func() bool {
		__antithesis_instrumentation__.Notify(264450)
		return err == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(264451)
		return nil, errors.NewAssertionErrorWithWrappedErrf(
			catalog.ErrDescriptorNotFound,
			"parentID=%d parentSchemaID=%d name=%q has ID=%d",
			parentID, parentSchemaID, name, id)
	} else {
		__antithesis_instrumentation__.Notify(264452)
	}
	__antithesis_instrumentation__.Notify(264445)
	if err != nil {
		__antithesis_instrumentation__.Notify(264453)
		return nil, sqlerrors.WrapErrorWhileConstructingObjectAlreadyExistsErr(err)
	} else {
		__antithesis_instrumentation__.Notify(264454)
	}
	__antithesis_instrumentation__.Notify(264446)
	return desc, nil
}

func (d *direct) CheckObjectCollision(
	ctx context.Context,
	txn *kv.Txn,
	parentID descpb.ID,
	parentSchemaID descpb.ID,
	name tree.ObjectName,
) error {
	__antithesis_instrumentation__.Notify(264455)
	desc, err := d.GetDescriptorCollidingWithObject(ctx, txn, parentID, parentSchemaID, name.Object())
	if err != nil {
		__antithesis_instrumentation__.Notify(264458)
		return err
	} else {
		__antithesis_instrumentation__.Notify(264459)
	}
	__antithesis_instrumentation__.Notify(264456)
	if desc != nil {
		__antithesis_instrumentation__.Notify(264460)
		maybeQualifiedName := name.Object()
		if name.Catalog() != "" && func() bool {
			__antithesis_instrumentation__.Notify(264462)
			return name.Schema() != "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(264463)
			maybeQualifiedName = name.FQString()
		} else {
			__antithesis_instrumentation__.Notify(264464)
		}
		__antithesis_instrumentation__.Notify(264461)
		return sqlerrors.MakeObjectAlreadyExistsError(desc.DescriptorProto(), maybeQualifiedName)
	} else {
		__antithesis_instrumentation__.Notify(264465)
	}
	__antithesis_instrumentation__.Notify(264457)
	return nil
}

func (d *direct) LookupObjectID(
	ctx context.Context, txn *kv.Txn, dbID descpb.ID, schemaID descpb.ID, objectName string,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(264466)
	return catkv.LookupID(ctx, txn, d.codec, dbID, schemaID, objectName)
}

func (d *direct) LookupSchemaID(
	ctx context.Context, txn *kv.Txn, dbID descpb.ID, schemaName string,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(264467)
	return catkv.LookupID(ctx, txn, d.codec, dbID, keys.RootNamespaceID, schemaName)
}

func (d *direct) LookupDatabaseID(
	ctx context.Context, txn *kv.Txn, dbName string,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(264468)
	return catkv.LookupID(ctx, txn, d.codec, keys.RootNamespaceID, keys.RootNamespaceID, dbName)
}

func (d *direct) WriteNewDescToBatch(
	ctx context.Context, kvTrace bool, b *kv.Batch, desc catalog.Descriptor,
) error {
	__antithesis_instrumentation__.Notify(264469)
	descKey := catalogkeys.MakeDescMetadataKey(d.codec, desc.GetID())
	proto := desc.DescriptorProto()
	if kvTrace {
		__antithesis_instrumentation__.Notify(264471)
		log.VEventf(ctx, 2, "CPut %s -> %s", descKey, proto)
	} else {
		__antithesis_instrumentation__.Notify(264472)
	}
	__antithesis_instrumentation__.Notify(264470)
	b.CPut(descKey, proto, nil)
	return nil
}
