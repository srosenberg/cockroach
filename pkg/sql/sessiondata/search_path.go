package sessiondata

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
)

var DefaultSearchPath = MakeSearchPath(
	[]string{catconstants.UserSchemaName, catconstants.PublicSchemaName},
)

type SearchPath struct {
	paths                []string
	containsPgCatalog    bool
	containsPgExtension  bool
	containsPgTempSchema bool
	tempSchemaName       string
	userSchemaName       string
}

var EmptySearchPath = SearchPath{}

func DefaultSearchPathForUser(username security.SQLUsername) SearchPath {
	__antithesis_instrumentation__.Notify(617806)
	return DefaultSearchPath.WithUserSchemaName(username.Normalized())
}

func MakeSearchPath(paths []string) SearchPath {
	__antithesis_instrumentation__.Notify(617807)
	containsPgCatalog := false
	containsPgExtension := false
	containsPgTempSchema := false
	for _, e := range paths {
		__antithesis_instrumentation__.Notify(617809)
		switch e {
		case catconstants.PgCatalogName:
			__antithesis_instrumentation__.Notify(617810)
			containsPgCatalog = true
		case catconstants.PgTempSchemaName:
			__antithesis_instrumentation__.Notify(617811)
			containsPgTempSchema = true
		case catconstants.PgExtensionSchemaName:
			__antithesis_instrumentation__.Notify(617812)
			containsPgExtension = true
		default:
			__antithesis_instrumentation__.Notify(617813)
		}
	}
	__antithesis_instrumentation__.Notify(617808)
	return SearchPath{
		paths:                paths,
		containsPgCatalog:    containsPgCatalog,
		containsPgExtension:  containsPgExtension,
		containsPgTempSchema: containsPgTempSchema,
	}
}

func (s SearchPath) WithTemporarySchemaName(tempSchemaName string) SearchPath {
	__antithesis_instrumentation__.Notify(617814)
	return SearchPath{
		paths:                s.paths,
		containsPgCatalog:    s.containsPgCatalog,
		containsPgTempSchema: s.containsPgTempSchema,
		containsPgExtension:  s.containsPgExtension,
		userSchemaName:       s.userSchemaName,
		tempSchemaName:       tempSchemaName,
	}
}

func (s SearchPath) WithUserSchemaName(userSchemaName string) SearchPath {
	__antithesis_instrumentation__.Notify(617815)
	return SearchPath{
		paths:                s.paths,
		containsPgCatalog:    s.containsPgCatalog,
		containsPgTempSchema: s.containsPgTempSchema,
		containsPgExtension:  s.containsPgExtension,
		userSchemaName:       userSchemaName,
		tempSchemaName:       s.tempSchemaName,
	}
}

func (s SearchPath) UpdatePaths(paths []string) SearchPath {
	__antithesis_instrumentation__.Notify(617816)
	return MakeSearchPath(paths).WithTemporarySchemaName(s.tempSchemaName).WithUserSchemaName(s.userSchemaName)
}

func (s SearchPath) MaybeResolveTemporarySchema(schemaName string) (string, error) {
	__antithesis_instrumentation__.Notify(617817)

	if strings.HasPrefix(schemaName, catconstants.PgTempSchemaName) && func() bool {
		__antithesis_instrumentation__.Notify(617820)
		return schemaName != catconstants.PgTempSchemaName == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(617821)
		return schemaName != s.tempSchemaName == true
	}() == true {
		__antithesis_instrumentation__.Notify(617822)
		return schemaName, pgerror.New(pgcode.FeatureNotSupported, "cannot access temporary tables of other sessions")
	} else {
		__antithesis_instrumentation__.Notify(617823)
	}
	__antithesis_instrumentation__.Notify(617818)

	if schemaName == catconstants.PgTempSchemaName && func() bool {
		__antithesis_instrumentation__.Notify(617824)
		return s.tempSchemaName != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(617825)
		return s.tempSchemaName, nil
	} else {
		__antithesis_instrumentation__.Notify(617826)
	}
	__antithesis_instrumentation__.Notify(617819)
	return schemaName, nil
}

func (s SearchPath) Iter() SearchPathIter {
	__antithesis_instrumentation__.Notify(617827)
	implicitPgTempSchema := !s.containsPgTempSchema && func() bool {
		__antithesis_instrumentation__.Notify(617828)
		return s.tempSchemaName != "" == true
	}() == true
	sp := SearchPathIter{
		paths:                s.paths,
		implicitPgCatalog:    !s.containsPgCatalog,
		implicitPgExtension:  !s.containsPgExtension,
		implicitPgTempSchema: implicitPgTempSchema,
		tempSchemaName:       s.tempSchemaName,
		userSchemaName:       s.userSchemaName,
	}
	return sp
}

func (s SearchPath) IterWithoutImplicitPGSchemas() SearchPathIter {
	__antithesis_instrumentation__.Notify(617829)
	sp := SearchPathIter{
		paths:                s.paths,
		implicitPgCatalog:    false,
		implicitPgTempSchema: false,
		tempSchemaName:       s.tempSchemaName,
		userSchemaName:       s.userSchemaName,
	}
	return sp
}

func (s SearchPath) GetPathArray() []string {
	__antithesis_instrumentation__.Notify(617830)
	return s.paths
}

func (s SearchPath) Contains(target string) bool {
	__antithesis_instrumentation__.Notify(617831)
	for _, candidate := range s.GetPathArray() {
		__antithesis_instrumentation__.Notify(617833)
		if candidate == target {
			__antithesis_instrumentation__.Notify(617834)
			return true
		} else {
			__antithesis_instrumentation__.Notify(617835)
		}
	}
	__antithesis_instrumentation__.Notify(617832)
	return false
}

func (s SearchPath) GetTemporarySchemaName() string {
	__antithesis_instrumentation__.Notify(617836)
	return s.tempSchemaName
}

func (s SearchPath) Equals(other *SearchPath) bool {
	__antithesis_instrumentation__.Notify(617837)
	if s.containsPgCatalog != other.containsPgCatalog {
		__antithesis_instrumentation__.Notify(617844)
		return false
	} else {
		__antithesis_instrumentation__.Notify(617845)
	}
	__antithesis_instrumentation__.Notify(617838)
	if s.containsPgExtension != other.containsPgExtension {
		__antithesis_instrumentation__.Notify(617846)
		return false
	} else {
		__antithesis_instrumentation__.Notify(617847)
	}
	__antithesis_instrumentation__.Notify(617839)
	if s.containsPgTempSchema != other.containsPgTempSchema {
		__antithesis_instrumentation__.Notify(617848)
		return false
	} else {
		__antithesis_instrumentation__.Notify(617849)
	}
	__antithesis_instrumentation__.Notify(617840)
	if len(s.paths) != len(other.paths) {
		__antithesis_instrumentation__.Notify(617850)
		return false
	} else {
		__antithesis_instrumentation__.Notify(617851)
	}
	__antithesis_instrumentation__.Notify(617841)
	if s.tempSchemaName != other.tempSchemaName {
		__antithesis_instrumentation__.Notify(617852)
		return false
	} else {
		__antithesis_instrumentation__.Notify(617853)
	}
	__antithesis_instrumentation__.Notify(617842)

	if &s.paths[0] != &other.paths[0] {
		__antithesis_instrumentation__.Notify(617854)
		for i := range s.paths {
			__antithesis_instrumentation__.Notify(617855)
			if s.paths[i] != other.paths[i] {
				__antithesis_instrumentation__.Notify(617856)
				return false
			} else {
				__antithesis_instrumentation__.Notify(617857)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(617858)
	}
	__antithesis_instrumentation__.Notify(617843)
	return true
}

func (s SearchPath) SQLIdentifiers() string {
	__antithesis_instrumentation__.Notify(617859)
	var buf bytes.Buffer
	for i, path := range s.paths {
		__antithesis_instrumentation__.Notify(617861)
		if i > 0 {
			__antithesis_instrumentation__.Notify(617863)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(617864)
		}
		__antithesis_instrumentation__.Notify(617862)
		lexbase.EncodeRestrictedSQLIdent(&buf, path, lexbase.EncNoFlags)
	}
	__antithesis_instrumentation__.Notify(617860)
	return buf.String()
}

func (s SearchPath) String() string {
	__antithesis_instrumentation__.Notify(617865)
	return strings.Join(s.paths, ",")
}

type SearchPathIter struct {
	paths                []string
	implicitPgCatalog    bool
	implicitPgExtension  bool
	implicitPgTempSchema bool
	tempSchemaName       string
	userSchemaName       string
	i                    int
}

func (iter *SearchPathIter) Next() (path string, ok bool) {
	__antithesis_instrumentation__.Notify(617866)

	if iter.implicitPgTempSchema && func() bool {
		__antithesis_instrumentation__.Notify(617870)
		return iter.tempSchemaName != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(617871)
		iter.implicitPgTempSchema = false
		return iter.tempSchemaName, true
	} else {
		__antithesis_instrumentation__.Notify(617872)
	}
	__antithesis_instrumentation__.Notify(617867)
	if iter.implicitPgCatalog {
		__antithesis_instrumentation__.Notify(617873)
		iter.implicitPgCatalog = false
		return catconstants.PgCatalogName, true
	} else {
		__antithesis_instrumentation__.Notify(617874)
	}
	__antithesis_instrumentation__.Notify(617868)

	if iter.i < len(iter.paths) {
		__antithesis_instrumentation__.Notify(617875)
		iter.i++

		if iter.paths[iter.i-1] == catconstants.PgTempSchemaName {
			__antithesis_instrumentation__.Notify(617879)

			if iter.tempSchemaName == "" {
				__antithesis_instrumentation__.Notify(617881)
				return iter.Next()
			} else {
				__antithesis_instrumentation__.Notify(617882)
			}
			__antithesis_instrumentation__.Notify(617880)
			return iter.tempSchemaName, true
		} else {
			__antithesis_instrumentation__.Notify(617883)
		}
		__antithesis_instrumentation__.Notify(617876)
		if iter.paths[iter.i-1] == catconstants.UserSchemaName {
			__antithesis_instrumentation__.Notify(617884)

			if iter.userSchemaName == "" {
				__antithesis_instrumentation__.Notify(617886)
				return iter.Next()
			} else {
				__antithesis_instrumentation__.Notify(617887)
			}
			__antithesis_instrumentation__.Notify(617885)
			return iter.userSchemaName, true
		} else {
			__antithesis_instrumentation__.Notify(617888)
		}
		__antithesis_instrumentation__.Notify(617877)

		if iter.paths[iter.i-1] == catconstants.PublicSchemaName && func() bool {
			__antithesis_instrumentation__.Notify(617889)
			return iter.implicitPgExtension == true
		}() == true {
			__antithesis_instrumentation__.Notify(617890)
			iter.implicitPgExtension = false

			iter.i--
			return catconstants.PgExtensionSchemaName, true
		} else {
			__antithesis_instrumentation__.Notify(617891)
		}
		__antithesis_instrumentation__.Notify(617878)
		return iter.paths[iter.i-1], true
	} else {
		__antithesis_instrumentation__.Notify(617892)
	}
	__antithesis_instrumentation__.Notify(617869)
	return "", false
}
