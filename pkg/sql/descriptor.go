package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

var (
	errEmptyDatabaseName = pgerror.New(pgcode.Syntax, "empty database name")
	errNoDatabase        = pgerror.New(pgcode.InvalidName, "no database specified")
	errNoSchema          = pgerror.Newf(pgcode.InvalidName, "no schema specified")
	errNoTable           = pgerror.New(pgcode.InvalidName, "no table specified")
	errNoType            = pgerror.New(pgcode.InvalidName, "no type specified")
	errNoMatch           = pgerror.New(pgcode.UndefinedObject, "no object matched")
)

func (p *planner) createDatabase(
	ctx context.Context, database *tree.CreateDatabase, jobDesc string,
) (*dbdesc.Mutable, bool, error) {
	__antithesis_instrumentation__.Notify(466038)

	dbName := string(database.Name)
	dKey := catalogkeys.MakeDatabaseNameKey(p.ExecCfg().Codec, dbName)

	if dbID, err := p.Descriptors().Direct().LookupDatabaseID(ctx, p.txn, dbName); err == nil && func() bool {
		__antithesis_instrumentation__.Notify(466048)
		return dbID != descpb.InvalidID == true
	}() == true {
		__antithesis_instrumentation__.Notify(466049)
		if database.IfNotExists {
			__antithesis_instrumentation__.Notify(466051)

			desc, err := p.Descriptors().Direct().MustGetDatabaseDescByID(ctx, p.txn, dbID)
			if err != nil {
				__antithesis_instrumentation__.Notify(466054)
				return nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(466055)
			}
			__antithesis_instrumentation__.Notify(466052)
			if desc.Dropped() {
				__antithesis_instrumentation__.Notify(466056)
				return nil, false, pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
					"database %q is being dropped, try again later",
					dbName)
			} else {
				__antithesis_instrumentation__.Notify(466057)
			}
			__antithesis_instrumentation__.Notify(466053)

			return nil, false, nil
		} else {
			__antithesis_instrumentation__.Notify(466058)
		}
		__antithesis_instrumentation__.Notify(466050)
		return nil, false, sqlerrors.NewDatabaseAlreadyExistsError(dbName)
	} else {
		__antithesis_instrumentation__.Notify(466059)
		if err != nil {
			__antithesis_instrumentation__.Notify(466060)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(466061)
		}
	}
	__antithesis_instrumentation__.Notify(466039)

	id, err := descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(466062)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(466063)
	}
	__antithesis_instrumentation__.Notify(466040)

	if database.PrimaryRegion != tree.PrimaryRegionNotSpecifiedName {
		__antithesis_instrumentation__.Notify(466064)
		telemetry.Inc(sqltelemetry.CreateMultiRegionDatabaseCounter)
		telemetry.Inc(
			sqltelemetry.CreateDatabaseSurvivalGoalCounter(
				database.SurvivalGoal.TelemetryName(),
			),
		)
		if database.Placement != tree.DataPlacementUnspecified {
			__antithesis_instrumentation__.Notify(466065)
			telemetry.Inc(
				sqltelemetry.CreateDatabasePlacementCounter(
					database.Placement.TelemetryName(),
				),
			)
		} else {
			__antithesis_instrumentation__.Notify(466066)
		}
	} else {
		__antithesis_instrumentation__.Notify(466067)
	}
	__antithesis_instrumentation__.Notify(466041)

	regionConfig, err := p.maybeInitializeMultiRegionMetadata(
		ctx,
		database.SurvivalGoal,
		database.PrimaryRegion,
		database.Regions,
		database.Placement,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(466068)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(466069)
	}
	__antithesis_instrumentation__.Notify(466042)

	publicSchemaID, err := p.createPublicSchema(ctx, id, database)
	if err != nil {
		__antithesis_instrumentation__.Notify(466070)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(466071)
	}
	__antithesis_instrumentation__.Notify(466043)

	owner := p.SessionData().User()
	if !database.Owner.Undefined() {
		__antithesis_instrumentation__.Notify(466072)
		owner, err = database.Owner.ToSQLUsername(p.SessionData(), security.UsernameValidation)
		if err != nil {
			__antithesis_instrumentation__.Notify(466073)
			return nil, true, err
		} else {
			__antithesis_instrumentation__.Notify(466074)
		}
	} else {
		__antithesis_instrumentation__.Notify(466075)
	}
	__antithesis_instrumentation__.Notify(466044)

	desc := dbdesc.NewInitial(
		id,
		string(database.Name),
		owner,
		dbdesc.MaybeWithDatabaseRegionConfig(regionConfig),
		dbdesc.WithPublicSchemaID(publicSchemaID),
	)

	if err := p.checkCanAlterToNewOwner(ctx, desc, owner); err != nil {
		__antithesis_instrumentation__.Notify(466076)
		return nil, true, err
	} else {
		__antithesis_instrumentation__.Notify(466077)
	}
	__antithesis_instrumentation__.Notify(466045)

	if err := p.createDescriptorWithID(ctx, dKey, id, desc, jobDesc); err != nil {
		__antithesis_instrumentation__.Notify(466078)
		return nil, true, err
	} else {
		__antithesis_instrumentation__.Notify(466079)
	}
	__antithesis_instrumentation__.Notify(466046)

	if err := p.maybeInitializeMultiRegionDatabase(ctx, desc, regionConfig); err != nil {
		__antithesis_instrumentation__.Notify(466080)
		return nil, true, err
	} else {
		__antithesis_instrumentation__.Notify(466081)
	}
	__antithesis_instrumentation__.Notify(466047)

	return desc, true, nil
}

func (p *planner) maybeCreatePublicSchemaWithDescriptor(
	ctx context.Context, dbID descpb.ID, database *tree.CreateDatabase,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(466082)
	if !p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) {
		__antithesis_instrumentation__.Notify(466086)
		return descpb.InvalidID, nil
	} else {
		__antithesis_instrumentation__.Notify(466087)
	}
	__antithesis_instrumentation__.Notify(466083)

	publicSchemaID, err := descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(466088)
		return descpb.InvalidID, err
	} else {
		__antithesis_instrumentation__.Notify(466089)
	}
	__antithesis_instrumentation__.Notify(466084)

	publicSchemaPrivileges := catpb.NewPublicSchemaPrivilegeDescriptor()
	publicSchemaDesc := schemadesc.NewBuilder(&descpb.SchemaDescriptor{
		ParentID:   dbID,
		Name:       tree.PublicSchema,
		ID:         publicSchemaID,
		Privileges: publicSchemaPrivileges,
		Version:    1,
	}).BuildCreatedMutableSchema()

	if err := p.createDescriptorWithID(
		ctx,
		catalogkeys.MakeSchemaNameKey(p.ExecCfg().Codec, dbID, tree.PublicSchema),
		publicSchemaDesc.GetID(),
		publicSchemaDesc,
		tree.AsStringWithFQNames(database, p.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(466090)
		return descpb.InvalidID, err
	} else {
		__antithesis_instrumentation__.Notify(466091)
	}
	__antithesis_instrumentation__.Notify(466085)

	return publicSchemaID, nil
}

func (p *planner) createPublicSchema(
	ctx context.Context, dbID descpb.ID, database *tree.CreateDatabase,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(466092)
	publicSchemaID, err := p.maybeCreatePublicSchemaWithDescriptor(ctx, dbID, database)
	if err != nil {
		__antithesis_instrumentation__.Notify(466096)
		return descpb.InvalidID, err
	} else {
		__antithesis_instrumentation__.Notify(466097)
	}
	__antithesis_instrumentation__.Notify(466093)
	if publicSchemaID != descpb.InvalidID {
		__antithesis_instrumentation__.Notify(466098)
		return publicSchemaID, nil
	} else {
		__antithesis_instrumentation__.Notify(466099)
	}
	__antithesis_instrumentation__.Notify(466094)

	key := catalogkeys.MakeSchemaNameKey(p.ExecCfg().Codec, dbID, tree.PublicSchema)
	if err := p.CreateSchemaNamespaceEntry(ctx, key, keys.PublicSchemaID); err != nil {
		__antithesis_instrumentation__.Notify(466100)
		return keys.PublicSchemaID, err
	} else {
		__antithesis_instrumentation__.Notify(466101)
	}
	__antithesis_instrumentation__.Notify(466095)
	return keys.PublicSchemaID, nil
}

func (p *planner) createDescriptorWithID(
	ctx context.Context,
	idKey roachpb.Key,
	id descpb.ID,
	descriptor catalog.Descriptor,
	jobDesc string,
) error {
	__antithesis_instrumentation__.Notify(466102)
	if descriptor.GetID() == 0 {
		__antithesis_instrumentation__.Notify(466112)

		log.Fatalf(ctx, "%v", errors.AssertionFailedf("cannot create descriptor with an empty ID: %v", descriptor))
	} else {
		__antithesis_instrumentation__.Notify(466113)
	}
	__antithesis_instrumentation__.Notify(466103)
	if descriptor.GetID() != id {
		__antithesis_instrumentation__.Notify(466114)
		log.Fatalf(ctx, "%v", errors.AssertionFailedf("cannot create descriptor with an unexpected (%v) ID: %v", id, descriptor))
	} else {
		__antithesis_instrumentation__.Notify(466115)
	}
	__antithesis_instrumentation__.Notify(466104)

	b := &kv.Batch{}
	descID := descriptor.GetID()
	if p.ExtendedEvalContext().Tracing.KVTracingEnabled() {
		__antithesis_instrumentation__.Notify(466116)
		log.VEventf(ctx, 2, "CPut %s -> %d", idKey, descID)
	} else {
		__antithesis_instrumentation__.Notify(466117)
	}
	__antithesis_instrumentation__.Notify(466105)
	b.CPut(idKey, descID, nil)
	if err := p.Descriptors().Direct().WriteNewDescToBatch(
		ctx,
		p.ExtendedEvalContext().Tracing.KVTracingEnabled(),
		b,
		descriptor,
	); err != nil {
		__antithesis_instrumentation__.Notify(466118)
		return err
	} else {
		__antithesis_instrumentation__.Notify(466119)
	}
	__antithesis_instrumentation__.Notify(466106)

	mutDesc, ok := descriptor.(catalog.MutableDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(466120)
		log.Fatalf(ctx, "unexpected type %T when creating descriptor", descriptor)
	} else {
		__antithesis_instrumentation__.Notify(466121)
	}
	__antithesis_instrumentation__.Notify(466107)

	isTable := false
	addUncommitted := false
	switch mutDesc.(type) {
	case *dbdesc.Mutable, *schemadesc.Mutable, *typedesc.Mutable:
		__antithesis_instrumentation__.Notify(466122)
		addUncommitted = true
	case *tabledesc.Mutable:
		__antithesis_instrumentation__.Notify(466123)
		addUncommitted = true
		isTable = true
	default:
		__antithesis_instrumentation__.Notify(466124)
		log.Fatalf(ctx, "unexpected type %T when creating descriptor", mutDesc)
	}
	__antithesis_instrumentation__.Notify(466108)
	if addUncommitted {
		__antithesis_instrumentation__.Notify(466125)
		if err := p.Descriptors().AddUncommittedDescriptor(mutDesc); err != nil {
			__antithesis_instrumentation__.Notify(466126)
			return err
		} else {
			__antithesis_instrumentation__.Notify(466127)
		}
	} else {
		__antithesis_instrumentation__.Notify(466128)
	}
	__antithesis_instrumentation__.Notify(466109)

	if err := p.txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(466129)
		return err
	} else {
		__antithesis_instrumentation__.Notify(466130)
	}
	__antithesis_instrumentation__.Notify(466110)
	if isTable && func() bool {
		__antithesis_instrumentation__.Notify(466131)
		return mutDesc.Adding() == true
	}() == true {
		__antithesis_instrumentation__.Notify(466132)

		if err := p.createOrUpdateSchemaChangeJob(
			ctx,
			mutDesc.(*tabledesc.Mutable),
			jobDesc,
			descpb.InvalidMutationID); err != nil {
			__antithesis_instrumentation__.Notify(466133)
			return err
		} else {
			__antithesis_instrumentation__.Notify(466134)
		}
	} else {
		__antithesis_instrumentation__.Notify(466135)
	}
	__antithesis_instrumentation__.Notify(466111)
	return nil
}

func TranslateSurvivalGoal(g tree.SurvivalGoal) (descpb.SurvivalGoal, error) {
	__antithesis_instrumentation__.Notify(466136)
	switch g {
	case tree.SurvivalGoalDefault:
		__antithesis_instrumentation__.Notify(466137)
		return descpb.SurvivalGoal_ZONE_FAILURE, nil
	case tree.SurvivalGoalZoneFailure:
		__antithesis_instrumentation__.Notify(466138)
		return descpb.SurvivalGoal_ZONE_FAILURE, nil
	case tree.SurvivalGoalRegionFailure:
		__antithesis_instrumentation__.Notify(466139)
		return descpb.SurvivalGoal_REGION_FAILURE, nil
	default:
		__antithesis_instrumentation__.Notify(466140)
		return 0, errors.Newf("unknown survival goal: %d", g)
	}
}

func TranslateDataPlacement(g tree.DataPlacement) (descpb.DataPlacement, error) {
	__antithesis_instrumentation__.Notify(466141)
	switch g {
	case tree.DataPlacementUnspecified:
		__antithesis_instrumentation__.Notify(466142)
		return descpb.DataPlacement_DEFAULT, nil
	case tree.DataPlacementDefault:
		__antithesis_instrumentation__.Notify(466143)
		return descpb.DataPlacement_DEFAULT, nil
	case tree.DataPlacementRestricted:
		__antithesis_instrumentation__.Notify(466144)
		return descpb.DataPlacement_RESTRICTED, nil
	default:
		__antithesis_instrumentation__.Notify(466145)
		return 0, errors.AssertionFailedf("unknown data placement: %d", g)
	}
}

func (p *planner) checkRegionIsCurrentlyActive(ctx context.Context, region catpb.RegionName) error {
	__antithesis_instrumentation__.Notify(466146)
	liveRegions, err := p.getLiveClusterRegions(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(466148)
		return err
	} else {
		__antithesis_instrumentation__.Notify(466149)
	}
	__antithesis_instrumentation__.Notify(466147)

	return CheckClusterRegionIsLive(liveRegions, region)
}

var InitializeMultiRegionMetadataCCL = func(
	ctx context.Context,
	execCfg *ExecutorConfig,
	liveClusterRegions LiveClusterRegions,
	survivalGoal tree.SurvivalGoal,
	primaryRegion catpb.RegionName,
	regions []tree.Name,
	dataPlacement tree.DataPlacement,
) (*multiregion.RegionConfig, error) {
	__antithesis_instrumentation__.Notify(466150)
	return nil, sqlerrors.NewCCLRequiredError(
		errors.New("creating multi-region databases requires a CCL binary"),
	)
}

const DefaultPrimaryRegionClusterSettingName = "sql.defaults.primary_region"

var DefaultPrimaryRegion = settings.RegisterStringSetting(
	settings.TenantWritable,
	DefaultPrimaryRegionClusterSettingName,
	`if not empty, all databases created without a PRIMARY REGION will `+
		`implicitly have the given PRIMARY REGION`,
	"",
).WithPublic()

const SecondaryTenantsMultiRegionAbstractionsEnabledSettingName = "sql.multi_region.allow_abstractions_for_secondary_tenants.enabled"

var SecondaryTenantsMultiRegionAbstractionsEnabled = settings.RegisterBoolSetting(
	settings.TenantReadOnly,
	SecondaryTenantsMultiRegionAbstractionsEnabledSettingName,
	"allow secondary tenants to use multi-region abstractions",
	false,
)

func (p *planner) maybeInitializeMultiRegionMetadata(
	ctx context.Context,
	survivalGoal tree.SurvivalGoal,
	primaryRegion tree.Name,
	regions []tree.Name,
	placement tree.DataPlacement,
) (*multiregion.RegionConfig, error) {
	__antithesis_instrumentation__.Notify(466151)
	if !p.execCfg.Codec.ForSystemTenant() && func() bool {
		__antithesis_instrumentation__.Notify(466156)
		return !SecondaryTenantsMultiRegionAbstractionsEnabled.Get(&p.execCfg.Settings.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(466157)

		if primaryRegion == "" && func() bool {
			__antithesis_instrumentation__.Notify(466159)
			return len(regions) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(466160)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(466161)
		}
		__antithesis_instrumentation__.Notify(466158)

		return nil, errors.WithHint(pgerror.Newf(
			pgcode.InvalidDatabaseDefinition,
			"setting %s disallows use of multi-region abstractions",
			SecondaryTenantsMultiRegionAbstractionsEnabledSettingName,
		),
			"consider omitting the primary region")
	} else {
		__antithesis_instrumentation__.Notify(466162)
	}
	__antithesis_instrumentation__.Notify(466152)

	if primaryRegion == "" && func() bool {
		__antithesis_instrumentation__.Notify(466163)
		return len(regions) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(466164)
		defaultPrimaryRegion := DefaultPrimaryRegion.Get(&p.execCfg.Settings.SV)
		if defaultPrimaryRegion == "" {
			__antithesis_instrumentation__.Notify(466166)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(466167)
		}
		__antithesis_instrumentation__.Notify(466165)
		primaryRegion = tree.Name(defaultPrimaryRegion)

		p.BufferClientNotice(
			ctx,
			pgnotice.Newf("setting %s as the PRIMARY REGION as no PRIMARY REGION was specified", primaryRegion),
		)
	} else {
		__antithesis_instrumentation__.Notify(466168)
	}
	__antithesis_instrumentation__.Notify(466153)

	liveRegions, err := p.getLiveClusterRegions(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(466169)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(466170)
	}
	__antithesis_instrumentation__.Notify(466154)

	regionConfig, err := InitializeMultiRegionMetadataCCL(
		ctx,
		p.ExecCfg(),
		liveRegions,
		survivalGoal,
		catpb.RegionName(primaryRegion),
		regions,
		placement,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(466171)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(466172)
	}
	__antithesis_instrumentation__.Notify(466155)

	return regionConfig, nil
}

func (p *planner) GetImmutableTableInterfaceByID(ctx context.Context, id int) (interface{}, error) {
	__antithesis_instrumentation__.Notify(466173)
	desc, err := p.Descriptors().GetImmutableTableByID(
		ctx,
		p.txn,
		descpb.ID(id),
		tree.ObjectLookupFlagsWithRequired(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(466175)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(466176)
	}
	__antithesis_instrumentation__.Notify(466174)
	return desc, nil
}
