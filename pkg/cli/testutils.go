package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cli/clierror"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/security/securitytest"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/kr/pretty"
)

func TestingReset() {
	__antithesis_instrumentation__.Notify(34659)

	initCLIDefaults()
}

type TestCLI struct {
	*server.TestServer
	tenant      serverutils.TestTenantInterface
	certsDir    string
	cleanupFunc func() error
	prevStderr  *os.File

	t *testing.T

	logScope *log.TestLogScope

	omitArgs bool

	reportExitCode bool
}

type TestCLIParams struct {
	T        *testing.T
	Insecure bool

	NoServer bool

	StoreSpecs []base.StoreSpec

	Locality roachpb.Locality

	NoNodelocal bool

	TenantArgs *base.TestTenantArgs
}

const testTempFilePrefix = "test-temp-prefix-"

const testUserfileUploadTempDirPrefix = "test-userfile-upload-temp-dir-"

func (c *TestCLI) fail(err interface{}) {
	__antithesis_instrumentation__.Notify(34660)
	if c.t != nil {
		__antithesis_instrumentation__.Notify(34661)
		defer c.logScope.Close(c.t)
		c.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(34662)
		panic(err)
	}
}

func NewCLITest(params TestCLIParams) TestCLI {
	__antithesis_instrumentation__.Notify(34663)
	return newCLITestWithArgs(params, nil)
}

func newCLITestWithArgs(params TestCLIParams, argsFn func(args *base.TestServerArgs)) TestCLI {
	__antithesis_instrumentation__.Notify(34664)
	c := TestCLI{t: params.T}

	certsDir, err := ioutil.TempDir("", "cli-test")
	if err != nil {
		__antithesis_instrumentation__.Notify(34670)
		c.fail(err)
	} else {
		__antithesis_instrumentation__.Notify(34671)
	}
	__antithesis_instrumentation__.Notify(34665)
	c.certsDir = certsDir

	if c.t != nil {
		__antithesis_instrumentation__.Notify(34672)
		c.logScope = log.Scope(c.t)
	} else {
		__antithesis_instrumentation__.Notify(34673)
	}
	__antithesis_instrumentation__.Notify(34666)

	c.cleanupFunc = func() error { __antithesis_instrumentation__.Notify(34674); return nil }
	__antithesis_instrumentation__.Notify(34667)

	if !params.NoServer {
		__antithesis_instrumentation__.Notify(34675)
		if !params.Insecure {
			__antithesis_instrumentation__.Notify(34680)
			c.cleanupFunc = securitytest.CreateTestCerts(certsDir)
		} else {
			__antithesis_instrumentation__.Notify(34681)
		}
		__antithesis_instrumentation__.Notify(34676)

		args := base.TestServerArgs{
			Insecure:      params.Insecure,
			SSLCertsDir:   c.certsDir,
			StoreSpecs:    params.StoreSpecs,
			Locality:      params.Locality,
			ExternalIODir: filepath.Join(certsDir, "extern"),
			Knobs: base.TestingKnobs{
				SQLStatsKnobs: &sqlstats.TestingKnobs{
					AOSTClause: "AS OF SYSTEM TIME '-1us'",
				},
			},
		}
		if argsFn != nil {
			__antithesis_instrumentation__.Notify(34682)
			argsFn(&args)
		} else {
			__antithesis_instrumentation__.Notify(34683)
		}
		__antithesis_instrumentation__.Notify(34677)
		if params.NoNodelocal {
			__antithesis_instrumentation__.Notify(34684)
			args.ExternalIODir = ""
		} else {
			__antithesis_instrumentation__.Notify(34685)
		}
		__antithesis_instrumentation__.Notify(34678)
		s, err := serverutils.StartServerRaw(args)
		if err != nil {
			__antithesis_instrumentation__.Notify(34686)
			c.fail(err)
		} else {
			__antithesis_instrumentation__.Notify(34687)
		}
		__antithesis_instrumentation__.Notify(34679)
		c.TestServer = s.(*server.TestServer)

		log.Infof(context.Background(), "server started at %s", c.ServingRPCAddr())
		log.Infof(context.Background(), "SQL listener at %s", c.ServingSQLAddr())
	} else {
		__antithesis_instrumentation__.Notify(34688)
	}
	__antithesis_instrumentation__.Notify(34668)

	if params.TenantArgs != nil {
		__antithesis_instrumentation__.Notify(34689)
		if c.TestServer == nil {
			__antithesis_instrumentation__.Notify(34692)
			c.fail(errors.AssertionFailedf("multitenant mode for CLI requires a DB server, try setting `NoServer` argument to false"))
		} else {
			__antithesis_instrumentation__.Notify(34693)
		}
		__antithesis_instrumentation__.Notify(34690)
		if c.Insecure() {
			__antithesis_instrumentation__.Notify(34694)
			params.TenantArgs.ForceInsecure = true
		} else {
			__antithesis_instrumentation__.Notify(34695)
		}
		__antithesis_instrumentation__.Notify(34691)
		c.tenant, _ = serverutils.StartTenant(c.t, c.TestServer, *params.TenantArgs)
	} else {
		__antithesis_instrumentation__.Notify(34696)
	}
	__antithesis_instrumentation__.Notify(34669)
	baseCfg.User = security.NodeUserName()

	c.prevStderr = stderr
	stderr = os.Stdout

	return c
}

func setCLIDefaultsForTests() {
	__antithesis_instrumentation__.Notify(34697)
	initCLIDefaults()
	sqlExecCtx.TerminalOutput = false
	sqlExecCtx.ShowTimes = false

	sqlExecCtx.TableDisplayFormat = clisqlexec.TableDisplayTable
}

func (c *TestCLI) stopServer() {
	__antithesis_instrumentation__.Notify(34698)
	if c.TestServer != nil {
		__antithesis_instrumentation__.Notify(34699)
		log.Infof(context.Background(), "stopping server at %s / %s",
			c.ServingRPCAddr(), c.ServingSQLAddr())
		c.Stopper().Stop(context.Background())
	} else {
		__antithesis_instrumentation__.Notify(34700)
	}
}

func (c *TestCLI) RestartServer(params TestCLIParams) {
	__antithesis_instrumentation__.Notify(34701)
	c.stopServer()
	log.Info(context.Background(), "restarting server")
	s, err := serverutils.StartServerRaw(base.TestServerArgs{
		Insecure:    params.Insecure,
		SSLCertsDir: c.certsDir,
		StoreSpecs:  params.StoreSpecs,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(34703)
		c.fail(err)
	} else {
		__antithesis_instrumentation__.Notify(34704)
	}
	__antithesis_instrumentation__.Notify(34702)
	c.TestServer = s.(*server.TestServer)
	log.Infof(context.Background(), "restarted server at %s / %s",
		c.ServingRPCAddr(), c.ServingSQLAddr())
	if params.TenantArgs != nil {
		__antithesis_instrumentation__.Notify(34705)
		if c.Insecure() {
			__antithesis_instrumentation__.Notify(34707)
			params.TenantArgs.ForceInsecure = true
		} else {
			__antithesis_instrumentation__.Notify(34708)
		}
		__antithesis_instrumentation__.Notify(34706)
		c.tenant, _ = serverutils.StartTenant(c.t, c.TestServer, *params.TenantArgs)
		log.Infof(context.Background(), "restarted tenant SQL only server at %s", c.tenant.SQLAddr())
	} else {
		__antithesis_instrumentation__.Notify(34709)
	}
}

func (c *TestCLI) Cleanup() {
	__antithesis_instrumentation__.Notify(34710)
	defer func() {
		__antithesis_instrumentation__.Notify(34712)
		if c.t != nil {
			__antithesis_instrumentation__.Notify(34713)
			c.logScope.Close(c.t)
		} else {
			__antithesis_instrumentation__.Notify(34714)
		}
	}()
	__antithesis_instrumentation__.Notify(34711)

	stderr = c.prevStderr

	log.Info(context.Background(), "stopping server and cleaning up CLI test")

	c.stopServer()

	if err := c.cleanupFunc(); err != nil {
		__antithesis_instrumentation__.Notify(34715)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(34716)
	}
}

func (c TestCLI) Run(line string) {
	__antithesis_instrumentation__.Notify(34717)
	a := strings.Fields(line)
	c.RunWithArgs(a)
}

func (c TestCLI) RunWithCapture(line string) (out string, err error) {
	__antithesis_instrumentation__.Notify(34718)
	return captureOutput(func() {
		__antithesis_instrumentation__.Notify(34719)
		c.Run(line)
	})
}

func (c TestCLI) RunWithCaptureArgs(args []string) (string, error) {
	__antithesis_instrumentation__.Notify(34720)
	return captureOutput(func() {
		__antithesis_instrumentation__.Notify(34721)
		c.RunWithArgs(args)
	})
}

func captureOutput(f func()) (out string, err error) {
	__antithesis_instrumentation__.Notify(34722)

	stdoutSave, stderrRedirSave := os.Stdout, stderr
	r, w, err := os.Pipe()
	if err != nil {
		__antithesis_instrumentation__.Notify(34726)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(34727)
	}
	__antithesis_instrumentation__.Notify(34723)
	os.Stdout = w
	stderr = w

	type captureResult struct {
		out string
		err error
	}
	outC := make(chan captureResult)
	go func() {
		__antithesis_instrumentation__.Notify(34728)
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		r.Close()
		outC <- captureResult{buf.String(), err}
	}()
	__antithesis_instrumentation__.Notify(34724)

	defer func() {
		__antithesis_instrumentation__.Notify(34729)

		w.Close()
		os.Stdout = stdoutSave
		stderr = stderrRedirSave
		outResult := <-outC
		out, err = outResult.out, outResult.err
		if x := recover(); x != nil {
			__antithesis_instrumentation__.Notify(34730)
			err = errors.Errorf("panic: %v", x)
		} else {
			__antithesis_instrumentation__.Notify(34731)
		}
	}()
	__antithesis_instrumentation__.Notify(34725)

	f()
	return
}

func isSQLCommand(args []string) (bool, error) {
	__antithesis_instrumentation__.Notify(34732)
	cmd, _, err := cockroachCmd.Find(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(34735)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(34736)
	}
	__antithesis_instrumentation__.Notify(34733)

	if f := flagSetForCmd(cmd).Lookup(cliflags.EchoSQL.Name); f != nil {
		__antithesis_instrumentation__.Notify(34737)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(34738)
	}
	__antithesis_instrumentation__.Notify(34734)
	return false, nil
}

func (c TestCLI) getRPCAddr() string {
	__antithesis_instrumentation__.Notify(34739)
	if c.tenant != nil {
		__antithesis_instrumentation__.Notify(34741)
		return c.tenant.RPCAddr()
	} else {
		__antithesis_instrumentation__.Notify(34742)
	}
	__antithesis_instrumentation__.Notify(34740)
	return c.ServingRPCAddr()
}

func (c TestCLI) getSQLAddr() string {
	__antithesis_instrumentation__.Notify(34743)
	if c.tenant != nil {
		__antithesis_instrumentation__.Notify(34745)
		return c.tenant.SQLAddr()
	} else {
		__antithesis_instrumentation__.Notify(34746)
	}
	__antithesis_instrumentation__.Notify(34744)
	return c.ServingSQLAddr()
}

func (c TestCLI) RunWithArgs(origArgs []string) {
	__antithesis_instrumentation__.Notify(34747)
	TestingReset()

	if err := func() error {
		__antithesis_instrumentation__.Notify(34748)
		args := append([]string(nil), origArgs[:1]...)
		if c.TestServer != nil {
			__antithesis_instrumentation__.Notify(34753)
			addr := c.getRPCAddr()
			if isSQL, err := isSQLCommand(origArgs); err != nil {
				__antithesis_instrumentation__.Notify(34756)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34757)
				if isSQL {
					__antithesis_instrumentation__.Notify(34758)
					addr = c.getSQLAddr()
				} else {
					__antithesis_instrumentation__.Notify(34759)
				}
			}
			__antithesis_instrumentation__.Notify(34754)
			h, p, err := net.SplitHostPort(addr)
			if err != nil {
				__antithesis_instrumentation__.Notify(34760)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34761)
			}
			__antithesis_instrumentation__.Notify(34755)
			args = append(args, fmt.Sprintf("--host=%s", net.JoinHostPort(h, p)))
			if c.Cfg.Insecure {
				__antithesis_instrumentation__.Notify(34762)
				args = append(args, "--insecure=true")
			} else {
				__antithesis_instrumentation__.Notify(34763)
				args = append(args, "--insecure=false")
				args = append(args, fmt.Sprintf("--certs-dir=%s", c.certsDir))
			}
		} else {
			__antithesis_instrumentation__.Notify(34764)
		}
		__antithesis_instrumentation__.Notify(34749)

		args = append(args, origArgs[1:]...)

		if len(origArgs) >= 3 && func() bool {
			__antithesis_instrumentation__.Notify(34765)
			return strings.Contains(origArgs[2], testTempFilePrefix) == true
		}() == true {
			__antithesis_instrumentation__.Notify(34766)
			splitFilePath := strings.Split(origArgs[2], testTempFilePrefix)
			origArgs[2] = splitFilePath[1]
		} else {
			__antithesis_instrumentation__.Notify(34767)
		}
		__antithesis_instrumentation__.Notify(34750)
		if len(origArgs) >= 4 && func() bool {
			__antithesis_instrumentation__.Notify(34768)
			return strings.Contains(origArgs[3], testUserfileUploadTempDirPrefix) == true
		}() == true {
			__antithesis_instrumentation__.Notify(34769)
			hasTrailingSlash := strings.HasSuffix(origArgs[3], "/")
			origArgs[3] = filepath.Base(origArgs[3])

			if hasTrailingSlash {
				__antithesis_instrumentation__.Notify(34770)
				origArgs[3] += "/"
			} else {
				__antithesis_instrumentation__.Notify(34771)
			}
		} else {
			__antithesis_instrumentation__.Notify(34772)
		}
		__antithesis_instrumentation__.Notify(34751)

		if !c.omitArgs {
			__antithesis_instrumentation__.Notify(34773)
			fmt.Fprintf(os.Stderr, "%s\n", args)
			fmt.Println(strings.Join(origArgs, " "))
		} else {
			__antithesis_instrumentation__.Notify(34774)
		}
		__antithesis_instrumentation__.Notify(34752)

		return Run(args)
	}(); err != nil {
		__antithesis_instrumentation__.Notify(34775)
		clierror.OutputError(os.Stdout, err, true, false)
		if c.reportExitCode {
			__antithesis_instrumentation__.Notify(34776)
			fmt.Fprintln(os.Stdout, "exit code:", getExitCode(err))
		} else {
			__antithesis_instrumentation__.Notify(34777)
		}
	} else {
		__antithesis_instrumentation__.Notify(34778)
		if c.reportExitCode {
			__antithesis_instrumentation__.Notify(34779)
			fmt.Fprintln(os.Stdout, "exit code:", exit.Success())
		} else {
			__antithesis_instrumentation__.Notify(34780)
		}
	}
}

func (c TestCLI) RunWithCAArgs(origArgs []string) {
	__antithesis_instrumentation__.Notify(34781)
	TestingReset()

	if err := func() error {
		__antithesis_instrumentation__.Notify(34782)
		args := append([]string(nil), origArgs[:1]...)
		if c.TestServer != nil {
			__antithesis_instrumentation__.Notify(34784)
			args = append(args, fmt.Sprintf("--ca-key=%s", filepath.Join(c.certsDir, security.EmbeddedCAKey)))
			args = append(args, fmt.Sprintf("--certs-dir=%s", c.certsDir))
		} else {
			__antithesis_instrumentation__.Notify(34785)
		}
		__antithesis_instrumentation__.Notify(34783)
		args = append(args, origArgs[1:]...)

		fmt.Fprintf(os.Stderr, "%s\n", args)
		fmt.Println(strings.Join(origArgs, " "))

		return Run(args)
	}(); err != nil {
		__antithesis_instrumentation__.Notify(34786)
		fmt.Println(err)
	} else {
		__antithesis_instrumentation__.Notify(34787)
	}
}

func ElideInsecureDeprecationNotice(csvStr string) string {
	__antithesis_instrumentation__.Notify(34788)

	lines := strings.SplitN(csvStr, "\n", 3)
	if len(lines) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(34790)
		return strings.HasPrefix(lines[0], "Flag --insecure has been deprecated") == true
	}() == true {
		__antithesis_instrumentation__.Notify(34791)
		csvStr = lines[2]
	} else {
		__antithesis_instrumentation__.Notify(34792)
	}
	__antithesis_instrumentation__.Notify(34789)
	return csvStr
}

func RemoveMatchingLines(output string, regexps []string) string {
	__antithesis_instrumentation__.Notify(34793)
	if len(regexps) == 0 {
		__antithesis_instrumentation__.Notify(34798)
		return output
	} else {
		__antithesis_instrumentation__.Notify(34799)
	}
	__antithesis_instrumentation__.Notify(34794)

	var patterns []*regexp.Regexp
	for _, weed := range regexps {
		__antithesis_instrumentation__.Notify(34800)
		p := regexp.MustCompile(weed)
		patterns = append(patterns, p)
	}
	__antithesis_instrumentation__.Notify(34795)
	filter := func(line string) bool {
		__antithesis_instrumentation__.Notify(34801)
		for _, pattern := range patterns {
			__antithesis_instrumentation__.Notify(34803)
			if pattern.MatchString(line) {
				__antithesis_instrumentation__.Notify(34804)
				return true
			} else {
				__antithesis_instrumentation__.Notify(34805)
			}
		}
		__antithesis_instrumentation__.Notify(34802)
		return false
	}
	__antithesis_instrumentation__.Notify(34796)

	result := strings.Builder{}
	for _, line := range strings.Split(output, "\n") {
		__antithesis_instrumentation__.Notify(34806)
		if filter(line) || func() bool {
			__antithesis_instrumentation__.Notify(34808)
			return len(line) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(34809)
			continue
		} else {
			__antithesis_instrumentation__.Notify(34810)
		}
		__antithesis_instrumentation__.Notify(34807)
		result.WriteString(line)
		result.WriteRune('\n')
	}
	__antithesis_instrumentation__.Notify(34797)
	return result.String()
}

func GetCsvNumCols(csvStr string) (cols int, err error) {
	__antithesis_instrumentation__.Notify(34811)
	csvStr = ElideInsecureDeprecationNotice(csvStr)
	reader := csv.NewReader(strings.NewReader(csvStr))
	records, err := reader.Read()
	if err != nil {
		__antithesis_instrumentation__.Notify(34813)
		return 0, errors.Wrapf(err, "error reading csv input:\n %v\n", csvStr)
	} else {
		__antithesis_instrumentation__.Notify(34814)
	}
	__antithesis_instrumentation__.Notify(34812)
	return len(records), nil
}

func MatchCSV(csvStr string, matchColRow [][]string) (err error) {
	__antithesis_instrumentation__.Notify(34815)
	defer func() {
		__antithesis_instrumentation__.Notify(34820)
		if err != nil {
			__antithesis_instrumentation__.Notify(34821)
			err = errors.Wrapf(err, "csv input:\n%v\nexpected:\n%s\n",
				csvStr, pretty.Sprint(matchColRow))
		} else {
			__antithesis_instrumentation__.Notify(34822)
		}
	}()
	__antithesis_instrumentation__.Notify(34816)

	csvStr = ElideInsecureDeprecationNotice(csvStr)
	reader := csv.NewReader(strings.NewReader(csvStr))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		__antithesis_instrumentation__.Notify(34823)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34824)
	}
	__antithesis_instrumentation__.Notify(34817)

	lr, lm := len(records), len(matchColRow)
	if lr < lm {
		__antithesis_instrumentation__.Notify(34825)
		return errors.Errorf("csv has %d rows, but expected at least %d", lr, lm)
	} else {
		__antithesis_instrumentation__.Notify(34826)
	}
	__antithesis_instrumentation__.Notify(34818)

	records = records[lr-lm:]

	for i := range records {
		__antithesis_instrumentation__.Notify(34827)
		if lr, lm := len(records[i]), len(matchColRow[i]); lr != lm {
			__antithesis_instrumentation__.Notify(34829)
			return errors.Errorf("row #%d: csv has %d columns, but expected %d", i+1, lr, lm)
		} else {
			__antithesis_instrumentation__.Notify(34830)
		}
		__antithesis_instrumentation__.Notify(34828)
		for j := range records[i] {
			__antithesis_instrumentation__.Notify(34831)
			pat, str := matchColRow[i][j], records[i][j]
			re := regexp.MustCompile(pat)
			if !re.MatchString(str) {
				__antithesis_instrumentation__.Notify(34832)
				err = errors.Wrapf(err, "row #%d, col #%d: found %q which does not match %q",
					i+1, j+1, str, pat)
			} else {
				__antithesis_instrumentation__.Notify(34833)
			}
		}
	}
	__antithesis_instrumentation__.Notify(34819)
	return err
}
