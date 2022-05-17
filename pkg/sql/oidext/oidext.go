// Package oidext contains oids that are not in `github.com/lib/pq/oid`
// as they are not shipped by default with postgres.
// As CRDB does not support extensions, we'll need to automatically assign
// a few OIDs of our own.
package oidext

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/lib/pq/oid"

const CockroachPredefinedOIDMax = 100000

const (
	T_geometry   = oid.Oid(90000)
	T__geometry  = oid.Oid(90001)
	T_geography  = oid.Oid(90002)
	T__geography = oid.Oid(90003)
	T_box2d      = oid.Oid(90004)
	T__box2d     = oid.Oid(90005)
)

var ExtensionTypeName = map[oid.Oid]string{
	T_geometry:   "GEOMETRY",
	T__geometry:  "_GEOMETRY",
	T_geography:  "GEOGRAPHY",
	T__geography: "_GEOGRAPHY",
	T_box2d:      "BOX2D",
	T__box2d:     "_BOX2D",
}

func TypeName(o oid.Oid) (string, bool) {
	__antithesis_instrumentation__.Notify(501859)
	name, ok := oid.TypeName[o]
	if ok {
		__antithesis_instrumentation__.Notify(501861)
		return name, ok
	} else {
		__antithesis_instrumentation__.Notify(501862)
	}
	__antithesis_instrumentation__.Notify(501860)
	name, ok = ExtensionTypeName[o]
	return name, ok
}
