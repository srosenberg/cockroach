// Package twkb implements the Tiny Well-known Binary (TWKB) encoding
// as described in https://github.com/TWKB/Specification/blob/master/twkb.md.
package twkb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"math"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/twpayne/go-geom"
)

type twkbType uint8

const (
	twkbTypePoint              twkbType = 1
	twkbTypeLineString         twkbType = 2
	twkbTypePolygon            twkbType = 3
	twkbTypeMultiPoint         twkbType = 4
	twkbTypeMultiLineString    twkbType = 5
	twkbTypeMultiPolygon       twkbType = 6
	twkbTypeGeometryCollection twkbType = 7
)

type marshalOptions struct {
	precisionXY int8
	precisionZ  int8
	precisionM  int8
}

type marshaller struct {
	bytes.Buffer
	o marshalOptions

	prevCoords []int64
}

type MarshalOption func(o *marshalOptions) error

func MarshalOptionPrecisionXY(p int64) MarshalOption {
	__antithesis_instrumentation__.Notify(64935)
	return func(o *marshalOptions) error {
		__antithesis_instrumentation__.Notify(64936)
		if p < -7 || func() bool {
			__antithesis_instrumentation__.Notify(64938)
			return p > 7 == true
		}() == true {
			__antithesis_instrumentation__.Notify(64939)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"XY precision must be between -7 and 7 inclusive",
			)
		} else {
			__antithesis_instrumentation__.Notify(64940)
		}
		__antithesis_instrumentation__.Notify(64937)
		o.precisionXY = int8(p)
		return nil
	}
}

func MarshalOptionPrecisionZ(p int64) MarshalOption {
	__antithesis_instrumentation__.Notify(64941)
	return func(o *marshalOptions) error {
		__antithesis_instrumentation__.Notify(64942)
		if p < 0 || func() bool {
			__antithesis_instrumentation__.Notify(64944)
			return p > 7 == true
		}() == true {
			__antithesis_instrumentation__.Notify(64945)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"Z precision must not be negative or greater than 7",
			)
		} else {
			__antithesis_instrumentation__.Notify(64946)
		}
		__antithesis_instrumentation__.Notify(64943)

		o.precisionZ = int8(p)
		return nil
	}
}

func MarshalOptionPrecisionM(p int64) MarshalOption {
	__antithesis_instrumentation__.Notify(64947)
	return func(o *marshalOptions) error {
		__antithesis_instrumentation__.Notify(64948)
		if p < 0 || func() bool {
			__antithesis_instrumentation__.Notify(64950)
			return p > 7 == true
		}() == true {
			__antithesis_instrumentation__.Notify(64951)
			return pgerror.Newf(
				pgcode.InvalidParameterValue,
				"M precision must not be negative or greater than 7",
			)
		} else {
			__antithesis_instrumentation__.Notify(64952)
		}
		__antithesis_instrumentation__.Notify(64949)

		o.precisionM = int8(p)
		return nil
	}
}

func Marshal(t geom.T, opts ...MarshalOption) ([]byte, error) {
	__antithesis_instrumentation__.Notify(64953)
	var o marshalOptions
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(64956)
		if err := opt(&o); err != nil {
			__antithesis_instrumentation__.Notify(64957)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(64958)
		}
	}
	__antithesis_instrumentation__.Notify(64954)

	m := marshaller{
		o:          o,
		prevCoords: make([]int64, t.Layout().Stride()),
	}
	if err := (&m).marshal(t); err != nil {
		__antithesis_instrumentation__.Notify(64959)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(64960)
	}
	__antithesis_instrumentation__.Notify(64955)
	return m.Bytes(), nil
}

func (m *marshaller) marshal(t geom.T) error {
	__antithesis_instrumentation__.Notify(64961)

	typeAndPrecisionHeader, err := typeAndPrecisionHeaderByte(t, m.o)
	if err != nil {
		__antithesis_instrumentation__.Notify(64967)
		return err
	} else {
		__antithesis_instrumentation__.Notify(64968)
	}
	__antithesis_instrumentation__.Notify(64962)
	if err := m.WriteByte(typeAndPrecisionHeader); err != nil {
		__antithesis_instrumentation__.Notify(64969)
		return err
	} else {
		__antithesis_instrumentation__.Notify(64970)
	}
	__antithesis_instrumentation__.Notify(64963)
	if err := m.WriteByte(metadataByte(t, m.o)); err != nil {
		__antithesis_instrumentation__.Notify(64971)
		return err
	} else {
		__antithesis_instrumentation__.Notify(64972)
	}
	__antithesis_instrumentation__.Notify(64964)
	if t.Layout().Stride() > 2 {
		__antithesis_instrumentation__.Notify(64973)
		if err := m.WriteByte(extendedDimensionsByte(t, m.o)); err != nil {
			__antithesis_instrumentation__.Notify(64974)
			return err
		} else {
			__antithesis_instrumentation__.Notify(64975)
		}
	} else {
		__antithesis_instrumentation__.Notify(64976)
	}
	__antithesis_instrumentation__.Notify(64965)
	if !t.Empty() {
		__antithesis_instrumentation__.Notify(64977)

		switch t := t.(type) {
		case *geom.Point:
			__antithesis_instrumentation__.Notify(64978)
			if err := m.writeFlatCoords(
				t.FlatCoords(),
				t.Layout(),
				false,
			); err != nil {
				__antithesis_instrumentation__.Notify(64988)
				return err
			} else {
				__antithesis_instrumentation__.Notify(64989)
			}
		case *geom.LineString:
			__antithesis_instrumentation__.Notify(64979)
			if err := m.writeFlatCoords(
				t.FlatCoords(),
				t.Layout(),
				true,
			); err != nil {
				__antithesis_instrumentation__.Notify(64990)
				return err
			} else {
				__antithesis_instrumentation__.Notify(64991)
			}
		case *geom.Polygon:
			__antithesis_instrumentation__.Notify(64980)
			if err := m.writeGeomWithEnds(
				t.FlatCoords(),
				t.Ends(),
				t.Layout(),
				0,
			); err != nil {
				__antithesis_instrumentation__.Notify(64992)
				return err
			} else {
				__antithesis_instrumentation__.Notify(64993)
			}
		case *geom.MultiPoint:
			__antithesis_instrumentation__.Notify(64981)

			if err := m.writeFlatCoords(
				t.FlatCoords(),
				t.Layout(),
				true,
			); err != nil {
				__antithesis_instrumentation__.Notify(64994)
				return err
			} else {
				__antithesis_instrumentation__.Notify(64995)
			}
		case *geom.MultiLineString:
			__antithesis_instrumentation__.Notify(64982)

			if err := m.writeGeomWithEnds(
				t.FlatCoords(),
				t.Ends(),
				t.Layout(),
				0,
			); err != nil {
				__antithesis_instrumentation__.Notify(64996)
				return err
			} else {
				__antithesis_instrumentation__.Notify(64997)
			}
		case *geom.MultiPolygon:
			__antithesis_instrumentation__.Notify(64983)

			flatCoords := t.FlatCoords()
			endss := t.Endss()

			if err := m.putUvarint(uint64(len(endss))); err != nil {
				__antithesis_instrumentation__.Notify(64998)
				return err
			} else {
				__antithesis_instrumentation__.Notify(64999)
			}
			__antithesis_instrumentation__.Notify(64984)
			startOffset := 0
			for _, ends := range endss {
				__antithesis_instrumentation__.Notify(65000)

				if err := m.writeGeomWithEnds(
					flatCoords,
					ends,
					t.Layout(),
					startOffset,
				); err != nil {
					__antithesis_instrumentation__.Notify(65002)
					return err
				} else {
					__antithesis_instrumentation__.Notify(65003)
				}
				__antithesis_instrumentation__.Notify(65001)

				if len(ends) > 0 {
					__antithesis_instrumentation__.Notify(65004)
					startOffset = ends[len(ends)-1]
				} else {
					__antithesis_instrumentation__.Notify(65005)
				}
			}
		case *geom.GeometryCollection:
			__antithesis_instrumentation__.Notify(64985)

			if err := m.putUvarint(uint64(t.NumGeoms())); err != nil {
				__antithesis_instrumentation__.Notify(65006)
				return err
			} else {
				__antithesis_instrumentation__.Notify(65007)
			}
			__antithesis_instrumentation__.Notify(64986)
			for _, gcT := range t.Geoms() {
				__antithesis_instrumentation__.Notify(65008)

				for i := range m.prevCoords {
					__antithesis_instrumentation__.Notify(65010)
					m.prevCoords[i] = 0
				}
				__antithesis_instrumentation__.Notify(65009)
				if err := m.marshal(gcT); err != nil {
					__antithesis_instrumentation__.Notify(65011)
					return err
				} else {
					__antithesis_instrumentation__.Notify(65012)
				}
			}
		default:
			__antithesis_instrumentation__.Notify(64987)
			return pgerror.Newf(pgcode.InvalidParameterValue, "unknown TWKB type: %T", t)
		}
	} else {
		__antithesis_instrumentation__.Notify(65013)
	}
	__antithesis_instrumentation__.Notify(64966)
	return nil
}

func (m *marshaller) writeGeomWithEnds(
	flatCoords []float64, ends []int, layout geom.Layout, startOffset int,
) error {
	__antithesis_instrumentation__.Notify(65014)

	if err := m.putUvarint(uint64(len(ends))); err != nil {
		__antithesis_instrumentation__.Notify(65017)
		return err
	} else {
		__antithesis_instrumentation__.Notify(65018)
	}
	__antithesis_instrumentation__.Notify(65015)
	start := startOffset

	for _, end := range ends {
		__antithesis_instrumentation__.Notify(65019)
		if err := m.writeFlatCoords(
			flatCoords[start:end],
			layout,
			true,
		); err != nil {
			__antithesis_instrumentation__.Notify(65021)
			return err
		} else {
			__antithesis_instrumentation__.Notify(65022)
		}
		__antithesis_instrumentation__.Notify(65020)
		start = end
	}
	__antithesis_instrumentation__.Notify(65016)
	return nil
}

func (m *marshaller) writeFlatCoords(
	flatCoords []float64, layout geom.Layout, writeLen bool,
) error {
	__antithesis_instrumentation__.Notify(65023)
	stride := layout.Stride()
	if writeLen {
		__antithesis_instrumentation__.Notify(65026)
		if err := m.putUvarint(uint64(len(flatCoords) / stride)); err != nil {
			__antithesis_instrumentation__.Notify(65027)
			return err
		} else {
			__antithesis_instrumentation__.Notify(65028)
		}
	} else {
		__antithesis_instrumentation__.Notify(65029)
	}
	__antithesis_instrumentation__.Notify(65024)

	zIndex := layout.ZIndex()
	mIndex := layout.MIndex()
	for i, coord := range flatCoords {
		__antithesis_instrumentation__.Notify(65030)

		precision := m.o.precisionXY
		if i%stride == zIndex {
			__antithesis_instrumentation__.Notify(65034)
			precision = m.o.precisionZ
		} else {
			__antithesis_instrumentation__.Notify(65035)
		}
		__antithesis_instrumentation__.Notify(65031)
		if i%stride == mIndex {
			__antithesis_instrumentation__.Notify(65036)
			precision = m.o.precisionM
		} else {
			__antithesis_instrumentation__.Notify(65037)
		}
		__antithesis_instrumentation__.Notify(65032)
		curr := int64(math.Round(coord * math.Pow(10, float64(precision))))
		prev := m.prevCoords[i%stride]

		if err := m.putVarint(curr - prev); err != nil {
			__antithesis_instrumentation__.Notify(65038)
			return err
		} else {
			__antithesis_instrumentation__.Notify(65039)
		}
		__antithesis_instrumentation__.Notify(65033)
		m.prevCoords[i%stride] = curr
	}
	__antithesis_instrumentation__.Notify(65025)
	return nil
}

func (m *marshaller) putVarint(x int64) error {
	__antithesis_instrumentation__.Notify(65040)
	ux := uint64(x) << 1
	if x < 0 {
		__antithesis_instrumentation__.Notify(65042)
		ux = ^ux
	} else {
		__antithesis_instrumentation__.Notify(65043)
	}
	__antithesis_instrumentation__.Notify(65041)
	return m.putUvarint(ux)
}

func (m *marshaller) putUvarint(x uint64) error {
	__antithesis_instrumentation__.Notify(65044)
	for x >= 0x80 {
		__antithesis_instrumentation__.Notify(65046)
		if err := m.WriteByte(byte(x) | 0x80); err != nil {
			__antithesis_instrumentation__.Notify(65048)
			return err
		} else {
			__antithesis_instrumentation__.Notify(65049)
		}
		__antithesis_instrumentation__.Notify(65047)
		x >>= 7
	}
	__antithesis_instrumentation__.Notify(65045)
	return m.WriteByte(byte(x))
}

func typeAndPrecisionHeaderByte(t geom.T, o marshalOptions) (byte, error) {
	__antithesis_instrumentation__.Notify(65050)
	var typ twkbType
	switch t.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(65052)
		typ = twkbTypePoint
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(65053)
		typ = twkbTypeLineString
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(65054)
		typ = twkbTypePolygon
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(65055)
		typ = twkbTypeMultiPoint
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(65056)
		typ = twkbTypeMultiLineString
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(65057)
		typ = twkbTypeMultiPolygon
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(65058)
		typ = twkbTypeGeometryCollection
	default:
		__antithesis_instrumentation__.Notify(65059)
		return 0, pgerror.Newf(pgcode.InvalidParameterValue, "unknown TWKB type: %T", t)
	}
	__antithesis_instrumentation__.Notify(65051)

	precisionZigZagged := zigzagInt8(o.precisionXY)
	return (uint8(typ) & 0x0F) | ((precisionZigZagged & 0x0F) << 4), nil
}

func metadataByte(t geom.T, o marshalOptions) byte {
	__antithesis_instrumentation__.Notify(65060)
	var b byte
	if t.Empty() {
		__antithesis_instrumentation__.Notify(65063)
		b |= 0b10000
	} else {
		__antithesis_instrumentation__.Notify(65064)
	}
	__antithesis_instrumentation__.Notify(65061)
	if t.Layout().Stride() > 2 {
		__antithesis_instrumentation__.Notify(65065)
		b |= 0b1000
	} else {
		__antithesis_instrumentation__.Notify(65066)
	}
	__antithesis_instrumentation__.Notify(65062)
	return b
}

func extendedDimensionsByte(t geom.T, o marshalOptions) byte {
	__antithesis_instrumentation__.Notify(65067)
	var extDimByte byte

	switch t.Layout() {
	case geom.XYZ:
		__antithesis_instrumentation__.Notify(65069)
		extDimByte |= 0b1
		extDimByte |= (byte(o.precisionZ) & 0b111) << 2
	case geom.XYM:
		__antithesis_instrumentation__.Notify(65070)
		extDimByte |= 0b10
		extDimByte |= (byte(o.precisionM) & 0b111) << 5
	case geom.XYZM:
		__antithesis_instrumentation__.Notify(65071)
		extDimByte |= 0b11
		extDimByte |= (byte(o.precisionZ) & 0b111) << 2
		extDimByte |= (byte(o.precisionM) & 0b111) << 5
	default:
		__antithesis_instrumentation__.Notify(65072)
	}
	__antithesis_instrumentation__.Notify(65068)
	return extDimByte
}

func zigzagInt8(x int8) byte {
	__antithesis_instrumentation__.Notify(65073)
	if x < 0 {
		__antithesis_instrumentation__.Notify(65075)
		return (uint8(-1-x) << 1) | 0x01
	} else {
		__antithesis_instrumentation__.Notify(65076)
	}
	__antithesis_instrumentation__.Notify(65074)
	return uint8(x) << 1
}
