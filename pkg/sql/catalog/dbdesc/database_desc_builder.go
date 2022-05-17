package dbdesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

type DatabaseDescriptorBuilder interface {
	catalog.DescriptorBuilder
	BuildImmutableDatabase() catalog.DatabaseDescriptor
	BuildExistingMutableDatabase() *Mutable
	BuildCreatedMutableDatabase() *Mutable
}

type databaseDescriptorBuilder struct {
	original      *descpb.DatabaseDescriptor
	maybeModified *descpb.DatabaseDescriptor

	isUncommittedVersion bool
	changes              catalog.PostDeserializationChanges
}

var _ DatabaseDescriptorBuilder = &databaseDescriptorBuilder{}

func NewBuilder(desc *descpb.DatabaseDescriptor) DatabaseDescriptorBuilder {
	__antithesis_instrumentation__.Notify(251405)
	return newBuilder(desc, false,
		catalog.PostDeserializationChanges{})
}

func newBuilder(
	desc *descpb.DatabaseDescriptor,
	isUncommittedVersion bool,
	changes catalog.PostDeserializationChanges,
) DatabaseDescriptorBuilder {
	__antithesis_instrumentation__.Notify(251406)
	return &databaseDescriptorBuilder{
		original:             protoutil.Clone(desc).(*descpb.DatabaseDescriptor),
		isUncommittedVersion: isUncommittedVersion,
		changes:              changes,
	}
}

func (ddb *databaseDescriptorBuilder) DescriptorType() catalog.DescriptorType {
	__antithesis_instrumentation__.Notify(251407)
	return catalog.Database
}

func (ddb *databaseDescriptorBuilder) RunPostDeserializationChanges() error {
	__antithesis_instrumentation__.Notify(251408)
	ddb.maybeModified = protoutil.Clone(ddb.original).(*descpb.DatabaseDescriptor)

	createdDefaultPrivileges := false
	removedIncompatibleDatabasePrivs := false

	if ddb.original.GetID() != keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(251412)
		if ddb.maybeModified.DefaultPrivileges == nil {
			__antithesis_instrumentation__.Notify(251414)
			ddb.maybeModified.DefaultPrivileges = catprivilege.MakeDefaultPrivilegeDescriptor(
				catpb.DefaultPrivilegeDescriptor_DATABASE)
			createdDefaultPrivileges = true
		} else {
			__antithesis_instrumentation__.Notify(251415)
		}
		__antithesis_instrumentation__.Notify(251413)

		removedIncompatibleDatabasePrivs = maybeConvertIncompatibleDBPrivilegesToDefaultPrivileges(
			ddb.maybeModified.Privileges, ddb.maybeModified.DefaultPrivileges,
		)
	} else {
		__antithesis_instrumentation__.Notify(251416)
	}
	__antithesis_instrumentation__.Notify(251409)

	privsChanged := catprivilege.MaybeFixPrivileges(
		&ddb.maybeModified.Privileges,
		descpb.InvalidID,
		descpb.InvalidID,
		privilege.Database,
		ddb.maybeModified.GetName())
	addedGrantOptions := catprivilege.MaybeUpdateGrantOptions(ddb.maybeModified.Privileges)

	if privsChanged || func() bool {
		__antithesis_instrumentation__.Notify(251417)
		return addedGrantOptions == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(251418)
		return removedIncompatibleDatabasePrivs == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(251419)
		return createdDefaultPrivileges == true
	}() == true {
		__antithesis_instrumentation__.Notify(251420)
		ddb.changes.Add(catalog.UpgradedPrivileges)
	} else {
		__antithesis_instrumentation__.Notify(251421)
	}
	__antithesis_instrumentation__.Notify(251410)
	if maybeRemoveDroppedSelfEntryFromSchemas(ddb.maybeModified) {
		__antithesis_instrumentation__.Notify(251422)
		ddb.changes.Add(catalog.RemovedSelfEntryInSchemas)
	} else {
		__antithesis_instrumentation__.Notify(251423)
	}
	__antithesis_instrumentation__.Notify(251411)
	return nil
}

func (ddb *databaseDescriptorBuilder) RunRestoreChanges(
	_ func(id descpb.ID) catalog.Descriptor,
) error {
	__antithesis_instrumentation__.Notify(251424)
	return nil
}

func maybeConvertIncompatibleDBPrivilegesToDefaultPrivileges(
	privileges *catpb.PrivilegeDescriptor, defaultPrivileges *catpb.DefaultPrivilegeDescriptor,
) (hasChanged bool) {
	__antithesis_instrumentation__.Notify(251425)

	if privileges == nil {
		__antithesis_instrumentation__.Notify(251428)
		return false
	} else {
		__antithesis_instrumentation__.Notify(251429)
	}
	__antithesis_instrumentation__.Notify(251426)

	var pgIncompatibleDBPrivileges = privilege.List{
		privilege.SELECT, privilege.INSERT, privilege.UPDATE, privilege.DELETE,
	}

	for i, user := range privileges.Users {
		__antithesis_instrumentation__.Notify(251430)
		incompatiblePrivileges := user.Privileges & pgIncompatibleDBPrivileges.ToBitField()

		if incompatiblePrivileges == 0 {
			__antithesis_instrumentation__.Notify(251432)
			continue
		} else {
			__antithesis_instrumentation__.Notify(251433)
		}
		__antithesis_instrumentation__.Notify(251431)

		hasChanged = true

		user.Privileges ^= incompatiblePrivileges

		privileges.Users[i] = user

		role := defaultPrivileges.FindOrCreateUser(catpb.DefaultPrivilegesRole{ForAllRoles: true})
		tableDefaultPrivilegesForAllRoles := role.DefaultPrivilegesPerObject[tree.Tables]

		defaultPrivilegesForUser := tableDefaultPrivilegesForAllRoles.FindOrCreateUser(user.User())
		defaultPrivilegesForUser.Privileges |= incompatiblePrivileges

		role.DefaultPrivilegesPerObject[tree.Tables] = tableDefaultPrivilegesForAllRoles
	}
	__antithesis_instrumentation__.Notify(251427)

	return hasChanged
}

func (ddb *databaseDescriptorBuilder) BuildImmutable() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(251434)
	return ddb.BuildImmutableDatabase()
}

func (ddb *databaseDescriptorBuilder) BuildImmutableDatabase() catalog.DatabaseDescriptor {
	__antithesis_instrumentation__.Notify(251435)
	desc := ddb.maybeModified
	if desc == nil {
		__antithesis_instrumentation__.Notify(251437)
		desc = ddb.original
	} else {
		__antithesis_instrumentation__.Notify(251438)
	}
	__antithesis_instrumentation__.Notify(251436)
	return &immutable{
		DatabaseDescriptor:   *desc,
		isUncommittedVersion: ddb.isUncommittedVersion,
		changes:              ddb.changes,
	}
}

func (ddb *databaseDescriptorBuilder) BuildExistingMutable() catalog.MutableDescriptor {
	__antithesis_instrumentation__.Notify(251439)
	return ddb.BuildExistingMutableDatabase()
}

func (ddb *databaseDescriptorBuilder) BuildExistingMutableDatabase() *Mutable {
	__antithesis_instrumentation__.Notify(251440)
	if ddb.maybeModified == nil {
		__antithesis_instrumentation__.Notify(251442)
		ddb.maybeModified = protoutil.Clone(ddb.original).(*descpb.DatabaseDescriptor)
	} else {
		__antithesis_instrumentation__.Notify(251443)
	}
	__antithesis_instrumentation__.Notify(251441)
	return &Mutable{
		immutable: immutable{
			DatabaseDescriptor:   *ddb.maybeModified,
			changes:              ddb.changes,
			isUncommittedVersion: ddb.isUncommittedVersion,
		},
		ClusterVersion: &immutable{DatabaseDescriptor: *ddb.original},
	}
}

func (ddb *databaseDescriptorBuilder) BuildCreatedMutable() catalog.MutableDescriptor {
	__antithesis_instrumentation__.Notify(251444)
	return ddb.BuildCreatedMutableDatabase()
}

func (ddb *databaseDescriptorBuilder) BuildCreatedMutableDatabase() *Mutable {
	__antithesis_instrumentation__.Notify(251445)
	desc := ddb.maybeModified
	if desc == nil {
		__antithesis_instrumentation__.Notify(251447)
		desc = ddb.original
	} else {
		__antithesis_instrumentation__.Notify(251448)
	}
	__antithesis_instrumentation__.Notify(251446)
	return &Mutable{
		immutable: immutable{
			DatabaseDescriptor:   *desc,
			changes:              ddb.changes,
			isUncommittedVersion: ddb.isUncommittedVersion,
		},
	}
}

type NewInitialOption func(*descpb.DatabaseDescriptor)

func MaybeWithDatabaseRegionConfig(regionConfig *multiregion.RegionConfig) NewInitialOption {
	__antithesis_instrumentation__.Notify(251449)
	return func(desc *descpb.DatabaseDescriptor) {
		__antithesis_instrumentation__.Notify(251450)

		if regionConfig == nil {
			__antithesis_instrumentation__.Notify(251452)
			return
		} else {
			__antithesis_instrumentation__.Notify(251453)
		}
		__antithesis_instrumentation__.Notify(251451)
		desc.RegionConfig = &descpb.DatabaseDescriptor_RegionConfig{
			SurvivalGoal:  regionConfig.SurvivalGoal(),
			PrimaryRegion: regionConfig.PrimaryRegion(),
			RegionEnumID:  regionConfig.RegionEnumID(),
			Placement:     regionConfig.Placement(),
		}
	}
}

func WithPublicSchemaID(publicSchemaID descpb.ID) NewInitialOption {
	__antithesis_instrumentation__.Notify(251454)
	return func(desc *descpb.DatabaseDescriptor) {
		__antithesis_instrumentation__.Notify(251455)

		if publicSchemaID != keys.PublicSchemaID {
			__antithesis_instrumentation__.Notify(251456)
			desc.Schemas = map[string]descpb.DatabaseDescriptor_SchemaInfo{
				tree.PublicSchema: {
					ID:      publicSchemaID,
					Dropped: false,
				},
			}
		} else {
			__antithesis_instrumentation__.Notify(251457)
		}
	}
}

func NewInitial(
	id descpb.ID, name string, owner security.SQLUsername, options ...NewInitialOption,
) *Mutable {
	__antithesis_instrumentation__.Notify(251458)
	return newInitialWithPrivileges(
		id,
		name,
		catpb.NewBaseDatabasePrivilegeDescriptor(owner),
		catprivilege.MakeDefaultPrivilegeDescriptor(catpb.DefaultPrivilegeDescriptor_DATABASE),
		options...,
	)
}

func newInitialWithPrivileges(
	id descpb.ID,
	name string,
	privileges *catpb.PrivilegeDescriptor,
	defaultPrivileges *catpb.DefaultPrivilegeDescriptor,
	options ...NewInitialOption,
) *Mutable {
	__antithesis_instrumentation__.Notify(251459)
	ret := descpb.DatabaseDescriptor{
		Name:              name,
		ID:                id,
		Version:           1,
		Privileges:        privileges,
		DefaultPrivileges: defaultPrivileges,
	}
	for _, option := range options {
		__antithesis_instrumentation__.Notify(251461)
		option(&ret)
	}
	__antithesis_instrumentation__.Notify(251460)
	return NewBuilder(&ret).BuildCreatedMutableDatabase()
}
