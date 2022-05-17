package geo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/golang/geo/s1"
	"github.com/pierrre/geohash"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"github.com/twpayne/go-geom/encoding/geojson"
	"github.com/twpayne/go-geom/encoding/kml"
	"github.com/twpayne/go-geom/encoding/wkb"
	"github.com/twpayne/go-geom/encoding/wkbcommon"
	"github.com/twpayne/go-geom/encoding/wkbhex"
	"github.com/twpayne/go-geom/encoding/wkt"
)

const DefaultGeoJSONDecimalDigits = 9

func SpatialObjectToWKT(so geopb.SpatialObject, maxDecimalDigits int) (geopb.WKT, error) {
	__antithesis_instrumentation__.Notify(58681)
	t, err := ewkb.Unmarshal([]byte(so.EWKB))
	if err != nil {
		__antithesis_instrumentation__.Notify(58683)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(58684)
	}
	__antithesis_instrumentation__.Notify(58682)
	ret, err := wkt.Marshal(t, wkt.EncodeOptionWithMaxDecimalDigits(maxDecimalDigits))
	return geopb.WKT(ret), err
}

func SpatialObjectToEWKT(so geopb.SpatialObject, maxDecimalDigits int) (geopb.EWKT, error) {
	__antithesis_instrumentation__.Notify(58685)
	t, err := ewkb.Unmarshal([]byte(so.EWKB))
	if err != nil {
		__antithesis_instrumentation__.Notify(58689)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(58690)
	}
	__antithesis_instrumentation__.Notify(58686)
	ret, err := wkt.Marshal(t, wkt.EncodeOptionWithMaxDecimalDigits(maxDecimalDigits))
	if err != nil {
		__antithesis_instrumentation__.Notify(58691)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(58692)
	}
	__antithesis_instrumentation__.Notify(58687)
	if t.SRID() != 0 {
		__antithesis_instrumentation__.Notify(58693)
		ret = fmt.Sprintf("SRID=%d;%s", t.SRID(), ret)
	} else {
		__antithesis_instrumentation__.Notify(58694)
	}
	__antithesis_instrumentation__.Notify(58688)
	return geopb.EWKT(ret), err
}

func SpatialObjectToWKB(so geopb.SpatialObject, byteOrder binary.ByteOrder) (geopb.WKB, error) {
	__antithesis_instrumentation__.Notify(58695)
	t, err := ewkb.Unmarshal([]byte(so.EWKB))
	if err != nil {
		__antithesis_instrumentation__.Notify(58697)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(58698)
	}
	__antithesis_instrumentation__.Notify(58696)
	ret, err := wkb.Marshal(t, byteOrder, wkbcommon.WKBOptionEmptyPointHandling(wkbcommon.EmptyPointHandlingNaN))
	return geopb.WKB(ret), err
}

type SpatialObjectToGeoJSONFlag int

const (
	SpatialObjectToGeoJSONFlagIncludeBBox SpatialObjectToGeoJSONFlag = 1 << (iota)
	SpatialObjectToGeoJSONFlagShortCRS
	SpatialObjectToGeoJSONFlagLongCRS
	SpatialObjectToGeoJSONFlagShortCRSIfNot4326

	SpatialObjectToGeoJSONFlagZero = 0
)

func geomToGeoJSONCRS(t geom.T, long bool) (*geojson.CRS, error) {
	__antithesis_instrumentation__.Notify(58699)
	projection, err := geoprojbase.Projection(geopb.SRID(t.SRID()))
	if err != nil {
		__antithesis_instrumentation__.Notify(58702)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(58703)
	}
	__antithesis_instrumentation__.Notify(58700)
	var prop string
	if long {
		__antithesis_instrumentation__.Notify(58704)
		prop = fmt.Sprintf("urn:ogc:def:crs:%s::%d", projection.AuthName, projection.AuthSRID)
	} else {
		__antithesis_instrumentation__.Notify(58705)
		prop = fmt.Sprintf("%s:%d", projection.AuthName, projection.AuthSRID)
	}
	__antithesis_instrumentation__.Notify(58701)
	crs := &geojson.CRS{
		Type: "name",
		Properties: map[string]interface{}{
			"name": prop,
		},
	}
	return crs, nil
}

func SpatialObjectToGeoJSON(
	so geopb.SpatialObject, maxDecimalDigits int, flag SpatialObjectToGeoJSONFlag,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(58706)
	t, err := ewkb.Unmarshal([]byte(so.EWKB))
	if err != nil {
		__antithesis_instrumentation__.Notify(58710)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(58711)
	}
	__antithesis_instrumentation__.Notify(58707)
	options := []geojson.EncodeGeometryOption{
		geojson.EncodeGeometryWithMaxDecimalDigits(maxDecimalDigits),
	}
	if flag&SpatialObjectToGeoJSONFlagIncludeBBox != 0 {
		__antithesis_instrumentation__.Notify(58712)

		if so.BoundingBox != nil {
			__antithesis_instrumentation__.Notify(58713)
			options = append(
				options,
				geojson.EncodeGeometryWithBBox(),
			)
		} else {
			__antithesis_instrumentation__.Notify(58714)
		}
	} else {
		__antithesis_instrumentation__.Notify(58715)
	}
	__antithesis_instrumentation__.Notify(58708)

	if t.SRID() != 0 {
		__antithesis_instrumentation__.Notify(58716)
		if flag&SpatialObjectToGeoJSONFlagLongCRS != 0 {
			__antithesis_instrumentation__.Notify(58717)
			crs, err := geomToGeoJSONCRS(t, true)
			if err != nil {
				__antithesis_instrumentation__.Notify(58719)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(58720)
			}
			__antithesis_instrumentation__.Notify(58718)
			options = append(options, geojson.EncodeGeometryWithCRS(crs))
		} else {
			__antithesis_instrumentation__.Notify(58721)
			if flag&SpatialObjectToGeoJSONFlagShortCRS != 0 {
				__antithesis_instrumentation__.Notify(58722)
				crs, err := geomToGeoJSONCRS(t, false)
				if err != nil {
					__antithesis_instrumentation__.Notify(58724)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(58725)
				}
				__antithesis_instrumentation__.Notify(58723)
				options = append(options, geojson.EncodeGeometryWithCRS(crs))
			} else {
				__antithesis_instrumentation__.Notify(58726)
				if flag&SpatialObjectToGeoJSONFlagShortCRSIfNot4326 != 0 {
					__antithesis_instrumentation__.Notify(58727)
					if t.SRID() != 4326 {
						__antithesis_instrumentation__.Notify(58728)
						crs, err := geomToGeoJSONCRS(t, false)
						if err != nil {
							__antithesis_instrumentation__.Notify(58730)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(58731)
						}
						__antithesis_instrumentation__.Notify(58729)
						options = append(options, geojson.EncodeGeometryWithCRS(crs))
					} else {
						__antithesis_instrumentation__.Notify(58732)
					}
				} else {
					__antithesis_instrumentation__.Notify(58733)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(58734)
	}
	__antithesis_instrumentation__.Notify(58709)

	return geojson.Marshal(t, options...)
}

func SpatialObjectToWKBHex(so geopb.SpatialObject) (string, error) {
	__antithesis_instrumentation__.Notify(58735)
	t, err := ewkb.Unmarshal([]byte(so.EWKB))
	if err != nil {
		__antithesis_instrumentation__.Notify(58737)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(58738)
	}
	__antithesis_instrumentation__.Notify(58736)
	ret, err := wkbhex.Encode(t, DefaultEWKBEncodingFormat, wkbcommon.WKBOptionEmptyPointHandling(wkbcommon.EmptyPointHandlingNaN))
	return strings.ToUpper(ret), err
}

func SpatialObjectToKML(so geopb.SpatialObject) (string, error) {
	__antithesis_instrumentation__.Notify(58739)
	t, err := ewkb.Unmarshal([]byte(so.EWKB))
	if err != nil {
		__antithesis_instrumentation__.Notify(58743)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(58744)
	}
	__antithesis_instrumentation__.Notify(58740)
	kmlElement, err := kml.Encode(t)
	if err != nil {
		__antithesis_instrumentation__.Notify(58745)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(58746)
	}
	__antithesis_instrumentation__.Notify(58741)
	var buf bytes.Buffer
	if err := kmlElement.Write(&buf); err != nil {
		__antithesis_instrumentation__.Notify(58747)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(58748)
	}
	__antithesis_instrumentation__.Notify(58742)
	return buf.String(), nil
}

const GeoHashAutoPrecision = 0

const GeoHashMaxPrecision = 20

func SpatialObjectToGeoHash(so geopb.SpatialObject, p int) (string, error) {
	__antithesis_instrumentation__.Notify(58749)
	if so.BoundingBox == nil {
		__antithesis_instrumentation__.Notify(58755)
		return "", nil
	} else {
		__antithesis_instrumentation__.Notify(58756)
	}
	__antithesis_instrumentation__.Notify(58750)
	bbox := so.BoundingBox
	if so.Type == geopb.SpatialObjectType_GeographyType {
		__antithesis_instrumentation__.Notify(58757)

		bbox = &geopb.BoundingBox{
			LoX: s1.Angle(bbox.LoX).Degrees(),
			HiX: s1.Angle(bbox.HiX).Degrees(),
			LoY: s1.Angle(bbox.LoY).Degrees(),
			HiY: s1.Angle(bbox.HiY).Degrees(),
		}
	} else {
		__antithesis_instrumentation__.Notify(58758)
	}
	__antithesis_instrumentation__.Notify(58751)
	if bbox.LoX < -180 || func() bool {
		__antithesis_instrumentation__.Notify(58759)
		return bbox.HiX > 180 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(58760)
		return bbox.LoY < -90 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(58761)
		return bbox.HiY > 90 == true
	}() == true {
		__antithesis_instrumentation__.Notify(58762)
		return "", pgerror.Newf(
			pgcode.InvalidParameterValue,
			"object has bounds greater than the bounds of lat/lng, got (%f %f, %f %f)",
			bbox.LoX, bbox.LoY,
			bbox.HiX, bbox.HiY,
		)
	} else {
		__antithesis_instrumentation__.Notify(58763)
	}
	__antithesis_instrumentation__.Notify(58752)

	if p <= GeoHashAutoPrecision {
		__antithesis_instrumentation__.Notify(58764)
		p = getPrecisionForBBox(bbox)
	} else {
		__antithesis_instrumentation__.Notify(58765)
	}
	__antithesis_instrumentation__.Notify(58753)

	if p > GeoHashMaxPrecision {
		__antithesis_instrumentation__.Notify(58766)
		p = GeoHashMaxPrecision
	} else {
		__antithesis_instrumentation__.Notify(58767)
	}
	__antithesis_instrumentation__.Notify(58754)

	bbCenterLng := bbox.LoX + (bbox.HiX-bbox.LoX)/2.0
	bbCenterLat := bbox.LoY + (bbox.HiY-bbox.LoY)/2.0

	return geohash.Encode(bbCenterLat, bbCenterLng, p), nil
}

func getPrecisionForBBox(bbox *geopb.BoundingBox) int {
	__antithesis_instrumentation__.Notify(58768)
	bitPrecision := 0

	if bbox.LoX == bbox.HiX && func() bool {
		__antithesis_instrumentation__.Notify(58771)
		return bbox.LoY == bbox.HiY == true
	}() == true {
		__antithesis_instrumentation__.Notify(58772)
		return GeoHashMaxPrecision
	} else {
		__antithesis_instrumentation__.Notify(58773)
	}
	__antithesis_instrumentation__.Notify(58769)

	lonMin := -180.0
	lonMax := 180.0
	latMin := -90.0
	latMax := 90.0

	for {
		__antithesis_instrumentation__.Notify(58774)
		lonWidth := lonMax - lonMin
		latWidth := latMax - latMin
		latMaxDelta, lonMaxDelta, latMinDelta, lonMinDelta := 0.0, 0.0, 0.0, 0.0

		if bbox.LoX > lonMin+lonWidth/2.0 {
			__antithesis_instrumentation__.Notify(58779)
			lonMinDelta = lonWidth / 2.0
		} else {
			__antithesis_instrumentation__.Notify(58780)
			if bbox.HiX < lonMax-lonWidth/2.0 {
				__antithesis_instrumentation__.Notify(58781)
				lonMaxDelta = lonWidth / -2.0
			} else {
				__antithesis_instrumentation__.Notify(58782)
			}
		}
		__antithesis_instrumentation__.Notify(58775)

		if bbox.LoY > latMin+latWidth/2.0 {
			__antithesis_instrumentation__.Notify(58783)
			latMinDelta = latWidth / 2.0
		} else {
			__antithesis_instrumentation__.Notify(58784)
			if bbox.HiY < latMax-latWidth/2.0 {
				__antithesis_instrumentation__.Notify(58785)
				latMaxDelta = latWidth / -2.0
			} else {
				__antithesis_instrumentation__.Notify(58786)
			}
		}
		__antithesis_instrumentation__.Notify(58776)

		precisionDelta := 0
		if lonMinDelta != 0.0 || func() bool {
			__antithesis_instrumentation__.Notify(58787)
			return lonMaxDelta != 0.0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(58788)
			lonMin += lonMinDelta
			lonMax += lonMaxDelta
			precisionDelta++
		} else {
			__antithesis_instrumentation__.Notify(58789)
			break
		}
		__antithesis_instrumentation__.Notify(58777)
		if latMinDelta != 0.0 || func() bool {
			__antithesis_instrumentation__.Notify(58790)
			return latMaxDelta != 0.0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(58791)
			latMin += latMinDelta
			latMax += latMaxDelta
			precisionDelta++
		} else {
			__antithesis_instrumentation__.Notify(58792)
			break
		}
		__antithesis_instrumentation__.Notify(58778)
		bitPrecision += precisionDelta
	}
	__antithesis_instrumentation__.Notify(58770)

	return bitPrecision / 5
}

func StringToByteOrder(s string) binary.ByteOrder {
	__antithesis_instrumentation__.Notify(58793)
	switch strings.ToLower(s) {
	case "ndr":
		__antithesis_instrumentation__.Notify(58794)
		return binary.LittleEndian
	case "xdr":
		__antithesis_instrumentation__.Notify(58795)
		return binary.BigEndian
	default:
		__antithesis_instrumentation__.Notify(58796)
		return DefaultEWKBEncodingFormat
	}
}
