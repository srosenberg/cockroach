package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func (bq *baseQueue) testingAdd(
	ctx context.Context, repl replicaInQueue, priority float64,
) (bool, error) {
	__antithesis_instrumentation__.Notify(112578)
	return bq.addInternal(ctx, repl.Desc(), repl.ReplicaID(), priority)
}

func forceScanAndProcess(ctx context.Context, s *Store, q *baseQueue) error {
	__antithesis_instrumentation__.Notify(112579)

	if _, err := s.GetConfReader(ctx); err != nil {
		__antithesis_instrumentation__.Notify(112582)
		return errors.Wrap(err, "unable to retrieve conf reader")
	} else {
		__antithesis_instrumentation__.Notify(112583)
	}
	__antithesis_instrumentation__.Notify(112580)

	newStoreReplicaVisitor(s).Visit(func(repl *Replica) bool {
		__antithesis_instrumentation__.Notify(112584)
		q.maybeAdd(context.Background(), repl, s.cfg.Clock.NowAsClockTimestamp())
		return true
	})
	__antithesis_instrumentation__.Notify(112581)

	q.DrainQueue(s.stopper)
	return nil
}

func mustForceScanAndProcess(ctx context.Context, s *Store, q *baseQueue) {
	__antithesis_instrumentation__.Notify(112585)
	if err := forceScanAndProcess(ctx, s, q); err != nil {
		__antithesis_instrumentation__.Notify(112586)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(112587)
	}
}

func (s *Store) ForceReplicationScanAndProcess() error {
	__antithesis_instrumentation__.Notify(112588)
	return forceScanAndProcess(context.TODO(), s, s.replicateQueue.baseQueue)
}

func (s *Store) MustForceReplicaGCScanAndProcess() {
	__antithesis_instrumentation__.Notify(112589)
	mustForceScanAndProcess(context.TODO(), s, s.replicaGCQueue.baseQueue)
}

func (s *Store) MustForceMergeScanAndProcess() {
	__antithesis_instrumentation__.Notify(112590)
	mustForceScanAndProcess(context.TODO(), s, s.mergeQueue.baseQueue)
}

func (s *Store) ForceSplitScanAndProcess() error {
	__antithesis_instrumentation__.Notify(112591)
	return forceScanAndProcess(context.TODO(), s, s.splitQueue.baseQueue)
}

func (s *Store) MustForceRaftLogScanAndProcess() {
	__antithesis_instrumentation__.Notify(112592)
	mustForceScanAndProcess(context.TODO(), s, s.raftLogQueue.baseQueue)
}

func (s *Store) ForceTimeSeriesMaintenanceQueueProcess() error {
	__antithesis_instrumentation__.Notify(112593)
	return forceScanAndProcess(context.TODO(), s, s.tsMaintenanceQueue.baseQueue)
}

func (s *Store) ForceRaftSnapshotQueueProcess() error {
	__antithesis_instrumentation__.Notify(112594)
	return forceScanAndProcess(context.TODO(), s, s.raftSnapshotQueue.baseQueue)
}

func (s *Store) ForceConsistencyQueueProcess() error {
	__antithesis_instrumentation__.Notify(112595)
	return forceScanAndProcess(context.TODO(), s, s.consistencyQueue.baseQueue)
}

func (s *Store) setGCQueueActive(active bool) {
	__antithesis_instrumentation__.Notify(112596)
	s.mvccGCQueue.SetDisabled(!active)
}
func (s *Store) setMergeQueueActive(active bool) {
	__antithesis_instrumentation__.Notify(112597)
	s.mergeQueue.SetDisabled(!active)
}
func (s *Store) setRaftLogQueueActive(active bool) {
	__antithesis_instrumentation__.Notify(112598)
	s.raftLogQueue.SetDisabled(!active)
}
func (s *Store) setReplicaGCQueueActive(active bool) {
	__antithesis_instrumentation__.Notify(112599)
	s.replicaGCQueue.SetDisabled(!active)
}

func (s *Store) SetReplicateQueueActive(active bool) {
	__antithesis_instrumentation__.Notify(112600)
	s.replicateQueue.SetDisabled(!active)
}
func (s *Store) setSplitQueueActive(active bool) {
	__antithesis_instrumentation__.Notify(112601)
	s.splitQueue.SetDisabled(!active)
}
func (s *Store) setTimeSeriesMaintenanceQueueActive(active bool) {
	__antithesis_instrumentation__.Notify(112602)
	s.tsMaintenanceQueue.SetDisabled(!active)
}
func (s *Store) setRaftSnapshotQueueActive(active bool) {
	__antithesis_instrumentation__.Notify(112603)
	s.raftSnapshotQueue.SetDisabled(!active)
}
func (s *Store) setConsistencyQueueActive(active bool) {
	__antithesis_instrumentation__.Notify(112604)
	s.consistencyQueue.SetDisabled(!active)
}
func (s *Store) setScannerActive(active bool) {
	__antithesis_instrumentation__.Notify(112605)
	s.scanner.SetDisabled(!active)
}
