package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/seqexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type createViewNode struct {
	viewName *tree.TableName

	viewQuery    string
	ifNotExists  bool
	replace      bool
	persistence  tree.Persistence
	materialized bool
	dbDesc       catalog.DatabaseDescriptor
	columns      colinfo.ResultColumns

	planDeps planDependencies

	typeDeps typeDependencies
}

func (n *createViewNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(465008) }

func (n *createViewNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(465009)
	tableType := tree.GetTableType(
		false, true, n.materialized,
	)
	if n.replace {
		__antithesis_instrumentation__.Notify(465022)
		telemetry.Inc(sqltelemetry.SchemaChangeCreateCounter(fmt.Sprintf("or_replace_%s", tableType)))
	} else {
		__antithesis_instrumentation__.Notify(465023)
		telemetry.Inc(sqltelemetry.SchemaChangeCreateCounter(tableType))
	}
	__antithesis_instrumentation__.Notify(465010)

	viewName := n.viewName.Object()
	log.VEventf(params.ctx, 2, "dependencies for view %s:\n%s", viewName, n.planDeps.String())

	if !allowCrossDatabaseViews.Get(&params.p.execCfg.Settings.SV) {
		__antithesis_instrumentation__.Notify(465024)
		for _, dep := range n.planDeps {
			__antithesis_instrumentation__.Notify(465025)
			if dbID := dep.desc.GetParentID(); dbID != n.dbDesc.GetID() && func() bool {
				__antithesis_instrumentation__.Notify(465026)
				return dbID != keys.SystemDatabaseID == true
			}() == true {
				__antithesis_instrumentation__.Notify(465027)
				return errors.WithHintf(
					pgerror.Newf(pgcode.FeatureNotSupported,
						"the view cannot refer to other databases; (see the '%s' cluster setting)",
						allowCrossDatabaseViewsSetting),
					crossDBReferenceDeprecationHint(),
				)
			} else {
				__antithesis_instrumentation__.Notify(465028)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(465029)
	}
	__antithesis_instrumentation__.Notify(465011)

	backRefMutables := make(map[descpb.ID]*tabledesc.Mutable, len(n.planDeps))
	hasTempBackref := false
	for id, updated := range n.planDeps {
		__antithesis_instrumentation__.Notify(465030)
		backRefMutable, err := params.p.Descriptors().GetUncommittedMutableTableByID(id)
		if err != nil {
			__antithesis_instrumentation__.Notify(465034)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465035)
		}
		__antithesis_instrumentation__.Notify(465031)
		if backRefMutable == nil {
			__antithesis_instrumentation__.Notify(465036)
			backRefMutable = tabledesc.NewBuilder(updated.desc.TableDesc()).BuildExistingMutableTable()
		} else {
			__antithesis_instrumentation__.Notify(465037)
		}
		__antithesis_instrumentation__.Notify(465032)
		if !n.persistence.IsTemporary() && func() bool {
			__antithesis_instrumentation__.Notify(465038)
			return backRefMutable.Temporary == true
		}() == true {
			__antithesis_instrumentation__.Notify(465039)
			hasTempBackref = true
		} else {
			__antithesis_instrumentation__.Notify(465040)
		}
		__antithesis_instrumentation__.Notify(465033)
		backRefMutables[id] = backRefMutable
	}
	__antithesis_instrumentation__.Notify(465012)
	if hasTempBackref {
		__antithesis_instrumentation__.Notify(465041)
		n.persistence = tree.PersistenceTemporary

		params.p.BufferClientNotice(
			params.ctx,
			pgnotice.Newf(`view "%s" will be a temporary view`, viewName),
		)
	} else {
		__antithesis_instrumentation__.Notify(465042)
	}
	__antithesis_instrumentation__.Notify(465013)

	var replacingDesc *tabledesc.Mutable
	schema, err := getSchemaForCreateTable(params, n.dbDesc, n.persistence, n.viewName,
		tree.ResolveRequireViewDesc, n.ifNotExists)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(465043)
		return !sqlerrors.IsRelationAlreadyExistsError(err) == true
	}() == true {
		__antithesis_instrumentation__.Notify(465044)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465045)
	}
	__antithesis_instrumentation__.Notify(465014)
	if err != nil {
		__antithesis_instrumentation__.Notify(465046)
		switch {
		case n.ifNotExists:
			__antithesis_instrumentation__.Notify(465047)
			return nil
		case n.replace:
			__antithesis_instrumentation__.Notify(465048)

			id, err := params.p.Descriptors().Direct().LookupObjectID(
				params.ctx,
				params.p.txn,
				n.dbDesc.GetID(),
				schema.GetID(),
				n.viewName.Table(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(465054)
				return err
			} else {
				__antithesis_instrumentation__.Notify(465055)
			}
			__antithesis_instrumentation__.Notify(465049)
			desc, err := params.p.Descriptors().GetMutableTableVersionByID(params.ctx, id, params.p.txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(465056)
				return err
			} else {
				__antithesis_instrumentation__.Notify(465057)
			}
			__antithesis_instrumentation__.Notify(465050)
			if err := params.p.CheckPrivilege(params.ctx, desc, privilege.DROP); err != nil {
				__antithesis_instrumentation__.Notify(465058)
				return err
			} else {
				__antithesis_instrumentation__.Notify(465059)
			}
			__antithesis_instrumentation__.Notify(465051)
			if !desc.IsView() {
				__antithesis_instrumentation__.Notify(465060)
				return pgerror.Newf(pgcode.WrongObjectType, `%q is not a view`, viewName)
			} else {
				__antithesis_instrumentation__.Notify(465061)
			}
			__antithesis_instrumentation__.Notify(465052)
			replacingDesc = desc
		default:
			__antithesis_instrumentation__.Notify(465053)
			return err
		}
	} else {
		__antithesis_instrumentation__.Notify(465062)
	}
	__antithesis_instrumentation__.Notify(465015)

	if n.persistence.IsTemporary() {
		__antithesis_instrumentation__.Notify(465063)
		telemetry.Inc(sqltelemetry.CreateTempViewCounter)
	} else {
		__antithesis_instrumentation__.Notify(465064)
	}
	__antithesis_instrumentation__.Notify(465016)

	privs := catprivilege.CreatePrivilegesFromDefaultPrivileges(
		n.dbDesc.GetDefaultPrivilegeDescriptor(),
		schema.GetDefaultPrivilegeDescriptor(),
		n.dbDesc.GetID(),
		params.SessionData().User(),
		tree.Tables,
		n.dbDesc.GetPrivileges(),
	)

	var newDesc *tabledesc.Mutable
	applyGlobalMultiRegionZoneConfig := false

	if replacingDesc != nil {
		__antithesis_instrumentation__.Notify(465065)
		newDesc, err = params.p.replaceViewDesc(
			params.ctx,
			params.p,
			n,
			replacingDesc,
			backRefMutables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(465066)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465067)
		}
	} else {
		__antithesis_instrumentation__.Notify(465068)

		id, err := descidgen.GenerateUniqueDescID(params.ctx, params.p.ExecCfg().DB, params.p.ExecCfg().Codec)
		if err != nil {
			__antithesis_instrumentation__.Notify(465075)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465076)
		}
		__antithesis_instrumentation__.Notify(465069)

		var creationTime hlc.Timestamp
		desc, err := makeViewTableDesc(
			params.ctx,
			viewName,
			n.viewQuery,
			n.dbDesc.GetID(),
			schema.GetID(),
			id,
			n.columns,
			creationTime,
			privs,
			&params.p.semaCtx,
			params.p.EvalContext(),
			params.p.EvalContext().Settings,
			n.persistence,
			n.dbDesc.IsMultiRegion(),
			params.p)
		if err != nil {
			__antithesis_instrumentation__.Notify(465077)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465078)
		}
		__antithesis_instrumentation__.Notify(465070)

		if n.materialized {
			__antithesis_instrumentation__.Notify(465079)

			desc.IsMaterializedView = true
			desc.State = descpb.DescriptorState_ADD
			desc.CreateAsOfTime = params.p.Txn().ReadTimestamp()
			version := params.ExecCfg().Settings.Version.ActiveVersion(params.ctx)
			if err := desc.AllocateIDs(params.ctx, version); err != nil {
				__antithesis_instrumentation__.Notify(465081)
				return err
			} else {
				__antithesis_instrumentation__.Notify(465082)
			}
			__antithesis_instrumentation__.Notify(465080)

			if n.dbDesc.IsMultiRegion() {
				__antithesis_instrumentation__.Notify(465083)
				desc.SetTableLocalityGlobal()
				applyGlobalMultiRegionZoneConfig = true
			} else {
				__antithesis_instrumentation__.Notify(465084)
			}
		} else {
			__antithesis_instrumentation__.Notify(465085)
		}
		__antithesis_instrumentation__.Notify(465071)

		orderedDependsOn := catalog.DescriptorIDSet{}
		for backrefID := range n.planDeps {
			__antithesis_instrumentation__.Notify(465086)
			orderedDependsOn.Add(backrefID)
		}
		__antithesis_instrumentation__.Notify(465072)
		desc.DependsOn = append(desc.DependsOn, orderedDependsOn.Ordered()...)

		orderedTypeDeps := catalog.DescriptorIDSet{}
		for backrefID := range n.typeDeps {
			__antithesis_instrumentation__.Notify(465087)
			orderedTypeDeps.Add(backrefID)
		}
		__antithesis_instrumentation__.Notify(465073)
		desc.DependsOnTypes = append(desc.DependsOnTypes, orderedTypeDeps.Ordered()...)

		if err = params.p.createDescriptorWithID(
			params.ctx,
			catalogkeys.MakeObjectNameKey(params.ExecCfg().Codec, n.dbDesc.GetID(), schema.GetID(), n.viewName.Table()),
			id,
			&desc,
			fmt.Sprintf("CREATE VIEW %q AS %q", n.viewName, n.viewQuery),
		); err != nil {
			__antithesis_instrumentation__.Notify(465088)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465089)
		}
		__antithesis_instrumentation__.Notify(465074)
		newDesc = &desc
	}
	__antithesis_instrumentation__.Notify(465017)

	for id, updated := range n.planDeps {
		__antithesis_instrumentation__.Notify(465090)
		backRefMutable := backRefMutables[id]

		backRefMutable.DependedOnBy = removeMatchingReferences(
			backRefMutable.DependedOnBy,
			newDesc.ID,
		)
		for _, dep := range updated.deps {
			__antithesis_instrumentation__.Notify(465092)

			dep.ID = newDesc.ID
			dep.ByID = updated.desc.IsSequence()
			backRefMutable.DependedOnBy = append(backRefMutable.DependedOnBy, dep)
		}
		__antithesis_instrumentation__.Notify(465091)
		if err := params.p.writeSchemaChange(
			params.ctx,
			backRefMutable,
			descpb.InvalidMutationID,
			fmt.Sprintf("updating view reference %q in table %s(%d)", n.viewName,
				updated.desc.GetName(), updated.desc.GetID(),
			),
		); err != nil {
			__antithesis_instrumentation__.Notify(465093)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465094)
		}
	}
	__antithesis_instrumentation__.Notify(465018)

	for id := range n.typeDeps {
		__antithesis_instrumentation__.Notify(465095)
		jobDesc := fmt.Sprintf("updating type back reference %d for table %d", id, newDesc.ID)
		if err := params.p.addTypeBackReference(params.ctx, id, newDesc.ID, jobDesc); err != nil {
			__antithesis_instrumentation__.Notify(465096)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465097)
		}
	}
	__antithesis_instrumentation__.Notify(465019)

	if err := validateDescriptor(params.ctx, params.p, newDesc); err != nil {
		__antithesis_instrumentation__.Notify(465098)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465099)
	}
	__antithesis_instrumentation__.Notify(465020)

	if applyGlobalMultiRegionZoneConfig {
		__antithesis_instrumentation__.Notify(465100)
		regionConfig, err := SynthesizeRegionConfig(params.ctx, params.p.txn, n.dbDesc.GetID(), params.p.Descriptors())
		if err != nil {
			__antithesis_instrumentation__.Notify(465102)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465103)
		}
		__antithesis_instrumentation__.Notify(465101)
		if err := ApplyZoneConfigForMultiRegionTable(
			params.ctx,
			params.p.txn,
			params.p.ExecCfg(),
			regionConfig,
			newDesc,
			applyZoneConfigForMultiRegionTableOptionTableNewConfig(
				tabledesc.LocalityConfigGlobal(),
			),
		); err != nil {
			__antithesis_instrumentation__.Notify(465104)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465105)
		}
	} else {
		__antithesis_instrumentation__.Notify(465106)
	}
	__antithesis_instrumentation__.Notify(465021)

	return params.p.logEvent(params.ctx,
		newDesc.ID,
		&eventpb.CreateView{
			ViewName:  n.viewName.FQString(),
			ViewQuery: n.viewQuery,
		})
}

func (*createViewNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(465107)
	return false, nil
}
func (*createViewNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(465108)
	return tree.Datums{}
}
func (n *createViewNode) Close(ctx context.Context) { __antithesis_instrumentation__.Notify(465109) }

func makeViewTableDesc(
	ctx context.Context,
	viewName string,
	viewQuery string,
	parentID descpb.ID,
	schemaID descpb.ID,
	id descpb.ID,
	resultColumns []colinfo.ResultColumn,
	creationTime hlc.Timestamp,
	privileges *catpb.PrivilegeDescriptor,
	semaCtx *tree.SemaContext,
	evalCtx *tree.EvalContext,
	st *cluster.Settings,
	persistence tree.Persistence,
	isMultiRegion bool,
	sc resolver.SchemaResolver,
) (tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(465110)
	desc := tabledesc.InitTableDescriptor(
		id,
		parentID,
		schemaID,
		viewName,
		creationTime,
		privileges,
		persistence,
	)
	desc.ViewQuery = viewQuery
	if isMultiRegion {
		__antithesis_instrumentation__.Notify(465115)
		desc.SetTableLocalityRegionalByTable(tree.PrimaryRegionNotSpecifiedName)
	} else {
		__antithesis_instrumentation__.Notify(465116)
	}
	__antithesis_instrumentation__.Notify(465111)

	if sc != nil {
		__antithesis_instrumentation__.Notify(465117)
		sequenceReplacedQuery, err := replaceSeqNamesWithIDs(ctx, sc, viewQuery)
		if err != nil {
			__antithesis_instrumentation__.Notify(465119)
			return tabledesc.Mutable{}, err
		} else {
			__antithesis_instrumentation__.Notify(465120)
		}
		__antithesis_instrumentation__.Notify(465118)
		desc.ViewQuery = sequenceReplacedQuery
	} else {
		__antithesis_instrumentation__.Notify(465121)
	}
	__antithesis_instrumentation__.Notify(465112)

	typeReplacedQuery, err := serializeUserDefinedTypes(ctx, semaCtx, desc.ViewQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(465122)
		return tabledesc.Mutable{}, err
	} else {
		__antithesis_instrumentation__.Notify(465123)
	}
	__antithesis_instrumentation__.Notify(465113)
	desc.ViewQuery = typeReplacedQuery

	if err := addResultColumns(ctx, semaCtx, evalCtx, st, &desc, resultColumns); err != nil {
		__antithesis_instrumentation__.Notify(465124)
		return tabledesc.Mutable{}, err
	} else {
		__antithesis_instrumentation__.Notify(465125)
	}
	__antithesis_instrumentation__.Notify(465114)

	return desc, nil
}

func replaceSeqNamesWithIDs(
	ctx context.Context, sc resolver.SchemaResolver, viewQuery string,
) (string, error) {
	__antithesis_instrumentation__.Notify(465126)
	replaceSeqFunc := func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(465130)
		seqIdentifiers, err := seqexpr.GetUsedSequences(expr)
		if err != nil {
			__antithesis_instrumentation__.Notify(465134)
			return false, expr, err
		} else {
			__antithesis_instrumentation__.Notify(465135)
		}
		__antithesis_instrumentation__.Notify(465131)
		seqNameToID := make(map[string]int64)
		for _, seqIdentifier := range seqIdentifiers {
			__antithesis_instrumentation__.Notify(465136)
			seqDesc, err := GetSequenceDescFromIdentifier(ctx, sc, seqIdentifier)
			if err != nil {
				__antithesis_instrumentation__.Notify(465138)
				return false, expr, err
			} else {
				__antithesis_instrumentation__.Notify(465139)
			}
			__antithesis_instrumentation__.Notify(465137)
			seqNameToID[seqIdentifier.SeqName] = int64(seqDesc.ID)
		}
		__antithesis_instrumentation__.Notify(465132)
		newExpr, err = seqexpr.ReplaceSequenceNamesWithIDs(expr, seqNameToID)
		if err != nil {
			__antithesis_instrumentation__.Notify(465140)
			return false, expr, err
		} else {
			__antithesis_instrumentation__.Notify(465141)
		}
		__antithesis_instrumentation__.Notify(465133)
		return false, newExpr, nil
	}
	__antithesis_instrumentation__.Notify(465127)

	stmt, err := parser.ParseOne(viewQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(465142)
		return "", errors.Wrap(err, "failed to parse view query")
	} else {
		__antithesis_instrumentation__.Notify(465143)
	}
	__antithesis_instrumentation__.Notify(465128)

	newStmt, err := tree.SimpleStmtVisit(stmt.AST, replaceSeqFunc)
	if err != nil {
		__antithesis_instrumentation__.Notify(465144)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(465145)
	}
	__antithesis_instrumentation__.Notify(465129)
	return newStmt.String(), nil
}

func serializeUserDefinedTypes(
	ctx context.Context, semaCtx *tree.SemaContext, viewQuery string,
) (string, error) {
	__antithesis_instrumentation__.Notify(465146)
	replaceFunc := func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(465150)
		switch n := expr.(type) {
		case *tree.CastExpr, *tree.AnnotateTypeExpr:
			__antithesis_instrumentation__.Notify(465152)
			texpr, err := tree.TypeCheck(ctx, n, semaCtx, types.Any)
			if err != nil {
				__antithesis_instrumentation__.Notify(465156)
				return false, expr, err
			} else {
				__antithesis_instrumentation__.Notify(465157)
			}
			__antithesis_instrumentation__.Notify(465153)
			if !texpr.ResolvedType().UserDefined() {
				__antithesis_instrumentation__.Notify(465158)
				return true, expr, nil
			} else {
				__antithesis_instrumentation__.Notify(465159)
			}
			__antithesis_instrumentation__.Notify(465154)

			s := tree.Serialize(texpr)
			parsedExpr, err := parser.ParseExpr(s)
			if err != nil {
				__antithesis_instrumentation__.Notify(465160)
				return false, expr, err
			} else {
				__antithesis_instrumentation__.Notify(465161)
			}
			__antithesis_instrumentation__.Notify(465155)
			return false, parsedExpr, nil
		}
		__antithesis_instrumentation__.Notify(465151)
		return true, expr, nil
	}
	__antithesis_instrumentation__.Notify(465147)

	stmt, err := parser.ParseOne(viewQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(465162)
		return "", errors.Wrap(err, "failed to parse view query")
	} else {
		__antithesis_instrumentation__.Notify(465163)
	}
	__antithesis_instrumentation__.Notify(465148)

	newStmt, err := tree.SimpleStmtVisit(stmt.AST, replaceFunc)
	if err != nil {
		__antithesis_instrumentation__.Notify(465164)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(465165)
	}
	__antithesis_instrumentation__.Notify(465149)
	return newStmt.String(), nil
}

func (p *planner) replaceViewDesc(
	ctx context.Context,
	sc resolver.SchemaResolver,
	n *createViewNode,
	toReplace *tabledesc.Mutable,
	backRefMutables map[descpb.ID]*tabledesc.Mutable,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(465166)

	toReplace.ViewQuery = n.viewQuery

	if sc != nil {
		__antithesis_instrumentation__.Notify(465176)
		updatedQuery, err := replaceSeqNamesWithIDs(ctx, sc, n.viewQuery)
		if err != nil {
			__antithesis_instrumentation__.Notify(465178)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(465179)
		}
		__antithesis_instrumentation__.Notify(465177)
		toReplace.ViewQuery = updatedQuery
	} else {
		__antithesis_instrumentation__.Notify(465180)
	}
	__antithesis_instrumentation__.Notify(465167)

	toReplace.Columns = make([]descpb.ColumnDescriptor, 0, len(n.columns))
	toReplace.NextColumnID = 0
	if err := addResultColumns(ctx, &p.semaCtx, p.EvalContext(), p.EvalContext().Settings, toReplace, n.columns); err != nil {
		__antithesis_instrumentation__.Notify(465181)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465182)
	}
	__antithesis_instrumentation__.Notify(465168)

	if err := verifyReplacingViewColumns(
		toReplace.ClusterVersion().Columns,
		toReplace.Columns,
	); err != nil {
		__antithesis_instrumentation__.Notify(465183)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465184)
	}
	__antithesis_instrumentation__.Notify(465169)

	for _, id := range toReplace.DependsOn {
		__antithesis_instrumentation__.Notify(465185)
		desc, ok := backRefMutables[id]
		if !ok {
			__antithesis_instrumentation__.Notify(465187)
			var err error
			desc, err = p.Descriptors().GetMutableTableVersionByID(ctx, id, p.txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(465189)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(465190)
			}
			__antithesis_instrumentation__.Notify(465188)
			backRefMutables[id] = desc
		} else {
			__antithesis_instrumentation__.Notify(465191)
		}
		__antithesis_instrumentation__.Notify(465186)

		if _, ok := n.planDeps[id]; !ok {
			__antithesis_instrumentation__.Notify(465192)
			desc.DependedOnBy = removeMatchingReferences(desc.DependedOnBy, toReplace.ID)
			if err := p.writeSchemaChange(
				ctx,
				desc,
				descpb.InvalidMutationID,
				fmt.Sprintf("removing view reference for %q from %s(%d)", n.viewName,
					desc.Name, desc.ID,
				),
			); err != nil {
				__antithesis_instrumentation__.Notify(465193)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(465194)
			}
		} else {
			__antithesis_instrumentation__.Notify(465195)
		}
	}
	__antithesis_instrumentation__.Notify(465170)

	var outdatedTypeRefs []descpb.ID
	for _, id := range toReplace.DependsOnTypes {
		__antithesis_instrumentation__.Notify(465196)
		if _, ok := n.typeDeps[id]; !ok {
			__antithesis_instrumentation__.Notify(465197)
			outdatedTypeRefs = append(outdatedTypeRefs, id)
		} else {
			__antithesis_instrumentation__.Notify(465198)
		}
	}
	__antithesis_instrumentation__.Notify(465171)
	jobDesc := fmt.Sprintf("updating type back references %d for table %d", outdatedTypeRefs, toReplace.ID)
	if err := p.removeTypeBackReferences(ctx, outdatedTypeRefs, toReplace.ID, jobDesc); err != nil {
		__antithesis_instrumentation__.Notify(465199)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465200)
	}
	__antithesis_instrumentation__.Notify(465172)

	toReplace.DependsOn = make([]descpb.ID, 0, len(n.planDeps))
	for backrefID := range n.planDeps {
		__antithesis_instrumentation__.Notify(465201)
		toReplace.DependsOn = append(toReplace.DependsOn, backrefID)
	}
	__antithesis_instrumentation__.Notify(465173)
	toReplace.DependsOnTypes = make([]descpb.ID, 0, len(n.typeDeps))
	for backrefID := range n.typeDeps {
		__antithesis_instrumentation__.Notify(465202)
		toReplace.DependsOnTypes = append(toReplace.DependsOnTypes, backrefID)
	}
	__antithesis_instrumentation__.Notify(465174)

	if err := p.writeSchemaChange(ctx, toReplace, descpb.InvalidMutationID,
		fmt.Sprintf("CREATE OR REPLACE VIEW %q AS %q", n.viewName, n.viewQuery),
	); err != nil {
		__antithesis_instrumentation__.Notify(465203)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(465204)
	}
	__antithesis_instrumentation__.Notify(465175)
	return toReplace, nil
}

func addResultColumns(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	evalCtx *tree.EvalContext,
	st *cluster.Settings,
	desc *tabledesc.Mutable,
	resultColumns colinfo.ResultColumns,
) error {
	__antithesis_instrumentation__.Notify(465205)
	for _, colRes := range resultColumns {
		__antithesis_instrumentation__.Notify(465208)
		columnTableDef := tree.ColumnTableDef{Name: tree.Name(colRes.Name), Type: colRes.Typ}

		columnTableDef.Nullable.Nullability = tree.SilentNull

		cdd, err := tabledesc.MakeColumnDefDescs(ctx, &columnTableDef, semaCtx, evalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(465210)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465211)
		}
		__antithesis_instrumentation__.Notify(465209)
		desc.AddColumn(cdd.ColumnDescriptor)
	}
	__antithesis_instrumentation__.Notify(465206)
	version := st.Version.ActiveVersionOrEmpty(ctx)
	if err := desc.AllocateIDs(ctx, version); err != nil {
		__antithesis_instrumentation__.Notify(465212)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465213)
	}
	__antithesis_instrumentation__.Notify(465207)
	return nil
}

func verifyReplacingViewColumns(oldColumns, newColumns []descpb.ColumnDescriptor) error {
	__antithesis_instrumentation__.Notify(465214)
	if len(newColumns) < len(oldColumns) {
		__antithesis_instrumentation__.Notify(465217)
		return pgerror.Newf(pgcode.InvalidTableDefinition, "cannot drop columns from view")
	} else {
		__antithesis_instrumentation__.Notify(465218)
	}
	__antithesis_instrumentation__.Notify(465215)
	for i := range oldColumns {
		__antithesis_instrumentation__.Notify(465219)
		oldCol, newCol := &oldColumns[i], &newColumns[i]
		if oldCol.Name != newCol.Name {
			__antithesis_instrumentation__.Notify(465223)
			return pgerror.Newf(
				pgcode.InvalidTableDefinition,
				`cannot change name of view column %q to %q`,
				oldCol.Name,
				newCol.Name,
			)
		} else {
			__antithesis_instrumentation__.Notify(465224)
		}
		__antithesis_instrumentation__.Notify(465220)
		if !newCol.Type.Identical(oldCol.Type) {
			__antithesis_instrumentation__.Notify(465225)
			return pgerror.Newf(
				pgcode.InvalidTableDefinition,
				`cannot change type of view column %q from %s to %s`,
				oldCol.Name,
				oldCol.Type.String(),
				newCol.Type.String(),
			)
		} else {
			__antithesis_instrumentation__.Notify(465226)
		}
		__antithesis_instrumentation__.Notify(465221)
		if newCol.Hidden != oldCol.Hidden {
			__antithesis_instrumentation__.Notify(465227)
			return pgerror.Newf(
				pgcode.InvalidTableDefinition,
				`cannot change visibility of view column %q`,
				oldCol.Name,
			)
		} else {
			__antithesis_instrumentation__.Notify(465228)
		}
		__antithesis_instrumentation__.Notify(465222)
		if newCol.Nullable != oldCol.Nullable {
			__antithesis_instrumentation__.Notify(465229)
			return pgerror.Newf(
				pgcode.InvalidTableDefinition,
				`cannot change nullability of view column %q`,
				oldCol.Name,
			)
		} else {
			__antithesis_instrumentation__.Notify(465230)
		}
	}
	__antithesis_instrumentation__.Notify(465216)
	return nil
}

func overrideColumnNames(cols colinfo.ResultColumns, newNames tree.NameList) colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(465231)
	res := append(colinfo.ResultColumns(nil), cols...)
	for i := range res {
		__antithesis_instrumentation__.Notify(465233)
		res[i].Name = string(newNames[i])
	}
	__antithesis_instrumentation__.Notify(465232)
	return res
}

func crossDBReferenceDeprecationHint() string {
	__antithesis_instrumentation__.Notify(465234)
	return fmt.Sprintf("Note that cross-database references will be removed in future releases. See: %s",
		docs.ReleaseNotesURL(`#deprecations`))
}
