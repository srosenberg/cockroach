package span

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

type Splitter struct {
	numKeyColumns int

	neededFamilies []descpb.FamilyID
}

func NoopSplitter() Splitter {
	__antithesis_instrumentation__.Notify(623620)
	return Splitter{}
}

func MakeSplitter(
	table catalog.TableDescriptor, index catalog.Index, neededColOrdinals util.FastIntSet,
) Splitter {
	__antithesis_instrumentation__.Notify(623621)

	if catalog.IsSystemDescriptor(table) {
		__antithesis_instrumentation__.Notify(623627)
		return NoopSplitter()
	} else {
		__antithesis_instrumentation__.Notify(623628)
	}
	__antithesis_instrumentation__.Notify(623622)

	if !index.IsUnique() {
		__antithesis_instrumentation__.Notify(623629)
		return NoopSplitter()
	} else {
		__antithesis_instrumentation__.Notify(623630)
	}
	__antithesis_instrumentation__.Notify(623623)

	if index.GetID() != table.GetPrimaryIndexID() {
		__antithesis_instrumentation__.Notify(623631)

		if index.GetType() != descpb.IndexDescriptor_FORWARD {
			__antithesis_instrumentation__.Notify(623633)
			return NoopSplitter()
		} else {
			__antithesis_instrumentation__.Notify(623634)
		}
		__antithesis_instrumentation__.Notify(623632)

		if index.GetVersion() < descpb.SecondaryIndexFamilyFormatVersion {
			__antithesis_instrumentation__.Notify(623635)
			return NoopSplitter()
		} else {
			__antithesis_instrumentation__.Notify(623636)
		}
	} else {
		__antithesis_instrumentation__.Notify(623637)
	}
	__antithesis_instrumentation__.Notify(623624)

	neededFamilies := rowenc.NeededColumnFamilyIDs(neededColOrdinals, table, index)

	for i := range neededFamilies[1:] {
		__antithesis_instrumentation__.Notify(623638)
		if neededFamilies[i] >= neededFamilies[i+1] {
			__antithesis_instrumentation__.Notify(623639)
			panic(errors.AssertionFailedf("family IDs not increasing"))
		} else {
			__antithesis_instrumentation__.Notify(623640)
		}
	}
	__antithesis_instrumentation__.Notify(623625)

	numFamilies := len(table.GetFamilies())
	if numFamilies > 1 && func() bool {
		__antithesis_instrumentation__.Notify(623641)
		return len(neededFamilies) == numFamilies == true
	}() == true {
		__antithesis_instrumentation__.Notify(623642)
		return NoopSplitter()
	} else {
		__antithesis_instrumentation__.Notify(623643)
	}
	__antithesis_instrumentation__.Notify(623626)

	return Splitter{
		numKeyColumns:  index.NumKeyColumns(),
		neededFamilies: neededFamilies,
	}
}

func MakeSplitterWithFamilyIDs(numKeyColumns int, familyIDs []descpb.FamilyID) Splitter {
	__antithesis_instrumentation__.Notify(623644)
	if len(familyIDs) == 0 {
		__antithesis_instrumentation__.Notify(623647)
		return NoopSplitter()
	} else {
		__antithesis_instrumentation__.Notify(623648)
	}
	__antithesis_instrumentation__.Notify(623645)

	for i := range familyIDs[1:] {
		__antithesis_instrumentation__.Notify(623649)
		if familyIDs[i] >= familyIDs[i+1] {
			__antithesis_instrumentation__.Notify(623650)
			panic(errors.AssertionFailedf("family IDs not increasing"))
		} else {
			__antithesis_instrumentation__.Notify(623651)
		}
	}
	__antithesis_instrumentation__.Notify(623646)
	return Splitter{
		numKeyColumns:  numKeyColumns,
		neededFamilies: familyIDs,
	}
}

func (s *Splitter) FamilyIDs() []descpb.FamilyID {
	__antithesis_instrumentation__.Notify(623652)
	return s.neededFamilies
}

func (s *Splitter) IsNoop() bool {
	__antithesis_instrumentation__.Notify(623653)
	return s.numKeyColumns == 0
}

func (s *Splitter) MaybeSplitSpanIntoSeparateFamilies(
	appendTo roachpb.Spans, span roachpb.Span, prefixLen int, containsNull bool,
) roachpb.Spans {
	__antithesis_instrumentation__.Notify(623654)
	if s.CanSplitSpanIntoFamilySpans(prefixLen, containsNull) {
		__antithesis_instrumentation__.Notify(623656)
		return rowenc.SplitRowKeyIntoFamilySpans(appendTo, span.Key, s.neededFamilies)
	} else {
		__antithesis_instrumentation__.Notify(623657)
	}
	__antithesis_instrumentation__.Notify(623655)
	return append(appendTo, span)
}

func (s *Splitter) CanSplitSpanIntoFamilySpans(prefixLen int, containsNull bool) bool {
	__antithesis_instrumentation__.Notify(623658)
	if s.IsNoop() {
		__antithesis_instrumentation__.Notify(623662)

		return false
	} else {
		__antithesis_instrumentation__.Notify(623663)
	}
	__antithesis_instrumentation__.Notify(623659)

	if prefixLen != s.numKeyColumns {
		__antithesis_instrumentation__.Notify(623664)
		return false
	} else {
		__antithesis_instrumentation__.Notify(623665)
	}
	__antithesis_instrumentation__.Notify(623660)

	if containsNull {
		__antithesis_instrumentation__.Notify(623666)
		return false
	} else {
		__antithesis_instrumentation__.Notify(623667)
	}
	__antithesis_instrumentation__.Notify(623661)

	return true
}

func (s *Splitter) ExistenceCheckSpan(
	span roachpb.Span, prefixLen int, containsNull bool,
) roachpb.Span {
	__antithesis_instrumentation__.Notify(623668)
	if s.CanSplitSpanIntoFamilySpans(prefixLen, containsNull) {
		__antithesis_instrumentation__.Notify(623670)

		key := keys.MakeFamilyKey(span.Key, 0)
		return roachpb.Span{Key: key, EndKey: roachpb.Key(key).PrefixEnd()}
	} else {
		__antithesis_instrumentation__.Notify(623671)
	}
	__antithesis_instrumentation__.Notify(623669)
	return span
}
