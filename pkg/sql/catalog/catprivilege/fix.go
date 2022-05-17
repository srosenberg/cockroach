package catprivilege

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
)

func MaybeFixUsagePrivForTablesAndDBs(ptr **catpb.PrivilegeDescriptor) bool {
	__antithesis_instrumentation__.Notify(250812)
	if *ptr == nil {
		__antithesis_instrumentation__.Notify(250816)
		*ptr = &catpb.PrivilegeDescriptor{}
	} else {
		__antithesis_instrumentation__.Notify(250817)
	}
	__antithesis_instrumentation__.Notify(250813)
	p := *ptr

	if p.Version > catpb.InitialVersion {
		__antithesis_instrumentation__.Notify(250818)

		return false
	} else {
		__antithesis_instrumentation__.Notify(250819)
	}
	__antithesis_instrumentation__.Notify(250814)

	modified := false
	for i := range p.Users {
		__antithesis_instrumentation__.Notify(250820)

		userPrivileges := &p.Users[i]

		if privilege.USAGE.Mask()&userPrivileges.Privileges != 0 {
			__antithesis_instrumentation__.Notify(250821)

			userPrivileges.Privileges = (userPrivileges.Privileges - privilege.USAGE.Mask()) | privilege.ZONECONFIG.Mask()
			modified = true
		} else {
			__antithesis_instrumentation__.Notify(250822)
		}
	}
	__antithesis_instrumentation__.Notify(250815)

	return modified
}

func MaybeFixPrivileges(
	ptr **catpb.PrivilegeDescriptor,
	parentID, parentSchemaID descpb.ID,
	objectType privilege.ObjectType,
	objectName string,
) bool {
	__antithesis_instrumentation__.Notify(250823)
	if *ptr == nil {
		__antithesis_instrumentation__.Notify(250831)
		*ptr = &catpb.PrivilegeDescriptor{}
	} else {
		__antithesis_instrumentation__.Notify(250832)
	}
	__antithesis_instrumentation__.Notify(250824)
	p := *ptr
	allowedPrivilegesBits := privilege.GetValidPrivilegesForObject(objectType).ToBitField()
	systemPrivs := SystemSuperuserPrivileges(descpb.NameInfo{
		ParentID:       parentID,
		ParentSchemaID: parentSchemaID,
		Name:           objectName,
	})
	if systemPrivs != nil {
		__antithesis_instrumentation__.Notify(250833)

		allowedPrivilegesBits = systemPrivs.ToBitField()
	} else {
		__antithesis_instrumentation__.Notify(250834)
	}
	__antithesis_instrumentation__.Notify(250825)

	changed := false

	fixSuperUser := func(user security.SQLUsername) {
		__antithesis_instrumentation__.Notify(250835)
		privs := p.FindOrCreateUser(user)
		oldPrivilegeBits := privs.Privileges
		if oldPrivilegeBits != allowedPrivilegesBits {
			__antithesis_instrumentation__.Notify(250836)
			if privilege.ALL.IsSetIn(allowedPrivilegesBits) {
				__antithesis_instrumentation__.Notify(250838)
				privs.Privileges = privilege.ALL.Mask()
			} else {
				__antithesis_instrumentation__.Notify(250839)
				privs.Privileges = allowedPrivilegesBits
			}
			__antithesis_instrumentation__.Notify(250837)
			changed = (privs.Privileges != oldPrivilegeBits) || func() bool {
				__antithesis_instrumentation__.Notify(250840)
				return changed == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(250841)
		}
	}
	__antithesis_instrumentation__.Notify(250826)

	fixSuperUser(security.RootUserName())
	fixSuperUser(security.AdminRoleName())

	if objectType == privilege.Table || func() bool {
		__antithesis_instrumentation__.Notify(250842)
		return objectType == privilege.Database == true
	}() == true {
		__antithesis_instrumentation__.Notify(250843)
		changed = MaybeFixUsagePrivForTablesAndDBs(&p) || func() bool {
			__antithesis_instrumentation__.Notify(250844)
			return changed == true
		}() == true
	} else {
		__antithesis_instrumentation__.Notify(250845)
	}
	__antithesis_instrumentation__.Notify(250827)

	for i := range p.Users {
		__antithesis_instrumentation__.Notify(250846)

		u := &p.Users[i]
		if u.User().IsRootUser() || func() bool {
			__antithesis_instrumentation__.Notify(250849)
			return u.User().IsAdminRole() == true
		}() == true {
			__antithesis_instrumentation__.Notify(250850)

			continue
		} else {
			__antithesis_instrumentation__.Notify(250851)
		}
		__antithesis_instrumentation__.Notify(250847)

		if u.Privileges&allowedPrivilegesBits != u.Privileges {
			__antithesis_instrumentation__.Notify(250852)
			changed = true
		} else {
			__antithesis_instrumentation__.Notify(250853)
		}
		__antithesis_instrumentation__.Notify(250848)
		u.Privileges &= allowedPrivilegesBits
	}
	__antithesis_instrumentation__.Notify(250828)

	if p.Owner().Undefined() {
		__antithesis_instrumentation__.Notify(250854)
		if systemPrivs != nil {
			__antithesis_instrumentation__.Notify(250856)
			p.SetOwner(security.NodeUserName())
		} else {
			__antithesis_instrumentation__.Notify(250857)
			p.SetOwner(security.RootUserName())
		}
		__antithesis_instrumentation__.Notify(250855)
		changed = true
	} else {
		__antithesis_instrumentation__.Notify(250858)
	}
	__antithesis_instrumentation__.Notify(250829)

	if p.Version < catpb.Version21_2 {
		__antithesis_instrumentation__.Notify(250859)
		p.SetVersion(catpb.Version21_2)
		changed = true
	} else {
		__antithesis_instrumentation__.Notify(250860)
	}
	__antithesis_instrumentation__.Notify(250830)
	return changed
}

func MaybeUpdateGrantOptions(p *catpb.PrivilegeDescriptor) bool {
	__antithesis_instrumentation__.Notify(250861)

	if p.CheckGrantOptions(security.AdminRoleName(), privilege.List{privilege.SELECT}) {
		__antithesis_instrumentation__.Notify(250864)
		return false
	} else {
		__antithesis_instrumentation__.Notify(250865)
	}
	__antithesis_instrumentation__.Notify(250862)

	changed := false
	for i := range p.Users {
		__antithesis_instrumentation__.Notify(250866)
		u := &p.Users[i]
		if privilege.ALL.IsSetIn(u.Privileges) {
			__antithesis_instrumentation__.Notify(250868)
			if !privilege.ALL.IsSetIn(u.WithGrantOption) {
				__antithesis_instrumentation__.Notify(250870)
				changed = true
			} else {
				__antithesis_instrumentation__.Notify(250871)
			}
			__antithesis_instrumentation__.Notify(250869)
			u.WithGrantOption = privilege.ALL.Mask()
			continue
		} else {
			__antithesis_instrumentation__.Notify(250872)
		}
		__antithesis_instrumentation__.Notify(250867)
		if privilege.GRANT.IsSetIn(u.Privileges) {
			__antithesis_instrumentation__.Notify(250873)
			if u.Privileges != u.WithGrantOption {
				__antithesis_instrumentation__.Notify(250875)
				changed = true
			} else {
				__antithesis_instrumentation__.Notify(250876)
			}
			__antithesis_instrumentation__.Notify(250874)
			u.WithGrantOption |= u.Privileges
		} else {
			__antithesis_instrumentation__.Notify(250877)
		}
	}
	__antithesis_instrumentation__.Notify(250863)

	return changed
}
