package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "gopkg.in/yaml.v3"

type Clauses []Clause

func flattened(c Clauses) Clauses {
	__antithesis_instrumentation__.Notify(578905)
	hasAnd := func() bool {
		__antithesis_instrumentation__.Notify(578909)
		for _, cl := range c {
			__antithesis_instrumentation__.Notify(578911)
			if _, isAnd := cl.(and); isAnd {
				__antithesis_instrumentation__.Notify(578912)
				return true
			} else {
				__antithesis_instrumentation__.Notify(578913)
			}
		}
		__antithesis_instrumentation__.Notify(578910)
		return false
	}
	__antithesis_instrumentation__.Notify(578906)
	if !hasAnd() {
		__antithesis_instrumentation__.Notify(578914)
		return c
	} else {
		__antithesis_instrumentation__.Notify(578915)
	}
	__antithesis_instrumentation__.Notify(578907)
	var ret Clauses
	for _, cl := range c {
		__antithesis_instrumentation__.Notify(578916)
		switch cl := cl.(type) {
		case and:
			__antithesis_instrumentation__.Notify(578917)
			ret = append(ret, flattened(Clauses(cl))...)
		default:
			__antithesis_instrumentation__.Notify(578918)
			ret = append(ret, cl)
		}
	}
	__antithesis_instrumentation__.Notify(578908)
	return ret
}

func (c Clauses) MarshalYAML() (interface{}, error) {
	__antithesis_instrumentation__.Notify(578919)
	fc := flattened(c)
	var n yaml.Node
	if err := n.Encode([]Clause(fc)); err != nil {
		__antithesis_instrumentation__.Notify(578921)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(578922)
	}
	__antithesis_instrumentation__.Notify(578920)
	n.Style = yaml.LiteralStyle
	return &n, nil
}
