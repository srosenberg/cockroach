package desctestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/catkv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/errors"
)

var (
	latestBinaryVersion = clusterversion.TestingClusterVersion
)

func TestingGetDatabaseDescriptorWitVersion(
	kvDB *kv.DB, codec keys.SQLCodec, version clusterversion.ClusterVersion, database string,
) catalog.DatabaseDescriptor {
	__antithesis_instrumentation__.Notify(265355)
	ctx := context.Background()
	var desc catalog.Descriptor
	if err := kvDB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) (err error) {
		__antithesis_instrumentation__.Notify(265357)
		id, err := catkv.LookupID(ctx, txn, codec, keys.RootNamespaceID, keys.RootNamespaceID, database)
		if err != nil {
			__antithesis_instrumentation__.Notify(265360)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(265361)
			if id == descpb.InvalidID {
				__antithesis_instrumentation__.Notify(265362)
				panic(fmt.Sprintf("database %s not found", database))
			} else {
				__antithesis_instrumentation__.Notify(265363)
			}
		}
		__antithesis_instrumentation__.Notify(265358)
		desc, err = catkv.MustGetDescriptorByID(ctx, version, codec, txn, nil, id, catalog.Database)
		if err != nil {
			__antithesis_instrumentation__.Notify(265364)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(265365)
		}
		__antithesis_instrumentation__.Notify(265359)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(265366)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(265367)
	}
	__antithesis_instrumentation__.Notify(265356)
	return desc.(catalog.DatabaseDescriptor)
}

func TestingGetDatabaseDescriptor(
	kvDB *kv.DB, codec keys.SQLCodec, database string,
) catalog.DatabaseDescriptor {
	__antithesis_instrumentation__.Notify(265368)
	return TestingGetDatabaseDescriptorWitVersion(kvDB, codec, latestBinaryVersion, database)
}

func TestingGetSchemaDescriptorWithVersion(
	kvDB *kv.DB,
	codec keys.SQLCodec,
	version clusterversion.ClusterVersion,
	dbID descpb.ID,
	schemaName string,
) catalog.SchemaDescriptor {
	__antithesis_instrumentation__.Notify(265369)
	ctx := context.Background()
	var desc catalog.Descriptor
	if err := kvDB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) (err error) {
		__antithesis_instrumentation__.Notify(265371)
		schemaID, err := catkv.LookupID(ctx, txn, codec, dbID, keys.RootNamespaceID, schemaName)
		if err != nil {
			__antithesis_instrumentation__.Notify(265374)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(265375)
			if schemaID == descpb.InvalidID {
				__antithesis_instrumentation__.Notify(265376)
				panic(fmt.Sprintf("schema %s not found", schemaName))
			} else {
				__antithesis_instrumentation__.Notify(265377)
			}
		}
		__antithesis_instrumentation__.Notify(265372)
		desc, err = catkv.MustGetDescriptorByID(ctx, version, codec, txn, nil, schemaID, catalog.Schema)
		if err != nil {
			__antithesis_instrumentation__.Notify(265378)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(265379)
		}
		__antithesis_instrumentation__.Notify(265373)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(265380)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(265381)
	}
	__antithesis_instrumentation__.Notify(265370)
	return desc.(catalog.SchemaDescriptor)
}

func TestingGetSchemaDescriptor(
	kvDB *kv.DB, codec keys.SQLCodec, dbID descpb.ID, schemaName string,
) catalog.SchemaDescriptor {
	__antithesis_instrumentation__.Notify(265382)
	return TestingGetSchemaDescriptorWithVersion(
		kvDB,
		codec,
		latestBinaryVersion,
		dbID,
		schemaName,
	)
}

func TestingGetTableDescriptorWithVersion(
	kvDB *kv.DB,
	codec keys.SQLCodec,
	version clusterversion.ClusterVersion,
	database string,
	schema string,
	table string,
) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(265383)
	return testingGetObjectDescriptor(kvDB, codec, version, database, schema, table).(catalog.TableDescriptor)
}

func TestingGetTableDescriptor(
	kvDB *kv.DB, codec keys.SQLCodec, database string, schema string, table string,
) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(265384)
	return TestingGetTableDescriptorWithVersion(kvDB, codec, latestBinaryVersion, database, schema, table)
}

func TestingGetPublicTableDescriptor(
	kvDB *kv.DB, codec keys.SQLCodec, database string, table string,
) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(265385)
	return testingGetObjectDescriptor(kvDB, codec, latestBinaryVersion, database, "public", table).(catalog.TableDescriptor)
}

func TestingGetMutableExistingTableDescriptor(
	kvDB *kv.DB, codec keys.SQLCodec, database string, table string,
) *tabledesc.Mutable {
	__antithesis_instrumentation__.Notify(265386)
	imm := TestingGetPublicTableDescriptor(kvDB, codec, database, table)
	return tabledesc.NewBuilder(imm.TableDesc()).BuildExistingMutableTable()
}

func TestingGetTypeDescriptor(
	kvDB *kv.DB, codec keys.SQLCodec, database string, schema string, object string,
) catalog.TypeDescriptor {
	__antithesis_instrumentation__.Notify(265387)
	return testingGetObjectDescriptor(kvDB, codec, latestBinaryVersion, database, schema, object).(catalog.TypeDescriptor)
}

func TestingGetPublicTypeDescriptor(
	kvDB *kv.DB, codec keys.SQLCodec, database string, object string,
) catalog.TypeDescriptor {
	__antithesis_instrumentation__.Notify(265388)
	return TestingGetTypeDescriptor(kvDB, codec, database, "public", object)
}

func testingGetObjectDescriptor(
	kvDB *kv.DB,
	codec keys.SQLCodec,
	version clusterversion.ClusterVersion,
	database string,
	schema string,
	object string,
) (desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(265389)
	ctx := context.Background()
	if err := kvDB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) (err error) {
		__antithesis_instrumentation__.Notify(265391)
		dbID, err := catkv.LookupID(ctx, txn, codec, keys.RootNamespaceID, keys.RootNamespaceID, database)
		if err != nil {
			__antithesis_instrumentation__.Notify(265398)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265399)
		}
		__antithesis_instrumentation__.Notify(265392)
		if dbID == descpb.InvalidID {
			__antithesis_instrumentation__.Notify(265400)
			return errors.Errorf("database %s not found", database)
		} else {
			__antithesis_instrumentation__.Notify(265401)
		}
		__antithesis_instrumentation__.Notify(265393)
		schemaID, err := catkv.LookupID(ctx, txn, codec, dbID, keys.RootNamespaceID, schema)
		if err != nil {
			__antithesis_instrumentation__.Notify(265402)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265403)
		}
		__antithesis_instrumentation__.Notify(265394)
		if schemaID == descpb.InvalidID {
			__antithesis_instrumentation__.Notify(265404)
			return errors.Errorf("schema %s not found", schema)
		} else {
			__antithesis_instrumentation__.Notify(265405)
		}
		__antithesis_instrumentation__.Notify(265395)
		objectID, err := catkv.LookupID(ctx, txn, codec, dbID, schemaID, object)
		if err != nil {
			__antithesis_instrumentation__.Notify(265406)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265407)
		}
		__antithesis_instrumentation__.Notify(265396)
		if objectID == descpb.InvalidID {
			__antithesis_instrumentation__.Notify(265408)
			return errors.Errorf("object %s not found", object)
		} else {
			__antithesis_instrumentation__.Notify(265409)
		}
		__antithesis_instrumentation__.Notify(265397)
		desc, err = catkv.MustGetDescriptorByID(ctx, latestBinaryVersion, codec, txn, nil, objectID, catalog.Any)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(265410)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(265411)
	}
	__antithesis_instrumentation__.Notify(265390)
	return desc
}
