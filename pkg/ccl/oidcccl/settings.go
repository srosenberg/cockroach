package oidcccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"encoding/json"
	"net/url"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/errors"
	"github.com/coreos/go-oidc"
)

const (
	baseOIDCSettingName           = "server.oidc_authentication."
	OIDCEnabledSettingName        = baseOIDCSettingName + "enabled"
	OIDCClientIDSettingName       = baseOIDCSettingName + "client_id"
	OIDCClientSecretSettingName   = baseOIDCSettingName + "client_secret"
	OIDCRedirectURLSettingName    = baseOIDCSettingName + "redirect_url"
	OIDCProviderURLSettingName    = baseOIDCSettingName + "provider_url"
	OIDCScopesSettingName         = baseOIDCSettingName + "scopes"
	OIDCClaimJSONKeySettingName   = baseOIDCSettingName + "claim_json_key"
	OIDCPrincipalRegexSettingName = baseOIDCSettingName + "principal_regex"
	OIDCButtonTextSettingName     = baseOIDCSettingName + "button_text"
	OIDCAutoLoginSettingName      = baseOIDCSettingName + "autologin"
)

var OIDCEnabled = func() *settings.BoolSetting {
	__antithesis_instrumentation__.Notify(20555)
	s := settings.RegisterBoolSetting(
		settings.TenantWritable,
		OIDCEnabledSettingName,
		"enables or disabled OIDC login for the DB Console",
		false,
	).WithPublic()
	s.SetReportable(true)
	return s
}()

var OIDCClientID = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(20556)
	s := settings.RegisterStringSetting(
		settings.TenantWritable,
		OIDCClientIDSettingName,
		"sets OIDC client id",
		"",
	).WithPublic()
	s.SetReportable(true)
	return s
}()

var OIDCClientSecret = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(20557)
	s := settings.RegisterStringSetting(
		settings.TenantWritable,
		OIDCClientSecretSettingName,
		"sets OIDC client secret",
		"",
	).WithPublic()
	s.SetReportable(false)
	return s
}()

type redirectURLConf struct {
	mrru *multiRegionRedirectURLs
	sru  *singleRedirectURL
}

func (conf *redirectURLConf) getForRegion(region string) (string, bool) {
	__antithesis_instrumentation__.Notify(20558)
	if conf.mrru != nil {
		__antithesis_instrumentation__.Notify(20561)
		s, ok := conf.mrru.RedirectURLs[region]
		return s, ok
	} else {
		__antithesis_instrumentation__.Notify(20562)
	}
	__antithesis_instrumentation__.Notify(20559)
	if conf.sru != nil {
		__antithesis_instrumentation__.Notify(20563)
		return conf.sru.RedirectURL, true
	} else {
		__antithesis_instrumentation__.Notify(20564)
	}
	__antithesis_instrumentation__.Notify(20560)
	return "", false
}

func (conf *redirectURLConf) get() (string, bool) {
	__antithesis_instrumentation__.Notify(20565)
	if conf.sru != nil {
		__antithesis_instrumentation__.Notify(20567)
		return conf.sru.RedirectURL, true
	} else {
		__antithesis_instrumentation__.Notify(20568)
	}
	__antithesis_instrumentation__.Notify(20566)
	return "", false
}

type multiRegionRedirectURLs struct {
	RedirectURLs map[string]string `json:"redirect_urls"`
}

type singleRedirectURL struct {
	RedirectURL string
}

func mustParseOIDCRedirectURL(s string) redirectURLConf {
	__antithesis_instrumentation__.Notify(20569)
	var mrru = multiRegionRedirectURLs{}
	decoder := json.NewDecoder(bytes.NewReader([]byte(s)))
	err := decoder.Decode(&mrru)
	if err != nil {
		__antithesis_instrumentation__.Notify(20571)
		return redirectURLConf{sru: &singleRedirectURL{RedirectURL: s}}
	} else {
		__antithesis_instrumentation__.Notify(20572)
	}
	__antithesis_instrumentation__.Notify(20570)
	return redirectURLConf{mrru: &mrru}
}

func validateOIDCRedirectURL(values *settings.Values, s string) error {
	__antithesis_instrumentation__.Notify(20573)
	var mrru = multiRegionRedirectURLs{}

	var jsonCheck json.RawMessage
	if json.Unmarshal([]byte(s), &jsonCheck) != nil {
		__antithesis_instrumentation__.Notify(20577)

		if _, err := url.Parse(s); err != nil {
			__antithesis_instrumentation__.Notify(20579)
			return errors.Wrap(err, "OIDC redirect URL not valid")
		} else {
			__antithesis_instrumentation__.Notify(20580)
		}
		__antithesis_instrumentation__.Notify(20578)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(20581)
	}
	__antithesis_instrumentation__.Notify(20574)

	decoder := json.NewDecoder(bytes.NewReader([]byte(s)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&mrru); err != nil {
		__antithesis_instrumentation__.Notify(20582)
		return errors.Wrap(err, "OIDC redirect JSON not valid")
	} else {
		__antithesis_instrumentation__.Notify(20583)
	}
	__antithesis_instrumentation__.Notify(20575)
	for _, route := range mrru.RedirectURLs {
		__antithesis_instrumentation__.Notify(20584)
		if _, err := url.Parse(route); err != nil {
			__antithesis_instrumentation__.Notify(20585)
			return errors.Wrapf(err, "OIDC redirect JSON contains invalid URL: %s", route)
		} else {
			__antithesis_instrumentation__.Notify(20586)
		}
	}
	__antithesis_instrumentation__.Notify(20576)
	return nil
}

var OIDCRedirectURL = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(20587)
	s := settings.RegisterValidatedStringSetting(
		settings.TenantWritable,
		OIDCRedirectURLSettingName,
		"sets OIDC redirect URL via a URL string or a JSON string containing a required "+
			"`redirect_urls` key with an object that maps from region keys to URL strings "+
			"(URLs should point to your load balancer and must route to the path /oidc/v1/callback) ",
		"https://localhost:8080/oidc/v1/callback",
		validateOIDCRedirectURL,
	)
	s.SetReportable(true)
	s.SetVisibility(settings.Public)
	return s
}()

var OIDCProviderURL = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(20588)
	s := settings.RegisterValidatedStringSetting(
		settings.TenantWritable,
		OIDCProviderURLSettingName,
		"sets OIDC provider URL ({provider_url}/.well-known/openid-configuration must resolve)",
		"",
		func(values *settings.Values, s string) error {
			__antithesis_instrumentation__.Notify(20590)
			if _, err := url.Parse(s); err != nil {
				__antithesis_instrumentation__.Notify(20592)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20593)
			}
			__antithesis_instrumentation__.Notify(20591)
			return nil
		},
	)
	__antithesis_instrumentation__.Notify(20589)
	s.SetReportable(true)
	s.SetVisibility(settings.Public)
	return s
}()

var OIDCScopes = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(20594)
	s := settings.RegisterValidatedStringSetting(
		settings.TenantWritable,
		OIDCScopesSettingName,
		"sets OIDC scopes to include with authentication request "+
			"(space delimited list of strings, required to start with `openid`)",
		"openid",
		func(values *settings.Values, s string) error {
			__antithesis_instrumentation__.Notify(20596)
			if s != oidc.ScopeOpenID && func() bool {
				__antithesis_instrumentation__.Notify(20598)
				return !strings.HasPrefix(s, oidc.ScopeOpenID+" ") == true
			}() == true {
				__antithesis_instrumentation__.Notify(20599)
				return errors.New("Missing `openid` scope which is required for OIDC")
			} else {
				__antithesis_instrumentation__.Notify(20600)
			}
			__antithesis_instrumentation__.Notify(20597)
			return nil
		},
	)
	__antithesis_instrumentation__.Notify(20595)
	s.SetReportable(true)
	s.SetVisibility(settings.Public)
	return s
}()

var OIDCClaimJSONKey = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(20601)
	s := settings.RegisterStringSetting(
		settings.TenantWritable,
		OIDCClaimJSONKeySettingName,
		"sets JSON key of principal to extract from payload after OIDC authentication completes "+
			"(usually email or sid)",
		"",
	).WithPublic()
	return s
}()

var OIDCPrincipalRegex = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(20602)
	s := settings.RegisterValidatedStringSetting(
		settings.TenantWritable,
		OIDCPrincipalRegexSettingName,
		"regular expression to apply to extracted principal (see claim_json_key setting) to "+
			"translate to SQL user (golang regex format, must include 1 grouping to extract)",
		"(.+)",
		func(values *settings.Values, s string) error {
			__antithesis_instrumentation__.Notify(20604)
			if _, err := regexp.Compile(s); err != nil {
				__antithesis_instrumentation__.Notify(20606)
				return errors.Wrapf(err, "unable to initialize %s setting, regex does not compile",
					OIDCPrincipalRegexSettingName)
			} else {
				__antithesis_instrumentation__.Notify(20607)
			}
			__antithesis_instrumentation__.Notify(20605)
			return nil
		},
	)
	__antithesis_instrumentation__.Notify(20603)
	s.SetVisibility(settings.Public)
	return s
}()

var OIDCButtonText = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(20608)
	s := settings.RegisterStringSetting(
		settings.TenantWritable,
		OIDCButtonTextSettingName,
		"text to show on button on DB Console login page to login with your OIDC provider "+
			"(only shown if OIDC is enabled)",
		"Login with your OIDC provider",
	).WithPublic()
	return s
}()

var OIDCAutoLogin = func() *settings.BoolSetting {
	__antithesis_instrumentation__.Notify(20609)
	s := settings.RegisterBoolSetting(
		settings.TenantWritable,
		OIDCAutoLoginSettingName,
		"if true, logged-out visitors to the DB Console will be "+
			"automatically redirected to the OIDC login endpoint",
		false,
	).WithPublic()
	return s
}()
