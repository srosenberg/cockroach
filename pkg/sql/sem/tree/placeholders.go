package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type PlaceholderIdx uint16

const MaxPlaceholderIdx = math.MaxUint16

func (idx PlaceholderIdx) String() string {
	__antithesis_instrumentation__.Notify(611756)
	return fmt.Sprintf("$%d", idx+1)
}

type PlaceholderTypes []*types.T

func (pt PlaceholderTypes) Identical(other PlaceholderTypes) bool {
	__antithesis_instrumentation__.Notify(611757)
	if len(pt) != len(other) {
		__antithesis_instrumentation__.Notify(611760)
		return false
	} else {
		__antithesis_instrumentation__.Notify(611761)
	}
	__antithesis_instrumentation__.Notify(611758)
	for i, t := range pt {
		__antithesis_instrumentation__.Notify(611762)
		switch {
		case t == nil && func() bool {
			__antithesis_instrumentation__.Notify(611767)
			return other[i] == nil == true
		}() == true:
			__antithesis_instrumentation__.Notify(611763)
		case t == nil || func() bool {
			__antithesis_instrumentation__.Notify(611768)
			return other[i] == nil == true
		}() == true:
			__antithesis_instrumentation__.Notify(611764)
			return false
		case !t.Identical(other[i]):
			__antithesis_instrumentation__.Notify(611765)
			return false
		default:
			__antithesis_instrumentation__.Notify(611766)
		}
	}
	__antithesis_instrumentation__.Notify(611759)
	return true
}

func (pt PlaceholderTypes) AssertAllSet() error {
	__antithesis_instrumentation__.Notify(611769)
	for i := range pt {
		__antithesis_instrumentation__.Notify(611771)
		if pt[i] == nil {
			__antithesis_instrumentation__.Notify(611772)
			return placeholderTypeAmbiguityError(PlaceholderIdx(i))
		} else {
			__antithesis_instrumentation__.Notify(611773)
		}
	}
	__antithesis_instrumentation__.Notify(611770)
	return nil
}

type QueryArguments []TypedExpr

func (qa QueryArguments) String() string {
	__antithesis_instrumentation__.Notify(611774)
	if len(qa) == 0 {
		__antithesis_instrumentation__.Notify(611777)
		return "{}"
	} else {
		__antithesis_instrumentation__.Notify(611778)
	}
	__antithesis_instrumentation__.Notify(611775)
	var buf bytes.Buffer
	buf.WriteByte('{')
	sep := ""
	for k, v := range qa {
		__antithesis_instrumentation__.Notify(611779)
		fmt.Fprintf(&buf, "%s%s:%q", sep, PlaceholderIdx(k), v)
		sep = ", "
	}
	__antithesis_instrumentation__.Notify(611776)
	buf.WriteByte('}')
	return buf.String()
}

type PlaceholderTypesInfo struct {
	TypeHints PlaceholderTypes

	Types PlaceholderTypes
}

func (p *PlaceholderTypesInfo) Type(idx PlaceholderIdx) (_ *types.T, ok bool, _ error) {
	__antithesis_instrumentation__.Notify(611780)
	if len(p.Types) <= int(idx) {
		__antithesis_instrumentation__.Notify(611783)
		return nil, false, makeNoValueProvidedForPlaceholderErr(idx)
	} else {
		__antithesis_instrumentation__.Notify(611784)
	}
	__antithesis_instrumentation__.Notify(611781)
	t := p.Types[idx]
	if t == nil && func() bool {
		__antithesis_instrumentation__.Notify(611785)
		return len(p.TypeHints) > int(idx) == true
	}() == true {
		__antithesis_instrumentation__.Notify(611786)
		t = p.TypeHints[idx]
	} else {
		__antithesis_instrumentation__.Notify(611787)
	}
	__antithesis_instrumentation__.Notify(611782)
	return t, t != nil, nil
}

func (p *PlaceholderTypesInfo) ValueType(idx PlaceholderIdx) (_ *types.T, ok bool) {
	__antithesis_instrumentation__.Notify(611788)
	var t *types.T
	if len(p.TypeHints) >= int(idx) {
		__antithesis_instrumentation__.Notify(611791)
		t = p.TypeHints[idx]
	} else {
		__antithesis_instrumentation__.Notify(611792)
	}
	__antithesis_instrumentation__.Notify(611789)
	if t == nil {
		__antithesis_instrumentation__.Notify(611793)
		t = p.Types[idx]
	} else {
		__antithesis_instrumentation__.Notify(611794)
	}
	__antithesis_instrumentation__.Notify(611790)
	return t, (t != nil)
}

func (p *PlaceholderTypesInfo) SetType(idx PlaceholderIdx, typ *types.T) error {
	__antithesis_instrumentation__.Notify(611795)
	if t := p.Types[idx]; t != nil {
		__antithesis_instrumentation__.Notify(611797)
		if !typ.Equivalent(t) {
			__antithesis_instrumentation__.Notify(611799)
			return pgerror.Newf(
				pgcode.DatatypeMismatch,
				"placeholder %s already has type %s, cannot assign %s", idx, t, typ)
		} else {
			__antithesis_instrumentation__.Notify(611800)
		}
		__antithesis_instrumentation__.Notify(611798)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(611801)
	}
	__antithesis_instrumentation__.Notify(611796)
	p.Types[idx] = typ
	return nil
}

type PlaceholderInfo struct {
	PlaceholderTypesInfo

	Values QueryArguments
}

func (p *PlaceholderInfo) Init(numPlaceholders int, typeHints PlaceholderTypes) error {
	__antithesis_instrumentation__.Notify(611802)
	p.Types = make(PlaceholderTypes, numPlaceholders)
	if typeHints == nil {
		__antithesis_instrumentation__.Notify(611804)
		p.TypeHints = make(PlaceholderTypes, numPlaceholders)
	} else {
		__antithesis_instrumentation__.Notify(611805)
		if err := checkPlaceholderArity(len(typeHints), numPlaceholders); err != nil {
			__antithesis_instrumentation__.Notify(611807)
			return err
		} else {
			__antithesis_instrumentation__.Notify(611808)
		}
		__antithesis_instrumentation__.Notify(611806)
		p.TypeHints = typeHints
	}
	__antithesis_instrumentation__.Notify(611803)
	p.Values = nil
	return nil
}

func (p *PlaceholderInfo) Assign(src *PlaceholderInfo, numPlaceholders int) error {
	__antithesis_instrumentation__.Notify(611809)
	if src != nil {
		__antithesis_instrumentation__.Notify(611811)
		if err := checkPlaceholderArity(len(src.Types), numPlaceholders); err != nil {
			__antithesis_instrumentation__.Notify(611813)
			return err
		} else {
			__antithesis_instrumentation__.Notify(611814)
		}
		__antithesis_instrumentation__.Notify(611812)
		*p = *src
		return nil
	} else {
		__antithesis_instrumentation__.Notify(611815)
	}
	__antithesis_instrumentation__.Notify(611810)
	return p.Init(numPlaceholders, nil)
}

func checkPlaceholderArity(numTypes, numPlaceholders int) error {
	__antithesis_instrumentation__.Notify(611816)
	if numTypes > numPlaceholders {
		__antithesis_instrumentation__.Notify(611818)
		return errors.AssertionFailedf(
			"unexpected placeholder types: got %d, expected %d",
			numTypes, numPlaceholders)
	} else {
		__antithesis_instrumentation__.Notify(611819)
		if numTypes < numPlaceholders {
			__antithesis_instrumentation__.Notify(611820)
			return pgerror.Newf(pgcode.UndefinedParameter,
				"could not find types for all placeholders: got %d, expected %d",
				numTypes, numPlaceholders)
		} else {
			__antithesis_instrumentation__.Notify(611821)
		}
	}
	__antithesis_instrumentation__.Notify(611817)
	return nil
}

func (p *PlaceholderInfo) Value(idx PlaceholderIdx) (TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(611822)
	if len(p.Values) <= int(idx) || func() bool {
		__antithesis_instrumentation__.Notify(611824)
		return p.Values[idx] == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(611825)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(611826)
	}
	__antithesis_instrumentation__.Notify(611823)
	return p.Values[idx], true
}

func (p *PlaceholderInfo) IsUnresolvedPlaceholder(expr Expr) bool {
	__antithesis_instrumentation__.Notify(611827)
	if t, ok := StripParens(expr).(*Placeholder); ok {
		__antithesis_instrumentation__.Notify(611829)
		_, res, err := p.Type(t.Idx)
		return !(err == nil && func() bool {
			__antithesis_instrumentation__.Notify(611830)
			return res == true
		}() == true)
	} else {
		__antithesis_instrumentation__.Notify(611831)
	}
	__antithesis_instrumentation__.Notify(611828)
	return false
}
