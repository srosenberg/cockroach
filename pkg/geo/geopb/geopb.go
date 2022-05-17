package geopb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"unsafe"
)

func (b *SpatialObject) EWKBHex() string {
	__antithesis_instrumentation__.Notify(63664)
	return fmt.Sprintf("%X", b.EWKB)
}

func (b *SpatialObject) MemSize() uintptr {
	__antithesis_instrumentation__.Notify(63665)
	var bboxSize uintptr
	if bbox := b.BoundingBox; bbox != nil {
		__antithesis_instrumentation__.Notify(63667)
		bboxSize = unsafe.Sizeof(*bbox)
	} else {
		__antithesis_instrumentation__.Notify(63668)
	}
	__antithesis_instrumentation__.Notify(63666)
	return unsafe.Sizeof(*b) + bboxSize + uintptr(len(b.EWKB))
}

func (s ShapeType) MultiType() ShapeType {
	__antithesis_instrumentation__.Notify(63669)
	switch s {
	case ShapeType_Unset:
		__antithesis_instrumentation__.Notify(63670)
		return ShapeType_Unset
	case ShapeType_Point, ShapeType_MultiPoint:
		__antithesis_instrumentation__.Notify(63671)
		return ShapeType_MultiPoint
	case ShapeType_LineString, ShapeType_MultiLineString:
		__antithesis_instrumentation__.Notify(63672)
		return ShapeType_MultiLineString
	case ShapeType_Polygon, ShapeType_MultiPolygon:
		__antithesis_instrumentation__.Notify(63673)
		return ShapeType_MultiPolygon
	case ShapeType_GeometryCollection:
		__antithesis_instrumentation__.Notify(63674)
		return ShapeType_GeometryCollection
	default:
		__antithesis_instrumentation__.Notify(63675)
		return ShapeType_Unset
	}
}
