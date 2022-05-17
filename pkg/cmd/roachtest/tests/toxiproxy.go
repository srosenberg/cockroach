package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/client"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/errors"
)

const cockroachToxiWrapper = `#!/usr/bin/env bash
set -eu

cd "$(dirname "${0}")"

orig_port=""

args=()

if [[ "$1" != "start" ]]; then
	./cockroach.real "$@"
	exit $?
fi

for arg in "$@"; do
	capture=$(echo "${arg}" | sed -E 's/^--port=([0-9]+)$/\1/')
	if [[ "${capture}" != "${arg}"  ]] && [[ -z "${orig_port}" ]] && [[ -n "${capture}" ]]; then
		orig_port="${capture}"
	fi
	args+=("${arg}")
done

if [[ -z "${orig_port}" ]]; then
	orig_port=26257
fi

args+=("--advertise-port=$((orig_port+10000))")

echo "toxiproxy interception:"
echo "original args: $@"
echo "modified args: ${args[@]}"
./cockroach.real "${args[@]}"
`

const toxiServerWrapper = `#!/usr/bin/env bash
set -eu

mkdir -p logs
./toxiproxy-server -host 0.0.0.0 -port $1 2>&1 > logs/toxiproxy.log & </dev/null
until nc -z localhost $1; do sleep 0.1; echo "waiting for toxiproxy-server..."; done
`

type ToxiCluster struct {
	t test.Test
	cluster.Cluster
	toxClients map[int]*toxiproxy.Client
	toxProxies map[int]*toxiproxy.Proxy
}

func Toxify(
	ctx context.Context, t test.Test, c cluster.Cluster, node option.NodeListOption,
) (*ToxiCluster, error) {
	__antithesis_instrumentation__.Notify(51367)
	toxiURL := "https://github.com/Shopify/toxiproxy/releases/download/v2.1.4/toxiproxy-server-linux-amd64"
	if c.IsLocal() && func() bool {
		__antithesis_instrumentation__.Notify(51371)
		return runtime.GOOS == "darwin" == true
	}() == true {
		__antithesis_instrumentation__.Notify(51372)
		toxiURL = "https://github.com/Shopify/toxiproxy/releases/download/v2.1.4/toxiproxy-server-darwin-amd64"
	} else {
		__antithesis_instrumentation__.Notify(51373)
	}
	__antithesis_instrumentation__.Notify(51368)
	if err := func() error {
		__antithesis_instrumentation__.Notify(51374)
		if err := c.RunE(ctx, c.All(), "curl", "-Lfo", "toxiproxy-server", toxiURL); err != nil {
			__antithesis_instrumentation__.Notify(51379)
			return err
		} else {
			__antithesis_instrumentation__.Notify(51380)
		}
		__antithesis_instrumentation__.Notify(51375)
		if err := c.RunE(ctx, c.All(), "chmod", "+x", "toxiproxy-server"); err != nil {
			__antithesis_instrumentation__.Notify(51381)
			return err
		} else {
			__antithesis_instrumentation__.Notify(51382)
		}
		__antithesis_instrumentation__.Notify(51376)

		if err := c.RunE(ctx, node, "mv cockroach cockroach.real"); err != nil {
			__antithesis_instrumentation__.Notify(51383)
			return err
		} else {
			__antithesis_instrumentation__.Notify(51384)
		}
		__antithesis_instrumentation__.Notify(51377)
		if err := c.PutString(ctx, cockroachToxiWrapper, "./cockroach", 0755, node); err != nil {
			__antithesis_instrumentation__.Notify(51385)
			return err
		} else {
			__antithesis_instrumentation__.Notify(51386)
		}
		__antithesis_instrumentation__.Notify(51378)
		return c.PutString(ctx, toxiServerWrapper, "./toxiproxyd", 0755, node)
	}(); err != nil {
		__antithesis_instrumentation__.Notify(51387)
		return nil, errors.Wrap(err, "toxify")
	} else {
		__antithesis_instrumentation__.Notify(51388)
	}
	__antithesis_instrumentation__.Notify(51369)

	tc := &ToxiCluster{
		t:          t,
		Cluster:    c,
		toxClients: make(map[int]*toxiproxy.Client),
		toxProxies: make(map[int]*toxiproxy.Proxy),
	}

	for _, i := range node {
		__antithesis_instrumentation__.Notify(51389)
		n := c.Node(i)

		toxPort := 8474 + i
		if err := c.RunE(ctx, n, fmt.Sprintf("./toxiproxyd %d 2>/dev/null >/dev/null < /dev/null", toxPort)); err != nil {
			__antithesis_instrumentation__.Notify(51394)
			return nil, errors.Wrap(err, "toxify")
		} else {
			__antithesis_instrumentation__.Notify(51395)
		}
		__antithesis_instrumentation__.Notify(51390)

		externalAddrs, err := c.ExternalAddr(ctx, t.L(), n)
		if err != nil {
			__antithesis_instrumentation__.Notify(51396)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(51397)
		}
		__antithesis_instrumentation__.Notify(51391)
		externalAddr, port, err := tc.addrToHostPort(externalAddrs[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(51398)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(51399)
		}
		__antithesis_instrumentation__.Notify(51392)
		tc.toxClients[i] = toxiproxy.NewClient(fmt.Sprintf("http://%s:%d", externalAddr, toxPort))
		proxy, err := tc.toxClients[i].CreateProxy("cockroach", fmt.Sprintf(":%d", tc.poisonedPort(port)), fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			__antithesis_instrumentation__.Notify(51400)
			return nil, errors.Wrap(err, "toxify")
		} else {
			__antithesis_instrumentation__.Notify(51401)
		}
		__antithesis_instrumentation__.Notify(51393)
		tc.toxProxies[i] = proxy
	}
	__antithesis_instrumentation__.Notify(51370)

	return tc, nil
}

func (*ToxiCluster) addrToHostPort(addr string) (string, int, error) {
	__antithesis_instrumentation__.Notify(51402)
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		__antithesis_instrumentation__.Notify(51405)
		return "", 0, err
	} else {
		__antithesis_instrumentation__.Notify(51406)
	}
	__antithesis_instrumentation__.Notify(51403)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(51407)
		return "", 0, err
	} else {
		__antithesis_instrumentation__.Notify(51408)
	}
	__antithesis_instrumentation__.Notify(51404)
	return host, port, nil
}

func (tc *ToxiCluster) poisonedPort(port int) int {
	__antithesis_instrumentation__.Notify(51409)

	_ = cockroachToxiWrapper
	return port + 10000
}

func (tc *ToxiCluster) Proxy(i int) *toxiproxy.Proxy {
	__antithesis_instrumentation__.Notify(51410)
	proxy, found := tc.toxProxies[i]
	if !found {
		__antithesis_instrumentation__.Notify(51412)
		tc.t.Fatalf("proxy for node %d not found", i)
	} else {
		__antithesis_instrumentation__.Notify(51413)
	}
	__antithesis_instrumentation__.Notify(51411)
	return proxy
}

func (tc *ToxiCluster) ExternalAddr(
	ctx context.Context, node option.NodeListOption,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(51414)
	return tc.Cluster.ExternalAddr(ctx, tc.t.L(), node)
}

func (tc *ToxiCluster) PoisonedExternalAddr(
	ctx context.Context, node option.NodeListOption,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(51415)
	var out []string

	extAddrs, err := tc.ExternalAddr(ctx, node)
	if err != nil {
		__antithesis_instrumentation__.Notify(51418)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(51419)
	}
	__antithesis_instrumentation__.Notify(51416)
	for _, addr := range extAddrs {
		__antithesis_instrumentation__.Notify(51420)
		host, port, err := tc.addrToHostPort(addr)
		if err != nil {
			__antithesis_instrumentation__.Notify(51422)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(51423)
		}
		__antithesis_instrumentation__.Notify(51421)
		out = append(out, fmt.Sprintf("%s:%d", host, tc.poisonedPort(port)))
	}
	__antithesis_instrumentation__.Notify(51417)
	return out, nil
}

func (tc *ToxiCluster) PoisonedPGAddr(
	ctx context.Context, node option.NodeListOption,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(51424)
	var out []string

	urls, err := tc.ExternalPGUrl(ctx, tc.t.L(), node)
	if err != nil {
		__antithesis_instrumentation__.Notify(51428)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(51429)
	}
	__antithesis_instrumentation__.Notify(51425)
	exts, err := tc.PoisonedExternalAddr(ctx, node)
	if err != nil {
		__antithesis_instrumentation__.Notify(51430)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(51431)
	}
	__antithesis_instrumentation__.Notify(51426)
	for i, s := range urls {
		__antithesis_instrumentation__.Notify(51432)
		u, err := url.Parse(s)
		if err != nil {
			__antithesis_instrumentation__.Notify(51434)
			tc.t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51435)
		}
		__antithesis_instrumentation__.Notify(51433)
		u.Host = exts[i]
		out = append(out, u.String())
	}
	__antithesis_instrumentation__.Notify(51427)
	return out, nil
}

func (tc *ToxiCluster) PoisonedConn(ctx context.Context, node int) *gosql.DB {
	__antithesis_instrumentation__.Notify(51436)
	urls, err := tc.PoisonedPGAddr(ctx, tc.Cluster.Node(node))
	if err != nil {
		__antithesis_instrumentation__.Notify(51439)
		tc.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51440)
	}
	__antithesis_instrumentation__.Notify(51437)
	db, err := gosql.Open("postgres", urls[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(51441)
		tc.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51442)
	}
	__antithesis_instrumentation__.Notify(51438)
	return db
}

var _ = (*ToxiCluster)(nil).PoisonedConn
var _ = (*ToxiCluster)(nil).PoisonedPGAddr
var _ = (*ToxiCluster)(nil).PoisonedExternalAddr

var measureRE = regexp.MustCompile(`real[^0-9]+([0-9.]+)`)

func (tc *ToxiCluster) Measure(ctx context.Context, fromNode int, stmt string) time.Duration {
	__antithesis_instrumentation__.Notify(51443)
	externalAddrs, err := tc.ExternalAddr(ctx, tc.Node(fromNode))
	if err != nil {
		__antithesis_instrumentation__.Notify(51449)
		tc.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51450)
	}
	__antithesis_instrumentation__.Notify(51444)
	_, port, err := tc.addrToHostPort(externalAddrs[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(51451)
		tc.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51452)
	}
	__antithesis_instrumentation__.Notify(51445)
	result, err := tc.Cluster.RunWithDetailsSingleNode(ctx, tc.t.L(), tc.Cluster.Node(fromNode), "time", "-p", "./cockroach", "sql", "--insecure", "--port", strconv.Itoa(port), "-e", "'"+stmt+"'")
	output := []byte(result.Stdout + result.Stderr)
	tc.t.L().Printf("%s\n", output)
	if err != nil {
		__antithesis_instrumentation__.Notify(51453)
		tc.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51454)
	}
	__antithesis_instrumentation__.Notify(51446)
	matches := measureRE.FindSubmatch(output)
	if len(matches) != 2 {
		__antithesis_instrumentation__.Notify(51455)
		tc.t.Fatalf("unable to extract duration from output: %s", output)
	} else {
		__antithesis_instrumentation__.Notify(51456)
	}
	__antithesis_instrumentation__.Notify(51447)
	f, err := strconv.ParseFloat(string(matches[1]), 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(51457)
		tc.t.Fatalf("unable to parse %s as float: %s", output, err)
	} else {
		__antithesis_instrumentation__.Notify(51458)
	}
	__antithesis_instrumentation__.Notify(51448)
	return time.Duration(f * 1e9)
}
