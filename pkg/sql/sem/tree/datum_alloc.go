package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/util"
)

type DatumAlloc struct {
	_ util.NoCopy

	AllocSize int

	datumAlloc        []Datum
	dintAlloc         []DInt
	dfloatAlloc       []DFloat
	dstringAlloc      []DString
	dbytesAlloc       []DBytes
	dbitArrayAlloc    []DBitArray
	ddecimalAlloc     []DDecimal
	ddateAlloc        []DDate
	denumAlloc        []DEnum
	dbox2dAlloc       []DBox2D
	dgeometryAlloc    []DGeometry
	dgeographyAlloc   []DGeography
	dtimeAlloc        []DTime
	dtimetzAlloc      []DTimeTZ
	dtimestampAlloc   []DTimestamp
	dtimestampTzAlloc []DTimestampTZ
	dintervalAlloc    []DInterval
	duuidAlloc        []DUuid
	dipnetAlloc       []DIPAddr
	djsonAlloc        []DJSON
	dtupleAlloc       []DTuple
	doidAlloc         []DOid
	dvoidAlloc        []DVoid
	env               CollationEnvironment

	ewkbAlloc               []byte
	curEWKBAllocSize        int
	lastEWKBBeyondAllocSize bool
}

const defaultDatumAllocSize = 16
const datumAllocMultiplier = 4
const defaultEWKBAllocSize = 4096
const maxEWKBAllocSize = 16384

func (a *DatumAlloc) NewDatums(num int) Datums {
	__antithesis_instrumentation__.Notify(607403)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607406)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607407)
	}
	__antithesis_instrumentation__.Notify(607404)
	buf := &a.datumAlloc
	if len(*buf) < num {
		__antithesis_instrumentation__.Notify(607408)
		extensionSize := a.AllocSize
		if extTupleLen := num * datumAllocMultiplier; extensionSize < extTupleLen {
			__antithesis_instrumentation__.Notify(607410)
			extensionSize = extTupleLen
		} else {
			__antithesis_instrumentation__.Notify(607411)
		}
		__antithesis_instrumentation__.Notify(607409)
		*buf = make(Datums, extensionSize)
	} else {
		__antithesis_instrumentation__.Notify(607412)
	}
	__antithesis_instrumentation__.Notify(607405)
	r := (*buf)[:num]
	*buf = (*buf)[num:]
	return r
}

func (a *DatumAlloc) NewDInt(v DInt) *DInt {
	__antithesis_instrumentation__.Notify(607413)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607416)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607417)
	}
	__antithesis_instrumentation__.Notify(607414)
	buf := &a.dintAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607418)
		*buf = make([]DInt, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607419)
	}
	__antithesis_instrumentation__.Notify(607415)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDFloat(v DFloat) *DFloat {
	__antithesis_instrumentation__.Notify(607420)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607423)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607424)
	}
	__antithesis_instrumentation__.Notify(607421)
	buf := &a.dfloatAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607425)
		*buf = make([]DFloat, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607426)
	}
	__antithesis_instrumentation__.Notify(607422)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDString(v DString) *DString {
	__antithesis_instrumentation__.Notify(607427)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607430)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607431)
	}
	__antithesis_instrumentation__.Notify(607428)
	buf := &a.dstringAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607432)
		*buf = make([]DString, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607433)
	}
	__antithesis_instrumentation__.Notify(607429)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDCollatedString(contents string, locale string) (*DCollatedString, error) {
	__antithesis_instrumentation__.Notify(607434)
	return NewDCollatedString(contents, locale, &a.env)
}

func (a *DatumAlloc) NewDName(v DString) Datum {
	__antithesis_instrumentation__.Notify(607435)
	return NewDNameFromDString(a.NewDString(v))
}

func (a *DatumAlloc) NewDBytes(v DBytes) *DBytes {
	__antithesis_instrumentation__.Notify(607436)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607439)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607440)
	}
	__antithesis_instrumentation__.Notify(607437)
	buf := &a.dbytesAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607441)
		*buf = make([]DBytes, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607442)
	}
	__antithesis_instrumentation__.Notify(607438)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDBitArray(v DBitArray) *DBitArray {
	__antithesis_instrumentation__.Notify(607443)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607446)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607447)
	}
	__antithesis_instrumentation__.Notify(607444)
	buf := &a.dbitArrayAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607448)
		*buf = make([]DBitArray, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607449)
	}
	__antithesis_instrumentation__.Notify(607445)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDDecimal(v DDecimal) *DDecimal {
	__antithesis_instrumentation__.Notify(607450)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607453)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607454)
	}
	__antithesis_instrumentation__.Notify(607451)
	buf := &a.ddecimalAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607455)
		*buf = make([]DDecimal, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607456)
	}
	__antithesis_instrumentation__.Notify(607452)
	r := &(*buf)[0]
	r.Set(&v.Decimal)
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDDate(v DDate) *DDate {
	__antithesis_instrumentation__.Notify(607457)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607460)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607461)
	}
	__antithesis_instrumentation__.Notify(607458)
	buf := &a.ddateAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607462)
		*buf = make([]DDate, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607463)
	}
	__antithesis_instrumentation__.Notify(607459)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDEnum(v DEnum) *DEnum {
	__antithesis_instrumentation__.Notify(607464)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607467)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607468)
	}
	__antithesis_instrumentation__.Notify(607465)
	buf := &a.denumAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607469)
		*buf = make([]DEnum, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607470)
	}
	__antithesis_instrumentation__.Notify(607466)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDBox2D(v DBox2D) *DBox2D {
	__antithesis_instrumentation__.Notify(607471)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607474)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607475)
	}
	__antithesis_instrumentation__.Notify(607472)
	buf := &a.dbox2dAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607476)
		*buf = make([]DBox2D, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607477)
	}
	__antithesis_instrumentation__.Notify(607473)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDGeography(v DGeography) *DGeography {
	__antithesis_instrumentation__.Notify(607478)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607481)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607482)
	}
	__antithesis_instrumentation__.Notify(607479)
	buf := &a.dgeographyAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607483)
		*buf = make([]DGeography, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607484)
	}
	__antithesis_instrumentation__.Notify(607480)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDVoid() *DVoid {
	__antithesis_instrumentation__.Notify(607485)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607488)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607489)
	}
	__antithesis_instrumentation__.Notify(607486)
	buf := &a.dvoidAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607490)
		*buf = make([]DVoid, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607491)
	}
	__antithesis_instrumentation__.Notify(607487)
	r := &(*buf)[0]
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDGeographyEmpty() *DGeography {
	__antithesis_instrumentation__.Notify(607492)
	r := a.NewDGeography(DGeography{})
	a.giveBytesToEWKB(r.SpatialObjectRef())
	return r
}

func (a *DatumAlloc) DoneInitNewDGeo(so *geopb.SpatialObject) {
	__antithesis_instrumentation__.Notify(607493)

	a.lastEWKBBeyondAllocSize = len(so.EWKB) > maxEWKBAllocSize
	c := cap(so.EWKB)
	l := len(so.EWKB)
	if (c - l) > l {
		__antithesis_instrumentation__.Notify(607494)
		a.ewkbAlloc = so.EWKB[l:l:c]
		so.EWKB = so.EWKB[:l:l]
	} else {
		__antithesis_instrumentation__.Notify(607495)
	}
}

func (a *DatumAlloc) NewDGeometry(v DGeometry) *DGeometry {
	__antithesis_instrumentation__.Notify(607496)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607499)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607500)
	}
	__antithesis_instrumentation__.Notify(607497)
	buf := &a.dgeometryAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607501)
		*buf = make([]DGeometry, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607502)
	}
	__antithesis_instrumentation__.Notify(607498)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDGeometryEmpty() *DGeometry {
	__antithesis_instrumentation__.Notify(607503)
	r := a.NewDGeometry(DGeometry{})
	a.giveBytesToEWKB(r.SpatialObjectRef())
	return r
}

func (a *DatumAlloc) giveBytesToEWKB(so *geopb.SpatialObject) {
	__antithesis_instrumentation__.Notify(607504)
	if a.ewkbAlloc == nil && func() bool {
		__antithesis_instrumentation__.Notify(607505)
		return !a.lastEWKBBeyondAllocSize == true
	}() == true {
		__antithesis_instrumentation__.Notify(607506)
		if a.curEWKBAllocSize == 0 {
			__antithesis_instrumentation__.Notify(607508)
			a.curEWKBAllocSize = defaultEWKBAllocSize
		} else {
			__antithesis_instrumentation__.Notify(607509)
			if a.curEWKBAllocSize < maxEWKBAllocSize {
				__antithesis_instrumentation__.Notify(607510)
				a.curEWKBAllocSize *= 2
			} else {
				__antithesis_instrumentation__.Notify(607511)
			}
		}
		__antithesis_instrumentation__.Notify(607507)
		so.EWKB = make([]byte, 0, a.curEWKBAllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607512)
		so.EWKB = a.ewkbAlloc
		a.ewkbAlloc = nil
	}
}

func (a *DatumAlloc) NewDTime(v DTime) *DTime {
	__antithesis_instrumentation__.Notify(607513)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607516)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607517)
	}
	__antithesis_instrumentation__.Notify(607514)
	buf := &a.dtimeAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607518)
		*buf = make([]DTime, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607519)
	}
	__antithesis_instrumentation__.Notify(607515)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDTimeTZ(v DTimeTZ) *DTimeTZ {
	__antithesis_instrumentation__.Notify(607520)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607523)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607524)
	}
	__antithesis_instrumentation__.Notify(607521)
	buf := &a.dtimetzAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607525)
		*buf = make([]DTimeTZ, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607526)
	}
	__antithesis_instrumentation__.Notify(607522)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDTimestamp(v DTimestamp) *DTimestamp {
	__antithesis_instrumentation__.Notify(607527)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607530)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607531)
	}
	__antithesis_instrumentation__.Notify(607528)
	buf := &a.dtimestampAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607532)
		*buf = make([]DTimestamp, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607533)
	}
	__antithesis_instrumentation__.Notify(607529)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDTimestampTZ(v DTimestampTZ) *DTimestampTZ {
	__antithesis_instrumentation__.Notify(607534)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607537)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607538)
	}
	__antithesis_instrumentation__.Notify(607535)
	buf := &a.dtimestampTzAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607539)
		*buf = make([]DTimestampTZ, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607540)
	}
	__antithesis_instrumentation__.Notify(607536)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDInterval(v DInterval) *DInterval {
	__antithesis_instrumentation__.Notify(607541)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607544)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607545)
	}
	__antithesis_instrumentation__.Notify(607542)
	buf := &a.dintervalAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607546)
		*buf = make([]DInterval, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607547)
	}
	__antithesis_instrumentation__.Notify(607543)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDUuid(v DUuid) *DUuid {
	__antithesis_instrumentation__.Notify(607548)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607551)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607552)
	}
	__antithesis_instrumentation__.Notify(607549)
	buf := &a.duuidAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607553)
		*buf = make([]DUuid, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607554)
	}
	__antithesis_instrumentation__.Notify(607550)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDIPAddr(v DIPAddr) *DIPAddr {
	__antithesis_instrumentation__.Notify(607555)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607558)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607559)
	}
	__antithesis_instrumentation__.Notify(607556)
	buf := &a.dipnetAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607560)
		*buf = make([]DIPAddr, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607561)
	}
	__antithesis_instrumentation__.Notify(607557)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDJSON(v DJSON) *DJSON {
	__antithesis_instrumentation__.Notify(607562)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607565)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607566)
	}
	__antithesis_instrumentation__.Notify(607563)
	buf := &a.djsonAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607567)
		*buf = make([]DJSON, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607568)
	}
	__antithesis_instrumentation__.Notify(607564)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDTuple(v DTuple) *DTuple {
	__antithesis_instrumentation__.Notify(607569)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607572)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607573)
	}
	__antithesis_instrumentation__.Notify(607570)
	buf := &a.dtupleAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607574)
		*buf = make([]DTuple, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607575)
	}
	__antithesis_instrumentation__.Notify(607571)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}

func (a *DatumAlloc) NewDOid(v DOid) Datum {
	__antithesis_instrumentation__.Notify(607576)
	if a.AllocSize == 0 {
		__antithesis_instrumentation__.Notify(607579)
		a.AllocSize = defaultDatumAllocSize
	} else {
		__antithesis_instrumentation__.Notify(607580)
	}
	__antithesis_instrumentation__.Notify(607577)
	buf := &a.doidAlloc
	if len(*buf) == 0 {
		__antithesis_instrumentation__.Notify(607581)
		*buf = make([]DOid, a.AllocSize)
	} else {
		__antithesis_instrumentation__.Notify(607582)
	}
	__antithesis_instrumentation__.Notify(607578)
	r := &(*buf)[0]
	*r = v
	*buf = (*buf)[1:]
	return r
}
