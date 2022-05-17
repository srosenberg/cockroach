package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treebin"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treewindow"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type IndexedRows interface {
	Len() int
	GetRow(ctx context.Context, idx int) (IndexedRow, error)
}

type IndexedRow interface {
	GetIdx() int
	GetDatum(idx int) (Datum, error)
	GetDatums(startIdx, endIdx int) (Datums, error)
}

type WindowFrameRun struct {
	Rows             IndexedRows
	ArgsIdxs         []uint32
	Frame            *WindowFrame
	StartBoundOffset Datum
	EndBoundOffset   Datum
	FilterColIdx     int
	OrdColIdx        int
	OrdDirection     encoding.Direction
	PlusOp, MinusOp  *BinOp
	PeerHelper       PeerGroupsIndicesHelper

	err error

	CurRowPeerGroupNum int

	RowIdx int
}

type WindowFrameRangeOps struct{}

func (o WindowFrameRangeOps) LookupImpl(left, right *types.T) (*BinOp, *BinOp, bool) {
	__antithesis_instrumentation__.Notify(616740)
	plusOverloads, minusOverloads := BinOps[treebin.Plus], BinOps[treebin.Minus]
	plusOp, found := plusOverloads.lookupImpl(left, right)
	if !found {
		__antithesis_instrumentation__.Notify(616743)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(616744)
	}
	__antithesis_instrumentation__.Notify(616741)
	minusOp, found := minusOverloads.lookupImpl(left, right)
	if !found {
		__antithesis_instrumentation__.Notify(616745)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(616746)
	}
	__antithesis_instrumentation__.Notify(616742)
	return plusOp, minusOp, true
}

func (wfr *WindowFrameRun) getValueByOffset(
	ctx context.Context, evalCtx *EvalContext, offset Datum, negative bool,
) (Datum, error) {
	__antithesis_instrumentation__.Notify(616747)
	if wfr.OrdDirection == encoding.Descending {
		__antithesis_instrumentation__.Notify(616752)

		negative = !negative
	} else {
		__antithesis_instrumentation__.Notify(616753)
	}
	__antithesis_instrumentation__.Notify(616748)
	var binOp *BinOp
	if negative {
		__antithesis_instrumentation__.Notify(616754)
		binOp = wfr.MinusOp
	} else {
		__antithesis_instrumentation__.Notify(616755)
		binOp = wfr.PlusOp
	}
	__antithesis_instrumentation__.Notify(616749)
	value, err := wfr.valueAt(ctx, wfr.RowIdx)
	if err != nil {
		__antithesis_instrumentation__.Notify(616756)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(616757)
	}
	__antithesis_instrumentation__.Notify(616750)
	if value == DNull {
		__antithesis_instrumentation__.Notify(616758)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(616759)
	}
	__antithesis_instrumentation__.Notify(616751)
	return binOp.Fn(evalCtx, value, offset)
}

func (wfr *WindowFrameRun) FrameStartIdx(ctx context.Context, evalCtx *EvalContext) (int, error) {
	__antithesis_instrumentation__.Notify(616760)
	if wfr.Frame == nil {
		__antithesis_instrumentation__.Notify(616762)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(616763)
	}
	__antithesis_instrumentation__.Notify(616761)
	switch wfr.Frame.Mode {
	case treewindow.RANGE:
		__antithesis_instrumentation__.Notify(616764)
		switch wfr.Frame.Bounds.StartBound.BoundType {
		case treewindow.UnboundedPreceding:
			__antithesis_instrumentation__.Notify(616768)
			return 0, nil
		case treewindow.OffsetPreceding:
			__antithesis_instrumentation__.Notify(616769)
			value, err := wfr.getValueByOffset(ctx, evalCtx, wfr.StartBoundOffset, true)
			if err != nil {
				__antithesis_instrumentation__.Notify(616777)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(616778)
			}
			__antithesis_instrumentation__.Notify(616770)
			if wfr.OrdDirection == encoding.Descending {
				__antithesis_instrumentation__.Notify(616779)

				return sort.Search(wfr.RowIdx, func(i int) bool {
					__antithesis_instrumentation__.Notify(616780)
					if wfr.err != nil {
						__antithesis_instrumentation__.Notify(616784)
						return false
					} else {
						__antithesis_instrumentation__.Notify(616785)
					}
					__antithesis_instrumentation__.Notify(616781)
					valueAt, err := wfr.valueAt(ctx, i)
					if err != nil {
						__antithesis_instrumentation__.Notify(616786)
						wfr.err = err
						return false
					} else {
						__antithesis_instrumentation__.Notify(616787)
					}
					__antithesis_instrumentation__.Notify(616782)
					cmp, err := compareForWindow(evalCtx, valueAt, value)
					if err != nil {
						__antithesis_instrumentation__.Notify(616788)
						wfr.err = err
						return false
					} else {
						__antithesis_instrumentation__.Notify(616789)
					}
					__antithesis_instrumentation__.Notify(616783)
					return cmp <= 0
				}), wfr.err
			} else {
				__antithesis_instrumentation__.Notify(616790)
			}
			__antithesis_instrumentation__.Notify(616771)

			return sort.Search(wfr.RowIdx, func(i int) bool {
				__antithesis_instrumentation__.Notify(616791)
				if wfr.err != nil {
					__antithesis_instrumentation__.Notify(616795)
					return false
				} else {
					__antithesis_instrumentation__.Notify(616796)
				}
				__antithesis_instrumentation__.Notify(616792)
				valueAt, err := wfr.valueAt(ctx, i)
				if err != nil {
					__antithesis_instrumentation__.Notify(616797)
					wfr.err = err
					return false
				} else {
					__antithesis_instrumentation__.Notify(616798)
				}
				__antithesis_instrumentation__.Notify(616793)
				cmp, err := compareForWindow(evalCtx, valueAt, value)
				if err != nil {
					__antithesis_instrumentation__.Notify(616799)
					wfr.err = err
					return false
				} else {
					__antithesis_instrumentation__.Notify(616800)
				}
				__antithesis_instrumentation__.Notify(616794)
				return cmp >= 0
			}), wfr.err
		case treewindow.CurrentRow:
			__antithesis_instrumentation__.Notify(616772)

			return wfr.PeerHelper.GetFirstPeerIdx(wfr.CurRowPeerGroupNum), nil
		case treewindow.OffsetFollowing:
			__antithesis_instrumentation__.Notify(616773)
			value, err := wfr.getValueByOffset(ctx, evalCtx, wfr.StartBoundOffset, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(616801)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(616802)
			}
			__antithesis_instrumentation__.Notify(616774)
			if wfr.OrdDirection == encoding.Descending {
				__antithesis_instrumentation__.Notify(616803)

				return sort.Search(wfr.PartitionSize(), func(i int) bool {
					__antithesis_instrumentation__.Notify(616804)
					if wfr.err != nil {
						__antithesis_instrumentation__.Notify(616808)
						return false
					} else {
						__antithesis_instrumentation__.Notify(616809)
					}
					__antithesis_instrumentation__.Notify(616805)
					valueAt, err := wfr.valueAt(ctx, i)
					if err != nil {
						__antithesis_instrumentation__.Notify(616810)
						wfr.err = err
						return false
					} else {
						__antithesis_instrumentation__.Notify(616811)
					}
					__antithesis_instrumentation__.Notify(616806)
					cmp, err := compareForWindow(evalCtx, valueAt, value)
					if err != nil {
						__antithesis_instrumentation__.Notify(616812)
						wfr.err = err
						return false
					} else {
						__antithesis_instrumentation__.Notify(616813)
					}
					__antithesis_instrumentation__.Notify(616807)
					return cmp <= 0
				}), wfr.err
			} else {
				__antithesis_instrumentation__.Notify(616814)
			}
			__antithesis_instrumentation__.Notify(616775)

			return sort.Search(wfr.PartitionSize(), func(i int) bool {
				__antithesis_instrumentation__.Notify(616815)
				if wfr.err != nil {
					__antithesis_instrumentation__.Notify(616819)
					return false
				} else {
					__antithesis_instrumentation__.Notify(616820)
				}
				__antithesis_instrumentation__.Notify(616816)
				valueAt, err := wfr.valueAt(ctx, i)
				if err != nil {
					__antithesis_instrumentation__.Notify(616821)
					wfr.err = err
					return false
				} else {
					__antithesis_instrumentation__.Notify(616822)
				}
				__antithesis_instrumentation__.Notify(616817)
				cmp, err := compareForWindow(evalCtx, valueAt, value)
				if err != nil {
					__antithesis_instrumentation__.Notify(616823)
					wfr.err = err
					return false
				} else {
					__antithesis_instrumentation__.Notify(616824)
				}
				__antithesis_instrumentation__.Notify(616818)
				return cmp >= 0
			}), wfr.err
		default:
			__antithesis_instrumentation__.Notify(616776)
			return 0, errors.AssertionFailedf(
				"unexpected WindowFrameBoundType in RANGE mode: %d",
				redact.Safe(wfr.Frame.Bounds.StartBound.BoundType))
		}
	case treewindow.ROWS:
		__antithesis_instrumentation__.Notify(616765)
		switch wfr.Frame.Bounds.StartBound.BoundType {
		case treewindow.UnboundedPreceding:
			__antithesis_instrumentation__.Notify(616825)
			return 0, nil
		case treewindow.OffsetPreceding:
			__antithesis_instrumentation__.Notify(616826)
			offset := MustBeDInt(wfr.StartBoundOffset)
			idx := wfr.RowIdx - int(offset)
			if idx < 0 {
				__antithesis_instrumentation__.Notify(616832)
				idx = 0
			} else {
				__antithesis_instrumentation__.Notify(616833)
			}
			__antithesis_instrumentation__.Notify(616827)
			return idx, nil
		case treewindow.CurrentRow:
			__antithesis_instrumentation__.Notify(616828)
			return wfr.RowIdx, nil
		case treewindow.OffsetFollowing:
			__antithesis_instrumentation__.Notify(616829)
			offset := MustBeDInt(wfr.StartBoundOffset)
			idx := wfr.RowIdx + int(offset)
			if idx >= wfr.PartitionSize() || func() bool {
				__antithesis_instrumentation__.Notify(616834)
				return int(offset) >= wfr.PartitionSize() == true
			}() == true {
				__antithesis_instrumentation__.Notify(616835)

				idx = wfr.unboundedFollowing()
			} else {
				__antithesis_instrumentation__.Notify(616836)
			}
			__antithesis_instrumentation__.Notify(616830)
			return idx, nil
		default:
			__antithesis_instrumentation__.Notify(616831)
			return 0, errors.AssertionFailedf(
				"unexpected WindowFrameBoundType in ROWS mode: %d",
				redact.Safe(wfr.Frame.Bounds.StartBound.BoundType))
		}
	case treewindow.GROUPS:
		__antithesis_instrumentation__.Notify(616766)
		switch wfr.Frame.Bounds.StartBound.BoundType {
		case treewindow.UnboundedPreceding:
			__antithesis_instrumentation__.Notify(616837)
			return 0, nil
		case treewindow.OffsetPreceding:
			__antithesis_instrumentation__.Notify(616838)
			offset := MustBeDInt(wfr.StartBoundOffset)
			peerGroupNum := wfr.CurRowPeerGroupNum - int(offset)
			if peerGroupNum < 0 {
				__antithesis_instrumentation__.Notify(616844)
				peerGroupNum = 0
			} else {
				__antithesis_instrumentation__.Notify(616845)
			}
			__antithesis_instrumentation__.Notify(616839)
			return wfr.PeerHelper.GetFirstPeerIdx(peerGroupNum), nil
		case treewindow.CurrentRow:
			__antithesis_instrumentation__.Notify(616840)

			return wfr.PeerHelper.GetFirstPeerIdx(wfr.CurRowPeerGroupNum), nil
		case treewindow.OffsetFollowing:
			__antithesis_instrumentation__.Notify(616841)
			offset := MustBeDInt(wfr.StartBoundOffset)
			peerGroupNum := wfr.CurRowPeerGroupNum + int(offset)
			lastPeerGroupNum := wfr.PeerHelper.GetLastPeerGroupNum()
			if peerGroupNum > lastPeerGroupNum || func() bool {
				__antithesis_instrumentation__.Notify(616846)
				return peerGroupNum < 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(616847)

				return wfr.unboundedFollowing(), nil
			} else {
				__antithesis_instrumentation__.Notify(616848)
			}
			__antithesis_instrumentation__.Notify(616842)
			return wfr.PeerHelper.GetFirstPeerIdx(peerGroupNum), nil
		default:
			__antithesis_instrumentation__.Notify(616843)
			return 0, errors.AssertionFailedf(
				"unexpected WindowFrameBoundType in GROUPS mode: %d",
				redact.Safe(wfr.Frame.Bounds.StartBound.BoundType))
		}
	default:
		__antithesis_instrumentation__.Notify(616767)
		return 0, errors.AssertionFailedf("unexpected WindowFrameMode: %d", wfr.Frame.Mode)
	}
}

func (f *WindowFrame) IsDefaultFrame() bool {
	__antithesis_instrumentation__.Notify(616849)
	if f == nil {
		__antithesis_instrumentation__.Notify(616852)
		return true
	} else {
		__antithesis_instrumentation__.Notify(616853)
	}
	__antithesis_instrumentation__.Notify(616850)
	if f.Bounds.StartBound.BoundType == treewindow.UnboundedPreceding {
		__antithesis_instrumentation__.Notify(616854)
		return f.DefaultFrameExclusion() && func() bool {
			__antithesis_instrumentation__.Notify(616855)
			return f.Mode == treewindow.RANGE == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(616856)
			return (f.Bounds.EndBound == nil || func() bool {
				__antithesis_instrumentation__.Notify(616857)
				return f.Bounds.EndBound.BoundType == treewindow.CurrentRow == true
			}() == true) == true
		}() == true
	} else {
		__antithesis_instrumentation__.Notify(616858)
	}
	__antithesis_instrumentation__.Notify(616851)
	return false
}

func (f *WindowFrame) DefaultFrameExclusion() bool {
	__antithesis_instrumentation__.Notify(616859)
	return f == nil || func() bool {
		__antithesis_instrumentation__.Notify(616860)
		return f.Exclusion == treewindow.NoExclusion == true
	}() == true
}

func (wfr *WindowFrameRun) FrameEndIdx(ctx context.Context, evalCtx *EvalContext) (int, error) {
	__antithesis_instrumentation__.Notify(616861)
	if wfr.Frame == nil {
		__antithesis_instrumentation__.Notify(616863)
		return wfr.DefaultFrameSize(), nil
	} else {
		__antithesis_instrumentation__.Notify(616864)
	}
	__antithesis_instrumentation__.Notify(616862)
	switch wfr.Frame.Mode {
	case treewindow.RANGE:
		__antithesis_instrumentation__.Notify(616865)
		if wfr.Frame.Bounds.EndBound == nil {
			__antithesis_instrumentation__.Notify(616872)

			return wfr.DefaultFrameSize(), nil
		} else {
			__antithesis_instrumentation__.Notify(616873)
		}
		__antithesis_instrumentation__.Notify(616866)
		switch wfr.Frame.Bounds.EndBound.BoundType {
		case treewindow.OffsetPreceding:
			__antithesis_instrumentation__.Notify(616874)
			value, err := wfr.getValueByOffset(ctx, evalCtx, wfr.EndBoundOffset, true)
			if err != nil {
				__antithesis_instrumentation__.Notify(616883)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(616884)
			}
			__antithesis_instrumentation__.Notify(616875)
			if wfr.OrdDirection == encoding.Descending {
				__antithesis_instrumentation__.Notify(616885)

				return sort.Search(wfr.PartitionSize(), func(i int) bool {
					__antithesis_instrumentation__.Notify(616886)
					if wfr.err != nil {
						__antithesis_instrumentation__.Notify(616890)
						return false
					} else {
						__antithesis_instrumentation__.Notify(616891)
					}
					__antithesis_instrumentation__.Notify(616887)
					valueAt, err := wfr.valueAt(ctx, i)
					if err != nil {
						__antithesis_instrumentation__.Notify(616892)
						wfr.err = err
						return false
					} else {
						__antithesis_instrumentation__.Notify(616893)
					}
					__antithesis_instrumentation__.Notify(616888)
					cmp, err := compareForWindow(evalCtx, valueAt, value)
					if err != nil {
						__antithesis_instrumentation__.Notify(616894)
						wfr.err = err
						return false
					} else {
						__antithesis_instrumentation__.Notify(616895)
					}
					__antithesis_instrumentation__.Notify(616889)
					return cmp < 0
				}), wfr.err
			} else {
				__antithesis_instrumentation__.Notify(616896)
			}
			__antithesis_instrumentation__.Notify(616876)

			return sort.Search(wfr.PartitionSize(), func(i int) bool {
				__antithesis_instrumentation__.Notify(616897)
				if wfr.err != nil {
					__antithesis_instrumentation__.Notify(616901)
					return false
				} else {
					__antithesis_instrumentation__.Notify(616902)
				}
				__antithesis_instrumentation__.Notify(616898)
				valueAt, err := wfr.valueAt(ctx, i)
				if err != nil {
					__antithesis_instrumentation__.Notify(616903)
					wfr.err = err
					return false
				} else {
					__antithesis_instrumentation__.Notify(616904)
				}
				__antithesis_instrumentation__.Notify(616899)
				cmp, err := compareForWindow(evalCtx, valueAt, value)
				if err != nil {
					__antithesis_instrumentation__.Notify(616905)
					wfr.err = err
					return false
				} else {
					__antithesis_instrumentation__.Notify(616906)
				}
				__antithesis_instrumentation__.Notify(616900)
				return cmp > 0
			}), wfr.err
		case treewindow.CurrentRow:
			__antithesis_instrumentation__.Notify(616877)

			return wfr.DefaultFrameSize(), nil
		case treewindow.OffsetFollowing:
			__antithesis_instrumentation__.Notify(616878)
			value, err := wfr.getValueByOffset(ctx, evalCtx, wfr.EndBoundOffset, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(616907)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(616908)
			}
			__antithesis_instrumentation__.Notify(616879)
			if wfr.OrdDirection == encoding.Descending {
				__antithesis_instrumentation__.Notify(616909)

				return sort.Search(wfr.PartitionSize(), func(i int) bool {
					__antithesis_instrumentation__.Notify(616910)
					if wfr.err != nil {
						__antithesis_instrumentation__.Notify(616914)
						return false
					} else {
						__antithesis_instrumentation__.Notify(616915)
					}
					__antithesis_instrumentation__.Notify(616911)
					valueAt, err := wfr.valueAt(ctx, i)
					if err != nil {
						__antithesis_instrumentation__.Notify(616916)
						wfr.err = err
						return false
					} else {
						__antithesis_instrumentation__.Notify(616917)
					}
					__antithesis_instrumentation__.Notify(616912)
					cmp, err := compareForWindow(evalCtx, valueAt, value)
					if err != nil {
						__antithesis_instrumentation__.Notify(616918)
						wfr.err = err
						return false
					} else {
						__antithesis_instrumentation__.Notify(616919)
					}
					__antithesis_instrumentation__.Notify(616913)
					return cmp < 0
				}), wfr.err
			} else {
				__antithesis_instrumentation__.Notify(616920)
			}
			__antithesis_instrumentation__.Notify(616880)

			return sort.Search(wfr.PartitionSize(), func(i int) bool {
				__antithesis_instrumentation__.Notify(616921)
				if wfr.err != nil {
					__antithesis_instrumentation__.Notify(616925)
					return false
				} else {
					__antithesis_instrumentation__.Notify(616926)
				}
				__antithesis_instrumentation__.Notify(616922)
				valueAt, err := wfr.valueAt(ctx, i)
				if err != nil {
					__antithesis_instrumentation__.Notify(616927)
					wfr.err = err
					return false
				} else {
					__antithesis_instrumentation__.Notify(616928)
				}
				__antithesis_instrumentation__.Notify(616923)
				cmp, err := compareForWindow(evalCtx, valueAt, value)
				if err != nil {
					__antithesis_instrumentation__.Notify(616929)
					wfr.err = err
					return false
				} else {
					__antithesis_instrumentation__.Notify(616930)
				}
				__antithesis_instrumentation__.Notify(616924)
				return cmp > 0
			}), wfr.err
		case treewindow.UnboundedFollowing:
			__antithesis_instrumentation__.Notify(616881)
			return wfr.unboundedFollowing(), nil
		default:
			__antithesis_instrumentation__.Notify(616882)
			return 0, errors.AssertionFailedf(
				"unexpected WindowFrameBoundType in RANGE mode: %d",
				redact.Safe(wfr.Frame.Bounds.EndBound.BoundType))
		}
	case treewindow.ROWS:
		__antithesis_instrumentation__.Notify(616867)
		if wfr.Frame.Bounds.EndBound == nil {
			__antithesis_instrumentation__.Notify(616931)

			return wfr.RowIdx + 1, nil
		} else {
			__antithesis_instrumentation__.Notify(616932)
		}
		__antithesis_instrumentation__.Notify(616868)
		switch wfr.Frame.Bounds.EndBound.BoundType {
		case treewindow.OffsetPreceding:
			__antithesis_instrumentation__.Notify(616933)
			offset := MustBeDInt(wfr.EndBoundOffset)
			idx := wfr.RowIdx - int(offset) + 1
			if idx < 0 {
				__antithesis_instrumentation__.Notify(616940)
				idx = 0
			} else {
				__antithesis_instrumentation__.Notify(616941)
			}
			__antithesis_instrumentation__.Notify(616934)
			return idx, nil
		case treewindow.CurrentRow:
			__antithesis_instrumentation__.Notify(616935)
			return wfr.RowIdx + 1, nil
		case treewindow.OffsetFollowing:
			__antithesis_instrumentation__.Notify(616936)
			offset := MustBeDInt(wfr.EndBoundOffset)
			idx := wfr.RowIdx + int(offset) + 1
			if idx >= wfr.PartitionSize() || func() bool {
				__antithesis_instrumentation__.Notify(616942)
				return int(offset) >= wfr.PartitionSize() == true
			}() == true {
				__antithesis_instrumentation__.Notify(616943)

				idx = wfr.unboundedFollowing()
			} else {
				__antithesis_instrumentation__.Notify(616944)
			}
			__antithesis_instrumentation__.Notify(616937)
			return idx, nil
		case treewindow.UnboundedFollowing:
			__antithesis_instrumentation__.Notify(616938)
			return wfr.unboundedFollowing(), nil
		default:
			__antithesis_instrumentation__.Notify(616939)
			return 0, errors.AssertionFailedf(
				"unexpected WindowFrameBoundType in ROWS mode: %d",
				redact.Safe(wfr.Frame.Bounds.EndBound.BoundType))
		}
	case treewindow.GROUPS:
		__antithesis_instrumentation__.Notify(616869)
		if wfr.Frame.Bounds.EndBound == nil {
			__antithesis_instrumentation__.Notify(616945)

			return wfr.DefaultFrameSize(), nil
		} else {
			__antithesis_instrumentation__.Notify(616946)
		}
		__antithesis_instrumentation__.Notify(616870)
		switch wfr.Frame.Bounds.EndBound.BoundType {
		case treewindow.OffsetPreceding:
			__antithesis_instrumentation__.Notify(616947)
			offset := MustBeDInt(wfr.EndBoundOffset)
			peerGroupNum := wfr.CurRowPeerGroupNum - int(offset)
			if peerGroupNum < wfr.PeerHelper.headPeerGroupNum {
				__antithesis_instrumentation__.Notify(616954)

				return 0, nil
			} else {
				__antithesis_instrumentation__.Notify(616955)
			}
			__antithesis_instrumentation__.Notify(616948)
			return wfr.PeerHelper.GetFirstPeerIdx(peerGroupNum) + wfr.PeerHelper.GetRowCount(peerGroupNum), nil
		case treewindow.CurrentRow:
			__antithesis_instrumentation__.Notify(616949)
			return wfr.DefaultFrameSize(), nil
		case treewindow.OffsetFollowing:
			__antithesis_instrumentation__.Notify(616950)
			offset := MustBeDInt(wfr.EndBoundOffset)
			peerGroupNum := wfr.CurRowPeerGroupNum + int(offset)
			lastPeerGroupNum := wfr.PeerHelper.GetLastPeerGroupNum()
			if peerGroupNum > lastPeerGroupNum || func() bool {
				__antithesis_instrumentation__.Notify(616956)
				return peerGroupNum < 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(616957)

				return wfr.unboundedFollowing(), nil
			} else {
				__antithesis_instrumentation__.Notify(616958)
			}
			__antithesis_instrumentation__.Notify(616951)
			return wfr.PeerHelper.GetFirstPeerIdx(peerGroupNum) + wfr.PeerHelper.GetRowCount(peerGroupNum), nil
		case treewindow.UnboundedFollowing:
			__antithesis_instrumentation__.Notify(616952)
			return wfr.unboundedFollowing(), nil
		default:
			__antithesis_instrumentation__.Notify(616953)
			return 0, errors.AssertionFailedf(
				"unexpected WindowFrameBoundType in GROUPS mode: %d",
				redact.Safe(wfr.Frame.Bounds.EndBound.BoundType))
		}
	default:
		__antithesis_instrumentation__.Notify(616871)
		return 0, errors.AssertionFailedf(
			"unexpected WindowFrameMode: %d", redact.Safe(wfr.Frame.Mode))
	}
}

func (wfr *WindowFrameRun) FrameSize(ctx context.Context, evalCtx *EvalContext) (int, error) {
	__antithesis_instrumentation__.Notify(616959)
	if wfr.Frame == nil {
		__antithesis_instrumentation__.Notify(616965)
		return wfr.DefaultFrameSize(), nil
	} else {
		__antithesis_instrumentation__.Notify(616966)
	}
	__antithesis_instrumentation__.Notify(616960)
	frameEndIdx, err := wfr.FrameEndIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(616967)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(616968)
	}
	__antithesis_instrumentation__.Notify(616961)
	frameStartIdx, err := wfr.FrameStartIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(616969)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(616970)
	}
	__antithesis_instrumentation__.Notify(616962)
	size := frameEndIdx - frameStartIdx
	if !wfr.noFilter() || func() bool {
		__antithesis_instrumentation__.Notify(616971)
		return !wfr.Frame.DefaultFrameExclusion() == true
	}() == true {
		__antithesis_instrumentation__.Notify(616972)
		size = 0
		for idx := frameStartIdx; idx < frameEndIdx; idx++ {
			__antithesis_instrumentation__.Notify(616973)
			if skipped, err := wfr.IsRowSkipped(ctx, idx); err != nil {
				__antithesis_instrumentation__.Notify(616975)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(616976)
				if skipped {
					__antithesis_instrumentation__.Notify(616977)
					continue
				} else {
					__antithesis_instrumentation__.Notify(616978)
				}
			}
			__antithesis_instrumentation__.Notify(616974)
			size++
		}
	} else {
		__antithesis_instrumentation__.Notify(616979)
	}
	__antithesis_instrumentation__.Notify(616963)
	if size <= 0 {
		__antithesis_instrumentation__.Notify(616980)
		size = 0
	} else {
		__antithesis_instrumentation__.Notify(616981)
	}
	__antithesis_instrumentation__.Notify(616964)
	return size, nil
}

func (wfr *WindowFrameRun) Rank() int {
	__antithesis_instrumentation__.Notify(616982)
	return wfr.RowIdx + 1
}

func (wfr *WindowFrameRun) PartitionSize() int {
	__antithesis_instrumentation__.Notify(616983)
	return wfr.Rows.Len()
}

func (wfr *WindowFrameRun) unboundedFollowing() int {
	__antithesis_instrumentation__.Notify(616984)
	return wfr.PartitionSize()
}

func (wfr *WindowFrameRun) DefaultFrameSize() int {
	__antithesis_instrumentation__.Notify(616985)
	return wfr.PeerHelper.GetFirstPeerIdx(wfr.CurRowPeerGroupNum) + wfr.PeerHelper.GetRowCount(wfr.CurRowPeerGroupNum)
}

func (wfr *WindowFrameRun) FirstInPeerGroup() bool {
	__antithesis_instrumentation__.Notify(616986)
	return wfr.RowIdx == wfr.PeerHelper.GetFirstPeerIdx(wfr.CurRowPeerGroupNum)
}

func (wfr *WindowFrameRun) Args(ctx context.Context) (Datums, error) {
	__antithesis_instrumentation__.Notify(616987)
	return wfr.ArgsWithRowOffset(ctx, 0)
}

func (wfr *WindowFrameRun) ArgsWithRowOffset(ctx context.Context, offset int) (Datums, error) {
	__antithesis_instrumentation__.Notify(616988)
	return wfr.ArgsByRowIdx(ctx, wfr.RowIdx+offset)
}

func (wfr *WindowFrameRun) ArgsByRowIdx(ctx context.Context, idx int) (Datums, error) {
	__antithesis_instrumentation__.Notify(616989)
	row, err := wfr.Rows.GetRow(ctx, idx)
	if err != nil {
		__antithesis_instrumentation__.Notify(616992)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(616993)
	}
	__antithesis_instrumentation__.Notify(616990)
	datums := make(Datums, len(wfr.ArgsIdxs))
	for i, argIdx := range wfr.ArgsIdxs {
		__antithesis_instrumentation__.Notify(616994)
		datums[i], err = row.GetDatum(int(argIdx))
		if err != nil {
			__antithesis_instrumentation__.Notify(616995)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(616996)
		}
	}
	__antithesis_instrumentation__.Notify(616991)
	return datums, nil
}

func (wfr *WindowFrameRun) valueAt(ctx context.Context, idx int) (Datum, error) {
	__antithesis_instrumentation__.Notify(616997)
	row, err := wfr.Rows.GetRow(ctx, idx)
	if err != nil {
		__antithesis_instrumentation__.Notify(616999)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(617000)
	}
	__antithesis_instrumentation__.Notify(616998)
	return row.GetDatum(wfr.OrdColIdx)
}

func (wfr *WindowFrameRun) RangeModeWithOffsets() bool {
	__antithesis_instrumentation__.Notify(617001)
	return wfr.Frame.Mode == treewindow.RANGE && func() bool {
		__antithesis_instrumentation__.Notify(617002)
		return wfr.Frame.Bounds.HasOffset() == true
	}() == true
}

func (wfr *WindowFrameRun) FullPartitionIsInWindow() bool {
	__antithesis_instrumentation__.Notify(617003)

	if wfr.Frame == nil || func() bool {
		__antithesis_instrumentation__.Notify(617007)
		return !wfr.Frame.DefaultFrameExclusion() == true
	}() == true {
		__antithesis_instrumentation__.Notify(617008)
		return false
	} else {
		__antithesis_instrumentation__.Notify(617009)
	}
	__antithesis_instrumentation__.Notify(617004)
	if wfr.Frame.Bounds.EndBound == nil {
		__antithesis_instrumentation__.Notify(617010)

		return false
	} else {
		__antithesis_instrumentation__.Notify(617011)
	}
	__antithesis_instrumentation__.Notify(617005)

	precedingConfirmed := wfr.Frame.Bounds.StartBound.BoundType == treewindow.UnboundedPreceding
	followingConfirmed := wfr.Frame.Bounds.EndBound.BoundType == treewindow.UnboundedFollowing
	if wfr.Frame.Mode == treewindow.ROWS || func() bool {
		__antithesis_instrumentation__.Notify(617012)
		return wfr.Frame.Mode == treewindow.GROUPS == true
	}() == true {
		__antithesis_instrumentation__.Notify(617013)

		if wfr.Frame.Bounds.StartBound.BoundType == treewindow.OffsetPreceding {
			__antithesis_instrumentation__.Notify(617015)

			startOffset := wfr.StartBoundOffset.(*DInt)

			precedingConfirmed = precedingConfirmed || func() bool {
				__antithesis_instrumentation__.Notify(617016)
				return *startOffset >= DInt(wfr.Rows.Len()-1) == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(617017)
		}
		__antithesis_instrumentation__.Notify(617014)
		if wfr.Frame.Bounds.EndBound.BoundType == treewindow.OffsetFollowing {
			__antithesis_instrumentation__.Notify(617018)

			endOffset := wfr.EndBoundOffset.(*DInt)

			followingConfirmed = followingConfirmed || func() bool {
				__antithesis_instrumentation__.Notify(617019)
				return *endOffset >= DInt(wfr.Rows.Len()-1) == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(617020)
		}
	} else {
		__antithesis_instrumentation__.Notify(617021)
	}
	__antithesis_instrumentation__.Notify(617006)
	return precedingConfirmed && func() bool {
		__antithesis_instrumentation__.Notify(617022)
		return followingConfirmed == true
	}() == true
}

func (wfr *WindowFrameRun) noFilter() bool {
	__antithesis_instrumentation__.Notify(617023)
	return wfr.FilterColIdx == NoColumnIdx
}

func (wfr *WindowFrameRun) isRowExcluded(idx int) (bool, error) {
	__antithesis_instrumentation__.Notify(617024)
	if wfr.Frame.DefaultFrameExclusion() {
		__antithesis_instrumentation__.Notify(617026)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(617027)
	}
	__antithesis_instrumentation__.Notify(617025)
	switch wfr.Frame.Exclusion {
	case treewindow.ExcludeCurrentRow:
		__antithesis_instrumentation__.Notify(617028)
		return idx == wfr.RowIdx, nil
	case treewindow.ExcludeGroup:
		__antithesis_instrumentation__.Notify(617029)
		curRowFirstPeerIdx := wfr.PeerHelper.GetFirstPeerIdx(wfr.CurRowPeerGroupNum)
		curRowPeerGroupRowCount := wfr.PeerHelper.GetRowCount(wfr.CurRowPeerGroupNum)
		return curRowFirstPeerIdx <= idx && func() bool {
			__antithesis_instrumentation__.Notify(617032)
			return idx < curRowFirstPeerIdx+curRowPeerGroupRowCount == true
		}() == true, nil
	case treewindow.ExcludeTies:
		__antithesis_instrumentation__.Notify(617030)
		curRowFirstPeerIdx := wfr.PeerHelper.GetFirstPeerIdx(wfr.CurRowPeerGroupNum)
		curRowPeerGroupRowCount := wfr.PeerHelper.GetRowCount(wfr.CurRowPeerGroupNum)
		return curRowFirstPeerIdx <= idx && func() bool {
			__antithesis_instrumentation__.Notify(617033)
			return idx < curRowFirstPeerIdx+curRowPeerGroupRowCount == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(617034)
			return idx != wfr.RowIdx == true
		}() == true, nil
	default:
		__antithesis_instrumentation__.Notify(617031)
		return false, errors.AssertionFailedf("unexpected WindowFrameExclusion")
	}
}

func (wfr *WindowFrameRun) IsRowSkipped(ctx context.Context, idx int) (bool, error) {
	__antithesis_instrumentation__.Notify(617035)
	if !wfr.noFilter() {
		__antithesis_instrumentation__.Notify(617037)
		row, err := wfr.Rows.GetRow(ctx, idx)
		if err != nil {
			__antithesis_instrumentation__.Notify(617040)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(617041)
		}
		__antithesis_instrumentation__.Notify(617038)
		d, err := row.GetDatum(wfr.FilterColIdx)
		if err != nil {
			__antithesis_instrumentation__.Notify(617042)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(617043)
		}
		__antithesis_instrumentation__.Notify(617039)
		if d != DBoolTrue {
			__antithesis_instrumentation__.Notify(617044)

			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(617045)
		}
	} else {
		__antithesis_instrumentation__.Notify(617046)
	}
	__antithesis_instrumentation__.Notify(617036)

	return wfr.isRowExcluded(idx)
}

func compareForWindow(evalCtx *EvalContext, left, right Datum) (int, error) {
	__antithesis_instrumentation__.Notify(617047)
	if types.IsDateTimeType(left.ResolvedType()) && func() bool {
		__antithesis_instrumentation__.Notify(617049)
		return left.ResolvedType() != types.Interval == true
	}() == true {
		__antithesis_instrumentation__.Notify(617050)

		ts, err := timeFromDatumForComparison(evalCtx, left)
		if err != nil {
			__antithesis_instrumentation__.Notify(617052)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(617053)
		}
		__antithesis_instrumentation__.Notify(617051)
		left, err = MakeDTimestampTZ(ts, time.Microsecond)
		if err != nil {
			__antithesis_instrumentation__.Notify(617054)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(617055)
		}
	} else {
		__antithesis_instrumentation__.Notify(617056)
	}
	__antithesis_instrumentation__.Notify(617048)
	return left.Compare(evalCtx, right), nil
}

type WindowFunc interface {
	Compute(context.Context, *EvalContext, *WindowFrameRun) (Datum, error)

	Reset(context.Context)

	Close(context.Context, *EvalContext)
}
