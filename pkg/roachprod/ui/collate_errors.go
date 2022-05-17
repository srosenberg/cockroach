package ui

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type ErrorsByError []error

func (l ErrorsByError) Len() int {
	__antithesis_instrumentation__.Notify(182539)
	return len(l)
}
func (l ErrorsByError) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(182540)
	return l[i].Error() < l[j].Error()
}
func (l ErrorsByError) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(182541)
	l[i], l[j] = l[j], l[i]
}
