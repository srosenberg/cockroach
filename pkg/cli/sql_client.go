package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

var sqlConnTimeout = envutil.EnvOrDefaultString("COCKROACH_CONNECT_TIMEOUT", "15")

type defaultSQLDb int

const (
	useSystemDb defaultSQLDb = iota

	useDefaultDb
)

func makeSQLClient(appName string, defaultMode defaultSQLDb) (clisqlclient.Conn, error) {
	__antithesis_instrumentation__.Notify(33930)
	baseURL, err := cliCtx.makeClientConnURL()
	if err != nil {
		__antithesis_instrumentation__.Notify(33936)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(33937)
	}
	__antithesis_instrumentation__.Notify(33931)

	sqlCtx.ConnectTimeout, err = strconv.Atoi(sqlConnTimeout)
	if err != nil {
		__antithesis_instrumentation__.Notify(33938)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(33939)
	}
	__antithesis_instrumentation__.Notify(33932)

	if defaultMode == useSystemDb {
		__antithesis_instrumentation__.Notify(33940)

		sqlCtx.Database = "system"
	} else {
		__antithesis_instrumentation__.Notify(33941)
	}
	__antithesis_instrumentation__.Notify(33933)

	sqlCtx.User = security.RootUser

	sqlCtx.ApplicationName = catconstants.ReportableAppNamePrefix + appName

	usePw, _, _ := baseURL.GetAuthnPassword()
	if usePw && func() bool {
		__antithesis_instrumentation__.Notify(33942)
		return cliCtx.Insecure == true
	}() == true {
		__antithesis_instrumentation__.Notify(33943)

		return nil, errors.Errorf("password authentication not enabled in insecure mode")
	} else {
		__antithesis_instrumentation__.Notify(33944)
	}
	__antithesis_instrumentation__.Notify(33934)

	sqlURL := baseURL.ToPQ().String()

	if log.V(2) {
		__antithesis_instrumentation__.Notify(33945)
		log.Infof(context.Background(), "connecting with URL: %s", sqlURL)
	} else {
		__antithesis_instrumentation__.Notify(33946)
	}
	__antithesis_instrumentation__.Notify(33935)

	return sqlCtx.MakeConn(sqlURL)
}
