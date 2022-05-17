package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
)

type syntheticDescriptors struct {
	descs nstree.Map
}

func (sd *syntheticDescriptors) add(desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(264878)
	if mut, ok := desc.(catalog.MutableDescriptor); ok {
		__antithesis_instrumentation__.Notify(264879)
		desc = mut.ImmutableCopy()
		sd.descs.Upsert(desc)
	} else {
		__antithesis_instrumentation__.Notify(264880)

		sd.descs.Upsert(desc)
	}
}

func (sd *syntheticDescriptors) remove(id descpb.ID) {
	__antithesis_instrumentation__.Notify(264881)
	sd.descs.Remove(id)
}

func (sd *syntheticDescriptors) set(descs []catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(264882)
	sd.descs.Clear()
	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(264883)
		sd.add(desc)
	}
}

func (sd *syntheticDescriptors) reset() {
	__antithesis_instrumentation__.Notify(264884)
	sd.descs.Clear()
}

func (sd *syntheticDescriptors) getByName(
	dbID descpb.ID, schemaID descpb.ID, name string,
) (found bool, desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(264885)
	if entry := sd.descs.GetByName(dbID, schemaID, name); entry != nil {
		__antithesis_instrumentation__.Notify(264887)
		return true, entry.(catalog.Descriptor)
	} else {
		__antithesis_instrumentation__.Notify(264888)
	}
	__antithesis_instrumentation__.Notify(264886)
	return false, nil
}

func (sd *syntheticDescriptors) getByID(id descpb.ID) (found bool, desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(264889)
	if entry := sd.descs.GetByID(id); entry != nil {
		__antithesis_instrumentation__.Notify(264891)
		return true, entry.(catalog.Descriptor)
	} else {
		__antithesis_instrumentation__.Notify(264892)
	}
	__antithesis_instrumentation__.Notify(264890)
	return false, nil
}
