package opgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/errors"
)

type target struct {
	e           scpb.Element
	status      scpb.Status
	transitions []transition
	iterateFunc func(*rel.Database, func(*screl.Node) error) error
}

type transition struct {
	from, to   scpb.Status
	revertible bool
	ops        opsFunc
	minPhase   scop.Phase
}

func makeTarget(e scpb.Element, spec targetSpec) (t target, err error) {
	__antithesis_instrumentation__.Notify(594099)
	defer func() {
		__antithesis_instrumentation__.Notify(594104)
		err = errors.Wrapf(err, "target %s", spec.to)
	}()
	__antithesis_instrumentation__.Notify(594100)
	t = target{
		e:      e,
		status: spec.to,
	}
	t.transitions, err = makeTransitions(e, spec)
	if err != nil {
		__antithesis_instrumentation__.Notify(594105)
		return t, err
	} else {
		__antithesis_instrumentation__.Notify(594106)
	}
	__antithesis_instrumentation__.Notify(594101)

	var element, target, node, targetStatus rel.Var = "element", "target", "node", "target-status"
	q, err := rel.NewQuery(screl.Schema,
		element.Type(e),
		targetStatus.Eq(spec.to),
		screl.JoinTargetNode(element, target, node),
		target.AttrEqVar(screl.TargetStatus, targetStatus),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(594107)
		return t, errors.Wrap(err, "failed to construct query")
	} else {
		__antithesis_instrumentation__.Notify(594108)
	}
	__antithesis_instrumentation__.Notify(594102)
	t.iterateFunc = func(database *rel.Database, f func(*screl.Node) error) error {
		__antithesis_instrumentation__.Notify(594109)
		return q.Iterate(database, func(r rel.Result) error {
			__antithesis_instrumentation__.Notify(594110)
			return f(r.Var(node).(*screl.Node))
		})
	}
	__antithesis_instrumentation__.Notify(594103)

	return t, nil
}

func makeTransitions(e scpb.Element, spec targetSpec) (ret []transition, err error) {
	__antithesis_instrumentation__.Notify(594111)
	tbs := makeTransitionBuildState(spec.from)
	for _, s := range spec.transitionSpecs {
		__antithesis_instrumentation__.Notify(594114)
		var t transition
		if s.from == scpb.Status_UNKNOWN {
			__antithesis_instrumentation__.Notify(594116)
			t.from = tbs.from
			t.to = s.to
			if err := tbs.withTransition(s); err != nil {
				__antithesis_instrumentation__.Notify(594118)
				return nil, errors.Wrapf(err, "invalid transition %s -> %s", t.from, t.to)
			} else {
				__antithesis_instrumentation__.Notify(594119)
			}
			__antithesis_instrumentation__.Notify(594117)
			if len(s.emitFns) > 0 {
				__antithesis_instrumentation__.Notify(594120)
				t.ops, err = makeOpsFunc(e, s.emitFns)
				if err != nil {
					__antithesis_instrumentation__.Notify(594121)
					return nil, errors.Wrapf(err, "making ops func for transition %s -> %s", t.from, t.to)
				} else {
					__antithesis_instrumentation__.Notify(594122)
				}
			} else {
				__antithesis_instrumentation__.Notify(594123)
			}
		} else {
			__antithesis_instrumentation__.Notify(594124)
			t.from = s.from
			t.to = tbs.from
			if err := tbs.withEquivTransition(s); err != nil {
				__antithesis_instrumentation__.Notify(594125)
				return nil, errors.Wrapf(err, "invalid no-op transition %s -> %s", t.from, t.to)
			} else {
				__antithesis_instrumentation__.Notify(594126)
			}
		}
		__antithesis_instrumentation__.Notify(594115)
		t.revertible = tbs.isRevertible
		t.minPhase = tbs.currentMinPhase
		ret = append(ret, t)
	}
	__antithesis_instrumentation__.Notify(594112)

	if tbs.from != spec.to {
		__antithesis_instrumentation__.Notify(594127)
		return nil, errors.Errorf("expected %s as the final status, instead found %s", spec.to, tbs.from)
	} else {
		__antithesis_instrumentation__.Notify(594128)
	}
	__antithesis_instrumentation__.Notify(594113)

	return ret, nil
}

type transitionBuildState struct {
	from            scpb.Status
	currentMinPhase scop.Phase
	isRevertible    bool

	isEquivMapped map[scpb.Status]bool
	isTo          map[scpb.Status]bool
	isFrom        map[scpb.Status]bool
}

func makeTransitionBuildState(from scpb.Status) transitionBuildState {
	__antithesis_instrumentation__.Notify(594129)
	return transitionBuildState{
		from:          from,
		isRevertible:  true,
		isEquivMapped: map[scpb.Status]bool{from: true},
		isTo:          map[scpb.Status]bool{},
		isFrom:        map[scpb.Status]bool{},
	}
}

func (tbs *transitionBuildState) withTransition(s transitionSpec) error {
	__antithesis_instrumentation__.Notify(594130)

	if s.to == scpb.Status_UNKNOWN {
		__antithesis_instrumentation__.Notify(594135)
		return errors.Errorf("invalid 'to' status")
	} else {
		__antithesis_instrumentation__.Notify(594136)
	}
	__antithesis_instrumentation__.Notify(594131)
	if tbs.isTo[s.to] {
		__antithesis_instrumentation__.Notify(594137)
		return errors.Errorf("%s was featured as 'to' in a previous transition", s.to)
	} else {
		__antithesis_instrumentation__.Notify(594138)
		if tbs.isEquivMapped[s.to] {
			__antithesis_instrumentation__.Notify(594139)
			return errors.Errorf("%s was featured as 'from' in a previous equivalence mapping", s.to)
		} else {
			__antithesis_instrumentation__.Notify(594140)
		}
	}
	__antithesis_instrumentation__.Notify(594132)

	if s.minPhase > 0 && func() bool {
		__antithesis_instrumentation__.Notify(594141)
		return s.minPhase < tbs.currentMinPhase == true
	}() == true {
		__antithesis_instrumentation__.Notify(594142)
		return errors.Errorf("minimum phase %s is less than inherited minimum phase %s",
			s.minPhase.String(), tbs.currentMinPhase.String())
	} else {
		__antithesis_instrumentation__.Notify(594143)
	}
	__antithesis_instrumentation__.Notify(594133)

	tbs.isRevertible = tbs.isRevertible && func() bool {
		__antithesis_instrumentation__.Notify(594144)
		return s.revertible == true
	}() == true
	if s.minPhase > tbs.currentMinPhase {
		__antithesis_instrumentation__.Notify(594145)
		tbs.currentMinPhase = s.minPhase
	} else {
		__antithesis_instrumentation__.Notify(594146)
	}
	__antithesis_instrumentation__.Notify(594134)
	tbs.isEquivMapped[tbs.from] = true
	tbs.isTo[s.to] = true
	tbs.isFrom[tbs.from] = true
	tbs.from = s.to
	return nil
}

func (tbs *transitionBuildState) withEquivTransition(s transitionSpec) error {
	__antithesis_instrumentation__.Notify(594147)

	if s.to != scpb.Status_UNKNOWN {
		__antithesis_instrumentation__.Notify(594152)
		return errors.Errorf("invalid 'to' status %s", s.to)
	} else {
		__antithesis_instrumentation__.Notify(594153)
	}
	__antithesis_instrumentation__.Notify(594148)

	if tbs.isTo[s.from] {
		__antithesis_instrumentation__.Notify(594154)
		return errors.Errorf("%s was featured as 'to' in a previous transition", s.from)
	} else {
		__antithesis_instrumentation__.Notify(594155)
		if tbs.isEquivMapped[s.from] {
			__antithesis_instrumentation__.Notify(594156)
			return errors.Errorf("%s was featured as 'from' in a previous equivalence mapping", s.from)
		} else {
			__antithesis_instrumentation__.Notify(594157)
		}
	}
	__antithesis_instrumentation__.Notify(594149)

	if !s.revertible {
		__antithesis_instrumentation__.Notify(594158)
		return errors.Errorf("must be revertible")
	} else {
		__antithesis_instrumentation__.Notify(594159)
	}
	__antithesis_instrumentation__.Notify(594150)
	if s.minPhase > 0 {
		__antithesis_instrumentation__.Notify(594160)
		return errors.Errorf("must not set a minimum phase")
	} else {
		__antithesis_instrumentation__.Notify(594161)
	}
	__antithesis_instrumentation__.Notify(594151)

	tbs.isEquivMapped[s.from] = true
	return nil
}
