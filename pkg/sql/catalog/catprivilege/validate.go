package catprivilege

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
)

func Validate(
	p catpb.PrivilegeDescriptor, objectNameKey catalog.NameKey, objectType privilege.ObjectType,
) error {
	__antithesis_instrumentation__.Notify(250884)
	return p.Validate(
		objectNameKey.GetParentID(),
		objectType,
		objectNameKey.GetName(),
		allowedSuperuserPrivileges(objectNameKey),
	)
}

func ValidateSuperuserPrivileges(
	p catpb.PrivilegeDescriptor, objectNameKey catalog.NameKey, objectType privilege.ObjectType,
) error {
	__antithesis_instrumentation__.Notify(250885)
	return p.ValidateSuperuserPrivileges(
		objectNameKey.GetParentID(),
		objectType,
		objectNameKey.GetName(),
		allowedSuperuserPrivileges(objectNameKey),
	)
}

func ValidateDefaultPrivileges(p catpb.DefaultPrivilegeDescriptor) error {
	__antithesis_instrumentation__.Notify(250886)
	return p.Validate()
}

func allowedSuperuserPrivileges(objectNameKey catalog.NameKey) privilege.List {
	__antithesis_instrumentation__.Notify(250887)
	privs := SystemSuperuserPrivileges(objectNameKey)
	if privs != nil {
		__antithesis_instrumentation__.Notify(250889)
		return privs
	} else {
		__antithesis_instrumentation__.Notify(250890)
	}
	__antithesis_instrumentation__.Notify(250888)
	return catpb.DefaultSuperuserPrivileges
}
