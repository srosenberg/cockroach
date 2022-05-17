package clisqlclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/errors"
)

func ParseBool(s string) (bool, error) {
	__antithesis_instrumentation__.Notify(28825)
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "true", "on", "yes", "1":
		__antithesis_instrumentation__.Notify(28826)
		return true, nil
	case "false", "off", "no", "0":
		__antithesis_instrumentation__.Notify(28827)
		return false, nil
	default:
		__antithesis_instrumentation__.Notify(28828)
		return false, errors.Newf("invalid boolean value %q", s)
	}
}
