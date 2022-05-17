package catalog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/errors"
	prometheusgo "github.com/prometheus/client_model/go"
)

const (
	Process            = `Process`
	SQLLayer           = `SQL Layer`
	KVTransactionLayer = `KV Transaction Layer`
	DistributionLayer  = `Distribution Layer`
	ReplicationLayer   = `Replication Layer`
	StorageLayer       = `Storage Layer`
	Timeseries         = `Timeseries`
	Jobs               = `Jobs`
)

type sectionDescription struct {
	Organization [][]string

	Charts []chartDescription
}

type chartDescription struct {
	Title string

	Metrics []string

	Units AxisUnits

	AxisLabel string

	Downsampler DescribeAggregator

	Aggregator DescribeAggregator

	Rate DescribeDerivative

	Percentiles bool
}

type chartDefault struct {
	Downsampler DescribeAggregator
	Aggregator  DescribeAggregator
	Rate        DescribeDerivative
	Percentiles bool
}

var chartDefaultsPerMetricType = map[prometheusgo.MetricType]chartDefault{
	prometheusgo.MetricType_COUNTER: {
		Downsampler: DescribeAggregator_AVG,
		Aggregator:  DescribeAggregator_SUM,
		Rate:        DescribeDerivative_NON_NEGATIVE_DERIVATIVE,
		Percentiles: false,
	},
	prometheusgo.MetricType_GAUGE: {
		Downsampler: DescribeAggregator_AVG,
		Aggregator:  DescribeAggregator_AVG,
		Rate:        DescribeDerivative_NONE,
		Percentiles: false,
	},
	prometheusgo.MetricType_HISTOGRAM: {
		Downsampler: DescribeAggregator_AVG,
		Aggregator:  DescribeAggregator_AVG,
		Rate:        DescribeDerivative_NONE,
		Percentiles: true,
	},
}

var chartCatalog = []ChartSection{
	{
		Title:           Process,
		LongTitle:       Process,
		CollectionTitle: "process-all",
		Description: `These charts detail the overall performance of the 
		<code>cockroach</code> process running on this server.`,
		Level: 0,
	},
	{
		Title:           SQLLayer,
		LongTitle:       SQLLayer,
		CollectionTitle: "sql-layer-all",
		Description: `In the SQL layer, nodes receive commands and then parse, plan, 
		and execute them. <br/><br/><a class="catalog-link" 
		href="https://www.cockroachlabs.com/docs/stable/architecture/sql-layer.html">
		SQL Layer Architecture Docs >></a>"`,
		Level: 0,
	},
	{
		Title:           KVTransactionLayer,
		LongTitle:       KVTransactionLayer,
		CollectionTitle: "kv-transaction-layer-all",
		Description: `The KV Transaction Layer coordinates concurrent requests as 
		key-value operations. To maintain consistency, this is also where the cluster 
		manages time. <br/><br/><a class="catalog-link" 
		href="https://www.cockroachlabs.com/docs/stable/architecture/transaction-layer.html">
		Transaction Layer Architecture Docs >></a>`,
		Level: 0,
	},
	{
		Title:           DistributionLayer,
		LongTitle:       DistributionLayer,
		CollectionTitle: "distribution-layer-all",
		Description: `The Distribution Layer provides a unified view of your cluster’s data, 
		which are actually broken up into many key-value ranges. <br/><br/><a class="catalog-link" 
		href="https://www.cockroachlabs.com/docs/stable/architecture/distribution-layer.html"> 
		Distribution Layer Architecture Docs >></a>`,
		Level: 0,
	},
	{
		Title:           ReplicationLayer,
		LongTitle:       ReplicationLayer,
		CollectionTitle: "replication-layer-all",
		Description: `The Replication Layer maintains consistency between copies of ranges 
		(known as replicas) through our consensus algorithm, Raft. <br/><br/><a class="catalog-link" 
			href="https://www.cockroachlabs.com/docs/stable/architecture/replication-layer.html"> 
			Replication Layer Architecture Docs >></a>`,
		Level: 0,
	},
	{
		Title:           StorageLayer,
		LongTitle:       StorageLayer,
		CollectionTitle: "replication-layer-all",
		Description: `The Storage Layer reads and writes data to disk, as well as manages 
		garbage collection. <br/><br/><a class="catalog-link" 
		href="https://www.cockroachlabs.com/docs/stable/architecture/storage-layer.html">
		Storage Layer Architecture Docs >></a>`,
		Level: 0,
	},
	{
		Title:           Timeseries,
		LongTitle:       Timeseries,
		CollectionTitle: "timeseries-all",
		Description: `Your cluster collects data about its own performance, which is used to 
		power the very charts you\'re using, among other things.`,
		Level: 0,
	},
	{
		Title:           Jobs,
		LongTitle:       Jobs,
		CollectionTitle: "jobs-all",
		Description:     `Your cluster executes various background jobs, as well as scheduled jobs`,
		Level:           0,
	},
}

var catalogGenerated = false

var catalogKey = map[string]int{
	Process:            0,
	SQLLayer:           1,
	KVTransactionLayer: 2,
	DistributionLayer:  3,
	ReplicationLayer:   4,
	StorageLayer:       5,
	Timeseries:         6,
	Jobs:               7,
}

var unitsKey = map[metric.Unit]AxisUnits{
	metric.Unit_BYTES:         AxisUnits_BYTES,
	metric.Unit_CONST:         AxisUnits_COUNT,
	metric.Unit_COUNT:         AxisUnits_COUNT,
	metric.Unit_NANOSECONDS:   AxisUnits_DURATION,
	metric.Unit_PERCENT:       AxisUnits_COUNT,
	metric.Unit_SECONDS:       AxisUnits_DURATION,
	metric.Unit_TIMESTAMP_NS:  AxisUnits_DURATION,
	metric.Unit_TIMESTAMP_SEC: AxisUnits_DURATION,
}

var aggKey = map[DescribeAggregator]tspb.TimeSeriesQueryAggregator{
	DescribeAggregator_AVG: tspb.TimeSeriesQueryAggregator_AVG,
	DescribeAggregator_MAX: tspb.TimeSeriesQueryAggregator_MAX,
	DescribeAggregator_MIN: tspb.TimeSeriesQueryAggregator_MIN,
	DescribeAggregator_SUM: tspb.TimeSeriesQueryAggregator_SUM,
}

var derKey = map[DescribeDerivative]tspb.TimeSeriesQueryDerivative{
	DescribeDerivative_DERIVATIVE:              tspb.TimeSeriesQueryDerivative_DERIVATIVE,
	DescribeDerivative_NONE:                    tspb.TimeSeriesQueryDerivative_NONE,
	DescribeDerivative_NON_NEGATIVE_DERIVATIVE: tspb.TimeSeriesQueryDerivative_NON_NEGATIVE_DERIVATIVE,
}

func GenerateCatalog(metadata map[string]metric.Metadata) ([]ChartSection, error) {
	__antithesis_instrumentation__.Notify(647049)

	if !catalogGenerated {
		__antithesis_instrumentation__.Notify(647051)
		for _, sd := range charts {
			__antithesis_instrumentation__.Notify(647053)

			if err := createIndividualCharts(metadata, sd); err != nil {
				__antithesis_instrumentation__.Notify(647054)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(647055)
			}
		}
		__antithesis_instrumentation__.Notify(647052)
		catalogGenerated = true
	} else {
		__antithesis_instrumentation__.Notify(647056)
	}
	__antithesis_instrumentation__.Notify(647050)

	return chartCatalog, nil
}

func createIndividualCharts(metadata map[string]metric.Metadata, sd sectionDescription) error {
	__antithesis_instrumentation__.Notify(647057)

	var ics []*IndividualChart

	for _, cd := range sd.Charts {
		__antithesis_instrumentation__.Notify(647060)

		ic := new(IndividualChart)

		if err := ic.addMetrics(cd, metadata); err != nil {
			__antithesis_instrumentation__.Notify(647064)
			return err
		} else {
			__antithesis_instrumentation__.Notify(647065)
		}
		__antithesis_instrumentation__.Notify(647061)

		if len(ic.Metrics) == 0 {
			__antithesis_instrumentation__.Notify(647066)
			continue
		} else {
			__antithesis_instrumentation__.Notify(647067)
		}
		__antithesis_instrumentation__.Notify(647062)

		if err := ic.addDisplayProperties(cd); err != nil {
			__antithesis_instrumentation__.Notify(647068)
			return err
		} else {
			__antithesis_instrumentation__.Notify(647069)
		}
		__antithesis_instrumentation__.Notify(647063)

		ic.Title = cd.Title

		ics = append(ics, ic)
	}
	__antithesis_instrumentation__.Notify(647058)

	for _, org := range sd.Organization {
		__antithesis_instrumentation__.Notify(647070)

		if len(org) < 2 {
			__antithesis_instrumentation__.Notify(647074)
			return errors.Errorf(`Sections must have at least Level 0 and 
				Level 1 organization, but only have %v in %v`, org, sd)
		} else {
			__antithesis_instrumentation__.Notify(647075)
			if len(org) > 3 {
				__antithesis_instrumentation__.Notify(647076)
				return errors.Errorf(`Sections cannot be more than 3 levels deep,
				but %v has %d`, sd, len(org))
			} else {
				__antithesis_instrumentation__.Notify(647077)
			}
		}
		__antithesis_instrumentation__.Notify(647071)

		for _, ic := range ics {
			__antithesis_instrumentation__.Notify(647078)
			ic.addNames(org)
		}
		__antithesis_instrumentation__.Notify(647072)

		topLevelCatalogIndex, ok := catalogKey[org[0]]

		if !ok {
			__antithesis_instrumentation__.Notify(647079)
			return errors.Errorf(`Undefined Level 0 organization; you must 
			use a const defined in pkg/ts/catalog/catalog_generator.go for %v`, sd)
		} else {
			__antithesis_instrumentation__.Notify(647080)
		}
		__antithesis_instrumentation__.Notify(647073)

		chartCatalog[topLevelCatalogIndex].addChartAndSubsections(org, ics)
	}
	__antithesis_instrumentation__.Notify(647059)

	return nil
}

func (ic *IndividualChart) addMetrics(
	cd chartDescription, metadata map[string]metric.Metadata,
) error {
	__antithesis_instrumentation__.Notify(647081)
	for _, x := range cd.Metrics {
		__antithesis_instrumentation__.Notify(647083)

		md, ok := metadata[x]

		if !ok {
			__antithesis_instrumentation__.Notify(647086)
			continue
		} else {
			__antithesis_instrumentation__.Notify(647087)
		}
		__antithesis_instrumentation__.Notify(647084)

		unit, ok := unitsKey[md.Unit]

		if !ok {
			__antithesis_instrumentation__.Notify(647088)
			return errors.Errorf(
				"%s's metric.Metadata has an unrecognized Unit, %v", md.Name, md.Unit,
			)
		} else {
			__antithesis_instrumentation__.Notify(647089)
		}
		__antithesis_instrumentation__.Notify(647085)

		ic.Metrics = append(ic.Metrics, ChartMetric{
			Name:           md.Name,
			Help:           md.Help,
			AxisLabel:      md.Measurement,
			PreferredUnits: unit,
			MetricType:     md.MetricType,
		})

		if ic.Metrics[0].MetricType != md.MetricType {
			__antithesis_instrumentation__.Notify(647090)
			return errors.Errorf(`%s and %s have different MetricTypes, but are being 
			added to the same chart, %v`, ic.Metrics[0].Name, md.Name, ic)
		} else {
			__antithesis_instrumentation__.Notify(647091)
		}
	}
	__antithesis_instrumentation__.Notify(647082)

	return nil
}

func (ic *IndividualChart) addNames(organization []string) {
	__antithesis_instrumentation__.Notify(647092)

	nondashDelimiters := regexp.MustCompile("( )|/|,")

	for _, n := range organization {
		__antithesis_instrumentation__.Notify(647094)
		ic.LongTitle += n + string(" | ")
		ic.CollectionTitle += nondashDelimiters.ReplaceAllString(strings.ToLower(n), "-") + "-"
	}
	__antithesis_instrumentation__.Notify(647093)

	ic.LongTitle += ic.Title
	ic.CollectionTitle += nondashDelimiters.ReplaceAllString(strings.ToLower(ic.Title), "-")

}

func (ic *IndividualChart) addDisplayProperties(cd chartDescription) error {
	__antithesis_instrumentation__.Notify(647095)

	defaults := chartDefaultsPerMetricType[ic.Metrics[0].MetricType]

	cdFull := cd

	if cdFull.Downsampler == DescribeAggregator_UNSET_AGG {
		__antithesis_instrumentation__.Notify(647105)
		cdFull.Downsampler = defaults.Downsampler
	} else {
		__antithesis_instrumentation__.Notify(647106)
	}
	__antithesis_instrumentation__.Notify(647096)
	if cdFull.Aggregator == DescribeAggregator_UNSET_AGG {
		__antithesis_instrumentation__.Notify(647107)
		cdFull.Aggregator = defaults.Aggregator
	} else {
		__antithesis_instrumentation__.Notify(647108)
	}
	__antithesis_instrumentation__.Notify(647097)
	if cdFull.Rate == DescribeDerivative_UNSET_DER {
		__antithesis_instrumentation__.Notify(647109)
		cdFull.Rate = defaults.Rate
	} else {
		__antithesis_instrumentation__.Notify(647110)
	}
	__antithesis_instrumentation__.Notify(647098)
	if !cdFull.Percentiles {
		__antithesis_instrumentation__.Notify(647111)
		cdFull.Percentiles = defaults.Percentiles
	} else {
		__antithesis_instrumentation__.Notify(647112)
	}
	__antithesis_instrumentation__.Notify(647099)

	if cdFull.Units == AxisUnits_UNSET_UNITS {
		__antithesis_instrumentation__.Notify(647113)

		pu := ic.Metrics[0].PreferredUnits
		for _, m := range ic.Metrics {
			__antithesis_instrumentation__.Notify(647115)
			if m.PreferredUnits != pu {
				__antithesis_instrumentation__.Notify(647116)
				return errors.Errorf(`Chart %s has metrics with different preferred 
				units; need to specify Units in its chartDescription: %v`, cd.Title, ic)
			} else {
				__antithesis_instrumentation__.Notify(647117)
			}
		}
		__antithesis_instrumentation__.Notify(647114)

		cdFull.Units = pu
	} else {
		__antithesis_instrumentation__.Notify(647118)
	}
	__antithesis_instrumentation__.Notify(647100)

	if cdFull.AxisLabel == "" {
		__antithesis_instrumentation__.Notify(647119)
		al := ic.Metrics[0].AxisLabel

		for _, m := range ic.Metrics {
			__antithesis_instrumentation__.Notify(647121)
			if m.AxisLabel != al {
				__antithesis_instrumentation__.Notify(647122)
				return errors.Errorf(`Chart %s has metrics with different axis labels (%s vs %s); 
				need to specify an AxisLabel in its chartDescription: %v`, al, m.AxisLabel, cd.Title, ic)
			} else {
				__antithesis_instrumentation__.Notify(647123)
			}
		}
		__antithesis_instrumentation__.Notify(647120)

		cdFull.AxisLabel = al
	} else {
		__antithesis_instrumentation__.Notify(647124)
	}
	__antithesis_instrumentation__.Notify(647101)

	ds, ok := aggKey[cdFull.Downsampler]
	if !ok {
		__antithesis_instrumentation__.Notify(647125)
		return errors.Errorf(
			"%s's chartDescription has an unrecognized Downsampler, %v", cdFull.Title, cdFull.Downsampler,
		)
	} else {
		__antithesis_instrumentation__.Notify(647126)
	}
	__antithesis_instrumentation__.Notify(647102)
	ic.Downsampler = &ds

	agg, ok := aggKey[cdFull.Aggregator]
	if !ok {
		__antithesis_instrumentation__.Notify(647127)
		return errors.Errorf(
			"%s's chartDescription has an unrecognized Aggregator, %v", cdFull.Title, cdFull.Aggregator,
		)
	} else {
		__antithesis_instrumentation__.Notify(647128)
	}
	__antithesis_instrumentation__.Notify(647103)
	ic.Aggregator = &agg

	der, ok := derKey[cdFull.Rate]
	if !ok {
		__antithesis_instrumentation__.Notify(647129)
		return errors.Errorf(
			"%s's chartDescription has an unrecognized Rate, %v", cdFull.Title, cdFull.Rate,
		)
	} else {
		__antithesis_instrumentation__.Notify(647130)
	}
	__antithesis_instrumentation__.Notify(647104)
	ic.Derivative = &der

	ic.Percentiles = cdFull.Percentiles
	ic.Units = cdFull.Units
	ic.AxisLabel = cdFull.AxisLabel

	return nil
}

func (cs *ChartSection) addChartAndSubsections(organization []string, ics []*IndividualChart) {
	__antithesis_instrumentation__.Notify(647131)

	var subsection *ChartSection

	subsectionLevel := int(cs.Level + 1)

	for _, ss := range cs.Subsections {
		__antithesis_instrumentation__.Notify(647134)
		if ss.Title == organization[subsectionLevel] {
			__antithesis_instrumentation__.Notify(647135)
			subsection = ss
			break
		} else {
			__antithesis_instrumentation__.Notify(647136)
		}
	}
	__antithesis_instrumentation__.Notify(647132)

	if subsection == nil {
		__antithesis_instrumentation__.Notify(647137)

		nondashDelimiters := regexp.MustCompile("( )|/|,")

		subsection = &ChartSection{
			Title: organization[subsectionLevel],

			LongTitle: "All",

			CollectionTitle: nondashDelimiters.ReplaceAllString(strings.ToLower(organization[0]), "-"),
			Level:           int32(subsectionLevel),
		}

		for i := 1; i <= subsectionLevel; i++ {
			__antithesis_instrumentation__.Notify(647139)
			subsection.LongTitle += " " + organization[i]
			subsection.CollectionTitle += "-" + nondashDelimiters.ReplaceAllString(strings.ToLower(organization[i]), "-")
		}
		__antithesis_instrumentation__.Notify(647138)

		cs.Subsections = append(cs.Subsections, subsection)
	} else {
		__antithesis_instrumentation__.Notify(647140)
	}
	__antithesis_instrumentation__.Notify(647133)

	if subsectionLevel == (len(organization) - 1) {
		__antithesis_instrumentation__.Notify(647141)
		subsection.Charts = append(subsection.Charts, ics...)
	} else {
		__antithesis_instrumentation__.Notify(647142)
		subsection.addChartAndSubsections(organization, ics)
	}
}
