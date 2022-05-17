package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

type RevokeRoleNode struct {
	roles       []security.SQLUsername
	members     []security.SQLUsername
	adminOption bool

	run revokeRoleRun
}

type revokeRoleRun struct {
	rowsAffected int
}

func (p *planner) RevokeRole(ctx context.Context, n *tree.RevokeRole) (planNode, error) {
	__antithesis_instrumentation__.Notify(567193)
	return p.RevokeRoleNode(ctx, n)
}

func (p *planner) RevokeRoleNode(ctx context.Context, n *tree.RevokeRole) (*RevokeRoleNode, error) {
	__antithesis_instrumentation__.Notify(567194)
	sqltelemetry.IncIAMRevokeCounter(n.AdminOption)

	ctx, span := tracing.ChildSpan(ctx, n.StatementTag())
	defer span.Finish()

	hasAdminRole, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(567203)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(567204)
	}
	__antithesis_instrumentation__.Notify(567195)

	allRoles, err := p.MemberOfWithAdminOption(ctx, p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(567205)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(567206)
	}
	__antithesis_instrumentation__.Notify(567196)

	inputRoles, err := n.Roles.ToSQLUsernames()
	if err != nil {
		__antithesis_instrumentation__.Notify(567207)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(567208)
	}
	__antithesis_instrumentation__.Notify(567197)
	inputMembers, err := n.Members.ToSQLUsernames(p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(567209)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(567210)
	}
	__antithesis_instrumentation__.Notify(567198)

	for _, r := range inputRoles {
		__antithesis_instrumentation__.Notify(567211)

		if hasAdminRole && func() bool {
			__antithesis_instrumentation__.Notify(567213)
			return !r.IsAdminRole() == true
		}() == true {
			__antithesis_instrumentation__.Notify(567214)
			continue
		} else {
			__antithesis_instrumentation__.Notify(567215)
		}
		__antithesis_instrumentation__.Notify(567212)
		if isAdmin, ok := allRoles[r]; !ok || func() bool {
			__antithesis_instrumentation__.Notify(567216)
			return !isAdmin == true
		}() == true {
			__antithesis_instrumentation__.Notify(567217)
			if r.IsAdminRole() {
				__antithesis_instrumentation__.Notify(567219)
				return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
					"%s is not a role admin for role %s", p.User(), r)
			} else {
				__antithesis_instrumentation__.Notify(567220)
			}
			__antithesis_instrumentation__.Notify(567218)
			return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
				"%s is not a superuser or role admin for role %s", p.User(), r)
		} else {
			__antithesis_instrumentation__.Notify(567221)
		}
	}
	__antithesis_instrumentation__.Notify(567199)

	roles, err := p.GetAllRoles(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(567222)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(567223)
	}
	__antithesis_instrumentation__.Notify(567200)

	for _, r := range inputRoles {
		__antithesis_instrumentation__.Notify(567224)
		if _, ok := roles[r]; !ok {
			__antithesis_instrumentation__.Notify(567225)
			return nil, pgerror.Newf(pgcode.UndefinedObject, "role/user %s does not exist", r)
		} else {
			__antithesis_instrumentation__.Notify(567226)
		}
	}
	__antithesis_instrumentation__.Notify(567201)

	for _, m := range inputMembers {
		__antithesis_instrumentation__.Notify(567227)
		if _, ok := roles[m]; !ok {
			__antithesis_instrumentation__.Notify(567228)
			return nil, pgerror.Newf(pgcode.UndefinedObject, "role/user %s does not exist", m)
		} else {
			__antithesis_instrumentation__.Notify(567229)
		}
	}
	__antithesis_instrumentation__.Notify(567202)

	return &RevokeRoleNode{
		roles:       inputRoles,
		members:     inputMembers,
		adminOption: n.AdminOption,
	}, nil
}

func (n *RevokeRoleNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(567230)
	opName := "revoke-role"

	var memberStmt string
	if n.adminOption {
		__antithesis_instrumentation__.Notify(567234)

		memberStmt = `UPDATE system.role_members SET "isAdmin" = false WHERE "role" = $1 AND "member" = $2`
	} else {
		__antithesis_instrumentation__.Notify(567235)

		memberStmt = `DELETE FROM system.role_members WHERE "role" = $1 AND "member" = $2`
	}
	__antithesis_instrumentation__.Notify(567231)

	var rowsAffected int
	for _, r := range n.roles {
		__antithesis_instrumentation__.Notify(567236)
		for _, m := range n.members {
			__antithesis_instrumentation__.Notify(567237)
			if r.IsAdminRole() && func() bool {
				__antithesis_instrumentation__.Notify(567240)
				return m.IsRootUser() == true
			}() == true {
				__antithesis_instrumentation__.Notify(567241)

				return pgerror.Newf(pgcode.ObjectInUse,
					"role/user %s cannot be removed from role %s or lose the ADMIN OPTION",
					security.RootUser, security.AdminRole)
			} else {
				__antithesis_instrumentation__.Notify(567242)
			}
			__antithesis_instrumentation__.Notify(567238)
			affected, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.ExecEx(
				params.ctx,
				opName,
				params.p.txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				memberStmt,
				r.Normalized(), m.Normalized(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(567243)
				return err
			} else {
				__antithesis_instrumentation__.Notify(567244)
			}
			__antithesis_instrumentation__.Notify(567239)

			rowsAffected += affected
		}
	}
	__antithesis_instrumentation__.Notify(567232)

	if rowsAffected > 0 {
		__antithesis_instrumentation__.Notify(567245)
		if err := params.p.BumpRoleMembershipTableVersion(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(567246)
			return err
		} else {
			__antithesis_instrumentation__.Notify(567247)
		}
	} else {
		__antithesis_instrumentation__.Notify(567248)
	}
	__antithesis_instrumentation__.Notify(567233)

	n.run.rowsAffected += rowsAffected

	return nil
}

func (*RevokeRoleNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(567249)
	return false, nil
}

func (*RevokeRoleNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(567250)
	return tree.Datums{}
}

func (*RevokeRoleNode) Close(context.Context) { __antithesis_instrumentation__.Notify(567251) }
