package descmetadata

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessioninit"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
)

type metadataUpdater struct {
	txn               *kv.Txn
	ie                sqlutil.InternalExecutor
	collectionFactory *descs.CollectionFactory
	cacheEnabled      bool
}

func (mu metadataUpdater) UpsertDescriptorComment(
	id int64, subID int64, commentType keys.CommentType, comment string,
) error {
	__antithesis_instrumentation__.Notify(466012)
	_, err := mu.ie.ExecEx(context.Background(),
		fmt.Sprintf("upsert-%s-comment", commentType),
		mu.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"UPSERT INTO system.comments VALUES ($1, $2, $3, $4)",
		commentType,
		id,
		subID,
		comment,
	)
	return err
}

func (mu metadataUpdater) DeleteDescriptorComment(
	id int64, subID int64, commentType keys.CommentType,
) error {
	__antithesis_instrumentation__.Notify(466013)
	_, err := mu.ie.ExecEx(context.Background(),
		fmt.Sprintf("delete-%s-comment", commentType),
		mu.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"DELETE FROM system.comments WHERE object_id = $1 AND sub_id = $2 AND "+
			"type = $3;",
		id,
		subID,
		commentType,
	)
	return err
}

func (mu metadataUpdater) DeleteAllCommentsForTables(idSet catalog.DescriptorIDSet) error {
	__antithesis_instrumentation__.Notify(466014)
	if idSet.Empty() {
		__antithesis_instrumentation__.Notify(466017)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(466018)
	}
	__antithesis_instrumentation__.Notify(466015)
	var buf strings.Builder
	ids := idSet.Ordered()
	_, _ = fmt.Fprintf(&buf, `
DELETE FROM system.comments
      WHERE type IN (%d, %d, %d, %d)
        AND object_id IN (%d`,
		keys.TableCommentType, keys.ColumnCommentType, keys.ConstraintCommentType,
		keys.IndexCommentType, ids[0],
	)
	for _, id := range ids[1:] {
		__antithesis_instrumentation__.Notify(466019)
		_, _ = fmt.Fprintf(&buf, ", %d", id)
	}
	__antithesis_instrumentation__.Notify(466016)
	buf.WriteString(")")
	_, err := mu.ie.ExecEx(context.Background(),
		"delete-all-comments-for-tables",
		mu.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		buf.String(),
	)
	return err
}

func (mu metadataUpdater) UpsertConstraintComment(
	tableID descpb.ID, constraintID descpb.ConstraintID, comment string,
) error {
	__antithesis_instrumentation__.Notify(466020)
	return mu.UpsertDescriptorComment(int64(tableID), int64(constraintID), keys.ConstraintCommentType, comment)
}

func (mu metadataUpdater) DeleteConstraintComment(
	tableID descpb.ID, constraintID descpb.ConstraintID,
) error {
	__antithesis_instrumentation__.Notify(466021)
	return mu.DeleteDescriptorComment(int64(tableID), int64(constraintID), keys.ConstraintCommentType)
}

func (mu metadataUpdater) DeleteDatabaseRoleSettings(ctx context.Context, dbID descpb.ID) error {
	__antithesis_instrumentation__.Notify(466022)
	rowsDeleted, err := mu.ie.ExecEx(ctx,
		"delete-db-role-setting",
		mu.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf(
			`DELETE FROM %s WHERE database_id = $1`,
			sessioninit.DatabaseRoleSettingsTableName,
		),
		dbID,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(466025)
		return err
	} else {
		__antithesis_instrumentation__.Notify(466026)
	}
	__antithesis_instrumentation__.Notify(466023)

	if !mu.cacheEnabled || func() bool {
		__antithesis_instrumentation__.Notify(466027)
		return rowsDeleted == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(466028)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(466029)
	}
	__antithesis_instrumentation__.Notify(466024)

	return mu.collectionFactory.Txn(ctx,
		mu.ie,
		mu.txn.DB(),
		func(ctx context.Context, txn *kv.Txn, descriptors *descs.Collection) error {
			__antithesis_instrumentation__.Notify(466030)
			desc, err := descriptors.GetMutableTableByID(
				ctx,
				txn,
				keys.DatabaseRoleSettingsTableID,
				tree.ObjectLookupFlags{
					CommonLookupFlags: tree.CommonLookupFlags{
						Required:       true,
						RequireMutable: true,
					},
				})
			if err != nil {
				__antithesis_instrumentation__.Notify(466032)
				return err
			} else {
				__antithesis_instrumentation__.Notify(466033)
			}
			__antithesis_instrumentation__.Notify(466031)
			desc.MaybeIncrementVersion()
			return descriptors.WriteDesc(ctx, false, desc, txn)
		})
}

func (mu metadataUpdater) SwapDescriptorSubComment(
	id int64, oldSubID int64, newSubID int64, commentType keys.CommentType,
) error {
	__antithesis_instrumentation__.Notify(466034)
	_, err := mu.ie.ExecEx(context.Background(),
		fmt.Sprintf("upsert-%s-comment", commentType),
		mu.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"UPDATE system.comments  SET sub_id= $1 WHERE "+
			"object_id = $2 AND sub_id = $3 AND type = $4",
		newSubID,
		id,
		oldSubID,
		commentType,
	)
	return err
}

func (mu metadataUpdater) DeleteSchedule(ctx context.Context, scheduleID int64) error {
	__antithesis_instrumentation__.Notify(466035)
	_, err := mu.ie.ExecEx(
		ctx,
		"delete-schedule",
		mu.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"DELETE FROM system.scheduled_jobs WHERE schedule_id = $1",
		scheduleID,
	)
	return err
}
