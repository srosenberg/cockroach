package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
)

type EnumSetting struct {
	IntSetting
	enumValues map[int64]string
}

var _ numericSetting = &EnumSetting{}

func (e *EnumSetting) Typ() string {
	__antithesis_instrumentation__.Notify(239874)
	return "e"
}

func (e *EnumSetting) String(sv *Values) string {
	__antithesis_instrumentation__.Notify(239875)
	enumID := e.Get(sv)
	if str, ok := e.enumValues[enumID]; ok {
		__antithesis_instrumentation__.Notify(239877)
		return str
	} else {
		__antithesis_instrumentation__.Notify(239878)
	}
	__antithesis_instrumentation__.Notify(239876)
	return fmt.Sprintf("unknown(%d)", enumID)
}

func (e *EnumSetting) DecodeToString(encoded string) (string, error) {
	__antithesis_instrumentation__.Notify(239879)
	v, err := e.DecodeValue(encoded)
	if err != nil {
		__antithesis_instrumentation__.Notify(239882)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(239883)
	}
	__antithesis_instrumentation__.Notify(239880)
	if str, ok := e.enumValues[v]; ok {
		__antithesis_instrumentation__.Notify(239884)
		return str, nil
	} else {
		__antithesis_instrumentation__.Notify(239885)
	}
	__antithesis_instrumentation__.Notify(239881)
	return encoded, nil
}

func (e *EnumSetting) ParseEnum(raw string) (int64, bool) {
	__antithesis_instrumentation__.Notify(239886)
	rawLower := strings.ToLower(raw)
	for k, v := range e.enumValues {
		__antithesis_instrumentation__.Notify(239889)
		if v == rawLower {
			__antithesis_instrumentation__.Notify(239890)
			return k, true
		} else {
			__antithesis_instrumentation__.Notify(239891)
		}
	}
	__antithesis_instrumentation__.Notify(239887)

	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(239892)
		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(239893)
	}
	__antithesis_instrumentation__.Notify(239888)
	_, ok := e.enumValues[v]
	return v, ok
}

func (e *EnumSetting) GetAvailableValuesAsHint() string {
	__antithesis_instrumentation__.Notify(239894)

	valIdxs := make([]int, 0, len(e.enumValues))
	for i := range e.enumValues {
		__antithesis_instrumentation__.Notify(239897)
		valIdxs = append(valIdxs, int(i))
	}
	__antithesis_instrumentation__.Notify(239895)
	sort.Ints(valIdxs)

	vals := make([]string, 0, len(e.enumValues))
	for _, enumIdx := range valIdxs {
		__antithesis_instrumentation__.Notify(239898)
		vals = append(vals, fmt.Sprintf("%d: %s", enumIdx, e.enumValues[int64(enumIdx)]))
	}
	__antithesis_instrumentation__.Notify(239896)
	return "Available values: " + strings.Join(vals, ", ")
}

func (e *EnumSetting) set(ctx context.Context, sv *Values, k int64) error {
	__antithesis_instrumentation__.Notify(239899)
	if _, ok := e.enumValues[k]; !ok {
		__antithesis_instrumentation__.Notify(239901)
		return errors.Errorf("unrecognized value %d", k)
	} else {
		__antithesis_instrumentation__.Notify(239902)
	}
	__antithesis_instrumentation__.Notify(239900)
	return e.IntSetting.set(ctx, sv, k)
}

func enumValuesToDesc(enumValues map[int64]string) string {
	__antithesis_instrumentation__.Notify(239903)
	var buffer bytes.Buffer
	values := make([]int64, 0, len(enumValues))
	for k := range enumValues {
		__antithesis_instrumentation__.Notify(239907)
		values = append(values, k)
	}
	__antithesis_instrumentation__.Notify(239904)
	sort.Slice(values, func(i, j int) bool { __antithesis_instrumentation__.Notify(239908); return values[i] < values[j] })
	__antithesis_instrumentation__.Notify(239905)

	buffer.WriteString("[")
	for i, k := range values {
		__antithesis_instrumentation__.Notify(239909)
		if i > 0 {
			__antithesis_instrumentation__.Notify(239911)
			buffer.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(239912)
		}
		__antithesis_instrumentation__.Notify(239910)
		fmt.Fprintf(&buffer, "%s = %d", strings.ToLower(enumValues[k]), k)
	}
	__antithesis_instrumentation__.Notify(239906)
	buffer.WriteString("]")
	return buffer.String()
}

func (e *EnumSetting) WithPublic() *EnumSetting {
	__antithesis_instrumentation__.Notify(239913)
	e.SetVisibility(Public)
	return e
}

func RegisterEnumSetting(
	class Class, key, desc string, defaultValue string, enumValues map[int64]string,
) *EnumSetting {
	__antithesis_instrumentation__.Notify(239914)
	enumValuesLower := make(map[int64]string)
	var i int64
	var found bool
	for k, v := range enumValues {
		__antithesis_instrumentation__.Notify(239917)
		enumValuesLower[k] = strings.ToLower(v)
		if v == defaultValue {
			__antithesis_instrumentation__.Notify(239918)
			i = k
			found = true
		} else {
			__antithesis_instrumentation__.Notify(239919)
		}
	}
	__antithesis_instrumentation__.Notify(239915)

	if !found {
		__antithesis_instrumentation__.Notify(239920)
		panic(fmt.Sprintf("enum registered with default value %s not in map %s", defaultValue, enumValuesToDesc(enumValuesLower)))
	} else {
		__antithesis_instrumentation__.Notify(239921)
	}
	__antithesis_instrumentation__.Notify(239916)

	setting := &EnumSetting{
		IntSetting: IntSetting{defaultValue: i},
		enumValues: enumValuesLower,
	}

	register(class, key, fmt.Sprintf("%s %s", desc, enumValuesToDesc(enumValues)), setting)
	return setting
}
