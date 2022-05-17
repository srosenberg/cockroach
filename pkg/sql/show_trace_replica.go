package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type showTraceReplicaNode struct {
	optColumnsSlot

	plan planNode

	run struct {
		values tree.Datums
	}
}

func (n *showTraceReplicaNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(623324)
	return nil
}

func (n *showTraceReplicaNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(623325)
	var timestampD tree.Datum
	var tag string
	for {
		__antithesis_instrumentation__.Notify(623331)
		ok, err := n.plan.Next(params)
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(623333)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(623334)
			return ok, err
		} else {
			__antithesis_instrumentation__.Notify(623335)
		}
		__antithesis_instrumentation__.Notify(623332)
		values := n.plan.Values()

		const (
			tsCol  = 0
			msgCol = 2
			tagCol = 3
		)
		if replicaMsgRE.MatchString(string(*values[msgCol].(*tree.DString))) {
			__antithesis_instrumentation__.Notify(623336)
			timestampD = values[tsCol]
			tag = string(*values[tagCol].(*tree.DString))
			break
		} else {
			__antithesis_instrumentation__.Notify(623337)
		}
	}
	__antithesis_instrumentation__.Notify(623326)

	matches := nodeStoreRangeRE.FindStringSubmatch(tag)
	if matches == nil {
		__antithesis_instrumentation__.Notify(623338)
		return false, errors.Errorf(`could not extract node, store, range from: %s`, tag)
	} else {
		__antithesis_instrumentation__.Notify(623339)
	}
	__antithesis_instrumentation__.Notify(623327)
	nodeID, err := strconv.Atoi(matches[1])
	if err != nil {
		__antithesis_instrumentation__.Notify(623340)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(623341)
	}
	__antithesis_instrumentation__.Notify(623328)
	storeID, err := strconv.Atoi(matches[2])
	if err != nil {
		__antithesis_instrumentation__.Notify(623342)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(623343)
	}
	__antithesis_instrumentation__.Notify(623329)
	rangeID, err := strconv.Atoi(matches[3])
	if err != nil {
		__antithesis_instrumentation__.Notify(623344)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(623345)
	}
	__antithesis_instrumentation__.Notify(623330)

	n.run.values = append(
		n.run.values[:0],
		timestampD,
		tree.NewDInt(tree.DInt(nodeID)),
		tree.NewDInt(tree.DInt(storeID)),
		tree.NewDInt(tree.DInt(rangeID)),
	)
	return true, nil
}

func (n *showTraceReplicaNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(623346)
	return n.run.values
}

func (n *showTraceReplicaNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(623347)
	n.plan.Close(ctx)
}

var nodeStoreRangeRE = regexp.MustCompile(`^\[n(\d+),s(\d+),r(\d+)/`)

var replicaMsgRE = regexp.MustCompile(
	strings.Join([]string{
		"^read-write path$",
		"^read-only path$",
		"^admin path$",
	}, "|"),
)
