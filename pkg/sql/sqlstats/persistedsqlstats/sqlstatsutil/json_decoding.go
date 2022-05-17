package sqlstatsutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/errors"
)

func DecodeTxnStatsMetadataJSON(
	metadata json.JSON, result *roachpb.CollectedTransactionStatistics,
) error {
	__antithesis_instrumentation__.Notify(625029)
	return jsonFields{
		{"stmtFingerprintIDs", (*stmtFingerprintIDArray)(&result.StatementFingerprintIDs)},
	}.decodeJSON(metadata)
}

func DecodeTxnStatsStatisticsJSON(jsonVal json.JSON, result *roachpb.TransactionStatistics) error {
	__antithesis_instrumentation__.Notify(625030)
	return (*txnStats)(result).decodeJSON(jsonVal)
}

func DecodeStmtStatsMetadataJSON(
	metadata json.JSON, result *roachpb.CollectedStatementStatistics,
) error {
	__antithesis_instrumentation__.Notify(625031)
	return (*stmtStatsMetadata)(result).jsonFields().decodeJSON(metadata)
}

func DecodeAggregatedMetadataJSON(
	metadata json.JSON, result *roachpb.AggregatedStatementMetadata,
) error {
	__antithesis_instrumentation__.Notify(625032)
	return (*aggregatedMetadata)(result).jsonFields().decodeJSON(metadata)
}

func DecodeStmtStatsStatisticsJSON(jsonVal json.JSON, result *roachpb.StatementStatistics) error {
	__antithesis_instrumentation__.Notify(625033)
	return (*stmtStats)(result).decodeJSON(jsonVal)
}

func JSONToExplainTreePlanNode(jsonVal json.JSON) (*roachpb.ExplainTreePlanNode, error) {
	__antithesis_instrumentation__.Notify(625034)
	node := roachpb.ExplainTreePlanNode{}

	nameAttr, err := jsonVal.FetchValKey("Name")
	if err != nil {
		__antithesis_instrumentation__.Notify(625040)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625041)
	}
	__antithesis_instrumentation__.Notify(625035)

	if nameAttr != nil {
		__antithesis_instrumentation__.Notify(625042)
		str, err := nameAttr.AsText()
		if err != nil {
			__antithesis_instrumentation__.Notify(625044)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(625045)
		}
		__antithesis_instrumentation__.Notify(625043)
		node.Name = *str
	} else {
		__antithesis_instrumentation__.Notify(625046)
	}
	__antithesis_instrumentation__.Notify(625036)

	iter, err := jsonVal.ObjectIter()
	if err != nil {
		__antithesis_instrumentation__.Notify(625047)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625048)
	}
	__antithesis_instrumentation__.Notify(625037)

	if iter == nil {
		__antithesis_instrumentation__.Notify(625049)
		return nil, errors.New("unable to deconstruct json object")
	} else {
		__antithesis_instrumentation__.Notify(625050)
	}
	__antithesis_instrumentation__.Notify(625038)

	for iter.Next() {
		__antithesis_instrumentation__.Notify(625051)
		key := iter.Key()
		value := iter.Value()

		if key == "Name" {
			__antithesis_instrumentation__.Notify(625053)

			continue
		} else {
			__antithesis_instrumentation__.Notify(625054)
		}
		__antithesis_instrumentation__.Notify(625052)

		if key == "Children" {
			__antithesis_instrumentation__.Notify(625055)
			for childIdx := 0; childIdx < value.Len(); childIdx++ {
				__antithesis_instrumentation__.Notify(625056)
				childJSON, err := value.FetchValIdx(childIdx)
				if err != nil {
					__antithesis_instrumentation__.Notify(625058)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(625059)
				}
				__antithesis_instrumentation__.Notify(625057)
				if childJSON != nil {
					__antithesis_instrumentation__.Notify(625060)
					child, err := JSONToExplainTreePlanNode(childJSON)
					if err != nil {
						__antithesis_instrumentation__.Notify(625062)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(625063)
					}
					__antithesis_instrumentation__.Notify(625061)
					node.Children = append(node.Children, child)
				} else {
					__antithesis_instrumentation__.Notify(625064)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(625065)
			str, err := value.AsText()
			if err != nil {
				__antithesis_instrumentation__.Notify(625067)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(625068)
			}
			__antithesis_instrumentation__.Notify(625066)
			node.Attrs = append(node.Attrs, &roachpb.ExplainTreePlanNode_Attr{
				Key:   key,
				Value: *str,
			})
		}
	}
	__antithesis_instrumentation__.Notify(625039)

	return &node, nil
}
