package geo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func hilbertInverse(n, x, y uint64) uint64 {
	__antithesis_instrumentation__.Notify(64667)
	var d uint64
	for s := n / 2; s > 0; s /= 2 {
		__antithesis_instrumentation__.Notify(64669)
		var rx uint64
		if (x & s) > 0 {
			__antithesis_instrumentation__.Notify(64672)
			rx = 1
		} else {
			__antithesis_instrumentation__.Notify(64673)
		}
		__antithesis_instrumentation__.Notify(64670)
		var ry uint64
		if (y & s) > 0 {
			__antithesis_instrumentation__.Notify(64674)
			ry = 1
		} else {
			__antithesis_instrumentation__.Notify(64675)
		}
		__antithesis_instrumentation__.Notify(64671)
		d += s * s * ((3 * rx) ^ ry)
		x, y = hilbertRotate(n, x, y, rx, ry)
	}
	__antithesis_instrumentation__.Notify(64668)
	return d
}

func hilbertRotate(n, x, y, rx, ry uint64) (uint64, uint64) {
	__antithesis_instrumentation__.Notify(64676)
	if ry == 0 {
		__antithesis_instrumentation__.Notify(64678)
		if rx == 1 {
			__antithesis_instrumentation__.Notify(64680)
			x = n - 1 - x
			y = n - 1 - y
		} else {
			__antithesis_instrumentation__.Notify(64681)
		}
		__antithesis_instrumentation__.Notify(64679)

		x, y = y, x
	} else {
		__antithesis_instrumentation__.Notify(64682)
	}
	__antithesis_instrumentation__.Notify(64677)
	return x, y
}
