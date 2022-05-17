package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
)

func (p *planner) renameDatabase(
	ctx context.Context, desc *dbdesc.Mutable, newName string, stmt string,
) error {
	__antithesis_instrumentation__.Notify(465239)
	oldNameKey := descpb.NameInfo{
		ParentID:       desc.GetParentID(),
		ParentSchemaID: desc.GetParentSchemaID(),
		Name:           desc.GetName(),
	}

	if dbID, err := p.Descriptors().Direct().LookupDatabaseID(ctx, p.txn, newName); err == nil && func() bool {
		__antithesis_instrumentation__.Notify(465242)
		return dbID != descpb.InvalidID == true
	}() == true {
		__antithesis_instrumentation__.Notify(465243)
		return pgerror.Newf(pgcode.DuplicateDatabase,
			"the new database name %q already exists", newName)
	} else {
		__antithesis_instrumentation__.Notify(465244)
		if err != nil {
			__antithesis_instrumentation__.Notify(465245)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465246)
		}
	}
	__antithesis_instrumentation__.Notify(465240)

	desc.SetName(newName)

	b := p.txn.NewBatch()
	p.renameNamespaceEntry(ctx, b, oldNameKey, desc)

	if err := p.writeNonDropDatabaseChange(ctx, desc, stmt); err != nil {
		__antithesis_instrumentation__.Notify(465247)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465248)
	}
	__antithesis_instrumentation__.Notify(465241)

	return p.txn.Run(ctx, b)
}

func (p *planner) writeNonDropDatabaseChange(
	ctx context.Context, desc *dbdesc.Mutable, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(465249)
	if err := p.createNonDropDatabaseChangeJob(ctx, desc.ID, jobDesc); err != nil {
		__antithesis_instrumentation__.Notify(465252)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465253)
	}
	__antithesis_instrumentation__.Notify(465250)
	b := p.Txn().NewBatch()
	if err := p.writeDatabaseChangeToBatch(ctx, desc, b); err != nil {
		__antithesis_instrumentation__.Notify(465254)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465255)
	}
	__antithesis_instrumentation__.Notify(465251)
	return p.Txn().Run(ctx, b)
}

func (p *planner) writeDatabaseChangeToBatch(
	ctx context.Context, desc *dbdesc.Mutable, b *kv.Batch,
) error {
	__antithesis_instrumentation__.Notify(465256)
	return p.Descriptors().WriteDescToBatch(
		ctx,
		p.extendedEvalCtx.Tracing.KVTracingEnabled(),
		desc,
		b,
	)
}

func (p *planner) forEachMutableTableInDatabase(
	ctx context.Context,
	dbDesc catalog.DatabaseDescriptor,
	fn func(ctx context.Context, scName string, tbDesc *tabledesc.Mutable) error,
) error {
	__antithesis_instrumentation__.Notify(465257)
	all, err := p.Descriptors().GetAllDescriptors(ctx, p.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(465262)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465263)
	}
	__antithesis_instrumentation__.Notify(465258)

	lCtx := newInternalLookupCtx(all.OrderedDescriptors(), dbDesc)
	var droppedRemoved []descpb.ID
	for _, tbID := range lCtx.tbIDs {
		__antithesis_instrumentation__.Notify(465264)
		desc := lCtx.tbDescs[tbID]
		if desc.Dropped() {
			__antithesis_instrumentation__.Notify(465266)
			continue
		} else {
			__antithesis_instrumentation__.Notify(465267)
		}
		__antithesis_instrumentation__.Notify(465265)
		droppedRemoved = append(droppedRemoved, tbID)
	}
	__antithesis_instrumentation__.Notify(465259)
	descs, err := p.Descriptors().GetMutableDescriptorsByID(ctx, p.Txn(), droppedRemoved...)
	if err != nil {
		__antithesis_instrumentation__.Notify(465268)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465269)
	}
	__antithesis_instrumentation__.Notify(465260)
	for _, d := range descs {
		__antithesis_instrumentation__.Notify(465270)
		mutable := d.(*tabledesc.Mutable)
		schemaName, found, err := lCtx.GetSchemaName(ctx, d.GetParentSchemaID(), d.GetParentID(), p.ExecCfg().Settings.Version)
		if err != nil {
			__antithesis_instrumentation__.Notify(465273)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465274)
		}
		__antithesis_instrumentation__.Notify(465271)
		if !found {
			__antithesis_instrumentation__.Notify(465275)
			return errors.AssertionFailedf("schema id %d not found", d.GetParentSchemaID())
		} else {
			__antithesis_instrumentation__.Notify(465276)
		}
		__antithesis_instrumentation__.Notify(465272)
		if err := fn(ctx, schemaName, mutable); err != nil {
			__antithesis_instrumentation__.Notify(465277)
			return err
		} else {
			__antithesis_instrumentation__.Notify(465278)
		}
	}
	__antithesis_instrumentation__.Notify(465261)
	return nil
}
