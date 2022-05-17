package blobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/blobs/blobspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc/metadata"
)

type BlobClient interface {
	ReadFile(ctx context.Context, file string, offset int64) (ioctx.ReadCloserCtx, int64, error)

	Writer(ctx context.Context, file string) (io.WriteCloser, error)

	List(ctx context.Context, pattern string) ([]string, error)

	Delete(ctx context.Context, file string) error

	Stat(ctx context.Context, file string) (*blobspb.BlobStat, error)
}

var _ BlobClient = &remoteClient{}

type remoteClient struct {
	blobClient blobspb.BlobClient
}

func newRemoteClient(blobClient blobspb.BlobClient) BlobClient {
	__antithesis_instrumentation__.Notify(3198)
	return &remoteClient{blobClient: blobClient}
}

func (c *remoteClient) ReadFile(
	ctx context.Context, file string, offset int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(3199)

	st, err := c.Stat(ctx, file)
	if err != nil {
		__antithesis_instrumentation__.Notify(3201)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(3202)
	}
	__antithesis_instrumentation__.Notify(3200)
	stream, err := c.blobClient.GetStream(ctx, &blobspb.GetRequest{
		Filename: file,
		Offset:   offset,
	})
	return newGetStreamReader(stream), st.Filesize, errors.Wrap(err, "fetching file")
}

type streamWriter struct {
	s   blobspb.Blob_PutStreamClient
	buf blobspb.StreamChunk
}

func (w *streamWriter) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(3203)
	n := 0
	for len(p) > 0 {
		__antithesis_instrumentation__.Notify(3205)
		l := copy(w.buf.Payload[:cap(w.buf.Payload)], p)
		w.buf.Payload = w.buf.Payload[:l]
		p = p[l:]
		if l > 0 {
			__antithesis_instrumentation__.Notify(3207)
			if err := w.s.Send(&w.buf); err != nil {
				__antithesis_instrumentation__.Notify(3208)
				return n, err
			} else {
				__antithesis_instrumentation__.Notify(3209)
			}
		} else {
			__antithesis_instrumentation__.Notify(3210)
		}
		__antithesis_instrumentation__.Notify(3206)
		n += l
	}
	__antithesis_instrumentation__.Notify(3204)
	return n, nil
}

func (w *streamWriter) Close() error {
	__antithesis_instrumentation__.Notify(3211)
	_, err := w.s.CloseAndRecv()
	return err
}

func (c *remoteClient) Writer(ctx context.Context, file string) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(3212)
	ctx = metadata.AppendToOutgoingContext(ctx, "filename", file)
	stream, err := c.blobClient.PutStream(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(3214)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(3215)
	}
	__antithesis_instrumentation__.Notify(3213)
	buf := make([]byte, 0, chunkSize)
	return &streamWriter{s: stream, buf: blobspb.StreamChunk{Payload: buf}}, nil
}

func (c *remoteClient) List(ctx context.Context, pattern string) ([]string, error) {
	__antithesis_instrumentation__.Notify(3216)
	resp, err := c.blobClient.List(ctx, &blobspb.GlobRequest{
		Pattern: pattern,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(3218)
		return nil, errors.Wrap(err, "fetching list")
	} else {
		__antithesis_instrumentation__.Notify(3219)
	}
	__antithesis_instrumentation__.Notify(3217)
	return resp.Files, nil
}

func (c *remoteClient) Delete(ctx context.Context, file string) error {
	__antithesis_instrumentation__.Notify(3220)
	_, err := c.blobClient.Delete(ctx, &blobspb.DeleteRequest{
		Filename: file,
	})
	return err
}

func (c *remoteClient) Stat(ctx context.Context, file string) (*blobspb.BlobStat, error) {
	__antithesis_instrumentation__.Notify(3221)
	resp, err := c.blobClient.Stat(ctx, &blobspb.StatRequest{
		Filename: file,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(3223)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(3224)
	}
	__antithesis_instrumentation__.Notify(3222)
	return resp, nil
}

var _ BlobClient = &localClient{}

type localClient struct {
	localStorage *LocalStorage
}

func NewLocalClient(externalIODir string) (BlobClient, error) {
	__antithesis_instrumentation__.Notify(3225)
	storage, err := NewLocalStorage(externalIODir)
	if err != nil {
		__antithesis_instrumentation__.Notify(3227)
		return nil, errors.Wrap(err, "creating local client")
	} else {
		__antithesis_instrumentation__.Notify(3228)
	}
	__antithesis_instrumentation__.Notify(3226)
	return &localClient{localStorage: storage}, nil
}

func (c *localClient) ReadFile(
	ctx context.Context, file string, offset int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(3229)
	return c.localStorage.ReadFile(file, offset)
}

func (c *localClient) Writer(ctx context.Context, file string) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(3230)
	return c.localStorage.Writer(ctx, file)
}

func (c *localClient) List(ctx context.Context, pattern string) ([]string, error) {
	__antithesis_instrumentation__.Notify(3231)
	return c.localStorage.List(pattern)
}

func (c *localClient) Delete(ctx context.Context, file string) error {
	__antithesis_instrumentation__.Notify(3232)
	return c.localStorage.Delete(file)
}

func (c *localClient) Stat(ctx context.Context, file string) (*blobspb.BlobStat, error) {
	__antithesis_instrumentation__.Notify(3233)
	return c.localStorage.Stat(file)
}

type BlobClientFactory func(ctx context.Context, dialing roachpb.NodeID) (BlobClient, error)

func NewBlobClientFactory(
	localNodeIDContainer *base.NodeIDContainer, dialer *nodedialer.Dialer, externalIODir string,
) BlobClientFactory {
	__antithesis_instrumentation__.Notify(3234)
	return func(ctx context.Context, dialing roachpb.NodeID) (BlobClient, error) {
		__antithesis_instrumentation__.Notify(3235)
		localNodeID := localNodeIDContainer.Get()
		if dialing == 0 || func() bool {
			__antithesis_instrumentation__.Notify(3238)
			return localNodeID == dialing == true
		}() == true {
			__antithesis_instrumentation__.Notify(3239)
			return NewLocalClient(externalIODir)
		} else {
			__antithesis_instrumentation__.Notify(3240)
		}
		__antithesis_instrumentation__.Notify(3236)
		conn, err := dialer.Dial(ctx, dialing, rpc.DefaultClass)
		if err != nil {
			__antithesis_instrumentation__.Notify(3241)
			return nil, errors.Wrapf(err, "connecting to node %d", dialing)
		} else {
			__antithesis_instrumentation__.Notify(3242)
		}
		__antithesis_instrumentation__.Notify(3237)
		return newRemoteClient(blobspb.NewBlobClient(conn)), nil
	}
}

func NewLocalOnlyBlobClientFactory(externalIODir string) BlobClientFactory {
	__antithesis_instrumentation__.Notify(3243)
	return func(ctx context.Context, dialing roachpb.NodeID) (BlobClient, error) {
		__antithesis_instrumentation__.Notify(3244)
		if dialing == 0 {
			__antithesis_instrumentation__.Notify(3246)
			return NewLocalClient(externalIODir)
		} else {
			__antithesis_instrumentation__.Notify(3247)
		}
		__antithesis_instrumentation__.Notify(3245)
		return nil, errors.Errorf("connecting to remote node not supported")
	}
}
