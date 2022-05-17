package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobsprotectedts"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

const (
	jsonMetaSentinel = `__crdb__`
)

func emitResolvedTimestamp(
	ctx context.Context, encoder Encoder, sink Sink, resolved hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(15352)

	if err := sink.EmitResolvedTimestamp(ctx, encoder, resolved); err != nil {
		__antithesis_instrumentation__.Notify(15355)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15356)
	}
	__antithesis_instrumentation__.Notify(15353)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(15357)
		log.Infof(ctx, `resolved %s`, resolved)
	} else {
		__antithesis_instrumentation__.Notify(15358)
	}
	__antithesis_instrumentation__.Notify(15354)
	return nil
}

func createProtectedTimestampRecord(
	ctx context.Context,
	codec keys.SQLCodec,
	jobID jobspb.JobID,
	targets []jobspb.ChangefeedTargetSpecification,
	resolved hlc.Timestamp,
	progress *jobspb.ChangefeedProgress,
) *ptpb.Record {
	__antithesis_instrumentation__.Notify(15359)
	progress.ProtectedTimestampRecord = uuid.MakeV4()
	deprecatedSpansToProtect := makeSpansToProtect(codec, targets)
	targetToProtect := makeTargetToProtect(targets)

	log.VEventf(ctx, 2, "creating protected timestamp %v at %v", progress.ProtectedTimestampRecord, resolved)
	return jobsprotectedts.MakeRecord(
		progress.ProtectedTimestampRecord, int64(jobID), resolved, deprecatedSpansToProtect,
		jobsprotectedts.Jobs, targetToProtect)
}

func makeTargetToProtect(targets []jobspb.ChangefeedTargetSpecification) *ptpb.Target {
	__antithesis_instrumentation__.Notify(15360)

	tablesToProtect := make(descpb.IDs, 0, len(targets)+1)
	for _, t := range targets {
		__antithesis_instrumentation__.Notify(15362)
		tablesToProtect = append(tablesToProtect, t.TableID)
	}
	__antithesis_instrumentation__.Notify(15361)
	tablesToProtect = append(tablesToProtect, keys.DescriptorTableID)
	return ptpb.MakeSchemaObjectsTarget(tablesToProtect)
}

func makeSpansToProtect(
	codec keys.SQLCodec, targets []jobspb.ChangefeedTargetSpecification,
) []roachpb.Span {
	__antithesis_instrumentation__.Notify(15363)

	spansToProtect := make([]roachpb.Span, 0, len(targets)+1)
	addTablePrefix := func(id uint32) {
		__antithesis_instrumentation__.Notify(15366)
		tablePrefix := codec.TablePrefix(id)
		spansToProtect = append(spansToProtect, roachpb.Span{
			Key:    tablePrefix,
			EndKey: tablePrefix.PrefixEnd(),
		})
	}
	__antithesis_instrumentation__.Notify(15364)
	for _, t := range targets {
		__antithesis_instrumentation__.Notify(15367)
		addTablePrefix(uint32(t.TableID))
	}
	__antithesis_instrumentation__.Notify(15365)
	addTablePrefix(keys.DescriptorTableID)
	return spansToProtect
}

func initialScanTypeFromOpts(opts map[string]string) (changefeedbase.InitialScanType, error) {
	__antithesis_instrumentation__.Notify(15368)
	_, cursor := opts[changefeedbase.OptCursor]
	initialScanType, initialScanSet := opts[changefeedbase.OptInitialScan]
	_, initialScanOnlySet := opts[changefeedbase.OptInitialScanOnly]
	_, noInitialScanSet := opts[changefeedbase.OptNoInitialScan]

	if initialScanSet && func() bool {
		__antithesis_instrumentation__.Notify(15376)
		return noInitialScanSet == true
	}() == true {
		__antithesis_instrumentation__.Notify(15377)
		return changefeedbase.InitialScan, errors.Errorf(
			`cannot specify both %s and %s`, changefeedbase.OptInitialScan,
			changefeedbase.OptNoInitialScan)
	} else {
		__antithesis_instrumentation__.Notify(15378)
	}
	__antithesis_instrumentation__.Notify(15369)

	if initialScanSet && func() bool {
		__antithesis_instrumentation__.Notify(15379)
		return initialScanOnlySet == true
	}() == true {
		__antithesis_instrumentation__.Notify(15380)
		return changefeedbase.InitialScan, errors.Errorf(
			`cannot specify both %s and %s`, changefeedbase.OptInitialScan,
			changefeedbase.OptInitialScanOnly)
	} else {
		__antithesis_instrumentation__.Notify(15381)
	}
	__antithesis_instrumentation__.Notify(15370)

	if noInitialScanSet && func() bool {
		__antithesis_instrumentation__.Notify(15382)
		return initialScanOnlySet == true
	}() == true {
		__antithesis_instrumentation__.Notify(15383)
		return changefeedbase.InitialScan, errors.Errorf(
			`cannot specify both %s and %s`, changefeedbase.OptInitialScanOnly,
			changefeedbase.OptNoInitialScan)
	} else {
		__antithesis_instrumentation__.Notify(15384)
	}
	__antithesis_instrumentation__.Notify(15371)

	if initialScanSet {
		__antithesis_instrumentation__.Notify(15385)
		const opt = changefeedbase.OptInitialScan
		switch strings.ToLower(initialScanType) {
		case ``, `yes`:
			__antithesis_instrumentation__.Notify(15386)
			return changefeedbase.InitialScan, nil
		case `no`:
			__antithesis_instrumentation__.Notify(15387)
			return changefeedbase.NoInitialScan, nil
		case `only`:
			__antithesis_instrumentation__.Notify(15388)
			return changefeedbase.OnlyInitialScan, nil
		default:
			__antithesis_instrumentation__.Notify(15389)
			return changefeedbase.InitialScan, errors.Errorf(
				`unknown %s: %s`, opt, initialScanType)
		}
	} else {
		__antithesis_instrumentation__.Notify(15390)
	}
	__antithesis_instrumentation__.Notify(15372)

	if initialScanOnlySet {
		__antithesis_instrumentation__.Notify(15391)
		return changefeedbase.OnlyInitialScan, nil
	} else {
		__antithesis_instrumentation__.Notify(15392)
	}
	__antithesis_instrumentation__.Notify(15373)

	if noInitialScanSet {
		__antithesis_instrumentation__.Notify(15393)
		return changefeedbase.NoInitialScan, nil
	} else {
		__antithesis_instrumentation__.Notify(15394)
	}
	__antithesis_instrumentation__.Notify(15374)

	if !cursor {
		__antithesis_instrumentation__.Notify(15395)
		return changefeedbase.InitialScan, nil
	} else {
		__antithesis_instrumentation__.Notify(15396)
	}
	__antithesis_instrumentation__.Notify(15375)

	return changefeedbase.NoInitialScan, nil
}
