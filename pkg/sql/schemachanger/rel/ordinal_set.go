package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "math/bits"

type ordinal uint64

type ordinalSet uint64

func (m ordinalSet) forEach(f func(a ordinal) (wantMore bool)) {
	__antithesis_instrumentation__.Notify(578585)
	rem := m
	for rem > 0 {
		__antithesis_instrumentation__.Notify(578586)
		ord := ordinal(bits.TrailingZeros64(uint64(rem)))
		if !f(ord) {
			__antithesis_instrumentation__.Notify(578588)
			return
		} else {
			__antithesis_instrumentation__.Notify(578589)
		}
		__antithesis_instrumentation__.Notify(578587)
		rem = rem.remove(ord)
	}
}

func (m ordinalSet) remove(ord ordinal) ordinalSet {
	__antithesis_instrumentation__.Notify(578590)
	return m & ^(1 << ord)
}

func (m ordinalSet) contains(ord ordinal) bool {
	__antithesis_instrumentation__.Notify(578591)
	return m&(1<<ord) != 0
}

func (m ordinalSet) add(ord ordinal) ordinalSet {
	__antithesis_instrumentation__.Notify(578592)
	return m | (1 << ord)
}

func (m ordinalSet) without(other ordinalSet) ordinalSet {
	__antithesis_instrumentation__.Notify(578593)
	return m & ^other
}

func (m ordinalSet) intersection(other ordinalSet) ordinalSet {
	__antithesis_instrumentation__.Notify(578594)
	return m & other
}

func (m ordinalSet) union(other ordinalSet) ordinalSet {
	__antithesis_instrumentation__.Notify(578595)
	return m | other
}

func (m ordinalSet) len() int {
	__antithesis_instrumentation__.Notify(578596)
	return bits.OnesCount64(uint64(m))
}
