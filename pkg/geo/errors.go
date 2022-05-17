package geo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
)

func NewMismatchingSRIDsError(a geopb.SpatialObject, b geopb.SpatialObject) error {
	__antithesis_instrumentation__.Notify(58797)
	return pgerror.Newf(
		pgcode.InvalidParameterValue,
		"operation on mixed SRIDs forbidden: (%s, %d) != (%s, %d)",
		a.ShapeType,
		a.SRID,
		b.ShapeType,
		b.SRID,
	)
}

type EmptyGeometryError struct {
	cause error
}

var _ error = (*EmptyGeometryError)(nil)
var _ errors.SafeDetailer = (*EmptyGeometryError)(nil)
var _ fmt.Formatter = (*EmptyGeometryError)(nil)
var _ errors.Formatter = (*EmptyGeometryError)(nil)

func (w *EmptyGeometryError) Error() string {
	__antithesis_instrumentation__.Notify(58798)
	return w.cause.Error()
}

func (w *EmptyGeometryError) Cause() error {
	__antithesis_instrumentation__.Notify(58799)
	return w.cause
}

func (w *EmptyGeometryError) Unwrap() error {
	__antithesis_instrumentation__.Notify(58800)
	return w.cause
}

func (w *EmptyGeometryError) SafeDetails() []string {
	__antithesis_instrumentation__.Notify(58801)
	return []string{w.cause.Error()}
}

func (w *EmptyGeometryError) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(58802)
	errors.FormatError(w, s, verb)
}

func (w *EmptyGeometryError) FormatError(p errors.Printer) (next error) {
	__antithesis_instrumentation__.Notify(58803)
	return w.cause
}

func IsEmptyGeometryError(err error) bool {
	__antithesis_instrumentation__.Notify(58804)
	return errors.HasType(err, &EmptyGeometryError{})
}

func NewEmptyGeometryError() *EmptyGeometryError {
	__antithesis_instrumentation__.Notify(58805)
	return &EmptyGeometryError{cause: pgerror.Newf(pgcode.InvalidParameterValue, "empty shape found")}
}
