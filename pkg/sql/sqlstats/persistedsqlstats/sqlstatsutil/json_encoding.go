package sqlstatsutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/hex"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/json"
)

func ExplainTreePlanNodeToJSON(node *roachpb.ExplainTreePlanNode) json.JSON {
	__antithesis_instrumentation__.Notify(625069)

	nodePlan := json.NewObjectBuilder(len(node.Attrs) + 2)
	nodeChildren := json.NewArrayBuilder(len(node.Children))

	nodePlan.Add("Name", json.FromString(node.Name))

	for _, attr := range node.Attrs {
		__antithesis_instrumentation__.Notify(625072)
		nodePlan.Add(strings.Title(attr.Key), json.FromString(attr.Value))
	}
	__antithesis_instrumentation__.Notify(625070)

	for _, childNode := range node.Children {
		__antithesis_instrumentation__.Notify(625073)
		nodeChildren.Add(ExplainTreePlanNodeToJSON(childNode))
	}
	__antithesis_instrumentation__.Notify(625071)

	nodePlan.Add("Children", nodeChildren.Build())
	return nodePlan.Build()
}

func BuildStmtMetadataJSON(statistics *roachpb.CollectedStatementStatistics) (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625074)
	return (*stmtStatsMetadata)(statistics).jsonFields().encodeJSON()
}

func BuildStmtStatisticsJSON(statistics *roachpb.StatementStatistics) (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625075)
	return (*stmtStats)(statistics).encodeJSON()
}

func BuildTxnMetadataJSON(statistics *roachpb.CollectedTransactionStatistics) (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625076)
	return jsonFields{
		{"stmtFingerprintIDs", (*stmtFingerprintIDArray)(&statistics.StatementFingerprintIDs)},
	}.encodeJSON()
}

func BuildTxnStatisticsJSON(statistics *roachpb.CollectedTransactionStatistics) (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625077)
	return (*txnStats)(&statistics.Stats).encodeJSON()
}

func BuildStmtDetailsMetadataJSON(
	metadata *roachpb.AggregatedStatementMetadata,
) (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625078)
	return (*aggregatedMetadata)(metadata).jsonFields().encodeJSON()
}

func EncodeUint64ToBytes(id uint64) []byte {
	__antithesis_instrumentation__.Notify(625079)
	result := make([]byte, 0, 8)
	return encoding.EncodeUint64Ascending(result, id)
}

func encodeStmtFingerprintIDToString(id roachpb.StmtFingerprintID) string {
	__antithesis_instrumentation__.Notify(625080)
	return hex.EncodeToString(EncodeUint64ToBytes(uint64(id)))
}
