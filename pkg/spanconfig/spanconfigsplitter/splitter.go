// Package spanconfigsplitter is able to split sql descriptors into its
// constituent spans.
package spanconfigsplitter

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/dustin/go-humanize"
)

var _ spanconfig.Splitter = &Splitter{}

type Splitter struct {
	codec keys.SQLCodec
	knobs *spanconfig.TestingKnobs
}

func New(codec keys.SQLCodec, knobs *spanconfig.TestingKnobs) *Splitter {
	__antithesis_instrumentation__.Notify(240961)
	if knobs == nil {
		__antithesis_instrumentation__.Notify(240963)
		knobs = &spanconfig.TestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(240964)
	}
	__antithesis_instrumentation__.Notify(240962)
	return &Splitter{
		codec: codec,
		knobs: knobs,
	}
}

func (s *Splitter) Splits(ctx context.Context, table catalog.TableDescriptor) (int, error) {
	__antithesis_instrumentation__.Notify(240965)
	if isNil(table) {
		__antithesis_instrumentation__.Notify(240970)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(240971)
	}
	__antithesis_instrumentation__.Notify(240966)

	if s.knobs.ExcludeDroppedDescriptorsFromLookup && func() bool {
		__antithesis_instrumentation__.Notify(240972)
		return table.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(240973)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(240974)
	}
	__antithesis_instrumentation__.Notify(240967)

	s.log(0, "+ %-2d between start of table and start of %s index", 1, humanize.Ordinal(1))
	numIndexes, innerSplitCount := 0, 0
	if err := catalog.ForEachIndex(table, catalog.IndexOpts{}, func(index catalog.Index) error {
		__antithesis_instrumentation__.Notify(240975)
		splitCount, err := s.splitsInner(ctx, partition{
			Partitioning: index.GetPartitioning(),
			table:        table,
			index:        index,
			level:        1,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(240977)
			return err
		} else {
			__antithesis_instrumentation__.Notify(240978)
		}
		__antithesis_instrumentation__.Notify(240976)

		innerSplitCount += splitCount
		numIndexes++

		s.log(0, "+ %-2d for %s index", splitCount, humanize.Ordinal(numIndexes))
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(240979)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(240980)
	}
	__antithesis_instrumentation__.Notify(240968)

	if numIndexes > 1 {
		__antithesis_instrumentation__.Notify(240981)
		s.log(0, "+ %-2d gap(s) between %d indexes", numIndexes-1, numIndexes)
	} else {
		__antithesis_instrumentation__.Notify(240982)
	}
	__antithesis_instrumentation__.Notify(240969)

	s.log(0, "+ %-2d between end of %s index and end of table", 1, humanize.Ordinal(numIndexes))
	return numIndexes + 1 + innerSplitCount, nil
}

func (s *Splitter) splitsInner(ctx context.Context, part partition) (int, error) {
	__antithesis_instrumentation__.Notify(240983)
	if part.NumColumns() == 0 {
		__antithesis_instrumentation__.Notify(240988)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(240989)
	}
	__antithesis_instrumentation__.Notify(240984)
	if part.NumRanges() > 0 {
		__antithesis_instrumentation__.Notify(240990)
		return part.NumRanges(), nil
	} else {
		__antithesis_instrumentation__.Notify(240991)
	}
	__antithesis_instrumentation__.Notify(240985)

	s.log(part.level, "+ %-2d between start of index and start of %s partition-by-list value", 1, humanize.Ordinal(1))
	numListValues, innerSplitCount := 0, 0
	if err := part.ForEachList(func(_ string, values [][]byte, subPartitioning catalog.Partitioning) error {
		__antithesis_instrumentation__.Notify(240992)
		for i := 0; i < len(values); i++ {
			__antithesis_instrumentation__.Notify(240994)
			subPartition := partition{
				Partitioning: subPartitioning,
				table:        part.table,
				index:        part.index,
				level:        part.level + 1,
			}
			splitCount, err := s.splitsInner(ctx, subPartition)
			if err != nil {
				__antithesis_instrumentation__.Notify(240996)
				return err
			} else {
				__antithesis_instrumentation__.Notify(240997)
			}
			__antithesis_instrumentation__.Notify(240995)

			innerSplitCount += splitCount
			numListValues++

			s.log(part.level, "+ %-2d for %s partition-by-list value", splitCount, humanize.Ordinal(numListValues))
		}
		__antithesis_instrumentation__.Notify(240993)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(240998)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(240999)
	}
	__antithesis_instrumentation__.Notify(240986)

	if numListValues > 1 {
		__antithesis_instrumentation__.Notify(241000)
		s.log(part.level, "+ %-2d gap(s) between %d partition-by-list value spans", numListValues-1, numListValues)
	} else {
		__antithesis_instrumentation__.Notify(241001)
	}
	__antithesis_instrumentation__.Notify(240987)

	s.log(part.level, "+ %-2d between end of %s partition-by-list value span and end of index", 1, humanize.Ordinal(numListValues))
	return numListValues + 1 + innerSplitCount, nil
}

func (s *Splitter) log(level int, format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(241002)
	padding := strings.Repeat("    ", level)
	if s.knobs.SplitterStepLogger != nil {
		__antithesis_instrumentation__.Notify(241003)
		s.knobs.SplitterStepLogger(padding + fmt.Sprintf(format, args...))
	} else {
		__antithesis_instrumentation__.Notify(241004)
	}
}

type partition struct {
	catalog.Partitioning

	table catalog.TableDescriptor
	index catalog.Index
	level int
}

func isNil(table catalog.TableDescriptor) bool {
	__antithesis_instrumentation__.Notify(241005)
	vTable := reflect.ValueOf(table)
	return vTable.Kind() == reflect.Ptr && func() bool {
		__antithesis_instrumentation__.Notify(241006)
		return vTable.IsNil() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(241007)
		return table == nil == true
	}() == true
}
