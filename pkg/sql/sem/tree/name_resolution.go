package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

func classifyTablePattern(n *UnresolvedName) (TablePattern, error) {
	__antithesis_instrumentation__.Notify(610451)
	if n.NumParts < 1 || func() bool {
		__antithesis_instrumentation__.Notify(610457)
		return n.NumParts > 3 == true
	}() == true {
		__antithesis_instrumentation__.Notify(610458)
		return nil, newInvTableNameError(n)
	} else {
		__antithesis_instrumentation__.Notify(610459)
	}
	__antithesis_instrumentation__.Notify(610452)

	firstCheck := 0
	if n.Star {
		__antithesis_instrumentation__.Notify(610460)
		firstCheck = 1
	} else {
		__antithesis_instrumentation__.Notify(610461)
	}
	__antithesis_instrumentation__.Notify(610453)

	lastCheck := n.NumParts
	if lastCheck > 2 {
		__antithesis_instrumentation__.Notify(610462)
		lastCheck = 2
	} else {
		__antithesis_instrumentation__.Notify(610463)
	}
	__antithesis_instrumentation__.Notify(610454)
	for i := firstCheck; i < lastCheck; i++ {
		__antithesis_instrumentation__.Notify(610464)
		if len(n.Parts[i]) == 0 {
			__antithesis_instrumentation__.Notify(610465)
			return nil, newInvTableNameError(n)
		} else {
			__antithesis_instrumentation__.Notify(610466)
		}
	}
	__antithesis_instrumentation__.Notify(610455)

	if n.Star {
		__antithesis_instrumentation__.Notify(610467)
		return &AllTablesSelector{makeObjectNamePrefixFromUnresolvedName(n)}, nil
	} else {
		__antithesis_instrumentation__.Notify(610468)
	}
	__antithesis_instrumentation__.Notify(610456)
	tb := makeTableNameFromUnresolvedName(n)
	return &tb, nil
}

func classifyColumnItem(n *UnresolvedName) (VarName, error) {
	__antithesis_instrumentation__.Notify(610469)
	if n.NumParts < 1 || func() bool {
		__antithesis_instrumentation__.Notify(610476)
		return n.NumParts > 4 == true
	}() == true {
		__antithesis_instrumentation__.Notify(610477)
		return nil, newInvColRef(n)
	} else {
		__antithesis_instrumentation__.Notify(610478)
	}
	__antithesis_instrumentation__.Notify(610470)

	firstCheck := 0
	if n.Star {
		__antithesis_instrumentation__.Notify(610479)
		firstCheck = 1
	} else {
		__antithesis_instrumentation__.Notify(610480)
	}
	__antithesis_instrumentation__.Notify(610471)

	lastCheck := n.NumParts
	if lastCheck > 3 {
		__antithesis_instrumentation__.Notify(610481)
		lastCheck = 3
	} else {
		__antithesis_instrumentation__.Notify(610482)
	}
	__antithesis_instrumentation__.Notify(610472)
	for i := firstCheck; i < lastCheck; i++ {
		__antithesis_instrumentation__.Notify(610483)
		if len(n.Parts[i]) == 0 {
			__antithesis_instrumentation__.Notify(610484)
			return nil, newInvColRef(n)
		} else {
			__antithesis_instrumentation__.Notify(610485)
		}
	}
	__antithesis_instrumentation__.Notify(610473)

	var tn *UnresolvedObjectName
	if n.NumParts > 1 {
		__antithesis_instrumentation__.Notify(610486)
		var err error
		tn, err = NewUnresolvedObjectName(
			n.NumParts-1,
			[3]string{n.Parts[1], n.Parts[2], n.Parts[3]},
			NoAnnotation,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(610487)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(610488)
		}
	} else {
		__antithesis_instrumentation__.Notify(610489)
	}
	__antithesis_instrumentation__.Notify(610474)
	if n.Star {
		__antithesis_instrumentation__.Notify(610490)
		return &AllColumnsSelector{TableName: tn}, nil
	} else {
		__antithesis_instrumentation__.Notify(610491)
	}
	__antithesis_instrumentation__.Notify(610475)
	return &ColumnItem{TableName: tn, ColumnName: Name(n.Parts[0])}, nil
}

const (
	PublicSchema string = catconstants.PublicSchemaName

	PublicSchemaName Name = Name(PublicSchema)
)

type QualifiedNameResolver interface {
	GetQualifiedTableNameByID(ctx context.Context, id int64, requiredType RequiredTableKind) (*TableName, error)
	CurrentDatabase() string
}

func (n *UnresolvedName) ResolveFunction(
	searchPath sessiondata.SearchPath,
) (*FunctionDefinition, error) {
	__antithesis_instrumentation__.Notify(610492)
	if n.NumParts > 3 || func() bool {
		__antithesis_instrumentation__.Notify(610499)
		return len(n.Parts[0]) == 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(610500)
		return n.Star == true
	}() == true {
		__antithesis_instrumentation__.Notify(610501)

		return nil, pgerror.Newf(pgcode.InvalidName,
			"invalid function name: %s", n)
	} else {
		__antithesis_instrumentation__.Notify(610502)
	}
	__antithesis_instrumentation__.Notify(610493)

	function, prefix := n.Parts[0], n.Parts[1]

	if d, ok := FunDefs[function]; ok && func() bool {
		__antithesis_instrumentation__.Notify(610503)
		return prefix == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(610504)

		return d, nil
	} else {
		__antithesis_instrumentation__.Notify(610505)
	}
	__antithesis_instrumentation__.Notify(610494)

	fullName := function

	if prefix == catconstants.PgCatalogName {
		__antithesis_instrumentation__.Notify(610506)

		prefix = ""
	} else {
		__antithesis_instrumentation__.Notify(610507)
	}
	__antithesis_instrumentation__.Notify(610495)
	if prefix == catconstants.PublicSchemaName {
		__antithesis_instrumentation__.Notify(610508)

		if d, ok := FunDefs[function]; ok && func() bool {
			__antithesis_instrumentation__.Notify(610509)
			return d.AvailableOnPublicSchema == true
		}() == true {
			__antithesis_instrumentation__.Notify(610510)
			return d, nil
		} else {
			__antithesis_instrumentation__.Notify(610511)
		}
	} else {
		__antithesis_instrumentation__.Notify(610512)
	}
	__antithesis_instrumentation__.Notify(610496)

	if prefix != "" {
		__antithesis_instrumentation__.Notify(610513)
		fullName = prefix + "." + function
	} else {
		__antithesis_instrumentation__.Notify(610514)
	}
	__antithesis_instrumentation__.Notify(610497)
	def, ok := FunDefs[fullName]
	if !ok {
		__antithesis_instrumentation__.Notify(610515)
		found := false
		if prefix == "" {
			__antithesis_instrumentation__.Notify(610517)

			iter := searchPath.Iter()
			for alt, ok := iter.Next(); ok; alt, ok = iter.Next() {
				__antithesis_instrumentation__.Notify(610518)
				fullName = alt + "." + function
				if def, ok = FunDefs[fullName]; ok {
					__antithesis_instrumentation__.Notify(610519)
					found = true
					break
				} else {
					__antithesis_instrumentation__.Notify(610520)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(610521)
		}
		__antithesis_instrumentation__.Notify(610516)
		if !found {
			__antithesis_instrumentation__.Notify(610522)
			extraMsg := ""

			if rdef, ok := FunDefs[strings.ToLower(function)]; ok {
				__antithesis_instrumentation__.Notify(610524)
				extraMsg = fmt.Sprintf(", but %s() exists", rdef.Name)
			} else {
				__antithesis_instrumentation__.Notify(610525)
			}
			__antithesis_instrumentation__.Notify(610523)
			return nil, pgerror.Newf(
				pgcode.UndefinedFunction, "unknown function: %s()%s", ErrString(n), extraMsg)
		} else {
			__antithesis_instrumentation__.Notify(610526)
		}
	} else {
		__antithesis_instrumentation__.Notify(610527)
	}
	__antithesis_instrumentation__.Notify(610498)

	return def, nil
}

func newInvColRef(n *UnresolvedName) error {
	__antithesis_instrumentation__.Notify(610528)
	return pgerror.NewWithDepthf(1, pgcode.InvalidColumnReference,
		"invalid column name: %s", n)
}

func newInvTableNameError(n fmt.Stringer) error {
	__antithesis_instrumentation__.Notify(610529)
	return pgerror.NewWithDepthf(1, pgcode.InvalidName,
		"invalid table name: %s", n)
}

type CommonLookupFlags struct {
	Required bool

	RequireMutable bool

	AvoidLeased bool

	IncludeOffline bool

	IncludeDropped bool

	AvoidSynthetic bool
}

type SchemaLookupFlags = CommonLookupFlags

type DatabaseLookupFlags = CommonLookupFlags

type DatabaseListFlags struct {
	CommonLookupFlags

	ExplicitPrefix bool
}

type DesiredObjectKind int

const (
	TableObject DesiredObjectKind = iota

	TypeObject
)

func NewQualifiedObjectName(catalog, schema, object string, kind DesiredObjectKind) ObjectName {
	__antithesis_instrumentation__.Notify(610530)
	switch kind {
	case TableObject:
		__antithesis_instrumentation__.Notify(610532)
		name := MakeTableNameWithSchema(Name(catalog), Name(schema), Name(object))
		return &name
	case TypeObject:
		__antithesis_instrumentation__.Notify(610533)
		name := MakeQualifiedTypeName(catalog, schema, object)
		return &name
	default:
		__antithesis_instrumentation__.Notify(610534)
	}
	__antithesis_instrumentation__.Notify(610531)
	return nil
}

type RequiredTableKind int

const (
	ResolveAnyTableKind RequiredTableKind = iota
	ResolveRequireTableDesc
	ResolveRequireViewDesc
	ResolveRequireTableOrViewDesc
	ResolveRequireSequenceDesc
)

var requiredTypeNames = [...]string{
	ResolveAnyTableKind:           "any",
	ResolveRequireTableDesc:       "table",
	ResolveRequireViewDesc:        "view",
	ResolveRequireTableOrViewDesc: "table or view",
	ResolveRequireSequenceDesc:    "sequence",
}

func (r RequiredTableKind) String() string {
	__antithesis_instrumentation__.Notify(610535)
	return requiredTypeNames[r]
}

type ObjectLookupFlags struct {
	CommonLookupFlags
	AllowWithoutPrimaryKey bool

	DesiredObjectKind DesiredObjectKind

	DesiredTableDescKind RequiredTableKind
}

func ObjectLookupFlagsWithRequired() ObjectLookupFlags {
	__antithesis_instrumentation__.Notify(610536)
	return ObjectLookupFlags{
		CommonLookupFlags: CommonLookupFlags{Required: true},
	}
}

func ObjectLookupFlagsWithRequiredTableKind(kind RequiredTableKind) ObjectLookupFlags {
	__antithesis_instrumentation__.Notify(610537)
	return ObjectLookupFlags{
		CommonLookupFlags:    CommonLookupFlags{Required: true},
		DesiredObjectKind:    TableObject,
		DesiredTableDescKind: kind,
	}
}
