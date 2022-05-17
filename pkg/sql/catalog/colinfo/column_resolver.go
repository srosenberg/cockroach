package colinfo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func ProcessTargetColumns(
	tableDesc catalog.TableDescriptor, nameList tree.NameList, ensureColumns, allowMutations bool,
) ([]catalog.Column, error) {
	__antithesis_instrumentation__.Notify(251035)
	if len(nameList) == 0 {
		__antithesis_instrumentation__.Notify(251038)
		if ensureColumns {
			__antithesis_instrumentation__.Notify(251040)

			return tableDesc.VisibleColumns(), nil
		} else {
			__antithesis_instrumentation__.Notify(251041)
		}
		__antithesis_instrumentation__.Notify(251039)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(251042)
	}
	__antithesis_instrumentation__.Notify(251036)

	var colIDSet catalog.TableColSet
	cols := make([]catalog.Column, len(nameList))
	for i, colName := range nameList {
		__antithesis_instrumentation__.Notify(251043)
		col, err := tableDesc.FindColumnWithName(colName)
		if err != nil {
			__antithesis_instrumentation__.Notify(251047)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(251048)
		}
		__antithesis_instrumentation__.Notify(251044)
		if !allowMutations && func() bool {
			__antithesis_instrumentation__.Notify(251049)
			return !col.Public() == true
		}() == true {
			__antithesis_instrumentation__.Notify(251050)
			return nil, NewUndefinedColumnError(string(colName))
		} else {
			__antithesis_instrumentation__.Notify(251051)
		}
		__antithesis_instrumentation__.Notify(251045)

		if colIDSet.Contains(col.GetID()) {
			__antithesis_instrumentation__.Notify(251052)
			return nil, pgerror.Newf(pgcode.Syntax,
				"multiple assignments to the same column %q", &nameList[i])
		} else {
			__antithesis_instrumentation__.Notify(251053)
		}
		__antithesis_instrumentation__.Notify(251046)
		colIDSet.Add(col.GetID())
		cols[i] = col
	}
	__antithesis_instrumentation__.Notify(251037)

	return cols, nil
}

func sourceNameMatches(srcName *tree.TableName, toFind tree.TableName) bool {
	__antithesis_instrumentation__.Notify(251054)
	if srcName.ObjectName != toFind.ObjectName {
		__antithesis_instrumentation__.Notify(251057)
		return false
	} else {
		__antithesis_instrumentation__.Notify(251058)
	}
	__antithesis_instrumentation__.Notify(251055)
	if toFind.ExplicitSchema {
		__antithesis_instrumentation__.Notify(251059)
		if !srcName.ExplicitSchema || func() bool {
			__antithesis_instrumentation__.Notify(251061)
			return srcName.SchemaName != toFind.SchemaName == true
		}() == true {
			__antithesis_instrumentation__.Notify(251062)
			return false
		} else {
			__antithesis_instrumentation__.Notify(251063)
		}
		__antithesis_instrumentation__.Notify(251060)
		if toFind.ExplicitCatalog {
			__antithesis_instrumentation__.Notify(251064)
			if !srcName.ExplicitCatalog || func() bool {
				__antithesis_instrumentation__.Notify(251065)
				return srcName.CatalogName != toFind.CatalogName == true
			}() == true {
				__antithesis_instrumentation__.Notify(251066)
				return false
			} else {
				__antithesis_instrumentation__.Notify(251067)
			}
		} else {
			__antithesis_instrumentation__.Notify(251068)
		}
	} else {
		__antithesis_instrumentation__.Notify(251069)
	}
	__antithesis_instrumentation__.Notify(251056)
	return true
}

type ColumnResolver struct {
	Source *DataSourceInfo

	ResolverState struct {
		ColIdx int
	}
}

func (r *ColumnResolver) FindSourceMatchingName(
	ctx context.Context, tn tree.TableName,
) (res NumResolutionResults, prefix *tree.TableName, srcMeta ColumnSourceMeta, err error) {
	__antithesis_instrumentation__.Notify(251070)
	if !sourceNameMatches(&r.Source.SourceAlias, tn) {
		__antithesis_instrumentation__.Notify(251072)
		return NoResults, nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(251073)
	}
	__antithesis_instrumentation__.Notify(251071)
	prefix = &r.Source.SourceAlias
	return ExactlyOne, prefix, nil, nil
}

func (r *ColumnResolver) FindSourceProvidingColumn(
	ctx context.Context, col tree.Name,
) (prefix *tree.TableName, srcMeta ColumnSourceMeta, colHint int, err error) {
	__antithesis_instrumentation__.Notify(251074)
	colIdx := tree.NoColumnIdx
	colName := string(col)

	for idx := range r.Source.SourceColumns {
		__antithesis_instrumentation__.Notify(251077)
		colIdx, err = r.findColHelper(colName, colIdx, idx)
		if err != nil {
			__antithesis_instrumentation__.Notify(251079)
			return nil, nil, -1, err
		} else {
			__antithesis_instrumentation__.Notify(251080)
		}
		__antithesis_instrumentation__.Notify(251078)
		if colIdx != tree.NoColumnIdx {
			__antithesis_instrumentation__.Notify(251081)
			prefix = &r.Source.SourceAlias
			break
		} else {
			__antithesis_instrumentation__.Notify(251082)
		}
	}
	__antithesis_instrumentation__.Notify(251075)
	if colIdx == tree.NoColumnIdx {
		__antithesis_instrumentation__.Notify(251083)
		colAlloc := col
		return nil, nil, -1, NewUndefinedColumnError(tree.ErrString(&colAlloc))
	} else {
		__antithesis_instrumentation__.Notify(251084)
	}
	__antithesis_instrumentation__.Notify(251076)
	r.ResolverState.ColIdx = colIdx
	return prefix, nil, colIdx, nil
}

func (r *ColumnResolver) Resolve(
	ctx context.Context, prefix *tree.TableName, srcMeta ColumnSourceMeta, colHint int, col tree.Name,
) (ColumnResolutionResult, error) {
	__antithesis_instrumentation__.Notify(251085)
	if colHint != -1 {
		__antithesis_instrumentation__.Notify(251089)

		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(251090)
	}
	__antithesis_instrumentation__.Notify(251086)

	colIdx := tree.NoColumnIdx
	colName := string(col)
	for idx := range r.Source.SourceColumns {
		__antithesis_instrumentation__.Notify(251091)
		var err error
		colIdx, err = r.findColHelper(colName, colIdx, idx)
		if err != nil {
			__antithesis_instrumentation__.Notify(251092)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(251093)
		}
	}
	__antithesis_instrumentation__.Notify(251087)

	if colIdx == tree.NoColumnIdx {
		__antithesis_instrumentation__.Notify(251094)
		r.ResolverState.ColIdx = tree.NoColumnIdx
		return nil, NewUndefinedColumnError(
			tree.ErrString(tree.NewColumnItem(&r.Source.SourceAlias, tree.Name(colName))))
	} else {
		__antithesis_instrumentation__.Notify(251095)
	}
	__antithesis_instrumentation__.Notify(251088)
	r.ResolverState.ColIdx = colIdx
	return nil, nil
}

func (r *ColumnResolver) findColHelper(colName string, colIdx, idx int) (int, error) {
	__antithesis_instrumentation__.Notify(251096)
	col := r.Source.SourceColumns[idx]
	if col.Name == colName {
		__antithesis_instrumentation__.Notify(251098)
		if colIdx != tree.NoColumnIdx {
			__antithesis_instrumentation__.Notify(251100)
			colString := tree.ErrString(r.Source.NodeFormatter(idx))
			var msgBuf bytes.Buffer
			name := tree.ErrString(&r.Source.SourceAlias)
			if len(name) == 0 {
				__antithesis_instrumentation__.Notify(251102)
				name = "<anonymous>"
			} else {
				__antithesis_instrumentation__.Notify(251103)
			}
			__antithesis_instrumentation__.Notify(251101)
			fmt.Fprintf(&msgBuf, "%s.%s", name, colString)
			return tree.NoColumnIdx, pgerror.Newf(pgcode.AmbiguousColumn,
				"column reference %q is ambiguous (candidates: %s)", colString, msgBuf.String())
		} else {
			__antithesis_instrumentation__.Notify(251104)
		}
		__antithesis_instrumentation__.Notify(251099)
		colIdx = idx
	} else {
		__antithesis_instrumentation__.Notify(251105)
	}
	__antithesis_instrumentation__.Notify(251097)
	return colIdx, nil
}

func NewUndefinedColumnError(name string) error {
	__antithesis_instrumentation__.Notify(251106)
	return pgerror.Newf(pgcode.UndefinedColumn, "column %q does not exist", name)
}
