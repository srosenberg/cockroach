package geospatial

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"compress/gzip"
	"embed"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"gopkg.in/yaml.v2"
)

type geospatial struct{}

func init() {
	workload.Register(geospatialMeta)
}

var geospatialMeta = workload.Meta{
	Name:         `geospatial`,
	Description:  `geospatial contains PostGIS tutorial tables`,
	Version:      `1.0.0`,
	PublicFacing: true,
	New:          func() workload.Generator { __antithesis_instrumentation__.Notify(694132); return geospatial{} },
}

func (geospatial) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(694133)
	return geospatialMeta
}

var embedded embed.FS

func (geospatial) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(694134)

	load := func(name string) [][]interface{} {
		__antithesis_instrumentation__.Notify(694136)
		compressed, err := embedded.ReadFile(fmt.Sprintf("data/%s.yaml.gz", name))
		if err != nil {
			__antithesis_instrumentation__.Notify(694140)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(694141)
		}
		__antithesis_instrumentation__.Notify(694137)
		r, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			__antithesis_instrumentation__.Notify(694142)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(694143)
		}
		__antithesis_instrumentation__.Notify(694138)
		var result [][]interface{}
		if err := yaml.NewDecoder(r).Decode(&result); err != nil {
			__antithesis_instrumentation__.Notify(694144)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(694145)
		}
		__antithesis_instrumentation__.Notify(694139)
		return result
	}
	__antithesis_instrumentation__.Notify(694135)

	nycCensusBlocksRows := load("nyc_census_blocks")
	nycHomicidesRows := load("nyc_homicides")
	nycNeighborhoodsRows := load("nyc_neighborhoods")
	nycStreetsRows := load("nyc_streets")
	nycSubwayStationsRows := load("nyc_subway_stations")
	subwayLinesRows := load("subway_lines")

	return []workload.Table{
		{
			Name: `nyc_census_blocks`,
			Schema: `(
				gid INT8 NOT NULL DEFAULT unique_rowid(),
				lkid VARCHAR(15) NULL,
				popn_total FLOAT8 NULL,
				popn_white FLOAT8 NULL,
				popn_black FLOAT8 NULL,
				popn_nativ FLOAT8 NULL,
				popn_asian FLOAT8 NULL,
				popn_other FLOAT8 NULL,
				boroname VARCHAR(32) NULL,
				geom GEOMETRY(MULTIPOLYGON,26918) NULL,
				CONSTRAINT "primary" PRIMARY KEY (gid ASC)
			)`,
			InitialRows: workload.Tuples(
				len(nycCensusBlocksRows),
				func(rowIdx int) []interface{} {
					__antithesis_instrumentation__.Notify(694146)
					return nycCensusBlocksRows[rowIdx]
				},
			),
		},
		{
			Name: `nyc_homicides`,
			Schema: `(
				gid INT8 NOT NULL DEFAULT unique_rowid(),
				incident_d DATE NULL,
				boroname VARCHAR(13) NULL,
				num_victim VARCHAR(1) NULL,
				primary_mo VARCHAR(20) NULL,
				id FLOAT8 NULL,
				weapon VARCHAR(16) NULL,
				light_dark VARCHAR(1) NULL,
				year FLOAT8 NULL,
				geom GEOMETRY(POINT,26918) NULL,
				CONSTRAINT "primary" PRIMARY KEY (gid ASC)
			)`,
			InitialRows: workload.TypedTuples(
				len(nycHomicidesRows),
				[]*types.T{
					types.Int,
					types.String,
					types.String,
					types.String,
					types.String,
					types.Float,
					types.String,
					types.String,
					types.Float,
					types.String,
				},
				func(rowIdx int) []interface{} {
					__antithesis_instrumentation__.Notify(694147)
					return nycHomicidesRows[rowIdx]
				},
			),
		},
		{
			Name: `nyc_neighborhoods`,
			Schema: `(
				gid INT8 NOT NULL DEFAULT unique_rowid(),
				boroname VARCHAR(43) NULL,
				name VARCHAR(64) NULL,
				geom GEOMETRY(MULTIPOLYGON,26918) NULL,
				CONSTRAINT "primary" PRIMARY KEY (gid ASC)
			)`,
			InitialRows: workload.Tuples(
				len(nycNeighborhoodsRows),
				func(rowIdx int) []interface{} {
					__antithesis_instrumentation__.Notify(694148)
					return nycNeighborhoodsRows[rowIdx]
				},
			),
		},
		{
			Name: `nyc_streets`,
			Schema: `(
				gid INT8 NOT NULL DEFAULT unique_rowid(),
				id FLOAT8 NULL,
				name VARCHAR(200) NULL,
				oneway VARCHAR(10) NULL,
				type VARCHAR(50) NULL,
				geom GEOMETRY(MULTILINESTRING,26918) NULL,
				CONSTRAINT "primary" PRIMARY KEY (gid ASC)
			)`,
			InitialRows: workload.TypedTuples(
				len(nycStreetsRows),
				[]*types.T{
					types.Int,
					types.Float,
					types.String,
					types.String,
					types.String,
					types.String,
				},
				func(rowIdx int) []interface{} {
					__antithesis_instrumentation__.Notify(694149)
					return nycStreetsRows[rowIdx]
				},
			),
		},
		{
			Name: `nyc_subway_stations`,
			Schema: `(
				gid INT8 NOT NULL DEFAULT unique_rowid(),
				objectid DECIMAL NULL,
				id DECIMAL NULL,
				name VARCHAR(31) NULL,
				alt_name VARCHAR(38) NULL,
				cross_st VARCHAR(27) NULL,
				long_name VARCHAR(60) NULL,
				label VARCHAR(50) NULL,
				borough VARCHAR(15) NULL,
				nghbhd VARCHAR(30) NULL,
				routes VARCHAR(20) NULL,
				transfers VARCHAR(25) NULL,
				color VARCHAR(30) NULL,
				express VARCHAR(10) NULL,
				closed VARCHAR(10) NULL,
				geom GEOMETRY(POINT,26918) NULL,
				CONSTRAINT "primary" PRIMARY KEY (gid ASC)
			)`,
			InitialRows: workload.TypedTuples(
				len(nycSubwayStationsRows),
				[]*types.T{
					types.Int,
					types.Int,
					types.Float,
					types.String,
					types.String,
					types.String,
					types.String,
					types.String,
					types.String,
					types.String,
					types.String,
					types.String,
					types.String,
					types.String,
					types.String,
					types.String,
				},
				func(rowIdx int) []interface{} {
					__antithesis_instrumentation__.Notify(694150)
					return nycSubwayStationsRows[rowIdx]
				},
			),
		},
		{
			Name: `subway_lines`,
			Schema: `(
				route CHAR NULL
			)`,
			InitialRows: workload.Tuples(
				len(subwayLinesRows),
				func(rowIdx int) []interface{} {
					__antithesis_instrumentation__.Notify(694151)
					return subwayLinesRows[rowIdx]
				},
			),
		},
	}
}
