package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descidgen"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type postgreStream struct {
	ctx                   context.Context
	s                     *bufio.Scanner
	copy                  *postgreStreamCopy
	unsupportedStmtLogger *unsupportedStmtLogger
}

func newPostgreStream(
	ctx context.Context, r io.Reader, max int, unsupportedStmtLogger *unsupportedStmtLogger,
) *postgreStream {
	__antithesis_instrumentation__.Notify(495873)
	s := bufio.NewScanner(r)
	s.Buffer(nil, max)
	p := &postgreStream{ctx: ctx, s: s, unsupportedStmtLogger: unsupportedStmtLogger}
	s.Split(p.split)
	return p
}

func (p *postgreStream) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	__antithesis_instrumentation__.Notify(495874)
	if p.copy == nil {
		__antithesis_instrumentation__.Notify(495876)
		return splitSQLSemicolon(data, atEOF)
	} else {
		__antithesis_instrumentation__.Notify(495877)
	}
	__antithesis_instrumentation__.Notify(495875)
	return bufio.ScanLines(data, atEOF)
}

func splitSQLSemicolon(data []byte, atEOF bool) (advance int, token []byte, err error) {
	__antithesis_instrumentation__.Notify(495878)
	if atEOF && func() bool {
		__antithesis_instrumentation__.Notify(495882)
		return len(data) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(495883)
		return 0, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(495884)
	}
	__antithesis_instrumentation__.Notify(495879)

	if pos, ok := parser.SplitFirstStatement(string(data)); ok {
		__antithesis_instrumentation__.Notify(495885)
		return pos, data[:pos], nil
	} else {
		__antithesis_instrumentation__.Notify(495886)
	}
	__antithesis_instrumentation__.Notify(495880)

	if atEOF {
		__antithesis_instrumentation__.Notify(495887)
		return len(data), data, nil
	} else {
		__antithesis_instrumentation__.Notify(495888)
	}
	__antithesis_instrumentation__.Notify(495881)

	return 0, nil, nil
}

func (p *postgreStream) Next() (interface{}, error) {
	__antithesis_instrumentation__.Notify(495889)
	if p.copy != nil {
		__antithesis_instrumentation__.Notify(495893)
		row, err := p.copy.Next()
		if errors.Is(err, errCopyDone) {
			__antithesis_instrumentation__.Notify(495895)
			p.copy = nil
			return errCopyDone, nil
		} else {
			__antithesis_instrumentation__.Notify(495896)
		}
		__antithesis_instrumentation__.Notify(495894)
		return row, err
	} else {
		__antithesis_instrumentation__.Notify(495897)
	}
	__antithesis_instrumentation__.Notify(495890)

	for p.s.Scan() {
		__antithesis_instrumentation__.Notify(495898)
		t := p.s.Text()
		skipOverComments(t)

		stmts, err := parser.Parse(t)
		if err != nil {
			__antithesis_instrumentation__.Notify(495900)

			if p.unsupportedStmtLogger.ignoreUnsupported && func() bool {
				__antithesis_instrumentation__.Notify(495902)
				return errors.HasType(err, (*tree.UnsupportedError)(nil)) == true
			}() == true {
				__antithesis_instrumentation__.Notify(495903)
				if unsupportedErr := (*tree.UnsupportedError)(nil); errors.As(err, &unsupportedErr) {
					__antithesis_instrumentation__.Notify(495905)
					err := p.unsupportedStmtLogger.log(unsupportedErr.FeatureName, true)
					if err != nil {
						__antithesis_instrumentation__.Notify(495906)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(495907)
					}
				} else {
					__antithesis_instrumentation__.Notify(495908)
				}
				__antithesis_instrumentation__.Notify(495904)
				continue
			} else {
				__antithesis_instrumentation__.Notify(495909)
			}
			__antithesis_instrumentation__.Notify(495901)
			return nil, wrapErrorWithUnsupportedHint(err)
		} else {
			__antithesis_instrumentation__.Notify(495910)
		}
		__antithesis_instrumentation__.Notify(495899)
		switch len(stmts) {
		case 0:
			__antithesis_instrumentation__.Notify(495911)

		case 1:
			__antithesis_instrumentation__.Notify(495912)

			if cf, ok := stmts[0].AST.(*tree.CopyFrom); ok && func() bool {
				__antithesis_instrumentation__.Notify(495915)
				return cf.Stdin == true
			}() == true {
				__antithesis_instrumentation__.Notify(495916)

				p.copy = newPostgreStreamCopy(p.s, copyDefaultDelimiter, copyDefaultNull)

				if !p.s.Scan() {
					__antithesis_instrumentation__.Notify(495919)
					return nil, errors.Errorf("expected empty line")
				} else {
					__antithesis_instrumentation__.Notify(495920)
				}
				__antithesis_instrumentation__.Notify(495917)
				if err := p.s.Err(); err != nil {
					__antithesis_instrumentation__.Notify(495921)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(495922)
				}
				__antithesis_instrumentation__.Notify(495918)
				if len(p.s.Bytes()) != 0 {
					__antithesis_instrumentation__.Notify(495923)
					return nil, errors.Errorf("expected empty line")
				} else {
					__antithesis_instrumentation__.Notify(495924)
				}
			} else {
				__antithesis_instrumentation__.Notify(495925)
			}
			__antithesis_instrumentation__.Notify(495913)
			return stmts[0].AST, nil
		default:
			__antithesis_instrumentation__.Notify(495914)
			return nil, errors.Errorf("unexpected: got %d statements", len(stmts))
		}
	}
	__antithesis_instrumentation__.Notify(495891)
	if err := p.s.Err(); err != nil {
		__antithesis_instrumentation__.Notify(495926)
		if errors.Is(err, bufio.ErrTooLong) {
			__antithesis_instrumentation__.Notify(495928)
			err = wrapWithLineTooLongHint(
				errors.HandledWithMessage(err, "line too long"),
			)
		} else {
			__antithesis_instrumentation__.Notify(495929)
		}
		__antithesis_instrumentation__.Notify(495927)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(495930)
	}
	__antithesis_instrumentation__.Notify(495892)
	return nil, io.EOF
}

var ignoreComments = regexp.MustCompile(`^\s*(--.*)`)

func skipOverComments(s string) {
	__antithesis_instrumentation__.Notify(495931)

	for {
		__antithesis_instrumentation__.Notify(495932)
		m := ignoreComments.FindStringIndex(s)
		if m == nil {
			__antithesis_instrumentation__.Notify(495934)
			break
		} else {
			__antithesis_instrumentation__.Notify(495935)
		}
		__antithesis_instrumentation__.Notify(495933)
		s = s[m[1]:]
	}
}

type regclassRewriter struct{}

var _ tree.Visitor = regclassRewriter{}

func (regclassRewriter) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	__antithesis_instrumentation__.Notify(495936)
	switch t := expr.(type) {
	case *tree.FuncExpr:
		__antithesis_instrumentation__.Notify(495938)
		switch t.Func.String() {
		case "nextval":
			__antithesis_instrumentation__.Notify(495939)
			if len(t.Exprs) > 0 {
				__antithesis_instrumentation__.Notify(495941)
				switch e := t.Exprs[0].(type) {
				case *tree.CastExpr:
					__antithesis_instrumentation__.Notify(495942)
					if typ, ok := tree.GetStaticallyKnownType(e.Type); ok && func() bool {
						__antithesis_instrumentation__.Notify(495943)
						return typ.Oid() == oid.T_regclass == true
					}() == true {
						__antithesis_instrumentation__.Notify(495944)

						t.Exprs[0] = e.Expr
					} else {
						__antithesis_instrumentation__.Notify(495945)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(495946)
			}
		default:
			__antithesis_instrumentation__.Notify(495940)
		}
	}
	__antithesis_instrumentation__.Notify(495937)
	return true, expr
}

func (regclassRewriter) VisitPost(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(495947)
	return expr
}

func removeDefaultRegclass(create *tree.CreateTable) {
	__antithesis_instrumentation__.Notify(495948)
	for _, def := range create.Defs {
		__antithesis_instrumentation__.Notify(495949)
		switch def := def.(type) {
		case *tree.ColumnTableDef:
			__antithesis_instrumentation__.Notify(495950)
			if def.DefaultExpr.Expr != nil {
				__antithesis_instrumentation__.Notify(495951)
				def.DefaultExpr.Expr, _ = tree.WalkExpr(regclassRewriter{}, def.DefaultExpr.Expr)
			} else {
				__antithesis_instrumentation__.Notify(495952)
			}
		}
	}
}

type schemaAndTableName struct {
	schema string
	table  string
}

func (s *schemaAndTableName) String() string {
	__antithesis_instrumentation__.Notify(495953)
	var ret string
	if s.schema != "" {
		__antithesis_instrumentation__.Notify(495955)
		ret += s.schema + "."
	} else {
		__antithesis_instrumentation__.Notify(495956)
	}
	__antithesis_instrumentation__.Notify(495954)
	ret += s.table
	return ret
}

type schemaParsingObjects struct {
	createSchema map[string]*tree.CreateSchema
	createTbl    map[schemaAndTableName]*tree.CreateTable
	createSeq    map[schemaAndTableName]*tree.CreateSequence
	tableFKs     map[schemaAndTableName][]*tree.ForeignKeyConstraintTableDef
}

func createPostgresSchemas(
	ctx context.Context,
	parentID descpb.ID,
	schemasToCreate map[string]*tree.CreateSchema,
	execCfg *sql.ExecutorConfig,
	sessionData *sessiondata.SessionData,
) ([]*schemadesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(495957)
	createSchema := func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
		dbDesc catalog.DatabaseDescriptor, schema *tree.CreateSchema,
	) (*schemadesc.Mutable, error) {
		__antithesis_instrumentation__.Notify(495961)
		desc, _, err := sql.CreateUserDefinedSchemaDescriptor(
			ctx, sessionData, schema, txn, descriptors, execCfg, dbDesc, false,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(495965)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(495966)
		}
		__antithesis_instrumentation__.Notify(495962)

		if desc == nil {
			__antithesis_instrumentation__.Notify(495967)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(495968)
		}
		__antithesis_instrumentation__.Notify(495963)

		desc.ID, err = getNextPlaceholderDescID(ctx, execCfg)
		if err != nil {
			__antithesis_instrumentation__.Notify(495969)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(495970)
		}
		__antithesis_instrumentation__.Notify(495964)
		desc.SetOffline("importing")
		return desc, nil
	}
	__antithesis_instrumentation__.Notify(495958)
	var schemaDescs []*schemadesc.Mutable
	createSchemaDescs := func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(495971)
		schemaDescs = nil
		_, dbDesc, err := descriptors.GetImmutableDatabaseByID(ctx, txn, parentID, tree.DatabaseLookupFlags{
			Required:    true,
			AvoidLeased: true,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(495974)
			return err
		} else {
			__antithesis_instrumentation__.Notify(495975)
		}
		__antithesis_instrumentation__.Notify(495972)
		for _, schema := range schemasToCreate {
			__antithesis_instrumentation__.Notify(495976)
			scDesc, err := createSchema(ctx, txn, descriptors, dbDesc, schema)
			if err != nil {
				__antithesis_instrumentation__.Notify(495978)
				return err
			} else {
				__antithesis_instrumentation__.Notify(495979)
			}
			__antithesis_instrumentation__.Notify(495977)
			if scDesc != nil {
				__antithesis_instrumentation__.Notify(495980)
				schemaDescs = append(schemaDescs, scDesc)
			} else {
				__antithesis_instrumentation__.Notify(495981)
			}
		}
		__antithesis_instrumentation__.Notify(495973)
		return nil
	}
	__antithesis_instrumentation__.Notify(495959)
	if err := sql.DescsTxn(ctx, execCfg, createSchemaDescs); err != nil {
		__antithesis_instrumentation__.Notify(495982)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(495983)
	}
	__antithesis_instrumentation__.Notify(495960)
	return schemaDescs, nil
}

func createPostgresSequences(
	ctx context.Context,
	parentID descpb.ID,
	createSeq map[schemaAndTableName]*tree.CreateSequence,
	fks fkHandler,
	walltime int64,
	owner security.SQLUsername,
	schemaNameToDesc map[string]*schemadesc.Mutable,
	execCfg *sql.ExecutorConfig,
) ([]*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(495984)
	ret := make([]*tabledesc.Mutable, 0)
	for schemaAndTableName, seq := range createSeq {
		__antithesis_instrumentation__.Notify(495986)
		schema, err := getSchemaByNameFromMap(ctx, schemaAndTableName, schemaNameToDesc, execCfg.Settings.Version)
		if err != nil {
			__antithesis_instrumentation__.Notify(495990)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(495991)
		}
		__antithesis_instrumentation__.Notify(495987)
		id, err := getNextPlaceholderDescID(ctx, execCfg)
		if err != nil {
			__antithesis_instrumentation__.Notify(495992)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(495993)
		}
		__antithesis_instrumentation__.Notify(495988)
		desc, err := sql.NewSequenceTableDesc(
			ctx,
			nil,
			execCfg.Settings,
			schemaAndTableName.table,
			seq.Options,
			parentID,
			schema.GetID(),
			id,
			hlc.Timestamp{WallTime: walltime},
			catpb.NewBasePrivilegeDescriptor(owner),
			tree.PersistencePermanent,

			false,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(495994)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(495995)
		}
		__antithesis_instrumentation__.Notify(495989)
		fks.resolver.tableNameToDesc[schemaAndTableName.String()] = desc
		ret = append(ret, desc)
	}
	__antithesis_instrumentation__.Notify(495985)

	return ret, nil
}

func getSchemaByNameFromMap(
	ctx context.Context,
	schemaAndTableName schemaAndTableName,
	schemaNameToDesc map[string]*schemadesc.Mutable,
	version clusterversion.Handle,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(495996)
	var schema catalog.SchemaDescriptor
	if !version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) && func() bool {
		__antithesis_instrumentation__.Notify(495999)
		return (schemaAndTableName.schema == "" || func() bool {
			__antithesis_instrumentation__.Notify(496000)
			return schemaAndTableName.schema == tree.PublicSchema == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(496001)
		return schemadesc.GetPublicSchema(), nil
	} else {
		__antithesis_instrumentation__.Notify(496002)
	}
	__antithesis_instrumentation__.Notify(495997)
	var ok bool
	if schema, ok = schemaNameToDesc[schemaAndTableName.schema]; !ok {
		__antithesis_instrumentation__.Notify(496003)
		return nil, errors.Newf("schema %q not found in the schemas created from the pgdump",
			schema)
	} else {
		__antithesis_instrumentation__.Notify(496004)
	}
	__antithesis_instrumentation__.Notify(495998)
	return schema, nil
}

func createPostgresTables(
	evalCtx *tree.EvalContext,
	p sql.JobExecContext,
	createTbl map[schemaAndTableName]*tree.CreateTable,
	fks fkHandler,
	backrefs map[descpb.ID]*tabledesc.Mutable,
	parentDB catalog.DatabaseDescriptor,
	walltime int64,
	schemaNameToDesc map[string]*schemadesc.Mutable,
) ([]*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(496005)
	ret := make([]*tabledesc.Mutable, 0)
	for schemaAndTableName, create := range createTbl {
		__antithesis_instrumentation__.Notify(496007)
		if create == nil {
			__antithesis_instrumentation__.Notify(496012)
			continue
		} else {
			__antithesis_instrumentation__.Notify(496013)
		}
		__antithesis_instrumentation__.Notify(496008)
		schema, err := getSchemaByNameFromMap(evalCtx.Ctx(), schemaAndTableName, schemaNameToDesc, evalCtx.Settings.Version)
		if err != nil {
			__antithesis_instrumentation__.Notify(496014)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(496015)
		}
		__antithesis_instrumentation__.Notify(496009)
		removeDefaultRegclass(create)

		semaCtxPtr := makeSemaCtxWithoutTypeResolver(p.SemaCtx())
		id, err := getNextPlaceholderDescID(evalCtx.Ctx(), p.ExecCfg())
		if err != nil {
			__antithesis_instrumentation__.Notify(496016)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(496017)
		}
		__antithesis_instrumentation__.Notify(496010)
		desc, err := MakeSimpleTableDescriptor(evalCtx.Ctx(), semaCtxPtr, p.ExecCfg().Settings,
			create, parentDB, schema, id, fks, walltime)
		if err != nil {
			__antithesis_instrumentation__.Notify(496018)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(496019)
		}
		__antithesis_instrumentation__.Notify(496011)
		fks.resolver.tableNameToDesc[schemaAndTableName.String()] = desc
		backrefs[desc.ID] = desc
		ret = append(ret, desc)
	}
	__antithesis_instrumentation__.Notify(496006)

	return ret, nil
}

func resolvePostgresFKs(
	evalCtx *tree.EvalContext,
	parentDB catalog.DatabaseDescriptor,
	tableFKs map[schemaAndTableName][]*tree.ForeignKeyConstraintTableDef,
	fks fkHandler,
	backrefs map[descpb.ID]*tabledesc.Mutable,
	schemaNameToDesc map[string]*schemadesc.Mutable,
) error {
	__antithesis_instrumentation__.Notify(496020)
	for schemaAndTableName, constraints := range tableFKs {
		__antithesis_instrumentation__.Notify(496022)
		desc := fks.resolver.tableNameToDesc[schemaAndTableName.String()]
		if desc == nil {
			__antithesis_instrumentation__.Notify(496026)
			continue
		} else {
			__antithesis_instrumentation__.Notify(496027)
		}
		__antithesis_instrumentation__.Notify(496023)
		schema, err := getSchemaByNameFromMap(evalCtx.Ctx(), schemaAndTableName, schemaNameToDesc, evalCtx.Settings.Version)
		if err != nil {
			__antithesis_instrumentation__.Notify(496028)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496029)
		}
		__antithesis_instrumentation__.Notify(496024)
		for _, constraint := range constraints {
			__antithesis_instrumentation__.Notify(496030)
			if constraint.Table.Schema() == "" {
				__antithesis_instrumentation__.Notify(496033)
				return errors.Errorf("schema expected to be non-empty when resolving postgres FK %s",
					constraint.Name.String())
			} else {
				__antithesis_instrumentation__.Notify(496034)
			}
			__antithesis_instrumentation__.Notify(496031)
			constraint.Table.ExplicitSchema = true

			if constraint.Table.Catalog() == "" {
				__antithesis_instrumentation__.Notify(496035)
				constraint.Table.ExplicitCatalog = true
				constraint.Table.CatalogName = "defaultdb"
			} else {
				__antithesis_instrumentation__.Notify(496036)
			}
			__antithesis_instrumentation__.Notify(496032)
			if err := sql.ResolveFK(
				evalCtx.Ctx(), nil, &fks.resolver,
				parentDB, schema, desc,
				constraint, backrefs, sql.NewTable,
				tree.ValidationDefault, evalCtx,
			); err != nil {
				__antithesis_instrumentation__.Notify(496037)
				return err
			} else {
				__antithesis_instrumentation__.Notify(496038)
			}
		}
		__antithesis_instrumentation__.Notify(496025)
		if err := fixDescriptorFKState(desc); err != nil {
			__antithesis_instrumentation__.Notify(496039)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496040)
		}
	}
	__antithesis_instrumentation__.Notify(496021)

	return nil
}

func getNextPlaceholderDescID(
	ctx context.Context, execCfg *sql.ExecutorConfig,
) (_ descpb.ID, err error) {
	__antithesis_instrumentation__.Notify(496041)
	if placeholderID == 0 {
		__antithesis_instrumentation__.Notify(496043)
		placeholderID, err = descidgen.PeekNextUniqueDescID(ctx, execCfg.DB, execCfg.Codec)
		if err != nil {
			__antithesis_instrumentation__.Notify(496044)
			return descpb.InvalidID, err
		} else {
			__antithesis_instrumentation__.Notify(496045)
		}
	} else {
		__antithesis_instrumentation__.Notify(496046)
	}
	__antithesis_instrumentation__.Notify(496042)
	ret := placeholderID
	placeholderID++
	return ret, nil
}

var placeholderID descpb.ID

func readPostgresCreateTable(
	ctx context.Context,
	input io.Reader,
	evalCtx *tree.EvalContext,
	p sql.JobExecContext,
	match string,
	parentDB catalog.DatabaseDescriptor,
	walltime int64,
	fks fkHandler,
	max int,
	owner security.SQLUsername,
	unsupportedStmtLogger *unsupportedStmtLogger,
) ([]*tabledesc.Mutable, []*schemadesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(496047)

	schemaObjects := schemaParsingObjects{
		createSchema: make(map[string]*tree.CreateSchema),
		createTbl:    make(map[schemaAndTableName]*tree.CreateTable),
		createSeq:    make(map[schemaAndTableName]*tree.CreateSequence),
		tableFKs:     make(map[schemaAndTableName][]*tree.ForeignKeyConstraintTableDef),
	}
	ps := newPostgreStream(ctx, input, max, unsupportedStmtLogger)
	for {
		__antithesis_instrumentation__.Notify(496057)
		stmt, err := ps.Next()
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(496060)
			break
		} else {
			__antithesis_instrumentation__.Notify(496061)
		}
		__antithesis_instrumentation__.Notify(496058)
		if err != nil {
			__antithesis_instrumentation__.Notify(496062)
			return nil, nil, errors.Wrap(err, "postgres parse error")
		} else {
			__antithesis_instrumentation__.Notify(496063)
		}
		__antithesis_instrumentation__.Notify(496059)
		if err := readPostgresStmt(ctx, evalCtx, match, fks, &schemaObjects, stmt, p,
			parentDB.GetID(), unsupportedStmtLogger); err != nil {
			__antithesis_instrumentation__.Notify(496064)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(496065)
		}
	}
	__antithesis_instrumentation__.Notify(496048)

	tables := make([]*tabledesc.Mutable, 0, len(schemaObjects.createTbl))
	schemaNameToDesc := make(map[string]*schemadesc.Mutable)
	schemaDescs, err := createPostgresSchemas(ctx, parentDB.GetID(), schemaObjects.createSchema,
		p.ExecCfg(), p.SessionData())
	if err != nil {
		__antithesis_instrumentation__.Notify(496066)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(496067)
	}
	__antithesis_instrumentation__.Notify(496049)

	for _, schemaDesc := range schemaDescs {
		__antithesis_instrumentation__.Notify(496068)
		schemaNameToDesc[schemaDesc.GetName()] = schemaDesc
	}
	__antithesis_instrumentation__.Notify(496050)

	if p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.PublicSchemasWithDescriptors) {
		__antithesis_instrumentation__.Notify(496069)

		publicSchema, err := getPublicSchemaDescForDatabase(ctx, p.ExecCfg(), parentDB)
		if err != nil {
			__antithesis_instrumentation__.Notify(496071)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(496072)
		}
		__antithesis_instrumentation__.Notify(496070)
		schemaNameToDesc[tree.PublicSchema] =
			schemadesc.NewBuilder(publicSchema.SchemaDesc()).BuildExistingMutableSchema()
	} else {
		__antithesis_instrumentation__.Notify(496073)
	}
	__antithesis_instrumentation__.Notify(496051)

	seqs, err := createPostgresSequences(
		ctx,
		parentDB.GetID(),
		schemaObjects.createSeq,
		fks,
		walltime,
		owner,
		schemaNameToDesc,
		p.ExecCfg(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(496074)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(496075)
	}
	__antithesis_instrumentation__.Notify(496052)
	tables = append(tables, seqs...)

	backrefs := make(map[descpb.ID]*tabledesc.Mutable)
	tableDescs, err := createPostgresTables(evalCtx, p, schemaObjects.createTbl, fks, backrefs,
		parentDB, walltime, schemaNameToDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(496076)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(496077)
	}
	__antithesis_instrumentation__.Notify(496053)
	tables = append(tables, tableDescs...)

	err = resolvePostgresFKs(
		evalCtx, parentDB, schemaObjects.tableFKs, fks, backrefs, schemaNameToDesc,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(496078)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(496079)
	}
	__antithesis_instrumentation__.Notify(496054)
	if match != "" && func() bool {
		__antithesis_instrumentation__.Notify(496080)
		return len(tables) != 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(496081)
		found := make([]string, 0, len(schemaObjects.createTbl))
		for schemaAndTableName := range schemaObjects.createTbl {
			__antithesis_instrumentation__.Notify(496083)
			found = append(found, schemaAndTableName.String())
		}
		__antithesis_instrumentation__.Notify(496082)
		return nil, nil, errors.Errorf("table %q not found in file (found tables: %s)", match,
			strings.Join(found, ", "))
	} else {
		__antithesis_instrumentation__.Notify(496084)
	}
	__antithesis_instrumentation__.Notify(496055)
	if len(tables) == 0 {
		__antithesis_instrumentation__.Notify(496085)
		return nil, nil, errors.Errorf("no table definition found")
	} else {
		__antithesis_instrumentation__.Notify(496086)
	}
	__antithesis_instrumentation__.Notify(496056)
	return tables, schemaDescs, nil
}

func readPostgresStmt(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	match string,
	fks fkHandler,
	schemaObjects *schemaParsingObjects,
	stmt interface{},
	p sql.JobExecContext,
	parentID descpb.ID,
	unsupportedStmtLogger *unsupportedStmtLogger,
) error {
	__antithesis_instrumentation__.Notify(496087)
	ignoreUnsupportedStmts := unsupportedStmtLogger.ignoreUnsupported
	switch stmt := stmt.(type) {
	case *tree.CreateSchema:
		__antithesis_instrumentation__.Notify(496089)
		name, err := getSchemaName(&stmt.Schema)
		if err != nil {
			__antithesis_instrumentation__.Notify(496120)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496121)
		}
		__antithesis_instrumentation__.Notify(496090)

		if match != "" {
			__antithesis_instrumentation__.Notify(496122)
			break
		} else {
			__antithesis_instrumentation__.Notify(496123)
		}
		__antithesis_instrumentation__.Notify(496091)
		schemaObjects.createSchema[name] = stmt
	case *tree.CreateTable:
		__antithesis_instrumentation__.Notify(496092)

		for _, def := range stmt.Defs {
			__antithesis_instrumentation__.Notify(496124)
			if d, ok := def.(*tree.ColumnTableDef); ok {
				__antithesis_instrumentation__.Notify(496125)
				if dType, ok := d.Type.(*types.T); ok {
					__antithesis_instrumentation__.Notify(496126)
					if dType.Equivalent(types.Int) {
						__antithesis_instrumentation__.Notify(496127)
						d.Type = parser.NakedIntTypeFromDefaultIntSize(p.SessionData().DefaultIntSize)
					} else {
						__antithesis_instrumentation__.Notify(496128)
					}
				} else {
					__antithesis_instrumentation__.Notify(496129)
				}
			} else {
				__antithesis_instrumentation__.Notify(496130)
			}
		}
		__antithesis_instrumentation__.Notify(496093)
		schemaQualifiedName, err := getSchemaAndTableName(&stmt.Table)
		if err != nil {
			__antithesis_instrumentation__.Notify(496131)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496132)
		}
		__antithesis_instrumentation__.Notify(496094)
		isMatch := match == "" || func() bool {
			__antithesis_instrumentation__.Notify(496133)
			return match == schemaQualifiedName.String() == true
		}() == true
		if isMatch {
			__antithesis_instrumentation__.Notify(496134)
			schemaObjects.createTbl[schemaQualifiedName] = stmt
		} else {
			__antithesis_instrumentation__.Notify(496135)
			schemaObjects.createTbl[schemaQualifiedName] = nil
		}
	case *tree.CreateIndex:
		__antithesis_instrumentation__.Notify(496095)
		if stmt.Predicate != nil {
			__antithesis_instrumentation__.Notify(496136)
			return unimplemented.NewWithIssue(50225, "cannot import a table with partial indexes")
		} else {
			__antithesis_instrumentation__.Notify(496137)
		}
		__antithesis_instrumentation__.Notify(496096)
		schemaQualifiedTableName, err := getSchemaAndTableName(&stmt.Table)
		if err != nil {
			__antithesis_instrumentation__.Notify(496138)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496139)
		}
		__antithesis_instrumentation__.Notify(496097)
		create := schemaObjects.createTbl[schemaQualifiedTableName]
		if create == nil {
			__antithesis_instrumentation__.Notify(496140)
			break
		} else {
			__antithesis_instrumentation__.Notify(496141)
		}
		__antithesis_instrumentation__.Notify(496098)
		var idx tree.TableDef = &tree.IndexTableDef{
			Name:             stmt.Name,
			Columns:          stmt.Columns,
			Storing:          stmt.Storing,
			Inverted:         stmt.Inverted,
			PartitionByIndex: stmt.PartitionByIndex,
			StorageParams:    stmt.StorageParams,
		}
		if stmt.Unique {
			__antithesis_instrumentation__.Notify(496142)
			idx = &tree.UniqueConstraintTableDef{IndexTableDef: *idx.(*tree.IndexTableDef)}
		} else {
			__antithesis_instrumentation__.Notify(496143)
		}
		__antithesis_instrumentation__.Notify(496099)
		create.Defs = append(create.Defs, idx)
	case *tree.AlterSchema:
		__antithesis_instrumentation__.Notify(496100)
		switch stmt.Cmd {
		default:
			__antithesis_instrumentation__.Notify(496144)
		}
	case *tree.AlterTable:
		__antithesis_instrumentation__.Notify(496101)
		schemaQualifiedTableName, err := getSchemaAndTableName2(stmt.Table)
		if err != nil {
			__antithesis_instrumentation__.Notify(496145)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496146)
		}
		__antithesis_instrumentation__.Notify(496102)
		create := schemaObjects.createTbl[schemaQualifiedTableName]
		if create == nil {
			__antithesis_instrumentation__.Notify(496147)
			break
		} else {
			__antithesis_instrumentation__.Notify(496148)
		}
		__antithesis_instrumentation__.Notify(496103)
		for _, cmd := range stmt.Cmds {
			__antithesis_instrumentation__.Notify(496149)
			switch cmd := cmd.(type) {
			case *tree.AlterTableAddConstraint:
				__antithesis_instrumentation__.Notify(496150)
				switch con := cmd.ConstraintDef.(type) {
				case *tree.ForeignKeyConstraintTableDef:
					__antithesis_instrumentation__.Notify(496161)
					if !fks.skip {
						__antithesis_instrumentation__.Notify(496163)
						if con.Table.Schema() == "" {
							__antithesis_instrumentation__.Notify(496165)
							con.Table.SchemaName = tree.PublicSchemaName
						} else {
							__antithesis_instrumentation__.Notify(496166)
						}
						__antithesis_instrumentation__.Notify(496164)
						schemaObjects.tableFKs[schemaQualifiedTableName] = append(schemaObjects.tableFKs[schemaQualifiedTableName], con)
					} else {
						__antithesis_instrumentation__.Notify(496167)
					}
				default:
					__antithesis_instrumentation__.Notify(496162)
					create.Defs = append(create.Defs, cmd.ConstraintDef)
				}
			case *tree.AlterTableSetDefault:
				__antithesis_instrumentation__.Notify(496151)
				found := false
				for i, def := range create.Defs {
					__antithesis_instrumentation__.Notify(496168)
					def, ok := def.(*tree.ColumnTableDef)

					if !ok || func() bool {
						__antithesis_instrumentation__.Notify(496170)
						return def.Name != cmd.Column == true
					}() == true {
						__antithesis_instrumentation__.Notify(496171)
						continue
					} else {
						__antithesis_instrumentation__.Notify(496172)
					}
					__antithesis_instrumentation__.Notify(496169)
					def.DefaultExpr.Expr = cmd.Default
					create.Defs[i] = def
					found = true
					break
				}
				__antithesis_instrumentation__.Notify(496152)
				if !found {
					__antithesis_instrumentation__.Notify(496173)
					return colinfo.NewUndefinedColumnError(cmd.Column.String())
				} else {
					__antithesis_instrumentation__.Notify(496174)
				}
			case *tree.AlterTableSetVisible:
				__antithesis_instrumentation__.Notify(496153)
				found := false
				for i, def := range create.Defs {
					__antithesis_instrumentation__.Notify(496175)
					def, ok := def.(*tree.ColumnTableDef)

					if !ok || func() bool {
						__antithesis_instrumentation__.Notify(496177)
						return def.Name != cmd.Column == true
					}() == true {
						__antithesis_instrumentation__.Notify(496178)
						continue
					} else {
						__antithesis_instrumentation__.Notify(496179)
					}
					__antithesis_instrumentation__.Notify(496176)
					def.Hidden = !cmd.Visible
					create.Defs[i] = def
					found = true
					break
				}
				__antithesis_instrumentation__.Notify(496154)
				if !found {
					__antithesis_instrumentation__.Notify(496180)
					return colinfo.NewUndefinedColumnError(cmd.Column.String())
				} else {
					__antithesis_instrumentation__.Notify(496181)
				}
			case *tree.AlterTableAddColumn:
				__antithesis_instrumentation__.Notify(496155)
				if cmd.IfNotExists {
					__antithesis_instrumentation__.Notify(496182)
					if ignoreUnsupportedStmts {
						__antithesis_instrumentation__.Notify(496184)
						err := unsupportedStmtLogger.log(stmt.String(), false)
						if err != nil {
							__antithesis_instrumentation__.Notify(496186)
							return err
						} else {
							__antithesis_instrumentation__.Notify(496187)
						}
						__antithesis_instrumentation__.Notify(496185)
						continue
					} else {
						__antithesis_instrumentation__.Notify(496188)
					}
					__antithesis_instrumentation__.Notify(496183)
					return wrapErrorWithUnsupportedHint(errors.Errorf("unsupported statement: %s", stmt))
				} else {
					__antithesis_instrumentation__.Notify(496189)
				}
				__antithesis_instrumentation__.Notify(496156)
				create.Defs = append(create.Defs, cmd.ColumnDef)
			case *tree.AlterTableSetNotNull:
				__antithesis_instrumentation__.Notify(496157)
				found := false
				for i, def := range create.Defs {
					__antithesis_instrumentation__.Notify(496190)
					def, ok := def.(*tree.ColumnTableDef)

					if !ok || func() bool {
						__antithesis_instrumentation__.Notify(496192)
						return def.Name != cmd.Column == true
					}() == true {
						__antithesis_instrumentation__.Notify(496193)
						continue
					} else {
						__antithesis_instrumentation__.Notify(496194)
					}
					__antithesis_instrumentation__.Notify(496191)
					def.Nullable.Nullability = tree.NotNull
					create.Defs[i] = def
					found = true
					break
				}
				__antithesis_instrumentation__.Notify(496158)
				if !found {
					__antithesis_instrumentation__.Notify(496195)
					return colinfo.NewUndefinedColumnError(cmd.Column.String())
				} else {
					__antithesis_instrumentation__.Notify(496196)
				}
			default:
				__antithesis_instrumentation__.Notify(496159)
				if ignoreUnsupportedStmts {
					__antithesis_instrumentation__.Notify(496197)
					err := unsupportedStmtLogger.log(stmt.String(), false)
					if err != nil {
						__antithesis_instrumentation__.Notify(496199)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496200)
					}
					__antithesis_instrumentation__.Notify(496198)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496201)
				}
				__antithesis_instrumentation__.Notify(496160)
				return wrapErrorWithUnsupportedHint(errors.Errorf("unsupported statement: %s", stmt))
			}
		}
	case *tree.AlterTableOwner:
		__antithesis_instrumentation__.Notify(496104)
		if ignoreUnsupportedStmts {
			__antithesis_instrumentation__.Notify(496202)
			return unsupportedStmtLogger.log(stmt.String(), false)
		} else {
			__antithesis_instrumentation__.Notify(496203)
		}
		__antithesis_instrumentation__.Notify(496105)
		return wrapErrorWithUnsupportedHint(errors.Errorf("unsupported statement: %s", stmt))
	case *tree.CreateSequence:
		__antithesis_instrumentation__.Notify(496106)
		schemaQualifiedTableName, err := getSchemaAndTableName(&stmt.Name)
		if err != nil {
			__antithesis_instrumentation__.Notify(496204)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496205)
		}
		__antithesis_instrumentation__.Notify(496107)
		if match == "" || func() bool {
			__antithesis_instrumentation__.Notify(496206)
			return match == schemaQualifiedTableName.String() == true
		}() == true {
			__antithesis_instrumentation__.Notify(496207)
			schemaObjects.createSeq[schemaQualifiedTableName] = stmt
		} else {
			__antithesis_instrumentation__.Notify(496208)
		}
	case *tree.AlterSequence:
		__antithesis_instrumentation__.Notify(496108)
		if ignoreUnsupportedStmts {
			__antithesis_instrumentation__.Notify(496209)
			return unsupportedStmtLogger.log(stmt.String(), false)
		} else {
			__antithesis_instrumentation__.Notify(496210)
		}
		__antithesis_instrumentation__.Notify(496109)
		return wrapErrorWithUnsupportedHint(errors.Errorf("unsupported %T statement: %s", stmt, stmt))

	case *tree.Select:
		__antithesis_instrumentation__.Notify(496110)
		switch sel := stmt.Select.(type) {
		case *tree.SelectClause:
			__antithesis_instrumentation__.Notify(496211)
			for _, selExpr := range sel.Exprs {
				__antithesis_instrumentation__.Notify(496214)
				switch expr := selExpr.Expr.(type) {
				case *tree.FuncExpr:
					__antithesis_instrumentation__.Notify(496215)

					semaCtx := tree.MakeSemaContext()
					if _, err := expr.TypeCheck(ctx, &semaCtx, nil); err != nil {
						__antithesis_instrumentation__.Notify(496223)

						if f := expr.Func.String(); pgerror.GetPGCode(err) == pgcode.UndefinedColumn && func() bool {
							__antithesis_instrumentation__.Notify(496225)
							return f == "setval" == true
						}() == true {
							__antithesis_instrumentation__.Notify(496226)
							continue
						} else {
							__antithesis_instrumentation__.Notify(496227)
						}
						__antithesis_instrumentation__.Notify(496224)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496228)
					}
					__antithesis_instrumentation__.Notify(496216)
					ov := expr.ResolvedOverload()

					fn := ov.SQLFn
					if fn == nil {
						__antithesis_instrumentation__.Notify(496229)
						err := errors.Errorf("unsupported function call: %s in stmt: %s",
							expr.Func.String(), stmt.String())
						if ignoreUnsupportedStmts {
							__antithesis_instrumentation__.Notify(496231)
							err := unsupportedStmtLogger.log(err.Error(), false)
							if err != nil {
								__antithesis_instrumentation__.Notify(496233)
								return err
							} else {
								__antithesis_instrumentation__.Notify(496234)
							}
							__antithesis_instrumentation__.Notify(496232)
							continue
						} else {
							__antithesis_instrumentation__.Notify(496235)
						}
						__antithesis_instrumentation__.Notify(496230)
						return wrapErrorWithUnsupportedHint(err)
					} else {
						__antithesis_instrumentation__.Notify(496236)
					}
					__antithesis_instrumentation__.Notify(496217)

					datums := make(tree.Datums, len(expr.Exprs))
					for i, ex := range expr.Exprs {
						__antithesis_instrumentation__.Notify(496237)
						d, ok := ex.(tree.Datum)
						if !ok {
							__antithesis_instrumentation__.Notify(496239)

							return errors.Errorf("unsupported statement: %s", stmt)
						} else {
							__antithesis_instrumentation__.Notify(496240)
						}
						__antithesis_instrumentation__.Notify(496238)
						datums[i] = d
					}
					__antithesis_instrumentation__.Notify(496218)

					fnSQL, err := fn(evalCtx, datums)
					if err != nil {
						__antithesis_instrumentation__.Notify(496241)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496242)
					}
					__antithesis_instrumentation__.Notify(496219)

					fnStmts, err := parser.Parse(fnSQL)
					if err != nil {
						__antithesis_instrumentation__.Notify(496243)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496244)
					}
					__antithesis_instrumentation__.Notify(496220)
					for _, fnStmt := range fnStmts {
						__antithesis_instrumentation__.Notify(496245)
						switch ast := fnStmt.AST.(type) {
						case *tree.AlterTable:
							__antithesis_instrumentation__.Notify(496246)
							if err := readPostgresStmt(ctx, evalCtx, match, fks, schemaObjects, ast, p,
								parentID, unsupportedStmtLogger); err != nil {
								__antithesis_instrumentation__.Notify(496248)
								return err
							} else {
								__antithesis_instrumentation__.Notify(496249)
							}
						default:
							__antithesis_instrumentation__.Notify(496247)

							return errors.Errorf("unsupported statement: %s", stmt)
						}
					}
				default:
					__antithesis_instrumentation__.Notify(496221)
					err := errors.Errorf("unsupported %T SELECT expr: %s", expr, expr)
					if ignoreUnsupportedStmts {
						__antithesis_instrumentation__.Notify(496250)
						err := unsupportedStmtLogger.log(err.Error(), false)
						if err != nil {
							__antithesis_instrumentation__.Notify(496252)
							return err
						} else {
							__antithesis_instrumentation__.Notify(496253)
						}
						__antithesis_instrumentation__.Notify(496251)
						continue
					} else {
						__antithesis_instrumentation__.Notify(496254)
					}
					__antithesis_instrumentation__.Notify(496222)
					return wrapErrorWithUnsupportedHint(err)
				}
			}
		default:
			__antithesis_instrumentation__.Notify(496212)
			err := errors.Errorf("unsupported %T SELECT %s", sel, sel)
			if ignoreUnsupportedStmts {
				__antithesis_instrumentation__.Notify(496255)
				err := unsupportedStmtLogger.log(err.Error(), false)
				if err != nil {
					__antithesis_instrumentation__.Notify(496257)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496258)
				}
				__antithesis_instrumentation__.Notify(496256)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(496259)
			}
			__antithesis_instrumentation__.Notify(496213)
			return wrapErrorWithUnsupportedHint(err)
		}
	case *tree.DropTable:
		__antithesis_instrumentation__.Notify(496111)
		names := stmt.Names

		for _, name := range names {
			__antithesis_instrumentation__.Notify(496260)
			tableName := name.ToUnresolvedObjectName().String()
			if err := sql.DescsTxn(ctx, p.ExecCfg(), func(ctx context.Context, txn *kv.Txn, col *descs.Collection) error {
				__antithesis_instrumentation__.Notify(496261)
				dbDesc, err := col.Direct().MustGetDatabaseDescByID(ctx, txn, parentID)
				if err != nil {
					__antithesis_instrumentation__.Notify(496264)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496265)
				}
				__antithesis_instrumentation__.Notify(496262)
				err = col.Direct().CheckObjectCollision(
					ctx,
					txn,
					parentID,
					dbDesc.GetSchemaID(tree.PublicSchema),
					tree.NewUnqualifiedTableName(tree.Name(tableName)),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(496266)
					return errors.Wrapf(err, `drop table "%s" and then retry the import`, tableName)
				} else {
					__antithesis_instrumentation__.Notify(496267)
				}
				__antithesis_instrumentation__.Notify(496263)
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(496268)
				return err
			} else {
				__antithesis_instrumentation__.Notify(496269)
			}
		}
	case *tree.BeginTransaction, *tree.CommitTransaction:
		__antithesis_instrumentation__.Notify(496112)

	case *tree.Insert, *tree.CopyFrom, *tree.Delete, copyData:
		__antithesis_instrumentation__.Notify(496113)

	case *tree.CreateExtension, *tree.CommentOnDatabase, *tree.CommentOnTable,
		*tree.CommentOnIndex, *tree.CommentOnConstraint, *tree.CommentOnColumn, *tree.SetVar, *tree.Analyze,
		*tree.CommentOnSchema:
		__antithesis_instrumentation__.Notify(496114)

		if ignoreUnsupportedStmts {
			__antithesis_instrumentation__.Notify(496270)
			return unsupportedStmtLogger.log(fmt.Sprintf("%s", stmt), false)
		} else {
			__antithesis_instrumentation__.Notify(496271)
		}
		__antithesis_instrumentation__.Notify(496115)
		return wrapErrorWithUnsupportedHint(errors.Errorf("unsupported %T statement: %s", stmt, stmt))
	case *tree.CreateType:
		__antithesis_instrumentation__.Notify(496116)
		return errors.New("IMPORT PGDUMP does not support user defined types; please" +
			" remove all CREATE TYPE statements and their usages from the dump file")
	case error:
		__antithesis_instrumentation__.Notify(496117)
		if !errors.Is(stmt, errCopyDone) {
			__antithesis_instrumentation__.Notify(496272)
			return stmt
		} else {
			__antithesis_instrumentation__.Notify(496273)
		}
	default:
		__antithesis_instrumentation__.Notify(496118)
		if ignoreUnsupportedStmts {
			__antithesis_instrumentation__.Notify(496274)
			return unsupportedStmtLogger.log(fmt.Sprintf("%s", stmt), false)
		} else {
			__antithesis_instrumentation__.Notify(496275)
		}
		__antithesis_instrumentation__.Notify(496119)
		return wrapErrorWithUnsupportedHint(errors.Errorf("unsupported %T statement: %s", stmt, stmt))
	}
	__antithesis_instrumentation__.Notify(496088)
	return nil
}

func getSchemaName(sc *tree.ObjectNamePrefix) (string, error) {
	__antithesis_instrumentation__.Notify(496276)
	if sc.ExplicitCatalog {
		__antithesis_instrumentation__.Notify(496278)
		return "", unimplemented.Newf("import into database specified in dump file",
			"explicit catalog schemas unsupported: %s", sc.CatalogName.String()+sc.SchemaName.String())
	} else {
		__antithesis_instrumentation__.Notify(496279)
	}
	__antithesis_instrumentation__.Notify(496277)
	return sc.SchemaName.String(), nil
}

func getSchemaAndTableName(tn *tree.TableName) (schemaAndTableName, error) {
	__antithesis_instrumentation__.Notify(496280)
	var ret schemaAndTableName
	ret.schema = tree.PublicSchema
	if tn.Schema() != "" {
		__antithesis_instrumentation__.Notify(496282)
		ret.schema = tn.Schema()
	} else {
		__antithesis_instrumentation__.Notify(496283)
	}
	__antithesis_instrumentation__.Notify(496281)
	ret.table = tn.Table()
	return ret, nil
}

func getSchemaAndTableName2(u *tree.UnresolvedObjectName) (schemaAndTableName, error) {
	__antithesis_instrumentation__.Notify(496284)
	var ret schemaAndTableName
	ret.schema = tree.PublicSchema
	if u.NumParts >= 2 && func() bool {
		__antithesis_instrumentation__.Notify(496286)
		return u.Parts[1] != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(496287)
		ret.schema = u.Parts[1]
	} else {
		__antithesis_instrumentation__.Notify(496288)
	}
	__antithesis_instrumentation__.Notify(496285)
	ret.table = u.Parts[0]
	return ret, nil
}

type pgDumpReader struct {
	tableDescs            map[string]catalog.TableDescriptor
	tables                map[string]*row.DatumRowConverter
	descs                 map[string]*execinfrapb.ReadImportDataSpec_ImportTable
	kvCh                  chan row.KVBatch
	opts                  roachpb.PgDumpOptions
	walltime              int64
	colMap                map[*row.DatumRowConverter](map[string]int)
	jobID                 int64
	unsupportedStmtLogger *unsupportedStmtLogger
	evalCtx               *tree.EvalContext
}

var _ inputConverter = &pgDumpReader{}

func newPgDumpReader(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	jobID int64,
	kvCh chan row.KVBatch,
	opts roachpb.PgDumpOptions,
	walltime int64,
	descs map[string]*execinfrapb.ReadImportDataSpec_ImportTable,
	evalCtx *tree.EvalContext,
) (*pgDumpReader, error) {
	__antithesis_instrumentation__.Notify(496289)
	tableDescs := make(map[string]catalog.TableDescriptor, len(descs))
	converters := make(map[string]*row.DatumRowConverter, len(descs))
	colMap := make(map[*row.DatumRowConverter](map[string]int))
	for name, table := range descs {
		__antithesis_instrumentation__.Notify(496291)
		if table.Desc.IsTable() {
			__antithesis_instrumentation__.Notify(496292)
			tableDesc := tabledesc.NewBuilder(table.Desc).BuildImmutableTable()
			colSubMap := make(map[string]int, len(table.TargetCols))
			targetCols := make(tree.NameList, len(table.TargetCols))
			for i, colName := range table.TargetCols {
				__antithesis_instrumentation__.Notify(496296)
				targetCols[i] = tree.Name(colName)
			}
			__antithesis_instrumentation__.Notify(496293)
			for i, col := range tableDesc.VisibleColumns() {
				__antithesis_instrumentation__.Notify(496297)
				colSubMap[col.GetName()] = i
			}
			__antithesis_instrumentation__.Notify(496294)
			conv, err := row.NewDatumRowConverter(ctx, semaCtx, tableDesc, targetCols, evalCtx, kvCh,
				nil, nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(496298)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(496299)
			}
			__antithesis_instrumentation__.Notify(496295)
			converters[name] = conv
			colMap[conv] = colSubMap
			tableDescs[name] = tableDesc
		} else {
			__antithesis_instrumentation__.Notify(496300)
			if table.Desc.IsSequence() {
				__antithesis_instrumentation__.Notify(496301)
				seqDesc := tabledesc.NewBuilder(table.Desc).BuildImmutableTable()
				tableDescs[name] = seqDesc
			} else {
				__antithesis_instrumentation__.Notify(496302)
			}
		}
	}
	__antithesis_instrumentation__.Notify(496290)
	return &pgDumpReader{
		kvCh:       kvCh,
		tableDescs: tableDescs,
		tables:     converters,
		descs:      descs,
		opts:       opts,
		walltime:   walltime,
		colMap:     colMap,
		jobID:      jobID,
		evalCtx:    evalCtx,
	}, nil
}

func (m *pgDumpReader) start(ctx ctxgroup.Group) {
	__antithesis_instrumentation__.Notify(496303)
}

func (m *pgDumpReader) readFiles(
	ctx context.Context,
	dataFiles map[int32]string,
	resumePos map[int32]int64,
	format roachpb.IOFileFormat,
	makeExternalStorage cloud.ExternalStorageFactory,
	user security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(496304)

	m.unsupportedStmtLogger = makeUnsupportedStmtLogger(ctx, user,
		m.jobID, format.PgDump.IgnoreUnsupported, format.PgDump.IgnoreUnsupportedLog, dataIngestion,
		makeExternalStorage)

	err := readInputFiles(ctx, dataFiles, resumePos, format, m.readFile, makeExternalStorage, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(496306)
		return err
	} else {
		__antithesis_instrumentation__.Notify(496307)
	}
	__antithesis_instrumentation__.Notify(496305)

	return m.unsupportedStmtLogger.flush()
}

func wrapErrorWithUnsupportedHint(err error) error {
	__antithesis_instrumentation__.Notify(496308)
	return errors.WithHintf(err,
		"To ignore unsupported statements and log them for review post IMPORT, see the options listed"+
			" in the docs: %s", "https://www.cockroachlabs.com/docs/stable/import.html#import-options")
}

func (m *pgDumpReader) readFile(
	ctx context.Context, input *fileReader, inputIdx int32, resumePos int64, rejected chan string,
) error {
	__antithesis_instrumentation__.Notify(496309)
	tableNameToRowsProcessed := make(map[string]int64)
	var inserts, count int64
	rowLimit := m.opts.RowLimit
	ps := newPostgreStream(ctx, input, int(m.opts.MaxRowSize), m.unsupportedStmtLogger)
	semaCtx := tree.MakeSemaContext()
	for _, conv := range m.tables {
		__antithesis_instrumentation__.Notify(496313)
		conv.KvBatch.Source = inputIdx
		conv.FractionFn = input.ReadFraction
		conv.CompletedRowFn = func() int64 {
			__antithesis_instrumentation__.Notify(496314)
			return count
		}
	}
	__antithesis_instrumentation__.Notify(496310)

	for {
		__antithesis_instrumentation__.Notify(496315)
		stmt, err := ps.Next()
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(496318)
			break
		} else {
			__antithesis_instrumentation__.Notify(496319)
		}
		__antithesis_instrumentation__.Notify(496316)
		if err != nil {
			__antithesis_instrumentation__.Notify(496320)
			return errors.Wrap(err, "postgres parse error")
		} else {
			__antithesis_instrumentation__.Notify(496321)
		}
		__antithesis_instrumentation__.Notify(496317)
		switch i := stmt.(type) {
		case *tree.Insert:
			__antithesis_instrumentation__.Notify(496322)
			n, ok := i.Table.(*tree.TableName)
			if !ok {
				__antithesis_instrumentation__.Notify(496344)
				return errors.Errorf("unexpected: %T", i.Table)
			} else {
				__antithesis_instrumentation__.Notify(496345)
			}
			__antithesis_instrumentation__.Notify(496323)
			name, err := getSchemaAndTableName(n)
			if err != nil {
				__antithesis_instrumentation__.Notify(496346)
				return errors.Wrapf(err, "%s", i)
			} else {
				__antithesis_instrumentation__.Notify(496347)
			}
			__antithesis_instrumentation__.Notify(496324)
			conv, ok := m.tables[name.String()]
			if !ok {
				__antithesis_instrumentation__.Notify(496348)

				continue
			} else {
				__antithesis_instrumentation__.Notify(496349)
			}
			__antithesis_instrumentation__.Notify(496325)
			if ok && func() bool {
				__antithesis_instrumentation__.Notify(496350)
				return conv == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(496351)
				return errors.Errorf("missing schema info for requested table %q", name)
			} else {
				__antithesis_instrumentation__.Notify(496352)
			}
			__antithesis_instrumentation__.Notify(496326)
			expectedColLen := len(i.Columns)
			if expectedColLen == 0 {
				__antithesis_instrumentation__.Notify(496353)

				expectedColLen = len(conv.VisibleCols)
			} else {
				__antithesis_instrumentation__.Notify(496354)
			}
			__antithesis_instrumentation__.Notify(496327)
			timestamp := timestampAfterEpoch(m.walltime)
			values, ok := i.Rows.Select.(*tree.ValuesClause)
			if !ok {
				__antithesis_instrumentation__.Notify(496355)
				if m.unsupportedStmtLogger.ignoreUnsupported {
					__antithesis_instrumentation__.Notify(496357)
					logLine := fmt.Sprintf("%s: unsupported by IMPORT\n",
						i.Rows.Select.String())
					err := m.unsupportedStmtLogger.log(logLine, false)
					if err != nil {
						__antithesis_instrumentation__.Notify(496359)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496360)
					}
					__antithesis_instrumentation__.Notify(496358)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496361)
				}
				__antithesis_instrumentation__.Notify(496356)
				return wrapErrorWithUnsupportedHint(errors.Errorf("unsupported: %s", i.Rows.Select))
			} else {
				__antithesis_instrumentation__.Notify(496362)
			}
			__antithesis_instrumentation__.Notify(496328)
			inserts++
			startingCount := count
			var targetColMapIdx []int
			if len(i.Columns) != 0 {
				__antithesis_instrumentation__.Notify(496363)
				targetColMapIdx = make([]int, len(i.Columns))
				conv.TargetColOrds = util.FastIntSet{}
				for j := range i.Columns {
					__antithesis_instrumentation__.Notify(496365)
					colName := string(i.Columns[j])
					idx, ok := m.colMap[conv][colName]
					if !ok {
						__antithesis_instrumentation__.Notify(496367)
						return errors.Newf("targeted column %q not found", colName)
					} else {
						__antithesis_instrumentation__.Notify(496368)
					}
					__antithesis_instrumentation__.Notify(496366)
					conv.TargetColOrds.Add(idx)
					targetColMapIdx[j] = idx
				}
				__antithesis_instrumentation__.Notify(496364)

				for idx := range conv.VisibleCols {
					__antithesis_instrumentation__.Notify(496369)
					if !conv.TargetColOrds.Contains(idx) {
						__antithesis_instrumentation__.Notify(496370)
						conv.Datums[idx] = tree.DNull
					} else {
						__antithesis_instrumentation__.Notify(496371)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(496372)
			}
			__antithesis_instrumentation__.Notify(496329)
			for _, tuple := range values.Rows {
				__antithesis_instrumentation__.Notify(496373)
				count++
				tableNameToRowsProcessed[name.String()]++
				if count <= resumePos {
					__antithesis_instrumentation__.Notify(496378)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496379)
				}
				__antithesis_instrumentation__.Notify(496374)
				if rowLimit != 0 && func() bool {
					__antithesis_instrumentation__.Notify(496380)
					return tableNameToRowsProcessed[name.String()] > rowLimit == true
				}() == true {
					__antithesis_instrumentation__.Notify(496381)
					break
				} else {
					__antithesis_instrumentation__.Notify(496382)
				}
				__antithesis_instrumentation__.Notify(496375)
				if got := len(tuple); expectedColLen != got {
					__antithesis_instrumentation__.Notify(496383)
					return errors.Errorf("expected %d values, got %d: %v", expectedColLen, got, tuple)
				} else {
					__antithesis_instrumentation__.Notify(496384)
				}
				__antithesis_instrumentation__.Notify(496376)
				for j, expr := range tuple {
					__antithesis_instrumentation__.Notify(496385)
					idx := j
					if len(i.Columns) != 0 {
						__antithesis_instrumentation__.Notify(496389)
						idx = targetColMapIdx[j]
					} else {
						__antithesis_instrumentation__.Notify(496390)
					}
					__antithesis_instrumentation__.Notify(496386)
					typed, err := expr.TypeCheck(ctx, &semaCtx, conv.VisibleColTypes[idx])
					if err != nil {
						__antithesis_instrumentation__.Notify(496391)
						return errors.Wrapf(err, "reading row %d (%d in insert statement %d)",
							count, count-startingCount, inserts)
					} else {
						__antithesis_instrumentation__.Notify(496392)
					}
					__antithesis_instrumentation__.Notify(496387)
					converted, err := typed.Eval(conv.EvalCtx)
					if err != nil {
						__antithesis_instrumentation__.Notify(496393)
						return errors.Wrapf(err, "reading row %d (%d in insert statement %d)",
							count, count-startingCount, inserts)
					} else {
						__antithesis_instrumentation__.Notify(496394)
					}
					__antithesis_instrumentation__.Notify(496388)
					conv.Datums[idx] = converted
				}
				__antithesis_instrumentation__.Notify(496377)
				if err := conv.Row(ctx, inputIdx, count+int64(timestamp)); err != nil {
					__antithesis_instrumentation__.Notify(496395)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496396)
				}
			}
		case *tree.CopyFrom:
			__antithesis_instrumentation__.Notify(496330)
			if !i.Stdin {
				__antithesis_instrumentation__.Notify(496397)
				return errors.New("expected STDIN option on COPY FROM")
			} else {
				__antithesis_instrumentation__.Notify(496398)
			}
			__antithesis_instrumentation__.Notify(496331)
			name, err := getSchemaAndTableName(&i.Table)
			if err != nil {
				__antithesis_instrumentation__.Notify(496399)
				return errors.Wrapf(err, "%s", i)
			} else {
				__antithesis_instrumentation__.Notify(496400)
			}
			__antithesis_instrumentation__.Notify(496332)
			conv, importing := m.tables[name.String()]
			if importing && func() bool {
				__antithesis_instrumentation__.Notify(496401)
				return conv == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(496402)
				return errors.Errorf("missing schema info for requested table %q", name)
			} else {
				__antithesis_instrumentation__.Notify(496403)
			}
			__antithesis_instrumentation__.Notify(496333)
			var targetColMapIdx []int
			if conv != nil {
				__antithesis_instrumentation__.Notify(496404)
				targetColMapIdx = make([]int, len(i.Columns))
				conv.TargetColOrds = util.FastIntSet{}
				for j := range i.Columns {
					__antithesis_instrumentation__.Notify(496406)
					colName := string(i.Columns[j])
					idx, ok := m.colMap[conv][colName]
					if !ok {
						__antithesis_instrumentation__.Notify(496408)
						return errors.Newf("targeted column %q not found", colName)
					} else {
						__antithesis_instrumentation__.Notify(496409)
					}
					__antithesis_instrumentation__.Notify(496407)
					conv.TargetColOrds.Add(idx)
					targetColMapIdx[j] = idx
				}
				__antithesis_instrumentation__.Notify(496405)

				for idx := range conv.VisibleCols {
					__antithesis_instrumentation__.Notify(496410)
					if !conv.TargetColOrds.Contains(idx) {
						__antithesis_instrumentation__.Notify(496411)
						conv.Datums[idx] = tree.DNull
					} else {
						__antithesis_instrumentation__.Notify(496412)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(496413)
			}
			__antithesis_instrumentation__.Notify(496334)
			for {
				__antithesis_instrumentation__.Notify(496414)
				row, err := ps.Next()

				if err == io.EOF {
					__antithesis_instrumentation__.Notify(496420)
					return makeRowErr(count, pgcode.ProtocolViolation,
						"unexpected EOF")
				} else {
					__antithesis_instrumentation__.Notify(496421)
				}
				__antithesis_instrumentation__.Notify(496415)
				if row == errCopyDone {
					__antithesis_instrumentation__.Notify(496422)
					break
				} else {
					__antithesis_instrumentation__.Notify(496423)
				}
				__antithesis_instrumentation__.Notify(496416)
				count++
				tableNameToRowsProcessed[name.String()]++
				if err != nil {
					__antithesis_instrumentation__.Notify(496424)
					return wrapRowErr(err, count, pgcode.Uncategorized, "")
				} else {
					__antithesis_instrumentation__.Notify(496425)
				}
				__antithesis_instrumentation__.Notify(496417)
				if !importing {
					__antithesis_instrumentation__.Notify(496426)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496427)
				}
				__antithesis_instrumentation__.Notify(496418)
				if count <= resumePos {
					__antithesis_instrumentation__.Notify(496428)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496429)
				}
				__antithesis_instrumentation__.Notify(496419)
				switch row := row.(type) {
				case copyData:
					__antithesis_instrumentation__.Notify(496430)
					if expected, got := conv.TargetColOrds.Len(), len(row); expected != got {
						__antithesis_instrumentation__.Notify(496435)
						return makeRowErr(count, pgcode.Syntax,
							"expected %d values, got %d", expected, got)
					} else {
						__antithesis_instrumentation__.Notify(496436)
					}
					__antithesis_instrumentation__.Notify(496431)
					if rowLimit != 0 && func() bool {
						__antithesis_instrumentation__.Notify(496437)
						return tableNameToRowsProcessed[name.String()] > rowLimit == true
					}() == true {
						__antithesis_instrumentation__.Notify(496438)
						break
					} else {
						__antithesis_instrumentation__.Notify(496439)
					}
					__antithesis_instrumentation__.Notify(496432)
					for i, s := range row {
						__antithesis_instrumentation__.Notify(496440)
						idx := targetColMapIdx[i]
						if s == nil {
							__antithesis_instrumentation__.Notify(496441)
							conv.Datums[idx] = tree.DNull
						} else {
							__antithesis_instrumentation__.Notify(496442)

							conv.Datums[idx], _, err = tree.ParseAndRequireString(conv.VisibleColTypes[idx], *s, conv.EvalCtx)
							if err != nil {
								__antithesis_instrumentation__.Notify(496443)
								col := conv.VisibleCols[idx]
								return wrapRowErr(err, count, pgcode.Syntax,
									"parse %q as %s", col.GetName(), col.GetType().SQLString())
							} else {
								__antithesis_instrumentation__.Notify(496444)
							}
						}
					}
					__antithesis_instrumentation__.Notify(496433)
					if err := conv.Row(ctx, inputIdx, count); err != nil {
						__antithesis_instrumentation__.Notify(496445)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496446)
					}
				default:
					__antithesis_instrumentation__.Notify(496434)
					return makeRowErr(count, pgcode.Uncategorized,
						"unexpected: %v", row)
				}
			}
		case *tree.Select:
			__antithesis_instrumentation__.Notify(496335)

			sc, ok := i.Select.(*tree.SelectClause)
			if !ok {
				__antithesis_instrumentation__.Notify(496447)
				err := errors.Errorf("unsupported %T Select: %v", i.Select, i.Select)
				if m.unsupportedStmtLogger.ignoreUnsupported {
					__antithesis_instrumentation__.Notify(496449)
					err := m.unsupportedStmtLogger.log(err.Error(), false)
					if err != nil {
						__antithesis_instrumentation__.Notify(496451)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496452)
					}
					__antithesis_instrumentation__.Notify(496450)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496453)
				}
				__antithesis_instrumentation__.Notify(496448)
				return wrapErrorWithUnsupportedHint(err)
			} else {
				__antithesis_instrumentation__.Notify(496454)
			}
			__antithesis_instrumentation__.Notify(496336)
			if len(sc.Exprs) != 1 {
				__antithesis_instrumentation__.Notify(496455)
				err := errors.Errorf("unsupported %d select args: %v", len(sc.Exprs), sc.Exprs)
				if m.unsupportedStmtLogger.ignoreUnsupported {
					__antithesis_instrumentation__.Notify(496457)
					err := m.unsupportedStmtLogger.log(err.Error(), false)
					if err != nil {
						__antithesis_instrumentation__.Notify(496459)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496460)
					}
					__antithesis_instrumentation__.Notify(496458)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496461)
				}
				__antithesis_instrumentation__.Notify(496456)
				return wrapErrorWithUnsupportedHint(err)
			} else {
				__antithesis_instrumentation__.Notify(496462)
			}
			__antithesis_instrumentation__.Notify(496337)
			fn, ok := sc.Exprs[0].Expr.(*tree.FuncExpr)
			if !ok {
				__antithesis_instrumentation__.Notify(496463)
				err := errors.Errorf("unsupported select arg %T: %v", sc.Exprs[0].Expr, sc.Exprs[0].Expr)
				if m.unsupportedStmtLogger.ignoreUnsupported {
					__antithesis_instrumentation__.Notify(496465)
					err := m.unsupportedStmtLogger.log(err.Error(), false)
					if err != nil {
						__antithesis_instrumentation__.Notify(496467)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496468)
					}
					__antithesis_instrumentation__.Notify(496466)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496469)
				}
				__antithesis_instrumentation__.Notify(496464)
				return wrapErrorWithUnsupportedHint(err)
			} else {
				__antithesis_instrumentation__.Notify(496470)
			}
			__antithesis_instrumentation__.Notify(496338)

			switch funcName := strings.ToLower(fn.Func.String()); funcName {
			case "search_path", "pg_catalog.set_config":
				__antithesis_instrumentation__.Notify(496471)
				err := errors.Errorf("unsupported %d fn args in select: %v", len(fn.Exprs), fn.Exprs)
				if m.unsupportedStmtLogger.ignoreUnsupported {
					__antithesis_instrumentation__.Notify(496486)
					err := m.unsupportedStmtLogger.log(err.Error(), false)
					if err != nil {
						__antithesis_instrumentation__.Notify(496488)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496489)
					}
					__antithesis_instrumentation__.Notify(496487)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496490)
				}
				__antithesis_instrumentation__.Notify(496472)
				return wrapErrorWithUnsupportedHint(err)
			case "setval", "pg_catalog.setval":
				__antithesis_instrumentation__.Notify(496473)
				if args := len(fn.Exprs); args < 2 || func() bool {
					__antithesis_instrumentation__.Notify(496491)
					return args > 3 == true
				}() == true {
					__antithesis_instrumentation__.Notify(496492)
					err := errors.Errorf("unsupported %d fn args in select: %v", len(fn.Exprs), fn.Exprs)
					if m.unsupportedStmtLogger.ignoreUnsupported {
						__antithesis_instrumentation__.Notify(496494)
						err := m.unsupportedStmtLogger.log(err.Error(), false)
						if err != nil {
							__antithesis_instrumentation__.Notify(496496)
							return err
						} else {
							__antithesis_instrumentation__.Notify(496497)
						}
						__antithesis_instrumentation__.Notify(496495)
						continue
					} else {
						__antithesis_instrumentation__.Notify(496498)
					}
					__antithesis_instrumentation__.Notify(496493)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496499)
				}
				__antithesis_instrumentation__.Notify(496474)
				seqname, ok := fn.Exprs[0].(*tree.StrVal)
				if !ok {
					__antithesis_instrumentation__.Notify(496500)
					if nested, nestedOk := fn.Exprs[0].(*tree.FuncExpr); nestedOk && func() bool {
						__antithesis_instrumentation__.Notify(496502)
						return nested.Func.String() == "pg_get_serial_sequence" == true
					}() == true {
						__antithesis_instrumentation__.Notify(496503)

						continue
					} else {
						__antithesis_instrumentation__.Notify(496504)
					}
					__antithesis_instrumentation__.Notify(496501)
					return errors.Errorf("unsupported setval %T arg: %v", fn.Exprs[0], fn.Exprs[0])
				} else {
					__antithesis_instrumentation__.Notify(496505)
				}
				__antithesis_instrumentation__.Notify(496475)
				seqval, ok := fn.Exprs[1].(*tree.NumVal)
				if !ok {
					__antithesis_instrumentation__.Notify(496506)
					err := errors.Errorf("unsupported setval %T arg: %v", fn.Exprs[1], fn.Exprs[1])
					if m.unsupportedStmtLogger.ignoreUnsupported {
						__antithesis_instrumentation__.Notify(496508)
						err := m.unsupportedStmtLogger.log(err.Error(), false)
						if err != nil {
							__antithesis_instrumentation__.Notify(496510)
							return err
						} else {
							__antithesis_instrumentation__.Notify(496511)
						}
						__antithesis_instrumentation__.Notify(496509)
						continue
					} else {
						__antithesis_instrumentation__.Notify(496512)
					}
					__antithesis_instrumentation__.Notify(496507)
					return wrapErrorWithUnsupportedHint(err)
				} else {
					__antithesis_instrumentation__.Notify(496513)
				}
				__antithesis_instrumentation__.Notify(496476)
				val, err := seqval.AsInt64()
				if err != nil {
					__antithesis_instrumentation__.Notify(496514)
					return errors.Wrap(err, "unsupported setval arg")
				} else {
					__antithesis_instrumentation__.Notify(496515)
				}
				__antithesis_instrumentation__.Notify(496477)
				isCalled := false
				if len(fn.Exprs) == 3 {
					__antithesis_instrumentation__.Notify(496516)
					called, ok := fn.Exprs[2].(*tree.DBool)
					if !ok {
						__antithesis_instrumentation__.Notify(496518)
						err := errors.Errorf("unsupported setval %T arg: %v", fn.Exprs[2], fn.Exprs[2])
						if m.unsupportedStmtLogger.ignoreUnsupported {
							__antithesis_instrumentation__.Notify(496520)
							err := m.unsupportedStmtLogger.log(err.Error(), false)
							if err != nil {
								__antithesis_instrumentation__.Notify(496522)
								return err
							} else {
								__antithesis_instrumentation__.Notify(496523)
							}
							__antithesis_instrumentation__.Notify(496521)
							continue
						} else {
							__antithesis_instrumentation__.Notify(496524)
						}
						__antithesis_instrumentation__.Notify(496519)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496525)
					}
					__antithesis_instrumentation__.Notify(496517)
					isCalled = bool(*called)
				} else {
					__antithesis_instrumentation__.Notify(496526)
				}
				__antithesis_instrumentation__.Notify(496478)
				name, err := parser.ParseTableName(seqname.RawString())
				if err != nil {
					__antithesis_instrumentation__.Notify(496527)
					break
				} else {
					__antithesis_instrumentation__.Notify(496528)
				}
				__antithesis_instrumentation__.Notify(496479)

				seqName := name.Parts[0]
				if name.Schema() != "" {
					__antithesis_instrumentation__.Notify(496529)
					seqName = fmt.Sprintf("%s.%s", name.Schema(), name.Object())
				} else {
					__antithesis_instrumentation__.Notify(496530)
				}
				__antithesis_instrumentation__.Notify(496480)
				seq := m.tableDescs[seqName]
				if seq == nil {
					__antithesis_instrumentation__.Notify(496531)
					break
				} else {
					__antithesis_instrumentation__.Notify(496532)
				}
				__antithesis_instrumentation__.Notify(496481)
				key, val, err := sql.MakeSequenceKeyVal(m.evalCtx.Codec, seq, val, isCalled)
				if err != nil {
					__antithesis_instrumentation__.Notify(496533)
					return wrapRowErr(err, count, pgcode.Uncategorized, "")
				} else {
					__antithesis_instrumentation__.Notify(496534)
				}
				__antithesis_instrumentation__.Notify(496482)
				kv := roachpb.KeyValue{Key: key}
				kv.Value.SetInt(val)
				m.kvCh <- row.KVBatch{
					Source: inputIdx, KVs: []roachpb.KeyValue{kv}, Progress: input.ReadFraction(),
				}
			case "addgeometrycolumn":
				__antithesis_instrumentation__.Notify(496483)

			default:
				__antithesis_instrumentation__.Notify(496484)
				err := errors.Errorf("unsupported function %s in stmt %s", funcName, i.Select.String())
				if m.unsupportedStmtLogger.ignoreUnsupported {
					__antithesis_instrumentation__.Notify(496535)
					err := m.unsupportedStmtLogger.log(err.Error(), false)
					if err != nil {
						__antithesis_instrumentation__.Notify(496537)
						return err
					} else {
						__antithesis_instrumentation__.Notify(496538)
					}
					__antithesis_instrumentation__.Notify(496536)
					continue
				} else {
					__antithesis_instrumentation__.Notify(496539)
				}
				__antithesis_instrumentation__.Notify(496485)
				return wrapErrorWithUnsupportedHint(err)
			}
		case *tree.CreateExtension, *tree.CommentOnDatabase, *tree.CommentOnTable,
			*tree.CommentOnIndex, *tree.CommentOnConstraint, *tree.CommentOnColumn, *tree.AlterSequence,
			*tree.CommentOnSchema:
			__antithesis_instrumentation__.Notify(496339)

		case *tree.SetVar, *tree.BeginTransaction, *tree.CommitTransaction, *tree.Analyze:
			__antithesis_instrumentation__.Notify(496340)

		case *tree.CreateTable, *tree.CreateSchema, *tree.AlterTable, *tree.AlterTableOwner,
			*tree.CreateIndex, *tree.CreateSequence, *tree.DropTable:
			__antithesis_instrumentation__.Notify(496341)

		default:
			__antithesis_instrumentation__.Notify(496342)
			err := errors.Errorf("unsupported %T statement: %v", i, i)
			if m.unsupportedStmtLogger.ignoreUnsupported {
				__antithesis_instrumentation__.Notify(496540)
				err := m.unsupportedStmtLogger.log(err.Error(), false)
				if err != nil {
					__antithesis_instrumentation__.Notify(496542)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496543)
				}
				__antithesis_instrumentation__.Notify(496541)
				continue
			} else {
				__antithesis_instrumentation__.Notify(496544)
			}
			__antithesis_instrumentation__.Notify(496343)
			return wrapErrorWithUnsupportedHint(err)
		}
	}
	__antithesis_instrumentation__.Notify(496311)
	for _, conv := range m.tables {
		__antithesis_instrumentation__.Notify(496545)
		if err := conv.SendBatch(ctx); err != nil {
			__antithesis_instrumentation__.Notify(496546)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496547)
		}
	}
	__antithesis_instrumentation__.Notify(496312)
	return nil
}

func wrapWithLineTooLongHint(err error) error {
	__antithesis_instrumentation__.Notify(496548)
	return errors.WithHintf(
		err,
		"use `max_row_size` to increase the maximum line limit (default: %s).",
		humanizeutil.IBytes(defaultScanBuffer),
	)
}
