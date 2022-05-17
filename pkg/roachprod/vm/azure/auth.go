package azure

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/cockroachdb/errors"
)

func (p *Provider) getAuthorizer() (ret autorest.Authorizer, err error) {
	__antithesis_instrumentation__.Notify(183082)
	p.mu.Lock()
	ret = p.mu.authorizer
	p.mu.Unlock()
	if ret != nil {
		__antithesis_instrumentation__.Notify(183085)
		return
	} else {
		__antithesis_instrumentation__.Notify(183086)
	}
	__antithesis_instrumentation__.Notify(183083)

	ret, err = auth.NewAuthorizerFromCLI()
	if err == nil {
		__antithesis_instrumentation__.Notify(183087)
		p.mu.Lock()
		p.mu.authorizer = ret
		p.mu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(183088)
		err = errors.Wrap(err, "could got get Azure auth token")
	}
	__antithesis_instrumentation__.Notify(183084)
	return
}

func (p *Provider) getAuthToken() (string, error) {
	__antithesis_instrumentation__.Notify(183089)
	auth, err := p.getAuthorizer()
	if err != nil {
		__antithesis_instrumentation__.Notify(183092)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183093)
	}
	__antithesis_instrumentation__.Notify(183090)

	fake := &http.Request{}
	if _, err := auth.WithAuthorization()(&stealAuth{}).Prepare(fake); err != nil {
		__antithesis_instrumentation__.Notify(183094)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183095)
	}
	__antithesis_instrumentation__.Notify(183091)
	return fake.Header.Get("Authorization")[7:], nil
}

type stealAuth struct {
}

var _ autorest.Preparer = &stealAuth{}

func (*stealAuth) Prepare(r *http.Request) (*http.Request, error) {
	__antithesis_instrumentation__.Notify(183096)
	return r, nil
}
