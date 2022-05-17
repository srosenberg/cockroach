package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/constraint"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

const virtualSchemaNotImplementedMessage = "virtual schema table not implemented: %s.%s"

type virtualSchema struct {
	name            string
	undefinedTables map[string]struct{}
	tableDefs       map[descpb.ID]virtualSchemaDef
	tableValidator  func(*descpb.TableDescriptor) error

	validWithNoDatabaseContext bool

	containsTypes bool
}

type virtualSchemaDef interface {
	getSchema() string
	initVirtualTableDesc(
		ctx context.Context, st *cluster.Settings, sc catalog.SchemaDescriptor, id descpb.ID,
	) (descpb.TableDescriptor, error)
	getComment() string
	isUnimplemented() bool
}

type virtualIndex struct {
	populate func(ctx context.Context, constraint tree.Datum, p *planner, db catalog.DatabaseDescriptor,
		addRow func(...tree.Datum) error,
	) (matched bool, err error)

	partial bool
}

type virtualSchemaTable struct {
	schema string

	comment string

	populate func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error

	indexes []virtualIndex

	generator func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor, stopper *stop.Stopper) (virtualTableGenerator, cleanupFunc, error)

	unimplemented bool
}

type virtualSchemaView struct {
	schema        string
	resultColumns colinfo.ResultColumns
}

func (t virtualSchemaTable) getSchema() string {
	__antithesis_instrumentation__.Notify(632422)
	return t.schema
}

func (t virtualSchemaTable) initVirtualTableDesc(
	ctx context.Context, st *cluster.Settings, sc catalog.SchemaDescriptor, id descpb.ID,
) (descpb.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(632423)
	stmt, err := parser.ParseOne(t.schema)
	if err != nil {
		__antithesis_instrumentation__.Notify(632429)
		return descpb.TableDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(632430)
	}
	__antithesis_instrumentation__.Notify(632424)

	create := stmt.AST.(*tree.CreateTable)
	var firstColDef *tree.ColumnTableDef
	for _, def := range create.Defs {
		__antithesis_instrumentation__.Notify(632431)
		if d, ok := def.(*tree.ColumnTableDef); ok {
			__antithesis_instrumentation__.Notify(632433)
			if d.HasDefaultExpr() {
				__antithesis_instrumentation__.Notify(632435)
				return descpb.TableDescriptor{},
					errors.Errorf("virtual tables are not allowed to use default exprs "+
						"because bootstrapping: %s:%s", &create.Table, d.Name)
			} else {
				__antithesis_instrumentation__.Notify(632436)
			}
			__antithesis_instrumentation__.Notify(632434)
			if firstColDef == nil {
				__antithesis_instrumentation__.Notify(632437)
				firstColDef = d
			} else {
				__antithesis_instrumentation__.Notify(632438)
			}
		} else {
			__antithesis_instrumentation__.Notify(632439)
		}
		__antithesis_instrumentation__.Notify(632432)
		if _, ok := def.(*tree.UniqueConstraintTableDef); ok {
			__antithesis_instrumentation__.Notify(632440)
			return descpb.TableDescriptor{},
				errors.Errorf("virtual tables are not allowed to have unique constraints")
		} else {
			__antithesis_instrumentation__.Notify(632441)
		}
	}
	__antithesis_instrumentation__.Notify(632425)
	if firstColDef == nil {
		__antithesis_instrumentation__.Notify(632442)
		return descpb.TableDescriptor{},
			errors.Errorf("can't have empty virtual tables")
	} else {
		__antithesis_instrumentation__.Notify(632443)
	}
	__antithesis_instrumentation__.Notify(632426)

	mutDesc, err := NewTableDesc(
		ctx,
		nil,
		nil,
		st,
		create,
		nil,
		sc,
		id,
		nil,
		startTime,
		catpb.NewVirtualTablePrivilegeDescriptor(),
		nil,
		nil,
		nil,
		&sessiondata.SessionData{},
		tree.PersistencePermanent,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(632444)
		err = errors.Wrapf(err, "initVirtualDesc problem with schema: \n%s", t.schema)
		return descpb.TableDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(632445)
	}
	__antithesis_instrumentation__.Notify(632427)
	for _, index := range mutDesc.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(632446)
		if index.NumKeyColumns() > 1 {
			__antithesis_instrumentation__.Notify(632449)
			panic("we don't know how to deal with virtual composite indexes yet")
		} else {
			__antithesis_instrumentation__.Notify(632450)
		}
		__antithesis_instrumentation__.Notify(632447)
		idx := index.IndexDescDeepCopy()
		idx.StoreColumnNames, idx.StoreColumnIDs = nil, nil
		publicColumns := mutDesc.PublicColumns()
		presentInIndex := catalog.MakeTableColSet(idx.KeyColumnIDs...)
		for _, col := range publicColumns {
			__antithesis_instrumentation__.Notify(632451)
			if col.IsVirtual() || func() bool {
				__antithesis_instrumentation__.Notify(632453)
				return presentInIndex.Contains(col.GetID()) == true
			}() == true {
				__antithesis_instrumentation__.Notify(632454)
				continue
			} else {
				__antithesis_instrumentation__.Notify(632455)
			}
			__antithesis_instrumentation__.Notify(632452)
			idx.StoreColumnIDs = append(idx.StoreColumnIDs, col.GetID())
			idx.StoreColumnNames = append(idx.StoreColumnNames, col.GetName())
		}
		__antithesis_instrumentation__.Notify(632448)
		mutDesc.SetPublicNonPrimaryIndex(index.Ordinal(), idx)
	}
	__antithesis_instrumentation__.Notify(632428)
	return mutDesc.TableDescriptor, nil
}

func (t virtualSchemaTable) getComment() string {
	__antithesis_instrumentation__.Notify(632456)
	return t.comment
}

func (t virtualSchemaTable) getIndex(id descpb.IndexID) *virtualIndex {
	__antithesis_instrumentation__.Notify(632457)

	return &t.indexes[id-2]
}

func (t virtualSchemaTable) isUnimplemented() bool {
	__antithesis_instrumentation__.Notify(632458)
	return t.unimplemented
}

func (v virtualSchemaView) getSchema() string {
	__antithesis_instrumentation__.Notify(632459)
	return v.schema
}

func (v virtualSchemaView) initVirtualTableDesc(
	ctx context.Context, st *cluster.Settings, sc catalog.SchemaDescriptor, id descpb.ID,
) (descpb.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(632460)
	stmt, err := parser.ParseOne(v.schema)
	if err != nil {
		__antithesis_instrumentation__.Notify(632463)
		return descpb.TableDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(632464)
	}
	__antithesis_instrumentation__.Notify(632461)

	create := stmt.AST.(*tree.CreateView)

	columns := v.resultColumns
	if len(create.ColumnNames) != 0 {
		__antithesis_instrumentation__.Notify(632465)
		columns = overrideColumnNames(columns, create.ColumnNames)
	} else {
		__antithesis_instrumentation__.Notify(632466)
	}
	__antithesis_instrumentation__.Notify(632462)
	mutDesc, err := makeViewTableDesc(
		ctx,
		create.Name.Table(),
		tree.AsStringWithFlags(create.AsSource, tree.FmtParsable),
		0,
		sc.GetID(),
		id,
		columns,
		startTime,
		catpb.NewVirtualTablePrivilegeDescriptor(),
		nil,
		nil,
		st,
		tree.PersistencePermanent,
		false,
		nil,
	)
	return mutDesc.TableDescriptor, err
}

func (v virtualSchemaView) getComment() string {
	__antithesis_instrumentation__.Notify(632467)
	return ""
}

func (v virtualSchemaView) isUnimplemented() bool {
	__antithesis_instrumentation__.Notify(632468)
	return false
}

var virtualSchemas = map[descpb.ID]virtualSchema{
	catconstants.InformationSchemaID: informationSchema,
	catconstants.PgCatalogID:         pgCatalog,
	catconstants.CrdbInternalID:      crdbInternal,
	catconstants.PgExtensionSchemaID: pgExtension,
}

var startTime = hlc.Timestamp{
	WallTime: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC).UnixNano(),
}

type VirtualSchemaHolder struct {
	schemasByName map[string]*virtualSchemaEntry
	schemasByID   map[descpb.ID]*virtualSchemaEntry
	defsByID      map[descpb.ID]*virtualDefEntry
	orderedNames  []string
}

var _ VirtualTabler = (*VirtualSchemaHolder)(nil)

func (vs *VirtualSchemaHolder) GetVirtualSchema(schemaName string) (catalog.VirtualSchema, bool) {
	__antithesis_instrumentation__.Notify(632469)
	sc, ok := vs.schemasByName[schemaName]
	return sc, ok
}

func (vs *VirtualSchemaHolder) GetVirtualSchemaByID(id descpb.ID) (catalog.VirtualSchema, bool) {
	__antithesis_instrumentation__.Notify(632470)
	sc, ok := vs.schemasByID[id]
	return sc, ok
}

func (vs *VirtualSchemaHolder) GetVirtualObjectByID(id descpb.ID) (catalog.VirtualObject, bool) {
	__antithesis_instrumentation__.Notify(632471)
	entry, ok := vs.defsByID[id]
	if !ok {
		__antithesis_instrumentation__.Notify(632473)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(632474)
	}
	__antithesis_instrumentation__.Notify(632472)
	return entry, true
}

var _ catalog.VirtualSchemas = (*VirtualSchemaHolder)(nil)

type virtualSchemaEntry struct {
	desc            catalog.SchemaDescriptor
	defs            map[string]*virtualDefEntry
	orderedDefNames []string
	undefinedTables map[string]struct{}
	containsTypes   bool
}

func (v *virtualSchemaEntry) Desc() catalog.SchemaDescriptor {
	__antithesis_instrumentation__.Notify(632475)
	return v.desc
}

func (v *virtualSchemaEntry) NumTables() int {
	__antithesis_instrumentation__.Notify(632476)
	return len(v.defs)
}

func (v *virtualSchemaEntry) VisitTables(f func(object catalog.VirtualObject)) {
	__antithesis_instrumentation__.Notify(632477)
	for _, name := range v.orderedDefNames {
		__antithesis_instrumentation__.Notify(632478)
		f(v.defs[name])
	}
}

func (v *virtualSchemaEntry) GetObjectByName(
	name string, flags tree.ObjectLookupFlags,
) (catalog.VirtualObject, error) {
	__antithesis_instrumentation__.Notify(632479)
	switch flags.DesiredObjectKind {
	case tree.TableObject:
		__antithesis_instrumentation__.Notify(632480)
		if def, ok := v.defs[name]; ok {
			__antithesis_instrumentation__.Notify(632488)
			if flags.RequireMutable {
				__antithesis_instrumentation__.Notify(632490)
				return &mutableVirtualDefEntry{
					desc: tabledesc.NewBuilder(def.desc.TableDesc()).BuildExistingMutableTable(),
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(632491)
			}
			__antithesis_instrumentation__.Notify(632489)
			return def, nil
		} else {
			__antithesis_instrumentation__.Notify(632492)
		}
		__antithesis_instrumentation__.Notify(632481)
		if _, ok := v.undefinedTables[name]; ok {
			__antithesis_instrumentation__.Notify(632493)
			return nil, newUnimplementedVirtualTableError(v.desc.GetName(), name)
		} else {
			__antithesis_instrumentation__.Notify(632494)
		}
		__antithesis_instrumentation__.Notify(632482)
		return nil, nil
	case tree.TypeObject:
		__antithesis_instrumentation__.Notify(632483)
		if !v.containsTypes {
			__antithesis_instrumentation__.Notify(632495)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(632496)
		}
		__antithesis_instrumentation__.Notify(632484)

		typRef, err := parser.GetTypeReferenceFromName(tree.Name(name))
		if err != nil {
			__antithesis_instrumentation__.Notify(632497)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(632498)
		}
		__antithesis_instrumentation__.Notify(632485)

		typ, ok := tree.GetStaticallyKnownType(typRef)
		if !ok {
			__antithesis_instrumentation__.Notify(632499)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(632500)
		}
		__antithesis_instrumentation__.Notify(632486)

		return &virtualTypeEntry{
			desc:    typedesc.MakeSimpleAlias(typ, catconstants.PgCatalogID),
			mutable: flags.RequireMutable,
		}, nil
	default:
		__antithesis_instrumentation__.Notify(632487)
		return nil, errors.AssertionFailedf("unknown desired object kind %d", flags.DesiredObjectKind)
	}
}

type virtualDefEntry struct {
	virtualDef                 virtualSchemaDef
	desc                       catalog.TableDescriptor
	comment                    string
	validWithNoDatabaseContext bool
	unimplemented              bool
}

func (e *virtualDefEntry) Desc() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(632501)
	return e.desc
}

func canQueryVirtualTable(evalCtx *tree.EvalContext, e *virtualDefEntry) bool {
	__antithesis_instrumentation__.Notify(632502)
	return !e.unimplemented || func() bool {
		__antithesis_instrumentation__.Notify(632503)
		return evalCtx == nil == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(632504)
		return evalCtx.SessionData() == nil == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(632505)
		return evalCtx.SessionData().StubCatalogTablesEnabled == true
	}() == true
}

type mutableVirtualDefEntry struct {
	desc *tabledesc.Mutable
}

func (e *mutableVirtualDefEntry) Desc() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(632506)
	return e.desc
}

type virtualTypeEntry struct {
	desc    catalog.TypeDescriptor
	mutable bool
}

func (e *virtualTypeEntry) Desc() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(632507)

	return e.desc
}

type virtualTableConstructor func(context.Context, *planner, string) (planNode, error)

var errInvalidDbPrefix = errors.WithHint(
	pgerror.New(pgcode.UndefinedObject,
		"cannot access virtual schema in anonymous database"),
	"verify that the current database is set")

func newInvalidVirtualSchemaError() error {
	__antithesis_instrumentation__.Notify(632508)
	return errors.AssertionFailedf("virtualSchema cannot have both the populate and generator functions defined")
}

func newInvalidVirtualDefEntryError() error {
	__antithesis_instrumentation__.Notify(632509)
	return errors.AssertionFailedf("virtualDefEntry.virtualDef must be a virtualSchemaTable")
}

func (e *virtualDefEntry) validateRow(datums tree.Datums, columns colinfo.ResultColumns) error {
	__antithesis_instrumentation__.Notify(632510)
	if r, c := len(datums), len(columns); r != c {
		__antithesis_instrumentation__.Notify(632513)
		return errors.AssertionFailedf("datum row count and column count differ: %d vs %d", r, c)
	} else {
		__antithesis_instrumentation__.Notify(632514)
	}
	__antithesis_instrumentation__.Notify(632511)
	for i := range columns {
		__antithesis_instrumentation__.Notify(632515)
		col := &columns[i]
		datum := datums[i]
		if datum == tree.DNull {
			__antithesis_instrumentation__.Notify(632516)
			if !e.desc.PublicColumns()[i].IsNullable() {
				__antithesis_instrumentation__.Notify(632517)
				return errors.AssertionFailedf("column %s.%s not nullable, but found NULL value",
					e.desc.GetName(), col.Name)
			} else {
				__antithesis_instrumentation__.Notify(632518)
			}
		} else {
			__antithesis_instrumentation__.Notify(632519)
			if !datum.ResolvedType().Equivalent(col.Typ) {
				__antithesis_instrumentation__.Notify(632520)
				return errors.AssertionFailedf("datum column %q expected to be type %s; found type %s",
					col.Name, col.Typ, datum.ResolvedType())
			} else {
				__antithesis_instrumentation__.Notify(632521)
			}
		}
	}
	__antithesis_instrumentation__.Notify(632512)
	return nil
}

func (e *virtualDefEntry) getPlanInfo(
	table catalog.TableDescriptor,
	index catalog.Index,
	idxConstraint *constraint.Constraint,
	stopper *stop.Stopper,
) (colinfo.ResultColumns, virtualTableConstructor) {
	__antithesis_instrumentation__.Notify(632522)
	var columns colinfo.ResultColumns
	for _, col := range e.desc.PublicColumns() {
		__antithesis_instrumentation__.Notify(632525)
		columns = append(columns, colinfo.ResultColumn{
			Name:           col.GetName(),
			Typ:            col.GetType(),
			TableID:        table.GetID(),
			PGAttributeNum: col.GetPGAttributeNum(),
		})
	}
	__antithesis_instrumentation__.Notify(632523)

	constructor := func(ctx context.Context, p *planner, dbName string) (planNode, error) {
		__antithesis_instrumentation__.Notify(632526)
		var dbDesc catalog.DatabaseDescriptor
		var err error
		if dbName != "" {
			__antithesis_instrumentation__.Notify(632528)
			dbDesc, err = p.Descriptors().GetImmutableDatabaseByName(ctx, p.txn,
				dbName, tree.DatabaseLookupFlags{
					Required: true, AvoidLeased: p.avoidLeasedDescriptors,
				})
			if err != nil {
				__antithesis_instrumentation__.Notify(632529)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(632530)
			}
		} else {
			__antithesis_instrumentation__.Notify(632531)
			if !e.validWithNoDatabaseContext {
				__antithesis_instrumentation__.Notify(632532)
				return nil, errInvalidDbPrefix
			} else {
				__antithesis_instrumentation__.Notify(632533)
			}
		}
		__antithesis_instrumentation__.Notify(632527)

		switch def := e.virtualDef.(type) {
		case virtualSchemaTable:
			__antithesis_instrumentation__.Notify(632534)
			if def.generator != nil && func() bool {
				__antithesis_instrumentation__.Notify(632541)
				return def.populate != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(632542)
				return nil, newInvalidVirtualSchemaError()
			} else {
				__antithesis_instrumentation__.Notify(632543)
			}
			__antithesis_instrumentation__.Notify(632535)

			if def.generator != nil {
				__antithesis_instrumentation__.Notify(632544)
				next, cleanup, err := def.generator(ctx, p, dbDesc, stopper)
				if err != nil {
					__antithesis_instrumentation__.Notify(632546)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(632547)
				}
				__antithesis_instrumentation__.Notify(632545)
				return p.newVirtualTableNode(columns, next, cleanup), nil
			} else {
				__antithesis_instrumentation__.Notify(632548)
			}
			__antithesis_instrumentation__.Notify(632536)

			constrainedScan := idxConstraint != nil && func() bool {
				__antithesis_instrumentation__.Notify(632549)
				return !idxConstraint.IsUnconstrained() == true
			}() == true
			if !constrainedScan {
				__antithesis_instrumentation__.Notify(632550)
				generator, cleanup, setupError := setupGenerator(ctx, func(ctx context.Context, pusher rowPusher) error {
					__antithesis_instrumentation__.Notify(632553)
					return def.populate(ctx, p, dbDesc, func(row ...tree.Datum) error {
						__antithesis_instrumentation__.Notify(632554)
						if err := e.validateRow(row, columns); err != nil {
							__antithesis_instrumentation__.Notify(632556)
							return err
						} else {
							__antithesis_instrumentation__.Notify(632557)
						}
						__antithesis_instrumentation__.Notify(632555)
						return pusher.pushRow(row...)
					})
				}, stopper)
				__antithesis_instrumentation__.Notify(632551)
				if setupError != nil {
					__antithesis_instrumentation__.Notify(632558)
					return nil, setupError
				} else {
					__antithesis_instrumentation__.Notify(632559)
				}
				__antithesis_instrumentation__.Notify(632552)
				return p.newVirtualTableNode(columns, generator, cleanup), nil
			} else {
				__antithesis_instrumentation__.Notify(632560)
			}
			__antithesis_instrumentation__.Notify(632537)

			if index.GetID() == 1 {
				__antithesis_instrumentation__.Notify(632561)
				return nil, errors.AssertionFailedf(
					"programming error: can't constrain scan on primary virtual index of table %s", e.desc.GetName())
			} else {
				__antithesis_instrumentation__.Notify(632562)
			}
			__antithesis_instrumentation__.Notify(632538)

			columnIdxMap := catalog.ColumnIDToOrdinalMap(table.PublicColumns())
			indexKeyDatums := make([]tree.Datum, index.NumKeyColumns())

			generator, cleanup, setupError := setupGenerator(ctx, e.makeConstrainedRowsGenerator(
				p, dbDesc, index, indexKeyDatums, columnIdxMap, idxConstraint, columns), stopper)
			if setupError != nil {
				__antithesis_instrumentation__.Notify(632563)
				return nil, setupError
			} else {
				__antithesis_instrumentation__.Notify(632564)
			}
			__antithesis_instrumentation__.Notify(632539)
			return p.newVirtualTableNode(columns, generator, cleanup), nil

		default:
			__antithesis_instrumentation__.Notify(632540)
			return nil, newInvalidVirtualDefEntryError()
		}
	}
	__antithesis_instrumentation__.Notify(632524)

	return columns, constructor
}

func (e *virtualDefEntry) makeConstrainedRowsGenerator(
	p *planner,
	dbDesc catalog.DatabaseDescriptor,
	index catalog.Index,
	indexKeyDatums []tree.Datum,
	columnIdxMap catalog.TableColMap,
	idxConstraint *constraint.Constraint,
	columns colinfo.ResultColumns,
) func(ctx context.Context, pusher rowPusher) error {
	__antithesis_instrumentation__.Notify(632565)
	def := e.virtualDef.(virtualSchemaTable)
	return func(ctx context.Context, pusher rowPusher) error {
		__antithesis_instrumentation__.Notify(632566)
		var span constraint.Span
		addRowIfPassesFilter := func(idxConstraint *constraint.Constraint) func(datums ...tree.Datum) error {
			__antithesis_instrumentation__.Notify(632571)
			return func(datums ...tree.Datum) error {
				__antithesis_instrumentation__.Notify(632572)
				for i := 0; i < index.NumKeyColumns(); i++ {
					__antithesis_instrumentation__.Notify(632575)
					id := index.GetKeyColumnID(i)
					indexKeyDatums[i] = datums[columnIdxMap.GetDefault(id)]
				}
				__antithesis_instrumentation__.Notify(632573)

				key := constraint.MakeCompositeKey(indexKeyDatums...)
				span.Init(key, constraint.IncludeBoundary, key, constraint.IncludeBoundary)
				var err error
				if idxConstraint.ContainsSpan(p.EvalContext(), &span) {
					__antithesis_instrumentation__.Notify(632576)
					if err := e.validateRow(datums, columns); err != nil {
						__antithesis_instrumentation__.Notify(632578)
						return err
					} else {
						__antithesis_instrumentation__.Notify(632579)
					}
					__antithesis_instrumentation__.Notify(632577)
					return pusher.pushRow(datums...)
				} else {
					__antithesis_instrumentation__.Notify(632580)
				}
				__antithesis_instrumentation__.Notify(632574)
				return err
			}
		}
		__antithesis_instrumentation__.Notify(632567)

		var currentSpan int
		for ; currentSpan < idxConstraint.Spans.Count(); currentSpan++ {
			__antithesis_instrumentation__.Notify(632581)
			span := idxConstraint.Spans.Get(currentSpan)
			if span.StartKey().Length() > 1 {
				__antithesis_instrumentation__.Notify(632585)
				return errors.AssertionFailedf(
					"programming error: can't push down composite constraints into vtables")
			} else {
				__antithesis_instrumentation__.Notify(632586)
			}
			__antithesis_instrumentation__.Notify(632582)
			if !span.HasSingleKey(p.EvalContext()) {
				__antithesis_instrumentation__.Notify(632587)

				break
			} else {
				__antithesis_instrumentation__.Notify(632588)
			}
			__antithesis_instrumentation__.Notify(632583)
			constraintDatum := span.StartKey().Value(0)
			virtualIndex := def.getIndex(index.GetID())

			found, err := virtualIndex.populate(ctx, constraintDatum, p, dbDesc,
				addRowIfPassesFilter(idxConstraint))
			if err != nil {
				__antithesis_instrumentation__.Notify(632589)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632590)
			}
			__antithesis_instrumentation__.Notify(632584)
			if !found && func() bool {
				__antithesis_instrumentation__.Notify(632591)
				return virtualIndex.partial == true
			}() == true {
				__antithesis_instrumentation__.Notify(632592)

				break
			} else {
				__antithesis_instrumentation__.Notify(632593)
			}
		}
		__antithesis_instrumentation__.Notify(632568)
		if currentSpan == idxConstraint.Spans.Count() {
			__antithesis_instrumentation__.Notify(632594)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(632595)
		}
		__antithesis_instrumentation__.Notify(632569)

		newConstraint := *idxConstraint
		newConstraint.Spans = constraint.Spans{}
		nSpans := idxConstraint.Spans.Count() - currentSpan
		newConstraint.Spans.Alloc(nSpans)
		for ; currentSpan < idxConstraint.Spans.Count(); currentSpan++ {
			__antithesis_instrumentation__.Notify(632596)
			newConstraint.Spans.Append(idxConstraint.Spans.Get(currentSpan))
		}
		__antithesis_instrumentation__.Notify(632570)
		return def.populate(ctx, p, dbDesc, addRowIfPassesFilter(&newConstraint))
	}
}

func NewVirtualSchemaHolder(
	ctx context.Context, st *cluster.Settings,
) (*VirtualSchemaHolder, error) {
	__antithesis_instrumentation__.Notify(632597)
	vs := &VirtualSchemaHolder{
		schemasByName: make(map[string]*virtualSchemaEntry, len(virtualSchemas)),
		schemasByID:   make(map[descpb.ID]*virtualSchemaEntry, len(virtualSchemas)),
		orderedNames:  make([]string, len(virtualSchemas)),
		defsByID:      make(map[descpb.ID]*virtualDefEntry, math.MaxUint32-catconstants.MinVirtualID),
	}

	order := 0
	for schemaID, schema := range virtualSchemas {
		__antithesis_instrumentation__.Notify(632599)
		scDesc, ok := schemadesc.GetVirtualSchemaByID(schemaID)
		if !ok {
			__antithesis_instrumentation__.Notify(632603)
			return nil, errors.AssertionFailedf("failed to find virtual schema %d (%s)", schemaID, schema.name)
		} else {
			__antithesis_instrumentation__.Notify(632604)
		}
		__antithesis_instrumentation__.Notify(632600)
		if scDesc.GetName() != schema.name {
			__antithesis_instrumentation__.Notify(632605)
			return nil, errors.AssertionFailedf("schema name mismatch for virtual schema %d: expected %s, found %s",
				schemaID, schema.name, scDesc.GetName())
		} else {
			__antithesis_instrumentation__.Notify(632606)
		}
		__antithesis_instrumentation__.Notify(632601)
		defs := make(map[string]*virtualDefEntry, len(schema.tableDefs))
		orderedDefNames := make([]string, 0, len(schema.tableDefs))

		for id, def := range schema.tableDefs {
			__antithesis_instrumentation__.Notify(632607)
			tableDesc, err := def.initVirtualTableDesc(ctx, st, scDesc, id)

			if err != nil {
				__antithesis_instrumentation__.Notify(632611)
				return nil, errors.NewAssertionErrorWithWrappedErrf(err,
					"failed to initialize %s", errors.Safe(def.getSchema()))
			} else {
				__antithesis_instrumentation__.Notify(632612)
			}
			__antithesis_instrumentation__.Notify(632608)

			if schema.tableValidator != nil {
				__antithesis_instrumentation__.Notify(632613)
				if err := schema.tableValidator(&tableDesc); err != nil {
					__antithesis_instrumentation__.Notify(632614)
					return nil, errors.NewAssertionErrorWithWrappedErrf(err, "programmer error")
				} else {
					__antithesis_instrumentation__.Notify(632615)
				}
			} else {
				__antithesis_instrumentation__.Notify(632616)
			}
			__antithesis_instrumentation__.Notify(632609)
			td := tabledesc.NewBuilder(&tableDesc).BuildImmutableTable()
			version := st.Version.ActiveVersionOrEmpty(ctx)
			if err := descbuilder.ValidateSelf(td, version); err != nil {
				__antithesis_instrumentation__.Notify(632617)
				return nil, errors.NewAssertionErrorWithWrappedErrf(err,
					"failed to validate virtual table %s: programmer error", errors.Safe(td.GetName()))
			} else {
				__antithesis_instrumentation__.Notify(632618)
			}
			__antithesis_instrumentation__.Notify(632610)

			entry := &virtualDefEntry{
				virtualDef:                 def,
				desc:                       td,
				validWithNoDatabaseContext: schema.validWithNoDatabaseContext,
				comment:                    def.getComment(),
				unimplemented:              def.isUnimplemented(),
			}
			defs[tableDesc.Name] = entry
			vs.defsByID[tableDesc.ID] = entry
			orderedDefNames = append(orderedDefNames, tableDesc.Name)
		}
		__antithesis_instrumentation__.Notify(632602)

		sort.Strings(orderedDefNames)

		vse := &virtualSchemaEntry{
			desc:            scDesc,
			defs:            defs,
			orderedDefNames: orderedDefNames,
			undefinedTables: schema.undefinedTables,
			containsTypes:   schema.containsTypes,
		}
		vs.schemasByName[scDesc.GetName()] = vse
		vs.schemasByID[scDesc.GetID()] = vse
		vs.orderedNames[order] = scDesc.GetName()
		order++
	}
	__antithesis_instrumentation__.Notify(632598)
	sort.Strings(vs.orderedNames)
	return vs, nil
}

func newUnimplementedVirtualTableError(schema, tableName string) error {
	__antithesis_instrumentation__.Notify(632619)
	return unimplemented.Newf(
		fmt.Sprintf("%s.%s", schema, tableName),
		virtualSchemaNotImplementedMessage,
		schema,
		tableName,
	)
}

func (vs *VirtualSchemaHolder) getSchemas() map[string]*virtualSchemaEntry {
	__antithesis_instrumentation__.Notify(632620)
	return vs.schemasByName
}

func (vs *VirtualSchemaHolder) getSchemaNames() []string {
	__antithesis_instrumentation__.Notify(632621)
	return vs.orderedNames
}

func (vs *VirtualSchemaHolder) getVirtualSchemaEntry(name string) (*virtualSchemaEntry, bool) {
	__antithesis_instrumentation__.Notify(632622)
	e, ok := vs.schemasByName[name]
	return e, ok
}

func (vs *VirtualSchemaHolder) getVirtualTableEntry(tn *tree.TableName) (*virtualDefEntry, error) {
	__antithesis_instrumentation__.Notify(632623)
	if db, ok := vs.getVirtualSchemaEntry(tn.Schema()); ok {
		__antithesis_instrumentation__.Notify(632625)
		tableName := tn.Table()
		if t, ok := db.defs[tableName]; ok {
			__antithesis_instrumentation__.Notify(632628)
			sqltelemetry.IncrementGetVirtualTableEntry(tn.Schema(), tableName)
			return t, nil
		} else {
			__antithesis_instrumentation__.Notify(632629)
		}
		__antithesis_instrumentation__.Notify(632626)
		if _, ok := db.undefinedTables[tableName]; ok {
			__antithesis_instrumentation__.Notify(632630)
			return nil, unimplemented.NewWithIssueDetailf(
				8675,
				fmt.Sprintf("%s.%s", tn.Schema(), tableName),
				virtualSchemaNotImplementedMessage,
				tn.Schema(),
				tableName,
			)
		} else {
			__antithesis_instrumentation__.Notify(632631)
		}
		__antithesis_instrumentation__.Notify(632627)
		return nil, sqlerrors.NewUndefinedRelationError(tn)
	} else {
		__antithesis_instrumentation__.Notify(632632)
	}
	__antithesis_instrumentation__.Notify(632624)
	return nil, nil
}

type VirtualTabler interface {
	getVirtualTableDesc(tn *tree.TableName) (catalog.TableDescriptor, error)
	getVirtualTableEntry(tn *tree.TableName) (*virtualDefEntry, error)
	getSchemas() map[string]*virtualSchemaEntry
	getSchemaNames() []string
}

func (vs *VirtualSchemaHolder) getVirtualTableDesc(
	tn *tree.TableName,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(632633)
	t, err := vs.getVirtualTableEntry(tn)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(632635)
		return t == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(632636)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(632637)
	}
	__antithesis_instrumentation__.Notify(632634)
	return t.desc, nil
}
