package install

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

type Node int

type Nodes []Node

func ListNodes(s string, numNodesInCluster int) (Nodes, error) {
	__antithesis_instrumentation__.Notify(181615)
	if s == "" {
		__antithesis_instrumentation__.Notify(181622)
		return nil, errors.AssertionFailedf("empty node selector")
	} else {
		__antithesis_instrumentation__.Notify(181623)
	}
	__antithesis_instrumentation__.Notify(181616)
	if numNodesInCluster < 1 {
		__antithesis_instrumentation__.Notify(181624)
		return nil, errors.AssertionFailedf("invalid number of nodes %d", numNodesInCluster)
	} else {
		__antithesis_instrumentation__.Notify(181625)
	}
	__antithesis_instrumentation__.Notify(181617)

	if s == "all" {
		__antithesis_instrumentation__.Notify(181626)
		return allNodes(numNodesInCluster), nil
	} else {
		__antithesis_instrumentation__.Notify(181627)
	}
	__antithesis_instrumentation__.Notify(181618)

	var set util.FastIntSet
	for _, p := range strings.Split(s, ",") {
		__antithesis_instrumentation__.Notify(181628)
		parts := strings.Split(p, "-")
		switch len(parts) {
		case 1:
			__antithesis_instrumentation__.Notify(181629)
			i, err := strconv.Atoi(parts[0])
			if err != nil {
				__antithesis_instrumentation__.Notify(181635)
				return nil, errors.Wrapf(err, "unable to parse node selector '%s'", s)
			} else {
				__antithesis_instrumentation__.Notify(181636)
			}
			__antithesis_instrumentation__.Notify(181630)
			set.Add(i)

		case 2:
			__antithesis_instrumentation__.Notify(181631)
			from, err := strconv.Atoi(parts[0])
			if err != nil {
				__antithesis_instrumentation__.Notify(181637)
				return nil, errors.Wrapf(err, "unable to parse node selector '%s'", s)
			} else {
				__antithesis_instrumentation__.Notify(181638)
			}
			__antithesis_instrumentation__.Notify(181632)
			to, err := strconv.Atoi(parts[1])
			if err != nil {
				__antithesis_instrumentation__.Notify(181639)
				return nil, errors.Wrapf(err, "unable to parse node selector '%s'", s)
			} else {
				__antithesis_instrumentation__.Notify(181640)
			}
			__antithesis_instrumentation__.Notify(181633)
			set.AddRange(from, to)

		default:
			__antithesis_instrumentation__.Notify(181634)
			return nil, fmt.Errorf("unable to parse node selector '%s'", p)
		}
	}
	__antithesis_instrumentation__.Notify(181619)
	nodes := make(Nodes, 0, set.Len())
	set.ForEach(func(v int) {
		__antithesis_instrumentation__.Notify(181641)
		nodes = append(nodes, Node(v))
	})
	__antithesis_instrumentation__.Notify(181620)

	for _, n := range nodes {
		__antithesis_instrumentation__.Notify(181642)
		if n < 1 {
			__antithesis_instrumentation__.Notify(181644)
			return nil, fmt.Errorf("invalid node selector '%s', node values start at 1", s)
		} else {
			__antithesis_instrumentation__.Notify(181645)
		}
		__antithesis_instrumentation__.Notify(181643)
		if int(n) > numNodesInCluster {
			__antithesis_instrumentation__.Notify(181646)
			return nil, fmt.Errorf(
				"invalid node selector '%s', cluster contains %d nodes", s, numNodesInCluster,
			)
		} else {
			__antithesis_instrumentation__.Notify(181647)
		}
	}
	__antithesis_instrumentation__.Notify(181621)
	return nodes, nil
}

func allNodes(numNodesInCluster int) Nodes {
	__antithesis_instrumentation__.Notify(181648)
	r := make(Nodes, numNodesInCluster)
	for i := range r {
		__antithesis_instrumentation__.Notify(181650)
		r[i] = Node(i + 1)
	}
	__antithesis_instrumentation__.Notify(181649)
	return r
}
