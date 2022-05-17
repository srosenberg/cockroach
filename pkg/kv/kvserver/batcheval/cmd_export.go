package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/types"
)

const SSTTargetSizeSetting = "kv.bulk_sst.target_size"

var ExportRequestTargetFileSize = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	SSTTargetSizeSetting,
	fmt.Sprintf("target size for SSTs emitted from export requests; "+
		"export requests (i.e. BACKUP) may buffer up to the sum of %s and %s in memory",
		SSTTargetSizeSetting, MaxExportOverageSetting,
	),
	16<<20,
).WithPublic()

const MaxExportOverageSetting = "kv.bulk_sst.max_allowed_overage"

var ExportRequestMaxAllowedFileSizeOverage = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	MaxExportOverageSetting,
	fmt.Sprintf("if positive, allowed size in excess of target size for SSTs from export requests; "+
		"export requests (i.e. BACKUP) may buffer up to the sum of %s and %s in memory",
		SSTTargetSizeSetting, MaxExportOverageSetting,
	),
	64<<20,
).WithPublic()

var exportRequestMaxIterationTime = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.bulk_sst.max_request_time",
	"if set, limits amount of time spent in export requests; "+
		"if export request can not finish within allocated time it will resume from the point it stopped in "+
		"subsequent request",

	0,
)

func init() {
	RegisterReadOnlyCommand(roachpb.Export, declareKeysExport, evalExport)
}

func declareKeysExport(
	rs ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, lockSpans *spanset.SpanSet,
	maxOffset time.Duration,
) {
	__antithesis_instrumentation__.Notify(96814)
	DefaultDeclareIsolatedKeys(rs, header, req, latchSpans, lockSpans, maxOffset)
	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeGCThresholdKey(header.RangeID)})
}

func evalExport(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96815)
	args := cArgs.Args.(*roachpb.ExportRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.ExportResponse)

	ctx, evalExportSpan := tracing.ChildSpan(ctx, fmt.Sprintf("Export [%s,%s)", args.Key, args.EndKey))
	defer evalExportSpan.Finish()

	var evalExportTrace types.StringValue
	if cArgs.EvalCtx.NodeID() == h.GatewayNodeID {
		__antithesis_instrumentation__.Notify(96827)
		evalExportTrace.Value = fmt.Sprintf("evaluating Export on gateway node %d", cArgs.EvalCtx.NodeID())
	} else {
		__antithesis_instrumentation__.Notify(96828)
		evalExportTrace.Value = fmt.Sprintf("evaluating Export on remote node %d", cArgs.EvalCtx.NodeID())
	}
	__antithesis_instrumentation__.Notify(96816)
	evalExportSpan.RecordStructured(&evalExportTrace)

	if cArgs.EvalCtx.ExcludeDataFromBackup() {
		__antithesis_instrumentation__.Notify(96829)
		log.Infof(ctx, "[%s, %s) is part of a table excluded from backup, returning empty ExportResponse", args.Key, args.EndKey)
		return result.Result{}, nil
	} else {
		__antithesis_instrumentation__.Notify(96830)
	}
	__antithesis_instrumentation__.Notify(96817)

	if !args.ReturnSST {
		__antithesis_instrumentation__.Notify(96831)
		return result.Result{}, errors.New("ReturnSST is required")
	} else {
		__antithesis_instrumentation__.Notify(96832)
	}
	__antithesis_instrumentation__.Notify(96818)

	if args.Encryption != nil {
		__antithesis_instrumentation__.Notify(96833)
		return result.Result{}, errors.New("returned SSTs cannot be encrypted")
	} else {
		__antithesis_instrumentation__.Notify(96834)
	}
	__antithesis_instrumentation__.Notify(96819)

	if args.MVCCFilter == roachpb.MVCCFilter_All {
		__antithesis_instrumentation__.Notify(96835)
		reply.StartTime = cArgs.EvalCtx.GetGCThreshold()
	} else {
		__antithesis_instrumentation__.Notify(96836)
	}
	__antithesis_instrumentation__.Notify(96820)

	var exportAllRevisions bool
	switch args.MVCCFilter {
	case roachpb.MVCCFilter_Latest:
		__antithesis_instrumentation__.Notify(96837)
		exportAllRevisions = false
	case roachpb.MVCCFilter_All:
		__antithesis_instrumentation__.Notify(96838)
		exportAllRevisions = true
	default:
		__antithesis_instrumentation__.Notify(96839)
		return result.Result{}, errors.Errorf("unknown MVCC filter: %s", args.MVCCFilter)
	}
	__antithesis_instrumentation__.Notify(96821)

	targetSize := uint64(args.TargetFileSize)

	clusterSettingTargetSize := uint64(ExportRequestTargetFileSize.Get(&cArgs.EvalCtx.ClusterSettings().SV))
	if targetSize > clusterSettingTargetSize {
		__antithesis_instrumentation__.Notify(96840)
		targetSize = clusterSettingTargetSize
	} else {
		__antithesis_instrumentation__.Notify(96841)
	}
	__antithesis_instrumentation__.Notify(96822)

	var maxSize uint64
	allowedOverage := ExportRequestMaxAllowedFileSizeOverage.Get(&cArgs.EvalCtx.ClusterSettings().SV)
	if targetSize > 0 && func() bool {
		__antithesis_instrumentation__.Notify(96842)
		return allowedOverage > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(96843)
		maxSize = targetSize + uint64(allowedOverage)
	} else {
		__antithesis_instrumentation__.Notify(96844)
	}
	__antithesis_instrumentation__.Notify(96823)

	maxRunTime := exportRequestMaxIterationTime.Get(&cArgs.EvalCtx.ClusterSettings().SV)

	var maxIntents uint64
	if m := storage.MaxIntentsPerWriteIntentError.Get(&cArgs.EvalCtx.ClusterSettings().SV); m > 0 {
		__antithesis_instrumentation__.Notify(96845)
		maxIntents = uint64(m)
	} else {
		__antithesis_instrumentation__.Notify(96846)
	}
	__antithesis_instrumentation__.Notify(96824)

	useTBI := args.EnableTimeBoundIteratorOptimization && func() bool {
		__antithesis_instrumentation__.Notify(96847)
		return !args.StartTime.IsEmpty() == true
	}() == true

	resumeKeyTS := hlc.Timestamp{}
	if args.SplitMidKey {
		__antithesis_instrumentation__.Notify(96848)
		resumeKeyTS = args.ResumeKeyTS
	} else {
		__antithesis_instrumentation__.Notify(96849)
	}
	__antithesis_instrumentation__.Notify(96825)

	var curSizeOfExportedSSTs int64
	for start := args.Key; start != nil; {
		__antithesis_instrumentation__.Notify(96850)
		destFile := &storage.MemFile{}
		summary, resume, resumeTS, err := reader.ExportMVCCToSst(ctx, storage.ExportOptions{
			StartKey:           storage.MVCCKey{Key: start, Timestamp: resumeKeyTS},
			EndKey:             args.EndKey,
			StartTS:            args.StartTime,
			EndTS:              h.Timestamp,
			ExportAllRevisions: exportAllRevisions,
			TargetSize:         targetSize,
			MaxSize:            maxSize,
			MaxIntents:         maxIntents,
			StopMidKey:         args.SplitMidKey,
			UseTBI:             useTBI,
			ResourceLimiter:    storage.NewResourceLimiter(storage.ResourceLimiterOptions{MaxRunTime: maxRunTime}, timeutil.DefaultTimeSource{}),
		}, destFile)
		if err != nil {
			__antithesis_instrumentation__.Notify(96854)
			if errors.HasType(err, (*storage.ExceedMaxSizeError)(nil)) {
				__antithesis_instrumentation__.Notify(96856)
				err = errors.WithHintf(err,
					"consider increasing cluster setting %q", MaxExportOverageSetting)
			} else {
				__antithesis_instrumentation__.Notify(96857)
			}
			__antithesis_instrumentation__.Notify(96855)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96858)
		}
		__antithesis_instrumentation__.Notify(96851)
		data := destFile.Data()

		if summary.DataSize == 0 {
			__antithesis_instrumentation__.Notify(96859)
			break
		} else {
			__antithesis_instrumentation__.Notify(96860)
		}
		__antithesis_instrumentation__.Notify(96852)

		span := roachpb.Span{Key: start}
		if resume != nil {
			__antithesis_instrumentation__.Notify(96861)
			span.EndKey = resume
		} else {
			__antithesis_instrumentation__.Notify(96862)
			span.EndKey = args.EndKey
		}
		__antithesis_instrumentation__.Notify(96853)
		exported := roachpb.ExportResponse_File{
			Span:     span,
			EndKeyTS: resumeTS,
			Exported: summary,
			SST:      data,
		}
		reply.Files = append(reply.Files, exported)
		start = resume
		resumeKeyTS = resumeTS

		if h.TargetBytes > 0 {
			__antithesis_instrumentation__.Notify(96863)
			curSizeOfExportedSSTs += summary.DataSize

			targetSize := h.TargetBytes
			if curSizeOfExportedSSTs < targetSize {
				__antithesis_instrumentation__.Notify(96865)
				targetSize = curSizeOfExportedSSTs
			} else {
				__antithesis_instrumentation__.Notify(96866)
			}
			__antithesis_instrumentation__.Notify(96864)
			reply.NumBytes = targetSize

			if reply.NumBytes == h.TargetBytes {
				__antithesis_instrumentation__.Notify(96867)
				if resume != nil {
					__antithesis_instrumentation__.Notify(96869)
					reply.ResumeSpan = &roachpb.Span{
						Key:    resume,
						EndKey: args.EndKey,
					}
					reply.ResumeReason = roachpb.RESUME_BYTE_LIMIT
				} else {
					__antithesis_instrumentation__.Notify(96870)
				}
				__antithesis_instrumentation__.Notify(96868)
				break
			} else {
				__antithesis_instrumentation__.Notify(96871)
			}
		} else {
			__antithesis_instrumentation__.Notify(96872)
		}
	}
	__antithesis_instrumentation__.Notify(96826)

	return result.Result{}, nil
}
