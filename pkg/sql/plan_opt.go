package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/opt"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec/execbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec/explain"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/indexrec"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/memo"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/optbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/xform"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/querycache"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

var queryCacheEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.query_cache.enabled", "enable the query cache", true,
)

func (p *planner) prepareUsingOptimizer(ctx context.Context) (planFlags, error) {
	__antithesis_instrumentation__.Notify(562790)
	stmt := &p.stmt

	opc := &p.optPlanningCtx
	opc.reset()

	switch t := stmt.AST.(type) {
	case *tree.AlterIndex, *tree.AlterTable, *tree.AlterSequence,
		*tree.Analyze,
		*tree.BeginTransaction,
		*tree.CommentOnColumn, *tree.CommentOnConstraint, *tree.CommentOnDatabase, *tree.CommentOnIndex, *tree.CommentOnTable, *tree.CommentOnSchema,
		*tree.CommitTransaction,
		*tree.CopyFrom, *tree.CreateDatabase, *tree.CreateIndex, *tree.CreateView,
		*tree.CreateSequence,
		*tree.CreateStats,
		*tree.Deallocate, *tree.Discard, *tree.DropDatabase, *tree.DropIndex,
		*tree.DropTable, *tree.DropView, *tree.DropSequence, *tree.DropType,
		*tree.Grant, *tree.GrantRole,
		*tree.Prepare,
		*tree.ReleaseSavepoint, *tree.RenameColumn, *tree.RenameDatabase,
		*tree.RenameIndex, *tree.RenameTable, *tree.Revoke, *tree.RevokeRole,
		*tree.RollbackToSavepoint, *tree.RollbackTransaction,
		*tree.Savepoint, *tree.SetTransaction, *tree.SetTracing, *tree.SetSessionAuthorizationDefault,
		*tree.SetSessionCharacteristics:
		__antithesis_instrumentation__.Notify(562797)

		return opc.flags, nil

	case *tree.Execute:
		__antithesis_instrumentation__.Notify(562798)

		name := string(t.Name)
		prepared, ok := p.preparedStatements.Get(name)
		if !ok {
			__antithesis_instrumentation__.Notify(562802)

			return opc.flags, pgerror.Newf(pgcode.UndefinedPreparedStatement,
				"no such prepared statement %s", name)
		} else {
			__antithesis_instrumentation__.Notify(562803)
		}
		__antithesis_instrumentation__.Notify(562799)
		stmt.Prepared.Columns = prepared.Columns
		return opc.flags, nil

	case *tree.ExplainAnalyze:
		__antithesis_instrumentation__.Notify(562800)

		if len(p.semaCtx.Placeholders.Types) != 0 {
			__antithesis_instrumentation__.Notify(562804)
			return 0, errors.Errorf("%s does not support placeholders", stmt.AST.StatementTag())
		} else {
			__antithesis_instrumentation__.Notify(562805)
		}
		__antithesis_instrumentation__.Notify(562801)
		stmt.Prepared.Columns = colinfo.ExplainPlanColumns
		return opc.flags, nil
	}
	__antithesis_instrumentation__.Notify(562791)

	if opc.useCache {
		__antithesis_instrumentation__.Notify(562806)
		cachedData, ok := p.execCfg.QueryCache.Find(&p.queryCacheSession, stmt.SQL)
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(562808)
			return cachedData.PrepareMetadata != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(562809)
			pm := cachedData.PrepareMetadata

			if !pm.TypeHints.Identical(p.semaCtx.Placeholders.TypeHints) {
				__antithesis_instrumentation__.Notify(562810)
				opc.log(ctx, "query cache hit but type hints don't match")
			} else {
				__antithesis_instrumentation__.Notify(562811)
				isStale, err := cachedData.Memo.IsStale(ctx, p.EvalContext(), &opc.catalog)
				if err != nil {
					__antithesis_instrumentation__.Notify(562814)
					return 0, err
				} else {
					__antithesis_instrumentation__.Notify(562815)
				}
				__antithesis_instrumentation__.Notify(562812)
				if !isStale {
					__antithesis_instrumentation__.Notify(562816)
					opc.log(ctx, "query cache hit (prepare)")
					opc.flags.Set(planFlagOptCacheHit)
					stmt.Prepared.StatementNoConstants = pm.StatementNoConstants
					stmt.Prepared.Columns = pm.Columns
					stmt.Prepared.Types = pm.Types
					stmt.Prepared.Memo = cachedData.Memo
					return opc.flags, nil
				} else {
					__antithesis_instrumentation__.Notify(562817)
				}
				__antithesis_instrumentation__.Notify(562813)
				opc.log(ctx, "query cache hit but memo is stale (prepare)")
			}
		} else {
			__antithesis_instrumentation__.Notify(562818)
			if ok {
				__antithesis_instrumentation__.Notify(562819)
				opc.log(ctx, "query cache hit but there is no prepare metadata")
			} else {
				__antithesis_instrumentation__.Notify(562820)
				opc.log(ctx, "query cache miss")
			}
		}
		__antithesis_instrumentation__.Notify(562807)
		opc.flags.Set(planFlagOptCacheMiss)
	} else {
		__antithesis_instrumentation__.Notify(562821)
	}
	__antithesis_instrumentation__.Notify(562792)

	memo, err := opc.buildReusableMemo(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(562822)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(562823)
	}
	__antithesis_instrumentation__.Notify(562793)

	md := memo.Metadata()
	physical := memo.RootProps()
	resultCols := make(colinfo.ResultColumns, len(physical.Presentation))
	for i, col := range physical.Presentation {
		__antithesis_instrumentation__.Notify(562824)
		colMeta := md.ColumnMeta(col.ID)
		resultCols[i].Name = col.Alias
		resultCols[i].Typ = colMeta.Type
		if err := checkResultType(resultCols[i].Typ); err != nil {
			__antithesis_instrumentation__.Notify(562826)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(562827)
		}
		__antithesis_instrumentation__.Notify(562825)

		if colMeta.Table != opt.TableID(0) {
			__antithesis_instrumentation__.Notify(562828)

			tab := md.Table(colMeta.Table)
			resultCols[i].TableID = descpb.ID(tab.ID())

			colOrdinal := colMeta.Table.ColumnOrdinal(col.ID)

			var column catalog.Column
			if catTable, ok := tab.(optCatalogTableInterface); ok {
				__antithesis_instrumentation__.Notify(562830)
				column = catTable.getCol(colOrdinal)
			} else {
				__antithesis_instrumentation__.Notify(562831)
			}
			__antithesis_instrumentation__.Notify(562829)
			if column != nil {
				__antithesis_instrumentation__.Notify(562832)
				resultCols[i].PGAttributeNum = column.GetPGAttributeNum()
			} else {
				__antithesis_instrumentation__.Notify(562833)
				resultCols[i].PGAttributeNum = uint32(tab.Column(colOrdinal).ColID())
			}
		} else {
			__antithesis_instrumentation__.Notify(562834)
		}
	}
	__antithesis_instrumentation__.Notify(562794)

	if err := p.semaCtx.Placeholders.Types.AssertAllSet(); err != nil {
		__antithesis_instrumentation__.Notify(562835)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(562836)
	}
	__antithesis_instrumentation__.Notify(562795)

	stmt.Prepared.Columns = resultCols
	stmt.Prepared.Types = p.semaCtx.Placeholders.Types
	if opc.allowMemoReuse {
		__antithesis_instrumentation__.Notify(562837)
		stmt.Prepared.Memo = memo
		if opc.useCache {
			__antithesis_instrumentation__.Notify(562838)

			pm := stmt.Prepared.PrepareMetadata
			cachedData := querycache.CachedData{
				SQL:             stmt.SQL,
				Memo:            memo,
				PrepareMetadata: &pm,
			}
			p.execCfg.QueryCache.Add(&p.queryCacheSession, &cachedData)
		} else {
			__antithesis_instrumentation__.Notify(562839)
		}
	} else {
		__antithesis_instrumentation__.Notify(562840)
	}
	__antithesis_instrumentation__.Notify(562796)
	return opc.flags, nil
}

func (p *planner) makeOptimizerPlan(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(562841)
	p.curPlan.init(&p.stmt, &p.instrumentation)

	opc := &p.optPlanningCtx
	opc.reset()

	execMemo, err := opc.buildExecMemo(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(562844)
		return err
	} else {
		__antithesis_instrumentation__.Notify(562845)
	}
	__antithesis_instrumentation__.Notify(562842)

	if mode := p.SessionData().ExperimentalDistSQLPlanningMode; mode != sessiondatapb.ExperimentalDistSQLPlanningOff {
		__antithesis_instrumentation__.Notify(562846)
		planningMode := distSQLDefaultPlanning

		if p.Descriptors().HasUncommittedTypes() {
			__antithesis_instrumentation__.Notify(562849)
			planningMode = distSQLLocalOnlyPlanning
		} else {
			__antithesis_instrumentation__.Notify(562850)
		}
		__antithesis_instrumentation__.Notify(562847)
		err := opc.runExecBuilder(
			&p.curPlan,
			&p.stmt,
			newDistSQLSpecExecFactory(p, planningMode),
			execMemo,
			p.EvalContext(),
			p.autoCommit,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(562851)
			if mode == sessiondatapb.ExperimentalDistSQLPlanningAlways && func() bool {
				__antithesis_instrumentation__.Notify(562852)
				return !strings.Contains(p.stmt.AST.StatementTag(), "SET") == true
			}() == true {
				__antithesis_instrumentation__.Notify(562853)

				return err
			} else {
				__antithesis_instrumentation__.Notify(562854)
			}

		} else {
			__antithesis_instrumentation__.Notify(562855)

			m := p.curPlan.main
			isPartiallyDistributed := m.physPlan.Distribution == physicalplan.PartiallyDistributedPlan
			if isPartiallyDistributed && func() bool {
				__antithesis_instrumentation__.Notify(562857)
				return p.SessionData().PartiallyDistributedPlansDisabled == true
			}() == true {
				__antithesis_instrumentation__.Notify(562858)

				err = opc.runExecBuilder(
					&p.curPlan,
					&p.stmt,
					newDistSQLSpecExecFactory(p, distSQLLocalOnlyPlanning),
					execMemo,
					p.EvalContext(),
					p.autoCommit,
				)
			} else {
				__antithesis_instrumentation__.Notify(562859)
			}
			__antithesis_instrumentation__.Notify(562856)
			if err == nil {
				__antithesis_instrumentation__.Notify(562860)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(562861)
			}
		}
		__antithesis_instrumentation__.Notify(562848)

		log.Infof(
			ctx, "distSQLSpecExecFactory failed planning with %v, falling back to the old path", err,
		)
	} else {
		__antithesis_instrumentation__.Notify(562862)
	}
	__antithesis_instrumentation__.Notify(562843)

	return opc.runExecBuilder(
		&p.curPlan,
		&p.stmt,
		newExecFactory(p),
		execMemo,
		p.EvalContext(),
		p.autoCommit,
	)
}

type optPlanningCtx struct {
	p *planner

	catalog optCatalog

	optimizer xform.Optimizer

	allowMemoReuse bool

	useCache bool

	flags planFlags
}

func (opc *optPlanningCtx) init(p *planner) {
	opc.p = p
	opc.catalog.init(p)
}

func (opc *optPlanningCtx) reset() {
	__antithesis_instrumentation__.Notify(562863)
	p := opc.p
	opc.catalog.reset()
	opc.optimizer.Init(p.EvalContext(), &opc.catalog)
	opc.flags = 0

	switch p.stmt.AST.(type) {
	case *tree.ParenSelect, *tree.Select, *tree.SelectClause, *tree.UnionClause, *tree.ValuesClause,
		*tree.Insert, *tree.Update, *tree.Delete, *tree.CannedOptPlan:
		__antithesis_instrumentation__.Notify(562864)

		opc.allowMemoReuse = !p.Descriptors().HasUncommittedTables()
		opc.useCache = opc.allowMemoReuse && func() bool {
			__antithesis_instrumentation__.Notify(562866)
			return queryCacheEnabled.Get(&p.execCfg.Settings.SV) == true
		}() == true

		if _, isCanned := p.stmt.AST.(*tree.CannedOptPlan); isCanned {
			__antithesis_instrumentation__.Notify(562867)

			opc.useCache = false
		} else {
			__antithesis_instrumentation__.Notify(562868)
		}

	default:
		__antithesis_instrumentation__.Notify(562865)
		opc.allowMemoReuse = false
		opc.useCache = false
	}
}

func (opc *optPlanningCtx) log(ctx context.Context, msg string) {
	__antithesis_instrumentation__.Notify(562869)
	if log.VDepth(1, 1) {
		__antithesis_instrumentation__.Notify(562870)
		log.InfofDepth(ctx, 1, "%s: %s", redact.Safe(msg), opc.p.stmt)
	} else {
		__antithesis_instrumentation__.Notify(562871)
		log.Event(ctx, msg)
	}
}

func (opc *optPlanningCtx) buildReusableMemo(ctx context.Context) (_ *memo.Memo, _ error) {
	__antithesis_instrumentation__.Notify(562872)
	p := opc.p

	_, isCanned := opc.p.stmt.AST.(*tree.CannedOptPlan)
	if isCanned {
		__antithesis_instrumentation__.Notify(562879)
		if !p.EvalContext().SessionData().AllowPrepareAsOptPlan {
			__antithesis_instrumentation__.Notify(562881)
			return nil, pgerror.New(pgcode.InsufficientPrivilege,
				"PREPARE AS OPT PLAN is a testing facility that should not be used directly",
			)
		} else {
			__antithesis_instrumentation__.Notify(562882)
		}
		__antithesis_instrumentation__.Notify(562880)

		if !p.SessionData().User().IsRootUser() {
			__antithesis_instrumentation__.Notify(562883)
			return nil, pgerror.New(pgcode.InsufficientPrivilege,
				"PREPARE AS OPT PLAN may only be used by root",
			)
		} else {
			__antithesis_instrumentation__.Notify(562884)
		}
	} else {
		__antithesis_instrumentation__.Notify(562885)
	}
	__antithesis_instrumentation__.Notify(562873)

	if p.SessionData().SaveTablesPrefix != "" && func() bool {
		__antithesis_instrumentation__.Notify(562886)
		return !p.SessionData().User().IsRootUser() == true
	}() == true {
		__antithesis_instrumentation__.Notify(562887)
		return nil, pgerror.New(pgcode.InsufficientPrivilege,
			"sub-expression tables creation may only be used by root",
		)
	} else {
		__antithesis_instrumentation__.Notify(562888)
	}
	__antithesis_instrumentation__.Notify(562874)

	f := opc.optimizer.Factory()
	bld := optbuilder.New(ctx, &p.semaCtx, p.EvalContext(), &opc.catalog, f, opc.p.stmt.AST)
	bld.KeepPlaceholders = true
	if err := bld.Build(); err != nil {
		__antithesis_instrumentation__.Notify(562889)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(562890)
	}
	__antithesis_instrumentation__.Notify(562875)

	if bld.DisableMemoReuse {
		__antithesis_instrumentation__.Notify(562891)

		opc.allowMemoReuse = false
		opc.useCache = false
	} else {
		__antithesis_instrumentation__.Notify(562892)
	}
	__antithesis_instrumentation__.Notify(562876)

	if isCanned {
		__antithesis_instrumentation__.Notify(562893)
		if f.Memo().HasPlaceholders() {
			__antithesis_instrumentation__.Notify(562895)

			return nil, pgerror.Newf(pgcode.Syntax,
				"placeholders are not supported with PREPARE AS OPT PLAN")
		} else {
			__antithesis_instrumentation__.Notify(562896)
		}
		__antithesis_instrumentation__.Notify(562894)

		return opc.optimizer.DetachMemo(), nil
	} else {
		__antithesis_instrumentation__.Notify(562897)
	}
	__antithesis_instrumentation__.Notify(562877)

	if f.Memo().HasPlaceholders() {
		__antithesis_instrumentation__.Notify(562898)

		_, ok, err := opc.optimizer.TryPlaceholderFastPath()
		if err != nil {
			__antithesis_instrumentation__.Notify(562900)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(562901)
		}
		__antithesis_instrumentation__.Notify(562899)
		if ok {
			__antithesis_instrumentation__.Notify(562902)
			opc.log(ctx, "placeholder fast path")
		} else {
			__antithesis_instrumentation__.Notify(562903)
		}
	} else {
		__antithesis_instrumentation__.Notify(562904)

		if !f.FoldingControl().PreventedStableFold() {
			__antithesis_instrumentation__.Notify(562905)
			opc.log(ctx, "optimizing (no placeholders)")
			if _, err := opc.optimizer.Optimize(); err != nil {
				__antithesis_instrumentation__.Notify(562906)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(562907)
			}
		} else {
			__antithesis_instrumentation__.Notify(562908)
		}
	}
	__antithesis_instrumentation__.Notify(562878)

	return opc.optimizer.DetachMemo(), nil
}

func (opc *optPlanningCtx) reuseMemo(cachedMemo *memo.Memo) (*memo.Memo, error) {
	__antithesis_instrumentation__.Notify(562909)
	if cachedMemo.IsOptimized() {
		__antithesis_instrumentation__.Notify(562913)

		return cachedMemo, nil
	} else {
		__antithesis_instrumentation__.Notify(562914)
	}
	__antithesis_instrumentation__.Notify(562910)
	f := opc.optimizer.Factory()

	f.FoldingControl().AllowStableFolds()
	if err := f.AssignPlaceholders(cachedMemo); err != nil {
		__antithesis_instrumentation__.Notify(562915)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(562916)
	}
	__antithesis_instrumentation__.Notify(562911)
	if _, err := opc.optimizer.Optimize(); err != nil {
		__antithesis_instrumentation__.Notify(562917)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(562918)
	}
	__antithesis_instrumentation__.Notify(562912)
	return f.Memo(), nil
}

func (opc *optPlanningCtx) buildExecMemo(ctx context.Context) (_ *memo.Memo, _ error) {
	__antithesis_instrumentation__.Notify(562919)
	prepared := opc.p.stmt.Prepared
	p := opc.p
	if opc.allowMemoReuse && func() bool {
		__antithesis_instrumentation__.Notify(562926)
		return prepared != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(562927)
		return prepared.Memo != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(562928)

		if isStale, err := prepared.Memo.IsStale(ctx, p.EvalContext(), &opc.catalog); err != nil {
			__antithesis_instrumentation__.Notify(562930)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(562931)
			if isStale {
				__antithesis_instrumentation__.Notify(562932)
				prepared.Memo, err = opc.buildReusableMemo(ctx)
				opc.log(ctx, "rebuilding cached memo")
				if err != nil {
					__antithesis_instrumentation__.Notify(562933)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(562934)
				}
			} else {
				__antithesis_instrumentation__.Notify(562935)
			}
		}
		__antithesis_instrumentation__.Notify(562929)
		opc.log(ctx, "reusing cached memo")
		memo, err := opc.reuseMemo(prepared.Memo)
		return memo, err
	} else {
		__antithesis_instrumentation__.Notify(562936)
	}
	__antithesis_instrumentation__.Notify(562920)

	if opc.useCache {
		__antithesis_instrumentation__.Notify(562937)

		cachedData, ok := p.execCfg.QueryCache.Find(&p.queryCacheSession, opc.p.stmt.SQL)
		if ok {
			__antithesis_instrumentation__.Notify(562939)
			if isStale, err := cachedData.Memo.IsStale(ctx, p.EvalContext(), &opc.catalog); err != nil {
				__antithesis_instrumentation__.Notify(562941)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(562942)
				if isStale {
					__antithesis_instrumentation__.Notify(562943)
					cachedData.Memo, err = opc.buildReusableMemo(ctx)
					if err != nil {
						__antithesis_instrumentation__.Notify(562945)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(562946)
					}
					__antithesis_instrumentation__.Notify(562944)

					cachedData.PrepareMetadata = nil
					p.execCfg.QueryCache.Add(&p.queryCacheSession, &cachedData)
					opc.log(ctx, "query cache hit but needed update")
					opc.flags.Set(planFlagOptCacheMiss)
				} else {
					__antithesis_instrumentation__.Notify(562947)
					opc.log(ctx, "query cache hit")
					opc.flags.Set(planFlagOptCacheHit)
				}
			}
			__antithesis_instrumentation__.Notify(562940)
			memo, err := opc.reuseMemo(cachedData.Memo)
			return memo, err
		} else {
			__antithesis_instrumentation__.Notify(562948)
		}
		__antithesis_instrumentation__.Notify(562938)
		opc.flags.Set(planFlagOptCacheMiss)
		opc.log(ctx, "query cache miss")
	} else {
		__antithesis_instrumentation__.Notify(562949)
		opc.log(ctx, "not using query cache")
	}
	__antithesis_instrumentation__.Notify(562921)

	f := opc.optimizer.Factory()
	f.FoldingControl().AllowStableFolds()
	bld := optbuilder.New(ctx, &p.semaCtx, p.EvalContext(), &opc.catalog, f, opc.p.stmt.AST)
	if err := bld.Build(); err != nil {
		__antithesis_instrumentation__.Notify(562950)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(562951)
	}
	__antithesis_instrumentation__.Notify(562922)

	_, isExplain := opc.p.stmt.AST.(*tree.Explain)
	if isExplain && func() bool {
		__antithesis_instrumentation__.Notify(562952)
		return p.SessionData().IndexRecommendationsEnabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(562953)
		if err := opc.makeQueryIndexRecommendation(); err != nil {
			__antithesis_instrumentation__.Notify(562954)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(562955)
		}
	} else {
		__antithesis_instrumentation__.Notify(562956)
	}
	__antithesis_instrumentation__.Notify(562923)

	if _, isCanned := opc.p.stmt.AST.(*tree.CannedOptPlan); !isCanned {
		__antithesis_instrumentation__.Notify(562957)
		if _, err := opc.optimizer.Optimize(); err != nil {
			__antithesis_instrumentation__.Notify(562958)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(562959)
		}
	} else {
		__antithesis_instrumentation__.Notify(562960)
	}
	__antithesis_instrumentation__.Notify(562924)

	if opc.useCache && func() bool {
		__antithesis_instrumentation__.Notify(562961)
		return !bld.HadPlaceholders == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(562962)
		return !bld.DisableMemoReuse == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(562963)
		return !f.FoldingControl().PermittedStableFold() == true
	}() == true {
		__antithesis_instrumentation__.Notify(562964)
		memo := opc.optimizer.DetachMemo()
		cachedData := querycache.CachedData{
			SQL:  opc.p.stmt.SQL,
			Memo: memo,
		}
		p.execCfg.QueryCache.Add(&p.queryCacheSession, &cachedData)
		opc.log(ctx, "query cache add")
		return memo, nil
	} else {
		__antithesis_instrumentation__.Notify(562965)
	}
	__antithesis_instrumentation__.Notify(562925)

	return f.Memo(), nil
}

func (opc *optPlanningCtx) runExecBuilder(
	planTop *planTop,
	stmt *Statement,
	f exec.Factory,
	mem *memo.Memo,
	evalCtx *tree.EvalContext,
	allowAutoCommit bool,
) error {
	__antithesis_instrumentation__.Notify(562966)
	var result *planComponents
	var isDDL bool
	var containsFullTableScan bool
	var containsFullIndexScan bool
	var containsLargeFullTableScan bool
	var containsLargeFullIndexScan bool
	var containsMutation bool
	var gf *explain.PlanGistFactory
	if !opc.p.SessionData().DisablePlanGists {
		__antithesis_instrumentation__.Notify(562978)
		gf = explain.NewPlanGistFactory(f)
		f = gf
	} else {
		__antithesis_instrumentation__.Notify(562979)
	}
	__antithesis_instrumentation__.Notify(562967)
	if !planTop.instrumentation.ShouldBuildExplainPlan() {
		__antithesis_instrumentation__.Notify(562980)
		bld := execbuilder.New(f, &opc.optimizer, mem, &opc.catalog, mem.RootExpr(), evalCtx, allowAutoCommit)
		plan, err := bld.Build()
		if err != nil {
			__antithesis_instrumentation__.Notify(562982)
			return err
		} else {
			__antithesis_instrumentation__.Notify(562983)
		}
		__antithesis_instrumentation__.Notify(562981)
		result = plan.(*planComponents)
		isDDL = bld.IsDDL
		containsFullTableScan = bld.ContainsFullTableScan
		containsFullIndexScan = bld.ContainsFullIndexScan
		containsLargeFullTableScan = bld.ContainsLargeFullTableScan
		containsLargeFullIndexScan = bld.ContainsLargeFullIndexScan
		containsMutation = bld.ContainsMutation
	} else {
		__antithesis_instrumentation__.Notify(562984)

		explainFactory := explain.NewFactory(f)
		bld := execbuilder.New(
			explainFactory, &opc.optimizer, mem, &opc.catalog, mem.RootExpr(), evalCtx, allowAutoCommit,
		)
		plan, err := bld.Build()
		if err != nil {
			__antithesis_instrumentation__.Notify(562986)
			return err
		} else {
			__antithesis_instrumentation__.Notify(562987)
		}
		__antithesis_instrumentation__.Notify(562985)
		explainPlan := plan.(*explain.Plan)
		result = explainPlan.WrappedPlan.(*planComponents)
		isDDL = bld.IsDDL
		containsFullTableScan = bld.ContainsFullTableScan
		containsFullIndexScan = bld.ContainsFullIndexScan
		containsLargeFullTableScan = bld.ContainsLargeFullTableScan
		containsLargeFullIndexScan = bld.ContainsLargeFullIndexScan
		containsMutation = bld.ContainsMutation

		planTop.instrumentation.RecordExplainPlan(explainPlan)
	}
	__antithesis_instrumentation__.Notify(562968)
	if gf != nil {
		__antithesis_instrumentation__.Notify(562988)
		planTop.instrumentation.planGist = gf.PlanGist()
	} else {
		__antithesis_instrumentation__.Notify(562989)
	}
	__antithesis_instrumentation__.Notify(562969)
	planTop.instrumentation.costEstimate = float64(mem.RootExpr().(memo.RelExpr).Cost())

	if stmt.ExpectedTypes != nil {
		__antithesis_instrumentation__.Notify(562990)
		cols := result.main.planColumns()
		if !stmt.ExpectedTypes.TypesEqual(cols) {
			__antithesis_instrumentation__.Notify(562991)
			return pgerror.New(pgcode.FeatureNotSupported, "cached plan must not change result type")
		} else {
			__antithesis_instrumentation__.Notify(562992)
		}
	} else {
		__antithesis_instrumentation__.Notify(562993)
	}
	__antithesis_instrumentation__.Notify(562970)

	planTop.planComponents = *result
	planTop.stmt = stmt
	planTop.flags = opc.flags
	if isDDL {
		__antithesis_instrumentation__.Notify(562994)
		planTop.flags.Set(planFlagIsDDL)
	} else {
		__antithesis_instrumentation__.Notify(562995)
	}
	__antithesis_instrumentation__.Notify(562971)
	if containsFullTableScan {
		__antithesis_instrumentation__.Notify(562996)
		planTop.flags.Set(planFlagContainsFullTableScan)
	} else {
		__antithesis_instrumentation__.Notify(562997)
	}
	__antithesis_instrumentation__.Notify(562972)
	if containsFullIndexScan {
		__antithesis_instrumentation__.Notify(562998)
		planTop.flags.Set(planFlagContainsFullIndexScan)
	} else {
		__antithesis_instrumentation__.Notify(562999)
	}
	__antithesis_instrumentation__.Notify(562973)
	if containsLargeFullTableScan {
		__antithesis_instrumentation__.Notify(563000)
		planTop.flags.Set(planFlagContainsLargeFullTableScan)
	} else {
		__antithesis_instrumentation__.Notify(563001)
	}
	__antithesis_instrumentation__.Notify(562974)
	if containsLargeFullIndexScan {
		__antithesis_instrumentation__.Notify(563002)
		planTop.flags.Set(planFlagContainsLargeFullIndexScan)
	} else {
		__antithesis_instrumentation__.Notify(563003)
	}
	__antithesis_instrumentation__.Notify(562975)
	if containsMutation {
		__antithesis_instrumentation__.Notify(563004)
		planTop.flags.Set(planFlagContainsMutation)
	} else {
		__antithesis_instrumentation__.Notify(563005)
	}
	__antithesis_instrumentation__.Notify(562976)
	if planTop.instrumentation.ShouldSaveMemo() {
		__antithesis_instrumentation__.Notify(563006)
		planTop.mem = mem
		planTop.catalog = &opc.catalog
	} else {
		__antithesis_instrumentation__.Notify(563007)
	}
	__antithesis_instrumentation__.Notify(562977)
	return nil
}

func (p *planner) DecodeGist(gist string) ([]string, error) {
	__antithesis_instrumentation__.Notify(563008)
	return explain.DecodePlanGistToRows(gist, &p.optPlanningCtx.catalog)
}

func (opc *optPlanningCtx) makeQueryIndexRecommendation() error {
	__antithesis_instrumentation__.Notify(563009)

	savedMemo := opc.optimizer.DetachMemo()

	f := opc.optimizer.Factory()
	f.FoldingControl().AllowStableFolds()
	f.CopyAndReplace(
		savedMemo.RootExpr().(memo.RelExpr),
		savedMemo.RootProps(),
		f.CopyWithoutAssigningPlaceholders,
	)
	opc.optimizer.NotifyOnMatchedRule(func(ruleName opt.RuleName) bool {
		__antithesis_instrumentation__.Notify(563013)
		return ruleName.IsNormalize()
	})
	__antithesis_instrumentation__.Notify(563010)
	if _, err := opc.optimizer.Optimize(); err != nil {
		__antithesis_instrumentation__.Notify(563014)
		return err
	} else {
		__antithesis_instrumentation__.Notify(563015)
	}
	__antithesis_instrumentation__.Notify(563011)

	indexCandidates := indexrec.FindIndexCandidateSet(f.Memo().RootExpr(), f.Metadata())
	optTables, hypTables := indexrec.BuildOptAndHypTableMaps(indexCandidates)

	opc.optimizer.Init(f.EvalContext(), &opc.catalog)
	f.CopyAndReplace(
		savedMemo.RootExpr().(memo.RelExpr),
		savedMemo.RootProps(),
		f.CopyWithoutAssigningPlaceholders,
	)
	opc.optimizer.Memo().Metadata().UpdateTableMeta(hypTables)
	if _, err := opc.optimizer.Optimize(); err != nil {
		__antithesis_instrumentation__.Notify(563016)
		return err
	} else {
		__antithesis_instrumentation__.Notify(563017)
	}
	__antithesis_instrumentation__.Notify(563012)

	indexRecommendations := indexrec.FindIndexRecommendationSet(f.Memo().RootExpr(), f.Metadata())
	opc.p.instrumentation.indexRecommendations = indexRecommendations.Output()

	opc.optimizer.Init(f.EvalContext(), &opc.catalog)
	savedMemo.Metadata().UpdateTableMeta(optTables)
	f.CopyAndReplace(
		savedMemo.RootExpr().(memo.RelExpr),
		savedMemo.RootProps(),
		f.CopyWithoutAssigningPlaceholders,
	)

	return nil
}
