package migration

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
)

type TestingKnobs struct {
	ListBetweenOverride func(from, to clusterversion.ClusterVersion) []clusterversion.ClusterVersion

	RegistryOverride func(cv clusterversion.ClusterVersion) (Migration, bool)
}

func (t *TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(128756) }

var _ base.ModuleTestingKnobs = (*TestingKnobs)(nil)
