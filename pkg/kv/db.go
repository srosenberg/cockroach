package kv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

type KeyValue struct {
	Key   roachpb.Key
	Value *roachpb.Value
}

func (kv *KeyValue) String() string {
	__antithesis_instrumentation__.Notify(86650)
	return kv.Key.String() + "=" + kv.PrettyValue()
}

func (kv *KeyValue) Exists() bool {
	__antithesis_instrumentation__.Notify(86651)
	return kv.Value != nil
}

func (kv *KeyValue) PrettyValue() string {
	__antithesis_instrumentation__.Notify(86652)
	if kv.Value == nil {
		__antithesis_instrumentation__.Notify(86655)
		return "nil"
	} else {
		__antithesis_instrumentation__.Notify(86656)
	}
	__antithesis_instrumentation__.Notify(86653)
	switch kv.Value.GetTag() {
	case roachpb.ValueType_INT:
		__antithesis_instrumentation__.Notify(86657)
		v, err := kv.Value.GetInt()
		if err != nil {
			__antithesis_instrumentation__.Notify(86666)
			return fmt.Sprintf("%v", err)
		} else {
			__antithesis_instrumentation__.Notify(86667)
		}
		__antithesis_instrumentation__.Notify(86658)
		return fmt.Sprintf("%d", v)
	case roachpb.ValueType_FLOAT:
		__antithesis_instrumentation__.Notify(86659)
		v, err := kv.Value.GetFloat()
		if err != nil {
			__antithesis_instrumentation__.Notify(86668)
			return fmt.Sprintf("%v", err)
		} else {
			__antithesis_instrumentation__.Notify(86669)
		}
		__antithesis_instrumentation__.Notify(86660)
		return fmt.Sprintf("%v", v)
	case roachpb.ValueType_BYTES:
		__antithesis_instrumentation__.Notify(86661)
		v, err := kv.Value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(86670)
			return fmt.Sprintf("%v", err)
		} else {
			__antithesis_instrumentation__.Notify(86671)
		}
		__antithesis_instrumentation__.Notify(86662)
		return fmt.Sprintf("%q", v)
	case roachpb.ValueType_TIME:
		__antithesis_instrumentation__.Notify(86663)
		v, err := kv.Value.GetTime()
		if err != nil {
			__antithesis_instrumentation__.Notify(86672)
			return fmt.Sprintf("%v", err)
		} else {
			__antithesis_instrumentation__.Notify(86673)
		}
		__antithesis_instrumentation__.Notify(86664)
		return v.String()
	default:
		__antithesis_instrumentation__.Notify(86665)
	}
	__antithesis_instrumentation__.Notify(86654)
	return fmt.Sprintf("%x", kv.Value.RawBytes)
}

func (kv *KeyValue) ValueBytes() []byte {
	__antithesis_instrumentation__.Notify(86674)
	if kv.Value == nil {
		__antithesis_instrumentation__.Notify(86677)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(86678)
	}
	__antithesis_instrumentation__.Notify(86675)
	bytes, err := kv.Value.GetBytes()
	if err != nil {
		__antithesis_instrumentation__.Notify(86679)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(86680)
	}
	__antithesis_instrumentation__.Notify(86676)
	return bytes
}

func (kv *KeyValue) ValueInt() int64 {
	__antithesis_instrumentation__.Notify(86681)
	if kv.Value == nil {
		__antithesis_instrumentation__.Notify(86684)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(86685)
	}
	__antithesis_instrumentation__.Notify(86682)
	i, err := kv.Value.GetInt()
	if err != nil {
		__antithesis_instrumentation__.Notify(86686)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(86687)
	}
	__antithesis_instrumentation__.Notify(86683)
	return i
}

func (kv *KeyValue) ValueProto(msg protoutil.Message) error {
	__antithesis_instrumentation__.Notify(86688)
	if kv.Value == nil {
		__antithesis_instrumentation__.Notify(86690)
		msg.Reset()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(86691)
	}
	__antithesis_instrumentation__.Notify(86689)
	return kv.Value.GetProto(msg)
}

type Result struct {
	calls int

	Err error

	Rows []KeyValue

	Keys []roachpb.Key

	ResumeSpan *roachpb.Span

	ResumeReason roachpb.ResumeReason

	ResumeNextBytes int64
}

func (r *Result) ResumeSpanAsValue() roachpb.Span {
	__antithesis_instrumentation__.Notify(86692)
	if r.ResumeSpan == nil {
		__antithesis_instrumentation__.Notify(86694)
		return roachpb.Span{}
	} else {
		__antithesis_instrumentation__.Notify(86695)
	}
	__antithesis_instrumentation__.Notify(86693)
	return *r.ResumeSpan
}

func (r Result) String() string {
	__antithesis_instrumentation__.Notify(86696)
	if r.Err != nil {
		__antithesis_instrumentation__.Notify(86699)
		return r.Err.Error()
	} else {
		__antithesis_instrumentation__.Notify(86700)
	}
	__antithesis_instrumentation__.Notify(86697)
	var buf strings.Builder
	for i := range r.Rows {
		__antithesis_instrumentation__.Notify(86701)
		if i > 0 {
			__antithesis_instrumentation__.Notify(86703)
			buf.WriteString("\n")
		} else {
			__antithesis_instrumentation__.Notify(86704)
		}
		__antithesis_instrumentation__.Notify(86702)
		fmt.Fprintf(&buf, "%d: %s", i, &r.Rows[i])
	}
	__antithesis_instrumentation__.Notify(86698)
	return buf.String()
}

type DBContext struct {
	UserPriority roachpb.UserPriority

	NodeID *base.SQLIDContainer

	Stopper *stop.Stopper
}

func DefaultDBContext(stopper *stop.Stopper) DBContext {
	__antithesis_instrumentation__.Notify(86705)
	var c base.NodeIDContainer
	return DBContext{
		UserPriority: roachpb.NormalUserPriority,

		NodeID:  base.NewSQLIDContainerForNode(&c),
		Stopper: stopper,
	}
}

type CrossRangeTxnWrapperSender struct {
	db      *DB
	wrapped Sender
}

var _ Sender = &CrossRangeTxnWrapperSender{}

func (s *CrossRangeTxnWrapperSender) Send(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(86706)
	if ba.Txn != nil {
		__antithesis_instrumentation__.Notify(86711)
		log.Fatalf(ctx, "CrossRangeTxnWrapperSender can't handle transactional requests")
	} else {
		__antithesis_instrumentation__.Notify(86712)
	}
	__antithesis_instrumentation__.Notify(86707)

	br, pErr := s.wrapped.Send(ctx, ba)
	if _, ok := pErr.GetDetail().(*roachpb.OpRequiresTxnError); !ok {
		__antithesis_instrumentation__.Notify(86713)
		return br, pErr
	} else {
		__antithesis_instrumentation__.Notify(86714)
	}
	__antithesis_instrumentation__.Notify(86708)

	err := s.db.Txn(ctx, func(ctx context.Context, txn *Txn) error {
		__antithesis_instrumentation__.Notify(86715)
		txn.SetDebugName("auto-wrap")
		b := txn.NewBatch()
		b.Header = ba.Header
		for _, arg := range ba.Requests {
			__antithesis_instrumentation__.Notify(86717)
			req := arg.GetInner().ShallowCopy()
			b.AddRawRequest(req)
		}
		__antithesis_instrumentation__.Notify(86716)
		err := txn.CommitInBatch(ctx, b)
		br = b.RawResponse()
		return err
	})
	__antithesis_instrumentation__.Notify(86709)
	if err != nil {
		__antithesis_instrumentation__.Notify(86718)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(86719)
	}
	__antithesis_instrumentation__.Notify(86710)
	br.Txn = nil
	return br, nil
}

func (s *CrossRangeTxnWrapperSender) Wrapped() Sender {
	__antithesis_instrumentation__.Notify(86720)
	return s.wrapped
}

type DB struct {
	log.AmbientContext

	factory TxnSenderFactory
	clock   *hlc.Clock
	ctx     DBContext

	crs CrossRangeTxnWrapperSender

	SQLKVResponseAdmissionQ *admission.WorkQueue
}

func (db *DB) NonTransactionalSender() Sender {
	__antithesis_instrumentation__.Notify(86721)
	return &db.crs
}

func (db *DB) GetFactory() TxnSenderFactory {
	__antithesis_instrumentation__.Notify(86722)
	return db.factory
}

func (db *DB) Clock() *hlc.Clock {
	__antithesis_instrumentation__.Notify(86723)
	return db.clock
}

func NewDB(
	actx log.AmbientContext, factory TxnSenderFactory, clock *hlc.Clock, stopper *stop.Stopper,
) *DB {
	__antithesis_instrumentation__.Notify(86724)
	return NewDBWithContext(actx, factory, clock, DefaultDBContext(stopper))
}

func NewDBWithContext(
	actx log.AmbientContext, factory TxnSenderFactory, clock *hlc.Clock, ctx DBContext,
) *DB {
	__antithesis_instrumentation__.Notify(86725)
	if actx.Tracer == nil {
		__antithesis_instrumentation__.Notify(86727)
		panic("no tracer set in AmbientCtx")
	} else {
		__antithesis_instrumentation__.Notify(86728)
	}
	__antithesis_instrumentation__.Notify(86726)
	db := &DB{
		AmbientContext: actx,
		factory:        factory,
		clock:          clock,
		ctx:            ctx,
		crs: CrossRangeTxnWrapperSender{
			wrapped: factory.NonTransactionalSender(),
		},
	}
	db.crs.db = db
	return db
}

func (db *DB) Get(ctx context.Context, key interface{}) (KeyValue, error) {
	__antithesis_instrumentation__.Notify(86729)
	b := &Batch{}
	b.Get(key)
	return getOneRow(db.Run(ctx, b), b)
}

func (db *DB) GetForUpdate(ctx context.Context, key interface{}) (KeyValue, error) {
	__antithesis_instrumentation__.Notify(86730)
	b := &Batch{}
	b.GetForUpdate(key)
	return getOneRow(db.Run(ctx, b), b)
}

func (db *DB) GetProto(ctx context.Context, key interface{}, msg protoutil.Message) error {
	__antithesis_instrumentation__.Notify(86731)
	_, err := db.GetProtoTs(ctx, key, msg)
	return err
}

func (db *DB) GetProtoTs(
	ctx context.Context, key interface{}, msg protoutil.Message,
) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(86732)
	r, err := db.Get(ctx, key)
	if err != nil {
		__antithesis_instrumentation__.Notify(86735)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(86736)
	}
	__antithesis_instrumentation__.Notify(86733)
	if err := r.ValueProto(msg); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(86737)
		return r.Value == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(86738)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(86739)
	}
	__antithesis_instrumentation__.Notify(86734)
	return r.Value.Timestamp, nil
}

func (db *DB) Put(ctx context.Context, key, value interface{}) error {
	__antithesis_instrumentation__.Notify(86740)
	b := &Batch{}
	b.Put(key, value)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) PutInline(ctx context.Context, key, value interface{}) error {
	__antithesis_instrumentation__.Notify(86741)
	b := &Batch{}
	b.PutInline(key, value)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) CPut(ctx context.Context, key, value interface{}, expValue []byte) error {
	__antithesis_instrumentation__.Notify(86742)
	b := &Batch{}
	b.CPut(key, value, expValue)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) CPutInline(ctx context.Context, key, value interface{}, expValue []byte) error {
	__antithesis_instrumentation__.Notify(86743)
	b := &Batch{}
	b.CPutInline(key, value, expValue)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) InitPut(ctx context.Context, key, value interface{}, failOnTombstones bool) error {
	__antithesis_instrumentation__.Notify(86744)
	b := &Batch{}
	b.InitPut(key, value, failOnTombstones)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) Inc(ctx context.Context, key interface{}, value int64) (KeyValue, error) {
	__antithesis_instrumentation__.Notify(86745)
	b := &Batch{}
	b.Inc(key, value)
	return getOneRow(db.Run(ctx, b), b)
}

func (db *DB) scan(
	ctx context.Context,
	begin, end interface{},
	maxRows int64,
	isReverse bool,
	forUpdate bool,
	readConsistency roachpb.ReadConsistencyType,
) ([]KeyValue, error) {
	__antithesis_instrumentation__.Notify(86746)
	b := &Batch{}
	b.Header.ReadConsistency = readConsistency
	if maxRows > 0 {
		__antithesis_instrumentation__.Notify(86748)
		b.Header.MaxSpanRequestKeys = maxRows
	} else {
		__antithesis_instrumentation__.Notify(86749)
	}
	__antithesis_instrumentation__.Notify(86747)
	b.scan(begin, end, isReverse, forUpdate)
	r, err := getOneResult(db.Run(ctx, b), b)
	return r.Rows, err
}

func (db *DB) Scan(ctx context.Context, begin, end interface{}, maxRows int64) ([]KeyValue, error) {
	__antithesis_instrumentation__.Notify(86750)
	return db.scan(ctx, begin, end, maxRows, false, false, roachpb.CONSISTENT)
}

func (db *DB) ScanForUpdate(
	ctx context.Context, begin, end interface{}, maxRows int64,
) ([]KeyValue, error) {
	__antithesis_instrumentation__.Notify(86751)
	return db.scan(ctx, begin, end, maxRows, false, true, roachpb.CONSISTENT)
}

func (db *DB) ReverseScan(
	ctx context.Context, begin, end interface{}, maxRows int64,
) ([]KeyValue, error) {
	__antithesis_instrumentation__.Notify(86752)
	return db.scan(ctx, begin, end, maxRows, true, false, roachpb.CONSISTENT)
}

func (db *DB) ReverseScanForUpdate(
	ctx context.Context, begin, end interface{}, maxRows int64,
) ([]KeyValue, error) {
	__antithesis_instrumentation__.Notify(86753)
	return db.scan(ctx, begin, end, maxRows, true, true, roachpb.CONSISTENT)
}

func (db *DB) Del(ctx context.Context, keys ...interface{}) error {
	__antithesis_instrumentation__.Notify(86754)
	b := &Batch{}
	b.Del(keys...)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) DelRange(
	ctx context.Context, begin, end interface{}, returnKeys bool,
) ([]roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(86755)
	b := &Batch{}
	b.DelRange(begin, end, returnKeys)
	r, err := getOneResult(db.Run(ctx, b), b)
	return r.Keys, err
}

func (db *DB) AdminMerge(ctx context.Context, key interface{}) error {
	__antithesis_instrumentation__.Notify(86756)
	b := &Batch{}
	b.adminMerge(key)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) AdminSplit(
	ctx context.Context,
	splitKey interface{},
	expirationTime hlc.Timestamp,
	predicateKeys ...roachpb.Key,
) error {
	__antithesis_instrumentation__.Notify(86757)
	b := &Batch{}
	b.adminSplit(splitKey, expirationTime, predicateKeys)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) AdminScatter(
	ctx context.Context, key roachpb.Key, maxSize int64,
) (*roachpb.AdminScatterResponse, error) {
	__antithesis_instrumentation__.Notify(86758)
	scatterReq := &roachpb.AdminScatterRequest{
		RequestHeader:   roachpb.RequestHeaderFromSpan(roachpb.Span{Key: key, EndKey: key.Next()}),
		RandomizeLeases: true,
		MaxSize:         maxSize,
	}
	raw, pErr := SendWrapped(ctx, db.NonTransactionalSender(), scatterReq)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(86761)
		return nil, pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(86762)
	}
	__antithesis_instrumentation__.Notify(86759)
	resp, ok := raw.(*roachpb.AdminScatterResponse)
	if !ok {
		__antithesis_instrumentation__.Notify(86763)
		return nil, errors.Errorf("unexpected response of type %T for AdminScatter", raw)
	} else {
		__antithesis_instrumentation__.Notify(86764)
	}
	__antithesis_instrumentation__.Notify(86760)
	return resp, nil
}

func (db *DB) AdminUnsplit(ctx context.Context, splitKey interface{}) error {
	__antithesis_instrumentation__.Notify(86765)
	b := &Batch{}
	b.adminUnsplit(splitKey)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) AdminTransferLease(
	ctx context.Context, key interface{}, target roachpb.StoreID,
) error {
	__antithesis_instrumentation__.Notify(86766)
	b := &Batch{}
	b.adminTransferLease(key, target)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) AdminChangeReplicas(
	ctx context.Context,
	key interface{},
	expDesc roachpb.RangeDescriptor,
	chgs []roachpb.ReplicationChange,
) (*roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(86767)
	b := &Batch{}
	b.adminChangeReplicas(key, expDesc, chgs)
	if err := getOneErr(db.Run(ctx, b), b); err != nil {
		__antithesis_instrumentation__.Notify(86771)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(86772)
	}
	__antithesis_instrumentation__.Notify(86768)
	responses := b.response.Responses
	if len(responses) == 0 {
		__antithesis_instrumentation__.Notify(86773)
		return nil, errors.Errorf("unexpected empty responses for AdminChangeReplicas")
	} else {
		__antithesis_instrumentation__.Notify(86774)
	}
	__antithesis_instrumentation__.Notify(86769)
	resp, ok := responses[0].GetInner().(*roachpb.AdminChangeReplicasResponse)
	if !ok {
		__antithesis_instrumentation__.Notify(86775)
		return nil, errors.Errorf("unexpected response of type %T for AdminChangeReplicas",
			responses[0].GetInner())
	} else {
		__antithesis_instrumentation__.Notify(86776)
	}
	__antithesis_instrumentation__.Notify(86770)
	desc := resp.Desc
	return &desc, nil
}

func (db *DB) AdminRelocateRange(
	ctx context.Context,
	key interface{},
	voterTargets, nonVoterTargets []roachpb.ReplicationTarget,
	transferLeaseToFirstVoter bool,
) error {
	__antithesis_instrumentation__.Notify(86777)
	b := &Batch{}
	b.adminRelocateRange(key, voterTargets, nonVoterTargets, transferLeaseToFirstVoter)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) AddSSTable(
	ctx context.Context,
	begin, end interface{},
	data []byte,
	disallowConflicts bool,
	disallowShadowing bool,
	disallowShadowingBelow hlc.Timestamp,
	stats *enginepb.MVCCStats,
	ingestAsWrites bool,
	batchTs hlc.Timestamp,
) (roachpb.Span, int64, error) {
	__antithesis_instrumentation__.Notify(86778)
	b := &Batch{Header: roachpb.Header{Timestamp: batchTs}}
	b.addSSTable(begin, end, data, disallowConflicts, disallowShadowing, disallowShadowingBelow,
		stats, ingestAsWrites, hlc.Timestamp{})
	err := getOneErr(db.Run(ctx, b), b)
	if err != nil {
		__antithesis_instrumentation__.Notify(86781)
		return roachpb.Span{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(86782)
	}
	__antithesis_instrumentation__.Notify(86779)
	if l := len(b.response.Responses); l != 1 {
		__antithesis_instrumentation__.Notify(86783)
		return roachpb.Span{}, 0, errors.AssertionFailedf("expected single response, got %d", l)
	} else {
		__antithesis_instrumentation__.Notify(86784)
	}
	__antithesis_instrumentation__.Notify(86780)
	resp := b.response.Responses[0].GetAddSstable()
	return resp.RangeSpan, resp.AvailableBytes, nil
}

func (db *DB) AddSSTableAtBatchTimestamp(
	ctx context.Context,
	begin, end interface{},
	data []byte,
	disallowConflicts bool,
	disallowShadowing bool,
	disallowShadowingBelow hlc.Timestamp,
	stats *enginepb.MVCCStats,
	ingestAsWrites bool,
	batchTs hlc.Timestamp,
) (hlc.Timestamp, roachpb.Span, int64, error) {
	__antithesis_instrumentation__.Notify(86785)
	b := &Batch{Header: roachpb.Header{Timestamp: batchTs}}
	b.addSSTable(begin, end, data, disallowConflicts, disallowShadowing, disallowShadowingBelow,
		stats, ingestAsWrites, batchTs)
	err := getOneErr(db.Run(ctx, b), b)
	if err != nil {
		__antithesis_instrumentation__.Notify(86788)
		return hlc.Timestamp{}, roachpb.Span{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(86789)
	}
	__antithesis_instrumentation__.Notify(86786)
	if l := len(b.response.Responses); l != 1 {
		__antithesis_instrumentation__.Notify(86790)
		return hlc.Timestamp{}, roachpb.Span{}, 0, errors.AssertionFailedf("expected single response, got %d", l)
	} else {
		__antithesis_instrumentation__.Notify(86791)
	}
	__antithesis_instrumentation__.Notify(86787)
	resp := b.response.Responses[0].GetAddSstable()
	return b.response.Timestamp, resp.RangeSpan, resp.AvailableBytes, nil
}

func (db *DB) Migrate(ctx context.Context, begin, end interface{}, version roachpb.Version) error {
	__antithesis_instrumentation__.Notify(86792)
	b := &Batch{}
	b.migrate(begin, end, version)
	return getOneErr(db.Run(ctx, b), b)
}

func (db *DB) QueryResolvedTimestamp(
	ctx context.Context, begin, end interface{}, nearest bool,
) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(86793)
	b := &Batch{}
	b.queryResolvedTimestamp(begin, end)
	if nearest {
		__antithesis_instrumentation__.Notify(86796)
		b.Header.RoutingPolicy = roachpb.RoutingPolicy_NEAREST
	} else {
		__antithesis_instrumentation__.Notify(86797)
	}
	__antithesis_instrumentation__.Notify(86794)
	if err := getOneErr(db.Run(ctx, b), b); err != nil {
		__antithesis_instrumentation__.Notify(86798)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(86799)
	}
	__antithesis_instrumentation__.Notify(86795)
	r := b.RawResponse().Responses[0].GetQueryResolvedTimestamp()
	return r.ResolvedTS, nil
}

func (db *DB) ScanInterleavedIntents(
	ctx context.Context, begin, end interface{}, ts hlc.Timestamp,
) ([]roachpb.Intent, *roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(86800)
	b := &Batch{Header: roachpb.Header{Timestamp: ts}}
	b.scanInterleavedIntents(begin, end)
	result, err := getOneResult(db.Run(ctx, b), b)
	if err != nil {
		__antithesis_instrumentation__.Notify(86804)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(86805)
	}
	__antithesis_instrumentation__.Notify(86801)
	responses := b.response.Responses
	if len(responses) == 0 {
		__antithesis_instrumentation__.Notify(86806)
		return nil, nil, errors.Errorf("unexpected empty response for ScanInterleavedIntents")
	} else {
		__antithesis_instrumentation__.Notify(86807)
	}
	__antithesis_instrumentation__.Notify(86802)
	resp, ok := responses[0].GetInner().(*roachpb.ScanInterleavedIntentsResponse)
	if !ok {
		__antithesis_instrumentation__.Notify(86808)
		return nil, nil, errors.Errorf("unexpected response of type %T for ScanInterleavedIntents",
			responses[0].GetInner())
	} else {
		__antithesis_instrumentation__.Notify(86809)
	}
	__antithesis_instrumentation__.Notify(86803)
	return resp.Intents, result.ResumeSpan, nil
}

func (db *DB) Barrier(ctx context.Context, begin, end interface{}) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(86810)
	b := &Batch{}
	b.barrier(begin, end)
	err := getOneErr(db.Run(ctx, b), b)
	if err != nil {
		__antithesis_instrumentation__.Notify(86814)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(86815)
	}
	__antithesis_instrumentation__.Notify(86811)
	responses := b.response.Responses
	if len(responses) == 0 {
		__antithesis_instrumentation__.Notify(86816)
		return hlc.Timestamp{}, errors.Errorf("unexpected empty response for Barrier")
	} else {
		__antithesis_instrumentation__.Notify(86817)
	}
	__antithesis_instrumentation__.Notify(86812)
	resp, ok := responses[0].GetInner().(*roachpb.BarrierResponse)
	if !ok {
		__antithesis_instrumentation__.Notify(86818)
		return hlc.Timestamp{}, errors.Errorf("unexpected response of type %T for Barrier",
			responses[0].GetInner())
	} else {
		__antithesis_instrumentation__.Notify(86819)
	}
	__antithesis_instrumentation__.Notify(86813)
	return resp.Timestamp, nil
}

func sendAndFill(ctx context.Context, send SenderFunc, b *Batch) error {
	__antithesis_instrumentation__.Notify(86820)

	var ba roachpb.BatchRequest
	ba.Requests = b.reqs
	ba.Header = b.Header
	ba.AdmissionHeader = b.AdmissionHeader
	b.response, b.pErr = send(ctx, ba)
	b.fillResults(ctx)
	if b.pErr == nil {
		__antithesis_instrumentation__.Notify(86822)
		b.pErr = roachpb.NewError(b.resultErr())
	} else {
		__antithesis_instrumentation__.Notify(86823)
	}
	__antithesis_instrumentation__.Notify(86821)
	return b.pErr.GoError()
}

func (db *DB) Run(ctx context.Context, b *Batch) error {
	__antithesis_instrumentation__.Notify(86824)
	if err := b.validate(); err != nil {
		__antithesis_instrumentation__.Notify(86826)
		return err
	} else {
		__antithesis_instrumentation__.Notify(86827)
	}
	__antithesis_instrumentation__.Notify(86825)
	return sendAndFill(ctx, db.send, b)
}

func (db *DB) NewTxn(ctx context.Context, debugName string) *Txn {
	__antithesis_instrumentation__.Notify(86828)

	nodeID, _ := db.ctx.NodeID.OptionalNodeID()
	txn := NewTxn(ctx, db, nodeID)
	txn.SetDebugName(debugName)
	return txn
}

func (db *DB) Txn(ctx context.Context, retryable func(context.Context, *Txn) error) error {
	__antithesis_instrumentation__.Notify(86829)
	return db.TxnWithAdmissionControl(
		ctx, roachpb.AdmissionHeader_OTHER, admission.NormalPri, retryable)
}

func (db *DB) TxnWithAdmissionControl(
	ctx context.Context,
	source roachpb.AdmissionHeader_Source,
	priority admission.WorkPriority,
	retryable func(context.Context, *Txn) error,
) error {
	__antithesis_instrumentation__.Notify(86830)

	nodeID, _ := db.ctx.NodeID.OptionalNodeID()
	txn := NewTxnWithAdmissionControl(ctx, db, nodeID, source, priority)
	txn.SetDebugName("unnamed")
	return runTxn(ctx, txn, retryable)
}

func (db *DB) TxnWithSteppingEnabled(
	ctx context.Context,
	qualityOfService sessiondatapb.QoSLevel,
	retryable func(context.Context, *Txn) error,
) error {
	__antithesis_instrumentation__.Notify(86831)

	nodeID, _ := db.ctx.NodeID.OptionalNodeID()
	txn := NewTxnWithSteppingEnabled(ctx, db, nodeID, qualityOfService)
	txn.SetDebugName("unnamed")
	return runTxn(ctx, txn, retryable)
}

func (db *DB) TxnRootKV(ctx context.Context, retryable func(context.Context, *Txn) error) error {
	__antithesis_instrumentation__.Notify(86832)
	nodeID, _ := db.ctx.NodeID.OptionalNodeID()
	txn := NewTxnRootKV(ctx, db, nodeID)
	txn.SetDebugName("unnamed")
	return runTxn(ctx, txn, retryable)
}

func runTxn(ctx context.Context, txn *Txn, retryable func(context.Context, *Txn) error) error {
	__antithesis_instrumentation__.Notify(86833)
	err := txn.exec(ctx, func(ctx context.Context, txn *Txn) error {
		__antithesis_instrumentation__.Notify(86837)
		return retryable(ctx, txn)
	})
	__antithesis_instrumentation__.Notify(86834)
	if err != nil {
		__antithesis_instrumentation__.Notify(86838)
		txn.CleanupOnError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(86839)
	}
	__antithesis_instrumentation__.Notify(86835)

	if errors.HasType(err, (*roachpb.TransactionRetryWithProtoRefreshError)(nil)) {
		__antithesis_instrumentation__.Notify(86840)
		return errors.Wrapf(err, "terminated retryable error")
	} else {
		__antithesis_instrumentation__.Notify(86841)
	}
	__antithesis_instrumentation__.Notify(86836)
	return err
}

func (db *DB) send(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(86842)
	return db.sendUsingSender(ctx, ba, db.NonTransactionalSender())
}

func (db *DB) sendUsingSender(
	ctx context.Context, ba roachpb.BatchRequest, sender Sender,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(86843)
	if len(ba.Requests) == 0 {
		__antithesis_instrumentation__.Notify(86848)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(86849)
	}
	__antithesis_instrumentation__.Notify(86844)
	if err := ba.ReadConsistency.SupportsBatch(ba); err != nil {
		__antithesis_instrumentation__.Notify(86850)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(86851)
	}
	__antithesis_instrumentation__.Notify(86845)
	if ba.UserPriority == 0 && func() bool {
		__antithesis_instrumentation__.Notify(86852)
		return db.ctx.UserPriority != 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(86853)
		ba.UserPriority = db.ctx.UserPriority
	} else {
		__antithesis_instrumentation__.Notify(86854)
	}
	__antithesis_instrumentation__.Notify(86846)

	br, pErr := sender.Send(ctx, ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(86855)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(86857)
			log.Infof(ctx, "failed batch: %s", pErr)
		} else {
			__antithesis_instrumentation__.Notify(86858)
		}
		__antithesis_instrumentation__.Notify(86856)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(86859)
	}
	__antithesis_instrumentation__.Notify(86847)
	return br, nil
}

func getOneErr(runErr error, b *Batch) error {
	__antithesis_instrumentation__.Notify(86860)
	if runErr != nil && func() bool {
		__antithesis_instrumentation__.Notify(86862)
		return len(b.Results) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(86863)
		return b.Results[0].Err
	} else {
		__antithesis_instrumentation__.Notify(86864)
	}
	__antithesis_instrumentation__.Notify(86861)
	return runErr
}

func getOneResult(runErr error, b *Batch) (Result, error) {
	__antithesis_instrumentation__.Notify(86865)
	if runErr != nil {
		__antithesis_instrumentation__.Notify(86868)
		if len(b.Results) > 0 {
			__antithesis_instrumentation__.Notify(86870)
			return b.Results[0], b.Results[0].Err
		} else {
			__antithesis_instrumentation__.Notify(86871)
		}
		__antithesis_instrumentation__.Notify(86869)
		return Result{Err: runErr}, runErr
	} else {
		__antithesis_instrumentation__.Notify(86872)
	}
	__antithesis_instrumentation__.Notify(86866)
	res := b.Results[0]
	if res.Err != nil {
		__antithesis_instrumentation__.Notify(86873)
		panic("run succeeded even through the result has an error")
	} else {
		__antithesis_instrumentation__.Notify(86874)
	}
	__antithesis_instrumentation__.Notify(86867)
	return res, nil
}

func getOneRow(runErr error, b *Batch) (KeyValue, error) {
	__antithesis_instrumentation__.Notify(86875)
	res, err := getOneResult(runErr, b)
	if err != nil {
		__antithesis_instrumentation__.Notify(86877)
		return KeyValue{}, err
	} else {
		__antithesis_instrumentation__.Notify(86878)
	}
	__antithesis_instrumentation__.Notify(86876)
	return res.Rows[0], nil
}

func IncrementValRetryable(ctx context.Context, db *DB, key roachpb.Key, inc int64) (int64, error) {
	__antithesis_instrumentation__.Notify(86879)
	var err error
	var res KeyValue
	for r := retry.Start(base.DefaultRetryOptions()); r.Next(); {
		__antithesis_instrumentation__.Notify(86881)
		res, err = db.Inc(ctx, key, inc)
		if errors.HasType(err, (*roachpb.UnhandledRetryableError)(nil)) || func() bool {
			__antithesis_instrumentation__.Notify(86883)
			return errors.HasType(err, (*roachpb.AmbiguousResultError)(nil)) == true
		}() == true {
			__antithesis_instrumentation__.Notify(86884)
			continue
		} else {
			__antithesis_instrumentation__.Notify(86885)
		}
		__antithesis_instrumentation__.Notify(86882)
		break
	}
	__antithesis_instrumentation__.Notify(86880)
	return res.ValueInt(), err
}
