package democluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type DemoLocalityList []roachpb.Locality

func (l *DemoLocalityList) Type() string {
	__antithesis_instrumentation__.Notify(32458)
	return "demoLocalityList"
}

func (l *DemoLocalityList) String() string {
	__antithesis_instrumentation__.Notify(32459)
	s := ""
	for _, loc := range []roachpb.Locality(*l) {
		__antithesis_instrumentation__.Notify(32461)
		s += loc.String()
	}
	__antithesis_instrumentation__.Notify(32460)
	return s
}

func (l *DemoLocalityList) Set(value string) error {
	__antithesis_instrumentation__.Notify(32462)
	*l = []roachpb.Locality{}
	locs := strings.Split(value, ":")
	for _, value := range locs {
		__antithesis_instrumentation__.Notify(32464)
		parsedLoc := &roachpb.Locality{}
		if err := parsedLoc.Set(value); err != nil {
			__antithesis_instrumentation__.Notify(32466)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32467)
		}
		__antithesis_instrumentation__.Notify(32465)
		*l = append(*l, *parsedLoc)
	}
	__antithesis_instrumentation__.Notify(32463)
	return nil
}

var defaultLocalities = DemoLocalityList{

	{Tiers: []roachpb.Tier{{Key: "region", Value: "us-east1"}, {Key: "az", Value: "b"}}},
	{Tiers: []roachpb.Tier{{Key: "region", Value: "us-east1"}, {Key: "az", Value: "c"}}},
	{Tiers: []roachpb.Tier{{Key: "region", Value: "us-east1"}, {Key: "az", Value: "d"}}},

	{Tiers: []roachpb.Tier{{Key: "region", Value: "us-west1"}, {Key: "az", Value: "a"}}},
	{Tiers: []roachpb.Tier{{Key: "region", Value: "us-west1"}, {Key: "az", Value: "b"}}},
	{Tiers: []roachpb.Tier{{Key: "region", Value: "us-west1"}, {Key: "az", Value: "c"}}},

	{Tiers: []roachpb.Tier{{Key: "region", Value: "europe-west1"}, {Key: "az", Value: "b"}}},
	{Tiers: []roachpb.Tier{{Key: "region", Value: "europe-west1"}, {Key: "az", Value: "c"}}},
	{Tiers: []roachpb.Tier{{Key: "region", Value: "europe-west1"}, {Key: "az", Value: "d"}}},
}
