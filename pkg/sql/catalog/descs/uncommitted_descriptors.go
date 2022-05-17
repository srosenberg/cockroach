package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/errors"
)

type uncommittedDescriptorStatus int

const (
	notValidatedYet uncommittedDescriptorStatus = iota

	notCheckedOutYet

	checkedOutAtLeastOnce
)

type uncommittedDescriptor struct {
	immutable catalog.Descriptor

	mutable catalog.MutableDescriptor

	uncommittedDescriptorStatus
}

func (u *uncommittedDescriptor) GetName() string {
	__antithesis_instrumentation__.Notify(265120)
	return u.immutable.GetName()
}

func (u *uncommittedDescriptor) GetParentID() descpb.ID {
	__antithesis_instrumentation__.Notify(265121)
	return u.immutable.GetParentID()
}

func (u uncommittedDescriptor) GetParentSchemaID() descpb.ID {
	__antithesis_instrumentation__.Notify(265122)
	return u.immutable.GetParentSchemaID()
}

func (u uncommittedDescriptor) GetID() descpb.ID {
	__antithesis_instrumentation__.Notify(265123)
	return u.immutable.GetID()
}

func (u *uncommittedDescriptor) checkOut() catalog.MutableDescriptor {
	__antithesis_instrumentation__.Notify(265124)
	if u.mutable == nil {
		__antithesis_instrumentation__.Notify(265126)

		return u.immutable.NewBuilder().BuildExistingMutable()
	} else {
		__antithesis_instrumentation__.Notify(265127)
	}
	__antithesis_instrumentation__.Notify(265125)
	u.uncommittedDescriptorStatus = checkedOutAtLeastOnce
	return u.mutable
}

var _ catalog.NameEntry = (*uncommittedDescriptor)(nil)

type uncommittedDescriptors struct {
	descs nstree.Map

	descNames nstree.Set

	addedSystemDatabase bool
}

func (ud *uncommittedDescriptors) reset() {
	__antithesis_instrumentation__.Notify(265128)
	ud.descs.Clear()
	ud.descNames.Clear()
	ud.addedSystemDatabase = false
}

func (ud *uncommittedDescriptors) add(
	mut catalog.MutableDescriptor, status uncommittedDescriptorStatus,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265129)
	uNew, err := makeUncommittedDescriptor(mut, status)
	if err != nil {
		__antithesis_instrumentation__.Notify(265133)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265134)
	}
	__antithesis_instrumentation__.Notify(265130)
	for _, n := range uNew.immutable.GetDrainingNames() {
		__antithesis_instrumentation__.Notify(265135)
		ud.descNames.Add(n)
	}
	__antithesis_instrumentation__.Notify(265131)
	if prev, ok := ud.descs.GetByID(mut.GetID()).(*uncommittedDescriptor); ok {
		__antithesis_instrumentation__.Notify(265136)
		if prev.mutable.OriginalVersion() != mut.OriginalVersion() {
			__antithesis_instrumentation__.Notify(265137)
			return nil, errors.AssertionFailedf(
				"cannot add a version of descriptor with a different original version" +
					" than it was previously added with")
		} else {
			__antithesis_instrumentation__.Notify(265138)
		}
	} else {
		__antithesis_instrumentation__.Notify(265139)
	}
	__antithesis_instrumentation__.Notify(265132)
	ud.descs.Upsert(uNew)
	return uNew.immutable, err
}

func (ud *uncommittedDescriptors) checkOut(id descpb.ID) (_ catalog.MutableDescriptor, err error) {
	__antithesis_instrumentation__.Notify(265140)
	defer func() {
		__antithesis_instrumentation__.Notify(265145)
		err = errors.NewAssertionErrorWithWrappedErrf(
			err, "cannot check out uncommitted descriptor with ID %d", id,
		)
	}()
	__antithesis_instrumentation__.Notify(265141)
	if id == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(265146)
		ud.maybeAddSystemDatabase()
	} else {
		__antithesis_instrumentation__.Notify(265147)
	}
	__antithesis_instrumentation__.Notify(265142)
	entry := ud.descs.GetByID(id)
	if entry == nil {
		__antithesis_instrumentation__.Notify(265148)
		return nil, errors.New("descriptor hasn't been added yet")
	} else {
		__antithesis_instrumentation__.Notify(265149)
	}
	__antithesis_instrumentation__.Notify(265143)
	u := entry.(*uncommittedDescriptor)
	if u.uncommittedDescriptorStatus == notValidatedYet {
		__antithesis_instrumentation__.Notify(265150)
		return nil, errors.New("descriptor hasn't been validated yet")
	} else {
		__antithesis_instrumentation__.Notify(265151)
	}
	__antithesis_instrumentation__.Notify(265144)
	return u.checkOut(), nil
}

func (ud *uncommittedDescriptors) checkIn(mut catalog.MutableDescriptor) error {
	__antithesis_instrumentation__.Notify(265152)

	_, err := ud.add(mut, checkedOutAtLeastOnce)
	return err
}

func makeUncommittedDescriptor(
	desc catalog.MutableDescriptor, status uncommittedDescriptorStatus,
) (*uncommittedDescriptor, error) {
	__antithesis_instrumentation__.Notify(265153)
	version := desc.GetVersion()
	origVersion := desc.OriginalVersion()
	if version != origVersion && func() bool {
		__antithesis_instrumentation__.Notify(265156)
		return version != origVersion+1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(265157)
		return nil, errors.AssertionFailedf(
			"descriptor %d version %d not compatible with cluster version %d",
			desc.GetID(), version, origVersion)
	} else {
		__antithesis_instrumentation__.Notify(265158)
	}
	__antithesis_instrumentation__.Notify(265154)

	mutable, err := maybeRefreshCachedFieldsOnTypeDescriptor(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(265159)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265160)
	}
	__antithesis_instrumentation__.Notify(265155)

	return &uncommittedDescriptor{
		mutable:                     mutable,
		immutable:                   mutable.ImmutableCopy(),
		uncommittedDescriptorStatus: status,
	}, nil
}

func maybeRefreshCachedFieldsOnTypeDescriptor(
	desc catalog.MutableDescriptor,
) (catalog.MutableDescriptor, error) {
	__antithesis_instrumentation__.Notify(265161)
	typeDesc, ok := desc.(catalog.TypeDescriptor)
	if ok {
		__antithesis_instrumentation__.Notify(265163)
		return typedesc.UpdateCachedFieldsOnModifiedMutable(typeDesc)
	} else {
		__antithesis_instrumentation__.Notify(265164)
	}
	__antithesis_instrumentation__.Notify(265162)
	return desc, nil
}

func (ud *uncommittedDescriptors) getImmutableByID(
	id descpb.ID,
) (catalog.Descriptor, uncommittedDescriptorStatus) {
	__antithesis_instrumentation__.Notify(265165)
	if id == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(265168)
		ud.maybeAddSystemDatabase()
	} else {
		__antithesis_instrumentation__.Notify(265169)
	}
	__antithesis_instrumentation__.Notify(265166)
	entry := ud.descs.GetByID(id)
	if entry == nil {
		__antithesis_instrumentation__.Notify(265170)
		return nil, notValidatedYet
	} else {
		__antithesis_instrumentation__.Notify(265171)
	}
	__antithesis_instrumentation__.Notify(265167)
	u := entry.(*uncommittedDescriptor)
	return u.immutable, u.uncommittedDescriptorStatus
}

func (ud *uncommittedDescriptors) getByName(
	dbID descpb.ID, schemaID descpb.ID, name string,
) (hasKnownRename bool, desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(265172)
	if dbID == 0 && func() bool {
		__antithesis_instrumentation__.Notify(265176)
		return schemaID == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(265177)
		return name == systemschema.SystemDatabaseName == true
	}() == true {
		__antithesis_instrumentation__.Notify(265178)
		ud.maybeAddSystemDatabase()
	} else {
		__antithesis_instrumentation__.Notify(265179)
	}
	__antithesis_instrumentation__.Notify(265173)

	if got := ud.descs.GetByName(dbID, schemaID, name); got != nil {
		__antithesis_instrumentation__.Notify(265180)
		u := got.(*uncommittedDescriptor)
		if u.uncommittedDescriptorStatus == notValidatedYet {
			__antithesis_instrumentation__.Notify(265182)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(265183)
		}
		__antithesis_instrumentation__.Notify(265181)
		return false, u.immutable
	} else {
		__antithesis_instrumentation__.Notify(265184)
	}
	__antithesis_instrumentation__.Notify(265174)

	if ud.descNames.Empty() {
		__antithesis_instrumentation__.Notify(265185)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(265186)
	}
	__antithesis_instrumentation__.Notify(265175)
	return ud.descNames.Contains(descpb.NameInfo{
		ParentID:       dbID,
		ParentSchemaID: schemaID,
		Name:           name,
	}), nil
}

func (ud *uncommittedDescriptors) getUnvalidatedByName(
	dbID descpb.ID, schemaID descpb.ID, name string,
) catalog.Descriptor {
	__antithesis_instrumentation__.Notify(265187)
	if dbID == 0 && func() bool {
		__antithesis_instrumentation__.Notify(265191)
		return schemaID == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(265192)
		return name == systemschema.SystemDatabaseName == true
	}() == true {
		__antithesis_instrumentation__.Notify(265193)
		ud.maybeAddSystemDatabase()
	} else {
		__antithesis_instrumentation__.Notify(265194)
	}
	__antithesis_instrumentation__.Notify(265188)
	entry := ud.descs.GetByName(dbID, schemaID, name)
	if entry == nil {
		__antithesis_instrumentation__.Notify(265195)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(265196)
	}
	__antithesis_instrumentation__.Notify(265189)
	u := entry.(*uncommittedDescriptor)
	if u.uncommittedDescriptorStatus != notValidatedYet {
		__antithesis_instrumentation__.Notify(265197)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(265198)
	}
	__antithesis_instrumentation__.Notify(265190)
	return u.immutable
}

func (ud *uncommittedDescriptors) iterateNewVersionByID(
	fn func(originalVersion lease.IDVersion) error,
) error {
	__antithesis_instrumentation__.Notify(265199)
	return ud.descs.IterateByID(func(entry catalog.NameEntry) error {
		__antithesis_instrumentation__.Notify(265200)
		u := entry.(*uncommittedDescriptor)
		if u.uncommittedDescriptorStatus == notValidatedYet {
			__antithesis_instrumentation__.Notify(265203)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(265204)
		}
		__antithesis_instrumentation__.Notify(265201)
		mut := u.mutable
		if mut == nil || func() bool {
			__antithesis_instrumentation__.Notify(265205)
			return mut.IsNew() == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(265206)
			return !mut.IsUncommittedVersion() == true
		}() == true {
			__antithesis_instrumentation__.Notify(265207)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(265208)
		}
		__antithesis_instrumentation__.Notify(265202)
		return fn(lease.NewIDVersionPrev(mut.OriginalName(), mut.OriginalID(), mut.OriginalVersion()))
	})
}

func (ud *uncommittedDescriptors) iterateUncommittedByID(
	fn func(imm catalog.Descriptor) error,
) error {
	__antithesis_instrumentation__.Notify(265209)
	return ud.descs.IterateByID(func(entry catalog.NameEntry) error {
		__antithesis_instrumentation__.Notify(265210)
		u := entry.(*uncommittedDescriptor)
		if u.uncommittedDescriptorStatus != checkedOutAtLeastOnce || func() bool {
			__antithesis_instrumentation__.Notify(265212)
			return !u.immutable.IsUncommittedVersion() == true
		}() == true {
			__antithesis_instrumentation__.Notify(265213)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(265214)
		}
		__antithesis_instrumentation__.Notify(265211)
		return fn(u.immutable)
	})
}

func (ud *uncommittedDescriptors) getUncommittedTables() (tables []catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(265215)
	_ = ud.iterateUncommittedByID(func(desc catalog.Descriptor) error {
		__antithesis_instrumentation__.Notify(265217)
		if table, ok := desc.(catalog.TableDescriptor); ok {
			__antithesis_instrumentation__.Notify(265219)
			tables = append(tables, table)
		} else {
			__antithesis_instrumentation__.Notify(265220)
		}
		__antithesis_instrumentation__.Notify(265218)
		return nil
	})
	__antithesis_instrumentation__.Notify(265216)
	return tables
}

func (ud *uncommittedDescriptors) getUncommittedDescriptorsForValidation() (
	descs []catalog.Descriptor,
) {
	__antithesis_instrumentation__.Notify(265221)
	_ = ud.iterateUncommittedByID(func(desc catalog.Descriptor) error {
		__antithesis_instrumentation__.Notify(265223)
		descs = append(descs, desc)
		return nil
	})
	__antithesis_instrumentation__.Notify(265222)
	return descs
}

func (ud *uncommittedDescriptors) hasUncommittedTables() (has bool) {
	__antithesis_instrumentation__.Notify(265224)
	_ = ud.iterateUncommittedByID(func(desc catalog.Descriptor) error {
		__antithesis_instrumentation__.Notify(265226)
		if _, has = desc.(catalog.TableDescriptor); has {
			__antithesis_instrumentation__.Notify(265228)
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(265229)
		}
		__antithesis_instrumentation__.Notify(265227)
		return nil
	})
	__antithesis_instrumentation__.Notify(265225)
	return has
}

func (ud *uncommittedDescriptors) hasUncommittedTypes() (has bool) {
	__antithesis_instrumentation__.Notify(265230)
	_ = ud.iterateUncommittedByID(func(desc catalog.Descriptor) error {
		__antithesis_instrumentation__.Notify(265232)
		if _, has = desc.(catalog.TypeDescriptor); has {
			__antithesis_instrumentation__.Notify(265234)
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(265235)
		}
		__antithesis_instrumentation__.Notify(265233)
		return nil
	})
	__antithesis_instrumentation__.Notify(265231)
	return has
}

var systemUncommittedDatabase = &uncommittedDescriptor{
	immutable: dbdesc.NewBuilder(systemschema.SystemDB.DatabaseDesc()).BuildImmutableDatabase(),

	mutable:                     nil,
	uncommittedDescriptorStatus: notCheckedOutYet,
}

func (ud *uncommittedDescriptors) maybeAddSystemDatabase() {
	__antithesis_instrumentation__.Notify(265236)
	if !ud.addedSystemDatabase {
		__antithesis_instrumentation__.Notify(265237)
		ud.addedSystemDatabase = true
		ud.descs.Upsert(systemUncommittedDatabase)
	} else {
		__antithesis_instrumentation__.Notify(265238)
	}
}
