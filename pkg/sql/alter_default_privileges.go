package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

var targetObjectToPrivilegeObject = map[tree.AlterDefaultPrivilegesTargetObject]privilege.ObjectType{
	tree.Tables:    privilege.Table,
	tree.Sequences: privilege.Table,
	tree.Types:     privilege.Type,
	tree.Schemas:   privilege.Schema,
}

type alterDefaultPrivilegesNode struct {
	n *tree.AlterDefaultPrivileges

	dbDesc      *dbdesc.Mutable
	schemaDescs []*schemadesc.Mutable
}

func (n *alterDefaultPrivilegesNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(243023)
	return false, nil
}
func (n *alterDefaultPrivilegesNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(243024)
	return tree.Datums{}
}
func (n *alterDefaultPrivilegesNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(243025)
}

func (p *planner) alterDefaultPrivileges(
	ctx context.Context, n *tree.AlterDefaultPrivileges,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(243026)

	database := p.CurrentDatabase()
	if n.Database != nil {
		__antithesis_instrumentation__.Notify(243032)
		database = n.Database.Normalize()
	} else {
		__antithesis_instrumentation__.Notify(243033)
	}
	__antithesis_instrumentation__.Notify(243027)
	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, database,
		tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(243034)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243035)
	}
	__antithesis_instrumentation__.Notify(243028)

	objectType := n.Grant.Target
	if !n.IsGrant {
		__antithesis_instrumentation__.Notify(243036)
		objectType = n.Revoke.Target
	} else {
		__antithesis_instrumentation__.Notify(243037)
	}
	__antithesis_instrumentation__.Notify(243029)

	if len(n.Schemas) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(243038)
		return objectType == tree.Schemas == true
	}() == true {
		__antithesis_instrumentation__.Notify(243039)
		return nil, pgerror.WithCandidateCode(errors.New(
			"cannot use IN SCHEMA clause when using GRANT/REVOKE ON SCHEMAS"),
			pgcode.InvalidGrantOperation,
		)
	} else {
		__antithesis_instrumentation__.Notify(243040)
	}
	__antithesis_instrumentation__.Notify(243030)

	var schemaDescs []*schemadesc.Mutable
	for _, sc := range n.Schemas {
		__antithesis_instrumentation__.Notify(243041)
		schemaDesc, err := p.Descriptors().GetMutableSchemaByName(ctx, p.txn, dbDesc, sc.Schema(), tree.SchemaLookupFlags{Required: true})
		if err != nil {
			__antithesis_instrumentation__.Notify(243043)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(243044)
		}
		__antithesis_instrumentation__.Notify(243042)
		schemaDescs = append(schemaDescs, schemaDesc.(*schemadesc.Mutable))
	}
	__antithesis_instrumentation__.Notify(243031)

	return &alterDefaultPrivilegesNode{
		n:           n,
		dbDesc:      dbDesc,
		schemaDescs: schemaDescs,
	}, err
}

func (n *alterDefaultPrivilegesNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(243045)
	if (n.n.Grant.WithGrantOption || func() bool {
		__antithesis_instrumentation__.Notify(243056)
		return n.n.Revoke.GrantOptionFor == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(243057)
		return !params.p.ExecCfg().Settings.Version.IsActive(params.ctx, clusterversion.ValidateGrantOption) == true
	}() == true {
		__antithesis_instrumentation__.Notify(243058)
		return pgerror.Newf(pgcode.FeatureNotSupported,
			"version %v must be finalized to use grant options",
			clusterversion.ByKey(clusterversion.ValidateGrantOption))
	} else {
		__antithesis_instrumentation__.Notify(243059)
	}
	__antithesis_instrumentation__.Notify(243046)
	targetRoles, err := n.n.Roles.ToSQLUsernames(params.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(243060)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243061)
	}
	__antithesis_instrumentation__.Notify(243047)

	if len(targetRoles) == 0 {
		__antithesis_instrumentation__.Notify(243062)
		targetRoles = append(targetRoles, params.p.User())
	} else {
		__antithesis_instrumentation__.Notify(243063)
	}
	__antithesis_instrumentation__.Notify(243048)

	if err := params.p.validateRoles(params.ctx, targetRoles, false); err != nil {
		__antithesis_instrumentation__.Notify(243064)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243065)
	}
	__antithesis_instrumentation__.Notify(243049)

	privileges := n.n.Grant.Privileges
	grantees := n.n.Grant.Grantees
	objectType := n.n.Grant.Target
	grantOption := n.n.Grant.WithGrantOption
	if !n.n.IsGrant {
		__antithesis_instrumentation__.Notify(243066)
		privileges = n.n.Revoke.Privileges
		grantees = n.n.Revoke.Grantees
		objectType = n.n.Revoke.Target
		grantOption = n.n.Revoke.GrantOptionFor
	} else {
		__antithesis_instrumentation__.Notify(243067)
	}
	__antithesis_instrumentation__.Notify(243050)

	granteeSQLUsernames, err := grantees.ToSQLUsernames(params.p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(243068)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243069)
	}
	__antithesis_instrumentation__.Notify(243051)

	if err := params.p.validateRoles(params.ctx, granteeSQLUsernames, true); err != nil {
		__antithesis_instrumentation__.Notify(243070)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243071)
	}
	__antithesis_instrumentation__.Notify(243052)

	if n.n.ForAllRoles {
		__antithesis_instrumentation__.Notify(243072)
		if err := params.p.RequireAdminRole(params.ctx, "ALTER DEFAULT PRIVILEGES"); err != nil {
			__antithesis_instrumentation__.Notify(243073)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243074)
		}
	} else {
		__antithesis_instrumentation__.Notify(243075)

		for _, targetRole := range targetRoles {
			__antithesis_instrumentation__.Notify(243076)
			if targetRole != params.p.User() {
				__antithesis_instrumentation__.Notify(243077)
				memberOf, err := params.p.MemberOfWithAdminOption(params.ctx, params.p.User())
				if err != nil {
					__antithesis_instrumentation__.Notify(243079)
					return err
				} else {
					__antithesis_instrumentation__.Notify(243080)
				}
				__antithesis_instrumentation__.Notify(243078)

				if _, found := memberOf[targetRole]; !found {
					__antithesis_instrumentation__.Notify(243081)
					return pgerror.Newf(pgcode.InsufficientPrivilege,
						"must be a member of %s", targetRole.Normalized())
				} else {
					__antithesis_instrumentation__.Notify(243082)
				}
			} else {
				__antithesis_instrumentation__.Notify(243083)
			}
		}
	}
	__antithesis_instrumentation__.Notify(243053)

	if err := privilege.ValidatePrivileges(
		privileges,
		targetObjectToPrivilegeObject[objectType],
	); err != nil {
		__antithesis_instrumentation__.Notify(243084)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243085)
	}
	__antithesis_instrumentation__.Notify(243054)

	if len(n.schemaDescs) == 0 {
		__antithesis_instrumentation__.Notify(243086)
		return n.alterDefaultPrivilegesForDatabase(params, targetRoles, objectType, grantees, privileges, grantOption)
	} else {
		__antithesis_instrumentation__.Notify(243087)
	}
	__antithesis_instrumentation__.Notify(243055)
	return n.alterDefaultPrivilegesForSchemas(params, targetRoles, objectType, grantees, privileges, grantOption)
}

func (n *alterDefaultPrivilegesNode) alterDefaultPrivilegesForSchemas(
	params runParams,
	targetRoles []security.SQLUsername,
	objectType tree.AlterDefaultPrivilegesTargetObject,
	grantees tree.RoleSpecList,
	privileges privilege.List,
	grantOption bool,
) error {
	__antithesis_instrumentation__.Notify(243088)
	var events []eventLogEntry
	for _, schemaDesc := range n.schemaDescs {
		__antithesis_instrumentation__.Notify(243090)
		if schemaDesc.GetDefaultPrivileges() == nil {
			__antithesis_instrumentation__.Notify(243096)
			schemaDesc.SetDefaultPrivilegeDescriptor(catprivilege.MakeDefaultPrivilegeDescriptor(catpb.DefaultPrivilegeDescriptor_SCHEMA))
		} else {
			__antithesis_instrumentation__.Notify(243097)
		}
		__antithesis_instrumentation__.Notify(243091)

		defaultPrivs := schemaDesc.GetMutableDefaultPrivilegeDescriptor()

		var roles []catpb.DefaultPrivilegesRole
		if n.n.ForAllRoles {
			__antithesis_instrumentation__.Notify(243098)
			roles = append(roles, catpb.DefaultPrivilegesRole{
				ForAllRoles: true,
			})
		} else {
			__antithesis_instrumentation__.Notify(243099)
			roles = make([]catpb.DefaultPrivilegesRole, len(targetRoles))
			for i, role := range targetRoles {
				__antithesis_instrumentation__.Notify(243100)
				roles[i] = catpb.DefaultPrivilegesRole{
					Role: role,
				}
			}
		}
		__antithesis_instrumentation__.Notify(243092)

		granteeSQLUsernames, err := grantees.ToSQLUsernames(params.SessionData(), security.UsernameValidation)
		if err != nil {
			__antithesis_instrumentation__.Notify(243101)
			return err
		} else {
			__antithesis_instrumentation__.Notify(243102)
		}
		__antithesis_instrumentation__.Notify(243093)

		grantPresent, allPresent := false, false
		for _, priv := range privileges {
			__antithesis_instrumentation__.Notify(243103)
			grantPresent = grantPresent || func() bool {
				__antithesis_instrumentation__.Notify(243104)
				return priv == privilege.GRANT == true
			}() == true
			allPresent = allPresent || func() bool {
				__antithesis_instrumentation__.Notify(243105)
				return priv == privilege.ALL == true
			}() == true
		}
		__antithesis_instrumentation__.Notify(243094)
		if params.ExecCfg().Settings.Version.IsActive(params.ctx, clusterversion.ValidateGrantOption) {
			__antithesis_instrumentation__.Notify(243106)
			noticeMessage := ""

			if allPresent && func() bool {
				__antithesis_instrumentation__.Notify(243108)
				return n.n.IsGrant == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(243109)
				return !grantOption == true
			}() == true {
				__antithesis_instrumentation__.Notify(243110)
				noticeMessage = "grant options were automatically applied but this behavior is deprecated"
			} else {
				__antithesis_instrumentation__.Notify(243111)
				if grantPresent {
					__antithesis_instrumentation__.Notify(243112)
					noticeMessage = "the GRANT privilege is deprecated"
				} else {
					__antithesis_instrumentation__.Notify(243113)
				}
			}
			__antithesis_instrumentation__.Notify(243107)
			if len(noticeMessage) > 0 {
				__antithesis_instrumentation__.Notify(243114)
				params.p.BufferClientNotice(
					params.ctx,
					errors.WithHint(
						pgnotice.Newf("%s", noticeMessage),
						"please use WITH GRANT OPTION",
					),
				)
			} else {
				__antithesis_instrumentation__.Notify(243115)
			}
		} else {
			__antithesis_instrumentation__.Notify(243116)
		}
		__antithesis_instrumentation__.Notify(243095)

		for _, role := range roles {
			__antithesis_instrumentation__.Notify(243117)
			if n.n.IsGrant {
				__antithesis_instrumentation__.Notify(243121)
				defaultPrivs.GrantDefaultPrivileges(
					role, privileges, granteeSQLUsernames, objectType, grantOption, grantPresent || func() bool {
						__antithesis_instrumentation__.Notify(243122)
						return allPresent == true
					}() == true,
				)
			} else {
				__antithesis_instrumentation__.Notify(243123)
				defaultPrivs.RevokeDefaultPrivileges(
					role, privileges, granteeSQLUsernames, objectType, grantOption, grantPresent || func() bool {
						__antithesis_instrumentation__.Notify(243124)
						return allPresent == true
					}() == true,
				)
			}
			__antithesis_instrumentation__.Notify(243118)

			eventDetails := eventpb.CommonSQLPrivilegeEventDetails{}
			if n.n.IsGrant {
				__antithesis_instrumentation__.Notify(243125)
				eventDetails.GrantedPrivileges = privileges.SortedNames()
			} else {
				__antithesis_instrumentation__.Notify(243126)
				eventDetails.RevokedPrivileges = privileges.SortedNames()
			}
			__antithesis_instrumentation__.Notify(243119)
			event := eventpb.AlterDefaultPrivileges{
				CommonSQLPrivilegeEventDetails: eventDetails,
				SchemaName:                     schemaDesc.GetName(),
			}
			if n.n.ForAllRoles {
				__antithesis_instrumentation__.Notify(243127)
				event.ForAllRoles = true
			} else {
				__antithesis_instrumentation__.Notify(243128)
				event.RoleName = role.Role.Normalized()
			}
			__antithesis_instrumentation__.Notify(243120)

			events = append(events, eventLogEntry{
				targetID: int32(n.dbDesc.GetID()),
				event:    &event,
			})

			if err := params.p.writeSchemaDescChange(
				params.ctx, schemaDesc, tree.AsStringWithFQNames(n.n, params.Ann()),
			); err != nil {
				__antithesis_instrumentation__.Notify(243129)
				return err
			} else {
				__antithesis_instrumentation__.Notify(243130)
			}
		}
	}
	__antithesis_instrumentation__.Notify(243089)

	return params.p.logEvents(params.ctx, events...)
}

func (n *alterDefaultPrivilegesNode) alterDefaultPrivilegesForDatabase(
	params runParams,
	targetRoles []security.SQLUsername,
	objectType tree.AlterDefaultPrivilegesTargetObject,
	grantees tree.RoleSpecList,
	privileges privilege.List,
	grantOption bool,
) error {
	__antithesis_instrumentation__.Notify(243131)
	if n.dbDesc.GetDefaultPrivileges() == nil {
		__antithesis_instrumentation__.Notify(243139)
		n.dbDesc.SetDefaultPrivilegeDescriptor(catprivilege.MakeDefaultPrivilegeDescriptor(catpb.DefaultPrivilegeDescriptor_DATABASE))
	} else {
		__antithesis_instrumentation__.Notify(243140)
	}
	__antithesis_instrumentation__.Notify(243132)

	defaultPrivs := n.dbDesc.GetMutableDefaultPrivilegeDescriptor()

	var roles []catpb.DefaultPrivilegesRole
	if n.n.ForAllRoles {
		__antithesis_instrumentation__.Notify(243141)
		roles = append(roles, catpb.DefaultPrivilegesRole{
			ForAllRoles: true,
		})
	} else {
		__antithesis_instrumentation__.Notify(243142)
		roles = make([]catpb.DefaultPrivilegesRole, len(targetRoles))
		for i, role := range targetRoles {
			__antithesis_instrumentation__.Notify(243143)
			roles[i] = catpb.DefaultPrivilegesRole{
				Role: role,
			}
		}
	}
	__antithesis_instrumentation__.Notify(243133)

	var events []eventLogEntry
	granteeSQLUsernames, err := grantees.ToSQLUsernames(params.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(243144)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243145)
	}
	__antithesis_instrumentation__.Notify(243134)

	grantPresent, allPresent := false, false
	for _, priv := range privileges {
		__antithesis_instrumentation__.Notify(243146)
		grantPresent = grantPresent || func() bool {
			__antithesis_instrumentation__.Notify(243147)
			return priv == privilege.GRANT == true
		}() == true
		allPresent = allPresent || func() bool {
			__antithesis_instrumentation__.Notify(243148)
			return priv == privilege.ALL == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(243135)
	if params.ExecCfg().Settings.Version.IsActive(params.ctx, clusterversion.ValidateGrantOption) {
		__antithesis_instrumentation__.Notify(243149)
		noticeMessage := ""

		if allPresent && func() bool {
			__antithesis_instrumentation__.Notify(243151)
			return n.n.IsGrant == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(243152)
			return !grantOption == true
		}() == true {
			__antithesis_instrumentation__.Notify(243153)
			noticeMessage = "grant options were automatically applied but this behavior is deprecated"
		} else {
			__antithesis_instrumentation__.Notify(243154)
			if grantPresent {
				__antithesis_instrumentation__.Notify(243155)
				noticeMessage = "the GRANT privilege is deprecated"
			} else {
				__antithesis_instrumentation__.Notify(243156)
			}
		}
		__antithesis_instrumentation__.Notify(243150)

		if len(noticeMessage) > 0 {
			__antithesis_instrumentation__.Notify(243157)
			params.p.BufferClientNotice(
				params.ctx,
				errors.WithHint(
					pgnotice.Newf("%s", noticeMessage),
					"please use WITH GRANT OPTION",
				),
			)
		} else {
			__antithesis_instrumentation__.Notify(243158)
		}
	} else {
		__antithesis_instrumentation__.Notify(243159)
	}
	__antithesis_instrumentation__.Notify(243136)

	for _, role := range roles {
		__antithesis_instrumentation__.Notify(243160)
		if n.n.IsGrant {
			__antithesis_instrumentation__.Notify(243164)
			defaultPrivs.GrantDefaultPrivileges(
				role, privileges, granteeSQLUsernames, objectType, grantOption, grantPresent || func() bool {
					__antithesis_instrumentation__.Notify(243165)
					return allPresent == true
				}() == true,
			)
		} else {
			__antithesis_instrumentation__.Notify(243166)
			defaultPrivs.RevokeDefaultPrivileges(
				role, privileges, granteeSQLUsernames, objectType, grantOption, grantPresent || func() bool {
					__antithesis_instrumentation__.Notify(243167)
					return allPresent == true
				}() == true,
			)
		}
		__antithesis_instrumentation__.Notify(243161)

		eventDetails := eventpb.CommonSQLPrivilegeEventDetails{}
		if n.n.IsGrant {
			__antithesis_instrumentation__.Notify(243168)
			eventDetails.GrantedPrivileges = privileges.SortedNames()
		} else {
			__antithesis_instrumentation__.Notify(243169)
			eventDetails.RevokedPrivileges = privileges.SortedNames()
		}
		__antithesis_instrumentation__.Notify(243162)
		event := eventpb.AlterDefaultPrivileges{
			CommonSQLPrivilegeEventDetails: eventDetails,
			DatabaseName:                   n.dbDesc.GetName(),
		}
		if n.n.ForAllRoles {
			__antithesis_instrumentation__.Notify(243170)
			event.ForAllRoles = true
		} else {
			__antithesis_instrumentation__.Notify(243171)
			event.RoleName = role.Role.Normalized()
		}
		__antithesis_instrumentation__.Notify(243163)

		events = append(events, eventLogEntry{
			targetID: int32(n.dbDesc.GetID()),
			event:    &event,
		})
	}
	__antithesis_instrumentation__.Notify(243137)

	if err := params.p.writeNonDropDatabaseChange(
		params.ctx, n.dbDesc, tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(243172)
		return err
	} else {
		__antithesis_instrumentation__.Notify(243173)
	}
	__antithesis_instrumentation__.Notify(243138)

	return params.p.logEvents(params.ctx, events...)
}
