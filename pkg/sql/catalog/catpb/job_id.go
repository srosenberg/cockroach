package catpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type JobID int64

const InvalidJobID JobID = 0

func (j JobID) SafeValue() { __antithesis_instrumentation__.Notify(249369) }
