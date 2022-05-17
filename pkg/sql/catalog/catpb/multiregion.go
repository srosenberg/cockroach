package catpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type RegionName string

func (r RegionName) String() string {
	__antithesis_instrumentation__.Notify(249370)
	return string(r)
}

type RegionNames []RegionName

func (regions RegionNames) ToStrings() []string {
	__antithesis_instrumentation__.Notify(249371)
	ret := make([]string, len(regions))
	for i, region := range regions {
		__antithesis_instrumentation__.Notify(249373)
		ret[i] = string(region)
	}
	__antithesis_instrumentation__.Notify(249372)
	return ret
}

func (cfg *LocalityConfig) TelemetryName() (string, error) {
	__antithesis_instrumentation__.Notify(249374)
	switch l := cfg.Locality.(type) {
	case *LocalityConfig_Global_:
		__antithesis_instrumentation__.Notify(249376)
		return tree.TelemetryNameGlobal, nil
	case *LocalityConfig_RegionalByTable_:
		__antithesis_instrumentation__.Notify(249377)
		if l.RegionalByTable.Region != nil {
			__antithesis_instrumentation__.Notify(249381)
			return tree.TelemetryNameRegionalByTableIn, nil
		} else {
			__antithesis_instrumentation__.Notify(249382)
		}
		__antithesis_instrumentation__.Notify(249378)
		return tree.TelemetryNameRegionalByTable, nil
	case *LocalityConfig_RegionalByRow_:
		__antithesis_instrumentation__.Notify(249379)
		if l.RegionalByRow.As != nil {
			__antithesis_instrumentation__.Notify(249383)
			return tree.TelemetryNameRegionalByRowAs, nil
		} else {
			__antithesis_instrumentation__.Notify(249384)
		}
		__antithesis_instrumentation__.Notify(249380)
		return tree.TelemetryNameRegionalByRow, nil
	}
	__antithesis_instrumentation__.Notify(249375)
	return "", errors.AssertionFailedf(
		"unknown locality config TelemetryName: type %T",
		cfg.Locality,
	)
}
