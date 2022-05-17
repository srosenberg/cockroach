package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"reflect"

	"github.com/cockroachdb/errors"
)

type planObserver struct {
	replaceNode func(ctx context.Context, nodeName string, plan planNode) (planNode, error)

	enterNode func(ctx context.Context, nodeName string, plan planNode) (bool, error)

	leaveNode func(nodeName string, plan planNode) error
}

func walkPlan(ctx context.Context, plan planNode, observer planObserver) error {
	__antithesis_instrumentation__.Notify(632709)
	v := makePlanVisitor(ctx, observer)
	v.visit(plan)
	return v.err
}

type planVisitor struct {
	observer planObserver
	ctx      context.Context
	err      error
}

func makePlanVisitor(ctx context.Context, observer planObserver) planVisitor {
	__antithesis_instrumentation__.Notify(632710)
	return planVisitor{observer: observer, ctx: ctx}
}

func (v *planVisitor) visit(plan planNode) planNode {
	__antithesis_instrumentation__.Notify(632711)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(632714)
		return plan
	} else {
		__antithesis_instrumentation__.Notify(632715)
	}
	__antithesis_instrumentation__.Notify(632712)

	name := nodeName(plan)

	if v.observer.replaceNode != nil {
		__antithesis_instrumentation__.Notify(632716)
		newNode, err := v.observer.replaceNode(v.ctx, name, plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(632718)
			v.err = err
			return plan
		} else {
			__antithesis_instrumentation__.Notify(632719)
		}
		__antithesis_instrumentation__.Notify(632717)
		if newNode != nil {
			__antithesis_instrumentation__.Notify(632720)
			return newNode
		} else {
			__antithesis_instrumentation__.Notify(632721)
		}
	} else {
		__antithesis_instrumentation__.Notify(632722)
	}
	__antithesis_instrumentation__.Notify(632713)
	v.visitInternal(plan, name)
	return plan
}

func (v *planVisitor) visitConcrete(plan planNode) {
	__antithesis_instrumentation__.Notify(632723)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(632725)
		return
	} else {
		__antithesis_instrumentation__.Notify(632726)
	}
	__antithesis_instrumentation__.Notify(632724)

	name := nodeName(plan)
	v.visitInternal(plan, name)
}

func (v *planVisitor) visitInternal(plan planNode, name string) {
	__antithesis_instrumentation__.Notify(632727)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(632732)
		return
	} else {
		__antithesis_instrumentation__.Notify(632733)
	}
	__antithesis_instrumentation__.Notify(632728)
	recurse := true

	if v.observer.enterNode != nil {
		__antithesis_instrumentation__.Notify(632734)
		recurse, v.err = v.observer.enterNode(v.ctx, name, plan)
		if v.err != nil {
			__antithesis_instrumentation__.Notify(632735)
			return
		} else {
			__antithesis_instrumentation__.Notify(632736)
		}
	} else {
		__antithesis_instrumentation__.Notify(632737)
	}
	__antithesis_instrumentation__.Notify(632729)
	if v.observer.leaveNode != nil {
		__antithesis_instrumentation__.Notify(632738)
		defer func() {
			__antithesis_instrumentation__.Notify(632739)
			if v.err != nil {
				__antithesis_instrumentation__.Notify(632741)
				return
			} else {
				__antithesis_instrumentation__.Notify(632742)
			}
			__antithesis_instrumentation__.Notify(632740)
			v.err = v.observer.leaveNode(name, plan)
		}()
	} else {
		__antithesis_instrumentation__.Notify(632743)
	}
	__antithesis_instrumentation__.Notify(632730)

	if !recurse {
		__antithesis_instrumentation__.Notify(632744)
		return
	} else {
		__antithesis_instrumentation__.Notify(632745)
	}
	__antithesis_instrumentation__.Notify(632731)

	switch n := plan.(type) {
	case *valuesNode:
		__antithesis_instrumentation__.Notify(632746)
	case *scanNode:
		__antithesis_instrumentation__.Notify(632747)

	case *filterNode:
		__antithesis_instrumentation__.Notify(632748)
		n.source.plan = v.visit(n.source.plan)

	case *renderNode:
		__antithesis_instrumentation__.Notify(632749)
		n.source.plan = v.visit(n.source.plan)

	case *indexJoinNode:
		__antithesis_instrumentation__.Notify(632750)
		n.input = v.visit(n.input)

	case *lookupJoinNode:
		__antithesis_instrumentation__.Notify(632751)
		n.input = v.visit(n.input)

	case *vTableLookupJoinNode:
		__antithesis_instrumentation__.Notify(632752)
		n.input = v.visit(n.input)

	case *zigzagJoinNode:
		__antithesis_instrumentation__.Notify(632753)

	case *applyJoinNode:
		__antithesis_instrumentation__.Notify(632754)
		n.input.plan = v.visit(n.input.plan)

	case *joinNode:
		__antithesis_instrumentation__.Notify(632755)
		n.left.plan = v.visit(n.left.plan)
		n.right.plan = v.visit(n.right.plan)

	case *invertedFilterNode:
		__antithesis_instrumentation__.Notify(632756)
		n.input = v.visit(n.input)

	case *invertedJoinNode:
		__antithesis_instrumentation__.Notify(632757)
		n.input = v.visit(n.input)

	case *limitNode:
		__antithesis_instrumentation__.Notify(632758)
		n.plan = v.visit(n.plan)

	case *max1RowNode:
		__antithesis_instrumentation__.Notify(632759)
		n.plan = v.visit(n.plan)

	case *distinctNode:
		__antithesis_instrumentation__.Notify(632760)
		n.plan = v.visit(n.plan)

	case *sortNode:
		__antithesis_instrumentation__.Notify(632761)
		n.plan = v.visit(n.plan)

	case *topKNode:
		__antithesis_instrumentation__.Notify(632762)
		n.plan = v.visit(n.plan)

	case *groupNode:
		__antithesis_instrumentation__.Notify(632763)
		n.plan = v.visit(n.plan)

	case *windowNode:
		__antithesis_instrumentation__.Notify(632764)
		n.plan = v.visit(n.plan)

	case *unionNode:
		__antithesis_instrumentation__.Notify(632765)
		n.left = v.visit(n.left)
		n.right = v.visit(n.right)

	case *splitNode:
		__antithesis_instrumentation__.Notify(632766)
		n.rows = v.visit(n.rows)

	case *unsplitNode:
		__antithesis_instrumentation__.Notify(632767)
		n.rows = v.visit(n.rows)

	case *relocateNode:
		__antithesis_instrumentation__.Notify(632768)
		n.rows = v.visit(n.rows)

	case *relocateRange:
		__antithesis_instrumentation__.Notify(632769)
		n.rows = v.visit(n.rows)

	case *insertNode, *insertFastPathNode:
		__antithesis_instrumentation__.Notify(632770)
		if ins, ok := n.(*insertNode); ok {
			__antithesis_instrumentation__.Notify(632804)
			ins.source = v.visit(ins.source)
		} else {
			__antithesis_instrumentation__.Notify(632805)
		}

	case *upsertNode:
		__antithesis_instrumentation__.Notify(632771)
		n.source = v.visit(n.source)

	case *updateNode:
		__antithesis_instrumentation__.Notify(632772)
		n.source = v.visit(n.source)

	case *deleteNode:
		__antithesis_instrumentation__.Notify(632773)
		n.source = v.visit(n.source)

	case *deleteRangeNode:
		__antithesis_instrumentation__.Notify(632774)

	case *serializeNode:
		__antithesis_instrumentation__.Notify(632775)
		v.visitConcrete(n.source)

	case *rowCountNode:
		__antithesis_instrumentation__.Notify(632776)
		v.visitConcrete(n.source)

	case *createTableNode:
		__antithesis_instrumentation__.Notify(632777)
		if n.n.As() {
			__antithesis_instrumentation__.Notify(632806)
			n.sourcePlan = v.visit(n.sourcePlan)
		} else {
			__antithesis_instrumentation__.Notify(632807)
		}

	case *alterTenantSetClusterSettingNode:
		__antithesis_instrumentation__.Notify(632778)
	case *createViewNode:
		__antithesis_instrumentation__.Notify(632779)
	case *setVarNode:
		__antithesis_instrumentation__.Notify(632780)
	case *setClusterSettingNode:
		__antithesis_instrumentation__.Notify(632781)
	case *resetAllNode:
		__antithesis_instrumentation__.Notify(632782)

	case *delayedNode:
		__antithesis_instrumentation__.Notify(632783)
		if n.plan != nil {
			__antithesis_instrumentation__.Notify(632808)
			n.plan = v.visit(n.plan)
		} else {
			__antithesis_instrumentation__.Notify(632809)
		}

	case *explainVecNode:
		__antithesis_instrumentation__.Notify(632784)

		if n.plan.main.planNode == nil {
			__antithesis_instrumentation__.Notify(632810)
			return
		} else {
			__antithesis_instrumentation__.Notify(632811)
		}
		__antithesis_instrumentation__.Notify(632785)
		n.plan.main.planNode = v.visit(n.plan.main.planNode)

	case *explainDDLNode:
		__antithesis_instrumentation__.Notify(632786)

		if n.plan.main.planNode == nil {
			__antithesis_instrumentation__.Notify(632812)
			return
		} else {
			__antithesis_instrumentation__.Notify(632813)
		}
		__antithesis_instrumentation__.Notify(632787)
		n.plan.main.planNode = v.visit(n.plan.main.planNode)

	case *ordinalityNode:
		__antithesis_instrumentation__.Notify(632788)
		n.source = v.visit(n.source)

	case *spoolNode:
		__antithesis_instrumentation__.Notify(632789)
		n.source = v.visit(n.source)

	case *saveTableNode:
		__antithesis_instrumentation__.Notify(632790)
		n.source = v.visit(n.source)

	case *showTraceReplicaNode:
		__antithesis_instrumentation__.Notify(632791)
		n.plan = v.visit(n.plan)

	case *cancelQueriesNode:
		__antithesis_instrumentation__.Notify(632792)
		n.rows = v.visit(n.rows)

	case *cancelSessionsNode:
		__antithesis_instrumentation__.Notify(632793)
		n.rows = v.visit(n.rows)

	case *controlJobsNode:
		__antithesis_instrumentation__.Notify(632794)
		n.rows = v.visit(n.rows)

	case *controlSchedulesNode:
		__antithesis_instrumentation__.Notify(632795)
		n.rows = v.visit(n.rows)

	case *setZoneConfigNode:
		__antithesis_instrumentation__.Notify(632796)

	case *projectSetNode:
		__antithesis_instrumentation__.Notify(632797)
		n.source = v.visit(n.source)

	case *rowSourceToPlanNode:
		__antithesis_instrumentation__.Notify(632798)

	case *errorIfRowsNode:
		__antithesis_instrumentation__.Notify(632799)
		n.plan = v.visit(n.plan)

	case *scanBufferNode:
		__antithesis_instrumentation__.Notify(632800)

	case *bufferNode:
		__antithesis_instrumentation__.Notify(632801)
		n.plan = v.visit(n.plan)

	case *recursiveCTENode:
		__antithesis_instrumentation__.Notify(632802)
		n.initial = v.visit(n.initial)

	case *exportNode:
		__antithesis_instrumentation__.Notify(632803)
		n.source = v.visit(n.source)
	}
}

func nodeName(plan planNode) string {
	__antithesis_instrumentation__.Notify(632814)

	switch n := plan.(type) {
	case *scanNode:
		__antithesis_instrumentation__.Notify(632817)
		if n.reverse {
			__antithesis_instrumentation__.Notify(632822)
			return "revscan"
		} else {
			__antithesis_instrumentation__.Notify(632823)
		}
	case *unionNode:
		__antithesis_instrumentation__.Notify(632818)
		if n.emitAll {
			__antithesis_instrumentation__.Notify(632824)
			return "append"
		} else {
			__antithesis_instrumentation__.Notify(632825)
		}

	case *joinNode:
		__antithesis_instrumentation__.Notify(632819)
		if len(n.mergeJoinOrdering) > 0 {
			__antithesis_instrumentation__.Notify(632826)
			return "merge join"
		} else {
			__antithesis_instrumentation__.Notify(632827)
		}
		__antithesis_instrumentation__.Notify(632820)
		if len(n.pred.leftEqualityIndices) == 0 {
			__antithesis_instrumentation__.Notify(632828)
			return "cross join"
		} else {
			__antithesis_instrumentation__.Notify(632829)
		}
		__antithesis_instrumentation__.Notify(632821)
		return "hash join"
	}
	__antithesis_instrumentation__.Notify(632815)

	name, ok := planNodeNames[reflect.TypeOf(plan)]
	if !ok {
		__antithesis_instrumentation__.Notify(632830)
		panic(errors.AssertionFailedf("name missing for type %T", plan))
	} else {
		__antithesis_instrumentation__.Notify(632831)
	}
	__antithesis_instrumentation__.Notify(632816)

	return name
}

var planNodeNames = map[reflect.Type]string{
	reflect.TypeOf(&alterDatabaseOwnerNode{}):           "alter database owner",
	reflect.TypeOf(&alterDatabaseAddRegionNode{}):       "alter database add region",
	reflect.TypeOf(&alterDatabasePrimaryRegionNode{}):   "alter database primary region",
	reflect.TypeOf(&alterDatabasePlacementNode{}):       "alter database placement",
	reflect.TypeOf(&alterDatabaseSurvivalGoalNode{}):    "alter database survive",
	reflect.TypeOf(&alterDatabaseDropRegionNode{}):      "alter database drop region",
	reflect.TypeOf(&alterDatabaseAddSuperRegion{}):      "alter database add super region",
	reflect.TypeOf(&alterDatabaseDropSuperRegion{}):     "alter database alter super region",
	reflect.TypeOf(&alterDatabaseAlterSuperRegion{}):    "alter database drop super region",
	reflect.TypeOf(&alterDefaultPrivilegesNode{}):       "alter default privileges",
	reflect.TypeOf(&alterIndexNode{}):                   "alter index",
	reflect.TypeOf(&alterSequenceNode{}):                "alter sequence",
	reflect.TypeOf(&alterSchemaNode{}):                  "alter schema",
	reflect.TypeOf(&alterTableNode{}):                   "alter table",
	reflect.TypeOf(&alterTableOwnerNode{}):              "alter table owner",
	reflect.TypeOf(&alterTableSetLocalityNode{}):        "alter table set locality",
	reflect.TypeOf(&alterTableSetSchemaNode{}):          "alter table set schema",
	reflect.TypeOf(&alterTenantSetClusterSettingNode{}): "alter tenant set cluster setting",
	reflect.TypeOf(&alterTypeNode{}):                    "alter type",
	reflect.TypeOf(&alterRoleNode{}):                    "alter role",
	reflect.TypeOf(&alterRoleSetNode{}):                 "alter role set var",
	reflect.TypeOf(&applyJoinNode{}):                    "apply join",
	reflect.TypeOf(&bufferNode{}):                       "buffer",
	reflect.TypeOf(&cancelQueriesNode{}):                "cancel queries",
	reflect.TypeOf(&cancelSessionsNode{}):               "cancel sessions",
	reflect.TypeOf(&changePrivilegesNode{}):             "change privileges",
	reflect.TypeOf(&commentOnColumnNode{}):              "comment on column",
	reflect.TypeOf(&commentOnConstraintNode{}):          "comment on constraint",
	reflect.TypeOf(&commentOnDatabaseNode{}):            "comment on database",
	reflect.TypeOf(&commentOnIndexNode{}):               "comment on index",
	reflect.TypeOf(&commentOnTableNode{}):               "comment on table",
	reflect.TypeOf(&commentOnSchemaNode{}):              "comment on schema",
	reflect.TypeOf(&controlJobsNode{}):                  "control jobs",
	reflect.TypeOf(&controlSchedulesNode{}):             "control schedules",
	reflect.TypeOf(&createDatabaseNode{}):               "create database",
	reflect.TypeOf(&createExtensionNode{}):              "create extension",
	reflect.TypeOf(&createIndexNode{}):                  "create index",
	reflect.TypeOf(&createSequenceNode{}):               "create sequence",
	reflect.TypeOf(&createSchemaNode{}):                 "create schema",
	reflect.TypeOf(&createStatsNode{}):                  "create statistics",
	reflect.TypeOf(&createTableNode{}):                  "create table",
	reflect.TypeOf(&createTypeNode{}):                   "create type",
	reflect.TypeOf(&CreateRoleNode{}):                   "create user/role",
	reflect.TypeOf(&createViewNode{}):                   "create view",
	reflect.TypeOf(&delayedNode{}):                      "virtual table",
	reflect.TypeOf(&deleteNode{}):                       "delete",
	reflect.TypeOf(&deleteRangeNode{}):                  "delete range",
	reflect.TypeOf(&distinctNode{}):                     "distinct",
	reflect.TypeOf(&dropDatabaseNode{}):                 "drop database",
	reflect.TypeOf(&dropIndexNode{}):                    "drop index",
	reflect.TypeOf(&dropSequenceNode{}):                 "drop sequence",
	reflect.TypeOf(&dropSchemaNode{}):                   "drop schema",
	reflect.TypeOf(&dropTableNode{}):                    "drop table",
	reflect.TypeOf(&dropTypeNode{}):                     "drop type",
	reflect.TypeOf(&DropRoleNode{}):                     "drop user/role",
	reflect.TypeOf(&dropViewNode{}):                     "drop view",
	reflect.TypeOf(&errorIfRowsNode{}):                  "error if rows",
	reflect.TypeOf(&explainPlanNode{}):                  "explain plan",
	reflect.TypeOf(&explainVecNode{}):                   "explain vectorized",
	reflect.TypeOf(&explainDDLNode{}):                   "explain ddl",
	reflect.TypeOf(&exportNode{}):                       "export",
	reflect.TypeOf(&fetchNode{}):                        "fetch",
	reflect.TypeOf(&filterNode{}):                       "filter",
	reflect.TypeOf(&GrantRoleNode{}):                    "grant role",
	reflect.TypeOf(&groupNode{}):                        "group",
	reflect.TypeOf(&hookFnNode{}):                       "plugin",
	reflect.TypeOf(&indexJoinNode{}):                    "index join",
	reflect.TypeOf(&insertNode{}):                       "insert",
	reflect.TypeOf(&insertFastPathNode{}):               "insert fast path",
	reflect.TypeOf(&invertedFilterNode{}):               "inverted filter",
	reflect.TypeOf(&invertedJoinNode{}):                 "inverted join",
	reflect.TypeOf(&joinNode{}):                         "join",
	reflect.TypeOf(&limitNode{}):                        "limit",
	reflect.TypeOf(&lookupJoinNode{}):                   "lookup join",
	reflect.TypeOf(&max1RowNode{}):                      "max1row",
	reflect.TypeOf(&ordinalityNode{}):                   "ordinality",
	reflect.TypeOf(&projectSetNode{}):                   "project set",
	reflect.TypeOf(&reassignOwnedByNode{}):              "reassign owned by",
	reflect.TypeOf(&dropOwnedByNode{}):                  "drop owned by",
	reflect.TypeOf(&recursiveCTENode{}):                 "recursive cte",
	reflect.TypeOf(&refreshMaterializedViewNode{}):      "refresh materialized view",
	reflect.TypeOf(&relocateNode{}):                     "relocate",
	reflect.TypeOf(&relocateRange{}):                    "relocate range",
	reflect.TypeOf(&renameColumnNode{}):                 "rename column",
	reflect.TypeOf(&renameDatabaseNode{}):               "rename database",
	reflect.TypeOf(&renameIndexNode{}):                  "rename index",
	reflect.TypeOf(&renameTableNode{}):                  "rename table",
	reflect.TypeOf(&reparentDatabaseNode{}):             "reparent database",
	reflect.TypeOf(&renderNode{}):                       "render",
	reflect.TypeOf(&resetAllNode{}):                     "reset all",
	reflect.TypeOf(&RevokeRoleNode{}):                   "revoke role",
	reflect.TypeOf(&rowCountNode{}):                     "count",
	reflect.TypeOf(&rowSourceToPlanNode{}):              "row source to plan node",
	reflect.TypeOf(&saveTableNode{}):                    "save table",
	reflect.TypeOf(&scanBufferNode{}):                   "scan buffer",
	reflect.TypeOf(&scanNode{}):                         "scan",
	reflect.TypeOf(&scatterNode{}):                      "scatter",
	reflect.TypeOf(&scrubNode{}):                        "scrub",
	reflect.TypeOf(&sequenceSelectNode{}):               "sequence select",
	reflect.TypeOf(&serializeNode{}):                    "run",
	reflect.TypeOf(&setClusterSettingNode{}):            "set cluster setting",
	reflect.TypeOf(&setVarNode{}):                       "set",
	reflect.TypeOf(&setZoneConfigNode{}):                "configure zone",
	reflect.TypeOf(&showFingerprintsNode{}):             "show fingerprints",
	reflect.TypeOf(&showTraceNode{}):                    "show trace for",
	reflect.TypeOf(&showTraceReplicaNode{}):             "replica trace",
	reflect.TypeOf(&showVarNode{}):                      "show",
	reflect.TypeOf(&sortNode{}):                         "sort",
	reflect.TypeOf(&splitNode{}):                        "split",
	reflect.TypeOf(&topKNode{}):                         "top-k",
	reflect.TypeOf(&unsplitNode{}):                      "unsplit",
	reflect.TypeOf(&unsplitAllNode{}):                   "unsplit all",
	reflect.TypeOf(&spoolNode{}):                        "spool",
	reflect.TypeOf(&truncateNode{}):                     "truncate",
	reflect.TypeOf(&unaryNode{}):                        "emptyrow",
	reflect.TypeOf(&unionNode{}):                        "union",
	reflect.TypeOf(&updateNode{}):                       "update",
	reflect.TypeOf(&upsertNode{}):                       "upsert",
	reflect.TypeOf(&valuesNode{}):                       "values",
	reflect.TypeOf(&virtualTableNode{}):                 "virtual table values",
	reflect.TypeOf(&vTableLookupJoinNode{}):             "virtual table lookup join",
	reflect.TypeOf(&windowNode{}):                       "window",
	reflect.TypeOf(&zeroNode{}):                         "norows",
	reflect.TypeOf(&zigzagJoinNode{}):                   "zigzag join",
	reflect.TypeOf(&schemaChangePlanNode{}):             "schema change",
}
