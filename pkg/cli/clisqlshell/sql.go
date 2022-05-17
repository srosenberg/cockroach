package clisqlshell

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clicfg"
	"github.com/cockroachdb/cockroach/pkg/cli/clierror"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/server/pgurl"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/scanner"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlfsm"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
	readline "github.com/knz/go-libedit"
)

const (
	helpMessageFmt = `You are using 'cockroach sql', CockroachDB's lightweight SQL client.
General
  \q, quit, exit    exit the shell (Ctrl+C/Ctrl+D also supported).

Help
  \? or "help"      print this help.
  \h [NAME]         help on syntax of SQL commands.
  \hf [NAME]        help on SQL built-in functions.

Query Buffer
  \p                during a multi-line statement, show the SQL entered so far.
  \r                during a multi-line statement, erase all the SQL entered so far.
  \| CMD            run an external command and run its output as SQL statements.

Connection
  \c, \connect {[DB] [USER] [HOST] [PORT] | [URL]}
                    connect to a server or print the current connection URL.
                    (Omitted values reuse previous parameters. Use '-' to skip a field.)

Input/Output
  \echo [STRING]    write the provided string to standard output.
  \i                execute commands from the specified file.
  \ir               as \i, but relative to the location of the current script.

Informational
  \l                list all databases in the CockroachDB cluster.
  \dt               show the tables of the current schema in the current database.
  \dT               show the user defined types of the current database.
  \du [USER]        list the specified user, or list the users for all databases if no user is specified.
  \d [TABLE]        show details about columns in the specified table, or alias for '\dt' if no table is specified.
  \dd TABLE         show details about constraints on the specified table.

Formatting
  \x [on|off]       toggle records display format.

Operating System
  \! CMD            run an external command and print its results on standard output.

Configuration
  \set [NAME]       set a client-side flag or (without argument) print the current settings.
  \unset NAME       unset a flag.

Statement diagnostics
  \statement-diag list                               list available bundles.
  \statement-diag download <bundle-id> [<filename>]  download bundle.

%s
More documentation about our SQL dialect and the CLI shell is available online:
%s
%s`

	demoCommandsHelp = `
Commands specific to the demo shell (EXPERIMENTAL):
  \demo ls                     list the demo nodes and their connection URLs.
  \demo shutdown <nodeid>      stop a demo node.
  \demo restart <nodeid>       restart a stopped demo node.
  \demo decommission <nodeid>  decommission a node.
  \demo recommission <nodeid>  recommission a node.
  \demo add <locality>         add a node (locality specified as "region=<region>,zone=<zone>").
`

	defaultPromptPattern = "%n@%M/%/%x>"

	debugPromptPattern = "%n@%M>"
)

type cliState struct {
	cliCtx     *clicfg.Context
	sqlConnCtx *clisqlclient.Context
	sqlExecCtx *clisqlexec.Context
	sqlCtx     *Context
	iCtx       *internalContext

	conn clisqlclient.Conn

	ins readline.EditLine

	buf *bufio.Reader

	singleStatement bool

	levels int

	includeDir string

	fullPrompt string

	continuePrompt string

	useContinuePrompt bool

	currentPrompt string

	copyFromState *clisqlclient.CopyFromState

	lastInputLine string

	atEOF bool

	lastKnownTxnStatus string

	forwardLines []string

	partialLines []string

	partialStmtsLen int

	concatLines string

	exitErr error
}

type cliStateEnum int

const (
	cliStart cliStateEnum = iota
	cliStop

	cliRefreshPrompt

	cliStartLine

	cliContinueLine

	cliReadLine

	cliDecidePath

	cliProcessFirstLine

	cliHandleCliCmd

	cliPrepareStatementLine

	cliCheckStatement

	cliRunStatement
)

func (c *cliState) printCliHelp() {
	__antithesis_instrumentation__.Notify(29417)
	demoHelpStr := ""
	if c.sqlCtx.DemoCluster != nil {
		__antithesis_instrumentation__.Notify(29419)
		demoHelpStr = demoCommandsHelp
	} else {
		__antithesis_instrumentation__.Notify(29420)
	}
	__antithesis_instrumentation__.Notify(29418)
	fmt.Fprintf(c.iCtx.stdout, helpMessageFmt,
		demoHelpStr,
		docs.URL("sql-statements.html"),
		docs.URL("use-the-built-in-sql-client.html"),
	)
	fmt.Fprintln(c.iCtx.stdout)
}

const noLineEditor readline.EditLine = -1

func (c *cliState) hasEditor() bool {
	__antithesis_instrumentation__.Notify(29421)
	return c.ins != noLineEditor
}

func (c *cliState) addHistory(line string) {
	__antithesis_instrumentation__.Notify(29422)
	if !c.hasEditor() || func() bool {
		__antithesis_instrumentation__.Notify(29424)
		return len(line) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(29425)
		return
	} else {
		__antithesis_instrumentation__.Notify(29426)
	}
	__antithesis_instrumentation__.Notify(29423)

	if err := c.ins.AddHistory(line); err != nil {
		__antithesis_instrumentation__.Notify(29427)
		fmt.Fprintf(c.iCtx.stderr, "warning: cannot save command-line history: %v\n", err)
		c.ins.SetAutoSaveHistory("", false)
	} else {
		__antithesis_instrumentation__.Notify(29428)
	}
}

func (c *cliState) invalidSyntax(nextState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29429)
	return c.invalidSyntaxf(nextState, `%s. Try \? for help.`, c.lastInputLine)
}

func (c *cliState) inCopy() bool {
	__antithesis_instrumentation__.Notify(29430)
	return c.copyFromState != nil
}

func (c *cliState) resetCopy() {
	__antithesis_instrumentation__.Notify(29431)
	c.copyFromState = nil
}

func (c *cliState) invalidSyntaxf(
	nextState cliStateEnum, format string, args ...interface{},
) cliStateEnum {
	__antithesis_instrumentation__.Notify(29432)
	fmt.Fprint(c.iCtx.stderr, "invalid syntax: ")
	fmt.Fprintf(c.iCtx.stderr, format, args...)
	fmt.Fprintln(c.iCtx.stderr)
	c.exitErr = errInvalidSyntax
	return nextState
}

func (c *cliState) invalidOptionChange(nextState cliStateEnum, opt string) cliStateEnum {
	__antithesis_instrumentation__.Notify(29433)
	c.exitErr = errors.Newf("cannot change option during multi-line editing: %s\n", opt)
	fmt.Fprintln(c.iCtx.stderr, c.exitErr)
	return nextState
}

func (c *cliState) internalServerError(nextState cliStateEnum, err error) cliStateEnum {
	__antithesis_instrumentation__.Notify(29434)
	fmt.Fprintf(c.iCtx.stderr, "internal server error: %v\n", err)
	c.exitErr = err
	return nextState
}

var options = map[string]struct {
	description               string
	isBoolean                 bool
	validDuringMultilineEntry bool
	set                       func(c *cliState, val string) error
	reset                     func(c *cliState) error

	display    func(c *cliState) string
	deprecated bool
}{
	`auto_trace`: {
		description:               "automatically run statement tracing on each executed statement",
		isBoolean:                 false,
		validDuringMultilineEntry: false,
		set: func(c *cliState, val string) error {
			__antithesis_instrumentation__.Notify(29435)
			b, err := clisqlclient.ParseBool(val)
			if err != nil {
				__antithesis_instrumentation__.Notify(29437)
				c.iCtx.autoTrace = "on, " + val
			} else {
				__antithesis_instrumentation__.Notify(29438)
				if b {
					__antithesis_instrumentation__.Notify(29439)
					c.iCtx.autoTrace = "on, on"
				} else {
					__antithesis_instrumentation__.Notify(29440)
					c.iCtx.autoTrace = ""
				}
			}
			__antithesis_instrumentation__.Notify(29436)
			return nil
		},
		reset: func(c *cliState) error {
			__antithesis_instrumentation__.Notify(29441)
			c.iCtx.autoTrace = ""
			return nil
		},
		display: func(c *cliState) string {
			__antithesis_instrumentation__.Notify(29442)
			if c.iCtx.autoTrace == "" {
				__antithesis_instrumentation__.Notify(29444)
				return "off"
			} else {
				__antithesis_instrumentation__.Notify(29445)
			}
			__antithesis_instrumentation__.Notify(29443)
			return c.iCtx.autoTrace
		},
	},
	`border`: {
		description:               "the border style for the display format 'table'",
		isBoolean:                 false,
		validDuringMultilineEntry: true,
		set: func(c *cliState, val string) error {
			__antithesis_instrumentation__.Notify(29446)
			v, err := strconv.Atoi(val)
			if err != nil {
				__antithesis_instrumentation__.Notify(29449)
				return err
			} else {
				__antithesis_instrumentation__.Notify(29450)
			}
			__antithesis_instrumentation__.Notify(29447)
			if v < 0 || func() bool {
				__antithesis_instrumentation__.Notify(29451)
				return v > 3 == true
			}() == true {
				__antithesis_instrumentation__.Notify(29452)
				return errors.New("only values between 0 and 4 are supported")
			} else {
				__antithesis_instrumentation__.Notify(29453)
			}
			__antithesis_instrumentation__.Notify(29448)
			c.sqlExecCtx.TableBorderMode = v
			return nil
		},
		display: func(c *cliState) string {
			__antithesis_instrumentation__.Notify(29454)
			return strconv.Itoa(c.sqlExecCtx.TableBorderMode)
		},
	},
	`display_format`: {
		description:               "the output format for tabular data (table, csv, tsv, html, sql, records, raw)",
		isBoolean:                 false,
		validDuringMultilineEntry: true,
		set: func(c *cliState, val string) error {
			__antithesis_instrumentation__.Notify(29455)
			return c.sqlExecCtx.TableDisplayFormat.Set(val)
		},
		reset: func(c *cliState) error {
			__antithesis_instrumentation__.Notify(29456)
			displayFormat := clisqlexec.TableDisplayTSV
			if c.sqlExecCtx.TerminalOutput {
				__antithesis_instrumentation__.Notify(29458)
				displayFormat = clisqlexec.TableDisplayTable
			} else {
				__antithesis_instrumentation__.Notify(29459)
			}
			__antithesis_instrumentation__.Notify(29457)
			c.sqlExecCtx.TableDisplayFormat = displayFormat
			return nil
		},
		display: func(c *cliState) string {
			__antithesis_instrumentation__.Notify(29460)
			return c.sqlExecCtx.TableDisplayFormat.String()
		},
	},
	`echo`: {
		description:               "show SQL queries before they are sent to the server",
		isBoolean:                 true,
		validDuringMultilineEntry: false,
		set: func(c *cliState, _ string) error {
			__antithesis_instrumentation__.Notify(29461)
			c.sqlConnCtx.Echo = true
			return nil
		},
		reset: func(c *cliState) error {
			__antithesis_instrumentation__.Notify(29462)
			c.sqlConnCtx.Echo = false
			return nil
		},
		display: func(c *cliState) string {
			__antithesis_instrumentation__.Notify(29463)
			return strconv.FormatBool(c.sqlConnCtx.Echo)
		},
	},
	`errexit`: {
		description:               "exit the shell upon a query error",
		isBoolean:                 true,
		validDuringMultilineEntry: true,
		set: func(c *cliState, _ string) error {
			__antithesis_instrumentation__.Notify(29464)
			c.iCtx.errExit = true
			return nil
		},
		reset: func(c *cliState) error {
			__antithesis_instrumentation__.Notify(29465)
			c.iCtx.errExit = false
			return nil
		},
		display: func(c *cliState) string {
			__antithesis_instrumentation__.Notify(29466)
			return strconv.FormatBool(c.iCtx.errExit)
		},
	},
	`check_syntax`: {
		description:               "check the SQL syntax before running a query",
		isBoolean:                 true,
		validDuringMultilineEntry: false,
		set: func(c *cliState, _ string) error {
			__antithesis_instrumentation__.Notify(29467)
			c.iCtx.checkSyntax = true
			return nil
		},
		reset: func(c *cliState) error {
			__antithesis_instrumentation__.Notify(29468)
			c.iCtx.checkSyntax = false
			return nil
		},
		display: func(c *cliState) string {
			__antithesis_instrumentation__.Notify(29469)
			return strconv.FormatBool(c.iCtx.checkSyntax)
		},
	},
	`show_times`: {
		description:               "display the execution time after each query",
		isBoolean:                 true,
		validDuringMultilineEntry: true,
		set: func(c *cliState, _ string) error {
			__antithesis_instrumentation__.Notify(29470)
			c.sqlExecCtx.ShowTimes = true
			return nil
		},
		reset: func(c *cliState) error {
			__antithesis_instrumentation__.Notify(29471)
			c.sqlExecCtx.ShowTimes = false
			return nil
		},
		display: func(c *cliState) string {
			__antithesis_instrumentation__.Notify(29472)
			return strconv.FormatBool(c.sqlExecCtx.ShowTimes)
		},
	},
	`show_server_times`: {
		description:               "display the server execution times for queries (requires show_times to be set)",
		isBoolean:                 true,
		validDuringMultilineEntry: true,
		set: func(c *cliState, _ string) error {
			__antithesis_instrumentation__.Notify(29473)
			c.sqlConnCtx.EnableServerExecutionTimings = true
			return nil
		},
		reset: func(c *cliState) error {
			__antithesis_instrumentation__.Notify(29474)
			c.sqlConnCtx.EnableServerExecutionTimings = false
			return nil
		},
		display: func(c *cliState) string {
			__antithesis_instrumentation__.Notify(29475)
			return strconv.FormatBool(c.sqlConnCtx.EnableServerExecutionTimings)
		},
	},
	`verbose_times`: {
		description:               "display execution times with more precision (requires show_times to be set)",
		isBoolean:                 true,
		validDuringMultilineEntry: true,
		set: func(c *cliState, _ string) error {
			__antithesis_instrumentation__.Notify(29476)
			c.sqlExecCtx.VerboseTimings = true
			return nil
		},
		reset: func(c *cliState) error {
			__antithesis_instrumentation__.Notify(29477)
			c.sqlExecCtx.VerboseTimings = false
			return nil
		},
		display: func(c *cliState) string {
			__antithesis_instrumentation__.Notify(29478)
			return strconv.FormatBool(c.sqlExecCtx.VerboseTimings)
		},
	},
	`smart_prompt`: {
		description:               "deprecated",
		isBoolean:                 true,
		validDuringMultilineEntry: false,
		set:                       func(c *cliState, _ string) error { __antithesis_instrumentation__.Notify(29479); return nil },
		reset:                     func(c *cliState) error { __antithesis_instrumentation__.Notify(29480); return nil },
		display:                   func(c *cliState) string { __antithesis_instrumentation__.Notify(29481); return "false" },
		deprecated:                true,
	},
	`prompt1`: {
		description:               "prompt string to use before each command (the following are expanded: %M full host, %m host, %> port number, %n user, %/ database, %x txn status)",
		isBoolean:                 false,
		validDuringMultilineEntry: true,
		set: func(c *cliState, val string) error {
			__antithesis_instrumentation__.Notify(29482)
			c.iCtx.customPromptPattern = val
			return nil
		},
		reset: func(c *cliState) error {
			__antithesis_instrumentation__.Notify(29483)
			c.iCtx.customPromptPattern = defaultPromptPattern
			return nil
		},
		display: func(c *cliState) string {
			__antithesis_instrumentation__.Notify(29484)
			return c.iCtx.customPromptPattern
		},
	},
}

var optionNames = func() []string {
	__antithesis_instrumentation__.Notify(29485)
	names := make([]string, 0, len(options))
	for k := range options {
		__antithesis_instrumentation__.Notify(29487)
		names = append(names, k)
	}
	__antithesis_instrumentation__.Notify(29486)
	sort.Strings(names)
	return names
}()

func (c *cliState) handleSet(args []string, nextState, errState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29488)
	if len(args) == 0 {
		__antithesis_instrumentation__.Notify(29496)
		optData := make([][]string, 0, len(options))
		for _, n := range optionNames {
			__antithesis_instrumentation__.Notify(29499)
			if options[n].deprecated {
				__antithesis_instrumentation__.Notify(29501)
				continue
			} else {
				__antithesis_instrumentation__.Notify(29502)
			}
			__antithesis_instrumentation__.Notify(29500)
			optData = append(optData, []string{n, options[n].display(c), options[n].description})
		}
		__antithesis_instrumentation__.Notify(29497)
		err := c.sqlExecCtx.PrintQueryOutput(c.iCtx.stdout, c.iCtx.stderr,
			[]string{"Option", "Value", "Description"},
			clisqlexec.NewRowSliceIter(optData, "lll"))
		if err != nil {
			__antithesis_instrumentation__.Notify(29503)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(29504)
		}
		__antithesis_instrumentation__.Notify(29498)

		return nextState
	} else {
		__antithesis_instrumentation__.Notify(29505)
	}
	__antithesis_instrumentation__.Notify(29489)

	if len(args) == 1 {
		__antithesis_instrumentation__.Notify(29506)

		args = strings.SplitN(args[0], "=", 2)
	} else {
		__antithesis_instrumentation__.Notify(29507)
	}
	__antithesis_instrumentation__.Notify(29490)

	opt, ok := options[args[0]]
	if !ok {
		__antithesis_instrumentation__.Notify(29508)
		return c.invalidSyntax(errState)
	} else {
		__antithesis_instrumentation__.Notify(29509)
	}
	__antithesis_instrumentation__.Notify(29491)
	if len(c.partialLines) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(29510)
		return !opt.validDuringMultilineEntry == true
	}() == true {
		__antithesis_instrumentation__.Notify(29511)
		return c.invalidOptionChange(errState, args[0])
	} else {
		__antithesis_instrumentation__.Notify(29512)
	}
	__antithesis_instrumentation__.Notify(29492)

	var val string
	switch len(args) {
	case 1:
		__antithesis_instrumentation__.Notify(29513)
		val = "true"
	case 2:
		__antithesis_instrumentation__.Notify(29514)
		val = args[1]
	default:
		__antithesis_instrumentation__.Notify(29515)
		return c.invalidSyntax(errState)
	}
	__antithesis_instrumentation__.Notify(29493)

	var err error
	if !opt.isBoolean {
		__antithesis_instrumentation__.Notify(29516)
		err = opt.set(c, val)
	} else {
		__antithesis_instrumentation__.Notify(29517)
		if b, e := clisqlclient.ParseBool(val); e != nil {
			__antithesis_instrumentation__.Notify(29518)
			return c.invalidSyntax(errState)
		} else {
			__antithesis_instrumentation__.Notify(29519)
			if b {
				__antithesis_instrumentation__.Notify(29520)
				err = opt.set(c, "true")
			} else {
				__antithesis_instrumentation__.Notify(29521)
				err = opt.reset(c)
			}
		}
	}
	__antithesis_instrumentation__.Notify(29494)

	if err != nil {
		__antithesis_instrumentation__.Notify(29522)
		fmt.Fprintf(c.iCtx.stderr, "\\set %s: %v\n", strings.Join(args, " "), err)
		c.exitErr = err
		return errState
	} else {
		__antithesis_instrumentation__.Notify(29523)
	}
	__antithesis_instrumentation__.Notify(29495)

	return nextState
}

func (c *cliState) handleUnset(args []string, nextState, errState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29524)
	if len(args) != 1 {
		__antithesis_instrumentation__.Notify(29529)
		return c.invalidSyntax(errState)
	} else {
		__antithesis_instrumentation__.Notify(29530)
	}
	__antithesis_instrumentation__.Notify(29525)
	opt, ok := options[args[0]]
	if !ok {
		__antithesis_instrumentation__.Notify(29531)
		return c.invalidSyntax(errState)
	} else {
		__antithesis_instrumentation__.Notify(29532)
	}
	__antithesis_instrumentation__.Notify(29526)
	if len(c.partialLines) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(29533)
		return !opt.validDuringMultilineEntry == true
	}() == true {
		__antithesis_instrumentation__.Notify(29534)
		return c.invalidOptionChange(errState, args[0])
	} else {
		__antithesis_instrumentation__.Notify(29535)
	}
	__antithesis_instrumentation__.Notify(29527)
	if err := opt.reset(c); err != nil {
		__antithesis_instrumentation__.Notify(29536)
		fmt.Fprintf(c.iCtx.stderr, "\\unset %s: %v\n", args[0], err)
		c.exitErr = err
		return errState
	} else {
		__antithesis_instrumentation__.Notify(29537)
	}
	__antithesis_instrumentation__.Notify(29528)
	return nextState
}

func isEndOfStatement(lastTok int) bool {
	__antithesis_instrumentation__.Notify(29538)
	return lastTok == ';' || func() bool {
		__antithesis_instrumentation__.Notify(29539)
		return lastTok == lexbase.HELPTOKEN == true
	}() == true
}

func (c *cliState) handleDemo(cmd []string, nextState, errState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29540)

	if c.sqlCtx.DemoCluster == nil {
		__antithesis_instrumentation__.Notify(29545)
		return c.invalidSyntaxf(errState, `\demo can only be run with cockroach demo`)
	} else {
		__antithesis_instrumentation__.Notify(29546)
	}
	__antithesis_instrumentation__.Notify(29541)

	if len(cmd) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(29547)
		return cmd[0] == "ls" == true
	}() == true {
		__antithesis_instrumentation__.Notify(29548)
		c.sqlCtx.DemoCluster.ListDemoNodes(c.iCtx.stdout, c.iCtx.stderr, false)
		return nextState
	} else {
		__antithesis_instrumentation__.Notify(29549)
	}
	__antithesis_instrumentation__.Notify(29542)

	if len(cmd) != 2 {
		__antithesis_instrumentation__.Notify(29550)
		return c.invalidSyntax(errState)
	} else {
		__antithesis_instrumentation__.Notify(29551)
	}
	__antithesis_instrumentation__.Notify(29543)

	if cmd[0] == "add" {
		__antithesis_instrumentation__.Notify(29552)
		return c.handleDemoAddNode(cmd, nextState, errState)
	} else {
		__antithesis_instrumentation__.Notify(29553)
	}
	__antithesis_instrumentation__.Notify(29544)

	return c.handleDemoNodeCommands(cmd, nextState, errState)
}

func (c *cliState) handleDemoAddNode(cmd []string, nextState, errState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29554)
	if cmd[0] != "add" {
		__antithesis_instrumentation__.Notify(29557)
		return c.internalServerError(errState, fmt.Errorf("bad call to handleDemoAddNode"))
	} else {
		__antithesis_instrumentation__.Notify(29558)
	}
	__antithesis_instrumentation__.Notify(29555)

	addedNodeID, err := c.sqlCtx.DemoCluster.AddNode(context.Background(), cmd[1])
	if err != nil {
		__antithesis_instrumentation__.Notify(29559)
		return c.internalServerError(errState, err)
	} else {
		__antithesis_instrumentation__.Notify(29560)
	}
	__antithesis_instrumentation__.Notify(29556)
	fmt.Fprintf(c.iCtx.stdout, "node %v has been added with locality \"%s\"\n",
		addedNodeID, c.sqlCtx.DemoCluster.GetLocality(addedNodeID))
	return nextState
}

func (c *cliState) handleDemoNodeCommands(
	cmd []string, nextState, errState cliStateEnum,
) cliStateEnum {
	__antithesis_instrumentation__.Notify(29561)
	nodeID, err := strconv.ParseInt(cmd[1], 10, 32)
	if err != nil {
		__antithesis_instrumentation__.Notify(29564)
		return c.invalidSyntaxf(
			errState,
			"%s",
			errors.Wrapf(err, "%q is not a valid node ID", cmd[1]),
		)
	} else {
		__antithesis_instrumentation__.Notify(29565)
	}
	__antithesis_instrumentation__.Notify(29562)

	ctx := context.Background()

	switch cmd[0] {
	case "shutdown":
		__antithesis_instrumentation__.Notify(29566)
		if err := c.sqlCtx.DemoCluster.DrainAndShutdown(ctx, int32(nodeID)); err != nil {
			__antithesis_instrumentation__.Notify(29575)
			return c.internalServerError(errState, err)
		} else {
			__antithesis_instrumentation__.Notify(29576)
		}
		__antithesis_instrumentation__.Notify(29567)
		fmt.Fprintf(c.iCtx.stdout, "node %d has been shutdown\n", nodeID)
		return nextState
	case "restart":
		__antithesis_instrumentation__.Notify(29568)
		if err := c.sqlCtx.DemoCluster.RestartNode(ctx, int32(nodeID)); err != nil {
			__antithesis_instrumentation__.Notify(29577)
			return c.internalServerError(errState, err)
		} else {
			__antithesis_instrumentation__.Notify(29578)
		}
		__antithesis_instrumentation__.Notify(29569)
		fmt.Fprintf(c.iCtx.stdout, "node %d has been restarted\n", nodeID)
		return nextState
	case "recommission":
		__antithesis_instrumentation__.Notify(29570)
		if err := c.sqlCtx.DemoCluster.Recommission(ctx, int32(nodeID)); err != nil {
			__antithesis_instrumentation__.Notify(29579)
			return c.internalServerError(errState, err)
		} else {
			__antithesis_instrumentation__.Notify(29580)
		}
		__antithesis_instrumentation__.Notify(29571)
		fmt.Fprintf(c.iCtx.stdout, "node %d has been recommissioned\n", nodeID)
		return nextState
	case "decommission":
		__antithesis_instrumentation__.Notify(29572)
		if err := c.sqlCtx.DemoCluster.Decommission(ctx, int32(nodeID)); err != nil {
			__antithesis_instrumentation__.Notify(29581)
			return c.internalServerError(errState, err)
		} else {
			__antithesis_instrumentation__.Notify(29582)
		}
		__antithesis_instrumentation__.Notify(29573)
		fmt.Fprintf(c.iCtx.stdout, "node %d has been decommissioned\n", nodeID)
		return nextState
	default:
		__antithesis_instrumentation__.Notify(29574)
	}
	__antithesis_instrumentation__.Notify(29563)
	return c.invalidSyntax(errState)
}

func (c *cliState) handleHelp(cmd []string, nextState, errState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29583)
	command := strings.TrimSpace(strings.Join(cmd, " "))
	helpText, _ := c.serverSideParse(command + " ??")
	if helpText != "" {
		__antithesis_instrumentation__.Notify(29585)
		fmt.Fprintln(c.iCtx.stdout, helpText)
	} else {
		__antithesis_instrumentation__.Notify(29586)
		fmt.Fprintf(c.iCtx.stderr,
			"no help available for %q.\nTry \\h with no argument to see available help.\n", command)
		c.exitErr = errors.New("no help available")
		return errState
	}
	__antithesis_instrumentation__.Notify(29584)
	return nextState
}

func (c *cliState) handleFunctionHelp(cmd []string, nextState, errState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29587)
	funcName := strings.TrimSpace(strings.Join(cmd, " "))
	helpText, _ := c.serverSideParse(fmt.Sprintf("select %s(??", funcName))
	if helpText != "" {
		__antithesis_instrumentation__.Notify(29589)
		fmt.Fprintln(c.iCtx.stdout, helpText)
	} else {
		__antithesis_instrumentation__.Notify(29590)
		fmt.Fprintf(c.iCtx.stderr,
			"no help available for %q.\nTry \\hf with no argument to see available help.\n", funcName)
		c.exitErr = errors.New("no help available")
		return errState
	}
	__antithesis_instrumentation__.Notify(29588)
	return nextState
}

func (c *cliState) execSyscmd(command string) (string, error) {
	__antithesis_instrumentation__.Notify(29591)
	var cmd *exec.Cmd

	shell := envutil.GetShellCommand(command)
	cmd = exec.Command(shell[0], shell[1:]...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = c.iCtx.stderr

	if err := cmd.Run(); err != nil {
		__antithesis_instrumentation__.Notify(29593)
		return "", errors.Wrap(err, "error in external command")
	} else {
		__antithesis_instrumentation__.Notify(29594)
	}
	__antithesis_instrumentation__.Notify(29592)

	return out.String(), nil
}

var errInvalidSyntax = errors.New("invalid syntax")

func (c *cliState) runSyscmd(line string, nextState, errState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29595)
	command := strings.Trim(line[2:], " \r\n\t\f")
	if command == "" {
		__antithesis_instrumentation__.Notify(29598)
		fmt.Fprintf(c.iCtx.stderr, "Usage:\n  \\! [command]\n")
		c.exitErr = errInvalidSyntax
		return errState
	} else {
		__antithesis_instrumentation__.Notify(29599)
	}
	__antithesis_instrumentation__.Notify(29596)

	cmdOut, err := c.execSyscmd(command)
	if err != nil {
		__antithesis_instrumentation__.Notify(29600)
		fmt.Fprintf(c.iCtx.stderr, "command failed: %s\n", err)
		c.exitErr = err
		return errState
	} else {
		__antithesis_instrumentation__.Notify(29601)
	}
	__antithesis_instrumentation__.Notify(29597)

	fmt.Fprint(c.iCtx.stdout, cmdOut)
	return nextState
}

func (c *cliState) pipeSyscmd(line string, nextState, errState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29602)
	command := strings.Trim(line[2:], " \n\r\t\f")
	if command == "" {
		__antithesis_instrumentation__.Notify(29605)
		fmt.Fprintf(c.iCtx.stderr, "Usage:\n  \\| [command]\n")
		c.exitErr = errInvalidSyntax
		return errState
	} else {
		__antithesis_instrumentation__.Notify(29606)
	}
	__antithesis_instrumentation__.Notify(29603)

	cmdOut, err := c.execSyscmd(command)
	if err != nil {
		__antithesis_instrumentation__.Notify(29607)
		fmt.Fprintf(c.iCtx.stderr, "command failed: %s\n", err)
		c.exitErr = err
		return errState
	} else {
		__antithesis_instrumentation__.Notify(29608)
	}
	__antithesis_instrumentation__.Notify(29604)

	c.lastInputLine = cmdOut
	return nextState
}

var rePromptFmt = regexp.MustCompile("(%.)")

var rePromptDbState = regexp.MustCompile("(?:^|[^%])%[/x]")

const unknownDbName = "?"

const unknownTxnStatus = " ?"

func (c *cliState) doRefreshPrompts(nextState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29609)
	if !c.hasEditor() {
		__antithesis_instrumentation__.Notify(29613)
		return nextState
	} else {
		__antithesis_instrumentation__.Notify(29614)
	}
	__antithesis_instrumentation__.Notify(29610)

	if c.useContinuePrompt {
		__antithesis_instrumentation__.Notify(29615)
		if c.inCopy() {
			__antithesis_instrumentation__.Notify(29617)
			c.continuePrompt = ">> "
		} else {
			__antithesis_instrumentation__.Notify(29618)
			if len(c.fullPrompt) < 3 {
				__antithesis_instrumentation__.Notify(29619)
				c.continuePrompt = "> "
			} else {
				__antithesis_instrumentation__.Notify(29620)

				c.continuePrompt = strings.Repeat(" ", len(c.fullPrompt)-3) + "-> "
			}
		}
		__antithesis_instrumentation__.Notify(29616)

		c.ins.SetLeftPrompt(c.continuePrompt)
		return nextState
	} else {
		__antithesis_instrumentation__.Notify(29621)
	}
	__antithesis_instrumentation__.Notify(29611)

	if c.inCopy() {
		__antithesis_instrumentation__.Notify(29622)
		c.fullPrompt = ">>"
	} else {
		__antithesis_instrumentation__.Notify(29623)

		parsedURL, err := url.Parse(c.conn.GetURL())
		if err != nil {
			__antithesis_instrumentation__.Notify(29628)

			c.fullPrompt = c.conn.GetURL() + "> "
			c.continuePrompt = strings.Repeat(" ", len(c.fullPrompt)-3) + "-> "
			return nextState
		} else {
			__antithesis_instrumentation__.Notify(29629)
		}
		__antithesis_instrumentation__.Notify(29624)

		userName := ""
		if parsedURL.User != nil {
			__antithesis_instrumentation__.Notify(29630)
			userName = parsedURL.User.Username()
		} else {
			__antithesis_instrumentation__.Notify(29631)
		}
		__antithesis_instrumentation__.Notify(29625)

		dbName := unknownDbName
		c.lastKnownTxnStatus = unknownTxnStatus

		wantDbStateInPrompt := rePromptDbState.MatchString(c.iCtx.customPromptPattern)
		if wantDbStateInPrompt {
			__antithesis_instrumentation__.Notify(29632)
			c.refreshTransactionStatus()

			dbName = c.refreshDatabaseName()
		} else {
			__antithesis_instrumentation__.Notify(29633)
		}
		__antithesis_instrumentation__.Notify(29626)

		c.fullPrompt = rePromptFmt.ReplaceAllStringFunc(c.iCtx.customPromptPattern, func(m string) string {
			__antithesis_instrumentation__.Notify(29634)
			switch m {
			case "%M":
				__antithesis_instrumentation__.Notify(29635)
				return parsedURL.Host
			case "%m":
				__antithesis_instrumentation__.Notify(29636)
				return parsedURL.Hostname()
			case "%>":
				__antithesis_instrumentation__.Notify(29637)
				return parsedURL.Port()
			case "%n":
				__antithesis_instrumentation__.Notify(29638)
				return userName
			case "%/":
				__antithesis_instrumentation__.Notify(29639)
				return dbName
			case "%x":
				__antithesis_instrumentation__.Notify(29640)
				return c.lastKnownTxnStatus
			case "%%":
				__antithesis_instrumentation__.Notify(29641)
				return "%"
			default:
				__antithesis_instrumentation__.Notify(29642)
				err = fmt.Errorf("unrecognized format code in prompt: %q", m)
				return ""
			}

		})
		__antithesis_instrumentation__.Notify(29627)
		if err != nil {
			__antithesis_instrumentation__.Notify(29643)
			c.fullPrompt = err.Error()
		} else {
			__antithesis_instrumentation__.Notify(29644)
		}
	}
	__antithesis_instrumentation__.Notify(29612)
	c.fullPrompt += " "
	c.currentPrompt = c.fullPrompt

	c.ins.SetLeftPrompt(c.currentPrompt)

	return nextState
}

func (c *cliState) refreshTransactionStatus() {
	__antithesis_instrumentation__.Notify(29645)
	c.lastKnownTxnStatus = unknownTxnStatus

	dbVal, dbColType, hasVal := c.conn.GetServerValue(
		context.Background(),
		"transaction status", `SHOW TRANSACTION STATUS`)
	if !hasVal {
		__antithesis_instrumentation__.Notify(29647)
		return
	} else {
		__antithesis_instrumentation__.Notify(29648)
	}
	__antithesis_instrumentation__.Notify(29646)

	txnString := clisqlexec.FormatVal(dbVal, dbColType,
		false, false)

	switch txnString {
	case sqlfsm.NoTxnStateStr:
		__antithesis_instrumentation__.Notify(29649)
		c.lastKnownTxnStatus = ""
	case sqlfsm.AbortedStateStr:
		__antithesis_instrumentation__.Notify(29650)
		c.lastKnownTxnStatus = " ERROR"
	case sqlfsm.CommitWaitStateStr:
		__antithesis_instrumentation__.Notify(29651)
		c.lastKnownTxnStatus = "  DONE"
	case sqlfsm.OpenStateStr:
		__antithesis_instrumentation__.Notify(29652)

		c.lastKnownTxnStatus = "  OPEN"
	default:
		__antithesis_instrumentation__.Notify(29653)
	}
}

func (c *cliState) refreshDatabaseName() string {
	__antithesis_instrumentation__.Notify(29654)
	if !(c.lastKnownTxnStatus == "" || func() bool {
		__antithesis_instrumentation__.Notify(29658)
		return c.lastKnownTxnStatus == "  OPEN" == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(29659)
		return c.lastKnownTxnStatus == unknownTxnStatus == true
	}() == true) {
		__antithesis_instrumentation__.Notify(29660)
		return unknownDbName
	} else {
		__antithesis_instrumentation__.Notify(29661)
	}
	__antithesis_instrumentation__.Notify(29655)

	dbVal, dbColType, hasVal := c.conn.GetServerValue(
		context.Background(),
		"database name", `SHOW DATABASE`)
	if !hasVal {
		__antithesis_instrumentation__.Notify(29662)
		return unknownDbName
	} else {
		__antithesis_instrumentation__.Notify(29663)
	}
	__antithesis_instrumentation__.Notify(29656)

	if dbVal == "" {
		__antithesis_instrumentation__.Notify(29664)

		fmt.Fprintln(c.iCtx.stderr, "warning: no current database set."+
			" Use SET database = <dbname> to change, CREATE DATABASE to make a new database.")
	} else {
		__antithesis_instrumentation__.Notify(29665)
	}
	__antithesis_instrumentation__.Notify(29657)

	dbName := clisqlexec.FormatVal(dbVal, dbColType,
		false, false)

	c.conn.SetCurrentDatabase(dbName)
	c.iCtx.dbName = dbName

	return dbName
}

var cmdHistFile = envutil.EnvOrDefaultString("COCKROACH_SQL_CLI_HISTORY", ".cockroachsql_history")

func (c *cliState) GetCompletions(s string) []string {
	__antithesis_instrumentation__.Notify(29666)
	sql, _ := c.ins.GetLineInfo()

	if c.inCopy() {
		__antithesis_instrumentation__.Notify(29670)
		return []string{s + "\t"}
	} else {
		__antithesis_instrumentation__.Notify(29671)
	}
	__antithesis_instrumentation__.Notify(29667)

	if !strings.HasSuffix(sql, "??") {
		__antithesis_instrumentation__.Notify(29672)
		query := fmt.Sprintf(`SHOW COMPLETIONS AT OFFSET %d FOR %s`, len(sql), lexbase.EscapeSQLString(sql))
		var rows [][]string
		var err error
		err = c.runWithInterruptableCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(29676)
			_, rows, err = c.sqlExecCtx.RunQuery(ctx, c.conn,
				clisqlclient.MakeQuery(query), true)
			return err
		})
		__antithesis_instrumentation__.Notify(29673)

		if err != nil {
			__antithesis_instrumentation__.Notify(29677)
			clierror.OutputError(c.iCtx.stdout, err, true, false)
		} else {
			__antithesis_instrumentation__.Notify(29678)
		}
		__antithesis_instrumentation__.Notify(29674)

		var completions []string
		for _, row := range rows {
			__antithesis_instrumentation__.Notify(29679)
			completions = append(completions, row[0])
		}
		__antithesis_instrumentation__.Notify(29675)

		return completions
	} else {
		__antithesis_instrumentation__.Notify(29680)
	}
	__antithesis_instrumentation__.Notify(29668)

	helpText, err := c.serverSideParse(sql)
	if helpText != "" {
		__antithesis_instrumentation__.Notify(29681)

		fmt.Fprintf(c.iCtx.stdout, "\nSuggestion:\n%s\n", helpText)
	} else {
		__antithesis_instrumentation__.Notify(29682)
		if err != nil {
			__antithesis_instrumentation__.Notify(29683)

			fmt.Fprintln(c.iCtx.stdout)
			clierror.OutputError(c.iCtx.stdout, err, true, false)
		} else {
			__antithesis_instrumentation__.Notify(29684)
		}
	}
	__antithesis_instrumentation__.Notify(29669)

	fmt.Fprint(c.iCtx.stdout, c.currentPrompt, sql)
	return nil
}

func (c *cliState) doStart(nextState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29685)

	c.partialLines = []string{}

	if c.cliCtx.IsInteractive {
		__antithesis_instrumentation__.Notify(29687)
		fmt.Fprintln(c.iCtx.stdout, "#\n# Enter \\? for a brief introduction.\n#")
	} else {
		__antithesis_instrumentation__.Notify(29688)
	}
	__antithesis_instrumentation__.Notify(29686)

	return nextState
}

func (c *cliState) doStartLine(nextState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29689)

	c.atEOF = false
	c.partialLines = c.partialLines[:0]
	c.partialStmtsLen = 0

	c.useContinuePrompt = false

	return nextState
}

func (c *cliState) doContinueLine(nextState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29690)
	c.atEOF = false

	c.useContinuePrompt = true

	return nextState
}

func (c *cliState) doReadLine(nextState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29691)
	if len(c.forwardLines) > 0 {
		__antithesis_instrumentation__.Notify(29695)

		c.lastInputLine = c.forwardLines[0]
		c.forwardLines = c.forwardLines[1:]
		return nextState
	} else {
		__antithesis_instrumentation__.Notify(29696)
	}
	__antithesis_instrumentation__.Notify(29692)

	var l string
	var err error
	if c.buf == nil {
		__antithesis_instrumentation__.Notify(29697)
		l, err = c.ins.GetLine()
		if len(l) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(29698)
			return l[len(l)-1] == '\n' == true
		}() == true {
			__antithesis_instrumentation__.Notify(29699)

			l = l[:len(l)-1]
		} else {
			__antithesis_instrumentation__.Notify(29700)

			fmt.Fprintln(c.iCtx.stdout)
		}
	} else {
		__antithesis_instrumentation__.Notify(29701)
		l, err = c.buf.ReadString('\n')

		if err == io.EOF && func() bool {
			__antithesis_instrumentation__.Notify(29702)
			return len(l) != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(29703)
			err = nil
		} else {
			__antithesis_instrumentation__.Notify(29704)
			if err == nil {
				__antithesis_instrumentation__.Notify(29705)

				l = l[:len(l)-1]
			} else {
				__antithesis_instrumentation__.Notify(29706)
			}
		}
	}
	__antithesis_instrumentation__.Notify(29693)

	switch {
	case err == nil:
		__antithesis_instrumentation__.Notify(29707)

		lines := strings.Split(l, "\n")
		if len(lines) > 1 {
			__antithesis_instrumentation__.Notify(29718)

			l = lines[0]
			c.forwardLines = lines[1:]
		} else {
			__antithesis_instrumentation__.Notify(29719)
		}

	case errors.Is(err, readline.ErrInterrupted):
		__antithesis_instrumentation__.Notify(29708)
		if !c.cliCtx.IsInteractive {
			__antithesis_instrumentation__.Notify(29720)

			c.exitErr = err
			return cliStop
		} else {
			__antithesis_instrumentation__.Notify(29721)
		}
		__antithesis_instrumentation__.Notify(29709)

		if c.inCopy() {
			__antithesis_instrumentation__.Notify(29722)

			defer func() {
				__antithesis_instrumentation__.Notify(29725)
				c.resetCopy()
				c.partialLines = c.partialLines[:0]
				c.partialStmtsLen = 0
				c.useContinuePrompt = false
			}()
			__antithesis_instrumentation__.Notify(29723)
			c.exitErr = errors.CombineErrors(
				pgerror.Newf(pgcode.QueryCanceled, "COPY canceled by user"),
				c.copyFromState.Cancel(),
			)
			if c.exitErr != nil {
				__antithesis_instrumentation__.Notify(29726)
				if !c.singleStatement {
					__antithesis_instrumentation__.Notify(29727)
					clierror.OutputError(c.iCtx.stderr, c.exitErr, true, false)
				} else {
					__antithesis_instrumentation__.Notify(29728)
				}
			} else {
				__antithesis_instrumentation__.Notify(29729)
			}
			__antithesis_instrumentation__.Notify(29724)
			return cliRefreshPrompt
		} else {
			__antithesis_instrumentation__.Notify(29730)
		}
		__antithesis_instrumentation__.Notify(29710)

		if l != "" {
			__antithesis_instrumentation__.Notify(29731)

			return cliReadLine
		} else {
			__antithesis_instrumentation__.Notify(29732)
		}
		__antithesis_instrumentation__.Notify(29711)

		if len(c.partialLines) > 0 {
			__antithesis_instrumentation__.Notify(29733)

			return cliStartLine
		} else {
			__antithesis_instrumentation__.Notify(29734)
		}
		__antithesis_instrumentation__.Notify(29712)

		if c.sqlExecCtx.TerminalOutput {
			__antithesis_instrumentation__.Notify(29735)
			fmt.Fprintf(c.iCtx.stdout, "^C\nUse \\q or terminate input to exit.\n")
		} else {
			__antithesis_instrumentation__.Notify(29736)
		}
		__antithesis_instrumentation__.Notify(29713)
		return cliStartLine

	case errors.Is(err, io.EOF):
		__antithesis_instrumentation__.Notify(29714)

		if c.inCopy() && func() bool {
			__antithesis_instrumentation__.Notify(29737)
			return c.cliCtx.IsInteractive == true
		}() == true {
			__antithesis_instrumentation__.Notify(29738)
			return cliRunStatement
		} else {
			__antithesis_instrumentation__.Notify(29739)
		}
		__antithesis_instrumentation__.Notify(29715)

		c.atEOF = true

		if c.cliCtx.IsInteractive {
			__antithesis_instrumentation__.Notify(29740)

			return cliStop
		} else {
			__antithesis_instrumentation__.Notify(29741)
		}
		__antithesis_instrumentation__.Notify(29716)

		if len(c.partialLines) == 0 {
			__antithesis_instrumentation__.Notify(29742)
			return cliStop
		} else {
			__antithesis_instrumentation__.Notify(29743)
		}

	default:
		__antithesis_instrumentation__.Notify(29717)

		fmt.Fprintf(c.iCtx.stderr, "input error: %s\n", err)
		c.exitErr = err
		return cliStop
	}
	__antithesis_instrumentation__.Notify(29694)

	c.lastInputLine = l
	return nextState
}

func (c *cliState) doProcessFirstLine(startState, nextState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29744)

	switch c.lastInputLine {
	case "":
		__antithesis_instrumentation__.Notify(29746)

		return startState

	case "help":
		__antithesis_instrumentation__.Notify(29747)
		c.printCliHelp()
		return startState

	case "exit", "quit":
		__antithesis_instrumentation__.Notify(29748)
		return cliStop
	default:
		__antithesis_instrumentation__.Notify(29749)
	}
	__antithesis_instrumentation__.Notify(29745)

	return nextState
}

func (c *cliState) doHandleCliCmd(loopState, nextState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29750)
	if len(c.lastInputLine) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(29754)
		return c.lastInputLine[0] != '\\' == true
	}() == true {
		__antithesis_instrumentation__.Notify(29755)
		return nextState
	} else {
		__antithesis_instrumentation__.Notify(29756)
	}
	__antithesis_instrumentation__.Notify(29751)

	errState := loopState
	if c.iCtx.errExit {
		__antithesis_instrumentation__.Notify(29757)

		errState = cliStop
	} else {
		__antithesis_instrumentation__.Notify(29758)
	}
	__antithesis_instrumentation__.Notify(29752)

	c.addHistory(c.lastInputLine)

	line := strings.TrimRight(c.lastInputLine, "; ")

	cmd := strings.Fields(line)
	switch cmd[0] {
	case `\q`, `\quit`, `\exit`:
		__antithesis_instrumentation__.Notify(29759)
		return cliStop

	case `\`, `\?`, `\help`:
		__antithesis_instrumentation__.Notify(29760)
		c.printCliHelp()

	case `\echo`:
		__antithesis_instrumentation__.Notify(29761)
		fmt.Fprintln(c.iCtx.stdout, strings.Join(cmd[1:], " "))

	case `\set`:
		__antithesis_instrumentation__.Notify(29762)
		return c.handleSet(cmd[1:], loopState, errState)

	case `\unset`:
		__antithesis_instrumentation__.Notify(29763)
		return c.handleUnset(cmd[1:], loopState, errState)

	case `\!`:
		__antithesis_instrumentation__.Notify(29764)
		return c.runSyscmd(c.lastInputLine, loopState, errState)

	case `\i`:
		__antithesis_instrumentation__.Notify(29765)
		return c.runInclude(cmd[1:], loopState, errState, false)

	case `\ir`:
		__antithesis_instrumentation__.Notify(29766)
		return c.runInclude(cmd[1:], loopState, errState, true)

	case `\p`:
		__antithesis_instrumentation__.Notify(29767)

		fmt.Fprintln(c.iCtx.stdout, strings.Join(c.partialLines, "\n"))

	case `\r`:
		__antithesis_instrumentation__.Notify(29768)

		return cliStartLine

	case `\show`:
		__antithesis_instrumentation__.Notify(29769)
		fmt.Fprintln(c.iCtx.stderr, `warning: \show is deprecated. Use \p.`)
		if len(c.partialLines) == 0 {
			__antithesis_instrumentation__.Notify(29796)
			fmt.Fprintf(c.iCtx.stderr, "No input so far. Did you mean SHOW?\n")
		} else {
			__antithesis_instrumentation__.Notify(29797)
			for _, s := range c.partialLines {
				__antithesis_instrumentation__.Notify(29798)
				fmt.Fprintln(c.iCtx.stdout, s)
			}
		}

	case `\|`:
		__antithesis_instrumentation__.Notify(29770)
		return c.pipeSyscmd(c.lastInputLine, nextState, errState)

	case `\h`:
		__antithesis_instrumentation__.Notify(29771)
		return c.handleHelp(cmd[1:], loopState, errState)

	case `\hf`:
		__antithesis_instrumentation__.Notify(29772)
		if len(cmd) == 1 {
			__antithesis_instrumentation__.Notify(29799)
			c.concatLines = `SELECT DISTINCT proname AS function FROM pg_proc ORDER BY 1`
			return cliRunStatement
		} else {
			__antithesis_instrumentation__.Notify(29800)
		}
		__antithesis_instrumentation__.Notify(29773)
		return c.handleFunctionHelp(cmd[1:], loopState, errState)

	case `\l`:
		__antithesis_instrumentation__.Notify(29774)
		c.concatLines = `SHOW DATABASES`
		return cliRunStatement

	case `\dt`:
		__antithesis_instrumentation__.Notify(29775)
		c.concatLines = `SHOW TABLES`
		return cliRunStatement

	case `\copy`:
		__antithesis_instrumentation__.Notify(29776)
		c.exitErr = c.runWithInterruptableCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(29801)
			return c.beginCopyFrom(ctx, c.concatLines)
		})
		__antithesis_instrumentation__.Notify(29777)
		if !c.singleStatement {
			__antithesis_instrumentation__.Notify(29802)
			clierror.OutputError(c.iCtx.stderr, c.exitErr, true, false)
		} else {
			__antithesis_instrumentation__.Notify(29803)
		}
		__antithesis_instrumentation__.Notify(29778)
		if c.exitErr != nil && func() bool {
			__antithesis_instrumentation__.Notify(29804)
			return c.iCtx.errExit == true
		}() == true {
			__antithesis_instrumentation__.Notify(29805)
			return cliStop
		} else {
			__antithesis_instrumentation__.Notify(29806)
		}
		__antithesis_instrumentation__.Notify(29779)
		return cliStartLine

	case `\.`:
		__antithesis_instrumentation__.Notify(29780)
		if c.inCopy() {
			__antithesis_instrumentation__.Notify(29807)
			c.concatLines += "\n" + `\.`
			return cliRunStatement
		} else {
			__antithesis_instrumentation__.Notify(29808)
		}
		__antithesis_instrumentation__.Notify(29781)
		return c.invalidSyntax(errState)

	case `\dT`:
		__antithesis_instrumentation__.Notify(29782)
		c.concatLines = `SHOW TYPES`
		return cliRunStatement

	case `\du`:
		__antithesis_instrumentation__.Notify(29783)
		if len(cmd) == 1 {
			__antithesis_instrumentation__.Notify(29809)
			c.concatLines = `SHOW USERS`
			return cliRunStatement
		} else {
			__antithesis_instrumentation__.Notify(29810)
			if len(cmd) == 2 {
				__antithesis_instrumentation__.Notify(29811)
				c.concatLines = fmt.Sprintf(`SELECT * FROM [SHOW USERS] WHERE username = %s`, lexbase.EscapeSQLString(cmd[1]))
				return cliRunStatement
			} else {
				__antithesis_instrumentation__.Notify(29812)
			}
		}
		__antithesis_instrumentation__.Notify(29784)
		return c.invalidSyntax(errState)

	case `\d`:
		__antithesis_instrumentation__.Notify(29785)
		if len(cmd) == 1 {
			__antithesis_instrumentation__.Notify(29813)
			c.concatLines = `SHOW TABLES`
			return cliRunStatement
		} else {
			__antithesis_instrumentation__.Notify(29814)
			if len(cmd) == 2 {
				__antithesis_instrumentation__.Notify(29815)
				c.concatLines = `SHOW COLUMNS FROM ` + cmd[1]
				return cliRunStatement
			} else {
				__antithesis_instrumentation__.Notify(29816)
			}
		}
		__antithesis_instrumentation__.Notify(29786)
		return c.invalidSyntax(errState)
	case `\dd`:
		__antithesis_instrumentation__.Notify(29787)
		if len(cmd) == 2 {
			__antithesis_instrumentation__.Notify(29817)
			c.concatLines = `SHOW CONSTRAINTS FROM ` + cmd[1] + ` WITH COMMENT`
			return cliRunStatement
		} else {
			__antithesis_instrumentation__.Notify(29818)
		}
		__antithesis_instrumentation__.Notify(29788)
		return c.invalidSyntax(errState)
	case `\connect`, `\c`:
		__antithesis_instrumentation__.Notify(29789)
		return c.handleConnect(cmd[1:], loopState, errState)

	case `\x`:
		__antithesis_instrumentation__.Notify(29790)
		format := clisqlexec.TableDisplayRecords
		switch len(cmd) {
		case 1:
			__antithesis_instrumentation__.Notify(29819)
			if c.sqlExecCtx.TableDisplayFormat == clisqlexec.TableDisplayRecords {
				__antithesis_instrumentation__.Notify(29822)
				format = clisqlexec.TableDisplayTable
			} else {
				__antithesis_instrumentation__.Notify(29823)
			}
		case 2:
			__antithesis_instrumentation__.Notify(29820)
			b, err := clisqlclient.ParseBool(cmd[1])
			if err != nil {
				__antithesis_instrumentation__.Notify(29824)
				return c.invalidSyntax(errState)
			} else {
				__antithesis_instrumentation__.Notify(29825)
				if b {
					__antithesis_instrumentation__.Notify(29826)
					format = clisqlexec.TableDisplayRecords
				} else {
					__antithesis_instrumentation__.Notify(29827)
					format = clisqlexec.TableDisplayTable
				}
			}
		default:
			__antithesis_instrumentation__.Notify(29821)
			return c.invalidSyntax(errState)
		}
		__antithesis_instrumentation__.Notify(29791)
		c.sqlExecCtx.TableDisplayFormat = format
		return loopState

	case `\demo`:
		__antithesis_instrumentation__.Notify(29792)
		return c.handleDemo(cmd[1:], loopState, errState)

	case `\statement-diag`:
		__antithesis_instrumentation__.Notify(29793)
		return c.handleStatementDiag(cmd[1:], loopState, errState)

	default:
		__antithesis_instrumentation__.Notify(29794)
		if strings.HasPrefix(cmd[0], `\d`) {
			__antithesis_instrumentation__.Notify(29828)

			fmt.Fprint(c.iCtx.stderr, "Suggestion: use the SQL SHOW statement to inspect your schema.\n")
		} else {
			__antithesis_instrumentation__.Notify(29829)
		}
		__antithesis_instrumentation__.Notify(29795)
		return c.invalidSyntax(errState)
	}
	__antithesis_instrumentation__.Notify(29753)

	return loopState
}

func (c *cliState) handleConnect(
	cmd []string, loopState, errState cliStateEnum,
) (resState cliStateEnum) {
	__antithesis_instrumentation__.Notify(29830)
	if err := c.handleConnectInternal(cmd); err != nil {
		__antithesis_instrumentation__.Notify(29832)
		fmt.Fprintln(c.iCtx.stderr, err)
		c.exitErr = err
		return errState
	} else {
		__antithesis_instrumentation__.Notify(29833)
	}
	__antithesis_instrumentation__.Notify(29831)
	return loopState
}

func (c *cliState) handleConnectInternal(cmd []string) error {
	__antithesis_instrumentation__.Notify(29834)
	firstArgIsURL := len(cmd) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(29843)
		return (strings.HasPrefix(cmd[0], "postgres://") || func() bool {
			__antithesis_instrumentation__.Notify(29844)
			return strings.HasPrefix(cmd[0], "postgresql://") == true
		}() == true) == true
	}() == true

	if len(cmd) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(29845)
		return firstArgIsURL == true
	}() == true {
		__antithesis_instrumentation__.Notify(29846)

		parseURL := c.sqlCtx.ParseURL
		if parseURL == nil {
			__antithesis_instrumentation__.Notify(29849)
			parseURL = pgurl.Parse
		} else {
			__antithesis_instrumentation__.Notify(29850)
		}
		__antithesis_instrumentation__.Notify(29847)
		purl, err := parseURL(cmd[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(29851)
			return err
		} else {
			__antithesis_instrumentation__.Notify(29852)
		}
		__antithesis_instrumentation__.Notify(29848)
		return c.switchToURL(purl)
	} else {
		__antithesis_instrumentation__.Notify(29853)
	}
	__antithesis_instrumentation__.Notify(29835)

	currURL, err := pgurl.Parse(c.conn.GetURL())
	if err != nil {
		__antithesis_instrumentation__.Notify(29854)
		return errors.Wrap(err, "parsing current connection URL")
	} else {
		__antithesis_instrumentation__.Notify(29855)
	}
	__antithesis_instrumentation__.Notify(29836)

	dbName := c.iCtx.dbName
	if dbName == "" {
		__antithesis_instrumentation__.Notify(29856)
		dbName = currURL.GetDatabase()
	} else {
		__antithesis_instrumentation__.Notify(29857)
	}
	__antithesis_instrumentation__.Notify(29837)

	prevproto, prevhost, prevport := currURL.GetNetworking()
	tlsUsed, mode, caCertPath := currURL.GetTLSOptions()

	newURL := pgurl.New()

	newURL.
		WithDefaultHost(prevhost).
		WithDefaultPort(prevport).
		WithDefaultDatabase(dbName).
		WithDefaultUsername(currURL.GetUsername())
	if tlsUsed {
		__antithesis_instrumentation__.Notify(29858)
		newURL.WithTransport(pgurl.TransportTLS(mode, caCertPath))
	} else {
		__antithesis_instrumentation__.Notify(29859)
		newURL.WithTransport(pgurl.TransportNone())
	}
	__antithesis_instrumentation__.Notify(29838)
	if err := newURL.AddOptions(currURL.GetExtraOptions()); err != nil {
		__antithesis_instrumentation__.Notify(29860)
		return err
	} else {
		__antithesis_instrumentation__.Notify(29861)
	}
	__antithesis_instrumentation__.Notify(29839)

	switch len(cmd) {
	case 4:
		__antithesis_instrumentation__.Notify(29862)
		if cmd[3] != "-" {
			__antithesis_instrumentation__.Notify(29872)
			if err := newURL.SetOption("port", cmd[3]); err != nil {
				__antithesis_instrumentation__.Notify(29873)
				return err
			} else {
				__antithesis_instrumentation__.Notify(29874)
			}
		} else {
			__antithesis_instrumentation__.Notify(29875)
		}
		__antithesis_instrumentation__.Notify(29863)
		fallthrough
	case 3:
		__antithesis_instrumentation__.Notify(29864)
		if cmd[2] != "-" {
			__antithesis_instrumentation__.Notify(29876)
			if err := newURL.SetOption("host", cmd[2]); err != nil {
				__antithesis_instrumentation__.Notify(29877)
				return err
			} else {
				__antithesis_instrumentation__.Notify(29878)
			}
		} else {
			__antithesis_instrumentation__.Notify(29879)
		}
		__antithesis_instrumentation__.Notify(29865)
		fallthrough
	case 2:
		__antithesis_instrumentation__.Notify(29866)
		if cmd[1] != "-" {
			__antithesis_instrumentation__.Notify(29880)
			if err := newURL.SetOption("user", cmd[1]); err != nil {
				__antithesis_instrumentation__.Notify(29881)
				return err
			} else {
				__antithesis_instrumentation__.Notify(29882)
			}
		} else {
			__antithesis_instrumentation__.Notify(29883)
		}
		__antithesis_instrumentation__.Notify(29867)
		fallthrough
	case 1:
		__antithesis_instrumentation__.Notify(29868)
		if cmd[0] != "-" {
			__antithesis_instrumentation__.Notify(29884)
			if err := newURL.SetOption("database", cmd[0]); err != nil {
				__antithesis_instrumentation__.Notify(29885)
				return err
			} else {
				__antithesis_instrumentation__.Notify(29886)
			}
		} else {
			__antithesis_instrumentation__.Notify(29887)
		}
	case 0:
		__antithesis_instrumentation__.Notify(29869)

		dbName := c.iCtx.dbName
		if dbName == "" {
			__antithesis_instrumentation__.Notify(29888)
			dbName = currURL.GetDatabase()
		} else {
			__antithesis_instrumentation__.Notify(29889)
		}
		__antithesis_instrumentation__.Notify(29870)
		fmt.Fprintf(c.iCtx.stdout, "Connection string: %s\n", currURL.ToPQ())
		fmt.Fprintf(c.iCtx.stdout, "You are connected to database %q as user %q.\n", dbName, currURL.GetUsername())
		return nil

	default:
		__antithesis_instrumentation__.Notify(29871)
		return errors.Newf(`unknown syntax: \c %s`, strings.Join(cmd, " "))
	}
	__antithesis_instrumentation__.Notify(29840)

	if proto, host, port := newURL.GetNetworking(); proto == prevproto && func() bool {
		__antithesis_instrumentation__.Notify(29890)
		return host == prevhost == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(29891)
		return port == prevport == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(29892)
		return newURL.GetUsername() == currURL.GetUsername() == true
	}() == true {
		__antithesis_instrumentation__.Notify(29893)

		currURL.ClearPassword()

		prevAuthn, err := currURL.GetAuthnOption()
		if err != nil {
			__antithesis_instrumentation__.Notify(29895)
			return err
		} else {
			__antithesis_instrumentation__.Notify(29896)
		}
		__antithesis_instrumentation__.Notify(29894)
		newURL.WithAuthn(prevAuthn)
	} else {
		__antithesis_instrumentation__.Notify(29897)
	}
	__antithesis_instrumentation__.Notify(29841)

	if err := newURL.Validate(); err != nil {
		__antithesis_instrumentation__.Notify(29898)
		return errors.Wrap(err, "validating the new URL")
	} else {
		__antithesis_instrumentation__.Notify(29899)
	}
	__antithesis_instrumentation__.Notify(29842)

	return c.switchToURL(newURL)
}

func (c *cliState) switchToURL(newURL *pgurl.URL) error {
	__antithesis_instrumentation__.Notify(29900)
	fmt.Fprintln(c.iCtx.stdout, "using new connection URL:", newURL)

	usePw, pwSet, _ := newURL.GetAuthnPassword()

	if err := c.conn.Close(); err != nil {
		__antithesis_instrumentation__.Notify(29902)
		fmt.Fprintf(c.iCtx.stderr, "warning: error while closing connection: %v\n", err)
	} else {
		__antithesis_instrumentation__.Notify(29903)
	}
	__antithesis_instrumentation__.Notify(29901)
	c.conn.SetURL(newURL.ToPQ().String())
	c.conn.SetMissingPassword(!usePw || func() bool {
		__antithesis_instrumentation__.Notify(29904)
		return !pwSet == true
	}() == true)
	return nil
}

const maxRecursionLevels = 10

func (c *cliState) runInclude(
	cmd []string, contState, errState cliStateEnum, relative bool,
) (resState cliStateEnum) {
	__antithesis_instrumentation__.Notify(29905)
	if len(cmd) != 1 {
		__antithesis_instrumentation__.Notify(29912)
		return c.invalidSyntax(errState)
	} else {
		__antithesis_instrumentation__.Notify(29913)
	}
	__antithesis_instrumentation__.Notify(29906)

	if c.levels >= maxRecursionLevels {
		__antithesis_instrumentation__.Notify(29914)
		c.exitErr = errors.Newf(`\i: too many recursion levels (max %d)`, maxRecursionLevels)
		fmt.Fprintf(c.iCtx.stderr, "%v\n", c.exitErr)
		return errState
	} else {
		__antithesis_instrumentation__.Notify(29915)
	}
	__antithesis_instrumentation__.Notify(29907)

	if len(c.partialLines) > 0 {
		__antithesis_instrumentation__.Notify(29916)
		return c.invalidSyntax(errState)
	} else {
		__antithesis_instrumentation__.Notify(29917)
	}
	__antithesis_instrumentation__.Notify(29908)

	filename := cmd[0]
	if !filepath.IsAbs(filename) && func() bool {
		__antithesis_instrumentation__.Notify(29918)
		return relative == true
	}() == true {
		__antithesis_instrumentation__.Notify(29919)

		filename = filepath.Join(c.includeDir, filename)
	} else {
		__antithesis_instrumentation__.Notify(29920)
	}
	__antithesis_instrumentation__.Notify(29909)

	f, err := os.Open(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(29921)
		fmt.Fprintln(c.iCtx.stderr, err)
		c.exitErr = err
		return errState
	} else {
		__antithesis_instrumentation__.Notify(29922)
	}
	__antithesis_instrumentation__.Notify(29910)

	defer func() {
		__antithesis_instrumentation__.Notify(29923)
		if err := f.Close(); err != nil {
			__antithesis_instrumentation__.Notify(29924)
			fmt.Fprintf(c.iCtx.stderr, "error: closing %s: %v\n", filename, err)
			c.exitErr = errors.CombineErrors(c.exitErr, err)
			resState = errState
		} else {
			__antithesis_instrumentation__.Notify(29925)
		}
	}()
	__antithesis_instrumentation__.Notify(29911)

	newLevel := c.levels + 1

	return c.runIncludeInternal(contState, errState,
		filename, bufio.NewReader(f), newLevel, false)
}

func (c *cliState) runString(
	contState, errState cliStateEnum, stmt string,
) (resState cliStateEnum) {
	__antithesis_instrumentation__.Notify(29926)
	r := strings.NewReader(stmt)
	buf := bufio.NewReader(r)

	newLevel := c.levels

	return c.runIncludeInternal(contState, errState,
		"-e", buf, newLevel, true)
}

func (c *cliState) runIncludeInternal(
	contState, errState cliStateEnum,
	filename string,
	input *bufio.Reader,
	level int,
	singleStatement bool,
) (resState cliStateEnum) {
	__antithesis_instrumentation__.Notify(29927)
	newState := cliState{
		cliCtx:     c.cliCtx,
		sqlConnCtx: c.sqlConnCtx,
		sqlExecCtx: c.sqlExecCtx,
		sqlCtx:     c.sqlCtx,
		iCtx:       c.iCtx,

		conn:       c.conn,
		includeDir: filepath.Dir(filename),
		ins:        noLineEditor,
		buf:        input,
		levels:     level,

		singleStatement: singleStatement,
	}

	if err := newState.doRunShell(cliStartLine, nil, nil, nil); err != nil {
		__antithesis_instrumentation__.Notify(29929)

		c.exitErr = errors.Wrapf(err, "%v", filename)
		return errState
	} else {
		__antithesis_instrumentation__.Notify(29930)
	}
	__antithesis_instrumentation__.Notify(29928)

	return contState
}

func (c *cliState) doPrepareStatementLine(
	startState, contState, checkState, execState cliStateEnum,
) cliStateEnum {
	__antithesis_instrumentation__.Notify(29931)
	c.partialLines = append(c.partialLines, c.lastInputLine)

	c.concatLines = strings.Trim(strings.Join(c.partialLines, "\n"), " \r\n\t\f")

	if c.concatLines == "" {
		__antithesis_instrumentation__.Notify(29938)

		return startState
	} else {
		__antithesis_instrumentation__.Notify(29939)
	}
	__antithesis_instrumentation__.Notify(29932)

	lastTok, ok := scanner.LastLexicalToken(c.concatLines)
	if c.partialStmtsLen == 0 && func() bool {
		__antithesis_instrumentation__.Notify(29940)
		return !ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(29941)

		if c.lastInputLine != "" {
			__antithesis_instrumentation__.Notify(29943)
			c.addHistory(c.lastInputLine)
		} else {
			__antithesis_instrumentation__.Notify(29944)
		}
		__antithesis_instrumentation__.Notify(29942)
		return startState
	} else {
		__antithesis_instrumentation__.Notify(29945)
	}
	__antithesis_instrumentation__.Notify(29933)
	endOfStmt := (!c.inCopy() && func() bool {
		__antithesis_instrumentation__.Notify(29946)
		return isEndOfStatement(lastTok) == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(29947)
		return (c.inCopy() && func() bool {
			__antithesis_instrumentation__.Notify(29948)
			return (strings.HasSuffix(c.concatLines, "\n"+`\.`) || func() bool {
				__antithesis_instrumentation__.Notify(29949)
				return c.atEOF == true
			}() == true) == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(29950)
		return (c.singleStatement && func() bool {
			__antithesis_instrumentation__.Notify(29951)
			return c.atEOF == true
		}() == true) == true
	}() == true
	if c.atEOF {
		__antithesis_instrumentation__.Notify(29952)

		if !endOfStmt {
			__antithesis_instrumentation__.Notify(29953)
			fmt.Fprintf(c.iCtx.stderr, "missing semicolon at end of statement: %s\n", c.concatLines)
			c.exitErr = fmt.Errorf("last statement was not executed: %s", c.concatLines)
			return cliStop
		} else {
			__antithesis_instrumentation__.Notify(29954)
		}
	} else {
		__antithesis_instrumentation__.Notify(29955)
	}
	__antithesis_instrumentation__.Notify(29934)

	if !endOfStmt {
		__antithesis_instrumentation__.Notify(29956)
		if lastTok == '?' {
			__antithesis_instrumentation__.Notify(29958)
			fmt.Fprintf(c.iCtx.stdout,
				"Note: a single '?' is a JSON operator. If you want contextual help, use '??'.\n")
		} else {
			__antithesis_instrumentation__.Notify(29959)
		}
		__antithesis_instrumentation__.Notify(29957)
		return contState
	} else {
		__antithesis_instrumentation__.Notify(29960)
	}
	__antithesis_instrumentation__.Notify(29935)

	if !c.inCopy() {
		__antithesis_instrumentation__.Notify(29961)
		c.addHistory(c.concatLines)
	} else {
		__antithesis_instrumentation__.Notify(29962)
	}
	__antithesis_instrumentation__.Notify(29936)

	if !c.iCtx.checkSyntax {
		__antithesis_instrumentation__.Notify(29963)
		return execState
	} else {
		__antithesis_instrumentation__.Notify(29964)
	}
	__antithesis_instrumentation__.Notify(29937)

	return checkState
}

func (c *cliState) doCheckStatement(startState, contState, execState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29965)

	if c.inCopy() {
		__antithesis_instrumentation__.Notify(29969)
		return execState
	} else {
		__antithesis_instrumentation__.Notify(29970)
	}
	__antithesis_instrumentation__.Notify(29966)

	helpText, err := c.serverSideParse(c.concatLines)
	if err != nil {
		__antithesis_instrumentation__.Notify(29971)
		if helpText != "" {
			__antithesis_instrumentation__.Notify(29975)

			fmt.Fprintln(c.iCtx.stdout, helpText)
		} else {
			__antithesis_instrumentation__.Notify(29976)
		}
		__antithesis_instrumentation__.Notify(29972)

		_ = c.invalidSyntaxf(
			cliStart, "statement ignored: %v",
			clierror.NewFormattedError(err, false, false),
		)

		if c.iCtx.errExit {
			__antithesis_instrumentation__.Notify(29977)
			return cliStop
		} else {
			__antithesis_instrumentation__.Notify(29978)
		}
		__antithesis_instrumentation__.Notify(29973)

		c.partialLines = c.partialLines[:c.partialStmtsLen]
		if len(c.partialLines) == 0 {
			__antithesis_instrumentation__.Notify(29979)
			return startState
		} else {
			__antithesis_instrumentation__.Notify(29980)
		}
		__antithesis_instrumentation__.Notify(29974)
		return contState
	} else {
		__antithesis_instrumentation__.Notify(29981)
	}
	__antithesis_instrumentation__.Notify(29967)

	if !c.cliCtx.IsInteractive {
		__antithesis_instrumentation__.Notify(29982)
		return execState
	} else {
		__antithesis_instrumentation__.Notify(29983)
	}
	__antithesis_instrumentation__.Notify(29968)

	nextState := execState

	c.partialStmtsLen = len(c.partialLines)

	return nextState
}

func (c *cliState) doRunStatements(nextState cliStateEnum) cliStateEnum {
	__antithesis_instrumentation__.Notify(29984)

	c.lastKnownTxnStatus = " ?"

	if c.iCtx.autoTrace != "" {
		__antithesis_instrumentation__.Notify(29990)

		c.exitErr = c.conn.Exec(
			context.Background(),
			"SET tracing = off; SET tracing = "+c.iCtx.autoTrace)
		if c.exitErr != nil {
			__antithesis_instrumentation__.Notify(29991)
			if !c.singleStatement {
				__antithesis_instrumentation__.Notify(29994)
				clierror.OutputError(c.iCtx.stderr, c.exitErr, true, false)
			} else {
				__antithesis_instrumentation__.Notify(29995)
			}
			__antithesis_instrumentation__.Notify(29992)
			if c.iCtx.errExit {
				__antithesis_instrumentation__.Notify(29996)
				return cliStop
			} else {
				__antithesis_instrumentation__.Notify(29997)
			}
			__antithesis_instrumentation__.Notify(29993)
			return nextState
		} else {
			__antithesis_instrumentation__.Notify(29998)
		}
	} else {
		__antithesis_instrumentation__.Notify(29999)
	}
	__antithesis_instrumentation__.Notify(29985)

	c.exitErr = c.runWithInterruptableCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(30000)
		if scanner.FirstLexicalToken(c.concatLines) == lexbase.COPY {
			__antithesis_instrumentation__.Notify(30003)
			return c.beginCopyFrom(ctx, c.concatLines)
		} else {
			__antithesis_instrumentation__.Notify(30004)
		}
		__antithesis_instrumentation__.Notify(30001)
		q := clisqlclient.MakeQuery(c.concatLines)
		if c.inCopy() {
			__antithesis_instrumentation__.Notify(30005)
			q = c.copyFromState.Commit(
				ctx,
				c.resetCopy,
				c.concatLines,
			)
		} else {
			__antithesis_instrumentation__.Notify(30006)
		}
		__antithesis_instrumentation__.Notify(30002)
		return c.sqlExecCtx.RunQueryAndFormatResults(
			ctx,
			c.conn,
			c.iCtx.stdout,
			c.iCtx.stderr,
			q,
		)
	})
	__antithesis_instrumentation__.Notify(29986)
	if c.exitErr != nil {
		__antithesis_instrumentation__.Notify(30007)
		if !c.singleStatement {
			__antithesis_instrumentation__.Notify(30008)
			clierror.OutputError(c.iCtx.stderr, c.exitErr, true, false)
		} else {
			__antithesis_instrumentation__.Notify(30009)
		}
	} else {
		__antithesis_instrumentation__.Notify(30010)
	}
	__antithesis_instrumentation__.Notify(29987)

	if c.iCtx.autoTrace != "" {
		__antithesis_instrumentation__.Notify(30011)

		if err := c.conn.Exec(context.Background(),
			"SET tracing = off"); err != nil {
			__antithesis_instrumentation__.Notify(30012)

			clierror.OutputError(c.iCtx.stderr, err, true, false)
			if c.exitErr == nil {
				__antithesis_instrumentation__.Notify(30013)

				c.exitErr = err
			} else {
				__antithesis_instrumentation__.Notify(30014)
			}

		} else {
			__antithesis_instrumentation__.Notify(30015)
			traceType := ""
			if strings.Contains(c.iCtx.autoTrace, "kv") {
				__antithesis_instrumentation__.Notify(30017)
				traceType = "kv"
			} else {
				__antithesis_instrumentation__.Notify(30018)
			}
			__antithesis_instrumentation__.Notify(30016)
			if err := c.runWithInterruptableCtx(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(30019)
				return c.sqlExecCtx.RunQueryAndFormatResults(ctx,
					c.conn, c.iCtx.stdout, c.iCtx.stderr,
					clisqlclient.MakeQuery(fmt.Sprintf("SHOW %s TRACE FOR SESSION", traceType)))
			}); err != nil {
				__antithesis_instrumentation__.Notify(30020)
				clierror.OutputError(c.iCtx.stderr, err, true, false)
				if c.exitErr == nil {
					__antithesis_instrumentation__.Notify(30021)

					c.exitErr = err
				} else {
					__antithesis_instrumentation__.Notify(30022)
				}

			} else {
				__antithesis_instrumentation__.Notify(30023)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(30024)
	}
	__antithesis_instrumentation__.Notify(29988)

	if c.exitErr != nil && func() bool {
		__antithesis_instrumentation__.Notify(30025)
		return c.iCtx.errExit == true
	}() == true {
		__antithesis_instrumentation__.Notify(30026)
		return cliStop
	} else {
		__antithesis_instrumentation__.Notify(30027)
	}
	__antithesis_instrumentation__.Notify(29989)

	return nextState
}

func (c *cliState) beginCopyFrom(ctx context.Context, sql string) error {
	__antithesis_instrumentation__.Notify(30028)
	c.refreshTransactionStatus()
	if c.lastKnownTxnStatus != "" {
		__antithesis_instrumentation__.Notify(30032)
		return unimplemented.Newf(
			"cli_copy_in_txn",
			"cannot use COPY inside a transaction",
		)
	} else {
		__antithesis_instrumentation__.Notify(30033)
	}
	__antithesis_instrumentation__.Notify(30029)
	copyFromState, err := clisqlclient.BeginCopyFrom(ctx, c.conn, sql)
	if err != nil {
		__antithesis_instrumentation__.Notify(30034)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30035)
	}
	__antithesis_instrumentation__.Notify(30030)
	c.copyFromState = copyFromState
	if c.cliCtx.IsInteractive {
		__antithesis_instrumentation__.Notify(30036)
		fmt.Fprintln(c.iCtx.stdout, `Enter data to be copied followed by a newline.`)
		fmt.Fprintln(c.iCtx.stdout, `End with a backslash and a period on a line by itself, or an EOF signal.`)
	} else {
		__antithesis_instrumentation__.Notify(30037)
	}
	__antithesis_instrumentation__.Notify(30031)
	return nil
}

func (c *cliState) doDecidePath() cliStateEnum {
	__antithesis_instrumentation__.Notify(30038)
	if len(c.partialLines) == 0 {
		__antithesis_instrumentation__.Notify(30040)
		return cliProcessFirstLine
	} else {
		__antithesis_instrumentation__.Notify(30041)
		if c.cliCtx.IsInteractive {
			__antithesis_instrumentation__.Notify(30042)

			return cliHandleCliCmd
		} else {
			__antithesis_instrumentation__.Notify(30043)
		}
	}
	__antithesis_instrumentation__.Notify(30039)

	return cliPrepareStatementLine
}

func NewShell(
	cliCtx *clicfg.Context,
	sqlConnCtx *clisqlclient.Context,
	sqlExecCtx *clisqlexec.Context,
	sqlCtx *Context,
	conn clisqlclient.Conn,
) Shell {
	__antithesis_instrumentation__.Notify(30044)
	return &cliState{
		cliCtx:     cliCtx,
		sqlConnCtx: sqlConnCtx,
		sqlExecCtx: sqlExecCtx,
		sqlCtx:     sqlCtx,
		iCtx: &internalContext{
			customPromptPattern: defaultPromptPattern,
		},
		conn:       conn,
		includeDir: ".",
	}
}

func (c *cliState) RunInteractive(cmdIn, cmdOut, cmdErr *os.File) (exitErr error) {
	__antithesis_instrumentation__.Notify(30045)
	finalFn := c.maybeHandleInterrupt()
	defer finalFn()

	return c.doRunShell(cliStart, cmdIn, cmdOut, cmdErr)
}

func (c *cliState) doRunShell(state cliStateEnum, cmdIn, cmdOut, cmdErr *os.File) (exitErr error) {
	__antithesis_instrumentation__.Notify(30046)
	for {
		__antithesis_instrumentation__.Notify(30048)
		if state == cliStop {
			__antithesis_instrumentation__.Notify(30050)
			break
		} else {
			__antithesis_instrumentation__.Notify(30051)
		}
		__antithesis_instrumentation__.Notify(30049)
		switch state {
		case cliStart:
			__antithesis_instrumentation__.Notify(30052)
			cleanupFn, err := c.configurePreShellDefaults(cmdIn, cmdOut, cmdErr)
			defer cleanupFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(30066)
				return err
			} else {
				__antithesis_instrumentation__.Notify(30067)
			}
			__antithesis_instrumentation__.Notify(30053)
			if len(c.sqlCtx.ExecStmts) > 0 {
				__antithesis_instrumentation__.Notify(30068)

				if err := c.runStatements(c.sqlCtx.ExecStmts); err != nil {
					__antithesis_instrumentation__.Notify(30070)
					return err
				} else {
					__antithesis_instrumentation__.Notify(30071)
				}
				__antithesis_instrumentation__.Notify(30069)
				if c.iCtx.quitAfterExecStmts {
					__antithesis_instrumentation__.Notify(30072)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(30073)
				}
			} else {
				__antithesis_instrumentation__.Notify(30074)
			}
			__antithesis_instrumentation__.Notify(30054)

			state = c.doStart(cliStartLine)

		case cliRefreshPrompt:
			__antithesis_instrumentation__.Notify(30055)
			state = c.doRefreshPrompts(cliReadLine)

		case cliStartLine:
			__antithesis_instrumentation__.Notify(30056)
			state = c.doStartLine(cliRefreshPrompt)

		case cliContinueLine:
			__antithesis_instrumentation__.Notify(30057)
			state = c.doContinueLine(cliRefreshPrompt)

		case cliReadLine:
			__antithesis_instrumentation__.Notify(30058)
			state = c.doReadLine(cliDecidePath)

		case cliDecidePath:
			__antithesis_instrumentation__.Notify(30059)
			state = c.doDecidePath()

		case cliProcessFirstLine:
			__antithesis_instrumentation__.Notify(30060)
			state = c.doProcessFirstLine(cliStartLine, cliHandleCliCmd)

		case cliHandleCliCmd:
			__antithesis_instrumentation__.Notify(30061)
			state = c.doHandleCliCmd(cliRefreshPrompt, cliPrepareStatementLine)

		case cliPrepareStatementLine:
			__antithesis_instrumentation__.Notify(30062)
			state = c.doPrepareStatementLine(
				cliStartLine, cliContinueLine, cliCheckStatement, cliRunStatement,
			)

		case cliCheckStatement:
			__antithesis_instrumentation__.Notify(30063)
			state = c.doCheckStatement(cliStartLine, cliContinueLine, cliRunStatement)

		case cliRunStatement:
			__antithesis_instrumentation__.Notify(30064)
			state = c.doRunStatements(cliStartLine)

		default:
			__antithesis_instrumentation__.Notify(30065)
			panic(fmt.Sprintf("unknown state: %d", state))
		}
	}
	__antithesis_instrumentation__.Notify(30047)

	return c.exitErr
}

func (c *cliState) configurePreShellDefaults(
	cmdIn, cmdOut, cmdErr *os.File,
) (cleanupFn func(), err error) {
	__antithesis_instrumentation__.Notify(30075)
	c.iCtx.stdout = cmdOut
	c.iCtx.stderr = cmdErr

	if c.sqlExecCtx.TerminalOutput {
		__antithesis_instrumentation__.Notify(30081)

		c.sqlExecCtx.ShowTimes = true
	} else {
		__antithesis_instrumentation__.Notify(30082)
	}
	__antithesis_instrumentation__.Notify(30076)

	if c.cliCtx.IsInteractive {
		__antithesis_instrumentation__.Notify(30083)

		c.iCtx.errExit = false
		if !c.sqlConnCtx.DebugMode {
			__antithesis_instrumentation__.Notify(30084)

			c.iCtx.checkSyntax = true
		} else {
			__antithesis_instrumentation__.Notify(30085)
		}
	} else {
		__antithesis_instrumentation__.Notify(30086)

		c.iCtx.errExit = true
		c.iCtx.checkSyntax = false

	}
	__antithesis_instrumentation__.Notify(30077)

	if c.cliCtx.IsInteractive && func() bool {
		__antithesis_instrumentation__.Notify(30087)
		return c.sqlExecCtx.TerminalOutput == true
	}() == true {
		__antithesis_instrumentation__.Notify(30088)

		c.ins, c.exitErr = readline.InitFiles("cockroach",
			true,
			cmdIn, c.iCtx.stdout, c.iCtx.stderr)
		if errors.Is(c.exitErr, readline.ErrWidecharNotSupported) {
			__antithesis_instrumentation__.Notify(30091)
			fmt.Fprintln(c.iCtx.stderr, "warning: wide character support disabled")
			c.ins, c.exitErr = readline.InitFiles("cockroach",
				false, cmdIn, c.iCtx.stdout, c.iCtx.stderr)
		} else {
			__antithesis_instrumentation__.Notify(30092)
		}
		__antithesis_instrumentation__.Notify(30089)
		if c.exitErr != nil {
			__antithesis_instrumentation__.Notify(30093)
			return cleanupFn, c.exitErr
		} else {
			__antithesis_instrumentation__.Notify(30094)
		}
		__antithesis_instrumentation__.Notify(30090)

		c.iCtx.stdout = c.ins.Stdout()

		c.ins.RebindControlKeys()
		cleanupFn = func() { __antithesis_instrumentation__.Notify(30095); c.ins.Close() }
	} else {
		__antithesis_instrumentation__.Notify(30096)
		c.ins = noLineEditor
		c.buf = bufio.NewReader(cmdIn)
		cleanupFn = func() { __antithesis_instrumentation__.Notify(30097) }
	}
	__antithesis_instrumentation__.Notify(30078)

	if c.hasEditor() {
		__antithesis_instrumentation__.Notify(30098)

		c.iCtx.customPromptPattern = defaultPromptPattern
		if c.sqlConnCtx.DebugMode {
			__antithesis_instrumentation__.Notify(30100)
			c.iCtx.customPromptPattern = debugPromptPattern
		} else {
			__antithesis_instrumentation__.Notify(30101)
		}
		__antithesis_instrumentation__.Notify(30099)

		c.ins.SetCompleter(c)
		if err := c.ins.UseHistory(-1, true); err != nil {
			__antithesis_instrumentation__.Notify(30102)
			fmt.Fprintf(c.iCtx.stderr, "warning: cannot enable history: %v\n ", err)
		} else {
			__antithesis_instrumentation__.Notify(30103)
			homeDir, err := envutil.HomeDir()
			if err != nil {
				__antithesis_instrumentation__.Notify(30104)
				fmt.Fprintf(c.iCtx.stderr, "warning: cannot retrieve user information: %v\nwarning: history will not be saved\n", err)
			} else {
				__antithesis_instrumentation__.Notify(30105)
				histFile := filepath.Join(homeDir, cmdHistFile)
				err = c.ins.LoadHistory(histFile)
				if err != nil {
					__antithesis_instrumentation__.Notify(30107)
					fmt.Fprintf(c.iCtx.stderr, "warning: cannot load the command-line history (file corrupted?): %v\n", err)
					fmt.Fprintf(c.iCtx.stderr, "note: the history file will be cleared upon first entry\n")
				} else {
					__antithesis_instrumentation__.Notify(30108)
				}
				__antithesis_instrumentation__.Notify(30106)
				c.ins.SetAutoSaveHistory(histFile, true)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(30109)
	}
	__antithesis_instrumentation__.Notify(30079)

	c.iCtx.quitAfterExecStmts = len(c.sqlCtx.ExecStmts) > 0
	if len(c.sqlCtx.SetStmts) > 0 {
		__antithesis_instrumentation__.Notify(30110)
		setStmts := make([]string, 0, len(c.sqlCtx.SetStmts)+len(c.sqlCtx.ExecStmts))
		for _, s := range c.sqlCtx.SetStmts {
			__antithesis_instrumentation__.Notify(30112)
			setStmts = append(setStmts, `\set `+s)
		}
		__antithesis_instrumentation__.Notify(30111)
		c.sqlCtx.SetStmts = nil
		c.sqlCtx.ExecStmts = append(setStmts, c.sqlCtx.ExecStmts...)
	} else {
		__antithesis_instrumentation__.Notify(30113)
	}
	__antithesis_instrumentation__.Notify(30080)

	return cleanupFn, nil
}

func (c *cliState) runStatements(stmts []string) error {
	__antithesis_instrumentation__.Notify(30114)
	for {
		__antithesis_instrumentation__.Notify(30117)
		for i, stmt := range stmts {
			__antithesis_instrumentation__.Notify(30120)
			c.exitErr = nil
			_ = c.runString(cliRunStatement, cliStop, stmt)
			if c.exitErr != nil {
				__antithesis_instrumentation__.Notify(30121)
				if !c.iCtx.errExit && func() bool {
					__antithesis_instrumentation__.Notify(30123)
					return i < len(stmts)-1 == true
				}() == true {
					__antithesis_instrumentation__.Notify(30124)

					clierror.OutputError(c.iCtx.stderr, c.exitErr, true, false)
				} else {
					__antithesis_instrumentation__.Notify(30125)
				}
				__antithesis_instrumentation__.Notify(30122)
				if c.iCtx.errExit {
					__antithesis_instrumentation__.Notify(30126)
					break
				} else {
					__antithesis_instrumentation__.Notify(30127)
				}
			} else {
				__antithesis_instrumentation__.Notify(30128)
			}
		}
		__antithesis_instrumentation__.Notify(30118)

		if c.sqlCtx.RepeatDelay > 0 && func() bool {
			__antithesis_instrumentation__.Notify(30129)
			return c.exitErr == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(30130)
			time.Sleep(c.sqlCtx.RepeatDelay)
			continue
		} else {
			__antithesis_instrumentation__.Notify(30131)
		}
		__antithesis_instrumentation__.Notify(30119)
		break
	}
	__antithesis_instrumentation__.Notify(30115)

	if c.exitErr != nil {
		__antithesis_instrumentation__.Notify(30132)

		c.exitErr = clierror.NewFormattedError(c.exitErr, true, false)
	} else {
		__antithesis_instrumentation__.Notify(30133)
	}
	__antithesis_instrumentation__.Notify(30116)
	return c.exitErr
}

func (c *cliState) serverSideParse(sql string) (helpText string, err error) {
	__antithesis_instrumentation__.Notify(30134)
	cols, rows, err := c.sqlExecCtx.RunQuery(
		context.Background(),
		c.conn,
		clisqlclient.MakeQuery("SHOW SYNTAX "+lexbase.EscapeSQLString(sql)),
		true)
	if err != nil {
		__antithesis_instrumentation__.Notify(30138)

		return "", errors.Wrap(err, "unexpected error")
	} else {
		__antithesis_instrumentation__.Notify(30139)
	}
	__antithesis_instrumentation__.Notify(30135)

	if len(cols) < 2 {
		__antithesis_instrumentation__.Notify(30140)
		return "", errors.Newf(
			"invalid results for SHOW SYNTAX: %q %q", cols, rows)
	} else {
		__antithesis_instrumentation__.Notify(30141)
	}
	__antithesis_instrumentation__.Notify(30136)

	if len(rows) >= 1 && func() bool {
		__antithesis_instrumentation__.Notify(30142)
		return rows[0][0] == "error" == true
	}() == true {
		__antithesis_instrumentation__.Notify(30143)
		var message, code, detail, hint string
		for _, row := range rows {
			__antithesis_instrumentation__.Notify(30148)
			switch row[0] {
			case "error":
				__antithesis_instrumentation__.Notify(30149)
				message = row[1]
			case "detail":
				__antithesis_instrumentation__.Notify(30150)
				detail = row[1]
			case "hint":
				__antithesis_instrumentation__.Notify(30151)
				hint = row[1]
			case "code":
				__antithesis_instrumentation__.Notify(30152)
				code = row[1]
			default:
				__antithesis_instrumentation__.Notify(30153)
			}
		}
		__antithesis_instrumentation__.Notify(30144)

		if strings.HasPrefix(message, "help token in input") && func() bool {
			__antithesis_instrumentation__.Notify(30154)
			return strings.HasPrefix(hint, "help:") == true
		}() == true {
			__antithesis_instrumentation__.Notify(30155)

			helpText = hint[6:]
			hint = ""
		} else {
			__antithesis_instrumentation__.Notify(30156)
		}
		__antithesis_instrumentation__.Notify(30145)

		err := errors.Newf("%s", message)
		err = pgerror.WithCandidateCode(err, pgcode.MakeCode(code))
		if hint != "" {
			__antithesis_instrumentation__.Notify(30157)
			err = errors.WithHint(err, hint)
		} else {
			__antithesis_instrumentation__.Notify(30158)
		}
		__antithesis_instrumentation__.Notify(30146)
		if detail != "" {
			__antithesis_instrumentation__.Notify(30159)
			err = errors.WithDetail(err, detail)
		} else {
			__antithesis_instrumentation__.Notify(30160)
		}
		__antithesis_instrumentation__.Notify(30147)
		return helpText, err
	} else {
		__antithesis_instrumentation__.Notify(30161)
	}
	__antithesis_instrumentation__.Notify(30137)
	return "", nil
}

const queryCancelEnabled = false

func (c *cliState) maybeHandleInterrupt() func() {
	__antithesis_instrumentation__.Notify(30162)
	if !c.cliCtx.IsInteractive {
		__antithesis_instrumentation__.Notify(30165)
		return func() { __antithesis_instrumentation__.Notify(30166) }
	} else {
		__antithesis_instrumentation__.Notify(30167)
	}
	__antithesis_instrumentation__.Notify(30163)
	intCh := make(chan os.Signal, 1)
	signal.Notify(intCh, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	if !queryCancelEnabled {
		__antithesis_instrumentation__.Notify(30168)
		go func() {
			__antithesis_instrumentation__.Notify(30169)
			for {
				__antithesis_instrumentation__.Notify(30170)
				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(30172)

					return
				case <-intCh:
					__antithesis_instrumentation__.Notify(30173)
				}
				__antithesis_instrumentation__.Notify(30171)
				fmt.Fprintln(c.iCtx.stderr,
					"query cancellation disabled in this client; a second interrupt will stop the shell.")
				signal.Reset(os.Interrupt)
			}
		}()
	} else {
		__antithesis_instrumentation__.Notify(30174)
		go func() {
			__antithesis_instrumentation__.Notify(30175)
			for {
				__antithesis_instrumentation__.Notify(30176)
				select {
				case <-intCh:
					__antithesis_instrumentation__.Notify(30177)
					c.iCtx.mu.Lock()
					cancelFn, doneCh := c.iCtx.mu.cancelFn, c.iCtx.mu.doneCh
					c.iCtx.mu.Unlock()
					if cancelFn == nil {
						__antithesis_instrumentation__.Notify(30181)

						continue
					} else {
						__antithesis_instrumentation__.Notify(30182)
					}
					__antithesis_instrumentation__.Notify(30178)

					fmt.Fprintf(c.iCtx.stderr, "\nattempting to cancel query...\n")

					cancelFn()

					tooLongTimer := time.After(3 * time.Second)
				wait:
					for {
						__antithesis_instrumentation__.Notify(30183)
						select {
						case <-doneCh:
							__antithesis_instrumentation__.Notify(30184)
							break wait
						case <-tooLongTimer:
							__antithesis_instrumentation__.Notify(30185)
							fmt.Fprintln(c.iCtx.stderr, "server does not respond to query cancellation; a second interrupt will stop the shell.")
							signal.Reset(os.Interrupt)
						}
					}
					__antithesis_instrumentation__.Notify(30179)

					signal.Notify(intCh, os.Interrupt)

				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(30180)

					return
				}
			}
		}()
	}
	__antithesis_instrumentation__.Notify(30164)
	return cancel
}

func (c *cliState) runWithInterruptableCtx(fn func(ctx context.Context) error) error {
	__antithesis_instrumentation__.Notify(30186)
	if !c.cliCtx.IsInteractive {
		__antithesis_instrumentation__.Notify(30190)
		return fn(context.Background())
	} else {
		__antithesis_instrumentation__.Notify(30191)
	}
	__antithesis_instrumentation__.Notify(30187)

	ctx, cancel := context.WithCancel(context.Background())

	doneCh := make(chan struct{})
	defer func() { __antithesis_instrumentation__.Notify(30192); close(doneCh) }()
	__antithesis_instrumentation__.Notify(30188)

	c.iCtx.mu.Lock()
	c.iCtx.mu.cancelFn = cancel
	c.iCtx.mu.doneCh = doneCh
	c.iCtx.mu.Unlock()
	defer func() {
		__antithesis_instrumentation__.Notify(30193)
		c.iCtx.mu.Lock()
		c.iCtx.mu.cancelFn = nil
		c.iCtx.mu.doneCh = nil
		c.iCtx.mu.Unlock()
	}()
	__antithesis_instrumentation__.Notify(30189)

	err := fn(ctx)
	return err
}
