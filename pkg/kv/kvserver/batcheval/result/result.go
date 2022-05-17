package result

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/kr/pretty"
)

type LocalResult struct {
	Reply *roachpb.BatchResponse

	EncounteredIntents []roachpb.Intent

	AcquiredLocks []roachpb.LockAcquisition

	ResolvedLocks []roachpb.LockUpdate

	UpdatedTxns []*roachpb.Transaction

	EndTxns []EndTxnIntents

	GossipFirstRange bool

	MaybeGossipSystemConfig bool

	MaybeGossipSystemConfigIfHaveFailure bool

	MaybeAddToSplitQueue bool

	MaybeGossipNodeLiveness *roachpb.Span

	Metrics *Metrics
}

func (lResult *LocalResult) IsZero() bool {
	__antithesis_instrumentation__.Notify(97671)

	return lResult.Reply == nil && func() bool {
		__antithesis_instrumentation__.Notify(97672)
		return lResult.EncounteredIntents == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(97673)
		return lResult.AcquiredLocks == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(97674)
		return lResult.ResolvedLocks == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(97675)
		return lResult.UpdatedTxns == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(97676)
		return lResult.EndTxns == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(97677)
		return !lResult.GossipFirstRange == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(97678)
		return !lResult.MaybeGossipSystemConfig == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(97679)
		return !lResult.MaybeGossipSystemConfigIfHaveFailure == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(97680)
		return lResult.MaybeGossipNodeLiveness == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(97681)
		return lResult.Metrics == nil == true
	}() == true
}

func (lResult *LocalResult) String() string {
	__antithesis_instrumentation__.Notify(97682)
	if lResult == nil {
		__antithesis_instrumentation__.Notify(97684)
		return "LocalResult: nil"
	} else {
		__antithesis_instrumentation__.Notify(97685)
	}
	__antithesis_instrumentation__.Notify(97683)
	return fmt.Sprintf("LocalResult (reply: %v, "+
		"#encountered intents: %d, #acquired locks: %d, #resolved locks: %d"+
		"#updated txns: %d #end txns: %d, "+
		"GossipFirstRange:%t MaybeGossipSystemConfig:%t "+
		"MaybeGossipSystemConfigIfHaveFailure:%t MaybeAddToSplitQueue:%t "+
		"MaybeGossipNodeLiveness:%s ",
		lResult.Reply,
		len(lResult.EncounteredIntents), len(lResult.AcquiredLocks), len(lResult.ResolvedLocks),
		len(lResult.UpdatedTxns), len(lResult.EndTxns),
		lResult.GossipFirstRange, lResult.MaybeGossipSystemConfig,
		lResult.MaybeGossipSystemConfigIfHaveFailure, lResult.MaybeAddToSplitQueue,
		lResult.MaybeGossipNodeLiveness)
}

func (lResult *LocalResult) RequiresRaft() bool {
	__antithesis_instrumentation__.Notify(97686)

	return lResult.MaybeGossipNodeLiveness != nil || func() bool {
		__antithesis_instrumentation__.Notify(97687)
		return lResult.MaybeGossipSystemConfig == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(97688)
		return lResult.MaybeGossipSystemConfigIfHaveFailure == true
	}() == true
}

func (lResult *LocalResult) DetachEncounteredIntents() []roachpb.Intent {
	__antithesis_instrumentation__.Notify(97689)
	if lResult == nil {
		__antithesis_instrumentation__.Notify(97691)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(97692)
	}
	__antithesis_instrumentation__.Notify(97690)
	r := lResult.EncounteredIntents
	lResult.EncounteredIntents = nil
	return r
}

func (lResult *LocalResult) DetachEndTxns(alwaysOnly bool) []EndTxnIntents {
	__antithesis_instrumentation__.Notify(97693)
	if lResult == nil {
		__antithesis_instrumentation__.Notify(97696)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(97697)
	}
	__antithesis_instrumentation__.Notify(97694)
	r := lResult.EndTxns
	if alwaysOnly {
		__antithesis_instrumentation__.Notify(97698)

		r = r[:0]
		for _, eti := range lResult.EndTxns {
			__antithesis_instrumentation__.Notify(97699)
			if eti.Always {
				__antithesis_instrumentation__.Notify(97700)
				r = append(r, eti)
			} else {
				__antithesis_instrumentation__.Notify(97701)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(97702)
	}
	__antithesis_instrumentation__.Notify(97695)
	lResult.EndTxns = nil
	return r
}

type Result struct {
	Local        LocalResult
	Replicated   kvserverpb.ReplicatedEvalResult
	WriteBatch   *kvserverpb.WriteBatch
	LogicalOpLog *kvserverpb.LogicalOpLog
}

func (p *Result) IsZero() bool {
	__antithesis_instrumentation__.Notify(97703)
	if !p.Local.IsZero() {
		__antithesis_instrumentation__.Notify(97708)
		return false
	} else {
		__antithesis_instrumentation__.Notify(97709)
	}
	__antithesis_instrumentation__.Notify(97704)
	if !p.Replicated.IsZero() {
		__antithesis_instrumentation__.Notify(97710)
		return false
	} else {
		__antithesis_instrumentation__.Notify(97711)
	}
	__antithesis_instrumentation__.Notify(97705)
	if p.WriteBatch != nil {
		__antithesis_instrumentation__.Notify(97712)
		return false
	} else {
		__antithesis_instrumentation__.Notify(97713)
	}
	__antithesis_instrumentation__.Notify(97706)
	if p.LogicalOpLog != nil {
		__antithesis_instrumentation__.Notify(97714)
		return false
	} else {
		__antithesis_instrumentation__.Notify(97715)
	}
	__antithesis_instrumentation__.Notify(97707)
	return true
}

func coalesceBool(lhs *bool, rhs *bool) {
	__antithesis_instrumentation__.Notify(97716)
	*lhs = *lhs || func() bool {
		__antithesis_instrumentation__.Notify(97717)
		return *rhs == true
	}() == true
	*rhs = false
}

func (p *Result) MergeAndDestroy(q Result) error {
	__antithesis_instrumentation__.Notify(97718)
	if q.Replicated.State != nil {
		__antithesis_instrumentation__.Notify(97739)
		if q.Replicated.State.RaftAppliedIndex != 0 {
			__antithesis_instrumentation__.Notify(97751)
			return errors.AssertionFailedf("must not specify RaftAppliedIndex")
		} else {
			__antithesis_instrumentation__.Notify(97752)
		}
		__antithesis_instrumentation__.Notify(97740)
		if q.Replicated.State.LeaseAppliedIndex != 0 {
			__antithesis_instrumentation__.Notify(97753)
			return errors.AssertionFailedf("must not specify LeaseAppliedIndex")
		} else {
			__antithesis_instrumentation__.Notify(97754)
		}
		__antithesis_instrumentation__.Notify(97741)
		if p.Replicated.State == nil {
			__antithesis_instrumentation__.Notify(97755)
			p.Replicated.State = &kvserverpb.ReplicaState{}
		} else {
			__antithesis_instrumentation__.Notify(97756)
		}
		__antithesis_instrumentation__.Notify(97742)
		if q.Replicated.State.RaftAppliedIndexTerm != 0 {
			__antithesis_instrumentation__.Notify(97757)
			if q.Replicated.State.RaftAppliedIndexTerm ==
				stateloader.RaftLogTermSignalForAddRaftAppliedIndexTermMigration {
				__antithesis_instrumentation__.Notify(97758)
				if p.Replicated.State.RaftAppliedIndexTerm != 0 && func() bool {
					__antithesis_instrumentation__.Notify(97760)
					return p.Replicated.State.RaftAppliedIndexTerm !=
						stateloader.RaftLogTermSignalForAddRaftAppliedIndexTermMigration == true
				}() == true {
					__antithesis_instrumentation__.Notify(97761)
					return errors.AssertionFailedf("invalid term value %d",
						p.Replicated.State.RaftAppliedIndexTerm)
				} else {
					__antithesis_instrumentation__.Notify(97762)
				}
				__antithesis_instrumentation__.Notify(97759)
				p.Replicated.State.RaftAppliedIndexTerm = q.Replicated.State.RaftAppliedIndexTerm
				q.Replicated.State.RaftAppliedIndexTerm = 0
			} else {
				__antithesis_instrumentation__.Notify(97763)
				return errors.AssertionFailedf("invalid term value %d",
					q.Replicated.State.RaftAppliedIndexTerm)
			}
		} else {
			__antithesis_instrumentation__.Notify(97764)
		}
		__antithesis_instrumentation__.Notify(97743)
		if p.Replicated.State.Desc == nil {
			__antithesis_instrumentation__.Notify(97765)
			p.Replicated.State.Desc = q.Replicated.State.Desc
		} else {
			__antithesis_instrumentation__.Notify(97766)
			if q.Replicated.State.Desc != nil {
				__antithesis_instrumentation__.Notify(97767)
				return errors.AssertionFailedf("conflicting RangeDescriptor")
			} else {
				__antithesis_instrumentation__.Notify(97768)
			}
		}
		__antithesis_instrumentation__.Notify(97744)
		q.Replicated.State.Desc = nil

		if p.Replicated.State.Lease == nil {
			__antithesis_instrumentation__.Notify(97769)
			p.Replicated.State.Lease = q.Replicated.State.Lease
		} else {
			__antithesis_instrumentation__.Notify(97770)
			if q.Replicated.State.Lease != nil {
				__antithesis_instrumentation__.Notify(97771)
				return errors.AssertionFailedf("conflicting Lease")
			} else {
				__antithesis_instrumentation__.Notify(97772)
			}
		}
		__antithesis_instrumentation__.Notify(97745)
		q.Replicated.State.Lease = nil

		if p.Replicated.State.TruncatedState == nil {
			__antithesis_instrumentation__.Notify(97773)
			p.Replicated.State.TruncatedState = q.Replicated.State.TruncatedState
			p.Replicated.RaftExpectedFirstIndex = q.Replicated.RaftExpectedFirstIndex
		} else {
			__antithesis_instrumentation__.Notify(97774)
			if q.Replicated.State.TruncatedState != nil {
				__antithesis_instrumentation__.Notify(97775)
				return errors.AssertionFailedf("conflicting TruncatedState")
			} else {
				__antithesis_instrumentation__.Notify(97776)
			}
		}
		__antithesis_instrumentation__.Notify(97746)
		q.Replicated.State.TruncatedState = nil
		q.Replicated.RaftExpectedFirstIndex = 0

		if q.Replicated.State.GCThreshold != nil {
			__antithesis_instrumentation__.Notify(97777)
			if p.Replicated.State.GCThreshold == nil {
				__antithesis_instrumentation__.Notify(97779)
				p.Replicated.State.GCThreshold = q.Replicated.State.GCThreshold
			} else {
				__antithesis_instrumentation__.Notify(97780)
				p.Replicated.State.GCThreshold.Forward(*q.Replicated.State.GCThreshold)
			}
			__antithesis_instrumentation__.Notify(97778)
			q.Replicated.State.GCThreshold = nil
		} else {
			__antithesis_instrumentation__.Notify(97781)
		}
		__antithesis_instrumentation__.Notify(97747)

		if p.Replicated.State.Version == nil {
			__antithesis_instrumentation__.Notify(97782)
			p.Replicated.State.Version = q.Replicated.State.Version
		} else {
			__antithesis_instrumentation__.Notify(97783)
			if q.Replicated.State.Version != nil {
				__antithesis_instrumentation__.Notify(97784)
				return errors.AssertionFailedf("conflicting Version")
			} else {
				__antithesis_instrumentation__.Notify(97785)
			}
		}
		__antithesis_instrumentation__.Notify(97748)
		q.Replicated.State.Version = nil

		if q.Replicated.State.Stats != nil {
			__antithesis_instrumentation__.Notify(97786)
			return errors.AssertionFailedf("must not specify Stats")
		} else {
			__antithesis_instrumentation__.Notify(97787)
		}
		__antithesis_instrumentation__.Notify(97749)
		if (*q.Replicated.State != kvserverpb.ReplicaState{}) {
			__antithesis_instrumentation__.Notify(97788)
			log.Fatalf(context.TODO(), "unhandled EvalResult: %s",
				pretty.Diff(*q.Replicated.State, kvserverpb.ReplicaState{}))
		} else {
			__antithesis_instrumentation__.Notify(97789)
		}
		__antithesis_instrumentation__.Notify(97750)
		q.Replicated.State = nil
	} else {
		__antithesis_instrumentation__.Notify(97790)
	}
	__antithesis_instrumentation__.Notify(97719)

	if p.Replicated.Split == nil {
		__antithesis_instrumentation__.Notify(97791)
		p.Replicated.Split = q.Replicated.Split
	} else {
		__antithesis_instrumentation__.Notify(97792)
		if q.Replicated.Split != nil {
			__antithesis_instrumentation__.Notify(97793)
			return errors.AssertionFailedf("conflicting Split")
		} else {
			__antithesis_instrumentation__.Notify(97794)
		}
	}
	__antithesis_instrumentation__.Notify(97720)
	q.Replicated.Split = nil

	if p.Replicated.Merge == nil {
		__antithesis_instrumentation__.Notify(97795)
		p.Replicated.Merge = q.Replicated.Merge
	} else {
		__antithesis_instrumentation__.Notify(97796)
		if q.Replicated.Merge != nil {
			__antithesis_instrumentation__.Notify(97797)
			return errors.AssertionFailedf("conflicting Merge")
		} else {
			__antithesis_instrumentation__.Notify(97798)
		}
	}
	__antithesis_instrumentation__.Notify(97721)
	q.Replicated.Merge = nil

	if p.Replicated.ChangeReplicas == nil {
		__antithesis_instrumentation__.Notify(97799)
		p.Replicated.ChangeReplicas = q.Replicated.ChangeReplicas
	} else {
		__antithesis_instrumentation__.Notify(97800)
		if q.Replicated.ChangeReplicas != nil {
			__antithesis_instrumentation__.Notify(97801)
			return errors.AssertionFailedf("conflicting ChangeReplicas")
		} else {
			__antithesis_instrumentation__.Notify(97802)
		}
	}
	__antithesis_instrumentation__.Notify(97722)
	q.Replicated.ChangeReplicas = nil

	if p.Replicated.ComputeChecksum == nil {
		__antithesis_instrumentation__.Notify(97803)
		p.Replicated.ComputeChecksum = q.Replicated.ComputeChecksum
	} else {
		__antithesis_instrumentation__.Notify(97804)
		if q.Replicated.ComputeChecksum != nil {
			__antithesis_instrumentation__.Notify(97805)
			return errors.AssertionFailedf("conflicting ComputeChecksum")
		} else {
			__antithesis_instrumentation__.Notify(97806)
		}
	}
	__antithesis_instrumentation__.Notify(97723)
	q.Replicated.ComputeChecksum = nil

	if p.Replicated.RaftLogDelta == 0 {
		__antithesis_instrumentation__.Notify(97807)
		p.Replicated.RaftLogDelta = q.Replicated.RaftLogDelta
	} else {
		__antithesis_instrumentation__.Notify(97808)
		if q.Replicated.RaftLogDelta != 0 {
			__antithesis_instrumentation__.Notify(97809)
			return errors.AssertionFailedf("conflicting RaftLogDelta")
		} else {
			__antithesis_instrumentation__.Notify(97810)
		}
	}
	__antithesis_instrumentation__.Notify(97724)
	q.Replicated.RaftLogDelta = 0

	if p.Replicated.AddSSTable == nil {
		__antithesis_instrumentation__.Notify(97811)
		p.Replicated.AddSSTable = q.Replicated.AddSSTable
	} else {
		__antithesis_instrumentation__.Notify(97812)
		if q.Replicated.AddSSTable != nil {
			__antithesis_instrumentation__.Notify(97813)
			return errors.AssertionFailedf("conflicting AddSSTable")
		} else {
			__antithesis_instrumentation__.Notify(97814)
		}
	}
	__antithesis_instrumentation__.Notify(97725)
	q.Replicated.AddSSTable = nil

	if p.Replicated.MVCCHistoryMutation == nil {
		__antithesis_instrumentation__.Notify(97815)
		p.Replicated.MVCCHistoryMutation = q.Replicated.MVCCHistoryMutation
	} else {
		__antithesis_instrumentation__.Notify(97816)
		if q.Replicated.MVCCHistoryMutation != nil {
			__antithesis_instrumentation__.Notify(97817)
			p.Replicated.MVCCHistoryMutation.Spans = append(p.Replicated.MVCCHistoryMutation.Spans,
				q.Replicated.MVCCHistoryMutation.Spans...)
		} else {
			__antithesis_instrumentation__.Notify(97818)
		}
	}
	__antithesis_instrumentation__.Notify(97726)
	q.Replicated.MVCCHistoryMutation = nil

	if p.Replicated.PrevLeaseProposal == nil {
		__antithesis_instrumentation__.Notify(97819)
		p.Replicated.PrevLeaseProposal = q.Replicated.PrevLeaseProposal
	} else {
		__antithesis_instrumentation__.Notify(97820)
		if q.Replicated.PrevLeaseProposal != nil {
			__antithesis_instrumentation__.Notify(97821)
			return errors.AssertionFailedf("conflicting lease expiration")
		} else {
			__antithesis_instrumentation__.Notify(97822)
		}
	}
	__antithesis_instrumentation__.Notify(97727)
	q.Replicated.PrevLeaseProposal = nil

	if p.Replicated.PriorReadSummary == nil {
		__antithesis_instrumentation__.Notify(97823)
		p.Replicated.PriorReadSummary = q.Replicated.PriorReadSummary
	} else {
		__antithesis_instrumentation__.Notify(97824)
		if q.Replicated.PriorReadSummary != nil {
			__antithesis_instrumentation__.Notify(97825)
			return errors.AssertionFailedf("conflicting prior read summary")
		} else {
			__antithesis_instrumentation__.Notify(97826)
		}
	}
	__antithesis_instrumentation__.Notify(97728)
	q.Replicated.PriorReadSummary = nil

	if !p.Replicated.IsProbe {
		__antithesis_instrumentation__.Notify(97827)
		p.Replicated.IsProbe = q.Replicated.IsProbe
	} else {
		__antithesis_instrumentation__.Notify(97828)
	}
	__antithesis_instrumentation__.Notify(97729)
	q.Replicated.IsProbe = false

	if p.Local.EncounteredIntents == nil {
		__antithesis_instrumentation__.Notify(97829)
		p.Local.EncounteredIntents = q.Local.EncounteredIntents
	} else {
		__antithesis_instrumentation__.Notify(97830)
		p.Local.EncounteredIntents = append(p.Local.EncounteredIntents, q.Local.EncounteredIntents...)
	}
	__antithesis_instrumentation__.Notify(97730)
	q.Local.EncounteredIntents = nil

	if p.Local.AcquiredLocks == nil {
		__antithesis_instrumentation__.Notify(97831)
		p.Local.AcquiredLocks = q.Local.AcquiredLocks
	} else {
		__antithesis_instrumentation__.Notify(97832)
		p.Local.AcquiredLocks = append(p.Local.AcquiredLocks, q.Local.AcquiredLocks...)
	}
	__antithesis_instrumentation__.Notify(97731)
	q.Local.AcquiredLocks = nil

	if p.Local.ResolvedLocks == nil {
		__antithesis_instrumentation__.Notify(97833)
		p.Local.ResolvedLocks = q.Local.ResolvedLocks
	} else {
		__antithesis_instrumentation__.Notify(97834)
		p.Local.ResolvedLocks = append(p.Local.ResolvedLocks, q.Local.ResolvedLocks...)
	}
	__antithesis_instrumentation__.Notify(97732)
	q.Local.ResolvedLocks = nil

	if p.Local.UpdatedTxns == nil {
		__antithesis_instrumentation__.Notify(97835)
		p.Local.UpdatedTxns = q.Local.UpdatedTxns
	} else {
		__antithesis_instrumentation__.Notify(97836)
		p.Local.UpdatedTxns = append(p.Local.UpdatedTxns, q.Local.UpdatedTxns...)
	}
	__antithesis_instrumentation__.Notify(97733)
	q.Local.UpdatedTxns = nil

	if p.Local.EndTxns == nil {
		__antithesis_instrumentation__.Notify(97837)
		p.Local.EndTxns = q.Local.EndTxns
	} else {
		__antithesis_instrumentation__.Notify(97838)
		p.Local.EndTxns = append(p.Local.EndTxns, q.Local.EndTxns...)
	}
	__antithesis_instrumentation__.Notify(97734)
	q.Local.EndTxns = nil

	if p.Local.MaybeGossipNodeLiveness == nil {
		__antithesis_instrumentation__.Notify(97839)
		p.Local.MaybeGossipNodeLiveness = q.Local.MaybeGossipNodeLiveness
	} else {
		__antithesis_instrumentation__.Notify(97840)
		if q.Local.MaybeGossipNodeLiveness != nil {
			__antithesis_instrumentation__.Notify(97841)
			return errors.AssertionFailedf("conflicting MaybeGossipNodeLiveness")
		} else {
			__antithesis_instrumentation__.Notify(97842)
		}
	}
	__antithesis_instrumentation__.Notify(97735)
	q.Local.MaybeGossipNodeLiveness = nil

	coalesceBool(&p.Local.GossipFirstRange, &q.Local.GossipFirstRange)
	coalesceBool(&p.Local.MaybeGossipSystemConfig, &q.Local.MaybeGossipSystemConfig)
	coalesceBool(&p.Local.MaybeGossipSystemConfigIfHaveFailure, &q.Local.MaybeGossipSystemConfigIfHaveFailure)
	coalesceBool(&p.Local.MaybeAddToSplitQueue, &q.Local.MaybeAddToSplitQueue)

	if p.Local.Metrics == nil {
		__antithesis_instrumentation__.Notify(97843)
		p.Local.Metrics = q.Local.Metrics
	} else {
		__antithesis_instrumentation__.Notify(97844)
		if q.Local.Metrics != nil {
			__antithesis_instrumentation__.Notify(97845)
			p.Local.Metrics.Add(*q.Local.Metrics)
		} else {
			__antithesis_instrumentation__.Notify(97846)
		}
	}
	__antithesis_instrumentation__.Notify(97736)
	q.Local.Metrics = nil

	if q.LogicalOpLog != nil {
		__antithesis_instrumentation__.Notify(97847)
		if p.LogicalOpLog == nil {
			__antithesis_instrumentation__.Notify(97848)
			p.LogicalOpLog = q.LogicalOpLog
		} else {
			__antithesis_instrumentation__.Notify(97849)
			p.LogicalOpLog.Ops = append(p.LogicalOpLog.Ops, q.LogicalOpLog.Ops...)
		}
	} else {
		__antithesis_instrumentation__.Notify(97850)
	}
	__antithesis_instrumentation__.Notify(97737)
	q.LogicalOpLog = nil

	if !q.IsZero() {
		__antithesis_instrumentation__.Notify(97851)
		log.Fatalf(context.TODO(), "unhandled EvalResult: %s", pretty.Diff(q, Result{}))
	} else {
		__antithesis_instrumentation__.Notify(97852)
	}
	__antithesis_instrumentation__.Notify(97738)

	return nil
}
