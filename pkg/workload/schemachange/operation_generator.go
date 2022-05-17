package schemachange

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachange"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"
)

type operationGeneratorParams struct {
	seqNum             *int64
	errorRate          int
	enumPct            int
	rng                *rand.Rand
	ops                *deck
	maxSourceTables    int
	sequenceOwnedByPct int
	fkParentInvalidPct int
	fkChildInvalidPct  int
}

type operationGenerator struct {
	params               *operationGeneratorParams
	expectedExecErrors   errorCodeSet
	expectedCommitErrors errorCodeSet

	candidateExpectedCommitErrors errorCodeSet

	opsInTxn []opType

	stmtsInTxt []string

	opGenLog strings.Builder
}

func (og *operationGenerator) LogQueryResults(queryName string, result string) {
	__antithesis_instrumentation__.Notify(695899)
	og.opGenLog.WriteString(fmt.Sprintf("QUERY [%s] :", queryName))
	og.opGenLog.WriteString(result)
	og.opGenLog.WriteString("\n")
}

func (og *operationGenerator) LogQueryResultArray(queryName string, results []string) {
	__antithesis_instrumentation__.Notify(695900)
	og.opGenLog.WriteString(fmt.Sprintf("QUERY [%s] : ", queryName))
	for _, result := range results {
		__antithesis_instrumentation__.Notify(695902)
		og.opGenLog.WriteString(result)
		og.opGenLog.WriteString(",")
	}
	__antithesis_instrumentation__.Notify(695901)

	og.opGenLog.WriteString("\n")
}

func (og *operationGenerator) GetOpGenLog() string {
	__antithesis_instrumentation__.Notify(695903)
	return og.opGenLog.String()
}

func makeOperationGenerator(params *operationGeneratorParams) *operationGenerator {
	__antithesis_instrumentation__.Notify(695904)
	return &operationGenerator{
		params:                        params,
		expectedExecErrors:            makeExpectedErrorSet(),
		expectedCommitErrors:          makeExpectedErrorSet(),
		candidateExpectedCommitErrors: makeExpectedErrorSet(),
	}
}

func (og *operationGenerator) resetOpState() {
	__antithesis_instrumentation__.Notify(695905)
	og.expectedExecErrors.reset()
	og.candidateExpectedCommitErrors.reset()
	og.opGenLog = strings.Builder{}
}

func (og *operationGenerator) resetTxnState() {
	__antithesis_instrumentation__.Notify(695906)
	og.expectedCommitErrors.reset()
	og.opsInTxn = nil
	og.stmtsInTxt = nil
}

type opType int

func (ot opType) isDDL() bool {
	__antithesis_instrumentation__.Notify(695907)
	return ot != insertRow && func() bool {
		__antithesis_instrumentation__.Notify(695908)
		return ot != validate == true
	}() == true
}

const (
	addColumn opType = iota
	addConstraint
	addForeignKeyConstraint
	addRegion
	addUniqueConstraint

	alterTableLocality

	createIndex
	createSequence
	createTable
	createTableAs
	createView
	createEnum
	createSchema

	dropColumn
	dropColumnDefault
	dropColumnNotNull
	dropColumnStored
	dropConstraint
	dropIndex
	dropSequence
	dropTable
	dropView
	dropSchema

	primaryRegion

	renameColumn
	renameIndex
	renameSequence
	renameTable
	renameView

	setColumnDefault
	setColumnNotNull
	setColumnType

	survive

	insertRow

	validate

	numOpTypes int = iota
)

var opFuncs = map[opType]func(*operationGenerator, context.Context, pgx.Tx) (string, error){
	addColumn:               (*operationGenerator).addColumn,
	addConstraint:           (*operationGenerator).addConstraint,
	addForeignKeyConstraint: (*operationGenerator).addForeignKeyConstraint,
	addRegion:               (*operationGenerator).addRegion,
	addUniqueConstraint:     (*operationGenerator).addUniqueConstraint,
	alterTableLocality:      (*operationGenerator).alterTableLocality,
	createIndex:             (*operationGenerator).createIndex,
	createSequence:          (*operationGenerator).createSequence,
	createTable:             (*operationGenerator).createTable,
	createTableAs:           (*operationGenerator).createTableAs,
	createView:              (*operationGenerator).createView,
	createEnum:              (*operationGenerator).createEnum,
	createSchema:            (*operationGenerator).createSchema,
	dropColumn:              (*operationGenerator).dropColumn,
	dropColumnDefault:       (*operationGenerator).dropColumnDefault,
	dropColumnNotNull:       (*operationGenerator).dropColumnNotNull,
	dropColumnStored:        (*operationGenerator).dropColumnStored,
	dropConstraint:          (*operationGenerator).dropConstraint,
	dropIndex:               (*operationGenerator).dropIndex,
	dropSequence:            (*operationGenerator).dropSequence,
	dropTable:               (*operationGenerator).dropTable,
	dropView:                (*operationGenerator).dropView,
	dropSchema:              (*operationGenerator).dropSchema,
	primaryRegion:           (*operationGenerator).primaryRegion,
	renameColumn:            (*operationGenerator).renameColumn,
	renameIndex:             (*operationGenerator).renameIndex,
	renameSequence:          (*operationGenerator).renameSequence,
	renameTable:             (*operationGenerator).renameTable,
	renameView:              (*operationGenerator).renameView,
	setColumnDefault:        (*operationGenerator).setColumnDefault,
	setColumnNotNull:        (*operationGenerator).setColumnNotNull,
	setColumnType:           (*operationGenerator).setColumnType,
	survive:                 (*operationGenerator).survive,
	insertRow:               (*operationGenerator).insertRow,
	validate:                (*operationGenerator).validate,
}

func init() {

	if len(opFuncs) != numOpTypes {
		panic(errors.Errorf("expected %d opFuncs, got %d", numOpTypes, len(opFuncs)))
	}
}

var opWeights = []int{
	addColumn:               1,
	addConstraint:           0,
	addForeignKeyConstraint: 0,
	addRegion:               1,
	addUniqueConstraint:     0,
	alterTableLocality:      1,
	createIndex:             1,
	createSequence:          1,
	createTable:             1,
	createTableAs:           1,
	createView:              1,
	createEnum:              1,
	createSchema:            1,
	dropColumn:              0,
	dropColumnDefault:       1,
	dropColumnNotNull:       1,
	dropColumnStored:        1,
	dropConstraint:          1,
	dropIndex:               1,
	dropSequence:            1,
	dropTable:               1,
	dropView:                1,
	dropSchema:              1,
	primaryRegion:           1,
	renameColumn:            1,
	renameIndex:             1,
	renameSequence:          1,
	renameTable:             1,
	renameView:              1,
	setColumnDefault:        1,
	setColumnNotNull:        1,
	setColumnType:           0,
	survive:                 1,
	insertRow:               0,
	validate:                2,
}

func (og *operationGenerator) randOp(ctx context.Context, tx pgx.Tx) (stmt string, err error) {
	__antithesis_instrumentation__.Notify(695909)

	for {
		__antithesis_instrumentation__.Notify(695911)
		op := opType(og.params.ops.Int())
		og.resetOpState()
		stmt, err = opFuncs[op](og, ctx, tx)
		if err != nil {
			__antithesis_instrumentation__.Notify(695913)
			if errors.Is(err, pgx.ErrNoRows) {
				__antithesis_instrumentation__.Notify(695916)
				continue
			} else {
				__antithesis_instrumentation__.Notify(695917)
			}
			__antithesis_instrumentation__.Notify(695914)

			if errors.Is(err, ErrSchemaChangesDisallowedDueToPkSwap) {
				__antithesis_instrumentation__.Notify(695918)
				continue
			} else {
				__antithesis_instrumentation__.Notify(695919)
			}
			__antithesis_instrumentation__.Notify(695915)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(695920)
		}
		__antithesis_instrumentation__.Notify(695912)

		og.checkIfOpViolatesDDLAfterWrite(op)
		og.stmtsInTxt = append(og.stmtsInTxt, stmt)

		og.expectedCommitErrors.merge(og.candidateExpectedCommitErrors)
		break
	}
	__antithesis_instrumentation__.Notify(695910)

	return stmt, err
}

func (og *operationGenerator) checkIfOpViolatesDDLAfterWrite(ot opType) {
	__antithesis_instrumentation__.Notify(695921)
	if ot.isDDL() && func() bool {
		__antithesis_instrumentation__.Notify(695923)
		return og.haveInsertBeforeAnyDDLs() == true
	}() == true {
		__antithesis_instrumentation__.Notify(695924)
		og.expectedExecErrors.add(pgcode.FeatureNotSupported)
	} else {
		__antithesis_instrumentation__.Notify(695925)
	}
	__antithesis_instrumentation__.Notify(695922)
	og.opsInTxn = append(og.opsInTxn, ot)
}

func (og *operationGenerator) haveInsertBeforeAnyDDLs() bool {
	__antithesis_instrumentation__.Notify(695926)
	for _, ot := range og.opsInTxn {
		__antithesis_instrumentation__.Notify(695928)
		if ot.isDDL() {
			__antithesis_instrumentation__.Notify(695930)
			break
		} else {
			__antithesis_instrumentation__.Notify(695931)
		}
		__antithesis_instrumentation__.Notify(695929)
		if ot == insertRow {
			__antithesis_instrumentation__.Notify(695932)
			return true
		} else {
			__antithesis_instrumentation__.Notify(695933)
		}
	}
	__antithesis_instrumentation__.Notify(695927)
	return false
}

func (og *operationGenerator) addColumn(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(695934)

	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(695947)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695948)
	}
	__antithesis_instrumentation__.Notify(695935)

	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(695949)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695950)
	}
	__antithesis_instrumentation__.Notify(695936)
	if !tableExists {
		__antithesis_instrumentation__.Notify(695951)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IrrelevantColumnName string`, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(695952)
	}
	__antithesis_instrumentation__.Notify(695937)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(695953)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695954)
	}
	__antithesis_instrumentation__.Notify(695938)

	columnName, err := og.randColumn(ctx, tx, *tableName, og.pctExisting(false))
	if err != nil {
		__antithesis_instrumentation__.Notify(695955)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695956)
	}
	__antithesis_instrumentation__.Notify(695939)

	typName, typ, err := og.randType(ctx, tx, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(695957)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695958)
	}
	__antithesis_instrumentation__.Notify(695940)

	def := &tree.ColumnTableDef{
		Name: tree.Name(columnName),
		Type: typName,
	}
	def.Nullable.Nullability = tree.Nullability(og.randIntn(1 + int(tree.SilentNull)))

	databaseHasRegionChange, err := og.databaseHasRegionChange(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(695959)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695960)
	}
	__antithesis_instrumentation__.Notify(695941)
	tableIsRegionalByRow, err := og.tableIsRegionalByRow(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(695961)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695962)
	}
	__antithesis_instrumentation__.Notify(695942)

	if !(tableIsRegionalByRow && func() bool {
		__antithesis_instrumentation__.Notify(695963)
		return databaseHasRegionChange == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(695964)
		return og.randIntn(10) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(695965)
		def.Unique.IsUnique = true
	} else {
		__antithesis_instrumentation__.Notify(695966)
	}
	__antithesis_instrumentation__.Notify(695943)

	columnExistsOnTable, err := og.columnExistsOnTable(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(695967)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695968)
	}
	__antithesis_instrumentation__.Notify(695944)
	var hasRows bool
	if tableExists {
		__antithesis_instrumentation__.Notify(695969)
		hasRows, err = og.tableHasRows(ctx, tx, tableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(695970)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(695971)
		}
	} else {
		__antithesis_instrumentation__.Notify(695972)
	}
	__antithesis_instrumentation__.Notify(695945)

	hasAlterPKSchemaChange, err := og.tableHasOngoingAlterPKSchemaChanges(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(695973)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695974)
	}
	__antithesis_instrumentation__.Notify(695946)

	codesWithConditions{
		{code: pgcode.DuplicateColumn, condition: columnExistsOnTable},
		{code: pgcode.UndefinedObject, condition: typ == nil},
		{code: pgcode.NotNullViolation, condition: hasRows && func() bool {
			__antithesis_instrumentation__.Notify(695975)
			return def.Nullable.Nullability == tree.NotNull == true
		}() == true},
		{code: pgcode.FeatureNotSupported, condition: hasAlterPKSchemaChange},

		{
			code: pgcode.FeatureNotSupported,
			condition: def.Unique.IsUnique && func() bool {
				__antithesis_instrumentation__.Notify(695976)
				return typ != nil == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(695977)
				return !colinfo.ColumnTypeIsIndexable(typ) == true
			}() == true,
		},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s`, tableName, tree.Serialize(def)), nil
}

func (og *operationGenerator) addConstraint(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(695978)

	return "", nil
}

func (og *operationGenerator) addUniqueConstraint(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(695979)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(695992)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695993)
	}
	__antithesis_instrumentation__.Notify(695980)
	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(695994)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695995)
	}
	__antithesis_instrumentation__.Notify(695981)
	if !tableExists {
		__antithesis_instrumentation__.Notify(695996)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER TABLE %s ADD CONSTRAINT IrrelevantConstraintName UNIQUE (IrrelevantColumnName)`, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(695997)
	}
	__antithesis_instrumentation__.Notify(695982)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(695998)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695999)
	}
	__antithesis_instrumentation__.Notify(695983)

	columnForConstraint, err := og.randColumnWithMeta(ctx, tx, *tableName, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696000)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696001)
	}
	__antithesis_instrumentation__.Notify(695984)

	constaintName := fmt.Sprintf("%s_%s_unique", tableName.Object(), columnForConstraint.name)

	columnExistsOnTable, err := og.columnExistsOnTable(ctx, tx, tableName, columnForConstraint.name)
	if err != nil {
		__antithesis_instrumentation__.Notify(696002)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696003)
	}
	__antithesis_instrumentation__.Notify(695985)
	constraintExists, err := og.constraintExists(ctx, tx, constaintName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696004)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696005)
	}
	__antithesis_instrumentation__.Notify(695986)

	canApplyConstraint := true
	if columnExistsOnTable {
		__antithesis_instrumentation__.Notify(696006)
		canApplyConstraint, err = og.canApplyUniqueConstraint(ctx, tx, tableName, []string{columnForConstraint.name})
		if err != nil {
			__antithesis_instrumentation__.Notify(696007)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(696008)
		}
	} else {
		__antithesis_instrumentation__.Notify(696009)
	}
	__antithesis_instrumentation__.Notify(695987)

	hasAlterPKSchemaChange, err := og.tableHasOngoingAlterPKSchemaChanges(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696010)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696011)
	}
	__antithesis_instrumentation__.Notify(695988)

	databaseHasRegionChange, err := og.databaseHasRegionChange(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696012)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696013)
	}
	__antithesis_instrumentation__.Notify(695989)
	tableIsRegionalByRow, err := og.tableIsRegionalByRow(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696014)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696015)
	}
	__antithesis_instrumentation__.Notify(695990)

	codesWithConditions{
		{code: pgcode.UndefinedColumn, condition: !columnExistsOnTable},
		{code: pgcode.DuplicateObject, condition: constraintExists},
		{code: pgcode.FeatureNotSupported, condition: columnExistsOnTable && func() bool {
			__antithesis_instrumentation__.Notify(696016)
			return !colinfo.ColumnTypeIsIndexable(columnForConstraint.typ) == true
		}() == true},
		{pgcode.FeatureNotSupported, hasAlterPKSchemaChange},
		{code: pgcode.ObjectNotInPrerequisiteState, condition: databaseHasRegionChange && func() bool {
			__antithesis_instrumentation__.Notify(696017)
			return tableIsRegionalByRow == true
		}() == true},
	}.add(og.expectedExecErrors)

	if !canApplyConstraint {
		__antithesis_instrumentation__.Notify(696018)
		og.candidateExpectedCommitErrors.add(pgcode.UniqueViolation)
	} else {
		__antithesis_instrumentation__.Notify(696019)
	}
	__antithesis_instrumentation__.Notify(695991)

	return fmt.Sprintf(`ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)`, tableName, constaintName, columnForConstraint.name), nil
}

func (og *operationGenerator) alterTableLocality(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696020)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696033)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696034)
	}
	__antithesis_instrumentation__.Notify(696021)
	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696035)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696036)
	}
	__antithesis_instrumentation__.Notify(696022)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696037)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER TABLE %s SET LOCALITY REGIONAL BY ROW`, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696038)
	}
	__antithesis_instrumentation__.Notify(696023)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696039)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696040)
	}
	__antithesis_instrumentation__.Notify(696024)

	databaseRegionNames, err := og.getDatabaseRegionNames(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696041)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696042)
	}
	__antithesis_instrumentation__.Notify(696025)
	if len(databaseRegionNames) == 0 {
		__antithesis_instrumentation__.Notify(696043)
		og.expectedExecErrors.add(pgcode.InvalidTableDefinition)
		return fmt.Sprintf(`ALTER TABLE %s SET LOCALITY REGIONAL BY ROW`, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696044)
	}
	__antithesis_instrumentation__.Notify(696026)

	hasSchemaChange, err := og.tableHasOngoingSchemaChanges(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696045)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696046)
	}
	__antithesis_instrumentation__.Notify(696027)
	databaseHasRegionChange, err := og.databaseHasRegionChange(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696047)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696048)
	}
	__antithesis_instrumentation__.Notify(696028)
	databaseHasMultiRegion, err := og.databaseIsMultiRegion(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696049)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696050)
	}
	__antithesis_instrumentation__.Notify(696029)
	if hasSchemaChange || func() bool {
		__antithesis_instrumentation__.Notify(696051)
		return databaseHasRegionChange == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(696052)
		return !databaseHasMultiRegion == true
	}() == true {
		__antithesis_instrumentation__.Notify(696053)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return `ALTER TABLE invalid_table SET LOCALITY REGIONAL BY ROW`, nil
	} else {
		__antithesis_instrumentation__.Notify(696054)
	}
	__antithesis_instrumentation__.Notify(696030)

	localityOptions := []func() (string, error){
		func() (string, error) {
			__antithesis_instrumentation__.Notify(696055)
			return "REGIONAL BY TABLE", nil
		},
		func() (string, error) {
			__antithesis_instrumentation__.Notify(696056)
			idx := og.params.rng.Intn(len(databaseRegionNames))
			regionName := tree.Name(databaseRegionNames[idx])
			return fmt.Sprintf(`REGIONAL BY TABLE IN %s`, regionName.String()), nil
		},
		func() (string, error) {
			__antithesis_instrumentation__.Notify(696057)
			return "GLOBAL", nil
		},
		func() (string, error) {
			__antithesis_instrumentation__.Notify(696058)
			columnForAs, err := og.randColumnWithMeta(ctx, tx, *tableName, og.alwaysExisting())
			columnForAsUsed := false
			if err != nil {
				__antithesis_instrumentation__.Notify(696062)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696063)
			}
			__antithesis_instrumentation__.Notify(696059)
			ret := "REGIONAL BY ROW"
			if columnForAs.typ.TypeMeta.Name != nil {
				__antithesis_instrumentation__.Notify(696064)
				if columnForAs.typ.TypeMeta.Name.Basename() == tree.RegionEnum && func() bool {
					__antithesis_instrumentation__.Notify(696065)
					return !columnForAs.nullable == true
				}() == true {
					__antithesis_instrumentation__.Notify(696066)
					ret += " AS " + columnForAs.name
					columnForAsUsed = true
				} else {
					__antithesis_instrumentation__.Notify(696067)
				}
			} else {
				__antithesis_instrumentation__.Notify(696068)
			}
			__antithesis_instrumentation__.Notify(696060)

			if !columnForAsUsed {
				__antithesis_instrumentation__.Notify(696069)
				columnNames, err := og.getTableColumns(ctx, tx, tableName.String(), true)
				if err != nil {
					__antithesis_instrumentation__.Notify(696071)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(696072)
				}
				__antithesis_instrumentation__.Notify(696070)
				for _, col := range columnNames {
					__antithesis_instrumentation__.Notify(696073)
					if col.name == tree.RegionalByRowRegionDefaultCol && func() bool {
						__antithesis_instrumentation__.Notify(696074)
						return col.nullable == true
					}() == true {
						__antithesis_instrumentation__.Notify(696075)
						og.expectedExecErrors.add(pgcode.InvalidTableDefinition)
					} else {
						__antithesis_instrumentation__.Notify(696076)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(696077)
			}
			__antithesis_instrumentation__.Notify(696061)

			return ret, nil
		},
	}
	__antithesis_instrumentation__.Notify(696031)
	idx := og.params.rng.Intn(len(localityOptions))
	toLocality, err := localityOptions[idx]()
	if err != nil {
		__antithesis_instrumentation__.Notify(696078)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696079)
	}
	__antithesis_instrumentation__.Notify(696032)
	return fmt.Sprintf(`ALTER TABLE %s SET LOCALITY %s`, tableName, toLocality), nil
}

func (og *operationGenerator) getClusterRegionNames(
	ctx context.Context, tx pgx.Tx,
) (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(696080)
	return og.scanRegionNames(ctx, tx, "SELECT region FROM [SHOW REGIONS FROM CLUSTER]")
}

func (og *operationGenerator) getDatabaseRegionNames(
	ctx context.Context, tx pgx.Tx,
) (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(696081)
	return og.scanRegionNames(ctx, tx, "SELECT region FROM [SHOW REGIONS FROM DATABASE]")
}

func (og *operationGenerator) getDatabase(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696082)
	var database string
	err := tx.QueryRow(ctx, "SHOW DATABASE").Scan(&database)
	og.LogQueryResults("SHOW DATABASE", database)
	return database, err
}

type getRegionsResult struct {
	regionNamesInDatabase catpb.RegionNames
	regionNamesInCluster  catpb.RegionNames

	regionNamesNotInDatabase catpb.RegionNames
}

func (og *operationGenerator) getRegions(ctx context.Context, tx pgx.Tx) (getRegionsResult, error) {
	__antithesis_instrumentation__.Notify(696083)
	regionNamesInCluster, err := og.getClusterRegionNames(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696089)
		return getRegionsResult{}, err
	} else {
		__antithesis_instrumentation__.Notify(696090)
	}
	__antithesis_instrumentation__.Notify(696084)
	regionNamesNotInDatabaseSet := make(map[catpb.RegionName]struct{}, len(regionNamesInCluster))
	for _, clusterRegionName := range regionNamesInCluster {
		__antithesis_instrumentation__.Notify(696091)
		regionNamesNotInDatabaseSet[clusterRegionName] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(696085)
	regionNamesInDatabase, err := og.getDatabaseRegionNames(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696092)
		return getRegionsResult{}, err
	} else {
		__antithesis_instrumentation__.Notify(696093)
	}
	__antithesis_instrumentation__.Notify(696086)
	for _, databaseRegionName := range regionNamesInDatabase {
		__antithesis_instrumentation__.Notify(696094)
		delete(regionNamesNotInDatabaseSet, databaseRegionName)
	}
	__antithesis_instrumentation__.Notify(696087)

	regionNamesNotInDatabase := make(catpb.RegionNames, 0, len(regionNamesNotInDatabaseSet))
	for regionName := range regionNamesNotInDatabaseSet {
		__antithesis_instrumentation__.Notify(696095)
		regionNamesNotInDatabase = append(regionNamesNotInDatabase, regionName)
	}
	__antithesis_instrumentation__.Notify(696088)
	return getRegionsResult{
		regionNamesInDatabase:    regionNamesInDatabase,
		regionNamesInCluster:     regionNamesInCluster,
		regionNamesNotInDatabase: regionNamesNotInDatabase,
	}, nil
}

func (og *operationGenerator) scanRegionNames(
	ctx context.Context, tx pgx.Tx, query string,
) (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(696096)
	var regionNames catpb.RegionNames
	var regionNamesForLog []string
	rows, err := tx.Query(ctx, query)
	if err != nil {
		__antithesis_instrumentation__.Notify(696100)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(696101)
	}
	__antithesis_instrumentation__.Notify(696097)
	defer rows.Close()

	for rows.Next() {
		__antithesis_instrumentation__.Notify(696102)
		var regionName catpb.RegionName
		if err := rows.Scan(&regionName); err != nil {
			__antithesis_instrumentation__.Notify(696104)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(696105)
		}
		__antithesis_instrumentation__.Notify(696103)
		regionNames = append(regionNames, regionName)
		regionNamesForLog = append(regionNamesForLog, regionName.String())
	}
	__antithesis_instrumentation__.Notify(696098)
	if rows.Err() != nil {
		__antithesis_instrumentation__.Notify(696106)
		return nil, errors.Wrapf(rows.Err(), "failed to get regions: %s", query)
	} else {
		__antithesis_instrumentation__.Notify(696107)
	}
	__antithesis_instrumentation__.Notify(696099)
	og.LogQueryResultArray(query, regionNamesForLog)
	return regionNames, nil
}

func (og *operationGenerator) addRegion(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696108)
	regionResult, err := og.getRegions(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696117)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696118)
	}
	__antithesis_instrumentation__.Notify(696109)
	database, err := og.getDatabase(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696119)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696120)
	}
	__antithesis_instrumentation__.Notify(696110)

	if len(regionResult.regionNamesInCluster) == 0 {
		__antithesis_instrumentation__.Notify(696121)
		og.expectedExecErrors.add(pgcode.InvalidDatabaseDefinition)
		return fmt.Sprintf(`ALTER DATABASE %s ADD REGION "invalid-region"`, database), nil
	} else {
		__antithesis_instrumentation__.Notify(696122)
	}
	__antithesis_instrumentation__.Notify(696111)

	if len(regionResult.regionNamesInDatabase) == 0 {
		__antithesis_instrumentation__.Notify(696123)
		idx := og.params.rng.Intn(len(regionResult.regionNamesInCluster))
		og.expectedExecErrors.add(pgcode.InvalidDatabaseDefinition)
		return fmt.Sprintf(
			`ALTER DATABASE %s ADD REGION "%s"`,
			database,
			regionResult.regionNamesInCluster[idx],
		), nil
	} else {
		__antithesis_instrumentation__.Notify(696124)
	}
	__antithesis_instrumentation__.Notify(696112)

	if len(regionResult.regionNamesInDatabase) > 0 {
		__antithesis_instrumentation__.Notify(696125)
		databaseHasRegionalByRowChange, err := og.databaseHasRegionalByRowChange(ctx, tx)
		if err != nil {
			__antithesis_instrumentation__.Notify(696127)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(696128)
		}
		__antithesis_instrumentation__.Notify(696126)
		if databaseHasRegionalByRowChange {
			__antithesis_instrumentation__.Notify(696129)

			og.expectedExecErrors.add(pgcode.InvalidName)
			og.expectedExecErrors.add(pgcode.ObjectNotInPrerequisiteState)
			return fmt.Sprintf(`ALTER DATABASE %s ADD REGION "invalid-region"`, database), nil
		} else {
			__antithesis_instrumentation__.Notify(696130)
		}
	} else {
		__antithesis_instrumentation__.Notify(696131)
	}
	__antithesis_instrumentation__.Notify(696113)

	if len(regionResult.regionNamesNotInDatabase) == 0 {
		__antithesis_instrumentation__.Notify(696132)
		idx := og.params.rng.Intn(len(regionResult.regionNamesInDatabase))
		og.expectedExecErrors.add(pgcode.DuplicateObject)
		return fmt.Sprintf(
			`ALTER DATABASE %s ADD REGION "%s"`,
			database,
			regionResult.regionNamesInDatabase[idx],
		), nil
	} else {
		__antithesis_instrumentation__.Notify(696133)
	}
	__antithesis_instrumentation__.Notify(696114)

	idx := og.params.rng.Intn(len(regionResult.regionNamesNotInDatabase))
	region := regionResult.regionNamesNotInDatabase[idx]
	valuePresent, err := og.enumMemberPresent(ctx, tx, tree.RegionEnum, string(region))
	if err != nil {
		__antithesis_instrumentation__.Notify(696134)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696135)
	}
	__antithesis_instrumentation__.Notify(696115)
	if valuePresent {
		__antithesis_instrumentation__.Notify(696136)
		og.expectedExecErrors.add(pgcode.DuplicateObject)
	} else {
		__antithesis_instrumentation__.Notify(696137)
	}
	__antithesis_instrumentation__.Notify(696116)
	return fmt.Sprintf(
		`ALTER DATABASE %s ADD REGION "%s"`,
		database,
		region,
	), nil
}

func (og *operationGenerator) primaryRegion(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696138)
	regionResult, err := og.getRegions(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696143)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696144)
	}
	__antithesis_instrumentation__.Notify(696139)
	database, err := og.getDatabase(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696145)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696146)
	}
	__antithesis_instrumentation__.Notify(696140)

	if len(regionResult.regionNamesInCluster) == 0 {
		__antithesis_instrumentation__.Notify(696147)
		og.expectedExecErrors.add(pgcode.InvalidDatabaseDefinition)
		return fmt.Sprintf(`ALTER DATABASE %s PRIMARY REGION "invalid-region"`, database), nil
	} else {
		__antithesis_instrumentation__.Notify(696148)
	}
	__antithesis_instrumentation__.Notify(696141)

	if len(regionResult.regionNamesInDatabase) == 0 {
		__antithesis_instrumentation__.Notify(696149)
		idx := og.params.rng.Intn(len(regionResult.regionNamesInCluster))
		return fmt.Sprintf(
			`ALTER DATABASE %s PRIMARY REGION "%s"`,
			database,
			regionResult.regionNamesInCluster[idx],
		), nil
	} else {
		__antithesis_instrumentation__.Notify(696150)
	}
	__antithesis_instrumentation__.Notify(696142)

	idx := og.params.rng.Intn(len(regionResult.regionNamesInDatabase))
	return fmt.Sprintf(
		`ALTER DATABASE %s PRIMARY REGION "%s"`,
		database,
		regionResult.regionNamesInDatabase[idx],
	), nil
}

func (og *operationGenerator) addForeignKeyConstraint(
	ctx context.Context, tx pgx.Tx,
) (string, error) {
	__antithesis_instrumentation__.Notify(696151)

	parentTable, parentColumn, err := og.randParentColumnForFkRelation(ctx, tx, og.randIntn(100) >= og.params.fkParentInvalidPct)
	if err != nil {
		__antithesis_instrumentation__.Notify(696160)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696161)
	}
	__antithesis_instrumentation__.Notify(696152)

	fetchInvalidChild := og.randIntn(100) < og.params.fkChildInvalidPct

	childType := parentColumn.typ
	if fetchInvalidChild {
		__antithesis_instrumentation__.Notify(696162)
		_, typ, err := og.randType(ctx, tx, og.pctExisting(true))
		if err != nil {
			__antithesis_instrumentation__.Notify(696164)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(696165)
		}
		__antithesis_instrumentation__.Notify(696163)
		if typ != nil {
			__antithesis_instrumentation__.Notify(696166)
			childType = typ
		} else {
			__antithesis_instrumentation__.Notify(696167)
		}
	} else {
		__antithesis_instrumentation__.Notify(696168)
	}
	__antithesis_instrumentation__.Notify(696153)

	childTable, childColumn, err := og.randChildColumnForFkRelation(ctx, tx, !fetchInvalidChild, childType.SQLString())
	if err != nil {
		__antithesis_instrumentation__.Notify(696169)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696170)
	}
	__antithesis_instrumentation__.Notify(696154)

	constraintName := tree.Name(fmt.Sprintf("%s_%s_%s_%s_fk", parentTable.Object(), parentColumn.name, childTable.Object(), childColumn.name))

	def := &tree.AlterTable{
		Table: childTable.ToUnresolvedObjectName(),
		Cmds: tree.AlterTableCmds{
			&tree.AlterTableAddConstraint{
				ConstraintDef: &tree.ForeignKeyConstraintTableDef{
					Name:     constraintName,
					Table:    *parentTable,
					FromCols: tree.NameList{tree.Name(childColumn.name)},
					ToCols:   tree.NameList{tree.Name(parentColumn.name)},
					Actions: tree.ReferenceActions{
						Update: tree.Cascade,
						Delete: tree.Cascade,
					},
				},
				ValidationBehavior: tree.ValidationDefault,
			},
		},
	}

	parentColumnHasUniqueConstraint, err := og.columnHasSingleUniqueConstraint(ctx, tx, parentTable, parentColumn.name)
	if err != nil {
		__antithesis_instrumentation__.Notify(696171)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696172)
	}
	__antithesis_instrumentation__.Notify(696155)
	childColumnIsComputed, err := og.columnIsComputed(ctx, tx, parentTable, parentColumn.name)
	if err != nil {
		__antithesis_instrumentation__.Notify(696173)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696174)
	}
	__antithesis_instrumentation__.Notify(696156)
	constraintExists, err := og.constraintExists(ctx, tx, string(constraintName))
	if err != nil {
		__antithesis_instrumentation__.Notify(696175)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696176)
	}
	__antithesis_instrumentation__.Notify(696157)
	rowsSatisfyConstraint, err := og.rowsSatisfyFkConstraint(ctx, tx, parentTable, parentColumn, childTable, childColumn)
	if err != nil {
		__antithesis_instrumentation__.Notify(696177)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696178)
	}
	__antithesis_instrumentation__.Notify(696158)

	codesWithConditions{
		{code: pgcode.ForeignKeyViolation, condition: !parentColumnHasUniqueConstraint},
		{code: pgcode.FeatureNotSupported, condition: childColumnIsComputed},
		{code: pgcode.DuplicateObject, condition: constraintExists},
		{code: pgcode.DatatypeMismatch, condition: !childColumn.typ.Equivalent(parentColumn.typ)},
	}.add(og.expectedExecErrors)

	if !rowsSatisfyConstraint {
		__antithesis_instrumentation__.Notify(696179)
		og.candidateExpectedCommitErrors.add(pgcode.ForeignKeyViolation)
	} else {
		__antithesis_instrumentation__.Notify(696180)
	}
	__antithesis_instrumentation__.Notify(696159)

	return tree.Serialize(def), nil
}

func (og *operationGenerator) createIndex(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696181)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696198)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696199)
	}
	__antithesis_instrumentation__.Notify(696182)

	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696200)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696201)
	}
	__antithesis_instrumentation__.Notify(696183)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696202)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		def := &tree.CreateIndex{
			Name:  tree.Name("IrrelevantName"),
			Table: *tableName,
			Columns: tree.IndexElemList{
				{Column: "IrrelevantColumn", Direction: tree.Ascending},
			},
		}
		return tree.Serialize(def), nil
	} else {
		__antithesis_instrumentation__.Notify(696203)
	}
	__antithesis_instrumentation__.Notify(696184)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696204)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696205)
	}
	__antithesis_instrumentation__.Notify(696185)

	columnNames, err := og.getTableColumns(ctx, tx, tableName.String(), true)
	if err != nil {
		__antithesis_instrumentation__.Notify(696206)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696207)
	}
	__antithesis_instrumentation__.Notify(696186)

	indexName, err := og.randIndex(ctx, tx, *tableName, og.pctExisting(false))
	if err != nil {
		__antithesis_instrumentation__.Notify(696208)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696209)
	}
	__antithesis_instrumentation__.Notify(696187)

	indexExists, err := og.indexExists(ctx, tx, tableName, indexName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696210)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696211)
	}
	__antithesis_instrumentation__.Notify(696188)

	def := &tree.CreateIndex{
		Name:        tree.Name(indexName),
		Table:       *tableName,
		Unique:      og.randIntn(4) == 0,
		Inverted:    og.randIntn(10) == 0,
		IfNotExists: og.randIntn(2) == 0,
	}

	regionColumn := ""
	tableIsRegionalByRow, err := og.tableIsRegionalByRow(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696212)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696213)
	}
	__antithesis_instrumentation__.Notify(696189)
	if tableIsRegionalByRow {
		__antithesis_instrumentation__.Notify(696214)
		regionColumn, err = og.getRegionColumn(ctx, tx, tableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(696215)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(696216)
		}
	} else {
		__antithesis_instrumentation__.Notify(696217)
	}
	__antithesis_instrumentation__.Notify(696190)

	duplicateRegionColumn := false
	nonIndexableType := false
	def.Columns = make(tree.IndexElemList, 1+og.randIntn(len(columnNames)))
	for i := range def.Columns {
		__antithesis_instrumentation__.Notify(696218)
		def.Columns[i].Column = tree.Name(columnNames[i].name)
		def.Columns[i].Direction = tree.Direction(og.randIntn(1 + int(tree.Descending)))

		if columnNames[i].name == regionColumn && func() bool {
			__antithesis_instrumentation__.Notify(696220)
			return i != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(696221)
			duplicateRegionColumn = true
		} else {
			__antithesis_instrumentation__.Notify(696222)
		}
		__antithesis_instrumentation__.Notify(696219)
		if def.Inverted {
			__antithesis_instrumentation__.Notify(696223)

			invertedIndexableType := colinfo.ColumnTypeIsInvertedIndexable(columnNames[i].typ)
			if (invertedIndexableType && func() bool {
				__antithesis_instrumentation__.Notify(696224)
				return i < len(def.Columns)-1 == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(696225)
				return (!invertedIndexableType && func() bool {
					__antithesis_instrumentation__.Notify(696226)
					return i == len(def.Columns)-1 == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(696227)
				nonIndexableType = true
			} else {
				__antithesis_instrumentation__.Notify(696228)
			}
		} else {
			__antithesis_instrumentation__.Notify(696229)
			if !colinfo.ColumnTypeIsIndexable(columnNames[i].typ) {
				__antithesis_instrumentation__.Notify(696230)
				nonIndexableType = true
			} else {
				__antithesis_instrumentation__.Notify(696231)
			}
		}
	}
	__antithesis_instrumentation__.Notify(696191)

	duplicateStore := false
	virtualComputedStored := false
	regionColStored := false
	columnNames = columnNames[len(def.Columns):]
	if n := len(columnNames); n > 0 {
		__antithesis_instrumentation__.Notify(696232)
		def.Storing = make(tree.NameList, og.randIntn(1+n))
		for i := range def.Storing {
			__antithesis_instrumentation__.Notify(696233)
			def.Storing[i] = tree.Name(columnNames[i].name)

			if tableIsRegionalByRow && func() bool {
				__antithesis_instrumentation__.Notify(696236)
				return columnNames[i].name == regionColumn == true
			}() == true {
				__antithesis_instrumentation__.Notify(696237)
				regionColStored = true
			} else {
				__antithesis_instrumentation__.Notify(696238)
			}
			__antithesis_instrumentation__.Notify(696234)

			if columnNames[i].generated && func() bool {
				__antithesis_instrumentation__.Notify(696239)
				return !virtualComputedStored == true
			}() == true {
				__antithesis_instrumentation__.Notify(696240)
				isStored, err := og.columnIsStoredComputed(ctx, tx, tableName, columnNames[i].name)
				if err != nil {
					__antithesis_instrumentation__.Notify(696242)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(696243)
				}
				__antithesis_instrumentation__.Notify(696241)
				if !isStored {
					__antithesis_instrumentation__.Notify(696244)
					virtualComputedStored = true
				} else {
					__antithesis_instrumentation__.Notify(696245)
				}
			} else {
				__antithesis_instrumentation__.Notify(696246)
			}
			__antithesis_instrumentation__.Notify(696235)

			if !duplicateStore {
				__antithesis_instrumentation__.Notify(696247)
				colUsedInPrimaryIdx, err := og.colIsPrimaryKey(ctx, tx, tableName, columnNames[i].name)
				if err != nil {
					__antithesis_instrumentation__.Notify(696249)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(696250)
				}
				__antithesis_instrumentation__.Notify(696248)
				if colUsedInPrimaryIdx {
					__antithesis_instrumentation__.Notify(696251)
					duplicateStore = true
				} else {
					__antithesis_instrumentation__.Notify(696252)
				}
			} else {
				__antithesis_instrumentation__.Notify(696253)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(696254)
	}
	__antithesis_instrumentation__.Notify(696192)

	uniqueViolationWillNotOccur := true
	if def.Unique {
		__antithesis_instrumentation__.Notify(696255)
		columns := []string{}
		for _, col := range def.Columns {
			__antithesis_instrumentation__.Notify(696257)
			columns = append(columns, string(col.Column))
		}
		__antithesis_instrumentation__.Notify(696256)
		uniqueViolationWillNotOccur, err = og.canApplyUniqueConstraint(ctx, tx, tableName, columns)
		if err != nil {
			__antithesis_instrumentation__.Notify(696258)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(696259)
		}
	} else {
		__antithesis_instrumentation__.Notify(696260)
	}
	__antithesis_instrumentation__.Notify(696193)

	hasAlterPKSchemaChange, err := og.tableHasOngoingAlterPKSchemaChanges(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696261)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696262)
	}
	__antithesis_instrumentation__.Notify(696194)

	databaseHasRegionChange, err := og.databaseHasRegionChange(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696263)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696264)
	}
	__antithesis_instrumentation__.Notify(696195)
	if databaseHasRegionChange && func() bool {
		__antithesis_instrumentation__.Notify(696265)
		return tableIsRegionalByRow == true
	}() == true {
		__antithesis_instrumentation__.Notify(696266)
		og.expectedExecErrors.add(pgcode.ObjectNotInPrerequisiteState)
	} else {
		__antithesis_instrumentation__.Notify(696267)
	}
	__antithesis_instrumentation__.Notify(696196)

	if !(indexExists && func() bool {
		__antithesis_instrumentation__.Notify(696268)
		return def.IfNotExists == true
	}() == true) {
		__antithesis_instrumentation__.Notify(696269)
		codesWithConditions{
			{code: pgcode.DuplicateRelation, condition: indexExists},

			{code: pgcode.InvalidSQLStatementName, condition: len(def.Storing) > 0 && func() bool {
				__antithesis_instrumentation__.Notify(696270)
				return def.Inverted == true
			}() == true},

			{code: pgcode.InvalidSQLStatementName, condition: def.Unique && func() bool {
				__antithesis_instrumentation__.Notify(696271)
				return def.Inverted == true
			}() == true},

			{code: pgcode.UniqueViolation, condition: !uniqueViolationWillNotOccur},
			{code: pgcode.DuplicateColumn, condition: duplicateStore},
			{code: pgcode.FeatureNotSupported, condition: nonIndexableType},
			{code: pgcode.FeatureNotSupported, condition: regionColStored},
			{code: pgcode.FeatureNotSupported, condition: duplicateRegionColumn},
			{code: pgcode.Uncategorized, condition: virtualComputedStored},
			{code: pgcode.FeatureNotSupported, condition: hasAlterPKSchemaChange},
		}.add(og.expectedExecErrors)
	} else {
		__antithesis_instrumentation__.Notify(696272)
	}
	__antithesis_instrumentation__.Notify(696197)

	return tree.Serialize(def), nil
}

func (og *operationGenerator) createSequence(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696273)
	seqName, err := og.randSequence(ctx, tx, og.pctExisting(false), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696279)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696280)
	}
	__antithesis_instrumentation__.Notify(696274)

	schemaExists, err := og.schemaExists(ctx, tx, seqName.Schema())
	if err != nil {
		__antithesis_instrumentation__.Notify(696281)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696282)
	}
	__antithesis_instrumentation__.Notify(696275)
	sequenceExists, err := og.sequenceExists(ctx, tx, seqName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696283)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696284)
	}
	__antithesis_instrumentation__.Notify(696276)

	ifNotExists := true
	if sequenceExists && func() bool {
		__antithesis_instrumentation__.Notify(696285)
		return og.produceError() == true
	}() == true {
		__antithesis_instrumentation__.Notify(696286)
		ifNotExists = false
	} else {
		__antithesis_instrumentation__.Notify(696287)
	}
	__antithesis_instrumentation__.Notify(696277)

	codesWithConditions{
		{code: pgcode.UndefinedSchema, condition: !schemaExists},
		{code: pgcode.DuplicateRelation, condition: sequenceExists && func() bool {
			__antithesis_instrumentation__.Notify(696288)
			return !ifNotExists == true
		}() == true},
	}.add(og.expectedExecErrors)

	var seqOptions tree.SequenceOptions

	if og.randIntn(100) < og.params.sequenceOwnedByPct {
		__antithesis_instrumentation__.Notify(696289)
		table, err := og.randTable(ctx, tx, og.pctExisting(true), "")
		if err != nil {
			__antithesis_instrumentation__.Notify(696292)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(696293)
		}
		__antithesis_instrumentation__.Notify(696290)
		tableExists, err := og.tableExists(ctx, tx, table)
		if err != nil {
			__antithesis_instrumentation__.Notify(696294)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(696295)
		}
		__antithesis_instrumentation__.Notify(696291)

		if !tableExists {
			__antithesis_instrumentation__.Notify(696296)
			seqOptions = append(
				seqOptions,
				tree.SequenceOption{
					Name:          tree.SeqOptOwnedBy,
					ColumnItemVal: &tree.ColumnItem{TableName: table.ToUnresolvedObjectName(), ColumnName: "IrrelevantColumnName"}},
			)
			if !(sequenceExists && func() bool {
				__antithesis_instrumentation__.Notify(696297)
				return ifNotExists == true
			}() == true) {
				__antithesis_instrumentation__.Notify(696298)
				og.expectedExecErrors.add(pgcode.UndefinedTable)
			} else {
				__antithesis_instrumentation__.Notify(696299)
			}
		} else {
			__antithesis_instrumentation__.Notify(696300)
			column, err := og.randColumn(ctx, tx, *table, og.pctExisting(true))
			if err != nil {
				__antithesis_instrumentation__.Notify(696304)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696305)
			}
			__antithesis_instrumentation__.Notify(696301)
			columnExists, err := og.columnExistsOnTable(ctx, tx, table, column)
			if err != nil {
				__antithesis_instrumentation__.Notify(696306)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696307)
			}
			__antithesis_instrumentation__.Notify(696302)

			if !columnExists && func() bool {
				__antithesis_instrumentation__.Notify(696308)
				return !sequenceExists == true
			}() == true {
				__antithesis_instrumentation__.Notify(696309)
				og.expectedExecErrors.add(pgcode.UndefinedColumn)
			} else {
				__antithesis_instrumentation__.Notify(696310)
			}
			__antithesis_instrumentation__.Notify(696303)

			seqOptions = append(
				seqOptions,
				tree.SequenceOption{
					Name:          tree.SeqOptOwnedBy,
					ColumnItemVal: &tree.ColumnItem{TableName: table.ToUnresolvedObjectName(), ColumnName: tree.Name(column)}},
			)
		}
	} else {
		__antithesis_instrumentation__.Notify(696311)
	}
	__antithesis_instrumentation__.Notify(696278)

	createSeq := &tree.CreateSequence{
		IfNotExists: ifNotExists,
		Name:        *seqName,
		Options:     seqOptions,
	}

	return tree.Serialize(createSeq), nil
}

func (og *operationGenerator) createTable(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696312)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(false), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696319)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696320)
	}
	__antithesis_instrumentation__.Notify(696313)

	tableIdx, err := strconv.Atoi(strings.TrimPrefix(tableName.Table(), "table"))
	if err != nil {
		__antithesis_instrumentation__.Notify(696321)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696322)
	}
	__antithesis_instrumentation__.Notify(696314)

	stmt := randgen.RandCreateTableWithColumnIndexNumberGenerator(og.params.rng, "table", tableIdx, og.newUniqueSeqNum)
	stmt.Table = *tableName
	stmt.IfNotExists = og.randIntn(2) == 0

	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696323)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696324)
	}
	__antithesis_instrumentation__.Notify(696315)
	schemaExists, err := og.schemaExists(ctx, tx, tableName.Schema())
	if err != nil {
		__antithesis_instrumentation__.Notify(696325)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696326)
	}
	__antithesis_instrumentation__.Notify(696316)

	computedColInIndex := false
	primaryKeyDisallowsComputedCols, err := isClusterVersionLessThan(ctx, tx,
		clusterversion.TestingBinaryMinSupportedVersion)
	if err != nil {
		__antithesis_instrumentation__.Notify(696327)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696328)
	}
	__antithesis_instrumentation__.Notify(696317)
	if primaryKeyDisallowsComputedCols {
		__antithesis_instrumentation__.Notify(696329)
		computedCols := make(map[string]struct{})
		for _, def := range stmt.Defs {
			__antithesis_instrumentation__.Notify(696330)
			if colDef, ok := def.(*tree.ColumnTableDef); ok {
				__antithesis_instrumentation__.Notify(696331)
				if colDef.Computed.Computed {
					__antithesis_instrumentation__.Notify(696332)
					computedCols[colDef.Name.String()] = struct{}{}
				} else {
					__antithesis_instrumentation__.Notify(696333)
				}
			} else {
				__antithesis_instrumentation__.Notify(696334)
				if indexDef, ok := def.(*tree.IndexTableDef); ok {
					__antithesis_instrumentation__.Notify(696335)
					for _, indexCol := range indexDef.Columns {
						__antithesis_instrumentation__.Notify(696336)
						if _, ok := computedCols[indexCol.Column.String()]; ok {
							__antithesis_instrumentation__.Notify(696337)
							computedColInIndex = true
						} else {
							__antithesis_instrumentation__.Notify(696338)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(696339)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(696340)
	}
	__antithesis_instrumentation__.Notify(696318)

	codesWithConditions{
		{code: pgcode.DuplicateRelation, condition: tableExists && func() bool {
			__antithesis_instrumentation__.Notify(696341)
			return !stmt.IfNotExists == true
		}() == true},
		{code: pgcode.UndefinedSchema, condition: !schemaExists},
		{code: pgcode.FeatureNotSupported, condition: computedColInIndex},
	}.add(og.expectedExecErrors)

	return tree.Serialize(stmt), nil
}

func (og *operationGenerator) createEnum(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696342)
	typName, typeExists, err := og.randEnum(ctx, tx, og.pctExisting(false))
	if err != nil {
		__antithesis_instrumentation__.Notify(696345)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696346)
	}
	__antithesis_instrumentation__.Notify(696343)
	schemaExists, err := og.schemaExists(ctx, tx, typName.Schema())
	if err != nil {
		__antithesis_instrumentation__.Notify(696347)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696348)
	}
	__antithesis_instrumentation__.Notify(696344)
	codesWithConditions{
		{code: pgcode.DuplicateObject, condition: typeExists},
		{code: pgcode.InvalidSchemaName, condition: !schemaExists},
	}.add(og.expectedExecErrors)
	stmt := randgen.RandCreateType(og.params.rng, typName.Object(), "asdf")
	stmt.(*tree.CreateType).TypeName = typName.ToUnresolvedObjectName()
	return tree.Serialize(stmt), nil
}

func (og *operationGenerator) createTableAs(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696349)
	numSourceTables := og.randIntn(og.params.maxSourceTables) + 1

	sourceTableNames := make([]tree.TableExpr, numSourceTables)
	sourceTableExistence := make([]bool, numSourceTables)

	uniqueTableNames := map[string]bool{}
	duplicateSourceTables := false

	for i := 0; i < numSourceTables; i++ {
		__antithesis_instrumentation__.Notify(696355)
		var tableName *tree.TableName
		var err error
		var sourceTableExists bool

		switch randInt := og.randIntn(1); randInt {
		case 0:
			__antithesis_instrumentation__.Notify(696357)
			tableName, err = og.randTable(ctx, tx, og.pctExisting(true), "")
			if err != nil {
				__antithesis_instrumentation__.Notify(696362)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696363)
			}
			__antithesis_instrumentation__.Notify(696358)
			sourceTableExists, err = og.tableExists(ctx, tx, tableName)
			if err != nil {
				__antithesis_instrumentation__.Notify(696364)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696365)
			}

		case 1:
			__antithesis_instrumentation__.Notify(696359)
			tableName, err = og.randView(ctx, tx, og.pctExisting(true), "")
			if err != nil {
				__antithesis_instrumentation__.Notify(696366)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696367)
			}
			__antithesis_instrumentation__.Notify(696360)
			sourceTableExists, err = og.viewExists(ctx, tx, tableName)
			if err != nil {
				__antithesis_instrumentation__.Notify(696368)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696369)
			}
		default:
			__antithesis_instrumentation__.Notify(696361)
		}
		__antithesis_instrumentation__.Notify(696356)

		sourceTableNames[i] = tableName
		sourceTableExistence[i] = sourceTableExists
		if _, exists := uniqueTableNames[tableName.String()]; exists {
			__antithesis_instrumentation__.Notify(696370)
			duplicateSourceTables = true
		} else {
			__antithesis_instrumentation__.Notify(696371)
			uniqueTableNames[tableName.String()] = true
		}
	}
	__antithesis_instrumentation__.Notify(696350)

	selectStatement := tree.SelectClause{
		From: tree.From{Tables: sourceTableNames},
	}

	uniqueColumnNames := map[string]bool{}
	duplicateColumns := false
	for i := 0; i < numSourceTables; i++ {
		__antithesis_instrumentation__.Notify(696372)
		tableName := sourceTableNames[i]
		tableExists := sourceTableExistence[i]

		if tableExists {
			__antithesis_instrumentation__.Notify(696373)
			columnNamesForTable, err := og.tableColumnsShuffled(ctx, tx, tableName.(*tree.TableName).String())
			if err != nil {
				__antithesis_instrumentation__.Notify(696375)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696376)
			}
			__antithesis_instrumentation__.Notify(696374)
			columnNamesForTable = columnNamesForTable[:1+og.randIntn(len(columnNamesForTable))]

			for j := range columnNamesForTable {
				__antithesis_instrumentation__.Notify(696377)
				colItem := tree.ColumnItem{
					TableName:  tableName.(*tree.TableName).ToUnresolvedObjectName(),
					ColumnName: tree.Name(columnNamesForTable[j]),
				}
				selectStatement.Exprs = append(selectStatement.Exprs, tree.SelectExpr{Expr: &colItem})

				if _, exists := uniqueColumnNames[columnNamesForTable[j]]; exists {
					__antithesis_instrumentation__.Notify(696378)
					duplicateColumns = true
				} else {
					__antithesis_instrumentation__.Notify(696379)
					uniqueColumnNames[columnNamesForTable[j]] = true
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(696380)
			og.expectedExecErrors.add(pgcode.UndefinedTable)
			colItem := tree.ColumnItem{
				ColumnName: tree.Name("IrrelevantColumnName"),
			}
			selectStatement.Exprs = append(selectStatement.Exprs, tree.SelectExpr{Expr: &colItem})
		}
	}
	__antithesis_instrumentation__.Notify(696351)

	destTableName, err := og.randTable(ctx, tx, og.pctExisting(false), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696381)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696382)
	}
	__antithesis_instrumentation__.Notify(696352)
	schemaExists, err := og.schemaExists(ctx, tx, destTableName.Schema())
	if err != nil {
		__antithesis_instrumentation__.Notify(696383)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696384)
	}
	__antithesis_instrumentation__.Notify(696353)
	tableExists, err := og.tableExists(ctx, tx, destTableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696385)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696386)
	}
	__antithesis_instrumentation__.Notify(696354)

	codesWithConditions{
		{code: pgcode.InvalidSchemaName, condition: !schemaExists},
		{code: pgcode.DuplicateRelation, condition: tableExists},
		{code: pgcode.Syntax, condition: len(selectStatement.Exprs) == 0},
		{code: pgcode.DuplicateAlias, condition: duplicateSourceTables},
		{code: pgcode.DuplicateColumn, condition: duplicateColumns},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`CREATE TABLE %s AS %s`,
		destTableName, selectStatement.String()), nil
}

func (og *operationGenerator) createView(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696387)

	numSourceTables := og.randIntn(og.params.maxSourceTables) + 1

	sourceTableNames := make([]tree.TableExpr, numSourceTables)
	sourceTableExistence := make([]bool, numSourceTables)

	uniqueTableNames := map[string]bool{}
	duplicateSourceTables := false

	for i := 0; i < numSourceTables; i++ {
		__antithesis_instrumentation__.Notify(696393)
		var tableName *tree.TableName
		var err error
		var sourceTableExists bool

		switch randInt := og.randIntn(1); randInt {
		case 0:
			__antithesis_instrumentation__.Notify(696395)
			tableName, err = og.randTable(ctx, tx, og.pctExisting(true), "")
			if err != nil {
				__antithesis_instrumentation__.Notify(696400)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696401)
			}
			__antithesis_instrumentation__.Notify(696396)
			sourceTableExists, err = og.tableExists(ctx, tx, tableName)
			if err != nil {
				__antithesis_instrumentation__.Notify(696402)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696403)
			}

		case 1:
			__antithesis_instrumentation__.Notify(696397)
			tableName, err = og.randView(ctx, tx, og.pctExisting(true), "")
			if err != nil {
				__antithesis_instrumentation__.Notify(696404)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696405)
			}
			__antithesis_instrumentation__.Notify(696398)
			sourceTableExists, err = og.viewExists(ctx, tx, tableName)
			if err != nil {
				__antithesis_instrumentation__.Notify(696406)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696407)
			}
		default:
			__antithesis_instrumentation__.Notify(696399)
		}
		__antithesis_instrumentation__.Notify(696394)

		sourceTableNames[i] = tableName
		sourceTableExistence[i] = sourceTableExists
		if _, exists := uniqueTableNames[tableName.String()]; exists {
			__antithesis_instrumentation__.Notify(696408)
			duplicateSourceTables = true
		} else {
			__antithesis_instrumentation__.Notify(696409)
			uniqueTableNames[tableName.String()] = true
		}
	}
	__antithesis_instrumentation__.Notify(696388)

	selectStatement := tree.SelectClause{
		From: tree.From{Tables: sourceTableNames},
	}

	uniqueColumnNames := map[string]bool{}
	duplicateColumns := false
	for i := 0; i < numSourceTables; i++ {
		__antithesis_instrumentation__.Notify(696410)
		tableName := sourceTableNames[i]
		tableExists := sourceTableExistence[i]

		if tableExists {
			__antithesis_instrumentation__.Notify(696411)
			columnNamesForTable, err := og.tableColumnsShuffled(ctx, tx, tableName.(*tree.TableName).String())
			if err != nil {
				__antithesis_instrumentation__.Notify(696413)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(696414)
			}
			__antithesis_instrumentation__.Notify(696412)
			columnNamesForTable = columnNamesForTable[:1+og.randIntn(len(columnNamesForTable))]

			for j := range columnNamesForTable {
				__antithesis_instrumentation__.Notify(696415)
				colItem := tree.ColumnItem{
					TableName:  tableName.(*tree.TableName).ToUnresolvedObjectName(),
					ColumnName: tree.Name(columnNamesForTable[j]),
				}
				selectStatement.Exprs = append(selectStatement.Exprs, tree.SelectExpr{Expr: &colItem})

				if _, exists := uniqueColumnNames[columnNamesForTable[j]]; exists {
					__antithesis_instrumentation__.Notify(696416)
					duplicateColumns = true
				} else {
					__antithesis_instrumentation__.Notify(696417)
					uniqueColumnNames[columnNamesForTable[j]] = true
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(696418)
			og.expectedExecErrors.add(pgcode.UndefinedTable)
			colItem := tree.ColumnItem{
				ColumnName: tree.Name("IrrelevantColumnName"),
			}
			selectStatement.Exprs = append(selectStatement.Exprs, tree.SelectExpr{Expr: &colItem})
		}
	}
	__antithesis_instrumentation__.Notify(696389)

	destViewName, err := og.randView(ctx, tx, og.pctExisting(false), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696419)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696420)
	}
	__antithesis_instrumentation__.Notify(696390)
	schemaExists, err := og.schemaExists(ctx, tx, destViewName.Schema())
	if err != nil {
		__antithesis_instrumentation__.Notify(696421)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696422)
	}
	__antithesis_instrumentation__.Notify(696391)
	viewExists, err := og.viewExists(ctx, tx, destViewName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696423)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696424)
	}
	__antithesis_instrumentation__.Notify(696392)

	codesWithConditions{
		{code: pgcode.InvalidSchemaName, condition: !schemaExists},
		{code: pgcode.DuplicateRelation, condition: viewExists},
		{code: pgcode.Syntax, condition: len(selectStatement.Exprs) == 0},
		{code: pgcode.DuplicateAlias, condition: duplicateSourceTables},
		{code: pgcode.DuplicateColumn, condition: duplicateColumns},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`CREATE VIEW %s AS %s`,
		destViewName, selectStatement.String()), nil
}

func (og *operationGenerator) dropColumn(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696425)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696436)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696437)
	}
	__antithesis_instrumentation__.Notify(696426)

	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696438)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696439)
	}
	__antithesis_instrumentation__.Notify(696427)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696440)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER TABLE %s DROP COLUMN "IrrelevantColumnName"`, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696441)
	}
	__antithesis_instrumentation__.Notify(696428)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696442)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696443)
	}
	__antithesis_instrumentation__.Notify(696429)

	columnName, err := og.randColumn(ctx, tx, *tableName, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696444)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696445)
	}
	__antithesis_instrumentation__.Notify(696430)
	columnExists, err := og.columnExistsOnTable(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696446)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696447)
	}
	__antithesis_instrumentation__.Notify(696431)
	colIsPrimaryKey, err := og.colIsPrimaryKey(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696448)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696449)
	}
	__antithesis_instrumentation__.Notify(696432)
	columnIsDependedOn, err := og.columnIsDependedOn(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696450)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696451)
	}
	__antithesis_instrumentation__.Notify(696433)
	columnIsInDroppingIndex, err := og.columnIsInDroppingIndex(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696452)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696453)
	}
	__antithesis_instrumentation__.Notify(696434)
	hasAlterPKSchemaChange, err := og.tableHasOngoingAlterPKSchemaChanges(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696454)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696455)
	}
	__antithesis_instrumentation__.Notify(696435)

	codesWithConditions{
		{code: pgcode.ObjectNotInPrerequisiteState, condition: columnIsInDroppingIndex},
		{code: pgcode.UndefinedColumn, condition: !columnExists},
		{code: pgcode.InvalidColumnReference, condition: colIsPrimaryKey},
		{code: pgcode.DependentObjectsStillExist, condition: columnIsDependedOn},
		{code: pgcode.FeatureNotSupported, condition: hasAlterPKSchemaChange},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`ALTER TABLE %s DROP COLUMN "%s"`, tableName, columnName), nil
}

func (og *operationGenerator) dropColumnDefault(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696456)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696464)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696465)
	}
	__antithesis_instrumentation__.Notify(696457)
	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696466)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696467)
	}
	__antithesis_instrumentation__.Notify(696458)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696468)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN "IrrelevantColumnName" DROP DEFAULT`, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696469)
	}
	__antithesis_instrumentation__.Notify(696459)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696470)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696471)
	}
	__antithesis_instrumentation__.Notify(696460)
	columnName, err := og.randColumn(ctx, tx, *tableName, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696472)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696473)
	}
	__antithesis_instrumentation__.Notify(696461)
	columnExists, err := og.columnExistsOnTable(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696474)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696475)
	}
	__antithesis_instrumentation__.Notify(696462)
	if !columnExists {
		__antithesis_instrumentation__.Notify(696476)
		og.expectedExecErrors.add(pgcode.UndefinedColumn)
	} else {
		__antithesis_instrumentation__.Notify(696477)
	}
	__antithesis_instrumentation__.Notify(696463)
	return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN "%s" DROP DEFAULT`, tableName, columnName), nil
}

func (og *operationGenerator) dropColumnNotNull(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696478)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696488)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696489)
	}
	__antithesis_instrumentation__.Notify(696479)
	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696490)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696491)
	}
	__antithesis_instrumentation__.Notify(696480)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696492)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN "IrrelevantColumnName" DROP NOT NULL`, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696493)
	}
	__antithesis_instrumentation__.Notify(696481)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696494)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696495)
	}
	__antithesis_instrumentation__.Notify(696482)
	columnName, err := og.randColumn(ctx, tx, *tableName, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696496)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696497)
	}
	__antithesis_instrumentation__.Notify(696483)
	columnExists, err := og.columnExistsOnTable(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696498)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696499)
	}
	__antithesis_instrumentation__.Notify(696484)
	colIsPrimaryKey, err := og.colIsPrimaryKey(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696500)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696501)
	}
	__antithesis_instrumentation__.Notify(696485)

	hasAlterPKSchemaChange, err := og.tableHasOngoingAlterPKSchemaChanges(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696502)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696503)
	}
	__antithesis_instrumentation__.Notify(696486)
	if hasAlterPKSchemaChange {
		__antithesis_instrumentation__.Notify(696504)

		return `SELECT 'avoiding timing hole'`, nil
	} else {
		__antithesis_instrumentation__.Notify(696505)
	}
	__antithesis_instrumentation__.Notify(696487)

	codesWithConditions{
		{pgcode.UndefinedColumn, !columnExists},
		{pgcode.InvalidTableDefinition, colIsPrimaryKey},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN "%s" DROP NOT NULL`, tableName, columnName), nil
}

func (og *operationGenerator) dropColumnStored(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696506)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696514)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696515)
	}
	__antithesis_instrumentation__.Notify(696507)
	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696516)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696517)
	}
	__antithesis_instrumentation__.Notify(696508)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696518)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN IrrelevantColumnName DROP STORED`, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696519)
	}
	__antithesis_instrumentation__.Notify(696509)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696520)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696521)
	}
	__antithesis_instrumentation__.Notify(696510)

	columnName, err := og.randColumn(ctx, tx, *tableName, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696522)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696523)
	}
	__antithesis_instrumentation__.Notify(696511)
	columnExists, err := og.columnExistsOnTable(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696524)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696525)
	}
	__antithesis_instrumentation__.Notify(696512)

	columnIsStored, err := og.columnIsStoredComputed(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696526)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696527)
	}
	__antithesis_instrumentation__.Notify(696513)

	codesWithConditions{
		{code: pgcode.InvalidColumnDefinition, condition: !columnIsStored},
		{code: pgcode.UndefinedColumn, condition: !columnExists},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN "%s" DROP STORED`, tableName, columnName), nil
}

func (og *operationGenerator) dropConstraint(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696528)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696540)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696541)
	}
	__antithesis_instrumentation__.Notify(696529)

	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696542)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696543)
	}
	__antithesis_instrumentation__.Notify(696530)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696544)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER TABLE %s DROP CONSTRAINT IrrelevantConstraintName`, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696545)
	}
	__antithesis_instrumentation__.Notify(696531)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696546)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696547)
	}
	__antithesis_instrumentation__.Notify(696532)

	constraintName, err := og.randConstraint(ctx, tx, tableName.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(696548)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696549)
	}
	__antithesis_instrumentation__.Notify(696533)

	constraintIsPrimary, err := og.constraintIsPrimary(ctx, tx, tableName, constraintName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696550)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696551)
	}
	__antithesis_instrumentation__.Notify(696534)
	if constraintIsPrimary {
		__antithesis_instrumentation__.Notify(696552)
		og.candidateExpectedCommitErrors.add(pgcode.FeatureNotSupported)
	} else {
		__antithesis_instrumentation__.Notify(696553)
	}
	__antithesis_instrumentation__.Notify(696535)

	constraintIsUnique, err := og.constraintIsUnique(ctx, tx, tableName, constraintName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696554)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696555)
	}
	__antithesis_instrumentation__.Notify(696536)
	if constraintIsUnique {
		__antithesis_instrumentation__.Notify(696556)
		og.expectedExecErrors.add(pgcode.FeatureNotSupported)
	} else {
		__antithesis_instrumentation__.Notify(696557)
	}
	__antithesis_instrumentation__.Notify(696537)

	constraintBeingDropped, err := og.constraintInDroppingState(ctx, tx, tableName, constraintName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696558)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696559)
	}
	__antithesis_instrumentation__.Notify(696538)
	if constraintBeingDropped {
		__antithesis_instrumentation__.Notify(696560)
		og.expectedExecErrors.add(pgcode.FeatureNotSupported)
	} else {
		__antithesis_instrumentation__.Notify(696561)
	}
	__antithesis_instrumentation__.Notify(696539)

	return fmt.Sprintf(`ALTER TABLE %s DROP CONSTRAINT "%s"`, tableName, constraintName), nil
}

func (og *operationGenerator) dropIndex(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696562)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696575)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696576)
	}
	__antithesis_instrumentation__.Notify(696563)
	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696577)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696578)
	}
	__antithesis_instrumentation__.Notify(696564)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696579)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`DROP INDEX %s@"IrrelevantIndexName"`, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696580)
	}
	__antithesis_instrumentation__.Notify(696565)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696581)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696582)
	}
	__antithesis_instrumentation__.Notify(696566)

	indexName, err := og.randIndex(ctx, tx, *tableName, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696583)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696584)
	}
	__antithesis_instrumentation__.Notify(696567)

	indexExists, err := og.indexExists(ctx, tx, tableName, indexName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696585)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696586)
	}
	__antithesis_instrumentation__.Notify(696568)
	if !indexExists {
		__antithesis_instrumentation__.Notify(696587)
		og.expectedExecErrors.add(pgcode.UndefinedObject)
	} else {
		__antithesis_instrumentation__.Notify(696588)
	}
	__antithesis_instrumentation__.Notify(696569)

	hasAlterPKSchemaChange, err := og.tableHasOngoingAlterPKSchemaChanges(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696589)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696590)
	}
	__antithesis_instrumentation__.Notify(696570)
	if hasAlterPKSchemaChange {
		__antithesis_instrumentation__.Notify(696591)
		og.expectedExecErrors.add(pgcode.FeatureNotSupported)
	} else {
		__antithesis_instrumentation__.Notify(696592)
	}
	__antithesis_instrumentation__.Notify(696571)

	databaseHasRegionChange, err := og.databaseHasRegionChange(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696593)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696594)
	}
	__antithesis_instrumentation__.Notify(696572)
	tableIsRegionalByRow, err := og.tableIsRegionalByRow(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696595)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696596)
	}
	__antithesis_instrumentation__.Notify(696573)
	if databaseHasRegionChange && func() bool {
		__antithesis_instrumentation__.Notify(696597)
		return tableIsRegionalByRow == true
	}() == true {
		__antithesis_instrumentation__.Notify(696598)
		og.expectedExecErrors.add(pgcode.ObjectNotInPrerequisiteState)
	} else {
		__antithesis_instrumentation__.Notify(696599)
	}
	__antithesis_instrumentation__.Notify(696574)

	return fmt.Sprintf(`DROP INDEX %s@"%s" CASCADE`, tableName, indexName), nil
}

func (og *operationGenerator) dropSequence(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696600)
	sequenceName, err := og.randSequence(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696604)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696605)
	}
	__antithesis_instrumentation__.Notify(696601)
	ifExists := og.randIntn(2) == 0
	dropSeq := &tree.DropSequence{
		Names:    tree.TableNames{*sequenceName},
		IfExists: ifExists,
	}

	sequenceExists, err := og.sequenceExists(ctx, tx, sequenceName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696606)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696607)
	}
	__antithesis_instrumentation__.Notify(696602)
	if !sequenceExists && func() bool {
		__antithesis_instrumentation__.Notify(696608)
		return !ifExists == true
	}() == true {
		__antithesis_instrumentation__.Notify(696609)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
	} else {
		__antithesis_instrumentation__.Notify(696610)
	}
	__antithesis_instrumentation__.Notify(696603)
	return tree.Serialize(dropSeq), nil
}

func (og *operationGenerator) dropTable(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696611)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696615)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696616)
	}
	__antithesis_instrumentation__.Notify(696612)
	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696617)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696618)
	}
	__antithesis_instrumentation__.Notify(696613)
	tableHasDependencies, err := og.tableHasDependencies(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696619)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696620)
	}
	__antithesis_instrumentation__.Notify(696614)

	dropBehavior := tree.DropBehavior(og.randIntn(3))

	ifExists := og.randIntn(2) == 0
	dropTable := tree.DropTable{
		Names:        []tree.TableName{*tableName},
		IfExists:     ifExists,
		DropBehavior: dropBehavior,
	}

	codesWithConditions{
		{pgcode.UndefinedTable, !ifExists && func() bool {
			__antithesis_instrumentation__.Notify(696621)
			return !tableExists == true
		}() == true},
		{pgcode.DependentObjectsStillExist, dropBehavior != tree.DropCascade && func() bool {
			__antithesis_instrumentation__.Notify(696622)
			return tableHasDependencies == true
		}() == true},
	}.add(og.expectedExecErrors)

	return dropTable.String(), nil
}

func (og *operationGenerator) dropView(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696623)
	viewName, err := og.randView(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696627)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696628)
	}
	__antithesis_instrumentation__.Notify(696624)
	viewExists, err := og.tableExists(ctx, tx, viewName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696629)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696630)
	}
	__antithesis_instrumentation__.Notify(696625)
	viewHasDependencies, err := og.tableHasDependencies(ctx, tx, viewName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696631)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696632)
	}
	__antithesis_instrumentation__.Notify(696626)

	dropBehavior := tree.DropBehavior(og.randIntn(3))

	ifExists := og.randIntn(2) == 0
	dropView := tree.DropView{
		Names:        []tree.TableName{*viewName},
		IfExists:     ifExists,
		DropBehavior: dropBehavior,
	}

	codesWithConditions{
		{pgcode.UndefinedTable, !ifExists && func() bool {
			__antithesis_instrumentation__.Notify(696633)
			return !viewExists == true
		}() == true},
		{pgcode.DependentObjectsStillExist, dropBehavior != tree.DropCascade && func() bool {
			__antithesis_instrumentation__.Notify(696634)
			return viewHasDependencies == true
		}() == true},
	}.add(og.expectedExecErrors)
	return dropView.String(), nil
}

func (og *operationGenerator) renameColumn(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696635)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696645)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696646)
	}
	__antithesis_instrumentation__.Notify(696636)

	srcTableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696647)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696648)
	}
	__antithesis_instrumentation__.Notify(696637)
	if !srcTableExists {
		__antithesis_instrumentation__.Notify(696649)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER TABLE %s RENAME COLUMN "IrrelevantColumnName" TO "OtherIrrelevantName"`,
			tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696650)
	}
	__antithesis_instrumentation__.Notify(696638)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696651)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696652)
	}
	__antithesis_instrumentation__.Notify(696639)

	srcColumnName, err := og.randColumn(ctx, tx, *tableName, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696653)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696654)
	}
	__antithesis_instrumentation__.Notify(696640)

	destColumnName, err := og.randColumn(ctx, tx, *tableName, og.pctExisting(false))
	if err != nil {
		__antithesis_instrumentation__.Notify(696655)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696656)
	}
	__antithesis_instrumentation__.Notify(696641)

	srcColumnExists, err := og.columnExistsOnTable(ctx, tx, tableName, srcColumnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696657)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696658)
	}
	__antithesis_instrumentation__.Notify(696642)
	destColumnExists, err := og.columnExistsOnTable(ctx, tx, tableName, destColumnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696659)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696660)
	}
	__antithesis_instrumentation__.Notify(696643)
	columnIsDependedOn, err := og.columnIsDependedOn(ctx, tx, tableName, srcColumnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696661)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696662)
	}
	__antithesis_instrumentation__.Notify(696644)

	codesWithConditions{
		{pgcode.UndefinedColumn, !srcColumnExists},
		{pgcode.DuplicateColumn, destColumnExists && func() bool {
			__antithesis_instrumentation__.Notify(696663)
			return srcColumnName != destColumnName == true
		}() == true},
		{pgcode.DependentObjectsStillExist, columnIsDependedOn},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`ALTER TABLE %s RENAME COLUMN "%s" TO "%s"`,
		tableName, srcColumnName, destColumnName), nil
}

func (og *operationGenerator) renameIndex(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696664)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696673)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696674)
	}
	__antithesis_instrumentation__.Notify(696665)

	srcTableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696675)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696676)
	}
	__antithesis_instrumentation__.Notify(696666)
	if !srcTableExists {
		__antithesis_instrumentation__.Notify(696677)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER INDEX %s@"IrrelevantConstraintName" RENAME TO "OtherConstraintName"`,
			tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696678)
	}
	__antithesis_instrumentation__.Notify(696667)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696679)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696680)
	}
	__antithesis_instrumentation__.Notify(696668)

	srcIndexName, err := og.randIndex(ctx, tx, *tableName, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696681)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696682)
	}
	__antithesis_instrumentation__.Notify(696669)

	destIndexName, err := og.randIndex(ctx, tx, *tableName, og.pctExisting(false))
	if err != nil {
		__antithesis_instrumentation__.Notify(696683)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696684)
	}
	__antithesis_instrumentation__.Notify(696670)

	srcIndexExists, err := og.indexExists(ctx, tx, tableName, srcIndexName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696685)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696686)
	}
	__antithesis_instrumentation__.Notify(696671)
	destIndexExists, err := og.indexExists(ctx, tx, tableName, destIndexName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696687)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696688)
	}
	__antithesis_instrumentation__.Notify(696672)

	codesWithConditions{
		{code: pgcode.UndefinedObject, condition: !srcIndexExists},
		{code: pgcode.DuplicateRelation, condition: destIndexExists && func() bool {
			__antithesis_instrumentation__.Notify(696689)
			return srcIndexName != destIndexName == true
		}() == true},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`ALTER INDEX %s@"%s" RENAME TO "%s"`,
		tableName, srcIndexName, destIndexName), nil
}

func (og *operationGenerator) renameSequence(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696690)
	srcSequenceName, err := og.randSequence(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696697)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696698)
	}
	__antithesis_instrumentation__.Notify(696691)

	desiredSchema := ""
	if !og.produceError() {
		__antithesis_instrumentation__.Notify(696699)
		desiredSchema = srcSequenceName.Schema()
	} else {
		__antithesis_instrumentation__.Notify(696700)
	}
	__antithesis_instrumentation__.Notify(696692)

	destSequenceName, err := og.randSequence(ctx, tx, og.pctExisting(false), desiredSchema)
	if err != nil {
		__antithesis_instrumentation__.Notify(696701)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696702)
	}
	__antithesis_instrumentation__.Notify(696693)

	srcSequenceExists, err := og.sequenceExists(ctx, tx, srcSequenceName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696703)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696704)
	}
	__antithesis_instrumentation__.Notify(696694)

	destSchemaExists, err := og.schemaExists(ctx, tx, destSequenceName.Schema())
	if err != nil {
		__antithesis_instrumentation__.Notify(696705)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696706)
	}
	__antithesis_instrumentation__.Notify(696695)

	destSequenceExists, err := og.sequenceExists(ctx, tx, destSequenceName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696707)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696708)
	}
	__antithesis_instrumentation__.Notify(696696)

	srcEqualsDest := srcSequenceName.String() == destSequenceName.String()
	codesWithConditions{
		{code: pgcode.UndefinedTable, condition: !srcSequenceExists},
		{code: pgcode.UndefinedSchema, condition: !destSchemaExists},
		{code: pgcode.DuplicateRelation, condition: !srcEqualsDest && func() bool {
			__antithesis_instrumentation__.Notify(696709)
			return destSequenceExists == true
		}() == true},
		{code: pgcode.InvalidName, condition: srcSequenceName.Schema() != destSequenceName.Schema()},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`ALTER SEQUENCE %s RENAME TO %s`, srcSequenceName, destSequenceName), nil
}

func (og *operationGenerator) renameTable(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696710)
	srcTableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696719)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696720)
	}
	__antithesis_instrumentation__.Notify(696711)

	desiredSchema := ""
	if !og.produceError() {
		__antithesis_instrumentation__.Notify(696721)
		desiredSchema = srcTableName.SchemaName.String()
	} else {
		__antithesis_instrumentation__.Notify(696722)
	}
	__antithesis_instrumentation__.Notify(696712)
	destTableName, err := og.randTable(ctx, tx, og.pctExisting(false), desiredSchema)
	if err != nil {
		__antithesis_instrumentation__.Notify(696723)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696724)
	}
	__antithesis_instrumentation__.Notify(696713)

	srcTableExists, err := og.tableExists(ctx, tx, srcTableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696725)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696726)
	}
	__antithesis_instrumentation__.Notify(696714)
	if srcTableExists {
		__antithesis_instrumentation__.Notify(696727)
		err = og.tableHasPrimaryKeySwapActive(ctx, tx, srcTableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(696728)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(696729)
		}
	} else {
		__antithesis_instrumentation__.Notify(696730)
	}
	__antithesis_instrumentation__.Notify(696715)

	destSchemaExists, err := og.schemaExists(ctx, tx, destTableName.Schema())
	if err != nil {
		__antithesis_instrumentation__.Notify(696731)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696732)
	}
	__antithesis_instrumentation__.Notify(696716)

	destTableExists, err := og.tableExists(ctx, tx, destTableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696733)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696734)
	}
	__antithesis_instrumentation__.Notify(696717)

	srcTableHasDependencies, err := og.tableHasDependencies(ctx, tx, srcTableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696735)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696736)
	}
	__antithesis_instrumentation__.Notify(696718)

	srcEqualsDest := destTableName.String() == srcTableName.String()
	codesWithConditions{
		{code: pgcode.UndefinedTable, condition: !srcTableExists},
		{code: pgcode.UndefinedSchema, condition: !destSchemaExists},
		{code: pgcode.DuplicateRelation, condition: !srcEqualsDest && func() bool {
			__antithesis_instrumentation__.Notify(696737)
			return destTableExists == true
		}() == true},
		{code: pgcode.DependentObjectsStillExist, condition: srcTableHasDependencies},
		{code: pgcode.InvalidName, condition: srcTableName.Schema() != destTableName.Schema()},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`ALTER TABLE %s RENAME TO %s`, srcTableName, destTableName), nil
}

func (og *operationGenerator) renameView(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696738)
	srcViewName, err := og.randView(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696746)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696747)
	}
	__antithesis_instrumentation__.Notify(696739)

	desiredSchema := ""
	if !og.produceError() {
		__antithesis_instrumentation__.Notify(696748)
		desiredSchema = srcViewName.SchemaName.String()
	} else {
		__antithesis_instrumentation__.Notify(696749)
	}
	__antithesis_instrumentation__.Notify(696740)
	destViewName, err := og.randView(ctx, tx, og.pctExisting(false), desiredSchema)
	if err != nil {
		__antithesis_instrumentation__.Notify(696750)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696751)
	}
	__antithesis_instrumentation__.Notify(696741)

	srcViewExists, err := og.viewExists(ctx, tx, srcViewName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696752)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696753)
	}
	__antithesis_instrumentation__.Notify(696742)

	destSchemaExists, err := og.schemaExists(ctx, tx, destViewName.Schema())
	if err != nil {
		__antithesis_instrumentation__.Notify(696754)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696755)
	}
	__antithesis_instrumentation__.Notify(696743)

	destViewExists, err := og.viewExists(ctx, tx, destViewName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696756)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696757)
	}
	__antithesis_instrumentation__.Notify(696744)

	srcTableHasDependencies, err := og.tableHasDependencies(ctx, tx, srcViewName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696758)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696759)
	}
	__antithesis_instrumentation__.Notify(696745)

	srcEqualsDest := destViewName.String() == srcViewName.String()
	codesWithConditions{
		{code: pgcode.UndefinedTable, condition: !srcViewExists},
		{code: pgcode.UndefinedSchema, condition: !destSchemaExists},
		{code: pgcode.DuplicateRelation, condition: !srcEqualsDest && func() bool {
			__antithesis_instrumentation__.Notify(696760)
			return destViewExists == true
		}() == true},
		{code: pgcode.DependentObjectsStillExist, condition: srcTableHasDependencies},
		{code: pgcode.InvalidName, condition: srcViewName.Schema() != destViewName.Schema()},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`ALTER VIEW %s RENAME TO %s`, srcViewName, destViewName), nil
}

func (og *operationGenerator) setColumnDefault(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696761)

	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696772)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696773)
	}
	__antithesis_instrumentation__.Notify(696762)

	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696774)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696775)
	}
	__antithesis_instrumentation__.Notify(696763)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696776)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN IrrelevantColumnName SET DEFAULT "IrrelevantValue"`,
			tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696777)
	}
	__antithesis_instrumentation__.Notify(696764)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696778)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696779)
	}
	__antithesis_instrumentation__.Notify(696765)

	columnForDefault, err := og.randColumnWithMeta(ctx, tx, *tableName, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696780)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696781)
	}
	__antithesis_instrumentation__.Notify(696766)
	columnExists, err := og.columnExistsOnTable(ctx, tx, tableName, columnForDefault.name)
	if err != nil {
		__antithesis_instrumentation__.Notify(696782)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696783)
	}
	__antithesis_instrumentation__.Notify(696767)
	if !columnExists {
		__antithesis_instrumentation__.Notify(696784)
		og.expectedExecErrors.add(pgcode.UndefinedColumn)
		return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN %s SET DEFAULT "IrrelevantValue"`,
			tableName, columnForDefault.name), nil
	} else {
		__antithesis_instrumentation__.Notify(696785)
	}
	__antithesis_instrumentation__.Notify(696768)

	datumTyp := columnForDefault.typ

	if og.produceError() {
		__antithesis_instrumentation__.Notify(696786)
		newTypeName, newTyp, err := og.randType(ctx, tx, og.pctExisting(true))
		if err != nil {
			__antithesis_instrumentation__.Notify(696789)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(696790)
		}
		__antithesis_instrumentation__.Notify(696787)
		if newTyp == nil {
			__antithesis_instrumentation__.Notify(696791)
			og.expectedExecErrors.add(pgcode.UndefinedObject)
			return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN %s SET DEFAULT 'IrrelevantValue':::%s`, tableName, columnForDefault.name, newTypeName.SQLString()), nil
		} else {
			__antithesis_instrumentation__.Notify(696792)
		}
		__antithesis_instrumentation__.Notify(696788)
		datumTyp = newTyp
	} else {
		__antithesis_instrumentation__.Notify(696793)
	}
	__antithesis_instrumentation__.Notify(696769)

	defaultDatum := randgen.RandDatum(og.params.rng, datumTyp, columnForDefault.nullable)

	if (!datumTyp.Equivalent(columnForDefault.typ)) && func() bool {
		__antithesis_instrumentation__.Notify(696794)
		return defaultDatum != tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(696795)
		og.expectedExecErrors.add(pgcode.DatatypeMismatch)
	} else {
		__antithesis_instrumentation__.Notify(696796)
	}
	__antithesis_instrumentation__.Notify(696770)

	if columnForDefault.generated {
		__antithesis_instrumentation__.Notify(696797)
		og.expectedExecErrors.add(pgcode.InvalidTableDefinition)
	} else {
		__antithesis_instrumentation__.Notify(696798)
	}
	__antithesis_instrumentation__.Notify(696771)

	return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s`, tableName, columnForDefault.name, tree.AsStringWithFlags(defaultDatum, tree.FmtParsable)), nil
}

func (og *operationGenerator) setColumnNotNull(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696799)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696811)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696812)
	}
	__antithesis_instrumentation__.Notify(696800)

	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696813)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696814)
	}
	__antithesis_instrumentation__.Notify(696801)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696815)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN IrrelevantColumnName SET NOT NULL`, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696816)
	}
	__antithesis_instrumentation__.Notify(696802)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696817)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696818)
	}
	__antithesis_instrumentation__.Notify(696803)

	columnName, err := og.randColumn(ctx, tx, *tableName, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696819)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696820)
	}
	__antithesis_instrumentation__.Notify(696804)
	columnExists, err := og.columnExistsOnTable(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696821)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696822)
	}
	__antithesis_instrumentation__.Notify(696805)
	constraintBeingAdded, err := og.columnNotNullConstraintInMutation(ctx, tx, tableName, columnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696823)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696824)
	}
	__antithesis_instrumentation__.Notify(696806)
	if constraintBeingAdded {
		__antithesis_instrumentation__.Notify(696825)
		og.expectedExecErrors.add(pgcode.ObjectNotInPrerequisiteState)
	} else {
		__antithesis_instrumentation__.Notify(696826)
	}
	__antithesis_instrumentation__.Notify(696807)

	if !columnExists {
		__antithesis_instrumentation__.Notify(696827)
		og.expectedExecErrors.add(pgcode.UndefinedColumn)
	} else {
		__antithesis_instrumentation__.Notify(696828)

		colContainsNull, err := og.columnContainsNull(ctx, tx, tableName, columnName)
		if err != nil {
			__antithesis_instrumentation__.Notify(696830)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(696831)
		}
		__antithesis_instrumentation__.Notify(696829)
		if colContainsNull {
			__antithesis_instrumentation__.Notify(696832)
			og.candidateExpectedCommitErrors.add(pgcode.CheckViolation)
		} else {
			__antithesis_instrumentation__.Notify(696833)
		}
	}
	__antithesis_instrumentation__.Notify(696808)

	hasPKSchemaChanges, err := og.tableHasOngoingAlterPKSchemaChanges(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696834)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696835)
	}
	__antithesis_instrumentation__.Notify(696809)
	if hasPKSchemaChanges {
		__antithesis_instrumentation__.Notify(696836)

		return `SELECT 'avoiding timing hole'`, nil
	} else {
		__antithesis_instrumentation__.Notify(696837)
	}
	__antithesis_instrumentation__.Notify(696810)

	return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN "%s" SET NOT NULL`, tableName, columnName), nil
}

func (og *operationGenerator) setColumnType(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696838)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696849)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696850)
	}
	__antithesis_instrumentation__.Notify(696839)

	const setSessionVariableString = `SET enable_experimental_alter_column_type_general = true;`

	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696851)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696852)
	}
	__antithesis_instrumentation__.Notify(696840)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696853)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(`%s ALTER TABLE %s ALTER COLUMN IrrelevantColumnName SET DATA TYPE IrrelevantDataType`, setSessionVariableString, tableName), nil
	} else {
		__antithesis_instrumentation__.Notify(696854)
	}
	__antithesis_instrumentation__.Notify(696841)
	err = og.tableHasPrimaryKeySwapActive(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696855)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696856)
	}
	__antithesis_instrumentation__.Notify(696842)

	columnForTypeChange, err := og.randColumnWithMeta(ctx, tx, *tableName, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696857)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696858)
	}
	__antithesis_instrumentation__.Notify(696843)

	columnExists, err := og.columnExistsOnTable(ctx, tx, tableName, columnForTypeChange.name)
	if err != nil {
		__antithesis_instrumentation__.Notify(696859)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696860)
	}
	__antithesis_instrumentation__.Notify(696844)
	if !columnExists {
		__antithesis_instrumentation__.Notify(696861)
		og.expectedExecErrors.add(pgcode.UndefinedColumn)
		return fmt.Sprintf(`%s ALTER TABLE %s ALTER COLUMN "%s" SET DATA TYPE IrrelevantTypeName`,
			setSessionVariableString, tableName, columnForTypeChange.name), nil
	} else {
		__antithesis_instrumentation__.Notify(696862)
	}
	__antithesis_instrumentation__.Notify(696845)

	newTypeName, newType, err := og.randType(ctx, tx, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(696863)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696864)
	}
	__antithesis_instrumentation__.Notify(696846)

	columnHasDependencies, err := og.columnIsDependedOn(ctx, tx, tableName, columnForTypeChange.name)
	if err != nil {
		__antithesis_instrumentation__.Notify(696865)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696866)
	}
	__antithesis_instrumentation__.Notify(696847)

	if newType != nil {
		__antithesis_instrumentation__.Notify(696867)

		kind, _ := schemachange.ClassifyConversion(context.Background(), columnForTypeChange.typ, newType)
		codesWithConditions{
			{code: pgcode.CannotCoerce, condition: kind == schemachange.ColumnConversionImpossible},
			{code: pgcode.FeatureNotSupported, condition: kind != schemachange.ColumnConversionTrivial},
		}.add(og.expectedExecErrors)
	} else {
		__antithesis_instrumentation__.Notify(696868)
	}
	__antithesis_instrumentation__.Notify(696848)

	codesWithConditions{
		{code: pgcode.UndefinedObject, condition: newType == nil},
		{code: pgcode.DependentObjectsStillExist, condition: columnHasDependencies},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`%s ALTER TABLE %s ALTER COLUMN "%s" SET DATA TYPE %s`,
		setSessionVariableString, tableName, columnForTypeChange.name, newTypeName.SQLString()), nil
}

func (og *operationGenerator) survive(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696869)
	dbRegions, err := og.getDatabaseRegionNames(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696873)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696874)
	}
	__antithesis_instrumentation__.Notify(696870)

	needsAtLeastThreeRegions := false
	survive := "ZONE FAILURE"
	if coinToss := og.randIntn(2); coinToss == 1 {
		__antithesis_instrumentation__.Notify(696875)
		survive = "REGION FAILURE"
		needsAtLeastThreeRegions = true
	} else {
		__antithesis_instrumentation__.Notify(696876)
	}
	__antithesis_instrumentation__.Notify(696871)

	codesWithConditions{
		{
			code:      pgcode.InvalidName,
			condition: len(dbRegions) == 0,
		},
		{
			code: pgcode.InvalidParameterValue,
			condition: needsAtLeastThreeRegions && func() bool {
				__antithesis_instrumentation__.Notify(696877)
				return len(dbRegions) < 3 == true
			}() == true,
		},
	}.add(og.expectedExecErrors)

	dbName, err := og.getDatabase(ctx, tx)
	if err != nil {
		__antithesis_instrumentation__.Notify(696878)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696879)
	}
	__antithesis_instrumentation__.Notify(696872)
	return fmt.Sprintf(`ALTER DATABASE %s SURVIVE %s`, dbName, survive), nil
}

func (og *operationGenerator) insertRow(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696880)
	tableName, err := og.randTable(ctx, tx, og.pctExisting(true), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(696890)
		return "", errors.Wrapf(err, "error getting random table name")
	} else {
		__antithesis_instrumentation__.Notify(696891)
	}
	__antithesis_instrumentation__.Notify(696881)
	tableExists, err := og.tableExists(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696892)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696893)
	}
	__antithesis_instrumentation__.Notify(696882)
	if !tableExists {
		__antithesis_instrumentation__.Notify(696894)
		og.expectedExecErrors.add(pgcode.UndefinedTable)
		return fmt.Sprintf(
			`INSERT INTO %s (IrrelevantColumnName) VALUES ("IrrelevantValue")`,
			tableName,
		), nil
	} else {
		__antithesis_instrumentation__.Notify(696895)
	}
	__antithesis_instrumentation__.Notify(696883)
	cols, err := og.getTableColumns(ctx, tx, tableName.String(), false)
	if err != nil {
		__antithesis_instrumentation__.Notify(696896)
		return "", errors.Wrapf(err, "error getting table columns for insert row")
	} else {
		__antithesis_instrumentation__.Notify(696897)
	}

	{
		__antithesis_instrumentation__.Notify(696898)
		truncated := cols[:0]
		for _, c := range cols {
			__antithesis_instrumentation__.Notify(696900)
			if !c.generated {
				__antithesis_instrumentation__.Notify(696901)
				truncated = append(truncated, c)
			} else {
				__antithesis_instrumentation__.Notify(696902)
			}
		}
		__antithesis_instrumentation__.Notify(696899)
		cols = truncated
	}
	__antithesis_instrumentation__.Notify(696884)
	colNames := []string{}
	rows := [][]string{}
	for _, col := range cols {
		__antithesis_instrumentation__.Notify(696903)
		colNames = append(colNames, col.name)
	}
	__antithesis_instrumentation__.Notify(696885)
	numRows := og.randIntn(3) + 1
	for i := 0; i < numRows; i++ {
		__antithesis_instrumentation__.Notify(696904)
		var row []string
		for _, col := range cols {
			__antithesis_instrumentation__.Notify(696906)
			d := randgen.RandDatum(og.params.rng, col.typ, col.nullable)
			row = append(row, tree.AsStringWithFlags(d, tree.FmtParsable))
		}
		__antithesis_instrumentation__.Notify(696905)

		rows = append(rows, row)
	}
	__antithesis_instrumentation__.Notify(696886)

	uniqueConstraintViolation, err := og.violatesUniqueConstraints(ctx, tx, tableName, colNames, rows)
	if err != nil {
		__antithesis_instrumentation__.Notify(696907)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696908)
	}
	__antithesis_instrumentation__.Notify(696887)

	foreignKeyViolation, err := og.violatesFkConstraints(ctx, tx, tableName, colNames, rows)
	if err != nil {
		__antithesis_instrumentation__.Notify(696909)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696910)
	}
	__antithesis_instrumentation__.Notify(696888)

	codesWithConditions{
		{code: pgcode.UniqueViolation, condition: uniqueConstraintViolation},
		{code: pgcode.ForeignKeyViolation, condition: foreignKeyViolation},
	}.add(og.expectedExecErrors)

	formattedRows := []string{}
	for _, row := range rows {
		__antithesis_instrumentation__.Notify(696911)
		formattedRows = append(formattedRows, fmt.Sprintf("(%s)", strings.Join(row, ",")))
	}
	__antithesis_instrumentation__.Notify(696889)

	return fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		tableName,
		strings.Join(colNames, ","),
		strings.Join(formattedRows, ","),
	), nil
}

func (og *operationGenerator) validate(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(696912)

	validateStmt := "SELECT 'validating all objects', crdb_internal.validate_multi_region_zone_configs()"
	rows, err := tx.Query(ctx, `SELECT * FROM "".crdb_internal.invalid_objects ORDER BY id`)
	if err != nil {
		__antithesis_instrumentation__.Notify(696917)
		return validateStmt, err
	} else {
		__antithesis_instrumentation__.Notify(696918)
	}
	__antithesis_instrumentation__.Notify(696913)
	defer rows.Close()

	var errs []string
	for rows.Next() {
		__antithesis_instrumentation__.Notify(696919)
		var id int64
		var dbName, schemaName, objName, errStr string
		if err := rows.Scan(&id, &dbName, &schemaName, &objName, &errStr); err != nil {
			__antithesis_instrumentation__.Notify(696921)
			return validateStmt, err
		} else {
			__antithesis_instrumentation__.Notify(696922)
		}
		__antithesis_instrumentation__.Notify(696920)
		errs = append(
			errs,
			fmt.Sprintf("id %d, db %s, schema %s, name %s: %s", id, dbName, schemaName, objName, errStr),
		)
	}
	__antithesis_instrumentation__.Notify(696914)

	if rows.Err() != nil {
		__antithesis_instrumentation__.Notify(696923)
		return "", errors.Wrap(rows.Err(), "querying for validation errors failed")
	} else {
		__antithesis_instrumentation__.Notify(696924)
	}
	__antithesis_instrumentation__.Notify(696915)

	if len(errs) == 0 {
		__antithesis_instrumentation__.Notify(696925)
		return validateStmt, nil
	} else {
		__antithesis_instrumentation__.Notify(696926)
	}
	__antithesis_instrumentation__.Notify(696916)
	return validateStmt, errors.Errorf("Validation FAIL:\n%s", strings.Join(errs, "\n"))
}

type column struct {
	name      string
	typ       *types.T
	nullable  bool
	generated bool
}

func (og *operationGenerator) getTableColumns(
	ctx context.Context, tx pgx.Tx, tableName string, shuffle bool,
) ([]column, error) {
	__antithesis_instrumentation__.Notify(696927)
	q := fmt.Sprintf(`
SELECT column_name,
       data_type,
       is_nullable,
       generation_expression != '' AS is_generated
  FROM [SHOW COLUMNS FROM %s];
`, tableName)
	rows, err := tx.Query(ctx, q)
	if err != nil {
		__antithesis_instrumentation__.Notify(696934)
		return nil, errors.Wrapf(err, "getting table columns from %s", tableName)
	} else {
		__antithesis_instrumentation__.Notify(696935)
	}
	__antithesis_instrumentation__.Notify(696928)
	defer rows.Close()
	var typNames []string
	var ret []column
	for rows.Next() {
		__antithesis_instrumentation__.Notify(696936)
		var c column
		var typName string
		err := rows.Scan(&c.name, &typName, &c.nullable, &c.generated)
		if err != nil {
			__antithesis_instrumentation__.Notify(696938)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(696939)
		}
		__antithesis_instrumentation__.Notify(696937)
		if c.name != "rowid" {
			__antithesis_instrumentation__.Notify(696940)
			typNames = append(typNames, typName)
			ret = append(ret, c)
		} else {
			__antithesis_instrumentation__.Notify(696941)
		}
	}
	__antithesis_instrumentation__.Notify(696929)
	if err := rows.Err(); err != nil {
		__antithesis_instrumentation__.Notify(696942)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(696943)
	}
	__antithesis_instrumentation__.Notify(696930)
	if len(ret) == 0 {
		__antithesis_instrumentation__.Notify(696944)
		return nil, pgx.ErrNoRows
	} else {
		__antithesis_instrumentation__.Notify(696945)
	}
	__antithesis_instrumentation__.Notify(696931)
	for i := range ret {
		__antithesis_instrumentation__.Notify(696946)
		c := &ret[i]
		c.typ, err = og.typeFromTypeName(ctx, tx, typNames[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(696947)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(696948)
		}
	}
	__antithesis_instrumentation__.Notify(696932)

	if shuffle {
		__antithesis_instrumentation__.Notify(696949)
		og.params.rng.Shuffle(len(ret), func(i, j int) {
			__antithesis_instrumentation__.Notify(696950)
			ret[i], ret[j] = ret[j], ret[i]
		})
	} else {
		__antithesis_instrumentation__.Notify(696951)
	}
	__antithesis_instrumentation__.Notify(696933)

	return ret, nil
}

func (og *operationGenerator) randColumn(
	ctx context.Context, tx pgx.Tx, tableName tree.TableName, pctExisting int,
) (string, error) {
	__antithesis_instrumentation__.Notify(696952)
	if og.randIntn(100) >= pctExisting {
		__antithesis_instrumentation__.Notify(696955)

		return fmt.Sprintf("col%s_%d",
			strings.TrimPrefix(tableName.Table(), "table"), og.newUniqueSeqNum()), nil
	} else {
		__antithesis_instrumentation__.Notify(696956)
	}
	__antithesis_instrumentation__.Notify(696953)
	q := fmt.Sprintf(`
  SELECT column_name
    FROM [SHOW COLUMNS FROM %s]
   WHERE column_name != 'rowid'
ORDER BY random()
   LIMIT 1;
`, tableName.String())
	var name string
	if err := tx.QueryRow(ctx, q).Scan(&name); err != nil {
		__antithesis_instrumentation__.Notify(696957)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696958)
	}
	__antithesis_instrumentation__.Notify(696954)
	return name, nil
}

func (og *operationGenerator) randColumnWithMeta(
	ctx context.Context, tx pgx.Tx, tableName tree.TableName, pctExisting int,
) (column, error) {
	__antithesis_instrumentation__.Notify(696959)
	if og.randIntn(100) >= pctExisting {
		__antithesis_instrumentation__.Notify(696963)

		return column{
			name: fmt.Sprintf("col%s_%d",
				strings.TrimPrefix(tableName.Table(), "table"), og.newUniqueSeqNum()),
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(696964)
	}
	__antithesis_instrumentation__.Notify(696960)
	q := fmt.Sprintf(`
 SELECT 
column_name, 
data_type, 
is_nullable, 
generation_expression != '' AS is_generated
   FROM [SHOW COLUMNS FROM %s]
  WHERE column_name != 'rowid'
ORDER BY random()
  LIMIT 1;
`, tableName.String())
	var col column
	var typ string
	if err := tx.QueryRow(ctx, q).Scan(&col.name, &typ, &col.nullable, &col.generated); err != nil {
		__antithesis_instrumentation__.Notify(696965)
		return column{}, errors.Wrapf(err, "randColumnWithMeta: %q", q)
	} else {
		__antithesis_instrumentation__.Notify(696966)
	}
	__antithesis_instrumentation__.Notify(696961)

	var err error
	col.typ, err = og.typeFromTypeName(ctx, tx, typ)
	if err != nil {
		__antithesis_instrumentation__.Notify(696967)
		return column{}, err
	} else {
		__antithesis_instrumentation__.Notify(696968)
	}
	__antithesis_instrumentation__.Notify(696962)

	return col, nil
}

func (og *operationGenerator) randChildColumnForFkRelation(
	ctx context.Context, tx pgx.Tx, isNotComputed bool, typ string,
) (*tree.TableName, *column, error) {
	__antithesis_instrumentation__.Notify(696969)

	query := strings.Builder{}
	query.WriteString(`
    SELECT table_schema, table_name, column_name, crdb_sql_type, is_nullable
      FROM information_schema.columns
		 WHERE table_name ~ 'table[0-9]+'
  `)
	query.WriteString(fmt.Sprintf(`
			AND crdb_sql_type = '%s'
	`, typ))

	if isNotComputed {
		__antithesis_instrumentation__.Notify(696973)
		query.WriteString(`AND is_generated = 'NO'`)
	} else {
		__antithesis_instrumentation__.Notify(696974)
		query.WriteString(`AND is_generated = 'YES'`)
	}
	__antithesis_instrumentation__.Notify(696970)

	var tableSchema string
	var tableName string
	var columnName string
	var typName string
	var nullable string

	err := tx.QueryRow(ctx, query.String()).Scan(&tableSchema, &tableName, &columnName, &typName, &nullable)
	if err != nil {
		__antithesis_instrumentation__.Notify(696975)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(696976)
	}
	__antithesis_instrumentation__.Notify(696971)

	columnToReturn := column{
		name:     columnName,
		nullable: nullable == "YES",
	}
	table := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
		SchemaName:     tree.Name(tableSchema),
		ExplicitSchema: true,
	}, tree.Name(tableName))

	columnToReturn.typ, err = og.typeFromTypeName(ctx, tx, typName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696977)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(696978)
	}
	__antithesis_instrumentation__.Notify(696972)

	return &table, &columnToReturn, nil
}

func (og *operationGenerator) randParentColumnForFkRelation(
	ctx context.Context, tx pgx.Tx, unique bool,
) (*tree.TableName, *column, error) {
	__antithesis_instrumentation__.Notify(696979)

	subQuery := strings.Builder{}
	subQuery.WriteString(`
		SELECT table_schema, table_name, column_name, crdb_sql_type, is_nullable, contype, conkey
      FROM (
        SELECT table_schema, table_name, column_name, crdb_sql_type, is_nullable, ordinal_position,
               concat(table_schema, '.', table_name)::REGCLASS::INT8 AS tableid
          FROM information_schema.columns
           ) AS cols
		  JOIN (
		        SELECT contype, conkey, conrelid
		          FROM pg_catalog.pg_constraint
		       ) AS cons ON cons.conrelid = cols.tableid
		 WHERE table_name ~ 'table[0-9]+'
  `)
	if unique {
		__antithesis_instrumentation__.Notify(696983)
		subQuery.WriteString(`
		 AND (contype = 'u' OR contype = 'p')
		 AND array_length(conkey, 1) = 1
		 AND conkey[1] = ordinal_position
		`)
	} else {
		__antithesis_instrumentation__.Notify(696984)
	}
	__antithesis_instrumentation__.Notify(696980)

	subQuery.WriteString(`
		ORDER BY random()
    LIMIT 1
  `)

	var tableSchema string
	var tableName string
	var columnName string
	var typName string
	var nullable string

	err := tx.QueryRow(ctx, fmt.Sprintf(`
	SELECT table_schema, table_name, column_name, crdb_sql_type, is_nullable FROM (
		%s
	)`, subQuery.String())).Scan(&tableSchema, &tableName, &columnName, &typName, &nullable)
	if err != nil {
		__antithesis_instrumentation__.Notify(696985)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(696986)
	}
	__antithesis_instrumentation__.Notify(696981)

	columnToReturn := column{
		name:     columnName,
		nullable: nullable == "YES",
	}
	table := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
		SchemaName:     tree.Name(tableSchema),
		ExplicitSchema: true,
	}, tree.Name(tableName))

	columnToReturn.typ, err = og.typeFromTypeName(ctx, tx, typName)
	if err != nil {
		__antithesis_instrumentation__.Notify(696987)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(696988)
	}
	__antithesis_instrumentation__.Notify(696982)

	return &table, &columnToReturn, nil
}

func (og *operationGenerator) randConstraint(
	ctx context.Context, tx pgx.Tx, tableName string,
) (string, error) {
	__antithesis_instrumentation__.Notify(696989)
	q := fmt.Sprintf(`
  SELECT constraint_name
    FROM [SHOW CONSTRAINTS FROM %s]
ORDER BY random()
   LIMIT 1;
`, tableName)
	var name string
	err := tx.QueryRow(ctx, q).Scan(&name)
	if err != nil {
		__antithesis_instrumentation__.Notify(696991)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696992)
	}
	__antithesis_instrumentation__.Notify(696990)
	return name, nil
}

func (og *operationGenerator) randIndex(
	ctx context.Context, tx pgx.Tx, tableName tree.TableName, pctExisting int,
) (string, error) {
	__antithesis_instrumentation__.Notify(696993)
	if og.randIntn(100) >= pctExisting {
		__antithesis_instrumentation__.Notify(696996)

		return fmt.Sprintf("index%s_%d",
			strings.TrimPrefix(tableName.Table(), "table"), og.newUniqueSeqNum()), nil
	} else {
		__antithesis_instrumentation__.Notify(696997)
	}
	__antithesis_instrumentation__.Notify(696994)
	q := fmt.Sprintf(`
  SELECT index_name
    FROM [SHOW INDEXES FROM %s]
	WHERE index_name LIKE 'index%%'
ORDER BY random()
   LIMIT 1;
`, tableName.String())
	var name string
	if err := tx.QueryRow(ctx, q).Scan(&name); err != nil {
		__antithesis_instrumentation__.Notify(696998)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(696999)
	}
	__antithesis_instrumentation__.Notify(696995)
	return name, nil
}

func (og *operationGenerator) randSequence(
	ctx context.Context, tx pgx.Tx, pctExisting int, desiredSchema string,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(697000)

	if desiredSchema != "" {
		__antithesis_instrumentation__.Notify(697004)
		if og.randIntn(100) >= pctExisting {
			__antithesis_instrumentation__.Notify(697007)
			treeSeqName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
				SchemaName:     tree.Name(desiredSchema),
				ExplicitSchema: true,
			}, tree.Name(fmt.Sprintf("seq%d", og.newUniqueSeqNum())))
			return &treeSeqName, nil
		} else {
			__antithesis_instrumentation__.Notify(697008)
		}
		__antithesis_instrumentation__.Notify(697005)
		q := fmt.Sprintf(`
   SELECT sequence_name
     FROM [SHOW SEQUENCES]
    WHERE sequence_name LIKE 'seq%%'
			AND sequence_schema = '%s'
 ORDER BY random()
		LIMIT 1;
		`, desiredSchema)

		var seqName string
		if err := tx.QueryRow(ctx, q).Scan(&seqName); err != nil {
			__antithesis_instrumentation__.Notify(697009)
			treeSeqName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{}, "")
			return &treeSeqName, err
		} else {
			__antithesis_instrumentation__.Notify(697010)
		}
		__antithesis_instrumentation__.Notify(697006)

		treeSeqName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
			SchemaName:     tree.Name(desiredSchema),
			ExplicitSchema: true,
		}, tree.Name(seqName))
		return &treeSeqName, nil
	} else {
		__antithesis_instrumentation__.Notify(697011)
	}
	__antithesis_instrumentation__.Notify(697001)

	if og.randIntn(100) >= pctExisting {
		__antithesis_instrumentation__.Notify(697012)

		randSchema, err := og.randSchema(ctx, tx, og.pctExisting(true))
		if err != nil {
			__antithesis_instrumentation__.Notify(697014)
			treeSeqName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{}, "")
			return &treeSeqName, err
		} else {
			__antithesis_instrumentation__.Notify(697015)
		}
		__antithesis_instrumentation__.Notify(697013)
		treeSeqName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
			SchemaName:     tree.Name(randSchema),
			ExplicitSchema: true,
		}, tree.Name(fmt.Sprintf("seq%d", og.newUniqueSeqNum())))
		return &treeSeqName, nil
	} else {
		__antithesis_instrumentation__.Notify(697016)
	}
	__antithesis_instrumentation__.Notify(697002)

	q := `
   SELECT sequence_schema, sequence_name
     FROM [SHOW SEQUENCES]
    WHERE sequence_name LIKE 'seq%%'
 ORDER BY random()
		LIMIT 1;
		`

	var schemaName string
	var seqName string
	if err := tx.QueryRow(ctx, q).Scan(&schemaName, &seqName); err != nil {
		__antithesis_instrumentation__.Notify(697017)
		treeTableName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{}, "")
		return &treeTableName, err
	} else {
		__antithesis_instrumentation__.Notify(697018)
	}
	__antithesis_instrumentation__.Notify(697003)

	treeSeqName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
		SchemaName:     tree.Name(schemaName),
		ExplicitSchema: true,
	}, tree.Name(seqName))
	return &treeSeqName, nil

}

func (og *operationGenerator) randEnum(
	ctx context.Context, tx pgx.Tx, pctExisting int,
) (name *tree.TypeName, exists bool, _ error) {
	__antithesis_instrumentation__.Notify(697019)
	if og.randIntn(100) >= pctExisting {
		__antithesis_instrumentation__.Notify(697022)

		randSchema, err := og.randSchema(ctx, tx, og.pctExisting(true))
		if err != nil {
			__antithesis_instrumentation__.Notify(697024)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(697025)
		}
		__antithesis_instrumentation__.Notify(697023)
		typeName := tree.MakeSchemaQualifiedTypeName(randSchema, fmt.Sprintf("enum%d", og.newUniqueSeqNum()))
		return &typeName, false, nil
	} else {
		__antithesis_instrumentation__.Notify(697026)
	}
	__antithesis_instrumentation__.Notify(697020)
	const q = `
  SELECT schema, name
    FROM [SHOW ENUMS]
   WHERE name LIKE 'enum%'
ORDER BY random()
   LIMIT 1;
`
	var schemaName string
	var typName string
	if err := tx.QueryRow(ctx, q).Scan(&schemaName, &typName); err != nil {
		__antithesis_instrumentation__.Notify(697027)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(697028)
	}
	__antithesis_instrumentation__.Notify(697021)
	typeName := tree.MakeSchemaQualifiedTypeName(schemaName, typName)
	return &typeName, true, nil
}

func (og *operationGenerator) randTable(
	ctx context.Context, tx pgx.Tx, pctExisting int, desiredSchema string,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(697029)

	if desiredSchema != "" {
		__antithesis_instrumentation__.Notify(697033)
		if og.randIntn(100) >= pctExisting {
			__antithesis_instrumentation__.Notify(697036)
			treeTableName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
				SchemaName:     tree.Name(desiredSchema),
				ExplicitSchema: true,
			}, tree.Name(fmt.Sprintf("table%d", og.newUniqueSeqNum())))
			return &treeTableName, nil
		} else {
			__antithesis_instrumentation__.Notify(697037)
		}
		__antithesis_instrumentation__.Notify(697034)
		q := fmt.Sprintf(`
		  SELECT table_name
		    FROM [SHOW TABLES]
		   WHERE table_name ~ 'table[0-9]+'
				 AND schema_name = '%s'
		ORDER BY random()
		   LIMIT 1;
		`, desiredSchema)

		var tableName string
		if err := tx.QueryRow(ctx, q).Scan(&tableName); err != nil {
			__antithesis_instrumentation__.Notify(697038)
			treeTableName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{}, "")
			return &treeTableName, err
		} else {
			__antithesis_instrumentation__.Notify(697039)
		}
		__antithesis_instrumentation__.Notify(697035)

		treeTableName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
			SchemaName:     tree.Name(desiredSchema),
			ExplicitSchema: true,
		}, tree.Name(tableName))
		return &treeTableName, nil
	} else {
		__antithesis_instrumentation__.Notify(697040)
	}
	__antithesis_instrumentation__.Notify(697030)

	if og.randIntn(100) >= pctExisting {
		__antithesis_instrumentation__.Notify(697041)

		randSchema, err := og.randSchema(ctx, tx, og.pctExisting(true))
		if err != nil {
			__antithesis_instrumentation__.Notify(697043)
			treeTableName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{}, "")
			return &treeTableName, err
		} else {
			__antithesis_instrumentation__.Notify(697044)
		}
		__antithesis_instrumentation__.Notify(697042)

		treeTableName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
			SchemaName:     tree.Name(randSchema),
			ExplicitSchema: true,
		}, tree.Name(fmt.Sprintf("table%d", og.newUniqueSeqNum())))
		return &treeTableName, nil
	} else {
		__antithesis_instrumentation__.Notify(697045)
	}
	__antithesis_instrumentation__.Notify(697031)

	const q = `
  SELECT schema_name, table_name
    FROM [SHOW TABLES]
   WHERE table_name ~ 'table[0-9]+'
ORDER BY random()
   LIMIT 1;
`
	var schemaName string
	var tableName string
	if err := tx.QueryRow(ctx, q).Scan(&schemaName, &tableName); err != nil {
		__antithesis_instrumentation__.Notify(697046)
		treeTableName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{}, "")
		return &treeTableName, err
	} else {
		__antithesis_instrumentation__.Notify(697047)
	}
	__antithesis_instrumentation__.Notify(697032)

	treeTableName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
		SchemaName:     tree.Name(schemaName),
		ExplicitSchema: true,
	}, tree.Name(tableName))
	return &treeTableName, nil
}

func (og *operationGenerator) randView(
	ctx context.Context, tx pgx.Tx, pctExisting int, desiredSchema string,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(697048)
	if desiredSchema != "" {
		__antithesis_instrumentation__.Notify(697052)
		if og.randIntn(100) >= pctExisting {
			__antithesis_instrumentation__.Notify(697055)
			treeViewName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
				SchemaName:     tree.Name(desiredSchema),
				ExplicitSchema: true,
			}, tree.Name(fmt.Sprintf("view%d", og.newUniqueSeqNum())))
			return &treeViewName, nil
		} else {
			__antithesis_instrumentation__.Notify(697056)
		}
		__antithesis_instrumentation__.Notify(697053)

		q := fmt.Sprintf(`
		  SELECT table_name
		    FROM [SHOW TABLES]
		   WHERE table_name LIKE 'view%%'
				 AND schema_name = '%s'
		ORDER BY random()
		   LIMIT 1;
		`, desiredSchema)

		var viewName string
		if err := tx.QueryRow(ctx, q).Scan(&viewName); err != nil {
			__antithesis_instrumentation__.Notify(697057)
			treeViewName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{}, "")
			return &treeViewName, err
		} else {
			__antithesis_instrumentation__.Notify(697058)
		}
		__antithesis_instrumentation__.Notify(697054)
		treeViewName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
			SchemaName:     tree.Name(desiredSchema),
			ExplicitSchema: true,
		}, tree.Name(viewName))
		return &treeViewName, nil
	} else {
		__antithesis_instrumentation__.Notify(697059)
	}
	__antithesis_instrumentation__.Notify(697049)

	if og.randIntn(100) >= pctExisting {
		__antithesis_instrumentation__.Notify(697060)

		randSchema, err := og.randSchema(ctx, tx, og.pctExisting(true))
		if err != nil {
			__antithesis_instrumentation__.Notify(697062)
			treeViewName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{}, "")
			return &treeViewName, err
		} else {
			__antithesis_instrumentation__.Notify(697063)
		}
		__antithesis_instrumentation__.Notify(697061)
		treeViewName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
			SchemaName:     tree.Name(randSchema),
			ExplicitSchema: true,
		}, tree.Name(fmt.Sprintf("view%d", og.newUniqueSeqNum())))
		return &treeViewName, nil
	} else {
		__antithesis_instrumentation__.Notify(697064)
	}
	__antithesis_instrumentation__.Notify(697050)
	const q = `
  SELECT schema_name, table_name
    FROM [SHOW TABLES]
   WHERE table_name LIKE 'view%'
ORDER BY random()
   LIMIT 1;
`
	var schemaName string
	var viewName string
	if err := tx.QueryRow(ctx, q).Scan(&schemaName, &viewName); err != nil {
		__antithesis_instrumentation__.Notify(697065)
		treeViewName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{}, "")
		return &treeViewName, err
	} else {
		__antithesis_instrumentation__.Notify(697066)
	}
	__antithesis_instrumentation__.Notify(697051)
	treeViewName := tree.MakeTableNameFromPrefix(tree.ObjectNamePrefix{
		SchemaName:     tree.Name(schemaName),
		ExplicitSchema: true,
	}, tree.Name(viewName))
	return &treeViewName, nil
}

func (og *operationGenerator) tableColumnsShuffled(
	ctx context.Context, tx pgx.Tx, tableName string,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(697067)
	q := fmt.Sprintf(`
SELECT column_name
FROM [SHOW COLUMNS FROM %s];
`, tableName)

	rows, err := tx.Query(ctx, q)
	if err != nil {
		__antithesis_instrumentation__.Notify(697073)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(697074)
	}
	__antithesis_instrumentation__.Notify(697068)
	defer rows.Close()

	var columnNames []string
	for rows.Next() {
		__antithesis_instrumentation__.Notify(697075)
		var name string
		if err := rows.Scan(&name); err != nil {
			__antithesis_instrumentation__.Notify(697077)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(697078)
		}
		__antithesis_instrumentation__.Notify(697076)
		if name != "rowid" {
			__antithesis_instrumentation__.Notify(697079)
			columnNames = append(columnNames, name)
		} else {
			__antithesis_instrumentation__.Notify(697080)
		}
	}
	__antithesis_instrumentation__.Notify(697069)
	if rows.Err() != nil {
		__antithesis_instrumentation__.Notify(697081)
		return nil, rows.Err()
	} else {
		__antithesis_instrumentation__.Notify(697082)
	}
	__antithesis_instrumentation__.Notify(697070)

	og.params.rng.Shuffle(len(columnNames), func(i, j int) {
		__antithesis_instrumentation__.Notify(697083)
		columnNames[i], columnNames[j] = columnNames[j], columnNames[i]
	})
	__antithesis_instrumentation__.Notify(697071)

	if len(columnNames) <= 0 {
		__antithesis_instrumentation__.Notify(697084)
		return nil, errors.Errorf("table %s has no columns", tableName)
	} else {
		__antithesis_instrumentation__.Notify(697085)
	}
	__antithesis_instrumentation__.Notify(697072)
	return columnNames, nil
}

func (og *operationGenerator) randType(
	ctx context.Context, tx pgx.Tx, enumPctExisting int,
) (*tree.TypeName, *types.T, error) {
	__antithesis_instrumentation__.Notify(697086)
	if og.randIntn(100) <= og.params.enumPct {
		__antithesis_instrumentation__.Notify(697088)

		typName, exists, err := og.randEnum(ctx, tx, enumPctExisting)
		if err != nil {
			__antithesis_instrumentation__.Notify(697092)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(697093)
		}
		__antithesis_instrumentation__.Notify(697089)
		if !exists {
			__antithesis_instrumentation__.Notify(697094)
			return typName, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(697095)
		}
		__antithesis_instrumentation__.Notify(697090)
		typ, err := og.typeFromTypeName(ctx, tx, typName.String())
		if err != nil {
			__antithesis_instrumentation__.Notify(697096)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(697097)
		}
		__antithesis_instrumentation__.Notify(697091)
		return typName, typ, nil
	} else {
		__antithesis_instrumentation__.Notify(697098)
	}
	__antithesis_instrumentation__.Notify(697087)
	typ := randgen.RandSortingType(og.params.rng)
	typeName := tree.MakeUnqualifiedTypeName(typ.SQLString())
	return &typeName, typ, nil
}

func (og *operationGenerator) createSchema(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(697099)
	schemaName, err := og.randSchema(ctx, tx, og.pctExisting(false))
	if err != nil {
		__antithesis_instrumentation__.Notify(697103)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(697104)
	}
	__antithesis_instrumentation__.Notify(697100)
	ifNotExists := og.randIntn(2) == 0

	schemaExists, err := og.schemaExists(ctx, tx, schemaName)
	if err != nil {
		__antithesis_instrumentation__.Notify(697105)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(697106)
	}
	__antithesis_instrumentation__.Notify(697101)
	if schemaExists && func() bool {
		__antithesis_instrumentation__.Notify(697107)
		return !ifNotExists == true
	}() == true {
		__antithesis_instrumentation__.Notify(697108)
		og.expectedExecErrors.add(pgcode.DuplicateSchema)
	} else {
		__antithesis_instrumentation__.Notify(697109)
	}
	__antithesis_instrumentation__.Notify(697102)

	stmt := randgen.MakeSchemaName(ifNotExists, schemaName, tree.MakeRoleSpecWithRoleName(security.RootUserName().Normalized()))
	return tree.Serialize(stmt), nil
}

func (og *operationGenerator) randSchema(
	ctx context.Context, tx pgx.Tx, pctExisting int,
) (string, error) {
	__antithesis_instrumentation__.Notify(697110)
	if og.randIntn(100) >= pctExisting {
		__antithesis_instrumentation__.Notify(697113)
		return fmt.Sprintf("schema%d", og.newUniqueSeqNum()), nil
	} else {
		__antithesis_instrumentation__.Notify(697114)
	}
	__antithesis_instrumentation__.Notify(697111)
	const q = `
  SELECT schema_name
    FROM information_schema.schemata
   WHERE schema_name
    LIKE 'schema%'
      OR schema_name = 'public'
ORDER BY random()
   LIMIT 1;
`
	var name string
	if err := tx.QueryRow(ctx, q).Scan(&name); err != nil {
		__antithesis_instrumentation__.Notify(697115)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(697116)
	}
	__antithesis_instrumentation__.Notify(697112)
	return name, nil
}

func (og *operationGenerator) dropSchema(ctx context.Context, tx pgx.Tx) (string, error) {
	__antithesis_instrumentation__.Notify(697117)
	schemaName, err := og.randSchema(ctx, tx, og.pctExisting(true))
	if err != nil {
		__antithesis_instrumentation__.Notify(697121)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(697122)
	}
	__antithesis_instrumentation__.Notify(697118)

	schemaExists, err := og.schemaExists(ctx, tx, schemaName)
	if err != nil {
		__antithesis_instrumentation__.Notify(697123)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(697124)
	}
	__antithesis_instrumentation__.Notify(697119)
	crossReferences, err := og.schemaContainsTypesWithCrossSchemaReferences(ctx, tx, schemaName)
	if err != nil {
		__antithesis_instrumentation__.Notify(697125)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(697126)
	}
	__antithesis_instrumentation__.Notify(697120)
	codesWithConditions{
		{pgcode.UndefinedSchema, !schemaExists},
		{pgcode.InvalidSchemaName, schemaName == tree.PublicSchema},
		{pgcode.FeatureNotSupported, crossReferences},
	}.add(og.expectedExecErrors)

	return fmt.Sprintf(`DROP SCHEMA "%s" CASCADE`, schemaName), nil
}

func (og *operationGenerator) pctExisting(shouldAlreadyExist bool) int {
	__antithesis_instrumentation__.Notify(697127)
	if shouldAlreadyExist {
		__antithesis_instrumentation__.Notify(697129)
		return 100 - og.params.errorRate
	} else {
		__antithesis_instrumentation__.Notify(697130)
	}
	__antithesis_instrumentation__.Notify(697128)
	return og.params.errorRate
}

func (og operationGenerator) alwaysExisting() int {
	__antithesis_instrumentation__.Notify(697131)
	return 100
}

func (og *operationGenerator) produceError() bool {
	__antithesis_instrumentation__.Notify(697132)
	return og.randIntn(100) < og.params.errorRate
}

func (og *operationGenerator) randIntn(topBound int) int {
	__antithesis_instrumentation__.Notify(697133)
	return og.params.rng.Intn(topBound)
}

func (og *operationGenerator) newUniqueSeqNum() int64 {
	__antithesis_instrumentation__.Notify(697134)
	return atomic.AddInt64(og.params.seqNum, 1)
}

func (og *operationGenerator) typeFromTypeName(
	ctx context.Context, tx pgx.Tx, typeName string,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(697135)
	stmt, err := parser.ParseOne(fmt.Sprintf("SELECT 'placeholder'::%s", typeName))
	if err != nil {
		__antithesis_instrumentation__.Notify(697138)
		return nil, errors.Wrapf(err, "typeFromTypeName: %s", typeName)
	} else {
		__antithesis_instrumentation__.Notify(697139)
	}
	__antithesis_instrumentation__.Notify(697136)
	typ, err := tree.ResolveType(
		context.Background(),
		stmt.AST.(*tree.Select).Select.(*tree.SelectClause).Exprs[0].Expr.(*tree.CastExpr).Type,
		&txTypeResolver{tx: tx},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(697140)
		return nil, errors.Wrapf(err, "ResolveType: %v", typeName)
	} else {
		__antithesis_instrumentation__.Notify(697141)
	}
	__antithesis_instrumentation__.Notify(697137)
	return typ, nil
}

func isClusterVersionLessThan(
	ctx context.Context, tx pgx.Tx, targetVersion roachpb.Version,
) (bool, error) {
	__antithesis_instrumentation__.Notify(697142)
	var clusterVersionStr string
	row := tx.QueryRow(ctx, `SHOW CLUSTER SETTING version`)
	if err := row.Scan(&clusterVersionStr); err != nil {
		__antithesis_instrumentation__.Notify(697145)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(697146)
	}
	__antithesis_instrumentation__.Notify(697143)
	clusterVersion, err := roachpb.ParseVersion(clusterVersionStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(697147)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(697148)
	}
	__antithesis_instrumentation__.Notify(697144)
	return clusterVersion.LessEq(targetVersion), nil
}
