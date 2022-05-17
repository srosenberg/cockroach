package ptstorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

func WithDatabase(s protectedts.Storage, db *kv.DB) protectedts.Storage {
	__antithesis_instrumentation__.Notify(112153)
	return &storageWithDatabase{s: s, db: db}
}

type storageWithDatabase struct {
	db *kv.DB
	s  protectedts.Storage
}

func (s *storageWithDatabase) Protect(ctx context.Context, txn *kv.Txn, r *ptpb.Record) error {
	__antithesis_instrumentation__.Notify(112154)
	if txn == nil {
		__antithesis_instrumentation__.Notify(112156)
		return s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(112157)
			return s.s.Protect(ctx, txn, r)
		})
	} else {
		__antithesis_instrumentation__.Notify(112158)
	}
	__antithesis_instrumentation__.Notify(112155)
	return s.s.Protect(ctx, txn, r)
}

func (s *storageWithDatabase) GetRecord(
	ctx context.Context, txn *kv.Txn, id uuid.UUID,
) (r *ptpb.Record, err error) {
	__antithesis_instrumentation__.Notify(112159)
	if txn == nil {
		__antithesis_instrumentation__.Notify(112161)
		err = s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(112163)
			r, err = s.s.GetRecord(ctx, txn, id)
			return err
		})
		__antithesis_instrumentation__.Notify(112162)
		return r, err
	} else {
		__antithesis_instrumentation__.Notify(112164)
	}
	__antithesis_instrumentation__.Notify(112160)
	return s.s.GetRecord(ctx, txn, id)
}

func (s *storageWithDatabase) MarkVerified(ctx context.Context, txn *kv.Txn, id uuid.UUID) error {
	__antithesis_instrumentation__.Notify(112165)
	if txn == nil {
		__antithesis_instrumentation__.Notify(112167)
		return s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(112168)
			return s.s.Release(ctx, txn, id)
		})
	} else {
		__antithesis_instrumentation__.Notify(112169)
	}
	__antithesis_instrumentation__.Notify(112166)
	return s.s.Release(ctx, txn, id)
}

func (s *storageWithDatabase) Release(ctx context.Context, txn *kv.Txn, id uuid.UUID) error {
	__antithesis_instrumentation__.Notify(112170)
	if txn == nil {
		__antithesis_instrumentation__.Notify(112172)
		return s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(112173)
			return s.s.Release(ctx, txn, id)
		})
	} else {
		__antithesis_instrumentation__.Notify(112174)
	}
	__antithesis_instrumentation__.Notify(112171)
	return s.s.Release(ctx, txn, id)
}

func (s *storageWithDatabase) GetMetadata(
	ctx context.Context, txn *kv.Txn,
) (md ptpb.Metadata, err error) {
	__antithesis_instrumentation__.Notify(112175)
	if txn == nil {
		__antithesis_instrumentation__.Notify(112177)
		err = s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(112179)
			md, err = s.s.GetMetadata(ctx, txn)
			return err
		})
		__antithesis_instrumentation__.Notify(112178)
		return md, err
	} else {
		__antithesis_instrumentation__.Notify(112180)
	}
	__antithesis_instrumentation__.Notify(112176)
	return s.s.GetMetadata(ctx, txn)
}

func (s *storageWithDatabase) GetState(
	ctx context.Context, txn *kv.Txn,
) (state ptpb.State, err error) {
	__antithesis_instrumentation__.Notify(112181)
	if txn == nil {
		__antithesis_instrumentation__.Notify(112183)
		err = s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(112185)
			state, err = s.s.GetState(ctx, txn)
			return err
		})
		__antithesis_instrumentation__.Notify(112184)
		return state, err
	} else {
		__antithesis_instrumentation__.Notify(112186)
	}
	__antithesis_instrumentation__.Notify(112182)
	return s.s.GetState(ctx, txn)
}

func (s *storageWithDatabase) UpdateTimestamp(
	ctx context.Context, txn *kv.Txn, id uuid.UUID, timestamp hlc.Timestamp,
) (err error) {
	__antithesis_instrumentation__.Notify(112187)
	if txn == nil {
		__antithesis_instrumentation__.Notify(112189)
		err = s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(112191)
			return s.s.UpdateTimestamp(ctx, txn, id, timestamp)
		})
		__antithesis_instrumentation__.Notify(112190)
		return err
	} else {
		__antithesis_instrumentation__.Notify(112192)
	}
	__antithesis_instrumentation__.Notify(112188)
	return s.s.UpdateTimestamp(ctx, txn, id, timestamp)
}
