package cluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"net"
	"testing"

	"github.com/cockroachdb/errors"
)

type Cluster interface {
	NumNodes() int

	NewDB(context.Context, int) (*gosql.DB, error)

	PGUrl(context.Context, int) string

	InternalIP(ctx context.Context, i int) net.IP

	Assert(context.Context, testing.TB)

	AssertAndStop(context.Context, testing.TB)

	ExecCLI(ctx context.Context, i int, args []string) (string, string, error)

	Kill(context.Context, int) error

	Restart(context.Context, int) error

	URL(context.Context, int) string

	Addr(ctx context.Context, i int, port string) string

	Hostname(i int) string
}

func Consistent(ctx context.Context, c Cluster, i int) error {
	__antithesis_instrumentation__.Notify(19)
	return errors.Errorf("Consistency checking is unimplmented and should be re-implemented using SQL")
}
