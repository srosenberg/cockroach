package spanset

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "sort"

type sortedSpans []Span

func (s *sortedSpans) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(123178)

	c := (*s)[i].Key.Compare((*s)[j].Key)
	if c != 0 {
		__antithesis_instrumentation__.Notify(123180)
		return c < 0
	} else {
		__antithesis_instrumentation__.Notify(123181)
	}
	__antithesis_instrumentation__.Notify(123179)
	return (*s)[i].EndKey.Compare((*s)[j].EndKey) < 0
}

func (s *sortedSpans) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(123182)
	(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
}

func (s *sortedSpans) Len() int {
	__antithesis_instrumentation__.Notify(123183)
	return len(*s)
}

func mergeSpans(latches *[]Span) ([]Span, bool) {
	__antithesis_instrumentation__.Notify(123184)
	if len(*latches) == 0 {
		__antithesis_instrumentation__.Notify(123187)
		return *latches, true
	} else {
		__antithesis_instrumentation__.Notify(123188)
	}
	__antithesis_instrumentation__.Notify(123185)

	sort.Sort((*sortedSpans)(latches))

	r := (*latches)[:1]
	distinct := true

	for _, cur := range (*latches)[1:] {
		__antithesis_instrumentation__.Notify(123189)
		prev := &r[len(r)-1]
		if len(cur.EndKey) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(123193)
			return len(prev.EndKey) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(123194)
			if cur.Key.Compare(prev.Key) != 0 {
				__antithesis_instrumentation__.Notify(123196)

				r = append(r, cur)
			} else {
				__antithesis_instrumentation__.Notify(123197)

				if cur.Timestamp != prev.Timestamp {
					__antithesis_instrumentation__.Notify(123199)
					r = append(r, cur)
				} else {
					__antithesis_instrumentation__.Notify(123200)
				}
				__antithesis_instrumentation__.Notify(123198)
				distinct = false
			}
			__antithesis_instrumentation__.Notify(123195)
			continue
		} else {
			__antithesis_instrumentation__.Notify(123201)
		}
		__antithesis_instrumentation__.Notify(123190)
		if len(prev.EndKey) == 0 {
			__antithesis_instrumentation__.Notify(123202)
			if cur.Key.Compare(prev.Key) == 0 {
				__antithesis_instrumentation__.Notify(123204)

				if cur.Timestamp != prev.Timestamp {
					__antithesis_instrumentation__.Notify(123206)
					r = append(r, cur)
				} else {
					__antithesis_instrumentation__.Notify(123207)
					prev.EndKey = cur.EndKey
				}
				__antithesis_instrumentation__.Notify(123205)
				distinct = false
			} else {
				__antithesis_instrumentation__.Notify(123208)

				r = append(r, cur)
			}
			__antithesis_instrumentation__.Notify(123203)
			continue
		} else {
			__antithesis_instrumentation__.Notify(123209)
		}
		__antithesis_instrumentation__.Notify(123191)

		if c := prev.EndKey.Compare(cur.Key); c >= 0 {
			__antithesis_instrumentation__.Notify(123210)
			if cur.EndKey != nil {
				__antithesis_instrumentation__.Notify(123212)
				if prev.EndKey.Compare(cur.EndKey) < 0 {
					__antithesis_instrumentation__.Notify(123213)

					if cur.Timestamp != prev.Timestamp {
						__antithesis_instrumentation__.Notify(123215)
						r = append(r, cur)
					} else {
						__antithesis_instrumentation__.Notify(123216)
						prev.EndKey = cur.EndKey
					}
					__antithesis_instrumentation__.Notify(123214)
					if c > 0 {
						__antithesis_instrumentation__.Notify(123217)
						distinct = false
					} else {
						__antithesis_instrumentation__.Notify(123218)
					}
				} else {
					__antithesis_instrumentation__.Notify(123219)

					if cur.Timestamp != prev.Timestamp {
						__antithesis_instrumentation__.Notify(123221)
						r = append(r, cur)
					} else {
						__antithesis_instrumentation__.Notify(123222)
					}
					__antithesis_instrumentation__.Notify(123220)
					distinct = false
				}
			} else {
				__antithesis_instrumentation__.Notify(123223)
				if c == 0 {
					__antithesis_instrumentation__.Notify(123224)

					if cur.Timestamp != prev.Timestamp {
						__antithesis_instrumentation__.Notify(123226)
						r = append(r, cur)
					} else {
						__antithesis_instrumentation__.Notify(123227)
					}
					__antithesis_instrumentation__.Notify(123225)
					prev.EndKey = cur.Key.Next()
				} else {
					__antithesis_instrumentation__.Notify(123228)

					if cur.Timestamp != prev.Timestamp {
						__antithesis_instrumentation__.Notify(123230)
						r = append(r, cur)
					} else {
						__antithesis_instrumentation__.Notify(123231)
					}
					__antithesis_instrumentation__.Notify(123229)
					distinct = false
				}
			}
			__antithesis_instrumentation__.Notify(123211)
			continue
		} else {
			__antithesis_instrumentation__.Notify(123232)
		}
		__antithesis_instrumentation__.Notify(123192)
		r = append(r, cur)
	}
	__antithesis_instrumentation__.Notify(123186)
	return r, distinct
}
