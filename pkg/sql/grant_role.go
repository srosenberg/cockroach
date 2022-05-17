package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

type GrantRoleNode struct {
	roles       []security.SQLUsername
	members     []security.SQLUsername
	adminOption bool

	run grantRoleRun
}

type grantRoleRun struct {
	rowsAffected int
}

func (p *planner) GrantRole(ctx context.Context, n *tree.GrantRole) (planNode, error) {
	__antithesis_instrumentation__.Notify(492906)
	return p.GrantRoleNode(ctx, n)
}

func (p *planner) GrantRoleNode(ctx context.Context, n *tree.GrantRole) (*GrantRoleNode, error) {
	__antithesis_instrumentation__.Notify(492907)
	sqltelemetry.IncIAMGrantCounter(n.AdminOption)

	ctx, span := tracing.ChildSpan(ctx, n.StatementTag())
	defer span.Finish()

	hasAdminRole, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(492918)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(492919)
	}
	__antithesis_instrumentation__.Notify(492908)

	allRoles, err := p.MemberOfWithAdminOption(ctx, p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(492920)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(492921)
	}
	__antithesis_instrumentation__.Notify(492909)

	inputRoles, err := n.Roles.ToSQLUsernames()
	if err != nil {
		__antithesis_instrumentation__.Notify(492922)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(492923)
	}
	__antithesis_instrumentation__.Notify(492910)
	inputMembers, err := n.Members.ToSQLUsernames(p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(492924)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(492925)
	}
	__antithesis_instrumentation__.Notify(492911)

	for _, r := range inputRoles {
		__antithesis_instrumentation__.Notify(492926)

		if hasAdminRole && func() bool {
			__antithesis_instrumentation__.Notify(492928)
			return !r.IsAdminRole() == true
		}() == true {
			__antithesis_instrumentation__.Notify(492929)
			continue
		} else {
			__antithesis_instrumentation__.Notify(492930)
		}
		__antithesis_instrumentation__.Notify(492927)
		if isAdmin, ok := allRoles[r]; !ok || func() bool {
			__antithesis_instrumentation__.Notify(492931)
			return !isAdmin == true
		}() == true {
			__antithesis_instrumentation__.Notify(492932)
			if r.IsAdminRole() {
				__antithesis_instrumentation__.Notify(492934)
				return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
					"%s is not a role admin for role %s", p.User(), r)
			} else {
				__antithesis_instrumentation__.Notify(492935)
			}
			__antithesis_instrumentation__.Notify(492933)
			return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
				"%s is not a superuser or role admin for role %s", p.User(), r)
		} else {
			__antithesis_instrumentation__.Notify(492936)
		}
	}
	__antithesis_instrumentation__.Notify(492912)

	roles, err := p.GetAllRoles(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(492937)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(492938)
	}
	__antithesis_instrumentation__.Notify(492913)

	for _, r := range inputRoles {
		__antithesis_instrumentation__.Notify(492939)
		if _, ok := roles[r]; !ok {
			__antithesis_instrumentation__.Notify(492940)
			maybeOption := strings.ToUpper(r.Normalized())
			for name := range roleoption.ByName {
				__antithesis_instrumentation__.Notify(492942)
				if maybeOption == name {
					__antithesis_instrumentation__.Notify(492943)
					return nil, errors.WithHintf(
						pgerror.Newf(pgcode.UndefinedObject,
							"role/user %s does not exist", r),
						"%s is a role option, try using ALTER ROLE to change a role's options.", maybeOption)
				} else {
					__antithesis_instrumentation__.Notify(492944)
				}
			}
			__antithesis_instrumentation__.Notify(492941)
			return nil, pgerror.Newf(pgcode.UndefinedObject, "role/user %s does not exist", r)
		} else {
			__antithesis_instrumentation__.Notify(492945)
		}
	}
	__antithesis_instrumentation__.Notify(492914)

	for _, m := range inputMembers {
		__antithesis_instrumentation__.Notify(492946)
		if _, ok := roles[m]; !ok {
			__antithesis_instrumentation__.Notify(492947)
			return nil, pgerror.Newf(pgcode.UndefinedObject, "role/user %s does not exist", m)
		} else {
			__antithesis_instrumentation__.Notify(492948)
		}
	}
	__antithesis_instrumentation__.Notify(492915)

	allRoleMemberships := make(map[security.SQLUsername]map[security.SQLUsername]bool)
	for _, r := range inputRoles {
		__antithesis_instrumentation__.Notify(492949)
		allRoles, err := p.MemberOfWithAdminOption(ctx, r)
		if err != nil {
			__antithesis_instrumentation__.Notify(492951)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(492952)
		}
		__antithesis_instrumentation__.Notify(492950)
		allRoleMemberships[r] = allRoles
	}
	__antithesis_instrumentation__.Notify(492916)

	for _, r := range inputRoles {
		__antithesis_instrumentation__.Notify(492953)
		for _, m := range inputMembers {
			__antithesis_instrumentation__.Notify(492954)
			if r == m {
				__antithesis_instrumentation__.Notify(492958)

				return nil, pgerror.Newf(pgcode.InvalidGrantOperation, "%s cannot be a member of itself", m)
			} else {
				__antithesis_instrumentation__.Notify(492959)
			}
			__antithesis_instrumentation__.Notify(492955)

			if memberOf, ok := allRoleMemberships[r]; ok {
				__antithesis_instrumentation__.Notify(492960)
				if _, ok = memberOf[m]; ok {
					__antithesis_instrumentation__.Notify(492961)
					return nil, pgerror.Newf(pgcode.InvalidGrantOperation,
						"making %s a member of %s would create a cycle", m, r)
				} else {
					__antithesis_instrumentation__.Notify(492962)
				}
			} else {
				__antithesis_instrumentation__.Notify(492963)
			}
			__antithesis_instrumentation__.Notify(492956)

			if _, ok := allRoleMemberships[m]; !ok {
				__antithesis_instrumentation__.Notify(492964)
				allRoleMemberships[m] = make(map[security.SQLUsername]bool)
			} else {
				__antithesis_instrumentation__.Notify(492965)
			}
			__antithesis_instrumentation__.Notify(492957)
			allRoleMemberships[m][r] = false
		}
	}
	__antithesis_instrumentation__.Notify(492917)

	return &GrantRoleNode{
		roles:       inputRoles,
		members:     inputMembers,
		adminOption: n.AdminOption,
	}, nil
}

func (n *GrantRoleNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(492966)
	opName := "grant-role"

	memberStmt := `INSERT INTO system.role_members ("role", "member", "isAdmin") VALUES ($1, $2, $3) ON CONFLICT ("role", "member")`
	if n.adminOption {
		__antithesis_instrumentation__.Notify(492971)

		memberStmt += ` DO UPDATE SET "isAdmin" = true`
	} else {
		__antithesis_instrumentation__.Notify(492972)

		memberStmt += ` DO NOTHING`
	}
	__antithesis_instrumentation__.Notify(492967)

	var rowsAffected int
	for _, r := range n.roles {
		__antithesis_instrumentation__.Notify(492973)
		for _, m := range n.members {
			__antithesis_instrumentation__.Notify(492974)
			affected, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.ExecEx(
				params.ctx,
				opName,
				params.p.txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				memberStmt,
				r.Normalized(), m.Normalized(), n.adminOption,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(492976)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492977)
			}
			__antithesis_instrumentation__.Notify(492975)

			rowsAffected += affected
		}
	}
	__antithesis_instrumentation__.Notify(492968)

	if rowsAffected > 0 {
		__antithesis_instrumentation__.Notify(492978)
		if err := params.p.BumpRoleMembershipTableVersion(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(492979)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492980)
		}
	} else {
		__antithesis_instrumentation__.Notify(492981)
	}
	__antithesis_instrumentation__.Notify(492969)

	n.run.rowsAffected += rowsAffected

	sqlUsernameToStrings := func(sqlUsernames []security.SQLUsername) []string {
		__antithesis_instrumentation__.Notify(492982)
		strings := make([]string, len(sqlUsernames))
		for i, sqlUsername := range sqlUsernames {
			__antithesis_instrumentation__.Notify(492984)
			strings[i] = sqlUsername.Normalized()
		}
		__antithesis_instrumentation__.Notify(492983)
		return strings
	}
	__antithesis_instrumentation__.Notify(492970)

	return params.p.logEvent(params.ctx,
		0,
		&eventpb.GrantRole{GranteeRoles: sqlUsernameToStrings(n.roles), Members: sqlUsernameToStrings(n.members)})
}

func (*GrantRoleNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(492985)
	return false, nil
}

func (*GrantRoleNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(492986)
	return tree.Datums{}
}

func (*GrantRoleNode) Close(context.Context) { __antithesis_instrumentation__.Notify(492987) }
