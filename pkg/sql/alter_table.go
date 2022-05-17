package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/paramparse"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type alterTableNode struct {
	n         *tree.AlterTable
	prefix    catalog.ResolvedObjectPrefix
	tableDesc *tabledesc.Mutable

	statsData map[int]tree.TypedExpr
}

func (p *planner) AlterTable(ctx context.Context, n *tree.AlterTable) (planNode, error) {
	__antithesis_instrumentation__.Notify(243918)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER TABLE",
	); err != nil {
		__antithesis_instrumentation__.Notify(243925)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243926)
	}
	__antithesis_instrumentation__.Notify(243919)

	prefix, tableDesc, err := p.ResolveMutableTableDescriptorEx(
		ctx, n.Table, !n.IfExists, tree.ResolveRequireTableDesc,
	)
	if errors.Is(err, resolver.ErrNoPrimaryKey) {
		__antithesis_instrumentation__.Notify(243927)
		if len(n.Cmds) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(243928)
			return isAlterCmdValidWithoutPrimaryKey(n.Cmds[0]) == true
		}() == true {
			__antithesis_instrumentation__.Notify(243929)
			prefix, tableDesc, err = p.ResolveMutableTableDescriptorExAllowNoPrimaryKey(
				ctx, n.Table, !n.IfExists, tree.ResolveRequireTableDesc,
			)
		} else {
			__antithesis_instrumentation__.Notify(243930)
		}
	} else {
		__antithesis_instrumentation__.Notify(243931)
	}
	__antithesis_instrumentation__.Notify(243920)
	if err != nil {
		__antithesis_instrumentation__.Notify(243932)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(243933)
	}
	__antithesis_instrumentation__.Notify(243921)

	if tableDesc == nil {
		__antithesis_instrumentation__.Notify(243934)
		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(243935)
	}
	__antithesis_instrumentation__.Notify(243922)

	if err := p.CheckPrivilege(ctx, tableDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(243936)
		return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
			"must be owner of table %s or have CREATE privilege on table %s",
			tree.Name(tableDesc.GetName()), tree.Name(tableDesc.GetName()))
	} else {
		__antithesis_instrumentation__.Notify(243937)
	}
	__antithesis_instrumentation__.Notify(243923)

	n.HoistAddColumnConstraints()

	statsData := make(map[int]tree.TypedExpr)
	for i, cmd := range n.Cmds {
		__antithesis_instrumentation__.Notify(243938)
		injectStats, ok := cmd.(*tree.AlterTableInjectStats)
		if !ok {
			__antithesis_instrumentation__.Notify(243941)
			continue
		} else {
			__antithesis_instrumentation__.Notify(243942)
		}
		__antithesis_instrumentation__.Notify(243939)
		typedExpr, err := p.analyzeExpr(
			ctx, injectStats.Stats,
			nil,
			tree.IndexedVarHelper{},
			types.Jsonb, true,
			"INJECT STATISTICS")
		if err != nil {
			__antithesis_instrumentation__.Notify(243943)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(243944)
		}
		__antithesis_instrumentation__.Notify(243940)
		statsData[i] = typedExpr
	}
	__antithesis_instrumentation__.Notify(243924)

	return &alterTableNode{
		n:         n,
		prefix:    prefix,
		tableDesc: tableDesc,
		statsData: statsData,
	}, nil
}

func isAlterCmdValidWithoutPrimaryKey(cmd tree.AlterTableCmd) bool {
	__antithesis_instrumentation__.Notify(243945)
	switch t := cmd.(type) {
	case *tree.AlterTableAlterPrimaryKey:
		__antithesis_instrumentation__.Notify(243947)
		return true
	case *tree.AlterTableAddConstraint:
		__antithesis_instrumentation__.Notify(243948)
		cs, ok := t.ConstraintDef.(*tree.UniqueConstraintTableDef)
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(243950)
			return cs.PrimaryKey == true
		}() == true {
			__antithesis_instrumentation__.Notify(243951)
			return true
		} else {
			__antithesis_instrumentation__.Notify(243952)
		}
	default:
		__antithesis_instrumentation__.Notify(243949)
		return false
	}
	__antithesis_instrumentation__.Notify(243946)
	return false
}

func (n *alterTableNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(243953) }

func (n *alterTableNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(243954)
	telemetry.Inc(sqltelemetry.SchemaChangeAlterCounter("table"))

	descriptorChanged := false
	origNumMutations := len(n.tableDesc.Mutations)
	var droppedViews []string
	resolved := params.p.ResolvedName(n.n.Table)
	tn, ok := resolved.(*tree.TableName)
	if !ok {
		__antithesis_instrumentation__.Notify(243961)
		return errors.AssertionFailedf(
			"%q was not resolved as a table but is %T", resolved, resolved)
	} else {
		__antithesis_instrumentation__.Notify(243962)
	}
	__antithesis_instrumentation__.Notify(243955)

	for i, cmd := range n.n.Cmds {
		__antithesis_instrumentation__.Notify(243963)
		telemetry.Inc(cmd.TelemetryCounter())

		if !n.tableDesc.HasPrimaryKey() && func() bool {
			__antithesis_instrumentation__.Notify(243966)
			return !isAlterCmdValidWithoutPrimaryKey(cmd) == true
		}() == true {
			__antithesis_instrumentation__.Notify(243967)
			return errors.Newf("table %q does not have a primary key, cannot perform%s", n.tableDesc.Name, tree.AsString(cmd))
		} else {
			__antithesis_instrumentation__.Notify(243968)
		}
		__antithesis_instrumentation__.Notify(243964)

		switch t := cmd.(type) {
		case *tree.AlterTableAddColumn:
			__antithesis_instrumentation__.Notify(243969)
			if t.ColumnDef.Unique.WithoutIndex {
				__antithesis_instrumentation__.Notify(244024)

				return errors.WithHint(
					pgerror.New(
						pgcode.FeatureNotSupported,
						"adding a column marked as UNIQUE WITHOUT INDEX is unsupported",
					),
					"add the column first, then run ALTER TABLE ... ADD CONSTRAINT to add a "+
						"UNIQUE WITHOUT INDEX constraint on the column",
				)
			} else {
				__antithesis_instrumentation__.Notify(244025)
			}
			__antithesis_instrumentation__.Notify(243970)
			var err error
			params.p.runWithOptions(resolveFlags{contextDatabaseID: n.tableDesc.ParentID}, func() {
				__antithesis_instrumentation__.Notify(244026)
				err = params.p.addColumnImpl(params, n, tn, n.tableDesc, t)
			})
			__antithesis_instrumentation__.Notify(243971)
			if err != nil {
				__antithesis_instrumentation__.Notify(244027)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244028)
			}
		case *tree.AlterTableAddConstraint:
			__antithesis_instrumentation__.Notify(243972)
			if skip, err := validateConstraintNameIsNotUsed(n.tableDesc, t); err != nil {
				__antithesis_instrumentation__.Notify(244029)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244030)
				if skip {
					__antithesis_instrumentation__.Notify(244031)
					continue
				} else {
					__antithesis_instrumentation__.Notify(244032)
				}
			}
			__antithesis_instrumentation__.Notify(243973)
			switch d := t.ConstraintDef.(type) {
			case *tree.UniqueConstraintTableDef:
				__antithesis_instrumentation__.Notify(244033)
				if d.WithoutIndex {
					__antithesis_instrumentation__.Notify(244054)
					if err := addUniqueWithoutIndexTableDef(
						params.ctx,
						params.EvalContext(),
						params.SessionData(),
						d,
						n.tableDesc,
						*tn,
						NonEmptyTable,
						t.ValidationBehavior,
						params.p.SemaCtx(),
					); err != nil {
						__antithesis_instrumentation__.Notify(244056)
						return err
					} else {
						__antithesis_instrumentation__.Notify(244057)
					}
					__antithesis_instrumentation__.Notify(244055)
					continue
				} else {
					__antithesis_instrumentation__.Notify(244058)
				}
				__antithesis_instrumentation__.Notify(244034)

				if d.PrimaryKey {
					__antithesis_instrumentation__.Notify(244059)

					alterPK := &tree.AlterTableAlterPrimaryKey{
						Columns:       d.Columns,
						Sharded:       d.Sharded,
						Name:          d.Name,
						StorageParams: d.StorageParams,
					}
					if err := params.p.AlterPrimaryKey(
						params.ctx,
						n.tableDesc,
						*alterPK,
						nil,
					); err != nil {
						__antithesis_instrumentation__.Notify(244061)
						return err
					} else {
						__antithesis_instrumentation__.Notify(244062)
					}
					__antithesis_instrumentation__.Notify(244060)
					continue
				} else {
					__antithesis_instrumentation__.Notify(244063)
				}
				__antithesis_instrumentation__.Notify(244035)

				if err := validateColumnsAreAccessible(n.tableDesc, d.Columns); err != nil {
					__antithesis_instrumentation__.Notify(244064)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244065)
				}
				__antithesis_instrumentation__.Notify(244036)

				tableName, err := params.p.getQualifiedTableName(params.ctx, n.tableDesc)
				if err != nil {
					__antithesis_instrumentation__.Notify(244066)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244067)
				}
				__antithesis_instrumentation__.Notify(244037)

				if err := replaceExpressionElemsWithVirtualCols(
					params.ctx,
					n.tableDesc,
					tableName,
					d.Columns,
					false,
					false,
					params.p.SemaCtx(),
				); err != nil {
					__antithesis_instrumentation__.Notify(244068)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244069)
				}
				__antithesis_instrumentation__.Notify(244038)

				for _, column := range d.Columns {
					__antithesis_instrumentation__.Notify(244070)
					if column.Expr != nil {
						__antithesis_instrumentation__.Notify(244072)
						return pgerror.New(
							pgcode.InvalidTableDefinition,
							"cannot create a unique constraint on an expression, use UNIQUE INDEX instead",
						)
					} else {
						__antithesis_instrumentation__.Notify(244073)
					}
					__antithesis_instrumentation__.Notify(244071)
					_, err := n.tableDesc.FindColumnWithName(column.Column)
					if err != nil {
						__antithesis_instrumentation__.Notify(244074)
						return err
					} else {
						__antithesis_instrumentation__.Notify(244075)
					}
				}
				__antithesis_instrumentation__.Notify(244039)
				idx := descpb.IndexDescriptor{
					Name:             string(d.Name),
					Unique:           true,
					StoreColumnNames: d.Storing.ToStrings(),
					CreatedAtNanos:   params.EvalContext().GetTxnTimestamp(time.Microsecond).UnixNano(),
				}
				if err := idx.FillColumns(d.Columns); err != nil {
					__antithesis_instrumentation__.Notify(244076)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244077)
				}
				__antithesis_instrumentation__.Notify(244040)

				if d.Predicate != nil {
					__antithesis_instrumentation__.Notify(244078)
					expr, err := schemaexpr.ValidatePartialIndexPredicate(
						params.ctx, n.tableDesc, d.Predicate, tableName, params.p.SemaCtx(),
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(244080)
						return err
					} else {
						__antithesis_instrumentation__.Notify(244081)
					}
					__antithesis_instrumentation__.Notify(244079)
					idx.Predicate = expr
				} else {
					__antithesis_instrumentation__.Notify(244082)
				}
				__antithesis_instrumentation__.Notify(244041)

				idx, err = params.p.configureIndexDescForNewIndexPartitioning(
					params.ctx,
					n.tableDesc,
					idx,
					d.PartitionByIndex,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(244083)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244084)
				}
				__antithesis_instrumentation__.Notify(244042)
				foundIndex, err := n.tableDesc.FindIndexWithName(string(d.Name))
				if err == nil {
					__antithesis_instrumentation__.Notify(244085)
					if foundIndex.Dropped() {
						__antithesis_instrumentation__.Notify(244086)
						return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
							"index %q being dropped, try again later", d.Name)
					} else {
						__antithesis_instrumentation__.Notify(244087)
					}
				} else {
					__antithesis_instrumentation__.Notify(244088)
				}
				__antithesis_instrumentation__.Notify(244043)
				if err := n.tableDesc.AddIndexMutation(params.ctx, &idx, descpb.DescriptorMutation_ADD, params.p.ExecCfg().Settings); err != nil {
					__antithesis_instrumentation__.Notify(244089)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244090)
				}
				__antithesis_instrumentation__.Notify(244044)

				version := params.ExecCfg().Settings.Version.ActiveVersion(params.ctx)
				if err := n.tableDesc.AllocateIDs(params.ctx, version); err != nil {
					__antithesis_instrumentation__.Notify(244091)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244092)
				}
				__antithesis_instrumentation__.Notify(244045)
				if err := params.p.configureZoneConfigForNewIndexPartitioning(
					params.ctx,
					n.tableDesc,
					idx,
				); err != nil {
					__antithesis_instrumentation__.Notify(244093)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244094)
				}
				__antithesis_instrumentation__.Notify(244046)

				if n.tableDesc.IsLocalityRegionalByRow() {
					__antithesis_instrumentation__.Notify(244095)
					if err := params.p.checkNoRegionChangeUnderway(
						params.ctx,
						n.tableDesc.GetParentID(),
						"create an UNIQUE CONSTRAINT on a REGIONAL BY ROW table",
					); err != nil {
						__antithesis_instrumentation__.Notify(244096)
						return err
					} else {
						__antithesis_instrumentation__.Notify(244097)
					}
				} else {
					__antithesis_instrumentation__.Notify(244098)
				}
			case *tree.CheckConstraintTableDef:
				__antithesis_instrumentation__.Notify(244047)
				var err error
				params.p.runWithOptions(resolveFlags{contextDatabaseID: n.tableDesc.ParentID}, func() {
					__antithesis_instrumentation__.Notify(244099)
					info, infoErr := n.tableDesc.GetConstraintInfo()
					if infoErr != nil {
						__antithesis_instrumentation__.Notify(244104)
						err = infoErr
						return
					} else {
						__antithesis_instrumentation__.Notify(244105)
					}
					__antithesis_instrumentation__.Notify(244100)
					ckBuilder := schemaexpr.MakeCheckConstraintBuilder(params.ctx, *tn, n.tableDesc, &params.p.semaCtx)
					for k := range info {
						__antithesis_instrumentation__.Notify(244106)
						ckBuilder.MarkNameInUse(k)
					}
					__antithesis_instrumentation__.Notify(244101)
					ck, buildErr := ckBuilder.Build(d)
					if buildErr != nil {
						__antithesis_instrumentation__.Notify(244107)
						err = buildErr
						return
					} else {
						__antithesis_instrumentation__.Notify(244108)
					}
					__antithesis_instrumentation__.Notify(244102)
					if t.ValidationBehavior == tree.ValidationDefault {
						__antithesis_instrumentation__.Notify(244109)
						ck.Validity = descpb.ConstraintValidity_Validating
					} else {
						__antithesis_instrumentation__.Notify(244110)
						ck.Validity = descpb.ConstraintValidity_Unvalidated
					}
					__antithesis_instrumentation__.Notify(244103)
					n.tableDesc.AddCheckMutation(ck, descpb.DescriptorMutation_ADD)
				})
				__antithesis_instrumentation__.Notify(244048)
				if err != nil {
					__antithesis_instrumentation__.Notify(244111)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244112)
				}

			case *tree.ForeignKeyConstraintTableDef:
				__antithesis_instrumentation__.Notify(244049)

				if d.Actions.Update != tree.NoAction && func() bool {
					__antithesis_instrumentation__.Notify(244113)
					return d.Actions.Update != tree.Restrict == true
				}() == true {
					__antithesis_instrumentation__.Notify(244114)
					for _, fromCol := range d.FromCols {
						__antithesis_instrumentation__.Notify(244115)
						for _, toCheck := range n.tableDesc.Columns {
							__antithesis_instrumentation__.Notify(244116)
							if fromCol == toCheck.ColName() && func() bool {
								__antithesis_instrumentation__.Notify(244117)
								return toCheck.HasOnUpdate() == true
							}() == true {
								__antithesis_instrumentation__.Notify(244118)
								return pgerror.Newf(
									pgcode.InvalidTableDefinition,
									"cannot specify a foreign key update action and an ON UPDATE"+
										" expression on the same column",
								)
							} else {
								__antithesis_instrumentation__.Notify(244119)
							}
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(244120)
				}
				__antithesis_instrumentation__.Notify(244050)

				affected := make(map[descpb.ID]*tabledesc.Mutable)

				var err error
				params.p.runWithOptions(resolveFlags{skipCache: true}, func() {
					__antithesis_instrumentation__.Notify(244121)

					span := n.tableDesc.PrimaryIndexSpan(params.ExecCfg().Codec)
					kvs, scanErr := params.p.txn.Scan(params.ctx, span.Key, span.EndKey, 1)
					if scanErr != nil {
						__antithesis_instrumentation__.Notify(244124)
						err = scanErr
						return
					} else {
						__antithesis_instrumentation__.Notify(244125)
					}
					__antithesis_instrumentation__.Notify(244122)
					var tableState TableState
					if len(kvs) == 0 {
						__antithesis_instrumentation__.Notify(244126)
						tableState = EmptyTable
					} else {
						__antithesis_instrumentation__.Notify(244127)
						tableState = NonEmptyTable
					}
					__antithesis_instrumentation__.Notify(244123)
					err = ResolveFK(
						params.ctx,
						params.p.txn,
						params.p,
						n.prefix.Database,
						n.prefix.Schema,
						n.tableDesc,
						d,
						affected,
						tableState,
						t.ValidationBehavior,
						params.p.EvalContext(),
					)
				})
				__antithesis_instrumentation__.Notify(244051)
				if err != nil {
					__antithesis_instrumentation__.Notify(244128)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244129)
				}
				__antithesis_instrumentation__.Notify(244052)
				descriptorChanged = true
				for _, updated := range affected {
					__antithesis_instrumentation__.Notify(244130)
					if err := params.p.writeSchemaChange(
						params.ctx, updated, descpb.InvalidMutationID, tree.AsStringWithFQNames(n.n, params.Ann()),
					); err != nil {
						__antithesis_instrumentation__.Notify(244131)
						return err
					} else {
						__antithesis_instrumentation__.Notify(244132)
					}
				}

			default:
				__antithesis_instrumentation__.Notify(244053)
				return errors.AssertionFailedf(
					"unsupported constraint: %T", t.ConstraintDef)
			}

		case *tree.AlterTableAlterPrimaryKey:
			__antithesis_instrumentation__.Notify(243974)
			if err := params.p.AlterPrimaryKey(
				params.ctx,
				n.tableDesc,
				*t,
				nil,
			); err != nil {
				__antithesis_instrumentation__.Notify(244133)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244134)
			}
			__antithesis_instrumentation__.Notify(243975)

			descriptorChanged = true

		case *tree.AlterTableDropColumn:
			__antithesis_instrumentation__.Notify(243976)
			if params.SessionData().SafeUpdates {
				__antithesis_instrumentation__.Notify(244135)
				err := pgerror.DangerousStatementf("ALTER TABLE DROP COLUMN will " +
					"remove all data in that column")
				if !params.extendedEvalCtx.TxnIsSingleStmt {
					__antithesis_instrumentation__.Notify(244137)
					err = errors.WithIssueLink(err, errors.IssueLink{
						IssueURL: "https://github.com/cockroachdb/cockroach/issues/46541",
						Detail: "when used in an explicit transaction combined with other " +
							"schema changes to the same table, DROP COLUMN can result in data " +
							"loss if one of the other schema change fails or is canceled",
					})
				} else {
					__antithesis_instrumentation__.Notify(244138)
				}
				__antithesis_instrumentation__.Notify(244136)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244139)
			}
			__antithesis_instrumentation__.Notify(243977)

			if t.Column == colinfo.TTLDefaultExpirationColumnName && func() bool {
				__antithesis_instrumentation__.Notify(244140)
				return n.tableDesc.HasRowLevelTTL() == true
			}() == true {
				__antithesis_instrumentation__.Notify(244141)
				return errors.WithHintf(
					pgerror.Newf(
						pgcode.InvalidTableDefinition,
						`cannot drop column %s while row-level TTL is active`,
						t.Column,
					),
					"use ALTER TABLE %s RESET (ttl) instead",
					tree.Name(n.tableDesc.GetName()),
				)
			} else {
				__antithesis_instrumentation__.Notify(244142)
			}
			__antithesis_instrumentation__.Notify(243978)

			colDroppedViews, err := dropColumnImpl(params, tn, n.tableDesc, t)
			if err != nil {
				__antithesis_instrumentation__.Notify(244143)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244144)
			}
			__antithesis_instrumentation__.Notify(243979)
			droppedViews = append(droppedViews, colDroppedViews...)
		case *tree.AlterTableDropConstraint:
			__antithesis_instrumentation__.Notify(243980)
			info, err := n.tableDesc.GetConstraintInfo()
			if err != nil {
				__antithesis_instrumentation__.Notify(244145)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244146)
			}
			__antithesis_instrumentation__.Notify(243981)
			name := string(t.Constraint)
			details, ok := info[name]
			if !ok {
				__antithesis_instrumentation__.Notify(244147)
				if t.IfExists {
					__antithesis_instrumentation__.Notify(244149)
					continue
				} else {
					__antithesis_instrumentation__.Notify(244150)
				}
				__antithesis_instrumentation__.Notify(244148)
				return pgerror.Newf(pgcode.UndefinedObject,
					"constraint %q of relation %q does not exist", t.Constraint, n.tableDesc.Name)
			} else {
				__antithesis_instrumentation__.Notify(244151)
			}
			__antithesis_instrumentation__.Notify(243982)
			if err := n.tableDesc.DropConstraint(
				params.ctx,
				name, details,
				func(desc *tabledesc.Mutable, ref *descpb.ForeignKeyConstraint) error {
					__antithesis_instrumentation__.Notify(244152)
					return params.p.removeFKBackReference(params.ctx, desc, ref)
				}, params.ExecCfg().Settings); err != nil {
				__antithesis_instrumentation__.Notify(244153)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244154)
			}
			__antithesis_instrumentation__.Notify(243983)
			descriptorChanged = true
			if err := validateDescriptor(params.ctx, params.p, n.tableDesc); err != nil {
				__antithesis_instrumentation__.Notify(244155)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244156)
			}

		case *tree.AlterTableValidateConstraint:
			__antithesis_instrumentation__.Notify(243984)
			info, err := n.tableDesc.GetConstraintInfo()
			if err != nil {
				__antithesis_instrumentation__.Notify(244157)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244158)
			}
			__antithesis_instrumentation__.Notify(243985)
			name := string(t.Constraint)
			constraint, ok := info[name]
			if !ok {
				__antithesis_instrumentation__.Notify(244159)
				return pgerror.Newf(pgcode.UndefinedObject,
					"constraint %q of relation %q does not exist", t.Constraint, n.tableDesc.Name)
			} else {
				__antithesis_instrumentation__.Notify(244160)
			}
			__antithesis_instrumentation__.Notify(243986)
			if !constraint.Unvalidated {
				__antithesis_instrumentation__.Notify(244161)
				continue
			} else {
				__antithesis_instrumentation__.Notify(244162)
			}
			__antithesis_instrumentation__.Notify(243987)
			switch constraint.Kind {
			case descpb.ConstraintTypeCheck:
				__antithesis_instrumentation__.Notify(244163)
				found := false
				var ck *descpb.TableDescriptor_CheckConstraint
				for _, c := range n.tableDesc.Checks {
					__antithesis_instrumentation__.Notify(244174)

					if c.Name == name && func() bool {
						__antithesis_instrumentation__.Notify(244175)
						return c.Validity != descpb.ConstraintValidity_Validating == true
					}() == true {
						__antithesis_instrumentation__.Notify(244176)
						found = true
						ck = c
						break
					} else {
						__antithesis_instrumentation__.Notify(244177)
					}
				}
				__antithesis_instrumentation__.Notify(244164)
				if !found {
					__antithesis_instrumentation__.Notify(244178)
					return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
						"constraint %q in the middle of being added, try again later", t.Constraint)
				} else {
					__antithesis_instrumentation__.Notify(244179)
				}
				__antithesis_instrumentation__.Notify(244165)
				if err := validateCheckInTxn(
					params.ctx, &params.p.semaCtx, params.ExecCfg().InternalExecutorFactory,
					params.SessionData(), n.tableDesc, params.EvalContext().Txn, ck.Expr,
				); err != nil {
					__antithesis_instrumentation__.Notify(244180)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244181)
				}
				__antithesis_instrumentation__.Notify(244166)
				ck.Validity = descpb.ConstraintValidity_Validated

			case descpb.ConstraintTypeFK:
				__antithesis_instrumentation__.Notify(244167)
				var foundFk *descpb.ForeignKeyConstraint
				for i := range n.tableDesc.OutboundFKs {
					__antithesis_instrumentation__.Notify(244182)
					fk := &n.tableDesc.OutboundFKs[i]

					if fk.Name == name && func() bool {
						__antithesis_instrumentation__.Notify(244183)
						return fk.Validity != descpb.ConstraintValidity_Validating == true
					}() == true {
						__antithesis_instrumentation__.Notify(244184)
						foundFk = fk
						break
					} else {
						__antithesis_instrumentation__.Notify(244185)
					}
				}
				__antithesis_instrumentation__.Notify(244168)
				if foundFk == nil {
					__antithesis_instrumentation__.Notify(244186)
					return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
						"constraint %q in the middle of being added, try again later", t.Constraint)
				} else {
					__antithesis_instrumentation__.Notify(244187)
				}
				__antithesis_instrumentation__.Notify(244169)
				if err := validateFkInTxn(
					params.ctx,
					params.ExecCfg().InternalExecutorFactory,
					params.p.SessionData(),
					n.tableDesc,
					params.EvalContext().Txn,
					params.p.Descriptors(),
					name,
				); err != nil {
					__antithesis_instrumentation__.Notify(244188)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244189)
				}
				__antithesis_instrumentation__.Notify(244170)
				foundFk.Validity = descpb.ConstraintValidity_Validated

			case descpb.ConstraintTypeUnique:
				__antithesis_instrumentation__.Notify(244171)
				if constraint.Index == nil {
					__antithesis_instrumentation__.Notify(244190)
					var foundUnique *descpb.UniqueWithoutIndexConstraint
					for i := range n.tableDesc.UniqueWithoutIndexConstraints {
						__antithesis_instrumentation__.Notify(244194)
						uc := &n.tableDesc.UniqueWithoutIndexConstraints[i]

						if uc.Name == name && func() bool {
							__antithesis_instrumentation__.Notify(244195)
							return uc.Validity != descpb.ConstraintValidity_Validating == true
						}() == true {
							__antithesis_instrumentation__.Notify(244196)
							foundUnique = uc
							break
						} else {
							__antithesis_instrumentation__.Notify(244197)
						}
					}
					__antithesis_instrumentation__.Notify(244191)
					if foundUnique == nil {
						__antithesis_instrumentation__.Notify(244198)
						return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
							"constraint %q in the middle of being added, try again later", t.Constraint)
					} else {
						__antithesis_instrumentation__.Notify(244199)
					}
					__antithesis_instrumentation__.Notify(244192)
					if err := validateUniqueWithoutIndexConstraintInTxn(
						params.ctx, params.ExecCfg().InternalExecutorFactory(
							params.ctx, params.SessionData(),
						), n.tableDesc, params.EvalContext().Txn, name,
					); err != nil {
						__antithesis_instrumentation__.Notify(244200)
						return err
					} else {
						__antithesis_instrumentation__.Notify(244201)
					}
					__antithesis_instrumentation__.Notify(244193)
					foundUnique.Validity = descpb.ConstraintValidity_Validated
					break
				} else {
					__antithesis_instrumentation__.Notify(244202)
				}
				__antithesis_instrumentation__.Notify(244172)

				fallthrough

			default:
				__antithesis_instrumentation__.Notify(244173)
				return pgerror.Newf(pgcode.WrongObjectType,
					"constraint %q of relation %q is not a foreign key, check, or unique without index"+
						" constraint", tree.ErrString(&t.Constraint), tree.ErrString(n.n.Table))
			}
			__antithesis_instrumentation__.Notify(243988)
			descriptorChanged = true

		case tree.ColumnMutationCmd:
			__antithesis_instrumentation__.Notify(243989)

			col, err := n.tableDesc.FindColumnWithName(t.GetColumn())
			if err != nil {
				__antithesis_instrumentation__.Notify(244203)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244204)
			}
			__antithesis_instrumentation__.Notify(243990)
			if col.Dropped() {
				__antithesis_instrumentation__.Notify(244205)
				return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
					"column %q in the middle of being dropped", t.GetColumn())
			} else {
				__antithesis_instrumentation__.Notify(244206)
			}
			__antithesis_instrumentation__.Notify(243991)

			if err := applyColumnMutation(params.ctx, n.tableDesc, col, t, params, n.n.Cmds, tn); err != nil {
				__antithesis_instrumentation__.Notify(244207)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244208)
			}
			__antithesis_instrumentation__.Notify(243992)
			descriptorChanged = true

		case *tree.AlterTablePartitionByTable:
			__antithesis_instrumentation__.Notify(243993)
			if t.All {
				__antithesis_instrumentation__.Notify(244209)
				return unimplemented.NewWithIssue(58736, "PARTITION ALL BY not yet implemented")
			} else {
				__antithesis_instrumentation__.Notify(244210)
			}
			__antithesis_instrumentation__.Notify(243994)
			if n.tableDesc.GetLocalityConfig() != nil {
				__antithesis_instrumentation__.Notify(244211)
				return pgerror.Newf(
					pgcode.FeatureNotSupported,
					"cannot set PARTITION BY on a table in a multi-region enabled database",
				)
			} else {
				__antithesis_instrumentation__.Notify(244212)
			}
			__antithesis_instrumentation__.Notify(243995)
			if n.tableDesc.IsPartitionAllBy() {
				__antithesis_instrumentation__.Notify(244213)
				return unimplemented.NewWithIssue(58736, "changing partition of table with PARTITION ALL BY not yet implemented")
			} else {
				__antithesis_instrumentation__.Notify(244214)
			}
			__antithesis_instrumentation__.Notify(243996)
			if n.tableDesc.GetPrimaryIndex().IsSharded() {
				__antithesis_instrumentation__.Notify(244215)
				return pgerror.New(
					pgcode.FeatureNotSupported,
					"cannot set explicit partitioning with PARTITION BY on hash sharded primary key",
				)
			} else {
				__antithesis_instrumentation__.Notify(244216)
			}
			__antithesis_instrumentation__.Notify(243997)
			oldPartitioning := n.tableDesc.GetPrimaryIndex().GetPartitioning().DeepCopy()
			if oldPartitioning.NumImplicitColumns() > 0 {
				__antithesis_instrumentation__.Notify(244217)
				return unimplemented.NewWithIssue(
					58731,
					"cannot ALTER TABLE PARTITION BY on a table which already has implicit column partitioning",
				)
			} else {
				__antithesis_instrumentation__.Notify(244218)
			}
			__antithesis_instrumentation__.Notify(243998)
			newPrimaryIndexDesc := n.tableDesc.GetPrimaryIndex().IndexDescDeepCopy()
			newImplicitCols, newPartitioning, err := CreatePartitioning(
				params.ctx, params.p.ExecCfg().Settings,
				params.EvalContext(),
				n.tableDesc,
				newPrimaryIndexDesc,
				t.PartitionBy,
				nil,
				params.p.EvalContext().SessionData().ImplicitColumnPartitioningEnabled || func() bool {
					__antithesis_instrumentation__.Notify(244219)
					return n.tableDesc.IsLocalityRegionalByRow() == true
				}() == true,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(244220)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244221)
			}
			__antithesis_instrumentation__.Notify(243999)
			if newPartitioning.NumImplicitColumns > 0 {
				__antithesis_instrumentation__.Notify(244222)
				return unimplemented.NewWithIssue(
					58731,
					"cannot ALTER TABLE and change the partitioning to contain implicit columns",
				)
			} else {
				__antithesis_instrumentation__.Notify(244223)
			}
			__antithesis_instrumentation__.Notify(244000)
			isIndexAltered := tabledesc.UpdateIndexPartitioning(&newPrimaryIndexDesc, true, newImplicitCols, newPartitioning)
			if isIndexAltered {
				__antithesis_instrumentation__.Notify(244224)
				n.tableDesc.SetPrimaryIndex(newPrimaryIndexDesc)
				descriptorChanged = true
				if err := deleteRemovedPartitionZoneConfigs(
					params.ctx,
					params.p.txn,
					n.tableDesc,
					n.tableDesc.GetPrimaryIndexID(),
					oldPartitioning,
					n.tableDesc.GetPrimaryIndex().GetPartitioning(),
					params.extendedEvalCtx.ExecCfg,
				); err != nil {
					__antithesis_instrumentation__.Notify(244225)
					return err
				} else {
					__antithesis_instrumentation__.Notify(244226)
				}
			} else {
				__antithesis_instrumentation__.Notify(244227)
			}

		case *tree.AlterTableSetAudit:
			__antithesis_instrumentation__.Notify(244001)
			changed, err := params.p.setAuditMode(params.ctx, n.tableDesc, t.Mode)
			if err != nil {
				__antithesis_instrumentation__.Notify(244228)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244229)
			}
			__antithesis_instrumentation__.Notify(244002)
			descriptorChanged = descriptorChanged || func() bool {
				__antithesis_instrumentation__.Notify(244230)
				return changed == true
			}() == true

		case *tree.AlterTableInjectStats:
			__antithesis_instrumentation__.Notify(244003)
			sd, ok := n.statsData[i]
			if !ok {
				__antithesis_instrumentation__.Notify(244231)
				return errors.AssertionFailedf("missing stats data")
			} else {
				__antithesis_instrumentation__.Notify(244232)
			}
			__antithesis_instrumentation__.Notify(244004)
			if !params.extendedEvalCtx.TxnIsSingleStmt {
				__antithesis_instrumentation__.Notify(244233)
				return errors.New("cannot inject statistics in an explicit transaction")
			} else {
				__antithesis_instrumentation__.Notify(244234)
			}
			__antithesis_instrumentation__.Notify(244005)
			if err := injectTableStats(params, n.tableDesc, sd); err != nil {
				__antithesis_instrumentation__.Notify(244235)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244236)
			}

		case *tree.AlterTableSetStorageParams:
			__antithesis_instrumentation__.Notify(244006)
			var ttlBefore *catpb.RowLevelTTL
			if ttl := n.tableDesc.GetRowLevelTTL(); ttl != nil {
				__antithesis_instrumentation__.Notify(244237)
				ttlBefore = protoutil.Clone(ttl).(*catpb.RowLevelTTL)
			} else {
				__antithesis_instrumentation__.Notify(244238)
			}
			__antithesis_instrumentation__.Notify(244007)
			if err := paramparse.SetStorageParameters(
				params.ctx,
				params.p.SemaCtx(),
				params.EvalContext(),
				t.StorageParams,
				paramparse.NewTableStorageParamObserver(n.tableDesc),
			); err != nil {
				__antithesis_instrumentation__.Notify(244239)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244240)
			}
			__antithesis_instrumentation__.Notify(244008)
			descriptorChanged = true

			if err := handleTTLStorageParamChange(
				params,
				tn,
				n.tableDesc,
				ttlBefore,
				n.tableDesc.GetRowLevelTTL(),
			); err != nil {
				__antithesis_instrumentation__.Notify(244241)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244242)
			}

		case *tree.AlterTableResetStorageParams:
			__antithesis_instrumentation__.Notify(244009)
			var ttlBefore *catpb.RowLevelTTL
			if ttl := n.tableDesc.GetRowLevelTTL(); ttl != nil {
				__antithesis_instrumentation__.Notify(244243)
				ttlBefore = protoutil.Clone(ttl).(*catpb.RowLevelTTL)
			} else {
				__antithesis_instrumentation__.Notify(244244)
			}
			__antithesis_instrumentation__.Notify(244010)
			if err := paramparse.ResetStorageParameters(
				params.ctx,
				params.EvalContext(),
				t.Params,
				paramparse.NewTableStorageParamObserver(n.tableDesc),
			); err != nil {
				__antithesis_instrumentation__.Notify(244245)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244246)
			}
			__antithesis_instrumentation__.Notify(244011)
			descriptorChanged = true

			if err := handleTTLStorageParamChange(
				params,
				tn,
				n.tableDesc,
				ttlBefore,
				n.tableDesc.GetRowLevelTTL(),
			); err != nil {
				__antithesis_instrumentation__.Notify(244247)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244248)
			}

		case *tree.AlterTableRenameColumn:
			__antithesis_instrumentation__.Notify(244012)
			descChanged, err := params.p.renameColumn(params.ctx, n.tableDesc, t.Column, t.NewName)
			if err != nil {
				__antithesis_instrumentation__.Notify(244249)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244250)
			}
			__antithesis_instrumentation__.Notify(244013)
			descriptorChanged = descriptorChanged || func() bool {
				__antithesis_instrumentation__.Notify(244251)
				return descChanged == true
			}() == true

		case *tree.AlterTableRenameConstraint:
			__antithesis_instrumentation__.Notify(244014)
			info, err := n.tableDesc.GetConstraintInfo()
			if err != nil {
				__antithesis_instrumentation__.Notify(244252)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244253)
			}
			__antithesis_instrumentation__.Notify(244015)
			details, ok := info[string(t.Constraint)]
			if !ok {
				__antithesis_instrumentation__.Notify(244254)
				return pgerror.Newf(pgcode.UndefinedObject,
					"constraint %q of relation %q does not exist", tree.ErrString(&t.Constraint), n.tableDesc.Name)
			} else {
				__antithesis_instrumentation__.Notify(244255)
			}
			__antithesis_instrumentation__.Notify(244016)
			if t.Constraint == t.NewName {
				__antithesis_instrumentation__.Notify(244256)

				break
			} else {
				__antithesis_instrumentation__.Notify(244257)
			}
			__antithesis_instrumentation__.Notify(244017)

			if _, ok := info[string(t.NewName)]; ok {
				__antithesis_instrumentation__.Notify(244258)
				return pgerror.Newf(pgcode.DuplicateObject,
					"duplicate constraint name: %q", tree.ErrString(&t.NewName))
			} else {
				__antithesis_instrumentation__.Notify(244259)
			}
			__antithesis_instrumentation__.Notify(244018)

			switch details.Kind {
			case descpb.ConstraintTypeUnique, descpb.ConstraintTypePK:
				__antithesis_instrumentation__.Notify(244260)
				if catalog.FindNonDropIndex(n.tableDesc, func(idx catalog.Index) bool {
					__antithesis_instrumentation__.Notify(244262)
					return idx.GetName() == string(t.NewName)
				}) != nil {
					__antithesis_instrumentation__.Notify(244263)
					return pgerror.Newf(pgcode.DuplicateRelation,
						"relation %v already exists", t.NewName)
				} else {
					__antithesis_instrumentation__.Notify(244264)
				}
			default:
				__antithesis_instrumentation__.Notify(244261)
			}
			__antithesis_instrumentation__.Notify(244019)

			if err := params.p.CheckPrivilege(params.ctx, n.tableDesc, privilege.CREATE); err != nil {
				__antithesis_instrumentation__.Notify(244265)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244266)
			}
			__antithesis_instrumentation__.Notify(244020)

			depViewRenameError := func(objType string, refTableID descpb.ID) error {
				__antithesis_instrumentation__.Notify(244267)
				return params.p.dependentViewError(params.ctx,
					objType, tree.ErrString(&t.NewName), n.tableDesc.ParentID, refTableID, "rename",
				)
			}
			__antithesis_instrumentation__.Notify(244021)

			if err := n.tableDesc.RenameConstraint(
				details, string(t.Constraint), string(t.NewName), depViewRenameError,
				func(desc *tabledesc.Mutable, ref *descpb.ForeignKeyConstraint, newName string) error {
					__antithesis_instrumentation__.Notify(244268)
					return params.p.updateFKBackReferenceName(params.ctx, desc, ref, newName)
				}); err != nil {
				__antithesis_instrumentation__.Notify(244269)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244270)
			}
			__antithesis_instrumentation__.Notify(244022)
			descriptorChanged = true
		default:
			__antithesis_instrumentation__.Notify(244023)
			return errors.AssertionFailedf("unsupported alter command: %T", cmd)
		}
		__antithesis_instrumentation__.Notify(243965)

		version := params.ExecCfg().Settings.Version.ActiveVersion(params.ctx)
		if err := n.tableDesc.AllocateIDs(params.ctx, version); err != nil {
			__antithesis_instrumentation__.Notify(244271)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244272)
		}
	}
	__antithesis_instrumentation__.Notify(243956)

	addedMutations := len(n.tableDesc.Mutations) > origNumMutations
	if !addedMutations && func() bool {
		__antithesis_instrumentation__.Notify(244273)
		return !descriptorChanged == true
	}() == true {
		__antithesis_instrumentation__.Notify(244274)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(244275)
	}
	__antithesis_instrumentation__.Notify(243957)

	mutationID := descpb.InvalidMutationID
	if addedMutations {
		__antithesis_instrumentation__.Notify(244276)
		mutationID = n.tableDesc.ClusterVersion().NextMutationID
	} else {
		__antithesis_instrumentation__.Notify(244277)
	}
	__antithesis_instrumentation__.Notify(243958)
	if err := params.p.writeSchemaChange(
		params.ctx, n.tableDesc, mutationID, tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(244278)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244279)
	}
	__antithesis_instrumentation__.Notify(243959)

	if err := params.p.addBackRefsFromAllTypesInTable(params.ctx, n.tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(244280)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244281)
	}
	__antithesis_instrumentation__.Notify(243960)

	return params.p.logEvent(params.ctx,
		n.tableDesc.ID,
		&eventpb.AlterTable{
			TableName:           params.p.ResolvedName(n.n.Table).FQString(),
			MutationID:          uint32(mutationID),
			CascadeDroppedViews: droppedViews,
		})
}

func (p *planner) setAuditMode(
	ctx context.Context, desc *tabledesc.Mutable, auditMode tree.AuditMode,
) (bool, error) {
	__antithesis_instrumentation__.Notify(244282)

	p.curPlan.auditEvents = append(p.curPlan.auditEvents,
		auditEvent{desc: desc, writing: true})

	if err := p.RequireAdminRole(ctx, "change auditing settings on a table"); err != nil {
		__antithesis_instrumentation__.Notify(244284)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(244285)
	}
	__antithesis_instrumentation__.Notify(244283)

	telemetry.Inc(sqltelemetry.SchemaSetAuditModeCounter(auditMode.TelemetryName()))

	return desc.SetAuditMode(auditMode)
}

func (n *alterTableNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(244286)
	return false, nil
}
func (n *alterTableNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(244287)
	return tree.Datums{}
}
func (n *alterTableNode) Close(context.Context) { __antithesis_instrumentation__.Notify(244288) }

func applyColumnMutation(
	ctx context.Context,
	tableDesc *tabledesc.Mutable,
	col catalog.Column,
	mut tree.ColumnMutationCmd,
	params runParams,
	cmds tree.AlterTableCmds,
	tn *tree.TableName,
) error {
	__antithesis_instrumentation__.Notify(244289)
	switch t := mut.(type) {
	case *tree.AlterTableAlterColumnType:
		__antithesis_instrumentation__.Notify(244291)
		return AlterColumnType(ctx, tableDesc, col, t, params, cmds, tn)

	case *tree.AlterTableSetDefault:
		__antithesis_instrumentation__.Notify(244292)
		if err := updateNonComputedColExpr(
			params,
			tableDesc,
			col,
			t.Default,
			&col.ColumnDesc().DefaultExpr,
			"DEFAULT",
		); err != nil {
			__antithesis_instrumentation__.Notify(244311)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244312)
		}

	case *tree.AlterTableSetOnUpdate:
		__antithesis_instrumentation__.Notify(244293)

		for _, fk := range tableDesc.OutboundFKs {
			__antithesis_instrumentation__.Notify(244313)
			for _, colID := range fk.OriginColumnIDs {
				__antithesis_instrumentation__.Notify(244314)
				if colID == col.GetID() && func() bool {
					__antithesis_instrumentation__.Notify(244315)
					return fk.OnUpdate != catpb.ForeignKeyAction_NO_ACTION == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(244316)
					return fk.OnUpdate != catpb.ForeignKeyAction_RESTRICT == true
				}() == true {
					__antithesis_instrumentation__.Notify(244317)
					return pgerror.Newf(
						pgcode.InvalidColumnDefinition,
						"column %s(%d) cannot have both an ON UPDATE expression and a foreign"+
							" key ON UPDATE action",
						col.GetName(),
						col.GetID(),
					)
				} else {
					__antithesis_instrumentation__.Notify(244318)
				}
			}
		}
		__antithesis_instrumentation__.Notify(244294)

		if err := updateNonComputedColExpr(
			params,
			tableDesc,
			col,
			t.Expr,
			&col.ColumnDesc().OnUpdateExpr,
			"ON UPDATE",
		); err != nil {
			__antithesis_instrumentation__.Notify(244319)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244320)
		}

	case *tree.AlterTableSetVisible:
		__antithesis_instrumentation__.Notify(244295)
		column, err := tableDesc.FindActiveOrNewColumnByName(col.ColName())
		if err != nil {
			__antithesis_instrumentation__.Notify(244321)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244322)
		}
		__antithesis_instrumentation__.Notify(244296)
		column.ColumnDesc().Hidden = !t.Visible

	case *tree.AlterTableSetNotNull:
		__antithesis_instrumentation__.Notify(244297)
		if !col.IsNullable() {
			__antithesis_instrumentation__.Notify(244323)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(244324)
		}
		__antithesis_instrumentation__.Notify(244298)

		for i := range tableDesc.Mutations {
			__antithesis_instrumentation__.Notify(244325)
			if constraint := tableDesc.Mutations[i].GetConstraint(); constraint != nil && func() bool {
				__antithesis_instrumentation__.Notify(244326)
				return constraint.ConstraintType == descpb.ConstraintToUpdate_NOT_NULL == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(244327)
				return constraint.NotNullColumn == col.GetID() == true
			}() == true {
				__antithesis_instrumentation__.Notify(244328)
				if tableDesc.Mutations[i].Direction == descpb.DescriptorMutation_ADD {
					__antithesis_instrumentation__.Notify(244330)
					return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
						"constraint in the middle of being added")
				} else {
					__antithesis_instrumentation__.Notify(244331)
				}
				__antithesis_instrumentation__.Notify(244329)
				return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
					"constraint in the middle of being dropped, try again later")
			} else {
				__antithesis_instrumentation__.Notify(244332)
			}
		}
		__antithesis_instrumentation__.Notify(244299)

		info, err := tableDesc.GetConstraintInfo()
		if err != nil {
			__antithesis_instrumentation__.Notify(244333)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244334)
		}
		__antithesis_instrumentation__.Notify(244300)
		inuseNames := make(map[string]struct{}, len(info))
		for k := range info {
			__antithesis_instrumentation__.Notify(244335)
			inuseNames[k] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(244301)
		check := tabledesc.MakeNotNullCheckConstraint(col.GetName(), col.GetID(), tableDesc.GetNextConstraintID(), inuseNames, descpb.ConstraintValidity_Validating)
		tableDesc.AddNotNullMutation(check, descpb.DescriptorMutation_ADD)
		tableDesc.NextConstraintID++

	case *tree.AlterTableDropNotNull:
		__antithesis_instrumentation__.Notify(244302)
		if col.IsNullable() {
			__antithesis_instrumentation__.Notify(244336)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(244337)
		}
		__antithesis_instrumentation__.Notify(244303)

		if tableDesc.GetPrimaryIndex().CollectKeyColumnIDs().Contains(col.GetID()) {
			__antithesis_instrumentation__.Notify(244338)
			return pgerror.Newf(pgcode.InvalidTableDefinition,
				`column "%s" is in a primary index`, col.GetName())
		} else {
			__antithesis_instrumentation__.Notify(244339)
		}
		__antithesis_instrumentation__.Notify(244304)

		for i := range tableDesc.Mutations {
			__antithesis_instrumentation__.Notify(244340)
			if constraint := tableDesc.Mutations[i].GetConstraint(); constraint != nil && func() bool {
				__antithesis_instrumentation__.Notify(244341)
				return constraint.ConstraintType == descpb.ConstraintToUpdate_NOT_NULL == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(244342)
				return constraint.NotNullColumn == col.GetID() == true
			}() == true {
				__antithesis_instrumentation__.Notify(244343)
				if tableDesc.Mutations[i].Direction == descpb.DescriptorMutation_ADD {
					__antithesis_instrumentation__.Notify(244345)
					return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
						"constraint in the middle of being added, try again later")
				} else {
					__antithesis_instrumentation__.Notify(244346)
				}
				__antithesis_instrumentation__.Notify(244344)
				return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
					"constraint in the middle of being dropped")
			} else {
				__antithesis_instrumentation__.Notify(244347)
			}
		}
		__antithesis_instrumentation__.Notify(244305)
		info, err := tableDesc.GetConstraintInfo()
		if err != nil {
			__antithesis_instrumentation__.Notify(244348)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244349)
		}
		__antithesis_instrumentation__.Notify(244306)
		inuseNames := make(map[string]struct{}, len(info))
		for k := range info {
			__antithesis_instrumentation__.Notify(244350)
			inuseNames[k] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(244307)
		col.ColumnDesc().Nullable = true

		check := tabledesc.MakeNotNullCheckConstraint(col.GetName(), col.GetID(), tableDesc.GetNextConstraintID(), inuseNames, descpb.ConstraintValidity_Dropping)
		tableDesc.Checks = append(tableDesc.Checks, check)
		tableDesc.NextConstraintID++
		tableDesc.AddNotNullMutation(check, descpb.DescriptorMutation_DROP)

	case *tree.AlterTableDropStored:
		__antithesis_instrumentation__.Notify(244308)
		if !col.IsComputed() {
			__antithesis_instrumentation__.Notify(244351)
			return pgerror.Newf(pgcode.InvalidColumnDefinition,
				"column %q is not a computed column", col.GetName())
		} else {
			__antithesis_instrumentation__.Notify(244352)
		}
		__antithesis_instrumentation__.Notify(244309)
		if col.IsVirtual() {
			__antithesis_instrumentation__.Notify(244353)
			return pgerror.Newf(pgcode.InvalidColumnDefinition,
				"column %q is not a stored computed column", col.GetName())
		} else {
			__antithesis_instrumentation__.Notify(244354)
		}
		__antithesis_instrumentation__.Notify(244310)
		col.ColumnDesc().ComputeExpr = nil
	}
	__antithesis_instrumentation__.Notify(244290)
	return nil
}

func labeledRowValues(cols []catalog.Column, values tree.Datums) string {
	__antithesis_instrumentation__.Notify(244355)
	var s bytes.Buffer
	for i := range cols {
		__antithesis_instrumentation__.Notify(244357)
		if i != 0 {
			__antithesis_instrumentation__.Notify(244359)
			s.WriteString(`, `)
		} else {
			__antithesis_instrumentation__.Notify(244360)
		}
		__antithesis_instrumentation__.Notify(244358)
		s.WriteString(cols[i].GetName())
		s.WriteString(`=`)
		s.WriteString(values[i].String())
	}
	__antithesis_instrumentation__.Notify(244356)
	return s.String()
}

func updateNonComputedColExpr(
	params runParams,
	tab *tabledesc.Mutable,
	col catalog.Column,
	newExpr tree.Expr,
	exprField **string,
	op string,
) error {
	__antithesis_instrumentation__.Notify(244361)

	if col.NumUsesSequences() > 0 {
		__antithesis_instrumentation__.Notify(244366)
		if err := params.p.removeSequenceDependencies(params.ctx, tab, col); err != nil {
			__antithesis_instrumentation__.Notify(244367)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244368)
		}
	} else {
		__antithesis_instrumentation__.Notify(244369)
	}
	__antithesis_instrumentation__.Notify(244362)

	if col.IsGeneratedAsIdentity() {
		__antithesis_instrumentation__.Notify(244370)
		return sqlerrors.NewSyntaxErrorf("column %q is an identity column", col.GetName())
	} else {
		__antithesis_instrumentation__.Notify(244371)
	}
	__antithesis_instrumentation__.Notify(244363)

	if newExpr == nil {
		__antithesis_instrumentation__.Notify(244372)
		*exprField = nil
	} else {
		__antithesis_instrumentation__.Notify(244373)
		_, s, err := sanitizeColumnExpression(params, newExpr, col, op)
		if err != nil {
			__antithesis_instrumentation__.Notify(244375)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244376)
		}
		__antithesis_instrumentation__.Notify(244374)

		*exprField = &s
	}
	__antithesis_instrumentation__.Notify(244364)

	if err := updateSequenceDependencies(params, tab, col); err != nil {
		__antithesis_instrumentation__.Notify(244377)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244378)
	}
	__antithesis_instrumentation__.Notify(244365)

	return nil
}

func sanitizeColumnExpression(
	p runParams, expr tree.Expr, col catalog.Column, opName string,
) (tree.TypedExpr, string, error) {
	__antithesis_instrumentation__.Notify(244379)
	colDatumType := col.GetType()
	typedExpr, err := schemaexpr.SanitizeVarFreeExpr(
		p.ctx, expr, colDatumType, opName, &p.p.semaCtx, tree.VolatilityVolatile,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(244381)
		return nil, "", pgerror.WithCandidateCode(err, pgcode.DatatypeMismatch)
	} else {
		__antithesis_instrumentation__.Notify(244382)
	}
	__antithesis_instrumentation__.Notify(244380)

	s := tree.Serialize(typedExpr)
	return typedExpr, s, nil
}

func updateSequenceDependencies(
	params runParams, tableDesc *tabledesc.Mutable, colDesc catalog.Column,
) error {
	__antithesis_instrumentation__.Notify(244383)
	var seqDescsToUpdate []*tabledesc.Mutable
	mergeNewSeqDescs := func(toAdd []*tabledesc.Mutable) {
		__antithesis_instrumentation__.Notify(244387)
		seqDescsToUpdate = append(seqDescsToUpdate, toAdd...)
		sort.Slice(seqDescsToUpdate,
			func(i, j int) bool {
				__antithesis_instrumentation__.Notify(244390)
				return seqDescsToUpdate[i].GetID() < seqDescsToUpdate[j].GetID()
			})
		__antithesis_instrumentation__.Notify(244388)
		truncated := make([]*tabledesc.Mutable, 0, len(seqDescsToUpdate))
		for i, v := range seqDescsToUpdate {
			__antithesis_instrumentation__.Notify(244391)
			if i == 0 || func() bool {
				__antithesis_instrumentation__.Notify(244392)
				return seqDescsToUpdate[i-1].GetID() != v.GetID() == true
			}() == true {
				__antithesis_instrumentation__.Notify(244393)
				truncated = append(truncated, v)
			} else {
				__antithesis_instrumentation__.Notify(244394)
			}
		}
		__antithesis_instrumentation__.Notify(244389)
		seqDescsToUpdate = truncated
	}
	__antithesis_instrumentation__.Notify(244384)
	for _, colExpr := range []struct {
		name   string
		exists func() bool
		get    func() string
	}{
		{
			name:   "DEFAULT",
			exists: colDesc.HasDefault,
			get:    colDesc.GetDefaultExpr,
		},
		{
			name:   "ON UPDATE",
			exists: colDesc.HasOnUpdate,
			get:    colDesc.GetOnUpdateExpr,
		},
	} {
		__antithesis_instrumentation__.Notify(244395)
		if !colExpr.exists() {
			__antithesis_instrumentation__.Notify(244400)
			continue
		} else {
			__antithesis_instrumentation__.Notify(244401)
		}
		__antithesis_instrumentation__.Notify(244396)
		untypedExpr, err := parser.ParseExpr(colExpr.get())
		if err != nil {
			__antithesis_instrumentation__.Notify(244402)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(244403)
		}
		__antithesis_instrumentation__.Notify(244397)

		typedExpr, _, err := sanitizeColumnExpression(
			params,
			untypedExpr,
			colDesc,
			"DEFAULT",
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(244404)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244405)
		}
		__antithesis_instrumentation__.Notify(244398)

		newSeqDescs, err := maybeAddSequenceDependencies(
			params.ctx,
			params.p.ExecCfg().Settings,
			params.p,
			tableDesc,
			colDesc.ColumnDesc(),
			typedExpr,
			nil,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(244406)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244407)
		}
		__antithesis_instrumentation__.Notify(244399)

		mergeNewSeqDescs(newSeqDescs)
	}
	__antithesis_instrumentation__.Notify(244385)

	for _, changedSeqDesc := range seqDescsToUpdate {
		__antithesis_instrumentation__.Notify(244408)
		if err := params.p.writeSchemaChange(
			params.ctx, changedSeqDesc, descpb.InvalidMutationID,
			fmt.Sprintf("updating dependent sequence %s(%d) for table %s(%d)",
				changedSeqDesc.Name, changedSeqDesc.ID, tableDesc.Name, tableDesc.ID,
			)); err != nil {
			__antithesis_instrumentation__.Notify(244409)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244410)
		}
	}
	__antithesis_instrumentation__.Notify(244386)

	return nil
}

func injectTableStats(
	params runParams, desc catalog.TableDescriptor, statsExpr tree.TypedExpr,
) error {
	__antithesis_instrumentation__.Notify(244411)
	val, err := statsExpr.Eval(params.EvalContext())
	if err != nil {
		__antithesis_instrumentation__.Notify(244417)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244418)
	}
	__antithesis_instrumentation__.Notify(244412)
	if val == tree.DNull {
		__antithesis_instrumentation__.Notify(244419)
		return pgerror.New(pgcode.Syntax,
			"statistics cannot be NULL")
	} else {
		__antithesis_instrumentation__.Notify(244420)
	}
	__antithesis_instrumentation__.Notify(244413)
	jsonStr := val.(*tree.DJSON).JSON.String()
	var jsonStats []stats.JSONStatistic
	if err := gojson.Unmarshal([]byte(jsonStr), &jsonStats); err != nil {
		__antithesis_instrumentation__.Notify(244421)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244422)
	}
	__antithesis_instrumentation__.Notify(244414)

	if _, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.Exec(
		params.ctx,
		"delete-stats",
		params.EvalContext().Txn,
		`DELETE FROM system.table_statistics WHERE "tableID" = $1`, desc.GetID(),
	); err != nil {
		__antithesis_instrumentation__.Notify(244423)
		return errors.Wrapf(err, "failed to delete old stats")
	} else {
		__antithesis_instrumentation__.Notify(244424)
	}
	__antithesis_instrumentation__.Notify(244415)

StatsLoop:
	for i := range jsonStats {
		__antithesis_instrumentation__.Notify(244425)
		s := &jsonStats[i]
		h, err := s.GetHistogram(&params.p.semaCtx, params.EvalContext())
		if err != nil {
			__antithesis_instrumentation__.Notify(244429)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244430)
		}
		__antithesis_instrumentation__.Notify(244426)

		var histogram interface{}
		if h != nil {
			__antithesis_instrumentation__.Notify(244431)
			histogram, err = protoutil.Marshal(h)
			if err != nil {
				__antithesis_instrumentation__.Notify(244432)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244433)
			}
		} else {
			__antithesis_instrumentation__.Notify(244434)
		}
		__antithesis_instrumentation__.Notify(244427)

		columnIDs := tree.NewDArray(types.Int)
		for _, colName := range s.Columns {
			__antithesis_instrumentation__.Notify(244435)
			col, err := desc.FindColumnWithName(tree.Name(colName))
			if err != nil {
				__antithesis_instrumentation__.Notify(244437)
				params.p.BufferClientNotice(
					params.ctx,
					pgnotice.Newf("column %q does not exist", colName),
				)
				continue StatsLoop
			} else {
				__antithesis_instrumentation__.Notify(244438)
			}
			__antithesis_instrumentation__.Notify(244436)
			if err := columnIDs.Append(tree.NewDInt(tree.DInt(col.GetID()))); err != nil {
				__antithesis_instrumentation__.Notify(244439)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244440)
			}
		}
		__antithesis_instrumentation__.Notify(244428)

		if err := insertJSONStatistic(params, desc.GetID(), columnIDs, s, histogram); err != nil {
			__antithesis_instrumentation__.Notify(244441)
			return errors.Wrap(err, "failed to insert stats")
		} else {
			__antithesis_instrumentation__.Notify(244442)
		}
	}
	__antithesis_instrumentation__.Notify(244416)

	params.extendedEvalCtx.ExecCfg.TableStatsCache.InvalidateTableStats(params.ctx, desc.GetID())

	return nil
}

func insertJSONStatistic(
	params runParams,
	tableID descpb.ID,
	columnIDs *tree.DArray,
	s *stats.JSONStatistic,
	histogram interface{},
) error {
	__antithesis_instrumentation__.Notify(244443)
	var (
		ctx      = params.ctx
		ie       = params.ExecCfg().InternalExecutor
		txn      = params.EvalContext().Txn
		settings = params.ExecCfg().Settings
	)

	var name interface{}
	if s.Name != "" {
		__antithesis_instrumentation__.Notify(244446)
		name = s.Name
	} else {
		__antithesis_instrumentation__.Notify(244447)
	}
	__antithesis_instrumentation__.Notify(244444)

	if !settings.Version.IsActive(params.ctx, clusterversion.AlterSystemTableStatisticsAddAvgSizeCol) {
		__antithesis_instrumentation__.Notify(244448)
		_, err := ie.Exec(
			ctx,
			"insert-stats",
			txn,
			`INSERT INTO system.table_statistics (
					"tableID",
					"name",
					"columnIDs",
					"createdAt",
					"rowCount",
					"distinctCount",
					"nullCount",
					histogram
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			tableID,
			name,
			columnIDs,
			s.CreatedAt,
			s.RowCount,
			s.DistinctCount,
			s.NullCount,
			histogram)
		return err
	} else {
		__antithesis_instrumentation__.Notify(244449)
	}
	__antithesis_instrumentation__.Notify(244445)
	_, err := ie.Exec(
		ctx,
		"insert-stats",
		txn,
		`INSERT INTO system.table_statistics (
					"tableID",
					"name",
					"columnIDs",
					"createdAt",
					"rowCount",
					"distinctCount",
					"nullCount",
					"avgSize",
					histogram
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		tableID,
		name,
		columnIDs,
		s.CreatedAt,
		s.RowCount,
		s.DistinctCount,
		s.NullCount,
		s.AvgSize,
		histogram)
	return err
}

func (p *planner) removeColumnComment(
	ctx context.Context, tableID descpb.ID, columnID descpb.ColumnID,
) error {
	__antithesis_instrumentation__.Notify(244450)
	_, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.ExecEx(
		ctx,
		"delete-column-comment",
		p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		"DELETE FROM system.comments WHERE type=$1 AND object_id=$2 AND sub_id=$3",
		keys.ColumnCommentType,
		tableID,
		columnID)

	return err
}

func validateConstraintNameIsNotUsed(
	tableDesc *tabledesc.Mutable, cmd *tree.AlterTableAddConstraint,
) (skipAddConstraint bool, _ error) {
	__antithesis_instrumentation__.Notify(244451)
	var name tree.Name
	var hasIfNotExists bool
	switch d := cmd.ConstraintDef.(type) {
	case *tree.CheckConstraintTableDef:
		__antithesis_instrumentation__.Notify(244458)
		name = d.Name
		hasIfNotExists = d.IfNotExists
	case *tree.ForeignKeyConstraintTableDef:
		__antithesis_instrumentation__.Notify(244459)
		name = d.Name
		hasIfNotExists = d.IfNotExists
	case *tree.UniqueConstraintTableDef:
		__antithesis_instrumentation__.Notify(244460)
		name = d.Name
		hasIfNotExists = d.IfNotExists
		if d.WithoutIndex {
			__antithesis_instrumentation__.Notify(244468)
			break
		} else {
			__antithesis_instrumentation__.Notify(244469)
		}
		__antithesis_instrumentation__.Notify(244461)

		if d.PrimaryKey {
			__antithesis_instrumentation__.Notify(244470)

			if tableDesc.HasPrimaryKey() && func() bool {
				__antithesis_instrumentation__.Notify(244473)
				return !tableDesc.IsPrimaryIndexDefaultRowID() == true
			}() == true {
				__antithesis_instrumentation__.Notify(244474)
				if d.IfNotExists {
					__antithesis_instrumentation__.Notify(244476)
					return true, nil
				} else {
					__antithesis_instrumentation__.Notify(244477)
				}
				__antithesis_instrumentation__.Notify(244475)
				return false, pgerror.Newf(pgcode.InvalidTableDefinition,
					"multiple primary keys for table %q are not allowed", tableDesc.Name)
			} else {
				__antithesis_instrumentation__.Notify(244478)
			}
			__antithesis_instrumentation__.Notify(244471)

			defaultPKName := tabledesc.PrimaryKeyIndexName(tableDesc.GetName())
			if tableDesc.HasPrimaryKey() && func() bool {
				__antithesis_instrumentation__.Notify(244479)
				return tableDesc.IsPrimaryIndexDefaultRowID() == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(244480)
				return tableDesc.PrimaryIndex.GetName() == defaultPKName == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(244481)
				return name == tree.Name(defaultPKName) == true
			}() == true {
				__antithesis_instrumentation__.Notify(244482)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(244483)
			}
			__antithesis_instrumentation__.Notify(244472)

			if !tableDesc.HasPrimaryKey() && func() bool {
				__antithesis_instrumentation__.Notify(244484)
				return tableDesc.PrimaryIndex.Name == name.String() == true
			}() == true {
				__antithesis_instrumentation__.Notify(244485)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(244486)
			}
		} else {
			__antithesis_instrumentation__.Notify(244487)
		}
		__antithesis_instrumentation__.Notify(244462)
		if name == "" {
			__antithesis_instrumentation__.Notify(244488)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(244489)
		}
		__antithesis_instrumentation__.Notify(244463)
		idx, _ := tableDesc.FindIndexWithName(string(name))

		if idx == nil {
			__antithesis_instrumentation__.Notify(244490)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(244491)
		}
		__antithesis_instrumentation__.Notify(244464)
		if d.IfNotExists {
			__antithesis_instrumentation__.Notify(244492)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(244493)
		}
		__antithesis_instrumentation__.Notify(244465)
		if idx.Dropped() {
			__antithesis_instrumentation__.Notify(244494)
			return false, pgerror.Newf(pgcode.DuplicateObject, "constraint with name %q already exists and is being dropped, try again later", name)
		} else {
			__antithesis_instrumentation__.Notify(244495)
		}
		__antithesis_instrumentation__.Notify(244466)
		return false, pgerror.Newf(pgcode.DuplicateObject, "constraint with name %q already exists", name)

	default:
		__antithesis_instrumentation__.Notify(244467)
		return false, errors.AssertionFailedf(
			"unsupported constraint: %T", cmd.ConstraintDef)
	}
	__antithesis_instrumentation__.Notify(244452)

	if name == "" {
		__antithesis_instrumentation__.Notify(244496)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(244497)
	}
	__antithesis_instrumentation__.Notify(244453)
	info, err := tableDesc.GetConstraintInfo()
	if err != nil {
		__antithesis_instrumentation__.Notify(244498)

		return false, errors.WithAssertionFailure(err)
	} else {
		__antithesis_instrumentation__.Notify(244499)
	}
	__antithesis_instrumentation__.Notify(244454)
	constraintInfo, isInUse := info[name.String()]
	if !isInUse {
		__antithesis_instrumentation__.Notify(244500)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(244501)
	}
	__antithesis_instrumentation__.Notify(244455)

	if isInUse && func() bool {
		__antithesis_instrumentation__.Notify(244502)
		return constraintInfo.Index != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(244503)
		return constraintInfo.Index.ID == tableDesc.PrimaryIndex.ID == true
	}() == true {
		__antithesis_instrumentation__.Notify(244504)
		for _, mut := range tableDesc.GetMutations() {
			__antithesis_instrumentation__.Notify(244505)
			if primaryKeySwap := mut.GetPrimaryKeySwap(); primaryKeySwap != nil && func() bool {
				__antithesis_instrumentation__.Notify(244506)
				return primaryKeySwap.OldPrimaryIndexId == tableDesc.PrimaryIndex.ID == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(244507)
				return primaryKeySwap.NewPrimaryIndexName != name.String() == true
			}() == true {
				__antithesis_instrumentation__.Notify(244508)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(244509)
			}
		}

	} else {
		__antithesis_instrumentation__.Notify(244510)
	}
	__antithesis_instrumentation__.Notify(244456)
	if hasIfNotExists {
		__antithesis_instrumentation__.Notify(244511)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(244512)
	}
	__antithesis_instrumentation__.Notify(244457)
	return false, pgerror.Newf(pgcode.DuplicateObject,
		"duplicate constraint name: %q", name)
}

func (p *planner) updateFKBackReferenceName(
	ctx context.Context,
	tableDesc *tabledesc.Mutable,
	ref *descpb.ForeignKeyConstraint,
	newName string,
) error {
	__antithesis_instrumentation__.Notify(244513)
	var referencedTableDesc *tabledesc.Mutable

	if tableDesc.ID == ref.ReferencedTableID {
		__antithesis_instrumentation__.Notify(244517)
		referencedTableDesc = tableDesc
	} else {
		__antithesis_instrumentation__.Notify(244518)
		lookup, err := p.Descriptors().GetMutableTableVersionByID(ctx, ref.ReferencedTableID, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(244520)
			return errors.Wrapf(err, "error resolving referenced table ID %d", ref.ReferencedTableID)
		} else {
			__antithesis_instrumentation__.Notify(244521)
		}
		__antithesis_instrumentation__.Notify(244519)
		referencedTableDesc = lookup
	}
	__antithesis_instrumentation__.Notify(244514)
	if referencedTableDesc.Dropped() {
		__antithesis_instrumentation__.Notify(244522)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(244523)
	}
	__antithesis_instrumentation__.Notify(244515)
	for i := range referencedTableDesc.InboundFKs {
		__antithesis_instrumentation__.Notify(244524)
		backref := &referencedTableDesc.InboundFKs[i]
		if backref.Name == ref.Name && func() bool {
			__antithesis_instrumentation__.Notify(244525)
			return backref.OriginTableID == tableDesc.ID == true
		}() == true {
			__antithesis_instrumentation__.Notify(244526)
			backref.Name = newName
			return p.writeSchemaChange(
				ctx, referencedTableDesc, descpb.InvalidMutationID,
				fmt.Sprintf("updating referenced FK table %s(%d) for table %s(%d)",
					referencedTableDesc.Name, referencedTableDesc.ID, tableDesc.Name, tableDesc.ID),
			)
		} else {
			__antithesis_instrumentation__.Notify(244527)
		}
	}
	__antithesis_instrumentation__.Notify(244516)
	return errors.Errorf("missing backreference for foreign key %s", ref.Name)
}

func dropColumnImpl(
	params runParams, tn *tree.TableName, tableDesc *tabledesc.Mutable, t *tree.AlterTableDropColumn,
) (droppedViews []string, err error) {
	__antithesis_instrumentation__.Notify(244528)
	if tableDesc.IsLocalityRegionalByRow() {
		__antithesis_instrumentation__.Notify(244549)
		rbrColName, err := tableDesc.GetRegionalByRowTableRegionColumnName()
		if err != nil {
			__antithesis_instrumentation__.Notify(244551)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(244552)
		}
		__antithesis_instrumentation__.Notify(244550)
		if rbrColName == t.Column {
			__antithesis_instrumentation__.Notify(244553)
			return nil, errors.WithHintf(
				pgerror.Newf(
					pgcode.InvalidColumnReference,
					"cannot drop column %s as it is used to store the region in a REGIONAL BY ROW table",
					t.Column,
				),
				"You must change the table locality before dropping this table or alter the table to use a different column to use for the region.",
			)
		} else {
			__antithesis_instrumentation__.Notify(244554)
		}
	} else {
		__antithesis_instrumentation__.Notify(244555)
	}
	__antithesis_instrumentation__.Notify(244529)

	colToDrop, err := tableDesc.FindColumnWithName(t.Column)
	if err != nil {
		__antithesis_instrumentation__.Notify(244556)
		if t.IfExists {
			__antithesis_instrumentation__.Notify(244558)

			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(244559)
		}
		__antithesis_instrumentation__.Notify(244557)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244560)
	}
	__antithesis_instrumentation__.Notify(244530)
	if colToDrop.Dropped() {
		__antithesis_instrumentation__.Notify(244561)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(244562)
	}
	__antithesis_instrumentation__.Notify(244531)

	if colToDrop.IsInaccessible() {
		__antithesis_instrumentation__.Notify(244563)
		return nil, pgerror.Newf(
			pgcode.InvalidColumnReference,
			"cannot drop inaccessible column %q",
			t.Column,
		)
	} else {
		__antithesis_instrumentation__.Notify(244564)
	}
	__antithesis_instrumentation__.Notify(244532)

	if colToDrop.NumUsesSequences() > 0 {
		__antithesis_instrumentation__.Notify(244565)
		if err := params.p.removeSequenceDependencies(params.ctx, tableDesc, colToDrop); err != nil {
			__antithesis_instrumentation__.Notify(244566)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(244567)
		}
	} else {
		__antithesis_instrumentation__.Notify(244568)
	}
	__antithesis_instrumentation__.Notify(244533)

	if err := params.p.canRemoveAllColumnOwnedSequences(params.ctx, tableDesc, colToDrop, t.DropBehavior); err != nil {
		__antithesis_instrumentation__.Notify(244569)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244570)
	}
	__antithesis_instrumentation__.Notify(244534)

	if err := params.p.dropSequencesOwnedByCol(params.ctx, colToDrop, true, t.DropBehavior); err != nil {
		__antithesis_instrumentation__.Notify(244571)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244572)
	}
	__antithesis_instrumentation__.Notify(244535)

	for _, ref := range tableDesc.DependedOnBy {
		__antithesis_instrumentation__.Notify(244573)
		found := false
		for _, colID := range ref.ColumnIDs {
			__antithesis_instrumentation__.Notify(244580)
			if colID == colToDrop.GetID() {
				__antithesis_instrumentation__.Notify(244581)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(244582)
			}
		}
		__antithesis_instrumentation__.Notify(244574)
		if !found {
			__antithesis_instrumentation__.Notify(244583)
			continue
		} else {
			__antithesis_instrumentation__.Notify(244584)
		}
		__antithesis_instrumentation__.Notify(244575)
		err := params.p.canRemoveDependentViewGeneric(
			params.ctx, "column", string(t.Column), tableDesc.ParentID, ref, t.DropBehavior,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(244585)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(244586)
		}
		__antithesis_instrumentation__.Notify(244576)
		viewDesc, err := params.p.getViewDescForCascade(
			params.ctx, "column", string(t.Column), tableDesc.ParentID, ref.ID, t.DropBehavior,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(244587)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(244588)
		}
		__antithesis_instrumentation__.Notify(244577)
		jobDesc := fmt.Sprintf("removing view %q dependent on column %q which is being dropped",
			viewDesc.Name, colToDrop.ColName())
		cascadedViews, err := params.p.removeDependentView(params.ctx, tableDesc, viewDesc, jobDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(244589)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(244590)
		}
		__antithesis_instrumentation__.Notify(244578)
		qualifiedView, err := params.p.getQualifiedTableName(params.ctx, viewDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(244591)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(244592)
		}
		__antithesis_instrumentation__.Notify(244579)

		droppedViews = append(droppedViews, cascadedViews...)
		droppedViews = append(droppedViews, qualifiedView.FQString())
	}
	__antithesis_instrumentation__.Notify(244536)

	if err := schemaexpr.ValidateColumnHasNoDependents(tableDesc, colToDrop); err != nil {
		__antithesis_instrumentation__.Notify(244593)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244594)
	}
	__antithesis_instrumentation__.Notify(244537)

	if tableDesc.GetPrimaryIndex().CollectKeyColumnIDs().Contains(colToDrop.GetID()) {
		__antithesis_instrumentation__.Notify(244595)
		return nil, pgerror.Newf(pgcode.InvalidColumnReference,
			"column %q is referenced by the primary key", colToDrop.GetName())
	} else {
		__antithesis_instrumentation__.Notify(244596)
	}
	__antithesis_instrumentation__.Notify(244538)
	var idxNamesToDelete []string
	for _, idx := range tableDesc.NonDropIndexes() {
		__antithesis_instrumentation__.Notify(244597)

		containsThisColumn := false

		for j := 0; j < idx.NumKeyColumns(); j++ {
			__antithesis_instrumentation__.Notify(244602)
			if idx.GetKeyColumnID(j) == colToDrop.GetID() {
				__antithesis_instrumentation__.Notify(244603)
				containsThisColumn = true
				break
			} else {
				__antithesis_instrumentation__.Notify(244604)
			}
		}
		__antithesis_instrumentation__.Notify(244598)
		if !containsThisColumn {
			__antithesis_instrumentation__.Notify(244605)
			for j := 0; j < idx.NumKeySuffixColumns(); j++ {
				__antithesis_instrumentation__.Notify(244606)
				id := idx.GetKeySuffixColumnID(j)
				if tableDesc.GetPrimaryIndex().CollectKeyColumnIDs().Contains(id) {
					__antithesis_instrumentation__.Notify(244608)

					continue
				} else {
					__antithesis_instrumentation__.Notify(244609)
				}
				__antithesis_instrumentation__.Notify(244607)
				if id == colToDrop.GetID() {
					__antithesis_instrumentation__.Notify(244610)
					containsThisColumn = true
					break
				} else {
					__antithesis_instrumentation__.Notify(244611)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(244612)
		}
		__antithesis_instrumentation__.Notify(244599)
		if !containsThisColumn {
			__antithesis_instrumentation__.Notify(244613)

			for j := 0; j < idx.NumSecondaryStoredColumns(); j++ {
				__antithesis_instrumentation__.Notify(244614)
				if idx.GetStoredColumnID(j) == colToDrop.GetID() {
					__antithesis_instrumentation__.Notify(244615)
					containsThisColumn = true
					break
				} else {
					__antithesis_instrumentation__.Notify(244616)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(244617)
		}
		__antithesis_instrumentation__.Notify(244600)

		if !containsThisColumn && func() bool {
			__antithesis_instrumentation__.Notify(244618)
			return idx.IsPartial() == true
		}() == true {
			__antithesis_instrumentation__.Notify(244619)
			expr, err := parser.ParseExpr(idx.GetPredicate())
			if err != nil {
				__antithesis_instrumentation__.Notify(244622)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(244623)
			}
			__antithesis_instrumentation__.Notify(244620)

			colIDs, err := schemaexpr.ExtractColumnIDs(tableDesc, expr)
			if err != nil {
				__antithesis_instrumentation__.Notify(244624)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(244625)
			}
			__antithesis_instrumentation__.Notify(244621)

			if colIDs.Contains(colToDrop.GetID()) {
				__antithesis_instrumentation__.Notify(244626)
				containsThisColumn = true
			} else {
				__antithesis_instrumentation__.Notify(244627)
			}
		} else {
			__antithesis_instrumentation__.Notify(244628)
		}
		__antithesis_instrumentation__.Notify(244601)

		if containsThisColumn {
			__antithesis_instrumentation__.Notify(244629)
			idxNamesToDelete = append(idxNamesToDelete, idx.GetName())
		} else {
			__antithesis_instrumentation__.Notify(244630)
		}
	}
	__antithesis_instrumentation__.Notify(244539)

	for _, idxName := range idxNamesToDelete {
		__antithesis_instrumentation__.Notify(244631)
		jobDesc := fmt.Sprintf(
			"removing index %q dependent on column %q which is being dropped; full details: %s",
			idxName,
			colToDrop.ColName(),
			tree.AsStringWithFQNames(tn, params.Ann()),
		)
		if err := params.p.dropIndexByName(
			params.ctx, tn, tree.UnrestrictedName(idxName), tableDesc, false,
			t.DropBehavior, ignoreIdxConstraint, jobDesc,
		); err != nil {
			__antithesis_instrumentation__.Notify(244632)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(244633)
		}
	}
	__antithesis_instrumentation__.Notify(244540)

	sliceIdx := 0
	for i := range tableDesc.UniqueWithoutIndexConstraints {
		__antithesis_instrumentation__.Notify(244634)
		constraint := &tableDesc.UniqueWithoutIndexConstraints[i]
		tableDesc.UniqueWithoutIndexConstraints[sliceIdx] = *constraint
		sliceIdx++
		if descpb.ColumnIDs(constraint.ColumnIDs).Contains(colToDrop.GetID()) {
			__antithesis_instrumentation__.Notify(244635)
			sliceIdx--

			if err := params.p.tryRemoveFKBackReferences(
				params.ctx, tableDesc, constraint, t.DropBehavior, nil,
			); err != nil {
				__antithesis_instrumentation__.Notify(244636)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(244637)
			}
		} else {
			__antithesis_instrumentation__.Notify(244638)
		}
	}
	__antithesis_instrumentation__.Notify(244541)
	tableDesc.UniqueWithoutIndexConstraints = tableDesc.UniqueWithoutIndexConstraints[:sliceIdx]

	constraintsToDrop := make([]string, 0, len(tableDesc.Checks))
	constraintInfo, err := tableDesc.GetConstraintInfo()
	if err != nil {
		__antithesis_instrumentation__.Notify(244639)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244640)
	}
	__antithesis_instrumentation__.Notify(244542)

	for _, check := range tableDesc.AllActiveAndInactiveChecks() {
		__antithesis_instrumentation__.Notify(244641)
		if used, err := tableDesc.CheckConstraintUsesColumn(check, colToDrop.GetID()); err != nil {
			__antithesis_instrumentation__.Notify(244642)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(244643)
			if used {
				__antithesis_instrumentation__.Notify(244644)
				if check.Validity == descpb.ConstraintValidity_Dropping {
					__antithesis_instrumentation__.Notify(244646)

					continue
				} else {
					__antithesis_instrumentation__.Notify(244647)
				}
				__antithesis_instrumentation__.Notify(244645)
				constraintsToDrop = append(constraintsToDrop, check.Name)
			} else {
				__antithesis_instrumentation__.Notify(244648)
			}
		}
	}
	__antithesis_instrumentation__.Notify(244543)

	for _, constraintName := range constraintsToDrop {
		__antithesis_instrumentation__.Notify(244649)
		err := tableDesc.DropConstraint(params.ctx, constraintName, constraintInfo[constraintName],
			func(*tabledesc.Mutable, *descpb.ForeignKeyConstraint) error {
				__antithesis_instrumentation__.Notify(244651)
				return nil
			},
			params.extendedEvalCtx.Settings,
		)
		__antithesis_instrumentation__.Notify(244650)
		if err != nil {
			__antithesis_instrumentation__.Notify(244652)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(244653)
		}
	}
	__antithesis_instrumentation__.Notify(244544)

	if err := params.p.removeColumnComment(params.ctx, tableDesc.ID, colToDrop.GetID()); err != nil {
		__antithesis_instrumentation__.Notify(244654)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(244655)
	}
	__antithesis_instrumentation__.Notify(244545)

	sliceIdx = 0
	for i := range tableDesc.OutboundFKs {
		__antithesis_instrumentation__.Notify(244656)
		tableDesc.OutboundFKs[sliceIdx] = tableDesc.OutboundFKs[i]
		sliceIdx++
		fk := &tableDesc.OutboundFKs[i]
		if descpb.ColumnIDs(fk.OriginColumnIDs).Contains(colToDrop.GetID()) {
			__antithesis_instrumentation__.Notify(244657)
			sliceIdx--
			if err := params.p.removeFKBackReference(params.ctx, tableDesc, fk); err != nil {
				__antithesis_instrumentation__.Notify(244658)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(244659)
			}
		} else {
			__antithesis_instrumentation__.Notify(244660)
		}
	}
	__antithesis_instrumentation__.Notify(244546)
	tableDesc.OutboundFKs = tableDesc.OutboundFKs[:sliceIdx]

	found := false
	for i := range tableDesc.Columns {
		__antithesis_instrumentation__.Notify(244661)
		if tableDesc.Columns[i].ID == colToDrop.GetID() {
			__antithesis_instrumentation__.Notify(244662)
			tableDesc.AddColumnMutation(colToDrop.ColumnDesc(), descpb.DescriptorMutation_DROP)

			tableDesc.Columns = append(tableDesc.Columns[:i:i], tableDesc.Columns[i+1:]...)
			found = true
			break
		} else {
			__antithesis_instrumentation__.Notify(244663)
		}
	}
	__antithesis_instrumentation__.Notify(244547)
	if !found {
		__antithesis_instrumentation__.Notify(244664)
		return nil, pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
			"column %q in the middle of being added, try again later", t.Column)
	} else {
		__antithesis_instrumentation__.Notify(244665)
	}
	__antithesis_instrumentation__.Notify(244548)

	return droppedViews, validateDescriptor(params.ctx, params.p, tableDesc)
}

func handleTTLStorageParamChange(
	params runParams,
	tn *tree.TableName,
	tableDesc *tabledesc.Mutable,
	before, after *catpb.RowLevelTTL,
) error {
	__antithesis_instrumentation__.Notify(244666)
	switch {
	case before == nil && func() bool {
		__antithesis_instrumentation__.Notify(244680)
		return after == nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(244668)

	case before != nil && func() bool {
		__antithesis_instrumentation__.Notify(244681)
		return after != nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(244669)

		if before.DeletionCron != after.DeletionCron {
			__antithesis_instrumentation__.Notify(244682)
			env := JobSchedulerEnv(params.ExecCfg())
			s, err := jobs.LoadScheduledJob(
				params.ctx,
				env,
				after.ScheduleID,
				params.ExecCfg().InternalExecutor,
				params.p.txn,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(244685)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244686)
			}
			__antithesis_instrumentation__.Notify(244683)
			if err := s.SetSchedule(after.DeletionCronOrDefault()); err != nil {
				__antithesis_instrumentation__.Notify(244687)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244688)
			}
			__antithesis_instrumentation__.Notify(244684)
			if err := s.Update(params.ctx, params.ExecCfg().InternalExecutor, params.p.txn); err != nil {
				__antithesis_instrumentation__.Notify(244689)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244690)
			}
		} else {
			__antithesis_instrumentation__.Notify(244691)
		}
		__antithesis_instrumentation__.Notify(244670)

		if before.DurationExpr != after.DurationExpr {
			__antithesis_instrumentation__.Notify(244692)
			col, err := tableDesc.FindColumnWithName(colinfo.TTLDefaultExpirationColumnName)
			if err != nil {
				__antithesis_instrumentation__.Notify(244696)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244697)
			}
			__antithesis_instrumentation__.Notify(244693)
			intervalExpr, err := parser.ParseExpr(string(after.DurationExpr))
			if err != nil {
				__antithesis_instrumentation__.Notify(244698)
				return errors.Wrapf(err, "unexpected expression for TTL duration")
			} else {
				__antithesis_instrumentation__.Notify(244699)
			}
			__antithesis_instrumentation__.Notify(244694)
			newExpr := rowLevelTTLAutomaticColumnExpr(intervalExpr)

			if err := updateNonComputedColExpr(
				params,
				tableDesc,
				col,
				newExpr,
				&col.ColumnDesc().DefaultExpr,
				"TTL DEFAULT",
			); err != nil {
				__antithesis_instrumentation__.Notify(244700)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244701)
			}
			__antithesis_instrumentation__.Notify(244695)

			if err := updateNonComputedColExpr(
				params,
				tableDesc,
				col,
				newExpr,
				&col.ColumnDesc().OnUpdateExpr,
				"TTL UPDATE",
			); err != nil {
				__antithesis_instrumentation__.Notify(244702)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244703)
			}
		} else {
			__antithesis_instrumentation__.Notify(244704)
		}
	case before == nil && func() bool {
		__antithesis_instrumentation__.Notify(244705)
		return after != nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(244671)
		if err := checkTTLEnabledForCluster(params.ctx, params.p.ExecCfg().Settings); err != nil {
			__antithesis_instrumentation__.Notify(244706)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244707)
		}
		__antithesis_instrumentation__.Notify(244672)

		tableDesc.RowLevelTTL = nil
		if _, err := tableDesc.FindColumnWithName(colinfo.TTLDefaultExpirationColumnName); err == nil {
			__antithesis_instrumentation__.Notify(244708)
			return pgerror.Newf(
				pgcode.InvalidTableDefinition,
				"cannot add TTL to table with the %s column already defined",
				colinfo.TTLDefaultExpirationColumnName,
			)
		} else {
			__antithesis_instrumentation__.Notify(244709)
		}
		__antithesis_instrumentation__.Notify(244673)
		col, err := rowLevelTTLAutomaticColumnDef(after)
		if err != nil {
			__antithesis_instrumentation__.Notify(244710)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244711)
		}
		__antithesis_instrumentation__.Notify(244674)
		addCol := &tree.AlterTableAddColumn{
			ColumnDef: col,
		}
		if err := params.p.addColumnImpl(
			params,
			&alterTableNode{
				tableDesc: tableDesc,
				n: &tree.AlterTable{
					Cmds: []tree.AlterTableCmd{addCol},
				},
			},
			tn,
			tableDesc,
			addCol,
		); err != nil {
			__antithesis_instrumentation__.Notify(244712)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244713)
		}
		__antithesis_instrumentation__.Notify(244675)
		tableDesc.AddModifyRowLevelTTLMutation(
			&descpb.ModifyRowLevelTTL{RowLevelTTL: after},
			descpb.DescriptorMutation_ADD,
		)
		version := params.ExecCfg().Settings.Version.ActiveVersion(params.ctx)
		if err := tableDesc.AllocateIDs(params.ctx, version); err != nil {
			__antithesis_instrumentation__.Notify(244714)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244715)
		}
	case before != nil && func() bool {
		__antithesis_instrumentation__.Notify(244716)
		return after == nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(244676)
		telemetry.Inc(sqltelemetry.RowLevelTTLDropped)

		tableDesc.RowLevelTTL = before

		droppedViews, err := dropColumnImpl(params, tn, tableDesc, &tree.AlterTableDropColumn{
			Column: colinfo.TTLDefaultExpirationColumnName,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(244717)
			return err
		} else {
			__antithesis_instrumentation__.Notify(244718)
		}
		__antithesis_instrumentation__.Notify(244677)

		if len(droppedViews) > 0 {
			__antithesis_instrumentation__.Notify(244719)
			return pgerror.Newf(pgcode.InvalidParameterValue, "cannot drop TTL automatic column if it is depended on by a view")
		} else {
			__antithesis_instrumentation__.Notify(244720)
		}
		__antithesis_instrumentation__.Notify(244678)

		tableDesc.AddModifyRowLevelTTLMutation(
			&descpb.ModifyRowLevelTTL{RowLevelTTL: before},
			descpb.DescriptorMutation_DROP,
		)
	default:
		__antithesis_instrumentation__.Notify(244679)
	}
	__antithesis_instrumentation__.Notify(244667)

	return nil
}

func (p *planner) tryRemoveFKBackReferences(
	ctx context.Context,
	tableDesc *tabledesc.Mutable,
	constraint descpb.UniqueConstraint,
	behavior tree.DropBehavior,
	candidateConstraints []descpb.UniqueConstraint,
) error {
	__antithesis_instrumentation__.Notify(244721)

	uniqueConstraintHasReplacementCandidate := func(
		referencedColumnIDs []descpb.ColumnID,
	) bool {
		__antithesis_instrumentation__.Notify(244724)
		for _, uc := range candidateConstraints {
			__antithesis_instrumentation__.Notify(244726)
			if uc.IsValidReferencedUniqueConstraint(referencedColumnIDs) {
				__antithesis_instrumentation__.Notify(244727)
				return true
			} else {
				__antithesis_instrumentation__.Notify(244728)
			}
		}
		__antithesis_instrumentation__.Notify(244725)
		return false
	}
	__antithesis_instrumentation__.Notify(244722)

	sliceIdx := 0
	for i := range tableDesc.InboundFKs {
		__antithesis_instrumentation__.Notify(244729)
		tableDesc.InboundFKs[sliceIdx] = tableDesc.InboundFKs[i]
		sliceIdx++
		fk := &tableDesc.InboundFKs[i]

		if constraint.IsValidReferencedUniqueConstraint(fk.ReferencedColumnIDs) && func() bool {
			__antithesis_instrumentation__.Notify(244730)
			return !uniqueConstraintHasReplacementCandidate(fk.ReferencedColumnIDs) == true
		}() == true {
			__antithesis_instrumentation__.Notify(244731)

			if err := p.canRemoveFKBackreference(ctx, constraint.GetName(), fk, behavior); err != nil {
				__antithesis_instrumentation__.Notify(244733)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244734)
			}
			__antithesis_instrumentation__.Notify(244732)
			sliceIdx--
			if err := p.removeFKForBackReference(ctx, tableDesc, fk); err != nil {
				__antithesis_instrumentation__.Notify(244735)
				return err
			} else {
				__antithesis_instrumentation__.Notify(244736)
			}
		} else {
			__antithesis_instrumentation__.Notify(244737)
		}
	}
	__antithesis_instrumentation__.Notify(244723)
	tableDesc.InboundFKs = tableDesc.InboundFKs[:sliceIdx]
	return nil
}
