package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Persistence int

const (
	PersistencePermanent Persistence = iota

	PersistenceTemporary

	PersistenceUnlogged
)

func (p Persistence) IsTemporary() bool {
	__antithesis_instrumentation__.Notify(611665)
	return p == PersistenceTemporary
}

func (p Persistence) IsUnlogged() bool {
	__antithesis_instrumentation__.Notify(611666)
	return p == PersistenceUnlogged
}
