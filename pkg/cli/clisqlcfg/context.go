// Package clisqlcfg defines configuration settings and mechanisms for
// instances of the SQL shell.
package clisqlcfg

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/cli/clicfg"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlshell"
	"github.com/cockroachdb/cockroach/pkg/server/pgurl"
	"github.com/cockroachdb/errors"
	isatty "github.com/mattn/go-isatty"
)

type Context struct {
	CliCtx *clicfg.Context

	ConnCtx *clisqlclient.Context

	ExecCtx *clisqlexec.Context

	ShellCtx clisqlshell.Context

	ApplicationName string

	Database string

	User string

	ConnectTimeout int

	CmdOut *os.File

	CmdErr *os.File

	InputFile string

	SafeUpdates OptBool

	ReadOnly bool

	opened bool
	cmdIn  *os.File
}

func (c *Context) LoadDefaults(cmdOut, cmdErr *os.File) {
	__antithesis_instrumentation__.Notify(28447)
	*c = Context{CliCtx: c.CliCtx, ConnCtx: c.ConnCtx, ExecCtx: c.ExecCtx}
	c.ExecCtx.TerminalOutput = isatty.IsTerminal(cmdOut.Fd())
	c.ExecCtx.TableDisplayFormat = clisqlexec.TableDisplayTSV
	c.ExecCtx.TableBorderMode = 0
	if c.ExecCtx.TerminalOutput {
		__antithesis_instrumentation__.Notify(28451)

		c.ExecCtx.TableDisplayFormat = clisqlexec.TableDisplayTable
	} else {
		__antithesis_instrumentation__.Notify(28452)
	}
	__antithesis_instrumentation__.Notify(28448)
	if cmdOut == nil {
		__antithesis_instrumentation__.Notify(28453)
		cmdOut = os.Stdout
	} else {
		__antithesis_instrumentation__.Notify(28454)
	}
	__antithesis_instrumentation__.Notify(28449)
	if cmdErr == nil {
		__antithesis_instrumentation__.Notify(28455)
		cmdErr = os.Stderr
	} else {
		__antithesis_instrumentation__.Notify(28456)
	}
	__antithesis_instrumentation__.Notify(28450)
	c.CmdOut = cmdOut
	c.CmdErr = cmdErr
}

func (c *Context) Open(defaultInput *os.File) (closeFn func(), err error) {
	__antithesis_instrumentation__.Notify(28457)
	if c.opened {
		__antithesis_instrumentation__.Notify(28460)
		return nil, errors.AssertionFailedf("programming error: Open called twice")
	} else {
		__antithesis_instrumentation__.Notify(28461)
	}
	__antithesis_instrumentation__.Notify(28458)

	c.cmdIn, closeFn, err = c.getInputFile(defaultInput)
	if err != nil {
		__antithesis_instrumentation__.Notify(28462)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28463)
	}
	__antithesis_instrumentation__.Notify(28459)
	c.checkInteractive()
	c.opened = true
	return closeFn, err
}

func (c *Context) getInputFile(defaultIn *os.File) (cmdIn *os.File, closeFn func(), err error) {
	__antithesis_instrumentation__.Notify(28464)
	if c.InputFile == "" {
		__antithesis_instrumentation__.Notify(28468)
		return defaultIn, func() { __antithesis_instrumentation__.Notify(28469) }, nil
	} else {
		__antithesis_instrumentation__.Notify(28470)
	}
	__antithesis_instrumentation__.Notify(28465)

	if len(c.ShellCtx.ExecStmts) != 0 {
		__antithesis_instrumentation__.Notify(28471)
		return nil, nil, errors.New("cannot specify both an input file and discrete statements")
	} else {
		__antithesis_instrumentation__.Notify(28472)
	}
	__antithesis_instrumentation__.Notify(28466)

	f, err := os.Open(c.InputFile)
	if err != nil {
		__antithesis_instrumentation__.Notify(28473)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(28474)
	}
	__antithesis_instrumentation__.Notify(28467)
	return f, func() { __antithesis_instrumentation__.Notify(28475); _ = f.Close() }, nil
}

func (c *Context) checkInteractive() {
	__antithesis_instrumentation__.Notify(28476)

	c.CliCtx.IsInteractive = len(c.ShellCtx.ExecStmts) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(28477)
		return isatty.IsTerminal(c.cmdIn.Fd()) == true
	}() == true
}

func (c *Context) MakeConn(url string) (clisqlclient.Conn, error) {
	__antithesis_instrumentation__.Notify(28478)
	baseURL, err := pgurl.Parse(url)
	if err != nil {
		__antithesis_instrumentation__.Notify(28483)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28484)
	}
	__antithesis_instrumentation__.Notify(28479)

	if c.Database != "" {
		__antithesis_instrumentation__.Notify(28485)
		baseURL.WithDefaultDatabase(c.Database)
	} else {
		__antithesis_instrumentation__.Notify(28486)
	}
	__antithesis_instrumentation__.Notify(28480)

	baseURL.WithDefaultUsername(c.User)

	if prevAppName := baseURL.GetOption("application_name"); prevAppName == "" && func() bool {
		__antithesis_instrumentation__.Notify(28487)
		return c.ApplicationName != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(28488)
		_ = baseURL.SetOption("application_name", c.ApplicationName)
	} else {
		__antithesis_instrumentation__.Notify(28489)
	}
	__antithesis_instrumentation__.Notify(28481)

	if baseURL.GetOption("connect_timeout") == "" && func() bool {
		__antithesis_instrumentation__.Notify(28490)
		return c.ConnectTimeout != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(28491)
		_ = baseURL.SetOption("connect_timeout", strconv.Itoa(c.ConnectTimeout))
	} else {
		__antithesis_instrumentation__.Notify(28492)
	}
	__antithesis_instrumentation__.Notify(28482)

	usePw, pwdSet, _ := baseURL.GetAuthnPassword()
	url = baseURL.ToPQ().String()

	conn := c.ConnCtx.MakeSQLConn(c.CmdOut, c.CmdErr, url)
	conn.SetMissingPassword(!usePw || func() bool {
		__antithesis_instrumentation__.Notify(28493)
		return !pwdSet == true
	}() == true)

	return conn, nil
}

func (c *Context) Run(conn clisqlclient.Conn) error {
	__antithesis_instrumentation__.Notify(28494)
	if !c.opened {
		__antithesis_instrumentation__.Notify(28498)
		return errors.AssertionFailedf("programming error: Open not called yet")
	} else {
		__antithesis_instrumentation__.Notify(28499)
	}
	__antithesis_instrumentation__.Notify(28495)

	if err := conn.EnsureConn(); err != nil {
		__antithesis_instrumentation__.Notify(28500)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28501)
	}
	__antithesis_instrumentation__.Notify(28496)

	c.maybeSetSafeUpdates(conn)
	if err := c.maybeSetReadOnly(conn); err != nil {
		__antithesis_instrumentation__.Notify(28502)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28503)
	}
	__antithesis_instrumentation__.Notify(28497)

	shell := clisqlshell.NewShell(c.CliCtx, c.ConnCtx, c.ExecCtx, &c.ShellCtx, conn)
	return shell.RunInteractive(c.cmdIn, c.CmdOut, c.CmdErr)
}

func (c *Context) maybeSetSafeUpdates(conn clisqlclient.Conn) {
	__antithesis_instrumentation__.Notify(28504)
	hasSafeUpdates, safeUpdates := c.SafeUpdates.Get()
	if !hasSafeUpdates {
		__antithesis_instrumentation__.Notify(28506)
		safeUpdates = c.CliCtx.IsInteractive
	} else {
		__antithesis_instrumentation__.Notify(28507)
	}
	__antithesis_instrumentation__.Notify(28505)
	if safeUpdates {
		__antithesis_instrumentation__.Notify(28508)
		if err := conn.Exec(context.Background(),
			"SET sql_safe_updates = TRUE"); err != nil {
			__antithesis_instrumentation__.Notify(28509)

			fmt.Fprintf(c.CmdErr, "warning: cannot enable safe updates: %v\n", err)
		} else {
			__antithesis_instrumentation__.Notify(28510)
		}
	} else {
		__antithesis_instrumentation__.Notify(28511)
	}
}

func (c *Context) maybeSetReadOnly(conn clisqlclient.Conn) error {
	__antithesis_instrumentation__.Notify(28512)
	if !c.ReadOnly {
		__antithesis_instrumentation__.Notify(28514)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(28515)
	}
	__antithesis_instrumentation__.Notify(28513)
	return conn.Exec(context.Background(),
		"SET default_transaction_read_only = TRUE")
}
