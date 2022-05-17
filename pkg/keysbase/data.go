package keysbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

var KeyMax = []byte{0xff, 0xff}

func PrefixEnd(b []byte) []byte {
	__antithesis_instrumentation__.Notify(85919)
	if len(b) == 0 {
		__antithesis_instrumentation__.Notify(85922)
		return KeyMax
	} else {
		__antithesis_instrumentation__.Notify(85923)
	}
	__antithesis_instrumentation__.Notify(85920)

	end := make([]byte, len(b))
	copy(end, b)
	for i := len(end) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(85924)
		end[i] = end[i] + 1
		if end[i] != 0 {
			__antithesis_instrumentation__.Notify(85925)
			return end[:i+1]
		} else {
			__antithesis_instrumentation__.Notify(85926)
		}
	}
	__antithesis_instrumentation__.Notify(85921)

	return b
}
