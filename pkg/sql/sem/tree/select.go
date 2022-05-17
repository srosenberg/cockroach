package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treewindow"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type SelectStatement interface {
	Statement
	selectStatement()
}

func (*ParenSelect) selectStatement()  { __antithesis_instrumentation__.Notify(613068) }
func (*SelectClause) selectStatement() { __antithesis_instrumentation__.Notify(613069) }
func (*UnionClause) selectStatement()  { __antithesis_instrumentation__.Notify(613070) }
func (*ValuesClause) selectStatement() { __antithesis_instrumentation__.Notify(613071) }

type Select struct {
	With    *With
	Select  SelectStatement
	OrderBy OrderBy
	Limit   *Limit
	Locking LockingClause
}

func (node *Select) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613072)
	ctx.FormatNode(node.With)
	ctx.FormatNode(node.Select)
	if len(node.OrderBy) > 0 {
		__antithesis_instrumentation__.Notify(613075)
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.OrderBy)
	} else {
		__antithesis_instrumentation__.Notify(613076)
	}
	__antithesis_instrumentation__.Notify(613073)
	if node.Limit != nil {
		__antithesis_instrumentation__.Notify(613077)
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Limit)
	} else {
		__antithesis_instrumentation__.Notify(613078)
	}
	__antithesis_instrumentation__.Notify(613074)
	ctx.FormatNode(&node.Locking)
}

type ParenSelect struct {
	Select *Select
}

func (node *ParenSelect) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613079)
	ctx.WriteByte('(')
	ctx.FormatNode(node.Select)
	ctx.WriteByte(')')
}

type SelectClause struct {
	From        From
	DistinctOn  DistinctOn
	Exprs       SelectExprs
	GroupBy     GroupBy
	Window      Window
	Having      *Where
	Where       *Where
	Distinct    bool
	TableSelect bool
}

func (node *SelectClause) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613080)
	f := ctx.flags
	if f.HasFlags(FmtSummary) {
		__antithesis_instrumentation__.Notify(613082)
		ctx.WriteString("SELECT")
		if len(node.From.Tables) > 0 {
			__antithesis_instrumentation__.Notify(613084)
			ctx.WriteByte(' ')
			ctx.FormatNode(&node.From)
		} else {
			__antithesis_instrumentation__.Notify(613085)
		}
		__antithesis_instrumentation__.Notify(613083)
		return
	} else {
		__antithesis_instrumentation__.Notify(613086)
	}
	__antithesis_instrumentation__.Notify(613081)
	if node.TableSelect {
		__antithesis_instrumentation__.Notify(613087)
		ctx.WriteString("TABLE ")
		ctx.FormatNode(node.From.Tables[0])
	} else {
		__antithesis_instrumentation__.Notify(613088)
		ctx.WriteString("SELECT ")
		if node.Distinct {
			__antithesis_instrumentation__.Notify(613094)
			if node.DistinctOn != nil {
				__antithesis_instrumentation__.Notify(613095)
				ctx.FormatNode(&node.DistinctOn)
				ctx.WriteByte(' ')
			} else {
				__antithesis_instrumentation__.Notify(613096)
				ctx.WriteString("DISTINCT ")
			}
		} else {
			__antithesis_instrumentation__.Notify(613097)
		}
		__antithesis_instrumentation__.Notify(613089)
		ctx.FormatNode(&node.Exprs)
		if len(node.From.Tables) > 0 {
			__antithesis_instrumentation__.Notify(613098)
			ctx.WriteByte(' ')
			ctx.FormatNode(&node.From)
		} else {
			__antithesis_instrumentation__.Notify(613099)
		}
		__antithesis_instrumentation__.Notify(613090)
		if node.Where != nil {
			__antithesis_instrumentation__.Notify(613100)
			ctx.WriteByte(' ')
			ctx.FormatNode(node.Where)
		} else {
			__antithesis_instrumentation__.Notify(613101)
		}
		__antithesis_instrumentation__.Notify(613091)
		if len(node.GroupBy) > 0 {
			__antithesis_instrumentation__.Notify(613102)
			ctx.WriteByte(' ')
			ctx.FormatNode(&node.GroupBy)
		} else {
			__antithesis_instrumentation__.Notify(613103)
		}
		__antithesis_instrumentation__.Notify(613092)
		if node.Having != nil {
			__antithesis_instrumentation__.Notify(613104)
			ctx.WriteByte(' ')
			ctx.FormatNode(node.Having)
		} else {
			__antithesis_instrumentation__.Notify(613105)
		}
		__antithesis_instrumentation__.Notify(613093)
		if len(node.Window) > 0 {
			__antithesis_instrumentation__.Notify(613106)
			ctx.WriteByte(' ')
			ctx.FormatNode(&node.Window)
		} else {
			__antithesis_instrumentation__.Notify(613107)
		}
	}
}

type SelectExprs []SelectExpr

func (node *SelectExprs) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613108)
	for i := range *node {
		__antithesis_instrumentation__.Notify(613109)
		if i > 0 {
			__antithesis_instrumentation__.Notify(613111)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(613112)
		}
		__antithesis_instrumentation__.Notify(613110)
		ctx.FormatNode(&(*node)[i])
	}
}

type SelectExpr struct {
	Expr Expr
	As   UnrestrictedName
}

func (node *SelectExpr) NormalizeTopLevelVarName() error {
	__antithesis_instrumentation__.Notify(613113)
	if vBase, ok := node.Expr.(VarName); ok {
		__antithesis_instrumentation__.Notify(613115)
		v, err := vBase.NormalizeVarName()
		if err != nil {
			__antithesis_instrumentation__.Notify(613117)
			return err
		} else {
			__antithesis_instrumentation__.Notify(613118)
		}
		__antithesis_instrumentation__.Notify(613116)
		node.Expr = v
	} else {
		__antithesis_instrumentation__.Notify(613119)
	}
	__antithesis_instrumentation__.Notify(613114)
	return nil
}

func StarSelectExpr() SelectExpr {
	__antithesis_instrumentation__.Notify(613120)
	return SelectExpr{Expr: StarExpr()}
}

func (node *SelectExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613121)
	ctx.FormatNode(node.Expr)
	if node.As != "" {
		__antithesis_instrumentation__.Notify(613122)
		ctx.WriteString(" AS ")
		ctx.FormatNode(&node.As)
	} else {
		__antithesis_instrumentation__.Notify(613123)
	}
}

type AliasClause struct {
	Alias Name
	Cols  NameList
}

func (a *AliasClause) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613124)
	ctx.FormatNode(&a.Alias)
	if len(a.Cols) != 0 {
		__antithesis_instrumentation__.Notify(613125)

		ctx.WriteString(" (")
		ctx.FormatNode(&a.Cols)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(613126)
	}
}

type AsOfClause struct {
	Expr Expr
}

func (a *AsOfClause) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613127)
	ctx.WriteString("AS OF SYSTEM TIME ")
	ctx.FormatNode(a.Expr)
}

type From struct {
	Tables TableExprs
	AsOf   AsOfClause
}

func (node *From) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613128)
	ctx.WriteString("FROM ")
	ctx.FormatNode(&node.Tables)
	if node.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(613129)
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.AsOf)
	} else {
		__antithesis_instrumentation__.Notify(613130)
	}
}

type TableExprs []TableExpr

func (node *TableExprs) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613131)
	prefix := ""
	for _, n := range *node {
		__antithesis_instrumentation__.Notify(613132)
		ctx.WriteString(prefix)
		ctx.FormatNode(n)
		prefix = ", "
	}
}

type TableExpr interface {
	NodeFormatter
	tableExpr()
	WalkTableExpr(Visitor) TableExpr
}

func (*AliasedTableExpr) tableExpr() { __antithesis_instrumentation__.Notify(613133) }
func (*ParenTableExpr) tableExpr()   { __antithesis_instrumentation__.Notify(613134) }
func (*JoinTableExpr) tableExpr()    { __antithesis_instrumentation__.Notify(613135) }
func (*RowsFromExpr) tableExpr()     { __antithesis_instrumentation__.Notify(613136) }
func (*Subquery) tableExpr()         { __antithesis_instrumentation__.Notify(613137) }
func (*StatementSource) tableExpr()  { __antithesis_instrumentation__.Notify(613138) }

type StatementSource struct {
	Statement Statement
}

func (node *StatementSource) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613139)
	ctx.WriteByte('[')
	ctx.FormatNode(node.Statement)
	ctx.WriteByte(']')
}

type IndexID = catid.IndexID

type IndexFlags struct {
	Index   UnrestrictedName
	IndexID IndexID

	Direction Direction

	NoIndexJoin bool

	NoZigzagJoin bool

	NoFullScan bool

	IgnoreForeignKeys bool

	IgnoreUniqueWithoutIndexKeys bool

	ForceZigzag    bool
	ZigzagIndexes  []UnrestrictedName
	ZigzagIndexIDs []IndexID
}

func (ih *IndexFlags) ForceIndex() bool {
	__antithesis_instrumentation__.Notify(613140)
	return ih.Index != "" || func() bool {
		__antithesis_instrumentation__.Notify(613141)
		return ih.IndexID != 0 == true
	}() == true
}

func (ih *IndexFlags) CombineWith(other *IndexFlags) error {
	__antithesis_instrumentation__.Notify(613142)
	if ih.NoIndexJoin && func() bool {
		__antithesis_instrumentation__.Notify(613153)
		return other.NoIndexJoin == true
	}() == true {
		__antithesis_instrumentation__.Notify(613154)
		return errors.New("NO_INDEX_JOIN specified multiple times")
	} else {
		__antithesis_instrumentation__.Notify(613155)
	}
	__antithesis_instrumentation__.Notify(613143)
	if ih.NoZigzagJoin && func() bool {
		__antithesis_instrumentation__.Notify(613156)
		return other.NoZigzagJoin == true
	}() == true {
		__antithesis_instrumentation__.Notify(613157)
		return errors.New("NO_ZIGZAG_JOIN specified multiple times")
	} else {
		__antithesis_instrumentation__.Notify(613158)
	}
	__antithesis_instrumentation__.Notify(613144)
	if ih.NoFullScan && func() bool {
		__antithesis_instrumentation__.Notify(613159)
		return other.NoFullScan == true
	}() == true {
		__antithesis_instrumentation__.Notify(613160)
		return errors.New("NO_FULL_SCAN specified multiple times")
	} else {
		__antithesis_instrumentation__.Notify(613161)
	}
	__antithesis_instrumentation__.Notify(613145)
	if ih.IgnoreForeignKeys && func() bool {
		__antithesis_instrumentation__.Notify(613162)
		return other.IgnoreForeignKeys == true
	}() == true {
		__antithesis_instrumentation__.Notify(613163)
		return errors.New("IGNORE_FOREIGN_KEYS specified multiple times")
	} else {
		__antithesis_instrumentation__.Notify(613164)
	}
	__antithesis_instrumentation__.Notify(613146)
	if ih.IgnoreUniqueWithoutIndexKeys && func() bool {
		__antithesis_instrumentation__.Notify(613165)
		return other.IgnoreUniqueWithoutIndexKeys == true
	}() == true {
		__antithesis_instrumentation__.Notify(613166)
		return errors.New("IGNORE_UNIQUE_WITHOUT_INDEX_KEYS specified multiple times")
	} else {
		__antithesis_instrumentation__.Notify(613167)
	}
	__antithesis_instrumentation__.Notify(613147)
	result := *ih
	result.NoIndexJoin = ih.NoIndexJoin || func() bool {
		__antithesis_instrumentation__.Notify(613168)
		return other.NoIndexJoin == true
	}() == true
	result.NoZigzagJoin = ih.NoZigzagJoin || func() bool {
		__antithesis_instrumentation__.Notify(613169)
		return other.NoZigzagJoin == true
	}() == true
	result.NoFullScan = ih.NoFullScan || func() bool {
		__antithesis_instrumentation__.Notify(613170)
		return other.NoFullScan == true
	}() == true
	result.IgnoreForeignKeys = ih.IgnoreForeignKeys || func() bool {
		__antithesis_instrumentation__.Notify(613171)
		return other.IgnoreForeignKeys == true
	}() == true
	result.IgnoreUniqueWithoutIndexKeys = ih.IgnoreUniqueWithoutIndexKeys || func() bool {
		__antithesis_instrumentation__.Notify(613172)
		return other.IgnoreUniqueWithoutIndexKeys == true
	}() == true

	if other.Direction != 0 {
		__antithesis_instrumentation__.Notify(613173)
		if ih.Direction != 0 {
			__antithesis_instrumentation__.Notify(613175)
			return errors.New("ASC/DESC specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(613176)
		}
		__antithesis_instrumentation__.Notify(613174)
		result.Direction = other.Direction
	} else {
		__antithesis_instrumentation__.Notify(613177)
	}
	__antithesis_instrumentation__.Notify(613148)

	if other.ForceIndex() {
		__antithesis_instrumentation__.Notify(613178)
		if ih.ForceIndex() {
			__antithesis_instrumentation__.Notify(613180)
			return errors.New("FORCE_INDEX specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(613181)
		}
		__antithesis_instrumentation__.Notify(613179)
		result.Index = other.Index
		result.IndexID = other.IndexID
	} else {
		__antithesis_instrumentation__.Notify(613182)
	}
	__antithesis_instrumentation__.Notify(613149)

	if other.ForceZigzag {
		__antithesis_instrumentation__.Notify(613183)
		if ih.ForceZigzag {
			__antithesis_instrumentation__.Notify(613185)
			return errors.New("FORCE_ZIGZAG specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(613186)
		}
		__antithesis_instrumentation__.Notify(613184)
		result.ForceZigzag = true
	} else {
		__antithesis_instrumentation__.Notify(613187)
	}
	__antithesis_instrumentation__.Notify(613150)

	if len(other.ZigzagIndexes) > 0 {
		__antithesis_instrumentation__.Notify(613188)
		if result.ForceZigzag {
			__antithesis_instrumentation__.Notify(613190)
			return errors.New("FORCE_ZIGZAG hints not distinct")
		} else {
			__antithesis_instrumentation__.Notify(613191)
		}
		__antithesis_instrumentation__.Notify(613189)
		result.ZigzagIndexes = append(result.ZigzagIndexes, other.ZigzagIndexes...)
	} else {
		__antithesis_instrumentation__.Notify(613192)
	}
	__antithesis_instrumentation__.Notify(613151)

	if len(other.ZigzagIndexIDs) > 0 {
		__antithesis_instrumentation__.Notify(613193)
		if result.ForceZigzag {
			__antithesis_instrumentation__.Notify(613195)
			return errors.New("FORCE_ZIGZAG hints not distinct")
		} else {
			__antithesis_instrumentation__.Notify(613196)
		}
		__antithesis_instrumentation__.Notify(613194)
		result.ZigzagIndexIDs = append(result.ZigzagIndexIDs, other.ZigzagIndexIDs...)
	} else {
		__antithesis_instrumentation__.Notify(613197)
	}
	__antithesis_instrumentation__.Notify(613152)

	*ih = result
	return nil
}

func (ih *IndexFlags) Check() error {
	__antithesis_instrumentation__.Notify(613198)
	if ih.NoIndexJoin && func() bool {
		__antithesis_instrumentation__.Notify(613205)
		return ih.ForceIndex() == true
	}() == true {
		__antithesis_instrumentation__.Notify(613206)
		return errors.New("FORCE_INDEX cannot be specified in conjunction with NO_INDEX_JOIN")
	} else {
		__antithesis_instrumentation__.Notify(613207)
	}
	__antithesis_instrumentation__.Notify(613199)
	if ih.Direction != 0 && func() bool {
		__antithesis_instrumentation__.Notify(613208)
		return !ih.ForceIndex() == true
	}() == true {
		__antithesis_instrumentation__.Notify(613209)
		return errors.New("ASC/DESC must be specified in conjunction with an index")
	} else {
		__antithesis_instrumentation__.Notify(613210)
	}
	__antithesis_instrumentation__.Notify(613200)
	if ih.zigzagForced() && func() bool {
		__antithesis_instrumentation__.Notify(613211)
		return ih.NoIndexJoin == true
	}() == true {
		__antithesis_instrumentation__.Notify(613212)
		return errors.New("FORCE_ZIGZAG cannot be specified in conjunction with NO_INDEX_JOIN")
	} else {
		__antithesis_instrumentation__.Notify(613213)
	}
	__antithesis_instrumentation__.Notify(613201)
	if ih.zigzagForced() && func() bool {
		__antithesis_instrumentation__.Notify(613214)
		return ih.ForceIndex() == true
	}() == true {
		__antithesis_instrumentation__.Notify(613215)
		return errors.New("FORCE_ZIGZAG cannot be specified in conjunction with FORCE_INDEX")
	} else {
		__antithesis_instrumentation__.Notify(613216)
	}
	__antithesis_instrumentation__.Notify(613202)
	if ih.zigzagForced() && func() bool {
		__antithesis_instrumentation__.Notify(613217)
		return ih.NoZigzagJoin == true
	}() == true {
		__antithesis_instrumentation__.Notify(613218)
		return errors.New("FORCE_ZIGZAG cannot be specified in conjunction with NO_ZIGZAG_JOIN")
	} else {
		__antithesis_instrumentation__.Notify(613219)
	}
	__antithesis_instrumentation__.Notify(613203)
	for _, name := range ih.ZigzagIndexes {
		__antithesis_instrumentation__.Notify(613220)
		if len(string(name)) == 0 {
			__antithesis_instrumentation__.Notify(613221)
			return errors.New("FORCE_ZIGZAG index name cannot be empty string")
		} else {
			__antithesis_instrumentation__.Notify(613222)
		}
	}
	__antithesis_instrumentation__.Notify(613204)

	return nil
}

func (ih *IndexFlags) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613223)
	ctx.WriteByte('@')
	if !ih.NoIndexJoin && func() bool {
		__antithesis_instrumentation__.Notify(613224)
		return !ih.NoZigzagJoin == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(613225)
		return !ih.NoFullScan == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(613226)
		return !ih.IgnoreForeignKeys == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(613227)
		return !ih.IgnoreUniqueWithoutIndexKeys == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(613228)
		return ih.Direction == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(613229)
		return !ih.zigzagForced() == true
	}() == true {
		__antithesis_instrumentation__.Notify(613230)
		if ih.Index != "" {
			__antithesis_instrumentation__.Notify(613231)
			ctx.FormatNode(&ih.Index)
		} else {
			__antithesis_instrumentation__.Notify(613232)
			ctx.Printf("[%d]", ih.IndexID)
		}
	} else {
		__antithesis_instrumentation__.Notify(613233)
		ctx.WriteByte('{')
		var sep func()
		sep = func() {
			__antithesis_instrumentation__.Notify(613242)
			sep = func() { __antithesis_instrumentation__.Notify(613243); ctx.WriteByte(',') }
		}
		__antithesis_instrumentation__.Notify(613234)
		if ih.Index != "" || func() bool {
			__antithesis_instrumentation__.Notify(613244)
			return ih.IndexID != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(613245)
			sep()
			ctx.WriteString("FORCE_INDEX=")
			if ih.Index != "" {
				__antithesis_instrumentation__.Notify(613247)
				ctx.FormatNode(&ih.Index)
			} else {
				__antithesis_instrumentation__.Notify(613248)
				ctx.Printf("[%d]", ih.IndexID)
			}
			__antithesis_instrumentation__.Notify(613246)

			if ih.Direction != 0 {
				__antithesis_instrumentation__.Notify(613249)
				ctx.Printf(",%s", ih.Direction)
			} else {
				__antithesis_instrumentation__.Notify(613250)
			}
		} else {
			__antithesis_instrumentation__.Notify(613251)
		}
		__antithesis_instrumentation__.Notify(613235)
		if ih.NoIndexJoin {
			__antithesis_instrumentation__.Notify(613252)
			sep()
			ctx.WriteString("NO_INDEX_JOIN")
		} else {
			__antithesis_instrumentation__.Notify(613253)
		}
		__antithesis_instrumentation__.Notify(613236)

		if ih.NoZigzagJoin {
			__antithesis_instrumentation__.Notify(613254)
			sep()
			ctx.WriteString("NO_ZIGZAG_JOIN")
		} else {
			__antithesis_instrumentation__.Notify(613255)
		}
		__antithesis_instrumentation__.Notify(613237)

		if ih.NoFullScan {
			__antithesis_instrumentation__.Notify(613256)
			sep()
			ctx.WriteString("NO_FULL_SCAN")
		} else {
			__antithesis_instrumentation__.Notify(613257)
		}
		__antithesis_instrumentation__.Notify(613238)

		if ih.IgnoreForeignKeys {
			__antithesis_instrumentation__.Notify(613258)
			sep()
			ctx.WriteString("IGNORE_FOREIGN_KEYS")
		} else {
			__antithesis_instrumentation__.Notify(613259)
		}
		__antithesis_instrumentation__.Notify(613239)

		if ih.IgnoreUniqueWithoutIndexKeys {
			__antithesis_instrumentation__.Notify(613260)
			sep()
			ctx.WriteString("IGNORE_UNIQUE_WITHOUT_INDEX_KEYS")
		} else {
			__antithesis_instrumentation__.Notify(613261)
		}
		__antithesis_instrumentation__.Notify(613240)

		if ih.ForceZigzag || func() bool {
			__antithesis_instrumentation__.Notify(613262)
			return len(ih.ZigzagIndexes) > 0 == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(613263)
			return len(ih.ZigzagIndexIDs) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(613264)
			sep()
			if ih.ForceZigzag {
				__antithesis_instrumentation__.Notify(613265)
				ctx.WriteString("FORCE_ZIGZAG")
			} else {
				__antithesis_instrumentation__.Notify(613266)
				needSep := false
				for _, name := range ih.ZigzagIndexes {
					__antithesis_instrumentation__.Notify(613268)
					if needSep {
						__antithesis_instrumentation__.Notify(613270)
						sep()
					} else {
						__antithesis_instrumentation__.Notify(613271)
					}
					__antithesis_instrumentation__.Notify(613269)
					ctx.WriteString("FORCE_ZIGZAG=")
					ctx.FormatNode(&name)
					needSep = true
				}
				__antithesis_instrumentation__.Notify(613267)
				for _, id := range ih.ZigzagIndexIDs {
					__antithesis_instrumentation__.Notify(613272)
					if needSep {
						__antithesis_instrumentation__.Notify(613274)
						sep()
					} else {
						__antithesis_instrumentation__.Notify(613275)
					}
					__antithesis_instrumentation__.Notify(613273)
					ctx.WriteString("FORCE_ZIGZAG=")
					ctx.Printf("[%d]", id)
					needSep = true
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(613276)
		}
		__antithesis_instrumentation__.Notify(613241)
		ctx.WriteString("}")
	}
}

func (ih *IndexFlags) zigzagForced() bool {
	__antithesis_instrumentation__.Notify(613277)
	return ih.ForceZigzag || func() bool {
		__antithesis_instrumentation__.Notify(613278)
		return len(ih.ZigzagIndexes) > 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(613279)
		return len(ih.ZigzagIndexIDs) > 0 == true
	}() == true
}

type AliasedTableExpr struct {
	Expr       TableExpr
	IndexFlags *IndexFlags
	Ordinality bool
	Lateral    bool
	As         AliasClause
}

func (node *AliasedTableExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613280)
	if node.Lateral {
		__antithesis_instrumentation__.Notify(613284)
		ctx.WriteString("LATERAL ")
	} else {
		__antithesis_instrumentation__.Notify(613285)
	}
	__antithesis_instrumentation__.Notify(613281)
	ctx.FormatNode(node.Expr)
	if node.IndexFlags != nil {
		__antithesis_instrumentation__.Notify(613286)
		ctx.FormatNode(node.IndexFlags)
	} else {
		__antithesis_instrumentation__.Notify(613287)
	}
	__antithesis_instrumentation__.Notify(613282)
	if node.Ordinality {
		__antithesis_instrumentation__.Notify(613288)
		ctx.WriteString(" WITH ORDINALITY")
	} else {
		__antithesis_instrumentation__.Notify(613289)
	}
	__antithesis_instrumentation__.Notify(613283)
	if node.As.Alias != "" {
		__antithesis_instrumentation__.Notify(613290)
		ctx.WriteString(" AS ")
		ctx.FormatNode(&node.As)
	} else {
		__antithesis_instrumentation__.Notify(613291)
	}
}

type ParenTableExpr struct {
	Expr TableExpr
}

func (node *ParenTableExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613292)
	ctx.WriteByte('(')
	ctx.FormatNode(node.Expr)
	ctx.WriteByte(')')
}

func StripTableParens(expr TableExpr) TableExpr {
	__antithesis_instrumentation__.Notify(613293)
	if p, ok := expr.(*ParenTableExpr); ok {
		__antithesis_instrumentation__.Notify(613295)
		return StripTableParens(p.Expr)
	} else {
		__antithesis_instrumentation__.Notify(613296)
	}
	__antithesis_instrumentation__.Notify(613294)
	return expr
}

type JoinTableExpr struct {
	JoinType string
	Left     TableExpr
	Right    TableExpr
	Cond     JoinCond
	Hint     string
}

const (
	AstFull  = "FULL"
	AstLeft  = "LEFT"
	AstRight = "RIGHT"
	AstCross = "CROSS"
	AstInner = "INNER"
)

const (
	AstHash     = "HASH"
	AstLookup   = "LOOKUP"
	AstMerge    = "MERGE"
	AstInverted = "INVERTED"
)

func (node *JoinTableExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613297)
	ctx.FormatNode(node.Left)
	ctx.WriteByte(' ')
	if _, isNatural := node.Cond.(NaturalJoinCond); isNatural {
		__antithesis_instrumentation__.Notify(613298)

		ctx.FormatNode(node.Cond)
		ctx.WriteByte(' ')
		if node.JoinType != "" {
			__antithesis_instrumentation__.Notify(613300)
			ctx.WriteString(node.JoinType)
			ctx.WriteByte(' ')
			if node.Hint != "" {
				__antithesis_instrumentation__.Notify(613301)
				ctx.WriteString(node.Hint)
				ctx.WriteByte(' ')
			} else {
				__antithesis_instrumentation__.Notify(613302)
			}
		} else {
			__antithesis_instrumentation__.Notify(613303)
		}
		__antithesis_instrumentation__.Notify(613299)
		ctx.WriteString("JOIN ")
		ctx.FormatNode(node.Right)
	} else {
		__antithesis_instrumentation__.Notify(613304)

		if node.JoinType != "" {
			__antithesis_instrumentation__.Notify(613306)
			ctx.WriteString(node.JoinType)
			ctx.WriteByte(' ')
			if node.Hint != "" {
				__antithesis_instrumentation__.Notify(613307)
				ctx.WriteString(node.Hint)
				ctx.WriteByte(' ')
			} else {
				__antithesis_instrumentation__.Notify(613308)
			}
		} else {
			__antithesis_instrumentation__.Notify(613309)
		}
		__antithesis_instrumentation__.Notify(613305)
		ctx.WriteString("JOIN ")
		ctx.FormatNode(node.Right)
		if node.Cond != nil {
			__antithesis_instrumentation__.Notify(613310)
			ctx.WriteByte(' ')
			ctx.FormatNode(node.Cond)
		} else {
			__antithesis_instrumentation__.Notify(613311)
		}
	}
}

type JoinCond interface {
	NodeFormatter
	joinCond()
}

func (NaturalJoinCond) joinCond() { __antithesis_instrumentation__.Notify(613312) }
func (*OnJoinCond) joinCond()     { __antithesis_instrumentation__.Notify(613313) }
func (*UsingJoinCond) joinCond()  { __antithesis_instrumentation__.Notify(613314) }

type NaturalJoinCond struct{}

func (NaturalJoinCond) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613315)
	ctx.WriteString("NATURAL")
}

type OnJoinCond struct {
	Expr Expr
}

func (node *OnJoinCond) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613316)
	ctx.WriteString("ON ")
	ctx.FormatNode(node.Expr)
}

type UsingJoinCond struct {
	Cols NameList
}

func (node *UsingJoinCond) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613317)
	ctx.WriteString("USING (")
	ctx.FormatNode(&node.Cols)
	ctx.WriteByte(')')
}

type Where struct {
	Type string
	Expr Expr
}

const (
	AstWhere  = "WHERE"
	AstHaving = "HAVING"
)

func NewWhere(typ string, expr Expr) *Where {
	__antithesis_instrumentation__.Notify(613318)
	if expr == nil {
		__antithesis_instrumentation__.Notify(613320)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(613321)
	}
	__antithesis_instrumentation__.Notify(613319)
	return &Where{Type: typ, Expr: expr}
}

func (node *Where) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613322)
	ctx.WriteString(node.Type)
	ctx.WriteByte(' ')
	ctx.FormatNode(node.Expr)
}

type GroupBy []Expr

func (node *GroupBy) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613323)
	prefix := "GROUP BY "
	for _, n := range *node {
		__antithesis_instrumentation__.Notify(613324)
		ctx.WriteString(prefix)
		ctx.FormatNode(n)
		prefix = ", "
	}
}

type DistinctOn []Expr

func (node *DistinctOn) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613325)
	ctx.WriteString("DISTINCT ON (")
	ctx.FormatNode((*Exprs)(node))
	ctx.WriteByte(')')
}

type OrderBy []*Order

func (node *OrderBy) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613326)
	prefix := "ORDER BY "
	for _, n := range *node {
		__antithesis_instrumentation__.Notify(613327)
		ctx.WriteString(prefix)
		ctx.FormatNode(n)
		prefix = ", "
	}
}

type Direction int8

const (
	DefaultDirection Direction = iota
	Ascending
	Descending
)

var directionName = [...]string{
	DefaultDirection: "",
	Ascending:        "ASC",
	Descending:       "DESC",
}

func (d Direction) String() string {
	__antithesis_instrumentation__.Notify(613328)
	if d < 0 || func() bool {
		__antithesis_instrumentation__.Notify(613330)
		return d > Direction(len(directionName)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(613331)
		return fmt.Sprintf("Direction(%d)", d)
	} else {
		__antithesis_instrumentation__.Notify(613332)
	}
	__antithesis_instrumentation__.Notify(613329)
	return directionName[d]
}

type NullsOrder int8

const (
	DefaultNullsOrder NullsOrder = iota
	NullsFirst
	NullsLast
)

var nullsOrderName = [...]string{
	DefaultNullsOrder: "",
	NullsFirst:        "NULLS FIRST",
	NullsLast:         "NULLS LAST",
}

func (n NullsOrder) String() string {
	__antithesis_instrumentation__.Notify(613333)
	if n < 0 || func() bool {
		__antithesis_instrumentation__.Notify(613335)
		return n > NullsOrder(len(nullsOrderName)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(613336)
		return fmt.Sprintf("NullsOrder(%d)", n)
	} else {
		__antithesis_instrumentation__.Notify(613337)
	}
	__antithesis_instrumentation__.Notify(613334)
	return nullsOrderName[n]
}

type OrderType int

const (
	OrderByColumn OrderType = iota

	OrderByIndex
)

type Order struct {
	OrderType  OrderType
	Expr       Expr
	Direction  Direction
	NullsOrder NullsOrder

	Table TableName

	Index UnrestrictedName
}

func (node *Order) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613338)
	if node.OrderType == OrderByColumn {
		__antithesis_instrumentation__.Notify(613341)
		ctx.FormatNode(node.Expr)
	} else {
		__antithesis_instrumentation__.Notify(613342)
		if node.Index == "" {
			__antithesis_instrumentation__.Notify(613343)
			ctx.WriteString("PRIMARY KEY ")
			ctx.FormatNode(&node.Table)
		} else {
			__antithesis_instrumentation__.Notify(613344)
			ctx.WriteString("INDEX ")
			ctx.FormatNode(&node.Table)
			ctx.WriteByte('@')
			ctx.FormatNode(&node.Index)
		}
	}
	__antithesis_instrumentation__.Notify(613339)
	if node.Direction != DefaultDirection {
		__antithesis_instrumentation__.Notify(613345)
		ctx.WriteByte(' ')
		ctx.WriteString(node.Direction.String())
	} else {
		__antithesis_instrumentation__.Notify(613346)
	}
	__antithesis_instrumentation__.Notify(613340)
	if node.NullsOrder != DefaultNullsOrder {
		__antithesis_instrumentation__.Notify(613347)
		ctx.WriteByte(' ')
		ctx.WriteString(node.NullsOrder.String())
	} else {
		__antithesis_instrumentation__.Notify(613348)
	}
}

func (node *Order) Equal(other *Order) bool {
	__antithesis_instrumentation__.Notify(613349)
	return node.Expr.String() == other.Expr.String() && func() bool {
		__antithesis_instrumentation__.Notify(613350)
		return node.Direction == other.Direction == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(613351)
		return node.Table == other.Table == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(613352)
		return node.OrderType == other.OrderType == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(613353)
		return node.NullsOrder == other.NullsOrder == true
	}() == true
}

type Limit struct {
	Offset, Count Expr
	LimitAll      bool
}

func (node *Limit) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613354)
	needSpace := false
	if node.Count != nil {
		__antithesis_instrumentation__.Notify(613356)
		ctx.WriteString("LIMIT ")
		ctx.FormatNode(node.Count)
		needSpace = true
	} else {
		__antithesis_instrumentation__.Notify(613357)
		if node.LimitAll {
			__antithesis_instrumentation__.Notify(613358)
			ctx.WriteString("LIMIT ALL")
			needSpace = true
		} else {
			__antithesis_instrumentation__.Notify(613359)
		}
	}
	__antithesis_instrumentation__.Notify(613355)
	if node.Offset != nil {
		__antithesis_instrumentation__.Notify(613360)
		if needSpace {
			__antithesis_instrumentation__.Notify(613362)
			ctx.WriteByte(' ')
		} else {
			__antithesis_instrumentation__.Notify(613363)
		}
		__antithesis_instrumentation__.Notify(613361)
		ctx.WriteString("OFFSET ")
		ctx.FormatNode(node.Offset)
	} else {
		__antithesis_instrumentation__.Notify(613364)
	}
}

type RowsFromExpr struct {
	Items Exprs
}

func (node *RowsFromExpr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613365)
	ctx.WriteString("ROWS FROM (")
	ctx.FormatNode(&node.Items)
	ctx.WriteByte(')')
}

type Window []*WindowDef

func (node *Window) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613366)
	prefix := "WINDOW "
	for _, n := range *node {
		__antithesis_instrumentation__.Notify(613367)
		ctx.WriteString(prefix)
		ctx.FormatNode(&n.Name)
		ctx.WriteString(" AS ")
		ctx.FormatNode(n)
		prefix = ", "
	}
}

type WindowDef struct {
	Name       Name
	RefName    Name
	Partitions Exprs
	OrderBy    OrderBy
	Frame      *WindowFrame
}

func (node *WindowDef) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613368)
	ctx.WriteByte('(')
	needSpaceSeparator := false
	if node.RefName != "" {
		__antithesis_instrumentation__.Notify(613373)
		ctx.FormatNode(&node.RefName)
		needSpaceSeparator = true
	} else {
		__antithesis_instrumentation__.Notify(613374)
	}
	__antithesis_instrumentation__.Notify(613369)
	if len(node.Partitions) > 0 {
		__antithesis_instrumentation__.Notify(613375)
		if needSpaceSeparator {
			__antithesis_instrumentation__.Notify(613377)
			ctx.WriteByte(' ')
		} else {
			__antithesis_instrumentation__.Notify(613378)
		}
		__antithesis_instrumentation__.Notify(613376)
		ctx.WriteString("PARTITION BY ")
		ctx.FormatNode(&node.Partitions)
		needSpaceSeparator = true
	} else {
		__antithesis_instrumentation__.Notify(613379)
	}
	__antithesis_instrumentation__.Notify(613370)
	if len(node.OrderBy) > 0 {
		__antithesis_instrumentation__.Notify(613380)
		if needSpaceSeparator {
			__antithesis_instrumentation__.Notify(613382)
			ctx.WriteByte(' ')
		} else {
			__antithesis_instrumentation__.Notify(613383)
		}
		__antithesis_instrumentation__.Notify(613381)
		ctx.FormatNode(&node.OrderBy)
		needSpaceSeparator = true
	} else {
		__antithesis_instrumentation__.Notify(613384)
	}
	__antithesis_instrumentation__.Notify(613371)
	if node.Frame != nil {
		__antithesis_instrumentation__.Notify(613385)
		if needSpaceSeparator {
			__antithesis_instrumentation__.Notify(613387)
			ctx.WriteByte(' ')
		} else {
			__antithesis_instrumentation__.Notify(613388)
		}
		__antithesis_instrumentation__.Notify(613386)
		ctx.FormatNode(node.Frame)
	} else {
		__antithesis_instrumentation__.Notify(613389)
	}
	__antithesis_instrumentation__.Notify(613372)
	ctx.WriteRune(')')
}

func OverrideWindowDef(base *WindowDef, override WindowDef) (WindowDef, error) {
	__antithesis_instrumentation__.Notify(613390)

	if len(override.Partitions) > 0 {
		__antithesis_instrumentation__.Notify(613394)
		return WindowDef{}, pgerror.Newf(pgcode.Windowing, "cannot override PARTITION BY clause of window %q", base.Name)
	} else {
		__antithesis_instrumentation__.Notify(613395)
	}
	__antithesis_instrumentation__.Notify(613391)
	override.Partitions = base.Partitions

	if len(base.OrderBy) > 0 {
		__antithesis_instrumentation__.Notify(613396)
		if len(override.OrderBy) > 0 {
			__antithesis_instrumentation__.Notify(613398)
			return WindowDef{}, pgerror.Newf(pgcode.Windowing, "cannot override ORDER BY clause of window %q", base.Name)
		} else {
			__antithesis_instrumentation__.Notify(613399)
		}
		__antithesis_instrumentation__.Notify(613397)
		override.OrderBy = base.OrderBy
	} else {
		__antithesis_instrumentation__.Notify(613400)
	}
	__antithesis_instrumentation__.Notify(613392)

	if base.Frame != nil {
		__antithesis_instrumentation__.Notify(613401)
		return WindowDef{}, pgerror.Newf(pgcode.Windowing, "cannot copy window %q because it has a frame clause", base.Name)
	} else {
		__antithesis_instrumentation__.Notify(613402)
	}
	__antithesis_instrumentation__.Notify(613393)

	return override, nil
}

type WindowFrameBound struct {
	BoundType  treewindow.WindowFrameBoundType
	OffsetExpr Expr
}

func (node *WindowFrameBound) HasOffset() bool {
	__antithesis_instrumentation__.Notify(613403)
	return node.BoundType.IsOffset()
}

type WindowFrameBounds struct {
	StartBound *WindowFrameBound
	EndBound   *WindowFrameBound
}

func (node *WindowFrameBounds) HasOffset() bool {
	__antithesis_instrumentation__.Notify(613404)
	return node.StartBound.HasOffset() || func() bool {
		__antithesis_instrumentation__.Notify(613405)
		return (node.EndBound != nil && func() bool {
			__antithesis_instrumentation__.Notify(613406)
			return node.EndBound.HasOffset() == true
		}() == true) == true
	}() == true
}

type WindowFrame struct {
	Mode      treewindow.WindowFrameMode
	Bounds    WindowFrameBounds
	Exclusion treewindow.WindowFrameExclusion
}

func (node *WindowFrameBound) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613407)
	switch node.BoundType {
	case treewindow.UnboundedPreceding:
		__antithesis_instrumentation__.Notify(613408)
		ctx.WriteString("UNBOUNDED PRECEDING")
	case treewindow.OffsetPreceding:
		__antithesis_instrumentation__.Notify(613409)
		ctx.FormatNode(node.OffsetExpr)
		ctx.WriteString(" PRECEDING")
	case treewindow.CurrentRow:
		__antithesis_instrumentation__.Notify(613410)
		ctx.WriteString("CURRENT ROW")
	case treewindow.OffsetFollowing:
		__antithesis_instrumentation__.Notify(613411)
		ctx.FormatNode(node.OffsetExpr)
		ctx.WriteString(" FOLLOWING")
	case treewindow.UnboundedFollowing:
		__antithesis_instrumentation__.Notify(613412)
		ctx.WriteString("UNBOUNDED FOLLOWING")
	default:
		__antithesis_instrumentation__.Notify(613413)
		panic(errors.AssertionFailedf("unhandled case: %d", redact.Safe(node.BoundType)))
	}
}

func (node *WindowFrame) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613414)
	ctx.WriteString(treewindow.WindowModeName(node.Mode))
	ctx.WriteByte(' ')
	if node.Bounds.EndBound != nil {
		__antithesis_instrumentation__.Notify(613416)
		ctx.WriteString("BETWEEN ")
		ctx.FormatNode(node.Bounds.StartBound)
		ctx.WriteString(" AND ")
		ctx.FormatNode(node.Bounds.EndBound)
	} else {
		__antithesis_instrumentation__.Notify(613417)
		ctx.FormatNode(node.Bounds.StartBound)
	}
	__antithesis_instrumentation__.Notify(613415)
	if node.Exclusion != treewindow.NoExclusion {
		__antithesis_instrumentation__.Notify(613418)
		ctx.WriteByte(' ')
		ctx.WriteString(node.Exclusion.String())
	} else {
		__antithesis_instrumentation__.Notify(613419)
	}
}

type LockingClause []*LockingItem

func (node *LockingClause) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613420)
	for _, n := range *node {
		__antithesis_instrumentation__.Notify(613421)
		ctx.FormatNode(n)
	}
}

type LockingItem struct {
	Strength   LockingStrength
	Targets    TableNames
	WaitPolicy LockingWaitPolicy
}

func (f *LockingItem) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613422)
	ctx.FormatNode(f.Strength)
	if len(f.Targets) > 0 {
		__antithesis_instrumentation__.Notify(613424)
		ctx.WriteString(" OF ")
		ctx.FormatNode(&f.Targets)
	} else {
		__antithesis_instrumentation__.Notify(613425)
	}
	__antithesis_instrumentation__.Notify(613423)
	ctx.FormatNode(f.WaitPolicy)
}

type LockingStrength byte

const (
	ForNone LockingStrength = iota

	ForKeyShare

	ForShare

	ForNoKeyUpdate

	ForUpdate
)

var lockingStrengthName = [...]string{
	ForNone:        "",
	ForKeyShare:    "FOR KEY SHARE",
	ForShare:       "FOR SHARE",
	ForNoKeyUpdate: "FOR NO KEY UPDATE",
	ForUpdate:      "FOR UPDATE",
}

func (s LockingStrength) String() string {
	__antithesis_instrumentation__.Notify(613426)
	return lockingStrengthName[s]
}

func (s LockingStrength) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613427)
	if s != ForNone {
		__antithesis_instrumentation__.Notify(613428)
		ctx.WriteString(" ")
		ctx.WriteString(s.String())
	} else {
		__antithesis_instrumentation__.Notify(613429)
	}
}

func (s LockingStrength) Max(s2 LockingStrength) LockingStrength {
	__antithesis_instrumentation__.Notify(613430)
	return LockingStrength(max(byte(s), byte(s2)))
}

type LockingWaitPolicy byte

const (
	LockWaitBlock LockingWaitPolicy = iota

	LockWaitSkip

	LockWaitError
)

var lockingWaitPolicyName = [...]string{
	LockWaitBlock: "",
	LockWaitSkip:  "SKIP LOCKED",
	LockWaitError: "NOWAIT",
}

func (p LockingWaitPolicy) String() string {
	__antithesis_instrumentation__.Notify(613431)
	return lockingWaitPolicyName[p]
}

func (p LockingWaitPolicy) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613432)
	if p != LockWaitBlock {
		__antithesis_instrumentation__.Notify(613433)
		ctx.WriteString(" ")
		ctx.WriteString(p.String())
	} else {
		__antithesis_instrumentation__.Notify(613434)
	}
}

func (p LockingWaitPolicy) Max(p2 LockingWaitPolicy) LockingWaitPolicy {
	__antithesis_instrumentation__.Notify(613435)
	return LockingWaitPolicy(max(byte(p), byte(p2)))
}

func max(a, b byte) byte {
	__antithesis_instrumentation__.Notify(613436)
	if a > b {
		__antithesis_instrumentation__.Notify(613438)
		return a
	} else {
		__antithesis_instrumentation__.Notify(613439)
	}
	__antithesis_instrumentation__.Notify(613437)
	return b
}
