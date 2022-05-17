package kvtenantccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errorspb"
)

func (c *Connector) runTenantSettingsSubscription(ctx context.Context, startupCh chan struct{}) {
	__antithesis_instrumentation__.Notify(19826)
	for ctx.Err() == nil {
		__antithesis_instrumentation__.Notify(19827)
		client, err := c.getClient(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(19830)
			continue
		} else {
			__antithesis_instrumentation__.Notify(19831)
		}
		__antithesis_instrumentation__.Notify(19828)
		stream, err := client.TenantSettings(ctx, &roachpb.TenantSettingsRequest{
			TenantID: c.tenantID,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(19832)
			log.Warningf(ctx, "error issuing TenantSettings RPC: %v", err)
			c.tryForgetClient(ctx, client)
			continue
		} else {
			__antithesis_instrumentation__.Notify(19833)
		}
		__antithesis_instrumentation__.Notify(19829)
		for firstEventInStream := true; ; firstEventInStream = false {
			__antithesis_instrumentation__.Notify(19834)
			e, err := stream.Recv()
			if err != nil {
				__antithesis_instrumentation__.Notify(19838)
				if err == io.EOF {
					__antithesis_instrumentation__.Notify(19840)
					break
				} else {
					__antithesis_instrumentation__.Notify(19841)
				}
				__antithesis_instrumentation__.Notify(19839)

				log.Warningf(ctx, "error consuming TenantSettings RPC: %v", err)
				c.tryForgetClient(ctx, client)
				break
			} else {
				__antithesis_instrumentation__.Notify(19842)
			}
			__antithesis_instrumentation__.Notify(19835)
			if e.Error != (errorspb.EncodedError{}) {
				__antithesis_instrumentation__.Notify(19843)

				log.Errorf(ctx, "error consuming TenantSettings RPC: %v", e.Error)
				continue
			} else {
				__antithesis_instrumentation__.Notify(19844)
			}
			__antithesis_instrumentation__.Notify(19836)

			if err := c.processSettingsEvent(e, firstEventInStream); err != nil {
				__antithesis_instrumentation__.Notify(19845)
				log.Errorf(ctx, "error processing tenant settings event: %v", err)
				_ = stream.CloseSend()
				c.tryForgetClient(ctx, client)
				break
			} else {
				__antithesis_instrumentation__.Notify(19846)
			}
			__antithesis_instrumentation__.Notify(19837)

			if startupCh != nil {
				__antithesis_instrumentation__.Notify(19847)
				close(startupCh)
				startupCh = nil
			} else {
				__antithesis_instrumentation__.Notify(19848)
			}
		}
	}
}

func (c *Connector) processSettingsEvent(
	e *roachpb.TenantSettingsEvent, firstEventInStream bool,
) error {
	__antithesis_instrumentation__.Notify(19849)
	if firstEventInStream && func() bool {
		__antithesis_instrumentation__.Notify(19855)
		return e.Incremental == true
	}() == true {
		__antithesis_instrumentation__.Notify(19856)
		return errors.Newf("first event must not be Incremental")
	} else {
		__antithesis_instrumentation__.Notify(19857)
	}
	__antithesis_instrumentation__.Notify(19850)
	c.settingsMu.Lock()
	defer c.settingsMu.Unlock()

	var m map[string]settings.EncodedValue
	switch e.Precedence {
	case roachpb.AllTenantsOverrides:
		__antithesis_instrumentation__.Notify(19858)
		m = c.settingsMu.allTenantOverrides
	case roachpb.SpecificTenantOverrides:
		__antithesis_instrumentation__.Notify(19859)
		m = c.settingsMu.specificOverrides
	default:
		__antithesis_instrumentation__.Notify(19860)
		return errors.Newf("unknown precedence value %d", e.Precedence)
	}
	__antithesis_instrumentation__.Notify(19851)

	if !e.Incremental {
		__antithesis_instrumentation__.Notify(19861)
		for k := range m {
			__antithesis_instrumentation__.Notify(19862)
			delete(m, k)
		}
	} else {
		__antithesis_instrumentation__.Notify(19863)
	}
	__antithesis_instrumentation__.Notify(19852)

	for _, o := range e.Overrides {
		__antithesis_instrumentation__.Notify(19864)
		if o.Value == (settings.EncodedValue{}) {
			__antithesis_instrumentation__.Notify(19865)

			delete(m, o.Name)
		} else {
			__antithesis_instrumentation__.Notify(19866)
			m[o.Name] = o.Value
		}
	}
	__antithesis_instrumentation__.Notify(19853)

	select {
	case c.settingsMu.notifyCh <- struct{}{}:
		__antithesis_instrumentation__.Notify(19867)
	default:
		__antithesis_instrumentation__.Notify(19868)
	}
	__antithesis_instrumentation__.Notify(19854)

	return nil
}

func (c *Connector) RegisterOverridesChannel() <-chan struct{} {
	__antithesis_instrumentation__.Notify(19869)
	c.settingsMu.Lock()
	defer c.settingsMu.Unlock()
	if c.settingsMu.notifyCh != nil {
		__antithesis_instrumentation__.Notify(19871)
		panic(errors.AssertionFailedf("multiple calls not supported"))
	} else {
		__antithesis_instrumentation__.Notify(19872)
	}
	__antithesis_instrumentation__.Notify(19870)
	ch := make(chan struct{}, 1)

	ch <- struct{}{}
	c.settingsMu.notifyCh = ch
	return ch
}

func (c *Connector) Overrides() map[string]settings.EncodedValue {
	__antithesis_instrumentation__.Notify(19873)

	res := make(map[string]settings.EncodedValue)
	c.settingsMu.Lock()
	defer c.settingsMu.Unlock()

	for name, val := range c.settingsMu.allTenantOverrides {
		__antithesis_instrumentation__.Notify(19876)
		res[name] = val
	}
	__antithesis_instrumentation__.Notify(19874)

	for name, val := range c.settingsMu.specificOverrides {
		__antithesis_instrumentation__.Notify(19877)
		res[name] = val
	}
	__antithesis_instrumentation__.Notify(19875)
	return res
}
