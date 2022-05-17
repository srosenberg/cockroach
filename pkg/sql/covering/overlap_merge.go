package covering

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"reflect"
	"sort"
)

type Range struct {
	Start   []byte
	End     []byte
	Payload interface{}
}

type endpoints [][]byte

var _ sort.Interface = endpoints{}

func (e endpoints) Len() int {
	__antithesis_instrumentation__.Notify(461056)
	return len(e)
}

func (e endpoints) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(461057)
	return bytes.Compare(e[i], e[j]) < 0
}

func (e endpoints) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(461058)
	e[i], e[j] = e[j], e[i]
}

type marker struct {
	payload       interface{}
	coveringIndex int
}

type markers []marker

var _ sort.Interface = markers{}

func (m markers) Len() int {
	__antithesis_instrumentation__.Notify(461059)
	return len(m)
}

func (m markers) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(461060)

	return m[i].coveringIndex < m[j].coveringIndex
}

func (m markers) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(461061)
	m[i], m[j] = m[j], m[i]
}

type Covering []Range

func OverlapCoveringMerge(coverings []Covering) []Range {
	__antithesis_instrumentation__.Notify(461062)

	var totalRange endpoints

	numsMap := map[string]struct{}{}

	emptySets := map[string]struct{}{}

	startKeys := map[string]markers{}

	endKeys := map[string]markers{}

	for i, covering := range coverings {
		__antithesis_instrumentation__.Notify(461065)
		for _, r := range covering {
			__antithesis_instrumentation__.Notify(461066)
			startKeys[string(r.Start)] = append(startKeys[string(r.Start)], marker{
				payload:       r.Payload,
				coveringIndex: i,
			})
			if _, exist := numsMap[string(r.Start)]; !exist {
				__antithesis_instrumentation__.Notify(461070)
				totalRange = append(totalRange, r.Start)
				numsMap[string(r.Start)] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(461071)
			}
			__antithesis_instrumentation__.Notify(461067)
			endKeys[string(r.End)] = append(endKeys[string(r.End)], marker{
				payload:       r.Payload,
				coveringIndex: i,
			})

			if _, exist := numsMap[string(r.End)]; !exist {
				__antithesis_instrumentation__.Notify(461072)
				totalRange = append(totalRange, r.End)
				numsMap[string(r.End)] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(461073)
			}
			__antithesis_instrumentation__.Notify(461068)

			if !bytes.Equal(r.Start, r.End) {
				__antithesis_instrumentation__.Notify(461074)
				continue
			} else {
				__antithesis_instrumentation__.Notify(461075)
			}
			__antithesis_instrumentation__.Notify(461069)

			if _, exists := emptySets[string(r.Start)]; !exists {
				__antithesis_instrumentation__.Notify(461076)
				totalRange = append(totalRange, r.End)
				emptySets[string(r.Start)] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(461077)
			}
		}
	}
	__antithesis_instrumentation__.Notify(461063)
	sort.Sort(totalRange)

	var prev []byte
	var payloadsMarkers markers
	var ret []Range

	for it, next := range totalRange {
		__antithesis_instrumentation__.Notify(461078)
		if len(prev) != 0 && func() bool {
			__antithesis_instrumentation__.Notify(461083)
			return len(payloadsMarkers) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(461084)
			var payloads []interface{}

			sort.Sort(payloadsMarkers)

			for _, marker := range payloadsMarkers {
				__antithesis_instrumentation__.Notify(461086)
				payloads = append(payloads, marker.payload)
			}
			__antithesis_instrumentation__.Notify(461085)

			ret = append(ret, Range{
				Start:   prev,
				End:     next,
				Payload: payloads,
			})
		} else {
			__antithesis_instrumentation__.Notify(461087)
		}
		__antithesis_instrumentation__.Notify(461079)

		if removeMarkers, ok := endKeys[string(next)]; ok {
			__antithesis_instrumentation__.Notify(461088)
			for _, marker := range removeMarkers {
				__antithesis_instrumentation__.Notify(461089)
				var index = -1
				for i, p := range payloadsMarkers {
					__antithesis_instrumentation__.Notify(461091)
					if reflect.DeepEqual(p.payload, marker.payload) {
						__antithesis_instrumentation__.Notify(461092)
						index = i
						break
					} else {
						__antithesis_instrumentation__.Notify(461093)
					}
				}
				__antithesis_instrumentation__.Notify(461090)
				if index != -1 {
					__antithesis_instrumentation__.Notify(461094)
					payloadsMarkers = append(payloadsMarkers[:index], payloadsMarkers[index+1:]...)
				} else {
					__antithesis_instrumentation__.Notify(461095)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(461096)
		}
		__antithesis_instrumentation__.Notify(461080)

		if bytes.Equal(prev, next) && func() bool {
			__antithesis_instrumentation__.Notify(461097)
			return it > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(461098)

			continue
		} else {
			__antithesis_instrumentation__.Notify(461099)
		}
		__antithesis_instrumentation__.Notify(461081)

		if addMarkers, ok := startKeys[string(next)]; ok {
			__antithesis_instrumentation__.Notify(461100)
			payloadsMarkers = append(payloadsMarkers, addMarkers...)
		} else {
			__antithesis_instrumentation__.Notify(461101)
		}
		__antithesis_instrumentation__.Notify(461082)

		prev = next
	}
	__antithesis_instrumentation__.Notify(461064)

	return ret
}
