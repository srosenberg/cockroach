package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessioninit"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

type MembershipCache struct {
	syncutil.Mutex
	tableVersion descpb.DescriptorVersion
	boundAccount mon.BoundAccount

	userCache map[security.SQLUsername]userRoleMembership

	populateCacheGroup singleflight.Group
	stopper            *stop.Stopper
}

func NewMembershipCache(account mon.BoundAccount, stopper *stop.Stopper) *MembershipCache {
	__antithesis_instrumentation__.Notify(245325)
	return &MembershipCache{
		boundAccount: account,
		stopper:      stopper,
	}
}

type userRoleMembership map[security.SQLUsername]bool

type AuthorizationAccessor interface {
	CheckPrivilegeForUser(
		ctx context.Context, descriptor catalog.Descriptor, privilege privilege.Kind, user security.SQLUsername,
	) error

	CheckPrivilege(
		ctx context.Context, descriptor catalog.Descriptor, privilege privilege.Kind,
	) error

	CheckAnyPrivilege(ctx context.Context, descriptor catalog.Descriptor) error

	UserHasAdminRole(ctx context.Context, user security.SQLUsername) (bool, error)

	HasAdminRole(ctx context.Context) (bool, error)

	RequireAdminRole(ctx context.Context, action string) error

	MemberOfWithAdminOption(ctx context.Context, member security.SQLUsername) (map[security.SQLUsername]bool, error)

	HasRoleOption(ctx context.Context, roleOption roleoption.Option) (bool, error)
}

var _ AuthorizationAccessor = &planner{}

func (p *planner) CheckPrivilegeForUser(
	ctx context.Context,
	descriptor catalog.Descriptor,
	privilege privilege.Kind,
	user security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(245326)

	if p.txn == nil || func() bool {
		__antithesis_instrumentation__.Notify(245332)
		return !p.txn.IsOpen() == true
	}() == true {
		__antithesis_instrumentation__.Notify(245333)
		return errors.AssertionFailedf("cannot use CheckPrivilege without a txn")
	} else {
		__antithesis_instrumentation__.Notify(245334)
	}
	__antithesis_instrumentation__.Notify(245327)

	p.maybeAudit(descriptor, privilege)

	privs := descriptor.GetPrivileges()

	if privs.CheckPrivilege(security.PublicRoleName(), privilege) {
		__antithesis_instrumentation__.Notify(245335)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245336)
	}
	__antithesis_instrumentation__.Notify(245328)

	hasPriv, err := p.checkRolePredicate(ctx, user, func(role security.SQLUsername) bool {
		__antithesis_instrumentation__.Notify(245337)
		return IsOwner(descriptor, role) || func() bool {
			__antithesis_instrumentation__.Notify(245338)
			return privs.CheckPrivilege(role, privilege) == true
		}() == true
	})
	__antithesis_instrumentation__.Notify(245329)
	if err != nil {
		__antithesis_instrumentation__.Notify(245339)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245340)
	}
	__antithesis_instrumentation__.Notify(245330)
	if hasPriv {
		__antithesis_instrumentation__.Notify(245341)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245342)
	}
	__antithesis_instrumentation__.Notify(245331)
	return pgerror.Newf(pgcode.InsufficientPrivilege,
		"user %s does not have %s privilege on %s %s",
		user, privilege, descriptor.DescriptorType(), descriptor.GetName())
}

func (p *planner) CheckPrivilege(
	ctx context.Context, descriptor catalog.Descriptor, privilege privilege.Kind,
) error {
	__antithesis_instrumentation__.Notify(245343)
	return p.CheckPrivilegeForUser(ctx, descriptor, privilege, p.User())
}

func (p *planner) CheckGrantOptionsForUser(
	ctx context.Context,
	descriptor catalog.Descriptor,
	privList privilege.List,
	user security.SQLUsername,
	isGrant bool,
) error {
	__antithesis_instrumentation__.Notify(245344)

	if isAdmin, err := p.UserHasAdminRole(ctx, user); err != nil {
		__antithesis_instrumentation__.Notify(245351)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245352)
		if isAdmin {
			__antithesis_instrumentation__.Notify(245353)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(245354)
		}
	}
	__antithesis_instrumentation__.Notify(245345)

	privs := descriptor.GetPrivileges()
	hasPriv, err := p.checkRolePredicate(ctx, user, func(role security.SQLUsername) bool {
		__antithesis_instrumentation__.Notify(245355)
		return IsOwner(descriptor, role) || func() bool {
			__antithesis_instrumentation__.Notify(245356)
			return privs.CheckGrantOptions(role, privList) == true
		}() == true
	})
	__antithesis_instrumentation__.Notify(245346)
	if err != nil {
		__antithesis_instrumentation__.Notify(245357)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245358)
	}
	__antithesis_instrumentation__.Notify(245347)
	if hasPriv {
		__antithesis_instrumentation__.Notify(245359)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245360)
	}
	__antithesis_instrumentation__.Notify(245348)

	code := pgcode.WarningPrivilegeNotGranted
	if !isGrant {
		__antithesis_instrumentation__.Notify(245361)
		code = pgcode.WarningPrivilegeNotRevoked
	} else {
		__antithesis_instrumentation__.Notify(245362)
	}
	__antithesis_instrumentation__.Notify(245349)
	if privList.Len() > 1 {
		__antithesis_instrumentation__.Notify(245363)
		return pgerror.Newf(
			code, "user %s missing WITH GRANT OPTION privilege on one or more of %s",
			user, privList.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(245364)
	}
	__antithesis_instrumentation__.Notify(245350)
	return pgerror.Newf(
		code, "user %s missing WITH GRANT OPTION privilege on %s",
		user, privList.String(),
	)
}

func getOwnerOfDesc(desc catalog.Descriptor) security.SQLUsername {
	__antithesis_instrumentation__.Notify(245365)

	owner := desc.GetPrivileges().Owner()
	if owner.Undefined() {
		__antithesis_instrumentation__.Notify(245367)

		if catalog.IsSystemDescriptor(desc) {
			__antithesis_instrumentation__.Notify(245368)
			owner = security.NodeUserName()
		} else {
			__antithesis_instrumentation__.Notify(245369)

			owner = security.AdminRoleName()
		}
	} else {
		__antithesis_instrumentation__.Notify(245370)
	}
	__antithesis_instrumentation__.Notify(245366)
	return owner
}

func IsOwner(desc catalog.Descriptor, role security.SQLUsername) bool {
	__antithesis_instrumentation__.Notify(245371)
	return role == getOwnerOfDesc(desc)
}

func (p *planner) HasOwnership(ctx context.Context, descriptor catalog.Descriptor) (bool, error) {
	__antithesis_instrumentation__.Notify(245372)
	user := p.SessionData().User()

	return p.checkRolePredicate(ctx, user, func(role security.SQLUsername) bool {
		__antithesis_instrumentation__.Notify(245373)
		return IsOwner(descriptor, role)
	})
}

func (p *planner) checkRolePredicate(
	ctx context.Context, user security.SQLUsername, predicate func(role security.SQLUsername) bool,
) (bool, error) {
	__antithesis_instrumentation__.Notify(245374)
	if ok := predicate(user); ok {
		__antithesis_instrumentation__.Notify(245378)
		return ok, nil
	} else {
		__antithesis_instrumentation__.Notify(245379)
	}
	__antithesis_instrumentation__.Notify(245375)
	memberOf, err := p.MemberOfWithAdminOption(ctx, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(245380)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(245381)
	}
	__antithesis_instrumentation__.Notify(245376)
	for role := range memberOf {
		__antithesis_instrumentation__.Notify(245382)
		if ok := predicate(role); ok {
			__antithesis_instrumentation__.Notify(245383)
			return ok, nil
		} else {
			__antithesis_instrumentation__.Notify(245384)
		}
	}
	__antithesis_instrumentation__.Notify(245377)
	return false, nil
}

func (p *planner) CheckAnyPrivilege(ctx context.Context, descriptor catalog.Descriptor) error {
	__antithesis_instrumentation__.Notify(245385)

	if p.txn == nil || func() bool {
		__antithesis_instrumentation__.Notify(245392)
		return !p.txn.IsOpen() == true
	}() == true {
		__antithesis_instrumentation__.Notify(245393)
		return errors.AssertionFailedf("cannot use CheckAnyPrivilege without a txn")
	} else {
		__antithesis_instrumentation__.Notify(245394)
	}
	__antithesis_instrumentation__.Notify(245386)

	user := p.SessionData().User()

	if user.IsNodeUser() {
		__antithesis_instrumentation__.Notify(245395)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(245396)
	}
	__antithesis_instrumentation__.Notify(245387)

	privs := descriptor.GetPrivileges()

	if privs.AnyPrivilege(user) {
		__antithesis_instrumentation__.Notify(245397)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245398)
	}
	__antithesis_instrumentation__.Notify(245388)

	if privs.AnyPrivilege(security.PublicRoleName()) {
		__antithesis_instrumentation__.Notify(245399)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245400)
	}
	__antithesis_instrumentation__.Notify(245389)

	memberOf, err := p.MemberOfWithAdminOption(ctx, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(245401)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245402)
	}
	__antithesis_instrumentation__.Notify(245390)

	for role := range memberOf {
		__antithesis_instrumentation__.Notify(245403)
		if privs.AnyPrivilege(role) {
			__antithesis_instrumentation__.Notify(245404)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(245405)
		}
	}
	__antithesis_instrumentation__.Notify(245391)

	return pgerror.Newf(pgcode.InsufficientPrivilege,
		"user %s has no privileges on %s %s",
		p.SessionData().User(), descriptor.DescriptorType(), descriptor.GetName())
}

func (p *planner) UserHasAdminRole(ctx context.Context, user security.SQLUsername) (bool, error) {
	__antithesis_instrumentation__.Notify(245406)
	if user.Undefined() {
		__antithesis_instrumentation__.Notify(245412)
		return false, errors.AssertionFailedf("empty user")
	} else {
		__antithesis_instrumentation__.Notify(245413)
	}
	__antithesis_instrumentation__.Notify(245407)

	if p.txn == nil || func() bool {
		__antithesis_instrumentation__.Notify(245414)
		return !p.txn.IsOpen() == true
	}() == true {
		__antithesis_instrumentation__.Notify(245415)
		return false, errors.AssertionFailedf("cannot use HasAdminRole without a txn")
	} else {
		__antithesis_instrumentation__.Notify(245416)
	}
	__antithesis_instrumentation__.Notify(245408)

	if user.IsAdminRole() || func() bool {
		__antithesis_instrumentation__.Notify(245417)
		return user.IsRootUser() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(245418)
		return user.IsNodeUser() == true
	}() == true {
		__antithesis_instrumentation__.Notify(245419)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(245420)
	}
	__antithesis_instrumentation__.Notify(245409)

	memberOf, err := p.MemberOfWithAdminOption(ctx, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(245421)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(245422)
	}
	__antithesis_instrumentation__.Notify(245410)

	if _, ok := memberOf[security.AdminRoleName()]; ok {
		__antithesis_instrumentation__.Notify(245423)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(245424)
	}
	__antithesis_instrumentation__.Notify(245411)

	return false, nil
}

func (p *planner) HasAdminRole(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(245425)
	return p.UserHasAdminRole(ctx, p.User())
}

func (p *planner) RequireAdminRole(ctx context.Context, action string) error {
	__antithesis_instrumentation__.Notify(245426)
	ok, err := p.HasAdminRole(ctx)

	if err != nil {
		__antithesis_instrumentation__.Notify(245429)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245430)
	}
	__antithesis_instrumentation__.Notify(245427)
	if !ok {
		__antithesis_instrumentation__.Notify(245431)

		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"only users with the admin role are allowed to %s", action)
	} else {
		__antithesis_instrumentation__.Notify(245432)
	}
	__antithesis_instrumentation__.Notify(245428)
	return nil
}

func (p *planner) MemberOfWithAdminOption(
	ctx context.Context, member security.SQLUsername,
) (map[security.SQLUsername]bool, error) {
	__antithesis_instrumentation__.Notify(245433)
	return MemberOfWithAdminOption(
		ctx,
		p.execCfg,
		p.ExecCfg().InternalExecutor,
		p.Descriptors(),
		p.Txn(),
		member,
	)
}

func MemberOfWithAdminOption(
	ctx context.Context,
	execCfg *ExecutorConfig,
	ie sqlutil.InternalExecutor,
	descsCol *descs.Collection,
	txn *kv.Txn,
	member security.SQLUsername,
) (map[security.SQLUsername]bool, error) {
	__antithesis_instrumentation__.Notify(245434)
	if txn == nil || func() bool {
		__antithesis_instrumentation__.Notify(245443)
		return !txn.IsOpen() == true
	}() == true {
		__antithesis_instrumentation__.Notify(245444)
		return nil, errors.AssertionFailedf("cannot use MemberOfWithAdminoption without a txn")
	} else {
		__antithesis_instrumentation__.Notify(245445)
	}
	__antithesis_instrumentation__.Notify(245435)

	roleMembersCache := execCfg.RoleMemberCache

	_, tableDesc, err := descsCol.GetImmutableTableByName(
		ctx,
		txn,
		&roleMembersTableName,
		tree.ObjectLookupFlagsWithRequired(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(245446)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(245447)
	}
	__antithesis_instrumentation__.Notify(245436)

	tableVersion := tableDesc.GetVersion()
	if tableDesc.IsUncommittedVersion() {
		__antithesis_instrumentation__.Notify(245448)
		return resolveMemberOfWithAdminOption(ctx, member, ie, txn, useSingleQueryForRoleMembershipCache.Get(execCfg.SV()))
	} else {
		__antithesis_instrumentation__.Notify(245449)
	}
	__antithesis_instrumentation__.Notify(245437)

	userMapping, found := func() (userRoleMembership, bool) {
		__antithesis_instrumentation__.Notify(245450)
		roleMembersCache.Lock()
		defer roleMembersCache.Unlock()
		if roleMembersCache.tableVersion < tableVersion {
			__antithesis_instrumentation__.Notify(245452)

			roleMembersCache.tableVersion = tableVersion
			roleMembersCache.userCache = make(map[security.SQLUsername]userRoleMembership)
			roleMembersCache.boundAccount.Empty(ctx)
		} else {
			__antithesis_instrumentation__.Notify(245453)
			if roleMembersCache.tableVersion > tableVersion {
				__antithesis_instrumentation__.Notify(245454)

				return nil, false
			} else {
				__antithesis_instrumentation__.Notify(245455)
			}
		}
		__antithesis_instrumentation__.Notify(245451)
		userMapping, ok := roleMembersCache.userCache[member]
		return userMapping, ok
	}()
	__antithesis_instrumentation__.Notify(245438)

	if found {
		__antithesis_instrumentation__.Notify(245456)

		return userMapping, nil
	} else {
		__antithesis_instrumentation__.Notify(245457)
	}
	__antithesis_instrumentation__.Notify(245439)

	ch, _ := roleMembersCache.populateCacheGroup.DoChan(
		fmt.Sprintf("%s-%d", member.Normalized(), tableVersion),
		func() (interface{}, error) {
			__antithesis_instrumentation__.Notify(245458)

			ctx, cancel := roleMembersCache.stopper.WithCancelOnQuiesce(
				logtags.WithTags(context.Background(), logtags.FromContext(ctx)))
			defer cancel()
			return resolveMemberOfWithAdminOption(
				ctx, member, ie, txn,
				useSingleQueryForRoleMembershipCache.Get(execCfg.SV()),
			)
		},
	)
	__antithesis_instrumentation__.Notify(245440)
	var memberships map[security.SQLUsername]bool
	select {
	case res := <-ch:
		__antithesis_instrumentation__.Notify(245459)
		if res.Err != nil {
			__antithesis_instrumentation__.Notify(245462)
			return nil, res.Err
		} else {
			__antithesis_instrumentation__.Notify(245463)
		}
		__antithesis_instrumentation__.Notify(245460)
		memberships = res.Val.(map[security.SQLUsername]bool)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(245461)
		return nil, ctx.Err()
	}
	__antithesis_instrumentation__.Notify(245441)

	func() {
		__antithesis_instrumentation__.Notify(245464)

		roleMembersCache.Lock()
		defer roleMembersCache.Unlock()
		if roleMembersCache.tableVersion != tableVersion {
			__antithesis_instrumentation__.Notify(245467)

			return
		} else {
			__antithesis_instrumentation__.Notify(245468)
		}
		__antithesis_instrumentation__.Notify(245465)

		sizeOfEntry := int64(len(member.Normalized()))
		for m := range memberships {
			__antithesis_instrumentation__.Notify(245469)
			sizeOfEntry += int64(len(m.Normalized()))
			sizeOfEntry += memsize.Bool
		}
		__antithesis_instrumentation__.Notify(245466)
		if err := roleMembersCache.boundAccount.Grow(ctx, sizeOfEntry); err != nil {
			__antithesis_instrumentation__.Notify(245470)

			log.Ops.Warningf(ctx, "no memory available to cache role membership info: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(245471)
			roleMembersCache.userCache[member] = memberships
		}
	}()
	__antithesis_instrumentation__.Notify(245442)
	return memberships, nil
}

var defaultSingleQueryForRoleMembershipCache = util.ConstantWithMetamorphicTestBool(
	"resolve-membership-single-scan-enabled",
	true,
)

var useSingleQueryForRoleMembershipCache = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.auth.resolve_membership_single_scan.enabled",
	"determines whether to populate the role membership cache with a single scan",
	defaultSingleQueryForRoleMembershipCache,
).WithPublic()

func resolveMemberOfWithAdminOption(
	ctx context.Context,
	member security.SQLUsername,
	ie sqlutil.InternalExecutor,
	txn *kv.Txn,
	singleQuery bool,
) (map[security.SQLUsername]bool, error) {
	__antithesis_instrumentation__.Notify(245472)
	ret := map[security.SQLUsername]bool{}
	if singleQuery {
		__antithesis_instrumentation__.Notify(245475)
		type membership struct {
			role    security.SQLUsername
			isAdmin bool
		}
		memberToRoles := make(map[security.SQLUsername][]membership)
		if err := forEachRoleMembership(ctx, ie, txn, func(role, member security.SQLUsername, isAdmin bool) error {
			__antithesis_instrumentation__.Notify(245478)
			memberToRoles[member] = append(memberToRoles[member], membership{role, isAdmin})
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(245479)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(245480)
		}
		__antithesis_instrumentation__.Notify(245476)

		var recurse func(u security.SQLUsername)
		recurse = func(u security.SQLUsername) {
			__antithesis_instrumentation__.Notify(245481)
			for _, membership := range memberToRoles[u] {
				__antithesis_instrumentation__.Notify(245482)

				prev, alreadySeen := ret[membership.role]
				ret[membership.role] = prev || func() bool {
					__antithesis_instrumentation__.Notify(245483)
					return membership.isAdmin == true
				}() == true
				if !alreadySeen {
					__antithesis_instrumentation__.Notify(245484)
					recurse(membership.role)
				} else {
					__antithesis_instrumentation__.Notify(245485)
				}
			}
		}
		__antithesis_instrumentation__.Notify(245477)
		recurse(member)
		return ret, nil
	} else {
		__antithesis_instrumentation__.Notify(245486)
	}
	__antithesis_instrumentation__.Notify(245473)

	visited := map[security.SQLUsername]struct{}{}
	toVisit := []security.SQLUsername{member}
	lookupRolesStmt := `SELECT "role", "isAdmin" FROM system.role_members WHERE "member" = $1`

	for len(toVisit) > 0 {
		__antithesis_instrumentation__.Notify(245487)

		m := toVisit[0]
		toVisit = toVisit[1:]
		if _, ok := visited[m]; ok {
			__antithesis_instrumentation__.Notify(245491)
			continue
		} else {
			__antithesis_instrumentation__.Notify(245492)
		}
		__antithesis_instrumentation__.Notify(245488)
		visited[m] = struct{}{}

		it, err := ie.QueryIterator(
			ctx, "expand-roles", txn, lookupRolesStmt, m.Normalized(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(245493)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(245494)
		}
		__antithesis_instrumentation__.Notify(245489)

		var ok bool
		for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
			__antithesis_instrumentation__.Notify(245495)
			row := it.Cur()
			roleName := tree.MustBeDString(row[0])
			isAdmin := row[1].(*tree.DBool)

			role := security.MakeSQLUsernameFromPreNormalizedString(string(roleName))
			ret[role] = bool(*isAdmin)

			toVisit = append(toVisit, role)
		}
		__antithesis_instrumentation__.Notify(245490)
		if err != nil {
			__antithesis_instrumentation__.Notify(245496)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(245497)
		}
	}
	__antithesis_instrumentation__.Notify(245474)

	return ret, nil
}

func (p *planner) HasRoleOption(ctx context.Context, roleOption roleoption.Option) (bool, error) {
	__antithesis_instrumentation__.Notify(245498)

	if p.txn == nil || func() bool {
		__antithesis_instrumentation__.Notify(245505)
		return !p.txn.IsOpen() == true
	}() == true {
		__antithesis_instrumentation__.Notify(245506)
		return false, errors.AssertionFailedf("cannot use HasRoleOption without a txn")
	} else {
		__antithesis_instrumentation__.Notify(245507)
	}
	__antithesis_instrumentation__.Notify(245499)

	user := p.SessionData().User()
	if user.IsRootUser() || func() bool {
		__antithesis_instrumentation__.Notify(245508)
		return user.IsNodeUser() == true
	}() == true {
		__antithesis_instrumentation__.Notify(245509)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(245510)
	}
	__antithesis_instrumentation__.Notify(245500)

	hasAdmin, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(245511)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(245512)
	}
	__antithesis_instrumentation__.Notify(245501)
	if hasAdmin {
		__antithesis_instrumentation__.Notify(245513)

		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(245514)
	}
	__antithesis_instrumentation__.Notify(245502)

	hasRolePrivilege, err := p.ExecCfg().InternalExecutor.QueryRowEx(
		ctx, "has-role-option", p.Txn(),
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf(
			`SELECT 1 from %s WHERE option = '%s' AND username = $1 LIMIT 1`,
			sessioninit.RoleOptionsTableName, roleOption.String()), user.Normalized())
	if err != nil {
		__antithesis_instrumentation__.Notify(245515)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(245516)
	}
	__antithesis_instrumentation__.Notify(245503)

	if len(hasRolePrivilege) != 0 {
		__antithesis_instrumentation__.Notify(245517)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(245518)
	}
	__antithesis_instrumentation__.Notify(245504)

	return false, nil
}

func (p *planner) CheckRoleOption(ctx context.Context, roleOption roleoption.Option) error {
	__antithesis_instrumentation__.Notify(245519)
	hasRoleOption, err := p.HasRoleOption(ctx, roleOption)
	if err != nil {
		__antithesis_instrumentation__.Notify(245522)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245523)
	}
	__antithesis_instrumentation__.Notify(245520)

	if !hasRoleOption {
		__antithesis_instrumentation__.Notify(245524)
		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"user %s does not have %s privilege", p.User(), roleOption)
	} else {
		__antithesis_instrumentation__.Notify(245525)
	}
	__antithesis_instrumentation__.Notify(245521)

	return nil
}

const ConnAuditingClusterSettingName = "server.auth_log.sql_connections.enabled"

const AuthAuditingClusterSettingName = "server.auth_log.sql_sessions.enabled"

type shouldCheckPublicSchema bool

const (
	checkPublicSchema     shouldCheckPublicSchema = true
	skipCheckPublicSchema shouldCheckPublicSchema = false
)

func (p *planner) canCreateOnSchema(
	ctx context.Context,
	schemaID descpb.ID,
	dbID descpb.ID,
	user security.SQLUsername,
	checkPublicSchema shouldCheckPublicSchema,
) error {
	__antithesis_instrumentation__.Notify(245526)
	scDesc, err := p.Descriptors().GetImmutableSchemaByID(
		ctx, p.Txn(), schemaID, tree.SchemaLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(245528)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245529)
	}
	__antithesis_instrumentation__.Notify(245527)

	switch kind := scDesc.SchemaKind(); kind {
	case catalog.SchemaPublic:
		__antithesis_instrumentation__.Notify(245530)

		if !checkPublicSchema {
			__antithesis_instrumentation__.Notify(245537)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(245538)
		}
		__antithesis_instrumentation__.Notify(245531)
		_, dbDesc, err := p.Descriptors().GetImmutableDatabaseByID(
			ctx, p.Txn(), dbID, tree.DatabaseLookupFlags{Required: true})
		if err != nil {
			__antithesis_instrumentation__.Notify(245539)
			return err
		} else {
			__antithesis_instrumentation__.Notify(245540)
		}
		__antithesis_instrumentation__.Notify(245532)
		return p.CheckPrivilegeForUser(ctx, dbDesc, privilege.CREATE, user)
	case catalog.SchemaTemporary:
		__antithesis_instrumentation__.Notify(245533)

		return nil
	case catalog.SchemaVirtual:
		__antithesis_instrumentation__.Notify(245534)
		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"cannot CREATE on schema %s", scDesc.GetName())
	case catalog.SchemaUserDefined:
		__antithesis_instrumentation__.Notify(245535)
		return p.CheckPrivilegeForUser(ctx, scDesc, privilege.CREATE, user)
	default:
		__antithesis_instrumentation__.Notify(245536)
		panic(errors.AssertionFailedf("unknown schema kind %d", kind))
	}
}

func (p *planner) canResolveDescUnderSchema(
	ctx context.Context, scDesc catalog.SchemaDescriptor, desc catalog.Descriptor,
) error {
	__antithesis_instrumentation__.Notify(245541)

	if tbl, ok := desc.(catalog.TableDescriptor); ok && func() bool {
		__antithesis_instrumentation__.Notify(245543)
		return tbl.IsTemporary() == true
	}() == true {
		__antithesis_instrumentation__.Notify(245544)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245545)
	}
	__antithesis_instrumentation__.Notify(245542)

	switch kind := scDesc.SchemaKind(); kind {
	case catalog.SchemaPublic, catalog.SchemaTemporary, catalog.SchemaVirtual:
		__antithesis_instrumentation__.Notify(245546)

		return nil
	case catalog.SchemaUserDefined:
		__antithesis_instrumentation__.Notify(245547)
		return p.CheckPrivilegeForUser(ctx, scDesc, privilege.USAGE, p.User())
	default:
		__antithesis_instrumentation__.Notify(245548)
		panic(errors.AssertionFailedf("unknown schema kind %d", kind))
	}
}

func (p *planner) checkCanAlterToNewOwner(
	ctx context.Context, desc catalog.MutableDescriptor, newOwner security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(245549)

	roleExists, err := RoleExists(ctx, p.ExecCfg(), p.Txn(), newOwner)
	if err != nil {
		__antithesis_instrumentation__.Notify(245560)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245561)
	}
	__antithesis_instrumentation__.Notify(245550)
	if !roleExists {
		__antithesis_instrumentation__.Notify(245562)
		return pgerror.Newf(pgcode.UndefinedObject, "role/user %q does not exist", newOwner)
	} else {
		__antithesis_instrumentation__.Notify(245563)
	}
	__antithesis_instrumentation__.Notify(245551)

	hasAdmin, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(245564)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245565)
	}
	__antithesis_instrumentation__.Notify(245552)
	if hasAdmin {
		__antithesis_instrumentation__.Notify(245566)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245567)
	}
	__antithesis_instrumentation__.Notify(245553)

	var objType string
	switch desc.(type) {
	case *typedesc.Mutable:
		__antithesis_instrumentation__.Notify(245568)
		objType = "type"
	case *tabledesc.Mutable:
		__antithesis_instrumentation__.Notify(245569)
		objType = "table"
	case *schemadesc.Mutable:
		__antithesis_instrumentation__.Notify(245570)
		objType = "schema"
	case *dbdesc.Mutable:
		__antithesis_instrumentation__.Notify(245571)
		objType = "database"
	default:
		__antithesis_instrumentation__.Notify(245572)
		return errors.AssertionFailedf("unknown object descriptor type %v", desc)
	}
	__antithesis_instrumentation__.Notify(245554)

	hasOwnership, err := p.HasOwnership(ctx, desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(245573)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245574)
	}
	__antithesis_instrumentation__.Notify(245555)
	if !hasOwnership {
		__antithesis_instrumentation__.Notify(245575)
		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"must be owner of %s %s", tree.Name(objType), tree.Name(desc.GetName()))
	} else {
		__antithesis_instrumentation__.Notify(245576)
	}
	__antithesis_instrumentation__.Notify(245556)

	if p.User() == newOwner {
		__antithesis_instrumentation__.Notify(245577)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245578)
	}
	__antithesis_instrumentation__.Notify(245557)
	memberOf, err := p.MemberOfWithAdminOption(ctx, p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(245579)
		return err
	} else {
		__antithesis_instrumentation__.Notify(245580)
	}
	__antithesis_instrumentation__.Notify(245558)
	if _, ok := memberOf[newOwner]; ok {
		__antithesis_instrumentation__.Notify(245581)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(245582)
	}
	__antithesis_instrumentation__.Notify(245559)
	return pgerror.Newf(pgcode.InsufficientPrivilege, "must be member of role %q", newOwner)
}

func (p *planner) HasOwnershipOnSchema(
	ctx context.Context, schemaID descpb.ID, dbID descpb.ID,
) (bool, error) {
	__antithesis_instrumentation__.Notify(245583)
	if dbID == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(245587)

		return p.User().IsNodeUser(), nil
	} else {
		__antithesis_instrumentation__.Notify(245588)
	}
	__antithesis_instrumentation__.Notify(245584)
	scDesc, err := p.Descriptors().GetImmutableSchemaByID(
		ctx, p.Txn(), schemaID, tree.SchemaLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(245589)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(245590)
	}
	__antithesis_instrumentation__.Notify(245585)

	hasOwnership := false
	switch kind := scDesc.SchemaKind(); kind {
	case catalog.SchemaPublic:
		__antithesis_instrumentation__.Notify(245591)

		hasOwnership, err = p.UserHasAdminRole(ctx, p.User())
		if err != nil {
			__antithesis_instrumentation__.Notify(245596)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(245597)
		}
	case catalog.SchemaVirtual:
		__antithesis_instrumentation__.Notify(245592)

	case catalog.SchemaTemporary:
		__antithesis_instrumentation__.Notify(245593)

		hasOwnership = p.SessionData() != nil && func() bool {
			__antithesis_instrumentation__.Notify(245598)
			return p.SessionData().IsTemporarySchemaID(uint32(scDesc.GetID())) == true
		}() == true
	case catalog.SchemaUserDefined:
		__antithesis_instrumentation__.Notify(245594)
		hasOwnership, err = p.HasOwnership(ctx, scDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(245599)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(245600)
		}
	default:
		__antithesis_instrumentation__.Notify(245595)
		panic(errors.AssertionFailedf("unknown schema kind %d", kind))
	}
	__antithesis_instrumentation__.Notify(245586)

	return hasOwnership, nil
}

func (p *planner) HasViewActivityOrViewActivityRedactedRole(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(245601)
	hasAdmin, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(245604)
		return hasAdmin, err
	} else {
		__antithesis_instrumentation__.Notify(245605)
	}
	__antithesis_instrumentation__.Notify(245602)
	if !hasAdmin {
		__antithesis_instrumentation__.Notify(245606)
		hasViewActivity, err := p.HasRoleOption(ctx, roleoption.VIEWACTIVITY)
		if err != nil {
			__antithesis_instrumentation__.Notify(245608)
			return hasViewActivity, err
		} else {
			__antithesis_instrumentation__.Notify(245609)
		}
		__antithesis_instrumentation__.Notify(245607)
		if !hasViewActivity {
			__antithesis_instrumentation__.Notify(245610)
			hasViewActivityRedacted, err := p.HasRoleOption(ctx, roleoption.VIEWACTIVITYREDACTED)
			if err != nil {
				__antithesis_instrumentation__.Notify(245612)
				return hasViewActivityRedacted, err
			} else {
				__antithesis_instrumentation__.Notify(245613)
			}
			__antithesis_instrumentation__.Notify(245611)
			if !hasViewActivityRedacted {
				__antithesis_instrumentation__.Notify(245614)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(245615)
			}
		} else {
			__antithesis_instrumentation__.Notify(245616)
		}
	} else {
		__antithesis_instrumentation__.Notify(245617)
	}
	__antithesis_instrumentation__.Notify(245603)

	return true, nil
}
