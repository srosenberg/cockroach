// Package migration captures the facilities needed to define and execute
// migrations for a crdb cluster. These migrations can be arbitrarily long
// running, are free to send out arbitrary requests cluster wide, change
// internal DB state, and much more. They're typically reserved for crdb
// internal operations and state. Each migration is idempotent in nature, is
// associated with a specific cluster version, and executed when the cluster
// version is made active on every node in the cluster.
//
// Examples of migrations that apply would be migrations to move all raft state
// from one storage engine to another, or purging all usage of the replicated
// truncated state in KV.
package migration

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
)

type Migration interface {
	ClusterVersion() clusterversion.ClusterVersion
	Name() string
	internal()
}

type JobDeps interface {
	GetMigration(key clusterversion.ClusterVersion) (Migration, bool)

	SystemDeps() SystemDeps
}

type migration struct {
	description string
	cv          clusterversion.ClusterVersion
}

func (m *migration) ClusterVersion() clusterversion.ClusterVersion {
	__antithesis_instrumentation__.Notify(128146)
	return m.cv
}

func (m *migration) Name() string {
	__antithesis_instrumentation__.Notify(128147)
	return fmt.Sprintf("Migration to %s: %q", m.cv.String(), m.description)
}

func (m *migration) internal() { __antithesis_instrumentation__.Notify(128148) }
