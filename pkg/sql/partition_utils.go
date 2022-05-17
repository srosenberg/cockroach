package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/covering"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

func GenerateSubzoneSpans(
	st *cluster.Settings,
	logicalClusterID uuid.UUID,
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
	subzones []zonepb.Subzone,
	hasNewSubzones bool,
) ([]zonepb.SubzoneSpan, error) {
	__antithesis_instrumentation__.Notify(557660)

	if hasNewSubzones {
		__antithesis_instrumentation__.Notify(557666)
		org := ClusterOrganization.Get(&st.SV)
		if err := base.CheckEnterpriseEnabled(st, logicalClusterID, org,
			"replication zones on indexes or partitions"); err != nil {
			__antithesis_instrumentation__.Notify(557667)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(557668)
		}
	} else {
		__antithesis_instrumentation__.Notify(557669)
	}
	__antithesis_instrumentation__.Notify(557661)

	if tableDesc.Dropped() {
		__antithesis_instrumentation__.Notify(557670)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(557671)
	}
	__antithesis_instrumentation__.Notify(557662)

	a := &tree.DatumAlloc{}

	subzoneIndexByIndexID := make(map[descpb.IndexID]int32)
	subzoneIndexByPartition := make(map[string]int32)
	for i, subzone := range subzones {
		__antithesis_instrumentation__.Notify(557672)
		if len(subzone.PartitionName) > 0 {
			__antithesis_instrumentation__.Notify(557673)
			subzoneIndexByPartition[subzone.PartitionName] = int32(i)
		} else {
			__antithesis_instrumentation__.Notify(557674)
			subzoneIndexByIndexID[descpb.IndexID(subzone.IndexID)] = int32(i)
		}
	}
	__antithesis_instrumentation__.Notify(557663)

	var indexCovering covering.Covering
	var partitionCoverings []covering.Covering
	if err := catalog.ForEachIndex(tableDesc, catalog.IndexOpts{
		AddMutations: true,
	}, func(idx catalog.Index) error {
		__antithesis_instrumentation__.Notify(557675)
		_, indexSubzoneExists := subzoneIndexByIndexID[idx.GetID()]
		if indexSubzoneExists {
			__antithesis_instrumentation__.Notify(557678)
			idxSpan := tableDesc.IndexSpan(codec, idx.GetID())

			indexCovering = append(indexCovering, covering.Range{
				Start: idxSpan.Key, End: idxSpan.EndKey,
				Payload: zonepb.Subzone{IndexID: uint32(idx.GetID())},
			})
		} else {
			__antithesis_instrumentation__.Notify(557679)
		}
		__antithesis_instrumentation__.Notify(557676)

		var emptyPrefix []tree.Datum
		indexPartitionCoverings, err := indexCoveringsForPartitioning(
			a, codec, tableDesc, idx, idx.GetPartitioning(), subzoneIndexByPartition, emptyPrefix)
		if err != nil {
			__antithesis_instrumentation__.Notify(557680)
			return err
		} else {
			__antithesis_instrumentation__.Notify(557681)
		}
		__antithesis_instrumentation__.Notify(557677)

		partitionCoverings = append(partitionCoverings, indexPartitionCoverings...)

		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(557682)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(557683)
	}
	__antithesis_instrumentation__.Notify(557664)

	ranges := covering.OverlapCoveringMerge(append(partitionCoverings, indexCovering))

	sharedPrefix := codec.TablePrefix(uint32(tableDesc.GetID()))

	var subzoneSpans []zonepb.SubzoneSpan
	for _, r := range ranges {
		__antithesis_instrumentation__.Notify(557684)
		payloads := r.Payload.([]interface{})
		if len(payloads) == 0 {
			__antithesis_instrumentation__.Notify(557689)
			continue
		} else {
			__antithesis_instrumentation__.Notify(557690)
		}
		__antithesis_instrumentation__.Notify(557685)
		subzoneSpan := zonepb.SubzoneSpan{
			Key:    bytes.TrimPrefix(r.Start, sharedPrefix),
			EndKey: bytes.TrimPrefix(r.End, sharedPrefix),
		}
		var ok bool
		if subzone := payloads[0].(zonepb.Subzone); len(subzone.PartitionName) > 0 {
			__antithesis_instrumentation__.Notify(557691)
			subzoneSpan.SubzoneIndex, ok = subzoneIndexByPartition[subzone.PartitionName]
		} else {
			__antithesis_instrumentation__.Notify(557692)
			subzoneSpan.SubzoneIndex, ok = subzoneIndexByIndexID[descpb.IndexID(subzone.IndexID)]
		}
		__antithesis_instrumentation__.Notify(557686)
		if !ok {
			__antithesis_instrumentation__.Notify(557693)
			continue
		} else {
			__antithesis_instrumentation__.Notify(557694)
		}
		__antithesis_instrumentation__.Notify(557687)
		if bytes.Equal(subzoneSpan.Key.PrefixEnd(), subzoneSpan.EndKey) {
			__antithesis_instrumentation__.Notify(557695)
			subzoneSpan.EndKey = nil
		} else {
			__antithesis_instrumentation__.Notify(557696)
		}
		__antithesis_instrumentation__.Notify(557688)
		subzoneSpans = append(subzoneSpans, subzoneSpan)
	}
	__antithesis_instrumentation__.Notify(557665)
	return subzoneSpans, nil
}

func indexCoveringsForPartitioning(
	a *tree.DatumAlloc,
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
	idx catalog.Index,
	part catalog.Partitioning,
	relevantPartitions map[string]int32,
	prefixDatums []tree.Datum,
) ([]covering.Covering, error) {
	__antithesis_instrumentation__.Notify(557697)
	if part.NumColumns() == 0 {
		__antithesis_instrumentation__.Notify(557701)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(557702)
	}
	__antithesis_instrumentation__.Notify(557698)

	var coverings []covering.Covering
	var descendentCoverings []covering.Covering

	if part.NumLists() > 0 {
		__antithesis_instrumentation__.Notify(557703)

		listCoverings := make([]covering.Covering, part.NumColumns()+1)
		err := part.ForEachList(func(name string, values [][]byte, subPartitioning catalog.Partitioning) error {
			__antithesis_instrumentation__.Notify(557706)
			for _, valueEncBuf := range values {
				__antithesis_instrumentation__.Notify(557708)
				t, keyPrefix, err := rowenc.DecodePartitionTuple(
					a, codec, tableDesc, idx, part, valueEncBuf, prefixDatums)
				if err != nil {
					__antithesis_instrumentation__.Notify(557712)
					return err
				} else {
					__antithesis_instrumentation__.Notify(557713)
				}
				__antithesis_instrumentation__.Notify(557709)
				if _, ok := relevantPartitions[name]; ok {
					__antithesis_instrumentation__.Notify(557714)
					listCoverings[len(t.Datums)] = append(listCoverings[len(t.Datums)], covering.Range{
						Start: keyPrefix, End: roachpb.Key(keyPrefix).PrefixEnd(),
						Payload: zonepb.Subzone{PartitionName: name},
					})
				} else {
					__antithesis_instrumentation__.Notify(557715)
				}
				__antithesis_instrumentation__.Notify(557710)
				newPrefixDatums := append(prefixDatums, t.Datums...)
				subpartitionCoverings, err := indexCoveringsForPartitioning(
					a, codec, tableDesc, idx, subPartitioning, relevantPartitions, newPrefixDatums)
				if err != nil {
					__antithesis_instrumentation__.Notify(557716)
					return err
				} else {
					__antithesis_instrumentation__.Notify(557717)
				}
				__antithesis_instrumentation__.Notify(557711)
				descendentCoverings = append(descendentCoverings, subpartitionCoverings...)
			}
			__antithesis_instrumentation__.Notify(557707)
			return nil
		})
		__antithesis_instrumentation__.Notify(557704)
		if err != nil {
			__antithesis_instrumentation__.Notify(557718)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(557719)
		}
		__antithesis_instrumentation__.Notify(557705)
		for i := range listCoverings {
			__antithesis_instrumentation__.Notify(557720)
			if covering := listCoverings[len(listCoverings)-i-1]; len(covering) > 0 {
				__antithesis_instrumentation__.Notify(557721)
				coverings = append(coverings, covering)
			} else {
				__antithesis_instrumentation__.Notify(557722)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(557723)
	}
	__antithesis_instrumentation__.Notify(557699)

	if part.NumRanges() > 0 {
		__antithesis_instrumentation__.Notify(557724)
		err := part.ForEachRange(func(name string, from, to []byte) error {
			__antithesis_instrumentation__.Notify(557726)
			if _, ok := relevantPartitions[name]; !ok {
				__antithesis_instrumentation__.Notify(557731)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(557732)
			}
			__antithesis_instrumentation__.Notify(557727)
			_, fromKey, err := rowenc.DecodePartitionTuple(
				a, codec, tableDesc, idx, part, from, prefixDatums)
			if err != nil {
				__antithesis_instrumentation__.Notify(557733)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557734)
			}
			__antithesis_instrumentation__.Notify(557728)
			_, toKey, err := rowenc.DecodePartitionTuple(
				a, codec, tableDesc, idx, part, to, prefixDatums)
			if err != nil {
				__antithesis_instrumentation__.Notify(557735)
				return err
			} else {
				__antithesis_instrumentation__.Notify(557736)
			}
			__antithesis_instrumentation__.Notify(557729)
			if _, ok := relevantPartitions[name]; ok {
				__antithesis_instrumentation__.Notify(557737)
				coverings = append(coverings, covering.Covering{{
					Start: fromKey, End: toKey,
					Payload: zonepb.Subzone{PartitionName: name},
				}})
			} else {
				__antithesis_instrumentation__.Notify(557738)
			}
			__antithesis_instrumentation__.Notify(557730)
			return nil
		})
		__antithesis_instrumentation__.Notify(557725)
		if err != nil {
			__antithesis_instrumentation__.Notify(557739)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(557740)
		}
	} else {
		__antithesis_instrumentation__.Notify(557741)
	}
	__antithesis_instrumentation__.Notify(557700)

	return append(descendentCoverings, coverings...), nil
}
