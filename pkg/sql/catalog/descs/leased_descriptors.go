package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type leaseManager interface {
	AcquireByName(
		ctx context.Context,
		timestamp hlc.Timestamp,
		parentID descpb.ID,
		parentSchemaID descpb.ID,
		name string,
	) (lease.LeasedDescriptor, error)

	Acquire(
		ctx context.Context, timestamp hlc.Timestamp, id descpb.ID,
	) (lease.LeasedDescriptor, error)
}

type deadlineHolder interface {
	ReadTimestamp() hlc.Timestamp
	UpdateDeadline(ctx context.Context, deadline hlc.Timestamp) error
}

type maxTimestampBoundDeadlineHolder struct {
	maxTimestampBound hlc.Timestamp
}

func (m maxTimestampBoundDeadlineHolder) ReadTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(264668)

	return m.maxTimestampBound.Prev()
}

func (m maxTimestampBoundDeadlineHolder) UpdateDeadline(
	ctx context.Context, deadline hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(264669)
	return nil
}

func makeLeasedDescriptors(lm leaseManager) leasedDescriptors {
	__antithesis_instrumentation__.Notify(264670)
	return leasedDescriptors{
		lm: lm,
	}
}

type leasedDescriptors struct {
	lm    leaseManager
	cache nstree.Map
}

func (ld *leasedDescriptors) getByName(
	ctx context.Context,
	txn deadlineHolder,
	parentID descpb.ID,
	parentSchemaID descpb.ID,
	name string,
) (desc catalog.Descriptor, shouldReadFromStore bool, err error) {
	__antithesis_instrumentation__.Notify(264671)

	if cached := ld.cache.GetByName(parentID, parentSchemaID, name); cached != nil {
		__antithesis_instrumentation__.Notify(264674)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(264676)
			log.Eventf(ctx, "found descriptor in collection for (%d, %d, '%s'): %d",
				parentID, parentSchemaID, name, cached.GetID())
		} else {
			__antithesis_instrumentation__.Notify(264677)
		}
		__antithesis_instrumentation__.Notify(264675)
		return cached.(lease.LeasedDescriptor).Underlying(), false, nil
	} else {
		__antithesis_instrumentation__.Notify(264678)
	}
	__antithesis_instrumentation__.Notify(264672)

	if systemschema.IsUnleasableSystemDescriptorByName(parentID, parentSchemaID, name) {
		__antithesis_instrumentation__.Notify(264679)
		return nil, true, nil
	} else {
		__antithesis_instrumentation__.Notify(264680)
	}
	__antithesis_instrumentation__.Notify(264673)

	readTimestamp := txn.ReadTimestamp()
	ldesc, err := ld.lm.AcquireByName(ctx, readTimestamp, parentID, parentSchemaID, name)
	const setTxnDeadline = true
	return ld.getResult(ctx, txn, setTxnDeadline, ldesc, err)
}

func (ld *leasedDescriptors) getByID(
	ctx context.Context, txn deadlineHolder, id descpb.ID,
) (_ catalog.Descriptor, shouldReadFromStore bool, _ error) {
	__antithesis_instrumentation__.Notify(264681)

	if cached := ld.getCachedByID(ctx, id); cached != nil {
		__antithesis_instrumentation__.Notify(264684)
		return cached, false, nil
	} else {
		__antithesis_instrumentation__.Notify(264685)
	}
	__antithesis_instrumentation__.Notify(264682)

	if systemschema.IsUnleasableSystemDescriptorByID(id) {
		__antithesis_instrumentation__.Notify(264686)
		return nil, true, nil
	} else {
		__antithesis_instrumentation__.Notify(264687)
	}
	__antithesis_instrumentation__.Notify(264683)

	readTimestamp := txn.ReadTimestamp()
	desc, err := ld.lm.Acquire(ctx, readTimestamp, id)
	const setTxnDeadline = false
	return ld.getResult(ctx, txn, setTxnDeadline, desc, err)
}

func (ld *leasedDescriptors) getCachedByID(ctx context.Context, id descpb.ID) catalog.Descriptor {
	__antithesis_instrumentation__.Notify(264688)
	cached := ld.cache.GetByID(id)
	if cached == nil {
		__antithesis_instrumentation__.Notify(264691)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(264692)
	}
	__antithesis_instrumentation__.Notify(264689)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(264693)
		log.Eventf(ctx, "found descriptor in collection for (%d, %d, '%s'): %d",
			cached.GetParentID(), cached.GetParentSchemaID(), cached.GetName(), id)
	} else {
		__antithesis_instrumentation__.Notify(264694)
	}
	__antithesis_instrumentation__.Notify(264690)
	return cached.(lease.LeasedDescriptor).Underlying()
}

func (ld *leasedDescriptors) getResult(
	ctx context.Context,
	txn deadlineHolder,
	setDeadline bool,
	ldesc lease.LeasedDescriptor,
	err error,
) (_ catalog.Descriptor, shouldReadFromStore bool, _ error) {
	__antithesis_instrumentation__.Notify(264695)
	if err != nil {
		__antithesis_instrumentation__.Notify(264700)
		_, isBoundedStalenessRead := txn.(*maxTimestampBoundDeadlineHolder)

		if shouldReadFromStore =
			!isBoundedStalenessRead && func() bool {
				__antithesis_instrumentation__.Notify(264702)
				return ((catalog.HasInactiveDescriptorError(err) && func() bool {
					__antithesis_instrumentation__.Notify(264703)
					return errors.Is(err, catalog.ErrDescriptorDropped) == true
				}() == true) || func() bool {
					__antithesis_instrumentation__.Notify(264704)
					return errors.Is(err, catalog.ErrDescriptorNotFound) == true
				}() == true) == true
			}() == true; shouldReadFromStore {
			__antithesis_instrumentation__.Notify(264705)
			return nil, true, nil
		} else {
			__antithesis_instrumentation__.Notify(264706)
		}
		__antithesis_instrumentation__.Notify(264701)

		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(264707)
	}
	__antithesis_instrumentation__.Notify(264696)

	expiration := ldesc.Expiration()
	readTimestamp := txn.ReadTimestamp()
	if expiration.LessEq(txn.ReadTimestamp()) {
		__antithesis_instrumentation__.Notify(264708)
		log.Fatalf(ctx, "bad descriptor for T=%s, expiration=%s", readTimestamp, expiration)
	} else {
		__antithesis_instrumentation__.Notify(264709)
	}
	__antithesis_instrumentation__.Notify(264697)

	ld.cache.Upsert(ldesc)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(264710)
		log.Eventf(ctx, "added descriptor '%s' to collection: %+v", ldesc.GetName(), ldesc.Underlying())
	} else {
		__antithesis_instrumentation__.Notify(264711)
	}
	__antithesis_instrumentation__.Notify(264698)

	if setDeadline {
		__antithesis_instrumentation__.Notify(264712)
		if err := ld.maybeUpdateDeadline(ctx, txn, nil); err != nil {
			__antithesis_instrumentation__.Notify(264713)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(264714)
		}
	} else {
		__antithesis_instrumentation__.Notify(264715)
	}
	__antithesis_instrumentation__.Notify(264699)
	return ldesc.Underlying(), false, nil
}

func (ld *leasedDescriptors) maybeUpdateDeadline(
	ctx context.Context, txn deadlineHolder, session sqlliveness.Session,
) error {
	__antithesis_instrumentation__.Notify(264716)

	var deadline hlc.Timestamp
	if session != nil {
		__antithesis_instrumentation__.Notify(264720)
		if expiration, txnTS := session.Expiration(), txn.ReadTimestamp(); txnTS.Less(expiration) {
			__antithesis_instrumentation__.Notify(264721)
			deadline = expiration
		} else {
			__antithesis_instrumentation__.Notify(264722)

			return errors.Errorf(
				"liveness session expired %s before transaction",
				txnTS.GoTime().Sub(expiration.GoTime()),
			)
		}
	} else {
		__antithesis_instrumentation__.Notify(264723)
	}
	__antithesis_instrumentation__.Notify(264717)
	if leaseDeadline, ok := ld.getDeadline(); ok && func() bool {
		__antithesis_instrumentation__.Notify(264724)
		return (deadline.IsEmpty() || func() bool {
			__antithesis_instrumentation__.Notify(264725)
			return leaseDeadline.Less(deadline) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(264726)

		deadline = leaseDeadline
	} else {
		__antithesis_instrumentation__.Notify(264727)
	}
	__antithesis_instrumentation__.Notify(264718)

	if !deadline.IsEmpty() {
		__antithesis_instrumentation__.Notify(264728)
		return txn.UpdateDeadline(ctx, deadline)
	} else {
		__antithesis_instrumentation__.Notify(264729)
	}
	__antithesis_instrumentation__.Notify(264719)
	return nil
}

func (ld *leasedDescriptors) getDeadline() (deadline hlc.Timestamp, haveDeadline bool) {
	__antithesis_instrumentation__.Notify(264730)
	_ = ld.cache.IterateByID(func(descriptor catalog.NameEntry) error {
		__antithesis_instrumentation__.Notify(264732)
		expiration := descriptor.(lease.LeasedDescriptor).Expiration()
		if !haveDeadline || func() bool {
			__antithesis_instrumentation__.Notify(264734)
			return expiration.Less(deadline) == true
		}() == true {
			__antithesis_instrumentation__.Notify(264735)
			deadline, haveDeadline = expiration, true
		} else {
			__antithesis_instrumentation__.Notify(264736)
		}
		__antithesis_instrumentation__.Notify(264733)
		return nil
	})
	__antithesis_instrumentation__.Notify(264731)
	return deadline, haveDeadline
}

func (ld *leasedDescriptors) releaseAll(ctx context.Context) {
	__antithesis_instrumentation__.Notify(264737)
	log.VEventf(ctx, 2, "releasing %d descriptors", ld.numDescriptors())
	_ = ld.cache.IterateByID(func(descriptor catalog.NameEntry) error {
		__antithesis_instrumentation__.Notify(264739)
		descriptor.(lease.LeasedDescriptor).Release(ctx)
		return nil
	})
	__antithesis_instrumentation__.Notify(264738)
	ld.cache.Clear()
}

func (ld *leasedDescriptors) release(ctx context.Context, descs []lease.IDVersion) {
	__antithesis_instrumentation__.Notify(264740)
	for _, idv := range descs {
		__antithesis_instrumentation__.Notify(264741)
		if removed := ld.cache.Remove(idv.ID); removed != nil {
			__antithesis_instrumentation__.Notify(264742)
			removed.(lease.LeasedDescriptor).Release(ctx)
		} else {
			__antithesis_instrumentation__.Notify(264743)
		}
	}
}

func (ld *leasedDescriptors) numDescriptors() int {
	__antithesis_instrumentation__.Notify(264744)
	return ld.cache.Len()
}
