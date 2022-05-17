package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	pbtypes "github.com/gogo/protobuf/types"
)

func validateCheckExpr(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
	exprStr string,
	tableDesc *tabledesc.Mutable,
	ie sqlutil.InternalExecutor,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(272377)
	expr, err := schemaexpr.FormatExprForDisplay(ctx, tableDesc, exprStr, semaCtx, sessionData, tree.FmtParsable)
	if err != nil {
		__antithesis_instrumentation__.Notify(272381)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272382)
	}
	__antithesis_instrumentation__.Notify(272378)
	colSelectors := tabledesc.ColumnsSelectors(tableDesc.AccessibleColumns())
	columns := tree.AsStringWithFlags(&colSelectors, tree.FmtSerializable)
	queryStr := fmt.Sprintf(`SELECT %s FROM [%d AS t] WHERE NOT (%s) LIMIT 1`, columns, tableDesc.GetID(), exprStr)
	log.Infof(ctx, "validating check constraint %q with query %q", expr, queryStr)

	rows, err := ie.QueryRow(ctx, "validate check constraint", txn, queryStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(272383)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272384)
	}
	__antithesis_instrumentation__.Notify(272379)
	if rows.Len() > 0 {
		__antithesis_instrumentation__.Notify(272385)
		return pgerror.Newf(pgcode.CheckViolation,
			"validation of CHECK %q failed on row: %s",
			expr, labeledRowValues(tableDesc.AccessibleColumns(), rows))
	} else {
		__antithesis_instrumentation__.Notify(272386)
	}
	__antithesis_instrumentation__.Notify(272380)
	return nil
}

func matchFullUnacceptableKeyQuery(
	srcTbl catalog.TableDescriptor, fk *descpb.ForeignKeyConstraint, limitResults bool,
) (sql string, colNames []string, _ error) {
	__antithesis_instrumentation__.Notify(272387)
	nCols := len(fk.OriginColumnIDs)
	srcCols := make([]string, nCols)
	srcNullExistsClause := make([]string, nCols)
	srcNotNullExistsClause := make([]string, nCols)

	returnedCols := srcCols
	for i := 0; i < nCols; i++ {
		__antithesis_instrumentation__.Notify(272391)
		col, err := srcTbl.FindColumnWithID(fk.OriginColumnIDs[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(272393)
			return "", nil, err
		} else {
			__antithesis_instrumentation__.Notify(272394)
		}
		__antithesis_instrumentation__.Notify(272392)
		srcCols[i] = tree.NameString(col.GetName())
		srcNullExistsClause[i] = fmt.Sprintf("%s IS NULL", srcCols[i])
		srcNotNullExistsClause[i] = fmt.Sprintf("%s IS NOT NULL", srcCols[i])
	}
	__antithesis_instrumentation__.Notify(272388)

	for i := 0; i < srcTbl.GetPrimaryIndex().NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(272395)
		id := srcTbl.GetPrimaryIndex().GetKeyColumnID(i)
		alreadyPresent := false
		for _, otherID := range fk.OriginColumnIDs {
			__antithesis_instrumentation__.Notify(272397)
			if id == otherID {
				__antithesis_instrumentation__.Notify(272398)
				alreadyPresent = true
				break
			} else {
				__antithesis_instrumentation__.Notify(272399)
			}
		}
		__antithesis_instrumentation__.Notify(272396)
		if !alreadyPresent {
			__antithesis_instrumentation__.Notify(272400)
			col, err := tabledesc.FindPublicColumnWithID(srcTbl, id)
			if err != nil {
				__antithesis_instrumentation__.Notify(272402)
				return "", nil, err
			} else {
				__antithesis_instrumentation__.Notify(272403)
			}
			__antithesis_instrumentation__.Notify(272401)
			returnedCols = append(returnedCols, col.GetName())
		} else {
			__antithesis_instrumentation__.Notify(272404)
		}
	}
	__antithesis_instrumentation__.Notify(272389)

	limit := ""
	if limitResults {
		__antithesis_instrumentation__.Notify(272405)
		limit = " LIMIT 1"
	} else {
		__antithesis_instrumentation__.Notify(272406)
	}
	__antithesis_instrumentation__.Notify(272390)
	return fmt.Sprintf(
		`SELECT %[1]s FROM [%[2]d AS tbl] WHERE (%[3]s) AND (%[4]s) %[5]s`,
		strings.Join(returnedCols, ","),
		srcTbl.GetID(),
		strings.Join(srcNullExistsClause, " OR "),
		strings.Join(srcNotNullExistsClause, " OR "),
		limit,
	), returnedCols, nil
}

func nonMatchingRowQuery(
	srcTbl catalog.TableDescriptor,
	fk *descpb.ForeignKeyConstraint,
	targetTbl catalog.TableDescriptor,
	limitResults bool,
) (sql string, originColNames []string, _ error) {
	__antithesis_instrumentation__.Notify(272407)
	originColNames, err := srcTbl.NamesForColumnIDs(fk.OriginColumnIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(272414)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(272415)
	}
	__antithesis_instrumentation__.Notify(272408)

	for i := 0; i < srcTbl.GetPrimaryIndex().NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(272416)
		pkColID := srcTbl.GetPrimaryIndex().GetKeyColumnID(i)
		found := false
		for _, id := range fk.OriginColumnIDs {
			__antithesis_instrumentation__.Notify(272418)
			if pkColID == id {
				__antithesis_instrumentation__.Notify(272419)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(272420)
			}
		}
		__antithesis_instrumentation__.Notify(272417)
		if !found {
			__antithesis_instrumentation__.Notify(272421)
			column, err := tabledesc.FindPublicColumnWithID(srcTbl, pkColID)
			if err != nil {
				__antithesis_instrumentation__.Notify(272423)
				return "", nil, err
			} else {
				__antithesis_instrumentation__.Notify(272424)
			}
			__antithesis_instrumentation__.Notify(272422)
			originColNames = append(originColNames, column.GetName())
		} else {
			__antithesis_instrumentation__.Notify(272425)
		}
	}
	__antithesis_instrumentation__.Notify(272409)
	srcCols := make([]string, len(originColNames))
	qualifiedSrcCols := make([]string, len(originColNames))
	for i, n := range originColNames {
		__antithesis_instrumentation__.Notify(272426)
		srcCols[i] = tree.NameString(n)

		qualifiedSrcCols[i] = fmt.Sprintf("s.%s", srcCols[i])
	}
	__antithesis_instrumentation__.Notify(272410)

	referencedColNames, err := targetTbl.NamesForColumnIDs(fk.ReferencedColumnIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(272427)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(272428)
	}
	__antithesis_instrumentation__.Notify(272411)
	nCols := len(fk.OriginColumnIDs)
	srcWhere := make([]string, nCols)
	targetCols := make([]string, nCols)
	on := make([]string, nCols)

	for i := 0; i < nCols; i++ {
		__antithesis_instrumentation__.Notify(272429)

		srcWhere[i] = fmt.Sprintf("%s IS NOT NULL", srcCols[i])
		targetCols[i] = fmt.Sprintf("t.%s", tree.NameString(referencedColNames[i]))
		on[i] = fmt.Sprintf("%s = %s", qualifiedSrcCols[i], targetCols[i])
	}
	__antithesis_instrumentation__.Notify(272412)

	limit := ""
	if limitResults {
		__antithesis_instrumentation__.Notify(272430)
		limit = " LIMIT 1"
	} else {
		__antithesis_instrumentation__.Notify(272431)
	}
	__antithesis_instrumentation__.Notify(272413)
	return fmt.Sprintf(
		`SELECT %[1]s FROM 
		  (SELECT %[2]s FROM [%[3]d AS src]@{IGNORE_FOREIGN_KEYS} WHERE %[4]s) AS s
			LEFT OUTER JOIN
			[%[5]d AS target] AS t
			ON %[6]s
		 WHERE %[7]s IS NULL %[8]s`,
		strings.Join(qualifiedSrcCols, ", "),
		strings.Join(srcCols, ", "),
		srcTbl.GetID(),
		strings.Join(srcWhere, " AND "),
		targetTbl.GetID(),
		strings.Join(on, " AND "),

		targetCols[0],
		limit,
	), originColNames, nil
}

func validateForeignKey(
	ctx context.Context,
	srcTable *tabledesc.Mutable,
	targetTable catalog.TableDescriptor,
	fk *descpb.ForeignKeyConstraint,
	ie sqlutil.InternalExecutor,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(272432)
	nCols := len(fk.OriginColumnIDs)

	referencedColumnNames, err := targetTable.NamesForColumnIDs(fk.ReferencedColumnIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(272438)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272439)
	}
	__antithesis_instrumentation__.Notify(272433)

	if nCols > 1 && func() bool {
		__antithesis_instrumentation__.Notify(272440)
		return fk.Match == descpb.ForeignKeyReference_FULL == true
	}() == true {
		__antithesis_instrumentation__.Notify(272441)
		query, colNames, err := matchFullUnacceptableKeyQuery(
			srcTable, fk, true,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(272444)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272445)
		}
		__antithesis_instrumentation__.Notify(272442)

		log.Infof(ctx, "validating MATCH FULL FK %q (%q [%v] -> %q [%v]) with query %q",
			fk.Name,
			srcTable.Name, colNames,
			targetTable.GetName(), referencedColumnNames,
			query,
		)

		values, err := ie.QueryRowEx(ctx, "validate foreign key constraint",
			txn, sessiondata.NodeUserSessionDataOverride, query)
		if err != nil {
			__antithesis_instrumentation__.Notify(272446)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272447)
		}
		__antithesis_instrumentation__.Notify(272443)
		if values.Len() > 0 {
			__antithesis_instrumentation__.Notify(272448)
			return pgerror.WithConstraintName(pgerror.Newf(pgcode.ForeignKeyViolation,
				"foreign key violation: MATCH FULL does not allow mixing of null and nonnull values %s for %s",
				formatValues(colNames, values), fk.Name,
			), fk.Name)
		} else {
			__antithesis_instrumentation__.Notify(272449)
		}
	} else {
		__antithesis_instrumentation__.Notify(272450)
	}
	__antithesis_instrumentation__.Notify(272434)
	query, colNames, err := nonMatchingRowQuery(
		srcTable, fk, targetTable,
		true,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(272451)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272452)
	}
	__antithesis_instrumentation__.Notify(272435)

	log.Infof(ctx, "validating FK %q (%q [%v] -> %q [%v]) with query %q",
		fk.Name,
		srcTable.Name, colNames, targetTable.GetName(), referencedColumnNames,
		query,
	)

	values, err := ie.QueryRowEx(ctx, "validate fk constraint", txn,
		sessiondata.NodeUserSessionDataOverride, query)
	if err != nil {
		__antithesis_instrumentation__.Notify(272453)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272454)
	}
	__antithesis_instrumentation__.Notify(272436)
	if values.Len() > 0 {
		__antithesis_instrumentation__.Notify(272455)
		return pgerror.WithConstraintName(pgerror.Newf(pgcode.ForeignKeyViolation,
			"foreign key violation: %q row %s has no match in %q",
			srcTable.Name, formatValues(colNames, values), targetTable.GetName()), fk.Name)
	} else {
		__antithesis_instrumentation__.Notify(272456)
	}
	__antithesis_instrumentation__.Notify(272437)
	return nil
}

func duplicateRowQuery(
	srcTbl catalog.TableDescriptor, columnIDs []descpb.ColumnID, pred string, limitResults bool,
) (sql string, colNames []string, _ error) {
	__antithesis_instrumentation__.Notify(272457)
	colNames, err := srcTbl.NamesForColumnIDs(columnIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(272463)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(272464)
	}
	__antithesis_instrumentation__.Notify(272458)

	srcCols := make([]string, len(colNames))
	for i, n := range colNames {
		__antithesis_instrumentation__.Notify(272465)
		srcCols[i] = tree.NameString(n)
	}
	__antithesis_instrumentation__.Notify(272459)

	srcWhere := make([]string, 0, len(srcCols)+1)
	for i := range srcCols {
		__antithesis_instrumentation__.Notify(272466)
		srcWhere = append(srcWhere, fmt.Sprintf("%s IS NOT NULL", srcCols[i]))
	}
	__antithesis_instrumentation__.Notify(272460)

	if pred != "" {
		__antithesis_instrumentation__.Notify(272467)
		srcWhere = append(srcWhere, fmt.Sprintf("(%s)", pred))
	} else {
		__antithesis_instrumentation__.Notify(272468)
	}
	__antithesis_instrumentation__.Notify(272461)

	limit := ""
	if limitResults {
		__antithesis_instrumentation__.Notify(272469)
		limit = " LIMIT 1"
	} else {
		__antithesis_instrumentation__.Notify(272470)
	}
	__antithesis_instrumentation__.Notify(272462)
	return fmt.Sprintf(
		`SELECT %[1]s FROM [%[2]d AS tbl] WHERE %[3]s GROUP BY %[1]s HAVING count(*) > 1 %[4]s`,
		strings.Join(srcCols, ", "),
		srcTbl.GetID(),
		strings.Join(srcWhere, " AND "),
		limit,
	), colNames, nil
}

func (p *planner) RevalidateUniqueConstraintsInCurrentDB(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(272471)
	dbName := p.CurrentDatabase()
	log.Infof(ctx, "validating unique constraints in database %s", dbName)
	db, err := p.Descriptors().GetImmutableDatabaseByName(
		ctx, p.Txn(), dbName, tree.DatabaseLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(272475)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272476)
	}
	__antithesis_instrumentation__.Notify(272472)
	tableDescs, err := p.Descriptors().GetAllTableDescriptorsInDatabase(ctx, p.Txn(), db.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(272477)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272478)
	}
	__antithesis_instrumentation__.Notify(272473)

	for _, tableDesc := range tableDescs {
		__antithesis_instrumentation__.Notify(272479)
		if err = RevalidateUniqueConstraintsInTable(ctx, p.Txn(), p.ExecCfg().InternalExecutor, tableDesc); err != nil {
			__antithesis_instrumentation__.Notify(272480)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272481)
		}
	}
	__antithesis_instrumentation__.Notify(272474)
	return nil
}

func (p *planner) RevalidateUniqueConstraintsInTable(ctx context.Context, tableID int) error {
	__antithesis_instrumentation__.Notify(272482)
	tableDesc, err := p.Descriptors().GetImmutableTableByID(
		ctx,
		p.Txn(),
		descpb.ID(tableID),
		tree.ObjectLookupFlagsWithRequired(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(272484)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272485)
	}
	__antithesis_instrumentation__.Notify(272483)
	return RevalidateUniqueConstraintsInTable(ctx, p.Txn(), p.ExecCfg().InternalExecutor, tableDesc)
}

func (p *planner) RevalidateUniqueConstraint(
	ctx context.Context, tableID int, constraintName string,
) error {
	__antithesis_instrumentation__.Notify(272486)
	tableDesc, err := p.Descriptors().GetImmutableTableByID(
		ctx,
		p.Txn(),
		descpb.ID(tableID),
		tree.ObjectLookupFlagsWithRequired(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(272490)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272491)
	}
	__antithesis_instrumentation__.Notify(272487)

	for _, index := range tableDesc.ActiveIndexes() {
		__antithesis_instrumentation__.Notify(272492)
		if index.GetName() == constraintName {
			__antithesis_instrumentation__.Notify(272493)
			if !index.IsUnique() {
				__antithesis_instrumentation__.Notify(272496)
				return errors.Newf("%s is not a unique constraint", constraintName)
			} else {
				__antithesis_instrumentation__.Notify(272497)
			}
			__antithesis_instrumentation__.Notify(272494)
			if index.GetPartitioning().NumImplicitColumns() > 0 {
				__antithesis_instrumentation__.Notify(272498)
				return validateUniqueConstraint(
					ctx,
					tableDesc,
					index.GetName(),
					index.IndexDesc().KeyColumnIDs[index.GetPartitioning().NumImplicitColumns():],
					index.GetPredicate(),
					p.ExecCfg().InternalExecutor,
					p.Txn(),
					true,
				)
			} else {
				__antithesis_instrumentation__.Notify(272499)
			}
			__antithesis_instrumentation__.Notify(272495)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(272500)
		}
	}
	__antithesis_instrumentation__.Notify(272488)

	for _, uc := range tableDesc.GetUniqueWithoutIndexConstraints() {
		__antithesis_instrumentation__.Notify(272501)
		if uc.Name == constraintName {
			__antithesis_instrumentation__.Notify(272502)
			return validateUniqueConstraint(
				ctx,
				tableDesc,
				uc.Name,
				uc.ColumnIDs,
				uc.Predicate,
				p.ExecCfg().InternalExecutor,
				p.Txn(),
				true,
			)
		} else {
			__antithesis_instrumentation__.Notify(272503)
		}
	}
	__antithesis_instrumentation__.Notify(272489)

	return errors.Newf("unique constraint %s does not exist", constraintName)
}

func HasVirtualUniqueConstraints(tableDesc catalog.TableDescriptor) bool {
	__antithesis_instrumentation__.Notify(272504)
	for _, index := range tableDesc.ActiveIndexes() {
		__antithesis_instrumentation__.Notify(272507)
		if index.IsUnique() && func() bool {
			__antithesis_instrumentation__.Notify(272508)
			return index.GetPartitioning().NumImplicitColumns() > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(272509)
			return true
		} else {
			__antithesis_instrumentation__.Notify(272510)
		}
	}
	__antithesis_instrumentation__.Notify(272505)
	for _, uc := range tableDesc.GetUniqueWithoutIndexConstraints() {
		__antithesis_instrumentation__.Notify(272511)
		if uc.Validity == descpb.ConstraintValidity_Validated {
			__antithesis_instrumentation__.Notify(272512)
			return true
		} else {
			__antithesis_instrumentation__.Notify(272513)
		}
	}
	__antithesis_instrumentation__.Notify(272506)
	return false
}

func RevalidateUniqueConstraintsInTable(
	ctx context.Context, txn *kv.Txn, ie sqlutil.InternalExecutor, tableDesc catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(272514)

	for _, index := range tableDesc.ActiveIndexes() {
		__antithesis_instrumentation__.Notify(272517)
		if index.IsUnique() && func() bool {
			__antithesis_instrumentation__.Notify(272518)
			return index.GetPartitioning().NumImplicitColumns() > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(272519)
			if err := validateUniqueConstraint(
				ctx,
				tableDesc,
				index.GetName(),
				index.IndexDesc().KeyColumnIDs[index.GetPartitioning().NumImplicitColumns():],
				index.GetPredicate(),
				ie,
				txn,
				true,
			); err != nil {
				__antithesis_instrumentation__.Notify(272520)
				log.Errorf(ctx, "validation of unique constraints failed for table %s: %s", tableDesc.GetName(), err)
				return errors.Wrapf(err, "for table %s", tableDesc.GetName())
			} else {
				__antithesis_instrumentation__.Notify(272521)
			}
		} else {
			__antithesis_instrumentation__.Notify(272522)
		}
	}
	__antithesis_instrumentation__.Notify(272515)

	for _, uc := range tableDesc.GetUniqueWithoutIndexConstraints() {
		__antithesis_instrumentation__.Notify(272523)
		if uc.Validity == descpb.ConstraintValidity_Validated {
			__antithesis_instrumentation__.Notify(272524)
			if err := validateUniqueConstraint(
				ctx,
				tableDesc,
				uc.Name,
				uc.ColumnIDs,
				uc.Predicate,
				ie,
				txn,
				true,
			); err != nil {
				__antithesis_instrumentation__.Notify(272525)
				log.Errorf(ctx, "validation of unique constraints failed for table %s: %s", tableDesc.GetName(), err)
				return errors.Wrapf(err, "for table %s", tableDesc.GetName())
			} else {
				__antithesis_instrumentation__.Notify(272526)
			}
		} else {
			__antithesis_instrumentation__.Notify(272527)
		}
	}
	__antithesis_instrumentation__.Notify(272516)

	log.Infof(ctx, "validated all unique constraints in table %s", tableDesc.GetName())
	return nil
}

func validateUniqueConstraint(
	ctx context.Context,
	srcTable catalog.TableDescriptor,
	constraintName string,
	columnIDs []descpb.ColumnID,
	pred string,
	ie sqlutil.InternalExecutor,
	txn *kv.Txn,
	preExisting bool,
) error {
	__antithesis_instrumentation__.Notify(272528)
	query, colNames, err := duplicateRowQuery(
		srcTable, columnIDs, pred, true,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(272532)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272533)
	}
	__antithesis_instrumentation__.Notify(272529)

	log.Infof(ctx, "validating unique constraint %q (%q [%v]) with query %q",
		constraintName,
		srcTable.GetName(),
		colNames,
		query,
	)

	values, err := ie.QueryRowEx(ctx, "validate unique constraint", txn,
		sessiondata.NodeUserSessionDataOverride, query)
	if err != nil {
		__antithesis_instrumentation__.Notify(272534)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272535)
	}
	__antithesis_instrumentation__.Notify(272530)
	if values.Len() > 0 {
		__antithesis_instrumentation__.Notify(272536)
		valuesStr := make([]string, len(values))
		for i := range values {
			__antithesis_instrumentation__.Notify(272539)
			valuesStr[i] = values[i].String()
		}
		__antithesis_instrumentation__.Notify(272537)

		errMsg := "could not create unique constraint"
		if preExisting {
			__antithesis_instrumentation__.Notify(272540)
			errMsg = "failed to validate unique constraint"
		} else {
			__antithesis_instrumentation__.Notify(272541)
		}
		__antithesis_instrumentation__.Notify(272538)
		return errors.WithDetail(
			pgerror.WithConstraintName(
				pgerror.Newf(
					pgcode.UniqueViolation, "%s %q", errMsg, constraintName,
				),
				constraintName,
			),
			fmt.Sprintf(
				"Key (%s)=(%s) is duplicated.", strings.Join(colNames, ","), strings.Join(valuesStr, ","),
			),
		)
	} else {
		__antithesis_instrumentation__.Notify(272542)
	}
	__antithesis_instrumentation__.Notify(272531)
	return nil
}

func (p *planner) ValidateTTLScheduledJobsInCurrentDB(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(272543)
	dbName := p.CurrentDatabase()
	log.Infof(ctx, "validating scheduled jobs in database %s", dbName)
	db, err := p.Descriptors().GetImmutableDatabaseByName(
		ctx, p.Txn(), dbName, tree.DatabaseLookupFlags{Required: true},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(272547)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272548)
	}
	__antithesis_instrumentation__.Notify(272544)
	tableDescs, err := p.Descriptors().GetAllTableDescriptorsInDatabase(ctx, p.Txn(), db.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(272549)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272550)
	}
	__antithesis_instrumentation__.Notify(272545)

	for _, tableDesc := range tableDescs {
		__antithesis_instrumentation__.Notify(272551)
		if err = p.validateTTLScheduledJobInTable(ctx, tableDesc); err != nil {
			__antithesis_instrumentation__.Notify(272552)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272553)
		}
	}
	__antithesis_instrumentation__.Notify(272546)
	return nil
}

var invalidTableTTLScheduledJobError = errors.Newf("invalid scheduled job for table")

func (p *planner) validateTTLScheduledJobInTable(
	ctx context.Context, tableDesc catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(272554)
	if !tableDesc.HasRowLevelTTL() {
		__antithesis_instrumentation__.Notify(272560)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(272561)
	}
	__antithesis_instrumentation__.Notify(272555)
	ttl := tableDesc.GetRowLevelTTL()

	execCfg := p.ExecCfg()
	env := JobSchedulerEnv(execCfg)

	wrapError := func(origErr error) error {
		__antithesis_instrumentation__.Notify(272562)
		return errors.WithHintf(
			errors.Mark(origErr, invalidTableTTLScheduledJobError),
			`use crdb_internal.repair_ttl_table_scheduled_job(%d) to repair the missing job`,
			tableDesc.GetID(),
		)
	}
	__antithesis_instrumentation__.Notify(272556)

	sj, err := jobs.LoadScheduledJob(
		ctx,
		env,
		ttl.ScheduleID,
		execCfg.InternalExecutor,
		p.txn,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(272563)
		if jobs.HasScheduledJobNotFoundError(err) {
			__antithesis_instrumentation__.Notify(272565)
			return wrapError(
				pgerror.Newf(
					pgcode.Internal,
					"table id %d maps to a non-existent schedule id %d",
					tableDesc.GetID(),
					ttl.ScheduleID,
				),
			)
		} else {
			__antithesis_instrumentation__.Notify(272566)
		}
		__antithesis_instrumentation__.Notify(272564)
		return errors.Wrapf(err, "error fetching schedule id %d for table id %d", ttl.ScheduleID, tableDesc.GetID())
	} else {
		__antithesis_instrumentation__.Notify(272567)
	}
	__antithesis_instrumentation__.Notify(272557)

	var args catpb.ScheduledRowLevelTTLArgs
	if err := pbtypes.UnmarshalAny(sj.ExecutionArgs().Args, &args); err != nil {
		__antithesis_instrumentation__.Notify(272568)
		return wrapError(
			pgerror.Wrapf(
				err,
				pgcode.Internal,
				"error unmarshalling scheduled jobs args for table id %d, schedule id %d",
				tableDesc.GetID(),
				ttl.ScheduleID,
			),
		)
	} else {
		__antithesis_instrumentation__.Notify(272569)
	}
	__antithesis_instrumentation__.Notify(272558)

	if args.TableID != tableDesc.GetID() {
		__antithesis_instrumentation__.Notify(272570)
		return wrapError(
			pgerror.Newf(
				pgcode.Internal,
				"schedule id %d points to table id %d instead of table id %d",
				ttl.ScheduleID,
				args.TableID,
				tableDesc.GetID(),
			),
		)
	} else {
		__antithesis_instrumentation__.Notify(272571)
	}
	__antithesis_instrumentation__.Notify(272559)

	return nil
}

func (p *planner) RepairTTLScheduledJobForTable(ctx context.Context, tableID int64) error {
	__antithesis_instrumentation__.Notify(272572)
	tableDesc, err := p.Descriptors().GetMutableTableByID(ctx, p.txn, descpb.ID(tableID), tree.ObjectLookupFlagsWithRequired())
	if err != nil {
		__antithesis_instrumentation__.Notify(272577)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272578)
	}
	__antithesis_instrumentation__.Notify(272573)
	validateErr := p.validateTTLScheduledJobInTable(ctx, tableDesc)
	if validateErr == nil {
		__antithesis_instrumentation__.Notify(272579)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(272580)
	}
	__antithesis_instrumentation__.Notify(272574)
	if !errors.HasType(validateErr, invalidTableTTLScheduledJobError) {
		__antithesis_instrumentation__.Notify(272581)
		return errors.Wrap(validateErr, "error validating TTL on table")
	} else {
		__antithesis_instrumentation__.Notify(272582)
	}
	__antithesis_instrumentation__.Notify(272575)
	sj, err := CreateRowLevelTTLScheduledJob(
		ctx,
		p.ExecCfg(),
		p.txn,
		p.User(),
		tableDesc.GetID(),
		tableDesc.GetRowLevelTTL(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(272583)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272584)
	}
	__antithesis_instrumentation__.Notify(272576)
	tableDesc.RowLevelTTL.ScheduleID = sj.ScheduleID()
	return p.Descriptors().WriteDesc(
		ctx, false, tableDesc, p.txn,
	)
}

func formatValues(colNames []string, values tree.Datums) string {
	__antithesis_instrumentation__.Notify(272585)
	var pairs bytes.Buffer
	for i := range values {
		__antithesis_instrumentation__.Notify(272587)
		if i > 0 {
			__antithesis_instrumentation__.Notify(272589)
			pairs.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(272590)
		}
		__antithesis_instrumentation__.Notify(272588)
		pairs.WriteString(fmt.Sprintf("%s=%v", colNames[i], values[i]))
	}
	__antithesis_instrumentation__.Notify(272586)
	return pairs.String()
}

type checkSet = util.FastIntSet

func checkMutationInput(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
	tabDesc catalog.TableDescriptor,
	checkOrds checkSet,
	checkVals tree.Datums,
) error {
	__antithesis_instrumentation__.Notify(272591)
	if len(checkVals) < checkOrds.Len() {
		__antithesis_instrumentation__.Notify(272594)
		return errors.AssertionFailedf(
			"mismatched check constraint columns: expected %d, got %d", checkOrds.Len(), len(checkVals))
	} else {
		__antithesis_instrumentation__.Notify(272595)
	}
	__antithesis_instrumentation__.Notify(272592)

	checks := tabDesc.ActiveChecks()
	colIdx := 0
	for i := range checks {
		__antithesis_instrumentation__.Notify(272596)
		if !checkOrds.Contains(i) {
			__antithesis_instrumentation__.Notify(272599)
			continue
		} else {
			__antithesis_instrumentation__.Notify(272600)
		}
		__antithesis_instrumentation__.Notify(272597)

		if res, err := tree.GetBool(checkVals[colIdx]); err != nil {
			__antithesis_instrumentation__.Notify(272601)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272602)
			if !res && func() bool {
				__antithesis_instrumentation__.Notify(272603)
				return checkVals[colIdx] != tree.DNull == true
			}() == true {
				__antithesis_instrumentation__.Notify(272604)

				expr, err := schemaexpr.FormatExprForDisplay(
					ctx, tabDesc, checks[i].Expr, semaCtx, sessionData, tree.FmtParsable,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(272606)

					return pgerror.WithConstraintName(errors.Wrapf(err, "failed to satisfy CHECK constraint (%s)", checks[i].Expr), checks[i].Name)
				} else {
					__antithesis_instrumentation__.Notify(272607)
				}
				__antithesis_instrumentation__.Notify(272605)
				return pgerror.WithConstraintName(pgerror.Newf(
					pgcode.CheckViolation, "failed to satisfy CHECK constraint (%s)", expr,
				), checks[i].Name)
			} else {
				__antithesis_instrumentation__.Notify(272608)
			}
		}
		__antithesis_instrumentation__.Notify(272598)
		colIdx++
	}
	__antithesis_instrumentation__.Notify(272593)
	return nil
}
