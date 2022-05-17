package catalog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util"
)

type TableColSet struct {
	set util.FastIntSet
}

func MakeTableColSet(vals ...descpb.ColumnID) TableColSet {
	__antithesis_instrumentation__.Notify(268467)
	var res TableColSet
	for _, v := range vals {
		__antithesis_instrumentation__.Notify(268469)
		res.Add(v)
	}
	__antithesis_instrumentation__.Notify(268468)
	return res
}

func (s *TableColSet) Add(col descpb.ColumnID) {
	__antithesis_instrumentation__.Notify(268470)
	s.set.Add(int(col))
}

func (s TableColSet) Contains(col descpb.ColumnID) bool {
	__antithesis_instrumentation__.Notify(268471)
	return s.set.Contains(int(col))
}

func (s TableColSet) Empty() bool {
	__antithesis_instrumentation__.Notify(268472)
	return s.set.Empty()
}

func (s TableColSet) Len() int { __antithesis_instrumentation__.Notify(268473); return s.set.Len() }

func (s TableColSet) Next(startVal descpb.ColumnID) (descpb.ColumnID, bool) {
	__antithesis_instrumentation__.Notify(268474)
	c, ok := s.set.Next(int(startVal))
	return descpb.ColumnID(c), ok
}

func (s TableColSet) ForEach(f func(col descpb.ColumnID)) {
	__antithesis_instrumentation__.Notify(268475)
	s.set.ForEach(func(i int) { __antithesis_instrumentation__.Notify(268476); f(descpb.ColumnID(i)) })
}

func (s TableColSet) Ordered() []descpb.ColumnID {
	__antithesis_instrumentation__.Notify(268477)
	if s.Empty() {
		__antithesis_instrumentation__.Notify(268480)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(268481)
	}
	__antithesis_instrumentation__.Notify(268478)
	result := make([]descpb.ColumnID, 0, s.Len())
	s.ForEach(func(i descpb.ColumnID) {
		__antithesis_instrumentation__.Notify(268482)
		result = append(result, i)
	})
	__antithesis_instrumentation__.Notify(268479)
	return result
}

func (s *TableColSet) UnionWith(rhs TableColSet) {
	__antithesis_instrumentation__.Notify(268483)
	s.set.UnionWith(rhs.set)
}

func (s TableColSet) String() string {
	__antithesis_instrumentation__.Notify(268484)
	return s.set.String()
}
