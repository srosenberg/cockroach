package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	clustersettings "github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
)

type createSequenceNode struct {
	n      *tree.CreateSequence
	dbDesc catalog.DatabaseDescriptor
}

func (p *planner) CreateSequence(ctx context.Context, n *tree.CreateSequence) (planNode, error) {
	__antithesis_instrumentation__.Notify(463462)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"CREATE SEQUENCE",
	); err != nil {
		__antithesis_instrumentation__.Notify(463466)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463467)
	}
	__antithesis_instrumentation__.Notify(463463)

	un := n.Name.ToUnresolvedObjectName()
	dbDesc, _, prefix, err := p.ResolveTargetObject(ctx, un)
	if err != nil {
		__antithesis_instrumentation__.Notify(463468)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463469)
	}
	__antithesis_instrumentation__.Notify(463464)
	n.Name.ObjectNamePrefix = prefix

	if err := p.CheckPrivilege(ctx, dbDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(463470)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463471)
	}
	__antithesis_instrumentation__.Notify(463465)

	return &createSequenceNode{
		n:      n,
		dbDesc: dbDesc,
	}, nil
}

func (n *createSequenceNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(463472) }

func (n *createSequenceNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(463473)
	telemetry.Inc(sqltelemetry.SchemaChangeCreateCounter("sequence"))

	schemaDesc, err := getSchemaForCreateTable(params, n.dbDesc, n.n.Persistence, &n.n.Name,
		tree.ResolveRequireSequenceDesc, n.n.IfNotExists)
	if err != nil {
		__antithesis_instrumentation__.Notify(463475)
		if sqlerrors.IsRelationAlreadyExistsError(err) && func() bool {
			__antithesis_instrumentation__.Notify(463477)
			return n.n.IfNotExists == true
		}() == true {
			__antithesis_instrumentation__.Notify(463478)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(463479)
		}
		__antithesis_instrumentation__.Notify(463476)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463480)
	}
	__antithesis_instrumentation__.Notify(463474)

	_, err = doCreateSequence(
		params.ctx, params.p, params.SessionData(), n.dbDesc, schemaDesc, &n.n.Name, n.n.Persistence, n.n.Options,
		tree.AsStringWithFQNames(n.n, params.Ann()),
	)

	return err
}

func doCreateSequence(
	ctx context.Context,
	p *planner,
	sessionData *sessiondata.SessionData,
	dbDesc catalog.DatabaseDescriptor,
	scDesc catalog.SchemaDescriptor,
	name *tree.TableName,
	persistence tree.Persistence,
	opts tree.SequenceOptions,
	jobDesc string,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(463481)
	id, err := descidgen.GenerateUniqueDescID(ctx, p.ExecCfg().DB, p.ExecCfg().Codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(463489)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463490)
	}
	__antithesis_instrumentation__.Notify(463482)

	privs := catprivilege.CreatePrivilegesFromDefaultPrivileges(
		dbDesc.GetDefaultPrivilegeDescriptor(),
		scDesc.GetDefaultPrivilegeDescriptor(),
		dbDesc.GetID(),
		sessionData.User(),
		tree.Sequences,
		dbDesc.GetPrivileges(),
	)

	if persistence.IsTemporary() {
		__antithesis_instrumentation__.Notify(463491)
		telemetry.Inc(sqltelemetry.CreateTempSequenceCounter)
	} else {
		__antithesis_instrumentation__.Notify(463492)
	}
	__antithesis_instrumentation__.Notify(463483)

	var creationTime hlc.Timestamp
	desc, err := NewSequenceTableDesc(
		ctx,
		p,
		p.EvalContext().Settings,
		name.Object(),
		opts,
		dbDesc.GetID(),
		scDesc.GetID(),
		id,
		creationTime,
		privs,
		persistence,
		dbDesc.IsMultiRegion(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(463493)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463494)
	}
	__antithesis_instrumentation__.Notify(463484)

	key := catalogkeys.MakeObjectNameKey(p.ExecCfg().Codec, dbDesc.GetID(), scDesc.GetID(), name.Object())
	if err = p.createDescriptorWithID(ctx, key, id, desc, jobDesc); err != nil {
		__antithesis_instrumentation__.Notify(463495)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463496)
	}
	__antithesis_instrumentation__.Notify(463485)

	seqValueKey := p.ExecCfg().Codec.SequenceKey(uint32(id))
	b := &kv.Batch{}
	if err := p.createdSequences.addCreatedSequence(id); err != nil {
		__antithesis_instrumentation__.Notify(463497)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463498)
	}
	__antithesis_instrumentation__.Notify(463486)
	b.Inc(seqValueKey, desc.SequenceOpts.Start-desc.SequenceOpts.Increment)
	if err := p.txn.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(463499)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463500)
	}
	__antithesis_instrumentation__.Notify(463487)

	if err := validateDescriptor(ctx, p, desc); err != nil {
		__antithesis_instrumentation__.Notify(463501)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463502)
	}
	__antithesis_instrumentation__.Notify(463488)

	return desc, p.logEvent(ctx,
		desc.ID,
		&eventpb.CreateSequence{
			SequenceName: name.FQString(),
		})
}

func createSequencesForSerialColumns(
	ctx context.Context,
	p *planner,
	sessionData *sessiondata.SessionData,
	db catalog.DatabaseDescriptor,
	sc catalog.SchemaDescriptor,
	n *tree.CreateTable,
) (map[tree.Name]*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(463503)
	colNameToSeqDesc := make(map[tree.Name]*tabledesc.Mutable)
	createStmt := n
	ensureCopy := func() {
		__antithesis_instrumentation__.Notify(463506)
		if createStmt == n {
			__antithesis_instrumentation__.Notify(463507)
			newCreateStmt := *n
			n.Defs = append(tree.TableDefs(nil), n.Defs...)
			createStmt = &newCreateStmt
		} else {
			__antithesis_instrumentation__.Notify(463508)
		}
	}
	__antithesis_instrumentation__.Notify(463504)

	tn := tree.MakeTableNameFromPrefix(catalog.ResolvedObjectPrefix{
		Database: db,
		Schema:   sc,
	}.NamePrefix(), tree.Name(n.Table.Table()))
	for i, def := range n.Defs {
		__antithesis_instrumentation__.Notify(463509)
		d, ok := def.(*tree.ColumnTableDef)
		if !ok {
			__antithesis_instrumentation__.Notify(463513)
			continue
		} else {
			__antithesis_instrumentation__.Notify(463514)
		}
		__antithesis_instrumentation__.Notify(463510)
		newDef, prefix, seqName, seqOpts, err := p.processSerialLikeInColumnDef(ctx, d, &tn)
		if err != nil {
			__antithesis_instrumentation__.Notify(463515)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463516)
		}
		__antithesis_instrumentation__.Notify(463511)

		if seqName != nil {
			__antithesis_instrumentation__.Notify(463517)
			seqDesc, err := doCreateSequence(
				ctx,
				p,
				sessionData,
				prefix.Database,
				prefix.Schema,
				seqName,
				n.Persistence,
				seqOpts,
				fmt.Sprintf("creating sequence %s for new table %s", seqName, n.Table.Table()),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(463519)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(463520)
			}
			__antithesis_instrumentation__.Notify(463518)
			colNameToSeqDesc[d.Name] = seqDesc
		} else {
			__antithesis_instrumentation__.Notify(463521)
		}
		__antithesis_instrumentation__.Notify(463512)
		if d != newDef {
			__antithesis_instrumentation__.Notify(463522)
			ensureCopy()
			n.Defs[i] = newDef
		} else {
			__antithesis_instrumentation__.Notify(463523)
		}
	}
	__antithesis_instrumentation__.Notify(463505)

	return colNameToSeqDesc, nil
}

func (*createSequenceNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(463524)
	return false, nil
}
func (*createSequenceNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(463525)
	return tree.Datums{}
}
func (*createSequenceNode) Close(context.Context) { __antithesis_instrumentation__.Notify(463526) }

func NewSequenceTableDesc(
	ctx context.Context,
	p *planner,
	settings *clustersettings.Settings,
	sequenceName string,
	sequenceOptions tree.SequenceOptions,
	parentID descpb.ID,
	schemaID descpb.ID,
	id descpb.ID,
	creationTime hlc.Timestamp,
	privileges *catpb.PrivilegeDescriptor,
	persistence tree.Persistence,
	isMultiRegion bool,
) (*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(463527)
	desc := tabledesc.InitTableDescriptor(
		id,
		parentID,
		schemaID,
		sequenceName,
		creationTime,
		privileges,
		persistence,
	)

	desc.Columns = []descpb.ColumnDescriptor{
		{
			ID:   tabledesc.SequenceColumnID,
			Name: tabledesc.SequenceColumnName,
			Type: types.Int,
		},
	}
	desc.SetPrimaryIndex(descpb.IndexDescriptor{
		ID:                  keys.SequenceIndexID,
		Name:                tabledesc.LegacyPrimaryKeyIndexName,
		KeyColumnIDs:        []descpb.ColumnID{tabledesc.SequenceColumnID},
		KeyColumnNames:      []string{tabledesc.SequenceColumnName},
		KeyColumnDirections: []descpb.IndexDescriptor_Direction{descpb.IndexDescriptor_ASC},
		EncodingType:        descpb.PrimaryIndexEncoding,
		Version:             descpb.PrimaryIndexWithStoredColumnsVersion,
		CreatedAtNanos:      creationTime.WallTime,
	})
	desc.Families = []descpb.ColumnFamilyDescriptor{
		{
			ID:              keys.SequenceColumnFamilyID,
			ColumnIDs:       []descpb.ColumnID{tabledesc.SequenceColumnID},
			ColumnNames:     []string{tabledesc.SequenceColumnName},
			Name:            "primary",
			DefaultColumnID: tabledesc.SequenceColumnID,
		},
	}

	opts := &descpb.TableDescriptor_SequenceOpts{
		Increment: 1,
	}
	if err := assignSequenceOptions(
		ctx,
		p,
		opts,
		sequenceOptions,
		true,
		id,
		parentID,
		nil,
	); err != nil {
		__antithesis_instrumentation__.Notify(463531)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463532)
	}
	__antithesis_instrumentation__.Notify(463528)
	desc.SequenceOpts = opts

	desc.State = descpb.DescriptorState_PUBLIC

	if isMultiRegion {
		__antithesis_instrumentation__.Notify(463533)
		desc.SetTableLocalityRegionalByTable(tree.PrimaryRegionNotSpecifiedName)
	} else {
		__antithesis_instrumentation__.Notify(463534)
	}
	__antithesis_instrumentation__.Notify(463529)

	version := settings.Version.ActiveVersion(ctx)
	if err := descbuilder.ValidateSelf(&desc, version); err != nil {
		__antithesis_instrumentation__.Notify(463535)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463536)
	}
	__antithesis_instrumentation__.Notify(463530)
	return &desc, nil
}
