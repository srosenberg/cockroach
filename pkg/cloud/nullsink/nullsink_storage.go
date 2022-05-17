package nullsink

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
)

func parseNullURL(_ cloud.ExternalStorageURIContext, _ *url.URL) (roachpb.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36614)
	return roachpb.ExternalStorage{Provider: roachpb.ExternalStorageProvider_null}, nil
}

func MakeNullSinkStorageURI(path string) string {
	__antithesis_instrumentation__.Notify(36615)
	return fmt.Sprintf("null:///%s", path)
}

type nullSinkStorage struct {
}

var _ cloud.ExternalStorage = &nullSinkStorage{}

func makeNullSinkStorage(
	_ context.Context, _ cloud.ExternalStorageContext, _ roachpb.ExternalStorage,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36616)
	telemetry.Count("external-io.nullsink")
	return &nullSinkStorage{}, nil
}

func (n *nullSinkStorage) Close() error {
	__antithesis_instrumentation__.Notify(36617)
	return nil
}

func (n *nullSinkStorage) Conf() roachpb.ExternalStorage {
	__antithesis_instrumentation__.Notify(36618)
	return roachpb.ExternalStorage{Provider: roachpb.ExternalStorageProvider_null}
}

func (n *nullSinkStorage) ExternalIOConf() base.ExternalIODirConfig {
	__antithesis_instrumentation__.Notify(36619)
	return base.ExternalIODirConfig{}
}

func (n *nullSinkStorage) Settings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(36620)
	return nil
}

func (n *nullSinkStorage) ReadFile(
	ctx context.Context, basename string,
) (ioctx.ReadCloserCtx, error) {
	__antithesis_instrumentation__.Notify(36621)
	reader, _, err := n.ReadFileAt(ctx, basename, 0)
	return reader, err
}

func (n *nullSinkStorage) ReadFileAt(
	_ context.Context, _ string, _ int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(36622)
	return nil, 0, io.EOF
}

type nullWriter struct{}

func (nullWriter) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(36623)
	return len(p), nil
}
func (nullWriter) Close() error { __antithesis_instrumentation__.Notify(36624); return nil }

func (n *nullSinkStorage) Writer(_ context.Context, _ string) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(36625)
	return nullWriter{}, nil
}

func (n *nullSinkStorage) List(_ context.Context, _, _ string, _ cloud.ListingFn) error {
	__antithesis_instrumentation__.Notify(36626)
	return nil
}

func (n *nullSinkStorage) Delete(_ context.Context, _ string) error {
	__antithesis_instrumentation__.Notify(36627)
	return nil
}

func (n *nullSinkStorage) Size(_ context.Context, _ string) (int64, error) {
	__antithesis_instrumentation__.Notify(36628)
	return 0, nil
}

var _ cloud.ExternalStorage = &nullSinkStorage{}

func init() {
	cloud.RegisterExternalStorageProvider(roachpb.ExternalStorageProvider_null,
		parseNullURL, makeNullSinkStorage, cloud.RedactedParams(), "null")
}
