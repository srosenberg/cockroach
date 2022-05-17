package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/querycache"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/fsm"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

func (ex *connExecutor) execPrepare(
	ctx context.Context, parseCmd PrepareStmt,
) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458535)
	retErr := func(err error) (fsm.Event, fsm.EventPayload) {
		__antithesis_instrumentation__.Notify(458540)
		return ex.makeErrEvent(err, parseCmd.AST)
	}
	__antithesis_instrumentation__.Notify(458536)

	if _, isNoTxn := ex.machine.CurState().(stateNoTxn); isNoTxn {
		__antithesis_instrumentation__.Notify(458541)
		return ex.beginImplicitTxn(ctx, parseCmd.AST)
	} else {
		__antithesis_instrumentation__.Notify(458542)
	}
	__antithesis_instrumentation__.Notify(458537)

	ctx, sp := tracing.EnsureChildSpan(ctx, ex.server.cfg.AmbientCtx.Tracer, "prepare stmt")
	defer sp.Finish()

	if parseCmd.Name != "" {
		__antithesis_instrumentation__.Notify(458543)
		if _, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[parseCmd.Name]; ok {
			__antithesis_instrumentation__.Notify(458544)
			err := pgerror.Newf(
				pgcode.DuplicatePreparedStatement,
				"prepared statement %q already exists", parseCmd.Name,
			)
			return retErr(err)
		} else {
			__antithesis_instrumentation__.Notify(458545)
		}
	} else {
		__antithesis_instrumentation__.Notify(458546)

		ex.deletePreparedStmt(ctx, "")
	}
	__antithesis_instrumentation__.Notify(458538)

	stmt := makeStatement(parseCmd.Statement, ex.generateID())
	_, err := ex.addPreparedStmt(
		ctx,
		parseCmd.Name,
		stmt,
		parseCmd.TypeHints,
		parseCmd.RawTypeHints,
		PreparedStatementOriginWire,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(458547)
		return retErr(err)
	} else {
		__antithesis_instrumentation__.Notify(458548)
	}
	__antithesis_instrumentation__.Notify(458539)

	return nil, nil
}

func (ex *connExecutor) addPreparedStmt(
	ctx context.Context,
	name string,
	stmt Statement,
	placeholderHints tree.PlaceholderTypes,
	rawTypeHints []oid.Oid,
	origin PreparedStatementOrigin,
) (*PreparedStatement, error) {
	__antithesis_instrumentation__.Notify(458549)
	if _, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[name]; ok {
		__antithesis_instrumentation__.Notify(458555)
		return nil, pgerror.Newf(
			pgcode.DuplicatePreparedStatement,
			"prepared statement %q already exists", name,
		)
	} else {
		__antithesis_instrumentation__.Notify(458556)
	}
	__antithesis_instrumentation__.Notify(458550)

	prepared, err := ex.prepare(ctx, stmt, placeholderHints, rawTypeHints, origin)
	if err != nil {
		__antithesis_instrumentation__.Notify(458557)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(458558)
	}
	__antithesis_instrumentation__.Notify(458551)

	if len(prepared.TypeHints) > pgwirebase.MaxPreparedStatementArgs {
		__antithesis_instrumentation__.Notify(458559)
		return nil, pgwirebase.NewProtocolViolationErrorf(
			"more than %d arguments to prepared statement: %d",
			pgwirebase.MaxPreparedStatementArgs, len(prepared.TypeHints))
	} else {
		__antithesis_instrumentation__.Notify(458560)
	}
	__antithesis_instrumentation__.Notify(458552)

	if err := prepared.memAcc.Grow(ctx, int64(len(name))); err != nil {
		__antithesis_instrumentation__.Notify(458561)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(458562)
	}
	__antithesis_instrumentation__.Notify(458553)
	ex.extraTxnState.prepStmtsNamespace.prepStmts[name] = prepared

	prepared.InferredTypes = make([]oid.Oid, len(prepared.Types))
	copy(prepared.InferredTypes, rawTypeHints)
	for i, it := range prepared.InferredTypes {
		__antithesis_instrumentation__.Notify(458563)

		if it == 0 || func() bool {
			__antithesis_instrumentation__.Notify(458564)
			return it == oid.T_unknown == true
		}() == true {
			__antithesis_instrumentation__.Notify(458565)
			t, _ := prepared.ValueType(tree.PlaceholderIdx(i))
			prepared.InferredTypes[i] = t.Oid()
		} else {
			__antithesis_instrumentation__.Notify(458566)
		}
	}
	__antithesis_instrumentation__.Notify(458554)

	return prepared, nil
}

func (ex *connExecutor) prepare(
	ctx context.Context,
	stmt Statement,
	placeholderHints tree.PlaceholderTypes,
	rawTypeHints []oid.Oid,
	origin PreparedStatementOrigin,
) (*PreparedStatement, error) {
	__antithesis_instrumentation__.Notify(458567)

	prepared := &PreparedStatement{
		memAcc:   ex.sessionMon.MakeBoundAccount(),
		refCount: 1,

		createdAt: timeutil.Now(),
		origin:    origin,
	}

	defer prepared.memAcc.Clear(ctx)

	if stmt.AST == nil {
		__antithesis_instrumentation__.Notify(458573)
		return prepared, nil
	} else {
		__antithesis_instrumentation__.Notify(458574)
	}
	__antithesis_instrumentation__.Notify(458568)

	origNumPlaceholders := stmt.NumPlaceholders
	switch stmt.AST.(type) {
	case *tree.Prepare:
		__antithesis_instrumentation__.Notify(458575)

		stmt.NumPlaceholders = 0
	}
	__antithesis_instrumentation__.Notify(458569)

	var flags planFlags
	prepare := func(ctx context.Context, txn *kv.Txn) (err error) {
		__antithesis_instrumentation__.Notify(458576)
		p := &ex.planner
		if origin == PreparedStatementOriginWire {
			__antithesis_instrumentation__.Notify(458580)

			ex.statsCollector.Reset(ex.applicationStats, ex.phaseTimes)
			ex.resetPlanner(ctx, p, txn, ex.server.cfg.Clock.PhysicalTime())
		} else {
			__antithesis_instrumentation__.Notify(458581)
		}
		__antithesis_instrumentation__.Notify(458577)

		if placeholderHints == nil {
			__antithesis_instrumentation__.Notify(458582)
			placeholderHints = make(tree.PlaceholderTypes, stmt.NumPlaceholders)
		} else {
			__antithesis_instrumentation__.Notify(458583)
			if rawTypeHints != nil {
				__antithesis_instrumentation__.Notify(458584)

				for i := range placeholderHints {
					__antithesis_instrumentation__.Notify(458585)
					if placeholderHints[i] == nil {
						__antithesis_instrumentation__.Notify(458586)
						if i >= len(rawTypeHints) {
							__antithesis_instrumentation__.Notify(458588)
							return pgwirebase.NewProtocolViolationErrorf(
								"expected %d arguments, got %d",
								len(placeholderHints),
								len(rawTypeHints),
							)
						} else {
							__antithesis_instrumentation__.Notify(458589)
						}
						__antithesis_instrumentation__.Notify(458587)
						if types.IsOIDUserDefinedType(rawTypeHints[i]) {
							__antithesis_instrumentation__.Notify(458590)
							var err error
							placeholderHints[i], err = ex.planner.ResolveTypeByOID(ctx, rawTypeHints[i])
							if err != nil {
								__antithesis_instrumentation__.Notify(458591)
								return err
							} else {
								__antithesis_instrumentation__.Notify(458592)
							}
						} else {
							__antithesis_instrumentation__.Notify(458593)
						}
					} else {
						__antithesis_instrumentation__.Notify(458594)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(458595)
			}
		}
		__antithesis_instrumentation__.Notify(458578)

		prepared.PrepareMetadata = querycache.PrepareMetadata{
			PlaceholderTypesInfo: tree.PlaceholderTypesInfo{
				TypeHints: placeholderHints,
				Types:     placeholderHints,
			},
		}
		prepared.Statement = stmt.Statement

		prepared.Statement.NumPlaceholders = origNumPlaceholders
		prepared.StatementNoConstants = stmt.StmtNoConstants
		prepared.StatementSummary = stmt.StmtSummary

		stmt.Prepared = prepared

		if err := tree.ProcessPlaceholderAnnotations(&ex.planner.semaCtx, stmt.AST, placeholderHints); err != nil {
			__antithesis_instrumentation__.Notify(458596)
			return err
		} else {
			__antithesis_instrumentation__.Notify(458597)
		}
		__antithesis_instrumentation__.Notify(458579)

		p.stmt = stmt
		p.semaCtx.Annotations = tree.MakeAnnotations(stmt.NumAnnotations)
		flags, err = ex.populatePrepared(ctx, txn, placeholderHints, p)
		return err
	}
	__antithesis_instrumentation__.Notify(458570)

	if err := prepare(ctx, ex.state.mu.txn); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(458598)
		return origin != PreparedStatementOriginSessionMigration == true
	}() == true {
		__antithesis_instrumentation__.Notify(458599)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(458600)
	}
	__antithesis_instrumentation__.Notify(458571)

	if err := prepared.memAcc.Grow(ctx, prepared.MemoryEstimate()); err != nil {
		__antithesis_instrumentation__.Notify(458601)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(458602)
	}
	__antithesis_instrumentation__.Notify(458572)
	ex.updateOptCounters(flags)
	return prepared, nil
}

func (ex *connExecutor) populatePrepared(
	ctx context.Context, txn *kv.Txn, placeholderHints tree.PlaceholderTypes, p *planner,
) (planFlags, error) {
	__antithesis_instrumentation__.Notify(458603)
	if before := ex.server.cfg.TestingKnobs.BeforePrepare; before != nil {
		__antithesis_instrumentation__.Notify(458608)
		if err := before(ctx, ex.planner.stmt.String(), txn); err != nil {
			__antithesis_instrumentation__.Notify(458609)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(458610)
		}
	} else {
		__antithesis_instrumentation__.Notify(458611)
	}
	__antithesis_instrumentation__.Notify(458604)
	stmt := &p.stmt
	if err := p.semaCtx.Placeholders.Init(stmt.NumPlaceholders, placeholderHints); err != nil {
		__antithesis_instrumentation__.Notify(458612)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(458613)
	}
	__antithesis_instrumentation__.Notify(458605)
	p.extendedEvalCtx.PrepareOnly = true
	if err := ex.handleAOST(ctx, p.stmt.AST); err != nil {
		__antithesis_instrumentation__.Notify(458614)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(458615)
	}
	__antithesis_instrumentation__.Notify(458606)

	flags, err := p.prepareUsingOptimizer(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(458616)
		log.VEventf(ctx, 1, "optimizer prepare failed: %v", err)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(458617)
	}
	__antithesis_instrumentation__.Notify(458607)
	log.VEvent(ctx, 2, "optimizer prepare succeeded")

	return flags, nil
}

func (ex *connExecutor) execBind(
	ctx context.Context, bindCmd BindStmt,
) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458618)
	retErr := func(err error) (fsm.Event, fsm.EventPayload) {
		__antithesis_instrumentation__.Notify(458629)
		return eventNonRetriableErr{IsCommit: fsm.False}, eventNonRetriableErrPayload{err: err}
	}
	__antithesis_instrumentation__.Notify(458619)

	ps, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[bindCmd.PreparedStatementName]
	if !ok {
		__antithesis_instrumentation__.Notify(458630)
		return retErr(pgerror.Newf(
			pgcode.InvalidSQLStatementName,
			"unknown prepared statement %q", bindCmd.PreparedStatementName))
	} else {
		__antithesis_instrumentation__.Notify(458631)
	}
	__antithesis_instrumentation__.Notify(458620)

	if _, isNoTxn := ex.machine.CurState().(stateNoTxn); isNoTxn {
		__antithesis_instrumentation__.Notify(458632)
		return ex.beginImplicitTxn(ctx, ps.AST)
	} else {
		__antithesis_instrumentation__.Notify(458633)
	}
	__antithesis_instrumentation__.Notify(458621)

	portalName := bindCmd.PortalName

	if portalName != "" {
		__antithesis_instrumentation__.Notify(458634)
		if _, ok := ex.extraTxnState.prepStmtsNamespace.portals[portalName]; ok {
			__antithesis_instrumentation__.Notify(458636)
			return retErr(pgerror.Newf(
				pgcode.DuplicateCursor, "portal %q already exists", portalName))
		} else {
			__antithesis_instrumentation__.Notify(458637)
		}
		__antithesis_instrumentation__.Notify(458635)
		if _, err := ex.getCursorAccessor().getCursor(portalName); err == nil {
			__antithesis_instrumentation__.Notify(458638)
			return retErr(pgerror.Newf(
				pgcode.DuplicateCursor, "portal %q already exists as cursor", portalName))
		} else {
			__antithesis_instrumentation__.Notify(458639)
		}
	} else {
		__antithesis_instrumentation__.Notify(458640)

		ex.deletePortal(ctx, "")
	}
	__antithesis_instrumentation__.Notify(458622)

	numQArgs := uint16(len(ps.InferredTypes))

	qargs := make(tree.QueryArguments, numQArgs)
	if bindCmd.internalArgs != nil {
		__antithesis_instrumentation__.Notify(458641)
		if len(bindCmd.internalArgs) != int(numQArgs) {
			__antithesis_instrumentation__.Notify(458643)
			return retErr(
				pgwirebase.NewProtocolViolationErrorf(
					"expected %d arguments, got %d", numQArgs, len(bindCmd.internalArgs)))
		} else {
			__antithesis_instrumentation__.Notify(458644)
		}
		__antithesis_instrumentation__.Notify(458642)
		for i, datum := range bindCmd.internalArgs {
			__antithesis_instrumentation__.Notify(458645)
			t := ps.InferredTypes[i]
			if oid := datum.ResolvedType().Oid(); datum != tree.DNull && func() bool {
				__antithesis_instrumentation__.Notify(458647)
				return oid != t == true
			}() == true {
				__antithesis_instrumentation__.Notify(458648)
				return retErr(
					pgwirebase.NewProtocolViolationErrorf(
						"for argument %d expected OID %d, got %d", i, t, oid))
			} else {
				__antithesis_instrumentation__.Notify(458649)
			}
			__antithesis_instrumentation__.Notify(458646)
			qargs[i] = datum
		}
	} else {
		__antithesis_instrumentation__.Notify(458650)
		qArgFormatCodes := bindCmd.ArgFormatCodes

		if len(qArgFormatCodes) != 1 && func() bool {
			__antithesis_instrumentation__.Notify(458655)
			return len(qArgFormatCodes) != int(numQArgs) == true
		}() == true {
			__antithesis_instrumentation__.Notify(458656)
			return retErr(pgwirebase.NewProtocolViolationErrorf(
				"wrong number of format codes specified: %d for %d arguments",
				len(qArgFormatCodes), numQArgs))
		} else {
			__antithesis_instrumentation__.Notify(458657)
		}
		__antithesis_instrumentation__.Notify(458651)

		if len(qArgFormatCodes) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(458658)
			return numQArgs > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(458659)
			fmtCode := qArgFormatCodes[0]
			qArgFormatCodes = make([]pgwirebase.FormatCode, numQArgs)
			for i := range qArgFormatCodes {
				__antithesis_instrumentation__.Notify(458660)
				qArgFormatCodes[i] = fmtCode
			}
		} else {
			__antithesis_instrumentation__.Notify(458661)
		}
		__antithesis_instrumentation__.Notify(458652)

		if len(bindCmd.Args) != int(numQArgs) {
			__antithesis_instrumentation__.Notify(458662)
			return retErr(
				pgwirebase.NewProtocolViolationErrorf(
					"expected %d arguments, got %d", numQArgs, len(bindCmd.Args)))
		} else {
			__antithesis_instrumentation__.Notify(458663)
		}
		__antithesis_instrumentation__.Notify(458653)

		resolve := func(ctx context.Context, txn *kv.Txn) (err error) {
			__antithesis_instrumentation__.Notify(458664)
			ex.statsCollector.Reset(ex.applicationStats, ex.phaseTimes)
			p := &ex.planner
			ex.resetPlanner(ctx, p, txn, ex.server.cfg.Clock.PhysicalTime())
			if err := ex.handleAOST(ctx, ps.AST); err != nil {
				__antithesis_instrumentation__.Notify(458667)
				return err
			} else {
				__antithesis_instrumentation__.Notify(458668)
			}
			__antithesis_instrumentation__.Notify(458665)

			for i, arg := range bindCmd.Args {
				__antithesis_instrumentation__.Notify(458669)
				k := tree.PlaceholderIdx(i)
				t := ps.InferredTypes[i]
				if arg == nil {
					__antithesis_instrumentation__.Notify(458670)

					qargs[k] = tree.DNull
				} else {
					__antithesis_instrumentation__.Notify(458671)
					typ, ok := types.OidToType[t]
					if !ok {
						__antithesis_instrumentation__.Notify(458674)
						var err error
						typ, err = ex.planner.ResolveTypeByOID(ctx, t)
						if err != nil {
							__antithesis_instrumentation__.Notify(458675)
							return err
						} else {
							__antithesis_instrumentation__.Notify(458676)
						}
					} else {
						__antithesis_instrumentation__.Notify(458677)
					}
					__antithesis_instrumentation__.Notify(458672)
					d, err := pgwirebase.DecodeDatum(
						ex.planner.EvalContext(),
						typ,
						qArgFormatCodes[i],
						arg,
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(458678)
						return pgerror.Wrapf(err, pgcode.ProtocolViolation, "error in argument for %s", k)
					} else {
						__antithesis_instrumentation__.Notify(458679)
					}
					__antithesis_instrumentation__.Notify(458673)
					qargs[k] = d
				}
			}
			__antithesis_instrumentation__.Notify(458666)
			return nil
		}
		__antithesis_instrumentation__.Notify(458654)

		if err := resolve(ctx, ex.state.mu.txn); err != nil {
			__antithesis_instrumentation__.Notify(458680)
			return retErr(err)
		} else {
			__antithesis_instrumentation__.Notify(458681)
		}
	}
	__antithesis_instrumentation__.Notify(458623)

	numCols := len(ps.Columns)
	if (len(bindCmd.OutFormats) > 1) && func() bool {
		__antithesis_instrumentation__.Notify(458682)
		return (len(bindCmd.OutFormats) != numCols) == true
	}() == true {
		__antithesis_instrumentation__.Notify(458683)
		return retErr(pgwirebase.NewProtocolViolationErrorf(
			"expected 1 or %d for number of format codes, got %d",
			numCols, len(bindCmd.OutFormats)))
	} else {
		__antithesis_instrumentation__.Notify(458684)
	}
	__antithesis_instrumentation__.Notify(458624)

	columnFormatCodes := bindCmd.OutFormats
	if len(bindCmd.OutFormats) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(458685)
		return numCols > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(458686)

		columnFormatCodes = make([]pgwirebase.FormatCode, numCols)
		for i := 0; i < numCols; i++ {
			__antithesis_instrumentation__.Notify(458687)
			columnFormatCodes[i] = bindCmd.OutFormats[0]
		}
	} else {
		__antithesis_instrumentation__.Notify(458688)
	}
	__antithesis_instrumentation__.Notify(458625)

	if ex.getTransactionState() == NoTxnStateStr {
		__antithesis_instrumentation__.Notify(458689)
		ex.planner.Descriptors().ReleaseAll(ctx)
	} else {
		__antithesis_instrumentation__.Notify(458690)
	}
	__antithesis_instrumentation__.Notify(458626)

	if err := ex.addPortal(ctx, portalName, ps, qargs, columnFormatCodes); err != nil {
		__antithesis_instrumentation__.Notify(458691)
		return retErr(err)
	} else {
		__antithesis_instrumentation__.Notify(458692)
	}
	__antithesis_instrumentation__.Notify(458627)

	if log.V(2) {
		__antithesis_instrumentation__.Notify(458693)
		log.Infof(ctx, "portal: %q for %q, args %q, formats %q",
			portalName, ps.Statement, qargs, columnFormatCodes)
	} else {
		__antithesis_instrumentation__.Notify(458694)
	}
	__antithesis_instrumentation__.Notify(458628)

	return nil, nil
}

func (ex *connExecutor) addPortal(
	ctx context.Context,
	portalName string,
	stmt *PreparedStatement,
	qargs tree.QueryArguments,
	outFormats []pgwirebase.FormatCode,
) error {
	__antithesis_instrumentation__.Notify(458695)
	if _, ok := ex.extraTxnState.prepStmtsNamespace.portals[portalName]; ok {
		__antithesis_instrumentation__.Notify(458699)
		panic(errors.AssertionFailedf("portal already exists: %q", portalName))
	} else {
		__antithesis_instrumentation__.Notify(458700)
	}
	__antithesis_instrumentation__.Notify(458696)
	if _, err := ex.getCursorAccessor().getCursor(portalName); err == nil {
		__antithesis_instrumentation__.Notify(458701)
		panic(errors.AssertionFailedf("portal already exists as cursor: %q", portalName))
	} else {
		__antithesis_instrumentation__.Notify(458702)
	}
	__antithesis_instrumentation__.Notify(458697)

	portal, err := ex.makePreparedPortal(ctx, portalName, stmt, qargs, outFormats)
	if err != nil {
		__antithesis_instrumentation__.Notify(458703)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458704)
	}
	__antithesis_instrumentation__.Notify(458698)

	ex.extraTxnState.prepStmtsNamespace.portals[portalName] = portal
	return nil
}

func (ex *connExecutor) exhaustPortal(portalName string) {
	__antithesis_instrumentation__.Notify(458705)
	portal, ok := ex.extraTxnState.prepStmtsNamespace.portals[portalName]
	if !ok {
		__antithesis_instrumentation__.Notify(458707)
		panic(errors.AssertionFailedf("portal %s doesn't exist", portalName))
	} else {
		__antithesis_instrumentation__.Notify(458708)
	}
	__antithesis_instrumentation__.Notify(458706)
	portal.exhausted = true
	ex.extraTxnState.prepStmtsNamespace.portals[portalName] = portal
}

func (ex *connExecutor) deletePreparedStmt(ctx context.Context, name string) {
	__antithesis_instrumentation__.Notify(458709)
	ps, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[name]
	if !ok {
		__antithesis_instrumentation__.Notify(458711)
		return
	} else {
		__antithesis_instrumentation__.Notify(458712)
	}
	__antithesis_instrumentation__.Notify(458710)
	ps.decRef(ctx)
	delete(ex.extraTxnState.prepStmtsNamespace.prepStmts, name)
}

func (ex *connExecutor) deletePortal(ctx context.Context, name string) {
	__antithesis_instrumentation__.Notify(458713)
	portal, ok := ex.extraTxnState.prepStmtsNamespace.portals[name]
	if !ok {
		__antithesis_instrumentation__.Notify(458715)
		return
	} else {
		__antithesis_instrumentation__.Notify(458716)
	}
	__antithesis_instrumentation__.Notify(458714)
	portal.close(ctx, &ex.extraTxnState.prepStmtsNamespaceMemAcc, name)
	delete(ex.extraTxnState.prepStmtsNamespace.portals, name)
}

func (ex *connExecutor) execDelPrepStmt(
	ctx context.Context, delCmd DeletePreparedStmt,
) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458717)
	switch delCmd.Type {
	case pgwirebase.PrepareStatement:
		__antithesis_instrumentation__.Notify(458719)
		_, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[delCmd.Name]
		if !ok {
			__antithesis_instrumentation__.Notify(458724)

			break
		} else {
			__antithesis_instrumentation__.Notify(458725)
		}
		__antithesis_instrumentation__.Notify(458720)

		ex.deletePreparedStmt(ctx, delCmd.Name)
	case pgwirebase.PreparePortal:
		__antithesis_instrumentation__.Notify(458721)
		_, ok := ex.extraTxnState.prepStmtsNamespace.portals[delCmd.Name]
		if !ok {
			__antithesis_instrumentation__.Notify(458726)
			break
		} else {
			__antithesis_instrumentation__.Notify(458727)
		}
		__antithesis_instrumentation__.Notify(458722)
		ex.deletePortal(ctx, delCmd.Name)
	default:
		__antithesis_instrumentation__.Notify(458723)
		panic(errors.AssertionFailedf("unknown del type: %s", delCmd.Type))
	}
	__antithesis_instrumentation__.Notify(458718)
	return nil, nil
}

func (ex *connExecutor) execDescribe(
	ctx context.Context, descCmd DescribeStmt, res DescribeResult,
) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458728)

	retErr := func(err error) (fsm.Event, fsm.EventPayload) {
		__antithesis_instrumentation__.Notify(458731)
		return eventNonRetriableErr{IsCommit: fsm.False}, eventNonRetriableErrPayload{err: err}
	}
	__antithesis_instrumentation__.Notify(458729)

	switch descCmd.Type {
	case pgwirebase.PrepareStatement:
		__antithesis_instrumentation__.Notify(458732)
		ps, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[descCmd.Name]
		if !ok {
			__antithesis_instrumentation__.Notify(458738)
			return retErr(pgerror.Newf(
				pgcode.InvalidSQLStatementName,
				"unknown prepared statement %q", descCmd.Name))
		} else {
			__antithesis_instrumentation__.Notify(458739)
		}
		__antithesis_instrumentation__.Notify(458733)

		res.SetInferredTypes(ps.InferredTypes)

		ast := ps.AST
		if execute, ok := ast.(*tree.Execute); ok {
			__antithesis_instrumentation__.Notify(458740)

			innerPs, found := ex.extraTxnState.prepStmtsNamespace.prepStmts[string(execute.Name)]
			if !found {
				__antithesis_instrumentation__.Notify(458742)
				return retErr(pgerror.Newf(
					pgcode.InvalidSQLStatementName,
					"unknown prepared statement %q", descCmd.Name))
			} else {
				__antithesis_instrumentation__.Notify(458743)
			}
			__antithesis_instrumentation__.Notify(458741)
			ast = innerPs.AST
		} else {
			__antithesis_instrumentation__.Notify(458744)
		}
		__antithesis_instrumentation__.Notify(458734)
		if stmtHasNoData(ast) {
			__antithesis_instrumentation__.Notify(458745)
			res.SetNoDataRowDescription()
		} else {
			__antithesis_instrumentation__.Notify(458746)
			res.SetPrepStmtOutput(ctx, ps.Columns)
		}
	case pgwirebase.PreparePortal:
		__antithesis_instrumentation__.Notify(458735)
		portal, ok := ex.extraTxnState.prepStmtsNamespace.portals[descCmd.Name]
		if !ok {
			__antithesis_instrumentation__.Notify(458747)
			return retErr(pgerror.Newf(
				pgcode.InvalidCursorName, "unknown portal %q", descCmd.Name))
		} else {
			__antithesis_instrumentation__.Notify(458748)
		}
		__antithesis_instrumentation__.Notify(458736)

		if stmtHasNoData(portal.Stmt.AST) {
			__antithesis_instrumentation__.Notify(458749)
			res.SetNoDataRowDescription()
		} else {
			__antithesis_instrumentation__.Notify(458750)
			res.SetPortalOutput(ctx, portal.Stmt.Columns, portal.OutFormats)
		}
	default:
		__antithesis_instrumentation__.Notify(458737)
		return retErr(errors.AssertionFailedf(
			"unknown describe type: %s", errors.Safe(descCmd.Type)))
	}
	__antithesis_instrumentation__.Notify(458730)
	return nil, nil
}
