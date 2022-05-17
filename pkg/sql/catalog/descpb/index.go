package descpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	types "github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

func (desc *IndexDescriptor) IsSharded() bool {
	__antithesis_instrumentation__.Notify(251599)
	return desc.Sharded.IsSharded
}

func (desc *IndexDescriptor) IsPartial() bool {
	__antithesis_instrumentation__.Notify(251600)
	return desc.Predicate != ""
}

func (desc *IndexDescriptor) ExplicitColumnStartIdx() int {
	__antithesis_instrumentation__.Notify(251601)
	start := int(desc.Partitioning.NumImplicitColumns)

	if desc.IsSharded() {
		__antithesis_instrumentation__.Notify(251603)
		start++
	} else {
		__antithesis_instrumentation__.Notify(251604)
	}
	__antithesis_instrumentation__.Notify(251602)
	return start
}

func (desc *IndexDescriptor) FillColumns(elems tree.IndexElemList) error {
	__antithesis_instrumentation__.Notify(251605)
	desc.KeyColumnNames = make([]string, 0, len(elems))
	desc.KeyColumnDirections = make([]IndexDescriptor_Direction, 0, len(elems))
	for _, c := range elems {
		__antithesis_instrumentation__.Notify(251607)
		if c.Expr != nil {
			__antithesis_instrumentation__.Notify(251609)
			return errors.AssertionFailedf("index elem expression should have been replaced with a column")
		} else {
			__antithesis_instrumentation__.Notify(251610)
		}
		__antithesis_instrumentation__.Notify(251608)
		desc.KeyColumnNames = append(desc.KeyColumnNames, string(c.Column))
		switch c.Direction {
		case tree.Ascending, tree.DefaultDirection:
			__antithesis_instrumentation__.Notify(251611)
			desc.KeyColumnDirections = append(desc.KeyColumnDirections, IndexDescriptor_ASC)
		case tree.Descending:
			__antithesis_instrumentation__.Notify(251612)
			desc.KeyColumnDirections = append(desc.KeyColumnDirections, IndexDescriptor_DESC)
		default:
			__antithesis_instrumentation__.Notify(251613)
			return fmt.Errorf("invalid direction %s for column %s", c.Direction, c.Column)
		}
	}
	__antithesis_instrumentation__.Notify(251606)
	return nil
}

func (desc *IndexDescriptor) IsValidOriginIndex(originColIDs ColumnIDs) bool {
	__antithesis_instrumentation__.Notify(251614)
	return !desc.IsPartial() && func() bool {
		__antithesis_instrumentation__.Notify(251615)
		return ColumnIDs(desc.KeyColumnIDs).HasPrefix(originColIDs) == true
	}() == true
}

func (desc *IndexDescriptor) explicitColumnIDsWithoutShardColumn() ColumnIDs {
	__antithesis_instrumentation__.Notify(251616)
	explicitColIDs := desc.KeyColumnIDs[desc.ExplicitColumnStartIdx():]
	explicitColNames := desc.KeyColumnNames[desc.ExplicitColumnStartIdx():]
	colIDs := make(ColumnIDs, 0, len(explicitColIDs))
	for i := range explicitColNames {
		__antithesis_instrumentation__.Notify(251618)
		if !desc.IsSharded() || func() bool {
			__antithesis_instrumentation__.Notify(251619)
			return explicitColNames[i] != desc.Sharded.Name == true
		}() == true {
			__antithesis_instrumentation__.Notify(251620)
			colIDs = append(colIDs, explicitColIDs[i])
		} else {
			__antithesis_instrumentation__.Notify(251621)
		}
	}
	__antithesis_instrumentation__.Notify(251617)
	return colIDs
}

func (desc *IndexDescriptor) IsValidReferencedUniqueConstraint(referencedColIDs ColumnIDs) bool {
	__antithesis_instrumentation__.Notify(251622)
	return desc.Unique && func() bool {
		__antithesis_instrumentation__.Notify(251623)
		return !desc.IsPartial() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(251624)
		return desc.explicitColumnIDsWithoutShardColumn().PermutationOf(referencedColIDs) == true
	}() == true
}

func (desc *IndexDescriptor) GetName() string {
	__antithesis_instrumentation__.Notify(251625)
	return desc.Name
}

func (desc *IndexDescriptor) InvertedColumnID() ColumnID {
	__antithesis_instrumentation__.Notify(251626)
	if desc.Type != IndexDescriptor_INVERTED {
		__antithesis_instrumentation__.Notify(251628)
		panic(errors.AssertionFailedf("index is not inverted"))
	} else {
		__antithesis_instrumentation__.Notify(251629)
	}
	__antithesis_instrumentation__.Notify(251627)
	return desc.KeyColumnIDs[len(desc.KeyColumnIDs)-1]
}

func (desc *IndexDescriptor) InvertedColumnName() string {
	__antithesis_instrumentation__.Notify(251630)
	if desc.Type != IndexDescriptor_INVERTED {
		__antithesis_instrumentation__.Notify(251632)
		panic(errors.AssertionFailedf("index is not inverted"))
	} else {
		__antithesis_instrumentation__.Notify(251633)
	}
	__antithesis_instrumentation__.Notify(251631)
	return desc.KeyColumnNames[len(desc.KeyColumnNames)-1]
}

func (desc *IndexDescriptor) InvertedColumnKeyType() *types.T {
	__antithesis_instrumentation__.Notify(251634)
	if desc.Type != IndexDescriptor_INVERTED {
		__antithesis_instrumentation__.Notify(251636)
		panic(errors.AssertionFailedf("index is not inverted"))
	} else {
		__antithesis_instrumentation__.Notify(251637)
	}
	__antithesis_instrumentation__.Notify(251635)
	return types.Bytes
}
