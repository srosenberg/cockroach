package scmutationexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/errors"
)

func (m *visitor) RemoveSchemaParent(ctx context.Context, op scop.RemoveSchemaParent) error {
	__antithesis_instrumentation__.Notify(582112)
	if desc, err := m.s.GetDescriptor(ctx, op.Parent.ParentDatabaseID); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(582116)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(582117)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582118)
	}
	__antithesis_instrumentation__.Notify(582113)
	db, err := m.checkOutDatabase(ctx, op.Parent.ParentDatabaseID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582119)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582120)
	}
	__antithesis_instrumentation__.Notify(582114)
	for name, info := range db.Schemas {
		__antithesis_instrumentation__.Notify(582121)
		if info.ID == op.Parent.SchemaID {
			__antithesis_instrumentation__.Notify(582122)
			delete(db.Schemas, name)
		} else {
			__antithesis_instrumentation__.Notify(582123)
		}
	}
	__antithesis_instrumentation__.Notify(582115)
	return nil
}

func (m *visitor) RemoveOwnerBackReferenceInSequence(
	ctx context.Context, op scop.RemoveOwnerBackReferenceInSequence,
) error {
	__antithesis_instrumentation__.Notify(582124)
	if desc, err := m.s.GetDescriptor(ctx, op.SequenceID); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(582127)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(582128)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582129)
	}
	__antithesis_instrumentation__.Notify(582125)
	seq, err := m.checkOutTable(ctx, op.SequenceID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582130)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582131)
	}
	__antithesis_instrumentation__.Notify(582126)
	seq.GetSequenceOpts().SequenceOwner.Reset()
	return nil
}

func (m *visitor) RemoveSequenceOwner(ctx context.Context, op scop.RemoveSequenceOwner) error {
	__antithesis_instrumentation__.Notify(582132)
	if desc, err := m.s.GetDescriptor(ctx, op.TableID); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(582136)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(582137)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582138)
	}
	__antithesis_instrumentation__.Notify(582133)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582139)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582140)
	}
	__antithesis_instrumentation__.Notify(582134)
	col, err := tbl.FindColumnWithID(op.ColumnID)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(582141)
		return col == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(582142)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582143)
	}
	__antithesis_instrumentation__.Notify(582135)
	ids := catalog.MakeDescriptorIDSet(col.ColumnDesc().OwnsSequenceIds...)
	ids.Remove(op.OwnedSequenceID)
	col.ColumnDesc().OwnsSequenceIds = ids.Ordered()
	return nil
}

func (m *visitor) RemoveCheckConstraint(ctx context.Context, op scop.RemoveCheckConstraint) error {
	__antithesis_instrumentation__.Notify(582144)
	if desc, err := m.s.GetDescriptor(ctx, op.TableID); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(582150)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(582151)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582152)
	}
	__antithesis_instrumentation__.Notify(582145)
	tbl, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582153)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582154)
	}
	__antithesis_instrumentation__.Notify(582146)
	var found bool
	for i, c := range tbl.Checks {
		__antithesis_instrumentation__.Notify(582155)
		if c.ConstraintID == op.ConstraintID {
			__antithesis_instrumentation__.Notify(582156)
			tbl.Checks = append(tbl.Checks[:i], tbl.Checks[i+1:]...)
			found = true
			break
		} else {
			__antithesis_instrumentation__.Notify(582157)
		}
	}
	__antithesis_instrumentation__.Notify(582147)
	for i, m := range tbl.Mutations {
		__antithesis_instrumentation__.Notify(582158)
		if c := m.GetConstraint(); c != nil && func() bool {
			__antithesis_instrumentation__.Notify(582159)
			return c.ConstraintType != descpb.ConstraintToUpdate_CHECK == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(582160)
			return c.Check.ConstraintID == op.ConstraintID == true
		}() == true {
			__antithesis_instrumentation__.Notify(582161)
			tbl.Mutations = append(tbl.Mutations[:i], tbl.Mutations[i+1:]...)
			found = true
			break
		} else {
			__antithesis_instrumentation__.Notify(582162)
		}
	}
	__antithesis_instrumentation__.Notify(582148)
	if !found {
		__antithesis_instrumentation__.Notify(582163)
		return errors.AssertionFailedf("failed to find check constraint %d in table %q (%d)",
			op.ConstraintID, tbl.GetName(), tbl.GetID())
	} else {
		__antithesis_instrumentation__.Notify(582164)
	}
	__antithesis_instrumentation__.Notify(582149)
	return nil
}

func (m *visitor) RemoveForeignKeyBackReference(
	ctx context.Context, op scop.RemoveForeignKeyBackReference,
) error {
	__antithesis_instrumentation__.Notify(582165)
	if desc, err := m.s.GetDescriptor(ctx, op.ReferencedTableID); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(582172)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(582173)

		return err
	} else {
		__antithesis_instrumentation__.Notify(582174)
	}
	__antithesis_instrumentation__.Notify(582166)

	var name string
	{
		__antithesis_instrumentation__.Notify(582175)
		out, err := m.s.GetDescriptor(ctx, op.OriginTableID)
		if err != nil {
			__antithesis_instrumentation__.Notify(582179)
			return err
		} else {
			__antithesis_instrumentation__.Notify(582180)
		}
		__antithesis_instrumentation__.Notify(582176)
		tbl, err := catalog.AsTableDescriptor(out)
		if err != nil {
			__antithesis_instrumentation__.Notify(582181)
			return err
		} else {
			__antithesis_instrumentation__.Notify(582182)
		}
		__antithesis_instrumentation__.Notify(582177)
		for _, fk := range tbl.AllActiveAndInactiveForeignKeys() {
			__antithesis_instrumentation__.Notify(582183)
			if fk.ConstraintID == op.OriginConstraintID {
				__antithesis_instrumentation__.Notify(582184)
				name = fk.Name
				break
			} else {
				__antithesis_instrumentation__.Notify(582185)
			}
		}
		__antithesis_instrumentation__.Notify(582178)
		if name == "" {
			__antithesis_instrumentation__.Notify(582186)
			return errors.AssertionFailedf("foreign key with ID %d not found in origin table %q (%d)",
				op.OriginConstraintID, out.GetName(), out.GetID())
		} else {
			__antithesis_instrumentation__.Notify(582187)
		}
	}
	__antithesis_instrumentation__.Notify(582167)

	in, err := m.checkOutTable(ctx, op.ReferencedTableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582188)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582189)
	}
	__antithesis_instrumentation__.Notify(582168)
	var found bool
	for i, fk := range in.InboundFKs {
		__antithesis_instrumentation__.Notify(582190)
		if fk.OriginTableID == op.OriginTableID && func() bool {
			__antithesis_instrumentation__.Notify(582191)
			return fk.Name == name == true
		}() == true {
			__antithesis_instrumentation__.Notify(582192)
			in.InboundFKs = append(in.InboundFKs[:i], in.InboundFKs[i+1:]...)
			found = true
			break
		} else {
			__antithesis_instrumentation__.Notify(582193)
		}
	}
	__antithesis_instrumentation__.Notify(582169)
	for i, m := range in.Mutations {
		__antithesis_instrumentation__.Notify(582194)
		if c := m.GetConstraint(); c != nil && func() bool {
			__antithesis_instrumentation__.Notify(582195)
			return c.ConstraintType != descpb.ConstraintToUpdate_FOREIGN_KEY == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(582196)
			return c.ForeignKey.OriginTableID == op.OriginTableID == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(582197)
			return c.Name == name == true
		}() == true {
			__antithesis_instrumentation__.Notify(582198)
			in.Mutations = append(in.Mutations[:i], in.Mutations[i+1:]...)
			found = true
			break
		} else {
			__antithesis_instrumentation__.Notify(582199)
		}
	}
	__antithesis_instrumentation__.Notify(582170)
	if !found {
		__antithesis_instrumentation__.Notify(582200)
		return errors.AssertionFailedf("foreign key %q not found in referenced table %q (%d)",
			name, in.GetName(), in.GetID())
	} else {
		__antithesis_instrumentation__.Notify(582201)
	}
	__antithesis_instrumentation__.Notify(582171)
	return nil
}

func (m *visitor) RemoveForeignKeyConstraint(
	ctx context.Context, op scop.RemoveForeignKeyConstraint,
) error {
	__antithesis_instrumentation__.Notify(582202)
	if desc, err := m.s.GetDescriptor(ctx, op.TableID); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(582207)
		return desc.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(582208)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582209)
	}
	__antithesis_instrumentation__.Notify(582203)
	out, err := m.checkOutTable(ctx, op.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582210)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582211)
	}
	__antithesis_instrumentation__.Notify(582204)
	for i, fk := range out.OutboundFKs {
		__antithesis_instrumentation__.Notify(582212)
		if fk.ConstraintID == op.ConstraintID {
			__antithesis_instrumentation__.Notify(582213)
			out.OutboundFKs = append(out.OutboundFKs[:i], out.OutboundFKs[i+1:]...)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(582214)
		}
	}
	__antithesis_instrumentation__.Notify(582205)
	for i, m := range out.Mutations {
		__antithesis_instrumentation__.Notify(582215)
		if c := m.GetConstraint(); c != nil && func() bool {
			__antithesis_instrumentation__.Notify(582216)
			return c.ConstraintType != descpb.ConstraintToUpdate_FOREIGN_KEY == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(582217)
			return c.ForeignKey.ConstraintID == op.ConstraintID == true
		}() == true {
			__antithesis_instrumentation__.Notify(582218)
			out.Mutations = append(out.Mutations[:i], out.Mutations[i+1:]...)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(582219)
		}
	}
	__antithesis_instrumentation__.Notify(582206)
	return errors.AssertionFailedf("foreign key with ID %d not found in origin table %q (%d)",
		op.ConstraintID, out.GetName(), out.GetID())
}

func (m *visitor) UpdateTableBackReferencesInTypes(
	ctx context.Context, op scop.UpdateTableBackReferencesInTypes,
) error {
	__antithesis_instrumentation__.Notify(582220)
	var forwardRefs catalog.DescriptorIDSet
	if desc, err := m.s.GetDescriptor(ctx, op.BackReferencedTableID); err != nil {
		__antithesis_instrumentation__.Notify(582222)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582223)
		if !desc.Dropped() {
			__antithesis_instrumentation__.Notify(582224)
			tbl, err := catalog.AsTableDescriptor(desc)
			if err != nil {
				__antithesis_instrumentation__.Notify(582230)
				return err
			} else {
				__antithesis_instrumentation__.Notify(582231)
			}
			__antithesis_instrumentation__.Notify(582225)
			parent, err := m.s.GetDescriptor(ctx, desc.GetParentID())
			if err != nil {
				__antithesis_instrumentation__.Notify(582232)
				return err
			} else {
				__antithesis_instrumentation__.Notify(582233)
			}
			__antithesis_instrumentation__.Notify(582226)
			db, err := catalog.AsDatabaseDescriptor(parent)
			if err != nil {
				__antithesis_instrumentation__.Notify(582234)
				return err
			} else {
				__antithesis_instrumentation__.Notify(582235)
			}
			__antithesis_instrumentation__.Notify(582227)
			ids, _, err := tbl.GetAllReferencedTypeIDs(db, func(id descpb.ID) (catalog.TypeDescriptor, error) {
				__antithesis_instrumentation__.Notify(582236)
				d, err := m.s.GetDescriptor(ctx, id)
				if err != nil {
					__antithesis_instrumentation__.Notify(582238)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(582239)
				}
				__antithesis_instrumentation__.Notify(582237)
				return catalog.AsTypeDescriptor(d)
			})
			__antithesis_instrumentation__.Notify(582228)
			if err != nil {
				__antithesis_instrumentation__.Notify(582240)
				return err
			} else {
				__antithesis_instrumentation__.Notify(582241)
			}
			__antithesis_instrumentation__.Notify(582229)
			for _, id := range ids {
				__antithesis_instrumentation__.Notify(582242)
				forwardRefs.Add(id)
			}
		} else {
			__antithesis_instrumentation__.Notify(582243)
		}
	}
	__antithesis_instrumentation__.Notify(582221)
	return updateBackReferencesInTypes(ctx, m, op.TypeIDs, op.BackReferencedTableID, forwardRefs)
}

func (m *visitor) RemoveBackReferenceInTypes(
	ctx context.Context, op scop.RemoveBackReferenceInTypes,
) error {
	__antithesis_instrumentation__.Notify(582244)
	return updateBackReferencesInTypes(ctx, m, op.TypeIDs, op.BackReferencedDescID, catalog.DescriptorIDSet{})
}

func updateBackReferencesInTypes(
	ctx context.Context,
	m *visitor,
	typeIDs []catid.DescID,
	backReferencedDescID catid.DescID,
	forwardRefs catalog.DescriptorIDSet,
) error {
	__antithesis_instrumentation__.Notify(582245)
	for _, typeID := range typeIDs {
		__antithesis_instrumentation__.Notify(582247)
		if desc, err := m.s.GetDescriptor(ctx, typeID); err != nil {
			__antithesis_instrumentation__.Notify(582251)
			return err
		} else {
			__antithesis_instrumentation__.Notify(582252)
			if desc.Dropped() {
				__antithesis_instrumentation__.Notify(582253)

				continue
			} else {
				__antithesis_instrumentation__.Notify(582254)
			}
		}
		__antithesis_instrumentation__.Notify(582248)
		typ, err := m.checkOutType(ctx, typeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(582255)
			return err
		} else {
			__antithesis_instrumentation__.Notify(582256)
		}
		__antithesis_instrumentation__.Notify(582249)
		backRefs := catalog.MakeDescriptorIDSet(typ.ReferencingDescriptorIDs...)
		if forwardRefs.Contains(typeID) {
			__antithesis_instrumentation__.Notify(582257)
			if backRefs.Contains(backReferencedDescID) {
				__antithesis_instrumentation__.Notify(582259)
				continue
			} else {
				__antithesis_instrumentation__.Notify(582260)
			}
			__antithesis_instrumentation__.Notify(582258)
			backRefs.Add(backReferencedDescID)
		} else {
			__antithesis_instrumentation__.Notify(582261)
			if !backRefs.Contains(backReferencedDescID) {
				__antithesis_instrumentation__.Notify(582263)
				continue
			} else {
				__antithesis_instrumentation__.Notify(582264)
			}
			__antithesis_instrumentation__.Notify(582262)
			backRefs.Remove(backReferencedDescID)
		}
		__antithesis_instrumentation__.Notify(582250)
		typ.ReferencingDescriptorIDs = backRefs.Ordered()
	}
	__antithesis_instrumentation__.Notify(582246)
	return nil
}

func (m *visitor) UpdateBackReferencesInSequences(
	ctx context.Context, op scop.UpdateBackReferencesInSequences,
) error {
	__antithesis_instrumentation__.Notify(582265)
	var forwardRefs catalog.DescriptorIDSet
	if desc, err := m.s.GetDescriptor(ctx, op.BackReferencedTableID); err != nil {
		__antithesis_instrumentation__.Notify(582268)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582269)
		if !desc.Dropped() {
			__antithesis_instrumentation__.Notify(582270)
			tbl, err := catalog.AsTableDescriptor(desc)
			if err != nil {
				__antithesis_instrumentation__.Notify(582272)
				return err
			} else {
				__antithesis_instrumentation__.Notify(582273)
			}
			__antithesis_instrumentation__.Notify(582271)
			if op.BackReferencedColumnID != 0 {
				__antithesis_instrumentation__.Notify(582274)
				col, err := tbl.FindColumnWithID(op.BackReferencedColumnID)
				if err != nil {
					__antithesis_instrumentation__.Notify(582277)
					return err
				} else {
					__antithesis_instrumentation__.Notify(582278)
				}
				__antithesis_instrumentation__.Notify(582275)
				for i, n := 0, col.NumUsesSequences(); i < n; i++ {
					__antithesis_instrumentation__.Notify(582279)
					forwardRefs.Add(col.GetUsesSequenceID(i))
				}
				__antithesis_instrumentation__.Notify(582276)
				for i, n := 0, col.NumOwnsSequences(); i < n; i++ {
					__antithesis_instrumentation__.Notify(582280)
					forwardRefs.Add(col.GetOwnsSequenceID(i))
				}
			} else {
				__antithesis_instrumentation__.Notify(582281)
				for _, c := range tbl.AllActiveAndInactiveChecks() {
					__antithesis_instrumentation__.Notify(582282)
					ids, err := sequenceIDsInExpr(c.Expr)
					if err != nil {
						__antithesis_instrumentation__.Notify(582284)
						return err
					} else {
						__antithesis_instrumentation__.Notify(582285)
					}
					__antithesis_instrumentation__.Notify(582283)
					ids.ForEach(forwardRefs.Add)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(582286)
		}
	}
	__antithesis_instrumentation__.Notify(582266)
	for _, seqID := range op.SequenceIDs {
		__antithesis_instrumentation__.Notify(582287)
		if desc, err := m.s.GetDescriptor(ctx, seqID); err != nil {
			__antithesis_instrumentation__.Notify(582289)
			return err
		} else {
			__antithesis_instrumentation__.Notify(582290)
			if desc.Dropped() {
				__antithesis_instrumentation__.Notify(582291)

				continue
			} else {
				__antithesis_instrumentation__.Notify(582292)
			}
		}
		__antithesis_instrumentation__.Notify(582288)
		if err := updateBackReferencesInSequences(
			ctx, m, seqID, op.BackReferencedTableID, op.BackReferencedColumnID, forwardRefs,
		); err != nil {
			__antithesis_instrumentation__.Notify(582293)
			return err
		} else {
			__antithesis_instrumentation__.Notify(582294)
		}
	}
	__antithesis_instrumentation__.Notify(582267)
	return nil
}

func updateBackReferencesInSequences(
	ctx context.Context,
	m *visitor,
	seqID, tblID descpb.ID,
	colID descpb.ColumnID,
	forwardRefs catalog.DescriptorIDSet,
) error {
	__antithesis_instrumentation__.Notify(582295)
	seq, err := m.checkOutTable(ctx, seqID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582299)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582300)
	}
	__antithesis_instrumentation__.Notify(582296)
	var current, updated catalog.TableColSet
	_ = seq.ForeachDependedOnBy(func(dep *descpb.TableDescriptor_Reference) error {
		__antithesis_instrumentation__.Notify(582301)
		if dep.ID == tblID {
			__antithesis_instrumentation__.Notify(582303)
			current = catalog.MakeTableColSet(dep.ColumnIDs...)
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(582304)
		}
		__antithesis_instrumentation__.Notify(582302)
		return nil
	})
	__antithesis_instrumentation__.Notify(582297)
	if forwardRefs.Contains(seqID) {
		__antithesis_instrumentation__.Notify(582305)
		if current.Contains(colID) {
			__antithesis_instrumentation__.Notify(582307)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(582308)
		}
		__antithesis_instrumentation__.Notify(582306)
		updated.UnionWith(current)
		updated.Add(colID)
	} else {
		__antithesis_instrumentation__.Notify(582309)
		if !current.Contains(colID) {
			__antithesis_instrumentation__.Notify(582311)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(582312)
		}
		__antithesis_instrumentation__.Notify(582310)
		current.ForEach(func(id descpb.ColumnID) {
			__antithesis_instrumentation__.Notify(582313)
			if id != colID {
				__antithesis_instrumentation__.Notify(582314)
				updated.Add(id)
			} else {
				__antithesis_instrumentation__.Notify(582315)
			}
		})
	}
	__antithesis_instrumentation__.Notify(582298)
	seq.UpdateColumnsDependedOnBy(tblID, updated)
	return nil
}

func (m *visitor) RemoveViewBackReferencesInRelations(
	ctx context.Context, op scop.RemoveViewBackReferencesInRelations,
) error {
	__antithesis_instrumentation__.Notify(582316)
	for _, relationID := range op.RelationIDs {
		__antithesis_instrumentation__.Notify(582318)
		if desc, err := m.s.GetDescriptor(ctx, relationID); err != nil {
			__antithesis_instrumentation__.Notify(582320)
			return err
		} else {
			__antithesis_instrumentation__.Notify(582321)
			if desc.Dropped() {
				__antithesis_instrumentation__.Notify(582322)

				continue
			} else {
				__antithesis_instrumentation__.Notify(582323)
			}
		}
		__antithesis_instrumentation__.Notify(582319)
		if err := removeViewBackReferencesInRelation(ctx, m, relationID, op.BackReferencedViewID); err != nil {
			__antithesis_instrumentation__.Notify(582324)
			return err
		} else {
			__antithesis_instrumentation__.Notify(582325)
		}
	}
	__antithesis_instrumentation__.Notify(582317)
	return nil
}

func removeViewBackReferencesInRelation(
	ctx context.Context, m *visitor, relationID, viewID descpb.ID,
) error {
	__antithesis_instrumentation__.Notify(582326)
	tbl, err := m.checkOutTable(ctx, relationID)
	if err != nil {
		__antithesis_instrumentation__.Notify(582329)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582330)
	}
	__antithesis_instrumentation__.Notify(582327)
	var newBackRefs []descpb.TableDescriptor_Reference
	for _, by := range tbl.DependedOnBy {
		__antithesis_instrumentation__.Notify(582331)
		if by.ID != viewID {
			__antithesis_instrumentation__.Notify(582332)
			newBackRefs = append(newBackRefs, by)
		} else {
			__antithesis_instrumentation__.Notify(582333)
		}
	}
	__antithesis_instrumentation__.Notify(582328)
	tbl.DependedOnBy = newBackRefs
	return nil
}
