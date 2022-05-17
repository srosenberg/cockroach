package descpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

var CompositeKeyMatchMethodValue = [...]ForeignKeyReference_Match{
	tree.MatchSimple:  ForeignKeyReference_SIMPLE,
	tree.MatchFull:    ForeignKeyReference_FULL,
	tree.MatchPartial: ForeignKeyReference_PARTIAL,
}

var ForeignKeyReferenceMatchValue = [...]tree.CompositeKeyMatchMethod{
	ForeignKeyReference_SIMPLE:  tree.MatchSimple,
	ForeignKeyReference_FULL:    tree.MatchFull,
	ForeignKeyReference_PARTIAL: tree.MatchPartial,
}

func (x ForeignKeyReference_Match) String() string {
	__antithesis_instrumentation__.Notify(251520)
	switch x {
	case ForeignKeyReference_SIMPLE:
		__antithesis_instrumentation__.Notify(251521)
		return "MATCH SIMPLE"
	case ForeignKeyReference_FULL:
		__antithesis_instrumentation__.Notify(251522)
		return "MATCH FULL"
	case ForeignKeyReference_PARTIAL:
		__antithesis_instrumentation__.Notify(251523)
		return "MATCH PARTIAL"
	default:
		__antithesis_instrumentation__.Notify(251524)
		return strconv.Itoa(int(x))
	}
}

var ForeignKeyReferenceActionType = [...]tree.ReferenceAction{
	catpb.ForeignKeyAction_NO_ACTION:   tree.NoAction,
	catpb.ForeignKeyAction_RESTRICT:    tree.Restrict,
	catpb.ForeignKeyAction_SET_DEFAULT: tree.SetDefault,
	catpb.ForeignKeyAction_SET_NULL:    tree.SetNull,
	catpb.ForeignKeyAction_CASCADE:     tree.Cascade,
}

var ForeignKeyReferenceActionValue = [...]catpb.ForeignKeyAction{
	tree.NoAction:   catpb.ForeignKeyAction_NO_ACTION,
	tree.Restrict:   catpb.ForeignKeyAction_RESTRICT,
	tree.SetDefault: catpb.ForeignKeyAction_SET_DEFAULT,
	tree.SetNull:    catpb.ForeignKeyAction_SET_NULL,
	tree.Cascade:    catpb.ForeignKeyAction_CASCADE,
}

type ConstraintType string

const (
	ConstraintTypePK ConstraintType = "PRIMARY KEY"

	ConstraintTypeFK ConstraintType = "FOREIGN KEY"

	ConstraintTypeUnique ConstraintType = "UNIQUE"

	ConstraintTypeCheck ConstraintType = "CHECK"
)

type ConstraintDetail struct {
	Kind         ConstraintType
	ConstraintID ConstraintID
	Columns      []string
	Details      string
	Unvalidated  bool

	Index *IndexDescriptor

	UniqueWithoutIndexConstraint *UniqueWithoutIndexConstraint

	FK              *ForeignKeyConstraint
	ReferencedTable *TableDescriptor

	CheckConstraint *TableDescriptor_CheckConstraint
}
