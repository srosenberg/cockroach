package a

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"errors"
	"fmt"
	"log"
)

func noReturn() {
	__antithesis_instrumentation__.Notify(645232)
	err := errors.New("foo")
	if err != nil {
		__antithesis_instrumentation__.Notify(645235)
		return
	} else {
		__antithesis_instrumentation__.Notify(645236)
	}
	__antithesis_instrumentation__.Notify(645233)
	_ = func() error {
		__antithesis_instrumentation__.Notify(645237)
		if err != nil {
			__antithesis_instrumentation__.Notify(645239)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(645240)
		}
		__antithesis_instrumentation__.Notify(645238)
		return nil
	}()
	__antithesis_instrumentation__.Notify(645234)
	return
}

func SingleReturn() error {
	__antithesis_instrumentation__.Notify(645241)
	err := errors.New("foo")
	if err != nil {
		__antithesis_instrumentation__.Notify(645255)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(645256)
	}
	__antithesis_instrumentation__.Notify(645242)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(645257)
		return false == true
	}() == true {
		__antithesis_instrumentation__.Notify(645258)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(645259)
	}
	__antithesis_instrumentation__.Notify(645243)
	if false || func() bool {
		__antithesis_instrumentation__.Notify(645260)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(645261)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(645262)
	}
	__antithesis_instrumentation__.Notify(645244)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(645263)
		return false == true
	}() == true {
		__antithesis_instrumentation__.Notify(645264)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(645265)
	}
	__antithesis_instrumentation__.Notify(645245)
	if false && func() bool {
		__antithesis_instrumentation__.Notify(645266)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(645267)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(645268)
	}
	__antithesis_instrumentation__.Notify(645246)
	if false || func() bool {
		__antithesis_instrumentation__.Notify(645269)
		return (err != nil && func() bool {
			__antithesis_instrumentation__.Notify(645270)
			return false == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(645271)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(645272)
	}
	__antithesis_instrumentation__.Notify(645247)
	if true && func() bool {
		__antithesis_instrumentation__.Notify(645273)
		return false == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(645274)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(645275)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(645276)
	}
	__antithesis_instrumentation__.Notify(645248)
	if err != nil {
		__antithesis_instrumentation__.Notify(645277)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(645278)
	}
	__antithesis_instrumentation__.Notify(645249)
	if err != nil {
		__antithesis_instrumentation__.Notify(645279)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(645280)
	}
	__antithesis_instrumentation__.Notify(645250)
	if err != nil {
		__antithesis_instrumentation__.Notify(645281)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(645282)
	}
	__antithesis_instrumentation__.Notify(645251)

	if err != nil {
		__antithesis_instrumentation__.Notify(645283)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(645284)
	}
	__antithesis_instrumentation__.Notify(645252)

	if err != nil {
		__antithesis_instrumentation__.Notify(645285)

		fmt.Println("hmm")

		return nil
	} else {
		__antithesis_instrumentation__.Notify(645286)
	}
	__antithesis_instrumentation__.Notify(645253)
	if err != nil {
		__antithesis_instrumentation__.Notify(645287)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(645288)
	}
	__antithesis_instrumentation__.Notify(645254)
	return nil
}

func MultipleReturns() (int, error) {
	__antithesis_instrumentation__.Notify(645289)
	err := errors.New("foo")
	if err != nil {
		__antithesis_instrumentation__.Notify(645292)
		if true {
			__antithesis_instrumentation__.Notify(645294)
			return 0, nil
		} else {
			__antithesis_instrumentation__.Notify(645295)
		}
		__antithesis_instrumentation__.Notify(645293)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(645296)
	}
	__antithesis_instrumentation__.Notify(645290)
	if err != nil {
		__antithesis_instrumentation__.Notify(645297)

		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(645298)
	}
	__antithesis_instrumentation__.Notify(645291)
	return -1, nil
}

func AcceptableBehavior() {
	__antithesis_instrumentation__.Notify(645299)
	type structThing struct {
		err error
	}
	_ = func() error {
		__antithesis_instrumentation__.Notify(645305)
		var t structThing

		if t.err != nil {
			__antithesis_instrumentation__.Notify(645307)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(645308)
		}
		__antithesis_instrumentation__.Notify(645306)
		return nil
	}
	__antithesis_instrumentation__.Notify(645300)
	_ = func() error {
		__antithesis_instrumentation__.Notify(645309)
		if err := errors.New("foo"); err != nil {
			__antithesis_instrumentation__.Notify(645311)
			log.Printf("%v", err)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(645312)
		}
		__antithesis_instrumentation__.Notify(645310)
		return nil
	}
	__antithesis_instrumentation__.Notify(645301)
	var t structThing
	_ = func() error {
		__antithesis_instrumentation__.Notify(645313)
		if err := errors.New("foo"); err != nil {
			__antithesis_instrumentation__.Notify(645315)
			t.err = err
			return nil
		} else {
			__antithesis_instrumentation__.Notify(645316)
		}
		__antithesis_instrumentation__.Notify(645314)
		return nil
	}
	__antithesis_instrumentation__.Notify(645302)
	func() error {
		__antithesis_instrumentation__.Notify(645317)
		if err := errors.New("foo"); err != nil {
			__antithesis_instrumentation__.Notify(645319)
			t = structThing{err: err}
			return nil
		} else {
			__antithesis_instrumentation__.Notify(645320)
		}
		__antithesis_instrumentation__.Notify(645318)
		return nil
	}()
	__antithesis_instrumentation__.Notify(645303)
	func() error {
		__antithesis_instrumentation__.Notify(645321)
		if err := errors.New("foo"); err != nil || func() bool {
			__antithesis_instrumentation__.Notify(645323)
			return true == true
		}() == true {
			__antithesis_instrumentation__.Notify(645324)
			if err != nil {
				__antithesis_instrumentation__.Notify(645326)
				return err
			} else {
				__antithesis_instrumentation__.Notify(645327)
			}
			__antithesis_instrumentation__.Notify(645325)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(645328)
		}
		__antithesis_instrumentation__.Notify(645322)
		return nil
	}()
	__antithesis_instrumentation__.Notify(645304)
	func() (string, error) {
		__antithesis_instrumentation__.Notify(645329)
		if err := errors.New("foo"); err != nil {
			__antithesis_instrumentation__.Notify(645331)
			return err.Error(), nil
		} else {
			__antithesis_instrumentation__.Notify(645332)
		}
		__antithesis_instrumentation__.Notify(645330)
		return "", nil
	}()
}
