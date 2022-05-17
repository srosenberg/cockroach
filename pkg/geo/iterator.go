package geo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

type GeomTIterator struct {
	g             geom.T
	emptyBehavior EmptyBehavior

	idx int

	subIt *GeomTIterator
}

func NewGeomTIterator(g geom.T, emptyBehavior EmptyBehavior) GeomTIterator {
	__antithesis_instrumentation__.Notify(64683)
	return GeomTIterator{g: g, emptyBehavior: emptyBehavior}
}

func (it *GeomTIterator) Next() (geom.T, bool, error) {
	__antithesis_instrumentation__.Notify(64684)
	next, hasNext, err := it.next()
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(64686)
		return !hasNext == true
	}() == true {
		__antithesis_instrumentation__.Notify(64687)
		return nil, hasNext, err
	} else {
		__antithesis_instrumentation__.Notify(64688)
	}
	__antithesis_instrumentation__.Notify(64685)
	for {
		__antithesis_instrumentation__.Notify(64689)
		if !next.Empty() {
			__antithesis_instrumentation__.Notify(64691)
			return next, hasNext, nil
		} else {
			__antithesis_instrumentation__.Notify(64692)
		}
		__antithesis_instrumentation__.Notify(64690)
		switch it.emptyBehavior {
		case EmptyBehaviorOmit:
			__antithesis_instrumentation__.Notify(64693)
			next, hasNext, err = it.next()
			if err != nil || func() bool {
				__antithesis_instrumentation__.Notify(64696)
				return !hasNext == true
			}() == true {
				__antithesis_instrumentation__.Notify(64697)
				return nil, hasNext, err
			} else {
				__antithesis_instrumentation__.Notify(64698)
			}
		case EmptyBehaviorError:
			__antithesis_instrumentation__.Notify(64694)
			return nil, false, NewEmptyGeometryError()
		default:
			__antithesis_instrumentation__.Notify(64695)
			return nil, false, errors.AssertionFailedf("programmer error: unknown behavior: %T", it.emptyBehavior)
		}
	}
}

func (it *GeomTIterator) next() (geom.T, bool, error) {
	__antithesis_instrumentation__.Notify(64699)
	switch t := it.g.(type) {
	case *geom.Point, *geom.LineString, *geom.Polygon:
		__antithesis_instrumentation__.Notify(64700)
		if it.idx == 1 {
			__antithesis_instrumentation__.Notify(64710)
			return nil, false, nil
		} else {
			__antithesis_instrumentation__.Notify(64711)
		}
		__antithesis_instrumentation__.Notify(64701)
		it.idx++
		return t, true, nil
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(64702)
		if it.idx == t.NumPoints() {
			__antithesis_instrumentation__.Notify(64712)
			return nil, false, nil
		} else {
			__antithesis_instrumentation__.Notify(64713)
		}
		__antithesis_instrumentation__.Notify(64703)
		p := t.Point(it.idx)
		it.idx++
		return p, true, nil
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(64704)
		if it.idx == t.NumLineStrings() {
			__antithesis_instrumentation__.Notify(64714)
			return nil, false, nil
		} else {
			__antithesis_instrumentation__.Notify(64715)
		}
		__antithesis_instrumentation__.Notify(64705)
		p := t.LineString(it.idx)
		it.idx++
		return p, true, nil
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(64706)
		if it.idx == t.NumPolygons() {
			__antithesis_instrumentation__.Notify(64716)
			return nil, false, nil
		} else {
			__antithesis_instrumentation__.Notify(64717)
		}
		__antithesis_instrumentation__.Notify(64707)
		p := t.Polygon(it.idx)
		it.idx++
		return p, true, nil
	case *geom.GeometryCollection:
		__antithesis_instrumentation__.Notify(64708)
		for {
			__antithesis_instrumentation__.Notify(64718)
			if it.idx == t.NumGeoms() {
				__antithesis_instrumentation__.Notify(64723)
				return nil, false, nil
			} else {
				__antithesis_instrumentation__.Notify(64724)
			}
			__antithesis_instrumentation__.Notify(64719)
			if it.subIt == nil {
				__antithesis_instrumentation__.Notify(64725)
				it.subIt = &GeomTIterator{g: t.Geom(it.idx), emptyBehavior: it.emptyBehavior}
			} else {
				__antithesis_instrumentation__.Notify(64726)
			}
			__antithesis_instrumentation__.Notify(64720)
			ret, next, err := it.subIt.next()
			if err != nil {
				__antithesis_instrumentation__.Notify(64727)
				return nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(64728)
			}
			__antithesis_instrumentation__.Notify(64721)
			if next {
				__antithesis_instrumentation__.Notify(64729)
				return ret, next, nil
			} else {
				__antithesis_instrumentation__.Notify(64730)
			}
			__antithesis_instrumentation__.Notify(64722)

			it.idx++
			it.subIt = nil
		}
	default:
		__antithesis_instrumentation__.Notify(64709)
		return nil, false, pgerror.Newf(pgcode.InvalidParameterValue, "unknown type: %T", t)
	}
}

func (it *GeomTIterator) Reset() {
	__antithesis_instrumentation__.Notify(64731)
	it.idx = 0
	it.subIt = nil
}
