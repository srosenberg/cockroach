package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/pgurl"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var importDumpFileCmd = &cobra.Command{
	Use:   "db <format> <source>",
	Short: "import a pgdump or mysqldump file into CockroachDB",
	Long: `
Uploads and imports a local dump file into the cockroach cluster via userfile storage.
`,
	Args: cobra.MinimumNArgs(2),
	RunE: clierrorplus.MaybeShoutError(runDumpFileImport),
}

var importDumpTableCmd = &cobra.Command{
	Use:   "table <table> <format> <source>",
	Short: "import a table from a pgdump or mysqldump file into CockroachDB",
	Long: `
Uploads and imports a table from the local dump file into the cockroach cluster via userfile storage.
`,
	Args: cobra.MinimumNArgs(3),
	RunE: clierrorplus.MaybeShoutError(runDumpTableImport),
}

type importCLITestingKnobs struct {
	returnQuery      bool
	pauseAfterUpload chan struct{}
	uploadComplete   chan struct{}
}

var importCLIKnobs importCLITestingKnobs

type importMode int

const (
	multiTable importMode = iota
	singleTable
)

func setImportCLITestingKnobs() (importCLITestingKnobs, func()) {
	__antithesis_instrumentation__.Notify(33113)
	importCLIKnobs = importCLITestingKnobs{
		pauseAfterUpload: make(chan struct{}, 1),
		uploadComplete:   make(chan struct{}, 1),
		returnQuery:      true,
	}

	return importCLIKnobs, func() {
		__antithesis_instrumentation__.Notify(33114)
		importCLIKnobs = importCLITestingKnobs{}
	}
}

func runDumpTableImport(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(33115)
	tableName := args[0]
	importFormat := strings.ToLower(args[1])
	source := args[2]
	conn, err := makeSQLClient("cockroach import table", useDefaultDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(33118)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33119)
	}
	__antithesis_instrumentation__.Notify(33116)
	defer func() {
		__antithesis_instrumentation__.Notify(33120)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(33117)
	ctx := context.Background()
	return runImport(ctx, conn, importFormat, source, tableName, singleTable)
}

func runDumpFileImport(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(33121)
	importFormat := strings.ToLower(args[0])
	source := args[1]
	conn, err := makeSQLClient("cockroach import db", useDefaultDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(33124)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33125)
	}
	__antithesis_instrumentation__.Notify(33122)
	defer func() {
		__antithesis_instrumentation__.Notify(33126)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(33123)
	ctx := context.Background()
	return runImport(ctx, conn, importFormat, source, "", multiTable)
}

func runImport(
	ctx context.Context,
	conn clisqlclient.Conn,
	importFormat, source, tableName string,
	mode importMode,
) error {
	__antithesis_instrumentation__.Notify(33127)
	if err := conn.EnsureConn(); err != nil {
		__antithesis_instrumentation__.Notify(33141)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33142)
	}
	__antithesis_instrumentation__.Notify(33128)

	connURL, err := url.Parse(conn.GetURL())
	if err != nil {
		__antithesis_instrumentation__.Notify(33143)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33144)
	}
	__antithesis_instrumentation__.Notify(33129)

	username, err := security.MakeSQLUsernameFromUserInput(connURL.User.Username(), security.UsernameCreation)
	if err != nil {
		__antithesis_instrumentation__.Notify(33145)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33146)
	}
	__antithesis_instrumentation__.Notify(33130)

	userfileDestinationURI := constructUserfileDestinationURI(source, "", username)
	unescapedUserfileURL, err := url.PathUnescape(userfileDestinationURI)
	if err != nil {
		__antithesis_instrumentation__.Notify(33147)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33148)
	}
	__antithesis_instrumentation__.Notify(33131)

	defer func() {
		__antithesis_instrumentation__.Notify(33149)

		_, _ = deleteUserFile(ctx, conn, unescapedUserfileURL)
	}()
	__antithesis_instrumentation__.Notify(33132)

	_, err = uploadUserFile(ctx, conn, source, userfileDestinationURI)
	if err != nil {
		__antithesis_instrumentation__.Notify(33150)
		return errors.Wrap(err, "failed to upload file to userfile before importing")
	} else {
		__antithesis_instrumentation__.Notify(33151)
	}
	__antithesis_instrumentation__.Notify(33133)

	if importCLIKnobs.uploadComplete != nil {
		__antithesis_instrumentation__.Notify(33152)
		importCLIKnobs.uploadComplete <- struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(33153)
	}
	__antithesis_instrumentation__.Notify(33134)

	if importCLIKnobs.pauseAfterUpload != nil {
		__antithesis_instrumentation__.Notify(33154)
		<-importCLIKnobs.pauseAfterUpload
	} else {
		__antithesis_instrumentation__.Notify(33155)
	}
	__antithesis_instrumentation__.Notify(33135)

	ex := conn.GetDriverConn()
	importCompletedMesssage := func() {
		__antithesis_instrumentation__.Notify(33156)
		switch mode {
		case singleTable:
			__antithesis_instrumentation__.Notify(33157)
			fmt.Printf("successfully imported table %s from %s file %s\n", tableName, importFormat,
				source)
		case multiTable:
			__antithesis_instrumentation__.Notify(33158)
			fmt.Printf("successfully imported %s file %s\n", importFormat, source)
		default:
			__antithesis_instrumentation__.Notify(33159)
		}
	}
	__antithesis_instrumentation__.Notify(33136)

	var importQuery string
	switch importFormat {
	case "pgdump":
		__antithesis_instrumentation__.Notify(33160)
		optionsClause := fmt.Sprintf("WITH max_row_size='%d'", importCtx.maxRowSize)
		if importCtx.skipForeignKeys {
			__antithesis_instrumentation__.Notify(33169)
			optionsClause = optionsClause + ", skip_foreign_keys"
		} else {
			__antithesis_instrumentation__.Notify(33170)
		}
		__antithesis_instrumentation__.Notify(33161)
		if importCtx.rowLimit > 0 {
			__antithesis_instrumentation__.Notify(33171)
			optionsClause = fmt.Sprintf("%s, row_limit='%d'", optionsClause, importCtx.rowLimit)
		} else {
			__antithesis_instrumentation__.Notify(33172)
		}
		__antithesis_instrumentation__.Notify(33162)
		if importCtx.ignoreUnsupported {
			__antithesis_instrumentation__.Notify(33173)
			optionsClause = fmt.Sprintf("%s, ignore_unsupported_statements", optionsClause)
		} else {
			__antithesis_instrumentation__.Notify(33174)
		}
		__antithesis_instrumentation__.Notify(33163)
		if importCtx.ignoreUnsupportedLog != "" {
			__antithesis_instrumentation__.Notify(33175)
			optionsClause = fmt.Sprintf("%s, log_ignored_statements=%s", optionsClause,
				importCtx.ignoreUnsupportedLog)
		} else {
			__antithesis_instrumentation__.Notify(33176)
		}
		__antithesis_instrumentation__.Notify(33164)
		switch mode {
		case singleTable:
			__antithesis_instrumentation__.Notify(33177)
			importQuery = fmt.Sprintf(`IMPORT TABLE %s FROM PGDUMP '%s' %s`, tableName,
				unescapedUserfileURL, optionsClause)
		case multiTable:
			__antithesis_instrumentation__.Notify(33178)
			importQuery = fmt.Sprintf(`IMPORT PGDUMP '%s' %s`, unescapedUserfileURL, optionsClause)
		default:
			__antithesis_instrumentation__.Notify(33179)
		}
	case "mysqldump":
		__antithesis_instrumentation__.Notify(33165)
		var optionsClause string
		if importCtx.skipForeignKeys {
			__antithesis_instrumentation__.Notify(33180)
			optionsClause = " WITH skip_foreign_keys"
		} else {
			__antithesis_instrumentation__.Notify(33181)
		}
		__antithesis_instrumentation__.Notify(33166)
		if importCtx.rowLimit > 0 {
			__antithesis_instrumentation__.Notify(33182)
			optionsClause = fmt.Sprintf("%s, row_limit='%d'", optionsClause, importCtx.rowLimit)
		} else {
			__antithesis_instrumentation__.Notify(33183)
		}
		__antithesis_instrumentation__.Notify(33167)
		switch mode {
		case singleTable:
			__antithesis_instrumentation__.Notify(33184)
			importQuery = fmt.Sprintf(`IMPORT TABLE %s FROM MYSQLDUMP '%s'%s`, tableName,
				unescapedUserfileURL, optionsClause)
		case multiTable:
			__antithesis_instrumentation__.Notify(33185)
			importQuery = fmt.Sprintf(`IMPORT MYSQLDUMP '%s'%s`, unescapedUserfileURL,
				optionsClause)
		default:
			__antithesis_instrumentation__.Notify(33186)
		}
	default:
		__antithesis_instrumentation__.Notify(33168)
		return errors.New("unsupported import format")
	}
	__antithesis_instrumentation__.Notify(33137)

	purl, err := pgurl.Parse(conn.GetURL())
	if err != nil {
		__antithesis_instrumentation__.Notify(33187)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33188)
	}
	__antithesis_instrumentation__.Notify(33138)

	if importCLIKnobs.returnQuery {
		__antithesis_instrumentation__.Notify(33189)
		fmt.Print(importQuery + "\n")
		fmt.Print(purl.GetDatabase())
		return nil
	} else {
		__antithesis_instrumentation__.Notify(33190)
	}
	__antithesis_instrumentation__.Notify(33139)

	_, err = ex.ExecContext(ctx, importQuery, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(33191)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33192)
	}
	__antithesis_instrumentation__.Notify(33140)

	importCompletedMesssage()
	return nil
}

var importCmds = []*cobra.Command{
	importDumpTableCmd,
	importDumpFileCmd,
}

var importCmd = &cobra.Command{
	Use:   "import [command]",
	Short: "import a db or table from a local PGDUMP or MYSQLDUMP file",
	Long:  "import a db or table from a local PGDUMP or MYSQLDUMP file",
	RunE:  UsageAndErr,
}

func init() {
	importCmd.AddCommand(importCmds...)
}
