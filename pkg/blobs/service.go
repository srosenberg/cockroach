/*
Package blobs contains a gRPC service to be used for remote file access.

It is used for bulk file reads and writes to files on any CockroachDB node.
Each node will run a blob service, which serves the file access for files on
that node. Each node will also have a blob client, which uses the nodedialer
to connect to another node's blob service, and access its files. The blob client
is the point of entry to this service and it supports the `BlobClient` interface,
which includes the following functionalities:
  - ReadFile
  - WriteFile
  - List
  - Delete
  - Stat
*/
package blobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"

	"github.com/cockroachdb/cockroach/pkg/blobs/blobspb"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Service struct {
	localStorage *LocalStorage
}

var _ blobspb.BlobServer = &Service{}

func NewBlobService(externalIODir string) (*Service, error) {
	__antithesis_instrumentation__.Notify(3360)
	localStorage, err := NewLocalStorage(externalIODir)
	return &Service{localStorage: localStorage}, err
}

func (s *Service) GetStream(req *blobspb.GetRequest, stream blobspb.Blob_GetStreamServer) error {
	__antithesis_instrumentation__.Notify(3361)
	content, _, err := s.localStorage.ReadFile(req.Filename, req.Offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(3363)
		return err
	} else {
		__antithesis_instrumentation__.Notify(3364)
	}
	__antithesis_instrumentation__.Notify(3362)
	defer content.Close(stream.Context())
	return streamContent(stream.Context(), stream, content)
}

func (s *Service) PutStream(stream blobspb.Blob_PutStreamServer) error {
	__antithesis_instrumentation__.Notify(3365)
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		__antithesis_instrumentation__.Notify(3370)
		return errors.New("could not fetch metadata")
	} else {
		__antithesis_instrumentation__.Notify(3371)
	}
	__antithesis_instrumentation__.Notify(3366)
	filename := md.Get("filename")
	if len(filename) < 1 || func() bool {
		__antithesis_instrumentation__.Notify(3372)
		return filename[0] == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(3373)
		return errors.New("no filename in metadata")
	} else {
		__antithesis_instrumentation__.Notify(3374)
	}
	__antithesis_instrumentation__.Notify(3367)
	reader := newPutStreamReader(stream)
	defer reader.Close(stream.Context())
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	w, err := s.localStorage.Writer(ctx, filename[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(3375)
		cancel()
		return err
	} else {
		__antithesis_instrumentation__.Notify(3376)
	}
	__antithesis_instrumentation__.Notify(3368)
	if _, err := io.Copy(w, ioctx.ReaderCtxAdapter(stream.Context(), reader)); err != nil {
		__antithesis_instrumentation__.Notify(3377)
		cancel()
		return errors.CombineErrors(w.Close(), err)
	} else {
		__antithesis_instrumentation__.Notify(3378)
	}
	__antithesis_instrumentation__.Notify(3369)
	err = w.Close()
	cancel()
	return err
}

func (s *Service) List(
	ctx context.Context, req *blobspb.GlobRequest,
) (*blobspb.GlobResponse, error) {
	__antithesis_instrumentation__.Notify(3379)
	matches, err := s.localStorage.List(req.Pattern)
	return &blobspb.GlobResponse{Files: matches}, err
}

func (s *Service) Delete(
	ctx context.Context, req *blobspb.DeleteRequest,
) (*blobspb.DeleteResponse, error) {
	__antithesis_instrumentation__.Notify(3380)
	return &blobspb.DeleteResponse{}, s.localStorage.Delete(req.Filename)
}

func (s *Service) Stat(ctx context.Context, req *blobspb.StatRequest) (*blobspb.BlobStat, error) {
	__antithesis_instrumentation__.Notify(3381)
	resp, err := s.localStorage.Stat(req.Filename)
	if oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(3383)

		return nil, status.Error(codes.NotFound, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(3384)
	}
	__antithesis_instrumentation__.Notify(3382)
	return resp, err
}
