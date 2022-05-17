package sctestdeps

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps/sctestutils"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec/scmutationexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scrun"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

var _ scbuild.Dependencies = (*TestState)(nil)

func (s *TestState) AuthorizationAccessor() scbuild.AuthorizationAccessor {
	__antithesis_instrumentation__.Notify(580864)
	return s
}

func (s *TestState) CatalogReader() scbuild.CatalogReader {
	__antithesis_instrumentation__.Notify(580865)
	return s
}

func (s *TestState) ClusterID() uuid.UUID {
	__antithesis_instrumentation__.Notify(580866)
	return uuid.Nil
}

func (s *TestState) Codec() keys.SQLCodec {
	__antithesis_instrumentation__.Notify(580867)
	return keys.SystemSQLCodec
}

func (s *TestState) SessionData() *sessiondata.SessionData {
	__antithesis_instrumentation__.Notify(580868)
	return &s.sessionData
}

func (s *TestState) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(580869)
	return cluster.MakeTestingClusterSettings()
}

func (s *TestState) Statements() []string {
	__antithesis_instrumentation__.Notify(580870)
	return s.statements
}

func (s *TestState) IncrementSchemaChangeAlterCounter(counterType string, extra ...string) {
	__antithesis_instrumentation__.Notify(580871)
	var maybeExtra string
	if len(extra) > 0 {
		__antithesis_instrumentation__.Notify(580873)
		maybeExtra = "." + extra[0]
	} else {
		__antithesis_instrumentation__.Notify(580874)
	}
	__antithesis_instrumentation__.Notify(580872)
	s.LogSideEffectf("increment telemetry for sql.schema.alter_%s%s", counterType, maybeExtra)
}

func (s *TestState) IncrementSchemaChangeDropCounter(counterType string) {
	__antithesis_instrumentation__.Notify(580875)
	s.LogSideEffectf("increment telemetry for sql.schema.drop_%s", counterType)
}

func (s *TestState) IncrementUserDefinedSchemaCounter(
	counterType sqltelemetry.UserDefinedSchemaTelemetryType,
) {
	__antithesis_instrumentation__.Notify(580876)
	s.LogSideEffectf("increment telemetry for sql.uds.%s", counterType)
}

func (s *TestState) IncrementEnumCounter(counterType sqltelemetry.EnumTelemetryType) {
	__antithesis_instrumentation__.Notify(580877)
	s.LogSideEffectf("increment telemetry for sql.udts.%s", counterType)
}

var _ scbuild.AuthorizationAccessor = (*TestState)(nil)

func (s *TestState) CheckPrivilege(
	ctx context.Context, descriptor catalog.Descriptor, privilege privilege.Kind,
) error {
	__antithesis_instrumentation__.Notify(580878)
	return nil
}

func (s *TestState) HasAdminRole(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(580879)
	return true, nil
}

func (s *TestState) HasOwnership(ctx context.Context, descriptor catalog.Descriptor) (bool, error) {
	__antithesis_instrumentation__.Notify(580880)
	return true, nil
}

func (s *TestState) IndexPartitioningCCLCallback() scbuild.CreatePartitioningCCLCallback {
	__antithesis_instrumentation__.Notify(580881)
	if ccl := scdeps.CreatePartitioningCCL; ccl != nil {
		__antithesis_instrumentation__.Notify(580883)
		return ccl
	} else {
		__antithesis_instrumentation__.Notify(580884)
	}
	__antithesis_instrumentation__.Notify(580882)
	return func(
		ctx context.Context,
		st *cluster.Settings,
		evalCtx *tree.EvalContext,
		columnLookupFn func(tree.Name) (catalog.Column, error),
		oldNumImplicitColumns int,
		oldKeyColumnNames []string,
		partBy *tree.PartitionBy,
		allowedNewColumnNames []tree.Name,
		allowImplicitPartitioning bool,
	) (newImplicitCols []catalog.Column, newPartitioning catpb.PartitioningDescriptor, err error) {
		__antithesis_instrumentation__.Notify(580885)
		newPartitioning.NumColumns = uint32(len(partBy.Fields))
		return nil, newPartitioning, nil
	}
}

var _ scbuild.CatalogReader = (*TestState)(nil)

func (s *TestState) MayResolveDatabase(
	ctx context.Context, name tree.Name,
) catalog.DatabaseDescriptor {
	__antithesis_instrumentation__.Notify(580886)
	desc := s.mayGetByName(0, 0, name.String())
	if desc == nil {
		__antithesis_instrumentation__.Notify(580890)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(580891)
	}
	__antithesis_instrumentation__.Notify(580887)
	db, err := catalog.AsDatabaseDescriptor(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(580892)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580893)
	}
	__antithesis_instrumentation__.Notify(580888)
	if db.Dropped() || func() bool {
		__antithesis_instrumentation__.Notify(580894)
		return db.Offline() == true
	}() == true {
		__antithesis_instrumentation__.Notify(580895)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(580896)
	}
	__antithesis_instrumentation__.Notify(580889)
	return db
}

func (s *TestState) MayResolveSchema(
	ctx context.Context, name tree.ObjectNamePrefix,
) (catalog.DatabaseDescriptor, catalog.SchemaDescriptor) {
	__antithesis_instrumentation__.Notify(580897)
	dbName := name.Catalog()
	scName := name.Schema()
	if !name.ExplicitCatalog && func() bool {
		__antithesis_instrumentation__.Notify(580904)
		return !name.ExplicitSchema == true
	}() == true {
		__antithesis_instrumentation__.Notify(580905)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(580906)
	}
	__antithesis_instrumentation__.Notify(580898)
	if !name.ExplicitCatalog || func() bool {
		__antithesis_instrumentation__.Notify(580907)
		return !name.ExplicitSchema == true
	}() == true {
		__antithesis_instrumentation__.Notify(580908)
		dbName = s.CurrentDatabase()
		if name.ExplicitCatalog {
			__antithesis_instrumentation__.Notify(580909)
			scName = name.Catalog()
		} else {
			__antithesis_instrumentation__.Notify(580910)
			scName = name.Schema()
		}
	} else {
		__antithesis_instrumentation__.Notify(580911)
	}
	__antithesis_instrumentation__.Notify(580899)
	dbDesc := s.mayGetByName(0, 0, dbName)
	if dbDesc == nil || func() bool {
		__antithesis_instrumentation__.Notify(580912)
		return dbDesc.Dropped() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(580913)
		return dbDesc.Offline() == true
	}() == true {
		__antithesis_instrumentation__.Notify(580914)
		if dbName == s.CurrentDatabase() {
			__antithesis_instrumentation__.Notify(580916)
			panic(errors.AssertionFailedf("Invalid current database %q", s.CurrentDatabase()))
		} else {
			__antithesis_instrumentation__.Notify(580917)
		}
		__antithesis_instrumentation__.Notify(580915)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(580918)
	}
	__antithesis_instrumentation__.Notify(580900)
	db, err := catalog.AsDatabaseDescriptor(dbDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(580919)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580920)
	}
	__antithesis_instrumentation__.Notify(580901)
	scDesc := s.mayGetByName(db.GetID(), 0, scName)
	if scDesc == nil || func() bool {
		__antithesis_instrumentation__.Notify(580921)
		return scDesc.Dropped() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(580922)
		return scDesc.Offline() == true
	}() == true {
		__antithesis_instrumentation__.Notify(580923)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(580924)
	}
	__antithesis_instrumentation__.Notify(580902)
	sc, err := catalog.AsSchemaDescriptor(scDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(580925)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580926)
	}
	__antithesis_instrumentation__.Notify(580903)
	return db, sc
}

func (s *TestState) MayResolveTable(
	ctx context.Context, name tree.UnresolvedObjectName,
) (catalog.ResolvedObjectPrefix, catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(580927)
	prefix, desc, err := s.mayResolveObject(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(580931)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580932)
	}
	__antithesis_instrumentation__.Notify(580928)
	if desc == nil {
		__antithesis_instrumentation__.Notify(580933)
		return prefix, nil
	} else {
		__antithesis_instrumentation__.Notify(580934)
	}
	__antithesis_instrumentation__.Notify(580929)
	table, err := catalog.AsTableDescriptor(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(580935)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580936)
	}
	__antithesis_instrumentation__.Notify(580930)
	return prefix, table
}

func (s *TestState) MayResolveType(
	ctx context.Context, name tree.UnresolvedObjectName,
) (catalog.ResolvedObjectPrefix, catalog.TypeDescriptor) {
	__antithesis_instrumentation__.Notify(580937)
	prefix, desc, err := s.mayResolveObject(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(580941)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580942)
	}
	__antithesis_instrumentation__.Notify(580938)
	if desc == nil {
		__antithesis_instrumentation__.Notify(580943)
		return prefix, nil
	} else {
		__antithesis_instrumentation__.Notify(580944)
	}
	__antithesis_instrumentation__.Notify(580939)
	typ, err := catalog.AsTypeDescriptor(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(580945)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(580946)
	}
	__antithesis_instrumentation__.Notify(580940)
	return prefix, typ
}

func (s *TestState) mayResolveObject(
	name tree.UnresolvedObjectName,
) (prefix catalog.ResolvedObjectPrefix, desc catalog.Descriptor, err error) {
	__antithesis_instrumentation__.Notify(580947)
	tn := name.ToTableName()
	{
		__antithesis_instrumentation__.Notify(580951)
		db, sc := s.mayResolvePrefix(tn.ObjectNamePrefix)
		if db == nil || func() bool {
			__antithesis_instrumentation__.Notify(580954)
			return sc == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(580955)
			return catalog.ResolvedObjectPrefix{}, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(580956)
		}
		__antithesis_instrumentation__.Notify(580952)
		prefix.ExplicitDatabase = true
		prefix.ExplicitSchema = true
		prefix.Database, err = catalog.AsDatabaseDescriptor(db)
		if err != nil {
			__antithesis_instrumentation__.Notify(580957)
			return catalog.ResolvedObjectPrefix{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(580958)
		}
		__antithesis_instrumentation__.Notify(580953)
		prefix.Schema, err = catalog.AsSchemaDescriptor(sc)
		if err != nil {
			__antithesis_instrumentation__.Notify(580959)
			return catalog.ResolvedObjectPrefix{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(580960)
		}
	}
	__antithesis_instrumentation__.Notify(580948)
	desc = s.mayGetByName(prefix.Database.GetID(), prefix.Schema.GetID(), name.Object())
	if desc == nil {
		__antithesis_instrumentation__.Notify(580961)
		return prefix, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(580962)
	}
	__antithesis_instrumentation__.Notify(580949)
	if desc.Dropped() || func() bool {
		__antithesis_instrumentation__.Notify(580963)
		return desc.Offline() == true
	}() == true {
		__antithesis_instrumentation__.Notify(580964)
		return prefix, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(580965)
	}
	__antithesis_instrumentation__.Notify(580950)
	return prefix, desc, nil
}

func (s *TestState) mayResolvePrefix(name tree.ObjectNamePrefix) (db, sc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(580966)
	if name.ExplicitCatalog && func() bool {
		__antithesis_instrumentation__.Notify(580974)
		return name.ExplicitSchema == true
	}() == true {
		__antithesis_instrumentation__.Notify(580975)
		db = s.mayGetByName(0, 0, name.Catalog())
		if db == nil || func() bool {
			__antithesis_instrumentation__.Notify(580978)
			return db.Dropped() == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(580979)
			return db.Offline() == true
		}() == true {
			__antithesis_instrumentation__.Notify(580980)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(580981)
		}
		__antithesis_instrumentation__.Notify(580976)
		sc = s.mayGetByName(db.GetID(), 0, name.Schema())
		if sc == nil || func() bool {
			__antithesis_instrumentation__.Notify(580982)
			return sc.Dropped() == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(580983)
			return sc.Offline() == true
		}() == true {
			__antithesis_instrumentation__.Notify(580984)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(580985)
		}
		__antithesis_instrumentation__.Notify(580977)
		return db, sc
	} else {
		__antithesis_instrumentation__.Notify(580986)
	}
	__antithesis_instrumentation__.Notify(580967)

	db = s.mayGetByName(0, 0, s.CurrentDatabase())
	if db == nil || func() bool {
		__antithesis_instrumentation__.Notify(580987)
		return db.Dropped() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(580988)
		return db.Offline() == true
	}() == true {
		__antithesis_instrumentation__.Notify(580989)
		panic(errors.AssertionFailedf("Invalid current database %q", s.CurrentDatabase()))
	} else {
		__antithesis_instrumentation__.Notify(580990)
	}
	__antithesis_instrumentation__.Notify(580968)

	if !name.ExplicitCatalog && func() bool {
		__antithesis_instrumentation__.Notify(580991)
		return !name.ExplicitSchema == true
	}() == true {
		__antithesis_instrumentation__.Notify(580992)
		sc = s.mayGetByName(db.GetID(), 0, catconstants.PublicSchemaName)
		if sc == nil || func() bool {
			__antithesis_instrumentation__.Notify(580994)
			return sc.Dropped() == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(580995)
			return sc.Offline() == true
		}() == true {
			__antithesis_instrumentation__.Notify(580996)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(580997)
		}
		__antithesis_instrumentation__.Notify(580993)
		return db, sc
	} else {
		__antithesis_instrumentation__.Notify(580998)
	}
	__antithesis_instrumentation__.Notify(580969)

	var prefixName string
	if name.ExplicitCatalog {
		__antithesis_instrumentation__.Notify(580999)
		prefixName = name.Catalog()
	} else {
		__antithesis_instrumentation__.Notify(581000)
		prefixName = name.Schema()
	}
	__antithesis_instrumentation__.Notify(580970)

	sc = s.mayGetByName(db.GetID(), 0, prefixName)
	if sc != nil && func() bool {
		__antithesis_instrumentation__.Notify(581001)
		return !sc.Dropped() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(581002)
		return !sc.Offline() == true
	}() == true {
		__antithesis_instrumentation__.Notify(581003)
		return db, sc
	} else {
		__antithesis_instrumentation__.Notify(581004)
	}
	__antithesis_instrumentation__.Notify(580971)

	db = s.mayGetByName(0, 0, prefixName)
	if db == nil || func() bool {
		__antithesis_instrumentation__.Notify(581005)
		return db.Dropped() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(581006)
		return db.Offline() == true
	}() == true {
		__antithesis_instrumentation__.Notify(581007)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(581008)
	}
	__antithesis_instrumentation__.Notify(580972)
	sc = s.mayGetByName(db.GetID(), 0, catconstants.PublicSchemaName)
	if sc == nil || func() bool {
		__antithesis_instrumentation__.Notify(581009)
		return sc.Dropped() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(581010)
		return sc.Offline() == true
	}() == true {
		__antithesis_instrumentation__.Notify(581011)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(581012)
	}
	__antithesis_instrumentation__.Notify(580973)
	return db, sc
}

func (s *TestState) mayGetByName(
	parentID, parentSchemaID descpb.ID, name string,
) catalog.Descriptor {
	__antithesis_instrumentation__.Notify(581013)
	key := descpb.NameInfo{
		ParentID:       parentID,
		ParentSchemaID: parentSchemaID,
		Name:           name,
	}
	ne := s.catalog.LookupNamespaceEntry(key)
	if ne == nil {
		__antithesis_instrumentation__.Notify(581018)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(581019)
	}
	__antithesis_instrumentation__.Notify(581014)
	id := ne.GetID()
	if id == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(581020)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(581021)
	}
	__antithesis_instrumentation__.Notify(581015)
	if id == keys.PublicSchemaID {
		__antithesis_instrumentation__.Notify(581022)
		return schemadesc.GetPublicSchema()
	} else {
		__antithesis_instrumentation__.Notify(581023)
	}
	__antithesis_instrumentation__.Notify(581016)
	b := s.descBuilder(id)
	if b == nil {
		__antithesis_instrumentation__.Notify(581024)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(581025)
	}
	__antithesis_instrumentation__.Notify(581017)
	return b.BuildImmutable()
}

func (s *TestState) ReadObjectNamesAndIDs(
	ctx context.Context, db catalog.DatabaseDescriptor, schema catalog.SchemaDescriptor,
) (names tree.TableNames, ids descpb.IDs) {
	__antithesis_instrumentation__.Notify(581026)
	m := make(map[string]descpb.ID)
	_ = s.catalog.ForEachNamespaceEntry(func(e catalog.NameEntry) error {
		__antithesis_instrumentation__.Notify(581030)
		if e.GetParentID() == db.GetID() && func() bool {
			__antithesis_instrumentation__.Notify(581032)
			return e.GetParentSchemaID() == schema.GetID() == true
		}() == true {
			__antithesis_instrumentation__.Notify(581033)
			m[e.GetName()] = e.GetID()
			names = append(names, tree.MakeTableNameWithSchema(
				tree.Name(db.GetName()),
				tree.Name(schema.GetName()),
				tree.Name(e.GetName()),
			))
		} else {
			__antithesis_instrumentation__.Notify(581034)
		}
		__antithesis_instrumentation__.Notify(581031)
		return nil
	})
	__antithesis_instrumentation__.Notify(581027)
	sort.Slice(names, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(581035)
		return names[i].Object() < names[j].Object()
	})
	__antithesis_instrumentation__.Notify(581028)
	for _, name := range names {
		__antithesis_instrumentation__.Notify(581036)
		ids = append(ids, m[name.Object()])
	}
	__antithesis_instrumentation__.Notify(581029)
	return names, ids
}

func (s *TestState) ResolveType(
	ctx context.Context, name *tree.UnresolvedObjectName,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(581037)
	prefix, obj, err := s.mayResolveObject(*name)
	if err != nil {
		__antithesis_instrumentation__.Notify(581041)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581042)
	}
	__antithesis_instrumentation__.Notify(581038)
	if obj == nil {
		__antithesis_instrumentation__.Notify(581043)
		return nil, errors.Wrapf(catalog.ErrDescriptorNotFound, "resolving type %q", name.String())
	} else {
		__antithesis_instrumentation__.Notify(581044)
	}
	__antithesis_instrumentation__.Notify(581039)
	typ, err := catalog.AsTypeDescriptor(obj)
	if err != nil {
		__antithesis_instrumentation__.Notify(581045)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581046)
	}
	__antithesis_instrumentation__.Notify(581040)
	tn := tree.MakeQualifiedTypeName(prefix.Database.GetName(), prefix.Schema.GetName(), typ.GetName())
	return typ.MakeTypesT(ctx, &tn, s)
}

func (s *TestState) ResolveTypeByOID(ctx context.Context, oid oid.Oid) (*types.T, error) {
	__antithesis_instrumentation__.Notify(581047)
	id, err := typedesc.UserDefinedTypeOIDToID(oid)
	if err != nil {
		__antithesis_instrumentation__.Notify(581050)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581051)
	}
	__antithesis_instrumentation__.Notify(581048)
	name, typ, err := s.GetTypeDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(581052)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581053)
	}
	__antithesis_instrumentation__.Notify(581049)
	return typ.MakeTypesT(ctx, &name, s)
}

var _ catalog.TypeDescriptorResolver = (*TestState)(nil)

func (s *TestState) GetTypeDescriptor(
	ctx context.Context, id descpb.ID,
) (tree.TypeName, catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(581054)
	desc, err := s.mustReadImmutableDescriptor(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(581058)
		return tree.TypeName{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(581059)
	}
	__antithesis_instrumentation__.Notify(581055)
	typ, err := catalog.AsTypeDescriptor(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(581060)
		return tree.TypeName{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(581061)
	}
	__antithesis_instrumentation__.Notify(581056)
	tn, err := s.getQualifiedObjectNameByID(typ.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(581062)
		return tree.TypeName{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(581063)
	}
	__antithesis_instrumentation__.Notify(581057)
	return tree.MakeTypeNameWithPrefix(tn.ObjectNamePrefix, tn.Object()), typ, nil
}

func (s *TestState) GetQualifiedTableNameByID(
	ctx context.Context, id int64, requiredType tree.RequiredTableKind,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(581064)
	return s.getQualifiedObjectNameByID(descpb.ID(id))
}

func (s *TestState) getQualifiedObjectNameByID(id descpb.ID) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(581065)
	obj, err := s.mustReadImmutableDescriptor(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(581069)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581070)
	}
	__antithesis_instrumentation__.Notify(581066)
	db, err := s.mustReadImmutableDescriptor(obj.GetParentID())
	if err != nil {
		__antithesis_instrumentation__.Notify(581071)
		return nil, errors.Wrapf(err, "parent database for object #%d", id)
	} else {
		__antithesis_instrumentation__.Notify(581072)
	}
	__antithesis_instrumentation__.Notify(581067)
	sc, err := s.mustReadImmutableDescriptor(obj.GetParentSchemaID())
	if err != nil {
		__antithesis_instrumentation__.Notify(581073)
		return nil, errors.Wrapf(err, "parent schema for object #%d", id)
	} else {
		__antithesis_instrumentation__.Notify(581074)
	}
	__antithesis_instrumentation__.Notify(581068)
	return tree.NewTableNameWithSchema(tree.Name(db.GetName()), tree.Name(sc.GetName()), tree.Name(obj.GetName())), nil
}

func (s *TestState) CurrentDatabase() string {
	__antithesis_instrumentation__.Notify(581075)
	return s.currentDatabase
}

func (s *TestState) MustGetSchemasForDatabase(
	ctx context.Context, database catalog.DatabaseDescriptor,
) map[descpb.ID]string {
	__antithesis_instrumentation__.Notify(581076)
	schemas := make(map[descpb.ID]string)
	err := database.ForEachNonDroppedSchema(func(id descpb.ID, name string) error {
		__antithesis_instrumentation__.Notify(581079)
		schemas[id] = name
		return nil
	})
	__antithesis_instrumentation__.Notify(581077)
	if err != nil {
		__antithesis_instrumentation__.Notify(581080)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(581081)
	}
	__antithesis_instrumentation__.Notify(581078)
	return schemas
}

func (s *TestState) MustReadDescriptor(ctx context.Context, id descpb.ID) catalog.Descriptor {
	__antithesis_instrumentation__.Notify(581082)
	desc, err := s.mustReadImmutableDescriptor(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(581084)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(581085)
	}
	__antithesis_instrumentation__.Notify(581083)
	return desc
}

func (s *TestState) mustReadMutableDescriptor(id descpb.ID) (catalog.MutableDescriptor, error) {
	__antithesis_instrumentation__.Notify(581086)
	if s.synthetic.LookupDescriptorEntry(id) != nil {
		__antithesis_instrumentation__.Notify(581089)
		return nil, errors.AssertionFailedf("attempted mutable access of synthetic descriptor %d", id)
	} else {
		__antithesis_instrumentation__.Notify(581090)
	}
	__antithesis_instrumentation__.Notify(581087)
	b := s.descBuilder(id)
	if b == nil {
		__antithesis_instrumentation__.Notify(581091)
		return nil, errors.Wrapf(catalog.ErrDescriptorNotFound, "reading mutable descriptor #%d", id)
	} else {
		__antithesis_instrumentation__.Notify(581092)
	}
	__antithesis_instrumentation__.Notify(581088)
	return b.BuildExistingMutable(), nil
}

func (s *TestState) mustReadImmutableDescriptor(id descpb.ID) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(581093)
	b := s.descBuilderWithSynthetic(id)
	if b == nil {
		__antithesis_instrumentation__.Notify(581095)
		return nil, errors.Wrapf(catalog.ErrDescriptorNotFound, "reading immutable descriptor #%d", id)
	} else {
		__antithesis_instrumentation__.Notify(581096)
	}
	__antithesis_instrumentation__.Notify(581094)
	return b.BuildImmutable(), nil
}

func (s *TestState) descBuilder(id descpb.ID) catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(581097)
	if desc := s.catalog.LookupDescriptorEntry(id); desc != nil {
		__antithesis_instrumentation__.Notify(581099)
		return desc.NewBuilder()
	} else {
		__antithesis_instrumentation__.Notify(581100)
	}
	__antithesis_instrumentation__.Notify(581098)
	return nil
}

func (s *TestState) descBuilderWithSynthetic(id descpb.ID) catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(581101)
	if desc := s.synthetic.LookupDescriptorEntry(id); desc != nil {
		__antithesis_instrumentation__.Notify(581103)
		return desc.NewBuilder()
	} else {
		__antithesis_instrumentation__.Notify(581104)
	}
	__antithesis_instrumentation__.Notify(581102)
	return s.descBuilder(id)
}

var _ scexec.Dependencies = (*TestState)(nil)

func (s *TestState) Clock() scmutationexec.Clock {
	__antithesis_instrumentation__.Notify(581105)
	return s
}

func (s *TestState) ApproximateTime() time.Time {
	__antithesis_instrumentation__.Notify(581106)
	return s.approximateTimestamp
}

func (s *TestState) Catalog() scexec.Catalog {
	__antithesis_instrumentation__.Notify(581107)
	return s
}

var _ scexec.Catalog = (*TestState)(nil)

func (s *TestState) MustReadImmutableDescriptors(
	ctx context.Context, ids ...descpb.ID,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(581108)
	out := make([]catalog.Descriptor, 0, len(ids))
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(581110)
		d, err := s.mustReadImmutableDescriptor(id)
		if err != nil {
			__antithesis_instrumentation__.Notify(581112)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(581113)
		}
		__antithesis_instrumentation__.Notify(581111)
		out = append(out, d)
	}
	__antithesis_instrumentation__.Notify(581109)
	return out, nil
}

func (s *TestState) AddSyntheticDescriptor(desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(581114)
	s.synthetic.UpsertDescriptorEntry(desc)
}

func (s *TestState) RemoveSyntheticDescriptor(id descpb.ID) {
	__antithesis_instrumentation__.Notify(581115)
	s.synthetic.DeleteDescriptorEntry(id)
}

func (s *TestState) MustReadMutableDescriptor(
	ctx context.Context, id descpb.ID,
) (catalog.MutableDescriptor, error) {
	__antithesis_instrumentation__.Notify(581116)
	return s.mustReadMutableDescriptor(id)
}

func (s *TestState) GetFullyQualifiedName(ctx context.Context, id descpb.ID) (string, error) {
	__antithesis_instrumentation__.Notify(581117)
	obj, err := s.mustReadImmutableDescriptor(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(581122)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(581123)
	}
	__antithesis_instrumentation__.Notify(581118)
	dbName := ""
	if obj.GetParentID() != descpb.InvalidID {
		__antithesis_instrumentation__.Notify(581124)
		db, err := s.mustReadImmutableDescriptor(obj.GetParentID())
		if err != nil {
			__antithesis_instrumentation__.Notify(581126)
			return "", errors.Wrapf(err, "parent database for object #%d", id)
		} else {
			__antithesis_instrumentation__.Notify(581127)
		}
		__antithesis_instrumentation__.Notify(581125)
		dbName = db.GetName()
	} else {
		__antithesis_instrumentation__.Notify(581128)
	}
	__antithesis_instrumentation__.Notify(581119)
	scName := ""
	if obj.GetParentSchemaID() != descpb.InvalidID {
		__antithesis_instrumentation__.Notify(581129)
		scName = tree.PublicSchema
		if obj.GetParentSchemaID() != keys.PublicSchemaID {
			__antithesis_instrumentation__.Notify(581130)
			sc, err := s.mustReadImmutableDescriptor(obj.GetParentSchemaID())
			if err != nil {
				__antithesis_instrumentation__.Notify(581132)
				return "", errors.Wrapf(err, "parent schema for object #%d", id)
			} else {
				__antithesis_instrumentation__.Notify(581133)
			}
			__antithesis_instrumentation__.Notify(581131)
			scName = sc.GetName()
		} else {
			__antithesis_instrumentation__.Notify(581134)
		}
	} else {
		__antithesis_instrumentation__.Notify(581135)
	}
	__antithesis_instrumentation__.Notify(581120)

	switch obj.DescriptorType() {
	case catalog.Table:
		__antithesis_instrumentation__.Notify(581136)
		fallthrough
	case catalog.Type:
		__antithesis_instrumentation__.Notify(581137)
		if scName == "" || func() bool {
			__antithesis_instrumentation__.Notify(581141)
			return dbName == "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(581142)
			return "", errors.AssertionFailedf("schema or database missing for type/relation %d", id)
		} else {
			__antithesis_instrumentation__.Notify(581143)
		}
	case catalog.Schema:
		__antithesis_instrumentation__.Notify(581138)
		if scName != "" || func() bool {
			__antithesis_instrumentation__.Notify(581144)
			return dbName == "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(581145)
			return "", errors.AssertionFailedf("schema or database are invalid for schema %d", id)
		} else {
			__antithesis_instrumentation__.Notify(581146)
		}
	case catalog.Database:
		__antithesis_instrumentation__.Notify(581139)
		if scName != "" || func() bool {
			__antithesis_instrumentation__.Notify(581147)
			return dbName != "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(581148)
			return "", errors.AssertionFailedf("schema or database are set for database %d", id)
		} else {
			__antithesis_instrumentation__.Notify(581149)
		}
	default:
		__antithesis_instrumentation__.Notify(581140)
	}
	__antithesis_instrumentation__.Notify(581121)
	return tree.NewTableNameWithSchema(tree.Name(dbName), tree.Name(scName), tree.Name(obj.GetName())).FQString(), nil
}

func (s *TestState) NewCatalogChangeBatcher() scexec.CatalogChangeBatcher {
	__antithesis_instrumentation__.Notify(581150)
	return &testCatalogChangeBatcher{
		s:             s,
		namesToDelete: make(map[descpb.NameInfo]descpb.ID),
	}
}

type testCatalogChangeBatcher struct {
	s                   *TestState
	descs               []catalog.Descriptor
	namesToDelete       map[descpb.NameInfo]descpb.ID
	descriptorsToDelete catalog.DescriptorIDSet
	zoneConfigsToDelete catalog.DescriptorIDSet
}

var _ scexec.CatalogChangeBatcher = (*testCatalogChangeBatcher)(nil)

func (b *testCatalogChangeBatcher) CreateOrUpdateDescriptor(
	ctx context.Context, desc catalog.MutableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(581151)
	b.descs = append(b.descs, desc)
	return nil
}

func (b *testCatalogChangeBatcher) DeleteName(
	ctx context.Context, nameInfo descpb.NameInfo, id descpb.ID,
) error {
	__antithesis_instrumentation__.Notify(581152)
	b.namesToDelete[nameInfo] = id
	return nil
}

func (b *testCatalogChangeBatcher) DeleteDescriptor(ctx context.Context, id descpb.ID) error {
	__antithesis_instrumentation__.Notify(581153)
	b.descriptorsToDelete.Add(id)
	return nil
}

func (b *testCatalogChangeBatcher) DeleteZoneConfig(ctx context.Context, id descpb.ID) error {
	__antithesis_instrumentation__.Notify(581154)
	b.zoneConfigsToDelete.Add(id)
	return nil
}

func (b *testCatalogChangeBatcher) ValidateAndRun(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(581155)
	names := make([]descpb.NameInfo, 0, len(b.namesToDelete))
	for nameInfo := range b.namesToDelete {
		__antithesis_instrumentation__.Notify(581162)
		names = append(names, nameInfo)
	}
	__antithesis_instrumentation__.Notify(581156)
	sort.Slice(names, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(581163)
		return b.namesToDelete[names[i]] < b.namesToDelete[names[j]]
	})
	__antithesis_instrumentation__.Notify(581157)
	for _, nameInfo := range names {
		__antithesis_instrumentation__.Notify(581164)
		expectedID := b.namesToDelete[nameInfo]
		ne := b.s.catalog.LookupNamespaceEntry(nameInfo)
		if ne == nil {
			__antithesis_instrumentation__.Notify(581168)
			return errors.AssertionFailedf(
				"cannot delete missing namespace entry %v", nameInfo)
		} else {
			__antithesis_instrumentation__.Notify(581169)
		}
		__antithesis_instrumentation__.Notify(581165)

		if actualID := ne.GetID(); actualID != expectedID {
			__antithesis_instrumentation__.Notify(581170)
			return errors.AssertionFailedf(
				"expected deleted namespace entry %v to have ID %d, instead is %d", nameInfo, expectedID, actualID)
		} else {
			__antithesis_instrumentation__.Notify(581171)
		}
		__antithesis_instrumentation__.Notify(581166)
		nameType := "object"
		if nameInfo.ParentSchemaID == 0 {
			__antithesis_instrumentation__.Notify(581172)
			if nameInfo.ParentID == 0 {
				__antithesis_instrumentation__.Notify(581173)
				nameType = "database"
			} else {
				__antithesis_instrumentation__.Notify(581174)
				nameType = "schema"
			}
		} else {
			__antithesis_instrumentation__.Notify(581175)
		}
		__antithesis_instrumentation__.Notify(581167)
		b.s.LogSideEffectf("delete %s namespace entry %v -> %d", nameType, nameInfo, expectedID)
		b.s.catalog.DeleteNamespaceEntry(nameInfo)
	}
	__antithesis_instrumentation__.Notify(581158)
	for _, desc := range b.descs {
		__antithesis_instrumentation__.Notify(581176)
		var old protoutil.Message
		if b := b.s.descBuilder(desc.GetID()); b != nil {
			__antithesis_instrumentation__.Notify(581178)
			old = b.BuildImmutable().DescriptorProto()
		} else {
			__antithesis_instrumentation__.Notify(581179)
		}
		__antithesis_instrumentation__.Notify(581177)
		diff := sctestutils.ProtoDiff(old, desc.DescriptorProto(), sctestutils.DiffArgs{
			Indent:       "  ",
			CompactLevel: 3,
		})
		b.s.LogSideEffectf("upsert descriptor #%d\n%s", desc.GetID(), diff)
		b.s.catalog.UpsertDescriptorEntry(desc)
	}
	__antithesis_instrumentation__.Notify(581159)
	for _, deletedID := range b.descriptorsToDelete.Ordered() {
		__antithesis_instrumentation__.Notify(581180)
		b.s.LogSideEffectf("delete descriptor #%d", deletedID)
		b.s.catalog.DeleteDescriptorEntry(deletedID)
	}
	__antithesis_instrumentation__.Notify(581160)
	for _, deletedID := range b.zoneConfigsToDelete.Ordered() {
		__antithesis_instrumentation__.Notify(581181)
		b.s.LogSideEffectf("deleting zone config for #%d", deletedID)
	}
	__antithesis_instrumentation__.Notify(581161)
	ve := b.s.catalog.Validate(ctx, clusterversion.TestingClusterVersion, catalog.NoValidationTelemetry, catalog.ValidationLevelAllPreTxnCommit, b.descs...)
	return ve.CombinedError()
}

func (s *TestState) IndexSpanSplitter() scexec.IndexSpanSplitter {
	__antithesis_instrumentation__.Notify(581182)
	return s.indexSpanSplitter
}

func (s *TestState) IndexBackfiller() scexec.Backfiller {
	__antithesis_instrumentation__.Notify(581183)
	return s.backfiller
}

func (s *TestState) PeriodicProgressFlusher() scexec.PeriodicProgressFlusher {
	__antithesis_instrumentation__.Notify(581184)
	return scdeps.NewNoopPeriodicProgressFlusher()
}

func (s *TestState) TransactionalJobRegistry() scexec.TransactionalJobRegistry {
	__antithesis_instrumentation__.Notify(581185)
	return s
}

var _ scexec.TransactionalJobRegistry = (*TestState)(nil)

func (s *TestState) CreateJob(ctx context.Context, record jobs.Record) error {
	__antithesis_instrumentation__.Notify(581186)
	if record.JobID == 0 {
		__antithesis_instrumentation__.Notify(581188)
		return errors.New("invalid 0 job ID")
	} else {
		__antithesis_instrumentation__.Notify(581189)
	}
	__antithesis_instrumentation__.Notify(581187)
	record.JobID = jobspb.JobID(1 + len(s.jobs))
	s.createdJobsInCurrentTxn = append(s.createdJobsInCurrentTxn, record.JobID)
	s.jobs = append(s.jobs, record)
	s.LogSideEffectf("create job #%d (non-cancelable: %v): %q\n  descriptor IDs: %v",
		record.JobID,
		record.NonCancelable,
		record.Description,
		record.DescriptorIDs,
	)
	return nil
}

func (s *TestState) CreatedJobs() []jobspb.JobID {
	__antithesis_instrumentation__.Notify(581190)
	return s.createdJobsInCurrentTxn
}

func (s *TestState) CheckPausepoint(name string) error {
	__antithesis_instrumentation__.Notify(581191)
	return nil
}

func (s *TestState) UpdateSchemaChangeJob(
	ctx context.Context, id jobspb.JobID, fn scexec.JobUpdateCallback,
) error {
	__antithesis_instrumentation__.Notify(581192)
	var scJob *jobs.Record
	for i, job := range s.jobs {
		__antithesis_instrumentation__.Notify(581197)
		if job.JobID == id {
			__antithesis_instrumentation__.Notify(581198)
			scJob = &s.jobs[i]
		} else {
			__antithesis_instrumentation__.Notify(581199)
		}
	}
	__antithesis_instrumentation__.Notify(581193)
	if scJob == nil {
		__antithesis_instrumentation__.Notify(581200)
		return errors.AssertionFailedf("schema change job not found")
	} else {
		__antithesis_instrumentation__.Notify(581201)
	}
	__antithesis_instrumentation__.Notify(581194)
	progress := jobspb.Progress{
		Progress:       nil,
		ModifiedMicros: 0,
		RunningStatus:  "",
		Details:        jobspb.WrapProgressDetails(scJob.Progress),
		TraceID:        0,
	}
	payload := jobspb.Payload{
		Description:                  scJob.Description,
		Statement:                    scJob.Statements,
		UsernameProto:                scJob.Username.EncodeProto(),
		StartedMicros:                0,
		FinishedMicros:               0,
		DescriptorIDs:                scJob.DescriptorIDs,
		Error:                        "",
		ResumeErrors:                 nil,
		CleanupErrors:                nil,
		FinalResumeError:             nil,
		Noncancelable:                scJob.NonCancelable,
		Details:                      jobspb.WrapPayloadDetails(scJob.Details),
		PauseReason:                  "",
		RetriableExecutionFailureLog: nil,
	}
	updateProgress := func(progress *jobspb.Progress) {
		__antithesis_instrumentation__.Notify(581202)
		scJob.Progress = *progress.GetNewSchemaChange()
		s.LogSideEffectf("update progress of schema change job #%d: %q", scJob.JobID, progress.RunningStatus)
	}
	__antithesis_instrumentation__.Notify(581195)
	setNonCancelable := func() {
		__antithesis_instrumentation__.Notify(581203)
		scJob.NonCancelable = true
		s.LogSideEffectf("set schema change job #%d to non-cancellable", scJob.JobID)
	}
	__antithesis_instrumentation__.Notify(581196)
	md := jobs.JobMetadata{
		ID:       scJob.JobID,
		Status:   jobs.StatusRunning,
		Payload:  &payload,
		Progress: &progress,
		RunStats: nil,
	}
	return fn(md, updateProgress, setNonCancelable)
}

func (s *TestState) MakeJobID() jobspb.JobID {
	__antithesis_instrumentation__.Notify(581204)
	if s.jobCounter == 0 {
		__antithesis_instrumentation__.Notify(581206)

		s.jobCounter = 1
	} else {
		__antithesis_instrumentation__.Notify(581207)
	}
	__antithesis_instrumentation__.Notify(581205)
	s.jobCounter++
	return jobspb.JobID(s.jobCounter)
}

func (s *TestState) SchemaChangerJobID() jobspb.JobID {
	__antithesis_instrumentation__.Notify(581208)
	return 1
}

func (s *TestState) TestingKnobs() *scrun.TestingKnobs {
	__antithesis_instrumentation__.Notify(581209)
	return s.testingKnobs
}

func (s *TestState) Phase() scop.Phase {
	__antithesis_instrumentation__.Notify(581210)
	return s.phase
}

func (s *TestState) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(581211)
	return security.RootUserName()
}

var _ scrun.JobRunDependencies = (*TestState)(nil)

func (s *TestState) WithTxnInJob(ctx context.Context, fn scrun.JobTxnFunc) (err error) {
	__antithesis_instrumentation__.Notify(581212)
	s.WithTxn(func(s *TestState) { __antithesis_instrumentation__.Notify(581214); err = fn(ctx, s) })
	__antithesis_instrumentation__.Notify(581213)
	return err
}

func (s *TestState) ValidateForwardIndexes(
	_ context.Context,
	tbl catalog.TableDescriptor,
	indexes []catalog.Index,
	_ sessiondata.InternalExecutorOverride,
) error {
	__antithesis_instrumentation__.Notify(581215)
	ids := make([]descpb.IndexID, len(indexes))
	for i, idx := range indexes {
		__antithesis_instrumentation__.Notify(581217)
		ids[i] = idx.GetID()
	}
	__antithesis_instrumentation__.Notify(581216)
	s.LogSideEffectf("validate forward indexes %v in table #%d", ids, tbl.GetID())
	return nil
}

func (s *TestState) ValidateInvertedIndexes(
	_ context.Context,
	tbl catalog.TableDescriptor,
	indexes []catalog.Index,
	_ sessiondata.InternalExecutorOverride,
) error {
	__antithesis_instrumentation__.Notify(581218)
	ids := make([]descpb.IndexID, len(indexes))
	for i, idx := range indexes {
		__antithesis_instrumentation__.Notify(581220)
		ids[i] = idx.GetID()
	}
	__antithesis_instrumentation__.Notify(581219)
	s.LogSideEffectf("validate inverted indexes %v in table #%d", ids, tbl.GetID())
	return nil
}

func (s *TestState) IndexValidator() scexec.IndexValidator {
	__antithesis_instrumentation__.Notify(581221)
	return s
}

func (s *TestState) LogEvent(
	_ context.Context,
	descID descpb.ID,
	details eventpb.CommonSQLEventDetails,
	event eventpb.EventPayload,
) error {
	__antithesis_instrumentation__.Notify(581222)
	s.LogSideEffectf("write %T to event log for descriptor #%d: %s",
		event, descID, details.Statement)
	return nil
}

func (s *TestState) EventLogger() scexec.EventLogger {
	__antithesis_instrumentation__.Notify(581223)
	return s
}

func (s *TestState) UpsertDescriptorComment(
	id int64, subID int64, commentType keys.CommentType, comment string,
) error {
	__antithesis_instrumentation__.Notify(581224)
	s.LogSideEffectf("upsert %s comment for descriptor #%d of type %s",
		comment, id, commentType)
	return nil
}

func (s *TestState) DeleteAllCommentsForTables(ids catalog.DescriptorIDSet) error {
	__antithesis_instrumentation__.Notify(581225)
	s.LogSideEffectf("delete all comments for table descriptors %v", ids.Ordered())
	return nil
}

func (s *TestState) DeleteDescriptorComment(
	id int64, subID int64, commentType keys.CommentType,
) error {
	__antithesis_instrumentation__.Notify(581226)
	s.LogSideEffectf("delete comment for descriptor #%d of type %s",
		id, commentType)
	return nil
}

func (s *TestState) UpsertConstraintComment(
	tableID descpb.ID, constraintID descpb.ConstraintID, comment string,
) error {
	__antithesis_instrumentation__.Notify(581227)
	s.LogSideEffectf("upsert comment %s for constraint on #%d, constraint id: %d"+
		comment, tableID, constraintID)
	return nil
}

func (s *TestState) DeleteConstraintComment(
	tableID descpb.ID, constraintID descpb.ConstraintID,
) error {
	__antithesis_instrumentation__.Notify(581228)
	s.LogSideEffectf("delete comment for constraint on #%d, constraint id: %d",
		tableID, constraintID)
	return nil
}

func (s *TestState) DeleteDatabaseRoleSettings(_ context.Context, dbID descpb.ID) error {
	__antithesis_instrumentation__.Notify(581229)
	s.LogSideEffectf("delete role settings for database on #%d", dbID)
	return nil
}

func (s *TestState) SwapDescriptorSubComment(
	id int64, oldSubID int64, newSubID int64, commentType keys.CommentType,
) error {
	__antithesis_instrumentation__.Notify(581230)
	s.LogSideEffectf("swapping sub comments on descriptor %d from "+
		"%d to %d of type %s",
		id,
		oldSubID,
		newSubID,
		commentType)
	return nil
}

func (s *TestState) DeleteSchedule(ctx context.Context, id int64) error {
	__antithesis_instrumentation__.Notify(581231)
	s.LogSideEffectf("delete scheduleId: %d", id)
	return nil
}

func (s *TestState) DescriptorMetadataUpdater(
	ctx context.Context,
) scexec.DescriptorMetadataUpdater {
	__antithesis_instrumentation__.Notify(581232)
	return s
}
