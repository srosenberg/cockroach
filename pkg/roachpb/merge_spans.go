package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "sort"

type sortedSpans []Span

func (s *sortedSpans) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(173761)

	c := (*s)[i].Key.Compare((*s)[j].Key)
	if c != 0 {
		__antithesis_instrumentation__.Notify(173763)
		return c < 0
	} else {
		__antithesis_instrumentation__.Notify(173764)
	}
	__antithesis_instrumentation__.Notify(173762)
	return (*s)[i].EndKey.Compare((*s)[j].EndKey) < 0
}

func (s *sortedSpans) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(173765)
	(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
}

func (s *sortedSpans) Len() int {
	__antithesis_instrumentation__.Notify(173766)
	return len(*s)
}

func MergeSpans(spans *[]Span) ([]Span, bool) {
	__antithesis_instrumentation__.Notify(173767)
	if len(*spans) == 0 {
		__antithesis_instrumentation__.Notify(173770)
		return *spans, true
	} else {
		__antithesis_instrumentation__.Notify(173771)
	}
	__antithesis_instrumentation__.Notify(173768)

	sort.Sort((*sortedSpans)(spans))

	r := (*spans)[:1]
	distinct := true

	for _, cur := range (*spans)[1:] {
		__antithesis_instrumentation__.Notify(173772)
		prev := &r[len(r)-1]
		if len(cur.EndKey) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(173776)
			return len(prev.EndKey) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(173777)
			if cur.Key.Compare(prev.Key) != 0 {
				__antithesis_instrumentation__.Notify(173779)

				r = append(r, cur)
			} else {
				__antithesis_instrumentation__.Notify(173780)

				distinct = false
			}
			__antithesis_instrumentation__.Notify(173778)
			continue
		} else {
			__antithesis_instrumentation__.Notify(173781)
		}
		__antithesis_instrumentation__.Notify(173773)
		if len(prev.EndKey) == 0 {
			__antithesis_instrumentation__.Notify(173782)
			if cur.Key.Compare(prev.Key) == 0 {
				__antithesis_instrumentation__.Notify(173784)

				prev.EndKey = cur.EndKey
				distinct = false
			} else {
				__antithesis_instrumentation__.Notify(173785)

				r = append(r, cur)
			}
			__antithesis_instrumentation__.Notify(173783)
			continue
		} else {
			__antithesis_instrumentation__.Notify(173786)
		}
		__antithesis_instrumentation__.Notify(173774)
		if c := prev.EndKey.Compare(cur.Key); c >= 0 {
			__antithesis_instrumentation__.Notify(173787)
			if cur.EndKey != nil {
				__antithesis_instrumentation__.Notify(173789)
				if prev.EndKey.Compare(cur.EndKey) < 0 {
					__antithesis_instrumentation__.Notify(173790)

					prev.EndKey = cur.EndKey
					if c > 0 {
						__antithesis_instrumentation__.Notify(173791)
						distinct = false
					} else {
						__antithesis_instrumentation__.Notify(173792)
					}
				} else {
					__antithesis_instrumentation__.Notify(173793)

					distinct = false
				}
			} else {
				__antithesis_instrumentation__.Notify(173794)
				if c == 0 {
					__antithesis_instrumentation__.Notify(173795)

					prev.EndKey = cur.Key.Next()
				} else {
					__antithesis_instrumentation__.Notify(173796)

					distinct = false
				}
			}
			__antithesis_instrumentation__.Notify(173788)
			continue
		} else {
			__antithesis_instrumentation__.Notify(173797)
		}
		__antithesis_instrumentation__.Notify(173775)
		r = append(r, cur)
	}
	__antithesis_instrumentation__.Notify(173769)
	return r, distinct
}

func SubtractSpans(todo, done Spans) Spans {
	__antithesis_instrumentation__.Notify(173798)
	if len(done) == 0 {
		__antithesis_instrumentation__.Notify(173803)
		return todo
	} else {
		__antithesis_instrumentation__.Notify(173804)
	}
	__antithesis_instrumentation__.Notify(173799)
	sort.Sort(todo)
	sort.Sort(done)

	remaining := make(Spans, 0, len(todo))
	appendRemaining := func(s Span) {
		__antithesis_instrumentation__.Notify(173805)
		if len(remaining) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(173806)
			return remaining[len(remaining)-1].EndKey.Equal(s.Key) == true
		}() == true {
			__antithesis_instrumentation__.Notify(173807)
			remaining[len(remaining)-1].EndKey = s.EndKey
		} else {
			__antithesis_instrumentation__.Notify(173808)
			remaining = append(remaining, s)
		}
	}
	__antithesis_instrumentation__.Notify(173800)

	var d int
	var t int
	for t < len(todo) && func() bool {
		__antithesis_instrumentation__.Notify(173809)
		return d < len(done) == true
	}() == true {
		__antithesis_instrumentation__.Notify(173810)
		tStart, tEnd := todo[t].Key, todo[t].EndKey
		dStart, dEnd := done[d].Key, done[d].EndKey
		if tStart.Equal(tEnd) {
			__antithesis_instrumentation__.Notify(173814)

			t++
			continue
		} else {
			__antithesis_instrumentation__.Notify(173815)
		}
		__antithesis_instrumentation__.Notify(173811)
		if dStart.Compare(tEnd) >= 0 {
			__antithesis_instrumentation__.Notify(173816)

			appendRemaining(todo[t])
			t++
			continue
		} else {
			__antithesis_instrumentation__.Notify(173817)
		}
		__antithesis_instrumentation__.Notify(173812)
		if dEnd.Compare(tStart) <= 0 {
			__antithesis_instrumentation__.Notify(173818)

			d++
			continue
		} else {
			__antithesis_instrumentation__.Notify(173819)
		}
		__antithesis_instrumentation__.Notify(173813)

		endCmp := dEnd.Compare(tEnd)
		if dStart.Compare(tStart) <= 0 {
			__antithesis_instrumentation__.Notify(173820)

			if endCmp < 0 {
				__antithesis_instrumentation__.Notify(173821)

				todo[t].Key = dEnd
				d++
			} else {
				__antithesis_instrumentation__.Notify(173822)
				if endCmp > 0 {
					__antithesis_instrumentation__.Notify(173823)

					t++
				} else {
					__antithesis_instrumentation__.Notify(173824)

					t++
					d++
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(173825)

			appendRemaining(Span{Key: tStart, EndKey: dStart})

			if endCmp < 0 {
				__antithesis_instrumentation__.Notify(173826)

				todo[t].Key = dEnd
				d++
			} else {
				__antithesis_instrumentation__.Notify(173827)
				if endCmp > 0 {
					__antithesis_instrumentation__.Notify(173828)

					t++
				} else {
					__antithesis_instrumentation__.Notify(173829)

					t++
					d++
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(173801)

	if t < len(todo) {
		__antithesis_instrumentation__.Notify(173830)
		remaining = append(remaining, todo[t:]...)
	} else {
		__antithesis_instrumentation__.Notify(173831)
	}
	__antithesis_instrumentation__.Notify(173802)
	return remaining
}
