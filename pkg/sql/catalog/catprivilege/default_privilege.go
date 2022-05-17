package catprivilege

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

var _ catalog.DefaultPrivilegeDescriptor = &immutable{}
var _ catalog.DefaultPrivilegeDescriptor = &Mutable{}

type immutable struct {
	defaultPrivilegeDescriptor *catpb.DefaultPrivilegeDescriptor
}

type Mutable struct {
	immutable
}

func MakeDefaultPrivilegeDescriptor(
	typ catpb.DefaultPrivilegeDescriptor_DefaultPrivilegeDescriptorType,
) *catpb.DefaultPrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(250662)
	var defaultPrivilegesForRole []catpb.DefaultPrivilegesForRole
	return &catpb.DefaultPrivilegeDescriptor{
		DefaultPrivilegesPerRole: defaultPrivilegesForRole,
		Type:                     typ,
	}
}

func MakeDefaultPrivileges(
	defaultPrivilegeDescriptor *catpb.DefaultPrivilegeDescriptor,
) catalog.DefaultPrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(250663)
	return &immutable{
		defaultPrivilegeDescriptor: defaultPrivilegeDescriptor,
	}
}

func NewMutableDefaultPrivileges(
	defaultPrivilegeDescriptor *catpb.DefaultPrivilegeDescriptor,
) *Mutable {
	__antithesis_instrumentation__.Notify(250664)
	return &Mutable{
		immutable{
			defaultPrivilegeDescriptor: defaultPrivilegeDescriptor,
		}}
}

func (d *immutable) grantOrRevokeDefaultPrivilegesHelper(
	defaultPrivilegesForRole *catpb.DefaultPrivilegesForRole,
	role catpb.DefaultPrivilegesRole,
	targetObject tree.AlterDefaultPrivilegesTargetObject,
	grantee security.SQLUsername,
	privList privilege.List,
	withGrantOption bool,
	isGrant bool,
	deprecateGrant bool,
) {
	__antithesis_instrumentation__.Notify(250665)
	defaultPrivileges := defaultPrivilegesForRole.DefaultPrivilegesPerObject[targetObject]

	if d.IsDatabaseDefaultPrivilege() {
		__antithesis_instrumentation__.Notify(250670)
		expandPrivileges(defaultPrivilegesForRole, role, &defaultPrivileges, targetObject)
	} else {
		__antithesis_instrumentation__.Notify(250671)
	}
	__antithesis_instrumentation__.Notify(250666)
	if isGrant {
		__antithesis_instrumentation__.Notify(250672)
		defaultPrivileges.Grant(grantee, privList, withGrantOption)
	} else {
		__antithesis_instrumentation__.Notify(250673)
		defaultPrivileges.Revoke(grantee, privList, targetObject.ToPrivilegeObjectType(), withGrantOption)
	}
	__antithesis_instrumentation__.Notify(250667)

	if deprecateGrant {
		__antithesis_instrumentation__.Notify(250674)
		defaultPrivileges.GrantPrivilegeToGrantOptions(grantee, isGrant)
	} else {
		__antithesis_instrumentation__.Notify(250675)
	}
	__antithesis_instrumentation__.Notify(250668)

	if d.IsDatabaseDefaultPrivilege() {
		__antithesis_instrumentation__.Notify(250676)
		foldPrivileges(defaultPrivilegesForRole, role, &defaultPrivileges, targetObject)
	} else {
		__antithesis_instrumentation__.Notify(250677)
	}
	__antithesis_instrumentation__.Notify(250669)
	defaultPrivilegesForRole.DefaultPrivilegesPerObject[targetObject] = defaultPrivileges
}

func (d *Mutable) GrantDefaultPrivileges(
	role catpb.DefaultPrivilegesRole,
	privileges privilege.List,
	grantees []security.SQLUsername,
	targetObject tree.AlterDefaultPrivilegesTargetObject,
	withGrantOption bool,
	deprecateGrant bool,
) {
	__antithesis_instrumentation__.Notify(250678)
	defaultPrivilegesForRole := d.defaultPrivilegeDescriptor.FindOrCreateUser(role)
	for _, grantee := range grantees {
		__antithesis_instrumentation__.Notify(250679)
		d.grantOrRevokeDefaultPrivilegesHelper(defaultPrivilegesForRole, role, targetObject, grantee, privileges, withGrantOption, true, deprecateGrant)
	}
}

func (d *Mutable) RevokeDefaultPrivileges(
	role catpb.DefaultPrivilegesRole,
	privileges privilege.List,
	grantees []security.SQLUsername,
	targetObject tree.AlterDefaultPrivilegesTargetObject,
	grantOptionFor bool,
	deprecateGrant bool,
) {
	__antithesis_instrumentation__.Notify(250680)
	defaultPrivilegesForRole := d.defaultPrivilegeDescriptor.FindOrCreateUser(role)
	for _, grantee := range grantees {
		__antithesis_instrumentation__.Notify(250684)
		d.grantOrRevokeDefaultPrivilegesHelper(defaultPrivilegesForRole, role, targetObject, grantee, privileges, grantOptionFor, false, deprecateGrant)
	}
	__antithesis_instrumentation__.Notify(250681)

	defaultPrivilegesPerObject := defaultPrivilegesForRole.DefaultPrivilegesPerObject

	for _, defaultPrivs := range defaultPrivilegesPerObject {
		__antithesis_instrumentation__.Notify(250685)
		if len(defaultPrivs.Users) != 0 {
			__antithesis_instrumentation__.Notify(250686)
			return
		} else {
			__antithesis_instrumentation__.Notify(250687)
		}
	}
	__antithesis_instrumentation__.Notify(250682)

	if defaultPrivilegesForRole.IsExplicitRole() && func() bool {
		__antithesis_instrumentation__.Notify(250688)
		return d.IsDatabaseDefaultPrivilege() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(250689)
		return (!GetRoleHasAllPrivilegesOnTargetObject(defaultPrivilegesForRole, tree.Tables) || func() bool {
			__antithesis_instrumentation__.Notify(250690)
			return !GetRoleHasAllPrivilegesOnTargetObject(defaultPrivilegesForRole, tree.Sequences) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(250691)
			return !GetRoleHasAllPrivilegesOnTargetObject(defaultPrivilegesForRole, tree.Types) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(250692)
			return !GetRoleHasAllPrivilegesOnTargetObject(defaultPrivilegesForRole, tree.Schemas) == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(250693)
		return !GetPublicHasUsageOnTypes(defaultPrivilegesForRole) == true
	}() == true {
		__antithesis_instrumentation__.Notify(250694)
		return
	} else {
		__antithesis_instrumentation__.Notify(250695)
	}
	__antithesis_instrumentation__.Notify(250683)

	d.defaultPrivilegeDescriptor.RemoveUser(role)
}

func CreatePrivilegesFromDefaultPrivileges(
	dbDefaultPrivilegeDescriptor catalog.DefaultPrivilegeDescriptor,
	schemaDefaultPrivilegeDescriptor catalog.DefaultPrivilegeDescriptor,
	dbID descpb.ID,
	user security.SQLUsername,
	targetObject tree.AlterDefaultPrivilegesTargetObject,
	databasePrivileges *catpb.PrivilegeDescriptor,
) *catpb.PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(250696)

	if dbID == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(250700)
		return catpb.NewBasePrivilegeDescriptor(security.NodeUserName())
	} else {
		__antithesis_instrumentation__.Notify(250701)
	}
	__antithesis_instrumentation__.Notify(250697)

	defaultPrivilegeDescriptors := []catalog.DefaultPrivilegeDescriptor{
		dbDefaultPrivilegeDescriptor,
	}

	if schemaDefaultPrivilegeDescriptor != nil {
		__antithesis_instrumentation__.Notify(250702)
		defaultPrivilegeDescriptors = append(defaultPrivilegeDescriptors, schemaDefaultPrivilegeDescriptor)
	} else {
		__antithesis_instrumentation__.Notify(250703)
	}
	__antithesis_instrumentation__.Notify(250698)

	newPrivs := catpb.NewBasePrivilegeDescriptor(user)
	role := catpb.DefaultPrivilegesRole{Role: user}
	for _, d := range defaultPrivilegeDescriptors {
		__antithesis_instrumentation__.Notify(250704)
		if defaultPrivilegesForRole, found := d.GetDefaultPrivilegesForRole(role); !found {
			__antithesis_instrumentation__.Notify(250706)

			defaultPrivilegesForCreatorRole := catpb.InitDefaultPrivilegesForRole(role, d.GetDefaultPrivilegeDescriptorType())
			for _, user := range GetUserPrivilegesForObject(defaultPrivilegesForCreatorRole, targetObject) {
				__antithesis_instrumentation__.Notify(250707)
				applyDefaultPrivileges(
					newPrivs,
					user.UserProto.Decode(),
					privilege.ListFromBitField(user.Privileges, targetObject.ToPrivilegeObjectType()),
					privilege.ListFromBitField(user.WithGrantOption, targetObject.ToPrivilegeObjectType()),
				)
			}
		} else {
			__antithesis_instrumentation__.Notify(250708)

			for _, user := range GetUserPrivilegesForObject(*defaultPrivilegesForRole, targetObject) {
				__antithesis_instrumentation__.Notify(250709)
				applyDefaultPrivileges(
					newPrivs,
					user.UserProto.Decode(),
					privilege.ListFromBitField(user.Privileges, targetObject.ToPrivilegeObjectType()),
					privilege.ListFromBitField(user.WithGrantOption, targetObject.ToPrivilegeObjectType()),
				)
			}
		}
		__antithesis_instrumentation__.Notify(250705)

		defaultPrivilegesForAllRoles, found := d.GetDefaultPrivilegesForRole(catpb.DefaultPrivilegesRole{ForAllRoles: true})
		if found {
			__antithesis_instrumentation__.Notify(250710)
			for _, user := range GetUserPrivilegesForObject(*defaultPrivilegesForAllRoles, targetObject) {
				__antithesis_instrumentation__.Notify(250711)
				applyDefaultPrivileges(
					newPrivs,
					user.UserProto.Decode(),
					privilege.ListFromBitField(user.Privileges, targetObject.ToPrivilegeObjectType()),
					privilege.ListFromBitField(user.WithGrantOption, targetObject.ToPrivilegeObjectType()),
				)
			}
		} else {
			__antithesis_instrumentation__.Notify(250712)
		}
	}
	__antithesis_instrumentation__.Notify(250699)

	newPrivs.Version = catpb.Version21_2
	return newPrivs
}

func (d *immutable) ForEachDefaultPrivilegeForRole(
	f func(defaultPrivilegesForRole catpb.DefaultPrivilegesForRole) error,
) error {
	__antithesis_instrumentation__.Notify(250713)
	if d.defaultPrivilegeDescriptor == nil {
		__antithesis_instrumentation__.Notify(250716)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(250717)
	}
	__antithesis_instrumentation__.Notify(250714)
	for _, defaultPrivilegesForRole := range d.defaultPrivilegeDescriptor.DefaultPrivilegesPerRole {
		__antithesis_instrumentation__.Notify(250718)
		if err := f(defaultPrivilegesForRole); err != nil {
			__antithesis_instrumentation__.Notify(250719)
			return err
		} else {
			__antithesis_instrumentation__.Notify(250720)
		}
	}
	__antithesis_instrumentation__.Notify(250715)
	return nil
}

func (d *immutable) GetDefaultPrivilegesForRole(
	role catpb.DefaultPrivilegesRole,
) (*catpb.DefaultPrivilegesForRole, bool) {
	__antithesis_instrumentation__.Notify(250721)
	idx := d.defaultPrivilegeDescriptor.FindUserIndex(role)
	if idx == -1 {
		__antithesis_instrumentation__.Notify(250723)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(250724)
	}
	__antithesis_instrumentation__.Notify(250722)
	return &d.defaultPrivilegeDescriptor.DefaultPrivilegesPerRole[idx], true
}

func (d *immutable) GetDefaultPrivilegeDescriptorType() catpb.DefaultPrivilegeDescriptor_DefaultPrivilegeDescriptorType {
	__antithesis_instrumentation__.Notify(250725)
	return d.defaultPrivilegeDescriptor.Type
}

func (d *immutable) IsDatabaseDefaultPrivilege() bool {
	__antithesis_instrumentation__.Notify(250726)
	return d.defaultPrivilegeDescriptor.Type == catpb.DefaultPrivilegeDescriptor_DATABASE
}

func foldPrivileges(
	defaultPrivilegesForRole *catpb.DefaultPrivilegesForRole,
	role catpb.DefaultPrivilegesRole,
	privileges *catpb.PrivilegeDescriptor,
	targetObject tree.AlterDefaultPrivilegesTargetObject,
) {
	__antithesis_instrumentation__.Notify(250727)
	if targetObject == tree.Types && func() bool {
		__antithesis_instrumentation__.Notify(250730)
		return privileges.CheckPrivilege(security.PublicRoleName(), privilege.USAGE) == true
	}() == true {
		__antithesis_instrumentation__.Notify(250731)
		publicUser, ok := privileges.FindUser(security.PublicRoleName())
		if ok {
			__antithesis_instrumentation__.Notify(250732)
			if !privilege.USAGE.IsSetIn(publicUser.WithGrantOption) {
				__antithesis_instrumentation__.Notify(250733)
				setPublicHasUsageOnTypes(defaultPrivilegesForRole, true)
				privileges.Revoke(
					security.PublicRoleName(),
					privilege.List{privilege.USAGE},
					privilege.Type,
					false,
				)
			} else {
				__antithesis_instrumentation__.Notify(250734)
			}
		} else {
			__antithesis_instrumentation__.Notify(250735)
		}
	} else {
		__antithesis_instrumentation__.Notify(250736)
	}
	__antithesis_instrumentation__.Notify(250728)

	if role.ForAllRoles {
		__antithesis_instrumentation__.Notify(250737)
		return
	} else {
		__antithesis_instrumentation__.Notify(250738)
	}
	__antithesis_instrumentation__.Notify(250729)
	if privileges.HasAllPrivileges(role.Role, targetObject.ToPrivilegeObjectType()) {
		__antithesis_instrumentation__.Notify(250739)
		user := privileges.FindOrCreateUser(role.Role)
		if user.WithGrantOption == 0 {
			__antithesis_instrumentation__.Notify(250740)
			setRoleHasAllOnTargetObject(defaultPrivilegesForRole, true, targetObject)
			privileges.RemoveUser(role.Role)
		} else {
			__antithesis_instrumentation__.Notify(250741)
		}
	} else {
		__antithesis_instrumentation__.Notify(250742)
	}
}

func expandPrivileges(
	defaultPrivilegesForRole *catpb.DefaultPrivilegesForRole,
	role catpb.DefaultPrivilegesRole,
	privileges *catpb.PrivilegeDescriptor,
	targetObject tree.AlterDefaultPrivilegesTargetObject,
) {
	__antithesis_instrumentation__.Notify(250743)
	if targetObject == tree.Types && func() bool {
		__antithesis_instrumentation__.Notify(250746)
		return GetPublicHasUsageOnTypes(defaultPrivilegesForRole) == true
	}() == true {
		__antithesis_instrumentation__.Notify(250747)
		privileges.Grant(security.PublicRoleName(), privilege.List{privilege.USAGE}, false)
		setPublicHasUsageOnTypes(defaultPrivilegesForRole, false)
	} else {
		__antithesis_instrumentation__.Notify(250748)
	}
	__antithesis_instrumentation__.Notify(250744)

	if role.ForAllRoles {
		__antithesis_instrumentation__.Notify(250749)
		return
	} else {
		__antithesis_instrumentation__.Notify(250750)
	}
	__antithesis_instrumentation__.Notify(250745)
	if GetRoleHasAllPrivilegesOnTargetObject(defaultPrivilegesForRole, targetObject) {
		__antithesis_instrumentation__.Notify(250751)
		privileges.Grant(defaultPrivilegesForRole.GetExplicitRole().UserProto.Decode(), privilege.List{privilege.ALL}, false)
		setRoleHasAllOnTargetObject(defaultPrivilegesForRole, false, targetObject)
	} else {
		__antithesis_instrumentation__.Notify(250752)
	}
}

func GetUserPrivilegesForObject(
	p catpb.DefaultPrivilegesForRole, targetObject tree.AlterDefaultPrivilegesTargetObject,
) []catpb.UserPrivileges {
	__antithesis_instrumentation__.Notify(250753)
	var userPrivileges []catpb.UserPrivileges
	if privileges, ok := p.DefaultPrivilegesPerObject[targetObject]; ok {
		__antithesis_instrumentation__.Notify(250758)
		userPrivileges = privileges.Users
	} else {
		__antithesis_instrumentation__.Notify(250759)
	}
	__antithesis_instrumentation__.Notify(250754)
	if GetPublicHasUsageOnTypes(&p) && func() bool {
		__antithesis_instrumentation__.Notify(250760)
		return targetObject == tree.Types == true
	}() == true {
		__antithesis_instrumentation__.Notify(250761)
		userPrivileges = append(userPrivileges, catpb.UserPrivileges{
			UserProto:  security.PublicRoleName().EncodeProto(),
			Privileges: privilege.USAGE.Mask(),
		})
	} else {
		__antithesis_instrumentation__.Notify(250762)
	}
	__antithesis_instrumentation__.Notify(250755)

	if !p.IsExplicitRole() {
		__antithesis_instrumentation__.Notify(250763)
		return userPrivileges
	} else {
		__antithesis_instrumentation__.Notify(250764)
	}
	__antithesis_instrumentation__.Notify(250756)
	userProto := p.GetExplicitRole().UserProto
	if GetRoleHasAllPrivilegesOnTargetObject(&p, targetObject) {
		__antithesis_instrumentation__.Notify(250765)
		return append(userPrivileges, catpb.UserPrivileges{
			UserProto:  userProto,
			Privileges: privilege.ALL.Mask(),
		})
	} else {
		__antithesis_instrumentation__.Notify(250766)
	}
	__antithesis_instrumentation__.Notify(250757)
	return userPrivileges
}

func GetPublicHasUsageOnTypes(defaultPrivilegesForRole *catpb.DefaultPrivilegesForRole) bool {
	__antithesis_instrumentation__.Notify(250767)
	if defaultPrivilegesForRole.IsExplicitRole() {
		__antithesis_instrumentation__.Notify(250769)
		return defaultPrivilegesForRole.GetExplicitRole().PublicHasUsageOnTypes
	} else {
		__antithesis_instrumentation__.Notify(250770)
	}
	__antithesis_instrumentation__.Notify(250768)
	return defaultPrivilegesForRole.GetForAllRoles().PublicHasUsageOnTypes
}

func GetRoleHasAllPrivilegesOnTargetObject(
	defaultPrivilegesForRole *catpb.DefaultPrivilegesForRole,
	targetObject tree.AlterDefaultPrivilegesTargetObject,
) bool {
	__antithesis_instrumentation__.Notify(250771)
	if !defaultPrivilegesForRole.IsExplicitRole() {
		__antithesis_instrumentation__.Notify(250773)

		return false
	} else {
		__antithesis_instrumentation__.Notify(250774)
	}
	__antithesis_instrumentation__.Notify(250772)
	switch targetObject {
	case tree.Tables:
		__antithesis_instrumentation__.Notify(250775)
		return defaultPrivilegesForRole.GetExplicitRole().RoleHasAllPrivilegesOnTables
	case tree.Sequences:
		__antithesis_instrumentation__.Notify(250776)
		return defaultPrivilegesForRole.GetExplicitRole().RoleHasAllPrivilegesOnSequences
	case tree.Types:
		__antithesis_instrumentation__.Notify(250777)
		return defaultPrivilegesForRole.GetExplicitRole().RoleHasAllPrivilegesOnTypes
	case tree.Schemas:
		__antithesis_instrumentation__.Notify(250778)
		return defaultPrivilegesForRole.GetExplicitRole().RoleHasAllPrivilegesOnSchemas
	default:
		__antithesis_instrumentation__.Notify(250779)
		panic(fmt.Sprintf("unknown target object %s", targetObject))
	}
}

func setPublicHasUsageOnTypes(
	defaultPrivilegesForRole *catpb.DefaultPrivilegesForRole, publicHasUsageOnTypes bool,
) {
	__antithesis_instrumentation__.Notify(250780)
	if defaultPrivilegesForRole.IsExplicitRole() {
		__antithesis_instrumentation__.Notify(250781)
		defaultPrivilegesForRole.GetExplicitRole().PublicHasUsageOnTypes = publicHasUsageOnTypes
	} else {
		__antithesis_instrumentation__.Notify(250782)
		defaultPrivilegesForRole.GetForAllRoles().PublicHasUsageOnTypes = publicHasUsageOnTypes
	}
}

func applyDefaultPrivileges(
	p *catpb.PrivilegeDescriptor,
	user security.SQLUsername,
	privList privilege.List,
	grantOptionList privilege.List,
) {
	__antithesis_instrumentation__.Notify(250783)
	userPriv := p.FindOrCreateUser(user)
	if privilege.ALL.IsSetIn(userPriv.WithGrantOption) && func() bool {
		__antithesis_instrumentation__.Notify(250787)
		return privilege.ALL.IsSetIn(userPriv.Privileges) == true
	}() == true {
		__antithesis_instrumentation__.Notify(250788)

		return
	} else {
		__antithesis_instrumentation__.Notify(250789)
	}
	__antithesis_instrumentation__.Notify(250784)

	privBits := privList.ToBitField()
	grantBits := grantOptionList.ToBitField()

	if !privilege.ALL.IsSetIn(privBits) {
		__antithesis_instrumentation__.Notify(250790)
		for _, grantOption := range grantOptionList {
			__antithesis_instrumentation__.Notify(250791)
			if privBits&grantOption.Mask() == 0 {
				__antithesis_instrumentation__.Notify(250792)
				return
			} else {
				__antithesis_instrumentation__.Notify(250793)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(250794)
	}
	__antithesis_instrumentation__.Notify(250785)

	if privilege.ALL.IsSetIn(privBits) {
		__antithesis_instrumentation__.Notify(250795)
		userPriv.Privileges = privilege.ALL.Mask()
	} else {
		__antithesis_instrumentation__.Notify(250796)
		if !privilege.ALL.IsSetIn(userPriv.Privileges) {
			__antithesis_instrumentation__.Notify(250797)
			userPriv.Privileges |= privBits
		} else {
			__antithesis_instrumentation__.Notify(250798)
		}
	}
	__antithesis_instrumentation__.Notify(250786)

	if privilege.ALL.IsSetIn(grantBits) {
		__antithesis_instrumentation__.Notify(250799)
		userPriv.WithGrantOption = privilege.ALL.Mask()
	} else {
		__antithesis_instrumentation__.Notify(250800)
		if !privilege.ALL.IsSetIn(userPriv.WithGrantOption) {
			__antithesis_instrumentation__.Notify(250801)
			userPriv.WithGrantOption |= grantBits
		} else {
			__antithesis_instrumentation__.Notify(250802)
		}
	}
}

func setRoleHasAllOnTargetObject(
	defaultPrivilegesForRole *catpb.DefaultPrivilegesForRole,
	roleHasAll bool,
	targetObject tree.AlterDefaultPrivilegesTargetObject,
) {
	__antithesis_instrumentation__.Notify(250803)
	if !defaultPrivilegesForRole.IsExplicitRole() {
		__antithesis_instrumentation__.Notify(250805)

		panic("DefaultPrivilegesForRole must be for an explicit role")
	} else {
		__antithesis_instrumentation__.Notify(250806)
	}
	__antithesis_instrumentation__.Notify(250804)
	switch targetObject {
	case tree.Tables:
		__antithesis_instrumentation__.Notify(250807)
		defaultPrivilegesForRole.GetExplicitRole().RoleHasAllPrivilegesOnTables = roleHasAll
	case tree.Sequences:
		__antithesis_instrumentation__.Notify(250808)
		defaultPrivilegesForRole.GetExplicitRole().RoleHasAllPrivilegesOnSequences = roleHasAll
	case tree.Types:
		__antithesis_instrumentation__.Notify(250809)
		defaultPrivilegesForRole.GetExplicitRole().RoleHasAllPrivilegesOnTypes = roleHasAll
	case tree.Schemas:
		__antithesis_instrumentation__.Notify(250810)
		defaultPrivilegesForRole.GetExplicitRole().RoleHasAllPrivilegesOnSchemas = roleHasAll
	default:
		__antithesis_instrumentation__.Notify(250811)
		panic(fmt.Sprintf("unknown target object %s", targetObject))
	}
}
