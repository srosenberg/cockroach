package streamingccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type EventType int

const (
	KVEvent EventType = iota

	CheckpointEvent

	GenerationEvent
)

type Event interface {
	Type() EventType

	GetKV() *roachpb.KeyValue

	GetResolved() *hlc.Timestamp
}

type kvEvent struct {
	kv roachpb.KeyValue
}

var _ Event = kvEvent{}

func (kve kvEvent) Type() EventType {
	__antithesis_instrumentation__.Notify(24941)
	return KVEvent
}

func (kve kvEvent) GetKV() *roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(24942)
	return &kve.kv
}

func (kve kvEvent) GetResolved() *hlc.Timestamp {
	__antithesis_instrumentation__.Notify(24943)
	return nil
}

type checkpointEvent struct {
	resolvedTimestamp hlc.Timestamp
}

var _ Event = checkpointEvent{}

func (ce checkpointEvent) Type() EventType {
	__antithesis_instrumentation__.Notify(24944)
	return CheckpointEvent
}

func (ce checkpointEvent) GetKV() *roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(24945)
	return nil
}

func (ce checkpointEvent) GetResolved() *hlc.Timestamp {
	__antithesis_instrumentation__.Notify(24946)
	return &ce.resolvedTimestamp
}

type generationEvent struct{}

var _ Event = generationEvent{}

func (ge generationEvent) Type() EventType {
	__antithesis_instrumentation__.Notify(24947)
	return GenerationEvent
}

func (ge generationEvent) GetKV() *roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(24948)
	return nil
}

func (ge generationEvent) GetResolved() *hlc.Timestamp {
	__antithesis_instrumentation__.Notify(24949)
	return nil
}

func MakeKVEvent(kv roachpb.KeyValue) Event {
	__antithesis_instrumentation__.Notify(24950)
	return kvEvent{kv: kv}
}

func MakeCheckpointEvent(resolvedTimestamp hlc.Timestamp) Event {
	__antithesis_instrumentation__.Notify(24951)
	return checkpointEvent{resolvedTimestamp: resolvedTimestamp}
}

func MakeGenerationEvent() Event {
	__antithesis_instrumentation__.Notify(24952)
	return generationEvent{}
}
