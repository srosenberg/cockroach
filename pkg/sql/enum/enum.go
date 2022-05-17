package enum

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "bytes"

const minToken int = 0
const maxToken int = 256
const midToken = maxToken / 2
const shiftInterval = 5

type ByteSpacing int

const (
	PackedSpacing ByteSpacing = iota

	SpreadSpacing
)

func (s ByteSpacing) String() string {
	__antithesis_instrumentation__.Notify(469975)
	switch s {
	case PackedSpacing:
		__antithesis_instrumentation__.Notify(469977)
		return "packed"
	case SpreadSpacing:
		__antithesis_instrumentation__.Notify(469978)
		return "spread"
	default:
		__antithesis_instrumentation__.Notify(469979)
	}
	__antithesis_instrumentation__.Notify(469976)
	panic("unknown spacing type")
}

func GenByteStringBetween(prev []byte, next []byte, spacing ByteSpacing) []byte {
	__antithesis_instrumentation__.Notify(469980)
	result := make([]byte, 0)
	if len(prev) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(469986)
		return len(next) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(469987)

		return append(result, byte(midToken))
	} else {
		__antithesis_instrumentation__.Notify(469988)
	}
	__antithesis_instrumentation__.Notify(469981)
	maxLen := len(prev)
	if len(next) > maxLen {
		__antithesis_instrumentation__.Notify(469989)
		maxLen = len(next)
	} else {
		__antithesis_instrumentation__.Notify(469990)
	}
	__antithesis_instrumentation__.Notify(469982)

	pos := 0
	for ; pos < maxLen; pos++ {
		__antithesis_instrumentation__.Notify(469991)
		p, n := get(prev, pos, minToken), get(next, pos, maxToken)
		if p != n {
			__antithesis_instrumentation__.Notify(469993)
			break
		} else {
			__antithesis_instrumentation__.Notify(469994)
		}
		__antithesis_instrumentation__.Notify(469992)
		result = append(result, byte(p))
	}
	__antithesis_instrumentation__.Notify(469983)

	p, n := get(prev, pos, minToken), get(next, pos, maxToken)
	var mid int
	switch spacing {
	case PackedSpacing:
		__antithesis_instrumentation__.Notify(469995)
		mid = byteBetweenPacked(p, n)
	case SpreadSpacing:
		__antithesis_instrumentation__.Notify(469996)
		mid = byteBetweenSpread(p, n)
	default:
		__antithesis_instrumentation__.Notify(469997)
	}
	__antithesis_instrumentation__.Notify(469984)

	if mid == p {
		__antithesis_instrumentation__.Notify(469998)
		result = append(result, byte(p))
		rest := GenByteStringBetween(slice(prev, pos+1), nil, spacing)
		return append(result, rest...)
	} else {
		__antithesis_instrumentation__.Notify(469999)
	}
	__antithesis_instrumentation__.Notify(469985)

	result = append(result, byte(mid))
	return result
}

func get(arr []byte, idx int, def int) int {
	__antithesis_instrumentation__.Notify(470000)
	if idx >= len(arr) {
		__antithesis_instrumentation__.Notify(470002)
		return def
	} else {
		__antithesis_instrumentation__.Notify(470003)
	}
	__antithesis_instrumentation__.Notify(470001)
	return int(arr[idx])
}

func slice(arr []byte, idx int) []byte {
	__antithesis_instrumentation__.Notify(470004)
	if idx > len(arr) {
		__antithesis_instrumentation__.Notify(470006)
		return []byte(nil)
	} else {
		__antithesis_instrumentation__.Notify(470007)
	}
	__antithesis_instrumentation__.Notify(470005)
	return arr[idx:]
}

func byteBetweenPacked(lo int, hi int) int {
	__antithesis_instrumentation__.Notify(470008)
	switch {
	case lo == minToken && func() bool {
		__antithesis_instrumentation__.Notify(470013)
		return hi == maxToken == true
	}() == true:
		__antithesis_instrumentation__.Notify(470009)
		return lo + (hi-lo)/2
	case lo == minToken && func() bool {
		__antithesis_instrumentation__.Notify(470014)
		return hi-lo > shiftInterval == true
	}() == true:
		__antithesis_instrumentation__.Notify(470010)
		return hi - shiftInterval
	case hi == maxToken && func() bool {
		__antithesis_instrumentation__.Notify(470015)
		return hi-lo > shiftInterval == true
	}() == true:
		__antithesis_instrumentation__.Notify(470011)
		return lo + shiftInterval
	default:
		__antithesis_instrumentation__.Notify(470012)
		return lo + (hi-lo)/2
	}
}

func byteBetweenSpread(lo int, hi int) int {
	__antithesis_instrumentation__.Notify(470016)
	return lo + (hi-lo)/2
}

func GenerateNEvenlySpacedBytes(n int) [][]byte {
	__antithesis_instrumentation__.Notify(470017)
	result := make([][]byte, n)
	genEvenlySpacedHelper(result, 0, n, nil, nil)
	return result
}

func genEvenlySpacedHelper(result [][]byte, lo, hi int, bot, top []byte) {
	__antithesis_instrumentation__.Notify(470018)
	if lo == hi {
		__antithesis_instrumentation__.Notify(470020)
		return
	} else {
		__antithesis_instrumentation__.Notify(470021)
	}
	__antithesis_instrumentation__.Notify(470019)
	mid := lo + (hi-lo)/2
	midBytes := GenByteStringBetween(bot, top, SpreadSpacing)
	result[mid] = midBytes
	genEvenlySpacedHelper(result, lo, mid, bot, midBytes)
	genEvenlySpacedHelper(result, mid+1, hi, midBytes, top)
}

func enumBytesAreLess(a []byte, b []byte) bool {
	__antithesis_instrumentation__.Notify(470022)
	if len(b) == 0 {
		__antithesis_instrumentation__.Notify(470024)
		return true
	} else {
		__antithesis_instrumentation__.Notify(470025)
	}
	__antithesis_instrumentation__.Notify(470023)
	return bytes.Compare(a, b) == -1
}
