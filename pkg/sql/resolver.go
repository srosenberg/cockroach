package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

var _ resolver.SchemaResolver = &planner{}

func (p *planner) runWithOptions(flags resolveFlags, fn func()) {
	__antithesis_instrumentation__.Notify(566600)
	if flags.skipCache {
		__antithesis_instrumentation__.Notify(566603)
		defer func(prev bool) { __antithesis_instrumentation__.Notify(566605); p.avoidLeasedDescriptors = prev }(p.avoidLeasedDescriptors)
		__antithesis_instrumentation__.Notify(566604)
		p.avoidLeasedDescriptors = true
	} else {
		__antithesis_instrumentation__.Notify(566606)
	}
	__antithesis_instrumentation__.Notify(566601)
	if flags.contextDatabaseID != descpb.InvalidID {
		__antithesis_instrumentation__.Notify(566607)
		defer func(prev descpb.ID) { __antithesis_instrumentation__.Notify(566609); p.contextDatabaseID = prev }(p.contextDatabaseID)
		__antithesis_instrumentation__.Notify(566608)
		p.contextDatabaseID = flags.contextDatabaseID
	} else {
		__antithesis_instrumentation__.Notify(566610)
	}
	__antithesis_instrumentation__.Notify(566602)
	fn()
}

type resolveFlags struct {
	skipCache         bool
	contextDatabaseID descpb.ID
}

func (p *planner) ResolveMutableTableDescriptor(
	ctx context.Context, tn *tree.TableName, required bool, requiredType tree.RequiredTableKind,
) (prefix catalog.ResolvedObjectPrefix, table *tabledesc.Mutable, err error) {
	__antithesis_instrumentation__.Notify(566611)
	prefix, desc, err := resolver.ResolveMutableExistingTableObject(ctx, p, tn, required, requiredType)
	if err != nil {
		__antithesis_instrumentation__.Notify(566614)
		return prefix, nil, err
	} else {
		__antithesis_instrumentation__.Notify(566615)
	}
	__antithesis_instrumentation__.Notify(566612)

	if desc != nil {
		__antithesis_instrumentation__.Notify(566616)
		if err := p.canResolveDescUnderSchema(ctx, prefix.Schema, desc); err != nil {
			__antithesis_instrumentation__.Notify(566617)
			return catalog.ResolvedObjectPrefix{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(566618)
		}
	} else {
		__antithesis_instrumentation__.Notify(566619)
	}
	__antithesis_instrumentation__.Notify(566613)

	return prefix, desc, nil
}

func (p *planner) resolveUncachedTableDescriptor(
	ctx context.Context, tn *tree.TableName, required bool, requiredType tree.RequiredTableKind,
) (table catalog.TableDescriptor, err error) {
	__antithesis_instrumentation__.Notify(566620)
	var prefix catalog.ResolvedObjectPrefix
	var desc catalog.Descriptor
	p.runWithOptions(resolveFlags{skipCache: true}, func() {
		__antithesis_instrumentation__.Notify(566624)
		lookupFlags := tree.ObjectLookupFlags{
			CommonLookupFlags:    tree.CommonLookupFlags{Required: required},
			DesiredObjectKind:    tree.TableObject,
			DesiredTableDescKind: requiredType,
		}
		desc, prefix, err = resolver.ResolveExistingObject(
			ctx, p, tn.ToUnresolvedObjectName(), lookupFlags,
		)
	})
	__antithesis_instrumentation__.Notify(566621)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(566625)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(566626)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566627)
	}
	__antithesis_instrumentation__.Notify(566622)
	table = desc.(catalog.TableDescriptor)

	if err := p.canResolveDescUnderSchema(ctx, prefix.Schema, table); err != nil {
		__antithesis_instrumentation__.Notify(566628)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566629)
	}
	__antithesis_instrumentation__.Notify(566623)

	return table, nil
}

func (p *planner) ResolveTargetObject(
	ctx context.Context, un *tree.UnresolvedObjectName,
) (
	db catalog.DatabaseDescriptor,
	schema catalog.SchemaDescriptor,
	namePrefix tree.ObjectNamePrefix,
	err error,
) {
	__antithesis_instrumentation__.Notify(566630)
	var prefix catalog.ResolvedObjectPrefix
	p.runWithOptions(resolveFlags{skipCache: true}, func() {
		__antithesis_instrumentation__.Notify(566633)
		prefix, namePrefix, err = resolver.ResolveTargetObject(ctx, p, un)
	})
	__antithesis_instrumentation__.Notify(566631)
	if err != nil {
		__antithesis_instrumentation__.Notify(566634)
		return nil, nil, namePrefix, err
	} else {
		__antithesis_instrumentation__.Notify(566635)
	}
	__antithesis_instrumentation__.Notify(566632)
	return prefix.Database, prefix.Schema, namePrefix, err
}

func (p *planner) LookupSchema(
	ctx context.Context, dbName, scName string,
) (found bool, scMeta catalog.ResolvedObjectPrefix, err error) {
	__antithesis_instrumentation__.Notify(566636)
	dbDesc, err := p.Descriptors().GetImmutableDatabaseByName(ctx, p.txn, dbName,
		tree.DatabaseLookupFlags{AvoidLeased: p.avoidLeasedDescriptors})
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(566639)
		return dbDesc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(566640)
		return false, catalog.ResolvedObjectPrefix{}, err
	} else {
		__antithesis_instrumentation__.Notify(566641)
	}
	__antithesis_instrumentation__.Notify(566637)
	sc := p.Accessor()
	var resolvedSchema catalog.SchemaDescriptor
	resolvedSchema, err = sc.GetSchemaByName(
		ctx, p.txn, dbDesc, scName, p.CommonLookupFlags(false),
	)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(566642)
		return resolvedSchema == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(566643)
		return false, catalog.ResolvedObjectPrefix{}, err
	} else {
		__antithesis_instrumentation__.Notify(566644)
	}
	__antithesis_instrumentation__.Notify(566638)
	return true, catalog.ResolvedObjectPrefix{
		Database: dbDesc,
		Schema:   resolvedSchema,
	}, nil
}

func (p *planner) SchemaExists(ctx context.Context, dbName, scName string) (found bool, err error) {
	__antithesis_instrumentation__.Notify(566645)
	found, _, err = p.LookupSchema(ctx, dbName, scName)
	return found, err
}

func (p *planner) LookupObject(
	ctx context.Context, flags tree.ObjectLookupFlags, dbName, scName, obName string,
) (found bool, prefix catalog.ResolvedObjectPrefix, objMeta catalog.Descriptor, err error) {
	__antithesis_instrumentation__.Notify(566646)
	sc := p.Accessor()
	flags.CommonLookupFlags.Required = false
	flags.CommonLookupFlags.AvoidLeased = p.avoidLeasedDescriptors

	if flags.DesiredObjectKind == tree.TypeObject && func() bool {
		__antithesis_instrumentation__.Notify(566648)
		return scName == tree.PublicSchema == true
	}() == true {
		__antithesis_instrumentation__.Notify(566649)
		if alias, ok := types.PublicSchemaAliases[obName]; ok {
			__antithesis_instrumentation__.Notify(566650)
			if flags.RequireMutable {
				__antithesis_instrumentation__.Notify(566655)
				return true, catalog.ResolvedObjectPrefix{}, nil, pgerror.Newf(pgcode.WrongObjectType, "type %q is a built-in type", obName)
			} else {
				__antithesis_instrumentation__.Notify(566656)
			}
			__antithesis_instrumentation__.Notify(566651)

			found, prefix, err = p.LookupSchema(ctx, dbName, scName)
			if err != nil || func() bool {
				__antithesis_instrumentation__.Notify(566657)
				return !found == true
			}() == true {
				__antithesis_instrumentation__.Notify(566658)
				return found, prefix, nil, err
			} else {
				__antithesis_instrumentation__.Notify(566659)
			}
			__antithesis_instrumentation__.Notify(566652)
			dbDesc, err := p.Descriptors().GetImmutableDatabaseByName(ctx, p.txn, dbName,
				tree.DatabaseLookupFlags{AvoidLeased: p.avoidLeasedDescriptors})
			if err != nil {
				__antithesis_instrumentation__.Notify(566660)
				return found, prefix, nil, err
			} else {
				__antithesis_instrumentation__.Notify(566661)
			}
			__antithesis_instrumentation__.Notify(566653)
			if dbDesc.HasPublicSchemaWithDescriptor() {
				__antithesis_instrumentation__.Notify(566662)
				publicSchemaID := dbDesc.GetSchemaID(tree.PublicSchema)
				return true, prefix, typedesc.MakeSimpleAlias(alias, publicSchemaID), nil
			} else {
				__antithesis_instrumentation__.Notify(566663)
			}
			__antithesis_instrumentation__.Notify(566654)
			return true, prefix, typedesc.MakeSimpleAlias(alias, keys.PublicSchemaID), nil
		} else {
			__antithesis_instrumentation__.Notify(566664)
		}
	} else {
		__antithesis_instrumentation__.Notify(566665)
	}
	__antithesis_instrumentation__.Notify(566647)

	prefix, objMeta, err = sc.GetObjectDesc(ctx, p.txn, dbName, scName, obName, flags)
	return objMeta != nil, prefix, objMeta, err
}

func (p *planner) CommonLookupFlags(required bool) tree.CommonLookupFlags {
	__antithesis_instrumentation__.Notify(566666)
	return tree.CommonLookupFlags{
		Required:    required,
		AvoidLeased: p.avoidLeasedDescriptors,
	}
}

func (p *planner) IsTableVisible(
	ctx context.Context, curDB string, searchPath sessiondata.SearchPath, tableID oid.Oid,
) (isVisible, exists bool, err error) {
	__antithesis_instrumentation__.Notify(566667)
	tableDesc, err := p.LookupTableByID(ctx, descpb.ID(tableID))
	if err != nil {
		__antithesis_instrumentation__.Notify(566672)

		if errors.Is(err, catalog.ErrDescriptorNotFound) || func() bool {
			__antithesis_instrumentation__.Notify(566674)
			return errors.Is(err, catalog.ErrDescriptorDropped) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(566675)
			return pgerror.GetPGCode(err) == pgcode.UndefinedTable == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(566676)
			return pgerror.GetPGCode(err) == pgcode.UndefinedObject == true
		}() == true {
			__antithesis_instrumentation__.Notify(566677)
			return false, false, nil
		} else {
			__antithesis_instrumentation__.Notify(566678)
		}
		__antithesis_instrumentation__.Notify(566673)
		return false, false, err
	} else {
		__antithesis_instrumentation__.Notify(566679)
	}
	__antithesis_instrumentation__.Notify(566668)
	schemaID := tableDesc.GetParentSchemaID()
	schemaDesc, err := p.Descriptors().GetImmutableSchemaByID(ctx, p.Txn(), schemaID,
		tree.SchemaLookupFlags{
			Required:    true,
			AvoidLeased: p.avoidLeasedDescriptors})
	if err != nil {
		__antithesis_instrumentation__.Notify(566680)
		return false, false, err
	} else {
		__antithesis_instrumentation__.Notify(566681)
	}
	__antithesis_instrumentation__.Notify(566669)
	if schemaDesc.SchemaKind() != catalog.SchemaVirtual {
		__antithesis_instrumentation__.Notify(566682)
		dbID := tableDesc.GetParentID()
		_, dbDesc, err := p.Descriptors().GetImmutableDatabaseByID(ctx, p.Txn(), dbID,
			tree.DatabaseLookupFlags{
				Required:    true,
				AvoidLeased: p.avoidLeasedDescriptors})
		if err != nil {
			__antithesis_instrumentation__.Notify(566684)
			return false, false, err
		} else {
			__antithesis_instrumentation__.Notify(566685)
		}
		__antithesis_instrumentation__.Notify(566683)
		if dbDesc.GetName() != curDB {
			__antithesis_instrumentation__.Notify(566686)

			return false, false, nil
		} else {
			__antithesis_instrumentation__.Notify(566687)
		}
	} else {
		__antithesis_instrumentation__.Notify(566688)
	}
	__antithesis_instrumentation__.Notify(566670)
	iter := searchPath.Iter()
	for scName, ok := iter.Next(); ok; scName, ok = iter.Next() {
		__antithesis_instrumentation__.Notify(566689)
		if schemaDesc.GetName() == scName {
			__antithesis_instrumentation__.Notify(566690)
			return true, true, nil
		} else {
			__antithesis_instrumentation__.Notify(566691)
		}
	}
	__antithesis_instrumentation__.Notify(566671)
	return false, true, nil
}

func (p *planner) IsTypeVisible(
	ctx context.Context, curDB string, searchPath sessiondata.SearchPath, typeID oid.Oid,
) (isVisible bool, exists bool, err error) {
	__antithesis_instrumentation__.Notify(566692)

	if _, ok := types.OidToType[typeID]; ok {
		__antithesis_instrumentation__.Notify(566699)
		return true, true, nil
	} else {
		__antithesis_instrumentation__.Notify(566700)
	}
	__antithesis_instrumentation__.Notify(566693)

	if !types.IsOIDUserDefinedType(typeID) {
		__antithesis_instrumentation__.Notify(566701)
		return false, false, nil
	} else {
		__antithesis_instrumentation__.Notify(566702)
	}
	__antithesis_instrumentation__.Notify(566694)

	id, err := typedesc.UserDefinedTypeOIDToID(typeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(566703)
		return false, false, err
	} else {
		__antithesis_instrumentation__.Notify(566704)
	}
	__antithesis_instrumentation__.Notify(566695)
	typName, _, err := p.GetTypeDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(566705)

		if errors.Is(err, catalog.ErrDescriptorNotFound) || func() bool {
			__antithesis_instrumentation__.Notify(566707)
			return errors.Is(err, catalog.ErrDescriptorDropped) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(566708)
			return pgerror.GetPGCode(err) == pgcode.UndefinedObject == true
		}() == true {
			__antithesis_instrumentation__.Notify(566709)
			return false, false, nil
		} else {
			__antithesis_instrumentation__.Notify(566710)
		}
		__antithesis_instrumentation__.Notify(566706)
		return false, false, err
	} else {
		__antithesis_instrumentation__.Notify(566711)
	}
	__antithesis_instrumentation__.Notify(566696)
	if typName.CatalogName.String() != curDB {
		__antithesis_instrumentation__.Notify(566712)

		return false, false, nil
	} else {
		__antithesis_instrumentation__.Notify(566713)
	}
	__antithesis_instrumentation__.Notify(566697)
	iter := searchPath.Iter()
	for scName, ok := iter.Next(); ok; scName, ok = iter.Next() {
		__antithesis_instrumentation__.Notify(566714)
		if typName.SchemaName.String() == scName {
			__antithesis_instrumentation__.Notify(566715)
			return true, true, nil
		} else {
			__antithesis_instrumentation__.Notify(566716)
		}
	}
	__antithesis_instrumentation__.Notify(566698)
	return false, true, nil
}

func (p *planner) HasAnyPrivilege(
	ctx context.Context,
	specifier tree.HasPrivilegeSpecifier,
	user security.SQLUsername,
	privs []privilege.Privilege,
) (tree.HasAnyPrivilegeResult, error) {
	__antithesis_instrumentation__.Notify(566717)
	desc, err := p.ResolveDescriptorForPrivilegeSpecifier(
		ctx,
		specifier,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(566721)
		return tree.HasNoPrivilege, err
	} else {
		__antithesis_instrumentation__.Notify(566722)
	}
	__antithesis_instrumentation__.Notify(566718)
	if desc == nil {
		__antithesis_instrumentation__.Notify(566723)
		return tree.ObjectNotFound, nil
	} else {
		__antithesis_instrumentation__.Notify(566724)
	}
	__antithesis_instrumentation__.Notify(566719)

	for _, priv := range privs {
		__antithesis_instrumentation__.Notify(566725)

		if priv.Kind == privilege.RULE {
			__antithesis_instrumentation__.Notify(566729)
			continue
		} else {
			__antithesis_instrumentation__.Notify(566730)
		}
		__antithesis_instrumentation__.Notify(566726)

		if err := p.CheckPrivilegeForUser(ctx, desc, priv.Kind, user); err != nil {
			__antithesis_instrumentation__.Notify(566731)
			if pgerror.GetPGCode(err) == pgcode.InsufficientPrivilege {
				__antithesis_instrumentation__.Notify(566733)
				continue
			} else {
				__antithesis_instrumentation__.Notify(566734)
			}
			__antithesis_instrumentation__.Notify(566732)
			return tree.HasNoPrivilege, err
		} else {
			__antithesis_instrumentation__.Notify(566735)
		}
		__antithesis_instrumentation__.Notify(566727)

		if priv.GrantOption {
			__antithesis_instrumentation__.Notify(566736)
			if !p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.ValidateGrantOption) {
				__antithesis_instrumentation__.Notify(566737)
				if err := p.CheckPrivilegeForUser(ctx, desc, privilege.GRANT, user); err != nil {
					__antithesis_instrumentation__.Notify(566738)
					if pgerror.GetPGCode(err) == pgcode.InsufficientPrivilege {
						__antithesis_instrumentation__.Notify(566740)
						continue
					} else {
						__antithesis_instrumentation__.Notify(566741)
					}
					__antithesis_instrumentation__.Notify(566739)
					return tree.HasNoPrivilege, err
				} else {
					__antithesis_instrumentation__.Notify(566742)
				}
			} else {
				__antithesis_instrumentation__.Notify(566743)
				if err := p.CheckGrantOptionsForUser(ctx, desc, []privilege.Kind{priv.Kind}, user, true); err != nil {
					__antithesis_instrumentation__.Notify(566744)
					if pgerror.GetPGCode(err) == pgcode.WarningPrivilegeNotGranted {
						__antithesis_instrumentation__.Notify(566746)
						continue
					} else {
						__antithesis_instrumentation__.Notify(566747)
					}
					__antithesis_instrumentation__.Notify(566745)
					return tree.HasNoPrivilege, err
				} else {
					__antithesis_instrumentation__.Notify(566748)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(566749)
		}
		__antithesis_instrumentation__.Notify(566728)
		return tree.HasPrivilege, nil
	}
	__antithesis_instrumentation__.Notify(566720)

	return tree.HasNoPrivilege, nil
}

func (p *planner) ResolveDescriptorForPrivilegeSpecifier(
	ctx context.Context, specifier tree.HasPrivilegeSpecifier,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(566750)
	if specifier.DatabaseName != nil {
		__antithesis_instrumentation__.Notify(566752)
		return p.Descriptors().GetImmutableDatabaseByName(
			ctx, p.txn, *specifier.DatabaseName, tree.DatabaseLookupFlags{Required: true},
		)
	} else {
		__antithesis_instrumentation__.Notify(566753)
		if specifier.DatabaseOID != nil {
			__antithesis_instrumentation__.Notify(566754)
			_, database, err := p.Descriptors().GetImmutableDatabaseByID(
				ctx, p.txn, descpb.ID(*specifier.DatabaseOID), tree.DatabaseLookupFlags{},
			)
			return database, err
		} else {
			__antithesis_instrumentation__.Notify(566755)
			if specifier.SchemaName != nil {
				__antithesis_instrumentation__.Notify(566756)
				database, err := p.Descriptors().GetImmutableDatabaseByName(
					ctx, p.txn, *specifier.SchemaDatabaseName, tree.DatabaseLookupFlags{Required: true},
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(566758)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(566759)
				}
				__antithesis_instrumentation__.Notify(566757)
				return p.Descriptors().GetImmutableSchemaByName(
					ctx, p.txn, database, *specifier.SchemaName, tree.SchemaLookupFlags{Required: *specifier.SchemaIsRequired},
				)
			} else {
				__antithesis_instrumentation__.Notify(566760)
				if specifier.TableName != nil || func() bool {
					__antithesis_instrumentation__.Notify(566761)
					return specifier.TableOID != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(566762)
					var table catalog.TableDescriptor
					var err error
					if specifier.TableName != nil {
						__antithesis_instrumentation__.Notify(566766)
						var tn *tree.TableName
						tn, err = parser.ParseQualifiedTableName(*specifier.TableName)
						if err != nil {
							__antithesis_instrumentation__.Notify(566770)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(566771)
						}
						__antithesis_instrumentation__.Notify(566767)
						if _, err = p.ResolveTableName(ctx, tn); err != nil {
							__antithesis_instrumentation__.Notify(566772)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(566773)
						}
						__antithesis_instrumentation__.Notify(566768)

						if p.SessionData().Database != "" && func() bool {
							__antithesis_instrumentation__.Notify(566774)
							return p.SessionData().Database != string(tn.CatalogName) == true
						}() == true {
							__antithesis_instrumentation__.Notify(566775)

							return nil, pgerror.Newf(pgcode.FeatureNotSupported,
								"cross-database references are not implemented: %s", tn)
						} else {
							__antithesis_instrumentation__.Notify(566776)
						}
						__antithesis_instrumentation__.Notify(566769)
						_, table, err = p.Descriptors().GetImmutableTableByName(
							ctx, p.txn, tn, tree.ObjectLookupFlags{},
						)
					} else {
						__antithesis_instrumentation__.Notify(566777)
						table, err = p.Descriptors().GetImmutableTableByID(
							ctx, p.txn, descpb.ID(*specifier.TableOID), tree.ObjectLookupFlags{},
						)

						if err != nil && func() bool {
							__antithesis_instrumentation__.Notify(566778)
							return sqlerrors.IsUndefinedRelationError(err) == true
						}() == true {
							__antithesis_instrumentation__.Notify(566779)

							return nil, nil
						} else {
							__antithesis_instrumentation__.Notify(566780)
						}
					}
					__antithesis_instrumentation__.Notify(566763)
					if err != nil {
						__antithesis_instrumentation__.Notify(566781)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(566782)
					}
					__antithesis_instrumentation__.Notify(566764)
					if *specifier.IsSequence {
						__antithesis_instrumentation__.Notify(566783)

						if !table.IsSequence() {
							__antithesis_instrumentation__.Notify(566784)
							return nil, pgerror.Newf(pgcode.WrongObjectType,
								"\"%s\" is not a sequence", table.GetName())
						} else {
							__antithesis_instrumentation__.Notify(566785)
						}
					} else {
						__antithesis_instrumentation__.Notify(566786)
						if err := validateColumnForHasPrivilegeSpecifier(
							table,
							specifier,
						); err != nil {
							__antithesis_instrumentation__.Notify(566787)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(566788)
						}
					}
					__antithesis_instrumentation__.Notify(566765)
					return table, nil
				} else {
					__antithesis_instrumentation__.Notify(566789)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(566751)
	return nil, errors.AssertionFailedf("invalid HasPrivilegeSpecifier")
}

func validateColumnForHasPrivilegeSpecifier(
	table catalog.TableDescriptor, specifier tree.HasPrivilegeSpecifier,
) error {
	__antithesis_instrumentation__.Notify(566790)
	if specifier.ColumnName != nil {
		__antithesis_instrumentation__.Notify(566793)
		_, err := table.FindColumnWithName(*specifier.ColumnName)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566794)
	}
	__antithesis_instrumentation__.Notify(566791)
	if specifier.ColumnAttNum != nil {
		__antithesis_instrumentation__.Notify(566795)
		for _, col := range table.PublicColumns() {
			__antithesis_instrumentation__.Notify(566797)
			if col.GetPGAttributeNum() == *specifier.ColumnAttNum {
				__antithesis_instrumentation__.Notify(566798)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(566799)
			}
		}
		__antithesis_instrumentation__.Notify(566796)
		return pgerror.Newf(
			pgcode.UndefinedColumn,
			"column %d of relation %s does not exist",
			*specifier.ColumnAttNum,
			tree.Name(table.GetName()),
		)

	} else {
		__antithesis_instrumentation__.Notify(566800)
	}
	__antithesis_instrumentation__.Notify(566792)
	return nil
}

func (p *planner) GetTypeDescriptor(
	ctx context.Context, id descpb.ID,
) (tree.TypeName, catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(566801)
	desc, err := p.Descriptors().GetImmutableTypeByID(ctx, p.txn, id, tree.ObjectLookupFlags{})
	if err != nil {
		__antithesis_instrumentation__.Notify(566805)
		return tree.TypeName{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(566806)
	}
	__antithesis_instrumentation__.Notify(566802)

	_, dbDesc, err := p.Descriptors().GetImmutableDatabaseByID(ctx, p.txn, desc.GetParentID(), p.CommonLookupFlags(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(566807)
		return tree.TypeName{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(566808)
	}
	__antithesis_instrumentation__.Notify(566803)
	sc, err := p.Descriptors().GetImmutableSchemaByID(
		ctx, p.txn, desc.GetParentSchemaID(), tree.SchemaLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(566809)
		return tree.TypeName{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(566810)
	}
	__antithesis_instrumentation__.Notify(566804)
	name := tree.MakeQualifiedTypeName(dbDesc.GetName(), sc.GetName(), desc.GetName())
	return name, desc, nil
}

func (p *planner) ResolveType(
	ctx context.Context, name *tree.UnresolvedObjectName,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(566811)
	lookupFlags := tree.ObjectLookupFlags{
		CommonLookupFlags: tree.CommonLookupFlags{Required: true, RequireMutable: false},
		DesiredObjectKind: tree.TypeObject,
	}
	desc, prefix, err := resolver.ResolveExistingObject(ctx, p, name, lookupFlags)
	if err != nil {
		__antithesis_instrumentation__.Notify(566815)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566816)
	}
	__antithesis_instrumentation__.Notify(566812)

	prefix.ExplicitDatabase = prefix.Database != nil
	prefix.ExplicitSchema = prefix.Schema != nil
	tn := tree.MakeTypeNameWithPrefix(prefix.NamePrefix(), name.Object())
	tdesc := desc.(catalog.TypeDescriptor)

	if p.contextDatabaseID != descpb.InvalidID && func() bool {
		__antithesis_instrumentation__.Notify(566817)
		return tdesc.GetParentID() != descpb.InvalidID == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(566818)
		return tdesc.GetParentID() != p.contextDatabaseID == true
	}() == true {
		__antithesis_instrumentation__.Notify(566819)
		return nil, pgerror.Newf(
			pgcode.FeatureNotSupported, "cross database type references are not supported: %s", tn.String())
	} else {
		__antithesis_instrumentation__.Notify(566820)
	}
	__antithesis_instrumentation__.Notify(566813)

	if err := p.canResolveDescUnderSchema(ctx, prefix.Schema, tdesc); err != nil {
		__antithesis_instrumentation__.Notify(566821)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566822)
	}
	__antithesis_instrumentation__.Notify(566814)

	return tdesc.MakeTypesT(ctx, &tn, p)
}

func (p *planner) ResolveTypeByOID(ctx context.Context, oid oid.Oid) (*types.T, error) {
	__antithesis_instrumentation__.Notify(566823)
	id, err := typedesc.UserDefinedTypeOIDToID(oid)
	if err != nil {
		__antithesis_instrumentation__.Notify(566826)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566827)
	}
	__antithesis_instrumentation__.Notify(566824)
	name, desc, err := p.GetTypeDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(566828)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566829)
	}
	__antithesis_instrumentation__.Notify(566825)
	return desc.MakeTypesT(ctx, &name, p)
}

func (p *planner) ObjectLookupFlags(required, requireMutable bool) tree.ObjectLookupFlags {
	__antithesis_instrumentation__.Notify(566830)
	flags := p.CommonLookupFlags(required)
	flags.RequireMutable = requireMutable
	return tree.ObjectLookupFlags{CommonLookupFlags: flags}
}

func getDescriptorsFromTargetListForPrivilegeChange(
	ctx context.Context, p *planner, targets tree.TargetList,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(566831)
	const required = true
	flags := tree.CommonLookupFlags{
		Required:       required,
		AvoidLeased:    p.avoidLeasedDescriptors,
		RequireMutable: true,
	}
	if targets.Databases != nil {
		__antithesis_instrumentation__.Notify(566838)
		if len(targets.Databases) == 0 {
			__antithesis_instrumentation__.Notify(566842)
			return nil, errNoDatabase
		} else {
			__antithesis_instrumentation__.Notify(566843)
		}
		__antithesis_instrumentation__.Notify(566839)
		descs := make([]catalog.Descriptor, 0, len(targets.Databases))
		for _, database := range targets.Databases {
			__antithesis_instrumentation__.Notify(566844)
			descriptor, err := p.Descriptors().
				GetMutableDatabaseByName(ctx, p.txn, string(database), flags)
			if err != nil {
				__antithesis_instrumentation__.Notify(566846)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(566847)
			}
			__antithesis_instrumentation__.Notify(566845)
			descs = append(descs, descriptor)
		}
		__antithesis_instrumentation__.Notify(566840)
		if len(descs) == 0 {
			__antithesis_instrumentation__.Notify(566848)
			return nil, errNoMatch
		} else {
			__antithesis_instrumentation__.Notify(566849)
		}
		__antithesis_instrumentation__.Notify(566841)
		return descs, nil
	} else {
		__antithesis_instrumentation__.Notify(566850)
	}
	__antithesis_instrumentation__.Notify(566832)

	if targets.Types != nil {
		__antithesis_instrumentation__.Notify(566851)
		if len(targets.Types) == 0 {
			__antithesis_instrumentation__.Notify(566855)
			return nil, errNoType
		} else {
			__antithesis_instrumentation__.Notify(566856)
		}
		__antithesis_instrumentation__.Notify(566852)
		descs := make([]catalog.Descriptor, 0, len(targets.Types))
		for _, typ := range targets.Types {
			__antithesis_instrumentation__.Notify(566857)
			_, descriptor, err := p.ResolveMutableTypeDescriptor(ctx, typ, required)
			if err != nil {
				__antithesis_instrumentation__.Notify(566859)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(566860)
			}
			__antithesis_instrumentation__.Notify(566858)

			descs = append(descs, descriptor)
		}
		__antithesis_instrumentation__.Notify(566853)

		if len(descs) == 0 {
			__antithesis_instrumentation__.Notify(566861)
			return nil, errNoMatch
		} else {
			__antithesis_instrumentation__.Notify(566862)
		}
		__antithesis_instrumentation__.Notify(566854)
		return descs, nil
	} else {
		__antithesis_instrumentation__.Notify(566863)
	}
	__antithesis_instrumentation__.Notify(566833)

	if targets.Schemas != nil {
		__antithesis_instrumentation__.Notify(566864)
		if len(targets.Schemas) == 0 {
			__antithesis_instrumentation__.Notify(566869)
			return nil, errNoSchema
		} else {
			__antithesis_instrumentation__.Notify(566870)
		}
		__antithesis_instrumentation__.Notify(566865)
		if targets.AllTablesInSchema {
			__antithesis_instrumentation__.Notify(566871)

			db, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, p.CurrentDatabase(), flags)
			if err != nil {
				__antithesis_instrumentation__.Notify(566874)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(566875)
			}
			__antithesis_instrumentation__.Notify(566872)
			var descs []catalog.Descriptor
			for _, sc := range targets.Schemas {
				__antithesis_instrumentation__.Notify(566876)
				_, objectIDs, err := resolver.GetObjectNamesAndIDs(
					ctx, p.txn, p, p.ExecCfg().Codec, db, sc.Schema(), true,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(566879)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(566880)
				}
				__antithesis_instrumentation__.Notify(566877)
				muts, err := p.Descriptors().GetMutableDescriptorsByID(ctx, p.txn, objectIDs...)
				if err != nil {
					__antithesis_instrumentation__.Notify(566881)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(566882)
				}
				__antithesis_instrumentation__.Notify(566878)
				for _, mut := range muts {
					__antithesis_instrumentation__.Notify(566883)
					if mut != nil && func() bool {
						__antithesis_instrumentation__.Notify(566884)
						return mut.DescriptorType() == catalog.Table == true
					}() == true {
						__antithesis_instrumentation__.Notify(566885)
						descs = append(descs, mut)
					} else {
						__antithesis_instrumentation__.Notify(566886)
					}
				}
			}
			__antithesis_instrumentation__.Notify(566873)

			return descs, nil
		} else {
			__antithesis_instrumentation__.Notify(566887)
		}
		__antithesis_instrumentation__.Notify(566866)

		descs := make([]catalog.Descriptor, 0, len(targets.Schemas))

		type schemaWithDBDesc struct {
			schema string
			dbDesc *dbdesc.Mutable
		}
		var targetSchemas []schemaWithDBDesc
		for _, sc := range targets.Schemas {
			__antithesis_instrumentation__.Notify(566888)
			dbName := p.CurrentDatabase()
			if sc.ExplicitCatalog {
				__antithesis_instrumentation__.Notify(566891)
				dbName = sc.Catalog()
			} else {
				__antithesis_instrumentation__.Notify(566892)
			}
			__antithesis_instrumentation__.Notify(566889)
			db, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, dbName, flags)
			if err != nil {
				__antithesis_instrumentation__.Notify(566893)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(566894)
			}
			__antithesis_instrumentation__.Notify(566890)
			targetSchemas = append(targetSchemas, schemaWithDBDesc{schema: sc.Schema(), dbDesc: db})
		}
		__antithesis_instrumentation__.Notify(566867)

		for _, sc := range targetSchemas {
			__antithesis_instrumentation__.Notify(566895)
			resSchema, err := p.Descriptors().GetSchemaByName(
				ctx, p.txn, sc.dbDesc, sc.schema, flags)
			if err != nil {
				__antithesis_instrumentation__.Notify(566897)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(566898)
			}
			__antithesis_instrumentation__.Notify(566896)
			switch resSchema.SchemaKind() {
			case catalog.SchemaUserDefined:
				__antithesis_instrumentation__.Notify(566899)
				descs = append(descs, resSchema)
			default:
				__antithesis_instrumentation__.Notify(566900)
				return nil, pgerror.Newf(pgcode.InvalidSchemaName,
					"cannot change privileges on schema %q", resSchema.GetName())
			}
		}
		__antithesis_instrumentation__.Notify(566868)
		return descs, nil
	} else {
		__antithesis_instrumentation__.Notify(566901)
	}
	__antithesis_instrumentation__.Notify(566834)

	if len(targets.Tables) == 0 {
		__antithesis_instrumentation__.Notify(566902)
		return nil, errNoTable
	} else {
		__antithesis_instrumentation__.Notify(566903)
	}
	__antithesis_instrumentation__.Notify(566835)
	descs := make([]catalog.Descriptor, 0, len(targets.Tables))
	for _, tableTarget := range targets.Tables {
		__antithesis_instrumentation__.Notify(566904)
		tableGlob, err := tableTarget.NormalizeTablePattern()
		if err != nil {
			__antithesis_instrumentation__.Notify(566908)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(566909)
		}
		__antithesis_instrumentation__.Notify(566905)
		_, objectIDs, err := expandTableGlob(ctx, p, tableGlob)
		if err != nil {
			__antithesis_instrumentation__.Notify(566910)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(566911)
		}
		__antithesis_instrumentation__.Notify(566906)
		muts, err := p.Descriptors().GetMutableDescriptorsByID(ctx, p.txn, objectIDs...)
		if err != nil {
			__antithesis_instrumentation__.Notify(566912)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(566913)
		}
		__antithesis_instrumentation__.Notify(566907)
		for _, mut := range muts {
			__antithesis_instrumentation__.Notify(566914)
			if mut != nil && func() bool {
				__antithesis_instrumentation__.Notify(566915)
				return mut.DescriptorType() == catalog.Table == true
			}() == true {
				__antithesis_instrumentation__.Notify(566916)
				descs = append(descs, mut)
			} else {
				__antithesis_instrumentation__.Notify(566917)
			}
		}
	}
	__antithesis_instrumentation__.Notify(566836)
	if len(descs) == 0 {
		__antithesis_instrumentation__.Notify(566918)
		return nil, errNoMatch
	} else {
		__antithesis_instrumentation__.Notify(566919)
	}
	__antithesis_instrumentation__.Notify(566837)
	return descs, nil
}

func (p *planner) getFullyQualifiedTableNamesFromIDs(
	ctx context.Context, ids []descpb.ID,
) (fullyQualifiedNames []*tree.TableName, _ error) {
	__antithesis_instrumentation__.Notify(566920)
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(566922)
		desc, err := p.Descriptors().GetImmutableTableByID(ctx, p.txn, id, tree.ObjectLookupFlags{
			CommonLookupFlags: tree.CommonLookupFlags{
				AvoidLeased:    true,
				IncludeDropped: true,
				IncludeOffline: true,
			},
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(566925)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(566926)
		}
		__antithesis_instrumentation__.Notify(566923)
		fqName, err := p.getQualifiedTableName(ctx, desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(566927)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(566928)
		}
		__antithesis_instrumentation__.Notify(566924)
		fullyQualifiedNames = append(fullyQualifiedNames, fqName)
	}
	__antithesis_instrumentation__.Notify(566921)
	return fullyQualifiedNames, nil
}

func (p *planner) getQualifiedTableName(
	ctx context.Context, desc catalog.TableDescriptor,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(566929)
	_, dbDesc, err := p.Descriptors().GetImmutableDatabaseByID(ctx, p.txn, desc.GetParentID(),
		tree.DatabaseLookupFlags{
			Required:       true,
			IncludeOffline: true,
			IncludeDropped: true,
			AvoidLeased:    true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(566932)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566933)
	}
	__antithesis_instrumentation__.Notify(566930)

	var schemaName tree.Name
	schemaID := desc.GetParentSchemaID()
	scDesc, err := p.Descriptors().GetImmutableSchemaByID(ctx, p.txn, schemaID,
		tree.SchemaLookupFlags{
			IncludeOffline: true,
			IncludeDropped: true,
			AvoidLeased:    true,
		})
	switch {
	case scDesc != nil:
		__antithesis_instrumentation__.Notify(566934)
		schemaName = tree.Name(scDesc.GetName())
	case desc.IsTemporary() && func() bool {
		__antithesis_instrumentation__.Notify(566937)
		return scDesc == nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(566935)

		schemaName = tree.Name(fmt.Sprintf("pg_temp_%d", schemaID))
	default:
		__antithesis_instrumentation__.Notify(566936)
		return nil, errors.Wrapf(err,
			"resolving schema name for %s.[%d].%s",
			tree.Name(dbDesc.GetName()),
			schemaID,
			tree.Name(desc.GetName()),
		)
	}
	__antithesis_instrumentation__.Notify(566931)

	tbName := tree.MakeTableNameWithSchema(
		tree.Name(dbDesc.GetName()),
		schemaName,
		tree.Name(desc.GetName()),
	)
	return &tbName, nil
}

func (p *planner) GetQualifiedTableNameByID(
	ctx context.Context, id int64, requiredType tree.RequiredTableKind,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(566938)
	lookupFlags := tree.ObjectLookupFlags{
		CommonLookupFlags:    tree.CommonLookupFlags{Required: true},
		DesiredObjectKind:    tree.TableObject,
		DesiredTableDescKind: requiredType,
	}

	table, err := p.Descriptors().GetImmutableTableByID(
		ctx, p.txn, descpb.ID(id), lookupFlags)
	if err != nil {
		__antithesis_instrumentation__.Notify(566940)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566941)
	}
	__antithesis_instrumentation__.Notify(566939)
	return p.getQualifiedTableName(ctx, table)
}

func (p *planner) getQualifiedSchemaName(
	ctx context.Context, desc catalog.SchemaDescriptor,
) (*tree.ObjectNamePrefix, error) {
	__antithesis_instrumentation__.Notify(566942)
	_, dbDesc, err := p.Descriptors().GetImmutableDatabaseByID(ctx, p.txn, desc.GetParentID(),
		tree.DatabaseLookupFlags{
			Required:    true,
			AvoidLeased: true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(566944)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566945)
	}
	__antithesis_instrumentation__.Notify(566943)
	return &tree.ObjectNamePrefix{
		CatalogName:     tree.Name(dbDesc.GetName()),
		SchemaName:      tree.Name(desc.GetName()),
		ExplicitCatalog: true,
		ExplicitSchema:  true,
	}, nil
}

func (p *planner) getQualifiedTypeName(
	ctx context.Context, desc catalog.TypeDescriptor,
) (*tree.TypeName, error) {
	__antithesis_instrumentation__.Notify(566946)
	_, dbDesc, err := p.Descriptors().GetImmutableDatabaseByID(ctx, p.txn, desc.GetParentID(),
		tree.DatabaseLookupFlags{
			Required:    true,
			AvoidLeased: true,
		})
	if err != nil {
		__antithesis_instrumentation__.Notify(566949)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566950)
	}
	__antithesis_instrumentation__.Notify(566947)

	schemaID := desc.GetParentSchemaID()
	scDesc, err := p.Descriptors().GetImmutableSchemaByID(
		ctx, p.txn, schemaID, tree.SchemaLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(566951)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566952)
	}
	__antithesis_instrumentation__.Notify(566948)

	typeName := tree.MakeQualifiedTypeName(
		dbDesc.GetName(),
		scDesc.GetName(),
		desc.GetName(),
	)

	return &typeName, nil
}

func findTableContainingIndex(
	ctx context.Context,
	txn *kv.Txn,
	sc resolver.SchemaResolver,
	codec keys.SQLCodec,
	dbName, scName string,
	idxName tree.UnrestrictedName,
	lookupFlags tree.CommonLookupFlags,
) (result *tree.TableName, desc *tabledesc.Mutable, err error) {
	__antithesis_instrumentation__.Notify(566953)
	sa := sc.Accessor()
	dbDesc, err := sa.GetDatabaseDesc(ctx, txn, dbName, lookupFlags)
	if dbDesc == nil || func() bool {
		__antithesis_instrumentation__.Notify(566958)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(566959)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(566960)
	}
	__antithesis_instrumentation__.Notify(566954)

	tns, _, err := sa.GetObjectNamesAndIDs(
		ctx, txn, dbDesc, scName, tree.DatabaseListFlags{
			CommonLookupFlags: lookupFlags,
			ExplicitPrefix:    true,
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(566961)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(566962)
	}
	__antithesis_instrumentation__.Notify(566955)

	result = nil
	for i := range tns {
		__antithesis_instrumentation__.Notify(566963)
		tn := &tns[i]
		_, tableDesc, err := resolver.ResolveMutableExistingTableObject(
			ctx, sc, tn, false, tree.ResolveAnyTableKind,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(566968)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(566969)
		}
		__antithesis_instrumentation__.Notify(566964)
		if tableDesc == nil || func() bool {
			__antithesis_instrumentation__.Notify(566970)
			return !(tableDesc.IsTable() || func() bool {
				__antithesis_instrumentation__.Notify(566971)
				return tableDesc.MaterializedView() == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(566972)
			continue
		} else {
			__antithesis_instrumentation__.Notify(566973)
		}
		__antithesis_instrumentation__.Notify(566965)

		idx, err := tableDesc.FindIndexWithName(string(idxName))
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(566974)
			return idx.Dropped() == true
		}() == true {
			__antithesis_instrumentation__.Notify(566975)

			continue
		} else {
			__antithesis_instrumentation__.Notify(566976)
		}
		__antithesis_instrumentation__.Notify(566966)
		if result != nil {
			__antithesis_instrumentation__.Notify(566977)
			return nil, nil, pgerror.Newf(pgcode.AmbiguousParameter,
				"index name %q is ambiguous (found in %s and %s)",
				idxName, tn.String(), result.String())
		} else {
			__antithesis_instrumentation__.Notify(566978)
		}
		__antithesis_instrumentation__.Notify(566967)
		result = tn
		desc = tableDesc
	}
	__antithesis_instrumentation__.Notify(566956)
	if result == nil && func() bool {
		__antithesis_instrumentation__.Notify(566979)
		return lookupFlags.Required == true
	}() == true {
		__antithesis_instrumentation__.Notify(566980)
		return nil, nil, pgerror.Newf(pgcode.UndefinedObject,
			"index %q does not exist", idxName)
	} else {
		__antithesis_instrumentation__.Notify(566981)
	}
	__antithesis_instrumentation__.Notify(566957)
	return result, desc, nil
}

func expandMutableIndexName(
	ctx context.Context, p *planner, index *tree.TableIndexName, requireTable bool,
) (tn *tree.TableName, desc *tabledesc.Mutable, err error) {
	__antithesis_instrumentation__.Notify(566982)
	p.runWithOptions(resolveFlags{skipCache: true}, func() {
		__antithesis_instrumentation__.Notify(566984)
		tn, desc, err = expandIndexName(ctx, p.txn, p, p.ExecCfg().Codec, index, requireTable)
	})
	__antithesis_instrumentation__.Notify(566983)
	return tn, desc, err
}

func expandIndexName(
	ctx context.Context,
	txn *kv.Txn,
	sc resolver.SchemaResolver,
	codec keys.SQLCodec,
	index *tree.TableIndexName,
	requireTable bool,
) (tn *tree.TableName, desc *tabledesc.Mutable, err error) {
	__antithesis_instrumentation__.Notify(566985)
	tn = &index.Table
	if tn.Table() != "" {
		__antithesis_instrumentation__.Notify(566991)

		_, desc, err = resolver.ResolveMutableExistingTableObject(ctx, sc, tn, requireTable, tree.ResolveRequireTableOrViewDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(566994)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(566995)
		}
		__antithesis_instrumentation__.Notify(566992)
		if desc != nil && func() bool {
			__antithesis_instrumentation__.Notify(566996)
			return desc.IsView() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(566997)
			return !desc.MaterializedView() == true
		}() == true {
			__antithesis_instrumentation__.Notify(566998)
			return nil, nil, pgerror.Newf(pgcode.WrongObjectType,
				"%q is not a table or materialized view", tn.Table())
		} else {
			__antithesis_instrumentation__.Notify(566999)
		}
		__antithesis_instrumentation__.Notify(566993)
		return tn, desc, nil
	} else {
		__antithesis_instrumentation__.Notify(567000)
	}
	__antithesis_instrumentation__.Notify(566986)

	found, _, err := resolver.ResolveObjectNamePrefix(
		ctx, sc, sc.CurrentDatabase(), sc.CurrentSearchPath(), &tn.ObjectNamePrefix,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(567001)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567002)
	}
	__antithesis_instrumentation__.Notify(566987)
	if !found {
		__antithesis_instrumentation__.Notify(567003)
		if requireTable {
			__antithesis_instrumentation__.Notify(567005)
			err = pgerror.Newf(pgcode.UndefinedObject,
				"schema or database was not found while searching index: %q",
				tree.ErrString(&index.Index))
			err = errors.WithHint(err, "check the current database and search_path are valid")
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(567006)
		}
		__antithesis_instrumentation__.Notify(567004)
		return nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(567007)
	}
	__antithesis_instrumentation__.Notify(566988)

	lookupFlags := sc.CommonLookupFlags(requireTable)
	var foundTn *tree.TableName
	foundTn, desc, err = findTableContainingIndex(
		ctx, txn, sc, codec, tn.Catalog(), tn.Schema(), index.Index, lookupFlags,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(567008)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567009)
	}
	__antithesis_instrumentation__.Notify(566989)

	if foundTn != nil {
		__antithesis_instrumentation__.Notify(567010)

		*tn = *foundTn
	} else {
		__antithesis_instrumentation__.Notify(567011)
	}
	__antithesis_instrumentation__.Notify(566990)
	return tn, desc, nil
}

func (p *planner) getTableAndIndex(
	ctx context.Context, tableWithIndex *tree.TableIndexName, privilege privilege.Kind,
) (*tabledesc.Mutable, catalog.Index, error) {
	__antithesis_instrumentation__.Notify(567012)
	var catalog optCatalog
	catalog.init(p)
	catalog.reset()

	idx, qualifiedName, err := cat.ResolveTableIndex(
		ctx, &catalog, cat.Flags{AvoidDescriptorCaches: true}, tableWithIndex,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(567018)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567019)
	}
	__antithesis_instrumentation__.Notify(567013)
	if err := catalog.CheckPrivilege(ctx, idx.Table(), privilege); err != nil {
		__antithesis_instrumentation__.Notify(567020)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567021)
	}
	__antithesis_instrumentation__.Notify(567014)
	optIdx := idx.(*optIndex)

	if tableWithIndex.Table.ObjectName == "" {
		__antithesis_instrumentation__.Notify(567022)
		tableWithIndex.Table = tree.MakeTableNameFromPrefix(
			qualifiedName.ObjectNamePrefix, qualifiedName.ObjectName,
		)
	} else {
		__antithesis_instrumentation__.Notify(567023)
	}
	__antithesis_instrumentation__.Notify(567015)

	tableID := optIdx.tab.desc.GetID()
	mut, err := p.Descriptors().GetMutableTableVersionByID(ctx, tableID, p.Txn())
	if err != nil {
		__antithesis_instrumentation__.Notify(567024)
		return nil, nil, errors.NewAssertionErrorWithWrappedErrf(err,
			"failed to re-resolve table %d for index %s", tableID, tableWithIndex)
	} else {
		__antithesis_instrumentation__.Notify(567025)
	}
	__antithesis_instrumentation__.Notify(567016)
	retIdx, err := mut.FindIndexWithID(optIdx.idx.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(567026)
		return nil, nil, errors.NewAssertionErrorWithWrappedErrf(err,
			"retrieving index %s (%d) from table which was known to already exist for table %d",
			tableWithIndex, optIdx.idx.GetID(), tableID)
	} else {
		__antithesis_instrumentation__.Notify(567027)
	}
	__antithesis_instrumentation__.Notify(567017)
	return mut, retIdx, nil
}

func expandTableGlob(
	ctx context.Context, p *planner, pattern tree.TablePattern,
) (tree.TableNames, descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(567028)
	var catalog optCatalog
	catalog.init(p)
	catalog.reset()

	return cat.ExpandDataSourceGlob(ctx, &catalog, cat.Flags{}, pattern)
}

type fkSelfResolver struct {
	resolver.SchemaResolver
	prefix       catalog.ResolvedObjectPrefix
	newTableName *tree.TableName
	newTableDesc *tabledesc.Mutable
}

var _ resolver.SchemaResolver = &fkSelfResolver{}

func (r *fkSelfResolver) LookupObject(
	ctx context.Context, flags tree.ObjectLookupFlags, dbName, scName, obName string,
) (found bool, prefix catalog.ResolvedObjectPrefix, objMeta catalog.Descriptor, err error) {
	__antithesis_instrumentation__.Notify(567029)
	if dbName == r.prefix.Database.GetName() && func() bool {
		__antithesis_instrumentation__.Notify(567031)
		return scName == r.prefix.Schema.GetName() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(567032)
		return obName == r.newTableName.Table() == true
	}() == true {
		__antithesis_instrumentation__.Notify(567033)
		table := r.newTableDesc
		if flags.RequireMutable {
			__antithesis_instrumentation__.Notify(567035)
			return true, r.prefix, table, nil
		} else {
			__antithesis_instrumentation__.Notify(567036)
		}
		__antithesis_instrumentation__.Notify(567034)
		return true, r.prefix, table.ImmutableCopy(), nil
	} else {
		__antithesis_instrumentation__.Notify(567037)
	}
	__antithesis_instrumentation__.Notify(567030)
	flags.IncludeOffline = false
	return r.SchemaResolver.LookupObject(ctx, flags, dbName, scName, obName)
}

type internalLookupCtx struct {
	dbNames     map[descpb.ID]string
	dbIDs       []descpb.ID
	dbDescs     map[descpb.ID]catalog.DatabaseDescriptor
	schemaDescs map[descpb.ID]catalog.SchemaDescriptor
	schemaNames map[descpb.ID]string
	schemaIDs   []descpb.ID
	tbDescs     map[descpb.ID]catalog.TableDescriptor
	tbIDs       []descpb.ID
	typDescs    map[descpb.ID]catalog.TypeDescriptor
	typIDs      []descpb.ID
}

func (l *internalLookupCtx) GetSchemaName(
	ctx context.Context, id, parentDBID descpb.ID, version clusterversion.Handle,
) (string, bool, error) {
	__antithesis_instrumentation__.Notify(567038)
	dbDesc, err := l.getDatabaseByID(parentDBID)
	if err != nil {
		__antithesis_instrumentation__.Notify(567042)
		return "", false, err
	} else {
		__antithesis_instrumentation__.Notify(567043)
	}
	__antithesis_instrumentation__.Notify(567039)

	if !dbDesc.HasPublicSchemaWithDescriptor() {
		__antithesis_instrumentation__.Notify(567044)
		if id == keys.PublicSchemaID {
			__antithesis_instrumentation__.Notify(567045)
			return tree.PublicSchema, true, nil
		} else {
			__antithesis_instrumentation__.Notify(567046)
		}
	} else {
		__antithesis_instrumentation__.Notify(567047)
	}
	__antithesis_instrumentation__.Notify(567040)

	if parentDBID == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(567048)
		if id == keys.SystemPublicSchemaID {
			__antithesis_instrumentation__.Notify(567049)
			return tree.PublicSchema, true, nil
		} else {
			__antithesis_instrumentation__.Notify(567050)
		}
	} else {
		__antithesis_instrumentation__.Notify(567051)
	}
	__antithesis_instrumentation__.Notify(567041)

	schemaName, found := l.schemaNames[id]
	return schemaName, found, nil
}

type tableLookupFn = *internalLookupCtx

func newInternalLookupCtxFromDescriptorProtos(
	ctx context.Context, rawDescs []descpb.Descriptor, prefix catalog.DatabaseDescriptor,
) (*internalLookupCtx, error) {
	__antithesis_instrumentation__.Notify(567052)
	descriptors := make([]catalog.Descriptor, len(rawDescs))
	for i := range rawDescs {
		__antithesis_instrumentation__.Notify(567055)
		descriptors[i] = descbuilder.NewBuilder(&rawDescs[i]).BuildImmutable()
	}
	__antithesis_instrumentation__.Notify(567053)
	lCtx := newInternalLookupCtx(descriptors, prefix)
	if err := descs.HydrateGivenDescriptors(ctx, descriptors); err != nil {
		__antithesis_instrumentation__.Notify(567056)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(567057)
	}
	__antithesis_instrumentation__.Notify(567054)
	return lCtx, nil
}

func newInternalLookupCtx(
	descs []catalog.Descriptor, prefix catalog.DatabaseDescriptor,
) *internalLookupCtx {
	__antithesis_instrumentation__.Notify(567058)
	dbNames := make(map[descpb.ID]string)
	dbDescs := make(map[descpb.ID]catalog.DatabaseDescriptor)
	schemaDescs := make(map[descpb.ID]catalog.SchemaDescriptor)
	schemaNames := make(map[descpb.ID]string)

	tbDescs := make(map[descpb.ID]catalog.TableDescriptor)
	typDescs := make(map[descpb.ID]catalog.TypeDescriptor)
	var tbIDs, typIDs, dbIDs, schemaIDs []descpb.ID

	for i := range descs {
		__antithesis_instrumentation__.Notify(567060)
		switch desc := descs[i].(type) {
		case catalog.DatabaseDescriptor:
			__antithesis_instrumentation__.Notify(567061)
			dbNames[desc.GetID()] = desc.GetName()
			dbDescs[desc.GetID()] = desc
			if prefix == nil || func() bool {
				__antithesis_instrumentation__.Notify(567065)
				return prefix.GetID() == desc.GetID() == true
			}() == true {
				__antithesis_instrumentation__.Notify(567066)

				dbIDs = append(dbIDs, desc.GetID())
			} else {
				__antithesis_instrumentation__.Notify(567067)
			}
		case catalog.TableDescriptor:
			__antithesis_instrumentation__.Notify(567062)
			tbDescs[desc.GetID()] = desc
			if prefix == nil || func() bool {
				__antithesis_instrumentation__.Notify(567068)
				return prefix.GetID() == desc.GetParentID() == true
			}() == true {
				__antithesis_instrumentation__.Notify(567069)

				tbIDs = append(tbIDs, desc.GetID())
			} else {
				__antithesis_instrumentation__.Notify(567070)
			}
		case catalog.TypeDescriptor:
			__antithesis_instrumentation__.Notify(567063)
			typDescs[desc.GetID()] = desc
			if prefix == nil || func() bool {
				__antithesis_instrumentation__.Notify(567071)
				return prefix.GetID() == desc.GetParentID() == true
			}() == true {
				__antithesis_instrumentation__.Notify(567072)

				typIDs = append(typIDs, desc.GetID())
			} else {
				__antithesis_instrumentation__.Notify(567073)
			}
		case catalog.SchemaDescriptor:
			__antithesis_instrumentation__.Notify(567064)
			schemaDescs[desc.GetID()] = desc
			if prefix == nil || func() bool {
				__antithesis_instrumentation__.Notify(567074)
				return prefix.GetID() == desc.GetParentID() == true
			}() == true {
				__antithesis_instrumentation__.Notify(567075)

				schemaIDs = append(schemaIDs, desc.GetID())
				schemaNames[desc.GetID()] = desc.GetName()
			} else {
				__antithesis_instrumentation__.Notify(567076)
			}
		}
	}
	__antithesis_instrumentation__.Notify(567059)

	return &internalLookupCtx{
		dbNames:     dbNames,
		dbDescs:     dbDescs,
		schemaDescs: schemaDescs,
		schemaNames: schemaNames,
		schemaIDs:   schemaIDs,
		tbDescs:     tbDescs,
		typDescs:    typDescs,
		tbIDs:       tbIDs,
		dbIDs:       dbIDs,
		typIDs:      typIDs,
	}
}

func (l *internalLookupCtx) getDatabaseByID(id descpb.ID) (catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(567077)
	db, ok := l.dbDescs[id]
	if !ok {
		__antithesis_instrumentation__.Notify(567079)
		return nil, sqlerrors.NewUndefinedDatabaseError(fmt.Sprintf("[%d]", id))
	} else {
		__antithesis_instrumentation__.Notify(567080)
	}
	__antithesis_instrumentation__.Notify(567078)
	return db, nil
}

func (l *internalLookupCtx) getTableByID(id descpb.ID) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(567081)
	tb, ok := l.tbDescs[id]
	if !ok {
		__antithesis_instrumentation__.Notify(567083)
		return nil, sqlerrors.NewUndefinedRelationError(
			tree.NewUnqualifiedTableName(tree.Name(fmt.Sprintf("[%d]", id))))
	} else {
		__antithesis_instrumentation__.Notify(567084)
	}
	__antithesis_instrumentation__.Notify(567082)
	return tb, nil
}

func (l *internalLookupCtx) getTypeByID(id descpb.ID) (catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(567085)
	typ, ok := l.typDescs[id]
	if !ok {
		__antithesis_instrumentation__.Notify(567087)
		return nil, sqlerrors.NewUndefinedRelationError(
			tree.NewUnqualifiedTableName(tree.Name(fmt.Sprintf("[%d]", id))))
	} else {
		__antithesis_instrumentation__.Notify(567088)
	}
	__antithesis_instrumentation__.Notify(567086)
	return typ, nil
}

func (l *internalLookupCtx) getSchemaByID(id descpb.ID) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(567089)
	sc, ok := l.schemaDescs[id]
	if !ok {
		__antithesis_instrumentation__.Notify(567091)
		return nil, sqlerrors.NewUndefinedSchemaError(fmt.Sprintf("[%d]", id))
	} else {
		__antithesis_instrumentation__.Notify(567092)
	}
	__antithesis_instrumentation__.Notify(567090)
	return sc, nil
}

func (l *internalLookupCtx) getSchemaNameByID(id descpb.ID) (string, error) {
	__antithesis_instrumentation__.Notify(567093)

	if id == keys.PublicSchemaID {
		__antithesis_instrumentation__.Notify(567096)
		return tree.PublicSchema, nil
	} else {
		__antithesis_instrumentation__.Notify(567097)
	}
	__antithesis_instrumentation__.Notify(567094)
	schema, err := l.getSchemaByID(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(567098)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(567099)
	}
	__antithesis_instrumentation__.Notify(567095)
	return schema.GetName(), nil
}

func (l *internalLookupCtx) getDatabaseName(table catalog.Descriptor) string {
	__antithesis_instrumentation__.Notify(567100)
	parentName := l.dbNames[table.GetParentID()]
	if parentName == "" {
		__antithesis_instrumentation__.Notify(567102)

		parentName = fmt.Sprintf("[%d]", table.GetParentID())
	} else {
		__antithesis_instrumentation__.Notify(567103)
	}
	__antithesis_instrumentation__.Notify(567101)
	return parentName
}

func (l *internalLookupCtx) getSchemaName(table catalog.TableDescriptor) string {
	__antithesis_instrumentation__.Notify(567104)
	schemaName := l.schemaNames[table.GetParentSchemaID()]
	if schemaName == "" {
		__antithesis_instrumentation__.Notify(567106)

		schemaName = fmt.Sprintf("[%d]", table.GetParentSchemaID())
	} else {
		__antithesis_instrumentation__.Notify(567107)
	}
	__antithesis_instrumentation__.Notify(567105)
	return schemaName
}

func getTableNameFromTableDescriptor(
	l simpleSchemaResolver, table catalog.TableDescriptor, dbPrefix string,
) (tree.TableName, error) {
	__antithesis_instrumentation__.Notify(567108)
	var tableName tree.TableName
	tableDbDesc, err := l.getDatabaseByID(table.GetParentID())
	if err != nil {
		__antithesis_instrumentation__.Notify(567111)
		return tree.TableName{}, err
	} else {
		__antithesis_instrumentation__.Notify(567112)
	}
	__antithesis_instrumentation__.Notify(567109)
	var parentSchemaName tree.Name

	if table.GetParentSchemaID() == keys.PublicSchemaID {
		__antithesis_instrumentation__.Notify(567113)
		parentSchemaName = tree.PublicSchemaName
	} else {
		__antithesis_instrumentation__.Notify(567114)
		parentSchema, err := l.getSchemaByID(table.GetParentSchemaID())
		if err != nil {
			__antithesis_instrumentation__.Notify(567116)
			return tree.TableName{}, err
		} else {
			__antithesis_instrumentation__.Notify(567117)
		}
		__antithesis_instrumentation__.Notify(567115)
		parentSchemaName = tree.Name(parentSchema.GetName())
	}
	__antithesis_instrumentation__.Notify(567110)
	tableName = tree.MakeTableNameWithSchema(tree.Name(tableDbDesc.GetName()),
		parentSchemaName, tree.Name(table.GetName()))
	tableName.ExplicitCatalog = tableDbDesc.GetName() != dbPrefix
	return tableName, nil
}

func getTypeNameFromTypeDescriptor(
	l simpleSchemaResolver, typ catalog.TypeDescriptor,
) (tree.TypeName, error) {
	__antithesis_instrumentation__.Notify(567118)
	var typeName tree.TypeName
	tableDbDesc, err := l.getDatabaseByID(typ.GetParentID())
	if err != nil {
		__antithesis_instrumentation__.Notify(567121)
		return typeName, err
	} else {
		__antithesis_instrumentation__.Notify(567122)
	}
	__antithesis_instrumentation__.Notify(567119)
	var parentSchemaName string

	if typ.GetParentSchemaID() == keys.PublicSchemaID {
		__antithesis_instrumentation__.Notify(567123)
		parentSchemaName = tree.PublicSchema
	} else {
		__antithesis_instrumentation__.Notify(567124)
		parentSchema, err := l.getSchemaByID(typ.GetParentSchemaID())
		if err != nil {
			__antithesis_instrumentation__.Notify(567126)
			return typeName, err
		} else {
			__antithesis_instrumentation__.Notify(567127)
		}
		__antithesis_instrumentation__.Notify(567125)
		parentSchemaName = parentSchema.GetName()
	}
	__antithesis_instrumentation__.Notify(567120)
	typeName = tree.MakeQualifiedTypeName(tableDbDesc.GetName(),
		parentSchemaName, typ.GetName())
	return typeName, nil
}

func (p *planner) ResolveMutableTypeDescriptor(
	ctx context.Context, name *tree.UnresolvedObjectName, required bool,
) (catalog.ResolvedObjectPrefix, *typedesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(567128)
	prefix, desc, err := resolver.ResolveMutableType(ctx, p, name, required)
	if err != nil {
		__antithesis_instrumentation__.Notify(567131)
		return catalog.ResolvedObjectPrefix{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567132)
	}
	__antithesis_instrumentation__.Notify(567129)

	if desc != nil {
		__antithesis_instrumentation__.Notify(567133)

		if err := p.canResolveDescUnderSchema(ctx, prefix.Schema, desc); err != nil {
			__antithesis_instrumentation__.Notify(567135)
			return catalog.ResolvedObjectPrefix{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(567136)
		}
		__antithesis_instrumentation__.Notify(567134)
		tn := tree.MakeTypeNameWithPrefix(prefix.NamePrefix(), desc.GetName())
		name.SetAnnotation(&p.semaCtx.Annotations, &tn)
	} else {
		__antithesis_instrumentation__.Notify(567137)
	}
	__antithesis_instrumentation__.Notify(567130)

	return prefix, desc, nil
}

func (p *planner) ResolveMutableTableDescriptorEx(
	ctx context.Context,
	name *tree.UnresolvedObjectName,
	required bool,
	requiredType tree.RequiredTableKind,
) (catalog.ResolvedObjectPrefix, *tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(567138)
	tn := name.ToTableName()
	prefix, table, err := resolver.ResolveMutableExistingTableObject(ctx, p, &tn, required, requiredType)
	if err != nil {
		__antithesis_instrumentation__.Notify(567141)
		return catalog.ResolvedObjectPrefix{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567142)
	}
	__antithesis_instrumentation__.Notify(567139)
	name.SetAnnotation(&p.semaCtx.Annotations, &tn)

	if table != nil {
		__antithesis_instrumentation__.Notify(567143)

		if err := p.canResolveDescUnderSchema(ctx, prefix.Schema, table); err != nil {
			__antithesis_instrumentation__.Notify(567144)
			return catalog.ResolvedObjectPrefix{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(567145)
		}
	} else {
		__antithesis_instrumentation__.Notify(567146)
	}
	__antithesis_instrumentation__.Notify(567140)

	return prefix, table, nil
}

func (p *planner) ResolveMutableTableDescriptorExAllowNoPrimaryKey(
	ctx context.Context,
	name *tree.UnresolvedObjectName,
	required bool,
	requiredType tree.RequiredTableKind,
) (catalog.ResolvedObjectPrefix, *tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(567147)
	lookupFlags := tree.ObjectLookupFlags{
		CommonLookupFlags:      tree.CommonLookupFlags{Required: required, RequireMutable: true},
		AllowWithoutPrimaryKey: true,
		DesiredObjectKind:      tree.TableObject,
		DesiredTableDescKind:   requiredType,
	}
	desc, prefix, err := resolver.ResolveExistingObject(ctx, p, name, lookupFlags)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(567150)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(567151)
		return catalog.ResolvedObjectPrefix{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567152)
	}
	__antithesis_instrumentation__.Notify(567148)
	tn := tree.MakeTableNameFromPrefix(prefix.NamePrefix(), tree.Name(name.Object()))
	name.SetAnnotation(&p.semaCtx.Annotations, &tn)
	table := desc.(*tabledesc.Mutable)

	if err := p.canResolveDescUnderSchema(ctx, prefix.Schema, table); err != nil {
		__antithesis_instrumentation__.Notify(567153)
		return catalog.ResolvedObjectPrefix{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(567154)
	}
	__antithesis_instrumentation__.Notify(567149)

	return prefix, table, err
}

func (p *planner) ResolveUncachedTableDescriptorEx(
	ctx context.Context,
	name *tree.UnresolvedObjectName,
	required bool,
	requiredType tree.RequiredTableKind,
) (table catalog.TableDescriptor, err error) {
	__antithesis_instrumentation__.Notify(567155)
	p.runWithOptions(resolveFlags{skipCache: true}, func() {
		__antithesis_instrumentation__.Notify(567157)
		table, err = p.ResolveExistingObjectEx(ctx, name, required, requiredType)
	})
	__antithesis_instrumentation__.Notify(567156)
	return table, err
}

func (p *planner) ResolveExistingObjectEx(
	ctx context.Context,
	name *tree.UnresolvedObjectName,
	required bool,
	requiredType tree.RequiredTableKind,
) (res catalog.TableDescriptor, err error) {
	__antithesis_instrumentation__.Notify(567158)
	lookupFlags := tree.ObjectLookupFlags{
		CommonLookupFlags:    p.CommonLookupFlags(required),
		DesiredObjectKind:    tree.TableObject,
		DesiredTableDescKind: requiredType,
	}
	desc, prefix, err := resolver.ResolveExistingObject(ctx, p, name, lookupFlags)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(567161)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(567162)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(567163)
	}
	__antithesis_instrumentation__.Notify(567159)
	tn := tree.MakeTableNameFromPrefix(prefix.NamePrefix(), tree.Name(name.Object()))
	name.SetAnnotation(&p.semaCtx.Annotations, &tn)
	table := desc.(catalog.TableDescriptor)

	if err := p.canResolveDescUnderSchema(ctx, prefix.Schema, table); err != nil {
		__antithesis_instrumentation__.Notify(567164)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(567165)
	}
	__antithesis_instrumentation__.Notify(567160)

	return table, nil
}

func (p *planner) ResolvedName(u *tree.UnresolvedObjectName) tree.ObjectName {
	__antithesis_instrumentation__.Notify(567166)
	return u.Resolved(&p.semaCtx.Annotations)
}

type simpleSchemaResolver interface {
	getDatabaseByID(id descpb.ID) (catalog.DatabaseDescriptor, error)
	getSchemaByID(id descpb.ID) (catalog.SchemaDescriptor, error)
	getTableByID(id descpb.ID) (catalog.TableDescriptor, error)
}
