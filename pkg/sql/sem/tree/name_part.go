package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
)

type Name string

func (n *Name) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(610391)
	f := ctx.flags
	if f.HasFlags(FmtAnonymize) && func() bool {
		__antithesis_instrumentation__.Notify(610392)
		return !isArityIndicatorString(string(*n)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(610393)
		ctx.WriteByte('_')
	} else {
		__antithesis_instrumentation__.Notify(610394)
		lexbase.EncodeRestrictedSQLIdent(&ctx.Buffer, string(*n), f.EncodeFlags())
	}
}

func NameStringP(s *string) string {
	__antithesis_instrumentation__.Notify(610395)
	return ((*Name)(s)).String()
}

func NameString(s string) string {
	__antithesis_instrumentation__.Notify(610396)
	return ((*Name)(&s)).String()
}

func ErrNameStringP(s *string) string {
	__antithesis_instrumentation__.Notify(610397)
	return ErrString(((*Name)(s)))
}

func ErrNameString(s string) string {
	__antithesis_instrumentation__.Notify(610398)
	return ErrString(((*Name)(&s)))
}

func (n Name) Normalize() string {
	__antithesis_instrumentation__.Notify(610399)
	return lexbase.NormalizeName(string(n))
}

type UnrestrictedName string

func (u *UnrestrictedName) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(610400)
	f := ctx.flags
	if f.HasFlags(FmtAnonymize) {
		__antithesis_instrumentation__.Notify(610401)
		ctx.WriteByte('_')
	} else {
		__antithesis_instrumentation__.Notify(610402)
		lexbase.EncodeUnrestrictedSQLIdent(&ctx.Buffer, string(*u), f.EncodeFlags())
	}
}

func (l NameList) ToStrings() []string {
	__antithesis_instrumentation__.Notify(610403)
	if l == nil {
		__antithesis_instrumentation__.Notify(610406)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(610407)
	}
	__antithesis_instrumentation__.Notify(610404)
	names := make([]string, len(l))
	for i, n := range l {
		__antithesis_instrumentation__.Notify(610408)
		names[i] = string(n)
	}
	__antithesis_instrumentation__.Notify(610405)
	return names
}

func (l NameList) ToSQLUsernames() ([]security.SQLUsername, error) {
	__antithesis_instrumentation__.Notify(610409)
	targetRoles := make([]security.SQLUsername, len(l))
	for i, role := range l {
		__antithesis_instrumentation__.Notify(610411)
		user, err := security.MakeSQLUsernameFromUserInput(string(role), security.UsernameValidation)
		if err != nil {
			__antithesis_instrumentation__.Notify(610413)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(610414)
		}
		__antithesis_instrumentation__.Notify(610412)
		targetRoles[i] = user
	}
	__antithesis_instrumentation__.Notify(610410)
	return targetRoles, nil
}

type NameList []Name

func (l *NameList) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(610415)
	for i := range *l {
		__antithesis_instrumentation__.Notify(610416)
		if i > 0 {
			__antithesis_instrumentation__.Notify(610418)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(610419)
		}
		__antithesis_instrumentation__.Notify(610417)
		ctx.FormatNode(&(*l)[i])
	}
}

type ArraySubscript struct {
	Begin Expr
	End   Expr
	Slice bool
}

func (a *ArraySubscript) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(610420)
	ctx.WriteByte('[')
	if a.Begin != nil {
		__antithesis_instrumentation__.Notify(610423)
		ctx.FormatNode(a.Begin)
	} else {
		__antithesis_instrumentation__.Notify(610424)
	}
	__antithesis_instrumentation__.Notify(610421)
	if a.Slice {
		__antithesis_instrumentation__.Notify(610425)
		ctx.WriteByte(':')
		if a.End != nil {
			__antithesis_instrumentation__.Notify(610426)
			ctx.FormatNode(a.End)
		} else {
			__antithesis_instrumentation__.Notify(610427)
		}
	} else {
		__antithesis_instrumentation__.Notify(610428)
	}
	__antithesis_instrumentation__.Notify(610422)
	ctx.WriteByte(']')
}

type UnresolvedName struct {
	NumParts int

	Star bool

	Parts NameParts
}

type NameParts = [4]string

func (u *UnresolvedName) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(610429)
	stopAt := 1
	if u.Star {
		__antithesis_instrumentation__.Notify(610432)
		stopAt = 2
	} else {
		__antithesis_instrumentation__.Notify(610433)
	}
	__antithesis_instrumentation__.Notify(610430)
	for i := u.NumParts; i >= stopAt; i-- {
		__antithesis_instrumentation__.Notify(610434)

		if i == u.NumParts {
			__antithesis_instrumentation__.Notify(610436)
			ctx.FormatNode((*Name)(&u.Parts[i-1]))
		} else {
			__antithesis_instrumentation__.Notify(610437)
			ctx.FormatNode((*UnrestrictedName)(&u.Parts[i-1]))
		}
		__antithesis_instrumentation__.Notify(610435)
		if i > 1 {
			__antithesis_instrumentation__.Notify(610438)
			ctx.WriteByte('.')
		} else {
			__antithesis_instrumentation__.Notify(610439)
		}
	}
	__antithesis_instrumentation__.Notify(610431)
	if u.Star {
		__antithesis_instrumentation__.Notify(610440)
		ctx.WriteByte('*')
	} else {
		__antithesis_instrumentation__.Notify(610441)
	}
}
func (u *UnresolvedName) String() string {
	__antithesis_instrumentation__.Notify(610442)
	return AsString(u)
}

func NewUnresolvedName(args ...string) *UnresolvedName {
	__antithesis_instrumentation__.Notify(610443)
	n := MakeUnresolvedName(args...)
	return &n
}

func MakeUnresolvedName(args ...string) UnresolvedName {
	__antithesis_instrumentation__.Notify(610444)
	n := UnresolvedName{NumParts: len(args)}
	for i := 0; i < len(args); i++ {
		__antithesis_instrumentation__.Notify(610446)
		n.Parts[i] = args[len(args)-1-i]
	}
	__antithesis_instrumentation__.Notify(610445)
	return n
}

func (u *UnresolvedName) ToUnresolvedObjectName(idx AnnotationIdx) (*UnresolvedObjectName, error) {
	__antithesis_instrumentation__.Notify(610447)
	if u.NumParts == 4 {
		__antithesis_instrumentation__.Notify(610449)
		return nil, pgerror.Newf(pgcode.Syntax, "improper qualified name (too many dotted names): %s", u)
	} else {
		__antithesis_instrumentation__.Notify(610450)
	}
	__antithesis_instrumentation__.Notify(610448)
	return NewUnresolvedObjectName(
		u.NumParts,
		[3]string{u.Parts[0], u.Parts[1], u.Parts[2]},
		idx,
	)
}
