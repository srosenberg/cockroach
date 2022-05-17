package descmetadata

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sessioninit"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
)

type MetadataUpdaterFactory struct {
	ieFactory         sqlutil.SessionBoundInternalExecutorFactory
	collectionFactory *descs.CollectionFactory
	settings          *settings.Values
}

func NewMetadataUpdaterFactory(
	ieFactory sqlutil.SessionBoundInternalExecutorFactory,
	collectionFactory *descs.CollectionFactory,
	settings *settings.Values,
) scexec.DescriptorMetadataUpdaterFactory {
	__antithesis_instrumentation__.Notify(466036)
	return MetadataUpdaterFactory{
		ieFactory:         ieFactory,
		collectionFactory: collectionFactory,
		settings:          settings,
	}
}

func (mf MetadataUpdaterFactory) NewMetadataUpdater(
	ctx context.Context, txn *kv.Txn, sessionData *sessiondata.SessionData,
) scexec.DescriptorMetadataUpdater {
	__antithesis_instrumentation__.Notify(466037)

	modifiedSessionData := sessionData.Clone()
	modifiedSessionData.ExperimentalDistSQLPlanningMode = sessiondatapb.ExperimentalDistSQLPlanningOn
	return metadataUpdater{
		txn:               txn,
		ie:                mf.ieFactory(ctx, modifiedSessionData),
		collectionFactory: mf.collectionFactory,
		cacheEnabled:      sessioninit.CacheEnabled.Get(mf.settings),
	}
}
