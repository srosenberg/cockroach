package scop

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"

type validationOp struct{ baseOp }

func (validationOp) Type() Type { __antithesis_instrumentation__.Notify(582485); return ValidationType }

type ValidateUniqueIndex struct {
	validationOp
	TableID descpb.ID
	IndexID descpb.IndexID
}

type ValidateCheckConstraint struct {
	validationOp
	TableID descpb.ID
	Name    string
}

var _ = validationOp{baseOp: baseOp{}}
