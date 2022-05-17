package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachange"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
)

var colInIndexNotSupportedErr = unimplemented.NewWithIssuef(
	47636, "ALTER COLUMN TYPE requiring rewrite of on-disk "+
		"data is currently not supported for columns that are part of an index")

var colOwnsSequenceNotSupportedErr = unimplemented.NewWithIssuef(
	48244, "ALTER COLUMN TYPE for a column that owns a sequence "+
		"is currently not supported")

var colWithConstraintNotSupportedErr = unimplemented.NewWithIssuef(
	48288, "ALTER COLUMN TYPE for a column that has a constraint "+
		"is currently not supported")

var AlterColTypeInTxnNotSupportedErr = unimplemented.NewWithIssuef(
	49351, "ALTER COLUMN TYPE is not supported inside a transaction")

var alterColTypeInCombinationNotSupportedErr = unimplemented.NewWithIssuef(
	49351, "ALTER COLUMN TYPE cannot be used in combination "+
		"with other ALTER TABLE commands")

func AlterColumnType(
	ctx context.Context,
	tableDesc *tabledesc.Mutable,
	col catalog.Column,
	t *tree.AlterTableAlterColumnType,
	params runParams,
	cmds tree.AlterTableCmds,
	tn *tree.TableName,
) error {
	__antithesis_instrumentation__.Notify(242304)
	for _, tableRef := range tableDesc.DependedOnBy {
		__antithesis_instrumentation__.Notify(242312)
		found := false
		for _, colID := range tableRef.ColumnIDs {
			__antithesis_instrumentation__.Notify(242314)
			if colID == col.GetID() {
				__antithesis_instrumentation__.Notify(242315)
				found = true
			} else {
				__antithesis_instrumentation__.Notify(242316)
			}
		}
		__antithesis_instrumentation__.Notify(242313)
		if found {
			__antithesis_instrumentation__.Notify(242317)
			return params.p.dependentViewError(
				ctx, "column", col.GetName(), tableDesc.ParentID, tableRef.ID, "alter type of",
			)
		} else {
			__antithesis_instrumentation__.Notify(242318)
		}
	}
	__antithesis_instrumentation__.Notify(242305)

	typ, err := tree.ResolveType(ctx, t.ToType, params.p.semaCtx.GetTypeResolver())
	if err != nil {
		__antithesis_instrumentation__.Notify(242319)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242320)
	}
	__antithesis_instrumentation__.Notify(242306)

	if t.Collation != "" {
		__antithesis_instrumentation__.Notify(242321)
		if types.IsStringType(typ) {
			__antithesis_instrumentation__.Notify(242322)
			typ = types.MakeCollatedString(typ, t.Collation)
		} else {
			__antithesis_instrumentation__.Notify(242323)
			return pgerror.New(pgcode.Syntax, "COLLATE can only be used with string types")
		}
	} else {
		__antithesis_instrumentation__.Notify(242324)
	}
	__antithesis_instrumentation__.Notify(242307)

	if col.IsGeneratedAsIdentity() {
		__antithesis_instrumentation__.Notify(242325)
		if typ.InternalType.Family != types.IntFamily {
			__antithesis_instrumentation__.Notify(242326)
			return sqlerrors.NewIdentityColumnTypeError()
		} else {
			__antithesis_instrumentation__.Notify(242327)
		}
	} else {
		__antithesis_instrumentation__.Notify(242328)
	}
	__antithesis_instrumentation__.Notify(242308)

	err = colinfo.ValidateColumnDefType(typ)
	if err != nil {
		__antithesis_instrumentation__.Notify(242329)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242330)
	}
	__antithesis_instrumentation__.Notify(242309)

	var kind schemachange.ColumnConversionKind
	if t.Using != nil {
		__antithesis_instrumentation__.Notify(242331)

		kind = schemachange.ColumnConversionGeneral
	} else {
		__antithesis_instrumentation__.Notify(242332)
		kind, err = schemachange.ClassifyConversion(ctx, col.GetType(), typ)
		if err != nil {
			__antithesis_instrumentation__.Notify(242333)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242334)
		}
	}
	__antithesis_instrumentation__.Notify(242310)

	switch kind {
	case schemachange.ColumnConversionDangerous, schemachange.ColumnConversionImpossible:
		__antithesis_instrumentation__.Notify(242335)

		return pgerror.Newf(pgcode.CannotCoerce,
			"the requested type conversion (%s -> %s) requires an explicit USING expression",
			col.GetType().SQLString(), typ.SQLString())
	case schemachange.ColumnConversionTrivial:
		__antithesis_instrumentation__.Notify(242336)
		if col.HasDefault() {
			__antithesis_instrumentation__.Notify(242343)
			if validCast := tree.ValidCast(col.GetType(), typ, tree.CastContextAssignment); !validCast {
				__antithesis_instrumentation__.Notify(242344)
				return pgerror.Wrapf(
					err,
					pgcode.DatatypeMismatch,
					"default for column %q cannot be cast automatically to type %s",
					col.GetName(),
					typ.SQLString(),
				)
			} else {
				__antithesis_instrumentation__.Notify(242345)
			}
		} else {
			__antithesis_instrumentation__.Notify(242346)
		}
		__antithesis_instrumentation__.Notify(242337)
		if col.HasOnUpdate() {
			__antithesis_instrumentation__.Notify(242347)
			if validCast := tree.ValidCast(col.GetType(), typ, tree.CastContextAssignment); !validCast {
				__antithesis_instrumentation__.Notify(242348)
				return pgerror.Wrapf(
					err,
					pgcode.DatatypeMismatch,
					"on update for column %q cannot be cast automatically to type %s",
					col.GetName(),
					typ.SQLString(),
				)
			} else {
				__antithesis_instrumentation__.Notify(242349)
			}
		} else {
			__antithesis_instrumentation__.Notify(242350)
		}
		__antithesis_instrumentation__.Notify(242338)

		col.ColumnDesc().Type = typ
	case schemachange.ColumnConversionGeneral, schemachange.ColumnConversionValidate:
		__antithesis_instrumentation__.Notify(242339)
		if err := alterColumnTypeGeneral(ctx, tableDesc, col, typ, t.Using, params, cmds, tn); err != nil {
			__antithesis_instrumentation__.Notify(242351)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242352)
		}
		__antithesis_instrumentation__.Notify(242340)
		if err := params.p.createOrUpdateSchemaChangeJob(params.ctx, tableDesc, tree.AsStringWithFQNames(t, params.Ann()), tableDesc.ClusterVersion().NextMutationID); err != nil {
			__antithesis_instrumentation__.Notify(242353)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242354)
		}
		__antithesis_instrumentation__.Notify(242341)
		params.p.BufferClientNotice(params.ctx, pgnotice.Newf("ALTER COLUMN TYPE changes are finalized asynchronously; "+
			"further schema changes on this table may be restricted until the job completes; "+
			"some writes to the altered column may be rejected until the schema change is finalized"))
	default:
		__antithesis_instrumentation__.Notify(242342)
		return errors.AssertionFailedf("unknown conversion for %s -> %s",
			col.GetType().SQLString(), typ.SQLString())
	}
	__antithesis_instrumentation__.Notify(242311)

	return nil
}

func alterColumnTypeGeneral(
	ctx context.Context,
	tableDesc *tabledesc.Mutable,
	col catalog.Column,
	toType *types.T,
	using tree.Expr,
	params runParams,
	cmds tree.AlterTableCmds,
	tn *tree.TableName,
) error {
	__antithesis_instrumentation__.Notify(242355)
	if !params.SessionData().AlterColumnTypeGeneralEnabled {
		__antithesis_instrumentation__.Notify(242373)
		return pgerror.WithCandidateCode(
			errors.WithHint(
				errors.WithIssueLink(
					errors.Newf("ALTER COLUMN TYPE from %v to %v is only "+
						"supported experimentally",
						col.GetType(), toType),
					errors.IssueLink{IssueURL: build.MakeIssueURL(49329)}),
				"you can enable alter column type general support by running "+
					"`SET enable_experimental_alter_column_type_general = true`"),
			pgcode.FeatureNotSupported)
	} else {
		__antithesis_instrumentation__.Notify(242374)
	}
	__antithesis_instrumentation__.Notify(242356)

	if col.NumOwnsSequences() != 0 {
		__antithesis_instrumentation__.Notify(242375)
		return colOwnsSequenceNotSupportedErr
	} else {
		__antithesis_instrumentation__.Notify(242376)
	}
	__antithesis_instrumentation__.Notify(242357)

	for i := range tableDesc.Checks {
		__antithesis_instrumentation__.Notify(242377)
		uses, err := tableDesc.CheckConstraintUsesColumn(tableDesc.Checks[i], col.GetID())
		if err != nil {
			__antithesis_instrumentation__.Notify(242379)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242380)
		}
		__antithesis_instrumentation__.Notify(242378)
		if uses {
			__antithesis_instrumentation__.Notify(242381)
			return colWithConstraintNotSupportedErr
		} else {
			__antithesis_instrumentation__.Notify(242382)
		}
	}
	__antithesis_instrumentation__.Notify(242358)

	for _, uc := range tableDesc.AllActiveAndInactiveUniqueWithoutIndexConstraints() {
		__antithesis_instrumentation__.Notify(242383)
		for _, id := range uc.ColumnIDs {
			__antithesis_instrumentation__.Notify(242384)
			if col.GetID() == id {
				__antithesis_instrumentation__.Notify(242385)
				return colWithConstraintNotSupportedErr
			} else {
				__antithesis_instrumentation__.Notify(242386)
			}
		}
	}
	__antithesis_instrumentation__.Notify(242359)

	for _, fk := range tableDesc.AllActiveAndInactiveForeignKeys() {
		__antithesis_instrumentation__.Notify(242387)
		if fk.OriginTableID == tableDesc.GetID() {
			__antithesis_instrumentation__.Notify(242389)
			for _, id := range fk.OriginColumnIDs {
				__antithesis_instrumentation__.Notify(242390)
				if col.GetID() == id {
					__antithesis_instrumentation__.Notify(242391)
					return colWithConstraintNotSupportedErr
				} else {
					__antithesis_instrumentation__.Notify(242392)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(242393)
		}
		__antithesis_instrumentation__.Notify(242388)
		if fk.ReferencedTableID == tableDesc.GetID() {
			__antithesis_instrumentation__.Notify(242394)
			for _, id := range fk.ReferencedColumnIDs {
				__antithesis_instrumentation__.Notify(242395)
				if col.GetID() == id {
					__antithesis_instrumentation__.Notify(242396)
					return colWithConstraintNotSupportedErr
				} else {
					__antithesis_instrumentation__.Notify(242397)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(242398)
		}
	}
	__antithesis_instrumentation__.Notify(242360)

	for _, idx := range tableDesc.NonDropIndexes() {
		__antithesis_instrumentation__.Notify(242399)
		for i := 0; i < idx.NumKeyColumns(); i++ {
			__antithesis_instrumentation__.Notify(242402)
			if idx.GetKeyColumnID(i) == col.GetID() {
				__antithesis_instrumentation__.Notify(242403)
				return colInIndexNotSupportedErr
			} else {
				__antithesis_instrumentation__.Notify(242404)
			}
		}
		__antithesis_instrumentation__.Notify(242400)
		for i := 0; i < idx.NumKeySuffixColumns(); i++ {
			__antithesis_instrumentation__.Notify(242405)
			if idx.GetKeySuffixColumnID(i) == col.GetID() {
				__antithesis_instrumentation__.Notify(242406)
				return colInIndexNotSupportedErr
			} else {
				__antithesis_instrumentation__.Notify(242407)
			}
		}
		__antithesis_instrumentation__.Notify(242401)
		if !idx.Primary() {
			__antithesis_instrumentation__.Notify(242408)
			for i := 0; i < idx.NumSecondaryStoredColumns(); i++ {
				__antithesis_instrumentation__.Notify(242409)
				if idx.GetStoredColumnID(i) == col.GetID() {
					__antithesis_instrumentation__.Notify(242410)
					return colInIndexNotSupportedErr
				} else {
					__antithesis_instrumentation__.Notify(242411)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(242412)
		}
	}
	__antithesis_instrumentation__.Notify(242361)

	if !params.extendedEvalCtx.TxnIsSingleStmt {
		__antithesis_instrumentation__.Notify(242413)
		return AlterColTypeInTxnNotSupportedErr
	} else {
		__antithesis_instrumentation__.Notify(242414)
	}
	__antithesis_instrumentation__.Notify(242362)

	if len(cmds) > 1 {
		__antithesis_instrumentation__.Notify(242415)
		return alterColTypeInCombinationNotSupportedErr
	} else {
		__antithesis_instrumentation__.Notify(242416)
	}
	__antithesis_instrumentation__.Notify(242363)

	currentMutationID := tableDesc.ClusterVersion().NextMutationID
	for i := range tableDesc.Mutations {
		__antithesis_instrumentation__.Notify(242417)
		mut := &tableDesc.Mutations[i]
		if mut.MutationID < currentMutationID {
			__antithesis_instrumentation__.Notify(242418)
			return unimplemented.NewWithIssuef(
				47137, "table %s is currently undergoing a schema change", tableDesc.Name)
		} else {
			__antithesis_instrumentation__.Notify(242419)
		}
	}
	__antithesis_instrumentation__.Notify(242364)

	nameExists := func(name string) bool {
		__antithesis_instrumentation__.Notify(242420)
		_, err := tableDesc.FindColumnWithName(tree.Name(name))
		return err == nil
	}
	__antithesis_instrumentation__.Notify(242365)

	shadowColName := tabledesc.GenerateUniqueName(col.GetName(), nameExists)

	var newColComputeExpr *string

	var inverseExpr string
	if using != nil {
		__antithesis_instrumentation__.Notify(242421)

		expr, _, _, err := schemaexpr.DequalifyAndValidateExpr(
			ctx,
			tableDesc,
			using,
			toType,
			"ALTER COLUMN TYPE USING EXPRESSION",
			&params.p.semaCtx,
			tree.VolatilityVolatile,
			tn,
		)

		if err != nil {
			__antithesis_instrumentation__.Notify(242423)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242424)
		}
		__antithesis_instrumentation__.Notify(242422)
		newColComputeExpr = &expr

		insertedValToString := tree.CastExpr{
			Expr:       &tree.ColumnItem{ColumnName: tree.Name(col.GetName())},
			Type:       types.String,
			SyntaxMode: tree.CastShort,
		}
		insertedVal := tree.Serialize(&insertedValToString)

		errMsg := fmt.Sprintf(
			"'column %s is undergoing the ALTER COLUMN TYPE USING EXPRESSION "+
				"schema change, inserts are not supported until the schema change is "+
				"finalized, '",
			col.GetName())
		failedInsertMsg := fmt.Sprintf(
			"'tried to insert ', %s, ' into %s'", insertedVal, col.GetName(),
		)
		inverseExpr = fmt.Sprintf(
			"crdb_internal.force_error('%s', concat(%s, %s))",
			pgcode.SQLStatementNotYetComplete, errMsg, failedInsertMsg)
	} else {
		__antithesis_instrumentation__.Notify(242425)

		newComputedExpr := tree.CastExpr{
			Expr:       &tree.ColumnItem{ColumnName: tree.Name(col.GetName())},
			Type:       toType,
			SyntaxMode: tree.CastShort,
		}
		s := tree.Serialize(&newComputedExpr)
		newColComputeExpr = &s

		oldColComputeExpr := tree.CastExpr{
			Expr:       &tree.ColumnItem{ColumnName: tree.Name(col.GetName())},
			Type:       col.GetType(),
			SyntaxMode: tree.CastShort,
		}
		inverseExpr = tree.Serialize(&oldColComputeExpr)
	}
	__antithesis_instrumentation__.Notify(242366)

	hasDefault := col.HasDefault()
	hasUpdate := col.HasOnUpdate()
	if hasDefault {
		__antithesis_instrumentation__.Notify(242426)
		if validCast := tree.ValidCast(col.GetType(), toType, tree.CastContextAssignment); !validCast {
			__antithesis_instrumentation__.Notify(242427)
			return pgerror.Newf(
				pgcode.DatatypeMismatch,
				"default for column %q cannot be cast automatically to type %s",
				col.GetName(),
				toType.SQLString(),
			)
		} else {
			__antithesis_instrumentation__.Notify(242428)
		}
	} else {
		__antithesis_instrumentation__.Notify(242429)
	}
	__antithesis_instrumentation__.Notify(242367)
	if hasUpdate {
		__antithesis_instrumentation__.Notify(242430)
		if validCast := tree.ValidCast(col.GetType(), toType, tree.CastContextAssignment); !validCast {
			__antithesis_instrumentation__.Notify(242431)
			return pgerror.Newf(
				pgcode.DatatypeMismatch,
				"on update for column %q cannot be cast automatically to type %s",
				col.GetName(),
				toType.SQLString(),
			)
		} else {
			__antithesis_instrumentation__.Notify(242432)
		}
	} else {
		__antithesis_instrumentation__.Notify(242433)
	}
	__antithesis_instrumentation__.Notify(242368)

	newCol := descpb.ColumnDescriptor{
		Name:            shadowColName,
		Type:            toType,
		Nullable:        col.IsNullable(),
		DefaultExpr:     col.ColumnDesc().DefaultExpr,
		UsesSequenceIds: col.ColumnDesc().UsesSequenceIds,
		OwnsSequenceIds: col.ColumnDesc().OwnsSequenceIds,
		ComputeExpr:     newColComputeExpr,
	}

	family, err := tableDesc.GetFamilyOfColumn(col.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(242434)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242435)
	}
	__antithesis_instrumentation__.Notify(242369)

	if err := tableDesc.AddColumnToFamilyMaybeCreate(
		newCol.Name, family.Name, false, false); err != nil {
		__antithesis_instrumentation__.Notify(242436)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242437)
	}
	__antithesis_instrumentation__.Notify(242370)

	tableDesc.AddColumnMutation(&newCol, descpb.DescriptorMutation_ADD)
	if !newCol.Virtual {
		__antithesis_instrumentation__.Notify(242438)

		primaryIndex := tableDesc.GetPrimaryIndex().IndexDescDeepCopy()
		primaryIndex.StoreColumnNames = append(primaryIndex.StoreColumnNames, newCol.Name)
		primaryIndex.StoreColumnIDs = append(primaryIndex.StoreColumnIDs, newCol.ID)
		tableDesc.SetPrimaryIndex(primaryIndex)
	} else {
		__antithesis_instrumentation__.Notify(242439)
	}
	__antithesis_instrumentation__.Notify(242371)

	version := params.ExecCfg().Settings.Version.ActiveVersion(ctx)
	if err := tableDesc.AllocateIDs(ctx, version); err != nil {
		__antithesis_instrumentation__.Notify(242440)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242441)
	}
	__antithesis_instrumentation__.Notify(242372)

	swapArgs := &descpb.ComputedColumnSwap{
		OldColumnId: col.GetID(),
		NewColumnId: newCol.ID,
		InverseExpr: inverseExpr,
	}

	tableDesc.AddComputedColumnSwapMutation(swapArgs)
	return nil
}
