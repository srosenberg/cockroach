package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/hex"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

func (p *planner) UnsafeUpsertDescriptor(
	ctx context.Context, descID int64, encodedDesc []byte, force bool,
) error {
	__antithesis_instrumentation__.Notify(566132)
	const method = "crdb_internal.unsafe_upsert_descriptor()"
	ev := eventpb.UnsafeUpsertDescriptor{Force: force}
	if err := checkPlannerStateForRepairFunctions(ctx, p, method); err != nil {
		__antithesis_instrumentation__.Notify(566152)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566153)
	}
	__antithesis_instrumentation__.Notify(566133)

	id := descpb.ID(descID)
	var desc descpb.Descriptor
	if err := protoutil.Unmarshal(encodedDesc, &desc); err != nil {
		__antithesis_instrumentation__.Notify(566154)
		return pgerror.Wrapf(err, pgcode.InvalidObjectDefinition, "failed to decode descriptor")
	} else {
		__antithesis_instrumentation__.Notify(566155)
	}
	__antithesis_instrumentation__.Notify(566134)
	newID, newVersion, _, _, newModTime, err := descpb.GetDescriptorMetadata(&desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(566156)
		return pgerror.Wrapf(err, pgcode.InvalidObjectDefinition, "invalid descriptor")
	} else {
		__antithesis_instrumentation__.Notify(566157)
	}
	__antithesis_instrumentation__.Notify(566135)
	if newID != id {
		__antithesis_instrumentation__.Notify(566158)
		if !force {
			__antithesis_instrumentation__.Notify(566160)
			return pgerror.Newf(pgcode.InvalidObjectDefinition, "invalid descriptor ID %d, expected %d", newID, id)
		} else {
			__antithesis_instrumentation__.Notify(566161)
		}
		__antithesis_instrumentation__.Notify(566159)
		newID = id
	} else {
		__antithesis_instrumentation__.Notify(566162)
	}
	__antithesis_instrumentation__.Notify(566136)

	mut, notice, err := unsafeReadDescriptor(ctx, p, id, force)
	if err != nil {
		__antithesis_instrumentation__.Notify(566163)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566164)
	}
	__antithesis_instrumentation__.Notify(566137)
	if notice != nil {
		__antithesis_instrumentation__.Notify(566165)
		ev.ForceNotice = notice.Error()
	} else {
		__antithesis_instrumentation__.Notify(566166)
	}
	__antithesis_instrumentation__.Notify(566138)

	var existingProto *descpb.Descriptor
	var existingVersion descpb.DescriptorVersion
	var existingModTime hlc.Timestamp
	var previousOwner string
	var previousUserPrivileges []catpb.UserPrivileges
	if mut != nil {
		__antithesis_instrumentation__.Notify(566167)
		if mut.IsUncommittedVersion() {
			__antithesis_instrumentation__.Notify(566169)
			return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
				"cannot modify a modified descriptor (%d) with UnsafeUpsertDescriptor", id)
		} else {
			__antithesis_instrumentation__.Notify(566170)
		}
		__antithesis_instrumentation__.Notify(566168)
		existingProto = protoutil.Clone(mut.DescriptorProto()).(*descpb.Descriptor)
		existingVersion = mut.GetVersion()
		existingModTime = mut.GetModificationTime()
		previousOwner = mut.GetPrivileges().Owner().Normalized()
		previousUserPrivileges = mut.GetPrivileges().Users
	} else {
		__antithesis_instrumentation__.Notify(566171)
	}
	__antithesis_instrumentation__.Notify(566139)

	if newVersion != existingVersion+1 {
		__antithesis_instrumentation__.Notify(566172)
		if !force {
			__antithesis_instrumentation__.Notify(566174)
			return pgerror.Newf(pgcode.InvalidObjectDefinition, "invalid new descriptor version %d, expected %v",
				newVersion, existingVersion+1)
		} else {
			__antithesis_instrumentation__.Notify(566175)
		}
		__antithesis_instrumentation__.Notify(566173)
		newVersion = existingVersion + 1
	} else {
		__antithesis_instrumentation__.Notify(566176)
	}
	__antithesis_instrumentation__.Notify(566140)
	if newModTime.IsEmpty() {
		__antithesis_instrumentation__.Notify(566177)
		if newVersion > 1 && func() bool {
			__antithesis_instrumentation__.Notify(566179)
			return existingModTime.IsEmpty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(566180)
			return pgerror.Newf(pgcode.InvalidObjectDefinition, "missing modification time in updated descriptor with version %d",
				newVersion)
		} else {
			__antithesis_instrumentation__.Notify(566181)
		}
		__antithesis_instrumentation__.Notify(566178)

		newModTime = existingModTime
	} else {
		__antithesis_instrumentation__.Notify(566182)
	}
	__antithesis_instrumentation__.Notify(566141)

	objectType := privilege.Any
	{
		__antithesis_instrumentation__.Notify(566183)

		if tbl := desc.GetTable(); tbl != nil {
			__antithesis_instrumentation__.Notify(566187)
			tbl.ID = newID
			tbl.Version = newVersion
			objectType = privilege.Table
		} else {
			__antithesis_instrumentation__.Notify(566188)
		}
		__antithesis_instrumentation__.Notify(566184)

		if db := desc.GetDatabase(); db != nil {
			__antithesis_instrumentation__.Notify(566189)
			db.ID = newID
			db.Version = newVersion
			objectType = privilege.Database
		} else {
			__antithesis_instrumentation__.Notify(566190)
		}
		__antithesis_instrumentation__.Notify(566185)

		if typ := desc.GetType(); typ != nil {
			__antithesis_instrumentation__.Notify(566191)
			typ.ID = newID
			typ.Version = newVersion
			objectType = privilege.Type
		} else {
			__antithesis_instrumentation__.Notify(566192)
		}
		__antithesis_instrumentation__.Notify(566186)

		if sc := desc.GetSchema(); sc != nil {
			__antithesis_instrumentation__.Notify(566193)
			sc.ID = newID
			sc.Version = newVersion
			objectType = privilege.Schema
		} else {
			__antithesis_instrumentation__.Notify(566194)
		}
	}
	__antithesis_instrumentation__.Notify(566142)
	if objectType == privilege.Any {
		__antithesis_instrumentation__.Notify(566195)
		return pgerror.Newf(pgcode.InvalidObjectDefinition, "invalid new descriptor %+v", desc)
	} else {
		__antithesis_instrumentation__.Notify(566196)
	}
	__antithesis_instrumentation__.Notify(566143)

	tbl, db, typ, schema := descpb.FromDescriptorWithMVCCTimestamp(&desc, newModTime)
	switch md := mut.(type) {
	case *tabledesc.Mutable:
		__antithesis_instrumentation__.Notify(566197)
		if objectType != privilege.Table {
			__antithesis_instrumentation__.Notify(566208)
			return pgerror.Newf(pgcode.InvalidObjectDefinition, "cannot replace table descriptor with %s", objectType)
		} else {
			__antithesis_instrumentation__.Notify(566209)
		}
		__antithesis_instrumentation__.Notify(566198)
		md.TableDescriptor = *tbl
	case *schemadesc.Mutable:
		__antithesis_instrumentation__.Notify(566199)
		if objectType != privilege.Schema {
			__antithesis_instrumentation__.Notify(566210)
			return pgerror.Newf(pgcode.InvalidObjectDefinition, "cannot replace schema descriptor with %s", objectType)
		} else {
			__antithesis_instrumentation__.Notify(566211)
		}
		__antithesis_instrumentation__.Notify(566200)
		md.SchemaDescriptor = *schema
	case *dbdesc.Mutable:
		__antithesis_instrumentation__.Notify(566201)
		if objectType != privilege.Database {
			__antithesis_instrumentation__.Notify(566212)
			return pgerror.Newf(pgcode.InvalidObjectDefinition, "cannot replace database descriptor with %s", objectType)
		} else {
			__antithesis_instrumentation__.Notify(566213)
		}
		__antithesis_instrumentation__.Notify(566202)
		md.DatabaseDescriptor = *db
	case *typedesc.Mutable:
		__antithesis_instrumentation__.Notify(566203)
		if objectType != privilege.Type {
			__antithesis_instrumentation__.Notify(566214)
			return pgerror.Newf(pgcode.InvalidObjectDefinition, "cannot replace type descriptor with %s", objectType)
		} else {
			__antithesis_instrumentation__.Notify(566215)
		}
		__antithesis_instrumentation__.Notify(566204)
		md.TypeDescriptor = *typ
	case nil:
		__antithesis_instrumentation__.Notify(566205)
		b := descbuilder.NewBuilderWithMVCCTimestamp(&desc, newModTime)
		if b == nil {
			__antithesis_instrumentation__.Notify(566216)
			return pgerror.Newf(pgcode.InvalidObjectDefinition, "invalid new descriptor %+v", desc)
		} else {
			__antithesis_instrumentation__.Notify(566217)
		}
		__antithesis_instrumentation__.Notify(566206)
		mut = b.BuildCreatedMutable()
	default:
		__antithesis_instrumentation__.Notify(566207)
		return errors.AssertionFailedf("unknown descriptor type %T for id %d", mut, id)
	}
	__antithesis_instrumentation__.Notify(566144)

	if existingProto != nil {
		__antithesis_instrumentation__.Notify(566218)
		marshaled, err := protoutil.Marshal(existingProto)
		if err != nil {
			__antithesis_instrumentation__.Notify(566220)
			return errors.NewAssertionErrorWithWrappedErrf(err, "failed to marshal existing descriptor %+v", existingProto)
		} else {
			__antithesis_instrumentation__.Notify(566221)
		}
		__antithesis_instrumentation__.Notify(566219)
		ev.PreviousDescriptor = hex.EncodeToString(marshaled)
	} else {
		__antithesis_instrumentation__.Notify(566222)
	}

	{
		__antithesis_instrumentation__.Notify(566223)
		marshaled, err := protoutil.Marshal(mut.DescriptorProto())
		if err != nil {
			__antithesis_instrumentation__.Notify(566225)
			return errors.NewAssertionErrorWithWrappedErrf(err, "failed to marshal new descriptor %+v", mut.DescriptorProto())
		} else {
			__antithesis_instrumentation__.Notify(566226)
		}
		__antithesis_instrumentation__.Notify(566224)
		ev.NewDescriptor = hex.EncodeToString(marshaled)
	}
	__antithesis_instrumentation__.Notify(566145)

	maxDescIDKeyVal, err := p.extendedEvalCtx.DB.Get(context.Background(), p.extendedEvalCtx.Codec.DescIDSequenceKey())
	if err != nil {
		__antithesis_instrumentation__.Notify(566227)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566228)
	}
	__antithesis_instrumentation__.Notify(566146)
	maxDescID, err := maxDescIDKeyVal.Value.GetInt()
	if err != nil {
		__antithesis_instrumentation__.Notify(566229)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566230)
	}
	__antithesis_instrumentation__.Notify(566147)
	if maxDescID <= descID {
		__antithesis_instrumentation__.Notify(566231)
		if !force {
			__antithesis_instrumentation__.Notify(566233)
			return pgerror.Newf(pgcode.InvalidObjectDefinition,
				"descriptor ID %d must be less than the descriptor ID sequence value %d", descID, maxDescID)
		} else {
			__antithesis_instrumentation__.Notify(566234)
		}
		__antithesis_instrumentation__.Notify(566232)
		inc := descID - maxDescID + 1
		_, err = kv.IncrementValRetryable(ctx, p.extendedEvalCtx.DB, p.extendedEvalCtx.Codec.DescIDSequenceKey(), inc)
		if err != nil {
			__antithesis_instrumentation__.Notify(566235)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566236)
		}
	} else {
		__antithesis_instrumentation__.Notify(566237)
	}
	__antithesis_instrumentation__.Notify(566148)

	if force {
		__antithesis_instrumentation__.Notify(566238)
		p.Descriptors().SkipValidationOnWrite()
	} else {
		__antithesis_instrumentation__.Notify(566239)
	}

	{
		__antithesis_instrumentation__.Notify(566240)
		b := p.txn.NewBatch()
		if err := p.Descriptors().WriteDescToBatch(
			ctx, p.extendedEvalCtx.Tracing.KVTracingEnabled(), mut, b,
		); err != nil {
			__antithesis_instrumentation__.Notify(566242)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566243)
		}
		__antithesis_instrumentation__.Notify(566241)
		if err := p.txn.Run(ctx, b); err != nil {
			__antithesis_instrumentation__.Notify(566244)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566245)
		}
	}
	__antithesis_instrumentation__.Notify(566149)

	newOwner := mut.GetPrivileges().Owner().Normalized()
	if previousOwner != newOwner {
		__antithesis_instrumentation__.Notify(566246)
		if err := logOwnerEvents(ctx, p, newOwner, mut); err != nil {
			__antithesis_instrumentation__.Notify(566247)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566248)
		}
	} else {
		__antithesis_instrumentation__.Notify(566249)
	}
	__antithesis_instrumentation__.Notify(566150)

	if err := comparePrivileges(
		ctx, p, mut, previousUserPrivileges, objectType,
	); err != nil {
		__antithesis_instrumentation__.Notify(566250)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566251)
	}
	__antithesis_instrumentation__.Notify(566151)

	return p.logEvent(ctx, id, &ev)
}

func comparePrivileges(
	ctx context.Context,
	p *planner,
	existing catalog.MutableDescriptor,
	prevUserPrivileges []catpb.UserPrivileges,
	objectType privilege.ObjectType,
) error {
	__antithesis_instrumentation__.Notify(566252)
	computePrivilegeChanges := func(prev, cur *catpb.UserPrivileges) (granted, revoked []string) {
		__antithesis_instrumentation__.Notify(566257)

		if cur == nil {
			__antithesis_instrumentation__.Notify(566263)
			revoked = privilege.ListFromBitField(prev.Privileges, objectType).SortedNames()
			return nil, revoked
		} else {
			__antithesis_instrumentation__.Notify(566264)
		}
		__antithesis_instrumentation__.Notify(566258)

		if prev.Privileges == cur.Privileges {
			__antithesis_instrumentation__.Notify(566265)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(566266)
		}
		__antithesis_instrumentation__.Notify(566259)

		prevPrivilegeSet := make(map[string]struct{})
		for _, priv := range privilege.ListFromBitField(prev.Privileges, objectType).SortedNames() {
			__antithesis_instrumentation__.Notify(566267)
			prevPrivilegeSet[priv] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(566260)

		for _, priv := range privilege.ListFromBitField(cur.Privileges, objectType).SortedNames() {
			__antithesis_instrumentation__.Notify(566268)
			if _, ok := prevPrivilegeSet[priv]; !ok {
				__antithesis_instrumentation__.Notify(566269)

				granted = append(granted, priv)
			} else {
				__antithesis_instrumentation__.Notify(566270)

				delete(prevPrivilegeSet, priv)
			}
		}
		__antithesis_instrumentation__.Notify(566261)

		for priv := range prevPrivilegeSet {
			__antithesis_instrumentation__.Notify(566271)
			revoked = append(revoked, priv)
		}
		__antithesis_instrumentation__.Notify(566262)
		sort.Strings(revoked)

		return granted, revoked
	}
	__antithesis_instrumentation__.Notify(566253)

	curUserPrivileges := existing.GetPrivileges().Users
	curUserMap := make(map[string]*catpb.UserPrivileges)
	for i := range curUserPrivileges {
		__antithesis_instrumentation__.Notify(566272)
		curUser := &curUserPrivileges[i]
		curUserMap[curUser.User().Normalized()] = curUser
	}
	__antithesis_instrumentation__.Notify(566254)

	for i := range prevUserPrivileges {
		__antithesis_instrumentation__.Notify(566273)
		prev := &prevUserPrivileges[i]
		username := prev.User().Normalized()
		cur := curUserMap[username]
		granted, revoked := computePrivilegeChanges(prev, cur)
		delete(curUserMap, username)
		if granted == nil && func() bool {
			__antithesis_instrumentation__.Notify(566275)
			return revoked == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(566276)
			continue
		} else {
			__antithesis_instrumentation__.Notify(566277)
		}
		__antithesis_instrumentation__.Notify(566274)

		if err := logPrivilegeEvents(
			ctx, p, existing, granted, revoked, username,
		); err != nil {
			__antithesis_instrumentation__.Notify(566278)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566279)
		}
	}
	__antithesis_instrumentation__.Notify(566255)

	for i := range curUserPrivileges {
		__antithesis_instrumentation__.Notify(566280)
		username := curUserPrivileges[i].User().Normalized()
		if _, ok := curUserMap[username]; ok {
			__antithesis_instrumentation__.Notify(566281)
			granted := privilege.ListFromBitField(curUserPrivileges[i].Privileges, objectType).SortedNames()
			if granted == nil {
				__antithesis_instrumentation__.Notify(566283)
				continue
			} else {
				__antithesis_instrumentation__.Notify(566284)
			}
			__antithesis_instrumentation__.Notify(566282)
			if err := logPrivilegeEvents(
				ctx, p, existing, granted, nil, username,
			); err != nil {
				__antithesis_instrumentation__.Notify(566285)
				return err
			} else {
				__antithesis_instrumentation__.Notify(566286)
			}
		} else {
			__antithesis_instrumentation__.Notify(566287)
		}
	}
	__antithesis_instrumentation__.Notify(566256)

	return nil
}

func logPrivilegeEvents(
	ctx context.Context,
	p *planner,
	existing catalog.MutableDescriptor,
	grantedPrivileges []string,
	revokedPrivileges []string,
	grantee string,
) error {
	__antithesis_instrumentation__.Notify(566288)

	eventDetails := eventpb.CommonSQLPrivilegeEventDetails{
		Grantee:           grantee,
		GrantedPrivileges: grantedPrivileges,
		RevokedPrivileges: revokedPrivileges,
	}

	switch md := existing.(type) {
	case *tabledesc.Mutable:
		__antithesis_instrumentation__.Notify(566290)
		return p.logEvent(ctx, existing.GetID(), &eventpb.ChangeTablePrivilege{
			CommonSQLPrivilegeEventDetails: eventDetails,
			TableName:                      md.GetName(),
		})
	case *schemadesc.Mutable:
		__antithesis_instrumentation__.Notify(566291)
		return p.logEvent(ctx, existing.GetID(), &eventpb.ChangeSchemaPrivilege{
			CommonSQLPrivilegeEventDetails: eventDetails,
			SchemaName:                     md.GetName(),
		})
	case *dbdesc.Mutable:
		__antithesis_instrumentation__.Notify(566292)
		return p.logEvent(ctx, existing.GetID(), &eventpb.ChangeDatabasePrivilege{
			CommonSQLPrivilegeEventDetails: eventDetails,
			DatabaseName:                   md.GetName(),
		})
	case *typedesc.Mutable:
		__antithesis_instrumentation__.Notify(566293)
		return p.logEvent(ctx, existing.GetID(), &eventpb.ChangeTypePrivilege{
			CommonSQLPrivilegeEventDetails: eventDetails,
			TypeName:                       md.GetName(),
		})
	}
	__antithesis_instrumentation__.Notify(566289)
	return nil
}

func logOwnerEvents(
	ctx context.Context, p *planner, newOwner string, existing catalog.MutableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(566294)
	switch md := existing.(type) {
	case *tabledesc.Mutable:
		__antithesis_instrumentation__.Notify(566296)
		return p.logEvent(ctx, md.GetID(), &eventpb.AlterTableOwner{
			TableName: md.GetName(),
			Owner:     newOwner,
		})
	case *schemadesc.Mutable:
		__antithesis_instrumentation__.Notify(566297)
		return p.logEvent(ctx, md.GetID(), &eventpb.AlterSchemaOwner{
			SchemaName: md.GetName(),
			Owner:      newOwner,
		})
	case *dbdesc.Mutable:
		__antithesis_instrumentation__.Notify(566298)
		return p.logEvent(ctx, md.GetID(), &eventpb.AlterDatabaseOwner{
			DatabaseName: md.GetName(),
			Owner:        newOwner,
		})
	case *typedesc.Mutable:
		__antithesis_instrumentation__.Notify(566299)
		return p.logEvent(ctx, md.GetID(), &eventpb.AlterTypeOwner{
			TypeName: md.GetName(),
			Owner:    newOwner,
		})
	}
	__antithesis_instrumentation__.Notify(566295)
	return nil
}

func (p *planner) UnsafeUpsertNamespaceEntry(
	ctx context.Context,
	parentIDInt, parentSchemaIDInt int64,
	name string,
	descIDInt int64,
	force bool,
) error {
	__antithesis_instrumentation__.Notify(566300)
	const method = "crdb_internal.unsafe_upsert_namespace_entry()"
	if err := checkPlannerStateForRepairFunctions(ctx, p, method); err != nil {
		__antithesis_instrumentation__.Notify(566310)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566311)
	}
	__antithesis_instrumentation__.Notify(566301)
	parentID, parentSchemaID, descID := descpb.ID(parentIDInt), descpb.ID(parentSchemaIDInt), descpb.ID(descIDInt)
	key := catalogkeys.MakeObjectNameKey(p.execCfg.Codec, parentID, parentSchemaID, name)
	val, err := p.txn.Get(ctx, key)
	if err != nil {
		__antithesis_instrumentation__.Notify(566312)
		return errors.Wrapf(err, "failed to read namespace entry (%d, %d, %s)",
			parentID, parentSchemaID, name)
	} else {
		__antithesis_instrumentation__.Notify(566313)
	}
	__antithesis_instrumentation__.Notify(566302)

	var existingID descpb.ID
	if val.Value != nil {
		__antithesis_instrumentation__.Notify(566314)
		existingID = descpb.ID(val.ValueInt())
	} else {
		__antithesis_instrumentation__.Notify(566315)
	}
	__antithesis_instrumentation__.Notify(566303)
	flags := p.CommonLookupFlags(true)
	flags.IncludeDropped = true
	flags.IncludeOffline = true
	validateDescriptor := func() error {
		__antithesis_instrumentation__.Notify(566316)
		desc, err := p.Descriptors().GetImmutableDescriptorByID(ctx, p.Txn(), descID, flags)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(566320)
			return descID != keys.PublicSchemaID == true
		}() == true {
			__antithesis_instrumentation__.Notify(566321)
			return errors.Wrapf(err, "failed to retrieve descriptor %d", descID)
		} else {
			__antithesis_instrumentation__.Notify(566322)
		}
		__antithesis_instrumentation__.Notify(566317)
		invalid := false
		switch desc.(type) {
		case nil:
			__antithesis_instrumentation__.Notify(566323)
			return nil
		case catalog.TableDescriptor, catalog.TypeDescriptor:
			__antithesis_instrumentation__.Notify(566324)
			invalid = parentID == descpb.InvalidID || func() bool {
				__antithesis_instrumentation__.Notify(566329)
				return parentSchemaID == descpb.InvalidID == true
			}() == true
		case catalog.SchemaDescriptor:
			__antithesis_instrumentation__.Notify(566325)
			invalid = parentID == descpb.InvalidID || func() bool {
				__antithesis_instrumentation__.Notify(566330)
				return parentSchemaID != descpb.InvalidID == true
			}() == true
		case catalog.DatabaseDescriptor:
			__antithesis_instrumentation__.Notify(566326)
			invalid = parentID != descpb.InvalidID || func() bool {
				__antithesis_instrumentation__.Notify(566331)
				return parentSchemaID != descpb.InvalidID == true
			}() == true
		default:
			__antithesis_instrumentation__.Notify(566327)

			if descID == keys.PublicSchemaID {
				__antithesis_instrumentation__.Notify(566332)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(566333)
			}
			__antithesis_instrumentation__.Notify(566328)
			return errors.AssertionFailedf(
				"unexpected descriptor type %T for descriptor %d", desc, descID)
		}
		__antithesis_instrumentation__.Notify(566318)

		if invalid {
			__antithesis_instrumentation__.Notify(566334)
			return pgerror.Newf(pgcode.InvalidCatalogName,
				"invalid prefix (%d, %d) for %s %d",
				parentID, parentSchemaID, desc.DescriptorType(), descID)
		} else {
			__antithesis_instrumentation__.Notify(566335)
		}
		__antithesis_instrumentation__.Notify(566319)
		return nil
	}
	__antithesis_instrumentation__.Notify(566304)
	validateParentDescriptor := func() error {
		__antithesis_instrumentation__.Notify(566336)
		if parentID == descpb.InvalidID {
			__antithesis_instrumentation__.Notify(566340)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(566341)
		}
		__antithesis_instrumentation__.Notify(566337)
		parent, err := p.Descriptors().GetImmutableDescriptorByID(ctx, p.Txn(), parentID, flags)
		if err != nil {
			__antithesis_instrumentation__.Notify(566342)
			return errors.Wrapf(err, "failed to look up parent %d", parentID)
		} else {
			__antithesis_instrumentation__.Notify(566343)
		}
		__antithesis_instrumentation__.Notify(566338)
		if _, isDatabase := parent.(catalog.DatabaseDescriptor); !isDatabase {
			__antithesis_instrumentation__.Notify(566344)
			return pgerror.Newf(pgcode.InvalidCatalogName,
				"parentID %d is a %T, not a database", parentID, parent)
		} else {
			__antithesis_instrumentation__.Notify(566345)
		}
		__antithesis_instrumentation__.Notify(566339)
		return nil
	}
	__antithesis_instrumentation__.Notify(566305)
	validateParentSchemaDescriptor := func() error {
		__antithesis_instrumentation__.Notify(566346)
		if parentSchemaID == descpb.InvalidID || func() bool {
			__antithesis_instrumentation__.Notify(566350)
			return parentSchemaID == keys.PublicSchemaID == true
		}() == true {
			__antithesis_instrumentation__.Notify(566351)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(566352)
		}
		__antithesis_instrumentation__.Notify(566347)
		schema, err := p.Descriptors().GetImmutableDescriptorByID(ctx, p.Txn(), parentSchemaID, flags)
		if err != nil {
			__antithesis_instrumentation__.Notify(566353)
			return err
		} else {
			__antithesis_instrumentation__.Notify(566354)
		}
		__antithesis_instrumentation__.Notify(566348)
		if _, isSchema := schema.(catalog.SchemaDescriptor); !isSchema {
			__antithesis_instrumentation__.Notify(566355)
			return pgerror.Newf(pgcode.InvalidCatalogName,
				"parentSchemaID %d is a %T, not a schema", parentSchemaID, schema)
		} else {
			__antithesis_instrumentation__.Notify(566356)
		}
		__antithesis_instrumentation__.Notify(566349)
		return nil
	}
	__antithesis_instrumentation__.Notify(566306)

	var validationErr error
	for _, f := range []func() error{
		validateDescriptor,
		validateParentDescriptor,
		validateParentSchemaDescriptor,
	} {
		__antithesis_instrumentation__.Notify(566357)
		if err := f(); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(566358)
			return force == true
		}() == true {
			__antithesis_instrumentation__.Notify(566359)
			validationErr = errors.CombineErrors(validationErr, err)
		} else {
			__antithesis_instrumentation__.Notify(566360)
			if err != nil {
				__antithesis_instrumentation__.Notify(566361)
				return err
			} else {
				__antithesis_instrumentation__.Notify(566362)
			}
		}
	}
	__antithesis_instrumentation__.Notify(566307)
	if err := p.txn.Put(ctx, key, descID); err != nil {
		__antithesis_instrumentation__.Notify(566363)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566364)
	}
	__antithesis_instrumentation__.Notify(566308)
	var validationErrStr string
	if validationErr != nil {
		__antithesis_instrumentation__.Notify(566365)
		validationErrStr = validationErr.Error()
	} else {
		__antithesis_instrumentation__.Notify(566366)
	}
	__antithesis_instrumentation__.Notify(566309)
	return p.logEvent(ctx, descID,
		&eventpb.UnsafeUpsertNamespaceEntry{
			ParentID:         uint32(parentID),
			ParentSchemaID:   uint32(parentSchemaID),
			Name:             name,
			PreviousID:       uint32(existingID),
			Force:            force,
			FailedValidation: validationErr != nil,
			ValidationErrors: validationErrStr,
		})
}

func (p *planner) UnsafeDeleteNamespaceEntry(
	ctx context.Context,
	parentIDInt, parentSchemaIDInt int64,
	name string,
	descIDInt int64,
	force bool,
) error {
	__antithesis_instrumentation__.Notify(566367)
	const method = "crdb_internal.unsafe_delete_namespace_entry()"
	if err := checkPlannerStateForRepairFunctions(ctx, p, method); err != nil {
		__antithesis_instrumentation__.Notify(566376)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566377)
	}
	__antithesis_instrumentation__.Notify(566368)
	parentID, parentSchemaID, descID := descpb.ID(parentIDInt), descpb.ID(parentSchemaIDInt), descpb.ID(descIDInt)
	key := catalogkeys.MakeObjectNameKey(p.execCfg.Codec, parentID, parentSchemaID, name)
	val, err := p.txn.Get(ctx, key)
	if err != nil {
		__antithesis_instrumentation__.Notify(566378)
		return errors.Wrapf(err, "failed to read namespace entry (%d, %d, %s)",
			parentID, parentSchemaID, name)
	} else {
		__antithesis_instrumentation__.Notify(566379)
	}
	__antithesis_instrumentation__.Notify(566369)
	if val.Value == nil {
		__antithesis_instrumentation__.Notify(566380)

		return pgerror.Newf(pgcode.InvalidCatalogName,
			"no namespace entry exists for (%d, %d, %s)",
			parentID, parentSchemaID, name)
	} else {
		__antithesis_instrumentation__.Notify(566381)
	}
	__antithesis_instrumentation__.Notify(566370)
	if val.Value != nil {
		__antithesis_instrumentation__.Notify(566382)
		existingID := descpb.ID(val.ValueInt())
		if existingID != descID {
			__antithesis_instrumentation__.Notify(566383)
			return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
				"namespace entry for (%d, %d, %s) has id %d, not %d",
				parentID, parentSchemaID, name, existingID, descID)
		} else {
			__antithesis_instrumentation__.Notify(566384)
		}
	} else {
		__antithesis_instrumentation__.Notify(566385)
	}
	__antithesis_instrumentation__.Notify(566371)
	desc, notice, err := unsafeReadDescriptor(ctx, p, descID, force)
	if err != nil {
		__antithesis_instrumentation__.Notify(566386)
		return errors.Wrapf(err, "failed to retrieve descriptor %d", descID)
	} else {
		__antithesis_instrumentation__.Notify(566387)
	}
	__antithesis_instrumentation__.Notify(566372)
	if desc != nil && func() bool {
		__antithesis_instrumentation__.Notify(566388)
		return !desc.Dropped() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(566389)
		return !force == true
	}() == true {
		__antithesis_instrumentation__.Notify(566390)
		return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
			"refusing to delete namespace entry for non-dropped descriptor")
	} else {
		__antithesis_instrumentation__.Notify(566391)
	}
	__antithesis_instrumentation__.Notify(566373)
	if err := p.txn.Del(ctx, key); err != nil {
		__antithesis_instrumentation__.Notify(566392)
		return errors.Wrap(err, "failed to delete entry")
	} else {
		__antithesis_instrumentation__.Notify(566393)
	}
	__antithesis_instrumentation__.Notify(566374)

	ev := eventpb.UnsafeDeleteNamespaceEntry{
		ParentID:       uint32(parentID),
		ParentSchemaID: uint32(parentSchemaID),
		Name:           name,
		Force:          force,
	}
	if notice != nil {
		__antithesis_instrumentation__.Notify(566394)
		ev.ForceNotice = notice.Error()
	} else {
		__antithesis_instrumentation__.Notify(566395)
	}
	__antithesis_instrumentation__.Notify(566375)
	return p.logEvent(ctx, descID, &ev)
}

func (p *planner) UnsafeDeleteDescriptor(ctx context.Context, descID int64, force bool) error {
	__antithesis_instrumentation__.Notify(566396)
	const method = "crdb_internal.unsafe_delete_descriptor()"
	if err := checkPlannerStateForRepairFunctions(ctx, p, method); err != nil {
		__antithesis_instrumentation__.Notify(566403)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566404)
	}
	__antithesis_instrumentation__.Notify(566397)
	id := descpb.ID(descID)
	mut, notice, err := unsafeReadDescriptor(ctx, p, id, force)
	if err != nil {
		__antithesis_instrumentation__.Notify(566405)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566406)
	}
	__antithesis_instrumentation__.Notify(566398)

	if mut != nil {
		__antithesis_instrumentation__.Notify(566407)
		mut.MaybeIncrementVersion()
		mut.SetDropped()
		if err := p.Descriptors().AddUncommittedDescriptor(mut); err != nil {
			__antithesis_instrumentation__.Notify(566409)
			return errors.WithAssertionFailure(err)
		} else {
			__antithesis_instrumentation__.Notify(566410)
		}
		__antithesis_instrumentation__.Notify(566408)
		if force {
			__antithesis_instrumentation__.Notify(566411)
			p.Descriptors().SkipValidationOnWrite()
		} else {
			__antithesis_instrumentation__.Notify(566412)
		}
	} else {
		__antithesis_instrumentation__.Notify(566413)
	}
	__antithesis_instrumentation__.Notify(566399)
	descKey := catalogkeys.MakeDescMetadataKey(p.execCfg.Codec, id)
	if err := p.txn.Del(ctx, descKey); err != nil {
		__antithesis_instrumentation__.Notify(566414)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566415)
	}
	__antithesis_instrumentation__.Notify(566400)

	ev := eventpb.UnsafeDeleteDescriptor{
		Force: force,
	}
	if mut != nil {
		__antithesis_instrumentation__.Notify(566416)
		ev.ParentID = uint32(mut.GetParentID())
		ev.ParentSchemaID = uint32(mut.GetParentSchemaID())
		ev.Name = mut.GetName()
	} else {
		__antithesis_instrumentation__.Notify(566417)
	}
	__antithesis_instrumentation__.Notify(566401)
	if notice != nil {
		__antithesis_instrumentation__.Notify(566418)
		ev.ForceNotice = notice.Error()
	} else {
		__antithesis_instrumentation__.Notify(566419)
	}
	__antithesis_instrumentation__.Notify(566402)
	return p.logEvent(ctx, id, &ev)
}

func unsafeReadDescriptor(
	ctx context.Context, p *planner, id descpb.ID, force bool,
) (mut catalog.MutableDescriptor, notice error, err error) {
	__antithesis_instrumentation__.Notify(566420)
	mut, err = p.Descriptors().GetMutableDescriptorByID(ctx, p.txn, id)
	if mut != nil {
		__antithesis_instrumentation__.Notify(566427)
		return mut, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(566428)
	}
	__antithesis_instrumentation__.Notify(566421)
	if errors.Is(err, catalog.ErrDescriptorNotFound) {
		__antithesis_instrumentation__.Notify(566429)
		return nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(566430)
	}
	__antithesis_instrumentation__.Notify(566422)
	if !force {
		__antithesis_instrumentation__.Notify(566431)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(566432)
	}
	__antithesis_instrumentation__.Notify(566423)
	notice = pgnotice.NewWithSeverityf("WARNING",
		"failed to retrieve existing descriptor, continuing with force flag: %v", err)
	p.BufferClientNotice(ctx, notice)

	descKey := catalogkeys.MakeDescMetadataKey(p.execCfg.Codec, id)
	descRow, err := p.txn.Get(ctx, descKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(566433)
		return nil, notice, err
	} else {
		__antithesis_instrumentation__.Notify(566434)
	}
	__antithesis_instrumentation__.Notify(566424)
	var descProto descpb.Descriptor
	if err := descRow.ValueProto(&descProto); err != nil {
		__antithesis_instrumentation__.Notify(566435)
		return nil, notice, err
	} else {
		__antithesis_instrumentation__.Notify(566436)
	}
	__antithesis_instrumentation__.Notify(566425)
	if b := descbuilder.NewBuilderWithMVCCTimestamp(&descProto, descRow.Value.Timestamp); b != nil {
		__antithesis_instrumentation__.Notify(566437)
		mut = b.BuildExistingMutable()
	} else {
		__antithesis_instrumentation__.Notify(566438)
	}
	__antithesis_instrumentation__.Notify(566426)
	return mut, notice, nil
}

func checkPlannerStateForRepairFunctions(ctx context.Context, p *planner, method string) error {
	__antithesis_instrumentation__.Notify(566439)
	if p.extendedEvalCtx.TxnReadOnly {
		__antithesis_instrumentation__.Notify(566443)
		return readOnlyError(method)
	} else {
		__antithesis_instrumentation__.Notify(566444)
	}
	__antithesis_instrumentation__.Notify(566440)
	hasAdmin, err := p.UserHasAdminRole(ctx, p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(566445)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566446)
	}
	__antithesis_instrumentation__.Notify(566441)
	if !hasAdmin {
		__antithesis_instrumentation__.Notify(566447)
		return pgerror.Newf(pgcode.InsufficientPrivilege, "admin role required for %s", method)
	} else {
		__antithesis_instrumentation__.Notify(566448)
	}
	__antithesis_instrumentation__.Notify(566442)
	return nil
}

func (p *planner) ForceDeleteTableData(ctx context.Context, descID int64) error {
	__antithesis_instrumentation__.Notify(566449)
	const method = "crdb_internal.force_delete_table_data()"
	err := checkPlannerStateForRepairFunctions(ctx, p, method)
	if err != nil {
		__antithesis_instrumentation__.Notify(566456)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566457)
	}
	__antithesis_instrumentation__.Notify(566450)

	id := descpb.ID(descID)
	desc, err := p.Descriptors().GetImmutableTableByID(ctx, p.txn, id,
		tree.ObjectLookupFlags{
			CommonLookupFlags: tree.CommonLookupFlags{
				Required:    true,
				AvoidLeased: true,
			},
			DesiredTableDescKind: tree.ResolveRequireTableDesc,
		})
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(566458)
		return pgerror.GetPGCode(err) != pgcode.UndefinedTable == true
	}() == true {
		__antithesis_instrumentation__.Notify(566459)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566460)
	}
	__antithesis_instrumentation__.Notify(566451)
	if desc != nil {
		__antithesis_instrumentation__.Notify(566461)
		return errors.New("descriptor still exists force deletion is blocked")
	} else {
		__antithesis_instrumentation__.Notify(566462)
	}
	__antithesis_instrumentation__.Notify(566452)

	maxDescID, err := p.extendedEvalCtx.DB.Get(context.Background(), p.extendedEvalCtx.Codec.DescIDSequenceKey())
	if err != nil {
		__antithesis_instrumentation__.Notify(566463)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566464)
	}
	__antithesis_instrumentation__.Notify(566453)
	if maxDescID.ValueInt() <= descID {
		__antithesis_instrumentation__.Notify(566465)
		return errors.Newf("descriptor id was never used (descID: %d exceeds maxDescID: %d)",
			descID, maxDescID)
	} else {
		__antithesis_instrumentation__.Notify(566466)
	}
	__antithesis_instrumentation__.Notify(566454)

	prefix := p.extendedEvalCtx.Codec.TablePrefix(uint32(id))
	tableSpans := roachpb.Span{Key: prefix, EndKey: prefix.PrefixEnd()}
	b := &kv.Batch{}
	b.AddRawRequest(&roachpb.ClearRangeRequest{
		RequestHeader: roachpb.RequestHeader{
			Key:    tableSpans.Key,
			EndKey: tableSpans.EndKey,
		},
	})

	err = p.txn.DB().Run(ctx, b)
	if err != nil {
		__antithesis_instrumentation__.Notify(566467)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566468)
	}
	__antithesis_instrumentation__.Notify(566455)

	return p.logEvent(ctx, id,
		&eventpb.ForceDeleteTableDataEntry{
			DescriptorID: uint32(descID),
		})
}

func (p *planner) ExternalReadFile(ctx context.Context, uri string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(566469)
	if err := p.RequireAdminRole(ctx, "network I/O"); err != nil {
		__antithesis_instrumentation__.Notify(566473)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566474)
	}
	__antithesis_instrumentation__.Notify(566470)

	conn, err := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, uri, p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(566475)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566476)
	}
	__antithesis_instrumentation__.Notify(566471)

	file, err := conn.ReadFile(ctx, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(566477)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(566478)
	}
	__antithesis_instrumentation__.Notify(566472)
	return ioctx.ReadAll(ctx, file)
}

func (p *planner) ExternalWriteFile(ctx context.Context, uri string, content []byte) error {
	__antithesis_instrumentation__.Notify(566479)
	if err := p.RequireAdminRole(ctx, "network I/O"); err != nil {
		__antithesis_instrumentation__.Notify(566482)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566483)
	}
	__antithesis_instrumentation__.Notify(566480)

	conn, err := p.ExecCfg().DistSQLSrv.ExternalStorageFromURI(ctx, uri, p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(566484)
		return err
	} else {
		__antithesis_instrumentation__.Notify(566485)
	}
	__antithesis_instrumentation__.Notify(566481)
	return cloud.WriteFile(ctx, conn, "", bytes.NewReader(content))
}
