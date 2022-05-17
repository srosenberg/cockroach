package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/errors"
)

type createdSequences interface {
	addCreatedSequence(id descpb.ID) error

	isCreatedSequence(id descpb.ID) bool
}

type connExCreatedSequencesAccessor struct {
	ex *connExecutor
}

func (c connExCreatedSequencesAccessor) addCreatedSequence(id descpb.ID) error {
	__antithesis_instrumentation__.Notify(465235)
	c.ex.extraTxnState.createdSequences[id] = struct{}{}
	return nil
}

func (c connExCreatedSequencesAccessor) isCreatedSequence(id descpb.ID) bool {
	__antithesis_instrumentation__.Notify(465236)
	_, ok := c.ex.extraTxnState.createdSequences[id]
	return ok
}

type emptyCreatedSequences struct{}

func (createdSequences emptyCreatedSequences) addCreatedSequence(id descpb.ID) error {
	__antithesis_instrumentation__.Notify(465237)
	return errors.AssertionFailedf("addCreatedSequence not supported in emptyCreatedSequences")
}

func (createdSequences emptyCreatedSequences) isCreatedSequence(id descpb.ID) bool {
	__antithesis_instrumentation__.Notify(465238)
	return false
}
