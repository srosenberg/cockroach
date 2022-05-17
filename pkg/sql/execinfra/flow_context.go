package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
)

type FlowCtx struct {
	AmbientContext log.AmbientContext

	Cfg *ServerConfig

	ID execinfrapb.FlowID

	EvalCtx *tree.EvalContext

	Txn *kv.Txn

	Descriptors *descs.Collection

	IsDescriptorsCleanupRequired bool

	NodeID *base.SQLIDContainer

	TraceKV bool

	CollectStats bool

	Local bool

	Gateway bool

	DiskMonitor *mon.BytesMonitor

	PreserveFlowSpecs bool
}

func (ctx *FlowCtx) NewEvalCtx() *tree.EvalContext {
	__antithesis_instrumentation__.Notify(470996)
	evalCopy := ctx.EvalCtx.Copy()
	return evalCopy
}

func (ctx *FlowCtx) TestingKnobs() TestingKnobs {
	__antithesis_instrumentation__.Notify(470997)
	return ctx.Cfg.TestingKnobs
}

func (ctx *FlowCtx) Stopper() *stop.Stopper {
	__antithesis_instrumentation__.Notify(470998)
	return ctx.Cfg.Stopper
}

func (ctx *FlowCtx) Codec() keys.SQLCodec {
	__antithesis_instrumentation__.Notify(470999)
	return ctx.EvalCtx.Codec
}

func (ctx *FlowCtx) TableDescriptor(desc *descpb.TableDescriptor) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(471000)
	if desc == nil {
		__antithesis_instrumentation__.Notify(471003)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(471004)
	}
	__antithesis_instrumentation__.Notify(471001)
	if ctx != nil && func() bool {
		__antithesis_instrumentation__.Notify(471005)
		return ctx.Descriptors != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(471006)
		return ctx.Txn != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(471007)
		leased, _ := ctx.Descriptors.GetLeasedImmutableTableByID(ctx.EvalCtx.Ctx(), ctx.Txn, desc.ID)
		if leased != nil && func() bool {
			__antithesis_instrumentation__.Notify(471008)
			return leased.GetVersion() == desc.Version == true
		}() == true {
			__antithesis_instrumentation__.Notify(471009)
			return leased
		} else {
			__antithesis_instrumentation__.Notify(471010)
		}
	} else {
		__antithesis_instrumentation__.Notify(471011)
	}
	__antithesis_instrumentation__.Notify(471002)
	return tabledesc.NewUnsafeImmutable(desc)
}

func (ctx *FlowCtx) NewTypeResolver(txn *kv.Txn) descs.DistSQLTypeResolver {
	__antithesis_instrumentation__.Notify(471012)
	if ctx == nil || func() bool {
		__antithesis_instrumentation__.Notify(471014)
		return ctx.Descriptors == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(471015)
		return descs.DistSQLTypeResolver{}
	} else {
		__antithesis_instrumentation__.Notify(471016)
	}
	__antithesis_instrumentation__.Notify(471013)
	return descs.NewDistSQLTypeResolver(ctx.Descriptors, txn)
}

func (ctx *FlowCtx) NewSemaContext(txn *kv.Txn) *tree.SemaContext {
	__antithesis_instrumentation__.Notify(471017)
	resolver := ctx.NewTypeResolver(txn)
	semaCtx := tree.MakeSemaContext()
	semaCtx.TypeResolver = &resolver
	return &semaCtx
}

func (ctx *FlowCtx) ProcessorComponentID(procID int32) execinfrapb.ComponentID {
	__antithesis_instrumentation__.Notify(471018)
	return execinfrapb.ProcessorComponentID(ctx.NodeID.SQLInstanceID(), ctx.ID, procID)
}

func (ctx *FlowCtx) StreamComponentID(streamID execinfrapb.StreamID) execinfrapb.ComponentID {
	__antithesis_instrumentation__.Notify(471019)
	return execinfrapb.StreamComponentID(ctx.NodeID.SQLInstanceID(), ctx.ID, streamID)
}
