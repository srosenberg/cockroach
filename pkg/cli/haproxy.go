package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var haProxyPath string
var haProxyLocality roachpb.Locality

var genHAProxyCmd = &cobra.Command{
	Use:   "haproxy",
	Short: "generate haproxy.cfg for the connected cluster",
	Long: `This command generates a minimal haproxy configuration file for the cluster
reached through the client flags.
The file is written to --out. Use "--out -" for stdout.

The addresses used are those advertised by the nodes themselves. Make sure haproxy
can resolve the hostnames in the configuration file, either by using full-qualified names, or
running haproxy in the same network.

Nodes that have been decommissioned are excluded from the generated configuration.

Nodes to include can be filtered by localities matching the '--locality' regular expression. eg:
  --locality=region=us-east                  # Nodes in region "us-east"
  --locality=region=us.*                     # Nodes in the US
  --locality=region=us.*,deployment=testing  # Nodes in the US AND in deployment tier "testing"

A regular expression can be specified per locality tier and all specified tiers must match.
The key (eg: 'region') must be fully specified, only values (eg: 'us-east1') can be regular expressions.
An error is returned if no nodes match the locality filter.
`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runGenHAProxyCmd),
}

type haProxyNodeInfo struct {
	NodeID   roachpb.NodeID
	NodeAddr string

	CheckPort string
	Locality  roachpb.Locality
}

func nodeStatusesToNodeInfos(nodes *serverpb.NodesResponse) []haProxyNodeInfo {
	__antithesis_instrumentation__.Notify(33040)
	fs := pflag.NewFlagSet("haproxy", pflag.ContinueOnError)

	httpAddr := ""
	httpPort := base.DefaultHTTPPort
	fs.Var(addrSetter{&httpAddr, &httpPort}, cliflags.ListenHTTPAddr.Name, "")
	fs.Var(aliasStrVar{&httpPort}, cliflags.ListenHTTPPort.Name, "")

	fs.SetOutput(ioutil.Discard)

	nodeInfos := make([]haProxyNodeInfo, 0, len(nodes.Nodes))

	nodeIDs := make([]int, 0, len(nodes.Nodes))
	statusByID := make(map[roachpb.NodeID]statuspb.NodeStatus)
	for _, status := range nodes.Nodes {
		__antithesis_instrumentation__.Notify(33043)
		statusByID[status.Desc.NodeID] = status
		nodeIDs = append(nodeIDs, int(status.Desc.NodeID))
	}
	__antithesis_instrumentation__.Notify(33041)
	sort.Ints(nodeIDs)

	for _, inodeID := range nodeIDs {
		__antithesis_instrumentation__.Notify(33044)
		nodeID := roachpb.NodeID(inodeID)
		status := statusByID[nodeID]
		liveness := nodes.LivenessByNodeID[nodeID]
		switch liveness {
		case livenesspb.NodeLivenessStatus_DECOMMISSIONING:
			__antithesis_instrumentation__.Notify(33047)
			fmt.Fprintf(stderr, "warning: node %d status is %s, excluding from haproxy configuration\n",
				nodeID, liveness)
			fallthrough
		case livenesspb.NodeLivenessStatus_DECOMMISSIONED:
			__antithesis_instrumentation__.Notify(33048)
			continue
		default:
			__antithesis_instrumentation__.Notify(33049)
		}
		__antithesis_instrumentation__.Notify(33045)

		info := haProxyNodeInfo{
			NodeID:   nodeID,
			NodeAddr: status.Desc.Address.AddressField,
			Locality: status.Desc.Locality,
		}

		httpPort = base.DefaultHTTPPort

		for j, arg := range status.Args {
			__antithesis_instrumentation__.Notify(33050)
			if strings.Contains(arg, cliflags.ListenHTTPPort.Name) || func() bool {
				__antithesis_instrumentation__.Notify(33051)
				return strings.Contains(arg, cliflags.ListenHTTPAddr.Name) == true
			}() == true {
				__antithesis_instrumentation__.Notify(33052)
				_ = fs.Parse(status.Args[j:])
				break
			} else {
				__antithesis_instrumentation__.Notify(33053)
			}
		}
		__antithesis_instrumentation__.Notify(33046)

		info.CheckPort = httpPort
		nodeInfos = append(nodeInfos, info)
	}
	__antithesis_instrumentation__.Notify(33042)

	return nodeInfos
}

func localityMatches(locality roachpb.Locality, desired roachpb.Locality) (bool, error) {
	__antithesis_instrumentation__.Notify(33054)
	for _, filterTier := range desired.Tiers {
		__antithesis_instrumentation__.Notify(33056)

		var b strings.Builder
		b.WriteString("^")
		b.WriteString(filterTier.Value)
		b.WriteString("$")
		re, err := regexp.Compile(b.String())
		if err != nil {
			__antithesis_instrumentation__.Notify(33059)
			return false, errors.Wrapf(err, "could not compile regular expression for %q", filterTier)
		} else {
			__antithesis_instrumentation__.Notify(33060)
		}
		__antithesis_instrumentation__.Notify(33057)

		keyFound := false
		for _, nodeTier := range locality.Tiers {
			__antithesis_instrumentation__.Notify(33061)
			if filterTier.Key != nodeTier.Key {
				__antithesis_instrumentation__.Notify(33064)
				continue
			} else {
				__antithesis_instrumentation__.Notify(33065)
			}
			__antithesis_instrumentation__.Notify(33062)

			keyFound = true
			if !re.MatchString(nodeTier.Value) {
				__antithesis_instrumentation__.Notify(33066)

				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(33067)
			}
			__antithesis_instrumentation__.Notify(33063)

			break
		}
		__antithesis_instrumentation__.Notify(33058)

		if !keyFound {
			__antithesis_instrumentation__.Notify(33068)

			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(33069)
		}
	}
	__antithesis_instrumentation__.Notify(33055)

	return true, nil
}

func filterByLocality(nodeInfos []haProxyNodeInfo) ([]haProxyNodeInfo, error) {
	__antithesis_instrumentation__.Notify(33070)
	if len(haProxyLocality.Tiers) == 0 {
		__antithesis_instrumentation__.Notify(33074)

		return nodeInfos, nil
	} else {
		__antithesis_instrumentation__.Notify(33075)
	}
	__antithesis_instrumentation__.Notify(33071)

	result := make([]haProxyNodeInfo, 0)
	availableLocalities := make(map[string]struct{})

	for _, info := range nodeInfos {
		__antithesis_instrumentation__.Notify(33076)
		l := info.Locality
		if len(l.Tiers) == 0 {
			__antithesis_instrumentation__.Notify(33079)
			continue
		} else {
			__antithesis_instrumentation__.Notify(33080)
		}
		__antithesis_instrumentation__.Notify(33077)

		availableLocalities[l.String()] = struct{}{}

		matches, err := localityMatches(l, haProxyLocality)
		if err != nil {
			__antithesis_instrumentation__.Notify(33081)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(33082)
		}
		__antithesis_instrumentation__.Notify(33078)

		if matches {
			__antithesis_instrumentation__.Notify(33083)
			result = append(result, info)
		} else {
			__antithesis_instrumentation__.Notify(33084)
		}
	}
	__antithesis_instrumentation__.Notify(33072)

	if len(result) == 0 {
		__antithesis_instrumentation__.Notify(33085)
		seenLocalities := make([]string, len(availableLocalities))
		i := 0
		for l := range availableLocalities {
			__antithesis_instrumentation__.Notify(33087)
			seenLocalities[i] = l
			i++
		}
		__antithesis_instrumentation__.Notify(33086)
		sort.Strings(seenLocalities)
		return nil, fmt.Errorf("no nodes match locality filter %s. Found localities: %v", haProxyLocality.String(), seenLocalities)
	} else {
		__antithesis_instrumentation__.Notify(33088)
	}
	__antithesis_instrumentation__.Notify(33073)

	return result, nil
}

func runGenHAProxyCmd(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(33089)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configTemplate, err := template.New("haproxy template").Parse(haProxyTemplate)
	if err != nil {
		__antithesis_instrumentation__.Notify(33097)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33098)
	}
	__antithesis_instrumentation__.Notify(33090)

	conn, _, finish, err := getClientGRPCConn(ctx, serverCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(33099)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33100)
	}
	__antithesis_instrumentation__.Notify(33091)
	defer finish()
	c := serverpb.NewStatusClient(conn)

	nodeStatuses, err := c.Nodes(ctx, &serverpb.NodesRequest{})
	if err != nil {
		__antithesis_instrumentation__.Notify(33101)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33102)
	}
	__antithesis_instrumentation__.Notify(33092)

	var w io.Writer
	var f *os.File
	if haProxyPath == "-" {
		__antithesis_instrumentation__.Notify(33103)
		w = os.Stdout
	} else {
		__antithesis_instrumentation__.Notify(33104)
		if f, err = os.OpenFile(haProxyPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
			__antithesis_instrumentation__.Notify(33105)
			return err
		} else {
			__antithesis_instrumentation__.Notify(33106)
			w = f
		}
	}
	__antithesis_instrumentation__.Notify(33093)

	nodeInfos := nodeStatusesToNodeInfos(nodeStatuses)
	filteredNodeInfos, err := filterByLocality(nodeInfos)
	if err != nil {
		__antithesis_instrumentation__.Notify(33107)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33108)
	}
	__antithesis_instrumentation__.Notify(33094)

	err = configTemplate.Execute(w, filteredNodeInfos)
	if err != nil {
		__antithesis_instrumentation__.Notify(33109)

		_ = f.Close()
		return err
	} else {
		__antithesis_instrumentation__.Notify(33110)
	}
	__antithesis_instrumentation__.Notify(33095)

	if f != nil {
		__antithesis_instrumentation__.Notify(33111)
		return f.Close()
	} else {
		__antithesis_instrumentation__.Notify(33112)
	}
	__antithesis_instrumentation__.Notify(33096)

	return nil
}

const haProxyTemplate = `
global
  maxconn 4096

defaults
    mode                tcp

    # Timeout values should be configured for your specific use.
    # See: https://cbonte.github.io/haproxy-dconv/1.8/configuration.html#4-timeout%20connect

    # With the timeout connect 5 secs,
    # if the backend server is not responding, haproxy will make a total
    # of 3 connection attempts waiting 5s each time before giving up on the server,
    # for a total of 15 seconds.
    retries             2
    timeout connect     5s

    # timeout client and server govern the maximum amount of time of TCP inactivity.
    # The server node may idle on a TCP connection either because it takes time to
    # execute a query before the first result set record is emitted, or in case of
    # some trouble on the server. So these timeout settings should be larger than the
    # time to execute the longest (most complex, under substantial concurrent workload)
    # query, yet not too large so truly failed connections are lingering too long
    # (resources associated with failed connections should be freed reasonably promptly).
    timeout client      10m
    timeout server      10m

    # TCP keep-alive on client side. Server already enables them.
    option              clitcpka

listen psql
    bind :26257
    mode tcp
    balance roundrobin
    option httpchk GET /health?ready=1
{{range .}}    server cockroach{{.NodeID}} {{.NodeAddr}} check port {{.CheckPort}}
{{end}}
`
