package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/errors"
)

type AlterDefaultPrivileges struct {
	Roles RoleSpecList

	ForAllRoles bool

	Schemas ObjectNamePrefixList

	Database *Name

	IsGrant bool
	Grant   AbbreviatedGrant
	Revoke  AbbreviatedRevoke
}

func (n *AlterDefaultPrivileges) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602971)
	ctx.WriteString("ALTER DEFAULT PRIVILEGES ")
	if n.ForAllRoles {
		__antithesis_instrumentation__.Notify(602974)
		ctx.WriteString("FOR ALL ROLES ")
	} else {
		__antithesis_instrumentation__.Notify(602975)
		if len(n.Roles) > 0 {
			__antithesis_instrumentation__.Notify(602976)
			ctx.WriteString("FOR ROLE ")
			for i, role := range n.Roles {
				__antithesis_instrumentation__.Notify(602978)
				if i > 0 {
					__antithesis_instrumentation__.Notify(602980)
					ctx.WriteString(", ")
				} else {
					__antithesis_instrumentation__.Notify(602981)
				}
				__antithesis_instrumentation__.Notify(602979)
				ctx.FormatNode(&role)
			}
			__antithesis_instrumentation__.Notify(602977)
			ctx.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(602982)
		}
	}
	__antithesis_instrumentation__.Notify(602972)
	if len(n.Schemas) > 0 {
		__antithesis_instrumentation__.Notify(602983)
		ctx.WriteString("IN SCHEMA ")
		ctx.FormatNode(n.Schemas)
		ctx.WriteString(" ")
	} else {
		__antithesis_instrumentation__.Notify(602984)
	}
	__antithesis_instrumentation__.Notify(602973)
	if n.IsGrant {
		__antithesis_instrumentation__.Notify(602985)
		n.Grant.Format(ctx)
	} else {
		__antithesis_instrumentation__.Notify(602986)
		n.Revoke.Format(ctx)
	}
}

type AlterDefaultPrivilegesTargetObject uint32

func (t AlterDefaultPrivilegesTargetObject) ToPrivilegeObjectType() privilege.ObjectType {
	__antithesis_instrumentation__.Notify(602987)
	targetObjectToPrivilegeObject := map[AlterDefaultPrivilegesTargetObject]privilege.ObjectType{
		Tables:    privilege.Table,
		Sequences: privilege.Table,
		Schemas:   privilege.Schema,
		Types:     privilege.Type,
	}
	return targetObjectToPrivilegeObject[t]
}

const (
	Tables    AlterDefaultPrivilegesTargetObject = 1
	Sequences AlterDefaultPrivilegesTargetObject = 2
	Types     AlterDefaultPrivilegesTargetObject = 3
	Schemas   AlterDefaultPrivilegesTargetObject = 4
)

func GetAlterDefaultPrivilegesTargetObjects() []AlterDefaultPrivilegesTargetObject {
	__antithesis_instrumentation__.Notify(602988)
	return []AlterDefaultPrivilegesTargetObject{
		Tables,
		Sequences,
		Types,
		Schemas,
	}
}

func (t AlterDefaultPrivilegesTargetObject) String() string {
	__antithesis_instrumentation__.Notify(602989)
	switch t {
	case Tables:
		__antithesis_instrumentation__.Notify(602990)
		return "tables"
	case Sequences:
		__antithesis_instrumentation__.Notify(602991)
		return "sequences"
	case Types:
		__antithesis_instrumentation__.Notify(602992)
		return "types"
	case Schemas:
		__antithesis_instrumentation__.Notify(602993)
		return "schemas"
	default:
		__antithesis_instrumentation__.Notify(602994)
		panic(errors.AssertionFailedf("unknown AlterDefaultPrivilegesTargetObject value: %d", t))
	}
}

type AbbreviatedGrant struct {
	Privileges      privilege.List
	Target          AlterDefaultPrivilegesTargetObject
	Grantees        RoleSpecList
	WithGrantOption bool
}

func (n *AbbreviatedGrant) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602995)
	ctx.WriteString("GRANT ")
	n.Privileges.Format(&ctx.Buffer)
	ctx.WriteString(" ON ")
	switch n.Target {
	case Tables:
		__antithesis_instrumentation__.Notify(602997)
		ctx.WriteString("TABLES ")
	case Sequences:
		__antithesis_instrumentation__.Notify(602998)
		ctx.WriteString("SEQUENCES ")
	case Types:
		__antithesis_instrumentation__.Notify(602999)
		ctx.WriteString("TYPES ")
	case Schemas:
		__antithesis_instrumentation__.Notify(603000)
		ctx.WriteString("SCHEMAS ")
	default:
		__antithesis_instrumentation__.Notify(603001)
	}
	__antithesis_instrumentation__.Notify(602996)
	ctx.WriteString("TO ")
	n.Grantees.Format(ctx)
	if n.WithGrantOption {
		__antithesis_instrumentation__.Notify(603002)
		ctx.WriteString(" WITH GRANT OPTION")
	} else {
		__antithesis_instrumentation__.Notify(603003)
	}
}

type AbbreviatedRevoke struct {
	Privileges     privilege.List
	Target         AlterDefaultPrivilegesTargetObject
	Grantees       RoleSpecList
	GrantOptionFor bool
}

func (n *AbbreviatedRevoke) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603004)
	ctx.WriteString("REVOKE ")
	if n.GrantOptionFor {
		__antithesis_instrumentation__.Notify(603007)
		ctx.WriteString("GRANT OPTION FOR ")
	} else {
		__antithesis_instrumentation__.Notify(603008)
	}
	__antithesis_instrumentation__.Notify(603005)
	n.Privileges.Format(&ctx.Buffer)
	ctx.WriteString(" ON ")
	switch n.Target {
	case Tables:
		__antithesis_instrumentation__.Notify(603009)
		ctx.WriteString("TABLES ")
	case Sequences:
		__antithesis_instrumentation__.Notify(603010)
		ctx.WriteString("SEQUENCES ")
	case Types:
		__antithesis_instrumentation__.Notify(603011)
		ctx.WriteString("TYPES ")
	case Schemas:
		__antithesis_instrumentation__.Notify(603012)
		ctx.WriteString("SCHEMAS ")
	default:
		__antithesis_instrumentation__.Notify(603013)
	}
	__antithesis_instrumentation__.Notify(603006)
	ctx.WriteString(" FROM ")
	n.Grantees.Format(ctx)
}
