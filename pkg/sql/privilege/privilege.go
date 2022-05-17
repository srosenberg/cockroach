package privilege

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/errors"
)

type Kind uint32

const (
	ALL        Kind = 1
	CREATE     Kind = 2
	DROP       Kind = 3
	GRANT      Kind = 4
	SELECT     Kind = 5
	INSERT     Kind = 6
	DELETE     Kind = 7
	UPDATE     Kind = 8
	USAGE      Kind = 9
	ZONECONFIG Kind = 10
	CONNECT    Kind = 11
	RULE       Kind = 12
)

type Privilege struct {
	Kind Kind

	GrantOption bool
}

type ObjectType string

const (
	Any ObjectType = "any"

	Database ObjectType = "database"

	Schema ObjectType = "schema"

	Table ObjectType = "table"

	Type ObjectType = "type"
)

var (
	AllPrivileges    = List{ALL, CONNECT, CREATE, DROP, GRANT, SELECT, INSERT, DELETE, UPDATE, USAGE, ZONECONFIG}
	ReadData         = List{GRANT, SELECT}
	ReadWriteData    = List{GRANT, SELECT, INSERT, DELETE, UPDATE}
	DBPrivileges     = List{ALL, CONNECT, CREATE, DROP, GRANT, ZONECONFIG}
	TablePrivileges  = List{ALL, CREATE, DROP, GRANT, SELECT, INSERT, DELETE, UPDATE, ZONECONFIG}
	SchemaPrivileges = List{ALL, GRANT, CREATE, USAGE}
	TypePrivileges   = List{ALL, GRANT, USAGE}
)

func (k Kind) Mask() uint32 {
	__antithesis_instrumentation__.Notify(563254)
	return 1 << k
}

func (k Kind) IsSetIn(bits uint32) bool {
	__antithesis_instrumentation__.Notify(563255)
	return bits&k.Mask() != 0
}

var ByValue = [...]Kind{
	ALL, CREATE, DROP, GRANT, SELECT, INSERT, DELETE, UPDATE, USAGE, ZONECONFIG, CONNECT, RULE,
}

var ByName = map[string]Kind{
	"ALL":        ALL,
	"CONNECT":    CONNECT,
	"CREATE":     CREATE,
	"DROP":       DROP,
	"GRANT":      GRANT,
	"SELECT":     SELECT,
	"INSERT":     INSERT,
	"DELETE":     DELETE,
	"UPDATE":     UPDATE,
	"ZONECONFIG": ZONECONFIG,
	"USAGE":      USAGE,
	"RULE":       RULE,
}

type List []Kind

func (pl List) Len() int {
	__antithesis_instrumentation__.Notify(563256)
	return len(pl)
}

func (pl List) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(563257)
	pl[i], pl[j] = pl[j], pl[i]
}

func (pl List) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(563258)
	return pl[i] < pl[j]
}

func (pl List) names() []string {
	__antithesis_instrumentation__.Notify(563259)
	ret := make([]string, len(pl))
	for i, p := range pl {
		__antithesis_instrumentation__.Notify(563261)
		ret[i] = p.String()
	}
	__antithesis_instrumentation__.Notify(563260)
	return ret
}

func (pl List) Format(buf *bytes.Buffer) {
	__antithesis_instrumentation__.Notify(563262)
	for i, p := range pl {
		__antithesis_instrumentation__.Notify(563263)
		if i > 0 {
			__antithesis_instrumentation__.Notify(563265)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(563266)
		}
		__antithesis_instrumentation__.Notify(563264)
		buf.WriteString(p.String())
	}
}

func (pl List) String() string {
	__antithesis_instrumentation__.Notify(563267)
	return strings.Join(pl.names(), ", ")
}

func (pl List) SortedString() string {
	__antithesis_instrumentation__.Notify(563268)
	names := pl.SortedNames()
	return strings.Join(names, ",")
}

func (pl List) SortedNames() []string {
	__antithesis_instrumentation__.Notify(563269)
	names := pl.names()
	sort.Strings(names)
	return names
}

func (pl List) ToBitField() uint32 {
	__antithesis_instrumentation__.Notify(563270)
	var ret uint32
	for _, p := range pl {
		__antithesis_instrumentation__.Notify(563272)
		ret |= p.Mask()
	}
	__antithesis_instrumentation__.Notify(563271)
	return ret
}

func (pl List) Contains(k Kind) bool {
	__antithesis_instrumentation__.Notify(563273)
	for _, p := range pl {
		__antithesis_instrumentation__.Notify(563275)
		if p == k {
			__antithesis_instrumentation__.Notify(563276)
			return true
		} else {
			__antithesis_instrumentation__.Notify(563277)
		}
	}
	__antithesis_instrumentation__.Notify(563274)
	return false
}

func ListFromBitField(m uint32, objectType ObjectType) List {
	__antithesis_instrumentation__.Notify(563278)
	ret := List{}

	privileges := GetValidPrivilegesForObject(objectType)

	for _, p := range privileges {
		__antithesis_instrumentation__.Notify(563280)
		if m&p.Mask() != 0 {
			__antithesis_instrumentation__.Notify(563281)
			ret = append(ret, p)
		} else {
			__antithesis_instrumentation__.Notify(563282)
		}
	}
	__antithesis_instrumentation__.Notify(563279)
	return ret
}

func PrivilegesFromBitFields(
	kindBits uint32, grantOptionBits uint32, objectType ObjectType,
) []Privilege {
	__antithesis_instrumentation__.Notify(563283)
	var ret []Privilege

	kinds := GetValidPrivilegesForObject(objectType)

	for _, kind := range kinds {
		__antithesis_instrumentation__.Notify(563285)
		if mask := kind.Mask(); kindBits&mask != 0 {
			__antithesis_instrumentation__.Notify(563286)
			ret = append(ret, Privilege{
				Kind:        kind,
				GrantOption: grantOptionBits&mask != 0,
			})
		} else {
			__antithesis_instrumentation__.Notify(563287)
		}
	}
	__antithesis_instrumentation__.Notify(563284)
	return ret
}

func ListFromStrings(strs []string) (List, error) {
	__antithesis_instrumentation__.Notify(563288)
	ret := make(List, len(strs))
	for i, s := range strs {
		__antithesis_instrumentation__.Notify(563290)
		k, ok := ByName[strings.ToUpper(s)]
		if !ok {
			__antithesis_instrumentation__.Notify(563292)
			return nil, errors.Errorf("not a valid privilege: %q", s)
		} else {
			__antithesis_instrumentation__.Notify(563293)
		}
		__antithesis_instrumentation__.Notify(563291)
		ret[i] = k
	}
	__antithesis_instrumentation__.Notify(563289)
	return ret, nil
}

func ValidatePrivileges(privileges List, objectType ObjectType) error {
	__antithesis_instrumentation__.Notify(563294)
	validPrivs := GetValidPrivilegesForObject(objectType)
	for _, priv := range privileges {
		__antithesis_instrumentation__.Notify(563296)
		if validPrivs.ToBitField()&priv.Mask() == 0 {
			__antithesis_instrumentation__.Notify(563297)
			return pgerror.Newf(pgcode.InvalidGrantOperation,
				"invalid privilege type %s for %s", priv.String(), objectType)
		} else {
			__antithesis_instrumentation__.Notify(563298)
		}
	}
	__antithesis_instrumentation__.Notify(563295)

	return nil
}

func GetValidPrivilegesForObject(objectType ObjectType) List {
	__antithesis_instrumentation__.Notify(563299)
	switch objectType {
	case Table:
		__antithesis_instrumentation__.Notify(563300)
		return TablePrivileges
	case Schema:
		__antithesis_instrumentation__.Notify(563301)
		return SchemaPrivileges
	case Database:
		__antithesis_instrumentation__.Notify(563302)
		return DBPrivileges
	case Type:
		__antithesis_instrumentation__.Notify(563303)
		return TypePrivileges
	case Any:
		__antithesis_instrumentation__.Notify(563304)
		return AllPrivileges
	default:
		__antithesis_instrumentation__.Notify(563305)
		panic(errors.AssertionFailedf("unknown object type %s", objectType))
	}
}

var privToACL = map[Kind]string{
	CREATE:  "C",
	SELECT:  "r",
	INSERT:  "a",
	DELETE:  "d",
	UPDATE:  "w",
	USAGE:   "U",
	CONNECT: "c",
}

var orderedPrivs = List{CREATE, USAGE, INSERT, CONNECT, DELETE, SELECT, UPDATE}

func (pl List) ListToACL(grantOptions List, objectType ObjectType) string {
	__antithesis_instrumentation__.Notify(563306)
	privileges := pl

	if pl.Contains(ALL) {
		__antithesis_instrumentation__.Notify(563309)
		privileges = GetValidPrivilegesForObject(objectType)
		if grantOptions.Contains(ALL) {
			__antithesis_instrumentation__.Notify(563310)
			grantOptions = GetValidPrivilegesForObject(objectType)
		} else {
			__antithesis_instrumentation__.Notify(563311)
		}
	} else {
		__antithesis_instrumentation__.Notify(563312)
	}
	__antithesis_instrumentation__.Notify(563307)
	chars := make([]string, len(privileges))
	for _, privilege := range orderedPrivs {
		__antithesis_instrumentation__.Notify(563313)
		if _, ok := privToACL[privilege]; !ok {
			__antithesis_instrumentation__.Notify(563316)
			panic(errors.AssertionFailedf("unknown privilege type %s", privilege.String()))
		} else {
			__antithesis_instrumentation__.Notify(563317)
		}
		__antithesis_instrumentation__.Notify(563314)
		if privileges.Contains(privilege) {
			__antithesis_instrumentation__.Notify(563318)
			chars = append(chars, privToACL[privilege])
		} else {
			__antithesis_instrumentation__.Notify(563319)
		}
		__antithesis_instrumentation__.Notify(563315)
		if grantOptions.Contains(privilege) {
			__antithesis_instrumentation__.Notify(563320)
			chars = append(chars, "*")
		} else {
			__antithesis_instrumentation__.Notify(563321)
		}
	}
	__antithesis_instrumentation__.Notify(563308)

	return strings.Join(chars, "")

}
