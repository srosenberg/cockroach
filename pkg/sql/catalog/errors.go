package catalog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func ValidateName(name, typ string) error {
	__antithesis_instrumentation__.Notify(265412)
	if len(name) == 0 {
		__antithesis_instrumentation__.Notify(265414)
		return pgerror.Newf(pgcode.Syntax, "empty %s name", typ)
	} else {
		__antithesis_instrumentation__.Notify(265415)
	}
	__antithesis_instrumentation__.Notify(265413)

	return nil
}

type inactiveDescriptorError struct {
	cause error
}

type addingTableError struct {
	cause error
}

func newAddingTableError(desc TableDescriptor) error {
	__antithesis_instrumentation__.Notify(265416)
	typStr := "table"
	if desc.IsView() && func() bool {
		__antithesis_instrumentation__.Notify(265418)
		return desc.IsPhysicalTable() == true
	}() == true {
		__antithesis_instrumentation__.Notify(265419)
		typStr = "materialized view"
	} else {
		__antithesis_instrumentation__.Notify(265420)
	}
	__antithesis_instrumentation__.Notify(265417)
	return &addingTableError{
		cause: errors.Errorf("%s %q is being added", typStr, desc.GetName()),
	}
}

func (a *addingTableError) Error() string {
	__antithesis_instrumentation__.Notify(265421)
	return a.cause.Error()
}

func (a *addingTableError) Unwrap() error {
	__antithesis_instrumentation__.Notify(265422)
	return a.cause
}

var ErrDescriptorDropped = errors.New("descriptor is being dropped")

func (i *inactiveDescriptorError) Error() string {
	__antithesis_instrumentation__.Notify(265423)
	return i.cause.Error()
}

func (i *inactiveDescriptorError) Unwrap() error {
	__antithesis_instrumentation__.Notify(265424)
	return i.cause
}

func HasAddingTableError(err error) bool {
	__antithesis_instrumentation__.Notify(265425)
	return errors.HasType(err, (*addingTableError)(nil))
}

func NewInactiveDescriptorError(err error) error {
	__antithesis_instrumentation__.Notify(265426)
	return &inactiveDescriptorError{err}
}

func HasInactiveDescriptorError(err error) bool {
	__antithesis_instrumentation__.Notify(265427)
	return errors.HasType(err, (*inactiveDescriptorError)(nil))
}

var ErrDescriptorNotFound = errors.New("descriptor not found")

var ErrReferencedDescriptorNotFound = errors.New("referenced descriptor not found")

var ErrDescriptorWrongType = errors.New("unexpected descriptor type")

func NewDescriptorTypeError(desc interface{}) error {
	__antithesis_instrumentation__.Notify(265428)
	return errors.Wrapf(ErrDescriptorWrongType, "descriptor is a %T", desc)
}

func AsDatabaseDescriptor(desc Descriptor) (DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(265429)
	db, ok := desc.(DatabaseDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(265431)
		if desc == nil {
			__antithesis_instrumentation__.Notify(265433)
			return nil, NewDescriptorTypeError(desc)
		} else {
			__antithesis_instrumentation__.Notify(265434)
		}
		__antithesis_instrumentation__.Notify(265432)
		return nil, WrapDatabaseDescRefErr(desc.GetID(), NewDescriptorTypeError(desc))
	} else {
		__antithesis_instrumentation__.Notify(265435)
	}
	__antithesis_instrumentation__.Notify(265430)
	return db, nil
}

func AsSchemaDescriptor(desc Descriptor) (SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(265436)
	schema, ok := desc.(SchemaDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(265438)
		if desc == nil {
			__antithesis_instrumentation__.Notify(265440)
			return nil, NewDescriptorTypeError(desc)
		} else {
			__antithesis_instrumentation__.Notify(265441)
		}
		__antithesis_instrumentation__.Notify(265439)
		return nil, WrapSchemaDescRefErr(desc.GetID(), NewDescriptorTypeError(desc))
	} else {
		__antithesis_instrumentation__.Notify(265442)
	}
	__antithesis_instrumentation__.Notify(265437)
	return schema, nil
}

func AsTableDescriptor(desc Descriptor) (TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(265443)
	table, ok := desc.(TableDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(265445)
		if desc == nil {
			__antithesis_instrumentation__.Notify(265447)
			return nil, NewDescriptorTypeError(desc)
		} else {
			__antithesis_instrumentation__.Notify(265448)
		}
		__antithesis_instrumentation__.Notify(265446)
		return nil, WrapTableDescRefErr(desc.GetID(), NewDescriptorTypeError(desc))
	} else {
		__antithesis_instrumentation__.Notify(265449)
	}
	__antithesis_instrumentation__.Notify(265444)
	return table, nil
}

func AsTypeDescriptor(desc Descriptor) (TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(265450)
	typ, ok := desc.(TypeDescriptor)
	if !ok {
		__antithesis_instrumentation__.Notify(265452)
		if desc == nil {
			__antithesis_instrumentation__.Notify(265454)
			return nil, NewDescriptorTypeError(desc)
		} else {
			__antithesis_instrumentation__.Notify(265455)
		}
		__antithesis_instrumentation__.Notify(265453)
		return nil, WrapTypeDescRefErr(desc.GetID(), NewDescriptorTypeError(desc))
	} else {
		__antithesis_instrumentation__.Notify(265456)
	}
	__antithesis_instrumentation__.Notify(265451)
	return typ, nil
}

func WrapDatabaseDescRefErr(id descpb.ID, err error) error {
	__antithesis_instrumentation__.Notify(265457)
	return errors.Wrapf(err, "referenced database ID %d", errors.Safe(id))
}

func WrapSchemaDescRefErr(id descpb.ID, err error) error {
	__antithesis_instrumentation__.Notify(265458)
	return errors.Wrapf(err, "referenced schema ID %d", errors.Safe(id))
}

func WrapTableDescRefErr(id descpb.ID, err error) error {
	__antithesis_instrumentation__.Notify(265459)
	return errors.Wrapf(err, "referenced table ID %d", errors.Safe(id))
}

func WrapTypeDescRefErr(id descpb.ID, err error) error {
	__antithesis_instrumentation__.Notify(265460)
	return errors.Wrapf(err, "referenced type ID %d", errors.Safe(id))
}

func NewMutableAccessToVirtualSchemaError(entry VirtualSchema, object string) error {
	__antithesis_instrumentation__.Notify(265461)
	switch entry.Desc().GetName() {
	case "pg_catalog":
		__antithesis_instrumentation__.Notify(265462)
		return pgerror.Newf(pgcode.InsufficientPrivilege,
			"%s is a system catalog", tree.ErrNameString(object))
	default:
		__antithesis_instrumentation__.Notify(265463)
		return pgerror.Newf(pgcode.WrongObjectType,
			"%s is a virtual object and cannot be modified", tree.ErrNameString(object))
	}
}
