package tabledesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/validate"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type Mutable struct {
	wrapper

	original *immutable
}

const (
	LegacyPrimaryKeyIndexName = "primary"

	SequenceColumnID = 1

	SequenceColumnName = "value"
)

var ErrMissingColumns = errors.New("table must contain at least 1 column")

var ErrMissingPrimaryKey = errors.New("table must contain a primary key")

var UseMVCCCompliantIndexCreation = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.mvcc_compliant_index_creation.enabled",
	"if true, schema changes will use the an index backfiller designed for MVCC-compliant bulk operations",
	true,
)

func (desc *wrapper) DescriptorType() catalog.DescriptorType {
	__antithesis_instrumentation__.Notify(269182)
	return catalog.Table
}

func (desc *Mutable) SetName(name string) {
	__antithesis_instrumentation__.Notify(269183)
	desc.Name = name
}

func (desc *wrapper) IsPartitionAllBy() bool {
	__antithesis_instrumentation__.Notify(269184)
	return desc.PartitionAllBy
}

func (desc *wrapper) GetParentSchemaID() descpb.ID {
	__antithesis_instrumentation__.Notify(269185)
	parentSchemaID := desc.GetUnexposedParentSchemaID()

	if parentSchemaID == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(269187)
		parentSchemaID = keys.PublicSchemaID
	} else {
		__antithesis_instrumentation__.Notify(269188)
	}
	__antithesis_instrumentation__.Notify(269186)
	return parentSchemaID
}

func (desc *wrapper) IndexKeysPerRow(idx catalog.Index) int {
	__antithesis_instrumentation__.Notify(269189)
	if desc.PrimaryIndex.ID == idx.GetID() {
		__antithesis_instrumentation__.Notify(269193)
		return len(desc.Families)
	} else {
		__antithesis_instrumentation__.Notify(269194)
	}
	__antithesis_instrumentation__.Notify(269190)
	if idx.NumSecondaryStoredColumns() == 0 || func() bool {
		__antithesis_instrumentation__.Notify(269195)
		return len(desc.Families) == 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(269196)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(269197)
	}
	__antithesis_instrumentation__.Notify(269191)

	numUsedFamilies := 1
	storedColumnIDs := idx.CollectSecondaryStoredColumnIDs()
	for _, family := range desc.Families[1:] {
		__antithesis_instrumentation__.Notify(269198)
		for _, columnID := range family.ColumnIDs {
			__antithesis_instrumentation__.Notify(269199)
			if storedColumnIDs.Contains(columnID) {
				__antithesis_instrumentation__.Notify(269200)
				numUsedFamilies++
				break
			} else {
				__antithesis_instrumentation__.Notify(269201)
			}
		}
	}
	__antithesis_instrumentation__.Notify(269192)
	return numUsedFamilies
}

func BuildIndexName(tableDesc *Mutable, idx *descpb.IndexDescriptor) (string, error) {
	__antithesis_instrumentation__.Notify(269202)

	segments := make([]string, 0, len(idx.KeyColumnNames)+2)

	segments = append(segments, tableDesc.Name)

	exprCount := 0
	for i, n := idx.ExplicitColumnStartIdx(), len(idx.KeyColumnNames); i < n; i++ {
		__antithesis_instrumentation__.Notify(269207)
		var segmentName string
		col, err := tableDesc.FindColumnWithName(tree.Name(idx.KeyColumnNames[i]))
		if err != nil {
			__antithesis_instrumentation__.Notify(269210)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(269211)
		}
		__antithesis_instrumentation__.Notify(269208)
		if col.IsExpressionIndexColumn() {
			__antithesis_instrumentation__.Notify(269212)
			if exprCount == 0 {
				__antithesis_instrumentation__.Notify(269214)
				segmentName = "expr"
			} else {
				__antithesis_instrumentation__.Notify(269215)
				segmentName = fmt.Sprintf("expr%d", exprCount)
			}
			__antithesis_instrumentation__.Notify(269213)
			exprCount++
		} else {
			__antithesis_instrumentation__.Notify(269216)
			segmentName = idx.KeyColumnNames[i]
		}
		__antithesis_instrumentation__.Notify(269209)
		segments = append(segments, segmentName)
	}
	__antithesis_instrumentation__.Notify(269203)

	if idx.UseDeletePreservingEncoding {
		__antithesis_instrumentation__.Notify(269217)
		segments = append(segments, "crdb_internal_dpe")
	} else {
		__antithesis_instrumentation__.Notify(269218)
	}
	__antithesis_instrumentation__.Notify(269204)

	if idx.Unique {
		__antithesis_instrumentation__.Notify(269219)
		segments = append(segments, "key")
	} else {
		__antithesis_instrumentation__.Notify(269220)
		segments = append(segments, "idx")
	}
	__antithesis_instrumentation__.Notify(269205)

	baseName := strings.Join(segments, "_")
	name := baseName
	for i := 1; ; i++ {
		__antithesis_instrumentation__.Notify(269221)
		foundIndex, _ := tableDesc.FindIndexWithName(name)
		if foundIndex == nil {
			__antithesis_instrumentation__.Notify(269223)
			break
		} else {
			__antithesis_instrumentation__.Notify(269224)
		}
		__antithesis_instrumentation__.Notify(269222)
		name = fmt.Sprintf("%s%d", baseName, i)
	}
	__antithesis_instrumentation__.Notify(269206)

	return name, nil
}

func (desc *wrapper) AllActiveAndInactiveChecks() []*descpb.TableDescriptor_CheckConstraint {
	__antithesis_instrumentation__.Notify(269225)

	checks := make([]*descpb.TableDescriptor_CheckConstraint, 0, len(desc.Checks))
	for _, c := range desc.Checks {
		__antithesis_instrumentation__.Notify(269228)

		if c.Validity != descpb.ConstraintValidity_Validating && func() bool {
			__antithesis_instrumentation__.Notify(269229)
			return c.Validity != descpb.ConstraintValidity_Dropping == true
		}() == true {
			__antithesis_instrumentation__.Notify(269230)
			checks = append(checks, c)
		} else {
			__antithesis_instrumentation__.Notify(269231)
		}
	}
	__antithesis_instrumentation__.Notify(269226)
	for _, m := range desc.Mutations {
		__antithesis_instrumentation__.Notify(269232)
		if c := m.GetConstraint(); c != nil && func() bool {
			__antithesis_instrumentation__.Notify(269233)
			return c.ConstraintType == descpb.ConstraintToUpdate_CHECK == true
		}() == true {
			__antithesis_instrumentation__.Notify(269234)

			if m.Direction != descpb.DescriptorMutation_DROP {
				__antithesis_instrumentation__.Notify(269235)
				checks = append(checks, &c.Check)
			} else {
				__antithesis_instrumentation__.Notify(269236)
			}
		} else {
			__antithesis_instrumentation__.Notify(269237)
		}
	}
	__antithesis_instrumentation__.Notify(269227)
	return checks
}

func GetColumnFamilyForShard(desc *Mutable, idxColumns []string) string {
	__antithesis_instrumentation__.Notify(269238)
	for _, f := range desc.Families {
		__antithesis_instrumentation__.Notify(269240)
		for _, fCol := range f.ColumnNames {
			__antithesis_instrumentation__.Notify(269241)
			if fCol == idxColumns[0] {
				__antithesis_instrumentation__.Notify(269242)
				return f.Name
			} else {
				__antithesis_instrumentation__.Notify(269243)
			}
		}
	}
	__antithesis_instrumentation__.Notify(269239)
	return ""
}

func (desc *wrapper) AllActiveAndInactiveUniqueWithoutIndexConstraints() []*descpb.UniqueWithoutIndexConstraint {
	__antithesis_instrumentation__.Notify(269244)
	ucs := make([]*descpb.UniqueWithoutIndexConstraint, 0, len(desc.UniqueWithoutIndexConstraints))
	for i := range desc.UniqueWithoutIndexConstraints {
		__antithesis_instrumentation__.Notify(269247)
		uc := &desc.UniqueWithoutIndexConstraints[i]

		if uc.Validity != descpb.ConstraintValidity_Validating && func() bool {
			__antithesis_instrumentation__.Notify(269248)
			return uc.Validity != descpb.ConstraintValidity_Dropping == true
		}() == true {
			__antithesis_instrumentation__.Notify(269249)
			ucs = append(ucs, uc)
		} else {
			__antithesis_instrumentation__.Notify(269250)
		}
	}
	__antithesis_instrumentation__.Notify(269245)
	for i := range desc.Mutations {
		__antithesis_instrumentation__.Notify(269251)
		if c := desc.Mutations[i].GetConstraint(); c != nil && func() bool {
			__antithesis_instrumentation__.Notify(269252)
			return c.ConstraintType == descpb.ConstraintToUpdate_UNIQUE_WITHOUT_INDEX == true
		}() == true {
			__antithesis_instrumentation__.Notify(269253)
			ucs = append(ucs, &c.UniqueWithoutIndexConstraint)
		} else {
			__antithesis_instrumentation__.Notify(269254)
		}
	}
	__antithesis_instrumentation__.Notify(269246)
	return ucs
}

func (desc *wrapper) AllActiveAndInactiveForeignKeys() []*descpb.ForeignKeyConstraint {
	__antithesis_instrumentation__.Notify(269255)
	fks := make([]*descpb.ForeignKeyConstraint, 0, len(desc.OutboundFKs))
	for i := range desc.OutboundFKs {
		__antithesis_instrumentation__.Notify(269258)
		fk := &desc.OutboundFKs[i]

		if fk.Validity != descpb.ConstraintValidity_Validating && func() bool {
			__antithesis_instrumentation__.Notify(269259)
			return fk.Validity != descpb.ConstraintValidity_Dropping == true
		}() == true {
			__antithesis_instrumentation__.Notify(269260)
			fks = append(fks, fk)
		} else {
			__antithesis_instrumentation__.Notify(269261)
		}
	}
	__antithesis_instrumentation__.Notify(269256)
	for i := range desc.Mutations {
		__antithesis_instrumentation__.Notify(269262)
		if c := desc.Mutations[i].GetConstraint(); c != nil && func() bool {
			__antithesis_instrumentation__.Notify(269263)
			return c.ConstraintType == descpb.ConstraintToUpdate_FOREIGN_KEY == true
		}() == true {
			__antithesis_instrumentation__.Notify(269264)
			fks = append(fks, &c.ForeignKey)
		} else {
			__antithesis_instrumentation__.Notify(269265)
		}
	}
	__antithesis_instrumentation__.Notify(269257)
	return fks
}

func (desc *wrapper) ForeachDependedOnBy(
	f func(dep *descpb.TableDescriptor_Reference) error,
) error {
	__antithesis_instrumentation__.Notify(269266)
	for i := range desc.DependedOnBy {
		__antithesis_instrumentation__.Notify(269268)
		if err := f(&desc.DependedOnBy[i]); err != nil {
			__antithesis_instrumentation__.Notify(269269)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(269271)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(269272)
			}
			__antithesis_instrumentation__.Notify(269270)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269273)
		}
	}
	__antithesis_instrumentation__.Notify(269267)
	return nil
}

func (desc *wrapper) ForeachOutboundFK(
	f func(constraint *descpb.ForeignKeyConstraint) error,
) error {
	__antithesis_instrumentation__.Notify(269274)
	for i := range desc.OutboundFKs {
		__antithesis_instrumentation__.Notify(269276)
		if err := f(&desc.OutboundFKs[i]); err != nil {
			__antithesis_instrumentation__.Notify(269277)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(269279)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(269280)
			}
			__antithesis_instrumentation__.Notify(269278)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269281)
		}
	}
	__antithesis_instrumentation__.Notify(269275)
	return nil
}

func (desc *wrapper) ForeachInboundFK(f func(fk *descpb.ForeignKeyConstraint) error) error {
	__antithesis_instrumentation__.Notify(269282)
	for i := range desc.InboundFKs {
		__antithesis_instrumentation__.Notify(269284)
		if err := f(&desc.InboundFKs[i]); err != nil {
			__antithesis_instrumentation__.Notify(269285)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(269287)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(269288)
			}
			__antithesis_instrumentation__.Notify(269286)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269289)
		}
	}
	__antithesis_instrumentation__.Notify(269283)
	return nil
}

func (desc *wrapper) NumFamilies() int {
	__antithesis_instrumentation__.Notify(269290)
	return len(desc.Families)
}

func (desc *wrapper) ForeachFamily(f func(family *descpb.ColumnFamilyDescriptor) error) error {
	__antithesis_instrumentation__.Notify(269291)
	for i := range desc.Families {
		__antithesis_instrumentation__.Notify(269293)
		if err := f(&desc.Families[i]); err != nil {
			__antithesis_instrumentation__.Notify(269294)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(269296)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(269297)
			}
			__antithesis_instrumentation__.Notify(269295)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269298)
		}
	}
	__antithesis_instrumentation__.Notify(269292)
	return nil
}

func generatedFamilyName(familyID descpb.FamilyID, columnNames []string) string {
	__antithesis_instrumentation__.Notify(269299)
	var buf strings.Builder
	fmt.Fprintf(&buf, "fam_%d", familyID)
	for _, n := range columnNames {
		__antithesis_instrumentation__.Notify(269301)
		buf.WriteString(`_`)
		buf.WriteString(n)
	}
	__antithesis_instrumentation__.Notify(269300)
	return buf.String()
}

func ForEachExprStringInTableDesc(descI catalog.TableDescriptor, f func(expr *string) error) error {
	__antithesis_instrumentation__.Notify(269302)
	var desc *wrapper
	switch descV := descI.(type) {
	case *wrapper:
		__antithesis_instrumentation__.Notify(269311)
		desc = descV
	case *immutable:
		__antithesis_instrumentation__.Notify(269312)
		desc = &descV.wrapper
	case *Mutable:
		__antithesis_instrumentation__.Notify(269313)
		desc = &descV.wrapper
	default:
		__antithesis_instrumentation__.Notify(269314)
		return errors.AssertionFailedf("unexpected type of table %T", descI)
	}
	__antithesis_instrumentation__.Notify(269303)

	doCol := func(c *descpb.ColumnDescriptor) error {
		__antithesis_instrumentation__.Notify(269315)
		if c.HasDefault() {
			__antithesis_instrumentation__.Notify(269319)
			if err := f(c.DefaultExpr); err != nil {
				__antithesis_instrumentation__.Notify(269320)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269321)
			}
		} else {
			__antithesis_instrumentation__.Notify(269322)
		}
		__antithesis_instrumentation__.Notify(269316)
		if c.IsComputed() {
			__antithesis_instrumentation__.Notify(269323)
			if err := f(c.ComputeExpr); err != nil {
				__antithesis_instrumentation__.Notify(269324)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269325)
			}
		} else {
			__antithesis_instrumentation__.Notify(269326)
		}
		__antithesis_instrumentation__.Notify(269317)
		if c.HasOnUpdate() {
			__antithesis_instrumentation__.Notify(269327)
			if err := f(c.OnUpdateExpr); err != nil {
				__antithesis_instrumentation__.Notify(269328)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269329)
			}
		} else {
			__antithesis_instrumentation__.Notify(269330)
		}
		__antithesis_instrumentation__.Notify(269318)
		return nil
	}
	__antithesis_instrumentation__.Notify(269304)
	doIndex := func(i catalog.Index) error {
		__antithesis_instrumentation__.Notify(269331)
		if i.IsPartial() {
			__antithesis_instrumentation__.Notify(269333)
			return f(&i.IndexDesc().Predicate)
		} else {
			__antithesis_instrumentation__.Notify(269334)
		}
		__antithesis_instrumentation__.Notify(269332)
		return nil
	}
	__antithesis_instrumentation__.Notify(269305)
	doCheck := func(c *descpb.TableDescriptor_CheckConstraint) error {
		__antithesis_instrumentation__.Notify(269335)
		return f(&c.Expr)
	}
	__antithesis_instrumentation__.Notify(269306)

	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(269336)
		if err := doCol(&desc.Columns[i]); err != nil {
			__antithesis_instrumentation__.Notify(269337)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269338)
		}
	}
	__antithesis_instrumentation__.Notify(269307)

	if err := catalog.ForEachIndex(descI, catalog.IndexOpts{
		NonPhysicalPrimaryIndex: true,
		DropMutations:           true,
		AddMutations:            true,
	}, doIndex); err != nil {
		__antithesis_instrumentation__.Notify(269339)
		return err
	} else {
		__antithesis_instrumentation__.Notify(269340)
	}
	__antithesis_instrumentation__.Notify(269308)

	for i := range desc.Checks {
		__antithesis_instrumentation__.Notify(269341)
		if err := doCheck(desc.Checks[i]); err != nil {
			__antithesis_instrumentation__.Notify(269342)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269343)
		}
	}
	__antithesis_instrumentation__.Notify(269309)

	for _, mut := range desc.Mutations {
		__antithesis_instrumentation__.Notify(269344)
		if c := mut.GetColumn(); c != nil {
			__antithesis_instrumentation__.Notify(269346)
			if err := doCol(c); err != nil {
				__antithesis_instrumentation__.Notify(269347)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269348)
			}
		} else {
			__antithesis_instrumentation__.Notify(269349)
		}
		__antithesis_instrumentation__.Notify(269345)
		if c := mut.GetConstraint(); c != nil && func() bool {
			__antithesis_instrumentation__.Notify(269350)
			return c.ConstraintType == descpb.ConstraintToUpdate_CHECK == true
		}() == true {
			__antithesis_instrumentation__.Notify(269351)
			if err := doCheck(&c.Check); err != nil {
				__antithesis_instrumentation__.Notify(269352)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269353)
			}
		} else {
			__antithesis_instrumentation__.Notify(269354)
		}
	}
	__antithesis_instrumentation__.Notify(269310)
	return nil
}

func (desc *wrapper) GetAllReferencedTypeIDs(
	dbDesc catalog.DatabaseDescriptor, getType func(descpb.ID) (catalog.TypeDescriptor, error),
) (referencedAnywhere, referencedInColumns descpb.IDs, _ error) {
	__antithesis_instrumentation__.Notify(269355)
	ids, err := desc.getAllReferencedTypesInTableColumns(getType)
	if err != nil {
		__antithesis_instrumentation__.Notify(269361)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(269362)
	}
	__antithesis_instrumentation__.Notify(269356)
	referencedInColumns = make(descpb.IDs, 0, len(ids))
	for id := range ids {
		__antithesis_instrumentation__.Notify(269363)
		referencedInColumns = append(referencedInColumns, id)
	}
	__antithesis_instrumentation__.Notify(269357)
	sort.Sort(referencedInColumns)

	exists := desc.GetMultiRegionEnumDependencyIfExists()
	if exists {
		__antithesis_instrumentation__.Notify(269364)
		regionEnumID, err := dbDesc.MultiRegionEnumID()
		if err != nil {
			__antithesis_instrumentation__.Notify(269366)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(269367)
		}
		__antithesis_instrumentation__.Notify(269365)
		ids[regionEnumID] = struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(269368)
	}
	__antithesis_instrumentation__.Notify(269358)

	for _, id := range desc.DependsOnTypes {
		__antithesis_instrumentation__.Notify(269369)
		ids[id] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(269359)

	result := make(descpb.IDs, 0, len(ids))
	for id := range ids {
		__antithesis_instrumentation__.Notify(269370)
		result = append(result, id)
	}
	__antithesis_instrumentation__.Notify(269360)

	sort.Sort(result)
	return result, referencedInColumns, nil
}

func (desc *wrapper) getAllReferencedTypesInTableColumns(
	getType func(descpb.ID) (catalog.TypeDescriptor, error),
) (map[descpb.ID]struct{}, error) {
	__antithesis_instrumentation__.Notify(269371)

	visitor := &tree.TypeCollectorVisitor{
		OIDs: make(map[oid.Oid]struct{}),
	}

	addOIDsInExpr := func(exprStr *string) error {
		__antithesis_instrumentation__.Notify(269378)
		expr, err := parser.ParseExpr(*exprStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(269380)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269381)
		}
		__antithesis_instrumentation__.Notify(269379)
		tree.WalkExpr(visitor, expr)
		return nil
	}
	__antithesis_instrumentation__.Notify(269372)

	if err := ForEachExprStringInTableDesc(desc, addOIDsInExpr); err != nil {
		__antithesis_instrumentation__.Notify(269382)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(269383)
	}
	__antithesis_instrumentation__.Notify(269373)

	ids := make(map[descpb.ID]struct{})
	for id := range visitor.OIDs {
		__antithesis_instrumentation__.Notify(269384)
		uid, err := typedesc.UserDefinedTypeOIDToID(id)
		if err != nil {
			__antithesis_instrumentation__.Notify(269388)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(269389)
		}
		__antithesis_instrumentation__.Notify(269385)
		typDesc, err := getType(uid)
		if err != nil {
			__antithesis_instrumentation__.Notify(269390)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(269391)
		}
		__antithesis_instrumentation__.Notify(269386)
		children, err := typDesc.GetIDClosure()
		if err != nil {
			__antithesis_instrumentation__.Notify(269392)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(269393)
		}
		__antithesis_instrumentation__.Notify(269387)
		for child := range children {
			__antithesis_instrumentation__.Notify(269394)
			ids[child] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(269374)

	addIDsInColumn := func(c *descpb.ColumnDescriptor) error {
		__antithesis_instrumentation__.Notify(269395)
		children, err := typedesc.GetTypeDescriptorClosure(c.Type)
		if err != nil {
			__antithesis_instrumentation__.Notify(269398)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269399)
		}
		__antithesis_instrumentation__.Notify(269396)
		for id := range children {
			__antithesis_instrumentation__.Notify(269400)
			ids[id] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(269397)
		return nil
	}
	__antithesis_instrumentation__.Notify(269375)
	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(269401)
		if err := addIDsInColumn(&desc.Columns[i]); err != nil {
			__antithesis_instrumentation__.Notify(269402)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(269403)
		}
	}
	__antithesis_instrumentation__.Notify(269376)
	for _, mut := range desc.Mutations {
		__antithesis_instrumentation__.Notify(269404)
		if c := mut.GetColumn(); c != nil {
			__antithesis_instrumentation__.Notify(269405)
			if err := addIDsInColumn(c); err != nil {
				__antithesis_instrumentation__.Notify(269406)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(269407)
			}
		} else {
			__antithesis_instrumentation__.Notify(269408)
		}
	}
	__antithesis_instrumentation__.Notify(269377)

	return ids, nil
}

func (desc *Mutable) initIDs() {
	__antithesis_instrumentation__.Notify(269409)
	if desc.NextColumnID == 0 {
		__antithesis_instrumentation__.Notify(269413)
		desc.NextColumnID = 1
	} else {
		__antithesis_instrumentation__.Notify(269414)
	}
	__antithesis_instrumentation__.Notify(269410)
	if desc.Version == 0 {
		__antithesis_instrumentation__.Notify(269415)
		desc.Version = 1
	} else {
		__antithesis_instrumentation__.Notify(269416)
	}
	__antithesis_instrumentation__.Notify(269411)
	if desc.NextMutationID == descpb.InvalidMutationID {
		__antithesis_instrumentation__.Notify(269417)
		desc.NextMutationID = 1
	} else {
		__antithesis_instrumentation__.Notify(269418)
	}
	__antithesis_instrumentation__.Notify(269412)
	if desc.NextConstraintID == 0 {
		__antithesis_instrumentation__.Notify(269419)
		desc.NextConstraintID = 1
	} else {
		__antithesis_instrumentation__.Notify(269420)
	}
}

func (desc *Mutable) MaybeFillColumnID(
	c *descpb.ColumnDescriptor, columnNames map[string]descpb.ColumnID,
) {
	__antithesis_instrumentation__.Notify(269421)
	desc.initIDs()

	columnID := c.ID
	if columnID == 0 {
		__antithesis_instrumentation__.Notify(269423)
		columnID = desc.NextColumnID
		desc.NextColumnID++
	} else {
		__antithesis_instrumentation__.Notify(269424)
	}
	__antithesis_instrumentation__.Notify(269422)
	columnNames[c.Name] = columnID
	c.ID = columnID
}

func (desc *Mutable) AllocateIDs(ctx context.Context, version clusterversion.ClusterVersion) error {
	__antithesis_instrumentation__.Notify(269425)
	if err := desc.AllocateIDsWithoutValidation(ctx); err != nil {
		__antithesis_instrumentation__.Notify(269428)
		return err
	} else {
		__antithesis_instrumentation__.Notify(269429)
	}
	__antithesis_instrumentation__.Notify(269426)

	savedID := desc.ID
	if desc.ID == 0 {
		__antithesis_instrumentation__.Notify(269430)
		desc.ID = keys.SystemDatabaseID
	} else {
		__antithesis_instrumentation__.Notify(269431)
	}
	__antithesis_instrumentation__.Notify(269427)
	err := validate.Self(version, desc)
	desc.ID = savedID

	return err
}

func (desc *Mutable) AllocateIDsWithoutValidation(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(269432)

	if desc.IsPhysicalTable() {
		__antithesis_instrumentation__.Notify(269437)
		if err := desc.ensurePrimaryKey(); err != nil {
			__antithesis_instrumentation__.Notify(269438)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269439)
		}
	} else {
		__antithesis_instrumentation__.Notify(269440)
	}
	__antithesis_instrumentation__.Notify(269433)

	desc.initIDs()
	columnNames := map[string]descpb.ColumnID{}
	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(269441)
		desc.MaybeFillColumnID(&desc.Columns[i], columnNames)
	}
	__antithesis_instrumentation__.Notify(269434)
	for _, m := range desc.Mutations {
		__antithesis_instrumentation__.Notify(269442)
		if c := m.GetColumn(); c != nil {
			__antithesis_instrumentation__.Notify(269443)
			desc.MaybeFillColumnID(c, columnNames)
		} else {
			__antithesis_instrumentation__.Notify(269444)
		}
	}
	__antithesis_instrumentation__.Notify(269435)

	if desc.IsTable() || func() bool {
		__antithesis_instrumentation__.Notify(269445)
		return desc.MaterializedView() == true
	}() == true {
		__antithesis_instrumentation__.Notify(269446)
		if err := desc.allocateIndexIDs(columnNames); err != nil {
			__antithesis_instrumentation__.Notify(269448)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269449)
		}
		__antithesis_instrumentation__.Notify(269447)

		if desc.IsPhysicalTable() {
			__antithesis_instrumentation__.Notify(269450)
			desc.allocateColumnFamilyIDs(columnNames)
		} else {
			__antithesis_instrumentation__.Notify(269451)
		}
	} else {
		__antithesis_instrumentation__.Notify(269452)
	}
	__antithesis_instrumentation__.Notify(269436)

	return nil
}

func (desc *Mutable) ensurePrimaryKey() error {
	__antithesis_instrumentation__.Notify(269453)
	if len(desc.PrimaryIndex.KeyColumnNames) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(269455)
		return desc.IsPhysicalTable() == true
	}() == true {
		__antithesis_instrumentation__.Notify(269456)

		nameExists := func(name string) bool {
			__antithesis_instrumentation__.Notify(269458)
			_, err := desc.FindColumnWithName(tree.Name(name))
			return err == nil
		}
		__antithesis_instrumentation__.Notify(269457)
		s := "unique_rowid()"
		col := &descpb.ColumnDescriptor{
			Name:        GenerateUniqueName("rowid", nameExists),
			Type:        types.Int,
			DefaultExpr: &s,
			Hidden:      true,
			Nullable:    false,
		}
		desc.AddColumn(col)
		idx := descpb.IndexDescriptor{
			Unique:              true,
			KeyColumnNames:      []string{col.Name},
			KeyColumnDirections: []descpb.IndexDescriptor_Direction{descpb.IndexDescriptor_ASC},
		}
		if err := desc.AddPrimaryIndex(idx); err != nil {
			__antithesis_instrumentation__.Notify(269459)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269460)
		}
	} else {
		__antithesis_instrumentation__.Notify(269461)
	}
	__antithesis_instrumentation__.Notify(269454)
	return nil
}

func (desc *Mutable) allocateIndexIDs(columnNames map[string]descpb.ColumnID) error {
	__antithesis_instrumentation__.Notify(269462)
	if desc.NextIndexID == 0 {
		__antithesis_instrumentation__.Notify(269469)
		desc.NextIndexID = 1
	} else {
		__antithesis_instrumentation__.Notify(269470)
	}
	__antithesis_instrumentation__.Notify(269463)
	if desc.NextConstraintID == 0 {
		__antithesis_instrumentation__.Notify(269471)
		desc.NextConstraintID = 1
	} else {
		__antithesis_instrumentation__.Notify(269472)
	}
	__antithesis_instrumentation__.Notify(269464)

	err := catalog.ForEachNonPrimaryIndex(desc, func(idx catalog.Index) error {
		__antithesis_instrumentation__.Notify(269473)
		if len(idx.GetName()) == 0 {
			__antithesis_instrumentation__.Notify(269476)
			name, err := BuildIndexName(desc, idx.IndexDesc())
			if err != nil {
				__antithesis_instrumentation__.Notify(269478)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269479)
			}
			__antithesis_instrumentation__.Notify(269477)
			idx.IndexDesc().Name = name
		} else {
			__antithesis_instrumentation__.Notify(269480)
		}
		__antithesis_instrumentation__.Notify(269474)
		if idx.GetConstraintID() == 0 && func() bool {
			__antithesis_instrumentation__.Notify(269481)
			return idx.IsUnique() == true
		}() == true {
			__antithesis_instrumentation__.Notify(269482)
			idx.IndexDesc().ConstraintID = desc.NextConstraintID
			desc.NextConstraintID++
		} else {
			__antithesis_instrumentation__.Notify(269483)
		}
		__antithesis_instrumentation__.Notify(269475)
		return nil
	})
	__antithesis_instrumentation__.Notify(269465)
	if err != nil {
		__antithesis_instrumentation__.Notify(269484)
		return err
	} else {
		__antithesis_instrumentation__.Notify(269485)
	}
	__antithesis_instrumentation__.Notify(269466)

	var compositeColIDs catalog.TableColSet
	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(269486)
		col := &desc.Columns[i]
		if colinfo.CanHaveCompositeKeyEncoding(col.Type) {
			__antithesis_instrumentation__.Notify(269487)
			compositeColIDs.Add(col.ID)
		} else {
			__antithesis_instrumentation__.Notify(269488)
		}
	}
	__antithesis_instrumentation__.Notify(269467)

	primaryColIDs := desc.GetPrimaryIndex().CollectKeyColumnIDs()
	for _, idx := range desc.AllIndexes() {
		__antithesis_instrumentation__.Notify(269489)
		if !idx.Primary() {
			__antithesis_instrumentation__.Notify(269498)
			maybeUpgradeSecondaryIndexFormatVersion(idx.IndexDesc())
		} else {
			__antithesis_instrumentation__.Notify(269499)
		}
		__antithesis_instrumentation__.Notify(269490)
		if idx.Primary() && func() bool {
			__antithesis_instrumentation__.Notify(269500)
			return idx.GetConstraintID() == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(269501)
			idx.IndexDesc().ConstraintID = desc.NextConstraintID
			desc.NextConstraintID++
		} else {
			__antithesis_instrumentation__.Notify(269502)
		}
		__antithesis_instrumentation__.Notify(269491)
		if idx.GetID() == 0 {
			__antithesis_instrumentation__.Notify(269503)
			idx.IndexDesc().ID = desc.NextIndexID
			desc.NextIndexID++
		} else {
			__antithesis_instrumentation__.Notify(269504)
			if !idx.Primary() {
				__antithesis_instrumentation__.Notify(269505)

				continue
			} else {
				__antithesis_instrumentation__.Notify(269506)
				if !idx.CollectPrimaryStoredColumnIDs().Contains(0) {
					__antithesis_instrumentation__.Notify(269507)

					continue
				} else {
					__antithesis_instrumentation__.Notify(269508)
				}
			}
		}
		__antithesis_instrumentation__.Notify(269492)

		for j, colName := range idx.IndexDesc().KeyColumnNames {
			__antithesis_instrumentation__.Notify(269509)
			if len(idx.IndexDesc().KeyColumnIDs) <= j {
				__antithesis_instrumentation__.Notify(269511)
				idx.IndexDesc().KeyColumnIDs = append(idx.IndexDesc().KeyColumnIDs, 0)
			} else {
				__antithesis_instrumentation__.Notify(269512)
			}
			__antithesis_instrumentation__.Notify(269510)
			if idx.IndexDesc().KeyColumnIDs[j] == 0 {
				__antithesis_instrumentation__.Notify(269513)
				idx.IndexDesc().KeyColumnIDs[j] = columnNames[colName]
			} else {
				__antithesis_instrumentation__.Notify(269514)
			}
		}
		__antithesis_instrumentation__.Notify(269493)

		indexHasOldStoredColumns := idx.HasOldStoredColumns()
		idx.IndexDesc().KeySuffixColumnIDs = nil
		idx.IndexDesc().StoreColumnIDs = nil
		idx.IndexDesc().CompositeColumnIDs = nil

		colIDs := idx.CollectKeyColumnIDs()
		var extraColumnIDs []descpb.ColumnID
		for _, primaryColID := range desc.PrimaryIndex.KeyColumnIDs {
			__antithesis_instrumentation__.Notify(269515)
			if !colIDs.Contains(primaryColID) {
				__antithesis_instrumentation__.Notify(269516)
				extraColumnIDs = append(extraColumnIDs, primaryColID)
				colIDs.Add(primaryColID)
			} else {
				__antithesis_instrumentation__.Notify(269517)
			}
		}
		__antithesis_instrumentation__.Notify(269494)
		if idx.GetEncodingType() == descpb.SecondaryIndexEncoding {
			__antithesis_instrumentation__.Notify(269518)
			idx.IndexDesc().KeySuffixColumnIDs = extraColumnIDs
		} else {
			__antithesis_instrumentation__.Notify(269519)
			colIDs = idx.CollectKeyColumnIDs()
		}
		__antithesis_instrumentation__.Notify(269495)

		for _, colName := range idx.IndexDesc().StoreColumnNames {
			__antithesis_instrumentation__.Notify(269520)
			col, err := desc.FindColumnWithName(tree.Name(colName))
			if err != nil {
				__antithesis_instrumentation__.Notify(269525)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269526)
			}
			__antithesis_instrumentation__.Notify(269521)
			if primaryColIDs.Contains(col.GetID()) && func() bool {
				__antithesis_instrumentation__.Notify(269527)
				return idx.GetEncodingType() == descpb.SecondaryIndexEncoding == true
			}() == true {
				__antithesis_instrumentation__.Notify(269528)

				err = pgerror.Newf(pgcode.DuplicateColumn,
					"index %q already contains column %q", idx.GetName(), col.GetName())
				err = errors.WithDetailf(err,
					"column %q is part of the primary index and therefore implicit in all indexes", col.GetName())
				return err
			} else {
				__antithesis_instrumentation__.Notify(269529)
			}
			__antithesis_instrumentation__.Notify(269522)
			if colIDs.Contains(col.GetID()) {
				__antithesis_instrumentation__.Notify(269530)
				return pgerror.Newf(
					pgcode.DuplicateColumn,
					"index %q already contains column %q", idx.GetName(), col.GetName())
			} else {
				__antithesis_instrumentation__.Notify(269531)
			}
			__antithesis_instrumentation__.Notify(269523)
			if indexHasOldStoredColumns {
				__antithesis_instrumentation__.Notify(269532)
				idx.IndexDesc().KeySuffixColumnIDs = append(idx.IndexDesc().KeySuffixColumnIDs, col.GetID())
			} else {
				__antithesis_instrumentation__.Notify(269533)
				idx.IndexDesc().StoreColumnIDs = append(idx.IndexDesc().StoreColumnIDs, col.GetID())
			}
			__antithesis_instrumentation__.Notify(269524)
			colIDs.Add(col.GetID())
		}
		__antithesis_instrumentation__.Notify(269496)

		for _, colID := range idx.IndexDesc().KeyColumnIDs {
			__antithesis_instrumentation__.Notify(269534)
			if compositeColIDs.Contains(colID) {
				__antithesis_instrumentation__.Notify(269535)
				idx.IndexDesc().CompositeColumnIDs = append(idx.IndexDesc().CompositeColumnIDs, colID)
			} else {
				__antithesis_instrumentation__.Notify(269536)
			}
		}
		__antithesis_instrumentation__.Notify(269497)
		for _, colID := range idx.IndexDesc().KeySuffixColumnIDs {
			__antithesis_instrumentation__.Notify(269537)
			if compositeColIDs.Contains(colID) {
				__antithesis_instrumentation__.Notify(269538)
				idx.IndexDesc().CompositeColumnIDs = append(idx.IndexDesc().CompositeColumnIDs, colID)
			} else {
				__antithesis_instrumentation__.Notify(269539)
			}
		}
	}
	__antithesis_instrumentation__.Notify(269468)
	return nil
}

func (desc *Mutable) allocateColumnFamilyIDs(columnNames map[string]descpb.ColumnID) {
	__antithesis_instrumentation__.Notify(269540)
	if desc.NextFamilyID == 0 {
		__antithesis_instrumentation__.Notify(269547)
		if len(desc.Families) == 0 {
			__antithesis_instrumentation__.Notify(269549)
			desc.Families = []descpb.ColumnFamilyDescriptor{
				{ID: 0, Name: "primary"},
			}
		} else {
			__antithesis_instrumentation__.Notify(269550)
		}
		__antithesis_instrumentation__.Notify(269548)
		desc.NextFamilyID = 1
	} else {
		__antithesis_instrumentation__.Notify(269551)
	}
	__antithesis_instrumentation__.Notify(269541)

	var columnsInFamilies catalog.TableColSet
	for i := range desc.Families {
		__antithesis_instrumentation__.Notify(269552)
		family := &desc.Families[i]
		if family.ID == 0 && func() bool {
			__antithesis_instrumentation__.Notify(269555)
			return i != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(269556)
			family.ID = desc.NextFamilyID
			desc.NextFamilyID++
		} else {
			__antithesis_instrumentation__.Notify(269557)
		}
		__antithesis_instrumentation__.Notify(269553)

		for j, colName := range family.ColumnNames {
			__antithesis_instrumentation__.Notify(269558)
			if len(family.ColumnIDs) <= j {
				__antithesis_instrumentation__.Notify(269561)
				family.ColumnIDs = append(family.ColumnIDs, 0)
			} else {
				__antithesis_instrumentation__.Notify(269562)
			}
			__antithesis_instrumentation__.Notify(269559)
			if family.ColumnIDs[j] == 0 {
				__antithesis_instrumentation__.Notify(269563)
				family.ColumnIDs[j] = columnNames[colName]
			} else {
				__antithesis_instrumentation__.Notify(269564)
			}
			__antithesis_instrumentation__.Notify(269560)
			columnsInFamilies.Add(family.ColumnIDs[j])
		}
		__antithesis_instrumentation__.Notify(269554)

		desc.Families[i] = *family
	}
	__antithesis_instrumentation__.Notify(269542)

	var primaryIndexColIDs catalog.TableColSet
	for _, colID := range desc.PrimaryIndex.KeyColumnIDs {
		__antithesis_instrumentation__.Notify(269565)
		primaryIndexColIDs.Add(colID)
	}
	__antithesis_instrumentation__.Notify(269543)

	ensureColumnInFamily := func(col *descpb.ColumnDescriptor) {
		__antithesis_instrumentation__.Notify(269566)
		if col.Virtual {
			__antithesis_instrumentation__.Notify(269571)

			return
		} else {
			__antithesis_instrumentation__.Notify(269572)
		}
		__antithesis_instrumentation__.Notify(269567)
		if columnsInFamilies.Contains(col.ID) {
			__antithesis_instrumentation__.Notify(269573)
			return
		} else {
			__antithesis_instrumentation__.Notify(269574)
		}
		__antithesis_instrumentation__.Notify(269568)
		if primaryIndexColIDs.Contains(col.ID) {
			__antithesis_instrumentation__.Notify(269575)

			desc.Families[0].ColumnNames = append(desc.Families[0].ColumnNames, col.Name)
			desc.Families[0].ColumnIDs = append(desc.Families[0].ColumnIDs, col.ID)
			return
		} else {
			__antithesis_instrumentation__.Notify(269576)
		}
		__antithesis_instrumentation__.Notify(269569)
		var familyID descpb.FamilyID
		if desc.ParentID == keys.SystemDatabaseID {
			__antithesis_instrumentation__.Notify(269577)

			familyID = descpb.FamilyID(col.ID)
			desc.Families = append(desc.Families, descpb.ColumnFamilyDescriptor{
				ID:          familyID,
				ColumnNames: []string{col.Name},
				ColumnIDs:   []descpb.ColumnID{col.ID},
			})
		} else {
			__antithesis_instrumentation__.Notify(269578)
			idx, ok := fitColumnToFamily(desc, *col)
			if !ok {
				__antithesis_instrumentation__.Notify(269580)
				idx = len(desc.Families)
				desc.Families = append(desc.Families, descpb.ColumnFamilyDescriptor{
					ID:          desc.NextFamilyID,
					ColumnNames: []string{},
					ColumnIDs:   []descpb.ColumnID{},
				})
			} else {
				__antithesis_instrumentation__.Notify(269581)
			}
			__antithesis_instrumentation__.Notify(269579)
			familyID = desc.Families[idx].ID
			desc.Families[idx].ColumnNames = append(desc.Families[idx].ColumnNames, col.Name)
			desc.Families[idx].ColumnIDs = append(desc.Families[idx].ColumnIDs, col.ID)
		}
		__antithesis_instrumentation__.Notify(269570)
		if familyID >= desc.NextFamilyID {
			__antithesis_instrumentation__.Notify(269582)
			desc.NextFamilyID = familyID + 1
		} else {
			__antithesis_instrumentation__.Notify(269583)
		}
	}
	__antithesis_instrumentation__.Notify(269544)
	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(269584)
		ensureColumnInFamily(&desc.Columns[i])
	}
	__antithesis_instrumentation__.Notify(269545)
	for _, m := range desc.Mutations {
		__antithesis_instrumentation__.Notify(269585)
		if c := m.GetColumn(); c != nil {
			__antithesis_instrumentation__.Notify(269586)
			ensureColumnInFamily(c)
		} else {
			__antithesis_instrumentation__.Notify(269587)
		}
	}
	__antithesis_instrumentation__.Notify(269546)

	for i := range desc.Families {
		__antithesis_instrumentation__.Notify(269588)
		family := &desc.Families[i]
		if len(family.Name) == 0 {
			__antithesis_instrumentation__.Notify(269591)
			family.Name = generatedFamilyName(family.ID, family.ColumnNames)
		} else {
			__antithesis_instrumentation__.Notify(269592)
		}
		__antithesis_instrumentation__.Notify(269589)

		if family.DefaultColumnID == 0 {
			__antithesis_instrumentation__.Notify(269593)
			defaultColumnID := descpb.ColumnID(0)
			for _, colID := range family.ColumnIDs {
				__antithesis_instrumentation__.Notify(269595)
				if !primaryIndexColIDs.Contains(colID) {
					__antithesis_instrumentation__.Notify(269596)
					if defaultColumnID == 0 {
						__antithesis_instrumentation__.Notify(269597)
						defaultColumnID = colID
					} else {
						__antithesis_instrumentation__.Notify(269598)
						defaultColumnID = descpb.ColumnID(0)
						break
					}
				} else {
					__antithesis_instrumentation__.Notify(269599)
				}
			}
			__antithesis_instrumentation__.Notify(269594)
			family.DefaultColumnID = defaultColumnID
		} else {
			__antithesis_instrumentation__.Notify(269600)
		}
		__antithesis_instrumentation__.Notify(269590)

		desc.Families[i] = *family
	}
}

func fitColumnToFamily(desc *Mutable, col descpb.ColumnDescriptor) (int, bool) {
	__antithesis_instrumentation__.Notify(269601)

	return 0, true
}

func (desc *Mutable) MaybeIncrementVersion() {
	__antithesis_instrumentation__.Notify(269602)

	if desc.Version == desc.ClusterVersion().Version+1 || func() bool {
		__antithesis_instrumentation__.Notify(269604)
		return desc.ClusterVersion().Version == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(269605)
		return
	} else {
		__antithesis_instrumentation__.Notify(269606)
	}
	__antithesis_instrumentation__.Notify(269603)
	desc.Version++

	desc.ModificationTime = hlc.Timestamp{}
}

func (desc *Mutable) OriginalName() string {
	__antithesis_instrumentation__.Notify(269607)
	return desc.ClusterVersion().Name
}

func (desc *Mutable) OriginalID() descpb.ID {
	__antithesis_instrumentation__.Notify(269608)
	return desc.ClusterVersion().ID
}

func (desc *Mutable) OriginalVersion() descpb.DescriptorVersion {
	__antithesis_instrumentation__.Notify(269609)
	return desc.ClusterVersion().Version
}

func (desc *Mutable) ClusterVersion() descpb.TableDescriptor {
	__antithesis_instrumentation__.Notify(269610)
	if desc.original == nil {
		__antithesis_instrumentation__.Notify(269612)
		return descpb.TableDescriptor{}
	} else {
		__antithesis_instrumentation__.Notify(269613)
	}
	__antithesis_instrumentation__.Notify(269611)
	return desc.original.TableDescriptor
}

func (desc *Mutable) OriginalDescriptor() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(269614)
	if desc.original != nil {
		__antithesis_instrumentation__.Notify(269616)
		return desc.original
	} else {
		__antithesis_instrumentation__.Notify(269617)
	}
	__antithesis_instrumentation__.Notify(269615)
	return nil
}

const FamilyHeuristicTargetBytes = 256

func notIndexableError(cols []descpb.ColumnDescriptor) error {
	__antithesis_instrumentation__.Notify(269618)
	if len(cols) == 0 {
		__antithesis_instrumentation__.Notify(269621)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(269622)
	}
	__antithesis_instrumentation__.Notify(269619)
	var msg string
	var typInfo string
	if len(cols) == 1 {
		__antithesis_instrumentation__.Notify(269623)
		col := &cols[0]
		msg = "column %s is of type %s and thus is not indexable"
		typInfo = col.Type.DebugString()
		msg = fmt.Sprintf(msg, col.Name, col.Type.Name())
	} else {
		__antithesis_instrumentation__.Notify(269624)
		msg = "the following columns are not indexable due to their type: "
		for i := range cols {
			__antithesis_instrumentation__.Notify(269625)
			col := &cols[i]
			msg += fmt.Sprintf("%s (type %s)", col.Name, col.Type.Name())
			typInfo += col.Type.DebugString()
			if i != len(cols)-1 {
				__antithesis_instrumentation__.Notify(269626)
				msg += ", "
				typInfo += ","
			} else {
				__antithesis_instrumentation__.Notify(269627)
			}
		}
	}
	__antithesis_instrumentation__.Notify(269620)
	return unimplemented.NewWithIssueDetailf(35730, typInfo, "%s", msg)
}

func checkColumnsValidForIndex(tableDesc *Mutable, indexColNames []string) error {
	__antithesis_instrumentation__.Notify(269628)
	invalidColumns := make([]descpb.ColumnDescriptor, 0, len(indexColNames))
	for _, indexCol := range indexColNames {
		__antithesis_instrumentation__.Notify(269631)
		for _, col := range tableDesc.NonDropColumns() {
			__antithesis_instrumentation__.Notify(269632)
			if col.GetName() == indexCol {
				__antithesis_instrumentation__.Notify(269633)
				if !colinfo.ColumnTypeIsIndexable(col.GetType()) {
					__antithesis_instrumentation__.Notify(269634)
					invalidColumns = append(invalidColumns, *col.ColumnDesc())
				} else {
					__antithesis_instrumentation__.Notify(269635)
				}
			} else {
				__antithesis_instrumentation__.Notify(269636)
			}
		}
	}
	__antithesis_instrumentation__.Notify(269629)
	if len(invalidColumns) > 0 {
		__antithesis_instrumentation__.Notify(269637)
		return notIndexableError(invalidColumns)
	} else {
		__antithesis_instrumentation__.Notify(269638)
	}
	__antithesis_instrumentation__.Notify(269630)
	return nil
}

func checkColumnsValidForInvertedIndex(tableDesc *Mutable, indexColNames []string) error {
	__antithesis_instrumentation__.Notify(269639)
	lastCol := len(indexColNames) - 1
	for i, indexCol := range indexColNames {
		__antithesis_instrumentation__.Notify(269641)
		for _, col := range tableDesc.NonDropColumns() {
			__antithesis_instrumentation__.Notify(269642)
			if col.GetName() == indexCol {
				__antithesis_instrumentation__.Notify(269643)

				if i == lastCol && func() bool {
					__antithesis_instrumentation__.Notify(269645)
					return !colinfo.ColumnTypeIsInvertedIndexable(col.GetType()) == true
				}() == true {
					__antithesis_instrumentation__.Notify(269646)
					return errors.WithHint(
						pgerror.Newf(
							pgcode.FeatureNotSupported,
							"column %s of type %s is not allowed as the last column in an inverted index",
							col.GetName(),
							col.GetType().Name(),
						),
						"see the documentation for more information about inverted indexes: "+docs.URL("inverted-indexes.html"),
					)

				} else {
					__antithesis_instrumentation__.Notify(269647)
				}
				__antithesis_instrumentation__.Notify(269644)

				if i < lastCol && func() bool {
					__antithesis_instrumentation__.Notify(269648)
					return !colinfo.ColumnTypeIsIndexable(col.GetType()) == true
				}() == true {
					__antithesis_instrumentation__.Notify(269649)
					return errors.WithHint(
						pgerror.Newf(
							pgcode.FeatureNotSupported,
							"column %s of type %s is only allowed as the last column in an inverted index",
							col.GetName(),
							col.GetType().Name(),
						),
						"see the documentation for more information about inverted indexes: "+docs.URL("inverted-indexes.html"),
					)
				} else {
					__antithesis_instrumentation__.Notify(269650)
				}
			} else {
				__antithesis_instrumentation__.Notify(269651)
			}
		}
	}
	__antithesis_instrumentation__.Notify(269640)
	return nil
}

func (desc *Mutable) AddColumn(col *descpb.ColumnDescriptor) {
	__antithesis_instrumentation__.Notify(269652)
	desc.Columns = append(desc.Columns, *col)
}

func (desc *Mutable) AddFamily(fam descpb.ColumnFamilyDescriptor) {
	__antithesis_instrumentation__.Notify(269653)
	desc.Families = append(desc.Families, fam)
}

func (desc *Mutable) AddPrimaryIndex(idx descpb.IndexDescriptor) error {
	__antithesis_instrumentation__.Notify(269654)
	if idx.Type == descpb.IndexDescriptor_INVERTED {
		__antithesis_instrumentation__.Notify(269660)
		return fmt.Errorf("primary index cannot be inverted")
	} else {
		__antithesis_instrumentation__.Notify(269661)
	}
	__antithesis_instrumentation__.Notify(269655)
	if err := checkColumnsValidForIndex(desc, idx.KeyColumnNames); err != nil {
		__antithesis_instrumentation__.Notify(269662)
		return err
	} else {
		__antithesis_instrumentation__.Notify(269663)
	}
	__antithesis_instrumentation__.Notify(269656)
	if desc.PrimaryIndex.Name != "" {
		__antithesis_instrumentation__.Notify(269664)
		return fmt.Errorf("multiple primary keys for table %q are not allowed", desc.Name)
	} else {
		__antithesis_instrumentation__.Notify(269665)
	}
	__antithesis_instrumentation__.Notify(269657)
	if idx.Name == "" {
		__antithesis_instrumentation__.Notify(269666)

		idx.Name = PrimaryKeyIndexName(desc.Name)
	} else {
		__antithesis_instrumentation__.Notify(269667)
	}
	__antithesis_instrumentation__.Notify(269658)
	idx.EncodingType = descpb.PrimaryIndexEncoding
	if idx.Version < descpb.PrimaryIndexWithStoredColumnsVersion {
		__antithesis_instrumentation__.Notify(269668)
		idx.Version = descpb.PrimaryIndexWithStoredColumnsVersion

		names := make(map[string]struct{})
		for _, name := range idx.KeyColumnNames {
			__antithesis_instrumentation__.Notify(269671)
			names[name] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(269669)
		cols := desc.DeletableColumns()
		idx.StoreColumnNames = make([]string, 0, len(cols))
		for _, col := range cols {
			__antithesis_instrumentation__.Notify(269672)
			if _, found := names[col.GetName()]; found || func() bool {
				__antithesis_instrumentation__.Notify(269674)
				return col.IsVirtual() == true
			}() == true {
				__antithesis_instrumentation__.Notify(269675)
				continue
			} else {
				__antithesis_instrumentation__.Notify(269676)
			}
			__antithesis_instrumentation__.Notify(269673)
			names[col.GetName()] = struct{}{}
			idx.StoreColumnNames = append(idx.StoreColumnNames, col.GetName())
		}
		__antithesis_instrumentation__.Notify(269670)
		if len(idx.StoreColumnNames) == 0 {
			__antithesis_instrumentation__.Notify(269677)
			idx.StoreColumnNames = nil
		} else {
			__antithesis_instrumentation__.Notify(269678)
		}
	} else {
		__antithesis_instrumentation__.Notify(269679)
	}
	__antithesis_instrumentation__.Notify(269659)
	desc.SetPrimaryIndex(idx)
	return nil
}

func (desc *Mutable) AddSecondaryIndex(idx descpb.IndexDescriptor) error {
	__antithesis_instrumentation__.Notify(269680)
	if idx.Type == descpb.IndexDescriptor_FORWARD {
		__antithesis_instrumentation__.Notify(269682)
		if err := checkColumnsValidForIndex(desc, idx.KeyColumnNames); err != nil {
			__antithesis_instrumentation__.Notify(269683)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269684)
		}
	} else {
		__antithesis_instrumentation__.Notify(269685)
		if err := checkColumnsValidForInvertedIndex(desc, idx.KeyColumnNames); err != nil {
			__antithesis_instrumentation__.Notify(269686)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269687)
		}
	}
	__antithesis_instrumentation__.Notify(269681)
	desc.AddPublicNonPrimaryIndex(idx)
	return nil
}

func (desc *Mutable) AddColumnToFamilyMaybeCreate(
	col string, family string, create bool, ifNotExists bool,
) error {
	__antithesis_instrumentation__.Notify(269688)
	idx := int(-1)
	if len(family) > 0 {
		__antithesis_instrumentation__.Notify(269692)
		for i := range desc.Families {
			__antithesis_instrumentation__.Notify(269693)
			if desc.Families[i].Name == family {
				__antithesis_instrumentation__.Notify(269694)
				idx = i
				break
			} else {
				__antithesis_instrumentation__.Notify(269695)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(269696)
	}
	__antithesis_instrumentation__.Notify(269689)

	if idx == -1 {
		__antithesis_instrumentation__.Notify(269697)
		if create {
			__antithesis_instrumentation__.Notify(269699)

			desc.AddFamily(descpb.ColumnFamilyDescriptor{Name: family, ColumnNames: []string{col}})
			return nil
		} else {
			__antithesis_instrumentation__.Notify(269700)
		}
		__antithesis_instrumentation__.Notify(269698)
		return fmt.Errorf("unknown family %q", family)
	} else {
		__antithesis_instrumentation__.Notify(269701)
	}
	__antithesis_instrumentation__.Notify(269690)

	if create && func() bool {
		__antithesis_instrumentation__.Notify(269702)
		return !ifNotExists == true
	}() == true {
		__antithesis_instrumentation__.Notify(269703)
		return fmt.Errorf("family %q already exists", family)
	} else {
		__antithesis_instrumentation__.Notify(269704)
	}
	__antithesis_instrumentation__.Notify(269691)
	desc.Families[idx].ColumnNames = append(desc.Families[idx].ColumnNames, col)
	return nil
}

func (desc *Mutable) RemoveColumnFromFamilyAndPrimaryIndex(colID descpb.ColumnID) {
	__antithesis_instrumentation__.Notify(269705)
	desc.removeColumnFromFamily(colID)
	idx := desc.GetPrimaryIndex().IndexDescDeepCopy()
	for i, id := range idx.StoreColumnIDs {
		__antithesis_instrumentation__.Notify(269706)
		if id == colID {
			__antithesis_instrumentation__.Notify(269707)
			idx.StoreColumnIDs = append(idx.StoreColumnIDs[:i], idx.StoreColumnIDs[i+1:]...)
			idx.StoreColumnNames = append(idx.StoreColumnNames[:i], idx.StoreColumnNames[i+1:]...)
			desc.SetPrimaryIndex(idx)
			return
		} else {
			__antithesis_instrumentation__.Notify(269708)
		}
	}
}

func (desc *Mutable) removeColumnFromFamily(colID descpb.ColumnID) {
	__antithesis_instrumentation__.Notify(269709)
	for i := range desc.Families {
		__antithesis_instrumentation__.Notify(269710)
		for j, c := range desc.Families[i].ColumnIDs {
			__antithesis_instrumentation__.Notify(269711)
			if c == colID {
				__antithesis_instrumentation__.Notify(269712)
				desc.Families[i].ColumnIDs = append(
					desc.Families[i].ColumnIDs[:j], desc.Families[i].ColumnIDs[j+1:]...)
				desc.Families[i].ColumnNames = append(
					desc.Families[i].ColumnNames[:j], desc.Families[i].ColumnNames[j+1:]...)

				if len(desc.Families[i].ColumnIDs) == 0 && func() bool {
					__antithesis_instrumentation__.Notify(269714)
					return desc.Families[i].ID != 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(269715)
					desc.Families = append(desc.Families[:i], desc.Families[i+1:]...)
				} else {
					__antithesis_instrumentation__.Notify(269716)
				}
				__antithesis_instrumentation__.Notify(269713)
				return
			} else {
				__antithesis_instrumentation__.Notify(269717)
			}
		}
	}
}

func (desc *Mutable) RenameColumnDescriptor(column catalog.Column, newColName string) {
	__antithesis_instrumentation__.Notify(269718)
	colID := column.GetID()
	column.ColumnDesc().Name = newColName

	for i := range desc.Families {
		__antithesis_instrumentation__.Notify(269720)
		for j := range desc.Families[i].ColumnIDs {
			__antithesis_instrumentation__.Notify(269721)
			if desc.Families[i].ColumnIDs[j] == colID {
				__antithesis_instrumentation__.Notify(269722)
				desc.Families[i].ColumnNames[j] = newColName
			} else {
				__antithesis_instrumentation__.Notify(269723)
			}
		}
	}
	__antithesis_instrumentation__.Notify(269719)

	for _, idx := range desc.AllIndexes() {
		__antithesis_instrumentation__.Notify(269724)
		idxDesc := idx.IndexDesc()
		for i, id := range idxDesc.KeyColumnIDs {
			__antithesis_instrumentation__.Notify(269726)
			if id == colID {
				__antithesis_instrumentation__.Notify(269727)
				idxDesc.KeyColumnNames[i] = newColName
			} else {
				__antithesis_instrumentation__.Notify(269728)
			}
		}
		__antithesis_instrumentation__.Notify(269725)
		for i, id := range idxDesc.StoreColumnIDs {
			__antithesis_instrumentation__.Notify(269729)
			if id == colID {
				__antithesis_instrumentation__.Notify(269730)
				idxDesc.StoreColumnNames[i] = newColName
			} else {
				__antithesis_instrumentation__.Notify(269731)
			}
		}
	}
}

func (desc *Mutable) FindActiveOrNewColumnByName(name tree.Name) (catalog.Column, error) {
	__antithesis_instrumentation__.Notify(269732)
	currentMutationID := desc.ClusterVersion().NextMutationID
	for _, col := range desc.DeletableColumns() {
		__antithesis_instrumentation__.Notify(269734)
		if col.ColName() == name && func() bool {
			__antithesis_instrumentation__.Notify(269735)
			return ((col.Public()) || func() bool {
				__antithesis_instrumentation__.Notify(269736)
				return (col.Adding() && func() bool {
					__antithesis_instrumentation__.Notify(269737)
					return col.MutationID() == currentMutationID == true
				}() == true) == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(269738)
			return col, nil
		} else {
			__antithesis_instrumentation__.Notify(269739)
		}
	}
	__antithesis_instrumentation__.Notify(269733)
	return nil, colinfo.NewUndefinedColumnError(string(name))
}

func (desc *wrapper) ContainsUserDefinedTypes() bool {
	__antithesis_instrumentation__.Notify(269740)
	return len(desc.UserDefinedTypeColumns()) > 0
}

func (desc *wrapper) FindFamilyByID(id descpb.FamilyID) (*descpb.ColumnFamilyDescriptor, error) {
	__antithesis_instrumentation__.Notify(269741)
	for i := range desc.Families {
		__antithesis_instrumentation__.Notify(269743)
		family := &desc.Families[i]
		if family.ID == id {
			__antithesis_instrumentation__.Notify(269744)
			return family, nil
		} else {
			__antithesis_instrumentation__.Notify(269745)
		}
	}
	__antithesis_instrumentation__.Notify(269742)
	return nil, fmt.Errorf("family-id \"%d\" does not exist", id)
}

func (desc *wrapper) NamesForColumnIDs(ids descpb.ColumnIDs) ([]string, error) {
	__antithesis_instrumentation__.Notify(269746)
	names := make([]string, len(ids))
	for i, id := range ids {
		__antithesis_instrumentation__.Notify(269748)
		col, err := desc.FindColumnWithID(id)
		if err != nil {
			__antithesis_instrumentation__.Notify(269750)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(269751)
		}
		__antithesis_instrumentation__.Notify(269749)
		names[i] = col.GetName()
	}
	__antithesis_instrumentation__.Notify(269747)
	return names, nil
}

func (desc *Mutable) DropConstraint(
	ctx context.Context,
	name string,
	detail descpb.ConstraintDetail,
	removeFK func(*Mutable, *descpb.ForeignKeyConstraint) error,
	settings *cluster.Settings,
) error {
	__antithesis_instrumentation__.Notify(269752)
	switch detail.Kind {
	case descpb.ConstraintTypePK:
		__antithesis_instrumentation__.Notify(269755)
		{
			__antithesis_instrumentation__.Notify(269769)
			primaryIndex := desc.PrimaryIndex
			primaryIndex.Disabled = true
			desc.SetPrimaryIndex(primaryIndex)
		}
		__antithesis_instrumentation__.Notify(269756)
		return nil

	case descpb.ConstraintTypeUnique:
		__antithesis_instrumentation__.Notify(269757)
		if detail.Index != nil {
			__antithesis_instrumentation__.Notify(269770)
			return unimplemented.NewWithIssueDetailf(42840, "drop-constraint-unique",
				"cannot drop UNIQUE constraint %q using ALTER TABLE DROP CONSTRAINT, use DROP INDEX CASCADE instead",
				tree.ErrNameStringP(&detail.Index.Name))
		} else {
			__antithesis_instrumentation__.Notify(269771)
		}
		__antithesis_instrumentation__.Notify(269758)
		if detail.UniqueWithoutIndexConstraint == nil {
			__antithesis_instrumentation__.Notify(269772)
			return errors.AssertionFailedf(
				"Index or UniqueWithoutIndexConstraint must be non-nil for a unique constraint",
			)
		} else {
			__antithesis_instrumentation__.Notify(269773)
		}
		__antithesis_instrumentation__.Notify(269759)
		if detail.UniqueWithoutIndexConstraint.Validity == descpb.ConstraintValidity_Validating {
			__antithesis_instrumentation__.Notify(269774)
			return unimplemented.NewWithIssueDetailf(42844,
				"drop-constraint-unique-validating",
				"constraint %q in the middle of being added, try again later", name)
		} else {
			__antithesis_instrumentation__.Notify(269775)
		}
		__antithesis_instrumentation__.Notify(269760)
		if detail.UniqueWithoutIndexConstraint.Validity == descpb.ConstraintValidity_Dropping {
			__antithesis_instrumentation__.Notify(269776)
			return unimplemented.NewWithIssueDetailf(42844,
				"drop-constraint-unique-mutation",
				"constraint %q in the middle of being dropped", name)
		} else {
			__antithesis_instrumentation__.Notify(269777)
		}
		__antithesis_instrumentation__.Notify(269761)

		for i := range desc.UniqueWithoutIndexConstraints {
			__antithesis_instrumentation__.Notify(269778)
			ref := &desc.UniqueWithoutIndexConstraints[i]
			if ref.Name == name {
				__antithesis_instrumentation__.Notify(269779)

				if detail.UniqueWithoutIndexConstraint.Validity == descpb.ConstraintValidity_Unvalidated {
					__antithesis_instrumentation__.Notify(269781)
					desc.UniqueWithoutIndexConstraints = append(
						desc.UniqueWithoutIndexConstraints[:i], desc.UniqueWithoutIndexConstraints[i+1:]...,
					)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(269782)
				}
				__antithesis_instrumentation__.Notify(269780)
				ref.Validity = descpb.ConstraintValidity_Dropping
				desc.AddUniqueWithoutIndexMutation(ref, descpb.DescriptorMutation_DROP)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(269783)
			}
		}

	case descpb.ConstraintTypeCheck:
		__antithesis_instrumentation__.Notify(269762)
		if detail.CheckConstraint.Validity == descpb.ConstraintValidity_Validating {
			__antithesis_instrumentation__.Notify(269784)
			return unimplemented.NewWithIssueDetailf(42844, "drop-constraint-check-mutation",
				"constraint %q in the middle of being added, try again later", name)
		} else {
			__antithesis_instrumentation__.Notify(269785)
		}
		__antithesis_instrumentation__.Notify(269763)
		if detail.CheckConstraint.Validity == descpb.ConstraintValidity_Dropping {
			__antithesis_instrumentation__.Notify(269786)
			return unimplemented.NewWithIssueDetailf(42844, "drop-constraint-check-mutation",
				"constraint %q in the middle of being dropped", name)
		} else {
			__antithesis_instrumentation__.Notify(269787)
		}
		__antithesis_instrumentation__.Notify(269764)
		for i, c := range desc.Checks {
			__antithesis_instrumentation__.Notify(269788)
			if c.Name == name {
				__antithesis_instrumentation__.Notify(269789)

				if detail.CheckConstraint.Validity == descpb.ConstraintValidity_Unvalidated {
					__antithesis_instrumentation__.Notify(269791)
					desc.Checks = append(desc.Checks[:i], desc.Checks[i+1:]...)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(269792)
				}
				__antithesis_instrumentation__.Notify(269790)
				c.Validity = descpb.ConstraintValidity_Dropping
				desc.AddCheckMutation(c, descpb.DescriptorMutation_DROP)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(269793)
			}
		}

	case descpb.ConstraintTypeFK:
		__antithesis_instrumentation__.Notify(269765)
		if detail.FK.Validity == descpb.ConstraintValidity_Validating {
			__antithesis_instrumentation__.Notify(269794)
			return unimplemented.NewWithIssueDetailf(42844,
				"drop-constraint-fk-validating",
				"constraint %q in the middle of being added, try again later", name)
		} else {
			__antithesis_instrumentation__.Notify(269795)
		}
		__antithesis_instrumentation__.Notify(269766)
		if detail.FK.Validity == descpb.ConstraintValidity_Dropping {
			__antithesis_instrumentation__.Notify(269796)
			return unimplemented.NewWithIssueDetailf(42844,
				"drop-constraint-fk-mutation",
				"constraint %q in the middle of being dropped", name)
		} else {
			__antithesis_instrumentation__.Notify(269797)
		}
		__antithesis_instrumentation__.Notify(269767)

		for i := range desc.OutboundFKs {
			__antithesis_instrumentation__.Notify(269798)
			ref := &desc.OutboundFKs[i]
			if ref.Name == name {
				__antithesis_instrumentation__.Notify(269799)

				if detail.FK.Validity == descpb.ConstraintValidity_Unvalidated {
					__antithesis_instrumentation__.Notify(269801)

					if err := removeFK(desc, detail.FK); err != nil {
						__antithesis_instrumentation__.Notify(269803)
						return err
					} else {
						__antithesis_instrumentation__.Notify(269804)
					}
					__antithesis_instrumentation__.Notify(269802)
					desc.OutboundFKs = append(desc.OutboundFKs[:i], desc.OutboundFKs[i+1:]...)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(269805)
				}
				__antithesis_instrumentation__.Notify(269800)
				ref.Validity = descpb.ConstraintValidity_Dropping
				desc.AddForeignKeyMutation(ref, descpb.DescriptorMutation_DROP)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(269806)
			}
		}

	default:
		__antithesis_instrumentation__.Notify(269768)
		return unimplemented.Newf(fmt.Sprintf("drop-constraint-%s", detail.Kind),
			"constraint %q has unsupported type", tree.ErrNameString(name))
	}
	__antithesis_instrumentation__.Notify(269753)

	for i := range desc.Mutations {
		__antithesis_instrumentation__.Notify(269807)
		m := &desc.Mutations[i]
		if m.GetConstraint() != nil && func() bool {
			__antithesis_instrumentation__.Notify(269808)
			return m.GetConstraint().Name == name == true
		}() == true {
			__antithesis_instrumentation__.Notify(269809)
			switch m.Direction {
			case descpb.DescriptorMutation_ADD:
				__antithesis_instrumentation__.Notify(269810)
				return unimplemented.NewWithIssueDetailf(42844,
					"drop-constraint-mutation",
					"constraint %q in the middle of being added, try again later", name)
			case descpb.DescriptorMutation_DROP:
				__antithesis_instrumentation__.Notify(269811)
				return unimplemented.NewWithIssueDetailf(42844,
					"drop-constraint-mutation",
					"constraint %q in the middle of being dropped", name)
			default:
				__antithesis_instrumentation__.Notify(269812)
			}
		} else {
			__antithesis_instrumentation__.Notify(269813)
		}
	}
	__antithesis_instrumentation__.Notify(269754)
	return errors.AssertionFailedf("constraint %q not found on table %q", name, desc.Name)
}

func (desc *Mutable) RenameConstraint(
	detail descpb.ConstraintDetail,
	oldName, newName string,
	dependentViewRenameError func(string, descpb.ID) error,
	renameFK func(*Mutable, *descpb.ForeignKeyConstraint, string) error,
) error {
	__antithesis_instrumentation__.Notify(269814)
	switch detail.Kind {
	case descpb.ConstraintTypePK:
		__antithesis_instrumentation__.Notify(269815)
		for _, tableRef := range desc.DependedOnBy {
			__antithesis_instrumentation__.Notify(269827)
			if tableRef.IndexID != detail.Index.ID {
				__antithesis_instrumentation__.Notify(269829)
				continue
			} else {
				__antithesis_instrumentation__.Notify(269830)
			}
			__antithesis_instrumentation__.Notify(269828)
			return dependentViewRenameError("index", tableRef.ID)
		}
		__antithesis_instrumentation__.Notify(269816)
		idx, err := desc.FindIndexWithID(detail.Index.ID)
		if err != nil {
			__antithesis_instrumentation__.Notify(269831)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269832)
		}
		__antithesis_instrumentation__.Notify(269817)
		idx.IndexDesc().Name = newName
		return nil

	case descpb.ConstraintTypeUnique:
		__antithesis_instrumentation__.Notify(269818)
		if detail.Index != nil {
			__antithesis_instrumentation__.Notify(269833)
			for _, tableRef := range desc.DependedOnBy {
				__antithesis_instrumentation__.Notify(269836)
				if tableRef.IndexID != detail.Index.ID {
					__antithesis_instrumentation__.Notify(269838)
					continue
				} else {
					__antithesis_instrumentation__.Notify(269839)
				}
				__antithesis_instrumentation__.Notify(269837)
				return dependentViewRenameError("index", tableRef.ID)
			}
			__antithesis_instrumentation__.Notify(269834)
			idx, err := desc.FindIndexWithID(detail.Index.ID)
			if err != nil {
				__antithesis_instrumentation__.Notify(269840)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269841)
			}
			__antithesis_instrumentation__.Notify(269835)
			idx.IndexDesc().Name = newName
		} else {
			__antithesis_instrumentation__.Notify(269842)
			if detail.UniqueWithoutIndexConstraint != nil {
				__antithesis_instrumentation__.Notify(269843)
				if detail.UniqueWithoutIndexConstraint.Validity == descpb.ConstraintValidity_Validating {
					__antithesis_instrumentation__.Notify(269845)
					return unimplemented.NewWithIssueDetailf(42844,
						"rename-constraint-unique-mutation",
						"constraint %q in the middle of being added, try again later",
						tree.ErrNameStringP(&detail.UniqueWithoutIndexConstraint.Name))
				} else {
					__antithesis_instrumentation__.Notify(269846)
				}
				__antithesis_instrumentation__.Notify(269844)
				detail.UniqueWithoutIndexConstraint.Name = newName
			} else {
				__antithesis_instrumentation__.Notify(269847)
				return errors.AssertionFailedf(
					"Index or UniqueWithoutIndexConstraint must be non-nil for a unique constraint",
				)
			}
		}
		__antithesis_instrumentation__.Notify(269819)
		return nil

	case descpb.ConstraintTypeFK:
		__antithesis_instrumentation__.Notify(269820)
		if detail.FK.Validity == descpb.ConstraintValidity_Validating {
			__antithesis_instrumentation__.Notify(269848)
			return unimplemented.NewWithIssueDetailf(42844,
				"rename-constraint-fk-mutation",
				"constraint %q in the middle of being added, try again later",
				tree.ErrNameStringP(&detail.FK.Name))
		} else {
			__antithesis_instrumentation__.Notify(269849)
		}
		__antithesis_instrumentation__.Notify(269821)

		if err := renameFK(desc, detail.FK, newName); err != nil {
			__antithesis_instrumentation__.Notify(269850)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269851)
		}
		__antithesis_instrumentation__.Notify(269822)

		fk, err := desc.FindFKByName(detail.FK.Name)
		if err != nil {
			__antithesis_instrumentation__.Notify(269852)
			return err
		} else {
			__antithesis_instrumentation__.Notify(269853)
		}
		__antithesis_instrumentation__.Notify(269823)
		fk.Name = newName
		return nil

	case descpb.ConstraintTypeCheck:
		__antithesis_instrumentation__.Notify(269824)
		if detail.CheckConstraint.Validity == descpb.ConstraintValidity_Validating {
			__antithesis_instrumentation__.Notify(269854)
			return unimplemented.NewWithIssueDetailf(42844,
				"rename-constraint-check-mutation",
				"constraint %q in the middle of being added, try again later",
				tree.ErrNameStringP(&detail.CheckConstraint.Name))
		} else {
			__antithesis_instrumentation__.Notify(269855)
		}
		__antithesis_instrumentation__.Notify(269825)
		detail.CheckConstraint.Name = newName
		return nil

	default:
		__antithesis_instrumentation__.Notify(269826)
		return unimplemented.Newf(fmt.Sprintf("rename-constraint-%s", detail.Kind),
			"constraint %q has unsupported type", tree.ErrNameString(oldName))
	}
}

func (desc *wrapper) GetIndexMutationCapabilities(id descpb.IndexID) (bool, bool) {
	__antithesis_instrumentation__.Notify(269856)
	for _, mutation := range desc.Mutations {
		__antithesis_instrumentation__.Notify(269858)
		if mutationIndex := mutation.GetIndex(); mutationIndex != nil {
			__antithesis_instrumentation__.Notify(269859)
			if mutationIndex.ID == id {
				__antithesis_instrumentation__.Notify(269860)
				return true,
					mutation.State == descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY
			} else {
				__antithesis_instrumentation__.Notify(269861)
			}
		} else {
			__antithesis_instrumentation__.Notify(269862)
		}
	}
	__antithesis_instrumentation__.Notify(269857)
	return false, false
}

func (desc *wrapper) FindFKByName(name string) (*descpb.ForeignKeyConstraint, error) {
	__antithesis_instrumentation__.Notify(269863)
	for i := range desc.OutboundFKs {
		__antithesis_instrumentation__.Notify(269865)
		fk := &desc.OutboundFKs[i]
		if fk.Name == name {
			__antithesis_instrumentation__.Notify(269866)
			return fk, nil
		} else {
			__antithesis_instrumentation__.Notify(269867)
		}
	}
	__antithesis_instrumentation__.Notify(269864)
	return nil, fmt.Errorf("fk %q does not exist", name)
}

func (desc *wrapper) IsPrimaryIndexDefaultRowID() bool {
	__antithesis_instrumentation__.Notify(269868)
	if len(desc.PrimaryIndex.KeyColumnIDs) != 1 {
		__antithesis_instrumentation__.Notify(269875)
		return false
	} else {
		__antithesis_instrumentation__.Notify(269876)
	}
	__antithesis_instrumentation__.Notify(269869)
	col, err := desc.FindColumnWithID(desc.PrimaryIndex.KeyColumnIDs[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(269877)

		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(269878)
	}
	__antithesis_instrumentation__.Notify(269870)
	if !col.IsHidden() {
		__antithesis_instrumentation__.Notify(269879)
		return false
	} else {
		__antithesis_instrumentation__.Notify(269880)
	}
	__antithesis_instrumentation__.Notify(269871)
	if !strings.HasPrefix(col.GetName(), "rowid") {
		__antithesis_instrumentation__.Notify(269881)
		return false
	} else {
		__antithesis_instrumentation__.Notify(269882)
	}
	__antithesis_instrumentation__.Notify(269872)
	if !col.GetType().Equal(types.Int) {
		__antithesis_instrumentation__.Notify(269883)
		return false
	} else {
		__antithesis_instrumentation__.Notify(269884)
	}
	__antithesis_instrumentation__.Notify(269873)
	if !col.HasDefault() {
		__antithesis_instrumentation__.Notify(269885)
		return false
	} else {
		__antithesis_instrumentation__.Notify(269886)
	}
	__antithesis_instrumentation__.Notify(269874)
	return col.GetDefaultExpr() == "unique_rowid()"
}

func (desc *Mutable) MakeMutationComplete(m descpb.DescriptorMutation) error {
	__antithesis_instrumentation__.Notify(269887)
	switch m.Direction {
	case descpb.DescriptorMutation_ADD:
		__antithesis_instrumentation__.Notify(269889)
		switch t := m.Descriptor_.(type) {
		case *descpb.DescriptorMutation_Column:
			__antithesis_instrumentation__.Notify(269892)
			desc.AddColumn(t.Column)

		case *descpb.DescriptorMutation_Index:
			__antithesis_instrumentation__.Notify(269893)
			if err := desc.AddSecondaryIndex(*t.Index); err != nil {
				__antithesis_instrumentation__.Notify(269902)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269903)
			}

		case *descpb.DescriptorMutation_Constraint:
			__antithesis_instrumentation__.Notify(269894)
			switch t.Constraint.ConstraintType {
			case descpb.ConstraintToUpdate_CHECK:
				__antithesis_instrumentation__.Notify(269904)
				switch t.Constraint.Check.Validity {
				case descpb.ConstraintValidity_Validating:
					__antithesis_instrumentation__.Notify(269911)

					for _, c := range desc.Checks {
						__antithesis_instrumentation__.Notify(269914)
						if c.Name == t.Constraint.Name {
							__antithesis_instrumentation__.Notify(269915)
							c.Validity = descpb.ConstraintValidity_Validated
							break
						} else {
							__antithesis_instrumentation__.Notify(269916)
						}
					}
				case descpb.ConstraintValidity_Unvalidated:
					__antithesis_instrumentation__.Notify(269912)

					desc.Checks = append(desc.Checks, &t.Constraint.Check)
				default:
					__antithesis_instrumentation__.Notify(269913)
					return errors.AssertionFailedf("invalid constraint validity state: %d", t.Constraint.Check.Validity)
				}
			case descpb.ConstraintToUpdate_FOREIGN_KEY:
				__antithesis_instrumentation__.Notify(269905)
				switch t.Constraint.ForeignKey.Validity {
				case descpb.ConstraintValidity_Validating:
					__antithesis_instrumentation__.Notify(269917)

					for i := range desc.OutboundFKs {
						__antithesis_instrumentation__.Notify(269920)
						fk := &desc.OutboundFKs[i]
						if fk.Name == t.Constraint.Name {
							__antithesis_instrumentation__.Notify(269921)
							fk.Validity = descpb.ConstraintValidity_Validated
							break
						} else {
							__antithesis_instrumentation__.Notify(269922)
						}
					}
				case descpb.ConstraintValidity_Unvalidated:
					__antithesis_instrumentation__.Notify(269918)

					desc.OutboundFKs = append(desc.OutboundFKs, t.Constraint.ForeignKey)
				default:
					__antithesis_instrumentation__.Notify(269919)
				}
			case descpb.ConstraintToUpdate_UNIQUE_WITHOUT_INDEX:
				__antithesis_instrumentation__.Notify(269906)
				switch t.Constraint.UniqueWithoutIndexConstraint.Validity {
				case descpb.ConstraintValidity_Validating:
					__antithesis_instrumentation__.Notify(269923)

					for i := range desc.UniqueWithoutIndexConstraints {
						__antithesis_instrumentation__.Notify(269926)
						uc := &desc.UniqueWithoutIndexConstraints[i]
						if uc.Name == t.Constraint.Name {
							__antithesis_instrumentation__.Notify(269927)
							uc.Validity = descpb.ConstraintValidity_Validated
							break
						} else {
							__antithesis_instrumentation__.Notify(269928)
						}
					}
				case descpb.ConstraintValidity_Unvalidated:
					__antithesis_instrumentation__.Notify(269924)

					desc.UniqueWithoutIndexConstraints = append(
						desc.UniqueWithoutIndexConstraints, t.Constraint.UniqueWithoutIndexConstraint,
					)
				default:
					__antithesis_instrumentation__.Notify(269925)
					return errors.AssertionFailedf("invalid constraint validity state: %d",
						t.Constraint.UniqueWithoutIndexConstraint.Validity,
					)
				}
			case descpb.ConstraintToUpdate_NOT_NULL:
				__antithesis_instrumentation__.Notify(269907)

				for i, c := range desc.Checks {
					__antithesis_instrumentation__.Notify(269929)
					if c.Name == t.Constraint.Check.Name {
						__antithesis_instrumentation__.Notify(269930)
						desc.Checks = append(desc.Checks[:i], desc.Checks[i+1:]...)
						break
					} else {
						__antithesis_instrumentation__.Notify(269931)
					}
				}
				__antithesis_instrumentation__.Notify(269908)
				col, err := desc.FindColumnWithID(t.Constraint.NotNullColumn)
				if err != nil {
					__antithesis_instrumentation__.Notify(269932)
					return err
				} else {
					__antithesis_instrumentation__.Notify(269933)
				}
				__antithesis_instrumentation__.Notify(269909)
				col.ColumnDesc().Nullable = false
			default:
				__antithesis_instrumentation__.Notify(269910)
				return errors.Errorf("unsupported constraint type: %d", t.Constraint.ConstraintType)
			}
		case *descpb.DescriptorMutation_PrimaryKeySwap:
			__antithesis_instrumentation__.Notify(269895)
			args := t.PrimaryKeySwap
			getIndexIdxByID := func(id descpb.IndexID) (int, error) {
				__antithesis_instrumentation__.Notify(269934)
				for i, idx := range desc.Indexes {
					__antithesis_instrumentation__.Notify(269936)
					if idx.ID == id {
						__antithesis_instrumentation__.Notify(269937)
						return i + 1, nil
					} else {
						__antithesis_instrumentation__.Notify(269938)
					}
				}
				__antithesis_instrumentation__.Notify(269935)
				return 0, errors.New("index was not in list of indexes")
			}
			__antithesis_instrumentation__.Notify(269896)

			primaryIndexCopy := desc.GetPrimaryIndex().IndexDescDeepCopy()

			if err := desc.AddDropIndexMutation(&primaryIndexCopy); err != nil {
				__antithesis_instrumentation__.Notify(269939)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269940)
			}
			__antithesis_instrumentation__.Notify(269897)

			newIndex, err := desc.FindIndexWithID(args.NewPrimaryIndexId)
			if err != nil {
				__antithesis_instrumentation__.Notify(269941)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269942)
			}

			{
				__antithesis_instrumentation__.Notify(269943)
				primaryIndex := newIndex.IndexDescDeepCopy()
				if args.NewPrimaryIndexName == "" {
					__antithesis_instrumentation__.Notify(269946)
					primaryIndex.Name = PrimaryKeyIndexName(desc.Name)
				} else {
					__antithesis_instrumentation__.Notify(269947)
					primaryIndex.Name = args.NewPrimaryIndexName
				}
				__antithesis_instrumentation__.Notify(269944)

				if primaryIndex.Version == descpb.StrictIndexColumnIDGuaranteesVersion {
					__antithesis_instrumentation__.Notify(269948)
					primaryIndex.Version = descpb.PrimaryIndexWithStoredColumnsVersion
				} else {
					__antithesis_instrumentation__.Notify(269949)
				}
				__antithesis_instrumentation__.Notify(269945)
				desc.SetPrimaryIndex(primaryIndex)
			}
			__antithesis_instrumentation__.Notify(269898)

			idx, err := getIndexIdxByID(newIndex.GetID())
			if err != nil {
				__antithesis_instrumentation__.Notify(269950)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269951)
			}
			__antithesis_instrumentation__.Notify(269899)
			desc.RemovePublicNonPrimaryIndex(idx)

			for j := range args.OldIndexes {
				__antithesis_instrumentation__.Notify(269952)
				oldID := args.OldIndexes[j]
				newID := args.NewIndexes[j]

				newIndex, err := desc.FindIndexWithID(newID)
				if err != nil {
					__antithesis_instrumentation__.Notify(269956)
					return err
				} else {
					__antithesis_instrumentation__.Notify(269957)
				}
				__antithesis_instrumentation__.Notify(269953)
				oldIndexIdx, err := getIndexIdxByID(oldID)
				if err != nil {
					__antithesis_instrumentation__.Notify(269958)
					return err
				} else {
					__antithesis_instrumentation__.Notify(269959)
				}
				__antithesis_instrumentation__.Notify(269954)
				if oldIndexIdx >= len(desc.ActiveIndexes()) {
					__antithesis_instrumentation__.Notify(269960)
					return errors.Errorf("invalid indexIdx %d", oldIndexIdx)
				} else {
					__antithesis_instrumentation__.Notify(269961)
				}
				__antithesis_instrumentation__.Notify(269955)
				oldIndex := desc.ActiveIndexes()[oldIndexIdx].IndexDesc()
				oldIndexCopy := protoutil.Clone(oldIndex).(*descpb.IndexDescriptor)
				newIndex.IndexDesc().Name = oldIndexCopy.Name

				desc.RemovePublicNonPrimaryIndex(oldIndexIdx)

				if err := desc.AddDropIndexMutation(oldIndexCopy); err != nil {
					__antithesis_instrumentation__.Notify(269962)
					return err
				} else {
					__antithesis_instrumentation__.Notify(269963)
				}
			}
		case *descpb.DescriptorMutation_ComputedColumnSwap:
			__antithesis_instrumentation__.Notify(269900)
			if err := desc.performComputedColumnSwap(t.ComputedColumnSwap); err != nil {
				__antithesis_instrumentation__.Notify(269964)
				return err
			} else {
				__antithesis_instrumentation__.Notify(269965)
			}

		case *descpb.DescriptorMutation_MaterializedViewRefresh:
			__antithesis_instrumentation__.Notify(269901)

			desc.SetPrimaryIndex(t.MaterializedViewRefresh.NewPrimaryIndex)
			desc.SetPublicNonPrimaryIndexes(t.MaterializedViewRefresh.NewIndexes)
		}

	case descpb.DescriptorMutation_DROP:
		__antithesis_instrumentation__.Notify(269890)
		switch t := m.Descriptor_.(type) {

		case *descpb.DescriptorMutation_Column:
			__antithesis_instrumentation__.Notify(269966)
			desc.RemoveColumnFromFamilyAndPrimaryIndex(t.Column.ID)
		}
	default:
		__antithesis_instrumentation__.Notify(269891)
	}
	__antithesis_instrumentation__.Notify(269888)
	return nil
}

func (desc *Mutable) performComputedColumnSwap(swap *descpb.ComputedColumnSwap) error {
	__antithesis_instrumentation__.Notify(269967)

	oldCol, err := desc.FindColumnWithID(swap.OldColumnId)
	if err != nil {
		__antithesis_instrumentation__.Notify(269977)
		return err
	} else {
		__antithesis_instrumentation__.Notify(269978)
	}
	__antithesis_instrumentation__.Notify(269968)
	newCol, err := desc.FindColumnWithID(swap.NewColumnId)
	if err != nil {
		__antithesis_instrumentation__.Notify(269979)
		return err
	} else {
		__antithesis_instrumentation__.Notify(269980)
	}
	__antithesis_instrumentation__.Notify(269969)

	newCol.ColumnDesc().ComputeExpr = nil

	oldCol.ColumnDesc().ComputeExpr = &swap.InverseExpr

	nameExists := func(name string) bool {
		__antithesis_instrumentation__.Notify(269981)
		_, err := desc.FindColumnWithName(tree.Name(name))
		return err == nil
	}
	__antithesis_instrumentation__.Notify(269970)

	uniqueName := GenerateUniqueName(newCol.GetName(), nameExists)

	oldColName := oldCol.GetName()

	desc.RenameColumnDescriptor(oldCol, uniqueName)
	desc.RenameColumnDescriptor(newCol, oldColName)

	oldColColumnFamily, err := desc.GetFamilyOfColumn(oldCol.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(269982)
		return err
	} else {
		__antithesis_instrumentation__.Notify(269983)
	}
	__antithesis_instrumentation__.Notify(269971)
	newColColumnFamily, err := desc.GetFamilyOfColumn(newCol.GetID())
	if err != nil {
		__antithesis_instrumentation__.Notify(269984)
		return err
	} else {
		__antithesis_instrumentation__.Notify(269985)
	}
	__antithesis_instrumentation__.Notify(269972)

	if oldColColumnFamily.ID != newColColumnFamily.ID {
		__antithesis_instrumentation__.Notify(269986)
		return errors.Newf("expected the column families of the old and new columns to match,"+
			"oldCol column family: %v, newCol column family: %v",
			oldColColumnFamily.ID, newColColumnFamily.ID)
	} else {
		__antithesis_instrumentation__.Notify(269987)
	}
	__antithesis_instrumentation__.Notify(269973)

	for i := range oldColColumnFamily.ColumnIDs {
		__antithesis_instrumentation__.Notify(269988)
		if oldColColumnFamily.ColumnIDs[i] == oldCol.GetID() {
			__antithesis_instrumentation__.Notify(269989)
			oldColColumnFamily.ColumnIDs[i] = newCol.GetID()
			oldColColumnFamily.ColumnNames[i] = newCol.GetName()
		} else {
			__antithesis_instrumentation__.Notify(269990)
			if oldColColumnFamily.ColumnIDs[i] == newCol.GetID() {
				__antithesis_instrumentation__.Notify(269991)
				oldColColumnFamily.ColumnIDs[i] = oldCol.GetID()
				oldColColumnFamily.ColumnNames[i] = oldCol.GetName()
			} else {
				__antithesis_instrumentation__.Notify(269992)
			}
		}
	}
	__antithesis_instrumentation__.Notify(269974)

	newCol.ColumnDesc().PGAttributeNum = oldCol.GetPGAttributeNum()
	oldCol.ColumnDesc().PGAttributeNum = 0

	oldCol.ColumnDesc().AlterColumnTypeInProgress = true

	oldColCopy := oldCol.ColumnDescDeepCopy()
	newColCopy := newCol.ColumnDescDeepCopy()
	desc.AddColumnMutation(&oldColCopy, descpb.DescriptorMutation_DROP)

	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(269993)
		if desc.Columns[i].ID == newCol.GetID() {
			__antithesis_instrumentation__.Notify(269994)
			desc.Columns = append(desc.Columns[:i:i], desc.Columns[i+1:]...)
			break
		} else {
			__antithesis_instrumentation__.Notify(269995)
		}
	}
	__antithesis_instrumentation__.Notify(269975)

	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(269996)
		if desc.Columns[i].ID == oldCol.GetID() {
			__antithesis_instrumentation__.Notify(269997)
			desc.Columns[i] = newColCopy
		} else {
			__antithesis_instrumentation__.Notify(269998)
		}
	}
	__antithesis_instrumentation__.Notify(269976)

	return nil
}

func (desc *Mutable) AddCheckMutation(
	ck *descpb.TableDescriptor_CheckConstraint, direction descpb.DescriptorMutation_Direction,
) {
	__antithesis_instrumentation__.Notify(269999)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_Constraint{
			Constraint: &descpb.ConstraintToUpdate{
				ConstraintType: descpb.ConstraintToUpdate_CHECK, Name: ck.Name, Check: *ck,
			},
		},
		Direction: direction,
	}
	desc.addMutation(m)
}

func (desc *Mutable) AddForeignKeyMutation(
	fk *descpb.ForeignKeyConstraint, direction descpb.DescriptorMutation_Direction,
) {
	__antithesis_instrumentation__.Notify(270000)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_Constraint{
			Constraint: &descpb.ConstraintToUpdate{
				ConstraintType: descpb.ConstraintToUpdate_FOREIGN_KEY,
				Name:           fk.Name,
				ForeignKey:     *fk,
			},
		},
		Direction: direction,
	}
	desc.addMutation(m)
}

func (desc *Mutable) AddUniqueWithoutIndexMutation(
	uc *descpb.UniqueWithoutIndexConstraint, direction descpb.DescriptorMutation_Direction,
) {
	__antithesis_instrumentation__.Notify(270001)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_Constraint{
			Constraint: &descpb.ConstraintToUpdate{
				ConstraintType:               descpb.ConstraintToUpdate_UNIQUE_WITHOUT_INDEX,
				Name:                         uc.Name,
				UniqueWithoutIndexConstraint: *uc,
			},
		},
		Direction: direction,
	}
	desc.addMutation(m)
}

func MakeNotNullCheckConstraint(
	colName string,
	colID descpb.ColumnID,
	constraintID descpb.ConstraintID,
	inuseNames map[string]struct{},
	validity descpb.ConstraintValidity,
) *descpb.TableDescriptor_CheckConstraint {
	__antithesis_instrumentation__.Notify(270002)
	name := fmt.Sprintf("%s_auto_not_null", colName)

	if _, ok := inuseNames[name]; ok {
		__antithesis_instrumentation__.Notify(270005)
		i := 1
		for {
			__antithesis_instrumentation__.Notify(270006)
			appended := fmt.Sprintf("%s%d", name, i)
			if _, ok := inuseNames[appended]; !ok {
				__antithesis_instrumentation__.Notify(270008)
				name = appended
				break
			} else {
				__antithesis_instrumentation__.Notify(270009)
			}
			__antithesis_instrumentation__.Notify(270007)
			i++
		}
	} else {
		__antithesis_instrumentation__.Notify(270010)
	}
	__antithesis_instrumentation__.Notify(270003)
	if inuseNames != nil {
		__antithesis_instrumentation__.Notify(270011)
		inuseNames[name] = struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(270012)
	}
	__antithesis_instrumentation__.Notify(270004)

	expr := &tree.IsNotNullExpr{
		Expr: &tree.ColumnItem{ColumnName: tree.Name(colName)},
	}

	return &descpb.TableDescriptor_CheckConstraint{
		Name:                name,
		Expr:                tree.Serialize(expr),
		Validity:            validity,
		ColumnIDs:           []descpb.ColumnID{colID},
		IsNonNullConstraint: true,
		ConstraintID:        constraintID,
	}
}

func (desc *Mutable) AddNotNullMutation(
	ck *descpb.TableDescriptor_CheckConstraint, direction descpb.DescriptorMutation_Direction,
) {
	__antithesis_instrumentation__.Notify(270013)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_Constraint{
			Constraint: &descpb.ConstraintToUpdate{
				ConstraintType: descpb.ConstraintToUpdate_NOT_NULL,
				Name:           ck.Name,
				NotNullColumn:  ck.ColumnIDs[0],
				Check:          *ck,
			},
		},
		Direction: direction,
	}
	desc.addMutation(m)
}

func (desc *Mutable) AddModifyRowLevelTTLMutation(
	ttl *descpb.ModifyRowLevelTTL, direction descpb.DescriptorMutation_Direction,
) {
	__antithesis_instrumentation__.Notify(270014)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_ModifyRowLevelTTL{ModifyRowLevelTTL: ttl},
		Direction:   direction,
	}
	desc.addMutation(m)
}

func (desc *Mutable) AddColumnMutation(
	c *descpb.ColumnDescriptor, direction descpb.DescriptorMutation_Direction,
) {
	__antithesis_instrumentation__.Notify(270015)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_Column{Column: c},
		Direction:   direction,
	}
	desc.addMutation(m)
}

func (desc *Mutable) AddDropIndexMutation(idx *descpb.IndexDescriptor) error {
	__antithesis_instrumentation__.Notify(270016)
	if err := desc.checkValidIndex(idx); err != nil {
		__antithesis_instrumentation__.Notify(270018)
		return err
	} else {
		__antithesis_instrumentation__.Notify(270019)
	}
	__antithesis_instrumentation__.Notify(270017)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_Index{Index: idx},
		Direction:   descpb.DescriptorMutation_DROP,
	}
	desc.addMutation(m)
	return nil
}

func (desc *Mutable) AddIndexMutation(
	ctx context.Context,
	idx *descpb.IndexDescriptor,
	direction descpb.DescriptorMutation_Direction,
	settings *cluster.Settings,
) error {
	__antithesis_instrumentation__.Notify(270020)
	if !settings.Version.IsActive(ctx, clusterversion.MVCCIndexBackfiller) || func() bool {
		__antithesis_instrumentation__.Notify(270023)
		return !UseMVCCCompliantIndexCreation.Get(&settings.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(270024)
		return desc.DeprecatedAddIndexMutation(idx, direction)
	} else {
		__antithesis_instrumentation__.Notify(270025)
	}
	__antithesis_instrumentation__.Notify(270021)

	if err := desc.checkValidIndex(idx); err != nil {
		__antithesis_instrumentation__.Notify(270026)
		return err
	} else {
		__antithesis_instrumentation__.Notify(270027)
	}
	__antithesis_instrumentation__.Notify(270022)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_Index{Index: idx},
		Direction:   direction,
	}
	desc.addMutation(m)
	return nil
}

func (desc *Mutable) DeprecatedAddIndexMutation(
	idx *descpb.IndexDescriptor, direction descpb.DescriptorMutation_Direction,
) error {
	__antithesis_instrumentation__.Notify(270028)
	if err := desc.checkValidIndex(idx); err != nil {
		__antithesis_instrumentation__.Notify(270030)
		return err
	} else {
		__antithesis_instrumentation__.Notify(270031)
	}
	__antithesis_instrumentation__.Notify(270029)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_Index{Index: idx},
		Direction:   direction,
	}
	desc.deprecatedAddMutation(m)
	return nil
}

func (desc *Mutable) checkValidIndex(idx *descpb.IndexDescriptor) error {
	__antithesis_instrumentation__.Notify(270032)
	switch idx.Type {
	case descpb.IndexDescriptor_FORWARD:
		__antithesis_instrumentation__.Notify(270034)
		if err := checkColumnsValidForIndex(desc, idx.KeyColumnNames); err != nil {
			__antithesis_instrumentation__.Notify(270037)
			return err
		} else {
			__antithesis_instrumentation__.Notify(270038)
		}
	case descpb.IndexDescriptor_INVERTED:
		__antithesis_instrumentation__.Notify(270035)
		if err := checkColumnsValidForInvertedIndex(desc, idx.KeyColumnNames); err != nil {
			__antithesis_instrumentation__.Notify(270039)
			return err
		} else {
			__antithesis_instrumentation__.Notify(270040)
		}
	default:
		__antithesis_instrumentation__.Notify(270036)
	}
	__antithesis_instrumentation__.Notify(270033)
	return nil
}

func (desc *Mutable) AddPrimaryKeySwapMutation(swap *descpb.PrimaryKeySwap) {
	__antithesis_instrumentation__.Notify(270041)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_PrimaryKeySwap{PrimaryKeySwap: swap},
		Direction:   descpb.DescriptorMutation_ADD,
	}
	desc.addMutation(m)
}

func (desc *Mutable) AddMaterializedViewRefreshMutation(refresh *descpb.MaterializedViewRefresh) {
	__antithesis_instrumentation__.Notify(270042)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_MaterializedViewRefresh{MaterializedViewRefresh: refresh},
		Direction:   descpb.DescriptorMutation_ADD,
	}
	desc.addMutation(m)
}

func (desc *Mutable) AddComputedColumnSwapMutation(swap *descpb.ComputedColumnSwap) {
	__antithesis_instrumentation__.Notify(270043)
	m := descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_ComputedColumnSwap{ComputedColumnSwap: swap},
		Direction:   descpb.DescriptorMutation_ADD,
	}
	desc.addMutation(m)
}

func (desc *Mutable) addMutation(m descpb.DescriptorMutation) {
	__antithesis_instrumentation__.Notify(270044)
	switch m.Direction {
	case descpb.DescriptorMutation_ADD:
		__antithesis_instrumentation__.Notify(270046)
		switch m.Descriptor_.(type) {
		case *descpb.DescriptorMutation_Index:
			__antithesis_instrumentation__.Notify(270049)
			m.State = descpb.DescriptorMutation_BACKFILLING
		default:
			__antithesis_instrumentation__.Notify(270050)
			m.State = descpb.DescriptorMutation_DELETE_ONLY
		}
	case descpb.DescriptorMutation_DROP:
		__antithesis_instrumentation__.Notify(270047)
		m.State = descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY
	default:
		__antithesis_instrumentation__.Notify(270048)
	}
	__antithesis_instrumentation__.Notify(270045)
	desc.addMutationWithNextID(m)

	if idxMut, ok := m.Descriptor_.(*descpb.DescriptorMutation_Index); ok {
		__antithesis_instrumentation__.Notify(270051)
		if m.Direction == descpb.DescriptorMutation_ADD {
			__antithesis_instrumentation__.Notify(270052)
			tempIndex := *protoutil.Clone(idxMut.Index).(*descpb.IndexDescriptor)
			tempIndex.UseDeletePreservingEncoding = true
			tempIndex.ID = 0
			tempIndex.Name = ""
			m2 := descpb.DescriptorMutation{
				Descriptor_: &descpb.DescriptorMutation_Index{Index: &tempIndex},
				Direction:   descpb.DescriptorMutation_ADD,
				State:       descpb.DescriptorMutation_DELETE_ONLY,
			}
			desc.addMutationWithNextID(m2)
		} else {
			__antithesis_instrumentation__.Notify(270053)
		}
	} else {
		__antithesis_instrumentation__.Notify(270054)
	}
}

func (desc *Mutable) deprecatedAddMutation(m descpb.DescriptorMutation) {
	__antithesis_instrumentation__.Notify(270055)
	switch m.Direction {
	case descpb.DescriptorMutation_ADD:
		__antithesis_instrumentation__.Notify(270057)
		m.State = descpb.DescriptorMutation_DELETE_ONLY
	case descpb.DescriptorMutation_DROP:
		__antithesis_instrumentation__.Notify(270058)
		m.State = descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY
	default:
		__antithesis_instrumentation__.Notify(270059)
	}
	__antithesis_instrumentation__.Notify(270056)
	desc.addMutationWithNextID(m)
}

func (desc *Mutable) addMutationWithNextID(m descpb.DescriptorMutation) {
	__antithesis_instrumentation__.Notify(270060)

	m.MutationID = desc.ClusterVersion().NextMutationID
	desc.NextMutationID = desc.ClusterVersion().NextMutationID + 1
	desc.Mutations = append(desc.Mutations, m)
}

func (desc *wrapper) MakeFirstMutationPublic(
	includeConstraints catalog.MutationPublicationFilter,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(270061)

	table := desc.NewBuilder().(TableDescriptorBuilder).BuildExistingMutableTable()
	mutationID := desc.Mutations[0].MutationID
	i := 0
	for _, mutation := range desc.Mutations {
		__antithesis_instrumentation__.Notify(270063)
		if mutation.MutationID != mutationID {
			__antithesis_instrumentation__.Notify(270066)

			break
		} else {
			__antithesis_instrumentation__.Notify(270067)
		}
		__antithesis_instrumentation__.Notify(270064)
		i++
		if mutation.GetPrimaryKeySwap() != nil && func() bool {
			__antithesis_instrumentation__.Notify(270068)
			return includeConstraints == catalog.IgnoreConstraintsAndPKSwaps == true
		}() == true {
			__antithesis_instrumentation__.Notify(270069)
			continue
		} else {
			__antithesis_instrumentation__.Notify(270070)
			if mutation.GetConstraint() != nil && func() bool {
				__antithesis_instrumentation__.Notify(270071)
				return includeConstraints > catalog.IncludeConstraints == true
			}() == true {
				__antithesis_instrumentation__.Notify(270072)
				continue
			} else {
				__antithesis_instrumentation__.Notify(270073)
			}
		}
		__antithesis_instrumentation__.Notify(270065)
		if err := table.MakeMutationComplete(mutation); err != nil {
			__antithesis_instrumentation__.Notify(270074)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(270075)
		}
	}
	__antithesis_instrumentation__.Notify(270062)
	table.Mutations = table.Mutations[i:]
	table.Version++
	return table, nil
}

func (desc *wrapper) MakePublic() catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(270076)

	table := desc.NewBuilder().(TableDescriptorBuilder).BuildExistingMutableTable()
	table.State = descpb.DescriptorState_PUBLIC
	table.Version++
	return table
}

func (desc *wrapper) HasPrimaryKey() bool {
	__antithesis_instrumentation__.Notify(270077)
	return !desc.PrimaryIndex.Disabled
}

func (desc *wrapper) HasColumnBackfillMutation() bool {
	__antithesis_instrumentation__.Notify(270078)
	for _, m := range desc.AllMutations() {
		__antithesis_instrumentation__.Notify(270080)
		if col := m.AsColumn(); col != nil {
			__antithesis_instrumentation__.Notify(270081)
			if catalog.ColumnNeedsBackfill(col) {
				__antithesis_instrumentation__.Notify(270082)
				return true
			} else {
				__antithesis_instrumentation__.Notify(270083)
			}
		} else {
			__antithesis_instrumentation__.Notify(270084)
		}
	}
	__antithesis_instrumentation__.Notify(270079)
	return false
}

func (desc *Mutable) IsNew() bool {
	__antithesis_instrumentation__.Notify(270085)
	return desc.ClusterVersion().ID == descpb.InvalidID
}

func ColumnsSelectors(cols []catalog.Column) tree.SelectExprs {
	__antithesis_instrumentation__.Notify(270086)
	exprs := make(tree.SelectExprs, len(cols))
	colItems := make([]tree.ColumnItem, len(cols))
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(270088)
		colItems[i].ColumnName = col.ColName()
		exprs[i].Expr = &colItems[i]
	}
	__antithesis_instrumentation__.Notify(270087)
	return exprs
}

func (desc *wrapper) InvalidateFKConstraints() {
	__antithesis_instrumentation__.Notify(270089)

	for i := range desc.OutboundFKs {
		__antithesis_instrumentation__.Notify(270090)
		fk := &desc.OutboundFKs[i]
		fk.Validity = descpb.ConstraintValidity_Unvalidated
	}
}

func (desc *wrapper) AllIndexSpans(codec keys.SQLCodec) roachpb.Spans {
	__antithesis_instrumentation__.Notify(270091)
	var spans roachpb.Spans
	for _, index := range desc.NonDropIndexes() {
		__antithesis_instrumentation__.Notify(270093)
		spans = append(spans, desc.IndexSpan(codec, index.GetID()))
	}
	__antithesis_instrumentation__.Notify(270092)
	return spans
}

func (desc *wrapper) PrimaryIndexSpan(codec keys.SQLCodec) roachpb.Span {
	__antithesis_instrumentation__.Notify(270094)
	return desc.IndexSpan(codec, desc.PrimaryIndex.ID)
}

func (desc *wrapper) IndexSpan(codec keys.SQLCodec, indexID descpb.IndexID) roachpb.Span {
	__antithesis_instrumentation__.Notify(270095)
	prefix := roachpb.Key(rowenc.MakeIndexKeyPrefix(codec, desc.GetID(), indexID))
	return roachpb.Span{Key: prefix, EndKey: prefix.PrefixEnd()}
}

func (desc *wrapper) TableSpan(codec keys.SQLCodec) roachpb.Span {
	__antithesis_instrumentation__.Notify(270096)
	prefix := codec.TablePrefix(uint32(desc.ID))
	return roachpb.Span{Key: prefix, EndKey: prefix.PrefixEnd()}
}

func (desc *wrapper) ColumnsUsed(
	cc *descpb.TableDescriptor_CheckConstraint,
) ([]descpb.ColumnID, error) {
	__antithesis_instrumentation__.Notify(270097)
	if len(cc.ColumnIDs) > 0 {
		__antithesis_instrumentation__.Notify(270103)

		return cc.ColumnIDs, nil
	} else {
		__antithesis_instrumentation__.Notify(270104)
	}
	__antithesis_instrumentation__.Notify(270098)

	parsed, err := parser.ParseExpr(cc.Expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(270105)
		return nil, pgerror.Wrapf(err, pgcode.Syntax,
			"could not parse check constraint %s", cc.Expr)
	} else {
		__antithesis_instrumentation__.Notify(270106)
	}
	__antithesis_instrumentation__.Notify(270099)

	var colIDsUsed catalog.TableColSet
	visitFn := func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(270107)
		if vBase, ok := expr.(tree.VarName); ok {
			__antithesis_instrumentation__.Notify(270109)
			v, err := vBase.NormalizeVarName()
			if err != nil {
				__antithesis_instrumentation__.Notify(270112)
				return false, nil, err
			} else {
				__antithesis_instrumentation__.Notify(270113)
			}
			__antithesis_instrumentation__.Notify(270110)
			if c, ok := v.(*tree.ColumnItem); ok {
				__antithesis_instrumentation__.Notify(270114)
				col, err := desc.FindColumnWithName(c.ColumnName)
				if err != nil || func() bool {
					__antithesis_instrumentation__.Notify(270116)
					return col.Dropped() == true
				}() == true {
					__antithesis_instrumentation__.Notify(270117)
					return false, nil, pgerror.Newf(pgcode.UndefinedColumn,
						"column %q not found for constraint %q",
						c.ColumnName, parsed.String())
				} else {
					__antithesis_instrumentation__.Notify(270118)
				}
				__antithesis_instrumentation__.Notify(270115)
				colIDsUsed.Add(col.GetID())
			} else {
				__antithesis_instrumentation__.Notify(270119)
			}
			__antithesis_instrumentation__.Notify(270111)
			return false, v, nil
		} else {
			__antithesis_instrumentation__.Notify(270120)
		}
		__antithesis_instrumentation__.Notify(270108)
		return true, expr, nil
	}
	__antithesis_instrumentation__.Notify(270100)
	if _, err := tree.SimpleVisit(parsed, visitFn); err != nil {
		__antithesis_instrumentation__.Notify(270121)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(270122)
	}
	__antithesis_instrumentation__.Notify(270101)

	cc.ColumnIDs = make([]descpb.ColumnID, 0, colIDsUsed.Len())
	for colID, ok := colIDsUsed.Next(0); ok; colID, ok = colIDsUsed.Next(colID + 1) {
		__antithesis_instrumentation__.Notify(270123)
		cc.ColumnIDs = append(cc.ColumnIDs, colID)
	}
	__antithesis_instrumentation__.Notify(270102)
	sort.Sort(descpb.ColumnIDs(cc.ColumnIDs))
	return cc.ColumnIDs, nil
}

func (desc *wrapper) CheckConstraintUsesColumn(
	cc *descpb.TableDescriptor_CheckConstraint, colID descpb.ColumnID,
) (bool, error) {
	__antithesis_instrumentation__.Notify(270124)
	colsUsed, err := desc.ColumnsUsed(cc)
	if err != nil {
		__antithesis_instrumentation__.Notify(270127)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(270128)
	}
	__antithesis_instrumentation__.Notify(270125)
	i := sort.Search(len(colsUsed), func(i int) bool {
		__antithesis_instrumentation__.Notify(270129)
		return colsUsed[i] >= colID
	})
	__antithesis_instrumentation__.Notify(270126)
	return i < len(colsUsed) && func() bool {
		__antithesis_instrumentation__.Notify(270130)
		return colsUsed[i] == colID == true
	}() == true, nil
}

func (desc *wrapper) GetFamilyOfColumn(
	colID descpb.ColumnID,
) (*descpb.ColumnFamilyDescriptor, error) {
	__antithesis_instrumentation__.Notify(270131)
	for _, fam := range desc.Families {
		__antithesis_instrumentation__.Notify(270133)
		for _, id := range fam.ColumnIDs {
			__antithesis_instrumentation__.Notify(270134)
			if id == colID {
				__antithesis_instrumentation__.Notify(270135)
				return &fam, nil
			} else {
				__antithesis_instrumentation__.Notify(270136)
			}
		}
	}
	__antithesis_instrumentation__.Notify(270132)

	return nil, errors.Newf("no column family found for column id %v", colID)
}

func (desc *Mutable) SetAuditMode(mode tree.AuditMode) (bool, error) {
	__antithesis_instrumentation__.Notify(270137)
	prev := desc.AuditMode
	switch mode {
	case tree.AuditModeDisable:
		__antithesis_instrumentation__.Notify(270139)
		desc.AuditMode = descpb.TableDescriptor_DISABLED
	case tree.AuditModeReadWrite:
		__antithesis_instrumentation__.Notify(270140)
		desc.AuditMode = descpb.TableDescriptor_READWRITE
	default:
		__antithesis_instrumentation__.Notify(270141)
		return false, pgerror.Newf(pgcode.InvalidParameterValue,
			"unknown audit mode: %s (%d)", mode, mode)
	}
	__antithesis_instrumentation__.Notify(270138)
	return prev != desc.AuditMode, nil
}

func (desc *wrapper) FindAllReferences() (map[descpb.ID]struct{}, error) {
	__antithesis_instrumentation__.Notify(270142)
	refs := map[descpb.ID]struct{}{}
	for i := range desc.OutboundFKs {
		__antithesis_instrumentation__.Notify(270148)
		fk := &desc.OutboundFKs[i]
		refs[fk.ReferencedTableID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(270143)
	for i := range desc.InboundFKs {
		__antithesis_instrumentation__.Notify(270149)
		fk := &desc.InboundFKs[i]
		refs[fk.OriginTableID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(270144)

	for _, c := range desc.NonDropColumns() {
		__antithesis_instrumentation__.Notify(270150)
		for i := 0; i < c.NumUsesSequences(); i++ {
			__antithesis_instrumentation__.Notify(270151)
			id := c.GetUsesSequenceID(i)
			refs[id] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(270145)

	for _, dest := range desc.DependsOn {
		__antithesis_instrumentation__.Notify(270152)
		refs[dest] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(270146)

	for _, c := range desc.DependedOnBy {
		__antithesis_instrumentation__.Notify(270153)
		refs[c.ID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(270147)
	return refs, nil
}

func (desc *immutable) ActiveChecks() []descpb.TableDescriptor_CheckConstraint {
	__antithesis_instrumentation__.Notify(270154)
	return desc.allChecks
}

func (desc *wrapper) IsShardColumn(col catalog.Column) bool {
	__antithesis_instrumentation__.Notify(270155)
	return nil != catalog.FindNonDropIndex(desc, func(idx catalog.Index) bool {
		__antithesis_instrumentation__.Notify(270156)
		return idx.IsSharded() && func() bool {
			__antithesis_instrumentation__.Notify(270157)
			return idx.GetShardColumnName() == col.GetName() == true
		}() == true
	})
}

func (desc *wrapper) TableDesc() *descpb.TableDescriptor {
	__antithesis_instrumentation__.Notify(270158)
	return &desc.TableDescriptor
}

func GenerateUniqueName(prefix string, nameExistsFunc func(name string) bool) string {
	__antithesis_instrumentation__.Notify(270159)
	name := prefix
	for i := 1; nameExistsFunc(name); i++ {
		__antithesis_instrumentation__.Notify(270161)
		name = fmt.Sprintf("%s_%d", prefix, i)
	}
	__antithesis_instrumentation__.Notify(270160)
	return name
}

func (desc *Mutable) SetParentSchemaID(schemaID descpb.ID) {
	__antithesis_instrumentation__.Notify(270162)
	desc.UnexposedParentSchemaID = schemaID
}

func (desc *Mutable) AddDrainingName(name descpb.NameInfo) {
	__antithesis_instrumentation__.Notify(270163)
	desc.DrainingNames = append(desc.DrainingNames, name)
}

func (desc *Mutable) SetPublic() {
	__antithesis_instrumentation__.Notify(270164)
	desc.State = descpb.DescriptorState_PUBLIC
	desc.OfflineReason = ""
}

func (desc *Mutable) SetDropped() {
	__antithesis_instrumentation__.Notify(270165)
	desc.State = descpb.DescriptorState_DROP
	desc.OfflineReason = ""
}

func (desc *Mutable) SetOffline(reason string) {
	__antithesis_instrumentation__.Notify(270166)
	desc.State = descpb.DescriptorState_OFFLINE
	desc.OfflineReason = reason
}

func (desc *wrapper) IsLocalityRegionalByRow() bool {
	__antithesis_instrumentation__.Notify(270167)
	return desc.LocalityConfig.GetRegionalByRow() != nil
}

func (desc *wrapper) IsLocalityRegionalByTable() bool {
	__antithesis_instrumentation__.Notify(270168)
	return desc.LocalityConfig.GetRegionalByTable() != nil
}

func (desc *wrapper) IsLocalityGlobal() bool {
	__antithesis_instrumentation__.Notify(270169)
	return desc.LocalityConfig.GetGlobal() != nil
}

func (desc *wrapper) GetRegionalByTableRegion() (catpb.RegionName, error) {
	__antithesis_instrumentation__.Notify(270170)
	if !desc.IsLocalityRegionalByTable() {
		__antithesis_instrumentation__.Notify(270173)
		return "", errors.AssertionFailedf("%s is not REGIONAL BY TABLE", desc.Name)
	} else {
		__antithesis_instrumentation__.Notify(270174)
	}
	__antithesis_instrumentation__.Notify(270171)
	region := desc.LocalityConfig.GetRegionalByTable().Region
	if region == nil {
		__antithesis_instrumentation__.Notify(270175)
		return catpb.RegionName(tree.PrimaryRegionNotSpecifiedName), nil
	} else {
		__antithesis_instrumentation__.Notify(270176)
	}
	__antithesis_instrumentation__.Notify(270172)
	return *region, nil
}

func (desc *wrapper) GetRegionalByRowTableRegionColumnName() (tree.Name, error) {
	__antithesis_instrumentation__.Notify(270177)
	if !desc.IsLocalityRegionalByRow() {
		__antithesis_instrumentation__.Notify(270180)
		return "", errors.AssertionFailedf("%q is not a REGIONAL BY ROW table", desc.Name)
	} else {
		__antithesis_instrumentation__.Notify(270181)
	}
	__antithesis_instrumentation__.Notify(270178)
	colName := desc.LocalityConfig.GetRegionalByRow().As
	if colName == nil {
		__antithesis_instrumentation__.Notify(270182)
		return tree.RegionalByRowRegionDefaultColName, nil
	} else {
		__antithesis_instrumentation__.Notify(270183)
	}
	__antithesis_instrumentation__.Notify(270179)
	return tree.Name(*colName), nil
}

func (desc *wrapper) GetRowLevelTTL() *catpb.RowLevelTTL {
	__antithesis_instrumentation__.Notify(270184)
	return desc.RowLevelTTL
}

func (desc *wrapper) HasRowLevelTTL() bool {
	__antithesis_instrumentation__.Notify(270185)
	return desc.RowLevelTTL != nil
}

func (desc *wrapper) GetExcludeDataFromBackup() bool {
	__antithesis_instrumentation__.Notify(270186)
	return desc.ExcludeDataFromBackup
}

func (desc *wrapper) GetStorageParams(spaceBetweenEqual bool) []string {
	__antithesis_instrumentation__.Notify(270187)
	var storageParams []string
	var spacing string
	if spaceBetweenEqual {
		__antithesis_instrumentation__.Notify(270192)
		spacing = ` `
	} else {
		__antithesis_instrumentation__.Notify(270193)
	}
	__antithesis_instrumentation__.Notify(270188)
	appendStorageParam := func(key, value string) {
		__antithesis_instrumentation__.Notify(270194)
		storageParams = append(storageParams, key+spacing+`=`+spacing+value)
	}
	__antithesis_instrumentation__.Notify(270189)
	if ttl := desc.GetRowLevelTTL(); ttl != nil {
		__antithesis_instrumentation__.Notify(270195)
		appendStorageParam(`ttl`, `'on'`)
		appendStorageParam(`ttl_automatic_column`, `'on'`)
		appendStorageParam(`ttl_expire_after`, string(ttl.DurationExpr))
		appendStorageParam(`ttl_job_cron`, fmt.Sprintf(`'%s'`, ttl.DeletionCronOrDefault()))
		if bs := ttl.SelectBatchSize; bs != 0 {
			__antithesis_instrumentation__.Notify(270202)
			appendStorageParam(`ttl_select_batch_size`, fmt.Sprintf(`%d`, bs))
		} else {
			__antithesis_instrumentation__.Notify(270203)
		}
		__antithesis_instrumentation__.Notify(270196)
		if bs := ttl.DeleteBatchSize; bs != 0 {
			__antithesis_instrumentation__.Notify(270204)
			appendStorageParam(`ttl_delete_batch_size`, fmt.Sprintf(`%d`, bs))
		} else {
			__antithesis_instrumentation__.Notify(270205)
		}
		__antithesis_instrumentation__.Notify(270197)
		if rc := ttl.RangeConcurrency; rc != 0 {
			__antithesis_instrumentation__.Notify(270206)
			appendStorageParam(`ttl_range_concurrency`, fmt.Sprintf(`%d`, rc))
		} else {
			__antithesis_instrumentation__.Notify(270207)
		}
		__antithesis_instrumentation__.Notify(270198)
		if rl := ttl.DeleteRateLimit; rl != 0 {
			__antithesis_instrumentation__.Notify(270208)
			appendStorageParam(`ttl_delete_rate_limit`, fmt.Sprintf(`%d`, rl))
		} else {
			__antithesis_instrumentation__.Notify(270209)
		}
		__antithesis_instrumentation__.Notify(270199)
		if pause := ttl.Pause; pause {
			__antithesis_instrumentation__.Notify(270210)
			appendStorageParam(`ttl_pause`, fmt.Sprintf(`%t`, pause))
		} else {
			__antithesis_instrumentation__.Notify(270211)
		}
		__antithesis_instrumentation__.Notify(270200)
		if p := ttl.RowStatsPollInterval; p != 0 {
			__antithesis_instrumentation__.Notify(270212)
			appendStorageParam(`ttl_row_stats_poll_interval`, fmt.Sprintf(`'%s'`, p.String()))
		} else {
			__antithesis_instrumentation__.Notify(270213)
		}
		__antithesis_instrumentation__.Notify(270201)
		if labelMetrics := ttl.LabelMetrics; labelMetrics {
			__antithesis_instrumentation__.Notify(270214)
			appendStorageParam(`ttl_label_metrics`, fmt.Sprintf(`%t`, labelMetrics))
		} else {
			__antithesis_instrumentation__.Notify(270215)
		}
	} else {
		__antithesis_instrumentation__.Notify(270216)
	}
	__antithesis_instrumentation__.Notify(270190)
	if exclude := desc.GetExcludeDataFromBackup(); exclude {
		__antithesis_instrumentation__.Notify(270217)
		appendStorageParam(`exclude_data_from_backup`, `true`)
	} else {
		__antithesis_instrumentation__.Notify(270218)
	}
	__antithesis_instrumentation__.Notify(270191)
	return storageParams
}

func (desc *wrapper) GetMultiRegionEnumDependencyIfExists() bool {
	__antithesis_instrumentation__.Notify(270219)
	if desc.IsLocalityRegionalByTable() {
		__antithesis_instrumentation__.Notify(270221)
		regionName, _ := desc.GetRegionalByTableRegion()
		return regionName != catpb.RegionName(tree.PrimaryRegionNotSpecifiedName)
	} else {
		__antithesis_instrumentation__.Notify(270222)
	}
	__antithesis_instrumentation__.Notify(270220)
	return false
}

func (desc *Mutable) SetTableLocalityRegionalByTable(region tree.Name) {
	__antithesis_instrumentation__.Notify(270223)
	lc := LocalityConfigRegionalByTable(region)
	desc.LocalityConfig = &lc
}

func LocalityConfigRegionalByTable(region tree.Name) catpb.LocalityConfig {
	__antithesis_instrumentation__.Notify(270224)
	l := &catpb.LocalityConfig_RegionalByTable_{
		RegionalByTable: &catpb.LocalityConfig_RegionalByTable{},
	}
	if region != tree.PrimaryRegionNotSpecifiedName {
		__antithesis_instrumentation__.Notify(270226)
		regionName := catpb.RegionName(region)
		l.RegionalByTable.Region = &regionName
	} else {
		__antithesis_instrumentation__.Notify(270227)
	}
	__antithesis_instrumentation__.Notify(270225)
	return catpb.LocalityConfig{Locality: l}
}

func (desc *Mutable) SetTableLocalityRegionalByRow(regionColName tree.Name) {
	__antithesis_instrumentation__.Notify(270228)
	lc := LocalityConfigRegionalByRow(regionColName)
	desc.LocalityConfig = &lc
}

func LocalityConfigRegionalByRow(regionColName tree.Name) catpb.LocalityConfig {
	__antithesis_instrumentation__.Notify(270229)
	rbr := &catpb.LocalityConfig_RegionalByRow{}
	if regionColName != tree.RegionalByRowRegionNotSpecifiedName {
		__antithesis_instrumentation__.Notify(270231)
		rbr.As = (*string)(&regionColName)
	} else {
		__antithesis_instrumentation__.Notify(270232)
	}
	__antithesis_instrumentation__.Notify(270230)
	return catpb.LocalityConfig{
		Locality: &catpb.LocalityConfig_RegionalByRow_{
			RegionalByRow: rbr,
		},
	}
}

func (desc *Mutable) SetTableLocalityGlobal() {
	__antithesis_instrumentation__.Notify(270233)
	lc := LocalityConfigGlobal()
	desc.LocalityConfig = &lc
}

func (desc *Mutable) SetDeclarativeSchemaChangerState(state *scpb.DescriptorState) {
	__antithesis_instrumentation__.Notify(270234)
	desc.DeclarativeSchemaChangerState = state
}

func LocalityConfigGlobal() catpb.LocalityConfig {
	__antithesis_instrumentation__.Notify(270235)
	return catpb.LocalityConfig{
		Locality: &catpb.LocalityConfig_Global_{
			Global: &catpb.LocalityConfig_Global{},
		},
	}
}

func PrimaryKeyIndexName(tableName string) string {
	__antithesis_instrumentation__.Notify(270236)
	return tableName + "_pkey"
}

func (desc *Mutable) UpdateColumnsDependedOnBy(id descpb.ID, colIDs catalog.TableColSet) {
	__antithesis_instrumentation__.Notify(270237)
	ref := descpb.TableDescriptor_Reference{
		ID:        id,
		ColumnIDs: colIDs.Ordered(),
		ByID:      true,
	}
	for i := range desc.DependedOnBy {
		__antithesis_instrumentation__.Notify(270239)
		by := &desc.DependedOnBy[i]
		if by.ID == id {
			__antithesis_instrumentation__.Notify(270240)
			if colIDs.Empty() {
				__antithesis_instrumentation__.Notify(270242)
				desc.DependedOnBy = append(desc.DependedOnBy[:i], desc.DependedOnBy[i+1:]...)
				return
			} else {
				__antithesis_instrumentation__.Notify(270243)
			}
			__antithesis_instrumentation__.Notify(270241)
			*by = ref
			return
		} else {
			__antithesis_instrumentation__.Notify(270244)
		}
	}
	__antithesis_instrumentation__.Notify(270238)
	desc.DependedOnBy = append(desc.DependedOnBy, ref)
}
