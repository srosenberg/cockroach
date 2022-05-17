package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
)

var errTwoVersionInvariantViolated = errors.Errorf("two version invariant violated")

func (cf *CollectionFactory) Txn(
	ctx context.Context,
	ie sqlutil.InternalExecutor,
	db *kv.DB,
	f func(ctx context.Context, txn *kv.Txn, descriptors *Collection) error,
) error {
	__antithesis_instrumentation__.Notify(265009)

	waitForDescriptors := func(modifiedDescriptors []lease.IDVersion, deletedDescs catalog.DescriptorIDSet) error {
		__antithesis_instrumentation__.Notify(265011)

		for _, ld := range modifiedDescriptors {
			__antithesis_instrumentation__.Notify(265013)
			waitForNoVersion := deletedDescs.Contains(ld.ID)

			if waitForNoVersion {
				__antithesis_instrumentation__.Notify(265014)
				err := cf.leaseMgr.WaitForNoVersion(ctx, ld.ID, retry.Options{})
				if err != nil {
					__antithesis_instrumentation__.Notify(265015)
					return err
				} else {
					__antithesis_instrumentation__.Notify(265016)
				}
			} else {
				__antithesis_instrumentation__.Notify(265017)
				_, err := cf.leaseMgr.WaitForOneVersion(ctx, ld.ID, retry.Options{})
				if err != nil {
					__antithesis_instrumentation__.Notify(265018)
					return err
				} else {
					__antithesis_instrumentation__.Notify(265019)
				}
			}
		}
		__antithesis_instrumentation__.Notify(265012)
		return nil
	}
	__antithesis_instrumentation__.Notify(265010)
	for {
		__antithesis_instrumentation__.Notify(265020)
		var modifiedDescriptors []lease.IDVersion
		var deletedDescs catalog.DescriptorIDSet
		var descsCol Collection
		if err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(265021)
			modifiedDescriptors = nil
			deletedDescs = catalog.DescriptorIDSet{}
			descsCol = cf.MakeCollection(ctx, nil)
			defer descsCol.ReleaseAll(ctx)
			if !cf.settings.Version.IsActive(
				ctx, clusterversion.DisableSystemConfigGossipTrigger,
			) {
				__antithesis_instrumentation__.Notify(265027)
				if err := txn.DeprecatedSetSystemConfigTrigger(
					cf.leaseMgr.Codec().ForSystemTenant(),
				); err != nil {
					__antithesis_instrumentation__.Notify(265028)
					return err
				} else {
					__antithesis_instrumentation__.Notify(265029)
				}
			} else {
				__antithesis_instrumentation__.Notify(265030)
			}
			__antithesis_instrumentation__.Notify(265022)
			if err := f(ctx, txn, &descsCol); err != nil {
				__antithesis_instrumentation__.Notify(265031)
				return err
			} else {
				__antithesis_instrumentation__.Notify(265032)
			}
			__antithesis_instrumentation__.Notify(265023)

			if err := descsCol.ValidateUncommittedDescriptors(ctx, txn); err != nil {
				__antithesis_instrumentation__.Notify(265033)
				return err
			} else {
				__antithesis_instrumentation__.Notify(265034)
			}
			__antithesis_instrumentation__.Notify(265024)
			modifiedDescriptors = descsCol.GetDescriptorsWithNewVersion()

			if err := CheckSpanCountLimit(
				ctx, &descsCol, cf.spanConfigSplitter, cf.spanConfigLimiter, txn,
			); err != nil {
				__antithesis_instrumentation__.Notify(265035)
				return err
			} else {
				__antithesis_instrumentation__.Notify(265036)
			}
			__antithesis_instrumentation__.Notify(265025)
			retryErr, err := CheckTwoVersionInvariant(
				ctx, db.Clock(), ie, &descsCol, txn, nil)
			if retryErr {
				__antithesis_instrumentation__.Notify(265037)
				return errTwoVersionInvariantViolated
			} else {
				__antithesis_instrumentation__.Notify(265038)
			}
			__antithesis_instrumentation__.Notify(265026)
			deletedDescs = descsCol.deletedDescs
			return err
		}); errors.Is(err, errTwoVersionInvariantViolated) {
			__antithesis_instrumentation__.Notify(265039)
			continue
		} else {
			__antithesis_instrumentation__.Notify(265040)
			if err == nil {
				__antithesis_instrumentation__.Notify(265042)
				err = waitForDescriptors(modifiedDescriptors, deletedDescs)
			} else {
				__antithesis_instrumentation__.Notify(265043)
			}
			__antithesis_instrumentation__.Notify(265041)
			return err
		}
	}
}

func CheckTwoVersionInvariant(
	ctx context.Context,
	clock *hlc.Clock,
	ie sqlutil.InternalExecutor,
	descsCol *Collection,
	txn *kv.Txn,
	onRetryBackoff func(),
) (retryDueToViolation bool, _ error) {
	__antithesis_instrumentation__.Notify(265044)
	descs := descsCol.GetDescriptorsWithNewVersion()
	if descs == nil {
		__antithesis_instrumentation__.Notify(265050)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(265051)
	}
	__antithesis_instrumentation__.Notify(265045)
	if txn.IsCommitted() {
		__antithesis_instrumentation__.Notify(265052)
		panic("transaction has already committed")
	} else {
		__antithesis_instrumentation__.Notify(265053)
	}
	__antithesis_instrumentation__.Notify(265046)

	descsCol.ReleaseSpecifiedLeases(ctx, descs)

	count, err := lease.CountLeases(ctx, ie, descs, txn.ProvisionalCommitTimestamp())
	if err != nil {
		__antithesis_instrumentation__.Notify(265054)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(265055)
	}
	__antithesis_instrumentation__.Notify(265047)
	if count == 0 {
		__antithesis_instrumentation__.Notify(265056)

		return false, descsCol.MaybeUpdateDeadline(ctx, txn)
	} else {
		__antithesis_instrumentation__.Notify(265057)
	}
	__antithesis_instrumentation__.Notify(265048)

	retryErr := txn.PrepareRetryableError(ctx,
		fmt.Sprintf(
			`cannot publish new versions for descriptors: %v, old versions still in use`,
			descs))

	txn.CleanupOnError(ctx, retryErr)

	descsCol.ReleaseLeases(ctx)

	for r := retry.StartWithCtx(ctx, base.DefaultRetryOptions()); r.Next(); {
		__antithesis_instrumentation__.Notify(265058)

		now := clock.Now()
		count, err := lease.CountLeases(ctx, ie, descs, now)
		if err != nil {
			__antithesis_instrumentation__.Notify(265061)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(265062)
		}
		__antithesis_instrumentation__.Notify(265059)
		if count == 0 {
			__antithesis_instrumentation__.Notify(265063)
			break
		} else {
			__antithesis_instrumentation__.Notify(265064)
		}
		__antithesis_instrumentation__.Notify(265060)
		if onRetryBackoff != nil {
			__antithesis_instrumentation__.Notify(265065)
			onRetryBackoff()
		} else {
			__antithesis_instrumentation__.Notify(265066)
		}
	}
	__antithesis_instrumentation__.Notify(265049)
	return true, retryErr
}

func CheckSpanCountLimit(
	ctx context.Context,
	descsCol *Collection,
	splitter spanconfig.Splitter,
	limiter spanconfig.Limiter,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(265067)
	if !descsCol.codec().ForSystemTenant() {
		__antithesis_instrumentation__.Notify(265069)
		var totalSpanCountDelta int
		for _, ut := range descsCol.GetUncommittedTables() {
			__antithesis_instrumentation__.Notify(265072)
			uncommittedMutTable, err := descsCol.GetUncommittedMutableTableByID(ut.GetID())
			if err != nil {
				__antithesis_instrumentation__.Notify(265076)
				return err
			} else {
				__antithesis_instrumentation__.Notify(265077)
			}
			__antithesis_instrumentation__.Notify(265073)

			var originalTableDesc catalog.TableDescriptor
			if originalDesc := uncommittedMutTable.OriginalDescriptor(); originalDesc != nil {
				__antithesis_instrumentation__.Notify(265078)
				originalTableDesc = originalDesc.(catalog.TableDescriptor)
			} else {
				__antithesis_instrumentation__.Notify(265079)
			}
			__antithesis_instrumentation__.Notify(265074)
			delta, err := spanconfig.Delta(ctx, splitter, originalTableDesc, uncommittedMutTable)
			if err != nil {
				__antithesis_instrumentation__.Notify(265080)
				return err
			} else {
				__antithesis_instrumentation__.Notify(265081)
			}
			__antithesis_instrumentation__.Notify(265075)
			totalSpanCountDelta += delta
		}
		__antithesis_instrumentation__.Notify(265070)

		shouldLimit, err := limiter.ShouldLimit(ctx, txn, totalSpanCountDelta)
		if err != nil {
			__antithesis_instrumentation__.Notify(265082)
			return err
		} else {
			__antithesis_instrumentation__.Notify(265083)
		}
		__antithesis_instrumentation__.Notify(265071)
		if shouldLimit {
			__antithesis_instrumentation__.Notify(265084)
			return pgerror.New(pgcode.ConfigurationLimitExceeded, "exceeded limit for number of table spans")
		} else {
			__antithesis_instrumentation__.Notify(265085)
		}
	} else {
		__antithesis_instrumentation__.Notify(265086)
	}
	__antithesis_instrumentation__.Notify(265068)

	return nil
}
