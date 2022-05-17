package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/pretty"
	"github.com/cockroachdb/errors"
	"golang.org/x/text/language"
)

type CreateDatabase struct {
	IfNotExists     bool
	Name            Name
	Template        string
	Encoding        string
	Collate         string
	CType           string
	ConnectionLimit int32
	PrimaryRegion   Name
	Regions         NameList
	SurvivalGoal    SurvivalGoal
	Placement       DataPlacement
	Owner           RoleSpec
}

func (node *CreateDatabase) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604708)
	ctx.WriteString("CREATE DATABASE ")
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(604719)
		ctx.WriteString("IF NOT EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(604720)
	}
	__antithesis_instrumentation__.Notify(604709)
	ctx.FormatNode(&node.Name)
	if node.Template != "" {
		__antithesis_instrumentation__.Notify(604721)

		ctx.WriteString(" TEMPLATE = ")
		lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, node.Template, ctx.flags.EncodeFlags())
	} else {
		__antithesis_instrumentation__.Notify(604722)
	}
	__antithesis_instrumentation__.Notify(604710)
	if node.Encoding != "" {
		__antithesis_instrumentation__.Notify(604723)

		ctx.WriteString(" ENCODING = ")
		lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, node.Encoding, ctx.flags.EncodeFlags())
	} else {
		__antithesis_instrumentation__.Notify(604724)
	}
	__antithesis_instrumentation__.Notify(604711)
	if node.Collate != "" {
		__antithesis_instrumentation__.Notify(604725)

		ctx.WriteString(" LC_COLLATE = ")
		lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, node.Collate, ctx.flags.EncodeFlags())
	} else {
		__antithesis_instrumentation__.Notify(604726)
	}
	__antithesis_instrumentation__.Notify(604712)
	if node.CType != "" {
		__antithesis_instrumentation__.Notify(604727)

		ctx.WriteString(" LC_CTYPE = ")
		lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, node.CType, ctx.flags.EncodeFlags())
	} else {
		__antithesis_instrumentation__.Notify(604728)
	}
	__antithesis_instrumentation__.Notify(604713)
	if node.ConnectionLimit != -1 {
		__antithesis_instrumentation__.Notify(604729)
		ctx.WriteString(" CONNECTION LIMIT = ")
		if ctx.flags.HasFlags(FmtHideConstants) {
			__antithesis_instrumentation__.Notify(604730)
			ctx.WriteByte('0')
		} else {
			__antithesis_instrumentation__.Notify(604731)

			ctx.WriteString(strconv.Itoa(int(node.ConnectionLimit)))
		}
	} else {
		__antithesis_instrumentation__.Notify(604732)
	}
	__antithesis_instrumentation__.Notify(604714)
	if node.PrimaryRegion != "" {
		__antithesis_instrumentation__.Notify(604733)
		ctx.WriteString(" PRIMARY REGION ")
		ctx.FormatNode(&node.PrimaryRegion)
	} else {
		__antithesis_instrumentation__.Notify(604734)
	}
	__antithesis_instrumentation__.Notify(604715)
	if node.Regions != nil {
		__antithesis_instrumentation__.Notify(604735)
		ctx.WriteString(" REGIONS = ")
		ctx.FormatNode(&node.Regions)
	} else {
		__antithesis_instrumentation__.Notify(604736)
	}
	__antithesis_instrumentation__.Notify(604716)
	if node.SurvivalGoal != SurvivalGoalDefault {
		__antithesis_instrumentation__.Notify(604737)
		ctx.WriteString(" ")
		ctx.FormatNode(&node.SurvivalGoal)
	} else {
		__antithesis_instrumentation__.Notify(604738)
	}
	__antithesis_instrumentation__.Notify(604717)
	if node.Placement != DataPlacementUnspecified {
		__antithesis_instrumentation__.Notify(604739)
		ctx.WriteString(" ")
		ctx.FormatNode(&node.Placement)
	} else {
		__antithesis_instrumentation__.Notify(604740)
	}
	__antithesis_instrumentation__.Notify(604718)

	if node.Owner.Name != "" {
		__antithesis_instrumentation__.Notify(604741)
		ctx.WriteString(" OWNER = ")
		ctx.FormatNode(&node.Owner)
	} else {
		__antithesis_instrumentation__.Notify(604742)
	}
}

type IndexElem struct {
	Column Name

	Expr       Expr
	Direction  Direction
	NullsOrder NullsOrder
}

func (node *IndexElem) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604743)
	if node.Expr == nil {
		__antithesis_instrumentation__.Notify(604746)
		ctx.FormatNode(&node.Column)
	} else {
		__antithesis_instrumentation__.Notify(604747)

		_, isFunc := node.Expr.(*FuncExpr)
		if !isFunc {
			__antithesis_instrumentation__.Notify(604749)
			ctx.WriteByte('(')
		} else {
			__antithesis_instrumentation__.Notify(604750)
		}
		__antithesis_instrumentation__.Notify(604748)
		ctx.FormatNode(node.Expr)
		if !isFunc {
			__antithesis_instrumentation__.Notify(604751)
			ctx.WriteByte(')')
		} else {
			__antithesis_instrumentation__.Notify(604752)
		}
	}
	__antithesis_instrumentation__.Notify(604744)
	if node.Direction != DefaultDirection {
		__antithesis_instrumentation__.Notify(604753)
		ctx.WriteByte(' ')
		ctx.WriteString(node.Direction.String())
	} else {
		__antithesis_instrumentation__.Notify(604754)
	}
	__antithesis_instrumentation__.Notify(604745)
	if node.NullsOrder != DefaultNullsOrder {
		__antithesis_instrumentation__.Notify(604755)
		ctx.WriteByte(' ')
		ctx.WriteString(node.NullsOrder.String())
	} else {
		__antithesis_instrumentation__.Notify(604756)
	}
}

func (node *IndexElem) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(604757)
	var d pretty.Doc
	if node.Expr == nil {
		__antithesis_instrumentation__.Notify(604761)
		d = p.Doc(&node.Column)
	} else {
		__antithesis_instrumentation__.Notify(604762)

		d = p.Doc(node.Expr)
		if _, isFunc := node.Expr.(*FuncExpr); !isFunc {
			__antithesis_instrumentation__.Notify(604763)
			d = p.bracket("(", d, ")")
		} else {
			__antithesis_instrumentation__.Notify(604764)
		}
	}
	__antithesis_instrumentation__.Notify(604758)
	if node.Direction != DefaultDirection {
		__antithesis_instrumentation__.Notify(604765)
		d = pretty.ConcatSpace(d, pretty.Keyword(node.Direction.String()))
	} else {
		__antithesis_instrumentation__.Notify(604766)
	}
	__antithesis_instrumentation__.Notify(604759)
	if node.NullsOrder != DefaultNullsOrder {
		__antithesis_instrumentation__.Notify(604767)
		d = pretty.ConcatSpace(d, pretty.Keyword(node.NullsOrder.String()))
	} else {
		__antithesis_instrumentation__.Notify(604768)
	}
	__antithesis_instrumentation__.Notify(604760)
	return d
}

type IndexElemList []IndexElem

func (l *IndexElemList) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604769)
	for i := range *l {
		__antithesis_instrumentation__.Notify(604770)
		if i > 0 {
			__antithesis_instrumentation__.Notify(604772)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(604773)
		}
		__antithesis_instrumentation__.Notify(604771)
		ctx.FormatNode(&(*l)[i])
	}
}

func (l *IndexElemList) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(604774)
	if l == nil || func() bool {
		__antithesis_instrumentation__.Notify(604777)
		return len(*l) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(604778)
		return pretty.Nil
	} else {
		__antithesis_instrumentation__.Notify(604779)
	}
	__antithesis_instrumentation__.Notify(604775)
	d := make([]pretty.Doc, len(*l))
	for i := range *l {
		__antithesis_instrumentation__.Notify(604780)
		d[i] = p.Doc(&(*l)[i])
	}
	__antithesis_instrumentation__.Notify(604776)
	return p.commaSeparated(d...)
}

type CreateIndex struct {
	Name        Name
	Table       TableName
	Unique      bool
	Inverted    bool
	IfNotExists bool
	Columns     IndexElemList
	Sharded     *ShardedIndexDef

	Storing          NameList
	PartitionByIndex *PartitionByIndex
	StorageParams    StorageParams
	Predicate        Expr
	Concurrently     bool
}

func (node *CreateIndex) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604781)

	ctx.WriteString("CREATE ")
	if node.Unique {
		__antithesis_instrumentation__.Notify(604791)
		ctx.WriteString("UNIQUE ")
	} else {
		__antithesis_instrumentation__.Notify(604792)
	}
	__antithesis_instrumentation__.Notify(604782)
	if node.Inverted {
		__antithesis_instrumentation__.Notify(604793)
		ctx.WriteString("INVERTED ")
	} else {
		__antithesis_instrumentation__.Notify(604794)
	}
	__antithesis_instrumentation__.Notify(604783)
	ctx.WriteString("INDEX ")
	if node.Concurrently {
		__antithesis_instrumentation__.Notify(604795)
		ctx.WriteString("CONCURRENTLY ")
	} else {
		__antithesis_instrumentation__.Notify(604796)
	}
	__antithesis_instrumentation__.Notify(604784)
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(604797)
		ctx.WriteString("IF NOT EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(604798)
	}
	__antithesis_instrumentation__.Notify(604785)
	if node.Name != "" {
		__antithesis_instrumentation__.Notify(604799)
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	} else {
		__antithesis_instrumentation__.Notify(604800)
	}
	__antithesis_instrumentation__.Notify(604786)
	ctx.WriteString("ON ")
	ctx.FormatNode(&node.Table)

	ctx.WriteString(" (")
	ctx.FormatNode(&node.Columns)
	ctx.WriteByte(')')
	if node.Sharded != nil {
		__antithesis_instrumentation__.Notify(604801)
		ctx.FormatNode(node.Sharded)
	} else {
		__antithesis_instrumentation__.Notify(604802)
	}
	__antithesis_instrumentation__.Notify(604787)
	if len(node.Storing) > 0 {
		__antithesis_instrumentation__.Notify(604803)
		ctx.WriteString(" STORING (")
		ctx.FormatNode(&node.Storing)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(604804)
	}
	__antithesis_instrumentation__.Notify(604788)
	if node.PartitionByIndex != nil {
		__antithesis_instrumentation__.Notify(604805)
		ctx.FormatNode(node.PartitionByIndex)
	} else {
		__antithesis_instrumentation__.Notify(604806)
	}
	__antithesis_instrumentation__.Notify(604789)
	if node.StorageParams != nil {
		__antithesis_instrumentation__.Notify(604807)
		ctx.WriteString(" WITH (")
		ctx.FormatNode(&node.StorageParams)
		ctx.WriteString(")")
	} else {
		__antithesis_instrumentation__.Notify(604808)
	}
	__antithesis_instrumentation__.Notify(604790)
	if node.Predicate != nil {
		__antithesis_instrumentation__.Notify(604809)
		ctx.WriteString(" WHERE ")
		ctx.FormatNode(node.Predicate)
	} else {
		__antithesis_instrumentation__.Notify(604810)
	}
}

type CreateTypeVariety int

const (
	_ CreateTypeVariety = iota

	Enum

	Composite

	Range

	Base

	Shell

	Domain
)

type EnumValue string

func (n *EnumValue) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604811)
	f := ctx.flags
	if f.HasFlags(FmtAnonymize) {
		__antithesis_instrumentation__.Notify(604812)
		ctx.WriteByte('_')
	} else {
		__antithesis_instrumentation__.Notify(604813)
		lexbase.EncodeSQLString(&ctx.Buffer, string(*n))
	}
}

type EnumValueList []EnumValue

func (l *EnumValueList) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604814)
	for i := range *l {
		__antithesis_instrumentation__.Notify(604815)
		if i > 0 {
			__antithesis_instrumentation__.Notify(604817)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(604818)
		}
		__antithesis_instrumentation__.Notify(604816)
		ctx.FormatNode(&(*l)[i])
	}
}

type CreateType struct {
	TypeName *UnresolvedObjectName
	Variety  CreateTypeVariety

	EnumLabels EnumValueList

	IfNotExists bool
}

var _ Statement = &CreateType{}

func (node *CreateType) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604819)
	ctx.WriteString("CREATE TYPE ")
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(604821)
		ctx.WriteString("IF NOT EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(604822)
	}
	__antithesis_instrumentation__.Notify(604820)
	ctx.FormatNode(node.TypeName)
	ctx.WriteString(" ")
	switch node.Variety {
	case Enum:
		__antithesis_instrumentation__.Notify(604823)
		ctx.WriteString("AS ENUM (")
		ctx.FormatNode(&node.EnumLabels)
		ctx.WriteString(")")
	default:
		__antithesis_instrumentation__.Notify(604824)
	}
}

func (node *CreateType) String() string {
	__antithesis_instrumentation__.Notify(604825)
	return AsString(node)
}

type TableDef interface {
	NodeFormatter

	tableDef()
}

func (*ColumnTableDef) tableDef()               { __antithesis_instrumentation__.Notify(604826) }
func (*IndexTableDef) tableDef()                { __antithesis_instrumentation__.Notify(604827) }
func (*FamilyTableDef) tableDef()               { __antithesis_instrumentation__.Notify(604828) }
func (*ForeignKeyConstraintTableDef) tableDef() { __antithesis_instrumentation__.Notify(604829) }
func (*CheckConstraintTableDef) tableDef()      { __antithesis_instrumentation__.Notify(604830) }
func (*LikeTableDef) tableDef()                 { __antithesis_instrumentation__.Notify(604831) }

type TableDefs []TableDef

func (node *TableDefs) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604832)
	for i, n := range *node {
		__antithesis_instrumentation__.Notify(604833)
		if i > 0 {
			__antithesis_instrumentation__.Notify(604835)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(604836)
		}
		__antithesis_instrumentation__.Notify(604834)
		ctx.FormatNode(n)
	}
}

type Nullability int

const (
	NotNull Nullability = iota
	Null
	SilentNull
)

type GeneratedIdentityType int

const (
	GeneratedAlways GeneratedIdentityType = iota
	GeneratedByDefault
)

type ColumnTableDef struct {
	Name              Name
	Type              ResolvableTypeReference
	IsSerial          bool
	GeneratedIdentity struct {
		IsGeneratedAsIdentity   bool
		GeneratedAsIdentityType GeneratedIdentityType
		SeqOptions              SequenceOptions
	}
	Hidden   bool
	Nullable struct {
		Nullability    Nullability
		ConstraintName Name
	}
	PrimaryKey struct {
		IsPrimaryKey  bool
		Sharded       bool
		ShardBuckets  Expr
		StorageParams StorageParams
	}
	Unique struct {
		IsUnique       bool
		WithoutIndex   bool
		ConstraintName Name
	}
	DefaultExpr struct {
		Expr           Expr
		ConstraintName Name
	}
	OnUpdateExpr struct {
		Expr           Expr
		ConstraintName Name
	}
	CheckExprs []ColumnTableDefCheckExpr
	References struct {
		Table          *TableName
		Col            Name
		ConstraintName Name
		Actions        ReferenceActions
		Match          CompositeKeyMatchMethod
	}
	Computed struct {
		Computed bool
		Expr     Expr
		Virtual  bool
	}
	Family struct {
		Name        Name
		Create      bool
		IfNotExists bool
	}
}

type ColumnTableDefCheckExpr struct {
	Expr           Expr
	ConstraintName Name
}

func processCollationOnType(
	name Name, ref ResolvableTypeReference, c ColumnCollation,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(604837)

	typ, ok := GetStaticallyKnownType(ref)
	if !ok {
		__antithesis_instrumentation__.Notify(604839)
		return nil, pgerror.Newf(pgcode.DatatypeMismatch,
			"COLLATE declaration for non-string-typed column %q", name)
	} else {
		__antithesis_instrumentation__.Notify(604840)
	}
	__antithesis_instrumentation__.Notify(604838)
	switch typ.Family() {
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(604841)
		return types.MakeCollatedString(typ, string(c)), nil
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(604842)
		return nil, pgerror.Newf(pgcode.Syntax,
			"multiple COLLATE declarations for column %q", name)
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(604843)
		elemTyp, err := processCollationOnType(name, typ.ArrayContents(), c)
		if err != nil {
			__antithesis_instrumentation__.Notify(604846)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(604847)
		}
		__antithesis_instrumentation__.Notify(604844)
		return types.MakeArray(elemTyp), nil
	default:
		__antithesis_instrumentation__.Notify(604845)
		return nil, pgerror.Newf(pgcode.DatatypeMismatch,
			"COLLATE declaration for non-string-typed column %q", name)
	}
}

func NewColumnTableDef(
	name Name,
	typRef ResolvableTypeReference,
	isSerial bool,
	qualifications []NamedColumnQualification,
) (*ColumnTableDef, error) {
	__antithesis_instrumentation__.Notify(604848)
	d := &ColumnTableDef{
		Name:     name,
		Type:     typRef,
		IsSerial: isSerial,
	}
	d.Nullable.Nullability = SilentNull
	for _, c := range qualifications {
		__antithesis_instrumentation__.Notify(604850)
		switch t := c.Qualification.(type) {
		case ColumnCollation:
			__antithesis_instrumentation__.Notify(604851)
			locale := string(t)

			if locale != DefaultCollationTag {
				__antithesis_instrumentation__.Notify(604880)
				_, err := language.Parse(locale)
				if err != nil {
					__antithesis_instrumentation__.Notify(604883)
					return nil, pgerror.Wrapf(err, pgcode.Syntax, "invalid locale %s", locale)
				} else {
					__antithesis_instrumentation__.Notify(604884)
				}
				__antithesis_instrumentation__.Notify(604881)
				collatedTyp, err := processCollationOnType(name, d.Type, t)
				if err != nil {
					__antithesis_instrumentation__.Notify(604885)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(604886)
				}
				__antithesis_instrumentation__.Notify(604882)
				d.Type = collatedTyp
			} else {
				__antithesis_instrumentation__.Notify(604887)
			}
		case *ColumnDefault:
			__antithesis_instrumentation__.Notify(604852)
			if d.HasDefaultExpr() || func() bool {
				__antithesis_instrumentation__.Notify(604888)
				return d.GeneratedIdentity.IsGeneratedAsIdentity == true
			}() == true {
				__antithesis_instrumentation__.Notify(604889)
				return nil, pgerror.Newf(pgcode.Syntax,
					"multiple default values specified for column %q", name)
			} else {
				__antithesis_instrumentation__.Notify(604890)
			}
			__antithesis_instrumentation__.Notify(604853)
			d.DefaultExpr.Expr = t.Expr
			d.DefaultExpr.ConstraintName = c.Name
		case *ColumnOnUpdate:
			__antithesis_instrumentation__.Notify(604854)
			if d.HasOnUpdateExpr() {
				__antithesis_instrumentation__.Notify(604891)
				return nil, pgerror.Newf(pgcode.Syntax,
					"multiple ON UPDATE values specified for column %q", name)
			} else {
				__antithesis_instrumentation__.Notify(604892)
			}
			__antithesis_instrumentation__.Notify(604855)
			if d.GeneratedIdentity.IsGeneratedAsIdentity {
				__antithesis_instrumentation__.Notify(604893)
				return nil, pgerror.Newf(pgcode.Syntax,
					"both generated identity and on update expression specified for column %q",
					name)
			} else {
				__antithesis_instrumentation__.Notify(604894)
			}
			__antithesis_instrumentation__.Notify(604856)
			d.OnUpdateExpr.Expr = t.Expr
			d.OnUpdateExpr.ConstraintName = c.Name
		case *GeneratedAlwaysAsIdentity, *GeneratedByDefAsIdentity:
			__antithesis_instrumentation__.Notify(604857)
			if typ, ok := typRef.(*types.T); !ok || func() bool {
				__antithesis_instrumentation__.Notify(604895)
				return typ.InternalType.Family != types.IntFamily == true
			}() == true {
				__antithesis_instrumentation__.Notify(604896)
				return nil, pgerror.Newf(
					pgcode.InvalidParameterValue,
					"identity column type must be an INT",
				)
			} else {
				__antithesis_instrumentation__.Notify(604897)
			}
			__antithesis_instrumentation__.Notify(604858)
			if d.GeneratedIdentity.IsGeneratedAsIdentity {
				__antithesis_instrumentation__.Notify(604898)
				return nil, pgerror.Newf(pgcode.Syntax,
					"multiple identity specifications for column %q", name)
			} else {
				__antithesis_instrumentation__.Notify(604899)
			}
			__antithesis_instrumentation__.Notify(604859)
			if d.HasDefaultExpr() {
				__antithesis_instrumentation__.Notify(604900)
				return nil, pgerror.Newf(pgcode.Syntax,
					"multiple default values specified for column %q", name)
			} else {
				__antithesis_instrumentation__.Notify(604901)
			}
			__antithesis_instrumentation__.Notify(604860)
			if d.Computed.Computed {
				__antithesis_instrumentation__.Notify(604902)
				return nil, pgerror.Newf(pgcode.Syntax,
					"both generated identity and computed expression specified for column %q", name)
			} else {
				__antithesis_instrumentation__.Notify(604903)
			}
			__antithesis_instrumentation__.Notify(604861)
			if d.Nullable.Nullability == Null {
				__antithesis_instrumentation__.Notify(604904)
				return nil, pgerror.Newf(pgcode.Syntax,
					"conflicting NULL/NOT NULL declarations for column %q", name)
			} else {
				__antithesis_instrumentation__.Notify(604905)
			}
			__antithesis_instrumentation__.Notify(604862)
			if d.HasOnUpdateExpr() {
				__antithesis_instrumentation__.Notify(604906)
				return nil, pgerror.Newf(pgcode.Syntax,
					"both generated identity and on update expression specified for column %q",
					name)
			} else {
				__antithesis_instrumentation__.Notify(604907)
			}
			__antithesis_instrumentation__.Notify(604863)
			d.GeneratedIdentity.IsGeneratedAsIdentity = true
			d.Nullable.Nullability = NotNull
			switch c.Qualification.(type) {
			case *GeneratedAlwaysAsIdentity:
				__antithesis_instrumentation__.Notify(604908)
				d.GeneratedIdentity.GeneratedAsIdentityType = GeneratedAlways
				d.GeneratedIdentity.SeqOptions = t.(*GeneratedAlwaysAsIdentity).SeqOptions
			case *GeneratedByDefAsIdentity:
				__antithesis_instrumentation__.Notify(604909)
				d.GeneratedIdentity.GeneratedAsIdentityType = GeneratedByDefault
				d.GeneratedIdentity.SeqOptions = t.(*GeneratedByDefAsIdentity).SeqOptions
			}
		case HiddenConstraint:
			__antithesis_instrumentation__.Notify(604864)
			d.Hidden = true
		case NotNullConstraint:
			__antithesis_instrumentation__.Notify(604865)
			if d.Nullable.Nullability == Null {
				__antithesis_instrumentation__.Notify(604910)
				return nil, pgerror.Newf(pgcode.Syntax,
					"conflicting NULL/NOT NULL declarations for column %q", name)
			} else {
				__antithesis_instrumentation__.Notify(604911)
			}
			__antithesis_instrumentation__.Notify(604866)
			d.Nullable.Nullability = NotNull
			d.Nullable.ConstraintName = c.Name
		case NullConstraint:
			__antithesis_instrumentation__.Notify(604867)
			if d.Nullable.Nullability == NotNull {
				__antithesis_instrumentation__.Notify(604912)
				return nil, pgerror.Newf(pgcode.Syntax,
					"conflicting NULL/NOT NULL declarations for column %q", name)
			} else {
				__antithesis_instrumentation__.Notify(604913)
			}
			__antithesis_instrumentation__.Notify(604868)
			d.Nullable.Nullability = Null
			d.Nullable.ConstraintName = c.Name
		case PrimaryKeyConstraint:
			__antithesis_instrumentation__.Notify(604869)
			d.PrimaryKey.IsPrimaryKey = true
			d.PrimaryKey.StorageParams = c.Qualification.(PrimaryKeyConstraint).StorageParams
			d.Unique.ConstraintName = c.Name
		case ShardedPrimaryKeyConstraint:
			__antithesis_instrumentation__.Notify(604870)
			d.PrimaryKey.IsPrimaryKey = true
			constraint := c.Qualification.(ShardedPrimaryKeyConstraint)
			d.PrimaryKey.Sharded = true
			d.PrimaryKey.ShardBuckets = constraint.ShardBuckets
			d.PrimaryKey.StorageParams = constraint.StorageParams
			d.Unique.ConstraintName = c.Name
		case UniqueConstraint:
			__antithesis_instrumentation__.Notify(604871)
			d.Unique.IsUnique = true
			d.Unique.WithoutIndex = t.WithoutIndex
			d.Unique.ConstraintName = c.Name
		case *ColumnCheckConstraint:
			__antithesis_instrumentation__.Notify(604872)
			d.CheckExprs = append(d.CheckExprs, ColumnTableDefCheckExpr{
				Expr:           t.Expr,
				ConstraintName: c.Name,
			})
		case *ColumnFKConstraint:
			__antithesis_instrumentation__.Notify(604873)
			if d.HasFKConstraint() {
				__antithesis_instrumentation__.Notify(604914)
				return nil, pgerror.Newf(pgcode.InvalidTableDefinition,
					"multiple foreign key constraints specified for column %q", name)
			} else {
				__antithesis_instrumentation__.Notify(604915)
			}
			__antithesis_instrumentation__.Notify(604874)
			d.References.Table = &t.Table
			d.References.Col = t.Col
			d.References.ConstraintName = c.Name
			d.References.Actions = t.Actions
			d.References.Match = t.Match
		case *ColumnComputedDef:
			__antithesis_instrumentation__.Notify(604875)
			if d.GeneratedIdentity.IsGeneratedAsIdentity {
				__antithesis_instrumentation__.Notify(604916)
				return nil, pgerror.Newf(pgcode.Syntax,
					"both generated identity and computed expression specified for column %q", name)
			} else {
				__antithesis_instrumentation__.Notify(604917)
			}
			__antithesis_instrumentation__.Notify(604876)
			d.Computed.Computed = true
			d.Computed.Expr = t.Expr
			d.Computed.Virtual = t.Virtual
		case *ColumnFamilyConstraint:
			__antithesis_instrumentation__.Notify(604877)
			if d.HasColumnFamily() {
				__antithesis_instrumentation__.Notify(604918)
				return nil, pgerror.Newf(pgcode.InvalidTableDefinition,
					"multiple column families specified for column %q", name)
			} else {
				__antithesis_instrumentation__.Notify(604919)
			}
			__antithesis_instrumentation__.Notify(604878)
			d.Family.Name = t.Family
			d.Family.Create = t.Create
			d.Family.IfNotExists = t.IfNotExists
		default:
			__antithesis_instrumentation__.Notify(604879)
			return nil, errors.AssertionFailedf("unexpected column qualification: %T", c)
		}
	}
	__antithesis_instrumentation__.Notify(604849)

	return d, nil
}

func (node *ColumnTableDef) HasDefaultExpr() bool {
	__antithesis_instrumentation__.Notify(604920)
	return node.DefaultExpr.Expr != nil
}

func (node *ColumnTableDef) HasOnUpdateExpr() bool {
	__antithesis_instrumentation__.Notify(604921)
	return node.OnUpdateExpr.Expr != nil
}

func (node *ColumnTableDef) HasFKConstraint() bool {
	__antithesis_instrumentation__.Notify(604922)
	return node.References.Table != nil
}

func (node *ColumnTableDef) IsComputed() bool {
	__antithesis_instrumentation__.Notify(604923)
	return node.Computed.Computed
}

func (node *ColumnTableDef) IsVirtual() bool {
	__antithesis_instrumentation__.Notify(604924)
	return node.Computed.Virtual
}

func (node *ColumnTableDef) HasColumnFamily() bool {
	__antithesis_instrumentation__.Notify(604925)
	return node.Family.Name != "" || func() bool {
		__antithesis_instrumentation__.Notify(604926)
		return node.Family.Create == true
	}() == true
}

func (node *ColumnTableDef) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604927)
	ctx.FormatNode(&node.Name)

	if node.Type != nil {
		__antithesis_instrumentation__.Notify(604939)
		ctx.WriteByte(' ')
		ctx.WriteString(node.columnTypeString())
	} else {
		__antithesis_instrumentation__.Notify(604940)
	}
	__antithesis_instrumentation__.Notify(604928)

	if node.Nullable.Nullability != SilentNull && func() bool {
		__antithesis_instrumentation__.Notify(604941)
		return node.Nullable.ConstraintName != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(604942)
		ctx.WriteString(" CONSTRAINT ")
		ctx.FormatNode(&node.Nullable.ConstraintName)
	} else {
		__antithesis_instrumentation__.Notify(604943)
	}
	__antithesis_instrumentation__.Notify(604929)
	switch node.Nullable.Nullability {
	case Null:
		__antithesis_instrumentation__.Notify(604944)
		ctx.WriteString(" NULL")
	case NotNull:
		__antithesis_instrumentation__.Notify(604945)
		ctx.WriteString(" NOT NULL")
	default:
		__antithesis_instrumentation__.Notify(604946)
	}
	__antithesis_instrumentation__.Notify(604930)
	if node.Hidden {
		__antithesis_instrumentation__.Notify(604947)
		ctx.WriteString(" NOT VISIBLE")
	} else {
		__antithesis_instrumentation__.Notify(604948)
	}
	__antithesis_instrumentation__.Notify(604931)
	if node.PrimaryKey.IsPrimaryKey || func() bool {
		__antithesis_instrumentation__.Notify(604949)
		return node.Unique.IsUnique == true
	}() == true {
		__antithesis_instrumentation__.Notify(604950)
		if node.Unique.ConstraintName != "" {
			__antithesis_instrumentation__.Notify(604952)
			ctx.WriteString(" CONSTRAINT ")
			ctx.FormatNode(&node.Unique.ConstraintName)
		} else {
			__antithesis_instrumentation__.Notify(604953)
		}
		__antithesis_instrumentation__.Notify(604951)
		if node.PrimaryKey.IsPrimaryKey {
			__antithesis_instrumentation__.Notify(604954)
			ctx.WriteString(" PRIMARY KEY")

			pkStorageParams := node.PrimaryKey.StorageParams
			if node.PrimaryKey.Sharded {
				__antithesis_instrumentation__.Notify(604956)
				ctx.WriteString(" USING HASH")
				bcStorageParam := node.PrimaryKey.StorageParams.GetVal(`bucket_count`)
				if _, ok := node.PrimaryKey.ShardBuckets.(DefaultVal); !ok && func() bool {
					__antithesis_instrumentation__.Notify(604957)
					return bcStorageParam == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(604958)
					pkStorageParams = append(
						pkStorageParams,
						StorageParam{
							Key:   `bucket_count`,
							Value: node.PrimaryKey.ShardBuckets,
						},
					)
				} else {
					__antithesis_instrumentation__.Notify(604959)
				}
			} else {
				__antithesis_instrumentation__.Notify(604960)
			}
			__antithesis_instrumentation__.Notify(604955)
			if len(pkStorageParams) > 0 {
				__antithesis_instrumentation__.Notify(604961)
				ctx.WriteString(" WITH (")
				ctx.FormatNode(&pkStorageParams)
				ctx.WriteString(")")
			} else {
				__antithesis_instrumentation__.Notify(604962)
			}
		} else {
			__antithesis_instrumentation__.Notify(604963)
			if node.Unique.IsUnique {
				__antithesis_instrumentation__.Notify(604964)
				ctx.WriteString(" UNIQUE")
				if node.Unique.WithoutIndex {
					__antithesis_instrumentation__.Notify(604965)
					ctx.WriteString(" WITHOUT INDEX")
				} else {
					__antithesis_instrumentation__.Notify(604966)
				}
			} else {
				__antithesis_instrumentation__.Notify(604967)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(604968)
	}
	__antithesis_instrumentation__.Notify(604932)
	if node.HasDefaultExpr() {
		__antithesis_instrumentation__.Notify(604969)
		if node.DefaultExpr.ConstraintName != "" {
			__antithesis_instrumentation__.Notify(604971)
			ctx.WriteString(" CONSTRAINT ")
			ctx.FormatNode(&node.DefaultExpr.ConstraintName)
		} else {
			__antithesis_instrumentation__.Notify(604972)
		}
		__antithesis_instrumentation__.Notify(604970)
		ctx.WriteString(" DEFAULT ")
		ctx.FormatNode(node.DefaultExpr.Expr)
	} else {
		__antithesis_instrumentation__.Notify(604973)
	}
	__antithesis_instrumentation__.Notify(604933)
	if node.HasOnUpdateExpr() {
		__antithesis_instrumentation__.Notify(604974)
		if node.OnUpdateExpr.ConstraintName != "" {
			__antithesis_instrumentation__.Notify(604976)
			ctx.WriteString(" CONSTRAINT ")
			ctx.FormatNode(&node.OnUpdateExpr.ConstraintName)
		} else {
			__antithesis_instrumentation__.Notify(604977)
		}
		__antithesis_instrumentation__.Notify(604975)
		ctx.WriteString(" ON UPDATE ")
		ctx.FormatNode(node.OnUpdateExpr.Expr)
	} else {
		__antithesis_instrumentation__.Notify(604978)
	}
	__antithesis_instrumentation__.Notify(604934)
	if node.GeneratedIdentity.IsGeneratedAsIdentity {
		__antithesis_instrumentation__.Notify(604979)
		switch node.GeneratedIdentity.GeneratedAsIdentityType {
		case GeneratedAlways:
			__antithesis_instrumentation__.Notify(604981)
			ctx.WriteString(" GENERATED ALWAYS AS IDENTITY")
		case GeneratedByDefault:
			__antithesis_instrumentation__.Notify(604982)
			ctx.WriteString(" GENERATED BY DEFAULT AS IDENTITY")
		default:
			__antithesis_instrumentation__.Notify(604983)
		}
		__antithesis_instrumentation__.Notify(604980)
		if genSeqOpt := node.GeneratedIdentity.SeqOptions; genSeqOpt != nil {
			__antithesis_instrumentation__.Notify(604984)
			ctx.WriteString(" (")

			genSeqOpt.Format(ctx)
			ctx.WriteString(" ) ")
		} else {
			__antithesis_instrumentation__.Notify(604985)
		}
	} else {
		__antithesis_instrumentation__.Notify(604986)
	}
	__antithesis_instrumentation__.Notify(604935)
	for _, checkExpr := range node.CheckExprs {
		__antithesis_instrumentation__.Notify(604987)
		if checkExpr.ConstraintName != "" {
			__antithesis_instrumentation__.Notify(604989)
			ctx.WriteString(" CONSTRAINT ")
			ctx.FormatNode(&checkExpr.ConstraintName)
		} else {
			__antithesis_instrumentation__.Notify(604990)
		}
		__antithesis_instrumentation__.Notify(604988)
		ctx.WriteString(" CHECK (")
		ctx.FormatNode(checkExpr.Expr)
		ctx.WriteByte(')')
	}
	__antithesis_instrumentation__.Notify(604936)
	if node.HasFKConstraint() {
		__antithesis_instrumentation__.Notify(604991)
		if node.References.ConstraintName != "" {
			__antithesis_instrumentation__.Notify(604995)
			ctx.WriteString(" CONSTRAINT ")
			ctx.FormatNode(&node.References.ConstraintName)
		} else {
			__antithesis_instrumentation__.Notify(604996)
		}
		__antithesis_instrumentation__.Notify(604992)
		ctx.WriteString(" REFERENCES ")
		ctx.FormatNode(node.References.Table)
		if node.References.Col != "" {
			__antithesis_instrumentation__.Notify(604997)
			ctx.WriteString(" (")
			ctx.FormatNode(&node.References.Col)
			ctx.WriteByte(')')
		} else {
			__antithesis_instrumentation__.Notify(604998)
		}
		__antithesis_instrumentation__.Notify(604993)
		if node.References.Match != MatchSimple {
			__antithesis_instrumentation__.Notify(604999)
			ctx.WriteByte(' ')
			ctx.WriteString(node.References.Match.String())
		} else {
			__antithesis_instrumentation__.Notify(605000)
		}
		__antithesis_instrumentation__.Notify(604994)
		ctx.FormatNode(&node.References.Actions)
	} else {
		__antithesis_instrumentation__.Notify(605001)
	}
	__antithesis_instrumentation__.Notify(604937)
	if node.IsComputed() {
		__antithesis_instrumentation__.Notify(605002)
		ctx.WriteString(" AS (")
		ctx.FormatNode(node.Computed.Expr)
		if node.Computed.Virtual {
			__antithesis_instrumentation__.Notify(605003)
			ctx.WriteString(") VIRTUAL")
		} else {
			__antithesis_instrumentation__.Notify(605004)
			ctx.WriteString(") STORED")
		}
	} else {
		__antithesis_instrumentation__.Notify(605005)
	}
	__antithesis_instrumentation__.Notify(604938)
	if node.HasColumnFamily() {
		__antithesis_instrumentation__.Notify(605006)
		if node.Family.Create {
			__antithesis_instrumentation__.Notify(605008)
			ctx.WriteString(" CREATE")
			if node.Family.IfNotExists {
				__antithesis_instrumentation__.Notify(605009)
				ctx.WriteString(" IF NOT EXISTS")
			} else {
				__antithesis_instrumentation__.Notify(605010)
			}
		} else {
			__antithesis_instrumentation__.Notify(605011)
		}
		__antithesis_instrumentation__.Notify(605007)
		ctx.WriteString(" FAMILY")
		if len(node.Family.Name) > 0 {
			__antithesis_instrumentation__.Notify(605012)
			ctx.WriteByte(' ')
			ctx.FormatNode(&node.Family.Name)
		} else {
			__antithesis_instrumentation__.Notify(605013)
		}
	} else {
		__antithesis_instrumentation__.Notify(605014)
	}
}

func (node *ColumnTableDef) columnTypeString() string {
	__antithesis_instrumentation__.Notify(605015)
	if node.IsSerial {
		__antithesis_instrumentation__.Notify(605017)

		switch MustBeStaticallyKnownType(node.Type).Width() {
		case 16:
			__antithesis_instrumentation__.Notify(605019)
			return "SERIAL2"
		case 32:
			__antithesis_instrumentation__.Notify(605020)
			return "SERIAL4"
		default:
			__antithesis_instrumentation__.Notify(605021)
		}
		__antithesis_instrumentation__.Notify(605018)
		return "SERIAL8"
	} else {
		__antithesis_instrumentation__.Notify(605022)
	}
	__antithesis_instrumentation__.Notify(605016)
	return node.Type.SQLString()
}

func (node *ColumnTableDef) String() string {
	__antithesis_instrumentation__.Notify(605023)
	return AsString(node)
}

type NamedColumnQualification struct {
	Name          Name
	Qualification ColumnQualification
}

type ColumnQualification interface {
	columnQualification()
}

func (ColumnCollation) columnQualification()      { __antithesis_instrumentation__.Notify(605024) }
func (*ColumnDefault) columnQualification()       { __antithesis_instrumentation__.Notify(605025) }
func (*ColumnOnUpdate) columnQualification()      { __antithesis_instrumentation__.Notify(605026) }
func (NotNullConstraint) columnQualification()    { __antithesis_instrumentation__.Notify(605027) }
func (NullConstraint) columnQualification()       { __antithesis_instrumentation__.Notify(605028) }
func (HiddenConstraint) columnQualification()     { __antithesis_instrumentation__.Notify(605029) }
func (PrimaryKeyConstraint) columnQualification() { __antithesis_instrumentation__.Notify(605030) }
func (ShardedPrimaryKeyConstraint) columnQualification() {
	__antithesis_instrumentation__.Notify(605031)
}
func (UniqueConstraint) columnQualification()        { __antithesis_instrumentation__.Notify(605032) }
func (*ColumnCheckConstraint) columnQualification()  { __antithesis_instrumentation__.Notify(605033) }
func (*ColumnComputedDef) columnQualification()      { __antithesis_instrumentation__.Notify(605034) }
func (*ColumnFKConstraint) columnQualification()     { __antithesis_instrumentation__.Notify(605035) }
func (*ColumnFamilyConstraint) columnQualification() { __antithesis_instrumentation__.Notify(605036) }
func (*GeneratedAlwaysAsIdentity) columnQualification() {
	__antithesis_instrumentation__.Notify(605037)
}
func (*GeneratedByDefAsIdentity) columnQualification() { __antithesis_instrumentation__.Notify(605038) }

type ColumnCollation string

type ColumnDefault struct {
	Expr Expr
}

type ColumnOnUpdate struct {
	Expr Expr
}

type GeneratedAlwaysAsIdentity struct {
	SeqOptions SequenceOptions
}

type GeneratedByDefAsIdentity struct {
	SeqOptions SequenceOptions
}

type NotNullConstraint struct{}

type NullConstraint struct{}

type HiddenConstraint struct{}

type PrimaryKeyConstraint struct {
	StorageParams StorageParams
}

type ShardedPrimaryKeyConstraint struct {
	Sharded       bool
	ShardBuckets  Expr
	StorageParams StorageParams
}

type UniqueConstraint struct {
	WithoutIndex bool
}

type ColumnCheckConstraint struct {
	Expr Expr
}

type ColumnFKConstraint struct {
	Table   TableName
	Col     Name
	Actions ReferenceActions
	Match   CompositeKeyMatchMethod
}

type ColumnComputedDef struct {
	Expr    Expr
	Virtual bool
}

type ColumnFamilyConstraint struct {
	Family      Name
	Create      bool
	IfNotExists bool
}

type IndexTableDef struct {
	Name             Name
	Columns          IndexElemList
	Sharded          *ShardedIndexDef
	Storing          NameList
	Inverted         bool
	PartitionByIndex *PartitionByIndex
	StorageParams    StorageParams
	Predicate        Expr
}

func (node *IndexTableDef) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605039)
	if node.Inverted {
		__antithesis_instrumentation__.Notify(605046)
		ctx.WriteString("INVERTED ")
	} else {
		__antithesis_instrumentation__.Notify(605047)
	}
	__antithesis_instrumentation__.Notify(605040)
	ctx.WriteString("INDEX ")
	if node.Name != "" {
		__antithesis_instrumentation__.Notify(605048)
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	} else {
		__antithesis_instrumentation__.Notify(605049)
	}
	__antithesis_instrumentation__.Notify(605041)
	ctx.WriteByte('(')
	ctx.FormatNode(&node.Columns)
	ctx.WriteByte(')')
	if node.Sharded != nil {
		__antithesis_instrumentation__.Notify(605050)
		ctx.FormatNode(node.Sharded)
	} else {
		__antithesis_instrumentation__.Notify(605051)
	}
	__antithesis_instrumentation__.Notify(605042)
	if node.Storing != nil {
		__antithesis_instrumentation__.Notify(605052)
		ctx.WriteString(" STORING (")
		ctx.FormatNode(&node.Storing)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(605053)
	}
	__antithesis_instrumentation__.Notify(605043)
	if node.PartitionByIndex != nil {
		__antithesis_instrumentation__.Notify(605054)
		ctx.FormatNode(node.PartitionByIndex)
	} else {
		__antithesis_instrumentation__.Notify(605055)
	}
	__antithesis_instrumentation__.Notify(605044)
	if node.StorageParams != nil {
		__antithesis_instrumentation__.Notify(605056)
		ctx.WriteString(" WITH (")
		ctx.FormatNode(&node.StorageParams)
		ctx.WriteString(")")
	} else {
		__antithesis_instrumentation__.Notify(605057)
	}
	__antithesis_instrumentation__.Notify(605045)
	if node.Predicate != nil {
		__antithesis_instrumentation__.Notify(605058)
		ctx.WriteString(" WHERE ")
		ctx.FormatNode(node.Predicate)
	} else {
		__antithesis_instrumentation__.Notify(605059)
	}
}

type ConstraintTableDef interface {
	TableDef

	constraintTableDef()

	SetName(name Name)

	SetIfNotExists()
}

func (*UniqueConstraintTableDef) constraintTableDef() { __antithesis_instrumentation__.Notify(605060) }
func (*ForeignKeyConstraintTableDef) constraintTableDef() {
	__antithesis_instrumentation__.Notify(605061)
}
func (*CheckConstraintTableDef) constraintTableDef() { __antithesis_instrumentation__.Notify(605062) }

type UniqueConstraintTableDef struct {
	IndexTableDef
	PrimaryKey   bool
	WithoutIndex bool
	IfNotExists  bool
}

func (node *UniqueConstraintTableDef) SetName(name Name) {
	__antithesis_instrumentation__.Notify(605063)
	node.Name = name
}

func (node *UniqueConstraintTableDef) SetIfNotExists() {
	__antithesis_instrumentation__.Notify(605064)
	node.IfNotExists = true
}

func (node *UniqueConstraintTableDef) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605065)
	if node.Name != "" {
		__antithesis_instrumentation__.Notify(605072)
		ctx.WriteString("CONSTRAINT ")
		if node.IfNotExists {
			__antithesis_instrumentation__.Notify(605074)
			ctx.WriteString("IF NOT EXISTS ")
		} else {
			__antithesis_instrumentation__.Notify(605075)
		}
		__antithesis_instrumentation__.Notify(605073)
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	} else {
		__antithesis_instrumentation__.Notify(605076)
	}
	__antithesis_instrumentation__.Notify(605066)
	if node.PrimaryKey {
		__antithesis_instrumentation__.Notify(605077)
		ctx.WriteString("PRIMARY KEY ")
	} else {
		__antithesis_instrumentation__.Notify(605078)
		ctx.WriteString("UNIQUE ")
	}
	__antithesis_instrumentation__.Notify(605067)
	if node.WithoutIndex {
		__antithesis_instrumentation__.Notify(605079)
		ctx.WriteString("WITHOUT INDEX ")
	} else {
		__antithesis_instrumentation__.Notify(605080)
	}
	__antithesis_instrumentation__.Notify(605068)
	ctx.WriteByte('(')
	ctx.FormatNode(&node.Columns)
	ctx.WriteByte(')')
	if node.Sharded != nil {
		__antithesis_instrumentation__.Notify(605081)
		ctx.FormatNode(node.Sharded)
	} else {
		__antithesis_instrumentation__.Notify(605082)
	}
	__antithesis_instrumentation__.Notify(605069)
	if node.Storing != nil {
		__antithesis_instrumentation__.Notify(605083)
		ctx.WriteString(" STORING (")
		ctx.FormatNode(&node.Storing)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(605084)
	}
	__antithesis_instrumentation__.Notify(605070)
	if node.PartitionByIndex != nil {
		__antithesis_instrumentation__.Notify(605085)
		ctx.FormatNode(node.PartitionByIndex)
	} else {
		__antithesis_instrumentation__.Notify(605086)
	}
	__antithesis_instrumentation__.Notify(605071)
	if node.Predicate != nil {
		__antithesis_instrumentation__.Notify(605087)
		ctx.WriteString(" WHERE ")
		ctx.FormatNode(node.Predicate)
	} else {
		__antithesis_instrumentation__.Notify(605088)
	}
}

type ReferenceAction int

const (
	NoAction ReferenceAction = iota
	Restrict
	SetNull
	SetDefault
	Cascade
)

var referenceActionName = [...]string{
	NoAction:   "NO ACTION",
	Restrict:   "RESTRICT",
	SetNull:    "SET NULL",
	SetDefault: "SET DEFAULT",
	Cascade:    "CASCADE",
}

func (ra ReferenceAction) String() string {
	__antithesis_instrumentation__.Notify(605089)
	return referenceActionName[ra]
}

type ReferenceActions struct {
	Delete ReferenceAction
	Update ReferenceAction
}

func (node *ReferenceActions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605090)
	if node.Delete != NoAction {
		__antithesis_instrumentation__.Notify(605092)
		ctx.WriteString(" ON DELETE ")
		ctx.WriteString(node.Delete.String())
	} else {
		__antithesis_instrumentation__.Notify(605093)
	}
	__antithesis_instrumentation__.Notify(605091)
	if node.Update != NoAction {
		__antithesis_instrumentation__.Notify(605094)
		ctx.WriteString(" ON UPDATE ")
		ctx.WriteString(node.Update.String())
	} else {
		__antithesis_instrumentation__.Notify(605095)
	}
}

type CompositeKeyMatchMethod int

const (
	MatchSimple CompositeKeyMatchMethod = iota
	MatchFull
	MatchPartial
)

var compositeKeyMatchMethodName = [...]string{
	MatchSimple:  "MATCH SIMPLE",
	MatchFull:    "MATCH FULL",
	MatchPartial: "MATCH PARTIAL",
}

func (c CompositeKeyMatchMethod) String() string {
	__antithesis_instrumentation__.Notify(605096)
	return compositeKeyMatchMethodName[c]
}

type ForeignKeyConstraintTableDef struct {
	Name        Name
	Table       TableName
	FromCols    NameList
	ToCols      NameList
	Actions     ReferenceActions
	Match       CompositeKeyMatchMethod
	IfNotExists bool
}

func (node *ForeignKeyConstraintTableDef) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605097)
	if node.Name != "" {
		__antithesis_instrumentation__.Notify(605101)
		ctx.WriteString("CONSTRAINT ")
		if node.IfNotExists {
			__antithesis_instrumentation__.Notify(605103)
			ctx.WriteString("IF NOT EXISTS ")
		} else {
			__antithesis_instrumentation__.Notify(605104)
		}
		__antithesis_instrumentation__.Notify(605102)
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	} else {
		__antithesis_instrumentation__.Notify(605105)
	}
	__antithesis_instrumentation__.Notify(605098)
	ctx.WriteString("FOREIGN KEY (")
	ctx.FormatNode(&node.FromCols)
	ctx.WriteString(") REFERENCES ")
	ctx.FormatNode(&node.Table)

	if len(node.ToCols) > 0 {
		__antithesis_instrumentation__.Notify(605106)
		ctx.WriteByte(' ')
		ctx.WriteByte('(')
		ctx.FormatNode(&node.ToCols)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(605107)
	}
	__antithesis_instrumentation__.Notify(605099)

	if node.Match != MatchSimple {
		__antithesis_instrumentation__.Notify(605108)
		ctx.WriteByte(' ')
		ctx.WriteString(node.Match.String())
	} else {
		__antithesis_instrumentation__.Notify(605109)
	}
	__antithesis_instrumentation__.Notify(605100)

	ctx.FormatNode(&node.Actions)
}

func (node *ForeignKeyConstraintTableDef) SetName(name Name) {
	__antithesis_instrumentation__.Notify(605110)
	node.Name = name
}

func (node *ForeignKeyConstraintTableDef) SetIfNotExists() {
	__antithesis_instrumentation__.Notify(605111)
	node.IfNotExists = true
}

type CheckConstraintTableDef struct {
	Name        Name
	Expr        Expr
	Hidden      bool
	IfNotExists bool
}

func (node *CheckConstraintTableDef) SetName(name Name) {
	__antithesis_instrumentation__.Notify(605112)
	node.Name = name
}

func (node *CheckConstraintTableDef) SetIfNotExists() {
	__antithesis_instrumentation__.Notify(605113)
	node.IfNotExists = true
}

func (node *CheckConstraintTableDef) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605114)
	if node.Name != "" {
		__antithesis_instrumentation__.Notify(605116)
		ctx.WriteString("CONSTRAINT ")
		if node.IfNotExists {
			__antithesis_instrumentation__.Notify(605118)
			ctx.WriteString("IF NOT EXISTS ")
		} else {
			__antithesis_instrumentation__.Notify(605119)
		}
		__antithesis_instrumentation__.Notify(605117)
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	} else {
		__antithesis_instrumentation__.Notify(605120)
	}
	__antithesis_instrumentation__.Notify(605115)
	ctx.WriteString("CHECK (")
	ctx.FormatNode(node.Expr)
	ctx.WriteByte(')')
}

type FamilyTableDef struct {
	Name    Name
	Columns NameList
}

func (node *FamilyTableDef) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605121)
	ctx.WriteString("FAMILY ")
	if node.Name != "" {
		__antithesis_instrumentation__.Notify(605123)
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	} else {
		__antithesis_instrumentation__.Notify(605124)
	}
	__antithesis_instrumentation__.Notify(605122)
	ctx.WriteByte('(')
	ctx.FormatNode(&node.Columns)
	ctx.WriteByte(')')
}

type ShardedIndexDef struct {
	ShardBuckets Expr
}

func (node *ShardedIndexDef) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605125)
	if _, ok := node.ShardBuckets.(DefaultVal); ok {
		__antithesis_instrumentation__.Notify(605127)
		ctx.WriteString(" USING HASH")
		return
	} else {
		__antithesis_instrumentation__.Notify(605128)
	}
	__antithesis_instrumentation__.Notify(605126)
	ctx.WriteString(" USING HASH WITH BUCKET_COUNT = ")
	ctx.FormatNode(node.ShardBuckets)
}

type PartitionByType string

const (
	PartitionByList PartitionByType = "LIST"

	PartitionByRange PartitionByType = "RANGE"
)

type PartitionByIndex struct {
	*PartitionBy
}

func (node *PartitionByIndex) ContainsPartitions() bool {
	__antithesis_instrumentation__.Notify(605129)
	return node != nil && func() bool {
		__antithesis_instrumentation__.Notify(605130)
		return node.PartitionBy != nil == true
	}() == true
}

func (node *PartitionByIndex) ContainsPartitioningClause() bool {
	__antithesis_instrumentation__.Notify(605131)
	return node != nil
}

type PartitionByTable struct {
	All bool

	*PartitionBy
}

func (node *PartitionByTable) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605132)
	if node == nil {
		__antithesis_instrumentation__.Notify(605135)
		ctx.WriteString(` PARTITION BY NOTHING`)
		return
	} else {
		__antithesis_instrumentation__.Notify(605136)
	}
	__antithesis_instrumentation__.Notify(605133)
	ctx.WriteString(` PARTITION `)
	if node.All {
		__antithesis_instrumentation__.Notify(605137)
		ctx.WriteString(`ALL `)
	} else {
		__antithesis_instrumentation__.Notify(605138)
	}
	__antithesis_instrumentation__.Notify(605134)
	ctx.WriteString(`BY `)
	node.PartitionBy.formatListOrRange(ctx)
}

func (node *PartitionByTable) ContainsPartitions() bool {
	__antithesis_instrumentation__.Notify(605139)
	return node != nil && func() bool {
		__antithesis_instrumentation__.Notify(605140)
		return node.PartitionBy != nil == true
	}() == true
}

func (node *PartitionByTable) ContainsPartitioningClause() bool {
	__antithesis_instrumentation__.Notify(605141)
	return node != nil
}

type PartitionBy struct {
	Fields NameList

	List  []ListPartition
	Range []RangePartition
}

func (node *PartitionBy) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605142)
	ctx.WriteString(` PARTITION BY `)
	node.formatListOrRange(ctx)
}

func (node *PartitionBy) formatListOrRange(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605143)
	if node == nil {
		__antithesis_instrumentation__.Notify(605148)
		ctx.WriteString(`NOTHING`)
		return
	} else {
		__antithesis_instrumentation__.Notify(605149)
	}
	__antithesis_instrumentation__.Notify(605144)
	if len(node.List) > 0 {
		__antithesis_instrumentation__.Notify(605150)
		ctx.WriteString(`LIST (`)
	} else {
		__antithesis_instrumentation__.Notify(605151)
		if len(node.Range) > 0 {
			__antithesis_instrumentation__.Notify(605152)
			ctx.WriteString(`RANGE (`)
		} else {
			__antithesis_instrumentation__.Notify(605153)
		}
	}
	__antithesis_instrumentation__.Notify(605145)
	ctx.FormatNode(&node.Fields)
	ctx.WriteString(`) (`)
	for i := range node.List {
		__antithesis_instrumentation__.Notify(605154)
		if i > 0 {
			__antithesis_instrumentation__.Notify(605156)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(605157)
		}
		__antithesis_instrumentation__.Notify(605155)
		ctx.FormatNode(&node.List[i])
	}
	__antithesis_instrumentation__.Notify(605146)
	for i := range node.Range {
		__antithesis_instrumentation__.Notify(605158)
		if i > 0 {
			__antithesis_instrumentation__.Notify(605160)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(605161)
		}
		__antithesis_instrumentation__.Notify(605159)
		ctx.FormatNode(&node.Range[i])
	}
	__antithesis_instrumentation__.Notify(605147)
	ctx.WriteString(`)`)
}

type ListPartition struct {
	Name         UnrestrictedName
	Exprs        Exprs
	Subpartition *PartitionBy
}

func (node *ListPartition) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605162)
	ctx.WriteString(`PARTITION `)
	ctx.FormatNode(&node.Name)
	ctx.WriteString(` VALUES IN (`)
	ctx.FormatNode(&node.Exprs)
	ctx.WriteByte(')')
	if node.Subpartition != nil {
		__antithesis_instrumentation__.Notify(605163)
		ctx.FormatNode(node.Subpartition)
	} else {
		__antithesis_instrumentation__.Notify(605164)
	}
}

type RangePartition struct {
	Name         UnrestrictedName
	From         Exprs
	To           Exprs
	Subpartition *PartitionBy
}

func (node *RangePartition) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605165)
	ctx.WriteString(`PARTITION `)
	ctx.FormatNode(&node.Name)
	ctx.WriteString(` VALUES FROM (`)
	ctx.FormatNode(&node.From)
	ctx.WriteString(`) TO (`)
	ctx.FormatNode(&node.To)
	ctx.WriteByte(')')
	if node.Subpartition != nil {
		__antithesis_instrumentation__.Notify(605166)
		ctx.FormatNode(node.Subpartition)
	} else {
		__antithesis_instrumentation__.Notify(605167)
	}
}

type StorageParam struct {
	Key   Name
	Value Expr
}

type StorageParams []StorageParam

func (o *StorageParams) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605168)
	for i := range *o {
		__antithesis_instrumentation__.Notify(605169)
		n := &(*o)[i]
		if i > 0 {
			__antithesis_instrumentation__.Notify(605171)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(605172)
		}
		__antithesis_instrumentation__.Notify(605170)

		ctx.FormatNode(&n.Key)
		if n.Value != nil {
			__antithesis_instrumentation__.Notify(605173)
			ctx.WriteString(` = `)
			ctx.FormatNode(n.Value)
		} else {
			__antithesis_instrumentation__.Notify(605174)
		}
	}
}

func (o *StorageParams) GetVal(key string) Expr {
	__antithesis_instrumentation__.Notify(605175)
	k := Name(key)
	for _, param := range *o {
		__antithesis_instrumentation__.Notify(605177)
		if param.Key == k {
			__antithesis_instrumentation__.Notify(605178)
			return param.Value
		} else {
			__antithesis_instrumentation__.Notify(605179)
		}
	}
	__antithesis_instrumentation__.Notify(605176)
	return nil
}

type CreateTableOnCommitSetting uint32

const (
	CreateTableOnCommitUnset CreateTableOnCommitSetting = iota

	CreateTableOnCommitPreserveRows
)

type CreateTable struct {
	IfNotExists      bool
	Table            TableName
	PartitionByTable *PartitionByTable
	Persistence      Persistence
	StorageParams    StorageParams
	OnCommit         CreateTableOnCommitSetting

	Defs     TableDefs
	AsSource *Select
	Locality *Locality
}

func (node *CreateTable) As() bool {
	__antithesis_instrumentation__.Notify(605180)
	return node.AsSource != nil
}

func (node *CreateTable) AsHasUserSpecifiedPrimaryKey() bool {
	__antithesis_instrumentation__.Notify(605181)
	if node.As() {
		__antithesis_instrumentation__.Notify(605183)
		for _, def := range node.Defs {
			__antithesis_instrumentation__.Notify(605184)
			if d, ok := def.(*ColumnTableDef); !ok {
				__antithesis_instrumentation__.Notify(605185)
				return false
			} else {
				__antithesis_instrumentation__.Notify(605186)
				if d.PrimaryKey.IsPrimaryKey {
					__antithesis_instrumentation__.Notify(605187)
					return true
				} else {
					__antithesis_instrumentation__.Notify(605188)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(605189)
	}
	__antithesis_instrumentation__.Notify(605182)
	return false
}

func (node *CreateTable) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605190)
	ctx.WriteString("CREATE ")
	switch node.Persistence {
	case PersistenceTemporary:
		__antithesis_instrumentation__.Notify(605193)
		ctx.WriteString("TEMPORARY ")
	case PersistenceUnlogged:
		__antithesis_instrumentation__.Notify(605194)
		ctx.WriteString("UNLOGGED ")
	default:
		__antithesis_instrumentation__.Notify(605195)
	}
	__antithesis_instrumentation__.Notify(605191)
	ctx.WriteString("TABLE ")
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(605196)
		ctx.WriteString("IF NOT EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(605197)
	}
	__antithesis_instrumentation__.Notify(605192)
	ctx.FormatNode(&node.Table)
	node.FormatBody(ctx)
}

func (node *CreateTable) FormatBody(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605198)
	if node.As() {
		__antithesis_instrumentation__.Notify(605199)
		if len(node.Defs) > 0 {
			__antithesis_instrumentation__.Notify(605201)
			ctx.WriteString(" (")
			ctx.FormatNode(&node.Defs)
			ctx.WriteByte(')')
		} else {
			__antithesis_instrumentation__.Notify(605202)
		}
		__antithesis_instrumentation__.Notify(605200)
		ctx.WriteString(" AS ")
		ctx.FormatNode(node.AsSource)
	} else {
		__antithesis_instrumentation__.Notify(605203)
		ctx.WriteString(" (")
		ctx.FormatNode(&node.Defs)
		ctx.WriteByte(')')
		if node.PartitionByTable != nil {
			__antithesis_instrumentation__.Notify(605206)
			ctx.FormatNode(node.PartitionByTable)
		} else {
			__antithesis_instrumentation__.Notify(605207)
		}
		__antithesis_instrumentation__.Notify(605204)
		if node.StorageParams != nil {
			__antithesis_instrumentation__.Notify(605208)
			ctx.WriteString(` WITH (`)
			ctx.FormatNode(&node.StorageParams)
			ctx.WriteByte(')')
		} else {
			__antithesis_instrumentation__.Notify(605209)
		}
		__antithesis_instrumentation__.Notify(605205)
		if node.Locality != nil {
			__antithesis_instrumentation__.Notify(605210)
			ctx.WriteString(" ")
			ctx.FormatNode(node.Locality)
		} else {
			__antithesis_instrumentation__.Notify(605211)
		}
	}
}

func (node *CreateTable) HoistConstraints() {
	__antithesis_instrumentation__.Notify(605212)
	for _, d := range node.Defs {
		__antithesis_instrumentation__.Notify(605213)
		if col, ok := d.(*ColumnTableDef); ok {
			__antithesis_instrumentation__.Notify(605214)
			for _, checkExpr := range col.CheckExprs {
				__antithesis_instrumentation__.Notify(605216)
				node.Defs = append(node.Defs,
					&CheckConstraintTableDef{
						Expr: checkExpr.Expr,
						Name: checkExpr.ConstraintName,
					},
				)
			}
			__antithesis_instrumentation__.Notify(605215)
			col.CheckExprs = nil
			if col.HasFKConstraint() {
				__antithesis_instrumentation__.Notify(605217)
				var targetCol NameList
				if col.References.Col != "" {
					__antithesis_instrumentation__.Notify(605219)
					targetCol = append(targetCol, col.References.Col)
				} else {
					__antithesis_instrumentation__.Notify(605220)
				}
				__antithesis_instrumentation__.Notify(605218)
				node.Defs = append(node.Defs, &ForeignKeyConstraintTableDef{
					Table:    *col.References.Table,
					FromCols: NameList{col.Name},
					ToCols:   targetCol,
					Name:     col.References.ConstraintName,
					Actions:  col.References.Actions,
					Match:    col.References.Match,
				})
				col.References.Table = nil
			} else {
				__antithesis_instrumentation__.Notify(605221)
			}
		} else {
			__antithesis_instrumentation__.Notify(605222)
		}
	}
}

type CreateSchema struct {
	IfNotExists bool
	AuthRole    RoleSpec
	Schema      ObjectNamePrefix
}

func (node *CreateSchema) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605223)
	ctx.WriteString("CREATE SCHEMA")

	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(605226)
		ctx.WriteString(" IF NOT EXISTS")
	} else {
		__antithesis_instrumentation__.Notify(605227)
	}
	__antithesis_instrumentation__.Notify(605224)

	if node.Schema.ExplicitSchema {
		__antithesis_instrumentation__.Notify(605228)
		ctx.WriteString(" ")
		ctx.FormatNode(&node.Schema)
	} else {
		__antithesis_instrumentation__.Notify(605229)
	}
	__antithesis_instrumentation__.Notify(605225)

	if !node.AuthRole.Undefined() {
		__antithesis_instrumentation__.Notify(605230)
		ctx.WriteString(" AUTHORIZATION ")
		ctx.FormatNode(&node.AuthRole)
	} else {
		__antithesis_instrumentation__.Notify(605231)
	}
}

type CreateSequence struct {
	IfNotExists bool
	Name        TableName
	Persistence Persistence
	Options     SequenceOptions
}

func (node *CreateSequence) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605232)
	ctx.WriteString("CREATE ")

	if node.Persistence == PersistenceTemporary {
		__antithesis_instrumentation__.Notify(605235)
		ctx.WriteString("TEMPORARY ")
	} else {
		__antithesis_instrumentation__.Notify(605236)
	}
	__antithesis_instrumentation__.Notify(605233)

	ctx.WriteString("SEQUENCE ")

	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(605237)
		ctx.WriteString("IF NOT EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(605238)
	}
	__antithesis_instrumentation__.Notify(605234)
	ctx.FormatNode(&node.Name)
	ctx.FormatNode(&node.Options)
}

type SequenceOptions []SequenceOption

func (node *SequenceOptions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605239)
	for i := range *node {
		__antithesis_instrumentation__.Notify(605240)
		option := &(*node)[i]
		ctx.WriteByte(' ')
		switch option.Name {
		case SeqOptAs:
			__antithesis_instrumentation__.Notify(605241)
			ctx.WriteString(option.Name)
			ctx.WriteByte(' ')
			ctx.WriteString(option.AsIntegerType.SQLString())
		case SeqOptCycle, SeqOptNoCycle:
			__antithesis_instrumentation__.Notify(605242)
			ctx.WriteString(option.Name)
		case SeqOptCache:
			__antithesis_instrumentation__.Notify(605243)
			ctx.WriteString(option.Name)
			ctx.WriteByte(' ')

			if ctx.flags.HasFlags(FmtHideConstants) {
				__antithesis_instrumentation__.Notify(605252)
				ctx.WriteByte('0')
			} else {
				__antithesis_instrumentation__.Notify(605253)
				ctx.Printf("%d", *option.IntVal)
			}
		case SeqOptMaxValue, SeqOptMinValue:
			__antithesis_instrumentation__.Notify(605244)
			if option.IntVal == nil {
				__antithesis_instrumentation__.Notify(605254)
				ctx.WriteString("NO ")
				ctx.WriteString(option.Name)
			} else {
				__antithesis_instrumentation__.Notify(605255)
				ctx.WriteString(option.Name)
				ctx.WriteByte(' ')

				if ctx.flags.HasFlags(FmtHideConstants) {
					__antithesis_instrumentation__.Notify(605256)
					ctx.WriteByte('0')
				} else {
					__antithesis_instrumentation__.Notify(605257)
					ctx.Printf("%d", *option.IntVal)
				}
			}
		case SeqOptStart:
			__antithesis_instrumentation__.Notify(605245)
			ctx.WriteString(option.Name)
			ctx.WriteByte(' ')
			if option.OptionalWord {
				__antithesis_instrumentation__.Notify(605258)
				ctx.WriteString("WITH ")
			} else {
				__antithesis_instrumentation__.Notify(605259)
			}
			__antithesis_instrumentation__.Notify(605246)

			if ctx.flags.HasFlags(FmtHideConstants) {
				__antithesis_instrumentation__.Notify(605260)
				ctx.WriteByte('0')
			} else {
				__antithesis_instrumentation__.Notify(605261)
				ctx.Printf("%d", *option.IntVal)
			}
		case SeqOptIncrement:
			__antithesis_instrumentation__.Notify(605247)
			ctx.WriteString(option.Name)
			ctx.WriteByte(' ')
			if option.OptionalWord {
				__antithesis_instrumentation__.Notify(605262)
				ctx.WriteString("BY ")
			} else {
				__antithesis_instrumentation__.Notify(605263)
			}
			__antithesis_instrumentation__.Notify(605248)

			if ctx.flags.HasFlags(FmtHideConstants) {
				__antithesis_instrumentation__.Notify(605264)
				ctx.WriteByte('0')
			} else {
				__antithesis_instrumentation__.Notify(605265)
				ctx.Printf("%d", *option.IntVal)
			}
		case SeqOptVirtual:
			__antithesis_instrumentation__.Notify(605249)
			ctx.WriteString(option.Name)
		case SeqOptOwnedBy:
			__antithesis_instrumentation__.Notify(605250)
			ctx.WriteString(option.Name)
			ctx.WriteByte(' ')
			switch option.ColumnItemVal {
			case nil:
				__antithesis_instrumentation__.Notify(605266)
				ctx.WriteString("NONE")
			default:
				__antithesis_instrumentation__.Notify(605267)
				ctx.FormatNode(option.ColumnItemVal)
			}
		default:
			__antithesis_instrumentation__.Notify(605251)
			panic(errors.AssertionFailedf("unexpected SequenceOption: %v", option))
		}
	}
}

type SequenceOption struct {
	Name string

	AsIntegerType *types.T

	IntVal *int64

	OptionalWord bool

	ColumnItemVal *ColumnItem
}

const (
	SeqOptAs        = "AS"
	SeqOptCycle     = "CYCLE"
	SeqOptNoCycle   = "NO CYCLE"
	SeqOptOwnedBy   = "OWNED BY"
	SeqOptCache     = "CACHE"
	SeqOptIncrement = "INCREMENT"
	SeqOptMinValue  = "MINVALUE"
	SeqOptMaxValue  = "MAXVALUE"
	SeqOptStart     = "START"
	SeqOptVirtual   = "VIRTUAL"

	_ = SeqOptAs
)

type LikeTableDef struct {
	Name    TableName
	Options []LikeTableOption
}

type LikeTableOption struct {
	Excluded bool
	Opt      LikeTableOpt
}

func (def *LikeTableDef) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605268)
	ctx.WriteString("LIKE ")
	ctx.FormatNode(&def.Name)
	for _, o := range def.Options {
		__antithesis_instrumentation__.Notify(605269)
		ctx.WriteString(" ")
		ctx.FormatNode(o)
	}
}

func (l LikeTableOption) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605270)
	if l.Excluded {
		__antithesis_instrumentation__.Notify(605272)
		ctx.WriteString("EXCLUDING ")
	} else {
		__antithesis_instrumentation__.Notify(605273)
		ctx.WriteString("INCLUDING ")
	}
	__antithesis_instrumentation__.Notify(605271)
	ctx.WriteString(l.Opt.String())
}

type LikeTableOpt int

const (
	LikeTableOptConstraints LikeTableOpt = 1 << iota
	LikeTableOptDefaults
	LikeTableOptGenerated
	LikeTableOptIndexes

	likeTableOptInvalid
)

const LikeTableOptAll = ^likeTableOptInvalid

func (o LikeTableOpt) Has(other LikeTableOpt) bool {
	__antithesis_instrumentation__.Notify(605274)
	return int(o)&int(other) != 0
}

func (o LikeTableOpt) String() string {
	__antithesis_instrumentation__.Notify(605275)
	switch o {
	case LikeTableOptConstraints:
		__antithesis_instrumentation__.Notify(605276)
		return "CONSTRAINTS"
	case LikeTableOptDefaults:
		__antithesis_instrumentation__.Notify(605277)
		return "DEFAULTS"
	case LikeTableOptGenerated:
		__antithesis_instrumentation__.Notify(605278)
		return "GENERATED"
	case LikeTableOptIndexes:
		__antithesis_instrumentation__.Notify(605279)
		return "INDEXES"
	case LikeTableOptAll:
		__antithesis_instrumentation__.Notify(605280)
		return "ALL"
	default:
		__antithesis_instrumentation__.Notify(605281)
		panic("unknown like table opt" + strconv.Itoa(int(o)))
	}
}

func (o KVOptions) ToRoleOptions(
	typeAsStringOrNull func(e Expr, op string) (func() (bool, string, error), error), op string,
) (roleoption.List, error) {
	__antithesis_instrumentation__.Notify(605282)
	roleOptions := make(roleoption.List, len(o))

	for i, ro := range o {
		__antithesis_instrumentation__.Notify(605284)

		option, err := roleoption.ToOption(string(ro.Key))
		if err != nil {
			__antithesis_instrumentation__.Notify(605286)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(605287)
		}
		__antithesis_instrumentation__.Notify(605285)

		if ro.Value != nil {
			__antithesis_instrumentation__.Notify(605288)
			if ro.Value == DNull {
				__antithesis_instrumentation__.Notify(605289)
				roleOptions[i] = roleoption.RoleOption{
					Option: option, HasValue: true, Value: func() (bool, string, error) {
						__antithesis_instrumentation__.Notify(605290)
						return true, "", nil
					},
				}
			} else {
				__antithesis_instrumentation__.Notify(605291)
				strFn, err := typeAsStringOrNull(ro.Value, op)
				if err != nil {
					__antithesis_instrumentation__.Notify(605293)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(605294)
				}
				__antithesis_instrumentation__.Notify(605292)
				roleOptions[i] = roleoption.RoleOption{
					Option: option, Value: strFn, HasValue: true,
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(605295)
			roleOptions[i] = roleoption.RoleOption{
				Option: option, HasValue: false,
			}
		}
	}
	__antithesis_instrumentation__.Notify(605283)

	return roleOptions, nil
}

func (o *KVOptions) formatAsRoleOptions(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605296)
	for _, option := range *o {
		__antithesis_instrumentation__.Notify(605297)
		ctx.WriteByte(' ')

		ctx.WriteString(strings.ToUpper(string(option.Key)))

		if strings.HasSuffix(string(option.Key), "password") {
			__antithesis_instrumentation__.Notify(605298)
			ctx.WriteByte(' ')
			if ctx.flags.HasFlags(FmtShowPasswords) {
				__antithesis_instrumentation__.Notify(605299)
				ctx.FormatNode(option.Value)
			} else {
				__antithesis_instrumentation__.Notify(605300)
				ctx.WriteString(PasswordSubstitution)
			}
		} else {
			__antithesis_instrumentation__.Notify(605301)
			if option.Value != nil {
				__antithesis_instrumentation__.Notify(605302)
				ctx.WriteByte(' ')
				if ctx.HasFlags(FmtHideConstants) {
					__antithesis_instrumentation__.Notify(605303)
					ctx.WriteString("'_'")
				} else {
					__antithesis_instrumentation__.Notify(605304)
					ctx.FormatNode(option.Value)
				}
			} else {
				__antithesis_instrumentation__.Notify(605305)
			}
		}
	}
}

type CreateRole struct {
	Name        RoleSpec
	IfNotExists bool
	IsRole      bool
	KVOptions   KVOptions
}

func (node *CreateRole) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605306)
	ctx.WriteString("CREATE")
	if node.IsRole {
		__antithesis_instrumentation__.Notify(605309)
		ctx.WriteString(" ROLE ")
	} else {
		__antithesis_instrumentation__.Notify(605310)
		ctx.WriteString(" USER ")
	}
	__antithesis_instrumentation__.Notify(605307)
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(605311)
		ctx.WriteString("IF NOT EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(605312)
	}
	__antithesis_instrumentation__.Notify(605308)
	ctx.FormatNode(&node.Name)

	if len(node.KVOptions) > 0 {
		__antithesis_instrumentation__.Notify(605313)
		ctx.WriteString(" WITH")
		node.KVOptions.formatAsRoleOptions(ctx)
	} else {
		__antithesis_instrumentation__.Notify(605314)
	}
}

type CreateView struct {
	Name         TableName
	ColumnNames  NameList
	AsSource     *Select
	IfNotExists  bool
	Persistence  Persistence
	Replace      bool
	Materialized bool
}

func (node *CreateView) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605315)
	ctx.WriteString("CREATE ")

	if node.Replace {
		__antithesis_instrumentation__.Notify(605321)
		ctx.WriteString("OR REPLACE ")
	} else {
		__antithesis_instrumentation__.Notify(605322)
	}
	__antithesis_instrumentation__.Notify(605316)

	if node.Persistence == PersistenceTemporary {
		__antithesis_instrumentation__.Notify(605323)
		ctx.WriteString("TEMPORARY ")
	} else {
		__antithesis_instrumentation__.Notify(605324)
	}
	__antithesis_instrumentation__.Notify(605317)

	if node.Materialized {
		__antithesis_instrumentation__.Notify(605325)
		ctx.WriteString("MATERIALIZED ")
	} else {
		__antithesis_instrumentation__.Notify(605326)
	}
	__antithesis_instrumentation__.Notify(605318)

	ctx.WriteString("VIEW ")

	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(605327)
		ctx.WriteString("IF NOT EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(605328)
	}
	__antithesis_instrumentation__.Notify(605319)
	ctx.FormatNode(&node.Name)

	if len(node.ColumnNames) > 0 {
		__antithesis_instrumentation__.Notify(605329)
		ctx.WriteByte(' ')
		ctx.WriteByte('(')
		ctx.FormatNode(&node.ColumnNames)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(605330)
	}
	__antithesis_instrumentation__.Notify(605320)

	ctx.WriteString(" AS ")
	ctx.FormatNode(node.AsSource)
}

type RefreshMaterializedView struct {
	Name              *UnresolvedObjectName
	Concurrently      bool
	RefreshDataOption RefreshDataOption
}

func (node *RefreshMaterializedView) TelemetryCounter() telemetry.Counter {
	__antithesis_instrumentation__.Notify(605331)
	return sqltelemetry.SchemaRefreshMaterializedView
}

type RefreshDataOption int

const (
	RefreshDataDefault RefreshDataOption = iota

	RefreshDataWithData

	RefreshDataClear
)

func (node *RefreshMaterializedView) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605332)
	ctx.WriteString("REFRESH MATERIALIZED VIEW ")
	if node.Concurrently {
		__antithesis_instrumentation__.Notify(605334)
		ctx.WriteString("CONCURRENTLY ")
	} else {
		__antithesis_instrumentation__.Notify(605335)
	}
	__antithesis_instrumentation__.Notify(605333)
	ctx.FormatNode(node.Name)
	switch node.RefreshDataOption {
	case RefreshDataWithData:
		__antithesis_instrumentation__.Notify(605336)
		ctx.WriteString(" WITH DATA")
	case RefreshDataClear:
		__antithesis_instrumentation__.Notify(605337)
		ctx.WriteString(" WITH NO DATA")
	default:
		__antithesis_instrumentation__.Notify(605338)
	}
}

type CreateStats struct {
	Name        Name
	ColumnNames NameList
	Table       TableExpr
	Options     CreateStatsOptions
}

func (node *CreateStats) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605339)
	ctx.WriteString("CREATE STATISTICS ")
	ctx.FormatNode(&node.Name)

	if len(node.ColumnNames) > 0 {
		__antithesis_instrumentation__.Notify(605341)
		ctx.WriteString(" ON ")
		ctx.FormatNode(&node.ColumnNames)
	} else {
		__antithesis_instrumentation__.Notify(605342)
	}
	__antithesis_instrumentation__.Notify(605340)

	ctx.WriteString(" FROM ")
	ctx.FormatNode(node.Table)

	if !node.Options.Empty() {
		__antithesis_instrumentation__.Notify(605343)
		ctx.WriteString(" WITH OPTIONS ")
		ctx.FormatNode(&node.Options)
	} else {
		__antithesis_instrumentation__.Notify(605344)
	}
}

type CreateStatsOptions struct {
	Throttling float64

	AsOf AsOfClause
}

func (o *CreateStatsOptions) Empty() bool {
	__antithesis_instrumentation__.Notify(605345)
	return o.Throttling == 0 && func() bool {
		__antithesis_instrumentation__.Notify(605346)
		return o.AsOf.Expr == nil == true
	}() == true
}

func (o *CreateStatsOptions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605347)
	sep := ""
	if o.Throttling != 0 {
		__antithesis_instrumentation__.Notify(605349)
		ctx.WriteString("THROTTLING ")

		if ctx.flags.HasFlags(FmtHideConstants) {
			__antithesis_instrumentation__.Notify(605351)

			ctx.WriteString("0.001")
		} else {
			__antithesis_instrumentation__.Notify(605352)
			fmt.Fprintf(ctx, "%g", o.Throttling)
		}
		__antithesis_instrumentation__.Notify(605350)
		sep = " "
	} else {
		__antithesis_instrumentation__.Notify(605353)
	}
	__antithesis_instrumentation__.Notify(605348)
	if o.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(605354)
		ctx.WriteString(sep)
		ctx.FormatNode(&o.AsOf)
		sep = " "
	} else {
		__antithesis_instrumentation__.Notify(605355)
	}
}

func (o *CreateStatsOptions) CombineWith(other *CreateStatsOptions) error {
	__antithesis_instrumentation__.Notify(605356)
	if other.Throttling != 0 {
		__antithesis_instrumentation__.Notify(605359)
		if o.Throttling != 0 {
			__antithesis_instrumentation__.Notify(605361)
			return errors.New("THROTTLING specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(605362)
		}
		__antithesis_instrumentation__.Notify(605360)
		o.Throttling = other.Throttling
	} else {
		__antithesis_instrumentation__.Notify(605363)
	}
	__antithesis_instrumentation__.Notify(605357)
	if other.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(605364)
		if o.AsOf.Expr != nil {
			__antithesis_instrumentation__.Notify(605366)
			return errors.New("AS OF specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(605367)
		}
		__antithesis_instrumentation__.Notify(605365)
		o.AsOf = other.AsOf
	} else {
		__antithesis_instrumentation__.Notify(605368)
	}
	__antithesis_instrumentation__.Notify(605358)
	return nil
}

type CreateExtension struct {
	Name        string
	IfNotExists bool
}

func (node *CreateExtension) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605369)
	ctx.WriteString("CREATE EXTENSION ")
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(605371)
		ctx.WriteString("IF NOT EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(605372)
	}
	__antithesis_instrumentation__.Notify(605370)

	ctx.WriteString(node.Name)
}
