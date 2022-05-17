package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type scrubNode struct {
	optColumnsSlot

	n *tree.Scrub

	run scrubRun
}

type checkOperation interface {
	Started() bool

	Start(params runParams) error

	Next(params runParams) (tree.Datums, error)

	Done(context.Context) bool

	Close(context.Context)
}

func (p *planner) Scrub(ctx context.Context, n *tree.Scrub) (planNode, error) {
	__antithesis_instrumentation__.Notify(595461)
	if err := p.RequireAdminRole(ctx, "SCRUB"); err != nil {
		__antithesis_instrumentation__.Notify(595463)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(595464)
	}
	__antithesis_instrumentation__.Notify(595462)
	return &scrubNode{n: n}, nil
}

type scrubRun struct {
	checkQueue []checkOperation
	row        tree.Datums
}

func (n *scrubNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(595465)
	switch n.n.Typ {
	case tree.ScrubTable:
		__antithesis_instrumentation__.Notify(595467)

		tableDesc, err := params.p.ResolveExistingObjectEx(
			params.ctx, n.n.Table, true, tree.ResolveRequireTableDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(595472)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595473)
		}
		__antithesis_instrumentation__.Notify(595468)
		tn, ok := params.p.ResolvedName(n.n.Table).(*tree.TableName)
		if !ok {
			__antithesis_instrumentation__.Notify(595474)
			return errors.AssertionFailedf("%q was not resolved as a table", n.n.Table)
		} else {
			__antithesis_instrumentation__.Notify(595475)
		}
		__antithesis_instrumentation__.Notify(595469)
		if err := n.startScrubTable(params.ctx, params.p, tableDesc, tn); err != nil {
			__antithesis_instrumentation__.Notify(595476)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595477)
		}
	case tree.ScrubDatabase:
		__antithesis_instrumentation__.Notify(595470)
		if err := n.startScrubDatabase(params.ctx, params.p, &n.n.Database); err != nil {
			__antithesis_instrumentation__.Notify(595478)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595479)
		}
	default:
		__antithesis_instrumentation__.Notify(595471)
		return errors.AssertionFailedf("unexpected SCRUB type received, got: %v", n.n.Typ)
	}
	__antithesis_instrumentation__.Notify(595466)
	return nil
}

func (n *scrubNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(595480)
	for len(n.run.checkQueue) > 0 {
		__antithesis_instrumentation__.Notify(595482)
		nextCheck := n.run.checkQueue[0]
		if !nextCheck.Started() {
			__antithesis_instrumentation__.Notify(595485)
			if err := nextCheck.Start(params); err != nil {
				__antithesis_instrumentation__.Notify(595486)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(595487)
			}
		} else {
			__antithesis_instrumentation__.Notify(595488)
		}
		__antithesis_instrumentation__.Notify(595483)

		if !nextCheck.Done(params.ctx) {
			__antithesis_instrumentation__.Notify(595489)
			var err error
			n.run.row, err = nextCheck.Next(params)
			if err != nil {
				__antithesis_instrumentation__.Notify(595491)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(595492)
			}
			__antithesis_instrumentation__.Notify(595490)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(595493)
		}
		__antithesis_instrumentation__.Notify(595484)

		nextCheck.Close(params.ctx)

		n.run.checkQueue = n.run.checkQueue[1:]
	}
	__antithesis_instrumentation__.Notify(595481)
	return false, nil
}

func (n *scrubNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(595494)
	return n.run.row
}

func (n *scrubNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595495)

	for len(n.run.checkQueue) > 0 {
		__antithesis_instrumentation__.Notify(595496)
		n.run.checkQueue[0].Close(ctx)
		n.run.checkQueue = n.run.checkQueue[1:]
	}
}

func (n *scrubNode) startScrubDatabase(ctx context.Context, p *planner, name *tree.Name) error {
	__antithesis_instrumentation__.Notify(595497)

	database := string(*name)
	dbDesc, err := p.Descriptors().GetImmutableDatabaseByName(ctx, p.txn,
		database, tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(595502)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595503)
	}
	__antithesis_instrumentation__.Notify(595498)

	schemas, err := p.Descriptors().GetSchemasForDatabase(ctx, p.txn, dbDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(595504)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595505)
	}
	__antithesis_instrumentation__.Notify(595499)

	var tbNames tree.TableNames
	for _, schema := range schemas {
		__antithesis_instrumentation__.Notify(595506)
		toAppend, _, err := resolver.GetObjectNamesAndIDs(
			ctx, p.txn, p, p.ExecCfg().Codec, dbDesc, schema, true,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(595508)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595509)
		}
		__antithesis_instrumentation__.Notify(595507)
		tbNames = append(tbNames, toAppend...)
	}
	__antithesis_instrumentation__.Notify(595500)

	for i := range tbNames {
		__antithesis_instrumentation__.Notify(595510)
		tableName := &tbNames[i]
		_, objDesc, err := p.Accessor().GetObjectDesc(
			ctx, p.txn, tableName.Catalog(), tableName.Schema(), tableName.Table(),
			p.ObjectLookupFlags(true, false),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(595513)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595514)
		}
		__antithesis_instrumentation__.Notify(595511)
		tableDesc := objDesc.(catalog.TableDescriptor)

		if !tableDesc.IsTable() {
			__antithesis_instrumentation__.Notify(595515)
			continue
		} else {
			__antithesis_instrumentation__.Notify(595516)
		}
		__antithesis_instrumentation__.Notify(595512)
		if err := n.startScrubTable(ctx, p, tableDesc, tableName); err != nil {
			__antithesis_instrumentation__.Notify(595517)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595518)
		}
	}
	__antithesis_instrumentation__.Notify(595501)
	return nil
}

func (n *scrubNode) startScrubTable(
	ctx context.Context, p *planner, tableDesc catalog.TableDescriptor, tableName *tree.TableName,
) error {
	__antithesis_instrumentation__.Notify(595519)
	ts, hasTS, err := p.getTimestamp(ctx, n.n.AsOf)
	if err != nil {
		__antithesis_instrumentation__.Notify(595523)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595524)
	}
	__antithesis_instrumentation__.Notify(595520)

	var indexesSet bool
	var physicalCheckSet bool
	var constraintsSet bool
	for _, option := range n.n.Options {
		__antithesis_instrumentation__.Notify(595525)
		switch v := option.(type) {
		case *tree.ScrubOptionIndex:
			__antithesis_instrumentation__.Notify(595526)
			if indexesSet {
				__antithesis_instrumentation__.Notify(595536)
				return pgerror.Newf(pgcode.Syntax,
					"cannot specify INDEX option more than once")
			} else {
				__antithesis_instrumentation__.Notify(595537)
			}
			__antithesis_instrumentation__.Notify(595527)
			indexesSet = true
			checks, err := createIndexCheckOperations(v.IndexNames, tableDesc, tableName, ts)
			if err != nil {
				__antithesis_instrumentation__.Notify(595538)
				return err
			} else {
				__antithesis_instrumentation__.Notify(595539)
			}
			__antithesis_instrumentation__.Notify(595528)
			n.run.checkQueue = append(n.run.checkQueue, checks...)

		case *tree.ScrubOptionPhysical:
			__antithesis_instrumentation__.Notify(595529)
			if physicalCheckSet {
				__antithesis_instrumentation__.Notify(595540)
				return pgerror.Newf(pgcode.Syntax,
					"cannot specify PHYSICAL option more than once")
			} else {
				__antithesis_instrumentation__.Notify(595541)
			}
			__antithesis_instrumentation__.Notify(595530)
			if hasTS {
				__antithesis_instrumentation__.Notify(595542)
				return pgerror.Newf(pgcode.Syntax,
					"cannot use AS OF SYSTEM TIME with PHYSICAL option")
			} else {
				__antithesis_instrumentation__.Notify(595543)
			}
			__antithesis_instrumentation__.Notify(595531)
			physicalCheckSet = true
			return pgerror.Newf(pgcode.FeatureNotSupported, "PHYSICAL scrub not implemented")

		case *tree.ScrubOptionConstraint:
			__antithesis_instrumentation__.Notify(595532)
			if constraintsSet {
				__antithesis_instrumentation__.Notify(595544)
				return pgerror.Newf(pgcode.Syntax,
					"cannot specify CONSTRAINT option more than once")
			} else {
				__antithesis_instrumentation__.Notify(595545)
			}
			__antithesis_instrumentation__.Notify(595533)
			constraintsSet = true
			constraintsToCheck, err := createConstraintCheckOperations(
				ctx, p, v.ConstraintNames, tableDesc, tableName, ts)
			if err != nil {
				__antithesis_instrumentation__.Notify(595546)
				return err
			} else {
				__antithesis_instrumentation__.Notify(595547)
			}
			__antithesis_instrumentation__.Notify(595534)
			n.run.checkQueue = append(n.run.checkQueue, constraintsToCheck...)

		default:
			__antithesis_instrumentation__.Notify(595535)
			panic(errors.AssertionFailedf("unhandled SCRUB option received: %+v", v))
		}
	}
	__antithesis_instrumentation__.Notify(595521)

	if len(n.n.Options) == 0 {
		__antithesis_instrumentation__.Notify(595548)
		indexesToCheck, err := createIndexCheckOperations(nil, tableDesc, tableName,
			ts)
		if err != nil {
			__antithesis_instrumentation__.Notify(595551)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595552)
		}
		__antithesis_instrumentation__.Notify(595549)
		n.run.checkQueue = append(n.run.checkQueue, indexesToCheck...)
		constraintsToCheck, err := createConstraintCheckOperations(
			ctx, p, nil, tableDesc, tableName, ts)
		if err != nil {
			__antithesis_instrumentation__.Notify(595553)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595554)
		}
		__antithesis_instrumentation__.Notify(595550)
		n.run.checkQueue = append(n.run.checkQueue, constraintsToCheck...)

	} else {
		__antithesis_instrumentation__.Notify(595555)
	}
	__antithesis_instrumentation__.Notify(595522)
	return nil
}

func getPrimaryColIdxs(
	tableDesc catalog.TableDescriptor, columns []catalog.Column,
) (primaryColIdxs []int, err error) {
	__antithesis_instrumentation__.Notify(595556)
	for i := 0; i < tableDesc.GetPrimaryIndex().NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(595558)
		colID := tableDesc.GetPrimaryIndex().GetKeyColumnID(i)
		rowIdx := -1
		for idx, col := range columns {
			__antithesis_instrumentation__.Notify(595561)
			if col.GetID() == colID {
				__antithesis_instrumentation__.Notify(595562)
				rowIdx = idx
				break
			} else {
				__antithesis_instrumentation__.Notify(595563)
			}
		}
		__antithesis_instrumentation__.Notify(595559)
		if rowIdx == -1 {
			__antithesis_instrumentation__.Notify(595564)
			return nil, errors.Errorf(
				"could not find primary index column in projection: columnID=%d columnName=%s",
				colID,
				tableDesc.GetPrimaryIndex().GetKeyColumnName(i))
		} else {
			__antithesis_instrumentation__.Notify(595565)
		}
		__antithesis_instrumentation__.Notify(595560)
		primaryColIdxs = append(primaryColIdxs, rowIdx)
	}
	__antithesis_instrumentation__.Notify(595557)
	return primaryColIdxs, nil
}

func colRef(tableAlias string, columnName string) string {
	__antithesis_instrumentation__.Notify(595566)
	u := tree.UnrestrictedName(columnName)
	if tableAlias == "" {
		__antithesis_instrumentation__.Notify(595568)
		return u.String()
	} else {
		__antithesis_instrumentation__.Notify(595569)
	}
	__antithesis_instrumentation__.Notify(595567)
	return fmt.Sprintf("%s.%s", tableAlias, &u)
}

func colRefs(tableAlias string, columnNames []string) []string {
	__antithesis_instrumentation__.Notify(595570)
	res := make([]string, len(columnNames))
	for i := range res {
		__antithesis_instrumentation__.Notify(595572)
		res[i] = colRef(tableAlias, columnNames[i])
	}
	__antithesis_instrumentation__.Notify(595571)
	return res
}

func pairwiseOp(left []string, right []string, op string) []string {
	__antithesis_instrumentation__.Notify(595573)
	if len(left) != len(right) {
		__antithesis_instrumentation__.Notify(595576)
		panic(errors.AssertionFailedf("slice length mismatch (%d vs %d)", len(left), len(right)))
	} else {
		__antithesis_instrumentation__.Notify(595577)
	}
	__antithesis_instrumentation__.Notify(595574)
	res := make([]string, len(left))
	for i := range res {
		__antithesis_instrumentation__.Notify(595578)
		res[i] = fmt.Sprintf("%s %s %s", left[i], op, right[i])
	}
	__antithesis_instrumentation__.Notify(595575)
	return res
}

func createIndexCheckOperations(
	indexNames tree.NameList,
	tableDesc catalog.TableDescriptor,
	tableName *tree.TableName,
	asOf hlc.Timestamp,
) (results []checkOperation, err error) {
	__antithesis_instrumentation__.Notify(595579)
	if indexNames == nil {
		__antithesis_instrumentation__.Notify(595584)

		for _, idx := range tableDesc.PublicNonPrimaryIndexes() {
			__antithesis_instrumentation__.Notify(595586)
			results = append(results, newIndexCheckOperation(
				tableName,
				tableDesc,
				idx,
				asOf,
			))
		}
		__antithesis_instrumentation__.Notify(595585)
		return results, nil
	} else {
		__antithesis_instrumentation__.Notify(595587)
	}
	__antithesis_instrumentation__.Notify(595580)

	names := make(map[string]struct{})
	for _, idxName := range indexNames {
		__antithesis_instrumentation__.Notify(595588)
		names[idxName.String()] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(595581)
	for _, idx := range tableDesc.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(595589)
		if _, ok := names[idx.GetName()]; ok {
			__antithesis_instrumentation__.Notify(595590)
			results = append(results, newIndexCheckOperation(
				tableName,
				tableDesc,
				idx,
				asOf,
			))
			delete(names, idx.GetName())
		} else {
			__antithesis_instrumentation__.Notify(595591)
		}
	}
	__antithesis_instrumentation__.Notify(595582)
	if len(names) > 0 {
		__antithesis_instrumentation__.Notify(595592)

		missingIndexNames := []string(nil)
		for _, idxName := range indexNames {
			__antithesis_instrumentation__.Notify(595594)
			if _, ok := names[idxName.String()]; ok {
				__antithesis_instrumentation__.Notify(595595)
				missingIndexNames = append(missingIndexNames, idxName.String())
			} else {
				__antithesis_instrumentation__.Notify(595596)
			}
		}
		__antithesis_instrumentation__.Notify(595593)
		return nil, pgerror.Newf(pgcode.UndefinedObject,
			"specified indexes to check that do not exist on table %q: %v",
			tableDesc.GetName(), strings.Join(missingIndexNames, ", "))
	} else {
		__antithesis_instrumentation__.Notify(595597)
	}
	__antithesis_instrumentation__.Notify(595583)
	return results, nil
}

func createConstraintCheckOperations(
	ctx context.Context,
	p *planner,
	constraintNames tree.NameList,
	tableDesc catalog.TableDescriptor,
	tableName *tree.TableName,
	asOf hlc.Timestamp,
) (results []checkOperation, err error) {
	__antithesis_instrumentation__.Notify(595598)
	constraints, err := tableDesc.GetConstraintInfoWithLookup(func(id descpb.ID) (catalog.TableDescriptor, error) {
		__antithesis_instrumentation__.Notify(595603)
		return p.Descriptors().GetImmutableTableByID(ctx, p.Txn(), id, tree.ObjectLookupFlagsWithRequired())
	})
	__antithesis_instrumentation__.Notify(595599)
	if err != nil {
		__antithesis_instrumentation__.Notify(595604)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(595605)
	}
	__antithesis_instrumentation__.Notify(595600)

	if constraintNames != nil {
		__antithesis_instrumentation__.Notify(595606)
		wantedConstraints := make(map[string]descpb.ConstraintDetail)
		for _, constraintName := range constraintNames {
			__antithesis_instrumentation__.Notify(595608)
			if v, ok := constraints[string(constraintName)]; ok {
				__antithesis_instrumentation__.Notify(595609)
				wantedConstraints[string(constraintName)] = v
			} else {
				__antithesis_instrumentation__.Notify(595610)
				return nil, pgerror.Newf(pgcode.UndefinedObject,
					"constraint %q of relation %q does not exist", constraintName, tableDesc.GetName())
			}
		}
		__antithesis_instrumentation__.Notify(595607)
		constraints = wantedConstraints
	} else {
		__antithesis_instrumentation__.Notify(595611)
	}
	__antithesis_instrumentation__.Notify(595601)

	for _, constraint := range constraints {
		__antithesis_instrumentation__.Notify(595612)
		switch constraint.Kind {
		case descpb.ConstraintTypeCheck:
			__antithesis_instrumentation__.Notify(595613)
			results = append(results, newSQLCheckConstraintCheckOperation(
				tableName,
				tableDesc,
				constraint.CheckConstraint,
				asOf,
			))
		case descpb.ConstraintTypeFK:
			__antithesis_instrumentation__.Notify(595614)
			results = append(results, newSQLForeignKeyCheckOperation(
				tableName,
				tableDesc,
				constraint,
				asOf,
			))
		default:
			__antithesis_instrumentation__.Notify(595615)
		}
	}
	__antithesis_instrumentation__.Notify(595602)
	return results, nil
}
