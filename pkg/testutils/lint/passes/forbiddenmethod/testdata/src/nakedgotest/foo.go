package nakedgotest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func toot() {
	__antithesis_instrumentation__.Notify(644789)

}

func A() {
	__antithesis_instrumentation__.Notify(644790)

	go func() { __antithesis_instrumentation__.Notify(644796) }()
	__antithesis_instrumentation__.Notify(644791)
	go toot()
	go toot()
	go toot()

	go func() { __antithesis_instrumentation__.Notify(644797) }()
	__antithesis_instrumentation__.Notify(644792)

	go func() { __antithesis_instrumentation__.Notify(644798) }()
	__antithesis_instrumentation__.Notify(644793)

	go toot()

	go func() {
		__antithesis_instrumentation__.Notify(644799)
		_ = 0
	}()
	__antithesis_instrumentation__.Notify(644794)

	go func() {
		__antithesis_instrumentation__.Notify(644800)
		_ = 0
	}()
	__antithesis_instrumentation__.Notify(644795)

	go func() {
		__antithesis_instrumentation__.Notify(644801)
		_ = 0
	}()
}
