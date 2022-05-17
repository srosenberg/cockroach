/*
Package simulation provides tools meant to visualize or test aspects
of a Cockroach cluster on a single host.

Gossip

Gossip creates a gossip network of up to 250 nodes and outputs
successive visualization of the gossip network graph via dot.

Uses tcp sockets for connecting 3, 10, 25, 50, 100 or 250
nodes. Generates .dot graph output files for each cycle of the
simulation.

To run:

    go install github.com/cockroachdb/cockroach/cmd/gossipsim
    gossipsim -size=(small|medium|large|huge|ginormous)

Log output includes instructions for displaying the graph output as a
series of images to visualize the evolution of the network.

Running the large through ginormous simulations will require the open
files limit be increased either for the shell running the simulation,
or system wide. For Linux:

    # For the current shell:
    ulimit -n 65536

    # System-wide:
    sysctl fs.file-max
    fs.file-max = 50384

For MacOS:

    # To view current limits (soft / hard):
    launchctl limit maxfiles

    # To edit, add/edit the following line in /etc/launchd.conf and
    # restart for the new file limit to take effect.
    #
    # limit maxfiles 16384 32768
    sudo vi /etc/launchd.conf
*/
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/gossip/simulation"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
)

const (
	minDotFontSize = 12

	maxDotFontSize = 24
)

var (
	size = flag.String("size", "medium", "size of network (tiny|small|medium|large|huge|ginormous)")
)

type edge struct {
	dest    roachpb.NodeID
	added   bool
	deleted bool
}

type edgeMap map[roachpb.NodeID][]edge

func (em edgeMap) addEdge(nodeID roachpb.NodeID, e edge) {
	__antithesis_instrumentation__.Notify(40972)
	if _, ok := em[nodeID]; !ok {
		__antithesis_instrumentation__.Notify(40974)
		em[nodeID] = make([]edge, 0, 1)
	} else {
		__antithesis_instrumentation__.Notify(40975)
	}
	__antithesis_instrumentation__.Notify(40973)
	em[nodeID] = append(em[nodeID], e)
}

func outputDotFile(
	dotFN string, cycle int, network *simulation.Network, edgeSet map[string]edge,
) (string, bool) {
	__antithesis_instrumentation__.Notify(40976)
	f, err := os.Create(dotFN)
	if err != nil {
		__antithesis_instrumentation__.Notify(40981)
		log.Fatalf(context.TODO(), "unable to create temp file: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(40982)
	}
	__antithesis_instrumentation__.Notify(40977)
	defer f.Close()

	outgoingMap := make(edgeMap)
	var maxIncoming int
	quiescent := true

	for _, simNode := range network.Nodes {
		__antithesis_instrumentation__.Notify(40983)
		node := simNode.Gossip
		incoming := node.Incoming()
		for _, iNode := range incoming {
			__antithesis_instrumentation__.Notify(40985)
			e := edge{dest: node.NodeID.Get()}
			key := fmt.Sprintf("%d:%d", iNode, node.NodeID.Get())
			if _, ok := edgeSet[key]; !ok {
				__antithesis_instrumentation__.Notify(40987)
				e.added = true
				quiescent = false
			} else {
				__antithesis_instrumentation__.Notify(40988)
			}
			__antithesis_instrumentation__.Notify(40986)
			delete(edgeSet, key)
			outgoingMap.addEdge(iNode, e)
		}
		__antithesis_instrumentation__.Notify(40984)
		if len(incoming) > maxIncoming {
			__antithesis_instrumentation__.Notify(40989)
			maxIncoming = len(incoming)
		} else {
			__antithesis_instrumentation__.Notify(40990)
		}
	}
	__antithesis_instrumentation__.Notify(40978)

	for key, e := range edgeSet {
		__antithesis_instrumentation__.Notify(40991)
		e.added = false
		e.deleted = true
		quiescent = false
		nodeID, err := strconv.Atoi(strings.Split(key, ":")[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(40993)
			log.Fatalf(context.TODO(), "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(40994)
		}
		__antithesis_instrumentation__.Notify(40992)
		outgoingMap.addEdge(roachpb.NodeID(nodeID), e)
		delete(edgeSet, key)
	}
	__antithesis_instrumentation__.Notify(40979)

	fmt.Fprintln(f, "digraph G {")
	fmt.Fprintln(f, "node [shape=record];")
	for _, simNode := range network.Nodes {
		__antithesis_instrumentation__.Notify(40995)
		node := simNode.Gossip
		var missing []roachpb.NodeID
		var totalAge int64
		for _, otherNode := range network.Nodes {
			__antithesis_instrumentation__.Notify(41000)
			if otherNode == simNode {
				__antithesis_instrumentation__.Notify(41002)
				continue
			} else {
				__antithesis_instrumentation__.Notify(41003)
			}
			__antithesis_instrumentation__.Notify(41001)
			infoKey := otherNode.Addr().String()

			if info, err := node.GetInfo(infoKey); err != nil {
				__antithesis_instrumentation__.Notify(41004)
				missing = append(missing, otherNode.Gossip.NodeID.Get())
				quiescent = false
			} else {
				__antithesis_instrumentation__.Notify(41005)
				_, val, err := encoding.DecodeUint64Ascending(info)
				if err != nil {
					__antithesis_instrumentation__.Notify(41007)
					log.Fatalf(context.TODO(), "bad decode of node info cycle: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(41008)
				}
				__antithesis_instrumentation__.Notify(41006)
				totalAge += int64(cycle) - int64(val)
			}
		}
		__antithesis_instrumentation__.Notify(40996)
		log.Infof(context.TODO(), "node %d: missing infos for nodes %s", node.NodeID.Get(), missing)

		var sentinelAge int64

		if info, err := node.GetInfo(gossip.KeySentinel); err != nil {
			__antithesis_instrumentation__.Notify(41009)
			log.Infof(context.TODO(), "error getting info for sentinel gossip key %q: %s", gossip.KeySentinel, err)
		} else {
			__antithesis_instrumentation__.Notify(41010)
			_, val, err := encoding.DecodeUint64Ascending(info)
			if err != nil {
				__antithesis_instrumentation__.Notify(41012)
				log.Fatalf(context.TODO(), "bad decode of sentinel cycle: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(41013)
			}
			__antithesis_instrumentation__.Notify(41011)
			sentinelAge = int64(cycle) - int64(val)
		}
		__antithesis_instrumentation__.Notify(40997)

		var age, nodeColor string
		if len(missing) > 0 {
			__antithesis_instrumentation__.Notify(41014)
			nodeColor = "color=red,"
			age = fmt.Sprintf("missing %d", len(missing))
		} else {
			__antithesis_instrumentation__.Notify(41015)
			age = strconv.FormatFloat(float64(totalAge)/float64(len(network.Nodes)-1-len(missing)), 'f', 4, 64)
		}
		__antithesis_instrumentation__.Notify(40998)
		fontSize := minDotFontSize
		if maxIncoming > 0 {
			__antithesis_instrumentation__.Notify(41016)
			fontSize = minDotFontSize + int(math.Floor(float64(len(node.Incoming())*
				(maxDotFontSize-minDotFontSize))/float64(maxIncoming)))
		} else {
			__antithesis_instrumentation__.Notify(41017)
		}
		__antithesis_instrumentation__.Notify(40999)
		fmt.Fprintf(f, "\t%s [%sfontsize=%d,label=\"{%s|AA=%s, MH=%d, SA=%d}\"]\n",
			node.NodeID.Get(), nodeColor, fontSize, node.NodeID.Get(), age, node.MaxHops(), sentinelAge)
		outgoing := outgoingMap[node.NodeID.Get()]
		for _, e := range outgoing {
			__antithesis_instrumentation__.Notify(41018)
			destSimNode, ok := network.GetNodeFromID(e.dest)
			if !ok {
				__antithesis_instrumentation__.Notify(41021)
				continue
			} else {
				__antithesis_instrumentation__.Notify(41022)
			}
			__antithesis_instrumentation__.Notify(41019)
			dest := destSimNode.Gossip
			style := ""
			if e.added {
				__antithesis_instrumentation__.Notify(41023)
				style = " [color=green]"
			} else {
				__antithesis_instrumentation__.Notify(41024)
				if e.deleted {
					__antithesis_instrumentation__.Notify(41025)
					style = " [color=red,style=dotted]"
				} else {
					__antithesis_instrumentation__.Notify(41026)
				}
			}
			__antithesis_instrumentation__.Notify(41020)
			fmt.Fprintf(f, "\t%s -> %s%s\n", node.NodeID.Get(), dest.NodeID.Get(), style)
			if !e.deleted {
				__antithesis_instrumentation__.Notify(41027)
				edgeSet[fmt.Sprintf("%d:%d", node.NodeID.Get(), e.dest)] = e
			} else {
				__antithesis_instrumentation__.Notify(41028)
			}
		}
	}
	__antithesis_instrumentation__.Notify(40980)
	fmt.Fprintln(f, "}")
	return f.Name(), quiescent
}

func main() {
	__antithesis_instrumentation__.Notify(41029)

	randutil.SeedForTests()

	if f := flag.Lookup("logtostderr"); f != nil {
		__antithesis_instrumentation__.Notify(41034)
		fmt.Println("Starting simulation. Add -logtostderr to see progress.")
	} else {
		__antithesis_instrumentation__.Notify(41035)
	}
	__antithesis_instrumentation__.Notify(41030)
	flag.Parse()

	dirName, err := ioutil.TempDir("", "gossip-simulation-")
	if err != nil {
		__antithesis_instrumentation__.Notify(41036)
		log.Fatalf(context.TODO(), "could not create temporary directory for gossip simulation output: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(41037)
	}
	__antithesis_instrumentation__.Notify(41031)

	nodeCount := 3
	switch *size {
	case "tiny":
		__antithesis_instrumentation__.Notify(41038)

	case "small":
		__antithesis_instrumentation__.Notify(41039)
		nodeCount = 10
	case "medium":
		__antithesis_instrumentation__.Notify(41040)
		nodeCount = 25
	case "large":
		__antithesis_instrumentation__.Notify(41041)
		nodeCount = 50
	case "huge":
		__antithesis_instrumentation__.Notify(41042)
		nodeCount = 100
	case "ginormous":
		__antithesis_instrumentation__.Notify(41043)
		nodeCount = 250
	default:
		__antithesis_instrumentation__.Notify(41044)
		log.Fatalf(context.TODO(), "unknown simulation size: %s", *size)
	}
	__antithesis_instrumentation__.Notify(41032)

	edgeSet := make(map[string]edge)

	stopper := stop.NewStopper()
	defer stopper.Stop(context.TODO())

	n := simulation.NewNetwork(stopper, nodeCount, true, zonepb.DefaultZoneConfigRef())
	n.SimulateNetwork(
		func(cycle int, network *simulation.Network) bool {
			__antithesis_instrumentation__.Notify(41045)

			dotFN := fmt.Sprintf("%s/sim-cycle-%03d.dot", dirName, cycle)
			_, quiescent := outputDotFile(dotFN, cycle, network, edgeSet)

			return !quiescent
		},
	)
	__antithesis_instrumentation__.Notify(41033)

	fmt.Printf("To view simulation graph output run (you must install graphviz):\n\nfor f in %s/*.dot ; do circo $f -Tpng -o $f.png ; echo $f.png ; done\n", dirName)
}
