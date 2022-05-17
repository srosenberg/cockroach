// Package leasemanager provides functionality for acquiring and managing leases
// via the kv api for use during startupmigrations.
package leasemanager

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

const DefaultLeaseDuration = 1 * time.Minute

type LeaseNotAvailableError struct {
	key        roachpb.Key
	expiration hlc.Timestamp
}

func (e *LeaseNotAvailableError) Error() string {
	__antithesis_instrumentation__.Notify(633088)
	return fmt.Sprintf("lease %q is not available until at least %s", e.key, e.expiration)
}

type LeaseManager struct {
	db            *kv.DB
	clock         *hlc.Clock
	clientID      string
	leaseDuration time.Duration
}

type Lease struct {
	key roachpb.Key
	val struct {
		sem      chan struct{}
		lease    *LeaseVal
		leaseRaw []byte
	}
}

type Options struct {
	ClientID      string
	LeaseDuration time.Duration
}

func New(db *kv.DB, clock *hlc.Clock, options Options) *LeaseManager {
	__antithesis_instrumentation__.Notify(633089)
	if options.ClientID == "" {
		__antithesis_instrumentation__.Notify(633092)
		options.ClientID = uuid.MakeV4().String()
	} else {
		__antithesis_instrumentation__.Notify(633093)
	}
	__antithesis_instrumentation__.Notify(633090)
	if options.LeaseDuration <= 0 {
		__antithesis_instrumentation__.Notify(633094)
		options.LeaseDuration = DefaultLeaseDuration
	} else {
		__antithesis_instrumentation__.Notify(633095)
	}
	__antithesis_instrumentation__.Notify(633091)
	return &LeaseManager{
		db:            db,
		clock:         clock,
		clientID:      options.ClientID,
		leaseDuration: options.LeaseDuration,
	}
}

func (m *LeaseManager) AcquireLease(ctx context.Context, key roachpb.Key) (*Lease, error) {
	__antithesis_instrumentation__.Notify(633096)
	lease := &Lease{
		key: key,
	}
	lease.val.sem = make(chan struct{}, 1)
	if err := m.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(633098)
		var val LeaseVal
		err := txn.GetProto(ctx, key, &val)
		if err != nil {
			__antithesis_instrumentation__.Notify(633103)
			return err
		} else {
			__antithesis_instrumentation__.Notify(633104)
		}
		__antithesis_instrumentation__.Notify(633099)
		if !m.leaseAvailable(&val) {
			__antithesis_instrumentation__.Notify(633105)
			return &LeaseNotAvailableError{key: key, expiration: val.Expiration}
		} else {
			__antithesis_instrumentation__.Notify(633106)
		}
		__antithesis_instrumentation__.Notify(633100)
		lease.val.lease = &LeaseVal{
			Owner:      m.clientID,
			Expiration: m.clock.Now().Add(m.leaseDuration.Nanoseconds(), 0),
		}
		var leaseRaw roachpb.Value
		if err := leaseRaw.SetProto(lease.val.lease); err != nil {
			__antithesis_instrumentation__.Notify(633107)
			return err
		} else {
			__antithesis_instrumentation__.Notify(633108)
		}
		__antithesis_instrumentation__.Notify(633101)
		if err := txn.Put(ctx, key, &leaseRaw); err != nil {
			__antithesis_instrumentation__.Notify(633109)
			return err
		} else {
			__antithesis_instrumentation__.Notify(633110)
		}
		__antithesis_instrumentation__.Notify(633102)
		lease.val.leaseRaw = leaseRaw.TagAndDataBytes()
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(633111)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(633112)
	}
	__antithesis_instrumentation__.Notify(633097)
	return lease, nil
}

func (m *LeaseManager) leaseAvailable(val *LeaseVal) bool {
	__antithesis_instrumentation__.Notify(633113)
	return val.Owner == m.clientID || func() bool {
		__antithesis_instrumentation__.Notify(633114)
		return m.timeRemaining(val) <= 0 == true
	}() == true
}

func (m *LeaseManager) TimeRemaining(l *Lease) time.Duration {
	__antithesis_instrumentation__.Notify(633115)
	l.val.sem <- struct{}{}
	defer func() { __antithesis_instrumentation__.Notify(633117); <-l.val.sem }()
	__antithesis_instrumentation__.Notify(633116)
	return m.timeRemaining(l.val.lease)
}

func (m *LeaseManager) timeRemaining(val *LeaseVal) time.Duration {
	__antithesis_instrumentation__.Notify(633118)
	maxOffset := m.clock.MaxOffset()
	return val.Expiration.GoTime().Sub(m.clock.Now().GoTime()) - maxOffset
}

func (m *LeaseManager) ExtendLease(ctx context.Context, l *Lease) error {
	__antithesis_instrumentation__.Notify(633119)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(633125)
		return ctx.Err()
	case l.val.sem <- struct{}{}:
		__antithesis_instrumentation__.Notify(633126)
	}
	__antithesis_instrumentation__.Notify(633120)
	defer func() { __antithesis_instrumentation__.Notify(633127); <-l.val.sem }()
	__antithesis_instrumentation__.Notify(633121)

	if m.timeRemaining(l.val.lease) < 0 {
		__antithesis_instrumentation__.Notify(633128)
		return errors.Errorf("can't extend lease that expired at time %s", l.val.lease.Expiration)
	} else {
		__antithesis_instrumentation__.Notify(633129)
	}
	__antithesis_instrumentation__.Notify(633122)

	newVal := &LeaseVal{
		Owner:      m.clientID,
		Expiration: m.clock.Now().Add(m.leaseDuration.Nanoseconds(), 0),
	}
	var newRaw roachpb.Value
	if err := newRaw.SetProto(newVal); err != nil {
		__antithesis_instrumentation__.Notify(633130)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633131)
	}
	__antithesis_instrumentation__.Notify(633123)
	if err := m.db.CPut(ctx, l.key, &newRaw, l.val.leaseRaw); err != nil {
		__antithesis_instrumentation__.Notify(633132)
		if errors.HasType(err, (*roachpb.ConditionFailedError)(nil)) {
			__antithesis_instrumentation__.Notify(633134)

			l.val.lease.Expiration = hlc.Timestamp{}
			return errors.Wrapf(err, "local lease state %v out of sync with DB state", l.val.lease)
		} else {
			__antithesis_instrumentation__.Notify(633135)
		}
		__antithesis_instrumentation__.Notify(633133)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633136)
	}
	__antithesis_instrumentation__.Notify(633124)
	l.val.lease = newVal
	l.val.leaseRaw = newRaw.TagAndDataBytes()
	return nil
}

func (m *LeaseManager) ReleaseLease(ctx context.Context, l *Lease) error {
	__antithesis_instrumentation__.Notify(633137)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(633140)
		return ctx.Err()
	case l.val.sem <- struct{}{}:
		__antithesis_instrumentation__.Notify(633141)
	}
	__antithesis_instrumentation__.Notify(633138)
	defer func() { __antithesis_instrumentation__.Notify(633142); <-l.val.sem }()
	__antithesis_instrumentation__.Notify(633139)

	return m.db.CPut(ctx, l.key, nil, l.val.leaseRaw)
}
