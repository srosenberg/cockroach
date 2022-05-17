package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/md5"
	cryptorand "crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	gojson "encoding/json"
	"fmt"
	"hash"
	"hash/crc32"
	"hash/fnv"
	"io/ioutil"
	"math"
	"math/bits"
	"math/rand"
	"net"
	"regexp/syntax"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/lex"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/persistedsqlstats/sqlstatsutil"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/fuzzystrmatch"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/ipaddr"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/cockroach/pkg/util/timetz"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/cockroachdb/cockroach/pkg/util/ulid"
	"github.com/cockroachdb/cockroach/pkg/util/unaccent"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/knz/strtime"
)

var (
	errEmptyInputString = pgerror.New(pgcode.InvalidParameterValue, "the input string must not be empty")
	errZeroIP           = pgerror.New(pgcode.InvalidParameterValue, "zero length IP")
	errChrValueTooSmall = pgerror.New(pgcode.InvalidParameterValue, "input value must be >= 0")
	errChrValueTooLarge = pgerror.Newf(pgcode.InvalidParameterValue,
		"input value must be <= %d (maximum Unicode code point)", utf8.MaxRune)
	errStringTooLarge = pgerror.Newf(pgcode.ProgramLimitExceeded,
		"requested length too large, exceeds %s", humanizeutil.IBytes(maxAllocatedStringSize))
	errInvalidNull = pgerror.New(pgcode.InvalidParameterValue, "input cannot be NULL")

	SequenceNameArg = "sequence_name"
)

const defaultFollowerReadDuration = -4800 * time.Millisecond

const maxAllocatedStringSize = 128 * 1024 * 1024

const errInsufficientArgsFmtString = "unknown signature: %s()"

const (
	categoryArray               = "Array"
	categoryComparison          = "Comparison"
	categoryCompatibility       = "Compatibility"
	categoryCrypto              = "Cryptographic"
	categoryDateAndTime         = "Date and time"
	categoryEnum                = "Enum"
	categoryFullTextSearch      = "Full Text Search"
	categoryGenerator           = "Set-returning"
	categoryTrigram             = "Trigrams"
	categoryFuzzyStringMatching = "Fuzzy String Matching"
	categoryIDGeneration        = "ID generation"
	categoryJSON                = "JSONB"
	categoryMultiRegion         = "Multi-region"
	categoryMultiTenancy        = "Multi-tenancy"
	categorySequences           = "Sequence"
	categorySpatial             = "Spatial"
	categoryString              = "String and byte"
	categorySystemInfo          = "System info"
	categorySystemRepair        = "System repair"
	categoryStreamIngestion     = "Stream Ingestion"
)

func categorizeType(t *types.T) string {
	__antithesis_instrumentation__.Notify(596950)
	switch t.Family() {
	case types.DateFamily, types.IntervalFamily, types.TimestampFamily, types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(596951)
		return categoryDateAndTime
	case types.StringFamily, types.BytesFamily:
		__antithesis_instrumentation__.Notify(596952)
		return categoryString
	default:
		__antithesis_instrumentation__.Notify(596953)
		return strings.ToUpper(t.String())
	}
}

const (
	GatewayRegionBuiltinName = "gateway_region"

	DefaultToDatabasePrimaryRegionBuiltinName = "default_to_database_primary_region"

	RehomeRowBuiltinName = "rehome_row"
)

var digitNames = [...]string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}

const regexpFlagInfo = `

CockroachDB supports the following flags:

| Flag           | Description                                                       |
|----------------|-------------------------------------------------------------------|
| **c**          | Case-sensitive matching                                           |
| **g**          | Global matching (match each substring instead of only the first)  |
| **i**          | Case-insensitive matching                                         |
| **m** or **n** | Newline-sensitive (see below)                                     |
| **p**          | Partial newline-sensitive matching (see below)                    |
| **s**          | Newline-insensitive (default)                                     |
| **w**          | Inverse partial newline-sensitive matching (see below)            |

| Mode | ` + "`.` and `[^...]` match newlines | `^` and `$` match line boundaries" + `|
|------|----------------------------------|--------------------------------------|
| s    | yes                              | no                                   |
| w    | yes                              | yes                                  |
| p    | no                               | no                                   |
| m/n  | no                               | yes                                  |`

type builtinDefinition struct {
	props     tree.FunctionProperties
	overloads []tree.Overload
}

func GetBuiltinProperties(name string) (*tree.FunctionProperties, []tree.Overload) {
	__antithesis_instrumentation__.Notify(596954)
	def, ok := builtins[name]
	if !ok {
		__antithesis_instrumentation__.Notify(596956)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(596957)
	}
	__antithesis_instrumentation__.Notify(596955)
	return &def.props, def.overloads
}

func defProps() tree.FunctionProperties {
	__antithesis_instrumentation__.Notify(596958)
	return tree.FunctionProperties{}
}

func arrayProps() tree.FunctionProperties {
	__antithesis_instrumentation__.Notify(596959)
	return tree.FunctionProperties{Category: categoryArray}
}

func arrayPropsNullableArgs() tree.FunctionProperties {
	__antithesis_instrumentation__.Notify(596960)
	p := arrayProps()
	p.NullableArgs = true
	return p
}

func makeBuiltin(props tree.FunctionProperties, overloads ...tree.Overload) builtinDefinition {
	__antithesis_instrumentation__.Notify(596961)
	return builtinDefinition{
		props:     props,
		overloads: overloads,
	}
}

func newDecodeError(enc string) error {
	__antithesis_instrumentation__.Notify(596962)
	return pgerror.Newf(pgcode.CharacterNotInRepertoire,
		"invalid byte sequence for encoding %q", enc)
}

func newEncodeError(c rune, enc string) error {
	__antithesis_instrumentation__.Notify(596963)
	return pgerror.Newf(pgcode.UntranslatableCharacter,
		"character %q has no representation in encoding %q", c, enc)
}

func mustBeDIntInTenantRange(e tree.Expr) (tree.DInt, error) {
	__antithesis_instrumentation__.Notify(596964)
	tenID := tree.MustBeDInt(e)
	if int64(tenID) <= 0 {
		__antithesis_instrumentation__.Notify(596966)
		return 0, pgerror.New(pgcode.InvalidParameterValue, "tenant ID must be positive")
	} else {
		__antithesis_instrumentation__.Notify(596967)
	}
	__antithesis_instrumentation__.Notify(596965)
	return tenID, nil
}

var builtins = map[string]builtinDefinition{

	"length":           lengthImpls(true),
	"char_length":      lengthImpls(false),
	"character_length": lengthImpls(false),

	"bit_length": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596968)
				return tree.NewDInt(tree.DInt(len(s) * 8)), nil
			},
			types.Int,
			"Calculates the number of bits used to represent `val`.",
			tree.VolatilityImmutable,
		),
		bytesOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596969)
				return tree.NewDInt(tree.DInt(len(s) * 8)), nil
			},
			types.Int,
			"Calculates the number of bits used to represent `val`.",
			tree.VolatilityImmutable,
		),
		bitsOverload1(
			func(_ *tree.EvalContext, s *tree.DBitArray) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596970)
				return tree.NewDInt(tree.DInt(s.BitArray.BitLen())), nil
			},
			types.Int,
			"Calculates the number of bits used to represent `val`.",
			tree.VolatilityImmutable,
		),
	),

	"octet_length": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596971)
				return tree.NewDInt(tree.DInt(len(s))), nil
			},
			types.Int,
			"Calculates the number of bytes used to represent `val`.",
			tree.VolatilityImmutable,
		),
		bytesOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596972)
				return tree.NewDInt(tree.DInt(len(s))), nil
			},
			types.Int,
			"Calculates the number of bytes used to represent `val`.",
			tree.VolatilityImmutable,
		),
		bitsOverload1(
			func(_ *tree.EvalContext, s *tree.DBitArray) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596973)
				return tree.NewDInt(tree.DInt((s.BitArray.BitLen() + 7) / 8)), nil
			},
			types.Int,
			"Calculates the number of bits used to represent `val`.",
			tree.VolatilityImmutable,
		),
	),

	"lower": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		stringOverload1(
			func(evalCtx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596974)
				return tree.NewDString(strings.ToLower(s)), nil
			},
			types.String,
			"Converts all characters in `val` to their lower-case equivalents.",
			tree.VolatilityImmutable,
		),
	),

	"unaccent": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		stringOverload1(
			func(evalCtx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596975)
				var b strings.Builder
				for _, ch := range s {
					__antithesis_instrumentation__.Notify(596977)
					v, ok := unaccent.Dictionary[ch]
					if ok {
						__antithesis_instrumentation__.Notify(596978)
						b.WriteString(v)
					} else {
						__antithesis_instrumentation__.Notify(596979)
						b.WriteRune(ch)
					}
				}
				__antithesis_instrumentation__.Notify(596976)
				return tree.NewDString(b.String()), nil
			},
			types.String,
			"Removes accents (diacritic signs) from the text provided in `val`.",
			tree.VolatilityImmutable,
		),
	),

	"upper": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		stringOverload1(
			func(evalCtx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596980)
				return tree.NewDString(strings.ToUpper(s)), nil
			},
			types.String,
			"Converts all characters in `val` to their to their upper-case equivalents.",
			tree.VolatilityImmutable,
		),
	),

	"prettify_statement": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		stringOverload1(
			func(evalCtx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596981)
				formattedStmt, err := prettyStatement(tree.DefaultPrettyCfg(), s)
				if err != nil {
					__antithesis_instrumentation__.Notify(596983)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(596984)
				}
				__antithesis_instrumentation__.Notify(596982)
				return tree.NewDString(formattedStmt), nil
			},
			types.String,
			"Prettifies a statement using a the default pretty-printing config.",
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types: tree.ArgTypes{
				{"statement", types.String},
				{"line_width", types.Int},
				{"align_mode", types.Int},
				{"case_mode", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596985)
				stmt := string(tree.MustBeDString(args[0]))
				lineWidth := int(tree.MustBeDInt(args[1]))
				alignMode := int(tree.MustBeDInt(args[2]))
				caseMode := int(tree.MustBeDInt(args[3]))
				formattedStmt, err := prettyStatementCustomConfig(stmt, lineWidth, alignMode, caseMode)
				if err != nil {
					__antithesis_instrumentation__.Notify(596987)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(596988)
				}
				__antithesis_instrumentation__.Notify(596986)
				return tree.NewDString(formattedStmt), nil
			},
			Info: "Prettifies a statement using a user-configured pretty-printing config.\n" +
				"Align mode values range from 0 - 3, representing no, partial, full, and extra alignment respectively.\n" +
				"Case mode values range between 0 - 1, representing lower casing and upper casing respectively.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"substr":    substringImpls,
	"substring": substringImpls,

	"concat": makeBuiltin(
		tree.FunctionProperties{
			NullableArgs: true,
		},
		tree.Overload{
			Types:      tree.VariadicType{VarType: types.String},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596989)
				var buffer bytes.Buffer
				length := 0
				for _, d := range args {
					__antithesis_instrumentation__.Notify(596991)
					if d == tree.DNull {
						__antithesis_instrumentation__.Notify(596994)
						continue
					} else {
						__antithesis_instrumentation__.Notify(596995)
					}
					__antithesis_instrumentation__.Notify(596992)
					length += len(string(tree.MustBeDString(d)))
					if length > maxAllocatedStringSize {
						__antithesis_instrumentation__.Notify(596996)
						return nil, errStringTooLarge
					} else {
						__antithesis_instrumentation__.Notify(596997)
					}
					__antithesis_instrumentation__.Notify(596993)
					buffer.WriteString(string(tree.MustBeDString(d)))
				}
				__antithesis_instrumentation__.Notify(596990)
				return tree.NewDString(buffer.String()), nil
			},
			Info:       "Concatenates a comma-separated list of strings.",
			Volatility: tree.VolatilityImmutable,

			IgnoreVolatilityCheck: true,
		},
	),

	"concat_ws": makeBuiltin(
		tree.FunctionProperties{
			NullableArgs: true,
		},
		tree.Overload{
			Types:      tree.VariadicType{VarType: types.String},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(596998)
				if len(args) == 0 {
					__antithesis_instrumentation__.Notify(597002)
					return nil, pgerror.Newf(pgcode.UndefinedFunction, errInsufficientArgsFmtString, "concat_ws")
				} else {
					__antithesis_instrumentation__.Notify(597003)
				}
				__antithesis_instrumentation__.Notify(596999)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597004)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597005)
				}
				__antithesis_instrumentation__.Notify(597000)
				sep := string(tree.MustBeDString(args[0]))
				var buf bytes.Buffer
				prefix := ""
				length := 0
				for _, d := range args[1:] {
					__antithesis_instrumentation__.Notify(597006)
					if d == tree.DNull {
						__antithesis_instrumentation__.Notify(597009)
						continue
					} else {
						__antithesis_instrumentation__.Notify(597010)
					}
					__antithesis_instrumentation__.Notify(597007)
					length += len(prefix) + len(string(tree.MustBeDString(d)))
					if length > maxAllocatedStringSize {
						__antithesis_instrumentation__.Notify(597011)
						return nil, errStringTooLarge
					} else {
						__antithesis_instrumentation__.Notify(597012)
					}
					__antithesis_instrumentation__.Notify(597008)

					buf.WriteString(prefix)
					prefix = sep
					buf.WriteString(string(tree.MustBeDString(d)))
				}
				__antithesis_instrumentation__.Notify(597001)
				return tree.NewDString(buf.String()), nil
			},
			Info: "Uses the first argument as a separator between the concatenation of the " +
				"subsequent arguments. \n\nFor example `concat_ws('!','wow','great')` " +
				"returns `wow!great`.",
			Volatility: tree.VolatilityImmutable,

			IgnoreVolatilityCheck: true,
		},
	),

	"convert_from": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"str", types.Bytes}, {"enc", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597013)
				str := []byte(tree.MustBeDBytes(args[0]))
				enc := CleanEncodingName(string(tree.MustBeDString(args[1])))
				switch enc {

				case "utf8", "unicode", "cp65001":
					__antithesis_instrumentation__.Notify(597015)
					if !utf8.Valid(str) {
						__antithesis_instrumentation__.Notify(597020)
						return nil, newDecodeError("UTF8")
					} else {
						__antithesis_instrumentation__.Notify(597021)
					}
					__antithesis_instrumentation__.Notify(597016)
					return tree.NewDString(string(str)), nil

				case "latin1", "iso88591", "cp28591":
					__antithesis_instrumentation__.Notify(597017)
					var buf strings.Builder
					for _, c := range str {
						__antithesis_instrumentation__.Notify(597022)
						buf.WriteRune(rune(c))
					}
					__antithesis_instrumentation__.Notify(597018)
					return tree.NewDString(buf.String()), nil
				default:
					__antithesis_instrumentation__.Notify(597019)
				}
				__antithesis_instrumentation__.Notify(597014)
				return nil, pgerror.Newf(pgcode.InvalidParameterValue,
					"invalid source encoding name %q", enc)
			},
			Info: "Decode the bytes in `str` into a string using encoding `enc`. " +
				"Supports encodings 'UTF8' and 'LATIN1'.",
			Volatility:            tree.VolatilityImmutable,
			IgnoreVolatilityCheck: true,
		}),

	"convert_to": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"str", types.String}, {"enc", types.String}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597023)
				str := string(tree.MustBeDString(args[0]))
				enc := CleanEncodingName(string(tree.MustBeDString(args[1])))
				switch enc {

				case "utf8", "unicode", "cp65001":
					__antithesis_instrumentation__.Notify(597025)
					return tree.NewDBytes(tree.DBytes([]byte(str))), nil

				case "latin1", "iso88591", "cp28591":
					__antithesis_instrumentation__.Notify(597026)
					res := make([]byte, 0, len(str))
					for _, c := range str {
						__antithesis_instrumentation__.Notify(597029)
						if c > 255 {
							__antithesis_instrumentation__.Notify(597031)
							return nil, newEncodeError(c, "LATIN1")
						} else {
							__antithesis_instrumentation__.Notify(597032)
						}
						__antithesis_instrumentation__.Notify(597030)
						res = append(res, byte(c))
					}
					__antithesis_instrumentation__.Notify(597027)
					return tree.NewDBytes(tree.DBytes(res)), nil
				default:
					__antithesis_instrumentation__.Notify(597028)
				}
				__antithesis_instrumentation__.Notify(597024)
				return nil, pgerror.Newf(pgcode.InvalidParameterValue,
					"invalid destination encoding name %q", enc)
			},
			Info: "Encode the string `str` as a byte array using encoding `enc`. " +
				"Supports encodings 'UTF8' and 'LATIN1'.",
			Volatility:            tree.VolatilityImmutable,
			IgnoreVolatilityCheck: true,
		}),

	"get_bit": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"bit_string", types.VarBit}, {"index", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597033)
				bitString := tree.MustBeDBitArray(args[0])
				index := int(tree.MustBeDInt(args[1]))
				bit, err := bitString.GetBitAtIndex(index)
				if err != nil {
					__antithesis_instrumentation__.Notify(597035)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597036)
				}
				__antithesis_instrumentation__.Notify(597034)
				return tree.NewDInt(tree.DInt(bit)), nil
			},
			Info:       "Extracts a bit at given index in the bit array.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"byte_string", types.Bytes}, {"index", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597037)
				byteString := []byte(*args[0].(*tree.DBytes))
				index := int(tree.MustBeDInt(args[1]))

				if index < 0 || func() bool {
					__antithesis_instrumentation__.Notify(597040)
					return index >= 8*len(byteString) == true
				}() == true {
					__antithesis_instrumentation__.Notify(597041)
					return nil, pgerror.Newf(pgcode.ArraySubscript,
						"bit index %d out of valid range (0..%d)", index, 8*len(byteString)-1)
				} else {
					__antithesis_instrumentation__.Notify(597042)
				}
				__antithesis_instrumentation__.Notify(597038)

				if byteString[index/8]&(byte(1)<<(byte(index)%8)) != 0 {
					__antithesis_instrumentation__.Notify(597043)
					return tree.NewDInt(tree.DInt(1)), nil
				} else {
					__antithesis_instrumentation__.Notify(597044)
				}
				__antithesis_instrumentation__.Notify(597039)
				return tree.NewDInt(tree.DInt(0)), nil
			},
			Info:       "Extracts a bit at the given index in the byte array.",
			Volatility: tree.VolatilityImmutable,
		}),

	"get_byte": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"byte_string", types.Bytes}, {"index", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597045)
				byteString := []byte(*args[0].(*tree.DBytes))
				index := int(tree.MustBeDInt(args[1]))

				if index < 0 || func() bool {
					__antithesis_instrumentation__.Notify(597047)
					return index >= len(byteString) == true
				}() == true {
					__antithesis_instrumentation__.Notify(597048)
					return nil, pgerror.Newf(pgcode.ArraySubscript,
						"byte index %d out of valid range (0..%d)", index, len(byteString)-1)
				} else {
					__antithesis_instrumentation__.Notify(597049)
				}
				__antithesis_instrumentation__.Notify(597046)
				return tree.NewDInt(tree.DInt(byteString[index])), nil
			},
			Info:       "Extracts a byte at the given index in the byte array.",
			Volatility: tree.VolatilityImmutable,
		}),

	"set_bit": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types: tree.ArgTypes{
				{"bit_string", types.VarBit},
				{"index", types.Int},
				{"to_set", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.VarBit),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597050)
				bitString := tree.MustBeDBitArray(args[0])
				index := int(tree.MustBeDInt(args[1]))
				toSet := int(tree.MustBeDInt(args[2]))

				if toSet != 0 && func() bool {
					__antithesis_instrumentation__.Notify(597053)
					return toSet != 1 == true
				}() == true {
					__antithesis_instrumentation__.Notify(597054)
					return nil, pgerror.Newf(pgcode.InvalidParameterValue,
						"new bit must be 0 or 1.")
				} else {
					__antithesis_instrumentation__.Notify(597055)
				}
				__antithesis_instrumentation__.Notify(597051)
				updatedBitString, err := bitString.SetBitAtIndex(index, toSet)
				if err != nil {
					__antithesis_instrumentation__.Notify(597056)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597057)
				}
				__antithesis_instrumentation__.Notify(597052)
				return &tree.DBitArray{BitArray: updatedBitString}, nil
			},
			Info:       "Updates a bit at given index in the bit array.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"byte_string", types.Bytes},
				{"index", types.Int},
				{"to_set", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597058)
				byteString := []byte(*args[0].(*tree.DBytes))
				index := int(tree.MustBeDInt(args[1]))
				toSet := int(tree.MustBeDInt(args[2]))

				if toSet != 0 && func() bool {
					__antithesis_instrumentation__.Notify(597061)
					return toSet != 1 == true
				}() == true {
					__antithesis_instrumentation__.Notify(597062)
					return nil, pgerror.Newf(pgcode.InvalidParameterValue,
						"new bit must be 0 or 1")
				} else {
					__antithesis_instrumentation__.Notify(597063)
				}
				__antithesis_instrumentation__.Notify(597059)

				if index < 0 || func() bool {
					__antithesis_instrumentation__.Notify(597064)
					return index >= 8*len(byteString) == true
				}() == true {
					__antithesis_instrumentation__.Notify(597065)
					return nil, pgerror.Newf(pgcode.ArraySubscript,
						"bit index %d out of valid range (0..%d)", index, 8*len(byteString)-1)
				} else {
					__antithesis_instrumentation__.Notify(597066)
				}
				__antithesis_instrumentation__.Notify(597060)

				byteString[index/8] &= ^(byte(1) << (byte(index) % 8))

				byteString[index/8] |= byte(toSet) << (byte(index) % 8)
				return tree.NewDBytes(tree.DBytes(byteString)), nil
			},
			Info:       "Updates a bit at the given index in the byte array.",
			Volatility: tree.VolatilityImmutable,
		}),

	"set_byte": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types: tree.ArgTypes{
				{"byte_string", types.Bytes},
				{"index", types.Int},
				{"to_set", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597067)
				byteString := []byte(*args[0].(*tree.DBytes))
				index := int(tree.MustBeDInt(args[1]))
				toSet := int(tree.MustBeDInt(args[2]))

				if index < 0 || func() bool {
					__antithesis_instrumentation__.Notify(597069)
					return index >= len(byteString) == true
				}() == true {
					__antithesis_instrumentation__.Notify(597070)
					return nil, pgerror.Newf(pgcode.ArraySubscript,
						"byte index %d out of valid range (0..%d)", index, len(byteString)-1)
				} else {
					__antithesis_instrumentation__.Notify(597071)
				}
				__antithesis_instrumentation__.Notify(597068)
				byteString[index] = byte(toSet)
				return tree.NewDBytes(tree.DBytes(byteString)), nil
			},
			Info:       "Updates a byte at the given index in the byte array.",
			Volatility: tree.VolatilityImmutable,
		}),

	"gen_random_uuid":  generateRandomUUIDImpl,
	"uuid_generate_v4": generateRandomUUIDImpl,

	"to_uuid": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.String}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597072)
				s := string(tree.MustBeDString(args[0]))
				uv, err := uuid.FromString(s)
				if err != nil {
					__antithesis_instrumentation__.Notify(597074)
					return nil, pgerror.Wrap(err, pgcode.InvalidParameterValue, "invalid UUID")
				} else {
					__antithesis_instrumentation__.Notify(597075)
				}
				__antithesis_instrumentation__.Notify(597073)
				return tree.NewDBytes(tree.DBytes(uv.GetBytes())), nil
			},
			Info: "Converts the character string representation of a UUID to its byte string " +
				"representation.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"from_uuid": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Bytes}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597076)
				b := []byte(*args[0].(*tree.DBytes))
				uv, err := uuid.FromBytes(b)
				if err != nil {
					__antithesis_instrumentation__.Notify(597078)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597079)
				}
				__antithesis_instrumentation__.Notify(597077)
				return tree.NewDString(uv.String()), nil
			},
			Info: "Converts the byte string representation of a UUID to its character string " +
				"representation.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"gen_random_ulid": makeBuiltin(
		tree.FunctionProperties{
			Category: categoryIDGeneration,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Uuid),
			Fn: func(_ *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597080)
				entropy := ulid.Monotonic(cryptorand.Reader, 0)
				uv := ulid.MustNew(ulid.Now(), entropy)
				return tree.NewDUuid(tree.DUuid{UUID: uuid.UUID(uv)}), nil
			},
			Info:       "Generates a random ULID and returns it as a value of UUID type.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"uuid_to_ulid": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Uuid}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597081)
				b := (*args[0].(*tree.DUuid)).GetBytes()
				var ul ulid.ULID
				if err := ul.UnmarshalBinary(b); err != nil {
					__antithesis_instrumentation__.Notify(597083)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597084)
				}
				__antithesis_instrumentation__.Notify(597082)
				return tree.NewDString(ul.String()), nil
			},
			Info:       "Converts a UUID-encoded ULID to its string representation.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"ulid_to_uuid": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.String}},
			ReturnType: tree.FixedReturnType(types.Uuid),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597085)
				s := tree.MustBeDString(args[0])
				u, err := ulid.Parse(string(s))
				if err != nil {
					__antithesis_instrumentation__.Notify(597089)
					return nil, pgerror.Wrap(err, pgcode.InvalidParameterValue, "invalid ULID")
				} else {
					__antithesis_instrumentation__.Notify(597090)
				}
				__antithesis_instrumentation__.Notify(597086)
				b, err := u.MarshalBinary()
				if err != nil {
					__antithesis_instrumentation__.Notify(597091)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597092)
				}
				__antithesis_instrumentation__.Notify(597087)
				uv, err := uuid.FromBytes(b)
				if err != nil {
					__antithesis_instrumentation__.Notify(597093)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597094)
				}
				__antithesis_instrumentation__.Notify(597088)
				return tree.NewDUuid(tree.DUuid{UUID: uv}), nil
			},
			Info:       "Converts a ULID string to its UUID-encoded representation.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"abbrev": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.INet}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597095)
				dIPAddr := tree.MustBeDIPAddr(args[0])
				return tree.NewDString(dIPAddr.IPAddr.String()), nil
			},
			Info: "Converts the combined IP address and prefix length to an abbreviated display format as text." +
				"For INET types, this will omit the prefix length if it's not the default (32 or IPv4, 128 for IPv6)" +
				"\n\nFor example, `abbrev('192.168.1.2/24')` returns `'192.168.1.2/24'`",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"broadcast": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.INet}},
			ReturnType: tree.FixedReturnType(types.INet),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597096)
				dIPAddr := tree.MustBeDIPAddr(args[0])
				broadcastIPAddr := dIPAddr.IPAddr.Broadcast()
				return &tree.DIPAddr{IPAddr: broadcastIPAddr}, nil
			},
			Info: "Gets the broadcast address for the network address represented by the value." +
				"\n\nFor example, `broadcast('192.168.1.2/24')` returns `'192.168.1.255/24'`",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"family": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.INet}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597097)
				dIPAddr := tree.MustBeDIPAddr(args[0])
				if dIPAddr.Family == ipaddr.IPv4family {
					__antithesis_instrumentation__.Notify(597099)
					return tree.NewDInt(tree.DInt(4)), nil
				} else {
					__antithesis_instrumentation__.Notify(597100)
				}
				__antithesis_instrumentation__.Notify(597098)
				return tree.NewDInt(tree.DInt(6)), nil
			},
			Info: "Extracts the IP family of the value; 4 for IPv4, 6 for IPv6." +
				"\n\nFor example, `family('::1')` returns `6`",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"host": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.INet}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597101)
				dIPAddr := tree.MustBeDIPAddr(args[0])
				s := dIPAddr.IPAddr.String()
				if i := strings.IndexByte(s, '/'); i != -1 {
					__antithesis_instrumentation__.Notify(597103)
					return tree.NewDString(s[:i]), nil
				} else {
					__antithesis_instrumentation__.Notify(597104)
				}
				__antithesis_instrumentation__.Notify(597102)
				return tree.NewDString(s), nil
			},
			Info: "Extracts the address part of the combined address/prefixlen value as text." +
				"\n\nFor example, `host('192.168.1.2/16')` returns `'192.168.1.2'`",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"hostmask": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.INet}},
			ReturnType: tree.FixedReturnType(types.INet),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597105)
				dIPAddr := tree.MustBeDIPAddr(args[0])
				ipAddr := dIPAddr.IPAddr.Hostmask()
				return &tree.DIPAddr{IPAddr: ipAddr}, nil
			},
			Info: "Creates an IP host mask corresponding to the prefix length in the value." +
				"\n\nFor example, `hostmask('192.168.1.2/16')` returns `'0.0.255.255'`",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"masklen": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.INet}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597106)
				dIPAddr := tree.MustBeDIPAddr(args[0])
				return tree.NewDInt(tree.DInt(dIPAddr.Mask)), nil
			},
			Info: "Retrieves the prefix length stored in the value." +
				"\n\nFor example, `masklen('192.168.1.2/16')` returns `16`",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"netmask": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.INet}},
			ReturnType: tree.FixedReturnType(types.INet),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597107)
				dIPAddr := tree.MustBeDIPAddr(args[0])
				ipAddr := dIPAddr.IPAddr.Netmask()
				return &tree.DIPAddr{IPAddr: ipAddr}, nil
			},
			Info: "Creates an IP network mask corresponding to the prefix length in the value." +
				"\n\nFor example, `netmask('192.168.1.2/16')` returns `'255.255.0.0'`",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"set_masklen": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"val", types.INet},
				{"prefixlen", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.INet),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597108)
				dIPAddr := tree.MustBeDIPAddr(args[0])
				mask := int(tree.MustBeDInt(args[1]))

				if !(dIPAddr.Family == ipaddr.IPv4family && func() bool {
					__antithesis_instrumentation__.Notify(597110)
					return mask >= 0 == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(597111)
					return mask <= 32 == true
				}() == true) && func() bool {
					__antithesis_instrumentation__.Notify(597112)
					return !(dIPAddr.Family == ipaddr.IPv6family && func() bool {
						__antithesis_instrumentation__.Notify(597113)
						return mask >= 0 == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(597114)
						return mask <= 128 == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(597115)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue, "invalid mask length: %d", mask)
				} else {
					__antithesis_instrumentation__.Notify(597116)
				}
				__antithesis_instrumentation__.Notify(597109)
				return &tree.DIPAddr{IPAddr: ipaddr.IPAddr{Family: dIPAddr.Family, Addr: dIPAddr.Addr, Mask: byte(mask)}}, nil
			},
			Info: "Sets the prefix length of `val` to `prefixlen`.\n\n" +
				"For example, `set_masklen('192.168.1.2', 16)` returns `'192.168.1.2/16'`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"text": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.INet}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597117)
				dIPAddr := tree.MustBeDIPAddr(args[0])
				s := dIPAddr.IPAddr.String()

				if strings.IndexByte(s, '/') == -1 {
					__antithesis_instrumentation__.Notify(597119)
					s += "/" + strconv.Itoa(int(dIPAddr.Mask))
				} else {
					__antithesis_instrumentation__.Notify(597120)
				}
				__antithesis_instrumentation__.Notify(597118)
				return tree.NewDString(s), nil
			},
			Info:       "Converts the IP address and prefix length to text.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"inet_same_family": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"val", types.INet},
				{"val", types.INet},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597121)
				first := tree.MustBeDIPAddr(args[0])
				other := tree.MustBeDIPAddr(args[1])
				return tree.MakeDBool(tree.DBool(first.Family == other.Family)), nil
			},
			Info:       "Checks if two IP addresses are of the same IP family.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"inet_contained_by_or_equals": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"val", types.INet},
				{"container", types.INet},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597122)
				ipAddr := tree.MustBeDIPAddr(args[0]).IPAddr
				other := tree.MustBeDIPAddr(args[1]).IPAddr
				return tree.MakeDBool(tree.DBool(ipAddr.ContainedByOrEquals(&other))), nil
			},
			Info: "Test for subnet inclusion or equality, using only the network parts of the addresses. " +
				"The host part of the addresses is ignored.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"inet_contains_or_equals": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"container", types.INet},
				{"val", types.INet},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597123)
				ipAddr := tree.MustBeDIPAddr(args[0]).IPAddr
				other := tree.MustBeDIPAddr(args[1]).IPAddr
				return tree.MakeDBool(tree.DBool(ipAddr.ContainsOrEquals(&other))), nil
			},
			Info: "Test for subnet inclusion or equality, using only the network parts of the addresses. " +
				"The host part of the addresses is ignored.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"from_ip": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Bytes}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597124)
				ipstr := args[0].(*tree.DBytes)
				nboip := net.IP(*ipstr)
				sv := nboip.String()

				if sv == "<nil>" {
					__antithesis_instrumentation__.Notify(597126)
					return nil, errZeroIP
				} else {
					__antithesis_instrumentation__.Notify(597127)
				}
				__antithesis_instrumentation__.Notify(597125)
				return tree.NewDString(sv), nil
			},
			Info: "Converts the byte string representation of an IP to its character string " +
				"representation.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"to_ip": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.String}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597128)
				ipdstr := tree.MustBeDString(args[0])
				ip := net.ParseIP(string(ipdstr))

				if ip == nil {
					__antithesis_instrumentation__.Notify(597130)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue, "invalid IP format: %s", ipdstr.String())
				} else {
					__antithesis_instrumentation__.Notify(597131)
				}
				__antithesis_instrumentation__.Notify(597129)
				return tree.NewDBytes(tree.DBytes(ip)), nil
			},
			Info: "Converts the character string representation of an IP to its byte string " +
				"representation.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"split_part": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"input", types.String},
				{"delimiter", types.String},
				{"return_index_pos", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597132)
				text := string(tree.MustBeDString(args[0]))
				sep := string(tree.MustBeDString(args[1]))
				field := int(tree.MustBeDInt(args[2]))

				if field <= 0 {
					__antithesis_instrumentation__.Notify(597135)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue, "field position %d must be greater than zero", field)
				} else {
					__antithesis_instrumentation__.Notify(597136)
				}
				__antithesis_instrumentation__.Notify(597133)

				splits := strings.Split(text, sep)
				if field > len(splits) {
					__antithesis_instrumentation__.Notify(597137)
					return tree.NewDString(""), nil
				} else {
					__antithesis_instrumentation__.Notify(597138)
				}
				__antithesis_instrumentation__.Notify(597134)
				return tree.NewDString(splits[field-1]), nil
			},
			Info: "Splits `input` on `delimiter` and return the value in the `return_index_pos`  " +
				"position (starting at 1). \n\nFor example, `split_part('123.456.789.0','.',3)`" +
				"returns `789`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"repeat": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.String}, {"repeat_counter", types.Int}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (_ tree.Datum, err error) {
				__antithesis_instrumentation__.Notify(597139)
				s := string(tree.MustBeDString(args[0]))
				count := int(tree.MustBeDInt(args[1]))

				ln := len(s) * count

				if count <= 0 {
					__antithesis_instrumentation__.Notify(597141)
					count = 0
				} else {
					__antithesis_instrumentation__.Notify(597142)
					if ln/count != len(s) {
						__antithesis_instrumentation__.Notify(597143)

						return nil, errStringTooLarge
					} else {
						__antithesis_instrumentation__.Notify(597144)
						if ln > maxAllocatedStringSize {
							__antithesis_instrumentation__.Notify(597145)
							return nil, errStringTooLarge
						} else {
							__antithesis_instrumentation__.Notify(597146)
						}
					}
				}
				__antithesis_instrumentation__.Notify(597140)

				return tree.NewDString(strings.Repeat(s, count)), nil
			},
			Info: "Concatenates `input` `repeat_counter` number of times.\n\nFor example, " +
				"`repeat('dog', 2)` returns `dogdog`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"encode": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"data", types.Bytes}, {"format", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (_ tree.Datum, err error) {
				__antithesis_instrumentation__.Notify(597147)
				data, format := *args[0].(*tree.DBytes), string(tree.MustBeDString(args[1]))
				be, ok := lex.BytesEncodeFormatFromString(format)
				if !ok {
					__antithesis_instrumentation__.Notify(597149)
					return nil, pgerror.New(pgcode.InvalidParameterValue,
						"only 'hex', 'escape', and 'base64' formats are supported for encode()")
				} else {
					__antithesis_instrumentation__.Notify(597150)
				}
				__antithesis_instrumentation__.Notify(597148)
				return tree.NewDString(lex.EncodeByteArrayToRawBytes(
					string(data), be, true)), nil
			},
			Info:       "Encodes `data` using `format` (`hex` / `escape` / `base64`).",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"decode": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"text", types.String}, {"format", types.String}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (_ tree.Datum, err error) {
				__antithesis_instrumentation__.Notify(597151)
				data, format := string(tree.MustBeDString(args[0])), string(tree.MustBeDString(args[1]))
				be, ok := lex.BytesEncodeFormatFromString(format)
				if !ok {
					__antithesis_instrumentation__.Notify(597154)
					return nil, pgerror.New(pgcode.InvalidParameterValue,
						"only 'hex', 'escape', and 'base64' formats are supported for decode()")
				} else {
					__antithesis_instrumentation__.Notify(597155)
				}
				__antithesis_instrumentation__.Notify(597152)
				res, err := lex.DecodeRawBytesToByteArray(data, be)
				if err != nil {
					__antithesis_instrumentation__.Notify(597156)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597157)
				}
				__antithesis_instrumentation__.Notify(597153)
				return tree.NewDBytes(tree.DBytes(res)), nil
			},
			Info:       "Decodes `data` using `format` (`hex` / `escape` / `base64`).",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"compress": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"data", types.Bytes}, {"codec", types.String}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (_ tree.Datum, err error) {
				__antithesis_instrumentation__.Notify(597158)
				uncompressedData := []byte(tree.MustBeDBytes(args[0]))
				codec := string(tree.MustBeDString(args[1]))
				switch strings.ToUpper(codec) {
				case "GZIP":
					__antithesis_instrumentation__.Notify(597159)
					gzipBuf := bytes.NewBuffer([]byte{})
					gz := gzip.NewWriter(gzipBuf)
					if _, err := gz.Write(uncompressedData); err != nil {
						__antithesis_instrumentation__.Notify(597163)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597164)
					}
					__antithesis_instrumentation__.Notify(597160)
					if err := gz.Close(); err != nil {
						__antithesis_instrumentation__.Notify(597165)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597166)
					}
					__antithesis_instrumentation__.Notify(597161)
					return tree.NewDBytes(tree.DBytes(gzipBuf.Bytes())), nil
				default:
					__antithesis_instrumentation__.Notify(597162)
					return nil, pgerror.New(pgcode.InvalidParameterValue,
						"only 'gzip' codec is supported for compress()")
				}
			},
			Info:       "Compress `data` with the specified `codec` (`gzip`).",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"decompress": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"data", types.Bytes}, {"codec", types.String}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (_ tree.Datum, err error) {
				__antithesis_instrumentation__.Notify(597167)
				compressedData := []byte(tree.MustBeDBytes(args[0]))
				codec := string(tree.MustBeDString(args[1]))
				switch strings.ToUpper(codec) {
				case "GZIP":
					__antithesis_instrumentation__.Notify(597168)
					r, err := gzip.NewReader(bytes.NewBuffer(compressedData))
					if err != nil {
						__antithesis_instrumentation__.Notify(597172)
						return nil, errors.Wrap(err, "failed to decompress")
					} else {
						__antithesis_instrumentation__.Notify(597173)
					}
					__antithesis_instrumentation__.Notify(597169)
					defer r.Close()
					decompressedBytes, err := ioutil.ReadAll(r)
					if err != nil {
						__antithesis_instrumentation__.Notify(597174)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597175)
					}
					__antithesis_instrumentation__.Notify(597170)
					return tree.NewDBytes(tree.DBytes(decompressedBytes)), nil
				default:
					__antithesis_instrumentation__.Notify(597171)
					return nil, pgerror.New(pgcode.InvalidParameterValue,
						"only 'gzip' codec is supported for decompress()")
				}
			},
			Info:       "Decompress `data` with the specified `codec` (`gzip`).",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"ascii": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597176)
				for _, ch := range s {
					__antithesis_instrumentation__.Notify(597178)
					return tree.NewDInt(tree.DInt(ch)), nil
				}
				__antithesis_instrumentation__.Notify(597177)
				return nil, errEmptyInputString
			},
			types.Int,
			"Returns the character code of the first character in `val`. Despite the name, the function supports Unicode too.",
			tree.VolatilityImmutable,
		)),

	"chr": makeBuiltin(tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Int}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597179)
				x := tree.MustBeDInt(args[0])
				var answer string
				switch {
				case x < 0:
					__antithesis_instrumentation__.Notify(597181)
					return nil, errChrValueTooSmall
				case x > utf8.MaxRune:
					__antithesis_instrumentation__.Notify(597182)
					return nil, errChrValueTooLarge
				default:
					__antithesis_instrumentation__.Notify(597183)
					answer = string(rune(x))
				}
				__antithesis_instrumentation__.Notify(597180)
				return tree.NewDString(answer), nil
			},
			Info:       "Returns the character with the code given in `val`. Inverse function of `ascii()`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"md5": hashBuiltin(
		func() hash.Hash { __antithesis_instrumentation__.Notify(597184); return md5.New() },
		"Calculates the MD5 hash value of a set of values.",
	),

	"sha1": hashBuiltin(
		func() hash.Hash { __antithesis_instrumentation__.Notify(597185); return sha1.New() },
		"Calculates the SHA1 hash value of a set of values.",
	),

	"sha224": hashBuiltin(
		func() hash.Hash { __antithesis_instrumentation__.Notify(597186); return sha256.New224() },
		"Calculates the SHA224 hash value of a set of values.",
	),

	"sha256": hashBuiltin(
		func() hash.Hash { __antithesis_instrumentation__.Notify(597187); return sha256.New() },
		"Calculates the SHA256 hash value of a set of values.",
	),

	"sha384": hashBuiltin(
		func() hash.Hash { __antithesis_instrumentation__.Notify(597188); return sha512.New384() },
		"Calculates the SHA384 hash value of a set of values.",
	),

	"sha512": hashBuiltin(
		func() hash.Hash { __antithesis_instrumentation__.Notify(597189); return sha512.New() },
		"Calculates the SHA512 hash value of a set of values.",
	),

	"fnv32": hash32Builtin(
		func() hash.Hash32 { __antithesis_instrumentation__.Notify(597190); return fnv.New32() },
		"Calculates the 32-bit FNV-1 hash value of a set of values.",
	),

	"fnv32a": hash32Builtin(
		func() hash.Hash32 { __antithesis_instrumentation__.Notify(597191); return fnv.New32a() },
		"Calculates the 32-bit FNV-1a hash value of a set of values.",
	),

	"fnv64": hash64Builtin(
		func() hash.Hash64 { __antithesis_instrumentation__.Notify(597192); return fnv.New64() },
		"Calculates the 64-bit FNV-1 hash value of a set of values.",
	),

	"fnv64a": hash64Builtin(
		func() hash.Hash64 { __antithesis_instrumentation__.Notify(597193); return fnv.New64a() },
		"Calculates the 64-bit FNV-1a hash value of a set of values.",
	),

	"crc32ieee": hash32Builtin(
		func() hash.Hash32 { __antithesis_instrumentation__.Notify(597194); return crc32.New(crc32.IEEETable) },
		"Calculates the CRC-32 hash using the IEEE polynomial.",
	),

	"crc32c": hash32Builtin(
		func() hash.Hash32 {
			__antithesis_instrumentation__.Notify(597195)
			return crc32.New(crc32.MakeTable(crc32.Castagnoli))
		},
		"Calculates the CRC-32 hash using the Castagnoli polynomial.",
	),

	"digest": makeBuiltin(
		tree.FunctionProperties{Category: categoryCrypto},
		tree.Overload{
			Types:      tree.ArgTypes{{"data", types.String}, {"type", types.String}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597196)
				alg := tree.MustBeDString(args[1])
				hashFunc, err := getHashFunc(string(alg))
				if err != nil {
					__antithesis_instrumentation__.Notify(597199)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597200)
				}
				__antithesis_instrumentation__.Notify(597197)
				h := hashFunc()
				if ok, err := feedHash(h, args[:1]); !ok || func() bool {
					__antithesis_instrumentation__.Notify(597201)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(597202)
					return tree.DNull, err
				} else {
					__antithesis_instrumentation__.Notify(597203)
				}
				__antithesis_instrumentation__.Notify(597198)
				return tree.NewDBytes(tree.DBytes(h.Sum(nil))), nil
			},
			Info: "Computes a binary hash of the given `data`. `type` is the algorithm " +
				"to use (md5, sha1, sha224, sha256, sha384, or sha512).",
			Volatility: tree.VolatilityLeakProof,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"data", types.Bytes}, {"type", types.String}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597204)
				alg := tree.MustBeDString(args[1])
				hashFunc, err := getHashFunc(string(alg))
				if err != nil {
					__antithesis_instrumentation__.Notify(597207)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597208)
				}
				__antithesis_instrumentation__.Notify(597205)
				h := hashFunc()
				if ok, err := feedHash(h, args[:1]); !ok || func() bool {
					__antithesis_instrumentation__.Notify(597209)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(597210)
					return tree.DNull, err
				} else {
					__antithesis_instrumentation__.Notify(597211)
				}
				__antithesis_instrumentation__.Notify(597206)
				return tree.NewDBytes(tree.DBytes(h.Sum(nil))), nil
			},
			Info: "Computes a binary hash of the given `data`. `type` is the algorithm " +
				"to use (md5, sha1, sha224, sha256, sha384, or sha512).",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"hmac": makeBuiltin(
		tree.FunctionProperties{Category: categoryCrypto},
		tree.Overload{
			Types:      tree.ArgTypes{{"data", types.String}, {"key", types.String}, {"type", types.String}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597212)
				key := tree.MustBeDString(args[1])
				alg := tree.MustBeDString(args[2])
				hashFunc, err := getHashFunc(string(alg))
				if err != nil {
					__antithesis_instrumentation__.Notify(597215)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597216)
				}
				__antithesis_instrumentation__.Notify(597213)
				h := hmac.New(hashFunc, []byte(key))
				if ok, err := feedHash(h, args[:1]); !ok || func() bool {
					__antithesis_instrumentation__.Notify(597217)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(597218)
					return tree.DNull, err
				} else {
					__antithesis_instrumentation__.Notify(597219)
				}
				__antithesis_instrumentation__.Notify(597214)
				return tree.NewDBytes(tree.DBytes(h.Sum(nil))), nil
			},
			Info:       "Calculates hashed MAC for `data` with key `key`. `type` is the same as in `digest()`.",
			Volatility: tree.VolatilityLeakProof,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"data", types.Bytes}, {"key", types.Bytes}, {"type", types.String}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597220)
				key := tree.MustBeDBytes(args[1])
				alg := tree.MustBeDString(args[2])
				hashFunc, err := getHashFunc(string(alg))
				if err != nil {
					__antithesis_instrumentation__.Notify(597223)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597224)
				}
				__antithesis_instrumentation__.Notify(597221)
				h := hmac.New(hashFunc, []byte(key))
				if ok, err := feedHash(h, args[:1]); !ok || func() bool {
					__antithesis_instrumentation__.Notify(597225)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(597226)
					return tree.DNull, err
				} else {
					__antithesis_instrumentation__.Notify(597227)
				}
				__antithesis_instrumentation__.Notify(597222)
				return tree.NewDBytes(tree.DBytes(h.Sum(nil))), nil
			},
			Info:       "Calculates hashed MAC for `data` with key `key`. `type` is the same as in `digest()`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"to_hex": makeBuiltin(
		tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Int}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597228)
				val := tree.MustBeDInt(args[0])

				return tree.NewDString(fmt.Sprintf("%x", uint64(val))), nil
			},
			Info:       "Converts `val` to its hexadecimal representation.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Bytes}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597229)
				return tree.NewDString(fmt.Sprintf("%x", tree.MustBeDBytes(args[0]))), nil
			},
			Info:       "Converts `val` to its hexadecimal representation.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597230)
				return tree.NewDString(fmt.Sprintf("%x", tree.MustBeDString(args[0]))), nil
			},
			Info:       "Converts `val` to its hexadecimal representation.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"to_english": makeBuiltin(
		tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Int}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597231)
				val := int(*args[0].(*tree.DInt))
				var buf bytes.Buffer
				var digits []string
				if val < 0 {
					__antithesis_instrumentation__.Notify(597235)
					buf.WriteString("minus-")
					if val == math.MinInt64 {
						__antithesis_instrumentation__.Notify(597237)

						digits = append(digits, digitNames[8])
						val /= 10
					} else {
						__antithesis_instrumentation__.Notify(597238)
					}
					__antithesis_instrumentation__.Notify(597236)
					val = -val
				} else {
					__antithesis_instrumentation__.Notify(597239)
				}
				__antithesis_instrumentation__.Notify(597232)
				digits = append(digits, digitNames[val%10])
				for val > 9 {
					__antithesis_instrumentation__.Notify(597240)
					val /= 10
					digits = append(digits, digitNames[val%10])
				}
				__antithesis_instrumentation__.Notify(597233)
				for i := len(digits) - 1; i >= 0; i-- {
					__antithesis_instrumentation__.Notify(597241)
					if i < len(digits)-1 {
						__antithesis_instrumentation__.Notify(597243)
						buf.WriteByte('-')
					} else {
						__antithesis_instrumentation__.Notify(597244)
					}
					__antithesis_instrumentation__.Notify(597242)
					buf.WriteString(digits[i])
				}
				__antithesis_instrumentation__.Notify(597234)
				return tree.NewDString(buf.String()), nil
			},
			Info:       "This function enunciates the value of its argument using English cardinals.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"strpos": makeBuiltin(
		tree.FunctionProperties{Category: categoryString},
		stringOverload2(
			"input",
			"find",
			func(_ *tree.EvalContext, s, substring string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597245)
				index := strings.Index(s, substring)
				if index < 0 {
					__antithesis_instrumentation__.Notify(597247)
					return tree.DZero, nil
				} else {
					__antithesis_instrumentation__.Notify(597248)
				}
				__antithesis_instrumentation__.Notify(597246)

				return tree.NewDInt(tree.DInt(utf8.RuneCountInString(s[:index]) + 1)), nil
			},
			types.Int,
			"Calculates the position where the string `find` begins in `input`. \n\nFor"+
				" example, `strpos('doggie', 'gie')` returns `4`.",
			tree.VolatilityImmutable,
		),
		bitsOverload2("input", "find",
			func(_ *tree.EvalContext, bitString, bitSubstring *tree.DBitArray) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597249)
				index := strings.Index(bitString.BitArray.String(), bitSubstring.BitArray.String())
				if index < 0 {
					__antithesis_instrumentation__.Notify(597251)
					return tree.DZero, nil
				} else {
					__antithesis_instrumentation__.Notify(597252)
				}
				__antithesis_instrumentation__.Notify(597250)
				return tree.NewDInt(tree.DInt(index + 1)), nil
			},
			types.Int,
			"Calculates the position where the bit subarray `find` begins in `input`.",
			tree.VolatilityImmutable,
		),
		bytesOverload2(
			"input",
			"find",
			func(_ *tree.EvalContext, byteString, byteSubstring string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597253)
				index := strings.Index(byteString, byteSubstring)
				if index < 0 {
					__antithesis_instrumentation__.Notify(597255)
					return tree.DZero, nil
				} else {
					__antithesis_instrumentation__.Notify(597256)
				}
				__antithesis_instrumentation__.Notify(597254)
				return tree.NewDInt(tree.DInt(index + 1)), nil
			},
			types.Int,
			"Calculates the position where the byte subarray `find` begins in `input`.",
			tree.VolatilityImmutable,
		)),

	"overlay": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"input", types.String},
				{"overlay_val", types.String},
				{"start_pos", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597257)
				s := string(tree.MustBeDString(args[0]))
				to := string(tree.MustBeDString(args[1]))
				pos := int(tree.MustBeDInt(args[2]))
				size := utf8.RuneCountInString(to)
				return overlay(s, to, pos, size)
			},
			Info: "Replaces characters in `input` with `overlay_val` starting at `start_pos` " +
				"(begins at 1). \n\nFor example, `overlay('doggie', 'CAT', 2)` returns " +
				"`dCATie`.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"input", types.String},
				{"overlay_val", types.String},
				{"start_pos", types.Int},
				{"end_pos", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597258)
				s := string(tree.MustBeDString(args[0]))
				to := string(tree.MustBeDString(args[1]))
				pos := int(tree.MustBeDInt(args[2]))
				size := int(tree.MustBeDInt(args[3]))
				return overlay(s, to, pos, size)
			},
			Info: "Deletes the characters in `input` between `start_pos` and `end_pos` (count " +
				"starts at 1), and then insert `overlay_val` at `start_pos`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"lpad": makeBuiltin(
		tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"string", types.String}, {"length", types.Int}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597259)
				s := string(tree.MustBeDString(args[0]))
				length := int(tree.MustBeDInt(args[1]))
				ret, err := lpad(s, length, " ")
				if err != nil {
					__antithesis_instrumentation__.Notify(597261)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597262)
				}
				__antithesis_instrumentation__.Notify(597260)
				return tree.NewDString(ret), nil
			},
			Info: "Pads `string` to `length` by adding ' ' to the left of `string`." +
				"If `string` is longer than `length` it is truncated.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"string", types.String}, {"length", types.Int}, {"fill", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597263)
				s := string(tree.MustBeDString(args[0]))
				length := int(tree.MustBeDInt(args[1]))
				fill := string(tree.MustBeDString(args[2]))
				ret, err := lpad(s, length, fill)
				if err != nil {
					__antithesis_instrumentation__.Notify(597265)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597266)
				}
				__antithesis_instrumentation__.Notify(597264)
				return tree.NewDString(ret), nil
			},
			Info: "Pads `string` by adding `fill` to the left of `string` to make it `length`. " +
				"If `string` is longer than `length` it is truncated.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"rpad": makeBuiltin(
		tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"string", types.String}, {"length", types.Int}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597267)
				s := string(tree.MustBeDString(args[0]))
				length := int(tree.MustBeDInt(args[1]))
				ret, err := rpad(s, length, " ")
				if err != nil {
					__antithesis_instrumentation__.Notify(597269)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597270)
				}
				__antithesis_instrumentation__.Notify(597268)
				return tree.NewDString(ret), nil
			},
			Info: "Pads `string` to `length` by adding ' ' to the right of string. " +
				"If `string` is longer than `length` it is truncated.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"string", types.String}, {"length", types.Int}, {"fill", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597271)
				s := string(tree.MustBeDString(args[0]))
				length := int(tree.MustBeDInt(args[1]))
				fill := string(tree.MustBeDString(args[2]))
				ret, err := rpad(s, length, fill)
				if err != nil {
					__antithesis_instrumentation__.Notify(597273)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597274)
				}
				__antithesis_instrumentation__.Notify(597272)
				return tree.NewDString(ret), nil
			},
			Info: "Pads `string` to `length` by adding `fill` to the right of `string`. " +
				"If `string` is longer than `length` it is truncated.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"btrim": makeBuiltin(defProps(),
		stringOverload2(
			"input",
			"trim_chars",
			func(_ *tree.EvalContext, s, chars string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597275)
				return tree.NewDString(strings.Trim(s, chars)), nil
			},
			types.String,
			"Removes any characters included in `trim_chars` from the beginning or end"+
				" of `input` (applies recursively). \n\nFor example, `btrim('doggie', 'eod')` "+
				"returns `ggi`.",
			tree.VolatilityImmutable,
		),
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597276)
				return tree.NewDString(strings.TrimSpace(s)), nil
			},
			types.String,
			"Removes all spaces from the beginning and end of `val`.",
			tree.VolatilityImmutable,
		),
	),

	"ltrim": makeBuiltin(defProps(),
		stringOverload2(
			"input",
			"trim_chars",
			func(_ *tree.EvalContext, s, chars string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597277)
				return tree.NewDString(strings.TrimLeft(s, chars)), nil
			},
			types.String,
			"Removes any characters included in `trim_chars` from the beginning "+
				"(left-hand side) of `input` (applies recursively). \n\nFor example, "+
				"`ltrim('doggie', 'od')` returns `ggie`.",
			tree.VolatilityImmutable,
		),
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597278)
				return tree.NewDString(strings.TrimLeftFunc(s, unicode.IsSpace)), nil
			},
			types.String,
			"Removes all spaces from the beginning (left-hand side) of `val`.",
			tree.VolatilityImmutable,
		),
	),

	"rtrim": makeBuiltin(defProps(),
		stringOverload2(
			"input",
			"trim_chars",
			func(_ *tree.EvalContext, s, chars string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597279)
				return tree.NewDString(strings.TrimRight(s, chars)), nil
			},
			types.String,
			"Removes any characters included in `trim_chars` from the end (right-hand "+
				"side) of `input` (applies recursively). \n\nFor example, `rtrim('doggie', 'ei')` "+
				"returns `dogg`.",
			tree.VolatilityImmutable,
		),
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597280)
				return tree.NewDString(strings.TrimRightFunc(s, unicode.IsSpace)), nil
			},
			types.String,
			"Removes all spaces from the end (right-hand side) of `val`.",
			tree.VolatilityImmutable,
		),
	),

	"reverse": makeBuiltin(defProps(),
		stringOverload1(
			func(evalCtx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597281)
				if len(s) > maxAllocatedStringSize {
					__antithesis_instrumentation__.Notify(597284)
					return nil, errStringTooLarge
				} else {
					__antithesis_instrumentation__.Notify(597285)
				}
				__antithesis_instrumentation__.Notify(597282)
				runes := []rune(s)
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					__antithesis_instrumentation__.Notify(597286)
					runes[i], runes[j] = runes[j], runes[i]
				}
				__antithesis_instrumentation__.Notify(597283)
				return tree.NewDString(string(runes)), nil
			},
			types.String,
			"Reverses the order of the string's characters.",
			tree.VolatilityImmutable,
		),
	),

	"replace": makeBuiltin(defProps(),
		stringOverload3(
			"input",
			"find",
			"replace",
			func(evalCtx *tree.EvalContext, input, from, to string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597287)

				var maxResultLen int64
				if len(from) == 0 {
					__antithesis_instrumentation__.Notify(597290)

					maxResultLen = int64(len(input) + (len(input)+1)*len(to))
				} else {
					__antithesis_instrumentation__.Notify(597291)
					if len(from) < len(to) {
						__antithesis_instrumentation__.Notify(597292)

						maxResultLen = int64(len(input) / len(from) * len(to))
					} else {
						__antithesis_instrumentation__.Notify(597293)

						maxResultLen = int64(len(input))
					}
				}
				__antithesis_instrumentation__.Notify(597288)
				if maxResultLen > maxAllocatedStringSize {
					__antithesis_instrumentation__.Notify(597294)
					return nil, errStringTooLarge
				} else {
					__antithesis_instrumentation__.Notify(597295)
				}
				__antithesis_instrumentation__.Notify(597289)
				result := strings.Replace(input, from, to, -1)
				return tree.NewDString(result), nil
			},
			types.String,
			"Replaces all occurrences of `find` with `replace` in `input`",
			tree.VolatilityImmutable,
		),
	),

	"translate": makeBuiltin(defProps(),
		stringOverload3("input", "find", "replace",
			func(evalCtx *tree.EvalContext, s, from, to string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597296)
				const deletionRune = utf8.MaxRune + 1
				translation := make(map[rune]rune, len(from))
				for _, fromRune := range from {
					__antithesis_instrumentation__.Notify(597299)
					toRune, size := utf8.DecodeRuneInString(to)
					if toRune == utf8.RuneError {
						__antithesis_instrumentation__.Notify(597301)
						toRune = deletionRune
					} else {
						__antithesis_instrumentation__.Notify(597302)
						to = to[size:]
					}
					__antithesis_instrumentation__.Notify(597300)
					translation[fromRune] = toRune
				}
				__antithesis_instrumentation__.Notify(597297)

				runes := make([]rune, 0, len(s))
				for _, c := range s {
					__antithesis_instrumentation__.Notify(597303)
					if t, ok := translation[c]; ok {
						__antithesis_instrumentation__.Notify(597304)
						if t != deletionRune {
							__antithesis_instrumentation__.Notify(597305)
							runes = append(runes, t)
						} else {
							__antithesis_instrumentation__.Notify(597306)
						}
					} else {
						__antithesis_instrumentation__.Notify(597307)
						runes = append(runes, c)
					}
				}
				__antithesis_instrumentation__.Notify(597298)
				return tree.NewDString(string(runes)), nil
			},
			types.String,
			"In `input`, replaces the first character from `find` with the first "+
				"character in `replace`; repeat for each character in `find`. \n\nFor example, "+
				"`translate('doggie', 'dog', '123');` returns `1233ie`.",
			tree.VolatilityImmutable,
		),
	),

	"regexp_extract": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.String}, {"regex", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597308)
				s := string(tree.MustBeDString(args[0]))
				pattern := string(tree.MustBeDString(args[1]))
				return regexpExtract(ctx, s, pattern, `\`)
			},
			Info:       "Returns the first match for the Regular Expression `regex` in `input`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"regexp_replace": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"input", types.String},
				{"regex", types.String},
				{"replace", types.String},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597309)
				s := string(tree.MustBeDString(args[0]))
				pattern := string(tree.MustBeDString(args[1]))
				to := string(tree.MustBeDString(args[2]))
				result, err := regexpReplace(evalCtx, s, pattern, to, "")
				if err != nil {
					__antithesis_instrumentation__.Notify(597311)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597312)
				}
				__antithesis_instrumentation__.Notify(597310)
				return result, nil
			},
			Info: "Replaces matches for the Regular Expression `regex` in `input` with the " +
				"Regular Expression `replace`.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"input", types.String},
				{"regex", types.String},
				{"replace", types.String},
				{"flags", types.String},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597313)
				s := string(tree.MustBeDString(args[0]))
				pattern := string(tree.MustBeDString(args[1]))
				to := string(tree.MustBeDString(args[2]))
				sqlFlags := string(tree.MustBeDString(args[3]))
				result, err := regexpReplace(evalCtx, s, pattern, to, sqlFlags)
				if err != nil {
					__antithesis_instrumentation__.Notify(597315)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597316)
				}
				__antithesis_instrumentation__.Notify(597314)
				return result, nil
			},
			Info: "Replaces matches for the regular expression `regex` in `input` with the regular " +
				"expression `replace` using `flags`." + regexpFlagInfo,
			Volatility: tree.VolatilityImmutable,
		},
	),

	"regexp_split_to_array": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"string", types.String},
				{"pattern", types.String},
			},
			ReturnType: tree.FixedReturnType(types.MakeArray(types.String)),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597317)
				return regexpSplitToArray(ctx, args, false)
			},
			Info:       "Split string using a POSIX regular expression as the delimiter.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"string", types.String},
				{"pattern", types.String},
				{"flags", types.String},
			},
			ReturnType: tree.FixedReturnType(types.MakeArray(types.String)),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597318)
				return regexpSplitToArray(ctx, args, true)
			},
			Info:       "Split string using a POSIX regular expression as the delimiter with flags." + regexpFlagInfo,
			Volatility: tree.VolatilityImmutable,
		},
	),

	"like_escape": makeBuiltin(defProps(),
		stringOverload3(
			"unescaped", "pattern", "escape",
			func(evalCtx *tree.EvalContext, unescaped, pattern, escape string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597319)
				return tree.MatchLikeEscape(evalCtx, unescaped, pattern, escape, false)
			},
			types.Bool,
			"Matches `unescaped` with `pattern` using `escape` as an escape token.",
			tree.VolatilityImmutable,
		),
	),

	"not_like_escape": makeBuiltin(defProps(),
		stringOverload3(
			"unescaped", "pattern", "escape",
			func(evalCtx *tree.EvalContext, unescaped, pattern, escape string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597320)
				dmatch, err := tree.MatchLikeEscape(evalCtx, unescaped, pattern, escape, false)
				if err != nil {
					__antithesis_instrumentation__.Notify(597322)
					return dmatch, err
				} else {
					__antithesis_instrumentation__.Notify(597323)
				}
				__antithesis_instrumentation__.Notify(597321)
				bmatch, err := tree.GetBool(dmatch)
				return tree.MakeDBool(!bmatch), err
			},
			types.Bool,
			"Checks whether `unescaped` not matches with `pattern` using `escape` as an escape token.",
			tree.VolatilityImmutable,
		)),

	"ilike_escape": makeBuiltin(defProps(),
		stringOverload3(
			"unescaped", "pattern", "escape",
			func(evalCtx *tree.EvalContext, unescaped, pattern, escape string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597324)
				return tree.MatchLikeEscape(evalCtx, unescaped, pattern, escape, true)
			},
			types.Bool,
			"Matches case insensetively `unescaped` with `pattern` using `escape` as an escape token.",
			tree.VolatilityImmutable,
		)),

	"not_ilike_escape": makeBuiltin(defProps(),
		stringOverload3(
			"unescaped", "pattern", "escape",
			func(evalCtx *tree.EvalContext, unescaped, pattern, escape string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597325)
				dmatch, err := tree.MatchLikeEscape(evalCtx, unescaped, pattern, escape, true)
				if err != nil {
					__antithesis_instrumentation__.Notify(597327)
					return dmatch, err
				} else {
					__antithesis_instrumentation__.Notify(597328)
				}
				__antithesis_instrumentation__.Notify(597326)
				bmatch, err := tree.GetBool(dmatch)
				return tree.MakeDBool(!bmatch), err
			},
			types.Bool,
			"Checks whether `unescaped` not matches case insensetively with `pattern` using `escape` as an escape token.",
			tree.VolatilityImmutable,
		)),

	"similar_escape": makeBuiltin(
		tree.FunctionProperties{
			NullableArgs: true,
		},
		similarOverloads...),

	"similar_to_escape": makeBuiltin(
		defProps(),
		append(
			similarOverloads,
			stringOverload3(
				"unescaped", "pattern", "escape",
				func(evalCtx *tree.EvalContext, unescaped, pattern, escape string) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(597329)
					return tree.SimilarToEscape(evalCtx, unescaped, pattern, escape)
				},
				types.Bool,
				"Matches `unescaped` with `pattern` using `escape` as an escape token.",
				tree.VolatilityImmutable,
			),
		)...,
	),

	"not_similar_to_escape": makeBuiltin(defProps(),
		stringOverload3(
			"unescaped", "pattern", "escape",
			func(evalCtx *tree.EvalContext, unescaped, pattern, escape string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597330)
				dmatch, err := tree.SimilarToEscape(evalCtx, unescaped, pattern, escape)
				if err != nil {
					__antithesis_instrumentation__.Notify(597332)
					return dmatch, err
				} else {
					__antithesis_instrumentation__.Notify(597333)
				}
				__antithesis_instrumentation__.Notify(597331)
				bmatch, err := tree.GetBool(dmatch)
				return tree.MakeDBool(!bmatch), err
			},
			types.Bool,
			"Checks whether `unescaped` not matches with `pattern` using `escape` as an escape token.",
			tree.VolatilityImmutable,
		)),

	"initcap": makeBuiltin(defProps(),
		stringOverload1(
			func(evalCtx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597334)
				return tree.NewDString(strings.Title(strings.ToLower(s))), nil
			},
			types.String,
			"Capitalizes the first letter of `val`.",
			tree.VolatilityImmutable,
		)),

	"quote_ident": makeBuiltin(defProps(),
		stringOverload1(
			func(evalCtx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597335)
				var buf bytes.Buffer
				lexbase.EncodeRestrictedSQLIdent(&buf, s, lexbase.EncNoFlags)
				return tree.NewDString(buf.String()), nil
			},
			types.String,
			"Return `val` suitably quoted to serve as identifier in a SQL statement.",
			tree.VolatilityImmutable,
		)),

	"quote_literal": makeBuiltin(defProps(),
		tree.Overload{
			Types:             tree.ArgTypes{{"val", types.String}},
			ReturnType:        tree.FixedReturnType(types.String),
			PreferredOverload: true,
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597336)
				s := tree.MustBeDString(args[0])
				return tree.NewDString(lexbase.EscapeSQLString(string(s))), nil
			},
			Info:       "Return `val` suitably quoted to serve as string literal in a SQL statement.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Any}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597337)

				d := tree.UnwrapDatum(ctx, args[0])
				strD, err := tree.PerformCast(ctx, d, types.String)
				if err != nil {
					__antithesis_instrumentation__.Notify(597339)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597340)
				}
				__antithesis_instrumentation__.Notify(597338)
				return tree.NewDString(strD.String()), nil
			},
			Info:       "Coerce `val` to a string and then quote it as a literal.",
			Volatility: tree.VolatilityStable,
		},
	),

	"quote_nullable": makeBuiltin(
		tree.FunctionProperties{
			Category:     categoryString,
			NullableArgs: true,
		},
		tree.Overload{
			Types:             tree.ArgTypes{{"val", types.String}},
			ReturnType:        tree.FixedReturnType(types.String),
			PreferredOverload: true,
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597341)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597343)
					return tree.NewDString("NULL"), nil
				} else {
					__antithesis_instrumentation__.Notify(597344)
				}
				__antithesis_instrumentation__.Notify(597342)
				s := tree.MustBeDString(args[0])
				return tree.NewDString(lexbase.EscapeSQLString(string(s))), nil
			},
			Info:       "Coerce `val` to a string and then quote it as a literal. If `val` is NULL, returns 'NULL'.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Any}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597345)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597348)
					return tree.NewDString("NULL"), nil
				} else {
					__antithesis_instrumentation__.Notify(597349)
				}
				__antithesis_instrumentation__.Notify(597346)

				d := tree.UnwrapDatum(ctx, args[0])
				strD, err := tree.PerformCast(ctx, d, types.String)
				if err != nil {
					__antithesis_instrumentation__.Notify(597350)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597351)
				}
				__antithesis_instrumentation__.Notify(597347)
				return tree.NewDString(strD.String()), nil
			},
			Info:       "Coerce `val` to a string and then quote it as a literal. If `val` is NULL, returns 'NULL'.",
			Volatility: tree.VolatilityStable,
		},
	),

	"left": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.Bytes}, {"return_set", types.Int}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597352)
				bytes := []byte(*args[0].(*tree.DBytes))
				n := int(tree.MustBeDInt(args[1]))

				if n < -len(bytes) {
					__antithesis_instrumentation__.Notify(597354)
					n = 0
				} else {
					__antithesis_instrumentation__.Notify(597355)
					if n < 0 {
						__antithesis_instrumentation__.Notify(597356)
						n = len(bytes) + n
					} else {
						__antithesis_instrumentation__.Notify(597357)
						if n > len(bytes) {
							__antithesis_instrumentation__.Notify(597358)
							n = len(bytes)
						} else {
							__antithesis_instrumentation__.Notify(597359)
						}
					}
				}
				__antithesis_instrumentation__.Notify(597353)
				return tree.NewDBytes(tree.DBytes(bytes[:n])), nil
			},
			Info:       "Returns the first `return_set` bytes from `input`.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.String}, {"return_set", types.Int}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597360)
				runes := []rune(string(tree.MustBeDString(args[0])))
				n := int(tree.MustBeDInt(args[1]))

				if n < -len(runes) {
					__antithesis_instrumentation__.Notify(597362)
					n = 0
				} else {
					__antithesis_instrumentation__.Notify(597363)
					if n < 0 {
						__antithesis_instrumentation__.Notify(597364)
						n = len(runes) + n
					} else {
						__antithesis_instrumentation__.Notify(597365)
						if n > len(runes) {
							__antithesis_instrumentation__.Notify(597366)
							n = len(runes)
						} else {
							__antithesis_instrumentation__.Notify(597367)
						}
					}
				}
				__antithesis_instrumentation__.Notify(597361)
				return tree.NewDString(string(runes[:n])), nil
			},
			Info:       "Returns the first `return_set` characters from `input`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"right": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.Bytes}, {"return_set", types.Int}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597368)
				bytes := []byte(*args[0].(*tree.DBytes))
				n := int(tree.MustBeDInt(args[1]))

				if n < -len(bytes) {
					__antithesis_instrumentation__.Notify(597370)
					n = 0
				} else {
					__antithesis_instrumentation__.Notify(597371)
					if n < 0 {
						__antithesis_instrumentation__.Notify(597372)
						n = len(bytes) + n
					} else {
						__antithesis_instrumentation__.Notify(597373)
						if n > len(bytes) {
							__antithesis_instrumentation__.Notify(597374)
							n = len(bytes)
						} else {
							__antithesis_instrumentation__.Notify(597375)
						}
					}
				}
				__antithesis_instrumentation__.Notify(597369)
				return tree.NewDBytes(tree.DBytes(bytes[len(bytes)-n:])), nil
			},
			Info:       "Returns the last `return_set` bytes from `input`.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.String}, {"return_set", types.Int}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597376)
				runes := []rune(string(tree.MustBeDString(args[0])))
				n := int(tree.MustBeDInt(args[1]))

				if n < -len(runes) {
					__antithesis_instrumentation__.Notify(597378)
					n = 0
				} else {
					__antithesis_instrumentation__.Notify(597379)
					if n < 0 {
						__antithesis_instrumentation__.Notify(597380)
						n = len(runes) + n
					} else {
						__antithesis_instrumentation__.Notify(597381)
						if n > len(runes) {
							__antithesis_instrumentation__.Notify(597382)
							n = len(runes)
						} else {
							__antithesis_instrumentation__.Notify(597383)
						}
					}
				}
				__antithesis_instrumentation__.Notify(597377)
				return tree.NewDString(string(runes[len(runes)-n:])), nil
			},
			Info:       "Returns the last `return_set` characters from `input`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"random": makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597384)
				return tree.NewDFloat(tree.DFloat(rand.Float64())), nil
			},
			Info: "Returns a random floating-point number between 0 (inclusive) and 1 (exclusive). " +
				"Note that the value contains at most 53 bits of randomness.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"unique_rowid": makeBuiltin(
		tree.FunctionProperties{
			Category: categoryIDGeneration,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597385)
				return tree.NewDInt(GenerateUniqueInt(ctx.NodeID.SQLInstanceID())), nil
			},
			Info: "Returns a unique ID used by CockroachDB to generate unique row IDs if a " +
				"Primary Key isn't defined for the table. The value is a combination of the " +
				"insert timestamp and the ID of the node executing the statement, which " +
				"guarantees this combination is globally unique. However, there can be " +
				"gaps and the order is not completely guaranteed.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"unordered_unique_rowid": makeBuiltin(
		tree.FunctionProperties{
			Category: categoryIDGeneration,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597386)
				v := GenerateUniqueUnorderedID(ctx.NodeID.SQLInstanceID())
				return tree.NewDInt(v), nil
			},
			Info: "Returns a unique ID. The value is a combination of the " +
				"insert timestamp and the ID of the node executing the statement, which " +
				"guarantees this combination is globally unique. The way it is generated " +
				"there is no ordering",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"nextval": makeBuiltin(
		tree.FunctionProperties{
			Category:             categorySequences,
			DistsqlBlocklist:     true,
			HasSequenceArguments: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{SequenceNameArg, types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597387)
				name := tree.MustBeDString(args[0])
				dOid, err := tree.ParseDOid(evalCtx, string(name), types.RegClass)
				if err != nil {
					__antithesis_instrumentation__.Notify(597390)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597391)
				}
				__antithesis_instrumentation__.Notify(597388)
				res, err := evalCtx.Sequence.IncrementSequenceByID(evalCtx.Ctx(), int64(dOid.DInt))
				if err != nil {
					__antithesis_instrumentation__.Notify(597392)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597393)
				}
				__antithesis_instrumentation__.Notify(597389)
				return tree.NewDInt(tree.DInt(res)), nil
			},
			Info:       "Advances the given sequence and returns its new value.",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{SequenceNameArg, types.RegClass}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597394)
				oid := tree.MustBeDOid(args[0])
				res, err := evalCtx.Sequence.IncrementSequenceByID(evalCtx.Ctx(), int64(oid.DInt))
				if err != nil {
					__antithesis_instrumentation__.Notify(597396)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597397)
				}
				__antithesis_instrumentation__.Notify(597395)
				return tree.NewDInt(tree.DInt(res)), nil
			},
			Info:       "Advances the given sequence and returns its new value.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"currval": makeBuiltin(
		tree.FunctionProperties{
			Category:             categorySequences,
			DistsqlBlocklist:     true,
			HasSequenceArguments: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{SequenceNameArg, types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597398)
				name := tree.MustBeDString(args[0])
				dOid, err := tree.ParseDOid(evalCtx, string(name), types.RegClass)
				if err != nil {
					__antithesis_instrumentation__.Notify(597401)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597402)
				}
				__antithesis_instrumentation__.Notify(597399)
				res, err := evalCtx.Sequence.GetLatestValueInSessionForSequenceByID(evalCtx.Ctx(), int64(dOid.DInt))
				if err != nil {
					__antithesis_instrumentation__.Notify(597403)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597404)
				}
				__antithesis_instrumentation__.Notify(597400)
				return tree.NewDInt(tree.DInt(res)), nil
			},
			Info:       "Returns the latest value obtained with nextval for this sequence in this session.",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{SequenceNameArg, types.RegClass}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597405)
				oid := tree.MustBeDOid(args[0])
				res, err := evalCtx.Sequence.GetLatestValueInSessionForSequenceByID(evalCtx.Ctx(), int64(oid.DInt))
				if err != nil {
					__antithesis_instrumentation__.Notify(597407)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597408)
				}
				__antithesis_instrumentation__.Notify(597406)
				return tree.NewDInt(tree.DInt(res)), nil
			},
			Info:       "Returns the latest value obtained with nextval for this sequence in this session.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"lastval": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySequences,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597409)
				val, err := evalCtx.SessionData().SequenceState.GetLastValue()
				if err != nil {
					__antithesis_instrumentation__.Notify(597411)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597412)
				}
				__antithesis_instrumentation__.Notify(597410)
				return tree.NewDInt(tree.DInt(val)), nil
			},
			Info:       "Return value most recently obtained with nextval in this session.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"setval": makeBuiltin(
		tree.FunctionProperties{
			Category:             categorySequences,
			DistsqlBlocklist:     true,
			HasSequenceArguments: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{SequenceNameArg, types.String}, {"value", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597413)
				name := tree.MustBeDString(args[0])
				dOid, err := tree.ParseDOid(evalCtx, string(name), types.RegClass)
				if err != nil {
					__antithesis_instrumentation__.Notify(597416)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597417)
				}
				__antithesis_instrumentation__.Notify(597414)

				newVal := tree.MustBeDInt(args[1])
				if err := evalCtx.Sequence.SetSequenceValueByID(
					evalCtx.Ctx(), uint32(dOid.DInt), int64(newVal), true); err != nil {
					__antithesis_instrumentation__.Notify(597418)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597419)
				}
				__antithesis_instrumentation__.Notify(597415)
				return args[1], nil
			},
			Info: "Set the given sequence's current value. The next call to nextval will return " +
				"`value + Increment`",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{SequenceNameArg, types.RegClass}, {"value", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597420)
				oid := tree.MustBeDOid(args[0])
				newVal := tree.MustBeDInt(args[1])
				if err := evalCtx.Sequence.SetSequenceValueByID(
					evalCtx.Ctx(), uint32(oid.DInt), int64(newVal), true); err != nil {
					__antithesis_instrumentation__.Notify(597422)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597423)
				}
				__antithesis_instrumentation__.Notify(597421)
				return args[1], nil
			},
			Info: "Set the given sequence's current value. The next call to nextval will return " +
				"`value + Increment`",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{SequenceNameArg, types.String}, {"value", types.Int}, {"is_called", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597424)
				name := tree.MustBeDString(args[0])
				dOid, err := tree.ParseDOid(evalCtx, string(name), types.RegClass)
				if err != nil {
					__antithesis_instrumentation__.Notify(597427)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597428)
				}
				__antithesis_instrumentation__.Notify(597425)
				isCalled := bool(tree.MustBeDBool(args[2]))

				newVal := tree.MustBeDInt(args[1])
				if err := evalCtx.Sequence.SetSequenceValueByID(
					evalCtx.Ctx(), uint32(dOid.DInt), int64(newVal), isCalled); err != nil {
					__antithesis_instrumentation__.Notify(597429)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597430)
				}
				__antithesis_instrumentation__.Notify(597426)
				return args[1], nil
			},
			Info: "Set the given sequence's current value. If is_called is false, the next call to " +
				"nextval will return `value`; otherwise `value + Increment`.",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{SequenceNameArg, types.RegClass}, {"value", types.Int}, {"is_called", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597431)
				oid := tree.MustBeDOid(args[0])
				isCalled := bool(tree.MustBeDBool(args[2]))

				newVal := tree.MustBeDInt(args[1])
				if err := evalCtx.Sequence.SetSequenceValueByID(
					evalCtx.Ctx(), uint32(oid.DInt), int64(newVal), isCalled); err != nil {
					__antithesis_instrumentation__.Notify(597433)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597434)
				}
				__antithesis_instrumentation__.Notify(597432)
				return args[1], nil
			},
			Info: "Set the given sequence's current value. If is_called is false, the next call to " +
				"nextval will return `value`; otherwise `value + Increment`.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"experimental_uuid_v4": uuidV4Impl,
	"uuid_v4":              uuidV4Impl,

	"greatest": makeBuiltin(
		tree.FunctionProperties{
			Category:     categoryComparison,
			NullableArgs: true,
		},
		tree.Overload{
			Types:      tree.HomogeneousType{},
			ReturnType: tree.FirstNonNullReturnType(),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597435)
				return tree.PickFromTuple(ctx, true, args)
			},
			Info:       "Returns the element with the greatest value.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"least": makeBuiltin(
		tree.FunctionProperties{
			Category:     categoryComparison,
			NullableArgs: true,
		},
		tree.Overload{
			Types:      tree.HomogeneousType{},
			ReturnType: tree.FirstNonNullReturnType(),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597436)
				return tree.PickFromTuple(ctx, false, args)
			},
			Info:       "Returns the element with the lowest value.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"experimental_strftime": makeBuiltin(
		tree.FunctionProperties{
			Category: categoryDateAndTime,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.Timestamp}, {"extract_format", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597437)
				fromTime := args[0].(*tree.DTimestamp).Time
				format := string(tree.MustBeDString(args[1]))
				t, err := strtime.Strftime(fromTime, format)
				if err != nil {
					__antithesis_instrumentation__.Notify(597439)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597440)
				}
				__antithesis_instrumentation__.Notify(597438)
				return tree.NewDString(t), nil
			},
			Info: "From `input`, extracts and formats the time as identified in `extract_format` " +
				"using standard `strftime` notation (though not all formatting is supported).",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.Date}, {"extract_format", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597441)
				fromTime, err := args[0].(*tree.DDate).ToTime()
				if err != nil {
					__antithesis_instrumentation__.Notify(597444)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597445)
				}
				__antithesis_instrumentation__.Notify(597442)
				format := string(tree.MustBeDString(args[1]))
				t, err := strtime.Strftime(fromTime, format)
				if err != nil {
					__antithesis_instrumentation__.Notify(597446)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597447)
				}
				__antithesis_instrumentation__.Notify(597443)
				return tree.NewDString(t), nil
			},
			Info: "From `input`, extracts and formats the time as identified in `extract_format` " +
				"using standard `strftime` notation (though not all formatting is supported).",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.TimestampTZ}, {"extract_format", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597448)
				fromTime := args[0].(*tree.DTimestampTZ).Time
				format := string(tree.MustBeDString(args[1]))
				t, err := strtime.Strftime(fromTime, format)
				if err != nil {
					__antithesis_instrumentation__.Notify(597450)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597451)
				}
				__antithesis_instrumentation__.Notify(597449)
				return tree.NewDString(t), nil
			},
			Info: "From `input`, extracts and formats the time as identified in `extract_format` " +
				"using standard `strftime` notation (though not all formatting is supported).",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"experimental_strptime": makeBuiltin(
		tree.FunctionProperties{
			Category: categoryDateAndTime,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.String}, {"format", types.String}},
			ReturnType: tree.FixedReturnType(types.TimestampTZ),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597452)
				toParse := string(tree.MustBeDString(args[0]))
				format := string(tree.MustBeDString(args[1]))
				t, err := strtime.Strptime(toParse, format)
				if err != nil {
					__antithesis_instrumentation__.Notify(597454)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597455)
				}
				__antithesis_instrumentation__.Notify(597453)
				return tree.MakeDTimestampTZ(t.UTC(), time.Microsecond)
			},
			Info: "Returns `input` as a timestamptz using `format` (which uses standard " +
				"`strptime` formatting).",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"to_char": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"interval", types.Interval}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597456)
				d := tree.MustBeDInterval(args[0]).Duration
				var buf bytes.Buffer
				d.FormatWithStyle(&buf, duration.IntervalStyle_POSTGRES)
				return tree.NewDString(buf.String()), nil
			},
			Info:       "Convert an interval to a string assuming the Postgres IntervalStyle.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"timestamp", types.Timestamp}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597457)
				ts := tree.MustBeDTimestamp(args[0])
				return tree.NewDString(tree.AsStringWithFlags(&ts, tree.FmtBareStrings)), nil
			},
			Info:       "Convert an timestamp to a string assuming the ISO, MDY DateStyle.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"date", types.Date}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597458)
				ts := tree.MustBeDDate(args[0])
				return tree.NewDString(tree.AsStringWithFlags(&ts, tree.FmtBareStrings)), nil
			},
			Info:       "Convert an date to a string assuming the ISO, MDY DateStyle.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"to_char_with_style": makeBuiltin(
		defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"interval", types.Interval}, {"style", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597459)
				d := tree.MustBeDInterval(args[0]).Duration
				styleStr := string(tree.MustBeDString(args[1]))
				styleVal, ok := duration.IntervalStyle_value[strings.ToUpper(styleStr)]
				if !ok {
					__antithesis_instrumentation__.Notify(597461)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"invalid IntervalStyle: %s",
						styleStr,
					)
				} else {
					__antithesis_instrumentation__.Notify(597462)
				}
				__antithesis_instrumentation__.Notify(597460)
				var buf bytes.Buffer
				d.FormatWithStyle(&buf, duration.IntervalStyle(styleVal))
				return tree.NewDString(buf.String()), nil
			},
			Info:       "Convert an interval to a string using the given IntervalStyle.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"timestamp", types.Timestamp}, {"datestyle", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597463)
				ts := tree.MustBeDTimestamp(args[0])
				dateStyleStr := string(tree.MustBeDString(args[1]))
				ds, err := pgdate.ParseDateStyle(dateStyleStr, pgdate.DefaultDateStyle())
				if err != nil {
					__antithesis_instrumentation__.Notify(597466)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597467)
				}
				__antithesis_instrumentation__.Notify(597464)
				if ds.Style != pgdate.Style_ISO {
					__antithesis_instrumentation__.Notify(597468)
					return nil, unimplemented.NewWithIssue(41773, "only ISO style is supported")
				} else {
					__antithesis_instrumentation__.Notify(597469)
				}
				__antithesis_instrumentation__.Notify(597465)
				return tree.NewDString(tree.AsStringWithFlags(&ts, tree.FmtBareStrings)), nil
			},
			Info:       "Convert an timestamp to a string assuming the string is formatted using the given DateStyle.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"date", types.Date}, {"datestyle", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597470)
				ts := tree.MustBeDDate(args[0])
				dateStyleStr := string(tree.MustBeDString(args[1]))
				ds, err := pgdate.ParseDateStyle(dateStyleStr, pgdate.DefaultDateStyle())
				if err != nil {
					__antithesis_instrumentation__.Notify(597473)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597474)
				}
				__antithesis_instrumentation__.Notify(597471)
				if ds.Style != pgdate.Style_ISO {
					__antithesis_instrumentation__.Notify(597475)
					return nil, unimplemented.NewWithIssue(41773, "only ISO style is supported")
				} else {
					__antithesis_instrumentation__.Notify(597476)
				}
				__antithesis_instrumentation__.Notify(597472)
				return tree.NewDString(tree.AsStringWithFlags(&ts, tree.FmtBareStrings)), nil
			},
			Info:       "Convert an date to a string assuming the string is formatted using the given DateStyle.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"age": makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.TimestampTZ}},
			ReturnType: tree.FixedReturnType(types.Interval),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597477)
				return &tree.DInterval{
					Duration: duration.Age(
						ctx.GetTxnTimestamp(time.Microsecond).Time,
						args[0].(*tree.DTimestampTZ).Time,
					),
				}, nil
			},
			Info: "Calculates the interval between `val` and the current time, normalized into years, months and days." + `

Note this may not be an accurate time span since years and months are normalized
from days, and years and months are out of context. To avoid normalizing days into
months and years, use ` + "`now() - timestamptz`" + `.`,
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"end", types.TimestampTZ}, {"begin", types.TimestampTZ}},
			ReturnType: tree.FixedReturnType(types.Interval),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597478)
				return &tree.DInterval{
					Duration: duration.Age(
						args[0].(*tree.DTimestampTZ).Time,
						args[1].(*tree.DTimestampTZ).Time,
					),
				}, nil
			},
			Info: "Calculates the interval between `begin` and `end`, normalized into years, months and days." + `

Note this may not be an accurate time span since years and months are normalized
from days, and years and months are out of context. To avoid normalizing days into
months and years, use the timestamptz subtraction operator.`,
			Volatility: tree.VolatilityImmutable,
		},
	),

	"current_date": makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Date),
			Fn:         currentDate,
			Info:       "Returns the date of the current transaction." + txnTSContextDoc,
			Volatility: tree.VolatilityStable,
		},
	),

	"now":                   txnTSImplBuiltin(true),
	"current_time":          txnTimeWithPrecisionBuiltin(true),
	"current_timestamp":     txnTSWithPrecisionImplBuiltin(true),
	"transaction_timestamp": txnTSImplBuiltin(true),

	"localtimestamp": txnTSWithPrecisionImplBuiltin(false),
	"localtime":      txnTimeWithPrecisionBuiltin(false),

	"statement_timestamp": makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types:             tree.ArgTypes{},
			ReturnType:        tree.FixedReturnType(types.TimestampTZ),
			PreferredOverload: true,
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597479)
				return tree.MakeDTimestampTZ(ctx.GetStmtTimestamp(), time.Microsecond)
			},
			Info:       "Returns the start time of the current statement.",
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Timestamp),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597480)
				return tree.MakeDTimestamp(ctx.GetStmtTimestamp(), time.Microsecond)
			},
			Info:       "Returns the start time of the current statement.",
			Volatility: tree.VolatilityStable,
		},
	),

	tree.FollowerReadTimestampFunctionName: makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.TimestampTZ),
			Fn:         followerReadTimestamp,
			Info: fmt.Sprintf(`Returns a timestamp which is very likely to be safe to perform
against a follower replica.

This function is intended to be used with an AS OF SYSTEM TIME clause to perform
historical reads against a time which is recent but sufficiently old for reads
to be performed against the closest replica as opposed to the currently
leaseholder for a given range.

Note that this function requires an enterprise license on a CCL distribution to
return a result that is less likely the closest replica. It is otherwise
hardcoded as %s from the statement time, which may not result in reading from the
nearest replica.`, defaultFollowerReadDuration),
			Volatility: tree.VolatilityVolatile,
		},
	),

	tree.FollowerReadTimestampExperimentalFunctionName: makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.TimestampTZ),
			Fn:         followerReadTimestamp,
			Info:       fmt.Sprintf("Same as %s. This name is deprecated.", tree.FollowerReadTimestampFunctionName),
			Volatility: tree.VolatilityVolatile,
		},
	),

	tree.WithMinTimestampFunctionName: makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types: tree.ArgTypes{
				{"min_timestamp", types.TimestampTZ},
			},
			ReturnType: tree.FixedReturnType(types.TimestampTZ),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597481)
				return withMinTimestamp(ctx, tree.MustBeDTimestampTZ(args[0]).Time)
			},
			Info:       withMinTimestampInfo(false),
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"min_timestamp", types.TimestampTZ},
				{"nearest_only", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.TimestampTZ),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597482)
				return withMinTimestamp(ctx, tree.MustBeDTimestampTZ(args[0]).Time)
			},
			Info:       withMinTimestampInfo(true),
			Volatility: tree.VolatilityVolatile,
		},
	),

	tree.WithMaxStalenessFunctionName: makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types: tree.ArgTypes{
				{"max_staleness", types.Interval},
			},
			ReturnType: tree.FixedReturnType(types.TimestampTZ),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597483)
				return withMaxStaleness(ctx, tree.MustBeDInterval(args[0]).Duration)
			},
			Info:       withMaxStalenessInfo(false),
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"max_staleness", types.Interval},
				{"nearest_only", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.TimestampTZ),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597484)
				return withMaxStaleness(ctx, tree.MustBeDInterval(args[0]).Duration)
			},
			Info:       withMaxStalenessInfo(true),
			Volatility: tree.VolatilityVolatile,
		},
	),

	"cluster_logical_timestamp": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Decimal),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597485)
				return ctx.GetClusterTimestamp(), nil
			},
			Info: `Returns the logical time of the current transaction as
a CockroachDB HLC in decimal form.

Note that uses of this function disable server-side optimizations and
may increase either contention or retry errors, or both.`,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"hlc_to_timestamp": makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types:      tree.ArgTypes{{"hlc", types.Decimal}},
			ReturnType: tree.FixedReturnType(types.TimestampTZ),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597486)
				d := tree.MustBeDDecimal(args[0])
				return tree.DecimalToInexactDTimestampTZ(&d)
			},
			Info: `Returns a TimestampTZ representation of a CockroachDB HLC in decimal form.

Note that a TimestampTZ has less precision than a CockroachDB HLC. It is intended as
a convenience function to display HLCs in a print-friendly form. Use the decimal
value if you rely on the HLC for accuracy.`,
			Volatility: tree.VolatilityImmutable,
		},
	),

	"clock_timestamp": makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types:             tree.ArgTypes{},
			ReturnType:        tree.FixedReturnType(types.TimestampTZ),
			PreferredOverload: true,
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597487)
				return tree.MakeDTimestampTZ(timeutil.Now(), time.Microsecond)
			},
			Info:       "Returns the current system time on one of the cluster nodes.",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Timestamp),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597488)
				return tree.MakeDTimestamp(timeutil.Now(), time.Microsecond)
			},
			Info:       "Returns the current system time on one of the cluster nodes.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"timeofday": makeBuiltin(
		tree.FunctionProperties{Category: categoryDateAndTime},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597489)
				ctxTime := ctx.GetRelativeParseTime()

				return tree.NewDString(ctxTime.Format("Mon Jan 2 15:04:05.000000 2006 -0700")), nil
			},
			Info:       "Returns the current system time on one of the cluster nodes as a string.",
			Volatility: tree.VolatilityStable,
		},
	),

	"extract":   extractBuiltin,
	"date_part": extractBuiltin,

	"extract_duration": makeBuiltin(
		tree.FunctionProperties{Category: categoryDateAndTime},
		tree.Overload{
			Types:      tree.ArgTypes{{"element", types.String}, {"input", types.Interval}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597490)

				fromInterval := *args[1].(*tree.DInterval)
				timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
				switch timeSpan {
				case "hour", "hours":
					__antithesis_instrumentation__.Notify(597491)
					return tree.NewDInt(tree.DInt(fromInterval.Nanos() / int64(time.Hour))), nil

				case "minute", "minutes":
					__antithesis_instrumentation__.Notify(597492)
					return tree.NewDInt(tree.DInt(fromInterval.Nanos() / int64(time.Minute))), nil

				case "second", "seconds":
					__antithesis_instrumentation__.Notify(597493)
					return tree.NewDInt(tree.DInt(fromInterval.Nanos() / int64(time.Second))), nil

				case "millisecond", "milliseconds":
					__antithesis_instrumentation__.Notify(597494)

					return tree.NewDInt(tree.DInt(fromInterval.Nanos() / int64(time.Millisecond))), nil

				case "microsecond", "microseconds":
					__antithesis_instrumentation__.Notify(597495)
					return tree.NewDInt(tree.DInt(fromInterval.Nanos() / int64(time.Microsecond))), nil

				default:
					__antithesis_instrumentation__.Notify(597496)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue, "unsupported timespan: %s", timeSpan)
				}
			},
			Info: "Extracts `element` from `input`.\n" +
				"Compatible elements: hour, minute, second, millisecond, microsecond.\n" +
				"This is deprecated in favor of `extract` which supports duration.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"date_trunc": makeBuiltin(
		tree.FunctionProperties{Category: categoryDateAndTime},
		tree.Overload{
			Types:      tree.ArgTypes{{"element", types.String}, {"input", types.Timestamp}},
			ReturnType: tree.FixedReturnType(types.Timestamp),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597497)
				timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
				fromTS := args[1].(*tree.DTimestamp)
				tsTZ, err := truncateTimestamp(fromTS.Time, timeSpan)
				if err != nil {
					__antithesis_instrumentation__.Notify(597499)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597500)
				}
				__antithesis_instrumentation__.Notify(597498)
				return tree.MakeDTimestamp(tsTZ.Time, time.Microsecond)
			},
			Info: "Truncates `input` to precision `element`.  Sets all fields that are less\n" +
				"significant than `element` to zero (or one, for day and month)\n\n" +
				"Compatible elements: millennium, century, decade, year, quarter, month,\n" +
				"week, day, hour, minute, second, millisecond, microsecond.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"element", types.String}, {"input", types.Date}},
			ReturnType: tree.FixedReturnType(types.TimestampTZ),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597501)
				timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
				date := args[1].(*tree.DDate)

				fromTSTZ, err := tree.MakeDTimestampTZFromDate(ctx.GetLocation(), date)
				if err != nil {
					__antithesis_instrumentation__.Notify(597503)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597504)
				}
				__antithesis_instrumentation__.Notify(597502)
				return truncateTimestamp(fromTSTZ.Time, timeSpan)
			},
			Info: "Truncates `input` to precision `element`.  Sets all fields that are less\n" +
				"significant than `element` to zero (or one, for day and month)\n\n" +
				"Compatible elements: millennium, century, decade, year, quarter, month,\n" +
				"week, day, hour, minute, second, millisecond, microsecond.",
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"element", types.String}, {"input", types.Time}},
			ReturnType: tree.FixedReturnType(types.Interval),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597505)
				timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
				fromTime := args[1].(*tree.DTime)
				time, err := truncateTime(fromTime, timeSpan)
				if err != nil {
					__antithesis_instrumentation__.Notify(597507)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597508)
				}
				__antithesis_instrumentation__.Notify(597506)
				return &tree.DInterval{Duration: duration.MakeDuration(int64(*time)*1000, 0, 0)}, nil
			},
			Info: "Truncates `input` to precision `element`.  Sets all fields that are less\n" +
				"significant than `element` to zero.\n\n" +
				"Compatible elements: hour, minute, second, millisecond, microsecond.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"element", types.String}, {"input", types.TimestampTZ}},
			ReturnType: tree.FixedReturnType(types.TimestampTZ),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597509)
				fromTSTZ := args[1].(*tree.DTimestampTZ)
				timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
				return truncateTimestamp(fromTSTZ.Time.In(ctx.GetLocation()), timeSpan)
			},
			Info: "Truncates `input` to precision `element`.  Sets all fields that are less\n" +
				"significant than `element` to zero (or one, for day and month)\n\n" +
				"Compatible elements: millennium, century, decade, year, quarter, month,\n" +
				"week, day, hour, minute, second, millisecond, microsecond.",
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"element", types.String}, {"input", types.Interval}},
			ReturnType: tree.FixedReturnType(types.Interval),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597510)
				fromInterval := args[1].(*tree.DInterval)
				timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
				return truncateInterval(fromInterval, timeSpan)
			},
			Info: "Truncates `input` to precision `element`.  Sets all fields that are less\n" +
				"significant than `element` to zero (or one, for day and month)\n\n" +
				"Compatible elements: millennium, century, decade, year, quarter, month,\n" +
				"week, day, hour, minute, second, millisecond, microsecond.",
			Volatility: tree.VolatilityStable,
		},
	),

	"row_to_json": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"row", types.AnyTuple}},
			ReturnType: tree.FixedReturnType(types.Jsonb),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597511)
				tuple := args[0].(*tree.DTuple)
				builder := json.NewObjectBuilder(len(tuple.D))
				typ := tuple.ResolvedType()
				labels := typ.TupleLabels()
				for i, d := range tuple.D {
					__antithesis_instrumentation__.Notify(597513)
					var label string
					if labels != nil {
						__antithesis_instrumentation__.Notify(597517)
						label = labels[i]
					} else {
						__antithesis_instrumentation__.Notify(597518)
					}
					__antithesis_instrumentation__.Notify(597514)
					if label == "" {
						__antithesis_instrumentation__.Notify(597519)
						label = fmt.Sprintf("f%d", i+1)
					} else {
						__antithesis_instrumentation__.Notify(597520)
					}
					__antithesis_instrumentation__.Notify(597515)
					val, err := tree.AsJSON(
						d,
						ctx.SessionData().DataConversionConfig,
						ctx.GetLocation(),
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(597521)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597522)
					}
					__antithesis_instrumentation__.Notify(597516)
					builder.Add(label, val)
				}
				__antithesis_instrumentation__.Notify(597512)
				return tree.NewDJSON(builder.Build()), nil
			},
			Info:       "Returns the row as a JSON object.",
			Volatility: tree.VolatilityStable,
		},
	),

	"timezone": makeBuiltin(defProps(),

		tree.Overload{
			Types: tree.ArgTypes{
				{"timezone", types.String},
				{"timestamptz_string", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Timestamp),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597523)
				tzArg := string(tree.MustBeDString(args[0]))
				tsArg := string(tree.MustBeDString(args[1]))
				ts, _, err := tree.ParseDTimestampTZ(ctx, tsArg, time.Microsecond)
				if err != nil {
					__antithesis_instrumentation__.Notify(597526)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597527)
				}
				__antithesis_instrumentation__.Notify(597524)
				loc, err := timeutil.TimeZoneStringToLocation(tzArg, timeutil.TimeZoneStringToLocationPOSIXStandard)
				if err != nil {
					__antithesis_instrumentation__.Notify(597528)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597529)
				}
				__antithesis_instrumentation__.Notify(597525)
				return ts.EvalAtTimeZone(ctx, loc)
			},
			Info:       "Convert given time stamp with time zone to the new time zone, with no time zone designation.",
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"timezone", types.String},
				{"timestamp", types.Timestamp},
			},
			ReturnType: tree.FixedReturnType(types.TimestampTZ),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597530)
				tzStr := string(tree.MustBeDString(args[0]))
				ts := tree.MustBeDTimestamp(args[1])
				loc, err := timeutil.TimeZoneStringToLocation(
					tzStr,
					timeutil.TimeZoneStringToLocationPOSIXStandard,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(597532)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597533)
				}
				__antithesis_instrumentation__.Notify(597531)
				_, beforeOffsetSecs := ts.Time.Zone()
				_, afterOffsetSecs := ts.Time.In(loc).Zone()
				durationDelta := time.Duration(beforeOffsetSecs-afterOffsetSecs) * time.Second
				return tree.MakeDTimestampTZ(ts.Time.Add(durationDelta), time.Microsecond)
			},
			Info:       "Treat given time stamp without time zone as located in the specified time zone.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"timezone", types.String},
				{"timestamptz", types.TimestampTZ},
			},
			ReturnType: tree.FixedReturnType(types.Timestamp),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597534)
				tzStr := string(tree.MustBeDString(args[0]))
				ts := tree.MustBeDTimestampTZ(args[1])
				loc, err := timeutil.TimeZoneStringToLocation(
					tzStr,
					timeutil.TimeZoneStringToLocationPOSIXStandard,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(597536)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597537)
				}
				__antithesis_instrumentation__.Notify(597535)
				return ts.EvalAtTimeZone(ctx, loc)
			},
			Info:       "Convert given time stamp with time zone to the new time zone, with no time zone designation.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"timezone", types.String},
				{"time", types.Time},
			},
			ReturnType: tree.FixedReturnType(types.TimeTZ),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597538)
				tzStr := string(tree.MustBeDString(args[0]))
				tArg := args[1].(*tree.DTime)
				loc, err := timeutil.TimeZoneStringToLocation(
					tzStr,
					timeutil.TimeZoneStringToLocationPOSIXStandard,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(597540)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597541)
				}
				__antithesis_instrumentation__.Notify(597539)
				tTime := timeofday.TimeOfDay(*tArg).ToTime()
				_, beforeOffsetSecs := tTime.In(ctx.GetLocation()).Zone()
				durationDelta := time.Duration(-beforeOffsetSecs) * time.Second
				return tree.NewDTimeTZ(timetz.MakeTimeTZFromTime(tTime.In(loc).Add(durationDelta))), nil
			},
			Info:       "Treat given time without time zone as located in the specified time zone.",
			Volatility: tree.VolatilityStable,

			IgnoreVolatilityCheck: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"timezone", types.String},
				{"timetz", types.TimeTZ},
			},
			ReturnType: tree.FixedReturnType(types.TimeTZ),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597542)

				tzStr := string(tree.MustBeDString(args[0]))
				tArg := args[1].(*tree.DTimeTZ)
				loc, err := timeutil.TimeZoneStringToLocation(
					tzStr,
					timeutil.TimeZoneStringToLocationPOSIXStandard,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(597544)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597545)
				}
				__antithesis_instrumentation__.Notify(597543)
				tTime := tArg.TimeTZ.ToTime()
				return tree.NewDTimeTZ(timetz.MakeTimeTZFromTime(tTime.In(loc))), nil
			},
			Info:       "Convert given time with time zone to the new time zone.",
			Volatility: tree.VolatilityStable,

			IgnoreVolatilityCheck: true,
		},
	),

	"parse_timestamp": makeBuiltin(
		defProps(),
		stringOverload1(
			func(ctx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597546)
				ts, dependsOnContext, err := tree.ParseDTimestamp(
					tree.NewParseTimeContext(ctx.GetTxnTimestamp(time.Microsecond).Time),
					s,
					time.Microsecond,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(597549)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597550)
				}
				__antithesis_instrumentation__.Notify(597547)
				if dependsOnContext {
					__antithesis_instrumentation__.Notify(597551)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"relative timestamps are not supported",
					)
				} else {
					__antithesis_instrumentation__.Notify(597552)
				}
				__antithesis_instrumentation__.Notify(597548)
				return ts, nil
			},
			types.Timestamp,
			"Convert a string containing an absolute timestamp to the corresponding timestamp assuming dates are in MDY format.",
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"string", types.String}, {"datestyle", types.String}},
			ReturnType: tree.FixedReturnType(types.Timestamp),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597553)
				arg := string(tree.MustBeDString(args[0]))
				dateStyle := string(tree.MustBeDString(args[1]))
				parseCtx, err := parseContextFromDateStyle(ctx, dateStyle)
				if err != nil {
					__antithesis_instrumentation__.Notify(597557)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597558)
				}
				__antithesis_instrumentation__.Notify(597554)
				ts, dependsOnContext, err := tree.ParseDTimestamp(
					parseCtx,
					arg,
					time.Microsecond,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(597559)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597560)
				}
				__antithesis_instrumentation__.Notify(597555)
				if dependsOnContext {
					__antithesis_instrumentation__.Notify(597561)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"relative timestamps are not supported",
					)
				} else {
					__antithesis_instrumentation__.Notify(597562)
				}
				__antithesis_instrumentation__.Notify(597556)
				return ts, nil
			},
			Info:       "Convert a string containing an absolute timestamp to the corresponding timestamp assuming dates formatted using the given DateStyle.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"parse_date": makeBuiltin(
		defProps(),
		stringOverload1(
			func(ctx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597563)
				ts, dependsOnContext, err := tree.ParseDDate(
					tree.NewParseTimeContext(ctx.GetTxnTimestamp(time.Microsecond).Time),
					s,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(597566)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597567)
				}
				__antithesis_instrumentation__.Notify(597564)
				if dependsOnContext {
					__antithesis_instrumentation__.Notify(597568)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"relative dates are not supported",
					)
				} else {
					__antithesis_instrumentation__.Notify(597569)
				}
				__antithesis_instrumentation__.Notify(597565)
				return ts, nil
			},
			types.Date,
			"Parses a date assuming it is in MDY format.",
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"string", types.String}, {"datestyle", types.String}},
			ReturnType: tree.FixedReturnType(types.Date),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597570)
				arg := string(tree.MustBeDString(args[0]))
				dateStyle := string(tree.MustBeDString(args[1]))
				parseCtx, err := parseContextFromDateStyle(ctx, dateStyle)
				if err != nil {
					__antithesis_instrumentation__.Notify(597574)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597575)
				}
				__antithesis_instrumentation__.Notify(597571)
				ts, dependsOnContext, err := tree.ParseDDate(parseCtx, arg)
				if err != nil {
					__antithesis_instrumentation__.Notify(597576)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597577)
				}
				__antithesis_instrumentation__.Notify(597572)
				if dependsOnContext {
					__antithesis_instrumentation__.Notify(597578)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"relative dates are not supported",
					)
				} else {
					__antithesis_instrumentation__.Notify(597579)
				}
				__antithesis_instrumentation__.Notify(597573)
				return ts, nil
			},
			Info:       "Parses a date assuming it is in format specified by DateStyle.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"parse_time": makeBuiltin(
		defProps(),
		stringOverload1(
			func(ctx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597580)
				t, dependsOnContext, err := tree.ParseDTime(
					tree.NewParseTimeContext(ctx.GetTxnTimestamp(time.Microsecond).Time),
					s,
					time.Microsecond,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(597583)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597584)
				}
				__antithesis_instrumentation__.Notify(597581)
				if dependsOnContext {
					__antithesis_instrumentation__.Notify(597585)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"relative times are not supported",
					)
				} else {
					__antithesis_instrumentation__.Notify(597586)
				}
				__antithesis_instrumentation__.Notify(597582)
				return t, nil
			},
			types.Time,
			"Parses a time assuming the date (if any) is in MDY format.",
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"string", types.String}, {"timestyle", types.String}},
			ReturnType: tree.FixedReturnType(types.Time),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597587)
				arg := string(tree.MustBeDString(args[0]))
				dateStyle := string(tree.MustBeDString(args[1]))
				parseCtx, err := parseContextFromDateStyle(ctx, dateStyle)
				if err != nil {
					__antithesis_instrumentation__.Notify(597591)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597592)
				}
				__antithesis_instrumentation__.Notify(597588)
				t, dependsOnContext, err := tree.ParseDTime(parseCtx, arg, time.Microsecond)
				if err != nil {
					__antithesis_instrumentation__.Notify(597593)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597594)
				}
				__antithesis_instrumentation__.Notify(597589)
				if dependsOnContext {
					__antithesis_instrumentation__.Notify(597595)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"relative times are not supported",
					)
				} else {
					__antithesis_instrumentation__.Notify(597596)
				}
				__antithesis_instrumentation__.Notify(597590)
				return t, nil
			},
			Info:       "Parses a time assuming the date (if any) is in format specified by DateStyle.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"parse_interval": makeBuiltin(
		defProps(),
		stringOverload1(
			func(evalCtx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597597)
				return tree.ParseDInterval(duration.IntervalStyle_POSTGRES, s)
			},
			types.Interval,
			"Convert a string to an interval assuming the Postgres IntervalStyle.",
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"string", types.String}, {"style", types.String}},
			ReturnType: tree.FixedReturnType(types.Interval),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597598)
				s := string(tree.MustBeDString(args[0]))
				styleStr := string(tree.MustBeDString(args[1]))
				styleVal, ok := duration.IntervalStyle_value[strings.ToUpper(styleStr)]
				if !ok {
					__antithesis_instrumentation__.Notify(597600)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"invalid IntervalStyle: %s",
						styleStr,
					)
				} else {
					__antithesis_instrumentation__.Notify(597601)
				}
				__antithesis_instrumentation__.Notify(597599)
				return tree.ParseDInterval(duration.IntervalStyle(styleVal), s)
			},
			Info:       "Convert a string to an interval using the given IntervalStyle.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"parse_timetz": makeBuiltin(
		defProps(),
		stringOverload1(
			func(ctx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597602)
				t, dependsOnContext, err := tree.ParseDTimeTZ(
					tree.NewParseTimeContext(ctx.GetTxnTimestamp(time.Microsecond).Time),
					s,
					time.Microsecond,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(597605)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597606)
				}
				__antithesis_instrumentation__.Notify(597603)
				if dependsOnContext {
					__antithesis_instrumentation__.Notify(597607)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"relative times are not supported",
					)
				} else {
					__antithesis_instrumentation__.Notify(597608)
				}
				__antithesis_instrumentation__.Notify(597604)
				return t, nil
			},
			types.TimeTZ,
			"Parses a timetz assuming the date (if any) is in MDY format.",
			tree.VolatilityImmutable,
		),
		tree.Overload{
			Types:      tree.ArgTypes{{"string", types.String}, {"timestyle", types.String}},
			ReturnType: tree.FixedReturnType(types.TimeTZ),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597609)
				arg := string(tree.MustBeDString(args[0]))
				dateStyle := string(tree.MustBeDString(args[1]))
				parseCtx, err := parseContextFromDateStyle(ctx, dateStyle)
				if err != nil {
					__antithesis_instrumentation__.Notify(597613)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597614)
				}
				__antithesis_instrumentation__.Notify(597610)
				t, dependsOnContext, err := tree.ParseDTimeTZ(parseCtx, arg, time.Microsecond)
				if err != nil {
					__antithesis_instrumentation__.Notify(597615)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597616)
				}
				__antithesis_instrumentation__.Notify(597611)
				if dependsOnContext {
					__antithesis_instrumentation__.Notify(597617)
					return nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"relative times are not supported",
					)
				} else {
					__antithesis_instrumentation__.Notify(597618)
				}
				__antithesis_instrumentation__.Notify(597612)
				return t, nil
			},
			Info:       "Parses a timetz assuming the date (if any) is in format specified by DateStyle.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"string_to_array": makeBuiltin(arrayPropsNullableArgs(),
		tree.Overload{
			Types:      tree.ArgTypes{{"str", types.String}, {"delimiter", types.String}},
			ReturnType: tree.FixedReturnType(types.StringArray),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597619)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597621)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597622)
				}
				__antithesis_instrumentation__.Notify(597620)
				str := string(tree.MustBeDString(args[0]))
				delimOrNil := stringOrNil(args[1])
				return stringToArray(str, delimOrNil, nil)
			},
			Info:       "Split a string into components on a delimiter.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"str", types.String}, {"delimiter", types.String}, {"null", types.String}},
			ReturnType: tree.FixedReturnType(types.StringArray),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597623)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597625)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597626)
				}
				__antithesis_instrumentation__.Notify(597624)
				str := string(tree.MustBeDString(args[0]))
				delimOrNil := stringOrNil(args[1])
				nullStr := stringOrNil(args[2])
				return stringToArray(str, delimOrNil, nullStr)
			},
			Info:       "Split a string into components on a delimiter with a specified string to consider NULL.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"array_to_string": makeBuiltin(arrayPropsNullableArgs(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.AnyArray}, {"delim", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597627)
				if args[0] == tree.DNull || func() bool {
					__antithesis_instrumentation__.Notify(597629)
					return args[1] == tree.DNull == true
				}() == true {
					__antithesis_instrumentation__.Notify(597630)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597631)
				}
				__antithesis_instrumentation__.Notify(597628)
				arr := tree.MustBeDArray(args[0])
				delim := string(tree.MustBeDString(args[1]))
				return arrayToString(evalCtx, arr, delim, nil)
			},
			Info:       "Join an array into a string with a delimiter.",
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.AnyArray}, {"delimiter", types.String}, {"null", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597632)
				if args[0] == tree.DNull || func() bool {
					__antithesis_instrumentation__.Notify(597634)
					return args[1] == tree.DNull == true
				}() == true {
					__antithesis_instrumentation__.Notify(597635)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597636)
				}
				__antithesis_instrumentation__.Notify(597633)
				arr := tree.MustBeDArray(args[0])
				delim := string(tree.MustBeDString(args[1]))
				nullStr := stringOrNil(args[2])
				return arrayToString(evalCtx, arr, delim, nullStr)
			},
			Info:       "Join an array into a string with a delimiter, replacing NULLs with a null string.",
			Volatility: tree.VolatilityStable,
		},
	),

	"array_length": makeBuiltin(arrayProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.AnyArray}, {"array_dimension", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597637)
				arr := tree.MustBeDArray(args[0])
				dimen := int64(tree.MustBeDInt(args[1]))
				return arrayLength(arr, dimen), nil
			},
			Info: "Calculates the length of `input` on the provided `array_dimension`. However, " +
				"because CockroachDB doesn't yet support multi-dimensional arrays, the only supported" +
				" `array_dimension` is **1**.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"cardinality": makeBuiltin(arrayProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.AnyArray}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597638)
				arr := tree.MustBeDArray(args[0])
				return cardinality(arr), nil
			},
			Info:       "Calculates the number of elements contained in `input`",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"array_lower": makeBuiltin(arrayProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.AnyArray}, {"array_dimension", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597639)
				arr := tree.MustBeDArray(args[0])
				dimen := int64(tree.MustBeDInt(args[1]))
				return arrayLower(arr, dimen), nil
			},
			Info: "Calculates the minimum value of `input` on the provided `array_dimension`. " +
				"However, because CockroachDB doesn't yet support multi-dimensional arrays, the only " +
				"supported `array_dimension` is **1**.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"array_upper": makeBuiltin(arrayProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.AnyArray}, {"array_dimension", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597640)
				arr := tree.MustBeDArray(args[0])
				dimen := int64(tree.MustBeDInt(args[1]))
				return arrayLength(arr, dimen), nil
			},
			Info: "Calculates the maximum value of `input` on the provided `array_dimension`. " +
				"However, because CockroachDB doesn't yet support multi-dimensional arrays, the only " +
				"supported `array_dimension` is **1**.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"array_append": setProps(arrayPropsNullableArgs(), arrayBuiltin(func(typ *types.T) tree.Overload {
		__antithesis_instrumentation__.Notify(597641)
		return tree.Overload{
			Types: tree.ArgTypes{{"array", types.MakeArray(typ)}, {"elem", typ}},
			ReturnType: func(args []tree.TypedExpr) *types.T {
				__antithesis_instrumentation__.Notify(597642)
				if len(args) > 0 {
					__antithesis_instrumentation__.Notify(597644)
					if argTyp := args[0].ResolvedType(); argTyp.Family() != types.UnknownFamily {
						__antithesis_instrumentation__.Notify(597645)
						return argTyp
					} else {
						__antithesis_instrumentation__.Notify(597646)
					}
				} else {
					__antithesis_instrumentation__.Notify(597647)
				}
				__antithesis_instrumentation__.Notify(597643)
				return types.MakeArray(typ)
			},
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597648)
				return tree.AppendToMaybeNullArray(typ, args[0], args[1])
			},
			Info:       "Appends `elem` to `array`, returning the result.",
			Volatility: tree.VolatilityImmutable,
		}
	})),

	"array_prepend": setProps(arrayPropsNullableArgs(), arrayBuiltin(func(typ *types.T) tree.Overload {
		__antithesis_instrumentation__.Notify(597649)
		return tree.Overload{
			Types: tree.ArgTypes{{"elem", typ}, {"array", types.MakeArray(typ)}},
			ReturnType: func(args []tree.TypedExpr) *types.T {
				__antithesis_instrumentation__.Notify(597650)
				if len(args) > 1 {
					__antithesis_instrumentation__.Notify(597652)
					if argTyp := args[1].ResolvedType(); argTyp.Family() != types.UnknownFamily {
						__antithesis_instrumentation__.Notify(597653)
						return argTyp
					} else {
						__antithesis_instrumentation__.Notify(597654)
					}
				} else {
					__antithesis_instrumentation__.Notify(597655)
				}
				__antithesis_instrumentation__.Notify(597651)
				return types.MakeArray(typ)
			},
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597656)
				return tree.PrependToMaybeNullArray(typ, args[0], args[1])
			},
			Info:       "Prepends `elem` to `array`, returning the result.",
			Volatility: tree.VolatilityImmutable,
		}
	})),

	"array_cat": setProps(arrayPropsNullableArgs(), arrayBuiltin(func(typ *types.T) tree.Overload {
		__antithesis_instrumentation__.Notify(597657)
		return tree.Overload{
			Types: tree.ArgTypes{{"left", types.MakeArray(typ)}, {"right", types.MakeArray(typ)}},
			ReturnType: func(args []tree.TypedExpr) *types.T {
				__antithesis_instrumentation__.Notify(597658)
				if len(args) > 1 {
					__antithesis_instrumentation__.Notify(597661)
					if argTyp := args[1].ResolvedType(); argTyp.Family() != types.UnknownFamily {
						__antithesis_instrumentation__.Notify(597662)
						return argTyp
					} else {
						__antithesis_instrumentation__.Notify(597663)
					}
				} else {
					__antithesis_instrumentation__.Notify(597664)
				}
				__antithesis_instrumentation__.Notify(597659)
				if len(args) > 2 {
					__antithesis_instrumentation__.Notify(597665)
					if argTyp := args[2].ResolvedType(); argTyp.Family() != types.UnknownFamily {
						__antithesis_instrumentation__.Notify(597666)
						return argTyp
					} else {
						__antithesis_instrumentation__.Notify(597667)
					}
				} else {
					__antithesis_instrumentation__.Notify(597668)
				}
				__antithesis_instrumentation__.Notify(597660)
				return types.MakeArray(typ)
			},
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597669)
				return tree.ConcatArrays(typ, args[0], args[1])
			},
			Info:       "Appends two arrays.",
			Volatility: tree.VolatilityImmutable,
		}
	})),

	"array_remove": setProps(arrayPropsNullableArgs(), arrayBuiltin(func(typ *types.T) tree.Overload {
		__antithesis_instrumentation__.Notify(597670)
		return tree.Overload{
			Types:      tree.ArgTypes{{"array", types.MakeArray(typ)}, {"elem", typ}},
			ReturnType: tree.IdentityReturnType(0),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597671)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597674)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597675)
				}
				__antithesis_instrumentation__.Notify(597672)
				result := tree.NewDArray(typ)
				for _, e := range tree.MustBeDArray(args[0]).Array {
					__antithesis_instrumentation__.Notify(597676)
					cmp, err := e.CompareError(ctx, args[1])
					if err != nil {
						__antithesis_instrumentation__.Notify(597678)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597679)
					}
					__antithesis_instrumentation__.Notify(597677)
					if cmp != 0 {
						__antithesis_instrumentation__.Notify(597680)
						if err := result.Append(e); err != nil {
							__antithesis_instrumentation__.Notify(597681)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(597682)
						}
					} else {
						__antithesis_instrumentation__.Notify(597683)
					}
				}
				__antithesis_instrumentation__.Notify(597673)
				return result, nil
			},
			Info:       "Remove from `array` all elements equal to `elem`.",
			Volatility: tree.VolatilityImmutable,
		}
	})),

	"array_replace": setProps(arrayPropsNullableArgs(), arrayBuiltin(func(typ *types.T) tree.Overload {
		__antithesis_instrumentation__.Notify(597684)
		return tree.Overload{
			Types:      tree.ArgTypes{{"array", types.MakeArray(typ)}, {"toreplace", typ}, {"replacewith", typ}},
			ReturnType: tree.IdentityReturnType(0),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597685)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597688)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597689)
				}
				__antithesis_instrumentation__.Notify(597686)
				result := tree.NewDArray(typ)
				for _, e := range tree.MustBeDArray(args[0]).Array {
					__antithesis_instrumentation__.Notify(597690)
					cmp, err := e.CompareError(ctx, args[1])
					if err != nil {
						__antithesis_instrumentation__.Notify(597692)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597693)
					}
					__antithesis_instrumentation__.Notify(597691)
					if cmp == 0 {
						__antithesis_instrumentation__.Notify(597694)
						if err := result.Append(args[2]); err != nil {
							__antithesis_instrumentation__.Notify(597695)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(597696)
						}
					} else {
						__antithesis_instrumentation__.Notify(597697)
						if err := result.Append(e); err != nil {
							__antithesis_instrumentation__.Notify(597698)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(597699)
						}
					}
				}
				__antithesis_instrumentation__.Notify(597687)
				return result, nil
			},
			Info:       "Replace all occurrences of `toreplace` in `array` with `replacewith`.",
			Volatility: tree.VolatilityImmutable,
		}
	})),

	"array_position": setProps(arrayPropsNullableArgs(), arrayBuiltin(func(typ *types.T) tree.Overload {
		__antithesis_instrumentation__.Notify(597700)
		return tree.Overload{
			Types:      tree.ArgTypes{{"array", types.MakeArray(typ)}, {"elem", typ}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597701)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597704)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597705)
				}
				__antithesis_instrumentation__.Notify(597702)
				for i, e := range tree.MustBeDArray(args[0]).Array {
					__antithesis_instrumentation__.Notify(597706)
					cmp, err := e.CompareError(ctx, args[1])
					if err != nil {
						__antithesis_instrumentation__.Notify(597708)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597709)
					}
					__antithesis_instrumentation__.Notify(597707)
					if cmp == 0 {
						__antithesis_instrumentation__.Notify(597710)
						return tree.NewDInt(tree.DInt(i + 1)), nil
					} else {
						__antithesis_instrumentation__.Notify(597711)
					}
				}
				__antithesis_instrumentation__.Notify(597703)
				return tree.DNull, nil
			},
			Info:       "Return the index of the first occurrence of `elem` in `array`.",
			Volatility: tree.VolatilityImmutable,
		}
	})),

	"array_positions": setProps(arrayPropsNullableArgs(), arrayBuiltin(func(typ *types.T) tree.Overload {
		__antithesis_instrumentation__.Notify(597712)
		return tree.Overload{
			Types:      tree.ArgTypes{{"array", types.MakeArray(typ)}, {"elem", typ}},
			ReturnType: tree.FixedReturnType(types.IntArray),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597713)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597716)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597717)
				}
				__antithesis_instrumentation__.Notify(597714)
				result := tree.NewDArray(types.Int)
				for i, e := range tree.MustBeDArray(args[0]).Array {
					__antithesis_instrumentation__.Notify(597718)
					cmp, err := e.CompareError(ctx, args[1])
					if err != nil {
						__antithesis_instrumentation__.Notify(597720)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597721)
					}
					__antithesis_instrumentation__.Notify(597719)
					if cmp == 0 {
						__antithesis_instrumentation__.Notify(597722)
						if err := result.Append(tree.NewDInt(tree.DInt(i + 1))); err != nil {
							__antithesis_instrumentation__.Notify(597723)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(597724)
						}
					} else {
						__antithesis_instrumentation__.Notify(597725)
					}
				}
				__antithesis_instrumentation__.Notify(597715)
				return result, nil
			},
			Info:       "Returns and array of indexes of all occurrences of `elem` in `array`.",
			Volatility: tree.VolatilityImmutable,
		}
	})),

	"ts_match_qv":                    makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"ts_match_vq":                    makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"tsvector_cmp":                   makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"tsvector_concat":                makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"ts_debug":                       makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"ts_headline":                    makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"ts_lexize":                      makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"websearch_to_tsquery":           makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"array_to_tsvector":              makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"get_current_ts_config":          makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"numnode":                        makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"plainto_tsquery":                makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"phraseto_tsquery":               makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"querytree":                      makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"setweight":                      makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"strip":                          makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"to_tsquery":                     makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"to_tsvector":                    makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"json_to_tsvector":               makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"jsonb_to_tsvector":              makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"ts_delete":                      makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"ts_filter":                      makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"ts_rank":                        makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"ts_rank_cd":                     makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"ts_rewrite":                     makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"tsquery_phrase":                 makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"tsvector_to_array":              makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"tsvector_update_trigger":        makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),
	"tsvector_update_trigger_column": makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 7821, Category: categoryFullTextSearch}),

	"soundex": makeBuiltin(
		tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"source", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597726)
				s := string(tree.MustBeDString(args[0]))
				t := fuzzystrmatch.Soundex(s)
				return tree.NewDString(t), nil
			},
			Info:       "Convert a string to its Soundex code.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"difference": makeBuiltin(
		tree.FunctionProperties{Category: categoryString},
		tree.Overload{
			Types:      tree.ArgTypes{{"source", types.String}, {"target", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597727)
				s, t := string(tree.MustBeDString(args[0])), string(tree.MustBeDString(args[1]))
				diff := fuzzystrmatch.Difference(s, t)
				return tree.NewDString(strconv.Itoa(diff)), nil
			},
			Info:       "Convert two strings to their Soundex codes and then reports the number of matching code positions.",
			Volatility: tree.VolatilityImmutable,
		},
	),
	"levenshtein": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"source", types.String}, {"target", types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597728)
				s, t := string(tree.MustBeDString(args[0])), string(tree.MustBeDString(args[1]))
				const maxLen = 255
				if len(s) > maxLen || func() bool {
					__antithesis_instrumentation__.Notify(597730)
					return len(t) > maxLen == true
				}() == true {
					__antithesis_instrumentation__.Notify(597731)
					return nil, pgerror.Newf(pgcode.InvalidParameterValue,
						"levenshtein argument exceeds maximum length of %d characters", maxLen)
				} else {
					__antithesis_instrumentation__.Notify(597732)
				}
				__antithesis_instrumentation__.Notify(597729)
				ld := fuzzystrmatch.LevenshteinDistance(s, t)
				return tree.NewDInt(tree.DInt(ld)), nil
			},
			Info:       "Calculates the Levenshtein distance between two strings. Maximum input length is 255 characters.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{{"source", types.String}, {"target", types.String},
				{"ins_cost", types.Int}, {"del_cost", types.Int}, {"sub_cost", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597733)
				s, t := string(tree.MustBeDString(args[0])), string(tree.MustBeDString(args[1]))
				ins, del, sub := int(tree.MustBeDInt(args[2])), int(tree.MustBeDInt(args[3])), int(tree.MustBeDInt(args[4]))
				const maxLen = 255
				if len(s) > maxLen || func() bool {
					__antithesis_instrumentation__.Notify(597735)
					return len(t) > maxLen == true
				}() == true {
					__antithesis_instrumentation__.Notify(597736)
					return nil, pgerror.Newf(pgcode.InvalidParameterValue,
						"levenshtein argument exceeds maximum length of %d characters", maxLen)
				} else {
					__antithesis_instrumentation__.Notify(597737)
				}
				__antithesis_instrumentation__.Notify(597734)
				ld := fuzzystrmatch.LevenshteinDistanceWithCost(s, t, ins, del, sub)
				return tree.NewDInt(tree.DInt(ld)), nil
			},
			Info: "Calculates the Levenshtein distance between two strings. The cost parameters specify how much to " +
				"charge for each edit operation. Maximum input length is 255 characters.",
			Volatility: tree.VolatilityImmutable,
		}),
	"levenshtein_less_equal": makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 56820, Category: categoryFuzzyStringMatching}),
	"metaphone":              makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 56820, Category: categoryFuzzyStringMatching}),
	"dmetaphone_alt":         makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 56820, Category: categoryFuzzyStringMatching}),

	"similarity":             makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 41285, Category: categoryTrigram}),
	"show_trgm":              makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 41285, Category: categoryTrigram}),
	"word_similarity":        makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 41285, Category: categoryTrigram}),
	"strict_word_similarity": makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 41285, Category: categoryTrigram}),
	"show_limit":             makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 41285, Category: categoryTrigram}),
	"set_limit":              makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 41285, Category: categoryTrigram}),

	"json_to_recordset":  makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 33285, Category: categoryJSON}),
	"jsonb_to_recordset": makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 33285, Category: categoryJSON}),

	"jsonb_path_exists":      makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 22513, Category: categoryJSON}),
	"jsonb_path_exists_opr":  makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 22513, Category: categoryJSON}),
	"jsonb_path_match":       makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 22513, Category: categoryJSON}),
	"jsonb_path_match_opr":   makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 22513, Category: categoryJSON}),
	"jsonb_path_query":       makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 22513, Category: categoryJSON}),
	"jsonb_path_query_array": makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 22513, Category: categoryJSON}),
	"jsonb_path_query_first": makeBuiltin(tree.FunctionProperties{UnsupportedWithIssue: 22513, Category: categoryJSON}),

	"json_remove_path": makeBuiltin(jsonProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Jsonb}, {"path", types.StringArray}},
			ReturnType: tree.FixedReturnType(types.Jsonb),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597738)
				ary := *tree.MustBeDArray(args[1])
				if err := checkHasNulls(ary); err != nil {
					__antithesis_instrumentation__.Notify(597741)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597742)
				}
				__antithesis_instrumentation__.Notify(597739)
				path, _ := darrayToStringSlice(ary)
				s, _, err := tree.MustBeDJSON(args[0]).JSON.RemovePath(path)
				if err != nil {
					__antithesis_instrumentation__.Notify(597743)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597744)
				}
				__antithesis_instrumentation__.Notify(597740)
				return &tree.DJSON{JSON: s}, nil
			},
			Info:       "Remove the specified path from the JSON object.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"json_extract_path": makeBuiltin(jsonProps(), jsonExtractPathImpl),

	"jsonb_extract_path": makeBuiltin(jsonProps(), jsonExtractPathImpl),

	"json_extract_path_text": makeBuiltin(jsonProps(), jsonExtractPathTextImpl),

	"jsonb_extract_path_text": makeBuiltin(jsonProps(), jsonExtractPathTextImpl),

	"json_set": makeBuiltin(jsonProps(), jsonSetImpl, jsonSetWithCreateMissingImpl),

	"jsonb_set": makeBuiltin(jsonProps(), jsonSetImpl, jsonSetWithCreateMissingImpl),

	"jsonb_insert": makeBuiltin(jsonProps(), jsonInsertImpl, jsonInsertWithInsertAfterImpl),

	"jsonb_pretty": makeBuiltin(jsonProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Jsonb}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597745)
				s, err := json.Pretty(tree.MustBeDJSON(args[0]).JSON)
				if err != nil {
					__antithesis_instrumentation__.Notify(597747)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597748)
				}
				__antithesis_instrumentation__.Notify(597746)
				return tree.NewDString(s), nil
			},
			Info:       "Returns the given JSON value as a STRING indented and with newlines.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"json_typeof": makeBuiltin(jsonProps(), jsonTypeOfImpl),

	"jsonb_typeof": makeBuiltin(jsonProps(), jsonTypeOfImpl),

	"array_to_json": arrayToJSONImpls,

	"to_json": makeBuiltin(jsonProps(), toJSONImpl),

	"to_jsonb": makeBuiltin(jsonProps(), toJSONImpl),

	"json_build_array": makeBuiltin(jsonPropsNullableArgs(), jsonBuildArrayImpl),

	"jsonb_build_array": makeBuiltin(jsonPropsNullableArgs(), jsonBuildArrayImpl),

	"json_build_object": makeBuiltin(jsonPropsNullableArgs(), jsonBuildObjectImpl),

	"jsonb_build_object": makeBuiltin(jsonPropsNullableArgs(), jsonBuildObjectImpl),

	"json_object": jsonObjectImpls,

	"jsonb_object": jsonObjectImpls,

	"json_strip_nulls": makeBuiltin(jsonProps(), jsonStripNullsImpl),

	"jsonb_strip_nulls": makeBuiltin(jsonProps(), jsonStripNullsImpl),

	"json_array_length": makeBuiltin(jsonProps(), jsonArrayLengthImpl),

	"jsonb_array_length": makeBuiltin(jsonProps(), jsonArrayLengthImpl),

	"jsonb_exists_any": makeBuiltin(
		jsonProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"json", types.Jsonb},
				{"array", types.StringArray},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(e *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597749)
				return tree.JSONExistsAny(e, tree.MustBeDJSON(args[0]), tree.MustBeDArray(args[1]))
			},
			Info:       "Returns whether any of the strings in the text array exist as top-level keys or array elements",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"json_valid": makeBuiltin(
		jsonProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"string", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(e *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597750)
				return jsonValidate(e, tree.MustBeDString(args[0])), nil
			},
			Info:       "Returns whether the given string is a valid JSON or not",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"crdb_internal.pb_to_json": makeBuiltin(
		jsonProps(),
		func() []tree.Overload {
			__antithesis_instrumentation__.Notify(597751)
			returnType := tree.FixedReturnType(types.Jsonb)
			const info = "Converts protocol message to its JSONB representation."
			volatility := tree.VolatilityImmutable
			pbToJSON := func(typ string, data []byte, flags protoreflect.FmtFlags) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597753)
				msg, err := protoreflect.DecodeMessage(typ, data)
				if err != nil {
					__antithesis_instrumentation__.Notify(597756)
					return nil, pgerror.Wrap(err, pgcode.InvalidParameterValue, "invalid protocol message")
				} else {
					__antithesis_instrumentation__.Notify(597757)
				}
				__antithesis_instrumentation__.Notify(597754)
				j, err := protoreflect.MessageToJSON(msg, flags)
				if err != nil {
					__antithesis_instrumentation__.Notify(597758)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597759)
				}
				__antithesis_instrumentation__.Notify(597755)
				return tree.NewDJSON(j), nil
			}
			__antithesis_instrumentation__.Notify(597752)

			return []tree.Overload{
				{
					Info:       info,
					Volatility: volatility,
					Types: tree.ArgTypes{
						{"pbname", types.String},
						{"data", types.Bytes},
					},
					ReturnType: returnType,
					Fn: func(context *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
						__antithesis_instrumentation__.Notify(597760)
						return pbToJSON(
							string(tree.MustBeDString(args[0])),
							[]byte(tree.MustBeDBytes(args[1])),
							protoreflect.FmtFlags{EmitDefaults: false, EmitRedacted: false},
						)
					},
				},
				{
					Info:       info,
					Volatility: volatility,
					Types: tree.ArgTypes{
						{"pbname", types.String},
						{"data", types.Bytes},
						{"emit_defaults", types.Bool},
					},
					ReturnType: returnType,
					Fn: func(context *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
						__antithesis_instrumentation__.Notify(597761)
						return pbToJSON(
							string(tree.MustBeDString(args[0])),
							[]byte(tree.MustBeDBytes(args[1])),
							protoreflect.FmtFlags{
								EmitDefaults: bool(tree.MustBeDBool(args[2])),
								EmitRedacted: false,
							},
						)
					},
				},
				{
					Info:       info,
					Volatility: volatility,
					Types: tree.ArgTypes{
						{"pbname", types.String},
						{"data", types.Bytes},
						{"emit_defaults", types.Bool},
						{"emit_redacted", types.Bool},
					},
					ReturnType: returnType,
					Fn: func(context *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
						__antithesis_instrumentation__.Notify(597762)
						return pbToJSON(
							string(tree.MustBeDString(args[0])),
							[]byte(tree.MustBeDBytes(args[1])),
							protoreflect.FmtFlags{
								EmitDefaults: bool(tree.MustBeDBool(args[2])),
								EmitRedacted: bool(tree.MustBeDBool(args[3])),
							},
						)
					},
				},
			}
		}()...),

	"crdb_internal.json_to_pb": makeBuiltin(
		jsonProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"pbname", types.String},
				{"json", types.Jsonb},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597763)
				msg, err := protoreflect.NewMessage(string(tree.MustBeDString(args[0])))
				if err != nil {
					__antithesis_instrumentation__.Notify(597766)
					return nil, pgerror.Wrap(err, pgcode.InvalidParameterValue, "invalid proto name")
				} else {
					__antithesis_instrumentation__.Notify(597767)
				}
				__antithesis_instrumentation__.Notify(597764)
				data, err := protoreflect.JSONBMarshalToMessage(tree.MustBeDJSON(args[1]).JSON, msg)
				if err != nil {
					__antithesis_instrumentation__.Notify(597768)
					return nil, pgerror.Wrap(err, pgcode.InvalidParameterValue, "invalid proto JSON")
				} else {
					__antithesis_instrumentation__.Notify(597769)
				}
				__antithesis_instrumentation__.Notify(597765)
				return tree.NewDBytes(tree.DBytes(data)), nil
			},
			Info:       "Convert JSONB data to protocol message bytes",
			Volatility: tree.VolatilityImmutable,
		}),

	"crdb_internal.read_file": makeBuiltin(
		jsonProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"uri", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597770)
				uri := string(tree.MustBeDString(args[0]))
				content, err := evalCtx.Planner.ExternalReadFile(evalCtx.Ctx(), uri)
				return tree.NewDBytes(tree.DBytes(content)), err
			},
			Info:       "Read the content of the file at the supplied external storage URI",
			Volatility: tree.VolatilityVolatile,
		}),

	"crdb_internal.write_file": makeBuiltin(
		jsonProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"data", types.Bytes},
				{"uri", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597771)
				data := tree.MustBeDBytes(args[0])
				uri := string(tree.MustBeDString(args[1]))
				if err := evalCtx.Planner.ExternalWriteFile(evalCtx.Ctx(), uri, []byte(data)); err != nil {
					__antithesis_instrumentation__.Notify(597773)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597774)
				}
				__antithesis_instrumentation__.Notify(597772)
				return tree.NewDInt(tree.DInt(len(data))), nil
			},
			Info:       "Write the content passed to a file at the supplied external storage URI",
			Volatility: tree.VolatilityVolatile,
		}),

	"crdb_internal.datums_to_bytes": makeBuiltin(
		tree.FunctionProperties{
			Category:             categorySystemInfo,
			NullableArgs:         true,
			Undocumented:         true,
			CompositeInsensitive: true,
		},
		tree.Overload{

			Info: "Converts datums into key-encoded bytes. " +
				"Supports NULLs and all data types which may be used in index keys",
			Types:      tree.VariadicType{VarType: types.Any},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597775)
				var out []byte
				for i, arg := range args {
					__antithesis_instrumentation__.Notify(597777)
					var err error
					out, err = keyside.Encode(out, arg, encoding.Ascending)
					if err != nil {
						__antithesis_instrumentation__.Notify(597778)
						return nil, pgerror.Newf(
							pgcode.DatatypeMismatch,
							"illegal argument %d of type %s",
							i, arg.ResolvedType(),
						)
					} else {
						__antithesis_instrumentation__.Notify(597779)
					}
				}
				__antithesis_instrumentation__.Notify(597776)
				return tree.NewDBytes(tree.DBytes(out)), nil
			},
			Volatility: tree.VolatilityImmutable,
		},
	),
	"crdb_internal.merge_statement_stats": makeBuiltin(arrayProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.JSONArray}},
			ReturnType: tree.FixedReturnType(types.Jsonb),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597780)
				arr := tree.MustBeDArray(args[0])
				var aggregatedStats roachpb.StatementStatistics
				for _, statsDatum := range arr.Array {
					__antithesis_instrumentation__.Notify(597783)
					if statsDatum == tree.DNull {
						__antithesis_instrumentation__.Notify(597786)
						continue
					} else {
						__antithesis_instrumentation__.Notify(597787)
					}
					__antithesis_instrumentation__.Notify(597784)
					var stats roachpb.StatementStatistics
					statsJSON := tree.MustBeDJSON(statsDatum).JSON
					if err := sqlstatsutil.DecodeStmtStatsStatisticsJSON(statsJSON, &stats); err != nil {
						__antithesis_instrumentation__.Notify(597788)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597789)
					}
					__antithesis_instrumentation__.Notify(597785)

					aggregatedStats.Add(&stats)
				}
				__antithesis_instrumentation__.Notify(597781)

				aggregatedJSON, err := sqlstatsutil.BuildStmtStatisticsJSON(&aggregatedStats)
				if err != nil {
					__antithesis_instrumentation__.Notify(597790)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597791)
				}
				__antithesis_instrumentation__.Notify(597782)

				return tree.NewDJSON(aggregatedJSON), nil
			},
			Info:       "Merge an array of roachpb.StatementStatistics into a single JSONB object",
			Volatility: tree.VolatilityImmutable,
		},
	),
	"crdb_internal.merge_transaction_stats": makeBuiltin(arrayProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.JSONArray}},
			ReturnType: tree.FixedReturnType(types.Jsonb),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597792)
				arr := tree.MustBeDArray(args[0])
				var aggregatedStats roachpb.TransactionStatistics
				for _, statsDatum := range arr.Array {
					__antithesis_instrumentation__.Notify(597795)
					if statsDatum == tree.DNull {
						__antithesis_instrumentation__.Notify(597798)
						continue
					} else {
						__antithesis_instrumentation__.Notify(597799)
					}
					__antithesis_instrumentation__.Notify(597796)
					var stats roachpb.TransactionStatistics
					statsJSON := tree.MustBeDJSON(statsDatum).JSON
					if err := sqlstatsutil.DecodeTxnStatsStatisticsJSON(statsJSON, &stats); err != nil {
						__antithesis_instrumentation__.Notify(597800)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597801)
					}
					__antithesis_instrumentation__.Notify(597797)

					aggregatedStats.Add(&stats)
				}
				__antithesis_instrumentation__.Notify(597793)

				aggregatedJSON, err := sqlstatsutil.BuildTxnStatisticsJSON(
					&roachpb.CollectedTransactionStatistics{
						Stats: aggregatedStats,
					})
				if err != nil {
					__antithesis_instrumentation__.Notify(597802)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597803)
				}
				__antithesis_instrumentation__.Notify(597794)

				return tree.NewDJSON(aggregatedJSON), nil
			},
			Info:       "Merge an array of roachpb.TransactionStatistics into a single JSONB object",
			Volatility: tree.VolatilityImmutable,
		},
	),
	"crdb_internal.merge_stats_metadata": makeBuiltin(arrayProps(),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.JSONArray}},
			ReturnType: tree.FixedReturnType(types.Jsonb),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597804)
				arr := tree.MustBeDArray(args[0])
				metadata := &roachpb.AggregatedStatementMetadata{}

				for _, metadataDatum := range arr.Array {
					__antithesis_instrumentation__.Notify(597807)
					if metadataDatum == tree.DNull {
						__antithesis_instrumentation__.Notify(597814)
						continue
					} else {
						__antithesis_instrumentation__.Notify(597815)
					}
					__antithesis_instrumentation__.Notify(597808)

					var statistics roachpb.CollectedStatementStatistics
					metadataJSON := tree.MustBeDJSON(metadataDatum).JSON
					err := sqlstatsutil.DecodeStmtStatsMetadataJSON(metadataJSON, &statistics)
					if err != nil {
						__antithesis_instrumentation__.Notify(597816)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597817)
					}
					__antithesis_instrumentation__.Notify(597809)
					metadata.ImplicitTxn = statistics.Key.ImplicitTxn
					metadata.Query = statistics.Key.Query
					metadata.QuerySummary = statistics.Key.QuerySummary
					metadata.StmtType = statistics.Stats.SQLType
					metadata.Databases = util.CombineUniqueString(metadata.Databases, []string{statistics.Key.Database})

					if statistics.Key.DistSQL {
						__antithesis_instrumentation__.Notify(597818)
						metadata.DistSQLCount++
					} else {
						__antithesis_instrumentation__.Notify(597819)
					}
					__antithesis_instrumentation__.Notify(597810)
					if statistics.Key.Failed {
						__antithesis_instrumentation__.Notify(597820)
						metadata.FailedCount++
					} else {
						__antithesis_instrumentation__.Notify(597821)
					}
					__antithesis_instrumentation__.Notify(597811)
					if statistics.Key.FullScan {
						__antithesis_instrumentation__.Notify(597822)
						metadata.FullScanCount++
					} else {
						__antithesis_instrumentation__.Notify(597823)
					}
					__antithesis_instrumentation__.Notify(597812)
					if statistics.Key.Vec {
						__antithesis_instrumentation__.Notify(597824)
						metadata.VecCount++
					} else {
						__antithesis_instrumentation__.Notify(597825)
					}
					__antithesis_instrumentation__.Notify(597813)
					metadata.TotalCount++
				}
				__antithesis_instrumentation__.Notify(597805)
				aggregatedJSON, err := sqlstatsutil.BuildStmtDetailsMetadataJSON(metadata)
				if err != nil {
					__antithesis_instrumentation__.Notify(597826)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597827)
				}
				__antithesis_instrumentation__.Notify(597806)

				return tree.NewDJSON(aggregatedJSON), nil
			},
			Info:       "Merge an array of StmtStatsMetadata into a single JSONB object",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"enum_first": makeBuiltin(
		tree.FunctionProperties{NullableArgs: true, Category: categoryEnum},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.AnyEnum}},
			ReturnType: tree.IdentityReturnType(0),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597828)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597831)
					return nil, pgerror.Newf(pgcode.NullValueNotAllowed, "argument cannot be NULL")
				} else {
					__antithesis_instrumentation__.Notify(597832)
				}
				__antithesis_instrumentation__.Notify(597829)
				arg := args[0].(*tree.DEnum)
				min, ok := arg.MinWriteable()
				if !ok {
					__antithesis_instrumentation__.Notify(597833)
					return nil, errors.Newf("enum %s contains no values", arg.ResolvedType().Name())
				} else {
					__antithesis_instrumentation__.Notify(597834)
				}
				__antithesis_instrumentation__.Notify(597830)
				return min, nil
			},
			Info:       "Returns the first value of the input enum type.",
			Volatility: tree.VolatilityStable,
		},
	),

	"enum_last": makeBuiltin(
		tree.FunctionProperties{NullableArgs: true, Category: categoryEnum},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.AnyEnum}},
			ReturnType: tree.IdentityReturnType(0),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597835)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597838)
					return nil, pgerror.Newf(pgcode.NullValueNotAllowed, "argument cannot be NULL")
				} else {
					__antithesis_instrumentation__.Notify(597839)
				}
				__antithesis_instrumentation__.Notify(597836)
				arg := args[0].(*tree.DEnum)
				max, ok := arg.MaxWriteable()
				if !ok {
					__antithesis_instrumentation__.Notify(597840)
					return nil, errors.Newf("enum %s contains no values", arg.ResolvedType().Name())
				} else {
					__antithesis_instrumentation__.Notify(597841)
				}
				__antithesis_instrumentation__.Notify(597837)
				return max, nil
			},
			Info:       "Returns the last value of the input enum type.",
			Volatility: tree.VolatilityStable,
		},
	),

	"enum_range": makeBuiltin(
		tree.FunctionProperties{NullableArgs: true, Category: categoryEnum},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.AnyEnum}},
			ReturnType: tree.ArrayOfFirstNonNullReturnType(),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597842)
				if args[0] == tree.DNull {
					__antithesis_instrumentation__.Notify(597845)
					return nil, pgerror.Newf(pgcode.NullValueNotAllowed, "argument cannot be NULL")
				} else {
					__antithesis_instrumentation__.Notify(597846)
				}
				__antithesis_instrumentation__.Notify(597843)
				arg := args[0].(*tree.DEnum)
				typ := arg.EnumTyp
				arr := tree.NewDArray(typ)
				for i := range typ.TypeMeta.EnumData.LogicalRepresentations {
					__antithesis_instrumentation__.Notify(597847)

					if typ.TypeMeta.EnumData.IsMemberReadOnly[i] {
						__antithesis_instrumentation__.Notify(597849)
						continue
					} else {
						__antithesis_instrumentation__.Notify(597850)
					}
					__antithesis_instrumentation__.Notify(597848)
					enum := &tree.DEnum{
						EnumTyp:     typ,
						PhysicalRep: typ.TypeMeta.EnumData.PhysicalRepresentations[i],
						LogicalRep:  typ.TypeMeta.EnumData.LogicalRepresentations[i],
					}
					if err := arr.Append(enum); err != nil {
						__antithesis_instrumentation__.Notify(597851)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597852)
					}
				}
				__antithesis_instrumentation__.Notify(597844)
				return arr, nil
			},
			Info:       "Returns all values of the input enum in an ordered array.",
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"lower", types.AnyEnum}, {"upper", types.AnyEnum}},
			ReturnType: tree.ArrayOfFirstNonNullReturnType(),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597853)
				if args[0] == tree.DNull && func() bool {
					__antithesis_instrumentation__.Notify(597857)
					return args[1] == tree.DNull == true
				}() == true {
					__antithesis_instrumentation__.Notify(597858)
					return nil, pgerror.Newf(pgcode.NullValueNotAllowed, "both arguments cannot be NULL")
				} else {
					__antithesis_instrumentation__.Notify(597859)
				}
				__antithesis_instrumentation__.Notify(597854)
				var bottom, top int
				var typ *types.T
				switch {
				case args[0] == tree.DNull:
					__antithesis_instrumentation__.Notify(597860)
					right := args[1].(*tree.DEnum)
					typ = right.ResolvedType()
					idx, err := typ.EnumGetIdxOfPhysical(right.PhysicalRep)
					if err != nil {
						__antithesis_instrumentation__.Notify(597867)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597868)
					}
					__antithesis_instrumentation__.Notify(597861)
					bottom, top = 0, idx
				case args[1] == tree.DNull:
					__antithesis_instrumentation__.Notify(597862)
					left := args[0].(*tree.DEnum)
					typ = left.ResolvedType()
					idx, err := typ.EnumGetIdxOfPhysical(left.PhysicalRep)
					if err != nil {
						__antithesis_instrumentation__.Notify(597869)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597870)
					}
					__antithesis_instrumentation__.Notify(597863)
					bottom, top = idx, len(typ.TypeMeta.EnumData.PhysicalRepresentations)-1
				default:
					__antithesis_instrumentation__.Notify(597864)
					left, right := args[0].(*tree.DEnum), args[1].(*tree.DEnum)
					if !left.ResolvedType().Equivalent(right.ResolvedType()) {
						__antithesis_instrumentation__.Notify(597871)
						return nil, pgerror.Newf(
							pgcode.DatatypeMismatch,
							"mismatched types %s and %s",
							left.ResolvedType(),
							right.ResolvedType(),
						)
					} else {
						__antithesis_instrumentation__.Notify(597872)
					}
					__antithesis_instrumentation__.Notify(597865)
					typ = left.ResolvedType()
					var err error
					bottom, err = typ.EnumGetIdxOfPhysical(left.PhysicalRep)
					if err != nil {
						__antithesis_instrumentation__.Notify(597873)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597874)
					}
					__antithesis_instrumentation__.Notify(597866)
					top, err = typ.EnumGetIdxOfPhysical(right.PhysicalRep)
					if err != nil {
						__antithesis_instrumentation__.Notify(597875)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597876)
					}
				}
				__antithesis_instrumentation__.Notify(597855)
				arr := tree.NewDArray(typ)
				for i := bottom; i <= top; i++ {
					__antithesis_instrumentation__.Notify(597877)

					if typ.TypeMeta.EnumData.IsMemberReadOnly[i] {
						__antithesis_instrumentation__.Notify(597879)
						continue
					} else {
						__antithesis_instrumentation__.Notify(597880)
					}
					__antithesis_instrumentation__.Notify(597878)
					enum := &tree.DEnum{
						EnumTyp:     typ,
						PhysicalRep: typ.TypeMeta.EnumData.PhysicalRepresentations[i],
						LogicalRep:  typ.TypeMeta.EnumData.LogicalRepresentations[i],
					}
					if err := arr.Append(enum); err != nil {
						__antithesis_instrumentation__.Notify(597881)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(597882)
					}
				}
				__antithesis_instrumentation__.Notify(597856)
				return arr, nil
			},
			Info:       "Returns all values of the input enum in an ordered array between the two arguments (inclusive).",
			Volatility: tree.VolatilityStable,
		},
	),

	"version": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597883)
				return tree.NewDString(build.GetInfo().Short()), nil
			},
			Info:       "Returns the node's version of CockroachDB.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"current_database": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597884)
				if len(ctx.SessionData().Database) == 0 {
					__antithesis_instrumentation__.Notify(597886)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597887)
				}
				__antithesis_instrumentation__.Notify(597885)
				return tree.NewDString(ctx.SessionData().Database), nil
			},
			Info:       "Returns the current database.",
			Volatility: tree.VolatilityStable,
		},
	),

	"current_schema": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemInfo,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597888)
				ctx := evalCtx.Ctx()
				curDb := evalCtx.SessionData().Database
				iter := evalCtx.SessionData().SearchPath.IterWithoutImplicitPGSchemas()
				for scName, ok := iter.Next(); ok; scName, ok = iter.Next() {
					__antithesis_instrumentation__.Notify(597890)
					if found, err := evalCtx.Planner.SchemaExists(ctx, curDb, scName); found || func() bool {
						__antithesis_instrumentation__.Notify(597891)
						return err != nil == true
					}() == true {
						__antithesis_instrumentation__.Notify(597892)
						if err != nil {
							__antithesis_instrumentation__.Notify(597894)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(597895)
						}
						__antithesis_instrumentation__.Notify(597893)
						return tree.NewDString(scName), nil
					} else {
						__antithesis_instrumentation__.Notify(597896)
					}
				}
				__antithesis_instrumentation__.Notify(597889)
				return tree.DNull, nil
			},
			Info:       "Returns the current schema.",
			Volatility: tree.VolatilityStable,
		},
	),

	"current_schemas": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemInfo,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"include_pg_catalog", types.Bool}},
			ReturnType: tree.FixedReturnType(types.StringArray),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597897)
				ctx := evalCtx.Ctx()
				curDb := evalCtx.SessionData().Database
				includeImplicitPgSchemas := *(args[0].(*tree.DBool))
				schemas := tree.NewDArray(types.String)
				var iter sessiondata.SearchPathIter
				if includeImplicitPgSchemas {
					__antithesis_instrumentation__.Notify(597900)
					iter = evalCtx.SessionData().SearchPath.Iter()
				} else {
					__antithesis_instrumentation__.Notify(597901)
					iter = evalCtx.SessionData().SearchPath.IterWithoutImplicitPGSchemas()
				}
				__antithesis_instrumentation__.Notify(597898)
				for scName, ok := iter.Next(); ok; scName, ok = iter.Next() {
					__antithesis_instrumentation__.Notify(597902)
					if found, err := evalCtx.Planner.SchemaExists(ctx, curDb, scName); found || func() bool {
						__antithesis_instrumentation__.Notify(597903)
						return err != nil == true
					}() == true {
						__antithesis_instrumentation__.Notify(597904)
						if err != nil {
							__antithesis_instrumentation__.Notify(597906)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(597907)
						}
						__antithesis_instrumentation__.Notify(597905)
						if err := schemas.Append(tree.NewDString(scName)); err != nil {
							__antithesis_instrumentation__.Notify(597908)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(597909)
						}
					} else {
						__antithesis_instrumentation__.Notify(597910)
					}
				}
				__antithesis_instrumentation__.Notify(597899)
				return schemas, nil
			},
			Info:       "Returns the valid schemas in the search path.",
			Volatility: tree.VolatilityStable,
		},
	),

	"current_user": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597911)
				if ctx.SessionData().User().Undefined() {
					__antithesis_instrumentation__.Notify(597913)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597914)
				}
				__antithesis_instrumentation__.Notify(597912)
				return tree.NewDString(ctx.SessionData().User().Normalized()), nil
			},
			Info: "Returns the current user. This function is provided for " +
				"compatibility with PostgreSQL.",
			Volatility: tree.VolatilityStable,
		},
	),

	"session_user": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597915)
				u := ctx.SessionData().SessionUser()
				if u.Undefined() {
					__antithesis_instrumentation__.Notify(597917)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597918)
				}
				__antithesis_instrumentation__.Notify(597916)
				return tree.NewDString(u.Normalized()), nil
			},
			Info: "Returns the session user. This function is provided for " +
				"compatibility with PostgreSQL.",
			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.trace_id": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597919)

				isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(597924)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597925)
				}
				__antithesis_instrumentation__.Notify(597920)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(597926)
					return nil, errInsufficientPriv
				} else {
					__antithesis_instrumentation__.Notify(597927)
				}
				__antithesis_instrumentation__.Notify(597921)

				sp := tracing.SpanFromContext(ctx.Context)
				if sp == nil {
					__antithesis_instrumentation__.Notify(597928)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597929)
				}
				__antithesis_instrumentation__.Notify(597922)

				traceID := sp.TraceID()
				if traceID == 0 {
					__antithesis_instrumentation__.Notify(597930)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(597931)
				}
				__antithesis_instrumentation__.Notify(597923)
				return tree.NewDInt(tree.DInt(traceID)), nil
			},
			Info: "Returns the current trace ID or an error if no trace is open.",

			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.set_trace_verbose": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types: tree.ArgTypes{
				{"trace_id", types.Int},
				{"verbosity", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597932)

				isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(597939)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597940)
				}
				__antithesis_instrumentation__.Notify(597933)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(597941)
					return nil, errInsufficientPriv
				} else {
					__antithesis_instrumentation__.Notify(597942)
				}
				__antithesis_instrumentation__.Notify(597934)

				traceID := tracingpb.TraceID(*(args[0].(*tree.DInt)))
				verbosity := bool(*(args[1].(*tree.DBool)))

				var rootSpan tracing.RegistrySpan
				if ctx.Tracer == nil {
					__antithesis_instrumentation__.Notify(597943)
					return nil, errors.AssertionFailedf("Tracer not configured")
				} else {
					__antithesis_instrumentation__.Notify(597944)
				}
				__antithesis_instrumentation__.Notify(597935)
				if err := ctx.Tracer.VisitSpans(func(span tracing.RegistrySpan) error {
					__antithesis_instrumentation__.Notify(597945)
					if span.TraceID() == traceID && func() bool {
						__antithesis_instrumentation__.Notify(597947)
						return rootSpan == nil == true
					}() == true {
						__antithesis_instrumentation__.Notify(597948)
						rootSpan = span
					} else {
						__antithesis_instrumentation__.Notify(597949)
					}
					__antithesis_instrumentation__.Notify(597946)

					return nil
				}); err != nil {
					__antithesis_instrumentation__.Notify(597950)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597951)
				}
				__antithesis_instrumentation__.Notify(597936)
				if rootSpan == nil {
					__antithesis_instrumentation__.Notify(597952)
					return tree.DBoolFalse, nil
				} else {
					__antithesis_instrumentation__.Notify(597953)
				}
				__antithesis_instrumentation__.Notify(597937)

				var recType tracing.RecordingType
				if verbosity {
					__antithesis_instrumentation__.Notify(597954)
					recType = tracing.RecordingVerbose
				} else {
					__antithesis_instrumentation__.Notify(597955)
					recType = tracing.RecordingOff
				}
				__antithesis_instrumentation__.Notify(597938)
				rootSpan.SetRecordingType(recType)
				return tree.DBoolTrue, nil
			},
			Info:       "Returns true if root span was found and verbosity was set, false otherwise.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.locality_value": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{{"key", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597956)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(597959)
					return nil, errors.Newf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(597960)
				}
				__antithesis_instrumentation__.Notify(597957)
				key := string(s)
				for i := range ctx.Locality.Tiers {
					__antithesis_instrumentation__.Notify(597961)
					tier := &ctx.Locality.Tiers[i]
					if tier.Key == key {
						__antithesis_instrumentation__.Notify(597962)
						return tree.NewDString(tier.Value), nil
					} else {
						__antithesis_instrumentation__.Notify(597963)
					}
				}
				__antithesis_instrumentation__.Notify(597958)
				return tree.DNull, nil
			},
			Info:       "Returns the value of the specified locality key.",
			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.cluster_setting_encoded_default": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{{"setting", types.String}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597964)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(597968)
					return nil, errors.AssertionFailedf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(597969)
				}
				__antithesis_instrumentation__.Notify(597965)
				name := strings.ToLower(string(s))
				rawSetting, ok := settings.Lookup(
					name, settings.LookupForLocalAccess, ctx.Codec.ForSystemTenant(),
				)
				if !ok {
					__antithesis_instrumentation__.Notify(597970)
					return nil, errors.Newf("unknown cluster setting '%s'", name)
				} else {
					__antithesis_instrumentation__.Notify(597971)
				}
				__antithesis_instrumentation__.Notify(597966)
				setting, ok := rawSetting.(settings.NonMaskedSetting)
				if !ok {
					__antithesis_instrumentation__.Notify(597972)

					return nil, errors.AssertionFailedf("setting '%s' is masked", name)
				} else {
					__antithesis_instrumentation__.Notify(597973)
				}
				__antithesis_instrumentation__.Notify(597967)

				return tree.NewDString(setting.EncodedDefault()), nil
			},
			Info:       "Returns the encoded default value of the given cluster setting.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"crdb_internal.decode_cluster_setting": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types: tree.ArgTypes{
				{"setting", types.String},
				{"value", types.String},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597974)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(597980)
					return nil, errors.AssertionFailedf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(597981)
				}
				__antithesis_instrumentation__.Notify(597975)
				encoded, ok := tree.AsDString(args[1])
				if !ok {
					__antithesis_instrumentation__.Notify(597982)
					return nil, errors.AssertionFailedf("expected string value, got %T", args[1])
				} else {
					__antithesis_instrumentation__.Notify(597983)
				}
				__antithesis_instrumentation__.Notify(597976)
				name := strings.ToLower(string(s))
				rawSetting, ok := settings.Lookup(
					name, settings.LookupForLocalAccess, ctx.Codec.ForSystemTenant(),
				)
				if !ok {
					__antithesis_instrumentation__.Notify(597984)
					return nil, errors.Newf("unknown cluster setting '%s'", name)
				} else {
					__antithesis_instrumentation__.Notify(597985)
				}
				__antithesis_instrumentation__.Notify(597977)
				setting, ok := rawSetting.(settings.NonMaskedSetting)
				if !ok {
					__antithesis_instrumentation__.Notify(597986)

					return nil, errors.AssertionFailedf("setting '%s' is masked", name)
				} else {
					__antithesis_instrumentation__.Notify(597987)
				}
				__antithesis_instrumentation__.Notify(597978)
				repr, err := setting.DecodeToString(string(encoded))
				if err != nil {
					__antithesis_instrumentation__.Notify(597988)
					return nil, errors.Wrapf(err, "%v", name)
				} else {
					__antithesis_instrumentation__.Notify(597989)
				}
				__antithesis_instrumentation__.Notify(597979)
				return tree.NewDString(repr), nil
			},
			Info:       "Decodes the given encoded value for a cluster setting.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"crdb_internal.node_executable_version": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597990)
				v := ctx.Settings.Version.BinaryVersion().String()
				return tree.NewDString(v), nil
			},
			Info:       "Returns the version of CockroachDB this node is running.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.active_version": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Jsonb),
			Fn: func(ctx *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597991)
				activeVersion := ctx.Settings.Version.ActiveVersionOrEmpty(ctx.Context)
				jsonStr, err := gojson.Marshal(&activeVersion.Version)
				if err != nil {
					__antithesis_instrumentation__.Notify(597994)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597995)
				}
				__antithesis_instrumentation__.Notify(597992)
				jsonDatum, err := tree.ParseDJSON(string(jsonStr))
				if err != nil {
					__antithesis_instrumentation__.Notify(597996)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(597997)
				}
				__antithesis_instrumentation__.Notify(597993)
				return jsonDatum, nil
			},
			Info:       "Returns the current active cluster version.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.is_at_least_version": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{{"version", types.String}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(597998)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(598003)
					return nil, errors.Newf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(598004)
				}
				__antithesis_instrumentation__.Notify(597999)
				arg, err := roachpb.ParseVersion(string(s))
				if err != nil {
					__antithesis_instrumentation__.Notify(598005)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598006)
				}
				__antithesis_instrumentation__.Notify(598000)
				activeVersion := ctx.Settings.Version.ActiveVersionOrEmpty(ctx.Context)
				if activeVersion == (clusterversion.ClusterVersion{}) {
					__antithesis_instrumentation__.Notify(598007)
					return nil, errors.AssertionFailedf("invalid uninitialized version")
				} else {
					__antithesis_instrumentation__.Notify(598008)
				}
				__antithesis_instrumentation__.Notify(598001)
				if arg.LessEq(activeVersion.Version) {
					__antithesis_instrumentation__.Notify(598009)
					return tree.DBoolTrue, nil
				} else {
					__antithesis_instrumentation__.Notify(598010)
				}
				__antithesis_instrumentation__.Notify(598002)
				return tree.DBoolFalse, nil
			},
			Info:       "Returns true if the cluster version is not older than the argument.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.approximate_timestamp": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{{"timestamp", types.Decimal}},
			ReturnType: tree.FixedReturnType(types.Timestamp),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598011)
				return tree.DecimalToInexactDTimestamp(args[0].(*tree.DDecimal))
			},
			Info:       "Converts the crdb_internal_mvcc_timestamp column into an approximate timestamp.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"crdb_internal.cluster_id": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Uuid),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598012)
				return tree.NewDUuid(tree.DUuid{UUID: ctx.ClusterID}), nil
			},
			Info:       "Returns the logical cluster ID for this tenant.",
			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.node_id": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598013)
				dNodeID := tree.DNull
				if nodeID, ok := ctx.NodeID.OptionalNodeID(); ok {
					__antithesis_instrumentation__.Notify(598015)
					dNodeID = tree.NewDInt(tree.DInt(nodeID))
				} else {
					__antithesis_instrumentation__.Notify(598016)
				}
				__antithesis_instrumentation__.Notify(598014)
				return dNodeID, nil
			},
			Info:       "Returns the node ID.",
			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.cluster_name": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598017)
				return tree.NewDString(ctx.ClusterName), nil
			},
			Info:       "Returns the cluster name.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.create_tenant": makeBuiltin(
		tree.FunctionProperties{
			Category:     categoryMultiTenancy,
			NullableArgs: true,
			Undocumented: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"id", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598018)
				if err := requireNonNull(args[0]); err != nil {
					__antithesis_instrumentation__.Notify(598022)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598023)
				}
				__antithesis_instrumentation__.Notify(598019)
				sTenID, err := mustBeDIntInTenantRange(args[0])
				if err != nil {
					__antithesis_instrumentation__.Notify(598024)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598025)
				}
				__antithesis_instrumentation__.Notify(598020)
				if err := ctx.Tenant.CreateTenant(ctx.Context, uint64(sTenID)); err != nil {
					__antithesis_instrumentation__.Notify(598026)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598027)
				}
				__antithesis_instrumentation__.Notify(598021)
				return args[0], nil
			},
			Info:       "Creates a new tenant with the provided ID. Must be run by the System tenant.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.create_join_token": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598028)
				token, err := ctx.JoinTokenCreator.CreateJoinToken(ctx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(598030)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598031)
				}
				__antithesis_instrumentation__.Notify(598029)
				return tree.NewDString(token), nil
			},
			Info:       "Creates a join token for use when adding a new node to a secure cluster.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.destroy_tenant": makeBuiltin(
		tree.FunctionProperties{
			Category:     categoryMultiTenancy,
			Undocumented: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"id", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598032)
				sTenID, err := mustBeDIntInTenantRange(args[0])
				if err != nil {
					__antithesis_instrumentation__.Notify(598035)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598036)
				}
				__antithesis_instrumentation__.Notify(598033)
				if err := ctx.Tenant.DestroyTenant(
					ctx.Context, uint64(sTenID), false,
				); err != nil {
					__antithesis_instrumentation__.Notify(598037)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598038)
				}
				__antithesis_instrumentation__.Notify(598034)
				return args[0], nil
			},
			Info:       "Destroys a tenant with the provided ID. Must be run by the System tenant.",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"id", types.Int},
				{"synchronous", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598039)
				sTenID, err := mustBeDIntInTenantRange(args[0])
				if err != nil {
					__antithesis_instrumentation__.Notify(598042)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598043)
				}
				__antithesis_instrumentation__.Notify(598040)
				synchronous := tree.MustBeDBool(args[1])
				if err := ctx.Tenant.DestroyTenant(
					ctx.Context, uint64(sTenID), bool(synchronous),
				); err != nil {
					__antithesis_instrumentation__.Notify(598044)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598045)
				}
				__antithesis_instrumentation__.Notify(598041)
				return args[0], nil
			},
			Info: "Destroys a tenant with the provided ID. Must be run by the System tenant. " +
				"Optionally, synchronously destroy the data",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.encode_key": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types: tree.ArgTypes{
				{"table_id", types.Int},
				{"index_id", types.Int},
				{"row_tuple", types.Any},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598046)
				tableID := int(tree.MustBeDInt(args[0]))
				indexID := int(tree.MustBeDInt(args[1]))
				rowDatums, ok := tree.AsDTuple(args[2])
				if !ok {
					__antithesis_instrumentation__.Notify(598057)
					return nil, pgerror.Newf(
						pgcode.DatatypeMismatch,
						"expected tuple argument for row_tuple, found %s",
						args[2],
					)
				} else {
					__antithesis_instrumentation__.Notify(598058)
				}
				__antithesis_instrumentation__.Notify(598047)

				tableDescI, err := ctx.Planner.GetImmutableTableInterfaceByID(ctx.Ctx(), tableID)
				if err != nil {
					__antithesis_instrumentation__.Notify(598059)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598060)
				}
				__antithesis_instrumentation__.Notify(598048)
				tableDesc := tableDescI.(catalog.TableDescriptor)
				index, err := tableDesc.FindIndexWithID(descpb.IndexID(indexID))
				if err != nil {
					__antithesis_instrumentation__.Notify(598061)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598062)
				}
				__antithesis_instrumentation__.Notify(598049)

				indexColIDs := make([]descpb.ColumnID, index.NumKeyColumns(), index.NumKeyColumns()+index.NumKeySuffixColumns())
				for i := 0; i < index.NumKeyColumns(); i++ {
					__antithesis_instrumentation__.Notify(598063)
					indexColIDs[i] = index.GetKeyColumnID(i)
				}
				__antithesis_instrumentation__.Notify(598050)
				if index.GetID() != tableDesc.GetPrimaryIndexID() && func() bool {
					__antithesis_instrumentation__.Notify(598064)
					return !index.IsUnique() == true
				}() == true {
					__antithesis_instrumentation__.Notify(598065)
					for i := 0; i < index.NumKeySuffixColumns(); i++ {
						__antithesis_instrumentation__.Notify(598066)
						indexColIDs = append(indexColIDs, index.GetKeySuffixColumnID(i))
					}
				} else {
					__antithesis_instrumentation__.Notify(598067)
				}
				__antithesis_instrumentation__.Notify(598051)

				if len(rowDatums.D) != len(indexColIDs) {
					__antithesis_instrumentation__.Notify(598068)
					err := errors.Newf(
						"number of values must equal number of columns in index %q",
						index.GetName(),
					)

					if index.GetID() != tableDesc.GetPrimaryIndexID() && func() bool {
						__antithesis_instrumentation__.Notify(598070)
						return !index.IsUnique() == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(598071)
						return index.NumKeySuffixColumns() > 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(598072)
						var extraColNames []string
						for i := 0; i < index.NumKeySuffixColumns(); i++ {
							__antithesis_instrumentation__.Notify(598075)
							id := index.GetKeySuffixColumnID(i)
							col, colErr := tableDesc.FindColumnWithID(id)
							if colErr != nil {
								__antithesis_instrumentation__.Notify(598077)
								return nil, errors.CombineErrors(err, colErr)
							} else {
								__antithesis_instrumentation__.Notify(598078)
							}
							__antithesis_instrumentation__.Notify(598076)
							extraColNames = append(extraColNames, col.GetName())
						}
						__antithesis_instrumentation__.Notify(598073)
						var allColNames []string
						for _, id := range indexColIDs {
							__antithesis_instrumentation__.Notify(598079)
							col, colErr := tableDesc.FindColumnWithID(id)
							if colErr != nil {
								__antithesis_instrumentation__.Notify(598081)
								return nil, errors.CombineErrors(err, colErr)
							} else {
								__antithesis_instrumentation__.Notify(598082)
							}
							__antithesis_instrumentation__.Notify(598080)
							allColNames = append(allColNames, col.GetName())
						}
						__antithesis_instrumentation__.Notify(598074)
						return nil, errors.WithHintf(
							err,
							"columns %v are implicitly part of index %q's key, include columns %v in this order",
							extraColNames,
							index.GetName(),
							allColNames,
						)
					} else {
						__antithesis_instrumentation__.Notify(598083)
					}
					__antithesis_instrumentation__.Notify(598069)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598084)
				}
				__antithesis_instrumentation__.Notify(598052)

				var datums tree.Datums
				for i, d := range rowDatums.D {
					__antithesis_instrumentation__.Notify(598085)

					var newDatum tree.Datum
					col, err := tableDesc.FindColumnWithID(indexColIDs[i])
					if err != nil {
						__antithesis_instrumentation__.Notify(598088)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(598089)
					}
					__antithesis_instrumentation__.Notify(598086)
					if d.ResolvedType() == types.Unknown {
						__antithesis_instrumentation__.Notify(598090)
						if !col.IsNullable() {
							__antithesis_instrumentation__.Notify(598092)
							return nil, pgerror.Newf(pgcode.NotNullViolation, "NULL provided as a value for a non-nullable column")
						} else {
							__antithesis_instrumentation__.Notify(598093)
						}
						__antithesis_instrumentation__.Notify(598091)
						newDatum = tree.DNull
					} else {
						__antithesis_instrumentation__.Notify(598094)
						expectedTyp := col.GetType()
						newDatum, err = tree.PerformCast(ctx, d, expectedTyp)
						if err != nil {
							__antithesis_instrumentation__.Notify(598095)
							return nil, errors.WithHint(err, "try to explicitly cast each value to the corresponding column type")
						} else {
							__antithesis_instrumentation__.Notify(598096)
						}
					}
					__antithesis_instrumentation__.Notify(598087)
					datums = append(datums, newDatum)
				}
				__antithesis_instrumentation__.Notify(598053)

				var colMap catalog.TableColMap
				for i, id := range indexColIDs {
					__antithesis_instrumentation__.Notify(598097)
					colMap.Set(id, i)
				}
				__antithesis_instrumentation__.Notify(598054)

				keyPrefix := rowenc.MakeIndexKeyPrefix(ctx.Codec, tableDesc.GetID(), index.GetID())
				keyAndSuffixCols := tableDesc.IndexFetchSpecKeyAndSuffixColumns(index)
				if len(datums) > len(keyAndSuffixCols) {
					__antithesis_instrumentation__.Notify(598098)
					return nil, errors.Errorf("encoding too many columns (%d)", len(datums))
				} else {
					__antithesis_instrumentation__.Notify(598099)
				}
				__antithesis_instrumentation__.Notify(598055)
				res, _, err := rowenc.EncodePartialIndexKey(keyAndSuffixCols[:len(datums)], colMap, datums, keyPrefix)
				if err != nil {
					__antithesis_instrumentation__.Notify(598100)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598101)
				}
				__antithesis_instrumentation__.Notify(598056)
				return tree.NewDBytes(tree.DBytes(res)), nil
			},
			Info:       "Generate the key for a row on a particular table and index.",
			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.force_error": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"errorCode", types.String}, {"msg", types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598102)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(598106)
					return nil, errors.Newf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(598107)
				}
				__antithesis_instrumentation__.Notify(598103)
				errCode := string(s)
				s, ok = tree.AsDString(args[1])
				if !ok {
					__antithesis_instrumentation__.Notify(598108)
					return nil, errors.Newf("expected string value, got %T", args[1])
				} else {
					__antithesis_instrumentation__.Notify(598109)
				}
				__antithesis_instrumentation__.Notify(598104)
				msg := string(s)

				if errCode == "" {
					__antithesis_instrumentation__.Notify(598110)
					return nil, errors.Newf("%s", msg)
				} else {
					__antithesis_instrumentation__.Notify(598111)
				}
				__antithesis_instrumentation__.Notify(598105)
				return nil, pgerror.Newf(pgcode.MakeCode(errCode), "%s", msg)
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.notice": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"msg", types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598112)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(598114)
					return nil, errors.Newf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(598115)
				}
				__antithesis_instrumentation__.Notify(598113)
				msg := string(s)
				return crdbInternalSendNotice(ctx, "NOTICE", msg)
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"severity", types.String}, {"msg", types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598116)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(598120)
					return nil, errors.Newf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(598121)
				}
				__antithesis_instrumentation__.Notify(598117)
				severityString := string(s)
				s, ok = tree.AsDString(args[1])
				if !ok {
					__antithesis_instrumentation__.Notify(598122)
					return nil, errors.Newf("expected string value, got %T", args[1])
				} else {
					__antithesis_instrumentation__.Notify(598123)
				}
				__antithesis_instrumentation__.Notify(598118)
				msg := string(s)
				if _, ok := pgnotice.ParseDisplaySeverity(severityString); !ok {
					__antithesis_instrumentation__.Notify(598124)
					return nil, pgerror.Newf(pgcode.InvalidParameterValue, "severity %s is invalid", severityString)
				} else {
					__antithesis_instrumentation__.Notify(598125)
				}
				__antithesis_instrumentation__.Notify(598119)
				return crdbInternalSendNotice(ctx, severityString, msg)
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.force_assertion_error": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"msg", types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598126)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(598128)
					return nil, errors.Newf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(598129)
				}
				__antithesis_instrumentation__.Notify(598127)
				msg := string(s)
				return nil, errors.AssertionFailedf("%s", msg)
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.void_func": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Void),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598130)
				return tree.DVoidDatum, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.force_panic": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"msg", types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598131)
				isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(598135)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598136)
				}
				__antithesis_instrumentation__.Notify(598132)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(598137)
					return nil, errInsufficientPriv
				} else {
					__antithesis_instrumentation__.Notify(598138)
				}
				__antithesis_instrumentation__.Notify(598133)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(598139)
					return nil, errors.Newf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(598140)
				}
				__antithesis_instrumentation__.Notify(598134)
				msg := string(s)

				colexecerror.NonCatchablePanic(msg)

				panic(msg)
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.force_log_fatal": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"msg", types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598141)
				isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(598145)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598146)
				}
				__antithesis_instrumentation__.Notify(598142)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(598147)
					return nil, errInsufficientPriv
				} else {
					__antithesis_instrumentation__.Notify(598148)
				}
				__antithesis_instrumentation__.Notify(598143)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(598149)
					return nil, errors.Newf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(598150)
				}
				__antithesis_instrumentation__.Notify(598144)
				msg := string(s)
				log.Fatalf(ctx.Ctx(), "force_log_fatal(): %s", msg)
				return nil, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.force_retry": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Interval}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598151)
				minDuration := args[0].(*tree.DInterval).Duration
				elapsed := duration.MakeDuration(int64(ctx.StmtTimestamp.Sub(ctx.TxnTimestamp)), 0, 0)
				if elapsed.Compare(minDuration) < 0 {
					__antithesis_instrumentation__.Notify(598153)
					return nil, ctx.Txn.GenerateForcedRetryableError(
						ctx.Ctx(), "forced by crdb_internal.force_retry()")
				} else {
					__antithesis_instrumentation__.Notify(598154)
				}
				__antithesis_instrumentation__.Notify(598152)
				return tree.DZero, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.lease_holder": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"key", types.Bytes}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598155)
				key := []byte(tree.MustBeDBytes(args[0]))
				b := &kv.Batch{}
				b.AddRawRequest(&roachpb.LeaseInfoRequest{
					RequestHeader: roachpb.RequestHeader{
						Key: key,
					},
				})
				if err := ctx.Txn.Run(ctx.Context, b); err != nil {
					__antithesis_instrumentation__.Notify(598157)
					return nil, pgerror.Wrap(err, pgcode.InvalidParameterValue, "error fetching leaseholder")
				} else {
					__antithesis_instrumentation__.Notify(598158)
				}
				__antithesis_instrumentation__.Notify(598156)
				resp := b.RawResponse().Responses[0].GetInner().(*roachpb.LeaseInfoResponse)

				return tree.NewDInt(tree.DInt(resp.Lease.Replica.StoreID)), nil
			},
			Info:       "This function is used to fetch the leaseholder corresponding to a request key",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.no_constant_folding": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.Any}},
			ReturnType: tree.IdentityReturnType(0),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598159)
				return args[0], nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.pretty_key": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"raw_key", types.Bytes},
				{"skip_fields", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598160)
				return tree.NewDString(catalogkeys.PrettyKey(
					nil,
					roachpb.Key(tree.MustBeDBytes(args[0])),
					int(tree.MustBeDInt(args[1])))), nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"crdb_internal.pretty_span": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"raw_key_start", types.Bytes},
				{"raw_key_end", types.Bytes},
				{"skip_fields", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598161)
				span := roachpb.Span{
					Key:    roachpb.Key(tree.MustBeDBytes(args[0])),
					EndKey: roachpb.Key(tree.MustBeDBytes(args[1])),
				}
				skip := int(tree.MustBeDInt(args[2]))
				return tree.NewDString(catalogkeys.PrettySpan(nil, span, skip)), nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"crdb_internal.range_stats": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"key", types.Bytes},
			},
			ReturnType: tree.FixedReturnType(types.Jsonb),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598162)
				key := []byte(tree.MustBeDBytes(args[0]))
				b := &kv.Batch{}
				b.AddRawRequest(&roachpb.RangeStatsRequest{
					RequestHeader: roachpb.RequestHeader{
						Key: key,
					},
				})
				if err := ctx.Txn.Run(ctx.Context, b); err != nil {
					__antithesis_instrumentation__.Notify(598166)
					return nil, pgerror.Wrap(err, pgcode.InvalidParameterValue, "error fetching range stats")
				} else {
					__antithesis_instrumentation__.Notify(598167)
				}
				__antithesis_instrumentation__.Notify(598163)
				resp := b.RawResponse().Responses[0].GetInner().(*roachpb.RangeStatsResponse).MVCCStats
				jsonStr, err := gojson.Marshal(&resp)
				if err != nil {
					__antithesis_instrumentation__.Notify(598168)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598169)
				}
				__antithesis_instrumentation__.Notify(598164)
				jsonDatum, err := tree.ParseDJSON(string(jsonStr))
				if err != nil {
					__antithesis_instrumentation__.Notify(598170)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598171)
				}
				__antithesis_instrumentation__.Notify(598165)
				return jsonDatum, nil
			},
			Info:       "This function is used to retrieve range statistics information as a JSON object.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.get_namespace_id": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{{"parent_id", types.Int}, {"name", types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598172)
				parentID := tree.MustBeDInt(args[0])
				name := tree.MustBeDString(args[1])
				id, found, err := ctx.PrivilegedAccessor.LookupNamespaceID(
					ctx.Context,
					int64(parentID),
					0,
					string(name),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(598175)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598176)
				}
				__antithesis_instrumentation__.Notify(598173)
				if !found {
					__antithesis_instrumentation__.Notify(598177)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(598178)
				}
				__antithesis_instrumentation__.Notify(598174)
				return tree.NewDInt(id), nil
			},
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"parent_id", types.Int},
				{"parent_schema_id", types.Int},
				{"name", types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598179)
				parentID := tree.MustBeDInt(args[0])
				parentSchemaID := tree.MustBeDInt(args[1])
				name := tree.MustBeDString(args[2])
				id, found, err := ctx.PrivilegedAccessor.LookupNamespaceID(
					ctx.Context,
					int64(parentID),
					int64(parentSchemaID),
					string(name),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(598182)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598183)
				}
				__antithesis_instrumentation__.Notify(598180)
				if !found {
					__antithesis_instrumentation__.Notify(598184)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(598185)
				}
				__antithesis_instrumentation__.Notify(598181)
				return tree.NewDInt(id), nil
			},
			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.get_database_id": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{{"name", types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598186)
				name := tree.MustBeDString(args[0])
				id, found, err := ctx.PrivilegedAccessor.LookupNamespaceID(
					ctx.Context,
					int64(0),
					int64(0),
					string(name),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(598189)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598190)
				}
				__antithesis_instrumentation__.Notify(598187)
				if !found {
					__antithesis_instrumentation__.Notify(598191)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(598192)
				}
				__antithesis_instrumentation__.Notify(598188)
				return tree.NewDInt(id), nil
			},
			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.get_zone_config": makeBuiltin(
		tree.FunctionProperties{Category: categorySystemInfo},
		tree.Overload{
			Types:      tree.ArgTypes{{"namespace_id", types.Int}},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598193)
				id := tree.MustBeDInt(args[0])
				bytes, found, err := ctx.PrivilegedAccessor.LookupZoneConfigByNamespaceID(
					ctx.Context,
					int64(id),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(598196)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598197)
				}
				__antithesis_instrumentation__.Notify(598194)
				if !found {
					__antithesis_instrumentation__.Notify(598198)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(598199)
				}
				__antithesis_instrumentation__.Notify(598195)
				return tree.NewDBytes(bytes), nil
			},
			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.set_vmodule": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"vmodule_string", types.String}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598200)
				isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(598204)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598205)
				}
				__antithesis_instrumentation__.Notify(598201)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(598206)
					return nil, errInsufficientPriv
				} else {
					__antithesis_instrumentation__.Notify(598207)
				}
				__antithesis_instrumentation__.Notify(598202)

				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(598208)
					return nil, errors.Newf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(598209)
				}
				__antithesis_instrumentation__.Notify(598203)
				vmodule := string(s)
				return tree.DZero, log.SetVModule(vmodule)
			},
			Info: "Set the equivalent of the `--vmodule` flag on the gateway node processing this request; " +
				"it affords control over the logging verbosity of different files. " +
				"Example syntax: `crdb_internal.set_vmodule('recordio=2,file=1,gfs*=3')`. " +
				"Reset with: `crdb_internal.set_vmodule('')`. " +
				"Raising the verbosity can severely affect performance.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.get_vmodule": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598210)

				isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(598213)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598214)
				}
				__antithesis_instrumentation__.Notify(598211)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(598215)
					return nil, errInsufficientPriv
				} else {
					__antithesis_instrumentation__.Notify(598216)
				}
				__antithesis_instrumentation__.Notify(598212)
				return tree.NewDString(log.GetVModule()), nil
			},
			Info:       "Returns the vmodule configuration on the gateway node processing this request.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.num_geo_inverted_index_entries": makeBuiltin(
		tree.FunctionProperties{
			Category:     categorySystemInfo,
			NullableArgs: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"table_id", types.Int},
				{"index_id", types.Int},
				{"val", types.Geography},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598217)
				if args[0] == tree.DNull || func() bool {
					__antithesis_instrumentation__.Notify(598223)
					return args[1] == tree.DNull == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(598224)
					return args[2] == tree.DNull == true
				}() == true {
					__antithesis_instrumentation__.Notify(598225)
					return tree.DZero, nil
				} else {
					__antithesis_instrumentation__.Notify(598226)
				}
				__antithesis_instrumentation__.Notify(598218)
				tableID := int(tree.MustBeDInt(args[0]))
				indexID := int(tree.MustBeDInt(args[1]))
				g := tree.MustBeDGeography(args[2])

				cf := descs.NewBareBonesCollectionFactory(ctx.Settings, ctx.Codec)
				descsCol := cf.MakeCollection(ctx.Context, descs.NewTemporarySchemaProvider(ctx.SessionDataStack))
				tableDesc, err := descsCol.Direct().MustGetTableDescByID(ctx.Ctx(), ctx.Txn, descpb.ID(tableID))
				if err != nil {
					__antithesis_instrumentation__.Notify(598227)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598228)
				}
				__antithesis_instrumentation__.Notify(598219)
				index, err := tableDesc.FindIndexWithID(descpb.IndexID(indexID))
				if err != nil {
					__antithesis_instrumentation__.Notify(598229)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598230)
				}
				__antithesis_instrumentation__.Notify(598220)
				if index.GetGeoConfig().S2Geography == nil {
					__antithesis_instrumentation__.Notify(598231)
					return nil, errors.Errorf("index_id %d is not a geography inverted index", indexID)
				} else {
					__antithesis_instrumentation__.Notify(598232)
				}
				__antithesis_instrumentation__.Notify(598221)
				keys, err := rowenc.EncodeGeoInvertedIndexTableKeys(g, nil, index.GetGeoConfig())
				if err != nil {
					__antithesis_instrumentation__.Notify(598233)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598234)
				}
				__antithesis_instrumentation__.Notify(598222)
				return tree.NewDInt(tree.DInt(len(keys))), nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"table_id", types.Int},
				{"index_id", types.Int},
				{"val", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598235)
				if args[0] == tree.DNull || func() bool {
					__antithesis_instrumentation__.Notify(598241)
					return args[1] == tree.DNull == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(598242)
					return args[2] == tree.DNull == true
				}() == true {
					__antithesis_instrumentation__.Notify(598243)
					return tree.DZero, nil
				} else {
					__antithesis_instrumentation__.Notify(598244)
				}
				__antithesis_instrumentation__.Notify(598236)
				tableID := int(tree.MustBeDInt(args[0]))
				indexID := int(tree.MustBeDInt(args[1]))
				g := tree.MustBeDGeometry(args[2])

				cf := descs.NewBareBonesCollectionFactory(ctx.Settings, ctx.Codec)
				descsCol := cf.MakeCollection(ctx.Context, descs.NewTemporarySchemaProvider(ctx.SessionDataStack))
				tableDesc, err := descsCol.Direct().MustGetTableDescByID(ctx.Ctx(), ctx.Txn, descpb.ID(tableID))
				if err != nil {
					__antithesis_instrumentation__.Notify(598245)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598246)
				}
				__antithesis_instrumentation__.Notify(598237)
				index, err := tableDesc.FindIndexWithID(descpb.IndexID(indexID))
				if err != nil {
					__antithesis_instrumentation__.Notify(598247)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598248)
				}
				__antithesis_instrumentation__.Notify(598238)
				if index.GetGeoConfig().S2Geometry == nil {
					__antithesis_instrumentation__.Notify(598249)
					return nil, errors.Errorf("index_id %d is not a geometry inverted index", indexID)
				} else {
					__antithesis_instrumentation__.Notify(598250)
				}
				__antithesis_instrumentation__.Notify(598239)
				keys, err := rowenc.EncodeGeoInvertedIndexTableKeys(g, nil, index.GetGeoConfig())
				if err != nil {
					__antithesis_instrumentation__.Notify(598251)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598252)
				}
				__antithesis_instrumentation__.Notify(598240)
				return tree.NewDInt(tree.DInt(len(keys))), nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityStable,
		}),

	"crdb_internal.num_inverted_index_entries": makeBuiltin(
		tree.FunctionProperties{
			Category:     categorySystemInfo,
			NullableArgs: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Jsonb}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598253)
				return jsonNumInvertedIndexEntries(ctx, args[0])
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.AnyArray}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598254)
				return arrayNumInvertedIndexEntries(ctx, args[0], tree.DNull)
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"val", types.Jsonb},
				{"version", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598255)

				return jsonNumInvertedIndexEntries(ctx, args[0])
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"val", types.AnyArray},
				{"version", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598256)
				return arrayNumInvertedIndexEntries(ctx, args[0], args[1])
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityStable,
		}),

	"crdb_internal.is_admin": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemInfo,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(evalCtx *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598257)
				if evalCtx.SessionAccessor == nil {
					__antithesis_instrumentation__.Notify(598260)
					return nil, errors.AssertionFailedf("session accessor not set")
				} else {
					__antithesis_instrumentation__.Notify(598261)
				}
				__antithesis_instrumentation__.Notify(598258)
				ctx := evalCtx.Ctx()
				isAdmin, err := evalCtx.SessionAccessor.HasAdminRole(ctx)
				if err != nil {
					__antithesis_instrumentation__.Notify(598262)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598263)
				}
				__antithesis_instrumentation__.Notify(598259)
				return tree.MakeDBool(tree.DBool(isAdmin)), nil
			},
			Info:       "Retrieves the current user's admin status.",
			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.has_role_option": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemInfo,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"option", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598264)
				if evalCtx.SessionAccessor == nil {
					__antithesis_instrumentation__.Notify(598268)
					return nil, errors.AssertionFailedf("session accessor not set")
				} else {
					__antithesis_instrumentation__.Notify(598269)
				}
				__antithesis_instrumentation__.Notify(598265)
				optionStr := string(tree.MustBeDString(args[0]))
				option, ok := roleoption.ByName[optionStr]
				if !ok {
					__antithesis_instrumentation__.Notify(598270)
					return nil, errors.Newf("unrecognized role option %s", optionStr)
				} else {
					__antithesis_instrumentation__.Notify(598271)
				}
				__antithesis_instrumentation__.Notify(598266)
				ctx := evalCtx.Ctx()
				ok, err := evalCtx.SessionAccessor.HasRoleOption(ctx, option)
				if err != nil {
					__antithesis_instrumentation__.Notify(598272)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598273)
				}
				__antithesis_instrumentation__.Notify(598267)
				return tree.MakeDBool(tree.DBool(ok)), nil
			},
			Info:       "Returns whether the current user has the specified role option",
			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.assignment_cast": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,

			NullableArgs: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"val", types.Any},
				{"type", types.Any},
			},
			ReturnType: tree.IdentityReturnType(1),
			FnWithExprs: func(evalCtx *tree.EvalContext, args tree.Exprs) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598274)
				targetType := args[1].(tree.TypedExpr).ResolvedType()
				val, err := args[0].(tree.TypedExpr).Eval(evalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(598276)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598277)
				}
				__antithesis_instrumentation__.Notify(598275)
				return tree.PerformAssignmentCast(evalCtx, val, targetType)
			},
			Info: "This function is used internally to perform assignment casts during mutations.",

			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.round_decimal_values": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"val", types.Decimal},
				{"scale", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Decimal),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598278)
				value := args[0].(*tree.DDecimal)
				scale := int32(tree.MustBeDInt(args[1]))
				return roundDDecimal(value, scale)
			},
			Info:       "This function is used internally to round decimal values during mutations.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"val", types.DecimalArray},
				{"scale", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.DecimalArray),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598279)
				value := args[0].(*tree.DArray)
				scale := int32(tree.MustBeDInt(args[1]))

				var newArr tree.Datums
				for i, elem := range value.Array {
					__antithesis_instrumentation__.Notify(598282)

					if elem == tree.DNull {
						__antithesis_instrumentation__.Notify(598285)
						continue
					} else {
						__antithesis_instrumentation__.Notify(598286)
					}
					__antithesis_instrumentation__.Notify(598283)

					rounded, err := roundDDecimal(elem.(*tree.DDecimal), scale)
					if err != nil {
						__antithesis_instrumentation__.Notify(598287)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(598288)
					}
					__antithesis_instrumentation__.Notify(598284)
					if rounded != elem {
						__antithesis_instrumentation__.Notify(598289)
						if newArr == nil {
							__antithesis_instrumentation__.Notify(598291)
							newArr = make(tree.Datums, len(value.Array))
							copy(newArr, value.Array)
						} else {
							__antithesis_instrumentation__.Notify(598292)
						}
						__antithesis_instrumentation__.Notify(598290)
						newArr[i] = rounded
					} else {
						__antithesis_instrumentation__.Notify(598293)
					}
				}
				__antithesis_instrumentation__.Notify(598280)
				if newArr != nil {
					__antithesis_instrumentation__.Notify(598294)
					ret := &tree.DArray{}
					*ret = *value
					ret.Array = newArr
					return ret, nil
				} else {
					__antithesis_instrumentation__.Notify(598295)
				}
				__antithesis_instrumentation__.Notify(598281)
				return value, nil
			},
			Info:       "This function is used internally to round decimal array values during mutations.",
			Volatility: tree.VolatilityStable,
		},
	),
	"crdb_internal.completed_migrations": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.StringArray),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598296)
				prefix := ctx.Codec.MigrationKeyPrefix()
				keyvals, err := ctx.Txn.Scan(ctx.Context, prefix, prefix.PrefixEnd(), 0)
				if err != nil {
					__antithesis_instrumentation__.Notify(598299)
					return nil, errors.Wrapf(err, "failed to get list of completed migrations")
				} else {
					__antithesis_instrumentation__.Notify(598300)
				}
				__antithesis_instrumentation__.Notify(598297)
				ret := &tree.DArray{ParamTyp: types.String, Array: make(tree.Datums, 0, len(keyvals))}
				for _, keyval := range keyvals {
					__antithesis_instrumentation__.Notify(598301)
					key := keyval.Key
					if len(key) > len(keys.MigrationPrefix) {
						__antithesis_instrumentation__.Notify(598303)
						key = key[len(keys.MigrationPrefix):]
					} else {
						__antithesis_instrumentation__.Notify(598304)
					}
					__antithesis_instrumentation__.Notify(598302)
					ret.Array = append(ret.Array, tree.NewDString(string(key)))
				}
				__antithesis_instrumentation__.Notify(598298)
				return ret, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),
	"crdb_internal.unsafe_upsert_descriptor": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemRepair,
			DistsqlBlocklist: true,
			Undocumented:     true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"id", types.Int},
				{"desc", types.Bytes},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598305)
				if err := ctx.Planner.UnsafeUpsertDescriptor(ctx.Context,
					int64(*args[0].(*tree.DInt)),
					[]byte(*args[1].(*tree.DBytes)),
					false); err != nil {
					__antithesis_instrumentation__.Notify(598307)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598308)
				}
				__antithesis_instrumentation__.Notify(598306)
				return tree.DBoolTrue, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"id", types.Int},
				{"desc", types.Bytes},
				{"force", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598309)
				if err := ctx.Planner.UnsafeUpsertDescriptor(ctx.Context,
					int64(*args[0].(*tree.DInt)),
					[]byte(*args[1].(*tree.DBytes)),
					bool(*args[2].(*tree.DBool))); err != nil {
					__antithesis_instrumentation__.Notify(598311)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598312)
				}
				__antithesis_instrumentation__.Notify(598310)
				return tree.DBoolTrue, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),
	"crdb_internal.unsafe_delete_descriptor": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemRepair,
			DistsqlBlocklist: true,
			Undocumented:     true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"id", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598313)
				if err := ctx.Planner.UnsafeDeleteDescriptor(ctx.Context,
					int64(*args[0].(*tree.DInt)),
					false,
				); err != nil {
					__antithesis_instrumentation__.Notify(598315)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598316)
				}
				__antithesis_instrumentation__.Notify(598314)
				return tree.DBoolTrue, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"id", types.Int},
				{"force", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598317)
				if err := ctx.Planner.UnsafeDeleteDescriptor(ctx.Context,
					int64(*args[0].(*tree.DInt)),
					bool(*args[1].(*tree.DBool)),
				); err != nil {
					__antithesis_instrumentation__.Notify(598319)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598320)
				}
				__antithesis_instrumentation__.Notify(598318)
				return tree.DBoolTrue, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),
	"crdb_internal.unsafe_upsert_namespace_entry": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemRepair,
			DistsqlBlocklist: true,
			Undocumented:     true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"parent_id", types.Int},
				{"parent_schema_id", types.Int},
				{"name", types.String},
				{"desc_id", types.Int},
				{"force", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598321)
				if err := ctx.Planner.UnsafeUpsertNamespaceEntry(
					ctx.Context,
					int64(*args[0].(*tree.DInt)),
					int64(*args[1].(*tree.DInt)),
					string(*args[2].(*tree.DString)),
					int64(*args[3].(*tree.DInt)),
					bool(*args[4].(*tree.DBool)),
				); err != nil {
					__antithesis_instrumentation__.Notify(598323)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598324)
				}
				__antithesis_instrumentation__.Notify(598322)
				return tree.DBoolTrue, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"parent_id", types.Int},
				{"parent_schema_id", types.Int},
				{"name", types.String},
				{"desc_id", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598325)
				if err := ctx.Planner.UnsafeUpsertNamespaceEntry(
					ctx.Context,
					int64(*args[0].(*tree.DInt)),
					int64(*args[1].(*tree.DInt)),
					string(*args[2].(*tree.DString)),
					int64(*args[3].(*tree.DInt)),
					false,
				); err != nil {
					__antithesis_instrumentation__.Notify(598327)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598328)
				}
				__antithesis_instrumentation__.Notify(598326)
				return tree.DBoolTrue, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),
	"crdb_internal.unsafe_delete_namespace_entry": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemRepair,
			DistsqlBlocklist: true,
			Undocumented:     true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"parent_id", types.Int},
				{"parent_schema_id", types.Int},
				{"name", types.String},
				{"desc_id", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598329)
				if err := ctx.Planner.UnsafeDeleteNamespaceEntry(
					ctx.Context,
					int64(*args[0].(*tree.DInt)),
					int64(*args[1].(*tree.DInt)),
					string(*args[2].(*tree.DString)),
					int64(*args[3].(*tree.DInt)),
					false,
				); err != nil {
					__antithesis_instrumentation__.Notify(598331)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598332)
				}
				__antithesis_instrumentation__.Notify(598330)
				return tree.DBoolTrue, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"parent_id", types.Int},
				{"parent_schema_id", types.Int},
				{"name", types.String},
				{"desc_id", types.Int},
				{"force", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598333)
				if err := ctx.Planner.UnsafeDeleteNamespaceEntry(
					ctx.Context,
					int64(*args[0].(*tree.DInt)),
					int64(*args[1].(*tree.DInt)),
					string(*args[2].(*tree.DString)),
					int64(*args[3].(*tree.DInt)),
					bool(*args[4].(*tree.DBool)),
				); err != nil {
					__antithesis_instrumentation__.Notify(598335)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598336)
				}
				__antithesis_instrumentation__.Notify(598334)
				return tree.DBoolTrue, nil
			},
			Info:       "This function is used only by CockroachDB's developers for testing purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.sql_liveness_is_alive": makeBuiltin(
		tree.FunctionProperties{Category: categoryMultiTenancy},
		tree.Overload{
			Types:      tree.ArgTypes{{"session_id", types.Bytes}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598337)
				sid := sqlliveness.SessionID(*(args[0].(*tree.DBytes)))
				live, err := evalCtx.SQLLivenessReader.IsAlive(evalCtx.Context, sid)
				if err != nil {
					__antithesis_instrumentation__.Notify(598339)
					return tree.MakeDBool(true), err
				} else {
					__antithesis_instrumentation__.Notify(598340)
				}
				__antithesis_instrumentation__.Notify(598338)
				return tree.MakeDBool(tree.DBool(live)), nil
			},
			Info:       "Checks is given sqlliveness session id is not expired",
			Volatility: tree.VolatilityStable,
		},
	),

	"crdb_internal.gc_tenant": makeBuiltin(

		tree.FunctionProperties{
			Category:     categoryMultiTenancy,
			Undocumented: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"id", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598341)
				sTenID, err := mustBeDIntInTenantRange(args[0])
				if err != nil {
					__antithesis_instrumentation__.Notify(598344)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598345)
				}
				__antithesis_instrumentation__.Notify(598342)
				if err := ctx.Tenant.GCTenant(ctx.Context, uint64(sTenID)); err != nil {
					__antithesis_instrumentation__.Notify(598346)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598347)
				}
				__antithesis_instrumentation__.Notify(598343)
				return args[0], nil
			},
			Info:       "Garbage collects a tenant with the provided ID. Must be run by the System tenant.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.update_tenant_resource_limits": makeBuiltin(
		tree.FunctionProperties{
			Category:     categoryMultiTenancy,
			Undocumented: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"tenant_id", types.Int},
				{"available_request_units", types.Float},
				{"refill_rate", types.Float},
				{"max_burst_request_units", types.Float},
				{"as_of", types.Timestamp},
				{"as_of_consumed_request_units", types.Float},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598348)
				sTenID, err := mustBeDIntInTenantRange(args[0])
				if err != nil {
					__antithesis_instrumentation__.Notify(598351)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598352)
				}
				__antithesis_instrumentation__.Notify(598349)
				availableRU := float64(tree.MustBeDFloat(args[1]))
				refillRate := float64(tree.MustBeDFloat(args[2]))
				maxBurstRU := float64(tree.MustBeDFloat(args[3]))
				asOf := tree.MustBeDTimestamp(args[4]).Time
				asOfConsumed := float64(tree.MustBeDFloat(args[5]))

				if err := ctx.Tenant.UpdateTenantResourceLimits(
					ctx.Context,
					uint64(sTenID),
					availableRU,
					refillRate,
					maxBurstRU,
					asOf,
					asOfConsumed,
				); err != nil {
					__antithesis_instrumentation__.Notify(598353)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598354)
				}
				__antithesis_instrumentation__.Notify(598350)
				return args[0], nil
			},
			Info:       "Updates resource limits for the tenant with the provided ID. Must be run by the System tenant.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.compact_engine_span": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemRepair,
			DistsqlBlocklist: true,
			Undocumented:     true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"node_id", types.Int},
				{"store_id", types.Int},
				{"start_key", types.Bytes},
				{"end_key", types.Bytes},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598355)
				nodeID := int32(tree.MustBeDInt(args[0]))
				storeID := int32(tree.MustBeDInt(args[1]))
				startKey := []byte(tree.MustBeDBytes(args[2]))
				endKey := []byte(tree.MustBeDBytes(args[3]))
				if err := ctx.CompactEngineSpan(
					ctx.Context, nodeID, storeID, startKey, endKey); err != nil {
					__antithesis_instrumentation__.Notify(598357)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598358)
				}
				__antithesis_instrumentation__.Notify(598356)
				return tree.DBoolTrue, nil
			},
			Info: "This function is used only by CockroachDB's developers for restoring engine health. " +
				"It is used to compact a span of the engine at the given node and store. The start and " +
				"end keys are bytes. To compact a particular rangeID, one can do: " +
				"SELECT crdb_internal.compact_engine_span(<node>, <store>, start_key, end_key) " +
				"FROM crdb_internal.ranges_no_leases WHERE range_id=<value>. If one has hex or escape " +
				"formatted bytea, one can use decode(<key string>, 'hex'|'escape') as the parameter. " +
				"The compaction is run synchronously, so this function may take a long time to return. " +
				"One can use the logs at the node to confirm that a compaction has started.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.increment_feature_counter": makeBuiltin(
		tree.FunctionProperties{
			Category:     categorySystemInfo,
			Undocumented: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"feature", types.String}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598359)
				s, ok := tree.AsDString(args[0])
				if !ok {
					__antithesis_instrumentation__.Notify(598361)
					return nil, errors.Newf("expected string value, got %T", args[0])
				} else {
					__antithesis_instrumentation__.Notify(598362)
				}
				__antithesis_instrumentation__.Notify(598360)
				feature := string(s)
				telemetry.Inc(sqltelemetry.HashedFeatureCounter(feature))
				return tree.DBoolTrue, nil
			},
			Info: "This function can be used to report the usage of an arbitrary feature. The " +
				"feature name is hashed for privacy purposes.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"num_nulls": makeBuiltin(
		tree.FunctionProperties{
			Category:     categoryComparison,
			NullableArgs: true,
		},
		tree.Overload{
			Types: tree.VariadicType{
				VarType: types.Any,
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598363)
				var numNulls int
				for _, arg := range args {
					__antithesis_instrumentation__.Notify(598365)
					if arg == tree.DNull {
						__antithesis_instrumentation__.Notify(598366)
						numNulls++
					} else {
						__antithesis_instrumentation__.Notify(598367)
					}
				}
				__antithesis_instrumentation__.Notify(598364)
				return tree.NewDInt(tree.DInt(numNulls)), nil
			},
			Info:       "Returns the number of null arguments.",
			Volatility: tree.VolatilityImmutable,
		},
	),
	"num_nonnulls": makeBuiltin(
		tree.FunctionProperties{
			Category:     categoryComparison,
			NullableArgs: true,
		},
		tree.Overload{
			Types: tree.VariadicType{
				VarType: types.Any,
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598368)
				var numNonNulls int
				for _, arg := range args {
					__antithesis_instrumentation__.Notify(598370)
					if arg != tree.DNull {
						__antithesis_instrumentation__.Notify(598371)
						numNonNulls++
					} else {
						__antithesis_instrumentation__.Notify(598372)
					}
				}
				__antithesis_instrumentation__.Notify(598369)
				return tree.NewDInt(tree.DInt(numNonNulls)), nil
			},
			Info:       "Returns the number of nonnull arguments.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	GatewayRegionBuiltinName: makeBuiltin(
		tree.FunctionProperties{
			Category: categoryMultiRegion,

			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, arg tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598373)
				region, found := evalCtx.Locality.Find("region")
				if !found {
					__antithesis_instrumentation__.Notify(598375)
					return nil, pgerror.Newf(
						pgcode.ConfigFile,
						"no region set on the locality flag on this node",
					)
				} else {
					__antithesis_instrumentation__.Notify(598376)
				}
				__antithesis_instrumentation__.Notify(598374)
				return tree.NewDString(region), nil
			},
			Info: `Returns the region of the connection's current node as defined by
the locality flag on node startup. Returns an error if no region is set.`,
			Volatility: tree.VolatilityStable,
		},
	),
	DefaultToDatabasePrimaryRegionBuiltinName: makeBuiltin(
		tree.FunctionProperties{Category: categoryMultiRegion},
		stringOverload1(
			func(evalCtx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598377)
				regionConfig, err := evalCtx.Regions.CurrentDatabaseRegionConfig(evalCtx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(598381)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598382)
				}
				__antithesis_instrumentation__.Notify(598378)
				if regionConfig == nil {
					__antithesis_instrumentation__.Notify(598383)
					return nil, pgerror.Newf(
						pgcode.InvalidDatabaseDefinition,
						"current database %s is not multi-region enabled",
						evalCtx.SessionData().Database,
					)
				} else {
					__antithesis_instrumentation__.Notify(598384)
				}
				__antithesis_instrumentation__.Notify(598379)
				if regionConfig.IsValidRegionNameString(s) {
					__antithesis_instrumentation__.Notify(598385)
					return tree.NewDString(s), nil
				} else {
					__antithesis_instrumentation__.Notify(598386)
				}
				__antithesis_instrumentation__.Notify(598380)
				primaryRegion := regionConfig.PrimaryRegionString()
				return tree.NewDString(primaryRegion), nil
			},
			types.String,
			`Returns the given region if the region has been added to the current database.
	Otherwise, this will return the primary region of the current database.
	This will error if the current database is not a multi-region database.`,
			tree.VolatilityStable,
		),
	),
	RehomeRowBuiltinName: makeBuiltin(
		tree.FunctionProperties{Category: categoryMultiRegion},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, arg tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598387)
				regionConfig, err := evalCtx.Regions.CurrentDatabaseRegionConfig(evalCtx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(598392)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598393)
				}
				__antithesis_instrumentation__.Notify(598388)
				if regionConfig == nil {
					__antithesis_instrumentation__.Notify(598394)
					return nil, pgerror.Newf(
						pgcode.InvalidDatabaseDefinition,
						"current database %s is not multi-region enabled",
						evalCtx.SessionData().Database,
					)
				} else {
					__antithesis_instrumentation__.Notify(598395)
				}
				__antithesis_instrumentation__.Notify(598389)
				gatewayRegion, found := evalCtx.Locality.Find("region")
				if !found {
					__antithesis_instrumentation__.Notify(598396)
					return nil, pgerror.Newf(
						pgcode.ConfigFile,
						"no region set on the locality flag on this node",
					)
				} else {
					__antithesis_instrumentation__.Notify(598397)
				}
				__antithesis_instrumentation__.Notify(598390)
				if regionConfig.IsValidRegionNameString(gatewayRegion) {
					__antithesis_instrumentation__.Notify(598398)
					return tree.NewDString(gatewayRegion), nil
				} else {
					__antithesis_instrumentation__.Notify(598399)
				}
				__antithesis_instrumentation__.Notify(598391)
				primaryRegion := regionConfig.PrimaryRegionString()
				return tree.NewDString(primaryRegion), nil
			},
			Info: `Returns the region of the connection's current node as defined by
the locality flag on node startup. Returns an error if no region is set.`,
			Volatility:       tree.VolatilityStable,
			DistsqlBlocklist: true,
		},
	),
	"crdb_internal.validate_multi_region_zone_configs": makeBuiltin(
		tree.FunctionProperties{Category: categoryMultiRegion},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598400)
				if err := evalCtx.Regions.ValidateAllMultiRegionZoneConfigsInCurrentDatabase(
					evalCtx.Context,
				); err != nil {
					__antithesis_instrumentation__.Notify(598402)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598403)
				}
				__antithesis_instrumentation__.Notify(598401)
				return tree.MakeDBool(true), nil
			},
			Info: `Validates all multi-region zone configurations are correctly setup
			for the current database, including all tables, indexes and partitions underneath.
			Returns an error if validation fails. This builtin uses un-leased versions of the
			each descriptor, requiring extra round trips.`,
			Volatility: tree.VolatilityVolatile,
		},
	),
	"crdb_internal.reset_multi_region_zone_configs_for_table": makeBuiltin(
		tree.FunctionProperties{Category: categoryMultiRegion},
		tree.Overload{
			Types:      tree.ArgTypes{{"id", types.Int}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598404)
				id := int64(*args[0].(*tree.DInt))

				if err := evalCtx.Regions.ResetMultiRegionZoneConfigsForTable(
					evalCtx.Context,
					id,
				); err != nil {
					__antithesis_instrumentation__.Notify(598406)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598407)
				}
				__antithesis_instrumentation__.Notify(598405)
				return tree.MakeDBool(true), nil
			},
			Info: `Resets the zone configuration for a multi-region table to
match its original state. No-ops if the given table ID is not a multi-region
table.`,
			Volatility: tree.VolatilityVolatile,
		},
	),
	"crdb_internal.reset_multi_region_zone_configs_for_database": makeBuiltin(
		tree.FunctionProperties{Category: categoryMultiRegion},
		tree.Overload{
			Types:      tree.ArgTypes{{"id", types.Int}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598408)
				id := int64(*args[0].(*tree.DInt))

				if err := evalCtx.Regions.ResetMultiRegionZoneConfigsForDatabase(
					evalCtx.Context,
					id,
				); err != nil {
					__antithesis_instrumentation__.Notify(598410)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598411)
				}
				__antithesis_instrumentation__.Notify(598409)
				return tree.MakeDBool(true), nil
			},
			Info: `Resets the zone configuration for a multi-region database to
match its original state. No-ops if the given database ID is not multi-region
enabled.`,
			Volatility: tree.VolatilityVolatile,
		},
	),
	"crdb_internal.filter_multiregion_fields_from_zone_config_sql": makeBuiltin(
		tree.FunctionProperties{Category: categoryMultiRegion},
		stringOverload1(
			func(evalCtx *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598412)
				stmt, err := parser.ParseOne(s)
				if err != nil {
					__antithesis_instrumentation__.Notify(598417)

					return tree.NewDString(s), nil
				} else {
					__antithesis_instrumentation__.Notify(598418)
				}
				__antithesis_instrumentation__.Notify(598413)
				zs, ok := stmt.AST.(*tree.SetZoneConfig)
				if !ok {
					__antithesis_instrumentation__.Notify(598419)
					return nil, errors.Newf("invalid CONFIGURE ZONE statement (type %T): %s", stmt.AST, stmt)
				} else {
					__antithesis_instrumentation__.Notify(598420)
				}
				__antithesis_instrumentation__.Notify(598414)
				newKVOptions := zs.Options[:0]
				for _, opt := range zs.Options {
					__antithesis_instrumentation__.Notify(598421)
					if _, ok := zonepb.MultiRegionZoneConfigFieldsSet[opt.Key]; !ok {
						__antithesis_instrumentation__.Notify(598422)
						newKVOptions = append(newKVOptions, opt)
					} else {
						__antithesis_instrumentation__.Notify(598423)
					}
				}
				__antithesis_instrumentation__.Notify(598415)
				if len(newKVOptions) == 0 {
					__antithesis_instrumentation__.Notify(598424)
					return tree.DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(598425)
				}
				__antithesis_instrumentation__.Notify(598416)
				zs.Options = newKVOptions
				return tree.NewDString(zs.String()), nil
			},
			types.String,
			`Takes in a CONFIGURE ZONE SQL statement and returns a modified
SQL statement omitting multi-region related zone configuration fields.
If the CONFIGURE ZONE statement can be inferred by the database's or
table's zone configuration this will return NULL.`,
			tree.VolatilityStable,
		),
	),
	"crdb_internal.reset_index_usage_stats": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemInfo,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598426)
				isAdmin, err := evalCtx.SessionAccessor.HasAdminRole(evalCtx.Ctx())
				if err != nil {
					__antithesis_instrumentation__.Notify(598431)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598432)
				}
				__antithesis_instrumentation__.Notify(598427)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(598433)
					return nil, errors.New("crdb_internal.reset_index_usage_stats() requires admin privilege")
				} else {
					__antithesis_instrumentation__.Notify(598434)
				}
				__antithesis_instrumentation__.Notify(598428)
				if evalCtx.IndexUsageStatsController == nil {
					__antithesis_instrumentation__.Notify(598435)
					return nil, errors.AssertionFailedf("index usage stats controller not set")
				} else {
					__antithesis_instrumentation__.Notify(598436)
				}
				__antithesis_instrumentation__.Notify(598429)
				ctx := evalCtx.Ctx()
				if err := evalCtx.IndexUsageStatsController.ResetIndexUsageStats(ctx); err != nil {
					__antithesis_instrumentation__.Notify(598437)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598438)
				}
				__antithesis_instrumentation__.Notify(598430)
				return tree.MakeDBool(true), nil
			},
			Info:       `This function is used to clear the collected index usage statistics.`,
			Volatility: tree.VolatilityVolatile,
		},
	),
	"crdb_internal.reset_sql_stats": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemInfo,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598439)
				isAdmin, err := evalCtx.SessionAccessor.HasAdminRole(evalCtx.Ctx())
				if err != nil {
					__antithesis_instrumentation__.Notify(598444)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598445)
				}
				__antithesis_instrumentation__.Notify(598440)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(598446)
					return nil, errors.New("crdb_internal.reset_sql_stats() requires admin privilege")
				} else {
					__antithesis_instrumentation__.Notify(598447)
				}
				__antithesis_instrumentation__.Notify(598441)
				if evalCtx.SQLStatsController == nil {
					__antithesis_instrumentation__.Notify(598448)
					return nil, errors.AssertionFailedf("sql stats controller not set")
				} else {
					__antithesis_instrumentation__.Notify(598449)
				}
				__antithesis_instrumentation__.Notify(598442)
				ctx := evalCtx.Ctx()
				if err := evalCtx.SQLStatsController.ResetClusterSQLStats(ctx); err != nil {
					__antithesis_instrumentation__.Notify(598450)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598451)
				}
				__antithesis_instrumentation__.Notify(598443)
				return tree.MakeDBool(true), nil
			},
			Info:       `This function is used to clear the collected SQL statistics.`,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.force_delete_table_data": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemRepair,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"id", types.Int}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598452)
				id := int64(*args[0].(*tree.DInt))

				err := ctx.Planner.ForceDeleteTableData(ctx.Context, id)
				if err != nil {
					__antithesis_instrumentation__.Notify(598454)
					return tree.DBoolFalse, err
				} else {
					__antithesis_instrumentation__.Notify(598455)
				}
				__antithesis_instrumentation__.Notify(598453)
				return tree.DBoolTrue, err
			},
			Info:       "This function can be used to clear the data belonging to a table, when the table cannot be dropped.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.serialize_session": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598456)
				return evalCtx.Planner.SerializeSessionState()
			},
			Info:       `This function serializes the variables in the current session.`,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.deserialize_session": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"session", types.Bytes}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598457)
				state := tree.MustBeDBytes(args[0])
				return evalCtx.Planner.DeserializeSessionState(tree.NewDBytes(state))
			},
			Info:       `This function deserializes the serialized variables into the current session.`,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.create_session_revival_token": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598458)
				return evalCtx.Planner.CreateSessionRevivalToken()
			},
			Info:       `Generate a token that can be used to create a new session for the current user.`,
			Volatility: tree.VolatilityVolatile,
		},
	),
	"crdb_internal.validate_session_revival_token": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"token", types.Bytes}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598459)
				token := tree.MustBeDBytes(args[0])
				return evalCtx.Planner.ValidateSessionRevivalToken(&token)
			},
			Info:       `Validate a token that was created by create_session_revival_token. Intended for testing.`,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.validate_ttl_scheduled_jobs": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Void),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598460)
				return tree.DVoidDatum, evalCtx.Planner.ValidateTTLScheduledJobsInCurrentDB(evalCtx.Context)
			},
			Info:       `Validate all TTL tables have a valid scheduled job attached.`,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.repair_ttl_table_scheduled_job": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"oid", types.Oid}},
			ReturnType: tree.FixedReturnType(types.Void),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598461)
				oid := tree.MustBeDOid(args[0])
				if err := evalCtx.Planner.RepairTTLScheduledJobForTable(evalCtx.Ctx(), int64(oid.DInt)); err != nil {
					__antithesis_instrumentation__.Notify(598463)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598464)
				}
				__antithesis_instrumentation__.Notify(598462)
				return tree.DVoidDatum, nil
			},
			Info:       `Repairs the scheduled job for a TTL table if it is missing.`,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.check_password_hash_format": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"password", types.Bytes}},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598465)
				arg := []byte(tree.MustBeDBytes(args[0]))
				ctx := evalCtx.Ctx()
				isHashed, _, _, schemeName, _, err := security.CheckPasswordHashValidity(ctx, arg)
				if err != nil {
					__antithesis_instrumentation__.Notify(598468)
					return tree.DNull, pgerror.WithCandidateCode(err, pgcode.Syntax)
				} else {
					__antithesis_instrumentation__.Notify(598469)
				}
				__antithesis_instrumentation__.Notify(598466)
				if !isHashed {
					__antithesis_instrumentation__.Notify(598470)
					return tree.DNull, pgerror.New(pgcode.Syntax, "hash format not recognized")
				} else {
					__antithesis_instrumentation__.Notify(598471)
				}
				__antithesis_instrumentation__.Notify(598467)
				return tree.NewDString(schemeName), nil
			},
			Info:       "This function checks whether a string is a precomputed password hash. Returns the hash algorithm.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"crdb_internal.schedule_sql_stats_compaction": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"session", types.Bytes}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598472)
				if evalCtx.SQLStatsController == nil {
					__antithesis_instrumentation__.Notify(598475)
					return nil, errors.AssertionFailedf("sql stats controller not set")
				} else {
					__antithesis_instrumentation__.Notify(598476)
				}
				__antithesis_instrumentation__.Notify(598473)
				ctx := evalCtx.Ctx()
				if err := evalCtx.SQLStatsController.CreateSQLStatsCompactionSchedule(ctx); err != nil {
					__antithesis_instrumentation__.Notify(598477)

					return tree.DNull, err
				} else {
					__antithesis_instrumentation__.Notify(598478)
				}
				__antithesis_instrumentation__.Notify(598474)
				return tree.DBoolTrue, nil
			},
			Info:       "This function is used to start a SQL stats compaction job.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.revalidate_unique_constraints_in_all_tables": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Void),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598479)
				if err := evalCtx.Planner.RevalidateUniqueConstraintsInCurrentDB(evalCtx.Ctx()); err != nil {
					__antithesis_instrumentation__.Notify(598481)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598482)
				}
				__antithesis_instrumentation__.Notify(598480)
				return tree.DVoidDatum, nil
			},
			Info: `This function is used to revalidate all unique constraints in tables
in the current database. Returns an error if validation fails.`,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.revalidate_unique_constraints_in_table": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"table_name", types.String}},
			ReturnType: tree.FixedReturnType(types.Void),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598483)
				name := tree.MustBeDString(args[0])
				dOid, err := tree.ParseDOid(evalCtx, string(name), types.RegClass)
				if err != nil {
					__antithesis_instrumentation__.Notify(598486)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598487)
				}
				__antithesis_instrumentation__.Notify(598484)
				if err := evalCtx.Planner.RevalidateUniqueConstraintsInTable(evalCtx.Ctx(), int(dOid.DInt)); err != nil {
					__antithesis_instrumentation__.Notify(598488)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598489)
				}
				__antithesis_instrumentation__.Notify(598485)
				return tree.DVoidDatum, nil
			},
			Info: `This function is used to revalidate all unique constraints in the given
table. Returns an error if validation fails.`,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.revalidate_unique_constraint": makeBuiltin(
		tree.FunctionProperties{
			Category: categorySystemInfo,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"table_name", types.String}, {"constraint_name", types.String}},
			ReturnType: tree.FixedReturnType(types.Void),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598490)
				tableName := tree.MustBeDString(args[0])
				constraintName := tree.MustBeDString(args[1])
				dOid, err := tree.ParseDOid(evalCtx, string(tableName), types.RegClass)
				if err != nil {
					__antithesis_instrumentation__.Notify(598493)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598494)
				}
				__antithesis_instrumentation__.Notify(598491)
				if err = evalCtx.Planner.RevalidateUniqueConstraint(
					evalCtx.Ctx(), int(dOid.DInt), string(constraintName),
				); err != nil {
					__antithesis_instrumentation__.Notify(598495)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598496)
				}
				__antithesis_instrumentation__.Notify(598492)
				return tree.DVoidDatum, nil
			},
			Info: `This function is used to revalidate the given unique constraint in the given
table. Returns an error if validation fails.`,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.kv_set_queue_active": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemRepair,
			DistsqlBlocklist: true,
			Undocumented:     true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"queue_name", types.String},
				{"active", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598497)
				isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(598501)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598502)
				}
				__antithesis_instrumentation__.Notify(598498)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(598503)
					return nil, errInsufficientPriv
				} else {
					__antithesis_instrumentation__.Notify(598504)
				}
				__antithesis_instrumentation__.Notify(598499)

				queue := string(tree.MustBeDString(args[0]))
				active := bool(tree.MustBeDBool(args[1]))

				if err := ctx.KVStoresIterator.ForEachStore(func(store kvserverbase.Store) error {
					__antithesis_instrumentation__.Notify(598505)
					return store.SetQueueActive(active, queue)
				}); err != nil {
					__antithesis_instrumentation__.Notify(598506)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598507)
				}
				__antithesis_instrumentation__.Notify(598500)

				return tree.DBoolTrue, nil
			},
			Info: `Used to enable/disable the named queue on all stores on the node it's run from.
One of 'mvccGC', 'merge', 'split', 'replicate', 'replicaGC', 'raftlog',
'raftsnapshot', 'consistencyChecker', and 'timeSeriesMaintenance'.`,
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"queue_name", types.String},
				{"active", types.Bool},
				{"store_id", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598508)
				isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(598513)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598514)
				}
				__antithesis_instrumentation__.Notify(598509)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(598515)
					return nil, errInsufficientPriv
				} else {
					__antithesis_instrumentation__.Notify(598516)
				}
				__antithesis_instrumentation__.Notify(598510)

				queue := string(tree.MustBeDString(args[0]))
				active := bool(tree.MustBeDBool(args[1]))
				storeID := roachpb.StoreID(tree.MustBeDInt(args[2]))

				var foundStore bool
				if err := ctx.KVStoresIterator.ForEachStore(func(store kvserverbase.Store) error {
					__antithesis_instrumentation__.Notify(598517)
					if storeID == store.StoreID() {
						__antithesis_instrumentation__.Notify(598519)
						foundStore = true
						return store.SetQueueActive(active, queue)
					} else {
						__antithesis_instrumentation__.Notify(598520)
					}
					__antithesis_instrumentation__.Notify(598518)
					return nil
				}); err != nil {
					__antithesis_instrumentation__.Notify(598521)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598522)
				}
				__antithesis_instrumentation__.Notify(598511)

				if !foundStore {
					__antithesis_instrumentation__.Notify(598523)
					return nil, errors.Errorf("store %s not found on this node", storeID)
				} else {
					__antithesis_instrumentation__.Notify(598524)
				}
				__antithesis_instrumentation__.Notify(598512)
				return tree.DBoolTrue, nil
			},
			Info: `Used to enable/disable the named queue on the specified store on the node it's
run from. One of 'mvccGC', 'merge', 'split', 'replicate', 'replicaGC',
'raftlog', 'raftsnapshot', 'consistencyChecker', and 'timeSeriesMaintenance'.`,
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.kv_enqueue_replica": makeBuiltin(
		tree.FunctionProperties{
			Category:         categorySystemRepair,
			DistsqlBlocklist: true,
			Undocumented:     true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"range_id", types.Int},
				{"queue_name", types.String},
				{"skip_should_queue", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598525)
				isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(598530)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598531)
				}
				__antithesis_instrumentation__.Notify(598526)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(598532)
					return nil, errInsufficientPriv
				} else {
					__antithesis_instrumentation__.Notify(598533)
				}
				__antithesis_instrumentation__.Notify(598527)

				rangeID := roachpb.RangeID(tree.MustBeDInt(args[0]))
				queue := string(tree.MustBeDString(args[1]))
				skipShouldQueue := bool(tree.MustBeDBool(args[2]))

				var foundRepl bool
				if err := ctx.KVStoresIterator.ForEachStore(func(store kvserverbase.Store) error {
					__antithesis_instrumentation__.Notify(598534)
					err := store.Enqueue(ctx.Context, queue, rangeID, skipShouldQueue)
					if err == nil {
						__antithesis_instrumentation__.Notify(598537)
						foundRepl = true
						return nil
					} else {
						__antithesis_instrumentation__.Notify(598538)
					}
					__antithesis_instrumentation__.Notify(598535)

					if errors.HasType(err, (*roachpb.RangeNotFoundError)(nil)) {
						__antithesis_instrumentation__.Notify(598539)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(598540)
					}
					__antithesis_instrumentation__.Notify(598536)
					return err
				}); err != nil {
					__antithesis_instrumentation__.Notify(598541)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598542)
				}
				__antithesis_instrumentation__.Notify(598528)

				if !foundRepl {
					__antithesis_instrumentation__.Notify(598543)
					return nil, errors.Errorf("replica with range id %s not found on this node", rangeID)
				} else {
					__antithesis_instrumentation__.Notify(598544)
				}
				__antithesis_instrumentation__.Notify(598529)
				return tree.DBoolTrue, nil
			},
			Info: `Enqueue the replica with the given range ID into the named queue, on the
specified store on the node it's run from. One of 'mvccGC', 'merge', 'split',
'replicate', 'replicaGC', 'raftlog', 'raftsnapshot', 'consistencyChecker', and
'timeSeriesMaintenance'.`,
			Volatility: tree.VolatilityVolatile,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"range_id", types.Int},
				{"queue_name", types.String},
				{"skip_should_queue", types.Bool},
				{"store_id", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598545)
				isAdmin, err := ctx.SessionAccessor.HasAdminRole(ctx.Context)
				if err != nil {
					__antithesis_instrumentation__.Notify(598550)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598551)
				}
				__antithesis_instrumentation__.Notify(598546)
				if !isAdmin {
					__antithesis_instrumentation__.Notify(598552)
					return nil, errInsufficientPriv
				} else {
					__antithesis_instrumentation__.Notify(598553)
				}
				__antithesis_instrumentation__.Notify(598547)

				rangeID := roachpb.RangeID(tree.MustBeDInt(args[0]))
				queue := string(tree.MustBeDString(args[1]))
				skipShouldQueue := bool(tree.MustBeDBool(args[2]))
				storeID := roachpb.StoreID(tree.MustBeDInt(args[3]))

				var foundStore bool
				if err := ctx.KVStoresIterator.ForEachStore(func(store kvserverbase.Store) error {
					__antithesis_instrumentation__.Notify(598554)
					if storeID == store.StoreID() {
						__antithesis_instrumentation__.Notify(598556)
						foundStore = true
						return store.Enqueue(ctx.Context, queue, rangeID, skipShouldQueue)
					} else {
						__antithesis_instrumentation__.Notify(598557)
					}
					__antithesis_instrumentation__.Notify(598555)
					return nil
				}); err != nil {
					__antithesis_instrumentation__.Notify(598558)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598559)
				}
				__antithesis_instrumentation__.Notify(598548)

				if !foundStore {
					__antithesis_instrumentation__.Notify(598560)
					return nil, errors.Errorf("store %s not found on this node", storeID)
				} else {
					__antithesis_instrumentation__.Notify(598561)
				}
				__antithesis_instrumentation__.Notify(598549)
				return tree.DBoolTrue, nil
			},
			Info: `Enqueue the replica with the given range ID into the named queue, on the
specified store on the node it's run from. One of 'mvccGC', 'merge', 'split',
'replicate', 'replicaGC', 'raftlog', 'raftsnapshot', 'consistencyChecker', and
'timeSeriesMaintenance'.`,
			Volatility: tree.VolatilityVolatile,
		},
	),
}

var lengthImpls = func(incBitOverload bool) builtinDefinition {
	__antithesis_instrumentation__.Notify(598562)
	b := makeBuiltin(tree.FunctionProperties{Category: categoryString},
		stringOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598565)
				return tree.NewDInt(tree.DInt(utf8.RuneCountInString(s))), nil
			},
			types.Int,
			"Calculates the number of characters in `val`.",
			tree.VolatilityImmutable,
		),
		bytesOverload1(
			func(_ *tree.EvalContext, s string) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598566)
				return tree.NewDInt(tree.DInt(len(s))), nil
			},
			types.Int,
			"Calculates the number of bytes in `val`.",
			tree.VolatilityImmutable,
		),
	)
	__antithesis_instrumentation__.Notify(598563)
	if incBitOverload {
		__antithesis_instrumentation__.Notify(598567)
		b.overloads = append(
			b.overloads,
			bitsOverload1(
				func(_ *tree.EvalContext, s *tree.DBitArray) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(598568)
					return tree.NewDInt(tree.DInt(s.BitArray.BitLen())), nil
				}, types.Int, "Calculates the number of bits in `val`.",
				tree.VolatilityImmutable,
			),
		)
	} else {
		__antithesis_instrumentation__.Notify(598569)
	}
	__antithesis_instrumentation__.Notify(598564)
	return b
}

var substringImpls = makeBuiltin(tree.FunctionProperties{Category: categoryString},
	tree.Overload{
		Types: tree.ArgTypes{
			{"input", types.String},
			{"start_pos", types.Int},
		},
		ReturnType: tree.FixedReturnType(types.String),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598570)
			substring := getSubstringFromIndex(string(tree.MustBeDString(args[0])), int(tree.MustBeDInt(args[1])))
			return tree.NewDString(substring), nil
		},
		Info:       "Returns a substring of `input` starting at `start_pos` (count starts at 1).",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types: tree.ArgTypes{
			{"input", types.String},
			{"start_pos", types.Int},
			{"length", types.Int},
		},
		SpecializedVecBuiltin: tree.SubstringStringIntInt,
		ReturnType:            tree.FixedReturnType(types.String),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598571)
			str := string(tree.MustBeDString(args[0]))
			start := int(tree.MustBeDInt(args[1]))
			length := int(tree.MustBeDInt(args[2]))

			substring, err := getSubstringFromIndexOfLength(str, "substring", start, length)
			if err != nil {
				__antithesis_instrumentation__.Notify(598573)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(598574)
			}
			__antithesis_instrumentation__.Notify(598572)
			return tree.NewDString(substring), nil
		},
		Info: "Returns a substring of `input` starting at `start_pos` (count starts at 1) and " +
			"including up to `length` characters.",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types: tree.ArgTypes{
			{"input", types.String},
			{"regex", types.String},
		},
		ReturnType: tree.FixedReturnType(types.String),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598575)
			s := string(tree.MustBeDString(args[0]))
			pattern := string(tree.MustBeDString(args[1]))
			return regexpExtract(ctx, s, pattern, `\`)
		},
		Info:       "Returns a substring of `input` that matches the regular expression `regex`.",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types: tree.ArgTypes{
			{"input", types.String},
			{"regex", types.String},
			{"escape_char", types.String},
		},
		ReturnType: tree.FixedReturnType(types.String),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598576)
			s := string(tree.MustBeDString(args[0]))
			pattern := string(tree.MustBeDString(args[1]))
			escape := string(tree.MustBeDString(args[2]))
			return regexpExtract(ctx, s, pattern, escape)
		},
		Info: "Returns a substring of `input` that matches the regular expression `regex` using " +
			"`escape_char` as your escape character instead of `\\`.",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types: tree.ArgTypes{
			{"input", types.VarBit},
			{"start_pos", types.Int},
		},
		ReturnType: tree.FixedReturnType(types.VarBit),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598577)
			bitString := tree.MustBeDBitArray(args[0])
			start := int(tree.MustBeDInt(args[1]))
			substring := getSubstringFromIndex(bitString.BitArray.String(), start)
			return tree.ParseDBitArray(substring)
		},
		Info:       "Returns a bit subarray of `input` starting at `start_pos` (count starts at 1).",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types: tree.ArgTypes{
			{"input", types.VarBit},
			{"start_pos", types.Int},
			{"length", types.Int},
		},
		ReturnType: tree.FixedReturnType(types.VarBit),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598578)
			bitString := tree.MustBeDBitArray(args[0])
			start := int(tree.MustBeDInt(args[1]))
			length := int(tree.MustBeDInt(args[2]))

			substring, err := getSubstringFromIndexOfLength(bitString.BitArray.String(), "bit subarray", start, length)
			if err != nil {
				__antithesis_instrumentation__.Notify(598580)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(598581)
			}
			__antithesis_instrumentation__.Notify(598579)
			return tree.ParseDBitArray(substring)
		},
		Info: "Returns a bit subarray of `input` starting at `start_pos` (count starts at 1) and " +
			"including up to `length` characters.",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types: tree.ArgTypes{
			{"input", types.Bytes},
			{"start_pos", types.Int},
		},
		ReturnType: tree.FixedReturnType(types.Bytes),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598582)
			byteString := string(*args[0].(*tree.DBytes))
			start := int(tree.MustBeDInt(args[1]))
			substring := getSubstringFromIndexBytes(byteString, start)
			return tree.NewDBytes(tree.DBytes(substring)), nil
		},
		Info:       "Returns a byte subarray of `input` starting at `start_pos` (count starts at 1).",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types: tree.ArgTypes{
			{"input", types.Bytes},
			{"start_pos", types.Int},
			{"length", types.Int},
		},
		ReturnType: tree.FixedReturnType(types.Bytes),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598583)
			byteString := string(*args[0].(*tree.DBytes))
			start := int(tree.MustBeDInt(args[1]))
			length := int(tree.MustBeDInt(args[2]))

			substring, err := getSubstringFromIndexOfLengthBytes(byteString, "byte subarray", start, length)
			if err != nil {
				__antithesis_instrumentation__.Notify(598585)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(598586)
			}
			__antithesis_instrumentation__.Notify(598584)
			return tree.NewDBytes(tree.DBytes(substring)), nil
		},
		Info: "Returns a byte subarray of `input` starting at `start_pos` (count starts at 1) and " +
			"including up to `length` characters.",
		Volatility: tree.VolatilityImmutable,
	},
)

func getSubstringFromIndex(str string, start int) string {
	__antithesis_instrumentation__.Notify(598587)
	runes := []rune(str)

	start--

	if start < 0 {
		__antithesis_instrumentation__.Notify(598589)
		start = 0
	} else {
		__antithesis_instrumentation__.Notify(598590)
		if start > len(runes) {
			__antithesis_instrumentation__.Notify(598591)
			start = len(runes)
		} else {
			__antithesis_instrumentation__.Notify(598592)
		}
	}
	__antithesis_instrumentation__.Notify(598588)
	return string(runes[start:])
}

func getSubstringFromIndexOfLength(str, errMsg string, start, length int) (string, error) {
	__antithesis_instrumentation__.Notify(598593)
	runes := []rune(str)

	start--

	if length < 0 {
		__antithesis_instrumentation__.Notify(598597)
		return "", pgerror.Newf(
			pgcode.InvalidParameterValue, "negative %s length %d not allowed", errMsg, length)
	} else {
		__antithesis_instrumentation__.Notify(598598)
	}
	__antithesis_instrumentation__.Notify(598594)

	end := start + length

	if end < start {
		__antithesis_instrumentation__.Notify(598599)
		end = len(runes)
	} else {
		__antithesis_instrumentation__.Notify(598600)
		if end < 0 {
			__antithesis_instrumentation__.Notify(598601)
			end = 0
		} else {
			__antithesis_instrumentation__.Notify(598602)
			if end > len(runes) {
				__antithesis_instrumentation__.Notify(598603)
				end = len(runes)
			} else {
				__antithesis_instrumentation__.Notify(598604)
			}
		}
	}
	__antithesis_instrumentation__.Notify(598595)

	if start < 0 {
		__antithesis_instrumentation__.Notify(598605)
		start = 0
	} else {
		__antithesis_instrumentation__.Notify(598606)
		if start > len(runes) {
			__antithesis_instrumentation__.Notify(598607)
			start = len(runes)
		} else {
			__antithesis_instrumentation__.Notify(598608)
		}
	}
	__antithesis_instrumentation__.Notify(598596)
	return string(runes[start:end]), nil
}

func getSubstringFromIndexBytes(str string, start int) string {
	__antithesis_instrumentation__.Notify(598609)
	bytes := []byte(str)

	start--

	if start < 0 {
		__antithesis_instrumentation__.Notify(598611)
		start = 0
	} else {
		__antithesis_instrumentation__.Notify(598612)
		if start > len(bytes) {
			__antithesis_instrumentation__.Notify(598613)
			start = len(bytes)
		} else {
			__antithesis_instrumentation__.Notify(598614)
		}
	}
	__antithesis_instrumentation__.Notify(598610)
	return string(bytes[start:])
}

func getSubstringFromIndexOfLengthBytes(str, errMsg string, start, length int) (string, error) {
	__antithesis_instrumentation__.Notify(598615)
	bytes := []byte(str)

	start--

	if length < 0 {
		__antithesis_instrumentation__.Notify(598619)
		return "", pgerror.Newf(
			pgcode.InvalidParameterValue, "negative %s length %d not allowed", errMsg, length)
	} else {
		__antithesis_instrumentation__.Notify(598620)
	}
	__antithesis_instrumentation__.Notify(598616)

	end := start + length

	if end < start {
		__antithesis_instrumentation__.Notify(598621)
		end = len(bytes)
	} else {
		__antithesis_instrumentation__.Notify(598622)
		if end < 0 {
			__antithesis_instrumentation__.Notify(598623)
			end = 0
		} else {
			__antithesis_instrumentation__.Notify(598624)
			if end > len(bytes) {
				__antithesis_instrumentation__.Notify(598625)
				end = len(bytes)
			} else {
				__antithesis_instrumentation__.Notify(598626)
			}
		}
	}
	__antithesis_instrumentation__.Notify(598617)

	if start < 0 {
		__antithesis_instrumentation__.Notify(598627)
		start = 0
	} else {
		__antithesis_instrumentation__.Notify(598628)
		if start > len(bytes) {
			__antithesis_instrumentation__.Notify(598629)
			start = len(bytes)
		} else {
			__antithesis_instrumentation__.Notify(598630)
		}
	}
	__antithesis_instrumentation__.Notify(598618)
	return string(bytes[start:end]), nil
}

var generateRandomUUIDImpl = makeBuiltin(
	tree.FunctionProperties{
		Category: categoryIDGeneration,
	},
	tree.Overload{
		Types:      tree.ArgTypes{},
		ReturnType: tree.FixedReturnType(types.Uuid),
		Fn: func(_ *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598631)
			uv := uuid.MakeV4()
			return tree.NewDUuid(tree.DUuid{UUID: uv}), nil
		},
		Info:       "Generates a random UUID and returns it as a value of UUID type.",
		Volatility: tree.VolatilityVolatile,
	},
)

var uuidV4Impl = makeBuiltin(
	tree.FunctionProperties{
		Category: categoryIDGeneration,
	},
	tree.Overload{
		Types:      tree.ArgTypes{},
		ReturnType: tree.FixedReturnType(types.Bytes),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598632)
			return tree.NewDBytes(tree.DBytes(uuid.MakeV4().GetBytes())), nil
		},
		Info:       "Returns a UUID.",
		Volatility: tree.VolatilityVolatile,
	},
)

const txnTSContextDoc = `

The value is based on a timestamp picked when the transaction starts
and which stays constant throughout the transaction. This timestamp
has no relationship with the commit order of concurrent transactions.`
const txnPreferredOverloadStr = `

This function is the preferred overload and will be evaluated by default.`
const txnTSDoc = `Returns the time of the current transaction.` + txnTSContextDoc

func getTimeAdditionalDesc(preferTZOverload bool) (string, string) {
	__antithesis_instrumentation__.Notify(598633)
	var tzAdditionalDesc, noTZAdditionalDesc string
	if preferTZOverload {
		__antithesis_instrumentation__.Notify(598635)
		tzAdditionalDesc = txnPreferredOverloadStr
	} else {
		__antithesis_instrumentation__.Notify(598636)
		noTZAdditionalDesc = txnPreferredOverloadStr
	}
	__antithesis_instrumentation__.Notify(598634)
	return tzAdditionalDesc, noTZAdditionalDesc
}

func txnTSOverloads(preferTZOverload bool) []tree.Overload {
	__antithesis_instrumentation__.Notify(598637)
	tzAdditionalDesc, noTZAdditionalDesc := getTimeAdditionalDesc(preferTZOverload)
	return []tree.Overload{
		{
			Types:             tree.ArgTypes{},
			ReturnType:        tree.FixedReturnType(types.TimestampTZ),
			PreferredOverload: preferTZOverload,
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598638)
				return ctx.GetTxnTimestamp(time.Microsecond), nil
			},
			Info:       txnTSDoc + tzAdditionalDesc,
			Volatility: tree.VolatilityStable,
		},
		{
			Types:             tree.ArgTypes{},
			ReturnType:        tree.FixedReturnType(types.Timestamp),
			PreferredOverload: !preferTZOverload,
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598639)
				return ctx.GetTxnTimestampNoZone(time.Microsecond), nil
			},
			Info:       txnTSDoc + noTZAdditionalDesc,
			Volatility: tree.VolatilityStable,
		},
		{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Date),
			Fn:         currentDate,
			Info:       txnTSDoc,
			Volatility: tree.VolatilityStable,
		},
	}
}

func txnTSWithPrecisionOverloads(preferTZOverload bool) []tree.Overload {
	__antithesis_instrumentation__.Notify(598640)
	tzAdditionalDesc, noTZAdditionalDesc := getTimeAdditionalDesc(preferTZOverload)
	return append(
		[]tree.Overload{
			{
				Types:             tree.ArgTypes{{"precision", types.Int}},
				ReturnType:        tree.FixedReturnType(types.TimestampTZ),
				PreferredOverload: preferTZOverload,
				Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(598641)
					prec := int32(tree.MustBeDInt(args[0]))
					if prec < 0 || func() bool {
						__antithesis_instrumentation__.Notify(598643)
						return prec > 6 == true
					}() == true {
						__antithesis_instrumentation__.Notify(598644)
						return nil, pgerror.Newf(pgcode.NumericValueOutOfRange, "precision %d out of range", prec)
					} else {
						__antithesis_instrumentation__.Notify(598645)
					}
					__antithesis_instrumentation__.Notify(598642)
					return ctx.GetTxnTimestamp(tree.TimeFamilyPrecisionToRoundDuration(prec)), nil
				},
				Info:       txnTSDoc + tzAdditionalDesc,
				Volatility: tree.VolatilityStable,
			},
			{
				Types:             tree.ArgTypes{{"precision", types.Int}},
				ReturnType:        tree.FixedReturnType(types.Timestamp),
				PreferredOverload: !preferTZOverload,
				Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(598646)
					prec := int32(tree.MustBeDInt(args[0]))
					if prec < 0 || func() bool {
						__antithesis_instrumentation__.Notify(598648)
						return prec > 6 == true
					}() == true {
						__antithesis_instrumentation__.Notify(598649)
						return nil, pgerror.Newf(pgcode.NumericValueOutOfRange, "precision %d out of range", prec)
					} else {
						__antithesis_instrumentation__.Notify(598650)
					}
					__antithesis_instrumentation__.Notify(598647)
					return ctx.GetTxnTimestampNoZone(tree.TimeFamilyPrecisionToRoundDuration(prec)), nil
				},
				Info:       txnTSDoc + noTZAdditionalDesc,
				Volatility: tree.VolatilityStable,
			},
			{
				Types:      tree.ArgTypes{{"precision", types.Int}},
				ReturnType: tree.FixedReturnType(types.Date),
				Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
					__antithesis_instrumentation__.Notify(598651)
					prec := int32(tree.MustBeDInt(args[0]))
					if prec < 0 || func() bool {
						__antithesis_instrumentation__.Notify(598653)
						return prec > 6 == true
					}() == true {
						__antithesis_instrumentation__.Notify(598654)
						return nil, pgerror.Newf(pgcode.NumericValueOutOfRange, "precision %d out of range", prec)
					} else {
						__antithesis_instrumentation__.Notify(598655)
					}
					__antithesis_instrumentation__.Notify(598652)
					return currentDate(ctx, args)
				},
				Info:       txnTSDoc,
				Volatility: tree.VolatilityStable,
			},
		},
		txnTSOverloads(preferTZOverload)...,
	)
}

func txnTSImplBuiltin(preferTZOverload bool) builtinDefinition {
	__antithesis_instrumentation__.Notify(598656)
	return makeBuiltin(
		tree.FunctionProperties{
			Category: categoryDateAndTime,
		},
		txnTSOverloads(preferTZOverload)...,
	)
}

func txnTSWithPrecisionImplBuiltin(preferTZOverload bool) builtinDefinition {
	__antithesis_instrumentation__.Notify(598657)
	return makeBuiltin(
		tree.FunctionProperties{
			Category: categoryDateAndTime,
		},
		txnTSWithPrecisionOverloads(preferTZOverload)...,
	)
}

func txnTimeWithPrecisionBuiltin(preferTZOverload bool) builtinDefinition {
	__antithesis_instrumentation__.Notify(598658)
	tzAdditionalDesc, noTZAdditionalDesc := getTimeAdditionalDesc(preferTZOverload)
	return makeBuiltin(
		tree.FunctionProperties{},
		tree.Overload{
			Types:             tree.ArgTypes{},
			ReturnType:        tree.FixedReturnType(types.TimeTZ),
			PreferredOverload: preferTZOverload,
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598659)
				return ctx.GetTxnTime(time.Microsecond), nil
			},
			Info:       "Returns the current transaction's time with time zone." + tzAdditionalDesc,
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:             tree.ArgTypes{},
			ReturnType:        tree.FixedReturnType(types.Time),
			PreferredOverload: !preferTZOverload,
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598660)
				return ctx.GetTxnTimeNoZone(time.Microsecond), nil
			},
			Info:       "Returns the current transaction's time with no time zone." + noTZAdditionalDesc,
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:             tree.ArgTypes{{"precision", types.Int}},
			ReturnType:        tree.FixedReturnType(types.TimeTZ),
			PreferredOverload: preferTZOverload,
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598661)
				prec := int32(tree.MustBeDInt(args[0]))
				if prec < 0 || func() bool {
					__antithesis_instrumentation__.Notify(598663)
					return prec > 6 == true
				}() == true {
					__antithesis_instrumentation__.Notify(598664)
					return nil, pgerror.Newf(pgcode.NumericValueOutOfRange, "precision %d out of range", prec)
				} else {
					__antithesis_instrumentation__.Notify(598665)
				}
				__antithesis_instrumentation__.Notify(598662)
				return ctx.GetTxnTime(tree.TimeFamilyPrecisionToRoundDuration(prec)), nil
			},
			Info:       "Returns the current transaction's time with time zone." + tzAdditionalDesc,
			Volatility: tree.VolatilityStable,
		},
		tree.Overload{
			Types:             tree.ArgTypes{{"precision", types.Int}},
			ReturnType:        tree.FixedReturnType(types.Time),
			PreferredOverload: !preferTZOverload,
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598666)
				prec := int32(tree.MustBeDInt(args[0]))
				if prec < 0 || func() bool {
					__antithesis_instrumentation__.Notify(598668)
					return prec > 6 == true
				}() == true {
					__antithesis_instrumentation__.Notify(598669)
					return nil, pgerror.Newf(pgcode.NumericValueOutOfRange, "precision %d out of range", prec)
				} else {
					__antithesis_instrumentation__.Notify(598670)
				}
				__antithesis_instrumentation__.Notify(598667)
				return ctx.GetTxnTimeNoZone(tree.TimeFamilyPrecisionToRoundDuration(prec)), nil
			},
			Info:       "Returns the current transaction's time with no time zone." + noTZAdditionalDesc,
			Volatility: tree.VolatilityStable,
		},
	)
}

func currentDate(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(598671)
	t := ctx.GetTxnTimestamp(time.Microsecond).Time
	t = t.In(ctx.GetLocation())
	return tree.NewDDateFromTime(t)
}

var (
	jsonNullDString    = tree.NewDString("null")
	jsonStringDString  = tree.NewDString("string")
	jsonNumberDString  = tree.NewDString("number")
	jsonBooleanDString = tree.NewDString("boolean")
	jsonArrayDString   = tree.NewDString("array")
	jsonObjectDString  = tree.NewDString("object")
)

var (
	errJSONObjectNotEvenNumberOfElements = pgerror.New(pgcode.InvalidParameterValue,
		"array must have even number of elements")
	errJSONObjectNullValueForKey = pgerror.New(pgcode.InvalidParameterValue,
		"null value not allowed for object key")
	errJSONObjectMismatchedArrayDim = pgerror.New(pgcode.InvalidParameterValue,
		"mismatched array dimensions")
)

var jsonExtractPathImpl = tree.Overload{
	Types:      tree.VariadicType{FixedTypes: []*types.T{types.Jsonb}, VarType: types.String},
	ReturnType: tree.FixedReturnType(types.Jsonb),
	Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598672)
		result, err := jsonExtractPathHelper(args)
		if err != nil {
			__antithesis_instrumentation__.Notify(598675)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(598676)
		}
		__antithesis_instrumentation__.Notify(598673)
		if result == nil {
			__antithesis_instrumentation__.Notify(598677)
			return tree.DNull, nil
		} else {
			__antithesis_instrumentation__.Notify(598678)
		}
		__antithesis_instrumentation__.Notify(598674)
		return &tree.DJSON{JSON: result}, nil
	},
	Info:       "Returns the JSON value pointed to by the variadic arguments.",
	Volatility: tree.VolatilityImmutable,
}

var jsonExtractPathTextImpl = tree.Overload{
	Types:      tree.VariadicType{FixedTypes: []*types.T{types.Jsonb}, VarType: types.String},
	ReturnType: tree.FixedReturnType(types.String),
	Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598679)
		result, err := jsonExtractPathHelper(args)
		if err != nil {
			__antithesis_instrumentation__.Notify(598684)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(598685)
		}
		__antithesis_instrumentation__.Notify(598680)
		if result == nil {
			__antithesis_instrumentation__.Notify(598686)
			return tree.DNull, nil
		} else {
			__antithesis_instrumentation__.Notify(598687)
		}
		__antithesis_instrumentation__.Notify(598681)
		text, err := result.AsText()
		if err != nil {
			__antithesis_instrumentation__.Notify(598688)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(598689)
		}
		__antithesis_instrumentation__.Notify(598682)
		if text == nil {
			__antithesis_instrumentation__.Notify(598690)
			return tree.DNull, nil
		} else {
			__antithesis_instrumentation__.Notify(598691)
		}
		__antithesis_instrumentation__.Notify(598683)
		return tree.NewDString(*text), nil
	},
	Info:       "Returns the JSON value as text pointed to by the variadic arguments.",
	Volatility: tree.VolatilityImmutable,
}

func jsonExtractPathHelper(args tree.Datums) (json.JSON, error) {
	__antithesis_instrumentation__.Notify(598692)
	j := tree.MustBeDJSON(args[0])
	path := make([]string, len(args)-1)
	for i, v := range args {
		__antithesis_instrumentation__.Notify(598694)
		if i == 0 {
			__antithesis_instrumentation__.Notify(598697)
			continue
		} else {
			__antithesis_instrumentation__.Notify(598698)
		}
		__antithesis_instrumentation__.Notify(598695)
		if v == tree.DNull {
			__antithesis_instrumentation__.Notify(598699)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(598700)
		}
		__antithesis_instrumentation__.Notify(598696)
		path[i-1] = string(tree.MustBeDString(v))
	}
	__antithesis_instrumentation__.Notify(598693)
	return json.FetchPath(j.JSON, path)
}

func darrayToStringSlice(d tree.DArray) (result []string, ok bool) {
	__antithesis_instrumentation__.Notify(598701)
	result = make([]string, len(d.Array))
	for i, s := range d.Array {
		__antithesis_instrumentation__.Notify(598703)
		if s == tree.DNull {
			__antithesis_instrumentation__.Notify(598705)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(598706)
		}
		__antithesis_instrumentation__.Notify(598704)
		result[i] = string(tree.MustBeDString(s))
	}
	__antithesis_instrumentation__.Notify(598702)
	return result, true
}

func checkHasNulls(ary tree.DArray) error {
	__antithesis_instrumentation__.Notify(598707)
	if ary.HasNulls {
		__antithesis_instrumentation__.Notify(598709)
		for i := range ary.Array {
			__antithesis_instrumentation__.Notify(598710)
			if ary.Array[i] == tree.DNull {
				__antithesis_instrumentation__.Notify(598711)
				return pgerror.Newf(pgcode.NullValueNotAllowed, "path element at position %d is null", i+1)
			} else {
				__antithesis_instrumentation__.Notify(598712)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(598713)
	}
	__antithesis_instrumentation__.Notify(598708)
	return nil
}

var jsonSetImpl = tree.Overload{
	Types: tree.ArgTypes{
		{"val", types.Jsonb},
		{"path", types.StringArray},
		{"to", types.Jsonb},
	},
	ReturnType: tree.FixedReturnType(types.Jsonb),
	Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598714)
		return jsonDatumSet(args[0], args[1], args[2], tree.DBoolTrue)
	},
	Info:       "Returns the JSON value pointed to by the variadic arguments.",
	Volatility: tree.VolatilityImmutable,
}

var jsonSetWithCreateMissingImpl = tree.Overload{
	Types: tree.ArgTypes{
		{"val", types.Jsonb},
		{"path", types.StringArray},
		{"to", types.Jsonb},
		{"create_missing", types.Bool},
	},
	ReturnType: tree.FixedReturnType(types.Jsonb),
	Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598715)
		return jsonDatumSet(args[0], args[1], args[2], args[3])
	},
	Info: "Returns the JSON value pointed to by the variadic arguments. " +
		"If `create_missing` is false, new keys will not be inserted to objects " +
		"and values will not be prepended or appended to arrays.",
	Volatility: tree.VolatilityImmutable,
}

func jsonDatumSet(
	targetD tree.Datum, pathD tree.Datum, toD tree.Datum, createMissingD tree.Datum,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(598716)
	ary := *tree.MustBeDArray(pathD)

	if err := checkHasNulls(ary); err != nil {
		__antithesis_instrumentation__.Notify(598720)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(598721)
	}
	__antithesis_instrumentation__.Notify(598717)
	path, ok := darrayToStringSlice(ary)
	if !ok {
		__antithesis_instrumentation__.Notify(598722)
		return targetD, nil
	} else {
		__antithesis_instrumentation__.Notify(598723)
	}
	__antithesis_instrumentation__.Notify(598718)
	j, err := json.DeepSet(tree.MustBeDJSON(targetD).JSON, path, tree.MustBeDJSON(toD).JSON, bool(tree.MustBeDBool(createMissingD)))
	if err != nil {
		__antithesis_instrumentation__.Notify(598724)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(598725)
	}
	__antithesis_instrumentation__.Notify(598719)
	return &tree.DJSON{JSON: j}, nil
}

var jsonInsertImpl = tree.Overload{
	Types: tree.ArgTypes{
		{"target", types.Jsonb},
		{"path", types.StringArray},
		{"new_val", types.Jsonb},
	},
	ReturnType: tree.FixedReturnType(types.Jsonb),
	Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598726)
		return insertToJSONDatum(args[0], args[1], args[2], tree.DBoolFalse)
	},
	Info:       "Returns the JSON value pointed to by the variadic arguments. `new_val` will be inserted before path target.",
	Volatility: tree.VolatilityImmutable,
}

var jsonInsertWithInsertAfterImpl = tree.Overload{
	Types: tree.ArgTypes{
		{"target", types.Jsonb},
		{"path", types.StringArray},
		{"new_val", types.Jsonb},
		{"insert_after", types.Bool},
	},
	ReturnType: tree.FixedReturnType(types.Jsonb),
	Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598727)
		return insertToJSONDatum(args[0], args[1], args[2], args[3])
	},
	Info: "Returns the JSON value pointed to by the variadic arguments. " +
		"If `insert_after` is true (default is false), `new_val` will be inserted after path target.",
	Volatility: tree.VolatilityImmutable,
}

func insertToJSONDatum(
	targetD tree.Datum, pathD tree.Datum, newValD tree.Datum, insertAfterD tree.Datum,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(598728)
	ary := *tree.MustBeDArray(pathD)

	if err := checkHasNulls(ary); err != nil {
		__antithesis_instrumentation__.Notify(598732)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(598733)
	}
	__antithesis_instrumentation__.Notify(598729)
	path, ok := darrayToStringSlice(ary)
	if !ok {
		__antithesis_instrumentation__.Notify(598734)
		return targetD, nil
	} else {
		__antithesis_instrumentation__.Notify(598735)
	}
	__antithesis_instrumentation__.Notify(598730)
	j, err := json.DeepInsert(tree.MustBeDJSON(targetD).JSON, path, tree.MustBeDJSON(newValD).JSON, bool(tree.MustBeDBool(insertAfterD)))
	if err != nil {
		__antithesis_instrumentation__.Notify(598736)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(598737)
	}
	__antithesis_instrumentation__.Notify(598731)
	return &tree.DJSON{JSON: j}, nil
}

var jsonTypeOfImpl = tree.Overload{
	Types:      tree.ArgTypes{{"val", types.Jsonb}},
	ReturnType: tree.FixedReturnType(types.String),
	Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598738)
		t := tree.MustBeDJSON(args[0]).JSON.Type()
		switch t {
		case json.NullJSONType:
			__antithesis_instrumentation__.Notify(598740)
			return jsonNullDString, nil
		case json.StringJSONType:
			__antithesis_instrumentation__.Notify(598741)
			return jsonStringDString, nil
		case json.NumberJSONType:
			__antithesis_instrumentation__.Notify(598742)
			return jsonNumberDString, nil
		case json.FalseJSONType, json.TrueJSONType:
			__antithesis_instrumentation__.Notify(598743)
			return jsonBooleanDString, nil
		case json.ArrayJSONType:
			__antithesis_instrumentation__.Notify(598744)
			return jsonArrayDString, nil
		case json.ObjectJSONType:
			__antithesis_instrumentation__.Notify(598745)
			return jsonObjectDString, nil
		default:
			__antithesis_instrumentation__.Notify(598746)
		}
		__antithesis_instrumentation__.Notify(598739)
		return nil, errors.AssertionFailedf("unexpected JSON type %d", t)
	},
	Info:       "Returns the type of the outermost JSON value as a text string.",
	Volatility: tree.VolatilityImmutable,
}

func jsonProps() tree.FunctionProperties {
	__antithesis_instrumentation__.Notify(598747)
	return tree.FunctionProperties{
		Category: categoryJSON,
	}
}

func jsonPropsNullableArgs() tree.FunctionProperties {
	__antithesis_instrumentation__.Notify(598748)
	d := jsonProps()
	d.NullableArgs = true
	return d
}

var jsonBuildObjectImpl = tree.Overload{
	Types:      tree.VariadicType{VarType: types.Any},
	ReturnType: tree.FixedReturnType(types.Jsonb),
	Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598749)
		if len(args)%2 != 0 {
			__antithesis_instrumentation__.Notify(598752)
			return nil, pgerror.New(pgcode.InvalidParameterValue,
				"argument list must have even number of elements")
		} else {
			__antithesis_instrumentation__.Notify(598753)
		}
		__antithesis_instrumentation__.Notify(598750)

		builder := json.NewObjectBuilder(len(args) / 2)
		for i := 0; i < len(args); i += 2 {
			__antithesis_instrumentation__.Notify(598754)
			if args[i] == tree.DNull {
				__antithesis_instrumentation__.Notify(598758)
				return nil, pgerror.Newf(pgcode.InvalidParameterValue,
					"argument %d cannot be null", i+1)
			} else {
				__antithesis_instrumentation__.Notify(598759)
			}
			__antithesis_instrumentation__.Notify(598755)

			key, err := asJSONBuildObjectKey(
				args[i],
				ctx.SessionData().DataConversionConfig,
				ctx.GetLocation(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(598760)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(598761)
			}
			__antithesis_instrumentation__.Notify(598756)

			val, err := tree.AsJSON(
				args[i+1],
				ctx.SessionData().DataConversionConfig,
				ctx.GetLocation(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(598762)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(598763)
			}
			__antithesis_instrumentation__.Notify(598757)

			builder.Add(key, val)
		}
		__antithesis_instrumentation__.Notify(598751)

		return tree.NewDJSON(builder.Build()), nil
	},
	Info:       "Builds a JSON object out of a variadic argument list.",
	Volatility: tree.VolatilityStable,
}

var toJSONImpl = tree.Overload{
	Types:      tree.ArgTypes{{"val", types.Any}},
	ReturnType: tree.FixedReturnType(types.Jsonb),
	Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598764)
		return toJSONObject(ctx, args[0])
	},
	Info:       "Returns the value as JSON or JSONB.",
	Volatility: tree.VolatilityStable,
}

var prettyPrintNotSupportedError = pgerror.Newf(pgcode.FeatureNotSupported, "pretty printing is not supported")

var arrayToJSONImpls = makeBuiltin(jsonProps(),
	tree.Overload{
		Types:      tree.ArgTypes{{"array", types.AnyArray}},
		ReturnType: tree.FixedReturnType(types.Jsonb),
		Fn:         toJSONImpl.Fn,
		Info:       "Returns the array as JSON or JSONB.",
		Volatility: tree.VolatilityStable,
	},
	tree.Overload{
		Types:      tree.ArgTypes{{"array", types.AnyArray}, {"pretty_bool", types.Bool}},
		ReturnType: tree.FixedReturnType(types.Jsonb),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598765)
			prettyPrint := bool(tree.MustBeDBool(args[1]))
			if prettyPrint {
				__antithesis_instrumentation__.Notify(598767)
				return nil, prettyPrintNotSupportedError
			} else {
				__antithesis_instrumentation__.Notify(598768)
			}
			__antithesis_instrumentation__.Notify(598766)
			return toJSONObject(ctx, args[0])
		},
		Info:       "Returns the array as JSON or JSONB.",
		Volatility: tree.VolatilityStable,
	},
)

var jsonBuildArrayImpl = tree.Overload{
	Types:      tree.VariadicType{VarType: types.Any},
	ReturnType: tree.FixedReturnType(types.Jsonb),
	Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598769)
		builder := json.NewArrayBuilder(len(args))
		for _, arg := range args {
			__antithesis_instrumentation__.Notify(598771)
			j, err := tree.AsJSON(
				arg,
				ctx.SessionData().DataConversionConfig,
				ctx.GetLocation(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(598773)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(598774)
			}
			__antithesis_instrumentation__.Notify(598772)
			builder.Add(j)
		}
		__antithesis_instrumentation__.Notify(598770)
		return tree.NewDJSON(builder.Build()), nil
	},
	Info:       "Builds a possibly-heterogeneously-typed JSON or JSONB array out of a variadic argument list.",
	Volatility: tree.VolatilityStable,
}

var jsonObjectImpls = makeBuiltin(jsonProps(),
	tree.Overload{
		Types:      tree.ArgTypes{{"texts", types.StringArray}},
		ReturnType: tree.FixedReturnType(types.Jsonb),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598775)
			arr := tree.MustBeDArray(args[0])
			if arr.Len()%2 != 0 {
				__antithesis_instrumentation__.Notify(598778)
				return nil, errJSONObjectNotEvenNumberOfElements
			} else {
				__antithesis_instrumentation__.Notify(598779)
			}
			__antithesis_instrumentation__.Notify(598776)
			builder := json.NewObjectBuilder(arr.Len() / 2)
			for i := 0; i < arr.Len(); i += 2 {
				__antithesis_instrumentation__.Notify(598780)
				if arr.Array[i] == tree.DNull {
					__antithesis_instrumentation__.Notify(598784)
					return nil, errJSONObjectNullValueForKey
				} else {
					__antithesis_instrumentation__.Notify(598785)
				}
				__antithesis_instrumentation__.Notify(598781)
				key, err := asJSONObjectKey(arr.Array[i])
				if err != nil {
					__antithesis_instrumentation__.Notify(598786)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598787)
				}
				__antithesis_instrumentation__.Notify(598782)
				val, err := tree.AsJSON(
					arr.Array[i+1],
					ctx.SessionData().DataConversionConfig,
					ctx.GetLocation(),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(598788)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598789)
				}
				__antithesis_instrumentation__.Notify(598783)
				builder.Add(key, val)
			}
			__antithesis_instrumentation__.Notify(598777)
			return tree.NewDJSON(builder.Build()), nil
		},
		Info: "Builds a JSON or JSONB object out of a text array. The array must have " +
			"exactly one dimension with an even number of members, in which case " +
			"they are taken as alternating key/value pairs.",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types: tree.ArgTypes{{"keys", types.StringArray},
			{"values", types.StringArray}},
		ReturnType: tree.FixedReturnType(types.Jsonb),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598790)
			keys := tree.MustBeDArray(args[0])
			values := tree.MustBeDArray(args[1])
			if keys.Len() != values.Len() {
				__antithesis_instrumentation__.Notify(598793)
				return nil, errJSONObjectMismatchedArrayDim
			} else {
				__antithesis_instrumentation__.Notify(598794)
			}
			__antithesis_instrumentation__.Notify(598791)
			builder := json.NewObjectBuilder(keys.Len())
			for i := 0; i < keys.Len(); i++ {
				__antithesis_instrumentation__.Notify(598795)
				if keys.Array[i] == tree.DNull {
					__antithesis_instrumentation__.Notify(598799)
					return nil, errJSONObjectNullValueForKey
				} else {
					__antithesis_instrumentation__.Notify(598800)
				}
				__antithesis_instrumentation__.Notify(598796)
				key, err := asJSONObjectKey(keys.Array[i])
				if err != nil {
					__antithesis_instrumentation__.Notify(598801)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598802)
				}
				__antithesis_instrumentation__.Notify(598797)
				val, err := tree.AsJSON(
					values.Array[i],
					ctx.SessionData().DataConversionConfig,
					ctx.GetLocation(),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(598803)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(598804)
				}
				__antithesis_instrumentation__.Notify(598798)
				builder.Add(key, val)
			}
			__antithesis_instrumentation__.Notify(598792)
			return tree.NewDJSON(builder.Build()), nil
		},
		Info: "This form of json_object takes keys and values pairwise from two " +
			"separate arrays. In all other respects it is identical to the " +
			"one-argument form.",
		Volatility: tree.VolatilityImmutable,
	},
)

var jsonStripNullsImpl = tree.Overload{
	Types:      tree.ArgTypes{{"from_json", types.Jsonb}},
	ReturnType: tree.FixedReturnType(types.Jsonb),
	Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598805)
		j, _, err := tree.MustBeDJSON(args[0]).StripNulls()
		return tree.NewDJSON(j), err
	},
	Info:       "Returns from_json with all object fields that have null values omitted. Other null values are untouched.",
	Volatility: tree.VolatilityImmutable,
}

var jsonArrayLengthImpl = tree.Overload{
	Types:      tree.ArgTypes{{"json", types.Jsonb}},
	ReturnType: tree.FixedReturnType(types.Int),
	Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(598806)
		j := tree.MustBeDJSON(args[0])
		switch j.Type() {
		case json.ArrayJSONType:
			__antithesis_instrumentation__.Notify(598807)
			return tree.NewDInt(tree.DInt(j.Len())), nil
		case json.ObjectJSONType:
			__antithesis_instrumentation__.Notify(598808)
			return nil, pgerror.New(pgcode.InvalidParameterValue,
				"cannot get array length of a non-array")
		default:
			__antithesis_instrumentation__.Notify(598809)
			return nil, pgerror.New(pgcode.InvalidParameterValue,
				"cannot get array length of a scalar")
		}
	},
	Info:       "Returns the number of elements in the outermost JSON or JSONB array.",
	Volatility: tree.VolatilityImmutable,
}

var similarOverloads = []tree.Overload{
	{
		Types:      tree.ArgTypes{{"pattern", types.String}},
		ReturnType: tree.FixedReturnType(types.String),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598810)
			if args[0] == tree.DNull {
				__antithesis_instrumentation__.Notify(598812)
				return tree.DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(598813)
			}
			__antithesis_instrumentation__.Notify(598811)
			pattern := string(tree.MustBeDString(args[0]))
			return tree.SimilarPattern(pattern, "")
		},
		Info:       "Converts a SQL regexp `pattern` to a POSIX regexp `pattern`.",
		Volatility: tree.VolatilityImmutable,
	},
	{
		Types:      tree.ArgTypes{{"pattern", types.String}, {"escape", types.String}},
		ReturnType: tree.FixedReturnType(types.String),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598814)
			if args[0] == tree.DNull {
				__antithesis_instrumentation__.Notify(598817)
				return tree.DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(598818)
			}
			__antithesis_instrumentation__.Notify(598815)
			pattern := string(tree.MustBeDString(args[0]))
			if args[1] == tree.DNull {
				__antithesis_instrumentation__.Notify(598819)
				return tree.SimilarPattern(pattern, "")
			} else {
				__antithesis_instrumentation__.Notify(598820)
			}
			__antithesis_instrumentation__.Notify(598816)
			escape := string(tree.MustBeDString(args[1]))
			return tree.SimilarPattern(pattern, escape)
		},
		Info:       "Converts a SQL regexp `pattern` to a POSIX regexp `pattern` using `escape` as an escape token.",
		Volatility: tree.VolatilityImmutable,
	},
}

func arrayBuiltin(impl func(*types.T) tree.Overload) builtinDefinition {
	__antithesis_instrumentation__.Notify(598821)
	overloads := make([]tree.Overload, 0, len(types.Scalar)+2)
	for _, typ := range append(types.Scalar, types.AnyEnum) {
		__antithesis_instrumentation__.Notify(598823)
		if ok, _ := types.IsValidArrayElementType(typ); ok {
			__antithesis_instrumentation__.Notify(598824)
			overloads = append(overloads, impl(typ))
		} else {
			__antithesis_instrumentation__.Notify(598825)
		}
	}
	__antithesis_instrumentation__.Notify(598822)

	tupleOverload := impl(types.AnyTuple)
	tupleOverload.DistsqlBlocklist = true
	overloads = append(overloads, tupleOverload)
	return builtinDefinition{
		props:     tree.FunctionProperties{Category: categoryArray},
		overloads: overloads,
	}
}

func setProps(props tree.FunctionProperties, d builtinDefinition) builtinDefinition {
	__antithesis_instrumentation__.Notify(598826)
	d.props = props
	return d
}

func jsonOverload1(
	f func(*tree.EvalContext, json.JSON) (tree.Datum, error),
	returnType *types.T,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(598827)
	return tree.Overload{
		Types:      tree.ArgTypes{{"val", types.Jsonb}},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598828)
			return f(evalCtx, tree.MustBeDJSON(args[0]).JSON)
		},
		Info:       info,
		Volatility: volatility,
	}
}

func stringOverload1(
	f func(*tree.EvalContext, string) (tree.Datum, error),
	returnType *types.T,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(598829)
	return tree.Overload{
		Types:      tree.ArgTypes{{"val", types.String}},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598830)
			return f(evalCtx, string(tree.MustBeDString(args[0])))
		},
		Info:       info,
		Volatility: volatility,
	}
}

func stringOverload2(
	a, b string,
	f func(*tree.EvalContext, string, string) (tree.Datum, error),
	returnType *types.T,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(598831)
	return tree.Overload{
		Types:      tree.ArgTypes{{a, types.String}, {b, types.String}},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598832)
			return f(evalCtx, string(tree.MustBeDString(args[0])), string(tree.MustBeDString(args[1])))
		},
		Info:       info,
		Volatility: volatility,
	}
}

func stringOverload3(
	a, b, c string,
	f func(*tree.EvalContext, string, string, string) (tree.Datum, error),
	returnType *types.T,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(598833)
	return tree.Overload{
		Types:      tree.ArgTypes{{a, types.String}, {b, types.String}, {c, types.String}},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598834)
			return f(evalCtx, string(tree.MustBeDString(args[0])), string(tree.MustBeDString(args[1])), string(tree.MustBeDString(args[2])))
		},
		Info:       info,
		Volatility: volatility,
	}
}

func bytesOverload1(
	f func(*tree.EvalContext, string) (tree.Datum, error),
	returnType *types.T,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(598835)
	return tree.Overload{
		Types:      tree.ArgTypes{{"val", types.Bytes}},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598836)
			return f(evalCtx, string(*args[0].(*tree.DBytes)))
		},
		Info:       info,
		Volatility: volatility,
	}
}

func bytesOverload2(
	a, b string,
	f func(*tree.EvalContext, string, string) (tree.Datum, error),
	returnType *types.T,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(598837)
	return tree.Overload{
		Types:      tree.ArgTypes{{a, types.Bytes}, {b, types.Bytes}},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598838)
			return f(evalCtx, string(*args[0].(*tree.DBytes)), string(*args[1].(*tree.DBytes)))
		},
		Info:       info,
		Volatility: volatility,
	}
}

func bitsOverload1(
	f func(*tree.EvalContext, *tree.DBitArray) (tree.Datum, error),
	returnType *types.T,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(598839)
	return tree.Overload{
		Types:      tree.ArgTypes{{"val", types.VarBit}},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598840)
			return f(evalCtx, args[0].(*tree.DBitArray))
		},
		Info:       info,
		Volatility: volatility,
	}
}

func bitsOverload2(
	a, b string,
	f func(*tree.EvalContext, *tree.DBitArray, *tree.DBitArray) (tree.Datum, error),
	returnType *types.T,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(598841)
	return tree.Overload{
		Types:      tree.ArgTypes{{a, types.VarBit}, {b, types.VarBit}},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(598842)
			return f(evalCtx, args[0].(*tree.DBitArray), args[1].(*tree.DBitArray))
		},
		Info:       info,
		Volatility: volatility,
	}
}

func getHashFunc(alg string) (func() hash.Hash, error) {
	__antithesis_instrumentation__.Notify(598843)
	switch strings.ToLower(alg) {
	case "md5":
		__antithesis_instrumentation__.Notify(598844)
		return md5.New, nil
	case "sha1":
		__antithesis_instrumentation__.Notify(598845)
		return sha1.New, nil
	case "sha224":
		__antithesis_instrumentation__.Notify(598846)
		return sha256.New224, nil
	case "sha256":
		__antithesis_instrumentation__.Notify(598847)
		return sha256.New, nil
	case "sha384":
		__antithesis_instrumentation__.Notify(598848)
		return sha512.New384, nil
	case "sha512":
		__antithesis_instrumentation__.Notify(598849)
		return sha512.New, nil
	default:
		__antithesis_instrumentation__.Notify(598850)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "cannot use %q, no such hash algorithm", alg)
	}
}

func feedHash(h hash.Hash, args tree.Datums) (bool, error) {
	__antithesis_instrumentation__.Notify(598851)
	var nonNullSeen bool
	for _, datum := range args {
		__antithesis_instrumentation__.Notify(598853)
		if datum == tree.DNull {
			__antithesis_instrumentation__.Notify(598856)
			continue
		} else {
			__antithesis_instrumentation__.Notify(598857)
			nonNullSeen = true
		}
		__antithesis_instrumentation__.Notify(598854)
		var buf string
		if d, ok := datum.(*tree.DBytes); ok {
			__antithesis_instrumentation__.Notify(598858)
			buf = string(*d)
		} else {
			__antithesis_instrumentation__.Notify(598859)
			buf = string(tree.MustBeDString(datum))
		}
		__antithesis_instrumentation__.Notify(598855)
		_, err := h.Write([]byte(buf))
		if err != nil {
			__antithesis_instrumentation__.Notify(598860)
			return false, errors.NewAssertionErrorWithWrappedErrf(err,
				`"It never returns an error." -- https://golang.org/pkg/hash: %T`, h)
		} else {
			__antithesis_instrumentation__.Notify(598861)
		}
	}
	__antithesis_instrumentation__.Notify(598852)
	return nonNullSeen, nil
}

func hashBuiltin(newHash func() hash.Hash, info string) builtinDefinition {
	__antithesis_instrumentation__.Notify(598862)
	return makeBuiltin(tree.FunctionProperties{NullableArgs: true},
		tree.Overload{
			Types:      tree.VariadicType{VarType: types.String},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598863)
				h := newHash()
				if ok, err := feedHash(h, args); !ok || func() bool {
					__antithesis_instrumentation__.Notify(598865)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(598866)
					return tree.DNull, err
				} else {
					__antithesis_instrumentation__.Notify(598867)
				}
				__antithesis_instrumentation__.Notify(598864)
				return tree.NewDString(fmt.Sprintf("%x", h.Sum(nil))), nil
			},
			Info:       info,
			Volatility: tree.VolatilityLeakProof,
		},
		tree.Overload{
			Types:      tree.VariadicType{VarType: types.Bytes},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598868)
				h := newHash()
				if ok, err := feedHash(h, args); !ok || func() bool {
					__antithesis_instrumentation__.Notify(598870)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(598871)
					return tree.DNull, err
				} else {
					__antithesis_instrumentation__.Notify(598872)
				}
				__antithesis_instrumentation__.Notify(598869)
				return tree.NewDString(fmt.Sprintf("%x", h.Sum(nil))), nil
			},
			Info:       info,
			Volatility: tree.VolatilityLeakProof,
		},
	)
}

func hash32Builtin(newHash func() hash.Hash32, info string) builtinDefinition {
	__antithesis_instrumentation__.Notify(598873)
	return makeBuiltin(tree.FunctionProperties{NullableArgs: true},
		tree.Overload{
			Types:      tree.VariadicType{VarType: types.String},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598874)
				h := newHash()
				if ok, err := feedHash(h, args); !ok || func() bool {
					__antithesis_instrumentation__.Notify(598876)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(598877)
					return tree.DNull, err
				} else {
					__antithesis_instrumentation__.Notify(598878)
				}
				__antithesis_instrumentation__.Notify(598875)
				return tree.NewDInt(tree.DInt(h.Sum32())), nil
			},
			Info:       info,
			Volatility: tree.VolatilityLeakProof,
		},
		tree.Overload{
			Types:      tree.VariadicType{VarType: types.Bytes},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598879)
				h := newHash()
				if ok, err := feedHash(h, args); !ok || func() bool {
					__antithesis_instrumentation__.Notify(598881)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(598882)
					return tree.DNull, err
				} else {
					__antithesis_instrumentation__.Notify(598883)
				}
				__antithesis_instrumentation__.Notify(598880)
				return tree.NewDInt(tree.DInt(h.Sum32())), nil
			},
			Info:       info,
			Volatility: tree.VolatilityLeakProof,
		},
	)
}

func hash64Builtin(newHash func() hash.Hash64, info string) builtinDefinition {
	__antithesis_instrumentation__.Notify(598884)
	return makeBuiltin(tree.FunctionProperties{NullableArgs: true},
		tree.Overload{
			Types:      tree.VariadicType{VarType: types.String},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598885)
				h := newHash()
				if ok, err := feedHash(h, args); !ok || func() bool {
					__antithesis_instrumentation__.Notify(598887)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(598888)
					return tree.DNull, err
				} else {
					__antithesis_instrumentation__.Notify(598889)
				}
				__antithesis_instrumentation__.Notify(598886)
				return tree.NewDInt(tree.DInt(h.Sum64())), nil
			},
			Info:       info,
			Volatility: tree.VolatilityLeakProof,
		},
		tree.Overload{
			Types:      tree.VariadicType{VarType: types.Bytes},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(598890)
				h := newHash()
				if ok, err := feedHash(h, args); !ok || func() bool {
					__antithesis_instrumentation__.Notify(598892)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(598893)
					return tree.DNull, err
				} else {
					__antithesis_instrumentation__.Notify(598894)
				}
				__antithesis_instrumentation__.Notify(598891)
				return tree.NewDInt(tree.DInt(h.Sum64())), nil
			},
			Info:       info,
			Volatility: tree.VolatilityLeakProof,
		},
	)
}

type regexpEscapeKey struct {
	sqlPattern string
	sqlEscape  string
}

func (k regexpEscapeKey) Pattern() (string, error) {
	__antithesis_instrumentation__.Notify(598895)
	pattern := k.sqlPattern
	if k.sqlEscape != `\` {
		__antithesis_instrumentation__.Notify(598897)
		pattern = strings.Replace(pattern, `\`, `\\`, -1)
		pattern = strings.Replace(pattern, k.sqlEscape, `\`, -1)
	} else {
		__antithesis_instrumentation__.Notify(598898)
	}
	__antithesis_instrumentation__.Notify(598896)
	return pattern, nil
}

func regexpExtract(ctx *tree.EvalContext, s, pattern, escape string) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(598899)
	patternRe, err := ctx.ReCache.GetRegexp(regexpEscapeKey{pattern, escape})
	if err != nil {
		__antithesis_instrumentation__.Notify(598903)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(598904)
	}
	__antithesis_instrumentation__.Notify(598900)

	match := patternRe.FindStringSubmatch(s)
	if match == nil {
		__antithesis_instrumentation__.Notify(598905)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(598906)
	}
	__antithesis_instrumentation__.Notify(598901)

	if len(match) > 1 {
		__antithesis_instrumentation__.Notify(598907)
		return tree.NewDString(match[1]), nil
	} else {
		__antithesis_instrumentation__.Notify(598908)
	}
	__antithesis_instrumentation__.Notify(598902)
	return tree.NewDString(match[0]), nil
}

type regexpFlagKey struct {
	sqlPattern string
	sqlFlags   string
}

func (k regexpFlagKey) Pattern() (string, error) {
	__antithesis_instrumentation__.Notify(598909)
	return regexpEvalFlags(k.sqlPattern, k.sqlFlags)
}

func regexpSplit(ctx *tree.EvalContext, args tree.Datums, hasFlags bool) ([]string, error) {
	__antithesis_instrumentation__.Notify(598910)
	s := string(tree.MustBeDString(args[0]))
	pattern := string(tree.MustBeDString(args[1]))
	sqlFlags := ""
	if hasFlags {
		__antithesis_instrumentation__.Notify(598913)
		sqlFlags = string(tree.MustBeDString(args[2]))
	} else {
		__antithesis_instrumentation__.Notify(598914)
	}
	__antithesis_instrumentation__.Notify(598911)
	patternRe, err := ctx.ReCache.GetRegexp(regexpFlagKey{pattern, sqlFlags})
	if err != nil {
		__antithesis_instrumentation__.Notify(598915)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(598916)
	}
	__antithesis_instrumentation__.Notify(598912)
	return patternRe.Split(s, -1), nil
}

func regexpSplitToArray(
	ctx *tree.EvalContext, args tree.Datums, hasFlags bool,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(598917)
	words, err := regexpSplit(ctx, args, hasFlags)
	if err != nil {
		__antithesis_instrumentation__.Notify(598920)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(598921)
	}
	__antithesis_instrumentation__.Notify(598918)
	result := tree.NewDArray(types.String)
	for _, word := range words {
		__antithesis_instrumentation__.Notify(598922)
		if err := result.Append(tree.NewDString(word)); err != nil {
			__antithesis_instrumentation__.Notify(598923)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(598924)
		}
	}
	__antithesis_instrumentation__.Notify(598919)
	return result, nil
}

func regexpReplace(ctx *tree.EvalContext, s, pattern, to, sqlFlags string) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(598925)
	patternRe, err := ctx.ReCache.GetRegexp(regexpFlagKey{pattern, sqlFlags})
	if err != nil {
		__antithesis_instrumentation__.Notify(598929)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(598930)
	}
	__antithesis_instrumentation__.Notify(598926)

	matchCount := 1
	if strings.ContainsRune(sqlFlags, 'g') {
		__antithesis_instrumentation__.Notify(598931)
		matchCount = -1
	} else {
		__antithesis_instrumentation__.Notify(598932)
	}
	__antithesis_instrumentation__.Notify(598927)

	finalIndex := 0
	var newString bytes.Buffer

	for _, matchIndex := range patternRe.FindAllStringSubmatchIndex(s, matchCount) {
		__antithesis_instrumentation__.Notify(598933)

		matchStart := matchIndex[0]
		matchEnd := matchIndex[1]

		preMatch := s[finalIndex:matchStart]
		newString.WriteString(preMatch)

		chunkStart := 0

		i := 0
		for i < len(to) {
			__antithesis_instrumentation__.Notify(598935)
			if to[i] == '\\' && func() bool {
				__antithesis_instrumentation__.Notify(598937)
				return i+1 < len(to) == true
			}() == true {
				__antithesis_instrumentation__.Notify(598938)
				i++
				if to[i] == '\\' {
					__antithesis_instrumentation__.Notify(598939)

					newString.WriteString(to[chunkStart:i])
					chunkStart = i + 1
				} else {
					__antithesis_instrumentation__.Notify(598940)
					if ('0' <= to[i] && func() bool {
						__antithesis_instrumentation__.Notify(598941)
						return to[i] <= '9' == true
					}() == true) || func() bool {
						__antithesis_instrumentation__.Notify(598942)
						return to[i] == '&' == true
					}() == true {
						__antithesis_instrumentation__.Notify(598943)
						newString.WriteString(to[chunkStart : i-1])
						chunkStart = i + 1
						if to[i] == '&' {
							__antithesis_instrumentation__.Notify(598944)

							newString.WriteString(s[matchStart:matchEnd])
						} else {
							__antithesis_instrumentation__.Notify(598945)
							captureGroupNumber := int(to[i] - '0')

							if matchIndexPos := 2 * captureGroupNumber; matchIndexPos < len(matchIndex) {
								__antithesis_instrumentation__.Notify(598946)
								startPos := matchIndex[matchIndexPos]
								endPos := matchIndex[matchIndexPos+1]
								if startPos >= 0 {
									__antithesis_instrumentation__.Notify(598947)
									newString.WriteString(s[startPos:endPos])
								} else {
									__antithesis_instrumentation__.Notify(598948)
								}
							} else {
								__antithesis_instrumentation__.Notify(598949)
							}
						}
					} else {
						__antithesis_instrumentation__.Notify(598950)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(598951)
			}
			__antithesis_instrumentation__.Notify(598936)
			i++
		}
		__antithesis_instrumentation__.Notify(598934)
		newString.WriteString(to[chunkStart:])

		finalIndex = matchEnd
	}
	__antithesis_instrumentation__.Notify(598928)

	newString.WriteString(s[finalIndex:])

	return tree.NewDString(newString.String()), nil
}

var flagToByte = map[syntax.Flags]byte{
	syntax.FoldCase: 'i',
	syntax.DotNL:    's',
}

var flagToNotByte = map[syntax.Flags]byte{
	syntax.OneLine: 'm',
}

func regexpEvalFlags(pattern, sqlFlags string) (string, error) {
	__antithesis_instrumentation__.Notify(598952)
	flags := syntax.DotNL | syntax.OneLine

	for _, sqlFlag := range sqlFlags {
		__antithesis_instrumentation__.Notify(598957)
		switch sqlFlag {
		case 'g':
			__antithesis_instrumentation__.Notify(598958)

		case 'i':
			__antithesis_instrumentation__.Notify(598959)
			flags |= syntax.FoldCase
		case 'c':
			__antithesis_instrumentation__.Notify(598960)
			flags &^= syntax.FoldCase
		case 's':
			__antithesis_instrumentation__.Notify(598961)
			flags &^= syntax.DotNL
		case 'm', 'n':
			__antithesis_instrumentation__.Notify(598962)
			flags &^= syntax.DotNL
			flags &^= syntax.OneLine
		case 'p':
			__antithesis_instrumentation__.Notify(598963)
			flags &^= syntax.DotNL
			flags |= syntax.OneLine
		case 'w':
			__antithesis_instrumentation__.Notify(598964)
			flags |= syntax.DotNL
			flags &^= syntax.OneLine
		default:
			__antithesis_instrumentation__.Notify(598965)
			return "", pgerror.Newf(
				pgcode.InvalidRegularExpression, "invalid regexp flag: %q", sqlFlag)
		}
	}
	__antithesis_instrumentation__.Notify(598953)

	var goFlags bytes.Buffer
	for flag, b := range flagToByte {
		__antithesis_instrumentation__.Notify(598966)
		if flags&flag != 0 {
			__antithesis_instrumentation__.Notify(598967)
			goFlags.WriteByte(b)
		} else {
			__antithesis_instrumentation__.Notify(598968)
		}
	}
	__antithesis_instrumentation__.Notify(598954)
	for flag, b := range flagToNotByte {
		__antithesis_instrumentation__.Notify(598969)
		if flags&flag == 0 {
			__antithesis_instrumentation__.Notify(598970)
			goFlags.WriteByte(b)
		} else {
			__antithesis_instrumentation__.Notify(598971)
		}
	}
	__antithesis_instrumentation__.Notify(598955)

	bs := goFlags.Bytes()
	if len(bs) == 0 {
		__antithesis_instrumentation__.Notify(598972)
		return pattern, nil
	} else {
		__antithesis_instrumentation__.Notify(598973)
	}
	__antithesis_instrumentation__.Notify(598956)
	return fmt.Sprintf("(?%s:%s)", bs, pattern), nil
}

func overlay(s, to string, pos, size int) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(598974)
	if pos < 1 {
		__antithesis_instrumentation__.Notify(598978)
		return nil, pgerror.Newf(
			pgcode.InvalidParameterValue, "non-positive substring length not allowed: %d", pos)
	} else {
		__antithesis_instrumentation__.Notify(598979)
	}
	__antithesis_instrumentation__.Notify(598975)
	pos--

	runes := []rune(s)
	if pos > len(runes) {
		__antithesis_instrumentation__.Notify(598980)
		pos = len(runes)
	} else {
		__antithesis_instrumentation__.Notify(598981)
	}
	__antithesis_instrumentation__.Notify(598976)
	after := pos + size
	if after < 0 {
		__antithesis_instrumentation__.Notify(598982)
		after = 0
	} else {
		__antithesis_instrumentation__.Notify(598983)
		if after > len(runes) {
			__antithesis_instrumentation__.Notify(598984)
			after = len(runes)
		} else {
			__antithesis_instrumentation__.Notify(598985)
		}
	}
	__antithesis_instrumentation__.Notify(598977)
	return tree.NewDString(string(runes[:pos]) + to + string(runes[after:])), nil
}

const NodeIDBits = 15

func GenerateUniqueUnorderedID(instanceID base.SQLInstanceID) tree.DInt {
	__antithesis_instrumentation__.Notify(598986)
	orig := uint64(GenerateUniqueInt(instanceID))
	uniqueUnorderedID := mapToUnorderedUniqueInt(orig)
	return tree.DInt(uniqueUnorderedID)
}

func mapToUnorderedUniqueInt(val uint64) uint64 {
	__antithesis_instrumentation__.Notify(598987)

	ts := (val & ((uint64(math.MaxUint64) >> 16) << 15)) >> 15
	v := (bits.Reverse64(ts) >> 1) | (val & (1<<15 - 1))
	return v
}

func GenerateUniqueInt(instanceID base.SQLInstanceID) tree.DInt {
	__antithesis_instrumentation__.Notify(598988)
	const precision = uint64(10 * time.Microsecond)

	nowNanos := timeutil.Now().UnixNano()

	if nowNanos < uniqueIntEpoch {
		__antithesis_instrumentation__.Notify(598991)
		nowNanos = uniqueIntEpoch
	} else {
		__antithesis_instrumentation__.Notify(598992)
	}
	__antithesis_instrumentation__.Notify(598989)
	timestamp := uint64(nowNanos-uniqueIntEpoch) / precision

	uniqueIntState.Lock()
	if timestamp <= uniqueIntState.timestamp {
		__antithesis_instrumentation__.Notify(598993)
		timestamp = uniqueIntState.timestamp + 1
	} else {
		__antithesis_instrumentation__.Notify(598994)
	}
	__antithesis_instrumentation__.Notify(598990)
	uniqueIntState.timestamp = timestamp
	uniqueIntState.Unlock()

	return GenerateUniqueID(int32(instanceID), timestamp)
}

func GenerateUniqueID(instanceID int32, timestamp uint64) tree.DInt {
	__antithesis_instrumentation__.Notify(598995)

	id := (timestamp << NodeIDBits) ^ uint64(instanceID)
	return tree.DInt(id)
}

func cardinality(arr *tree.DArray) tree.Datum {
	__antithesis_instrumentation__.Notify(598996)
	if arr.ParamTyp.Family() != types.ArrayFamily {
		__antithesis_instrumentation__.Notify(598999)
		return tree.NewDInt(tree.DInt(arr.Len()))
	} else {
		__antithesis_instrumentation__.Notify(599000)
	}
	__antithesis_instrumentation__.Notify(598997)
	card := 0
	for _, a := range arr.Array {
		__antithesis_instrumentation__.Notify(599001)
		card += int(tree.MustBeDInt(cardinality(tree.MustBeDArray(a))))
	}
	__antithesis_instrumentation__.Notify(598998)
	return tree.NewDInt(tree.DInt(card))
}

func arrayLength(arr *tree.DArray, dim int64) tree.Datum {
	__antithesis_instrumentation__.Notify(599002)
	if arr.Len() == 0 || func() bool {
		__antithesis_instrumentation__.Notify(599006)
		return dim < 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(599007)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(599008)
	}
	__antithesis_instrumentation__.Notify(599003)
	if dim == 1 {
		__antithesis_instrumentation__.Notify(599009)
		return tree.NewDInt(tree.DInt(arr.Len()))
	} else {
		__antithesis_instrumentation__.Notify(599010)
	}
	__antithesis_instrumentation__.Notify(599004)
	a, ok := tree.AsDArray(arr.Array[0])
	if !ok {
		__antithesis_instrumentation__.Notify(599011)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(599012)
	}
	__antithesis_instrumentation__.Notify(599005)
	return arrayLength(a, dim-1)
}

var intOne = tree.NewDInt(tree.DInt(1))

func arrayLower(arr *tree.DArray, dim int64) tree.Datum {
	__antithesis_instrumentation__.Notify(599013)
	if arr.Len() == 0 || func() bool {
		__antithesis_instrumentation__.Notify(599017)
		return dim < 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(599018)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(599019)
	}
	__antithesis_instrumentation__.Notify(599014)
	if dim == 1 {
		__antithesis_instrumentation__.Notify(599020)
		return intOne
	} else {
		__antithesis_instrumentation__.Notify(599021)
	}
	__antithesis_instrumentation__.Notify(599015)
	a, ok := tree.AsDArray(arr.Array[0])
	if !ok {
		__antithesis_instrumentation__.Notify(599022)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(599023)
	}
	__antithesis_instrumentation__.Notify(599016)
	return arrayLower(a, dim-1)
}

var extractBuiltin = makeBuiltin(
	tree.FunctionProperties{Category: categoryDateAndTime},
	tree.Overload{
		Types:      tree.ArgTypes{{"element", types.String}, {"input", types.Timestamp}},
		ReturnType: tree.FixedReturnType(types.Float),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(599024)

			fromTS := args[1].(*tree.DTimestamp)
			timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
			return extractTimeSpanFromTimestamp(ctx, fromTS.Time, timeSpan)
		},
		Info: "Extracts `element` from `input`.\n\n" +
			"Compatible elements: millennium, century, decade, year, isoyear,\n" +
			"quarter, month, week, dayofweek, isodow, dayofyear, julian,\n" +
			"hour, minute, second, millisecond, microsecond, epoch",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types:      tree.ArgTypes{{"element", types.String}, {"input", types.Interval}},
		ReturnType: tree.FixedReturnType(types.Float),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(599025)
			fromInterval := args[1].(*tree.DInterval)
			timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
			return extractTimeSpanFromInterval(fromInterval, timeSpan)
		},
		Info: "Extracts `element` from `input`.\n\n" +
			"Compatible elements: millennium, century, decade, year,\n" +
			"month, day, hour, minute, second, millisecond, microsecond, epoch",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types:      tree.ArgTypes{{"element", types.String}, {"input", types.Date}},
		ReturnType: tree.FixedReturnType(types.Float),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(599026)
			timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
			date := args[1].(*tree.DDate)
			fromTime, err := date.ToTime()
			if err != nil {
				__antithesis_instrumentation__.Notify(599028)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(599029)
			}
			__antithesis_instrumentation__.Notify(599027)
			return extractTimeSpanFromTimestamp(ctx, fromTime, timeSpan)
		},
		Info: "Extracts `element` from `input`.\n\n" +
			"Compatible elements: millennium, century, decade, year, isoyear,\n" +
			"quarter, month, week, dayofweek, isodow, dayofyear, julian,\n" +
			"hour, minute, second, millisecond, microsecond, epoch",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types:      tree.ArgTypes{{"element", types.String}, {"input", types.TimestampTZ}},
		ReturnType: tree.FixedReturnType(types.Float),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(599030)
			fromTSTZ := args[1].(*tree.DTimestampTZ)
			timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
			return extractTimeSpanFromTimestampTZ(ctx, fromTSTZ.Time.In(ctx.GetLocation()), timeSpan)
		},
		Info: "Extracts `element` from `input`.\n\n" +
			"Compatible elements: millennium, century, decade, year, isoyear,\n" +
			"quarter, month, week, dayofweek, isodow, dayofyear, julian,\n" +
			"hour, minute, second, millisecond, microsecond, epoch,\n" +
			"timezone, timezone_hour, timezone_minute",
		Volatility: tree.VolatilityStable,
	},
	tree.Overload{
		Types:      tree.ArgTypes{{"element", types.String}, {"input", types.Time}},
		ReturnType: tree.FixedReturnType(types.Float),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(599031)
			fromTime := args[1].(*tree.DTime)
			timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
			return extractTimeSpanFromTime(fromTime, timeSpan)
		},
		Info: "Extracts `element` from `input`.\n\n" +
			"Compatible elements: hour, minute, second, millisecond, microsecond, epoch",
		Volatility: tree.VolatilityImmutable,
	},
	tree.Overload{
		Types:      tree.ArgTypes{{"element", types.String}, {"input", types.TimeTZ}},
		ReturnType: tree.FixedReturnType(types.Float),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(599032)
			fromTime := args[1].(*tree.DTimeTZ)
			timeSpan := strings.ToLower(string(tree.MustBeDString(args[0])))
			return extractTimeSpanFromTimeTZ(fromTime, timeSpan)
		},
		Info: "Extracts `element` from `input`.\n\n" +
			"Compatible elements: hour, minute, second, millisecond, microsecond, epoch,\n" +
			"timezone, timezone_hour, timezone_minute",
		Volatility: tree.VolatilityImmutable,
	},
)

func extractTimeSpanFromTime(fromTime *tree.DTime, timeSpan string) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599033)
	t := timeofday.TimeOfDay(*fromTime)
	return extractTimeSpanFromTimeOfDay(t, timeSpan)
}

func extractTimezoneFromOffset(offsetSecs int32, timeSpan string) tree.Datum {
	__antithesis_instrumentation__.Notify(599034)
	switch timeSpan {
	case "timezone":
		__antithesis_instrumentation__.Notify(599036)
		return tree.NewDFloat(tree.DFloat(float64(offsetSecs)))
	case "timezone_hour", "timezone_hours":
		__antithesis_instrumentation__.Notify(599037)
		numHours := offsetSecs / duration.SecsPerHour
		return tree.NewDFloat(tree.DFloat(float64(numHours)))
	case "timezone_minute", "timezone_minutes":
		__antithesis_instrumentation__.Notify(599038)
		numMinutes := offsetSecs / duration.SecsPerMinute
		return tree.NewDFloat(tree.DFloat(float64(numMinutes % 60)))
	default:
		__antithesis_instrumentation__.Notify(599039)
	}
	__antithesis_instrumentation__.Notify(599035)
	return nil
}

func extractTimeSpanFromTimeTZ(fromTime *tree.DTimeTZ, timeSpan string) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599040)
	if ret := extractTimezoneFromOffset(-fromTime.OffsetSecs, timeSpan); ret != nil {
		__antithesis_instrumentation__.Notify(599043)
		return ret, nil
	} else {
		__antithesis_instrumentation__.Notify(599044)
	}
	__antithesis_instrumentation__.Notify(599041)
	switch timeSpan {
	case "epoch":
		__antithesis_instrumentation__.Notify(599045)

		seconds := float64(time.Duration(fromTime.TimeOfDay))*float64(time.Microsecond)/float64(time.Second) + float64(fromTime.OffsetSecs)
		return tree.NewDFloat(tree.DFloat(seconds)), nil
	default:
		__antithesis_instrumentation__.Notify(599046)
	}
	__antithesis_instrumentation__.Notify(599042)
	return extractTimeSpanFromTimeOfDay(fromTime.TimeOfDay, timeSpan)
}

func extractTimeSpanFromTimeOfDay(t timeofday.TimeOfDay, timeSpan string) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599047)
	switch timeSpan {
	case "hour", "hours":
		__antithesis_instrumentation__.Notify(599048)
		return tree.NewDFloat(tree.DFloat(t.Hour())), nil
	case "minute", "minutes":
		__antithesis_instrumentation__.Notify(599049)
		return tree.NewDFloat(tree.DFloat(t.Minute())), nil
	case "second", "seconds":
		__antithesis_instrumentation__.Notify(599050)
		return tree.NewDFloat(tree.DFloat(float64(t.Second()) + float64(t.Microsecond())/(duration.MicrosPerMilli*duration.MillisPerSec))), nil
	case "millisecond", "milliseconds":
		__antithesis_instrumentation__.Notify(599051)
		return tree.NewDFloat(tree.DFloat(float64(t.Second()*duration.MillisPerSec) + float64(t.Microsecond())/duration.MicrosPerMilli)), nil
	case "microsecond", "microseconds":
		__antithesis_instrumentation__.Notify(599052)
		return tree.NewDFloat(tree.DFloat((t.Second() * duration.MillisPerSec * duration.MicrosPerMilli) + t.Microsecond())), nil
	case "epoch":
		__antithesis_instrumentation__.Notify(599053)
		seconds := float64(time.Duration(t)) * float64(time.Microsecond) / float64(time.Second)
		return tree.NewDFloat(tree.DFloat(seconds)), nil
	default:
		__antithesis_instrumentation__.Notify(599054)
		return nil, pgerror.Newf(
			pgcode.InvalidParameterValue, "unsupported timespan: %s", timeSpan)
	}
}

func dateToJulianDay(year int, month int, day int) int {
	__antithesis_instrumentation__.Notify(599055)
	if month > 2 {
		__antithesis_instrumentation__.Notify(599057)
		month++
		year += 4800
	} else {
		__antithesis_instrumentation__.Notify(599058)
		month += 13
		year += 4799
	}
	__antithesis_instrumentation__.Notify(599056)

	century := year / 100
	jd := year*365 - 32167
	jd += year/4 - century + century/4
	jd += 7834*month/256 + day

	return jd
}

func extractTimeSpanFromTimestampTZ(
	ctx *tree.EvalContext, fromTime time.Time, timeSpan string,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599059)
	_, offsetSecs := fromTime.Zone()
	if ret := extractTimezoneFromOffset(int32(offsetSecs), timeSpan); ret != nil {
		__antithesis_instrumentation__.Notify(599061)
		return ret, nil
	} else {
		__antithesis_instrumentation__.Notify(599062)
	}
	__antithesis_instrumentation__.Notify(599060)

	switch timeSpan {
	case "epoch":
		__antithesis_instrumentation__.Notify(599063)
		return extractTimeSpanFromTimestamp(ctx, fromTime, timeSpan)
	default:
		__antithesis_instrumentation__.Notify(599064)

		pretendTime := fromTime.In(time.UTC).Add(time.Duration(offsetSecs) * time.Second)
		return extractTimeSpanFromTimestamp(ctx, pretendTime, timeSpan)
	}
}

func extractTimeSpanFromInterval(
	fromInterval *tree.DInterval, timeSpan string,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599065)
	switch timeSpan {
	case "millennia", "millennium", "millenniums":
		__antithesis_instrumentation__.Notify(599066)
		return tree.NewDFloat(tree.DFloat(fromInterval.Months / (duration.MonthsPerYear * 1000))), nil

	case "centuries", "century":
		__antithesis_instrumentation__.Notify(599067)
		return tree.NewDFloat(tree.DFloat(fromInterval.Months / (duration.MonthsPerYear * 100))), nil

	case "decade", "decades":
		__antithesis_instrumentation__.Notify(599068)
		return tree.NewDFloat(tree.DFloat(fromInterval.Months / (duration.MonthsPerYear * 10))), nil

	case "year", "years":
		__antithesis_instrumentation__.Notify(599069)
		return tree.NewDFloat(tree.DFloat(fromInterval.Months / duration.MonthsPerYear)), nil

	case "month", "months":
		__antithesis_instrumentation__.Notify(599070)
		return tree.NewDFloat(tree.DFloat(fromInterval.Months % duration.MonthsPerYear)), nil

	case "day", "days":
		__antithesis_instrumentation__.Notify(599071)
		return tree.NewDFloat(tree.DFloat(fromInterval.Days)), nil

	case "hour", "hours":
		__antithesis_instrumentation__.Notify(599072)
		return tree.NewDFloat(tree.DFloat(fromInterval.Nanos() / int64(time.Hour))), nil

	case "minute", "minutes":
		__antithesis_instrumentation__.Notify(599073)

		return tree.NewDFloat(tree.DFloat((fromInterval.Nanos() % int64(time.Second*duration.SecsPerHour)) / int64(time.Minute))), nil

	case "second", "seconds":
		__antithesis_instrumentation__.Notify(599074)
		return tree.NewDFloat(tree.DFloat(float64(fromInterval.Nanos()%int64(time.Minute)) / float64(time.Second))), nil

	case "millisecond", "milliseconds":
		__antithesis_instrumentation__.Notify(599075)

		return tree.NewDFloat(tree.DFloat(float64(fromInterval.Nanos()%int64(time.Minute)) / float64(time.Millisecond))), nil

	case "microsecond", "microseconds":
		__antithesis_instrumentation__.Notify(599076)
		return tree.NewDFloat(tree.DFloat(float64(fromInterval.Nanos()%int64(time.Minute)) / float64(time.Microsecond))), nil
	case "epoch":
		__antithesis_instrumentation__.Notify(599077)
		return tree.NewDFloat(tree.DFloat(fromInterval.AsFloat64())), nil
	default:
		__antithesis_instrumentation__.Notify(599078)
		return nil, pgerror.Newf(
			pgcode.InvalidParameterValue, "unsupported timespan: %s", timeSpan)
	}
}

func extractTimeSpanFromTimestamp(
	_ *tree.EvalContext, fromTime time.Time, timeSpan string,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599079)
	switch timeSpan {
	case "millennia", "millennium", "millenniums":
		__antithesis_instrumentation__.Notify(599080)
		year := fromTime.Year()
		if year > 0 {
			__antithesis_instrumentation__.Notify(599104)
			return tree.NewDFloat(tree.DFloat((year + 999) / 1000)), nil
		} else {
			__antithesis_instrumentation__.Notify(599105)
		}
		__antithesis_instrumentation__.Notify(599081)
		return tree.NewDFloat(tree.DFloat(-((999 - (year - 1)) / 1000))), nil

	case "centuries", "century":
		__antithesis_instrumentation__.Notify(599082)
		year := fromTime.Year()
		if year > 0 {
			__antithesis_instrumentation__.Notify(599106)
			return tree.NewDFloat(tree.DFloat((year + 99) / 100)), nil
		} else {
			__antithesis_instrumentation__.Notify(599107)
		}
		__antithesis_instrumentation__.Notify(599083)
		return tree.NewDFloat(tree.DFloat(-((99 - (year - 1)) / 100))), nil

	case "decade", "decades":
		__antithesis_instrumentation__.Notify(599084)
		year := fromTime.Year()
		if year >= 0 {
			__antithesis_instrumentation__.Notify(599108)
			return tree.NewDFloat(tree.DFloat(year / 10)), nil
		} else {
			__antithesis_instrumentation__.Notify(599109)
		}
		__antithesis_instrumentation__.Notify(599085)
		return tree.NewDFloat(tree.DFloat(-((8 - (year - 1)) / 10))), nil

	case "year", "years":
		__antithesis_instrumentation__.Notify(599086)
		return tree.NewDFloat(tree.DFloat(fromTime.Year())), nil

	case "isoyear":
		__antithesis_instrumentation__.Notify(599087)
		year, _ := fromTime.ISOWeek()
		return tree.NewDFloat(tree.DFloat(year)), nil

	case "quarter":
		__antithesis_instrumentation__.Notify(599088)
		return tree.NewDFloat(tree.DFloat((fromTime.Month()-1)/3 + 1)), nil

	case "month", "months":
		__antithesis_instrumentation__.Notify(599089)
		return tree.NewDFloat(tree.DFloat(fromTime.Month())), nil

	case "week", "weeks":
		__antithesis_instrumentation__.Notify(599090)
		_, week := fromTime.ISOWeek()
		return tree.NewDFloat(tree.DFloat(week)), nil

	case "day", "days":
		__antithesis_instrumentation__.Notify(599091)
		return tree.NewDFloat(tree.DFloat(fromTime.Day())), nil

	case "dayofweek", "dow":
		__antithesis_instrumentation__.Notify(599092)
		return tree.NewDFloat(tree.DFloat(fromTime.Weekday())), nil

	case "isodow":
		__antithesis_instrumentation__.Notify(599093)
		day := fromTime.Weekday()
		if day == 0 {
			__antithesis_instrumentation__.Notify(599110)
			return tree.NewDFloat(tree.DFloat(7)), nil
		} else {
			__antithesis_instrumentation__.Notify(599111)
		}
		__antithesis_instrumentation__.Notify(599094)
		return tree.NewDFloat(tree.DFloat(day)), nil

	case "dayofyear", "doy":
		__antithesis_instrumentation__.Notify(599095)
		return tree.NewDFloat(tree.DFloat(fromTime.YearDay())), nil

	case "julian":
		__antithesis_instrumentation__.Notify(599096)
		julianDay := float64(dateToJulianDay(fromTime.Year(), int(fromTime.Month()), fromTime.Day())) +
			(float64(fromTime.Hour()*duration.SecsPerHour+fromTime.Minute()*duration.SecsPerMinute+fromTime.Second())+
				float64(fromTime.Nanosecond())/float64(time.Second))/duration.SecsPerDay
		return tree.NewDFloat(tree.DFloat(julianDay)), nil

	case "hour", "hours":
		__antithesis_instrumentation__.Notify(599097)
		return tree.NewDFloat(tree.DFloat(fromTime.Hour())), nil

	case "minute", "minutes":
		__antithesis_instrumentation__.Notify(599098)
		return tree.NewDFloat(tree.DFloat(fromTime.Minute())), nil

	case "second", "seconds":
		__antithesis_instrumentation__.Notify(599099)
		return tree.NewDFloat(tree.DFloat(float64(fromTime.Second()) + float64(fromTime.Nanosecond())/float64(time.Second))), nil

	case "millisecond", "milliseconds":
		__antithesis_instrumentation__.Notify(599100)

		return tree.NewDFloat(
			tree.DFloat(
				float64(fromTime.Second()*duration.MillisPerSec) + float64(fromTime.Nanosecond())/
					float64(time.Millisecond),
			),
		), nil

	case "microsecond", "microseconds":
		__antithesis_instrumentation__.Notify(599101)
		return tree.NewDFloat(
			tree.DFloat(
				float64(fromTime.Second()*duration.MillisPerSec*duration.MicrosPerMilli) + float64(fromTime.Nanosecond())/
					float64(time.Microsecond),
			),
		), nil

	case "epoch":
		__antithesis_instrumentation__.Notify(599102)
		return tree.NewDFloat(tree.DFloat(float64(fromTime.UnixNano()) / float64(time.Second))), nil

	default:
		__antithesis_instrumentation__.Notify(599103)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unsupported timespan: %s", timeSpan)
	}
}

func truncateTime(fromTime *tree.DTime, timeSpan string) (*tree.DTime, error) {
	__antithesis_instrumentation__.Notify(599112)
	t := timeofday.TimeOfDay(*fromTime)
	hour := t.Hour()
	min := t.Minute()
	sec := t.Second()
	micro := t.Microsecond()

	minTrunc := 0
	secTrunc := 0
	microTrunc := 0

	switch timeSpan {
	case "hour", "hours":
		__antithesis_instrumentation__.Notify(599114)
		min, sec, micro = minTrunc, secTrunc, microTrunc
	case "minute", "minutes":
		__antithesis_instrumentation__.Notify(599115)
		sec, micro = secTrunc, microTrunc
	case "second", "seconds":
		__antithesis_instrumentation__.Notify(599116)
		micro = microTrunc
	case "millisecond", "milliseconds":
		__antithesis_instrumentation__.Notify(599117)

		micro = (micro / duration.MicrosPerMilli) * duration.MicrosPerMilli
	case "microsecond", "microseconds":
		__antithesis_instrumentation__.Notify(599118)
	default:
		__antithesis_instrumentation__.Notify(599119)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unsupported timespan: %s", timeSpan)
	}
	__antithesis_instrumentation__.Notify(599113)

	return tree.MakeDTime(timeofday.New(hour, min, sec, micro)), nil
}

func stringOrNil(d tree.Datum) *string {
	__antithesis_instrumentation__.Notify(599120)
	if d == tree.DNull {
		__antithesis_instrumentation__.Notify(599122)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(599123)
	}
	__antithesis_instrumentation__.Notify(599121)
	s := string(tree.MustBeDString(d))
	return &s
}

func stringToArray(str string, delimPtr *string, nullStr *string) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599124)
	var split []string

	if delimPtr != nil {
		__antithesis_instrumentation__.Notify(599127)
		delim := *delimPtr
		if str == "" {
			__antithesis_instrumentation__.Notify(599128)
			split = nil
		} else {
			__antithesis_instrumentation__.Notify(599129)
			if delim == "" {
				__antithesis_instrumentation__.Notify(599130)
				split = []string{str}
			} else {
				__antithesis_instrumentation__.Notify(599131)
				split = strings.Split(str, delim)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(599132)

		split = make([]string, len(str))
		for i, c := range str {
			__antithesis_instrumentation__.Notify(599133)
			split[i] = string(c)
		}
	}
	__antithesis_instrumentation__.Notify(599125)

	result := tree.NewDArray(types.String)
	for _, s := range split {
		__antithesis_instrumentation__.Notify(599134)
		var next tree.Datum = tree.NewDString(s)
		if nullStr != nil && func() bool {
			__antithesis_instrumentation__.Notify(599136)
			return s == *nullStr == true
		}() == true {
			__antithesis_instrumentation__.Notify(599137)
			next = tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(599138)
		}
		__antithesis_instrumentation__.Notify(599135)
		if err := result.Append(next); err != nil {
			__antithesis_instrumentation__.Notify(599139)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(599140)
		}
	}
	__antithesis_instrumentation__.Notify(599126)
	return result, nil
}

func arrayToString(
	evalCtx *tree.EvalContext, arr *tree.DArray, delim string, nullStr *string,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599141)
	f := evalCtx.FmtCtx(tree.FmtArrayToString)

	for i := range arr.Array {
		__antithesis_instrumentation__.Notify(599143)
		if arr.Array[i] == tree.DNull {
			__antithesis_instrumentation__.Notify(599145)
			if nullStr == nil {
				__antithesis_instrumentation__.Notify(599147)
				continue
			} else {
				__antithesis_instrumentation__.Notify(599148)
			}
			__antithesis_instrumentation__.Notify(599146)
			f.WriteString(*nullStr)
		} else {
			__antithesis_instrumentation__.Notify(599149)
			f.FormatNode(arr.Array[i])
		}
		__antithesis_instrumentation__.Notify(599144)
		if i < len(arr.Array)-1 {
			__antithesis_instrumentation__.Notify(599150)
			f.WriteString(delim)
		} else {
			__antithesis_instrumentation__.Notify(599151)
		}
	}
	__antithesis_instrumentation__.Notify(599142)
	return tree.NewDString(f.CloseAndGetString()), nil
}

func encodeEscape(input []byte) string {
	__antithesis_instrumentation__.Notify(599152)
	var result bytes.Buffer
	start := 0
	for i := range input {
		__antithesis_instrumentation__.Notify(599154)
		if input[i] == 0 || func() bool {
			__antithesis_instrumentation__.Notify(599155)
			return input[i]&128 != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(599156)
			result.Write(input[start:i])
			start = i + 1
			result.WriteString(fmt.Sprintf(`\%03o`, input[i]))
		} else {
			__antithesis_instrumentation__.Notify(599157)
			if input[i] == '\\' {
				__antithesis_instrumentation__.Notify(599158)
				result.Write(input[start:i])
				start = i + 1
				result.WriteString(`\\`)
			} else {
				__antithesis_instrumentation__.Notify(599159)
			}
		}
	}
	__antithesis_instrumentation__.Notify(599153)
	result.Write(input[start:])
	return result.String()
}

var errInvalidSyntaxForDecode = pgerror.New(pgcode.InvalidParameterValue, "invalid syntax for decode(..., 'escape')")

func isOctalDigit(c byte) bool {
	__antithesis_instrumentation__.Notify(599160)
	return '0' <= c && func() bool {
		__antithesis_instrumentation__.Notify(599161)
		return c <= '7' == true
	}() == true
}

func decodeOctalTriplet(input string) byte {
	__antithesis_instrumentation__.Notify(599162)
	return (input[0]-'0')*64 + (input[1]-'0')*8 + (input[2] - '0')
}

func decodeEscape(input string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(599163)
	result := make([]byte, 0, len(input))
	for i := 0; i < len(input); i++ {
		__antithesis_instrumentation__.Notify(599165)
		if input[i] == '\\' {
			__antithesis_instrumentation__.Notify(599166)
			if i+1 < len(input) && func() bool {
				__antithesis_instrumentation__.Notify(599167)
				return input[i+1] == '\\' == true
			}() == true {
				__antithesis_instrumentation__.Notify(599168)
				result = append(result, '\\')
				i++
			} else {
				__antithesis_instrumentation__.Notify(599169)
				if i+3 < len(input) && func() bool {
					__antithesis_instrumentation__.Notify(599170)
					return isOctalDigit(input[i+1]) == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(599171)
					return isOctalDigit(input[i+2]) == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(599172)
					return isOctalDigit(input[i+3]) == true
				}() == true {
					__antithesis_instrumentation__.Notify(599173)
					result = append(result, decodeOctalTriplet(input[i+1:i+4]))
					i += 3
				} else {
					__antithesis_instrumentation__.Notify(599174)
					return nil, errInvalidSyntaxForDecode
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(599175)
			result = append(result, input[i])
		}
	}
	__antithesis_instrumentation__.Notify(599164)
	return result, nil
}

func truncateTimestamp(fromTime time.Time, timeSpan string) (*tree.DTimestampTZ, error) {
	__antithesis_instrumentation__.Notify(599176)
	year := fromTime.Year()
	month := fromTime.Month()
	day := fromTime.Day()
	hour := fromTime.Hour()
	min := fromTime.Minute()
	sec := fromTime.Second()
	nsec := fromTime.Nanosecond()
	loc := fromTime.Location()

	_, origZoneOffset := fromTime.Zone()

	monthTrunc := time.January
	dayTrunc := 1
	hourTrunc := 0
	minTrunc := 0
	secTrunc := 0
	nsecTrunc := 0

	switch timeSpan {
	case "millennia", "millennium", "millenniums":
		__antithesis_instrumentation__.Notify(599179)
		if year > 0 {
			__antithesis_instrumentation__.Notify(599197)
			year = ((year+999)/1000)*1000 - 999
		} else {
			__antithesis_instrumentation__.Notify(599198)
			year = -((999-(year-1))/1000)*1000 + 1
		}
		__antithesis_instrumentation__.Notify(599180)
		month, day, hour, min, sec, nsec = monthTrunc, dayTrunc, hourTrunc, minTrunc, secTrunc, nsecTrunc

	case "centuries", "century":
		__antithesis_instrumentation__.Notify(599181)
		if year > 0 {
			__antithesis_instrumentation__.Notify(599199)
			year = ((year+99)/100)*100 - 99
		} else {
			__antithesis_instrumentation__.Notify(599200)
			year = -((99-(year-1))/100)*100 + 1
		}
		__antithesis_instrumentation__.Notify(599182)
		month, day, hour, min, sec, nsec = monthTrunc, dayTrunc, hourTrunc, minTrunc, secTrunc, nsecTrunc

	case "decade", "decades":
		__antithesis_instrumentation__.Notify(599183)
		if year >= 0 {
			__antithesis_instrumentation__.Notify(599201)
			year = (year / 10) * 10
		} else {
			__antithesis_instrumentation__.Notify(599202)
			year = -((8 - (year - 1)) / 10) * 10
		}
		__antithesis_instrumentation__.Notify(599184)
		month, day, hour, min, sec, nsec = monthTrunc, dayTrunc, hourTrunc, minTrunc, secTrunc, nsecTrunc

	case "year", "years":
		__antithesis_instrumentation__.Notify(599185)
		month, day, hour, min, sec, nsec = monthTrunc, dayTrunc, hourTrunc, minTrunc, secTrunc, nsecTrunc

	case "quarter":
		__antithesis_instrumentation__.Notify(599186)
		firstMonthInQuarter := ((month-1)/3)*3 + 1
		month, day, hour, min, sec, nsec = firstMonthInQuarter, dayTrunc, hourTrunc, minTrunc, secTrunc, nsecTrunc

	case "month", "months":
		__antithesis_instrumentation__.Notify(599187)
		day, hour, min, sec, nsec = dayTrunc, hourTrunc, minTrunc, secTrunc, nsecTrunc

	case "week", "weeks":
		__antithesis_instrumentation__.Notify(599188)

		previousMonday := fromTime.Add(-1 * time.Hour * 24 * time.Duration(fromTime.Weekday()-1))
		if fromTime.Weekday() == time.Sunday {
			__antithesis_instrumentation__.Notify(599203)

			previousMonday = fromTime.Add(-6 * time.Hour * 24)
		} else {
			__antithesis_instrumentation__.Notify(599204)
		}
		__antithesis_instrumentation__.Notify(599189)
		year, month, day = previousMonday.Year(), previousMonday.Month(), previousMonday.Day()
		hour, min, sec, nsec = hourTrunc, minTrunc, secTrunc, nsecTrunc

	case "day", "days":
		__antithesis_instrumentation__.Notify(599190)
		hour, min, sec, nsec = hourTrunc, minTrunc, secTrunc, nsecTrunc

	case "hour", "hours":
		__antithesis_instrumentation__.Notify(599191)
		min, sec, nsec = minTrunc, secTrunc, nsecTrunc

	case "minute", "minutes":
		__antithesis_instrumentation__.Notify(599192)
		sec, nsec = secTrunc, nsecTrunc

	case "second", "seconds":
		__antithesis_instrumentation__.Notify(599193)
		nsec = nsecTrunc

	case "millisecond", "milliseconds":
		__antithesis_instrumentation__.Notify(599194)

		milliseconds := (nsec / int(time.Millisecond)) * int(time.Millisecond)
		nsec = milliseconds

	case "microsecond", "microseconds":
		__antithesis_instrumentation__.Notify(599195)
		microseconds := (nsec / int(time.Microsecond)) * int(time.Microsecond)
		nsec = microseconds

	default:
		__antithesis_instrumentation__.Notify(599196)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unsupported timespan: %s", timeSpan)
	}
	__antithesis_instrumentation__.Notify(599177)

	toTime := time.Date(year, month, day, hour, min, sec, nsec, loc)
	_, newZoneOffset := toTime.Zone()

	if origZoneOffset != newZoneOffset {
		__antithesis_instrumentation__.Notify(599205)

		fixedOffsetLoc := timeutil.FixedTimeZoneOffsetToLocation(origZoneOffset, "date_trunc")
		fixedOffsetTime := time.Date(year, month, day, hour, min, sec, nsec, fixedOffsetLoc)
		locCorrectedOffsetTime := fixedOffsetTime.In(loc)

		if _, zoneOffset := locCorrectedOffsetTime.Zone(); origZoneOffset == zoneOffset {
			__antithesis_instrumentation__.Notify(599206)
			toTime = locCorrectedOffsetTime
		} else {
			__antithesis_instrumentation__.Notify(599207)
		}
	} else {
		__antithesis_instrumentation__.Notify(599208)
	}
	__antithesis_instrumentation__.Notify(599178)
	return tree.MakeDTimestampTZ(toTime, time.Microsecond)
}

func truncateInterval(fromInterval *tree.DInterval, timeSpan string) (*tree.DInterval, error) {
	__antithesis_instrumentation__.Notify(599209)

	toInterval := tree.DInterval{}

	switch timeSpan {
	case "millennia", "millennium", "millenniums":
		__antithesis_instrumentation__.Notify(599211)
		toInterval.Months = fromInterval.Months - fromInterval.Months%(12*1000)

	case "centuries", "century":
		__antithesis_instrumentation__.Notify(599212)
		toInterval.Months = fromInterval.Months - fromInterval.Months%(12*100)

	case "decade", "decades":
		__antithesis_instrumentation__.Notify(599213)
		toInterval.Months = fromInterval.Months - fromInterval.Months%(12*10)

	case "year", "years":
		__antithesis_instrumentation__.Notify(599214)
		toInterval.Months = fromInterval.Months - fromInterval.Months%12

	case "quarter":
		__antithesis_instrumentation__.Notify(599215)
		toInterval.Months = fromInterval.Months - fromInterval.Months%3

	case "month", "months":
		__antithesis_instrumentation__.Notify(599216)
		toInterval.Months = fromInterval.Months

	case "week", "weeks":
		__antithesis_instrumentation__.Notify(599217)

		return nil, pgerror.Newf(pgcode.FeatureNotSupported, "interval units %q not supported because months usually have fractional weeks", timeSpan)

	case "day", "days":
		__antithesis_instrumentation__.Notify(599218)
		toInterval.Months = fromInterval.Months
		toInterval.Days = fromInterval.Days

	case "hour", "hours":
		__antithesis_instrumentation__.Notify(599219)
		toInterval.Months = fromInterval.Months
		toInterval.Days = fromInterval.Days
		toInterval.SetNanos(fromInterval.Nanos() - fromInterval.Nanos()%time.Hour.Nanoseconds())

	case "minute", "minutes":
		__antithesis_instrumentation__.Notify(599220)
		toInterval.Months = fromInterval.Months
		toInterval.Days = fromInterval.Days
		toInterval.SetNanos(fromInterval.Nanos() - fromInterval.Nanos()%time.Minute.Nanoseconds())

	case "second", "seconds":
		__antithesis_instrumentation__.Notify(599221)
		toInterval.Months = fromInterval.Months
		toInterval.Days = fromInterval.Days
		toInterval.SetNanos(fromInterval.Nanos() - fromInterval.Nanos()%time.Second.Nanoseconds())

	case "millisecond", "milliseconds":
		__antithesis_instrumentation__.Notify(599222)
		toInterval.Months = fromInterval.Months
		toInterval.Days = fromInterval.Days
		toInterval.SetNanos(fromInterval.Nanos() - fromInterval.Nanos()%time.Millisecond.Nanoseconds())

	case "microsecond", "microseconds":
		__antithesis_instrumentation__.Notify(599223)
		toInterval.Months = fromInterval.Months
		toInterval.Days = fromInterval.Days
		toInterval.SetNanos(fromInterval.Nanos() - fromInterval.Nanos()%time.Microsecond.Nanoseconds())

	default:
		__antithesis_instrumentation__.Notify(599224)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "interval units %q not recognized", timeSpan)
	}
	__antithesis_instrumentation__.Notify(599210)

	return &toInterval, nil
}

func asJSONBuildObjectKey(
	d tree.Datum, dcc sessiondatapb.DataConversionConfig, loc *time.Location,
) (string, error) {
	__antithesis_instrumentation__.Notify(599225)
	switch t := d.(type) {
	case *tree.DJSON, *tree.DArray, *tree.DTuple:
		__antithesis_instrumentation__.Notify(599226)
		return "", pgerror.New(pgcode.InvalidParameterValue,
			"key value must be scalar, not array, tuple, or json")
	case *tree.DString:
		__antithesis_instrumentation__.Notify(599227)
		return string(*t), nil
	case *tree.DCollatedString:
		__antithesis_instrumentation__.Notify(599228)
		return t.Contents, nil
	case *tree.DTimestampTZ:
		__antithesis_instrumentation__.Notify(599229)
		ts, err := tree.MakeDTimestampTZ(t.Time.In(loc), time.Microsecond)
		if err != nil {
			__antithesis_instrumentation__.Notify(599233)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(599234)
		}
		__antithesis_instrumentation__.Notify(599230)
		return tree.AsStringWithFlags(
			ts,
			tree.FmtBareStrings,
			tree.FmtDataConversionConfig(dcc),
		), nil
	case *tree.DBool, *tree.DInt, *tree.DFloat, *tree.DDecimal, *tree.DTimestamp,
		*tree.DDate, *tree.DUuid, *tree.DInterval, *tree.DBytes, *tree.DIPAddr, *tree.DOid,
		*tree.DTime, *tree.DTimeTZ, *tree.DBitArray, *tree.DGeography, *tree.DGeometry, *tree.DBox2D:
		__antithesis_instrumentation__.Notify(599231)
		return tree.AsStringWithFlags(d, tree.FmtBareStrings), nil
	default:
		__antithesis_instrumentation__.Notify(599232)
		return "", errors.AssertionFailedf("unexpected type %T for key value", d)
	}
}

func asJSONObjectKey(d tree.Datum) (string, error) {
	__antithesis_instrumentation__.Notify(599235)
	switch t := d.(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(599236)
		return string(*t), nil
	default:
		__antithesis_instrumentation__.Notify(599237)
		return "", errors.AssertionFailedf("unexpected type %T for asJSONObjectKey", d)
	}
}

func toJSONObject(ctx *tree.EvalContext, d tree.Datum) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599238)
	j, err := tree.AsJSON(d, ctx.SessionData().DataConversionConfig, ctx.GetLocation())
	if err != nil {
		__antithesis_instrumentation__.Notify(599240)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599241)
	}
	__antithesis_instrumentation__.Notify(599239)
	return tree.NewDJSON(j), nil
}

func jsonValidate(_ *tree.EvalContext, string tree.DString) *tree.DBool {
	__antithesis_instrumentation__.Notify(599242)
	var js interface{}
	return tree.MakeDBool(gojson.Unmarshal([]byte(string), &js) == nil)
}

func padMaybeTruncate(s string, length int, fill string) (ok bool, slen int, ret string) {
	__antithesis_instrumentation__.Notify(599243)
	if length < 0 {
		__antithesis_instrumentation__.Notify(599248)

		length = 0
	} else {
		__antithesis_instrumentation__.Notify(599249)
	}
	__antithesis_instrumentation__.Notify(599244)
	slen = utf8.RuneCountInString(s)
	if length == slen {
		__antithesis_instrumentation__.Notify(599250)
		return true, slen, s
	} else {
		__antithesis_instrumentation__.Notify(599251)
	}
	__antithesis_instrumentation__.Notify(599245)

	if length < slen {
		__antithesis_instrumentation__.Notify(599252)
		return true, slen, string([]rune(s)[:length])
	} else {
		__antithesis_instrumentation__.Notify(599253)
	}
	__antithesis_instrumentation__.Notify(599246)

	if len(fill) == 0 {
		__antithesis_instrumentation__.Notify(599254)
		return true, slen, s
	} else {
		__antithesis_instrumentation__.Notify(599255)
	}
	__antithesis_instrumentation__.Notify(599247)

	return false, slen, s
}

func lpad(s string, length int, fill string) (string, error) {
	__antithesis_instrumentation__.Notify(599256)
	if length > maxAllocatedStringSize {
		__antithesis_instrumentation__.Notify(599260)
		return "", errStringTooLarge
	} else {
		__antithesis_instrumentation__.Notify(599261)
	}
	__antithesis_instrumentation__.Notify(599257)
	ok, slen, ret := padMaybeTruncate(s, length, fill)
	if ok {
		__antithesis_instrumentation__.Notify(599262)
		return ret, nil
	} else {
		__antithesis_instrumentation__.Notify(599263)
	}
	__antithesis_instrumentation__.Notify(599258)
	var buf strings.Builder
	fillRunes := []rune(fill)
	for i := 0; i < length-slen; i++ {
		__antithesis_instrumentation__.Notify(599264)
		buf.WriteRune(fillRunes[i%len(fillRunes)])
	}
	__antithesis_instrumentation__.Notify(599259)
	buf.WriteString(s)

	return buf.String(), nil
}

func rpad(s string, length int, fill string) (string, error) {
	__antithesis_instrumentation__.Notify(599265)
	if length > maxAllocatedStringSize {
		__antithesis_instrumentation__.Notify(599269)
		return "", errStringTooLarge
	} else {
		__antithesis_instrumentation__.Notify(599270)
	}
	__antithesis_instrumentation__.Notify(599266)
	ok, slen, ret := padMaybeTruncate(s, length, fill)
	if ok {
		__antithesis_instrumentation__.Notify(599271)
		return ret, nil
	} else {
		__antithesis_instrumentation__.Notify(599272)
	}
	__antithesis_instrumentation__.Notify(599267)
	var buf strings.Builder
	buf.WriteString(s)
	fillRunes := []rune(fill)
	for i := 0; i < length-slen; i++ {
		__antithesis_instrumentation__.Notify(599273)
		buf.WriteRune(fillRunes[i%len(fillRunes)])
	}
	__antithesis_instrumentation__.Notify(599268)

	return buf.String(), nil
}

func CleanEncodingName(s string) string {
	__antithesis_instrumentation__.Notify(599274)
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		__antithesis_instrumentation__.Notify(599276)
		c := s[i]
		if c >= 'A' && func() bool {
			__antithesis_instrumentation__.Notify(599277)
			return c <= 'Z' == true
		}() == true {
			__antithesis_instrumentation__.Notify(599278)
			b = append(b, c-'A'+'a')
		} else {
			__antithesis_instrumentation__.Notify(599279)
			if (c >= 'a' && func() bool {
				__antithesis_instrumentation__.Notify(599280)
				return c <= 'z' == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(599281)
				return (c >= '0' && func() bool {
					__antithesis_instrumentation__.Notify(599282)
					return c <= '9' == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(599283)
				b = append(b, c)
			} else {
				__antithesis_instrumentation__.Notify(599284)
			}
		}
	}
	__antithesis_instrumentation__.Notify(599275)
	return string(b)
}

var errInsufficientPriv = pgerror.New(
	pgcode.InsufficientPrivilege, "insufficient privilege",
)

var EvalFollowerReadOffset func(logicalClusterID uuid.UUID, _ *cluster.Settings) (time.Duration, error)

func recentTimestamp(ctx *tree.EvalContext) (time.Time, error) {
	__antithesis_instrumentation__.Notify(599285)
	if EvalFollowerReadOffset == nil {
		__antithesis_instrumentation__.Notify(599288)
		telemetry.Inc(sqltelemetry.FollowerReadDisabledCCLCounter)
		ctx.ClientNoticeSender.BufferClientNotice(
			ctx.Context,
			pgnotice.Newf("follower reads disabled because you are running a non-CCL distribution"),
		)
		return ctx.StmtTimestamp.Add(defaultFollowerReadDuration), nil
	} else {
		__antithesis_instrumentation__.Notify(599289)
	}
	__antithesis_instrumentation__.Notify(599286)
	offset, err := EvalFollowerReadOffset(ctx.ClusterID, ctx.Settings)
	if err != nil {
		__antithesis_instrumentation__.Notify(599290)
		if code := pgerror.GetPGCode(err); code == pgcode.CCLValidLicenseRequired {
			__antithesis_instrumentation__.Notify(599292)
			telemetry.Inc(sqltelemetry.FollowerReadDisabledNoEnterpriseLicense)
			ctx.ClientNoticeSender.BufferClientNotice(
				ctx.Context, pgnotice.Newf("follower reads disabled: %s", err.Error()),
			)
			return ctx.StmtTimestamp.Add(defaultFollowerReadDuration), nil
		} else {
			__antithesis_instrumentation__.Notify(599293)
		}
		__antithesis_instrumentation__.Notify(599291)
		return time.Time{}, err
	} else {
		__antithesis_instrumentation__.Notify(599294)
	}
	__antithesis_instrumentation__.Notify(599287)
	return ctx.StmtTimestamp.Add(offset), nil
}

func requireNonNull(d tree.Datum) error {
	__antithesis_instrumentation__.Notify(599295)
	if d == tree.DNull {
		__antithesis_instrumentation__.Notify(599297)
		return errInvalidNull
	} else {
		__antithesis_instrumentation__.Notify(599298)
	}
	__antithesis_instrumentation__.Notify(599296)
	return nil
}

func followerReadTimestamp(ctx *tree.EvalContext, _ tree.Datums) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599299)
	ts, err := recentTimestamp(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(599301)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599302)
	}
	__antithesis_instrumentation__.Notify(599300)
	return tree.MakeDTimestampTZ(ts, time.Microsecond)
}

var (
	WithMinTimestamp = func(ctx *tree.EvalContext, t time.Time) (time.Time, error) {
		__antithesis_instrumentation__.Notify(599303)
		return time.Time{}, pgerror.Newf(
			pgcode.CCLRequired,
			"%s can only be used with a CCL distribution",
			tree.WithMinTimestampFunctionName,
		)
	}

	WithMaxStaleness = func(ctx *tree.EvalContext, d duration.Duration) (time.Time, error) {
		__antithesis_instrumentation__.Notify(599304)
		return time.Time{}, pgerror.Newf(
			pgcode.CCLRequired,
			"%s can only be used with a CCL distribution",
			tree.WithMaxStalenessFunctionName,
		)
	}
)

const nearestOnlyInfo = `

If nearest_only is set to true, reads that cannot be served using the nearest
available replica will error.
`

func withMinTimestampInfo(nearestOnly bool) string {
	__antithesis_instrumentation__.Notify(599305)
	var nearestOnlyText string
	if nearestOnly {
		__antithesis_instrumentation__.Notify(599307)
		nearestOnlyText = nearestOnlyInfo
	} else {
		__antithesis_instrumentation__.Notify(599308)
	}
	__antithesis_instrumentation__.Notify(599306)
	return fmt.Sprintf(
		`When used in the AS OF SYSTEM TIME clause of an single-statement,
read-only transaction, CockroachDB chooses the newest timestamp before the min_timestamp
that allows execution of the reads at the nearest available replica without blocking.%s

Note this function requires an enterprise license on a CCL distribution.`,
		nearestOnlyText,
	)
}

func withMaxStalenessInfo(nearestOnly bool) string {
	__antithesis_instrumentation__.Notify(599309)
	var nearestOnlyText string
	if nearestOnly {
		__antithesis_instrumentation__.Notify(599311)
		nearestOnlyText = nearestOnlyInfo
	} else {
		__antithesis_instrumentation__.Notify(599312)
	}
	__antithesis_instrumentation__.Notify(599310)
	return fmt.Sprintf(
		`When used in the AS OF SYSTEM TIME clause of an single-statement,
read-only transaction, CockroachDB chooses the newest timestamp within the staleness
bound that allows execution of the reads at the nearest available replica without blocking.%s

Note this function requires an enterprise license on a CCL distribution.`,
		nearestOnlyText,
	)
}

func withMinTimestamp(ctx *tree.EvalContext, t time.Time) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599313)
	t, err := WithMinTimestamp(ctx, t)
	if err != nil {
		__antithesis_instrumentation__.Notify(599315)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599316)
	}
	__antithesis_instrumentation__.Notify(599314)
	return tree.MakeDTimestampTZ(t, time.Microsecond)
}

func withMaxStaleness(ctx *tree.EvalContext, d duration.Duration) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599317)
	t, err := WithMaxStaleness(ctx, d)
	if err != nil {
		__antithesis_instrumentation__.Notify(599319)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599320)
	}
	__antithesis_instrumentation__.Notify(599318)
	return tree.MakeDTimestampTZ(t, time.Microsecond)
}

func jsonNumInvertedIndexEntries(_ *tree.EvalContext, val tree.Datum) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599321)
	if val == tree.DNull {
		__antithesis_instrumentation__.Notify(599324)
		return tree.DZero, nil
	} else {
		__antithesis_instrumentation__.Notify(599325)
	}
	__antithesis_instrumentation__.Notify(599322)
	n, err := json.NumInvertedIndexEntries(tree.MustBeDJSON(val).JSON)
	if err != nil {
		__antithesis_instrumentation__.Notify(599326)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599327)
	}
	__antithesis_instrumentation__.Notify(599323)
	return tree.NewDInt(tree.DInt(n)), nil
}

func arrayNumInvertedIndexEntries(
	ctx *tree.EvalContext, val, version tree.Datum,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(599328)
	if val == tree.DNull {
		__antithesis_instrumentation__.Notify(599332)
		return tree.DZero, nil
	} else {
		__antithesis_instrumentation__.Notify(599333)
	}
	__antithesis_instrumentation__.Notify(599329)
	arr := tree.MustBeDArray(val)

	v := descpb.EmptyArraysInInvertedIndexesVersion
	if version != tree.DNull {
		__antithesis_instrumentation__.Notify(599334)
		v = descpb.IndexDescriptorVersion(tree.MustBeDInt(version))
	} else {
		__antithesis_instrumentation__.Notify(599335)
	}
	__antithesis_instrumentation__.Notify(599330)

	keys, err := rowenc.EncodeInvertedIndexTableKeys(arr, nil, v)
	if err != nil {
		__antithesis_instrumentation__.Notify(599336)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599337)
	}
	__antithesis_instrumentation__.Notify(599331)
	return tree.NewDInt(tree.DInt(len(keys))), nil
}

func parseContextFromDateStyle(
	ctx *tree.EvalContext, dateStyleStr string,
) (tree.ParseTimeContext, error) {
	__antithesis_instrumentation__.Notify(599338)
	ds, err := pgdate.ParseDateStyle(dateStyleStr, pgdate.DefaultDateStyle())
	if err != nil {
		__antithesis_instrumentation__.Notify(599341)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(599342)
	}
	__antithesis_instrumentation__.Notify(599339)
	if ds.Style != pgdate.Style_ISO {
		__antithesis_instrumentation__.Notify(599343)
		return nil, unimplemented.NewWithIssue(41773, "only ISO style is supported")
	} else {
		__antithesis_instrumentation__.Notify(599344)
	}
	__antithesis_instrumentation__.Notify(599340)
	return tree.NewParseTimeContext(
		ctx.GetTxnTimestamp(time.Microsecond).Time,
		tree.NewParseTimeContextOptionDateStyle(ds),
	), nil
}

func prettyStatementCustomConfig(
	stmt string, lineWidth int, alignMode int, caseSetting int,
) (string, error) {
	__antithesis_instrumentation__.Notify(599345)
	cfg := tree.DefaultPrettyCfg()
	cfg.LineWidth = lineWidth
	cfg.Align = tree.PrettyAlignMode(alignMode)
	caseMode := tree.CaseMode(caseSetting)
	if caseMode == tree.LowerCase {
		__antithesis_instrumentation__.Notify(599347)
		cfg.Case = func(str string) string { __antithesis_instrumentation__.Notify(599348); return strings.ToLower(str) }
	} else {
		__antithesis_instrumentation__.Notify(599349)
		if caseMode == tree.UpperCase {
			__antithesis_instrumentation__.Notify(599350)
			cfg.Case = func(str string) string { __antithesis_instrumentation__.Notify(599351); return strings.ToUpper(str) }
		} else {
			__antithesis_instrumentation__.Notify(599352)
		}
	}
	__antithesis_instrumentation__.Notify(599346)
	return prettyStatement(cfg, stmt)
}

func prettyStatement(p tree.PrettyCfg, stmt string) (string, error) {
	__antithesis_instrumentation__.Notify(599353)
	stmts, err := parser.Parse(stmt)
	if err != nil {
		__antithesis_instrumentation__.Notify(599356)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(599357)
	}
	__antithesis_instrumentation__.Notify(599354)
	var formattedStmt strings.Builder
	for idx := range stmts {
		__antithesis_instrumentation__.Notify(599358)
		formattedStmt.WriteString(p.Pretty(stmts[idx].AST))
		if len(stmts) > 1 {
			__antithesis_instrumentation__.Notify(599360)
			formattedStmt.WriteString(";")
		} else {
			__antithesis_instrumentation__.Notify(599361)
		}
		__antithesis_instrumentation__.Notify(599359)
		formattedStmt.WriteString("\n")
	}
	__antithesis_instrumentation__.Notify(599355)
	return formattedStmt.String(), nil
}
