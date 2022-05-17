package flagstub

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/errors"
)

func New(delegate vm.Provider, unimplemented string) vm.Provider {
	__antithesis_instrumentation__.Notify(183562)
	return &provider{delegate: delegate, unimplemented: unimplemented}
}

type provider struct {
	delegate      vm.Provider
	unimplemented string
}

func (p *provider) CleanSSH() error {
	__antithesis_instrumentation__.Notify(183563)
	return nil
}

func (p *provider) ConfigSSH(zones []string) error {
	__antithesis_instrumentation__.Notify(183564)
	return nil
}

func (p *provider) Create(
	l *logger.Logger, names []string, opts vm.CreateOpts, providerOpts vm.ProviderOpts,
) error {
	__antithesis_instrumentation__.Notify(183565)
	return errors.Newf("%s", p.unimplemented)
}

func (p *provider) Delete(vms vm.List) error {
	__antithesis_instrumentation__.Notify(183566)
	return errors.Newf("%s", p.unimplemented)
}

func (p *provider) Reset(vms vm.List) error {
	__antithesis_instrumentation__.Notify(183567)
	return nil
}

func (p *provider) Extend(vms vm.List, lifetime time.Duration) error {
	__antithesis_instrumentation__.Notify(183568)
	return errors.Newf("%s", p.unimplemented)
}

func (p *provider) FindActiveAccount() (string, error) {
	__antithesis_instrumentation__.Notify(183569)
	return "", nil
}

func (p *provider) List(l *logger.Logger) (vm.List, error) {
	__antithesis_instrumentation__.Notify(183570)
	return nil, nil
}

func (p *provider) Name() string {
	__antithesis_instrumentation__.Notify(183571)
	return p.delegate.Name()
}

func (p *provider) Active() bool {
	__antithesis_instrumentation__.Notify(183572)
	return false
}

func (p *provider) ProjectActive(project string) bool {
	__antithesis_instrumentation__.Notify(183573)
	return false
}

func (p *provider) CreateProviderOpts() vm.ProviderOpts {
	__antithesis_instrumentation__.Notify(183574)
	return nil
}
