package scbuild

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild/internal/scbuildstmt"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

var _ scbuildstmt.TreeAnnotator = (*astAnnotator)(nil)

type astAnnotator struct {
	statement        tree.Statement
	annotation       tree.Annotations
	nonExistentNames map[*tree.TableName]struct{}
}

func newAstAnnotator(original tree.Statement) (*astAnnotator, error) {
	__antithesis_instrumentation__.Notify(579289)

	statement, err := parser.ParseOne(original.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(579291)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(579292)
	}
	__antithesis_instrumentation__.Notify(579290)
	return &astAnnotator{
		nonExistentNames: map[*tree.TableName]struct{}{},
		statement:        statement.AST,
		annotation:       tree.MakeAnnotations(statement.NumAnnotations),
	}, nil
}

func (ann *astAnnotator) GetStatement() tree.Statement {
	__antithesis_instrumentation__.Notify(579293)
	return ann.statement
}

func (ann *astAnnotator) GetAnnotations() *tree.Annotations {
	__antithesis_instrumentation__.Notify(579294)

	return &ann.annotation
}

func (ann *astAnnotator) MarkNameAsNonExistent(name *tree.TableName) {
	__antithesis_instrumentation__.Notify(579295)
	ann.nonExistentNames[name] = struct{}{}
}

func (ann *astAnnotator) SetUnresolvedNameAnnotation(
	unresolvedName *tree.UnresolvedObjectName, annotation interface{},
) {
	__antithesis_instrumentation__.Notify(579296)
	unresolvedName.SetAnnotation(&ann.annotation, annotation)
}

func (ann *astAnnotator) ValidateAnnotations() {
	__antithesis_instrumentation__.Notify(579297)

	f := tree.NewFmtCtx(
		tree.FmtAlwaysQualifyTableNames|tree.FmtMarkRedactionNode,
		tree.FmtAnnotations(&ann.annotation),
		tree.FmtReformatTableNames(func(ctx *tree.FmtCtx, name *tree.TableName) {
			__antithesis_instrumentation__.Notify(579299)

			if _, ok := ann.nonExistentNames[name]; ok {
				__antithesis_instrumentation__.Notify(579301)
				return
			} else {
				__antithesis_instrumentation__.Notify(579302)
			}
			__antithesis_instrumentation__.Notify(579300)
			if name.CatalogName == "" || func() bool {
				__antithesis_instrumentation__.Notify(579303)
				return name.SchemaName == "" == true
			}() == true {
				__antithesis_instrumentation__.Notify(579304)
				panic(errors.AssertionFailedf("unresolved name inside annotated AST "+
					"(%v)", name.String()))
			} else {
				__antithesis_instrumentation__.Notify(579305)
			}
		}))
	__antithesis_instrumentation__.Notify(579298)
	f.FormatNode(ann.statement)
}
