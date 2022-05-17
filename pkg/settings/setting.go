package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"
)

type Setting interface {
	Class() Class

	Key() string

	Typ() string

	String(sv *Values) string

	Description() string

	Visibility() Visibility
}

type NonMaskedSetting interface {
	Setting

	Encoded(sv *Values) string

	EncodedDefault() string

	DecodeToString(encoded string) (repr string, err error)

	SetOnChange(sv *Values, fn func(ctx context.Context))

	ErrorHint() (bool, string)
}

type Class int8

const (
	SystemOnly Class = iota

	TenantReadOnly

	TenantWritable
)

type Visibility int8

const (
	Reserved Visibility = iota

	Public
)

func AdminOnly(name string) bool {
	__antithesis_instrumentation__.Notify(240096)
	return !strings.HasPrefix(name, "sql.defaults.")
}
