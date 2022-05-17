package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

var uniqueRowIDExpr = &tree.FuncExpr{Func: tree.WrapFunction("unique_rowid")}

var unorderedUniqueRowIDExpr = &tree.FuncExpr{Func: tree.WrapFunction("unordered_unique_rowid")}

var realSequenceOpts tree.SequenceOptions

var virtualSequenceOpts = tree.SequenceOptions{
	tree.SequenceOption{Name: tree.SeqOptVirtual},
}

var cachedSequencesCacheSizeSetting = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.defaults.serial_sequences_cache_size",
	"the default cache size when the session's serial normalization mode is set to cached sequences"+
		"A cache size of 1 means no caching. Any cache size less than 1 is invalid.",
	256,
	settings.PositiveInt,
)

func (p *planner) generateSequenceForSerial(
	ctx context.Context, d *tree.ColumnTableDef, tableName *tree.TableName,
) (*tree.TableName, *tree.FuncExpr, catalog.DatabaseDescriptor, catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(617645)
	log.VEventf(ctx, 2, "creating sequence for new column %q of %q", d, tableName)

	seqName := tree.NewTableNameWithSchema(
		tableName.CatalogName,
		tableName.SchemaName,
		tree.Name(tableName.Table()+"_"+string(d.Name)+"_seq"))

	un := seqName.ToUnresolvedObjectName()
	dbDesc, schemaDesc, prefix, err := p.ResolveTargetObject(ctx, un)
	if err != nil {
		__antithesis_instrumentation__.Notify(617648)
		return nil, nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(617649)
	}
	__antithesis_instrumentation__.Notify(617646)
	seqName.ObjectNamePrefix = prefix

	nameBase := seqName.ObjectName
	for i := 0; ; i++ {
		__antithesis_instrumentation__.Notify(617650)
		if i > 0 {
			__antithesis_instrumentation__.Notify(617653)
			seqName.ObjectName = tree.Name(fmt.Sprintf("%s%d", nameBase, i))
		} else {
			__antithesis_instrumentation__.Notify(617654)
		}
		__antithesis_instrumentation__.Notify(617651)
		res, err := p.resolveUncachedTableDescriptor(ctx, seqName, false, tree.ResolveAnyTableKind)
		if err != nil {
			__antithesis_instrumentation__.Notify(617655)
			return nil, nil, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(617656)
		}
		__antithesis_instrumentation__.Notify(617652)
		if res == nil {
			__antithesis_instrumentation__.Notify(617657)
			break
		} else {
			__antithesis_instrumentation__.Notify(617658)
		}
	}
	__antithesis_instrumentation__.Notify(617647)

	defaultExpr := &tree.FuncExpr{
		Func:  tree.WrapFunction("nextval"),
		Exprs: tree.Exprs{tree.NewStrVal(seqName.String())},
	}

	return seqName, defaultExpr, dbDesc, schemaDesc, nil
}

func (p *planner) generateSerialInColumnDef(
	ctx context.Context,
	d *tree.ColumnTableDef,
	tableName *tree.TableName,
	serialNormalizationMode sessiondatapb.SerialNormalizationMode,
) (
	*tree.ColumnTableDef,
	*catalog.ResolvedObjectPrefix,
	*tree.TableName,
	tree.SequenceOptions,
	error,
) {
	__antithesis_instrumentation__.Notify(617659)

	if err := assertValidSerialColumnDef(d, tableName); err != nil {
		__antithesis_instrumentation__.Notify(617666)
		return nil, nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(617667)
	}
	__antithesis_instrumentation__.Notify(617660)

	newSpec := *d

	newSpec.Nullable.Nullability = tree.NotNull

	newSpec.IsSerial = false

	defType, err := tree.ResolveType(ctx, d.Type, p.semaCtx.GetTypeResolver())
	if err != nil {
		__antithesis_instrumentation__.Notify(617668)
		return nil, nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(617669)
	}
	__antithesis_instrumentation__.Notify(617661)

	switch serialNormalizationMode {
	case sessiondatapb.SerialUsesRowID, sessiondatapb.SerialUsesUnorderedRowID, sessiondatapb.SerialUsesVirtualSequences:
		__antithesis_instrumentation__.Notify(617670)

		upgradeType := types.Int
		if defType.Width() < upgradeType.Width() {
			__antithesis_instrumentation__.Notify(617674)
			p.BufferClientNotice(
				ctx,
				errors.WithHintf(
					pgnotice.Newf(
						"upgrading the column %s to %s to utilize the session serial_normalization setting",
						d.Name.String(),
						upgradeType.SQLString(),
					),
					"change the serial_normalization to sql_sequence or sql_sequence_cached if you wish "+
						"to use a smaller sized serial column at the cost of performance. See %s",
					docs.URL("serial.html"),
				),
			)
		} else {
			__antithesis_instrumentation__.Notify(617675)
		}
		__antithesis_instrumentation__.Notify(617671)
		newSpec.Type = upgradeType

	case sessiondatapb.SerialUsesSQLSequences, sessiondatapb.SerialUsesCachedSQLSequences:
		__antithesis_instrumentation__.Notify(617672)

	default:
		__antithesis_instrumentation__.Notify(617673)
		return nil, nil, nil, nil,
			errors.AssertionFailedf("unknown serial normalization mode: %s", serialNormalizationMode)
	}
	__antithesis_instrumentation__.Notify(617662)
	telemetry.Inc(sqltelemetry.SerialColumnNormalizationCounter(
		defType.Name(), serialNormalizationMode.String()))

	if serialNormalizationMode == sessiondatapb.SerialUsesRowID {
		__antithesis_instrumentation__.Notify(617676)

		newSpec.DefaultExpr.Expr = uniqueRowIDExpr
		return &newSpec, nil, nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(617677)
		if serialNormalizationMode == sessiondatapb.SerialUsesUnorderedRowID {
			__antithesis_instrumentation__.Notify(617678)
			newSpec.DefaultExpr.Expr = unorderedUniqueRowIDExpr
			return &newSpec, nil, nil, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(617679)
		}
	}
	__antithesis_instrumentation__.Notify(617663)

	log.VEventf(ctx, 2, "creating sequence for new column %q of %q", d, tableName)

	seqName, defaultExpr, dbDesc, schemaDesc, err := p.generateSequenceForSerial(ctx, d, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(617680)
		return nil, nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(617681)
	}
	__antithesis_instrumentation__.Notify(617664)

	seqType := ""
	seqOpts := realSequenceOpts
	if serialNormalizationMode == sessiondatapb.SerialUsesVirtualSequences {
		__antithesis_instrumentation__.Notify(617682)
		seqType = "virtual "
		seqOpts = virtualSequenceOpts
	} else {
		__antithesis_instrumentation__.Notify(617683)
		if serialNormalizationMode == sessiondatapb.SerialUsesCachedSQLSequences {
			__antithesis_instrumentation__.Notify(617684)
			seqType = "cached "

			value := cachedSequencesCacheSizeSetting.Get(&p.ExecCfg().Settings.SV)
			seqOpts = tree.SequenceOptions{
				tree.SequenceOption{Name: tree.SeqOptCache, IntVal: &value},
			}
		} else {
			__antithesis_instrumentation__.Notify(617685)
		}
	}
	__antithesis_instrumentation__.Notify(617665)
	log.VEventf(ctx, 2, "new column %q of %q will have %s sequence name %q and default %q",
		d, tableName, seqType, seqName, defaultExpr)

	newSpec.DefaultExpr.Expr = defaultExpr

	return &newSpec, &catalog.ResolvedObjectPrefix{
		Database: dbDesc, Schema: schemaDesc,
	}, seqName, seqOpts, nil
}

func (p *planner) processGeneratedAsIdentityColumnDef(
	ctx context.Context, d *tree.ColumnTableDef, tableName *tree.TableName,
) (
	*tree.ColumnTableDef,
	*catalog.ResolvedObjectPrefix,
	*tree.TableName,
	tree.SequenceOptions,
	error,
) {
	__antithesis_instrumentation__.Notify(617686)

	curSerialNormalizationMode := sessiondatapb.SerialUsesSQLSequences
	newSpecPtr, catalogPrefixPtr, seqName, seqOpts, err := p.generateSerialInColumnDef(ctx, d, tableName, curSerialNormalizationMode)
	if err != nil {
		__antithesis_instrumentation__.Notify(617688)
		return nil, nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(617689)
	}
	__antithesis_instrumentation__.Notify(617687)
	return newSpecPtr, catalogPrefixPtr, seqName, seqOpts, nil
}

func (p *planner) processSerialInColumnDef(
	ctx context.Context, d *tree.ColumnTableDef, tableName *tree.TableName,
) (
	*tree.ColumnTableDef,
	*catalog.ResolvedObjectPrefix,
	*tree.TableName,
	tree.SequenceOptions,
	error,
) {
	__antithesis_instrumentation__.Notify(617690)
	serialNormalizationMode := p.SessionData().SerialNormalizationMode
	newSpecPtr, catalogPrefixPtr, seqName, seqOpts, err := p.generateSerialInColumnDef(ctx, d, tableName, serialNormalizationMode)
	if err != nil {
		__antithesis_instrumentation__.Notify(617692)
		return nil, nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(617693)
	}
	__antithesis_instrumentation__.Notify(617691)
	return newSpecPtr, catalogPrefixPtr, seqName, seqOpts, nil
}

func (p *planner) processSerialLikeInColumnDef(
	ctx context.Context, d *tree.ColumnTableDef, tableName *tree.TableName,
) (
	*tree.ColumnTableDef,
	*catalog.ResolvedObjectPrefix,
	*tree.TableName,
	tree.SequenceOptions,
	error,
) {
	__antithesis_instrumentation__.Notify(617694)
	var newSpecPtr *tree.ColumnTableDef
	var catalogPrefixPtr *catalog.ResolvedObjectPrefix
	var seqName *tree.TableName
	var seqOpts tree.SequenceOptions
	var err error

	if d.IsSerial {
		__antithesis_instrumentation__.Notify(617696)
		newSpecPtr, catalogPrefixPtr, seqName, seqOpts, err = p.processSerialInColumnDef(ctx, d, tableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(617697)
			return nil, nil, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(617698)
		}

	} else {
		__antithesis_instrumentation__.Notify(617699)
		if d.GeneratedIdentity.IsGeneratedAsIdentity {
			__antithesis_instrumentation__.Notify(617700)
			newSpecPtr, catalogPrefixPtr, seqName, seqOpts, err = p.processGeneratedAsIdentityColumnDef(ctx, d, tableName)
			if d.GeneratedIdentity.SeqOptions != nil {
				__antithesis_instrumentation__.Notify(617702)
				seqOpts = d.GeneratedIdentity.SeqOptions
			} else {
				__antithesis_instrumentation__.Notify(617703)
			}
			__antithesis_instrumentation__.Notify(617701)
			if err != nil {
				__antithesis_instrumentation__.Notify(617704)
				return nil, nil, nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(617705)
			}
		} else {
			__antithesis_instrumentation__.Notify(617706)
			return d, nil, nil, nil, nil
		}
	}
	__antithesis_instrumentation__.Notify(617695)
	return newSpecPtr, catalogPrefixPtr, seqName, seqOpts, nil
}

func SimplifySerialInColumnDefWithRowID(
	ctx context.Context, d *tree.ColumnTableDef, tableName *tree.TableName,
) error {
	__antithesis_instrumentation__.Notify(617707)
	if !d.IsSerial {
		__antithesis_instrumentation__.Notify(617710)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(617711)
	}
	__antithesis_instrumentation__.Notify(617708)

	if err := assertValidSerialColumnDef(d, tableName); err != nil {
		__antithesis_instrumentation__.Notify(617712)
		return err
	} else {
		__antithesis_instrumentation__.Notify(617713)
	}
	__antithesis_instrumentation__.Notify(617709)

	d.Nullable.Nullability = tree.NotNull

	d.Type = types.Int
	d.DefaultExpr.Expr = uniqueRowIDExpr

	d.IsSerial = false

	return nil
}

func assertValidSerialColumnDef(d *tree.ColumnTableDef, tableName *tree.TableName) error {
	__antithesis_instrumentation__.Notify(617714)
	if d.HasDefaultExpr() {
		__antithesis_instrumentation__.Notify(617718)

		return pgerror.Newf(pgcode.Syntax,
			"multiple default values specified for column %q of table %q",
			tree.ErrString(&d.Name), tree.ErrString(tableName))
	} else {
		__antithesis_instrumentation__.Notify(617719)
	}
	__antithesis_instrumentation__.Notify(617715)

	if d.Nullable.Nullability == tree.Null {
		__antithesis_instrumentation__.Notify(617720)

		return pgerror.Newf(pgcode.Syntax,
			"conflicting NULL/NOT NULL declarations for column %q of table %q",
			tree.ErrString(&d.Name), tree.ErrString(tableName))
	} else {
		__antithesis_instrumentation__.Notify(617721)
	}
	__antithesis_instrumentation__.Notify(617716)

	if d.Computed.Expr != nil {
		__antithesis_instrumentation__.Notify(617722)

		return pgerror.Newf(pgcode.Syntax,
			"SERIAL column %q of table %q cannot be computed",
			tree.ErrString(&d.Name), tree.ErrString(tableName))
	} else {
		__antithesis_instrumentation__.Notify(617723)
	}
	__antithesis_instrumentation__.Notify(617717)

	return nil
}
