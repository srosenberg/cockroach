package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(470991)

	var x [1]struct{}
	_ = x[NeedMoreRows-0]
	_ = x[DrainRequested-1]
	_ = x[ConsumerClosed-2]
}

const _ConsumerStatus_name = "NeedMoreRowsDrainRequestedConsumerClosed"

var _ConsumerStatus_index = [...]uint8{0, 12, 26, 40}

func (i ConsumerStatus) String() string {
	__antithesis_instrumentation__.Notify(470992)
	if i >= ConsumerStatus(len(_ConsumerStatus_index)-1) {
		__antithesis_instrumentation__.Notify(470994)
		return "ConsumerStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(470995)
	}
	__antithesis_instrumentation__.Notify(470993)
	return _ConsumerStatus_name[_ConsumerStatus_index[i]:_ConsumerStatus_index[i+1]]
}
