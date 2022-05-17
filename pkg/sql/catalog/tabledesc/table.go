package tabledesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type ColumnDefDescs struct {
	*tree.ColumnTableDef

	*descpb.ColumnDescriptor

	PrimaryKeyOrUniqueIndexDescriptor *descpb.IndexDescriptor

	DefaultExpr, OnUpdateExpr tree.TypedExpr
}

const MaxBucketAllowed = 2048

func (cdd *ColumnDefDescs) ForEachTypedExpr(fn func(tree.TypedExpr) error) error {
	__antithesis_instrumentation__.Notify(270245)
	if cdd.ColumnTableDef.HasDefaultExpr() {
		__antithesis_instrumentation__.Notify(270248)
		if err := fn(cdd.DefaultExpr); err != nil {
			__antithesis_instrumentation__.Notify(270249)
			return err
		} else {
			__antithesis_instrumentation__.Notify(270250)
		}
	} else {
		__antithesis_instrumentation__.Notify(270251)
	}
	__antithesis_instrumentation__.Notify(270246)
	if cdd.ColumnTableDef.HasOnUpdateExpr() {
		__antithesis_instrumentation__.Notify(270252)
		if err := fn(cdd.OnUpdateExpr); err != nil {
			__antithesis_instrumentation__.Notify(270253)
			return err
		} else {
			__antithesis_instrumentation__.Notify(270254)
		}
	} else {
		__antithesis_instrumentation__.Notify(270255)
	}
	__antithesis_instrumentation__.Notify(270247)
	return nil
}

func MakeColumnDefDescs(
	ctx context.Context, d *tree.ColumnTableDef, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext,
) (*ColumnDefDescs, error) {
	__antithesis_instrumentation__.Notify(270256)
	if d.IsSerial {
		__antithesis_instrumentation__.Notify(270267)

		return nil, pgerror.New(pgcode.FeatureNotSupported,
			"SERIAL cannot be used in this context")
	} else {
		__antithesis_instrumentation__.Notify(270268)
	}
	__antithesis_instrumentation__.Notify(270257)

	if len(d.CheckExprs) > 0 {
		__antithesis_instrumentation__.Notify(270269)

		return nil, errors.New("unexpected column CHECK constraint")
	} else {
		__antithesis_instrumentation__.Notify(270270)
	}
	__antithesis_instrumentation__.Notify(270258)
	if d.HasFKConstraint() {
		__antithesis_instrumentation__.Notify(270271)

		return nil, errors.New("unexpected column REFERENCED constraint")
	} else {
		__antithesis_instrumentation__.Notify(270272)
	}
	__antithesis_instrumentation__.Notify(270259)

	col := &descpb.ColumnDescriptor{
		Name: string(d.Name),
		Nullable: d.Nullable.Nullability != tree.NotNull && func() bool {
			__antithesis_instrumentation__.Notify(270273)
			return !d.PrimaryKey.IsPrimaryKey == true
		}() == true,
		Virtual: d.IsVirtual(),
		Hidden:  d.Hidden,
	}
	ret := &ColumnDefDescs{
		ColumnTableDef:   d,
		ColumnDescriptor: col,
	}

	if d.GeneratedIdentity.IsGeneratedAsIdentity {
		__antithesis_instrumentation__.Notify(270274)
		switch d.GeneratedIdentity.GeneratedAsIdentityType {
		case tree.GeneratedAlways:
			__antithesis_instrumentation__.Notify(270276)
			col.GeneratedAsIdentityType = catpb.GeneratedAsIdentityType_GENERATED_ALWAYS
		case tree.GeneratedByDefault:
			__antithesis_instrumentation__.Notify(270277)
			col.GeneratedAsIdentityType = catpb.GeneratedAsIdentityType_GENERATED_BY_DEFAULT
		default:
			__antithesis_instrumentation__.Notify(270278)
			return nil, errors.AssertionFailedf(
				"column %s is of invalid generated as identity type (neither ALWAYS nor BY DEFAULT)", string(d.Name))
		}
		__antithesis_instrumentation__.Notify(270275)
		if genSeqOpt := d.GeneratedIdentity.SeqOptions; genSeqOpt != nil {
			__antithesis_instrumentation__.Notify(270279)
			s := tree.Serialize(&d.GeneratedIdentity.SeqOptions)
			col.GeneratedAsIdentitySequenceOption = &s
		} else {
			__antithesis_instrumentation__.Notify(270280)
		}
	} else {
		__antithesis_instrumentation__.Notify(270281)
	}
	__antithesis_instrumentation__.Notify(270260)

	resType, err := tree.ResolveType(ctx, d.Type, semaCtx.GetTypeResolver())
	if err != nil {
		__antithesis_instrumentation__.Notify(270282)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(270283)
	}
	__antithesis_instrumentation__.Notify(270261)
	if err = colinfo.ValidateColumnDefType(resType); err != nil {
		__antithesis_instrumentation__.Notify(270284)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(270285)
	}
	__antithesis_instrumentation__.Notify(270262)
	col.Type = resType

	if d.HasDefaultExpr() {
		__antithesis_instrumentation__.Notify(270286)

		ret.DefaultExpr, err = schemaexpr.SanitizeVarFreeExpr(
			ctx, d.DefaultExpr.Expr, resType, "DEFAULT", semaCtx, tree.VolatilityVolatile,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(270288)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(270289)
		}
		__antithesis_instrumentation__.Notify(270287)

		if ret.DefaultExpr != tree.DNull {
			__antithesis_instrumentation__.Notify(270290)
			d.DefaultExpr.Expr = ret.DefaultExpr
			s := tree.Serialize(d.DefaultExpr.Expr)
			col.DefaultExpr = &s
		} else {
			__antithesis_instrumentation__.Notify(270291)
		}
	} else {
		__antithesis_instrumentation__.Notify(270292)
	}
	__antithesis_instrumentation__.Notify(270263)

	if d.HasOnUpdateExpr() {
		__antithesis_instrumentation__.Notify(270293)

		ret.OnUpdateExpr, err = schemaexpr.SanitizeVarFreeExpr(
			ctx, d.OnUpdateExpr.Expr, resType, "ON UPDATE", semaCtx, tree.VolatilityVolatile,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(270295)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(270296)
		}
		__antithesis_instrumentation__.Notify(270294)

		d.OnUpdateExpr.Expr = ret.OnUpdateExpr
		s := tree.Serialize(d.OnUpdateExpr.Expr)
		col.OnUpdateExpr = &s
	} else {
		__antithesis_instrumentation__.Notify(270297)
	}
	__antithesis_instrumentation__.Notify(270264)

	if d.IsComputed() {
		__antithesis_instrumentation__.Notify(270298)

		s := tree.Serialize(d.Computed.Expr)
		col.ComputeExpr = &s
	} else {
		__antithesis_instrumentation__.Notify(270299)
	}
	__antithesis_instrumentation__.Notify(270265)

	if d.PrimaryKey.IsPrimaryKey || func() bool {
		__antithesis_instrumentation__.Notify(270300)
		return (d.Unique.IsUnique && func() bool {
			__antithesis_instrumentation__.Notify(270301)
			return !d.Unique.WithoutIndex == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(270302)
		if !d.PrimaryKey.Sharded {
			__antithesis_instrumentation__.Notify(270304)
			ret.PrimaryKeyOrUniqueIndexDescriptor = &descpb.IndexDescriptor{
				Unique:              true,
				KeyColumnNames:      []string{string(d.Name)},
				KeyColumnDirections: []descpb.IndexDescriptor_Direction{descpb.IndexDescriptor_ASC},
			}
		} else {
			__antithesis_instrumentation__.Notify(270305)
			buckets, err := EvalShardBucketCount(ctx, semaCtx, evalCtx, d.PrimaryKey.ShardBuckets, d.PrimaryKey.StorageParams)
			if err != nil {
				__antithesis_instrumentation__.Notify(270307)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(270308)
			}
			__antithesis_instrumentation__.Notify(270306)
			shardColName := GetShardColumnName([]string{string(d.Name)}, buckets)
			ret.PrimaryKeyOrUniqueIndexDescriptor = &descpb.IndexDescriptor{
				Unique:              true,
				KeyColumnNames:      []string{shardColName, string(d.Name)},
				KeyColumnDirections: []descpb.IndexDescriptor_Direction{descpb.IndexDescriptor_ASC, descpb.IndexDescriptor_ASC},
				Sharded: catpb.ShardedDescriptor{
					IsSharded:    true,
					Name:         shardColName,
					ShardBuckets: buckets,
					ColumnNames:  []string{string(d.Name)},
				},
			}
		}
		__antithesis_instrumentation__.Notify(270303)
		if d.Unique.ConstraintName != "" {
			__antithesis_instrumentation__.Notify(270309)
			ret.PrimaryKeyOrUniqueIndexDescriptor.Name = string(d.Unique.ConstraintName)
		} else {
			__antithesis_instrumentation__.Notify(270310)
		}
	} else {
		__antithesis_instrumentation__.Notify(270311)
	}
	__antithesis_instrumentation__.Notify(270266)

	return ret, nil
}

func EvalShardBucketCount(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	evalCtx *tree.EvalContext,
	shardBuckets tree.Expr,
	storageParams tree.StorageParams,
) (int32, error) {
	__antithesis_instrumentation__.Notify(270312)
	_, legacyBucketNotGiven := shardBuckets.(tree.DefaultVal)
	paramVal := storageParams.GetVal(`bucket_count`)

	if !legacyBucketNotGiven && func() bool {
		__antithesis_instrumentation__.Notify(270317)
		return paramVal != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(270318)
		return 0, pgerror.New(
			pgcode.InvalidParameterValue,
			`"bucket_count" storage parameter and "BUCKET_COUNT" cannot be set at the same time`,
		)
	} else {
		__antithesis_instrumentation__.Notify(270319)
	}
	__antithesis_instrumentation__.Notify(270313)

	var buckets int64
	const invalidBucketCountMsg = `hash sharded index bucket count must be in range [2, 2048], got %v`

	if legacyBucketNotGiven && func() bool {
		__antithesis_instrumentation__.Notify(270320)
		return paramVal == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(270321)
		buckets = catconstants.DefaultHashShardedIndexBucketCount.Get(&evalCtx.Settings.SV)
	} else {
		__antithesis_instrumentation__.Notify(270322)
		if paramVal != nil {
			__antithesis_instrumentation__.Notify(270326)
			shardBuckets = paramVal
		} else {
			__antithesis_instrumentation__.Notify(270327)
		}
		__antithesis_instrumentation__.Notify(270323)
		typedExpr, err := schemaexpr.SanitizeVarFreeExpr(
			ctx, shardBuckets, types.Int, "BUCKET_COUNT", semaCtx, tree.VolatilityVolatile,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(270328)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(270329)
		}
		__antithesis_instrumentation__.Notify(270324)
		d, err := typedExpr.Eval(evalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(270330)
			return 0, pgerror.Wrapf(err, pgcode.InvalidParameterValue, invalidBucketCountMsg, typedExpr)
		} else {
			__antithesis_instrumentation__.Notify(270331)
		}
		__antithesis_instrumentation__.Notify(270325)
		buckets = int64(tree.MustBeDInt(d))
	}
	__antithesis_instrumentation__.Notify(270314)
	if buckets < 2 {
		__antithesis_instrumentation__.Notify(270332)
		return 0, pgerror.Newf(pgcode.InvalidParameterValue, invalidBucketCountMsg, buckets)
	} else {
		__antithesis_instrumentation__.Notify(270333)
	}
	__antithesis_instrumentation__.Notify(270315)
	if buckets > MaxBucketAllowed {
		__antithesis_instrumentation__.Notify(270334)
		return 0, pgerror.Newf(pgcode.InvalidParameterValue, invalidBucketCountMsg, buckets)
	} else {
		__antithesis_instrumentation__.Notify(270335)
	}
	__antithesis_instrumentation__.Notify(270316)
	return int32(buckets), nil
}

func GetShardColumnName(colNames []string, buckets int32) string {
	__antithesis_instrumentation__.Notify(270336)

	sort.Strings(colNames)
	return strings.Join(
		append(append([]string{`crdb_internal`}, colNames...), fmt.Sprintf(`shard_%v`, buckets)), `_`,
	)
}

func (desc *wrapper) GetConstraintInfo() (map[string]descpb.ConstraintDetail, error) {
	__antithesis_instrumentation__.Notify(270337)
	return desc.collectConstraintInfo(nil)
}

func (desc *wrapper) GetConstraintInfoWithLookup(
	tableLookup catalog.TableLookupFn,
) (map[string]descpb.ConstraintDetail, error) {
	__antithesis_instrumentation__.Notify(270338)
	return desc.collectConstraintInfo(tableLookup)
}

func (desc *wrapper) CheckUniqueConstraints() error {
	__antithesis_instrumentation__.Notify(270339)
	_, err := desc.collectConstraintInfo(nil)
	return err
}

func (desc *wrapper) collectConstraintInfo(
	tableLookup catalog.TableLookupFn,
) (map[string]descpb.ConstraintDetail, error) {
	__antithesis_instrumentation__.Notify(270340)
	info := make(map[string]descpb.ConstraintDetail)

	for _, indexI := range desc.NonDropIndexes() {
		__antithesis_instrumentation__.Notify(270345)
		index := indexI.IndexDesc()
		if index.ID == desc.PrimaryIndex.ID {
			__antithesis_instrumentation__.Notify(270346)
			if _, ok := info[index.Name]; ok {
				__antithesis_instrumentation__.Notify(270352)
				return nil, pgerror.Newf(pgcode.DuplicateObject,
					"duplicate constraint name: %q", index.Name)
			} else {
				__antithesis_instrumentation__.Notify(270353)
			}
			__antithesis_instrumentation__.Notify(270347)
			colHiddenMap := make(map[descpb.ColumnID]bool, len(desc.Columns))
			for i := range desc.Columns {
				__antithesis_instrumentation__.Notify(270354)
				col := &desc.Columns[i]
				colHiddenMap[col.ID] = col.Hidden
			}
			__antithesis_instrumentation__.Notify(270348)

			hidden := true
			for _, id := range index.KeyColumnIDs {
				__antithesis_instrumentation__.Notify(270355)
				if !colHiddenMap[id] {
					__antithesis_instrumentation__.Notify(270356)
					hidden = false
					break
				} else {
					__antithesis_instrumentation__.Notify(270357)
				}
			}
			__antithesis_instrumentation__.Notify(270349)
			if hidden {
				__antithesis_instrumentation__.Notify(270358)
				continue
			} else {
				__antithesis_instrumentation__.Notify(270359)
			}
			__antithesis_instrumentation__.Notify(270350)
			indexName := index.Name

			for _, mutation := range desc.GetMutations() {
				__antithesis_instrumentation__.Notify(270360)
				if mutation.GetPrimaryKeySwap() != nil {
					__antithesis_instrumentation__.Notify(270361)
					indexName = mutation.GetPrimaryKeySwap().NewPrimaryIndexName
				} else {
					__antithesis_instrumentation__.Notify(270362)
				}
			}
			__antithesis_instrumentation__.Notify(270351)
			detail := descpb.ConstraintDetail{
				Kind:         descpb.ConstraintTypePK,
				ConstraintID: index.ConstraintID,
			}
			detail.Columns = index.KeyColumnNames
			detail.Index = index
			info[indexName] = detail
		} else {
			__antithesis_instrumentation__.Notify(270363)
			if index.Unique {
				__antithesis_instrumentation__.Notify(270364)
				if _, ok := info[index.Name]; ok {
					__antithesis_instrumentation__.Notify(270366)
					return nil, pgerror.Newf(pgcode.DuplicateObject,
						"duplicate constraint name: %q", index.Name)
				} else {
					__antithesis_instrumentation__.Notify(270367)
				}
				__antithesis_instrumentation__.Notify(270365)
				detail := descpb.ConstraintDetail{
					Kind:         descpb.ConstraintTypeUnique,
					ConstraintID: index.ConstraintID,
				}
				detail.Columns = index.KeyColumnNames
				detail.Index = index
				info[index.Name] = detail
			} else {
				__antithesis_instrumentation__.Notify(270368)
			}
		}
	}
	__antithesis_instrumentation__.Notify(270341)

	ucs := desc.AllActiveAndInactiveUniqueWithoutIndexConstraints()
	for _, uc := range ucs {
		__antithesis_instrumentation__.Notify(270369)
		if _, ok := info[uc.Name]; ok {
			__antithesis_instrumentation__.Notify(270372)
			return nil, pgerror.Newf(pgcode.DuplicateObject,
				"duplicate constraint name: %q", uc.Name)
		} else {
			__antithesis_instrumentation__.Notify(270373)
		}
		__antithesis_instrumentation__.Notify(270370)
		detail := descpb.ConstraintDetail{
			Kind:         descpb.ConstraintTypeUnique,
			ConstraintID: uc.ConstraintID,
		}

		detail.Unvalidated = uc.Validity != descpb.ConstraintValidity_Validated
		var err error
		detail.Columns, err = desc.NamesForColumnIDs(uc.ColumnIDs)
		if err != nil {
			__antithesis_instrumentation__.Notify(270374)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(270375)
		}
		__antithesis_instrumentation__.Notify(270371)
		detail.UniqueWithoutIndexConstraint = uc
		info[uc.Name] = detail
	}
	__antithesis_instrumentation__.Notify(270342)

	fks := desc.AllActiveAndInactiveForeignKeys()
	for _, fk := range fks {
		__antithesis_instrumentation__.Notify(270376)
		if _, ok := info[fk.Name]; ok {
			__antithesis_instrumentation__.Notify(270380)
			return nil, pgerror.Newf(pgcode.DuplicateObject,
				"duplicate constraint name: %q", fk.Name)
		} else {
			__antithesis_instrumentation__.Notify(270381)
		}
		__antithesis_instrumentation__.Notify(270377)
		detail := descpb.ConstraintDetail{
			Kind:         descpb.ConstraintTypeFK,
			ConstraintID: fk.ConstraintID,
		}

		detail.Unvalidated = fk.Validity != descpb.ConstraintValidity_Validated
		var err error
		detail.Columns, err = desc.NamesForColumnIDs(fk.OriginColumnIDs)
		if err != nil {
			__antithesis_instrumentation__.Notify(270382)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(270383)
		}
		__antithesis_instrumentation__.Notify(270378)
		detail.FK = fk

		if tableLookup != nil {
			__antithesis_instrumentation__.Notify(270384)
			other, err := tableLookup(fk.ReferencedTableID)
			if err != nil {
				__antithesis_instrumentation__.Notify(270387)
				return nil, errors.NewAssertionErrorWithWrappedErrf(err,
					"error resolving table %d referenced in foreign key",
					redact.Safe(fk.ReferencedTableID))
			} else {
				__antithesis_instrumentation__.Notify(270388)
			}
			__antithesis_instrumentation__.Notify(270385)
			referencedColumnNames, err := other.NamesForColumnIDs(fk.ReferencedColumnIDs)
			if err != nil {
				__antithesis_instrumentation__.Notify(270389)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(270390)
			}
			__antithesis_instrumentation__.Notify(270386)
			detail.Details = fmt.Sprintf("%s.%v", other.GetName(), referencedColumnNames)
			detail.ReferencedTable = other.TableDesc()
		} else {
			__antithesis_instrumentation__.Notify(270391)
		}
		__antithesis_instrumentation__.Notify(270379)
		info[fk.Name] = detail
	}
	__antithesis_instrumentation__.Notify(270343)

	for _, c := range desc.AllActiveAndInactiveChecks() {
		__antithesis_instrumentation__.Notify(270392)
		if _, ok := info[c.Name]; ok {
			__antithesis_instrumentation__.Notify(270395)
			return nil, pgerror.Newf(pgcode.DuplicateObject,
				"duplicate constraint name: %q", c.Name)
		} else {
			__antithesis_instrumentation__.Notify(270396)
		}
		__antithesis_instrumentation__.Notify(270393)
		detail := descpb.ConstraintDetail{
			Kind:         descpb.ConstraintTypeCheck,
			ConstraintID: c.ConstraintID,
		}

		detail.Unvalidated = c.Validity != descpb.ConstraintValidity_Validated
		detail.CheckConstraint = c
		detail.Details = c.Expr
		if tableLookup != nil {
			__antithesis_instrumentation__.Notify(270397)
			colsUsed, err := desc.ColumnsUsed(c)
			if err != nil {
				__antithesis_instrumentation__.Notify(270399)
				return nil, errors.NewAssertionErrorWithWrappedErrf(err,
					"error computing columns used in check constraint %q", c.Name)
			} else {
				__antithesis_instrumentation__.Notify(270400)
			}
			__antithesis_instrumentation__.Notify(270398)
			for _, colID := range colsUsed {
				__antithesis_instrumentation__.Notify(270401)
				col, err := desc.FindColumnWithID(colID)
				if err != nil {
					__antithesis_instrumentation__.Notify(270403)
					return nil, errors.NewAssertionErrorWithWrappedErrf(err,
						"error finding column %d in table %s", redact.Safe(colID), desc.Name)
				} else {
					__antithesis_instrumentation__.Notify(270404)
				}
				__antithesis_instrumentation__.Notify(270402)
				detail.Columns = append(detail.Columns, col.GetName())
			}
		} else {
			__antithesis_instrumentation__.Notify(270405)
		}
		__antithesis_instrumentation__.Notify(270394)
		info[c.Name] = detail
	}
	__antithesis_instrumentation__.Notify(270344)
	return info, nil
}

func FindFKReferencedUniqueConstraint(
	referencedTable catalog.TableDescriptor, referencedColIDs descpb.ColumnIDs,
) (descpb.UniqueConstraint, error) {
	__antithesis_instrumentation__.Notify(270406)

	primaryIndex := referencedTable.GetPrimaryIndex()
	if primaryIndex.IsValidReferencedUniqueConstraint(referencedColIDs) {
		__antithesis_instrumentation__.Notify(270410)
		return primaryIndex.IndexDesc(), nil
	} else {
		__antithesis_instrumentation__.Notify(270411)
	}
	__antithesis_instrumentation__.Notify(270407)

	for _, idx := range referencedTable.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(270412)
		if idx.IsValidReferencedUniqueConstraint(referencedColIDs) {
			__antithesis_instrumentation__.Notify(270413)
			return idx.IndexDesc(), nil
		} else {
			__antithesis_instrumentation__.Notify(270414)
		}
	}
	__antithesis_instrumentation__.Notify(270408)

	uniqueWithoutIndexConstraints := referencedTable.GetUniqueWithoutIndexConstraints()
	for i := range uniqueWithoutIndexConstraints {
		__antithesis_instrumentation__.Notify(270415)
		c := &uniqueWithoutIndexConstraints[i]

		if c.IsPartial() {
			__antithesis_instrumentation__.Notify(270417)
			continue
		} else {
			__antithesis_instrumentation__.Notify(270418)
		}
		__antithesis_instrumentation__.Notify(270416)

		if c.IsValidReferencedUniqueConstraint(referencedColIDs) {
			__antithesis_instrumentation__.Notify(270419)
			return c, nil
		} else {
			__antithesis_instrumentation__.Notify(270420)
		}
	}
	__antithesis_instrumentation__.Notify(270409)
	return nil, pgerror.Newf(
		pgcode.ForeignKeyViolation,
		"there is no unique constraint matching given keys for referenced table %s",
		referencedTable.GetName(),
	)
}

func InitTableDescriptor(
	id, parentID, parentSchemaID descpb.ID,
	name string,
	creationTime hlc.Timestamp,
	privileges *catpb.PrivilegeDescriptor,
	persistence tree.Persistence,
) Mutable {
	__antithesis_instrumentation__.Notify(270421)
	return Mutable{
		wrapper: wrapper{
			TableDescriptor: descpb.TableDescriptor{
				ID:                      id,
				Name:                    name,
				ParentID:                parentID,
				UnexposedParentSchemaID: parentSchemaID,
				FormatVersion:           descpb.InterleavedFormatVersion,
				Version:                 1,
				ModificationTime:        creationTime,
				Privileges:              privileges,
				CreateAsOfTime:          creationTime,
				Temporary:               persistence.IsTemporary(),
			},
		},
	}
}

func FindPublicColumnsWithNames(
	desc catalog.TableDescriptor, names tree.NameList,
) ([]catalog.Column, error) {
	__antithesis_instrumentation__.Notify(270422)
	cols := make([]catalog.Column, len(names))
	for i, name := range names {
		__antithesis_instrumentation__.Notify(270424)
		c, err := FindPublicColumnWithName(desc, name)
		if err != nil {
			__antithesis_instrumentation__.Notify(270426)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(270427)
		}
		__antithesis_instrumentation__.Notify(270425)
		cols[i] = c
	}
	__antithesis_instrumentation__.Notify(270423)
	return cols, nil
}

func FindPublicColumnWithName(
	desc catalog.TableDescriptor, name tree.Name,
) (catalog.Column, error) {
	__antithesis_instrumentation__.Notify(270428)
	col, err := desc.FindColumnWithName(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(270431)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(270432)
	}
	__antithesis_instrumentation__.Notify(270429)
	if !col.Public() {
		__antithesis_instrumentation__.Notify(270433)
		return nil, colinfo.NewUndefinedColumnError(string(name))
	} else {
		__antithesis_instrumentation__.Notify(270434)
	}
	__antithesis_instrumentation__.Notify(270430)
	return col, nil
}

func FindPublicColumnWithID(
	desc catalog.TableDescriptor, id descpb.ColumnID,
) (catalog.Column, error) {
	__antithesis_instrumentation__.Notify(270435)
	col, err := desc.FindColumnWithID(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(270438)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(270439)
	}
	__antithesis_instrumentation__.Notify(270436)
	if !col.Public() {
		__antithesis_instrumentation__.Notify(270440)
		return nil, fmt.Errorf("column-id \"%d\" does not exist", id)
	} else {
		__antithesis_instrumentation__.Notify(270441)
	}
	__antithesis_instrumentation__.Notify(270437)
	return col, nil
}

func FindInvertedColumn(
	desc catalog.TableDescriptor, invertedColDesc *descpb.ColumnDescriptor,
) catalog.Column {
	__antithesis_instrumentation__.Notify(270442)
	if invertedColDesc == nil {
		__antithesis_instrumentation__.Notify(270445)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(270446)
	}
	__antithesis_instrumentation__.Notify(270443)
	found, err := desc.FindColumnWithID(invertedColDesc.ID)
	if err != nil {
		__antithesis_instrumentation__.Notify(270447)
		panic(errors.HandleAsAssertionFailure(err))
	} else {
		__antithesis_instrumentation__.Notify(270448)
	}
	__antithesis_instrumentation__.Notify(270444)
	invertedColumn := found.DeepCopy()
	*invertedColumn.ColumnDesc() = *invertedColDesc
	return invertedColumn
}

func PrimaryKeyString(desc catalog.TableDescriptor) string {
	__antithesis_instrumentation__.Notify(270449)
	primaryIdx := desc.GetPrimaryIndex()
	f := tree.NewFmtCtx(tree.FmtSimple)
	f.WriteString("PRIMARY KEY (")
	startIdx := primaryIdx.ExplicitColumnStartIdx()
	for i, n := startIdx, primaryIdx.NumKeyColumns(); i < n; i++ {
		__antithesis_instrumentation__.Notify(270452)
		if i > startIdx {
			__antithesis_instrumentation__.Notify(270454)
			f.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(270455)
		}
		__antithesis_instrumentation__.Notify(270453)

		name := primaryIdx.GetKeyColumnName(i)
		f.FormatNameP(&name)
		f.WriteByte(' ')
		f.WriteString(primaryIdx.GetKeyColumnDirection(i).String())
	}
	__antithesis_instrumentation__.Notify(270450)
	f.WriteByte(')')
	if primaryIdx.IsSharded() {
		__antithesis_instrumentation__.Notify(270456)
		f.WriteString(
			fmt.Sprintf(" USING HASH WITH (bucket_count=%v)", primaryIdx.GetSharded().ShardBuckets),
		)
	} else {
		__antithesis_instrumentation__.Notify(270457)
	}
	__antithesis_instrumentation__.Notify(270451)
	return f.CloseAndGetString()
}

func ColumnNamePlaceholder(id descpb.ColumnID) string {
	__antithesis_instrumentation__.Notify(270458)
	return fmt.Sprintf("crdb_internal_column_%d_name_placeholder", id)
}

func IndexNamePlaceholder(id descpb.IndexID) string {
	__antithesis_instrumentation__.Notify(270459)
	return fmt.Sprintf("crdb_internal_index_%d_name_placeholder", id)
}

func RenameColumnInTable(
	tableDesc *Mutable,
	col catalog.Column,
	newName tree.Name,
	isShardColumnRenameable func(shardCol catalog.Column, newShardColName tree.Name) (bool, error),
) error {
	__antithesis_instrumentation__.Notify(270460)
	renameInExpr := func(expr *string) error {
		__antithesis_instrumentation__.Notify(270470)
		newExpr, renameErr := schemaexpr.RenameColumn(*expr, col.ColName(), newName)
		if renameErr != nil {
			__antithesis_instrumentation__.Notify(270472)
			return renameErr
		} else {
			__antithesis_instrumentation__.Notify(270473)
		}
		__antithesis_instrumentation__.Notify(270471)
		*expr = newExpr
		return nil
	}
	__antithesis_instrumentation__.Notify(270461)

	for i := range tableDesc.Checks {
		__antithesis_instrumentation__.Notify(270474)
		if err := renameInExpr(&tableDesc.Checks[i].Expr); err != nil {
			__antithesis_instrumentation__.Notify(270475)
			return err
		} else {
			__antithesis_instrumentation__.Notify(270476)
		}
	}
	__antithesis_instrumentation__.Notify(270462)

	for i := range tableDesc.Columns {
		__antithesis_instrumentation__.Notify(270477)
		if otherCol := &tableDesc.Columns[i]; otherCol.IsComputed() {
			__antithesis_instrumentation__.Notify(270478)
			if err := renameInExpr(otherCol.ComputeExpr); err != nil {
				__antithesis_instrumentation__.Notify(270479)
				return err
			} else {
				__antithesis_instrumentation__.Notify(270480)
			}
		} else {
			__antithesis_instrumentation__.Notify(270481)
		}
	}
	__antithesis_instrumentation__.Notify(270463)

	for _, idx := range tableDesc.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(270482)
		if idx.IsPartial() {
			__antithesis_instrumentation__.Notify(270483)
			if err := renameInExpr(&idx.IndexDesc().Predicate); err != nil {
				__antithesis_instrumentation__.Notify(270484)
				return err
			} else {
				__antithesis_instrumentation__.Notify(270485)
			}
		} else {
			__antithesis_instrumentation__.Notify(270486)
		}
	}
	__antithesis_instrumentation__.Notify(270464)

	for i := range tableDesc.Mutations {
		__antithesis_instrumentation__.Notify(270487)
		m := &tableDesc.Mutations[i]
		if constraint := m.GetConstraint(); constraint != nil {
			__antithesis_instrumentation__.Notify(270488)
			if constraint.ConstraintType == descpb.ConstraintToUpdate_CHECK || func() bool {
				__antithesis_instrumentation__.Notify(270489)
				return constraint.ConstraintType == descpb.ConstraintToUpdate_NOT_NULL == true
			}() == true {
				__antithesis_instrumentation__.Notify(270490)
				if err := renameInExpr(&constraint.Check.Expr); err != nil {
					__antithesis_instrumentation__.Notify(270491)
					return err
				} else {
					__antithesis_instrumentation__.Notify(270492)
				}
			} else {
				__antithesis_instrumentation__.Notify(270493)
			}
		} else {
			__antithesis_instrumentation__.Notify(270494)
			if otherCol := m.GetColumn(); otherCol != nil {
				__antithesis_instrumentation__.Notify(270495)
				if otherCol.IsComputed() {
					__antithesis_instrumentation__.Notify(270496)
					if err := renameInExpr(otherCol.ComputeExpr); err != nil {
						__antithesis_instrumentation__.Notify(270497)
						return err
					} else {
						__antithesis_instrumentation__.Notify(270498)
					}
				} else {
					__antithesis_instrumentation__.Notify(270499)
				}
			} else {
				__antithesis_instrumentation__.Notify(270500)
				if idx := m.GetIndex(); idx != nil {
					__antithesis_instrumentation__.Notify(270501)
					if idx.IsPartial() {
						__antithesis_instrumentation__.Notify(270502)
						if err := renameInExpr(&idx.Predicate); err != nil {
							__antithesis_instrumentation__.Notify(270503)
							return err
						} else {
							__antithesis_instrumentation__.Notify(270504)
						}
					} else {
						__antithesis_instrumentation__.Notify(270505)
					}
				} else {
					__antithesis_instrumentation__.Notify(270506)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(270465)

	shardColumnsToRename := make(map[tree.Name]tree.Name)
	maybeUpdateShardedDesc := func(shardedDesc *catpb.ShardedDescriptor) {
		__antithesis_instrumentation__.Notify(270507)
		if !shardedDesc.IsSharded {
			__antithesis_instrumentation__.Notify(270512)
			return
		} else {
			__antithesis_instrumentation__.Notify(270513)
		}
		__antithesis_instrumentation__.Notify(270508)
		oldShardColName := tree.Name(GetShardColumnName(
			shardedDesc.ColumnNames, shardedDesc.ShardBuckets))
		var changed bool
		for i, c := range shardedDesc.ColumnNames {
			__antithesis_instrumentation__.Notify(270514)
			if c == string(col.ColName()) {
				__antithesis_instrumentation__.Notify(270515)
				changed = true
				shardedDesc.ColumnNames[i] = string(newName)
			} else {
				__antithesis_instrumentation__.Notify(270516)
			}
		}
		__antithesis_instrumentation__.Notify(270509)
		if !changed {
			__antithesis_instrumentation__.Notify(270517)
			return
		} else {
			__antithesis_instrumentation__.Notify(270518)
		}
		__antithesis_instrumentation__.Notify(270510)
		newShardColName, alreadyRenamed := shardColumnsToRename[oldShardColName]
		if !alreadyRenamed {
			__antithesis_instrumentation__.Notify(270519)
			newShardColName = tree.Name(GetShardColumnName(shardedDesc.ColumnNames, shardedDesc.ShardBuckets))
			shardColumnsToRename[oldShardColName] = newShardColName
		} else {
			__antithesis_instrumentation__.Notify(270520)
		}
		__antithesis_instrumentation__.Notify(270511)

		shardedDesc.Name = string(newShardColName)
	}
	__antithesis_instrumentation__.Notify(270466)
	for _, idx := range tableDesc.NonDropIndexes() {
		__antithesis_instrumentation__.Notify(270521)
		maybeUpdateShardedDesc(&idx.IndexDesc().Sharded)
	}
	__antithesis_instrumentation__.Notify(270467)

	if tableDesc.IsLocalityRegionalByRow() {
		__antithesis_instrumentation__.Notify(270522)
		rbrColName, err := tableDesc.GetRegionalByRowTableRegionColumnName()
		if err != nil {
			__antithesis_instrumentation__.Notify(270524)
			return err
		} else {
			__antithesis_instrumentation__.Notify(270525)
		}
		__antithesis_instrumentation__.Notify(270523)
		if rbrColName == col.ColName() {
			__antithesis_instrumentation__.Notify(270526)
			tableDesc.SetTableLocalityRegionalByRow(newName)
		} else {
			__antithesis_instrumentation__.Notify(270527)
		}
	} else {
		__antithesis_instrumentation__.Notify(270528)
	}
	__antithesis_instrumentation__.Notify(270468)

	tableDesc.RenameColumnDescriptor(col, string(newName))

	for oldShardColName, newShardColName := range shardColumnsToRename {
		__antithesis_instrumentation__.Notify(270529)
		shardCol, err := tableDesc.FindColumnWithName(oldShardColName)
		if err != nil {
			__antithesis_instrumentation__.Notify(270533)
			return err
		} else {
			__antithesis_instrumentation__.Notify(270534)
		}
		__antithesis_instrumentation__.Notify(270530)
		var canBeRenamed bool
		if isShardColumnRenameable == nil {
			__antithesis_instrumentation__.Notify(270535)
			canBeRenamed = true
		} else {
			__antithesis_instrumentation__.Notify(270536)
			if canBeRenamed, err = isShardColumnRenameable(shardCol, newShardColName); err != nil {
				__antithesis_instrumentation__.Notify(270537)
				return err
			} else {
				__antithesis_instrumentation__.Notify(270538)
			}
		}
		__antithesis_instrumentation__.Notify(270531)
		if !canBeRenamed {
			__antithesis_instrumentation__.Notify(270539)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(270540)
		}
		__antithesis_instrumentation__.Notify(270532)

		return RenameColumnInTable(tableDesc, shardCol, newShardColName, nil)
	}
	__antithesis_instrumentation__.Notify(270469)

	return nil
}
