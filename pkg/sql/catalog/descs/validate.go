package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/catkv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/validate"
)

func (tc *Collection) Validate(
	ctx context.Context,
	txn *kv.Txn,
	telemetry catalog.ValidationTelemetry,
	targetLevel catalog.ValidationLevel,
	descriptors ...catalog.Descriptor,
) (err error) {
	__antithesis_instrumentation__.Notify(265239)
	vd := tc.newValidationDereferencer(txn)
	version := tc.settings.Version.ActiveVersion(ctx)
	return validate.Validate(
		ctx,
		version,
		vd,
		telemetry,
		targetLevel,
		descriptors...).CombinedError()
}

func (tc *Collection) ValidateUncommittedDescriptors(ctx context.Context, txn *kv.Txn) (err error) {
	__antithesis_instrumentation__.Notify(265240)
	if tc.skipValidationOnWrite || func() bool {
		__antithesis_instrumentation__.Notify(265243)
		return !ValidateOnWriteEnabled.Get(&tc.settings.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(265244)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(265245)
	}
	__antithesis_instrumentation__.Notify(265241)
	descs := tc.uncommitted.getUncommittedDescriptorsForValidation()
	if len(descs) == 0 {
		__antithesis_instrumentation__.Notify(265246)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(265247)
	}
	__antithesis_instrumentation__.Notify(265242)
	return tc.Validate(ctx, txn, catalog.ValidationWriteTelemetry, catalog.ValidationLevelAllPreTxnCommit, descs...)
}

func (tc *Collection) newValidationDereferencer(txn *kv.Txn) validate.ValidationDereferencer {
	__antithesis_instrumentation__.Notify(265248)
	return &collectionBackedDereferencer{tc: tc, txn: txn}
}

type collectionBackedDereferencer struct {
	tc  *Collection
	txn *kv.Txn
}

var _ validate.ValidationDereferencer = &collectionBackedDereferencer{}

func (c collectionBackedDereferencer) DereferenceDescriptors(
	ctx context.Context, version clusterversion.ClusterVersion, reqs []descpb.ID,
) (ret []catalog.Descriptor, _ error) {
	__antithesis_instrumentation__.Notify(265249)
	ret = make([]catalog.Descriptor, len(reqs))
	fallbackReqs := make([]descpb.ID, 0, len(reqs))
	fallbackRetIndexes := make([]int, 0, len(reqs))
	for i, id := range reqs {
		__antithesis_instrumentation__.Notify(265252)
		desc, err := c.fastDescLookup(ctx, id)
		if err != nil {
			__antithesis_instrumentation__.Notify(265254)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(265255)
		}
		__antithesis_instrumentation__.Notify(265253)
		if desc == nil {
			__antithesis_instrumentation__.Notify(265256)
			fallbackReqs = append(fallbackReqs, id)
			fallbackRetIndexes = append(fallbackRetIndexes, i)
		} else {
			__antithesis_instrumentation__.Notify(265257)
			ret[i] = desc
		}
	}
	__antithesis_instrumentation__.Notify(265250)
	if len(fallbackReqs) > 0 {
		__antithesis_instrumentation__.Notify(265258)
		fallbackRet, err := catkv.GetCrossReferencedDescriptorsForValidation(
			ctx,
			version,
			c.tc.codec(),
			c.txn,
			fallbackReqs,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(265260)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(265261)
		}
		__antithesis_instrumentation__.Notify(265259)
		for j, desc := range fallbackRet {
			__antithesis_instrumentation__.Notify(265262)
			if desc == nil {
				__antithesis_instrumentation__.Notify(265265)
				continue
			} else {
				__antithesis_instrumentation__.Notify(265266)
			}
			__antithesis_instrumentation__.Notify(265263)
			if uc, _ := c.tc.uncommitted.getImmutableByID(desc.GetID()); uc == nil {
				__antithesis_instrumentation__.Notify(265267)
				desc, err = c.tc.uncommitted.add(desc.NewBuilder().BuildExistingMutable(), notValidatedYet)
				if err != nil {
					__antithesis_instrumentation__.Notify(265268)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(265269)
				}
			} else {
				__antithesis_instrumentation__.Notify(265270)
			}
			__antithesis_instrumentation__.Notify(265264)
			ret[fallbackRetIndexes[j]] = desc
		}
	} else {
		__antithesis_instrumentation__.Notify(265271)
	}
	__antithesis_instrumentation__.Notify(265251)
	return ret, nil
}

func (c collectionBackedDereferencer) fastDescLookup(
	ctx context.Context, id descpb.ID,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265272)
	if uc, _ := c.tc.uncommitted.getImmutableByID(id); uc != nil {
		__antithesis_instrumentation__.Notify(265274)
		return uc, nil
	} else {
		__antithesis_instrumentation__.Notify(265275)
	}
	__antithesis_instrumentation__.Notify(265273)
	return nil, nil
}

func (c collectionBackedDereferencer) DereferenceDescriptorIDs(
	ctx context.Context, reqs []descpb.NameInfo,
) (ret []descpb.ID, _ error) {
	__antithesis_instrumentation__.Notify(265276)
	ret = make([]descpb.ID, len(reqs))
	fallbackReqs := make([]descpb.NameInfo, 0, len(reqs))
	fallbackRetIndexes := make([]int, 0, len(reqs))
	for i, req := range reqs {
		__antithesis_instrumentation__.Notify(265279)
		found, id, err := c.fastNamespaceLookup(ctx, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(265281)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(265282)
		}
		__antithesis_instrumentation__.Notify(265280)
		if found {
			__antithesis_instrumentation__.Notify(265283)
			ret[i] = id
		} else {
			__antithesis_instrumentation__.Notify(265284)
			fallbackReqs = append(fallbackReqs, req)
			fallbackRetIndexes = append(fallbackRetIndexes, i)
		}
	}
	__antithesis_instrumentation__.Notify(265277)
	if len(fallbackReqs) > 0 {
		__antithesis_instrumentation__.Notify(265285)

		fallbackRet, err := catkv.LookupIDs(ctx, c.txn, c.tc.codec(), fallbackReqs)
		if err != nil {
			__antithesis_instrumentation__.Notify(265287)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(265288)
		}
		__antithesis_instrumentation__.Notify(265286)
		for j, id := range fallbackRet {
			__antithesis_instrumentation__.Notify(265289)
			ret[fallbackRetIndexes[j]] = id
		}
	} else {
		__antithesis_instrumentation__.Notify(265290)
	}
	__antithesis_instrumentation__.Notify(265278)
	return ret, nil
}

func (c collectionBackedDereferencer) fastNamespaceLookup(
	ctx context.Context, req descpb.NameInfo,
) (found bool, id descpb.ID, err error) {
	__antithesis_instrumentation__.Notify(265291)

	switch req.ParentID {
	case descpb.InvalidID:
		__antithesis_instrumentation__.Notify(265293)
		if req.ParentSchemaID == descpb.InvalidID && func() bool {
			__antithesis_instrumentation__.Notify(265296)
			return req.Name == catconstants.SystemDatabaseName == true
		}() == true {
			__antithesis_instrumentation__.Notify(265297)

			return true, keys.SystemDatabaseID, nil
		} else {
			__antithesis_instrumentation__.Notify(265298)
		}
	case keys.SystemDatabaseID:
		__antithesis_instrumentation__.Notify(265294)

		id = c.tc.kv.systemNamespace.lookup(req.ParentSchemaID, req.Name)
		return id != descpb.InvalidID, id, nil
	default:
		__antithesis_instrumentation__.Notify(265295)
	}
	__antithesis_instrumentation__.Notify(265292)
	return false, descpb.InvalidID, nil
}
