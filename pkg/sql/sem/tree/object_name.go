package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type ObjectName interface {
	NodeFormatter
	Object() string
	Schema() string
	Catalog() string
	FQString() string
	objectName()
}

var _ ObjectName = &TableName{}
var _ ObjectName = &TypeName{}

type objName struct {
	ObjectName Name

	ObjectNamePrefix
}

func (o *objName) Object() string {
	__antithesis_instrumentation__.Notify(610963)
	return string(o.ObjectName)
}

func (o *objName) ToUnresolvedObjectName() *UnresolvedObjectName {
	__antithesis_instrumentation__.Notify(610964)
	u := &UnresolvedObjectName{}

	u.NumParts = 1
	u.Parts[0] = string(o.ObjectName)
	if o.ExplicitSchema {
		__antithesis_instrumentation__.Notify(610967)
		u.Parts[u.NumParts] = string(o.SchemaName)
		u.NumParts++
	} else {
		__antithesis_instrumentation__.Notify(610968)
	}
	__antithesis_instrumentation__.Notify(610965)
	if o.ExplicitCatalog {
		__antithesis_instrumentation__.Notify(610969)
		u.Parts[u.NumParts] = string(o.CatalogName)
		u.NumParts++
	} else {
		__antithesis_instrumentation__.Notify(610970)
	}
	__antithesis_instrumentation__.Notify(610966)
	return u
}

type ObjectNamePrefix struct {
	CatalogName Name
	SchemaName  Name

	ExplicitCatalog bool

	ExplicitSchema bool
}

func (tp *ObjectNamePrefix) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(610971)
	alwaysFormat := ctx.alwaysFormatTablePrefix()
	if tp.ExplicitSchema || func() bool {
		__antithesis_instrumentation__.Notify(610972)
		return alwaysFormat == true
	}() == true {
		__antithesis_instrumentation__.Notify(610973)
		if tp.ExplicitCatalog || func() bool {
			__antithesis_instrumentation__.Notify(610975)
			return alwaysFormat == true
		}() == true {
			__antithesis_instrumentation__.Notify(610976)
			ctx.FormatNode(&tp.CatalogName)
			ctx.WriteByte('.')
		} else {
			__antithesis_instrumentation__.Notify(610977)
		}
		__antithesis_instrumentation__.Notify(610974)
		ctx.FormatNode(&tp.SchemaName)
	} else {
		__antithesis_instrumentation__.Notify(610978)
	}
}

func (tp *ObjectNamePrefix) String() string {
	__antithesis_instrumentation__.Notify(610979)
	return AsString(tp)
}

func (tp *ObjectNamePrefix) Schema() string {
	__antithesis_instrumentation__.Notify(610980)
	return string(tp.SchemaName)
}

func (tp *ObjectNamePrefix) Catalog() string {
	__antithesis_instrumentation__.Notify(610981)
	return string(tp.CatalogName)
}

type ObjectNamePrefixList []ObjectNamePrefix

func (tp ObjectNamePrefixList) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(610982)
	for idx, objectNamePrefix := range tp {
		__antithesis_instrumentation__.Notify(610983)
		ctx.FormatNode(&objectNamePrefix)
		if idx != len(tp)-1 {
			__antithesis_instrumentation__.Notify(610984)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(610985)
		}
	}
}

type UnresolvedObjectName struct {
	NumParts int

	Parts [3]string

	AnnotatedNode
}

func (*UnresolvedObjectName) tableExpr() { __antithesis_instrumentation__.Notify(610986) }

func NewUnresolvedObjectName(
	numParts int, parts [3]string, annotationIdx AnnotationIdx,
) (*UnresolvedObjectName, error) {
	__antithesis_instrumentation__.Notify(610987)
	u := &UnresolvedObjectName{
		NumParts:      numParts,
		Parts:         parts,
		AnnotatedNode: AnnotatedNode{AnnIdx: annotationIdx},
	}
	if u.NumParts < 1 {
		__antithesis_instrumentation__.Notify(610991)
		return nil, newInvTableNameError(u)
	} else {
		__antithesis_instrumentation__.Notify(610992)
	}
	__antithesis_instrumentation__.Notify(610988)

	lastCheck := u.NumParts
	if lastCheck > 2 {
		__antithesis_instrumentation__.Notify(610993)
		lastCheck = 2
	} else {
		__antithesis_instrumentation__.Notify(610994)
	}
	__antithesis_instrumentation__.Notify(610989)
	for i := 0; i < lastCheck; i++ {
		__antithesis_instrumentation__.Notify(610995)
		if len(u.Parts[i]) == 0 {
			__antithesis_instrumentation__.Notify(610996)
			return nil, newInvTableNameError(u)
		} else {
			__antithesis_instrumentation__.Notify(610997)
		}
	}
	__antithesis_instrumentation__.Notify(610990)
	return u, nil
}

func (u *UnresolvedObjectName) Resolved(ann *Annotations) ObjectName {
	__antithesis_instrumentation__.Notify(610998)
	r := u.GetAnnotation(ann)
	if r == nil {
		__antithesis_instrumentation__.Notify(611000)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(611001)
	}
	__antithesis_instrumentation__.Notify(610999)
	return r.(ObjectName)
}

func (u *UnresolvedObjectName) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(611002)

	if ctx.HasFlags(FmtAlwaysQualifyTableNames) || func() bool {
		__antithesis_instrumentation__.Notify(611004)
		return ctx.tableNameFormatter != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(611005)
		if ctx.tableNameFormatter != nil && func() bool {
			__antithesis_instrumentation__.Notify(611007)
			return ctx.ann == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(611008)

			tn := u.ToTableName()
			tn.Format(ctx)
			return
		} else {
			__antithesis_instrumentation__.Notify(611009)
		}
		__antithesis_instrumentation__.Notify(611006)

		if n := u.Resolved(ctx.ann); n != nil {
			__antithesis_instrumentation__.Notify(611010)
			n.Format(ctx)
			return
		} else {
			__antithesis_instrumentation__.Notify(611011)
		}
	} else {
		__antithesis_instrumentation__.Notify(611012)
	}
	__antithesis_instrumentation__.Notify(611003)

	for i := u.NumParts; i > 0; i-- {
		__antithesis_instrumentation__.Notify(611013)

		if i == u.NumParts {
			__antithesis_instrumentation__.Notify(611014)
			ctx.FormatNode((*Name)(&u.Parts[i-1]))
		} else {
			__antithesis_instrumentation__.Notify(611015)
			ctx.WriteByte('.')
			ctx.FormatNode((*UnrestrictedName)(&u.Parts[i-1]))
		}
	}
}

func (u *UnresolvedObjectName) String() string {
	__antithesis_instrumentation__.Notify(611016)
	return AsString(u)
}

func (u *UnresolvedObjectName) ToTableName() TableName {
	__antithesis_instrumentation__.Notify(611017)
	return TableName{objName{
		ObjectName: Name(u.Parts[0]),
		ObjectNamePrefix: ObjectNamePrefix{
			SchemaName:      Name(u.Parts[1]),
			CatalogName:     Name(u.Parts[2]),
			ExplicitSchema:  u.NumParts >= 2,
			ExplicitCatalog: u.NumParts >= 3,
		},
	}}
}

func (u *UnresolvedObjectName) ToUnresolvedName() *UnresolvedName {
	__antithesis_instrumentation__.Notify(611018)
	return &UnresolvedName{
		NumParts: u.NumParts,
		Parts:    NameParts{u.Parts[0], u.Parts[1], u.Parts[2]},
	}
}

func (u *UnresolvedObjectName) Object() string {
	__antithesis_instrumentation__.Notify(611019)
	return u.Parts[0]
}

func (u *UnresolvedObjectName) Schema() string {
	__antithesis_instrumentation__.Notify(611020)
	return u.Parts[1]
}

func (u *UnresolvedObjectName) Catalog() string {
	__antithesis_instrumentation__.Notify(611021)
	return u.Parts[2]
}

func (u *UnresolvedObjectName) HasExplicitSchema() bool {
	__antithesis_instrumentation__.Notify(611022)
	return u.NumParts >= 2
}

func (u *UnresolvedObjectName) HasExplicitCatalog() bool {
	__antithesis_instrumentation__.Notify(611023)
	return u.NumParts >= 3
}
