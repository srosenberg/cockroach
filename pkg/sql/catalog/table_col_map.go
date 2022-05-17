package catalog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util"
)

type TableColMap struct {
	m util.FastIntMap

	systemColMap [NumSystemColumns]int

	systemColIsSet [NumSystemColumns]bool
}

func (s *TableColMap) Set(col descpb.ColumnID, val int) {
	__antithesis_instrumentation__.Notify(268437)
	if col < SmallestSystemColumnColumnID {
		__antithesis_instrumentation__.Notify(268438)
		s.m.Set(int(col), val)
	} else {
		__antithesis_instrumentation__.Notify(268439)
		pos := col - SmallestSystemColumnColumnID
		s.systemColMap[pos] = val
		s.systemColIsSet[pos] = true
	}
}

func (s *TableColMap) Get(col descpb.ColumnID) (val int, ok bool) {
	__antithesis_instrumentation__.Notify(268440)
	if col < SmallestSystemColumnColumnID {
		__antithesis_instrumentation__.Notify(268442)
		return s.m.Get(int(col))
	} else {
		__antithesis_instrumentation__.Notify(268443)
	}
	__antithesis_instrumentation__.Notify(268441)
	pos := col - SmallestSystemColumnColumnID
	return s.systemColMap[pos], s.systemColIsSet[pos]
}

func (s *TableColMap) GetDefault(col descpb.ColumnID) (val int) {
	__antithesis_instrumentation__.Notify(268444)
	if col < SmallestSystemColumnColumnID {
		__antithesis_instrumentation__.Notify(268446)
		return s.m.GetDefault(int(col))
	} else {
		__antithesis_instrumentation__.Notify(268447)
	}
	__antithesis_instrumentation__.Notify(268445)
	pos := col - SmallestSystemColumnColumnID
	return s.systemColMap[pos]
}

func (s *TableColMap) Len() (val int) {
	__antithesis_instrumentation__.Notify(268448)
	l := s.m.Len()
	for _, isSet := range s.systemColIsSet {
		__antithesis_instrumentation__.Notify(268450)
		if isSet {
			__antithesis_instrumentation__.Notify(268451)
			l++
		} else {
			__antithesis_instrumentation__.Notify(268452)
		}
	}
	__antithesis_instrumentation__.Notify(268449)
	return l
}

func (s *TableColMap) ForEach(f func(colID descpb.ColumnID, returnIndex int)) {
	__antithesis_instrumentation__.Notify(268453)
	s.m.ForEach(func(k, v int) {
		__antithesis_instrumentation__.Notify(268455)
		f(descpb.ColumnID(k), v)
	})
	__antithesis_instrumentation__.Notify(268454)
	for pos, isSet := range s.systemColIsSet {
		__antithesis_instrumentation__.Notify(268456)
		if isSet {
			__antithesis_instrumentation__.Notify(268457)
			id := SmallestSystemColumnColumnID + pos
			f(descpb.ColumnID(id), s.systemColMap[pos])
		} else {
			__antithesis_instrumentation__.Notify(268458)
		}
	}
}

func (s *TableColMap) String() string {
	__antithesis_instrumentation__.Notify(268459)
	var buf bytes.Buffer
	buf.WriteString("map[")
	s.m.ContentsIntoBuffer(&buf)
	first := buf.Len() == len("map[")
	for pos, isSet := range s.systemColIsSet {
		__antithesis_instrumentation__.Notify(268461)
		if isSet {
			__antithesis_instrumentation__.Notify(268462)
			if !first {
				__antithesis_instrumentation__.Notify(268464)
				buf.WriteByte(' ')
			} else {
				__antithesis_instrumentation__.Notify(268465)
			}
			__antithesis_instrumentation__.Notify(268463)
			first = false
			id := SmallestSystemColumnColumnID + pos
			fmt.Fprintf(&buf, "%d:%d", id, s.systemColMap[pos])
		} else {
			__antithesis_instrumentation__.Notify(268466)
		}
	}
	__antithesis_instrumentation__.Notify(268460)
	buf.WriteByte(']')
	return buf.String()
}
