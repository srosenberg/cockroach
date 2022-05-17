package colexecargs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type MonitorRegistry struct {
	accounts []*mon.BoundAccount
	monitors []*mon.BytesMonitor
}

func (r MonitorRegistry) GetMonitors() []*mon.BytesMonitor {
	__antithesis_instrumentation__.Notify(286141)
	return r.monitors
}

func (r *MonitorRegistry) NewStreamingMemAccount(flowCtx *execinfra.FlowCtx) *mon.BoundAccount {
	__antithesis_instrumentation__.Notify(286142)
	streamingMemAccount := flowCtx.EvalCtx.Mon.MakeBoundAccount()
	r.accounts = append(r.accounts, &streamingMemAccount)
	return &streamingMemAccount
}

func (r MonitorRegistry) getMemMonitorName(opName string, processorID int32, suffix string) string {
	__antithesis_instrumentation__.Notify(286143)

	return opName + "-" + strconv.Itoa(int(processorID)) + "-" + suffix + "-" + strconv.Itoa(len(r.monitors))
}

func (r *MonitorRegistry) CreateMemAccountForSpillStrategy(
	ctx context.Context, flowCtx *execinfra.FlowCtx, opName string, processorID int32,
) (*mon.BoundAccount, string) {
	__antithesis_instrumentation__.Notify(286144)
	monitorName := r.getMemMonitorName(opName, processorID, "limited")
	bufferingOpMemMonitor := execinfra.NewLimitedMonitor(
		ctx, flowCtx.EvalCtx.Mon, flowCtx, monitorName,
	)
	r.monitors = append(r.monitors, bufferingOpMemMonitor)
	bufferingMemAccount := bufferingOpMemMonitor.MakeBoundAccount()
	r.accounts = append(r.accounts, &bufferingMemAccount)
	return &bufferingMemAccount, monitorName
}

func (r *MonitorRegistry) CreateMemAccountForSpillStrategyWithLimit(
	ctx context.Context, flowCtx *execinfra.FlowCtx, limit int64, opName string, processorID int32,
) (*mon.BoundAccount, string) {
	__antithesis_instrumentation__.Notify(286145)
	if flowCtx.Cfg.TestingKnobs.ForceDiskSpill {
		__antithesis_instrumentation__.Notify(286147)
		if limit != 1 {
			__antithesis_instrumentation__.Notify(286148)
			colexecerror.InternalError(errors.AssertionFailedf(
				"expected limit of 1 when forcing disk spilling, got %d", limit,
			))
		} else {
			__antithesis_instrumentation__.Notify(286149)
		}
	} else {
		__antithesis_instrumentation__.Notify(286150)
	}
	__antithesis_instrumentation__.Notify(286146)
	monitorName := r.getMemMonitorName(opName, processorID, "limited")
	bufferingOpMemMonitor := mon.NewMonitorInheritWithLimit(monitorName, limit, flowCtx.EvalCtx.Mon)
	bufferingOpMemMonitor.Start(ctx, flowCtx.EvalCtx.Mon, mon.BoundAccount{})
	r.monitors = append(r.monitors, bufferingOpMemMonitor)
	bufferingMemAccount := bufferingOpMemMonitor.MakeBoundAccount()
	r.accounts = append(r.accounts, &bufferingMemAccount)
	return &bufferingMemAccount, monitorName
}

func (r *MonitorRegistry) CreateUnlimitedMemAccount(
	ctx context.Context, flowCtx *execinfra.FlowCtx, opName string, processorID int32,
) *mon.BoundAccount {
	__antithesis_instrumentation__.Notify(286151)
	monitorName := r.getMemMonitorName(opName, processorID, "unlimited")
	bufferingOpUnlimitedMemMonitor := execinfra.NewMonitor(
		ctx, flowCtx.EvalCtx.Mon, monitorName,
	)
	r.monitors = append(r.monitors, bufferingOpUnlimitedMemMonitor)
	bufferingMemAccount := bufferingOpUnlimitedMemMonitor.MakeBoundAccount()
	r.accounts = append(r.accounts, &bufferingMemAccount)
	return &bufferingMemAccount
}

func (r *MonitorRegistry) CreateUnlimitedMemAccounts(
	ctx context.Context, flowCtx *execinfra.FlowCtx, name string, numAccounts int,
) (*mon.BytesMonitor, []*mon.BoundAccount) {
	__antithesis_instrumentation__.Notify(286152)
	bufferingOpUnlimitedMemMonitor := execinfra.NewMonitor(
		ctx, flowCtx.EvalCtx.Mon, name+"-unlimited",
	)
	r.monitors = append(r.monitors, bufferingOpUnlimitedMemMonitor)
	oldLen := len(r.accounts)
	for i := 0; i < numAccounts; i++ {
		__antithesis_instrumentation__.Notify(286154)
		acc := bufferingOpUnlimitedMemMonitor.MakeBoundAccount()
		r.accounts = append(r.accounts, &acc)
	}
	__antithesis_instrumentation__.Notify(286153)
	return bufferingOpUnlimitedMemMonitor, r.accounts[oldLen:len(r.accounts)]
}

func (r *MonitorRegistry) CreateDiskMonitor(
	ctx context.Context, flowCtx *execinfra.FlowCtx, opName string, processorID int32,
) *mon.BytesMonitor {
	__antithesis_instrumentation__.Notify(286155)
	monitorName := r.getMemMonitorName(opName, processorID, "disk")
	opDiskMonitor := execinfra.NewMonitor(ctx, flowCtx.DiskMonitor, monitorName)
	r.monitors = append(r.monitors, opDiskMonitor)
	return opDiskMonitor
}

func (r *MonitorRegistry) CreateDiskAccount(
	ctx context.Context, flowCtx *execinfra.FlowCtx, opName string, processorID int32,
) *mon.BoundAccount {
	__antithesis_instrumentation__.Notify(286156)
	opDiskMonitor := r.CreateDiskMonitor(ctx, flowCtx, opName, processorID)
	opDiskAccount := opDiskMonitor.MakeBoundAccount()
	r.accounts = append(r.accounts, &opDiskAccount)
	return &opDiskAccount
}

func (r *MonitorRegistry) CreateDiskAccounts(
	ctx context.Context, flowCtx *execinfra.FlowCtx, name string, numAccounts int,
) (*mon.BytesMonitor, []*mon.BoundAccount) {
	__antithesis_instrumentation__.Notify(286157)
	diskMonitor := execinfra.NewMonitor(ctx, flowCtx.DiskMonitor, name)
	r.monitors = append(r.monitors, diskMonitor)
	oldLen := len(r.accounts)
	for i := 0; i < numAccounts; i++ {
		__antithesis_instrumentation__.Notify(286159)
		diskAcc := diskMonitor.MakeBoundAccount()
		r.accounts = append(r.accounts, &diskAcc)
	}
	__antithesis_instrumentation__.Notify(286158)
	return diskMonitor, r.accounts[oldLen:len(r.accounts)]
}

func (r *MonitorRegistry) AssertInvariants() {
	__antithesis_instrumentation__.Notify(286160)

	names := make(map[string]struct{}, len(r.monitors))
	for _, m := range r.monitors {
		__antithesis_instrumentation__.Notify(286161)
		if _, seen := names[m.Name()]; seen {
			__antithesis_instrumentation__.Notify(286163)
			colexecerror.InternalError(errors.AssertionFailedf("monitor named %q encountered twice", m.Name()))
		} else {
			__antithesis_instrumentation__.Notify(286164)
		}
		__antithesis_instrumentation__.Notify(286162)
		names[m.Name()] = struct{}{}
	}
}

func (r *MonitorRegistry) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(286165)
	for i := range r.accounts {
		__antithesis_instrumentation__.Notify(286167)
		r.accounts[i].Close(ctx)
	}
	__antithesis_instrumentation__.Notify(286166)
	for i := range r.monitors {
		__antithesis_instrumentation__.Notify(286168)
		r.monitors[i].Stop(ctx)
	}
}

func (r *MonitorRegistry) Reset() {
	__antithesis_instrumentation__.Notify(286169)

	r.accounts = r.accounts[:0]
	r.monitors = r.monitors[:0]
}
