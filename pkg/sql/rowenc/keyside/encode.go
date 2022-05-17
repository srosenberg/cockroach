package keyside

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
)

func Encode(b []byte, val tree.Datum, dir encoding.Direction) ([]byte, error) {
	__antithesis_instrumentation__.Notify(570794)
	if (dir != encoding.Ascending) && func() bool {
		__antithesis_instrumentation__.Notify(570798)
		return (dir != encoding.Descending) == true
	}() == true {
		__antithesis_instrumentation__.Notify(570799)
		return nil, errors.Errorf("invalid direction: %d", dir)
	} else {
		__antithesis_instrumentation__.Notify(570800)
	}
	__antithesis_instrumentation__.Notify(570795)

	if val == tree.DNull {
		__antithesis_instrumentation__.Notify(570801)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570803)
			return encoding.EncodeNullAscending(b), nil
		} else {
			__antithesis_instrumentation__.Notify(570804)
		}
		__antithesis_instrumentation__.Notify(570802)
		return encoding.EncodeNullDescending(b), nil
	} else {
		__antithesis_instrumentation__.Notify(570805)
	}
	__antithesis_instrumentation__.Notify(570796)

	switch t := tree.UnwrapDatum(nil, val).(type) {
	case *tree.DBool:
		__antithesis_instrumentation__.Notify(570806)
		var x int64
		if *t {
			__antithesis_instrumentation__.Notify(570855)
			x = 1
		} else {
			__antithesis_instrumentation__.Notify(570856)
			x = 0
		}
		__antithesis_instrumentation__.Notify(570807)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570857)
			return encoding.EncodeVarintAscending(b, x), nil
		} else {
			__antithesis_instrumentation__.Notify(570858)
		}
		__antithesis_instrumentation__.Notify(570808)
		return encoding.EncodeVarintDescending(b, x), nil
	case *tree.DInt:
		__antithesis_instrumentation__.Notify(570809)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570859)
			return encoding.EncodeVarintAscending(b, int64(*t)), nil
		} else {
			__antithesis_instrumentation__.Notify(570860)
		}
		__antithesis_instrumentation__.Notify(570810)
		return encoding.EncodeVarintDescending(b, int64(*t)), nil
	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(570811)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570861)
			return encoding.EncodeFloatAscending(b, float64(*t)), nil
		} else {
			__antithesis_instrumentation__.Notify(570862)
		}
		__antithesis_instrumentation__.Notify(570812)
		return encoding.EncodeFloatDescending(b, float64(*t)), nil
	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(570813)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570863)
			return encoding.EncodeDecimalAscending(b, &t.Decimal), nil
		} else {
			__antithesis_instrumentation__.Notify(570864)
		}
		__antithesis_instrumentation__.Notify(570814)
		return encoding.EncodeDecimalDescending(b, &t.Decimal), nil
	case *tree.DString:
		__antithesis_instrumentation__.Notify(570815)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570865)
			return encoding.EncodeStringAscending(b, string(*t)), nil
		} else {
			__antithesis_instrumentation__.Notify(570866)
		}
		__antithesis_instrumentation__.Notify(570816)
		return encoding.EncodeStringDescending(b, string(*t)), nil
	case *tree.DBytes:
		__antithesis_instrumentation__.Notify(570817)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570867)
			return encoding.EncodeStringAscending(b, string(*t)), nil
		} else {
			__antithesis_instrumentation__.Notify(570868)
		}
		__antithesis_instrumentation__.Notify(570818)
		return encoding.EncodeStringDescending(b, string(*t)), nil
	case *tree.DVoid:
		__antithesis_instrumentation__.Notify(570819)
		return encoding.EncodeVoidAscendingOrDescending(b), nil
	case *tree.DBox2D:
		__antithesis_instrumentation__.Notify(570820)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570869)
			return encoding.EncodeBox2DAscending(b, t.CartesianBoundingBox.BoundingBox)
		} else {
			__antithesis_instrumentation__.Notify(570870)
		}
		__antithesis_instrumentation__.Notify(570821)
		return encoding.EncodeBox2DDescending(b, t.CartesianBoundingBox.BoundingBox)
	case *tree.DGeography:
		__antithesis_instrumentation__.Notify(570822)
		so := t.Geography.SpatialObjectRef()
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570871)
			return encoding.EncodeGeoAscending(b, t.Geography.SpaceCurveIndex(), so)
		} else {
			__antithesis_instrumentation__.Notify(570872)
		}
		__antithesis_instrumentation__.Notify(570823)
		return encoding.EncodeGeoDescending(b, t.Geography.SpaceCurveIndex(), so)
	case *tree.DGeometry:
		__antithesis_instrumentation__.Notify(570824)
		so := t.Geometry.SpatialObjectRef()
		spaceCurveIndex, err := t.Geometry.SpaceCurveIndex()
		if err != nil {
			__antithesis_instrumentation__.Notify(570873)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(570874)
		}
		__antithesis_instrumentation__.Notify(570825)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570875)
			return encoding.EncodeGeoAscending(b, spaceCurveIndex, so)
		} else {
			__antithesis_instrumentation__.Notify(570876)
		}
		__antithesis_instrumentation__.Notify(570826)
		return encoding.EncodeGeoDescending(b, spaceCurveIndex, so)
	case *tree.DDate:
		__antithesis_instrumentation__.Notify(570827)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570877)
			return encoding.EncodeVarintAscending(b, t.UnixEpochDaysWithOrig()), nil
		} else {
			__antithesis_instrumentation__.Notify(570878)
		}
		__antithesis_instrumentation__.Notify(570828)
		return encoding.EncodeVarintDescending(b, t.UnixEpochDaysWithOrig()), nil
	case *tree.DTime:
		__antithesis_instrumentation__.Notify(570829)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570879)
			return encoding.EncodeVarintAscending(b, int64(*t)), nil
		} else {
			__antithesis_instrumentation__.Notify(570880)
		}
		__antithesis_instrumentation__.Notify(570830)
		return encoding.EncodeVarintDescending(b, int64(*t)), nil
	case *tree.DTimestamp:
		__antithesis_instrumentation__.Notify(570831)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570881)
			return encoding.EncodeTimeAscending(b, t.Time), nil
		} else {
			__antithesis_instrumentation__.Notify(570882)
		}
		__antithesis_instrumentation__.Notify(570832)
		return encoding.EncodeTimeDescending(b, t.Time), nil
	case *tree.DTimestampTZ:
		__antithesis_instrumentation__.Notify(570833)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570883)
			return encoding.EncodeTimeAscending(b, t.Time), nil
		} else {
			__antithesis_instrumentation__.Notify(570884)
		}
		__antithesis_instrumentation__.Notify(570834)
		return encoding.EncodeTimeDescending(b, t.Time), nil
	case *tree.DTimeTZ:
		__antithesis_instrumentation__.Notify(570835)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570885)
			return encoding.EncodeTimeTZAscending(b, t.TimeTZ), nil
		} else {
			__antithesis_instrumentation__.Notify(570886)
		}
		__antithesis_instrumentation__.Notify(570836)
		return encoding.EncodeTimeTZDescending(b, t.TimeTZ), nil
	case *tree.DInterval:
		__antithesis_instrumentation__.Notify(570837)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570887)
			return encoding.EncodeDurationAscending(b, t.Duration)
		} else {
			__antithesis_instrumentation__.Notify(570888)
		}
		__antithesis_instrumentation__.Notify(570838)
		return encoding.EncodeDurationDescending(b, t.Duration)
	case *tree.DUuid:
		__antithesis_instrumentation__.Notify(570839)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570889)
			return encoding.EncodeBytesAscending(b, t.GetBytes()), nil
		} else {
			__antithesis_instrumentation__.Notify(570890)
		}
		__antithesis_instrumentation__.Notify(570840)
		return encoding.EncodeBytesDescending(b, t.GetBytes()), nil
	case *tree.DIPAddr:
		__antithesis_instrumentation__.Notify(570841)
		data := t.ToBuffer(nil)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570891)
			return encoding.EncodeBytesAscending(b, data), nil
		} else {
			__antithesis_instrumentation__.Notify(570892)
		}
		__antithesis_instrumentation__.Notify(570842)
		return encoding.EncodeBytesDescending(b, data), nil
	case *tree.DTuple:
		__antithesis_instrumentation__.Notify(570843)
		for _, datum := range t.D {
			__antithesis_instrumentation__.Notify(570893)
			var err error
			b, err = Encode(b, datum, dir)
			if err != nil {
				__antithesis_instrumentation__.Notify(570894)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(570895)
			}
		}
		__antithesis_instrumentation__.Notify(570844)
		return b, nil
	case *tree.DArray:
		__antithesis_instrumentation__.Notify(570845)
		return encodeArrayKey(b, t, dir)
	case *tree.DCollatedString:
		__antithesis_instrumentation__.Notify(570846)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570896)
			return encoding.EncodeBytesAscending(b, t.Key), nil
		} else {
			__antithesis_instrumentation__.Notify(570897)
		}
		__antithesis_instrumentation__.Notify(570847)
		return encoding.EncodeBytesDescending(b, t.Key), nil
	case *tree.DBitArray:
		__antithesis_instrumentation__.Notify(570848)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570898)
			return encoding.EncodeBitArrayAscending(b, t.BitArray), nil
		} else {
			__antithesis_instrumentation__.Notify(570899)
		}
		__antithesis_instrumentation__.Notify(570849)
		return encoding.EncodeBitArrayDescending(b, t.BitArray), nil
	case *tree.DOid:
		__antithesis_instrumentation__.Notify(570850)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570900)
			return encoding.EncodeVarintAscending(b, int64(t.DInt)), nil
		} else {
			__antithesis_instrumentation__.Notify(570901)
		}
		__antithesis_instrumentation__.Notify(570851)
		return encoding.EncodeVarintDescending(b, int64(t.DInt)), nil
	case *tree.DEnum:
		__antithesis_instrumentation__.Notify(570852)
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570902)
			return encoding.EncodeBytesAscending(b, t.PhysicalRep), nil
		} else {
			__antithesis_instrumentation__.Notify(570903)
		}
		__antithesis_instrumentation__.Notify(570853)
		return encoding.EncodeBytesDescending(b, t.PhysicalRep), nil
	case *tree.DJSON:
		__antithesis_instrumentation__.Notify(570854)
		return nil, unimplemented.NewWithIssue(35706, "unable to encode JSON as a table key")
	}
	__antithesis_instrumentation__.Notify(570797)
	return nil, errors.Errorf("unable to encode table key: %T", val)
}
