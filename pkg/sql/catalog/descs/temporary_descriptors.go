package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

type temporaryDescriptors struct {
	settings *cluster.Settings
	codec    keys.SQLCodec
	tsp      TemporarySchemaProvider
}

func makeTemporaryDescriptors(
	settings *cluster.Settings, codec keys.SQLCodec, temporarySchemaProvider TemporarySchemaProvider,
) temporaryDescriptors {
	__antithesis_instrumentation__.Notify(264982)
	return temporaryDescriptors{
		settings: settings,
		codec:    codec,
		tsp:      temporarySchemaProvider,
	}
}

type TemporarySchemaProvider interface {
	GetTemporarySchemaName() string
	GetTemporarySchemaIDForDB(descpb.ID) (descpb.ID, bool)
	MaybeGetDatabaseForTemporarySchemaID(descpb.ID) (descpb.ID, bool)
}

type temporarySchemaProviderImpl sessiondata.Stack

var _ TemporarySchemaProvider = (*temporarySchemaProviderImpl)(nil)

func NewTemporarySchemaProvider(sds *sessiondata.Stack) TemporarySchemaProvider {
	__antithesis_instrumentation__.Notify(264983)
	return (*temporarySchemaProviderImpl)(sds)
}

func (impl *temporarySchemaProviderImpl) GetTemporarySchemaName() string {
	__antithesis_instrumentation__.Notify(264984)
	return (*sessiondata.Stack)(impl).Top().SearchPath.GetTemporarySchemaName()
}

func (impl *temporarySchemaProviderImpl) GetTemporarySchemaIDForDB(id descpb.ID) (descpb.ID, bool) {
	__antithesis_instrumentation__.Notify(264985)
	ret, found := (*sessiondata.Stack)(impl).Top().GetTemporarySchemaIDForDB(uint32(id))
	return descpb.ID(ret), found
}

func (impl *temporarySchemaProviderImpl) MaybeGetDatabaseForTemporarySchemaID(
	id descpb.ID,
) (descpb.ID, bool) {
	__antithesis_instrumentation__.Notify(264986)
	ret, found := (*sessiondata.Stack)(impl).Top().MaybeGetDatabaseForTemporarySchemaID(uint32(id))
	return descpb.ID(ret), found
}

func (td *temporaryDescriptors) getSchemaByName(
	ctx context.Context, dbID descpb.ID, schemaName string,
) (avoidFurtherLookups bool, _ catalog.SchemaDescriptor) {
	__antithesis_instrumentation__.Notify(264987)

	if tsp := td.tsp; tsp != nil {
		__antithesis_instrumentation__.Notify(264990)
		if schemaName == catconstants.PgTempSchemaName || func() bool {
			__antithesis_instrumentation__.Notify(264991)
			return schemaName == tsp.GetTemporarySchemaName() == true
		}() == true {
			__antithesis_instrumentation__.Notify(264992)
			schemaID, found := tsp.GetTemporarySchemaIDForDB(dbID)
			if !found {
				__antithesis_instrumentation__.Notify(264994)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(264995)
			}
			__antithesis_instrumentation__.Notify(264993)
			return true, schemadesc.NewTemporarySchema(
				tsp.GetTemporarySchemaName(),
				schemaID,
				dbID,
			)
		} else {
			__antithesis_instrumentation__.Notify(264996)
		}
	} else {
		__antithesis_instrumentation__.Notify(264997)
	}
	__antithesis_instrumentation__.Notify(264988)
	if !td.settings.Version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) {
		__antithesis_instrumentation__.Notify(264998)

		if schemaName == tree.PublicSchema {
			__antithesis_instrumentation__.Notify(264999)
			return true, schemadesc.NewTemporarySchema(
				schemaName,
				keys.PublicSchemaID,
				dbID,
			)
		} else {
			__antithesis_instrumentation__.Notify(265000)
		}
	} else {
		__antithesis_instrumentation__.Notify(265001)
	}
	__antithesis_instrumentation__.Notify(264989)
	return false, nil
}

func (td *temporaryDescriptors) getSchemaByID(
	ctx context.Context, schemaID descpb.ID,
) catalog.SchemaDescriptor {
	__antithesis_instrumentation__.Notify(265002)
	tsp := td.tsp
	if tsp == nil {
		__antithesis_instrumentation__.Notify(265005)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(265006)
	}
	__antithesis_instrumentation__.Notify(265003)
	if dbID, exists := tsp.MaybeGetDatabaseForTemporarySchemaID(schemaID); exists {
		__antithesis_instrumentation__.Notify(265007)
		return schemadesc.NewTemporarySchema(
			tsp.GetTemporarySchemaName(),
			schemaID,
			dbID,
		)
	} else {
		__antithesis_instrumentation__.Notify(265008)
	}
	__antithesis_instrumentation__.Notify(265004)
	return nil
}
