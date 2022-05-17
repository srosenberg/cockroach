package oidcccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/ui"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	oidc "github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

const (
	idTokenKey               = "id_token"
	codeKey                  = "code"
	stateKey                 = "state"
	secretCookieName         = "oidc_secret"
	oidcLoginPath            = "/oidc/v1/login"
	oidcCallbackPath         = "/oidc/v1/callback"
	genericCallbackHTTPError = "OIDC: unable to complete authentication"
	genericLoginHTTPError    = "OIDC: unable to initiate authentication"
	counterPrefix            = "auth.oidc."
	beginAuthCounterName     = counterPrefix + "begin_auth"
	beginCallbackCounterName = counterPrefix + "begin_callback"
	loginSuccessCounterName  = counterPrefix + "login_success"
	enableCounterName        = counterPrefix + "enable"
	hmacKeySize              = 32
	stateTokenSize           = 32
)

var (
	beginAuthUseCounter     = telemetry.GetCounterOnce(beginAuthCounterName)
	beginCallbackUseCounter = telemetry.GetCounterOnce(beginCallbackCounterName)
	loginSuccessUseCounter  = telemetry.GetCounterOnce(loginSuccessCounterName)
	enableUseCounter        = telemetry.GetCounterOnce(enableCounterName)
)

type oidcAuthenticationServer struct {
	mutex        syncutil.RWMutex
	conf         oidcAuthenticationConf
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier

	enabled     bool
	initialized bool
}

type oidcAuthenticationConf struct {
	clientID        string
	clientSecret    string
	redirectURLConf redirectURLConf
	providerURL     string
	scopes          string
	enabled         bool
	claimJSONKey    string
	principalRegex  *regexp.Regexp
	buttonText      string
	autoLogin       bool
}

func (s *oidcAuthenticationServer) GetOIDCConf() ui.OIDCUIConf {
	__antithesis_instrumentation__.Notify(20452)
	return ui.OIDCUIConf{
		ButtonText: s.conf.buttonText,
		Enabled:    s.enabled,
		AutoLogin:  s.conf.autoLogin,
	}
}

func reloadConfig(
	ctx context.Context,
	server *oidcAuthenticationServer,
	locality roachpb.Locality,
	st *cluster.Settings,
) {
	__antithesis_instrumentation__.Notify(20453)
	server.mutex.Lock()
	defer server.mutex.Unlock()
	reloadConfigLocked(ctx, server, locality, st)
}

func reloadConfigLocked(
	ctx context.Context,
	server *oidcAuthenticationServer,
	locality roachpb.Locality,
	st *cluster.Settings,
) {
	__antithesis_instrumentation__.Notify(20454)
	conf := oidcAuthenticationConf{
		clientID:        OIDCClientID.Get(&st.SV),
		clientSecret:    OIDCClientSecret.Get(&st.SV),
		redirectURLConf: mustParseOIDCRedirectURL(OIDCRedirectURL.Get(&st.SV)),
		providerURL:     OIDCProviderURL.Get(&st.SV),
		scopes:          OIDCScopes.Get(&st.SV),
		claimJSONKey:    OIDCClaimJSONKey.Get(&st.SV),
		enabled:         OIDCEnabled.Get(&st.SV),

		principalRegex: regexp.MustCompile(OIDCPrincipalRegex.Get(&st.SV)),
		buttonText:     OIDCButtonText.Get(&st.SV),
		autoLogin:      OIDCAutoLogin.Get(&st.SV),
	}

	if !server.conf.enabled && func() bool {
		__antithesis_instrumentation__.Notify(20459)
		return conf.enabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(20460)
		telemetry.Inc(enableUseCounter)
	} else {
		__antithesis_instrumentation__.Notify(20461)
	}
	__antithesis_instrumentation__.Notify(20455)

	server.initialized = false
	server.conf = conf
	if server.conf.enabled {
		__antithesis_instrumentation__.Notify(20462)

		server.enabled = true
	} else {
		__antithesis_instrumentation__.Notify(20463)
		server.enabled = false
		return
	}
	__antithesis_instrumentation__.Notify(20456)

	provider, err := oidc.NewProvider(ctx, server.conf.providerURL)
	if err != nil {
		__antithesis_instrumentation__.Notify(20464)
		log.Warningf(ctx, "unable to initialize OIDC provider, disabling OIDC: %v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(20465)
	}
	__antithesis_instrumentation__.Notify(20457)

	scopesForOauth := strings.Split(server.conf.scopes, " ")

	redirectURL, err := getRegionSpecificRedirectURL(locality, server.conf.redirectURLConf)
	if err != nil {
		__antithesis_instrumentation__.Notify(20466)
		log.Warningf(ctx, "unable to initialize OIDC provider, disabling OIDC: %v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(20467)
	}
	__antithesis_instrumentation__.Notify(20458)

	server.oauth2Config = oauth2.Config{
		ClientID:     server.conf.clientID,
		ClientSecret: server.conf.clientSecret,
		RedirectURL:  redirectURL,

		Endpoint: provider.Endpoint(),
		Scopes:   scopesForOauth,
	}

	server.verifier = provider.Verifier(&oidc.Config{ClientID: server.conf.clientID})
	server.initialized = true
	log.Infof(ctx, "initialized OIDC server")
}

func getRegionSpecificRedirectURL(locality roachpb.Locality, conf redirectURLConf) (string, error) {
	__antithesis_instrumentation__.Notify(20468)
	if len(locality.Tiers) > 0 {
		__antithesis_instrumentation__.Notify(20471)
		region, containsRegion := locality.Find("region")
		if containsRegion {
			__antithesis_instrumentation__.Notify(20472)
			if redirectURL, ok := conf.getForRegion(region); ok {
				__antithesis_instrumentation__.Notify(20474)
				return redirectURL, nil
			} else {
				__antithesis_instrumentation__.Notify(20475)
			}
			__antithesis_instrumentation__.Notify(20473)
			return "", errors.Newf("OIDC: no matching redirect URL found for region %s", region)
		} else {
			__antithesis_instrumentation__.Notify(20476)
		}
	} else {
		__antithesis_instrumentation__.Notify(20477)
	}
	__antithesis_instrumentation__.Notify(20469)
	s, ok := conf.get()
	if !ok {
		__antithesis_instrumentation__.Notify(20478)
		return "", errors.New("OIDC: redirect URL config expects region setting, which is unset")
	} else {
		__antithesis_instrumentation__.Notify(20479)
	}
	__antithesis_instrumentation__.Notify(20470)
	return s, nil
}

var ConfigureOIDC = func(
	serverCtx context.Context,
	st *cluster.Settings,
	locality roachpb.Locality,
	handleHTTP func(pattern string, handler http.Handler),
	userLoginFromSSO func(ctx context.Context, username string) (*http.Cookie, error),
	ambientCtx log.AmbientContext,
	cluster uuid.UUID,
) (server.OIDC, error) {
	__antithesis_instrumentation__.Notify(20480)
	oidcAuthentication := &oidcAuthenticationServer{}

	handleHTTP(oidcCallbackPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		__antithesis_instrumentation__.Notify(20493)
		ctx := r.Context()

		oidcAuthentication.mutex.Lock()
		defer oidcAuthentication.mutex.Unlock()

		if oidcAuthentication.enabled && func() bool {
			__antithesis_instrumentation__.Notify(20507)
			return !oidcAuthentication.initialized == true
		}() == true {
			__antithesis_instrumentation__.Notify(20508)
			reloadConfigLocked(ctx, oidcAuthentication, locality, st)
		} else {
			__antithesis_instrumentation__.Notify(20509)
		}
		__antithesis_instrumentation__.Notify(20494)

		if !oidcAuthentication.enabled {
			__antithesis_instrumentation__.Notify(20510)
			http.Error(w, "OIDC: disabled", http.StatusBadRequest)
			return
		} else {
			__antithesis_instrumentation__.Notify(20511)
		}
		__antithesis_instrumentation__.Notify(20495)

		telemetry.Inc(beginCallbackUseCounter)

		state := r.URL.Query().Get(stateKey)

		secretCookie, err := r.Cookie(secretCookieName)
		if err != nil {
			__antithesis_instrumentation__.Notify(20512)
			log.Errorf(ctx, "OIDC: missing client side cookie: %v", err)
			http.Error(w, genericCallbackHTTPError, http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(20513)
		}
		__antithesis_instrumentation__.Notify(20496)

		kast := keyAndSignedToken{
			secretCookie,
			state,
		}

		valid, err := kast.validate()
		if err != nil {
			__antithesis_instrumentation__.Notify(20514)
			log.Errorf(ctx, "OIDC: validating client cookie and state token pair: %v", err)
			http.Error(w, genericCallbackHTTPError, http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(20515)
		}
		__antithesis_instrumentation__.Notify(20497)
		if !valid {
			__antithesis_instrumentation__.Notify(20516)
			log.Error(ctx, "OIDC: invalid client cooke and state token pair")
			http.Error(w, genericCallbackHTTPError, http.StatusBadRequest)
			return
		} else {
			__antithesis_instrumentation__.Notify(20517)
		}
		__antithesis_instrumentation__.Notify(20498)

		oauth2Token, err := oidcAuthentication.oauth2Config.Exchange(ctx, r.URL.Query().Get(codeKey))
		if err != nil {
			__antithesis_instrumentation__.Notify(20518)
			log.Errorf(ctx, "OIDC: failed to exchange code for token: %v", err)
			http.Error(w, genericCallbackHTTPError, http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(20519)
		}
		__antithesis_instrumentation__.Notify(20499)

		rawIDToken, ok := oauth2Token.Extra(idTokenKey).(string)
		if !ok {
			__antithesis_instrumentation__.Notify(20520)
			log.Error(ctx, "OIDC: failed to extract ID token from OAuth2 token")
			http.Error(w, genericCallbackHTTPError, http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(20521)
		}
		__antithesis_instrumentation__.Notify(20500)

		idToken, err := oidcAuthentication.verifier.Verify(ctx, rawIDToken)
		if err != nil {
			__antithesis_instrumentation__.Notify(20522)
			log.Errorf(ctx, "OIDC: unable to verify ID token: %v", err)
			http.Error(w, genericCallbackHTTPError, http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(20523)
		}
		__antithesis_instrumentation__.Notify(20501)

		var claims map[string]json.RawMessage
		if err := idToken.Claims(&claims); err != nil {
			__antithesis_instrumentation__.Notify(20524)
			log.Errorf(ctx, "OIDC: unable to deserialize token claims: %v", err)
			http.Error(w, genericCallbackHTTPError, http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(20525)
		}
		__antithesis_instrumentation__.Notify(20502)

		var principal string
		claim := claims[oidcAuthentication.conf.claimJSONKey]
		if err := json.Unmarshal(claim, &principal); err != nil {
			__antithesis_instrumentation__.Notify(20526)
			log.Errorf(ctx, "OIDC: failed to complete authentication: failed to extract claim key %s: %v", oidcAuthentication.conf.claimJSONKey, err)
			http.Error(w, genericCallbackHTTPError, http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(20527)
		}
		__antithesis_instrumentation__.Notify(20503)

		match := oidcAuthentication.conf.principalRegex.FindStringSubmatch(principal)
		numGroups := len(match)
		if numGroups != 2 {
			__antithesis_instrumentation__.Notify(20528)
			log.Errorf(ctx, "OIDC: failed to complete authentication: expected one group in regexp, got %d", numGroups)
			http.Error(w, genericCallbackHTTPError, http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(20529)
		}
		__antithesis_instrumentation__.Notify(20504)

		username := match[1]
		cookie, err := userLoginFromSSO(ctx, username)
		if err != nil {
			__antithesis_instrumentation__.Notify(20530)
			log.Errorf(ctx, "OIDC: failed to complete authentication: unable to create session for %s: %v", username, err)
			http.Error(w, genericCallbackHTTPError, http.StatusForbidden)
			return
		} else {
			__antithesis_instrumentation__.Notify(20531)
		}
		__antithesis_instrumentation__.Notify(20505)

		org := sql.ClusterOrganization.Get(&st.SV)
		if err := utilccl.CheckEnterpriseEnabled(st, cluster, org, "OIDC"); err != nil {
			__antithesis_instrumentation__.Notify(20532)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			__antithesis_instrumentation__.Notify(20533)
		}
		__antithesis_instrumentation__.Notify(20506)

		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

		telemetry.Inc(loginSuccessUseCounter)
	}))
	__antithesis_instrumentation__.Notify(20481)

	handleHTTP(oidcLoginPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		__antithesis_instrumentation__.Notify(20534)
		ctx := r.Context()

		oidcAuthentication.mutex.Lock()
		defer oidcAuthentication.mutex.Unlock()

		if oidcAuthentication.enabled && func() bool {
			__antithesis_instrumentation__.Notify(20538)
			return !oidcAuthentication.initialized == true
		}() == true {
			__antithesis_instrumentation__.Notify(20539)
			reloadConfigLocked(ctx, oidcAuthentication, locality, st)
		} else {
			__antithesis_instrumentation__.Notify(20540)
		}
		__antithesis_instrumentation__.Notify(20535)

		if !oidcAuthentication.enabled {
			__antithesis_instrumentation__.Notify(20541)
			http.Error(w, "OIDC: disabled", http.StatusBadRequest)
			return
		} else {
			__antithesis_instrumentation__.Notify(20542)
		}
		__antithesis_instrumentation__.Notify(20536)

		telemetry.Inc(beginAuthUseCounter)

		kast, err := newKeyAndSignedToken(hmacKeySize, stateTokenSize)
		if err != nil {
			__antithesis_instrumentation__.Notify(20543)
			log.Errorf(ctx, "OIDC: unable to generate key and signed message: %v", err)
			http.Error(w, genericLoginHTTPError, http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(20544)
		}
		__antithesis_instrumentation__.Notify(20537)

		http.SetCookie(w, kast.secretKeyCookie)
		http.Redirect(
			w, r, oidcAuthentication.oauth2Config.AuthCodeURL(kast.signedTokenEncoded), http.StatusFound,
		)
	}))
	__antithesis_instrumentation__.Notify(20482)

	reloadConfig(serverCtx, oidcAuthentication, locality, st)

	OIDCEnabled.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20545)
		reloadConfig(ambientCtx.AnnotateCtx(ctx), oidcAuthentication, locality, st)
	})
	__antithesis_instrumentation__.Notify(20483)
	OIDCClientID.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20546)
		reloadConfig(ambientCtx.AnnotateCtx(ctx), oidcAuthentication, locality, st)
	})
	__antithesis_instrumentation__.Notify(20484)
	OIDCClientSecret.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20547)
		reloadConfig(ambientCtx.AnnotateCtx(ctx), oidcAuthentication, locality, st)
	})
	__antithesis_instrumentation__.Notify(20485)
	OIDCRedirectURL.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20548)
		reloadConfig(ambientCtx.AnnotateCtx(ctx), oidcAuthentication, locality, st)
	})
	__antithesis_instrumentation__.Notify(20486)
	OIDCProviderURL.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20549)
		reloadConfig(ambientCtx.AnnotateCtx(ctx), oidcAuthentication, locality, st)
	})
	__antithesis_instrumentation__.Notify(20487)
	OIDCScopes.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20550)
		reloadConfig(ambientCtx.AnnotateCtx(ctx), oidcAuthentication, locality, st)
	})
	__antithesis_instrumentation__.Notify(20488)
	OIDCClaimJSONKey.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20551)
		reloadConfig(ambientCtx.AnnotateCtx(ctx), oidcAuthentication, locality, st)
	})
	__antithesis_instrumentation__.Notify(20489)
	OIDCPrincipalRegex.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20552)
		reloadConfig(ambientCtx.AnnotateCtx(ctx), oidcAuthentication, locality, st)
	})
	__antithesis_instrumentation__.Notify(20490)
	OIDCButtonText.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20553)
		reloadConfig(ambientCtx.AnnotateCtx(ctx), oidcAuthentication, locality, st)
	})
	__antithesis_instrumentation__.Notify(20491)
	OIDCAutoLogin.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(20554)
		reloadConfig(ambientCtx.AnnotateCtx(ctx), oidcAuthentication, locality, st)
	})
	__antithesis_instrumentation__.Notify(20492)

	return oidcAuthentication, nil
}

func init() {
	server.ConfigureOIDC = ConfigureOIDC
}
