package slstorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type FakeStorage struct {
	mu struct {
		syncutil.Mutex
		sessions map[sqlliveness.SessionID]hlc.Timestamp
	}
}

func NewFakeStorage() *FakeStorage {
	__antithesis_instrumentation__.Notify(624398)
	fs := &FakeStorage{}
	fs.mu.sessions = make(map[sqlliveness.SessionID]hlc.Timestamp)
	return fs
}

func (s *FakeStorage) IsAlive(
	_ context.Context, sid sqlliveness.SessionID,
) (alive bool, err error) {
	__antithesis_instrumentation__.Notify(624399)
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.mu.sessions[sid]
	return ok, nil
}

func (s *FakeStorage) Insert(
	_ context.Context, sid sqlliveness.SessionID, expiration hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(624400)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mu.sessions[sid]; ok {
		__antithesis_instrumentation__.Notify(624402)
		return errors.Errorf("session %s already exists", sid)
	} else {
		__antithesis_instrumentation__.Notify(624403)
	}
	__antithesis_instrumentation__.Notify(624401)
	s.mu.sessions[sid] = expiration
	return nil
}

func (s *FakeStorage) Update(
	_ context.Context, sid sqlliveness.SessionID, expiration hlc.Timestamp,
) (bool, error) {
	__antithesis_instrumentation__.Notify(624404)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mu.sessions[sid]; !ok {
		__antithesis_instrumentation__.Notify(624406)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(624407)
	}
	__antithesis_instrumentation__.Notify(624405)
	s.mu.sessions[sid] = expiration
	return true, nil
}

func (s *FakeStorage) Delete(_ context.Context, sid sqlliveness.SessionID) error {
	__antithesis_instrumentation__.Notify(624408)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.mu.sessions, sid)
	return nil
}
