package pgwirebase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/lex"
	"github.com/cockroachdb/cockroach/pkg/sql/oidext"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bitarray"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/ipaddr"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/cockroach/pkg/util/uint128"
	"github.com/cockroachdb/errors"
	"github.com/dustin/go-humanize"
	"github.com/jackc/pgtype"
	"github.com/lib/pq/oid"
)

const (
	defaultMaxReadBufferMessageSize = 1 << 24
	minReadBufferMessageSize        = 1 << 14
)

const readBufferMaxMessageSizeClusterSettingName = "sql.conn.max_read_buffer_message_size"

var ReadBufferMaxMessageSizeClusterSetting = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	readBufferMaxMessageSizeClusterSettingName,
	"maximum buffer size to allow for ingesting sql statements. Connections must be restarted for this to take effect.",
	defaultMaxReadBufferMessageSize,
	func(val int64) error {
		__antithesis_instrumentation__.Notify(560917)
		if val < minReadBufferMessageSize {
			__antithesis_instrumentation__.Notify(560919)
			return errors.Newf("buffer message size must be at least %s", humanize.Bytes(minReadBufferMessageSize))
		} else {
			__antithesis_instrumentation__.Notify(560920)
		}
		__antithesis_instrumentation__.Notify(560918)
		return nil
	},
)

type FormatCode uint16

const (
	FormatText FormatCode = 0

	FormatBinary FormatCode = 1
)

var _ BufferedReader = &bufio.Reader{}
var _ BufferedReader = &bytes.Buffer{}

type BufferedReader interface {
	io.Reader
	ReadString(delim byte) (string, error)
	ReadByte() (byte, error)
}

type ReadBuffer struct {
	Msg            []byte
	tmp            [4]byte
	maxMessageSize int
}

type ReadBufferOption func(*ReadBuffer)

func ReadBufferOptionWithClusterSettings(sv *settings.Values) ReadBufferOption {
	__antithesis_instrumentation__.Notify(560921)
	return func(b *ReadBuffer) {
		__antithesis_instrumentation__.Notify(560922)
		if sv != nil {
			__antithesis_instrumentation__.Notify(560923)
			b.maxMessageSize = int(ReadBufferMaxMessageSizeClusterSetting.Get(sv))
		} else {
			__antithesis_instrumentation__.Notify(560924)
		}
	}
}

func MakeReadBuffer(opts ...ReadBufferOption) ReadBuffer {
	__antithesis_instrumentation__.Notify(560925)
	buf := ReadBuffer{
		maxMessageSize: defaultMaxReadBufferMessageSize,
	}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(560927)
		opt(&buf)
	}
	__antithesis_instrumentation__.Notify(560926)
	return buf
}

func (b *ReadBuffer) reset(size int) {
	__antithesis_instrumentation__.Notify(560928)
	if b.Msg != nil {
		__antithesis_instrumentation__.Notify(560932)
		b.Msg = b.Msg[len(b.Msg):]
	} else {
		__antithesis_instrumentation__.Notify(560933)
	}
	__antithesis_instrumentation__.Notify(560929)

	if cap(b.Msg) >= size {
		__antithesis_instrumentation__.Notify(560934)
		b.Msg = b.Msg[:size]
		return
	} else {
		__antithesis_instrumentation__.Notify(560935)
	}
	__antithesis_instrumentation__.Notify(560930)

	allocSize := size
	if allocSize < 4096 {
		__antithesis_instrumentation__.Notify(560936)
		allocSize = 4096
	} else {
		__antithesis_instrumentation__.Notify(560937)
	}
	__antithesis_instrumentation__.Notify(560931)
	b.Msg = make([]byte, size, allocSize)
}

func (b *ReadBuffer) ReadUntypedMsg(rd io.Reader) (int, error) {
	__antithesis_instrumentation__.Notify(560938)
	nread, err := io.ReadFull(rd, b.tmp[:])
	if err != nil {
		__antithesis_instrumentation__.Notify(560941)
		return nread, err
	} else {
		__antithesis_instrumentation__.Notify(560942)
	}
	__antithesis_instrumentation__.Notify(560939)
	size := int(binary.BigEndian.Uint32(b.tmp[:]))

	size -= 4
	if size > b.maxMessageSize || func() bool {
		__antithesis_instrumentation__.Notify(560943)
		return size < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(560944)
		err := errors.WithHintf(
			NewProtocolViolationErrorf(
				"message size %s bigger than maximum allowed message size %s",
				humanize.IBytes(uint64(size)),
				humanize.IBytes(uint64(b.maxMessageSize)),
			),
			"the maximum message size can be configured using the %s cluster setting",
			readBufferMaxMessageSizeClusterSettingName,
		)
		if size > 0 {
			__antithesis_instrumentation__.Notify(560946)
			err = withMessageTooBigError(err, size)
		} else {
			__antithesis_instrumentation__.Notify(560947)
		}
		__antithesis_instrumentation__.Notify(560945)
		return nread, err
	} else {
		__antithesis_instrumentation__.Notify(560948)
	}
	__antithesis_instrumentation__.Notify(560940)

	b.reset(size)
	n, err := io.ReadFull(rd, b.Msg)
	return nread + n, err
}

func (b *ReadBuffer) SlurpBytes(rd io.Reader, n int) (int, error) {
	__antithesis_instrumentation__.Notify(560949)
	var nRead int
	if b.maxMessageSize > 0 {
		__antithesis_instrumentation__.Notify(560951)
		sizeRemaining := n
		for sizeRemaining > 0 {
			__antithesis_instrumentation__.Notify(560952)
			toRead := sizeRemaining
			if b.maxMessageSize < sizeRemaining {
				__antithesis_instrumentation__.Notify(560954)
				toRead = b.maxMessageSize
			} else {
				__antithesis_instrumentation__.Notify(560955)
			}
			__antithesis_instrumentation__.Notify(560953)
			b.reset(toRead)
			readBatch, err := io.ReadFull(rd, b.Msg)
			nRead += readBatch
			sizeRemaining -= readBatch
			if err != nil {
				__antithesis_instrumentation__.Notify(560956)
				return nRead, err
			} else {
				__antithesis_instrumentation__.Notify(560957)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(560958)
	}
	__antithesis_instrumentation__.Notify(560950)
	return nRead, nil
}

func (b *ReadBuffer) ReadTypedMsg(rd BufferedReader) (ClientMessageType, int, error) {
	__antithesis_instrumentation__.Notify(560959)
	typ, err := rd.ReadByte()
	if err != nil {
		__antithesis_instrumentation__.Notify(560961)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(560962)
	}
	__antithesis_instrumentation__.Notify(560960)
	n, err := b.ReadUntypedMsg(rd)
	return ClientMessageType(typ), n, err
}

func (b *ReadBuffer) GetString() (string, error) {
	__antithesis_instrumentation__.Notify(560963)
	pos := bytes.IndexByte(b.Msg, 0)
	if pos == -1 {
		__antithesis_instrumentation__.Notify(560965)
		return "", NewProtocolViolationErrorf("NUL terminator not found")
	} else {
		__antithesis_instrumentation__.Notify(560966)
	}
	__antithesis_instrumentation__.Notify(560964)

	s := b.Msg[:pos]
	b.Msg = b.Msg[pos+1:]
	return *((*string)(unsafe.Pointer(&s))), nil
}

func (b *ReadBuffer) GetPrepareType() (PrepareType, error) {
	__antithesis_instrumentation__.Notify(560967)
	v, err := b.GetBytes(1)
	if err != nil {
		__antithesis_instrumentation__.Notify(560969)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(560970)
	}
	__antithesis_instrumentation__.Notify(560968)
	return PrepareType(v[0]), nil
}

func (b *ReadBuffer) GetBytes(n int) ([]byte, error) {
	__antithesis_instrumentation__.Notify(560971)
	if len(b.Msg) < n {
		__antithesis_instrumentation__.Notify(560973)
		return nil, NewProtocolViolationErrorf("insufficient data: %d", len(b.Msg))
	} else {
		__antithesis_instrumentation__.Notify(560974)
	}
	__antithesis_instrumentation__.Notify(560972)
	v := b.Msg[:n]
	b.Msg = b.Msg[n:]
	return v, nil
}

func (b *ReadBuffer) GetUint16() (uint16, error) {
	__antithesis_instrumentation__.Notify(560975)
	if len(b.Msg) < 2 {
		__antithesis_instrumentation__.Notify(560977)
		return 0, NewProtocolViolationErrorf("insufficient data: %d", len(b.Msg))
	} else {
		__antithesis_instrumentation__.Notify(560978)
	}
	__antithesis_instrumentation__.Notify(560976)
	v := binary.BigEndian.Uint16(b.Msg[:2])
	b.Msg = b.Msg[2:]
	return v, nil
}

func (b *ReadBuffer) GetUint32() (uint32, error) {
	__antithesis_instrumentation__.Notify(560979)
	if len(b.Msg) < 4 {
		__antithesis_instrumentation__.Notify(560981)
		return 0, NewProtocolViolationErrorf("insufficient data: %d", len(b.Msg))
	} else {
		__antithesis_instrumentation__.Notify(560982)
	}
	__antithesis_instrumentation__.Notify(560980)
	v := binary.BigEndian.Uint32(b.Msg[:4])
	b.Msg = b.Msg[4:]
	return v, nil
}

func (b *ReadBuffer) GetUint64() (uint64, error) {
	__antithesis_instrumentation__.Notify(560983)
	if len(b.Msg) < 8 {
		__antithesis_instrumentation__.Notify(560985)
		return 0, NewProtocolViolationErrorf("insufficient data: %d", len(b.Msg))
	} else {
		__antithesis_instrumentation__.Notify(560986)
	}
	__antithesis_instrumentation__.Notify(560984)
	v := binary.BigEndian.Uint64(b.Msg[:8])
	b.Msg = b.Msg[8:]
	return v, nil
}

func NewUnrecognizedMsgTypeErr(typ ClientMessageType) error {
	__antithesis_instrumentation__.Notify(560987)
	return NewProtocolViolationErrorf("unrecognized client message type %v", typ)
}

func NewProtocolViolationErrorf(format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(560988)
	return pgerror.Newf(pgcode.ProtocolViolation, format, args...)
}

func NewInvalidBinaryRepresentationErrorf(format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(560989)
	return pgerror.Newf(pgcode.InvalidBinaryRepresentation, format, args...)
}

func validateArrayDimensions(nDimensions int, nElements int) error {
	__antithesis_instrumentation__.Notify(560990)
	switch nDimensions {
	case 1:
		__antithesis_instrumentation__.Notify(560992)
		break
	case 0:
		__antithesis_instrumentation__.Notify(560993)

		if nElements == 0 {
			__antithesis_instrumentation__.Notify(560996)
			break
		} else {
			__antithesis_instrumentation__.Notify(560997)
		}
		__antithesis_instrumentation__.Notify(560994)
		fallthrough
	default:
		__antithesis_instrumentation__.Notify(560995)
		return unimplemented.NewWithIssuef(32552,
			"%d-dimension arrays not supported; only 1-dimension", nDimensions)
	}
	__antithesis_instrumentation__.Notify(560991)
	return nil
}

func DecodeDatum(
	evalCtx *tree.EvalContext, t *types.T, code FormatCode, b []byte,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(560998)
	id := t.Oid()
	switch code {
	case FormatText:
		__antithesis_instrumentation__.Notify(561002)
		switch id {
		case oid.T_record:
			__antithesis_instrumentation__.Notify(561006)
			d, _, err := tree.ParseDTupleFromString(evalCtx, string(b), t)
			if err != nil {
				__antithesis_instrumentation__.Notify(561058)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561059)
			}
			__antithesis_instrumentation__.Notify(561007)
			return d, nil
		case oid.T_bool:
			__antithesis_instrumentation__.Notify(561008)
			t, err := strconv.ParseBool(string(b))
			if err != nil {
				__antithesis_instrumentation__.Notify(561060)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561061)
			}
			__antithesis_instrumentation__.Notify(561009)
			return tree.MakeDBool(tree.DBool(t)), nil
		case oid.T_bit, oid.T_varbit:
			__antithesis_instrumentation__.Notify(561010)
			t, err := tree.ParseDBitArray(string(b))
			if err != nil {
				__antithesis_instrumentation__.Notify(561062)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561063)
			}
			__antithesis_instrumentation__.Notify(561011)
			return t, nil
		case oid.T_int2, oid.T_int4, oid.T_int8:
			__antithesis_instrumentation__.Notify(561012)
			i, err := strconv.ParseInt(string(b), 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(561064)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561065)
			}
			__antithesis_instrumentation__.Notify(561013)
			return tree.NewDInt(tree.DInt(i)), nil
		case oid.T_oid,
			oid.T_regoper,
			oid.T_regproc,
			oid.T_regrole,
			oid.T_regclass,
			oid.T_regtype,
			oid.T_regconfig,
			oid.T_regoperator,
			oid.T_regnamespace,
			oid.T_regprocedure,
			oid.T_regdictionary:
			__antithesis_instrumentation__.Notify(561014)
			return tree.ParseDOid(evalCtx, string(b), t)
		case oid.T_float4, oid.T_float8:
			__antithesis_instrumentation__.Notify(561015)
			f, err := strconv.ParseFloat(string(b), 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(561066)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561067)
			}
			__antithesis_instrumentation__.Notify(561016)
			return tree.NewDFloat(tree.DFloat(f)), nil
		case oidext.T_box2d:
			__antithesis_instrumentation__.Notify(561017)
			d, err := tree.ParseDBox2D(string(b))
			if err != nil {
				__antithesis_instrumentation__.Notify(561068)
				return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as box2d", b)
			} else {
				__antithesis_instrumentation__.Notify(561069)
			}
			__antithesis_instrumentation__.Notify(561018)
			return d, nil
		case oidext.T_geography:
			__antithesis_instrumentation__.Notify(561019)
			d, err := tree.ParseDGeography(string(b))
			if err != nil {
				__antithesis_instrumentation__.Notify(561070)
				return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as geography", b)
			} else {
				__antithesis_instrumentation__.Notify(561071)
			}
			__antithesis_instrumentation__.Notify(561020)
			return d, nil
		case oidext.T_geometry:
			__antithesis_instrumentation__.Notify(561021)
			d, err := tree.ParseDGeometry(string(b))
			if err != nil {
				__antithesis_instrumentation__.Notify(561072)
				return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as geometry", b)
			} else {
				__antithesis_instrumentation__.Notify(561073)
			}
			__antithesis_instrumentation__.Notify(561022)
			return d, nil
		case oid.T_void:
			__antithesis_instrumentation__.Notify(561023)
			return tree.DVoidDatum, nil
		case oid.T_numeric:
			__antithesis_instrumentation__.Notify(561024)
			d, err := tree.ParseDDecimal(string(b))
			if err != nil {
				__antithesis_instrumentation__.Notify(561074)
				return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as decimal", b)
			} else {
				__antithesis_instrumentation__.Notify(561075)
			}
			__antithesis_instrumentation__.Notify(561025)
			return d, nil
		case oid.T_bytea:
			__antithesis_instrumentation__.Notify(561026)
			res, err := lex.DecodeRawBytesToByteArrayAuto(b)
			if err != nil {
				__antithesis_instrumentation__.Notify(561076)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561077)
			}
			__antithesis_instrumentation__.Notify(561027)
			return tree.NewDBytes(tree.DBytes(res)), nil
		case oid.T_timestamp:
			__antithesis_instrumentation__.Notify(561028)
			d, _, err := tree.ParseDTimestamp(evalCtx, string(b), time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(561078)
				return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as timestamp", b)
			} else {
				__antithesis_instrumentation__.Notify(561079)
			}
			__antithesis_instrumentation__.Notify(561029)
			return d, nil
		case oid.T_timestamptz:
			__antithesis_instrumentation__.Notify(561030)
			d, _, err := tree.ParseDTimestampTZ(evalCtx, string(b), time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(561080)
				return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as timestamptz", b)
			} else {
				__antithesis_instrumentation__.Notify(561081)
			}
			__antithesis_instrumentation__.Notify(561031)
			return d, nil
		case oid.T_date:
			__antithesis_instrumentation__.Notify(561032)
			d, _, err := tree.ParseDDate(evalCtx, string(b))
			if err != nil {
				__antithesis_instrumentation__.Notify(561082)
				return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as date", b)
			} else {
				__antithesis_instrumentation__.Notify(561083)
			}
			__antithesis_instrumentation__.Notify(561033)
			return d, nil
		case oid.T_time:
			__antithesis_instrumentation__.Notify(561034)
			d, _, err := tree.ParseDTime(nil, string(b), time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(561084)
				return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as time", b)
			} else {
				__antithesis_instrumentation__.Notify(561085)
			}
			__antithesis_instrumentation__.Notify(561035)
			return d, nil
		case oid.T_timetz:
			__antithesis_instrumentation__.Notify(561036)
			d, _, err := tree.ParseDTimeTZ(evalCtx, string(b), time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(561086)
				return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as timetz", b)
			} else {
				__antithesis_instrumentation__.Notify(561087)
			}
			__antithesis_instrumentation__.Notify(561037)
			return d, nil
		case oid.T_interval:
			__antithesis_instrumentation__.Notify(561038)
			d, err := tree.ParseDInterval(evalCtx.GetIntervalStyle(), string(b))
			if err != nil {
				__antithesis_instrumentation__.Notify(561088)
				return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as interval", b)
			} else {
				__antithesis_instrumentation__.Notify(561089)
			}
			__antithesis_instrumentation__.Notify(561039)
			return d, nil
		case oid.T_uuid:
			__antithesis_instrumentation__.Notify(561040)
			d, err := tree.ParseDUuidFromString(string(b))
			if err != nil {
				__antithesis_instrumentation__.Notify(561090)
				return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as uuid", b)
			} else {
				__antithesis_instrumentation__.Notify(561091)
			}
			__antithesis_instrumentation__.Notify(561041)
			return d, nil
		case oid.T_inet:
			__antithesis_instrumentation__.Notify(561042)
			d, err := tree.ParseDIPAddrFromINetString(string(b))
			if err != nil {
				__antithesis_instrumentation__.Notify(561092)
				return nil, pgerror.Newf(pgcode.Syntax,
					"could not parse string %q as inet", b)
			} else {
				__antithesis_instrumentation__.Notify(561093)
			}
			__antithesis_instrumentation__.Notify(561043)
			return d, nil
		case oid.T__int2, oid.T__int4, oid.T__int8:
			__antithesis_instrumentation__.Notify(561044)
			var arr pgtype.Int8Array
			if err := arr.DecodeText(nil, b); err != nil {
				__antithesis_instrumentation__.Notify(561094)
				return nil, pgerror.Wrapf(err, pgcode.Syntax,
					"could not parse string %q as int array", b)
			} else {
				__antithesis_instrumentation__.Notify(561095)
			}
			__antithesis_instrumentation__.Notify(561045)
			if arr.Status != pgtype.Present {
				__antithesis_instrumentation__.Notify(561096)
				return tree.DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(561097)
			}
			__antithesis_instrumentation__.Notify(561046)
			if err := validateArrayDimensions(len(arr.Dimensions), len(arr.Elements)); err != nil {
				__antithesis_instrumentation__.Notify(561098)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561099)
			}
			__antithesis_instrumentation__.Notify(561047)
			out := tree.NewDArray(types.Int)
			var d tree.Datum
			for _, v := range arr.Elements {
				__antithesis_instrumentation__.Notify(561100)
				if v.Status != pgtype.Present {
					__antithesis_instrumentation__.Notify(561102)
					d = tree.DNull
				} else {
					__antithesis_instrumentation__.Notify(561103)
					d = tree.NewDInt(tree.DInt(v.Int))
				}
				__antithesis_instrumentation__.Notify(561101)
				if err := out.Append(d); err != nil {
					__antithesis_instrumentation__.Notify(561104)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(561105)
				}
			}
			__antithesis_instrumentation__.Notify(561048)
			return out, nil
		case oid.T__text, oid.T__name:
			__antithesis_instrumentation__.Notify(561049)
			var arr pgtype.TextArray
			if err := arr.DecodeText(nil, b); err != nil {
				__antithesis_instrumentation__.Notify(561106)
				return nil, pgerror.Wrapf(err, pgcode.Syntax,
					"could not parse string %q as text array", b)
			} else {
				__antithesis_instrumentation__.Notify(561107)
			}
			__antithesis_instrumentation__.Notify(561050)
			if arr.Status != pgtype.Present {
				__antithesis_instrumentation__.Notify(561108)
				return tree.DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(561109)
			}
			__antithesis_instrumentation__.Notify(561051)
			if err := validateArrayDimensions(len(arr.Dimensions), len(arr.Elements)); err != nil {
				__antithesis_instrumentation__.Notify(561110)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561111)
			}
			__antithesis_instrumentation__.Notify(561052)
			out := tree.NewDArray(types.String)
			if id == oid.T__name {
				__antithesis_instrumentation__.Notify(561112)
				out.ParamTyp = types.Name
			} else {
				__antithesis_instrumentation__.Notify(561113)
			}
			__antithesis_instrumentation__.Notify(561053)
			var d tree.Datum
			for _, v := range arr.Elements {
				__antithesis_instrumentation__.Notify(561114)
				if v.Status != pgtype.Present {
					__antithesis_instrumentation__.Notify(561116)
					d = tree.DNull
				} else {
					__antithesis_instrumentation__.Notify(561117)
					d = tree.NewDString(v.String)
					if id == oid.T__name {
						__antithesis_instrumentation__.Notify(561118)
						d = tree.NewDNameFromDString(d.(*tree.DString))
					} else {
						__antithesis_instrumentation__.Notify(561119)
					}
				}
				__antithesis_instrumentation__.Notify(561115)
				if err := out.Append(d); err != nil {
					__antithesis_instrumentation__.Notify(561120)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(561121)
				}
			}
			__antithesis_instrumentation__.Notify(561054)
			return out, nil
		case oid.T_jsonb:
			__antithesis_instrumentation__.Notify(561055)
			if err := validateStringBytes(b); err != nil {
				__antithesis_instrumentation__.Notify(561122)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561123)
			}
			__antithesis_instrumentation__.Notify(561056)
			return tree.ParseDJSON(string(b))
		default:
			__antithesis_instrumentation__.Notify(561057)
		}
		__antithesis_instrumentation__.Notify(561003)
		if t.Family() == types.ArrayFamily {
			__antithesis_instrumentation__.Notify(561124)

			if err := validateStringBytes(b); err != nil {
				__antithesis_instrumentation__.Notify(561126)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561127)
			}
			__antithesis_instrumentation__.Notify(561125)
			return tree.NewDString(string(b)), nil
		} else {
			__antithesis_instrumentation__.Notify(561128)
		}
	case FormatBinary:
		__antithesis_instrumentation__.Notify(561004)
		switch id {
		case oid.T_record:
			__antithesis_instrumentation__.Notify(561129)
			return decodeBinaryTuple(evalCtx, b)
		case oid.T_bool:
			__antithesis_instrumentation__.Notify(561130)
			if len(b) > 0 {
				__antithesis_instrumentation__.Notify(561176)
				switch b[0] {
				case 0:
					__antithesis_instrumentation__.Notify(561177)
					return tree.MakeDBool(false), nil
				case 1:
					__antithesis_instrumentation__.Notify(561178)
					return tree.MakeDBool(true), nil
				default:
					__antithesis_instrumentation__.Notify(561179)
				}
			} else {
				__antithesis_instrumentation__.Notify(561180)
			}
			__antithesis_instrumentation__.Notify(561131)
			return nil, pgerror.Newf(pgcode.Syntax, "unsupported binary bool: %x", b)
		case oid.T_int2:
			__antithesis_instrumentation__.Notify(561132)
			if len(b) < 2 {
				__antithesis_instrumentation__.Notify(561181)
				return nil, pgerror.Newf(pgcode.Syntax, "int2 requires 2 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561182)
			}
			__antithesis_instrumentation__.Notify(561133)
			i := int16(binary.BigEndian.Uint16(b))
			return tree.NewDInt(tree.DInt(i)), nil
		case oid.T_int4:
			__antithesis_instrumentation__.Notify(561134)
			if len(b) < 4 {
				__antithesis_instrumentation__.Notify(561183)
				return nil, pgerror.Newf(pgcode.Syntax, "int4 requires 4 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561184)
			}
			__antithesis_instrumentation__.Notify(561135)
			i := int32(binary.BigEndian.Uint32(b))
			return tree.NewDInt(tree.DInt(i)), nil
		case oid.T_int8:
			__antithesis_instrumentation__.Notify(561136)
			if len(b) < 8 {
				__antithesis_instrumentation__.Notify(561185)
				return nil, pgerror.Newf(pgcode.Syntax, "int8 requires 8 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561186)
			}
			__antithesis_instrumentation__.Notify(561137)
			i := int64(binary.BigEndian.Uint64(b))
			return tree.NewDInt(tree.DInt(i)), nil
		case oid.T_oid:
			__antithesis_instrumentation__.Notify(561138)
			if len(b) < 4 {
				__antithesis_instrumentation__.Notify(561187)
				return nil, pgerror.Newf(pgcode.Syntax, "oid requires 4 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561188)
			}
			__antithesis_instrumentation__.Notify(561139)
			u := binary.BigEndian.Uint32(b)
			return tree.NewDOid(tree.DInt(u)), nil
		case oid.T_float4:
			__antithesis_instrumentation__.Notify(561140)
			if len(b) < 4 {
				__antithesis_instrumentation__.Notify(561189)
				return nil, pgerror.Newf(pgcode.Syntax, "float4 requires 4 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561190)
			}
			__antithesis_instrumentation__.Notify(561141)
			f := math.Float32frombits(binary.BigEndian.Uint32(b))
			return tree.NewDFloat(tree.DFloat(f)), nil
		case oid.T_float8:
			__antithesis_instrumentation__.Notify(561142)
			if len(b) < 8 {
				__antithesis_instrumentation__.Notify(561191)
				return nil, pgerror.Newf(pgcode.Syntax, "float8 requires 8 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561192)
			}
			__antithesis_instrumentation__.Notify(561143)
			f := math.Float64frombits(binary.BigEndian.Uint64(b))
			return tree.NewDFloat(tree.DFloat(f)), nil
		case oid.T_numeric:
			__antithesis_instrumentation__.Notify(561144)
			r := bytes.NewReader(b)

			alloc := struct {
				pgNum PGNumeric
				i16   int16

				dd tree.DDecimal
			}{}

			for _, ptr := range []interface{}{
				&alloc.pgNum.Ndigits,
				&alloc.pgNum.Weight,
				&alloc.pgNum.Sign,
				&alloc.pgNum.Dscale,
			} {
				__antithesis_instrumentation__.Notify(561193)
				if err := binary.Read(r, binary.BigEndian, ptr); err != nil {
					__antithesis_instrumentation__.Notify(561194)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(561195)
				}
			}
			__antithesis_instrumentation__.Notify(561145)

			if alloc.pgNum.Ndigits > 0 {
				__antithesis_instrumentation__.Notify(561196)
				decDigits := make([]byte, 0, int(alloc.pgNum.Ndigits)*PGDecDigits)
				for i := int16(0); i < alloc.pgNum.Ndigits; i++ {
					__antithesis_instrumentation__.Notify(561200)
					if err := binary.Read(r, binary.BigEndian, &alloc.i16); err != nil {
						__antithesis_instrumentation__.Notify(561204)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(561205)
					}
					__antithesis_instrumentation__.Notify(561201)

					numZeroes := PGDecDigits
					for i16 := alloc.i16; i16 > 0; i16 /= 10 {
						__antithesis_instrumentation__.Notify(561206)
						numZeroes--
					}
					__antithesis_instrumentation__.Notify(561202)
					for ; numZeroes > 0; numZeroes-- {
						__antithesis_instrumentation__.Notify(561207)
						decDigits = append(decDigits, '0')
					}
					__antithesis_instrumentation__.Notify(561203)
					if alloc.i16 > 0 {
						__antithesis_instrumentation__.Notify(561208)
						decDigits = strconv.AppendUint(decDigits, uint64(alloc.i16), 10)
					} else {
						__antithesis_instrumentation__.Notify(561209)
					}
				}
				__antithesis_instrumentation__.Notify(561197)

				dscale := (alloc.pgNum.Ndigits - (alloc.pgNum.Weight + 1)) * PGDecDigits
				if overScale := dscale - alloc.pgNum.Dscale; overScale > 0 {
					__antithesis_instrumentation__.Notify(561210)
					dscale -= overScale
					decDigits = decDigits[:len(decDigits)-int(overScale)]
				} else {
					__antithesis_instrumentation__.Notify(561211)
				}
				__antithesis_instrumentation__.Notify(561198)

				decString := string(decDigits)
				if _, ok := alloc.dd.Coeff.SetString(decString, 10); !ok {
					__antithesis_instrumentation__.Notify(561212)
					return nil, pgerror.Newf(pgcode.Syntax, "could not parse string %q as decimal", decString)
				} else {
					__antithesis_instrumentation__.Notify(561213)
				}
				__antithesis_instrumentation__.Notify(561199)
				alloc.dd.Exponent = -int32(dscale)
			} else {
				__antithesis_instrumentation__.Notify(561214)
			}
			__antithesis_instrumentation__.Notify(561146)

			switch alloc.pgNum.Sign {
			case PGNumericPos:
				__antithesis_instrumentation__.Notify(561215)
			case PGNumericNeg:
				__antithesis_instrumentation__.Notify(561216)
				alloc.dd.Neg(&alloc.dd.Decimal)
			case 0xc000:
				__antithesis_instrumentation__.Notify(561217)

				return tree.ParseDDecimal("NaN")
			case 0xd000:
				__antithesis_instrumentation__.Notify(561218)

				return tree.ParseDDecimal("Inf")

			case 0xf000:
				__antithesis_instrumentation__.Notify(561219)

				return tree.ParseDDecimal("-Inf")
			default:
				__antithesis_instrumentation__.Notify(561220)
				return nil, pgerror.Newf(pgcode.Syntax, "unsupported numeric sign: %d", alloc.pgNum.Sign)
			}
			__antithesis_instrumentation__.Notify(561147)

			return &alloc.dd, nil
		case oid.T_bytea:
			__antithesis_instrumentation__.Notify(561148)
			return tree.NewDBytes(tree.DBytes(b)), nil
		case oid.T_timestamp:
			__antithesis_instrumentation__.Notify(561149)
			if len(b) < 8 {
				__antithesis_instrumentation__.Notify(561221)
				return nil, pgerror.Newf(pgcode.Syntax, "timestamp requires 8 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561222)
			}
			__antithesis_instrumentation__.Notify(561150)
			i := int64(binary.BigEndian.Uint64(b))
			return tree.MakeDTimestamp(pgBinaryToTime(i), time.Microsecond)
		case oid.T_timestamptz:
			__antithesis_instrumentation__.Notify(561151)
			if len(b) < 8 {
				__antithesis_instrumentation__.Notify(561223)
				return nil, pgerror.Newf(pgcode.Syntax, "timestamptz requires 8 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561224)
			}
			__antithesis_instrumentation__.Notify(561152)
			i := int64(binary.BigEndian.Uint64(b))
			return tree.MakeDTimestampTZ(pgBinaryToTime(i), time.Microsecond)
		case oid.T_date:
			__antithesis_instrumentation__.Notify(561153)
			if len(b) < 4 {
				__antithesis_instrumentation__.Notify(561225)
				return nil, pgerror.Newf(pgcode.Syntax, "date requires 4 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561226)
			}
			__antithesis_instrumentation__.Notify(561154)
			i := int32(binary.BigEndian.Uint32(b))
			return pgBinaryToDate(i)
		case oid.T_time:
			__antithesis_instrumentation__.Notify(561155)
			if len(b) < 8 {
				__antithesis_instrumentation__.Notify(561227)
				return nil, pgerror.Newf(pgcode.Syntax, "time requires 8 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561228)
			}
			__antithesis_instrumentation__.Notify(561156)
			i := int64(binary.BigEndian.Uint64(b))
			return tree.MakeDTime(timeofday.TimeOfDay(i)), nil
		case oid.T_timetz:
			__antithesis_instrumentation__.Notify(561157)
			if len(b) < 12 {
				__antithesis_instrumentation__.Notify(561229)
				return nil, pgerror.Newf(pgcode.Syntax, "timetz requires 12 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561230)
			}
			__antithesis_instrumentation__.Notify(561158)
			timeOfDayMicros := int64(binary.BigEndian.Uint64(b))
			offsetSecs := int32(binary.BigEndian.Uint32(b[8:]))
			return tree.NewDTimeTZFromOffset(timeofday.TimeOfDay(timeOfDayMicros), offsetSecs), nil
		case oid.T_interval:
			__antithesis_instrumentation__.Notify(561159)
			if len(b) < 16 {
				__antithesis_instrumentation__.Notify(561231)
				return nil, pgerror.Newf(pgcode.Syntax, "interval requires 16 bytes for binary format")
			} else {
				__antithesis_instrumentation__.Notify(561232)
			}
			__antithesis_instrumentation__.Notify(561160)
			nanos := (int64(binary.BigEndian.Uint64(b)) / int64(time.Nanosecond)) * int64(time.Microsecond)
			days := int32(binary.BigEndian.Uint32(b[8:]))
			months := int32(binary.BigEndian.Uint32(b[12:]))

			duration := duration.MakeDuration(nanos, int64(days), int64(months))
			return &tree.DInterval{Duration: duration}, nil
		case oid.T_uuid:
			__antithesis_instrumentation__.Notify(561161)
			u, err := tree.ParseDUuidFromBytes(b)
			if err != nil {
				__antithesis_instrumentation__.Notify(561233)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561234)
			}
			__antithesis_instrumentation__.Notify(561162)
			return u, nil
		case oid.T_inet:
			__antithesis_instrumentation__.Notify(561163)
			ipAddr, err := pgBinaryToIPAddr(b)
			if err != nil {
				__antithesis_instrumentation__.Notify(561235)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561236)
			}
			__antithesis_instrumentation__.Notify(561164)
			return tree.NewDIPAddr(tree.DIPAddr{IPAddr: ipAddr}), nil
		case oid.T_jsonb:
			__antithesis_instrumentation__.Notify(561165)
			if len(b) < 1 {
				__antithesis_instrumentation__.Notify(561237)
				return nil, NewProtocolViolationErrorf("no data to decode")
			} else {
				__antithesis_instrumentation__.Notify(561238)
			}
			__antithesis_instrumentation__.Notify(561166)
			if b[0] != 1 {
				__antithesis_instrumentation__.Notify(561239)
				return nil, NewProtocolViolationErrorf("expected JSONB version 1")
			} else {
				__antithesis_instrumentation__.Notify(561240)
			}
			__antithesis_instrumentation__.Notify(561167)

			b = b[1:]
			if err := validateStringBytes(b); err != nil {
				__antithesis_instrumentation__.Notify(561241)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561242)
			}
			__antithesis_instrumentation__.Notify(561168)
			return tree.ParseDJSON(string(b))
		case oid.T_varbit, oid.T_bit:
			__antithesis_instrumentation__.Notify(561169)
			if len(b) < 4 {
				__antithesis_instrumentation__.Notify(561243)
				return nil, NewProtocolViolationErrorf("insufficient data: %d", len(b))
			} else {
				__antithesis_instrumentation__.Notify(561244)
			}
			__antithesis_instrumentation__.Notify(561170)
			bitlen := binary.BigEndian.Uint32(b)
			b = b[4:]
			lastBitsUsed := uint64(bitlen % 64)
			if bitlen != 0 && func() bool {
				__antithesis_instrumentation__.Notify(561245)
				return lastBitsUsed == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(561246)
				lastBitsUsed = 64
			} else {
				__antithesis_instrumentation__.Notify(561247)
			}
			__antithesis_instrumentation__.Notify(561171)
			if len(b)*8 < int(bitlen) {
				__antithesis_instrumentation__.Notify(561248)
				return nil, pgerror.Newf(pgcode.Syntax, "unexpected varbit bitlen %d (b: %d)", bitlen, len(b))
			} else {
				__antithesis_instrumentation__.Notify(561249)
			}
			__antithesis_instrumentation__.Notify(561172)
			words := make([]uint64, (len(b)+7)/8)

			for i := 0; i < len(words)-1; i++ {
				__antithesis_instrumentation__.Notify(561250)
				words[i] = binary.BigEndian.Uint64(b)
				b = b[8:]
			}
			__antithesis_instrumentation__.Notify(561173)
			if len(words) > 0 {
				__antithesis_instrumentation__.Notify(561251)
				var w uint64
				i := uint(0)
				for ; i < uint(lastBitsUsed); i += 8 {
					__antithesis_instrumentation__.Notify(561253)
					if len(b) == 0 {
						__antithesis_instrumentation__.Notify(561255)
						return nil, NewInvalidBinaryRepresentationErrorf("incorrect binary data")
					} else {
						__antithesis_instrumentation__.Notify(561256)
					}
					__antithesis_instrumentation__.Notify(561254)
					w = (w << 8) | uint64(b[0])
					b = b[1:]
				}
				__antithesis_instrumentation__.Notify(561252)
				words[len(words)-1] = w << (64 - i)
			} else {
				__antithesis_instrumentation__.Notify(561257)
			}
			__antithesis_instrumentation__.Notify(561174)
			ba, err := bitarray.FromEncodingParts(words, lastBitsUsed)
			return &tree.DBitArray{BitArray: ba}, err
		default:
			__antithesis_instrumentation__.Notify(561175)
			if t.Family() == types.ArrayFamily {
				__antithesis_instrumentation__.Notify(561258)
				return decodeBinaryArray(evalCtx, t.ArrayContents(), b, code)
			} else {
				__antithesis_instrumentation__.Notify(561259)
			}
		}
	default:
		__antithesis_instrumentation__.Notify(561005)
		return nil, errors.AssertionFailedf(
			"unexpected format code: %d", errors.Safe(code))
	}
	__antithesis_instrumentation__.Notify(560999)

	switch t.Family() {
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(561260)
		if err := validateStringBytes(b); err != nil {
			__antithesis_instrumentation__.Notify(561263)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(561264)
		}
		__antithesis_instrumentation__.Notify(561261)
		return tree.MakeDEnumFromLogicalRepresentation(t, string(b))
	default:
		__antithesis_instrumentation__.Notify(561262)
	}
	__antithesis_instrumentation__.Notify(561000)
	switch id {
	case oid.T_text, oid.T_varchar, oid.T_unknown:
		__antithesis_instrumentation__.Notify(561265)
		if err := validateStringBytes(b); err != nil {
			__antithesis_instrumentation__.Notify(561274)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(561275)
		}
		__antithesis_instrumentation__.Notify(561266)
		return tree.NewDString(string(b)), nil
	case oid.T_bpchar:
		__antithesis_instrumentation__.Notify(561267)
		if err := validateStringBytes(b); err != nil {
			__antithesis_instrumentation__.Notify(561276)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(561277)
		}
		__antithesis_instrumentation__.Notify(561268)

		sv := strings.TrimRight(string(b), " ")
		return tree.NewDString(sv), nil
	case oid.T_char:
		__antithesis_instrumentation__.Notify(561269)
		sv := string(b)

		if len(b) >= 1 {
			__antithesis_instrumentation__.Notify(561278)
			if b[0] == 0 {
				__antithesis_instrumentation__.Notify(561279)
				sv = ""
			} else {
				__antithesis_instrumentation__.Notify(561280)
				sv = string(b[:1])
			}
		} else {
			__antithesis_instrumentation__.Notify(561281)
		}
		__antithesis_instrumentation__.Notify(561270)
		return tree.NewDString(sv), nil
	case oid.T_name:
		__antithesis_instrumentation__.Notify(561271)
		if err := validateStringBytes(b); err != nil {
			__antithesis_instrumentation__.Notify(561282)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(561283)
		}
		__antithesis_instrumentation__.Notify(561272)
		return tree.NewDName(string(b)), nil
	default:
		__antithesis_instrumentation__.Notify(561273)
	}
	__antithesis_instrumentation__.Notify(561001)

	return nil, errors.AssertionFailedf(
		"unsupported OID %v with format code %s", errors.Safe(id), errors.Safe(code))
}

func validateStringBytes(b []byte) error {
	__antithesis_instrumentation__.Notify(561284)
	if !utf8.Valid(b) {
		__antithesis_instrumentation__.Notify(561286)
		return invalidUTF8Error
	} else {
		__antithesis_instrumentation__.Notify(561287)
	}
	__antithesis_instrumentation__.Notify(561285)
	return nil
}

type PGNumericSign uint16

const (
	PGNumericPos PGNumericSign = 0x0000

	PGNumericNeg PGNumericSign = 0x4000
)

const PGDecDigits = 4

type PGNumeric struct {
	Ndigits, Weight, Dscale int16
	Sign                    PGNumericSign
}

func pgBinaryToTime(i int64) time.Time {
	__antithesis_instrumentation__.Notify(561288)
	return duration.AddMicros(PGEpochJDate, i)
}

func pgBinaryToDate(i int32) (*tree.DDate, error) {
	__antithesis_instrumentation__.Notify(561289)
	d, err := pgdate.MakeDateFromPGEpoch(i)
	if err != nil {
		__antithesis_instrumentation__.Notify(561291)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(561292)
	}
	__antithesis_instrumentation__.Notify(561290)
	return tree.NewDDate(d), nil
}

func pgBinaryToIPAddr(b []byte) (ipaddr.IPAddr, error) {
	__antithesis_instrumentation__.Notify(561293)
	if len(b) < 4 {
		__antithesis_instrumentation__.Notify(561297)
		return ipaddr.IPAddr{}, NewProtocolViolationErrorf("insufficient data: %d", len(b))
	} else {
		__antithesis_instrumentation__.Notify(561298)
	}
	__antithesis_instrumentation__.Notify(561294)

	mask := b[1]
	familyByte := b[0]
	var addr ipaddr.Addr
	var family ipaddr.IPFamily
	b = b[4:]

	if familyByte == PGBinaryIPv4family {
		__antithesis_instrumentation__.Notify(561299)
		family = ipaddr.IPv4family
	} else {
		__antithesis_instrumentation__.Notify(561300)
		if familyByte == PGBinaryIPv6family {
			__antithesis_instrumentation__.Notify(561301)
			family = ipaddr.IPv6family
		} else {
			__antithesis_instrumentation__.Notify(561302)
			return ipaddr.IPAddr{}, NewInvalidBinaryRepresentationErrorf("unknown family received: %d", familyByte)
		}
	}
	__antithesis_instrumentation__.Notify(561295)

	if family == ipaddr.IPv4family {
		__antithesis_instrumentation__.Notify(561303)
		if len(b) != 4 {
			__antithesis_instrumentation__.Notify(561305)
			return ipaddr.IPAddr{}, NewInvalidBinaryRepresentationErrorf("unexpected data: %d", len(b))
		} else {
			__antithesis_instrumentation__.Notify(561306)
		}
		__antithesis_instrumentation__.Notify(561304)

		var tmp [16]byte
		tmp[10] = 0xff
		tmp[11] = 0xff
		copy(tmp[12:], b)
		addr = ipaddr.Addr(uint128.FromBytes(tmp[:]))
	} else {
		__antithesis_instrumentation__.Notify(561307)
		if len(b) != 16 {
			__antithesis_instrumentation__.Notify(561309)
			return ipaddr.IPAddr{}, NewInvalidBinaryRepresentationErrorf("unexpected data: %d", len(b))
		} else {
			__antithesis_instrumentation__.Notify(561310)
		}
		__antithesis_instrumentation__.Notify(561308)
		addr = ipaddr.Addr(uint128.FromBytes(b))
	}
	__antithesis_instrumentation__.Notify(561296)

	return ipaddr.IPAddr{
		Family: family,
		Mask:   mask,
		Addr:   addr,
	}, nil
}

func decodeBinaryArray(
	evalCtx *tree.EvalContext, t *types.T, b []byte, code FormatCode,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(561311)
	var hdr struct {
		Ndims int32

		_       int32
		ElemOid int32
	}
	var dim struct {
		DimSize int32

		_ int32
	}
	r := bytes.NewBuffer(b)
	if err := binary.Read(r, binary.BigEndian, &hdr); err != nil {
		__antithesis_instrumentation__.Notify(561318)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(561319)
	}
	__antithesis_instrumentation__.Notify(561312)
	if t.Oid() != oid.Oid(hdr.ElemOid) {
		__antithesis_instrumentation__.Notify(561320)
		return nil, pgerror.Newf(pgcode.ProtocolViolation, "wrong element type")
	} else {
		__antithesis_instrumentation__.Notify(561321)
	}
	__antithesis_instrumentation__.Notify(561313)
	arr := tree.NewDArray(t)
	if hdr.Ndims == 0 {
		__antithesis_instrumentation__.Notify(561322)
		return arr, nil
	} else {
		__antithesis_instrumentation__.Notify(561323)
	}
	__antithesis_instrumentation__.Notify(561314)
	if err := binary.Read(r, binary.BigEndian, &dim); err != nil {
		__antithesis_instrumentation__.Notify(561324)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(561325)
	}
	__antithesis_instrumentation__.Notify(561315)
	if err := validateArrayDimensions(int(hdr.Ndims), int(dim.DimSize)); err != nil {
		__antithesis_instrumentation__.Notify(561326)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(561327)
	}
	__antithesis_instrumentation__.Notify(561316)
	var vlen int32
	for i := int32(0); i < dim.DimSize; i++ {
		__antithesis_instrumentation__.Notify(561328)
		if err := binary.Read(r, binary.BigEndian, &vlen); err != nil {
			__antithesis_instrumentation__.Notify(561332)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(561333)
		}
		__antithesis_instrumentation__.Notify(561329)
		if vlen < 0 {
			__antithesis_instrumentation__.Notify(561334)
			if err := arr.Append(tree.DNull); err != nil {
				__antithesis_instrumentation__.Notify(561336)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561337)
			}
			__antithesis_instrumentation__.Notify(561335)
			continue
		} else {
			__antithesis_instrumentation__.Notify(561338)
		}
		__antithesis_instrumentation__.Notify(561330)
		buf := r.Next(int(vlen))
		elem, err := DecodeDatum(evalCtx, t, code, buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(561339)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(561340)
		}
		__antithesis_instrumentation__.Notify(561331)
		if err := arr.Append(elem); err != nil {
			__antithesis_instrumentation__.Notify(561341)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(561342)
		}
	}
	__antithesis_instrumentation__.Notify(561317)
	return arr, nil
}

const tupleHeaderSize, oidSize, elementSize = 4, 4, 4

func decodeBinaryTuple(evalCtx *tree.EvalContext, b []byte) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(561343)

	bufferLength := int32(len(b))
	if bufferLength < tupleHeaderSize {
		__antithesis_instrumentation__.Notify(561348)
		return nil, pgerror.Newf(
			pgcode.Syntax,
			"tuple requires a %d byte header for binary format. bufferLength=%d",
			tupleHeaderSize, bufferLength)
	} else {
		__antithesis_instrumentation__.Notify(561349)
	}
	__antithesis_instrumentation__.Notify(561344)

	bufferStartIdx := int32(0)
	bufferEndIdx := bufferStartIdx + tupleHeaderSize
	numberOfElements := int32(binary.BigEndian.Uint32(b[bufferStartIdx:bufferEndIdx]))
	if numberOfElements < 0 {
		__antithesis_instrumentation__.Notify(561350)
		return nil, pgerror.Newf(
			pgcode.Syntax,
			"tuple must have non-negative number of elements. numberOfElements=%d",
			numberOfElements)
	} else {
		__antithesis_instrumentation__.Notify(561351)
	}
	__antithesis_instrumentation__.Notify(561345)
	bufferStartIdx = bufferEndIdx

	typs := make([]*types.T, numberOfElements)
	datums := make(tree.Datums, numberOfElements)

	elementIdx := int32(0)

	getSyntaxError := func(message string, args ...interface{}) error {
		__antithesis_instrumentation__.Notify(561352)
		formattedMessage := fmt.Sprintf(message, args...)
		return pgerror.Newf(
			pgcode.Syntax,
			"%s elementIdx=%d bufferLength=%d bufferStartIdx=%d bufferEndIdx=%d",
			formattedMessage, elementIdx, bufferLength, bufferStartIdx, bufferEndIdx)
	}
	__antithesis_instrumentation__.Notify(561346)

	for elementIdx < numberOfElements {
		__antithesis_instrumentation__.Notify(561353)

		bytesToRead := int32(oidSize)
		bufferEndIdx = bufferStartIdx + bytesToRead
		if bufferEndIdx < bufferStartIdx {
			__antithesis_instrumentation__.Notify(561360)
			return nil, getSyntaxError("integer overflow reading element OID for binary format. ")
		} else {
			__antithesis_instrumentation__.Notify(561361)
		}
		__antithesis_instrumentation__.Notify(561354)
		if bufferLength < bufferEndIdx {
			__antithesis_instrumentation__.Notify(561362)
			return nil, getSyntaxError("insufficient bytes reading element OID for binary format. ")
		} else {
			__antithesis_instrumentation__.Notify(561363)
		}
		__antithesis_instrumentation__.Notify(561355)

		elementOID := int32(binary.BigEndian.Uint32(b[bufferStartIdx:bufferEndIdx]))
		elementType, ok := types.OidToType[oid.Oid(elementOID)]
		if !ok {
			__antithesis_instrumentation__.Notify(561364)
			return nil, getSyntaxError("element type not found for OID %d. ", elementOID)
		} else {
			__antithesis_instrumentation__.Notify(561365)
		}
		__antithesis_instrumentation__.Notify(561356)
		typs[elementIdx] = elementType
		bufferStartIdx = bufferEndIdx

		bytesToRead = int32(elementSize)
		bufferEndIdx = bufferStartIdx + bytesToRead
		if bufferEndIdx < bufferStartIdx {
			__antithesis_instrumentation__.Notify(561366)
			return nil, getSyntaxError("integer overflow reading element size for binary format. ")
		} else {
			__antithesis_instrumentation__.Notify(561367)
		}
		__antithesis_instrumentation__.Notify(561357)
		if bufferLength < bufferEndIdx {
			__antithesis_instrumentation__.Notify(561368)
			return nil, getSyntaxError("insufficient bytes reading element size for binary format. ")
		} else {
			__antithesis_instrumentation__.Notify(561369)
		}
		__antithesis_instrumentation__.Notify(561358)

		bytesToRead = int32(binary.BigEndian.Uint32(b[bufferStartIdx:bufferEndIdx]))
		bufferStartIdx = bufferEndIdx
		if bytesToRead == -1 {
			__antithesis_instrumentation__.Notify(561370)
			datums[elementIdx] = tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(561371)
			bufferEndIdx = bufferStartIdx + bytesToRead
			if bufferEndIdx < bufferStartIdx {
				__antithesis_instrumentation__.Notify(561375)
				return nil, getSyntaxError("integer overflow reading element for binary format. ")
			} else {
				__antithesis_instrumentation__.Notify(561376)
			}
			__antithesis_instrumentation__.Notify(561372)
			if bufferLength < bufferEndIdx {
				__antithesis_instrumentation__.Notify(561377)
				return nil, getSyntaxError("insufficient bytes reading element for binary format. ")
			} else {
				__antithesis_instrumentation__.Notify(561378)
			}
			__antithesis_instrumentation__.Notify(561373)

			colDatum, err := DecodeDatum(evalCtx, elementType, FormatBinary, b[bufferStartIdx:bufferEndIdx])

			if err != nil {
				__antithesis_instrumentation__.Notify(561379)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(561380)
			}
			__antithesis_instrumentation__.Notify(561374)

			bufferStartIdx = bufferEndIdx
			datums[elementIdx] = colDatum
		}
		__antithesis_instrumentation__.Notify(561359)
		elementIdx++
	}
	__antithesis_instrumentation__.Notify(561347)

	tupleTyps := types.MakeTuple(typs)
	return tree.NewDTuple(tupleTyps, datums...), nil

}

var invalidUTF8Error = pgerror.Newf(pgcode.CharacterNotInRepertoire, "invalid UTF-8 sequence")

var (
	PGEpochJDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
)

const (
	PGBinaryIPv4family byte = 2

	PGBinaryIPv6family byte = 3
)
