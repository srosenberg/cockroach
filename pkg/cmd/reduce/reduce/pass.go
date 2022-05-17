package reduce

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type intPass struct {
	name string
	fn   func(string, int) (string, bool, error)
}

func MakeIntPass(name string, f func(s string, i int) (out string, ok bool, err error)) Pass {
	__antithesis_instrumentation__.Notify(41827)
	return intPass{
		name: name,
		fn:   f,
	}
}

func (p intPass) Name() string {
	__antithesis_instrumentation__.Notify(41828)
	return p.name
}

func (p intPass) New(string) State {
	__antithesis_instrumentation__.Notify(41829)
	return 0
}

func (p intPass) Transform(f string, s State) (string, Result, error) {
	__antithesis_instrumentation__.Notify(41830)
	i := s.(int)
	data, ok, err := p.fn(f, i)
	res := OK
	if !ok {
		__antithesis_instrumentation__.Notify(41832)
		res = STOP
	} else {
		__antithesis_instrumentation__.Notify(41833)
	}
	__antithesis_instrumentation__.Notify(41831)
	return data, res, err
}

func (p intPass) Advance(f string, s State) State {
	__antithesis_instrumentation__.Notify(41834)
	return s.(int) + 1
}
