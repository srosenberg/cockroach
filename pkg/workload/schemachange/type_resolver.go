package schemachange

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq/oid"
)

type txTypeResolver struct {
	tx pgx.Tx
}

func (t txTypeResolver) ResolveType(
	ctx context.Context, name *tree.UnresolvedObjectName,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(697318)

	if name.HasExplicitSchema() {
		__antithesis_instrumentation__.Notify(697323)
		rows, err := t.tx.Query(ctx, `
  SELECT enumlabel, enumsortorder, pgt.oid::int
    FROM pg_enum AS pge, pg_type AS pgt, pg_namespace AS pgn
   WHERE (pgt.typnamespace = pgn.oid AND pgt.oid = pge.enumtypid)
         AND typcategory = 'E'
         AND typname = $1
				 AND nspname = $2
ORDER BY enumsortorder`, name.Object(), name.Schema())
		if err != nil {
			__antithesis_instrumentation__.Notify(697327)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(697328)
		}
		__antithesis_instrumentation__.Notify(697324)
		var logicalReps []string
		var physicalReps [][]byte
		var readOnly []bool
		var objectID oid.Oid
		for rows.Next() {
			__antithesis_instrumentation__.Notify(697329)
			var logicalRep string
			var order int64
			if err := rows.Scan(&logicalRep, &order, &objectID); err != nil {
				__antithesis_instrumentation__.Notify(697331)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(697332)
			}
			__antithesis_instrumentation__.Notify(697330)
			logicalReps = append(logicalReps, logicalRep)
			physicalReps = append(physicalReps, encoding.EncodeUntaggedIntValue(nil, order))
			readOnly = append(readOnly, false)
		}
		__antithesis_instrumentation__.Notify(697325)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(697333)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(697334)
		}
		__antithesis_instrumentation__.Notify(697326)

		n := types.UserDefinedTypeName{Name: name.Object()}
		n.Schema = name.Schema()
		n.ExplicitSchema = true
		return &types.T{
			InternalType: types.InternalType{
				Family: types.EnumFamily,
				Oid:    objectID,
			},
			TypeMeta: types.UserDefinedTypeMetadata{
				Name: &n,
				EnumData: &types.EnumMetadata{
					LogicalRepresentations:  logicalReps,
					PhysicalRepresentations: physicalReps,
					IsMemberReadOnly:        readOnly,
				},
			},
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(697335)
	}
	__antithesis_instrumentation__.Notify(697319)

	var objectID oid.Oid
	if err := t.tx.QueryRow(ctx, `
  SELECT oid::int
    FROM pg_type
   WHERE typname = $1
  `, name.Object(),
	).Scan(&objectID); err != nil {
		__antithesis_instrumentation__.Notify(697336)
		if errors.Is(err, pgx.ErrNoRows) {
			__antithesis_instrumentation__.Notify(697338)
			err = errors.Errorf("failed to resolve primitive type %s", name)
		} else {
			__antithesis_instrumentation__.Notify(697339)
		}
		__antithesis_instrumentation__.Notify(697337)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(697340)
	}
	__antithesis_instrumentation__.Notify(697320)

	if _, exists := types.OidToType[objectID]; !exists {
		__antithesis_instrumentation__.Notify(697341)
		return nil, pgerror.Newf(pgcode.UndefinedObject, "type %s with oid %s does not exist", name.Object(), objectID)
	} else {
		__antithesis_instrumentation__.Notify(697342)
	}
	__antithesis_instrumentation__.Notify(697321)

	if objectID == oid.T_bpchar {
		__antithesis_instrumentation__.Notify(697343)
		t := *types.OidToType[objectID]
		t.InternalType.Width = 1
		return &t, nil
	} else {
		__antithesis_instrumentation__.Notify(697344)
	}
	__antithesis_instrumentation__.Notify(697322)
	return types.OidToType[objectID], nil
}

func (t txTypeResolver) ResolveTypeByOID(ctx context.Context, oid oid.Oid) (*types.T, error) {
	__antithesis_instrumentation__.Notify(697345)
	return nil, pgerror.Newf(pgcode.UndefinedObject, "type %d does not exist", oid)
}

var _ tree.TypeReferenceResolver = (*txTypeResolver)(nil)
