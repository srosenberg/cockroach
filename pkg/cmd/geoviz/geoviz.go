package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/sha1"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/golang/geo/s2"
)

type Image struct {
	Objects []Object `json:"objects"`
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Polygon struct {
	Geodesic    bool       `json:"geodesic"`
	StrokeColor string     `json:"strokeColor"`
	FillColor   string     `json:"fillColor"`
	Paths       [][]LatLng `json:"paths"`
	Marker      Marker     `json:"marker"`
	Label       string     `json:"label"`
}

type Polyline struct {
	Geodesic    bool     `json:"geodesic"`
	StrokeColor string   `json:"strokeColor"`
	Path        []LatLng `json:"path"`
	Marker      Marker   `json:"marker"`
}

type Marker struct {
	Title string `json:"title"`

	Content  string `json:"content"`
	Position LatLng `json:"position"`
	Label    string `json:"label"`
}

type Object struct {
	Title     string     `json:"title"`
	Color     string     `json:"color"`
	Polygons  []Polygon  `json:"polygons"`
	Polylines []Polyline `json:"polylines"`
	Markers   []Marker   `json:"markers"`
}

func s2PointToLatLng(fromPoint s2.Point) LatLng {
	__antithesis_instrumentation__.Notify(40574)
	from := s2.LatLngFromPoint(fromPoint)
	return LatLng{Lat: from.Lat.Degrees(), Lng: from.Lng.Degrees()}
}

func (img *Image) AddGeography(g geo.Geography, title string, color string) {
	__antithesis_instrumentation__.Notify(40575)
	regions, err := g.AsS2(geo.EmptyBehaviorOmit)
	if err != nil {
		__antithesis_instrumentation__.Notify(40578)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(40579)
	}
	__antithesis_instrumentation__.Notify(40576)
	object := Object{Title: title, Color: color}
	for _, region := range regions {
		__antithesis_instrumentation__.Notify(40580)
		switch region := region.(type) {
		case s2.Point:
			__antithesis_instrumentation__.Notify(40581)
			latlng := s2PointToLatLng(region)
			content := fmt.Sprintf("<h3>%s (point) (%f,%f)</h3>", title, latlng.Lat, latlng.Lng)
			object.Markers = append(
				object.Markers,
				Marker{
					Title:    title,
					Position: latlng,
					Content:  content,
					Label:    title,
				},
			)
		case *s2.Polyline:
			__antithesis_instrumentation__.Notify(40582)
			polyline := Polyline{
				Geodesic:    true,
				StrokeColor: color,
				Marker:      Marker{Title: title, Content: fmt.Sprintf("<h3>%s (line)</h3>", title)},
			}
			for _, point := range *region {
				__antithesis_instrumentation__.Notify(40587)
				polyline.Path = append(polyline.Path, s2PointToLatLng(point))
			}
			__antithesis_instrumentation__.Notify(40583)
			object.Polylines = append(object.Polylines, polyline)
		case *s2.Polygon:
			__antithesis_instrumentation__.Notify(40584)
			polygon := Polygon{
				Geodesic:    true,
				StrokeColor: color,
				FillColor:   color,
				Marker:      Marker{Title: title, Content: fmt.Sprintf("<h3>%s (polygon)</h3>", title)},
				Label:       title,
			}
			outerLoop := []LatLng{}
			for _, point := range region.Loop(0).Vertices() {
				__antithesis_instrumentation__.Notify(40588)
				outerLoop = append(outerLoop, s2PointToLatLng(point))
			}
			__antithesis_instrumentation__.Notify(40585)
			innerLoops := [][]LatLng{}
			for _, loop := range region.Loops()[1:] {
				__antithesis_instrumentation__.Notify(40589)
				vertices := loop.Vertices()
				loopRet := make([]LatLng, len(vertices))
				for i, vertex := range vertices {
					__antithesis_instrumentation__.Notify(40591)
					loopRet[len(vertices)-1-i] = s2PointToLatLng(vertex)
				}
				__antithesis_instrumentation__.Notify(40590)
				innerLoops = append(innerLoops, loopRet)
			}
			__antithesis_instrumentation__.Notify(40586)
			polygon.Paths = append([][]LatLng{outerLoop}, innerLoops...)
			object.Polygons = append(object.Polygons, polygon)
		}
	}
	__antithesis_instrumentation__.Notify(40577)
	img.Objects = append(img.Objects, object)
}

func (img *Image) AddS2Cells(title string, color string, cellIDs ...s2.CellID) {
	__antithesis_instrumentation__.Notify(40592)
	object := Object{Title: title, Color: color}
	for _, cellID := range cellIDs {
		__antithesis_instrumentation__.Notify(40594)
		cell := s2.CellFromCellID(cellID)
		loop := []LatLng{}

		for i := 0; i < 4; i++ {
			__antithesis_instrumentation__.Notify(40596)
			point := cell.Vertex(i)
			loop = append(loop, s2PointToLatLng(point))
		}
		__antithesis_instrumentation__.Notify(40595)
		content := fmt.Sprintf(
			"<h3>%s (S2 cell)</h3><br/>cell ID: %d<br/>cell string: %s<br/>cell token: %s<br/>cell binary: %064b",
			title,
			cellID,
			cellID.String(),
			cellID.ToToken(),
			cellID,
		)
		polygon := Polygon{
			Geodesic:    true,
			StrokeColor: color,
			FillColor:   color,
			Paths:       [][]LatLng{loop},
			Marker:      Marker{Title: title, Content: content},
			Label:       title,
		}
		object.Polygons = append(object.Polygons, polygon)
	}
	__antithesis_instrumentation__.Notify(40593)
	img.Objects = append(img.Objects, object)
}

func ImageFromReader(r io.Reader) (*Image, error) {
	__antithesis_instrumentation__.Notify(40597)
	csvReader := csv.NewReader(r)
	gviz := &Image{}
	recordNum := 0
	for {
		__antithesis_instrumentation__.Notify(40599)
		record, err := csvReader.Read()
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(40605)
			break
		} else {
			__antithesis_instrumentation__.Notify(40606)
		}
		__antithesis_instrumentation__.Notify(40600)
		if err != nil {
			__antithesis_instrumentation__.Notify(40607)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(40608)
		}
		__antithesis_instrumentation__.Notify(40601)
		recordNum++

		if len(record) != 4 {
			__antithesis_instrumentation__.Notify(40609)
			return nil, fmt.Errorf("records must be of format op,name(can be blank),color(can be blank),data")
		} else {
			__antithesis_instrumentation__.Notify(40610)
		}
		__antithesis_instrumentation__.Notify(40602)
		op, name, color, data := record[0], record[1], record[2], record[3]

		if color == "" {
			__antithesis_instrumentation__.Notify(40611)
			hashed := sha1.Sum([]byte(op + name + data))
			color = fmt.Sprintf("#%x", hashed[:3])
		} else {
			__antithesis_instrumentation__.Notify(40612)
		}
		__antithesis_instrumentation__.Notify(40603)

		if name == "" {
			__antithesis_instrumentation__.Notify(40613)
			name = fmt.Sprintf("shape #%d", recordNum)
		} else {
			__antithesis_instrumentation__.Notify(40614)
		}
		__antithesis_instrumentation__.Notify(40604)
		switch op {
		case "draw", "drawgeography":
			__antithesis_instrumentation__.Notify(40615)
			g, err := geo.ParseGeography(data)
			if err != nil {
				__antithesis_instrumentation__.Notify(40632)
				return nil, fmt.Errorf("error parsing: %s", data)
			} else {
				__antithesis_instrumentation__.Notify(40633)
			}
			__antithesis_instrumentation__.Notify(40616)
			gviz.AddGeography(g, name, color)
		case "innercovering":
			__antithesis_instrumentation__.Notify(40617)
			g, err := geo.ParseGeography(data)
			if err != nil {
				__antithesis_instrumentation__.Notify(40634)
				return nil, fmt.Errorf("error parsing: %s", data)
			} else {
				__antithesis_instrumentation__.Notify(40635)
			}
			__antithesis_instrumentation__.Notify(40618)
			index := geoindex.NewS2GeographyIndex(geoindex.S2GeographyConfig{
				S2Config: &geoindex.S2Config{
					MinLevel: 4,
					MaxLevel: 30,
					MaxCells: 20,
				},
			})
			gviz.AddS2Cells(name, color, index.TestingInnerCovering(g)...)
		case "drawcellid":
			__antithesis_instrumentation__.Notify(40619)
			cellIDs := []s2.CellID{}
			for _, d := range strings.Split(data, ",") {
				__antithesis_instrumentation__.Notify(40636)
				parsed, err := strconv.ParseInt(d, 10, 64)
				if err != nil {
					__antithesis_instrumentation__.Notify(40638)
					return nil, fmt.Errorf("error parsing: %s", data)
				} else {
					__antithesis_instrumentation__.Notify(40639)
				}
				__antithesis_instrumentation__.Notify(40637)
				cellIDs = append(cellIDs, s2.CellID(parsed))
			}
			__antithesis_instrumentation__.Notify(40620)
			gviz.AddS2Cells(name, color, cellIDs...)
		case "drawcelltoken":
			__antithesis_instrumentation__.Notify(40621)
			cellIDs := []s2.CellID{}
			for _, d := range strings.Split(data, ",") {
				__antithesis_instrumentation__.Notify(40640)
				cellID := s2.CellIDFromToken(d)
				if cellID == 0 {
					__antithesis_instrumentation__.Notify(40642)
					return nil, fmt.Errorf("error parsing: %s", data)
				} else {
					__antithesis_instrumentation__.Notify(40643)
				}
				__antithesis_instrumentation__.Notify(40641)
				cellIDs = append(cellIDs, cellID)
			}
			__antithesis_instrumentation__.Notify(40622)
			gviz.AddS2Cells(name, color, cellIDs...)
		case "covering":
			__antithesis_instrumentation__.Notify(40623)
			g, err := geo.ParseGeography(data)
			if err != nil {
				__antithesis_instrumentation__.Notify(40644)
				return nil, fmt.Errorf("error parsing: %s", data)
			} else {
				__antithesis_instrumentation__.Notify(40645)
			}
			__antithesis_instrumentation__.Notify(40624)
			gS2, err := g.AsS2(geo.EmptyBehaviorOmit)
			if err != nil {
				__antithesis_instrumentation__.Notify(40646)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(40647)
			}
			__antithesis_instrumentation__.Notify(40625)
			rc := &s2.RegionCoverer{MinLevel: 0, MaxLevel: 30, MaxCells: 4}
			cellIDs := []s2.CellID{}
			for _, s := range gS2 {
				__antithesis_instrumentation__.Notify(40648)
				cellIDs = append(cellIDs, rc.Covering(s)...)
			}
			__antithesis_instrumentation__.Notify(40626)
			gviz.AddS2Cells(name, color, cellIDs...)
		case "interiorcovering":
			__antithesis_instrumentation__.Notify(40627)
			g, err := geo.ParseGeography(data)
			if err != nil {
				__antithesis_instrumentation__.Notify(40649)
				return nil, fmt.Errorf("error parsing: %s", data)
			} else {
				__antithesis_instrumentation__.Notify(40650)
			}
			__antithesis_instrumentation__.Notify(40628)
			gS2, err := g.AsS2(geo.EmptyBehaviorOmit)
			if err != nil {
				__antithesis_instrumentation__.Notify(40651)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(40652)
			}
			__antithesis_instrumentation__.Notify(40629)
			rc := &s2.RegionCoverer{MinLevel: 0, MaxLevel: 30, MaxCells: 4}
			cellIDs := []s2.CellID{}
			for _, s := range gS2 {
				__antithesis_instrumentation__.Notify(40653)
				cellIDs = append(cellIDs, rc.Covering(s)...)
			}
			__antithesis_instrumentation__.Notify(40630)
			gviz.AddS2Cells(name, color, cellIDs...)
		default:
			__antithesis_instrumentation__.Notify(40631)
		}
	}
	__antithesis_instrumentation__.Notify(40598)
	return gviz, nil
}
