package blobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"

	"github.com/cockroachdb/cockroach/pkg/blobs/blobspb"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
)

var chunkSize = 128 * 1 << 10

var _ ioctx.ReadCloserCtx = &blobStreamReader{}

type streamReceiver interface {
	SendAndClose(*blobspb.StreamResponse) error
	Recv() (*blobspb.StreamChunk, error)
}

type nopSendAndClose struct {
	blobspb.Blob_GetStreamClient
}

func (*nopSendAndClose) SendAndClose(*blobspb.StreamResponse) error {
	__antithesis_instrumentation__.Notify(3385)
	return nil
}

func newGetStreamReader(client blobspb.Blob_GetStreamClient) ioctx.ReadCloserCtx {
	__antithesis_instrumentation__.Notify(3386)
	return &blobStreamReader{
		stream: &nopSendAndClose{client},
	}
}

func newPutStreamReader(client blobspb.Blob_PutStreamServer) ioctx.ReadCloserCtx {
	__antithesis_instrumentation__.Notify(3387)
	return &blobStreamReader{stream: client}
}

type blobStreamReader struct {
	lastPayload []byte
	lastOffset  int
	stream      streamReceiver
	EOFReached  bool
}

func (r *blobStreamReader) Read(ctx context.Context, out []byte) (int, error) {
	__antithesis_instrumentation__.Notify(3388)
	if r.EOFReached {
		__antithesis_instrumentation__.Notify(3392)
		return 0, io.EOF
	} else {
		__antithesis_instrumentation__.Notify(3393)
	}
	__antithesis_instrumentation__.Notify(3389)

	offset := 0

	if r.lastPayload != nil {
		__antithesis_instrumentation__.Notify(3394)
		offset = len(r.lastPayload) - r.lastOffset
		if len(out) < offset {
			__antithesis_instrumentation__.Notify(3396)
			copy(out, r.lastPayload[r.lastOffset:])
			r.lastOffset += len(out)
			return len(out), nil
		} else {
			__antithesis_instrumentation__.Notify(3397)
		}
		__antithesis_instrumentation__.Notify(3395)
		copy(out[:offset], r.lastPayload[r.lastOffset:])
		r.lastPayload = nil
	} else {
		__antithesis_instrumentation__.Notify(3398)
	}
	__antithesis_instrumentation__.Notify(3390)
	for offset < len(out) {
		__antithesis_instrumentation__.Notify(3399)
		chunk, err := r.stream.Recv()
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(3403)
			r.EOFReached = true
			break
		} else {
			__antithesis_instrumentation__.Notify(3404)
		}
		__antithesis_instrumentation__.Notify(3400)
		if err != nil {
			__antithesis_instrumentation__.Notify(3405)
			return offset, err
		} else {
			__antithesis_instrumentation__.Notify(3406)
		}
		__antithesis_instrumentation__.Notify(3401)
		var lenToWrite int
		if len(out)-offset >= len(chunk.Payload) {
			__antithesis_instrumentation__.Notify(3407)
			lenToWrite = len(chunk.Payload)
		} else {
			__antithesis_instrumentation__.Notify(3408)
			lenToWrite = len(out) - offset

			r.lastPayload = chunk.Payload
			r.lastOffset = lenToWrite
		}
		__antithesis_instrumentation__.Notify(3402)
		copy(out[offset:offset+lenToWrite], chunk.Payload[:lenToWrite])
		offset += lenToWrite
	}
	__antithesis_instrumentation__.Notify(3391)
	return offset, nil
}

func (r *blobStreamReader) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(3409)
	return r.stream.SendAndClose(&blobspb.StreamResponse{})
}

type streamSender interface {
	Send(*blobspb.StreamChunk) error
}

func streamContent(ctx context.Context, sender streamSender, content ioctx.ReaderCtx) error {
	__antithesis_instrumentation__.Notify(3410)
	payload := make([]byte, chunkSize)
	var chunk blobspb.StreamChunk
	for {
		__antithesis_instrumentation__.Notify(3411)
		n, err := content.Read(ctx, payload)
		if n > 0 {
			__antithesis_instrumentation__.Notify(3414)
			chunk.Payload = payload[:n]
			err = sender.Send(&chunk)
		} else {
			__antithesis_instrumentation__.Notify(3415)
		}
		__antithesis_instrumentation__.Notify(3412)
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(3416)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(3417)
		}
		__antithesis_instrumentation__.Notify(3413)
		if err != nil {
			__antithesis_instrumentation__.Notify(3418)
			return err
		} else {
			__antithesis_instrumentation__.Notify(3419)
		}
	}
}
