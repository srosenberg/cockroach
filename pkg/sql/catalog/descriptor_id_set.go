package catalog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util"
)

type DescriptorIDSet struct {
	set util.FastIntSet
}

func MakeDescriptorIDSet(ids ...descpb.ID) DescriptorIDSet {
	__antithesis_instrumentation__.Notify(264059)
	s := DescriptorIDSet{}
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(264061)
		s.Add(id)
	}
	__antithesis_instrumentation__.Notify(264060)
	return s
}

var _ = MakeDescriptorIDSet

func (d *DescriptorIDSet) Add(id descpb.ID) {
	__antithesis_instrumentation__.Notify(264062)
	d.set.Add(int(id))
}

func (d DescriptorIDSet) Len() int {
	__antithesis_instrumentation__.Notify(264063)
	return d.set.Len()
}

func (d DescriptorIDSet) Contains(id descpb.ID) bool {
	__antithesis_instrumentation__.Notify(264064)
	return d.set.Contains(int(id))
}

func (d DescriptorIDSet) ForEach(f func(id descpb.ID)) {
	__antithesis_instrumentation__.Notify(264065)
	d.set.ForEach(func(i int) { __antithesis_instrumentation__.Notify(264066); f(descpb.ID(i)) })
}

func (d DescriptorIDSet) Empty() bool {
	__antithesis_instrumentation__.Notify(264067)
	return d.set.Empty()
}

func (d DescriptorIDSet) Ordered() []descpb.ID {
	__antithesis_instrumentation__.Notify(264068)
	if d.Empty() {
		__antithesis_instrumentation__.Notify(264071)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(264072)
	}
	__antithesis_instrumentation__.Notify(264069)
	result := make([]descpb.ID, 0, d.Len())
	d.ForEach(func(i descpb.ID) {
		__antithesis_instrumentation__.Notify(264073)
		result = append(result, i)
	})
	__antithesis_instrumentation__.Notify(264070)
	return result
}

func (d *DescriptorIDSet) Remove(id descpb.ID) {
	__antithesis_instrumentation__.Notify(264074)
	d.set.Remove(int(id))
}
