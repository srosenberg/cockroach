package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/faketreeeval"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type fkHandler struct {
	allowed  bool
	skip     bool
	resolver fkResolver
}

var NoFKs = fkHandler{resolver: fkResolver{
	tableNameToDesc: make(map[string]*tabledesc.Mutable),
}}

func MakeTestingSimpleTableDescriptor(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	st *cluster.Settings,
	create *tree.CreateTable,
	parentID, parentSchemaID, tableID descpb.ID,
	fks fkHandler,
	walltime int64,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(494763)
	db := dbdesc.NewInitial(parentID, "foo", security.RootUserName())
	var sc catalog.SchemaDescriptor
	if !st.Version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) && func() bool {
		__antithesis_instrumentation__.Notify(494765)
		return parentSchemaID == keys.PublicSchemaIDForBackup == true
	}() == true {
		__antithesis_instrumentation__.Notify(494766)

		sc = schemadesc.GetPublicSchema()
	} else {
		__antithesis_instrumentation__.Notify(494767)
		sc = schemadesc.NewBuilder(&descpb.SchemaDescriptor{
			Name:     "foo",
			ID:       parentSchemaID,
			Version:  1,
			ParentID: parentID,
			Privileges: catpb.NewPrivilegeDescriptor(
				security.PublicRoleName(),
				privilege.SchemaPrivileges,
				privilege.List{},
				security.RootUserName(),
			),
		}).BuildCreatedMutableSchema()
	}
	__antithesis_instrumentation__.Notify(494764)
	return MakeSimpleTableDescriptor(ctx, semaCtx, st, create, db, sc, tableID, fks, walltime)
}

func makeSemaCtxWithoutTypeResolver(semaCtx *tree.SemaContext) *tree.SemaContext {
	__antithesis_instrumentation__.Notify(494768)
	semaCtxCopy := *semaCtx
	semaCtxCopy.TypeResolver = nil
	return &semaCtxCopy
}

func MakeSimpleTableDescriptor(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	st *cluster.Settings,
	create *tree.CreateTable,
	db catalog.DatabaseDescriptor,
	sc catalog.SchemaDescriptor,
	tableID descpb.ID,
	fks fkHandler,
	walltime int64,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(494769)
	create.HoistConstraints()
	if create.IfNotExists {
		__antithesis_instrumentation__.Notify(494775)
		return nil, unimplemented.NewWithIssueDetailf(42846, "import.if-no-exists", "unsupported IF NOT EXISTS")
	} else {
		__antithesis_instrumentation__.Notify(494776)
	}
	__antithesis_instrumentation__.Notify(494770)
	if create.AsSource != nil {
		__antithesis_instrumentation__.Notify(494777)
		return nil, unimplemented.NewWithIssueDetailf(42846, "import.create-as", "CREATE AS not supported")
	} else {
		__antithesis_instrumentation__.Notify(494778)
	}
	__antithesis_instrumentation__.Notify(494771)

	filteredDefs := create.Defs[:0]
	for i := range create.Defs {
		__antithesis_instrumentation__.Notify(494779)
		switch def := create.Defs[i].(type) {
		case *tree.CheckConstraintTableDef,
			*tree.FamilyTableDef,
			*tree.UniqueConstraintTableDef:
			__antithesis_instrumentation__.Notify(494781)

		case *tree.IndexTableDef:
			__antithesis_instrumentation__.Notify(494782)
			for i := range def.Columns {
				__antithesis_instrumentation__.Notify(494789)
				if def.Columns[i].Expr != nil {
					__antithesis_instrumentation__.Notify(494790)
					return nil, unimplemented.NewWithIssueDetail(56002, "import.expression-index",
						"to import into a table with expression indexes, use IMPORT INTO")
				} else {
					__antithesis_instrumentation__.Notify(494791)
				}
			}

		case *tree.ColumnTableDef:
			__antithesis_instrumentation__.Notify(494783)
			if def.IsComputed() && func() bool {
				__antithesis_instrumentation__.Notify(494792)
				return def.IsVirtual() == true
			}() == true {
				__antithesis_instrumentation__.Notify(494793)
				return nil, unimplemented.NewWithIssueDetail(56002, "import.computed",
					"to import into a table with virtual computed columns, use IMPORT INTO")
			} else {
				__antithesis_instrumentation__.Notify(494794)
			}
			__antithesis_instrumentation__.Notify(494784)

			if err := sql.SimplifySerialInColumnDefWithRowID(ctx, def, &create.Table); err != nil {
				__antithesis_instrumentation__.Notify(494795)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(494796)
			}

		case *tree.ForeignKeyConstraintTableDef:
			__antithesis_instrumentation__.Notify(494785)
			if !fks.allowed {
				__antithesis_instrumentation__.Notify(494797)
				return nil, unimplemented.NewWithIssueDetailf(42846, "import.fk",
					"this IMPORT format does not support foreign keys")
			} else {
				__antithesis_instrumentation__.Notify(494798)
			}
			__antithesis_instrumentation__.Notify(494786)
			if fks.skip {
				__antithesis_instrumentation__.Notify(494799)
				continue
			} else {
				__antithesis_instrumentation__.Notify(494800)
			}
			__antithesis_instrumentation__.Notify(494787)

			def.Table = tree.MakeUnqualifiedTableName(def.Table.ObjectName)

		default:
			__antithesis_instrumentation__.Notify(494788)
			return nil, unimplemented.Newf(fmt.Sprintf("import.%T", def), "unsupported table definition: %s", tree.AsString(def))
		}
		__antithesis_instrumentation__.Notify(494780)

		filteredDefs = append(filteredDefs, create.Defs[i])
	}
	__antithesis_instrumentation__.Notify(494772)
	create.Defs = filteredDefs

	evalCtx := tree.EvalContext{
		Context:            ctx,
		Sequence:           &importSequenceOperators{},
		Regions:            makeImportRegionOperator(""),
		SessionDataStack:   sessiondata.NewStack(&sessiondata.SessionData{}),
		ClientNoticeSender: &faketreeeval.DummyClientNoticeSender{},
		TxnTimestamp:       timeutil.Unix(0, walltime),
		Settings:           st,
	}
	affected := make(map[descpb.ID]*tabledesc.Mutable)

	tableDesc, err := sql.NewTableDesc(
		ctx,
		nil,
		&fks.resolver,
		st,
		create,
		db,
		sc,
		tableID,
		nil,
		hlc.Timestamp{WallTime: walltime},
		catpb.NewBasePrivilegeDescriptor(security.AdminRoleName()),
		affected,
		semaCtx,
		&evalCtx,
		evalCtx.SessionData(),
		tree.PersistencePermanent,

		sql.NewTableDescOptionBypassLocalityOnNonMultiRegionDatabaseCheck(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(494801)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(494802)
	}
	__antithesis_instrumentation__.Notify(494773)
	if err := fixDescriptorFKState(tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(494803)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(494804)
	}
	__antithesis_instrumentation__.Notify(494774)

	return tableDesc, nil
}

func fixDescriptorFKState(tableDesc *tabledesc.Mutable) error {
	__antithesis_instrumentation__.Notify(494805)
	tableDesc.SetPublic()
	for i := range tableDesc.OutboundFKs {
		__antithesis_instrumentation__.Notify(494807)
		tableDesc.OutboundFKs[i].Validity = descpb.ConstraintValidity_Unvalidated
	}
	__antithesis_instrumentation__.Notify(494806)
	return nil
}

var (
	errSequenceOperators = errors.New("sequence operations unsupported")
	errRegionOperator    = errors.New("region operations unsupported")
	errSchemaResolver    = errors.New("schema resolver unsupported")
)

type importRegionOperator struct {
	primaryRegion catpb.RegionName
}

func makeImportRegionOperator(primaryRegion catpb.RegionName) *importRegionOperator {
	__antithesis_instrumentation__.Notify(494808)
	return &importRegionOperator{primaryRegion: primaryRegion}
}

type importDatabaseRegionConfig struct {
	primaryRegion catpb.RegionName
}

func (i importDatabaseRegionConfig) IsValidRegionNameString(_ string) bool {
	__antithesis_instrumentation__.Notify(494809)

	return false
}

func (i importDatabaseRegionConfig) PrimaryRegionString() string {
	__antithesis_instrumentation__.Notify(494810)
	return string(i.primaryRegion)
}

var _ tree.DatabaseRegionConfig = &importDatabaseRegionConfig{}

func (so *importRegionOperator) CurrentDatabaseRegionConfig(
	_ context.Context,
) (tree.DatabaseRegionConfig, error) {
	__antithesis_instrumentation__.Notify(494811)
	return importDatabaseRegionConfig{primaryRegion: so.primaryRegion}, nil
}

func (so *importRegionOperator) ValidateAllMultiRegionZoneConfigsInCurrentDatabase(
	_ context.Context,
) error {
	__antithesis_instrumentation__.Notify(494812)
	return errors.WithStack(errRegionOperator)
}

func (so *importRegionOperator) ResetMultiRegionZoneConfigsForTable(
	_ context.Context, _ int64,
) error {
	__antithesis_instrumentation__.Notify(494813)
	return errors.WithStack(errRegionOperator)
}

func (so *importRegionOperator) ResetMultiRegionZoneConfigsForDatabase(
	_ context.Context, _ int64,
) error {
	__antithesis_instrumentation__.Notify(494814)
	return errors.WithStack(errRegionOperator)
}

type importSequenceOperators struct{}

func (so *importSequenceOperators) GetSerialSequenceNameFromColumn(
	ctx context.Context, tn *tree.TableName, columnName tree.Name,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(494815)
	return nil, errors.WithStack(errSequenceOperators)
}

func (so *importSequenceOperators) ParseQualifiedTableName(sql string) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(494816)
	name, err := parser.ParseTableName(sql)
	if err != nil {
		__antithesis_instrumentation__.Notify(494818)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(494819)
	}
	__antithesis_instrumentation__.Notify(494817)
	tn := name.ToTableName()
	return &tn, nil
}

func (so *importSequenceOperators) ResolveTableName(
	ctx context.Context, tn *tree.TableName,
) (tree.ID, error) {
	__antithesis_instrumentation__.Notify(494820)
	return 0, errSequenceOperators
}

func (so *importSequenceOperators) SchemaExists(
	ctx context.Context, dbName, scName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(494821)
	return false, errSequenceOperators
}

func (so *importSequenceOperators) IsTableVisible(
	ctx context.Context, curDB string, searchPath sessiondata.SearchPath, tableID oid.Oid,
) (bool, bool, error) {
	__antithesis_instrumentation__.Notify(494822)
	return false, false, errors.WithStack(errSequenceOperators)
}

func (so *importSequenceOperators) IsTypeVisible(
	ctx context.Context, curDB string, searchPath sessiondata.SearchPath, typeID oid.Oid,
) (bool, bool, error) {
	__antithesis_instrumentation__.Notify(494823)
	return false, false, errors.WithStack(errSequenceOperators)
}

func (so *importSequenceOperators) HasAnyPrivilege(
	ctx context.Context,
	specifier tree.HasPrivilegeSpecifier,
	user security.SQLUsername,
	privs []privilege.Privilege,
) (tree.HasAnyPrivilegeResult, error) {
	__antithesis_instrumentation__.Notify(494824)
	return tree.HasNoPrivilege, errors.WithStack(errSequenceOperators)
}

func (so *importSequenceOperators) IncrementSequenceByID(
	ctx context.Context, seqID int64,
) (int64, error) {
	__antithesis_instrumentation__.Notify(494825)
	return 0, errSequenceOperators
}

func (so *importSequenceOperators) GetLatestValueInSessionForSequenceByID(
	ctx context.Context, seqID int64,
) (int64, error) {
	__antithesis_instrumentation__.Notify(494826)
	return 0, errSequenceOperators
}

func (so *importSequenceOperators) SetSequenceValueByID(
	ctx context.Context, seqID uint32, newVal int64, isCalled bool,
) error {
	__antithesis_instrumentation__.Notify(494827)
	return errSequenceOperators
}

type fkResolver struct {
	tableNameToDesc map[string]*tabledesc.Mutable
	format          roachpb.IOFileFormat
}

var _ resolver.SchemaResolver = &fkResolver{}

func (r *fkResolver) Accessor() catalog.Accessor {
	__antithesis_instrumentation__.Notify(494828)
	return nil
}

func (r *fkResolver) CurrentDatabase() string {
	__antithesis_instrumentation__.Notify(494829)
	return ""
}

func (r *fkResolver) CurrentSearchPath() sessiondata.SearchPath {
	__antithesis_instrumentation__.Notify(494830)
	return sessiondata.SearchPath{}
}

func (r *fkResolver) CommonLookupFlags(required bool) tree.CommonLookupFlags {
	__antithesis_instrumentation__.Notify(494831)
	return tree.CommonLookupFlags{}
}

func (r *fkResolver) LookupObject(
	ctx context.Context, flags tree.ObjectLookupFlags, dbName, scName, obName string,
) (found bool, prefix catalog.ResolvedObjectPrefix, objMeta catalog.Descriptor, err error) {
	__antithesis_instrumentation__.Notify(494832)

	var lookupName string
	if r.format.Format == roachpb.IOFileFormat_PgDump {
		__antithesis_instrumentation__.Notify(494836)
		if scName == "" || func() bool {
			__antithesis_instrumentation__.Notify(494838)
			return dbName == "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(494839)
			return false, prefix, nil, errors.Errorf("expected catalog and schema name to be set when resolving"+
				" table %q in PGDUMP", obName)
		} else {
			__antithesis_instrumentation__.Notify(494840)
		}
		__antithesis_instrumentation__.Notify(494837)
		lookupName = fmt.Sprintf("%s.%s", scName, obName)
	} else {
		__antithesis_instrumentation__.Notify(494841)
		if scName != "" {
			__antithesis_instrumentation__.Notify(494842)
			lookupName = strings.TrimPrefix(obName, scName+".")
		} else {
			__antithesis_instrumentation__.Notify(494843)
		}
	}
	__antithesis_instrumentation__.Notify(494833)
	tbl, ok := r.tableNameToDesc[lookupName]
	if ok {
		__antithesis_instrumentation__.Notify(494844)
		return true, prefix, tbl, nil
	} else {
		__antithesis_instrumentation__.Notify(494845)
	}
	__antithesis_instrumentation__.Notify(494834)
	names := make([]string, 0, len(r.tableNameToDesc))
	for k := range r.tableNameToDesc {
		__antithesis_instrumentation__.Notify(494846)
		names = append(names, k)
	}
	__antithesis_instrumentation__.Notify(494835)
	suggestions := strings.Join(names, ",")
	return false, prefix, nil, errors.Errorf("referenced table %q not found in tables being imported (%s)",
		lookupName, suggestions)
}

func (r fkResolver) LookupSchema(
	ctx context.Context, dbName, scName string,
) (found bool, scMeta catalog.ResolvedObjectPrefix, err error) {
	__antithesis_instrumentation__.Notify(494847)
	return false, scMeta, errSchemaResolver
}

func (r fkResolver) ResolveTypeByOID(ctx context.Context, oid oid.Oid) (*types.T, error) {
	__antithesis_instrumentation__.Notify(494848)
	return nil, errSchemaResolver
}

func (r fkResolver) ResolveType(
	ctx context.Context, name *tree.UnresolvedObjectName,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(494849)
	return nil, errSchemaResolver
}

func (r fkResolver) GetQualifiedTableNameByID(
	ctx context.Context, id int64, requiredType tree.RequiredTableKind,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(494850)
	return nil, errSchemaResolver
}
