package metrictestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"regexp"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

func GetMetricsText(registry *metric.Registry, re *regexp.Regexp) (string, error) {
	__antithesis_instrumentation__.Notify(645557)
	ex := metric.MakePrometheusExporter()
	scrape := func(ex *metric.PrometheusExporter) {
		__antithesis_instrumentation__.Notify(645562)
		ex.ScrapeRegistry(registry, true)
	}
	__antithesis_instrumentation__.Notify(645558)
	var in bytes.Buffer
	if err := ex.ScrapeAndPrintAsText(&in, scrape); err != nil {
		__antithesis_instrumentation__.Notify(645563)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(645564)
	}
	__antithesis_instrumentation__.Notify(645559)
	sc := bufio.NewScanner(&in)
	var outLines []string
	for sc.Scan() {
		__antithesis_instrumentation__.Notify(645565)
		if bytes.HasPrefix(sc.Bytes(), []byte{'#'}) || func() bool {
			__antithesis_instrumentation__.Notify(645567)
			return !re.Match(sc.Bytes()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(645568)
			continue
		} else {
			__antithesis_instrumentation__.Notify(645569)
		}
		__antithesis_instrumentation__.Notify(645566)
		outLines = append(outLines, sc.Text())
	}
	__antithesis_instrumentation__.Notify(645560)
	if err := sc.Err(); err != nil {
		__antithesis_instrumentation__.Notify(645570)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(645571)
	}
	__antithesis_instrumentation__.Notify(645561)
	sort.Strings(outLines)
	metricsText := strings.Join(outLines, "\n")
	return metricsText, nil
}
