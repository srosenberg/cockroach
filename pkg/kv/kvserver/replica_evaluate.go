package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/uncertainty"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/kr/pretty"
)

func optimizePuts(
	reader storage.Reader, origReqs []roachpb.RequestUnion, distinctSpans bool,
) []roachpb.RequestUnion {
	__antithesis_instrumentation__.Notify(117229)
	var minKey, maxKey roachpb.Key
	var unique map[string]struct{}
	if !distinctSpans {
		__antithesis_instrumentation__.Notify(117236)
		unique = make(map[string]struct{}, len(origReqs))
	} else {
		__antithesis_instrumentation__.Notify(117237)
	}
	__antithesis_instrumentation__.Notify(117230)

	maybeAddPut := func(key roachpb.Key) bool {
		__antithesis_instrumentation__.Notify(117238)

		if unique != nil {
			__antithesis_instrumentation__.Notify(117242)
			if _, ok := unique[string(key)]; ok {
				__antithesis_instrumentation__.Notify(117244)
				return false
			} else {
				__antithesis_instrumentation__.Notify(117245)
			}
			__antithesis_instrumentation__.Notify(117243)
			unique[string(key)] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(117246)
		}
		__antithesis_instrumentation__.Notify(117239)
		if minKey == nil || func() bool {
			__antithesis_instrumentation__.Notify(117247)
			return bytes.Compare(key, minKey) < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(117248)
			minKey = key
		} else {
			__antithesis_instrumentation__.Notify(117249)
		}
		__antithesis_instrumentation__.Notify(117240)
		if maxKey == nil || func() bool {
			__antithesis_instrumentation__.Notify(117250)
			return bytes.Compare(key, maxKey) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(117251)
			maxKey = key
		} else {
			__antithesis_instrumentation__.Notify(117252)
		}
		__antithesis_instrumentation__.Notify(117241)
		return true
	}
	__antithesis_instrumentation__.Notify(117231)

	firstUnoptimizedIndex := len(origReqs)
	for i, r := range origReqs {
		__antithesis_instrumentation__.Notify(117253)
		switch t := r.GetInner().(type) {
		case *roachpb.PutRequest:
			__antithesis_instrumentation__.Notify(117255)
			if maybeAddPut(t.Key) {
				__antithesis_instrumentation__.Notify(117258)
				continue
			} else {
				__antithesis_instrumentation__.Notify(117259)
			}
		case *roachpb.ConditionalPutRequest:
			__antithesis_instrumentation__.Notify(117256)
			if maybeAddPut(t.Key) {
				__antithesis_instrumentation__.Notify(117260)
				continue
			} else {
				__antithesis_instrumentation__.Notify(117261)
			}
		case *roachpb.InitPutRequest:
			__antithesis_instrumentation__.Notify(117257)
			if maybeAddPut(t.Key) {
				__antithesis_instrumentation__.Notify(117262)
				continue
			} else {
				__antithesis_instrumentation__.Notify(117263)
			}
		}
		__antithesis_instrumentation__.Notify(117254)
		firstUnoptimizedIndex = i
		break
	}
	__antithesis_instrumentation__.Notify(117232)

	if firstUnoptimizedIndex < optimizePutThreshold {
		__antithesis_instrumentation__.Notify(117264)
		return origReqs
	} else {
		__antithesis_instrumentation__.Notify(117265)
	}
	__antithesis_instrumentation__.Notify(117233)

	iter := reader.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{

		UpperBound: maxKey.Next(),
	})
	defer iter.Close()

	iter.SeekGE(storage.MakeMVCCMetadataKey(minKey))
	var iterKey roachpb.Key
	if ok, err := iter.Valid(); err != nil {
		__antithesis_instrumentation__.Notify(117266)

		log.Errorf(context.TODO(), "Seek returned error; disabling blind-put optimization: %+v", err)
		return origReqs
	} else {
		__antithesis_instrumentation__.Notify(117267)
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(117268)
			return bytes.Compare(iter.Key().Key, maxKey) <= 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(117269)
			iterKey = iter.Key().Key
		} else {
			__antithesis_instrumentation__.Notify(117270)
		}
	}
	__antithesis_instrumentation__.Notify(117234)

	reqs := append([]roachpb.RequestUnion(nil), origReqs...)
	for i := range reqs[:firstUnoptimizedIndex] {
		__antithesis_instrumentation__.Notify(117271)
		inner := reqs[i].GetInner()
		if iterKey == nil || func() bool {
			__antithesis_instrumentation__.Notify(117272)
			return bytes.Compare(iterKey, inner.Header().Key) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(117273)
			switch t := inner.(type) {
			case *roachpb.PutRequest:
				__antithesis_instrumentation__.Notify(117274)
				shallow := *t
				shallow.Blind = true
				reqs[i].MustSetInner(&shallow)
			case *roachpb.ConditionalPutRequest:
				__antithesis_instrumentation__.Notify(117275)
				shallow := *t
				shallow.Blind = true
				reqs[i].MustSetInner(&shallow)
			case *roachpb.InitPutRequest:
				__antithesis_instrumentation__.Notify(117276)
				shallow := *t
				shallow.Blind = true
				reqs[i].MustSetInner(&shallow)
			default:
				__antithesis_instrumentation__.Notify(117277)
				log.Fatalf(context.TODO(), "unexpected non-put request: %s", t)
			}
		} else {
			__antithesis_instrumentation__.Notify(117278)
		}
	}
	__antithesis_instrumentation__.Notify(117235)
	return reqs
}

func evaluateBatch(
	ctx context.Context,
	idKey kvserverbase.CmdIDKey,
	readWriter storage.ReadWriter,
	rec batcheval.EvalContext,
	ms *enginepb.MVCCStats,
	ba *roachpb.BatchRequest,
	ui uncertainty.Interval,
	readOnly bool,
) (_ *roachpb.BatchResponse, _ result.Result, retErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(117279)

	defer func() {
		__antithesis_instrumentation__.Notify(117286)

		if retErr != nil && func() bool {
			__antithesis_instrumentation__.Notify(117287)
			return retErr.GetTxn() != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(117288)
			retErr.GetTxn().WriteTooOld = false
		} else {
			__antithesis_instrumentation__.Notify(117289)
		}
	}()
	__antithesis_instrumentation__.Notify(117280)

	baReqs := ba.Requests

	baHeader := ba.Header

	br := ba.CreateReply()

	if len(baReqs) >= optimizePutThreshold && func() bool {
		__antithesis_instrumentation__.Notify(117290)
		return !readOnly == true
	}() == true {
		__antithesis_instrumentation__.Notify(117291)
		baReqs = optimizePuts(readWriter, baReqs, baHeader.DistinctSpans)
	} else {
		__antithesis_instrumentation__.Notify(117292)
	}
	__antithesis_instrumentation__.Notify(117281)

	if baHeader.Txn != nil {
		__antithesis_instrumentation__.Notify(117293)
		baHeader.Txn = baHeader.Txn.Clone()

		if baHeader.Txn.IsLocking() {
			__antithesis_instrumentation__.Notify(117294)

			if !ba.IsSingleAbortTxnRequest() && func() bool {
				__antithesis_instrumentation__.Notify(117295)
				return !ba.IsSingleHeartbeatTxnRequest() == true
			}() == true {
				__antithesis_instrumentation__.Notify(117296)
				if pErr := checkIfTxnAborted(ctx, rec, readWriter, *baHeader.Txn); pErr != nil {
					__antithesis_instrumentation__.Notify(117297)
					return nil, result.Result{}, pErr
				} else {
					__antithesis_instrumentation__.Notify(117298)
				}
			} else {
				__antithesis_instrumentation__.Notify(117299)
			}
		} else {
			__antithesis_instrumentation__.Notify(117300)
		}
	} else {
		__antithesis_instrumentation__.Notify(117301)
	}
	__antithesis_instrumentation__.Notify(117282)

	var mergedResult result.Result

	var writeTooOldState struct {
		err *roachpb.WriteTooOldError

		cantDeferWTOE bool
	}

	for index, union := range baReqs {
		__antithesis_instrumentation__.Notify(117302)

		args := union.GetInner()

		if baHeader.Txn != nil {
			__antithesis_instrumentation__.Notify(117311)

			baHeader.Txn.Sequence = args.Header().Sequence
		} else {
			__antithesis_instrumentation__.Notify(117312)
		}
		__antithesis_instrumentation__.Notify(117303)

		if filter := rec.EvalKnobs().TestingEvalFilter; filter != nil {
			__antithesis_instrumentation__.Notify(117313)
			filterArgs := kvserverbase.FilterArgs{
				Ctx:     ctx,
				CmdID:   idKey,
				Index:   index,
				Sid:     rec.StoreID(),
				Req:     args,
				Version: rec.ClusterSettings().Version.ActiveVersionOrEmpty(ctx).Version,
				Hdr:     baHeader,
			}
			if pErr := filter(filterArgs); pErr != nil {
				__antithesis_instrumentation__.Notify(117314)
				if pErr.GetTxn() == nil {
					__antithesis_instrumentation__.Notify(117316)
					pErr.SetTxn(baHeader.Txn)
				} else {
					__antithesis_instrumentation__.Notify(117317)
				}
				__antithesis_instrumentation__.Notify(117315)
				log.Infof(ctx, "test injecting error: %s", pErr)
				return nil, result.Result{}, pErr
			} else {
				__antithesis_instrumentation__.Notify(117318)
			}
		} else {
			__antithesis_instrumentation__.Notify(117319)
		}
		__antithesis_instrumentation__.Notify(117304)

		reply := br.Responses[index].GetInner()

		curResult, err := evaluateCommand(
			ctx, readWriter, rec, ms, baHeader, args, reply, ui)

		if filter := rec.EvalKnobs().TestingPostEvalFilter; filter != nil {
			__antithesis_instrumentation__.Notify(117320)
			filterArgs := kvserverbase.FilterArgs{
				Ctx:   ctx,
				CmdID: idKey,
				Index: index,
				Sid:   rec.StoreID(),
				Req:   args,
				Hdr:   baHeader,
				Err:   err,
			}
			if pErr := filter(filterArgs); pErr != nil {
				__antithesis_instrumentation__.Notify(117321)
				if pErr.GetTxn() == nil {
					__antithesis_instrumentation__.Notify(117323)
					pErr.SetTxn(baHeader.Txn)
				} else {
					__antithesis_instrumentation__.Notify(117324)
				}
				__antithesis_instrumentation__.Notify(117322)
				log.Infof(ctx, "test injecting error: %s", pErr)
				return nil, result.Result{}, pErr
			} else {
				__antithesis_instrumentation__.Notify(117325)
			}
		} else {
			__antithesis_instrumentation__.Notify(117326)
		}
		__antithesis_instrumentation__.Notify(117305)

		if headerCopy := reply.Header(); baHeader.Txn != nil && func() bool {
			__antithesis_instrumentation__.Notify(117327)
			return headerCopy.Txn != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(117328)
			baHeader.Txn.Update(headerCopy.Txn)
			headerCopy.Txn = nil
			reply.SetHeader(headerCopy)
		} else {
			__antithesis_instrumentation__.Notify(117329)
		}
		__antithesis_instrumentation__.Notify(117306)

		if err != nil {
			__antithesis_instrumentation__.Notify(117330)

			if retErr := (*roachpb.TransactionRetryError)(nil); errors.As(err, &retErr) && func() bool {
				__antithesis_instrumentation__.Notify(117331)
				return retErr.Reason == roachpb.RETRY_WRITE_TOO_OLD == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(117332)
				return args.Method() == roachpb.EndTxn == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(117333)
				return writeTooOldState.err != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(117334)
				err = writeTooOldState.err

				writeTooOldState.cantDeferWTOE = true
			} else {
				__antithesis_instrumentation__.Notify(117335)
				if wtoErr := (*roachpb.WriteTooOldError)(nil); errors.As(err, &wtoErr) {
					__antithesis_instrumentation__.Notify(117336)

					if writeTooOldState.err != nil {
						__antithesis_instrumentation__.Notify(117340)
						writeTooOldState.err.ActualTimestamp.Forward(
							wtoErr.ActualTimestamp)
					} else {
						__antithesis_instrumentation__.Notify(117341)
						writeTooOldState.err = wtoErr
					}
					__antithesis_instrumentation__.Notify(117337)

					if !roachpb.IsBlindWrite(args) {
						__antithesis_instrumentation__.Notify(117342)
						writeTooOldState.cantDeferWTOE = true
					} else {
						__antithesis_instrumentation__.Notify(117343)
					}
					__antithesis_instrumentation__.Notify(117338)

					if baHeader.Txn != nil {
						__antithesis_instrumentation__.Notify(117344)
						log.VEventf(ctx, 2, "setting WriteTooOld because of key: %s. wts: %s -> %s",
							args.Header().Key, baHeader.Txn.WriteTimestamp, wtoErr.ActualTimestamp)
						baHeader.Txn.WriteTimestamp.Forward(wtoErr.ActualTimestamp)
						baHeader.Txn.WriteTooOld = true
					} else {
						__antithesis_instrumentation__.Notify(117345)

						writeTooOldState.cantDeferWTOE = true
					}
					__antithesis_instrumentation__.Notify(117339)

					err = nil
				} else {
					__antithesis_instrumentation__.Notify(117346)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(117347)
		}
		__antithesis_instrumentation__.Notify(117307)

		if err := mergedResult.MergeAndDestroy(curResult); err != nil {
			__antithesis_instrumentation__.Notify(117348)
			log.Fatalf(
				ctx,
				"unable to absorb Result: %s\ndiff(new, old): %s",
				err, pretty.Diff(curResult, mergedResult),
			)
		} else {
			__antithesis_instrumentation__.Notify(117349)
		}
		__antithesis_instrumentation__.Notify(117308)

		if err != nil {
			__antithesis_instrumentation__.Notify(117350)
			pErr := roachpb.NewErrorWithTxn(err, baHeader.Txn)

			pErr.SetErrorIndex(int32(index))

			return nil, mergedResult, pErr
		} else {
			__antithesis_instrumentation__.Notify(117351)
		}
		__antithesis_instrumentation__.Notify(117309)

		h := reply.Header()
		if limit, retResults := baHeader.MaxSpanRequestKeys, h.NumKeys; limit != 0 && func() bool {
			__antithesis_instrumentation__.Notify(117352)
			return retResults > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(117353)
			if retResults > limit {
				__antithesis_instrumentation__.Notify(117354)
				index, retResults, limit := index, retResults, limit
				return nil, mergedResult, roachpb.NewError(errors.AssertionFailedf(
					"received %d results, limit was %d (original limit: %d, batch=%s idx=%d)",
					retResults, limit, ba.Header.MaxSpanRequestKeys,
					redact.Safe(ba.Summary()), index))
			} else {
				__antithesis_instrumentation__.Notify(117355)
				if retResults < limit {
					__antithesis_instrumentation__.Notify(117356)
					baHeader.MaxSpanRequestKeys -= retResults
				} else {
					__antithesis_instrumentation__.Notify(117357)

					baHeader.MaxSpanRequestKeys = -1
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(117358)
		}
		__antithesis_instrumentation__.Notify(117310)

		if baHeader.TargetBytes > 0 {
			__antithesis_instrumentation__.Notify(117359)
			if h.ResumeReason == roachpb.RESUME_BYTE_LIMIT {
				__antithesis_instrumentation__.Notify(117360)
				baHeader.TargetBytes = -1
			} else {
				__antithesis_instrumentation__.Notify(117361)
				if baHeader.TargetBytes > h.NumBytes {
					__antithesis_instrumentation__.Notify(117362)
					baHeader.TargetBytes -= h.NumBytes
				} else {
					__antithesis_instrumentation__.Notify(117363)
					baHeader.TargetBytes = -1
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(117364)
		}
	}
	__antithesis_instrumentation__.Notify(117283)

	if writeTooOldState.cantDeferWTOE {
		__antithesis_instrumentation__.Notify(117365)

		return nil, mergedResult, roachpb.NewErrorWithTxn(writeTooOldState.err, baHeader.Txn)
	} else {
		__antithesis_instrumentation__.Notify(117366)
	}
	__antithesis_instrumentation__.Notify(117284)

	if baHeader.Txn != nil {
		__antithesis_instrumentation__.Notify(117367)

		br.Txn = baHeader.Txn

		if br.Txn.ReadTimestamp.Less(baHeader.Timestamp) {
			__antithesis_instrumentation__.Notify(117369)
			log.Fatalf(ctx, "br.Txn.ReadTimestamp < ba.Timestamp (%s < %s). ba: %s",
				br.Txn.ReadTimestamp, baHeader.Timestamp, ba)
		} else {
			__antithesis_instrumentation__.Notify(117370)
		}
		__antithesis_instrumentation__.Notify(117368)
		br.Timestamp = br.Txn.ReadTimestamp
	} else {
		__antithesis_instrumentation__.Notify(117371)
		br.Timestamp = baHeader.Timestamp
	}
	__antithesis_instrumentation__.Notify(117285)

	return br, mergedResult, nil
}

func evaluateCommand(
	ctx context.Context,
	readWriter storage.ReadWriter,
	rec batcheval.EvalContext,
	ms *enginepb.MVCCStats,
	h roachpb.Header,
	args roachpb.Request,
	reply roachpb.Response,
	ui uncertainty.Interval,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(117372)
	var err error
	var pd result.Result

	if cmd, ok := batcheval.LookupCommand(args.Method()); ok {
		__antithesis_instrumentation__.Notify(117375)
		cArgs := batcheval.CommandArgs{
			EvalCtx:     rec,
			Header:      h,
			Args:        args,
			Stats:       ms,
			Uncertainty: ui,
		}

		if cmd.EvalRW != nil {
			__antithesis_instrumentation__.Notify(117376)
			pd, err = cmd.EvalRW(ctx, readWriter, cArgs, reply)
		} else {
			__antithesis_instrumentation__.Notify(117377)
			pd, err = cmd.EvalRO(ctx, readWriter, cArgs, reply)
		}
	} else {
		__antithesis_instrumentation__.Notify(117378)
		return result.Result{}, errors.Errorf("unrecognized command %s", args.Method())
	}
	__antithesis_instrumentation__.Notify(117373)

	if log.ExpensiveLogEnabled(ctx, 2) {
		__antithesis_instrumentation__.Notify(117379)
		trunc := func(s string) string {
			__antithesis_instrumentation__.Notify(117381)
			const maxLen = 256
			if len(s) > maxLen {
				__antithesis_instrumentation__.Notify(117383)
				return s[:maxLen-3] + "..."
			} else {
				__antithesis_instrumentation__.Notify(117384)
			}
			__antithesis_instrumentation__.Notify(117382)
			return s
		}
		__antithesis_instrumentation__.Notify(117380)
		log.VEventf(ctx, 2, "evaluated %s command %s, txn=%v : resp=%s, err=%v",
			args.Method(), trunc(args.String()), h.Txn, trunc(reply.String()), err)
	} else {
		__antithesis_instrumentation__.Notify(117385)
	}
	__antithesis_instrumentation__.Notify(117374)
	return pd, err
}

func canDoServersideRetry(
	ctx context.Context,
	pErr *roachpb.Error,
	ba *roachpb.BatchRequest,
	br *roachpb.BatchResponse,
	g *concurrency.Guard,
	deadline *hlc.Timestamp,
) bool {
	__antithesis_instrumentation__.Notify(117386)
	if ba.Txn != nil {
		__antithesis_instrumentation__.Notify(117390)
		if !ba.CanForwardReadTimestamp {
			__antithesis_instrumentation__.Notify(117393)
			return false
		} else {
			__antithesis_instrumentation__.Notify(117394)
		}
		__antithesis_instrumentation__.Notify(117391)
		if deadline != nil {
			__antithesis_instrumentation__.Notify(117395)
			log.Fatal(ctx, "deadline passed for transactional request")
		} else {
			__antithesis_instrumentation__.Notify(117396)
		}
		__antithesis_instrumentation__.Notify(117392)
		if etArg, ok := ba.GetArg(roachpb.EndTxn); ok {
			__antithesis_instrumentation__.Notify(117397)
			et := etArg.(*roachpb.EndTxnRequest)
			deadline = et.Deadline
		} else {
			__antithesis_instrumentation__.Notify(117398)
		}
	} else {
		__antithesis_instrumentation__.Notify(117399)
	}
	__antithesis_instrumentation__.Notify(117387)

	var newTimestamp hlc.Timestamp
	if ba.Txn != nil {
		__antithesis_instrumentation__.Notify(117400)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(117401)
			var ok bool
			ok, newTimestamp = roachpb.TransactionRefreshTimestamp(pErr)
			if !ok {
				__antithesis_instrumentation__.Notify(117402)
				return false
			} else {
				__antithesis_instrumentation__.Notify(117403)
			}
		} else {
			__antithesis_instrumentation__.Notify(117404)
			if !br.Txn.WriteTooOld {
				__antithesis_instrumentation__.Notify(117406)
				log.Fatalf(ctx, "expected the WriteTooOld flag to be set")
			} else {
				__antithesis_instrumentation__.Notify(117407)
			}
			__antithesis_instrumentation__.Notify(117405)
			newTimestamp = br.Txn.WriteTimestamp
		}
	} else {
		__antithesis_instrumentation__.Notify(117408)
		if pErr == nil {
			__antithesis_instrumentation__.Notify(117410)
			log.Fatalf(ctx, "canDoServersideRetry called for non-txn request without error")
		} else {
			__antithesis_instrumentation__.Notify(117411)
		}
		__antithesis_instrumentation__.Notify(117409)
		switch tErr := pErr.GetDetail().(type) {
		case *roachpb.WriteTooOldError:
			__antithesis_instrumentation__.Notify(117412)
			newTimestamp = tErr.RetryTimestamp()

		case *roachpb.ReadWithinUncertaintyIntervalError:
			__antithesis_instrumentation__.Notify(117413)
			newTimestamp = tErr.RetryTimestamp()

		default:
			__antithesis_instrumentation__.Notify(117414)
			return false
		}
	}
	__antithesis_instrumentation__.Notify(117388)

	if batcheval.IsEndTxnExceedingDeadline(newTimestamp, deadline) {
		__antithesis_instrumentation__.Notify(117415)
		return false
	} else {
		__antithesis_instrumentation__.Notify(117416)
	}
	__antithesis_instrumentation__.Notify(117389)
	return tryBumpBatchTimestamp(ctx, ba, g, newTimestamp)
}
