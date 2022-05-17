package scop

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"

type backfillOp struct{ baseOp }

func (backfillOp) Type() Type { __antithesis_instrumentation__.Notify(582381); return BackfillType }

type BackfillIndex struct {
	backfillOp
	TableID       descpb.ID
	SourceIndexID descpb.IndexID
	IndexID       descpb.IndexID
}

var _ = backfillOp{baseOp: baseOp{}}
