package catkv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func lookupDescriptorsUnvalidated(
	ctx context.Context, txn *kv.Txn, cq catalogQuerier, ids []descpb.ID,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265668)
	if len(ids) == 0 {
		__antithesis_instrumentation__.Notify(265673)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265674)
	}
	__antithesis_instrumentation__.Notify(265669)
	cb, err := cq.query(ctx, txn, func(codec keys.SQLCodec, b *kv.Batch) {
		__antithesis_instrumentation__.Notify(265675)
		for _, id := range ids {
			__antithesis_instrumentation__.Notify(265677)
			key := catalogkeys.MakeDescMetadataKey(codec, id)
			b.Get(key)
		}
		__antithesis_instrumentation__.Notify(265676)
		if log.ExpensiveLogEnabled(ctx, 2) {
			__antithesis_instrumentation__.Notify(265678)
			log.Infof(ctx, "looking up unvalidated descriptors by id: %v", ids)
		} else {
			__antithesis_instrumentation__.Notify(265679)
		}
	})
	__antithesis_instrumentation__.Notify(265670)
	if err != nil {
		__antithesis_instrumentation__.Notify(265680)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265681)
	}
	__antithesis_instrumentation__.Notify(265671)
	ret := make([]catalog.Descriptor, len(ids))
	for i, id := range ids {
		__antithesis_instrumentation__.Notify(265682)
		desc := cb.LookupDescriptorEntry(id)
		if desc == nil {
			__antithesis_instrumentation__.Notify(265684)
			if cq.isRequired {
				__antithesis_instrumentation__.Notify(265686)
				return nil, wrapError(cq.expectedType, id, requiredError(cq.expectedType, id))
			} else {
				__antithesis_instrumentation__.Notify(265687)
			}
			__antithesis_instrumentation__.Notify(265685)
			continue
		} else {
			__antithesis_instrumentation__.Notify(265688)
		}
		__antithesis_instrumentation__.Notify(265683)
		ret[i] = desc
	}
	__antithesis_instrumentation__.Notify(265672)
	return ret, nil
}

func lookupIDs(
	ctx context.Context, txn *kv.Txn, cq catalogQuerier, nameInfos []descpb.NameInfo,
) ([]descpb.ID, error) {
	__antithesis_instrumentation__.Notify(265689)
	if len(nameInfos) == 0 {
		__antithesis_instrumentation__.Notify(265694)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265695)
	}
	__antithesis_instrumentation__.Notify(265690)
	cb, err := cq.query(ctx, txn, func(codec keys.SQLCodec, b *kv.Batch) {
		__antithesis_instrumentation__.Notify(265696)
		for _, nameInfo := range nameInfos {
			__antithesis_instrumentation__.Notify(265697)
			if nameInfo.Name == "" {
				__antithesis_instrumentation__.Notify(265699)
				continue
			} else {
				__antithesis_instrumentation__.Notify(265700)
			}
			__antithesis_instrumentation__.Notify(265698)
			b.Get(catalogkeys.EncodeNameKey(codec, nameInfo))
		}
	})
	__antithesis_instrumentation__.Notify(265691)
	if err != nil {
		__antithesis_instrumentation__.Notify(265701)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265702)
	}
	__antithesis_instrumentation__.Notify(265692)
	ret := make([]descpb.ID, len(nameInfos))
	for i, nameInfo := range nameInfos {
		__antithesis_instrumentation__.Notify(265703)
		ne := cb.LookupNamespaceEntry(nameInfo)
		if ne == nil {
			__antithesis_instrumentation__.Notify(265705)
			if cq.isRequired {
				__antithesis_instrumentation__.Notify(265707)
				return nil, errors.AssertionFailedf("expected namespace entry for %s, none found", nameInfo.String())
			} else {
				__antithesis_instrumentation__.Notify(265708)
			}
			__antithesis_instrumentation__.Notify(265706)
			continue
		} else {
			__antithesis_instrumentation__.Notify(265709)
		}
		__antithesis_instrumentation__.Notify(265704)
		ret[i] = ne.GetID()
	}
	__antithesis_instrumentation__.Notify(265693)
	return ret, nil
}

type catalogQuerier struct {
	isRequired   bool
	expectedType catalog.DescriptorType
	codec        keys.SQLCodec
}

func (cq catalogQuerier) query(
	ctx context.Context, txn *kv.Txn, in func(codec keys.SQLCodec, b *kv.Batch),
) (nstree.Catalog, error) {
	__antithesis_instrumentation__.Notify(265710)
	b := txn.NewBatch()
	in(cq.codec, b)
	if err := txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(265713)
		return nstree.Catalog{}, err
	} else {
		__antithesis_instrumentation__.Notify(265714)
	}
	__antithesis_instrumentation__.Notify(265711)
	cb := &nstree.MutableCatalog{}
	for _, result := range b.Results {
		__antithesis_instrumentation__.Notify(265715)
		if result.Err != nil {
			__antithesis_instrumentation__.Notify(265717)
			return nstree.Catalog{}, result.Err
		} else {
			__antithesis_instrumentation__.Notify(265718)
		}
		__antithesis_instrumentation__.Notify(265716)
		for _, row := range result.Rows {
			__antithesis_instrumentation__.Notify(265719)
			_, catTableID, err := cq.codec.DecodeTablePrefix(row.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(265722)
				return nstree.Catalog{}, err
			} else {
				__antithesis_instrumentation__.Notify(265723)
			}
			__antithesis_instrumentation__.Notify(265720)
			switch catTableID {
			case keys.NamespaceTableID:
				__antithesis_instrumentation__.Notify(265724)
				err = cq.processNamespaceResultRow(row, cb)
			case keys.DescriptorTableID:
				__antithesis_instrumentation__.Notify(265725)
				err = cq.processDescriptorResultRow(row, cb)
			default:
				__antithesis_instrumentation__.Notify(265726)
				err = errors.AssertionFailedf("unexpected catalog key %s", row.Key.String())
			}
			__antithesis_instrumentation__.Notify(265721)
			if err != nil {
				__antithesis_instrumentation__.Notify(265727)
				return nstree.Catalog{}, err
			} else {
				__antithesis_instrumentation__.Notify(265728)
			}
		}
	}
	__antithesis_instrumentation__.Notify(265712)
	return cb.Catalog, nil
}

func (cq catalogQuerier) processNamespaceResultRow(
	row kv.KeyValue, cb *nstree.MutableCatalog,
) error {
	__antithesis_instrumentation__.Notify(265729)
	nameInfo, err := catalogkeys.DecodeNameMetadataKey(cq.codec, row.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(265732)
		return err
	} else {
		__antithesis_instrumentation__.Notify(265733)
	}
	__antithesis_instrumentation__.Notify(265730)
	if row.Exists() {
		__antithesis_instrumentation__.Notify(265734)
		cb.UpsertNamespaceEntry(nameInfo, descpb.ID(row.ValueInt()))
	} else {
		__antithesis_instrumentation__.Notify(265735)
	}
	__antithesis_instrumentation__.Notify(265731)
	return nil
}

func (cq catalogQuerier) processDescriptorResultRow(
	row kv.KeyValue, cb *nstree.MutableCatalog,
) error {
	__antithesis_instrumentation__.Notify(265736)
	u32ID, err := cq.codec.DecodeDescMetadataID(row.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(265739)
		return err
	} else {
		__antithesis_instrumentation__.Notify(265740)
	}
	__antithesis_instrumentation__.Notify(265737)
	id := descpb.ID(u32ID)
	desc, err := build(cq.expectedType, id, row.Value, cq.isRequired)
	if err != nil {
		__antithesis_instrumentation__.Notify(265741)
		return wrapError(cq.expectedType, id, err)
	} else {
		__antithesis_instrumentation__.Notify(265742)
	}
	__antithesis_instrumentation__.Notify(265738)
	cb.UpsertDescriptorEntry(desc)
	return nil
}

func wrapError(expectedType catalog.DescriptorType, id descpb.ID, err error) error {
	__antithesis_instrumentation__.Notify(265743)
	switch expectedType {
	case catalog.Table:
		__antithesis_instrumentation__.Notify(265745)
		return catalog.WrapTableDescRefErr(id, err)
	case catalog.Database:
		__antithesis_instrumentation__.Notify(265746)
		return catalog.WrapDatabaseDescRefErr(id, err)
	case catalog.Schema:
		__antithesis_instrumentation__.Notify(265747)
		return catalog.WrapSchemaDescRefErr(id, err)
	case catalog.Type:
		__antithesis_instrumentation__.Notify(265748)
		return catalog.WrapTypeDescRefErr(id, err)
	default:
		__antithesis_instrumentation__.Notify(265749)
	}
	__antithesis_instrumentation__.Notify(265744)
	return errors.Wrapf(err, "referenced descriptor ID %d", id)
}

func build(
	expectedType catalog.DescriptorType, id descpb.ID, rowValue *roachpb.Value, isRequired bool,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265750)
	var b catalog.DescriptorBuilder
	if rowValue != nil {
		__antithesis_instrumentation__.Notify(265756)
		var descProto descpb.Descriptor
		if err := rowValue.GetProto(&descProto); err != nil {
			__antithesis_instrumentation__.Notify(265758)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(265759)
		}
		__antithesis_instrumentation__.Notify(265757)
		b = descbuilder.NewBuilderWithMVCCTimestamp(&descProto, rowValue.Timestamp)
	} else {
		__antithesis_instrumentation__.Notify(265760)
	}
	__antithesis_instrumentation__.Notify(265751)
	if b == nil {
		__antithesis_instrumentation__.Notify(265761)
		if isRequired {
			__antithesis_instrumentation__.Notify(265763)
			return nil, requiredError(expectedType, id)
		} else {
			__antithesis_instrumentation__.Notify(265764)
		}
		__antithesis_instrumentation__.Notify(265762)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265765)
	}
	__antithesis_instrumentation__.Notify(265752)
	if expectedType != catalog.Any && func() bool {
		__antithesis_instrumentation__.Notify(265766)
		return b.DescriptorType() != expectedType == true
	}() == true {
		__antithesis_instrumentation__.Notify(265767)
		return nil, pgerror.Newf(pgcode.WrongObjectType, "descriptor is a %s", b.DescriptorType())
	} else {
		__antithesis_instrumentation__.Notify(265768)
	}
	__antithesis_instrumentation__.Notify(265753)
	if err := b.RunPostDeserializationChanges(); err != nil {
		__antithesis_instrumentation__.Notify(265769)
		return nil, errors.NewAssertionErrorWithWrappedErrf(err, "error during RunPostDeserializationChanges")
	} else {
		__antithesis_instrumentation__.Notify(265770)
	}
	__antithesis_instrumentation__.Notify(265754)
	desc := b.BuildImmutable()
	if id != desc.GetID() {
		__antithesis_instrumentation__.Notify(265771)
		return nil, errors.AssertionFailedf("unexpected ID %d in descriptor", desc.GetID())
	} else {
		__antithesis_instrumentation__.Notify(265772)
	}
	__antithesis_instrumentation__.Notify(265755)
	return desc, nil
}

func requiredError(expectedType catalog.DescriptorType, id descpb.ID) (err error) {
	__antithesis_instrumentation__.Notify(265773)
	switch expectedType {
	case catalog.Table:
		__antithesis_instrumentation__.Notify(265775)
		err = sqlerrors.NewUndefinedRelationError(&tree.TableRef{TableID: int64(id)})
	case catalog.Database:
		__antithesis_instrumentation__.Notify(265776)
		err = sqlerrors.NewUndefinedDatabaseError(fmt.Sprintf("[%d]", id))
	case catalog.Schema:
		__antithesis_instrumentation__.Notify(265777)
		err = sqlerrors.NewUndefinedSchemaError(fmt.Sprintf("[%d]", id))
	case catalog.Type:
		__antithesis_instrumentation__.Notify(265778)
		err = sqlerrors.NewUndefinedTypeError(tree.NewUnqualifiedTypeName(fmt.Sprintf("[%d]", id)))
	default:
		__antithesis_instrumentation__.Notify(265779)
		err = errors.Errorf("failed to find descriptor [%d]", id)
	}
	__antithesis_instrumentation__.Notify(265774)
	return errors.CombineErrors(catalog.ErrDescriptorNotFound, err)
}
