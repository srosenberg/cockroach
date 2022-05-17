package cmdutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"log"
	"os"
)

func RequireEnv(s string) string {
	__antithesis_instrumentation__.Notify(37787)
	v := os.Getenv(s)
	if v == "" {
		__antithesis_instrumentation__.Notify(37789)
		log.Fatalf("missing required environment variable %q", s)
	} else {
		__antithesis_instrumentation__.Notify(37790)
	}
	__antithesis_instrumentation__.Notify(37788)
	return v
}
