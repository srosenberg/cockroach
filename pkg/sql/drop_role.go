package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessioninit"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type DropRoleNode struct {
	ifExists  bool
	isRole    bool
	roleNames []security.SQLUsername
}

func (p *planner) DropRole(ctx context.Context, n *tree.DropRole) (planNode, error) {
	__antithesis_instrumentation__.Notify(469126)
	return p.DropRoleNode(ctx, n.Names, n.IfExists, n.IsRole, "DROP ROLE")
}

func (p *planner) DropRoleNode(
	ctx context.Context, roleSpecs tree.RoleSpecList, ifExists bool, isRole bool, opName string,
) (*DropRoleNode, error) {
	__antithesis_instrumentation__.Notify(469127)
	if err := p.CheckRoleOption(ctx, roleoption.CREATEROLE); err != nil {
		__antithesis_instrumentation__.Notify(469131)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469132)
	}
	__antithesis_instrumentation__.Notify(469128)

	for _, r := range roleSpecs {
		__antithesis_instrumentation__.Notify(469133)
		if r.RoleSpecType != tree.RoleName {
			__antithesis_instrumentation__.Notify(469134)
			return nil, pgerror.Newf(pgcode.InvalidParameterValue, "cannot use special role specifier in DROP ROLE")
		} else {
			__antithesis_instrumentation__.Notify(469135)
		}
	}
	__antithesis_instrumentation__.Notify(469129)
	roleNames, err := roleSpecs.ToSQLUsernames(p.SessionData(), security.UsernameCreation)
	if err != nil {
		__antithesis_instrumentation__.Notify(469136)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469137)
	}
	__antithesis_instrumentation__.Notify(469130)

	return &DropRoleNode{
		ifExists:  ifExists,
		isRole:    isRole,
		roleNames: roleNames,
	}, nil
}

type objectType string

const (
	database         objectType = "database"
	table            objectType = "table"
	schema           objectType = "schema"
	typeObject       objectType = "type"
	defaultPrivilege objectType = "default_privilege"
)

type objectAndType struct {
	ObjectType   objectType
	ObjectName   string
	ErrorMessage error
}

func (n *DropRoleNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(469138)
	var opName string
	if n.isRole {
		__antithesis_instrumentation__.Notify(469154)
		sqltelemetry.IncIAMDropCounter(sqltelemetry.Role)
		opName = "drop-role"
	} else {
		__antithesis_instrumentation__.Notify(469155)
		sqltelemetry.IncIAMDropCounter(sqltelemetry.User)
		opName = "drop-user"
	}
	__antithesis_instrumentation__.Notify(469139)

	hasAdmin, err := params.p.HasAdminRole(params.ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(469156)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469157)
	}
	__antithesis_instrumentation__.Notify(469140)

	userNames := make(map[security.SQLUsername][]objectAndType)
	for i, name := range n.roleNames {
		__antithesis_instrumentation__.Notify(469158)

		userNames[n.roleNames[i]] = make([]objectAndType, 0)
		if name.IsReserved() {
			__antithesis_instrumentation__.Notify(469160)
			return pgerror.Newf(pgcode.ReservedName, "role name %q is reserved", name.Normalized())
		} else {
			__antithesis_instrumentation__.Notify(469161)
		}
		__antithesis_instrumentation__.Notify(469159)

		if !hasAdmin {
			__antithesis_instrumentation__.Notify(469162)
			targetIsAdmin, err := params.p.UserHasAdminRole(params.ctx, name)
			if err != nil {
				__antithesis_instrumentation__.Notify(469164)
				return err
			} else {
				__antithesis_instrumentation__.Notify(469165)
			}
			__antithesis_instrumentation__.Notify(469163)
			if targetIsAdmin {
				__antithesis_instrumentation__.Notify(469166)
				return pgerror.New(pgcode.InsufficientPrivilege, "must be superuser to drop superusers")
			} else {
				__antithesis_instrumentation__.Notify(469167)
			}
		} else {
			__antithesis_instrumentation__.Notify(469168)
		}
	}
	__antithesis_instrumentation__.Notify(469141)

	privilegeObjectFormatter := tree.NewFmtCtx(tree.FmtSimple)
	defer privilegeObjectFormatter.Close()

	if err := forEachDatabaseDesc(params.ctx, params.p, nil, true,
		func(db catalog.DatabaseDescriptor) error {
			__antithesis_instrumentation__.Notify(469169)
			if _, ok := userNames[db.GetPrivileges().Owner()]; ok {
				__antithesis_instrumentation__.Notify(469172)
				userNames[db.GetPrivileges().Owner()] = append(
					userNames[db.GetPrivileges().Owner()],
					objectAndType{
						ObjectType: database,
						ObjectName: db.GetName(),
					})
			} else {
				__antithesis_instrumentation__.Notify(469173)
			}
			__antithesis_instrumentation__.Notify(469170)
			for _, u := range db.GetPrivileges().Users {
				__antithesis_instrumentation__.Notify(469174)
				if _, ok := userNames[u.User()]; ok {
					__antithesis_instrumentation__.Notify(469175)
					if privilegeObjectFormatter.Len() > 0 {
						__antithesis_instrumentation__.Notify(469177)
						privilegeObjectFormatter.WriteString(", ")
					} else {
						__antithesis_instrumentation__.Notify(469178)
					}
					__antithesis_instrumentation__.Notify(469176)
					privilegeObjectFormatter.FormatName(db.GetName())
					break
				} else {
					__antithesis_instrumentation__.Notify(469179)
				}
			}
			__antithesis_instrumentation__.Notify(469171)
			return accumulateDependentDefaultPrivileges(db.GetDefaultPrivilegeDescriptor(), userNames, db.GetName(), "")
		}); err != nil {
		__antithesis_instrumentation__.Notify(469180)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469181)
	}
	__antithesis_instrumentation__.Notify(469142)

	all, err := params.p.Descriptors().GetAllDescriptors(params.ctx, params.p.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(469182)
		return err
	} else {
		__antithesis_instrumentation__.Notify(469183)
	}
	__antithesis_instrumentation__.Notify(469143)

	lCtx := newInternalLookupCtx(all.OrderedDescriptors(), nil)

	for _, tbID := range lCtx.tbIDs {
		__antithesis_instrumentation__.Notify(469184)
		tableDescriptor := lCtx.tbDescs[tbID]
		if !descriptorIsVisible(tableDescriptor, true) {
			__antithesis_instrumentation__.Notify(469187)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469188)
		}
		__antithesis_instrumentation__.Notify(469185)
		if _, ok := userNames[tableDescriptor.GetPrivileges().Owner()]; ok {
			__antithesis_instrumentation__.Notify(469189)
			tn, err := getTableNameFromTableDescriptor(lCtx, tableDescriptor, "")
			if err != nil {
				__antithesis_instrumentation__.Notify(469191)
				return err
			} else {
				__antithesis_instrumentation__.Notify(469192)
			}
			__antithesis_instrumentation__.Notify(469190)
			userNames[tableDescriptor.GetPrivileges().Owner()] = append(
				userNames[tableDescriptor.GetPrivileges().Owner()],
				objectAndType{
					ObjectType: table,
					ObjectName: tn.String(),
				})
		} else {
			__antithesis_instrumentation__.Notify(469193)
		}
		__antithesis_instrumentation__.Notify(469186)
		for _, u := range tableDescriptor.GetPrivileges().Users {
			__antithesis_instrumentation__.Notify(469194)
			if _, ok := userNames[u.User()]; ok {
				__antithesis_instrumentation__.Notify(469195)
				if privilegeObjectFormatter.Len() > 0 {
					__antithesis_instrumentation__.Notify(469197)
					privilegeObjectFormatter.WriteString(", ")
				} else {
					__antithesis_instrumentation__.Notify(469198)
				}
				__antithesis_instrumentation__.Notify(469196)
				parentName := lCtx.getDatabaseName(tableDescriptor)
				schemaName := lCtx.getSchemaName(tableDescriptor)
				tn := tree.MakeTableNameWithSchema(tree.Name(parentName), tree.Name(schemaName), tree.Name(tableDescriptor.GetName()))
				privilegeObjectFormatter.FormatNode(&tn)
				break
			} else {
				__antithesis_instrumentation__.Notify(469199)
			}
		}
	}
	__antithesis_instrumentation__.Notify(469144)
	for _, schemaDesc := range lCtx.schemaDescs {
		__antithesis_instrumentation__.Notify(469200)
		if !descriptorIsVisible(schemaDesc, true) {
			__antithesis_instrumentation__.Notify(469204)
			continue
		} else {
			__antithesis_instrumentation__.Notify(469205)
		}
		__antithesis_instrumentation__.Notify(469201)

		if _, ok := userNames[schemaDesc.GetPrivileges().Owner()]; ok {
			__antithesis_instrumentation__.Notify(469206)
			userNames[schemaDesc.GetPrivileges().Owner()] = append(
				userNames[schemaDesc.GetPrivileges().Owner()],
				objectAndType{
					ObjectType: schema,
					ObjectName: schemaDesc.GetName(),
				})
		} else {
			__antithesis_instrumentation__.Notify(469207)
		}
		__antithesis_instrumentation__.Notify(469202)

		dbDesc, err := lCtx.getDatabaseByID(schemaDesc.GetParentID())
		if err != nil {
			__antithesis_instrumentation__.Notify(469208)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469209)
		}
		__antithesis_instrumentation__.Notify(469203)

		if err := accumulateDependentDefaultPrivileges(schemaDesc.GetDefaultPrivilegeDescriptor(), userNames, dbDesc.GetName(), schemaDesc.GetName()); err != nil {
			__antithesis_instrumentation__.Notify(469210)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469211)
		}
	}
	__antithesis_instrumentation__.Notify(469145)
	for _, typDesc := range lCtx.typDescs {
		__antithesis_instrumentation__.Notify(469212)
		if _, ok := userNames[typDesc.GetPrivileges().Owner()]; ok {
			__antithesis_instrumentation__.Notify(469213)
			if !descriptorIsVisible(typDesc, true) {
				__antithesis_instrumentation__.Notify(469216)
				continue
			} else {
				__antithesis_instrumentation__.Notify(469217)
			}
			__antithesis_instrumentation__.Notify(469214)
			tn, err := getTypeNameFromTypeDescriptor(lCtx, typDesc)
			if err != nil {
				__antithesis_instrumentation__.Notify(469218)
				return err
			} else {
				__antithesis_instrumentation__.Notify(469219)
			}
			__antithesis_instrumentation__.Notify(469215)
			userNames[typDesc.GetPrivileges().Owner()] = append(
				userNames[typDesc.GetPrivileges().Owner()],
				objectAndType{
					ObjectType: typeObject,
					ObjectName: tn.String(),
				})
		} else {
			__antithesis_instrumentation__.Notify(469220)
		}
	}
	__antithesis_instrumentation__.Notify(469146)

	if privilegeObjectFormatter.Len() > 0 {
		__antithesis_instrumentation__.Notify(469221)
		fnl := tree.NewFmtCtx(tree.FmtSimple)
		defer fnl.Close()
		for i, name := range n.roleNames {
			__antithesis_instrumentation__.Notify(469223)
			if i > 0 {
				__antithesis_instrumentation__.Notify(469225)
				fnl.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(469226)
			}
			__antithesis_instrumentation__.Notify(469224)
			fnl.FormatName(name.Normalized())
		}
		__antithesis_instrumentation__.Notify(469222)
		return pgerror.Newf(pgcode.DependentObjectsStillExist,
			"cannot drop role%s/user%s %s: grants still exist on %s",
			util.Pluralize(int64(len(n.roleNames))), util.Pluralize(int64(len(n.roleNames))),
			fnl.String(), privilegeObjectFormatter.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(469227)
	}
	__antithesis_instrumentation__.Notify(469147)

	hasDependentDefaultPrivilege := false
	for _, name := range n.roleNames {
		__antithesis_instrumentation__.Notify(469228)

		dependentObjects := userNames[name]

		sort.SliceStable(dependentObjects, func(i int, j int) bool {
			__antithesis_instrumentation__.Notify(469230)
			if dependentObjects[i].ObjectType != dependentObjects[j].ObjectType {
				__antithesis_instrumentation__.Notify(469233)
				return dependentObjects[i].ObjectType < dependentObjects[j].ObjectType
			} else {
				__antithesis_instrumentation__.Notify(469234)
			}
			__antithesis_instrumentation__.Notify(469231)

			if dependentObjects[i].ObjectName != dependentObjects[j].ObjectName {
				__antithesis_instrumentation__.Notify(469235)
				return dependentObjects[i].ObjectName < dependentObjects[j].ObjectName
			} else {
				__antithesis_instrumentation__.Notify(469236)
			}
			__antithesis_instrumentation__.Notify(469232)

			return dependentObjects[i].ErrorMessage.Error() < dependentObjects[j].ErrorMessage.Error()
		})
		__antithesis_instrumentation__.Notify(469229)
		var hints []string
		if len(dependentObjects) > 0 {
			__antithesis_instrumentation__.Notify(469237)
			objectsMsg := tree.NewFmtCtx(tree.FmtSimple)
			for _, obj := range dependentObjects {
				__antithesis_instrumentation__.Notify(469240)
				switch obj.ObjectType {
				case database, table, schema, typeObject:
					__antithesis_instrumentation__.Notify(469241)
					objectsMsg.WriteString(fmt.Sprintf("\nowner of %s %s", obj.ObjectType, obj.ObjectName))
				case defaultPrivilege:
					__antithesis_instrumentation__.Notify(469242)
					hasDependentDefaultPrivilege = true
					objectsMsg.WriteString(fmt.Sprintf("\n%s", obj.ErrorMessage))
					hints = append(hints, errors.GetAllHints(obj.ErrorMessage)...)
				default:
					__antithesis_instrumentation__.Notify(469243)
				}
			}
			__antithesis_instrumentation__.Notify(469238)
			objects := objectsMsg.CloseAndGetString()
			err := pgerror.Newf(pgcode.DependentObjectsStillExist,
				"role %s cannot be dropped because some objects depend on it%s",
				name, objects)
			if hasDependentDefaultPrivilege {
				__antithesis_instrumentation__.Notify(469244)
				err = errors.WithHint(err,
					strings.Join(hints, "\n"),
				)
			} else {
				__antithesis_instrumentation__.Notify(469245)
			}
			__antithesis_instrumentation__.Notify(469239)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469246)
		}
	}
	__antithesis_instrumentation__.Notify(469148)

	var numRoleMembershipsDeleted, numRoleSettingsRowsDeleted int
	for normalizedUsername := range userNames {
		__antithesis_instrumentation__.Notify(469247)

		if normalizedUsername.IsAdminRole() || func() bool {
			__antithesis_instrumentation__.Notify(469258)
			return normalizedUsername.IsPublicRole() == true
		}() == true {
			__antithesis_instrumentation__.Notify(469259)
			return pgerror.Newf(
				pgcode.InvalidParameterValue, "cannot drop special role %s", normalizedUsername)
		} else {
			__antithesis_instrumentation__.Notify(469260)
		}
		__antithesis_instrumentation__.Notify(469248)
		if normalizedUsername.IsRootUser() {
			__antithesis_instrumentation__.Notify(469261)
			return pgerror.Newf(
				pgcode.InvalidParameterValue, "cannot drop special user %s", normalizedUsername)
		} else {
			__antithesis_instrumentation__.Notify(469262)
		}
		__antithesis_instrumentation__.Notify(469249)

		numSchedulesRow, err := params.ExecCfg().InternalExecutor.QueryRow(
			params.ctx,
			"check-user-schedules",
			params.p.txn,
			"SELECT count(*) FROM system.scheduled_jobs WHERE owner=$1",
			normalizedUsername,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(469263)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469264)
		}
		__antithesis_instrumentation__.Notify(469250)
		if numSchedulesRow == nil {
			__antithesis_instrumentation__.Notify(469265)
			return errors.New("failed to check user schedules")
		} else {
			__antithesis_instrumentation__.Notify(469266)
		}
		__antithesis_instrumentation__.Notify(469251)
		numSchedules := int64(tree.MustBeDInt(numSchedulesRow[0]))
		if numSchedules > 0 {
			__antithesis_instrumentation__.Notify(469267)
			return pgerror.Newf(pgcode.DependentObjectsStillExist,
				"cannot drop role/user %s; it owns %d scheduled jobs.",
				normalizedUsername, numSchedules)
		} else {
			__antithesis_instrumentation__.Notify(469268)
		}
		__antithesis_instrumentation__.Notify(469252)

		numUsersDeleted, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.Exec(
			params.ctx,
			opName,
			params.p.txn,
			`DELETE FROM system.users WHERE username=$1`,
			normalizedUsername,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(469269)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469270)
		}
		__antithesis_instrumentation__.Notify(469253)

		if numUsersDeleted == 0 && func() bool {
			__antithesis_instrumentation__.Notify(469271)
			return !n.ifExists == true
		}() == true {
			__antithesis_instrumentation__.Notify(469272)
			return errors.Errorf("role/user %s does not exist", normalizedUsername)
		} else {
			__antithesis_instrumentation__.Notify(469273)
		}
		__antithesis_instrumentation__.Notify(469254)

		rowsDeleted, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.Exec(
			params.ctx,
			"drop-role-membership",
			params.p.txn,
			`DELETE FROM system.role_members WHERE "role" = $1 OR "member" = $1`,
			normalizedUsername,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(469274)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469275)
		}
		__antithesis_instrumentation__.Notify(469255)
		numRoleMembershipsDeleted += rowsDeleted

		_, err = params.extendedEvalCtx.ExecCfg.InternalExecutor.Exec(
			params.ctx,
			opName,
			params.p.txn,
			fmt.Sprintf(
				`DELETE FROM %s WHERE username=$1`,
				sessioninit.RoleOptionsTableName,
			),
			normalizedUsername,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(469276)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469277)
		}
		__antithesis_instrumentation__.Notify(469256)

		rowsDeleted, err = params.extendedEvalCtx.ExecCfg.InternalExecutor.Exec(
			params.ctx,
			opName,
			params.p.txn,
			fmt.Sprintf(
				`DELETE FROM %s WHERE role_name = $1`,
				sessioninit.DatabaseRoleSettingsTableName,
			),
			normalizedUsername,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(469278)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469279)
		}
		__antithesis_instrumentation__.Notify(469257)
		numRoleSettingsRowsDeleted += rowsDeleted

	}
	__antithesis_instrumentation__.Notify(469149)

	if sessioninit.CacheEnabled.Get(&params.p.ExecCfg().Settings.SV) {
		__antithesis_instrumentation__.Notify(469280)
		if err := params.p.bumpUsersTableVersion(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(469283)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469284)
		}
		__antithesis_instrumentation__.Notify(469281)
		if err := params.p.bumpRoleOptionsTableVersion(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(469285)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469286)
		}
		__antithesis_instrumentation__.Notify(469282)
		if numRoleSettingsRowsDeleted > 0 {
			__antithesis_instrumentation__.Notify(469287)
			if err := params.p.bumpDatabaseRoleSettingsTableVersion(params.ctx); err != nil {
				__antithesis_instrumentation__.Notify(469288)
				return err
			} else {
				__antithesis_instrumentation__.Notify(469289)
			}
		} else {
			__antithesis_instrumentation__.Notify(469290)
		}
	} else {
		__antithesis_instrumentation__.Notify(469291)
	}
	__antithesis_instrumentation__.Notify(469150)
	if numRoleMembershipsDeleted > 0 {
		__antithesis_instrumentation__.Notify(469292)
		if err := params.p.BumpRoleMembershipTableVersion(params.ctx); err != nil {
			__antithesis_instrumentation__.Notify(469293)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469294)
		}
	} else {
		__antithesis_instrumentation__.Notify(469295)
	}
	__antithesis_instrumentation__.Notify(469151)

	normalizedNames := make([]string, len(n.roleNames))
	for i, name := range n.roleNames {
		__antithesis_instrumentation__.Notify(469296)
		normalizedNames[i] = name.Normalized()
	}
	__antithesis_instrumentation__.Notify(469152)
	sort.Strings(normalizedNames)
	for _, name := range normalizedNames {
		__antithesis_instrumentation__.Notify(469297)
		if err := params.p.logEvent(params.ctx,
			0,
			&eventpb.DropRole{RoleName: name}); err != nil {
			__antithesis_instrumentation__.Notify(469298)
			return err
		} else {
			__antithesis_instrumentation__.Notify(469299)
		}
	}
	__antithesis_instrumentation__.Notify(469153)
	return nil
}

func (*DropRoleNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(469300)
	return false, nil
}

func (*DropRoleNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(469301)
	return tree.Datums{}
}

func (*DropRoleNode) Close(context.Context) { __antithesis_instrumentation__.Notify(469302) }

func accumulateDependentDefaultPrivileges(
	defaultPrivilegeDescriptor catalog.DefaultPrivilegeDescriptor,
	userNames map[security.SQLUsername][]objectAndType,
	dbName string,
	schemaName string,
) error {
	__antithesis_instrumentation__.Notify(469303)

	return defaultPrivilegeDescriptor.ForEachDefaultPrivilegeForRole(func(
		defaultPrivilegesForRole catpb.DefaultPrivilegesForRole) error {
		__antithesis_instrumentation__.Notify(469304)
		role := catpb.DefaultPrivilegesRole{}
		if defaultPrivilegesForRole.IsExplicitRole() {
			__antithesis_instrumentation__.Notify(469307)
			role.Role = defaultPrivilegesForRole.GetExplicitRole().UserProto.Decode()
		} else {
			__antithesis_instrumentation__.Notify(469308)
			role.ForAllRoles = true
		}
		__antithesis_instrumentation__.Notify(469305)
		for object, defaultPrivs := range defaultPrivilegesForRole.DefaultPrivilegesPerObject {
			__antithesis_instrumentation__.Notify(469309)
			addDependentPrivileges(object, defaultPrivs, role, userNames, dbName, schemaName)
		}
		__antithesis_instrumentation__.Notify(469306)
		return nil
	})
}

func addDependentPrivileges(
	object tree.AlterDefaultPrivilegesTargetObject,
	defaultPrivs catpb.PrivilegeDescriptor,
	role catpb.DefaultPrivilegesRole,
	userNames map[security.SQLUsername][]objectAndType,
	dbName string,
	schemaName string,
) {
	__antithesis_instrumentation__.Notify(469310)
	var objectType string
	switch object {
	case tree.Tables:
		__antithesis_instrumentation__.Notify(469314)
		objectType = "relations"
	case tree.Sequences:
		__antithesis_instrumentation__.Notify(469315)
		objectType = "sequences"
	case tree.Types:
		__antithesis_instrumentation__.Notify(469316)
		objectType = "types"
	case tree.Schemas:
		__antithesis_instrumentation__.Notify(469317)
		objectType = "schemas"
	default:
		__antithesis_instrumentation__.Notify(469318)
	}
	__antithesis_instrumentation__.Notify(469311)

	inSchemaMsg := ""
	if schemaName != "" {
		__antithesis_instrumentation__.Notify(469319)
		inSchemaMsg = fmt.Sprintf(" in schema %s", schemaName)
	} else {
		__antithesis_instrumentation__.Notify(469320)
	}
	__antithesis_instrumentation__.Notify(469312)

	createHint := func(
		role catpb.DefaultPrivilegesRole,
		grantee security.SQLUsername,
	) string {
		__antithesis_instrumentation__.Notify(469321)

		roleString := "ALL ROLES"
		if !role.ForAllRoles {
			__antithesis_instrumentation__.Notify(469323)
			roleString = fmt.Sprintf("ROLE %s", role.Role.SQLIdentifier())
		} else {
			__antithesis_instrumentation__.Notify(469324)
		}
		__antithesis_instrumentation__.Notify(469322)

		return fmt.Sprintf("USE %s; ALTER DEFAULT PRIVILEGES FOR %s%s REVOKE ALL ON %s FROM %s;",
			dbName, roleString, strings.ToUpper(inSchemaMsg), strings.ToUpper(object.String()), grantee.SQLIdentifier())
	}
	__antithesis_instrumentation__.Notify(469313)

	for _, privs := range defaultPrivs.Users {
		__antithesis_instrumentation__.Notify(469325)
		grantee := privs.User()
		if !role.ForAllRoles {
			__antithesis_instrumentation__.Notify(469327)
			if _, ok := userNames[role.Role]; ok {
				__antithesis_instrumentation__.Notify(469328)
				hint := createHint(role, grantee)
				userNames[role.Role] = append(userNames[role.Role],
					objectAndType{
						ObjectType: defaultPrivilege,
						ErrorMessage: errors.WithHint(
							errors.Newf(
								"owner of default privileges on new %s belonging to role %s in database %s%s",
								objectType, role.Role, dbName, inSchemaMsg,
							), hint),
					})
			} else {
				__antithesis_instrumentation__.Notify(469329)
			}
		} else {
			__antithesis_instrumentation__.Notify(469330)
		}
		__antithesis_instrumentation__.Notify(469326)
		if _, ok := userNames[grantee]; ok {
			__antithesis_instrumentation__.Notify(469331)
			hint := createHint(role, grantee)
			var err error
			if role.ForAllRoles {
				__antithesis_instrumentation__.Notify(469333)
				err = errors.Newf(
					"privileges for default privileges on new %s for all roles in database %s%s",
					objectType, dbName, inSchemaMsg,
				)
			} else {
				__antithesis_instrumentation__.Notify(469334)
				err = errors.Newf(
					"privileges for default privileges on new %s belonging to role %s in database %s%s",
					objectType, role.Role, dbName, inSchemaMsg,
				)
			}
			__antithesis_instrumentation__.Notify(469332)
			userNames[grantee] = append(userNames[grantee],
				objectAndType{
					ObjectType:   defaultPrivilege,
					ErrorMessage: errors.WithHint(err, hint),
				})
		} else {
			__antithesis_instrumentation__.Notify(469335)
		}
	}
}
