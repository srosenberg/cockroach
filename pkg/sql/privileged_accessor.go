package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/errors"
)

func (p *planner) LookupNamespaceID(
	ctx context.Context, parentID int64, parentSchemaID int64, name string,
) (tree.DInt, bool, error) {
	__antithesis_instrumentation__.Notify(563322)
	query := fmt.Sprintf(
		`SELECT id FROM [%d AS namespace] WHERE "parentID" = $1 AND "parentSchemaID" = $2 AND name = $3`,
		keys.NamespaceTableID,
	)
	r, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryRowEx(
		ctx,
		"crdb-internal-get-descriptor-id",
		p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		query,
		parentID,
		parentSchemaID,
		name,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(563326)
		return 0, false, err
	} else {
		__antithesis_instrumentation__.Notify(563327)
	}
	__antithesis_instrumentation__.Notify(563323)
	if r == nil {
		__antithesis_instrumentation__.Notify(563328)
		return 0, false, nil
	} else {
		__antithesis_instrumentation__.Notify(563329)
	}
	__antithesis_instrumentation__.Notify(563324)
	id := tree.MustBeDInt(r[0])
	if err := p.checkDescriptorPermissions(ctx, descpb.ID(id)); err != nil {
		__antithesis_instrumentation__.Notify(563330)
		return 0, false, err
	} else {
		__antithesis_instrumentation__.Notify(563331)
	}
	__antithesis_instrumentation__.Notify(563325)
	return id, true, nil
}

func (p *planner) LookupZoneConfigByNamespaceID(
	ctx context.Context, id int64,
) (tree.DBytes, bool, error) {
	__antithesis_instrumentation__.Notify(563332)
	if err := p.checkDescriptorPermissions(ctx, descpb.ID(id)); err != nil {
		__antithesis_instrumentation__.Notify(563336)
		return "", false, err
	} else {
		__antithesis_instrumentation__.Notify(563337)
	}
	__antithesis_instrumentation__.Notify(563333)

	const query = `SELECT config FROM system.zones WHERE id = $1`
	r, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryRowEx(
		ctx,
		"crdb-internal-get-zone",
		p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		query,
		id,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(563338)
		return "", false, err
	} else {
		__antithesis_instrumentation__.Notify(563339)
	}
	__antithesis_instrumentation__.Notify(563334)
	if r == nil {
		__antithesis_instrumentation__.Notify(563340)
		return "", false, nil
	} else {
		__antithesis_instrumentation__.Notify(563341)
	}
	__antithesis_instrumentation__.Notify(563335)
	return tree.MustBeDBytes(r[0]), true, nil
}

func (p *planner) checkDescriptorPermissions(ctx context.Context, id descpb.ID) error {
	__antithesis_instrumentation__.Notify(563342)
	desc, err := p.Descriptors().GetImmutableDescriptorByID(
		ctx, p.txn, id,
		tree.CommonLookupFlags{
			IncludeDropped: true,
			IncludeOffline: true,

			Required: true,
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(563345)

		if errors.Is(err, catalog.ErrDescriptorNotFound) {
			__antithesis_instrumentation__.Notify(563347)
			err = nil
		} else {
			__antithesis_instrumentation__.Notify(563348)
		}
		__antithesis_instrumentation__.Notify(563346)
		return err
	} else {
		__antithesis_instrumentation__.Notify(563349)
	}
	__antithesis_instrumentation__.Notify(563343)
	if err := p.CheckAnyPrivilege(ctx, desc); err != nil {
		__antithesis_instrumentation__.Notify(563350)
		return pgerror.New(pgcode.InsufficientPrivilege, "insufficient privilege")
	} else {
		__antithesis_instrumentation__.Notify(563351)
	}
	__antithesis_instrumentation__.Notify(563344)
	return nil
}
