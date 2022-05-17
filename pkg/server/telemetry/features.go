package telemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"math"
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

func Bucket10(num int64) int64 {
	__antithesis_instrumentation__.Notify(238207)
	if num == math.MinInt64 {
		__antithesis_instrumentation__.Notify(238212)

		return -1000000000000000000
	} else {
		__antithesis_instrumentation__.Notify(238213)
	}
	__antithesis_instrumentation__.Notify(238208)
	sign := int64(1)
	if num < 0 {
		__antithesis_instrumentation__.Notify(238214)
		sign = -1
		num = -num
	} else {
		__antithesis_instrumentation__.Notify(238215)
	}
	__antithesis_instrumentation__.Notify(238209)
	if num < 10 {
		__antithesis_instrumentation__.Notify(238216)
		return num * sign
	} else {
		__antithesis_instrumentation__.Notify(238217)
	}
	__antithesis_instrumentation__.Notify(238210)
	res := int64(10)
	for ; res < 1000000000000000000 && func() bool {
		__antithesis_instrumentation__.Notify(238218)
		return res*10 <= num == true
	}() == true; res *= 10 {
		__antithesis_instrumentation__.Notify(238219)
	}
	__antithesis_instrumentation__.Notify(238211)
	return res * sign
}

func CountBucketed(prefix string, value int64) {
	__antithesis_instrumentation__.Notify(238220)
	Count(fmt.Sprintf("%s.%d", prefix, Bucket10(value)))
}

func Count(feature string) {
	__antithesis_instrumentation__.Notify(238221)
	Inc(GetCounter(feature))
}

type Counter *int32

func Inc(c Counter) {
	__antithesis_instrumentation__.Notify(238222)
	atomic.AddInt32(c, 1)
}

func Read(c Counter) int32 {
	__antithesis_instrumentation__.Notify(238223)
	return atomic.LoadInt32(c)
}

func GetCounterOnce(feature string) Counter {
	__antithesis_instrumentation__.Notify(238224)
	counters.RLock()
	_, ok := counters.m[feature]
	counters.RUnlock()
	if ok {
		__antithesis_instrumentation__.Notify(238226)
		panic("counter already exists: " + feature)
	} else {
		__antithesis_instrumentation__.Notify(238227)
	}
	__antithesis_instrumentation__.Notify(238225)
	return GetCounter(feature)
}

func GetCounter(feature string) Counter {
	__antithesis_instrumentation__.Notify(238228)
	counters.RLock()
	i, ok := counters.m[feature]
	counters.RUnlock()
	if ok {
		__antithesis_instrumentation__.Notify(238231)
		return i
	} else {
		__antithesis_instrumentation__.Notify(238232)
	}
	__antithesis_instrumentation__.Notify(238229)

	counters.Lock()
	defer counters.Unlock()
	i, ok = counters.m[feature]
	if !ok {
		__antithesis_instrumentation__.Notify(238233)
		i = new(int32)
		counters.m[feature] = i
	} else {
		__antithesis_instrumentation__.Notify(238234)
	}
	__antithesis_instrumentation__.Notify(238230)
	return i
}

type CounterWithMetric struct {
	telemetry Counter
	metric    *metric.Counter
}

var _ metric.Iterable = CounterWithMetric{}

func NewCounterWithMetric(metadata metric.Metadata) CounterWithMetric {
	__antithesis_instrumentation__.Notify(238235)
	return CounterWithMetric{
		telemetry: GetCounter(metadata.Name),
		metric:    metric.NewCounter(metadata),
	}
}

func (c CounterWithMetric) Inc() {
	__antithesis_instrumentation__.Notify(238236)
	Inc(c.telemetry)
	c.metric.Inc(1)
}

func (c CounterWithMetric) Count() int64 {
	__antithesis_instrumentation__.Notify(238237)
	return c.metric.Count()
}

func (c CounterWithMetric) GetName() string {
	__antithesis_instrumentation__.Notify(238238)
	return c.metric.GetName()
}

func (c CounterWithMetric) GetHelp() string {
	__antithesis_instrumentation__.Notify(238239)
	return c.metric.GetHelp()
}

func (c CounterWithMetric) GetMeasurement() string {
	__antithesis_instrumentation__.Notify(238240)
	return c.metric.GetMeasurement()
}

func (c CounterWithMetric) GetUnit() metric.Unit {
	__antithesis_instrumentation__.Notify(238241)
	return c.metric.GetUnit()
}

func (c CounterWithMetric) GetMetadata() metric.Metadata {
	__antithesis_instrumentation__.Notify(238242)
	return c.metric.GetMetadata()
}

func (c CounterWithMetric) Inspect(f func(interface{})) {
	__antithesis_instrumentation__.Notify(238243)
	c.metric.Inspect(f)
}

func init() {
	counters.m = make(map[string]Counter, approxFeatureCount)
}

var approxFeatureCount = 1500

var counters struct {
	syncutil.RWMutex
	m map[string]Counter
}

type QuantizeCounts bool

type ResetCounters bool

const (
	Quantized QuantizeCounts = true

	Raw QuantizeCounts = false

	ResetCounts ResetCounters = true

	ReadOnly ResetCounters = false
)

func GetRawFeatureCounts() map[string]int32 {
	__antithesis_instrumentation__.Notify(238244)
	return GetFeatureCounts(Raw, ReadOnly)
}

func GetFeatureCounts(quantize QuantizeCounts, reset ResetCounters) map[string]int32 {
	__antithesis_instrumentation__.Notify(238245)
	counters.RLock()
	m := make(map[string]int32, len(counters.m))
	for k, cnt := range counters.m {
		__antithesis_instrumentation__.Notify(238248)
		var val int32
		if reset {
			__antithesis_instrumentation__.Notify(238250)
			val = atomic.SwapInt32(cnt, 0)
		} else {
			__antithesis_instrumentation__.Notify(238251)
			val = atomic.LoadInt32(cnt)
		}
		__antithesis_instrumentation__.Notify(238249)
		if val != 0 {
			__antithesis_instrumentation__.Notify(238252)
			m[k] = val
		} else {
			__antithesis_instrumentation__.Notify(238253)
		}
	}
	__antithesis_instrumentation__.Notify(238246)
	counters.RUnlock()
	if quantize {
		__antithesis_instrumentation__.Notify(238254)
		for k := range m {
			__antithesis_instrumentation__.Notify(238255)
			m[k] = int32(Bucket10(int64(m[k])))
		}
	} else {
		__antithesis_instrumentation__.Notify(238256)
	}
	__antithesis_instrumentation__.Notify(238247)
	return m
}

const ValidationTelemetryKeyPrefix = "sql.schema.validation_errors."

func RecordError(err error) {
	__antithesis_instrumentation__.Notify(238257)
	if err == nil {
		__antithesis_instrumentation__.Notify(238259)
		return
	} else {
		__antithesis_instrumentation__.Notify(238260)
	}
	__antithesis_instrumentation__.Notify(238258)

	code := pgerror.GetPGCode(err)
	Count("errorcodes." + code.String())

	tkeys := errors.GetTelemetryKeys(err)
	if len(tkeys) > 0 {
		__antithesis_instrumentation__.Notify(238261)
		var prefix string
		switch code {
		case pgcode.FeatureNotSupported:
			__antithesis_instrumentation__.Notify(238263)
			prefix = "unimplemented."
		case pgcode.Internal:
			__antithesis_instrumentation__.Notify(238264)
			prefix = "internalerror."
		default:
			__antithesis_instrumentation__.Notify(238265)
			prefix = "othererror." + code.String() + "."
		}
		__antithesis_instrumentation__.Notify(238262)
		for _, tk := range tkeys {
			__antithesis_instrumentation__.Notify(238266)
			prefixedTelemetryKey := prefix + tk
			if strings.HasPrefix(tk, ValidationTelemetryKeyPrefix) {
				__antithesis_instrumentation__.Notify(238268)

				prefixedTelemetryKey = tk
			} else {
				__antithesis_instrumentation__.Notify(238269)
			}
			__antithesis_instrumentation__.Notify(238267)
			Count(prefixedTelemetryKey)
		}
	} else {
		__antithesis_instrumentation__.Notify(238270)
	}
}
