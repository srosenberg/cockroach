package geopb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

const (
	UnknownSRID = SRID(0)

	DefaultGeometrySRID = UnknownSRID

	DefaultGeographySRID = SRID(4326)
)

const (
	ZShapeTypeFlag = 1 << 30

	MShapeTypeFlag = 1 << 29
)

func (s ShapeType) To2D() ShapeType {
	__antithesis_instrumentation__.Notify(64027)
	return ShapeType(uint32(s) & (MShapeTypeFlag - 1))
}

type SRID int32

type WKT string

type EWKT string

type WKB []byte

type EWKB []byte
