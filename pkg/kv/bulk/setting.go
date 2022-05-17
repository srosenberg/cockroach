package bulk

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
)

var (
	IngestBatchSize = settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"kv.bulk_ingest.batch_size",
		"the maximum size of the payload in an AddSSTable request",
		16<<20,
	)
)

func ingestFileSize(st *cluster.Settings) int64 {
	__antithesis_instrumentation__.Notify(86404)
	desiredSize := IngestBatchSize.Get(&st.SV)
	maxCommandSize := kvserver.MaxCommandSize.Get(&st.SV)
	if desiredSize > maxCommandSize {
		__antithesis_instrumentation__.Notify(86406)
		return maxCommandSize
	} else {
		__antithesis_instrumentation__.Notify(86407)
	}
	__antithesis_instrumentation__.Notify(86405)
	return desiredSize
}
