package option

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
)

type NodeListOption []int

func NewNodeListOptionRange(start, end int) NodeListOption {
	__antithesis_instrumentation__.Notify(44270)
	ret := make(NodeListOption, (end-start)+1)
	for i := 0; i <= (end - start); i++ {
		__antithesis_instrumentation__.Notify(44272)
		ret[i] = start + i
	}
	__antithesis_instrumentation__.Notify(44271)
	return ret
}

func (n NodeListOption) Equals(o NodeListOption) bool {
	__antithesis_instrumentation__.Notify(44273)
	if len(n) != len(o) {
		__antithesis_instrumentation__.Notify(44276)
		return false
	} else {
		__antithesis_instrumentation__.Notify(44277)
	}
	__antithesis_instrumentation__.Notify(44274)
	for i := range n {
		__antithesis_instrumentation__.Notify(44278)
		if n[i] != o[i] {
			__antithesis_instrumentation__.Notify(44279)
			return false
		} else {
			__antithesis_instrumentation__.Notify(44280)
		}
	}
	__antithesis_instrumentation__.Notify(44275)
	return true
}

func (n NodeListOption) Option() { __antithesis_instrumentation__.Notify(44281) }

func (n NodeListOption) Merge(o NodeListOption) NodeListOption {
	__antithesis_instrumentation__.Notify(44282)
	t := make(NodeListOption, 0, len(n)+len(o))
	t = append(t, n...)
	t = append(t, o...)
	sort.Ints(t)
	r := t[:1]
	for i := 1; i < len(t); i++ {
		__antithesis_instrumentation__.Notify(44284)
		if r[len(r)-1] != t[i] {
			__antithesis_instrumentation__.Notify(44285)
			r = append(r, t[i])
		} else {
			__antithesis_instrumentation__.Notify(44286)
		}
	}
	__antithesis_instrumentation__.Notify(44283)
	return r
}

func (n NodeListOption) RandNode() NodeListOption {
	__antithesis_instrumentation__.Notify(44287)
	return NodeListOption{n[rand.Intn(len(n))]}
}

func (n NodeListOption) NodeIDsString() string {
	__antithesis_instrumentation__.Notify(44288)
	result := ""
	for _, i := range n {
		__antithesis_instrumentation__.Notify(44290)
		result += fmt.Sprintf("%s ", strconv.Itoa(i))
	}
	__antithesis_instrumentation__.Notify(44289)
	return result
}

func (n NodeListOption) String() string {
	__antithesis_instrumentation__.Notify(44291)
	if len(n) == 0 {
		__antithesis_instrumentation__.Notify(44296)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(44297)
	}
	__antithesis_instrumentation__.Notify(44292)

	var buf bytes.Buffer
	buf.WriteByte(':')

	appendRange := func(start, end int) {
		__antithesis_instrumentation__.Notify(44298)
		if buf.Len() > 1 {
			__antithesis_instrumentation__.Notify(44300)
			buf.WriteByte(',')
		} else {
			__antithesis_instrumentation__.Notify(44301)
		}
		__antithesis_instrumentation__.Notify(44299)
		if start == end {
			__antithesis_instrumentation__.Notify(44302)
			fmt.Fprintf(&buf, "%d", start)
		} else {
			__antithesis_instrumentation__.Notify(44303)
			fmt.Fprintf(&buf, "%d-%d", start, end)
		}
	}
	__antithesis_instrumentation__.Notify(44293)

	start, end := -1, -1
	for _, i := range n {
		__antithesis_instrumentation__.Notify(44304)
		if start != -1 && func() bool {
			__antithesis_instrumentation__.Notify(44307)
			return end == i-1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(44308)
			end = i
			continue
		} else {
			__antithesis_instrumentation__.Notify(44309)
		}
		__antithesis_instrumentation__.Notify(44305)
		if start != -1 {
			__antithesis_instrumentation__.Notify(44310)
			appendRange(start, end)
		} else {
			__antithesis_instrumentation__.Notify(44311)
		}
		__antithesis_instrumentation__.Notify(44306)
		start, end = i, i
	}
	__antithesis_instrumentation__.Notify(44294)
	if start != -1 {
		__antithesis_instrumentation__.Notify(44312)
		appendRange(start, end)
	} else {
		__antithesis_instrumentation__.Notify(44313)
	}
	__antithesis_instrumentation__.Notify(44295)
	return buf.String()
}
