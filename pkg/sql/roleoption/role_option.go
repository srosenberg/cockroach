package roleoption

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/errors"
)

type Option uint32

type RoleOption struct {
	Option
	HasValue bool

	Value func() (bool, string, error)
}

const (
	_ Option = iota
	CREATEROLE
	NOCREATEROLE
	PASSWORD
	LOGIN
	NOLOGIN
	VALIDUNTIL
	CONTROLJOB
	NOCONTROLJOB
	CONTROLCHANGEFEED
	NOCONTROLCHANGEFEED
	CREATEDB
	NOCREATEDB
	CREATELOGIN
	NOCREATELOGIN

	VIEWACTIVITY
	NOVIEWACTIVITY
	CANCELQUERY
	NOCANCELQUERY
	MODIFYCLUSTERSETTING
	NOMODIFYCLUSTERSETTING
	DEFAULTSETTINGS
	VIEWACTIVITYREDACTED
	NOVIEWACTIVITYREDACTED
	SQLLOGIN
	NOSQLLOGIN
	VIEWCLUSTERSETTING
	NOVIEWCLUSTERSETTING
)

var toSQLStmts = map[Option]string{
	CREATEROLE:             `UPSERT INTO system.role_options (username, option) VALUES ($1, 'CREATEROLE')`,
	NOCREATEROLE:           `DELETE FROM system.role_options WHERE username = $1 AND option = 'CREATEROLE'`,
	LOGIN:                  `DELETE FROM system.role_options WHERE username = $1 AND option = 'NOLOGIN'`,
	NOLOGIN:                `UPSERT INTO system.role_options (username, option) VALUES ($1, 'NOLOGIN')`,
	VALIDUNTIL:             `UPSERT INTO system.role_options (username, option, value) VALUES ($1, 'VALID UNTIL', $2::timestamptz::string)`,
	CONTROLJOB:             `UPSERT INTO system.role_options (username, option) VALUES ($1, 'CONTROLJOB')`,
	NOCONTROLJOB:           `DELETE FROM system.role_options WHERE username = $1 AND option = 'CONTROLJOB'`,
	CONTROLCHANGEFEED:      `UPSERT INTO system.role_options (username, option) VALUES ($1, 'CONTROLCHANGEFEED')`,
	NOCONTROLCHANGEFEED:    `DELETE FROM system.role_options WHERE username = $1 AND option = 'CONTROLCHANGEFEED'`,
	CREATEDB:               `UPSERT INTO system.role_options (username, option) VALUES ($1, 'CREATEDB')`,
	NOCREATEDB:             `DELETE FROM system.role_options WHERE username = $1 AND option = 'CREATEDB'`,
	CREATELOGIN:            `UPSERT INTO system.role_options (username, option) VALUES ($1, 'CREATELOGIN')`,
	NOCREATELOGIN:          `DELETE FROM system.role_options WHERE username = $1 AND option = 'CREATELOGIN'`,
	VIEWACTIVITY:           `UPSERT INTO system.role_options (username, option) VALUES ($1, 'VIEWACTIVITY')`,
	NOVIEWACTIVITY:         `DELETE FROM system.role_options WHERE username = $1 AND option = 'VIEWACTIVITY'`,
	CANCELQUERY:            `UPSERT INTO system.role_options (username, option) VALUES ($1, 'CANCELQUERY')`,
	NOCANCELQUERY:          `DELETE FROM system.role_options WHERE username = $1 AND option = 'CANCELQUERY'`,
	MODIFYCLUSTERSETTING:   `UPSERT INTO system.role_options (username, option) VALUES ($1, 'MODIFYCLUSTERSETTING')`,
	NOMODIFYCLUSTERSETTING: `DELETE FROM system.role_options WHERE username = $1 AND option = 'MODIFYCLUSTERSETTING'`,
	SQLLOGIN:               `DELETE FROM system.role_options WHERE username = $1 AND option = 'NOSQLLOGIN'`,
	NOSQLLOGIN:             `UPSERT INTO system.role_options (username, option) VALUES ($1, 'NOSQLLOGIN')`,
	VIEWACTIVITYREDACTED:   `UPSERT INTO system.role_options (username, option) VALUES ($1, 'VIEWACTIVITYREDACTED')`,
	NOVIEWACTIVITYREDACTED: `DELETE FROM system.role_options WHERE username = $1 AND option = 'VIEWACTIVITYREDACTED'`,
	VIEWCLUSTERSETTING:     `UPSERT INTO system.role_options (username, option) VALUES ($1, 'VIEWCLUSTERSETTING')`,
	NOVIEWCLUSTERSETTING:   `DELETE FROM system.role_options WHERE username = $1 AND option = 'VIEWCLUSTERSETTING'`,
}

func (o Option) Mask() uint32 {
	__antithesis_instrumentation__.Notify(567257)
	return 1 << o
}

var ByName = map[string]Option{
	"CREATEROLE":             CREATEROLE,
	"NOCREATEROLE":           NOCREATEROLE,
	"PASSWORD":               PASSWORD,
	"LOGIN":                  LOGIN,
	"NOLOGIN":                NOLOGIN,
	"VALID UNTIL":            VALIDUNTIL,
	"CONTROLJOB":             CONTROLJOB,
	"NOCONTROLJOB":           NOCONTROLJOB,
	"CONTROLCHANGEFEED":      CONTROLCHANGEFEED,
	"NOCONTROLCHANGEFEED":    NOCONTROLCHANGEFEED,
	"CREATEDB":               CREATEDB,
	"NOCREATEDB":             NOCREATEDB,
	"CREATELOGIN":            CREATELOGIN,
	"NOCREATELOGIN":          NOCREATELOGIN,
	"VIEWACTIVITY":           VIEWACTIVITY,
	"NOVIEWACTIVITY":         NOVIEWACTIVITY,
	"CANCELQUERY":            CANCELQUERY,
	"NOCANCELQUERY":          NOCANCELQUERY,
	"MODIFYCLUSTERSETTING":   MODIFYCLUSTERSETTING,
	"NOMODIFYCLUSTERSETTING": NOMODIFYCLUSTERSETTING,
	"DEFAULTSETTINGS":        DEFAULTSETTINGS,
	"VIEWACTIVITYREDACTED":   VIEWACTIVITYREDACTED,
	"NOVIEWACTIVITYREDACTED": NOVIEWACTIVITYREDACTED,
	"SQLLOGIN":               SQLLOGIN,
	"NOSQLLOGIN":             NOSQLLOGIN,
	"VIEWCLUSTERSETTING":     VIEWCLUSTERSETTING,
	"NOVIEWCLUSTERSETTING":   NOVIEWCLUSTERSETTING,
}

func ToOption(str string) (Option, error) {
	__antithesis_instrumentation__.Notify(567258)
	ret := ByName[strings.ToUpper(str)]
	if ret == 0 {
		__antithesis_instrumentation__.Notify(567260)
		return 0, pgerror.Newf(pgcode.Syntax, "unrecognized role option %s", str)
	} else {
		__antithesis_instrumentation__.Notify(567261)
	}
	__antithesis_instrumentation__.Notify(567259)

	return ret, nil
}

type List []RoleOption

func (rol List) GetSQLStmts(op string) (map[string]func() (bool, string, error), error) {
	__antithesis_instrumentation__.Notify(567262)
	if len(rol) <= 0 {
		__antithesis_instrumentation__.Notify(567266)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(567267)
	}
	__antithesis_instrumentation__.Notify(567263)

	stmts := make(map[string]func() (bool, string, error), len(rol))

	err := rol.CheckRoleOptionConflicts()
	if err != nil {
		__antithesis_instrumentation__.Notify(567268)
		return stmts, err
	} else {
		__antithesis_instrumentation__.Notify(567269)
	}
	__antithesis_instrumentation__.Notify(567264)

	for _, ro := range rol {
		__antithesis_instrumentation__.Notify(567270)
		sqltelemetry.IncIAMOptionCounter(
			op,
			strings.ToLower(ro.Option.String()),
		)

		if ro.Option == PASSWORD || func() bool {
			__antithesis_instrumentation__.Notify(567272)
			return ro.Option == DEFAULTSETTINGS == true
		}() == true {
			__antithesis_instrumentation__.Notify(567273)
			continue
		} else {
			__antithesis_instrumentation__.Notify(567274)
		}
		__antithesis_instrumentation__.Notify(567271)

		stmt := toSQLStmts[ro.Option]
		if ro.HasValue {
			__antithesis_instrumentation__.Notify(567275)
			stmts[stmt] = ro.Value
		} else {
			__antithesis_instrumentation__.Notify(567276)
			stmts[stmt] = nil
		}
	}
	__antithesis_instrumentation__.Notify(567265)

	return stmts, nil
}

func (rol List) ToBitField() (uint32, error) {
	__antithesis_instrumentation__.Notify(567277)
	var ret uint32
	for _, p := range rol {
		__antithesis_instrumentation__.Notify(567279)
		if ret&p.Option.Mask() != 0 {
			__antithesis_instrumentation__.Notify(567281)
			return 0, pgerror.Newf(pgcode.Syntax, "redundant role options")
		} else {
			__antithesis_instrumentation__.Notify(567282)
		}
		__antithesis_instrumentation__.Notify(567280)
		ret |= p.Option.Mask()
	}
	__antithesis_instrumentation__.Notify(567278)
	return ret, nil
}

func (rol List) Contains(p Option) bool {
	__antithesis_instrumentation__.Notify(567283)
	for _, ro := range rol {
		__antithesis_instrumentation__.Notify(567285)
		if ro.Option == p {
			__antithesis_instrumentation__.Notify(567286)
			return true
		} else {
			__antithesis_instrumentation__.Notify(567287)
		}
	}
	__antithesis_instrumentation__.Notify(567284)

	return false
}

func (rol List) CheckRoleOptionConflicts() error {
	__antithesis_instrumentation__.Notify(567288)
	roleOptionBits, err := rol.ToBitField()

	if err != nil {
		__antithesis_instrumentation__.Notify(567291)
		return err
	} else {
		__antithesis_instrumentation__.Notify(567292)
	}
	__antithesis_instrumentation__.Notify(567289)

	if (roleOptionBits&CREATEROLE.Mask() != 0 && func() bool {
		__antithesis_instrumentation__.Notify(567293)
		return roleOptionBits&NOCREATEROLE.Mask() != 0 == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(567294)
		return (roleOptionBits&LOGIN.Mask() != 0 && func() bool {
			__antithesis_instrumentation__.Notify(567295)
			return roleOptionBits&NOLOGIN.Mask() != 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(567296)
		return (roleOptionBits&CONTROLJOB.Mask() != 0 && func() bool {
			__antithesis_instrumentation__.Notify(567297)
			return roleOptionBits&NOCONTROLJOB.Mask() != 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(567298)
		return (roleOptionBits&CONTROLCHANGEFEED.Mask() != 0 && func() bool {
			__antithesis_instrumentation__.Notify(567299)
			return roleOptionBits&NOCONTROLCHANGEFEED.Mask() != 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(567300)
		return (roleOptionBits&CREATEDB.Mask() != 0 && func() bool {
			__antithesis_instrumentation__.Notify(567301)
			return roleOptionBits&NOCREATEDB.Mask() != 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(567302)
		return (roleOptionBits&CREATELOGIN.Mask() != 0 && func() bool {
			__antithesis_instrumentation__.Notify(567303)
			return roleOptionBits&NOCREATELOGIN.Mask() != 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(567304)
		return (roleOptionBits&VIEWACTIVITY.Mask() != 0 && func() bool {
			__antithesis_instrumentation__.Notify(567305)
			return roleOptionBits&NOVIEWACTIVITY.Mask() != 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(567306)
		return (roleOptionBits&CANCELQUERY.Mask() != 0 && func() bool {
			__antithesis_instrumentation__.Notify(567307)
			return roleOptionBits&NOCANCELQUERY.Mask() != 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(567308)
		return (roleOptionBits&MODIFYCLUSTERSETTING.Mask() != 0 && func() bool {
			__antithesis_instrumentation__.Notify(567309)
			return roleOptionBits&NOMODIFYCLUSTERSETTING.Mask() != 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(567310)
		return (roleOptionBits&VIEWACTIVITYREDACTED.Mask() != 0 && func() bool {
			__antithesis_instrumentation__.Notify(567311)
			return roleOptionBits&NOVIEWACTIVITYREDACTED.Mask() != 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(567312)
		return (roleOptionBits&SQLLOGIN.Mask() != 0 && func() bool {
			__antithesis_instrumentation__.Notify(567313)
			return roleOptionBits&NOSQLLOGIN.Mask() != 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(567314)
		return (roleOptionBits&VIEWCLUSTERSETTING.Mask() != 0 && func() bool {
			__antithesis_instrumentation__.Notify(567315)
			return roleOptionBits&NOVIEWCLUSTERSETTING.Mask() != 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(567316)
		return pgerror.Newf(pgcode.Syntax, "conflicting role options")
	} else {
		__antithesis_instrumentation__.Notify(567317)
	}
	__antithesis_instrumentation__.Notify(567290)
	return nil
}

func (rol List) GetPassword() (isNull bool, password string, err error) {
	__antithesis_instrumentation__.Notify(567318)
	for _, ro := range rol {
		__antithesis_instrumentation__.Notify(567320)
		if ro.Option == PASSWORD {
			__antithesis_instrumentation__.Notify(567321)
			return ro.Value()
		} else {
			__antithesis_instrumentation__.Notify(567322)
		}
	}
	__antithesis_instrumentation__.Notify(567319)

	return false, "", errors.New("password not found in role options")
}
