package rowenc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
)

type PartitionSpecialValCode uint64

const (
	PartitionDefaultVal PartitionSpecialValCode = 0

	PartitionMaxVal PartitionSpecialValCode = 1

	PartitionMinVal PartitionSpecialValCode = 2
)

func (c PartitionSpecialValCode) String() string {
	__antithesis_instrumentation__.Notify(570904)
	switch c {
	case PartitionDefaultVal:
		__antithesis_instrumentation__.Notify(570906)
		return (tree.DefaultVal{}).String()
	case PartitionMinVal:
		__antithesis_instrumentation__.Notify(570907)
		return (tree.PartitionMinVal{}).String()
	case PartitionMaxVal:
		__antithesis_instrumentation__.Notify(570908)
		return (tree.PartitionMaxVal{}).String()
	default:
		__antithesis_instrumentation__.Notify(570909)
	}
	__antithesis_instrumentation__.Notify(570905)
	panic("unreachable")
}

type PartitionTuple struct {
	Datums       tree.Datums
	Special      PartitionSpecialValCode
	SpecialCount int
}

func (t *PartitionTuple) String() string {
	__antithesis_instrumentation__.Notify(570910)
	f := tree.NewFmtCtx(tree.FmtSimple)
	f.WriteByte('(')
	for i := 0; i < len(t.Datums)+t.SpecialCount; i++ {
		__antithesis_instrumentation__.Notify(570912)
		if i > 0 {
			__antithesis_instrumentation__.Notify(570914)
			f.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(570915)
		}
		__antithesis_instrumentation__.Notify(570913)
		if i < len(t.Datums) {
			__antithesis_instrumentation__.Notify(570916)
			f.FormatNode(t.Datums[i])
		} else {
			__antithesis_instrumentation__.Notify(570917)
			f.WriteString(t.Special.String())
		}
	}
	__antithesis_instrumentation__.Notify(570911)
	f.WriteByte(')')
	return f.CloseAndGetString()
}

func DecodePartitionTuple(
	a *tree.DatumAlloc,
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
	index catalog.Index,
	part catalog.Partitioning,
	valueEncBuf []byte,
	prefixDatums tree.Datums,
) (*PartitionTuple, []byte, error) {
	__antithesis_instrumentation__.Notify(570918)
	if len(prefixDatums)+part.NumColumns() > index.NumKeyColumns() {
		__antithesis_instrumentation__.Notify(570926)
		return nil, nil, fmt.Errorf("not enough columns in index for this partitioning")
	} else {
		__antithesis_instrumentation__.Notify(570927)
	}
	__antithesis_instrumentation__.Notify(570919)

	t := &PartitionTuple{
		Datums: make(tree.Datums, 0, part.NumColumns()),
	}

	for i := len(prefixDatums); i < index.NumKeyColumns() && func() bool {
		__antithesis_instrumentation__.Notify(570928)
		return i < len(prefixDatums)+part.NumColumns() == true
	}() == true; i++ {
		__antithesis_instrumentation__.Notify(570929)
		colID := index.GetKeyColumnID(i)
		col, err := tableDesc.FindColumnWithID(colID)
		if err != nil {
			__antithesis_instrumentation__.Notify(570931)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(570932)
		}
		__antithesis_instrumentation__.Notify(570930)
		if _, dataOffset, _, typ, err := encoding.DecodeValueTag(valueEncBuf); err != nil {
			__antithesis_instrumentation__.Notify(570933)
			return nil, nil, errors.Wrapf(err, "decoding")
		} else {
			__antithesis_instrumentation__.Notify(570934)
			if typ == encoding.NotNull {
				__antithesis_instrumentation__.Notify(570935)

				var valCode uint64
				valueEncBuf, _, valCode, err = encoding.DecodeNonsortingUvarint(valueEncBuf[dataOffset:])
				if err != nil {
					__antithesis_instrumentation__.Notify(570938)
					return nil, nil, err
				} else {
					__antithesis_instrumentation__.Notify(570939)
				}
				__antithesis_instrumentation__.Notify(570936)
				nextSpecial := PartitionSpecialValCode(valCode)
				if t.SpecialCount > 0 && func() bool {
					__antithesis_instrumentation__.Notify(570940)
					return t.Special != nextSpecial == true
				}() == true {
					__antithesis_instrumentation__.Notify(570941)
					return nil, nil, errors.Newf("non-%[1]s value (%[2]s) not allowed after %[1]s",
						t.Special, nextSpecial)
				} else {
					__antithesis_instrumentation__.Notify(570942)
				}
				__antithesis_instrumentation__.Notify(570937)
				t.Special = nextSpecial
				t.SpecialCount++
			} else {
				__antithesis_instrumentation__.Notify(570943)
				var datum tree.Datum
				datum, valueEncBuf, err = valueside.Decode(a, col.GetType(), valueEncBuf)
				if err != nil {
					__antithesis_instrumentation__.Notify(570946)
					return nil, nil, errors.Wrapf(err, "decoding")
				} else {
					__antithesis_instrumentation__.Notify(570947)
				}
				__antithesis_instrumentation__.Notify(570944)
				if t.SpecialCount > 0 {
					__antithesis_instrumentation__.Notify(570948)
					return nil, nil, errors.Newf("non-%[1]s value (%[2]s) not allowed after %[1]s",
						t.Special, datum)
				} else {
					__antithesis_instrumentation__.Notify(570949)
				}
				__antithesis_instrumentation__.Notify(570945)
				t.Datums = append(t.Datums, datum)
			}
		}
	}
	__antithesis_instrumentation__.Notify(570920)
	if len(valueEncBuf) > 0 {
		__antithesis_instrumentation__.Notify(570950)
		return nil, nil, errors.New("superfluous data in encoded value")
	} else {
		__antithesis_instrumentation__.Notify(570951)
	}
	__antithesis_instrumentation__.Notify(570921)

	allDatums := append(prefixDatums, t.Datums...)
	var colMap catalog.TableColMap
	for i := range allDatums {
		__antithesis_instrumentation__.Notify(570952)
		colMap.Set(index.GetKeyColumnID(i), i)
	}
	__antithesis_instrumentation__.Notify(570922)

	indexKeyPrefix := MakeIndexKeyPrefix(codec, tableDesc.GetID(), index.GetID())
	keyAndSuffixCols := tableDesc.IndexFetchSpecKeyAndSuffixColumns(index)
	if len(allDatums) > len(keyAndSuffixCols) {
		__antithesis_instrumentation__.Notify(570953)
		return nil, nil, errors.Errorf("encoding too many columns (%d)", len(allDatums))
	} else {
		__antithesis_instrumentation__.Notify(570954)
	}
	__antithesis_instrumentation__.Notify(570923)
	key, _, err := EncodePartialIndexKey(keyAndSuffixCols[:len(allDatums)], colMap, allDatums, indexKeyPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(570955)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(570956)
	}
	__antithesis_instrumentation__.Notify(570924)

	if t.SpecialCount > 0 && func() bool {
		__antithesis_instrumentation__.Notify(570957)
		return t.Special == PartitionMaxVal == true
	}() == true {
		__antithesis_instrumentation__.Notify(570958)
		key = roachpb.Key(key).PrefixEnd()
	} else {
		__antithesis_instrumentation__.Notify(570959)
	}
	__antithesis_instrumentation__.Notify(570925)

	return t, key, nil
}
