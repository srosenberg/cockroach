package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/geo/geoproj"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase/embeddedproj"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

var (
	flagSRC = flag.String(
		"src",
		"",
		"The source of where the spatial_ref_sys data lives. Assumes a CSV with no header row separated by ;.",
	)
	flagDEST = flag.String(
		"dest",
		"proj.json.gz",
		"The resulting .json.gz file.",
	)
)

func main() {
	__antithesis_instrumentation__.Notify(40414)
	flag.Parse()

	data := buildData()

	out, err := os.Create(*flagDEST)
	if err != nil {
		__antithesis_instrumentation__.Notify(40417)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(40418)
	}
	__antithesis_instrumentation__.Notify(40415)
	if err := embeddedproj.Encode(data, out); err != nil {
		__antithesis_instrumentation__.Notify(40419)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(40420)
	}
	__antithesis_instrumentation__.Notify(40416)

	if err := out.Close(); err != nil {
		__antithesis_instrumentation__.Notify(40421)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(40422)
	}
}

func buildData() embeddedproj.Data {
	__antithesis_instrumentation__.Notify(40423)
	type spheroidKey struct {
		majorAxis           float64
		eccentricitySquared float64
	}
	foundSpheroids := make(map[spheroidKey]int64)
	var mu syncutil.Mutex
	var d embeddedproj.Data

	g := ctxgroup.WithContext(context.Background())
	records := readRecords()
	const numWorkers = 8
	batchSize := int(math.Ceil(float64(len(records)) / numWorkers))
	for i := 0; i < len(records); i += batchSize {
		__antithesis_instrumentation__.Notify(40428)
		start := i
		end := i + batchSize
		if end > len(records) {
			__antithesis_instrumentation__.Notify(40430)
			end = len(records)
		} else {
			__antithesis_instrumentation__.Notify(40431)
		}
		__antithesis_instrumentation__.Notify(40429)
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(40432)
			for _, record := range records[start:end] {
				__antithesis_instrumentation__.Notify(40434)
				proj4text := strings.TrimRight(record[4], " \t")
				isLatLng, s, err := geoproj.GetProjMetadata(geoprojbase.MakeProj4Text(proj4text))
				if err != nil {
					__antithesis_instrumentation__.Notify(40440)
					log.Printf("error processing %s: %s, skipping", record[2], err)
					continue
				} else {
					__antithesis_instrumentation__.Notify(40441)
				}
				__antithesis_instrumentation__.Notify(40435)

				key := spheroidKey{s.Radius, s.Flattening}
				mu.Lock()
				spheroidHash, ok := foundSpheroids[key]
				if !ok {
					__antithesis_instrumentation__.Notify(40442)
					shaBytes := sha256.Sum256([]byte(
						strconv.FormatFloat(s.Radius, 'f', -1, 64) + "," + strconv.FormatFloat(s.Flattening, 'f', -1, 64),
					))
					spheroidHash = 0
					for _, b := range shaBytes[:6] {
						__antithesis_instrumentation__.Notify(40444)
						spheroidHash = (spheroidHash << 8) | int64(b)
					}
					__antithesis_instrumentation__.Notify(40443)
					foundSpheroids[key] = spheroidHash
					d.Spheroids = append(
						d.Spheroids,
						embeddedproj.Spheroid{
							Hash:       spheroidHash,
							Radius:     s.Radius,
							Flattening: s.Flattening,
						},
					)
				} else {
					__antithesis_instrumentation__.Notify(40445)
				}
				__antithesis_instrumentation__.Notify(40436)
				mu.Unlock()

				var bounds embeddedproj.Bounds
				if record[1] == "EPSG" {
					__antithesis_instrumentation__.Notify(40446)
					var results struct {
						Results []struct {
							BBox interface{} `json:"bbox,omitempty"`
							Code string      `json:"code"`
						} `json:"results"`
					}
					for _, searchArgs := range []string{
						record[2],
						fmt.Sprintf("%s%%20deprecated%%3A1", record[2]),
					} {
						__antithesis_instrumentation__.Notify(40456)
						var resp *http.Response
						for i := 0; i < 5; i++ {
							__antithesis_instrumentation__.Notify(40462)
							resp, err = httputil.Get(ctx, fmt.Sprintf("http://epsg.io/?q=%s&format=json", searchArgs))
							if err == nil {
								__antithesis_instrumentation__.Notify(40464)
								break
							} else {
								__antithesis_instrumentation__.Notify(40465)
							}
							__antithesis_instrumentation__.Notify(40463)
							log.Printf("http failure on %s, retrying; %v", record[2], err)
							time.Sleep(time.Duration(i) * time.Second * 2)
						}
						__antithesis_instrumentation__.Notify(40457)
						if err != nil {
							__antithesis_instrumentation__.Notify(40466)
							return err
						} else {
							__antithesis_instrumentation__.Notify(40467)
						}
						__antithesis_instrumentation__.Notify(40458)

						body, err := ioutil.ReadAll(resp.Body)
						resp.Body.Close()
						if err != nil {
							__antithesis_instrumentation__.Notify(40468)
							return err
						} else {
							__antithesis_instrumentation__.Notify(40469)
						}
						__antithesis_instrumentation__.Notify(40459)

						if err := json.Unmarshal(body, &results); err != nil {
							__antithesis_instrumentation__.Notify(40470)
							return err
						} else {
							__antithesis_instrumentation__.Notify(40471)
						}
						__antithesis_instrumentation__.Notify(40460)
						newResults := results.Results[:0]
						for i := range results.Results {
							__antithesis_instrumentation__.Notify(40472)
							if results.Results[i].Code == record[2] && func() bool {
								__antithesis_instrumentation__.Notify(40473)
								return results.Results[i].BBox != interface{}("") == true
							}() == true {
								__antithesis_instrumentation__.Notify(40474)
								newResults = append(newResults, results.Results[i])
							} else {
								__antithesis_instrumentation__.Notify(40475)
							}
						}
						__antithesis_instrumentation__.Notify(40461)
						results.Results = newResults
						if len(results.Results) > 0 {
							__antithesis_instrumentation__.Notify(40476)
							break
						} else {
							__antithesis_instrumentation__.Notify(40477)
						}
					}
					__antithesis_instrumentation__.Notify(40447)

					if len(results.Results) != 1 {
						__antithesis_instrumentation__.Notify(40478)
						log.Printf("WARNING: expected 1 result for %s, found %#v", record[2], results.Results)
					} else {
						__antithesis_instrumentation__.Notify(40479)
					}
					__antithesis_instrumentation__.Notify(40448)
					bbox := results.Results[0].BBox.([]interface{})

					xCoords := []float64{bbox[1].(float64), bbox[1].(float64), bbox[3].(float64), bbox[3].(float64)}
					yCoords := []float64{bbox[0].(float64), bbox[2].(float64), bbox[0].(float64), bbox[2].(float64)}
					if !isLatLng {
						__antithesis_instrumentation__.Notify(40480)
						if err := geoproj.Project(
							geoprojbase.MakeProj4Text("+proj=longlat +datum=WGS84 +no_defs"),
							geoprojbase.MakeProj4Text(proj4text),
							xCoords,
							yCoords,
							[]float64{0, 0, 0, 0},
						); err != nil {
							__antithesis_instrumentation__.Notify(40481)
							log.Printf("error processing %s: %s, skipping", record[2], err)
							continue
						} else {
							__antithesis_instrumentation__.Notify(40482)
						}
					} else {
						__antithesis_instrumentation__.Notify(40483)
					}
					__antithesis_instrumentation__.Notify(40449)

					sort.Slice(xCoords, func(i, j int) bool {
						__antithesis_instrumentation__.Notify(40484)
						return xCoords[i] < xCoords[j]
					})
					__antithesis_instrumentation__.Notify(40450)
					sort.Slice(yCoords, func(i, j int) bool {
						__antithesis_instrumentation__.Notify(40485)
						return yCoords[i] < yCoords[j]
					})
					__antithesis_instrumentation__.Notify(40451)
					skip := false
					for _, coord := range xCoords {
						__antithesis_instrumentation__.Notify(40486)
						if math.IsInf(coord, 1) || func() bool {
							__antithesis_instrumentation__.Notify(40487)
							return math.IsInf(coord, -1) == true
						}() == true {
							__antithesis_instrumentation__.Notify(40488)
							log.Printf("infinite coord at SRID %s, skipping", record[2])
							skip = true
							break
						} else {
							__antithesis_instrumentation__.Notify(40489)
						}
					}
					__antithesis_instrumentation__.Notify(40452)
					if skip {
						__antithesis_instrumentation__.Notify(40490)
						continue
					} else {
						__antithesis_instrumentation__.Notify(40491)
					}
					__antithesis_instrumentation__.Notify(40453)
					for _, coord := range yCoords {
						__antithesis_instrumentation__.Notify(40492)
						if math.IsInf(coord, 1) || func() bool {
							__antithesis_instrumentation__.Notify(40493)
							return math.IsInf(coord, -1) == true
						}() == true {
							__antithesis_instrumentation__.Notify(40494)
							log.Printf("infinite coord at SRID %s, skipping", record[2])
							skip = true
						} else {
							__antithesis_instrumentation__.Notify(40495)
						}
					}
					__antithesis_instrumentation__.Notify(40454)
					if skip {
						__antithesis_instrumentation__.Notify(40496)
						continue
					} else {
						__antithesis_instrumentation__.Notify(40497)
					}
					__antithesis_instrumentation__.Notify(40455)
					bounds = embeddedproj.Bounds{
						MinX: xCoords[0],
						MaxX: xCoords[3],
						MinY: yCoords[0],
						MaxY: yCoords[3],
					}
				} else {
					__antithesis_instrumentation__.Notify(40498)
				}
				__antithesis_instrumentation__.Notify(40437)

				srid, err := strconv.ParseInt(record[0], 0, 64)
				if err != nil {
					__antithesis_instrumentation__.Notify(40499)
					return err
				} else {
					__antithesis_instrumentation__.Notify(40500)
				}
				__antithesis_instrumentation__.Notify(40438)

				authSRID, err := strconv.ParseInt(record[2], 0, 64)
				if err != nil {
					__antithesis_instrumentation__.Notify(40501)
					return err
				} else {
					__antithesis_instrumentation__.Notify(40502)
				}
				__antithesis_instrumentation__.Notify(40439)

				mu.Lock()
				d.Projections = append(
					d.Projections,
					embeddedproj.Projection{
						SRID:      int(srid),
						AuthName:  record[1],
						AuthSRID:  int(authSRID),
						SRText:    record[3],
						Proj4Text: proj4text,

						Bounds:   bounds,
						IsLatLng: isLatLng,
						Spheroid: spheroidHash,
					},
				)
				mu.Unlock()
			}
			__antithesis_instrumentation__.Notify(40433)
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(40424)

	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(40503)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(40504)
	}
	__antithesis_instrumentation__.Notify(40425)
	sort.Slice(d.Projections, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(40505)
		return d.Projections[i].SRID < d.Projections[j].SRID
	})
	__antithesis_instrumentation__.Notify(40426)
	sort.Slice(d.Spheroids, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(40506)
		return d.Spheroids[i].Hash < d.Spheroids[j].Hash
	})
	__antithesis_instrumentation__.Notify(40427)
	return d
}

func readRecords() [][]string {
	__antithesis_instrumentation__.Notify(40507)
	in, err := os.Open(*flagSRC)
	if err != nil {
		__antithesis_instrumentation__.Notify(40511)
		log.Fatalf("cannot open CSV file '%s'", *flagSRC)
	} else {
		__antithesis_instrumentation__.Notify(40512)
	}
	__antithesis_instrumentation__.Notify(40508)
	defer func() {
		__antithesis_instrumentation__.Notify(40513)
		if err := in.Close(); err != nil {
			__antithesis_instrumentation__.Notify(40514)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(40515)
		}
	}()
	__antithesis_instrumentation__.Notify(40509)

	r := csv.NewReader(in)
	r.Comma = ';'

	records, err := r.ReadAll()
	if err != nil {
		__antithesis_instrumentation__.Notify(40516)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(40517)
	}
	__antithesis_instrumentation__.Notify(40510)
	return records
}
