package catpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/errors"
)

type PrivilegeDescVersion uint32

const (
	InitialVersion PrivilegeDescVersion = iota

	OwnerVersion

	Version21_2
)

func (p PrivilegeDescriptor) Owner() security.SQLUsername {
	__antithesis_instrumentation__.Notify(249385)
	return p.OwnerProto.Decode()
}

func (u UserPrivileges) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(249386)
	return u.UserProto.Decode()
}

func (p PrivilegeDescriptor) findUserIndex(user security.SQLUsername) int {
	__antithesis_instrumentation__.Notify(249387)
	idx := sort.Search(len(p.Users), func(i int) bool {
		__antithesis_instrumentation__.Notify(249390)
		return !p.Users[i].User().LessThan(user)
	})
	__antithesis_instrumentation__.Notify(249388)
	if idx < len(p.Users) && func() bool {
		__antithesis_instrumentation__.Notify(249391)
		return p.Users[idx].User() == user == true
	}() == true {
		__antithesis_instrumentation__.Notify(249392)
		return idx
	} else {
		__antithesis_instrumentation__.Notify(249393)
	}
	__antithesis_instrumentation__.Notify(249389)
	return -1
}

func (p PrivilegeDescriptor) FindUser(user security.SQLUsername) (*UserPrivileges, bool) {
	__antithesis_instrumentation__.Notify(249394)
	idx := p.findUserIndex(user)
	if idx == -1 {
		__antithesis_instrumentation__.Notify(249396)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(249397)
	}
	__antithesis_instrumentation__.Notify(249395)
	return &p.Users[idx], true
}

func (p *PrivilegeDescriptor) FindOrCreateUser(user security.SQLUsername) *UserPrivileges {
	__antithesis_instrumentation__.Notify(249398)
	idx := sort.Search(len(p.Users), func(i int) bool {
		__antithesis_instrumentation__.Notify(249401)
		return !p.Users[i].User().LessThan(user)
	})
	__antithesis_instrumentation__.Notify(249399)
	if idx == len(p.Users) {
		__antithesis_instrumentation__.Notify(249402)

		p.Users = append(p.Users, UserPrivileges{UserProto: user.EncodeProto()})
	} else {
		__antithesis_instrumentation__.Notify(249403)
		if p.Users[idx].User() == user {
			__antithesis_instrumentation__.Notify(249404)

		} else {
			__antithesis_instrumentation__.Notify(249405)

			p.Users = append(p.Users, UserPrivileges{})
			copy(p.Users[idx+1:], p.Users[idx:])
			p.Users[idx] = UserPrivileges{UserProto: user.EncodeProto()}
		}
	}
	__antithesis_instrumentation__.Notify(249400)
	return &p.Users[idx]
}

func (p *PrivilegeDescriptor) RemoveUser(user security.SQLUsername) {
	__antithesis_instrumentation__.Notify(249406)
	idx := p.findUserIndex(user)
	if idx == -1 {
		__antithesis_instrumentation__.Notify(249408)

		return
	} else {
		__antithesis_instrumentation__.Notify(249409)
	}
	__antithesis_instrumentation__.Notify(249407)
	p.Users = append(p.Users[:idx], p.Users[idx+1:]...)
}

func NewCustomSuperuserPrivilegeDescriptor(
	priv privilege.List, owner security.SQLUsername,
) *PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(249410)
	return &PrivilegeDescriptor{
		OwnerProto: owner.EncodeProto(),
		Users: []UserPrivileges{
			{
				UserProto:       security.AdminRoleName().EncodeProto(),
				Privileges:      priv.ToBitField(),
				WithGrantOption: priv.ToBitField(),
			},
			{
				UserProto:       security.RootUserName().EncodeProto(),
				Privileges:      priv.ToBitField(),
				WithGrantOption: priv.ToBitField(),
			},
		},
		Version: Version21_2,
	}
}

func NewVirtualTablePrivilegeDescriptor() *PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(249411)
	return NewPrivilegeDescriptor(
		security.PublicRoleName(), privilege.List{privilege.SELECT}, privilege.List{}, security.NodeUserName(),
	)
}

func NewVirtualSchemaPrivilegeDescriptor() *PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(249412)
	return NewPrivilegeDescriptor(
		security.PublicRoleName(), privilege.List{privilege.USAGE}, privilege.List{}, security.NodeUserName(),
	)
}

func NewTemporarySchemaPrivilegeDescriptor() *PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(249413)
	p := NewBasePrivilegeDescriptor(security.AdminRoleName())
	p.Grant(security.PublicRoleName(), privilege.List{privilege.CREATE, privilege.USAGE}, false)
	return p
}

func NewPrivilegeDescriptor(
	user security.SQLUsername,
	priv privilege.List,
	grantOption privilege.List,
	owner security.SQLUsername,
) *PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(249414)
	return &PrivilegeDescriptor{
		OwnerProto: owner.EncodeProto(),
		Users: []UserPrivileges{
			{
				UserProto:       user.EncodeProto(),
				Privileges:      priv.ToBitField(),
				WithGrantOption: grantOption.ToBitField(),
			},
		},
		Version: Version21_2,
	}
}

var DefaultSuperuserPrivileges = privilege.List{privilege.ALL}

func NewBasePrivilegeDescriptor(owner security.SQLUsername) *PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(249415)
	return NewCustomSuperuserPrivilegeDescriptor(DefaultSuperuserPrivileges, owner)
}

func NewBaseDatabasePrivilegeDescriptor(owner security.SQLUsername) *PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(249416)
	p := NewBasePrivilegeDescriptor(owner)
	p.Grant(security.PublicRoleName(), privilege.List{privilege.CONNECT}, false)
	return p
}

func NewPublicSchemaPrivilegeDescriptor() *PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(249417)

	p := NewBasePrivilegeDescriptor(security.AdminRoleName())

	p.Grant(security.PublicRoleName(), privilege.List{privilege.CREATE, privilege.USAGE}, false)
	return p
}

func (p *PrivilegeDescriptor) CheckGrantOptions(
	user security.SQLUsername, privList privilege.List,
) bool {
	__antithesis_instrumentation__.Notify(249418)
	userPriv, exists := p.FindUser(user)
	if !exists {
		__antithesis_instrumentation__.Notify(249422)
		return false
	} else {
		__antithesis_instrumentation__.Notify(249423)
	}
	__antithesis_instrumentation__.Notify(249419)

	if privilege.ALL.IsSetIn(userPriv.WithGrantOption) {
		__antithesis_instrumentation__.Notify(249424)
		return true
	} else {
		__antithesis_instrumentation__.Notify(249425)
	}
	__antithesis_instrumentation__.Notify(249420)

	for _, priv := range privList {
		__antithesis_instrumentation__.Notify(249426)
		if userPriv.WithGrantOption&priv.Mask() == 0 {
			__antithesis_instrumentation__.Notify(249427)
			return false
		} else {
			__antithesis_instrumentation__.Notify(249428)
		}
	}
	__antithesis_instrumentation__.Notify(249421)

	return true
}

func (p *PrivilegeDescriptor) Grant(
	user security.SQLUsername, privList privilege.List, withGrantOption bool,
) {
	__antithesis_instrumentation__.Notify(249429)
	userPriv := p.FindOrCreateUser(user)
	if privilege.ALL.IsSetIn(userPriv.WithGrantOption) && func() bool {
		__antithesis_instrumentation__.Notify(249434)
		return privilege.ALL.IsSetIn(userPriv.Privileges) == true
	}() == true {
		__antithesis_instrumentation__.Notify(249435)

		return
	} else {
		__antithesis_instrumentation__.Notify(249436)
	}
	__antithesis_instrumentation__.Notify(249430)

	if privilege.ALL.IsSetIn(userPriv.Privileges) && func() bool {
		__antithesis_instrumentation__.Notify(249437)
		return !withGrantOption == true
	}() == true {
		__antithesis_instrumentation__.Notify(249438)

		return
	} else {
		__antithesis_instrumentation__.Notify(249439)
	}
	__antithesis_instrumentation__.Notify(249431)

	bits := privList.ToBitField()
	if privilege.ALL.IsSetIn(bits) {
		__antithesis_instrumentation__.Notify(249440)

		userPriv.Privileges = privilege.ALL.Mask()
		if withGrantOption {
			__antithesis_instrumentation__.Notify(249442)
			userPriv.WithGrantOption = privilege.ALL.Mask()
		} else {
			__antithesis_instrumentation__.Notify(249443)
		}
		__antithesis_instrumentation__.Notify(249441)
		return
	} else {
		__antithesis_instrumentation__.Notify(249444)
	}
	__antithesis_instrumentation__.Notify(249432)

	if withGrantOption {
		__antithesis_instrumentation__.Notify(249445)
		userPriv.WithGrantOption |= bits
	} else {
		__antithesis_instrumentation__.Notify(249446)
	}
	__antithesis_instrumentation__.Notify(249433)
	userPriv.Privileges |= bits
}

func (p *PrivilegeDescriptor) Revoke(
	user security.SQLUsername,
	privList privilege.List,
	objectType privilege.ObjectType,
	grantOptionFor bool,
) {
	__antithesis_instrumentation__.Notify(249447)
	userPriv, ok := p.FindUser(user)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(249452)
		return userPriv.Privileges == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(249453)

		return
	} else {
		__antithesis_instrumentation__.Notify(249454)
	}
	__antithesis_instrumentation__.Notify(249448)

	bits := privList.ToBitField()
	if privilege.ALL.IsSetIn(bits) {
		__antithesis_instrumentation__.Notify(249455)
		userPriv.WithGrantOption = 0
		if !grantOptionFor {
			__antithesis_instrumentation__.Notify(249457)

			p.RemoveUser(user)
		} else {
			__antithesis_instrumentation__.Notify(249458)
			if privilege.ALL.IsSetIn(userPriv.Privileges) {
				__antithesis_instrumentation__.Notify(249459)

				userPriv.Privileges = privilege.ALL.Mask()
			} else {
				__antithesis_instrumentation__.Notify(249460)
			}
		}
		__antithesis_instrumentation__.Notify(249456)
		return
	} else {
		__antithesis_instrumentation__.Notify(249461)
	}
	__antithesis_instrumentation__.Notify(249449)

	if privilege.ALL.IsSetIn(userPriv.Privileges) && func() bool {
		__antithesis_instrumentation__.Notify(249462)
		return !grantOptionFor == true
	}() == true {
		__antithesis_instrumentation__.Notify(249463)

		validPrivs := privilege.GetValidPrivilegesForObject(objectType)
		userPriv.Privileges = 0
		for _, v := range validPrivs {
			__antithesis_instrumentation__.Notify(249464)
			if v != privilege.ALL {
				__antithesis_instrumentation__.Notify(249465)
				userPriv.Privileges |= v.Mask()
			} else {
				__antithesis_instrumentation__.Notify(249466)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(249467)
	}
	__antithesis_instrumentation__.Notify(249450)

	if privilege.ALL.IsSetIn(userPriv.WithGrantOption) {
		__antithesis_instrumentation__.Notify(249468)

		validPrivs := privilege.GetValidPrivilegesForObject(objectType)
		userPriv.WithGrantOption = 0
		for _, v := range validPrivs {
			__antithesis_instrumentation__.Notify(249469)
			if v != privilege.ALL {
				__antithesis_instrumentation__.Notify(249470)
				userPriv.WithGrantOption |= v.Mask()
			} else {
				__antithesis_instrumentation__.Notify(249471)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(249472)
	}
	__antithesis_instrumentation__.Notify(249451)

	userPriv.WithGrantOption &^= bits
	if !grantOptionFor {
		__antithesis_instrumentation__.Notify(249473)
		userPriv.Privileges &^= bits

		if userPriv.Privileges == 0 {
			__antithesis_instrumentation__.Notify(249474)
			p.RemoveUser(user)
		} else {
			__antithesis_instrumentation__.Notify(249475)
		}
	} else {
		__antithesis_instrumentation__.Notify(249476)
	}
}

func (p *PrivilegeDescriptor) GrantPrivilegeToGrantOptions(
	user security.SQLUsername, isGrant bool,
) {
	__antithesis_instrumentation__.Notify(249477)
	if isGrant {
		__antithesis_instrumentation__.Notify(249478)
		userPriv := p.FindOrCreateUser(user)
		userPriv.WithGrantOption = userPriv.Privileges
	} else {
		__antithesis_instrumentation__.Notify(249479)
		userPriv, ok := p.FindUser(user)
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(249481)
			return userPriv.Privileges == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(249482)

			return
		} else {
			__antithesis_instrumentation__.Notify(249483)
		}
		__antithesis_instrumentation__.Notify(249480)
		userPriv.WithGrantOption = 0
	}
}

func (p PrivilegeDescriptor) ValidateSuperuserPrivileges(
	parentID catid.DescID,
	objectType privilege.ObjectType,
	objectName string,
	allowedSuperuserPrivileges privilege.List,
) error {
	__antithesis_instrumentation__.Notify(249484)
	if parentID == catid.InvalidDescID && func() bool {
		__antithesis_instrumentation__.Notify(249487)
		return objectType != privilege.Database == true
	}() == true {
		__antithesis_instrumentation__.Notify(249488)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(249489)
	}
	__antithesis_instrumentation__.Notify(249485)
	for _, user := range []security.SQLUsername{

		security.RootUserName(),

		security.AdminRoleName(),
	} {
		__antithesis_instrumentation__.Notify(249490)
		superPriv, ok := p.FindUser(user)
		if !ok {
			__antithesis_instrumentation__.Notify(249492)
			return fmt.Errorf(
				"user %s does not have privileges over %s",
				user,
				privilegeObject(parentID, objectType, objectName),
			)
		} else {
			__antithesis_instrumentation__.Notify(249493)
		}
		__antithesis_instrumentation__.Notify(249491)

		if superPriv.Privileges != allowedSuperuserPrivileges.ToBitField() {
			__antithesis_instrumentation__.Notify(249494)
			return fmt.Errorf(
				"user %s must have exactly %s privileges on %s",
				user,
				allowedSuperuserPrivileges,
				privilegeObject(parentID, objectType, objectName),
			)
		} else {
			__antithesis_instrumentation__.Notify(249495)
		}
	}
	__antithesis_instrumentation__.Notify(249486)
	return nil
}

func (p PrivilegeDescriptor) Validate(
	parentID catid.DescID,
	objectType privilege.ObjectType,
	objectName string,
	allowedSuperuserPrivileges privilege.List,
) error {
	__antithesis_instrumentation__.Notify(249496)
	if err := p.ValidateSuperuserPrivileges(parentID, objectType, objectName, allowedSuperuserPrivileges); err != nil {
		__antithesis_instrumentation__.Notify(249500)
		return errors.HandleAsAssertionFailure(err)
	} else {
		__antithesis_instrumentation__.Notify(249501)
	}
	__antithesis_instrumentation__.Notify(249497)

	if p.Version >= OwnerVersion {
		__antithesis_instrumentation__.Notify(249502)
		if p.Owner().Undefined() {
			__antithesis_instrumentation__.Notify(249503)
			return errors.AssertionFailedf("found no owner for %s", privilegeObject(parentID, objectType, objectName))
		} else {
			__antithesis_instrumentation__.Notify(249504)
		}
	} else {
		__antithesis_instrumentation__.Notify(249505)
	}
	__antithesis_instrumentation__.Notify(249498)

	valid, u, remaining := p.IsValidPrivilegesForObjectType(objectType)
	if !valid {
		__antithesis_instrumentation__.Notify(249506)
		return errors.AssertionFailedf(
			"user %s must not have %s privileges on %s",
			u.User(),
			privilege.ListFromBitField(remaining, privilege.Any),
			privilegeObject(parentID, objectType, objectName),
		)
	} else {
		__antithesis_instrumentation__.Notify(249507)
	}
	__antithesis_instrumentation__.Notify(249499)

	return nil
}

func (p PrivilegeDescriptor) IsValidPrivilegesForObjectType(
	objectType privilege.ObjectType,
) (bool, UserPrivileges, uint32) {
	__antithesis_instrumentation__.Notify(249508)
	allowedPrivilegesBits := privilege.GetValidPrivilegesForObject(objectType).ToBitField()

	if p.Version < Version21_2 {
		__antithesis_instrumentation__.Notify(249511)
		if objectType == privilege.Schema {
			__antithesis_instrumentation__.Notify(249513)

			allowedPrivilegesBits |= privilege.GetValidPrivilegesForObject(privilege.Database).ToBitField()
		} else {
			__antithesis_instrumentation__.Notify(249514)
		}
		__antithesis_instrumentation__.Notify(249512)
		if objectType == privilege.Table || func() bool {
			__antithesis_instrumentation__.Notify(249515)
			return objectType == privilege.Database == true
		}() == true {
			__antithesis_instrumentation__.Notify(249516)

			allowedPrivilegesBits |= privilege.USAGE.Mask()
		} else {
			__antithesis_instrumentation__.Notify(249517)
		}
	} else {
		__antithesis_instrumentation__.Notify(249518)
	}
	__antithesis_instrumentation__.Notify(249509)

	for _, u := range p.Users {
		__antithesis_instrumentation__.Notify(249519)
		if u.User().IsRootUser() || func() bool {
			__antithesis_instrumentation__.Notify(249521)
			return u.User().IsAdminRole() == true
		}() == true {
			__antithesis_instrumentation__.Notify(249522)

			continue
		} else {
			__antithesis_instrumentation__.Notify(249523)
		}
		__antithesis_instrumentation__.Notify(249520)

		if remaining := u.Privileges &^ allowedPrivilegesBits; remaining != 0 {
			__antithesis_instrumentation__.Notify(249524)
			return false, u, remaining
		} else {
			__antithesis_instrumentation__.Notify(249525)
		}
	}
	__antithesis_instrumentation__.Notify(249510)

	return true, UserPrivileges{}, 0
}

type UserPrivilege struct {
	User       security.SQLUsername
	Privileges []privilege.Privilege
}

func (p PrivilegeDescriptor) Show(objectType privilege.ObjectType) []UserPrivilege {
	__antithesis_instrumentation__.Notify(249526)
	ret := make([]UserPrivilege, 0, len(p.Users))
	for _, userPriv := range p.Users {
		__antithesis_instrumentation__.Notify(249528)
		privileges := privilege.PrivilegesFromBitFields(userPriv.Privileges, userPriv.WithGrantOption, objectType)
		sort.Slice(privileges, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(249530)
			return strings.Compare(privileges[i].Kind.String(), privileges[j].Kind.String()) < 0
		})
		__antithesis_instrumentation__.Notify(249529)
		ret = append(ret, UserPrivilege{
			User:       userPriv.User(),
			Privileges: privileges,
		})
	}
	__antithesis_instrumentation__.Notify(249527)
	return ret
}

func (p PrivilegeDescriptor) CheckPrivilege(user security.SQLUsername, priv privilege.Kind) bool {
	__antithesis_instrumentation__.Notify(249531)
	userPriv, ok := p.FindUser(user)
	if !ok {
		__antithesis_instrumentation__.Notify(249534)

		return user.IsNodeUser()
	} else {
		__antithesis_instrumentation__.Notify(249535)
	}
	__antithesis_instrumentation__.Notify(249532)

	if privilege.ALL.IsSetIn(userPriv.Privileges) {
		__antithesis_instrumentation__.Notify(249536)
		return true
	} else {
		__antithesis_instrumentation__.Notify(249537)
	}
	__antithesis_instrumentation__.Notify(249533)
	return priv.IsSetIn(userPriv.Privileges)
}

func (p PrivilegeDescriptor) AnyPrivilege(user security.SQLUsername) bool {
	__antithesis_instrumentation__.Notify(249538)
	if p.Owner() == user {
		__antithesis_instrumentation__.Notify(249541)
		return true
	} else {
		__antithesis_instrumentation__.Notify(249542)
	}
	__antithesis_instrumentation__.Notify(249539)
	userPriv, ok := p.FindUser(user)
	if !ok {
		__antithesis_instrumentation__.Notify(249543)
		return false
	} else {
		__antithesis_instrumentation__.Notify(249544)
	}
	__antithesis_instrumentation__.Notify(249540)
	return userPriv.Privileges != 0
}

func (p PrivilegeDescriptor) HasAllPrivileges(
	user security.SQLUsername, objectType privilege.ObjectType,
) bool {
	__antithesis_instrumentation__.Notify(249545)
	if p.CheckPrivilege(user, privilege.ALL) {
		__antithesis_instrumentation__.Notify(249548)
		return true
	} else {
		__antithesis_instrumentation__.Notify(249549)
	}
	__antithesis_instrumentation__.Notify(249546)

	validPrivileges := privilege.GetValidPrivilegesForObject(objectType)
	for _, priv := range validPrivileges {
		__antithesis_instrumentation__.Notify(249550)
		if priv == privilege.ALL {
			__antithesis_instrumentation__.Notify(249552)
			continue
		} else {
			__antithesis_instrumentation__.Notify(249553)
		}
		__antithesis_instrumentation__.Notify(249551)
		if !p.CheckPrivilege(user, priv) {
			__antithesis_instrumentation__.Notify(249554)
			return false
		} else {
			__antithesis_instrumentation__.Notify(249555)
		}
	}
	__antithesis_instrumentation__.Notify(249547)

	return true
}

func (p *PrivilegeDescriptor) SetOwner(owner security.SQLUsername) {
	__antithesis_instrumentation__.Notify(249556)
	p.OwnerProto = owner.EncodeProto()
}

func (p *PrivilegeDescriptor) SetVersion(version PrivilegeDescVersion) {
	__antithesis_instrumentation__.Notify(249557)
	p.Version = version
}

func privilegeObject(
	parentID catid.DescID, objectType privilege.ObjectType, objectName string,
) string {
	__antithesis_instrumentation__.Notify(249558)
	if parentID == keys.SystemDatabaseID || func() bool {
		__antithesis_instrumentation__.Notify(249560)
		return (parentID == catid.InvalidDescID && func() bool {
			__antithesis_instrumentation__.Notify(249561)
			return objectName == catconstants.SystemDatabaseName == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(249562)
		return fmt.Sprintf("system %s %q", objectType, objectName)
	} else {
		__antithesis_instrumentation__.Notify(249563)
	}
	__antithesis_instrumentation__.Notify(249559)
	return fmt.Sprintf("%s %q", objectType, objectName)
}
