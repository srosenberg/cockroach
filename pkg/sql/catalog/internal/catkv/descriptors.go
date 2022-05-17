package catkv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/validate"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func GetCatalogUnvalidated(
	ctx context.Context, codec keys.SQLCodec, txn *kv.Txn,
) (nstree.Catalog, error) {
	__antithesis_instrumentation__.Notify(265780)
	cq := catalogQuerier{
		isRequired:   true,
		expectedType: catalog.Any,
		codec:        codec,
	}
	log.Eventf(ctx, "fetching all descriptors and namespace entries")
	return cq.query(ctx, txn, func(codec keys.SQLCodec, b *kv.Batch) {
		__antithesis_instrumentation__.Notify(265781)
		b.Header.MaxSpanRequestKeys = 0
		descsPrefix := catalogkeys.MakeAllDescsMetadataKey(codec)
		b.Scan(descsPrefix, descsPrefix.PrefixEnd())
		nsPrefix := codec.IndexPrefix(
			uint32(systemschema.NamespaceTable.GetID()),
			uint32(systemschema.NamespaceTable.GetPrimaryIndexID()))
		b.Scan(nsPrefix, nsPrefix.PrefixEnd())
	})
}

func lookupDescriptorsAndValidate(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	txn *kv.Txn,
	cq catalogQuerier,
	vd validate.ValidationDereferencer,
	ids []descpb.ID,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265782)
	descs, err := lookupDescriptorsUnvalidated(ctx, txn, cq, ids)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(265786)
		return len(descs) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(265787)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265788)
	}
	__antithesis_instrumentation__.Notify(265783)
	if vd == nil {
		__antithesis_instrumentation__.Notify(265789)
		vd = &readValidationDereferencer{
			catalogQuerier: catalogQuerier{
				expectedType: catalog.Any,
				codec:        cq.codec,
			},
			txn: txn,
		}
	} else {
		__antithesis_instrumentation__.Notify(265790)
	}
	__antithesis_instrumentation__.Notify(265784)
	ve := validate.Validate(ctx, version, vd, catalog.ValidationReadTelemetry, catalog.ValidationLevelCrossReferences, descs...)
	if err := ve.CombinedError(); err != nil {
		__antithesis_instrumentation__.Notify(265791)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265792)
	}
	__antithesis_instrumentation__.Notify(265785)
	return descs, nil
}

type readValidationDereferencer struct {
	catalogQuerier
	txn *kv.Txn
}

var _ validate.ValidationDereferencer = (*readValidationDereferencer)(nil)

func (t *readValidationDereferencer) DereferenceDescriptors(
	ctx context.Context, version clusterversion.ClusterVersion, reqs []descpb.ID,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265793)
	return GetCrossReferencedDescriptorsForValidation(ctx, version, t.codec, t.txn, reqs)
}

func (t *readValidationDereferencer) DereferenceDescriptorIDs(
	ctx context.Context, requests []descpb.NameInfo,
) ([]descpb.ID, error) {
	__antithesis_instrumentation__.Notify(265794)
	return LookupIDs(ctx, t.txn, t.codec, requests)
}

func GetCrossReferencedDescriptorsForValidation(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	codec keys.SQLCodec,
	txn *kv.Txn,
	ids []descpb.ID,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265795)
	cq := catalogQuerier{
		expectedType: catalog.Any,
		codec:        codec,
	}
	descs, err := lookupDescriptorsUnvalidated(ctx, txn, cq, ids)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(265798)
		return len(descs) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(265799)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265800)
	}
	__antithesis_instrumentation__.Notify(265796)
	if err := validate.Self(version, descs...); err != nil {
		__antithesis_instrumentation__.Notify(265801)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265802)
	}
	__antithesis_instrumentation__.Notify(265797)
	return descs, nil
}

func MaybeGetDescriptorByID(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	codec keys.SQLCodec,
	txn *kv.Txn,
	vd validate.ValidationDereferencer,
	id descpb.ID,
	expectedType catalog.DescriptorType,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265803)
	cq := catalogQuerier{
		expectedType: expectedType,
		codec:        codec,
	}
	descs, err := lookupDescriptorsAndValidate(ctx, version, txn, cq, vd, []descpb.ID{id})
	if err != nil {
		__antithesis_instrumentation__.Notify(265805)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265806)
	}
	__antithesis_instrumentation__.Notify(265804)
	return descs[0], nil
}

func MustGetDescriptorsByID(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	codec keys.SQLCodec,
	txn *kv.Txn,
	vd validate.ValidationDereferencer,
	ids []descpb.ID,
	expectedType catalog.DescriptorType,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265807)
	cq := catalogQuerier{
		codec:        codec,
		isRequired:   true,
		expectedType: expectedType,
	}
	return lookupDescriptorsAndValidate(ctx, version, txn, cq, vd, ids)
}

func MustGetDescriptorByID(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	codec keys.SQLCodec,
	txn *kv.Txn,
	vd validate.ValidationDereferencer,
	id descpb.ID,
	expectedType catalog.DescriptorType,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265808)
	descs, err := MustGetDescriptorsByID(ctx, version, codec, txn, vd, []descpb.ID{id}, expectedType)
	if err != nil {
		__antithesis_instrumentation__.Notify(265810)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265811)
	}
	__antithesis_instrumentation__.Notify(265809)
	return descs[0], err
}
