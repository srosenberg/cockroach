package catpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type DefaultPrivilegesRole struct {
	Role        security.SQLUsername
	ForAllRoles bool
}

func (p *DefaultPrivilegesForRole) toDefaultPrivilegesRole() DefaultPrivilegesRole {
	__antithesis_instrumentation__.Notify(249315)
	if p.IsExplicitRole() {
		__antithesis_instrumentation__.Notify(249317)
		return DefaultPrivilegesRole{
			Role: p.GetExplicitRole().UserProto.Decode(),
		}
	} else {
		__antithesis_instrumentation__.Notify(249318)
	}
	__antithesis_instrumentation__.Notify(249316)
	return DefaultPrivilegesRole{
		ForAllRoles: true,
	}
}

func (r DefaultPrivilegesRole) lessThan(other DefaultPrivilegesRole) bool {
	__antithesis_instrumentation__.Notify(249319)

	if r.ForAllRoles {
		__antithesis_instrumentation__.Notify(249322)
		return false
	} else {
		__antithesis_instrumentation__.Notify(249323)
	}
	__antithesis_instrumentation__.Notify(249320)
	if other.ForAllRoles {
		__antithesis_instrumentation__.Notify(249324)
		return true
	} else {
		__antithesis_instrumentation__.Notify(249325)
	}
	__antithesis_instrumentation__.Notify(249321)

	return r.Role.LessThan(other.Role)
}

func (p *DefaultPrivilegeDescriptor) FindUserIndex(role DefaultPrivilegesRole) int {
	__antithesis_instrumentation__.Notify(249326)
	idx := sort.Search(len(p.DefaultPrivilegesPerRole), func(i int) bool {
		__antithesis_instrumentation__.Notify(249329)
		return !p.DefaultPrivilegesPerRole[i].toDefaultPrivilegesRole().lessThan(role)
	})
	__antithesis_instrumentation__.Notify(249327)
	if idx < len(p.DefaultPrivilegesPerRole) && func() bool {
		__antithesis_instrumentation__.Notify(249330)
		return p.DefaultPrivilegesPerRole[idx].toDefaultPrivilegesRole() == role == true
	}() == true {
		__antithesis_instrumentation__.Notify(249331)
		return idx
	} else {
		__antithesis_instrumentation__.Notify(249332)
	}
	__antithesis_instrumentation__.Notify(249328)
	return -1
}

func (p *DefaultPrivilegeDescriptor) FindOrCreateUser(
	role DefaultPrivilegesRole,
) *DefaultPrivilegesForRole {
	__antithesis_instrumentation__.Notify(249333)
	idx := sort.Search(len(p.DefaultPrivilegesPerRole), func(i int) bool {
		__antithesis_instrumentation__.Notify(249336)
		return !p.DefaultPrivilegesPerRole[i].toDefaultPrivilegesRole().lessThan(role)
	})
	__antithesis_instrumentation__.Notify(249334)
	if idx == len(p.DefaultPrivilegesPerRole) {
		__antithesis_instrumentation__.Notify(249337)

		p.DefaultPrivilegesPerRole = append(p.DefaultPrivilegesPerRole,
			InitDefaultPrivilegesForRole(role, p.Type),
		)
	} else {
		__antithesis_instrumentation__.Notify(249338)
		if p.DefaultPrivilegesPerRole[idx].toDefaultPrivilegesRole() == role {
			__antithesis_instrumentation__.Notify(249339)

		} else {
			__antithesis_instrumentation__.Notify(249340)

			p.DefaultPrivilegesPerRole = append(p.DefaultPrivilegesPerRole, DefaultPrivilegesForRole{})
			copy(p.DefaultPrivilegesPerRole[idx+1:], p.DefaultPrivilegesPerRole[idx:])
			p.DefaultPrivilegesPerRole[idx] = InitDefaultPrivilegesForRole(role, p.Type)
		}
	}
	__antithesis_instrumentation__.Notify(249335)
	return &p.DefaultPrivilegesPerRole[idx]
}

func InitDefaultPrivilegesForRole(
	role DefaultPrivilegesRole,
	defaultPrivilegeDescType DefaultPrivilegeDescriptor_DefaultPrivilegeDescriptorType,
) DefaultPrivilegesForRole {
	__antithesis_instrumentation__.Notify(249341)
	var defaultPrivilegesRole isDefaultPrivilegesForRole_Role
	if role.ForAllRoles {
		__antithesis_instrumentation__.Notify(249344)
		defaultPrivilegesRole = &DefaultPrivilegesForRole_ForAllRoles{
			ForAllRoles: &DefaultPrivilegesForRole_ForAllRolesPseudoRole{
				PublicHasUsageOnTypes: true,
			},
		}
		return DefaultPrivilegesForRole{
			Role:                       defaultPrivilegesRole,
			DefaultPrivilegesPerObject: map[tree.AlterDefaultPrivilegesTargetObject]PrivilegeDescriptor{},
		}
	} else {
		__antithesis_instrumentation__.Notify(249345)
	}
	__antithesis_instrumentation__.Notify(249342)

	if defaultPrivilegeDescType == DefaultPrivilegeDescriptor_DATABASE {
		__antithesis_instrumentation__.Notify(249346)
		defaultPrivilegesRole = &DefaultPrivilegesForRole_ExplicitRole_{
			ExplicitRole: &DefaultPrivilegesForRole_ExplicitRole{
				UserProto:                       role.Role.EncodeProto(),
				PublicHasUsageOnTypes:           true,
				RoleHasAllPrivilegesOnTables:    true,
				RoleHasAllPrivilegesOnSequences: true,
				RoleHasAllPrivilegesOnSchemas:   true,
				RoleHasAllPrivilegesOnTypes:     true,
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(249347)

		defaultPrivilegesRole = &DefaultPrivilegesForRole_ExplicitRole_{
			ExplicitRole: &DefaultPrivilegesForRole_ExplicitRole{
				UserProto: role.Role.EncodeProto(),
			},
		}
	}
	__antithesis_instrumentation__.Notify(249343)
	return DefaultPrivilegesForRole{
		Role:                       defaultPrivilegesRole,
		DefaultPrivilegesPerObject: map[tree.AlterDefaultPrivilegesTargetObject]PrivilegeDescriptor{},
	}
}

func (p *DefaultPrivilegeDescriptor) RemoveUser(role DefaultPrivilegesRole) {
	__antithesis_instrumentation__.Notify(249348)
	idx := p.FindUserIndex(role)
	if idx == -1 {
		__antithesis_instrumentation__.Notify(249350)

		return
	} else {
		__antithesis_instrumentation__.Notify(249351)
	}
	__antithesis_instrumentation__.Notify(249349)
	p.DefaultPrivilegesPerRole = append(p.DefaultPrivilegesPerRole[:idx], p.DefaultPrivilegesPerRole[idx+1:]...)
}

func (p *DefaultPrivilegeDescriptor) Validate() error {
	__antithesis_instrumentation__.Notify(249352)
	entryForAllRolesFound := false
	for i, defaultPrivilegesForRole := range p.DefaultPrivilegesPerRole {
		__antithesis_instrumentation__.Notify(249354)
		if !defaultPrivilegesForRole.IsExplicitRole() {
			__antithesis_instrumentation__.Notify(249357)
			if entryForAllRolesFound {
				__antithesis_instrumentation__.Notify(249359)
				return errors.AssertionFailedf("multiple entries found in map for all roles")
			} else {
				__antithesis_instrumentation__.Notify(249360)
			}
			__antithesis_instrumentation__.Notify(249358)
			entryForAllRolesFound = true
		} else {
			__antithesis_instrumentation__.Notify(249361)
		}
		__antithesis_instrumentation__.Notify(249355)
		if i+1 < len(p.DefaultPrivilegesPerRole) && func() bool {
			__antithesis_instrumentation__.Notify(249362)
			return !defaultPrivilegesForRole.toDefaultPrivilegesRole().
				lessThan(p.DefaultPrivilegesPerRole[i+1].toDefaultPrivilegesRole()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(249363)
			return errors.AssertionFailedf("default privilege list is not sorted")
		} else {
			__antithesis_instrumentation__.Notify(249364)
		}
		__antithesis_instrumentation__.Notify(249356)
		for objectType, defaultPrivileges := range defaultPrivilegesForRole.DefaultPrivilegesPerObject {
			__antithesis_instrumentation__.Notify(249365)
			privilegeObjectType := objectType.ToPrivilegeObjectType()
			valid, u, remaining := defaultPrivileges.IsValidPrivilegesForObjectType(privilegeObjectType)
			if !valid {
				__antithesis_instrumentation__.Notify(249366)
				return errors.AssertionFailedf("user %s must not have %s privileges on %s",
					u.User(), privilege.ListFromBitField(remaining, privilege.Any), privilegeObjectType)
			} else {
				__antithesis_instrumentation__.Notify(249367)
			}
		}
	}
	__antithesis_instrumentation__.Notify(249353)

	return nil
}

func (p DefaultPrivilegesForRole) IsExplicitRole() bool {
	__antithesis_instrumentation__.Notify(249368)
	return p.GetExplicitRole() != nil
}
