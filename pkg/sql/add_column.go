package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

func (p *planner) addColumnImpl(
	params runParams,
	n *alterTableNode,
	tn *tree.TableName,
	desc *tabledesc.Mutable,
	t *tree.AlterTableAddColumn,
) error {
	__antithesis_instrumentation__.Notify(242189)
	d := t.ColumnDef

	if d.IsComputed() {
		__antithesis_instrumentation__.Notify(242206)
		d.Computed.Expr = schemaexpr.MaybeRewriteComputedColumn(d.Computed.Expr, params.SessionData())
	} else {
		__antithesis_instrumentation__.Notify(242207)
	}
	__antithesis_instrumentation__.Notify(242190)

	toType, err := tree.ResolveType(params.ctx, d.Type, params.p.semaCtx.GetTypeResolver())
	if err != nil {
		__antithesis_instrumentation__.Notify(242208)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242209)
	}
	__antithesis_instrumentation__.Notify(242191)
	switch toType.Oid() {
	case oid.T_int2vector, oid.T_oidvector:
		__antithesis_instrumentation__.Notify(242210)
		return pgerror.Newf(
			pgcode.FeatureNotSupported,
			"VECTOR column types are unsupported",
		)
	default:
		__antithesis_instrumentation__.Notify(242211)
	}
	__antithesis_instrumentation__.Notify(242192)

	var colOwnedSeqDesc *tabledesc.Mutable
	newDef, seqPrefix, seqName, seqOpts, err := params.p.processSerialLikeInColumnDef(params.ctx, d, tn)
	if err != nil {
		__antithesis_instrumentation__.Notify(242212)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242213)
	}
	__antithesis_instrumentation__.Notify(242193)
	if seqName != nil {
		__antithesis_instrumentation__.Notify(242214)
		colOwnedSeqDesc, err = doCreateSequence(
			params.ctx,
			params.p,
			params.SessionData(),
			seqPrefix.Database,
			seqPrefix.Schema,
			seqName,
			n.tableDesc.Persistence(),
			seqOpts,
			tree.AsStringWithFQNames(n.n, params.Ann()),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(242215)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242216)
		}
	} else {
		__antithesis_instrumentation__.Notify(242217)
	}
	__antithesis_instrumentation__.Notify(242194)
	d = newDef

	cdd, err := tabledesc.MakeColumnDefDescs(params.ctx, d, &params.p.semaCtx, params.EvalContext())
	if err != nil {
		__antithesis_instrumentation__.Notify(242218)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242219)
	}
	__antithesis_instrumentation__.Notify(242195)
	col := cdd.ColumnDescriptor
	idx := cdd.PrimaryKeyOrUniqueIndexDescriptor
	incTelemetryForNewColumn(d, col)

	if idx != nil {
		__antithesis_instrumentation__.Notify(242220)
		if n.tableDesc.IsLocalityRegionalByRow() {
			__antithesis_instrumentation__.Notify(242222)
			if err := params.p.checkNoRegionChangeUnderway(
				params.ctx,
				n.tableDesc.GetParentID(),
				"add an UNIQUE COLUMN on a REGIONAL BY ROW table",
			); err != nil {
				__antithesis_instrumentation__.Notify(242223)
				return err
			} else {
				__antithesis_instrumentation__.Notify(242224)
			}
		} else {
			__antithesis_instrumentation__.Notify(242225)
		}
		__antithesis_instrumentation__.Notify(242221)

		*idx, err = p.configureIndexDescForNewIndexPartitioning(
			params.ctx,
			desc,
			*idx,
			nil,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(242226)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242227)
		}
	} else {
		__antithesis_instrumentation__.Notify(242228)
	}
	__antithesis_instrumentation__.Notify(242196)

	if !col.Nullable && func() bool {
		__antithesis_instrumentation__.Notify(242229)
		return (col.DefaultExpr == nil && func() bool {
			__antithesis_instrumentation__.Notify(242230)
			return !col.IsComputed() == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(242231)
		span := n.tableDesc.PrimaryIndexSpan(params.ExecCfg().Codec)
		kvs, err := params.p.txn.Scan(params.ctx, span.Key, span.EndKey, 1)
		if err != nil {
			__antithesis_instrumentation__.Notify(242233)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242234)
		}
		__antithesis_instrumentation__.Notify(242232)
		if len(kvs) > 0 {
			__antithesis_instrumentation__.Notify(242235)
			return sqlerrors.NewNonNullViolationError(col.Name)
		} else {
			__antithesis_instrumentation__.Notify(242236)
		}
	} else {
		__antithesis_instrumentation__.Notify(242237)
	}
	__antithesis_instrumentation__.Notify(242197)
	if isPublic, err := checkColumnDoesNotExist(n.tableDesc, d.Name); err != nil {
		__antithesis_instrumentation__.Notify(242238)
		if isPublic && func() bool {
			__antithesis_instrumentation__.Notify(242240)
			return t.IfNotExists == true
		}() == true {
			__antithesis_instrumentation__.Notify(242241)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(242242)
		}
		__antithesis_instrumentation__.Notify(242239)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242243)
	}
	__antithesis_instrumentation__.Notify(242198)

	n.tableDesc.AddColumnMutation(col, descpb.DescriptorMutation_ADD)
	if idx != nil {
		__antithesis_instrumentation__.Notify(242244)
		if err := n.tableDesc.AddIndexMutation(params.ctx, idx, descpb.DescriptorMutation_ADD, params.p.ExecCfg().Settings); err != nil {
			__antithesis_instrumentation__.Notify(242245)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242246)
		}
	} else {
		__antithesis_instrumentation__.Notify(242247)
	}
	__antithesis_instrumentation__.Notify(242199)
	if d.HasColumnFamily() {
		__antithesis_instrumentation__.Notify(242248)
		err := n.tableDesc.AddColumnToFamilyMaybeCreate(
			col.Name, string(d.Family.Name), d.Family.Create,
			d.Family.IfNotExists)
		if err != nil {
			__antithesis_instrumentation__.Notify(242249)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242250)
		}
	} else {
		__antithesis_instrumentation__.Notify(242251)
	}
	__antithesis_instrumentation__.Notify(242200)

	if d.IsComputed() {
		__antithesis_instrumentation__.Notify(242252)
		serializedExpr, _, err := schemaexpr.ValidateComputedColumnExpression(
			params.ctx, n.tableDesc, d, tn, "computed column", params.p.SemaCtx(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(242254)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242255)
		}
		__antithesis_instrumentation__.Notify(242253)
		col.ComputeExpr = &serializedExpr
	} else {
		__antithesis_instrumentation__.Notify(242256)
	}
	__antithesis_instrumentation__.Notify(242201)

	if !col.Virtual {
		__antithesis_instrumentation__.Notify(242257)

		primaryIndex := n.tableDesc.GetPrimaryIndex().IndexDescDeepCopy()
		primaryIndex.StoreColumnNames = append(primaryIndex.StoreColumnNames, col.Name)
		primaryIndex.StoreColumnIDs = append(primaryIndex.StoreColumnIDs, col.ID)
		n.tableDesc.SetPrimaryIndex(primaryIndex)
	} else {
		__antithesis_instrumentation__.Notify(242258)
	}
	__antithesis_instrumentation__.Notify(242202)

	version := params.ExecCfg().Settings.Version.ActiveVersion(params.ctx)
	if err := n.tableDesc.AllocateIDs(params.ctx, version); err != nil {
		__antithesis_instrumentation__.Notify(242259)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242260)
	}
	__antithesis_instrumentation__.Notify(242203)

	if err := cdd.ForEachTypedExpr(func(expr tree.TypedExpr) error {
		__antithesis_instrumentation__.Notify(242261)
		changedSeqDescs, err := maybeAddSequenceDependencies(
			params.ctx, params.ExecCfg().Settings, params.p, n.tableDesc, col, expr, nil,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(242264)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242265)
		}
		__antithesis_instrumentation__.Notify(242262)
		for _, changedSeqDesc := range changedSeqDescs {
			__antithesis_instrumentation__.Notify(242266)

			if colOwnedSeqDesc != nil && func() bool {
				__antithesis_instrumentation__.Notify(242268)
				return colOwnedSeqDesc.ID == changedSeqDesc.ID == true
			}() == true {
				__antithesis_instrumentation__.Notify(242269)
				if err := setSequenceOwner(changedSeqDesc, d.Name, desc); err != nil {
					__antithesis_instrumentation__.Notify(242270)
					return err
				} else {
					__antithesis_instrumentation__.Notify(242271)
				}
			} else {
				__antithesis_instrumentation__.Notify(242272)
			}
			__antithesis_instrumentation__.Notify(242267)
			if err := params.p.writeSchemaChange(
				params.ctx, changedSeqDesc, descpb.InvalidMutationID, tree.AsStringWithFQNames(n.n, params.Ann()),
			); err != nil {
				__antithesis_instrumentation__.Notify(242273)
				return err
			} else {
				__antithesis_instrumentation__.Notify(242274)
			}
		}
		__antithesis_instrumentation__.Notify(242263)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(242275)
		return err
	} else {
		__antithesis_instrumentation__.Notify(242276)
	}
	__antithesis_instrumentation__.Notify(242204)

	if n.tableDesc.IsLocalityRegionalByRow() && func() bool {
		__antithesis_instrumentation__.Notify(242277)
		return idx != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(242278)

		if err := p.configureZoneConfigForNewIndexPartitioning(
			params.ctx,
			n.tableDesc,
			*idx,
		); err != nil {
			__antithesis_instrumentation__.Notify(242279)
			return err
		} else {
			__antithesis_instrumentation__.Notify(242280)
		}
	} else {
		__antithesis_instrumentation__.Notify(242281)
	}
	__antithesis_instrumentation__.Notify(242205)

	return nil
}

func checkColumnDoesNotExist(
	tableDesc catalog.TableDescriptor, name tree.Name,
) (isPublic bool, err error) {
	__antithesis_instrumentation__.Notify(242282)
	col, _ := tableDesc.FindColumnWithName(name)
	if col == nil {
		__antithesis_instrumentation__.Notify(242288)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(242289)
	}
	__antithesis_instrumentation__.Notify(242283)
	if col.IsSystemColumn() {
		__antithesis_instrumentation__.Notify(242290)
		return false, pgerror.Newf(pgcode.DuplicateColumn,
			"column name %q conflicts with a system column name",
			col.GetName())
	} else {
		__antithesis_instrumentation__.Notify(242291)
	}
	__antithesis_instrumentation__.Notify(242284)
	if col.Public() {
		__antithesis_instrumentation__.Notify(242292)
		return true, sqlerrors.NewColumnAlreadyExistsError(tree.ErrString(&name), tableDesc.GetName())
	} else {
		__antithesis_instrumentation__.Notify(242293)
	}
	__antithesis_instrumentation__.Notify(242285)
	if col.Adding() {
		__antithesis_instrumentation__.Notify(242294)
		return false, pgerror.Newf(pgcode.DuplicateColumn,
			"duplicate: column %q in the middle of being added, not yet public",
			col.GetName())
	} else {
		__antithesis_instrumentation__.Notify(242295)
	}
	__antithesis_instrumentation__.Notify(242286)
	if col.Dropped() {
		__antithesis_instrumentation__.Notify(242296)
		return false, pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
			"column %q being dropped, try again later", col.GetName())
	} else {
		__antithesis_instrumentation__.Notify(242297)
	}
	__antithesis_instrumentation__.Notify(242287)
	return false, errors.AssertionFailedf("mutation in direction NONE")
}
