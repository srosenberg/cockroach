package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

func findTransitioningMembers(desc *typedesc.Mutable) ([][]byte, bool) {
	__antithesis_instrumentation__.Notify(629015)
	var transitioningMembers [][]byte
	beingDropped := false

	if desc.IsNew() {
		__antithesis_instrumentation__.Notify(629018)
		for _, member := range desc.EnumMembers {
			__antithesis_instrumentation__.Notify(629020)
			if member.Capability != descpb.TypeDescriptor_EnumMember_ALL {
				__antithesis_instrumentation__.Notify(629021)
				transitioningMembers = append(transitioningMembers, member.PhysicalRepresentation)
				if member.Direction == descpb.TypeDescriptor_EnumMember_REMOVE {
					__antithesis_instrumentation__.Notify(629022)
					beingDropped = true
				} else {
					__antithesis_instrumentation__.Notify(629023)
				}
			} else {
				__antithesis_instrumentation__.Notify(629024)
			}
		}
		__antithesis_instrumentation__.Notify(629019)
		return transitioningMembers, beingDropped
	} else {
		__antithesis_instrumentation__.Notify(629025)
	}
	__antithesis_instrumentation__.Notify(629016)

	for _, member := range desc.EnumMembers {
		__antithesis_instrumentation__.Notify(629026)
		found := false
		for _, clusterMember := range desc.ClusterVersion.EnumMembers {
			__antithesis_instrumentation__.Notify(629028)
			if bytes.Equal(member.PhysicalRepresentation, clusterMember.PhysicalRepresentation) {
				__antithesis_instrumentation__.Notify(629029)
				found = true
				if member.Capability != clusterMember.Capability {
					__antithesis_instrumentation__.Notify(629031)
					transitioningMembers = append(transitioningMembers, member.PhysicalRepresentation)
					if member.Capability == descpb.TypeDescriptor_EnumMember_READ_ONLY && func() bool {
						__antithesis_instrumentation__.Notify(629032)
						return member.Direction == descpb.TypeDescriptor_EnumMember_REMOVE == true
					}() == true {
						__antithesis_instrumentation__.Notify(629033)
						beingDropped = true
					} else {
						__antithesis_instrumentation__.Notify(629034)
					}
				} else {
					__antithesis_instrumentation__.Notify(629035)
				}
				__antithesis_instrumentation__.Notify(629030)
				break
			} else {
				__antithesis_instrumentation__.Notify(629036)
			}
		}
		__antithesis_instrumentation__.Notify(629027)

		if !found {
			__antithesis_instrumentation__.Notify(629037)
			transitioningMembers = append(transitioningMembers, member.PhysicalRepresentation)
		} else {
			__antithesis_instrumentation__.Notify(629038)
		}
	}
	__antithesis_instrumentation__.Notify(629017)
	return transitioningMembers, beingDropped
}

func (p *planner) writeTypeSchemaChange(
	ctx context.Context, typeDesc *typedesc.Mutable, jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(629039)

	record, recordExists := p.extendedEvalCtx.SchemaChangeJobRecords[typeDesc.ID]
	transitioningMembers, beingDropped := findTransitioningMembers(typeDesc)
	if recordExists {
		__antithesis_instrumentation__.Notify(629041)

		newDetails := jobspb.TypeSchemaChangeDetails{
			TypeID:               typeDesc.ID,
			TransitioningMembers: transitioningMembers,
		}
		record.Details = newDetails
		record.AppendDescription(jobDesc)
		record.SetNonCancelable(ctx,
			func(ctx context.Context, nonCancelable bool) bool {
				__antithesis_instrumentation__.Notify(629043)

				if !nonCancelable {
					__antithesis_instrumentation__.Notify(629045)
					return nonCancelable
				} else {
					__antithesis_instrumentation__.Notify(629046)
				}
				__antithesis_instrumentation__.Notify(629044)

				return !beingDropped
			})
		__antithesis_instrumentation__.Notify(629042)
		log.Infof(ctx, "job %d: updated with type change for type %d", record.JobID, typeDesc.ID)
	} else {
		__antithesis_instrumentation__.Notify(629047)

		newRecord := jobs.Record{
			JobID:         p.extendedEvalCtx.ExecCfg.JobRegistry.MakeJobID(),
			Description:   jobDesc,
			Username:      p.User(),
			DescriptorIDs: descpb.IDs{typeDesc.ID},
			Details: jobspb.TypeSchemaChangeDetails{
				TypeID:               typeDesc.ID,
				TransitioningMembers: transitioningMembers,
			},
			Progress: jobspb.TypeSchemaChangeProgress{},

			NonCancelable: !beingDropped,
		}
		p.extendedEvalCtx.SchemaChangeJobRecords[typeDesc.ID] = &newRecord
		log.Infof(ctx, "queued new type change job %d for type %d", newRecord.JobID, typeDesc.ID)
	}
	__antithesis_instrumentation__.Notify(629040)

	return p.writeTypeDesc(ctx, typeDesc)
}

func (p *planner) writeTypeDesc(ctx context.Context, typeDesc *typedesc.Mutable) error {
	__antithesis_instrumentation__.Notify(629048)

	b := p.txn.NewBatch()
	if err := p.Descriptors().WriteDescToBatch(
		ctx, p.extendedEvalCtx.Tracing.KVTracingEnabled(), typeDesc, b,
	); err != nil {
		__antithesis_instrumentation__.Notify(629050)
		return err
	} else {
		__antithesis_instrumentation__.Notify(629051)
	}
	__antithesis_instrumentation__.Notify(629049)
	return p.txn.Run(ctx, b)
}

type typeSchemaChanger struct {
	typeID descpb.ID

	transitioningMembers [][]byte
	execCfg              *ExecutorConfig
}

type TypeSchemaChangerTestingKnobs struct {
	TypeSchemaChangeJobNoOp func() bool

	RunBeforeExec func() error

	RunBeforeEnumMemberPromotion func(ctx context.Context) error

	RunAfterOnFailOrCancel func() error

	RunBeforeMultiRegionUpdates func() error
}

func (TypeSchemaChangerTestingKnobs) ModuleTestingKnobs() {
	__antithesis_instrumentation__.Notify(629052)
}

func (t *typeSchemaChanger) getTypeDescFromStore(
	ctx context.Context,
) (catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(629053)
	var typeDesc catalog.TypeDescriptor
	if err := DescsTxn(ctx, t.execCfg, func(ctx context.Context, txn *kv.Txn, col *descs.Collection) (err error) {
		__antithesis_instrumentation__.Notify(629055)
		typeDesc, err = col.Direct().MustGetTypeDescByID(ctx, txn, t.typeID)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(629056)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(629057)
	}
	__antithesis_instrumentation__.Notify(629054)
	return typeDesc, nil
}

func refreshTypeDescriptorLeases(
	ctx context.Context, leaseMgr *lease.Manager, typeDesc catalog.TypeDescriptor,
) error {
	__antithesis_instrumentation__.Notify(629058)
	var err error
	var ids = []descpb.ID{typeDesc.GetID()}
	if typeDesc.GetArrayTypeID() != descpb.InvalidID {
		__antithesis_instrumentation__.Notify(629061)
		ids = append(ids, typeDesc.GetArrayTypeID())
	} else {
		__antithesis_instrumentation__.Notify(629062)
	}
	__antithesis_instrumentation__.Notify(629059)
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(629063)
		if _, updateErr := WaitToUpdateLeases(ctx, leaseMgr, id); updateErr != nil {
			__antithesis_instrumentation__.Notify(629064)

			if errors.Is(updateErr, catalog.ErrDescriptorNotFound) {
				__antithesis_instrumentation__.Notify(629065)
				log.Infof(ctx,
					"could not find type descriptor %d to refresh lease; "+
						"assuming it was dropped and moving on",
					id,
				)
			} else {
				__antithesis_instrumentation__.Notify(629066)
				err = errors.CombineErrors(err, updateErr)
			}
		} else {
			__antithesis_instrumentation__.Notify(629067)
		}
	}
	__antithesis_instrumentation__.Notify(629060)
	return err
}

func (t *typeSchemaChanger) exec(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(629068)
	if t.execCfg.TypeSchemaChangerTestingKnobs.RunBeforeExec != nil {
		__antithesis_instrumentation__.Notify(629075)
		if err := t.execCfg.TypeSchemaChangerTestingKnobs.RunBeforeExec(); err != nil {
			__antithesis_instrumentation__.Notify(629076)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629077)
		}
	} else {
		__antithesis_instrumentation__.Notify(629078)
	}
	__antithesis_instrumentation__.Notify(629069)
	ctx = logtags.AddTags(ctx, t.logTags())
	leaseMgr := t.execCfg.LeaseManager
	codec := t.execCfg.Codec

	typeDesc, err := t.getTypeDescFromStore(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(629079)
		return err
	} else {
		__antithesis_instrumentation__.Notify(629080)
	}
	__antithesis_instrumentation__.Notify(629070)

	if len(typeDesc.GetDrainingNames()) > 0 {
		__antithesis_instrumentation__.Notify(629081)
		if err := drainNamesForDescriptor(
			ctx, typeDesc.GetID(), t.execCfg.CollectionFactory, t.execCfg.DB,
			t.execCfg.InternalExecutor, codec, nil,
		); err != nil {
			__antithesis_instrumentation__.Notify(629082)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629083)
		}
	} else {
		__antithesis_instrumentation__.Notify(629084)
	}
	__antithesis_instrumentation__.Notify(629071)

	if err := refreshTypeDescriptorLeases(ctx, leaseMgr, typeDesc); err != nil {
		__antithesis_instrumentation__.Notify(629085)
		return err
	} else {
		__antithesis_instrumentation__.Notify(629086)
	}
	__antithesis_instrumentation__.Notify(629072)

	if (typeDesc.GetKind() == descpb.TypeDescriptor_ENUM || func() bool {
		__antithesis_instrumentation__.Notify(629087)
		return typeDesc.GetKind() == descpb.TypeDescriptor_MULTIREGION_ENUM == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(629088)
		return len(t.transitioningMembers) != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(629089)
		if fn := t.execCfg.TypeSchemaChangerTestingKnobs.RunBeforeEnumMemberPromotion; fn != nil {
			__antithesis_instrumentation__.Notify(629099)
			if err := fn(ctx); err != nil {
				__antithesis_instrumentation__.Notify(629100)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629101)
			}
		} else {
			__antithesis_instrumentation__.Notify(629102)
		}
		__antithesis_instrumentation__.Notify(629090)

		var multiRegionPreDropIsNecessary bool
		withDatabaseRegionChangeFinalizer := func(
			ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
			f func(finalizer *databaseRegionChangeFinalizer) error,
		) error {
			__antithesis_instrumentation__.Notify(629103)
			typeDesc, err := descsCol.GetMutableTypeVersionByID(ctx, txn, t.typeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(629106)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629107)
			}
			__antithesis_instrumentation__.Notify(629104)
			regionChangeFinalizer, err := newDatabaseRegionChangeFinalizer(
				ctx,
				txn,
				t.execCfg,
				descsCol,
				typeDesc.GetParentID(),
				typeDesc.GetID(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(629108)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629109)
			}
			__antithesis_instrumentation__.Notify(629105)
			defer regionChangeFinalizer.cleanup()
			return f(regionChangeFinalizer)
		}
		__antithesis_instrumentation__.Notify(629091)
		prepareRepartitionedRegionalByRowTables := func(
			ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
		) (repartitioned []*tabledesc.Mutable, err error) {
			__antithesis_instrumentation__.Notify(629110)
			err = withDatabaseRegionChangeFinalizer(ctx, txn, descsCol, func(
				finalizer *databaseRegionChangeFinalizer,
			) (err error) {
				__antithesis_instrumentation__.Notify(629112)
				repartitioned, _, err = finalizer.repartitionRegionalByRowTables(ctx, txn)
				return err
			})
			__antithesis_instrumentation__.Notify(629111)
			return repartitioned, err
		}
		__antithesis_instrumentation__.Notify(629092)
		repartitionRegionalByRowTables := func(
			ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
		) error {
			__antithesis_instrumentation__.Notify(629113)
			return withDatabaseRegionChangeFinalizer(ctx, txn, descsCol, func(
				finalizer *databaseRegionChangeFinalizer,
			) error {
				__antithesis_instrumentation__.Notify(629114)
				return finalizer.preDrop(ctx, txn)
			})
		}
		__antithesis_instrumentation__.Notify(629093)

		validateDrops := func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
			__antithesis_instrumentation__.Notify(629115)
			typeDesc, err := descsCol.GetMutableTypeVersionByID(ctx, txn, t.typeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(629120)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629121)
			}
			__antithesis_instrumentation__.Notify(629116)
			var toDrop []descpb.TypeDescriptor_EnumMember
			for _, member := range typeDesc.EnumMembers {
				__antithesis_instrumentation__.Notify(629122)
				if t.isTransitioningInCurrentJob(&member) && func() bool {
					__antithesis_instrumentation__.Notify(629123)
					return enumMemberIsRemoving(&member) == true
				}() == true {
					__antithesis_instrumentation__.Notify(629124)
					if typeDesc.Kind == descpb.TypeDescriptor_MULTIREGION_ENUM {
						__antithesis_instrumentation__.Notify(629126)
						multiRegionPreDropIsNecessary = true
					} else {
						__antithesis_instrumentation__.Notify(629127)
					}
					__antithesis_instrumentation__.Notify(629125)
					toDrop = append(toDrop, member)
				} else {
					__antithesis_instrumentation__.Notify(629128)
				}
			}
			__antithesis_instrumentation__.Notify(629117)

			if multiRegionPreDropIsNecessary {
				__antithesis_instrumentation__.Notify(629129)
				repartitioned, err := prepareRepartitionedRegionalByRowTables(ctx, txn, descsCol)
				if err != nil {
					__antithesis_instrumentation__.Notify(629132)
					return err
				} else {
					__antithesis_instrumentation__.Notify(629133)
				}
				__antithesis_instrumentation__.Notify(629130)
				synthetic := make([]catalog.Descriptor, len(repartitioned))
				for i, d := range repartitioned {
					__antithesis_instrumentation__.Notify(629134)
					synthetic[i] = d
				}
				__antithesis_instrumentation__.Notify(629131)
				descsCol.SetSyntheticDescriptors(synthetic)
			} else {
				__antithesis_instrumentation__.Notify(629135)
			}
			__antithesis_instrumentation__.Notify(629118)
			for _, member := range toDrop {
				__antithesis_instrumentation__.Notify(629136)
				if err := t.canRemoveEnumValue(ctx, typeDesc, txn, &member, descsCol); err != nil {
					__antithesis_instrumentation__.Notify(629137)
					return err
				} else {
					__antithesis_instrumentation__.Notify(629138)
				}
			}
			__antithesis_instrumentation__.Notify(629119)
			return nil
		}
		__antithesis_instrumentation__.Notify(629094)
		if err := DescsTxn(ctx, t.execCfg, validateDrops); err != nil {
			__antithesis_instrumentation__.Notify(629139)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629140)
		}
		__antithesis_instrumentation__.Notify(629095)
		if multiRegionPreDropIsNecessary {
			__antithesis_instrumentation__.Notify(629141)
			if err := DescsTxn(ctx, t.execCfg, repartitionRegionalByRowTables); err != nil {
				__antithesis_instrumentation__.Notify(629142)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629143)
			}
		} else {
			__antithesis_instrumentation__.Notify(629144)
		}
		__antithesis_instrumentation__.Notify(629096)

		run := func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
			__antithesis_instrumentation__.Notify(629145)
			typeDesc, err := descsCol.GetMutableTypeVersionByID(ctx, txn, t.typeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(629155)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629156)
			}
			__antithesis_instrumentation__.Notify(629146)

			for i := range typeDesc.EnumMembers {
				__antithesis_instrumentation__.Notify(629157)
				member := &typeDesc.EnumMembers[i]
				if t.isTransitioningInCurrentJob(member) && func() bool {
					__antithesis_instrumentation__.Notify(629158)
					return enumMemberIsAdding(member) == true
				}() == true {
					__antithesis_instrumentation__.Notify(629159)
					member.Capability = descpb.TypeDescriptor_EnumMember_ALL
					member.Direction = descpb.TypeDescriptor_EnumMember_NONE
				} else {
					__antithesis_instrumentation__.Notify(629160)
				}
			}
			__antithesis_instrumentation__.Notify(629147)

			applyFilterOnEnumMembers(typeDesc, func(member *descpb.TypeDescriptor_EnumMember) bool {
				__antithesis_instrumentation__.Notify(629161)
				return t.isTransitioningInCurrentJob(member) && func() bool {
					__antithesis_instrumentation__.Notify(629162)
					return enumMemberIsRemoving(member) == true
				}() == true
			})
			__antithesis_instrumentation__.Notify(629148)

			regionChangeFinalizer, err := newDatabaseRegionChangeFinalizer(
				ctx,
				txn,
				t.execCfg,
				descsCol,
				typeDesc.GetParentID(),
				typeDesc.GetID(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(629163)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629164)
			}
			__antithesis_instrumentation__.Notify(629149)
			defer regionChangeFinalizer.cleanup()

			b := txn.NewBatch()
			if err := descsCol.WriteDescToBatch(
				ctx, true, typeDesc, b,
			); err != nil {
				__antithesis_instrumentation__.Notify(629165)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629166)
			}
			__antithesis_instrumentation__.Notify(629150)

			arrayTypeDesc, err := descsCol.GetMutableTypeVersionByID(ctx, txn, typeDesc.ArrayTypeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(629167)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629168)
			}
			__antithesis_instrumentation__.Notify(629151)
			if err := descsCol.WriteDescToBatch(
				ctx, true, arrayTypeDesc, b,
			); err != nil {
				__antithesis_instrumentation__.Notify(629169)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629170)
			}
			__antithesis_instrumentation__.Notify(629152)

			if err := txn.Run(ctx, b); err != nil {
				__antithesis_instrumentation__.Notify(629171)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629172)
			}
			__antithesis_instrumentation__.Notify(629153)

			if typeDesc.Kind == descpb.TypeDescriptor_MULTIREGION_ENUM {
				__antithesis_instrumentation__.Notify(629173)
				if fn := t.execCfg.TypeSchemaChangerTestingKnobs.RunBeforeMultiRegionUpdates; fn != nil {
					__antithesis_instrumentation__.Notify(629175)
					if err := fn(); err != nil {
						__antithesis_instrumentation__.Notify(629176)
						return err
					} else {
						__antithesis_instrumentation__.Notify(629177)
					}
				} else {
					__antithesis_instrumentation__.Notify(629178)
				}
				__antithesis_instrumentation__.Notify(629174)
				if err := regionChangeFinalizer.finalize(ctx, txn); err != nil {
					__antithesis_instrumentation__.Notify(629179)
					return err
				} else {
					__antithesis_instrumentation__.Notify(629180)
				}
			} else {
				__antithesis_instrumentation__.Notify(629181)
			}
			__antithesis_instrumentation__.Notify(629154)

			return nil
		}
		__antithesis_instrumentation__.Notify(629097)
		if err := DescsTxn(ctx, t.execCfg, run); err != nil {
			__antithesis_instrumentation__.Notify(629182)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629183)
		}
		__antithesis_instrumentation__.Notify(629098)

		if err := refreshTypeDescriptorLeases(ctx, leaseMgr, typeDesc); err != nil {
			__antithesis_instrumentation__.Notify(629184)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629185)
		}
	} else {
		__antithesis_instrumentation__.Notify(629186)
	}
	__antithesis_instrumentation__.Notify(629073)

	if typeDesc.Dropped() {
		__antithesis_instrumentation__.Notify(629187)
		if err := t.execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(629188)
			b := txn.NewBatch()
			b.Del(catalogkeys.MakeDescMetadataKey(codec, typeDesc.GetID()))
			return txn.Run(ctx, b)
		}); err != nil {
			__antithesis_instrumentation__.Notify(629189)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629190)
		}
	} else {
		__antithesis_instrumentation__.Notify(629191)
	}
	__antithesis_instrumentation__.Notify(629074)
	return nil
}

func (t *typeSchemaChanger) isTransitioningInCurrentJob(
	member *descpb.TypeDescriptor_EnumMember,
) bool {
	__antithesis_instrumentation__.Notify(629192)
	for _, rep := range t.transitioningMembers {
		__antithesis_instrumentation__.Notify(629194)
		if bytes.Equal(member.PhysicalRepresentation, rep) {
			__antithesis_instrumentation__.Notify(629195)
			return true
		} else {
			__antithesis_instrumentation__.Notify(629196)
		}
	}
	__antithesis_instrumentation__.Notify(629193)
	return false
}

func applyFilterOnEnumMembers(
	typeDesc *typedesc.Mutable, shouldRemove func(member *descpb.TypeDescriptor_EnumMember) bool,
) {
	__antithesis_instrumentation__.Notify(629197)
	idx := 0
	for _, member := range typeDesc.EnumMembers {
		__antithesis_instrumentation__.Notify(629199)
		if shouldRemove(&member) {
			__antithesis_instrumentation__.Notify(629201)

			continue
		} else {
			__antithesis_instrumentation__.Notify(629202)
		}
		__antithesis_instrumentation__.Notify(629200)
		typeDesc.EnumMembers[idx] = member
		idx++
	}
	__antithesis_instrumentation__.Notify(629198)
	typeDesc.EnumMembers = typeDesc.EnumMembers[:idx]
}

func (t *typeSchemaChanger) cleanupEnumValues(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(629203)
	var regionChangeFinalizer *databaseRegionChangeFinalizer

	cleanup := func(ctx context.Context, txn *kv.Txn, descsCol *descs.Collection) error {
		__antithesis_instrumentation__.Notify(629205)
		typeDesc, err := descsCol.GetMutableTypeVersionByID(ctx, txn, t.typeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(629213)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629214)
		}
		__antithesis_instrumentation__.Notify(629206)

		if !enumHasNonPublic(typeDesc) {
			__antithesis_instrumentation__.Notify(629215)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(629216)
		}
		__antithesis_instrumentation__.Notify(629207)

		if typeDesc.Kind == descpb.TypeDescriptor_MULTIREGION_ENUM {
			__antithesis_instrumentation__.Notify(629217)
			regionChangeFinalizer, err = newDatabaseRegionChangeFinalizer(
				ctx,
				txn,
				t.execCfg,
				descsCol,
				typeDesc.GetParentID(),
				typeDesc.GetID(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(629219)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629220)
			}
			__antithesis_instrumentation__.Notify(629218)
			defer regionChangeFinalizer.cleanup()
		} else {
			__antithesis_instrumentation__.Notify(629221)
		}
		__antithesis_instrumentation__.Notify(629208)

		for i := range typeDesc.EnumMembers {
			__antithesis_instrumentation__.Notify(629222)
			member := &typeDesc.EnumMembers[i]
			if t.isTransitioningInCurrentJob(member) && func() bool {
				__antithesis_instrumentation__.Notify(629223)
				return enumMemberIsRemoving(member) == true
			}() == true {
				__antithesis_instrumentation__.Notify(629224)
				member.Capability = descpb.TypeDescriptor_EnumMember_ALL
				member.Direction = descpb.TypeDescriptor_EnumMember_NONE
			} else {
				__antithesis_instrumentation__.Notify(629225)
			}
		}
		__antithesis_instrumentation__.Notify(629209)

		applyFilterOnEnumMembers(typeDesc, func(member *descpb.TypeDescriptor_EnumMember) bool {
			__antithesis_instrumentation__.Notify(629226)
			return t.isTransitioningInCurrentJob(member) && func() bool {
				__antithesis_instrumentation__.Notify(629227)
				return enumMemberIsAdding(member) == true
			}() == true
		})
		__antithesis_instrumentation__.Notify(629210)

		if err := descsCol.WriteDesc(ctx, true, typeDesc, txn); err != nil {
			__antithesis_instrumentation__.Notify(629228)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629229)
		}
		__antithesis_instrumentation__.Notify(629211)

		if regionChangeFinalizer != nil {
			__antithesis_instrumentation__.Notify(629230)
			if err := regionChangeFinalizer.finalize(ctx, txn); err != nil {
				__antithesis_instrumentation__.Notify(629231)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629232)
			}
		} else {
			__antithesis_instrumentation__.Notify(629233)
		}
		__antithesis_instrumentation__.Notify(629212)

		return nil
	}
	__antithesis_instrumentation__.Notify(629204)
	return DescsTxn(ctx, t.execCfg, cleanup)
}

func convertToSQLStringRepresentation(bytes []byte) (string, error) {
	__antithesis_instrumentation__.Notify(629234)
	var byteRep strings.Builder
	byteRep.WriteString("x'")
	if _, err := hex.NewEncoder(&byteRep).Write(bytes); err != nil {
		__antithesis_instrumentation__.Notify(629236)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(629237)
	}
	__antithesis_instrumentation__.Notify(629235)
	byteRep.WriteString("'")
	return byteRep.String(), nil
}

func doesArrayContainEnumValues(s string, member *descpb.TypeDescriptor_EnumMember) bool {
	__antithesis_instrumentation__.Notify(629238)
	enumValues := strings.Split(s[1:len(s)-1], ",")
	for _, val := range enumValues {
		__antithesis_instrumentation__.Notify(629240)
		if strings.TrimSpace(val) == member.LogicalRepresentation {
			__antithesis_instrumentation__.Notify(629241)
			return true
		} else {
			__antithesis_instrumentation__.Notify(629242)
		}
	}
	__antithesis_instrumentation__.Notify(629239)
	return false
}

func findUsagesOfEnumValue(
	exprStr string, member *descpb.TypeDescriptor_EnumMember, typeID descpb.ID,
) (bool, error) {
	__antithesis_instrumentation__.Notify(629243)
	expr, err := parser.ParseExpr(exprStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(629247)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(629248)
	}
	__antithesis_instrumentation__.Notify(629244)
	var foundUsage bool

	visitFunc := func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(629249)
		switch t := expr.(type) {

		case *tree.AnnotateTypeExpr:
			__antithesis_instrumentation__.Notify(629250)

			typeOid, ok := t.Type.(*tree.OIDTypeReference)
			if !ok {
				__antithesis_instrumentation__.Notify(629263)
				return true, expr, nil
			} else {
				__antithesis_instrumentation__.Notify(629264)
			}
			__antithesis_instrumentation__.Notify(629251)
			id, err := typedesc.UserDefinedTypeOIDToID(typeOid.OID)
			if err != nil {
				__antithesis_instrumentation__.Notify(629265)
				return false, expr, err
			} else {
				__antithesis_instrumentation__.Notify(629266)
			}
			__antithesis_instrumentation__.Notify(629252)
			if id != typeID {
				__antithesis_instrumentation__.Notify(629267)
				return true, expr, nil
			} else {
				__antithesis_instrumentation__.Notify(629268)
			}
			__antithesis_instrumentation__.Notify(629253)

			strVal, ok := t.Expr.(*tree.StrVal)
			if !ok {
				__antithesis_instrumentation__.Notify(629269)
				return true, expr, nil
			} else {
				__antithesis_instrumentation__.Notify(629270)
			}
			__antithesis_instrumentation__.Notify(629254)
			physicalRep := []byte(strVal.RawString())
			if bytes.Equal(physicalRep, member.PhysicalRepresentation) {
				__antithesis_instrumentation__.Notify(629271)
				foundUsage = true
			} else {
				__antithesis_instrumentation__.Notify(629272)
			}
			__antithesis_instrumentation__.Notify(629255)
			return false, expr, nil

		case *tree.CastExpr:
			__antithesis_instrumentation__.Notify(629256)
			typeOid, ok := t.Type.(*tree.OIDTypeReference)
			if !ok {
				__antithesis_instrumentation__.Notify(629273)
				return true, expr, nil
			} else {
				__antithesis_instrumentation__.Notify(629274)
			}
			__antithesis_instrumentation__.Notify(629257)
			id, err := typedesc.UserDefinedTypeOIDToID(typeOid.OID)
			if err != nil {
				__antithesis_instrumentation__.Notify(629275)
				return false, expr, err
			} else {
				__antithesis_instrumentation__.Notify(629276)
			}
			__antithesis_instrumentation__.Notify(629258)

			id = id - 1
			if id != typeID {
				__antithesis_instrumentation__.Notify(629277)
				return true, expr, nil
			} else {
				__antithesis_instrumentation__.Notify(629278)
			}
			__antithesis_instrumentation__.Notify(629259)

			annotateType, ok := t.Expr.(*tree.AnnotateTypeExpr)
			if !ok {
				__antithesis_instrumentation__.Notify(629279)
				return true, expr, nil
			} else {
				__antithesis_instrumentation__.Notify(629280)
			}
			__antithesis_instrumentation__.Notify(629260)
			strVal, ok := annotateType.Expr.(*tree.StrVal)
			if !ok {
				__antithesis_instrumentation__.Notify(629281)
				return true, expr, nil
			} else {
				__antithesis_instrumentation__.Notify(629282)
			}
			__antithesis_instrumentation__.Notify(629261)
			foundUsage = doesArrayContainEnumValues(strVal.RawString(), member)
			return false, expr, nil
		default:
			__antithesis_instrumentation__.Notify(629262)
			return true, expr, nil
		}
	}
	__antithesis_instrumentation__.Notify(629245)

	_, err = tree.SimpleVisit(expr, visitFunc)
	if err != nil {
		__antithesis_instrumentation__.Notify(629283)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(629284)
	}
	__antithesis_instrumentation__.Notify(629246)
	return foundUsage, nil
}

func findUsagesOfEnumValueInViewQuery(
	viewQuery string, member *descpb.TypeDescriptor_EnumMember, typeID descpb.ID,
) (bool, error) {
	__antithesis_instrumentation__.Notify(629285)
	var foundUsage bool
	visitFunc := func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(629289)
		annotateType, ok := expr.(*tree.AnnotateTypeExpr)
		if !ok {
			__antithesis_instrumentation__.Notify(629296)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(629297)
		}
		__antithesis_instrumentation__.Notify(629290)

		typeOid, ok := annotateType.Type.(*tree.OIDTypeReference)
		if !ok {
			__antithesis_instrumentation__.Notify(629298)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(629299)
		}
		__antithesis_instrumentation__.Notify(629291)
		id, err := typedesc.UserDefinedTypeOIDToID(typeOid.OID)
		if err != nil {
			__antithesis_instrumentation__.Notify(629300)
			return false, expr, err
		} else {
			__antithesis_instrumentation__.Notify(629301)
		}
		__antithesis_instrumentation__.Notify(629292)
		if id != typeID {
			__antithesis_instrumentation__.Notify(629302)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(629303)
		}
		__antithesis_instrumentation__.Notify(629293)

		strVal, ok := annotateType.Expr.(*tree.StrVal)
		if !ok {
			__antithesis_instrumentation__.Notify(629304)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(629305)
		}
		__antithesis_instrumentation__.Notify(629294)
		physicalRep := []byte(strVal.RawString())
		if bytes.Equal(physicalRep, member.PhysicalRepresentation) {
			__antithesis_instrumentation__.Notify(629306)
			foundUsage = true
			return false, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(629307)
		}
		__antithesis_instrumentation__.Notify(629295)

		return false, expr, nil
	}
	__antithesis_instrumentation__.Notify(629286)

	stmt, err := parser.ParseOne(viewQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(629308)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(629309)
	}
	__antithesis_instrumentation__.Notify(629287)
	_, err = tree.SimpleStmtVisit(stmt.AST, visitFunc)
	if err != nil {
		__antithesis_instrumentation__.Notify(629310)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(629311)
	}
	__antithesis_instrumentation__.Notify(629288)
	return foundUsage, nil
}

func (t *typeSchemaChanger) canRemoveEnumValue(
	ctx context.Context,
	typeDesc *typedesc.Mutable,
	txn *kv.Txn,
	member *descpb.TypeDescriptor_EnumMember,
	descsCol *descs.Collection,
) error {
	__antithesis_instrumentation__.Notify(629312)
	for _, ID := range typeDesc.ReferencingDescriptorIDs {
		__antithesis_instrumentation__.Notify(629315)
		desc, err := descsCol.GetImmutableTableByID(ctx, txn, ID, tree.ObjectLookupFlags{
			CommonLookupFlags: tree.CommonLookupFlags{
				AvoidLeased: true,
				Required:    true,
			},
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(629322)
			return errors.Wrapf(err,
				"could not validate enum value removal for %q", member.LogicalRepresentation)
		} else {
			__antithesis_instrumentation__.Notify(629323)
		}
		__antithesis_instrumentation__.Notify(629316)
		if desc.IsView() {
			__antithesis_instrumentation__.Notify(629324)
			foundUsage, err := findUsagesOfEnumValueInViewQuery(desc.GetViewQuery(), member, typeDesc.ID)
			if err != nil {
				__antithesis_instrumentation__.Notify(629326)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629327)
			}
			__antithesis_instrumentation__.Notify(629325)
			if foundUsage {
				__antithesis_instrumentation__.Notify(629328)
				return pgerror.Newf(pgcode.DependentObjectsStillExist,
					"could not remove enum value %q as it is being used in view %q",
					member.LogicalRepresentation, desc.GetName())
			} else {
				__antithesis_instrumentation__.Notify(629329)
			}
		} else {
			__antithesis_instrumentation__.Notify(629330)
		}
		__antithesis_instrumentation__.Notify(629317)

		var query strings.Builder
		colSelectors := tabledesc.ColumnsSelectors(desc.PublicColumns())
		columns := tree.AsStringWithFlags(&colSelectors, tree.FmtSerializable)
		query.WriteString(fmt.Sprintf("SELECT %s FROM [%d as t] WHERE", columns, ID))
		firstClause := true
		validationQueryConstructed := false

		for _, idx := range desc.AllIndexes() {
			__antithesis_instrumentation__.Notify(629331)
			if pred := idx.GetPredicate(); pred != "" {
				__antithesis_instrumentation__.Notify(629335)
				foundUsage, err := findUsagesOfEnumValue(pred, member, typeDesc.ID)
				if err != nil {
					__antithesis_instrumentation__.Notify(629337)
					return err
				} else {
					__antithesis_instrumentation__.Notify(629338)
				}
				__antithesis_instrumentation__.Notify(629336)
				if foundUsage {
					__antithesis_instrumentation__.Notify(629339)
					return pgerror.Newf(pgcode.DependentObjectsStillExist,
						"could not remove enum value %q as it is being used in a predicate of index %s",
						member.LogicalRepresentation, &tree.TableIndexName{
							Table: tree.MakeUnqualifiedTableName(tree.Name(desc.GetName())),
							Index: tree.UnrestrictedName(idx.GetName()),
						})
				} else {
					__antithesis_instrumentation__.Notify(629340)
				}
			} else {
				__antithesis_instrumentation__.Notify(629341)
			}
			__antithesis_instrumentation__.Notify(629332)
			keyColumns := make([]catalog.Column, 0, idx.NumKeyColumns())
			for i := 0; i < idx.NumKeyColumns(); i++ {
				__antithesis_instrumentation__.Notify(629342)
				col, err := desc.FindColumnWithID(idx.GetKeyColumnID(i))
				if err != nil {
					__antithesis_instrumentation__.Notify(629344)
					return errors.WithAssertionFailure(err)
				} else {
					__antithesis_instrumentation__.Notify(629345)
				}
				__antithesis_instrumentation__.Notify(629343)
				keyColumns = append(keyColumns, col)
			}
			__antithesis_instrumentation__.Notify(629333)
			foundUsage, err := findUsagesOfEnumValueInPartitioning(
				idx.GetPartitioning(), t.execCfg.Codec, keyColumns, desc, idx, member, nil, typeDesc,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(629346)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629347)
			}
			__antithesis_instrumentation__.Notify(629334)
			if foundUsage {
				__antithesis_instrumentation__.Notify(629348)
				return pgerror.Newf(pgcode.DependentObjectsStillExist,
					"could not remove enum value %q as it is being used in the partitioning of index %s",
					member.LogicalRepresentation, &tree.TableIndexName{
						Table: tree.MakeUnqualifiedTableName(tree.Name(desc.GetName())),
						Index: tree.UnrestrictedName(idx.GetName()),
					})
			} else {
				__antithesis_instrumentation__.Notify(629349)
			}
		}
		__antithesis_instrumentation__.Notify(629318)

		for _, chk := range desc.AllActiveAndInactiveChecks() {
			__antithesis_instrumentation__.Notify(629350)
			foundUsage, err := findUsagesOfEnumValue(chk.Expr, member, typeDesc.ID)
			if err != nil {
				__antithesis_instrumentation__.Notify(629352)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629353)
			}
			__antithesis_instrumentation__.Notify(629351)
			if foundUsage {
				__antithesis_instrumentation__.Notify(629354)
				return pgerror.Newf(pgcode.DependentObjectsStillExist,
					"could not remove enum value %q as it is being used in a check constraint of %q",
					member.LogicalRepresentation, desc.GetName())
			} else {
				__antithesis_instrumentation__.Notify(629355)
			}
		}
		__antithesis_instrumentation__.Notify(629319)

		for _, col := range desc.PublicColumns() {
			__antithesis_instrumentation__.Notify(629356)

			if col.HasDefault() {
				__antithesis_instrumentation__.Notify(629360)
				foundUsage, err := findUsagesOfEnumValue(col.GetDefaultExpr(), member, typeDesc.ID)
				if err != nil {
					__antithesis_instrumentation__.Notify(629362)
					return err
				} else {
					__antithesis_instrumentation__.Notify(629363)
				}
				__antithesis_instrumentation__.Notify(629361)
				if foundUsage {
					__antithesis_instrumentation__.Notify(629364)
					return pgerror.Newf(pgcode.DependentObjectsStillExist,
						"could not remove enum value %q as it is being used in a default expresion of %q",
						member.LogicalRepresentation, desc.GetName())
				} else {
					__antithesis_instrumentation__.Notify(629365)
				}
			} else {
				__antithesis_instrumentation__.Notify(629366)
			}
			__antithesis_instrumentation__.Notify(629357)

			if col.IsComputed() {
				__antithesis_instrumentation__.Notify(629367)
				foundUsage, err := findUsagesOfEnumValue(col.GetComputeExpr(), member, typeDesc.ID)
				if err != nil {
					__antithesis_instrumentation__.Notify(629369)
					return err
				} else {
					__antithesis_instrumentation__.Notify(629370)
				}
				__antithesis_instrumentation__.Notify(629368)
				if foundUsage {
					__antithesis_instrumentation__.Notify(629371)
					return pgerror.Newf(pgcode.DependentObjectsStillExist,
						"could not remove enum value %q as it is being used in a computed column of %q",
						member.LogicalRepresentation, desc.GetName())
				} else {
					__antithesis_instrumentation__.Notify(629372)
				}
			} else {
				__antithesis_instrumentation__.Notify(629373)
			}
			__antithesis_instrumentation__.Notify(629358)

			if col.HasOnUpdate() {
				__antithesis_instrumentation__.Notify(629374)
				foundUsage, err := findUsagesOfEnumValue(col.GetOnUpdateExpr(), member, typeDesc.ID)
				if err != nil {
					__antithesis_instrumentation__.Notify(629376)
					return err
				} else {
					__antithesis_instrumentation__.Notify(629377)
				}
				__antithesis_instrumentation__.Notify(629375)
				if foundUsage {
					__antithesis_instrumentation__.Notify(629378)
					return pgerror.Newf(pgcode.DependentObjectsStillExist,
						"could not remove enum value %q as it is being used in an ON UPDATE expression"+
							" of %q",
						member.LogicalRepresentation, desc.GetName())
				} else {
					__antithesis_instrumentation__.Notify(629379)
				}
			} else {
				__antithesis_instrumentation__.Notify(629380)
			}
			__antithesis_instrumentation__.Notify(629359)

			if col.GetType().UserDefined() {
				__antithesis_instrumentation__.Notify(629381)
				tid, terr := typedesc.GetUserDefinedTypeDescID(col.GetType())
				if terr != nil {
					__antithesis_instrumentation__.Notify(629383)
					return terr
				} else {
					__antithesis_instrumentation__.Notify(629384)
				}
				__antithesis_instrumentation__.Notify(629382)
				if typeDesc.ID == tid {
					__antithesis_instrumentation__.Notify(629385)
					if !firstClause {
						__antithesis_instrumentation__.Notify(629388)
						query.WriteString(" OR")
					} else {
						__antithesis_instrumentation__.Notify(629389)
					}
					__antithesis_instrumentation__.Notify(629386)
					sqlPhysRep, err := convertToSQLStringRepresentation(member.PhysicalRepresentation)
					if err != nil {
						__antithesis_instrumentation__.Notify(629390)
						return err
					} else {
						__antithesis_instrumentation__.Notify(629391)
					}
					__antithesis_instrumentation__.Notify(629387)
					colName := col.ColName()
					query.WriteString(fmt.Sprintf(
						" t.%s = %s",
						colName.String(),
						sqlPhysRep,
					))
					firstClause = false
					validationQueryConstructed = true
				} else {
					__antithesis_instrumentation__.Notify(629392)
				}
			} else {
				__antithesis_instrumentation__.Notify(629393)
			}
		}
		__antithesis_instrumentation__.Notify(629320)
		query.WriteString(" LIMIT 1")

		if validationQueryConstructed {
			__antithesis_instrumentation__.Notify(629394)

			_, dbDesc, err := descsCol.GetImmutableDatabaseByID(
				ctx, txn, typeDesc.ParentID, tree.DatabaseLookupFlags{
					Required:    true,
					AvoidLeased: true,
				})
			const validationErr = "could not validate removal of enum value %q"
			if err != nil {
				__antithesis_instrumentation__.Notify(629397)
				return errors.Wrapf(err, validationErr, member.LogicalRepresentation)
			} else {
				__antithesis_instrumentation__.Notify(629398)
			}
			__antithesis_instrumentation__.Notify(629395)
			override := sessiondata.InternalExecutorOverride{
				User:     security.RootUserName(),
				Database: dbDesc.GetName(),
			}
			rows, err := t.execCfg.InternalExecutor.QueryRowEx(ctx, "count-value-usage", txn, override, query.String())
			if err != nil {
				__antithesis_instrumentation__.Notify(629399)
				return errors.Wrapf(err, validationErr, member.LogicalRepresentation)
			} else {
				__antithesis_instrumentation__.Notify(629400)
			}
			__antithesis_instrumentation__.Notify(629396)

			if len(rows) > 0 {
				__antithesis_instrumentation__.Notify(629401)
				return pgerror.Newf(pgcode.DependentObjectsStillExist,
					"could not remove enum value %q as it is being used by %q in row: %s",
					member.LogicalRepresentation, desc.GetName(), labeledRowValues(desc.PublicColumns(), rows))
			} else {
				__antithesis_instrumentation__.Notify(629402)
			}
		} else {
			__antithesis_instrumentation__.Notify(629403)
		}
		__antithesis_instrumentation__.Notify(629321)

		if typeDesc.Kind == descpb.TypeDescriptor_MULTIREGION_ENUM && func() bool {
			__antithesis_instrumentation__.Notify(629404)
			return desc.IsLocalityRegionalByTable() == true
		}() == true {
			__antithesis_instrumentation__.Notify(629405)
			homedRegion, err := desc.GetRegionalByTableRegion()
			if err != nil {
				__antithesis_instrumentation__.Notify(629407)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629408)
			}
			__antithesis_instrumentation__.Notify(629406)
			if catpb.RegionName(member.LogicalRepresentation) == homedRegion {
				__antithesis_instrumentation__.Notify(629409)
				return errors.Newf("could not remove enum value %q as it is the home region for table %q",
					member.LogicalRepresentation, desc.GetName())
			} else {
				__antithesis_instrumentation__.Notify(629410)
			}
		} else {
			__antithesis_instrumentation__.Notify(629411)
		}
	}
	__antithesis_instrumentation__.Notify(629313)

	arrayTypeDesc, err := descsCol.GetImmutableTypeByID(
		ctx, txn, typeDesc.ArrayTypeID, tree.ObjectLookupFlags{})
	if err != nil {
		__antithesis_instrumentation__.Notify(629412)
		return err
	} else {
		__antithesis_instrumentation__.Notify(629413)
	}
	__antithesis_instrumentation__.Notify(629314)

	return t.canRemoveEnumValueFromArrayUsages(ctx, arrayTypeDesc, member, txn, descsCol)
}

func findUsagesOfEnumValueInPartitioning(
	partitioning catalog.Partitioning,
	codec keys.SQLCodec,
	columns []catalog.Column,
	table catalog.TableDescriptor,
	index catalog.Index,
	member *descpb.TypeDescriptor_EnumMember,
	fakePrefixDatums []tree.Datum,
	typ *typedesc.Mutable,
) (foundUsage bool, _ error) {
	__antithesis_instrumentation__.Notify(629414)
	if partitioning == nil || func() bool {
		__antithesis_instrumentation__.Notify(629421)
		return partitioning.NumColumns() == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(629422)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(629423)
	}
	__antithesis_instrumentation__.Notify(629415)

	var colsToCheck util.FastIntSet
	for i, c := range columns[:partitioning.NumColumns()] {
		__antithesis_instrumentation__.Notify(629424)
		typT := c.GetType()
		if !typT.UserDefined() {
			__antithesis_instrumentation__.Notify(629428)
			continue
		} else {
			__antithesis_instrumentation__.Notify(629429)
		}
		__antithesis_instrumentation__.Notify(629425)
		id, err := typedesc.UserDefinedTypeOIDToID(typT.Oid())
		if err != nil {
			__antithesis_instrumentation__.Notify(629430)
			return false, errors.WithAssertionFailure(err)
		} else {
			__antithesis_instrumentation__.Notify(629431)
		}
		__antithesis_instrumentation__.Notify(629426)
		if id != typ.GetID() {
			__antithesis_instrumentation__.Notify(629432)
			continue
		} else {
			__antithesis_instrumentation__.Notify(629433)
		}
		__antithesis_instrumentation__.Notify(629427)
		colsToCheck.Add(i)
	}
	__antithesis_instrumentation__.Notify(629416)
	makeFakeDatums := func(n int) []tree.Datum {
		__antithesis_instrumentation__.Notify(629434)
		ret := fakePrefixDatums
		for i := 0; i < n; i++ {
			__antithesis_instrumentation__.Notify(629436)
			ret = append(ret, tree.DNull)
		}
		__antithesis_instrumentation__.Notify(629435)
		return ret
	}
	__antithesis_instrumentation__.Notify(629417)

	maybeFindUsageInValues := func(values [][]byte) (err error) {
		__antithesis_instrumentation__.Notify(629437)
		if colsToCheck.Empty() {
			__antithesis_instrumentation__.Notify(629440)
			return
		} else {
			__antithesis_instrumentation__.Notify(629441)
		}
		__antithesis_instrumentation__.Notify(629438)
		for _, v := range values {
			__antithesis_instrumentation__.Notify(629442)
			foundUsage, err = findUsageOfEnumValueInEncodedPartitioningValue(
				codec, table, index, partitioning, v, fakePrefixDatums, colsToCheck, foundUsage, member,
			)
			if foundUsage {
				__antithesis_instrumentation__.Notify(629444)
				err = iterutil.StopIteration()
			} else {
				__antithesis_instrumentation__.Notify(629445)
			}
			__antithesis_instrumentation__.Notify(629443)
			if err != nil {
				__antithesis_instrumentation__.Notify(629446)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629447)
			}
		}
		__antithesis_instrumentation__.Notify(629439)
		return nil
	}
	__antithesis_instrumentation__.Notify(629418)
	if err := partitioning.ForEachList(func(
		name string, values [][]byte, subPartitioning catalog.Partitioning,
	) (err error) {
		__antithesis_instrumentation__.Notify(629448)
		if err = maybeFindUsageInValues(values); err != nil {
			__antithesis_instrumentation__.Notify(629451)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629452)
		}
		__antithesis_instrumentation__.Notify(629449)
		foundUsage, err = findUsagesOfEnumValueInPartitioning(
			subPartitioning, codec, columns[partitioning.NumColumns():],
			table, index, member, makeFakeDatums(partitioning.NumColumns()), typ)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(629453)
			return foundUsage == true
		}() == true {
			__antithesis_instrumentation__.Notify(629454)
			err = iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(629455)
		}
		__antithesis_instrumentation__.Notify(629450)
		return err
	}); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(629456)
		return foundUsage == true
	}() == true {
		__antithesis_instrumentation__.Notify(629457)
		return foundUsage, err
	} else {
		__antithesis_instrumentation__.Notify(629458)
	}
	__antithesis_instrumentation__.Notify(629419)
	if err := partitioning.ForEachRange(func(
		name string, from, to []byte,
	) error {
		__antithesis_instrumentation__.Notify(629459)
		return maybeFindUsageInValues([][]byte{from, to})
	}); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(629460)
		return foundUsage == true
	}() == true {
		__antithesis_instrumentation__.Notify(629461)
		return foundUsage, err
	} else {
		__antithesis_instrumentation__.Notify(629462)
	}
	__antithesis_instrumentation__.Notify(629420)
	return false, nil
}

func findUsageOfEnumValueInEncodedPartitioningValue(
	codec keys.SQLCodec,
	table catalog.TableDescriptor,
	index catalog.Index,
	partitioning catalog.Partitioning,
	v []byte,
	fakePrefixDatums []tree.Datum,
	colsToCheck util.FastIntSet,
	foundUsage bool,
	member *descpb.TypeDescriptor_EnumMember,
) (bool, error) {
	__antithesis_instrumentation__.Notify(629463)
	var d tree.DatumAlloc
	tuple, _, err := rowenc.DecodePartitionTuple(
		&d, codec, table, index, partitioning, v, fakePrefixDatums,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(629466)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(629467)
	}
	__antithesis_instrumentation__.Notify(629464)
	colsToCheck.ForEach(func(i int) {
		__antithesis_instrumentation__.Notify(629468)
		foundUsage = foundUsage || func() bool {
			__antithesis_instrumentation__.Notify(629469)
			return bytes.Equal(
				member.PhysicalRepresentation,
				tuple.Datums[i].(*tree.DEnum).PhysicalRep,
			) == true
		}() == true
	})
	__antithesis_instrumentation__.Notify(629465)
	return foundUsage, nil
}

func (t *typeSchemaChanger) canRemoveEnumValueFromArrayUsages(
	ctx context.Context,
	arrayTypeDesc catalog.TypeDescriptor,
	member *descpb.TypeDescriptor_EnumMember,
	txn *kv.Txn,
	descsCol *descs.Collection,
) error {
	__antithesis_instrumentation__.Notify(629470)
	const validationErr = "could not validate removal of enum value %q"
	for i := 0; i < arrayTypeDesc.NumReferencingDescriptors(); i++ {
		__antithesis_instrumentation__.Notify(629472)
		id := arrayTypeDesc.GetReferencingDescriptorID(i)
		desc, err := descsCol.GetImmutableTableByID(ctx, txn, id, tree.ObjectLookupFlags{})
		if err != nil {
			__antithesis_instrumentation__.Notify(629479)
			return errors.Wrapf(err, validationErr, member.LogicalRepresentation)
		} else {
			__antithesis_instrumentation__.Notify(629480)
		}
		__antithesis_instrumentation__.Notify(629473)
		var unionUnnests strings.Builder
		var query strings.Builder

		firstClause := true
		for _, col := range desc.PublicColumns() {
			__antithesis_instrumentation__.Notify(629481)
			if !col.GetType().UserDefined() {
				__antithesis_instrumentation__.Notify(629484)
				continue
			} else {
				__antithesis_instrumentation__.Notify(629485)
			}
			__antithesis_instrumentation__.Notify(629482)
			tid, terr := typedesc.GetUserDefinedTypeDescID(col.GetType())
			if terr != nil {
				__antithesis_instrumentation__.Notify(629486)
				return terr
			} else {
				__antithesis_instrumentation__.Notify(629487)
			}
			__antithesis_instrumentation__.Notify(629483)
			if arrayTypeDesc.GetID() == tid {
				__antithesis_instrumentation__.Notify(629488)
				if !firstClause {
					__antithesis_instrumentation__.Notify(629490)
					unionUnnests.WriteString(" UNION ")
				} else {
					__antithesis_instrumentation__.Notify(629491)
				}
				__antithesis_instrumentation__.Notify(629489)
				colName := col.ColName()
				unionUnnests.WriteString(fmt.Sprintf(
					"SELECT unnest(t.%s) FROM [%d AS t]",
					colName.String(),
					id,
				))
				firstClause = false
			} else {
				__antithesis_instrumentation__.Notify(629492)
			}
		}
		__antithesis_instrumentation__.Notify(629474)

		if firstClause {
			__antithesis_instrumentation__.Notify(629493)
			continue
		} else {
			__antithesis_instrumentation__.Notify(629494)
		}
		__antithesis_instrumentation__.Notify(629475)
		query.WriteString("SELECT unnest FROM (")
		query.WriteString(unionUnnests.String())

		sqlPhysRep, err := convertToSQLStringRepresentation(member.PhysicalRepresentation)
		if err != nil {
			__antithesis_instrumentation__.Notify(629495)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629496)
		}
		__antithesis_instrumentation__.Notify(629476)
		query.WriteString(fmt.Sprintf(") WHERE unnest = %s", sqlPhysRep))

		_, dbDesc, err := descsCol.GetImmutableDatabaseByID(
			ctx, txn, arrayTypeDesc.GetParentID(), tree.DatabaseLookupFlags{
				Required:    true,
				AvoidLeased: true,
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(629497)
			return errors.Wrapf(err, validationErr, member.LogicalRepresentation)
		} else {
			__antithesis_instrumentation__.Notify(629498)
		}
		__antithesis_instrumentation__.Notify(629477)
		override := sessiondata.InternalExecutorOverride{
			User:     security.RootUserName(),
			Database: dbDesc.GetName(),
		}
		rows, err := t.execCfg.InternalExecutor.QueryRowEx(
			ctx,
			"count-array-type-value-usage",
			txn,
			override,
			query.String(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(629499)
			return errors.Wrapf(err, validationErr, member.LogicalRepresentation)
		} else {
			__antithesis_instrumentation__.Notify(629500)
		}
		__antithesis_instrumentation__.Notify(629478)
		if len(rows) > 0 {
			__antithesis_instrumentation__.Notify(629501)

			parentSchema, err := descsCol.GetImmutableSchemaByID(
				ctx, txn, desc.GetParentSchemaID(), tree.SchemaLookupFlags{Required: true})
			if err != nil {
				__antithesis_instrumentation__.Notify(629503)
				return err
			} else {
				__antithesis_instrumentation__.Notify(629504)
			}
			__antithesis_instrumentation__.Notify(629502)
			fqName := tree.MakeTableNameWithSchema(
				tree.Name(dbDesc.GetName()),
				tree.Name(parentSchema.GetName()),
				tree.Name(desc.GetName()),
			)
			return pgerror.Newf(pgcode.DependentObjectsStillExist, "could not remove enum value %q as it is being used by table %q",
				member.LogicalRepresentation, fqName.FQString(),
			)
		} else {
			__antithesis_instrumentation__.Notify(629505)
		}
	}
	__antithesis_instrumentation__.Notify(629471)

	return nil
}

func enumHasNonPublic(typeDesc catalog.TypeDescriptor) bool {
	__antithesis_instrumentation__.Notify(629506)
	hasNonPublic := false
	for i := 0; i < typeDesc.NumEnumMembers(); i++ {
		__antithesis_instrumentation__.Notify(629508)
		if typeDesc.IsMemberReadOnly(i) {
			__antithesis_instrumentation__.Notify(629509)
			hasNonPublic = true
			break
		} else {
			__antithesis_instrumentation__.Notify(629510)
		}
	}
	__antithesis_instrumentation__.Notify(629507)
	return hasNonPublic
}

func enumMemberIsAdding(member *descpb.TypeDescriptor_EnumMember) bool {
	__antithesis_instrumentation__.Notify(629511)
	if member.Capability == descpb.TypeDescriptor_EnumMember_READ_ONLY && func() bool {
		__antithesis_instrumentation__.Notify(629513)
		return member.Direction == descpb.TypeDescriptor_EnumMember_ADD == true
	}() == true {
		__antithesis_instrumentation__.Notify(629514)
		return true
	} else {
		__antithesis_instrumentation__.Notify(629515)
	}
	__antithesis_instrumentation__.Notify(629512)
	return false
}

func enumMemberIsRemoving(member *descpb.TypeDescriptor_EnumMember) bool {
	__antithesis_instrumentation__.Notify(629516)
	if member.Capability == descpb.TypeDescriptor_EnumMember_READ_ONLY && func() bool {
		__antithesis_instrumentation__.Notify(629518)
		return member.Direction == descpb.TypeDescriptor_EnumMember_REMOVE == true
	}() == true {
		__antithesis_instrumentation__.Notify(629519)
		return true
	} else {
		__antithesis_instrumentation__.Notify(629520)
	}
	__antithesis_instrumentation__.Notify(629517)
	return false
}

func (t *typeSchemaChanger) execWithRetry(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(629521)

	opts := retry.Options{
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     20 * time.Second,
		Multiplier:     1.5,
	}
	for r := retry.StartWithCtx(ctx, opts); r.Next(); {
		__antithesis_instrumentation__.Notify(629523)
		if err := t.execCfg.JobRegistry.CheckPausepoint("typeschemachanger.before.exec"); err != nil {
			__antithesis_instrumentation__.Notify(629525)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629526)
		}
		__antithesis_instrumentation__.Notify(629524)
		tcErr := t.exec(ctx)
		switch {
		case tcErr == nil:
			__antithesis_instrumentation__.Notify(629527)
			return nil
		case errors.Is(tcErr, catalog.ErrDescriptorNotFound):
			__antithesis_instrumentation__.Notify(629528)

			log.Infof(
				ctx,
				"descriptor %d not found for type change job; assuming it was dropped, and exiting",
				t.typeID,
			)
			return nil
		case !IsPermanentSchemaChangeError(tcErr):
			__antithesis_instrumentation__.Notify(629529)

			log.Infof(ctx, "retrying type schema change due to retriable error %v", tcErr)
		default:
			__antithesis_instrumentation__.Notify(629530)
			return tcErr
		}
	}
	__antithesis_instrumentation__.Notify(629522)
	return nil
}

func (t *typeSchemaChanger) logTags() *logtags.Buffer {
	__antithesis_instrumentation__.Notify(629531)
	buf := &logtags.Buffer{}
	buf.Add("typeChangeExec", nil)
	buf.Add("type", t.typeID)
	return buf
}

type typeChangeResumer struct {
	job *jobs.Job
}

func (t *typeChangeResumer) Resume(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(629532)
	p := execCtx.(JobExecContext)
	if p.ExecCfg().TypeSchemaChangerTestingKnobs.TypeSchemaChangeJobNoOp != nil {
		__antithesis_instrumentation__.Notify(629534)
		if p.ExecCfg().TypeSchemaChangerTestingKnobs.TypeSchemaChangeJobNoOp() {
			__antithesis_instrumentation__.Notify(629535)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(629536)
		}
	} else {
		__antithesis_instrumentation__.Notify(629537)
	}
	__antithesis_instrumentation__.Notify(629533)
	tc := &typeSchemaChanger{
		typeID:               t.job.Details().(jobspb.TypeSchemaChangeDetails).TypeID,
		transitioningMembers: t.job.Details().(jobspb.TypeSchemaChangeDetails).TransitioningMembers,
		execCfg:              p.ExecCfg(),
	}
	return tc.execWithRetry(ctx)
}

func (t *typeChangeResumer) OnFailOrCancel(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(629538)

	tc := &typeSchemaChanger{
		typeID:               t.job.Details().(jobspb.TypeSchemaChangeDetails).TypeID,
		transitioningMembers: t.job.Details().(jobspb.TypeSchemaChangeDetails).TransitioningMembers,
		execCfg:              execCtx.(JobExecContext).ExecCfg(),
	}

	if rollbackErr := func() error {
		__antithesis_instrumentation__.Notify(629540)
		if err := tc.cleanupEnumValues(ctx); err != nil {
			__antithesis_instrumentation__.Notify(629544)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629545)
		}
		__antithesis_instrumentation__.Notify(629541)

		if err := drainNamesForDescriptor(
			ctx, tc.typeID, tc.execCfg.CollectionFactory, tc.execCfg.DB,
			tc.execCfg.InternalExecutor, tc.execCfg.Codec,
			nil,
		); err != nil {
			__antithesis_instrumentation__.Notify(629546)
			return err
		} else {
			__antithesis_instrumentation__.Notify(629547)
		}
		__antithesis_instrumentation__.Notify(629542)

		if fn := tc.execCfg.TypeSchemaChangerTestingKnobs.RunAfterOnFailOrCancel; fn != nil {
			__antithesis_instrumentation__.Notify(629548)
			return fn()
		} else {
			__antithesis_instrumentation__.Notify(629549)
		}
		__antithesis_instrumentation__.Notify(629543)

		return nil
	}(); rollbackErr != nil {
		__antithesis_instrumentation__.Notify(629550)
		switch {
		case errors.Is(rollbackErr, catalog.ErrDescriptorNotFound):
			__antithesis_instrumentation__.Notify(629551)

			log.Infof(
				ctx,
				"descriptor %d not found for type change job; assuming it was dropped, and exiting",
				tc.typeID,
			)
		case !IsPermanentSchemaChangeError(rollbackErr):
			__antithesis_instrumentation__.Notify(629552)
			return jobs.MarkAsRetryJobError(rollbackErr)
		default:
			__antithesis_instrumentation__.Notify(629553)
			return rollbackErr
		}
	} else {
		__antithesis_instrumentation__.Notify(629554)
	}
	__antithesis_instrumentation__.Notify(629539)

	return nil
}

func init() {
	createResumerFn := func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
		return &typeChangeResumer{job: job}
	}
	jobs.RegisterConstructor(jobspb.TypeTypeSchemaChange, createResumerFn)
}
