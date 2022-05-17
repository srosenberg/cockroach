package install

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/alessio/shellescape"
	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/ssh"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/errors"
)

var startScript string

func cockroachNodeBinary(c *SyncedCluster, node Node) string {
	__antithesis_instrumentation__.Notify(181208)
	if filepath.IsAbs(c.Binary) {
		__antithesis_instrumentation__.Notify(181214)
		return c.Binary
	} else {
		__antithesis_instrumentation__.Notify(181215)
	}
	__antithesis_instrumentation__.Notify(181209)
	if !c.IsLocal() {
		__antithesis_instrumentation__.Notify(181216)
		return "./" + c.Binary
	} else {
		__antithesis_instrumentation__.Notify(181217)
	}
	__antithesis_instrumentation__.Notify(181210)

	path := filepath.Join(c.localVMDir(node), c.Binary)
	if _, err := os.Stat(path); err == nil {
		__antithesis_instrumentation__.Notify(181218)
		return path
	} else {
		__antithesis_instrumentation__.Notify(181219)
	}
	__antithesis_instrumentation__.Notify(181211)

	path, err := exec.LookPath(c.Binary)
	if err != nil {
		__antithesis_instrumentation__.Notify(181220)
		if strings.HasPrefix(c.Binary, "/") {
			__antithesis_instrumentation__.Notify(181223)
			return c.Binary
		} else {
			__antithesis_instrumentation__.Notify(181224)
		}
		__antithesis_instrumentation__.Notify(181221)

		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			__antithesis_instrumentation__.Notify(181225)
			return c.Binary
		} else {
			__antithesis_instrumentation__.Notify(181226)
		}
		__antithesis_instrumentation__.Notify(181222)
		path = gopath + "/src/github.com/cockroachdb/cockroach/" + c.Binary
		var err2 error
		path, err2 = exec.LookPath(path)
		if err2 != nil {
			__antithesis_instrumentation__.Notify(181227)
			return c.Binary
		} else {
			__antithesis_instrumentation__.Notify(181228)
		}
	} else {
		__antithesis_instrumentation__.Notify(181229)
	}
	__antithesis_instrumentation__.Notify(181212)
	path, err = filepath.Abs(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(181230)
		return c.Binary
	} else {
		__antithesis_instrumentation__.Notify(181231)
	}
	__antithesis_instrumentation__.Notify(181213)
	return path
}

func getCockroachVersion(
	ctx context.Context, c *SyncedCluster, node Node,
) (*version.Version, error) {
	__antithesis_instrumentation__.Notify(181232)
	sess, err := c.newSession(node)
	if err != nil {
		__antithesis_instrumentation__.Notify(181235)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181236)
	}
	__antithesis_instrumentation__.Notify(181233)
	defer sess.Close()

	cmd := cockroachNodeBinary(c, node) + " version"
	out, err := sess.CombinedOutput(ctx, cmd+" --build-tag")
	if err != nil {
		__antithesis_instrumentation__.Notify(181237)
		return nil, errors.Wrapf(err, "~ %s --build-tag\n%s", cmd, out)
	} else {
		__antithesis_instrumentation__.Notify(181238)
	}
	__antithesis_instrumentation__.Notify(181234)

	verString := strings.TrimSpace(string(out))
	return version.Parse(verString)
}

func argExists(args []string, target string) int {
	__antithesis_instrumentation__.Notify(181239)
	for i, arg := range args {
		__antithesis_instrumentation__.Notify(181241)
		if arg == target || func() bool {
			__antithesis_instrumentation__.Notify(181242)
			return strings.HasPrefix(arg, target+"=") == true
		}() == true {
			__antithesis_instrumentation__.Notify(181243)
			return i
		} else {
			__antithesis_instrumentation__.Notify(181244)
		}
	}
	__antithesis_instrumentation__.Notify(181240)
	return -1
}

type StartOpts struct {
	Target     StartTarget
	Sequential bool
	ExtraArgs  []string

	NumFilesLimit int64

	SkipInit        bool
	StoreCount      int
	EncryptedStores bool

	TenantID int
	KVAddrs  string
}

type StartTarget int

const (
	StartDefault StartTarget = iota

	StartTenantSQL

	StartTenantProxy
)

func (st StartTarget) String() string {
	__antithesis_instrumentation__.Notify(181245)
	return [...]string{
		StartDefault:     "default",
		StartTenantSQL:   "tenant SQL",
		StartTenantProxy: "tenant proxy",
	}[st]
}

func (c *SyncedCluster) Start(ctx context.Context, l *logger.Logger, startOpts StartOpts) error {
	__antithesis_instrumentation__.Notify(181246)
	if startOpts.Target == StartTenantProxy {
		__antithesis_instrumentation__.Notify(181250)
		return fmt.Errorf("start tenant proxy not implemented")
	} else {
		__antithesis_instrumentation__.Notify(181251)
	}
	__antithesis_instrumentation__.Notify(181247)
	if err := c.distributeCerts(ctx, l); err != nil {
		__antithesis_instrumentation__.Notify(181252)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181253)
	}
	__antithesis_instrumentation__.Notify(181248)

	nodes := c.TargetNodes()
	var parallelism = 0
	if startOpts.Sequential {
		__antithesis_instrumentation__.Notify(181254)
		parallelism = 1
	} else {
		__antithesis_instrumentation__.Notify(181255)
	}
	__antithesis_instrumentation__.Notify(181249)

	l.Printf("%s: starting nodes", c.Name)
	return c.Parallel(l, "", len(nodes), parallelism, func(nodeIdx int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(181256)
		node := nodes[nodeIdx]
		vers, err := getCockroachVersion(ctx, c, node)
		if err != nil {
			__antithesis_instrumentation__.Notify(181265)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(181266)
		}
		__antithesis_instrumentation__.Notify(181257)

		if _, err := c.startNode(ctx, l, node, startOpts, vers); err != nil {
			__antithesis_instrumentation__.Notify(181267)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(181268)
		}
		__antithesis_instrumentation__.Notify(181258)

		if startOpts.Target != StartDefault {
			__antithesis_instrumentation__.Notify(181269)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(181270)
		}
		__antithesis_instrumentation__.Notify(181259)

		if node != 1 {
			__antithesis_instrumentation__.Notify(181271)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(181272)
		}
		__antithesis_instrumentation__.Notify(181260)

		if startOpts.SkipInit {
			__antithesis_instrumentation__.Notify(181273)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(181274)
		}
		__antithesis_instrumentation__.Notify(181261)

		shouldInit := !c.useStartSingleNode()
		if shouldInit {
			__antithesis_instrumentation__.Notify(181275)
			l.Printf("%s: initializing cluster", c.Name)
			initOut, err := c.initializeCluster(ctx, node)
			if err != nil {
				__antithesis_instrumentation__.Notify(181277)
				return nil, errors.WithDetail(err, "unable to initialize cluster")
			} else {
				__antithesis_instrumentation__.Notify(181278)
			}
			__antithesis_instrumentation__.Notify(181276)

			if initOut != "" {
				__antithesis_instrumentation__.Notify(181279)
				l.Printf(initOut)
			} else {
				__antithesis_instrumentation__.Notify(181280)
			}
		} else {
			__antithesis_instrumentation__.Notify(181281)
		}
		__antithesis_instrumentation__.Notify(181262)

		l.Printf("%s: setting cluster settings", c.Name)
		clusterSettingsOut, err := c.setClusterSettings(ctx, l, node)
		if err != nil {
			__antithesis_instrumentation__.Notify(181282)
			return nil, errors.Wrap(err, "unable to set cluster settings")
		} else {
			__antithesis_instrumentation__.Notify(181283)
		}
		__antithesis_instrumentation__.Notify(181263)
		if clusterSettingsOut != "" {
			__antithesis_instrumentation__.Notify(181284)
			l.Printf(clusterSettingsOut)
		} else {
			__antithesis_instrumentation__.Notify(181285)
		}
		__antithesis_instrumentation__.Notify(181264)
		return nil, nil
	})
}

func (c *SyncedCluster) NodeDir(node Node, storeIndex int) string {
	__antithesis_instrumentation__.Notify(181286)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(181288)
		if storeIndex != 1 {
			__antithesis_instrumentation__.Notify(181290)
			panic("NodeDir only supports one store for local deployments")
		} else {
			__antithesis_instrumentation__.Notify(181291)
		}
		__antithesis_instrumentation__.Notify(181289)
		return filepath.Join(c.localVMDir(node), "data")
	} else {
		__antithesis_instrumentation__.Notify(181292)
	}
	__antithesis_instrumentation__.Notify(181287)
	return fmt.Sprintf("/mnt/data%d/cockroach", storeIndex)
}

func (c *SyncedCluster) LogDir(node Node) string {
	__antithesis_instrumentation__.Notify(181293)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(181295)
		return filepath.Join(c.localVMDir(node), "logs")
	} else {
		__antithesis_instrumentation__.Notify(181296)
	}
	__antithesis_instrumentation__.Notify(181294)
	return "logs"
}

func (c *SyncedCluster) CertsDir(node Node) string {
	__antithesis_instrumentation__.Notify(181297)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(181299)
		return filepath.Join(c.localVMDir(node), "certs")
	} else {
		__antithesis_instrumentation__.Notify(181300)
	}
	__antithesis_instrumentation__.Notify(181298)
	return "certs"
}

func (c *SyncedCluster) NodeURL(host string, port int) string {
	__antithesis_instrumentation__.Notify(181301)
	var u url.URL
	u.User = url.User("root")
	u.Scheme = "postgres"
	u.Host = fmt.Sprintf("%s:%d", host, port)
	v := url.Values{}
	if c.Secure {
		__antithesis_instrumentation__.Notify(181303)
		v.Add("sslcert", c.PGUrlCertsDir+"/client.root.crt")
		v.Add("sslkey", c.PGUrlCertsDir+"/client.root.key")
		v.Add("sslrootcert", c.PGUrlCertsDir+"/ca.crt")
		v.Add("sslmode", "verify-full")
	} else {
		__antithesis_instrumentation__.Notify(181304)
		v.Add("sslmode", "disable")
	}
	__antithesis_instrumentation__.Notify(181302)
	u.RawQuery = v.Encode()
	return "'" + u.String() + "'"
}

func (c *SyncedCluster) NodePort(node Node) int {
	__antithesis_instrumentation__.Notify(181305)
	return c.VMs[node-1].SQLPort
}

func (c *SyncedCluster) NodeUIPort(node Node) int {
	__antithesis_instrumentation__.Notify(181306)
	return c.VMs[node-1].AdminUIPort
}

func (c *SyncedCluster) SQL(ctx context.Context, l *logger.Logger, args []string) error {
	__antithesis_instrumentation__.Notify(181307)
	if len(args) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(181309)
		return len(c.Nodes) == 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(181310)

		if len(args) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(181312)
			return len(c.Nodes) != 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(181313)
			return fmt.Errorf("invalid number of nodes for interactive sql: %d", len(c.Nodes))
		} else {
			__antithesis_instrumentation__.Notify(181314)
		}
		__antithesis_instrumentation__.Notify(181311)
		url := c.NodeURL("localhost", c.NodePort(c.Nodes[0]))
		binary := cockroachNodeBinary(c, c.Nodes[0])
		allArgs := []string{binary, "sql", "--url", url}
		allArgs = append(allArgs, ssh.Escape(args))
		return c.SSH(ctx, l, []string{"-t"}, allArgs)
	} else {
		__antithesis_instrumentation__.Notify(181315)
	}
	__antithesis_instrumentation__.Notify(181308)

	return c.RunSQL(ctx, l, args)
}

func (c *SyncedCluster) RunSQL(ctx context.Context, l *logger.Logger, args []string) error {
	__antithesis_instrumentation__.Notify(181316)
	type result struct {
		node   Node
		output string
	}
	resultChan := make(chan result, len(c.Nodes))

	display := fmt.Sprintf("%s: executing sql", c.Name)
	if err := c.Parallel(l, display, len(c.Nodes), 0, func(nodeIdx int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(181321)
		node := c.Nodes[nodeIdx]
		sess, err := c.newSession(node)
		if err != nil {
			__antithesis_instrumentation__.Notify(181325)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(181326)
		}
		__antithesis_instrumentation__.Notify(181322)
		defer sess.Close()

		var cmd string
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(181327)
			cmd = fmt.Sprintf(`cd %s ; `, c.localVMDir(node))
		} else {
			__antithesis_instrumentation__.Notify(181328)
		}
		__antithesis_instrumentation__.Notify(181323)
		cmd += cockroachNodeBinary(c, node) + " sql --url " +
			c.NodeURL("localhost", c.NodePort(node)) + " " +
			ssh.Escape(args)

		out, err := sess.CombinedOutput(ctx, cmd)
		if err != nil {
			__antithesis_instrumentation__.Notify(181329)
			return nil, errors.Wrapf(err, "~ %s\n%s", cmd, out)
		} else {
			__antithesis_instrumentation__.Notify(181330)
		}
		__antithesis_instrumentation__.Notify(181324)

		resultChan <- result{node: node, output: string(out)}
		return nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(181331)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181332)
	}
	__antithesis_instrumentation__.Notify(181317)

	results := make([]result, 0, len(c.Nodes))
	for range c.Nodes {
		__antithesis_instrumentation__.Notify(181333)
		results = append(results, <-resultChan)
	}
	__antithesis_instrumentation__.Notify(181318)
	sort.Slice(results, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(181334)
		return results[i].node < results[j].node
	})
	__antithesis_instrumentation__.Notify(181319)
	for _, r := range results {
		__antithesis_instrumentation__.Notify(181335)
		l.Printf("node %d:\n%s", r.node, r.output)
	}
	__antithesis_instrumentation__.Notify(181320)

	return nil
}

func (c *SyncedCluster) startNode(
	ctx context.Context, l *logger.Logger, node Node, startOpts StartOpts, vers *version.Version,
) (string, error) {
	__antithesis_instrumentation__.Notify(181336)
	startCmd, err := c.generateStartCmd(ctx, l, node, startOpts, vers)
	if err != nil {
		__antithesis_instrumentation__.Notify(181342)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(181343)
	}
	__antithesis_instrumentation__.Notify(181337)

	if err := func() error {
		__antithesis_instrumentation__.Notify(181344)
		sess, err := c.newSession(node)
		if err != nil {
			__antithesis_instrumentation__.Notify(181348)
			return err
		} else {
			__antithesis_instrumentation__.Notify(181349)
		}
		__antithesis_instrumentation__.Notify(181345)
		defer sess.Close()

		sess.SetStdin(strings.NewReader(startCmd))
		var cmd string
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(181350)
			cmd = fmt.Sprintf(`cd %s ; `, c.localVMDir(node))
		} else {
			__antithesis_instrumentation__.Notify(181351)
		}
		__antithesis_instrumentation__.Notify(181346)
		cmd += `cat > cockroach.sh && chmod +x cockroach.sh`
		if out, err := sess.CombinedOutput(ctx, cmd); err != nil {
			__antithesis_instrumentation__.Notify(181352)
			return errors.Wrapf(err, "failed to upload start script: %s", out)
		} else {
			__antithesis_instrumentation__.Notify(181353)
		}
		__antithesis_instrumentation__.Notify(181347)

		return nil
	}(); err != nil {
		__antithesis_instrumentation__.Notify(181354)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(181355)
	}
	__antithesis_instrumentation__.Notify(181338)

	sess, err := c.newSession(node)
	if err != nil {
		__antithesis_instrumentation__.Notify(181356)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(181357)
	}
	__antithesis_instrumentation__.Notify(181339)
	defer sess.Close()

	var cmd string
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(181358)
		cmd = fmt.Sprintf(`cd %s ; `, c.localVMDir(node))
	} else {
		__antithesis_instrumentation__.Notify(181359)
	}
	__antithesis_instrumentation__.Notify(181340)
	cmd += "./cockroach.sh"
	out, err := sess.CombinedOutput(ctx, cmd)
	if err != nil {
		__antithesis_instrumentation__.Notify(181360)
		return "", errors.Wrapf(err, "~ %s\n%s", cmd, out)
	} else {
		__antithesis_instrumentation__.Notify(181361)
	}
	__antithesis_instrumentation__.Notify(181341)
	return strings.TrimSpace(string(out)), nil
}

func (c *SyncedCluster) generateStartCmd(
	ctx context.Context, l *logger.Logger, node Node, startOpts StartOpts, vers *version.Version,
) (string, error) {
	__antithesis_instrumentation__.Notify(181362)
	args, err := c.generateStartArgs(ctx, l, node, startOpts, vers)
	if err != nil {
		__antithesis_instrumentation__.Notify(181364)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(181365)
	}
	__antithesis_instrumentation__.Notify(181363)

	return execStartTemplate(startTemplateData{
		LogDir: c.LogDir(node),
		KeyCmd: c.generateKeyCmd(node, startOpts),
		EnvVars: append(append([]string{
			fmt.Sprintf("ROACHPROD=%s", c.roachprodEnvValue(node)),
			"GOTRACEBACK=crash",
			"COCKROACH_SKIP_ENABLING_DIAGNOSTIC_REPORTING=1",
		}, c.Env...), getEnvVars()...),
		Binary:        cockroachNodeBinary(c, node),
		Args:          args,
		MemoryMax:     config.MemoryMax,
		NumFilesLimit: startOpts.NumFilesLimit,
		Local:         c.IsLocal(),
	})
}

type startTemplateData struct {
	Local         bool
	LogDir        string
	Binary        string
	KeyCmd        string
	MemoryMax     string
	NumFilesLimit int64
	Args          []string
	EnvVars       []string
}

func execStartTemplate(data startTemplateData) (string, error) {
	__antithesis_instrumentation__.Notify(181366)
	tpl, err := template.New("start").
		Funcs(template.FuncMap{"shesc": func(i interface{}) string {
			__antithesis_instrumentation__.Notify(181370)
			return shellescape.Quote(fmt.Sprint(i))
		}}).
		Delims("#{", "#}").
		Parse(startScript)
	__antithesis_instrumentation__.Notify(181367)
	if err != nil {
		__antithesis_instrumentation__.Notify(181371)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(181372)
	}
	__antithesis_instrumentation__.Notify(181368)
	var buf strings.Builder
	if err := tpl.Execute(&buf, data); err != nil {
		__antithesis_instrumentation__.Notify(181373)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(181374)
	}
	__antithesis_instrumentation__.Notify(181369)
	return buf.String(), nil
}

func (c *SyncedCluster) generateStartArgs(
	ctx context.Context, l *logger.Logger, node Node, startOpts StartOpts, vers *version.Version,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(181375)
	var args []string

	switch startOpts.Target {
	case StartDefault:
		__antithesis_instrumentation__.Notify(181387)
		if c.useStartSingleNode() {
			__antithesis_instrumentation__.Notify(181391)
			args = []string{"start-single-node"}
		} else {
			__antithesis_instrumentation__.Notify(181392)
			args = []string{"start"}
		}

	case StartTenantSQL:
		__antithesis_instrumentation__.Notify(181388)
		args = []string{"mt", "start-sql"}

	case StartTenantProxy:
		__antithesis_instrumentation__.Notify(181389)

		panic("unimplemented")

	default:
		__antithesis_instrumentation__.Notify(181390)
		return nil, errors.Errorf("unsupported start target %v", startOpts.Target)
	}
	__antithesis_instrumentation__.Notify(181376)

	if c.Secure {
		__antithesis_instrumentation__.Notify(181393)
		args = append(args, `--certs-dir`, c.CertsDir(node))
	} else {
		__antithesis_instrumentation__.Notify(181394)
		args = append(args, "--insecure")
	}
	__antithesis_instrumentation__.Notify(181377)

	logDir := c.LogDir(node)
	idx1 := argExists(startOpts.ExtraArgs, "--log")
	idx2 := argExists(startOpts.ExtraArgs, "--log-config-file")

	if idx1 == -1 && func() bool {
		__antithesis_instrumentation__.Notify(181395)
		return idx2 == -1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(181396)
		if vers.AtLeast(version.MustParse("v21.1.0-alpha.0")) {
			__antithesis_instrumentation__.Notify(181397)

			args = append(args, "--log", `file-defaults: {dir: '`+logDir+`', exit-on-error: false}`)
		} else {
			__antithesis_instrumentation__.Notify(181398)
			args = append(args, `--log-dir`, logDir)
		}
	} else {
		__antithesis_instrumentation__.Notify(181399)
	}
	__antithesis_instrumentation__.Notify(181378)

	listenHost := ""
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(181400)

		listenHost = "127.0.0.1"
	} else {
		__antithesis_instrumentation__.Notify(181401)
	}
	__antithesis_instrumentation__.Notify(181379)

	if startOpts.Target == StartTenantSQL {
		__antithesis_instrumentation__.Notify(181402)
		args = append(args, fmt.Sprintf("--sql-addr=%s:%d", listenHost, c.NodePort(node)))
	} else {
		__antithesis_instrumentation__.Notify(181403)
		args = append(args, fmt.Sprintf("--listen-addr=%s:%d", listenHost, c.NodePort(node)))
	}
	__antithesis_instrumentation__.Notify(181380)
	args = append(args, fmt.Sprintf("--http-addr=%s:%d", listenHost, c.NodeUIPort(node)))

	if !c.IsLocal() {
		__antithesis_instrumentation__.Notify(181404)
		advertiseHost := ""
		if c.shouldAdvertisePublicIP() {
			__antithesis_instrumentation__.Notify(181406)
			advertiseHost = c.Host(node)
		} else {
			__antithesis_instrumentation__.Notify(181407)
			advertiseHost = c.VMs[node-1].PrivateIP
		}
		__antithesis_instrumentation__.Notify(181405)
		args = append(args,
			fmt.Sprintf("--advertise-addr=%s:%d", advertiseHost, c.NodePort(node)),
		)
	} else {
		__antithesis_instrumentation__.Notify(181408)
	}
	__antithesis_instrumentation__.Notify(181381)

	if startOpts.Target == StartDefault && func() bool {
		__antithesis_instrumentation__.Notify(181409)
		return !c.useStartSingleNode() == true
	}() == true {
		__antithesis_instrumentation__.Notify(181410)
		args = append(args, fmt.Sprintf("--join=%s:%d", c.Host(1), c.NodePort(1)))
	} else {
		__antithesis_instrumentation__.Notify(181411)
	}
	__antithesis_instrumentation__.Notify(181382)
	if startOpts.Target == StartTenantSQL {
		__antithesis_instrumentation__.Notify(181412)
		args = append(args, fmt.Sprintf("--kv-addrs=%s", startOpts.KVAddrs))
		args = append(args, fmt.Sprintf("--tenant-id=%d", startOpts.TenantID))
	} else {
		__antithesis_instrumentation__.Notify(181413)
	}
	__antithesis_instrumentation__.Notify(181383)

	if startOpts.Target == StartDefault {
		__antithesis_instrumentation__.Notify(181414)
		args = append(args, c.generateStartFlagsKV(node, startOpts, vers)...)
	} else {
		__antithesis_instrumentation__.Notify(181415)
	}
	__antithesis_instrumentation__.Notify(181384)

	if startOpts.Target == StartDefault || func() bool {
		__antithesis_instrumentation__.Notify(181416)
		return startOpts.Target == StartTenantSQL == true
	}() == true {
		__antithesis_instrumentation__.Notify(181417)
		args = append(args, c.generateStartFlagsSQL(node, startOpts, vers)...)
	} else {
		__antithesis_instrumentation__.Notify(181418)
	}
	__antithesis_instrumentation__.Notify(181385)

	e := expander{
		node: node,
	}
	for _, arg := range startOpts.ExtraArgs {
		__antithesis_instrumentation__.Notify(181419)
		expandedArg, err := e.expand(ctx, l, c, arg)
		if err != nil {
			__antithesis_instrumentation__.Notify(181421)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(181422)
		}
		__antithesis_instrumentation__.Notify(181420)
		args = append(args, strings.Split(expandedArg, " ")...)
	}
	__antithesis_instrumentation__.Notify(181386)

	return args, nil
}

func (c *SyncedCluster) generateStartFlagsKV(
	node Node, startOpts StartOpts, vers *version.Version,
) []string {
	__antithesis_instrumentation__.Notify(181423)
	var args []string
	var storeDirs []string
	if idx := argExists(startOpts.ExtraArgs, "--store"); idx == -1 {
		__antithesis_instrumentation__.Notify(181427)
		for i := 1; i <= startOpts.StoreCount; i++ {
			__antithesis_instrumentation__.Notify(181428)
			storeDir := c.NodeDir(node, i)
			storeDirs = append(storeDirs, storeDir)

			args = append(args, `--store`,
				`path=`+storeDir+`,attrs=`+fmt.Sprintf("store%d", i))
		}
	} else {
		__antithesis_instrumentation__.Notify(181429)
		storeDir := strings.TrimPrefix(startOpts.ExtraArgs[idx], "--store=")
		storeDirs = append(storeDirs, storeDir)
	}
	__antithesis_instrumentation__.Notify(181424)

	if startOpts.EncryptedStores {
		__antithesis_instrumentation__.Notify(181430)

		for _, storeDir := range storeDirs {
			__antithesis_instrumentation__.Notify(181431)

			encryptArgs := "path=%s,key=%s/aes-128.key,old-key=plain"
			encryptArgs = fmt.Sprintf(encryptArgs, storeDir, storeDir)
			args = append(args, `--enterprise-encryption`, encryptArgs)
		}
	} else {
		__antithesis_instrumentation__.Notify(181432)
	}
	__antithesis_instrumentation__.Notify(181425)

	args = append(args, fmt.Sprintf("--cache=%d%%", c.maybeScaleMem(25)))

	if locality := c.locality(node); locality != "" {
		__antithesis_instrumentation__.Notify(181433)
		if idx := argExists(startOpts.ExtraArgs, "--locality"); idx == -1 {
			__antithesis_instrumentation__.Notify(181434)
			args = append(args, "--locality="+locality)
		} else {
			__antithesis_instrumentation__.Notify(181435)
		}
	} else {
		__antithesis_instrumentation__.Notify(181436)
	}
	__antithesis_instrumentation__.Notify(181426)
	return args
}

func (c *SyncedCluster) generateStartFlagsSQL(
	node Node, startOpts StartOpts, vers *version.Version,
) []string {
	__antithesis_instrumentation__.Notify(181437)
	var args []string
	args = append(args, fmt.Sprintf("--max-sql-memory=%d%%", c.maybeScaleMem(25)))
	return args
}

func (c *SyncedCluster) maybeScaleMem(val int) int {
	__antithesis_instrumentation__.Notify(181438)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(181440)
		val /= len(c.Nodes)
		if val == 0 {
			__antithesis_instrumentation__.Notify(181441)
			val = 1
		} else {
			__antithesis_instrumentation__.Notify(181442)
		}
	} else {
		__antithesis_instrumentation__.Notify(181443)
	}
	__antithesis_instrumentation__.Notify(181439)
	return val
}

func (c *SyncedCluster) initializeCluster(ctx context.Context, node Node) (string, error) {
	__antithesis_instrumentation__.Notify(181444)
	initCmd := c.generateInitCmd(node)

	sess, err := c.newSession(node)
	if err != nil {
		__antithesis_instrumentation__.Notify(181447)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(181448)
	}
	__antithesis_instrumentation__.Notify(181445)
	defer sess.Close()

	out, err := sess.CombinedOutput(ctx, initCmd)
	if err != nil {
		__antithesis_instrumentation__.Notify(181449)
		return "", errors.Wrapf(err, "~ %s\n%s", initCmd, out)
	} else {
		__antithesis_instrumentation__.Notify(181450)
	}
	__antithesis_instrumentation__.Notify(181446)
	return strings.TrimSpace(string(out)), nil
}

func (c *SyncedCluster) setClusterSettings(
	ctx context.Context, l *logger.Logger, node Node,
) (string, error) {
	__antithesis_instrumentation__.Notify(181451)
	clusterSettingCmd := c.generateClusterSettingCmd(l, node)

	sess, err := c.newSession(node)
	if err != nil {
		__antithesis_instrumentation__.Notify(181454)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(181455)
	}
	__antithesis_instrumentation__.Notify(181452)
	defer sess.Close()

	out, err := sess.CombinedOutput(ctx, clusterSettingCmd)
	if err != nil {
		__antithesis_instrumentation__.Notify(181456)
		return "", errors.Wrapf(err, "~ %s\n%s", clusterSettingCmd, out)
	} else {
		__antithesis_instrumentation__.Notify(181457)
	}
	__antithesis_instrumentation__.Notify(181453)
	return strings.TrimSpace(string(out)), nil
}

func (c *SyncedCluster) generateClusterSettingCmd(l *logger.Logger, node Node) string {
	__antithesis_instrumentation__.Notify(181458)
	if config.CockroachDevLicense == "" {
		__antithesis_instrumentation__.Notify(181461)
		l.Printf("%s: COCKROACH_DEV_LICENSE unset: enterprise features will be unavailable\n",
			c.Name)
	} else {
		__antithesis_instrumentation__.Notify(181462)
	}
	__antithesis_instrumentation__.Notify(181459)

	var clusterSettingCmd string
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(181463)
		clusterSettingCmd = fmt.Sprintf(`cd %s ; `, c.localVMDir(1))
	} else {
		__antithesis_instrumentation__.Notify(181464)
	}
	__antithesis_instrumentation__.Notify(181460)

	binary := cockroachNodeBinary(c, node)
	path := fmt.Sprintf("%s/%s", c.NodeDir(node, 1), "settings-initialized")
	url := c.NodeURL("localhost", c.NodePort(1))

	clusterSettingCmd += fmt.Sprintf(`
		if ! test -e %s ; then
			COCKROACH_CONNECT_TIMEOUT=0 %s sql --url %s -e "SET CLUSTER SETTING server.remote_debugging.mode = 'any'" || true;
			COCKROACH_CONNECT_TIMEOUT=0 %s sql --url %s -e "
				SET CLUSTER SETTING cluster.organization = 'Cockroach Labs - Production Testing';
				SET CLUSTER SETTING enterprise.license = '%s';" \
			&& touch %s
		fi`, path, binary, url, binary, url, config.CockroachDevLicense, path)
	return clusterSettingCmd
}

func (c *SyncedCluster) generateInitCmd(node Node) string {
	__antithesis_instrumentation__.Notify(181465)
	var initCmd string
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(181467)
		initCmd = fmt.Sprintf(`cd %s ; `, c.localVMDir(1))
	} else {
		__antithesis_instrumentation__.Notify(181468)
	}
	__antithesis_instrumentation__.Notify(181466)

	path := fmt.Sprintf("%s/%s", c.NodeDir(node, 1), "cluster-bootstrapped")
	url := c.NodeURL("localhost", c.NodePort(node))
	binary := cockroachNodeBinary(c, node)
	initCmd += fmt.Sprintf(`
		if ! test -e %[1]s ; then
			COCKROACH_CONNECT_TIMEOUT=0 %[2]s init --url %[3]s && touch %[1]s
		fi`, path, binary, url)
	return initCmd
}

func (c *SyncedCluster) generateKeyCmd(node Node, startOpts StartOpts) string {
	__antithesis_instrumentation__.Notify(181469)
	if !startOpts.EncryptedStores {
		__antithesis_instrumentation__.Notify(181473)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(181474)
	}
	__antithesis_instrumentation__.Notify(181470)

	var storeDirs []string
	if storeArgIdx := argExists(startOpts.ExtraArgs, "--store"); storeArgIdx == -1 {
		__antithesis_instrumentation__.Notify(181475)
		for i := 1; i <= startOpts.StoreCount; i++ {
			__antithesis_instrumentation__.Notify(181476)
			storeDir := c.NodeDir(node, i)
			storeDirs = append(storeDirs, storeDir)
		}
	} else {
		__antithesis_instrumentation__.Notify(181477)
		storeDir := strings.TrimPrefix(startOpts.ExtraArgs[storeArgIdx], "--store=")
		storeDirs = append(storeDirs, storeDir)
	}
	__antithesis_instrumentation__.Notify(181471)

	var keyCmd strings.Builder
	for _, storeDir := range storeDirs {
		__antithesis_instrumentation__.Notify(181478)
		fmt.Fprintf(&keyCmd, `
			mkdir -p %[1]s;
			if [ ! -e %[1]s/aes-128.key ]; then
				openssl rand -out %[1]s/aes-128.key 48;
			fi;`, storeDir)
	}
	__antithesis_instrumentation__.Notify(181472)
	return keyCmd.String()
}

func (c *SyncedCluster) useStartSingleNode() bool {
	__antithesis_instrumentation__.Notify(181479)
	return len(c.VMs) == 1
}

func (c *SyncedCluster) distributeCerts(ctx context.Context, l *logger.Logger) error {
	__antithesis_instrumentation__.Notify(181480)
	for _, node := range c.TargetNodes() {
		__antithesis_instrumentation__.Notify(181482)
		if node == 1 && func() bool {
			__antithesis_instrumentation__.Notify(181483)
			return c.Secure == true
		}() == true {
			__antithesis_instrumentation__.Notify(181484)
			return c.DistributeCerts(ctx, l)
		} else {
			__antithesis_instrumentation__.Notify(181485)
		}
	}
	__antithesis_instrumentation__.Notify(181481)
	return nil
}

func (c *SyncedCluster) shouldAdvertisePublicIP() bool {
	__antithesis_instrumentation__.Notify(181486)

	for i := range c.VMs {
		__antithesis_instrumentation__.Notify(181488)
		if i > 0 && func() bool {
			__antithesis_instrumentation__.Notify(181489)
			return c.VMs[i].VPC != c.VMs[0].VPC == true
		}() == true {
			__antithesis_instrumentation__.Notify(181490)
			return true
		} else {
			__antithesis_instrumentation__.Notify(181491)
		}
	}
	__antithesis_instrumentation__.Notify(181487)
	return false
}

func getEnvVars() []string {
	__antithesis_instrumentation__.Notify(181492)
	var sl []string
	for _, v := range os.Environ() {
		__antithesis_instrumentation__.Notify(181494)
		if strings.HasPrefix(v, "COCKROACH_") {
			__antithesis_instrumentation__.Notify(181495)
			sl = append(sl, v)
		} else {
			__antithesis_instrumentation__.Notify(181496)
		}
	}
	__antithesis_instrumentation__.Notify(181493)
	return sl
}
