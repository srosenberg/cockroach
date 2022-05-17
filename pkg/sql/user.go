package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessioninit"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func GetUserSessionInitInfo(
	ctx context.Context,
	execCfg *ExecutorConfig,
	ie *InternalExecutor,
	username security.SQLUsername,
	databaseName string,
) (
	exists bool,
	canLoginSQL bool,
	canLoginDBConsole bool,
	isSuperuser bool,
	defaultSettings []sessioninit.SettingsCacheEntry,
	pwRetrieveFn func(ctx context.Context) (expired bool, hashedPassword security.PasswordHash, err error),
	err error,
) {
	__antithesis_instrumentation__.Notify(631454)
	runFn := getUserInfoRunFn(execCfg, username, "get-user-timeout")

	if username.IsRootUser() {
		__antithesis_instrumentation__.Notify(631457)

		rootFn := func(ctx context.Context) (expired bool, ret security.PasswordHash, err error) {
			__antithesis_instrumentation__.Notify(631459)
			err = runFn(ctx, func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(631462)
				authInfo, _, err := retrieveSessionInitInfoWithCache(ctx, execCfg, ie, username, databaseName)
				if err != nil {
					__antithesis_instrumentation__.Notify(631464)
					return err
				} else {
					__antithesis_instrumentation__.Notify(631465)
				}
				__antithesis_instrumentation__.Notify(631463)
				ret = authInfo.HashedPassword
				return nil
			})
			__antithesis_instrumentation__.Notify(631460)
			if ret == nil {
				__antithesis_instrumentation__.Notify(631466)
				ret = security.MissingPasswordHash
			} else {
				__antithesis_instrumentation__.Notify(631467)
			}
			__antithesis_instrumentation__.Notify(631461)

			return false, ret, err
		}
		__antithesis_instrumentation__.Notify(631458)

		return true, true, true, true, nil, rootFn, nil
	} else {
		__antithesis_instrumentation__.Notify(631468)
	}
	__antithesis_instrumentation__.Notify(631455)

	var authInfo sessioninit.AuthInfo
	var settingsEntries []sessioninit.SettingsCacheEntry

	if err = runFn(ctx, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(631469)

		authInfo, settingsEntries, err = retrieveSessionInitInfoWithCache(
			ctx, execCfg, ie, username, databaseName,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(631471)
			return err
		} else {
			__antithesis_instrumentation__.Notify(631472)
		}
		__antithesis_instrumentation__.Notify(631470)

		return execCfg.CollectionFactory.Txn(
			ctx,
			ie,
			execCfg.DB,
			func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
				__antithesis_instrumentation__.Notify(631473)
				memberships, err := MemberOfWithAdminOption(
					ctx,
					execCfg,
					ie,
					descsCol,
					txn,
					username,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(631475)
					return err
				} else {
					__antithesis_instrumentation__.Notify(631476)
				}
				__antithesis_instrumentation__.Notify(631474)
				_, isSuperuser = memberships[security.AdminRoleName()]
				return nil
			},
		)
	}); err != nil {
		__antithesis_instrumentation__.Notify(631477)
		log.Warningf(ctx, "user membership lookup for %q failed: %v", username, err)
		err = errors.Wrap(errors.Handled(err), "internal error while retrieving user account memberships")
	} else {
		__antithesis_instrumentation__.Notify(631478)
	}
	__antithesis_instrumentation__.Notify(631456)

	return authInfo.UserExists,
		authInfo.CanLoginSQL,
		authInfo.CanLoginDBConsole,
		isSuperuser,
		settingsEntries,
		func(ctx context.Context) (expired bool, ret security.PasswordHash, err error) {
			__antithesis_instrumentation__.Notify(631479)
			ret = authInfo.HashedPassword
			if authInfo.ValidUntil != nil {
				__antithesis_instrumentation__.Notify(631482)

				if authInfo.ValidUntil.Time.Sub(timeutil.Now()) < 0 {
					__antithesis_instrumentation__.Notify(631483)
					expired = true
					ret = nil
				} else {
					__antithesis_instrumentation__.Notify(631484)
				}
			} else {
				__antithesis_instrumentation__.Notify(631485)
			}
			__antithesis_instrumentation__.Notify(631480)
			if ret == nil {
				__antithesis_instrumentation__.Notify(631486)
				ret = security.MissingPasswordHash
			} else {
				__antithesis_instrumentation__.Notify(631487)
			}
			__antithesis_instrumentation__.Notify(631481)
			return expired, ret, nil
		},
		err
}

func getUserInfoRunFn(
	execCfg *ExecutorConfig, username security.SQLUsername, opName string,
) func(context.Context, func(context.Context) error) error {
	__antithesis_instrumentation__.Notify(631488)

	timeout := userLoginTimeout.Get(&execCfg.Settings.SV)

	const maxRootTimeout = 4*time.Second + 500*time.Millisecond
	if username.IsRootUser() && func() bool {
		__antithesis_instrumentation__.Notify(631492)
		return (timeout == 0 || func() bool {
			__antithesis_instrumentation__.Notify(631493)
			return timeout > maxRootTimeout == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(631494)
		timeout = maxRootTimeout
	} else {
		__antithesis_instrumentation__.Notify(631495)
	}
	__antithesis_instrumentation__.Notify(631489)

	runFn := func(ctx context.Context, fn func(ctx context.Context) error) error {
		__antithesis_instrumentation__.Notify(631496)
		return fn(ctx)
	}
	__antithesis_instrumentation__.Notify(631490)
	if timeout != 0 {
		__antithesis_instrumentation__.Notify(631497)
		runFn = func(ctx context.Context, fn func(ctx context.Context) error) error {
			__antithesis_instrumentation__.Notify(631498)
			return contextutil.RunWithTimeout(ctx, opName, timeout, fn)
		}
	} else {
		__antithesis_instrumentation__.Notify(631499)
	}
	__antithesis_instrumentation__.Notify(631491)
	return runFn
}

func retrieveSessionInitInfoWithCache(
	ctx context.Context,
	execCfg *ExecutorConfig,
	ie *InternalExecutor,
	username security.SQLUsername,
	databaseName string,
) (aInfo sessioninit.AuthInfo, settingsEntries []sessioninit.SettingsCacheEntry, err error) {
	__antithesis_instrumentation__.Notify(631500)
	if err = func() (retErr error) {
		__antithesis_instrumentation__.Notify(631502)
		aInfo, retErr = execCfg.SessionInitCache.GetAuthInfo(
			ctx,
			execCfg.Settings,
			ie,
			execCfg.DB,
			execCfg.CollectionFactory,
			username,
			retrieveAuthInfo,
		)
		if retErr != nil {
			__antithesis_instrumentation__.Notify(631505)
			return retErr
		} else {
			__antithesis_instrumentation__.Notify(631506)
		}
		__antithesis_instrumentation__.Notify(631503)

		if username.IsRootUser() || func() bool {
			__antithesis_instrumentation__.Notify(631507)
			return !aInfo.UserExists == true
		}() == true {
			__antithesis_instrumentation__.Notify(631508)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(631509)
		}
		__antithesis_instrumentation__.Notify(631504)
		settingsEntries, retErr = execCfg.SessionInitCache.GetDefaultSettings(
			ctx,
			execCfg.Settings,
			ie,
			execCfg.DB,
			execCfg.CollectionFactory,
			username,
			databaseName,
			retrieveDefaultSettings,
		)
		return retErr
	}(); err != nil {
		__antithesis_instrumentation__.Notify(631510)

		log.Warningf(ctx, "user lookup for %q failed: %v", username, err)
		err = errors.Wrap(errors.Handled(err), "internal error while retrieving user account")
	} else {
		__antithesis_instrumentation__.Notify(631511)
	}
	__antithesis_instrumentation__.Notify(631501)
	return aInfo, settingsEntries, err
}

func retrieveAuthInfo(
	ctx context.Context, ie sqlutil.InternalExecutor, username security.SQLUsername,
) (aInfo sessioninit.AuthInfo, retErr error) {
	__antithesis_instrumentation__.Notify(631512)

	const getHashedPassword = `SELECT "hashedPassword" FROM system.public.users ` +
		`WHERE username=$1`
	values, err := ie.QueryRowEx(
		ctx, "get-hashed-pwd", nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		getHashedPassword, username)
	if err != nil {
		__antithesis_instrumentation__.Notify(631520)
		return aInfo, errors.Wrapf(err, "error looking up user %s", username)
	} else {
		__antithesis_instrumentation__.Notify(631521)
	}
	__antithesis_instrumentation__.Notify(631513)
	var hashedPassword []byte
	if values != nil {
		__antithesis_instrumentation__.Notify(631522)
		aInfo.UserExists = true
		if v := values[0]; v != tree.DNull {
			__antithesis_instrumentation__.Notify(631523)
			hashedPassword = []byte(*(v.(*tree.DBytes)))
		} else {
			__antithesis_instrumentation__.Notify(631524)
		}
	} else {
		__antithesis_instrumentation__.Notify(631525)
	}
	__antithesis_instrumentation__.Notify(631514)
	aInfo.HashedPassword = security.LoadPasswordHash(ctx, hashedPassword)

	if !aInfo.UserExists {
		__antithesis_instrumentation__.Notify(631526)
		return aInfo, nil
	} else {
		__antithesis_instrumentation__.Notify(631527)
	}
	__antithesis_instrumentation__.Notify(631515)

	if username.IsRootUser() {
		__antithesis_instrumentation__.Notify(631528)
		return aInfo, nil
	} else {
		__antithesis_instrumentation__.Notify(631529)
	}
	__antithesis_instrumentation__.Notify(631516)

	const getLoginDependencies = `SELECT option, value FROM system.public.role_options ` +
		`WHERE username=$1 AND option IN ('NOLOGIN', 'VALID UNTIL', 'NOSQLLOGIN')`

	roleOptsIt, err := ie.QueryIteratorEx(
		ctx, "get-login-dependencies", nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		getLoginDependencies,
		username,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(631530)
		return aInfo, errors.Wrapf(err, "error looking up user %s", username)
	} else {
		__antithesis_instrumentation__.Notify(631531)
	}
	__antithesis_instrumentation__.Notify(631517)

	defer func() {
		__antithesis_instrumentation__.Notify(631532)
		retErr = errors.CombineErrors(retErr, roleOptsIt.Close())
	}()
	__antithesis_instrumentation__.Notify(631518)

	aInfo.CanLoginSQL = true
	aInfo.CanLoginDBConsole = true
	var ok bool
	for ok, err = roleOptsIt.Next(ctx); ok; ok, err = roleOptsIt.Next(ctx) {
		__antithesis_instrumentation__.Notify(631533)
		row := roleOptsIt.Cur()
		option := string(tree.MustBeDString(row[0]))

		if option == "NOLOGIN" {
			__antithesis_instrumentation__.Notify(631536)
			aInfo.CanLoginSQL = false
			aInfo.CanLoginDBConsole = false
		} else {
			__antithesis_instrumentation__.Notify(631537)
		}
		__antithesis_instrumentation__.Notify(631534)
		if option == "NOSQLLOGIN" {
			__antithesis_instrumentation__.Notify(631538)
			aInfo.CanLoginSQL = false
		} else {
			__antithesis_instrumentation__.Notify(631539)
		}
		__antithesis_instrumentation__.Notify(631535)

		if option == "VALID UNTIL" {
			__antithesis_instrumentation__.Notify(631540)
			if tree.DNull.Compare(nil, row[1]) != 0 {
				__antithesis_instrumentation__.Notify(631541)
				ts := string(tree.MustBeDString(row[1]))

				timeCtx := tree.NewParseTimeContext(timeutil.Now())
				aInfo.ValidUntil, _, err = tree.ParseDTimestamp(timeCtx, ts, time.Microsecond)
				if err != nil {
					__antithesis_instrumentation__.Notify(631542)
					return aInfo, errors.Wrap(err,
						"error trying to parse timestamp while retrieving password valid until value")
				} else {
					__antithesis_instrumentation__.Notify(631543)
				}
			} else {
				__antithesis_instrumentation__.Notify(631544)
			}
		} else {
			__antithesis_instrumentation__.Notify(631545)
		}
	}
	__antithesis_instrumentation__.Notify(631519)

	return aInfo, err
}

func retrieveDefaultSettings(
	ctx context.Context,
	ie sqlutil.InternalExecutor,
	username security.SQLUsername,
	databaseID descpb.ID,
) (settingsEntries []sessioninit.SettingsCacheEntry, retErr error) {
	__antithesis_instrumentation__.Notify(631546)

	keys := sessioninit.GenerateSettingsCacheKeys(databaseID, username)
	settingsEntries = make([]sessioninit.SettingsCacheEntry, len(keys))
	for i, k := range keys {
		__antithesis_instrumentation__.Notify(631552)
		settingsEntries[i] = sessioninit.SettingsCacheEntry{
			SettingsCacheKey: k,
			Settings:         []string{},
		}
	}
	__antithesis_instrumentation__.Notify(631547)

	if username.IsRootUser() {
		__antithesis_instrumentation__.Notify(631553)
		return settingsEntries, nil
	} else {
		__antithesis_instrumentation__.Notify(631554)
	}
	__antithesis_instrumentation__.Notify(631548)

	const getDefaultSettings = `
SELECT
  database_id, role_name, settings
FROM
  system.public.database_role_settings
WHERE
  (database_id = 0 AND role_name = $1)
  OR (database_id = $2 AND role_name = $1)
  OR (database_id = $2 AND role_name = '')
  OR (database_id = 0 AND role_name = '');
`

	defaultSettingsIt, err := ie.QueryIteratorEx(
		ctx, "get-default-settings", nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		getDefaultSettings,
		username,
		databaseID,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(631555)
		return nil, errors.Wrapf(err, "error looking up user %s", username)
	} else {
		__antithesis_instrumentation__.Notify(631556)
	}
	__antithesis_instrumentation__.Notify(631549)

	defer func() {
		__antithesis_instrumentation__.Notify(631557)
		retErr = errors.CombineErrors(retErr, defaultSettingsIt.Close())
	}()
	__antithesis_instrumentation__.Notify(631550)

	var ok bool
	for ok, err = defaultSettingsIt.Next(ctx); ok; ok, err = defaultSettingsIt.Next(ctx) {
		__antithesis_instrumentation__.Notify(631558)
		row := defaultSettingsIt.Cur()
		fetechedDatabaseID := descpb.ID(tree.MustBeDOid(row[0]).DInt)
		fetchedUsername := security.MakeSQLUsernameFromPreNormalizedString(string(tree.MustBeDString(row[1])))
		settingsDatum := tree.MustBeDArray(row[2])
		fetchedSettings := make([]string, settingsDatum.Len())
		for i, s := range settingsDatum.Array {
			__antithesis_instrumentation__.Notify(631560)
			fetchedSettings[i] = string(tree.MustBeDString(s))
		}
		__antithesis_instrumentation__.Notify(631559)

		thisKey := sessioninit.SettingsCacheKey{
			DatabaseID: fetechedDatabaseID,
			Username:   fetchedUsername,
		}

		for i, s := range settingsEntries {
			__antithesis_instrumentation__.Notify(631561)
			if s.SettingsCacheKey == thisKey {
				__antithesis_instrumentation__.Notify(631562)
				settingsEntries[i].Settings = fetchedSettings
			} else {
				__antithesis_instrumentation__.Notify(631563)
			}
		}
	}
	__antithesis_instrumentation__.Notify(631551)

	return settingsEntries, err
}

var userLoginTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"server.user_login.timeout",
	"timeout after which client authentication times out if some system range is unavailable (0 = no timeout)",
	10*time.Second,
	settings.NonNegativeDuration,
).WithPublic()

func (p *planner) GetAllRoles(ctx context.Context) (map[security.SQLUsername]bool, error) {
	__antithesis_instrumentation__.Notify(631564)
	query := `SELECT username FROM system.users`
	it, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryIteratorEx(
		ctx, "read-users", p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		query)
	if err != nil {
		__antithesis_instrumentation__.Notify(631568)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(631569)
	}
	__antithesis_instrumentation__.Notify(631565)

	users := make(map[security.SQLUsername]bool)
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(631570)
		username := tree.MustBeDString(it.Cur()[0])

		users[security.MakeSQLUsernameFromPreNormalizedString(string(username))] = true
	}
	__antithesis_instrumentation__.Notify(631566)
	if err != nil {
		__antithesis_instrumentation__.Notify(631571)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(631572)
	}
	__antithesis_instrumentation__.Notify(631567)
	return users, nil
}

func (p *planner) RoleExists(ctx context.Context, role security.SQLUsername) (bool, error) {
	__antithesis_instrumentation__.Notify(631573)
	return RoleExists(ctx, p.ExecCfg(), p.Txn(), role)
}

func RoleExists(
	ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, role security.SQLUsername,
) (bool, error) {
	__antithesis_instrumentation__.Notify(631574)
	query := `SELECT username FROM system.users WHERE username = $1`
	row, err := execCfg.InternalExecutor.QueryRowEx(
		ctx, "read-users", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		query, role,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(631576)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(631577)
	}
	__antithesis_instrumentation__.Notify(631575)

	return row != nil, nil
}

var roleMembersTableName = tree.MakeTableNameWithSchema("system", tree.PublicSchemaName, "role_members")

func (p *planner) BumpRoleMembershipTableVersion(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(631578)
	_, tableDesc, err := p.ResolveMutableTableDescriptor(ctx, &roleMembersTableName, true, tree.ResolveAnyTableKind)
	if err != nil {
		__antithesis_instrumentation__.Notify(631580)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631581)
	}
	__antithesis_instrumentation__.Notify(631579)

	return p.writeSchemaChange(
		ctx, tableDesc, descpb.InvalidMutationID, "updating version for role membership table",
	)
}

func (p *planner) bumpUsersTableVersion(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(631582)
	_, tableDesc, err := p.ResolveMutableTableDescriptor(ctx, sessioninit.UsersTableName, true, tree.ResolveAnyTableKind)
	if err != nil {
		__antithesis_instrumentation__.Notify(631584)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631585)
	}
	__antithesis_instrumentation__.Notify(631583)

	return p.writeSchemaChange(
		ctx, tableDesc, descpb.InvalidMutationID, "updating version for users table",
	)
}

func (p *planner) bumpRoleOptionsTableVersion(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(631586)
	_, tableDesc, err := p.ResolveMutableTableDescriptor(ctx, sessioninit.RoleOptionsTableName, true, tree.ResolveAnyTableKind)
	if err != nil {
		__antithesis_instrumentation__.Notify(631588)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631589)
	}
	__antithesis_instrumentation__.Notify(631587)

	return p.writeSchemaChange(
		ctx, tableDesc, descpb.InvalidMutationID, "updating version for role options table",
	)
}

func (p *planner) bumpDatabaseRoleSettingsTableVersion(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(631590)
	_, tableDesc, err := p.ResolveMutableTableDescriptor(ctx, sessioninit.DatabaseRoleSettingsTableName, true, tree.ResolveAnyTableKind)
	if err != nil {
		__antithesis_instrumentation__.Notify(631592)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631593)
	}
	__antithesis_instrumentation__.Notify(631591)

	return p.writeSchemaChange(
		ctx, tableDesc, descpb.InvalidMutationID, "updating version for database_role_settings table",
	)
}

func (p *planner) setRole(ctx context.Context, local bool, s security.SQLUsername) error {
	__antithesis_instrumentation__.Notify(631594)
	sessionUser := p.SessionData().SessionUser()
	becomeUser := sessionUser

	if !s.IsNoneRole() {
		__antithesis_instrumentation__.Notify(631599)
		becomeUser = s

		exists, err := p.RoleExists(ctx, becomeUser)
		if err != nil {
			__antithesis_instrumentation__.Notify(631601)
			return err
		} else {
			__antithesis_instrumentation__.Notify(631602)
		}
		__antithesis_instrumentation__.Notify(631600)
		if !exists {
			__antithesis_instrumentation__.Notify(631603)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"role %s does not exist",
				becomeUser.Normalized(),
			)
		} else {
			__antithesis_instrumentation__.Notify(631604)
		}
	} else {
		__antithesis_instrumentation__.Notify(631605)
	}
	__antithesis_instrumentation__.Notify(631595)

	if err := p.checkCanBecomeUser(ctx, becomeUser); err != nil {
		__antithesis_instrumentation__.Notify(631606)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631607)
	}
	__antithesis_instrumentation__.Notify(631596)

	updateStr := "off"
	willBecomeAdmin, err := p.UserHasAdminRole(ctx, becomeUser)
	if err != nil {
		__antithesis_instrumentation__.Notify(631608)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631609)
	}
	__antithesis_instrumentation__.Notify(631597)
	if willBecomeAdmin {
		__antithesis_instrumentation__.Notify(631610)
		updateStr = "on"
	} else {
		__antithesis_instrumentation__.Notify(631611)
	}
	__antithesis_instrumentation__.Notify(631598)

	return p.applyOnSessionDataMutators(
		ctx,
		local,
		func(m sessionDataMutator) error {
			__antithesis_instrumentation__.Notify(631612)
			m.data.IsSuperuser = willBecomeAdmin
			m.bufferParamStatusUpdate("is_superuser", updateStr)

			if becomeUser.IsNoneRole() {
				__antithesis_instrumentation__.Notify(631615)
				if m.data.SessionUserProto.Decode().Normalized() != "" {
					__antithesis_instrumentation__.Notify(631617)
					m.data.UserProto = m.data.SessionUserProto
					m.data.SessionUserProto = ""
				} else {
					__antithesis_instrumentation__.Notify(631618)
				}
				__antithesis_instrumentation__.Notify(631616)
				m.data.SearchPath = m.data.SearchPath.WithUserSchemaName(m.data.User().Normalized())
				return nil
			} else {
				__antithesis_instrumentation__.Notify(631619)
			}
			__antithesis_instrumentation__.Notify(631613)

			if m.data.SessionUserProto == "" {
				__antithesis_instrumentation__.Notify(631620)
				m.data.SessionUserProto = m.data.UserProto
			} else {
				__antithesis_instrumentation__.Notify(631621)
			}
			__antithesis_instrumentation__.Notify(631614)
			m.data.UserProto = becomeUser.EncodeProto()
			m.data.SearchPath = m.data.SearchPath.WithUserSchemaName(m.data.User().Normalized())
			return nil
		},
	)

}

func (p *planner) checkCanBecomeUser(ctx context.Context, becomeUser security.SQLUsername) error {
	__antithesis_instrumentation__.Notify(631622)
	sessionUser := p.SessionData().SessionUser()

	if becomeUser.IsNoneRole() {
		__antithesis_instrumentation__.Notify(631630)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(631631)
	}
	__antithesis_instrumentation__.Notify(631623)

	if sessionUser.IsRootUser() {
		__antithesis_instrumentation__.Notify(631632)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(631633)
	}
	__antithesis_instrumentation__.Notify(631624)

	if becomeUser.Normalized() == sessionUser.Normalized() {
		__antithesis_instrumentation__.Notify(631634)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(631635)
	}
	__antithesis_instrumentation__.Notify(631625)

	if becomeUser.IsRootUser() {
		__antithesis_instrumentation__.Notify(631636)
		return pgerror.Newf(
			pgcode.InsufficientPrivilege,
			"only root can become root",
		)
	} else {
		__antithesis_instrumentation__.Notify(631637)
	}
	__antithesis_instrumentation__.Notify(631626)

	memberships, err := p.MemberOfWithAdminOption(ctx, sessionUser)
	if err != nil {
		__antithesis_instrumentation__.Notify(631638)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631639)
	}
	__antithesis_instrumentation__.Notify(631627)

	if _, ok := memberships[security.AdminRoleName()]; ok {
		__antithesis_instrumentation__.Notify(631640)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(631641)
	}
	__antithesis_instrumentation__.Notify(631628)

	if _, ok := memberships[becomeUser]; !ok {
		__antithesis_instrumentation__.Notify(631642)
		return pgerror.Newf(
			pgcode.InsufficientPrivilege,
			`permission denied to set role "%s"`,
			becomeUser.Normalized(),
		)
	} else {
		__antithesis_instrumentation__.Notify(631643)
	}
	__antithesis_instrumentation__.Notify(631629)
	return nil
}

func MaybeUpgradeStoredPasswordHash(
	ctx context.Context,
	execCfg *ExecutorConfig,
	username security.SQLUsername,
	cleartext string,
	currentHash security.PasswordHash,
) {
	__antithesis_instrumentation__.Notify(631644)

	converted, prevHash, newHash, newMethod, err := security.MaybeUpgradePasswordHash(ctx, &execCfg.Settings.SV, cleartext, currentHash)
	if err != nil {
		__antithesis_instrumentation__.Notify(631646)

		log.Warningf(ctx, "password hash conversion failed: %+v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(631647)
		if !converted {
			__antithesis_instrumentation__.Notify(631648)

			return
		} else {
			__antithesis_instrumentation__.Notify(631649)
		}
	}
	__antithesis_instrumentation__.Notify(631645)

	if err := updateUserPasswordHash(ctx, execCfg, username, prevHash, newHash); err != nil {
		__antithesis_instrumentation__.Notify(631650)

		log.Warningf(ctx, "storing the new password hash after conversion failed: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(631651)

		log.StructuredEvent(ctx, &eventpb.PasswordHashConverted{
			RoleName:  username.Normalized(),
			OldMethod: currentHash.Method().String(),
			NewMethod: newMethod,
		})
	}
}

func updateUserPasswordHash(
	ctx context.Context,
	execCfg *ExecutorConfig,
	username security.SQLUsername,
	prevHash, newHash []byte,
) error {
	__antithesis_instrumentation__.Notify(631652)
	runFn := getUserInfoRunFn(execCfg, username, "set-hash-timeout")

	return runFn(ctx, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(631653)
		return DescsTxn(ctx, execCfg, func(ctx context.Context, txn *kv.Txn, d *descs.Collection) error {
			__antithesis_instrumentation__.Notify(631654)

			rowsAffected, err := execCfg.InternalExecutor.Exec(
				ctx,
				"set-password-hash",
				txn,
				`UPDATE system.users SET "hashedPassword" = $3 WHERE username = $1 AND "hashedPassword" = $2`,
				username.Normalized(),
				prevHash,
				newHash,
			)
			if err != nil || func() bool {
				__antithesis_instrumentation__.Notify(631657)
				return rowsAffected == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(631658)

				return err
			} else {
				__antithesis_instrumentation__.Notify(631659)
			}
			__antithesis_instrumentation__.Notify(631655)
			usersTable, err := d.GetMutableTableByID(
				ctx, txn, keys.UsersTableID, tree.ObjectLookupFlagsWithRequired(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(631660)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631661)
			}
			__antithesis_instrumentation__.Notify(631656)

			return d.WriteDesc(ctx, false, usersTable, txn)
		})
	})
}
