package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/metric"

var (
	MetaMaxRowSizeLog = metric.Metadata{
		Name:        "sql.guardrails.max_row_size_log.count",
		Help:        "Number of rows observed violating sql.guardrails.max_row_size_log",
		Measurement: "Rows",
		Unit:        metric.Unit_COUNT,
	}

	MetaMaxRowSizeErr = metric.Metadata{
		Name:        "sql.guardrails.max_row_size_err.count",
		Help:        "Number of rows observed violating sql.guardrails.max_row_size_err",
		Measurement: "Rows",
		Unit:        metric.Unit_COUNT,
	}
)

type Metrics struct {
	MaxRowSizeLogCount *metric.Counter
	MaxRowSizeErrCount *metric.Counter
}

var _ metric.Struct = Metrics{}

func (Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(568482) }
