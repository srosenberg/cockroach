package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

type httpTestServer struct {
	t struct {
		authentication *authenticationServer
		sqlServer      *SQLServer
	}

	authClient [2]struct {
		httpClient http.Client
		cookie     *serverpb.SessionCookie
		once       sync.Once
		err        error
	}
}

func (ts *httpTestServer) AdminURL() string {
	__antithesis_instrumentation__.Notify(239459)
	return ts.t.sqlServer.execCfg.RPCContext.Config.AdminURL().String()
}

func (ts *httpTestServer) GetHTTPClient() (http.Client, error) {
	__antithesis_instrumentation__.Notify(239460)
	return ts.t.sqlServer.execCfg.RPCContext.GetHTTPClient()
}

func (ts *httpTestServer) GetAdminAuthenticatedHTTPClient() (http.Client, error) {
	__antithesis_instrumentation__.Notify(239461)
	httpClient, _, err := ts.getAuthenticatedHTTPClientAndCookie(authenticatedUserName(), true)
	return httpClient, err
}

func (ts *httpTestServer) GetAuthenticatedHTTPClient(isAdmin bool) (http.Client, error) {
	__antithesis_instrumentation__.Notify(239462)
	authUser := authenticatedUserName()
	if !isAdmin {
		__antithesis_instrumentation__.Notify(239464)
		authUser = authenticatedUserNameNoAdmin()
	} else {
		__antithesis_instrumentation__.Notify(239465)
	}
	__antithesis_instrumentation__.Notify(239463)
	httpClient, _, err := ts.getAuthenticatedHTTPClientAndCookie(authUser, isAdmin)
	return httpClient, err
}

func (ts *httpTestServer) getAuthenticatedHTTPClientAndCookie(
	authUser security.SQLUsername, isAdmin bool,
) (http.Client, *serverpb.SessionCookie, error) {
	__antithesis_instrumentation__.Notify(239466)
	authIdx := 0
	if isAdmin {
		__antithesis_instrumentation__.Notify(239469)
		authIdx = 1
	} else {
		__antithesis_instrumentation__.Notify(239470)
	}
	__antithesis_instrumentation__.Notify(239467)
	authClient := &ts.authClient[authIdx]
	authClient.once.Do(func() {
		__antithesis_instrumentation__.Notify(239471)

		authClient.err = func() error {
			__antithesis_instrumentation__.Notify(239472)

			if err := ts.createAuthUser(authUser, isAdmin); err != nil {
				__antithesis_instrumentation__.Notify(239480)
				return err
			} else {
				__antithesis_instrumentation__.Notify(239481)
			}
			__antithesis_instrumentation__.Notify(239473)

			id, secret, err := ts.t.authentication.newAuthSession(context.TODO(), authUser)
			if err != nil {
				__antithesis_instrumentation__.Notify(239482)
				return err
			} else {
				__antithesis_instrumentation__.Notify(239483)
			}
			__antithesis_instrumentation__.Notify(239474)
			rawCookie := &serverpb.SessionCookie{
				ID:     id,
				Secret: secret,
			}

			cookie, err := EncodeSessionCookie(rawCookie, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(239484)
				return err
			} else {
				__antithesis_instrumentation__.Notify(239485)
			}
			__antithesis_instrumentation__.Notify(239475)
			cookieJar, err := cookiejar.New(nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(239486)
				return err
			} else {
				__antithesis_instrumentation__.Notify(239487)
			}
			__antithesis_instrumentation__.Notify(239476)
			url, err := url.Parse(ts.t.sqlServer.execCfg.RPCContext.Config.AdminURL().String())
			if err != nil {
				__antithesis_instrumentation__.Notify(239488)
				return err
			} else {
				__antithesis_instrumentation__.Notify(239489)
			}
			__antithesis_instrumentation__.Notify(239477)
			cookieJar.SetCookies(url, []*http.Cookie{cookie})

			authClient.httpClient, err = ts.t.sqlServer.execCfg.RPCContext.GetHTTPClient()
			if err != nil {
				__antithesis_instrumentation__.Notify(239490)
				return err
			} else {
				__antithesis_instrumentation__.Notify(239491)
			}
			__antithesis_instrumentation__.Notify(239478)
			rawCookieBytes, err := protoutil.Marshal(rawCookie)
			if err != nil {
				__antithesis_instrumentation__.Notify(239492)
				return err
			} else {
				__antithesis_instrumentation__.Notify(239493)
			}
			__antithesis_instrumentation__.Notify(239479)
			authClient.httpClient.Transport = &v2AuthDecorator{
				RoundTripper: authClient.httpClient.Transport,
				session:      base64.StdEncoding.EncodeToString(rawCookieBytes),
			}
			authClient.httpClient.Jar = cookieJar
			authClient.cookie = rawCookie
			return nil
		}()
	})
	__antithesis_instrumentation__.Notify(239468)

	return authClient.httpClient, authClient.cookie, authClient.err
}

func (ts *httpTestServer) createAuthUser(userName security.SQLUsername, isAdmin bool) error {
	__antithesis_instrumentation__.Notify(239494)
	if _, err := ts.t.sqlServer.internalExecutor.ExecEx(context.TODO(),
		"create-auth-user", nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf("CREATE USER %s", userName.Normalized()),
	); err != nil {
		__antithesis_instrumentation__.Notify(239497)
		return err
	} else {
		__antithesis_instrumentation__.Notify(239498)
	}
	__antithesis_instrumentation__.Notify(239495)
	if isAdmin {
		__antithesis_instrumentation__.Notify(239499)

		if _, err := ts.t.sqlServer.internalExecutor.ExecEx(context.TODO(),
			"grant-admin", nil,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			"INSERT INTO system.role_members (role, member, \"isAdmin\") VALUES ('admin', $1, true)", userName.Normalized(),
		); err != nil {
			__antithesis_instrumentation__.Notify(239500)
			return err
		} else {
			__antithesis_instrumentation__.Notify(239501)
		}
	} else {
		__antithesis_instrumentation__.Notify(239502)
	}
	__antithesis_instrumentation__.Notify(239496)
	return nil
}
