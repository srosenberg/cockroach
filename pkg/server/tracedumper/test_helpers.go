package tracedumper

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func TestingGetTraceDumpDir(td *TraceDumper) string {
	__antithesis_instrumentation__.Notify(239503)
	return td.store.GetFullPath("")
}
