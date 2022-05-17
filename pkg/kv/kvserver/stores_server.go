package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/redact"
)

type Server struct {
	stores *Stores
}

var _ PerReplicaServer = Server{}
var _ PerStoreServer = Server{}

func MakeServer(descriptor *roachpb.NodeDescriptor, stores *Stores) Server {
	__antithesis_instrumentation__.Notify(126559)
	return Server{stores}
}

func (is Server) execStoreCommand(
	ctx context.Context, h StoreRequestHeader, f func(context.Context, *Store) error,
) error {
	__antithesis_instrumentation__.Notify(126560)
	store, err := is.stores.GetStore(h.StoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(126562)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126563)
	}
	__antithesis_instrumentation__.Notify(126561)

	return store.stopper.RunTaskWithErr(ctx, "store command", func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(126564)
		return f(ctx, store)
	})
}

func (is Server) CollectChecksum(
	ctx context.Context, req *CollectChecksumRequest,
) (*CollectChecksumResponse, error) {
	__antithesis_instrumentation__.Notify(126565)
	resp := &CollectChecksumResponse{}
	err := is.execStoreCommand(ctx, req.StoreRequestHeader,
		func(ctx context.Context, s *Store) error {
			__antithesis_instrumentation__.Notify(126567)
			r, err := s.GetReplica(req.RangeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(126571)
				return err
			} else {
				__antithesis_instrumentation__.Notify(126572)
			}
			__antithesis_instrumentation__.Notify(126568)
			c, err := r.getChecksum(ctx, req.ChecksumID)
			if err != nil {
				__antithesis_instrumentation__.Notify(126573)
				return err
			} else {
				__antithesis_instrumentation__.Notify(126574)
			}
			__antithesis_instrumentation__.Notify(126569)
			ccr := c.CollectChecksumResponse
			if !bytes.Equal(req.Checksum, ccr.Checksum) {
				__antithesis_instrumentation__.Notify(126575)

				if len(req.Checksum) > 0 {
					__antithesis_instrumentation__.Notify(126576)
					log.Errorf(ctx, "consistency check failed on range r%d: expected checksum %x, got %x",
						req.RangeID, redact.Safe(req.Checksum), redact.Safe(ccr.Checksum))

				} else {
					__antithesis_instrumentation__.Notify(126577)
				}
			} else {
				__antithesis_instrumentation__.Notify(126578)
				ccr.Snapshot = nil
			}
			__antithesis_instrumentation__.Notify(126570)
			resp = &ccr
			return nil
		})
	__antithesis_instrumentation__.Notify(126566)
	return resp, err
}

func (is Server) WaitForApplication(
	ctx context.Context, req *WaitForApplicationRequest,
) (*WaitForApplicationResponse, error) {
	__antithesis_instrumentation__.Notify(126579)
	resp := &WaitForApplicationResponse{}
	err := is.execStoreCommand(ctx, req.StoreRequestHeader, func(ctx context.Context, s *Store) error {
		__antithesis_instrumentation__.Notify(126581)

		retryOpts := retry.Options{InitialBackoff: 10 * time.Millisecond}
		for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
			__antithesis_instrumentation__.Notify(126584)

			repl, err := s.GetReplica(req.RangeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(126586)
				return err
			} else {
				__antithesis_instrumentation__.Notify(126587)
			}
			__antithesis_instrumentation__.Notify(126585)
			repl.mu.RLock()
			leaseAppliedIndex := repl.mu.state.LeaseAppliedIndex
			repl.mu.RUnlock()
			if leaseAppliedIndex >= req.LeaseIndex {
				__antithesis_instrumentation__.Notify(126588)

				return storage.WriteSyncNoop(ctx, s.engine)
			} else {
				__antithesis_instrumentation__.Notify(126589)
			}
		}
		__antithesis_instrumentation__.Notify(126582)
		if ctx.Err() == nil {
			__antithesis_instrumentation__.Notify(126590)
			log.Fatal(ctx, "infinite retry loop exited but context has no error")
		} else {
			__antithesis_instrumentation__.Notify(126591)
		}
		__antithesis_instrumentation__.Notify(126583)
		return ctx.Err()
	})
	__antithesis_instrumentation__.Notify(126580)
	return resp, err
}

func (is Server) WaitForReplicaInit(
	ctx context.Context, req *WaitForReplicaInitRequest,
) (*WaitForReplicaInitResponse, error) {
	__antithesis_instrumentation__.Notify(126592)
	resp := &WaitForReplicaInitResponse{}
	err := is.execStoreCommand(ctx, req.StoreRequestHeader, func(ctx context.Context, s *Store) error {
		__antithesis_instrumentation__.Notify(126594)
		retryOpts := retry.Options{InitialBackoff: 10 * time.Millisecond}
		for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
			__antithesis_instrumentation__.Notify(126597)

			if repl, err := s.GetReplica(req.RangeID); err == nil && func() bool {
				__antithesis_instrumentation__.Notify(126598)
				return repl.IsInitialized() == true
			}() == true {
				__antithesis_instrumentation__.Notify(126599)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(126600)
			}
		}
		__antithesis_instrumentation__.Notify(126595)
		if ctx.Err() == nil {
			__antithesis_instrumentation__.Notify(126601)
			log.Fatal(ctx, "infinite retry loop exited but context has no error")
		} else {
			__antithesis_instrumentation__.Notify(126602)
		}
		__antithesis_instrumentation__.Notify(126596)
		return ctx.Err()
	})
	__antithesis_instrumentation__.Notify(126593)
	return resp, err
}

func (is Server) CompactEngineSpan(
	ctx context.Context, req *CompactEngineSpanRequest,
) (*CompactEngineSpanResponse, error) {
	__antithesis_instrumentation__.Notify(126603)
	resp := &CompactEngineSpanResponse{}
	err := is.execStoreCommand(ctx, req.StoreRequestHeader,
		func(ctx context.Context, s *Store) error {
			__antithesis_instrumentation__.Notify(126605)
			return s.Engine().CompactRange(req.Span.Key, req.Span.EndKey)
		})
	__antithesis_instrumentation__.Notify(126604)
	return resp, err
}
