package ts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func MakeDataKey(name string, source string, r Resolution, timestamp int64) roachpb.Key {
	__antithesis_instrumentation__.Notify(648033)
	k := makeDataKeySeriesPrefix(name, r)

	timeslot := timestamp / r.SlabDuration()
	k = encoding.EncodeVarintAscending(k, timeslot)
	k = append(k, source...)
	return k
}

func makeDataKeySeriesPrefix(name string, r Resolution) roachpb.Key {
	__antithesis_instrumentation__.Notify(648034)
	k := append(roachpb.Key(nil), keys.TimeseriesPrefix...)
	k = encoding.EncodeBytesAscending(k, []byte(name))
	k = encoding.EncodeVarintAscending(k, int64(r))
	return k
}

func DecodeDataKey(key roachpb.Key) (string, string, Resolution, int64, error) {
	__antithesis_instrumentation__.Notify(648035)

	remainder := key
	if !bytes.HasPrefix(key, keys.TimeseriesPrefix) {
		__antithesis_instrumentation__.Notify(648037)
		return "", "", 0, 0, errors.Errorf("malformed time series data key %v: improper prefix", key)
	} else {
		__antithesis_instrumentation__.Notify(648038)
	}
	__antithesis_instrumentation__.Notify(648036)
	remainder = remainder[len(keys.TimeseriesPrefix):]

	return decodeDataKeySuffix(remainder)
}

func decodeDataKeySuffix(key roachpb.Key) (string, string, Resolution, int64, error) {
	__antithesis_instrumentation__.Notify(648039)

	remainder, name, err := encoding.DecodeBytesAscending(key, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(648043)
		return "", "", 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(648044)
	}
	__antithesis_instrumentation__.Notify(648040)

	remainder, resolutionInt, err := encoding.DecodeVarintAscending(remainder)
	if err != nil {
		__antithesis_instrumentation__.Notify(648045)
		return "", "", 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(648046)
	}
	__antithesis_instrumentation__.Notify(648041)
	resolution := Resolution(resolutionInt)

	remainder, timeslot, err := encoding.DecodeVarintAscending(remainder)
	if err != nil {
		__antithesis_instrumentation__.Notify(648047)
		return "", "", 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(648048)
	}
	__antithesis_instrumentation__.Notify(648042)
	timestamp := timeslot * resolution.SlabDuration()

	source := remainder

	return string(name), string(source), resolution, timestamp, nil
}

func prettyPrintKey(key roachpb.Key) string {
	__antithesis_instrumentation__.Notify(648049)
	name, source, resolution, timestamp, err := decodeDataKeySuffix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(648051)

		return encoding.PrettyPrintValue(nil, key, "/")
	} else {
		__antithesis_instrumentation__.Notify(648052)
	}
	__antithesis_instrumentation__.Notify(648050)
	return fmt.Sprintf("/%s/%s/%s/%s", name, source, resolution,
		timeutil.Unix(0, timestamp).Format(time.RFC3339Nano))
}

func init() {
	keys.PrettyPrintTimeseriesKey = prettyPrintKey
}
