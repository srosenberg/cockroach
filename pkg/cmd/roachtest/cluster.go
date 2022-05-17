package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/armon/circbuf"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	_ "github.com/lib/pq"
	"github.com/spf13/pflag"
)

func init() {
	_ = roachprod.InitProviders()
}

var (
	local bool

	cockroach                 string
	libraryFilePaths          []string
	cloud                                  = spec.GCE
	encrypt                   encryptValue = "false"
	instanceType              string
	localSSDArg               bool
	workload                  string
	deprecatedRoachprodBinary string

	overrideOpts vm.CreateOpts

	overrideFlagset  *pflag.FlagSet
	overrideNumNodes = -1
	buildTag         string
	clusterName      string
	clusterWipe      bool
	zonesF           string
	teamCity         bool
	disableIssue     bool
)

type encryptValue string

func (v *encryptValue) String() string {
	__antithesis_instrumentation__.Notify(43240)
	return string(*v)
}

func (v *encryptValue) Set(s string) error {
	__antithesis_instrumentation__.Notify(43241)
	if s == "random" {
		__antithesis_instrumentation__.Notify(43244)
		*v = encryptValue(s)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(43245)
	}
	__antithesis_instrumentation__.Notify(43242)
	t, err := strconv.ParseBool(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(43246)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43247)
	}
	__antithesis_instrumentation__.Notify(43243)
	*v = encryptValue(fmt.Sprint(t))
	return nil
}

func (v *encryptValue) asBool() bool {
	__antithesis_instrumentation__.Notify(43248)
	if *v == "random" {
		__antithesis_instrumentation__.Notify(43251)
		return rand.Intn(2) == 0
	} else {
		__antithesis_instrumentation__.Notify(43252)
	}
	__antithesis_instrumentation__.Notify(43249)
	t, err := strconv.ParseBool(string(*v))
	if err != nil {
		__antithesis_instrumentation__.Notify(43253)
		return false
	} else {
		__antithesis_instrumentation__.Notify(43254)
	}
	__antithesis_instrumentation__.Notify(43250)
	return t
}

func (v *encryptValue) Type() string {
	__antithesis_instrumentation__.Notify(43255)
	return "string"
}

type errBinaryOrLibraryNotFound struct {
	binary string
}

func (e errBinaryOrLibraryNotFound) Error() string {
	__antithesis_instrumentation__.Notify(43256)
	return fmt.Sprintf("binary or library %q not found (or was not executable)", e.binary)
}

func filepathAbs(path string) (string, error) {
	__antithesis_instrumentation__.Notify(43257)
	path, err := filepath.Abs(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(43259)
		return "", errors.WithStack(err)
	} else {
		__antithesis_instrumentation__.Notify(43260)
	}
	__antithesis_instrumentation__.Notify(43258)
	return path, nil
}

func findBinary(binary, defValue string) (abspath string, err error) {
	__antithesis_instrumentation__.Notify(43261)
	if binary == "" {
		__antithesis_instrumentation__.Notify(43264)
		binary = defValue
	} else {
		__antithesis_instrumentation__.Notify(43265)
	}
	__antithesis_instrumentation__.Notify(43262)

	if fi, err := os.Stat(binary); err == nil && func() bool {
		__antithesis_instrumentation__.Notify(43266)
		return fi.Mode().IsRegular() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(43267)
		return (fi.Mode() & 0111) != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(43268)
		return filepathAbs(binary)
	} else {
		__antithesis_instrumentation__.Notify(43269)
	}
	__antithesis_instrumentation__.Notify(43263)
	return findBinaryOrLibrary("bin", binary)
}

func findLibrary(libraryName string) (string, error) {
	__antithesis_instrumentation__.Notify(43270)
	suffix := ".so"
	if local {
		__antithesis_instrumentation__.Notify(43272)
		switch runtime.GOOS {
		case "linux":
			__antithesis_instrumentation__.Notify(43273)
		case "freebsd":
			__antithesis_instrumentation__.Notify(43274)
		case "openbsd":
			__antithesis_instrumentation__.Notify(43275)
		case "dragonfly":
			__antithesis_instrumentation__.Notify(43276)
		case "windows":
			__antithesis_instrumentation__.Notify(43277)
			suffix = ".dll"
		case "darwin":
			__antithesis_instrumentation__.Notify(43278)
			suffix = ".dylib"
		default:
			__antithesis_instrumentation__.Notify(43279)
			return "", errors.Newf("failed to find suffix for runtime %s", runtime.GOOS)
		}
	} else {
		__antithesis_instrumentation__.Notify(43280)
	}
	__antithesis_instrumentation__.Notify(43271)
	return findBinaryOrLibrary("lib", libraryName+suffix)
}

func findBinaryOrLibrary(binOrLib string, name string) (string, error) {
	__antithesis_instrumentation__.Notify(43281)

	path, err := exec.LookPath(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(43283)
		if strings.HasPrefix(name, "/") {
			__antithesis_instrumentation__.Notify(43288)
			return "", errors.WithStack(err)
		} else {
			__antithesis_instrumentation__.Notify(43289)
		}
		__antithesis_instrumentation__.Notify(43284)

		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			__antithesis_instrumentation__.Notify(43290)
			gopath = filepath.Join(os.Getenv("HOME"), "go")
		} else {
			__antithesis_instrumentation__.Notify(43291)
		}
		__antithesis_instrumentation__.Notify(43285)

		var suffix string
		if !local {
			__antithesis_instrumentation__.Notify(43292)
			suffix = ".docker_amd64"
		} else {
			__antithesis_instrumentation__.Notify(43293)
		}
		__antithesis_instrumentation__.Notify(43286)
		dirs := []string{
			filepath.Join(gopath, "/src/github.com/cockroachdb/cockroach/"),
			filepath.Join(gopath, "/src/github.com/cockroachdb/cockroach", binOrLib+suffix),
			filepath.Join(os.ExpandEnv("$PWD"), binOrLib+suffix),
			filepath.Join(gopath, "/src/github.com/cockroachdb/cockroach", binOrLib),
		}
		for _, dir := range dirs {
			__antithesis_instrumentation__.Notify(43294)
			path = filepath.Join(dir, name)
			var err2 error
			path, err2 = exec.LookPath(path)
			if err2 == nil {
				__antithesis_instrumentation__.Notify(43295)
				return filepathAbs(path)
			} else {
				__antithesis_instrumentation__.Notify(43296)
			}
		}
		__antithesis_instrumentation__.Notify(43287)
		return "", errBinaryOrLibraryNotFound{name}
	} else {
		__antithesis_instrumentation__.Notify(43297)
	}
	__antithesis_instrumentation__.Notify(43282)
	return filepathAbs(path)
}

func initBinariesAndLibraries() {
	__antithesis_instrumentation__.Notify(43298)

	if clusterName == "local" {
		__antithesis_instrumentation__.Notify(43304)
		local = true
	} else {
		__antithesis_instrumentation__.Notify(43305)
	}
	__antithesis_instrumentation__.Notify(43299)
	if local {
		__antithesis_instrumentation__.Notify(43306)
		cloud = spec.Local
	} else {
		__antithesis_instrumentation__.Notify(43307)
	}
	__antithesis_instrumentation__.Notify(43300)

	cockroachDefault := "cockroach"
	if !local {
		__antithesis_instrumentation__.Notify(43308)
		cockroachDefault = "cockroach-linux-2.6.32-gnu-amd64"
	} else {
		__antithesis_instrumentation__.Notify(43309)
	}
	__antithesis_instrumentation__.Notify(43301)
	var err error
	cockroach, err = findBinary(cockroach, cockroachDefault)
	if err != nil {
		__antithesis_instrumentation__.Notify(43310)
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(43311)
	}
	__antithesis_instrumentation__.Notify(43302)

	workload, err = findBinary(workload, "workload")
	if errors.As(err, &errBinaryOrLibraryNotFound{}) {
		__antithesis_instrumentation__.Notify(43312)
		fmt.Fprintln(os.Stderr, "workload binary not provided, proceeding anyway")
	} else {
		__antithesis_instrumentation__.Notify(43313)
		if err != nil {
			__antithesis_instrumentation__.Notify(43314)
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			os.Exit(1)
		} else {
			__antithesis_instrumentation__.Notify(43315)
		}
	}
	__antithesis_instrumentation__.Notify(43303)

	for _, libraryName := range []string{"libgeos", "libgeos_c"} {
		__antithesis_instrumentation__.Notify(43316)
		if libraryFilePath, err := findLibrary(libraryName); err != nil {
			__antithesis_instrumentation__.Notify(43317)
			fmt.Fprintf(os.Stderr, "error finding library %s, ignoring: %+v\n", libraryName, err)
		} else {
			__antithesis_instrumentation__.Notify(43318)
			libraryFilePaths = append(libraryFilePaths, libraryFilePath)
		}
	}
}

func execCmd(ctx context.Context, l *logger.Logger, clusterName string, args ...string) error {
	__antithesis_instrumentation__.Notify(43319)
	return execCmdEx(ctx, l, clusterName, args...).err
}

type cmdRes struct {
	err error

	stdout, stderr string
}

func execCmdEx(ctx context.Context, l *logger.Logger, clusterName string, args ...string) cmdRes {
	__antithesis_instrumentation__.Notify(43320)
	var cancel func()
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	l.Printf("> %s\n", strings.Join(args, " "))
	var roachprodRunStdout, roachprodRunStderr io.Writer

	debugStdoutBuffer, _ := circbuf.NewBuffer(4096)
	debugStderrBuffer, _ := circbuf.NewBuffer(4096)

	var closePipes func(ctx context.Context)
	var wg sync.WaitGroup
	{
		__antithesis_instrumentation__.Notify(43325)

		var wOut, wErr, rOut, rErr *os.File
		var cwOnce sync.Once
		closePipes = func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(43330)

			cwOnce.Do(func() {
				__antithesis_instrumentation__.Notify(43331)
				if wOut != nil {
					__antithesis_instrumentation__.Notify(43336)
					_ = wOut.Close()
				} else {
					__antithesis_instrumentation__.Notify(43337)
				}
				__antithesis_instrumentation__.Notify(43332)
				if wErr != nil {
					__antithesis_instrumentation__.Notify(43338)
					_ = wErr.Close()
				} else {
					__antithesis_instrumentation__.Notify(43339)
				}
				__antithesis_instrumentation__.Notify(43333)
				dur := 10 * time.Second
				if ctx.Err() != nil {
					__antithesis_instrumentation__.Notify(43340)
					dur = 10 * time.Millisecond
				} else {
					__antithesis_instrumentation__.Notify(43341)
				}
				__antithesis_instrumentation__.Notify(43334)
				deadline := timeutil.Now().Add(dur)
				if rOut != nil {
					__antithesis_instrumentation__.Notify(43342)
					_ = rOut.SetReadDeadline(deadline)
				} else {
					__antithesis_instrumentation__.Notify(43343)
				}
				__antithesis_instrumentation__.Notify(43335)
				if rErr != nil {
					__antithesis_instrumentation__.Notify(43344)
					_ = rErr.SetReadDeadline(deadline)
				} else {
					__antithesis_instrumentation__.Notify(43345)
				}
			})
		}
		__antithesis_instrumentation__.Notify(43326)
		defer closePipes(ctx)

		var err error
		rOut, wOut, err = os.Pipe()
		if err != nil {
			__antithesis_instrumentation__.Notify(43346)
			return cmdRes{err: err}
		} else {
			__antithesis_instrumentation__.Notify(43347)
		}
		__antithesis_instrumentation__.Notify(43327)

		rErr, wErr, err = os.Pipe()
		if err != nil {
			__antithesis_instrumentation__.Notify(43348)
			return cmdRes{err: err}
		} else {
			__antithesis_instrumentation__.Notify(43349)
		}
		__antithesis_instrumentation__.Notify(43328)
		roachprodRunStdout = wOut
		wg.Add(1)
		go func() {
			__antithesis_instrumentation__.Notify(43350)
			defer wg.Done()
			_, _ = io.Copy(l.Stdout, io.TeeReader(rOut, debugStdoutBuffer))
		}()
		__antithesis_instrumentation__.Notify(43329)

		if l.Stderr == l.Stdout {
			__antithesis_instrumentation__.Notify(43351)

			roachprodRunStderr = wOut
		} else {
			__antithesis_instrumentation__.Notify(43352)
			roachprodRunStderr = wErr
			wg.Add(1)
			go func() {
				__antithesis_instrumentation__.Notify(43353)
				defer wg.Done()
				_, _ = io.Copy(l.Stderr, io.TeeReader(rErr, debugStderrBuffer))
			}()
		}
	}
	__antithesis_instrumentation__.Notify(43321)

	err := roachprod.Run(ctx, l, clusterName, "", "", false, roachprodRunStdout, roachprodRunStderr, args)
	closePipes(ctx)
	wg.Wait()

	stdoutString := debugStdoutBuffer.String()
	if debugStdoutBuffer.TotalWritten() > debugStdoutBuffer.Size() {
		__antithesis_instrumentation__.Notify(43354)
		stdoutString = "<... some data truncated by circular buffer; go to artifacts for details ...>\n" + stdoutString
	} else {
		__antithesis_instrumentation__.Notify(43355)
	}
	__antithesis_instrumentation__.Notify(43322)
	stderrString := debugStderrBuffer.String()
	if debugStderrBuffer.TotalWritten() > debugStderrBuffer.Size() {
		__antithesis_instrumentation__.Notify(43356)
		stderrString = "<... some data truncated by circular buffer; go to artifacts for details ...>\n" + stderrString
	} else {
		__antithesis_instrumentation__.Notify(43357)
	}
	__antithesis_instrumentation__.Notify(43323)

	if err != nil {
		__antithesis_instrumentation__.Notify(43358)

		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(43360)
			err = errors.CombineErrors(ctx.Err(), err)
		} else {
			__antithesis_instrumentation__.Notify(43361)
		}
		__antithesis_instrumentation__.Notify(43359)

		if err != nil {
			__antithesis_instrumentation__.Notify(43362)
			err = &cluster.WithCommandDetails{
				Wrapped: err,
				Cmd:     strings.Join(args, " "),
				Stderr:  stderrString,
				Stdout:  stdoutString,
			}
		} else {
			__antithesis_instrumentation__.Notify(43363)
		}
	} else {
		__antithesis_instrumentation__.Notify(43364)
	}
	__antithesis_instrumentation__.Notify(43324)

	return cmdRes{
		err:    err,
		stdout: stdoutString,
		stderr: stderrString,
	}
}

type clusterRegistry struct {
	mu struct {
		syncutil.Mutex
		clusters map[string]*clusterImpl
		tagCount map[string]int

		savedClusters map[*clusterImpl]string
	}
}

func newClusterRegistry() *clusterRegistry {
	__antithesis_instrumentation__.Notify(43365)
	cr := &clusterRegistry{}
	cr.mu.clusters = make(map[string]*clusterImpl)
	cr.mu.savedClusters = make(map[*clusterImpl]string)
	return cr
}

func (r *clusterRegistry) registerCluster(c *clusterImpl) error {
	__antithesis_instrumentation__.Notify(43366)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.mu.clusters[c.name] != nil {
		__antithesis_instrumentation__.Notify(43368)
		return fmt.Errorf("cluster named %q already exists in registry", c.name)
	} else {
		__antithesis_instrumentation__.Notify(43369)
	}
	__antithesis_instrumentation__.Notify(43367)
	r.mu.clusters[c.name] = c
	return nil
}

func (r *clusterRegistry) unregisterCluster(c *clusterImpl) bool {
	__antithesis_instrumentation__.Notify(43370)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.mu.clusters[c.name]; !ok {
		__antithesis_instrumentation__.Notify(43373)

		return false
	} else {
		__antithesis_instrumentation__.Notify(43374)
	}
	__antithesis_instrumentation__.Notify(43371)
	delete(r.mu.clusters, c.name)
	if c.tag != "" {
		__antithesis_instrumentation__.Notify(43375)
		if _, ok := r.mu.tagCount[c.tag]; !ok {
			__antithesis_instrumentation__.Notify(43377)
			panic(fmt.Sprintf("tagged cluster not accounted for: %s", c))
		} else {
			__antithesis_instrumentation__.Notify(43378)
		}
		__antithesis_instrumentation__.Notify(43376)
		r.mu.tagCount[c.tag]--
	} else {
		__antithesis_instrumentation__.Notify(43379)
	}
	__antithesis_instrumentation__.Notify(43372)
	return true
}

func (r *clusterRegistry) countForTag(tag string) int {
	__antithesis_instrumentation__.Notify(43380)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.mu.tagCount[tag]
}

func (r *clusterRegistry) markClusterAsSaved(c *clusterImpl, msg string) {
	__antithesis_instrumentation__.Notify(43381)
	r.mu.Lock()
	r.mu.savedClusters[c] = msg
	r.mu.Unlock()
}

type clusterWithMsg struct {
	*clusterImpl
	savedMsg string
}

func (r *clusterRegistry) savedClusters() []clusterWithMsg {
	__antithesis_instrumentation__.Notify(43382)
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]clusterWithMsg, len(r.mu.savedClusters))
	i := 0
	for c, msg := range r.mu.savedClusters {
		__antithesis_instrumentation__.Notify(43385)
		res[i] = clusterWithMsg{
			clusterImpl: c,
			savedMsg:    msg,
		}
		i++
	}
	__antithesis_instrumentation__.Notify(43383)
	sort.Slice(res, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(43386)
		return strings.Compare(res[i].name, res[j].name) < 0
	})
	__antithesis_instrumentation__.Notify(43384)
	return res
}

func (r *clusterRegistry) destroyAllClusters(ctx context.Context, l *logger.Logger) {
	__antithesis_instrumentation__.Notify(43387)

	done := make(chan struct{})
	go func() {
		__antithesis_instrumentation__.Notify(43389)
		defer close(done)

		var clusters []*clusterImpl
		savedClusters := make(map[*clusterImpl]struct{})
		r.mu.Lock()
		for _, c := range r.mu.clusters {
			__antithesis_instrumentation__.Notify(43393)
			clusters = append(clusters, c)
		}
		__antithesis_instrumentation__.Notify(43390)
		for c := range r.mu.savedClusters {
			__antithesis_instrumentation__.Notify(43394)
			savedClusters[c] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(43391)
		r.mu.Unlock()

		var wg sync.WaitGroup
		wg.Add(len(clusters))
		for _, c := range clusters {
			__antithesis_instrumentation__.Notify(43395)
			go func(c *clusterImpl) {
				__antithesis_instrumentation__.Notify(43396)
				defer wg.Done()
				if _, ok := savedClusters[c]; !ok {
					__antithesis_instrumentation__.Notify(43397)

					c.Destroy(ctx, dontCloseLogger, l)
				} else {
					__antithesis_instrumentation__.Notify(43398)
				}
			}(c)
		}
		__antithesis_instrumentation__.Notify(43392)

		wg.Wait()
	}()
	__antithesis_instrumentation__.Notify(43388)

	select {
	case <-done:
		__antithesis_instrumentation__.Notify(43399)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(43400)
	}
}

func makeGCEClusterName(name string) string {
	__antithesis_instrumentation__.Notify(43401)
	name = strings.ToLower(name)
	name = regexp.MustCompile(`[^-a-z0-9]+`).ReplaceAllString(name, "-")
	name = regexp.MustCompile(`-+`).ReplaceAllString(name, "-")
	return name
}

func makeClusterName(name string) string {
	__antithesis_instrumentation__.Notify(43402)
	return makeGCEClusterName(name)
}

func MachineTypeToCPUs(s string) int {
	__antithesis_instrumentation__.Notify(43403)
	{
		__antithesis_instrumentation__.Notify(43407)

		var v int
		if _, err := fmt.Sscanf(s, "n1-standard-%d", &v); err == nil {
			__antithesis_instrumentation__.Notify(43410)
			return v
		} else {
			__antithesis_instrumentation__.Notify(43411)
		}
		__antithesis_instrumentation__.Notify(43408)
		if _, err := fmt.Sscanf(s, "n1-highcpu-%d", &v); err == nil {
			__antithesis_instrumentation__.Notify(43412)
			return v
		} else {
			__antithesis_instrumentation__.Notify(43413)
		}
		__antithesis_instrumentation__.Notify(43409)
		if _, err := fmt.Sscanf(s, "n1-highmem-%d", &v); err == nil {
			__antithesis_instrumentation__.Notify(43414)
			return v
		} else {
			__antithesis_instrumentation__.Notify(43415)
		}
	}
	__antithesis_instrumentation__.Notify(43404)

	typeAndSize := strings.Split(s, ".")

	if len(typeAndSize) == 2 {
		__antithesis_instrumentation__.Notify(43416)
		size := typeAndSize[1]

		switch size {
		case "large":
			__antithesis_instrumentation__.Notify(43417)
			return 2
		case "xlarge":
			__antithesis_instrumentation__.Notify(43418)
			return 4
		case "2xlarge":
			__antithesis_instrumentation__.Notify(43419)
			return 8
		case "4xlarge":
			__antithesis_instrumentation__.Notify(43420)
			return 16
		case "9xlarge":
			__antithesis_instrumentation__.Notify(43421)
			return 36
		case "12xlarge":
			__antithesis_instrumentation__.Notify(43422)
			return 48
		case "18xlarge":
			__antithesis_instrumentation__.Notify(43423)
			return 72
		case "24xlarge":
			__antithesis_instrumentation__.Notify(43424)
			return 96
		default:
			__antithesis_instrumentation__.Notify(43425)
		}
	} else {
		__antithesis_instrumentation__.Notify(43426)
	}
	__antithesis_instrumentation__.Notify(43405)

	switch s {
	case "Standard_D2_v3":
		__antithesis_instrumentation__.Notify(43427)
		return 2
	case "Standard_D4_v3":
		__antithesis_instrumentation__.Notify(43428)
		return 4
	case "Standard_D8_v3":
		__antithesis_instrumentation__.Notify(43429)
		return 8
	case "Standard_D16_v3":
		__antithesis_instrumentation__.Notify(43430)
		return 16
	case "Standard_D32_v3":
		__antithesis_instrumentation__.Notify(43431)
		return 32
	case "Standard_D48_v3":
		__antithesis_instrumentation__.Notify(43432)
		return 48
	case "Standard_D64_v3":
		__antithesis_instrumentation__.Notify(43433)
		return 64
	default:
		__antithesis_instrumentation__.Notify(43434)
	}
	__antithesis_instrumentation__.Notify(43406)

	fmt.Fprintf(os.Stderr, "unknown machine type: %s\n", s)
	os.Exit(1)
	return -1
}

type nodeSelector interface {
	option.Option
	Merge(option.NodeListOption) option.NodeListOption
}

type clusterImpl struct {
	name string
	tag  string
	spec spec.ClusterSpec
	t    test.Test

	r *clusterRegistry

	l *logger.Logger

	localCertsDir string
	expiration    time.Time
	encAtRest     bool

	destroyState destroyState
}

func (c *clusterImpl) Name() string {
	__antithesis_instrumentation__.Notify(43435)
	return c.name
}

func (c *clusterImpl) Spec() spec.ClusterSpec {
	__antithesis_instrumentation__.Notify(43436)
	return c.spec
}

func (c *clusterImpl) status(args ...interface{}) {
	__antithesis_instrumentation__.Notify(43437)
	if c.t == nil {
		__antithesis_instrumentation__.Notify(43439)
		return
	} else {
		__antithesis_instrumentation__.Notify(43440)
	}
	__antithesis_instrumentation__.Notify(43438)
	c.t.Status(args...)
}

func (c *clusterImpl) workerStatus(args ...interface{}) {
	__antithesis_instrumentation__.Notify(43441)
	if impl, ok := c.t.(*testImpl); ok {
		__antithesis_instrumentation__.Notify(43442)
		impl.WorkerStatus(args...)
	} else {
		__antithesis_instrumentation__.Notify(43443)
	}
}

func (c *clusterImpl) String() string {
	__antithesis_instrumentation__.Notify(43444)
	return fmt.Sprintf("%s [tag:%s] (%d nodes)", c.name, c.tag, c.Spec().NodeCount)
}

type destroyState struct {
	owned bool

	alloc *quotapool.IntAlloc

	mu struct {
		syncutil.Mutex
		loggerClosed bool

		destroyed chan struct{}

		saved bool

		savedMsg string
	}
}

func (c *clusterImpl) closeLogger() {
	__antithesis_instrumentation__.Notify(43445)
	c.destroyState.mu.Lock()
	defer c.destroyState.mu.Unlock()
	if c.destroyState.mu.loggerClosed {
		__antithesis_instrumentation__.Notify(43447)
		return
	} else {
		__antithesis_instrumentation__.Notify(43448)
	}
	__antithesis_instrumentation__.Notify(43446)
	c.destroyState.mu.loggerClosed = true
	c.l.Close()
}

type clusterConfig struct {
	spec spec.ClusterSpec

	artifactsDir string

	username     string
	localCluster bool
	useIOBarrier bool
	alloc        *quotapool.IntAlloc
}

type clusterFactory struct {
	namePrefix string

	counter uint64

	r *clusterRegistry

	artifactsDir string

	sem chan struct{}
}

func newClusterFactory(
	user string, clustersID string, artifactsDir string, r *clusterRegistry, concurrentCreations int,
) *clusterFactory {
	__antithesis_instrumentation__.Notify(43449)
	secs := timeutil.Now().Unix()
	var prefix string
	if clustersID != "" {
		__antithesis_instrumentation__.Notify(43451)
		prefix = fmt.Sprintf("%s-%s-%d-", user, clustersID, secs)
	} else {
		__antithesis_instrumentation__.Notify(43452)
		prefix = fmt.Sprintf("%s-%d-", user, secs)
	}
	__antithesis_instrumentation__.Notify(43450)
	return &clusterFactory{
		sem:          make(chan struct{}, concurrentCreations),
		namePrefix:   prefix,
		artifactsDir: artifactsDir,
		r:            r,
	}
}

func (f *clusterFactory) acquireSem() func() {
	__antithesis_instrumentation__.Notify(43453)
	f.sem <- struct{}{}
	return f.releaseSem
}

func (f *clusterFactory) releaseSem() {
	__antithesis_instrumentation__.Notify(43454)
	<-f.sem
}

func (f *clusterFactory) genName(cfg clusterConfig) string {
	__antithesis_instrumentation__.Notify(43455)
	if cfg.localCluster {
		__antithesis_instrumentation__.Notify(43457)
		return "local"
	} else {
		__antithesis_instrumentation__.Notify(43458)
	}
	__antithesis_instrumentation__.Notify(43456)
	count := atomic.AddUint64(&f.counter, 1)
	return makeClusterName(
		fmt.Sprintf("%s-%02d-%s", f.namePrefix, count, cfg.spec.String()))
}

func createFlagsOverride(flags *pflag.FlagSet, opts *vm.CreateOpts) {
	__antithesis_instrumentation__.Notify(43459)
	if flags != nil {
		__antithesis_instrumentation__.Notify(43460)
		if flags.Changed("lifetime") {
			__antithesis_instrumentation__.Notify(43466)
			opts.Lifetime = overrideOpts.Lifetime
		} else {
			__antithesis_instrumentation__.Notify(43467)
		}
		__antithesis_instrumentation__.Notify(43461)
		if flags.Changed("roachprod-local-ssd") {
			__antithesis_instrumentation__.Notify(43468)
			opts.SSDOpts.UseLocalSSD = overrideOpts.SSDOpts.UseLocalSSD
		} else {
			__antithesis_instrumentation__.Notify(43469)
		}
		__antithesis_instrumentation__.Notify(43462)
		if flags.Changed("filesystem") {
			__antithesis_instrumentation__.Notify(43470)
			opts.SSDOpts.FileSystem = overrideOpts.SSDOpts.FileSystem
		} else {
			__antithesis_instrumentation__.Notify(43471)
		}
		__antithesis_instrumentation__.Notify(43463)
		if flags.Changed("local-ssd-no-ext4-barrier") {
			__antithesis_instrumentation__.Notify(43472)
			opts.SSDOpts.NoExt4Barrier = overrideOpts.SSDOpts.NoExt4Barrier
		} else {
			__antithesis_instrumentation__.Notify(43473)
		}
		__antithesis_instrumentation__.Notify(43464)
		if flags.Changed("os-volume-size") {
			__antithesis_instrumentation__.Notify(43474)
			opts.OsVolumeSize = overrideOpts.OsVolumeSize
		} else {
			__antithesis_instrumentation__.Notify(43475)
		}
		__antithesis_instrumentation__.Notify(43465)
		if flags.Changed("geo") {
			__antithesis_instrumentation__.Notify(43476)
			opts.GeoDistributed = overrideOpts.GeoDistributed
		} else {
			__antithesis_instrumentation__.Notify(43477)
		}
	} else {
		__antithesis_instrumentation__.Notify(43478)
	}
}

func (f *clusterFactory) clusterMock(cfg clusterConfig) *clusterImpl {
	__antithesis_instrumentation__.Notify(43479)
	return &clusterImpl{
		name:       f.genName(cfg),
		expiration: timeutil.Now().Add(24 * time.Hour),
		r:          f.r,
	}
}

func (f *clusterFactory) newCluster(
	ctx context.Context, cfg clusterConfig, setStatus func(string), teeOpt logger.TeeOptType,
) (*clusterImpl, error) {
	__antithesis_instrumentation__.Notify(43480)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43487)
		return nil, errors.Wrap(ctx.Err(), "newCluster")
	} else {
		__antithesis_instrumentation__.Notify(43488)
	}
	__antithesis_instrumentation__.Notify(43481)

	if overrideFlagset != nil && func() bool {
		__antithesis_instrumentation__.Notify(43489)
		return overrideFlagset.Changed("nodes") == true
	}() == true {
		__antithesis_instrumentation__.Notify(43490)
		cfg.spec.NodeCount = overrideNumNodes
	} else {
		__antithesis_instrumentation__.Notify(43491)
	}
	__antithesis_instrumentation__.Notify(43482)

	if cfg.spec.NodeCount == 0 {
		__antithesis_instrumentation__.Notify(43492)

		c := f.clusterMock(cfg)
		if err := f.r.registerCluster(c); err != nil {
			__antithesis_instrumentation__.Notify(43494)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(43495)
		}
		__antithesis_instrumentation__.Notify(43493)
		return c, nil
	} else {
		__antithesis_instrumentation__.Notify(43496)
	}
	__antithesis_instrumentation__.Notify(43483)

	if cfg.localCluster {
		__antithesis_instrumentation__.Notify(43497)

		cfg.spec.Lifetime = 100000 * time.Hour
	} else {
		__antithesis_instrumentation__.Notify(43498)
	}
	__antithesis_instrumentation__.Notify(43484)

	setStatus("acquiring cluster creation semaphore")
	release := f.acquireSem()
	defer release()

	setStatus("roachprod create")
	defer setStatus("idle")

	providerOptsContainer := vm.CreateProviderOptionsContainer()

	createVMOpts, providerOpts, err := cfg.spec.RoachprodOpts("", cfg.useIOBarrier)
	if err != nil {
		__antithesis_instrumentation__.Notify(43499)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(43500)
	}
	__antithesis_instrumentation__.Notify(43485)
	if cfg.spec.Cloud != spec.Local {
		__antithesis_instrumentation__.Notify(43501)
		providerOptsContainer.SetProviderOpts(cfg.spec.Cloud, providerOpts)
	} else {
		__antithesis_instrumentation__.Notify(43502)
	}
	__antithesis_instrumentation__.Notify(43486)

	createFlagsOverride(overrideFlagset, &createVMOpts)

	cfg.spec.Lifetime = createVMOpts.Lifetime

	const maxAttempts = 3

	for i := 1; ; i++ {
		__antithesis_instrumentation__.Notify(43503)
		c := &clusterImpl{

			name:       f.genName(cfg),
			spec:       cfg.spec,
			expiration: cfg.spec.Expiration(),
			r:          f.r,
			destroyState: destroyState{
				owned: true,
				alloc: cfg.alloc,
			},
		}
		c.status("creating cluster")

		var retryStr string
		if i > 1 {
			__antithesis_instrumentation__.Notify(43508)
			retryStr = "-retry" + strconv.Itoa(i-1)
		} else {
			__antithesis_instrumentation__.Notify(43509)
		}
		__antithesis_instrumentation__.Notify(43504)
		logPath := filepath.Join(f.artifactsDir, runnerLogsDir, "cluster-create", c.name+retryStr+".log")
		l, err := logger.RootLogger(logPath, teeOpt)
		if err != nil {
			__antithesis_instrumentation__.Notify(43510)
			log.Fatalf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(43511)
		}
		__antithesis_instrumentation__.Notify(43505)

		l.PrintfCtx(ctx, "Attempting cluster creation (attempt #%d/%d)", i, maxAttempts)
		createVMOpts.ClusterName = c.name
		err = roachprod.Create(ctx, l, cfg.username, cfg.spec.NodeCount, createVMOpts, providerOptsContainer)
		if err == nil {
			__antithesis_instrumentation__.Notify(43512)
			if err := f.r.registerCluster(c); err != nil {
				__antithesis_instrumentation__.Notify(43514)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(43515)
			}
			__antithesis_instrumentation__.Notify(43513)
			c.status("idle")
			l.Close()
			return c, nil
		} else {
			__antithesis_instrumentation__.Notify(43516)
		}
		__antithesis_instrumentation__.Notify(43506)

		if errors.HasType(err, (*roachprod.ClusterAlreadyExistsError)(nil)) {
			__antithesis_instrumentation__.Notify(43517)

			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(43518)
		}
		__antithesis_instrumentation__.Notify(43507)

		l.PrintfCtx(ctx, "cluster creation failed, cleaning up in case it was partially created: %s", err)

		c.destroyState.alloc = nil
		c.Destroy(ctx, closeLogger, l)
		if i >= maxAttempts {
			__antithesis_instrumentation__.Notify(43519)

			cfg.alloc.Release()
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(43520)
		}

	}
}

type attachOpt struct {
	skipValidation bool

	skipStop bool
	skipWipe bool
}

func attachToExistingCluster(
	ctx context.Context,
	name string,
	l *logger.Logger,
	spec spec.ClusterSpec,
	opt attachOpt,
	r *clusterRegistry,
) (*clusterImpl, error) {
	__antithesis_instrumentation__.Notify(43521)
	exp := spec.Expiration()
	if name == "local" {
		__antithesis_instrumentation__.Notify(43526)
		exp = timeutil.Now().Add(100000 * time.Hour)
	} else {
		__antithesis_instrumentation__.Notify(43527)
	}
	__antithesis_instrumentation__.Notify(43522)
	c := &clusterImpl{
		name:       name,
		spec:       spec,
		l:          l,
		expiration: exp,
		destroyState: destroyState{

			owned: false,
		},
		r: r,
	}

	if err := r.registerCluster(c); err != nil {
		__antithesis_instrumentation__.Notify(43528)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(43529)
	}
	__antithesis_instrumentation__.Notify(43523)

	if !opt.skipValidation {
		__antithesis_instrumentation__.Notify(43530)
		if err := c.validate(ctx, spec, l); err != nil {
			__antithesis_instrumentation__.Notify(43531)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(43532)
		}
	} else {
		__antithesis_instrumentation__.Notify(43533)
	}
	__antithesis_instrumentation__.Notify(43524)

	if !opt.skipStop {
		__antithesis_instrumentation__.Notify(43534)
		c.status("stopping cluster")
		if err := c.StopE(ctx, l, option.DefaultStopOpts(), c.All()); err != nil {
			__antithesis_instrumentation__.Notify(43536)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(43537)
		}
		__antithesis_instrumentation__.Notify(43535)
		if !opt.skipWipe {
			__antithesis_instrumentation__.Notify(43538)
			if clusterWipe {
				__antithesis_instrumentation__.Notify(43539)
				if err := c.WipeE(ctx, l, c.All()); err != nil {
					__antithesis_instrumentation__.Notify(43540)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(43541)
				}
			} else {
				__antithesis_instrumentation__.Notify(43542)
				l.Printf("skipping cluster wipe\n")
			}
		} else {
			__antithesis_instrumentation__.Notify(43543)
		}
	} else {
		__antithesis_instrumentation__.Notify(43544)
	}
	__antithesis_instrumentation__.Notify(43525)

	c.status("idle")
	return c, nil
}

func (c *clusterImpl) setTest(t test.Test) {
	__antithesis_instrumentation__.Notify(43545)
	c.t = t
	c.l = t.L()
}

func (c *clusterImpl) StopCockroachGracefullyOnNode(
	ctx context.Context, l *logger.Logger, node int,
) error {
	__antithesis_instrumentation__.Notify(43546)
	port := fmt.Sprintf("{pgport:%d}", node)

	if err := c.RunE(ctx, c.Node(node), "./cockroach quit --insecure --port="+port); err != nil {
		__antithesis_instrumentation__.Notify(43548)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43549)
	}
	__antithesis_instrumentation__.Notify(43547)

	c.Stop(ctx, l, option.DefaultStopOpts(), c.Node(node))

	return nil
}

func (c *clusterImpl) Save(ctx context.Context, msg string, l *logger.Logger) {
	__antithesis_instrumentation__.Notify(43550)
	l.PrintfCtx(ctx, "saving cluster %s for debugging (--debug specified)", c)

	if c.destroyState.owned {
		__antithesis_instrumentation__.Notify(43552)
		c.destroyState.alloc.Freeze()
	} else {
		__antithesis_instrumentation__.Notify(43553)
	}
	__antithesis_instrumentation__.Notify(43551)
	c.r.markClusterAsSaved(c, msg)
	c.destroyState.mu.Lock()
	c.destroyState.mu.saved = true
	c.destroyState.mu.savedMsg = msg
	c.destroyState.mu.Unlock()
}

func (c *clusterImpl) validate(
	ctx context.Context, nodes spec.ClusterSpec, l *logger.Logger,
) error {
	__antithesis_instrumentation__.Notify(43554)

	c.status("checking that existing cluster matches spec")
	pattern := "^" + regexp.QuoteMeta(c.name) + "$"
	cloudClusters, err := roachprod.List(l, false, pattern)
	if err != nil {
		__antithesis_instrumentation__.Notify(43559)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43560)
	}
	__antithesis_instrumentation__.Notify(43555)
	cDetails, ok := cloudClusters.Clusters[c.name]
	if !ok {
		__antithesis_instrumentation__.Notify(43561)
		return fmt.Errorf("cluster %q not found", c.name)
	} else {
		__antithesis_instrumentation__.Notify(43562)
	}
	__antithesis_instrumentation__.Notify(43556)
	if len(cDetails.VMs) < c.spec.NodeCount {
		__antithesis_instrumentation__.Notify(43563)
		return fmt.Errorf("cluster has %d nodes, test requires at least %d", len(cDetails.VMs), c.spec.NodeCount)
	} else {
		__antithesis_instrumentation__.Notify(43564)
	}
	__antithesis_instrumentation__.Notify(43557)
	if cpus := nodes.CPUs; cpus != 0 {
		__antithesis_instrumentation__.Notify(43565)
		for i, vm := range cDetails.VMs {
			__antithesis_instrumentation__.Notify(43566)
			vmCPUs := MachineTypeToCPUs(vm.MachineType)

			if vmCPUs > 0 && func() bool {
				__antithesis_instrumentation__.Notify(43567)
				return vmCPUs < cpus == true
			}() == true {
				__antithesis_instrumentation__.Notify(43568)
				return fmt.Errorf("node %d has %d CPUs, test requires %d", i, vmCPUs, cpus)
			} else {
				__antithesis_instrumentation__.Notify(43569)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(43570)
	}
	__antithesis_instrumentation__.Notify(43558)
	return nil
}

func (c *clusterImpl) lister() option.NodeLister {
	__antithesis_instrumentation__.Notify(43571)
	fatalf := func(string, ...interface{}) { __antithesis_instrumentation__.Notify(43574) }
	__antithesis_instrumentation__.Notify(43572)
	if c.t != nil {
		__antithesis_instrumentation__.Notify(43575)
		fatalf = c.t.Fatalf
	} else {
		__antithesis_instrumentation__.Notify(43576)
	}
	__antithesis_instrumentation__.Notify(43573)
	return option.NodeLister{NodeCount: c.spec.NodeCount, Fatalf: fatalf}
}

func (c *clusterImpl) All() option.NodeListOption {
	__antithesis_instrumentation__.Notify(43577)
	return c.lister().All()
}

func (c *clusterImpl) Range(begin, end int) option.NodeListOption {
	__antithesis_instrumentation__.Notify(43578)
	return c.lister().Range(begin, end)
}

func (c *clusterImpl) Nodes(ns ...int) option.NodeListOption {
	__antithesis_instrumentation__.Notify(43579)
	return c.lister().Nodes(ns...)
}

func (c *clusterImpl) Node(i int) option.NodeListOption {
	__antithesis_instrumentation__.Notify(43580)
	return c.lister().Node(i)
}

func (c *clusterImpl) FetchLogs(ctx context.Context, t test.Test) error {
	__antithesis_instrumentation__.Notify(43581)
	if c.spec.NodeCount == 0 {
		__antithesis_instrumentation__.Notify(43583)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(43584)
	}
	__antithesis_instrumentation__.Notify(43582)

	t.L().Printf("fetching logs\n")
	c.status("fetching logs")

	return contextutil.RunWithTimeout(ctx, "fetch logs", 2*time.Minute, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(43585)
		path := filepath.Join(c.t.ArtifactsDir(), "logs", "unredacted")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			__antithesis_instrumentation__.Notify(43589)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43590)
		}
		__antithesis_instrumentation__.Notify(43586)

		if err := c.Get(ctx, c.l, "logs", path); err != nil {
			__antithesis_instrumentation__.Notify(43591)
			t.L().Printf("failed to fetch logs: %v", err)
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(43592)
				return errors.Wrap(err, "cluster.FetchLogs")
			} else {
				__antithesis_instrumentation__.Notify(43593)
			}
		} else {
			__antithesis_instrumentation__.Notify(43594)
		}
		__antithesis_instrumentation__.Notify(43587)

		if err := c.RunE(ctx, c.All(), "mkdir -p logs/redacted && ./cockroach debug merge-logs --redact logs/*.log > logs/redacted/combined.log"); err != nil {
			__antithesis_instrumentation__.Notify(43595)
			t.L().Printf("failed to redact logs: %v", err)
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(43596)
				return err
			} else {
				__antithesis_instrumentation__.Notify(43597)
			}
		} else {
			__antithesis_instrumentation__.Notify(43598)
		}
		__antithesis_instrumentation__.Notify(43588)
		dest := filepath.Join(c.t.ArtifactsDir(), "logs/cockroach.log")
		return errors.Wrap(c.Get(ctx, c.l, "logs/redacted/combined.log", dest), "cluster.FetchLogs")
	})
}

func saveDiskUsageToLogsDir(ctx context.Context, c cluster.Cluster) error {
	__antithesis_instrumentation__.Notify(43599)

	if c.Spec().NodeCount == 0 || func() bool {
		__antithesis_instrumentation__.Notify(43601)
		return c.IsLocal() == true
	}() == true {
		__antithesis_instrumentation__.Notify(43602)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(43603)
	}
	__antithesis_instrumentation__.Notify(43600)

	return contextutil.RunWithTimeout(ctx, "disk usage", 20*time.Second, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(43604)
		return c.RunE(ctx, c.All(),
			"du -c /mnt/data1 --exclude lost+found >> logs/diskusage.txt")
	})
}

func (c *clusterImpl) CopyRoachprodState(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(43605)
	if c.spec.NodeCount == 0 {
		__antithesis_instrumentation__.Notify(43608)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(43609)
	}
	__antithesis_instrumentation__.Notify(43606)

	const roachprodStateDirName = ".roachprod"
	const roachprodStateName = "roachprod_state"
	u, err := user.Current()
	if err != nil {
		__antithesis_instrumentation__.Notify(43610)
		return errors.Wrap(err, "failed to get current user")
	} else {
		__antithesis_instrumentation__.Notify(43611)
	}
	__antithesis_instrumentation__.Notify(43607)
	src := filepath.Join(u.HomeDir, roachprodStateDirName)
	dest := filepath.Join(c.t.ArtifactsDir(), roachprodStateName)
	cmd := exec.CommandContext(ctx, "cp", "-r", src, dest)
	output, err := cmd.CombinedOutput()
	return errors.Wrapf(err, "command %q failed: output: %v", cmd.Args, string(output))
}

func (c *clusterImpl) FetchTimeseriesData(ctx context.Context, t test.Test) error {
	__antithesis_instrumentation__.Notify(43612)
	return contextutil.RunWithTimeout(ctx, "fetch tsdata", 5*time.Minute, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(43613)
		node := 1
		for ; node <= c.spec.NodeCount; node++ {
			__antithesis_instrumentation__.Notify(43623)
			db, err := c.ConnE(ctx, t.L(), node)
			if err == nil {
				__antithesis_instrumentation__.Notify(43626)
				err = db.Ping()
				db.Close()
			} else {
				__antithesis_instrumentation__.Notify(43627)
			}
			__antithesis_instrumentation__.Notify(43624)
			if err != nil {
				__antithesis_instrumentation__.Notify(43628)
				t.L().Printf("node %d not responding to SQL, trying next one", node)
				continue
			} else {
				__antithesis_instrumentation__.Notify(43629)
			}
			__antithesis_instrumentation__.Notify(43625)
			break
		}
		__antithesis_instrumentation__.Notify(43614)
		if node > c.spec.NodeCount {
			__antithesis_instrumentation__.Notify(43630)
			return errors.New("no node responds to SQL, cannot fetch tsdata")
		} else {
			__antithesis_instrumentation__.Notify(43631)
		}
		__antithesis_instrumentation__.Notify(43615)
		if err := c.RunE(
			ctx, c.Node(node), "./cockroach debug tsdump --insecure --format=raw > tsdump.gob",
		); err != nil {
			__antithesis_instrumentation__.Notify(43632)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43633)
		}
		__antithesis_instrumentation__.Notify(43616)
		tsDumpGob := filepath.Join(c.t.ArtifactsDir(), "tsdump.gob")
		if err := c.Get(
			ctx, c.l, "tsdump.gob", tsDumpGob, c.Node(node),
		); err != nil {
			__antithesis_instrumentation__.Notify(43634)
			return errors.Wrap(err, "cluster.FetchTimeseriesData")
		} else {
			__antithesis_instrumentation__.Notify(43635)
		}
		__antithesis_instrumentation__.Notify(43617)
		db, err := c.ConnE(ctx, t.L(), node)
		if err != nil {
			__antithesis_instrumentation__.Notify(43636)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43637)
		}
		__antithesis_instrumentation__.Notify(43618)
		defer db.Close()
		rows, err := db.QueryContext(
			ctx,
			` SELECT store_id, node_id FROM crdb_internal.kv_store_status`,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(43638)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43639)
		}
		__antithesis_instrumentation__.Notify(43619)
		defer rows.Close()
		var buf bytes.Buffer
		for rows.Next() {
			__antithesis_instrumentation__.Notify(43640)
			var storeID, nodeID int
			if err := rows.Scan(&storeID, &nodeID); err != nil {
				__antithesis_instrumentation__.Notify(43642)
				return err
			} else {
				__antithesis_instrumentation__.Notify(43643)
			}
			__antithesis_instrumentation__.Notify(43641)
			fmt.Fprintf(&buf, "%d: %d\n", storeID, nodeID)
		}
		__antithesis_instrumentation__.Notify(43620)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(43644)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43645)
		}
		__antithesis_instrumentation__.Notify(43621)
		if err := ioutil.WriteFile(tsDumpGob+".yaml", buf.Bytes(), 0644); err != nil {
			__antithesis_instrumentation__.Notify(43646)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43647)
		}
		__antithesis_instrumentation__.Notify(43622)
		return ioutil.WriteFile(tsDumpGob+"-run.sh", []byte(`#!/usr/bin/env bash

COCKROACH_DEBUG_TS_IMPORT_FILE=tsdump.gob cockroach start-single-node --insecure
`), 0755)
	})
}

func (c *clusterImpl) FetchDebugZip(ctx context.Context, t test.Test) error {
	__antithesis_instrumentation__.Notify(43648)
	if c.spec.NodeCount == 0 {
		__antithesis_instrumentation__.Notify(43650)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(43651)
	}
	__antithesis_instrumentation__.Notify(43649)

	t.L().Printf("fetching debug zip\n")
	c.status("fetching debug zip")

	return contextutil.RunWithTimeout(ctx, "debug zip", 5*time.Minute, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(43652)
		const zipName = "debug.zip"
		path := filepath.Join(c.t.ArtifactsDir(), zipName)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			__antithesis_instrumentation__.Notify(43655)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43656)
		}
		__antithesis_instrumentation__.Notify(43653)

		for i := 1; i <= c.spec.NodeCount; i++ {
			__antithesis_instrumentation__.Notify(43657)

			si := strconv.Itoa(i)
			cmd := []string{"./cockroach", "debug", "zip", "--exclude-files='*.log,*.txt,*.pprof'", "--url", "{pgurl:" + si + "}", zipName}
			if err := c.RunE(ctx, c.All(), cmd...); err != nil {
				__antithesis_instrumentation__.Notify(43659)
				t.L().Printf("./cockroach debug zip failed: %v", err)
				if i < c.spec.NodeCount {
					__antithesis_instrumentation__.Notify(43661)
					continue
				} else {
					__antithesis_instrumentation__.Notify(43662)
				}
				__antithesis_instrumentation__.Notify(43660)
				return err
			} else {
				__antithesis_instrumentation__.Notify(43663)
			}
			__antithesis_instrumentation__.Notify(43658)
			return errors.Wrap(c.Get(ctx, c.l, zipName, path, c.Node(i)), "cluster.FetchDebugZip")
		}
		__antithesis_instrumentation__.Notify(43654)
		return nil
	})
}

func (c *clusterImpl) assertNoDeadNode(ctx context.Context, t test.Test) {
	__antithesis_instrumentation__.Notify(43664)
	if c.spec.NodeCount == 0 {
		__antithesis_instrumentation__.Notify(43666)

		return
	} else {
		__antithesis_instrumentation__.Notify(43667)
	}
	__antithesis_instrumentation__.Notify(43665)

	_, err := roachprod.Monitor(ctx, t.L(), c.name, install.MonitorOpts{OneShot: true, IgnoreEmptyNodes: true})

	if err != nil {
		__antithesis_instrumentation__.Notify(43668)
		t.Errorf("dead node detection: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(43669)
	}
}

func (c *clusterImpl) CheckReplicaDivergenceOnDB(
	ctx context.Context, l *logger.Logger, db *gosql.DB,
) error {
	__antithesis_instrumentation__.Notify(43670)

	rows, err := db.QueryContext(ctx, `
SET statement_timeout = '5m';
SELECT t.range_id, t.start_key_pretty, t.status, t.detail
FROM
crdb_internal.check_consistency(true, '', '') as t
WHERE t.status NOT IN ('RANGE_CONSISTENT', 'RANGE_INDETERMINATE')`)
	if err != nil {
		__antithesis_instrumentation__.Notify(43674)

		l.Printf("consistency check failed with %v; ignoring", err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(43675)
	}
	__antithesis_instrumentation__.Notify(43671)
	defer rows.Close()
	var finalErr error
	for rows.Next() {
		__antithesis_instrumentation__.Notify(43676)
		var rangeID int32
		var prettyKey, status, detail string
		if scanErr := rows.Scan(&rangeID, &prettyKey, &status, &detail); scanErr != nil {
			__antithesis_instrumentation__.Notify(43678)
			l.Printf("consistency check failed with %v; ignoring", scanErr)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(43679)
		}
		__antithesis_instrumentation__.Notify(43677)
		finalErr = errors.CombineErrors(finalErr,
			errors.Newf("r%d (%s) is inconsistent: %s %s\n", rangeID, prettyKey, status, detail))
	}
	__antithesis_instrumentation__.Notify(43672)
	if err := rows.Err(); err != nil {
		__antithesis_instrumentation__.Notify(43680)
		l.Printf("consistency check failed with %v; ignoring", err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(43681)
	}
	__antithesis_instrumentation__.Notify(43673)
	return finalErr
}

func (c *clusterImpl) FailOnReplicaDivergence(ctx context.Context, t *testImpl) {
	__antithesis_instrumentation__.Notify(43682)
	if c.spec.NodeCount < 1 {
		__antithesis_instrumentation__.Notify(43686)
		return
	} else {
		__antithesis_instrumentation__.Notify(43687)
	}
	__antithesis_instrumentation__.Notify(43683)

	var db *gosql.DB
	for i := 1; i <= c.spec.NodeCount; i++ {
		__antithesis_instrumentation__.Notify(43688)

		if err := contextutil.RunWithTimeout(
			ctx, "find live node", 5*time.Second,
			func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(43690)
				db = c.Conn(ctx, t.L(), i)
				_, err := db.ExecContext(ctx, `;`)
				return err
			},
		); err != nil {
			__antithesis_instrumentation__.Notify(43691)
			_ = db.Close()
			db = nil
			continue
		} else {
			__antithesis_instrumentation__.Notify(43692)
		}
		__antithesis_instrumentation__.Notify(43689)
		t.L().Printf("running (fast) consistency checks on node %d", i)
		break
	}
	__antithesis_instrumentation__.Notify(43684)
	if db == nil {
		__antithesis_instrumentation__.Notify(43693)
		t.L().Printf("no live node found, skipping consistency check")
		return
	} else {
		__antithesis_instrumentation__.Notify(43694)
	}
	__antithesis_instrumentation__.Notify(43685)
	defer db.Close()

	if err := contextutil.RunWithTimeout(
		ctx, "consistency check", 5*time.Minute,
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(43695)
			return c.CheckReplicaDivergenceOnDB(ctx, t.L(), db)
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(43696)
		t.Errorf("consistency check failed: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(43697)
	}
}

func (c *clusterImpl) FetchDmesg(ctx context.Context, t test.Test) error {
	__antithesis_instrumentation__.Notify(43698)
	if c.spec.NodeCount == 0 || func() bool {
		__antithesis_instrumentation__.Notify(43700)
		return c.IsLocal() == true
	}() == true {
		__antithesis_instrumentation__.Notify(43701)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(43702)
	}
	__antithesis_instrumentation__.Notify(43699)

	t.L().Printf("fetching dmesg\n")
	c.status("fetching dmesg")

	return contextutil.RunWithTimeout(ctx, "dmesg", 20*time.Second, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(43703)
		const name = "dmesg.txt"
		path := filepath.Join(c.t.ArtifactsDir(), name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			__antithesis_instrumentation__.Notify(43708)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43709)
		}
		__antithesis_instrumentation__.Notify(43704)

		cmd := []string{"/bin/bash", "-c", "'sudo dmesg > " + name + "'"}
		var results []install.RunResultDetails
		var combinedDmesgError error
		if results, combinedDmesgError = c.RunWithDetails(ctx, nil, c.All(), cmd...); combinedDmesgError != nil {
			__antithesis_instrumentation__.Notify(43710)
			return errors.Wrap(combinedDmesgError, "cluster.FetchDmesg")
		} else {
			__antithesis_instrumentation__.Notify(43711)
		}
		__antithesis_instrumentation__.Notify(43705)

		var successfulNodes option.NodeListOption
		for _, result := range results {
			__antithesis_instrumentation__.Notify(43712)
			if result.Err != nil {
				__antithesis_instrumentation__.Notify(43713)

				combinedDmesgError = errors.CombineErrors(combinedDmesgError, result.Err)
				t.L().Printf("running dmesg failed on node %d: %v", result.Node, result.Err)
			} else {
				__antithesis_instrumentation__.Notify(43714)

				successfulNodes = append(successfulNodes, int(result.Node))
			}
		}
		__antithesis_instrumentation__.Notify(43706)

		if err := c.Get(ctx, c.l, name, path, successfulNodes); err != nil {
			__antithesis_instrumentation__.Notify(43715)
			t.L().Printf("getting dmesg files failed: %v", err)
			return errors.Wrap(err, "cluster.FetchDmesg")
		} else {
			__antithesis_instrumentation__.Notify(43716)
		}
		__antithesis_instrumentation__.Notify(43707)

		return combinedDmesgError
	})
}

func (c *clusterImpl) FetchJournalctl(ctx context.Context, t test.Test) error {
	__antithesis_instrumentation__.Notify(43717)
	if c.spec.NodeCount == 0 || func() bool {
		__antithesis_instrumentation__.Notify(43719)
		return c.IsLocal() == true
	}() == true {
		__antithesis_instrumentation__.Notify(43720)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(43721)
	}
	__antithesis_instrumentation__.Notify(43718)

	t.L().Printf("fetching journalctl\n")
	c.status("fetching journalctl")

	return contextutil.RunWithTimeout(ctx, "journalctl", 20*time.Second, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(43722)
		const name = "journalctl.txt"
		path := filepath.Join(c.t.ArtifactsDir(), name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			__antithesis_instrumentation__.Notify(43727)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43728)
		}
		__antithesis_instrumentation__.Notify(43723)

		cmd := []string{"/bin/bash", "-c", "'sudo journalctl > " + name + "'"}
		var results []install.RunResultDetails
		var combinedJournalctlError error
		if results, combinedJournalctlError = c.RunWithDetails(ctx, nil, c.All(), cmd...); combinedJournalctlError != nil {
			__antithesis_instrumentation__.Notify(43729)
			return errors.Wrap(combinedJournalctlError, "cluster.FetchJournalctl")
		} else {
			__antithesis_instrumentation__.Notify(43730)
		}
		__antithesis_instrumentation__.Notify(43724)

		var successfulNodes option.NodeListOption
		for _, result := range results {
			__antithesis_instrumentation__.Notify(43731)
			if result.Err != nil {
				__antithesis_instrumentation__.Notify(43732)

				combinedJournalctlError = errors.CombineErrors(combinedJournalctlError, result.Err)
				t.L().Printf("running journalctl failed on node %d: %v", result.Node, result.Err)
			} else {
				__antithesis_instrumentation__.Notify(43733)

				successfulNodes = append(successfulNodes, int(result.Node))
			}
		}
		__antithesis_instrumentation__.Notify(43725)

		if err := c.Get(ctx, c.l, name, path, successfulNodes); err != nil {
			__antithesis_instrumentation__.Notify(43734)
			t.L().Printf("getting files failed: %v", err)
			return errors.Wrap(err, "cluster.FetchJournalctl")
		} else {
			__antithesis_instrumentation__.Notify(43735)
		}
		__antithesis_instrumentation__.Notify(43726)

		return combinedJournalctlError
	})
}

func (c *clusterImpl) FetchCores(ctx context.Context, t test.Test) error {
	__antithesis_instrumentation__.Notify(43736)
	if c.spec.NodeCount == 0 || func() bool {
		__antithesis_instrumentation__.Notify(43739)
		return c.IsLocal() == true
	}() == true {
		__antithesis_instrumentation__.Notify(43740)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(43741)
	}
	__antithesis_instrumentation__.Notify(43737)

	if true {
		__antithesis_instrumentation__.Notify(43742)

		t.L().Printf("skipped fetching cores\n")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(43743)
	}
	__antithesis_instrumentation__.Notify(43738)

	t.L().Printf("fetching cores\n")
	c.status("fetching cores")

	return contextutil.RunWithTimeout(ctx, "cores", 60*time.Second, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(43744)
		path := filepath.Join(c.t.ArtifactsDir(), "cores")
		return errors.Wrap(c.Get(ctx, c.l, "/mnt/data1/cores", path), "cluster.FetchCores")
	})
}

type closeLoggerOpt bool

const (
	closeLogger     closeLoggerOpt = true
	dontCloseLogger                = false
)

func (c *clusterImpl) Destroy(ctx context.Context, lo closeLoggerOpt, l *logger.Logger) {
	__antithesis_instrumentation__.Notify(43745)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43748)
		return
	} else {
		__antithesis_instrumentation__.Notify(43749)
	}
	__antithesis_instrumentation__.Notify(43746)
	if c.spec.NodeCount == 0 {
		__antithesis_instrumentation__.Notify(43750)

		c.r.unregisterCluster(c)
		return
	} else {
		__antithesis_instrumentation__.Notify(43751)
	}
	__antithesis_instrumentation__.Notify(43747)

	ch := c.doDestroy(ctx, l)
	<-ch

	if lo == closeLogger && func() bool {
		__antithesis_instrumentation__.Notify(43752)
		return c.l != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(43753)
		c.closeLogger()
	} else {
		__antithesis_instrumentation__.Notify(43754)
	}
}

func (c *clusterImpl) doDestroy(ctx context.Context, l *logger.Logger) <-chan struct{} {
	__antithesis_instrumentation__.Notify(43755)
	var inFlight <-chan struct{}
	c.destroyState.mu.Lock()
	if c.destroyState.mu.saved {
		__antithesis_instrumentation__.Notify(43760)

		c.destroyState.mu.Unlock()
		ch := make(chan struct{})
		close(ch)
		return ch
	} else {
		__antithesis_instrumentation__.Notify(43761)
	}
	__antithesis_instrumentation__.Notify(43756)
	if c.destroyState.mu.destroyed == nil {
		__antithesis_instrumentation__.Notify(43762)
		c.destroyState.mu.destroyed = make(chan struct{})
	} else {
		__antithesis_instrumentation__.Notify(43763)
		inFlight = c.destroyState.mu.destroyed
	}
	__antithesis_instrumentation__.Notify(43757)
	c.destroyState.mu.Unlock()
	if inFlight != nil {
		__antithesis_instrumentation__.Notify(43764)
		return inFlight
	} else {
		__antithesis_instrumentation__.Notify(43765)
	}
	__antithesis_instrumentation__.Notify(43758)

	if clusterWipe {
		__antithesis_instrumentation__.Notify(43766)
		if c.destroyState.owned {
			__antithesis_instrumentation__.Notify(43767)
			l.PrintfCtx(ctx, "destroying cluster %s...", c)
			c.status("destroying cluster")

			if err := roachprod.Destroy(l, false, false, c.name); err != nil {
				__antithesis_instrumentation__.Notify(43769)
				l.ErrorfCtx(ctx, "error destroying cluster %s: %s", c, err)
			} else {
				__antithesis_instrumentation__.Notify(43770)
				l.PrintfCtx(ctx, "destroying cluster %s... done", c)
			}
			__antithesis_instrumentation__.Notify(43768)
			if c.destroyState.alloc != nil {
				__antithesis_instrumentation__.Notify(43771)

				c.destroyState.alloc.Release()
			} else {
				__antithesis_instrumentation__.Notify(43772)
			}
		} else {
			__antithesis_instrumentation__.Notify(43773)
			l.PrintfCtx(ctx, "wiping cluster %s", c)
			c.status("wiping cluster")
			if err := roachprod.Wipe(ctx, l, c.name, false); err != nil {
				__antithesis_instrumentation__.Notify(43775)
				l.Errorf("%s", err)
			} else {
				__antithesis_instrumentation__.Notify(43776)
			}
			__antithesis_instrumentation__.Notify(43774)
			if c.localCertsDir != "" {
				__antithesis_instrumentation__.Notify(43777)
				if err := os.RemoveAll(c.localCertsDir); err != nil {
					__antithesis_instrumentation__.Notify(43779)
					l.Errorf("failed to remove local certs in %s: %s", c.localCertsDir, err)
				} else {
					__antithesis_instrumentation__.Notify(43780)
				}
				__antithesis_instrumentation__.Notify(43778)
				c.localCertsDir = ""
			} else {
				__antithesis_instrumentation__.Notify(43781)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(43782)
		l.Printf("skipping cluster wipe\n")
	}
	__antithesis_instrumentation__.Notify(43759)
	c.r.unregisterCluster(c)
	c.destroyState.mu.Lock()
	ch := c.destroyState.mu.destroyed
	close(ch)
	c.destroyState.mu.Unlock()
	return ch
}

func (c *clusterImpl) Put(ctx context.Context, src, dest string, nodes ...option.Option) {
	__antithesis_instrumentation__.Notify(43783)
	if err := c.PutE(ctx, c.l, src, dest, nodes...); err != nil {
		__antithesis_instrumentation__.Notify(43784)
		c.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(43785)
	}
}

func (c *clusterImpl) PutE(
	ctx context.Context, l *logger.Logger, src, dest string, nodes ...option.Option,
) error {
	__antithesis_instrumentation__.Notify(43786)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43788)
		return errors.Wrap(ctx.Err(), "cluster.Put")
	} else {
		__antithesis_instrumentation__.Notify(43789)
	}
	__antithesis_instrumentation__.Notify(43787)

	c.status("uploading file")
	defer c.status("")
	return errors.Wrap(roachprod.Put(ctx, l, c.MakeNodes(nodes...), src, dest, true), "cluster.PutE")
}

func (c *clusterImpl) PutLibraries(ctx context.Context, libraryDir string) error {
	__antithesis_instrumentation__.Notify(43790)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43794)
		return errors.Wrap(ctx.Err(), "cluster.Put")
	} else {
		__antithesis_instrumentation__.Notify(43795)
	}
	__antithesis_instrumentation__.Notify(43791)

	c.status("uploading library files")
	defer c.status("")

	if err := c.RunE(ctx, c.All(), "mkdir", "-p", libraryDir); err != nil {
		__antithesis_instrumentation__.Notify(43796)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43797)
	}
	__antithesis_instrumentation__.Notify(43792)
	for _, libraryFilePath := range libraryFilePaths {
		__antithesis_instrumentation__.Notify(43798)
		putPath := filepath.Join(libraryDir, filepath.Base(libraryFilePath))
		if err := c.PutE(
			ctx,
			c.l,
			libraryFilePath,
			putPath,
		); err != nil {
			__antithesis_instrumentation__.Notify(43799)
			return errors.Wrap(err, "cluster.PutLibraries")
		} else {
			__antithesis_instrumentation__.Notify(43800)
		}
	}
	__antithesis_instrumentation__.Notify(43793)
	return nil
}

func (c *clusterImpl) Stage(
	ctx context.Context,
	l *logger.Logger,
	application, versionOrSHA, dir string,
	opts ...option.Option,
) error {
	__antithesis_instrumentation__.Notify(43801)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43803)
		return errors.Wrap(ctx.Err(), "cluster.Stage")
	} else {
		__antithesis_instrumentation__.Notify(43804)
	}
	__antithesis_instrumentation__.Notify(43802)
	c.status("staging binary")
	defer c.status("")
	return errors.Wrap(roachprod.Stage(ctx, l, c.MakeNodes(opts...), "", dir, application, versionOrSHA), "cluster.Stage")
}

func (c *clusterImpl) Get(
	ctx context.Context, l *logger.Logger, src, dest string, opts ...option.Option,
) error {
	__antithesis_instrumentation__.Notify(43805)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43807)
		return errors.Wrap(ctx.Err(), "cluster.Get")
	} else {
		__antithesis_instrumentation__.Notify(43808)
	}
	__antithesis_instrumentation__.Notify(43806)
	c.status(fmt.Sprintf("getting %v", src))
	defer c.status("")
	return errors.Wrap(roachprod.Get(l, c.MakeNodes(opts...), src, dest), "cluster.Get")
}

func (c *clusterImpl) PutString(
	ctx context.Context, content, dest string, mode os.FileMode, nodes ...option.Option,
) error {
	__antithesis_instrumentation__.Notify(43809)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43814)
		return errors.Wrap(ctx.Err(), "cluster.PutString")
	} else {
		__antithesis_instrumentation__.Notify(43815)
	}
	__antithesis_instrumentation__.Notify(43810)
	c.status("uploading string")
	defer c.status("")

	temp, err := ioutil.TempFile("", filepath.Base(dest))
	if err != nil {
		__antithesis_instrumentation__.Notify(43816)
		return errors.Wrap(err, "cluster.PutString")
	} else {
		__antithesis_instrumentation__.Notify(43817)
	}
	__antithesis_instrumentation__.Notify(43811)
	if _, err := temp.WriteString(content); err != nil {
		__antithesis_instrumentation__.Notify(43818)
		return errors.Wrap(err, "cluster.PutString")
	} else {
		__antithesis_instrumentation__.Notify(43819)
	}
	__antithesis_instrumentation__.Notify(43812)
	temp.Close()
	src := temp.Name()

	if err := os.Chmod(src, mode); err != nil {
		__antithesis_instrumentation__.Notify(43820)
		return errors.Wrap(err, "cluster.PutString")
	} else {
		__antithesis_instrumentation__.Notify(43821)
	}
	__antithesis_instrumentation__.Notify(43813)

	return errors.Wrap(c.PutE(ctx, c.l, src, dest, nodes...), "cluster.PutString")
}

func (c *clusterImpl) GitClone(
	ctx context.Context, l *logger.Logger, src, dest, branch string, nodes option.NodeListOption,
) error {
	__antithesis_instrumentation__.Notify(43822)
	cmd := []string{"bash", "-e", "-c", fmt.Sprintf(`'
		if ! test -d %[1]s; then
			git clone -b %[2]s --depth 1 %[3]s %[1]s
  		else
			cd %[1]s
		git fetch origin
		git checkout origin/%[2]s
  		fi
  		'`, dest, branch, src),
	}
	return errors.Wrap(c.RunE(ctx, nodes, cmd...), "cluster.GitClone")
}

func (c *clusterImpl) setStatusForClusterOpt(
	operation string, worker bool, nodesOptions ...option.Option,
) {
	__antithesis_instrumentation__.Notify(43823)
	var nodes option.NodeListOption
	for _, o := range nodesOptions {
		__antithesis_instrumentation__.Notify(43826)
		if s, ok := o.(nodeSelector); ok {
			__antithesis_instrumentation__.Notify(43827)
			nodes = s.Merge(nodes)
		} else {
			__antithesis_instrumentation__.Notify(43828)
		}
	}
	__antithesis_instrumentation__.Notify(43824)

	nodesString := " cluster"
	if len(nodes) != 0 {
		__antithesis_instrumentation__.Notify(43829)
		nodesString = " nodes " + nodes.String()
	} else {
		__antithesis_instrumentation__.Notify(43830)
	}
	__antithesis_instrumentation__.Notify(43825)
	msg := operation + nodesString
	if worker {
		__antithesis_instrumentation__.Notify(43831)
		c.workerStatus(msg)
	} else {
		__antithesis_instrumentation__.Notify(43832)
		c.status(msg)
	}
}

func (c *clusterImpl) clearStatusForClusterOpt(worker bool) {
	__antithesis_instrumentation__.Notify(43833)
	if worker {
		__antithesis_instrumentation__.Notify(43834)
		c.workerStatus()
	} else {
		__antithesis_instrumentation__.Notify(43835)
		c.status()
	}
}

func (c *clusterImpl) StartE(
	ctx context.Context,
	l *logger.Logger,
	startOpts option.StartOpts,
	settings install.ClusterSettings,
	opts ...option.Option,
) error {
	__antithesis_instrumentation__.Notify(43836)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43844)
		return errors.Wrap(ctx.Err(), "cluster.StartE")
	} else {
		__antithesis_instrumentation__.Notify(43845)
	}
	__antithesis_instrumentation__.Notify(43837)

	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43846)
		return ctx.Err()
	} else {
		__antithesis_instrumentation__.Notify(43847)
	}
	__antithesis_instrumentation__.Notify(43838)
	c.setStatusForClusterOpt("starting", startOpts.RoachtestOpts.Worker, opts...)
	defer c.clearStatusForClusterOpt(startOpts.RoachtestOpts.Worker)

	startOpts.RoachprodOpts.EncryptedStores = c.encAtRest

	if !envExists(settings.Env, "COCKROACH_ENABLE_RPC_COMPRESSION") {
		__antithesis_instrumentation__.Notify(43848)

		settings.Env = append(settings.Env, "COCKROACH_ENABLE_RPC_COMPRESSION=false")
	} else {
		__antithesis_instrumentation__.Notify(43849)
	}
	__antithesis_instrumentation__.Notify(43839)

	if !envExists(settings.Env, "COCKROACH_UI_RELEASE_NOTES_SIGNUP_DISMISSED") {
		__antithesis_instrumentation__.Notify(43850)

		settings.Env = append(settings.Env, "COCKROACH_UI_RELEASE_NOTES_SIGNUP_DISMISSED=true")
	} else {
		__antithesis_instrumentation__.Notify(43851)
	}
	__antithesis_instrumentation__.Notify(43840)

	if !envExists(settings.Env, "COCKROACH_CRASH_ON_SPAN_USE_AFTER_FINISH") {
		__antithesis_instrumentation__.Notify(43852)

		settings.Env = append(settings.Env, "COCKROACH_CRASH_ON_SPAN_USE_AFTER_FINISH=true")
	} else {
		__antithesis_instrumentation__.Notify(43853)
	}
	__antithesis_instrumentation__.Notify(43841)

	clusterSettingsOpts := []install.ClusterSettingOption{
		install.TagOption(settings.Tag),
		install.PGUrlCertsDirOption(settings.PGUrlCertsDir),
		install.SecureOption(settings.Secure),
		install.UseTreeDistOption(settings.UseTreeDist),
		install.EnvOption(settings.Env),
		install.NumRacksOption(settings.NumRacks),
		install.BinaryOption(settings.Binary),
	}

	if err := roachprod.Start(ctx, l, c.MakeNodes(opts...), startOpts.RoachprodOpts, clusterSettingsOpts...); err != nil {
		__antithesis_instrumentation__.Notify(43854)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43855)
	}
	__antithesis_instrumentation__.Notify(43842)

	if settings.Secure {
		__antithesis_instrumentation__.Notify(43856)
		var err error
		c.localCertsDir, err = ioutil.TempDir("", "roachtest-certs")
		if err != nil {
			__antithesis_instrumentation__.Notify(43859)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43860)
		}
		__antithesis_instrumentation__.Notify(43857)

		c.localCertsDir = filepath.Join(c.localCertsDir, "certs")

		if err := c.Get(ctx, c.l, "./certs", c.localCertsDir, c.Node(1)); err != nil {
			__antithesis_instrumentation__.Notify(43861)
			return errors.Wrap(err, "cluster.StartE")
		} else {
			__antithesis_instrumentation__.Notify(43862)
		}
		__antithesis_instrumentation__.Notify(43858)

		if err := filepath.Walk(c.localCertsDir, func(path string, info fs.FileInfo, err error) error {
			__antithesis_instrumentation__.Notify(43863)
			if info.IsDir() {
				__antithesis_instrumentation__.Notify(43865)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(43866)
			}
			__antithesis_instrumentation__.Notify(43864)
			return os.Chmod(path, 0600)
		}); err != nil {
			__antithesis_instrumentation__.Notify(43867)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43868)
		}
	} else {
		__antithesis_instrumentation__.Notify(43869)
	}
	__antithesis_instrumentation__.Notify(43843)
	return nil
}

func (c *clusterImpl) Start(
	ctx context.Context,
	l *logger.Logger,
	startOpts option.StartOpts,
	settings install.ClusterSettings,
	opts ...option.Option,
) {
	__antithesis_instrumentation__.Notify(43870)
	if err := c.StartE(ctx, l, startOpts, settings, opts...); err != nil {
		__antithesis_instrumentation__.Notify(43871)
		c.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(43872)
	}
}

func envExists(envs []string, prefix string) bool {
	__antithesis_instrumentation__.Notify(43873)
	for _, env := range envs {
		__antithesis_instrumentation__.Notify(43875)
		if strings.HasPrefix(env, prefix) {
			__antithesis_instrumentation__.Notify(43876)
			return true
		} else {
			__antithesis_instrumentation__.Notify(43877)
		}
	}
	__antithesis_instrumentation__.Notify(43874)
	return false
}

func (c *clusterImpl) StopE(
	ctx context.Context, l *logger.Logger, stopOpts option.StopOpts, nodes ...option.Option,
) error {
	__antithesis_instrumentation__.Notify(43878)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43881)
		return errors.Wrap(ctx.Err(), "cluster.StopE")
	} else {
		__antithesis_instrumentation__.Notify(43882)
	}
	__antithesis_instrumentation__.Notify(43879)
	if c.spec.NodeCount == 0 {
		__antithesis_instrumentation__.Notify(43883)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(43884)
	}
	__antithesis_instrumentation__.Notify(43880)
	c.setStatusForClusterOpt("stopping", stopOpts.RoachtestOpts.Worker, nodes...)
	defer c.clearStatusForClusterOpt(stopOpts.RoachtestOpts.Worker)
	return errors.Wrap(roachprod.Stop(ctx, l, c.MakeNodes(nodes...), stopOpts.RoachprodOpts), "cluster.StopE")
}

func (c *clusterImpl) Stop(
	ctx context.Context, l *logger.Logger, stopOpts option.StopOpts, opts ...option.Option,
) {
	__antithesis_instrumentation__.Notify(43885)
	if c.t.Failed() {
		__antithesis_instrumentation__.Notify(43887)

		return
	} else {
		__antithesis_instrumentation__.Notify(43888)
	}
	__antithesis_instrumentation__.Notify(43886)
	if err := c.StopE(ctx, l, stopOpts, opts...); err != nil {
		__antithesis_instrumentation__.Notify(43889)
		c.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(43890)
	}
}

func (c *clusterImpl) Reset(ctx context.Context, l *logger.Logger) error {
	__antithesis_instrumentation__.Notify(43891)
	if c.t.Failed() {
		__antithesis_instrumentation__.Notify(43894)
		return errors.New("already failed")
	} else {
		__antithesis_instrumentation__.Notify(43895)
	}
	__antithesis_instrumentation__.Notify(43892)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43896)
		return errors.Wrap(ctx.Err(), "cluster.Reset")
	} else {
		__antithesis_instrumentation__.Notify(43897)
	}
	__antithesis_instrumentation__.Notify(43893)
	c.status("resetting cluster")
	defer c.status()
	return errors.Wrap(roachprod.Reset(l, c.name), "cluster.Reset")
}

func (c *clusterImpl) WipeE(ctx context.Context, l *logger.Logger, nodes ...option.Option) error {
	__antithesis_instrumentation__.Notify(43898)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43901)
		return errors.Wrap(ctx.Err(), "cluster.WipeE")
	} else {
		__antithesis_instrumentation__.Notify(43902)
	}
	__antithesis_instrumentation__.Notify(43899)
	if c.spec.NodeCount == 0 {
		__antithesis_instrumentation__.Notify(43903)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(43904)
	}
	__antithesis_instrumentation__.Notify(43900)
	c.setStatusForClusterOpt("wiping", false, nodes...)
	defer c.clearStatusForClusterOpt(false)
	return roachprod.Wipe(ctx, l, c.MakeNodes(nodes...), false)
}

func (c *clusterImpl) Wipe(ctx context.Context, nodes ...option.Option) {
	__antithesis_instrumentation__.Notify(43905)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43907)
		return
	} else {
		__antithesis_instrumentation__.Notify(43908)
	}
	__antithesis_instrumentation__.Notify(43906)
	if err := c.WipeE(ctx, c.l, nodes...); err != nil {
		__antithesis_instrumentation__.Notify(43909)
		c.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(43910)
	}
}

func (c *clusterImpl) Run(ctx context.Context, node option.NodeListOption, args ...string) {
	__antithesis_instrumentation__.Notify(43911)
	err := c.RunE(ctx, node, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(43912)
		c.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(43913)
	}
}

func (c *clusterImpl) RunE(ctx context.Context, node option.NodeListOption, args ...string) error {
	__antithesis_instrumentation__.Notify(43914)
	if len(args) == 0 {
		__antithesis_instrumentation__.Notify(43920)
		return errors.New("No command passed")
	} else {
		__antithesis_instrumentation__.Notify(43921)
	}
	__antithesis_instrumentation__.Notify(43915)
	l, logFile, err := c.loggerForCmd(node, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(43922)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43923)
	}
	__antithesis_instrumentation__.Notify(43916)

	if err := errors.Wrap(ctx.Err(), "cluster.RunE"); err != nil {
		__antithesis_instrumentation__.Notify(43924)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43925)
	}
	__antithesis_instrumentation__.Notify(43917)
	err = execCmd(ctx, l, c.MakeNodes(node), args...)

	l.Printf("> result: %+v", err)
	if err := ctx.Err(); err != nil {
		__antithesis_instrumentation__.Notify(43926)
		l.Printf("(note: incoming context was canceled: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(43927)
	}
	__antithesis_instrumentation__.Notify(43918)
	physicalFileName := l.File.Name()
	l.Close()
	if err != nil {
		__antithesis_instrumentation__.Notify(43928)
		failedPhysicalFileName := strings.TrimSuffix(physicalFileName, ".log") + ".failed"
		if failedFile, err2 := os.Create(failedPhysicalFileName); err2 != nil {
			__antithesis_instrumentation__.Notify(43929)
			failedFile.Close()
		} else {
			__antithesis_instrumentation__.Notify(43930)
		}
	} else {
		__antithesis_instrumentation__.Notify(43931)
	}
	__antithesis_instrumentation__.Notify(43919)
	err = errors.Wrapf(err, "output in %s", logFile)
	return err
}

func (c *clusterImpl) RunWithDetailsSingleNode(
	ctx context.Context, testLogger *logger.Logger, nodes option.NodeListOption, args ...string,
) (install.RunResultDetails, error) {
	__antithesis_instrumentation__.Notify(43932)
	if len(nodes) != 1 {
		__antithesis_instrumentation__.Notify(43935)
		return install.RunResultDetails{}, errors.Newf("RunWithDetailsSingleNode received %d nodes. Use RunWithDetails if you need to run on multiple nodes.", len(nodes))
	} else {
		__antithesis_instrumentation__.Notify(43936)
	}
	__antithesis_instrumentation__.Notify(43933)
	results, err := c.RunWithDetails(ctx, testLogger, nodes, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(43937)
		return install.RunResultDetails{}, err
	} else {
		__antithesis_instrumentation__.Notify(43938)
	}
	__antithesis_instrumentation__.Notify(43934)
	return results[0], results[0].Err
}

func (c *clusterImpl) RunWithDetails(
	ctx context.Context, testLogger *logger.Logger, nodes option.NodeListOption, args ...string,
) ([]install.RunResultDetails, error) {
	__antithesis_instrumentation__.Notify(43939)
	if len(args) == 0 {
		__antithesis_instrumentation__.Notify(43947)
		return nil, errors.New("No command passed")
	} else {
		__antithesis_instrumentation__.Notify(43948)
	}
	__antithesis_instrumentation__.Notify(43940)
	l, _, err := c.loggerForCmd(nodes, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(43949)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(43950)
	}
	__antithesis_instrumentation__.Notify(43941)
	physicalFileName := l.File.Name()

	if err := ctx.Err(); err != nil {
		__antithesis_instrumentation__.Notify(43951)
		l.Printf("(note: incoming context was canceled: %s", err)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(43952)
	}
	__antithesis_instrumentation__.Notify(43942)

	l.Printf("running %s on nodes: %v", strings.Join(args, " "), nodes)
	if testLogger != nil {
		__antithesis_instrumentation__.Notify(43953)
		testLogger.Printf("> %s\n", strings.Join(args, " "))
	} else {
		__antithesis_instrumentation__.Notify(43954)
	}
	__antithesis_instrumentation__.Notify(43943)

	results, err := roachprod.RunWithDetails(ctx, l, c.MakeNodes(nodes), "", "", false, args)
	if err != nil {
		__antithesis_instrumentation__.Notify(43955)
		l.Printf("> result: %+v", err)
		createFailedFile(physicalFileName)
		return results, err
	} else {
		__antithesis_instrumentation__.Notify(43956)
	}
	__antithesis_instrumentation__.Notify(43944)

	for _, result := range results {
		__antithesis_instrumentation__.Notify(43957)
		if result.Err != nil {
			__antithesis_instrumentation__.Notify(43958)
			err = result.Err
			l.Printf("> Error for Node %d: %+v", int(result.Node), result.Err)
		} else {
			__antithesis_instrumentation__.Notify(43959)
		}
	}
	__antithesis_instrumentation__.Notify(43945)
	if err != nil {
		__antithesis_instrumentation__.Notify(43960)
		createFailedFile(physicalFileName)
	} else {
		__antithesis_instrumentation__.Notify(43961)
	}
	__antithesis_instrumentation__.Notify(43946)
	l.Close()
	return results, nil
}

func createFailedFile(logFileName string) {
	__antithesis_instrumentation__.Notify(43962)
	failedPhysicalFileName := strings.TrimSuffix(logFileName, ".log") + ".failed"
	if failedFile, err2 := os.Create(failedPhysicalFileName); err2 != nil {
		__antithesis_instrumentation__.Notify(43963)
		failedFile.Close()
	} else {
		__antithesis_instrumentation__.Notify(43964)
	}
}

func (c *clusterImpl) Reformat(
	ctx context.Context, l *logger.Logger, node option.NodeListOption, filesystem string,
) error {
	__antithesis_instrumentation__.Notify(43965)
	return roachprod.Reformat(ctx, l, c.name, filesystem)
}

var _ = (&clusterImpl{}).Reformat

func (c *clusterImpl) Install(
	ctx context.Context, l *logger.Logger, nodes option.NodeListOption, software ...string,
) error {
	__antithesis_instrumentation__.Notify(43966)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(43969)
		return errors.Wrap(ctx.Err(), "cluster.Install")
	} else {
		__antithesis_instrumentation__.Notify(43970)
	}
	__antithesis_instrumentation__.Notify(43967)
	if len(software) == 0 {
		__antithesis_instrumentation__.Notify(43971)
		return errors.New("Error running cluster.Install: no software passed")
	} else {
		__antithesis_instrumentation__.Notify(43972)
	}
	__antithesis_instrumentation__.Notify(43968)
	return errors.Wrap(roachprod.Install(ctx, l, c.MakeNodes(nodes), software), "cluster.Install")
}

var reOnlyAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func cmdLogFileName(t time.Time, nodes option.NodeListOption, args ...string) string {
	__antithesis_instrumentation__.Notify(43973)

	args = strings.Split(strings.Join(args, " "), " ")
	prefix := []string{reOnlyAlphanumeric.ReplaceAllString(args[0], "")}
	for _, arg := range args[1:] {
		__antithesis_instrumentation__.Notify(43976)
		if s := reOnlyAlphanumeric.ReplaceAllString(arg, ""); s != arg {
			__antithesis_instrumentation__.Notify(43978)
			break
		} else {
			__antithesis_instrumentation__.Notify(43979)
		}
		__antithesis_instrumentation__.Notify(43977)
		prefix = append(prefix, arg)
	}
	__antithesis_instrumentation__.Notify(43974)
	s := strings.Join(prefix, "_")
	const maxLen = 70
	if len(s) > maxLen {
		__antithesis_instrumentation__.Notify(43980)
		s = s[:maxLen]
	} else {
		__antithesis_instrumentation__.Notify(43981)
	}
	__antithesis_instrumentation__.Notify(43975)
	logFile := fmt.Sprintf(
		"run_%s_n%s_%s",
		t.Format(`150405.000000000`),
		nodes.String()[1:],
		s,
	)
	return logFile
}

func (c *clusterImpl) loggerForCmd(
	node option.NodeListOption, args ...string,
) (*logger.Logger, string, error) {
	__antithesis_instrumentation__.Notify(43982)
	logFile := cmdLogFileName(timeutil.Now(), node, args...)

	l, err := c.l.ChildLogger(logFile, logger.QuietStderr, logger.QuietStdout)
	if err != nil {
		__antithesis_instrumentation__.Notify(43984)
		return nil, "", err
	} else {
		__antithesis_instrumentation__.Notify(43985)
	}
	__antithesis_instrumentation__.Notify(43983)
	return l, logFile, nil
}

func (c *clusterImpl) pgURLErr(
	ctx context.Context, l *logger.Logger, node option.NodeListOption, external bool,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(43986)
	urls, err := roachprod.PgURL(ctx, l, c.MakeNodes(node), c.localCertsDir, external, c.localCertsDir != "")
	if err != nil {
		__antithesis_instrumentation__.Notify(43989)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(43990)
	}
	__antithesis_instrumentation__.Notify(43987)
	for i, url := range urls {
		__antithesis_instrumentation__.Notify(43991)
		urls[i] = strings.Trim(url, "'")
	}
	__antithesis_instrumentation__.Notify(43988)
	return urls, nil
}

func (c *clusterImpl) InternalPGUrl(
	ctx context.Context, l *logger.Logger, node option.NodeListOption,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(43992)
	return c.pgURLErr(ctx, l, node, false)
}

var _ = (&clusterImpl{}).InternalPGUrl

func (c *clusterImpl) ExternalPGUrl(
	ctx context.Context, l *logger.Logger, node option.NodeListOption,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(43993)
	return c.pgURLErr(ctx, l, node, true)
}

func addrToAdminUIAddr(c *clusterImpl, addr string) (string, error) {
	__antithesis_instrumentation__.Notify(43994)
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		__antithesis_instrumentation__.Notify(43997)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(43998)
	}
	__antithesis_instrumentation__.Notify(43995)
	webPort, err := strconv.Atoi(port)
	if err != nil {
		__antithesis_instrumentation__.Notify(43999)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(44000)
	}
	__antithesis_instrumentation__.Notify(43996)

	return fmt.Sprintf("%s:%d", host, webPort+1), nil
}

func urlToAddr(pgURL string) (string, error) {
	__antithesis_instrumentation__.Notify(44001)
	u, err := url.Parse(pgURL)
	if err != nil {
		__antithesis_instrumentation__.Notify(44003)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(44004)
	}
	__antithesis_instrumentation__.Notify(44002)
	return u.Host, nil
}

func addrToHost(addr string) (string, error) {
	__antithesis_instrumentation__.Notify(44005)
	host, _, err := addrToHostPort(addr)
	if err != nil {
		__antithesis_instrumentation__.Notify(44007)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(44008)
	}
	__antithesis_instrumentation__.Notify(44006)
	return host, nil
}

func addrToHostPort(addr string) (string, int, error) {
	__antithesis_instrumentation__.Notify(44009)
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		__antithesis_instrumentation__.Notify(44012)
		return "", 0, err
	} else {
		__antithesis_instrumentation__.Notify(44013)
	}
	__antithesis_instrumentation__.Notify(44010)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(44014)
		return "", 0, err
	} else {
		__antithesis_instrumentation__.Notify(44015)
	}
	__antithesis_instrumentation__.Notify(44011)
	return host, port, nil
}

func (c *clusterImpl) InternalAdminUIAddr(
	ctx context.Context, l *logger.Logger, node option.NodeListOption,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(44016)
	var addrs []string
	urls, err := c.InternalAddr(ctx, l, node)
	if err != nil {
		__antithesis_instrumentation__.Notify(44019)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44020)
	}
	__antithesis_instrumentation__.Notify(44017)
	for _, u := range urls {
		__antithesis_instrumentation__.Notify(44021)
		adminUIAddr, err := addrToAdminUIAddr(c, u)
		if err != nil {
			__antithesis_instrumentation__.Notify(44023)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(44024)
		}
		__antithesis_instrumentation__.Notify(44022)
		addrs = append(addrs, adminUIAddr)
	}
	__antithesis_instrumentation__.Notify(44018)
	return addrs, nil
}

func (c *clusterImpl) ExternalAdminUIAddr(
	ctx context.Context, l *logger.Logger, node option.NodeListOption,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(44025)
	var addrs []string
	externalAddrs, err := c.ExternalAddr(ctx, l, node)
	if err != nil {
		__antithesis_instrumentation__.Notify(44028)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44029)
	}
	__antithesis_instrumentation__.Notify(44026)
	for _, u := range externalAddrs {
		__antithesis_instrumentation__.Notify(44030)
		adminUIAddr, err := addrToAdminUIAddr(c, u)
		if err != nil {
			__antithesis_instrumentation__.Notify(44032)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(44033)
		}
		__antithesis_instrumentation__.Notify(44031)
		addrs = append(addrs, adminUIAddr)
	}
	__antithesis_instrumentation__.Notify(44027)
	return addrs, nil
}

func (c *clusterImpl) InternalAddr(
	ctx context.Context, l *logger.Logger, node option.NodeListOption,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(44034)
	var addrs []string
	urls, err := c.pgURLErr(ctx, l, node, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(44037)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44038)
	}
	__antithesis_instrumentation__.Notify(44035)
	for _, u := range urls {
		__antithesis_instrumentation__.Notify(44039)
		addr, err := urlToAddr(u)
		if err != nil {
			__antithesis_instrumentation__.Notify(44041)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(44042)
		}
		__antithesis_instrumentation__.Notify(44040)
		addrs = append(addrs, addr)
	}
	__antithesis_instrumentation__.Notify(44036)
	return addrs, nil
}

func (c *clusterImpl) InternalIP(
	ctx context.Context, l *logger.Logger, node option.NodeListOption,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(44043)
	var ips []string
	addrs, err := c.InternalAddr(ctx, l, node)
	if err != nil {
		__antithesis_instrumentation__.Notify(44046)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44047)
	}
	__antithesis_instrumentation__.Notify(44044)
	for _, addr := range addrs {
		__antithesis_instrumentation__.Notify(44048)
		host, err := addrToHost(addr)
		if err != nil {
			__antithesis_instrumentation__.Notify(44050)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(44051)
		}
		__antithesis_instrumentation__.Notify(44049)
		ips = append(ips, host)
	}
	__antithesis_instrumentation__.Notify(44045)
	return ips, nil
}

func (c *clusterImpl) ExternalAddr(
	ctx context.Context, l *logger.Logger, node option.NodeListOption,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(44052)
	var addrs []string
	urls, err := c.pgURLErr(ctx, l, node, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(44055)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44056)
	}
	__antithesis_instrumentation__.Notify(44053)
	for _, u := range urls {
		__antithesis_instrumentation__.Notify(44057)
		addr, err := urlToAddr(u)
		if err != nil {
			__antithesis_instrumentation__.Notify(44059)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(44060)
		}
		__antithesis_instrumentation__.Notify(44058)
		addrs = append(addrs, addr)
	}
	__antithesis_instrumentation__.Notify(44054)
	return addrs, nil
}

func (c *clusterImpl) ExternalIP(
	ctx context.Context, l *logger.Logger, node option.NodeListOption,
) ([]string, error) {
	__antithesis_instrumentation__.Notify(44061)
	var ips []string
	addrs, err := c.ExternalAddr(ctx, l, node)
	if err != nil {
		__antithesis_instrumentation__.Notify(44064)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44065)
	}
	__antithesis_instrumentation__.Notify(44062)
	for _, addr := range addrs {
		__antithesis_instrumentation__.Notify(44066)
		host, err := addrToHost(addr)
		if err != nil {
			__antithesis_instrumentation__.Notify(44068)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(44069)
		}
		__antithesis_instrumentation__.Notify(44067)
		ips = append(ips, host)
	}
	__antithesis_instrumentation__.Notify(44063)
	return ips, nil
}

var _ = (&clusterImpl{}).ExternalIP

func (c *clusterImpl) Conn(ctx context.Context, l *logger.Logger, node int) *gosql.DB {
	__antithesis_instrumentation__.Notify(44070)
	urls, err := c.ExternalPGUrl(ctx, l, c.Node(node))
	if err != nil {
		__antithesis_instrumentation__.Notify(44073)
		c.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(44074)
	}
	__antithesis_instrumentation__.Notify(44071)
	db, err := gosql.Open("postgres", urls[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(44075)
		c.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(44076)
	}
	__antithesis_instrumentation__.Notify(44072)
	return db
}

func (c *clusterImpl) ConnE(ctx context.Context, l *logger.Logger, node int) (*gosql.DB, error) {
	__antithesis_instrumentation__.Notify(44077)
	urls, err := c.ExternalPGUrl(ctx, l, c.Node(node))
	if err != nil {
		__antithesis_instrumentation__.Notify(44080)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44081)
	}
	__antithesis_instrumentation__.Notify(44078)
	db, err := gosql.Open("postgres", urls[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(44082)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44083)
	}
	__antithesis_instrumentation__.Notify(44079)
	return db, nil
}

func (c *clusterImpl) ConnEAsUser(
	ctx context.Context, l *logger.Logger, node int, user string,
) (*gosql.DB, error) {
	__antithesis_instrumentation__.Notify(44084)
	urls, err := c.ExternalPGUrl(ctx, l, c.Node(node))
	if err != nil {
		__antithesis_instrumentation__.Notify(44088)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44089)
	}
	__antithesis_instrumentation__.Notify(44085)

	u, err := url.Parse(urls[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(44090)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44091)
	}
	__antithesis_instrumentation__.Notify(44086)
	u.User = url.User(user)
	dataSourceName := u.String()
	db, err := gosql.Open("postgres", dataSourceName)
	if err != nil {
		__antithesis_instrumentation__.Notify(44092)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(44093)
	}
	__antithesis_instrumentation__.Notify(44087)
	return db, nil
}

func (c *clusterImpl) MakeNodes(opts ...option.Option) string {
	__antithesis_instrumentation__.Notify(44094)
	var r option.NodeListOption
	for _, o := range opts {
		__antithesis_instrumentation__.Notify(44096)
		if s, ok := o.(nodeSelector); ok {
			__antithesis_instrumentation__.Notify(44097)
			r = s.Merge(r)
		} else {
			__antithesis_instrumentation__.Notify(44098)
		}
	}
	__antithesis_instrumentation__.Notify(44095)
	return c.name + r.String()
}

func (c *clusterImpl) IsLocal() bool {
	__antithesis_instrumentation__.Notify(44099)
	return c.name == "local"
}

func (c *clusterImpl) Extend(ctx context.Context, d time.Duration, l *logger.Logger) error {
	__antithesis_instrumentation__.Notify(44100)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(44103)
		return errors.Wrap(ctx.Err(), "cluster.Extend")
	} else {
		__antithesis_instrumentation__.Notify(44104)
	}
	__antithesis_instrumentation__.Notify(44101)
	l.PrintfCtx(ctx, "extending cluster by %s", d.String())
	if err := roachprod.Extend(l, c.name, d); err != nil {
		__antithesis_instrumentation__.Notify(44105)
		l.PrintfCtx(ctx, "roachprod extend failed: %v", err)
		return errors.Wrap(err, "roachprod extend failed")
	} else {
		__antithesis_instrumentation__.Notify(44106)
	}
	__antithesis_instrumentation__.Notify(44102)

	c.expiration = c.expiration.Add(d)
	return nil
}

func (c *clusterImpl) NewMonitor(ctx context.Context, opts ...option.Option) cluster.Monitor {
	__antithesis_instrumentation__.Notify(44107)
	return newMonitor(ctx, c.t, c, opts...)
}
