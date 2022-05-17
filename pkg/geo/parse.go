package geo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
	"github.com/pierrre/geohash"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"github.com/twpayne/go-geom/encoding/ewkbhex"
	"github.com/twpayne/go-geom/encoding/geojson"
	"github.com/twpayne/go-geom/encoding/wkb"
	"github.com/twpayne/go-geom/encoding/wkbcommon"
	"github.com/twpayne/go-geom/encoding/wkt"
)

func parseEWKBRaw(soType geopb.SpatialObjectType, in geopb.EWKB) (geopb.SpatialObject, error) {
	__antithesis_instrumentation__.Notify(64740)
	t, err := ewkb.Unmarshal(in)
	if err != nil {
		__antithesis_instrumentation__.Notify(64742)
		return geopb.SpatialObject{}, pgerror.Wrapf(err, pgcode.InvalidParameterValue, "error parsing EWKB")
	} else {
		__antithesis_instrumentation__.Notify(64743)
	}
	__antithesis_instrumentation__.Notify(64741)
	return spatialObjectFromGeomT(t, soType)
}

func parseAmbiguousText(
	soType geopb.SpatialObjectType, str string, defaultSRID geopb.SRID,
) (geopb.SpatialObject, error) {
	__antithesis_instrumentation__.Notify(64744)
	if len(str) == 0 {
		__antithesis_instrumentation__.Notify(64747)
		return geopb.SpatialObject{}, pgerror.Newf(pgcode.InvalidParameterValue, "geo: parsing empty string to geo type")
	} else {
		__antithesis_instrumentation__.Notify(64748)
	}
	__antithesis_instrumentation__.Notify(64745)

	switch str[0] {
	case '0':
		__antithesis_instrumentation__.Notify(64749)
		return parseEWKBHex(soType, str, defaultSRID)
	case 0x00, 0x01:
		__antithesis_instrumentation__.Notify(64750)
		return parseEWKB(soType, []byte(str), defaultSRID, DefaultSRIDIsHint)
	case '{':
		__antithesis_instrumentation__.Notify(64751)
		return parseGeoJSON(soType, []byte(str), defaultSRID)
	default:
		__antithesis_instrumentation__.Notify(64752)
	}
	__antithesis_instrumentation__.Notify(64746)

	return parseEWKT(soType, geopb.EWKT(str), defaultSRID, DefaultSRIDIsHint)
}

func parseEWKBHex(
	soType geopb.SpatialObjectType, str string, defaultSRID geopb.SRID,
) (geopb.SpatialObject, error) {
	__antithesis_instrumentation__.Notify(64753)
	t, err := ewkbhex.Decode(str)
	if err != nil {
		__antithesis_instrumentation__.Notify(64756)
		return geopb.SpatialObject{}, pgerror.Wrapf(err, pgcode.InvalidParameterValue, "error parsing EWKB hex")
	} else {
		__antithesis_instrumentation__.Notify(64757)
	}
	__antithesis_instrumentation__.Notify(64754)
	if (defaultSRID != 0 && func() bool {
		__antithesis_instrumentation__.Notify(64758)
		return t.SRID() == 0 == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(64759)
		return int32(t.SRID()) < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(64760)
		AdjustGeomTSRID(t, defaultSRID)
	} else {
		__antithesis_instrumentation__.Notify(64761)
	}
	__antithesis_instrumentation__.Notify(64755)
	return spatialObjectFromGeomT(t, soType)
}

func parseEWKB(
	soType geopb.SpatialObjectType,
	b []byte,
	defaultSRID geopb.SRID,
	overwrite defaultSRIDOverwriteSetting,
) (geopb.SpatialObject, error) {
	__antithesis_instrumentation__.Notify(64762)
	t, err := ewkb.Unmarshal(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(64765)
		return geopb.SpatialObject{}, pgerror.Wrapf(err, pgcode.InvalidParameterValue, "error parsing EWKB")
	} else {
		__antithesis_instrumentation__.Notify(64766)
	}
	__antithesis_instrumentation__.Notify(64763)
	if overwrite == DefaultSRIDShouldOverwrite || func() bool {
		__antithesis_instrumentation__.Notify(64767)
		return (defaultSRID != 0 && func() bool {
			__antithesis_instrumentation__.Notify(64768)
			return t.SRID() == 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(64769)
		return int32(t.SRID()) < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(64770)
		AdjustGeomTSRID(t, defaultSRID)
	} else {
		__antithesis_instrumentation__.Notify(64771)
	}
	__antithesis_instrumentation__.Notify(64764)
	return spatialObjectFromGeomT(t, soType)
}

func parseWKB(
	soType geopb.SpatialObjectType, b []byte, defaultSRID geopb.SRID,
) (geopb.SpatialObject, error) {
	__antithesis_instrumentation__.Notify(64772)
	t, err := wkb.Unmarshal(b, wkbcommon.WKBOptionEmptyPointHandling(wkbcommon.EmptyPointHandlingNaN))
	if err != nil {
		__antithesis_instrumentation__.Notify(64774)
		return geopb.SpatialObject{}, pgerror.Wrapf(err, pgcode.InvalidParameterValue, "error parsing WKB")
	} else {
		__antithesis_instrumentation__.Notify(64775)
	}
	__antithesis_instrumentation__.Notify(64773)
	AdjustGeomTSRID(t, defaultSRID)
	return spatialObjectFromGeomT(t, soType)
}

func parseGeoJSON(
	soType geopb.SpatialObjectType, b []byte, defaultSRID geopb.SRID,
) (geopb.SpatialObject, error) {
	__antithesis_instrumentation__.Notify(64776)
	var t geom.T
	if err := geojson.Unmarshal(b, &t); err != nil {
		__antithesis_instrumentation__.Notify(64780)
		return geopb.SpatialObject{}, pgerror.Wrapf(err, pgcode.InvalidParameterValue, "error parsing GeoJSON")
	} else {
		__antithesis_instrumentation__.Notify(64781)
	}
	__antithesis_instrumentation__.Notify(64777)
	if t == nil {
		__antithesis_instrumentation__.Notify(64782)
		return geopb.SpatialObject{}, pgerror.Newf(pgcode.InvalidParameterValue, "invalid GeoJSON input")
	} else {
		__antithesis_instrumentation__.Notify(64783)
	}
	__antithesis_instrumentation__.Notify(64778)
	if defaultSRID != 0 && func() bool {
		__antithesis_instrumentation__.Notify(64784)
		return t.SRID() == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(64785)
		AdjustGeomTSRID(t, defaultSRID)
	} else {
		__antithesis_instrumentation__.Notify(64786)
	}
	__antithesis_instrumentation__.Notify(64779)
	return spatialObjectFromGeomT(t, soType)
}

const sridPrefix = "SRID="
const sridPrefixLen = len(sridPrefix)

type defaultSRIDOverwriteSetting bool

const (
	DefaultSRIDShouldOverwrite defaultSRIDOverwriteSetting = true

	DefaultSRIDIsHint defaultSRIDOverwriteSetting = false
)

func parseEWKT(
	soType geopb.SpatialObjectType,
	str geopb.EWKT,
	defaultSRID geopb.SRID,
	overwrite defaultSRIDOverwriteSetting,
) (geopb.SpatialObject, error) {
	__antithesis_instrumentation__.Notify(64787)
	srid := defaultSRID
	if hasPrefixIgnoreCase(string(str), sridPrefix) {
		__antithesis_instrumentation__.Notify(64790)
		end := strings.Index(string(str[sridPrefixLen:]), ";")
		if end != -1 {
			__antithesis_instrumentation__.Notify(64791)
			if overwrite != DefaultSRIDShouldOverwrite {
				__antithesis_instrumentation__.Notify(64793)
				sridInt64, err := strconv.ParseInt(string(str[sridPrefixLen:sridPrefixLen+end]), 10, 32)
				if err != nil {
					__antithesis_instrumentation__.Notify(64795)
					return geopb.SpatialObject{}, pgerror.Wrapf(
						err,
						pgcode.InvalidParameterValue,
						"error parsing SRID for EWKT",
					)
				} else {
					__antithesis_instrumentation__.Notify(64796)
				}
				__antithesis_instrumentation__.Notify(64794)

				if sridInt64 > 0 {
					__antithesis_instrumentation__.Notify(64797)
					srid = geopb.SRID(sridInt64)
				} else {
					__antithesis_instrumentation__.Notify(64798)
				}
			} else {
				__antithesis_instrumentation__.Notify(64799)
			}
			__antithesis_instrumentation__.Notify(64792)
			str = str[sridPrefixLen+end+1:]
		} else {
			__antithesis_instrumentation__.Notify(64800)
			return geopb.SpatialObject{}, pgerror.Newf(
				pgcode.InvalidParameterValue,
				"geo: failed to find ; character with SRID declaration during EWKT decode: %q",
				str,
			)
		}
	} else {
		__antithesis_instrumentation__.Notify(64801)
	}
	__antithesis_instrumentation__.Notify(64788)

	g, wktUnmarshalErr := wkt.Unmarshal(string(str))
	if wktUnmarshalErr != nil {
		__antithesis_instrumentation__.Notify(64802)
		return geopb.SpatialObject{}, pgerror.Wrap(
			wktUnmarshalErr,
			pgcode.InvalidParameterValue,
			"error parsing EWKT",
		)
	} else {
		__antithesis_instrumentation__.Notify(64803)
	}
	__antithesis_instrumentation__.Notify(64789)
	AdjustGeomTSRID(g, srid)
	return spatialObjectFromGeomT(g, soType)
}

func hasPrefixIgnoreCase(str string, prefix string) bool {
	__antithesis_instrumentation__.Notify(64804)
	if len(str) < len(prefix) {
		__antithesis_instrumentation__.Notify(64807)
		return false
	} else {
		__antithesis_instrumentation__.Notify(64808)
	}
	__antithesis_instrumentation__.Notify(64805)
	for i := 0; i < len(prefix); i++ {
		__antithesis_instrumentation__.Notify(64809)
		if util.ToLowerSingleByte(str[i]) != util.ToLowerSingleByte(prefix[i]) {
			__antithesis_instrumentation__.Notify(64810)
			return false
		} else {
			__antithesis_instrumentation__.Notify(64811)
		}
	}
	__antithesis_instrumentation__.Notify(64806)
	return true
}

func ParseGeometryPointFromGeoHash(g string, precision int) (Geometry, error) {
	__antithesis_instrumentation__.Notify(64812)
	box, err := parseGeoHash(g, precision)
	if err != nil {
		__antithesis_instrumentation__.Notify(64815)
		return Geometry{}, err
	} else {
		__antithesis_instrumentation__.Notify(64816)
	}
	__antithesis_instrumentation__.Notify(64813)
	point := box.Center()
	geom, gErr := MakeGeometryFromPointCoords(point.Lon, point.Lat)
	if gErr != nil {
		__antithesis_instrumentation__.Notify(64817)
		return Geometry{}, gErr
	} else {
		__antithesis_instrumentation__.Notify(64818)
	}
	__antithesis_instrumentation__.Notify(64814)
	return geom, nil
}

func ParseCartesianBoundingBoxFromGeoHash(g string, precision int) (CartesianBoundingBox, error) {
	__antithesis_instrumentation__.Notify(64819)
	box, err := parseGeoHash(g, precision)
	if err != nil {
		__antithesis_instrumentation__.Notify(64821)
		return CartesianBoundingBox{}, err
	} else {
		__antithesis_instrumentation__.Notify(64822)
	}
	__antithesis_instrumentation__.Notify(64820)
	return CartesianBoundingBox{
		BoundingBox: geopb.BoundingBox{
			LoX: box.Lon.Min,
			HiX: box.Lon.Max,
			LoY: box.Lat.Min,
			HiY: box.Lat.Max,
		},
	}, nil
}

func parseGeoHash(g string, precision int) (geohash.Box, error) {
	__antithesis_instrumentation__.Notify(64823)
	if len(g) == 0 {
		__antithesis_instrumentation__.Notify(64827)
		return geohash.Box{}, pgerror.Newf(pgcode.InvalidParameterValue, "length of GeoHash must be greater than 0")
	} else {
		__antithesis_instrumentation__.Notify(64828)
	}
	__antithesis_instrumentation__.Notify(64824)

	g = strings.ToLower(g)

	if precision > len(g) || func() bool {
		__antithesis_instrumentation__.Notify(64829)
		return precision < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(64830)
		precision = len(g)
	} else {
		__antithesis_instrumentation__.Notify(64831)
	}
	__antithesis_instrumentation__.Notify(64825)
	box, err := geohash.Decode(g[:precision])
	if err != nil {
		__antithesis_instrumentation__.Notify(64832)
		return geohash.Box{}, pgerror.Wrap(err, pgcode.InvalidParameterValue, "invalid GeoHash")
	} else {
		__antithesis_instrumentation__.Notify(64833)
	}
	__antithesis_instrumentation__.Notify(64826)
	return box, nil
}

func GeometryToEncodedPolyline(g Geometry, p int) (string, error) {
	__antithesis_instrumentation__.Notify(64834)
	gt, err := g.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(64837)
		return "", errors.Wrap(err, "error parsing input geometry")
	} else {
		__antithesis_instrumentation__.Notify(64838)
	}
	__antithesis_instrumentation__.Notify(64835)
	if gt.SRID() != 4326 {
		__antithesis_instrumentation__.Notify(64839)
		return "", pgerror.Newf(pgcode.InvalidParameterValue, "only SRID 4326 is supported")
	} else {
		__antithesis_instrumentation__.Notify(64840)
	}
	__antithesis_instrumentation__.Notify(64836)

	return encodePolylinePoints(gt.FlatCoords(), p), nil
}

func ParseEncodedPolyline(encodedPolyline string, precision int) (Geometry, error) {
	__antithesis_instrumentation__.Notify(64841)
	flatCoords := decodePolylinePoints(encodedPolyline, precision)
	ls := geom.NewLineStringFlat(geom.XY, flatCoords).SetSRID(4326)

	g, err := MakeGeometryFromGeomT(ls)
	if err != nil {
		__antithesis_instrumentation__.Notify(64843)
		return Geometry{}, errors.Wrap(err, "parsing geography error")
	} else {
		__antithesis_instrumentation__.Notify(64844)
	}
	__antithesis_instrumentation__.Notify(64842)
	return g, nil
}
