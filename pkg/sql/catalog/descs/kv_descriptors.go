package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/catkv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/validate"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type kvDescriptors struct {
	codec keys.SQLCodec

	systemNamespace *systemDatabaseNamespaceCache

	allDescriptors allDescriptors

	allDatabaseDescriptors []catalog.DatabaseDescriptor

	allSchemasForDatabase map[descpb.ID]map[descpb.ID]string
}

type allDescriptors struct {
	c nstree.Catalog
}

func (d *allDescriptors) init(c nstree.Catalog) {
	d.c = c
}

func (d *allDescriptors) clear() {
	__antithesis_instrumentation__.Notify(264577)
	*d = allDescriptors{}
}

func (d *allDescriptors) isUnset() bool {
	__antithesis_instrumentation__.Notify(264578)
	return !d.c.IsInitialized()
}

func (d *allDescriptors) contains(id descpb.ID) bool {
	__antithesis_instrumentation__.Notify(264579)
	return d.c.IsInitialized() && func() bool {
		__antithesis_instrumentation__.Notify(264580)
		return d.c.LookupDescriptorEntry(id) != nil == true
	}() == true
}

func makeKVDescriptors(
	codec keys.SQLCodec, systemNamespace *systemDatabaseNamespaceCache,
) kvDescriptors {
	__antithesis_instrumentation__.Notify(264581)
	return kvDescriptors{
		codec:           codec,
		systemNamespace: systemNamespace,
	}
}

func (kd *kvDescriptors) reset() {
	__antithesis_instrumentation__.Notify(264582)
	kd.releaseAllDescriptors()
}

func (kd *kvDescriptors) releaseAllDescriptors() {
	__antithesis_instrumentation__.Notify(264583)
	kd.allDescriptors.clear()
	kd.allDatabaseDescriptors = nil
	kd.allSchemasForDatabase = nil
}

func (kd *kvDescriptors) lookupName(
	ctx context.Context,
	txn *kv.Txn,
	maybeDB catalog.DatabaseDescriptor,
	parentID descpb.ID,
	parentSchemaID descpb.ID,
	name string,
) (id descpb.ID, err error) {
	__antithesis_instrumentation__.Notify(264584)

	switch parentID {
	case descpb.InvalidID:
		__antithesis_instrumentation__.Notify(264586)
		if name == systemschema.SystemDatabaseName {
			__antithesis_instrumentation__.Notify(264590)

			return keys.SystemDatabaseID, nil
		} else {
			__antithesis_instrumentation__.Notify(264591)
		}
	case keys.SystemDatabaseID:
		__antithesis_instrumentation__.Notify(264587)

		id = kd.systemNamespace.lookup(parentSchemaID, name)
		if id != descpb.InvalidID {
			__antithesis_instrumentation__.Notify(264592)
			return id, err
		} else {
			__antithesis_instrumentation__.Notify(264593)
		}
		__antithesis_instrumentation__.Notify(264588)

		defer func() {
			__antithesis_instrumentation__.Notify(264594)
			if err == nil && func() bool {
				__antithesis_instrumentation__.Notify(264595)
				return id != descpb.InvalidID == true
			}() == true {
				__antithesis_instrumentation__.Notify(264596)
				kd.systemNamespace.add(descpb.NameInfo{
					ParentID:       keys.SystemDatabaseID,
					ParentSchemaID: parentSchemaID,
					Name:           name,
				}, id)
			} else {
				__antithesis_instrumentation__.Notify(264597)
			}
		}()
	default:
		__antithesis_instrumentation__.Notify(264589)
		if parentSchemaID == descpb.InvalidID {
			__antithesis_instrumentation__.Notify(264598)

			if maybeDB != nil {
				__antithesis_instrumentation__.Notify(264599)

				id := maybeDB.GetSchemaID(name)
				return id, nil
			} else {
				__antithesis_instrumentation__.Notify(264600)
			}
		} else {
			__antithesis_instrumentation__.Notify(264601)
		}
	}
	__antithesis_instrumentation__.Notify(264585)

	return catkv.LookupID(
		ctx, txn, kd.codec, parentID, parentSchemaID, name,
	)
}

func (kd *kvDescriptors) getByName(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	txn *kv.Txn,
	vd validate.ValidationDereferencer,
	maybeDB catalog.DatabaseDescriptor,
	parentID descpb.ID,
	parentSchemaID descpb.ID,
	name string,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(264602)
	descID, err := kd.lookupName(ctx, txn, maybeDB, parentID, parentSchemaID, name)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(264606)
		return descID == descpb.InvalidID == true
	}() == true {
		__antithesis_instrumentation__.Notify(264607)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264608)
	}
	__antithesis_instrumentation__.Notify(264603)
	descs, err := kd.getByIDs(ctx, version, txn, vd, []descpb.ID{descID})
	if err != nil {
		__antithesis_instrumentation__.Notify(264609)
		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(264611)

			return nil, errors.WithAssertionFailure(err)
		} else {
			__antithesis_instrumentation__.Notify(264612)
		}
		__antithesis_instrumentation__.Notify(264610)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264613)
	}
	__antithesis_instrumentation__.Notify(264604)
	if descs[0].GetName() != name {
		__antithesis_instrumentation__.Notify(264614)

		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(264615)
	}
	__antithesis_instrumentation__.Notify(264605)
	return descs[0], nil
}

func (kd *kvDescriptors) getByIDs(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	txn *kv.Txn,
	vd validate.ValidationDereferencer,
	ids []descpb.ID,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(264616)
	ret := make([]catalog.Descriptor, len(ids))
	kvIDs := make([]descpb.ID, 0, len(ids))
	indexes := make([]int, 0, len(ids))
	for i, id := range ids {
		__antithesis_instrumentation__.Notify(264621)
		if id == keys.SystemDatabaseID {
			__antithesis_instrumentation__.Notify(264622)

			ret[i] = dbdesc.NewBuilder(systemschema.SystemDB.DatabaseDesc()).BuildExistingMutable()
		} else {
			__antithesis_instrumentation__.Notify(264623)
			kvIDs = append(kvIDs, id)
			indexes = append(indexes, i)
		}
	}
	__antithesis_instrumentation__.Notify(264617)
	if len(kvIDs) == 0 {
		__antithesis_instrumentation__.Notify(264624)
		return ret, nil
	} else {
		__antithesis_instrumentation__.Notify(264625)
	}
	__antithesis_instrumentation__.Notify(264618)
	kvDescs, err := catkv.MustGetDescriptorsByID(ctx, version, kd.codec, txn, vd, kvIDs, catalog.Any)
	if err != nil {
		__antithesis_instrumentation__.Notify(264626)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264627)
	}
	__antithesis_instrumentation__.Notify(264619)
	for j, desc := range kvDescs {
		__antithesis_instrumentation__.Notify(264628)
		ret[indexes[j]] = desc
	}
	__antithesis_instrumentation__.Notify(264620)
	return ret, nil
}

func (kd *kvDescriptors) getAllDescriptors(
	ctx context.Context, txn *kv.Txn, version clusterversion.ClusterVersion,
) (nstree.Catalog, error) {
	__antithesis_instrumentation__.Notify(264629)
	if kd.allDescriptors.isUnset() {
		__antithesis_instrumentation__.Notify(264631)
		c, err := catkv.GetCatalogUnvalidated(ctx, kd.codec, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(264635)
			return nstree.Catalog{}, err
		} else {
			__antithesis_instrumentation__.Notify(264636)
		}
		__antithesis_instrumentation__.Notify(264632)

		descs := c.OrderedDescriptors()
		ve := c.Validate(ctx, version, catalog.ValidationReadTelemetry, catalog.ValidationLevelCrossReferences, descs...)
		if err := ve.CombinedError(); err != nil {
			__antithesis_instrumentation__.Notify(264637)
			return nstree.Catalog{}, err
		} else {
			__antithesis_instrumentation__.Notify(264638)
		}
		__antithesis_instrumentation__.Notify(264633)

		if err := HydrateGivenDescriptors(ctx, descs); err != nil {
			__antithesis_instrumentation__.Notify(264639)

			log.Errorf(ctx, "%s", err.Error())
		} else {
			__antithesis_instrumentation__.Notify(264640)
		}
		__antithesis_instrumentation__.Notify(264634)

		kd.allDescriptors.init(c)
	} else {
		__antithesis_instrumentation__.Notify(264641)
	}
	__antithesis_instrumentation__.Notify(264630)
	return kd.allDescriptors.c, nil
}

func (kd *kvDescriptors) getAllDatabaseDescriptors(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	txn *kv.Txn,
	vd validate.ValidationDereferencer,
) ([]catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(264642)
	if kd.allDatabaseDescriptors == nil {
		__antithesis_instrumentation__.Notify(264644)
		c, err := catkv.GetAllDatabaseDescriptorIDs(ctx, txn, kd.codec)
		if err != nil {
			__antithesis_instrumentation__.Notify(264647)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(264648)
		}
		__antithesis_instrumentation__.Notify(264645)
		dbDescs, err := catkv.MustGetDescriptorsByID(ctx, version, kd.codec, txn, vd, c.OrderedDescriptorIDs(), catalog.Database)
		if err != nil {
			__antithesis_instrumentation__.Notify(264649)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(264650)
		}
		__antithesis_instrumentation__.Notify(264646)
		kd.allDatabaseDescriptors = make([]catalog.DatabaseDescriptor, len(dbDescs))
		for i, dbDesc := range dbDescs {
			__antithesis_instrumentation__.Notify(264651)
			kd.allDatabaseDescriptors[i] = dbDesc.(catalog.DatabaseDescriptor)
		}
	} else {
		__antithesis_instrumentation__.Notify(264652)
	}
	__antithesis_instrumentation__.Notify(264643)
	return kd.allDatabaseDescriptors, nil
}

func (kd *kvDescriptors) getSchemasForDatabase(
	ctx context.Context, txn *kv.Txn, db catalog.DatabaseDescriptor,
) (map[descpb.ID]string, error) {
	__antithesis_instrumentation__.Notify(264653)
	if kd.allSchemasForDatabase == nil {
		__antithesis_instrumentation__.Notify(264656)
		kd.allSchemasForDatabase = make(map[descpb.ID]map[descpb.ID]string)
	} else {
		__antithesis_instrumentation__.Notify(264657)
	}
	__antithesis_instrumentation__.Notify(264654)
	if _, ok := kd.allSchemasForDatabase[db.GetID()]; !ok {
		__antithesis_instrumentation__.Notify(264658)
		var err error
		allSchemas, err := resolver.GetForDatabase(ctx, txn, kd.codec, db)
		if err != nil {
			__antithesis_instrumentation__.Notify(264660)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(264661)
		}
		__antithesis_instrumentation__.Notify(264659)
		kd.allSchemasForDatabase[db.GetID()] = make(map[descpb.ID]string)
		for id, entry := range allSchemas {
			__antithesis_instrumentation__.Notify(264662)
			kd.allSchemasForDatabase[db.GetID()][id] = entry.Name
		}
	} else {
		__antithesis_instrumentation__.Notify(264663)
	}
	__antithesis_instrumentation__.Notify(264655)
	return kd.allSchemasForDatabase[db.GetID()], nil
}

func (kd *kvDescriptors) idDefinitelyDoesNotExist(id descpb.ID) bool {
	__antithesis_instrumentation__.Notify(264664)
	if kd.allDescriptors.isUnset() {
		__antithesis_instrumentation__.Notify(264666)
		return false
	} else {
		__antithesis_instrumentation__.Notify(264667)
	}
	__antithesis_instrumentation__.Notify(264665)
	return !kd.allDescriptors.contains(id)
}
