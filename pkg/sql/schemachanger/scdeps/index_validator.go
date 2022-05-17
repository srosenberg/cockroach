package scdeps

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
)

type ValidateForwardIndexesFn func(
	ctx context.Context,
	tbl catalog.TableDescriptor,
	indexes []catalog.Index,
	runHistoricalTxn sqlutil.HistoricalInternalExecTxnRunner,
	withFirstMutationPublic bool,
	gatherAllInvalid bool,
	execOverride sessiondata.InternalExecutorOverride,
) error

type ValidateInvertedIndexesFn func(
	ctx context.Context,
	codec keys.SQLCodec,
	tbl catalog.TableDescriptor,
	indexes []catalog.Index,
	runHistoricalTxn sqlutil.HistoricalInternalExecTxnRunner,
	withFirstMutationPublic bool,
	gatherAllInvalid bool,
	execOverride sessiondata.InternalExecutorOverride,
) error

type NewFakeSessionDataFn func(sv *settings.Values) *sessiondata.SessionData

type indexValidator struct {
	db                      *kv.DB
	codec                   keys.SQLCodec
	settings                *cluster.Settings
	ieFactory               sqlutil.SessionBoundInternalExecutorFactory
	validateForwardIndexes  ValidateForwardIndexesFn
	validateInvertedIndexes ValidateInvertedIndexesFn
	newFakeSessionData      NewFakeSessionDataFn
}

func (iv indexValidator) ValidateForwardIndexes(
	ctx context.Context,
	tbl catalog.TableDescriptor,
	indexes []catalog.Index,
	override sessiondata.InternalExecutorOverride,
) error {
	__antithesis_instrumentation__.Notify(580719)

	txnRunner := func(ctx context.Context, fn sqlutil.InternalExecFn) error {
		__antithesis_instrumentation__.Notify(580721)
		validationTxn := iv.db.NewTxn(ctx, "validation")
		err := validationTxn.SetFixedTimestamp(ctx, iv.db.Clock().Now())
		if err != nil {
			__antithesis_instrumentation__.Notify(580723)
			return err
		} else {
			__antithesis_instrumentation__.Notify(580724)
		}
		__antithesis_instrumentation__.Notify(580722)
		return fn(ctx, validationTxn, iv.ieFactory(ctx, iv.newFakeSessionData(&iv.settings.SV)))
	}
	__antithesis_instrumentation__.Notify(580720)
	const withFirstMutationPublic = true
	const gatherAllInvalid = false
	return iv.validateForwardIndexes(ctx, tbl, indexes, txnRunner, withFirstMutationPublic, gatherAllInvalid, override)
}

func (iv indexValidator) ValidateInvertedIndexes(
	ctx context.Context,
	tbl catalog.TableDescriptor,
	indexes []catalog.Index,
	override sessiondata.InternalExecutorOverride,
) error {
	__antithesis_instrumentation__.Notify(580725)

	txnRunner := func(ctx context.Context, fn sqlutil.InternalExecFn) error {
		__antithesis_instrumentation__.Notify(580727)
		validationTxn := iv.db.NewTxn(ctx, "validation")
		err := validationTxn.SetFixedTimestamp(ctx, iv.db.Clock().Now())
		if err != nil {
			__antithesis_instrumentation__.Notify(580729)
			return err
		} else {
			__antithesis_instrumentation__.Notify(580730)
		}
		__antithesis_instrumentation__.Notify(580728)
		return fn(ctx, validationTxn, iv.ieFactory(ctx, iv.newFakeSessionData(&iv.settings.SV)))
	}
	__antithesis_instrumentation__.Notify(580726)
	const withFirstMutationPublic = true
	const gatherAllInvalid = false
	return iv.validateInvertedIndexes(ctx, iv.codec, tbl, indexes, txnRunner, withFirstMutationPublic, gatherAllInvalid, override)
}

func NewIndexValidator(
	db *kv.DB,
	codec keys.SQLCodec,
	settings *cluster.Settings,
	ieFactory sqlutil.SessionBoundInternalExecutorFactory,
	validateForwardIndexes ValidateForwardIndexesFn,
	validateInvertedIndexes ValidateInvertedIndexesFn,
	newFakeSessionData NewFakeSessionDataFn,
) scexec.IndexValidator {
	__antithesis_instrumentation__.Notify(580731)
	return indexValidator{
		db:                      db,
		codec:                   codec,
		settings:                settings,
		ieFactory:               ieFactory,
		validateForwardIndexes:  validateForwardIndexes,
		validateInvertedIndexes: validateInvertedIndexes,
		newFakeSessionData:      newFakeSessionData,
	}
}
