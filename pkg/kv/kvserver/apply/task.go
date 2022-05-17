package apply

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type StateMachine interface {
	NewBatch(ephemeral bool) Batch

	ApplySideEffects(context.Context, CheckedCommand) (AppliedCommand, error)
}

var ErrRemoved = errors.New("replica removed")

type Batch interface {
	Stage(context.Context, Command) (CheckedCommand, error)

	ApplyToStateMachine(context.Context) error

	Close()
}

type Decoder interface {
	DecodeAndBind(context.Context, []raftpb.Entry) (anyLocal bool, _ error)

	NewCommandIter() CommandIterator

	Reset()
}

type Task struct {
	sm  StateMachine
	dec Decoder

	decoded bool

	anyLocal bool

	batchSize int32
}

func MakeTask(sm StateMachine, dec Decoder) Task {
	__antithesis_instrumentation__.Notify(96251)
	return Task{sm: sm, dec: dec}
}

func (t *Task) Decode(ctx context.Context, committedEntries []raftpb.Entry) error {
	__antithesis_instrumentation__.Notify(96252)
	var err error
	t.anyLocal, err = t.dec.DecodeAndBind(ctx, committedEntries)
	t.decoded = true
	return err
}

func (t *Task) assertDecoded() {
	__antithesis_instrumentation__.Notify(96253)
	if !t.decoded {
		__antithesis_instrumentation__.Notify(96254)
		panic("Task.Decode not called yet")
	} else {
		__antithesis_instrumentation__.Notify(96255)
	}
}

func (t *Task) AckCommittedEntriesBeforeApplication(ctx context.Context, maxIndex uint64) error {
	__antithesis_instrumentation__.Notify(96256)
	t.assertDecoded()
	if !t.anyLocal {
		__antithesis_instrumentation__.Notify(96260)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(96261)
	}
	__antithesis_instrumentation__.Notify(96257)

	batch := t.sm.NewBatch(true)
	defer batch.Close()

	iter := t.dec.NewCommandIter()
	defer iter.Close()

	batchIter := takeWhileCmdIter(iter, func(cmd Command) bool {
		__antithesis_instrumentation__.Notify(96262)
		if cmd.Index() > maxIndex {
			__antithesis_instrumentation__.Notify(96264)
			return false
		} else {
			__antithesis_instrumentation__.Notify(96265)
		}
		__antithesis_instrumentation__.Notify(96263)
		return cmd.IsTrivial()
	})
	__antithesis_instrumentation__.Notify(96258)

	stagedIter, err := mapCmdIter(batchIter, batch.Stage)
	if err != nil {
		__antithesis_instrumentation__.Notify(96266)
		return err
	} else {
		__antithesis_instrumentation__.Notify(96267)
	}
	__antithesis_instrumentation__.Notify(96259)

	return forEachCheckedCmdIter(ctx, stagedIter, func(cmd CheckedCommand, ctx context.Context) error {
		__antithesis_instrumentation__.Notify(96268)
		if !cmd.Rejected() && func() bool {
			__antithesis_instrumentation__.Notify(96270)
			return cmd.IsLocal() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(96271)
			return cmd.CanAckBeforeApplication() == true
		}() == true {
			__antithesis_instrumentation__.Notify(96272)
			return cmd.AckSuccess(cmd.Ctx())
		} else {
			__antithesis_instrumentation__.Notify(96273)
		}
		__antithesis_instrumentation__.Notify(96269)
		return nil
	})
}

func (t *Task) SetMaxBatchSize(size int) {
	__antithesis_instrumentation__.Notify(96274)
	t.batchSize = int32(size)
}

func (t *Task) ApplyCommittedEntries(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(96275)
	t.assertDecoded()

	iter := t.dec.NewCommandIter()
	for iter.Valid() {
		__antithesis_instrumentation__.Notify(96277)
		if err := t.applyOneBatch(ctx, iter); err != nil {
			__antithesis_instrumentation__.Notify(96278)

			if rejectErr := forEachCmdIter(ctx, iter, func(cmd Command, ctx context.Context) error {
				__antithesis_instrumentation__.Notify(96280)
				return cmd.AckErrAndFinish(ctx, err)
			}); rejectErr != nil {
				__antithesis_instrumentation__.Notify(96281)
				return rejectErr
			} else {
				__antithesis_instrumentation__.Notify(96282)
			}
			__antithesis_instrumentation__.Notify(96279)
			return err
		} else {
			__antithesis_instrumentation__.Notify(96283)
		}
	}
	__antithesis_instrumentation__.Notify(96276)
	iter.Close()
	return nil
}

func (t *Task) applyOneBatch(ctx context.Context, iter CommandIterator) error {
	__antithesis_instrumentation__.Notify(96284)

	batch := t.sm.NewBatch(false)
	defer batch.Close()

	pol := trivialPolicy{maxCount: t.batchSize}
	batchIter := takeWhileCmdIter(iter, func(cmd Command) bool {
		__antithesis_instrumentation__.Notify(96289)
		return pol.maybeAdd(cmd.IsTrivial())
	})
	__antithesis_instrumentation__.Notify(96285)

	stagedIter, err := mapCmdIter(batchIter, batch.Stage)
	if err != nil {
		__antithesis_instrumentation__.Notify(96290)
		return err
	} else {
		__antithesis_instrumentation__.Notify(96291)
	}
	__antithesis_instrumentation__.Notify(96286)

	if err := batch.ApplyToStateMachine(ctx); err != nil {
		__antithesis_instrumentation__.Notify(96292)
		return err
	} else {
		__antithesis_instrumentation__.Notify(96293)
	}
	__antithesis_instrumentation__.Notify(96287)

	appliedIter, err := mapCheckedCmdIter(stagedIter, t.sm.ApplySideEffects)
	if err != nil {
		__antithesis_instrumentation__.Notify(96294)
		return err
	} else {
		__antithesis_instrumentation__.Notify(96295)
	}
	__antithesis_instrumentation__.Notify(96288)

	return forEachAppliedCmdIter(ctx, appliedIter, AppliedCommand.AckOutcomeAndFinish)
}

type trivialPolicy struct {
	maxCount int32

	trivialCount    int32
	nonTrivialCount int32
}

func (p *trivialPolicy) maybeAdd(trivial bool) bool {
	__antithesis_instrumentation__.Notify(96296)
	if !trivial {
		__antithesis_instrumentation__.Notify(96300)
		if p.trivialCount+p.nonTrivialCount > 0 {
			__antithesis_instrumentation__.Notify(96302)
			return false
		} else {
			__antithesis_instrumentation__.Notify(96303)
		}
		__antithesis_instrumentation__.Notify(96301)
		p.nonTrivialCount++
		return true
	} else {
		__antithesis_instrumentation__.Notify(96304)
	}
	__antithesis_instrumentation__.Notify(96297)
	if p.nonTrivialCount > 0 {
		__antithesis_instrumentation__.Notify(96305)
		return false
	} else {
		__antithesis_instrumentation__.Notify(96306)
	}
	__antithesis_instrumentation__.Notify(96298)
	if p.maxCount > 0 && func() bool {
		__antithesis_instrumentation__.Notify(96307)
		return p.maxCount == p.trivialCount == true
	}() == true {
		__antithesis_instrumentation__.Notify(96308)
		return false
	} else {
		__antithesis_instrumentation__.Notify(96309)
	}
	__antithesis_instrumentation__.Notify(96299)
	p.trivialCount++
	return true
}

func (t *Task) Close() {
	__antithesis_instrumentation__.Notify(96310)
	t.dec.Reset()
	*t = Task{}
}
