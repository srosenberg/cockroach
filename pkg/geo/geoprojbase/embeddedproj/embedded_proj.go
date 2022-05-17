// Package embeddedproj defines the format used to embed static projection data
// in the Cockroach binary. Is is used to load the data as well as to generate
// the data file (from cmd/generate-spatial-ref-sys).
package embeddedproj

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"compress/gzip"
	"encoding/json"
	"io"
)

type Spheroid struct {
	Hash       int64
	Radius     float64
	Flattening float64
}

type Projection struct {
	SRID      int
	AuthName  string
	AuthSRID  int
	SRText    string
	Proj4Text string
	Bounds    Bounds
	IsLatLng  bool

	Spheroid int64
}

type Bounds struct {
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
}

type Data struct {
	Spheroids   []Spheroid
	Projections []Projection
}

func Encode(d Data, w io.Writer) error {
	__antithesis_instrumentation__.Notify(64047)
	data, err := json.MarshalIndent(d, "", " ")
	if err != nil {
		__antithesis_instrumentation__.Notify(64051)
		return err
	} else {
		__antithesis_instrumentation__.Notify(64052)
	}
	__antithesis_instrumentation__.Notify(64048)
	zw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		__antithesis_instrumentation__.Notify(64053)
		return err
	} else {
		__antithesis_instrumentation__.Notify(64054)
	}
	__antithesis_instrumentation__.Notify(64049)
	if _, err := zw.Write(data); err != nil {
		__antithesis_instrumentation__.Notify(64055)
		return err
	} else {
		__antithesis_instrumentation__.Notify(64056)
	}
	__antithesis_instrumentation__.Notify(64050)
	return zw.Close()
}

func Decode(r io.Reader) (Data, error) {
	__antithesis_instrumentation__.Notify(64057)
	zr, err := gzip.NewReader(r)
	if err != nil {
		__antithesis_instrumentation__.Notify(64060)
		return Data{}, err
	} else {
		__antithesis_instrumentation__.Notify(64061)
	}
	__antithesis_instrumentation__.Notify(64058)
	var result Data
	if err := json.NewDecoder(zr).Decode(&result); err != nil {
		__antithesis_instrumentation__.Notify(64062)
		return Data{}, err
	} else {
		__antithesis_instrumentation__.Notify(64063)
	}
	__antithesis_instrumentation__.Notify(64059)
	return result, nil
}
