package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func (v Version) Less(otherV Version) bool {
	__antithesis_instrumentation__.Notify(179907)
	if v.Major < otherV.Major {
		__antithesis_instrumentation__.Notify(179912)
		return true
	} else {
		__antithesis_instrumentation__.Notify(179913)
		if v.Major > otherV.Major {
			__antithesis_instrumentation__.Notify(179914)
			return false
		} else {
			__antithesis_instrumentation__.Notify(179915)
		}
	}
	__antithesis_instrumentation__.Notify(179908)
	if v.Minor < otherV.Minor {
		__antithesis_instrumentation__.Notify(179916)
		return true
	} else {
		__antithesis_instrumentation__.Notify(179917)
		if v.Minor > otherV.Minor {
			__antithesis_instrumentation__.Notify(179918)
			return false
		} else {
			__antithesis_instrumentation__.Notify(179919)
		}
	}
	__antithesis_instrumentation__.Notify(179909)
	if v.Patch < otherV.Patch {
		__antithesis_instrumentation__.Notify(179920)
		return true
	} else {
		__antithesis_instrumentation__.Notify(179921)
		if v.Patch > otherV.Patch {
			__antithesis_instrumentation__.Notify(179922)
			return false
		} else {
			__antithesis_instrumentation__.Notify(179923)
		}
	}
	__antithesis_instrumentation__.Notify(179910)
	if v.Internal < otherV.Internal {
		__antithesis_instrumentation__.Notify(179924)
		return true
	} else {
		__antithesis_instrumentation__.Notify(179925)
		if v.Internal > otherV.Internal {
			__antithesis_instrumentation__.Notify(179926)
			return false
		} else {
			__antithesis_instrumentation__.Notify(179927)
		}
	}
	__antithesis_instrumentation__.Notify(179911)
	return false
}

func (v Version) LessEq(otherV Version) bool {
	__antithesis_instrumentation__.Notify(179928)
	return v.Equal(otherV) || func() bool {
		__antithesis_instrumentation__.Notify(179929)
		return v.Less(otherV) == true
	}() == true
}

func (v Version) String() string {
	__antithesis_instrumentation__.Notify(179930)
	return redact.StringWithoutMarkers(v)
}

func (v Version) SafeFormat(p redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(179931)
	if v.Internal == 0 {
		__antithesis_instrumentation__.Notify(179933)
		p.Printf("%d.%d", v.Major, v.Minor)
		return
	} else {
		__antithesis_instrumentation__.Notify(179934)
	}
	__antithesis_instrumentation__.Notify(179932)
	p.Printf("%d.%d-%d", v.Major, v.Minor, v.Internal)
}

func ParseVersion(s string) (Version, error) {
	__antithesis_instrumentation__.Notify(179935)
	var c Version
	dotParts := strings.Split(s, ".")

	if len(dotParts) != 2 {
		__antithesis_instrumentation__.Notify(179940)
		return Version{}, errors.Errorf("invalid version %s", s)
	} else {
		__antithesis_instrumentation__.Notify(179941)
	}
	__antithesis_instrumentation__.Notify(179936)

	parts := append(dotParts[:1], strings.Split(dotParts[1], "-")...)
	if len(parts) == 2 {
		__antithesis_instrumentation__.Notify(179942)
		parts = append(parts, "0")
	} else {
		__antithesis_instrumentation__.Notify(179943)
	}
	__antithesis_instrumentation__.Notify(179937)

	if len(parts) != 3 {
		__antithesis_instrumentation__.Notify(179944)
		return c, errors.Errorf("invalid version %s", s)
	} else {
		__antithesis_instrumentation__.Notify(179945)
	}
	__antithesis_instrumentation__.Notify(179938)

	ints := make([]int64, len(parts))
	for i := range parts {
		__antithesis_instrumentation__.Notify(179946)
		var err error
		if ints[i], err = strconv.ParseInt(parts[i], 10, 32); err != nil {
			__antithesis_instrumentation__.Notify(179947)
			return c, errors.Wrapf(err, "invalid version %s", s)
		} else {
			__antithesis_instrumentation__.Notify(179948)
		}
	}
	__antithesis_instrumentation__.Notify(179939)

	c.Major = int32(ints[0])
	c.Minor = int32(ints[1])
	c.Internal = int32(ints[2])

	return c, nil
}

func MustParseVersion(s string) Version {
	__antithesis_instrumentation__.Notify(179949)
	v, err := ParseVersion(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(179951)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(179952)
	}
	__antithesis_instrumentation__.Notify(179950)
	return v
}
