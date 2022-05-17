package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/paramparse"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessioninit"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type alterRoleNode struct {
	roleName    security.SQLUsername
	ifExists    bool
	isRole      bool
	roleOptions roleoption.List
}

type alterRoleSetNode struct {
	roleName security.SQLUsername
	ifExists bool
	isRole   bool
	allRoles bool

	dbDescID    descpb.ID
	setVarKind  setVarBehavior
	varName     string
	sVar        sessionVar
	typedValues []tree.TypedExpr
}

type setVarBehavior int

const (
	setSingleVar   setVarBehavior = 0
	resetSingleVar setVarBehavior = 1
	resetAllVars   setVarBehavior = 2
	unknown        setVarBehavior = 3
)

func (p *planner) AlterRole(ctx context.Context, n *tree.AlterRole) (planNode, error) {
	__antithesis_instrumentation__.Notify(243515)
	return p.AlterRoleNode(ctx, n.Name, n.IfExists, n.IsRole, "ALTER ROLE", n.KVOptions)
}

func (p *planner) AlterRoleNode(
	ctx context.Context,
	roleSpec tree.RoleSpec,
	ifExists bool,
	isRole bool,
	opName string,
	kvOptions tree.KVOptions,
) (*alterRoleNode, error) {
	__antithesis_instrumentation__.Notify(243516)

	if err := p.CheckRoleOption(ctx, roleoption.CREATEROLE); err != nil {
		__antithesis_instrumentation__.Notify(243523)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243524)
	}
	__antithesis_instrumentation__.Notify(243517)

	asStringOrNull := func(e tree.Expr, op string) (func() (bool, string, error), error) {
		__antithesis_instrumentation__.Notify(243525)
		return p.TypeAsStringOrNull(ctx, e, op)
	}
	__antithesis_instrumentation__.Notify(243518)
	roleOptions, err := kvOptions.ToRoleOptions(asStringOrNull, opName)
	if err != nil {
		__antithesis_instrumentation__.Notify(243526)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243527)
	}
	__antithesis_instrumentation__.Notify(243519)
	if err := roleOptions.CheckRoleOptionConflicts(); err != nil {
		__antithesis_instrumentation__.Notify(243528)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243529)
	}
	__antithesis_instrumentation__.Notify(243520)

	if err := p.checkPasswordOptionConstraints(ctx, roleOptions, false); err != nil {
		__antithesis_instrumentation__.Notify(243530)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243531)
	}
	__antithesis_instrumentation__.Notify(243521)

	roleName, err := roleSpec.ToSQLUsername(p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(243532)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243533)
	}
	__antithesis_instrumentation__.Notify(243522)

	return &alterRoleNode{
		roleName:    roleName,
		ifExists:    ifExists,
		isRole:      isRole,
		roleOptions: roleOptions,
	}, nil
}

func (p *planner) checkPasswordOptionConstraints(
	ctx context.Context, roleOptions roleoption.List, newUser bool,
) error {
	__antithesis_instrumentation__.Notify(243534)
	if roleOptions.Contains(roleoption.CREATELOGIN) || func() bool {
		__antithesis_instrumentation__.Notify(243536)
		return roleOptions.Contains(roleoption.NOCREATELOGIN) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(243537)
		return roleOptions.Contains(roleoption.PASSWORD) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(243538)
		return roleOptions.Contains(roleoption.VALIDUNTIL) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(243539)
		return roleOptions.Contains(roleoption.LOGIN) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(243540)
		return (roleOptions.Contains(roleoption.NOLOGIN) && func() bool {
			__antithesis_instrumentation__.Notify(243541)
			return !newUser == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(243542)
		return (newUser && func() bool {
			__antithesis_instrumentation__.Notify(243543)
			return !roleOptions.Contains(roleoption.NOLOGIN) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(243544)
			return !roleOptions.Contains(roleoption.LOGIN) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(243545)

		if err := p.CheckRoleOption(ctx, roleoption.CREATELOGIN); err != nil {
			__antithesis_instrumentation__.Notify(243546)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243547)
		}
	} else {
		__antithesis_instrumentation__.Notify(243548)
	}
	__antithesis_instrumentation__.Notify(243535)
	return nil
}

func (n *alterRoleNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(243549)
	var opName string
	if n.isRole {
		__antithesis_instrumentation__.Notify(243563)
		sqltelemetry.IncIAMAlterCounter(sqltelemetry.Role)
		opName = "alter-role"
	} else {
		__antithesis_instrumentation__.Notify(243564)
		sqltelemetry.IncIAMAlterCounter(sqltelemetry.User)
		opName = "alter-user"
	}
	__antithesis_instrumentation__.Notify(243550)
	if n.roleName.Undefined() {
		__antithesis_instrumentation__.Notify(243565)
		return pgerror.New(pgcode.InvalidParameterValue, "no username specified")
	} else {
		__antithesis_instrumentation__.Notify(243566)
	}
	__antithesis_instrumentation__.Notify(243551)
	if n.roleName.IsAdminRole() {
		__antithesis_instrumentation__.Notify(243567)
		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"cannot edit admin role")
	} else {
		__antithesis_instrumentation__.Notify(243568)
	}
	__antithesis_instrumentation__.Notify(243552)

	row, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.QueryRowEx(
		params.ctx,
		opName,
		params.p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf("SELECT 1 FROM %s WHERE username = $1", sessioninit.UsersTableName),
		n.roleName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(243569)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243570)
	}
	__antithesis_instrumentation__.Notify(243553)
	if row == nil {
		__antithesis_instrumentation__.Notify(243571)
		if n.ifExists {
			__antithesis_instrumentation__.Notify(243573)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(243574)
		}
		__antithesis_instrumentation__.Notify(243572)
		return pgerror.Newf(pgcode.UndefinedObject, "role/user %s does not exist", n.roleName)
	} else {
		__antithesis_instrumentation__.Notify(243575)
	}
	__antithesis_instrumentation__.Notify(243554)

	isAdmin, err := params.p.UserHasAdminRole(params.ctx, n.roleName)
	if err != nil {
		__antithesis_instrumentation__.Notify(243576)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243577)
	}
	__antithesis_instrumentation__.Notify(243555)
	if isAdmin {
		__antithesis_instrumentation__.Notify(243578)
		if err := params.p.RequireAdminRole(params.ctx, "ALTER ROLE admin"); err != nil {
			__antithesis_instrumentation__.Notify(243579)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243580)
		}
	} else {
		__antithesis_instrumentation__.Notify(243581)
	}
	__antithesis_instrumentation__.Notify(243556)

	hasPasswordOpt, hashedPassword, err := retrievePasswordFromRoleOptions(params, n.roleOptions)
	if err != nil {
		__antithesis_instrumentation__.Notify(243582)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243583)
	}
	__antithesis_instrumentation__.Notify(243557)
	if hasPasswordOpt {
		__antithesis_instrumentation__.Notify(243584)

		_, err = params.extendedEvalCtx.ExecCfg.InternalExecutor.Exec(
			params.ctx,
			opName,
			params.p.txn,
			`UPDATE system.users SET "hashedPassword" = $2 WHERE username = $1`,
			n.roleName,
			hashedPassword,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(243586)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243587)
		}
		__antithesis_instrumentation__.Notify(243585)
		if sessioninit.CacheEnabled.Get(&params.p.ExecCfg().Settings.SV) {
			__antithesis_instrumentation__.Notify(243588)

			if err := params.p.bumpUsersTableVersion(params.ctx); err != nil {
				__antithesis_instrumentation__.Notify(243589)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243590)
			}
		} else {
			__antithesis_instrumentation__.Notify(243591)
		}
	} else {
		__antithesis_instrumentation__.Notify(243592)
	}
	__antithesis_instrumentation__.Notify(243558)

	stmts, err := n.roleOptions.GetSQLStmts(sqltelemetry.AlterRole)
	if err != nil {
		__antithesis_instrumentation__.Notify(243593)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243594)
	}
	__antithesis_instrumentation__.Notify(243559)

	for stmt, value := range stmts {
		__antithesis_instrumentation__.Notify(243595)
		qargs := []interface{}{n.roleName}

		if value != nil {
			__antithesis_instrumentation__.Notify(243597)
			isNull, val, err := value()
			if err != nil {
				__antithesis_instrumentation__.Notify(243599)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243600)
			}
			__antithesis_instrumentation__.Notify(243598)
			if isNull {
				__antithesis_instrumentation__.Notify(243601)

				qargs = append(qargs, nil)
			} else {
				__antithesis_instrumentation__.Notify(243602)
				qargs = append(qargs, val)
			}
		} else {
			__antithesis_instrumentation__.Notify(243603)
		}
		__antithesis_instrumentation__.Notify(243596)

		_, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.ExecEx(
			params.ctx,
			opName,
			params.p.txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			stmt,
			qargs...,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(243604)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243605)
		}
	}
	__antithesis_instrumentation__.Notify(243560)

	optStrs := make([]string, len(n.roleOptions))
	for i := range optStrs {
		__antithesis_instrumentation__.Notify(243606)
		optStrs[i] = n.roleOptions[i].String()
	}
	__antithesis_instrumentation__.Notify(243561)

	if sessioninit.CacheEnabled.Get(&params.p.ExecCfg().Settings.SV) {
		__antithesis_instrumentation__.Notify(243607)

		if err := params.p.bumpRoleOptionsTableVersion(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(243608)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243609)
		}
	} else {
		__antithesis_instrumentation__.Notify(243610)
	}
	__antithesis_instrumentation__.Notify(243562)

	return params.p.logEvent(params.ctx,
		0,
		&eventpb.AlterRole{
			RoleName: n.roleName.Normalized(),
			Options:  optStrs,
		})
}

func (*alterRoleNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(243611)
	return false, nil
}
func (*alterRoleNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(243612)
	return tree.Datums{}
}
func (*alterRoleNode) Close(context.Context) { __antithesis_instrumentation__.Notify(243613) }

func (p *planner) AlterRoleSet(ctx context.Context, n *tree.AlterRoleSet) (planNode, error) {
	__antithesis_instrumentation__.Notify(243614)

	if n.AllRoles {
		__antithesis_instrumentation__.Notify(243619)
		if err := p.RequireAdminRole(ctx, "ALTER ROLE ALL"); err != nil {
			__antithesis_instrumentation__.Notify(243620)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(243621)
		}
	} else {
		__antithesis_instrumentation__.Notify(243622)
		if err := p.CheckRoleOption(ctx, roleoption.CREATEROLE); err != nil {
			__antithesis_instrumentation__.Notify(243623)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(243624)
		}
	}
	__antithesis_instrumentation__.Notify(243615)

	var roleName security.SQLUsername
	if !n.AllRoles {
		__antithesis_instrumentation__.Notify(243625)
		var err error
		roleName, err = n.RoleName.ToSQLUsername(p.SessionData(), security.UsernameValidation)
		if err != nil {
			__antithesis_instrumentation__.Notify(243626)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(243627)
		}
	} else {
		__antithesis_instrumentation__.Notify(243628)
	}
	__antithesis_instrumentation__.Notify(243616)

	dbDescID := descpb.ID(0)
	if n.DatabaseName != "" {
		__antithesis_instrumentation__.Notify(243629)
		dbDesc, err := p.Descriptors().GetImmutableDatabaseByName(ctx, p.txn, string(n.DatabaseName),
			tree.DatabaseLookupFlags{Required: true})
		if err != nil {
			__antithesis_instrumentation__.Notify(243631)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(243632)
		}
		__antithesis_instrumentation__.Notify(243630)
		dbDescID = dbDesc.GetID()
	} else {
		__antithesis_instrumentation__.Notify(243633)
	}
	__antithesis_instrumentation__.Notify(243617)

	setVarKind, varName, sVar, typedValues, err := p.processSetOrResetClause(ctx, n.SetOrReset)
	if err != nil {
		__antithesis_instrumentation__.Notify(243634)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243635)
	}
	__antithesis_instrumentation__.Notify(243618)

	return &alterRoleSetNode{
		roleName:    roleName,
		ifExists:    n.IfExists,
		isRole:      n.IsRole,
		allRoles:    n.AllRoles,
		dbDescID:    dbDescID,
		setVarKind:  setVarKind,
		varName:     varName,
		sVar:        sVar,
		typedValues: typedValues,
	}, nil
}

func (p *planner) processSetOrResetClause(
	ctx context.Context, setOrResetClause *tree.SetVar,
) (
	setVarKind setVarBehavior,
	varName string,
	sVar sessionVar,
	typedValues []tree.TypedExpr,
	err error,
) {
	__antithesis_instrumentation__.Notify(243636)
	if setOrResetClause.ResetAll {
		__antithesis_instrumentation__.Notify(243645)
		return resetAllVars, "", sessionVar{}, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(243646)
	}
	__antithesis_instrumentation__.Notify(243637)

	if setOrResetClause.Name == "" {
		__antithesis_instrumentation__.Notify(243647)

		return unknown, "", sessionVar{}, nil, pgerror.Newf(pgcode.Syntax, "invalid variable name: %q", setOrResetClause.Name)
	} else {
		__antithesis_instrumentation__.Notify(243648)
	}
	__antithesis_instrumentation__.Notify(243638)

	isReset := false
	if len(setOrResetClause.Values) == 1 {
		__antithesis_instrumentation__.Notify(243649)
		if _, ok := setOrResetClause.Values[0].(tree.DefaultVal); ok {
			__antithesis_instrumentation__.Notify(243650)

			isReset = true
		} else {
			__antithesis_instrumentation__.Notify(243651)
		}
	} else {
		__antithesis_instrumentation__.Notify(243652)
	}
	__antithesis_instrumentation__.Notify(243639)
	varName = strings.ToLower(setOrResetClause.Name)

	if isReset {
		__antithesis_instrumentation__.Notify(243653)
		return resetSingleVar, varName, sessionVar{}, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(243654)
	}
	__antithesis_instrumentation__.Notify(243640)

	switch varName {

	case "database", "role":
		__antithesis_instrumentation__.Notify(243655)
		return unknown, "", sessionVar{}, nil, newCannotChangeParameterError(varName)
	default:
		__antithesis_instrumentation__.Notify(243656)
	}
	__antithesis_instrumentation__.Notify(243641)
	_, sVar, err = getSessionVar(varName, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(243657)
		return unknown, "", sessionVar{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(243658)
	}
	__antithesis_instrumentation__.Notify(243642)

	if sVar.Set == nil {
		__antithesis_instrumentation__.Notify(243659)
		return unknown, "", sessionVar{}, nil, newCannotChangeParameterError(varName)
	} else {
		__antithesis_instrumentation__.Notify(243660)
	}
	__antithesis_instrumentation__.Notify(243643)

	for _, expr := range setOrResetClause.Values {
		__antithesis_instrumentation__.Notify(243661)
		expr = paramparse.UnresolvedNameToStrVal(expr)

		typedValue, err := p.analyzeExpr(
			ctx, expr, nil, tree.IndexedVarHelper{}, types.String, false, "ALTER ROLE ... SET ",
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(243663)
			return unknown, "", sessionVar{}, nil, wrapSetVarError(err, varName, expr.String())
		} else {
			__antithesis_instrumentation__.Notify(243664)
		}
		__antithesis_instrumentation__.Notify(243662)
		typedValues = append(typedValues, typedValue)
	}
	__antithesis_instrumentation__.Notify(243644)

	return setSingleVar, varName, sVar, typedValues, nil
}

func (n *alterRoleSetNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(243665)
	var opName string
	if n.isRole {
		__antithesis_instrumentation__.Notify(243674)
		sqltelemetry.IncIAMAlterCounter(sqltelemetry.Role)
		opName = "alter-role"
	} else {
		__antithesis_instrumentation__.Notify(243675)
		sqltelemetry.IncIAMAlterCounter(sqltelemetry.User)
		opName = "alter-user"
	}
	__antithesis_instrumentation__.Notify(243666)

	needsUpdate, roleName, err := n.getRoleName(params, opName)
	if err != nil {
		__antithesis_instrumentation__.Notify(243676)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243677)
	}
	__antithesis_instrumentation__.Notify(243667)
	if !needsUpdate {
		__antithesis_instrumentation__.Notify(243678)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(243679)
	}
	__antithesis_instrumentation__.Notify(243668)

	var deleteQuery = fmt.Sprintf(
		`DELETE FROM %s WHERE database_id = $1 AND role_name = $2`,
		sessioninit.DatabaseRoleSettingsTableName,
	)
	var upsertQuery = fmt.Sprintf(
		`UPSERT INTO %s (database_id, role_name, settings) VALUES ($1, $2, $3)`,
		sessioninit.DatabaseRoleSettingsTableName,
	)

	upsertOrDeleteFunc := func(newSettings []string) error {
		__antithesis_instrumentation__.Notify(243680)
		var rowsAffected int
		var internalExecErr error
		if newSettings == nil {
			__antithesis_instrumentation__.Notify(243684)
			rowsAffected, internalExecErr = params.extendedEvalCtx.ExecCfg.InternalExecutor.ExecEx(
				params.ctx,
				opName,
				params.p.txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				deleteQuery,
				n.dbDescID,
				roleName,
			)
		} else {
			__antithesis_instrumentation__.Notify(243685)
			rowsAffected, internalExecErr = params.extendedEvalCtx.ExecCfg.InternalExecutor.ExecEx(
				params.ctx,
				opName,
				params.p.txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				upsertQuery,
				n.dbDescID,
				roleName,
				newSettings,
			)
		}
		__antithesis_instrumentation__.Notify(243681)
		if internalExecErr != nil {
			__antithesis_instrumentation__.Notify(243686)
			return internalExecErr
		} else {
			__antithesis_instrumentation__.Notify(243687)
		}
		__antithesis_instrumentation__.Notify(243682)

		if rowsAffected > 0 && func() bool {
			__antithesis_instrumentation__.Notify(243688)
			return sessioninit.CacheEnabled.Get(&params.p.ExecCfg().Settings.SV) == true
		}() == true {
			__antithesis_instrumentation__.Notify(243689)

			if err := params.p.bumpDatabaseRoleSettingsTableVersion(params.ctx); err != nil {
				__antithesis_instrumentation__.Notify(243690)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243691)
			}
		} else {
			__antithesis_instrumentation__.Notify(243692)
		}
		__antithesis_instrumentation__.Notify(243683)
		return params.p.logEvent(params.ctx,
			0,
			&eventpb.AlterRole{
				RoleName: roleName.Normalized(),
				Options:  []string{roleoption.DEFAULTSETTINGS.String()},
			})
	}
	__antithesis_instrumentation__.Notify(243669)

	if n.setVarKind == resetAllVars {
		__antithesis_instrumentation__.Notify(243693)
		return upsertOrDeleteFunc(nil)
	} else {
		__antithesis_instrumentation__.Notify(243694)
	}
	__antithesis_instrumentation__.Notify(243670)

	hasOldSettings, newSettings, err := n.makeNewSettings(params, opName, roleName)
	if err != nil {
		__antithesis_instrumentation__.Notify(243695)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243696)
	}
	__antithesis_instrumentation__.Notify(243671)

	if n.setVarKind == resetSingleVar {
		__antithesis_instrumentation__.Notify(243697)
		if !hasOldSettings {
			__antithesis_instrumentation__.Notify(243699)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(243700)
		}
		__antithesis_instrumentation__.Notify(243698)
		return upsertOrDeleteFunc(newSettings)
	} else {
		__antithesis_instrumentation__.Notify(243701)
	}
	__antithesis_instrumentation__.Notify(243672)

	strVal, err := n.getSessionVarVal(params)
	if err != nil {
		__antithesis_instrumentation__.Notify(243702)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243703)
	}
	__antithesis_instrumentation__.Notify(243673)

	newSetting := fmt.Sprintf("%s=%s", n.varName, strVal)
	newSettings = append(newSettings, newSetting)
	return upsertOrDeleteFunc(newSettings)
}

func (n *alterRoleSetNode) getRoleName(
	params runParams, opName string,
) (needsUpdate bool, retRoleName security.SQLUsername, err error) {
	__antithesis_instrumentation__.Notify(243704)
	if n.allRoles {
		__antithesis_instrumentation__.Notify(243714)
		return true, security.MakeSQLUsernameFromPreNormalizedString(""), nil
	} else {
		__antithesis_instrumentation__.Notify(243715)
	}
	__antithesis_instrumentation__.Notify(243705)
	if n.roleName.Undefined() {
		__antithesis_instrumentation__.Notify(243716)
		return false, security.SQLUsername{}, pgerror.New(pgcode.InvalidParameterValue, "no username specified")
	} else {
		__antithesis_instrumentation__.Notify(243717)
	}
	__antithesis_instrumentation__.Notify(243706)
	if n.roleName.IsAdminRole() {
		__antithesis_instrumentation__.Notify(243718)
		return false, security.SQLUsername{}, pgerror.Newf(pgcode.InsufficientPrivilege, "cannot edit admin role")
	} else {
		__antithesis_instrumentation__.Notify(243719)
	}
	__antithesis_instrumentation__.Notify(243707)
	if n.roleName.IsRootUser() {
		__antithesis_instrumentation__.Notify(243720)
		return false, security.SQLUsername{}, pgerror.Newf(pgcode.InsufficientPrivilege, "cannot edit root user")
	} else {
		__antithesis_instrumentation__.Notify(243721)
	}
	__antithesis_instrumentation__.Notify(243708)
	if n.roleName.IsPublicRole() {
		__antithesis_instrumentation__.Notify(243722)
		return false, security.SQLUsername{}, pgerror.Newf(pgcode.InsufficientPrivilege, "cannot edit public role")
	} else {
		__antithesis_instrumentation__.Notify(243723)
	}
	__antithesis_instrumentation__.Notify(243709)

	row, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.QueryRowEx(
		params.ctx,
		opName,
		params.p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf("SELECT 1 FROM %s WHERE username = $1", sessioninit.UsersTableName),
		n.roleName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(243724)
		return false, security.SQLUsername{}, err
	} else {
		__antithesis_instrumentation__.Notify(243725)
	}
	__antithesis_instrumentation__.Notify(243710)
	if row == nil {
		__antithesis_instrumentation__.Notify(243726)
		if n.ifExists {
			__antithesis_instrumentation__.Notify(243728)
			return false, security.SQLUsername{}, nil
		} else {
			__antithesis_instrumentation__.Notify(243729)
		}
		__antithesis_instrumentation__.Notify(243727)
		return false, security.SQLUsername{}, errors.Newf("role/user %s does not exist", n.roleName)
	} else {
		__antithesis_instrumentation__.Notify(243730)
	}
	__antithesis_instrumentation__.Notify(243711)
	isAdmin, err := params.p.UserHasAdminRole(params.ctx, n.roleName)
	if err != nil {
		__antithesis_instrumentation__.Notify(243731)
		return false, security.SQLUsername{}, err
	} else {
		__antithesis_instrumentation__.Notify(243732)
	}
	__antithesis_instrumentation__.Notify(243712)
	if isAdmin {
		__antithesis_instrumentation__.Notify(243733)
		if err := params.p.RequireAdminRole(params.ctx, "ALTER ROLE admin"); err != nil {
			__antithesis_instrumentation__.Notify(243734)
			return false, security.SQLUsername{}, err
		} else {
			__antithesis_instrumentation__.Notify(243735)
		}
	} else {
		__antithesis_instrumentation__.Notify(243736)
	}
	__antithesis_instrumentation__.Notify(243713)
	return true, n.roleName, nil
}

func (n *alterRoleSetNode) makeNewSettings(
	params runParams, opName string, roleName security.SQLUsername,
) (hasOldSettings bool, newSettings []string, err error) {
	__antithesis_instrumentation__.Notify(243737)
	var selectQuery = fmt.Sprintf(
		`SELECT settings FROM %s WHERE database_id = $1 AND role_name = $2`,
		sessioninit.DatabaseRoleSettingsTableName,
	)
	datums, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.QueryRowEx(
		params.ctx,
		opName,
		params.p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		selectQuery,
		n.dbDescID,
		roleName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(243740)
		return false, nil, err
	} else {
		__antithesis_instrumentation__.Notify(243741)
	}
	__antithesis_instrumentation__.Notify(243738)
	var oldSettings *tree.DArray
	if datums != nil {
		__antithesis_instrumentation__.Notify(243742)
		oldSettings = tree.MustBeDArray(datums[0])
		for _, s := range oldSettings.Array {
			__antithesis_instrumentation__.Notify(243743)
			oldSetting := string(tree.MustBeDString(s))
			keyVal := strings.SplitN(oldSetting, "=", 2)
			if !strings.EqualFold(n.varName, keyVal[0]) {
				__antithesis_instrumentation__.Notify(243744)
				newSettings = append(newSettings, oldSetting)
			} else {
				__antithesis_instrumentation__.Notify(243745)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(243746)
	}
	__antithesis_instrumentation__.Notify(243739)
	return oldSettings != nil, newSettings, nil
}

func (n *alterRoleSetNode) getSessionVarVal(params runParams) (string, error) {
	__antithesis_instrumentation__.Notify(243747)
	if n.varName == "" || func() bool {
		__antithesis_instrumentation__.Notify(243753)
		return n.typedValues == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(243754)
		return "", nil
	} else {
		__antithesis_instrumentation__.Notify(243755)
	}
	__antithesis_instrumentation__.Notify(243748)
	for i, v := range n.typedValues {
		__antithesis_instrumentation__.Notify(243756)
		d, err := v.Eval(params.EvalContext())
		if err != nil {
			__antithesis_instrumentation__.Notify(243758)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(243759)
		}
		__antithesis_instrumentation__.Notify(243757)
		n.typedValues[i] = d
	}
	__antithesis_instrumentation__.Notify(243749)
	var strVal string
	var err error
	if n.sVar.GetStringVal != nil {
		__antithesis_instrumentation__.Notify(243760)
		strVal, err = n.sVar.GetStringVal(params.ctx, params.extendedEvalCtx, n.typedValues)
	} else {
		__antithesis_instrumentation__.Notify(243761)

		strVal, err = getStringVal(params.EvalContext(), n.varName, n.typedValues)
	}
	__antithesis_instrumentation__.Notify(243750)
	if err != nil {
		__antithesis_instrumentation__.Notify(243762)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(243763)
	}
	__antithesis_instrumentation__.Notify(243751)

	if err := CheckSessionVariableValueValid(params.ctx, params.ExecCfg().Settings, n.varName, strVal); err != nil {
		__antithesis_instrumentation__.Notify(243764)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(243765)
	}
	__antithesis_instrumentation__.Notify(243752)
	return strVal, nil
}

func (*alterRoleSetNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(243766)
	return false, nil
}
func (*alterRoleSetNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(243767)
	return tree.Datums{}
}
func (*alterRoleSetNode) Close(context.Context) { __antithesis_instrumentation__.Notify(243768) }
