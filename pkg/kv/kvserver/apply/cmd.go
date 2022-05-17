package apply

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "context"

type Command interface {
	Index() uint64

	IsTrivial() bool

	IsLocal() bool

	Ctx() context.Context

	AckErrAndFinish(context.Context, error) error
}

type CheckedCommand interface {
	Command

	Rejected() bool

	CanAckBeforeApplication() bool

	AckSuccess(context.Context) error
}

type AppliedCommand interface {
	CheckedCommand

	AckOutcomeAndFinish(context.Context) error
}

type CommandIteratorBase interface {
	Valid() bool

	Next()

	Close()
}

type CommandIterator interface {
	CommandIteratorBase

	Cur() Command

	NewList() CommandList

	NewCheckedList() CheckedCommandList
}

type CommandList interface {
	CommandIterator

	Append(Command)
}

type CheckedCommandIterator interface {
	CommandIteratorBase

	CurChecked() CheckedCommand

	NewAppliedList() AppliedCommandList
}

type CheckedCommandList interface {
	CheckedCommandIterator

	AppendChecked(CheckedCommand)
}

type AppliedCommandIterator interface {
	CommandIteratorBase

	CurApplied() AppliedCommand
}

type AppliedCommandList interface {
	AppliedCommandIterator

	AppendApplied(AppliedCommand)
}

func takeWhileCmdIter(iter CommandIterator, pred func(Command) bool) CommandIterator {
	__antithesis_instrumentation__.Notify(96215)
	ret := iter.NewList()
	for iter.Valid() {
		__antithesis_instrumentation__.Notify(96217)
		cmd := iter.Cur()
		if !pred(cmd) {
			__antithesis_instrumentation__.Notify(96219)
			break
		} else {
			__antithesis_instrumentation__.Notify(96220)
		}
		__antithesis_instrumentation__.Notify(96218)
		iter.Next()
		ret.Append(cmd)
	}
	__antithesis_instrumentation__.Notify(96216)
	return ret
}

func mapCmdIter(
	iter CommandIterator, fn func(context.Context, Command) (CheckedCommand, error),
) (CheckedCommandIterator, error) {
	__antithesis_instrumentation__.Notify(96221)
	defer iter.Close()
	ret := iter.NewCheckedList()
	for iter.Valid() {
		__antithesis_instrumentation__.Notify(96223)
		cur := iter.Cur()
		checked, err := fn(cur.Ctx(), cur)
		if err != nil {
			__antithesis_instrumentation__.Notify(96225)
			ret.Close()
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(96226)
		}
		__antithesis_instrumentation__.Notify(96224)
		iter.Next()
		ret.AppendChecked(checked)
	}
	__antithesis_instrumentation__.Notify(96222)
	return ret, nil
}

func mapCheckedCmdIter(
	iter CheckedCommandIterator, fn func(context.Context, CheckedCommand) (AppliedCommand, error),
) (AppliedCommandIterator, error) {
	__antithesis_instrumentation__.Notify(96227)
	defer iter.Close()
	ret := iter.NewAppliedList()
	for iter.Valid() {
		__antithesis_instrumentation__.Notify(96229)
		curChecked := iter.CurChecked()
		applied, err := fn(curChecked.Ctx(), curChecked)
		if err != nil {
			__antithesis_instrumentation__.Notify(96231)
			ret.Close()
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(96232)
		}
		__antithesis_instrumentation__.Notify(96230)
		iter.Next()
		ret.AppendApplied(applied)
	}
	__antithesis_instrumentation__.Notify(96228)
	return ret, nil
}

func forEachCmdIter(
	ctx context.Context, iter CommandIterator, fn func(Command, context.Context) error,
) error {
	__antithesis_instrumentation__.Notify(96233)
	defer iter.Close()
	for iter.Valid() {
		__antithesis_instrumentation__.Notify(96235)
		if err := fn(iter.Cur(), ctx); err != nil {
			__antithesis_instrumentation__.Notify(96237)
			return err
		} else {
			__antithesis_instrumentation__.Notify(96238)
		}
		__antithesis_instrumentation__.Notify(96236)
		iter.Next()
	}
	__antithesis_instrumentation__.Notify(96234)
	return nil
}

func forEachCheckedCmdIter(
	ctx context.Context, iter CheckedCommandIterator, fn func(CheckedCommand, context.Context) error,
) error {
	__antithesis_instrumentation__.Notify(96239)
	defer iter.Close()
	for iter.Valid() {
		__antithesis_instrumentation__.Notify(96241)
		if err := fn(iter.CurChecked(), ctx); err != nil {
			__antithesis_instrumentation__.Notify(96243)
			return err
		} else {
			__antithesis_instrumentation__.Notify(96244)
		}
		__antithesis_instrumentation__.Notify(96242)
		iter.Next()
	}
	__antithesis_instrumentation__.Notify(96240)
	return nil
}

func forEachAppliedCmdIter(
	ctx context.Context, iter AppliedCommandIterator, fn func(AppliedCommand, context.Context) error,
) error {
	__antithesis_instrumentation__.Notify(96245)
	defer iter.Close()
	for iter.Valid() {
		__antithesis_instrumentation__.Notify(96247)
		if err := fn(iter.CurApplied(), ctx); err != nil {
			__antithesis_instrumentation__.Notify(96249)
			return err
		} else {
			__antithesis_instrumentation__.Notify(96250)
		}
		__antithesis_instrumentation__.Notify(96248)
		iter.Next()
	}
	__antithesis_instrumentation__.Notify(96246)
	return nil
}
