package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treewindow"
	"github.com/cockroachdb/cockroach/pkg/util/ring"
)

type PeerGroupChecker interface {
	InSameGroup(i, j int) (bool, error)
}

type peerGroup struct {
	firstPeerIdx int
	rowCount     int
}

type PeerGroupsIndicesHelper struct {
	groups               ring.Buffer
	peerGrouper          PeerGroupChecker
	headPeerGroupNum     int
	allPeerGroupsSkipped bool
	allRowsProcessed     bool
	unboundedFollowing   int
}

func (p *PeerGroupsIndicesHelper) Init(wfr *WindowFrameRun, peerGrouper PeerGroupChecker) error {
	__antithesis_instrumentation__.Notify(617057)

	p.groups.Reset()
	p.headPeerGroupNum = 0
	p.allPeerGroupsSkipped = false
	p.allRowsProcessed = false
	p.unboundedFollowing = wfr.unboundedFollowing()

	var group *peerGroup
	p.peerGrouper = peerGrouper
	startIdxOfFirstPeerGroupWithinFrame := 0
	if wfr.Frame != nil && func() bool {
		__antithesis_instrumentation__.Notify(617062)
		return wfr.Frame.Mode == treewindow.GROUPS == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(617063)
		return wfr.Frame.Bounds.StartBound.BoundType == treewindow.OffsetFollowing == true
	}() == true {
		__antithesis_instrumentation__.Notify(617064)

		peerGroupOffset := int(MustBeDInt(wfr.StartBoundOffset))
		group = &peerGroup{firstPeerIdx: 0, rowCount: 1}
		for group.firstPeerIdx < wfr.PartitionSize() && func() bool {
			__antithesis_instrumentation__.Notify(617067)
			return p.groups.Len() < peerGroupOffset == true
		}() == true {
			__antithesis_instrumentation__.Notify(617068)
			p.groups.AddLast(group)
			for ; group.firstPeerIdx+group.rowCount < wfr.PartitionSize(); group.rowCount++ {
				__antithesis_instrumentation__.Notify(617070)
				idx := group.firstPeerIdx + group.rowCount
				if sameGroup, err := p.peerGrouper.InSameGroup(idx-1, idx); err != nil {
					__antithesis_instrumentation__.Notify(617071)
					return err
				} else {
					__antithesis_instrumentation__.Notify(617072)
					if !sameGroup {
						__antithesis_instrumentation__.Notify(617073)
						break
					} else {
						__antithesis_instrumentation__.Notify(617074)
					}
				}
			}
			__antithesis_instrumentation__.Notify(617069)
			group = &peerGroup{firstPeerIdx: group.firstPeerIdx + group.rowCount, rowCount: 1}
		}
		__antithesis_instrumentation__.Notify(617065)

		if group.firstPeerIdx == wfr.PartitionSize() {
			__antithesis_instrumentation__.Notify(617075)

			p.allPeerGroupsSkipped = true
			return nil
		} else {
			__antithesis_instrumentation__.Notify(617076)
		}
		__antithesis_instrumentation__.Notify(617066)

		startIdxOfFirstPeerGroupWithinFrame = group.firstPeerIdx
	} else {
		__antithesis_instrumentation__.Notify(617077)
	}
	__antithesis_instrumentation__.Notify(617058)

	group = &peerGroup{firstPeerIdx: startIdxOfFirstPeerGroupWithinFrame, rowCount: 1}
	p.groups.AddLast(group)
	for ; group.firstPeerIdx+group.rowCount < wfr.PartitionSize(); group.rowCount++ {
		__antithesis_instrumentation__.Notify(617078)
		idx := group.firstPeerIdx + group.rowCount
		if sameGroup, err := p.peerGrouper.InSameGroup(idx-1, idx); err != nil {
			__antithesis_instrumentation__.Notify(617079)
			return err
		} else {
			__antithesis_instrumentation__.Notify(617080)
			if !sameGroup {
				__antithesis_instrumentation__.Notify(617081)
				break
			} else {
				__antithesis_instrumentation__.Notify(617082)
			}
		}
	}
	__antithesis_instrumentation__.Notify(617059)
	if group.firstPeerIdx+group.rowCount == wfr.PartitionSize() {
		__antithesis_instrumentation__.Notify(617083)
		p.allRowsProcessed = true
		return nil
	} else {
		__antithesis_instrumentation__.Notify(617084)
	}
	__antithesis_instrumentation__.Notify(617060)

	if wfr.Frame != nil && func() bool {
		__antithesis_instrumentation__.Notify(617085)
		return wfr.Frame.Mode == treewindow.GROUPS == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(617086)
		return wfr.Frame.Bounds.EndBound != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(617087)
		return wfr.Frame.Bounds.EndBound.BoundType == treewindow.OffsetFollowing == true
	}() == true {
		__antithesis_instrumentation__.Notify(617088)

		peerGroupOffset := int(MustBeDInt(wfr.EndBoundOffset))
		group = &peerGroup{firstPeerIdx: group.firstPeerIdx + group.rowCount, rowCount: 1}
		for group.firstPeerIdx < wfr.PartitionSize() && func() bool {
			__antithesis_instrumentation__.Notify(617090)
			return p.groups.Len() <= peerGroupOffset == true
		}() == true {
			__antithesis_instrumentation__.Notify(617091)
			p.groups.AddLast(group)
			for ; group.firstPeerIdx+group.rowCount < wfr.PartitionSize(); group.rowCount++ {
				__antithesis_instrumentation__.Notify(617093)
				idx := group.firstPeerIdx + group.rowCount
				if sameGroup, err := p.peerGrouper.InSameGroup(idx-1, idx); err != nil {
					__antithesis_instrumentation__.Notify(617094)
					return err
				} else {
					__antithesis_instrumentation__.Notify(617095)
					if !sameGroup {
						__antithesis_instrumentation__.Notify(617096)
						break
					} else {
						__antithesis_instrumentation__.Notify(617097)
					}
				}
			}
			__antithesis_instrumentation__.Notify(617092)
			group = &peerGroup{firstPeerIdx: group.firstPeerIdx + group.rowCount, rowCount: 1}
		}
		__antithesis_instrumentation__.Notify(617089)
		if group.firstPeerIdx == wfr.PartitionSize() {
			__antithesis_instrumentation__.Notify(617098)
			p.allRowsProcessed = true
		} else {
			__antithesis_instrumentation__.Notify(617099)
		}
	} else {
		__antithesis_instrumentation__.Notify(617100)
	}
	__antithesis_instrumentation__.Notify(617061)
	return nil
}

func (p *PeerGroupsIndicesHelper) Update(wfr *WindowFrameRun) error {
	__antithesis_instrumentation__.Notify(617101)
	if p.allPeerGroupsSkipped {
		__antithesis_instrumentation__.Notify(617107)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(617108)
	}
	__antithesis_instrumentation__.Notify(617102)

	lastPeerGroup := p.groups.GetLast().(*peerGroup)
	nextPeerGroupStartIdx := lastPeerGroup.firstPeerIdx + lastPeerGroup.rowCount

	if (wfr.Frame == nil || func() bool {
		__antithesis_instrumentation__.Notify(617109)
		return wfr.Frame.Mode == treewindow.ROWS == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(617110)
		return wfr.Frame.Mode == treewindow.RANGE == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(617111)
		return (wfr.Frame.Bounds.StartBound.BoundType == treewindow.OffsetPreceding && func() bool {
			__antithesis_instrumentation__.Notify(617112)
			return wfr.CurRowPeerGroupNum-p.headPeerGroupNum > int(MustBeDInt(wfr.StartBoundOffset)) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(617113)
			return wfr.Frame.Bounds.StartBound.BoundType == treewindow.CurrentRow == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(617114)
			return (wfr.Frame.Bounds.StartBound.BoundType == treewindow.OffsetFollowing && func() bool {
				__antithesis_instrumentation__.Notify(617115)
				return p.headPeerGroupNum-wfr.CurRowPeerGroupNum > int(MustBeDInt(wfr.StartBoundOffset)) == true
			}() == true) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(617116)

		p.groups.RemoveFirst()
		p.headPeerGroupNum++
	} else {
		__antithesis_instrumentation__.Notify(617117)
	}
	__antithesis_instrumentation__.Notify(617103)

	if p.allRowsProcessed {
		__antithesis_instrumentation__.Notify(617118)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(617119)
	}
	__antithesis_instrumentation__.Notify(617104)

	peerGroup := &peerGroup{firstPeerIdx: nextPeerGroupStartIdx, rowCount: 1}
	p.groups.AddLast(peerGroup)
	for ; peerGroup.firstPeerIdx+peerGroup.rowCount < wfr.PartitionSize(); peerGroup.rowCount++ {
		__antithesis_instrumentation__.Notify(617120)
		idx := peerGroup.firstPeerIdx + peerGroup.rowCount
		if sameGroup, err := p.peerGrouper.InSameGroup(idx-1, idx); err != nil {
			__antithesis_instrumentation__.Notify(617121)
			return err
		} else {
			__antithesis_instrumentation__.Notify(617122)
			if !sameGroup {
				__antithesis_instrumentation__.Notify(617123)
				break
			} else {
				__antithesis_instrumentation__.Notify(617124)
			}
		}
	}
	__antithesis_instrumentation__.Notify(617105)
	if peerGroup.firstPeerIdx+peerGroup.rowCount == wfr.PartitionSize() {
		__antithesis_instrumentation__.Notify(617125)
		p.allRowsProcessed = true
	} else {
		__antithesis_instrumentation__.Notify(617126)
	}
	__antithesis_instrumentation__.Notify(617106)
	return nil
}

func (p *PeerGroupsIndicesHelper) GetFirstPeerIdx(peerGroupNum int) int {
	__antithesis_instrumentation__.Notify(617127)
	posInBuffer := peerGroupNum - p.headPeerGroupNum
	if posInBuffer < 0 || func() bool {
		__antithesis_instrumentation__.Notify(617129)
		return p.groups.Len() < posInBuffer == true
	}() == true {
		__antithesis_instrumentation__.Notify(617130)
		panic("peerGroupNum out of bounds")
	} else {
		__antithesis_instrumentation__.Notify(617131)
	}
	__antithesis_instrumentation__.Notify(617128)
	return p.groups.Get(posInBuffer).(*peerGroup).firstPeerIdx
}

func (p *PeerGroupsIndicesHelper) GetRowCount(peerGroupNum int) int {
	__antithesis_instrumentation__.Notify(617132)
	posInBuffer := peerGroupNum - p.headPeerGroupNum
	if posInBuffer < 0 || func() bool {
		__antithesis_instrumentation__.Notify(617134)
		return p.groups.Len() < posInBuffer == true
	}() == true {
		__antithesis_instrumentation__.Notify(617135)
		panic("peerGroupNum out of bounds")
	} else {
		__antithesis_instrumentation__.Notify(617136)
	}
	__antithesis_instrumentation__.Notify(617133)
	return p.groups.Get(posInBuffer).(*peerGroup).rowCount
}

func (p *PeerGroupsIndicesHelper) GetLastPeerGroupNum() int {
	__antithesis_instrumentation__.Notify(617137)
	if p.groups.Len() == 0 {
		__antithesis_instrumentation__.Notify(617139)
		panic("GetLastPeerGroupNum on empty RingBuffer")
	} else {
		__antithesis_instrumentation__.Notify(617140)
	}
	__antithesis_instrumentation__.Notify(617138)
	return p.headPeerGroupNum + p.groups.Len() - 1
}
