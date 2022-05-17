package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	mysqltypes "vitess.io/vitess/go/sqltypes"
	mysql "vitess.io/vitess/go/vt/sqlparser"
)

type mysqldumpReader struct {
	evalCtx  *tree.EvalContext
	tables   map[string]*row.DatumRowConverter
	kvCh     chan row.KVBatch
	debugRow func(tree.Datums)
	walltime int64
	opts     roachpb.MysqldumpOptions
}

var _ inputConverter = &mysqldumpReader{}

func newMysqldumpReader(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	kvCh chan row.KVBatch,
	walltime int64,
	tables map[string]*execinfrapb.ReadImportDataSpec_ImportTable,
	evalCtx *tree.EvalContext,
	opts roachpb.MysqldumpOptions,
) (*mysqldumpReader, error) {
	__antithesis_instrumentation__.Notify(495319)
	res := &mysqldumpReader{evalCtx: evalCtx, kvCh: kvCh, walltime: walltime, opts: opts}

	converters := make(map[string]*row.DatumRowConverter, len(tables))
	for name, table := range tables {
		__antithesis_instrumentation__.Notify(495321)
		if table.Desc == nil {
			__antithesis_instrumentation__.Notify(495324)
			converters[name] = nil
			continue
		} else {
			__antithesis_instrumentation__.Notify(495325)
		}
		__antithesis_instrumentation__.Notify(495322)
		conv, err := row.NewDatumRowConverter(ctx, semaCtx, tabledesc.NewBuilder(table.Desc).
			BuildImmutableTable(), nil, evalCtx, kvCh,
			nil, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(495326)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(495327)
		}
		__antithesis_instrumentation__.Notify(495323)
		converters[name] = conv
	}
	__antithesis_instrumentation__.Notify(495320)
	res.tables = converters
	return res, nil
}

func (m *mysqldumpReader) start(ctx ctxgroup.Group) {
	__antithesis_instrumentation__.Notify(495328)
}

func (m *mysqldumpReader) readFiles(
	ctx context.Context,
	dataFiles map[int32]string,
	resumePos map[int32]int64,
	format roachpb.IOFileFormat,
	makeExternalStorage cloud.ExternalStorageFactory,
	user security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(495329)
	return readInputFiles(ctx, dataFiles, resumePos, format, m.readFile, makeExternalStorage, user)
}

func (m *mysqldumpReader) readFile(
	ctx context.Context, input *fileReader, inputIdx int32, resumePos int64, rejected chan string,
) error {
	__antithesis_instrumentation__.Notify(495330)
	var inserts, count int64
	r := bufio.NewReaderSize(input, 1024*64)
	tableNameToRowsProcessed := make(map[string]int64)
	rowLimit := m.opts.RowLimit
	tokens := mysql.NewTokenizer(r)
	tokens.SkipSpecialComments = true

	for _, conv := range m.tables {
		__antithesis_instrumentation__.Notify(495334)
		conv.KvBatch.Source = inputIdx
		conv.FractionFn = input.ReadFraction
		conv.CompletedRowFn = func() int64 {
			__antithesis_instrumentation__.Notify(495335)
			return count
		}
	}
	__antithesis_instrumentation__.Notify(495331)

	for {
		__antithesis_instrumentation__.Notify(495336)
		stmt, err := mysql.ParseNextStrictDDL(tokens)
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(495340)
			break
		} else {
			__antithesis_instrumentation__.Notify(495341)
		}
		__antithesis_instrumentation__.Notify(495337)
		if errors.Is(err, mysql.ErrEmpty) {
			__antithesis_instrumentation__.Notify(495342)
			continue
		} else {
			__antithesis_instrumentation__.Notify(495343)
		}
		__antithesis_instrumentation__.Notify(495338)
		if err != nil {
			__antithesis_instrumentation__.Notify(495344)
			return errors.Wrap(err, "mysql parse error")
		} else {
			__antithesis_instrumentation__.Notify(495345)
		}
		__antithesis_instrumentation__.Notify(495339)
		switch i := stmt.(type) {
		case *mysql.Insert:
			__antithesis_instrumentation__.Notify(495346)
			name := safeString(i.Table.Name)
			conv, ok := m.tables[lexbase.NormalizeName(name)]
			if !ok {
				__antithesis_instrumentation__.Notify(495352)

				continue
			} else {
				__antithesis_instrumentation__.Notify(495353)
			}
			__antithesis_instrumentation__.Notify(495347)
			if conv == nil {
				__antithesis_instrumentation__.Notify(495354)
				return errors.Errorf("missing schema info for requested table %q", name)
			} else {
				__antithesis_instrumentation__.Notify(495355)
			}
			__antithesis_instrumentation__.Notify(495348)
			inserts++
			timestamp := timestampAfterEpoch(m.walltime)
			rows, ok := i.Rows.(mysql.Values)
			if !ok {
				__antithesis_instrumentation__.Notify(495356)
				return errors.Errorf(
					"insert statement %d: unexpected insert row type %T: %v", inserts, rows, i.Rows,
				)
			} else {
				__antithesis_instrumentation__.Notify(495357)
			}
			__antithesis_instrumentation__.Notify(495349)
			startingCount := count
			for _, inputRow := range rows {
				__antithesis_instrumentation__.Notify(495358)
				count++
				tableNameToRowsProcessed[name]++

				if count <= resumePos {
					__antithesis_instrumentation__.Notify(495364)
					continue
				} else {
					__antithesis_instrumentation__.Notify(495365)
				}
				__antithesis_instrumentation__.Notify(495359)
				if rowLimit != 0 && func() bool {
					__antithesis_instrumentation__.Notify(495366)
					return tableNameToRowsProcessed[name] > rowLimit == true
				}() == true {
					__antithesis_instrumentation__.Notify(495367)
					break
				} else {
					__antithesis_instrumentation__.Notify(495368)
				}
				__antithesis_instrumentation__.Notify(495360)
				if expected, got := len(conv.VisibleCols), len(inputRow); expected != got {
					__antithesis_instrumentation__.Notify(495369)
					return errors.Errorf("expected %d values, got %d: %v", expected, got, inputRow)
				} else {
					__antithesis_instrumentation__.Notify(495370)
				}
				__antithesis_instrumentation__.Notify(495361)
				for i, raw := range inputRow {
					__antithesis_instrumentation__.Notify(495371)
					converted, err := mysqlValueToDatum(raw, conv.VisibleColTypes[i], conv.EvalCtx)
					if err != nil {
						__antithesis_instrumentation__.Notify(495373)
						return errors.Wrapf(err, "reading row %d (%d in insert statement %d)",
							count, count-startingCount, inserts)
					} else {
						__antithesis_instrumentation__.Notify(495374)
					}
					__antithesis_instrumentation__.Notify(495372)
					conv.Datums[i] = converted
				}
				__antithesis_instrumentation__.Notify(495362)
				if err := conv.Row(ctx, inputIdx, count+int64(timestamp)); err != nil {
					__antithesis_instrumentation__.Notify(495375)
					return err
				} else {
					__antithesis_instrumentation__.Notify(495376)
				}
				__antithesis_instrumentation__.Notify(495363)
				if m.debugRow != nil {
					__antithesis_instrumentation__.Notify(495377)
					m.debugRow(conv.Datums)
				} else {
					__antithesis_instrumentation__.Notify(495378)
				}
			}
		default:
			__antithesis_instrumentation__.Notify(495350)
			if log.V(3) {
				__antithesis_instrumentation__.Notify(495379)
				log.Infof(ctx, "ignoring %T stmt: %v", i, i)
			} else {
				__antithesis_instrumentation__.Notify(495380)
			}
			__antithesis_instrumentation__.Notify(495351)
			continue
		}
	}
	__antithesis_instrumentation__.Notify(495332)
	for _, conv := range m.tables {
		__antithesis_instrumentation__.Notify(495381)
		if err := conv.SendBatch(ctx); err != nil {
			__antithesis_instrumentation__.Notify(495382)
			return err
		} else {
			__antithesis_instrumentation__.Notify(495383)
		}
	}
	__antithesis_instrumentation__.Notify(495333)
	return nil
}

const (
	zeroDate = "0000-00-00"
	zeroYear = "0000"
	zeroTime = "0000-00-00 00:00:00"
)

func mysqlValueToDatum(
	raw mysql.Expr, desired *types.T, evalContext *tree.EvalContext,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(495384)
	switch v := raw.(type) {
	case mysql.BoolVal:
		__antithesis_instrumentation__.Notify(495385)
		if v {
			__antithesis_instrumentation__.Notify(495391)
			return tree.DBoolTrue, nil
		} else {
			__antithesis_instrumentation__.Notify(495392)
		}
		__antithesis_instrumentation__.Notify(495386)
		return tree.DBoolFalse, nil
	case *mysql.Literal:
		__antithesis_instrumentation__.Notify(495387)
		switch v.Type {
		case mysql.StrVal:
			__antithesis_instrumentation__.Notify(495393)
			s := string(v.Val)

			if strings.HasPrefix(s, zeroYear) {
				__antithesis_instrumentation__.Notify(495399)
				switch desired.Family() {
				case types.TimestampTZFamily, types.TimestampFamily:
					__antithesis_instrumentation__.Notify(495400)
					if s == zeroTime {
						__antithesis_instrumentation__.Notify(495403)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(495404)
					}
				case types.DateFamily:
					__antithesis_instrumentation__.Notify(495401)
					if s == zeroDate {
						__antithesis_instrumentation__.Notify(495405)
						return tree.DNull, nil
					} else {
						__antithesis_instrumentation__.Notify(495406)
					}
				default:
					__antithesis_instrumentation__.Notify(495402)
				}
			} else {
				__antithesis_instrumentation__.Notify(495407)
			}
			__antithesis_instrumentation__.Notify(495394)

			return rowenc.ParseDatumStringAsWithRawBytes(desired, s, evalContext)
		case mysql.IntVal:
			__antithesis_instrumentation__.Notify(495395)
			return rowenc.ParseDatumStringAs(desired, string(v.Val), evalContext)
		case mysql.FloatVal:
			__antithesis_instrumentation__.Notify(495396)
			return rowenc.ParseDatumStringAs(desired, string(v.Val), evalContext)
		case mysql.HexVal:
			__antithesis_instrumentation__.Notify(495397)
			v, err := v.HexDecode()
			return tree.NewDBytes(tree.DBytes(v)), err

		default:
			__antithesis_instrumentation__.Notify(495398)
			return nil, fmt.Errorf("unsupported value type %c: %v", v.Type, v)
		}

	case *mysql.UnaryExpr:
		__antithesis_instrumentation__.Notify(495388)
		switch v.Operator {
		case mysql.UMinusOp:
			__antithesis_instrumentation__.Notify(495408)
			parsed, err := mysqlValueToDatum(v.Expr, desired, evalContext)
			if err != nil {
				__antithesis_instrumentation__.Notify(495412)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(495413)
			}
			__antithesis_instrumentation__.Notify(495409)
			switch i := parsed.(type) {
			case *tree.DInt:
				__antithesis_instrumentation__.Notify(495414)
				return tree.NewDInt(-*i), nil
			case *tree.DFloat:
				__antithesis_instrumentation__.Notify(495415)
				return tree.NewDFloat(-*i), nil
			case *tree.DDecimal:
				__antithesis_instrumentation__.Notify(495416)
				dec := &i.Decimal
				dd := &tree.DDecimal{}
				dd.Decimal.Neg(dec)
				return dd, nil
			default:
				__antithesis_instrumentation__.Notify(495417)
				return nil, errors.Errorf("unsupported negation of %T", i)
			}
		case mysql.UBinaryOp:
			__antithesis_instrumentation__.Notify(495410)

			return mysqlValueToDatum(v.Expr, desired, evalContext)
		default:
			__antithesis_instrumentation__.Notify(495411)
			return nil, errors.Errorf("unexpected operator: %q", v.Operator)
		}

	case *mysql.NullVal:
		__antithesis_instrumentation__.Notify(495389)
		return tree.DNull, nil

	default:
		__antithesis_instrumentation__.Notify(495390)
		return nil, errors.Errorf("unexpected value type %T: %v", v, v)
	}
}

func readMysqlCreateTable(
	ctx context.Context,
	input io.Reader,
	evalCtx *tree.EvalContext,
	p sql.JobExecContext,
	startingID descpb.ID,
	parentDB catalog.DatabaseDescriptor,
	match string,
	fks fkHandler,
	seqVals map[descpb.ID]int64,
	owner security.SQLUsername,
	walltime int64,
) ([]*tabledesc.Mutable, error) {
	__antithesis_instrumentation__.Notify(495418)
	match = lexbase.NormalizeName(match)
	r := bufio.NewReaderSize(input, 1024*64)
	tokens := mysql.NewTokenizer(r)
	tokens.SkipSpecialComments = true

	var ret []*tabledesc.Mutable
	var fkDefs []delayedFK
	var found bool
	var names []string
	for {
		__antithesis_instrumentation__.Notify(495423)
		stmt, err := mysql.ParseNextStrictDDL(tokens)
		if err == nil {
			__antithesis_instrumentation__.Notify(495428)
			err = tokens.LastError
		} else {
			__antithesis_instrumentation__.Notify(495429)
		}
		__antithesis_instrumentation__.Notify(495424)
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(495430)
			break
		} else {
			__antithesis_instrumentation__.Notify(495431)
		}
		__antithesis_instrumentation__.Notify(495425)
		if errors.Is(err, mysql.ErrEmpty) {
			__antithesis_instrumentation__.Notify(495432)
			continue
		} else {
			__antithesis_instrumentation__.Notify(495433)
		}
		__antithesis_instrumentation__.Notify(495426)
		if err != nil {
			__antithesis_instrumentation__.Notify(495434)
			return nil, errors.Wrap(err, "mysql parse error")
		} else {
			__antithesis_instrumentation__.Notify(495435)
		}
		__antithesis_instrumentation__.Notify(495427)
		if i, ok := stmt.(*mysql.DDL); ok && func() bool {
			__antithesis_instrumentation__.Notify(495436)
			return i.Action == mysql.CreateDDLAction == true
		}() == true {
			__antithesis_instrumentation__.Notify(495437)
			name := safeString(i.Table.Name)
			if match != "" && func() bool {
				__antithesis_instrumentation__.Notify(495440)
				return match != name == true
			}() == true {
				__antithesis_instrumentation__.Notify(495441)
				names = append(names, name)
				continue
			} else {
				__antithesis_instrumentation__.Notify(495442)
			}
			__antithesis_instrumentation__.Notify(495438)
			id := descpb.ID(int(startingID) + len(ret))
			tbl, moreFKs, err := mysqlTableToCockroach(ctx, evalCtx, p, parentDB, id, name, i.TableSpec, fks, seqVals, owner, walltime)
			if err != nil {
				__antithesis_instrumentation__.Notify(495443)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(495444)
			}
			__antithesis_instrumentation__.Notify(495439)
			fkDefs = append(fkDefs, moreFKs...)
			ret = append(ret, tbl...)
			if match == name {
				__antithesis_instrumentation__.Notify(495445)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(495446)
			}
		} else {
			__antithesis_instrumentation__.Notify(495447)
		}
	}
	__antithesis_instrumentation__.Notify(495419)
	if ret == nil {
		__antithesis_instrumentation__.Notify(495448)
		return nil, errors.Errorf("no table definitions found")
	} else {
		__antithesis_instrumentation__.Notify(495449)
	}
	__antithesis_instrumentation__.Notify(495420)
	if match != "" && func() bool {
		__antithesis_instrumentation__.Notify(495450)
		return !found == true
	}() == true {
		__antithesis_instrumentation__.Notify(495451)
		return nil, errors.Errorf("table %q not found in file (found tables: %s)", match, strings.Join(names, ", "))
	} else {
		__antithesis_instrumentation__.Notify(495452)
	}
	__antithesis_instrumentation__.Notify(495421)
	if err := addDelayedFKs(ctx, fkDefs, fks.resolver, evalCtx); err != nil {
		__antithesis_instrumentation__.Notify(495453)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(495454)
	}
	__antithesis_instrumentation__.Notify(495422)
	return ret, nil
}

type mysqlIdent interface{ CompliantName() string }

func safeString(in mysqlIdent) string {
	__antithesis_instrumentation__.Notify(495455)
	return lexbase.NormalizeName(in.CompliantName())
}

func safeName(in mysqlIdent) tree.Name {
	__antithesis_instrumentation__.Notify(495456)
	return tree.Name(safeString(in))
}

func mysqlTableToCockroach(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	p sql.JobExecContext,
	parentDB catalog.DatabaseDescriptor,
	id descpb.ID,
	name string,
	in *mysql.TableSpec,
	fks fkHandler,
	seqVals map[descpb.ID]int64,
	owner security.SQLUsername,
	walltime int64,
) ([]*tabledesc.Mutable, []delayedFK, error) {
	__antithesis_instrumentation__.Notify(495457)
	if in == nil {
		__antithesis_instrumentation__.Notify(495470)
		return nil, nil, errors.Errorf("could not read definition for table %q (possible unsupported type?)", name)
	} else {
		__antithesis_instrumentation__.Notify(495471)
	}
	__antithesis_instrumentation__.Notify(495458)

	time := hlc.Timestamp{WallTime: walltime}

	const seqOpt = "auto_increment="
	var seqName string
	var startingValue int64
	for _, opt := range strings.Fields(strings.ToLower(in.Options)) {
		__antithesis_instrumentation__.Notify(495472)
		if strings.HasPrefix(opt, seqOpt) {
			__antithesis_instrumentation__.Notify(495473)
			seqName = name + "_auto_inc"
			i, err := strconv.Atoi(strings.TrimPrefix(opt, seqOpt))
			if err != nil {
				__antithesis_instrumentation__.Notify(495475)
				return nil, nil, errors.Wrapf(err, "parsing AUTO_INCREMENT value")
			} else {
				__antithesis_instrumentation__.Notify(495476)
			}
			__antithesis_instrumentation__.Notify(495474)
			startingValue = int64(i)
			break
		} else {
			__antithesis_instrumentation__.Notify(495477)
		}
	}
	__antithesis_instrumentation__.Notify(495459)

	if seqName == "" {
		__antithesis_instrumentation__.Notify(495478)
		for _, raw := range in.Columns {
			__antithesis_instrumentation__.Notify(495479)
			if raw.Type.Autoincrement {
				__antithesis_instrumentation__.Notify(495480)
				seqName = name + "_auto_inc"
				break
			} else {
				__antithesis_instrumentation__.Notify(495481)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(495482)
	}
	__antithesis_instrumentation__.Notify(495460)
	publicSchema, err := getPublicSchemaDescForDatabase(ctx, p.ExecCfg(), parentDB)
	if err != nil {
		__antithesis_instrumentation__.Notify(495483)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(495484)
	}
	__antithesis_instrumentation__.Notify(495461)
	var seqDesc *tabledesc.Mutable

	if seqName != "" {
		__antithesis_instrumentation__.Notify(495485)
		var opts tree.SequenceOptions
		if startingValue != 0 {
			__antithesis_instrumentation__.Notify(495488)
			opts = tree.SequenceOptions{{Name: tree.SeqOptStart, IntVal: &startingValue}}
			seqVals[id] = startingValue
		} else {
			__antithesis_instrumentation__.Notify(495489)
		}
		__antithesis_instrumentation__.Notify(495486)
		var err error
		privilegeDesc := catpb.NewBasePrivilegeDescriptor(owner)
		seqDesc, err = sql.NewSequenceTableDesc(
			ctx,
			nil,
			evalCtx.Settings,
			seqName,
			opts,
			parentDB.GetID(),
			publicSchema.GetID(),
			id,
			time,
			privilegeDesc,
			tree.PersistencePermanent,

			false,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(495490)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(495491)
		}
		__antithesis_instrumentation__.Notify(495487)
		fks.resolver.tableNameToDesc[seqName] = seqDesc
		id++
	} else {
		__antithesis_instrumentation__.Notify(495492)
	}
	__antithesis_instrumentation__.Notify(495462)

	stmt := &tree.CreateTable{Table: tree.MakeUnqualifiedTableName(tree.Name(name))}

	checks := make(map[string]*tree.CheckConstraintTableDef)

	for _, raw := range in.Columns {
		__antithesis_instrumentation__.Notify(495493)
		def, err := mysqlColToCockroach(safeString(raw.Name), raw.Type, checks)
		if err != nil {
			__antithesis_instrumentation__.Notify(495497)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(495498)
		}
		__antithesis_instrumentation__.Notify(495494)
		if raw.Type.Autoincrement {
			__antithesis_instrumentation__.Notify(495499)

			expr, err := parser.ParseExpr(fmt.Sprintf("nextval('%s':::STRING)", seqName))
			if err != nil {
				__antithesis_instrumentation__.Notify(495501)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(495502)
			}
			__antithesis_instrumentation__.Notify(495500)
			def.DefaultExpr.Expr = expr
		} else {
			__antithesis_instrumentation__.Notify(495503)
		}
		__antithesis_instrumentation__.Notify(495495)

		if dType, ok := def.Type.(*types.T); ok {
			__antithesis_instrumentation__.Notify(495504)
			if dType.Equivalent(types.Int) && func() bool {
				__antithesis_instrumentation__.Notify(495505)
				return p != nil == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(495506)
				return p.SessionData() != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(495507)
				def.Type = parser.NakedIntTypeFromDefaultIntSize(p.SessionData().DefaultIntSize)
			} else {
				__antithesis_instrumentation__.Notify(495508)
			}
		} else {
			__antithesis_instrumentation__.Notify(495509)
		}
		__antithesis_instrumentation__.Notify(495496)
		stmt.Defs = append(stmt.Defs, def)
	}
	__antithesis_instrumentation__.Notify(495463)

	for _, raw := range in.Indexes {
		__antithesis_instrumentation__.Notify(495510)
		var elems tree.IndexElemList
		for _, col := range raw.Columns {
			__antithesis_instrumentation__.Notify(495513)
			elems = append(elems, tree.IndexElem{Column: safeName(col.Column)})
		}
		__antithesis_instrumentation__.Notify(495511)

		idxName := safeName(raw.Info.Name)

		if raw.Info.Primary {
			__antithesis_instrumentation__.Notify(495514)
			idxName = tree.Name(tabledesc.PrimaryKeyIndexName(name))
		} else {
			__antithesis_instrumentation__.Notify(495515)
		}
		__antithesis_instrumentation__.Notify(495512)
		idx := tree.IndexTableDef{Name: idxName, Columns: elems}
		if raw.Info.Primary || func() bool {
			__antithesis_instrumentation__.Notify(495516)
			return raw.Info.Unique == true
		}() == true {
			__antithesis_instrumentation__.Notify(495517)
			stmt.Defs = append(stmt.Defs, &tree.UniqueConstraintTableDef{IndexTableDef: idx, PrimaryKey: raw.Info.Primary})
		} else {
			__antithesis_instrumentation__.Notify(495518)
			stmt.Defs = append(stmt.Defs, &idx)
		}
	}
	__antithesis_instrumentation__.Notify(495464)

	for _, c := range checks {
		__antithesis_instrumentation__.Notify(495519)
		stmt.Defs = append(stmt.Defs, c)
	}
	__antithesis_instrumentation__.Notify(495465)

	semaCtx := tree.MakeSemaContext()
	semaCtxPtr := &semaCtx

	if p != nil && func() bool {
		__antithesis_instrumentation__.Notify(495520)
		return p.SemaCtx() != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(495521)
		semaCtxPtr = p.SemaCtx()
	} else {
		__antithesis_instrumentation__.Notify(495522)
	}
	__antithesis_instrumentation__.Notify(495466)

	semaCtxPtr = makeSemaCtxWithoutTypeResolver(semaCtxPtr)
	desc, err := MakeSimpleTableDescriptor(
		ctx, semaCtxPtr, evalCtx.Settings, stmt, parentDB,
		publicSchema, id, fks, time.WallTime,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(495523)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(495524)
	}
	__antithesis_instrumentation__.Notify(495467)

	var fkDefs []delayedFK
	for _, raw := range in.Constraints {
		__antithesis_instrumentation__.Notify(495525)
		switch i := raw.Details.(type) {
		case *mysql.ForeignKeyDefinition:
			__antithesis_instrumentation__.Notify(495526)
			if !fks.allowed {
				__antithesis_instrumentation__.Notify(495531)
				return nil, nil, errors.Errorf("foreign keys not supported: %s", mysql.String(raw))
			} else {
				__antithesis_instrumentation__.Notify(495532)
			}
			__antithesis_instrumentation__.Notify(495527)
			if fks.skip {
				__antithesis_instrumentation__.Notify(495533)
				continue
			} else {
				__antithesis_instrumentation__.Notify(495534)
			}
			__antithesis_instrumentation__.Notify(495528)
			fromCols := i.Source
			toTable := tree.MakeTableNameWithSchema(
				safeName(i.ReferencedTable.Qualifier),
				tree.PublicSchemaName,
				safeName(i.ReferencedTable.Name),
			)
			toCols := i.ReferencedColumns
			d := &tree.ForeignKeyConstraintTableDef{
				Name:     tree.Name(lexbase.NormalizeName(raw.Name)),
				FromCols: toNameList(fromCols),
				ToCols:   toNameList(toCols),
			}

			if i.OnDelete != mysql.NoAction {
				__antithesis_instrumentation__.Notify(495535)
				d.Actions.Delete = mysqlActionToCockroach(i.OnDelete)
			} else {
				__antithesis_instrumentation__.Notify(495536)
			}
			__antithesis_instrumentation__.Notify(495529)
			if i.OnUpdate != mysql.NoAction {
				__antithesis_instrumentation__.Notify(495537)
				d.Actions.Update = mysqlActionToCockroach(i.OnUpdate)
			} else {
				__antithesis_instrumentation__.Notify(495538)
			}
			__antithesis_instrumentation__.Notify(495530)

			d.Table = toTable
			fkDefs = append(fkDefs, delayedFK{
				db:  parentDB,
				sc:  schemadesc.GetPublicSchema(),
				tbl: desc,
				def: d,
			})
		}
	}
	__antithesis_instrumentation__.Notify(495468)
	fks.resolver.tableNameToDesc[desc.Name] = desc
	if seqDesc != nil {
		__antithesis_instrumentation__.Notify(495539)
		return []*tabledesc.Mutable{seqDesc, desc}, fkDefs, nil
	} else {
		__antithesis_instrumentation__.Notify(495540)
	}
	__antithesis_instrumentation__.Notify(495469)
	return []*tabledesc.Mutable{desc}, fkDefs, nil
}

func mysqlActionToCockroach(action mysql.ReferenceAction) tree.ReferenceAction {
	__antithesis_instrumentation__.Notify(495541)
	switch action {
	case mysql.Restrict:
		__antithesis_instrumentation__.Notify(495543)
		return tree.Restrict
	case mysql.Cascade:
		__antithesis_instrumentation__.Notify(495544)
		return tree.Cascade
	case mysql.SetNull:
		__antithesis_instrumentation__.Notify(495545)
		return tree.SetNull
	case mysql.SetDefault:
		__antithesis_instrumentation__.Notify(495546)
		return tree.SetDefault
	default:
		__antithesis_instrumentation__.Notify(495547)
	}
	__antithesis_instrumentation__.Notify(495542)
	return tree.NoAction
}

type delayedFK struct {
	db  catalog.DatabaseDescriptor
	sc  catalog.SchemaDescriptor
	tbl *tabledesc.Mutable
	def *tree.ForeignKeyConstraintTableDef
}

func addDelayedFKs(
	ctx context.Context, defs []delayedFK, resolver fkResolver, evalCtx *tree.EvalContext,
) error {
	__antithesis_instrumentation__.Notify(495548)
	for _, def := range defs {
		__antithesis_instrumentation__.Notify(495550)
		backrefs := map[descpb.ID]*tabledesc.Mutable{}
		if err := sql.ResolveFK(
			ctx, nil, &resolver, def.db, def.sc, def.tbl, def.def,
			backrefs, sql.NewTable,
			tree.ValidationDefault, evalCtx,
		); err != nil {
			__antithesis_instrumentation__.Notify(495553)
			return err
		} else {
			__antithesis_instrumentation__.Notify(495554)
		}
		__antithesis_instrumentation__.Notify(495551)
		if err := fixDescriptorFKState(def.tbl); err != nil {
			__antithesis_instrumentation__.Notify(495555)
			return err
		} else {
			__antithesis_instrumentation__.Notify(495556)
		}
		__antithesis_instrumentation__.Notify(495552)
		version := evalCtx.Settings.Version.ActiveVersion(ctx)
		if err := def.tbl.AllocateIDs(ctx, version); err != nil {
			__antithesis_instrumentation__.Notify(495557)
			return err
		} else {
			__antithesis_instrumentation__.Notify(495558)
		}
	}
	__antithesis_instrumentation__.Notify(495549)
	return nil
}

func toNameList(cols mysql.Columns) tree.NameList {
	__antithesis_instrumentation__.Notify(495559)
	res := make([]tree.Name, len(cols))
	for i := range cols {
		__antithesis_instrumentation__.Notify(495561)
		res[i] = safeName(cols[i])
	}
	__antithesis_instrumentation__.Notify(495560)
	return res
}

func mysqlColToCockroach(
	name string, col mysql.ColumnType, checks map[string]*tree.CheckConstraintTableDef,
) (*tree.ColumnTableDef, error) {
	__antithesis_instrumentation__.Notify(495562)
	def := &tree.ColumnTableDef{Name: tree.Name(name)}

	var length, scale int

	if col.Length != nil {
		__antithesis_instrumentation__.Notify(495568)
		num, err := strconv.Atoi(string(col.Length.Val))
		if err != nil {
			__antithesis_instrumentation__.Notify(495570)
			return nil, errors.Wrapf(err, "could not parse length for column %q", name)
		} else {
			__antithesis_instrumentation__.Notify(495571)
		}
		__antithesis_instrumentation__.Notify(495569)
		length = num
	} else {
		__antithesis_instrumentation__.Notify(495572)
	}
	__antithesis_instrumentation__.Notify(495563)

	if col.Scale != nil {
		__antithesis_instrumentation__.Notify(495573)
		num, err := strconv.Atoi(string(col.Scale.Val))
		if err != nil {
			__antithesis_instrumentation__.Notify(495575)
			return nil, errors.Wrapf(err, "could not parse scale for column %q", name)
		} else {
			__antithesis_instrumentation__.Notify(495576)
		}
		__antithesis_instrumentation__.Notify(495574)
		scale = num
	} else {
		__antithesis_instrumentation__.Notify(495577)
	}
	__antithesis_instrumentation__.Notify(495564)

	switch typ := col.SQLType(); typ {

	case mysqltypes.Char:
		__antithesis_instrumentation__.Notify(495578)
		def.Type = types.MakeChar(int32(length))
	case mysqltypes.VarChar:
		__antithesis_instrumentation__.Notify(495579)
		def.Type = types.MakeVarChar(int32(length))
	case mysqltypes.Text:
		__antithesis_instrumentation__.Notify(495580)
		def.Type = types.MakeString(int32(length))

	case mysqltypes.Blob:
		__antithesis_instrumentation__.Notify(495581)
		def.Type = types.Bytes
	case mysqltypes.VarBinary:
		__antithesis_instrumentation__.Notify(495582)
		def.Type = types.Bytes
	case mysqltypes.Binary:
		__antithesis_instrumentation__.Notify(495583)
		def.Type = types.Bytes

	case mysqltypes.Int8:
		__antithesis_instrumentation__.Notify(495584)
		def.Type = types.Int2
	case mysqltypes.Uint8:
		__antithesis_instrumentation__.Notify(495585)
		def.Type = types.Int2
	case mysqltypes.Int16:
		__antithesis_instrumentation__.Notify(495586)
		def.Type = types.Int2
	case mysqltypes.Uint16:
		__antithesis_instrumentation__.Notify(495587)
		def.Type = types.Int4
	case mysqltypes.Int24:
		__antithesis_instrumentation__.Notify(495588)
		def.Type = types.Int4
	case mysqltypes.Uint24:
		__antithesis_instrumentation__.Notify(495589)
		def.Type = types.Int4
	case mysqltypes.Int32:
		__antithesis_instrumentation__.Notify(495590)
		def.Type = types.Int4
	case mysqltypes.Uint32:
		__antithesis_instrumentation__.Notify(495591)
		def.Type = types.Int
	case mysqltypes.Int64:
		__antithesis_instrumentation__.Notify(495592)
		def.Type = types.Int
	case mysqltypes.Uint64:
		__antithesis_instrumentation__.Notify(495593)
		def.Type = types.Int

	case mysqltypes.Float32:
		__antithesis_instrumentation__.Notify(495594)
		def.Type = types.Float4
	case mysqltypes.Float64:
		__antithesis_instrumentation__.Notify(495595)
		def.Type = types.Float

	case mysqltypes.Decimal:
		__antithesis_instrumentation__.Notify(495596)
		def.Type = types.MakeDecimal(int32(length), int32(scale))

	case mysqltypes.Date:
		__antithesis_instrumentation__.Notify(495597)
		def.Type = types.Date
		if col.Default != nil {
			__antithesis_instrumentation__.Notify(495609)
			if lit, ok := col.Default.(*mysql.Literal); ok && func() bool {
				__antithesis_instrumentation__.Notify(495610)
				return bytes.Equal(lit.Val, []byte(zeroDate)) == true
			}() == true {
				__antithesis_instrumentation__.Notify(495611)
				col.Default = nil
			} else {
				__antithesis_instrumentation__.Notify(495612)
			}
		} else {
			__antithesis_instrumentation__.Notify(495613)
		}
	case mysqltypes.Time:
		__antithesis_instrumentation__.Notify(495598)
		def.Type = types.Time
	case mysqltypes.Timestamp:
		__antithesis_instrumentation__.Notify(495599)
		def.Type = types.TimestampTZ
		if col.Default != nil {
			__antithesis_instrumentation__.Notify(495614)
			if lit, ok := col.Default.(*mysql.Literal); ok && func() bool {
				__antithesis_instrumentation__.Notify(495615)
				return bytes.Equal(lit.Val, []byte(zeroTime)) == true
			}() == true {
				__antithesis_instrumentation__.Notify(495616)
				col.Default = nil
			} else {
				__antithesis_instrumentation__.Notify(495617)
			}
		} else {
			__antithesis_instrumentation__.Notify(495618)
		}
	case mysqltypes.Datetime:
		__antithesis_instrumentation__.Notify(495600)
		def.Type = types.TimestampTZ
		if col.Default != nil {
			__antithesis_instrumentation__.Notify(495619)
			if lit, ok := col.Default.(*mysql.Literal); ok && func() bool {
				__antithesis_instrumentation__.Notify(495620)
				return bytes.Equal(lit.Val, []byte(zeroTime)) == true
			}() == true {
				__antithesis_instrumentation__.Notify(495621)
				col.Default = nil
			} else {
				__antithesis_instrumentation__.Notify(495622)
			}
		} else {
			__antithesis_instrumentation__.Notify(495623)
		}
	case mysqltypes.Year:
		__antithesis_instrumentation__.Notify(495601)
		def.Type = types.Int2

	case mysqltypes.Enum:
		__antithesis_instrumentation__.Notify(495602)
		def.Type = types.String

		expr, err := parser.ParseExpr(fmt.Sprintf("%s in (%s)", name, strings.Join(col.EnumValues, ",")))
		if err != nil {
			__antithesis_instrumentation__.Notify(495624)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(495625)
		}
		__antithesis_instrumentation__.Notify(495603)
		checks[name] = &tree.CheckConstraintTableDef{
			Name: tree.Name(fmt.Sprintf("imported_from_enum_%s", name)),
			Expr: expr,
		}

	case mysqltypes.TypeJSON:
		__antithesis_instrumentation__.Notify(495604)
		def.Type = types.Jsonb

	case mysqltypes.Set:
		__antithesis_instrumentation__.Notify(495605)
		return nil, errors.WithHint(
			unimplemented.NewWithIssue(32560, "cannot import SET columns at this time"),
			"try converting the column to a 64-bit integer before import")
	case mysqltypes.Geometry:
		__antithesis_instrumentation__.Notify(495606)
		return nil, unimplemented.NewWithIssue(32559, "cannot import GEOMETRY columns at this time")
	case mysqltypes.Bit:
		__antithesis_instrumentation__.Notify(495607)
		return nil, errors.WithHint(
			unimplemented.NewWithIssue(32561, "cannot import BIT columns at this time"),
			"try converting the column to a 64-bit integer before import")
	default:
		__antithesis_instrumentation__.Notify(495608)
		return nil, unimplemented.Newf(fmt.Sprintf("import.mysqlcoltype.%s", typ),
			"unsupported mysql type %q", col.Type)
	}
	__antithesis_instrumentation__.Notify(495565)

	if col.NotNull {
		__antithesis_instrumentation__.Notify(495626)
		def.Nullable.Nullability = tree.NotNull
	} else {
		__antithesis_instrumentation__.Notify(495627)
		def.Nullable.Nullability = tree.Null
	}
	__antithesis_instrumentation__.Notify(495566)

	if col.Default != nil {
		__antithesis_instrumentation__.Notify(495628)
		if _, ok := col.Default.(*mysql.NullVal); !ok {
			__antithesis_instrumentation__.Notify(495629)
			if literal, ok := col.Default.(*mysql.Literal); ok && func() bool {
				__antithesis_instrumentation__.Notify(495630)
				return literal.Type == mysql.StrVal == true
			}() == true {
				__antithesis_instrumentation__.Notify(495631)

				def.DefaultExpr.Expr = tree.NewStrVal(string(literal.Val))
			} else {
				__antithesis_instrumentation__.Notify(495632)
				exprString := mysql.String(col.Default)
				expr, err := parser.ParseExpr(exprString)
				if err != nil {
					__antithesis_instrumentation__.Notify(495634)

					return nil, errors.Wrapf(
						unimplemented.Newf("import.mysql.default", "unsupported default expression"),
						"error parsing %q for column %q: %v",
						exprString,
						name,
						err,
					)
				} else {
					__antithesis_instrumentation__.Notify(495635)
				}
				__antithesis_instrumentation__.Notify(495633)
				def.DefaultExpr.Expr = expr
			}
		} else {
			__antithesis_instrumentation__.Notify(495636)
		}
	} else {
		__antithesis_instrumentation__.Notify(495637)
	}
	__antithesis_instrumentation__.Notify(495567)
	return def, nil
}
