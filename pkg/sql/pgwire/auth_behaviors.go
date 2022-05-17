package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/errors"
)

type AuthBehaviors struct {
	authenticator       Authenticator
	connClose           func()
	replacementIdentity security.SQLUsername
	replacedIdentity    bool
	roleMapper          RoleMapper
}

var _ Authenticator = (*AuthBehaviors)(nil).Authenticate
var _ func() = (*AuthBehaviors)(nil).ConnClose
var _ RoleMapper = (*AuthBehaviors)(nil).MapRole

var _ = (*AuthBehaviors)(nil).SetConnClose
var _ = (*AuthBehaviors)(nil).SetReplacementIdentity

func (b *AuthBehaviors) Authenticate(
	ctx context.Context,
	systemIdentity security.SQLUsername,
	clientConnection bool,
	pwRetrieveFn PasswordRetrievalFn,
) error {
	__antithesis_instrumentation__.Notify(558996)
	if found := b.authenticator; found != nil {
		__antithesis_instrumentation__.Notify(558998)
		return found(ctx, systemIdentity, clientConnection, pwRetrieveFn)
	} else {
		__antithesis_instrumentation__.Notify(558999)
	}
	__antithesis_instrumentation__.Notify(558997)
	return errors.New("no Authenticator provided to AuthBehaviors")
}

func (b *AuthBehaviors) SetAuthenticator(a Authenticator) {
	__antithesis_instrumentation__.Notify(559000)
	b.authenticator = a
}

func (b *AuthBehaviors) ConnClose() {
	__antithesis_instrumentation__.Notify(559001)
	if fn := b.connClose; fn != nil {
		__antithesis_instrumentation__.Notify(559002)
		fn()
	} else {
		__antithesis_instrumentation__.Notify(559003)
	}
}

func (b *AuthBehaviors) SetConnClose(fn func()) {
	__antithesis_instrumentation__.Notify(559004)
	b.connClose = fn
}

func (b *AuthBehaviors) ReplacementIdentity() (_ security.SQLUsername, ok bool) {
	__antithesis_instrumentation__.Notify(559005)
	return b.replacementIdentity, b.replacedIdentity
}

func (b *AuthBehaviors) SetReplacementIdentity(id security.SQLUsername) {
	__antithesis_instrumentation__.Notify(559006)
	b.replacementIdentity = id
	b.replacedIdentity = true
}

func (b *AuthBehaviors) MapRole(
	ctx context.Context, systemIdentity security.SQLUsername,
) ([]security.SQLUsername, error) {
	__antithesis_instrumentation__.Notify(559007)
	if found := b.roleMapper; found != nil {
		__antithesis_instrumentation__.Notify(559009)
		return found(ctx, systemIdentity)
	} else {
		__antithesis_instrumentation__.Notify(559010)
	}
	__antithesis_instrumentation__.Notify(559008)
	return nil, errors.New("no RoleMapper provided to AuthBehaviors")
}

func (b *AuthBehaviors) SetRoleMapper(m RoleMapper) {
	__antithesis_instrumentation__.Notify(559011)
	b.roleMapper = m
}
