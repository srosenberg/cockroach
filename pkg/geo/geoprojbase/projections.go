package geoprojbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	_ "embed"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/geo/geographiclib"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase/embeddedproj"
	"github.com/cockroachdb/errors"
)

var projData []byte

var once sync.Once
var projectionsInternal map[geopb.SRID]ProjInfo

func getProjections() map[geopb.SRID]ProjInfo {
	__antithesis_instrumentation__.Notify(64081)
	once.Do(func() {
		__antithesis_instrumentation__.Notify(64083)
		d, err := embeddedproj.Decode(bytes.NewReader(projData))
		if err != nil {
			__antithesis_instrumentation__.Notify(64086)
			panic(errors.NewAssertionErrorWithWrappedErrf(err, "error decoding embedded projection data"))
		} else {
			__antithesis_instrumentation__.Notify(64087)
		}
		__antithesis_instrumentation__.Notify(64084)

		spheroids := make(map[int64]*geographiclib.Spheroid, len(d.Spheroids))
		for _, s := range d.Spheroids {
			__antithesis_instrumentation__.Notify(64088)
			spheroids[s.Hash] = geographiclib.NewSpheroid(s.Radius, s.Flattening)
		}
		__antithesis_instrumentation__.Notify(64085)

		projectionsInternal = make(map[geopb.SRID]ProjInfo, len(d.Projections))
		for _, p := range d.Projections {
			__antithesis_instrumentation__.Notify(64089)
			srid := geopb.SRID(p.SRID)
			spheroid, ok := spheroids[p.Spheroid]
			if !ok {
				__antithesis_instrumentation__.Notify(64091)
				panic(errors.AssertionFailedf("embedded projection data contains invalid spheroid %x", p.Spheroid))
			} else {
				__antithesis_instrumentation__.Notify(64092)
			}
			__antithesis_instrumentation__.Notify(64090)
			projectionsInternal[srid] = ProjInfo{
				SRID:      srid,
				AuthName:  "EPSG",
				AuthSRID:  p.AuthSRID,
				SRText:    p.SRText,
				Proj4Text: MakeProj4Text(p.Proj4Text),
				Bounds: Bounds{
					MinX: p.Bounds.MinX,
					MaxX: p.Bounds.MaxX,
					MinY: p.Bounds.MinY,
					MaxY: p.Bounds.MaxY,
				},
				IsLatLng: p.IsLatLng,
				Spheroid: spheroid,
			}
		}
	})
	__antithesis_instrumentation__.Notify(64082)

	return projectionsInternal
}
