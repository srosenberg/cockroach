package certmgr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
)

type CertManager struct {
	syncutil.RWMutex
	ctx           context.Context
	monitorCancel context.CancelFunc
	certs         map[string]Cert
}

func NewCertManager(ctx context.Context) *CertManager {
	__antithesis_instrumentation__.Notify(186344)
	cm := &CertManager{
		ctx:   ctx,
		certs: make(map[string]Cert),
	}
	return cm
}

func (cm *CertManager) ManageCert(id string, cert Cert) {
	__antithesis_instrumentation__.Notify(186345)
	cm.Lock()
	defer cm.Unlock()
	if len(cm.certs) == 0 {
		__antithesis_instrumentation__.Notify(186347)
		cm.startMonitorLocked()
	} else {
		__antithesis_instrumentation__.Notify(186348)
	}
	__antithesis_instrumentation__.Notify(186346)
	cm.certs[id] = cert
}

func (cm *CertManager) RemoveCert(id string) {
	__antithesis_instrumentation__.Notify(186349)
	cm.Lock()
	defer cm.Unlock()
	delete(cm.certs, id)
	if len(cm.certs) == 0 {
		__antithesis_instrumentation__.Notify(186350)
		cm.stopMonitorLocked()
	} else {
		__antithesis_instrumentation__.Notify(186351)
	}
}

func (cm *CertManager) Cert(id string) Cert {
	__antithesis_instrumentation__.Notify(186352)
	cm.RLock()
	defer cm.RUnlock()
	return cm.certs[id]
}

func (cm *CertManager) startMonitorLocked() {
	__antithesis_instrumentation__.Notify(186353)
	ctx, cancel := context.WithCancel(cm.ctx)
	cm.monitorCancel = cancel
	refresh := sysutil.RefreshSignaledChan()

	go func() {
		__antithesis_instrumentation__.Notify(186354)
		for {
			__antithesis_instrumentation__.Notify(186355)
			select {
			case sig := <-refresh:
				__antithesis_instrumentation__.Notify(186356)
				log.Ops.Infof(ctx, "received signal %q, triggering certificate reload", sig)
				cm.Reload(ctx)
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(186357)
				return
			}
		}
	}()
}

func (cm *CertManager) stopMonitorLocked() {
	__antithesis_instrumentation__.Notify(186358)
	cm.monitorCancel()
	cm.monitorCancel = nil
}

func (cm *CertManager) Reload(ctx context.Context) {
	__antithesis_instrumentation__.Notify(186359)
	cm.RLock()
	defer cm.RUnlock()

	errCount := 0
	for _, cert := range cm.certs {
		__antithesis_instrumentation__.Notify(186361)
		cert.Reload(cm.ctx)
		if cert.Err() != nil {
			__antithesis_instrumentation__.Notify(186363)
			log.Ops.Warningf(ctx, "could not reload cert: %v", cert.Err())
			errCount++
		} else {
			__antithesis_instrumentation__.Notify(186364)
		}
		__antithesis_instrumentation__.Notify(186362)

		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(186365)
			return
		} else {
			__antithesis_instrumentation__.Notify(186366)
		}
	}
	__antithesis_instrumentation__.Notify(186360)
	if errCount > 0 {
		__antithesis_instrumentation__.Notify(186367)
		log.StructuredEvent(cm.ctx, &eventpb.CertsReload{
			Success: false,
			ErrorMessage: fmt.Sprintf(
				"%d certs (out of %d) failed to reload", errCount, len(cm.certs),
			)},
		)
	} else {
		__antithesis_instrumentation__.Notify(186368)
		log.StructuredEvent(cm.ctx, &eventpb.CertsReload{Success: true})
	}
}
