package spanconfigmanager

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "context"

func (m *Manager) TestingCreateAndStartJobIfNoneExists(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(240767)
	return m.createAndStartJobIfNoneExists(ctx)
}
