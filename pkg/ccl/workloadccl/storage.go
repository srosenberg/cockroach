package workloadccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"

	_ "github.com/cockroachdb/cockroach/pkg/cloud/amazon"
	_ "github.com/cockroachdb/cockroach/pkg/cloud/azure"
	_ "github.com/cockroachdb/cockroach/pkg/cloud/gcp"
	"github.com/cockroachdb/cockroach/pkg/security"
	clustersettings "github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/errors"
)

const storageError = `failed to create google cloud client ` +
	`(You may need to setup the GCS application default credentials: ` +
	`'gcloud auth application-default login --project=cockroach-shared')`

func GetStorage(ctx context.Context, cfg FixtureConfig) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(27903)
	switch cfg.StorageProvider {
	case "gs", "s3", "azure":
		__antithesis_instrumentation__.Notify(27906)
	default:
		__antithesis_instrumentation__.Notify(27907)
		return nil, errors.AssertionFailedf("unsupported external storage provider; valid providers are gs, s3, and azure")
	}
	__antithesis_instrumentation__.Notify(27904)

	s, err := cloud.ExternalStorageFromURI(ctx, cfg.ObjectPathToURI(),
		base.ExternalIODirConfig{}, clustersettings.MakeClusterSettings(),
		nil, security.SQLUsername{}, nil, nil, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(27908)
		return nil, errors.Wrap(err, storageError)
	} else {
		__antithesis_instrumentation__.Notify(27909)
	}
	__antithesis_instrumentation__.Notify(27905)
	return s, nil
}
