package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func (tc *Collection) GetMutableDescriptorsByID(
	ctx context.Context, txn *kv.Txn, ids ...descpb.ID,
) ([]catalog.MutableDescriptor, error) {
	__antithesis_instrumentation__.Notify(264192)
	flags := tree.CommonLookupFlags{
		Required:       true,
		RequireMutable: true,
		IncludeOffline: true,
		IncludeDropped: true,
	}
	descs, err := tc.getDescriptorsByID(ctx, txn, flags, ids...)
	if err != nil {
		__antithesis_instrumentation__.Notify(264195)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264196)
	}
	__antithesis_instrumentation__.Notify(264193)
	ret := make([]catalog.MutableDescriptor, len(descs))
	for i, desc := range descs {
		__antithesis_instrumentation__.Notify(264197)
		ret[i] = desc.(catalog.MutableDescriptor)
	}
	__antithesis_instrumentation__.Notify(264194)
	return ret, nil
}

func (tc *Collection) GetMutableDescriptorByID(
	ctx context.Context, txn *kv.Txn, id descpb.ID,
) (catalog.MutableDescriptor, error) {
	__antithesis_instrumentation__.Notify(264198)
	descs, err := tc.GetMutableDescriptorsByID(ctx, txn, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(264200)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264201)
	}
	__antithesis_instrumentation__.Notify(264199)
	return descs[0], nil
}

func (tc *Collection) GetImmutableDescriptorsByID(
	ctx context.Context, txn *kv.Txn, flags tree.CommonLookupFlags, ids ...descpb.ID,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(264202)
	flags.RequireMutable = false
	return tc.getDescriptorsByID(ctx, txn, flags, ids...)
}

func (tc *Collection) GetImmutableDescriptorByID(
	ctx context.Context, txn *kv.Txn, id descpb.ID, flags tree.CommonLookupFlags,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(264203)
	descs, err := tc.GetImmutableDescriptorsByID(ctx, txn, flags, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(264205)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264206)
	}
	__antithesis_instrumentation__.Notify(264204)
	return descs[0], nil
}

func (tc *Collection) getDescriptorsByID(
	ctx context.Context, txn *kv.Txn, flags tree.CommonLookupFlags, ids ...descpb.ID,
) (descs []catalog.Descriptor, err error) {
	__antithesis_instrumentation__.Notify(264207)
	defer func() {
		__antithesis_instrumentation__.Notify(264215)
		if err == nil {
			__antithesis_instrumentation__.Notify(264217)
			err = filterDescriptorsStates(descs, flags)
		} else {
			__antithesis_instrumentation__.Notify(264218)
		}
		__antithesis_instrumentation__.Notify(264216)
		if err != nil {
			__antithesis_instrumentation__.Notify(264219)
			descs = nil
		} else {
			__antithesis_instrumentation__.Notify(264220)
		}
	}()
	__antithesis_instrumentation__.Notify(264208)

	log.VEventf(ctx, 2, "looking up descriptors for ids %d", ids)
	descs = make([]catalog.Descriptor, len(ids))
	{
		__antithesis_instrumentation__.Notify(264221)

		q := byIDLookupContext{
			ctx:   ctx,
			txn:   txn,
			tc:    tc,
			flags: flags,
		}
		for _, fn := range []func(id descpb.ID) (catalog.Descriptor, error){
			q.lookupVirtual,
			q.lookupSynthetic,
			q.lookupUncommitted,
			q.lookupLeased,
		} {
			__antithesis_instrumentation__.Notify(264222)
			for i, id := range ids {
				__antithesis_instrumentation__.Notify(264223)
				if descs[i] != nil {
					__antithesis_instrumentation__.Notify(264226)
					continue
				} else {
					__antithesis_instrumentation__.Notify(264227)
				}
				__antithesis_instrumentation__.Notify(264224)
				desc, err := fn(id)
				if err != nil {
					__antithesis_instrumentation__.Notify(264228)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(264229)
				}
				__antithesis_instrumentation__.Notify(264225)
				descs[i] = desc
			}
		}
	}
	__antithesis_instrumentation__.Notify(264209)

	remainingIDs := make([]descpb.ID, 0, len(ids))
	indexes := make([]int, 0, len(ids))
	for i, id := range ids {
		__antithesis_instrumentation__.Notify(264230)
		if descs[i] != nil {
			__antithesis_instrumentation__.Notify(264232)
			continue
		} else {
			__antithesis_instrumentation__.Notify(264233)
		}
		__antithesis_instrumentation__.Notify(264231)
		remainingIDs = append(remainingIDs, id)
		indexes = append(indexes, i)
	}
	__antithesis_instrumentation__.Notify(264210)
	if len(remainingIDs) == 0 {
		__antithesis_instrumentation__.Notify(264234)

		return descs, nil
	} else {
		__antithesis_instrumentation__.Notify(264235)
	}
	__antithesis_instrumentation__.Notify(264211)
	kvDescs, err := tc.withReadFromStore(flags.RequireMutable, func() ([]catalog.Descriptor, error) {
		__antithesis_instrumentation__.Notify(264236)
		ret := make([]catalog.Descriptor, len(remainingIDs))

		kvIDs := make([]descpb.ID, 0, len(remainingIDs))
		kvIndexes := make([]int, 0, len(remainingIDs))
		for i, id := range remainingIDs {
			__antithesis_instrumentation__.Notify(264239)
			if imm, status := tc.uncommitted.getImmutableByID(id); imm != nil && func() bool {
				__antithesis_instrumentation__.Notify(264241)
				return status == notValidatedYet == true
			}() == true {
				__antithesis_instrumentation__.Notify(264242)
				err := tc.Validate(ctx, txn, catalog.ValidationReadTelemetry, catalog.ValidationLevelCrossReferences, imm)
				if err != nil {
					__antithesis_instrumentation__.Notify(264244)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(264245)
				}
				__antithesis_instrumentation__.Notify(264243)
				ret[i] = imm
				continue
			} else {
				__antithesis_instrumentation__.Notify(264246)
			}
			__antithesis_instrumentation__.Notify(264240)
			kvIDs = append(kvIDs, id)
			kvIndexes = append(kvIndexes, i)
		}
		__antithesis_instrumentation__.Notify(264237)

		if len(kvIDs) > 0 {
			__antithesis_instrumentation__.Notify(264247)
			vd := tc.newValidationDereferencer(txn)
			kvDescs, err := tc.kv.getByIDs(ctx, tc.version, txn, vd, kvIDs)
			if err != nil {
				__antithesis_instrumentation__.Notify(264249)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(264250)
			}
			__antithesis_instrumentation__.Notify(264248)
			for k, imm := range kvDescs {
				__antithesis_instrumentation__.Notify(264251)
				ret[kvIndexes[k]] = imm
			}
		} else {
			__antithesis_instrumentation__.Notify(264252)
		}
		__antithesis_instrumentation__.Notify(264238)
		return ret, nil
	})
	__antithesis_instrumentation__.Notify(264212)
	if err != nil {
		__antithesis_instrumentation__.Notify(264253)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264254)
	}
	__antithesis_instrumentation__.Notify(264213)
	for j, desc := range kvDescs {
		__antithesis_instrumentation__.Notify(264255)

		if table, isTable := desc.(catalog.TableDescriptor); isTable {
			__antithesis_instrumentation__.Notify(264257)
			desc, err = tc.hydrateTypesInTableDesc(ctx, txn, table)
			if err != nil {
				__antithesis_instrumentation__.Notify(264258)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(264259)
			}
		} else {
			__antithesis_instrumentation__.Notify(264260)
		}
		__antithesis_instrumentation__.Notify(264256)
		descs[indexes[j]] = desc
	}
	__antithesis_instrumentation__.Notify(264214)
	return descs, nil
}

type byIDLookupContext struct {
	ctx   context.Context
	txn   *kv.Txn
	tc    *Collection
	flags tree.CommonLookupFlags
}

func (q *byIDLookupContext) lookupVirtual(id descpb.ID) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(264261)
	return q.tc.virtual.getByID(q.ctx, id, q.flags.RequireMutable)
}

func (q *byIDLookupContext) lookupSynthetic(id descpb.ID) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(264262)
	if q.flags.AvoidSynthetic {
		__antithesis_instrumentation__.Notify(264266)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(264267)
	}
	__antithesis_instrumentation__.Notify(264263)
	_, sd := q.tc.synthetic.getByID(id)
	if sd == nil {
		__antithesis_instrumentation__.Notify(264268)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(264269)
	}
	__antithesis_instrumentation__.Notify(264264)
	if q.flags.RequireMutable {
		__antithesis_instrumentation__.Notify(264270)
		return nil, newMutableSyntheticDescriptorAssertionError(sd.GetID())
	} else {
		__antithesis_instrumentation__.Notify(264271)
	}
	__antithesis_instrumentation__.Notify(264265)
	return sd, nil
}

func (q *byIDLookupContext) lookupUncommitted(id descpb.ID) (_ catalog.Descriptor, err error) {
	__antithesis_instrumentation__.Notify(264272)
	ud, status := q.tc.uncommitted.getImmutableByID(id)
	if ud == nil || func() bool {
		__antithesis_instrumentation__.Notify(264275)
		return status == notValidatedYet == true
	}() == true {
		__antithesis_instrumentation__.Notify(264276)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(264277)
	}
	__antithesis_instrumentation__.Notify(264273)
	log.VEventf(q.ctx, 2, "found uncommitted descriptor %d", id)
	if !q.flags.RequireMutable {
		__antithesis_instrumentation__.Notify(264278)
		return ud, nil
	} else {
		__antithesis_instrumentation__.Notify(264279)
	}
	__antithesis_instrumentation__.Notify(264274)
	return q.tc.uncommitted.checkOut(id)
}

func (q *byIDLookupContext) lookupLeased(id descpb.ID) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(264280)
	if q.flags.AvoidLeased || func() bool {
		__antithesis_instrumentation__.Notify(264284)
		return q.flags.RequireMutable == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(264285)
		return lease.TestingTableLeasesAreDisabled() == true
	}() == true {
		__antithesis_instrumentation__.Notify(264286)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(264287)
	}
	__antithesis_instrumentation__.Notify(264281)

	if q.tc.kv.idDefinitelyDoesNotExist(id) {
		__antithesis_instrumentation__.Notify(264288)
		return nil, catalog.ErrDescriptorNotFound
	} else {
		__antithesis_instrumentation__.Notify(264289)
	}
	__antithesis_instrumentation__.Notify(264282)
	desc, shouldReadFromStore, err := q.tc.leased.getByID(q.ctx, q.tc.deadlineHolder(q.txn), id)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(264290)
		return shouldReadFromStore == true
	}() == true {
		__antithesis_instrumentation__.Notify(264291)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264292)
	}
	__antithesis_instrumentation__.Notify(264283)
	return desc, nil
}

func filterDescriptorsStates(descs []catalog.Descriptor, flags tree.CommonLookupFlags) error {
	__antithesis_instrumentation__.Notify(264293)
	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(264295)

		_, err := filterDescriptorState(desc, true, flags)
		if err == nil {
			__antithesis_instrumentation__.Notify(264298)
			continue
		} else {
			__antithesis_instrumentation__.Notify(264299)
		}
		__antithesis_instrumentation__.Notify(264296)
		if desc.Adding() && func() bool {
			__antithesis_instrumentation__.Notify(264300)
			return (desc.IsUncommittedVersion() || func() bool {
				__antithesis_instrumentation__.Notify(264301)
				return flags.AvoidLeased == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(264302)
				return flags.RequireMutable == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(264303)

			continue
		} else {
			__antithesis_instrumentation__.Notify(264304)
		}
		__antithesis_instrumentation__.Notify(264297)
		return err
	}
	__antithesis_instrumentation__.Notify(264294)
	return nil
}

func (tc *Collection) getByName(
	ctx context.Context,
	txn *kv.Txn,
	db catalog.DatabaseDescriptor,
	sc catalog.SchemaDescriptor,
	name string,
	avoidLeased, mutable, avoidSynthetic bool,
	alwaysLookupLeasedPublicSchema bool,
) (found bool, desc catalog.Descriptor, err error) {
	__antithesis_instrumentation__.Notify(264305)
	var parentID, parentSchemaID descpb.ID
	if db != nil {
		__antithesis_instrumentation__.Notify(264311)
		if sc == nil {
			__antithesis_instrumentation__.Notify(264313)

			return getSchemaByName(
				ctx, tc, txn, db, name, avoidLeased, mutable, avoidSynthetic,
				alwaysLookupLeasedPublicSchema,
			)
		} else {
			__antithesis_instrumentation__.Notify(264314)
		}
		__antithesis_instrumentation__.Notify(264312)
		parentID, parentSchemaID = db.GetID(), sc.GetID()
	} else {
		__antithesis_instrumentation__.Notify(264315)
	}
	__antithesis_instrumentation__.Notify(264306)

	if found, sd := tc.synthetic.getByName(parentID, parentSchemaID, name); found && func() bool {
		__antithesis_instrumentation__.Notify(264316)
		return !avoidSynthetic == true
	}() == true {
		__antithesis_instrumentation__.Notify(264317)
		if mutable {
			__antithesis_instrumentation__.Notify(264319)
			return false, nil, newMutableSyntheticDescriptorAssertionError(sd.GetID())
		} else {
			__antithesis_instrumentation__.Notify(264320)
		}
		__antithesis_instrumentation__.Notify(264318)
		return true, sd, nil
	} else {
		__antithesis_instrumentation__.Notify(264321)
	}

	{
		__antithesis_instrumentation__.Notify(264322)
		refuseFurtherLookup, ud := tc.uncommitted.getByName(parentID, parentSchemaID, name)
		if ud != nil {
			__antithesis_instrumentation__.Notify(264324)
			log.VEventf(ctx, 2, "found uncommitted descriptor %d", ud.GetID())
			if mutable {
				__antithesis_instrumentation__.Notify(264326)
				ud, err = tc.uncommitted.checkOut(ud.GetID())
				if err != nil {
					__antithesis_instrumentation__.Notify(264327)
					return false, nil, err
				} else {
					__antithesis_instrumentation__.Notify(264328)
				}
			} else {
				__antithesis_instrumentation__.Notify(264329)
			}
			__antithesis_instrumentation__.Notify(264325)
			return true, ud, nil
		} else {
			__antithesis_instrumentation__.Notify(264330)
		}
		__antithesis_instrumentation__.Notify(264323)
		if refuseFurtherLookup {
			__antithesis_instrumentation__.Notify(264331)
			return false, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(264332)
		}
	}
	__antithesis_instrumentation__.Notify(264307)

	if !avoidLeased && func() bool {
		__antithesis_instrumentation__.Notify(264333)
		return !mutable == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(264334)
		return !lease.TestingTableLeasesAreDisabled() == true
	}() == true {
		__antithesis_instrumentation__.Notify(264335)
		var shouldReadFromStore bool
		desc, shouldReadFromStore, err = tc.leased.getByName(ctx, tc.deadlineHolder(txn), parentID, parentSchemaID, name)
		if err != nil {
			__antithesis_instrumentation__.Notify(264337)
			return false, nil, err
		} else {
			__antithesis_instrumentation__.Notify(264338)
		}
		__antithesis_instrumentation__.Notify(264336)
		if !shouldReadFromStore {
			__antithesis_instrumentation__.Notify(264339)
			return desc != nil, desc, nil
		} else {
			__antithesis_instrumentation__.Notify(264340)
		}
	} else {
		__antithesis_instrumentation__.Notify(264341)
	}
	__antithesis_instrumentation__.Notify(264308)

	var descs []catalog.Descriptor
	descs, err = tc.withReadFromStore(mutable, func() ([]catalog.Descriptor, error) {
		__antithesis_instrumentation__.Notify(264342)

		if imm := tc.uncommitted.getUnvalidatedByName(parentID, parentSchemaID, name); imm != nil {
			__antithesis_instrumentation__.Notify(264345)
			return []catalog.Descriptor{imm},
				tc.Validate(ctx, txn, catalog.ValidationReadTelemetry, catalog.ValidationLevelCrossReferences, imm)
		} else {
			__antithesis_instrumentation__.Notify(264346)
		}
		__antithesis_instrumentation__.Notify(264343)

		uncommittedParent, _ := tc.uncommitted.getImmutableByID(parentID)
		uncommittedDB, _ := catalog.AsDatabaseDescriptor(uncommittedParent)
		version := tc.settings.Version.ActiveVersion(ctx)
		vd := tc.newValidationDereferencer(txn)
		imm, err := tc.kv.getByName(ctx, version, txn, vd, uncommittedDB, parentID, parentSchemaID, name)
		if err != nil {
			__antithesis_instrumentation__.Notify(264347)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(264348)
		}
		__antithesis_instrumentation__.Notify(264344)
		return []catalog.Descriptor{imm}, nil
	})
	__antithesis_instrumentation__.Notify(264309)
	if err != nil {
		__antithesis_instrumentation__.Notify(264349)
		return false, nil, err
	} else {
		__antithesis_instrumentation__.Notify(264350)
	}
	__antithesis_instrumentation__.Notify(264310)
	return descs[0] != nil, descs[0], err
}

func (tc *Collection) withReadFromStore(
	requireMutable bool, readFn func() ([]catalog.Descriptor, error),
) (descs []catalog.Descriptor, _ error) {
	__antithesis_instrumentation__.Notify(264351)
	descs, err := readFn()
	if err != nil {
		__antithesis_instrumentation__.Notify(264354)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264355)
	}
	__antithesis_instrumentation__.Notify(264352)
	for i, desc := range descs {
		__antithesis_instrumentation__.Notify(264356)
		if desc == nil {
			__antithesis_instrumentation__.Notify(264360)
			continue
		} else {
			__antithesis_instrumentation__.Notify(264361)
		}
		__antithesis_instrumentation__.Notify(264357)
		desc, err = tc.uncommitted.add(desc.NewBuilder().BuildExistingMutable(), notCheckedOutYet)
		if err != nil {
			__antithesis_instrumentation__.Notify(264362)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(264363)
		}
		__antithesis_instrumentation__.Notify(264358)
		if requireMutable {
			__antithesis_instrumentation__.Notify(264364)
			desc, err = tc.uncommitted.checkOut(desc.GetID())
			if err != nil {
				__antithesis_instrumentation__.Notify(264365)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(264366)
			}
		} else {
			__antithesis_instrumentation__.Notify(264367)
		}
		__antithesis_instrumentation__.Notify(264359)
		descs[i] = desc
	}
	__antithesis_instrumentation__.Notify(264353)
	return descs, nil
}

func (tc *Collection) deadlineHolder(txn *kv.Txn) deadlineHolder {
	__antithesis_instrumentation__.Notify(264368)
	if tc.maxTimestampBoundDeadlineHolder.maxTimestampBound.IsEmpty() {
		__antithesis_instrumentation__.Notify(264370)
		return txn
	} else {
		__antithesis_instrumentation__.Notify(264371)
	}
	__antithesis_instrumentation__.Notify(264369)
	return &tc.maxTimestampBoundDeadlineHolder
}

func getSchemaByName(
	ctx context.Context,
	tc *Collection,
	txn *kv.Txn,
	db catalog.DatabaseDescriptor,
	name string,
	avoidLeased, mutable, avoidSynthetic bool,
	alwaysLookupLeasedPublicSchema bool,
) (bool, catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(264372)
	if !db.HasPublicSchemaWithDescriptor() && func() bool {
		__antithesis_instrumentation__.Notify(264377)
		return name == tree.PublicSchema == true
	}() == true {
		__antithesis_instrumentation__.Notify(264378)

		if alwaysLookupLeasedPublicSchema {
			__antithesis_instrumentation__.Notify(264380)
			desc, _, err := tc.leased.getByName(ctx, txn, db.GetID(), 0, catconstants.PublicSchemaName)
			if err != nil {
				__antithesis_instrumentation__.Notify(264382)
				return false, desc, err
			} else {
				__antithesis_instrumentation__.Notify(264383)
			}
			__antithesis_instrumentation__.Notify(264381)
			return true, desc, nil
		} else {
			__antithesis_instrumentation__.Notify(264384)
		}
		__antithesis_instrumentation__.Notify(264379)
		return true, schemadesc.GetPublicSchema(), nil
	} else {
		__antithesis_instrumentation__.Notify(264385)
	}
	__antithesis_instrumentation__.Notify(264373)
	if sc := tc.virtual.getSchemaByName(name); sc != nil {
		__antithesis_instrumentation__.Notify(264386)
		return true, sc, nil
	} else {
		__antithesis_instrumentation__.Notify(264387)
	}
	__antithesis_instrumentation__.Notify(264374)
	if isTemporarySchema(name) {
		__antithesis_instrumentation__.Notify(264388)
		if isDone, sc := tc.temporary.getSchemaByName(ctx, db.GetID(), name); sc != nil || func() bool {
			__antithesis_instrumentation__.Notify(264391)
			return isDone == true
		}() == true {
			__antithesis_instrumentation__.Notify(264392)
			return sc != nil, sc, nil
		} else {
			__antithesis_instrumentation__.Notify(264393)
		}
		__antithesis_instrumentation__.Notify(264389)
		scID, err := tc.kv.lookupName(ctx, txn, nil, db.GetID(), keys.RootNamespaceID, name)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(264394)
			return scID == descpb.InvalidID == true
		}() == true {
			__antithesis_instrumentation__.Notify(264395)
			return false, nil, err
		} else {
			__antithesis_instrumentation__.Notify(264396)
		}
		__antithesis_instrumentation__.Notify(264390)
		return true, schemadesc.NewTemporarySchema(name, scID, db.GetID()), nil
	} else {
		__antithesis_instrumentation__.Notify(264397)
	}
	__antithesis_instrumentation__.Notify(264375)
	if id := db.GetSchemaID(name); id != descpb.InvalidID {
		__antithesis_instrumentation__.Notify(264398)

		sc, err := tc.getSchemaByID(ctx, txn, id, tree.SchemaLookupFlags{
			RequireMutable: mutable,
			AvoidLeased:    avoidLeased,
			AvoidSynthetic: avoidSynthetic,
		})
		if errors.Is(err, catalog.ErrDescriptorDropped) {
			__antithesis_instrumentation__.Notify(264400)
			err = nil
		} else {
			__antithesis_instrumentation__.Notify(264401)
		}
		__antithesis_instrumentation__.Notify(264399)
		return sc != nil, sc, err
	} else {
		__antithesis_instrumentation__.Notify(264402)
	}
	__antithesis_instrumentation__.Notify(264376)
	return false, nil, nil
}

func isTemporarySchema(name string) bool {
	__antithesis_instrumentation__.Notify(264403)
	return strings.HasPrefix(name, catconstants.PgTempSchemaName)
}

func filterDescriptorState(
	desc catalog.Descriptor, required bool, flags tree.CommonLookupFlags,
) (dropped bool, _ error) {
	__antithesis_instrumentation__.Notify(264404)
	flags = tree.CommonLookupFlags{
		Required:       required,
		IncludeOffline: flags.IncludeOffline,
		IncludeDropped: flags.IncludeDropped,
	}
	if err := catalog.FilterDescriptorState(desc, flags); err != nil {
		__antithesis_instrumentation__.Notify(264406)
		if required || func() bool {
			__antithesis_instrumentation__.Notify(264408)
			return !errors.Is(err, catalog.ErrDescriptorDropped) == true
		}() == true {
			__antithesis_instrumentation__.Notify(264409)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(264410)
		}
		__antithesis_instrumentation__.Notify(264407)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(264411)
	}
	__antithesis_instrumentation__.Notify(264405)
	return false, nil
}
