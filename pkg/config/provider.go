package config

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type SystemConfigProvider interface {
	GetSystemConfig() *SystemConfig

	RegisterSystemConfigChannel() (_ <-chan struct{}, unregister func())
}

type ConstantSystemConfigProvider struct {
	cfg *SystemConfig
}

func NewConstantSystemConfigProvider(cfg *SystemConfig) *ConstantSystemConfigProvider {
	__antithesis_instrumentation__.Notify(55791)
	p := &ConstantSystemConfigProvider{cfg: cfg}
	return p
}

func (c *ConstantSystemConfigProvider) GetSystemConfig() *SystemConfig {
	__antithesis_instrumentation__.Notify(55792)
	return c.cfg
}

func (c *ConstantSystemConfigProvider) RegisterSystemConfigChannel() (
	_ <-chan struct{},
	unregister func(),
) {
	__antithesis_instrumentation__.Notify(55793)

	return nil, func() { __antithesis_instrumentation__.Notify(55794) }
}
