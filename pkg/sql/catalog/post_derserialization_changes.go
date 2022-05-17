package catalog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util"

type PostDeserializationChangeType int

type PostDeserializationChanges struct{ s util.FastIntSet }

func (c PostDeserializationChanges) HasChanges() bool {
	__antithesis_instrumentation__.Notify(267151)
	return !c.s.Empty()
}

func (c *PostDeserializationChanges) Add(change PostDeserializationChangeType) {
	__antithesis_instrumentation__.Notify(267152)
	c.s.Add(int(change))
}

func (c PostDeserializationChanges) ForEach(f func(change PostDeserializationChangeType)) {
	__antithesis_instrumentation__.Notify(267153)
	c.s.ForEach(func(i int) { __antithesis_instrumentation__.Notify(267154); f(PostDeserializationChangeType(i)) })
}

func (c PostDeserializationChanges) Contains(change PostDeserializationChangeType) bool {
	__antithesis_instrumentation__.Notify(267155)
	return c.s.Contains(int(change))
}

const (
	UpgradedFormatVersion PostDeserializationChangeType = iota

	FixedIndexEncodingType

	UpgradedIndexFormatVersion

	UpgradedForeignKeyRepresentation

	UpgradedNamespaceName

	UpgradedPrivileges

	RemovedDefaultExprFromComputedColumn

	RemovedDuplicateIDsInRefs

	AddedConstraintIDs

	RemovedSelfEntryInSchemas

	FixedDateStyleIntervalStyleCast
)
