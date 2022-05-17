package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/errors"
)

type RoleSpecType int

const (
	RoleName RoleSpecType = iota

	CurrentUser

	SessionUser
)

func (r RoleSpecType) String() string {
	__antithesis_instrumentation__.Notify(612947)
	switch r {
	case RoleName:
		__antithesis_instrumentation__.Notify(612948)
		return "ROLE_NAME"
	case CurrentUser:
		__antithesis_instrumentation__.Notify(612949)
		return "CURRENT_USER"
	case SessionUser:
		__antithesis_instrumentation__.Notify(612950)
		return "SESSION_USER"
	default:
		__antithesis_instrumentation__.Notify(612951)
		panic(fmt.Sprintf("unknown role spec type: %d", r))
	}
}

type RoleSpecList []RoleSpec

type RoleSpec struct {
	RoleSpecType RoleSpecType
	Name         string
}

func MakeRoleSpecWithRoleName(name string) RoleSpec {
	__antithesis_instrumentation__.Notify(612952)
	return RoleSpec{RoleSpecType: RoleName, Name: name}
}

func (r RoleSpec) ToSQLUsername(
	sessionData *sessiondata.SessionData, purpose security.UsernamePurpose,
) (security.SQLUsername, error) {
	__antithesis_instrumentation__.Notify(612953)
	if r.RoleSpecType == CurrentUser {
		__antithesis_instrumentation__.Notify(612956)
		return sessionData.User(), nil
	} else {
		__antithesis_instrumentation__.Notify(612957)
		if r.RoleSpecType == SessionUser {
			__antithesis_instrumentation__.Notify(612958)
			return sessionData.SessionUser(), nil
		} else {
			__antithesis_instrumentation__.Notify(612959)
		}
	}
	__antithesis_instrumentation__.Notify(612954)
	username, err := security.MakeSQLUsernameFromUserInput(r.Name, purpose)
	if err != nil {
		__antithesis_instrumentation__.Notify(612960)
		if errors.Is(err, security.ErrUsernameTooLong) {
			__antithesis_instrumentation__.Notify(612962)
			err = pgerror.WithCandidateCode(err, pgcode.NameTooLong)
		} else {
			__antithesis_instrumentation__.Notify(612963)
			if errors.IsAny(err, security.ErrUsernameInvalid, security.ErrUsernameEmpty) {
				__antithesis_instrumentation__.Notify(612964)
				err = pgerror.WithCandidateCode(err, pgcode.InvalidName)
			} else {
				__antithesis_instrumentation__.Notify(612965)
			}
		}
		__antithesis_instrumentation__.Notify(612961)
		return username, errors.Wrapf(err, "%q", username)
	} else {
		__antithesis_instrumentation__.Notify(612966)
	}
	__antithesis_instrumentation__.Notify(612955)
	return username, nil
}

func (l RoleSpecList) ToSQLUsernames(
	sessionData *sessiondata.SessionData, purpose security.UsernamePurpose,
) ([]security.SQLUsername, error) {
	__antithesis_instrumentation__.Notify(612967)
	targetRoles := make([]security.SQLUsername, len(l))
	for i, role := range l {
		__antithesis_instrumentation__.Notify(612969)
		user, err := role.ToSQLUsername(sessionData, purpose)
		if err != nil {
			__antithesis_instrumentation__.Notify(612971)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(612972)
		}
		__antithesis_instrumentation__.Notify(612970)
		targetRoles[i] = user
	}
	__antithesis_instrumentation__.Notify(612968)
	return targetRoles, nil
}

func (r RoleSpec) Undefined() bool {
	__antithesis_instrumentation__.Notify(612973)
	return r.RoleSpecType == RoleName && func() bool {
		__antithesis_instrumentation__.Notify(612974)
		return len(r.Name) == 0 == true
	}() == true
}

func (r *RoleSpec) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612975)
	f := ctx.flags
	if f.HasFlags(FmtAnonymize) && func() bool {
		__antithesis_instrumentation__.Notify(612976)
		return !isArityIndicatorString(r.Name) == true
	}() == true {
		__antithesis_instrumentation__.Notify(612977)
		ctx.WriteByte('_')
	} else {
		__antithesis_instrumentation__.Notify(612978)
		switch r.RoleSpecType {
		case RoleName:
			__antithesis_instrumentation__.Notify(612979)
			lexbase.EncodeRestrictedSQLIdent(&ctx.Buffer, r.Name, f.EncodeFlags())
			return
		case CurrentUser, SessionUser:
			__antithesis_instrumentation__.Notify(612980)
			ctx.WriteString(r.RoleSpecType.String())
		default:
			__antithesis_instrumentation__.Notify(612981)
		}
	}
}

func (l *RoleSpecList) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612982)
	for i := range *l {
		__antithesis_instrumentation__.Notify(612983)
		if i > 0 {
			__antithesis_instrumentation__.Notify(612985)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(612986)
		}
		__antithesis_instrumentation__.Notify(612984)
		ctx.FormatNode(&(*l)[i])
	}
}
