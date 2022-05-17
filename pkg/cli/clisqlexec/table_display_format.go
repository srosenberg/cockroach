package clisqlexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

type TableDisplayFormat int

const (
	TableDisplayTSV TableDisplayFormat = iota

	TableDisplayCSV

	TableDisplayTable

	TableDisplayRecords

	TableDisplaySQL

	TableDisplayHTML

	TableDisplayRawHTML

	TableDisplayRaw

	TableDisplayLastFormat
)

var _ pflag.Value = (*TableDisplayFormat)(nil)

func (f *TableDisplayFormat) Type() string {
	__antithesis_instrumentation__.Notify(29394)
	return "string"
}

func (f *TableDisplayFormat) String() string {
	__antithesis_instrumentation__.Notify(29395)
	switch *f {
	case TableDisplayTSV:
		__antithesis_instrumentation__.Notify(29397)
		return "tsv"
	case TableDisplayCSV:
		__antithesis_instrumentation__.Notify(29398)
		return "csv"
	case TableDisplayTable:
		__antithesis_instrumentation__.Notify(29399)
		return "table"
	case TableDisplayRecords:
		__antithesis_instrumentation__.Notify(29400)
		return "records"
	case TableDisplaySQL:
		__antithesis_instrumentation__.Notify(29401)
		return "sql"
	case TableDisplayHTML:
		__antithesis_instrumentation__.Notify(29402)
		return "html"
	case TableDisplayRawHTML:
		__antithesis_instrumentation__.Notify(29403)
		return "rawhtml"
	case TableDisplayRaw:
		__antithesis_instrumentation__.Notify(29404)
		return "raw"
	default:
		__antithesis_instrumentation__.Notify(29405)
	}
	__antithesis_instrumentation__.Notify(29396)
	return ""
}

func (f *TableDisplayFormat) Set(s string) error {
	__antithesis_instrumentation__.Notify(29406)
	switch s {
	case "tsv":
		__antithesis_instrumentation__.Notify(29408)
		*f = TableDisplayTSV
	case "csv":
		__antithesis_instrumentation__.Notify(29409)
		*f = TableDisplayCSV
	case "table":
		__antithesis_instrumentation__.Notify(29410)
		*f = TableDisplayTable
	case "records":
		__antithesis_instrumentation__.Notify(29411)
		*f = TableDisplayRecords
	case "sql":
		__antithesis_instrumentation__.Notify(29412)
		*f = TableDisplaySQL
	case "html":
		__antithesis_instrumentation__.Notify(29413)
		*f = TableDisplayHTML
	case "rawhtml":
		__antithesis_instrumentation__.Notify(29414)
		*f = TableDisplayRawHTML
	case "raw":
		__antithesis_instrumentation__.Notify(29415)
		*f = TableDisplayRaw
	default:
		__antithesis_instrumentation__.Notify(29416)
		return errors.Newf("invalid table display format: %s "+

			"(possible values: tsv, csv, table, records, sql, html, raw)", s)
	}
	__antithesis_instrumentation__.Notify(29407)
	return nil
}
