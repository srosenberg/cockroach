package registry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
)

type TestSpec struct {
	Skip string

	SkipDetails string

	Name string

	Owner Owner

	Timeout time.Duration

	Tags []string

	Cluster spec.ClusterSpec

	UseIOBarrier bool

	NonReleaseBlocker bool

	RequiresLicense bool

	EncryptAtRandom bool

	Run func(ctx context.Context, t test.Test, c cluster.Cluster)
}

func (t *TestSpec) MatchOrSkip(filter *TestFilter) bool {
	__antithesis_instrumentation__.Notify(44381)
	if !filter.Name.MatchString(t.Name) {
		__antithesis_instrumentation__.Notify(44385)
		return false
	} else {
		__antithesis_instrumentation__.Notify(44386)
	}
	__antithesis_instrumentation__.Notify(44382)
	if len(t.Tags) == 0 {
		__antithesis_instrumentation__.Notify(44387)
		if !filter.Tag.MatchString("default") {
			__antithesis_instrumentation__.Notify(44389)
			t.Skip = fmt.Sprintf("%s does not match [default]", filter.RawTag)
		} else {
			__antithesis_instrumentation__.Notify(44390)
		}
		__antithesis_instrumentation__.Notify(44388)
		return true
	} else {
		__antithesis_instrumentation__.Notify(44391)
	}
	__antithesis_instrumentation__.Notify(44383)
	for _, t := range t.Tags {
		__antithesis_instrumentation__.Notify(44392)
		if filter.Tag.MatchString(t) {
			__antithesis_instrumentation__.Notify(44393)
			return true
		} else {
			__antithesis_instrumentation__.Notify(44394)
		}
	}
	__antithesis_instrumentation__.Notify(44384)
	t.Skip = fmt.Sprintf("%s does not match %s", filter.RawTag, t.Tags)
	return true
}
