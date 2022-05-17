package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type VarName interface {
	TypedExpr

	NormalizeVarName() (VarName, error)
}

var _ VarName = &UnresolvedName{}
var _ VarName = UnqualifiedStar{}
var _ VarName = &AllColumnsSelector{}
var _ VarName = &TupleStar{}
var _ VarName = &ColumnItem{}

type UnqualifiedStar struct{}

func (UnqualifiedStar) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(615898)
	ctx.WriteByte('*')
}
func (u UnqualifiedStar) String() string {
	__antithesis_instrumentation__.Notify(615899)
	return AsString(u)
}

func (u UnqualifiedStar) NormalizeVarName() (VarName, error) {
	__antithesis_instrumentation__.Notify(615900)
	return u, nil
}

var singletonStarName VarName = UnqualifiedStar{}

func StarExpr() VarName { __antithesis_instrumentation__.Notify(615901); return singletonStarName }

func (UnqualifiedStar) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(615902)
	panic(errors.AssertionFailedf("unqualified stars ought to be replaced before this point"))
}

func (UnqualifiedStar) Variable() { __antithesis_instrumentation__.Notify(615903) }

func (*UnresolvedName) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(615904)
	panic(errors.AssertionFailedf("unresolved names ought to be replaced before this point"))
}

func (*UnresolvedName) Variable() { __antithesis_instrumentation__.Notify(615905) }

func (n *UnresolvedName) NormalizeVarName() (VarName, error) {
	__antithesis_instrumentation__.Notify(615906)
	return classifyColumnItem(n)
}

type AllColumnsSelector struct {
	TableName *UnresolvedObjectName
}

func (a *AllColumnsSelector) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(615907)
	ctx.FormatNode(a.TableName)
	ctx.WriteString(".*")
}
func (a *AllColumnsSelector) String() string {
	__antithesis_instrumentation__.Notify(615908)
	return AsString(a)
}

func (a *AllColumnsSelector) NormalizeVarName() (VarName, error) {
	__antithesis_instrumentation__.Notify(615909)
	return a, nil
}

func (a *AllColumnsSelector) Variable() { __antithesis_instrumentation__.Notify(615910) }

func (*AllColumnsSelector) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(615911)
	panic(errors.AssertionFailedf("all-columns selectors ought to be replaced before this point"))
}

type ColumnItem struct {
	TableName *UnresolvedObjectName

	ColumnName Name
}

func (c *ColumnItem) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(615912)
	if c.TableName != nil {
		__antithesis_instrumentation__.Notify(615914)
		ctx.FormatNode(c.TableName)
		ctx.WriteByte('.')
	} else {
		__antithesis_instrumentation__.Notify(615915)
	}
	__antithesis_instrumentation__.Notify(615913)
	ctx.FormatNode(&c.ColumnName)
}
func (c *ColumnItem) String() string {
	__antithesis_instrumentation__.Notify(615916)
	return AsString(c)
}

func (c *ColumnItem) NormalizeVarName() (VarName, error) {
	__antithesis_instrumentation__.Notify(615917)
	return c, nil
}

func (c *ColumnItem) Column() string {
	__antithesis_instrumentation__.Notify(615918)
	return string(c.ColumnName)
}

func (c *ColumnItem) Variable() { __antithesis_instrumentation__.Notify(615919) }

func (c *ColumnItem) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(615920)
	if presetTypesForTesting == nil {
		__antithesis_instrumentation__.Notify(615922)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(615923)
	}
	__antithesis_instrumentation__.Notify(615921)
	return presetTypesForTesting[c.String()]
}

func NewColumnItem(tn *TableName, colName Name) *ColumnItem {
	__antithesis_instrumentation__.Notify(615924)
	c := MakeColumnItem(tn, colName)
	return &c
}

func MakeColumnItem(tn *TableName, colName Name) ColumnItem {
	__antithesis_instrumentation__.Notify(615925)
	c := ColumnItem{ColumnName: colName}
	if tn.Table() != "" {
		__antithesis_instrumentation__.Notify(615927)
		numParts := 1
		if tn.ExplicitCatalog {
			__antithesis_instrumentation__.Notify(615929)
			numParts = 3
		} else {
			__antithesis_instrumentation__.Notify(615930)
			if tn.ExplicitSchema {
				__antithesis_instrumentation__.Notify(615931)
				numParts = 2
			} else {
				__antithesis_instrumentation__.Notify(615932)
			}
		}
		__antithesis_instrumentation__.Notify(615928)

		c.TableName = &UnresolvedObjectName{
			NumParts: numParts,
			Parts:    [3]string{tn.Table(), tn.Schema(), tn.Catalog()},
		}
	} else {
		__antithesis_instrumentation__.Notify(615933)
	}
	__antithesis_instrumentation__.Notify(615926)
	return c
}
