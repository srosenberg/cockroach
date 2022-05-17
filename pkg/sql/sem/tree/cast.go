package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/lex"
	"github.com/cockroachdb/cockroach/pkg/sql/oidext"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/bitarray"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type CastContext uint8

const (
	_ CastContext = iota

	CastContextExplicit

	CastContextAssignment

	CastContextImplicit
)

func (cc CastContext) String() string {
	__antithesis_instrumentation__.Notify(603603)
	switch cc {
	case CastContextExplicit:
		__antithesis_instrumentation__.Notify(603604)
		return "explicit"
	case CastContextAssignment:
		__antithesis_instrumentation__.Notify(603605)
		return "assignment"
	case CastContextImplicit:
		__antithesis_instrumentation__.Notify(603606)
		return "implicit"
	default:
		__antithesis_instrumentation__.Notify(603607)
		return "invalid"
	}
}

type contextOrigin uint8

const (
	_ contextOrigin = iota

	contextOriginPgCast

	contextOriginAutomaticIOConversion

	contextOriginLegacyConversion
)

type cast struct {
	maxContext CastContext

	origin contextOrigin

	volatility Volatility

	volatilityHint string

	intervalStyleAffected bool

	dateStyleAffected bool
}

var castMap = map[oid.Oid]map[oid.Oid]cast{
	oid.T_bit: {
		oid.T_bit:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:   {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:   {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_bool: {
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float4:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:    {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_char: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oidext.T_box2d: {
		oidext.T_geometry: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_bpchar: {
		oid.T_bpchar:  {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bit:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_box2d: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bytea:    {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_date: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "CHAR to DATE casts depend on session DateStyle; use parse_date(string) instead",
			dateStyleAffected: true,
		},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_inet:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_interval: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,

			volatility:            VolatilityImmutable,
			volatilityHint:        "CHAR to INTERVAL casts depend on session IntervalStyle; use parse_interval(string) instead",
			intervalStyleAffected: true,
		},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_record:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_time: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "CHAR to TIME casts depend on session DateStyle; use parse_time(string) instead",
			dateStyleAffected: true,
		},
		oid.T_timestamp: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "CHAR to TIMESTAMP casts are context-dependent because of relative timestamp strings " +
				"like 'now' and session settings such as DateStyle; use parse_timestamp(string) instead.",
			dateStyleAffected: true,
		},
		oid.T_timestamptz: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
		},
		oid.T_timetz: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "CHAR to TIMETZ casts depend on session DateStyle; use parse_timetz(char) instead",
			dateStyleAffected: true,
		},
		oid.T_uuid:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_void:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_bytea: {
		oidext.T_geography: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_uuid:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_char: {
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:    {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_name: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},

		oid.T_bit:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_box2d: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bytea:    {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_date: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    `"char" to DATE casts depend on session DateStyle; use parse_date(string) instead`,
			dateStyleAffected: true,
		},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_inet:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_interval: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,

			volatility:            VolatilityImmutable,
			volatilityHint:        `"char" to INTERVAL casts depend on session IntervalStyle; use parse_interval(string) instead`,
			intervalStyleAffected: true,
		},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_record:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_time: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    `"char" to TIME casts depend on session DateStyle; use parse_time(string) instead`,
			dateStyleAffected: true,
		},
		oid.T_timestamp: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: `"char" to TIMESTAMP casts are context-dependent because of relative timestamp strings ` +
				"like 'now' and session settings such as DateStyle; use parse_timestamp(string) instead.",
			dateStyleAffected: true,
		},
		oid.T_timestamptz: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
		},
		oid.T_timetz: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    `"char" to TIMETZ casts depend on session DateStyle; use parse_timetz(string) instead`,
			dateStyleAffected: true,
		},
		oid.T_uuid:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_void:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_date: {
		oid.T_float4:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int8:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_timestamp:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamptz: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityStable},

		oid.T_bpchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,

			volatility: VolatilityImmutable,
			volatilityHint: "DATE to CHAR casts are dependent on DateStyle; consider " +
				"using to_char(date) instead.",
			dateStyleAffected: true,
		},
		oid.T_char: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,

			volatility: VolatilityImmutable,
			volatilityHint: `DATE to "char" casts are dependent on DateStyle; consider ` +
				"using to_char(date) instead.",
			dateStyleAffected: true,
		},
		oid.T_name: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,

			volatility: VolatilityImmutable,
			volatilityHint: "DATE to NAME casts are dependent on DateStyle; consider " +
				"using to_char(date) instead.",
			dateStyleAffected: true,
		},
		oid.T_text: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,

			volatility: VolatilityImmutable,
			volatilityHint: "DATE to STRING casts are dependent on DateStyle; consider " +
				"using to_char(date) instead.",
			dateStyleAffected: true,
		},
		oid.T_varchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,

			volatility: VolatilityImmutable,
			volatilityHint: "DATE to VARCHAR casts are dependent on DateStyle; consider " +
				"using to_char(date) instead.",
			dateStyleAffected: true,
		},
	},
	oid.T_float4: {
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_float8: {
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float4:   {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oidext.T_geography: {
		oid.T_bytea:        {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oidext.T_geometry: {
		oidext.T_box2d:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_bytea:        {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_text:         {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_inet: {
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_char: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_int2: {
		oid.T_bit:          {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_bool:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_date:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float4:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regnamespace: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regproc:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regprocedure: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regrole:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regtype:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamp:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_timestamptz:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_varbit:       {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_int4: {
		oid.T_bit:          {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_bool:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_char:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_date:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float4:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regnamespace: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regproc:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regprocedure: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regrole:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regtype:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamp:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_timestamptz:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_varbit:       {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_int8: {
		oid.T_bit:          {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_bool:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_date:         {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float4:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regnamespace: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regproc:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regprocedure: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regrole:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regtype:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamp:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_timestamptz:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_varbit:       {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_interval: {
		oid.T_float4:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int8:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_interval: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_numeric:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_time:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bpchar: {
			maxContext:            CastContextAssignment,
			origin:                contextOriginAutomaticIOConversion,
			volatility:            VolatilityImmutable,
			volatilityHint:        "INTERVAL to CHAR casts depend on IntervalStyle; consider using to_char(interval)",
			intervalStyleAffected: true,
		},
		oid.T_char: {
			maxContext:            CastContextAssignment,
			origin:                contextOriginAutomaticIOConversion,
			volatility:            VolatilityImmutable,
			volatilityHint:        `INTERVAL to "char" casts depend on IntervalStyle; consider using to_char(interval)`,
			intervalStyleAffected: true,
		},
		oid.T_name: {
			maxContext:            CastContextAssignment,
			origin:                contextOriginAutomaticIOConversion,
			volatility:            VolatilityImmutable,
			volatilityHint:        "INTERVAL to NAME casts depend on IntervalStyle; consider using to_char(interval)",
			intervalStyleAffected: true,
		},
		oid.T_text: {
			maxContext:            CastContextAssignment,
			origin:                contextOriginAutomaticIOConversion,
			volatility:            VolatilityImmutable,
			volatilityHint:        "INTERVAL to STRING casts depend on IntervalStyle; consider using to_char(interval)",
			intervalStyleAffected: true,
		},
		oid.T_varchar: {
			maxContext:            CastContextAssignment,
			origin:                contextOriginAutomaticIOConversion,
			volatility:            VolatilityImmutable,
			volatilityHint:        "INTERVAL to VARCHAR casts depend on IntervalStyle; consider using to_char(interval)",
			intervalStyleAffected: true,
		},
	},
	oid.T_jsonb: {
		oid.T_bool:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_name: {
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_char: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},

		oid.T_bit:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_box2d: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bytea:    {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_date: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "NAME to DATE casts depend on session DateStyle; use parse_date(string) instead",
			dateStyleAffected: true,
		},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_inet:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_interval: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,

			volatility:            VolatilityImmutable,
			volatilityHint:        "NAME to INTERVAL casts depend on session IntervalStyle; use parse_interval(string) instead",
			intervalStyleAffected: true,
		},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_record:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_time: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "NAME to TIME casts depend on session DateStyle; use parse_time(string) instead",
			dateStyleAffected: true,
		},
		oid.T_timestamp: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "NAME to TIMESTAMP casts are context-dependent because of relative timestamp strings " +
				"like 'now' and session settings such as DateStyle; use parse_timestamp(string) instead.",
			dateStyleAffected: true,
		},
		oid.T_timestamptz: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
		},
		oid.T_timetz: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "NAME to TIMETZ casts depend on session DateStyle; use parse_timetz(string) instead",
			dateStyleAffected: true,
		},
		oid.T_uuid:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_void:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_numeric: {
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float4:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float8:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int4:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_interval: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:  {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_oid: {

		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regnamespace: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regproc:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regprocedure: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regrole:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regtype:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_record: {

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regclass: {

		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regnamespace: {

		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regproc: {

		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regprocedure: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regprocedure: {

		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regproc:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regrole: {

		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_regtype: {

		oid.T_int2:         {maxContext: CastContextAssignment, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityStable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
	},
	oid.T_text: {
		oid.T_bpchar:      {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_char:        {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oidext.T_geometry: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_name:        {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityStable},

		oid.T_text:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bit:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_box2d: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bytea:    {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_date: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "STRING to DATE casts depend on session DateStyle; use parse_date(string) instead",
			dateStyleAffected: true,
		},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_inet:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_interval: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,

			volatility:            VolatilityImmutable,
			volatilityHint:        "STRING to INTERVAL casts depend on session IntervalStyle; use parse_interval(string) instead",
			intervalStyleAffected: true,
		},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_record:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_time: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "STRING to TIME casts depend on session DateStyle; use parse_time(string) instead",
			dateStyleAffected: true,
		},
		oid.T_timestamp: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "STRING to TIMESTAMP casts are context-dependent because of relative timestamp strings " +
				"like 'now' and session settings such as DateStyle; use parse_timestamp(string) instead.",
			dateStyleAffected: true,
		},
		oid.T_timestamptz: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
		},
		oid.T_timetz: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "STRING to TIMETZ casts depend on session DateStyle; use parse_timetz(string) instead",
			dateStyleAffected: true,
		},
		oid.T_uuid:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_void:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_time: {
		oid.T_interval: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_time:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timetz:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityStable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_timestamp: {
		oid.T_date:        {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_float4:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:      {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int8:        {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric:     {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_time:        {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamp:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timestamptz: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityStable},

		oid.T_bpchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,

			volatility: VolatilityImmutable,
			volatilityHint: "TIMESTAMP to CHAR casts are dependent on DateStyle; consider " +
				"using to_char(timestamp) instead.",
			dateStyleAffected: true,
		},
		oid.T_char: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,

			volatility: VolatilityImmutable,
			volatilityHint: `TIMESTAMP to "char" casts are dependent on DateStyle; consider ` +
				"using to_char(timestamp) instead.",
			dateStyleAffected: true,
		},
		oid.T_name: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,

			volatility: VolatilityImmutable,
			volatilityHint: "TIMESTAMP to NAME casts are dependent on DateStyle; consider " +
				"using to_char(timestamp) instead.",
			dateStyleAffected: true,
		},
		oid.T_text: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,

			volatility: VolatilityImmutable,
			volatilityHint: "TIMESTAMP to STRING casts are dependent on DateStyle; consider " +
				"using to_char(timestamp) instead.",
			dateStyleAffected: true,
		},
		oid.T_varchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,

			volatility: VolatilityImmutable,
			volatilityHint: "TIMESTAMP to VARCHAR casts are dependent on DateStyle; consider " +
				"using to_char(timestamp) instead.",
			dateStyleAffected: true,
		},
	},
	oid.T_timestamptz: {
		oid.T_date:    {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityStable},
		oid.T_float4:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_float8:  {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int2:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int8:    {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_numeric: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_time:    {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityStable},
		oid.T_timestamp: {
			maxContext:     CastContextAssignment,
			origin:         contextOriginPgCast,
			volatility:     VolatilityStable,
			volatilityHint: "TIMESTAMPTZ to TIMESTAMP casts depend on the current timezone; consider using AT TIME ZONE 'UTC' instead",
		},
		oid.T_timestamptz: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timetz:      {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityStable},

		oid.T_bpchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "TIMESTAMPTZ to CHAR casts depend on the current timezone; consider " +
				"using to_char(t AT TIME ZONE 'UTC') instead.",
		},
		oid.T_char: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: `TIMESTAMPTZ to "char" casts depend on the current timezone; consider ` +
				"using to_char(t AT TIME ZONE 'UTC') instead.",
		},
		oid.T_name: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "TIMESTAMPTZ to NAME casts depend on the current timezone; consider " +
				"using to_char(t AT TIME ZONE 'UTC') instead.",
		},
		oid.T_text: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "TIMESTAMPTZ to STRING casts depend on the current timezone; consider " +
				"using to_char(t AT TIME ZONE 'UTC') instead.",
		},
		oid.T_varchar: {
			maxContext: CastContextAssignment,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "TIMESTAMPTZ to VARCHAR casts depend on the current timezone; consider " +
				"using to_char(t AT TIME ZONE 'UTC') instead.",
		},
	},
	oid.T_timetz: {
		oid.T_time:   {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_timetz: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_uuid: {
		oid.T_bytea: {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_varbit: {
		oid.T_bit:    {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_int2:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int4:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_int8:   {maxContext: CastContextExplicit, origin: contextOriginLegacyConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_varchar: {
		oid.T_bpchar:   {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_char:     {maxContext: CastContextAssignment, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_name:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_regclass: {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityStable},
		oid.T_text:     {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},
		oid.T_varchar:  {maxContext: CastContextImplicit, origin: contextOriginPgCast, volatility: VolatilityImmutable},

		oid.T_bit:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bool:     {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_box2d: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_bytea:    {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_date: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "VARCHAR to DATE casts depend on session DateStyle; use parse_date(string) instead",
			dateStyleAffected: true,
		},
		oid.T_float4:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_float8:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geography: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oidext.T_geometry:  {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_inet:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int2:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int4:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_int8:         {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_interval: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,

			volatility:            VolatilityImmutable,
			volatilityHint:        "VARCHAR to INTERVAL casts depend on session IntervalStyle; use parse_interval(string) instead",
			intervalStyleAffected: true,
		},
		oid.T_jsonb:        {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_numeric:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_oid:          {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_record:       {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regnamespace: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regproc:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regprocedure: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regrole:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_regtype:      {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityStable},
		oid.T_time: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "VARCHAR to TIME casts depend on session DateStyle; use parse_time(string) instead",
			dateStyleAffected: true,
		},
		oid.T_timestamp: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
			volatilityHint: "VARCHAR to TIMESTAMP casts are context-dependent because of relative timestamp strings " +
				"like 'now' and session settings such as DateStyle; use parse_timestamp(string) instead.",
			dateStyleAffected: true,
		},
		oid.T_timestamptz: {
			maxContext: CastContextExplicit,
			origin:     contextOriginAutomaticIOConversion,
			volatility: VolatilityStable,
		},
		oid.T_timetz: {
			maxContext:        CastContextExplicit,
			origin:            contextOriginAutomaticIOConversion,
			volatility:        VolatilityStable,
			volatilityHint:    "VARCHAR to TIMETZ casts depend on session DateStyle; use parse_timetz(string) instead",
			dateStyleAffected: true,
		},
		oid.T_uuid:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varbit: {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_void:   {maxContext: CastContextExplicit, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
	oid.T_void: {
		oid.T_bpchar:  {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_char:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_name:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_text:    {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
		oid.T_varchar: {maxContext: CastContextAssignment, origin: contextOriginAutomaticIOConversion, volatility: VolatilityImmutable},
	},
}

func init() {
	var stringTypes = [...]oid.Oid{
		oid.T_bpchar,
		oid.T_name,
		oid.T_char,
		oid.T_varchar,
		oid.T_text,
	}
	isStringType := func(o oid.Oid) bool {
		for _, strOid := range stringTypes {
			if o == strOid {
				return true
			}
		}
		return false
	}

	typeName := func(o oid.Oid) string {
		if name, ok := oidext.TypeName(o); ok {
			return name
		}
		panic(errors.AssertionFailedf("no type name for oid %d", o))
	}

	for _, strType := range stringTypes {
		for otherType := range castMap {
			if strType == otherType {
				continue
			}
			strTypeName := typeName(strType)
			otherTypeName := typeName(otherType)
			if _, from := castMap[strType][otherType]; !from && otherType != oid.T_unknown {
				panic(errors.AssertionFailedf("there must be a cast from %s to %s", strTypeName, otherTypeName))
			}
			if _, to := castMap[otherType][strType]; !to {
				panic(errors.AssertionFailedf("there must be a cast from %s to %s", otherTypeName, strTypeName))
			}
		}
	}

	for src, tgts := range castMap {
		for tgt, ent := range tgts {
			srcStr := typeName(src)
			tgtStr := typeName(tgt)

			if ent.maxContext == CastContext(0) {
				panic(errors.AssertionFailedf("cast from %s to %s has no maxContext set", srcStr, tgtStr))
			}
			if ent.origin == contextOrigin(0) {
				panic(errors.AssertionFailedf("cast from %s to %s has no origin set", srcStr, tgtStr))
			}

			if src == tgt {
				if ent.maxContext != CastContextImplicit {
					panic(errors.AssertionFailedf(
						"cast from %s to %s must be an implicit cast",
						srcStr, tgtStr,
					))
				}
			}

			if isStringType(tgt) && ent.origin == contextOriginAutomaticIOConversion &&
				ent.maxContext != CastContextAssignment {
				panic(errors.AssertionFailedf(
					"automatic conversion from %s to %s must be an assignment cast",
					srcStr, tgtStr,
				))
			}

			if isStringType(src) && !isStringType(tgt) && ent.origin == contextOriginAutomaticIOConversion &&
				ent.maxContext != CastContextExplicit {
				panic(errors.AssertionFailedf(
					"automatic conversion from %s to %s must be an explicit cast",
					srcStr, tgtStr,
				))
			}
		}
	}
}

func ForEachCast(fn func(src, tgt oid.Oid)) {
	__antithesis_instrumentation__.Notify(603608)
	for src, tgts := range castMap {
		__antithesis_instrumentation__.Notify(603609)
		for tgt := range tgts {
			__antithesis_instrumentation__.Notify(603610)
			fn(src, tgt)
		}
	}
}

func ValidCast(src, tgt *types.T, ctx CastContext) bool {
	__antithesis_instrumentation__.Notify(603611)
	srcFamily := src.Family()
	tgtFamily := tgt.Family()

	if srcFamily == types.ArrayFamily && func() bool {
		__antithesis_instrumentation__.Notify(603615)
		return tgtFamily == types.ArrayFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(603616)
		return ValidCast(src.ArrayContents(), tgt.ArrayContents(), ctx)
	} else {
		__antithesis_instrumentation__.Notify(603617)
	}
	__antithesis_instrumentation__.Notify(603612)

	if srcFamily == types.TupleFamily && func() bool {
		__antithesis_instrumentation__.Notify(603618)
		return tgtFamily == types.TupleFamily == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603619)
		return tgt != types.AnyTuple == true
	}() == true {
		__antithesis_instrumentation__.Notify(603620)
		srcTypes := src.TupleContents()
		tgtTypes := tgt.TupleContents()

		if len(srcTypes) != len(tgtTypes) {
			__antithesis_instrumentation__.Notify(603623)
			return false
		} else {
			__antithesis_instrumentation__.Notify(603624)
		}
		__antithesis_instrumentation__.Notify(603621)
		for i := range srcTypes {
			__antithesis_instrumentation__.Notify(603625)
			if ok := ValidCast(srcTypes[i], tgtTypes[i], ctx); !ok {
				__antithesis_instrumentation__.Notify(603626)
				return false
			} else {
				__antithesis_instrumentation__.Notify(603627)
			}
		}
		__antithesis_instrumentation__.Notify(603622)
		return true
	} else {
		__antithesis_instrumentation__.Notify(603628)
	}
	__antithesis_instrumentation__.Notify(603613)

	c, ok := lookupCast(src, tgt, false, false)
	if ok {
		__antithesis_instrumentation__.Notify(603629)
		return c.maxContext >= ctx
	} else {
		__antithesis_instrumentation__.Notify(603630)
	}
	__antithesis_instrumentation__.Notify(603614)

	return false
}

func lookupCast(src, tgt *types.T, intervalStyleEnabled, dateStyleEnabled bool) (cast, bool) {
	__antithesis_instrumentation__.Notify(603631)
	srcFamily := src.Family()
	tgtFamily := tgt.Family()
	srcFamily.Name()

	if srcFamily == types.UnknownFamily {
		__antithesis_instrumentation__.Notify(603641)
		return cast{
			maxContext: CastContextImplicit,
			volatility: VolatilityImmutable,
		}, true
	} else {
		__antithesis_instrumentation__.Notify(603642)
	}
	__antithesis_instrumentation__.Notify(603632)

	if srcFamily == types.EnumFamily && func() bool {
		__antithesis_instrumentation__.Notify(603643)
		return tgtFamily == types.StringFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(603644)

		return cast{
			maxContext: CastContextAssignment,
			volatility: VolatilityImmutable,
		}, true
	} else {
		__antithesis_instrumentation__.Notify(603645)
	}
	__antithesis_instrumentation__.Notify(603633)
	if tgtFamily == types.EnumFamily {
		__antithesis_instrumentation__.Notify(603646)
		switch srcFamily {
		case types.StringFamily:
			__antithesis_instrumentation__.Notify(603647)

			return cast{
				maxContext: CastContextExplicit,
				volatility: VolatilityImmutable,
			}, true
		case types.UnknownFamily:
			__antithesis_instrumentation__.Notify(603648)

			return cast{
				maxContext: CastContextImplicit,
				volatility: VolatilityImmutable,
			}, true
		default:
			__antithesis_instrumentation__.Notify(603649)
		}
	} else {
		__antithesis_instrumentation__.Notify(603650)
	}
	__antithesis_instrumentation__.Notify(603634)

	if srcFamily == types.ArrayFamily && func() bool {
		__antithesis_instrumentation__.Notify(603651)
		return tgtFamily == types.StringFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(603652)
		return cast{
			maxContext: CastContextAssignment,
			volatility: VolatilityStable,
		}, true
	} else {
		__antithesis_instrumentation__.Notify(603653)
	}
	__antithesis_instrumentation__.Notify(603635)

	if srcFamily == types.TupleFamily && func() bool {
		__antithesis_instrumentation__.Notify(603654)
		return tgtFamily == types.StringFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(603655)
		return cast{
			maxContext: CastContextAssignment,
			volatility: VolatilityImmutable,
		}, true
	} else {
		__antithesis_instrumentation__.Notify(603656)
	}
	__antithesis_instrumentation__.Notify(603636)

	if srcFamily == types.TupleFamily && func() bool {
		__antithesis_instrumentation__.Notify(603657)
		return tgt == types.AnyTuple == true
	}() == true {
		__antithesis_instrumentation__.Notify(603658)
		return cast{
			maxContext: CastContextImplicit,
			volatility: VolatilityImmutable,
		}, true
	} else {
		__antithesis_instrumentation__.Notify(603659)
	}
	__antithesis_instrumentation__.Notify(603637)

	if srcFamily == types.StringFamily && func() bool {
		__antithesis_instrumentation__.Notify(603660)
		return (tgtFamily == types.ArrayFamily || func() bool {
			__antithesis_instrumentation__.Notify(603661)
			return tgtFamily == types.TupleFamily == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(603662)
		return cast{
			maxContext: CastContextExplicit,
			volatility: VolatilityStable,
		}, true
	} else {
		__antithesis_instrumentation__.Notify(603663)
	}
	__antithesis_instrumentation__.Notify(603638)

	if tgts, ok := castMap[src.Oid()]; ok {
		__antithesis_instrumentation__.Notify(603664)
		if c, ok := tgts[tgt.Oid()]; ok {
			__antithesis_instrumentation__.Notify(603665)
			if intervalStyleEnabled && func() bool {
				__antithesis_instrumentation__.Notify(603667)
				return c.intervalStyleAffected == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(603668)
				return (dateStyleEnabled && func() bool {
					__antithesis_instrumentation__.Notify(603669)
					return c.dateStyleAffected == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(603670)
				c.volatility = VolatilityStable
			} else {
				__antithesis_instrumentation__.Notify(603671)
			}
			__antithesis_instrumentation__.Notify(603666)
			return c, true
		} else {
			__antithesis_instrumentation__.Notify(603672)
		}
	} else {
		__antithesis_instrumentation__.Notify(603673)
	}
	__antithesis_instrumentation__.Notify(603639)

	if src.Oid() == tgt.Oid() {
		__antithesis_instrumentation__.Notify(603674)
		return cast{
			maxContext: CastContextImplicit,
			volatility: VolatilityImmutable,
		}, true
	} else {
		__antithesis_instrumentation__.Notify(603675)
	}
	__antithesis_instrumentation__.Notify(603640)

	return cast{}, false
}

func LookupCastVolatility(from, to *types.T, sd *sessiondata.SessionData) (_ Volatility, ok bool) {
	__antithesis_instrumentation__.Notify(603676)
	fromFamily := from.Family()
	toFamily := to.Family()

	if fromFamily == types.ArrayFamily && func() bool {
		__antithesis_instrumentation__.Notify(603681)
		return toFamily == types.ArrayFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(603682)
		return LookupCastVolatility(from.ArrayContents(), to.ArrayContents(), sd)
	} else {
		__antithesis_instrumentation__.Notify(603683)
	}
	__antithesis_instrumentation__.Notify(603677)

	if fromFamily == types.TupleFamily && func() bool {
		__antithesis_instrumentation__.Notify(603684)
		return toFamily == types.TupleFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(603685)
		fromTypes := from.TupleContents()
		toTypes := to.TupleContents()

		if len(toTypes) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(603689)
			return toTypes[0].Family() == types.AnyFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(603690)
			return VolatilityStable, true
		} else {
			__antithesis_instrumentation__.Notify(603691)
		}
		__antithesis_instrumentation__.Notify(603686)
		if len(fromTypes) != len(toTypes) {
			__antithesis_instrumentation__.Notify(603692)
			return 0, false
		} else {
			__antithesis_instrumentation__.Notify(603693)
		}
		__antithesis_instrumentation__.Notify(603687)
		maxVolatility := VolatilityLeakProof
		for i := range fromTypes {
			__antithesis_instrumentation__.Notify(603694)
			v, lookupOk := LookupCastVolatility(fromTypes[i], toTypes[i], sd)
			if !lookupOk {
				__antithesis_instrumentation__.Notify(603696)
				return 0, false
			} else {
				__antithesis_instrumentation__.Notify(603697)
			}
			__antithesis_instrumentation__.Notify(603695)
			if v > maxVolatility {
				__antithesis_instrumentation__.Notify(603698)
				maxVolatility = v
			} else {
				__antithesis_instrumentation__.Notify(603699)
			}
		}
		__antithesis_instrumentation__.Notify(603688)
		return maxVolatility, true
	} else {
		__antithesis_instrumentation__.Notify(603700)
	}
	__antithesis_instrumentation__.Notify(603678)

	intervalStyleEnabled := false
	dateStyleEnabled := false
	if sd != nil {
		__antithesis_instrumentation__.Notify(603701)
		intervalStyleEnabled = sd.IntervalStyleEnabled
		dateStyleEnabled = sd.DateStyleEnabled
	} else {
		__antithesis_instrumentation__.Notify(603702)
	}
	__antithesis_instrumentation__.Notify(603679)

	cast, ok := lookupCast(from, to, intervalStyleEnabled, dateStyleEnabled)
	if !ok {
		__antithesis_instrumentation__.Notify(603703)
		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(603704)
	}
	__antithesis_instrumentation__.Notify(603680)
	return cast.volatility, true
}

func PerformCast(ctx *EvalContext, d Datum, t *types.T) (Datum, error) {
	__antithesis_instrumentation__.Notify(603705)
	ret, err := performCastWithoutPrecisionTruncation(ctx, d, t, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(603707)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(603708)
	}
	__antithesis_instrumentation__.Notify(603706)
	return AdjustValueToType(t, ret)
}

func PerformAssignmentCast(ctx *EvalContext, d Datum, t *types.T) (Datum, error) {
	__antithesis_instrumentation__.Notify(603709)
	if !ValidCast(d.ResolvedType(), t, CastContextAssignment) {
		__antithesis_instrumentation__.Notify(603712)
		return nil, pgerror.Newf(
			pgcode.CannotCoerce,
			"invalid assignment cast: %s -> %s", d.ResolvedType(), t,
		)
	} else {
		__antithesis_instrumentation__.Notify(603713)
	}
	__antithesis_instrumentation__.Notify(603710)
	d, err := performCastWithoutPrecisionTruncation(ctx, d, t, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(603714)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(603715)
	}
	__antithesis_instrumentation__.Notify(603711)
	return AdjustValueToType(t, d)
}

func AdjustValueToType(typ *types.T, inVal Datum) (outVal Datum, err error) {
	__antithesis_instrumentation__.Notify(603716)
	switch typ.Family() {
	case types.StringFamily, types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(603718)
		var sv string
		if v, ok := AsDString(inVal); ok {
			__antithesis_instrumentation__.Notify(603733)
			sv = string(v)
		} else {
			__antithesis_instrumentation__.Notify(603734)
			if v, ok := inVal.(*DCollatedString); ok {
				__antithesis_instrumentation__.Notify(603735)
				sv = v.Contents
			} else {
				__antithesis_instrumentation__.Notify(603736)
			}
		}
		__antithesis_instrumentation__.Notify(603719)
		sv = adjustStringValueToType(typ, sv)
		if typ.Width() > 0 && func() bool {
			__antithesis_instrumentation__.Notify(603737)
			return utf8.RuneCountInString(sv) > int(typ.Width()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(603738)
			return nil, pgerror.Newf(pgcode.StringDataRightTruncation,
				"value too long for type %s",
				typ.SQLString())
		} else {
			__antithesis_instrumentation__.Notify(603739)
		}
		__antithesis_instrumentation__.Notify(603720)

		if typ.Oid() == oid.T_bpchar || func() bool {
			__antithesis_instrumentation__.Notify(603740)
			return typ.Oid() == oid.T_char == true
		}() == true {
			__antithesis_instrumentation__.Notify(603741)
			if _, ok := AsDString(inVal); ok {
				__antithesis_instrumentation__.Notify(603742)
				return NewDString(sv), nil
			} else {
				__antithesis_instrumentation__.Notify(603743)
				if _, ok := inVal.(*DCollatedString); ok {
					__antithesis_instrumentation__.Notify(603744)
					return NewDCollatedString(sv, typ.Locale(), &CollationEnvironment{})
				} else {
					__antithesis_instrumentation__.Notify(603745)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(603746)
		}
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(603721)
		if v, ok := AsDInt(inVal); ok {
			__antithesis_instrumentation__.Notify(603747)
			if typ.Width() == 32 || func() bool {
				__antithesis_instrumentation__.Notify(603748)
				return typ.Width() == 16 == true
			}() == true {
				__antithesis_instrumentation__.Notify(603749)

				width := uint(typ.Width() - 1)

				shifted := v >> width
				if (v >= 0 && func() bool {
					__antithesis_instrumentation__.Notify(603750)
					return shifted > 0 == true
				}() == true) || func() bool {
					__antithesis_instrumentation__.Notify(603751)
					return (v < 0 && func() bool {
						__antithesis_instrumentation__.Notify(603752)
						return shifted < -1 == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(603753)
					if typ.Width() == 16 {
						__antithesis_instrumentation__.Notify(603755)
						return nil, ErrInt2OutOfRange
					} else {
						__antithesis_instrumentation__.Notify(603756)
					}
					__antithesis_instrumentation__.Notify(603754)
					return nil, ErrInt4OutOfRange
				} else {
					__antithesis_instrumentation__.Notify(603757)
				}
			} else {
				__antithesis_instrumentation__.Notify(603758)
			}
		} else {
			__antithesis_instrumentation__.Notify(603759)
		}
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(603722)
		if v, ok := AsDBitArray(inVal); ok {
			__antithesis_instrumentation__.Notify(603760)
			if typ.Width() > 0 {
				__antithesis_instrumentation__.Notify(603761)
				bitLen := v.BitLen()
				switch typ.Oid() {
				case oid.T_varbit:
					__antithesis_instrumentation__.Notify(603762)
					if bitLen > uint(typ.Width()) {
						__antithesis_instrumentation__.Notify(603764)
						return nil, pgerror.Newf(pgcode.StringDataRightTruncation,
							"bit string length %d too large for type %s", bitLen, typ.SQLString())
					} else {
						__antithesis_instrumentation__.Notify(603765)
					}
				default:
					__antithesis_instrumentation__.Notify(603763)
					if bitLen != uint(typ.Width()) {
						__antithesis_instrumentation__.Notify(603766)
						return nil, pgerror.Newf(pgcode.StringDataLengthMismatch,
							"bit string length %d does not match type %s", bitLen, typ.SQLString())
					} else {
						__antithesis_instrumentation__.Notify(603767)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(603768)
			}
		} else {
			__antithesis_instrumentation__.Notify(603769)
		}
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(603723)
		if inDec, ok := inVal.(*DDecimal); ok {
			__antithesis_instrumentation__.Notify(603770)
			if inDec.Form != apd.Finite || func() bool {
				__antithesis_instrumentation__.Notify(603774)
				return typ.Precision() == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(603775)

				break
			} else {
				__antithesis_instrumentation__.Notify(603776)
			}
			__antithesis_instrumentation__.Notify(603771)
			if int64(typ.Precision()) >= inDec.NumDigits() && func() bool {
				__antithesis_instrumentation__.Notify(603777)
				return typ.Scale() == inDec.Exponent == true
			}() == true {
				__antithesis_instrumentation__.Notify(603778)

				break
			} else {
				__antithesis_instrumentation__.Notify(603779)
			}
			__antithesis_instrumentation__.Notify(603772)

			var outDec DDecimal
			outDec.Set(&inDec.Decimal)
			err := LimitDecimalWidth(&outDec.Decimal, int(typ.Precision()), int(typ.Scale()))
			if err != nil {
				__antithesis_instrumentation__.Notify(603780)
				return nil, errors.Wrapf(err, "type %s", typ.SQLString())
			} else {
				__antithesis_instrumentation__.Notify(603781)
			}
			__antithesis_instrumentation__.Notify(603773)
			return &outDec, nil
		} else {
			__antithesis_instrumentation__.Notify(603782)
		}
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(603724)
		if inArr, ok := inVal.(*DArray); ok {
			__antithesis_instrumentation__.Notify(603783)
			var outArr *DArray
			elementType := typ.ArrayContents()
			for i, inElem := range inArr.Array {
				__antithesis_instrumentation__.Notify(603785)
				outElem, err := AdjustValueToType(elementType, inElem)
				if err != nil {
					__antithesis_instrumentation__.Notify(603788)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(603789)
				}
				__antithesis_instrumentation__.Notify(603786)
				if outElem != inElem {
					__antithesis_instrumentation__.Notify(603790)
					if outArr == nil {
						__antithesis_instrumentation__.Notify(603791)
						outArr = &DArray{}
						*outArr = *inArr
						outArr.Array = make(Datums, len(inArr.Array))
						copy(outArr.Array, inArr.Array[:i])
					} else {
						__antithesis_instrumentation__.Notify(603792)
					}
				} else {
					__antithesis_instrumentation__.Notify(603793)
				}
				__antithesis_instrumentation__.Notify(603787)
				if outArr != nil {
					__antithesis_instrumentation__.Notify(603794)
					outArr.Array[i] = inElem
				} else {
					__antithesis_instrumentation__.Notify(603795)
				}
			}
			__antithesis_instrumentation__.Notify(603784)
			if outArr != nil {
				__antithesis_instrumentation__.Notify(603796)
				return outArr, nil
			} else {
				__antithesis_instrumentation__.Notify(603797)
			}
		} else {
			__antithesis_instrumentation__.Notify(603798)
		}
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(603725)
		if in, ok := inVal.(*DTime); ok {
			__antithesis_instrumentation__.Notify(603799)
			return in.Round(TimeFamilyPrecisionToRoundDuration(typ.Precision())), nil
		} else {
			__antithesis_instrumentation__.Notify(603800)
		}
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(603726)
		if in, ok := inVal.(*DTimestamp); ok {
			__antithesis_instrumentation__.Notify(603801)
			return in.Round(TimeFamilyPrecisionToRoundDuration(typ.Precision()))
		} else {
			__antithesis_instrumentation__.Notify(603802)
		}
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(603727)
		if in, ok := inVal.(*DTimestampTZ); ok {
			__antithesis_instrumentation__.Notify(603803)
			return in.Round(TimeFamilyPrecisionToRoundDuration(typ.Precision()))
		} else {
			__antithesis_instrumentation__.Notify(603804)
		}
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(603728)
		if in, ok := inVal.(*DTimeTZ); ok {
			__antithesis_instrumentation__.Notify(603805)
			return in.Round(TimeFamilyPrecisionToRoundDuration(typ.Precision())), nil
		} else {
			__antithesis_instrumentation__.Notify(603806)
		}
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(603729)
		if in, ok := inVal.(*DInterval); ok {
			__antithesis_instrumentation__.Notify(603807)
			itm, err := typ.IntervalTypeMetadata()
			if err != nil {
				__antithesis_instrumentation__.Notify(603809)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(603810)
			}
			__antithesis_instrumentation__.Notify(603808)
			return NewDInterval(in.Duration, itm), nil
		} else {
			__antithesis_instrumentation__.Notify(603811)
		}
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(603730)
		if in, ok := inVal.(*DGeometry); ok {
			__antithesis_instrumentation__.Notify(603812)
			if err := geo.SpatialObjectFitsColumnMetadata(
				in.Geometry.SpatialObject(),
				typ.InternalType.GeoMetadata.SRID,
				typ.InternalType.GeoMetadata.ShapeType,
			); err != nil {
				__antithesis_instrumentation__.Notify(603813)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(603814)
			}
		} else {
			__antithesis_instrumentation__.Notify(603815)
		}
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(603731)
		if in, ok := inVal.(*DGeography); ok {
			__antithesis_instrumentation__.Notify(603816)
			if err := geo.SpatialObjectFitsColumnMetadata(
				in.Geography.SpatialObject(),
				typ.InternalType.GeoMetadata.SRID,
				typ.InternalType.GeoMetadata.ShapeType,
			); err != nil {
				__antithesis_instrumentation__.Notify(603817)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(603818)
			}
		} else {
			__antithesis_instrumentation__.Notify(603819)
		}
	default:
		__antithesis_instrumentation__.Notify(603732)
	}
	__antithesis_instrumentation__.Notify(603717)
	return inVal, nil
}

func adjustStringValueToType(typ *types.T, sv string) string {
	__antithesis_instrumentation__.Notify(603820)
	switch typ.Oid() {
	case oid.T_char:
		__antithesis_instrumentation__.Notify(603822)

		return util.TruncateString(sv, 1)
	case oid.T_bpchar:
		__antithesis_instrumentation__.Notify(603823)

		return strings.TrimRight(sv, " ")
	default:
		__antithesis_instrumentation__.Notify(603824)
	}
	__antithesis_instrumentation__.Notify(603821)
	return sv
}

func formatBitArrayToType(d *DBitArray, t *types.T) *DBitArray {
	__antithesis_instrumentation__.Notify(603825)
	if t.Width() == 0 || func() bool {
		__antithesis_instrumentation__.Notify(603828)
		return d.BitLen() == uint(t.Width()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(603829)
		return d
	} else {
		__antithesis_instrumentation__.Notify(603830)
	}
	__antithesis_instrumentation__.Notify(603826)
	a := d.BitArray.Clone()
	switch t.Oid() {
	case oid.T_varbit:
		__antithesis_instrumentation__.Notify(603831)

		if uint(t.Width()) < a.BitLen() {
			__antithesis_instrumentation__.Notify(603833)
			a = a.ToWidth(uint(t.Width()))
		} else {
			__antithesis_instrumentation__.Notify(603834)
		}
	default:
		__antithesis_instrumentation__.Notify(603832)
		a = a.ToWidth(uint(t.Width()))
	}
	__antithesis_instrumentation__.Notify(603827)
	return &DBitArray{a}
}

func performCastWithoutPrecisionTruncation(
	ctx *EvalContext, d Datum, t *types.T, truncateWidth bool,
) (Datum, error) {
	__antithesis_instrumentation__.Notify(603835)

	if d == DNull {
		__antithesis_instrumentation__.Notify(603838)
		return d, nil
	} else {
		__antithesis_instrumentation__.Notify(603839)
	}
	__antithesis_instrumentation__.Notify(603836)

	d = UnwrapDatum(nil, d)
	switch t.Family() {
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(603840)
		var ba *DBitArray
		switch v := d.(type) {
		case *DBitArray:
			__antithesis_instrumentation__.Notify(603872)
			ba = v
		case *DInt:
			__antithesis_instrumentation__.Notify(603873)
			var err error
			ba, err = NewDBitArrayFromInt(int64(*v), uint(t.Width()))
			if err != nil {
				__antithesis_instrumentation__.Notify(603878)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(603879)
			}
		case *DString:
			__antithesis_instrumentation__.Notify(603874)
			res, err := bitarray.Parse(string(*v))
			if err != nil {
				__antithesis_instrumentation__.Notify(603880)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(603881)
			}
			__antithesis_instrumentation__.Notify(603875)
			ba = &DBitArray{res}
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(603876)
			res, err := bitarray.Parse(v.Contents)
			if err != nil {
				__antithesis_instrumentation__.Notify(603882)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(603883)
			}
			__antithesis_instrumentation__.Notify(603877)
			ba = &DBitArray{res}
		}
		__antithesis_instrumentation__.Notify(603841)
		if truncateWidth {
			__antithesis_instrumentation__.Notify(603884)
			ba = formatBitArrayToType(ba, t)
		} else {
			__antithesis_instrumentation__.Notify(603885)
		}
		__antithesis_instrumentation__.Notify(603842)
		return ba, nil

	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(603843)
		switch v := d.(type) {
		case *DBool:
			__antithesis_instrumentation__.Notify(603886)
			return d, nil
		case *DInt:
			__antithesis_instrumentation__.Notify(603887)
			return MakeDBool(*v != 0), nil
		case *DFloat:
			__antithesis_instrumentation__.Notify(603888)
			return MakeDBool(*v != 0), nil
		case *DDecimal:
			__antithesis_instrumentation__.Notify(603889)
			return MakeDBool(v.Sign() != 0), nil
		case *DString:
			__antithesis_instrumentation__.Notify(603890)
			return ParseDBool(strings.TrimSpace(string(*v)))
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(603891)
			return ParseDBool(v.Contents)
		case *DJSON:
			__antithesis_instrumentation__.Notify(603892)
			b, ok := v.AsBool()
			if !ok {
				__antithesis_instrumentation__.Notify(603894)
				return nil, failedCastFromJSON(v, t)
			} else {
				__antithesis_instrumentation__.Notify(603895)
			}
			__antithesis_instrumentation__.Notify(603893)
			return MakeDBool(DBool(b)), nil
		}

	case types.IntFamily:
		__antithesis_instrumentation__.Notify(603844)
		var res *DInt
		switch v := d.(type) {
		case *DBitArray:
			__antithesis_instrumentation__.Notify(603896)
			res = v.AsDInt(uint(t.Width()))
		case *DBool:
			__antithesis_instrumentation__.Notify(603897)
			if *v {
				__antithesis_instrumentation__.Notify(603915)
				res = NewDInt(1)
			} else {
				__antithesis_instrumentation__.Notify(603916)
				res = DZero
			}
		case *DInt:
			__antithesis_instrumentation__.Notify(603898)

			res = v
		case *DFloat:
			__antithesis_instrumentation__.Notify(603899)
			f := float64(*v)

			if math.IsNaN(f) || func() bool {
				__antithesis_instrumentation__.Notify(603917)
				return f <= float64(math.MinInt64) == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(603918)
				return f >= float64(math.MaxInt64) == true
			}() == true {
				__antithesis_instrumentation__.Notify(603919)
				return nil, ErrIntOutOfRange
			} else {
				__antithesis_instrumentation__.Notify(603920)
			}
			__antithesis_instrumentation__.Notify(603900)
			res = NewDInt(DInt(f))
		case *DDecimal:
			__antithesis_instrumentation__.Notify(603901)
			i, err := roundDecimalToInt(ctx, &v.Decimal)
			if err != nil {
				__antithesis_instrumentation__.Notify(603921)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(603922)
			}
			__antithesis_instrumentation__.Notify(603902)
			res = NewDInt(DInt(i))
		case *DString:
			__antithesis_instrumentation__.Notify(603903)
			var err error
			if res, err = ParseDInt(strings.TrimSpace(string(*v))); err != nil {
				__antithesis_instrumentation__.Notify(603923)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(603924)
			}
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(603904)
			var err error
			if res, err = ParseDInt(v.Contents); err != nil {
				__antithesis_instrumentation__.Notify(603925)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(603926)
			}
		case *DTimestamp:
			__antithesis_instrumentation__.Notify(603905)
			res = NewDInt(DInt(v.Unix()))
		case *DTimestampTZ:
			__antithesis_instrumentation__.Notify(603906)
			res = NewDInt(DInt(v.Unix()))
		case *DDate:
			__antithesis_instrumentation__.Notify(603907)

			if !v.IsFinite() {
				__antithesis_instrumentation__.Notify(603927)
				return nil, ErrIntOutOfRange
			} else {
				__antithesis_instrumentation__.Notify(603928)
			}
			__antithesis_instrumentation__.Notify(603908)
			res = NewDInt(DInt(v.UnixEpochDays()))
		case *DInterval:
			__antithesis_instrumentation__.Notify(603909)
			iv, ok := v.AsInt64()
			if !ok {
				__antithesis_instrumentation__.Notify(603929)
				return nil, ErrIntOutOfRange
			} else {
				__antithesis_instrumentation__.Notify(603930)
			}
			__antithesis_instrumentation__.Notify(603910)
			res = NewDInt(DInt(iv))
		case *DOid:
			__antithesis_instrumentation__.Notify(603911)
			res = &v.DInt
		case *DJSON:
			__antithesis_instrumentation__.Notify(603912)
			dec, ok := v.AsDecimal()
			if !ok {
				__antithesis_instrumentation__.Notify(603931)
				return nil, failedCastFromJSON(v, t)
			} else {
				__antithesis_instrumentation__.Notify(603932)
			}
			__antithesis_instrumentation__.Notify(603913)
			i, err := dec.Int64()
			if err != nil {
				__antithesis_instrumentation__.Notify(603933)

				i, err = roundDecimalToInt(ctx, dec)
				if err != nil {
					__antithesis_instrumentation__.Notify(603934)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(603935)
				}
			} else {
				__antithesis_instrumentation__.Notify(603936)
			}
			__antithesis_instrumentation__.Notify(603914)
			res = NewDInt(DInt(i))
		}
		__antithesis_instrumentation__.Notify(603845)
		if res != nil {
			__antithesis_instrumentation__.Notify(603937)
			return res, nil
		} else {
			__antithesis_instrumentation__.Notify(603938)
		}

	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(603846)
		switch v := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(603939)
			return MakeDEnumFromLogicalRepresentation(t, string(*v))
		case *DBytes:
			__antithesis_instrumentation__.Notify(603940)
			return MakeDEnumFromPhysicalRepresentation(t, []byte(*v))
		case *DEnum:
			__antithesis_instrumentation__.Notify(603941)
			return d, nil
		}

	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(603847)
		switch v := d.(type) {
		case *DBool:
			__antithesis_instrumentation__.Notify(603942)
			if *v {
				__antithesis_instrumentation__.Notify(603958)
				return NewDFloat(1), nil
			} else {
				__antithesis_instrumentation__.Notify(603959)
			}
			__antithesis_instrumentation__.Notify(603943)
			return NewDFloat(0), nil
		case *DInt:
			__antithesis_instrumentation__.Notify(603944)
			return NewDFloat(DFloat(*v)), nil
		case *DFloat:
			__antithesis_instrumentation__.Notify(603945)
			return d, nil
		case *DDecimal:
			__antithesis_instrumentation__.Notify(603946)
			f, err := v.Float64()
			if err != nil {
				__antithesis_instrumentation__.Notify(603960)
				return nil, ErrFloatOutOfRange
			} else {
				__antithesis_instrumentation__.Notify(603961)
			}
			__antithesis_instrumentation__.Notify(603947)
			return NewDFloat(DFloat(f)), nil
		case *DString:
			__antithesis_instrumentation__.Notify(603948)
			return ParseDFloat(strings.TrimSpace(string(*v)))
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(603949)
			return ParseDFloat(v.Contents)
		case *DTimestamp:
			__antithesis_instrumentation__.Notify(603950)
			micros := float64(v.Nanosecond() / int(time.Microsecond))
			return NewDFloat(DFloat(float64(v.Unix()) + micros*1e-6)), nil
		case *DTimestampTZ:
			__antithesis_instrumentation__.Notify(603951)
			micros := float64(v.Nanosecond() / int(time.Microsecond))
			return NewDFloat(DFloat(float64(v.Unix()) + micros*1e-6)), nil
		case *DDate:
			__antithesis_instrumentation__.Notify(603952)

			if !v.IsFinite() {
				__antithesis_instrumentation__.Notify(603962)
				return nil, ErrFloatOutOfRange
			} else {
				__antithesis_instrumentation__.Notify(603963)
			}
			__antithesis_instrumentation__.Notify(603953)
			return NewDFloat(DFloat(float64(v.UnixEpochDays()))), nil
		case *DInterval:
			__antithesis_instrumentation__.Notify(603954)
			return NewDFloat(DFloat(v.AsFloat64())), nil
		case *DJSON:
			__antithesis_instrumentation__.Notify(603955)
			dec, ok := v.AsDecimal()
			if !ok {
				__antithesis_instrumentation__.Notify(603964)
				return nil, failedCastFromJSON(v, t)
			} else {
				__antithesis_instrumentation__.Notify(603965)
			}
			__antithesis_instrumentation__.Notify(603956)
			fl, err := dec.Float64()
			if err != nil {
				__antithesis_instrumentation__.Notify(603966)
				return nil, ErrFloatOutOfRange
			} else {
				__antithesis_instrumentation__.Notify(603967)
			}
			__antithesis_instrumentation__.Notify(603957)
			return NewDFloat(DFloat(fl)), nil
		}

	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(603848)
		var dd DDecimal
		var err error
		unset := false
		switch v := d.(type) {
		case *DBool:
			__antithesis_instrumentation__.Notify(603968)
			if *v {
				__antithesis_instrumentation__.Notify(603983)
				dd.SetInt64(1)
			} else {
				__antithesis_instrumentation__.Notify(603984)
			}
		case *DInt:
			__antithesis_instrumentation__.Notify(603969)
			dd.SetInt64(int64(*v))
		case *DDate:
			__antithesis_instrumentation__.Notify(603970)

			if !v.IsFinite() {
				__antithesis_instrumentation__.Notify(603985)
				return nil, ErrDecOutOfRange
			} else {
				__antithesis_instrumentation__.Notify(603986)
			}
			__antithesis_instrumentation__.Notify(603971)
			dd.SetInt64(v.UnixEpochDays())
		case *DFloat:
			__antithesis_instrumentation__.Notify(603972)
			_, err = dd.SetFloat64(float64(*v))
		case *DDecimal:
			__antithesis_instrumentation__.Notify(603973)

			if t.Precision() == 0 {
				__antithesis_instrumentation__.Notify(603987)
				return d, nil
			} else {
				__antithesis_instrumentation__.Notify(603988)
			}
			__antithesis_instrumentation__.Notify(603974)
			dd.Set(&v.Decimal)
		case *DString:
			__antithesis_instrumentation__.Notify(603975)
			err = dd.SetString(strings.TrimSpace(string(*v)))
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(603976)
			err = dd.SetString(v.Contents)
		case *DTimestamp:
			__antithesis_instrumentation__.Notify(603977)
			val := &dd.Coeff
			val.SetInt64(v.Unix())
			val.Mul(val, big10E6)
			micros := v.Nanosecond() / int(time.Microsecond)
			val.Add(val, apd.NewBigInt(int64(micros)))
			dd.Exponent = -6
		case *DTimestampTZ:
			__antithesis_instrumentation__.Notify(603978)
			val := &dd.Coeff
			val.SetInt64(v.Unix())
			val.Mul(val, big10E6)
			micros := v.Nanosecond() / int(time.Microsecond)
			val.Add(val, apd.NewBigInt(int64(micros)))
			dd.Exponent = -6
		case *DInterval:
			__antithesis_instrumentation__.Notify(603979)
			v.AsBigInt(&dd.Coeff)
			dd.Exponent = -9
		case *DJSON:
			__antithesis_instrumentation__.Notify(603980)
			dec, ok := v.AsDecimal()
			if !ok {
				__antithesis_instrumentation__.Notify(603989)
				return nil, failedCastFromJSON(v, t)
			} else {
				__antithesis_instrumentation__.Notify(603990)
			}
			__antithesis_instrumentation__.Notify(603981)
			dd.Set(dec)
		default:
			__antithesis_instrumentation__.Notify(603982)
			unset = true
		}
		__antithesis_instrumentation__.Notify(603849)
		if err != nil {
			__antithesis_instrumentation__.Notify(603991)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(603992)
		}
		__antithesis_instrumentation__.Notify(603850)
		if !unset {
			__antithesis_instrumentation__.Notify(603993)

			if dd.Coeff.Sign() < 0 {
				__antithesis_instrumentation__.Notify(603996)
				dd.Negative = true
				dd.Coeff.Abs(&dd.Coeff)
			} else {
				__antithesis_instrumentation__.Notify(603997)
			}
			__antithesis_instrumentation__.Notify(603994)
			err = LimitDecimalWidth(&dd.Decimal, int(t.Precision()), int(t.Scale()))
			if err != nil {
				__antithesis_instrumentation__.Notify(603998)
				return nil, errors.Wrapf(err, "type %s", t.SQLString())
			} else {
				__antithesis_instrumentation__.Notify(603999)
			}
			__antithesis_instrumentation__.Notify(603995)
			return &dd, nil
		} else {
			__antithesis_instrumentation__.Notify(604000)
		}

	case types.StringFamily, types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(603851)
		var s string
		typ := t
		switch t := d.(type) {
		case *DBitArray:
			__antithesis_instrumentation__.Notify(604001)
			s = t.BitArray.String()
		case *DFloat:
			__antithesis_instrumentation__.Notify(604002)
			s = strconv.FormatFloat(float64(*t), 'g',
				ctx.SessionData().DataConversionConfig.GetFloatPrec(), 64)
		case *DInt:
			__antithesis_instrumentation__.Notify(604003)
			if typ.Oid() == oid.T_char {
				__antithesis_instrumentation__.Notify(604020)

				if *t > math.MaxInt8 || func() bool {
					__antithesis_instrumentation__.Notify(604021)
					return *t < math.MinInt8 == true
				}() == true {
					__antithesis_instrumentation__.Notify(604022)
					return nil, errCharOutOfRange
				} else {
					__antithesis_instrumentation__.Notify(604023)
					if *t == 0 {
						__antithesis_instrumentation__.Notify(604024)
						s = ""
					} else {
						__antithesis_instrumentation__.Notify(604025)
						s = string([]byte{byte(*t)})
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(604026)
				s = d.String()
			}
		case *DBool, *DDecimal:
			__antithesis_instrumentation__.Notify(604004)
			s = d.String()
		case *DTimestamp, *DDate, *DTime, *DTimeTZ, *DGeography, *DGeometry, *DBox2D:
			__antithesis_instrumentation__.Notify(604005)
			s = AsStringWithFlags(d, FmtBareStrings)
		case *DTimestampTZ:
			__antithesis_instrumentation__.Notify(604006)

			ts, err := MakeDTimestampTZ(t.In(ctx.GetLocation()), time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(604027)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604028)
			}
			__antithesis_instrumentation__.Notify(604007)
			s = AsStringWithFlags(
				ts,
				FmtBareStrings,
			)
		case *DTuple:
			__antithesis_instrumentation__.Notify(604008)
			s = AsStringWithFlags(
				d,
				FmtPgwireText,
				FmtDataConversionConfig(ctx.SessionData().DataConversionConfig),
			)
		case *DArray:
			__antithesis_instrumentation__.Notify(604009)
			s = AsStringWithFlags(
				d,
				FmtPgwireText,
				FmtDataConversionConfig(ctx.SessionData().DataConversionConfig),
			)
		case *DInterval:
			__antithesis_instrumentation__.Notify(604010)

			s = AsStringWithFlags(
				d,
				FmtPgwireText,
				FmtDataConversionConfig(ctx.SessionData().DataConversionConfig),
			)
		case *DUuid:
			__antithesis_instrumentation__.Notify(604011)
			s = t.UUID.String()
		case *DIPAddr:
			__antithesis_instrumentation__.Notify(604012)
			s = AsStringWithFlags(d, FmtBareStrings)
		case *DString:
			__antithesis_instrumentation__.Notify(604013)
			s = string(*t)
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604014)
			s = t.Contents
		case *DBytes:
			__antithesis_instrumentation__.Notify(604015)
			s = lex.EncodeByteArrayToRawBytes(
				string(*t),
				ctx.SessionData().DataConversionConfig.BytesEncodeFormat,
				false,
			)
		case *DOid:
			__antithesis_instrumentation__.Notify(604016)
			s = t.String()
		case *DJSON:
			__antithesis_instrumentation__.Notify(604017)
			s = t.JSON.String()
		case *DEnum:
			__antithesis_instrumentation__.Notify(604018)
			s = t.LogicalRep
		case *DVoid:
			__antithesis_instrumentation__.Notify(604019)
			s = ""
		}
		__antithesis_instrumentation__.Notify(603852)
		switch t.Family() {
		case types.StringFamily:
			__antithesis_instrumentation__.Notify(604029)
			if t.Oid() == oid.T_name {
				__antithesis_instrumentation__.Notify(604037)
				return NewDName(s), nil
			} else {
				__antithesis_instrumentation__.Notify(604038)
			}
			__antithesis_instrumentation__.Notify(604030)

			if t.Oid() == oid.T_bpchar {
				__antithesis_instrumentation__.Notify(604039)
				s = strings.TrimRight(s, " ")
			} else {
				__antithesis_instrumentation__.Notify(604040)
			}
			__antithesis_instrumentation__.Notify(604031)

			if truncateWidth && func() bool {
				__antithesis_instrumentation__.Notify(604041)
				return t.Width() > 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(604042)
				s = util.TruncateString(s, int(t.Width()))
			} else {
				__antithesis_instrumentation__.Notify(604043)
			}
			__antithesis_instrumentation__.Notify(604032)
			return NewDString(s), nil
		case types.CollatedStringFamily:
			__antithesis_instrumentation__.Notify(604033)

			if t.Oid() == oid.T_bpchar {
				__antithesis_instrumentation__.Notify(604044)
				s = strings.TrimRight(s, " ")
			} else {
				__antithesis_instrumentation__.Notify(604045)
			}
			__antithesis_instrumentation__.Notify(604034)

			if truncateWidth && func() bool {
				__antithesis_instrumentation__.Notify(604046)
				return t.Width() > 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(604047)
				s = util.TruncateString(s, int(t.Width()))
			} else {
				__antithesis_instrumentation__.Notify(604048)
			}
			__antithesis_instrumentation__.Notify(604035)
			return NewDCollatedString(s, t.Locale(), &ctx.CollationEnv)
		default:
			__antithesis_instrumentation__.Notify(604036)
		}

	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(603853)
		switch t := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604049)
			return ParseDByte(string(*t))
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604050)
			return NewDBytes(DBytes(t.Contents)), nil
		case *DUuid:
			__antithesis_instrumentation__.Notify(604051)
			return NewDBytes(DBytes(t.GetBytes())), nil
		case *DBytes:
			__antithesis_instrumentation__.Notify(604052)
			return d, nil
		case *DGeography:
			__antithesis_instrumentation__.Notify(604053)
			return NewDBytes(DBytes(t.Geography.EWKB())), nil
		case *DGeometry:
			__antithesis_instrumentation__.Notify(604054)
			return NewDBytes(DBytes(t.Geometry.EWKB())), nil
		}

	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(603854)
		switch t := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604055)
			return ParseDUuidFromString(string(*t))
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604056)
			return ParseDUuidFromString(t.Contents)
		case *DBytes:
			__antithesis_instrumentation__.Notify(604057)
			return ParseDUuidFromBytes([]byte(*t))
		case *DUuid:
			__antithesis_instrumentation__.Notify(604058)
			return d, nil
		}

	case types.INetFamily:
		__antithesis_instrumentation__.Notify(603855)
		switch t := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604059)
			return ParseDIPAddrFromINetString(string(*t))
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604060)
			return ParseDIPAddrFromINetString(t.Contents)
		case *DIPAddr:
			__antithesis_instrumentation__.Notify(604061)
			return d, nil
		}

	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(603856)
		switch d := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604062)
			return ParseDBox2D(string(*d))
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604063)
			return ParseDBox2D(d.Contents)
		case *DBox2D:
			__antithesis_instrumentation__.Notify(604064)
			return d, nil
		case *DGeometry:
			__antithesis_instrumentation__.Notify(604065)
			bbox := d.CartesianBoundingBox()
			if bbox == nil {
				__antithesis_instrumentation__.Notify(604067)
				return DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(604068)
			}
			__antithesis_instrumentation__.Notify(604066)
			return NewDBox2D(*bbox), nil
		}

	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(603857)
		switch d := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604069)
			return ParseDGeography(string(*d))
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604070)
			return ParseDGeography(d.Contents)
		case *DGeography:
			__antithesis_instrumentation__.Notify(604071)
			if err := geo.SpatialObjectFitsColumnMetadata(
				d.Geography.SpatialObject(),
				t.InternalType.GeoMetadata.SRID,
				t.InternalType.GeoMetadata.ShapeType,
			); err != nil {
				__antithesis_instrumentation__.Notify(604082)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604083)
			}
			__antithesis_instrumentation__.Notify(604072)
			return d, nil
		case *DGeometry:
			__antithesis_instrumentation__.Notify(604073)
			g, err := d.AsGeography()
			if err != nil {
				__antithesis_instrumentation__.Notify(604084)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604085)
			}
			__antithesis_instrumentation__.Notify(604074)
			if err := geo.SpatialObjectFitsColumnMetadata(
				g.SpatialObject(),
				t.InternalType.GeoMetadata.SRID,
				t.InternalType.GeoMetadata.ShapeType,
			); err != nil {
				__antithesis_instrumentation__.Notify(604086)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604087)
			}
			__antithesis_instrumentation__.Notify(604075)
			return &DGeography{g}, nil
		case *DJSON:
			__antithesis_instrumentation__.Notify(604076)
			t, err := d.AsText()
			if err != nil {
				__antithesis_instrumentation__.Notify(604088)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604089)
			}
			__antithesis_instrumentation__.Notify(604077)
			if t == nil {
				__antithesis_instrumentation__.Notify(604090)
				return DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(604091)
			}
			__antithesis_instrumentation__.Notify(604078)
			g, err := geo.ParseGeographyFromGeoJSON([]byte(*t))
			if err != nil {
				__antithesis_instrumentation__.Notify(604092)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604093)
			}
			__antithesis_instrumentation__.Notify(604079)
			return &DGeography{g}, nil
		case *DBytes:
			__antithesis_instrumentation__.Notify(604080)
			g, err := geo.ParseGeographyFromEWKB(geopb.EWKB(*d))
			if err != nil {
				__antithesis_instrumentation__.Notify(604094)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604095)
			}
			__antithesis_instrumentation__.Notify(604081)
			return &DGeography{g}, nil
		}
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(603858)
		switch d := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604096)
			return ParseDGeometry(string(*d))
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604097)
			return ParseDGeometry(d.Contents)
		case *DGeometry:
			__antithesis_instrumentation__.Notify(604098)
			if err := geo.SpatialObjectFitsColumnMetadata(
				d.Geometry.SpatialObject(),
				t.InternalType.GeoMetadata.SRID,
				t.InternalType.GeoMetadata.ShapeType,
			); err != nil {
				__antithesis_instrumentation__.Notify(604111)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604112)
			}
			__antithesis_instrumentation__.Notify(604099)
			return d, nil
		case *DGeography:
			__antithesis_instrumentation__.Notify(604100)
			if err := geo.SpatialObjectFitsColumnMetadata(
				d.Geography.SpatialObject(),
				t.InternalType.GeoMetadata.SRID,
				t.InternalType.GeoMetadata.ShapeType,
			); err != nil {
				__antithesis_instrumentation__.Notify(604113)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604114)
			}
			__antithesis_instrumentation__.Notify(604101)
			g, err := d.AsGeometry()
			if err != nil {
				__antithesis_instrumentation__.Notify(604115)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604116)
			}
			__antithesis_instrumentation__.Notify(604102)
			return &DGeometry{g}, nil
		case *DJSON:
			__antithesis_instrumentation__.Notify(604103)
			t, err := d.AsText()
			if err != nil {
				__antithesis_instrumentation__.Notify(604117)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604118)
			}
			__antithesis_instrumentation__.Notify(604104)
			if t == nil {
				__antithesis_instrumentation__.Notify(604119)
				return DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(604120)
			}
			__antithesis_instrumentation__.Notify(604105)
			g, err := geo.ParseGeometryFromGeoJSON([]byte(*t))
			if err != nil {
				__antithesis_instrumentation__.Notify(604121)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604122)
			}
			__antithesis_instrumentation__.Notify(604106)
			return &DGeometry{g}, nil
		case *DBox2D:
			__antithesis_instrumentation__.Notify(604107)
			g, err := geo.MakeGeometryFromGeomT(d.ToGeomT(geopb.DefaultGeometrySRID))
			if err != nil {
				__antithesis_instrumentation__.Notify(604123)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604124)
			}
			__antithesis_instrumentation__.Notify(604108)
			return &DGeometry{g}, nil
		case *DBytes:
			__antithesis_instrumentation__.Notify(604109)
			g, err := geo.ParseGeometryFromEWKB(geopb.EWKB(*d))
			if err != nil {
				__antithesis_instrumentation__.Notify(604125)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604126)
			}
			__antithesis_instrumentation__.Notify(604110)
			return &DGeometry{g}, nil
		}

	case types.DateFamily:
		__antithesis_instrumentation__.Notify(603859)
		switch d := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604127)
			res, _, err := ParseDDate(ctx, string(*d))
			return res, err
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604128)
			res, _, err := ParseDDate(ctx, d.Contents)
			return res, err
		case *DDate:
			__antithesis_instrumentation__.Notify(604129)
			return d, nil
		case *DInt:
			__antithesis_instrumentation__.Notify(604130)

			t, err := pgdate.MakeDateFromUnixEpoch(int64(*d))
			return NewDDate(t), err
		case *DTimestampTZ:
			__antithesis_instrumentation__.Notify(604131)
			return NewDDateFromTime(d.Time.In(ctx.GetLocation()))
		case *DTimestamp:
			__antithesis_instrumentation__.Notify(604132)
			return NewDDateFromTime(d.Time)
		}

	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(603860)
		roundTo := TimeFamilyPrecisionToRoundDuration(t.Precision())
		switch d := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604133)
			res, _, err := ParseDTime(ctx, string(*d), roundTo)
			return res, err
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604134)
			res, _, err := ParseDTime(ctx, d.Contents, roundTo)
			return res, err
		case *DTime:
			__antithesis_instrumentation__.Notify(604135)
			return d.Round(roundTo), nil
		case *DTimeTZ:
			__antithesis_instrumentation__.Notify(604136)
			return MakeDTime(d.TimeOfDay.Round(roundTo)), nil
		case *DTimestamp:
			__antithesis_instrumentation__.Notify(604137)
			return MakeDTime(timeofday.FromTime(d.Time).Round(roundTo)), nil
		case *DTimestampTZ:
			__antithesis_instrumentation__.Notify(604138)

			stripped, err := d.stripTimeZone(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(604141)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604142)
			}
			__antithesis_instrumentation__.Notify(604139)
			return MakeDTime(timeofday.FromTime(stripped.Time).Round(roundTo)), nil
		case *DInterval:
			__antithesis_instrumentation__.Notify(604140)
			return MakeDTime(timeofday.Min.Add(d.Duration).Round(roundTo)), nil
		}

	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(603861)
		roundTo := TimeFamilyPrecisionToRoundDuration(t.Precision())
		switch d := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604143)
			res, _, err := ParseDTimeTZ(ctx, string(*d), roundTo)
			return res, err
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604144)
			res, _, err := ParseDTimeTZ(ctx, d.Contents, roundTo)
			return res, err
		case *DTime:
			__antithesis_instrumentation__.Notify(604145)
			return NewDTimeTZFromLocation(timeofday.TimeOfDay(*d).Round(roundTo), ctx.GetLocation()), nil
		case *DTimeTZ:
			__antithesis_instrumentation__.Notify(604146)
			return d.Round(roundTo), nil
		case *DTimestampTZ:
			__antithesis_instrumentation__.Notify(604147)
			return NewDTimeTZFromTime(d.Time.In(ctx.GetLocation()).Round(roundTo)), nil
		}

	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(603862)
		roundTo := TimeFamilyPrecisionToRoundDuration(t.Precision())

		switch d := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604148)
			res, _, err := ParseDTimestamp(ctx, string(*d), roundTo)
			return res, err
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604149)
			res, _, err := ParseDTimestamp(ctx, d.Contents, roundTo)
			return res, err
		case *DDate:
			__antithesis_instrumentation__.Notify(604150)
			t, err := d.ToTime()
			if err != nil {
				__antithesis_instrumentation__.Notify(604156)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604157)
			}
			__antithesis_instrumentation__.Notify(604151)
			return MakeDTimestamp(t, roundTo)
		case *DInt:
			__antithesis_instrumentation__.Notify(604152)
			return MakeDTimestamp(timeutil.Unix(int64(*d), 0), roundTo)
		case *DTimestamp:
			__antithesis_instrumentation__.Notify(604153)
			return d.Round(roundTo)
		case *DTimestampTZ:
			__antithesis_instrumentation__.Notify(604154)

			stripped, err := d.stripTimeZone(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(604158)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604159)
			}
			__antithesis_instrumentation__.Notify(604155)
			return stripped.Round(roundTo)
		}

	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(603863)
		roundTo := TimeFamilyPrecisionToRoundDuration(t.Precision())

		switch d := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604160)
			res, _, err := ParseDTimestampTZ(ctx, string(*d), roundTo)
			return res, err
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604161)
			res, _, err := ParseDTimestampTZ(ctx, d.Contents, roundTo)
			return res, err
		case *DDate:
			__antithesis_instrumentation__.Notify(604162)
			t, err := d.ToTime()
			if err != nil {
				__antithesis_instrumentation__.Notify(604167)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604168)
			}
			__antithesis_instrumentation__.Notify(604163)
			_, before := t.Zone()
			_, after := t.In(ctx.GetLocation()).Zone()
			return MakeDTimestampTZ(t.Add(time.Duration(before-after)*time.Second), roundTo)
		case *DTimestamp:
			__antithesis_instrumentation__.Notify(604164)
			_, before := d.Time.Zone()
			_, after := d.Time.In(ctx.GetLocation()).Zone()
			return MakeDTimestampTZ(d.Time.Add(time.Duration(before-after)*time.Second), roundTo)
		case *DInt:
			__antithesis_instrumentation__.Notify(604165)
			return MakeDTimestampTZ(timeutil.Unix(int64(*d), 0), roundTo)
		case *DTimestampTZ:
			__antithesis_instrumentation__.Notify(604166)
			return d.Round(roundTo)
		}

	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(603864)
		itm, err := t.IntervalTypeMetadata()
		if err != nil {
			__antithesis_instrumentation__.Notify(604169)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(604170)
		}
		__antithesis_instrumentation__.Notify(603865)
		switch v := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604171)
			return ParseDIntervalWithTypeMetadata(ctx.GetIntervalStyle(), string(*v), itm)
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(604172)
			return ParseDIntervalWithTypeMetadata(ctx.GetIntervalStyle(), v.Contents, itm)
		case *DInt:
			__antithesis_instrumentation__.Notify(604173)
			return NewDInterval(duration.FromInt64(int64(*v)), itm), nil
		case *DFloat:
			__antithesis_instrumentation__.Notify(604174)
			return NewDInterval(duration.FromFloat64(float64(*v)), itm), nil
		case *DTime:
			__antithesis_instrumentation__.Notify(604175)
			return NewDInterval(duration.MakeDuration(int64(*v)*1000, 0, 0), itm), nil
		case *DDecimal:
			__antithesis_instrumentation__.Notify(604176)
			var d apd.Decimal
			dnanos := v.Decimal
			dnanos.Exponent += 9

			_, err := HighPrecisionCtx.Quantize(&d, &dnanos, 0)
			if err != nil {
				__antithesis_instrumentation__.Notify(604181)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604182)
			}
			__antithesis_instrumentation__.Notify(604177)
			if dnanos.Negative {
				__antithesis_instrumentation__.Notify(604183)
				d.Coeff.Neg(&d.Coeff)
			} else {
				__antithesis_instrumentation__.Notify(604184)
			}
			__antithesis_instrumentation__.Notify(604178)
			dv, ok := duration.FromBigInt(&d.Coeff)
			if !ok {
				__antithesis_instrumentation__.Notify(604185)
				return nil, ErrDecOutOfRange
			} else {
				__antithesis_instrumentation__.Notify(604186)
			}
			__antithesis_instrumentation__.Notify(604179)
			return NewDInterval(dv, itm), nil
		case *DInterval:
			__antithesis_instrumentation__.Notify(604180)
			return NewDInterval(v.Duration, itm), nil
		}
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(603866)
		switch v := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604187)
			return ParseDJSON(string(*v))
		case *DJSON:
			__antithesis_instrumentation__.Notify(604188)
			return v, nil
		case *DGeography:
			__antithesis_instrumentation__.Notify(604189)
			j, err := geo.SpatialObjectToGeoJSON(v.Geography.SpatialObject(), -1, geo.SpatialObjectToGeoJSONFlagZero)
			if err != nil {
				__antithesis_instrumentation__.Notify(604193)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604194)
			}
			__antithesis_instrumentation__.Notify(604190)
			return ParseDJSON(string(j))
		case *DGeometry:
			__antithesis_instrumentation__.Notify(604191)
			j, err := geo.SpatialObjectToGeoJSON(v.Geometry.SpatialObject(), -1, geo.SpatialObjectToGeoJSONFlagZero)
			if err != nil {
				__antithesis_instrumentation__.Notify(604195)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604196)
			}
			__antithesis_instrumentation__.Notify(604192)
			return ParseDJSON(string(j))
		}
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(603867)
		switch v := d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604197)
			res, _, err := ParseDArrayFromString(ctx, string(*v), t.ArrayContents())
			return res, err
		case *DArray:
			__antithesis_instrumentation__.Notify(604198)
			dcast := NewDArray(t.ArrayContents())
			if err := dcast.MaybeSetCustomOid(t); err != nil {
				__antithesis_instrumentation__.Notify(604201)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604202)
			}
			__antithesis_instrumentation__.Notify(604199)
			for _, e := range v.Array {
				__antithesis_instrumentation__.Notify(604203)
				ecast := DNull
				if e != DNull {
					__antithesis_instrumentation__.Notify(604205)
					var err error
					ecast, err = PerformCast(ctx, e, t.ArrayContents())
					if err != nil {
						__antithesis_instrumentation__.Notify(604206)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(604207)
					}
				} else {
					__antithesis_instrumentation__.Notify(604208)
				}
				__antithesis_instrumentation__.Notify(604204)

				if err := dcast.Append(ecast); err != nil {
					__antithesis_instrumentation__.Notify(604209)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(604210)
				}
			}
			__antithesis_instrumentation__.Notify(604200)
			return dcast, nil
		}
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(603868)
		switch v := d.(type) {
		case *DOid:
			__antithesis_instrumentation__.Notify(604211)
			return performIntToOidCast(ctx, t, v.DInt)
		case *DInt:
			__antithesis_instrumentation__.Notify(604212)

			i := DInt(uint32(*v))
			return performIntToOidCast(ctx, t, i)
		case *DString:
			__antithesis_instrumentation__.Notify(604213)
			if t.Oid() != oid.T_oid && func() bool {
				__antithesis_instrumentation__.Notify(604215)
				return string(*v) == ZeroOidValue == true
			}() == true {
				__antithesis_instrumentation__.Notify(604216)
				return wrapAsZeroOid(t), nil
			} else {
				__antithesis_instrumentation__.Notify(604217)
			}
			__antithesis_instrumentation__.Notify(604214)
			return ParseDOid(ctx, string(*v), t)
		}
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(603869)
		switch v := d.(type) {
		case *DTuple:
			__antithesis_instrumentation__.Notify(604218)
			if t == types.AnyTuple {
				__antithesis_instrumentation__.Notify(604223)

				return v, nil
			} else {
				__antithesis_instrumentation__.Notify(604224)
			}
			__antithesis_instrumentation__.Notify(604219)

			if len(v.D) != len(t.TupleContents()) {
				__antithesis_instrumentation__.Notify(604225)
				return nil, pgerror.New(
					pgcode.CannotCoerce, "cannot cast tuple; wrong number of columns")
			} else {
				__antithesis_instrumentation__.Notify(604226)
			}
			__antithesis_instrumentation__.Notify(604220)
			ret := NewDTupleWithLen(t, len(v.D))
			for i := range v.D {
				__antithesis_instrumentation__.Notify(604227)
				var err error
				ret.D[i], err = PerformCast(ctx, v.D[i], t.TupleContents()[i])
				if err != nil {
					__antithesis_instrumentation__.Notify(604228)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(604229)
				}
			}
			__antithesis_instrumentation__.Notify(604221)
			return ret, nil
		case *DString:
			__antithesis_instrumentation__.Notify(604222)
			res, _, err := ParseDTupleFromString(ctx, string(*v), t)
			return res, err
		}
	case types.VoidFamily:
		__antithesis_instrumentation__.Notify(603870)
		switch d.(type) {
		case *DString:
			__antithesis_instrumentation__.Notify(604230)
			return DVoidDatum, nil
		}
	default:
		__antithesis_instrumentation__.Notify(603871)
	}
	__antithesis_instrumentation__.Notify(603837)

	return nil, pgerror.Newf(
		pgcode.CannotCoerce, "invalid cast: %s -> %s", d.ResolvedType(), t)
}

func performIntToOidCast(ctx *EvalContext, t *types.T, v DInt) (Datum, error) {
	__antithesis_instrumentation__.Notify(604231)
	switch t.Oid() {
	case oid.T_oid:
		__antithesis_instrumentation__.Notify(604232)
		return &DOid{semanticType: t, DInt: v}, nil
	case oid.T_regtype:
		__antithesis_instrumentation__.Notify(604233)

		ret := &DOid{semanticType: t, DInt: v}
		if typ, ok := types.OidToType[oid.Oid(v)]; ok {
			__antithesis_instrumentation__.Notify(604240)
			ret.name = typ.PGName()
		} else {
			__antithesis_instrumentation__.Notify(604241)
			if types.IsOIDUserDefinedType(oid.Oid(v)) {
				__antithesis_instrumentation__.Notify(604242)
				typ, err := ctx.Planner.ResolveTypeByOID(ctx.Context, oid.Oid(v))
				if err != nil {
					__antithesis_instrumentation__.Notify(604244)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(604245)
				}
				__antithesis_instrumentation__.Notify(604243)
				ret.name = typ.PGName()
			} else {
				__antithesis_instrumentation__.Notify(604246)
				if v == 0 {
					__antithesis_instrumentation__.Notify(604247)
					return wrapAsZeroOid(t), nil
				} else {
					__antithesis_instrumentation__.Notify(604248)
				}
			}
		}
		__antithesis_instrumentation__.Notify(604234)
		return ret, nil

	case oid.T_regproc, oid.T_regprocedure:
		__antithesis_instrumentation__.Notify(604235)

		name, ok := OidToBuiltinName[oid.Oid(v)]
		ret := &DOid{semanticType: t, DInt: v}
		if !ok {
			__antithesis_instrumentation__.Notify(604249)
			if v == 0 {
				__antithesis_instrumentation__.Notify(604251)
				return wrapAsZeroOid(t), nil
			} else {
				__antithesis_instrumentation__.Notify(604252)
			}
			__antithesis_instrumentation__.Notify(604250)
			return ret, nil
		} else {
			__antithesis_instrumentation__.Notify(604253)
		}
		__antithesis_instrumentation__.Notify(604236)
		ret.name = name
		return ret, nil

	default:
		__antithesis_instrumentation__.Notify(604237)
		if v == 0 {
			__antithesis_instrumentation__.Notify(604254)
			return wrapAsZeroOid(t), nil
		} else {
			__antithesis_instrumentation__.Notify(604255)
		}
		__antithesis_instrumentation__.Notify(604238)

		dOid, err := ctx.Planner.ResolveOIDFromOID(ctx.Ctx(), t, NewDOid(v))
		if err != nil {
			__antithesis_instrumentation__.Notify(604256)
			dOid = NewDOid(v)
			dOid.semanticType = t
		} else {
			__antithesis_instrumentation__.Notify(604257)
		}
		__antithesis_instrumentation__.Notify(604239)
		return dOid, nil
	}
}

func roundDecimalToInt(ctx *EvalContext, d *apd.Decimal) (int64, error) {
	__antithesis_instrumentation__.Notify(604258)
	var newD apd.Decimal
	if _, err := DecimalCtx.RoundToIntegralValue(&newD, d); err != nil {
		__antithesis_instrumentation__.Notify(604261)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(604262)
	}
	__antithesis_instrumentation__.Notify(604259)
	i, err := newD.Int64()
	if err != nil {
		__antithesis_instrumentation__.Notify(604263)
		return 0, ErrIntOutOfRange
	} else {
		__antithesis_instrumentation__.Notify(604264)
	}
	__antithesis_instrumentation__.Notify(604260)
	return i, nil
}

func failedCastFromJSON(j *DJSON, t *types.T) error {
	__antithesis_instrumentation__.Notify(604265)
	return pgerror.Newf(
		pgcode.InvalidParameterValue,
		"cannot cast jsonb %s to type %s",
		j.Type(), t,
	)
}

func PopulateRecordWithJSON(
	ctx *EvalContext, j json.JSON, desiredType *types.T, tup *DTuple,
) error {
	__antithesis_instrumentation__.Notify(604266)
	if j.Type() != json.ObjectJSONType {
		__antithesis_instrumentation__.Notify(604270)
		return pgerror.Newf(pgcode.InvalidParameterValue, "expected JSON object")
	} else {
		__antithesis_instrumentation__.Notify(604271)
	}
	__antithesis_instrumentation__.Notify(604267)
	tupleTypes := desiredType.TupleContents()
	labels := desiredType.TupleLabels()
	if labels == nil {
		__antithesis_instrumentation__.Notify(604272)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			"anonymous records cannot be used with json{b}_populate_record{set}",
		)
	} else {
		__antithesis_instrumentation__.Notify(604273)
	}
	__antithesis_instrumentation__.Notify(604268)
	for i := range tupleTypes {
		__antithesis_instrumentation__.Notify(604274)
		val, err := j.FetchValKey(labels[i])
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(604276)
			return val == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(604277)

			continue
		} else {
			__antithesis_instrumentation__.Notify(604278)
		}
		__antithesis_instrumentation__.Notify(604275)
		tup.D[i], err = PopulateDatumWithJSON(ctx, val, tupleTypes[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(604279)
			return err
		} else {
			__antithesis_instrumentation__.Notify(604280)
		}
	}
	__antithesis_instrumentation__.Notify(604269)
	return nil
}

func PopulateDatumWithJSON(ctx *EvalContext, j json.JSON, desiredType *types.T) (Datum, error) {
	__antithesis_instrumentation__.Notify(604281)
	if j == json.NullJSONValue {
		__antithesis_instrumentation__.Notify(604285)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(604286)
	}
	__antithesis_instrumentation__.Notify(604282)
	switch desiredType.Family() {
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(604287)
		if j.Type() != json.ArrayJSONType {
			__antithesis_instrumentation__.Notify(604293)
			return nil, pgerror.Newf(pgcode.InvalidTextRepresentation, "expected JSON array")
		} else {
			__antithesis_instrumentation__.Notify(604294)
		}
		__antithesis_instrumentation__.Notify(604288)
		n := j.Len()
		elementTyp := desiredType.ArrayContents()
		d := NewDArray(elementTyp)
		d.Array = make(Datums, n)
		for i := 0; i < n; i++ {
			__antithesis_instrumentation__.Notify(604295)
			elt, err := j.FetchValIdx(i)
			if err != nil {
				__antithesis_instrumentation__.Notify(604297)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604298)
			}
			__antithesis_instrumentation__.Notify(604296)
			d.Array[i], err = PopulateDatumWithJSON(ctx, elt, elementTyp)
			if err != nil {
				__antithesis_instrumentation__.Notify(604299)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604300)
			}
		}
		__antithesis_instrumentation__.Notify(604289)
		return d, nil
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(604290)
		tup := NewDTupleWithLen(desiredType, len(desiredType.TupleContents()))
		for i := range tup.D {
			__antithesis_instrumentation__.Notify(604301)
			tup.D[i] = DNull
		}
		__antithesis_instrumentation__.Notify(604291)
		err := PopulateRecordWithJSON(ctx, j, desiredType, tup)
		return tup, err
	default:
		__antithesis_instrumentation__.Notify(604292)
	}
	__antithesis_instrumentation__.Notify(604283)
	var s string
	switch j.Type() {
	case json.StringJSONType:
		__antithesis_instrumentation__.Notify(604302)
		t, err := j.AsText()
		if err != nil {
			__antithesis_instrumentation__.Notify(604305)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(604306)
		}
		__antithesis_instrumentation__.Notify(604303)
		s = *t
	default:
		__antithesis_instrumentation__.Notify(604304)
		s = j.String()
	}
	__antithesis_instrumentation__.Notify(604284)
	return PerformCast(ctx, NewDString(s), desiredType)
}

type castCounterType struct {
	from, to types.Family
}

var castCounters map[castCounterType]telemetry.Counter

func init() {
	castCounters = make(map[castCounterType]telemetry.Counter)
	for fromID := range types.Family_name {
		for toID := range types.Family_name {
			from := types.Family(fromID)
			to := types.Family(toID)
			var c telemetry.Counter
			switch {
			case from == types.ArrayFamily && to == types.ArrayFamily:
				c = sqltelemetry.ArrayCastCounter
			case from == types.TupleFamily && to == types.TupleFamily:
				c = sqltelemetry.TupleCastCounter
			case from == types.EnumFamily && to == types.EnumFamily:
				c = sqltelemetry.EnumCastCounter
			default:
				c = sqltelemetry.CastOpCounter(from.Name(), to.Name())
			}
			castCounters[castCounterType{from, to}] = c
		}
	}
}

func GetCastCounter(from, to types.Family) telemetry.Counter {
	__antithesis_instrumentation__.Notify(604307)
	if c, ok := castCounters[castCounterType{from, to}]; ok {
		__antithesis_instrumentation__.Notify(604309)
		return c
	} else {
		__antithesis_instrumentation__.Notify(604310)
	}
	__antithesis_instrumentation__.Notify(604308)
	panic(errors.AssertionFailedf(
		"no cast counter found for cast from %s to %s", from.Name(), to.Name(),
	))
}
