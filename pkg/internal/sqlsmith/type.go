package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

func (s *Smither) typeFromSQLTypeSyntax(typeStr string) (*types.T, error) {
	__antithesis_instrumentation__.Notify(69980)
	typRef, err := parser.GetTypeFromValidSQLSyntax(typeStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(69983)
		return nil, errors.AssertionFailedf("failed to parse type: %v", typeStr)
	} else {
		__antithesis_instrumentation__.Notify(69984)
	}
	__antithesis_instrumentation__.Notify(69981)
	typ, err := tree.ResolveType(context.Background(), typRef, s)
	if err != nil {
		__antithesis_instrumentation__.Notify(69985)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(69986)
	}
	__antithesis_instrumentation__.Notify(69982)
	return typ, nil
}

func (s *Smither) pickAnyType(typ *types.T) *types.T {
	__antithesis_instrumentation__.Notify(69987)
	switch typ.Family() {
	case types.AnyFamily:
		__antithesis_instrumentation__.Notify(69989)
		typ = s.randType()
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(69990)
		if typ.ArrayContents().Family() == types.AnyFamily {
			__antithesis_instrumentation__.Notify(69992)
			typ = randgen.RandArrayContentsType(s.rnd)
		} else {
			__antithesis_instrumentation__.Notify(69993)
		}
	default:
		__antithesis_instrumentation__.Notify(69991)
	}
	__antithesis_instrumentation__.Notify(69988)
	return typ
}

func (s *Smither) randScalarType() *types.T {
	__antithesis_instrumentation__.Notify(69994)
	s.lock.RLock()
	defer s.lock.RUnlock()
	scalarTypes := types.Scalar
	if s.types != nil {
		__antithesis_instrumentation__.Notify(69996)
		scalarTypes = s.types.scalarTypes
	} else {
		__antithesis_instrumentation__.Notify(69997)
	}
	__antithesis_instrumentation__.Notify(69995)
	return randgen.RandTypeFromSlice(s.rnd, scalarTypes)
}

func (s *Smither) isScalarType(t *types.T) bool {
	__antithesis_instrumentation__.Notify(69998)
	s.lock.AssertRHeld()
	scalarTypes := types.Scalar
	if s.types != nil {
		__antithesis_instrumentation__.Notify(70001)
		scalarTypes = s.types.scalarTypes
	} else {
		__antithesis_instrumentation__.Notify(70002)
	}
	__antithesis_instrumentation__.Notify(69999)
	for i := range scalarTypes {
		__antithesis_instrumentation__.Notify(70003)
		if t.Identical(scalarTypes[i]) {
			__antithesis_instrumentation__.Notify(70004)
			return true
		} else {
			__antithesis_instrumentation__.Notify(70005)
		}
	}
	__antithesis_instrumentation__.Notify(70000)
	return false
}

func (s *Smither) randType() *types.T {
	__antithesis_instrumentation__.Notify(70006)
	s.lock.RLock()
	defer s.lock.RUnlock()
	seedTypes := randgen.SeedTypes
	if s.types != nil {
		__antithesis_instrumentation__.Notify(70008)
		seedTypes = s.types.seedTypes
	} else {
		__antithesis_instrumentation__.Notify(70009)
	}
	__antithesis_instrumentation__.Notify(70007)
	return randgen.RandTypeFromSlice(s.rnd, seedTypes)
}

func (s *Smither) makeDesiredTypes() []*types.T {
	__antithesis_instrumentation__.Notify(70010)
	var typs []*types.T
	for {
		__antithesis_instrumentation__.Notify(70012)
		typs = append(typs, s.randType())
		if s.d6() < 2 || func() bool {
			__antithesis_instrumentation__.Notify(70013)
			return !s.canRecurse() == true
		}() == true {
			__antithesis_instrumentation__.Notify(70014)
			break
		} else {
			__antithesis_instrumentation__.Notify(70015)
		}
	}
	__antithesis_instrumentation__.Notify(70011)
	return typs
}

type typeInfo struct {
	udts        map[tree.TypeName]*types.T
	seedTypes   []*types.T
	scalarTypes []*types.T
}

func (s *Smither) ResolveType(
	_ context.Context, name *tree.UnresolvedObjectName,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(70016)
	key := tree.MakeSchemaQualifiedTypeName(name.Schema(), name.Object())
	res, ok := s.types.udts[key]
	if !ok {
		__antithesis_instrumentation__.Notify(70018)
		return nil, errors.Newf("type name %s not found by smither", name.Object())
	} else {
		__antithesis_instrumentation__.Notify(70019)
	}
	__antithesis_instrumentation__.Notify(70017)
	return res, nil
}

func (s *Smither) ResolveTypeByOID(context.Context, oid.Oid) (*types.T, error) {
	__antithesis_instrumentation__.Notify(70020)
	return nil, errors.AssertionFailedf("smither cannot resolve types by OID")
}
