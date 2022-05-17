package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessioninit"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type CreateRoleNode struct {
	ifNotExists bool
	isRole      bool
	roleOptions roleoption.List
	roleName    security.SQLUsername
}

func (p *planner) CreateRole(ctx context.Context, n *tree.CreateRole) (planNode, error) {
	__antithesis_instrumentation__.Notify(463259)
	return p.CreateRoleNode(ctx, n.Name, n.IfNotExists, n.IsRole,
		"CREATE ROLE", n.KVOptions)
}

func (p *planner) CreateRoleNode(
	ctx context.Context,
	roleSpec tree.RoleSpec,
	ifNotExists bool,
	isRole bool,
	opName string,
	kvOptions tree.KVOptions,
) (*CreateRoleNode, error) {
	__antithesis_instrumentation__.Notify(463260)
	if err := p.CheckRoleOption(ctx, roleoption.CREATEROLE); err != nil {
		__antithesis_instrumentation__.Notify(463270)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463271)
	}
	__antithesis_instrumentation__.Notify(463261)

	if roleSpec.RoleSpecType != tree.RoleName {
		__antithesis_instrumentation__.Notify(463272)
		return nil, pgerror.Newf(pgcode.ReservedName, "%s cannot be used as a role name here", roleSpec.RoleSpecType)
	} else {
		__antithesis_instrumentation__.Notify(463273)
	}
	__antithesis_instrumentation__.Notify(463262)

	asStringOrNull := func(e tree.Expr, op string) (func() (bool, string, error), error) {
		__antithesis_instrumentation__.Notify(463274)
		return p.TypeAsStringOrNull(ctx, e, op)
	}
	__antithesis_instrumentation__.Notify(463263)
	roleOptions, err := kvOptions.ToRoleOptions(asStringOrNull, opName)
	if err != nil {
		__antithesis_instrumentation__.Notify(463275)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463276)
	}
	__antithesis_instrumentation__.Notify(463264)

	if err := roleOptions.CheckRoleOptionConflicts(); err != nil {
		__antithesis_instrumentation__.Notify(463277)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463278)
	}
	__antithesis_instrumentation__.Notify(463265)

	if isRole && func() bool {
		__antithesis_instrumentation__.Notify(463279)
		return !roleOptions.Contains(roleoption.LOGIN) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(463280)
		return !roleOptions.Contains(roleoption.NOLOGIN) == true
	}() == true {
		__antithesis_instrumentation__.Notify(463281)
		roleOptions = append(roleOptions,
			roleoption.RoleOption{Option: roleoption.NOLOGIN, HasValue: false})
	} else {
		__antithesis_instrumentation__.Notify(463282)
	}
	__antithesis_instrumentation__.Notify(463266)

	if err := p.checkPasswordOptionConstraints(ctx, roleOptions, true); err != nil {
		__antithesis_instrumentation__.Notify(463283)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463284)
	}
	__antithesis_instrumentation__.Notify(463267)

	roleName, err := roleSpec.ToSQLUsername(p.SessionData(), security.UsernameCreation)
	if err != nil {
		__antithesis_instrumentation__.Notify(463285)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463286)
	}
	__antithesis_instrumentation__.Notify(463268)

	if roleName.IsReserved() {
		__antithesis_instrumentation__.Notify(463287)
		return nil, pgerror.Newf(
			pgcode.ReservedName,
			"role name %q is reserved",
			roleName.Normalized(),
		)
	} else {
		__antithesis_instrumentation__.Notify(463288)
	}
	__antithesis_instrumentation__.Notify(463269)

	return &CreateRoleNode{
		roleName:    roleName,
		ifNotExists: ifNotExists,
		isRole:      isRole,
		roleOptions: roleOptions,
	}, nil
}

func (n *CreateRoleNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(463289)
	var opName string
	if n.isRole {
		__antithesis_instrumentation__.Notify(463298)
		sqltelemetry.IncIAMCreateCounter(sqltelemetry.Role)
		opName = "create-role"
	} else {
		__antithesis_instrumentation__.Notify(463299)
		sqltelemetry.IncIAMCreateCounter(sqltelemetry.User)
		opName = "create-user"
	}
	__antithesis_instrumentation__.Notify(463290)

	_, hashedPassword, err := retrievePasswordFromRoleOptions(params, n.roleOptions)
	if err != nil {
		__antithesis_instrumentation__.Notify(463300)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463301)
	}
	__antithesis_instrumentation__.Notify(463291)

	row, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.QueryRowEx(
		params.ctx,
		opName,
		params.p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf(`select "isRole" from %s where username = $1`, sessioninit.UsersTableName),
		n.roleName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(463302)
		return errors.Wrapf(err, "error looking up user")
	} else {
		__antithesis_instrumentation__.Notify(463303)
	}
	__antithesis_instrumentation__.Notify(463292)
	if row != nil {
		__antithesis_instrumentation__.Notify(463304)
		if n.ifNotExists {
			__antithesis_instrumentation__.Notify(463306)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(463307)
		}
		__antithesis_instrumentation__.Notify(463305)
		return pgerror.Newf(pgcode.DuplicateObject,
			"a role/user named %s already exists", n.roleName.Normalized())
	} else {
		__antithesis_instrumentation__.Notify(463308)
	}
	__antithesis_instrumentation__.Notify(463293)

	rowsAffected, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.Exec(
		params.ctx,
		opName,
		params.p.txn,
		fmt.Sprintf("insert into %s values ($1, $2, $3)", sessioninit.UsersTableName),
		n.roleName,
		hashedPassword,
		n.isRole,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(463309)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463310)
		if rowsAffected != 1 {
			__antithesis_instrumentation__.Notify(463311)
			return errors.AssertionFailedf("%d rows affected by user creation; expected exactly one row affected",
				rowsAffected,
			)
		} else {
			__antithesis_instrumentation__.Notify(463312)
		}
	}
	__antithesis_instrumentation__.Notify(463294)

	stmts, err := n.roleOptions.GetSQLStmts(sqltelemetry.CreateRole)
	if err != nil {
		__antithesis_instrumentation__.Notify(463313)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463314)
	}
	__antithesis_instrumentation__.Notify(463295)

	for stmt, value := range stmts {
		__antithesis_instrumentation__.Notify(463315)
		qargs := []interface{}{n.roleName}

		if value != nil {
			__antithesis_instrumentation__.Notify(463317)
			isNull, val, err := value()
			if err != nil {
				__antithesis_instrumentation__.Notify(463319)
				return err
			} else {
				__antithesis_instrumentation__.Notify(463320)
			}
			__antithesis_instrumentation__.Notify(463318)
			if isNull {
				__antithesis_instrumentation__.Notify(463321)

				qargs = append(qargs, nil)
			} else {
				__antithesis_instrumentation__.Notify(463322)
				qargs = append(qargs, val)
			}
		} else {
			__antithesis_instrumentation__.Notify(463323)
		}
		__antithesis_instrumentation__.Notify(463316)

		_, err = params.extendedEvalCtx.ExecCfg.InternalExecutor.ExecEx(
			params.ctx,
			opName,
			params.p.txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			stmt,
			qargs...,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(463324)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463325)
		}
	}
	__antithesis_instrumentation__.Notify(463296)

	if sessioninit.CacheEnabled.Get(&params.p.ExecCfg().Settings.SV) {
		__antithesis_instrumentation__.Notify(463326)

		if err := params.p.bumpUsersTableVersion(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(463328)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463329)
		}
		__antithesis_instrumentation__.Notify(463327)
		if err := params.p.bumpRoleOptionsTableVersion(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(463330)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463331)
		}
	} else {
		__antithesis_instrumentation__.Notify(463332)
	}
	__antithesis_instrumentation__.Notify(463297)

	return params.p.logEvent(params.ctx,
		0,
		&eventpb.CreateRole{RoleName: n.roleName.Normalized()})
}

func (*CreateRoleNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(463333)
	return false, nil
}

func (*CreateRoleNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(463334)
	return tree.Datums{}
}

func (*CreateRoleNode) Close(context.Context) { __antithesis_instrumentation__.Notify(463335) }

func retrievePasswordFromRoleOptions(
	params runParams, roleOptions roleoption.List,
) (hasPasswordOpt bool, hashedPassword []byte, err error) {
	__antithesis_instrumentation__.Notify(463336)
	if !roleOptions.Contains(roleoption.PASSWORD) {
		__antithesis_instrumentation__.Notify(463341)
		return false, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(463342)
	}
	__antithesis_instrumentation__.Notify(463337)
	isNull, password, err := roleOptions.GetPassword()
	if err != nil {
		__antithesis_instrumentation__.Notify(463343)
		return true, nil, err
	} else {
		__antithesis_instrumentation__.Notify(463344)
	}
	__antithesis_instrumentation__.Notify(463338)
	if !isNull && func() bool {
		__antithesis_instrumentation__.Notify(463345)
		return params.extendedEvalCtx.ExecCfg.RPCContext.Config.Insecure == true
	}() == true {
		__antithesis_instrumentation__.Notify(463346)

		return true, nil, pgerror.New(pgcode.InvalidPassword,
			"setting or updating a password is not supported in insecure mode")
	} else {
		__antithesis_instrumentation__.Notify(463347)
	}
	__antithesis_instrumentation__.Notify(463339)

	if !isNull {
		__antithesis_instrumentation__.Notify(463348)
		if hashedPassword, err = params.p.checkPasswordAndGetHash(params.ctx, password); err != nil {
			__antithesis_instrumentation__.Notify(463349)
			return true, nil, err
		} else {
			__antithesis_instrumentation__.Notify(463350)
		}
	} else {
		__antithesis_instrumentation__.Notify(463351)
	}
	__antithesis_instrumentation__.Notify(463340)

	return true, hashedPassword, nil
}

func (p *planner) checkPasswordAndGetHash(
	ctx context.Context, password string,
) (hashedPassword []byte, err error) {
	__antithesis_instrumentation__.Notify(463352)
	if password == "" {
		__antithesis_instrumentation__.Notify(463357)
		return hashedPassword, security.ErrEmptyPassword
	} else {
		__antithesis_instrumentation__.Notify(463358)
	}
	__antithesis_instrumentation__.Notify(463353)

	st := p.ExecCfg().Settings
	if security.AutoDetectPasswordHashes.Get(&st.SV) {
		__antithesis_instrumentation__.Notify(463359)
		var isPreHashed, schemeSupported bool
		var schemeName string
		var issueNum int
		isPreHashed, schemeSupported, issueNum, schemeName, hashedPassword, err = security.CheckPasswordHashValidity(ctx, []byte(password))
		if err != nil {
			__antithesis_instrumentation__.Notify(463361)
			return hashedPassword, pgerror.WithCandidateCode(err, pgcode.Syntax)
		} else {
			__antithesis_instrumentation__.Notify(463362)
		}
		__antithesis_instrumentation__.Notify(463360)
		if isPreHashed {
			__antithesis_instrumentation__.Notify(463363)
			if !schemeSupported {
				__antithesis_instrumentation__.Notify(463365)
				return hashedPassword, unimplemented.NewWithIssueDetailf(issueNum, schemeName, "the password hash scheme %q is not supported", schemeName)
			} else {
				__antithesis_instrumentation__.Notify(463366)
			}
			__antithesis_instrumentation__.Notify(463364)
			return hashedPassword, nil
		} else {
			__antithesis_instrumentation__.Notify(463367)
		}
	} else {
		__antithesis_instrumentation__.Notify(463368)
	}
	__antithesis_instrumentation__.Notify(463354)

	if minLength := security.MinPasswordLength.Get(&st.SV); minLength >= 1 && func() bool {
		__antithesis_instrumentation__.Notify(463369)
		return int64(len(password)) < minLength == true
	}() == true {
		__antithesis_instrumentation__.Notify(463370)
		return nil, errors.WithHintf(security.ErrPasswordTooShort,
			"Passwords must be %d characters or longer.", minLength)
	} else {
		__antithesis_instrumentation__.Notify(463371)
	}
	__antithesis_instrumentation__.Notify(463355)

	hashedPassword, err = security.HashPassword(ctx, &st.SV, password)
	if err != nil {
		__antithesis_instrumentation__.Notify(463372)
		return hashedPassword, err
	} else {
		__antithesis_instrumentation__.Notify(463373)
	}
	__antithesis_instrumentation__.Notify(463356)

	return hashedPassword, nil
}
