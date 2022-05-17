package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type tableComments struct {
	comment     *string
	columns     []comment
	indexes     []comment
	constraints []comment
}

type comment struct {
	subID   int
	comment string
}

func selectComment(ctx context.Context, p PlanHookState, tableID descpb.ID) (tc *tableComments) {
	__antithesis_instrumentation__.Notify(622856)
	query := fmt.Sprintf("SELECT type, object_id, sub_id, comment FROM system.comments WHERE object_id = %d", tableID)

	txn := p.ExtendedEvalContext().Txn
	it, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryIterator(
		ctx, "show-tables-with-comment", txn, query)
	if err != nil {
		__antithesis_instrumentation__.Notify(622858)
		log.VEventf(ctx, 1, "%q", err)
	} else {
		__antithesis_instrumentation__.Notify(622859)
		var ok bool
		for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
			__antithesis_instrumentation__.Notify(622861)
			row := it.Cur()
			commentType := keys.CommentType(tree.MustBeDInt(row[0]))
			switch commentType {
			case keys.TableCommentType, keys.ColumnCommentType,
				keys.IndexCommentType, keys.ConstraintCommentType:
				__antithesis_instrumentation__.Notify(622862)
				subID := int(tree.MustBeDInt(row[2]))
				cmt := string(tree.MustBeDString(row[3]))

				if tc == nil {
					__antithesis_instrumentation__.Notify(622865)
					tc = &tableComments{}
				} else {
					__antithesis_instrumentation__.Notify(622866)
				}
				__antithesis_instrumentation__.Notify(622863)

				switch commentType {
				case keys.TableCommentType:
					__antithesis_instrumentation__.Notify(622867)
					tc.comment = &cmt
				case keys.ColumnCommentType:
					__antithesis_instrumentation__.Notify(622868)
					tc.columns = append(tc.columns, comment{subID, cmt})
				case keys.IndexCommentType:
					__antithesis_instrumentation__.Notify(622869)
					tc.indexes = append(tc.indexes, comment{subID, cmt})
				case keys.ConstraintCommentType:
					__antithesis_instrumentation__.Notify(622870)
					tc.constraints = append(tc.constraints, comment{subID, cmt})
				default:
					__antithesis_instrumentation__.Notify(622871)
				}
			default:
				__antithesis_instrumentation__.Notify(622864)
			}
		}
		__antithesis_instrumentation__.Notify(622860)
		if err != nil {
			__antithesis_instrumentation__.Notify(622872)
			log.VEventf(ctx, 1, "%q", err)
			tc = nil
		} else {
			__antithesis_instrumentation__.Notify(622873)
		}
	}
	__antithesis_instrumentation__.Notify(622857)

	return tc
}

func ShowCreateView(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
	tn *tree.TableName,
	desc catalog.TableDescriptor,
) (string, error) {
	__antithesis_instrumentation__.Notify(622874)
	f := tree.NewFmtCtx(tree.FmtSimple)
	f.WriteString("CREATE ")
	if desc.IsTemporary() {
		__antithesis_instrumentation__.Notify(622878)
		f.WriteString("TEMP ")
	} else {
		__antithesis_instrumentation__.Notify(622879)
	}
	__antithesis_instrumentation__.Notify(622875)
	f.WriteString("VIEW ")
	f.FormatNode(tn)
	f.WriteString(" (")
	cols := desc.PublicColumns()
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(622880)
		f.WriteString("\n\t")
		name := col.GetName()
		f.FormatNameP(&name)
		if i == len(cols)-1 {
			__antithesis_instrumentation__.Notify(622881)
			f.WriteRune('\n')
		} else {
			__antithesis_instrumentation__.Notify(622882)
			f.WriteRune(',')
		}
	}
	__antithesis_instrumentation__.Notify(622876)
	f.WriteString(") AS ")

	cfg := tree.DefaultPrettyCfg()
	cfg.UseTabs = true
	cfg.LineWidth = 100 - cfg.TabWidth
	q := formatViewQueryForDisplay(ctx, semaCtx, sessionData, desc, cfg)
	for i, line := range strings.Split(q, "\n") {
		__antithesis_instrumentation__.Notify(622883)
		if i > 0 {
			__antithesis_instrumentation__.Notify(622885)
			f.WriteString("\n\t")
		} else {
			__antithesis_instrumentation__.Notify(622886)
		}
		__antithesis_instrumentation__.Notify(622884)
		f.WriteString(line)
	}
	__antithesis_instrumentation__.Notify(622877)
	return f.CloseAndGetString(), nil
}

func formatViewQueryForDisplay(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
	desc catalog.TableDescriptor,
	cfg tree.PrettyCfg,
) (query string) {
	__antithesis_instrumentation__.Notify(622887)
	defer func() {
		__antithesis_instrumentation__.Notify(622891)
		parsed, err := parser.ParseOne(query)
		if err != nil {
			__antithesis_instrumentation__.Notify(622893)
			log.Warningf(ctx, "error parsing query for view %s (%v): %+v",
				desc.GetName(), desc.GetID(), err)
			return
		} else {
			__antithesis_instrumentation__.Notify(622894)
		}
		__antithesis_instrumentation__.Notify(622892)
		query = cfg.Pretty(parsed.AST)
	}()
	__antithesis_instrumentation__.Notify(622888)

	typeReplacedViewQuery, err := formatViewQueryTypesForDisplay(ctx, semaCtx, sessionData, desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(622895)
		log.Warningf(ctx, "error deserializing user defined types for view %s (%v): %+v",
			desc.GetName(), desc.GetID(), err)
		return desc.GetViewQuery()
	} else {
		__antithesis_instrumentation__.Notify(622896)
	}
	__antithesis_instrumentation__.Notify(622889)

	sequenceReplacedViewQuery, err := formatViewQuerySequencesForDisplay(ctx, semaCtx, typeReplacedViewQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(622897)
		log.Warningf(ctx, "error converting sequence IDs to names for view %s (%v): %+v",
			desc.GetName(), desc.GetID(), err)
		return typeReplacedViewQuery
	} else {
		__antithesis_instrumentation__.Notify(622898)
	}
	__antithesis_instrumentation__.Notify(622890)

	return sequenceReplacedViewQuery
}

func formatViewQuerySequencesForDisplay(
	ctx context.Context, semaCtx *tree.SemaContext, viewQuery string,
) (string, error) {
	__antithesis_instrumentation__.Notify(622899)
	replaceFunc := func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(622903)
		newExpr, err = schemaexpr.ReplaceIDsWithFQNames(ctx, expr, semaCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(622905)
			return false, expr, err
		} else {
			__antithesis_instrumentation__.Notify(622906)
		}
		__antithesis_instrumentation__.Notify(622904)
		return false, newExpr, nil
	}
	__antithesis_instrumentation__.Notify(622900)

	stmt, err := parser.ParseOne(viewQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(622907)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(622908)
	}
	__antithesis_instrumentation__.Notify(622901)

	newStmt, err := tree.SimpleStmtVisit(stmt.AST, replaceFunc)
	if err != nil {
		__antithesis_instrumentation__.Notify(622909)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(622910)
	}
	__antithesis_instrumentation__.Notify(622902)
	return newStmt.String(), nil
}

func formatViewQueryTypesForDisplay(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
	desc catalog.TableDescriptor,
) (string, error) {
	__antithesis_instrumentation__.Notify(622911)
	replaceFunc := func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(622915)
		switch n := expr.(type) {
		case *tree.AnnotateTypeExpr, *tree.CastExpr:
			__antithesis_instrumentation__.Notify(622916)
			texpr, err := tree.TypeCheck(ctx, n, semaCtx, types.Any)
			if err != nil {
				__antithesis_instrumentation__.Notify(622922)
				return false, expr, err
			} else {
				__antithesis_instrumentation__.Notify(622923)
			}
			__antithesis_instrumentation__.Notify(622917)
			if !texpr.ResolvedType().UserDefined() {
				__antithesis_instrumentation__.Notify(622924)
				return true, expr, nil
			} else {
				__antithesis_instrumentation__.Notify(622925)
			}
			__antithesis_instrumentation__.Notify(622918)

			formattedExpr, err := schemaexpr.FormatExprForDisplay(
				ctx, desc, expr.String(), semaCtx, sessionData, tree.FmtParsable,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(622926)
				return false, expr, err
			} else {
				__antithesis_instrumentation__.Notify(622927)
			}
			__antithesis_instrumentation__.Notify(622919)
			newExpr, err = parser.ParseExpr(formattedExpr)
			if err != nil {
				__antithesis_instrumentation__.Notify(622928)
				return false, expr, err
			} else {
				__antithesis_instrumentation__.Notify(622929)
			}
			__antithesis_instrumentation__.Notify(622920)
			return false, newExpr, nil
		default:
			__antithesis_instrumentation__.Notify(622921)
			return true, expr, nil
		}
	}
	__antithesis_instrumentation__.Notify(622912)

	viewQuery := desc.GetViewQuery()
	stmt, err := parser.ParseOne(viewQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(622930)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(622931)
	}
	__antithesis_instrumentation__.Notify(622913)

	newStmt, err := tree.SimpleStmtVisit(stmt.AST, replaceFunc)
	if err != nil {
		__antithesis_instrumentation__.Notify(622932)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(622933)
	}
	__antithesis_instrumentation__.Notify(622914)
	return newStmt.String(), nil
}

func showComments(
	tn *tree.TableName, table catalog.TableDescriptor, tc *tableComments, buf *bytes.Buffer,
) error {
	__antithesis_instrumentation__.Notify(622934)
	if tc == nil {
		__antithesis_instrumentation__.Notify(622942)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(622943)
	}
	__antithesis_instrumentation__.Notify(622935)
	f := tree.NewFmtCtx(tree.FmtSimple)
	un := tn.ToUnresolvedObjectName()
	if tc.comment != nil {
		__antithesis_instrumentation__.Notify(622944)
		f.WriteString(";\n")
		f.FormatNode(&tree.CommentOnTable{
			Table:   un,
			Comment: tc.comment,
		})
	} else {
		__antithesis_instrumentation__.Notify(622945)
	}
	__antithesis_instrumentation__.Notify(622936)

	for _, columnComment := range tc.columns {
		__antithesis_instrumentation__.Notify(622946)
		col, err := table.FindColumnWithID(descpb.ColumnID(columnComment.subID))
		if err != nil {
			__antithesis_instrumentation__.Notify(622948)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622949)
		}
		__antithesis_instrumentation__.Notify(622947)

		f.WriteString(";\n")
		f.FormatNode(&tree.CommentOnColumn{
			ColumnItem: &tree.ColumnItem{
				TableName:  tn.ToUnresolvedObjectName(),
				ColumnName: tree.Name(col.GetName()),
			},
			Comment: &columnComment.comment,
		})
	}
	__antithesis_instrumentation__.Notify(622937)

	for _, indexComment := range tc.indexes {
		__antithesis_instrumentation__.Notify(622950)
		idx, err := table.FindIndexWithID(descpb.IndexID(indexComment.subID))
		if err != nil {
			__antithesis_instrumentation__.Notify(622952)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622953)
		}
		__antithesis_instrumentation__.Notify(622951)

		f.WriteString(";\n")
		f.FormatNode(&tree.CommentOnIndex{
			Index: tree.TableIndexName{
				Table: *tn,
				Index: tree.UnrestrictedName(idx.GetName()),
			},
			Comment: &indexComment.comment,
		})
	}
	__antithesis_instrumentation__.Notify(622938)

	constraints, err := table.GetConstraintInfo()
	if err != nil {
		__antithesis_instrumentation__.Notify(622954)
		return err
	} else {
		__antithesis_instrumentation__.Notify(622955)
	}
	__antithesis_instrumentation__.Notify(622939)
	constraintIDToConstraint := make(map[descpb.ConstraintID]string)
	for constraintName, constraint := range constraints {
		__antithesis_instrumentation__.Notify(622956)
		constraintIDToConstraint[constraint.ConstraintID] = constraintName
	}
	__antithesis_instrumentation__.Notify(622940)
	for _, constraintComment := range tc.constraints {
		__antithesis_instrumentation__.Notify(622957)
		f.WriteString(";\n")
		constraintName := constraintIDToConstraint[descpb.ConstraintID(constraintComment.subID)]
		f.FormatNode(&tree.CommentOnConstraint{
			Constraint: tree.Name(constraintName),
			Table:      tn.ToUnresolvedObjectName(),
			Comment:    &constraintComment.comment,
		})
	}
	__antithesis_instrumentation__.Notify(622941)

	buf.WriteString(f.CloseAndGetString())
	return nil
}

func showForeignKeyConstraint(
	buf *bytes.Buffer,
	dbPrefix string,
	originTable catalog.TableDescriptor,
	fk *descpb.ForeignKeyConstraint,
	lCtx simpleSchemaResolver,
	searchPath sessiondata.SearchPath,
) error {
	__antithesis_instrumentation__.Notify(622958)
	var refNames []string
	var originNames []string
	var fkTableName tree.TableName
	if lCtx != nil {
		__antithesis_instrumentation__.Notify(622964)
		fkTable, err := lCtx.getTableByID(fk.ReferencedTableID)
		if err != nil {
			__antithesis_instrumentation__.Notify(622968)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622969)
		}
		__antithesis_instrumentation__.Notify(622965)
		fkTableName, err = getTableNameFromTableDescriptor(lCtx, fkTable, dbPrefix)
		if err != nil {
			__antithesis_instrumentation__.Notify(622970)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622971)
		}
		__antithesis_instrumentation__.Notify(622966)
		fkTableName.ExplicitSchema = !searchPath.Contains(fkTableName.SchemaName.String())
		refNames, err = fkTable.NamesForColumnIDs(fk.ReferencedColumnIDs)
		if err != nil {
			__antithesis_instrumentation__.Notify(622972)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622973)
		}
		__antithesis_instrumentation__.Notify(622967)
		originNames, err = originTable.NamesForColumnIDs(fk.OriginColumnIDs)
		if err != nil {
			__antithesis_instrumentation__.Notify(622974)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622975)
		}
	} else {
		__antithesis_instrumentation__.Notify(622976)
		refNames = []string{"???"}
		originNames = []string{"???"}
		fkTableName = tree.MakeTableNameWithSchema(tree.Name(""), tree.PublicSchemaName, tree.Name(fmt.Sprintf("[%d as ref]", fk.ReferencedTableID)))
		fkTableName.ExplicitSchema = false
	}
	__antithesis_instrumentation__.Notify(622959)
	buf.WriteString("FOREIGN KEY (")
	formatQuoteNames(buf, originNames...)
	buf.WriteString(") REFERENCES ")
	fmtCtx := tree.NewFmtCtx(tree.FmtSimple)
	fmtCtx.FormatNode(&fkTableName)
	buf.WriteString(fmtCtx.CloseAndGetString())
	buf.WriteString("(")
	formatQuoteNames(buf, refNames...)
	buf.WriteByte(')')

	if fk.Match != descpb.ForeignKeyReference_SIMPLE {
		__antithesis_instrumentation__.Notify(622977)
		buf.WriteByte(' ')
		buf.WriteString(fk.Match.String())
	} else {
		__antithesis_instrumentation__.Notify(622978)
	}
	__antithesis_instrumentation__.Notify(622960)
	if fk.OnDelete != catpb.ForeignKeyAction_NO_ACTION {
		__antithesis_instrumentation__.Notify(622979)
		buf.WriteString(" ON DELETE ")
		buf.WriteString(fk.OnDelete.String())
	} else {
		__antithesis_instrumentation__.Notify(622980)
	}
	__antithesis_instrumentation__.Notify(622961)
	if fk.OnUpdate != catpb.ForeignKeyAction_NO_ACTION {
		__antithesis_instrumentation__.Notify(622981)
		buf.WriteString(" ON UPDATE ")
		buf.WriteString(fk.OnUpdate.String())
	} else {
		__antithesis_instrumentation__.Notify(622982)
	}
	__antithesis_instrumentation__.Notify(622962)
	if fk.Validity != descpb.ConstraintValidity_Validated {
		__antithesis_instrumentation__.Notify(622983)
		buf.WriteString(" NOT VALID")
	} else {
		__antithesis_instrumentation__.Notify(622984)
	}
	__antithesis_instrumentation__.Notify(622963)
	return nil
}

func ShowCreateSequence(
	ctx context.Context, tn *tree.TableName, desc catalog.TableDescriptor,
) (string, error) {
	__antithesis_instrumentation__.Notify(622985)
	f := tree.NewFmtCtx(tree.FmtSimple)
	f.WriteString("CREATE ")
	if desc.IsTemporary() {
		__antithesis_instrumentation__.Notify(622990)
		f.WriteString("TEMP ")
	} else {
		__antithesis_instrumentation__.Notify(622991)
	}
	__antithesis_instrumentation__.Notify(622986)
	f.WriteString("SEQUENCE ")
	f.FormatNode(tn)
	opts := desc.GetSequenceOpts()
	if opts.AsIntegerType != "" {
		__antithesis_instrumentation__.Notify(622992)
		f.Printf(" AS %s", opts.AsIntegerType)
	} else {
		__antithesis_instrumentation__.Notify(622993)
	}
	__antithesis_instrumentation__.Notify(622987)
	f.Printf(" MINVALUE %d", opts.MinValue)
	f.Printf(" MAXVALUE %d", opts.MaxValue)
	f.Printf(" INCREMENT %d", opts.Increment)
	f.Printf(" START %d", opts.Start)
	if opts.Virtual {
		__antithesis_instrumentation__.Notify(622994)
		f.Printf(" VIRTUAL")
	} else {
		__antithesis_instrumentation__.Notify(622995)
	}
	__antithesis_instrumentation__.Notify(622988)
	if opts.CacheSize > 1 {
		__antithesis_instrumentation__.Notify(622996)
		f.Printf(" CACHE %d", opts.CacheSize)
	} else {
		__antithesis_instrumentation__.Notify(622997)
	}
	__antithesis_instrumentation__.Notify(622989)
	return f.CloseAndGetString(), nil
}

func showFamilyClause(desc catalog.TableDescriptor, f *tree.FmtCtx) {
	__antithesis_instrumentation__.Notify(622998)

	families := desc.GetFamilies()
	if len(families) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(623000)
		return families[0].Name == tabledesc.FamilyPrimaryName == true
	}() == true {
		__antithesis_instrumentation__.Notify(623001)
		return
	} else {
		__antithesis_instrumentation__.Notify(623002)
	}
	__antithesis_instrumentation__.Notify(622999)
	for _, fam := range families {
		__antithesis_instrumentation__.Notify(623003)
		activeColumnNames := make([]string, 0, len(fam.ColumnNames))
		for i, colID := range fam.ColumnIDs {
			__antithesis_instrumentation__.Notify(623006)
			if col, _ := desc.FindColumnWithID(colID); col != nil && func() bool {
				__antithesis_instrumentation__.Notify(623007)
				return col.Public() == true
			}() == true {
				__antithesis_instrumentation__.Notify(623008)
				activeColumnNames = append(activeColumnNames, fam.ColumnNames[i])
			} else {
				__antithesis_instrumentation__.Notify(623009)
			}
		}
		__antithesis_instrumentation__.Notify(623004)
		if len(desc.PublicColumns()) == 0 {
			__antithesis_instrumentation__.Notify(623010)
			f.WriteString("FAMILY ")
		} else {
			__antithesis_instrumentation__.Notify(623011)
			f.WriteString(",\n\tFAMILY ")
		}
		__antithesis_instrumentation__.Notify(623005)
		formatQuoteNames(&f.Buffer, fam.Name)
		f.WriteString(" (")
		formatQuoteNames(&f.Buffer, activeColumnNames...)
		f.WriteString(")")
	}
}

func showCreateLocality(desc catalog.TableDescriptor, f *tree.FmtCtx) error {
	__antithesis_instrumentation__.Notify(623012)
	if c := desc.GetLocalityConfig(); c != nil {
		__antithesis_instrumentation__.Notify(623014)
		f.WriteString(" LOCALITY ")
		return multiregion.FormatTableLocalityConfig(c, f)
	} else {
		__antithesis_instrumentation__.Notify(623015)
	}
	__antithesis_instrumentation__.Notify(623013)
	return nil
}

func ShowCreatePartitioning(
	a *tree.DatumAlloc,
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
	idx catalog.Index,
	part catalog.Partitioning,
	buf *bytes.Buffer,
	indent int,
	colOffset int,
) error {
	__antithesis_instrumentation__.Notify(623016)
	isPrimaryKeyOfPartitionAllByTable :=
		tableDesc.IsPartitionAllBy() && func() bool {
			__antithesis_instrumentation__.Notify(623028)
			return tableDesc.GetPrimaryIndexID() == idx.GetID() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(623029)
			return colOffset == 0 == true
		}() == true

	if part.NumColumns() == 0 && func() bool {
		__antithesis_instrumentation__.Notify(623030)
		return !isPrimaryKeyOfPartitionAllByTable == true
	}() == true {
		__antithesis_instrumentation__.Notify(623031)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(623032)
	}
	__antithesis_instrumentation__.Notify(623017)

	if tableDesc.IsPartitionAllBy() && func() bool {
		__antithesis_instrumentation__.Notify(623033)
		return tableDesc.GetPrimaryIndexID() != idx.GetID() == true
	}() == true {
		__antithesis_instrumentation__.Notify(623034)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(623035)
	}
	__antithesis_instrumentation__.Notify(623018)

	if c := tableDesc.GetLocalityConfig(); c != nil {
		__antithesis_instrumentation__.Notify(623036)
		switch c.Locality.(type) {
		case *catpb.LocalityConfig_RegionalByRow_:
			__antithesis_instrumentation__.Notify(623037)
			return nil
		}
	} else {
		__antithesis_instrumentation__.Notify(623038)
	}
	__antithesis_instrumentation__.Notify(623019)

	fakePrefixDatums := make([]tree.Datum, colOffset)
	for i := range fakePrefixDatums {
		__antithesis_instrumentation__.Notify(623039)
		fakePrefixDatums[i] = tree.DNull
	}
	__antithesis_instrumentation__.Notify(623020)

	indentStr := strings.Repeat("\t", indent)
	buf.WriteString(` PARTITION `)
	if isPrimaryKeyOfPartitionAllByTable {
		__antithesis_instrumentation__.Notify(623040)
		buf.WriteString(`ALL `)
	} else {
		__antithesis_instrumentation__.Notify(623041)
	}
	__antithesis_instrumentation__.Notify(623021)
	buf.WriteString(`BY `)
	if part.NumLists() > 0 {
		__antithesis_instrumentation__.Notify(623042)
		buf.WriteString(`LIST`)
	} else {
		__antithesis_instrumentation__.Notify(623043)
		if part.NumRanges() > 0 {
			__antithesis_instrumentation__.Notify(623044)
			buf.WriteString(`RANGE`)
		} else {
			__antithesis_instrumentation__.Notify(623045)
			if isPrimaryKeyOfPartitionAllByTable {
				__antithesis_instrumentation__.Notify(623046)
				buf.WriteString(`NOTHING`)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(623047)
				return errors.Errorf(`invalid partition descriptor: %v`, part.PartitioningDesc())
			}
		}
	}
	__antithesis_instrumentation__.Notify(623022)
	buf.WriteString(` (`)
	for i := 0; i < part.NumColumns(); i++ {
		__antithesis_instrumentation__.Notify(623048)
		if i != 0 {
			__antithesis_instrumentation__.Notify(623050)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(623051)
		}
		__antithesis_instrumentation__.Notify(623049)
		buf.WriteString(idx.GetKeyColumnName(colOffset + i))
	}
	__antithesis_instrumentation__.Notify(623023)
	buf.WriteString(`) (`)
	fmtCtx := tree.NewFmtCtx(tree.FmtSimple)
	isFirst := true
	err := part.ForEachList(func(name string, values [][]byte, subPartitioning catalog.Partitioning) error {
		__antithesis_instrumentation__.Notify(623052)
		if !isFirst {
			__antithesis_instrumentation__.Notify(623055)
			buf.WriteString(`, `)
		} else {
			__antithesis_instrumentation__.Notify(623056)
		}
		__antithesis_instrumentation__.Notify(623053)
		isFirst = false
		buf.WriteString("\n")
		buf.WriteString(indentStr)
		buf.WriteString("\tPARTITION ")
		fmtCtx.FormatNameP(&name)
		_, _ = fmtCtx.Buffer.WriteTo(buf)
		buf.WriteString(` VALUES IN (`)
		for j, values := range values {
			__antithesis_instrumentation__.Notify(623057)
			if j != 0 {
				__antithesis_instrumentation__.Notify(623060)
				buf.WriteString(`, `)
			} else {
				__antithesis_instrumentation__.Notify(623061)
			}
			__antithesis_instrumentation__.Notify(623058)
			tuple, _, err := rowenc.DecodePartitionTuple(
				a, codec, tableDesc, idx, part, values, fakePrefixDatums)
			if err != nil {
				__antithesis_instrumentation__.Notify(623062)
				return err
			} else {
				__antithesis_instrumentation__.Notify(623063)
			}
			__antithesis_instrumentation__.Notify(623059)
			buf.WriteString(tuple.String())
		}
		__antithesis_instrumentation__.Notify(623054)
		buf.WriteString(`)`)
		return ShowCreatePartitioning(
			a, codec, tableDesc, idx, subPartitioning, buf, indent+1, colOffset+part.NumColumns(),
		)
	})
	__antithesis_instrumentation__.Notify(623024)
	if err != nil {
		__antithesis_instrumentation__.Notify(623064)
		return err
	} else {
		__antithesis_instrumentation__.Notify(623065)
	}
	__antithesis_instrumentation__.Notify(623025)
	isFirst = true
	err = part.ForEachRange(func(name string, from, to []byte) error {
		__antithesis_instrumentation__.Notify(623066)
		if !isFirst {
			__antithesis_instrumentation__.Notify(623070)
			buf.WriteString(`, `)
		} else {
			__antithesis_instrumentation__.Notify(623071)
		}
		__antithesis_instrumentation__.Notify(623067)
		isFirst = false
		buf.WriteString("\n")
		buf.WriteString(indentStr)
		buf.WriteString("\tPARTITION ")
		buf.WriteString(name)
		buf.WriteString(" VALUES FROM ")
		fromTuple, _, err := rowenc.DecodePartitionTuple(
			a, codec, tableDesc, idx, part, from, fakePrefixDatums)
		if err != nil {
			__antithesis_instrumentation__.Notify(623072)
			return err
		} else {
			__antithesis_instrumentation__.Notify(623073)
		}
		__antithesis_instrumentation__.Notify(623068)
		buf.WriteString(fromTuple.String())
		buf.WriteString(" TO ")
		toTuple, _, err := rowenc.DecodePartitionTuple(
			a, codec, tableDesc, idx, part, to, fakePrefixDatums)
		if err != nil {
			__antithesis_instrumentation__.Notify(623074)
			return err
		} else {
			__antithesis_instrumentation__.Notify(623075)
		}
		__antithesis_instrumentation__.Notify(623069)
		buf.WriteString(toTuple.String())
		return nil
	})
	__antithesis_instrumentation__.Notify(623026)
	if err != nil {
		__antithesis_instrumentation__.Notify(623076)
		return err
	} else {
		__antithesis_instrumentation__.Notify(623077)
	}
	__antithesis_instrumentation__.Notify(623027)
	buf.WriteString("\n")
	buf.WriteString(indentStr)
	buf.WriteString(")")
	return nil
}

func showConstraintClause(
	ctx context.Context,
	desc catalog.TableDescriptor,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
	f *tree.FmtCtx,
) error {
	__antithesis_instrumentation__.Notify(623078)
	for _, e := range desc.AllActiveAndInactiveChecks() {
		__antithesis_instrumentation__.Notify(623081)
		if e.Hidden {
			__antithesis_instrumentation__.Notify(623085)
			continue
		} else {
			__antithesis_instrumentation__.Notify(623086)
		}
		__antithesis_instrumentation__.Notify(623082)
		f.WriteString(",\n\t")
		if len(e.Name) > 0 {
			__antithesis_instrumentation__.Notify(623087)
			f.WriteString("CONSTRAINT ")
			formatQuoteNames(&f.Buffer, e.Name)
			f.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(623088)
		}
		__antithesis_instrumentation__.Notify(623083)
		f.WriteString("CHECK (")
		expr, err := schemaexpr.FormatExprForDisplay(ctx, desc, e.Expr, semaCtx, sessionData, tree.FmtParsable)
		if err != nil {
			__antithesis_instrumentation__.Notify(623089)
			return err
		} else {
			__antithesis_instrumentation__.Notify(623090)
		}
		__antithesis_instrumentation__.Notify(623084)
		f.WriteString(expr)
		f.WriteString(")")
		if e.Validity != descpb.ConstraintValidity_Validated {
			__antithesis_instrumentation__.Notify(623091)
			f.WriteString(" NOT VALID")
		} else {
			__antithesis_instrumentation__.Notify(623092)
		}
	}
	__antithesis_instrumentation__.Notify(623079)
	for _, c := range desc.AllActiveAndInactiveUniqueWithoutIndexConstraints() {
		__antithesis_instrumentation__.Notify(623093)
		f.WriteString(",\n\t")
		if len(c.Name) > 0 {
			__antithesis_instrumentation__.Notify(623097)
			f.WriteString("CONSTRAINT ")
			formatQuoteNames(&f.Buffer, c.Name)
			f.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(623098)
		}
		__antithesis_instrumentation__.Notify(623094)
		f.WriteString("UNIQUE WITHOUT INDEX (")
		colNames, err := desc.NamesForColumnIDs(c.ColumnIDs)
		if err != nil {
			__antithesis_instrumentation__.Notify(623099)
			return err
		} else {
			__antithesis_instrumentation__.Notify(623100)
		}
		__antithesis_instrumentation__.Notify(623095)
		f.WriteString(strings.Join(colNames, ", "))
		f.WriteString(")")
		if c.IsPartial() {
			__antithesis_instrumentation__.Notify(623101)
			f.WriteString(" WHERE ")
			pred, err := schemaexpr.FormatExprForDisplay(ctx, desc, c.Predicate, semaCtx, sessionData, tree.FmtParsable)
			if err != nil {
				__antithesis_instrumentation__.Notify(623103)
				return err
			} else {
				__antithesis_instrumentation__.Notify(623104)
			}
			__antithesis_instrumentation__.Notify(623102)
			f.WriteString(pred)
		} else {
			__antithesis_instrumentation__.Notify(623105)
		}
		__antithesis_instrumentation__.Notify(623096)
		if c.Validity != descpb.ConstraintValidity_Validated {
			__antithesis_instrumentation__.Notify(623106)
			f.WriteString(" NOT VALID")
		} else {
			__antithesis_instrumentation__.Notify(623107)
		}
	}
	__antithesis_instrumentation__.Notify(623080)
	f.WriteString("\n)")
	return nil
}
