package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
)

type dropCascadeState struct {
	schemasToDelete []schemaWithDbDesc

	objectNamesToDelete []tree.ObjectName

	td                      []toDelete
	toDeleteByID            map[descpb.ID]*toDelete
	allTableObjectsToDelete []*tabledesc.Mutable
	typesToDelete           []*typedesc.Mutable

	droppedNames []string
}

type schemaWithDbDesc struct {
	schema catalog.SchemaDescriptor
	dbDesc *dbdesc.Mutable
}

func newDropCascadeState() *dropCascadeState {
	__antithesis_instrumentation__.Notify(468695)
	return &dropCascadeState{

		droppedNames: []string{},
	}
}

func (d *dropCascadeState) collectObjectsInSchema(
	ctx context.Context, p *planner, db *dbdesc.Mutable, schema catalog.SchemaDescriptor,
) error {
	__antithesis_instrumentation__.Notify(468696)
	names, _, err := resolver.GetObjectNamesAndIDs(
		ctx, p.txn, p, p.ExecCfg().Codec, db, schema.GetName(), true,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(468699)
		return err
	} else {
		__antithesis_instrumentation__.Notify(468700)
	}
	__antithesis_instrumentation__.Notify(468697)
	for i := range names {
		__antithesis_instrumentation__.Notify(468701)
		d.objectNamesToDelete = append(d.objectNamesToDelete, &names[i])
	}
	__antithesis_instrumentation__.Notify(468698)
	d.schemasToDelete = append(d.schemasToDelete, schemaWithDbDesc{schema: schema, dbDesc: db})
	return nil
}

func (d *dropCascadeState) resolveCollectedObjects(
	ctx context.Context, p *planner, db *dbdesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(468702)
	d.td = make([]toDelete, 0, len(d.objectNamesToDelete))

	for i := range d.objectNamesToDelete {
		__antithesis_instrumentation__.Notify(468706)
		objName := d.objectNamesToDelete[i]

		found, _, desc, err := p.LookupObject(
			ctx,
			tree.ObjectLookupFlags{

				CommonLookupFlags: tree.CommonLookupFlags{
					Required:       false,
					RequireMutable: true,
					IncludeOffline: true,
				},
				DesiredObjectKind: tree.TableObject,
			},
			objName.Catalog(),
			objName.Schema(),
			objName.Object(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(468708)
			return err
		} else {
			__antithesis_instrumentation__.Notify(468709)
		}
		__antithesis_instrumentation__.Notify(468707)
		if found {
			__antithesis_instrumentation__.Notify(468710)
			tbDesc, ok := desc.(*tabledesc.Mutable)
			if !ok {
				__antithesis_instrumentation__.Notify(468716)
				return errors.AssertionFailedf(
					"descriptor for %q is not Mutable",
					objName.Object(),
				)
			} else {
				__antithesis_instrumentation__.Notify(468717)
			}
			__antithesis_instrumentation__.Notify(468711)
			if db != nil {
				__antithesis_instrumentation__.Notify(468718)
				if tbDesc.State == descpb.DescriptorState_OFFLINE {
					__antithesis_instrumentation__.Notify(468719)
					dbName := db.GetName()
					return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
						"cannot drop a database with OFFLINE tables, ensure %s is"+
							" dropped or made public before dropping database %s",
						objName.FQString(), tree.AsString((*tree.Name)(&dbName)))
				} else {
					__antithesis_instrumentation__.Notify(468720)
				}
			} else {
				__antithesis_instrumentation__.Notify(468721)
			}
			__antithesis_instrumentation__.Notify(468712)
			checkOwnership := true

			if tbDesc.Temporary && func() bool {
				__antithesis_instrumentation__.Notify(468722)
				return !p.SessionData().IsTemporarySchemaID(uint32(tbDesc.GetParentSchemaID())) == true
			}() == true {
				__antithesis_instrumentation__.Notify(468723)
				checkOwnership = false
			} else {
				__antithesis_instrumentation__.Notify(468724)
			}
			__antithesis_instrumentation__.Notify(468713)
			if err := p.canDropTable(ctx, tbDesc, checkOwnership); err != nil {
				__antithesis_instrumentation__.Notify(468725)
				return err
			} else {
				__antithesis_instrumentation__.Notify(468726)
			}
			__antithesis_instrumentation__.Notify(468714)

			for _, ref := range tbDesc.DependedOnBy {
				__antithesis_instrumentation__.Notify(468727)
				if err := p.canRemoveDependentView(ctx, tbDesc, ref, tree.DropCascade); err != nil {
					__antithesis_instrumentation__.Notify(468728)
					return err
				} else {
					__antithesis_instrumentation__.Notify(468729)
				}
			}
			__antithesis_instrumentation__.Notify(468715)
			d.td = append(d.td, toDelete{objName, tbDesc})
		} else {
			__antithesis_instrumentation__.Notify(468730)

			found, _, desc, err := p.LookupObject(
				ctx,
				tree.ObjectLookupFlags{
					CommonLookupFlags: tree.CommonLookupFlags{
						Required:       true,
						RequireMutable: true,
						IncludeOffline: true,
					},
					DesiredObjectKind: tree.TypeObject,
				},
				objName.Catalog(),
				objName.Schema(),
				objName.Object(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(468734)
				return err
			} else {
				__antithesis_instrumentation__.Notify(468735)
			}
			__antithesis_instrumentation__.Notify(468731)

			if !found {
				__antithesis_instrumentation__.Notify(468736)
				continue
			} else {
				__antithesis_instrumentation__.Notify(468737)
			}
			__antithesis_instrumentation__.Notify(468732)
			typDesc, ok := desc.(*typedesc.Mutable)
			if !ok {
				__antithesis_instrumentation__.Notify(468738)
				return errors.AssertionFailedf(
					"descriptor for %q is not Mutable",
					objName.Object(),
				)
			} else {
				__antithesis_instrumentation__.Notify(468739)
			}
			__antithesis_instrumentation__.Notify(468733)

			d.typesToDelete = append(d.typesToDelete, typDesc)
		}
	}
	__antithesis_instrumentation__.Notify(468703)

	allObjectsToDelete, implicitDeleteMap, err := p.accumulateAllObjectsToDelete(ctx, d.td)
	if err != nil {
		__antithesis_instrumentation__.Notify(468740)
		return err
	} else {
		__antithesis_instrumentation__.Notify(468741)
	}
	__antithesis_instrumentation__.Notify(468704)
	d.allTableObjectsToDelete = allObjectsToDelete
	d.td = filterImplicitlyDeletedObjects(d.td, implicitDeleteMap)
	d.toDeleteByID = make(map[descpb.ID]*toDelete)
	for i := range d.td {
		__antithesis_instrumentation__.Notify(468742)
		d.toDeleteByID[d.td[i].desc.GetID()] = &d.td[i]
	}
	__antithesis_instrumentation__.Notify(468705)
	return nil
}

func (d *dropCascadeState) dropAllCollectedObjects(ctx context.Context, p *planner) error {
	__antithesis_instrumentation__.Notify(468743)

	for _, toDel := range d.td {
		__antithesis_instrumentation__.Notify(468746)
		desc := toDel.desc
		var cascadedObjects []string
		var err error
		if desc.IsView() {
			__antithesis_instrumentation__.Notify(468749)
			cascadedObjects, err = p.dropViewImpl(ctx, desc, false, "", tree.DropCascade)
		} else {
			__antithesis_instrumentation__.Notify(468750)
			if desc.IsSequence() {
				__antithesis_instrumentation__.Notify(468751)
				err = p.dropSequenceImpl(ctx, desc, false, "", tree.DropCascade)
			} else {
				__antithesis_instrumentation__.Notify(468752)
				cascadedObjects, err = p.dropTableImpl(ctx, desc, true, "", tree.DropCascade)
			}
		}
		__antithesis_instrumentation__.Notify(468747)
		if err != nil {
			__antithesis_instrumentation__.Notify(468753)
			return err
		} else {
			__antithesis_instrumentation__.Notify(468754)
		}
		__antithesis_instrumentation__.Notify(468748)
		d.droppedNames = append(d.droppedNames, cascadedObjects...)
		d.droppedNames = append(d.droppedNames, toDel.tn.FQString())
	}
	__antithesis_instrumentation__.Notify(468744)

	for _, typ := range d.typesToDelete {
		__antithesis_instrumentation__.Notify(468755)
		if err := d.canDropType(ctx, p, typ); err != nil {
			__antithesis_instrumentation__.Notify(468757)
			return err
		} else {
			__antithesis_instrumentation__.Notify(468758)
		}
		__antithesis_instrumentation__.Notify(468756)

		if err := p.dropTypeImpl(ctx, typ, "", false); err != nil {
			__antithesis_instrumentation__.Notify(468759)
			return err
		} else {
			__antithesis_instrumentation__.Notify(468760)
		}
	}
	__antithesis_instrumentation__.Notify(468745)

	return nil
}

func (d *dropCascadeState) canDropType(
	ctx context.Context, p *planner, typ *typedesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(468761)
	var referencedButNotDropping []descpb.ID
	for _, id := range typ.ReferencingDescriptorIDs {
		__antithesis_instrumentation__.Notify(468766)
		if _, exists := d.toDeleteByID[id]; exists {
			__antithesis_instrumentation__.Notify(468768)
			continue
		} else {
			__antithesis_instrumentation__.Notify(468769)
		}
		__antithesis_instrumentation__.Notify(468767)
		referencedButNotDropping = append(referencedButNotDropping, id)
	}
	__antithesis_instrumentation__.Notify(468762)
	if len(referencedButNotDropping) == 0 {
		__antithesis_instrumentation__.Notify(468770)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(468771)
	}
	__antithesis_instrumentation__.Notify(468763)
	dependentNames, err := p.getFullyQualifiedTableNamesFromIDs(ctx, referencedButNotDropping)
	if err != nil {
		__antithesis_instrumentation__.Notify(468772)
		return errors.Wrapf(err, "type %q has dependent objects", typ.Name)
	} else {
		__antithesis_instrumentation__.Notify(468773)
	}
	__antithesis_instrumentation__.Notify(468764)
	fqName, err := getTypeNameFromTypeDescriptor(
		oneAtATimeSchemaResolver{ctx, p},
		typ,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(468774)
		return errors.Wrapf(err, "type %q has dependent objects", typ.Name)
	} else {
		__antithesis_instrumentation__.Notify(468775)
	}
	__antithesis_instrumentation__.Notify(468765)
	return unimplemented.NewWithIssueDetailf(51480, "DROP TYPE CASCADE is not yet supported",
		"cannot drop type %q because other objects (%v) still depend on it",
		fqName.FQString(),
		dependentNames,
	)
}

func (d *dropCascadeState) getDroppedTableDetails() []jobspb.DroppedTableDetails {
	__antithesis_instrumentation__.Notify(468776)
	res := make([]jobspb.DroppedTableDetails, len(d.allTableObjectsToDelete))
	for i := range d.allTableObjectsToDelete {
		__antithesis_instrumentation__.Notify(468778)
		tbl := d.allTableObjectsToDelete[i]
		res[i] = jobspb.DroppedTableDetails{
			ID:   tbl.ID,
			Name: tbl.Name,
		}
	}
	__antithesis_instrumentation__.Notify(468777)
	return res
}
