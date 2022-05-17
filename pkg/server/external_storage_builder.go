package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/blobs"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/errors"
)

type externalStorageBuilder struct {
	conf              base.ExternalIODirConfig
	settings          *cluster.Settings
	blobClientFactory blobs.BlobClientFactory
	initCalled        bool
	ie                *sql.InternalExecutor
	db                *kv.DB
	limiters          cloud.Limiters
}

func (e *externalStorageBuilder) init(
	ctx context.Context,
	conf base.ExternalIODirConfig,
	settings *cluster.Settings,
	nodeIDContainer *base.NodeIDContainer,
	nodeDialer *nodedialer.Dialer,
	testingKnobs base.TestingKnobs,
	ie *sql.InternalExecutor,
	db *kv.DB,
) {
	var blobClientFactory blobs.BlobClientFactory
	if p, ok := testingKnobs.Server.(*TestingKnobs); ok && p.BlobClientFactory != nil {
		blobClientFactory = p.BlobClientFactory
	}
	if blobClientFactory == nil {
		blobClientFactory = blobs.NewBlobClientFactory(nodeIDContainer, nodeDialer, settings.ExternalIODir)
	}
	e.conf = conf
	e.settings = settings
	e.blobClientFactory = blobClientFactory
	e.initCalled = true
	e.ie = ie
	e.db = db
	e.limiters = cloud.MakeLimiters(ctx, &settings.SV)

}

func (e *externalStorageBuilder) makeExternalStorage(
	ctx context.Context, dest roachpb.ExternalStorage,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(193402)
	if !e.initCalled {
		__antithesis_instrumentation__.Notify(193404)
		return nil, errors.New("cannot create external storage before init")
	} else {
		__antithesis_instrumentation__.Notify(193405)
	}
	__antithesis_instrumentation__.Notify(193403)
	return cloud.MakeExternalStorage(ctx, dest, e.conf, e.settings, e.blobClientFactory, e.ie,
		e.db, e.limiters)
}

func (e *externalStorageBuilder) makeExternalStorageFromURI(
	ctx context.Context, uri string, user security.SQLUsername,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(193406)
	if !e.initCalled {
		__antithesis_instrumentation__.Notify(193408)
		return nil, errors.New("cannot create external storage before init")
	} else {
		__antithesis_instrumentation__.Notify(193409)
	}
	__antithesis_instrumentation__.Notify(193407)
	return cloud.ExternalStorageFromURI(ctx, uri, e.conf, e.settings, e.blobClientFactory, user, e.ie, e.db, e.limiters)
}
