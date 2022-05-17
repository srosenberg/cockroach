package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/ui"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	loginPath  = "/login"
	logoutPath = "/logout"

	secretLength = 16

	SessionCookieName = "session"

	DemoLoginPath = "/demologin"
)

type noOIDCConfigured struct{}

func (c *noOIDCConfigured) GetOIDCConf() ui.OIDCUIConf {
	__antithesis_instrumentation__.Notify(189203)
	return ui.OIDCUIConf{
		Enabled: false,
	}
}

type OIDC interface {
	ui.OIDCUI
}

var ConfigureOIDC = func(
	ctx context.Context,
	st *cluster.Settings,
	locality roachpb.Locality,
	handleHTTP func(pattern string, handler http.Handler),
	userLoginFromSSO func(ctx context.Context, username string) (*http.Cookie, error),
	ambientCtx log.AmbientContext,
	cluster uuid.UUID,
) (OIDC, error) {
	__antithesis_instrumentation__.Notify(189204)
	return &noOIDCConfigured{}, nil
}

var webSessionTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"server.web_session_timeout",
	"the duration that a newly created web session will be valid",
	7*24*time.Hour,
	settings.NonNegativeDuration,
).WithPublic()

type authenticationServer struct {
	cfg       *base.Config
	sqlServer *SQLServer
}

func newAuthenticationServer(cfg *base.Config, s *SQLServer) *authenticationServer {
	__antithesis_instrumentation__.Notify(189205)
	return &authenticationServer{
		cfg:       cfg,
		sqlServer: s,
	}
}

func (s *authenticationServer) RegisterService(g *grpc.Server) {
	__antithesis_instrumentation__.Notify(189206)
	serverpb.RegisterLogInServer(g, s)
	serverpb.RegisterLogOutServer(g, s)
}

func (s *authenticationServer) RegisterGateway(
	ctx context.Context, mux *gwruntime.ServeMux, conn *grpc.ClientConn,
) error {
	__antithesis_instrumentation__.Notify(189207)
	if err := serverpb.RegisterLogInHandler(ctx, mux, conn); err != nil {
		__antithesis_instrumentation__.Notify(189209)
		return err
	} else {
		__antithesis_instrumentation__.Notify(189210)
	}
	__antithesis_instrumentation__.Notify(189208)
	return serverpb.RegisterLogOutHandler(ctx, mux, conn)
}

func (s *authenticationServer) UserLogin(
	ctx context.Context, req *serverpb.UserLoginRequest,
) (*serverpb.UserLoginResponse, error) {
	__antithesis_instrumentation__.Notify(189211)
	if req.Username == "" {
		__antithesis_instrumentation__.Notify(189218)
		return nil, status.Errorf(
			codes.Unauthenticated,
			"no username was provided",
		)
	} else {
		__antithesis_instrumentation__.Notify(189219)
	}
	__antithesis_instrumentation__.Notify(189212)

	username, _ := security.MakeSQLUsernameFromUserInput(req.Username, security.UsernameValidation)

	verified, expired, err := s.verifyPasswordDBConsole(ctx, username, req.Password)
	if err != nil {
		__antithesis_instrumentation__.Notify(189220)
		return nil, apiInternalError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189221)
	}
	__antithesis_instrumentation__.Notify(189213)
	if expired {
		__antithesis_instrumentation__.Notify(189222)
		return nil, status.Errorf(
			codes.Unauthenticated,
			"the password for %s has expired",
			username,
		)
	} else {
		__antithesis_instrumentation__.Notify(189223)
	}
	__antithesis_instrumentation__.Notify(189214)
	if !verified {
		__antithesis_instrumentation__.Notify(189224)
		return nil, errWebAuthenticationFailure
	} else {
		__antithesis_instrumentation__.Notify(189225)
	}
	__antithesis_instrumentation__.Notify(189215)

	cookie, err := s.createSessionFor(ctx, username)
	if err != nil {
		__antithesis_instrumentation__.Notify(189226)
		return nil, apiInternalError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189227)
	}
	__antithesis_instrumentation__.Notify(189216)

	if err := grpc.SetHeader(ctx, metadata.Pairs("set-cookie", cookie.String())); err != nil {
		__antithesis_instrumentation__.Notify(189228)
		return nil, apiInternalError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189229)
	}
	__antithesis_instrumentation__.Notify(189217)

	return &serverpb.UserLoginResponse{}, nil
}

func (s *authenticationServer) demoLogin(w http.ResponseWriter, req *http.Request) {
	__antithesis_instrumentation__.Notify(189230)
	ctx := context.Background()
	ctx = logtags.AddTag(ctx, "client", log.SafeOperational(req.RemoteAddr))
	ctx = logtags.AddTag(ctx, "demologin", nil)

	fail := func(err error) {
		__antithesis_instrumentation__.Notify(189239)
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprintf("invalid request: %v", err)))
	}
	__antithesis_instrumentation__.Notify(189231)

	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(189240)
		fail(err)
		return
	} else {
		__antithesis_instrumentation__.Notify(189241)
	}
	__antithesis_instrumentation__.Notify(189232)

	var userInput, password string
	if len(req.Form["username"]) != 1 {
		__antithesis_instrumentation__.Notify(189242)
		fail(errors.New("username not passed right"))
		return
	} else {
		__antithesis_instrumentation__.Notify(189243)
	}
	__antithesis_instrumentation__.Notify(189233)
	if len(req.Form["password"]) != 1 {
		__antithesis_instrumentation__.Notify(189244)
		fail(errors.New("password not passed right"))
		return
	} else {
		__antithesis_instrumentation__.Notify(189245)
	}
	__antithesis_instrumentation__.Notify(189234)
	userInput = req.Form["username"][0]
	password = req.Form["password"][0]

	username, _ := security.MakeSQLUsernameFromUserInput(userInput, security.UsernameValidation)

	verified, expired, err := s.verifyPasswordDBConsole(ctx, username, password)
	if err != nil {
		__antithesis_instrumentation__.Notify(189246)
		fail(err)
		return
	} else {
		__antithesis_instrumentation__.Notify(189247)
	}
	__antithesis_instrumentation__.Notify(189235)
	if expired {
		__antithesis_instrumentation__.Notify(189248)
		fail(errors.New("password expired"))
		return
	} else {
		__antithesis_instrumentation__.Notify(189249)
	}
	__antithesis_instrumentation__.Notify(189236)
	if !verified {
		__antithesis_instrumentation__.Notify(189250)
		fail(errors.New("password invalid"))
		return
	} else {
		__antithesis_instrumentation__.Notify(189251)
	}
	__antithesis_instrumentation__.Notify(189237)

	cookie, err := s.createSessionFor(ctx, username)
	if err != nil {
		__antithesis_instrumentation__.Notify(189252)
		fail(err)
		return
	} else {
		__antithesis_instrumentation__.Notify(189253)
	}
	__antithesis_instrumentation__.Notify(189238)

	w.Header()["Set-Cookie"] = []string{cookie.String()}
	w.Header()["Location"] = []string{"/"}
	w.WriteHeader(302)
	_, _ = w.Write([]byte("you can use the UI now"))
}

var errWebAuthenticationFailure = status.Errorf(
	codes.Unauthenticated,
	"the provided credentials did not match any account on the server",
)

func (s *authenticationServer) UserLoginFromSSO(
	ctx context.Context, reqUsername string,
) (*http.Cookie, error) {
	__antithesis_instrumentation__.Notify(189254)

	username, _ := security.MakeSQLUsernameFromUserInput(reqUsername, security.UsernameValidation)

	exists, _, canLoginDBConsole, _, _, _, err := sql.GetUserSessionInitInfo(
		ctx,
		s.sqlServer.execCfg,
		s.sqlServer.execCfg.InternalExecutor,
		username,
		"",
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(189257)
		return nil, errors.Wrap(err, "failed creating session for username")
	} else {
		__antithesis_instrumentation__.Notify(189258)
	}
	__antithesis_instrumentation__.Notify(189255)
	if !exists || func() bool {
		__antithesis_instrumentation__.Notify(189259)
		return !canLoginDBConsole == true
	}() == true {
		__antithesis_instrumentation__.Notify(189260)
		return nil, errWebAuthenticationFailure
	} else {
		__antithesis_instrumentation__.Notify(189261)
	}
	__antithesis_instrumentation__.Notify(189256)

	return s.createSessionFor(ctx, username)
}

func (s *authenticationServer) createSessionFor(
	ctx context.Context, username security.SQLUsername,
) (*http.Cookie, error) {
	__antithesis_instrumentation__.Notify(189262)

	id, secret, err := s.newAuthSession(ctx, username)
	if err != nil {
		__antithesis_instrumentation__.Notify(189264)
		return nil, apiInternalError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189265)
	}
	__antithesis_instrumentation__.Notify(189263)

	cookieValue := &serverpb.SessionCookie{
		ID:     id,
		Secret: secret,
	}
	return EncodeSessionCookie(cookieValue, !s.cfg.DisableTLSForHTTP)
}

func (s *authenticationServer) UserLogout(
	ctx context.Context, req *serverpb.UserLogoutRequest,
) (*serverpb.UserLogoutResponse, error) {
	__antithesis_instrumentation__.Notify(189266)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		__antithesis_instrumentation__.Notify(189272)
		return nil, apiInternalError(ctx, fmt.Errorf("couldn't get incoming context"))
	} else {
		__antithesis_instrumentation__.Notify(189273)
	}
	__antithesis_instrumentation__.Notify(189267)
	sessionIDs := md.Get(webSessionIDKeyStr)
	if len(sessionIDs) != 1 {
		__antithesis_instrumentation__.Notify(189274)
		return nil, apiInternalError(ctx, fmt.Errorf("couldn't get incoming context"))
	} else {
		__antithesis_instrumentation__.Notify(189275)
	}
	__antithesis_instrumentation__.Notify(189268)

	sessionID, err := strconv.Atoi(sessionIDs[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(189276)
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid session id: %d", sessionID)
	} else {
		__antithesis_instrumentation__.Notify(189277)
	}
	__antithesis_instrumentation__.Notify(189269)

	if n, err := s.sqlServer.internalExecutor.ExecEx(
		ctx,
		"revoke-auth-session",
		nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		`UPDATE system.web_sessions SET "revokedAt" = now() WHERE id = $1`,
		sessionID,
	); err != nil {
		__antithesis_instrumentation__.Notify(189278)
		return nil, apiInternalError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189279)
		if n == 0 {
			__antithesis_instrumentation__.Notify(189280)
			err := status.Errorf(
				codes.InvalidArgument,
				"session with id %d nonexistent", sessionID)
			log.Infof(ctx, "%v", err)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(189281)
		}
	}
	__antithesis_instrumentation__.Notify(189270)

	cookie := makeCookieWithValue("", false)
	cookie.MaxAge = -1

	if err := grpc.SetHeader(ctx, metadata.Pairs("set-cookie", cookie.String())); err != nil {
		__antithesis_instrumentation__.Notify(189282)
		return nil, apiInternalError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(189283)
	}
	__antithesis_instrumentation__.Notify(189271)

	return &serverpb.UserLogoutResponse{}, nil
}

func (s *authenticationServer) verifySession(
	ctx context.Context, cookie *serverpb.SessionCookie,
) (bool, string, error) {
	__antithesis_instrumentation__.Notify(189284)

	const sessionQuery = `
SELECT "hashedSecret", "username", "expiresAt", "revokedAt"
FROM system.web_sessions
WHERE id = $1`

	var (
		hashedSecret []byte
		username     string
		expiresAt    time.Time
		isRevoked    bool
	)

	row, err := s.sqlServer.internalExecutor.QueryRowEx(
		ctx,
		"lookup-auth-session",
		nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		sessionQuery, cookie.ID)
	if row == nil || func() bool {
		__antithesis_instrumentation__.Notify(189290)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(189291)
		return false, "", err
	} else {
		__antithesis_instrumentation__.Notify(189292)
	}
	__antithesis_instrumentation__.Notify(189285)

	if row.Len() != 4 || func() bool {
		__antithesis_instrumentation__.Notify(189293)
		return row[0].ResolvedType().Family() != types.BytesFamily == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(189294)
		return row[1].ResolvedType().Family() != types.StringFamily == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(189295)
		return row[2].ResolvedType().Family() != types.TimestampFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(189296)
		return false, "", errors.Errorf("values returned from auth session lookup do not match expectation")
	} else {
		__antithesis_instrumentation__.Notify(189297)
	}
	__antithesis_instrumentation__.Notify(189286)

	hashedSecret = []byte(*row[0].(*tree.DBytes))
	username = string(*row[1].(*tree.DString))
	expiresAt = row[2].(*tree.DTimestamp).Time
	isRevoked = row[3].ResolvedType().Family() != types.UnknownFamily

	if isRevoked {
		__antithesis_instrumentation__.Notify(189298)
		return false, "", nil
	} else {
		__antithesis_instrumentation__.Notify(189299)
	}
	__antithesis_instrumentation__.Notify(189287)

	if now := s.sqlServer.execCfg.Clock.PhysicalTime(); !now.Before(expiresAt) {
		__antithesis_instrumentation__.Notify(189300)
		return false, "", nil
	} else {
		__antithesis_instrumentation__.Notify(189301)
	}
	__antithesis_instrumentation__.Notify(189288)

	hasher := sha256.New()
	_, _ = hasher.Write(cookie.Secret)
	hashedCookieSecret := hasher.Sum(nil)
	if !bytes.Equal(hashedSecret, hashedCookieSecret) {
		__antithesis_instrumentation__.Notify(189302)
		return false, "", nil
	} else {
		__antithesis_instrumentation__.Notify(189303)
	}
	__antithesis_instrumentation__.Notify(189289)

	return true, username, nil
}

func (s *authenticationServer) verifyPasswordDBConsole(
	ctx context.Context, username security.SQLUsername, password string,
) (valid bool, expired bool, err error) {
	__antithesis_instrumentation__.Notify(189304)
	exists, _, canLoginDBConsole, _, _, pwRetrieveFn, err := sql.GetUserSessionInitInfo(
		ctx,
		s.sqlServer.execCfg,
		s.sqlServer.execCfg.InternalExecutor,
		username,
		"",
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(189310)
		return false, false, err
	} else {
		__antithesis_instrumentation__.Notify(189311)
	}
	__antithesis_instrumentation__.Notify(189305)
	if !exists || func() bool {
		__antithesis_instrumentation__.Notify(189312)
		return !canLoginDBConsole == true
	}() == true {
		__antithesis_instrumentation__.Notify(189313)
		return false, false, nil
	} else {
		__antithesis_instrumentation__.Notify(189314)
	}
	__antithesis_instrumentation__.Notify(189306)
	expired, hashedPassword, err := pwRetrieveFn(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(189315)
		return false, false, err
	} else {
		__antithesis_instrumentation__.Notify(189316)
	}
	__antithesis_instrumentation__.Notify(189307)

	if expired {
		__antithesis_instrumentation__.Notify(189317)
		return false, true, nil
	} else {
		__antithesis_instrumentation__.Notify(189318)
	}
	__antithesis_instrumentation__.Notify(189308)

	ok, err := security.CompareHashAndCleartextPassword(ctx, hashedPassword, password)
	if ok && func() bool {
		__antithesis_instrumentation__.Notify(189319)
		return err == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(189320)

		sql.MaybeUpgradeStoredPasswordHash(ctx,
			s.sqlServer.execCfg,
			username,
			password, hashedPassword)
	} else {
		__antithesis_instrumentation__.Notify(189321)
	}
	__antithesis_instrumentation__.Notify(189309)
	return ok, false, err
}

func CreateAuthSecret() (secret, hashedSecret []byte, err error) {
	__antithesis_instrumentation__.Notify(189322)
	secret = make([]byte, secretLength)
	if _, err := rand.Read(secret); err != nil {
		__antithesis_instrumentation__.Notify(189324)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(189325)
	}
	__antithesis_instrumentation__.Notify(189323)

	hasher := sha256.New()
	_, _ = hasher.Write(secret)
	hashedSecret = hasher.Sum(nil)
	return secret, hashedSecret, nil
}

func (s *authenticationServer) newAuthSession(
	ctx context.Context, username security.SQLUsername,
) (int64, []byte, error) {
	__antithesis_instrumentation__.Notify(189326)
	secret, hashedSecret, err := CreateAuthSecret()
	if err != nil {
		__antithesis_instrumentation__.Notify(189330)
		return 0, nil, err
	} else {
		__antithesis_instrumentation__.Notify(189331)
	}
	__antithesis_instrumentation__.Notify(189327)

	expiration := s.sqlServer.execCfg.Clock.PhysicalTime().Add(webSessionTimeout.Get(&s.sqlServer.execCfg.Settings.SV))

	insertSessionStmt := `
INSERT INTO system.web_sessions ("hashedSecret", username, "expiresAt")
VALUES($1, $2, $3)
RETURNING id
`
	var id int64

	row, err := s.sqlServer.internalExecutor.QueryRowEx(
		ctx,
		"create-auth-session",
		nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		insertSessionStmt,
		hashedSecret,
		username.Normalized(),
		expiration,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(189332)
		return 0, nil, err
	} else {
		__antithesis_instrumentation__.Notify(189333)
	}
	__antithesis_instrumentation__.Notify(189328)
	if row.Len() != 1 || func() bool {
		__antithesis_instrumentation__.Notify(189334)
		return row[0].ResolvedType().Family() != types.IntFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(189335)
		return 0, nil, errors.Errorf(
			"expected create auth session statement to return exactly one integer, returned %v",
			row,
		)
	} else {
		__antithesis_instrumentation__.Notify(189336)
	}
	__antithesis_instrumentation__.Notify(189329)

	id = int64(*row[0].(*tree.DInt))

	return id, secret, nil
}

type authenticationMux struct {
	server *authenticationServer
	inner  http.Handler

	allowAnonymous bool
}

func newAuthenticationMuxAllowAnonymous(
	s *authenticationServer, inner http.Handler,
) *authenticationMux {
	__antithesis_instrumentation__.Notify(189337)
	return &authenticationMux{
		server:         s,
		inner:          inner,
		allowAnonymous: true,
	}
}

func newAuthenticationMux(s *authenticationServer, inner http.Handler) *authenticationMux {
	__antithesis_instrumentation__.Notify(189338)
	return &authenticationMux{
		server:         s,
		inner:          inner,
		allowAnonymous: false,
	}
}

type webSessionUserKey struct{}
type webSessionIDKey struct{}

const webSessionUserKeyStr = "websessionuser"
const webSessionIDKeyStr = "websessionid"

func (am *authenticationMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	__antithesis_instrumentation__.Notify(189339)
	username, cookie, err := am.getSession(w, req)
	if err == nil {
		__antithesis_instrumentation__.Notify(189341)
		ctx := req.Context()
		ctx = context.WithValue(ctx, webSessionUserKey{}, username)
		ctx = context.WithValue(ctx, webSessionIDKey{}, cookie.ID)
		req = req.WithContext(ctx)
	} else {
		__antithesis_instrumentation__.Notify(189342)
		if !am.allowAnonymous {
			__antithesis_instrumentation__.Notify(189343)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(189345)
				log.Infof(req.Context(), "web session error: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(189346)
			}
			__antithesis_instrumentation__.Notify(189344)
			http.Error(w, "a valid authentication cookie is required", http.StatusUnauthorized)
			return
		} else {
			__antithesis_instrumentation__.Notify(189347)
		}
	}
	__antithesis_instrumentation__.Notify(189340)
	am.inner.ServeHTTP(w, req)
}

func EncodeSessionCookie(
	sessionCookie *serverpb.SessionCookie, forHTTPSOnly bool,
) (*http.Cookie, error) {
	__antithesis_instrumentation__.Notify(189348)
	cookieValueBytes, err := protoutil.Marshal(sessionCookie)
	if err != nil {
		__antithesis_instrumentation__.Notify(189350)
		return nil, errors.Wrap(err, "session cookie could not be encoded")
	} else {
		__antithesis_instrumentation__.Notify(189351)
	}
	__antithesis_instrumentation__.Notify(189349)
	value := base64.StdEncoding.EncodeToString(cookieValueBytes)
	return makeCookieWithValue(value, forHTTPSOnly), nil
}

func makeCookieWithValue(value string, forHTTPSOnly bool) *http.Cookie {
	__antithesis_instrumentation__.Notify(189352)
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   forHTTPSOnly,
	}
}

func (am *authenticationMux) getSession(
	w http.ResponseWriter, req *http.Request,
) (string, *serverpb.SessionCookie, error) {
	__antithesis_instrumentation__.Notify(189353)
	ctx := req.Context()

	cookies := req.Cookies()
	found := false
	var cookie *serverpb.SessionCookie
	var err error
	for _, c := range cookies {
		__antithesis_instrumentation__.Notify(189358)
		if c.Name != SessionCookieName {
			__antithesis_instrumentation__.Notify(189361)
			continue
		} else {
			__antithesis_instrumentation__.Notify(189362)
		}
		__antithesis_instrumentation__.Notify(189359)
		found = true
		cookie, err = decodeSessionCookie(c)
		if err != nil {
			__antithesis_instrumentation__.Notify(189363)

			log.Infof(ctx, "found a matching cookie that failed decoding: %v", err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(189364)
		}
		__antithesis_instrumentation__.Notify(189360)
		break
	}
	__antithesis_instrumentation__.Notify(189354)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(189365)
		return !found == true
	}() == true {
		__antithesis_instrumentation__.Notify(189366)
		return "", nil, http.ErrNoCookie
	} else {
		__antithesis_instrumentation__.Notify(189367)
	}
	__antithesis_instrumentation__.Notify(189355)

	valid, username, err := am.server.verifySession(req.Context(), cookie)
	if err != nil {
		__antithesis_instrumentation__.Notify(189368)
		err := apiInternalError(req.Context(), err)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(189369)
	}
	__antithesis_instrumentation__.Notify(189356)
	if !valid {
		__antithesis_instrumentation__.Notify(189370)
		err := errors.New("the provided authentication session could not be validated")
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(189371)
	}
	__antithesis_instrumentation__.Notify(189357)

	return username, cookie, nil
}

func decodeSessionCookie(encodedCookie *http.Cookie) (*serverpb.SessionCookie, error) {
	__antithesis_instrumentation__.Notify(189372)

	cookieBytes, err := base64.StdEncoding.DecodeString(encodedCookie.Value)
	if err != nil {
		__antithesis_instrumentation__.Notify(189375)
		return nil, errors.Wrap(err, "session cookie could not be decoded")
	} else {
		__antithesis_instrumentation__.Notify(189376)
	}
	__antithesis_instrumentation__.Notify(189373)
	var sessionCookieValue serverpb.SessionCookie
	if err := protoutil.Unmarshal(cookieBytes, &sessionCookieValue); err != nil {
		__antithesis_instrumentation__.Notify(189377)
		return nil, errors.Wrap(err, "session cookie could not be unmarshaled")
	} else {
		__antithesis_instrumentation__.Notify(189378)
	}
	__antithesis_instrumentation__.Notify(189374)
	return &sessionCookieValue, nil
}

func authenticationHeaderMatcher(key string) (string, bool) {
	__antithesis_instrumentation__.Notify(189379)

	if key == "set-cookie" {
		__antithesis_instrumentation__.Notify(189381)
		return key, true
	} else {
		__antithesis_instrumentation__.Notify(189382)
	}
	__antithesis_instrumentation__.Notify(189380)

	return fmt.Sprintf("%s%s", gwruntime.MetadataHeaderPrefix, key), true
}

func forwardAuthenticationMetadata(ctx context.Context, _ *http.Request) metadata.MD {
	__antithesis_instrumentation__.Notify(189383)
	md := metadata.MD{}
	if user := ctx.Value(webSessionUserKey{}); user != nil {
		__antithesis_instrumentation__.Notify(189386)
		md.Set(webSessionUserKeyStr, user.(string))
	} else {
		__antithesis_instrumentation__.Notify(189387)
	}
	__antithesis_instrumentation__.Notify(189384)
	if sessionID := ctx.Value(webSessionIDKey{}); sessionID != nil {
		__antithesis_instrumentation__.Notify(189388)
		md.Set(webSessionIDKeyStr, fmt.Sprintf("%v", sessionID))
	} else {
		__antithesis_instrumentation__.Notify(189389)
	}
	__antithesis_instrumentation__.Notify(189385)
	return md
}
