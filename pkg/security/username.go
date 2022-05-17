package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type SQLUsername struct {
	u string
}

const NodeUser = "node"

func NodeUserName() SQLUsername {
	__antithesis_instrumentation__.Notify(187338)
	return SQLUsername{NodeUser}
}

func (s SQLUsername) IsNodeUser() bool {
	__antithesis_instrumentation__.Notify(187339)
	return s.u == NodeUser
}

const RootUser = "root"

func RootUserName() SQLUsername {
	__antithesis_instrumentation__.Notify(187340)
	return SQLUsername{RootUser}
}

func (s SQLUsername) IsRootUser() bool {
	__antithesis_instrumentation__.Notify(187341)
	return s.u == RootUser
}

const AdminRole = "admin"

func AdminRoleName() SQLUsername {
	__antithesis_instrumentation__.Notify(187342)
	return SQLUsername{AdminRole}
}

func (s SQLUsername) IsAdminRole() bool {
	__antithesis_instrumentation__.Notify(187343)
	return s.u == AdminRole
}

const PublicRole = "public"

const NoneRole = "none"

func (s SQLUsername) IsNoneRole() bool {
	__antithesis_instrumentation__.Notify(187344)
	return s.u == NoneRole
}

func PublicRoleName() SQLUsername {
	__antithesis_instrumentation__.Notify(187345)
	return SQLUsername{PublicRole}
}

func (s SQLUsername) IsPublicRole() bool {
	__antithesis_instrumentation__.Notify(187346)
	return s.u == PublicRole
}

func (s SQLUsername) IsReserved() bool {
	__antithesis_instrumentation__.Notify(187347)
	return s.IsPublicRole() || func() bool {
		__antithesis_instrumentation__.Notify(187348)
		return s.u == NoneRole == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(187349)
		return s.IsNodeUser() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(187350)
		return strings.HasPrefix(s.u, "pg_") == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(187351)
		return strings.HasPrefix(s.u, "crdb_internal_") == true
	}() == true
}

func (s SQLUsername) Undefined() bool {
	__antithesis_instrumentation__.Notify(187352)
	return len(s.u) == 0
}

const TestUser = "testuser"

func TestUserName() SQLUsername {
	__antithesis_instrumentation__.Notify(187353)
	return SQLUsername{TestUser}
}

func MakeSQLUsernameFromUserInput(u string, purpose UsernamePurpose) (res SQLUsername, err error) {
	__antithesis_instrumentation__.Notify(187354)

	res.u = lexbase.NormalizeName(u)
	if purpose == UsernameCreation {
		__antithesis_instrumentation__.Notify(187356)
		err = res.ValidateForCreation()
	} else {
		__antithesis_instrumentation__.Notify(187357)
	}
	__antithesis_instrumentation__.Notify(187355)
	return res, err
}

type UsernamePurpose bool

const (
	UsernameCreation UsernamePurpose = false

	UsernameValidation UsernamePurpose = true
)

const usernameHelp = "Usernames are case insensitive, must start with a letter, " +
	"digit or underscore, may contain letters, digits, dashes, periods, or underscores, and must not exceed 63 characters."

const maxUsernameLengthForCreation = 63

var validUsernameCreationRE = regexp.MustCompile(`^[\p{Ll}0-9_][---\p{Ll}0-9_.]*$`)

func (s SQLUsername) ValidateForCreation() error {
	__antithesis_instrumentation__.Notify(187358)
	if len(s.u) == 0 {
		__antithesis_instrumentation__.Notify(187362)
		return ErrUsernameEmpty
	} else {
		__antithesis_instrumentation__.Notify(187363)
	}
	__antithesis_instrumentation__.Notify(187359)
	if len(s.u) > maxUsernameLengthForCreation {
		__antithesis_instrumentation__.Notify(187364)
		return ErrUsernameTooLong
	} else {
		__antithesis_instrumentation__.Notify(187365)
	}
	__antithesis_instrumentation__.Notify(187360)
	if !validUsernameCreationRE.MatchString(s.u) {
		__antithesis_instrumentation__.Notify(187366)
		return ErrUsernameInvalid
	} else {
		__antithesis_instrumentation__.Notify(187367)
	}
	__antithesis_instrumentation__.Notify(187361)
	return nil
}

var ErrUsernameTooLong = errors.WithHint(errors.New("username is too long"), usernameHelp)

var ErrUsernameInvalid = errors.WithHint(errors.New("username is invalid"), usernameHelp)

var ErrUsernameEmpty = errors.WithHint(errors.New("username is empty"), usernameHelp)

var ErrUsernameNotNormalized = errors.WithHint(errors.New("username is not normalized"),
	"The username should be converted to lowercase and unicode characters normalized to NFC.")

func MakeSQLUsernameFromPreNormalizedString(u string) SQLUsername {
	__antithesis_instrumentation__.Notify(187368)
	return SQLUsername{u: u}
}

func MakeSQLUsernameFromPreNormalizedStringChecked(u string) (SQLUsername, error) {
	__antithesis_instrumentation__.Notify(187369)
	res := SQLUsername{u: lexbase.NormalizeName(u)}
	if res.u != u {
		__antithesis_instrumentation__.Notify(187371)
		return res, ErrUsernameNotNormalized
	} else {
		__antithesis_instrumentation__.Notify(187372)
	}
	__antithesis_instrumentation__.Notify(187370)
	return res, nil
}

func (s SQLUsername) Normalized() string { __antithesis_instrumentation__.Notify(187373); return s.u }

func (s SQLUsername) SQLIdentifier() string {
	__antithesis_instrumentation__.Notify(187374)
	var buf bytes.Buffer
	lexbase.EncodeRestrictedSQLIdent(&buf, s.u, lexbase.EncNoFlags)
	return buf.String()
}

func (s SQLUsername) Format(fs fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(187375)
	_, f := redact.MakeFormat(fs, verb)
	fmt.Fprintf(fs, f, s.u)
}

func (s SQLUsername) LessThan(u SQLUsername) bool {
	__antithesis_instrumentation__.Notify(187376)
	return s.u < u.u
}

type SQLUsernameProto string

func (s SQLUsernameProto) Decode() SQLUsername {
	__antithesis_instrumentation__.Notify(187377)
	return SQLUsername{u: string(s)}
}

func (s SQLUsername) EncodeProto() SQLUsernameProto {
	__antithesis_instrumentation__.Notify(187378)
	return SQLUsernameProto(s.u)
}
