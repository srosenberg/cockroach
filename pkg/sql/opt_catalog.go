package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/indexrec"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type optCatalog struct {
	planner *planner

	cfg *config.SystemConfig

	dataSources map[catalog.TableDescriptor]cat.DataSource

	tn tree.TableName
}

var _ cat.Catalog = &optCatalog{}

func (oc *optCatalog) init(planner *planner) {
	oc.planner = planner
	oc.dataSources = make(map[catalog.TableDescriptor]cat.DataSource)
}

func (oc *optCatalog) reset() {
	__antithesis_instrumentation__.Notify(550840)

	if len(oc.dataSources) > 100 {
		__antithesis_instrumentation__.Notify(550842)
		oc.dataSources = make(map[catalog.TableDescriptor]cat.DataSource)
	} else {
		__antithesis_instrumentation__.Notify(550843)
	}
	__antithesis_instrumentation__.Notify(550841)

	oc.cfg = oc.planner.execCfg.SystemConfig.GetSystemConfig()
}

type optSchema struct {
	planner *planner

	database catalog.DatabaseDescriptor
	schema   catalog.SchemaDescriptor

	name cat.SchemaName
}

func (os *optSchema) ID() cat.StableID {
	__antithesis_instrumentation__.Notify(550844)
	switch os.schema.SchemaKind() {
	case catalog.SchemaUserDefined, catalog.SchemaTemporary:
		__antithesis_instrumentation__.Notify(550845)

		return cat.StableID(os.schema.GetID())
	default:
		__antithesis_instrumentation__.Notify(550846)

		return cat.StableID(os.database.GetID())
	}
}

func (os *optSchema) PostgresDescriptorID() cat.StableID {
	__antithesis_instrumentation__.Notify(550847)
	return os.ID()
}

func (os *optSchema) Equals(other cat.Object) bool {
	__antithesis_instrumentation__.Notify(550848)
	otherSchema, ok := other.(*optSchema)
	return ok && func() bool {
		__antithesis_instrumentation__.Notify(550849)
		return os.ID() == otherSchema.ID() == true
	}() == true
}

func (os *optSchema) Name() *cat.SchemaName {
	__antithesis_instrumentation__.Notify(550850)
	return &os.name
}

func (os *optSchema) GetDataSourceNames(
	ctx context.Context,
) ([]cat.DataSourceName, descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(550851)
	return resolver.GetObjectNamesAndIDs(
		ctx,
		os.planner.Txn(),
		os.planner,
		os.planner.ExecCfg().Codec,
		os.database,
		os.name.Schema(),
		true,
	)
}

func (os *optSchema) getDescriptorForPermissionsCheck() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(550852)

	if os.schema.SchemaKind() == catalog.SchemaUserDefined {
		__antithesis_instrumentation__.Notify(550854)
		return os.schema
	} else {
		__antithesis_instrumentation__.Notify(550855)
	}
	__antithesis_instrumentation__.Notify(550853)

	return os.database
}

func (oc *optCatalog) ResolveSchema(
	ctx context.Context, flags cat.Flags, name *cat.SchemaName,
) (cat.Schema, cat.SchemaName, error) {
	__antithesis_instrumentation__.Notify(550856)
	if flags.AvoidDescriptorCaches {
		__antithesis_instrumentation__.Notify(550860)
		defer func(prev bool) {
			__antithesis_instrumentation__.Notify(550862)
			oc.planner.avoidLeasedDescriptors = prev
		}(oc.planner.avoidLeasedDescriptors)
		__antithesis_instrumentation__.Notify(550861)
		oc.planner.avoidLeasedDescriptors = true
	} else {
		__antithesis_instrumentation__.Notify(550863)
	}
	__antithesis_instrumentation__.Notify(550857)

	oc.tn.ObjectNamePrefix = *name
	found, prefix, err := resolver.ResolveObjectNamePrefix(
		ctx, oc.planner, oc.planner.CurrentDatabase(),
		oc.planner.CurrentSearchPath(), &oc.tn.ObjectNamePrefix,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(550864)
		return nil, cat.SchemaName{}, err
	} else {
		__antithesis_instrumentation__.Notify(550865)
	}
	__antithesis_instrumentation__.Notify(550858)
	if !found {
		__antithesis_instrumentation__.Notify(550866)
		if !name.ExplicitSchema && func() bool {
			__antithesis_instrumentation__.Notify(550868)
			return !name.ExplicitCatalog == true
		}() == true {
			__antithesis_instrumentation__.Notify(550869)
			return nil, cat.SchemaName{}, pgerror.New(
				pgcode.InvalidName, "no database or schema specified",
			)
		} else {
			__antithesis_instrumentation__.Notify(550870)
		}
		__antithesis_instrumentation__.Notify(550867)
		return nil, cat.SchemaName{}, pgerror.Newf(
			pgcode.InvalidSchemaName, "target database or schema does not exist",
		)
	} else {
		__antithesis_instrumentation__.Notify(550871)
	}
	__antithesis_instrumentation__.Notify(550859)

	return &optSchema{
		planner:  oc.planner,
		database: prefix.Database,
		schema:   prefix.Schema,
		name:     oc.tn.ObjectNamePrefix,
	}, oc.tn.ObjectNamePrefix, nil
}

func (oc *optCatalog) ResolveDataSource(
	ctx context.Context, flags cat.Flags, name *cat.DataSourceName,
) (cat.DataSource, cat.DataSourceName, error) {
	__antithesis_instrumentation__.Notify(550872)
	if flags.AvoidDescriptorCaches {
		__antithesis_instrumentation__.Notify(550877)
		defer func(prev bool) {
			__antithesis_instrumentation__.Notify(550879)
			oc.planner.avoidLeasedDescriptors = prev
		}(oc.planner.avoidLeasedDescriptors)
		__antithesis_instrumentation__.Notify(550878)
		oc.planner.avoidLeasedDescriptors = true
	} else {
		__antithesis_instrumentation__.Notify(550880)
	}
	__antithesis_instrumentation__.Notify(550873)

	oc.tn = *name
	lflags := tree.ObjectLookupFlagsWithRequiredTableKind(tree.ResolveAnyTableKind)
	prefix, desc, err := resolver.ResolveExistingTableObject(ctx, oc.planner, &oc.tn, lflags)
	if err != nil {
		__antithesis_instrumentation__.Notify(550881)
		return nil, cat.DataSourceName{}, err
	} else {
		__antithesis_instrumentation__.Notify(550882)
	}
	__antithesis_instrumentation__.Notify(550874)

	if err := oc.planner.canResolveDescUnderSchema(ctx, prefix.Schema, desc); err != nil {
		__antithesis_instrumentation__.Notify(550883)
		return nil, cat.DataSourceName{}, err
	} else {
		__antithesis_instrumentation__.Notify(550884)
	}
	__antithesis_instrumentation__.Notify(550875)

	ds, err := oc.dataSourceForDesc(ctx, flags, desc, &oc.tn)
	if err != nil {
		__antithesis_instrumentation__.Notify(550885)
		return nil, cat.DataSourceName{}, err
	} else {
		__antithesis_instrumentation__.Notify(550886)
	}
	__antithesis_instrumentation__.Notify(550876)
	return ds, oc.tn, nil
}

func (oc *optCatalog) ResolveDataSourceByID(
	ctx context.Context, flags cat.Flags, dataSourceID cat.StableID,
) (_ cat.DataSource, isAdding bool, _ error) {
	__antithesis_instrumentation__.Notify(550887)
	if flags.AvoidDescriptorCaches {
		__antithesis_instrumentation__.Notify(550890)
		defer func(prev bool) {
			__antithesis_instrumentation__.Notify(550892)
			oc.planner.avoidLeasedDescriptors = prev
		}(oc.planner.avoidLeasedDescriptors)
		__antithesis_instrumentation__.Notify(550891)
		oc.planner.avoidLeasedDescriptors = true
	} else {
		__antithesis_instrumentation__.Notify(550893)
	}
	__antithesis_instrumentation__.Notify(550888)

	tableLookup, err := oc.planner.LookupTableByID(ctx, descpb.ID(dataSourceID))

	if err != nil {
		__antithesis_instrumentation__.Notify(550894)
		isAdding := catalog.HasAddingTableError(err)
		if errors.Is(err, catalog.ErrDescriptorNotFound) || func() bool {
			__antithesis_instrumentation__.Notify(550896)
			return isAdding == true
		}() == true {
			__antithesis_instrumentation__.Notify(550897)
			return nil, isAdding, sqlerrors.NewUndefinedRelationError(&tree.TableRef{TableID: int64(dataSourceID)})
		} else {
			__antithesis_instrumentation__.Notify(550898)
		}
		__antithesis_instrumentation__.Notify(550895)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(550899)
	}
	__antithesis_instrumentation__.Notify(550889)

	ds, err := oc.dataSourceForDesc(ctx, cat.Flags{}, tableLookup, &tree.TableName{})
	return ds, false, err
}

func (oc *optCatalog) ResolveTypeByOID(ctx context.Context, oid oid.Oid) (*types.T, error) {
	__antithesis_instrumentation__.Notify(550900)
	return oc.planner.ResolveTypeByOID(ctx, oid)
}

func (oc *optCatalog) ResolveType(
	ctx context.Context, name *tree.UnresolvedObjectName,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(550901)
	return oc.planner.ResolveType(ctx, name)
}

func getDescFromCatalogObjectForPermissions(o cat.Object) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(550902)
	switch t := o.(type) {
	case *optSchema:
		__antithesis_instrumentation__.Notify(550903)
		return t.getDescriptorForPermissionsCheck(), nil
	case *optTable:
		__antithesis_instrumentation__.Notify(550904)
		return t.desc, nil
	case *optVirtualTable:
		__antithesis_instrumentation__.Notify(550905)
		return t.desc, nil
	case *optView:
		__antithesis_instrumentation__.Notify(550906)
		return t.desc, nil
	case *optSequence:
		__antithesis_instrumentation__.Notify(550907)
		return t.desc, nil
	default:
		__antithesis_instrumentation__.Notify(550908)
		return nil, errors.AssertionFailedf("invalid object type: %T", o)
	}
}

func getDescForDataSource(o cat.DataSource) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(550909)
	switch t := o.(type) {
	case *optTable:
		__antithesis_instrumentation__.Notify(550910)
		return t.desc, nil
	case *optVirtualTable:
		__antithesis_instrumentation__.Notify(550911)
		return t.desc, nil
	case *optView:
		__antithesis_instrumentation__.Notify(550912)
		return t.desc, nil
	case *optSequence:
		__antithesis_instrumentation__.Notify(550913)
		return t.desc, nil
	default:
		__antithesis_instrumentation__.Notify(550914)
		return nil, errors.AssertionFailedf("invalid object type: %T", o)
	}
}

func (oc *optCatalog) CheckPrivilege(ctx context.Context, o cat.Object, priv privilege.Kind) error {
	__antithesis_instrumentation__.Notify(550915)
	desc, err := getDescFromCatalogObjectForPermissions(o)
	if err != nil {
		__antithesis_instrumentation__.Notify(550917)
		return err
	} else {
		__antithesis_instrumentation__.Notify(550918)
	}
	__antithesis_instrumentation__.Notify(550916)
	return oc.planner.CheckPrivilege(ctx, desc, priv)
}

func (oc *optCatalog) CheckAnyPrivilege(ctx context.Context, o cat.Object) error {
	__antithesis_instrumentation__.Notify(550919)
	desc, err := getDescFromCatalogObjectForPermissions(o)
	if err != nil {
		__antithesis_instrumentation__.Notify(550921)
		return err
	} else {
		__antithesis_instrumentation__.Notify(550922)
	}
	__antithesis_instrumentation__.Notify(550920)
	return oc.planner.CheckAnyPrivilege(ctx, desc)
}

func (oc *optCatalog) HasAdminRole(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(550923)
	return oc.planner.HasAdminRole(ctx)
}

func (oc *optCatalog) RequireAdminRole(ctx context.Context, action string) error {
	__antithesis_instrumentation__.Notify(550924)
	return oc.planner.RequireAdminRole(ctx, action)
}

func (oc *optCatalog) HasRoleOption(
	ctx context.Context, roleOption roleoption.Option,
) (bool, error) {
	__antithesis_instrumentation__.Notify(550925)
	return oc.planner.HasRoleOption(ctx, roleOption)
}

func (oc *optCatalog) FullyQualifiedName(
	ctx context.Context, ds cat.DataSource,
) (cat.DataSourceName, error) {
	__antithesis_instrumentation__.Notify(550926)
	return oc.fullyQualifiedNameWithTxn(ctx, ds, oc.planner.Txn())
}

func (oc *optCatalog) fullyQualifiedNameWithTxn(
	ctx context.Context, ds cat.DataSource, txn *kv.Txn,
) (cat.DataSourceName, error) {
	__antithesis_instrumentation__.Notify(550927)
	if vt, ok := ds.(*optVirtualTable); ok {
		__antithesis_instrumentation__.Notify(550932)

		return vt.name, nil
	} else {
		__antithesis_instrumentation__.Notify(550933)
	}
	__antithesis_instrumentation__.Notify(550928)

	desc, err := getDescForDataSource(ds)
	if err != nil {
		__antithesis_instrumentation__.Notify(550934)
		return cat.DataSourceName{}, err
	} else {
		__antithesis_instrumentation__.Notify(550935)
	}
	__antithesis_instrumentation__.Notify(550929)

	dbID := desc.GetParentID()
	dbDesc, err := oc.planner.Descriptors().Direct().MustGetDatabaseDescByID(ctx, txn, dbID)
	if err != nil {
		__antithesis_instrumentation__.Notify(550936)
		return cat.DataSourceName{}, err
	} else {
		__antithesis_instrumentation__.Notify(550937)
	}
	__antithesis_instrumentation__.Notify(550930)
	scID := desc.GetParentSchemaID()
	var scName tree.Name

	if scID == keys.PublicSchemaID {
		__antithesis_instrumentation__.Notify(550938)
		scName = tree.PublicSchemaName
	} else {
		__antithesis_instrumentation__.Notify(550939)
		scDesc, err := oc.planner.Descriptors().Direct().MustGetSchemaDescByID(ctx, txn, scID)
		if err != nil {
			__antithesis_instrumentation__.Notify(550941)
			return cat.DataSourceName{}, err
		} else {
			__antithesis_instrumentation__.Notify(550942)
		}
		__antithesis_instrumentation__.Notify(550940)
		scName = tree.Name(scDesc.GetName())
	}
	__antithesis_instrumentation__.Notify(550931)

	return tree.MakeTableNameWithSchema(
			tree.Name(dbDesc.GetName()),
			scName,
			tree.Name(desc.GetName())),
		nil
}

func (oc *optCatalog) RoleExists(ctx context.Context, role security.SQLUsername) (bool, error) {
	__antithesis_instrumentation__.Notify(550943)
	return RoleExists(ctx, oc.planner.ExecCfg(), oc.planner.Txn(), role)
}

func (oc *optCatalog) dataSourceForDesc(
	ctx context.Context, flags cat.Flags, desc catalog.TableDescriptor, name *cat.DataSourceName,
) (cat.DataSource, error) {
	__antithesis_instrumentation__.Notify(550944)

	if desc.IsTable() || func() bool {
		__antithesis_instrumentation__.Notify(550948)
		return desc.MaterializedView() == true
	}() == true {
		__antithesis_instrumentation__.Notify(550949)

		return oc.dataSourceForTable(ctx, flags, desc, name)
	} else {
		__antithesis_instrumentation__.Notify(550950)
	}
	__antithesis_instrumentation__.Notify(550945)

	ds, ok := oc.dataSources[desc]
	if ok {
		__antithesis_instrumentation__.Notify(550951)
		return ds, nil
	} else {
		__antithesis_instrumentation__.Notify(550952)
	}
	__antithesis_instrumentation__.Notify(550946)

	switch {
	case desc.IsView():
		__antithesis_instrumentation__.Notify(550953)
		ds = newOptView(desc)

	case desc.IsSequence():
		__antithesis_instrumentation__.Notify(550954)
		ds = newOptSequence(desc)

	default:
		__antithesis_instrumentation__.Notify(550955)
		return nil, errors.AssertionFailedf("unexpected table descriptor: %+v", desc)
	}
	__antithesis_instrumentation__.Notify(550947)

	oc.dataSources[desc] = ds
	return ds, nil
}

func (oc *optCatalog) dataSourceForTable(
	ctx context.Context, flags cat.Flags, desc catalog.TableDescriptor, name *cat.DataSourceName,
) (cat.DataSource, error) {
	__antithesis_instrumentation__.Notify(550956)
	if desc.IsVirtualTable() {
		__antithesis_instrumentation__.Notify(550962)

		return newOptVirtualTable(ctx, oc, desc, name)
	} else {
		__antithesis_instrumentation__.Notify(550963)
	}
	__antithesis_instrumentation__.Notify(550957)

	var tableStats []*stats.TableStatistic
	if !flags.NoTableStats {
		__antithesis_instrumentation__.Notify(550964)
		var err error
		tableStats, err = oc.planner.execCfg.TableStatsCache.GetTableStats(context.TODO(), desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(550965)

			tableStats = nil
		} else {
			__antithesis_instrumentation__.Notify(550966)
		}
	} else {
		__antithesis_instrumentation__.Notify(550967)
	}
	__antithesis_instrumentation__.Notify(550958)

	zoneConfig, err := oc.getZoneConfig(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(550968)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(550969)
	}
	__antithesis_instrumentation__.Notify(550959)

	if ds, ok := oc.dataSources[desc]; ok && func() bool {
		__antithesis_instrumentation__.Notify(550970)
		return !ds.(*optTable).isStale(desc, tableStats, zoneConfig) == true
	}() == true {
		__antithesis_instrumentation__.Notify(550971)
		return ds, nil
	} else {
		__antithesis_instrumentation__.Notify(550972)
	}
	__antithesis_instrumentation__.Notify(550960)

	ds, err := newOptTable(desc, oc.codec(), tableStats, zoneConfig)
	if err != nil {
		__antithesis_instrumentation__.Notify(550973)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(550974)
	}
	__antithesis_instrumentation__.Notify(550961)
	oc.dataSources[desc] = ds
	return ds, nil
}

var emptyZoneConfig = cat.EmptyZone()

func (oc *optCatalog) getZoneConfig(desc catalog.TableDescriptor) (cat.Zone, error) {
	__antithesis_instrumentation__.Notify(550975)

	if oc.cfg == nil || func() bool {
		__antithesis_instrumentation__.Notify(550979)
		return desc.IsVirtualTable() == true
	}() == true {
		__antithesis_instrumentation__.Notify(550980)
		return emptyZoneConfig, nil
	} else {
		__antithesis_instrumentation__.Notify(550981)
	}
	__antithesis_instrumentation__.Notify(550976)
	zone, err := oc.cfg.GetZoneConfigForObject(
		oc.codec(), oc.version(), config.ObjectID(desc.GetID()),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(550982)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(550983)
	}
	__antithesis_instrumentation__.Notify(550977)
	if zone == nil {
		__antithesis_instrumentation__.Notify(550984)

		return emptyZoneConfig, nil
	} else {
		__antithesis_instrumentation__.Notify(550985)
	}
	__antithesis_instrumentation__.Notify(550978)
	return cat.AsZone(zone), nil
}

func (oc *optCatalog) codec() keys.SQLCodec {
	__antithesis_instrumentation__.Notify(550986)
	return oc.planner.ExecCfg().Codec
}

func (oc *optCatalog) version() clusterversion.ClusterVersion {
	__antithesis_instrumentation__.Notify(550987)
	return oc.planner.ExecCfg().Settings.Version.ActiveVersionOrEmpty(
		oc.planner.EvalContext().Context,
	)
}

type optView struct {
	desc catalog.TableDescriptor
}

var _ cat.View = &optView{}

func newOptView(desc catalog.TableDescriptor) *optView {
	__antithesis_instrumentation__.Notify(550988)
	return &optView{desc: desc}
}

func (ov *optView) ID() cat.StableID {
	__antithesis_instrumentation__.Notify(550989)
	return cat.StableID(ov.desc.GetID())
}

func (ov *optView) PostgresDescriptorID() cat.StableID {
	__antithesis_instrumentation__.Notify(550990)
	return cat.StableID(ov.desc.GetID())
}

func (ov *optView) Equals(other cat.Object) bool {
	__antithesis_instrumentation__.Notify(550991)
	otherView, ok := other.(*optView)
	if !ok {
		__antithesis_instrumentation__.Notify(550993)
		return false
	} else {
		__antithesis_instrumentation__.Notify(550994)
	}
	__antithesis_instrumentation__.Notify(550992)
	return ov.desc.GetID() == otherView.desc.GetID() && func() bool {
		__antithesis_instrumentation__.Notify(550995)
		return ov.desc.GetVersion() == otherView.desc.GetVersion() == true
	}() == true
}

func (ov *optView) Name() tree.Name {
	__antithesis_instrumentation__.Notify(550996)
	return tree.Name(ov.desc.GetName())
}

func (ov *optView) IsSystemView() bool {
	__antithesis_instrumentation__.Notify(550997)
	return ov.desc.IsVirtualTable()
}

func (ov *optView) Query() string {
	__antithesis_instrumentation__.Notify(550998)
	return ov.desc.GetViewQuery()
}

func (ov *optView) ColumnNameCount() int {
	__antithesis_instrumentation__.Notify(550999)
	return len(ov.desc.PublicColumns())
}

func (ov *optView) ColumnName(i int) tree.Name {
	__antithesis_instrumentation__.Notify(551000)
	return ov.desc.PublicColumns()[i].ColName()
}

func (ov *optView) CollectTypes(ord int) (descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(551001)
	col := ov.desc.AllColumns()[ord]
	return collectTypes(col)
}

type optSequence struct {
	desc catalog.TableDescriptor
}

var _ cat.DataSource = &optSequence{}
var _ cat.Sequence = &optSequence{}

func newOptSequence(desc catalog.TableDescriptor) *optSequence {
	__antithesis_instrumentation__.Notify(551002)
	return &optSequence{desc: desc}
}

func (os *optSequence) ID() cat.StableID {
	__antithesis_instrumentation__.Notify(551003)
	return cat.StableID(os.desc.GetID())
}

func (os *optSequence) PostgresDescriptorID() cat.StableID {
	__antithesis_instrumentation__.Notify(551004)
	return cat.StableID(os.desc.GetID())
}

func (os *optSequence) Equals(other cat.Object) bool {
	__antithesis_instrumentation__.Notify(551005)
	otherSeq, ok := other.(*optSequence)
	if !ok {
		__antithesis_instrumentation__.Notify(551007)
		return false
	} else {
		__antithesis_instrumentation__.Notify(551008)
	}
	__antithesis_instrumentation__.Notify(551006)
	return os.desc.GetID() == otherSeq.desc.GetID() && func() bool {
		__antithesis_instrumentation__.Notify(551009)
		return os.desc.GetVersion() == otherSeq.desc.GetVersion() == true
	}() == true
}

func (os *optSequence) Name() tree.Name {
	__antithesis_instrumentation__.Notify(551010)
	return tree.Name(os.desc.GetName())
}

func (os *optSequence) SequenceMarker() { __antithesis_instrumentation__.Notify(551011) }

func (os *optSequence) CollectTypes(ord int) (descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(551012)
	col := os.desc.AllColumns()[ord]
	return collectTypes(col)
}

type optTable struct {
	desc catalog.TableDescriptor

	columns []cat.Column

	indexes []optIndex

	codec keys.SQLCodec

	rawStats []*stats.TableStatistic

	stats []optTableStat

	zone cat.Zone

	primaryFamily optFamily

	families []optFamily

	uniqueConstraints []optUniqueConstraint

	outboundFKs []optForeignKeyConstraint
	inboundFKs  []optForeignKeyConstraint

	checkConstraints []cat.CheckConstraint

	colMap catalog.TableColMap
}

var _ cat.Table = &optTable{}

func newOptTable(
	desc catalog.TableDescriptor,
	codec keys.SQLCodec,
	stats []*stats.TableStatistic,
	tblZone cat.Zone,
) (*optTable, error) {
	__antithesis_instrumentation__.Notify(551013)
	ot := &optTable{
		desc:     desc,
		codec:    codec,
		rawStats: stats,
		zone:     tblZone,
	}

	pkCols := desc.GetPrimaryIndex().CollectKeyColumnIDs()

	cols := ot.desc.DeletableColumns()
	numCols := len(ot.desc.AllColumns())

	secondaryIndexes := ot.desc.DeletableNonPrimaryIndexes()
	for _, index := range secondaryIndexes {
		__antithesis_instrumentation__.Notify(551027)
		if index.GetType() == descpb.IndexDescriptor_INVERTED {
			__antithesis_instrumentation__.Notify(551028)
			numCols++
		} else {
			__antithesis_instrumentation__.Notify(551029)
		}
	}
	__antithesis_instrumentation__.Notify(551014)

	ot.columns = make([]cat.Column, len(cols), numCols)
	for _, col := range cols {
		__antithesis_instrumentation__.Notify(551030)
		var kind cat.ColumnKind
		visibility := cat.Visible
		switch {
		case col.Public():
			__antithesis_instrumentation__.Notify(551032)
			kind = cat.Ordinary
			if col.IsInaccessible() {
				__antithesis_instrumentation__.Notify(551035)
				visibility = cat.Inaccessible
			} else {
				__antithesis_instrumentation__.Notify(551036)
				if col.IsHidden() {
					__antithesis_instrumentation__.Notify(551037)
					visibility = cat.Hidden
				} else {
					__antithesis_instrumentation__.Notify(551038)
				}
			}
		case col.WriteAndDeleteOnly():
			__antithesis_instrumentation__.Notify(551033)
			kind = cat.WriteOnly
			visibility = cat.Inaccessible
		default:
			__antithesis_instrumentation__.Notify(551034)
			kind = cat.DeleteOnly
			visibility = cat.Inaccessible
		}
		__antithesis_instrumentation__.Notify(551031)

		if !col.IsVirtual() || func() bool {
			__antithesis_instrumentation__.Notify(551039)
			return pkCols.Contains(col.GetID()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(551040)
			ot.columns[col.Ordinal()].Init(
				col.Ordinal(),
				cat.StableID(col.GetID()),
				col.ColName(),
				kind,
				col.GetType(),
				col.IsNullable(),
				visibility,
				col.ColumnDesc().DefaultExpr,
				col.ColumnDesc().ComputeExpr,
				col.ColumnDesc().OnUpdateExpr,
				mapGeneratedAsIdentityType(col.GetGeneratedAsIdentityType()),
				col.ColumnDesc().GeneratedAsIdentitySequenceOption,
			)
		} else {
			__antithesis_instrumentation__.Notify(551041)

			ot.columns[col.Ordinal()].InitVirtualComputed(
				col.Ordinal(),
				cat.StableID(col.GetID()),
				col.ColName(),
				col.GetType(),
				col.IsNullable(),
				visibility,
				col.GetComputeExpr(),
			)
		}
	}
	__antithesis_instrumentation__.Notify(551015)

	newColumn := func() (col *cat.Column, ordinal int) {
		__antithesis_instrumentation__.Notify(551042)
		ordinal = len(ot.columns)
		ot.columns = ot.columns[:ordinal+1]
		return &ot.columns[ordinal], ordinal
	}
	__antithesis_instrumentation__.Notify(551016)

	for _, sysCol := range ot.desc.SystemColumns() {
		__antithesis_instrumentation__.Notify(551043)
		found, _ := desc.FindColumnWithName(sysCol.ColName())
		if found == nil || func() bool {
			__antithesis_instrumentation__.Notify(551044)
			return found.IsSystemColumn() == true
		}() == true {
			__antithesis_instrumentation__.Notify(551045)
			col, ord := newColumn()
			col.Init(
				ord,
				cat.StableID(sysCol.GetID()),
				sysCol.ColName(),
				cat.System,
				sysCol.GetType(),
				sysCol.IsNullable(),
				cat.MaybeHidden(sysCol.IsHidden()),
				sysCol.ColumnDesc().DefaultExpr,
				sysCol.ColumnDesc().ComputeExpr,
				sysCol.ColumnDesc().OnUpdateExpr,
				mapGeneratedAsIdentityType(sysCol.GetGeneratedAsIdentityType()),
				sysCol.ColumnDesc().GeneratedAsIdentitySequenceOption,
			)
		} else {
			__antithesis_instrumentation__.Notify(551046)
		}
	}
	__antithesis_instrumentation__.Notify(551017)

	for i := range ot.columns {
		__antithesis_instrumentation__.Notify(551047)
		ot.colMap.Set(descpb.ColumnID(ot.columns[i].ColID()), i)
	}
	__antithesis_instrumentation__.Notify(551018)

	ot.uniqueConstraints = make([]optUniqueConstraint, 0, len(ot.desc.GetUniqueWithoutIndexConstraints()))
	for i := range ot.desc.GetUniqueWithoutIndexConstraints() {
		__antithesis_instrumentation__.Notify(551048)
		u := &ot.desc.GetUniqueWithoutIndexConstraints()[i]
		ot.uniqueConstraints = append(ot.uniqueConstraints, optUniqueConstraint{
			name:         u.Name,
			table:        ot.ID(),
			columns:      u.ColumnIDs,
			predicate:    u.Predicate,
			withoutIndex: true,
			validity:     u.Validity,
		})
	}
	__antithesis_instrumentation__.Notify(551019)

	ot.indexes = make([]optIndex, 1+len(secondaryIndexes))

	for i := range ot.indexes {
		__antithesis_instrumentation__.Notify(551049)
		var idx catalog.Index
		if i == 0 {
			__antithesis_instrumentation__.Notify(551053)
			idx = desc.GetPrimaryIndex()
		} else {
			__antithesis_instrumentation__.Notify(551054)
			idx = secondaryIndexes[i-1]
		}
		__antithesis_instrumentation__.Notify(551050)

		idxZone := tblZone
		partZones := make(map[string]cat.Zone)
		for j := 0; j < tblZone.SubzoneCount(); j++ {
			__antithesis_instrumentation__.Notify(551055)
			subzone := tblZone.Subzone(j)
			if subzone.Index() == cat.StableID(idx.GetID()) {
				__antithesis_instrumentation__.Notify(551056)
				copyZone := subzone.Zone().InheritFromParent(tblZone)
				if subzone.Partition() == "" {
					__antithesis_instrumentation__.Notify(551057)

					idxZone = copyZone
				} else {
					__antithesis_instrumentation__.Notify(551058)

					partZones[subzone.Partition()] = copyZone
				}
			} else {
				__antithesis_instrumentation__.Notify(551059)
			}
		}
		__antithesis_instrumentation__.Notify(551051)
		if idx.GetType() == descpb.IndexDescriptor_INVERTED {
			__antithesis_instrumentation__.Notify(551060)

			invertedColumnID := idx.InvertedColumnID()
			invertedColumnName := idx.InvertedColumnName()
			invertedColumnType := idx.InvertedColumnKeyType()

			invertedSourceColOrdinal, _ := ot.lookupColumnOrdinal(invertedColumnID)

			invertedCol, invertedColOrd := newColumn()

			invertedCol.InitInverted(
				invertedColOrd,
				tree.Name(invertedColumnName+"_inverted_key"),
				invertedColumnType,
				false,
				invertedSourceColOrdinal,
			)
			ot.indexes[i].init(ot, i, idx, idxZone, partZones, invertedColOrd)
		} else {
			__antithesis_instrumentation__.Notify(551061)
			ot.indexes[i].init(ot, i, idx, idxZone, partZones, -1)
		}
		__antithesis_instrumentation__.Notify(551052)

		if idx.IsUnique() {
			__antithesis_instrumentation__.Notify(551062)
			if idx.GetPartitioning().NumImplicitColumns() > 0 {
				__antithesis_instrumentation__.Notify(551063)

				ot.uniqueConstraints = append(ot.uniqueConstraints, optUniqueConstraint{
					name:         idx.GetName(),
					table:        ot.ID(),
					columns:      idx.IndexDesc().KeyColumnIDs[idx.IndexDesc().ExplicitColumnStartIdx():],
					withoutIndex: true,
					predicate:    idx.GetPredicate(),

					validity: descpb.ConstraintValidity_Validated,
				})
			} else {
				__antithesis_instrumentation__.Notify(551064)
				if idx.IsSharded() {
					__antithesis_instrumentation__.Notify(551065)

					ot.uniqueConstraints = append(ot.uniqueConstraints, optUniqueConstraint{
						name:                               idx.GetName(),
						table:                              ot.ID(),
						columns:                            idx.IndexDesc().KeyColumnIDs[idx.IndexDesc().ExplicitColumnStartIdx():],
						withoutIndex:                       true,
						predicate:                          idx.GetPredicate(),
						validity:                           descpb.ConstraintValidity_Validated,
						uniquenessGuaranteedByAnotherIndex: true,
					})
				} else {
					__antithesis_instrumentation__.Notify(551066)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(551067)
		}
	}
	__antithesis_instrumentation__.Notify(551020)

	_ = ot.desc.ForeachOutboundFK(func(fk *descpb.ForeignKeyConstraint) error {
		__antithesis_instrumentation__.Notify(551068)
		ot.outboundFKs = append(ot.outboundFKs, optForeignKeyConstraint{
			name:              fk.Name,
			originTable:       ot.ID(),
			originColumns:     fk.OriginColumnIDs,
			referencedTable:   cat.StableID(fk.ReferencedTableID),
			referencedColumns: fk.ReferencedColumnIDs,
			validity:          fk.Validity,
			match:             fk.Match,
			deleteAction:      fk.OnDelete,
			updateAction:      fk.OnUpdate,
		})
		return nil
	})
	__antithesis_instrumentation__.Notify(551021)
	_ = ot.desc.ForeachInboundFK(func(fk *descpb.ForeignKeyConstraint) error {
		__antithesis_instrumentation__.Notify(551069)
		ot.inboundFKs = append(ot.inboundFKs, optForeignKeyConstraint{
			name:              fk.Name,
			originTable:       cat.StableID(fk.OriginTableID),
			originColumns:     fk.OriginColumnIDs,
			referencedTable:   ot.ID(),
			referencedColumns: fk.ReferencedColumnIDs,
			validity:          fk.Validity,
			match:             fk.Match,
			deleteAction:      fk.OnDelete,
			updateAction:      fk.OnUpdate,
		})
		return nil
	})
	__antithesis_instrumentation__.Notify(551022)

	ot.primaryFamily.init(ot, &desc.GetFamilies()[0])
	ot.families = make([]optFamily, len(desc.GetFamilies())-1)
	for i := range ot.families {
		__antithesis_instrumentation__.Notify(551070)
		ot.families[i].init(ot, &desc.GetFamilies()[i+1])
	}
	__antithesis_instrumentation__.Notify(551023)

	var synthesizedChecks []cat.CheckConstraint
	for i := 0; i < ot.ColumnCount(); i++ {
		__antithesis_instrumentation__.Notify(551071)
		col := ot.Column(i)
		if col.IsMutation() {
			__antithesis_instrumentation__.Notify(551073)

			continue
		} else {
			__antithesis_instrumentation__.Notify(551074)
		}
		__antithesis_instrumentation__.Notify(551072)
		colType := col.DatumType()
		if colType.UserDefined() {
			__antithesis_instrumentation__.Notify(551075)
			switch colType.Family() {
			case types.EnumFamily:
				__antithesis_instrumentation__.Notify(551076)

				expr := &tree.ComparisonExpr{
					Operator: treecmp.MakeComparisonOperator(treecmp.In),
					Left:     &tree.ColumnItem{ColumnName: col.ColName()},
					Right:    tree.NewDTuple(colType, tree.MakeAllDEnumsInType(colType)...),
				}
				synthesizedChecks = append(synthesizedChecks, cat.CheckConstraint{
					Constraint: tree.Serialize(expr),
					Validated:  true,
				})
			default:
				__antithesis_instrumentation__.Notify(551077)
			}
		} else {
			__antithesis_instrumentation__.Notify(551078)
		}
	}
	__antithesis_instrumentation__.Notify(551024)

	activeChecks := desc.ActiveChecks()
	ot.checkConstraints = make([]cat.CheckConstraint, 0, len(activeChecks)+len(synthesizedChecks))
	for i := range activeChecks {
		__antithesis_instrumentation__.Notify(551079)
		ot.checkConstraints = append(ot.checkConstraints, cat.CheckConstraint{
			Constraint: activeChecks[i].Expr,
			Validated:  activeChecks[i].Validity == descpb.ConstraintValidity_Validated,
		})
	}
	__antithesis_instrumentation__.Notify(551025)
	ot.checkConstraints = append(ot.checkConstraints, synthesizedChecks...)

	if stats != nil {
		__antithesis_instrumentation__.Notify(551080)
		ot.stats = make([]optTableStat, len(stats))
		n := 0
		for i := range stats {
			__antithesis_instrumentation__.Notify(551082)

			if ok, err := ot.stats[n].init(ot, stats[i]); err != nil {
				__antithesis_instrumentation__.Notify(551083)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(551084)
				if ok {
					__antithesis_instrumentation__.Notify(551085)
					n++
				} else {
					__antithesis_instrumentation__.Notify(551086)
				}
			}
		}
		__antithesis_instrumentation__.Notify(551081)
		ot.stats = ot.stats[:n]
	} else {
		__antithesis_instrumentation__.Notify(551087)
	}
	__antithesis_instrumentation__.Notify(551026)

	return ot, nil
}

func (ot *optTable) ID() cat.StableID {
	__antithesis_instrumentation__.Notify(551088)
	return cat.StableID(ot.desc.GetID())
}

func (ot *optTable) PostgresDescriptorID() cat.StableID {
	__antithesis_instrumentation__.Notify(551089)
	return cat.StableID(ot.desc.GetID())
}

func (ot *optTable) isStale(
	rawDesc catalog.TableDescriptor, tableStats []*stats.TableStatistic, zone cat.Zone,
) bool {
	__antithesis_instrumentation__.Notify(551090)

	if len(tableStats) != len(ot.rawStats) {
		__antithesis_instrumentation__.Notify(551095)
		return true
	} else {
		__antithesis_instrumentation__.Notify(551096)
	}
	__antithesis_instrumentation__.Notify(551091)
	if len(tableStats) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(551097)
		return &tableStats[0] != &ot.rawStats[0] == true
	}() == true {
		__antithesis_instrumentation__.Notify(551098)
		return true
	} else {
		__antithesis_instrumentation__.Notify(551099)
	}
	__antithesis_instrumentation__.Notify(551092)
	if !zone.Equal(ot.zone) {
		__antithesis_instrumentation__.Notify(551100)
		return true
	} else {
		__antithesis_instrumentation__.Notify(551101)
	}
	__antithesis_instrumentation__.Notify(551093)

	if !catalog.UserDefinedTypeColsHaveSameVersion(ot.desc, rawDesc) {
		__antithesis_instrumentation__.Notify(551102)
		return true
	} else {
		__antithesis_instrumentation__.Notify(551103)
	}
	__antithesis_instrumentation__.Notify(551094)
	return false
}

func (ot *optTable) Equals(other cat.Object) bool {
	__antithesis_instrumentation__.Notify(551104)
	otherTable, ok := other.(*optTable)
	if !ok {
		__antithesis_instrumentation__.Notify(551112)
		return false
	} else {
		__antithesis_instrumentation__.Notify(551113)
	}
	__antithesis_instrumentation__.Notify(551105)
	if ot == otherTable {
		__antithesis_instrumentation__.Notify(551114)

		return true
	} else {
		__antithesis_instrumentation__.Notify(551115)
	}
	__antithesis_instrumentation__.Notify(551106)
	if ot.desc.GetID() != otherTable.desc.GetID() || func() bool {
		__antithesis_instrumentation__.Notify(551116)
		return ot.desc.GetVersion() != otherTable.desc.GetVersion() == true
	}() == true {
		__antithesis_instrumentation__.Notify(551117)
		return false
	} else {
		__antithesis_instrumentation__.Notify(551118)
	}
	__antithesis_instrumentation__.Notify(551107)

	if len(ot.stats) != len(otherTable.stats) {
		__antithesis_instrumentation__.Notify(551119)
		return false
	} else {
		__antithesis_instrumentation__.Notify(551120)
	}
	__antithesis_instrumentation__.Notify(551108)
	for i := range ot.stats {
		__antithesis_instrumentation__.Notify(551121)
		if !ot.stats[i].equals(&otherTable.stats[i]) {
			__antithesis_instrumentation__.Notify(551122)
			return false
		} else {
			__antithesis_instrumentation__.Notify(551123)
		}
	}
	__antithesis_instrumentation__.Notify(551109)

	if !catalog.UserDefinedTypeColsHaveSameVersion(ot.desc, otherTable.desc) {
		__antithesis_instrumentation__.Notify(551124)
		return false
	} else {
		__antithesis_instrumentation__.Notify(551125)
	}
	__antithesis_instrumentation__.Notify(551110)

	var prevLeftZone, prevRightZone cat.Zone
	for i := range ot.indexes {
		__antithesis_instrumentation__.Notify(551126)
		leftZone := ot.indexes[i].zone
		rightZone := otherTable.indexes[i].zone
		if leftZone == prevLeftZone && func() bool {
			__antithesis_instrumentation__.Notify(551129)
			return rightZone == prevRightZone == true
		}() == true {
			__antithesis_instrumentation__.Notify(551130)
			continue
		} else {
			__antithesis_instrumentation__.Notify(551131)
		}
		__antithesis_instrumentation__.Notify(551127)
		if !leftZone.Equal(rightZone) {
			__antithesis_instrumentation__.Notify(551132)
			return false
		} else {
			__antithesis_instrumentation__.Notify(551133)
		}
		__antithesis_instrumentation__.Notify(551128)
		prevLeftZone = leftZone
		prevRightZone = rightZone
	}
	__antithesis_instrumentation__.Notify(551111)

	return true
}

func (ot *optTable) Name() tree.Name {
	__antithesis_instrumentation__.Notify(551134)
	return tree.Name(ot.desc.GetName())
}

func (ot *optTable) IsVirtualTable() bool {
	__antithesis_instrumentation__.Notify(551135)
	return false
}

func (ot *optTable) IsMaterializedView() bool {
	__antithesis_instrumentation__.Notify(551136)
	return ot.desc.MaterializedView()
}

func (ot *optTable) ColumnCount() int {
	__antithesis_instrumentation__.Notify(551137)
	return len(ot.columns)
}

func (ot *optTable) Column(i int) *cat.Column {
	__antithesis_instrumentation__.Notify(551138)
	return &ot.columns[i]
}

func (ot *optTable) getCol(i int) catalog.Column {
	__antithesis_instrumentation__.Notify(551139)
	if i < len(ot.desc.AllColumns()) {
		__antithesis_instrumentation__.Notify(551141)
		return ot.desc.AllColumns()[i]
	} else {
		__antithesis_instrumentation__.Notify(551142)
	}
	__antithesis_instrumentation__.Notify(551140)
	return nil
}

func (ot *optTable) IndexCount() int {
	__antithesis_instrumentation__.Notify(551143)

	return len(ot.desc.ActiveIndexes())
}

func (ot *optTable) WritableIndexCount() int {
	__antithesis_instrumentation__.Notify(551144)

	return 1 + len(ot.desc.WritableNonPrimaryIndexes())
}

func (ot *optTable) DeletableIndexCount() int {
	__antithesis_instrumentation__.Notify(551145)

	return len(ot.desc.DeletableNonPrimaryIndexes()) + 1
}

func (ot *optTable) Index(i cat.IndexOrdinal) cat.Index {
	__antithesis_instrumentation__.Notify(551146)
	return &ot.indexes[i]
}

func (ot *optTable) StatisticCount() int {
	__antithesis_instrumentation__.Notify(551147)
	return len(ot.stats)
}

func (ot *optTable) Statistic(i int) cat.TableStatistic {
	__antithesis_instrumentation__.Notify(551148)
	return &ot.stats[i]
}

func (ot *optTable) CheckCount() int {
	__antithesis_instrumentation__.Notify(551149)
	return len(ot.checkConstraints)
}

func (ot *optTable) Check(i int) cat.CheckConstraint {
	__antithesis_instrumentation__.Notify(551150)
	return ot.checkConstraints[i]
}

func (ot *optTable) FamilyCount() int {
	__antithesis_instrumentation__.Notify(551151)
	return 1 + len(ot.families)
}

func (ot *optTable) Family(i int) cat.Family {
	__antithesis_instrumentation__.Notify(551152)
	if i == 0 {
		__antithesis_instrumentation__.Notify(551154)
		return &ot.primaryFamily
	} else {
		__antithesis_instrumentation__.Notify(551155)
	}
	__antithesis_instrumentation__.Notify(551153)
	return &ot.families[i-1]
}

func (ot *optTable) OutboundForeignKeyCount() int {
	__antithesis_instrumentation__.Notify(551156)
	return len(ot.outboundFKs)
}

func (ot *optTable) OutboundForeignKey(i int) cat.ForeignKeyConstraint {
	__antithesis_instrumentation__.Notify(551157)
	return &ot.outboundFKs[i]
}

func (ot *optTable) InboundForeignKeyCount() int {
	__antithesis_instrumentation__.Notify(551158)
	return len(ot.inboundFKs)
}

func (ot *optTable) InboundForeignKey(i int) cat.ForeignKeyConstraint {
	__antithesis_instrumentation__.Notify(551159)
	return &ot.inboundFKs[i]
}

func (ot *optTable) UniqueCount() int {
	__antithesis_instrumentation__.Notify(551160)
	return len(ot.uniqueConstraints)
}

func (ot *optTable) Unique(i cat.UniqueOrdinal) cat.UniqueConstraint {
	__antithesis_instrumentation__.Notify(551161)
	return &ot.uniqueConstraints[i]
}

func (ot *optTable) Zone() cat.Zone {
	__antithesis_instrumentation__.Notify(551162)
	return ot.zone
}

func (ot *optTable) IsPartitionAllBy() bool {
	__antithesis_instrumentation__.Notify(551163)
	return ot.desc.IsPartitionAllBy()
}

func (ot *optTable) lookupColumnOrdinal(colID descpb.ColumnID) (int, error) {
	__antithesis_instrumentation__.Notify(551164)
	col, ok := ot.colMap.Get(colID)
	if ok {
		__antithesis_instrumentation__.Notify(551166)
		return col, nil
	} else {
		__antithesis_instrumentation__.Notify(551167)
	}
	__antithesis_instrumentation__.Notify(551165)
	return col, pgerror.Newf(pgcode.UndefinedColumn,
		"column [%d] does not exist", colID)
}

func convertTableToOptTable(tab cat.Table) *optTable {
	__antithesis_instrumentation__.Notify(551168)
	var optTab *optTable
	switch t := tab.(type) {
	case *optTable:
		__antithesis_instrumentation__.Notify(551170)
		optTab = t
	case *indexrec.HypotheticalTable:
		__antithesis_instrumentation__.Notify(551171)
		optTab = t.Table.(*optTable)
	}
	__antithesis_instrumentation__.Notify(551169)
	return optTab
}

func (ot *optTable) CollectTypes(ord int) (descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(551172)
	col := ot.desc.AllColumns()[ord]
	return collectTypes(col)
}

type optIndex struct {
	tab  *optTable
	idx  catalog.Index
	zone cat.Zone

	columnOrds []int

	storedCols []descpb.ColumnID

	indexOrdinal  int
	numCols       int
	numKeyCols    int
	numLaxKeyCols int

	partitions []optPartition

	invertedColOrd int
}

var _ cat.Index = &optIndex{}

func (oi *optIndex) init(
	tab *optTable,
	indexOrdinal int,
	idx catalog.Index,
	zone cat.Zone,
	partZones map[string]cat.Zone,
	invertedColOrd int,
) {
	oi.tab = tab
	oi.idx = idx
	oi.zone = zone
	oi.indexOrdinal = indexOrdinal
	oi.invertedColOrd = invertedColOrd
	if idx.Primary() {

		oi.storedCols = make([]descpb.ColumnID, 0, tab.ColumnCount()-idx.NumKeyColumns())
		pkCols := idx.CollectKeyColumnIDs()
		for i, n := 0, tab.ColumnCount(); i < n; i++ {
			if col := tab.Column(i); col.Kind() != cat.Inverted && !col.IsVirtualComputed() {
				if id := descpb.ColumnID(col.ColID()); !pkCols.Contains(id) {
					oi.storedCols = append(oi.storedCols, id)
				}
			}
		}
		oi.numCols = idx.NumKeyColumns() + len(oi.storedCols)
	} else {
		oi.storedCols = idx.IndexDesc().StoreColumnIDs
		oi.numCols = idx.NumKeyColumns() + idx.NumKeySuffixColumns() + idx.NumSecondaryStoredColumns()
	}

	idxPartitioning := idx.GetPartitioning()
	oi.partitions = make([]optPartition, 0, idxPartitioning.NumLists())
	_ = idxPartitioning.ForEachList(func(name string, values [][]byte, subPartitioning catalog.Partitioning) error {
		op := optPartition{
			name:   name,
			zone:   cat.EmptyZone(),
			datums: make([]tree.Datums, 0, len(values)),
		}

		if zone, ok := partZones[name]; ok {
			op.zone = zone
		}

		var a tree.DatumAlloc
		for _, valueEncBuf := range values {
			t, _, err := rowenc.DecodePartitionTuple(
				&a, oi.tab.codec, oi.tab.desc, oi.idx, oi.idx.GetPartitioning(),
				valueEncBuf, nil,
			)
			if err != nil {
				log.Fatalf(context.TODO(), "error while decoding partition tuple: %+v %+v",
					oi.tab.desc, oi.tab.desc.GetDependsOnTypes())
			}
			op.datums = append(op.datums, t.Datums)

		}

		oi.partitions = append(oi.partitions, op)
		return nil
	})

	if idx.IsUnique() {
		notNull := true
		for i := 0; i < idx.NumKeyColumns(); i++ {
			id := idx.GetKeyColumnID(i)
			ord, _ := tab.lookupColumnOrdinal(id)
			if tab.Column(ord).IsNullable() {
				notNull = false
				break
			}
		}

		if notNull {

			oi.numLaxKeyCols = idx.NumKeyColumns()
			oi.numKeyCols = oi.numLaxKeyCols
		} else {

			oi.numLaxKeyCols = idx.NumKeyColumns()
			oi.numKeyCols = oi.numLaxKeyCols + idx.NumKeySuffixColumns()
		}
	} else {

		oi.numLaxKeyCols = idx.NumKeyColumns() + idx.NumKeySuffixColumns()
		oi.numKeyCols = oi.numLaxKeyCols
	}

	inverted := oi.IsInverted()
	numKeyCols := idx.NumKeyColumns()
	numKeySuffixCols := idx.NumKeySuffixColumns()
	oi.columnOrds = make([]int, oi.numCols)
	for i := 0; i < oi.numCols; i++ {
		var ord int
		switch {
		case inverted && i == numKeyCols-1:
			ord = oi.invertedColOrd
		case i < numKeyCols:
			ord, _ = oi.tab.lookupColumnOrdinal(oi.idx.GetKeyColumnID(i))
		case i < numKeyCols+numKeySuffixCols:
			ord, _ = oi.tab.lookupColumnOrdinal(oi.idx.GetKeySuffixColumnID(i - numKeyCols))
		default:
			ord, _ = oi.tab.lookupColumnOrdinal(oi.storedCols[i-numKeyCols-numKeySuffixCols])
		}
		oi.columnOrds[i] = ord
	}
}

func (oi *optIndex) ID() cat.StableID {
	__antithesis_instrumentation__.Notify(551173)
	return cat.StableID(oi.idx.GetID())
}

func (oi *optIndex) Name() tree.Name {
	__antithesis_instrumentation__.Notify(551174)
	return tree.Name(oi.idx.GetName())
}

func (oi *optIndex) IsUnique() bool {
	__antithesis_instrumentation__.Notify(551175)
	return oi.idx.IsUnique()
}

func (oi *optIndex) IsInverted() bool {
	__antithesis_instrumentation__.Notify(551176)
	return oi.idx.GetType() == descpb.IndexDescriptor_INVERTED
}

func (oi *optIndex) ColumnCount() int {
	__antithesis_instrumentation__.Notify(551177)
	return oi.numCols
}

func (oi *optIndex) ExplicitColumnCount() int {
	__antithesis_instrumentation__.Notify(551178)
	return oi.idx.NumKeyColumns()
}

func (oi *optIndex) KeyColumnCount() int {
	__antithesis_instrumentation__.Notify(551179)
	return oi.numKeyCols
}

func (oi *optIndex) LaxKeyColumnCount() int {
	__antithesis_instrumentation__.Notify(551180)
	return oi.numLaxKeyCols
}

func (oi *optIndex) NonInvertedPrefixColumnCount() int {
	__antithesis_instrumentation__.Notify(551181)
	if !oi.IsInverted() {
		__antithesis_instrumentation__.Notify(551183)
		panic("non-inverted indexes do not have inverted prefix columns")
	} else {
		__antithesis_instrumentation__.Notify(551184)
	}
	__antithesis_instrumentation__.Notify(551182)
	return oi.idx.NumKeyColumns() - 1
}

func (oi *optIndex) Column(i int) cat.IndexColumn {
	__antithesis_instrumentation__.Notify(551185)
	ord := oi.columnOrds[i]

	descending := i < oi.idx.NumKeyColumns() && func() bool {
		__antithesis_instrumentation__.Notify(551186)
		return oi.idx.GetKeyColumnDirection(i) == descpb.IndexDescriptor_DESC == true
	}() == true
	return cat.IndexColumn{
		Column:     oi.tab.Column(ord),
		Descending: descending,
	}
}

func (oi *optIndex) InvertedColumn() cat.IndexColumn {
	__antithesis_instrumentation__.Notify(551187)
	if !oi.IsInverted() {
		__antithesis_instrumentation__.Notify(551189)
		panic(errors.AssertionFailedf("non-inverted indexes do not have inverted columns"))
	} else {
		__antithesis_instrumentation__.Notify(551190)
	}
	__antithesis_instrumentation__.Notify(551188)
	ord := oi.idx.NumKeyColumns() - 1
	return oi.Column(ord)
}

func (oi *optIndex) Predicate() (string, bool) {
	__antithesis_instrumentation__.Notify(551191)
	return oi.idx.GetPredicate(), oi.idx.GetPredicate() != ""
}

func (oi *optIndex) Zone() cat.Zone {
	__antithesis_instrumentation__.Notify(551192)
	return oi.zone
}

func (oi *optIndex) Span() roachpb.Span {
	__antithesis_instrumentation__.Notify(551193)
	desc := oi.tab.desc

	if desc.GetID() <= keys.MaxSystemConfigDescID {
		__antithesis_instrumentation__.Notify(551195)
		return keys.SystemConfigSpan
	} else {
		__antithesis_instrumentation__.Notify(551196)
	}
	__antithesis_instrumentation__.Notify(551194)
	return desc.IndexSpan(oi.tab.codec, oi.idx.GetID())
}

func (oi *optIndex) Table() cat.Table {
	__antithesis_instrumentation__.Notify(551197)
	return oi.tab
}

func (oi *optIndex) Ordinal() int {
	__antithesis_instrumentation__.Notify(551198)
	return oi.indexOrdinal
}

func (oi *optIndex) ImplicitColumnCount() int {
	__antithesis_instrumentation__.Notify(551199)
	implicitColCnt := oi.idx.GetPartitioning().NumImplicitColumns()
	if oi.idx.IsSharded() {
		__antithesis_instrumentation__.Notify(551201)
		implicitColCnt++
	} else {
		__antithesis_instrumentation__.Notify(551202)
	}
	__antithesis_instrumentation__.Notify(551200)
	return implicitColCnt
}

func (oi *optIndex) ImplicitPartitioningColumnCount() int {
	__antithesis_instrumentation__.Notify(551203)
	return oi.idx.GetPartitioning().NumImplicitColumns()
}

func (oi *optIndex) GeoConfig() *geoindex.Config {
	__antithesis_instrumentation__.Notify(551204)
	return &oi.idx.IndexDesc().GeoConfig
}

func (oi *optIndex) Version() descpb.IndexDescriptorVersion {
	__antithesis_instrumentation__.Notify(551205)
	return oi.idx.GetVersion()
}

func (oi *optIndex) PartitionCount() int {
	__antithesis_instrumentation__.Notify(551206)
	return len(oi.partitions)
}

func (oi *optIndex) Partition(i int) cat.Partition {
	__antithesis_instrumentation__.Notify(551207)
	return &oi.partitions[i]
}

type optPartition struct {
	name   string
	zone   cat.Zone
	datums []tree.Datums
}

var _ cat.Partition = &optPartition{}

func (op *optPartition) Name() string {
	__antithesis_instrumentation__.Notify(551208)
	return op.name
}

func (op *optPartition) Zone() cat.Zone {
	__antithesis_instrumentation__.Notify(551209)
	return op.zone
}

func (op *optPartition) PartitionByListPrefixes() []tree.Datums {
	__antithesis_instrumentation__.Notify(551210)
	return op.datums
}

type optTableStat struct {
	stat           *stats.TableStatistic
	columnOrdinals []int
}

var _ cat.TableStatistic = &optTableStat{}

func (os *optTableStat) init(tab *optTable, stat *stats.TableStatistic) (ok bool, _ error) {
	os.stat = stat
	os.columnOrdinals = make([]int, len(stat.ColumnIDs))
	for i, c := range stat.ColumnIDs {
		var ok bool
		os.columnOrdinals[i], ok = tab.colMap.Get(c)
		if !ok {

			return false, nil
		}
	}

	return true, nil
}

func (os *optTableStat) equals(other *optTableStat) bool {
	__antithesis_instrumentation__.Notify(551211)

	if os.CreatedAt() != other.CreatedAt() || func() bool {
		__antithesis_instrumentation__.Notify(551214)
		return len(os.columnOrdinals) != len(other.columnOrdinals) == true
	}() == true {
		__antithesis_instrumentation__.Notify(551215)
		return false
	} else {
		__antithesis_instrumentation__.Notify(551216)
	}
	__antithesis_instrumentation__.Notify(551212)
	for i, c := range os.columnOrdinals {
		__antithesis_instrumentation__.Notify(551217)
		if c != other.columnOrdinals[i] {
			__antithesis_instrumentation__.Notify(551218)
			return false
		} else {
			__antithesis_instrumentation__.Notify(551219)
		}
	}
	__antithesis_instrumentation__.Notify(551213)
	return true
}

func (os *optTableStat) CreatedAt() time.Time {
	__antithesis_instrumentation__.Notify(551220)
	return os.stat.CreatedAt
}

func (os *optTableStat) ColumnCount() int {
	__antithesis_instrumentation__.Notify(551221)
	return len(os.columnOrdinals)
}

func (os *optTableStat) ColumnOrdinal(i int) int {
	__antithesis_instrumentation__.Notify(551222)
	return os.columnOrdinals[i]
}

func (os *optTableStat) RowCount() uint64 {
	__antithesis_instrumentation__.Notify(551223)
	return os.stat.RowCount
}

func (os *optTableStat) DistinctCount() uint64 {
	__antithesis_instrumentation__.Notify(551224)
	return os.stat.DistinctCount
}

func (os *optTableStat) NullCount() uint64 {
	__antithesis_instrumentation__.Notify(551225)
	return os.stat.NullCount
}

func (os *optTableStat) AvgSize() uint64 {
	__antithesis_instrumentation__.Notify(551226)
	return os.stat.AvgSize
}

func (os *optTableStat) Histogram() []cat.HistogramBucket {
	__antithesis_instrumentation__.Notify(551227)
	return os.stat.Histogram
}

type optFamily struct {
	tab  *optTable
	desc *descpb.ColumnFamilyDescriptor
}

var _ cat.Family = &optFamily{}

func (oi *optFamily) init(tab *optTable, desc *descpb.ColumnFamilyDescriptor) {
	oi.tab = tab
	oi.desc = desc
}

func (oi *optFamily) ID() cat.StableID {
	__antithesis_instrumentation__.Notify(551228)
	return cat.StableID(oi.desc.ID)
}

func (oi *optFamily) Name() tree.Name {
	__antithesis_instrumentation__.Notify(551229)
	return tree.Name(oi.desc.Name)
}

func (oi *optFamily) ColumnCount() int {
	__antithesis_instrumentation__.Notify(551230)
	return len(oi.desc.ColumnIDs)
}

func (oi *optFamily) Column(i int) cat.FamilyColumn {
	__antithesis_instrumentation__.Notify(551231)
	ord, _ := oi.tab.lookupColumnOrdinal(oi.desc.ColumnIDs[i])
	return cat.FamilyColumn{Column: oi.tab.Column(ord), Ordinal: ord}
}

func (oi *optFamily) Table() cat.Table {
	__antithesis_instrumentation__.Notify(551232)
	return oi.tab
}

type optUniqueConstraint struct {
	name string

	table     cat.StableID
	columns   []descpb.ColumnID
	predicate string

	withoutIndex bool
	validity     descpb.ConstraintValidity

	uniquenessGuaranteedByAnotherIndex bool
}

var _ cat.UniqueConstraint = &optUniqueConstraint{}

func (u *optUniqueConstraint) Name() string {
	__antithesis_instrumentation__.Notify(551233)
	return u.name
}

func (u *optUniqueConstraint) TableID() cat.StableID {
	__antithesis_instrumentation__.Notify(551234)
	return u.table
}

func (u *optUniqueConstraint) ColumnCount() int {
	__antithesis_instrumentation__.Notify(551235)
	return len(u.columns)
}

func (u *optUniqueConstraint) ColumnOrdinal(tab cat.Table, i int) int {
	__antithesis_instrumentation__.Notify(551236)
	if tab.ID() != u.table {
		__antithesis_instrumentation__.Notify(551238)
		panic(errors.AssertionFailedf(
			"invalid table %d passed to ColumnOrdinal (expected %d)",
			tab.ID(), u.table,
		))
	} else {
		__antithesis_instrumentation__.Notify(551239)
	}
	__antithesis_instrumentation__.Notify(551237)
	optTab := convertTableToOptTable(tab)
	ord, _ := optTab.lookupColumnOrdinal(u.columns[i])
	return ord
}

func (u *optUniqueConstraint) Predicate() (string, bool) {
	__antithesis_instrumentation__.Notify(551240)
	return u.predicate, u.predicate != ""
}

func (u *optUniqueConstraint) WithoutIndex() bool {
	__antithesis_instrumentation__.Notify(551241)
	return u.withoutIndex
}

func (u *optUniqueConstraint) Validated() bool {
	__antithesis_instrumentation__.Notify(551242)
	return u.validity == descpb.ConstraintValidity_Validated
}

func (u *optUniqueConstraint) UniquenessGuaranteedByAnotherIndex() bool {
	__antithesis_instrumentation__.Notify(551243)
	return u.uniquenessGuaranteedByAnotherIndex
}

type optForeignKeyConstraint struct {
	name string

	originTable   cat.StableID
	originColumns []descpb.ColumnID

	referencedTable   cat.StableID
	referencedColumns []descpb.ColumnID

	validity     descpb.ConstraintValidity
	match        descpb.ForeignKeyReference_Match
	deleteAction catpb.ForeignKeyAction
	updateAction catpb.ForeignKeyAction
}

var _ cat.ForeignKeyConstraint = &optForeignKeyConstraint{}

func (fk *optForeignKeyConstraint) Name() string {
	__antithesis_instrumentation__.Notify(551244)
	return fk.name
}

func (fk *optForeignKeyConstraint) OriginTableID() cat.StableID {
	__antithesis_instrumentation__.Notify(551245)
	return fk.originTable
}

func (fk *optForeignKeyConstraint) ReferencedTableID() cat.StableID {
	__antithesis_instrumentation__.Notify(551246)
	return fk.referencedTable
}

func (fk *optForeignKeyConstraint) ColumnCount() int {
	__antithesis_instrumentation__.Notify(551247)
	return len(fk.originColumns)
}

func (fk *optForeignKeyConstraint) OriginColumnOrdinal(originTable cat.Table, i int) int {
	__antithesis_instrumentation__.Notify(551248)
	if originTable.ID() != fk.originTable {
		__antithesis_instrumentation__.Notify(551250)
		panic(errors.AssertionFailedf(
			"invalid table %d passed to OriginColumnOrdinal (expected %d)",
			originTable.ID(), fk.originTable,
		))
	} else {
		__antithesis_instrumentation__.Notify(551251)
	}
	__antithesis_instrumentation__.Notify(551249)

	tab := convertTableToOptTable(originTable)
	ord, _ := tab.lookupColumnOrdinal(fk.originColumns[i])
	return ord
}

func (fk *optForeignKeyConstraint) ReferencedColumnOrdinal(referencedTable cat.Table, i int) int {
	__antithesis_instrumentation__.Notify(551252)
	if referencedTable.ID() != fk.referencedTable {
		__antithesis_instrumentation__.Notify(551254)
		panic(errors.AssertionFailedf(
			"invalid table %d passed to ReferencedColumnOrdinal (expected %d)",
			referencedTable.ID(), fk.referencedTable,
		))
	} else {
		__antithesis_instrumentation__.Notify(551255)
	}
	__antithesis_instrumentation__.Notify(551253)
	tab := convertTableToOptTable(referencedTable)
	ord, _ := tab.lookupColumnOrdinal(fk.referencedColumns[i])
	return ord
}

func (fk *optForeignKeyConstraint) Validated() bool {
	__antithesis_instrumentation__.Notify(551256)
	return fk.validity == descpb.ConstraintValidity_Validated
}

func (fk *optForeignKeyConstraint) MatchMethod() tree.CompositeKeyMatchMethod {
	__antithesis_instrumentation__.Notify(551257)
	return descpb.ForeignKeyReferenceMatchValue[fk.match]
}

func (fk *optForeignKeyConstraint) DeleteReferenceAction() tree.ReferenceAction {
	__antithesis_instrumentation__.Notify(551258)
	return descpb.ForeignKeyReferenceActionType[fk.deleteAction]
}

func (fk *optForeignKeyConstraint) UpdateReferenceAction() tree.ReferenceAction {
	__antithesis_instrumentation__.Notify(551259)
	return descpb.ForeignKeyReferenceActionType[fk.updateAction]
}

type optVirtualTable struct {
	desc catalog.TableDescriptor

	columns []cat.Column

	id cat.StableID

	name cat.DataSourceName

	indexes []optVirtualIndex

	family optVirtualFamily

	colMap catalog.TableColMap
}

var _ cat.Table = &optVirtualTable{}

func newOptVirtualTable(
	ctx context.Context, oc *optCatalog, desc catalog.TableDescriptor, name *cat.DataSourceName,
) (*optVirtualTable, error) {
	__antithesis_instrumentation__.Notify(551260)

	id := cat.StableID(desc.GetID())
	if name.Catalog() != "" {
		__antithesis_instrumentation__.Notify(551265)

		found, prefix, err := oc.planner.LookupSchema(ctx, name.Catalog(), name.Schema())
		if err != nil {
			__antithesis_instrumentation__.Notify(551267)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(551268)
		}
		__antithesis_instrumentation__.Notify(551266)
		if !found {
			__antithesis_instrumentation__.Notify(551269)

			id |= cat.StableID(math.MaxUint32) << 32
		} else {
			__antithesis_instrumentation__.Notify(551270)
			id |= cat.StableID(prefix.Database.GetID()) << 32
		}
	} else {
		__antithesis_instrumentation__.Notify(551271)
	}
	__antithesis_instrumentation__.Notify(551261)

	ot := &optVirtualTable{
		desc: desc,
		id:   id,
		name: *name,
	}

	ot.columns = make([]cat.Column, len(desc.PublicColumns())+1)

	ot.columns[0].Init(
		0,
		math.MaxInt64,
		"crdb_internal_vtable_pk",
		cat.Ordinary,
		types.Int,
		false,
		cat.Hidden,
		nil,
		nil,
		nil,
		cat.NotGeneratedAsIdentity,
		nil,
	)
	for i, d := range desc.PublicColumns() {
		__antithesis_instrumentation__.Notify(551272)
		ot.columns[i+1].Init(
			i+1,
			cat.StableID(d.GetID()),
			tree.Name(d.GetName()),
			cat.Ordinary,
			d.GetType(),
			d.IsNullable(),
			cat.MaybeHidden(d.IsHidden()),
			d.ColumnDesc().DefaultExpr,
			d.ColumnDesc().ComputeExpr,
			d.ColumnDesc().OnUpdateExpr,
			mapGeneratedAsIdentityType(d.GetGeneratedAsIdentityType()),
			d.ColumnDesc().GeneratedAsIdentitySequenceOption,
		)
	}
	__antithesis_instrumentation__.Notify(551262)

	for i := range ot.columns {
		__antithesis_instrumentation__.Notify(551273)
		ot.colMap.Set(descpb.ColumnID(ot.columns[i].ColID()), i)
	}
	__antithesis_instrumentation__.Notify(551263)

	ot.name.ExplicitSchema = true
	ot.name.ExplicitCatalog = true

	ot.family.init(ot)

	ot.indexes = make([]optVirtualIndex, len(ot.desc.ActiveIndexes()))

	ot.indexes[0] = optVirtualIndex{
		tab:          ot,
		indexOrdinal: 0,
		numCols:      ot.ColumnCount(),
	}

	for _, idx := range ot.desc.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(551274)
		if idx.NumKeyColumns() > 1 {
			__antithesis_instrumentation__.Notify(551276)
			panic(errors.AssertionFailedf("virtual indexes with more than 1 col not supported"))
		} else {
			__antithesis_instrumentation__.Notify(551277)
		}
		__antithesis_instrumentation__.Notify(551275)

		ot.indexes[idx.Ordinal()] = optVirtualIndex{
			tab:          ot,
			idx:          idx,
			indexOrdinal: idx.Ordinal(),

			numCols: ot.ColumnCount(),
		}
	}
	__antithesis_instrumentation__.Notify(551264)

	return ot, nil
}

func (ot *optVirtualTable) ID() cat.StableID {
	__antithesis_instrumentation__.Notify(551278)
	return ot.id
}

func (ot *optVirtualTable) PostgresDescriptorID() cat.StableID {
	__antithesis_instrumentation__.Notify(551279)
	return cat.StableID(ot.desc.GetID())
}

func (ot *optVirtualTable) Equals(other cat.Object) bool {
	__antithesis_instrumentation__.Notify(551280)
	otherTable, ok := other.(*optVirtualTable)
	if !ok {
		__antithesis_instrumentation__.Notify(551284)
		return false
	} else {
		__antithesis_instrumentation__.Notify(551285)
	}
	__antithesis_instrumentation__.Notify(551281)
	if ot == otherTable {
		__antithesis_instrumentation__.Notify(551286)

		return true
	} else {
		__antithesis_instrumentation__.Notify(551287)
	}
	__antithesis_instrumentation__.Notify(551282)
	if ot.id != otherTable.id || func() bool {
		__antithesis_instrumentation__.Notify(551288)
		return ot.desc.GetVersion() != otherTable.desc.GetVersion() == true
	}() == true {
		__antithesis_instrumentation__.Notify(551289)
		return false
	} else {
		__antithesis_instrumentation__.Notify(551290)
	}
	__antithesis_instrumentation__.Notify(551283)

	return true
}

func (ot *optVirtualTable) Name() tree.Name {
	__antithesis_instrumentation__.Notify(551291)
	return ot.name.ObjectName
}

func (ot *optVirtualTable) IsVirtualTable() bool {
	__antithesis_instrumentation__.Notify(551292)
	return true
}

func (ot *optVirtualTable) IsMaterializedView() bool {
	__antithesis_instrumentation__.Notify(551293)
	return false
}

func (ot *optVirtualTable) ColumnCount() int {
	__antithesis_instrumentation__.Notify(551294)
	return len(ot.columns)
}

func (ot *optVirtualTable) Column(i int) *cat.Column {
	__antithesis_instrumentation__.Notify(551295)
	return &ot.columns[i]
}

func (ot *optVirtualTable) getCol(i int) catalog.Column {
	__antithesis_instrumentation__.Notify(551296)
	if i > 0 && func() bool {
		__antithesis_instrumentation__.Notify(551298)
		return i <= len(ot.desc.PublicColumns()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(551299)
		return ot.desc.PublicColumns()[i-1]
	} else {
		__antithesis_instrumentation__.Notify(551300)
	}
	__antithesis_instrumentation__.Notify(551297)
	return nil
}

func (ot *optVirtualTable) IndexCount() int {
	__antithesis_instrumentation__.Notify(551301)

	return len(ot.desc.ActiveIndexes())
}

func (ot *optVirtualTable) WritableIndexCount() int {
	__antithesis_instrumentation__.Notify(551302)

	return 1 + len(ot.desc.WritableNonPrimaryIndexes())
}

func (ot *optVirtualTable) DeletableIndexCount() int {
	__antithesis_instrumentation__.Notify(551303)

	return len(ot.desc.AllIndexes())
}

func (ot *optVirtualTable) Index(i cat.IndexOrdinal) cat.Index {
	__antithesis_instrumentation__.Notify(551304)
	return &ot.indexes[i]
}

func (ot *optVirtualTable) StatisticCount() int {
	__antithesis_instrumentation__.Notify(551305)
	return 0
}

func (ot *optVirtualTable) Statistic(i int) cat.TableStatistic {
	__antithesis_instrumentation__.Notify(551306)
	panic(errors.AssertionFailedf("no stats"))
}

func (ot *optVirtualTable) CheckCount() int {
	__antithesis_instrumentation__.Notify(551307)
	return len(ot.desc.ActiveChecks())
}

func (ot *optVirtualTable) Check(i int) cat.CheckConstraint {
	__antithesis_instrumentation__.Notify(551308)
	check := ot.desc.ActiveChecks()[i]
	return cat.CheckConstraint{
		Constraint: check.Expr,
		Validated:  check.Validity == descpb.ConstraintValidity_Validated,
	}
}

func (ot *optVirtualTable) FamilyCount() int {
	__antithesis_instrumentation__.Notify(551309)
	return 1
}

func (ot *optVirtualTable) Family(i int) cat.Family {
	__antithesis_instrumentation__.Notify(551310)
	return &ot.family
}

func (ot *optVirtualTable) OutboundForeignKeyCount() int {
	__antithesis_instrumentation__.Notify(551311)
	return 0
}

func (ot *optVirtualTable) OutboundForeignKey(i int) cat.ForeignKeyConstraint {
	__antithesis_instrumentation__.Notify(551312)
	panic(errors.AssertionFailedf("no FKs"))
}

func (ot *optVirtualTable) InboundForeignKeyCount() int {
	__antithesis_instrumentation__.Notify(551313)
	return 0
}

func (ot *optVirtualTable) InboundForeignKey(i int) cat.ForeignKeyConstraint {
	__antithesis_instrumentation__.Notify(551314)
	panic(errors.AssertionFailedf("no FKs"))
}

func (ot *optVirtualTable) UniqueCount() int {
	__antithesis_instrumentation__.Notify(551315)
	return 0
}

func (ot *optVirtualTable) Unique(i cat.UniqueOrdinal) cat.UniqueConstraint {
	__antithesis_instrumentation__.Notify(551316)
	panic(errors.AssertionFailedf("no unique constraints"))
}

func (ot *optVirtualTable) Zone() cat.Zone {
	__antithesis_instrumentation__.Notify(551317)
	panic(errors.AssertionFailedf("no zone"))
}

func (ot *optVirtualTable) IsPartitionAllBy() bool {
	__antithesis_instrumentation__.Notify(551318)
	return false
}

func (ot *optVirtualTable) CollectTypes(ord int) (descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(551319)
	col := ot.desc.AllColumns()[ord]
	return collectTypes(col)
}

type optVirtualIndex struct {
	tab *optVirtualTable

	idx catalog.Index

	numCols int

	indexOrdinal int
}

func (oi *optVirtualIndex) ID() cat.StableID {
	__antithesis_instrumentation__.Notify(551320)
	if oi.idx == nil {
		__antithesis_instrumentation__.Notify(551322)

		return cat.StableID(0)
	} else {
		__antithesis_instrumentation__.Notify(551323)
	}
	__antithesis_instrumentation__.Notify(551321)
	return cat.StableID(oi.idx.GetID())
}

func (oi *optVirtualIndex) Name() tree.Name {
	__antithesis_instrumentation__.Notify(551324)
	if oi.idx == nil {
		__antithesis_instrumentation__.Notify(551326)

		return "primary"
	} else {
		__antithesis_instrumentation__.Notify(551327)
	}
	__antithesis_instrumentation__.Notify(551325)
	return tree.Name(oi.idx.GetName())
}

func (oi *optVirtualIndex) IsUnique() bool {
	__antithesis_instrumentation__.Notify(551328)
	if oi.idx == nil {
		__antithesis_instrumentation__.Notify(551330)

		return false
	} else {
		__antithesis_instrumentation__.Notify(551331)
	}
	__antithesis_instrumentation__.Notify(551329)
	return oi.idx.IsUnique()
}

func (oi *optVirtualIndex) IsInverted() bool {
	__antithesis_instrumentation__.Notify(551332)
	return false
}

func (oi *optVirtualIndex) ExplicitColumnCount() int {
	__antithesis_instrumentation__.Notify(551333)
	return oi.idx.NumKeyColumns()
}

func (oi *optVirtualIndex) ColumnCount() int {
	__antithesis_instrumentation__.Notify(551334)
	return oi.numCols
}

func (oi *optVirtualIndex) KeyColumnCount() int {
	__antithesis_instrumentation__.Notify(551335)

	return 2
}

func (oi *optVirtualIndex) LaxKeyColumnCount() int {
	__antithesis_instrumentation__.Notify(551336)

	return 2
}

func (oi *optVirtualIndex) NonInvertedPrefixColumnCount() int {
	__antithesis_instrumentation__.Notify(551337)
	panic("virtual indexes are not inverted")
}

func (ot *optVirtualTable) lookupColumnOrdinal(colID descpb.ColumnID) (int, error) {
	__antithesis_instrumentation__.Notify(551338)
	col, ok := ot.colMap.Get(colID)
	if ok {
		__antithesis_instrumentation__.Notify(551340)
		return col, nil
	} else {
		__antithesis_instrumentation__.Notify(551341)
	}
	__antithesis_instrumentation__.Notify(551339)
	return col, pgerror.Newf(pgcode.UndefinedColumn,
		"column [%d] does not exist", colID)
}

func (oi *optVirtualIndex) Column(i int) cat.IndexColumn {
	__antithesis_instrumentation__.Notify(551342)
	if oi.idx == nil {
		__antithesis_instrumentation__.Notify(551346)

		return cat.IndexColumn{Column: oi.tab.Column(i)}
	} else {
		__antithesis_instrumentation__.Notify(551347)
	}
	__antithesis_instrumentation__.Notify(551343)
	length := oi.idx.NumKeyColumns()
	if i < length {
		__antithesis_instrumentation__.Notify(551348)
		ord, _ := oi.tab.lookupColumnOrdinal(oi.idx.GetKeyColumnID(i))
		return cat.IndexColumn{
			Column: oi.tab.Column(ord),
		}
	} else {
		__antithesis_instrumentation__.Notify(551349)
	}
	__antithesis_instrumentation__.Notify(551344)
	if i == length {
		__antithesis_instrumentation__.Notify(551350)

		return cat.IndexColumn{Column: oi.tab.Column(0)}
	} else {
		__antithesis_instrumentation__.Notify(551351)
	}
	__antithesis_instrumentation__.Notify(551345)

	i -= length + 1
	ord, _ := oi.tab.lookupColumnOrdinal(oi.idx.GetStoredColumnID(i))
	return cat.IndexColumn{Column: oi.tab.Column(ord)}
}

func (oi *optVirtualIndex) InvertedColumn() cat.IndexColumn {
	__antithesis_instrumentation__.Notify(551352)
	panic(errors.AssertionFailedf("virtual indexes are not inverted"))
}

func (oi *optVirtualIndex) Predicate() (string, bool) {
	__antithesis_instrumentation__.Notify(551353)
	return "", false
}

func (oi *optVirtualIndex) Zone() cat.Zone {
	__antithesis_instrumentation__.Notify(551354)
	panic(errors.AssertionFailedf("no zone"))
}

func (oi *optVirtualIndex) Span() roachpb.Span {
	__antithesis_instrumentation__.Notify(551355)
	panic(errors.AssertionFailedf("no span"))
}

func (oi *optVirtualIndex) Table() cat.Table {
	__antithesis_instrumentation__.Notify(551356)
	return oi.tab
}

func (oi *optVirtualIndex) Ordinal() int {
	__antithesis_instrumentation__.Notify(551357)
	return oi.indexOrdinal
}

func (oi *optVirtualIndex) ImplicitColumnCount() int {
	__antithesis_instrumentation__.Notify(551358)
	return 0
}

func (oi *optVirtualIndex) ImplicitPartitioningColumnCount() int {
	__antithesis_instrumentation__.Notify(551359)
	return 0
}

func (oi *optVirtualIndex) GeoConfig() *geoindex.Config {
	__antithesis_instrumentation__.Notify(551360)
	return nil
}

func (oi *optVirtualIndex) Version() descpb.IndexDescriptorVersion {
	__antithesis_instrumentation__.Notify(551361)
	return 0
}

func (oi *optVirtualIndex) PartitionCount() int {
	__antithesis_instrumentation__.Notify(551362)
	return 0
}

func (oi *optVirtualIndex) Partition(i int) cat.Partition {
	__antithesis_instrumentation__.Notify(551363)
	return nil
}

type optVirtualFamily struct {
	tab *optVirtualTable
}

var _ cat.Family = &optVirtualFamily{}

func (oi *optVirtualFamily) init(tab *optVirtualTable) {
	oi.tab = tab
}

func (oi *optVirtualFamily) ID() cat.StableID {
	__antithesis_instrumentation__.Notify(551364)
	return 0
}

func (oi *optVirtualFamily) Name() tree.Name {
	__antithesis_instrumentation__.Notify(551365)
	return "primary"
}

func (oi *optVirtualFamily) ColumnCount() int {
	__antithesis_instrumentation__.Notify(551366)
	return oi.tab.ColumnCount()
}

func (oi *optVirtualFamily) Column(i int) cat.FamilyColumn {
	__antithesis_instrumentation__.Notify(551367)
	return cat.FamilyColumn{Column: oi.tab.Column(i), Ordinal: i}
}

func (oi *optVirtualFamily) Table() cat.Table {
	__antithesis_instrumentation__.Notify(551368)
	return oi.tab
}

type optCatalogTableInterface interface {
	getCol(i int) catalog.Column
}

var _ optCatalogTableInterface = &optTable{}
var _ optCatalogTableInterface = &optVirtualTable{}

func collectTypes(col catalog.Column) (descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(551369)
	visitor := &tree.TypeCollectorVisitor{
		OIDs: make(map[oid.Oid]struct{}),
	}
	addOIDsInExpr := func(exprStr string) error {
		__antithesis_instrumentation__.Notify(551375)
		expr, err := parser.ParseExpr(exprStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(551377)
			return err
		} else {
			__antithesis_instrumentation__.Notify(551378)
		}
		__antithesis_instrumentation__.Notify(551376)
		tree.WalkExpr(visitor, expr)
		return nil
	}
	__antithesis_instrumentation__.Notify(551370)

	if col.HasDefault() {
		__antithesis_instrumentation__.Notify(551379)
		if err := addOIDsInExpr(col.GetDefaultExpr()); err != nil {
			__antithesis_instrumentation__.Notify(551380)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(551381)
		}
	} else {
		__antithesis_instrumentation__.Notify(551382)
	}
	__antithesis_instrumentation__.Notify(551371)
	if col.IsComputed() {
		__antithesis_instrumentation__.Notify(551383)
		if err := addOIDsInExpr(col.GetComputeExpr()); err != nil {
			__antithesis_instrumentation__.Notify(551384)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(551385)
		}
	} else {
		__antithesis_instrumentation__.Notify(551386)
	}
	__antithesis_instrumentation__.Notify(551372)
	if typ := col.GetType(); typ != nil && func() bool {
		__antithesis_instrumentation__.Notify(551387)
		return typ.UserDefined() == true
	}() == true {
		__antithesis_instrumentation__.Notify(551388)
		visitor.OIDs[typ.Oid()] = struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(551389)
	}
	__antithesis_instrumentation__.Notify(551373)

	ids := make(descpb.IDs, 0, len(visitor.OIDs))
	for collectedOid := range visitor.OIDs {
		__antithesis_instrumentation__.Notify(551390)
		id, err := typedesc.UserDefinedTypeOIDToID(collectedOid)
		if err != nil {
			__antithesis_instrumentation__.Notify(551392)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(551393)
		}
		__antithesis_instrumentation__.Notify(551391)
		ids = append(ids, id)
	}
	__antithesis_instrumentation__.Notify(551374)
	return ids, nil
}

func mapGeneratedAsIdentityType(inType catpb.GeneratedAsIdentityType) cat.GeneratedAsIdentityType {
	__antithesis_instrumentation__.Notify(551394)
	mapGeneratedAsIdentityType := map[catpb.GeneratedAsIdentityType]cat.GeneratedAsIdentityType{
		catpb.GeneratedAsIdentityType_NOT_IDENTITY_COLUMN:  cat.NotGeneratedAsIdentity,
		catpb.GeneratedAsIdentityType_GENERATED_ALWAYS:     cat.GeneratedAlwaysAsIdentity,
		catpb.GeneratedAsIdentityType_GENERATED_BY_DEFAULT: cat.GeneratedByDefaultAsIdentity,
	}
	return mapGeneratedAsIdentityType[inType]
}
