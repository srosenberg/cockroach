package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [options] <session-username>",
	Short: "create a HTTP session and token for the given user",
	Long: `
Creates a HTTP session for the given user and print out a login cookie for use
in non-interactive programs.

Example use of the session cookie using 'curl':

   curl -k -b "<cookie>" https://localhost:8080/_admin/v1/settings

The user invoking the 'login' CLI command must be an admin on the cluster.
The user for which the HTTP session is opened can be arbitrary.
`,
	Args: cobra.ExactArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runLogin),
}

func runLogin(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(27948)

	username := tree.Name(args[0]).Normalize()

	id, httpCookie, err := createAuthSessionToken(username)
	if err != nil {
		__antithesis_instrumentation__.Notify(27951)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27952)
	}
	__antithesis_instrumentation__.Notify(27949)
	hC := httpCookie.String()

	if authCtx.onlyCookie {
		__antithesis_instrumentation__.Notify(27953)

		fmt.Println(hC)
	} else {
		__antithesis_instrumentation__.Notify(27954)

		cols := []string{"username", "session ID", "authentication cookie"}
		rows := [][]string{
			{username, fmt.Sprintf("%d", id), hC},
		}
		if err := sqlExecCtx.PrintQueryOutput(os.Stdout, stderr, cols, clisqlexec.NewRowSliceIter(rows, "ll")); err != nil {
			__antithesis_instrumentation__.Notify(27956)
			return err
		} else {
			__antithesis_instrumentation__.Notify(27957)
		}
		__antithesis_instrumentation__.Notify(27955)

		if isatty.IsTerminal(os.Stdin.Fd()) {
			__antithesis_instrumentation__.Notify(27958)
			fmt.Fprintf(stderr, `#
# Example uses:
#
#     curl [-k] --cookie '%[1]s' https://...
#
#     wget [--no-check-certificate] --header='Cookie: %[1]s' https://...
#
`, hC)
		} else {
			__antithesis_instrumentation__.Notify(27959)
		}
	}
	__antithesis_instrumentation__.Notify(27950)

	return nil
}

func createAuthSessionToken(
	username string,
) (sessionID int64, httpCookie *http.Cookie, resErr error) {
	__antithesis_instrumentation__.Notify(27960)
	sqlConn, err := makeSQLClient("cockroach auth-session login", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(27969)
		return -1, nil, err
	} else {
		__antithesis_instrumentation__.Notify(27970)
	}
	__antithesis_instrumentation__.Notify(27961)
	defer func() {
		__antithesis_instrumentation__.Notify(27971)
		resErr = errors.CombineErrors(resErr, sqlConn.Close())
	}()
	__antithesis_instrumentation__.Notify(27962)

	ctx := context.Background()

	_, rows, err := sqlExecCtx.RunQuery(ctx,
		sqlConn,
		clisqlclient.MakeQuery(`SELECT count(username) FROM system.users WHERE username = $1 AND NOT "isRole"`, username), false)
	if err != nil {
		__antithesis_instrumentation__.Notify(27972)
		return -1, nil, err
	} else {
		__antithesis_instrumentation__.Notify(27973)
	}
	__antithesis_instrumentation__.Notify(27963)
	if rows[0][0] != "1" {
		__antithesis_instrumentation__.Notify(27974)
		return -1, nil, fmt.Errorf("user %q does not exist", username)
	} else {
		__antithesis_instrumentation__.Notify(27975)
	}
	__antithesis_instrumentation__.Notify(27964)

	secret, hashedSecret, err := server.CreateAuthSecret()
	if err != nil {
		__antithesis_instrumentation__.Notify(27976)
		return -1, nil, err
	} else {
		__antithesis_instrumentation__.Notify(27977)
	}
	__antithesis_instrumentation__.Notify(27965)
	expiration := timeutil.Now().Add(authCtx.validityPeriod)

	insertSessionStmt := `
INSERT INTO system.web_sessions ("hashedSecret", username, "expiresAt")
VALUES($1, $2, $3)
RETURNING id
`
	var id int64
	row, err := sqlConn.QueryRow(ctx,
		insertSessionStmt,
		hashedSecret,
		username,
		expiration,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(27978)
		return -1, nil, err
	} else {
		__antithesis_instrumentation__.Notify(27979)
	}
	__antithesis_instrumentation__.Notify(27966)
	if len(row) != 1 {
		__antithesis_instrumentation__.Notify(27980)
		return -1, nil, errors.Newf("expected 1 column, got %d", len(row))
	} else {
		__antithesis_instrumentation__.Notify(27981)
	}
	__antithesis_instrumentation__.Notify(27967)
	id, ok := row[0].(int64)
	if !ok {
		__antithesis_instrumentation__.Notify(27982)
		return -1, nil, errors.Newf("expected integer, got %T", row[0])
	} else {
		__antithesis_instrumentation__.Notify(27983)
	}
	__antithesis_instrumentation__.Notify(27968)

	sCookie := &serverpb.SessionCookie{ID: id, Secret: secret}
	httpCookie, err = server.EncodeSessionCookie(sCookie, false)
	return id, httpCookie, err
}

var logoutCmd = &cobra.Command{
	Use:   "logout [options] <session-username>",
	Short: "invalidates all the HTTP session tokens previously created for the given user",
	Long: `
Revokes all previously issued HTTP authentication tokens for the given user.

The user invoking the 'login' CLI command must be an admin on the cluster.
The user for which the HTTP sessions are revoked can be arbitrary.
`,
	Args: cobra.ExactArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runLogout),
}

func runLogout(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(27984)
	username := tree.Name(args[0]).Normalize()

	sqlConn, err := makeSQLClient("cockroach auth-session logout", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(27987)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27988)
	}
	__antithesis_instrumentation__.Notify(27985)
	defer func() {
		__antithesis_instrumentation__.Notify(27989)
		resErr = errors.CombineErrors(resErr, sqlConn.Close())
	}()
	__antithesis_instrumentation__.Notify(27986)

	logoutQuery := clisqlclient.MakeQuery(
		`UPDATE system.web_sessions SET "revokedAt" = if("revokedAt"::timestamptz<now(),"revokedAt",now())
      WHERE username = $1
  RETURNING username,
            id AS "session ID",
            "revokedAt" AS "revoked"`,
		username)
	return sqlExecCtx.RunQueryAndFormatResults(
		context.Background(),
		sqlConn, os.Stdout, stderr, logoutQuery)
}

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "lists the currently active HTTP sessions",
	Long: `
Prints out the currently active HTTP sessions.

The user invoking the 'list' CLI command must be an admin on the cluster.
`,
	Args: cobra.ExactArgs(0),
	RunE: clierrorplus.MaybeDecorateError(runAuthList),
}

func runAuthList(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(27990)
	sqlConn, err := makeSQLClient("cockroach auth-session list", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(27993)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27994)
	}
	__antithesis_instrumentation__.Notify(27991)
	defer func() {
		__antithesis_instrumentation__.Notify(27995)
		resErr = errors.CombineErrors(resErr, sqlConn.Close())
	}()
	__antithesis_instrumentation__.Notify(27992)

	logoutQuery := clisqlclient.MakeQuery(`
SELECT username,
       id AS "session ID",
       "createdAt" as "created",
       "expiresAt" as "expires",
       "revokedAt" as "revoked",
       "lastUsedAt" as "last used"
  FROM system.web_sessions`)
	return sqlExecCtx.RunQueryAndFormatResults(
		context.Background(),
		sqlConn, os.Stdout, stderr, logoutQuery)
}

var authCmds = []*cobra.Command{
	loginCmd,
	logoutCmd,
	authListCmd,
}

var authCmd = &cobra.Command{
	Use:   "auth-session",
	Short: "log in and out of HTTP sessions",
	RunE:  UsageAndErr,
}

func init() {
	authCmd.AddCommand(authCmds...)
}
