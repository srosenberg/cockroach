package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

func (p *planner) Grant(ctx context.Context, n *tree.Grant) (planNode, error) {
	__antithesis_instrumentation__.Notify(492750)
	grantOn := getGrantOnObject(n.Targets, sqltelemetry.IncIAMGrantPrivilegesCounter)

	if err := privilege.ValidatePrivileges(n.Privileges, grantOn); err != nil {
		__antithesis_instrumentation__.Notify(492753)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(492754)
	}
	__antithesis_instrumentation__.Notify(492751)

	grantees, err := n.Grantees.ToSQLUsernames(p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(492755)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(492756)
	}
	__antithesis_instrumentation__.Notify(492752)

	return &changePrivilegesNode{
		isGrant:         true,
		withGrantOption: n.WithGrantOption,
		targets:         n.Targets,
		grantees:        grantees,
		desiredprivs:    n.Privileges,
		changePrivilege: func(privDesc *catpb.PrivilegeDescriptor, privileges privilege.List, grantee security.SQLUsername) {
			__antithesis_instrumentation__.Notify(492757)
			privDesc.Grant(grantee, privileges, n.WithGrantOption)
		},
		grantOn:          grantOn,
		granteesNameList: n.Grantees,
	}, nil
}

func (p *planner) Revoke(ctx context.Context, n *tree.Revoke) (planNode, error) {
	__antithesis_instrumentation__.Notify(492758)
	grantOn := getGrantOnObject(n.Targets, sqltelemetry.IncIAMRevokePrivilegesCounter)

	if err := privilege.ValidatePrivileges(n.Privileges, grantOn); err != nil {
		__antithesis_instrumentation__.Notify(492761)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(492762)
	}
	__antithesis_instrumentation__.Notify(492759)

	grantees, err := n.Grantees.ToSQLUsernames(p.SessionData(), security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(492763)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(492764)
	}
	__antithesis_instrumentation__.Notify(492760)
	return &changePrivilegesNode{
		isGrant:         false,
		withGrantOption: n.GrantOptionFor,
		targets:         n.Targets,
		grantees:        grantees,
		desiredprivs:    n.Privileges,
		changePrivilege: func(privDesc *catpb.PrivilegeDescriptor, privileges privilege.List, grantee security.SQLUsername) {
			__antithesis_instrumentation__.Notify(492765)
			privDesc.Revoke(grantee, privileges, grantOn, n.GrantOptionFor)
		},
		grantOn:          grantOn,
		granteesNameList: n.Grantees,
	}, nil
}

type changePrivilegesNode struct {
	isGrant         bool
	withGrantOption bool
	targets         tree.TargetList
	grantees        []security.SQLUsername
	desiredprivs    privilege.List
	changePrivilege func(*catpb.PrivilegeDescriptor, privilege.List, security.SQLUsername)
	grantOn         privilege.ObjectType

	granteesNameList tree.RoleSpecList
}

func (n *changePrivilegesNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(492766) }

func (n *changePrivilegesNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(492767)
	ctx := params.ctx
	p := params.p

	if n.withGrantOption && func() bool {
		__antithesis_instrumentation__.Notify(492777)
		return !p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.ValidateGrantOption) == true
	}() == true {
		__antithesis_instrumentation__.Notify(492778)
		return pgerror.Newf(pgcode.FeatureNotSupported,
			"version %v must be finalized to use grant options",
			clusterversion.ByKey(clusterversion.ValidateGrantOption))
	} else {
		__antithesis_instrumentation__.Notify(492779)
	}
	__antithesis_instrumentation__.Notify(492768)

	if err := p.validateRoles(ctx, n.grantees, true); err != nil {
		__antithesis_instrumentation__.Notify(492780)
		return err
	} else {
		__antithesis_instrumentation__.Notify(492781)
	}
	__antithesis_instrumentation__.Notify(492769)

	if n.isGrant && func() bool {
		__antithesis_instrumentation__.Notify(492782)
		return n.withGrantOption == true
	}() == true {
		__antithesis_instrumentation__.Notify(492783)
		for _, grantee := range n.grantees {
			__antithesis_instrumentation__.Notify(492784)
			if grantee.IsPublicRole() {
				__antithesis_instrumentation__.Notify(492785)
				return pgerror.Newf(
					pgcode.InvalidGrantOperation,
					"grant options cannot be granted to %q role",
					security.PublicRoleName(),
				)
			} else {
				__antithesis_instrumentation__.Notify(492786)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(492787)
	}
	__antithesis_instrumentation__.Notify(492770)

	var err error
	var descriptors []catalog.Descriptor

	p.runWithOptions(resolveFlags{skipCache: true}, func() {
		__antithesis_instrumentation__.Notify(492788)
		descriptors, err = getDescriptorsFromTargetListForPrivilegeChange(ctx, p, n.targets)
	})
	__antithesis_instrumentation__.Notify(492771)
	if err != nil {
		__antithesis_instrumentation__.Notify(492789)
		return err
	} else {
		__antithesis_instrumentation__.Notify(492790)
	}
	__antithesis_instrumentation__.Notify(492772)

	if len(descriptors) == 0 {
		__antithesis_instrumentation__.Notify(492791)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(492792)
	}
	__antithesis_instrumentation__.Notify(492773)

	var events []eventLogEntry

	b := p.txn.NewBatch()
	for _, descriptor := range descriptors {
		__antithesis_instrumentation__.Notify(492793)

		op := "REVOKE"
		if n.isGrant {
			__antithesis_instrumentation__.Notify(492799)
			op = "GRANT"
		} else {
			__antithesis_instrumentation__.Notify(492800)
		}
		__antithesis_instrumentation__.Notify(492794)
		if catalog.IsSystemDescriptor(descriptor) {
			__antithesis_instrumentation__.Notify(492801)
			return pgerror.Newf(pgcode.InsufficientPrivilege, "cannot %s on system object", op)
		} else {
			__antithesis_instrumentation__.Notify(492802)
		}
		__antithesis_instrumentation__.Notify(492795)

		if !p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.ValidateGrantOption) {
			__antithesis_instrumentation__.Notify(492803)
			if err := p.CheckPrivilege(ctx, descriptor, privilege.GRANT); err != nil {
				__antithesis_instrumentation__.Notify(492804)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492805)
			}
		} else {
			__antithesis_instrumentation__.Notify(492806)
		}
		__antithesis_instrumentation__.Notify(492796)

		if len(n.desiredprivs) > 0 {
			__antithesis_instrumentation__.Notify(492807)
			grantPresent, allPresent := false, false
			for _, priv := range n.desiredprivs {
				__antithesis_instrumentation__.Notify(492813)

				if err := p.CheckPrivilege(ctx, descriptor, priv); err != nil {
					__antithesis_instrumentation__.Notify(492815)
					return err
				} else {
					__antithesis_instrumentation__.Notify(492816)
				}
				__antithesis_instrumentation__.Notify(492814)
				grantPresent = grantPresent || func() bool {
					__antithesis_instrumentation__.Notify(492817)
					return priv == privilege.GRANT == true
				}() == true
				allPresent = allPresent || func() bool {
					__antithesis_instrumentation__.Notify(492818)
					return priv == privilege.ALL == true
				}() == true
			}
			__antithesis_instrumentation__.Notify(492808)
			privileges := descriptor.GetPrivileges()

			noticeMessage := ""
			if p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.ValidateGrantOption) {
				__antithesis_instrumentation__.Notify(492819)
				err := p.CheckGrantOptionsForUser(ctx, descriptor, n.desiredprivs, p.User(), n.isGrant)
				if err != nil {
					__antithesis_instrumentation__.Notify(492821)
					return err
				} else {
					__antithesis_instrumentation__.Notify(492822)
				}
				__antithesis_instrumentation__.Notify(492820)

				if allPresent && func() bool {
					__antithesis_instrumentation__.Notify(492823)
					return n.isGrant == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(492824)
					return !n.withGrantOption == true
				}() == true {
					__antithesis_instrumentation__.Notify(492825)
					noticeMessage = "grant options were automatically applied but this behavior is deprecated"
				} else {
					__antithesis_instrumentation__.Notify(492826)
					if grantPresent {
						__antithesis_instrumentation__.Notify(492827)
						noticeMessage = "the GRANT privilege is deprecated"
					} else {
						__antithesis_instrumentation__.Notify(492828)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(492829)
			}
			__antithesis_instrumentation__.Notify(492809)

			for _, grantee := range n.grantees {
				__antithesis_instrumentation__.Notify(492830)
				n.changePrivilege(privileges, n.desiredprivs, grantee)

				granteeHasGrantPriv := privileges.CheckPrivilege(grantee, privilege.GRANT)

				if granteeHasGrantPriv && func() bool {
					__antithesis_instrumentation__.Notify(492832)
					return n.isGrant == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(492833)
					return !n.withGrantOption == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(492834)
					return len(noticeMessage) == 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(492835)
					noticeMessage = "grant options were automatically applied but this behavior is deprecated"
				} else {
					__antithesis_instrumentation__.Notify(492836)
				}
				__antithesis_instrumentation__.Notify(492831)
				if !n.withGrantOption && func() bool {
					__antithesis_instrumentation__.Notify(492837)
					return (grantPresent || func() bool {
						__antithesis_instrumentation__.Notify(492838)
						return allPresent == true
					}() == true || func() bool {
						__antithesis_instrumentation__.Notify(492839)
						return (granteeHasGrantPriv && func() bool {
							__antithesis_instrumentation__.Notify(492840)
							return n.isGrant == true
						}() == true) == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(492841)
					if n.isGrant {
						__antithesis_instrumentation__.Notify(492842)
						privileges.GrantPrivilegeToGrantOptions(grantee, true)
					} else {
						__antithesis_instrumentation__.Notify(492843)
						privileges.GrantPrivilegeToGrantOptions(grantee, false)
					}
				} else {
					__antithesis_instrumentation__.Notify(492844)
				}
			}
			__antithesis_instrumentation__.Notify(492810)

			if len(noticeMessage) > 0 {
				__antithesis_instrumentation__.Notify(492845)
				params.p.BufferClientNotice(
					ctx,
					errors.WithHint(
						pgnotice.Newf("%s", noticeMessage),
						"please use WITH GRANT OPTION",
					),
				)
			} else {
				__antithesis_instrumentation__.Notify(492846)
			}
			__antithesis_instrumentation__.Notify(492811)

			err = catprivilege.ValidateSuperuserPrivileges(*privileges, descriptor, n.grantOn)
			if err != nil {
				__antithesis_instrumentation__.Notify(492847)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492848)
			}
			__antithesis_instrumentation__.Notify(492812)

			err = catprivilege.Validate(*privileges, descriptor, n.grantOn)
			if err != nil {
				__antithesis_instrumentation__.Notify(492849)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492850)
			}
		} else {
			__antithesis_instrumentation__.Notify(492851)
		}
		__antithesis_instrumentation__.Notify(492797)

		eventDetails := eventpb.CommonSQLPrivilegeEventDetails{}
		if n.isGrant {
			__antithesis_instrumentation__.Notify(492852)
			eventDetails.GrantedPrivileges = n.desiredprivs.SortedNames()
		} else {
			__antithesis_instrumentation__.Notify(492853)
			eventDetails.RevokedPrivileges = n.desiredprivs.SortedNames()
		}
		__antithesis_instrumentation__.Notify(492798)

		switch d := descriptor.(type) {
		case *dbdesc.Mutable:
			__antithesis_instrumentation__.Notify(492854)
			if err := p.writeDatabaseChangeToBatch(ctx, d, b); err != nil {
				__antithesis_instrumentation__.Notify(492864)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492865)
			}
			__antithesis_instrumentation__.Notify(492855)
			if err := p.createNonDropDatabaseChangeJob(ctx, d.ID,
				fmt.Sprintf("updating privileges for database %d", d.ID)); err != nil {
				__antithesis_instrumentation__.Notify(492866)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492867)
			}
			__antithesis_instrumentation__.Notify(492856)
			for _, grantee := range n.grantees {
				__antithesis_instrumentation__.Notify(492868)
				privs := eventDetails
				privs.Grantee = grantee.Normalized()
				events = append(events, eventLogEntry{
					targetID: int32(d.ID),
					event: &eventpb.ChangeDatabasePrivilege{
						CommonSQLPrivilegeEventDetails: privs,
						DatabaseName:                   (*tree.Name)(&d.Name).String(),
					}})
			}

		case *tabledesc.Mutable:
			__antithesis_instrumentation__.Notify(492857)

			if err := p.createOrUpdateSchemaChangeJob(
				ctx, d,
				fmt.Sprintf("updating privileges for table %d", d.ID),
				descpb.InvalidMutationID,
			); err != nil {
				__antithesis_instrumentation__.Notify(492869)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492870)
			}
			__antithesis_instrumentation__.Notify(492858)
			if !d.Dropped() {
				__antithesis_instrumentation__.Notify(492871)
				if err := p.writeSchemaChangeToBatch(ctx, d, b); err != nil {
					__antithesis_instrumentation__.Notify(492872)
					return err
				} else {
					__antithesis_instrumentation__.Notify(492873)
				}
			} else {
				__antithesis_instrumentation__.Notify(492874)
			}
			__antithesis_instrumentation__.Notify(492859)
			for _, grantee := range n.grantees {
				__antithesis_instrumentation__.Notify(492875)
				privs := eventDetails
				privs.Grantee = grantee.Normalized()
				events = append(events, eventLogEntry{
					targetID: int32(d.ID),
					event: &eventpb.ChangeTablePrivilege{
						CommonSQLPrivilegeEventDetails: privs,
						TableName:                      d.Name,
					}})
			}
		case *typedesc.Mutable:
			__antithesis_instrumentation__.Notify(492860)
			err := p.writeTypeSchemaChange(ctx, d, fmt.Sprintf("updating privileges for type %d", d.ID))
			if err != nil {
				__antithesis_instrumentation__.Notify(492876)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492877)
			}
			__antithesis_instrumentation__.Notify(492861)
			for _, grantee := range n.grantees {
				__antithesis_instrumentation__.Notify(492878)
				privs := eventDetails
				privs.Grantee = grantee.Normalized()
				events = append(events, eventLogEntry{
					targetID: int32(d.ID),
					event: &eventpb.ChangeTypePrivilege{
						CommonSQLPrivilegeEventDetails: privs,
						TypeName:                       d.Name,
					}})
			}
		case *schemadesc.Mutable:
			__antithesis_instrumentation__.Notify(492862)
			if err := p.writeSchemaDescChange(
				ctx,
				d,
				fmt.Sprintf("updating privileges for schema %d", d.ID),
			); err != nil {
				__antithesis_instrumentation__.Notify(492879)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492880)
			}
			__antithesis_instrumentation__.Notify(492863)
			for _, grantee := range n.grantees {
				__antithesis_instrumentation__.Notify(492881)
				privs := eventDetails
				privs.Grantee = grantee.Normalized()
				events = append(events, eventLogEntry{
					targetID: int32(d.ID),
					event: &eventpb.ChangeSchemaPrivilege{
						CommonSQLPrivilegeEventDetails: privs,
						SchemaName:                     d.Name,
					}})
			}
		}
	}
	__antithesis_instrumentation__.Notify(492774)

	if err := p.txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(492882)
		return err
	} else {
		__antithesis_instrumentation__.Notify(492883)
	}
	__antithesis_instrumentation__.Notify(492775)

	if err := params.p.logEvents(params.ctx, events...); err != nil {
		__antithesis_instrumentation__.Notify(492884)
		return err
	} else {
		__antithesis_instrumentation__.Notify(492885)
	}
	__antithesis_instrumentation__.Notify(492776)
	return nil
}

func (*changePrivilegesNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(492886)
	return false, nil
}
func (*changePrivilegesNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(492887)
	return tree.Datums{}
}
func (*changePrivilegesNode) Close(context.Context) { __antithesis_instrumentation__.Notify(492888) }

func getGrantOnObject(targets tree.TargetList, incIAMFunc func(on string)) privilege.ObjectType {
	__antithesis_instrumentation__.Notify(492889)
	switch {
	case targets.Databases != nil:
		__antithesis_instrumentation__.Notify(492890)
		incIAMFunc(sqltelemetry.OnDatabase)
		return privilege.Database
	case targets.AllTablesInSchema:
		__antithesis_instrumentation__.Notify(492891)
		incIAMFunc(sqltelemetry.OnAllTablesInSchema)
		return privilege.Table
	case targets.Schemas != nil:
		__antithesis_instrumentation__.Notify(492892)
		incIAMFunc(sqltelemetry.OnSchema)
		return privilege.Schema
	case targets.Types != nil:
		__antithesis_instrumentation__.Notify(492893)
		incIAMFunc(sqltelemetry.OnType)
		return privilege.Type
	default:
		__antithesis_instrumentation__.Notify(492894)
		incIAMFunc(sqltelemetry.OnTable)
		return privilege.Table
	}
}

func (p *planner) validateRoles(
	ctx context.Context, roles []security.SQLUsername, isPublicValid bool,
) error {
	__antithesis_instrumentation__.Notify(492895)
	users, err := p.GetAllRoles(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(492899)
		return err
	} else {
		__antithesis_instrumentation__.Notify(492900)
	}
	__antithesis_instrumentation__.Notify(492896)
	if isPublicValid {
		__antithesis_instrumentation__.Notify(492901)
		users[security.PublicRoleName()] = true
	} else {
		__antithesis_instrumentation__.Notify(492902)
	}
	__antithesis_instrumentation__.Notify(492897)
	for i, grantee := range roles {
		__antithesis_instrumentation__.Notify(492903)
		if _, ok := users[grantee]; !ok {
			__antithesis_instrumentation__.Notify(492904)
			sqlName := tree.Name(roles[i].Normalized())
			return errors.Errorf("user or role %s does not exist", &sqlName)
		} else {
			__antithesis_instrumentation__.Notify(492905)
		}
	}
	__antithesis_instrumentation__.Notify(492898)

	return nil
}
