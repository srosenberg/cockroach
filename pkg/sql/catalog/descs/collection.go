// Package descs provides abstractions for dealing with sets of descriptors.
// It is utilized during schema changes and by catalog.Accessor implementations.
package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/hydratedtables"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/validate"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func makeCollection(
	ctx context.Context,
	leaseMgr *lease.Manager,
	settings *cluster.Settings,
	codec keys.SQLCodec,
	hydratedTables *hydratedtables.Cache,
	systemNamespace *systemDatabaseNamespaceCache,
	virtualSchemas catalog.VirtualSchemas,
	temporarySchemaProvider TemporarySchemaProvider,
) Collection {
	__antithesis_instrumentation__.Notify(264075)
	return Collection{
		settings:       settings,
		version:        settings.Version.ActiveVersion(ctx),
		hydratedTables: hydratedTables,
		virtual:        makeVirtualDescriptors(virtualSchemas),
		leased:         makeLeasedDescriptors(leaseMgr),
		kv:             makeKVDescriptors(codec, systemNamespace),
		temporary:      makeTemporaryDescriptors(settings, codec, temporarySchemaProvider),
		direct:         makeDirect(ctx, codec, settings),
	}
}

type Collection struct {
	settings *cluster.Settings

	version clusterversion.ClusterVersion

	virtual virtualDescriptors

	leased leasedDescriptors

	uncommitted uncommittedDescriptors

	kv kvDescriptors

	synthetic syntheticDescriptors

	temporary temporaryDescriptors

	hydratedTables *hydratedtables.Cache

	skipValidationOnWrite bool

	deletedDescs catalog.DescriptorIDSet

	maxTimestampBoundDeadlineHolder maxTimestampBoundDeadlineHolder

	sqlLivenessSession sqlliveness.Session

	direct direct
}

var _ catalog.Accessor = (*Collection)(nil)

func (tc *Collection) MaybeUpdateDeadline(ctx context.Context, txn *kv.Txn) (err error) {
	__antithesis_instrumentation__.Notify(264076)
	return tc.leased.maybeUpdateDeadline(ctx, txn, tc.sqlLivenessSession)
}

func (tc *Collection) SetMaxTimestampBound(maxTimestampBound hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(264077)
	tc.maxTimestampBoundDeadlineHolder.maxTimestampBound = maxTimestampBound
}

func (tc *Collection) ResetMaxTimestampBound() {
	__antithesis_instrumentation__.Notify(264078)
	tc.maxTimestampBoundDeadlineHolder.maxTimestampBound = hlc.Timestamp{}
}

func (tc *Collection) SkipValidationOnWrite() {
	__antithesis_instrumentation__.Notify(264079)
	tc.skipValidationOnWrite = true
}

func (tc *Collection) ReleaseSpecifiedLeases(ctx context.Context, descs []lease.IDVersion) {
	__antithesis_instrumentation__.Notify(264080)
	tc.leased.release(ctx, descs)
}

func (tc *Collection) ReleaseLeases(ctx context.Context) {
	__antithesis_instrumentation__.Notify(264081)
	tc.leased.releaseAll(ctx)

	tc.sqlLivenessSession = nil
}

func (tc *Collection) ReleaseAll(ctx context.Context) {
	__antithesis_instrumentation__.Notify(264082)
	tc.ReleaseLeases(ctx)
	tc.uncommitted.reset()
	tc.kv.reset()
	tc.synthetic.reset()
	tc.deletedDescs = catalog.DescriptorIDSet{}
	tc.skipValidationOnWrite = false
}

func (tc *Collection) HasUncommittedTables() bool {
	__antithesis_instrumentation__.Notify(264083)
	return tc.uncommitted.hasUncommittedTables()
}

func (tc *Collection) HasUncommittedTypes() bool {
	__antithesis_instrumentation__.Notify(264084)
	return tc.uncommitted.hasUncommittedTypes()
}

func (tc *Collection) AddUncommittedDescriptor(desc catalog.MutableDescriptor) error {
	__antithesis_instrumentation__.Notify(264085)

	tc.kv.releaseAllDescriptors()
	return tc.uncommitted.checkIn(desc)
}

var ValidateOnWriteEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.catalog.descs.validate_on_write.enabled",
	"set to true to validate descriptors prior to writing, false to disable; default is true",
	true,
)

func (tc *Collection) WriteDescToBatch(
	ctx context.Context, kvTrace bool, desc catalog.MutableDescriptor, b *kv.Batch,
) error {
	__antithesis_instrumentation__.Notify(264086)
	desc.MaybeIncrementVersion()
	if !tc.skipValidationOnWrite && func() bool {
		__antithesis_instrumentation__.Notify(264090)
		return ValidateOnWriteEnabled.Get(&tc.settings.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(264091)
		if err := validate.Self(tc.version, desc); err != nil {
			__antithesis_instrumentation__.Notify(264092)
			return err
		} else {
			__antithesis_instrumentation__.Notify(264093)
		}
	} else {
		__antithesis_instrumentation__.Notify(264094)
	}
	__antithesis_instrumentation__.Notify(264087)
	if err := tc.AddUncommittedDescriptor(desc); err != nil {
		__antithesis_instrumentation__.Notify(264095)
		return err
	} else {
		__antithesis_instrumentation__.Notify(264096)
	}
	__antithesis_instrumentation__.Notify(264088)
	descKey := catalogkeys.MakeDescMetadataKey(tc.codec(), desc.GetID())
	proto := desc.DescriptorProto()
	if kvTrace {
		__antithesis_instrumentation__.Notify(264097)
		log.VEventf(ctx, 2, "Put %s -> %s", descKey, proto)
	} else {
		__antithesis_instrumentation__.Notify(264098)
	}
	__antithesis_instrumentation__.Notify(264089)
	b.Put(descKey, proto)
	return nil
}

func (tc *Collection) WriteDesc(
	ctx context.Context, kvTrace bool, desc catalog.MutableDescriptor, txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(264099)
	b := txn.NewBatch()
	if err := tc.WriteDescToBatch(ctx, kvTrace, desc, b); err != nil {
		__antithesis_instrumentation__.Notify(264101)
		return err
	} else {
		__antithesis_instrumentation__.Notify(264102)
	}
	__antithesis_instrumentation__.Notify(264100)
	return txn.Run(ctx, b)
}

func (tc *Collection) GetDescriptorsWithNewVersion() (originalVersions []lease.IDVersion) {
	__antithesis_instrumentation__.Notify(264103)
	_ = tc.uncommitted.iterateNewVersionByID(func(originalVersion lease.IDVersion) error {
		__antithesis_instrumentation__.Notify(264105)
		originalVersions = append(originalVersions, originalVersion)
		return nil
	})
	__antithesis_instrumentation__.Notify(264104)
	return originalVersions
}

func (tc *Collection) GetUncommittedTables() (tables []catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(264106)
	return tc.uncommitted.getUncommittedTables()
}

func newMutableSyntheticDescriptorAssertionError(id descpb.ID) error {
	__antithesis_instrumentation__.Notify(264107)
	return errors.AssertionFailedf("attempted mutable access of synthetic descriptor %d", id)
}

func (tc *Collection) GetAllDescriptors(ctx context.Context, txn *kv.Txn) (nstree.Catalog, error) {
	__antithesis_instrumentation__.Notify(264108)
	return tc.kv.getAllDescriptors(ctx, txn, tc.version)
}

func (tc *Collection) GetAllDatabaseDescriptors(
	ctx context.Context, txn *kv.Txn,
) ([]catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(264109)
	vd := tc.newValidationDereferencer(txn)
	return tc.kv.getAllDatabaseDescriptors(ctx, tc.version, txn, vd)
}

func (tc *Collection) GetAllTableDescriptorsInDatabase(
	ctx context.Context, txn *kv.Txn, dbID descpb.ID,
) ([]catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(264110)

	found, _, err := tc.getDatabaseByID(ctx, txn, dbID, tree.DatabaseLookupFlags{
		AvoidLeased: false,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(264115)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264116)
	}
	__antithesis_instrumentation__.Notify(264111)
	if !found {
		__antithesis_instrumentation__.Notify(264117)
		return nil, sqlerrors.NewUndefinedDatabaseError(fmt.Sprintf("[%d]", dbID))
	} else {
		__antithesis_instrumentation__.Notify(264118)
	}
	__antithesis_instrumentation__.Notify(264112)
	all, err := tc.GetAllDescriptors(ctx, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(264119)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264120)
	}
	__antithesis_instrumentation__.Notify(264113)
	var ret []catalog.TableDescriptor
	for _, desc := range all.OrderedDescriptors() {
		__antithesis_instrumentation__.Notify(264121)
		if desc.GetParentID() == dbID {
			__antithesis_instrumentation__.Notify(264122)
			if table, ok := desc.(catalog.TableDescriptor); ok {
				__antithesis_instrumentation__.Notify(264123)
				ret = append(ret, table)
			} else {
				__antithesis_instrumentation__.Notify(264124)
			}
		} else {
			__antithesis_instrumentation__.Notify(264125)
		}
	}
	__antithesis_instrumentation__.Notify(264114)
	return ret, nil
}

func (tc *Collection) GetSchemasForDatabase(
	ctx context.Context, txn *kv.Txn, dbDesc catalog.DatabaseDescriptor,
) (map[descpb.ID]string, error) {
	__antithesis_instrumentation__.Notify(264126)
	return tc.kv.getSchemasForDatabase(ctx, txn, dbDesc)
}

func (tc *Collection) GetObjectNamesAndIDs(
	ctx context.Context,
	txn *kv.Txn,
	dbDesc catalog.DatabaseDescriptor,
	scName string,
	flags tree.DatabaseListFlags,
) (tree.TableNames, descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(264127)
	if ok, names, ds := tc.virtual.maybeGetObjectNamesAndIDs(
		scName, dbDesc, flags,
	); ok {
		__antithesis_instrumentation__.Notify(264133)
		return names, ds, nil
	} else {
		__antithesis_instrumentation__.Notify(264134)
	}
	__antithesis_instrumentation__.Notify(264128)

	schemaFlags := tree.SchemaLookupFlags{
		Required: flags.Required,
		AvoidLeased: flags.RequireMutable || func() bool {
			__antithesis_instrumentation__.Notify(264135)
			return flags.AvoidLeased == true
		}() == true,
		IncludeDropped: flags.IncludeDropped,
		IncludeOffline: flags.IncludeOffline,
	}
	schema, err := tc.getSchemaByName(ctx, txn, dbDesc, scName, schemaFlags)
	if err != nil {
		__antithesis_instrumentation__.Notify(264136)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(264137)
	}
	__antithesis_instrumentation__.Notify(264129)
	if schema == nil {
		__antithesis_instrumentation__.Notify(264138)
		return nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(264139)
	}
	__antithesis_instrumentation__.Notify(264130)

	log.Eventf(ctx, "fetching list of objects for %q", dbDesc.GetName())
	prefix := catalogkeys.MakeObjectNameKey(tc.codec(), dbDesc.GetID(), schema.GetID(), "")
	sr, err := txn.Scan(ctx, prefix, prefix.PrefixEnd(), 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(264140)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(264141)
	}
	__antithesis_instrumentation__.Notify(264131)

	alreadySeen := make(map[string]bool)
	var tableNames tree.TableNames
	var tableIDs descpb.IDs

	for _, row := range sr {
		__antithesis_instrumentation__.Notify(264142)
		_, tableName, err := encoding.DecodeUnsafeStringAscending(bytes.TrimPrefix(
			row.Key, prefix), nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(264144)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(264145)
		}
		__antithesis_instrumentation__.Notify(264143)
		alreadySeen[tableName] = true
		tn := tree.MakeTableNameWithSchema(tree.Name(dbDesc.GetName()), tree.Name(scName), tree.Name(tableName))
		tn.ExplicitCatalog = flags.ExplicitPrefix
		tn.ExplicitSchema = flags.ExplicitPrefix
		tableNames = append(tableNames, tn)
		tableIDs = append(tableIDs, descpb.ID(row.ValueInt()))
	}
	__antithesis_instrumentation__.Notify(264132)

	return tableNames, tableIDs, nil
}

func (tc *Collection) SetSyntheticDescriptors(descs []catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(264146)
	tc.synthetic.set(descs)
}

func (tc *Collection) AddSyntheticDescriptor(desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(264147)
	tc.synthetic.add(desc)
}

func (tc *Collection) RemoveSyntheticDescriptor(id descpb.ID) {
	__antithesis_instrumentation__.Notify(264148)
	tc.synthetic.remove(id)
}

func (tc *Collection) codec() keys.SQLCodec {
	__antithesis_instrumentation__.Notify(264149)
	return tc.kv.codec
}

func (tc *Collection) AddDeletedDescriptor(id descpb.ID) {
	__antithesis_instrumentation__.Notify(264150)
	tc.deletedDescs.Add(id)
}

func (tc *Collection) SetSession(session sqlliveness.Session) {
	__antithesis_instrumentation__.Notify(264151)
	tc.sqlLivenessSession = session
}
