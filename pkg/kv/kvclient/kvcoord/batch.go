package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
)

var emptySpan = roachpb.Span{}

func Truncate(
	reqs []roachpb.RequestUnion, rs roachpb.RSpan,
) ([]roachpb.RequestUnion, []int, error) {
	__antithesis_instrumentation__.Notify(86886)
	truncateOne := func(args roachpb.Request) (bool, roachpb.Span, error) {
		__antithesis_instrumentation__.Notify(86889)
		header := args.Header().Span()
		if !roachpb.IsRange(args) {
			__antithesis_instrumentation__.Notify(86897)

			if len(header.EndKey) > 0 {
				__antithesis_instrumentation__.Notify(86901)
				return false, emptySpan, errors.Errorf("%T is not a range command, but EndKey is set", args)
			} else {
				__antithesis_instrumentation__.Notify(86902)
			}
			__antithesis_instrumentation__.Notify(86898)
			keyAddr, err := keys.Addr(header.Key)
			if err != nil {
				__antithesis_instrumentation__.Notify(86903)
				return false, emptySpan, err
			} else {
				__antithesis_instrumentation__.Notify(86904)
			}
			__antithesis_instrumentation__.Notify(86899)
			if !rs.ContainsKey(keyAddr) {
				__antithesis_instrumentation__.Notify(86905)
				return false, emptySpan, nil
			} else {
				__antithesis_instrumentation__.Notify(86906)
			}
			__antithesis_instrumentation__.Notify(86900)
			return true, header, nil
		} else {
			__antithesis_instrumentation__.Notify(86907)
		}
		__antithesis_instrumentation__.Notify(86890)

		local := false
		keyAddr, err := keys.Addr(header.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(86908)
			return false, emptySpan, err
		} else {
			__antithesis_instrumentation__.Notify(86909)
		}
		__antithesis_instrumentation__.Notify(86891)
		endKeyAddr, err := keys.Addr(header.EndKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(86910)
			return false, emptySpan, err
		} else {
			__antithesis_instrumentation__.Notify(86911)
		}
		__antithesis_instrumentation__.Notify(86892)
		if l, r := keys.IsLocal(header.Key), keys.IsLocal(header.EndKey); l || func() bool {
			__antithesis_instrumentation__.Notify(86912)
			return r == true
		}() == true {
			__antithesis_instrumentation__.Notify(86913)
			if !l || func() bool {
				__antithesis_instrumentation__.Notify(86915)
				return !r == true
			}() == true {
				__antithesis_instrumentation__.Notify(86916)
				return false, emptySpan, errors.Errorf("local key mixed with global key in range")
			} else {
				__antithesis_instrumentation__.Notify(86917)
			}
			__antithesis_instrumentation__.Notify(86914)
			local = true
		} else {
			__antithesis_instrumentation__.Notify(86918)
		}
		__antithesis_instrumentation__.Notify(86893)
		if keyAddr.Less(rs.Key) {
			__antithesis_instrumentation__.Notify(86919)

			if !local {
				__antithesis_instrumentation__.Notify(86920)
				header.Key = rs.Key.AsRawKey()
			} else {
				__antithesis_instrumentation__.Notify(86921)

				header.Key = keys.MakeRangeKeyPrefix(rs.Key)
			}
		} else {
			__antithesis_instrumentation__.Notify(86922)
		}
		__antithesis_instrumentation__.Notify(86894)
		if !endKeyAddr.Less(rs.EndKey) {
			__antithesis_instrumentation__.Notify(86923)

			if !local {
				__antithesis_instrumentation__.Notify(86924)
				header.EndKey = rs.EndKey.AsRawKey()
			} else {
				__antithesis_instrumentation__.Notify(86925)

				header.EndKey = keys.MakeRangeKeyPrefix(rs.EndKey)
			}
		} else {
			__antithesis_instrumentation__.Notify(86926)
		}
		__antithesis_instrumentation__.Notify(86895)

		if header.Key.Compare(header.EndKey) >= 0 {
			__antithesis_instrumentation__.Notify(86927)
			return false, emptySpan, nil
		} else {
			__antithesis_instrumentation__.Notify(86928)
		}
		__antithesis_instrumentation__.Notify(86896)
		return true, header, nil
	}
	__antithesis_instrumentation__.Notify(86887)

	var positions []int
	var truncReqs []roachpb.RequestUnion
	for pos, arg := range reqs {
		__antithesis_instrumentation__.Notify(86929)
		hasRequest, newSpan, err := truncateOne(arg.GetInner())
		if hasRequest {
			__antithesis_instrumentation__.Notify(86931)

			inner := reqs[pos].GetInner()
			oldHeader := inner.Header()
			if newSpan.EqualValue(oldHeader.Span()) {
				__antithesis_instrumentation__.Notify(86933)
				truncReqs = append(truncReqs, reqs[pos])
			} else {
				__antithesis_instrumentation__.Notify(86934)
				oldHeader.SetSpan(newSpan)
				shallowCopy := inner.ShallowCopy()
				shallowCopy.SetHeader(oldHeader)
				truncReqs = append(truncReqs, roachpb.RequestUnion{})
				truncReqs[len(truncReqs)-1].MustSetInner(shallowCopy)
			}
			__antithesis_instrumentation__.Notify(86932)
			positions = append(positions, pos)
		} else {
			__antithesis_instrumentation__.Notify(86935)
		}
		__antithesis_instrumentation__.Notify(86930)
		if err != nil {
			__antithesis_instrumentation__.Notify(86936)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(86937)
		}
	}
	__antithesis_instrumentation__.Notify(86888)
	return truncReqs, positions, nil
}

func prev(reqs []roachpb.RequestUnion, k roachpb.RKey) (roachpb.RKey, error) {
	__antithesis_instrumentation__.Notify(86938)
	candidate := roachpb.RKeyMin
	for _, union := range reqs {
		__antithesis_instrumentation__.Notify(86940)
		inner := union.GetInner()
		h := inner.Header()
		addr, err := keys.Addr(h.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(86945)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(86946)
		}
		__antithesis_instrumentation__.Notify(86941)
		endKey := h.EndKey
		if len(endKey) == 0 {
			__antithesis_instrumentation__.Notify(86947)

			endKey = h.Key.Next()
		} else {
			__antithesis_instrumentation__.Notify(86948)
		}
		__antithesis_instrumentation__.Notify(86942)
		eAddr, err := keys.AddrUpperBound(endKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(86949)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(86950)
		}
		__antithesis_instrumentation__.Notify(86943)
		if !eAddr.Less(k) {
			__antithesis_instrumentation__.Notify(86951)

			if addr.Less(k) {
				__antithesis_instrumentation__.Notify(86953)

				return k, nil
			} else {
				__antithesis_instrumentation__.Notify(86954)
			}
			__antithesis_instrumentation__.Notify(86952)

			continue
		} else {
			__antithesis_instrumentation__.Notify(86955)
		}
		__antithesis_instrumentation__.Notify(86944)

		if candidate.Less(eAddr) {
			__antithesis_instrumentation__.Notify(86956)
			candidate = eAddr
		} else {
			__antithesis_instrumentation__.Notify(86957)
		}
	}
	__antithesis_instrumentation__.Notify(86939)
	return candidate, nil
}

func Next(reqs []roachpb.RequestUnion, k roachpb.RKey) (roachpb.RKey, error) {
	__antithesis_instrumentation__.Notify(86958)
	candidate := roachpb.RKeyMax
	for _, union := range reqs {
		__antithesis_instrumentation__.Notify(86960)
		inner := union.GetInner()
		h := inner.Header()
		addr, err := keys.Addr(h.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(86963)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(86964)
		}
		__antithesis_instrumentation__.Notify(86961)
		if addr.Less(k) {
			__antithesis_instrumentation__.Notify(86965)
			if len(h.EndKey) == 0 {
				__antithesis_instrumentation__.Notify(86969)

				continue
			} else {
				__antithesis_instrumentation__.Notify(86970)
			}
			__antithesis_instrumentation__.Notify(86966)
			eAddr, err := keys.AddrUpperBound(h.EndKey)
			if err != nil {
				__antithesis_instrumentation__.Notify(86971)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(86972)
			}
			__antithesis_instrumentation__.Notify(86967)
			if k.Less(eAddr) {
				__antithesis_instrumentation__.Notify(86973)

				return k, nil
			} else {
				__antithesis_instrumentation__.Notify(86974)
			}
			__antithesis_instrumentation__.Notify(86968)

			continue
		} else {
			__antithesis_instrumentation__.Notify(86975)
		}
		__antithesis_instrumentation__.Notify(86962)

		if addr.Less(candidate) {
			__antithesis_instrumentation__.Notify(86976)
			candidate = addr
		} else {
			__antithesis_instrumentation__.Notify(86977)
		}
	}
	__antithesis_instrumentation__.Notify(86959)
	return candidate, nil
}
