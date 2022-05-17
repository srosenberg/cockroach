package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/hba"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/identmap"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

const (
	authOK int32 = 0

	authCleartextPassword int32 = 3

	authReqSASL int32 = 10

	authReqSASLContinue int32 = 11

	authReqSASLFin int32 = 12
)

type authOptions struct {
	insecure bool

	connType hba.ConnType

	connDetails eventpb.CommonConnectionDetails

	auth *hba.Conf

	identMap *identmap.Conf

	ie *sql.InternalExecutor

	testingSkipAuth bool

	testingAuthHook func(ctx context.Context) error
}

func (c *conn) handleAuthentication(
	ctx context.Context, ac AuthConn, authOpt authOptions, execCfg *sql.ExecutorConfig,
) (connClose func(), _ error) {
	__antithesis_instrumentation__.Notify(558877)
	if authOpt.testingSkipAuth {
		__antithesis_instrumentation__.Notify(558889)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(558890)
	}
	__antithesis_instrumentation__.Notify(558878)
	if authOpt.testingAuthHook != nil {
		__antithesis_instrumentation__.Notify(558891)
		return nil, authOpt.testingAuthHook(ctx)
	} else {
		__antithesis_instrumentation__.Notify(558892)
	}
	__antithesis_instrumentation__.Notify(558879)

	tlsState, hbaEntry, authMethod, err := c.findAuthenticationMethod(authOpt)
	if err != nil {
		__antithesis_instrumentation__.Notify(558893)
		ac.LogAuthFailed(ctx, eventpb.AuthFailReason_METHOD_NOT_FOUND, err)
		return nil, c.sendError(ctx, execCfg, pgerror.WithCandidateCode(err, pgcode.InvalidAuthorizationSpecification))
	} else {
		__antithesis_instrumentation__.Notify(558894)
	}
	__antithesis_instrumentation__.Notify(558880)

	ac.SetAuthMethod(hbaEntry.Method.String())
	ac.LogAuthInfof(ctx, "HBA rule: %s", hbaEntry.Input)

	behaviors, err := authMethod(ctx, ac, tlsState, execCfg, hbaEntry, authOpt.identMap)
	connClose = behaviors.ConnClose
	if err != nil {
		__antithesis_instrumentation__.Notify(558895)
		ac.LogAuthFailed(ctx, eventpb.AuthFailReason_UNKNOWN, err)
		return connClose, c.sendError(ctx, execCfg, pgerror.WithCandidateCode(err, pgcode.InvalidAuthorizationSpecification))
	} else {
		__antithesis_instrumentation__.Notify(558896)
	}
	__antithesis_instrumentation__.Notify(558881)

	var systemIdentity security.SQLUsername
	if found, ok := behaviors.ReplacementIdentity(); ok {
		__antithesis_instrumentation__.Notify(558897)
		systemIdentity = found
		ac.SetSystemIdentity(systemIdentity)
	} else {
		__antithesis_instrumentation__.Notify(558898)
		systemIdentity = c.sessionArgs.User
	}
	__antithesis_instrumentation__.Notify(558882)

	if err := c.chooseDbRole(ctx, ac, behaviors.MapRole, systemIdentity); err != nil {
		__antithesis_instrumentation__.Notify(558899)
		log.Warningf(ctx, "unable to map incoming identity %q to any database user: %+v", systemIdentity, err)
		ac.LogAuthFailed(ctx, eventpb.AuthFailReason_USER_NOT_FOUND, err)
		return connClose, c.sendError(ctx, execCfg, pgerror.WithCandidateCode(err, pgcode.InvalidAuthorizationSpecification))
	} else {
		__antithesis_instrumentation__.Notify(558900)
	}
	__antithesis_instrumentation__.Notify(558883)

	dbUser := c.sessionArgs.User

	exists, canLoginSQL, _, isSuperuser, defaultSettings, pwRetrievalFn, err :=
		sql.GetUserSessionInitInfo(
			ctx,
			execCfg,
			authOpt.ie,
			dbUser,
			c.sessionArgs.SessionDefaults["database"],
		)
	if err != nil {
		__antithesis_instrumentation__.Notify(558901)
		log.Warningf(ctx, "user retrieval failed for user=%q: %+v", dbUser, err)
		ac.LogAuthFailed(ctx, eventpb.AuthFailReason_USER_RETRIEVAL_ERROR, err)
		return connClose, c.sendError(ctx, execCfg, pgerror.WithCandidateCode(err, pgcode.InvalidAuthorizationSpecification))
	} else {
		__antithesis_instrumentation__.Notify(558902)
	}
	__antithesis_instrumentation__.Notify(558884)
	c.sessionArgs.IsSuperuser = isSuperuser

	if !exists {
		__antithesis_instrumentation__.Notify(558903)
		ac.LogAuthFailed(ctx, eventpb.AuthFailReason_USER_NOT_FOUND, nil)

		return connClose, c.sendError(ctx, execCfg, pgerror.WithCandidateCode(security.NewErrPasswordUserAuthFailed(dbUser), pgcode.InvalidPassword))
	} else {
		__antithesis_instrumentation__.Notify(558904)
	}
	__antithesis_instrumentation__.Notify(558885)

	if !canLoginSQL {
		__antithesis_instrumentation__.Notify(558905)
		ac.LogAuthFailed(ctx, eventpb.AuthFailReason_LOGIN_DISABLED, nil)
		return connClose, c.sendError(ctx, execCfg, pgerror.Newf(pgcode.InvalidAuthorizationSpecification, "%s does not have login privilege", dbUser))
	} else {
		__antithesis_instrumentation__.Notify(558906)
	}
	__antithesis_instrumentation__.Notify(558886)

	if err := behaviors.Authenticate(ctx, systemIdentity, true, pwRetrievalFn); err != nil {
		__antithesis_instrumentation__.Notify(558907)
		ac.LogAuthFailed(ctx, eventpb.AuthFailReason_CREDENTIALS_INVALID, err)
		if pErr := (*security.PasswordUserAuthError)(nil); errors.As(err, &pErr) {
			__antithesis_instrumentation__.Notify(558909)
			err = pgerror.WithCandidateCode(err, pgcode.InvalidPassword)
		} else {
			__antithesis_instrumentation__.Notify(558910)
			err = pgerror.WithCandidateCode(err, pgcode.InvalidAuthorizationSpecification)
		}
		__antithesis_instrumentation__.Notify(558908)
		return connClose, c.sendError(ctx, execCfg, err)
	} else {
		__antithesis_instrumentation__.Notify(558911)
	}
	__antithesis_instrumentation__.Notify(558887)

	for _, settingEntry := range defaultSettings {
		__antithesis_instrumentation__.Notify(558912)
		for _, setting := range settingEntry.Settings {
			__antithesis_instrumentation__.Notify(558913)
			keyVal := strings.SplitN(setting, "=", 2)
			if len(keyVal) != 2 {
				__antithesis_instrumentation__.Notify(558916)
				log.Ops.Warningf(ctx, "%s has malformed default setting: %q", dbUser, setting)
				continue
			} else {
				__antithesis_instrumentation__.Notify(558917)
			}
			__antithesis_instrumentation__.Notify(558914)
			if err := sql.CheckSessionVariableValueValid(ctx, execCfg.Settings, keyVal[0], keyVal[1]); err != nil {
				__antithesis_instrumentation__.Notify(558918)
				log.Ops.Warningf(ctx, "%s has invalid default setting: %v", dbUser, err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(558919)
			}
			__antithesis_instrumentation__.Notify(558915)
			if _, ok := c.sessionArgs.SessionDefaults[keyVal[0]]; !ok {
				__antithesis_instrumentation__.Notify(558920)
				c.sessionArgs.SessionDefaults[keyVal[0]] = keyVal[1]
			} else {
				__antithesis_instrumentation__.Notify(558921)
			}
		}
	}
	__antithesis_instrumentation__.Notify(558888)

	ac.LogAuthOK(ctx)
	return connClose, nil
}

func (c *conn) authOKMessage() error {
	__antithesis_instrumentation__.Notify(558922)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgAuth)
	c.msgBuilder.putInt32(authOK)
	return c.msgBuilder.finishMsg(c.conn)
}

func (c *conn) chooseDbRole(
	ctx context.Context, ac AuthConn, mapper RoleMapper, systemIdentity security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(558923)
	if mapped, err := mapper(ctx, systemIdentity); err != nil {
		__antithesis_instrumentation__.Notify(558925)
		return err
	} else {
		__antithesis_instrumentation__.Notify(558926)
		if len(mapped) == 0 {
			__antithesis_instrumentation__.Notify(558927)
			return errors.Newf("system identity %q did not map to a database role", systemIdentity.Normalized())
		} else {
			__antithesis_instrumentation__.Notify(558928)
			c.sessionArgs.User = mapped[0]
			ac.SetDbUser(mapped[0])
		}
	}
	__antithesis_instrumentation__.Notify(558924)
	return nil
}

func (c *conn) findAuthenticationMethod(
	authOpt authOptions,
) (tlsState tls.ConnectionState, hbaEntry *hba.Entry, methodFn AuthMethod, err error) {
	__antithesis_instrumentation__.Notify(558929)
	if authOpt.insecure {
		__antithesis_instrumentation__.Notify(558935)

		methodFn = authTrust
		hbaEntry = &insecureEntry
		return
	} else {
		__antithesis_instrumentation__.Notify(558936)
	}
	__antithesis_instrumentation__.Notify(558930)
	if c.sessionArgs.SessionRevivalToken != nil {
		__antithesis_instrumentation__.Notify(558937)
		methodFn = authSessionRevivalToken(c.sessionArgs.SessionRevivalToken)
		c.sessionArgs.SessionRevivalToken = nil
		hbaEntry = &sessionRevivalEntry
		return
	} else {
		__antithesis_instrumentation__.Notify(558938)
	}
	__antithesis_instrumentation__.Notify(558931)

	var mi methodInfo
	mi, hbaEntry, err = c.lookupAuthenticationMethodUsingRules(authOpt.connType, authOpt.auth)
	if err != nil {
		__antithesis_instrumentation__.Notify(558939)
		return
	} else {
		__antithesis_instrumentation__.Notify(558940)
	}
	__antithesis_instrumentation__.Notify(558932)
	methodFn = mi.fn

	if authOpt.connType&mi.validConnTypes == 0 {
		__antithesis_instrumentation__.Notify(558941)
		err = errors.Newf("method %q required for this user, but unusable over this connection type",
			hbaEntry.Method.Value)
		return
	} else {
		__antithesis_instrumentation__.Notify(558942)
	}
	__antithesis_instrumentation__.Notify(558933)

	if authOpt.connType == hba.ConnHostSSL {
		__antithesis_instrumentation__.Notify(558943)
		tlsConn, ok := c.conn.(*readTimeoutConn).Conn.(*tls.Conn)
		if !ok {
			__antithesis_instrumentation__.Notify(558945)
			err = errors.AssertionFailedf("server reports hostssl conn without TLS state")
			return
		} else {
			__antithesis_instrumentation__.Notify(558946)
		}
		__antithesis_instrumentation__.Notify(558944)
		tlsState = tlsConn.ConnectionState()
	} else {
		__antithesis_instrumentation__.Notify(558947)
	}
	__antithesis_instrumentation__.Notify(558934)

	return
}

func (c *conn) lookupAuthenticationMethodUsingRules(
	connType hba.ConnType, auth *hba.Conf,
) (mi methodInfo, entry *hba.Entry, err error) {
	__antithesis_instrumentation__.Notify(558948)
	var ip net.IP
	if connType != hba.ConnLocal {
		__antithesis_instrumentation__.Notify(558951)

		tcpAddr, ok := c.sessionArgs.RemoteAddr.(*net.TCPAddr)
		if !ok {
			__antithesis_instrumentation__.Notify(558953)
			err = errors.AssertionFailedf("client address type %T unsupported", c.sessionArgs.RemoteAddr)
			return
		} else {
			__antithesis_instrumentation__.Notify(558954)
		}
		__antithesis_instrumentation__.Notify(558952)
		ip = tcpAddr.IP
	} else {
		__antithesis_instrumentation__.Notify(558955)
	}
	__antithesis_instrumentation__.Notify(558949)

	for i := range auth.Entries {
		__antithesis_instrumentation__.Notify(558956)
		entry = &auth.Entries[i]
		var connMatch bool
		connMatch, err = entry.ConnMatches(connType, ip)
		if err != nil {
			__antithesis_instrumentation__.Notify(558960)

			return
		} else {
			__antithesis_instrumentation__.Notify(558961)
		}
		__antithesis_instrumentation__.Notify(558957)
		if !connMatch {
			__antithesis_instrumentation__.Notify(558962)

			continue
		} else {
			__antithesis_instrumentation__.Notify(558963)
		}
		__antithesis_instrumentation__.Notify(558958)
		if !entry.UserMatches(c.sessionArgs.User) {
			__antithesis_instrumentation__.Notify(558964)

			continue
		} else {
			__antithesis_instrumentation__.Notify(558965)
		}
		__antithesis_instrumentation__.Notify(558959)
		return entry.MethodFn.(methodInfo), entry, nil
	}
	__antithesis_instrumentation__.Notify(558950)

	err = errors.Errorf("no %s entry for host %q, user %q", serverHBAConfSetting, ip, c.sessionArgs.User)
	return
}

type authenticatorIO interface {
	sendPwdData(data []byte) error

	noMorePwdData()

	authResult() error
}

type AuthConn interface {
	SendAuthRequest(authType int32, data []byte) error

	GetPwdData() ([]byte, error)

	AuthOK(context.Context)

	AuthFail(err error)

	SetAuthMethod(method string)

	SetDbUser(dbUser security.SQLUsername)

	SetSystemIdentity(systemIdentity security.SQLUsername)

	LogAuthInfof(ctx context.Context, format string, args ...interface{})

	LogAuthFailed(ctx context.Context, reason eventpb.AuthFailReason, err error)

	LogAuthOK(ctx context.Context)
}

type authPipe struct {
	c   *conn
	log bool

	connDetails eventpb.CommonConnectionDetails
	authDetails eventpb.CommonSessionDetails
	authMethod  string

	ch chan []byte

	writerDone chan struct{}
	readerDone chan authRes
}

type authRes struct {
	err error
}

func newAuthPipe(
	c *conn, logAuthn bool, authOpt authOptions, systemIdentity security.SQLUsername,
) *authPipe {
	__antithesis_instrumentation__.Notify(558966)
	ap := &authPipe{
		c:           c,
		log:         logAuthn,
		connDetails: authOpt.connDetails,
		authDetails: eventpb.CommonSessionDetails{
			SystemIdentity: systemIdentity.Normalized(),
			Transport:      authOpt.connType.String(),
		},
		ch:         make(chan []byte),
		writerDone: make(chan struct{}),
		readerDone: make(chan authRes, 1),
	}
	return ap
}

var _ authenticatorIO = &authPipe{}
var _ AuthConn = &authPipe{}

func (p *authPipe) sendPwdData(data []byte) error {
	__antithesis_instrumentation__.Notify(558967)
	select {
	case p.ch <- data:
		__antithesis_instrumentation__.Notify(558968)
		return nil
	case <-p.readerDone:
		__antithesis_instrumentation__.Notify(558969)
		return pgwirebase.NewProtocolViolationErrorf("unexpected auth data")
	}
}

func (p *authPipe) noMorePwdData() {
	__antithesis_instrumentation__.Notify(558970)
	if p.writerDone == nil {
		__antithesis_instrumentation__.Notify(558972)
		return
	} else {
		__antithesis_instrumentation__.Notify(558973)
	}
	__antithesis_instrumentation__.Notify(558971)

	close(p.writerDone)
	p.writerDone = nil
}

func (p *authPipe) GetPwdData() ([]byte, error) {
	__antithesis_instrumentation__.Notify(558974)
	select {
	case data := <-p.ch:
		__antithesis_instrumentation__.Notify(558975)
		return data, nil
	case <-p.writerDone:
		__antithesis_instrumentation__.Notify(558976)
		return nil, pgwirebase.NewProtocolViolationErrorf("client didn't send required auth data")
	}
}

func (p *authPipe) AuthOK(ctx context.Context) {
	__antithesis_instrumentation__.Notify(558977)
	p.readerDone <- authRes{err: nil}
}

func (p *authPipe) AuthFail(err error) {
	__antithesis_instrumentation__.Notify(558978)
	p.readerDone <- authRes{err: err}
}

func (p *authPipe) SetAuthMethod(method string) {
	__antithesis_instrumentation__.Notify(558979)
	p.authMethod = method
}

func (p *authPipe) SetDbUser(dbUser security.SQLUsername) {
	__antithesis_instrumentation__.Notify(558980)
	p.authDetails.User = dbUser.Normalized()
}

func (p *authPipe) SetSystemIdentity(systemIdentity security.SQLUsername) {
	__antithesis_instrumentation__.Notify(558981)
	p.authDetails.SystemIdentity = systemIdentity.Normalized()
}

func (p *authPipe) LogAuthOK(ctx context.Context) {
	__antithesis_instrumentation__.Notify(558982)
	if p.log {
		__antithesis_instrumentation__.Notify(558983)
		ev := &eventpb.ClientAuthenticationOk{
			CommonConnectionDetails: p.connDetails,
			CommonSessionDetails:    p.authDetails,
			Method:                  p.authMethod,
		}
		log.StructuredEvent(ctx, ev)
	} else {
		__antithesis_instrumentation__.Notify(558984)
	}
}

func (p *authPipe) LogAuthInfof(ctx context.Context, format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(558985)
	if p.log {
		__antithesis_instrumentation__.Notify(558986)
		ev := &eventpb.ClientAuthenticationInfo{
			CommonConnectionDetails: p.connDetails,
			CommonSessionDetails:    p.authDetails,
			Info:                    fmt.Sprintf(format, args...),
			Method:                  p.authMethod,
		}
		log.StructuredEvent(ctx, ev)
	} else {
		__antithesis_instrumentation__.Notify(558987)
	}
}

func (p *authPipe) LogAuthFailed(
	ctx context.Context, reason eventpb.AuthFailReason, detailedErr error,
) {
	__antithesis_instrumentation__.Notify(558988)
	if p.log {
		__antithesis_instrumentation__.Notify(558989)
		var errStr string
		if detailedErr != nil {
			__antithesis_instrumentation__.Notify(558991)
			errStr = detailedErr.Error()
		} else {
			__antithesis_instrumentation__.Notify(558992)
		}
		__antithesis_instrumentation__.Notify(558990)
		ev := &eventpb.ClientAuthenticationFailed{
			CommonConnectionDetails: p.connDetails,
			CommonSessionDetails:    p.authDetails,
			Reason:                  reason,
			Detail:                  errStr,
			Method:                  p.authMethod,
		}
		log.StructuredEvent(ctx, ev)
	} else {
		__antithesis_instrumentation__.Notify(558993)
	}
}

func (p *authPipe) authResult() error {
	__antithesis_instrumentation__.Notify(558994)
	p.noMorePwdData()
	res := <-p.readerDone
	return res.err
}

func (p *authPipe) SendAuthRequest(authType int32, data []byte) error {
	__antithesis_instrumentation__.Notify(558995)
	c := p.c
	c.msgBuilder.initMsg(pgwirebase.ServerMsgAuth)
	c.msgBuilder.putInt32(authType)
	c.msgBuilder.write(data)
	return c.msgBuilder.finishMsg(c.conn)
}
