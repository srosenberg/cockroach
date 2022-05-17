// Package geoprojbase is a minimal dependency package that contains
// basic metadata and data structures for SRIDs and their CRS
// transformations.
package geoprojbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/geo/geographiclib"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
)

type Proj4Text struct {
	cStr []byte
}

func MakeProj4Text(str string) Proj4Text {
	__antithesis_instrumentation__.Notify(64064)
	return Proj4Text{
		cStr: []byte(str + "\u0000"),
	}
}

func (p *Proj4Text) String() string {
	__antithesis_instrumentation__.Notify(64065)
	return string(p.cStr[:len(p.cStr)-1])
}

func (p *Proj4Text) Bytes() []byte {
	__antithesis_instrumentation__.Notify(64066)
	return p.cStr
}

func (p *Proj4Text) Equal(o Proj4Text) bool {
	__antithesis_instrumentation__.Notify(64067)
	return bytes.Equal(p.cStr, o.cStr)
}

type Bounds struct {
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
}

type ProjInfo struct {
	SRID geopb.SRID

	AuthName string

	AuthSRID int

	SRText string

	Proj4Text Proj4Text

	Bounds Bounds

	IsLatLng bool

	Spheroid *geographiclib.Spheroid
}

var ErrProjectionNotFound error = errors.Newf("projection not found")

func Projection(srid geopb.SRID) (ProjInfo, error) {
	__antithesis_instrumentation__.Notify(64068)
	projections := getProjections()
	p, exists := projections[srid]
	if !exists {
		__antithesis_instrumentation__.Notify(64070)
		return ProjInfo{}, errors.Mark(
			pgerror.Newf(pgcode.InvalidParameterValue, "projection for SRID %d does not exist", srid),
			ErrProjectionNotFound,
		)
	} else {
		__antithesis_instrumentation__.Notify(64071)
	}
	__antithesis_instrumentation__.Notify(64069)
	return p, nil
}

func MustProjection(srid geopb.SRID) ProjInfo {
	__antithesis_instrumentation__.Notify(64072)
	ret, err := Projection(srid)
	if err != nil {
		__antithesis_instrumentation__.Notify(64074)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(64075)
	}
	__antithesis_instrumentation__.Notify(64073)
	return ret
}

func AllProjections() []ProjInfo {
	__antithesis_instrumentation__.Notify(64076)
	projections := getProjections()
	ret := make([]ProjInfo, 0, len(projections))
	for _, p := range projections {
		__antithesis_instrumentation__.Notify(64079)
		ret = append(ret, p)
	}
	__antithesis_instrumentation__.Notify(64077)
	sort.Slice(ret, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(64080)
		return ret[i].SRID < ret[j].SRID
	})
	__antithesis_instrumentation__.Notify(64078)
	return ret
}
