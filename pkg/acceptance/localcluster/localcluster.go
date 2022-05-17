package localcluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"net"
	"os/exec"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/acceptance/cluster"
	"github.com/cockroachdb/errors"
)

type LocalCluster struct {
	*Cluster
}

var _ cluster.Cluster = &LocalCluster{}

func (b *LocalCluster) Port(ctx context.Context, i int) string {
	__antithesis_instrumentation__.Notify(1096)
	return b.RPCPort(i)
}

func (b *LocalCluster) NumNodes() int {
	__antithesis_instrumentation__.Notify(1097)
	return len(b.Nodes)
}

func (b *LocalCluster) NewDB(ctx context.Context, i int) (*gosql.DB, error) {
	__antithesis_instrumentation__.Notify(1098)
	return gosql.Open("postgres", b.PGUrl(ctx, i))
}

func (b *LocalCluster) PGUrl(ctx context.Context, i int) string {
	__antithesis_instrumentation__.Notify(1099)
	return b.Nodes[i].PGUrl()
}

func (b *LocalCluster) InternalIP(ctx context.Context, i int) net.IP {
	__antithesis_instrumentation__.Notify(1100)
	ips, err := net.LookupIP(b.IPAddr(i))
	if err != nil {
		__antithesis_instrumentation__.Notify(1102)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(1103)
	}
	__antithesis_instrumentation__.Notify(1101)
	return ips[0]
}

func (b *LocalCluster) Assert(ctx context.Context, t testing.TB) {
	__antithesis_instrumentation__.Notify(1104)

}

func (b *LocalCluster) AssertAndStop(ctx context.Context, t testing.TB) {
	__antithesis_instrumentation__.Notify(1105)
	b.Assert(ctx, t)
	b.Close()
}

func (b *LocalCluster) ExecCLI(ctx context.Context, i int, cmd []string) (string, string, error) {
	__antithesis_instrumentation__.Notify(1106)
	cmd = append([]string{b.Cfg.Binary}, cmd...)
	cmd = append(cmd, "--insecure", "--host", ":"+b.Port(ctx, i))
	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	var o, e bytes.Buffer
	c.Stdout, c.Stderr = &o, &e
	err := c.Run()
	if err != nil {
		__antithesis_instrumentation__.Notify(1108)
		err = errors.Wrapf(err, "cmd: %v\nstderr:\n %s\nstdout:\n %s", cmd, o.String(), e.String())
	} else {
		__antithesis_instrumentation__.Notify(1109)
	}
	__antithesis_instrumentation__.Notify(1107)
	return o.String(), e.String(), err
}

func (b *LocalCluster) Kill(ctx context.Context, i int) error {
	__antithesis_instrumentation__.Notify(1110)
	b.Nodes[i].Kill()
	return nil
}

func (b *LocalCluster) RestartAsync(ctx context.Context, i int) <-chan error {
	__antithesis_instrumentation__.Notify(1111)
	b.Nodes[i].Kill()
	joins := b.joins()
	ch := b.Nodes[i].StartAsync(ctx, joins...)
	if len(joins) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(1113)
		return len(b.Nodes) > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(1114)

		for {
			__antithesis_instrumentation__.Notify(1115)
			if gossipAddr := b.Nodes[i].AdvertiseAddr(); gossipAddr != "" {
				__antithesis_instrumentation__.Notify(1117)
				return ch
			} else {
				__antithesis_instrumentation__.Notify(1118)
			}
			__antithesis_instrumentation__.Notify(1116)
			time.Sleep(10 * time.Millisecond)
		}
	} else {
		__antithesis_instrumentation__.Notify(1119)
	}
	__antithesis_instrumentation__.Notify(1112)
	return ch
}

func (b *LocalCluster) Restart(ctx context.Context, i int) error {
	__antithesis_instrumentation__.Notify(1120)
	return <-b.RestartAsync(ctx, i)
}

func (b *LocalCluster) URL(ctx context.Context, i int) string {
	__antithesis_instrumentation__.Notify(1121)
	rest := b.Nodes[i].HTTPAddr()
	if rest == "" {
		__antithesis_instrumentation__.Notify(1123)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(1124)
	}
	__antithesis_instrumentation__.Notify(1122)
	return "http://" + rest
}

func (b *LocalCluster) Addr(ctx context.Context, i int, port string) string {
	__antithesis_instrumentation__.Notify(1125)
	return net.JoinHostPort(b.Nodes[i].AdvertiseAddr(), port)
}

func (b *LocalCluster) Hostname(i int) string {
	__antithesis_instrumentation__.Notify(1126)
	return b.IPAddr(i)
}
