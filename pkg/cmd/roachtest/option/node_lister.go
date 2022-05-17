package option

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type NodeLister struct {
	NodeCount int
	Fatalf    func(string, ...interface{})
}

func (l NodeLister) All() NodeListOption {
	__antithesis_instrumentation__.Notify(44314)
	return l.Range(1, l.NodeCount)
}

func (l NodeLister) Range(begin, end int) NodeListOption {
	__antithesis_instrumentation__.Notify(44315)
	if begin < 1 || func() bool {
		__antithesis_instrumentation__.Notify(44318)
		return end > l.NodeCount == true
	}() == true {
		__antithesis_instrumentation__.Notify(44319)
		l.Fatalf("invalid node range: %d-%d (1-%d)", begin, end, l.NodeCount)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(44320)
	}
	__antithesis_instrumentation__.Notify(44316)
	r := make(NodeListOption, 0, 1+end-begin)
	for i := begin; i <= end; i++ {
		__antithesis_instrumentation__.Notify(44321)
		r = append(r, i)
	}
	__antithesis_instrumentation__.Notify(44317)
	return r
}

func (l NodeLister) Nodes(ns ...int) NodeListOption {
	__antithesis_instrumentation__.Notify(44322)
	r := make(NodeListOption, 0, len(ns))
	for _, n := range ns {
		__antithesis_instrumentation__.Notify(44324)
		if n < 1 || func() bool {
			__antithesis_instrumentation__.Notify(44326)
			return n > l.NodeCount == true
		}() == true {
			__antithesis_instrumentation__.Notify(44327)
			l.Fatalf("invalid node range: %d (1-%d)", n, l.NodeCount)
		} else {
			__antithesis_instrumentation__.Notify(44328)
		}
		__antithesis_instrumentation__.Notify(44325)

		r = append(r, n)
	}
	__antithesis_instrumentation__.Notify(44323)
	return r
}

func (l NodeLister) Node(n int) NodeListOption {
	__antithesis_instrumentation__.Notify(44329)
	return l.Nodes(n)
}
