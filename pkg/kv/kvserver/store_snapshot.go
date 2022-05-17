package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/raftentry"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rditer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

const (
	storeDrainingMsg = "store is draining"

	IntersectingSnapshotMsg = "snapshot intersects existing range"
)

type incomingSnapshotStream interface {
	Send(*kvserverpb.SnapshotResponse) error
	Recv() (*kvserverpb.SnapshotRequest, error)
}

type outgoingSnapshotStream interface {
	Send(*kvserverpb.SnapshotRequest) error
	Recv() (*kvserverpb.SnapshotResponse, error)
}

type snapshotStrategy interface {
	Receive(
		context.Context,
		incomingSnapshotStream,
		kvserverpb.SnapshotRequest_Header,
		*metric.Counter,
	) (IncomingSnapshot, error)

	Send(
		context.Context,
		outgoingSnapshotStream,
		kvserverpb.SnapshotRequest_Header,
		*OutgoingSnapshot,
		*metric.Counter,
	) (int64, error)

	Status() redact.RedactableString

	Close(context.Context)
}

func assertStrategy(
	ctx context.Context,
	header kvserverpb.SnapshotRequest_Header,
	expect kvserverpb.SnapshotRequest_Strategy,
) {
	__antithesis_instrumentation__.Notify(126015)
	if header.Strategy != expect {
		__antithesis_instrumentation__.Notify(126016)
		log.Fatalf(ctx, "expected strategy %s, found strategy %s", expect, header.Strategy)
	} else {
		__antithesis_instrumentation__.Notify(126017)
	}
}

type kvBatchSnapshotStrategy struct {
	status redact.RedactableString

	batchSize int64

	limiter *rate.Limiter

	newBatch func() storage.Batch

	sstChunkSize int64

	scratch *SSTSnapshotStorageScratch
	st      *cluster.Settings
}

type multiSSTWriter struct {
	st        *cluster.Settings
	scratch   *SSTSnapshotStorageScratch
	currSST   storage.SSTWriter
	keyRanges []rditer.KeyRange
	currRange int

	sstChunkSize int64

	dataSize int64
}

func newMultiSSTWriter(
	ctx context.Context,
	st *cluster.Settings,
	scratch *SSTSnapshotStorageScratch,
	keyRanges []rditer.KeyRange,
	sstChunkSize int64,
) (multiSSTWriter, error) {
	__antithesis_instrumentation__.Notify(126018)
	msstw := multiSSTWriter{
		st:           st,
		scratch:      scratch,
		keyRanges:    keyRanges,
		sstChunkSize: sstChunkSize,
	}
	if err := msstw.initSST(ctx); err != nil {
		__antithesis_instrumentation__.Notify(126020)
		return msstw, err
	} else {
		__antithesis_instrumentation__.Notify(126021)
	}
	__antithesis_instrumentation__.Notify(126019)
	return msstw, nil
}

func (msstw *multiSSTWriter) initSST(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(126022)
	newSSTFile, err := msstw.scratch.NewFile(ctx, msstw.sstChunkSize)
	if err != nil {
		__antithesis_instrumentation__.Notify(126025)
		return errors.Wrap(err, "failed to create new sst file")
	} else {
		__antithesis_instrumentation__.Notify(126026)
	}
	__antithesis_instrumentation__.Notify(126023)
	newSST := storage.MakeIngestionSSTWriter(ctx, msstw.st, newSSTFile)
	msstw.currSST = newSST
	if err := msstw.currSST.ClearRawRange(
		msstw.keyRanges[msstw.currRange].Start, msstw.keyRanges[msstw.currRange].End); err != nil {
		__antithesis_instrumentation__.Notify(126027)
		msstw.currSST.Close()
		return errors.Wrap(err, "failed to clear range on sst file writer")
	} else {
		__antithesis_instrumentation__.Notify(126028)
	}
	__antithesis_instrumentation__.Notify(126024)
	return nil
}

func (msstw *multiSSTWriter) finalizeSST(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(126029)
	err := msstw.currSST.Finish()
	if err != nil {
		__antithesis_instrumentation__.Notify(126031)
		return errors.Wrap(err, "failed to finish sst")
	} else {
		__antithesis_instrumentation__.Notify(126032)
	}
	__antithesis_instrumentation__.Notify(126030)
	msstw.dataSize += msstw.currSST.DataSize
	msstw.currRange++
	msstw.currSST.Close()
	return nil
}

func (msstw *multiSSTWriter) Put(ctx context.Context, key storage.EngineKey, value []byte) error {
	__antithesis_instrumentation__.Notify(126033)
	for msstw.keyRanges[msstw.currRange].End.Compare(key.Key) <= 0 {
		__antithesis_instrumentation__.Notify(126037)

		if err := msstw.finalizeSST(ctx); err != nil {
			__antithesis_instrumentation__.Notify(126039)
			return err
		} else {
			__antithesis_instrumentation__.Notify(126040)
		}
		__antithesis_instrumentation__.Notify(126038)
		if err := msstw.initSST(ctx); err != nil {
			__antithesis_instrumentation__.Notify(126041)
			return err
		} else {
			__antithesis_instrumentation__.Notify(126042)
		}
	}
	__antithesis_instrumentation__.Notify(126034)
	if msstw.keyRanges[msstw.currRange].Start.Compare(key.Key) > 0 {
		__antithesis_instrumentation__.Notify(126043)
		return errors.AssertionFailedf("client error: expected %s to fall in one of %s", key.Key, msstw.keyRanges)
	} else {
		__antithesis_instrumentation__.Notify(126044)
	}
	__antithesis_instrumentation__.Notify(126035)
	if err := msstw.currSST.PutEngineKey(key, value); err != nil {
		__antithesis_instrumentation__.Notify(126045)
		return errors.Wrap(err, "failed to put in sst")
	} else {
		__antithesis_instrumentation__.Notify(126046)
	}
	__antithesis_instrumentation__.Notify(126036)
	return nil
}

func (msstw *multiSSTWriter) Finish(ctx context.Context) (int64, error) {
	__antithesis_instrumentation__.Notify(126047)
	if msstw.currRange < len(msstw.keyRanges) {
		__antithesis_instrumentation__.Notify(126049)
		for {
			__antithesis_instrumentation__.Notify(126050)
			if err := msstw.finalizeSST(ctx); err != nil {
				__antithesis_instrumentation__.Notify(126053)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(126054)
			}
			__antithesis_instrumentation__.Notify(126051)
			if msstw.currRange >= len(msstw.keyRanges) {
				__antithesis_instrumentation__.Notify(126055)
				break
			} else {
				__antithesis_instrumentation__.Notify(126056)
			}
			__antithesis_instrumentation__.Notify(126052)
			if err := msstw.initSST(ctx); err != nil {
				__antithesis_instrumentation__.Notify(126057)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(126058)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(126059)
	}
	__antithesis_instrumentation__.Notify(126048)
	return msstw.dataSize, nil
}

func (msstw *multiSSTWriter) Close() {
	__antithesis_instrumentation__.Notify(126060)
	msstw.currSST.Close()
}

func (kvSS *kvBatchSnapshotStrategy) Receive(
	ctx context.Context,
	stream incomingSnapshotStream,
	header kvserverpb.SnapshotRequest_Header,
	bytesRcvdCounter *metric.Counter,
) (IncomingSnapshot, error) {
	__antithesis_instrumentation__.Notify(126061)
	assertStrategy(ctx, header, kvserverpb.SnapshotRequest_KV_BATCH)

	keyRanges := rditer.MakeReplicatedKeyRanges(header.State.Desc)
	msstw, err := newMultiSSTWriter(ctx, kvSS.st, kvSS.scratch, keyRanges, kvSS.sstChunkSize)
	if err != nil {
		__antithesis_instrumentation__.Notify(126063)
		return noSnap, err
	} else {
		__antithesis_instrumentation__.Notify(126064)
	}
	__antithesis_instrumentation__.Notify(126062)
	defer msstw.Close()

	for {
		__antithesis_instrumentation__.Notify(126065)
		req, err := stream.Recv()
		if err != nil {
			__antithesis_instrumentation__.Notify(126069)
			return noSnap, err
		} else {
			__antithesis_instrumentation__.Notify(126070)
		}
		__antithesis_instrumentation__.Notify(126066)
		if req.Header != nil {
			__antithesis_instrumentation__.Notify(126071)
			err := errors.New("client error: provided a header mid-stream")
			return noSnap, sendSnapshotError(stream, err)
		} else {
			__antithesis_instrumentation__.Notify(126072)
		}
		__antithesis_instrumentation__.Notify(126067)

		if req.KVBatch != nil {
			__antithesis_instrumentation__.Notify(126073)
			bytesRcvdCounter.Inc(int64(len(req.KVBatch)))
			batchReader, err := storage.NewRocksDBBatchReader(req.KVBatch)
			if err != nil {
				__antithesis_instrumentation__.Notify(126075)
				return noSnap, errors.Wrap(err, "failed to decode batch")
			} else {
				__antithesis_instrumentation__.Notify(126076)
			}
			__antithesis_instrumentation__.Notify(126074)

			for batchReader.Next() {
				__antithesis_instrumentation__.Notify(126077)
				if batchReader.BatchType() != storage.BatchTypeValue {
					__antithesis_instrumentation__.Notify(126080)
					return noSnap, errors.AssertionFailedf("expected type %d, found type %d", storage.BatchTypeValue, batchReader.BatchType())
				} else {
					__antithesis_instrumentation__.Notify(126081)
				}
				__antithesis_instrumentation__.Notify(126078)
				key, err := batchReader.EngineKey()
				if err != nil {
					__antithesis_instrumentation__.Notify(126082)
					return noSnap, errors.Wrap(err, "failed to decode mvcc key")
				} else {
					__antithesis_instrumentation__.Notify(126083)
				}
				__antithesis_instrumentation__.Notify(126079)
				if err := msstw.Put(ctx, key, batchReader.Value()); err != nil {
					__antithesis_instrumentation__.Notify(126084)
					return noSnap, errors.Wrapf(err, "writing sst for raft snapshot")
				} else {
					__antithesis_instrumentation__.Notify(126085)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(126086)
		}
		__antithesis_instrumentation__.Notify(126068)
		if req.Final {
			__antithesis_instrumentation__.Notify(126087)

			dataSize, err := msstw.Finish(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(126090)
				return noSnap, errors.Wrapf(err, "finishing sst for raft snapshot")
			} else {
				__antithesis_instrumentation__.Notify(126091)
			}
			__antithesis_instrumentation__.Notify(126088)
			msstw.Close()

			snapUUID, err := uuid.FromBytes(header.RaftMessageRequest.Message.Snapshot.Data)
			if err != nil {
				__antithesis_instrumentation__.Notify(126092)
				err = errors.Wrap(err, "client error: invalid snapshot")
				return noSnap, sendSnapshotError(stream, err)
			} else {
				__antithesis_instrumentation__.Notify(126093)
			}
			__antithesis_instrumentation__.Notify(126089)

			inSnap := IncomingSnapshot{
				SnapUUID:          snapUUID,
				SSTStorageScratch: kvSS.scratch,
				FromReplica:       header.RaftMessageRequest.FromReplica,
				Desc:              header.State.Desc,
				DataSize:          dataSize,
				snapType:          header.Type,
				raftAppliedIndex:  header.State.RaftAppliedIndex,
			}

			kvSS.status = redact.Sprintf("ssts: %d", len(kvSS.scratch.SSTs()))
			return inSnap, nil
		} else {
			__antithesis_instrumentation__.Notify(126094)
		}
	}
}

var errMalformedSnapshot = errors.New("malformed snapshot generated")

func (kvSS *kvBatchSnapshotStrategy) Send(
	ctx context.Context,
	stream outgoingSnapshotStream,
	header kvserverpb.SnapshotRequest_Header,
	snap *OutgoingSnapshot,
	bytesSentMetric *metric.Counter,
) (int64, error) {
	__antithesis_instrumentation__.Notify(126095)
	assertStrategy(ctx, header, kvserverpb.SnapshotRequest_KV_BATCH)

	bytesSent := int64(0)

	kvs := 0
	var b storage.Batch
	defer func() {
		__antithesis_instrumentation__.Notify(126099)
		if b != nil {
			__antithesis_instrumentation__.Notify(126100)
			b.Close()
		} else {
			__antithesis_instrumentation__.Notify(126101)
		}
	}()
	__antithesis_instrumentation__.Notify(126096)
	for iter := snap.Iter; ; iter.Next() {
		__antithesis_instrumentation__.Notify(126102)
		if ok, err := iter.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(126106)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(126107)
			if !ok {
				__antithesis_instrumentation__.Notify(126108)
				break
			} else {
				__antithesis_instrumentation__.Notify(126109)
			}
		}
		__antithesis_instrumentation__.Notify(126103)
		kvs++
		unsafeKey := iter.UnsafeKey()
		unsafeValue := iter.UnsafeValue()
		if b == nil {
			__antithesis_instrumentation__.Notify(126110)
			b = kvSS.newBatch()
		} else {
			__antithesis_instrumentation__.Notify(126111)
		}
		__antithesis_instrumentation__.Notify(126104)
		if err := b.PutEngineKey(unsafeKey, unsafeValue); err != nil {
			__antithesis_instrumentation__.Notify(126112)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(126113)
		}
		__antithesis_instrumentation__.Notify(126105)

		if bLen := int64(b.Len()); bLen >= kvSS.batchSize {
			__antithesis_instrumentation__.Notify(126114)
			if err := kvSS.sendBatch(ctx, stream, b); err != nil {
				__antithesis_instrumentation__.Notify(126116)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(126117)
			}
			__antithesis_instrumentation__.Notify(126115)
			bytesSent += bLen
			bytesSentMetric.Inc(bLen)
			b.Close()
			b = nil
		} else {
			__antithesis_instrumentation__.Notify(126118)
		}
	}
	__antithesis_instrumentation__.Notify(126097)
	if b != nil {
		__antithesis_instrumentation__.Notify(126119)
		if err := kvSS.sendBatch(ctx, stream, b); err != nil {
			__antithesis_instrumentation__.Notify(126121)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(126122)
		}
		__antithesis_instrumentation__.Notify(126120)
		bytesSent += int64(b.Len())
		bytesSentMetric.Inc(int64(b.Len()))
	} else {
		__antithesis_instrumentation__.Notify(126123)
	}
	__antithesis_instrumentation__.Notify(126098)

	kvSS.status = redact.Sprintf("kv pairs: %d", kvs)
	return bytesSent, nil
}

func (kvSS *kvBatchSnapshotStrategy) sendBatch(
	ctx context.Context, stream outgoingSnapshotStream, batch storage.Batch,
) error {
	__antithesis_instrumentation__.Notify(126124)
	if err := kvSS.limiter.WaitN(ctx, 1); err != nil {
		__antithesis_instrumentation__.Notify(126126)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126127)
	}
	__antithesis_instrumentation__.Notify(126125)
	return stream.Send(&kvserverpb.SnapshotRequest{KVBatch: batch.Repr()})
}

func (kvSS *kvBatchSnapshotStrategy) Status() redact.RedactableString {
	__antithesis_instrumentation__.Notify(126128)
	return kvSS.status
}

func (kvSS *kvBatchSnapshotStrategy) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(126129)
	if kvSS.scratch != nil {
		__antithesis_instrumentation__.Notify(126130)

		if err := kvSS.scratch.Clear(); err != nil {
			__antithesis_instrumentation__.Notify(126131)
			log.Warningf(ctx, "error closing kvBatchSnapshotStrategy: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(126132)
		}
	} else {
		__antithesis_instrumentation__.Notify(126133)
	}
}

func (s *Store) reserveSnapshot(
	ctx context.Context, header *kvserverpb.SnapshotRequest_Header,
) (_cleanup func(), _err error) {
	__antithesis_instrumentation__.Notify(126134)
	tBegin := timeutil.Now()

	if header.RangeSize != 0 {
		__antithesis_instrumentation__.Notify(126137)
		queueCtx := ctx
		if deadline, ok := queueCtx.Deadline(); ok {
			__antithesis_instrumentation__.Notify(126139)

			timeoutFrac := snapshotReservationQueueTimeoutFraction.Get(&s.ClusterSettings().SV)
			timeout := time.Duration(timeoutFrac * float64(timeutil.Until(deadline)))
			var cancel func()
			queueCtx, cancel = context.WithTimeout(queueCtx, timeout)
			defer cancel()
		} else {
			__antithesis_instrumentation__.Notify(126140)
		}
		__antithesis_instrumentation__.Notify(126138)
		select {
		case s.snapshotApplySem <- struct{}{}:
			__antithesis_instrumentation__.Notify(126141)
		case <-queueCtx.Done():
			__antithesis_instrumentation__.Notify(126142)
			if err := ctx.Err(); err != nil {
				__antithesis_instrumentation__.Notify(126145)
				return nil, errors.Wrap(err, "acquiring snapshot reservation")
			} else {
				__antithesis_instrumentation__.Notify(126146)
			}
			__antithesis_instrumentation__.Notify(126143)
			return nil, errors.Wrapf(queueCtx.Err(),
				"giving up during snapshot reservation due to %q",
				snapshotReservationQueueTimeoutFraction.Key())
		case <-s.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(126144)
			return nil, errors.Errorf("stopped")
		}
	} else {
		__antithesis_instrumentation__.Notify(126147)
	}
	__antithesis_instrumentation__.Notify(126135)

	const snapshotReservationWaitWarnThreshold = 32 * time.Second
	if elapsed := timeutil.Since(tBegin); elapsed > snapshotReservationWaitWarnThreshold {
		__antithesis_instrumentation__.Notify(126148)
		replDesc, _ := header.State.Desc.GetReplicaDescriptor(s.StoreID())
		log.Infof(
			ctx,
			"waited for %.1fs to acquire snapshot reservation to r%d/%d",
			elapsed.Seconds(),
			header.State.Desc.RangeID,
			replDesc.ReplicaID,
		)
	} else {
		__antithesis_instrumentation__.Notify(126149)
	}
	__antithesis_instrumentation__.Notify(126136)

	s.metrics.ReservedReplicaCount.Inc(1)
	s.metrics.Reserved.Inc(header.RangeSize)
	return func() {
		__antithesis_instrumentation__.Notify(126150)
		s.metrics.ReservedReplicaCount.Dec(1)
		s.metrics.Reserved.Dec(header.RangeSize)
		if header.RangeSize != 0 {
			__antithesis_instrumentation__.Notify(126151)
			<-s.snapshotApplySem
		} else {
			__antithesis_instrumentation__.Notify(126152)
		}
	}, nil
}

func (s *Store) canAcceptSnapshotLocked(
	ctx context.Context, snapHeader *kvserverpb.SnapshotRequest_Header,
) (*ReplicaPlaceholder, error) {
	__antithesis_instrumentation__.Notify(126153)

	desc := *snapHeader.State.Desc

	existingRepl, ok := s.mu.replicasByRangeID.Load(desc.RangeID)
	if !ok {
		__antithesis_instrumentation__.Notify(126158)
		return nil, errors.Errorf("canAcceptSnapshotLocked requires a replica present")
	} else {
		__antithesis_instrumentation__.Notify(126159)
	}
	__antithesis_instrumentation__.Notify(126154)

	existingRepl.raftMu.AssertHeld()

	existingRepl.mu.RLock()
	existingDesc := existingRepl.mu.state.Desc
	existingIsInitialized := existingDesc.IsInitialized()
	existingDestroyStatus := existingRepl.mu.destroyStatus
	existingRepl.mu.RUnlock()

	if existingIsInitialized {
		__antithesis_instrumentation__.Notify(126160)

		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(126161)
	}
	__antithesis_instrumentation__.Notify(126155)

	if existingDestroyStatus.Removed() {
		__antithesis_instrumentation__.Notify(126162)
		return nil, existingDestroyStatus.err
	} else {
		__antithesis_instrumentation__.Notify(126163)
	}
	__antithesis_instrumentation__.Notify(126156)

	if err := s.checkSnapshotOverlapLocked(ctx, snapHeader); err != nil {
		__antithesis_instrumentation__.Notify(126164)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(126165)
	}
	__antithesis_instrumentation__.Notify(126157)

	placeholder := &ReplicaPlaceholder{
		rangeDesc: desc,
	}
	return placeholder, nil
}

func (s *Store) checkSnapshotOverlapLocked(
	ctx context.Context, snapHeader *kvserverpb.SnapshotRequest_Header,
) error {
	__antithesis_instrumentation__.Notify(126166)
	desc := *snapHeader.State.Desc

	if exRng, ok := s.mu.replicaPlaceholders[desc.RangeID]; ok {
		__antithesis_instrumentation__.Notify(126169)
		return errors.Errorf("%s: canAcceptSnapshotLocked: cannot add placeholder, have an existing placeholder %s %v", s, exRng, snapHeader.RaftMessageRequest.FromReplica)
	} else {
		__antithesis_instrumentation__.Notify(126170)
	}
	__antithesis_instrumentation__.Notify(126167)

	if it := s.getOverlappingKeyRangeLocked(&desc); it.item != nil {
		__antithesis_instrumentation__.Notify(126171)

		exReplica, err := s.GetReplica(it.Desc().RangeID)
		msg := IntersectingSnapshotMsg
		if err != nil {
			__antithesis_instrumentation__.Notify(126173)
			log.Warningf(ctx, "unable to look up overlapping replica on %s: %v", exReplica, err)
		} else {
			__antithesis_instrumentation__.Notify(126174)
			inactive := func(r *Replica) bool {
				__antithesis_instrumentation__.Notify(126177)
				if r.RaftStatus() == nil {
					__antithesis_instrumentation__.Notify(126179)
					return true
				} else {
					__antithesis_instrumentation__.Notify(126180)
				}
				__antithesis_instrumentation__.Notify(126178)

				return !r.CurrentLeaseStatus(ctx).IsValid()
			}
			__antithesis_instrumentation__.Notify(126175)

			gcPriority := replicaGCPriorityDefault
			if inactive(exReplica) {
				__antithesis_instrumentation__.Notify(126181)
				gcPriority = replicaGCPrioritySuspect
			} else {
				__antithesis_instrumentation__.Notify(126182)
			}
			__antithesis_instrumentation__.Notify(126176)

			msg += "; initiated GC:"
			s.replicaGCQueue.AddAsync(ctx, exReplica, gcPriority)
		}
		__antithesis_instrumentation__.Notify(126172)
		return errors.Errorf("%s %v (incoming %v)", msg, exReplica, snapHeader.State.Desc.RSpan())
	} else {
		__antithesis_instrumentation__.Notify(126183)
	}
	__antithesis_instrumentation__.Notify(126168)
	return nil
}

func (s *Store) receiveSnapshot(
	ctx context.Context, header *kvserverpb.SnapshotRequest_Header, stream incomingSnapshotStream,
) error {
	__antithesis_instrumentation__.Notify(126184)

	if s.IsDraining() {
		__antithesis_instrumentation__.Notify(126196)
		switch t := header.Priority; t {
		case kvserverpb.SnapshotRequest_RECOVERY:
			__antithesis_instrumentation__.Notify(126197)

		case kvserverpb.SnapshotRequest_REBALANCE:
			__antithesis_instrumentation__.Notify(126198)
			return sendSnapshotError(stream, errors.New(storeDrainingMsg))
		default:
			__antithesis_instrumentation__.Notify(126199)

		}
	} else {
		__antithesis_instrumentation__.Notify(126200)
	}
	__antithesis_instrumentation__.Notify(126185)

	if fn := s.cfg.TestingKnobs.ReceiveSnapshot; fn != nil {
		__antithesis_instrumentation__.Notify(126201)
		if err := fn(header); err != nil {
			__antithesis_instrumentation__.Notify(126202)
			return sendSnapshotError(stream, err)
		} else {
			__antithesis_instrumentation__.Notify(126203)
		}
	} else {
		__antithesis_instrumentation__.Notify(126204)
	}
	__antithesis_instrumentation__.Notify(126186)

	storeID := s.StoreID()
	if _, ok := header.State.Desc.GetReplicaDescriptor(storeID); !ok {
		__antithesis_instrumentation__.Notify(126205)
		return errors.AssertionFailedf(
			`snapshot of type %s was sent to s%d which did not contain it as a replica: %s`,
			header.Type, storeID, header.State.Desc.Replicas())
	} else {
		__antithesis_instrumentation__.Notify(126206)
	}
	__antithesis_instrumentation__.Notify(126187)

	cleanup, err := s.reserveSnapshot(ctx, header)
	if err != nil {
		__antithesis_instrumentation__.Notify(126207)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126208)
	}
	__antithesis_instrumentation__.Notify(126188)
	defer cleanup()

	var placeholder *ReplicaPlaceholder
	if pErr := s.withReplicaForRequest(
		ctx, &header.RaftMessageRequest, func(ctx context.Context, r *Replica,
		) *roachpb.Error {
			__antithesis_instrumentation__.Notify(126209)
			var err error
			s.mu.Lock()
			defer s.mu.Unlock()
			placeholder, err = s.canAcceptSnapshotLocked(ctx, header)
			if err != nil {
				__antithesis_instrumentation__.Notify(126212)
				return roachpb.NewError(err)
			} else {
				__antithesis_instrumentation__.Notify(126213)
			}
			__antithesis_instrumentation__.Notify(126210)
			if placeholder != nil {
				__antithesis_instrumentation__.Notify(126214)
				if err := s.addPlaceholderLocked(placeholder); err != nil {
					__antithesis_instrumentation__.Notify(126215)
					return roachpb.NewError(err)
				} else {
					__antithesis_instrumentation__.Notify(126216)
				}
			} else {
				__antithesis_instrumentation__.Notify(126217)
			}
			__antithesis_instrumentation__.Notify(126211)
			return nil
		}); pErr != nil {
		__antithesis_instrumentation__.Notify(126218)
		log.Infof(ctx, "cannot accept snapshot: %s", pErr)
		return pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(126219)
	}
	__antithesis_instrumentation__.Notify(126189)

	defer func() {
		__antithesis_instrumentation__.Notify(126220)
		if placeholder != nil {
			__antithesis_instrumentation__.Notify(126221)

			if _, err := s.removePlaceholder(ctx, placeholder, removePlaceholderFailed); err != nil {
				__antithesis_instrumentation__.Notify(126222)
				log.Fatalf(ctx, "unable to remove placeholder: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(126223)
			}
		} else {
			__antithesis_instrumentation__.Notify(126224)
		}
	}()
	__antithesis_instrumentation__.Notify(126190)

	var ss snapshotStrategy
	switch header.Strategy {
	case kvserverpb.SnapshotRequest_KV_BATCH:
		__antithesis_instrumentation__.Notify(126225)
		snapUUID, err := uuid.FromBytes(header.RaftMessageRequest.Message.Snapshot.Data)
		if err != nil {
			__antithesis_instrumentation__.Notify(126228)
			err = errors.Wrap(err, "invalid snapshot")
			return sendSnapshotError(stream, err)
		} else {
			__antithesis_instrumentation__.Notify(126229)
		}
		__antithesis_instrumentation__.Notify(126226)

		ss = &kvBatchSnapshotStrategy{
			scratch:      s.sstSnapshotStorage.NewScratchSpace(header.State.Desc.RangeID, snapUUID),
			sstChunkSize: snapshotSSTWriteSyncRate.Get(&s.cfg.Settings.SV),
			st:           s.ClusterSettings(),
		}
		defer ss.Close(ctx)
	default:
		__antithesis_instrumentation__.Notify(126227)
		return sendSnapshotError(stream,
			errors.Errorf("%s,r%d: unknown snapshot strategy: %s",
				s, header.State.Desc.RangeID, header.Strategy),
		)
	}
	__antithesis_instrumentation__.Notify(126191)

	if err := stream.Send(&kvserverpb.SnapshotResponse{Status: kvserverpb.SnapshotResponse_ACCEPTED}); err != nil {
		__antithesis_instrumentation__.Notify(126230)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126231)
	}
	__antithesis_instrumentation__.Notify(126192)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(126232)
		log.Infof(ctx, "accepted snapshot reservation for r%d", header.State.Desc.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(126233)
	}
	__antithesis_instrumentation__.Notify(126193)

	inSnap, err := ss.Receive(ctx, stream, *header, s.metrics.RangeSnapshotRcvdBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(126234)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126235)
	}
	__antithesis_instrumentation__.Notify(126194)
	inSnap.placeholder = placeholder

	applyCtx := s.AnnotateCtx(context.Background())
	if err := s.processRaftSnapshotRequest(applyCtx, header, inSnap); err != nil {
		__antithesis_instrumentation__.Notify(126236)
		return sendSnapshotError(stream, errors.Wrap(err.GoError(), "failed to apply snapshot"))
	} else {
		__antithesis_instrumentation__.Notify(126237)
	}
	__antithesis_instrumentation__.Notify(126195)
	return stream.Send(&kvserverpb.SnapshotResponse{Status: kvserverpb.SnapshotResponse_APPLIED})
}

func sendSnapshotError(stream incomingSnapshotStream, err error) error {
	__antithesis_instrumentation__.Notify(126238)
	return stream.Send(&kvserverpb.SnapshotResponse{
		Status:  kvserverpb.SnapshotResponse_ERROR,
		Message: err.Error(),
	})
}

type SnapshotStorePool interface {
	throttle(reason throttleReason, why string, toStoreID roachpb.StoreID)
}

const minSnapshotRate = 1 << 20

var rebalanceSnapshotRate = settings.RegisterByteSizeSetting(
	settings.SystemOnly,
	"kv.snapshot_rebalance.max_rate",
	"the rate limit (bytes/sec) to use for rebalance and upreplication snapshots",
	32<<20,
	func(v int64) error {
		__antithesis_instrumentation__.Notify(126239)
		if v < minSnapshotRate {
			__antithesis_instrumentation__.Notify(126241)
			return errors.Errorf("snapshot rate cannot be set to a value below %s: %s",
				humanizeutil.IBytes(minSnapshotRate), humanizeutil.IBytes(v))
		} else {
			__antithesis_instrumentation__.Notify(126242)
		}
		__antithesis_instrumentation__.Notify(126240)
		return nil
	},
).WithPublic()

var recoverySnapshotRate = settings.RegisterByteSizeSetting(
	settings.SystemOnly,
	"kv.snapshot_recovery.max_rate",
	"the rate limit (bytes/sec) to use for recovery snapshots",
	32<<20,
	func(v int64) error {
		__antithesis_instrumentation__.Notify(126243)
		if v < minSnapshotRate {
			__antithesis_instrumentation__.Notify(126245)
			return errors.Errorf("snapshot rate cannot be set to a value below %s: %s",
				humanizeutil.IBytes(minSnapshotRate), humanizeutil.IBytes(v))
		} else {
			__antithesis_instrumentation__.Notify(126246)
		}
		__antithesis_instrumentation__.Notify(126244)
		return nil
	},
).WithPublic()

var snapshotSenderBatchSize = settings.RegisterByteSizeSetting(
	settings.SystemOnly,
	"kv.snapshot_sender.batch_size",
	"size of key-value batches sent over the network during snapshots",
	256<<10,
	settings.PositiveInt,
)

var snapshotReservationQueueTimeoutFraction = settings.RegisterFloatSetting(
	settings.SystemOnly,
	"kv.snapshot_receiver.reservation_queue_timeout_fraction",
	"the fraction of a snapshot's total timeout that it is allowed to spend "+
		"queued on the receiver waiting for a reservation",
	0.4,
	func(v float64) error {
		__antithesis_instrumentation__.Notify(126247)
		const min, max = 0.25, 1.0
		if v < min {
			__antithesis_instrumentation__.Notify(126249)
			return errors.Errorf("cannot set to a value less than %f: %f", min, v)
		} else {
			__antithesis_instrumentation__.Notify(126250)
			if v > max {
				__antithesis_instrumentation__.Notify(126251)
				return errors.Errorf("cannot set to a value greater than %f: %f", max, v)
			} else {
				__antithesis_instrumentation__.Notify(126252)
			}
		}
		__antithesis_instrumentation__.Notify(126248)
		return nil
	},
)

var snapshotSSTWriteSyncRate = settings.RegisterByteSizeSetting(
	settings.SystemOnly,
	"kv.snapshot_sst.sync_size",
	"threshold after which snapshot SST writes must fsync",
	bulkIOWriteBurst,
	settings.PositiveInt,
)

func snapshotRateLimit(
	st *cluster.Settings, priority kvserverpb.SnapshotRequest_Priority,
) (rate.Limit, error) {
	__antithesis_instrumentation__.Notify(126253)
	switch priority {
	case kvserverpb.SnapshotRequest_RECOVERY:
		__antithesis_instrumentation__.Notify(126254)
		return rate.Limit(recoverySnapshotRate.Get(&st.SV)), nil
	case kvserverpb.SnapshotRequest_REBALANCE:
		__antithesis_instrumentation__.Notify(126255)
		return rate.Limit(rebalanceSnapshotRate.Get(&st.SV)), nil
	default:
		__antithesis_instrumentation__.Notify(126256)
		return 0, errors.Errorf("unknown snapshot priority: %s", priority)
	}
}

func SendEmptySnapshot(
	ctx context.Context,
	st *cluster.Settings,
	cc *grpc.ClientConn,
	now hlc.Timestamp,
	desc roachpb.RangeDescriptor,
	to roachpb.ReplicaDescriptor,
) error {
	__antithesis_instrumentation__.Notify(126257)

	eng, err := storage.Open(
		context.Background(),
		storage.InMemory(),
		storage.CacheSize(1<<20),
		storage.MaxSize(512<<20))
	if err != nil {
		__antithesis_instrumentation__.Notify(126269)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126270)
	}
	__antithesis_instrumentation__.Notify(126258)
	defer eng.Close()

	var ms enginepb.MVCCStats

	if err := storage.MVCCPutProto(
		ctx, eng, &ms, keys.RangeDescriptorKey(desc.StartKey), now, nil, &desc,
	); err != nil {
		__antithesis_instrumentation__.Notify(126271)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126272)
	}
	__antithesis_instrumentation__.Notify(126259)

	writeAppliedIndexTerm := st.Version.IsActive(ctx, clusterversion.AddRaftAppliedIndexTermMigration)
	if !writeAppliedIndexTerm {
		__antithesis_instrumentation__.Notify(126273)
		return errors.Errorf("cluster version is too old %s",
			st.Version.ActiveVersionOrEmpty(ctx))
	} else {
		__antithesis_instrumentation__.Notify(126274)
	}
	__antithesis_instrumentation__.Notify(126260)
	ms, err = stateloader.WriteInitialReplicaState(
		ctx,
		eng,
		ms,
		desc,
		roachpb.Lease{},
		hlc.Timestamp{},
		st.Version.ActiveVersionOrEmpty(ctx).Version,
		writeAppliedIndexTerm,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(126275)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126276)
	}
	__antithesis_instrumentation__.Notify(126261)

	sl := stateloader.Make(desc.RangeID)
	state, err := sl.Load(ctx, eng, &desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(126277)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126278)
	}
	__antithesis_instrumentation__.Notify(126262)

	state.DeprecatedUsingAppliedStateKey = true

	hs, err := sl.LoadHardState(ctx, eng)
	if err != nil {
		__antithesis_instrumentation__.Notify(126279)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126280)
	}
	__antithesis_instrumentation__.Notify(126263)

	snapUUID, err := uuid.NewV4()
	if err != nil {
		__antithesis_instrumentation__.Notify(126281)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126282)
	}
	__antithesis_instrumentation__.Notify(126264)

	outgoingSnap, err := snapshot(
		ctx,
		snapUUID,
		sl,

		kvserverpb.SnapshotRequest_VIA_SNAPSHOT_QUEUE,
		eng,
		desc.RangeID,
		raftentry.NewCache(1),
		func(func(SideloadStorage) error) error { __antithesis_instrumentation__.Notify(126283); return nil },
		desc.StartKey,
	)
	__antithesis_instrumentation__.Notify(126265)
	if err != nil {
		__antithesis_instrumentation__.Notify(126284)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126285)
	}
	__antithesis_instrumentation__.Notify(126266)
	defer outgoingSnap.Close()

	from := to
	req := kvserverpb.RaftMessageRequest{
		RangeID:     desc.RangeID,
		FromReplica: from,
		ToReplica:   to,
		Message: raftpb.Message{
			Type:     raftpb.MsgSnap,
			To:       uint64(to.ReplicaID),
			From:     uint64(from.ReplicaID),
			Term:     hs.Term,
			Snapshot: outgoingSnap.RaftSnap,
		},
	}

	header := kvserverpb.SnapshotRequest_Header{
		State:                                state,
		RaftMessageRequest:                   req,
		RangeSize:                            ms.Total(),
		Priority:                             kvserverpb.SnapshotRequest_RECOVERY,
		Strategy:                             kvserverpb.SnapshotRequest_KV_BATCH,
		Type:                                 kvserverpb.SnapshotRequest_VIA_SNAPSHOT_QUEUE,
		DeprecatedUnreplicatedTruncatedState: true,
	}

	stream, err := NewMultiRaftClient(cc).RaftSnapshot(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(126286)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126287)
	}
	__antithesis_instrumentation__.Notify(126267)

	defer func() {
		__antithesis_instrumentation__.Notify(126288)
		if err := stream.CloseSend(); err != nil {
			__antithesis_instrumentation__.Notify(126289)
			log.Warningf(ctx, "failed to close snapshot stream: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(126290)
		}
	}()
	__antithesis_instrumentation__.Notify(126268)

	return sendSnapshot(
		ctx,
		st,
		stream,
		noopStorePool{},
		header,
		&outgoingSnap,
		eng.NewBatch,
		func() { __antithesis_instrumentation__.Notify(126291) },
		nil,
	)
}

type noopStorePool struct{}

func (n noopStorePool) throttle(throttleReason, string, roachpb.StoreID) {
	__antithesis_instrumentation__.Notify(126292)
}

func sendSnapshot(
	ctx context.Context,
	st *cluster.Settings,
	stream outgoingSnapshotStream,
	storePool SnapshotStorePool,
	header kvserverpb.SnapshotRequest_Header,
	snap *OutgoingSnapshot,
	newBatch func() storage.Batch,
	sent func(),
	bytesSentCounter *metric.Counter,
) error {
	__antithesis_instrumentation__.Notify(126293)
	if bytesSentCounter == nil {
		__antithesis_instrumentation__.Notify(126304)

		bytesSentCounter = metric.NewCounter(metric.Metadata{Name: "range.snapshots.sent-bytes"})
	} else {
		__antithesis_instrumentation__.Notify(126305)
	}
	__antithesis_instrumentation__.Notify(126294)
	start := timeutil.Now()
	to := header.RaftMessageRequest.ToReplica
	if err := stream.Send(&kvserverpb.SnapshotRequest{Header: &header}); err != nil {
		__antithesis_instrumentation__.Notify(126306)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126307)
	}
	__antithesis_instrumentation__.Notify(126295)

	resp, err := stream.Recv()
	if err != nil {
		__antithesis_instrumentation__.Notify(126308)
		storePool.throttle(throttleFailed, err.Error(), to.StoreID)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126309)
	}
	__antithesis_instrumentation__.Notify(126296)
	switch resp.Status {
	case kvserverpb.SnapshotResponse_ERROR:
		__antithesis_instrumentation__.Notify(126310)
		storePool.throttle(throttleFailed, resp.Message, to.StoreID)
		return errors.Errorf("%s: remote couldn't accept %s with error: %s",
			to, snap, resp.Message)
	case kvserverpb.SnapshotResponse_ACCEPTED:
		__antithesis_instrumentation__.Notify(126311)

	default:
		__antithesis_instrumentation__.Notify(126312)
		err := errors.Errorf("%s: server sent an invalid status while negotiating %s: %s",
			to, snap, resp.Status)
		storePool.throttle(throttleFailed, err.Error(), to.StoreID)
		return err
	}
	__antithesis_instrumentation__.Notify(126297)

	durQueued := timeutil.Since(start)
	start = timeutil.Now()

	targetRate, err := snapshotRateLimit(st, header.Priority)
	if err != nil {
		__antithesis_instrumentation__.Notify(126313)
		return errors.Wrapf(err, "%s", to)
	} else {
		__antithesis_instrumentation__.Notify(126314)
	}
	__antithesis_instrumentation__.Notify(126298)
	batchSize := snapshotSenderBatchSize.Get(&st.SV)

	limiter := rate.NewLimiter(targetRate/rate.Limit(batchSize), 1)

	var ss snapshotStrategy
	switch header.Strategy {
	case kvserverpb.SnapshotRequest_KV_BATCH:
		__antithesis_instrumentation__.Notify(126315)
		ss = &kvBatchSnapshotStrategy{
			batchSize: batchSize,
			limiter:   limiter,
			newBatch:  newBatch,
			st:        st,
		}
	default:
		__antithesis_instrumentation__.Notify(126316)
		log.Fatalf(ctx, "unknown snapshot strategy: %s", header.Strategy)
	}
	__antithesis_instrumentation__.Notify(126299)

	numBytesSent, err := ss.Send(ctx, stream, header, snap, bytesSentCounter)
	if err != nil {
		__antithesis_instrumentation__.Notify(126317)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126318)
	}
	__antithesis_instrumentation__.Notify(126300)
	durSent := timeutil.Since(start)

	sent()
	if err := stream.Send(&kvserverpb.SnapshotRequest{Final: true}); err != nil {
		__antithesis_instrumentation__.Notify(126319)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126320)
	}
	__antithesis_instrumentation__.Notify(126301)
	log.Infof(
		ctx,
		"streamed %s to %s with %s in %.2fs @ %s/s: %s, rate-limit: %s/s, queued: %.2fs",
		snap,
		to,
		humanizeutil.IBytes(numBytesSent),
		durSent.Seconds(),
		humanizeutil.IBytes(int64(float64(numBytesSent)/durSent.Seconds())),
		ss.Status(),
		humanizeutil.IBytes(int64(targetRate)),
		durQueued.Seconds(),
	)

	resp, err = stream.Recv()
	if err != nil {
		__antithesis_instrumentation__.Notify(126321)
		return errors.Wrapf(err, "%s: remote failed to apply snapshot", to)
	} else {
		__antithesis_instrumentation__.Notify(126322)
	}
	__antithesis_instrumentation__.Notify(126302)

	if unexpectedResp, err := stream.Recv(); err != io.EOF {
		__antithesis_instrumentation__.Notify(126323)
		if err != nil {
			__antithesis_instrumentation__.Notify(126325)
			return errors.Wrapf(err, "%s: expected EOF, got resp=%v with error", to, unexpectedResp)
		} else {
			__antithesis_instrumentation__.Notify(126326)
		}
		__antithesis_instrumentation__.Notify(126324)
		return errors.Newf("%s: expected EOF, got resp=%v", to, unexpectedResp)
	} else {
		__antithesis_instrumentation__.Notify(126327)
	}
	__antithesis_instrumentation__.Notify(126303)
	switch resp.Status {
	case kvserverpb.SnapshotResponse_ERROR:
		__antithesis_instrumentation__.Notify(126328)
		return errors.Errorf("%s: remote failed to apply snapshot for reason %s", to, resp.Message)
	case kvserverpb.SnapshotResponse_APPLIED:
		__antithesis_instrumentation__.Notify(126329)
		return nil
	default:
		__antithesis_instrumentation__.Notify(126330)
		return errors.Errorf("%s: server sent an invalid status during finalization: %s",
			to, resp.Status)
	}
}
