package catconstants

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/keys"

var StaticSchemaIDMapVirtualPublicSchema = map[uint32]string{
	keys.PublicSchemaID: PublicSchemaName,
	PgCatalogID:         PgCatalogName,
	InformationSchemaID: InformationSchemaName,
	CrdbInternalID:      CRDBInternalSchemaName,
	PgExtensionSchemaID: PgExtensionSchemaName,
}

var StaticSchemaIDMap = map[uint32]string{
	PgCatalogID:         PgCatalogName,
	InformationSchemaID: InformationSchemaName,
	CrdbInternalID:      CRDBInternalSchemaName,
	PgExtensionSchemaID: PgExtensionSchemaName,
}

func GetStaticSchemaIDMap() map[uint32]string {
	__antithesis_instrumentation__.Notify(247364)
	return StaticSchemaIDMapVirtualPublicSchema
}

const PgCatalogName = "pg_catalog"

const PublicSchemaName = "public"

const UserSchemaName = "$user"

const InformationSchemaName = "information_schema"

const CRDBInternalSchemaName = "crdb_internal"

const PgSchemaPrefix = "pg_"

const PgTempSchemaName = "pg_temp"

const PgExtensionSchemaName = "pg_extension"

var VirtualSchemaNames = map[string]struct{}{
	PgCatalogName:          {},
	InformationSchemaName:  {},
	CRDBInternalSchemaName: {},
	PgExtensionSchemaName:  {},
}
