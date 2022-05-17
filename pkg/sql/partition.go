package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func partitionByFromTableDesc(
	codec keys.SQLCodec, tableDesc *tabledesc.Mutable,
) (*tree.PartitionBy, error) {
	__antithesis_instrumentation__.Notify(557615)
	idx := tableDesc.GetPrimaryIndex()
	return partitionByFromTableDescImpl(codec, tableDesc, idx, idx.GetPartitioning(), 0)
}

func partitionByFromTableDescImpl(
	codec keys.SQLCodec,
	tableDesc *tabledesc.Mutable,
	idx catalog.Index,
	part catalog.Partitioning,
	colOffset int,
) (*tree.PartitionBy, error) {
	__antithesis_instrumentation__.Notify(557616)
	if part.NumColumns() == 0 {
		__antithesis_instrumentation__.Notify(557624)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(557625)
	}
	__antithesis_instrumentation__.Notify(557617)

	fakePrefixDatums := make([]tree.Datum, colOffset)
	for i := range fakePrefixDatums {
		__antithesis_instrumentation__.Notify(557626)
		fakePrefixDatums[i] = tree.DNull
	}
	__antithesis_instrumentation__.Notify(557618)

	partitionBy := &tree.PartitionBy{
		Fields: make(tree.NameList, part.NumColumns()),
		List:   make([]tree.ListPartition, 0, part.NumLists()),
		Range:  make([]tree.RangePartition, 0, part.NumRanges()),
	}
	for i := 0; i < part.NumColumns(); i++ {
		__antithesis_instrumentation__.Notify(557627)
		partitionBy.Fields[i] = tree.Name(idx.GetKeyColumnName(colOffset + i))
	}
	__antithesis_instrumentation__.Notify(557619)

	a := &tree.DatumAlloc{}
	err := part.ForEachList(func(name string, values [][]byte, subPartitioning catalog.Partitioning) (err error) {
		__antithesis_instrumentation__.Notify(557628)
		lp := tree.ListPartition{
			Name:  tree.UnrestrictedName(name),
			Exprs: make(tree.Exprs, len(values)),
		}
		for j, values := range values {
			__antithesis_instrumentation__.Notify(557630)
			tuple, _, err := rowenc.DecodePartitionTuple(
				a,
				codec,
				tableDesc,
				idx,
				part,
				values,
				fakePrefixDatums,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(557633)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557634)
			}
			__antithesis_instrumentation__.Notify(557631)
			exprs, err := partitionTupleToExprs(tuple)
			if err != nil {
				__antithesis_instrumentation__.Notify(557635)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557636)
			}
			__antithesis_instrumentation__.Notify(557632)
			lp.Exprs[j] = &tree.Tuple{
				Exprs: exprs,
			}
		}
		__antithesis_instrumentation__.Notify(557629)
		lp.Subpartition, err = partitionByFromTableDescImpl(
			codec,
			tableDesc,
			idx,
			subPartitioning,
			colOffset+part.NumColumns(),
		)
		partitionBy.List = append(partitionBy.List, lp)
		return err
	})
	__antithesis_instrumentation__.Notify(557620)
	if err != nil {
		__antithesis_instrumentation__.Notify(557637)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(557638)
	}
	__antithesis_instrumentation__.Notify(557621)

	err = part.ForEachRange(func(name string, from, to []byte) error {
		__antithesis_instrumentation__.Notify(557639)
		rp := tree.RangePartition{Name: tree.UnrestrictedName(name)}
		fromTuple, _, err := rowenc.DecodePartitionTuple(
			a, codec, tableDesc, idx, part, from, fakePrefixDatums)
		if err != nil {
			__antithesis_instrumentation__.Notify(557643)
			return err
		} else {
			__antithesis_instrumentation__.Notify(557644)
		}
		__antithesis_instrumentation__.Notify(557640)
		rp.From, err = partitionTupleToExprs(fromTuple)
		if err != nil {
			__antithesis_instrumentation__.Notify(557645)
			return err
		} else {
			__antithesis_instrumentation__.Notify(557646)
		}
		__antithesis_instrumentation__.Notify(557641)
		toTuple, _, err := rowenc.DecodePartitionTuple(
			a, codec, tableDesc, idx, part, to, fakePrefixDatums)
		if err != nil {
			__antithesis_instrumentation__.Notify(557647)
			return err
		} else {
			__antithesis_instrumentation__.Notify(557648)
		}
		__antithesis_instrumentation__.Notify(557642)
		rp.To, err = partitionTupleToExprs(toTuple)
		partitionBy.Range = append(partitionBy.Range, rp)
		return err
	})
	__antithesis_instrumentation__.Notify(557622)
	if err != nil {
		__antithesis_instrumentation__.Notify(557649)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(557650)
	}
	__antithesis_instrumentation__.Notify(557623)

	return partitionBy, nil
}

func partitionTupleToExprs(t *rowenc.PartitionTuple) (tree.Exprs, error) {
	__antithesis_instrumentation__.Notify(557651)
	exprs := make(tree.Exprs, len(t.Datums)+t.SpecialCount)
	for i, d := range t.Datums {
		__antithesis_instrumentation__.Notify(557654)
		exprs[i] = d
	}
	__antithesis_instrumentation__.Notify(557652)
	for i := 0; i < t.SpecialCount; i++ {
		__antithesis_instrumentation__.Notify(557655)
		switch t.Special {
		case rowenc.PartitionDefaultVal:
			__antithesis_instrumentation__.Notify(557656)
			exprs[i+len(t.Datums)] = &tree.DefaultVal{}
		case rowenc.PartitionMinVal:
			__antithesis_instrumentation__.Notify(557657)
			exprs[i+len(t.Datums)] = &tree.PartitionMinVal{}
		case rowenc.PartitionMaxVal:
			__antithesis_instrumentation__.Notify(557658)
			exprs[i+len(t.Datums)] = &tree.PartitionMaxVal{}
		default:
			__antithesis_instrumentation__.Notify(557659)
			return nil, errors.AssertionFailedf("unknown special value found: %v", t.Special)
		}
	}
	__antithesis_instrumentation__.Notify(557653)
	return exprs, nil
}
