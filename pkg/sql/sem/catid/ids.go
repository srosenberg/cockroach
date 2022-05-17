// Package catid is a low-level package exporting ID types.
package catid

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type DescID uint32

const InvalidDescID DescID = 0

func (DescID) SafeValue() { __antithesis_instrumentation__.Notify(602892) }

type ColumnID uint32

func (ColumnID) SafeValue() { __antithesis_instrumentation__.Notify(602893) }

type FamilyID uint32

func (FamilyID) SafeValue() { __antithesis_instrumentation__.Notify(602894) }

type IndexID uint32

func (IndexID) SafeValue() { __antithesis_instrumentation__.Notify(602895) }

type ConstraintID uint32

func (ConstraintID) SafeValue() { __antithesis_instrumentation__.Notify(602896) }
