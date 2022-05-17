package kv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"reflect"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

func marshalKey(k interface{}) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(128111)
	switch t := k.(type) {
	case *roachpb.Key:
		__antithesis_instrumentation__.Notify(128113)
		return *t, nil
	case roachpb.Key:
		__antithesis_instrumentation__.Notify(128114)
		return t, nil
	case *roachpb.RKey:
		__antithesis_instrumentation__.Notify(128115)
		return t.AsRawKey(), nil
	case roachpb.RKey:
		__antithesis_instrumentation__.Notify(128116)
		return t.AsRawKey(), nil
	case string:
		__antithesis_instrumentation__.Notify(128117)
		return roachpb.Key(t), nil
	case []byte:
		__antithesis_instrumentation__.Notify(128118)
		return roachpb.Key(t), nil
	}
	__antithesis_instrumentation__.Notify(128112)
	return nil, fmt.Errorf("unable to marshal key: %T %q", k, k)
}

func marshalValue(v interface{}) (roachpb.Value, error) {
	__antithesis_instrumentation__.Notify(128119)
	var r roachpb.Value

	switch t := v.(type) {
	case *roachpb.Value:
		__antithesis_instrumentation__.Notify(128122)
		if err := t.VerifyHeader(); err != nil {
			__antithesis_instrumentation__.Notify(128134)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(128135)
		}
		__antithesis_instrumentation__.Notify(128123)
		return *t, nil

	case nil:
		__antithesis_instrumentation__.Notify(128124)
		return r, nil

	case bool:
		__antithesis_instrumentation__.Notify(128125)
		r.SetBool(t)
		return r, nil

	case string:
		__antithesis_instrumentation__.Notify(128126)
		r.SetBytes([]byte(t))
		return r, nil

	case []byte:
		__antithesis_instrumentation__.Notify(128127)
		r.SetBytes(t)
		return r, nil

	case apd.Decimal:
		__antithesis_instrumentation__.Notify(128128)
		err := r.SetDecimal(&t)
		return r, err

	case roachpb.Key:
		__antithesis_instrumentation__.Notify(128129)
		r.SetBytes([]byte(t))
		return r, nil

	case time.Time:
		__antithesis_instrumentation__.Notify(128130)
		r.SetTime(t)
		return r, nil

	case duration.Duration:
		__antithesis_instrumentation__.Notify(128131)
		err := r.SetDuration(t)
		return r, err

	case protoutil.Message:
		__antithesis_instrumentation__.Notify(128132)
		err := r.SetProto(t)
		return r, err

	case roachpb.Value:
		__antithesis_instrumentation__.Notify(128133)
		panic("unexpected type roachpb.Value (use *roachpb.Value)")
	}
	__antithesis_instrumentation__.Notify(128120)

	switch v := reflect.ValueOf(v); v.Kind() {
	case reflect.Bool:
		__antithesis_instrumentation__.Notify(128136)
		r.SetBool(v.Bool())
		return r, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		__antithesis_instrumentation__.Notify(128137)
		r.SetInt(v.Int())
		return r, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		__antithesis_instrumentation__.Notify(128138)
		r.SetInt(int64(v.Uint()))
		return r, nil

	case reflect.Float32, reflect.Float64:
		__antithesis_instrumentation__.Notify(128139)
		r.SetFloat(v.Float())
		return r, nil

	case reflect.String:
		__antithesis_instrumentation__.Notify(128140)
		r.SetBytes([]byte(v.String()))
		return r, nil
	default:
		__antithesis_instrumentation__.Notify(128141)
	}
	__antithesis_instrumentation__.Notify(128121)

	return r, fmt.Errorf("unable to marshal %T: %v", v, v)
}
