package streamingccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streampb"
	"github.com/cockroachdb/cockroach/pkg/streaming"
)

type StreamStatusErr struct {
	StreamID     streaming.StreamID
	StreamStatus streampb.StreamReplicationStatus_StreamStatus
}

func NewStreamStatusErr(
	streamID streaming.StreamID, streamStatus streampb.StreamReplicationStatus_StreamStatus,
) StreamStatusErr {
	__antithesis_instrumentation__.Notify(24939)
	return StreamStatusErr{
		StreamID:     streamID,
		StreamStatus: streamStatus,
	}
}

func (e StreamStatusErr) Error() string {
	__antithesis_instrumentation__.Notify(24940)
	return fmt.Sprintf("replication stream %d is not running, status is %s", e.StreamID, e.StreamStatus)
}
