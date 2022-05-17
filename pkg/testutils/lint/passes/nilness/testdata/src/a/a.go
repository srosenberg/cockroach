package a

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type X struct{ f, g int }

func f(x, y *X) {
	__antithesis_instrumentation__.Notify(644992)
	if x == nil {
		__antithesis_instrumentation__.Notify(644995)
		print(x.f)
	} else {
		__antithesis_instrumentation__.Notify(644996)
		print(x.f)
	}
	__antithesis_instrumentation__.Notify(644993)

	if x == nil {
		__antithesis_instrumentation__.Notify(644997)
		if nil != y {
			__antithesis_instrumentation__.Notify(644999)
			print(1)
			panic(0)
		} else {
			__antithesis_instrumentation__.Notify(645000)
		}
		__antithesis_instrumentation__.Notify(644998)
		x.f = 1
		y.f = 1
	} else {
		__antithesis_instrumentation__.Notify(645001)
	}
	__antithesis_instrumentation__.Notify(644994)

	var f func()
	if f == nil {
		__antithesis_instrumentation__.Notify(645002)
		go f()
	} else {
		__antithesis_instrumentation__.Notify(645003)

		defer f()
	}
}

func f2(ptr *[3]int, i interface{}) {
	__antithesis_instrumentation__.Notify(645004)
	if ptr != nil {
		__antithesis_instrumentation__.Notify(645006)
		print(ptr[:])
		*ptr = [3]int{}
		print(*ptr)
	} else {
		__antithesis_instrumentation__.Notify(645007)
		print(ptr[:])
		*ptr = [3]int{}
		print(*ptr)

		if ptr != nil {
			__antithesis_instrumentation__.Notify(645008)

			print(*ptr)
		} else {
			__antithesis_instrumentation__.Notify(645009)
		}
	}
	__antithesis_instrumentation__.Notify(645005)

	if i != nil {
		__antithesis_instrumentation__.Notify(645010)
		print(i.(interface{ f() }))
	} else {
		__antithesis_instrumentation__.Notify(645011)
		print(i.(interface{ f() }))
	}
}

func g() error

func f3() error {
	__antithesis_instrumentation__.Notify(645012)
	err := g()
	if err != nil {
		__antithesis_instrumentation__.Notify(645017)
		return err
	} else {
		__antithesis_instrumentation__.Notify(645018)
	}
	__antithesis_instrumentation__.Notify(645013)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(645019)
		return err.Error() == "foo" == true
	}() == true {
		__antithesis_instrumentation__.Notify(645020)
		print(0)
	} else {
		__antithesis_instrumentation__.Notify(645021)
	}
	__antithesis_instrumentation__.Notify(645014)
	ch := make(chan int)
	if ch == nil {
		__antithesis_instrumentation__.Notify(645022)
		print(0)
	} else {
		__antithesis_instrumentation__.Notify(645023)
	}
	__antithesis_instrumentation__.Notify(645015)
	if ch != nil {
		__antithesis_instrumentation__.Notify(645024)
		print(0)
	} else {
		__antithesis_instrumentation__.Notify(645025)
	}
	__antithesis_instrumentation__.Notify(645016)
	return nil
}

func h(err error, b bool) {
	__antithesis_instrumentation__.Notify(645026)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(645027)
		return b == true
	}() == true {
		__antithesis_instrumentation__.Notify(645028)
		return
	} else {
		__antithesis_instrumentation__.Notify(645029)
		if err != nil {
			__antithesis_instrumentation__.Notify(645030)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(645031)
		}
	}
}

func i(*int) error {
	__antithesis_instrumentation__.Notify(645032)
	for {
		__antithesis_instrumentation__.Notify(645033)
		if err := g(); err != nil {
			__antithesis_instrumentation__.Notify(645034)
			return err
		} else {
			__antithesis_instrumentation__.Notify(645035)
		}
	}
}

func f4(x *X) {
	__antithesis_instrumentation__.Notify(645036)
	if x == nil {
		__antithesis_instrumentation__.Notify(645037)
		panic(x)
	} else {
		__antithesis_instrumentation__.Notify(645038)
	}
}

func f5(x *X) {
	__antithesis_instrumentation__.Notify(645039)
	panic(nil)
}

func f6(x *X) {
	__antithesis_instrumentation__.Notify(645040)
	var err error
	panic(err)
}

func f7() {
	__antithesis_instrumentation__.Notify(645041)
	x, err := bad()
	if err != nil {
		__antithesis_instrumentation__.Notify(645043)
		panic(0)
	} else {
		__antithesis_instrumentation__.Notify(645044)
	}
	__antithesis_instrumentation__.Notify(645042)
	if x == nil {
		__antithesis_instrumentation__.Notify(645045)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(645046)
	}
}

func bad() (*X, error) {
	__antithesis_instrumentation__.Notify(645047)
	return nil, nil
}

func f8() {
	__antithesis_instrumentation__.Notify(645048)
	var e error
	v, _ := e.(interface{})
	print(v)
}

func f9(x interface {
	a()
	b()
	c()
}) {
	__antithesis_instrumentation__.Notify(645049)

	x.b()
	xx := interface {
		a()
		b()
	}(x)
	if xx != nil {
		__antithesis_instrumentation__.Notify(645053)
		return
	} else {
		__antithesis_instrumentation__.Notify(645054)
	}
	__antithesis_instrumentation__.Notify(645050)
	x.c()
	xx.b()
	xxx := interface{ a() }(xx)
	xxx.a()

	if unknown() {
		__antithesis_instrumentation__.Notify(645055)
		panic(x)
	} else {
		__antithesis_instrumentation__.Notify(645056)
	}
	__antithesis_instrumentation__.Notify(645051)
	if unknown() {
		__antithesis_instrumentation__.Notify(645057)
		panic(xx)
	} else {
		__antithesis_instrumentation__.Notify(645058)
	}
	__antithesis_instrumentation__.Notify(645052)
	if unknown() {
		__antithesis_instrumentation__.Notify(645059)
		panic(xxx)
	} else {
		__antithesis_instrumentation__.Notify(645060)
	}
}

func unknown() bool {
	__antithesis_instrumentation__.Notify(645061)
	return false
}
