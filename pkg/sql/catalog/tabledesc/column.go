package tabledesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

var _ catalog.Column = (*column)(nil)

type column struct {
	maybeMutation
	desc    *descpb.ColumnDescriptor
	ordinal int
}

func (w column) ColumnDesc() *descpb.ColumnDescriptor {
	__antithesis_instrumentation__.Notify(268592)
	return w.desc
}

func (w column) ColumnDescDeepCopy() descpb.ColumnDescriptor {
	__antithesis_instrumentation__.Notify(268593)
	return *protoutil.Clone(w.desc).(*descpb.ColumnDescriptor)
}

func (w column) DeepCopy() catalog.Column {
	__antithesis_instrumentation__.Notify(268594)
	desc := w.ColumnDescDeepCopy()
	return &column{
		maybeMutation: w.maybeMutation,
		desc:          &desc,
		ordinal:       w.ordinal,
	}
}

func (w column) Ordinal() int {
	__antithesis_instrumentation__.Notify(268595)
	return w.ordinal
}

func (w column) Public() bool {
	__antithesis_instrumentation__.Notify(268596)
	return !w.IsMutation() && func() bool {
		__antithesis_instrumentation__.Notify(268597)
		return !w.IsSystemColumn() == true
	}() == true
}

func (w column) GetID() descpb.ColumnID {
	__antithesis_instrumentation__.Notify(268598)
	return w.desc.ID
}

func (w column) GetName() string {
	__antithesis_instrumentation__.Notify(268599)
	return w.desc.Name
}

func (w column) ColName() tree.Name {
	__antithesis_instrumentation__.Notify(268600)
	return w.desc.ColName()
}

func (w column) HasType() bool {
	__antithesis_instrumentation__.Notify(268601)
	return w.desc.Type != nil
}

func (w column) GetType() *types.T {
	__antithesis_instrumentation__.Notify(268602)
	return w.desc.Type
}

func (w column) IsNullable() bool {
	__antithesis_instrumentation__.Notify(268603)
	return w.desc.Nullable
}

func (w column) HasDefault() bool {
	__antithesis_instrumentation__.Notify(268604)
	return w.desc.HasDefault()
}

func (w column) GetDefaultExpr() string {
	__antithesis_instrumentation__.Notify(268605)
	if !w.HasDefault() {
		__antithesis_instrumentation__.Notify(268607)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(268608)
	}
	__antithesis_instrumentation__.Notify(268606)
	return *w.desc.DefaultExpr
}

func (w column) HasOnUpdate() bool {
	__antithesis_instrumentation__.Notify(268609)
	return w.desc.HasOnUpdate()
}

func (w column) GetOnUpdateExpr() string {
	__antithesis_instrumentation__.Notify(268610)
	if !w.HasOnUpdate() {
		__antithesis_instrumentation__.Notify(268612)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(268613)
	}
	__antithesis_instrumentation__.Notify(268611)
	return *w.desc.OnUpdateExpr
}

func (w column) IsComputed() bool {
	__antithesis_instrumentation__.Notify(268614)
	return w.desc.IsComputed()
}

func (w column) GetComputeExpr() string {
	__antithesis_instrumentation__.Notify(268615)
	if !w.IsComputed() {
		__antithesis_instrumentation__.Notify(268617)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(268618)
	}
	__antithesis_instrumentation__.Notify(268616)
	return *w.desc.ComputeExpr
}

func (w column) IsHidden() bool {
	__antithesis_instrumentation__.Notify(268619)
	return w.desc.Hidden
}

func (w column) IsInaccessible() bool {
	__antithesis_instrumentation__.Notify(268620)
	return w.desc.Inaccessible
}

func (w column) IsExpressionIndexColumn() bool {
	__antithesis_instrumentation__.Notify(268621)
	return w.IsInaccessible() && func() bool {
		__antithesis_instrumentation__.Notify(268622)
		return w.IsVirtual() == true
	}() == true
}

func (w column) NumUsesSequences() int {
	__antithesis_instrumentation__.Notify(268623)
	return len(w.desc.UsesSequenceIds)
}

func (w column) GetUsesSequenceID(usesSequenceOrdinal int) descpb.ID {
	__antithesis_instrumentation__.Notify(268624)
	return w.desc.UsesSequenceIds[usesSequenceOrdinal]
}

func (w column) NumOwnsSequences() int {
	__antithesis_instrumentation__.Notify(268625)
	return len(w.desc.OwnsSequenceIds)
}

func (w column) GetOwnsSequenceID(ownsSequenceOrdinal int) descpb.ID {
	__antithesis_instrumentation__.Notify(268626)
	return w.desc.OwnsSequenceIds[ownsSequenceOrdinal]
}

func (w column) IsVirtual() bool {
	__antithesis_instrumentation__.Notify(268627)
	return w.desc.Virtual
}

func (w column) CheckCanBeInboundFKRef() error {
	__antithesis_instrumentation__.Notify(268628)
	return w.desc.CheckCanBeInboundFKRef()
}

func (w column) CheckCanBeOutboundFKRef() error {
	__antithesis_instrumentation__.Notify(268629)
	return w.desc.CheckCanBeOutboundFKRef()
}

func (w column) GetPGAttributeNum() uint32 {
	__antithesis_instrumentation__.Notify(268630)
	return w.desc.GetPGAttributeNum()
}

func (w column) IsSystemColumn() bool {
	__antithesis_instrumentation__.Notify(268631)
	return w.desc.SystemColumnKind != catpb.SystemColumnKind_NONE
}

func (w column) IsGeneratedAsIdentity() bool {
	__antithesis_instrumentation__.Notify(268632)
	return w.desc.GeneratedAsIdentityType != catpb.GeneratedAsIdentityType_NOT_IDENTITY_COLUMN
}

func (w column) IsGeneratedAlwaysAsIdentity() bool {
	__antithesis_instrumentation__.Notify(268633)
	return w.desc.GeneratedAsIdentityType == catpb.GeneratedAsIdentityType_GENERATED_ALWAYS
}

func (w column) IsGeneratedByDefaultAsIdentity() bool {
	__antithesis_instrumentation__.Notify(268634)
	return w.desc.GeneratedAsIdentityType == catpb.GeneratedAsIdentityType_GENERATED_BY_DEFAULT
}

func (w column) GetGeneratedAsIdentityType() catpb.GeneratedAsIdentityType {
	__antithesis_instrumentation__.Notify(268635)
	return w.desc.GeneratedAsIdentityType
}

func (w column) GetGeneratedAsIdentitySequenceOption() string {
	__antithesis_instrumentation__.Notify(268636)
	if !w.HasGeneratedAsIdentitySequenceOption() {
		__antithesis_instrumentation__.Notify(268638)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(268639)
	}
	__antithesis_instrumentation__.Notify(268637)
	return strings.TrimSpace(*w.desc.GeneratedAsIdentitySequenceOption)
}

func (w column) HasGeneratedAsIdentitySequenceOption() bool {
	__antithesis_instrumentation__.Notify(268640)
	return w.desc.GeneratedAsIdentitySequenceOption != nil
}

type columnCache struct {
	all                  []catalog.Column
	public               []catalog.Column
	writable             []catalog.Column
	deletable            []catalog.Column
	nonDrop              []catalog.Column
	visible              []catalog.Column
	accessible           []catalog.Column
	readable             []catalog.Column
	withUDTs             []catalog.Column
	system               []catalog.Column
	familyDefaultColumns []descpb.IndexFetchSpec_FamilyDefaultColumn
	index                []indexColumnCache
}

type indexColumnCache struct {
	all          []catalog.Column
	allDirs      []descpb.IndexDescriptor_Direction
	key          []catalog.Column
	keyDirs      []descpb.IndexDescriptor_Direction
	stored       []catalog.Column
	keySuffix    []catalog.Column
	full         []catalog.Column
	fullDirs     []descpb.IndexDescriptor_Direction
	keyAndSuffix []descpb.IndexFetchSpec_KeyColumn
}

func newColumnCache(desc *descpb.TableDescriptor, mutations *mutationCache) *columnCache {
	__antithesis_instrumentation__.Notify(268641)
	c := columnCache{}

	numPublic := len(desc.Columns)
	backingStructs := make([]column, numPublic, numPublic+len(colinfo.AllSystemColumnDescs))
	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(268652)
		backingStructs[i] = column{desc: &desc.Columns[i], ordinal: i}
	}
	__antithesis_instrumentation__.Notify(268642)
	numMutations := len(mutations.columns)
	numDeletable := numPublic + numMutations
	for i := range colinfo.AllSystemColumnDescs {
		__antithesis_instrumentation__.Notify(268653)
		col := column{
			desc:    &colinfo.AllSystemColumnDescs[i],
			ordinal: numDeletable + i,
		}
		backingStructs = append(backingStructs, col)
	}
	__antithesis_instrumentation__.Notify(268643)

	c.all = make([]catalog.Column, 0, numDeletable+len(colinfo.AllSystemColumnDescs))
	for i := range backingStructs[:numPublic] {
		__antithesis_instrumentation__.Notify(268654)
		c.all = append(c.all, &backingStructs[i])
	}
	__antithesis_instrumentation__.Notify(268644)
	for _, m := range mutations.columns {
		__antithesis_instrumentation__.Notify(268655)
		c.all = append(c.all, m.AsColumn())
	}
	__antithesis_instrumentation__.Notify(268645)
	for i := range backingStructs[numPublic:] {
		__antithesis_instrumentation__.Notify(268656)
		c.all = append(c.all, &backingStructs[numPublic+i])
	}
	__antithesis_instrumentation__.Notify(268646)

	c.deletable = c.all[:numDeletable]
	c.system = c.all[numDeletable:]
	c.public = c.all[:numPublic]
	if numMutations == 0 {
		__antithesis_instrumentation__.Notify(268657)
		c.readable = c.public
		c.writable = c.public
		c.nonDrop = c.public
	} else {
		__antithesis_instrumentation__.Notify(268658)
		for _, col := range c.deletable {
			__antithesis_instrumentation__.Notify(268659)
			if !col.DeleteOnly() {
				__antithesis_instrumentation__.Notify(268662)
				lazyAllocAppendColumn(&c.writable, col, numDeletable)
			} else {
				__antithesis_instrumentation__.Notify(268663)
			}
			__antithesis_instrumentation__.Notify(268660)
			if !col.Dropped() {
				__antithesis_instrumentation__.Notify(268664)
				lazyAllocAppendColumn(&c.nonDrop, col, numDeletable)
			} else {
				__antithesis_instrumentation__.Notify(268665)
			}
			__antithesis_instrumentation__.Notify(268661)
			lazyAllocAppendColumn(&c.readable, col, numDeletable)
		}
	}
	__antithesis_instrumentation__.Notify(268647)
	for _, col := range c.deletable {
		__antithesis_instrumentation__.Notify(268666)
		if col.Public() && func() bool {
			__antithesis_instrumentation__.Notify(268669)
			return !col.IsHidden() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(268670)
			return !col.IsInaccessible() == true
		}() == true {
			__antithesis_instrumentation__.Notify(268671)
			lazyAllocAppendColumn(&c.visible, col, numPublic)
		} else {
			__antithesis_instrumentation__.Notify(268672)
		}
		__antithesis_instrumentation__.Notify(268667)
		if col.Public() && func() bool {
			__antithesis_instrumentation__.Notify(268673)
			return !col.IsInaccessible() == true
		}() == true {
			__antithesis_instrumentation__.Notify(268674)
			lazyAllocAppendColumn(&c.accessible, col, numPublic)
		} else {
			__antithesis_instrumentation__.Notify(268675)
		}
		__antithesis_instrumentation__.Notify(268668)
		if col.HasType() && func() bool {
			__antithesis_instrumentation__.Notify(268676)
			return col.GetType().UserDefined() == true
		}() == true {
			__antithesis_instrumentation__.Notify(268677)
			lazyAllocAppendColumn(&c.withUDTs, col, numDeletable)
		} else {
			__antithesis_instrumentation__.Notify(268678)
		}
	}
	__antithesis_instrumentation__.Notify(268648)

	for i := range desc.Families {
		__antithesis_instrumentation__.Notify(268679)
		if f := &desc.Families[i]; f.DefaultColumnID != 0 {
			__antithesis_instrumentation__.Notify(268680)
			if c.familyDefaultColumns == nil {
				__antithesis_instrumentation__.Notify(268682)
				c.familyDefaultColumns = make([]descpb.IndexFetchSpec_FamilyDefaultColumn, 0, len(desc.Families)-i)
			} else {
				__antithesis_instrumentation__.Notify(268683)
			}
			__antithesis_instrumentation__.Notify(268681)
			c.familyDefaultColumns = append(c.familyDefaultColumns, descpb.IndexFetchSpec_FamilyDefaultColumn{
				FamilyID:        f.ID,
				DefaultColumnID: f.DefaultColumnID,
			})
		} else {
			__antithesis_instrumentation__.Notify(268684)
		}
	}
	__antithesis_instrumentation__.Notify(268649)

	c.index = make([]indexColumnCache, 0, 1+len(desc.Indexes)+len(mutations.indexes))
	c.index = append(c.index, makeIndexColumnCache(&desc.PrimaryIndex, c.all))
	for i := range desc.Indexes {
		__antithesis_instrumentation__.Notify(268685)
		c.index = append(c.index, makeIndexColumnCache(&desc.Indexes[i], c.all))
	}
	__antithesis_instrumentation__.Notify(268650)
	for i := range mutations.indexes {
		__antithesis_instrumentation__.Notify(268686)
		c.index = append(c.index, makeIndexColumnCache(mutations.indexes[i].AsIndex().IndexDesc(), c.all))
	}
	__antithesis_instrumentation__.Notify(268651)
	return &c
}

func makeIndexColumnCache(idx *descpb.IndexDescriptor, all []catalog.Column) (ic indexColumnCache) {
	__antithesis_instrumentation__.Notify(268687)
	nKey := len(idx.KeyColumnIDs)
	nKeySuffix := len(idx.KeySuffixColumnIDs)
	nStored := len(idx.StoreColumnIDs)
	nAll := nKey + nKeySuffix + nStored
	ic.allDirs = make([]descpb.IndexDescriptor_Direction, nAll)

	copy(ic.allDirs, idx.KeyColumnDirections)
	ic.all = make([]catalog.Column, 0, nAll)
	appendColumnsByID(&ic.all, all, idx.KeyColumnIDs)
	appendColumnsByID(&ic.all, all, idx.KeySuffixColumnIDs)
	appendColumnsByID(&ic.all, all, idx.StoreColumnIDs)
	ic.key = ic.all[:nKey]
	ic.keyDirs = ic.allDirs[:nKey]
	ic.keySuffix = ic.all[nKey : nKey+nKeySuffix]
	ic.stored = ic.all[nKey+nKeySuffix:]
	nFull := nKey
	if !idx.Unique {
		__antithesis_instrumentation__.Notify(268692)
		nFull = nFull + nKeySuffix
	} else {
		__antithesis_instrumentation__.Notify(268693)
	}
	__antithesis_instrumentation__.Notify(268688)
	ic.full = ic.all[:nFull]
	ic.fullDirs = ic.allDirs[:nFull]

	var invertedColumnID descpb.ColumnID
	if nKey > 0 && func() bool {
		__antithesis_instrumentation__.Notify(268694)
		return idx.Type == descpb.IndexDescriptor_INVERTED == true
	}() == true {
		__antithesis_instrumentation__.Notify(268695)
		invertedColumnID = idx.InvertedColumnID()
	} else {
		__antithesis_instrumentation__.Notify(268696)
	}
	__antithesis_instrumentation__.Notify(268689)
	var compositeIDs catalog.TableColSet
	for _, colID := range idx.CompositeColumnIDs {
		__antithesis_instrumentation__.Notify(268697)
		compositeIDs.Add(colID)
	}
	__antithesis_instrumentation__.Notify(268690)
	ic.keyAndSuffix = make([]descpb.IndexFetchSpec_KeyColumn, nKey+nKeySuffix)
	for i := range ic.keyAndSuffix {
		__antithesis_instrumentation__.Notify(268698)
		col := ic.all[i]
		if col == nil {
			__antithesis_instrumentation__.Notify(268701)
			ic.keyAndSuffix[i].Name = "invalid"
			continue
		} else {
			__antithesis_instrumentation__.Notify(268702)
		}
		__antithesis_instrumentation__.Notify(268699)
		colID := col.GetID()
		typ := col.GetType()
		if colID != 0 && func() bool {
			__antithesis_instrumentation__.Notify(268703)
			return colID == invertedColumnID == true
		}() == true {
			__antithesis_instrumentation__.Notify(268704)
			typ = idx.InvertedColumnKeyType()
		} else {
			__antithesis_instrumentation__.Notify(268705)
		}
		__antithesis_instrumentation__.Notify(268700)
		ic.keyAndSuffix[i] = descpb.IndexFetchSpec_KeyColumn{
			IndexFetchSpec_Column: descpb.IndexFetchSpec_Column{
				Name:          col.GetName(),
				ColumnID:      colID,
				Type:          typ,
				IsNonNullable: !col.IsNullable(),
			},
			Direction:   ic.allDirs[i],
			IsComposite: compositeIDs.Contains(colID),
			IsInverted:  colID == invertedColumnID,
		}
	}
	__antithesis_instrumentation__.Notify(268691)
	return ic
}

func appendColumnsByID(slice *[]catalog.Column, source []catalog.Column, ids []descpb.ColumnID) {
	__antithesis_instrumentation__.Notify(268706)
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(268707)
		var col catalog.Column
		for _, candidate := range source {
			__antithesis_instrumentation__.Notify(268709)
			if candidate.GetID() == id {
				__antithesis_instrumentation__.Notify(268710)
				col = candidate
				break
			} else {
				__antithesis_instrumentation__.Notify(268711)
			}
		}
		__antithesis_instrumentation__.Notify(268708)
		*slice = append(*slice, col)
	}
}

func lazyAllocAppendColumn(slice *[]catalog.Column, col catalog.Column, cap int) {
	__antithesis_instrumentation__.Notify(268712)
	if *slice == nil {
		__antithesis_instrumentation__.Notify(268714)
		*slice = make([]catalog.Column, 0, cap)
	} else {
		__antithesis_instrumentation__.Notify(268715)
	}
	__antithesis_instrumentation__.Notify(268713)
	*slice = append(*slice, col)
}
