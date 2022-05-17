package physicalplan

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
)

var flowSpecPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(562534)
		return &execinfrapb.FlowSpec{}
	},
}

func NewFlowSpec(flowID execinfrapb.FlowID, gateway base.SQLInstanceID) *execinfrapb.FlowSpec {
	__antithesis_instrumentation__.Notify(562535)
	spec := flowSpecPool.Get().(*execinfrapb.FlowSpec)
	spec.FlowID = flowID
	spec.Gateway = gateway
	return spec
}

func ReleaseFlowSpec(spec *execinfrapb.FlowSpec) {
	__antithesis_instrumentation__.Notify(562536)
	for i := range spec.Processors {
		__antithesis_instrumentation__.Notify(562538)
		if tr := spec.Processors[i].Core.TableReader; tr != nil {
			__antithesis_instrumentation__.Notify(562539)
			releaseTableReaderSpec(tr)
		} else {
			__antithesis_instrumentation__.Notify(562540)
		}
	}
	__antithesis_instrumentation__.Notify(562537)
	*spec = execinfrapb.FlowSpec{
		Processors: spec.Processors[:0],
	}
	flowSpecPool.Put(spec)
}

var trSpecPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(562541)
		return &execinfrapb.TableReaderSpec{}
	},
}

func NewTableReaderSpec() *execinfrapb.TableReaderSpec {
	__antithesis_instrumentation__.Notify(562542)
	return trSpecPool.Get().(*execinfrapb.TableReaderSpec)
}

func releaseTableReaderSpec(s *execinfrapb.TableReaderSpec) {
	__antithesis_instrumentation__.Notify(562543)
	*s = execinfrapb.TableReaderSpec{}
	trSpecPool.Put(s)
}
