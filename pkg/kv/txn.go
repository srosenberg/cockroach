package kv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

const asyncRollbackTimeout = time.Minute

type Txn struct {
	db *DB

	typ TxnType

	gatewayNodeID roachpb.NodeID

	commitTriggers []func(ctx context.Context)

	systemConfigTrigger bool

	mu struct {
		syncutil.Mutex
		ID           uuid.UUID
		debugName    string
		userPriority roachpb.UserPriority

		previousIDs map[uuid.UUID]struct{}

		sender TxnSender

		deadline *hlc.Timestamp
	}

	admissionHeader roachpb.AdmissionHeader
}

func NewTxn(ctx context.Context, db *DB, gatewayNodeID roachpb.NodeID) *Txn {
	__antithesis_instrumentation__.Notify(127730)
	return NewTxnWithAdmissionControl(
		ctx, db, gatewayNodeID, roachpb.AdmissionHeader_OTHER, admission.NormalPri)
}

func NewTxnWithAdmissionControl(
	ctx context.Context,
	db *DB,
	gatewayNodeID roachpb.NodeID,
	source roachpb.AdmissionHeader_Source,
	priority admission.WorkPriority,
) *Txn {
	__antithesis_instrumentation__.Notify(127731)
	if db == nil {
		__antithesis_instrumentation__.Notify(127733)
		panic(errors.WithContextTags(
			errors.AssertionFailedf("attempting to create txn with nil db"), ctx))
	} else {
		__antithesis_instrumentation__.Notify(127734)
	}
	__antithesis_instrumentation__.Notify(127732)

	now := db.clock.NowAsClockTimestamp()
	kvTxn := roachpb.MakeTransaction(
		"unnamed",
		nil,
		roachpb.NormalUserPriority,
		now.ToTimestamp(),
		db.clock.MaxOffset().Nanoseconds(),
		int32(db.ctx.NodeID.SQLInstanceID()),
	)
	txn := NewTxnFromProto(ctx, db, gatewayNodeID, now, RootTxn, &kvTxn)
	txn.admissionHeader = roachpb.AdmissionHeader{
		CreateTime: db.clock.PhysicalNow(),
		Priority:   int32(priority),
		Source:     source,
	}
	return txn
}

func NewTxnWithSteppingEnabled(
	ctx context.Context,
	db *DB,
	gatewayNodeID roachpb.NodeID,
	qualityOfService sessiondatapb.QoSLevel,
) *Txn {
	__antithesis_instrumentation__.Notify(127735)
	txn := NewTxnWithAdmissionControl(ctx, db, gatewayNodeID,
		roachpb.AdmissionHeader_FROM_SQL, admission.WorkPriority(qualityOfService))
	_ = txn.ConfigureStepping(ctx, SteppingEnabled)
	return txn
}

func NewTxnRootKV(ctx context.Context, db *DB, gatewayNodeID roachpb.NodeID) *Txn {
	__antithesis_instrumentation__.Notify(127736)
	return NewTxnWithAdmissionControl(
		ctx, db, gatewayNodeID, roachpb.AdmissionHeader_ROOT_KV, admission.NormalPri)
}

func NewTxnFromProto(
	ctx context.Context,
	db *DB,
	gatewayNodeID roachpb.NodeID,
	now hlc.ClockTimestamp,
	typ TxnType,
	proto *roachpb.Transaction,
) *Txn {
	__antithesis_instrumentation__.Notify(127737)

	if gatewayNodeID != 0 && func() bool {
		__antithesis_instrumentation__.Notify(127739)
		return typ == RootTxn == true
	}() == true {
		__antithesis_instrumentation__.Notify(127740)
		proto.UpdateObservedTimestamp(gatewayNodeID, now)
	} else {
		__antithesis_instrumentation__.Notify(127741)
	}
	__antithesis_instrumentation__.Notify(127738)

	txn := &Txn{db: db, typ: typ, gatewayNodeID: gatewayNodeID}
	txn.mu.ID = proto.ID
	txn.mu.userPriority = roachpb.NormalUserPriority
	txn.mu.sender = db.factory.RootTransactionalSender(proto, txn.mu.userPriority)
	return txn
}

func NewLeafTxn(
	ctx context.Context, db *DB, gatewayNodeID roachpb.NodeID, tis *roachpb.LeafTxnInputState,
) *Txn {
	__antithesis_instrumentation__.Notify(127742)
	if db == nil {
		__antithesis_instrumentation__.Notify(127745)
		panic(errors.WithContextTags(
			errors.AssertionFailedf("attempting to create leaf txn with nil db for Transaction: %s", tis.Txn), ctx))
	} else {
		__antithesis_instrumentation__.Notify(127746)
	}
	__antithesis_instrumentation__.Notify(127743)
	if tis.Txn.Status != roachpb.PENDING {
		__antithesis_instrumentation__.Notify(127747)
		panic(errors.WithContextTags(
			errors.AssertionFailedf("can't create leaf txn with non-PENDING proto: %s", tis.Txn), ctx))
	} else {
		__antithesis_instrumentation__.Notify(127748)
	}
	__antithesis_instrumentation__.Notify(127744)
	tis.Txn.AssertInitialized(ctx)
	txn := &Txn{db: db, typ: LeafTxn, gatewayNodeID: gatewayNodeID}
	txn.mu.ID = tis.Txn.ID
	txn.mu.userPriority = roachpb.NormalUserPriority
	txn.mu.sender = db.factory.LeafTransactionalSender(tis)
	return txn
}

func (txn *Txn) DB() *DB {
	__antithesis_instrumentation__.Notify(127749)
	return txn.db
}

func (txn *Txn) Sender() TxnSender {
	__antithesis_instrumentation__.Notify(127750)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender
}

func (txn *Txn) ID() uuid.UUID {
	__antithesis_instrumentation__.Notify(127751)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.ID
}

func (txn *Txn) Epoch() enginepb.TxnEpoch {
	__antithesis_instrumentation__.Notify(127752)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.Epoch()
}

func (txn *Txn) statusLocked() roachpb.TransactionStatus {
	__antithesis_instrumentation__.Notify(127753)
	return txn.mu.sender.TxnStatus()
}

func (txn *Txn) IsCommitted() bool {
	__antithesis_instrumentation__.Notify(127754)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.statusLocked() == roachpb.COMMITTED
}

func (txn *Txn) IsAborted() bool {
	__antithesis_instrumentation__.Notify(127755)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.statusLocked() == roachpb.ABORTED
}

func (txn *Txn) IsOpen() bool {
	__antithesis_instrumentation__.Notify(127756)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.statusLocked() == roachpb.PENDING
}

func (txn *Txn) SetUserPriority(userPriority roachpb.UserPriority) error {
	__antithesis_instrumentation__.Notify(127757)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127761)
		panic(errors.AssertionFailedf("SetUserPriority() called on leaf txn"))
	} else {
		__antithesis_instrumentation__.Notify(127762)
	}
	__antithesis_instrumentation__.Notify(127758)

	txn.mu.Lock()
	defer txn.mu.Unlock()
	if txn.mu.userPriority == userPriority {
		__antithesis_instrumentation__.Notify(127763)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(127764)
	}
	__antithesis_instrumentation__.Notify(127759)

	if userPriority < roachpb.MinUserPriority || func() bool {
		__antithesis_instrumentation__.Notify(127765)
		return userPriority > roachpb.MaxUserPriority == true
	}() == true {
		__antithesis_instrumentation__.Notify(127766)
		return errors.AssertionFailedf("the given user priority %f is out of the allowed range [%f, %f]",
			userPriority, roachpb.MinUserPriority, roachpb.MaxUserPriority)
	} else {
		__antithesis_instrumentation__.Notify(127767)
	}
	__antithesis_instrumentation__.Notify(127760)

	txn.mu.userPriority = userPriority
	return txn.mu.sender.SetUserPriority(userPriority)
}

func (txn *Txn) TestingSetPriority(priority enginepb.TxnPriority) {
	__antithesis_instrumentation__.Notify(127768)
	txn.mu.Lock()

	txn.mu.userPriority = roachpb.UserPriority(-priority)
	if err := txn.mu.sender.SetUserPriority(txn.mu.userPriority); err != nil {
		__antithesis_instrumentation__.Notify(127770)
		log.Fatalf(context.TODO(), "%+v", err)
	} else {
		__antithesis_instrumentation__.Notify(127771)
	}
	__antithesis_instrumentation__.Notify(127769)
	txn.mu.Unlock()
}

func (txn *Txn) UserPriority() roachpb.UserPriority {
	__antithesis_instrumentation__.Notify(127772)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.userPriority
}

func (txn *Txn) SetDebugName(name string) {
	__antithesis_instrumentation__.Notify(127773)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127775)
		panic(errors.AssertionFailedf("SetDebugName() called on leaf txn"))
	} else {
		__antithesis_instrumentation__.Notify(127776)
	}
	__antithesis_instrumentation__.Notify(127774)

	txn.mu.Lock()
	defer txn.mu.Unlock()

	txn.mu.sender.SetDebugName(name)
	txn.mu.debugName = name
}

func (txn *Txn) DebugName() string {
	__antithesis_instrumentation__.Notify(127777)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.debugNameLocked()
}

func (txn *Txn) debugNameLocked() string {
	__antithesis_instrumentation__.Notify(127778)
	return fmt.Sprintf("%s (id: %s)", txn.mu.debugName, txn.mu.ID)
}

func (txn *Txn) String() string {
	__antithesis_instrumentation__.Notify(127779)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.String()
}

func (txn *Txn) ReadTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127780)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.readTimestampLocked()
}

func (txn *Txn) readTimestampLocked() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127781)
	return txn.mu.sender.ReadTimestamp()
}

func (txn *Txn) CommitTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127782)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.CommitTimestamp()
}

func (txn *Txn) CommitTimestampFixed() bool {
	__antithesis_instrumentation__.Notify(127783)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.CommitTimestampFixed()
}

func (txn *Txn) ProvisionalCommitTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127784)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.ProvisionalCommitTimestamp()
}

func (txn *Txn) RequiredFrontier() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127785)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.RequiredFrontier()
}

func (txn *Txn) DeprecatedSetSystemConfigTrigger(forSystemTenant bool) error {
	__antithesis_instrumentation__.Notify(127786)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127790)
		return errors.AssertionFailedf("DeprecatedSetSystemConfigTrigger() called on leaf txn")
	} else {
		__antithesis_instrumentation__.Notify(127791)
	}
	__antithesis_instrumentation__.Notify(127787)
	if !forSystemTenant {
		__antithesis_instrumentation__.Notify(127792)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(127793)
	}
	__antithesis_instrumentation__.Notify(127788)

	txn.mu.Lock()
	defer txn.mu.Unlock()
	if err := txn.mu.sender.AnchorOnSystemConfigRange(); err != nil {
		__antithesis_instrumentation__.Notify(127794)
		return err
	} else {
		__antithesis_instrumentation__.Notify(127795)
	}
	__antithesis_instrumentation__.Notify(127789)
	txn.systemConfigTrigger = true
	return nil
}

func (txn *Txn) DisablePipelining() error {
	__antithesis_instrumentation__.Notify(127796)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127798)
		return errors.AssertionFailedf("DisablePipelining() called on leaf txn")
	} else {
		__antithesis_instrumentation__.Notify(127799)
	}
	__antithesis_instrumentation__.Notify(127797)

	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.DisablePipelining()
}

func (txn *Txn) NewBatch() *Batch {
	__antithesis_instrumentation__.Notify(127800)
	return &Batch{txn: txn, AdmissionHeader: txn.AdmissionHeader()}
}

func (txn *Txn) Get(ctx context.Context, key interface{}) (KeyValue, error) {
	__antithesis_instrumentation__.Notify(127801)
	b := txn.NewBatch()
	b.Get(key)
	return getOneRow(txn.Run(ctx, b), b)
}

func (txn *Txn) GetForUpdate(ctx context.Context, key interface{}) (KeyValue, error) {
	__antithesis_instrumentation__.Notify(127802)
	b := txn.NewBatch()
	b.GetForUpdate(key)
	return getOneRow(txn.Run(ctx, b), b)
}

func (txn *Txn) GetProto(ctx context.Context, key interface{}, msg protoutil.Message) error {
	__antithesis_instrumentation__.Notify(127803)
	_, err := txn.GetProtoTs(ctx, key, msg)
	return err
}

func (txn *Txn) GetProtoTs(
	ctx context.Context, key interface{}, msg protoutil.Message,
) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(127804)
	r, err := txn.Get(ctx, key)
	if err != nil {
		__antithesis_instrumentation__.Notify(127807)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(127808)
	}
	__antithesis_instrumentation__.Notify(127805)
	if err := r.ValueProto(msg); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(127809)
		return r.Value == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(127810)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(127811)
	}
	__antithesis_instrumentation__.Notify(127806)
	return r.Value.Timestamp, nil
}

func (txn *Txn) Put(ctx context.Context, key, value interface{}) error {
	__antithesis_instrumentation__.Notify(127812)
	b := txn.NewBatch()
	b.Put(key, value)
	return getOneErr(txn.Run(ctx, b), b)
}

func (txn *Txn) CPut(ctx context.Context, key, value interface{}, expValue []byte) error {
	__antithesis_instrumentation__.Notify(127813)
	b := txn.NewBatch()
	b.CPut(key, value, expValue)
	return getOneErr(txn.Run(ctx, b), b)
}

func (txn *Txn) InitPut(ctx context.Context, key, value interface{}, failOnTombstones bool) error {
	__antithesis_instrumentation__.Notify(127814)
	b := txn.NewBatch()
	b.InitPut(key, value, failOnTombstones)
	return getOneErr(txn.Run(ctx, b), b)
}

func (txn *Txn) Inc(ctx context.Context, key interface{}, value int64) (KeyValue, error) {
	__antithesis_instrumentation__.Notify(127815)
	b := txn.NewBatch()
	b.Inc(key, value)
	return getOneRow(txn.Run(ctx, b), b)
}

func (txn *Txn) scan(
	ctx context.Context, begin, end interface{}, maxRows int64, isReverse, forUpdate bool,
) ([]KeyValue, error) {
	__antithesis_instrumentation__.Notify(127816)
	b := txn.NewBatch()
	if maxRows > 0 {
		__antithesis_instrumentation__.Notify(127818)
		b.Header.MaxSpanRequestKeys = maxRows
	} else {
		__antithesis_instrumentation__.Notify(127819)
	}
	__antithesis_instrumentation__.Notify(127817)
	b.scan(begin, end, isReverse, forUpdate)
	r, err := getOneResult(txn.Run(ctx, b), b)
	return r.Rows, err
}

func (txn *Txn) Scan(
	ctx context.Context, begin, end interface{}, maxRows int64,
) ([]KeyValue, error) {
	__antithesis_instrumentation__.Notify(127820)
	return txn.scan(ctx, begin, end, maxRows, false, false)
}

func (txn *Txn) ScanForUpdate(
	ctx context.Context, begin, end interface{}, maxRows int64,
) ([]KeyValue, error) {
	__antithesis_instrumentation__.Notify(127821)
	return txn.scan(ctx, begin, end, maxRows, false, true)
}

func (txn *Txn) ReverseScan(
	ctx context.Context, begin, end interface{}, maxRows int64,
) ([]KeyValue, error) {
	__antithesis_instrumentation__.Notify(127822)
	return txn.scan(ctx, begin, end, maxRows, true, false)
}

func (txn *Txn) ReverseScanForUpdate(
	ctx context.Context, begin, end interface{}, maxRows int64,
) ([]KeyValue, error) {
	__antithesis_instrumentation__.Notify(127823)
	return txn.scan(ctx, begin, end, maxRows, true, true)
}

func (txn *Txn) Iterate(
	ctx context.Context, begin, end interface{}, pageSize int, f func([]KeyValue) error,
) error {
	__antithesis_instrumentation__.Notify(127824)
	for {
		__antithesis_instrumentation__.Notify(127825)
		rows, err := txn.Scan(ctx, begin, end, int64(pageSize))
		if err != nil {
			__antithesis_instrumentation__.Notify(127830)
			return err
		} else {
			__antithesis_instrumentation__.Notify(127831)
		}
		__antithesis_instrumentation__.Notify(127826)
		if len(rows) == 0 {
			__antithesis_instrumentation__.Notify(127832)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(127833)
		}
		__antithesis_instrumentation__.Notify(127827)
		if err := f(rows); err != nil {
			__antithesis_instrumentation__.Notify(127834)
			return errors.Wrap(err, "running iterate callback")
		} else {
			__antithesis_instrumentation__.Notify(127835)
		}
		__antithesis_instrumentation__.Notify(127828)
		if len(rows) < pageSize {
			__antithesis_instrumentation__.Notify(127836)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(127837)
		}
		__antithesis_instrumentation__.Notify(127829)
		begin = rows[len(rows)-1].Key.Next()
	}
}

func (txn *Txn) Del(ctx context.Context, keys ...interface{}) error {
	__antithesis_instrumentation__.Notify(127838)
	b := txn.NewBatch()
	b.Del(keys...)
	return getOneErr(txn.Run(ctx, b), b)
}

func (txn *Txn) DelRange(
	ctx context.Context, begin, end interface{}, returnKeys bool,
) ([]roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(127839)
	b := txn.NewBatch()
	b.DelRange(begin, end, returnKeys)
	r, err := getOneResult(txn.Run(ctx, b), b)
	return r.Keys, err
}

func (txn *Txn) Run(ctx context.Context, b *Batch) error {
	__antithesis_instrumentation__.Notify(127840)
	if err := b.validate(); err != nil {
		__antithesis_instrumentation__.Notify(127842)
		return err
	} else {
		__antithesis_instrumentation__.Notify(127843)
	}
	__antithesis_instrumentation__.Notify(127841)
	return sendAndFill(ctx, txn.Send, b)
}

func (txn *Txn) commit(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(127844)

	et := endTxnReq(true, txn.deadline(), txn.systemConfigTrigger)
	ba := roachpb.BatchRequest{Requests: et.unionArr[:]}
	_, pErr := txn.Send(ctx, ba)
	if pErr == nil {
		__antithesis_instrumentation__.Notify(127846)
		for _, t := range txn.commitTriggers {
			__antithesis_instrumentation__.Notify(127847)
			t(ctx)
		}
	} else {
		__antithesis_instrumentation__.Notify(127848)
	}
	__antithesis_instrumentation__.Notify(127845)
	return pErr.GoError()
}

func (txn *Txn) CleanupOnError(ctx context.Context, err error) {
	__antithesis_instrumentation__.Notify(127849)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127852)
		panic(errors.WithContextTags(errors.AssertionFailedf("CleanupOnError() called on leaf txn"), ctx))
	} else {
		__antithesis_instrumentation__.Notify(127853)
	}
	__antithesis_instrumentation__.Notify(127850)

	if err == nil {
		__antithesis_instrumentation__.Notify(127854)
		panic(errors.WithContextTags(errors.AssertionFailedf("CleanupOnError() called with nil error"), ctx))
	} else {
		__antithesis_instrumentation__.Notify(127855)
	}
	__antithesis_instrumentation__.Notify(127851)
	if replyErr := txn.rollback(ctx); replyErr != nil {
		__antithesis_instrumentation__.Notify(127856)
		if _, ok := replyErr.GetDetail().(*roachpb.TransactionStatusError); ok || func() bool {
			__antithesis_instrumentation__.Notify(127857)
			return txn.IsAborted() == true
		}() == true {
			__antithesis_instrumentation__.Notify(127858)
			log.Eventf(ctx, "failure aborting transaction: %s; abort caused by: %s", replyErr, err)
		} else {
			__antithesis_instrumentation__.Notify(127859)
			log.Warningf(ctx, "failure aborting transaction: %s; abort caused by: %s", replyErr, err)
		}
	} else {
		__antithesis_instrumentation__.Notify(127860)
	}
}

func (txn *Txn) Commit(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(127861)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127863)
		return errors.WithContextTags(errors.AssertionFailedf("Commit() called on leaf txn"), ctx)
	} else {
		__antithesis_instrumentation__.Notify(127864)
	}
	__antithesis_instrumentation__.Notify(127862)

	return txn.commit(ctx)
}

func (txn *Txn) CommitInBatch(ctx context.Context, b *Batch) error {
	__antithesis_instrumentation__.Notify(127865)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127868)
		return errors.WithContextTags(errors.AssertionFailedf("CommitInBatch() called on leaf txn"), ctx)
	} else {
		__antithesis_instrumentation__.Notify(127869)
	}
	__antithesis_instrumentation__.Notify(127866)

	if txn != b.txn {
		__antithesis_instrumentation__.Notify(127870)
		return errors.Errorf("a batch b can only be committed by b.txn")
	} else {
		__antithesis_instrumentation__.Notify(127871)
	}
	__antithesis_instrumentation__.Notify(127867)
	et := endTxnReq(true, txn.deadline(), txn.systemConfigTrigger)
	b.growReqs(1)
	b.reqs[len(b.reqs)-1].Value = &et.union
	b.initResult(1, 0, b.raw, nil)
	return txn.Run(ctx, b)
}

func (txn *Txn) CommitOrCleanup(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(127872)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127875)
		return errors.WithContextTags(errors.AssertionFailedf("CommitOrCleanup() called on leaf txn"), ctx)
	} else {
		__antithesis_instrumentation__.Notify(127876)
	}
	__antithesis_instrumentation__.Notify(127873)

	err := txn.commit(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(127877)
		txn.CleanupOnError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(127878)
	}
	__antithesis_instrumentation__.Notify(127874)
	return err
}

func (txn *Txn) UpdateDeadline(ctx context.Context, deadline hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(127879)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127882)
		panic(errors.WithContextTags(errors.AssertionFailedf("UpdateDeadline() called on leaf txn"), ctx))
	} else {
		__antithesis_instrumentation__.Notify(127883)
	}
	__antithesis_instrumentation__.Notify(127880)

	txn.mu.Lock()
	defer txn.mu.Unlock()

	readTimestamp := txn.readTimestampLocked()
	if deadline.Less(readTimestamp) {
		__antithesis_instrumentation__.Notify(127884)
		return errors.AssertionFailedf("deadline below read timestamp is nonsensical; "+
			"txn has would have no chance to commit. Deadline: %s. Read timestamp: %s Previous Deadline: %s.",
			deadline, readTimestamp, txn.mu.deadline)
	} else {
		__antithesis_instrumentation__.Notify(127885)
	}
	__antithesis_instrumentation__.Notify(127881)
	txn.mu.deadline = new(hlc.Timestamp)
	*txn.mu.deadline = deadline
	return nil
}

func (txn *Txn) DeadlineLikelySufficient(sv *settings.Values) bool {
	__antithesis_instrumentation__.Notify(127886)
	txn.mu.Lock()
	defer txn.mu.Unlock()

	getTargetTS := func() hlc.Timestamp {
		__antithesis_instrumentation__.Notify(127888)
		now := txn.db.Clock().NowAsClockTimestamp()
		maxClockOffset := txn.db.Clock().MaxOffset()
		lagTargetDuration := closedts.TargetDuration.Get(sv)
		leadTargetOverride := closedts.LeadForGlobalReadsOverride.Get(sv)
		sideTransportCloseInterval := closedts.SideTransportCloseInterval.Get(sv)
		return closedts.TargetForPolicy(now, maxClockOffset,
			lagTargetDuration, leadTargetOverride, sideTransportCloseInterval,
			roachpb.LEAD_FOR_GLOBAL_READS).Add(int64(time.Second), 0)
	}
	__antithesis_instrumentation__.Notify(127887)

	return txn.mu.deadline != nil && func() bool {
		__antithesis_instrumentation__.Notify(127889)
		return !txn.mu.deadline.IsEmpty() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(127890)
		return (txn.mu.deadline.Less(txn.mu.sender.ProvisionalCommitTimestamp()) || func() bool {
			__antithesis_instrumentation__.Notify(127891)
			return txn.mu.deadline.Less(getTargetTS()) == true
		}() == true) == true
	}() == true
}

func (txn *Txn) resetDeadlineLocked() {
	__antithesis_instrumentation__.Notify(127892)
	txn.mu.deadline = nil
}

func (txn *Txn) Rollback(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(127893)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127895)
		return errors.WithContextTags(errors.AssertionFailedf("Rollback() called on leaf txn"), ctx)
	} else {
		__antithesis_instrumentation__.Notify(127896)
	}
	__antithesis_instrumentation__.Notify(127894)

	return txn.rollback(ctx).GoError()
}

func (txn *Txn) rollback(ctx context.Context) *roachpb.Error {
	__antithesis_instrumentation__.Notify(127897)
	log.VEventf(ctx, 2, "rolling back transaction")

	if ctx.Err() == nil {
		__antithesis_instrumentation__.Notify(127900)

		et := endTxnReq(false, nil, false)
		ba := roachpb.BatchRequest{Requests: et.unionArr[:]}
		_, pErr := txn.Send(ctx, ba)
		if pErr == nil {
			__antithesis_instrumentation__.Notify(127902)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(127903)
		}
		__antithesis_instrumentation__.Notify(127901)

		if ctx.Err() == nil {
			__antithesis_instrumentation__.Notify(127904)
			return pErr
		} else {
			__antithesis_instrumentation__.Notify(127905)
		}
	} else {
		__antithesis_instrumentation__.Notify(127906)
	}
	__antithesis_instrumentation__.Notify(127898)

	stopper := txn.db.ctx.Stopper
	ctx, cancel := stopper.WithCancelOnQuiesce(txn.db.AnnotateCtx(context.Background()))
	if err := stopper.RunAsyncTask(ctx, "async-rollback", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(127907)
		defer cancel()

		et := endTxnReq(false, nil, false)
		ba := roachpb.BatchRequest{Requests: et.unionArr[:]}
		_ = contextutil.RunWithTimeout(ctx, "async txn rollback", asyncRollbackTimeout,
			func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(127908)
				if _, pErr := txn.Send(ctx, ba); pErr != nil {
					__antithesis_instrumentation__.Notify(127910)
					if statusErr, ok := pErr.GetDetail().(*roachpb.TransactionStatusError); ok && func() bool {
						__antithesis_instrumentation__.Notify(127911)
						return statusErr.Reason == roachpb.TransactionStatusError_REASON_TXN_COMMITTED == true
					}() == true {
						__antithesis_instrumentation__.Notify(127912)

						log.VEventf(ctx, 2, "async rollback failed: %s", pErr)
					} else {
						__antithesis_instrumentation__.Notify(127913)
						log.Infof(ctx, "async rollback failed: %s", pErr)
					}
				} else {
					__antithesis_instrumentation__.Notify(127914)
				}
				__antithesis_instrumentation__.Notify(127909)
				return nil
			})
	}); err != nil {
		__antithesis_instrumentation__.Notify(127915)
		cancel()
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(127916)
	}
	__antithesis_instrumentation__.Notify(127899)
	return nil
}

func (txn *Txn) AddCommitTrigger(trigger func(ctx context.Context)) {
	__antithesis_instrumentation__.Notify(127917)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127919)
		panic(errors.AssertionFailedf("AddCommitTrigger() called on leaf txn"))
	} else {
		__antithesis_instrumentation__.Notify(127920)
	}
	__antithesis_instrumentation__.Notify(127918)

	txn.commitTriggers = append(txn.commitTriggers, trigger)
}

type endTxnReqAlloc struct {
	req      roachpb.EndTxnRequest
	union    roachpb.RequestUnion_EndTxn
	unionArr [1]roachpb.RequestUnion
}

func endTxnReq(commit bool, deadline *hlc.Timestamp, hasTrigger bool) *endTxnReqAlloc {
	__antithesis_instrumentation__.Notify(127921)
	alloc := new(endTxnReqAlloc)
	alloc.req.Commit = commit
	alloc.req.Deadline = deadline
	if hasTrigger {
		__antithesis_instrumentation__.Notify(127923)
		alloc.req.InternalCommitTrigger = &roachpb.InternalCommitTrigger{
			ModifiedSpanTrigger: &roachpb.ModifiedSpanTrigger{
				SystemConfigSpan: true,
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(127924)
	}
	__antithesis_instrumentation__.Notify(127922)
	alloc.union.EndTxn = &alloc.req
	alloc.unionArr[0].Value = &alloc.union
	return alloc
}

type AutoCommitError struct {
	cause error
}

func (e *AutoCommitError) Cause() error {
	__antithesis_instrumentation__.Notify(127925)
	return e.cause
}

func (e *AutoCommitError) Error() string {
	__antithesis_instrumentation__.Notify(127926)
	return e.cause.Error()
}

func (txn *Txn) exec(ctx context.Context, fn func(context.Context, *Txn) error) (err error) {
	__antithesis_instrumentation__.Notify(127927)

	for {
		__antithesis_instrumentation__.Notify(127929)
		if err := ctx.Err(); err != nil {
			__antithesis_instrumentation__.Notify(127934)
			return err
		} else {
			__antithesis_instrumentation__.Notify(127935)
		}
		__antithesis_instrumentation__.Notify(127930)
		err = fn(ctx, txn)

		if err == nil {
			__antithesis_instrumentation__.Notify(127936)
			if !txn.IsCommitted() {
				__antithesis_instrumentation__.Notify(127937)
				err = txn.Commit(ctx)
				log.Eventf(ctx, "client.Txn did AutoCommit. err: %v", err)
				if err != nil {
					__antithesis_instrumentation__.Notify(127938)
					if !errors.HasType(err, (*roachpb.TransactionRetryWithProtoRefreshError)(nil)) {
						__antithesis_instrumentation__.Notify(127939)

						err = &AutoCommitError{cause: err}
					} else {
						__antithesis_instrumentation__.Notify(127940)
					}
				} else {
					__antithesis_instrumentation__.Notify(127941)
				}
			} else {
				__antithesis_instrumentation__.Notify(127942)
			}
		} else {
			__antithesis_instrumentation__.Notify(127943)
		}
		__antithesis_instrumentation__.Notify(127931)

		var retryable bool
		if err != nil {
			__antithesis_instrumentation__.Notify(127944)
			if errors.HasType(err, (*roachpb.UnhandledRetryableError)(nil)) {
				__antithesis_instrumentation__.Notify(127945)
				if txn.typ == RootTxn {
					__antithesis_instrumentation__.Notify(127946)

					log.Fatalf(ctx, "unexpected UnhandledRetryableError at the txn.exec() level: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(127947)
				}
			} else {
				__antithesis_instrumentation__.Notify(127948)
				if t := (*roachpb.TransactionRetryWithProtoRefreshError)(nil); errors.As(err, &t) {
					__antithesis_instrumentation__.Notify(127949)
					if !txn.IsRetryableErrMeantForTxn(*t) {
						__antithesis_instrumentation__.Notify(127951)

						return errors.Wrapf(err, "retryable error from another txn")
					} else {
						__antithesis_instrumentation__.Notify(127952)
					}
					__antithesis_instrumentation__.Notify(127950)
					retryable = true
				} else {
					__antithesis_instrumentation__.Notify(127953)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(127954)
		}
		__antithesis_instrumentation__.Notify(127932)

		if !retryable {
			__antithesis_instrumentation__.Notify(127955)
			break
		} else {
			__antithesis_instrumentation__.Notify(127956)
		}
		__antithesis_instrumentation__.Notify(127933)

		txn.PrepareForRetry(ctx)
	}
	__antithesis_instrumentation__.Notify(127928)

	return err
}

func (txn *Txn) PrepareForRetry(ctx context.Context) {
	__antithesis_instrumentation__.Notify(127957)

	txn.commitTriggers = nil

	txn.mu.Lock()
	defer txn.mu.Unlock()

	retryErr := txn.mu.sender.GetTxnRetryableErr(ctx)
	if retryErr == nil {
		__antithesis_instrumentation__.Notify(127960)
		return
	} else {
		__antithesis_instrumentation__.Notify(127961)
	}
	__antithesis_instrumentation__.Notify(127958)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127962)
		panic(errors.WithContextTags(errors.NewAssertionErrorWithWrappedErrf(
			retryErr, "PrepareForRetry() called on leaf txn"), ctx))
	} else {
		__antithesis_instrumentation__.Notify(127963)
	}
	__antithesis_instrumentation__.Notify(127959)
	log.VEventf(ctx, 2, "retrying transaction: %s because of a retryable error: %s",
		txn.debugNameLocked(), retryErr)
	txn.handleRetryableErrLocked(ctx, retryErr)
}

func (txn *Txn) IsRetryableErrMeantForTxn(
	retryErr roachpb.TransactionRetryWithProtoRefreshError,
) bool {
	__antithesis_instrumentation__.Notify(127964)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(127967)
		panic(errors.AssertionFailedf("IsRetryableErrMeantForTxn() called on leaf txn"))
	} else {
		__antithesis_instrumentation__.Notify(127968)
	}
	__antithesis_instrumentation__.Notify(127965)

	txn.mu.Lock()
	defer txn.mu.Unlock()

	errTxnID := retryErr.TxnID

	if _, ok := txn.mu.previousIDs[errTxnID]; ok {
		__antithesis_instrumentation__.Notify(127969)
		return true
	} else {
		__antithesis_instrumentation__.Notify(127970)
	}
	__antithesis_instrumentation__.Notify(127966)

	return errTxnID == txn.mu.ID
}

func (txn *Txn) Send(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127971)

	if txn.gatewayNodeID != 0 {
		__antithesis_instrumentation__.Notify(127977)
		ba.Header.GatewayNodeID = txn.gatewayNodeID
	} else {
		__antithesis_instrumentation__.Notify(127978)
	}
	__antithesis_instrumentation__.Notify(127972)

	if ba.BoundedStaleness != nil {
		__antithesis_instrumentation__.Notify(127979)
		return nil, roachpb.NewError(errors.AssertionFailedf(
			"bounded staleness header passed to Txn.Send: %s", ba.String()))
	} else {
		__antithesis_instrumentation__.Notify(127980)
	}
	__antithesis_instrumentation__.Notify(127973)

	if ba.AdmissionHeader.CreateTime == 0 {
		__antithesis_instrumentation__.Notify(127981)
		noMem := ba.AdmissionHeader.NoMemoryReservedAtSource
		ba.AdmissionHeader = txn.AdmissionHeader()
		ba.AdmissionHeader.NoMemoryReservedAtSource = noMem
	} else {
		__antithesis_instrumentation__.Notify(127982)
	}
	__antithesis_instrumentation__.Notify(127974)

	txn.mu.Lock()
	requestTxnID := txn.mu.ID
	sender := txn.mu.sender
	txn.mu.Unlock()
	br, pErr := txn.db.sendUsingSender(ctx, ba, sender)
	if pErr == nil {
		__antithesis_instrumentation__.Notify(127983)
		return br, nil
	} else {
		__antithesis_instrumentation__.Notify(127984)
	}
	__antithesis_instrumentation__.Notify(127975)

	if retryErr, ok := pErr.GetDetail().(*roachpb.TransactionRetryWithProtoRefreshError); ok {
		__antithesis_instrumentation__.Notify(127985)
		if requestTxnID != retryErr.TxnID {
			__antithesis_instrumentation__.Notify(127986)

			log.Fatalf(ctx, "retryable error for the wrong txn. "+
				"requestTxnID: %s, retryErr.TxnID: %s. retryErr: %s",
				requestTxnID, retryErr.TxnID, retryErr)
		} else {
			__antithesis_instrumentation__.Notify(127987)
		}
	} else {
		__antithesis_instrumentation__.Notify(127988)
	}
	__antithesis_instrumentation__.Notify(127976)
	return br, pErr
}

func (txn *Txn) handleRetryableErrLocked(
	ctx context.Context, retryErr *roachpb.TransactionRetryWithProtoRefreshError,
) {
	__antithesis_instrumentation__.Notify(127989)
	txn.resetDeadlineLocked()
	txn.replaceRootSenderIfTxnAbortedLocked(ctx, retryErr, retryErr.TxnID)
}

func (txn *Txn) NegotiateAndSend(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127990)
	if err := txn.checkNegotiateAndSendPreconditions(ctx, ba); err != nil {
		__antithesis_instrumentation__.Notify(127995)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(127996)
	}
	__antithesis_instrumentation__.Notify(127991)
	if err := txn.applyDeadlineToBoundedStaleness(ctx, ba.BoundedStaleness); err != nil {
		__antithesis_instrumentation__.Notify(127997)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(127998)
	}
	__antithesis_instrumentation__.Notify(127992)

	br, pErr := txn.DB().GetFactory().NonTransactionalSender().Send(ctx, ba)
	if pErr == nil {
		__antithesis_instrumentation__.Notify(127999)

		if err := txn.SetFixedTimestamp(ctx, br.Timestamp); err != nil {
			__antithesis_instrumentation__.Notify(128001)
			return nil, roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(128002)
		}
		__antithesis_instrumentation__.Notify(128000)

		return br, nil
	} else {
		__antithesis_instrumentation__.Notify(128003)
	}
	__antithesis_instrumentation__.Notify(127993)
	if _, ok := pErr.GetDetail().(*roachpb.OpRequiresTxnError); !ok {
		__antithesis_instrumentation__.Notify(128004)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(128005)
	}
	__antithesis_instrumentation__.Notify(127994)

	return nil, roachpb.NewError(unimplemented.NewWithIssue(67554,
		"cross-range bounded staleness reads not yet implemented"))
}

func (txn *Txn) checkNegotiateAndSendPreconditions(
	ctx context.Context, ba roachpb.BatchRequest,
) (err error) {
	__antithesis_instrumentation__.Notify(128006)
	assert := func(b bool, s string) {
		__antithesis_instrumentation__.Notify(128009)
		if !b {
			__antithesis_instrumentation__.Notify(128010)
			err = errors.CombineErrors(err,
				errors.WithContextTags(errors.AssertionFailedf(
					"%s: ba=%s, txn=%s", s, ba.String(), txn.String()), ctx),
			)
		} else {
			__antithesis_instrumentation__.Notify(128011)
		}
	}
	__antithesis_instrumentation__.Notify(128007)
	if cfg := ba.BoundedStaleness; cfg == nil {
		__antithesis_instrumentation__.Notify(128012)
		assert(false, "bounded_staleness configuration must be set")
	} else {
		__antithesis_instrumentation__.Notify(128013)
		assert(!cfg.MinTimestampBound.IsEmpty(), "min_timestamp_bound must be set")
		assert(cfg.MaxTimestampBound.IsEmpty() || func() bool {
			__antithesis_instrumentation__.Notify(128014)
			return cfg.MinTimestampBound.Less(cfg.MaxTimestampBound) == true
		}() == true,
			"max_timestamp_bound, if set, must be greater than min_timestamp_bound")
	}
	__antithesis_instrumentation__.Notify(128008)
	assert(ba.Timestamp.IsEmpty(), "timestamp must not be set")
	assert(ba.Txn == nil, "txn must not be set")
	assert(ba.ReadConsistency == roachpb.CONSISTENT, "read consistency must be set to CONSISTENT")
	assert(ba.IsReadOnly(), "batch must be read-only")
	assert(!ba.IsLocking(), "batch must not be locking")
	assert(txn.typ == RootTxn, "txn must be root")
	assert(!txn.CommitTimestampFixed(), "txn commit timestamp must not be fixed")
	return err
}

func (txn *Txn) applyDeadlineToBoundedStaleness(
	ctx context.Context, bs *roachpb.BoundedStalenessHeader,
) error {
	__antithesis_instrumentation__.Notify(128015)
	d := txn.deadline()
	if d == nil {
		__antithesis_instrumentation__.Notify(128019)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128020)
	}
	__antithesis_instrumentation__.Notify(128016)
	if d.LessEq(bs.MinTimestampBound) {
		__antithesis_instrumentation__.Notify(128021)
		return errors.WithContextTags(errors.AssertionFailedf(
			"transaction deadline %s equal to or below min_timestamp_bound %s",
			*d, bs.MinTimestampBound), ctx)
	} else {
		__antithesis_instrumentation__.Notify(128022)
	}
	__antithesis_instrumentation__.Notify(128017)
	if bs.MaxTimestampBound.IsEmpty() {
		__antithesis_instrumentation__.Notify(128023)
		bs.MaxTimestampBound = *d
	} else {
		__antithesis_instrumentation__.Notify(128024)
		bs.MaxTimestampBound.Backward(*d)
	}
	__antithesis_instrumentation__.Notify(128018)
	return nil
}

func (txn *Txn) GetLeafTxnInputState(ctx context.Context) *roachpb.LeafTxnInputState {
	__antithesis_instrumentation__.Notify(128025)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(128028)
		panic(errors.WithContextTags(errors.AssertionFailedf("GetLeafTxnInputState() called on leaf txn"), ctx))
	} else {
		__antithesis_instrumentation__.Notify(128029)
	}
	__antithesis_instrumentation__.Notify(128026)

	txn.mu.Lock()
	defer txn.mu.Unlock()
	ts, err := txn.mu.sender.GetLeafTxnInputState(ctx, AnyTxnStatus)
	if err != nil {
		__antithesis_instrumentation__.Notify(128030)
		log.Fatalf(ctx, "unexpected error from GetLeafTxnInputState(AnyTxnStatus): %s", err)
	} else {
		__antithesis_instrumentation__.Notify(128031)
	}
	__antithesis_instrumentation__.Notify(128027)
	return ts
}

func (txn *Txn) GetLeafTxnInputStateOrRejectClient(
	ctx context.Context,
) (*roachpb.LeafTxnInputState, error) {
	__antithesis_instrumentation__.Notify(128032)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(128035)
		return nil, errors.WithContextTags(
			errors.AssertionFailedf("GetLeafTxnInputStateOrRejectClient() called on leaf txn"), ctx)
	} else {
		__antithesis_instrumentation__.Notify(128036)
	}
	__antithesis_instrumentation__.Notify(128033)

	txn.mu.Lock()
	defer txn.mu.Unlock()
	tfs, err := txn.mu.sender.GetLeafTxnInputState(ctx, OnlyPending)
	if err != nil {
		__antithesis_instrumentation__.Notify(128037)
		var retryErr *roachpb.TransactionRetryWithProtoRefreshError
		if errors.As(err, &retryErr) {
			__antithesis_instrumentation__.Notify(128039)
			txn.handleRetryableErrLocked(ctx, retryErr)
		} else {
			__antithesis_instrumentation__.Notify(128040)
		}
		__antithesis_instrumentation__.Notify(128038)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(128041)
	}
	__antithesis_instrumentation__.Notify(128034)
	return tfs, nil
}

func (txn *Txn) GetLeafTxnFinalState(ctx context.Context) (*roachpb.LeafTxnFinalState, error) {
	__antithesis_instrumentation__.Notify(128042)
	if txn.typ != LeafTxn {
		__antithesis_instrumentation__.Notify(128045)
		return nil, errors.WithContextTags(
			errors.AssertionFailedf("GetLeafTxnFinalState() called on root txn"), ctx)
	} else {
		__antithesis_instrumentation__.Notify(128046)
	}
	__antithesis_instrumentation__.Notify(128043)

	txn.mu.Lock()
	defer txn.mu.Unlock()
	tfs, err := txn.mu.sender.GetLeafTxnFinalState(ctx, AnyTxnStatus)
	if err != nil {
		__antithesis_instrumentation__.Notify(128047)
		return nil, errors.WithContextTags(
			errors.NewAssertionErrorWithWrappedErrf(err,
				"unexpected error from GetLeafTxnFinalState(AnyTxnStatus)"), ctx)
	} else {
		__antithesis_instrumentation__.Notify(128048)
	}
	__antithesis_instrumentation__.Notify(128044)
	return tfs, nil
}

func (txn *Txn) UpdateRootWithLeafFinalState(
	ctx context.Context, tfs *roachpb.LeafTxnFinalState,
) error {
	__antithesis_instrumentation__.Notify(128049)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(128051)
		return errors.WithContextTags(
			errors.AssertionFailedf("UpdateRootWithLeafFinalState() called on leaf txn"), ctx)
	} else {
		__antithesis_instrumentation__.Notify(128052)
	}
	__antithesis_instrumentation__.Notify(128050)

	txn.mu.Lock()
	defer txn.mu.Unlock()
	txn.mu.sender.UpdateRootWithLeafFinalState(ctx, tfs)
	return nil
}

func (txn *Txn) UpdateStateOnRemoteRetryableErr(ctx context.Context, pErr *roachpb.Error) error {
	__antithesis_instrumentation__.Notify(128053)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(128057)
		return errors.AssertionFailedf("UpdateStateOnRemoteRetryableErr() called on leaf txn")
	} else {
		__antithesis_instrumentation__.Notify(128058)
	}
	__antithesis_instrumentation__.Notify(128054)

	txn.mu.Lock()
	defer txn.mu.Unlock()

	if pErr.TransactionRestart() == roachpb.TransactionRestart_NONE {
		__antithesis_instrumentation__.Notify(128059)
		log.Fatalf(ctx, "unexpected non-retryable error: %s", pErr)
	} else {
		__antithesis_instrumentation__.Notify(128060)
	}
	__antithesis_instrumentation__.Notify(128055)

	origTxnID := pErr.GetTxn().ID
	if origTxnID != txn.mu.ID {
		__antithesis_instrumentation__.Notify(128061)
		return errors.Errorf("retryable error for an older version of txn (current: %s), err: %s",
			txn.mu.ID, pErr)
	} else {
		__antithesis_instrumentation__.Notify(128062)
	}
	__antithesis_instrumentation__.Notify(128056)

	pErr = txn.mu.sender.UpdateStateOnRemoteRetryableErr(ctx, pErr)
	txn.replaceRootSenderIfTxnAbortedLocked(ctx, pErr.GetDetail().(*roachpb.TransactionRetryWithProtoRefreshError), origTxnID)

	return pErr.GoError()
}

func (txn *Txn) replaceRootSenderIfTxnAbortedLocked(
	ctx context.Context, retryErr *roachpb.TransactionRetryWithProtoRefreshError, origTxnID uuid.UUID,
) {
	__antithesis_instrumentation__.Notify(128063)

	newTxn := &retryErr.Transaction

	if txn.mu.ID != origTxnID {
		__antithesis_instrumentation__.Notify(128066)

		log.VEventf(ctx, 2, "retriable error for old incarnation of the transaction")
		return
	} else {
		__antithesis_instrumentation__.Notify(128067)
	}
	__antithesis_instrumentation__.Notify(128064)
	if !retryErr.PrevTxnAborted() {
		__antithesis_instrumentation__.Notify(128068)

		txn.mu.sender.ClearTxnRetryableErr(ctx)
		return
	} else {
		__antithesis_instrumentation__.Notify(128069)
	}
	__antithesis_instrumentation__.Notify(128065)

	txn.recordPreviousTxnIDLocked(txn.mu.ID)
	txn.mu.ID = newTxn.ID

	prevSteppingMode := txn.mu.sender.GetSteppingMode(ctx)
	txn.mu.sender = txn.db.factory.RootTransactionalSender(newTxn, txn.mu.userPriority)
	txn.mu.sender.ConfigureStepping(ctx, prevSteppingMode)
}

func (txn *Txn) recordPreviousTxnIDLocked(prevTxnID uuid.UUID) {
	__antithesis_instrumentation__.Notify(128070)
	if txn.mu.previousIDs == nil {
		__antithesis_instrumentation__.Notify(128072)
		txn.mu.previousIDs = make(map[uuid.UUID]struct{})
	} else {
		__antithesis_instrumentation__.Notify(128073)
	}
	__antithesis_instrumentation__.Notify(128071)
	txn.mu.previousIDs[txn.mu.ID] = struct{}{}
}

func (txn *Txn) SetFixedTimestamp(ctx context.Context, ts hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(128074)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(128077)
		return errors.WithContextTags(errors.AssertionFailedf(
			"SetFixedTimestamp() called on leaf txn"), ctx)
	} else {
		__antithesis_instrumentation__.Notify(128078)
	}
	__antithesis_instrumentation__.Notify(128075)

	if ts.IsEmpty() {
		__antithesis_instrumentation__.Notify(128079)
		return errors.WithContextTags(errors.AssertionFailedf(
			"empty timestamp is invalid for SetFixedTimestamp()"), ctx)
	} else {
		__antithesis_instrumentation__.Notify(128080)
	}
	__antithesis_instrumentation__.Notify(128076)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.SetFixedTimestamp(ctx, ts)
}

func (txn *Txn) GenerateForcedRetryableError(ctx context.Context, msg string) error {
	__antithesis_instrumentation__.Notify(128081)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	now := txn.db.clock.NowAsClockTimestamp()
	txn.mu.sender.ManualRestart(ctx, txn.mu.userPriority, now.ToTimestamp())
	txn.resetDeadlineLocked()
	return txn.mu.sender.PrepareRetryableError(ctx, msg)
}

func (txn *Txn) PrepareRetryableError(ctx context.Context, msg string) error {
	__antithesis_instrumentation__.Notify(128082)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(128084)
		return errors.WithContextTags(
			errors.AssertionFailedf("PrepareRetryableError() called on leaf txn"), ctx)
	} else {
		__antithesis_instrumentation__.Notify(128085)
	}
	__antithesis_instrumentation__.Notify(128083)

	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.PrepareRetryableError(ctx, msg)
}

func (txn *Txn) ManualRestart(ctx context.Context, ts hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(128086)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(128088)
		panic(errors.WithContextTags(
			errors.AssertionFailedf("ManualRestart() called on leaf txn"), ctx))
	} else {
		__antithesis_instrumentation__.Notify(128089)
	}
	__antithesis_instrumentation__.Notify(128087)

	txn.mu.Lock()
	defer txn.mu.Unlock()
	txn.mu.sender.ManualRestart(ctx, txn.mu.userPriority, ts)
}

func (txn *Txn) IsSerializablePushAndRefreshNotPossible() bool {
	__antithesis_instrumentation__.Notify(128090)
	return txn.mu.sender.IsSerializablePushAndRefreshNotPossible()
}

func (txn *Txn) Type() TxnType {
	__antithesis_instrumentation__.Notify(128091)
	return txn.typ
}

func (txn *Txn) TestingCloneTxn() *roachpb.Transaction {
	__antithesis_instrumentation__.Notify(128092)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.TestingCloneTxn()
}

func (txn *Txn) deadline() *hlc.Timestamp {
	__antithesis_instrumentation__.Notify(128093)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.deadline
}

func (txn *Txn) Active() bool {
	__antithesis_instrumentation__.Notify(128094)
	if txn.typ != RootTxn {
		__antithesis_instrumentation__.Notify(128096)
		panic(errors.AssertionFailedf("Active() called on leaf txn"))
	} else {
		__antithesis_instrumentation__.Notify(128097)
	}
	__antithesis_instrumentation__.Notify(128095)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.Active()
}

func (txn *Txn) Step(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(128098)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.Step(ctx)
}

func (txn *Txn) SetReadSeqNum(seq enginepb.TxnSeq) error {
	__antithesis_instrumentation__.Notify(128099)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.SetReadSeqNum(seq)
}

func (txn *Txn) ConfigureStepping(ctx context.Context, mode SteppingMode) (prevMode SteppingMode) {
	__antithesis_instrumentation__.Notify(128100)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.ConfigureStepping(ctx, mode)
}

func (txn *Txn) CreateSavepoint(ctx context.Context) (SavepointToken, error) {
	__antithesis_instrumentation__.Notify(128101)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.CreateSavepoint(ctx)
}

func (txn *Txn) RollbackToSavepoint(ctx context.Context, s SavepointToken) error {
	__antithesis_instrumentation__.Notify(128102)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.RollbackToSavepoint(ctx, s)
}

func (txn *Txn) ReleaseSavepoint(ctx context.Context, s SavepointToken) error {
	__antithesis_instrumentation__.Notify(128103)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.ReleaseSavepoint(ctx, s)
}

func (txn *Txn) ManualRefresh(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(128104)
	txn.mu.Lock()
	sender := txn.mu.sender
	txn.mu.Unlock()
	return sender.ManualRefresh(ctx)
}

func (txn *Txn) DeferCommitWait(ctx context.Context) func(context.Context) error {
	__antithesis_instrumentation__.Notify(128105)
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.mu.sender.DeferCommitWait(ctx)
}

func (txn *Txn) AdmissionHeader() roachpb.AdmissionHeader {
	__antithesis_instrumentation__.Notify(128106)
	h := txn.admissionHeader
	if txn.mu.sender.IsLocking() {
		__antithesis_instrumentation__.Notify(128108)

		h.Priority = int32(admission.LockingPri)
	} else {
		__antithesis_instrumentation__.Notify(128109)
	}
	__antithesis_instrumentation__.Notify(128107)
	return h
}

type OnePCNotAllowedError struct{}

var _ error = OnePCNotAllowedError{}

func (OnePCNotAllowedError) Error() string {
	__antithesis_instrumentation__.Notify(128110)
	return "could not commit in one phase as requested"
}
