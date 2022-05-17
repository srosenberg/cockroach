package status

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"
	"text/template"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/redact"
	humanize "github.com/dustin/go-humanize"
)

var statsTemplate = template.Must(template.New("runtime stats").Funcs(template.FuncMap{
	"iBytes": humanize.IBytes,
}).Parse(`{{iBytes .MemRSSBytes}} RSS, {{.GoroutineCount}} goroutines (stacks: {{iBytes .MemStackSysBytes}}), ` +
	`{{iBytes .GoAllocBytes}}/{{iBytes .GoTotalBytes}} Go alloc/total{{if .GoStatsStaleness}}(stale){{end}} ` +
	`(heap fragmentation: {{iBytes .HeapFragmentBytes}}, heap reserved: {{iBytes .HeapReservedBytes}}, heap released: {{iBytes .HeapReleasedBytes}}), ` +
	`{{iBytes .CGoAllocBytes}}/{{iBytes .CGoTotalBytes}} CGO alloc/total ({{printf "%.1f" .CGoCallRate}} CGO/sec), ` +
	`{{printf "%.1f" .CPUUserPercent}}/{{printf "%.1f" .CPUSysPercent}} %(u/s)time, {{printf "%.1f" .GCPausePercent}} %gc ({{.GCRunCount}}x), ` +
	`{{iBytes .NetHostRecvBytes}}/{{iBytes .NetHostSendBytes}} (r/w)net`))

func logStats(ctx context.Context, stats *eventpb.RuntimeStats) {
	__antithesis_instrumentation__.Notify(235700)

	log.StructuredEvent(ctx, stats)

	var buf strings.Builder
	if err := statsTemplate.Execute(&buf, stats); err != nil {
		__antithesis_instrumentation__.Notify(235702)
		log.Warningf(ctx, "failed to render runtime stats: %v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(235703)
	}
	__antithesis_instrumentation__.Notify(235701)
	log.Health.Infof(ctx, "runtime stats: %s", redact.SafeString(buf.String()))
}
