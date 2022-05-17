package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func (r *Replica) setCorruptRaftMuLocked(
	ctx context.Context, cErr *roachpb.ReplicaCorruptionError,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(117131)
	r.readOnlyCmdMu.Lock()
	defer r.readOnlyCmdMu.Unlock()
	r.mu.Lock()
	defer r.mu.Unlock()

	log.ErrorfDepth(ctx, 1, "stalling replica due to: %s", cErr.ErrorMsg)
	cErr.Processed = true
	r.mu.destroyStatus.Set(cErr, destroyReasonRemoved)

	auxDir := r.store.engine.GetAuxiliaryDir()
	_ = r.store.engine.MkdirAll(auxDir)
	path := base.PreventedStartupFile(auxDir)

	preventStartupMsg := fmt.Sprintf(`ATTENTION:

this node is terminating because replica %s detected an inconsistent state.
Please contact the CockroachDB support team. It is not necessarily safe
to replace this node; cluster data may still be at risk of corruption.

A file preventing this node from restarting was placed at:
%s
`, r, path)

	if err := fs.WriteFile(r.store.engine, path, []byte(preventStartupMsg)); err != nil {
		__antithesis_instrumentation__.Notify(117133)
		log.Warningf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(117134)
	}
	__antithesis_instrumentation__.Notify(117132)

	log.FatalfDepth(ctx, 1, "replica is corrupted: %s", cErr)
	return roachpb.NewError(cErr)
}
