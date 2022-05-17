package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func (p *planner) dropNamespaceEntry(
	ctx context.Context, b *kv.Batch, desc catalog.MutableDescriptor,
) {
	__antithesis_instrumentation__.Notify(501829)

	deleteNamespaceEntryAndMaybeAddDrainingName(ctx, b, p, desc, desc)
}

func (p *planner) renameNamespaceEntry(
	ctx context.Context, b *kv.Batch, oldNameKey catalog.NameKey, desc catalog.MutableDescriptor,
) {
	__antithesis_instrumentation__.Notify(501830)

	deleteNamespaceEntryAndMaybeAddDrainingName(ctx, b, p, oldNameKey, desc)

	marshalledKey := catalogkeys.EncodeNameKey(p.ExecCfg().Codec, desc)
	if p.extendedEvalCtx.Tracing.KVTracingEnabled() {
		__antithesis_instrumentation__.Notify(501832)
		log.VEventf(ctx, 2, "CPut %s -> %d", marshalledKey, desc.GetID())
	} else {
		__antithesis_instrumentation__.Notify(501833)
	}
	__antithesis_instrumentation__.Notify(501831)
	b.CPut(marshalledKey, desc.GetID(), nil)
}

func deleteNamespaceEntryAndMaybeAddDrainingName(
	ctx context.Context,
	b *kv.Batch,
	p *planner,
	nameKeyToDelete catalog.NameKey,
	desc catalog.MutableDescriptor,
) {
	__antithesis_instrumentation__.Notify(501834)
	if !p.execCfg.Settings.Version.IsActive(ctx, clusterversion.AvoidDrainingNames) {
		__antithesis_instrumentation__.Notify(501837)
		desc.AddDrainingName(descpb.NameInfo{
			ParentID:       nameKeyToDelete.GetParentID(),
			ParentSchemaID: nameKeyToDelete.GetParentSchemaID(),
			Name:           nameKeyToDelete.GetName(),
		})
		return
	} else {
		__antithesis_instrumentation__.Notify(501838)
	}
	__antithesis_instrumentation__.Notify(501835)
	marshalledKey := catalogkeys.EncodeNameKey(p.ExecCfg().Codec, nameKeyToDelete)
	if p.extendedEvalCtx.Tracing.KVTracingEnabled() {
		__antithesis_instrumentation__.Notify(501839)
		log.VEventf(ctx, 2, "Del %s", marshalledKey)
	} else {
		__antithesis_instrumentation__.Notify(501840)
	}
	__antithesis_instrumentation__.Notify(501836)
	b.Del(marshalledKey)
}
