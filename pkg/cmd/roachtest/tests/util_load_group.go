package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"

type loadGroup struct {
	roachNodes option.NodeListOption
	loadNodes  option.NodeListOption
}

type loadGroupList []loadGroup

func (lg loadGroupList) roachNodes() option.NodeListOption {
	__antithesis_instrumentation__.Notify(52253)
	var roachNodes option.NodeListOption
	for _, g := range lg {
		__antithesis_instrumentation__.Notify(52255)
		roachNodes = roachNodes.Merge(g.roachNodes)
	}
	__antithesis_instrumentation__.Notify(52254)
	return roachNodes
}

func (lg loadGroupList) loadNodes() option.NodeListOption {
	__antithesis_instrumentation__.Notify(52256)
	var loadNodes option.NodeListOption
	for _, g := range lg {
		__antithesis_instrumentation__.Notify(52258)
		loadNodes = loadNodes.Merge(g.loadNodes)
	}
	__antithesis_instrumentation__.Notify(52257)
	return loadNodes
}

func makeLoadGroups(
	c interface {
		Node(int) option.NodeListOption
		Range(int, int) option.NodeListOption
	},
	numZones, numRoachNodes, numLoadNodes int,
) loadGroupList {
	__antithesis_instrumentation__.Notify(52259)
	if numLoadNodes > numZones {
		__antithesis_instrumentation__.Notify(52262)
		panic("cannot have more than one load node per zone")
	} else {
		__antithesis_instrumentation__.Notify(52263)
		if numZones%numLoadNodes != 0 {
			__antithesis_instrumentation__.Notify(52264)
			panic("numZones must be divisible by numLoadNodes")
		} else {
			__antithesis_instrumentation__.Notify(52265)
		}
	}
	__antithesis_instrumentation__.Notify(52260)

	loadNodesAtTheEnd := numLoadNodes%numZones != 0
	loadGroups := make(loadGroupList, numLoadNodes)
	roachNodesPerGroup := numRoachNodes / numLoadNodes
	for i := range loadGroups {
		__antithesis_instrumentation__.Notify(52266)
		if loadNodesAtTheEnd {
			__antithesis_instrumentation__.Notify(52267)
			first := i*roachNodesPerGroup + 1
			loadGroups[i].roachNodes = c.Range(first, first+roachNodesPerGroup-1)
			loadGroups[i].loadNodes = c.Node(numRoachNodes + i + 1)
		} else {
			__antithesis_instrumentation__.Notify(52268)
			first := i*(roachNodesPerGroup+1) + 1
			loadGroups[i].roachNodes = c.Range(first, first+roachNodesPerGroup-1)
			loadGroups[i].loadNodes = c.Node((i + 1) * (roachNodesPerGroup + 1))
		}
	}
	__antithesis_instrumentation__.Notify(52261)
	return loadGroups
}
