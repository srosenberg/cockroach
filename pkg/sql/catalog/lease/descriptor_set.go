package lease

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/errors"
)

type descriptorSet struct {
	data []*descriptorVersionState
}

func (l *descriptorSet) String() string {
	__antithesis_instrumentation__.Notify(266047)
	var buf bytes.Buffer
	for i, s := range l.data {
		__antithesis_instrumentation__.Notify(266049)
		if i > 0 {
			__antithesis_instrumentation__.Notify(266051)
			buf.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(266052)
		}
		__antithesis_instrumentation__.Notify(266050)
		buf.WriteString(fmt.Sprintf("%d:%d", s.GetVersion(), s.getExpiration().WallTime))
	}
	__antithesis_instrumentation__.Notify(266048)
	return buf.String()
}

func (l *descriptorSet) insert(s *descriptorVersionState) {
	__antithesis_instrumentation__.Notify(266053)
	i, match := l.findIndex(s.GetVersion())
	if match {
		__antithesis_instrumentation__.Notify(266056)
		panic("unable to insert duplicate lease")
	} else {
		__antithesis_instrumentation__.Notify(266057)
	}
	__antithesis_instrumentation__.Notify(266054)
	if i == len(l.data) {
		__antithesis_instrumentation__.Notify(266058)
		l.data = append(l.data, s)
		return
	} else {
		__antithesis_instrumentation__.Notify(266059)
	}
	__antithesis_instrumentation__.Notify(266055)
	l.data = append(l.data, nil)
	copy(l.data[i+1:], l.data[i:])
	l.data[i] = s
}

func (l *descriptorSet) remove(s *descriptorVersionState) {
	__antithesis_instrumentation__.Notify(266060)
	i, match := l.findIndex(s.GetVersion())
	if !match {
		__antithesis_instrumentation__.Notify(266062)
		panic(errors.AssertionFailedf("can't find lease to remove: %s", s))
	} else {
		__antithesis_instrumentation__.Notify(266063)
	}
	__antithesis_instrumentation__.Notify(266061)
	l.data = append(l.data[:i], l.data[i+1:]...)
}

func (l *descriptorSet) find(version descpb.DescriptorVersion) *descriptorVersionState {
	__antithesis_instrumentation__.Notify(266064)
	if i, match := l.findIndex(version); match {
		__antithesis_instrumentation__.Notify(266066)
		return l.data[i]
	} else {
		__antithesis_instrumentation__.Notify(266067)
	}
	__antithesis_instrumentation__.Notify(266065)
	return nil
}

func (l *descriptorSet) findIndex(version descpb.DescriptorVersion) (int, bool) {
	__antithesis_instrumentation__.Notify(266068)
	i := sort.Search(len(l.data), func(i int) bool {
		__antithesis_instrumentation__.Notify(266071)
		s := l.data[i]
		return s.GetVersion() >= version
	})
	__antithesis_instrumentation__.Notify(266069)
	if i < len(l.data) {
		__antithesis_instrumentation__.Notify(266072)
		s := l.data[i]
		if s.GetVersion() == version {
			__antithesis_instrumentation__.Notify(266073)
			return i, true
		} else {
			__antithesis_instrumentation__.Notify(266074)
		}
	} else {
		__antithesis_instrumentation__.Notify(266075)
	}
	__antithesis_instrumentation__.Notify(266070)
	return i, false
}

func (l *descriptorSet) findNewest() *descriptorVersionState {
	__antithesis_instrumentation__.Notify(266076)
	if len(l.data) == 0 {
		__antithesis_instrumentation__.Notify(266078)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266079)
	}
	__antithesis_instrumentation__.Notify(266077)
	return l.data[len(l.data)-1]
}

func (l *descriptorSet) findVersion(version descpb.DescriptorVersion) *descriptorVersionState {
	__antithesis_instrumentation__.Notify(266080)
	if len(l.data) == 0 {
		__antithesis_instrumentation__.Notify(266085)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266086)
	}
	__antithesis_instrumentation__.Notify(266081)

	i := sort.Search(len(l.data), func(i int) bool {
		__antithesis_instrumentation__.Notify(266087)
		return l.data[i].GetVersion() > version
	})
	__antithesis_instrumentation__.Notify(266082)
	if i == 0 {
		__antithesis_instrumentation__.Notify(266088)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266089)
	}
	__antithesis_instrumentation__.Notify(266083)

	s := l.data[i-1]
	if s.GetVersion() == version {
		__antithesis_instrumentation__.Notify(266090)
		return s
	} else {
		__antithesis_instrumentation__.Notify(266091)
	}
	__antithesis_instrumentation__.Notify(266084)
	return nil
}
