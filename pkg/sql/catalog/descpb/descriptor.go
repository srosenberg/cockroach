package descpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func GetDescriptorMetadata(
	desc *Descriptor,
) (
	id ID,
	version DescriptorVersion,
	name string,
	state DescriptorState,
	modTime hlc.Timestamp,
	err error,
) {
	__antithesis_instrumentation__.Notify(251525)
	switch t := desc.Union.(type) {
	case *Descriptor_Table:
		__antithesis_instrumentation__.Notify(251527)
		id = t.Table.ID
		version = t.Table.Version
		name = t.Table.Name
		state = t.Table.State
		modTime = t.Table.ModificationTime
	case *Descriptor_Database:
		__antithesis_instrumentation__.Notify(251528)
		id = t.Database.ID
		version = t.Database.Version
		name = t.Database.Name
		state = t.Database.State
		modTime = t.Database.ModificationTime
	case *Descriptor_Type:
		__antithesis_instrumentation__.Notify(251529)
		id = t.Type.ID
		version = t.Type.Version
		name = t.Type.Name
		state = t.Type.State
		modTime = t.Type.ModificationTime
	case *Descriptor_Schema:
		__antithesis_instrumentation__.Notify(251530)
		id = t.Schema.ID
		version = t.Schema.Version
		name = t.Schema.Name
		state = t.Schema.State
		modTime = t.Schema.ModificationTime
	case nil:
		__antithesis_instrumentation__.Notify(251531)
		err = errors.AssertionFailedf("Table/Database/Type/Schema not set in descpb.Descriptor")
	default:
		__antithesis_instrumentation__.Notify(251532)
		err = errors.AssertionFailedf("Unknown descpb.Descriptor type %T", t)
	}
	__antithesis_instrumentation__.Notify(251526)
	return id, version, name, state, modTime, err
}

func GetDescriptorID(desc *Descriptor) ID {
	__antithesis_instrumentation__.Notify(251533)
	id, _, _, _, _, err := GetDescriptorMetadata(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(251535)
		panic(errors.Wrap(err, "GetDescriptorID"))
	} else {
		__antithesis_instrumentation__.Notify(251536)
	}
	__antithesis_instrumentation__.Notify(251534)
	return id
}

func GetDescriptorName(desc *Descriptor) string {
	__antithesis_instrumentation__.Notify(251537)
	_, _, name, _, _, err := GetDescriptorMetadata(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(251539)
		panic(errors.Wrap(err, "GetDescriptorName"))
	} else {
		__antithesis_instrumentation__.Notify(251540)
	}
	__antithesis_instrumentation__.Notify(251538)
	return name
}

func GetDescriptorVersion(desc *Descriptor) DescriptorVersion {
	__antithesis_instrumentation__.Notify(251541)
	_, version, _, _, _, err := GetDescriptorMetadata(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(251543)
		panic(errors.Wrap(err, "GetDescriptorVersion"))
	} else {
		__antithesis_instrumentation__.Notify(251544)
	}
	__antithesis_instrumentation__.Notify(251542)
	return version
}

func GetDescriptorModificationTime(desc *Descriptor) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(251545)
	_, _, _, _, modTime, err := GetDescriptorMetadata(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(251547)
		panic(errors.Wrap(err, "GetDescriptorModificationTime"))
	} else {
		__antithesis_instrumentation__.Notify(251548)
	}
	__antithesis_instrumentation__.Notify(251546)
	return modTime
}

func GetDescriptorState(desc *Descriptor) DescriptorState {
	__antithesis_instrumentation__.Notify(251549)
	_, _, _, state, _, err := GetDescriptorMetadata(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(251551)
		panic(errors.Wrap(err, "GetDescriptorState"))
	} else {
		__antithesis_instrumentation__.Notify(251552)
	}
	__antithesis_instrumentation__.Notify(251550)
	return state
}

func setDescriptorModificationTime(desc *Descriptor, ts hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(251553)
	switch t := desc.Union.(type) {
	case *Descriptor_Table:
		__antithesis_instrumentation__.Notify(251554)
		t.Table.ModificationTime = ts
	case *Descriptor_Database:
		__antithesis_instrumentation__.Notify(251555)
		t.Database.ModificationTime = ts
	case *Descriptor_Type:
		__antithesis_instrumentation__.Notify(251556)
		t.Type.ModificationTime = ts
	case *Descriptor_Schema:
		__antithesis_instrumentation__.Notify(251557)
		t.Schema.ModificationTime = ts
	default:
		__antithesis_instrumentation__.Notify(251558)
		panic(errors.AssertionFailedf("setModificationTime: unknown Descriptor type %T", t))
	}
}

func MaybeSetDescriptorModificationTimeFromMVCCTimestamp(desc *Descriptor, ts hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(251559)
	switch t := desc.Union.(type) {
	case nil:
		__antithesis_instrumentation__.Notify(251561)

		return
	case *Descriptor_Table:
		__antithesis_instrumentation__.Notify(251562)

		if !ts.IsEmpty() && func() bool {
			__antithesis_instrumentation__.Notify(251564)
			return t.Table.ModificationTime.IsEmpty() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(251565)
			return t.Table.CreateAsOfTime.IsEmpty() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(251566)
			return t.Table.Version == 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(251567)
			t.Table.CreateAsOfTime = ts
		} else {
			__antithesis_instrumentation__.Notify(251568)
		}
		__antithesis_instrumentation__.Notify(251563)

		if t.Table.Adding() && func() bool {
			__antithesis_instrumentation__.Notify(251569)
			return t.Table.IsAs() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(251570)
			return t.Table.CreateAsOfTime.IsEmpty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(251571)
			log.Fatalf(context.TODO(), "table descriptor for %q (%d.%d) is in the "+
				"ADD state and was created with CREATE TABLE ... AS but does not have a "+
				"CreateAsOfTime set", t.Table.Name, t.Table.ParentID, t.Table.ID)
		} else {
			__antithesis_instrumentation__.Notify(251572)
		}
	}
	__antithesis_instrumentation__.Notify(251560)

	if modTime := GetDescriptorModificationTime(desc); modTime.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(251573)
		return ts.IsEmpty() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(251574)
		return GetDescriptorVersion(desc) > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(251575)

		log.Fatalf(context.TODO(), "read table descriptor for %q (%d) without ModificationTime "+
			"with zero MVCC timestamp; full descriptor:\n%s", GetDescriptorName(desc), GetDescriptorID(desc), desc)
	} else {
		__antithesis_instrumentation__.Notify(251576)
		if modTime.IsEmpty() {
			__antithesis_instrumentation__.Notify(251577)
			setDescriptorModificationTime(desc, ts)
		} else {
			__antithesis_instrumentation__.Notify(251578)
			if !ts.IsEmpty() && func() bool {
				__antithesis_instrumentation__.Notify(251579)
				return ts.Less(modTime) == true
			}() == true {
				__antithesis_instrumentation__.Notify(251580)
				log.Fatalf(context.TODO(), "read table descriptor %q (%d) which has a ModificationTime "+
					"after its MVCC timestamp: has %v, expected %v",
					GetDescriptorName(desc), GetDescriptorID(desc), modTime, ts)
			} else {
				__antithesis_instrumentation__.Notify(251581)
			}
		}
	}
}

func FromDescriptorWithMVCCTimestamp(
	desc *Descriptor, ts hlc.Timestamp,
) (
	table *TableDescriptor,
	database *DatabaseDescriptor,
	typ *TypeDescriptor,
	schema *SchemaDescriptor,
) {
	__antithesis_instrumentation__.Notify(251582)
	if desc == nil {
		__antithesis_instrumentation__.Notify(251584)
		return nil, nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(251585)
	}
	__antithesis_instrumentation__.Notify(251583)

	table = desc.GetTable()

	database = desc.GetDatabase()

	typ = desc.GetType()

	schema = desc.GetSchema()
	MaybeSetDescriptorModificationTimeFromMVCCTimestamp(desc, ts)
	return table, database, typ, schema
}

func FromDescriptor(
	desc *Descriptor,
) (*TableDescriptor, *DatabaseDescriptor, *TypeDescriptor, *SchemaDescriptor) {
	__antithesis_instrumentation__.Notify(251586)
	return FromDescriptorWithMVCCTimestamp(desc, hlc.Timestamp{})
}
