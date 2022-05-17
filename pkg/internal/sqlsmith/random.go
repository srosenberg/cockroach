package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func (s *Smither) coin() bool {
	__antithesis_instrumentation__.Notify(68971)
	return s.rnd.Intn(2) == 0
}

func (s *Smither) d6() int {
	__antithesis_instrumentation__.Notify(68972)
	return s.rnd.Intn(6) + 1
}

func (s *Smither) d9() int {
	__antithesis_instrumentation__.Notify(68973)
	return s.rnd.Intn(9) + 1
}

func (s *Smither) d100() int {
	__antithesis_instrumentation__.Notify(68974)
	return s.rnd.Intn(100) + 1
}

func (s *Smither) geom(p float64) int {
	__antithesis_instrumentation__.Notify(68975)
	if p <= 0 || func() bool {
		__antithesis_instrumentation__.Notify(68978)
		return p >= 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(68979)
		panic("bad p")
	} else {
		__antithesis_instrumentation__.Notify(68980)
	}
	__antithesis_instrumentation__.Notify(68976)
	count := 1
	for s.rnd.Float64() < p {
		__antithesis_instrumentation__.Notify(68981)
		count++
	}
	__antithesis_instrumentation__.Notify(68977)
	return count
}

func (s *Smither) sample(n, mean int, fn func(i int)) {
	__antithesis_instrumentation__.Notify(68982)
	if n <= 0 {
		__antithesis_instrumentation__.Notify(68985)
		return
	} else {
		__antithesis_instrumentation__.Notify(68986)
	}
	__antithesis_instrumentation__.Notify(68983)
	perms := s.rnd.Perm(n)
	m := float64(mean)
	p := (m - 1) / m
	k := s.geom(p)
	if k > n {
		__antithesis_instrumentation__.Notify(68987)
		k = n
	} else {
		__antithesis_instrumentation__.Notify(68988)
	}
	__antithesis_instrumentation__.Notify(68984)
	for ki := 0; ki < k; ki++ {
		__antithesis_instrumentation__.Notify(68989)
		fn(perms[ki])
	}
}
