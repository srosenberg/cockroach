package democluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type regionPair struct {
	regionA string
	regionB string
}

var regionToRegionToLatency map[string]map[string]int

func insertPair(pair regionPair, latency int) {
	__antithesis_instrumentation__.Notify(32468)
	regionToLatency, ok := regionToRegionToLatency[pair.regionA]
	if !ok {
		__antithesis_instrumentation__.Notify(32470)
		regionToLatency = make(map[string]int)
		regionToRegionToLatency[pair.regionA] = regionToLatency
	} else {
		__antithesis_instrumentation__.Notify(32471)
	}
	__antithesis_instrumentation__.Notify(32469)
	regionToLatency[pair.regionB] = latency
}

var regionRoundTripLatencies = map[regionPair]int{
	{regionA: "us-east1", regionB: "us-west1"}:     66,
	{regionA: "us-east1", regionB: "europe-west1"}: 64,
	{regionA: "us-west1", regionB: "europe-west1"}: 146,
}

var regionOneWayLatencies = make(map[regionPair]int)

func init() {

	for pair, latency := range regionRoundTripLatencies {
		regionOneWayLatencies[pair] = latency / 2
	}
	regionToRegionToLatency = make(map[string]map[string]int)
	for pair, latency := range regionOneWayLatencies {
		insertPair(pair, latency)
		insertPair(regionPair{
			regionA: pair.regionB,
			regionB: pair.regionA,
		}, latency)
	}
}
