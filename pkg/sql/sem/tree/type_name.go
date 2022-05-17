package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type TypeName struct {
	objName
}

var _ = (*TypeName).Type
var _ = (*TypeName).FQString
var _ = NewUnqualifiedTypeName

func (t *TypeName) Type() string {
	__antithesis_instrumentation__.Notify(615777)
	return string(t.ObjectName)
}

func (t *TypeName) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(615778)
	ctx.FormatNode(&t.ObjectNamePrefix)
	if t.ExplicitSchema || func() bool {
		__antithesis_instrumentation__.Notify(615780)
		return ctx.alwaysFormatTablePrefix() == true
	}() == true {
		__antithesis_instrumentation__.Notify(615781)
		ctx.WriteByte('.')
	} else {
		__antithesis_instrumentation__.Notify(615782)
	}
	__antithesis_instrumentation__.Notify(615779)
	ctx.FormatNode(&t.ObjectName)
}

func (t *TypeName) String() string {
	__antithesis_instrumentation__.Notify(615783)
	return AsString(t)
}

func (t *TypeName) SQLString() string {
	__antithesis_instrumentation__.Notify(615784)

	return AsStringWithFlags(t, FmtBareIdentifiers)
}

func (t *TypeName) FQString() string {
	__antithesis_instrumentation__.Notify(615785)
	ctx := NewFmtCtx(FmtSimple)
	ctx.FormatNode(&t.CatalogName)
	ctx.WriteByte('.')
	ctx.FormatNode(&t.SchemaName)
	ctx.WriteByte('.')
	ctx.FormatNode(&t.ObjectName)
	return ctx.CloseAndGetString()
}

func (t *TypeName) objectName() { __antithesis_instrumentation__.Notify(615786) }

func NewUnqualifiedTypeName(typ string) *TypeName {
	__antithesis_instrumentation__.Notify(615787)
	tn := MakeUnqualifiedTypeName(typ)
	return &tn
}

func MakeUnqualifiedTypeName(typ string) TypeName {
	__antithesis_instrumentation__.Notify(615788)
	return MakeTypeNameWithPrefix(ObjectNamePrefix{}, typ)
}

func MakeSchemaQualifiedTypeName(schema, typ string) TypeName {
	__antithesis_instrumentation__.Notify(615789)
	return MakeTypeNameWithPrefix(ObjectNamePrefix{
		ExplicitSchema: true,
		SchemaName:     Name(schema),
	}, typ)
}

func MakeTypeNameWithPrefix(prefix ObjectNamePrefix, typ string) TypeName {
	__antithesis_instrumentation__.Notify(615790)
	return TypeName{objName{
		ObjectNamePrefix: prefix,
		ObjectName:       Name(typ),
	}}
}

func MakeQualifiedTypeName(db, schema, typ string) TypeName {
	__antithesis_instrumentation__.Notify(615791)
	return MakeTypeNameWithPrefix(ObjectNamePrefix{
		ExplicitCatalog: true,
		CatalogName:     Name(db),
		ExplicitSchema:  true,
		SchemaName:      Name(schema),
	}, typ)
}

func NewQualifiedTypeName(db, schema, typ string) *TypeName {
	__antithesis_instrumentation__.Notify(615792)
	tn := MakeQualifiedTypeName(db, schema, typ)
	return &tn
}

type TypeReferenceResolver interface {
	ResolveType(ctx context.Context, name *UnresolvedObjectName) (*types.T, error)
	ResolveTypeByOID(ctx context.Context, oid oid.Oid) (*types.T, error)
}

type ResolvableTypeReference interface {
	SQLString() string
}

var _ ResolvableTypeReference = &UnresolvedObjectName{}
var _ ResolvableTypeReference = &ArrayTypeReference{}
var _ ResolvableTypeReference = &types.T{}
var _ ResolvableTypeReference = &OIDTypeReference{}
var _ NodeFormatter = &UnresolvedName{}
var _ NodeFormatter = &ArrayTypeReference{}

func ResolveType(
	ctx context.Context, ref ResolvableTypeReference, resolver TypeReferenceResolver,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(615793)
	switch t := ref.(type) {
	case *types.T:
		__antithesis_instrumentation__.Notify(615794)
		return t, nil
	case *ArrayTypeReference:
		__antithesis_instrumentation__.Notify(615795)
		typ, err := ResolveType(ctx, t.ElementType, resolver)
		if err != nil {
			__antithesis_instrumentation__.Notify(615802)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(615803)
		}
		__antithesis_instrumentation__.Notify(615796)
		return types.MakeArray(typ), nil
	case *UnresolvedObjectName:
		__antithesis_instrumentation__.Notify(615797)
		if resolver == nil {
			__antithesis_instrumentation__.Notify(615804)

			return nil, pgerror.Newf(pgcode.UndefinedObject, "type %q does not exist", t)
		} else {
			__antithesis_instrumentation__.Notify(615805)
		}
		__antithesis_instrumentation__.Notify(615798)
		return resolver.ResolveType(ctx, t)
	case *OIDTypeReference:
		__antithesis_instrumentation__.Notify(615799)
		if resolver == nil {
			__antithesis_instrumentation__.Notify(615806)
			return nil, pgerror.Newf(pgcode.UndefinedObject, "type resolver unavailable to resolve type OID %d", t.OID)
		} else {
			__antithesis_instrumentation__.Notify(615807)
		}
		__antithesis_instrumentation__.Notify(615800)
		return resolver.ResolveTypeByOID(ctx, t.OID)
	default:
		__antithesis_instrumentation__.Notify(615801)
		return nil, errors.AssertionFailedf("unknown resolvable type reference type %s", t)
	}
}

func (ctx *FmtCtx) FormatTypeReference(ref ResolvableTypeReference) {
	__antithesis_instrumentation__.Notify(615808)
	switch t := ref.(type) {
	case *types.T:
		__antithesis_instrumentation__.Notify(615809)
		if t.UserDefined() {
			__antithesis_instrumentation__.Notify(615815)
			if ctx.HasFlags(FmtAnonymize) {
				__antithesis_instrumentation__.Notify(615816)
				ctx.WriteByte('_')
				return
			} else {
				__antithesis_instrumentation__.Notify(615817)
				if ctx.HasFlags(fmtStaticallyFormatUserDefinedTypes) {
					__antithesis_instrumentation__.Notify(615818)
					idRef := OIDTypeReference{OID: t.Oid()}
					ctx.WriteString(idRef.SQLString())
					return
				} else {
					__antithesis_instrumentation__.Notify(615819)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(615820)
		}
		__antithesis_instrumentation__.Notify(615810)
		ctx.WriteString(t.SQLString())

	case *OIDTypeReference:
		__antithesis_instrumentation__.Notify(615811)
		if ctx.indexedTypeFormatter != nil {
			__antithesis_instrumentation__.Notify(615821)
			ctx.indexedTypeFormatter(ctx, t)
			return
		} else {
			__antithesis_instrumentation__.Notify(615822)
		}
		__antithesis_instrumentation__.Notify(615812)
		ctx.WriteString(ref.SQLString())

	case NodeFormatter:
		__antithesis_instrumentation__.Notify(615813)
		ctx.FormatNode(t)

	default:
		__antithesis_instrumentation__.Notify(615814)
		panic(errors.AssertionFailedf("type reference must implement NodeFormatter"))
	}
}

func GetStaticallyKnownType(ref ResolvableTypeReference) (typ *types.T, ok bool) {
	__antithesis_instrumentation__.Notify(615823)
	typ, ok = ref.(*types.T)
	return typ, ok
}

func MustBeStaticallyKnownType(ref ResolvableTypeReference) *types.T {
	__antithesis_instrumentation__.Notify(615824)
	if typ, ok := ref.(*types.T); ok {
		__antithesis_instrumentation__.Notify(615826)
		return typ
	} else {
		__antithesis_instrumentation__.Notify(615827)
	}
	__antithesis_instrumentation__.Notify(615825)
	panic(errors.AssertionFailedf("type reference was not a statically known type"))
}

type OIDTypeReference struct {
	OID oid.Oid
}

func (node *OIDTypeReference) SQLString() string {
	__antithesis_instrumentation__.Notify(615828)
	return fmt.Sprintf("@%d", node.OID)
}

type ArrayTypeReference struct {
	ElementType ResolvableTypeReference
}

func (node *ArrayTypeReference) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(615829)
	if typ, ok := GetStaticallyKnownType(node.ElementType); ok {
		__antithesis_instrumentation__.Notify(615830)
		ctx.FormatTypeReference(types.MakeArray(typ))
	} else {
		__antithesis_instrumentation__.Notify(615831)
		ctx.FormatTypeReference(node.ElementType)
		ctx.WriteString("[]")
	}
}

func (node *ArrayTypeReference) SQLString() string {
	__antithesis_instrumentation__.Notify(615832)

	return AsStringWithFlags(node, FmtBareIdentifiers)
}

func (name *UnresolvedObjectName) SQLString() string {
	__antithesis_instrumentation__.Notify(615833)

	return AsStringWithFlags(name, FmtBareIdentifiers)
}

func IsReferenceSerialType(ref ResolvableTypeReference) bool {
	__antithesis_instrumentation__.Notify(615834)
	if typ, ok := GetStaticallyKnownType(ref); ok {
		__antithesis_instrumentation__.Notify(615836)
		return types.IsSerialType(typ)
	} else {
		__antithesis_instrumentation__.Notify(615837)
	}
	__antithesis_instrumentation__.Notify(615835)
	return false
}

type TypeCollectorVisitor struct {
	OIDs map[oid.Oid]struct{}
}

func (v *TypeCollectorVisitor) VisitPre(expr Expr) (bool, Expr) {
	__antithesis_instrumentation__.Notify(615838)
	switch t := expr.(type) {
	case Datum:
		__antithesis_instrumentation__.Notify(615840)
		if t.ResolvedType().UserDefined() {
			__antithesis_instrumentation__.Notify(615844)
			v.OIDs[t.ResolvedType().Oid()] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(615845)
		}
	case *IsOfTypeExpr:
		__antithesis_instrumentation__.Notify(615841)
		for _, ref := range t.Types {
			__antithesis_instrumentation__.Notify(615846)
			if idref, ok := ref.(*OIDTypeReference); ok {
				__antithesis_instrumentation__.Notify(615847)
				v.OIDs[idref.OID] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(615848)
			}
		}
	case *CastExpr:
		__antithesis_instrumentation__.Notify(615842)
		if idref, ok := t.Type.(*OIDTypeReference); ok {
			__antithesis_instrumentation__.Notify(615849)
			v.OIDs[idref.OID] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(615850)
		}
	case *AnnotateTypeExpr:
		__antithesis_instrumentation__.Notify(615843)
		if idref, ok := t.Type.(*OIDTypeReference); ok {
			__antithesis_instrumentation__.Notify(615851)
			v.OIDs[idref.OID] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(615852)
		}
	}
	__antithesis_instrumentation__.Notify(615839)
	return true, expr
}

func (v *TypeCollectorVisitor) VisitPost(e Expr) Expr {
	__antithesis_instrumentation__.Notify(615853)
	return e
}

type TestingMapTypeResolver struct {
	typeMap map[string]*types.T
}

func (dtr *TestingMapTypeResolver) ResolveType(
	_ context.Context, name *UnresolvedObjectName,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(615854)
	typ, ok := dtr.typeMap[name.String()]
	if !ok {
		__antithesis_instrumentation__.Notify(615856)
		return nil, errors.Newf("type %q does not exist", name)
	} else {
		__antithesis_instrumentation__.Notify(615857)
	}
	__antithesis_instrumentation__.Notify(615855)
	return typ, nil
}

func (dtr *TestingMapTypeResolver) ResolveTypeByOID(context.Context, oid.Oid) (*types.T, error) {
	__antithesis_instrumentation__.Notify(615858)
	return nil, errors.AssertionFailedf("unimplemented")
}

func MakeTestingMapTypeResolver(typeMap map[string]*types.T) TypeReferenceResolver {
	__antithesis_instrumentation__.Notify(615859)
	return &TestingMapTypeResolver{typeMap: typeMap}
}
