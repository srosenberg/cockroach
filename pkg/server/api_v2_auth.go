package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authenticationV2Server struct {
	ctx        context.Context
	sqlServer  *SQLServer
	authServer *authenticationServer
	mux        *http.ServeMux
	basePath   string
}

func newAuthenticationV2Server(
	ctx context.Context, s *Server, basePath string,
) *authenticationV2Server {
	__antithesis_instrumentation__.Notify(188925)
	simpleMux := http.NewServeMux()

	authServer := &authenticationV2Server{
		sqlServer:  s.sqlServer,
		authServer: newAuthenticationServer(s.cfg.Config, s.sqlServer),
		mux:        simpleMux,
		ctx:        ctx,
		basePath:   basePath,
	}

	authServer.registerRoutes()
	return authServer
}

func (a *authenticationV2Server) registerRoutes() {
	__antithesis_instrumentation__.Notify(188926)
	a.bindEndpoint("login/", a.login)
	a.bindEndpoint("logout/", a.logout)
}

func (a *authenticationV2Server) bindEndpoint(endpoint string, handler http.HandlerFunc) {
	__antithesis_instrumentation__.Notify(188927)
	a.mux.HandleFunc(a.basePath+endpoint, handler)
}

func (a *authenticationV2Server) createSessionFor(
	ctx context.Context, username security.SQLUsername,
) (string, error) {
	__antithesis_instrumentation__.Notify(188928)

	id, secret, err := a.authServer.newAuthSession(ctx, username)
	if err != nil {
		__antithesis_instrumentation__.Notify(188931)
		return "", apiInternalError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188932)
	}
	__antithesis_instrumentation__.Notify(188929)

	cookieValue := &serverpb.SessionCookie{
		ID:     id,
		Secret: secret,
	}
	cookieValueBytes, err := protoutil.Marshal(cookieValue)
	if err != nil {
		__antithesis_instrumentation__.Notify(188933)
		return "", errors.Wrap(err, "session cookie could not be encoded")
	} else {
		__antithesis_instrumentation__.Notify(188934)
	}
	__antithesis_instrumentation__.Notify(188930)
	value := base64.StdEncoding.EncodeToString(cookieValueBytes)
	return value, nil
}

type loginResponse struct {
	Session string `json:"session"`
}

func (a *authenticationV2Server) login(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(188935)
	if r.Method != "POST" {
		__antithesis_instrumentation__.Notify(188943)
		http.Error(w, "not found", http.StatusNotFound)
	} else {
		__antithesis_instrumentation__.Notify(188944)
	}
	__antithesis_instrumentation__.Notify(188936)
	if err := r.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(188945)
		apiV2InternalError(r.Context(), err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(188946)
	}
	__antithesis_instrumentation__.Notify(188937)
	if r.Form.Get("username") == "" {
		__antithesis_instrumentation__.Notify(188947)
		http.Error(w, "username not specified", http.StatusBadRequest)
		return
	} else {
		__antithesis_instrumentation__.Notify(188948)
	}
	__antithesis_instrumentation__.Notify(188938)

	username, _ := security.MakeSQLUsernameFromUserInput(r.Form.Get("username"), security.UsernameValidation)

	verified, expired, err := a.authServer.verifyPasswordDBConsole(a.ctx, username, r.Form.Get("password"))
	if err != nil {
		__antithesis_instrumentation__.Notify(188949)
		apiV2InternalError(r.Context(), err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(188950)
	}
	__antithesis_instrumentation__.Notify(188939)
	if expired {
		__antithesis_instrumentation__.Notify(188951)
		http.Error(w, "the password has expired", http.StatusUnauthorized)
		return
	} else {
		__antithesis_instrumentation__.Notify(188952)
	}
	__antithesis_instrumentation__.Notify(188940)
	if !verified {
		__antithesis_instrumentation__.Notify(188953)
		http.Error(w, "the provided credentials did not match any account on the server", http.StatusUnauthorized)
		return
	} else {
		__antithesis_instrumentation__.Notify(188954)
	}
	__antithesis_instrumentation__.Notify(188941)

	session, err := a.createSessionFor(a.ctx, username)
	if err != nil {
		__antithesis_instrumentation__.Notify(188955)
		apiV2InternalError(r.Context(), err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(188956)
	}
	__antithesis_instrumentation__.Notify(188942)

	writeJSONResponse(r.Context(), w, http.StatusOK, &loginResponse{Session: session})
}

type logoutResponse struct {
	LoggedOut bool `json:"logged_out"`
}

func (a *authenticationV2Server) logout(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(188957)
	if r.Method != "POST" {
		__antithesis_instrumentation__.Notify(188963)
		http.Error(w, "not found", http.StatusNotFound)
	} else {
		__antithesis_instrumentation__.Notify(188964)
	}
	__antithesis_instrumentation__.Notify(188958)
	session := r.Header.Get(apiV2AuthHeader)
	if session == "" {
		__antithesis_instrumentation__.Notify(188965)
		http.Error(w, "invalid or unspecified session", http.StatusBadRequest)
		return
	} else {
		__antithesis_instrumentation__.Notify(188966)
	}
	__antithesis_instrumentation__.Notify(188959)
	var sessionCookie serverpb.SessionCookie
	decoded, err := base64.StdEncoding.DecodeString(session)
	if err != nil {
		__antithesis_instrumentation__.Notify(188967)
		apiV2InternalError(r.Context(), err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(188968)
	}
	__antithesis_instrumentation__.Notify(188960)
	if err := protoutil.Unmarshal(decoded, &sessionCookie); err != nil {
		__antithesis_instrumentation__.Notify(188969)
		apiV2InternalError(r.Context(), err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(188970)
	}
	__antithesis_instrumentation__.Notify(188961)

	if n, err := a.sqlServer.internalExecutor.ExecEx(
		a.ctx,
		"revoke-auth-session",
		nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		`UPDATE system.web_sessions SET "revokedAt" = now() WHERE id = $1`,
		sessionCookie.ID,
	); err != nil {
		__antithesis_instrumentation__.Notify(188971)
		apiV2InternalError(r.Context(), err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(188972)
		if n == 0 {
			__antithesis_instrumentation__.Notify(188973)
			err := status.Errorf(
				codes.InvalidArgument,
				"session with id %d nonexistent", sessionCookie.ID)
			log.Infof(a.ctx, "%v", err)
			http.Error(w, "invalid session", http.StatusBadRequest)
			return
		} else {
			__antithesis_instrumentation__.Notify(188974)
		}
	}
	__antithesis_instrumentation__.Notify(188962)

	writeJSONResponse(r.Context(), w, http.StatusOK, &logoutResponse{LoggedOut: true})
}

func (a *authenticationV2Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(188975)
	a.mux.ServeHTTP(w, r)
}

type authenticationV2Mux struct {
	s     *authenticationV2Server
	inner http.Handler
}

func newAuthenticationV2Mux(s *authenticationV2Server, inner http.Handler) *authenticationV2Mux {
	__antithesis_instrumentation__.Notify(188976)
	return &authenticationV2Mux{
		s:     s,
		inner: inner,
	}
}

func (a *authenticationV2Mux) getSession(
	w http.ResponseWriter, req *http.Request,
) (string, *serverpb.SessionCookie, error) {
	__antithesis_instrumentation__.Notify(188977)

	rawSession := req.Header.Get(apiV2AuthHeader)
	if len(rawSession) == 0 {
		__antithesis_instrumentation__.Notify(188983)
		err := errors.New("invalid session header")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(188984)
	}
	__antithesis_instrumentation__.Notify(188978)
	sessionCookie := &serverpb.SessionCookie{}
	decoded, err := base64.StdEncoding.DecodeString(rawSession)
	if err != nil {
		__antithesis_instrumentation__.Notify(188985)
		err := errors.New("invalid session header")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(188986)
	}
	__antithesis_instrumentation__.Notify(188979)
	if err := protoutil.Unmarshal(decoded, sessionCookie); err != nil {
		__antithesis_instrumentation__.Notify(188987)
		err := errors.New("invalid session header")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(188988)
	}
	__antithesis_instrumentation__.Notify(188980)

	valid, username, err := a.s.authServer.verifySession(req.Context(), sessionCookie)
	if err != nil {
		__antithesis_instrumentation__.Notify(188989)
		apiV2InternalError(req.Context(), err, w)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(188990)
	}
	__antithesis_instrumentation__.Notify(188981)
	if !valid {
		__antithesis_instrumentation__.Notify(188991)
		err := errors.New("the provided authentication session could not be validated")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(188992)
	}
	__antithesis_instrumentation__.Notify(188982)

	return username, sessionCookie, nil
}

func (a *authenticationV2Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	__antithesis_instrumentation__.Notify(188993)
	username, cookie, err := a.getSession(w, req)
	if err == nil {
		__antithesis_instrumentation__.Notify(188995)

		ctx := req.Context()
		ctx = context.WithValue(ctx, webSessionUserKey{}, username)
		ctx = context.WithValue(ctx, webSessionIDKey{}, cookie.ID)
		req = req.WithContext(ctx)
	} else {
		__antithesis_instrumentation__.Notify(188996)

		return
	}
	__antithesis_instrumentation__.Notify(188994)
	a.inner.ServeHTTP(w, req)
}

type apiRole int

const (
	regularRole apiRole = iota
	adminRole
	superUserRole
)

type roleAuthorizationMux struct {
	ie     *sql.InternalExecutor
	role   apiRole
	option roleoption.Option
	inner  http.Handler
}

func (r *roleAuthorizationMux) getRoleForUser(
	ctx context.Context, user security.SQLUsername,
) (apiRole, error) {
	__antithesis_instrumentation__.Notify(188997)
	if user.IsRootUser() {
		__antithesis_instrumentation__.Notify(189004)

		return superUserRole, nil
	} else {
		__antithesis_instrumentation__.Notify(189005)
	}
	__antithesis_instrumentation__.Notify(188998)
	row, err := r.ie.QueryRowEx(
		ctx, "check-is-admin", nil,
		sessiondata.InternalExecutorOverride{User: user},
		"SELECT crdb_internal.is_admin()")
	if err != nil {
		__antithesis_instrumentation__.Notify(189006)
		return regularRole, err
	} else {
		__antithesis_instrumentation__.Notify(189007)
	}
	__antithesis_instrumentation__.Notify(188999)
	if row == nil {
		__antithesis_instrumentation__.Notify(189008)
		return regularRole, errors.AssertionFailedf("hasAdminRole: expected 1 row, got 0")
	} else {
		__antithesis_instrumentation__.Notify(189009)
	}
	__antithesis_instrumentation__.Notify(189000)
	if len(row) != 1 {
		__antithesis_instrumentation__.Notify(189010)
		return regularRole, errors.AssertionFailedf("hasAdminRole: expected 1 column, got %d", len(row))
	} else {
		__antithesis_instrumentation__.Notify(189011)
	}
	__antithesis_instrumentation__.Notify(189001)
	dbDatum, ok := tree.AsDBool(row[0])
	if !ok {
		__antithesis_instrumentation__.Notify(189012)
		return regularRole, errors.AssertionFailedf("hasAdminRole: expected bool, got %T", row[0])
	} else {
		__antithesis_instrumentation__.Notify(189013)
	}
	__antithesis_instrumentation__.Notify(189002)
	if dbDatum {
		__antithesis_instrumentation__.Notify(189014)
		return adminRole, nil
	} else {
		__antithesis_instrumentation__.Notify(189015)
	}
	__antithesis_instrumentation__.Notify(189003)
	return regularRole, nil
}

func (r *roleAuthorizationMux) hasRoleOption(
	ctx context.Context, user security.SQLUsername, roleOption roleoption.Option,
) (bool, error) {
	__antithesis_instrumentation__.Notify(189016)
	if user.IsRootUser() {
		__antithesis_instrumentation__.Notify(189022)

		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(189023)
	}
	__antithesis_instrumentation__.Notify(189017)
	row, err := r.ie.QueryRowEx(
		ctx, "check-role-option", nil,
		sessiondata.InternalExecutorOverride{User: user},
		"SELECT crdb_internal.has_role_option($1)", roleOption.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(189024)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(189025)
	}
	__antithesis_instrumentation__.Notify(189018)
	if row == nil {
		__antithesis_instrumentation__.Notify(189026)
		return false, errors.AssertionFailedf("hasRoleOption: expected 1 row, got 0")
	} else {
		__antithesis_instrumentation__.Notify(189027)
	}
	__antithesis_instrumentation__.Notify(189019)
	if len(row) != 1 {
		__antithesis_instrumentation__.Notify(189028)
		return false, errors.AssertionFailedf("hasRoleOption: expected 1 column, got %d", len(row))
	} else {
		__antithesis_instrumentation__.Notify(189029)
	}
	__antithesis_instrumentation__.Notify(189020)
	dbDatum, ok := tree.AsDBool(row[0])
	if !ok {
		__antithesis_instrumentation__.Notify(189030)
		return false, errors.AssertionFailedf("hasRoleOption: expected bool, got %T", row[0])
	} else {
		__antithesis_instrumentation__.Notify(189031)
	}
	__antithesis_instrumentation__.Notify(189021)
	return bool(dbDatum), nil
}

func (r *roleAuthorizationMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	__antithesis_instrumentation__.Notify(189032)

	username := security.MakeSQLUsernameFromPreNormalizedString(
		req.Context().Value(webSessionUserKey{}).(string))
	if role, err := r.getRoleForUser(req.Context(), username); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(189035)
		return role < r.role == true
	}() == true {
		__antithesis_instrumentation__.Notify(189036)
		if err != nil {
			__antithesis_instrumentation__.Notify(189038)
			apiV2InternalError(req.Context(), err, w)
		} else {
			__antithesis_instrumentation__.Notify(189039)
			http.Error(w, "user not allowed to access this endpoint", http.StatusForbidden)
		}
		__antithesis_instrumentation__.Notify(189037)
		return
	} else {
		__antithesis_instrumentation__.Notify(189040)
	}
	__antithesis_instrumentation__.Notify(189033)
	if r.option > 0 {
		__antithesis_instrumentation__.Notify(189041)
		ok, err := r.hasRoleOption(req.Context(), username, r.option)
		if err != nil {
			__antithesis_instrumentation__.Notify(189042)
			apiV2InternalError(req.Context(), err, w)
			return
		} else {
			__antithesis_instrumentation__.Notify(189043)
			if !ok {
				__antithesis_instrumentation__.Notify(189044)
				http.Error(w, "user not allowed to access this endpoint", http.StatusForbidden)
				return
			} else {
				__antithesis_instrumentation__.Notify(189045)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(189046)
	}
	__antithesis_instrumentation__.Notify(189034)
	r.inner.ServeHTTP(w, req)
}

func apiToOutgoingGatewayCtx(ctx context.Context, r *http.Request) context.Context {
	__antithesis_instrumentation__.Notify(189047)
	return metadata.NewOutgoingContext(ctx, forwardAuthenticationMetadata(ctx, r))
}
