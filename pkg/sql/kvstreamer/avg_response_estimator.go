package kvstreamer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type avgResponseEstimator struct {
	responseBytes int64
	numResponses  int64
}

const initialAvgResponseSize = 1 << 10

func (e *avgResponseEstimator) getAvgResponseSize() int64 {
	__antithesis_instrumentation__.Notify(498859)
	if e.numResponses == 0 {
		__antithesis_instrumentation__.Notify(498861)
		return initialAvgResponseSize
	} else {
		__antithesis_instrumentation__.Notify(498862)
	}
	__antithesis_instrumentation__.Notify(498860)

	return e.responseBytes / e.numResponses
}

func (e *avgResponseEstimator) update(responseBytes int64, numResponses int64) {
	__antithesis_instrumentation__.Notify(498863)
	e.responseBytes += responseBytes
	e.numResponses += numResponses
}
