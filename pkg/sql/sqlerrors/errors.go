// Package sqlerrors exports errors which can occur in the sql package.
package sqlerrors

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

const (
	txnAbortedMsg = "current transaction is aborted, commands ignored " +
		"until end of transaction block"
	txnCommittedMsg = "current transaction is committed, commands ignored " +
		"until end of transaction block"
)

func NewTransactionAbortedError(customMsg string) error {
	__antithesis_instrumentation__.Notify(623864)
	if customMsg != "" {
		__antithesis_instrumentation__.Notify(623866)
		return pgerror.Newf(
			pgcode.InFailedSQLTransaction, "%s: %s", customMsg, txnAbortedMsg)
	} else {
		__antithesis_instrumentation__.Notify(623867)
	}
	__antithesis_instrumentation__.Notify(623865)
	return pgerror.New(pgcode.InFailedSQLTransaction, txnAbortedMsg)
}

func NewTransactionCommittedError() error {
	__antithesis_instrumentation__.Notify(623868)
	return pgerror.New(pgcode.InvalidTransactionState, txnCommittedMsg)
}

func NewNonNullViolationError(columnName string) error {
	__antithesis_instrumentation__.Notify(623869)
	return pgerror.Newf(pgcode.NotNullViolation, "null value in column %q violates not-null constraint", columnName)
}

func NewInvalidAssignmentCastError(
	sourceType *types.T, targetType *types.T, targetColName string,
) error {
	__antithesis_instrumentation__.Notify(623870)
	return errors.WithHint(
		pgerror.Newf(
			pgcode.DatatypeMismatch,
			"value type %s doesn't match type %s of column %q",
			sourceType, targetType, tree.ErrNameString(targetColName),
		),
		"you will need to rewrite or cast the expression",
	)
}

func NewGeneratedAlwaysAsIdentityColumnOverrideError(columnName string) error {
	__antithesis_instrumentation__.Notify(623871)
	return errors.WithDetailf(
		pgerror.Newf(pgcode.GeneratedAlways, "cannot insert into column %q", columnName),
		"Column %q is an identity column defined as GENERATED ALWAYS", columnName,
	)
}

func NewGeneratedAlwaysAsIdentityColumnUpdateError(columnName string) error {
	__antithesis_instrumentation__.Notify(623872)
	return errors.WithDetailf(
		pgerror.Newf(pgcode.GeneratedAlways, "column %q can only be updated to DEFAULT", columnName),
		"Column %q is an identity column defined as GENERATED ALWAYS", columnName,
	)
}

func NewIdentityColumnTypeError() error {
	__antithesis_instrumentation__.Notify(623873)
	return pgerror.Newf(pgcode.InvalidParameterValue,
		"identity column type must be INT, INT2, INT4 or INT8")
}

func NewInvalidSchemaDefinitionError(err error) error {
	__antithesis_instrumentation__.Notify(623874)
	return pgerror.WithCandidateCode(err, pgcode.InvalidSchemaDefinition)
}

func NewUndefinedSchemaError(name string) error {
	__antithesis_instrumentation__.Notify(623875)
	return pgerror.Newf(pgcode.InvalidSchemaName, "unknown schema %q", name)
}

func NewCCLRequiredError(err error) error {
	__antithesis_instrumentation__.Notify(623876)
	return pgerror.WithCandidateCode(err, pgcode.CCLRequired)
}

func IsCCLRequiredError(err error) bool {
	__antithesis_instrumentation__.Notify(623877)
	return errHasCode(err, pgcode.CCLRequired)
}

func NewUndefinedDatabaseError(name string) error {
	__antithesis_instrumentation__.Notify(623878)

	return pgerror.Newf(
		pgcode.InvalidCatalogName, "database %q does not exist", name)
}

func NewInvalidWildcardError(name string) error {
	__antithesis_instrumentation__.Notify(623879)
	return pgerror.Newf(
		pgcode.InvalidCatalogName,
		"%q does not match any valid database or schema", name)
}

func NewUndefinedObjectError(name tree.NodeFormatter, kind tree.DesiredObjectKind) error {
	__antithesis_instrumentation__.Notify(623880)
	switch kind {
	case tree.TableObject:
		__antithesis_instrumentation__.Notify(623881)
		return NewUndefinedRelationError(name)
	case tree.TypeObject:
		__antithesis_instrumentation__.Notify(623882)
		return NewUndefinedTypeError(name)
	default:
		__antithesis_instrumentation__.Notify(623883)
		return errors.AssertionFailedf("unknown object kind %d", kind)
	}
}

func NewUndefinedTypeError(name tree.NodeFormatter) error {
	__antithesis_instrumentation__.Notify(623884)
	return pgerror.Newf(pgcode.UndefinedObject, "type %q does not exist", tree.ErrString(name))
}

func NewUndefinedRelationError(name tree.NodeFormatter) error {
	__antithesis_instrumentation__.Notify(623885)
	return pgerror.Newf(pgcode.UndefinedTable,
		"relation %q does not exist", tree.ErrString(name))
}

func NewColumnAlreadyExistsError(name, relation string) error {
	__antithesis_instrumentation__.Notify(623886)
	return pgerror.Newf(pgcode.DuplicateColumn, "column %q of relation %q already exists", name, relation)
}

func NewDatabaseAlreadyExistsError(name string) error {
	__antithesis_instrumentation__.Notify(623887)
	return pgerror.Newf(pgcode.DuplicateDatabase, "database %q already exists", name)
}

func NewSchemaAlreadyExistsError(name string) error {
	__antithesis_instrumentation__.Notify(623888)
	return pgerror.Newf(pgcode.DuplicateSchema, "schema %q already exists", name)
}

func WrapErrorWhileConstructingObjectAlreadyExistsErr(err error) error {
	__antithesis_instrumentation__.Notify(623889)
	return pgerror.WithCandidateCode(errors.Wrap(err, "object already exists"), pgcode.DuplicateObject)
}

func MakeObjectAlreadyExistsError(collidingObject *descpb.Descriptor, name string) error {
	__antithesis_instrumentation__.Notify(623890)
	switch collidingObject.Union.(type) {
	case *descpb.Descriptor_Table:
		__antithesis_instrumentation__.Notify(623891)
		return NewRelationAlreadyExistsError(name)
	case *descpb.Descriptor_Type:
		__antithesis_instrumentation__.Notify(623892)
		return NewTypeAlreadyExistsError(name)
	case *descpb.Descriptor_Database:
		__antithesis_instrumentation__.Notify(623893)
		return NewDatabaseAlreadyExistsError(name)
	case *descpb.Descriptor_Schema:
		__antithesis_instrumentation__.Notify(623894)
		return NewSchemaAlreadyExistsError(name)
	default:
		__antithesis_instrumentation__.Notify(623895)
		return errors.AssertionFailedf("unknown type %T exists with name %v", collidingObject.Union, name)
	}
}

func NewRelationAlreadyExistsError(name string) error {
	__antithesis_instrumentation__.Notify(623896)
	return pgerror.Newf(pgcode.DuplicateRelation, "relation %q already exists", name)
}

func NewTypeAlreadyExistsError(name string) error {
	__antithesis_instrumentation__.Notify(623897)
	return pgerror.Newf(pgcode.DuplicateObject, "type %q already exists", name)
}

func IsRelationAlreadyExistsError(err error) bool {
	__antithesis_instrumentation__.Notify(623898)
	return errHasCode(err, pgcode.DuplicateRelation)
}

func NewWrongObjectTypeError(name tree.NodeFormatter, desiredObjType string) error {
	__antithesis_instrumentation__.Notify(623899)
	return pgerror.Newf(pgcode.WrongObjectType, "%q is not a %s",
		tree.ErrString(name), desiredObjType)
}

func NewSyntaxErrorf(format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(623900)
	return pgerror.Newf(pgcode.Syntax, format, args...)
}

func NewDependentObjectErrorf(format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(623901)
	return pgerror.Newf(pgcode.DependentObjectsStillExist, format, args...)
}

func NewRangeUnavailableError(rangeID roachpb.RangeID, origErr error) error {
	__antithesis_instrumentation__.Notify(623902)
	return pgerror.Wrapf(origErr, pgcode.RangeUnavailable, "key range id:%d is unavailable", rangeID)
}

func NewWindowInAggError() error {
	__antithesis_instrumentation__.Notify(623903)
	return pgerror.New(pgcode.Grouping,
		"window functions are not allowed in aggregate")
}

func NewAggInAggError() error {
	__antithesis_instrumentation__.Notify(623904)
	return pgerror.New(pgcode.Grouping, "aggregate function calls cannot be nested")
}

var QueryTimeoutError = pgerror.New(
	pgcode.QueryCanceled, "query execution canceled due to statement timeout")

func IsOutOfMemoryError(err error) bool {
	__antithesis_instrumentation__.Notify(623905)
	return errHasCode(err, pgcode.OutOfMemory)
}

func IsDiskFullError(err error) bool {
	__antithesis_instrumentation__.Notify(623906)
	return errHasCode(err, pgcode.DiskFull)
}

func IsUndefinedColumnError(err error) bool {
	__antithesis_instrumentation__.Notify(623907)
	return errHasCode(err, pgcode.UndefinedColumn)
}

func IsUndefinedRelationError(err error) bool {
	__antithesis_instrumentation__.Notify(623908)
	return errHasCode(err, pgcode.UndefinedTable)
}

func IsUndefinedDatabaseError(err error) bool {
	__antithesis_instrumentation__.Notify(623909)
	return errHasCode(err, pgcode.UndefinedDatabase)
}

func IsUndefinedSchemaError(err error) bool {
	__antithesis_instrumentation__.Notify(623910)
	return errHasCode(err, pgcode.UndefinedSchema)
}

func errHasCode(err error, code ...pgcode.Code) bool {
	__antithesis_instrumentation__.Notify(623911)
	pgCode := pgerror.GetPGCode(err)
	for _, c := range code {
		__antithesis_instrumentation__.Notify(623913)
		if pgCode == c {
			__antithesis_instrumentation__.Notify(623914)
			return true
		} else {
			__antithesis_instrumentation__.Notify(623915)
		}
	}
	__antithesis_instrumentation__.Notify(623912)
	return false
}
