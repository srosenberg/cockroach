package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util"
)

type PartialIndexUpdateHelper struct {
	IgnoreForPut util.FastIntSet

	IgnoreForDel util.FastIntSet
}

func (pm *PartialIndexUpdateHelper) Init(
	partialIndexPutVals tree.Datums, partialIndexDelVals tree.Datums, tabDesc catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(568483)
	colIdx := 0

	for _, idx := range tabDesc.PartialIndexes() {
		__antithesis_instrumentation__.Notify(568485)

		if colIdx < len(partialIndexPutVals) {
			__antithesis_instrumentation__.Notify(568488)
			val, err := tree.GetBool(partialIndexPutVals[colIdx])
			if err != nil {
				__antithesis_instrumentation__.Notify(568490)
				return err
			} else {
				__antithesis_instrumentation__.Notify(568491)
			}
			__antithesis_instrumentation__.Notify(568489)
			if !val {
				__antithesis_instrumentation__.Notify(568492)

				pm.IgnoreForPut.Add(int(idx.GetID()))
			} else {
				__antithesis_instrumentation__.Notify(568493)
			}
		} else {
			__antithesis_instrumentation__.Notify(568494)
		}
		__antithesis_instrumentation__.Notify(568486)

		if colIdx < len(partialIndexDelVals) {
			__antithesis_instrumentation__.Notify(568495)
			val, err := tree.GetBool(partialIndexDelVals[colIdx])
			if err != nil {
				__antithesis_instrumentation__.Notify(568497)
				return err
			} else {
				__antithesis_instrumentation__.Notify(568498)
			}
			__antithesis_instrumentation__.Notify(568496)
			if !val {
				__antithesis_instrumentation__.Notify(568499)

				pm.IgnoreForDel.Add(int(idx.GetID()))
			} else {
				__antithesis_instrumentation__.Notify(568500)
			}
		} else {
			__antithesis_instrumentation__.Notify(568501)
		}
		__antithesis_instrumentation__.Notify(568487)

		colIdx++
	}
	__antithesis_instrumentation__.Notify(568484)

	return nil
}
