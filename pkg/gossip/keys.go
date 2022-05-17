package gossip

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
)

const separator = ":"

const (
	KeyClusterID = "cluster-id"

	KeyStorePrefix = "store"

	KeyNodeIDPrefix = "node"

	KeyNodeHealthAlertPrefix = "health-alert"

	KeyNodeLivenessPrefix = "liveness"

	KeySentinel = "sentinel"

	KeyFirstRangeDescriptor = "first-range"

	KeyDeprecatedSystemConfig = "system-db"

	KeyDistSQLNodeVersionKeyPrefix = "distsql-version"

	KeyDistSQLDrainingPrefix = "distsql-draining"

	KeyGossipClientsPrefix = "gossip-clients"

	KeyGossipStatementDiagnosticsRequest = "stmt-diag-req"

	KeyGossipStatementDiagnosticsRequestCancellation = "stmt-diag-cancel-req"
)

func MakeKey(components ...string) string {
	__antithesis_instrumentation__.Notify(67958)
	return strings.Join(components, separator)
}

func MakePrefixPattern(prefix string) string {
	__antithesis_instrumentation__.Notify(67959)
	return regexp.QuoteMeta(prefix+separator) + ".*"
}

func MakeNodeIDKey(nodeID roachpb.NodeID) string {
	__antithesis_instrumentation__.Notify(67960)
	return MakeKey(KeyNodeIDPrefix, nodeID.String())
}

func IsNodeIDKey(key string) bool {
	__antithesis_instrumentation__.Notify(67961)
	return strings.HasPrefix(key, KeyNodeIDPrefix+separator)
}

func NodeIDFromKey(key string, prefix string) (roachpb.NodeID, error) {
	__antithesis_instrumentation__.Notify(67962)
	trimmedKey, err := removePrefixFromKey(key, prefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(67965)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(67966)
	}
	__antithesis_instrumentation__.Notify(67963)
	nodeID, err := strconv.ParseInt(trimmedKey, 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(67967)
		return 0, errors.Wrapf(err, "failed parsing NodeID from key %q", key)
	} else {
		__antithesis_instrumentation__.Notify(67968)
	}
	__antithesis_instrumentation__.Notify(67964)
	return roachpb.NodeID(nodeID), nil
}

func MakeGossipClientsKey(nodeID roachpb.NodeID) string {
	__antithesis_instrumentation__.Notify(67969)
	return MakeKey(KeyGossipClientsPrefix, nodeID.String())
}

func MakeNodeHealthAlertKey(nodeID roachpb.NodeID) string {
	__antithesis_instrumentation__.Notify(67970)
	return MakeKey(KeyNodeHealthAlertPrefix, strconv.Itoa(int(nodeID)))
}

func MakeNodeLivenessKey(nodeID roachpb.NodeID) string {
	__antithesis_instrumentation__.Notify(67971)
	return MakeKey(KeyNodeLivenessPrefix, nodeID.String())
}

func MakeStoreKey(storeID roachpb.StoreID) string {
	__antithesis_instrumentation__.Notify(67972)
	return MakeKey(KeyStorePrefix, storeID.String())
}

func StoreIDFromKey(storeKey string) (roachpb.StoreID, error) {
	__antithesis_instrumentation__.Notify(67973)
	trimmedKey, err := removePrefixFromKey(storeKey, KeyStorePrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(67976)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(67977)
	}
	__antithesis_instrumentation__.Notify(67974)
	storeID, err := strconv.ParseInt(trimmedKey, 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(67978)
		return 0, errors.Wrapf(err, "failed parsing StoreID from key %q", storeKey)
	} else {
		__antithesis_instrumentation__.Notify(67979)
	}
	__antithesis_instrumentation__.Notify(67975)
	return roachpb.StoreID(storeID), nil
}

func MakeDistSQLNodeVersionKey(instanceID base.SQLInstanceID) string {
	__antithesis_instrumentation__.Notify(67980)
	return MakeKey(KeyDistSQLNodeVersionKeyPrefix, instanceID.String())
}

func MakeDistSQLDrainingKey(instanceID base.SQLInstanceID) string {
	__antithesis_instrumentation__.Notify(67981)
	return MakeKey(KeyDistSQLDrainingPrefix, instanceID.String())
}

func removePrefixFromKey(key, prefix string) (string, error) {
	__antithesis_instrumentation__.Notify(67982)
	trimmedKey := strings.TrimPrefix(key, prefix+separator)
	if trimmedKey == key {
		__antithesis_instrumentation__.Notify(67984)
		return "", errors.Errorf("%q does not have expected prefix %q%s", key, prefix, separator)
	} else {
		__antithesis_instrumentation__.Notify(67985)
	}
	__antithesis_instrumentation__.Notify(67983)
	return trimmedKey, nil
}
