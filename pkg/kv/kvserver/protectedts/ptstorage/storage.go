// Package ptstorage implements protectedts.Storage.
package ptstorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type storage struct {
	settings *cluster.Settings
	ex       sqlutil.InternalExecutor

	knobs *protectedts.TestingKnobs
}

var _ protectedts.Storage = (*storage)(nil)

func useDeprecatedProtectedTSStorage(
	ctx context.Context, st *cluster.Settings, knobs *protectedts.TestingKnobs,
) bool {
	__antithesis_instrumentation__.Notify(111791)
	return !st.Version.IsActive(ctx, clusterversion.AlterSystemProtectedTimestampAddColumn) || func() bool {
		__antithesis_instrumentation__.Notify(111792)
		return knobs.DisableProtectedTimestampForMultiTenant == true
	}() == true
}

func New(
	settings *cluster.Settings, ex sqlutil.InternalExecutor, knobs *protectedts.TestingKnobs,
) protectedts.Storage {
	__antithesis_instrumentation__.Notify(111793)
	if knobs == nil {
		__antithesis_instrumentation__.Notify(111795)
		knobs = &protectedts.TestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(111796)
	}
	__antithesis_instrumentation__.Notify(111794)
	return &storage{settings: settings, ex: ex, knobs: knobs}
}

var errNoTxn = errors.New("must provide a non-nil transaction")

func (p *storage) UpdateTimestamp(
	ctx context.Context, txn *kv.Txn, id uuid.UUID, timestamp hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(111797)
	row, err := p.ex.QueryRowEx(ctx, "protectedts-update", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		updateTimestampQuery, id.GetBytesMut(), timestamp.AsOfSystemTime())
	if err != nil {
		__antithesis_instrumentation__.Notify(111800)
		return errors.Wrapf(err, "failed to update record %v", id)
	} else {
		__antithesis_instrumentation__.Notify(111801)
	}
	__antithesis_instrumentation__.Notify(111798)
	if len(row) == 0 {
		__antithesis_instrumentation__.Notify(111802)
		return protectedts.ErrNotExists
	} else {
		__antithesis_instrumentation__.Notify(111803)
	}
	__antithesis_instrumentation__.Notify(111799)
	return nil
}

func (p *storage) deprecatedProtect(
	ctx context.Context, txn *kv.Txn, r *ptpb.Record, meta []byte,
) error {
	__antithesis_instrumentation__.Notify(111804)
	s := makeSettings(p.settings)
	encodedSpans, err := protoutil.Marshal(&Spans{Spans: r.DeprecatedSpans})
	if err != nil {
		__antithesis_instrumentation__.Notify(111811)
		return errors.Wrap(err, "failed to marshal spans")
	} else {
		__antithesis_instrumentation__.Notify(111812)
	}
	__antithesis_instrumentation__.Notify(111805)
	it, err := p.ex.QueryIteratorEx(ctx, "protectedts-deprecated-protect", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		protectQueryWithoutTarget,
		s.maxSpans, s.maxBytes, len(r.DeprecatedSpans),
		r.ID, r.Timestamp.AsOfSystemTime(),
		r.MetaType, meta,
		len(r.DeprecatedSpans), encodedSpans)
	if err != nil {
		__antithesis_instrumentation__.Notify(111813)
		return errors.Wrapf(err, "failed to write record %v", r.ID)
	} else {
		__antithesis_instrumentation__.Notify(111814)
	}
	__antithesis_instrumentation__.Notify(111806)
	ok, err := it.Next(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(111815)
		return errors.Wrapf(err, "failed to write record %v", r.ID)
	} else {
		__antithesis_instrumentation__.Notify(111816)
	}
	__antithesis_instrumentation__.Notify(111807)
	if !ok {
		__antithesis_instrumentation__.Notify(111817)
		return errors.Newf("failed to write record %v", r.ID)
	} else {
		__antithesis_instrumentation__.Notify(111818)
	}
	__antithesis_instrumentation__.Notify(111808)
	row := it.Cur()
	if err := it.Close(); err != nil {
		__antithesis_instrumentation__.Notify(111819)
		log.Infof(ctx, "encountered %v when writing record %v", err, r.ID)
	} else {
		__antithesis_instrumentation__.Notify(111820)
	}
	__antithesis_instrumentation__.Notify(111809)
	if failed := *row[0].(*tree.DBool); failed {
		__antithesis_instrumentation__.Notify(111821)
		curNumSpans := int64(*row[1].(*tree.DInt))
		if s.maxSpans > 0 && func() bool {
			__antithesis_instrumentation__.Notify(111824)
			return curNumSpans+int64(len(r.DeprecatedSpans)) > s.maxSpans == true
		}() == true {
			__antithesis_instrumentation__.Notify(111825)
			return errors.WithHint(
				errors.Errorf("protectedts: limit exceeded: %d+%d > %d spans", curNumSpans,
					len(r.DeprecatedSpans), s.maxSpans),
				"SET CLUSTER SETTING kv.protectedts.max_spans to a higher value")
		} else {
			__antithesis_instrumentation__.Notify(111826)
		}
		__antithesis_instrumentation__.Notify(111822)
		curBytes := int64(*row[2].(*tree.DInt))
		recordBytes := int64(len(encodedSpans) + len(r.Meta) + len(r.MetaType))
		if s.maxBytes > 0 && func() bool {
			__antithesis_instrumentation__.Notify(111827)
			return curBytes+recordBytes > s.maxBytes == true
		}() == true {
			__antithesis_instrumentation__.Notify(111828)
			return errors.WithHint(
				errors.Errorf("protectedts: limit exceeded: %d+%d > %d bytes", curBytes, recordBytes,
					s.maxBytes),
				"SET CLUSTER SETTING kv.protectedts.max_bytes to a higher value")
		} else {
			__antithesis_instrumentation__.Notify(111829)
		}
		__antithesis_instrumentation__.Notify(111823)
		return protectedts.ErrExists
	} else {
		__antithesis_instrumentation__.Notify(111830)
	}
	__antithesis_instrumentation__.Notify(111810)
	return nil
}

func (p *storage) Protect(ctx context.Context, txn *kv.Txn, r *ptpb.Record) error {
	__antithesis_instrumentation__.Notify(111831)
	if err := validateRecordForProtect(ctx, r, p.settings, p.knobs); err != nil {
		__antithesis_instrumentation__.Notify(111842)
		return err
	} else {
		__antithesis_instrumentation__.Notify(111843)
	}
	__antithesis_instrumentation__.Notify(111832)
	if txn == nil {
		__antithesis_instrumentation__.Notify(111844)
		return errNoTxn
	} else {
		__antithesis_instrumentation__.Notify(111845)
	}
	__antithesis_instrumentation__.Notify(111833)

	meta := r.Meta
	if meta == nil {
		__antithesis_instrumentation__.Notify(111846)

		meta = []byte{}
	} else {
		__antithesis_instrumentation__.Notify(111847)
	}
	__antithesis_instrumentation__.Notify(111834)

	if useDeprecatedProtectedTSStorage(ctx, p.settings, p.knobs) {
		__antithesis_instrumentation__.Notify(111848)
		return p.deprecatedProtect(ctx, txn, r, meta)
	} else {
		__antithesis_instrumentation__.Notify(111849)
	}
	__antithesis_instrumentation__.Notify(111835)

	r.DeprecatedSpans = nil
	s := makeSettings(p.settings)
	encodedTarget, err := protoutil.Marshal(&ptpb.Target{Union: r.Target.GetUnion(),
		IgnoreIfExcludedFromBackup: r.Target.IgnoreIfExcludedFromBackup})
	if err != nil {
		__antithesis_instrumentation__.Notify(111850)
		return errors.Wrap(err, "failed to marshal spans")
	} else {
		__antithesis_instrumentation__.Notify(111851)
	}
	__antithesis_instrumentation__.Notify(111836)
	it, err := p.ex.QueryIteratorEx(ctx, "protectedts-protect", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		protectQuery,
		s.maxSpans, s.maxBytes, len(r.DeprecatedSpans),
		r.ID, r.Timestamp.AsOfSystemTime(),
		r.MetaType, meta,
		len(r.DeprecatedSpans), encodedTarget, encodedTarget)
	if err != nil {
		__antithesis_instrumentation__.Notify(111852)
		return errors.Wrapf(err, "failed to write record %v", r.ID)
	} else {
		__antithesis_instrumentation__.Notify(111853)
	}
	__antithesis_instrumentation__.Notify(111837)
	ok, err := it.Next(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(111854)
		return errors.Wrapf(err, "failed to write record %v", r.ID)
	} else {
		__antithesis_instrumentation__.Notify(111855)
	}
	__antithesis_instrumentation__.Notify(111838)
	if !ok {
		__antithesis_instrumentation__.Notify(111856)
		return errors.Newf("failed to write record %v", r.ID)
	} else {
		__antithesis_instrumentation__.Notify(111857)
	}
	__antithesis_instrumentation__.Notify(111839)
	row := it.Cur()
	if err := it.Close(); err != nil {
		__antithesis_instrumentation__.Notify(111858)
		log.Infof(ctx, "encountered %v when writing record %v", err, r.ID)
	} else {
		__antithesis_instrumentation__.Notify(111859)
	}
	__antithesis_instrumentation__.Notify(111840)
	if failed := *row[0].(*tree.DBool); failed {
		__antithesis_instrumentation__.Notify(111860)
		curBytes := int64(*row[1].(*tree.DInt))
		recordBytes := int64(len(encodedTarget) + len(r.Meta) + len(r.MetaType))
		if s.maxBytes > 0 && func() bool {
			__antithesis_instrumentation__.Notify(111862)
			return curBytes+recordBytes > s.maxBytes == true
		}() == true {
			__antithesis_instrumentation__.Notify(111863)
			return errors.WithHint(
				errors.Errorf("protectedts: limit exceeded: %d+%d > %d bytes", curBytes, recordBytes,
					s.maxBytes),
				"SET CLUSTER SETTING kv.protectedts.max_bytes to a higher value")
		} else {
			__antithesis_instrumentation__.Notify(111864)
		}
		__antithesis_instrumentation__.Notify(111861)
		return protectedts.ErrExists
	} else {
		__antithesis_instrumentation__.Notify(111865)
	}
	__antithesis_instrumentation__.Notify(111841)

	return nil
}

func (p *storage) deprecatedGetRecord(
	ctx context.Context, txn *kv.Txn, id uuid.UUID,
) (*ptpb.Record, error) {
	__antithesis_instrumentation__.Notify(111866)
	row, err := p.ex.QueryRowEx(ctx, "protectedts-deprecated-GetRecord", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		getRecordWithoutTargetQuery, id.GetBytesMut())
	if err != nil {
		__antithesis_instrumentation__.Notify(111870)
		return nil, errors.Wrapf(err, "failed to read record %v", id)
	} else {
		__antithesis_instrumentation__.Notify(111871)
	}
	__antithesis_instrumentation__.Notify(111867)
	if len(row) == 0 {
		__antithesis_instrumentation__.Notify(111872)
		return nil, protectedts.ErrNotExists
	} else {
		__antithesis_instrumentation__.Notify(111873)
	}
	__antithesis_instrumentation__.Notify(111868)
	var r ptpb.Record
	if err := rowToRecord(ctx, row, &r, p.settings, p.knobs); err != nil {
		__antithesis_instrumentation__.Notify(111874)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(111875)
	}
	__antithesis_instrumentation__.Notify(111869)
	return &r, nil
}

func (p *storage) GetRecord(ctx context.Context, txn *kv.Txn, id uuid.UUID) (*ptpb.Record, error) {
	__antithesis_instrumentation__.Notify(111876)
	if txn == nil {
		__antithesis_instrumentation__.Notify(111882)
		return nil, errNoTxn
	} else {
		__antithesis_instrumentation__.Notify(111883)
	}
	__antithesis_instrumentation__.Notify(111877)

	if useDeprecatedProtectedTSStorage(ctx, p.settings, p.knobs) {
		__antithesis_instrumentation__.Notify(111884)
		return p.deprecatedGetRecord(ctx, txn, id)
	} else {
		__antithesis_instrumentation__.Notify(111885)
	}
	__antithesis_instrumentation__.Notify(111878)

	row, err := p.ex.QueryRowEx(ctx, "protectedts-GetRecord", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		getRecordQuery, id.GetBytesMut())
	if err != nil {
		__antithesis_instrumentation__.Notify(111886)
		return nil, errors.Wrapf(err, "failed to read record %v", id)
	} else {
		__antithesis_instrumentation__.Notify(111887)
	}
	__antithesis_instrumentation__.Notify(111879)
	if len(row) == 0 {
		__antithesis_instrumentation__.Notify(111888)
		return nil, protectedts.ErrNotExists
	} else {
		__antithesis_instrumentation__.Notify(111889)
	}
	__antithesis_instrumentation__.Notify(111880)
	var r ptpb.Record
	if err := rowToRecord(ctx, row, &r, p.settings, p.knobs); err != nil {
		__antithesis_instrumentation__.Notify(111890)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(111891)
	}
	__antithesis_instrumentation__.Notify(111881)
	return &r, nil
}

func (p *storage) MarkVerified(ctx context.Context, txn *kv.Txn, id uuid.UUID) error {
	__antithesis_instrumentation__.Notify(111892)
	if txn == nil {
		__antithesis_instrumentation__.Notify(111896)
		return errNoTxn
	} else {
		__antithesis_instrumentation__.Notify(111897)
	}
	__antithesis_instrumentation__.Notify(111893)
	numRows, err := p.ex.ExecEx(ctx, "protectedts-MarkVerified", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		markVerifiedQuery, id.GetBytesMut())
	if err != nil {
		__antithesis_instrumentation__.Notify(111898)
		return errors.Wrapf(err, "failed to mark record %v as verified", id)
	} else {
		__antithesis_instrumentation__.Notify(111899)
	}
	__antithesis_instrumentation__.Notify(111894)
	if numRows == 0 {
		__antithesis_instrumentation__.Notify(111900)
		return protectedts.ErrNotExists
	} else {
		__antithesis_instrumentation__.Notify(111901)
	}
	__antithesis_instrumentation__.Notify(111895)
	return nil
}

func (p *storage) Release(ctx context.Context, txn *kv.Txn, id uuid.UUID) error {
	__antithesis_instrumentation__.Notify(111902)
	if txn == nil {
		__antithesis_instrumentation__.Notify(111906)
		return errNoTxn
	} else {
		__antithesis_instrumentation__.Notify(111907)
	}
	__antithesis_instrumentation__.Notify(111903)
	numRows, err := p.ex.ExecEx(ctx, "protectedts-Release", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		releaseQuery, id.GetBytesMut())
	if err != nil {
		__antithesis_instrumentation__.Notify(111908)
		return errors.Wrapf(err, "failed to release record %v", id)
	} else {
		__antithesis_instrumentation__.Notify(111909)
	}
	__antithesis_instrumentation__.Notify(111904)
	if numRows == 0 {
		__antithesis_instrumentation__.Notify(111910)
		return protectedts.ErrNotExists
	} else {
		__antithesis_instrumentation__.Notify(111911)
	}
	__antithesis_instrumentation__.Notify(111905)
	return nil
}

func (p *storage) GetMetadata(ctx context.Context, txn *kv.Txn) (ptpb.Metadata, error) {
	__antithesis_instrumentation__.Notify(111912)
	if txn == nil {
		__antithesis_instrumentation__.Notify(111916)
		return ptpb.Metadata{}, errNoTxn
	} else {
		__antithesis_instrumentation__.Notify(111917)
	}
	__antithesis_instrumentation__.Notify(111913)
	row, err := p.ex.QueryRowEx(ctx, "protectedts-GetMetadata", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		getMetadataQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(111918)
		return ptpb.Metadata{}, errors.Wrap(err, "failed to read metadata")
	} else {
		__antithesis_instrumentation__.Notify(111919)
	}
	__antithesis_instrumentation__.Notify(111914)
	if row == nil {
		__antithesis_instrumentation__.Notify(111920)
		return ptpb.Metadata{}, errors.New("failed to read metadata")
	} else {
		__antithesis_instrumentation__.Notify(111921)
	}
	__antithesis_instrumentation__.Notify(111915)
	return ptpb.Metadata{
		Version:    uint64(*row[0].(*tree.DInt)),
		NumRecords: uint64(*row[1].(*tree.DInt)),
		NumSpans:   uint64(*row[2].(*tree.DInt)),
		TotalBytes: uint64(*row[3].(*tree.DInt)),
	}, nil
}

func (p *storage) GetState(ctx context.Context, txn *kv.Txn) (ptpb.State, error) {
	__antithesis_instrumentation__.Notify(111922)
	if txn == nil {
		__antithesis_instrumentation__.Notify(111926)
		return ptpb.State{}, errNoTxn
	} else {
		__antithesis_instrumentation__.Notify(111927)
	}
	__antithesis_instrumentation__.Notify(111923)
	md, err := p.GetMetadata(ctx, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(111928)
		return ptpb.State{}, err
	} else {
		__antithesis_instrumentation__.Notify(111929)
	}
	__antithesis_instrumentation__.Notify(111924)
	records, err := p.getRecords(ctx, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(111930)
		return ptpb.State{}, err
	} else {
		__antithesis_instrumentation__.Notify(111931)
	}
	__antithesis_instrumentation__.Notify(111925)
	return ptpb.State{
		Metadata: md,
		Records:  records,
	}, nil
}

func (p *storage) deprecatedGetRecords(ctx context.Context, txn *kv.Txn) ([]ptpb.Record, error) {
	__antithesis_instrumentation__.Notify(111932)
	it, err := p.ex.QueryIteratorEx(ctx, "protectedts-deprecated-GetRecords", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		getRecordsWithoutTargetQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(111936)
		return nil, errors.Wrap(err, "failed to read records")
	} else {
		__antithesis_instrumentation__.Notify(111937)
	}
	__antithesis_instrumentation__.Notify(111933)

	var ok bool
	var records []ptpb.Record
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(111938)
		var record ptpb.Record
		if err := rowToRecord(ctx, it.Cur(), &record, p.settings, p.knobs); err != nil {
			__antithesis_instrumentation__.Notify(111940)
			log.Errorf(ctx, "failed to parse row as record: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(111941)
		}
		__antithesis_instrumentation__.Notify(111939)
		records = append(records, record)
	}
	__antithesis_instrumentation__.Notify(111934)
	if err != nil {
		__antithesis_instrumentation__.Notify(111942)
		return nil, errors.Wrap(err, "failed to read records")
	} else {
		__antithesis_instrumentation__.Notify(111943)
	}
	__antithesis_instrumentation__.Notify(111935)
	return records, nil
}

func (p *storage) getRecords(ctx context.Context, txn *kv.Txn) ([]ptpb.Record, error) {
	__antithesis_instrumentation__.Notify(111944)
	if useDeprecatedProtectedTSStorage(ctx, p.settings, p.knobs) {
		__antithesis_instrumentation__.Notify(111949)
		return p.deprecatedGetRecords(ctx, txn)
	} else {
		__antithesis_instrumentation__.Notify(111950)
	}
	__antithesis_instrumentation__.Notify(111945)

	it, err := p.ex.QueryIteratorEx(ctx, "protectedts-GetRecords", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()}, getRecordsQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(111951)
		return nil, errors.Wrap(err, "failed to read records")
	} else {
		__antithesis_instrumentation__.Notify(111952)
	}
	__antithesis_instrumentation__.Notify(111946)

	var ok bool
	var records []ptpb.Record
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(111953)
		var record ptpb.Record
		if err := rowToRecord(ctx, it.Cur(), &record, p.settings, p.knobs); err != nil {
			__antithesis_instrumentation__.Notify(111955)
			log.Errorf(ctx, "failed to parse row as record: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(111956)
		}
		__antithesis_instrumentation__.Notify(111954)
		records = append(records, record)
	}
	__antithesis_instrumentation__.Notify(111947)
	if err != nil {
		__antithesis_instrumentation__.Notify(111957)
		return nil, errors.Wrap(err, "failed to read records")
	} else {
		__antithesis_instrumentation__.Notify(111958)
	}
	__antithesis_instrumentation__.Notify(111948)
	return records, nil
}

func rowToRecord(
	ctx context.Context,
	row tree.Datums,
	r *ptpb.Record,
	st *cluster.Settings,
	knobs *protectedts.TestingKnobs,
) error {
	__antithesis_instrumentation__.Notify(111959)
	r.ID = row[0].(*tree.DUuid).UUID.GetBytes()
	tsDecimal := row[1].(*tree.DDecimal)
	ts, err := tree.DecimalToHLC(&tsDecimal.Decimal)
	if err != nil {
		__antithesis_instrumentation__.Notify(111964)
		return errors.Wrapf(err, "failed to parse timestamp for %v", r.ID)
	} else {
		__antithesis_instrumentation__.Notify(111965)
	}
	__antithesis_instrumentation__.Notify(111960)
	r.Timestamp = ts

	r.MetaType = string(*row[2].(*tree.DString))
	if row[3] != tree.DNull {
		__antithesis_instrumentation__.Notify(111966)
		if meta := row[3].(*tree.DBytes); len(*meta) > 0 {
			__antithesis_instrumentation__.Notify(111967)
			r.Meta = []byte(*meta)
		} else {
			__antithesis_instrumentation__.Notify(111968)
		}
	} else {
		__antithesis_instrumentation__.Notify(111969)
	}
	__antithesis_instrumentation__.Notify(111961)
	var spans Spans
	if err := protoutil.Unmarshal([]byte(*row[4].(*tree.DBytes)), &spans); err != nil {
		__antithesis_instrumentation__.Notify(111970)
		return errors.Wrapf(err, "failed to unmarshal span for %v", r.ID)
	} else {
		__antithesis_instrumentation__.Notify(111971)
	}
	__antithesis_instrumentation__.Notify(111962)
	r.DeprecatedSpans = spans.Spans
	r.Verified = bool(*row[5].(*tree.DBool))

	if !useDeprecatedProtectedTSStorage(ctx, st, knobs) {
		__antithesis_instrumentation__.Notify(111972)
		target := &ptpb.Target{}
		targetDBytes, ok := row[6].(*tree.DBytes)
		if !ok {
			__antithesis_instrumentation__.Notify(111975)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(111976)
		}
		__antithesis_instrumentation__.Notify(111973)
		if err := protoutil.Unmarshal([]byte(*targetDBytes), target); err != nil {
			__antithesis_instrumentation__.Notify(111977)
			return errors.Wrapf(err, "failed to unmarshal target for %v", r.ID)
		} else {
			__antithesis_instrumentation__.Notify(111978)
		}
		__antithesis_instrumentation__.Notify(111974)
		r.Target = target
	} else {
		__antithesis_instrumentation__.Notify(111979)
	}
	__antithesis_instrumentation__.Notify(111963)
	return nil
}

type settings struct {
	maxSpans int64
	maxBytes int64
}

func makeSettings(s *cluster.Settings) settings {
	__antithesis_instrumentation__.Notify(111980)
	return settings{
		maxSpans: protectedts.MaxSpans.Get(&s.SV),
		maxBytes: protectedts.MaxBytes.Get(&s.SV),
	}
}

var (
	errZeroTimestamp        = errors.New("invalid zero value timestamp")
	errZeroID               = errors.New("invalid zero value ID")
	errEmptySpans           = errors.Errorf("invalid empty set of spans")
	errNilTarget            = errors.Errorf("invalid nil target")
	errInvalidMeta          = errors.Errorf("invalid Meta with empty MetaType")
	errCreateVerifiedRecord = errors.Errorf("cannot create a verified record")
)

func validateRecordForProtect(
	ctx context.Context, r *ptpb.Record, st *cluster.Settings, knobs *protectedts.TestingKnobs,
) error {
	__antithesis_instrumentation__.Notify(111981)
	if r.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(111988)
		return errZeroTimestamp
	} else {
		__antithesis_instrumentation__.Notify(111989)
	}
	__antithesis_instrumentation__.Notify(111982)
	if r.ID.GetUUID() == uuid.Nil {
		__antithesis_instrumentation__.Notify(111990)
		return errZeroID
	} else {
		__antithesis_instrumentation__.Notify(111991)
	}
	__antithesis_instrumentation__.Notify(111983)
	useDeprecatedPTSStorage := useDeprecatedProtectedTSStorage(ctx, st, knobs)
	if !useDeprecatedPTSStorage && func() bool {
		__antithesis_instrumentation__.Notify(111992)
		return r.Target == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(111993)
		return errNilTarget
	} else {
		__antithesis_instrumentation__.Notify(111994)
	}
	__antithesis_instrumentation__.Notify(111984)
	if useDeprecatedPTSStorage && func() bool {
		__antithesis_instrumentation__.Notify(111995)
		return len(r.DeprecatedSpans) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(111996)
		return errEmptySpans
	} else {
		__antithesis_instrumentation__.Notify(111997)
	}
	__antithesis_instrumentation__.Notify(111985)
	if len(r.Meta) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(111998)
		return len(r.MetaType) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(111999)
		return errInvalidMeta
	} else {
		__antithesis_instrumentation__.Notify(112000)
	}
	__antithesis_instrumentation__.Notify(111986)
	if r.Verified {
		__antithesis_instrumentation__.Notify(112001)
		return errCreateVerifiedRecord
	} else {
		__antithesis_instrumentation__.Notify(112002)
	}
	__antithesis_instrumentation__.Notify(111987)
	return nil
}
