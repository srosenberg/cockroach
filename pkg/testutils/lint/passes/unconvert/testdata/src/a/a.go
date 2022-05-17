package a

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

var (
	f  float64 = -1
	ff         = float64(f)
	fi         = int(f)

	fff = float64(f)

	_ = float64(f) +
		float64(f)

	_ = 1 +

		float64(f) +
		2

	_ = 1 +
		2 +
		float64(f)

	_ = 1 +
		float64(f) +
		2

	_ = 1 +
		float64(f) + 2

	_ = 1 +
		float64(f) + 2 + 3

	_ = 1 +
		(float64(f) + 2 +
			3)

	_ = 1 +
		2 +
		3 +

		float64(f)

	_ = 1 +
		2 +
		3 +
		float64(f)

	_ = 1 +
		(float64(f) + 2 +
			3)

	_ = 1 +

		2 +
		3 +
		float64(f)

	_ = 1 +

		2 +
		float64(f)

	_ = 1 +
		(float64(f) + 2 +
			3)

	_ = 1 +
		(float64(f) + 2 +
			3)

	_ = 1 +
		(float64(f) + 2 +
			3)

	_ = 1 +
		(float64(f) + 2 +
			3)

	_ = 1 +
		2 +
		float64(f)
)

func foo() {
	__antithesis_instrumentation__.Notify(645435)

	if fff := float64(f); fff > 0 {
		__antithesis_instrumentation__.Notify(645437)
		panic("foo")
	} else {
		__antithesis_instrumentation__.Notify(645438)
	}
	__antithesis_instrumentation__.Notify(645436)

	if fff := float64(f); fff > 0 {
		__antithesis_instrumentation__.Notify(645439)
		panic("foo")
	} else {
		__antithesis_instrumentation__.Notify(645440)
	}
}
