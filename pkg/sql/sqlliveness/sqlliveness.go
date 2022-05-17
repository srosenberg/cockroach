// Package sqlliveness provides interfaces to associate resources at the SQL
// level with tenant SQL processes.
//
// For more info see:
// https://github.com/cockroachdb/cockroach/blob/master/docs/RFCS/20200615_sql_liveness.md
package sqlliveness

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/hex"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/errors"
)

type SessionID string

type Provider interface {
	Start(ctx context.Context)
	Metrics() metric.Struct
	Liveness

	CachedReader() Reader
}

type Liveness interface {
	Reader
	Instance
}

func (s SessionID) String() string {
	__antithesis_instrumentation__.Notify(624409)
	return hex.EncodeToString(encoding.UnsafeConvertStringToBytes(string(s)))
}

func (s SessionID) SafeValue() { __antithesis_instrumentation__.Notify(624410) }

func (s SessionID) UnsafeBytes() []byte {
	__antithesis_instrumentation__.Notify(624411)
	return encoding.UnsafeConvertStringToBytes(string(s))
}

type Instance interface {
	Session(context.Context) (Session, error)
}

type Session interface {
	ID() SessionID

	Start() hlc.Timestamp

	Expiration() hlc.Timestamp

	RegisterCallbackForSessionExpiry(func(ctx context.Context))
}

type Reader interface {
	IsAlive(context.Context, SessionID) (alive bool, err error)
}

type TestingKnobs struct {
	SessionOverride func(ctx context.Context) (Session, error)
}

var _ base.ModuleTestingKnobs = &TestingKnobs{}

func (*TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(624412) }

var NotStartedError = errors.Errorf("sqlliveness subsystem has not yet been started")
