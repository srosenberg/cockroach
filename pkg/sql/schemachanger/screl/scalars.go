package screl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/errors"
)

func GetDescID(e scpb.Element) catid.DescID {
	__antithesis_instrumentation__.Notify(595027)
	id, err := Schema.GetAttribute(DescID, e)
	if err != nil {
		__antithesis_instrumentation__.Notify(595029)

		panic(errors.NewAssertionErrorWithWrappedErrf(
			err, "failed to retrieve descriptor ID for %T", e,
		))
	} else {
		__antithesis_instrumentation__.Notify(595030)
	}
	__antithesis_instrumentation__.Notify(595028)
	return id.(catid.DescID)
}

func AllTargetDescIDs(s scpb.TargetState) (ids catalog.DescriptorIDSet) {
	__antithesis_instrumentation__.Notify(595031)
	for i := range s.Targets {
		__antithesis_instrumentation__.Notify(595033)
		e := s.Targets[i].Element()

		switch te := e.(type) {
		case *scpb.Namespace:
			__antithesis_instrumentation__.Notify(595034)

			ids.Add(te.DescriptorID)
		case *scpb.ObjectParent:
			__antithesis_instrumentation__.Notify(595035)

			ids.Add(te.ObjectID)
		default:
			__antithesis_instrumentation__.Notify(595036)
			_ = WalkDescIDs(e, func(id *catid.DescID) error {
				__antithesis_instrumentation__.Notify(595037)
				ids.Add(*id)
				return nil
			})
		}
	}
	__antithesis_instrumentation__.Notify(595032)
	return ids
}

func AllDescIDs(e scpb.Element) (ids catalog.DescriptorIDSet) {
	__antithesis_instrumentation__.Notify(595038)
	if e == nil {
		__antithesis_instrumentation__.Notify(595041)
		return ids
	} else {
		__antithesis_instrumentation__.Notify(595042)
	}
	__antithesis_instrumentation__.Notify(595039)
	_ = WalkDescIDs(e, func(id *catid.DescID) error {
		__antithesis_instrumentation__.Notify(595043)
		ids.Add(*id)
		return nil
	})
	__antithesis_instrumentation__.Notify(595040)
	return ids
}

func ContainsDescID(haystack scpb.Element, needle catid.DescID) (contains bool) {
	__antithesis_instrumentation__.Notify(595044)
	_ = WalkDescIDs(haystack, func(id *catid.DescID) error {
		__antithesis_instrumentation__.Notify(595046)
		if contains = *id == needle; contains {
			__antithesis_instrumentation__.Notify(595048)
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(595049)
		}
		__antithesis_instrumentation__.Notify(595047)
		return nil
	})
	__antithesis_instrumentation__.Notify(595045)
	return contains
}
