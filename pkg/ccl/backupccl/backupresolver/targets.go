package backupresolver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type DescriptorsMatched struct {
	Descs []catalog.Descriptor

	ExpandedDB []descpb.ID

	RequestedDBs []catalog.DatabaseDescriptor
}

type MissingTableErr struct {
	wrapped   error
	tableName string
}

func (e *MissingTableErr) Error() string {
	__antithesis_instrumentation__.Notify(8984)
	return fmt.Sprintf("table %q does not exist, or invalid RESTORE timestamp: %v", e.GetTableName(), e.wrapped.Error())
}

func (e *MissingTableErr) Unwrap() error {
	__antithesis_instrumentation__.Notify(8985)
	return e.wrapped
}

func (e *MissingTableErr) GetTableName() string {
	__antithesis_instrumentation__.Notify(8986)
	return e.tableName
}

func (d DescriptorsMatched) CheckExpansions(coveredDBs []descpb.ID) error {
	__antithesis_instrumentation__.Notify(8987)
	covered := make(map[descpb.ID]bool)
	for _, i := range coveredDBs {
		__antithesis_instrumentation__.Notify(8991)
		covered[i] = true
	}
	__antithesis_instrumentation__.Notify(8988)
	for _, i := range d.RequestedDBs {
		__antithesis_instrumentation__.Notify(8992)
		if !covered[i.GetID()] {
			__antithesis_instrumentation__.Notify(8993)
			return errors.Errorf("cannot RESTORE DATABASE from a backup of individual tables (use SHOW BACKUP to determine available tables)")
		} else {
			__antithesis_instrumentation__.Notify(8994)
		}
	}
	__antithesis_instrumentation__.Notify(8989)
	for _, i := range d.ExpandedDB {
		__antithesis_instrumentation__.Notify(8995)
		if !covered[i] {
			__antithesis_instrumentation__.Notify(8996)
			return errors.Errorf("cannot RESTORE <database>.* from a backup of individual tables (use SHOW BACKUP to determine available tables)")
		} else {
			__antithesis_instrumentation__.Notify(8997)
		}
	}
	__antithesis_instrumentation__.Notify(8990)
	return nil
}

type DescriptorResolver struct {
	DescByID map[descpb.ID]catalog.Descriptor

	DbsByName map[string]descpb.ID

	SchemasByName map[descpb.ID]map[string]descpb.ID

	ObjsByName map[descpb.ID]map[string]map[string]descpb.ID
}

func (r *DescriptorResolver) LookupSchema(
	ctx context.Context, dbName, scName string,
) (bool, catalog.ResolvedObjectPrefix, error) {
	__antithesis_instrumentation__.Notify(8998)
	dbID, ok := r.DbsByName[dbName]
	if !ok {
		__antithesis_instrumentation__.Notify(9001)
		return false, catalog.ResolvedObjectPrefix{}, nil
	} else {
		__antithesis_instrumentation__.Notify(9002)
	}
	__antithesis_instrumentation__.Notify(8999)
	schemas := r.SchemasByName[dbID]
	if scID, ok := schemas[scName]; ok {
		__antithesis_instrumentation__.Notify(9003)
		dbDesc, dbOk := r.DescByID[dbID].(catalog.DatabaseDescriptor)
		scDesc, scOk := r.DescByID[scID].(catalog.SchemaDescriptor)

		if !scOk && func() bool {
			__antithesis_instrumentation__.Notify(9005)
			return scID == keys.SystemPublicSchemaID == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(9006)
			return (!scOk && func() bool {
				__antithesis_instrumentation__.Notify(9007)
				return scID == keys.PublicSchemaIDForBackup == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(9008)
			scDesc, scOk = schemadesc.GetPublicSchema(), true
		} else {
			__antithesis_instrumentation__.Notify(9009)
		}
		__antithesis_instrumentation__.Notify(9004)
		if dbOk && func() bool {
			__antithesis_instrumentation__.Notify(9010)
			return scOk == true
		}() == true {
			__antithesis_instrumentation__.Notify(9011)
			return true, catalog.ResolvedObjectPrefix{
				Database: dbDesc,
				Schema:   scDesc,
			}, nil
		} else {
			__antithesis_instrumentation__.Notify(9012)
		}
	} else {
		__antithesis_instrumentation__.Notify(9013)
	}
	__antithesis_instrumentation__.Notify(9000)
	return false, catalog.ResolvedObjectPrefix{}, nil
}

func (r *DescriptorResolver) LookupObject(
	ctx context.Context, flags tree.ObjectLookupFlags, dbName, scName, obName string,
) (bool, catalog.ResolvedObjectPrefix, catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(9014)

	resolvedPrefix := catalog.ResolvedObjectPrefix{}
	if flags.RequireMutable {
		__antithesis_instrumentation__.Notify(9020)
		panic("did not expect request for mutable descriptor")
	} else {
		__antithesis_instrumentation__.Notify(9021)
	}
	__antithesis_instrumentation__.Notify(9015)
	dbID, ok := r.DbsByName[dbName]
	if !ok {
		__antithesis_instrumentation__.Notify(9022)
		return false, resolvedPrefix, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(9023)
	}
	__antithesis_instrumentation__.Notify(9016)
	resolvedPrefix.Database, ok = r.DescByID[dbID].(catalog.DatabaseDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(9024)
		return false, resolvedPrefix, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(9025)
	}
	__antithesis_instrumentation__.Notify(9017)
	scID, ok := r.SchemasByName[dbID][scName]
	if !ok {
		__antithesis_instrumentation__.Notify(9026)
		return false, resolvedPrefix, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(9027)
	}
	__antithesis_instrumentation__.Notify(9018)
	if scMap, ok := r.ObjsByName[dbID]; ok {
		__antithesis_instrumentation__.Notify(9028)
		if objMap, ok := scMap[scName]; ok {
			__antithesis_instrumentation__.Notify(9029)
			if objID, ok := objMap[obName]; ok {
				__antithesis_instrumentation__.Notify(9030)
				if scID == keys.PublicSchemaID {
					__antithesis_instrumentation__.Notify(9032)
					resolvedPrefix.Schema = schemadesc.GetPublicSchema()
				} else {
					__antithesis_instrumentation__.Notify(9033)
					resolvedPrefix.Schema, ok = r.DescByID[scID].(catalog.SchemaDescriptor)
					if !ok {
						__antithesis_instrumentation__.Notify(9034)
						return false, resolvedPrefix, nil, errors.AssertionFailedf(
							"expected schema for ID %d, got %T", scID, r.DescByID[scID])
					} else {
						__antithesis_instrumentation__.Notify(9035)
					}
				}
				__antithesis_instrumentation__.Notify(9031)
				return true, resolvedPrefix, r.DescByID[objID], nil
			} else {
				__antithesis_instrumentation__.Notify(9036)
			}
		} else {
			__antithesis_instrumentation__.Notify(9037)
		}
	} else {
		__antithesis_instrumentation__.Notify(9038)
	}
	__antithesis_instrumentation__.Notify(9019)
	return false, catalog.ResolvedObjectPrefix{}, nil, nil
}

func NewDescriptorResolver(descs []catalog.Descriptor) (*DescriptorResolver, error) {
	__antithesis_instrumentation__.Notify(9039)
	r := &DescriptorResolver{
		DescByID:      make(map[descpb.ID]catalog.Descriptor),
		SchemasByName: make(map[descpb.ID]map[string]descpb.ID),
		DbsByName:     make(map[string]descpb.ID),
		ObjsByName:    make(map[descpb.ID]map[string]map[string]descpb.ID),
	}

	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(9044)
		if desc.Dropped() {
			__antithesis_instrumentation__.Notify(9048)
			continue
		} else {
			__antithesis_instrumentation__.Notify(9049)
		}
		__antithesis_instrumentation__.Notify(9045)
		if _, isDB := desc.(catalog.DatabaseDescriptor); isDB {
			__antithesis_instrumentation__.Notify(9050)
			if _, ok := r.DbsByName[desc.GetName()]; ok {
				__antithesis_instrumentation__.Notify(9052)
				return nil, errors.Errorf("duplicate database name: %q used for ID %d and %d",
					desc.GetName(), r.DbsByName[desc.GetName()], desc.GetID())
			} else {
				__antithesis_instrumentation__.Notify(9053)
			}
			__antithesis_instrumentation__.Notify(9051)
			r.DbsByName[desc.GetName()] = desc.GetID()
			r.ObjsByName[desc.GetID()] = make(map[string]map[string]descpb.ID)
			r.SchemasByName[desc.GetID()] = make(map[string]descpb.ID)

			if !desc.(catalog.DatabaseDescriptor).HasPublicSchemaWithDescriptor() {
				__antithesis_instrumentation__.Notify(9054)
				r.ObjsByName[desc.GetID()][tree.PublicSchema] = make(map[string]descpb.ID)
				r.SchemasByName[desc.GetID()][tree.PublicSchema] = keys.PublicSchemaIDForBackup
			} else {
				__antithesis_instrumentation__.Notify(9055)
			}
		} else {
			__antithesis_instrumentation__.Notify(9056)
		}
		__antithesis_instrumentation__.Notify(9046)

		if prevDesc, ok := r.DescByID[desc.GetID()]; ok {
			__antithesis_instrumentation__.Notify(9057)
			return nil, errors.Errorf("duplicate descriptor ID: %d used by %q and %q",
				desc.GetID(), prevDesc.GetName(), desc.GetName())
		} else {
			__antithesis_instrumentation__.Notify(9058)
		}
		__antithesis_instrumentation__.Notify(9047)
		r.DescByID[desc.GetID()] = desc
	}
	__antithesis_instrumentation__.Notify(9040)

	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(9059)
		if desc.Dropped() {
			__antithesis_instrumentation__.Notify(9061)
			continue
		} else {
			__antithesis_instrumentation__.Notify(9062)
		}
		__antithesis_instrumentation__.Notify(9060)
		if sc, ok := desc.(catalog.SchemaDescriptor); ok {
			__antithesis_instrumentation__.Notify(9063)
			schemaMap := r.ObjsByName[sc.GetParentID()]
			if schemaMap == nil {
				__antithesis_instrumentation__.Notify(9066)
				schemaMap = make(map[string]map[string]descpb.ID)
			} else {
				__antithesis_instrumentation__.Notify(9067)
			}
			__antithesis_instrumentation__.Notify(9064)
			schemaMap[sc.GetName()] = make(map[string]descpb.ID)
			r.ObjsByName[sc.GetParentID()] = schemaMap

			schemaNameMap := r.SchemasByName[sc.GetParentID()]
			if schemaNameMap == nil {
				__antithesis_instrumentation__.Notify(9068)
				schemaNameMap = make(map[string]descpb.ID)
			} else {
				__antithesis_instrumentation__.Notify(9069)
			}
			__antithesis_instrumentation__.Notify(9065)
			schemaNameMap[sc.GetName()] = sc.GetID()
			r.SchemasByName[sc.GetParentID()] = schemaNameMap
		} else {
			__antithesis_instrumentation__.Notify(9070)
		}
	}
	__antithesis_instrumentation__.Notify(9041)

	registerDesc := func(parentID descpb.ID, desc catalog.Descriptor, kind string) error {
		__antithesis_instrumentation__.Notify(9071)
		parentDesc, ok := r.DescByID[parentID]
		if !ok {
			__antithesis_instrumentation__.Notify(9077)
			return errors.Errorf("%s %q has unknown ParentID %d", kind, desc.GetName(), parentID)
		} else {
			__antithesis_instrumentation__.Notify(9078)
		}
		__antithesis_instrumentation__.Notify(9072)
		if _, ok := r.DbsByName[parentDesc.GetName()]; !ok {
			__antithesis_instrumentation__.Notify(9079)
			return errors.Errorf("%s %q's ParentID %d (%q) is not a database",
				kind, desc.GetName(), parentID, parentDesc.GetName())
		} else {
			__antithesis_instrumentation__.Notify(9080)
		}
		__antithesis_instrumentation__.Notify(9073)

		schemaMap := r.ObjsByName[parentDesc.GetID()]
		scID := desc.GetParentSchemaID()
		var scName string

		if scID == keys.PublicSchemaIDForBackup {
			__antithesis_instrumentation__.Notify(9081)
			scName = tree.PublicSchema
		} else {
			__antithesis_instrumentation__.Notify(9082)
			scDescI, ok := r.DescByID[scID]
			if !ok {
				__antithesis_instrumentation__.Notify(9085)
				return errors.Errorf("schema %d not found for desc %d", scID, desc.GetID())
			} else {
				__antithesis_instrumentation__.Notify(9086)
			}
			__antithesis_instrumentation__.Notify(9083)
			scDesc, err := catalog.AsSchemaDescriptor(scDescI)
			if err != nil {
				__antithesis_instrumentation__.Notify(9087)
				return err
			} else {
				__antithesis_instrumentation__.Notify(9088)
			}
			__antithesis_instrumentation__.Notify(9084)
			scName = scDesc.GetName()
		}
		__antithesis_instrumentation__.Notify(9074)

		objMap := schemaMap[scName]
		if objMap == nil {
			__antithesis_instrumentation__.Notify(9089)
			objMap = make(map[string]descpb.ID)
		} else {
			__antithesis_instrumentation__.Notify(9090)
		}
		__antithesis_instrumentation__.Notify(9075)
		if _, ok := objMap[desc.GetName()]; ok {
			__antithesis_instrumentation__.Notify(9091)
			return errors.Errorf("duplicate %s name: %q.%q.%q used for ID %d and %d",
				kind, parentDesc.GetName(), scName, desc.GetName(), desc.GetID(), objMap[desc.GetName()])
		} else {
			__antithesis_instrumentation__.Notify(9092)
		}
		__antithesis_instrumentation__.Notify(9076)
		objMap[desc.GetName()] = desc.GetID()
		r.ObjsByName[parentDesc.GetID()][scName] = objMap
		return nil
	}
	__antithesis_instrumentation__.Notify(9042)

	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(9093)
		if desc.Dropped() {
			__antithesis_instrumentation__.Notify(9096)
			continue
		} else {
			__antithesis_instrumentation__.Notify(9097)
		}
		__antithesis_instrumentation__.Notify(9094)
		var typeToRegister string
		switch desc := desc.(type) {
		case catalog.TableDescriptor:
			__antithesis_instrumentation__.Notify(9098)
			if desc.IsTemporary() {
				__antithesis_instrumentation__.Notify(9101)
				continue
			} else {
				__antithesis_instrumentation__.Notify(9102)
			}
			__antithesis_instrumentation__.Notify(9099)
			typeToRegister = "table"
		case catalog.TypeDescriptor:
			__antithesis_instrumentation__.Notify(9100)
			typeToRegister = "type"
		}
		__antithesis_instrumentation__.Notify(9095)
		if typeToRegister != "" {
			__antithesis_instrumentation__.Notify(9103)
			if err := registerDesc(desc.GetParentID(), desc, typeToRegister); err != nil {
				__antithesis_instrumentation__.Notify(9104)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(9105)
			}
		} else {
			__antithesis_instrumentation__.Notify(9106)
		}
	}
	__antithesis_instrumentation__.Notify(9043)

	return r, nil
}

func DescriptorsMatchingTargets(
	ctx context.Context,
	currentDatabase string,
	searchPath sessiondata.SearchPath,
	descriptors []catalog.Descriptor,
	targets tree.TargetList,
	asOf hlc.Timestamp,
) (DescriptorsMatched, error) {
	__antithesis_instrumentation__.Notify(9107)
	ret := DescriptorsMatched{}

	r, err := NewDescriptorResolver(descriptors)
	if err != nil {
		__antithesis_instrumentation__.Notify(9117)
		return ret, err
	} else {
		__antithesis_instrumentation__.Notify(9118)
	}
	__antithesis_instrumentation__.Notify(9108)

	alreadyRequestedDBs := make(map[descpb.ID]struct{})
	alreadyExpandedDBs := make(map[descpb.ID]struct{})
	invalidRestoreTsErr := errors.Errorf("supplied backups do not cover requested time")

	for _, d := range targets.Databases {
		__antithesis_instrumentation__.Notify(9119)
		dbID, ok := r.DbsByName[string(d)]
		if !ok {
			__antithesis_instrumentation__.Notify(9121)
			if asOf.IsEmpty() {
				__antithesis_instrumentation__.Notify(9123)
				return ret, errors.Errorf("database %q does not exist", d)
			} else {
				__antithesis_instrumentation__.Notify(9124)
			}
			__antithesis_instrumentation__.Notify(9122)
			return ret, errors.Wrapf(invalidRestoreTsErr, "database %q does not exist, or invalid RESTORE timestamp", d)
		} else {
			__antithesis_instrumentation__.Notify(9125)
		}
		__antithesis_instrumentation__.Notify(9120)
		if _, ok := alreadyRequestedDBs[dbID]; !ok {
			__antithesis_instrumentation__.Notify(9126)
			desc := r.DescByID[dbID]

			doesNotExistErr := errors.Errorf(`database %q does not exist`, d)
			if err := catalog.FilterDescriptorState(
				desc, tree.CommonLookupFlags{},
			); err != nil {
				__antithesis_instrumentation__.Notify(9128)

				return ret, doesNotExistErr
			} else {
				__antithesis_instrumentation__.Notify(9129)
			}
			__antithesis_instrumentation__.Notify(9127)
			ret.Descs = append(ret.Descs, desc)
			ret.RequestedDBs = append(ret.RequestedDBs,
				desc.(catalog.DatabaseDescriptor))
			ret.ExpandedDB = append(ret.ExpandedDB, dbID)
			alreadyRequestedDBs[dbID] = struct{}{}
			alreadyExpandedDBs[dbID] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(9130)
		}
	}
	__antithesis_instrumentation__.Notify(9109)

	alreadyRequestedSchemas := make(map[descpb.ID]struct{})
	maybeAddSchemaDesc := func(id descpb.ID, requirePublic bool) error {
		__antithesis_instrumentation__.Notify(9131)

		if id == keys.PublicSchemaIDForBackup {
			__antithesis_instrumentation__.Notify(9134)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(9135)
		}
		__antithesis_instrumentation__.Notify(9132)
		if _, ok := alreadyRequestedSchemas[id]; !ok {
			__antithesis_instrumentation__.Notify(9136)
			schemaDesc := r.DescByID[id]
			if err := catalog.FilterDescriptorState(
				schemaDesc, tree.CommonLookupFlags{},
			); err != nil {
				__antithesis_instrumentation__.Notify(9138)
				if requirePublic {
					__antithesis_instrumentation__.Notify(9140)
					return errors.Wrapf(err, "schema %d was expected to be PUBLIC", id)
				} else {
					__antithesis_instrumentation__.Notify(9141)
				}
				__antithesis_instrumentation__.Notify(9139)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(9142)
			}
			__antithesis_instrumentation__.Notify(9137)
			alreadyRequestedSchemas[id] = struct{}{}
			ret.Descs = append(ret.Descs, r.DescByID[id])
		} else {
			__antithesis_instrumentation__.Notify(9143)
		}
		__antithesis_instrumentation__.Notify(9133)

		return nil
	}
	__antithesis_instrumentation__.Notify(9110)
	getSchemaIDByName := func(scName string, dbID descpb.ID) (descpb.ID, error) {
		__antithesis_instrumentation__.Notify(9144)
		schemas, ok := r.SchemasByName[dbID]
		if !ok {
			__antithesis_instrumentation__.Notify(9147)
			return 0, errors.Newf("database with ID %d not found", dbID)
		} else {
			__antithesis_instrumentation__.Notify(9148)
		}
		__antithesis_instrumentation__.Notify(9145)
		schemaID, ok := schemas[scName]
		if !ok {
			__antithesis_instrumentation__.Notify(9149)
			return 0, errors.Newf("schema with name %s not found in DB %d", scName, dbID)
		} else {
			__antithesis_instrumentation__.Notify(9150)
		}
		__antithesis_instrumentation__.Notify(9146)
		return schemaID, nil
	}
	__antithesis_instrumentation__.Notify(9111)

	alreadyRequestedTypes := make(map[descpb.ID]struct{})
	maybeAddTypeDesc := func(id descpb.ID) {
		__antithesis_instrumentation__.Notify(9151)
		if _, ok := alreadyRequestedTypes[id]; !ok {
			__antithesis_instrumentation__.Notify(9152)

			alreadyRequestedTypes[id] = struct{}{}
			ret.Descs = append(ret.Descs, r.DescByID[id])
		} else {
			__antithesis_instrumentation__.Notify(9153)
		}
	}
	__antithesis_instrumentation__.Notify(9112)
	getTypeByID := func(id descpb.ID) (catalog.TypeDescriptor, error) {
		__antithesis_instrumentation__.Notify(9154)
		desc, ok := r.DescByID[id]
		if !ok {
			__antithesis_instrumentation__.Notify(9157)
			return nil, errors.Newf("type with ID %d not found", id)
		} else {
			__antithesis_instrumentation__.Notify(9158)
		}
		__antithesis_instrumentation__.Notify(9155)
		typeDesc, ok := desc.(catalog.TypeDescriptor)
		if !ok {
			__antithesis_instrumentation__.Notify(9159)
			return nil, errors.Newf("descriptor %d is not a type, but a %T", id, desc)
		} else {
			__antithesis_instrumentation__.Notify(9160)
		}
		__antithesis_instrumentation__.Notify(9156)
		return typeDesc, nil
	}
	__antithesis_instrumentation__.Notify(9113)

	alreadyRequestedTables := make(map[descpb.ID]struct{})

	alreadyRequestedSchemasByDBs := make(map[descpb.ID]map[string]struct{})
	for _, pattern := range targets.Tables {
		__antithesis_instrumentation__.Notify(9161)
		var err error
		pattern, err = pattern.NormalizeTablePattern()
		if err != nil {
			__antithesis_instrumentation__.Notify(9163)
			return ret, err
		} else {
			__antithesis_instrumentation__.Notify(9164)
		}
		__antithesis_instrumentation__.Notify(9162)

		switch p := pattern.(type) {
		case *tree.TableName:
			__antithesis_instrumentation__.Notify(9165)
			un := p.ToUnresolvedObjectName()
			found, prefix, descI, err := resolver.ResolveExisting(ctx, un, r, tree.ObjectLookupFlags{}, currentDatabase, searchPath)
			if err != nil {
				__antithesis_instrumentation__.Notify(9182)
				return ret, err
			} else {
				__antithesis_instrumentation__.Notify(9183)
			}
			__antithesis_instrumentation__.Notify(9166)

			if (prefix.ExplicitDatabase && func() bool {
				__antithesis_instrumentation__.Notify(9184)
				return !p.ExplicitCatalog == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(9185)
				return (prefix.ExplicitSchema && func() bool {
					__antithesis_instrumentation__.Notify(9186)
					return !p.ExplicitSchema == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(9187)
				p.ObjectNamePrefix = prefix.NamePrefix()
			} else {
				__antithesis_instrumentation__.Notify(9188)
			}
			__antithesis_instrumentation__.Notify(9167)
			doesNotExistErr := errors.Errorf(`table %q does not exist`, tree.ErrString(p))
			if !found {
				__antithesis_instrumentation__.Notify(9189)
				if asOf.IsEmpty() {
					__antithesis_instrumentation__.Notify(9191)
					return ret, doesNotExistErr
				} else {
					__antithesis_instrumentation__.Notify(9192)
				}
				__antithesis_instrumentation__.Notify(9190)
				return ret, &MissingTableErr{invalidRestoreTsErr, tree.ErrString(p)}
			} else {
				__antithesis_instrumentation__.Notify(9193)
			}
			__antithesis_instrumentation__.Notify(9168)
			tableDesc, isTable := descI.(catalog.TableDescriptor)

			if !isTable {
				__antithesis_instrumentation__.Notify(9194)
				return ret, doesNotExistErr
			} else {
				__antithesis_instrumentation__.Notify(9195)
			}
			__antithesis_instrumentation__.Notify(9169)

			if err := catalog.FilterDescriptorState(
				tableDesc, tree.CommonLookupFlags{},
			); err != nil {
				__antithesis_instrumentation__.Notify(9196)

				return ret, doesNotExistErr
			} else {
				__antithesis_instrumentation__.Notify(9197)
			}
			__antithesis_instrumentation__.Notify(9170)

			parentID := tableDesc.GetParentID()
			if _, ok := alreadyRequestedDBs[parentID]; !ok {
				__antithesis_instrumentation__.Notify(9198)
				parentDesc := r.DescByID[parentID]
				ret.Descs = append(ret.Descs, parentDesc)
				alreadyRequestedDBs[parentID] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(9199)
			}
			__antithesis_instrumentation__.Notify(9171)

			if _, ok := alreadyRequestedTables[tableDesc.GetID()]; !ok {
				__antithesis_instrumentation__.Notify(9200)
				alreadyRequestedTables[tableDesc.GetID()] = struct{}{}
				ret.Descs = append(ret.Descs, tableDesc)
			} else {
				__antithesis_instrumentation__.Notify(9201)
			}
			__antithesis_instrumentation__.Notify(9172)

			if err := maybeAddSchemaDesc(tableDesc.GetParentSchemaID(), true); err != nil {
				__antithesis_instrumentation__.Notify(9202)
				return ret, err
			} else {
				__antithesis_instrumentation__.Notify(9203)
			}
			__antithesis_instrumentation__.Notify(9173)

			desc := r.DescByID[tableDesc.GetParentID()]
			dbDesc := desc.(catalog.DatabaseDescriptor)
			typeIDs, _, err := tableDesc.GetAllReferencedTypeIDs(dbDesc, getTypeByID)
			if err != nil {
				__antithesis_instrumentation__.Notify(9204)
				return ret, err
			} else {
				__antithesis_instrumentation__.Notify(9205)
			}
			__antithesis_instrumentation__.Notify(9174)
			for _, id := range typeIDs {
				__antithesis_instrumentation__.Notify(9206)
				maybeAddTypeDesc(id)
			}

		case *tree.AllTablesSelector:
			__antithesis_instrumentation__.Notify(9175)

			hasSchemaScope := p.ExplicitSchema && func() bool {
				__antithesis_instrumentation__.Notify(9207)
				return p.ExplicitCatalog == true
			}() == true
			found, prefix, err := resolver.ResolveObjectNamePrefix(ctx, r, currentDatabase, searchPath, &p.ObjectNamePrefix)
			if err != nil {
				__antithesis_instrumentation__.Notify(9208)
				return ret, err
			} else {
				__antithesis_instrumentation__.Notify(9209)
			}
			__antithesis_instrumentation__.Notify(9176)
			if !found {
				__antithesis_instrumentation__.Notify(9210)
				return ret, sqlerrors.NewInvalidWildcardError(tree.ErrString(p))
			} else {
				__antithesis_instrumentation__.Notify(9211)
			}
			__antithesis_instrumentation__.Notify(9177)

			dbID := prefix.Database.GetID()
			if _, ok := alreadyRequestedDBs[dbID]; !ok {
				__antithesis_instrumentation__.Notify(9212)
				ret.Descs = append(ret.Descs, prefix.Database)
				alreadyRequestedDBs[dbID] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(9213)
			}
			__antithesis_instrumentation__.Notify(9178)

			if _, ok := alreadyExpandedDBs[prefix.Database.GetID()]; !ok {
				__antithesis_instrumentation__.Notify(9214)
				ret.ExpandedDB = append(ret.ExpandedDB, prefix.Database.GetID())
				alreadyExpandedDBs[prefix.Database.GetID()] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(9215)
			}
			__antithesis_instrumentation__.Notify(9179)

			if !hasSchemaScope && func() bool {
				__antithesis_instrumentation__.Notify(9216)
				return !p.ExplicitCatalog == true
			}() == true {
				__antithesis_instrumentation__.Notify(9217)
				hasSchemaScope = true
			} else {
				__antithesis_instrumentation__.Notify(9218)
			}
			__antithesis_instrumentation__.Notify(9180)

			if hasSchemaScope {
				__antithesis_instrumentation__.Notify(9219)
				if _, ok := alreadyRequestedSchemasByDBs[dbID]; !ok {
					__antithesis_instrumentation__.Notify(9221)
					scMap := make(map[string]struct{})
					alreadyRequestedSchemasByDBs[dbID] = scMap
				} else {
					__antithesis_instrumentation__.Notify(9222)
				}
				__antithesis_instrumentation__.Notify(9220)
				scMap := alreadyRequestedSchemasByDBs[dbID]
				scMap[p.Schema()] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(9223)
			}
		default:
			__antithesis_instrumentation__.Notify(9181)
			return ret, errors.Errorf("unknown pattern %T: %+v", pattern, pattern)
		}
	}
	__antithesis_instrumentation__.Notify(9114)

	addTableDescsInSchema := func(schemas map[string]descpb.ID) error {
		__antithesis_instrumentation__.Notify(9224)
		for _, id := range schemas {
			__antithesis_instrumentation__.Notify(9226)
			desc := r.DescByID[id]
			switch desc := desc.(type) {
			case catalog.TableDescriptor:
				__antithesis_instrumentation__.Notify(9227)
				if err := catalog.FilterDescriptorState(
					desc, tree.CommonLookupFlags{},
				); err != nil {
					__antithesis_instrumentation__.Notify(9233)

					continue
				} else {
					__antithesis_instrumentation__.Notify(9234)
				}
				__antithesis_instrumentation__.Notify(9228)
				if _, ok := alreadyRequestedTables[id]; !ok {
					__antithesis_instrumentation__.Notify(9235)
					ret.Descs = append(ret.Descs, desc)
				} else {
					__antithesis_instrumentation__.Notify(9236)
				}
				__antithesis_instrumentation__.Notify(9229)

				if desc.GetParentSchemaID() != keys.PublicSchemaIDForBackup {
					__antithesis_instrumentation__.Notify(9237)

					if err := maybeAddSchemaDesc(desc.GetParentSchemaID(), true); err != nil {
						__antithesis_instrumentation__.Notify(9238)
						return err
					} else {
						__antithesis_instrumentation__.Notify(9239)
					}
				} else {
					__antithesis_instrumentation__.Notify(9240)
				}
				__antithesis_instrumentation__.Notify(9230)

				dbRaw := r.DescByID[desc.GetParentID()]
				dbDesc := dbRaw.(catalog.DatabaseDescriptor)
				typeIDs, _, err := desc.GetAllReferencedTypeIDs(dbDesc, getTypeByID)
				if err != nil {
					__antithesis_instrumentation__.Notify(9241)
					return err
				} else {
					__antithesis_instrumentation__.Notify(9242)
				}
				__antithesis_instrumentation__.Notify(9231)
				for _, id := range typeIDs {
					__antithesis_instrumentation__.Notify(9243)
					maybeAddTypeDesc(id)
				}
			case catalog.TypeDescriptor:
				__antithesis_instrumentation__.Notify(9232)
				maybeAddTypeDesc(desc.GetID())
			}
		}
		__antithesis_instrumentation__.Notify(9225)
		return nil
	}
	__antithesis_instrumentation__.Notify(9115)

	for dbID := range alreadyExpandedDBs {
		__antithesis_instrumentation__.Notify(9244)
		if requestedSchemas, ok := alreadyRequestedSchemasByDBs[dbID]; !ok {
			__antithesis_instrumentation__.Notify(9245)
			for schemaName, schemas := range r.ObjsByName[dbID] {
				__antithesis_instrumentation__.Notify(9246)
				schemaID, err := getSchemaIDByName(schemaName, dbID)
				if err != nil {
					__antithesis_instrumentation__.Notify(9249)
					return ret, err
				} else {
					__antithesis_instrumentation__.Notify(9250)
				}
				__antithesis_instrumentation__.Notify(9247)
				if err := maybeAddSchemaDesc(schemaID, false); err != nil {
					__antithesis_instrumentation__.Notify(9251)
					return ret, err
				} else {
					__antithesis_instrumentation__.Notify(9252)
				}
				__antithesis_instrumentation__.Notify(9248)
				if err := addTableDescsInSchema(schemas); err != nil {
					__antithesis_instrumentation__.Notify(9253)
					return ret, err
				} else {
					__antithesis_instrumentation__.Notify(9254)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(9255)
			for schemaName := range requestedSchemas {
				__antithesis_instrumentation__.Notify(9256)
				schemas := r.ObjsByName[dbID][schemaName]
				if err := addTableDescsInSchema(schemas); err != nil {
					__antithesis_instrumentation__.Notify(9257)
					return ret, err
				} else {
					__antithesis_instrumentation__.Notify(9258)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(9116)

	return ret, nil
}

func LoadAllDescs(
	ctx context.Context, execCfg *sql.ExecutorConfig, asOf hlc.Timestamp,
) (allDescs []catalog.Descriptor, _ error) {
	__antithesis_instrumentation__.Notify(9259)
	if err := sql.DescsTxn(ctx, execCfg, func(ctx context.Context, txn *kv.Txn, col *descs.Collection) error {
		__antithesis_instrumentation__.Notify(9261)
		err := txn.SetFixedTimestamp(ctx, asOf)
		if err != nil {
			__antithesis_instrumentation__.Notify(9263)
			return err
		} else {
			__antithesis_instrumentation__.Notify(9264)
		}
		__antithesis_instrumentation__.Notify(9262)
		all, err := col.GetAllDescriptors(ctx, txn)
		allDescs = all.OrderedDescriptors()
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(9265)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(9266)
	}
	__antithesis_instrumentation__.Notify(9260)
	return allDescs, nil
}

func ResolveTargetsToDescriptors(
	ctx context.Context, p sql.PlanHookState, endTime hlc.Timestamp, targets *tree.TargetList,
) ([]catalog.Descriptor, []descpb.ID, error) {
	__antithesis_instrumentation__.Notify(9267)
	allDescs, err := LoadAllDescs(ctx, p.ExecCfg(), endTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(9271)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(9272)
	}
	__antithesis_instrumentation__.Notify(9268)

	var matched DescriptorsMatched
	if matched, err = DescriptorsMatchingTargets(ctx,
		p.CurrentDatabase(), p.CurrentSearchPath(), allDescs, *targets, endTime); err != nil {
		__antithesis_instrumentation__.Notify(9273)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(9274)
	}
	__antithesis_instrumentation__.Notify(9269)

	sort.Slice(matched.Descs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(9275)
		return matched.Descs[i].GetID() < matched.Descs[j].GetID()
	})
	__antithesis_instrumentation__.Notify(9270)

	return matched.Descs, matched.ExpandedDB, nil
}
