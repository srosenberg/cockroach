package tabledesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/redact"
)

func (desc *immutable) SafeMessage() string {
	__antithesis_instrumentation__.Notify(269010)
	return formatSafeTableDesc("tabledesc.immutable", desc)
}

func (desc *Mutable) SafeMessage() string {
	__antithesis_instrumentation__.Notify(269011)
	return formatSafeTableDesc("tabledesc.Mutable", desc)
}

func formatSafeTableDesc(typeName string, desc catalog.TableDescriptor) string {
	__antithesis_instrumentation__.Notify(269012)
	var buf redact.StringBuilder
	buf.Printf(typeName + ": {")
	formatSafeTableProperties(&buf, desc)
	buf.Printf("}")
	return buf.String()
}
func formatSafeTableProperties(w *redact.StringBuilder, desc catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(269013)
	catalog.FormatSafeDescriptorProperties(w, desc)
	if desc.IsTemporary() {
		__antithesis_instrumentation__.Notify(269018)
		w.Printf(", Temporary: true")
	} else {
		__antithesis_instrumentation__.Notify(269019)
	}
	__antithesis_instrumentation__.Notify(269014)
	if desc.IsView() {
		__antithesis_instrumentation__.Notify(269020)
		w.Printf(", View: true")
	} else {
		__antithesis_instrumentation__.Notify(269021)
	}
	__antithesis_instrumentation__.Notify(269015)
	if desc.IsSequence() {
		__antithesis_instrumentation__.Notify(269022)
		w.Printf(", Sequence: true")
	} else {
		__antithesis_instrumentation__.Notify(269023)
	}
	__antithesis_instrumentation__.Notify(269016)
	if desc.IsVirtualTable() {
		__antithesis_instrumentation__.Notify(269024)
		w.Printf(", Virtual: true")
	} else {
		__antithesis_instrumentation__.Notify(269025)
	}
	__antithesis_instrumentation__.Notify(269017)
	formatSafeTableColumns(w, desc)
	formatSafeTableColumnFamilies(w, desc)
	formatSafeTableMutationJobs(w, desc)
	formatSafeMutations(w, desc)
	formatSafeTableIndexes(w, desc)
	formatSafeTableConstraints(w, desc)

}

func formatSafeTableColumns(w *redact.StringBuilder, desc catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(269026)
	var printed bool
	formatColumn := func(c *descpb.ColumnDescriptor, m *descpb.DescriptorMutation) {
		__antithesis_instrumentation__.Notify(269029)
		if printed {
			__antithesis_instrumentation__.Notify(269031)
			w.Printf(", ")
		} else {
			__antithesis_instrumentation__.Notify(269032)
		}
		__antithesis_instrumentation__.Notify(269030)
		formatSafeColumn(w, c, m)
		printed = true
	}
	__antithesis_instrumentation__.Notify(269027)
	td := desc.TableDesc()
	w.Printf(", NextColumnID: %d", td.NextColumnID)
	w.Printf(", Columns: [")
	for i := range td.Columns {
		__antithesis_instrumentation__.Notify(269033)
		formatColumn(&td.Columns[i], nil)
	}
	__antithesis_instrumentation__.Notify(269028)
	w.Printf("]")
}

func formatSafeColumn(
	w *redact.StringBuilder, c *descpb.ColumnDescriptor, m *descpb.DescriptorMutation,
) {
	__antithesis_instrumentation__.Notify(269034)
	w.Printf("{ID: %d, TypeID: %d", c.ID, c.Type.InternalType.Oid)
	w.Printf(", Null: %t", c.Nullable)
	if c.Hidden {
		__antithesis_instrumentation__.Notify(269042)
		w.Printf(", Hidden: true")
	} else {
		__antithesis_instrumentation__.Notify(269043)
	}
	__antithesis_instrumentation__.Notify(269035)
	if c.HasDefault() {
		__antithesis_instrumentation__.Notify(269044)
		w.Printf(", HasDefault: true")
	} else {
		__antithesis_instrumentation__.Notify(269045)
	}
	__antithesis_instrumentation__.Notify(269036)
	if c.IsComputed() {
		__antithesis_instrumentation__.Notify(269046)
		w.Printf(", IsComputed: true")
	} else {
		__antithesis_instrumentation__.Notify(269047)
	}
	__antithesis_instrumentation__.Notify(269037)
	if c.AlterColumnTypeInProgress {
		__antithesis_instrumentation__.Notify(269048)
		w.Printf(", AlterColumnTypeInProgress: t")
	} else {
		__antithesis_instrumentation__.Notify(269049)
	}
	__antithesis_instrumentation__.Notify(269038)
	if len(c.OwnsSequenceIds) > 0 {
		__antithesis_instrumentation__.Notify(269050)
		w.Printf(", OwnsSequenceIDs: ")
		formatSafeIDs(w, c.OwnsSequenceIds)
	} else {
		__antithesis_instrumentation__.Notify(269051)
	}
	__antithesis_instrumentation__.Notify(269039)
	if len(c.UsesSequenceIds) > 0 {
		__antithesis_instrumentation__.Notify(269052)
		w.Printf(", UsesSequenceIDs: ")
		formatSafeIDs(w, c.UsesSequenceIds)
	} else {
		__antithesis_instrumentation__.Notify(269053)
	}
	__antithesis_instrumentation__.Notify(269040)
	if m != nil {
		__antithesis_instrumentation__.Notify(269054)
		w.Printf(", State: %s, MutationID: %d", m.Direction, m.MutationID)
	} else {
		__antithesis_instrumentation__.Notify(269055)
	}
	__antithesis_instrumentation__.Notify(269041)
	w.Printf("}")
}

func formatSafeTableIndexes(w *redact.StringBuilder, desc catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(269056)
	w.Printf(", PrimaryIndex: %d", desc.GetPrimaryIndexID())
	w.Printf(", NextIndexID: %d", desc.GetNextIndexID())
	w.Printf(", Indexes: [")
	_ = catalog.ForEachActiveIndex(desc, func(idx catalog.Index) error {
		__antithesis_instrumentation__.Notify(269058)
		if !idx.Primary() {
			__antithesis_instrumentation__.Notify(269060)
			w.Printf(", ")
		} else {
			__antithesis_instrumentation__.Notify(269061)
		}
		__antithesis_instrumentation__.Notify(269059)
		formatSafeIndex(w, idx.IndexDesc(), nil)
		return nil
	})
	__antithesis_instrumentation__.Notify(269057)
	w.Printf("]")
}

func formatSafeIndex(
	w *redact.StringBuilder, idx *descpb.IndexDescriptor, mut *descpb.DescriptorMutation,
) {
	__antithesis_instrumentation__.Notify(269062)
	w.Printf("{")
	w.Printf("ID: %d", idx.ID)
	w.Printf(", Unique: %t", idx.Unique)
	if idx.Predicate != "" {
		__antithesis_instrumentation__.Notify(269068)
		w.Printf(", Partial: true")
	} else {
		__antithesis_instrumentation__.Notify(269069)
	}
	__antithesis_instrumentation__.Notify(269063)
	w.Printf(", KeyColumns: [")
	for i := range idx.KeyColumnIDs {
		__antithesis_instrumentation__.Notify(269070)
		if i > 0 {
			__antithesis_instrumentation__.Notify(269072)
			w.Printf(", ")
		} else {
			__antithesis_instrumentation__.Notify(269073)
		}
		__antithesis_instrumentation__.Notify(269071)
		w.Printf("{ID: %d, Dir: %s}", idx.KeyColumnIDs[i], idx.KeyColumnDirections[i])
	}
	__antithesis_instrumentation__.Notify(269064)
	w.Printf("]")
	if len(idx.KeySuffixColumnIDs) > 0 {
		__antithesis_instrumentation__.Notify(269074)
		w.Printf(", KeySuffixColumns: ")
		formatSafeColumnIDs(w, idx.KeySuffixColumnIDs)
	} else {
		__antithesis_instrumentation__.Notify(269075)
	}
	__antithesis_instrumentation__.Notify(269065)
	if len(idx.StoreColumnIDs) > 0 {
		__antithesis_instrumentation__.Notify(269076)
		w.Printf(", StoreColumns: ")
		formatSafeColumnIDs(w, idx.StoreColumnIDs)
	} else {
		__antithesis_instrumentation__.Notify(269077)
	}
	__antithesis_instrumentation__.Notify(269066)
	if mut != nil {
		__antithesis_instrumentation__.Notify(269078)
		w.Printf(", State: %s, MutationID: %d", mut.Direction, mut.MutationID)
	} else {
		__antithesis_instrumentation__.Notify(269079)
	}
	__antithesis_instrumentation__.Notify(269067)
	w.Printf("}")
}

func formatSafeTableConstraints(w *redact.StringBuilder, desc catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(269080)
	td := desc.TableDesc()
	formatSafeTableChecks(w, td.Checks)
	formatSafeTableUniqueWithoutIndexConstraints(w, td.UniqueWithoutIndexConstraints)
	formatSafeTableFKs(w, "InboundFKs", td.InboundFKs)
	formatSafeTableFKs(w, "OutboundFKs", td.OutboundFKs)
}

func formatSafeTableFKs(
	w *redact.StringBuilder, fksType string, fks []descpb.ForeignKeyConstraint,
) {
	__antithesis_instrumentation__.Notify(269081)
	for i := range fks {
		__antithesis_instrumentation__.Notify(269083)
		w.Printf(", ")
		if i == 0 {
			__antithesis_instrumentation__.Notify(269085)
			w.Printf(fksType)
			w.Printf(": [")
		} else {
			__antithesis_instrumentation__.Notify(269086)
		}
		__antithesis_instrumentation__.Notify(269084)
		fk := &fks[i]
		formatSafeFK(w, fk, nil)
	}
	__antithesis_instrumentation__.Notify(269082)
	if len(fks) > 0 {
		__antithesis_instrumentation__.Notify(269087)
		w.Printf("]")
	} else {
		__antithesis_instrumentation__.Notify(269088)
	}
}

func formatSafeFK(
	w *redact.StringBuilder, fk *descpb.ForeignKeyConstraint, m *descpb.DescriptorMutation,
) {
	__antithesis_instrumentation__.Notify(269089)
	w.Printf("{OriginTableID: %d", fk.OriginTableID)
	w.Printf(", OriginColumns: ")
	formatSafeColumnIDs(w, fk.OriginColumnIDs)
	w.Printf(", ReferencedTableID: %d", fk.ReferencedTableID)
	w.Printf(", ReferencedColumnIDs: ")
	formatSafeColumnIDs(w, fk.ReferencedColumnIDs)
	w.Printf(", Validity: %s", fk.Validity.String())
	if m != nil {
		__antithesis_instrumentation__.Notify(269091)
		w.Printf(", State: %s, MutationID: %d", m.Direction, m.MutationID)
	} else {
		__antithesis_instrumentation__.Notify(269092)
	}
	__antithesis_instrumentation__.Notify(269090)
	w.Printf("}")
}

func formatSafeTableChecks(
	w *redact.StringBuilder, checks []*descpb.TableDescriptor_CheckConstraint,
) {
	__antithesis_instrumentation__.Notify(269093)
	for i, c := range checks {
		__antithesis_instrumentation__.Notify(269095)
		if i == 0 {
			__antithesis_instrumentation__.Notify(269097)
			w.Printf(", Checks: [")
		} else {
			__antithesis_instrumentation__.Notify(269098)
			w.Printf(", ")
		}
		__antithesis_instrumentation__.Notify(269096)
		formatSafeCheck(w, c, nil)
	}
	__antithesis_instrumentation__.Notify(269094)
	if len(checks) > 0 {
		__antithesis_instrumentation__.Notify(269099)
		w.Printf("]")
	} else {
		__antithesis_instrumentation__.Notify(269100)
	}
}

func formatSafeTableUniqueWithoutIndexConstraints(
	w *redact.StringBuilder, constraints []descpb.UniqueWithoutIndexConstraint,
) {
	__antithesis_instrumentation__.Notify(269101)
	for i := range constraints {
		__antithesis_instrumentation__.Notify(269103)
		c := &constraints[i]
		if i == 0 {
			__antithesis_instrumentation__.Notify(269105)
			w.Printf(", Unique Without Index Constraints: [")
		} else {
			__antithesis_instrumentation__.Notify(269106)
			w.Printf(", ")
		}
		__antithesis_instrumentation__.Notify(269104)
		formatSafeUniqueWithoutIndexConstraint(w, c, nil)
	}
	__antithesis_instrumentation__.Notify(269102)
	if len(constraints) > 0 {
		__antithesis_instrumentation__.Notify(269107)
		w.Printf("]")
	} else {
		__antithesis_instrumentation__.Notify(269108)
	}
}

func formatSafeTableColumnFamilies(w *redact.StringBuilder, desc catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(269109)
	td := desc.TableDesc()
	w.Printf(", NextFamilyID: %d", td.NextFamilyID)
	for i := range td.Families {
		__antithesis_instrumentation__.Notify(269111)
		w.Printf(", ")
		if i == 0 {
			__antithesis_instrumentation__.Notify(269113)
			w.Printf("Families: [")
		} else {
			__antithesis_instrumentation__.Notify(269114)
		}
		__antithesis_instrumentation__.Notify(269112)
		formatSafeTableColumnFamily(w, &td.Families[i])
	}
	__antithesis_instrumentation__.Notify(269110)
	if len(td.Families) > 0 {
		__antithesis_instrumentation__.Notify(269115)
		w.Printf("]")
	} else {
		__antithesis_instrumentation__.Notify(269116)
	}
}

func formatSafeTableColumnFamily(w *redact.StringBuilder, f *descpb.ColumnFamilyDescriptor) {
	__antithesis_instrumentation__.Notify(269117)
	w.Printf("{")
	w.Printf("ID: %d", f.ID)
	w.Printf(", Columns: ")
	formatSafeColumnIDs(w, f.ColumnIDs)
	w.Printf("}")
}

func formatSafeTableMutationJobs(w *redact.StringBuilder, td catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(269118)
	mutationJobs := td.GetMutationJobs()
	for i := range mutationJobs {
		__antithesis_instrumentation__.Notify(269120)
		w.Printf(", ")
		if i == 0 {
			__antithesis_instrumentation__.Notify(269122)
			w.Printf("MutationJobs: [")
		} else {
			__antithesis_instrumentation__.Notify(269123)
		}
		__antithesis_instrumentation__.Notify(269121)
		m := &mutationJobs[i]
		w.Printf("{MutationID: %d, JobID: %d}", m.MutationID, m.JobID)
	}
	__antithesis_instrumentation__.Notify(269119)
	if len(mutationJobs) > 0 {
		__antithesis_instrumentation__.Notify(269124)
		w.Printf("]")
	} else {
		__antithesis_instrumentation__.Notify(269125)
	}
}

func formatSafeMutations(w *redact.StringBuilder, td catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(269126)
	mutations := td.TableDesc().Mutations
	for i := range mutations {
		__antithesis_instrumentation__.Notify(269128)
		w.Printf(", ")
		m := &mutations[i]
		if i == 0 {
			__antithesis_instrumentation__.Notify(269130)
			w.Printf("Mutations: [")
		} else {
			__antithesis_instrumentation__.Notify(269131)
		}
		__antithesis_instrumentation__.Notify(269129)
		formatSafeMutation(w, m)
	}
	__antithesis_instrumentation__.Notify(269127)
	if len(mutations) > 0 {
		__antithesis_instrumentation__.Notify(269132)
		w.Printf("]")
	} else {
		__antithesis_instrumentation__.Notify(269133)
	}
}

func formatSafeMutation(w *redact.StringBuilder, m *descpb.DescriptorMutation) {
	__antithesis_instrumentation__.Notify(269134)
	w.Printf("{MutationID: %d", m.MutationID)
	w.Printf(", Direction: %s", m.Direction)
	w.Printf(", State: %s", m.State)
	switch md := m.Descriptor_.(type) {
	case *descpb.DescriptorMutation_Constraint:
		__antithesis_instrumentation__.Notify(269136)
		w.Printf(", ConstraintType: %s", md.Constraint.ConstraintType)
		if md.Constraint.NotNullColumn != 0 {
			__antithesis_instrumentation__.Notify(269144)
			w.Printf(", NotNullColumn: %d", md.Constraint.NotNullColumn)
		} else {
			__antithesis_instrumentation__.Notify(269145)
		}
		__antithesis_instrumentation__.Notify(269137)
		switch {
		case !md.Constraint.ForeignKey.Equal(&descpb.ForeignKeyConstraint{}):
			__antithesis_instrumentation__.Notify(269146)
			w.Printf(", ForeignKey: ")
			formatSafeFK(w, &md.Constraint.ForeignKey, m)
		case !md.Constraint.Check.Equal(&descpb.TableDescriptor_CheckConstraint{}):
			__antithesis_instrumentation__.Notify(269147)
			w.Printf(", Check: ")
			formatSafeCheck(w, &md.Constraint.Check, m)
		default:
			__antithesis_instrumentation__.Notify(269148)
		}
	case *descpb.DescriptorMutation_Index:
		__antithesis_instrumentation__.Notify(269138)
		w.Printf(", Index: ")
		formatSafeIndex(w, md.Index, m)
	case *descpb.DescriptorMutation_Column:
		__antithesis_instrumentation__.Notify(269139)
		w.Printf(", Column: ")
		formatSafeColumn(w, md.Column, m)
	case *descpb.DescriptorMutation_PrimaryKeySwap:
		__antithesis_instrumentation__.Notify(269140)
		w.Printf(", PrimaryKeySwap: {")
		w.Printf("OldPrimaryIndexID: %d", md.PrimaryKeySwap.OldPrimaryIndexId)
		w.Printf(", OldIndexes: ")
		formatSafeIndexIDs(w, md.PrimaryKeySwap.NewIndexes)
		w.Printf(", NewPrimaryIndexID: %d", md.PrimaryKeySwap.NewPrimaryIndexId)
		w.Printf(", NewIndexes: ")
		formatSafeIndexIDs(w, md.PrimaryKeySwap.NewIndexes)
		w.Printf("}")
	case *descpb.DescriptorMutation_ComputedColumnSwap:
		__antithesis_instrumentation__.Notify(269141)
		w.Printf(", ComputedColumnSwap: {OldColumnID: %d, NewColumnID: %d}",
			md.ComputedColumnSwap.OldColumnId, md.ComputedColumnSwap.NewColumnId)
	case *descpb.DescriptorMutation_MaterializedViewRefresh:
		__antithesis_instrumentation__.Notify(269142)
		w.Printf(", MaterializedViewRefresh: {")
		w.Printf("NewPrimaryIndex: ")
		formatSafeIndex(w, &md.MaterializedViewRefresh.NewPrimaryIndex, m)
		w.Printf(", NewIndexes: [")
		for i := range md.MaterializedViewRefresh.NewIndexes {
			__antithesis_instrumentation__.Notify(269149)
			if i > 0 {
				__antithesis_instrumentation__.Notify(269151)
				w.Printf(", ")
			} else {
				__antithesis_instrumentation__.Notify(269152)
			}
			__antithesis_instrumentation__.Notify(269150)
			formatSafeIndex(w, &md.MaterializedViewRefresh.NewIndexes[i], m)
		}
		__antithesis_instrumentation__.Notify(269143)
		w.Printf("]")
		w.Printf(", AsOf: %s, ShouldBackfill: %b",
			md.MaterializedViewRefresh.AsOf, md.MaterializedViewRefresh.ShouldBackfill)
		w.Printf("}")
	}
	__antithesis_instrumentation__.Notify(269135)
	w.Printf("}")
}

func formatSafeCheck(
	w *redact.StringBuilder, c *descpb.TableDescriptor_CheckConstraint, m *descpb.DescriptorMutation,
) {
	__antithesis_instrumentation__.Notify(269153)

	w.Printf("{Columns: ")
	formatSafeColumnIDs(w, c.ColumnIDs)
	w.Printf(", Validity: %s", c.Validity.String())
	if c.Hidden {
		__antithesis_instrumentation__.Notify(269156)
		w.Printf(", Hidden: true")
	} else {
		__antithesis_instrumentation__.Notify(269157)
	}
	__antithesis_instrumentation__.Notify(269154)
	if m != nil {
		__antithesis_instrumentation__.Notify(269158)
		w.Printf(", State: %s, MutationID: %d", m.Direction, m.MutationID)
	} else {
		__antithesis_instrumentation__.Notify(269159)
	}
	__antithesis_instrumentation__.Notify(269155)
	w.Printf("}")
}

func formatSafeUniqueWithoutIndexConstraint(
	w *redact.StringBuilder, c *descpb.UniqueWithoutIndexConstraint, m *descpb.DescriptorMutation,
) {
	__antithesis_instrumentation__.Notify(269160)

	w.Printf("{TableID: %d", c.TableID)
	w.Printf(", Columns: ")
	formatSafeColumnIDs(w, c.ColumnIDs)
	w.Printf(", Validity: %s", c.Validity.String())
	if m != nil {
		__antithesis_instrumentation__.Notify(269162)
		w.Printf(", State: %s, MutationID: %d", m.Direction, m.MutationID)
	} else {
		__antithesis_instrumentation__.Notify(269163)
	}
	__antithesis_instrumentation__.Notify(269161)
	w.Printf("}")
}

func formatSafeColumnIDs(w *redact.StringBuilder, colIDs []descpb.ColumnID) {
	__antithesis_instrumentation__.Notify(269164)
	w.Printf("[")
	for i, colID := range colIDs {
		__antithesis_instrumentation__.Notify(269166)
		if i > 0 {
			__antithesis_instrumentation__.Notify(269168)
			w.Printf(", ")
		} else {
			__antithesis_instrumentation__.Notify(269169)
		}
		__antithesis_instrumentation__.Notify(269167)
		w.Printf("%d", colID)
	}
	__antithesis_instrumentation__.Notify(269165)
	w.Printf("]")
}

func formatSafeIndexIDs(w *redact.StringBuilder, indexIDs []descpb.IndexID) {
	__antithesis_instrumentation__.Notify(269170)
	w.Printf("[")
	for i, idxID := range indexIDs {
		__antithesis_instrumentation__.Notify(269172)
		if i > 0 {
			__antithesis_instrumentation__.Notify(269174)
			w.Printf(", ")
		} else {
			__antithesis_instrumentation__.Notify(269175)
		}
		__antithesis_instrumentation__.Notify(269173)
		w.Printf("%d", idxID)
	}
	__antithesis_instrumentation__.Notify(269171)
	w.Printf("]")
}

func formatSafeIDs(w *redact.StringBuilder, ids []descpb.ID) {
	__antithesis_instrumentation__.Notify(269176)
	w.Printf("[")
	for i, id := range ids {
		__antithesis_instrumentation__.Notify(269178)
		if i > 0 {
			__antithesis_instrumentation__.Notify(269180)
			w.Printf(", ")
		} else {
			__antithesis_instrumentation__.Notify(269181)
		}
		__antithesis_instrumentation__.Notify(269179)
		w.Printf("%d", id)
	}
	__antithesis_instrumentation__.Notify(269177)
	w.Printf("]")
}
