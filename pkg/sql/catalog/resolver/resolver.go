package resolver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type SchemaResolver interface {
	ObjectNameExistingResolver
	ObjectNameTargetResolver
	tree.QualifiedNameResolver
	tree.TypeReferenceResolver

	Accessor() catalog.Accessor
	CurrentSearchPath() sessiondata.SearchPath
	CommonLookupFlags(required bool) tree.CommonLookupFlags
}

type ObjectNameExistingResolver interface {
	LookupObject(
		ctx context.Context, flags tree.ObjectLookupFlags,
		dbName, scName, obName string,
	) (
		found bool,
		prefix catalog.ResolvedObjectPrefix,
		objMeta catalog.Descriptor,
		err error,
	)
}

type ObjectNameTargetResolver interface {
	LookupSchema(
		ctx context.Context, dbName, scName string,
	) (found bool, scMeta catalog.ResolvedObjectPrefix, err error)
}

var ErrNoPrimaryKey = pgerror.Newf(pgcode.NoPrimaryKey,
	"requested table does not have a primary key")

func GetObjectNamesAndIDs(
	ctx context.Context,
	txn *kv.Txn,
	sc SchemaResolver,
	codec keys.SQLCodec,
	dbDesc catalog.DatabaseDescriptor,
	scName string,
	explicitPrefix bool,
) (tree.TableNames, descpb.IDs, error) {
	__antithesis_instrumentation__.Notify(267156)
	return sc.Accessor().GetObjectNamesAndIDs(ctx, txn, dbDesc, scName, tree.DatabaseListFlags{
		CommonLookupFlags: sc.CommonLookupFlags(true),
		ExplicitPrefix:    explicitPrefix,
	})
}

func ResolveExistingTableObject(
	ctx context.Context, sc SchemaResolver, tn *tree.TableName, lookupFlags tree.ObjectLookupFlags,
) (prefix catalog.ResolvedObjectPrefix, res catalog.TableDescriptor, err error) {
	__antithesis_instrumentation__.Notify(267157)

	un := tn.ToUnresolvedObjectName()
	desc, prefix, err := ResolveExistingObject(ctx, sc, un, lookupFlags)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(267159)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(267160)
		return prefix, nil, err
	} else {
		__antithesis_instrumentation__.Notify(267161)
	}
	__antithesis_instrumentation__.Notify(267158)
	tn.ObjectNamePrefix = prefix.NamePrefix()
	return prefix, desc.(catalog.TableDescriptor), nil
}

func ResolveMutableExistingTableObject(
	ctx context.Context,
	sc SchemaResolver,
	tn *tree.TableName,
	required bool,
	requiredType tree.RequiredTableKind,
) (prefix catalog.ResolvedObjectPrefix, res *tabledesc.Mutable, err error) {
	__antithesis_instrumentation__.Notify(267162)
	lookupFlags := tree.ObjectLookupFlags{
		CommonLookupFlags:    tree.CommonLookupFlags{Required: required, RequireMutable: true},
		DesiredObjectKind:    tree.TableObject,
		DesiredTableDescKind: requiredType,
	}

	un := tn.ToUnresolvedObjectName()
	var desc catalog.Descriptor
	desc, prefix, err = ResolveExistingObject(ctx, sc, un, lookupFlags)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(267164)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(267165)
		return prefix, nil, err
	} else {
		__antithesis_instrumentation__.Notify(267166)
	}
	__antithesis_instrumentation__.Notify(267163)
	tn.ObjectNamePrefix = prefix.NamePrefix()
	return prefix, desc.(*tabledesc.Mutable), nil
}

func ResolveMutableType(
	ctx context.Context, sc SchemaResolver, un *tree.UnresolvedObjectName, required bool,
) (catalog.ResolvedObjectPrefix, *typedesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(267167)
	lookupFlags := tree.ObjectLookupFlags{
		CommonLookupFlags: tree.CommonLookupFlags{Required: required, RequireMutable: true},
		DesiredObjectKind: tree.TypeObject,
	}
	desc, prefix, err := ResolveExistingObject(ctx, sc, un, lookupFlags)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(267169)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(267170)
		return catalog.ResolvedObjectPrefix{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(267171)
	}
	__antithesis_instrumentation__.Notify(267168)
	switch t := desc.(type) {
	case *typedesc.Mutable:
		__antithesis_instrumentation__.Notify(267172)
		return prefix, t, nil
	case *typedesc.TableImplicitRecordType:
		__antithesis_instrumentation__.Notify(267173)
		return catalog.ResolvedObjectPrefix{}, nil, pgerror.Newf(pgcode.DependentObjectsStillExist,
			"cannot modify table record type %q", desc.GetName())
	default:
		__antithesis_instrumentation__.Notify(267174)
		return catalog.ResolvedObjectPrefix{}, nil,
			errors.AssertionFailedf("unhandled type descriptor type %T during resolve mutable desc", t)
	}
}

func ResolveExistingObject(
	ctx context.Context,
	sc SchemaResolver,
	un *tree.UnresolvedObjectName,
	lookupFlags tree.ObjectLookupFlags,
) (res catalog.Descriptor, _ catalog.ResolvedObjectPrefix, err error) {
	__antithesis_instrumentation__.Notify(267175)
	found, prefix, obj, err := ResolveExisting(ctx, un, sc, lookupFlags, sc.CurrentDatabase(), sc.CurrentSearchPath())
	if err != nil {
		__antithesis_instrumentation__.Notify(267179)
		return nil, prefix, err
	} else {
		__antithesis_instrumentation__.Notify(267180)
	}
	__antithesis_instrumentation__.Notify(267176)

	if !found {
		__antithesis_instrumentation__.Notify(267181)
		if lookupFlags.Required {
			__antithesis_instrumentation__.Notify(267183)

			if un.HasExplicitCatalog() && func() bool {
				__antithesis_instrumentation__.Notify(267185)
				return un.Catalog() != "" == true
			}() == true {
				__antithesis_instrumentation__.Notify(267186)
				if prefix.Database == nil {
					__antithesis_instrumentation__.Notify(267188)
					return nil, prefix, sqlerrors.NewUndefinedDatabaseError(un.Catalog())
				} else {
					__antithesis_instrumentation__.Notify(267189)
				}
				__antithesis_instrumentation__.Notify(267187)
				if un.HasExplicitSchema() && func() bool {
					__antithesis_instrumentation__.Notify(267190)
					return prefix.Schema == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(267191)
					return nil, prefix, sqlerrors.NewUndefinedSchemaError(un.Schema())
				} else {
					__antithesis_instrumentation__.Notify(267192)
				}
			} else {
				__antithesis_instrumentation__.Notify(267193)
			}
			__antithesis_instrumentation__.Notify(267184)
			return nil, prefix, sqlerrors.NewUndefinedObjectError(un, lookupFlags.DesiredObjectKind)
		} else {
			__antithesis_instrumentation__.Notify(267194)
		}
		__antithesis_instrumentation__.Notify(267182)
		return nil, prefix, nil
	} else {
		__antithesis_instrumentation__.Notify(267195)
	}
	__antithesis_instrumentation__.Notify(267177)
	getResolvedTn := func() *tree.TableName {
		__antithesis_instrumentation__.Notify(267196)
		tn := tree.MakeTableNameFromPrefix(prefix.NamePrefix(), tree.Name(un.Object()))
		return &tn
	}
	__antithesis_instrumentation__.Notify(267178)

	switch lookupFlags.DesiredObjectKind {
	case tree.TypeObject:
		__antithesis_instrumentation__.Notify(267197)
		typ, isType := obj.(catalog.TypeDescriptor)
		if !isType {
			__antithesis_instrumentation__.Notify(267205)
			return nil, prefix, sqlerrors.NewUndefinedTypeError(getResolvedTn())
		} else {
			__antithesis_instrumentation__.Notify(267206)
		}
		__antithesis_instrumentation__.Notify(267198)
		return typ, prefix, nil
	case tree.TableObject:
		__antithesis_instrumentation__.Notify(267199)
		table, ok := obj.(catalog.TableDescriptor)
		if !ok {
			__antithesis_instrumentation__.Notify(267207)
			return nil, prefix, sqlerrors.NewUndefinedRelationError(getResolvedTn())
		} else {
			__antithesis_instrumentation__.Notify(267208)
		}
		__antithesis_instrumentation__.Notify(267200)
		goodType := true
		switch lookupFlags.DesiredTableDescKind {
		case tree.ResolveRequireTableDesc:
			__antithesis_instrumentation__.Notify(267209)
			goodType = table.IsTable()
		case tree.ResolveRequireViewDesc:
			__antithesis_instrumentation__.Notify(267210)
			goodType = table.IsView()
		case tree.ResolveRequireTableOrViewDesc:
			__antithesis_instrumentation__.Notify(267211)
			goodType = table.IsTable() || func() bool {
				__antithesis_instrumentation__.Notify(267214)
				return table.IsView() == true
			}() == true
		case tree.ResolveRequireSequenceDesc:
			__antithesis_instrumentation__.Notify(267212)
			goodType = table.IsSequence()
		default:
			__antithesis_instrumentation__.Notify(267213)
		}
		__antithesis_instrumentation__.Notify(267201)
		if !goodType {
			__antithesis_instrumentation__.Notify(267215)
			return nil, prefix, sqlerrors.NewWrongObjectTypeError(getResolvedTn(), lookupFlags.DesiredTableDescKind.String())
		} else {
			__antithesis_instrumentation__.Notify(267216)
		}
		__antithesis_instrumentation__.Notify(267202)

		if !lookupFlags.AllowWithoutPrimaryKey && func() bool {
			__antithesis_instrumentation__.Notify(267217)
			return table.IsTable() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(267218)
			return !table.HasPrimaryKey() == true
		}() == true {
			__antithesis_instrumentation__.Notify(267219)
			return nil, prefix, ErrNoPrimaryKey
		} else {
			__antithesis_instrumentation__.Notify(267220)
		}
		__antithesis_instrumentation__.Notify(267203)

		return obj.(catalog.TableDescriptor), prefix, nil
	default:
		__antithesis_instrumentation__.Notify(267204)
		return nil, prefix, errors.AssertionFailedf(
			"unknown desired object kind %d", lookupFlags.DesiredObjectKind)
	}
}

func ResolveTargetObject(
	ctx context.Context, sc SchemaResolver, un *tree.UnresolvedObjectName,
) (catalog.ResolvedObjectPrefix, tree.ObjectNamePrefix, error) {
	__antithesis_instrumentation__.Notify(267221)
	found, prefix, scInfo, err := ResolveTarget(ctx, un, sc, sc.CurrentDatabase(), sc.CurrentSearchPath())
	if err != nil {
		__antithesis_instrumentation__.Notify(267225)
		return catalog.ResolvedObjectPrefix{}, prefix, err
	} else {
		__antithesis_instrumentation__.Notify(267226)
	}
	__antithesis_instrumentation__.Notify(267222)
	if !found {
		__antithesis_instrumentation__.Notify(267227)
		if !un.HasExplicitSchema() && func() bool {
			__antithesis_instrumentation__.Notify(267229)
			return !un.HasExplicitCatalog() == true
		}() == true {
			__antithesis_instrumentation__.Notify(267230)
			return catalog.ResolvedObjectPrefix{}, prefix,
				pgerror.New(pgcode.InvalidName, "no database specified")
		} else {
			__antithesis_instrumentation__.Notify(267231)
		}
		__antithesis_instrumentation__.Notify(267228)
		err = pgerror.Newf(pgcode.InvalidSchemaName,
			"cannot create %q because the target database or schema does not exist",
			tree.ErrString(un))
		err = errors.WithHint(err, "verify that the current database and search_path are valid and/or the target database exists")
		return catalog.ResolvedObjectPrefix{}, prefix, err
	} else {
		__antithesis_instrumentation__.Notify(267232)
	}
	__antithesis_instrumentation__.Notify(267223)
	if scInfo.Schema.SchemaKind() == catalog.SchemaVirtual {
		__antithesis_instrumentation__.Notify(267233)
		return catalog.ResolvedObjectPrefix{}, prefix, pgerror.Newf(pgcode.InsufficientPrivilege,
			"schema cannot be modified: %q", tree.ErrString(&prefix))
	} else {
		__antithesis_instrumentation__.Notify(267234)
	}
	__antithesis_instrumentation__.Notify(267224)
	return scInfo, prefix, nil
}

func ResolveSchemaNameByID(
	ctx context.Context,
	txn *kv.Txn,
	codec keys.SQLCodec,
	db catalog.DatabaseDescriptor,
	schemaID descpb.ID,
) (string, error) {
	__antithesis_instrumentation__.Notify(267235)

	staticSchemaMap := catconstants.GetStaticSchemaIDMap()
	if schemaName, ok := staticSchemaMap[uint32(schemaID)]; ok {
		__antithesis_instrumentation__.Notify(267239)
		return schemaName, nil
	} else {
		__antithesis_instrumentation__.Notify(267240)
	}
	__antithesis_instrumentation__.Notify(267236)
	schemas, err := GetForDatabase(ctx, txn, codec, db)
	if err != nil {
		__antithesis_instrumentation__.Notify(267241)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(267242)
	}
	__antithesis_instrumentation__.Notify(267237)
	if schema, ok := schemas[schemaID]; ok {
		__antithesis_instrumentation__.Notify(267243)
		return schema.Name, nil
	} else {
		__antithesis_instrumentation__.Notify(267244)
	}
	__antithesis_instrumentation__.Notify(267238)
	return "", errors.Newf("unable to resolve schema id %d for db %d", schemaID, db.GetID())
}

type SchemaEntryForDB struct {
	Name      string
	Timestamp hlc.Timestamp
}

func GetForDatabase(
	ctx context.Context, txn *kv.Txn, codec keys.SQLCodec, db catalog.DatabaseDescriptor,
) (map[descpb.ID]SchemaEntryForDB, error) {
	__antithesis_instrumentation__.Notify(267245)
	log.Eventf(ctx, "fetching all schema descriptor IDs for database %q (%d)", db.GetName(), db.GetID())

	nameKey := catalogkeys.MakeSchemaNameKey(codec, db.GetID(), "")
	kvs, err := txn.Scan(ctx, nameKey, nameKey.PrefixEnd(), 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(267249)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(267250)
	}
	__antithesis_instrumentation__.Notify(267246)

	ret := make(map[descpb.ID]SchemaEntryForDB, len(kvs)+1)

	if !db.HasPublicSchemaWithDescriptor() {
		__antithesis_instrumentation__.Notify(267251)
		ret[descpb.ID(keys.PublicSchemaID)] = SchemaEntryForDB{
			Name:      tree.PublicSchema,
			Timestamp: txn.ReadTimestamp(),
		}
	} else {
		__antithesis_instrumentation__.Notify(267252)
	}
	__antithesis_instrumentation__.Notify(267247)

	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(267253)
		id := descpb.ID(kv.ValueInt())
		if _, ok := ret[id]; ok {
			__antithesis_instrumentation__.Notify(267256)
			continue
		} else {
			__antithesis_instrumentation__.Notify(267257)
		}
		__antithesis_instrumentation__.Notify(267254)
		k, err := catalogkeys.DecodeNameMetadataKey(codec, kv.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(267258)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(267259)
		}
		__antithesis_instrumentation__.Notify(267255)
		ret[id] = SchemaEntryForDB{
			Name:      k.GetName(),
			Timestamp: kv.Value.Timestamp,
		}
	}
	__antithesis_instrumentation__.Notify(267248)
	return ret, nil
}

func ResolveExisting(
	ctx context.Context,
	u *tree.UnresolvedObjectName,
	r ObjectNameExistingResolver,
	lookupFlags tree.ObjectLookupFlags,
	curDb string,
	searchPath sessiondata.SearchPath,
) (found bool, prefix catalog.ResolvedObjectPrefix, result catalog.Descriptor, err error) {
	__antithesis_instrumentation__.Notify(267260)
	if u.HasExplicitSchema() {
		__antithesis_instrumentation__.Notify(267264)
		if u.HasExplicitCatalog() {
			__antithesis_instrumentation__.Notify(267268)

			found, prefix, result, err = r.LookupObject(ctx, lookupFlags, u.Catalog(), u.Schema(), u.Object())
			prefix.ExplicitDatabase, prefix.ExplicitSchema = true, true
			return found, prefix, result, err
		} else {
			__antithesis_instrumentation__.Notify(267269)
		}
		__antithesis_instrumentation__.Notify(267265)

		_, isVirtualSchema := catconstants.VirtualSchemaNames[u.Schema()]
		if isVirtualSchema || func() bool {
			__antithesis_instrumentation__.Notify(267270)
			return curDb != "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(267271)
			if found, prefix, result, err = r.LookupObject(
				ctx, lookupFlags, curDb, u.Schema(), u.Object(),
			); found || func() bool {
				__antithesis_instrumentation__.Notify(267272)
				return err != nil == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(267273)
				return isVirtualSchema == true
			}() == true {
				__antithesis_instrumentation__.Notify(267274)
				if !found && func() bool {
					__antithesis_instrumentation__.Notify(267276)
					return err == nil == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(267277)
					return prefix.Database == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(267278)

					err = sqlerrors.NewUndefinedDatabaseError(curDb)
				} else {
					__antithesis_instrumentation__.Notify(267279)
				}
				__antithesis_instrumentation__.Notify(267275)

				prefix.ExplicitDatabase = false
				prefix.ExplicitSchema = true
				return found, prefix, result, err
			} else {
				__antithesis_instrumentation__.Notify(267280)
			}
		} else {
			__antithesis_instrumentation__.Notify(267281)
		}
		__antithesis_instrumentation__.Notify(267266)

		found, prefix, result, err = r.LookupObject(ctx, lookupFlags, u.Schema(), tree.PublicSchema, u.Object())
		if found && func() bool {
			__antithesis_instrumentation__.Notify(267282)
			return err == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(267283)
			prefix.ExplicitSchema = true
			prefix.ExplicitDatabase = true
		} else {
			__antithesis_instrumentation__.Notify(267284)
		}
		__antithesis_instrumentation__.Notify(267267)
		return found, prefix, result, err
	} else {
		__antithesis_instrumentation__.Notify(267285)
	}
	__antithesis_instrumentation__.Notify(267261)

	iter := searchPath.Iter()
	foundDatabase := false
	for next, ok := iter.Next(); ok; next, ok = iter.Next() {
		__antithesis_instrumentation__.Notify(267286)
		if found, prefix, result, err = r.LookupObject(
			ctx, lookupFlags, curDb, next, u.Object(),
		); found || func() bool {
			__antithesis_instrumentation__.Notify(267288)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(267289)
			return found, prefix, result, err
		} else {
			__antithesis_instrumentation__.Notify(267290)
		}
		__antithesis_instrumentation__.Notify(267287)
		foundDatabase = foundDatabase || func() bool {
			__antithesis_instrumentation__.Notify(267291)
			return prefix.Database != nil == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(267262)

	if curDb != "" && func() bool {
		__antithesis_instrumentation__.Notify(267292)
		return !foundDatabase == true
	}() == true {
		__antithesis_instrumentation__.Notify(267293)
		return false, prefix, nil, sqlerrors.NewUndefinedDatabaseError(curDb)
	} else {
		__antithesis_instrumentation__.Notify(267294)
	}
	__antithesis_instrumentation__.Notify(267263)
	return false, prefix, nil, nil
}

func ResolveTarget(
	ctx context.Context,
	u *tree.UnresolvedObjectName,
	r ObjectNameTargetResolver,
	curDb string,
	searchPath sessiondata.SearchPath,
) (found bool, _ tree.ObjectNamePrefix, scMeta catalog.ResolvedObjectPrefix, err error) {
	__antithesis_instrumentation__.Notify(267295)
	if u.HasExplicitSchema() {
		__antithesis_instrumentation__.Notify(267298)
		if u.HasExplicitCatalog() {
			__antithesis_instrumentation__.Notify(267302)

			found, scMeta, err = r.LookupSchema(ctx, u.Catalog(), u.Schema())
			scMeta.ExplicitDatabase, scMeta.ExplicitSchema = true, true
			return found, scMeta.NamePrefix(), scMeta, err
		} else {
			__antithesis_instrumentation__.Notify(267303)
		}
		__antithesis_instrumentation__.Notify(267299)

		if found, scMeta, err = r.LookupSchema(ctx, curDb, u.Schema()); found || func() bool {
			__antithesis_instrumentation__.Notify(267304)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(267305)
			if err == nil {
				__antithesis_instrumentation__.Notify(267307)
				scMeta.ExplicitDatabase, scMeta.ExplicitSchema = false, true
			} else {
				__antithesis_instrumentation__.Notify(267308)
			}
			__antithesis_instrumentation__.Notify(267306)
			return found, scMeta.NamePrefix(), scMeta, err
		} else {
			__antithesis_instrumentation__.Notify(267309)
		}
		__antithesis_instrumentation__.Notify(267300)

		if found, scMeta, err = r.LookupSchema(ctx, u.Schema(), tree.PublicSchema); found || func() bool {
			__antithesis_instrumentation__.Notify(267310)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(267311)
			if err == nil {
				__antithesis_instrumentation__.Notify(267313)
				scMeta.ExplicitDatabase, scMeta.ExplicitSchema = true, true
			} else {
				__antithesis_instrumentation__.Notify(267314)
			}
			__antithesis_instrumentation__.Notify(267312)
			return found, scMeta.NamePrefix(), scMeta, err
		} else {
			__antithesis_instrumentation__.Notify(267315)
		}
		__antithesis_instrumentation__.Notify(267301)

		return false, tree.ObjectNamePrefix{}, catalog.ResolvedObjectPrefix{}, nil
	} else {
		__antithesis_instrumentation__.Notify(267316)
	}
	__antithesis_instrumentation__.Notify(267296)

	iter := searchPath.IterWithoutImplicitPGSchemas()
	for scName, ok := iter.Next(); ok; scName, ok = iter.Next() {
		__antithesis_instrumentation__.Notify(267317)
		if found, scMeta, err = r.LookupSchema(ctx, curDb, scName); found || func() bool {
			__antithesis_instrumentation__.Notify(267318)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(267319)
			if err == nil {
				__antithesis_instrumentation__.Notify(267321)
				scMeta.ExplicitDatabase, scMeta.ExplicitSchema = false, false
			} else {
				__antithesis_instrumentation__.Notify(267322)
			}
			__antithesis_instrumentation__.Notify(267320)
			break
		} else {
			__antithesis_instrumentation__.Notify(267323)
		}
	}
	__antithesis_instrumentation__.Notify(267297)
	return found, scMeta.NamePrefix(), scMeta, err
}

func ResolveObjectNamePrefix(
	ctx context.Context,
	r ObjectNameTargetResolver,
	curDb string,
	searchPath sessiondata.SearchPath,
	tp *tree.ObjectNamePrefix,
) (found bool, scMeta catalog.ResolvedObjectPrefix, err error) {
	__antithesis_instrumentation__.Notify(267324)
	if tp.ExplicitSchema {
		__antithesis_instrumentation__.Notify(267327)

		scName, err := searchPath.MaybeResolveTemporarySchema(tp.Schema())
		if err != nil {
			__antithesis_instrumentation__.Notify(267332)
			return false, catalog.ResolvedObjectPrefix{}, err
		} else {
			__antithesis_instrumentation__.Notify(267333)
		}
		__antithesis_instrumentation__.Notify(267328)
		if tp.ExplicitCatalog {
			__antithesis_instrumentation__.Notify(267334)

			tp.SchemaName = tree.Name(scName)
			return r.LookupSchema(ctx, tp.Catalog(), scName)
		} else {
			__antithesis_instrumentation__.Notify(267335)
		}
		__antithesis_instrumentation__.Notify(267329)

		if found, scMeta, err = r.LookupSchema(ctx, curDb, scName); found || func() bool {
			__antithesis_instrumentation__.Notify(267336)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(267337)
			if err == nil {
				__antithesis_instrumentation__.Notify(267339)
				tp.CatalogName = tree.Name(curDb)
				tp.SchemaName = tree.Name(scName)
			} else {
				__antithesis_instrumentation__.Notify(267340)
			}
			__antithesis_instrumentation__.Notify(267338)
			return found, scMeta, err
		} else {
			__antithesis_instrumentation__.Notify(267341)
		}
		__antithesis_instrumentation__.Notify(267330)

		if found, scMeta, err = r.LookupSchema(ctx, tp.Schema(), tree.PublicSchema); found || func() bool {
			__antithesis_instrumentation__.Notify(267342)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(267343)
			if err == nil {
				__antithesis_instrumentation__.Notify(267345)
				tp.CatalogName = tp.SchemaName
				tp.SchemaName = tree.PublicSchemaName
				tp.ExplicitCatalog = true
			} else {
				__antithesis_instrumentation__.Notify(267346)
			}
			__antithesis_instrumentation__.Notify(267344)
			return found, scMeta, err
		} else {
			__antithesis_instrumentation__.Notify(267347)
		}
		__antithesis_instrumentation__.Notify(267331)

		return false, catalog.ResolvedObjectPrefix{}, nil
	} else {
		__antithesis_instrumentation__.Notify(267348)
	}
	__antithesis_instrumentation__.Notify(267325)

	iter := searchPath.IterWithoutImplicitPGSchemas()
	for scName, ok := iter.Next(); ok; scName, ok = iter.Next() {
		__antithesis_instrumentation__.Notify(267349)
		if found, scMeta, err = r.LookupSchema(ctx, curDb, scName); found || func() bool {
			__antithesis_instrumentation__.Notify(267350)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(267351)
			if err == nil {
				__antithesis_instrumentation__.Notify(267353)
				tp.CatalogName = tree.Name(curDb)
				tp.SchemaName = tree.Name(scName)
			} else {
				__antithesis_instrumentation__.Notify(267354)
			}
			__antithesis_instrumentation__.Notify(267352)
			break
		} else {
			__antithesis_instrumentation__.Notify(267355)
		}
	}
	__antithesis_instrumentation__.Notify(267326)
	return found, scMeta, err
}
