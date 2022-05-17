package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "sync"

type valuesMap struct {
	attrs ordinalSet
	m     map[ordinal]interface{}
}

var valuesSyncPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(579268)
		return &valuesMap{
			m: make(map[ordinal]interface{}),
		}
	},
}

func getValues() *valuesMap {
	__antithesis_instrumentation__.Notify(579269)
	return valuesSyncPool.Get().(*valuesMap)
}

func putValues(v *valuesMap) {
	__antithesis_instrumentation__.Notify(579270)
	v.clear()
	valuesSyncPool.Put(v)
}

func (vm *valuesMap) clear() {
	__antithesis_instrumentation__.Notify(579271)
	for k := range vm.m {
		__antithesis_instrumentation__.Notify(579273)
		delete(vm.m, k)
	}
	__antithesis_instrumentation__.Notify(579272)
	vm.attrs = 0
}

func (vm valuesMap) get(a ordinal) interface{} {
	__antithesis_instrumentation__.Notify(579274)
	return vm.m[a]
}

func (vm *valuesMap) add(ord ordinal, v interface{}) {
	__antithesis_instrumentation__.Notify(579275)
	vm.attrs = vm.attrs.add(ord)
	vm.m[ord] = v
}
