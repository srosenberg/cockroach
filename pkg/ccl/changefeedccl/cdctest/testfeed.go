package cdctest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
)

type TestFeedFactory interface {
	Feed(create string, args ...interface{}) (TestFeed, error)

	Server() serverutils.TestServerInterface
}

type TestFeedMessage struct {
	Topic, Partition string
	Key, Value       []byte
	Resolved         []byte
}

func (m TestFeedMessage) String() string {
	__antithesis_instrumentation__.Notify(15004)
	if m.Resolved != nil {
		__antithesis_instrumentation__.Notify(15006)
		return string(m.Resolved)
	} else {
		__antithesis_instrumentation__.Notify(15007)
	}
	__antithesis_instrumentation__.Notify(15005)
	return fmt.Sprintf(`%s: %s->%s`, m.Topic, m.Key, m.Value)
}

type TestFeed interface {
	Partitions() []string

	Next() (*TestFeedMessage, error)

	Close() error
}

type EnterpriseTestFeed interface {
	JobID() jobspb.JobID

	Pause() error

	Resume() error

	WaitForStatus(func(s jobs.Status) bool) error

	FetchTerminalJobErr() error

	FetchRunningStatus() (string, error)

	Details() (*jobspb.ChangefeedDetails, error)
}
