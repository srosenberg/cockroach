package install

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/roachprod/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	rperrors "github.com/cockroachdb/cockroach/pkg/roachprod/errors"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/ssh"
	"github.com/cockroachdb/cockroach/pkg/roachprod/ui"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/aws"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm/local"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
)

type SyncedCluster struct {
	cloud.Cluster

	Nodes Nodes

	ClusterSettings

	Localities []string

	AuthorizedKeys []byte
}

func NewSyncedCluster(
	metadata *cloud.Cluster, nodeSelector string, settings ClusterSettings,
) (*SyncedCluster, error) {
	__antithesis_instrumentation__.Notify(180465)
	c := &SyncedCluster{
		Cluster:         *metadata,
		ClusterSettings: settings,
	}
	c.Localities = make([]string, len(c.VMs))
	for i := range c.VMs {
		__antithesis_instrumentation__.Notify(180468)
		locality, err := c.VMs[i].Locality()
		if err != nil {
			__antithesis_instrumentation__.Notify(180471)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180472)
		}
		__antithesis_instrumentation__.Notify(180469)
		if c.NumRacks > 0 {
			__antithesis_instrumentation__.Notify(180473)
			if locality != "" {
				__antithesis_instrumentation__.Notify(180475)
				locality += ","
			} else {
				__antithesis_instrumentation__.Notify(180476)
			}
			__antithesis_instrumentation__.Notify(180474)
			locality += fmt.Sprintf("rack=%d", i%c.NumRacks)
		} else {
			__antithesis_instrumentation__.Notify(180477)
		}
		__antithesis_instrumentation__.Notify(180470)
		c.Localities[i] = locality
	}
	__antithesis_instrumentation__.Notify(180466)

	nodes, err := ListNodes(nodeSelector, len(c.VMs))
	if err != nil {
		__antithesis_instrumentation__.Notify(180478)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(180479)
	}
	__antithesis_instrumentation__.Notify(180467)
	c.Nodes = nodes
	return c, nil
}

func (c *SyncedCluster) Host(n Node) string {
	__antithesis_instrumentation__.Notify(180480)
	return c.VMs[n-1].PublicIP
}

func (c *SyncedCluster) user(n Node) string {
	__antithesis_instrumentation__.Notify(180481)
	return c.VMs[n-1].RemoteUser
}

func (c *SyncedCluster) locality(n Node) string {
	__antithesis_instrumentation__.Notify(180482)
	return c.Localities[n-1]
}

func (c *SyncedCluster) IsLocal() bool {
	__antithesis_instrumentation__.Notify(180483)
	return config.IsLocalClusterName(c.Name)
}

func (c *SyncedCluster) localVMDir(n Node) string {
	__antithesis_instrumentation__.Notify(180484)
	return local.VMDir(c.Name, int(n))
}

func (c *SyncedCluster) TargetNodes() Nodes {
	__antithesis_instrumentation__.Notify(180485)
	return append(Nodes{}, c.Nodes...)
}

func (c *SyncedCluster) GetInternalIP(ctx context.Context, n Node) (string, error) {
	__antithesis_instrumentation__.Notify(180486)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(180491)
		return c.Host(n), nil
	} else {
		__antithesis_instrumentation__.Notify(180492)
	}
	__antithesis_instrumentation__.Notify(180487)

	session, err := c.newSession(n)
	if err != nil {
		__antithesis_instrumentation__.Notify(180493)
		return "", errors.Wrapf(err, "GetInternalIP: failed dial %s:%d", c.Name, n)
	} else {
		__antithesis_instrumentation__.Notify(180494)
	}
	__antithesis_instrumentation__.Notify(180488)
	defer session.Close()

	var stdout, stderr strings.Builder
	session.SetStdout(&stdout)
	session.SetStderr(&stderr)
	cmd := `hostname --all-ip-addresses`
	if err := session.Run(ctx, cmd); err != nil {
		__antithesis_instrumentation__.Notify(180495)
		return "", errors.Wrapf(err,
			"GetInternalIP: failed to execute hostname on %s:%d:\n(stdout) %s\n(stderr) %s",
			c.Name, n, stdout.String(), stderr.String())
	} else {
		__antithesis_instrumentation__.Notify(180496)
	}
	__antithesis_instrumentation__.Notify(180489)
	ip := strings.TrimSpace(stdout.String())
	if ip == "" {
		__antithesis_instrumentation__.Notify(180497)
		return "", errors.Errorf(
			"empty internal IP returned, stdout:\n%s\nstderr:\n%s",
			stdout.String(), stderr.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(180498)
	}
	__antithesis_instrumentation__.Notify(180490)
	return ip, nil
}

func (c *SyncedCluster) roachprodEnvValue(node Node) string {
	__antithesis_instrumentation__.Notify(180499)
	var parts []string
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(180502)
		parts = append(parts, c.Name)
	} else {
		__antithesis_instrumentation__.Notify(180503)
	}
	__antithesis_instrumentation__.Notify(180500)
	parts = append(parts, fmt.Sprintf("%d", node))
	if c.Tag != "" {
		__antithesis_instrumentation__.Notify(180504)
		parts = append(parts, c.Tag)
	} else {
		__antithesis_instrumentation__.Notify(180505)
	}
	__antithesis_instrumentation__.Notify(180501)
	return strings.Join(parts, "/")
}

func (c *SyncedCluster) roachprodEnvRegex(node Node) string {
	__antithesis_instrumentation__.Notify(180506)
	escaped := strings.Replace(c.roachprodEnvValue(node), "/", "\\/", -1)

	return fmt.Sprintf(`ROACHPROD=%s[ \/]`, escaped)
}

func (c *SyncedCluster) newSession(node Node) (session, error) {
	__antithesis_instrumentation__.Notify(180507)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(180509)
		return newLocalSession(), nil
	} else {
		__antithesis_instrumentation__.Notify(180510)
	}
	__antithesis_instrumentation__.Notify(180508)
	return newRemoteSession(c.user(node), c.Host(node), c.DebugDir)
}

func (c *SyncedCluster) Stop(ctx context.Context, l *logger.Logger, sig int, wait bool) error {
	__antithesis_instrumentation__.Notify(180511)
	display := fmt.Sprintf("%s: stopping", c.Name)
	if wait {
		__antithesis_instrumentation__.Notify(180513)
		display += " and waiting"
	} else {
		__antithesis_instrumentation__.Notify(180514)
	}
	__antithesis_instrumentation__.Notify(180512)
	return c.Parallel(l, display, len(c.Nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180515)
		sess, err := c.newSession(c.Nodes[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(180518)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180519)
		}
		__antithesis_instrumentation__.Notify(180516)
		defer sess.Close()

		var waitCmd string
		if wait {
			__antithesis_instrumentation__.Notify(180520)
			waitCmd = fmt.Sprintf(`
  for pid in ${pids}; do
    echo "${pid}: checking" >> %[1]s/roachprod.log
    while kill -0 ${pid}; do
      kill -0 ${pid} >> %[1]s/roachprod.log 2>&1
      echo "${pid}: still alive [$?]" >> %[1]s/roachprod.log
      ps axeww -o pid -o command >> %[1]s/roachprod.log
      sleep 1
    done
    echo "${pid}: dead" >> %[1]s/roachprod.log
  done`,
				c.LogDir(c.Nodes[i]),
			)
		} else {
			__antithesis_instrumentation__.Notify(180521)
		}
		__antithesis_instrumentation__.Notify(180517)

		cmd := fmt.Sprintf(`
mkdir -p %[1]s
echo ">>> roachprod stop: $(date)" >> %[1]s/roachprod.log
ps axeww -o pid -o command >> %[1]s/roachprod.log
pids=$(ps axeww -o pid -o command | \
  sed 's/export ROACHPROD=//g' | \
  awk '/%[2]s/ { print $1 }')
if [ -n "${pids}" ]; then
  kill -%[3]d ${pids}
%[4]s
fi`,
			c.LogDir(c.Nodes[i]),
			c.roachprodEnvRegex(c.Nodes[i]),
			sig,
			waitCmd,
		)
		return sess.CombinedOutput(ctx, cmd)
	})
}

func (c *SyncedCluster) Wipe(ctx context.Context, l *logger.Logger, preserveCerts bool) error {
	__antithesis_instrumentation__.Notify(180522)
	display := fmt.Sprintf("%s: wiping", c.Name)
	if err := c.Stop(ctx, l, 9, true); err != nil {
		__antithesis_instrumentation__.Notify(180524)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180525)
	}
	__antithesis_instrumentation__.Notify(180523)
	return c.Parallel(l, display, len(c.Nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180526)
		sess, err := c.newSession(c.Nodes[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(180529)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180530)
		}
		__antithesis_instrumentation__.Notify(180527)
		defer sess.Close()

		var cmd string
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(180531)

			dirs := []string{"data", "logs"}
			if !preserveCerts {
				__antithesis_instrumentation__.Notify(180533)
				dirs = append(dirs, "certs*")
			} else {
				__antithesis_instrumentation__.Notify(180534)
			}
			__antithesis_instrumentation__.Notify(180532)
			for _, dir := range dirs {
				__antithesis_instrumentation__.Notify(180535)
				cmd += fmt.Sprintf(`rm -fr %s/%s ;`, c.localVMDir(c.Nodes[i]), dir)
			}
		} else {
			__antithesis_instrumentation__.Notify(180536)
			cmd = `sudo find /mnt/data* -maxdepth 1 -type f -exec rm -f {} \; &&
sudo rm -fr /mnt/data*/{auxiliary,local,tmp,cassandra,cockroach,cockroach-temp*,mongo-data} &&
sudo rm -fr logs &&
`
			if !preserveCerts {
				__antithesis_instrumentation__.Notify(180537)
				cmd += "sudo rm -fr certs* ;\n"
			} else {
				__antithesis_instrumentation__.Notify(180538)
			}
		}
		__antithesis_instrumentation__.Notify(180528)
		return sess.CombinedOutput(ctx, cmd)
	})
}

func (c *SyncedCluster) Status(ctx context.Context, l *logger.Logger) error {
	__antithesis_instrumentation__.Notify(180539)
	display := fmt.Sprintf("%s: status", c.Name)
	results := make([]string, len(c.Nodes))
	if err := c.Parallel(l, display, len(c.Nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180542)
		sess, err := c.newSession(c.Nodes[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(180546)
			results[i] = err.Error()
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(180547)
		}
		__antithesis_instrumentation__.Notify(180543)
		defer sess.Close()

		binary := cockroachNodeBinary(c, c.Nodes[i])
		cmd := fmt.Sprintf(`out=$(ps axeww -o pid -o ucomm -o command | \
  sed 's/export ROACHPROD=//g' | \
  awk '/%s/ {print $2, $1}'`,
			c.roachprodEnvRegex(c.Nodes[i]))
		cmd += ` | sort | uniq);
vers=$(` + binary + ` version 2>/dev/null | awk '/Build Tag:/ {print $NF}')
if [ -n "${out}" -a -n "${vers}" ]; then
  echo ${out} | sed "s/cockroach/cockroach-${vers}/g"
else
  echo ${out}
fi
`
		out, err := sess.CombinedOutput(ctx, cmd)
		var msg string
		if err != nil {
			__antithesis_instrumentation__.Notify(180548)
			return nil, errors.Wrapf(err, "~ %s\n%s", cmd, out)
		} else {
			__antithesis_instrumentation__.Notify(180549)
		}
		__antithesis_instrumentation__.Notify(180544)
		msg = strings.TrimSpace(string(out))
		if msg == "" {
			__antithesis_instrumentation__.Notify(180550)
			msg = "not running"
		} else {
			__antithesis_instrumentation__.Notify(180551)
		}
		__antithesis_instrumentation__.Notify(180545)
		results[i] = msg
		return nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(180552)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180553)
	}
	__antithesis_instrumentation__.Notify(180540)
	for i, r := range results {
		__antithesis_instrumentation__.Notify(180554)
		l.Printf("  %2d: %s\n", c.Nodes[i], r)
	}
	__antithesis_instrumentation__.Notify(180541)
	return nil
}

type NodeMonitorInfo struct {
	Node Node

	Msg string

	Err error
}

type MonitorOpts struct {
	OneShot          bool
	IgnoreEmptyNodes bool
}

func (c *SyncedCluster) Monitor(ctx context.Context, opts MonitorOpts) chan NodeMonitorInfo {
	__antithesis_instrumentation__.Notify(180555)
	ch := make(chan NodeMonitorInfo)
	nodes := c.TargetNodes()
	var wg sync.WaitGroup

	for i := range nodes {
		__antithesis_instrumentation__.Notify(180558)
		wg.Add(1)
		go func(i int) {
			__antithesis_instrumentation__.Notify(180559)
			defer wg.Done()
			sess, err := c.newSession(nodes[i])
			if err != nil {
				__antithesis_instrumentation__.Notify(180567)
				ch <- NodeMonitorInfo{Node: nodes[i], Err: err}
				wg.Done()
				return
			} else {
				__antithesis_instrumentation__.Notify(180568)
			}
			__antithesis_instrumentation__.Notify(180560)
			defer sess.Close()

			p, err := sess.StdoutPipe()
			if err != nil {
				__antithesis_instrumentation__.Notify(180569)
				ch <- NodeMonitorInfo{Node: nodes[i], Err: err}
				wg.Done()
				return
			} else {
				__antithesis_instrumentation__.Notify(180570)
			}
			__antithesis_instrumentation__.Notify(180561)

			data := struct {
				OneShot     bool
				IgnoreEmpty bool
				Store       string
				Port        int
				Local       bool
			}{
				OneShot:     opts.OneShot,
				IgnoreEmpty: opts.IgnoreEmptyNodes,
				Store:       c.NodeDir(nodes[i], 1),
				Port:        c.NodePort(nodes[i]),
				Local:       c.IsLocal(),
			}

			snippet := `
{{ if .IgnoreEmpty }}
if [ ! -f "{{.Store}}/CURRENT" ]; then
  echo "skipped"
  exit 0
fi
{{- end}}
# Init with -1 so that when cockroach is initially dead, we print
# a dead event for it.
lastpid=-1
while :; do
{{ if .Local }}
  pid=$(lsof -i :{{.Port}} -sTCP:LISTEN | awk '!/COMMAND/ {print $2}')
	pid=${pid:-0} # default to 0
	status="unknown"
{{- else }}
  # When CRDB is not running, this is zero.
	pid=$(systemctl show cockroach --property MainPID --value)
	status=$(systemctl show cockroach --property ExecMainStatus --value)
{{- end }}
  if [[ "${lastpid}" == -1 && "${pid}" != 0 ]]; then
    # On the first iteration through the loop, if the process is running,
    # don't register a PID change (which would trigger an erroneous dead
    # event).
    lastpid=0
  fi
  # Output a dead event whenever the PID changes from a nonzero value to
  # any other value. In particular, we emit a dead event when the node stops
  # (lastpid is nonzero, pid is zero), but not when the process then starts
  # again (lastpid is zero, pid is nonzero).
  if [ "${pid}" != "${lastpid}" ]; then
    if [ "${lastpid}" != 0 ]; then
      if [ "${pid}" != 0 ]; then
        # If the PID changed but neither is zero, then the status refers to
        # the new incarnation. We lost the actual exit status of the old PID.
        status="unknown"
      fi
    	echo "dead (exit status ${status})"
    fi
		if [ "${pid}" != 0 ]; then
			echo "${pid}"
    fi
    lastpid=${pid}
  fi
{{ if .OneShot }}
  exit 0
{{- end }}
  sleep 1
  if [ "${pid}" != 0 ]; then
    while kill -0 "${pid}"; do
      sleep 1
    done
  fi
done
`

			t := template.Must(template.New("script").Parse(snippet))
			var buf bytes.Buffer
			if err := t.Execute(&buf, data); err != nil {
				__antithesis_instrumentation__.Notify(180571)
				ch <- NodeMonitorInfo{Node: nodes[i], Err: err}
				return
			} else {
				__antithesis_instrumentation__.Notify(180572)
			}
			__antithesis_instrumentation__.Notify(180562)

			if err := sess.RequestPty(); err != nil {
				__antithesis_instrumentation__.Notify(180573)
				ch <- NodeMonitorInfo{Node: nodes[i], Err: err}
				return
			} else {
				__antithesis_instrumentation__.Notify(180574)
			}
			__antithesis_instrumentation__.Notify(180563)

			var readerWg sync.WaitGroup
			readerWg.Add(1)
			go func(p io.Reader) {
				__antithesis_instrumentation__.Notify(180575)
				defer readerWg.Done()
				r := bufio.NewReader(p)
				for {
					__antithesis_instrumentation__.Notify(180576)
					line, _, err := r.ReadLine()
					if err == io.EOF {
						__antithesis_instrumentation__.Notify(180578)
						return
					} else {
						__antithesis_instrumentation__.Notify(180579)
					}
					__antithesis_instrumentation__.Notify(180577)
					ch <- NodeMonitorInfo{Node: nodes[i], Msg: string(line)}
				}
			}(p)
			__antithesis_instrumentation__.Notify(180564)

			if err := sess.Start(buf.String()); err != nil {
				__antithesis_instrumentation__.Notify(180580)
				ch <- NodeMonitorInfo{Node: nodes[i], Err: err}
				return
			} else {
				__antithesis_instrumentation__.Notify(180581)
			}
			__antithesis_instrumentation__.Notify(180565)

			go func() {
				__antithesis_instrumentation__.Notify(180582)
				<-ctx.Done()
				sess.Close()
			}()
			__antithesis_instrumentation__.Notify(180566)

			readerWg.Wait()

			if err := sess.Wait(); err != nil {
				__antithesis_instrumentation__.Notify(180583)
				ch <- NodeMonitorInfo{Node: nodes[i], Err: err}
				return
			} else {
				__antithesis_instrumentation__.Notify(180584)
			}
		}(i)
	}
	__antithesis_instrumentation__.Notify(180556)
	go func() {
		__antithesis_instrumentation__.Notify(180585)
		wg.Wait()
		close(ch)
	}()
	__antithesis_instrumentation__.Notify(180557)

	return ch
}

type RunResultDetails struct {
	Node             Node
	Stdout           string
	Stderr           string
	Err              error
	RemoteExitStatus string
}

func processStdout(stdout string) (string, string) {
	__antithesis_instrumentation__.Notify(180586)
	retStdout := stdout
	exitStatusPattern := "LAST EXIT STATUS: "
	exitStatusIndex := strings.LastIndex(retStdout, exitStatusPattern)
	remoteExitStatus := "-1"

	if exitStatusIndex != -1 {
		__antithesis_instrumentation__.Notify(180588)
		retStdout = stdout[:exitStatusIndex]
		remoteExitStatus = strings.TrimSpace(stdout[exitStatusIndex+len(exitStatusPattern):])
	} else {
		__antithesis_instrumentation__.Notify(180589)
	}
	__antithesis_instrumentation__.Notify(180587)
	return retStdout, remoteExitStatus
}

func runCmdOnSingleNode(
	ctx context.Context, l *logger.Logger, c *SyncedCluster, node Node, cmd string,
) (RunResultDetails, error) {
	__antithesis_instrumentation__.Notify(180590)
	result := RunResultDetails{Node: node}
	sess, err := c.newSession(node)
	if err != nil {
		__antithesis_instrumentation__.Notify(180595)
		return result, err
	} else {
		__antithesis_instrumentation__.Notify(180596)
	}
	__antithesis_instrumentation__.Notify(180591)
	defer sess.Close()

	sess.SetWithExitStatus(true)
	var stdoutBuffer, stderrBuffer bytes.Buffer
	sess.SetStdout(&stdoutBuffer)
	sess.SetStderr(&stderrBuffer)

	e := expander{
		node: node,
	}
	expandedCmd, err := e.expand(ctx, l, c, cmd)
	if err != nil {
		__antithesis_instrumentation__.Notify(180597)
		return result, err
	} else {
		__antithesis_instrumentation__.Notify(180598)
	}
	__antithesis_instrumentation__.Notify(180592)

	nodeCmd := fmt.Sprintf(`export ROACHPROD=%s GOTRACEBACK=crash && bash -c %s`,
		c.roachprodEnvValue(node), ssh.Escape1(expandedCmd))
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(180599)
		nodeCmd = fmt.Sprintf("cd %s; %s", c.localVMDir(node), nodeCmd)
	} else {
		__antithesis_instrumentation__.Notify(180600)
	}
	__antithesis_instrumentation__.Notify(180593)

	err = sess.Run(ctx, nodeCmd)
	result.Stderr = stderrBuffer.String()
	result.Stdout, result.RemoteExitStatus = processStdout(stdoutBuffer.String())

	if err != nil {
		__antithesis_instrumentation__.Notify(180601)
		detailMsg := fmt.Sprintf("Node %d. Command with error:\n```\n%s\n```\n", node, cmd)
		err = errors.WithDetail(err, detailMsg)
		err = rperrors.ClassifyCmdError(err)
		if reflect.TypeOf(err) == reflect.TypeOf(rperrors.SSH{}) {
			__antithesis_instrumentation__.Notify(180603)
			result.RemoteExitStatus = "255"
		} else {
			__antithesis_instrumentation__.Notify(180604)
		}
		__antithesis_instrumentation__.Notify(180602)
		result.Err = err
	} else {
		__antithesis_instrumentation__.Notify(180605)
		if result.RemoteExitStatus != "0" {
			__antithesis_instrumentation__.Notify(180606)
			result.Err = &NonZeroExitCode{fmt.Sprintf("Non-zero exit code: %s", result.RemoteExitStatus)}
		} else {
			__antithesis_instrumentation__.Notify(180607)
		}
	}
	__antithesis_instrumentation__.Notify(180594)
	return result, nil
}

type NonZeroExitCode struct {
	message string
}

func (e *NonZeroExitCode) Error() string {
	__antithesis_instrumentation__.Notify(180608)
	return e.message
}

func (c *SyncedCluster) Run(
	ctx context.Context, l *logger.Logger, stdout, stderr io.Writer, nodes Nodes, title, cmd string,
) error {
	__antithesis_instrumentation__.Notify(180609)

	stream := len(nodes) == 1
	var display string
	if !stream {
		__antithesis_instrumentation__.Notify(180613)
		display = fmt.Sprintf("%s: %s", c.Name, title)
	} else {
		__antithesis_instrumentation__.Notify(180614)
	}
	__antithesis_instrumentation__.Notify(180610)

	errs := make([]error, len(nodes))
	results := make([]string, len(nodes))
	if err := c.Parallel(l, display, len(nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180615)
		sess, err := c.newSession(nodes[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(180621)
			errs[i] = err
			results[i] = err.Error()
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(180622)
		}
		__antithesis_instrumentation__.Notify(180616)
		defer sess.Close()

		e := expander{
			node: nodes[i],
		}
		expandedCmd, err := e.expand(ctx, l, c, cmd)
		if err != nil {
			__antithesis_instrumentation__.Notify(180623)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180624)
		}
		__antithesis_instrumentation__.Notify(180617)

		nodeCmd := fmt.Sprintf(`export ROACHPROD=%s GOTRACEBACK=crash && bash -c %s`,
			c.roachprodEnvValue(nodes[i]), ssh.Escape1(expandedCmd))
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(180625)
			nodeCmd = fmt.Sprintf("cd %s; %s", c.localVMDir(nodes[i]), nodeCmd)
		} else {
			__antithesis_instrumentation__.Notify(180626)
		}
		__antithesis_instrumentation__.Notify(180618)

		if stream {
			__antithesis_instrumentation__.Notify(180627)
			sess.SetStdout(stdout)
			sess.SetStderr(stderr)
			errs[i] = sess.Run(ctx, nodeCmd)
			if errs[i] != nil {
				__antithesis_instrumentation__.Notify(180629)
				detailMsg := fmt.Sprintf("Node %d. Command with error:\n```\n%s\n```\n", nodes[i], cmd)
				err = errors.WithDetail(errs[i], detailMsg)
				err = rperrors.ClassifyCmdError(err)
				errs[i] = err
			} else {
				__antithesis_instrumentation__.Notify(180630)
			}
			__antithesis_instrumentation__.Notify(180628)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(180631)
		}
		__antithesis_instrumentation__.Notify(180619)

		out, err := sess.CombinedOutput(ctx, nodeCmd)
		msg := strings.TrimSpace(string(out))
		if err != nil {
			__antithesis_instrumentation__.Notify(180632)
			detailMsg := fmt.Sprintf("Node %d. Command with error:\n```\n%s\n```\n", nodes[i], cmd)
			err = errors.WithDetail(err, detailMsg)
			err = rperrors.ClassifyCmdError(err)
			errs[i] = err
			msg += fmt.Sprintf("\n%v", err)
		} else {
			__antithesis_instrumentation__.Notify(180633)
		}
		__antithesis_instrumentation__.Notify(180620)
		results[i] = msg
		return nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(180634)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180635)
	}
	__antithesis_instrumentation__.Notify(180611)

	if !stream {
		__antithesis_instrumentation__.Notify(180636)
		for i, r := range results {
			__antithesis_instrumentation__.Notify(180637)
			fmt.Fprintf(stdout, "  %2d: %s\n", nodes[i], r)
		}
	} else {
		__antithesis_instrumentation__.Notify(180638)
	}
	__antithesis_instrumentation__.Notify(180612)
	return rperrors.SelectPriorityError(errs)
}

func (c *SyncedCluster) RunWithDetails(
	ctx context.Context, l *logger.Logger, nodes Nodes, title, cmd string,
) ([]RunResultDetails, error) {
	__antithesis_instrumentation__.Notify(180639)
	display := fmt.Sprintf("%s: %s", c.Name, title)
	results := make([]RunResultDetails, len(nodes))

	failed, err := c.ParallelE(l, display, len(nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180642)
		result, err := runCmdOnSingleNode(ctx, l, c, nodes[i], cmd)
		if err != nil {
			__antithesis_instrumentation__.Notify(180644)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180645)
		}
		__antithesis_instrumentation__.Notify(180643)
		results[i] = result
		return nil, nil
	})
	__antithesis_instrumentation__.Notify(180640)
	if err != nil {
		__antithesis_instrumentation__.Notify(180646)
		for _, node := range failed {
			__antithesis_instrumentation__.Notify(180647)
			results[node.Index].Err = node.Err
		}
	} else {
		__antithesis_instrumentation__.Notify(180648)
	}
	__antithesis_instrumentation__.Notify(180641)
	return results, nil
}

func (c *SyncedCluster) Wait(ctx context.Context, l *logger.Logger) error {
	__antithesis_instrumentation__.Notify(180649)
	display := fmt.Sprintf("%s: waiting for nodes to start", c.Name)
	errs := make([]error, len(c.Nodes))
	if err := c.Parallel(l, display, len(c.Nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180653)
		for j := 0; j < 600; j++ {
			__antithesis_instrumentation__.Notify(180655)
			sess, err := c.newSession(c.Nodes[i])
			if err != nil {
				__antithesis_instrumentation__.Notify(180658)
				time.Sleep(500 * time.Millisecond)
				continue
			} else {
				__antithesis_instrumentation__.Notify(180659)
			}
			__antithesis_instrumentation__.Notify(180656)
			defer sess.Close()

			_, err = sess.CombinedOutput(ctx, "test -e /mnt/data1/.roachprod-initialized")
			if err != nil {
				__antithesis_instrumentation__.Notify(180660)
				time.Sleep(500 * time.Millisecond)
				continue
			} else {
				__antithesis_instrumentation__.Notify(180661)
			}
			__antithesis_instrumentation__.Notify(180657)
			return nil, nil
		}
		__antithesis_instrumentation__.Notify(180654)
		errs[i] = errors.New("timed out after 5m")
		return nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(180662)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180663)
	}
	__antithesis_instrumentation__.Notify(180650)

	var foundErr bool
	for i, err := range errs {
		__antithesis_instrumentation__.Notify(180664)
		if err != nil {
			__antithesis_instrumentation__.Notify(180665)
			l.Printf("  %2d: %v", c.Nodes[i], err)
			foundErr = true
		} else {
			__antithesis_instrumentation__.Notify(180666)
		}
	}
	__antithesis_instrumentation__.Notify(180651)
	if foundErr {
		__antithesis_instrumentation__.Notify(180667)
		return errors.New("not all nodes booted successfully")
	} else {
		__antithesis_instrumentation__.Notify(180668)
	}
	__antithesis_instrumentation__.Notify(180652)
	return nil
}

func (c *SyncedCluster) SetupSSH(ctx context.Context, l *logger.Logger) error {
	__antithesis_instrumentation__.Notify(180669)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(180679)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(180680)
	}
	__antithesis_instrumentation__.Notify(180670)

	if len(c.Nodes) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(180681)
		return len(c.VMs) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(180682)
		return fmt.Errorf("%s: invalid cluster: nodes=%d hosts=%d",
			c.Name, len(c.Nodes), len(c.VMs))
	} else {
		__antithesis_instrumentation__.Notify(180683)
	}
	__antithesis_instrumentation__.Notify(180671)

	var sshTar []byte
	if err := c.Parallel(l, "generating ssh key", 1, 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180684)
		sess, err := c.newSession(1)
		if err != nil {
			__antithesis_instrumentation__.Notify(180687)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180688)
		}
		__antithesis_instrumentation__.Notify(180685)
		defer sess.Close()

		cmd := `
test -f .ssh/id_rsa || \
  (ssh-keygen -q -f .ssh/id_rsa -t rsa -N '' && \
   cat .ssh/id_rsa.pub >> .ssh/authorized_keys);
tar cf - .ssh/id_rsa .ssh/id_rsa.pub .ssh/authorized_keys
`

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		sess.SetStdout(&stdout)
		sess.SetStderr(&stderr)

		if err := sess.Run(ctx, cmd); err != nil {
			__antithesis_instrumentation__.Notify(180689)
			return nil, errors.Wrapf(err, "%s: stderr:\n%s", cmd, stderr.String())
		} else {
			__antithesis_instrumentation__.Notify(180690)
		}
		__antithesis_instrumentation__.Notify(180686)
		sshTar = stdout.Bytes()
		return nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(180691)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180692)
	}
	__antithesis_instrumentation__.Notify(180672)

	nodes := c.Nodes[1:]
	if err := c.Parallel(l, "distributing ssh key", len(nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180693)
		sess, err := c.newSession(nodes[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(180696)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180697)
		}
		__antithesis_instrumentation__.Notify(180694)
		defer sess.Close()

		sess.SetStdin(bytes.NewReader(sshTar))
		cmd := `tar xf -`
		if out, err := sess.CombinedOutput(ctx, cmd); err != nil {
			__antithesis_instrumentation__.Notify(180698)
			return nil, errors.Wrapf(err, "%s: output:\n%s", cmd, out)
		} else {
			__antithesis_instrumentation__.Notify(180699)
		}
		__antithesis_instrumentation__.Notify(180695)
		return nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(180700)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180701)
	}
	__antithesis_instrumentation__.Notify(180673)

	ips := make([]string, len(c.Nodes), len(c.Nodes)*2)
	if err := c.Parallel(l, "retrieving hosts", len(c.Nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180702)
		for j := 0; j < 20 && func() bool {
			__antithesis_instrumentation__.Notify(180705)
			return ips[i] == "" == true
		}() == true; j++ {
			__antithesis_instrumentation__.Notify(180706)
			var err error
			ips[i], err = c.GetInternalIP(ctx, c.Nodes[i])
			if err != nil {
				__antithesis_instrumentation__.Notify(180708)
				return nil, errors.Wrapf(err, "pgurls")
			} else {
				__antithesis_instrumentation__.Notify(180709)
			}
			__antithesis_instrumentation__.Notify(180707)
			time.Sleep(time.Second)
		}
		__antithesis_instrumentation__.Notify(180703)
		if ips[i] == "" {
			__antithesis_instrumentation__.Notify(180710)
			return nil, fmt.Errorf("retrieved empty IP address")
		} else {
			__antithesis_instrumentation__.Notify(180711)
		}
		__antithesis_instrumentation__.Notify(180704)
		return nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(180712)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180713)
	}
	__antithesis_instrumentation__.Notify(180674)

	for _, i := range c.Nodes {
		__antithesis_instrumentation__.Notify(180714)
		ips = append(ips, c.Host(i))
	}
	__antithesis_instrumentation__.Notify(180675)
	var knownHostsData []byte
	if err := c.Parallel(l, "scanning hosts", 1, 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180715)
		sess, err := c.newSession(c.Nodes[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(180718)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180719)
		}
		__antithesis_instrumentation__.Notify(180716)
		defer sess.Close()

		cmd := `
set -e
tmp="$(tempfile -d ~/.ssh -p 'roachprod' )"
on_exit() {
    rm -f "${tmp}"
}
trap on_exit EXIT
for i in {1..20}; do
  ssh-keyscan -T 60 -t rsa ` + strings.Join(ips, " ") + ` > "${tmp}"
  if [[ "$(wc < ${tmp} -l)" -eq "` + fmt.Sprint(len(ips)) + `" ]]; then
    [[ -f .ssh/known_hosts ]] && cat .ssh/known_hosts >> "${tmp}"
    sort -u < "${tmp}"
    exit 0
  fi
  sleep 1
done
exit 1
`
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		sess.SetStdout(&stdout)
		sess.SetStderr(&stderr)
		if err := sess.Run(ctx, cmd); err != nil {
			__antithesis_instrumentation__.Notify(180720)
			return nil, errors.Wrapf(err, "%s: stderr:\n%s", cmd, stderr.String())
		} else {
			__antithesis_instrumentation__.Notify(180721)
		}
		__antithesis_instrumentation__.Notify(180717)
		knownHostsData = stdout.Bytes()
		return nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(180722)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180723)
	}
	__antithesis_instrumentation__.Notify(180676)

	if err := c.Parallel(l, "distributing known_hosts", len(c.Nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180724)
		sess, err := c.newSession(c.Nodes[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(180727)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180728)
		}
		__antithesis_instrumentation__.Notify(180725)
		defer sess.Close()

		sess.SetStdin(bytes.NewReader(knownHostsData))
		const cmd = `
known_hosts_data="$(cat)"
set -e
tmp="$(tempfile -p 'roachprod' -m 0644 )"
on_exit() {
    rm -f "${tmp}"
}
trap on_exit EXIT
echo "${known_hosts_data}" > "${tmp}"
cat "${tmp}" >> ~/.ssh/known_hosts
# If our bootstrapping user is not the shared user install all of the
# relevant ssh files from the bootstrapping user into the shared user's
# .ssh directory.
if [[ "$(whoami)" != "` + config.SharedUser + `" ]]; then
    # Ensure that the shared user has a .ssh directory
    sudo -u ` + config.SharedUser +
			` bash -c "mkdir -p ~` + config.SharedUser + `/.ssh"
    # This somewhat absurd incantation ensures that we properly shell quote
    # filenames so that they both aren't expanded and work even if the filenames
    # include spaces.
    sudo find ~/.ssh -type f -execdir bash -c 'install \
        --owner ` + config.SharedUser + ` \
        --group ` + config.SharedUser + ` \
        --mode $(stat -c "%a" '"'"'{}'"'"') \
        '"'"'{}'"'"' ~` + config.SharedUser + `/.ssh' \;
fi
`
		if out, err := sess.CombinedOutput(ctx, cmd); err != nil {
			__antithesis_instrumentation__.Notify(180729)
			return nil, errors.Wrapf(err, "%s: output:\n%s", cmd, out)
		} else {
			__antithesis_instrumentation__.Notify(180730)
		}
		__antithesis_instrumentation__.Notify(180726)
		return nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(180731)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180732)
	}
	__antithesis_instrumentation__.Notify(180677)

	if len(c.AuthorizedKeys) > 0 {
		__antithesis_instrumentation__.Notify(180733)

		if err := c.Parallel(l, "adding additional authorized keys", len(c.Nodes), 0, func(i int) ([]byte, error) {
			__antithesis_instrumentation__.Notify(180734)
			sess, err := c.newSession(c.Nodes[i])
			if err != nil {
				__antithesis_instrumentation__.Notify(180737)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(180738)
			}
			__antithesis_instrumentation__.Notify(180735)
			defer sess.Close()

			sess.SetStdin(bytes.NewReader(c.AuthorizedKeys))
			const cmd = `
keys_data="$(cat)"
set -e
tmp1="$(tempfile -d ~/.ssh -p 'roachprod' )"
tmp2="$(tempfile -d ~/.ssh -p 'roachprod' )"
on_exit() {
    rm -f "${tmp1}" "${tmp2}"
}
trap on_exit EXIT
if [[ -f ~/.ssh/authorized_keys ]]; then
    cat ~/.ssh/authorized_keys > "${tmp1}"
fi
echo "${keys_data}" >> "${tmp1}"
sort -u < "${tmp1}" > "${tmp2}"
install --mode 0600 "${tmp2}" ~/.ssh/authorized_keys
if [[ "$(whoami)" != "` + config.SharedUser + `" ]]; then
    sudo install --mode 0600 \
        --owner ` + config.SharedUser + `\
        --group ` + config.SharedUser + `\
        "${tmp2}" ~` + config.SharedUser + `/.ssh/authorized_keys
fi
`
			if out, err := sess.CombinedOutput(ctx, cmd); err != nil {
				__antithesis_instrumentation__.Notify(180739)
				return nil, errors.Wrapf(err, "~ %s\n%s", cmd, out)
			} else {
				__antithesis_instrumentation__.Notify(180740)
			}
			__antithesis_instrumentation__.Notify(180736)
			return nil, nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(180741)
			return err
		} else {
			__antithesis_instrumentation__.Notify(180742)
		}
	} else {
		__antithesis_instrumentation__.Notify(180743)
	}
	__antithesis_instrumentation__.Notify(180678)

	return nil
}

func (c *SyncedCluster) DistributeCerts(ctx context.Context, l *logger.Logger) error {
	__antithesis_instrumentation__.Notify(180744)
	dir := ""
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(180753)
		dir = c.localVMDir(1)
	} else {
		__antithesis_instrumentation__.Notify(180754)
	}
	__antithesis_instrumentation__.Notify(180745)

	var existsErr error
	display := fmt.Sprintf("%s: checking certs", c.Name)
	if err := c.Parallel(l, display, 1, 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180755)
		sess, err := c.newSession(1)
		if err != nil {
			__antithesis_instrumentation__.Notify(180757)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180758)
		}
		__antithesis_instrumentation__.Notify(180756)
		defer sess.Close()
		_, existsErr = sess.CombinedOutput(ctx, `test -e `+filepath.Join(dir, `certs.tar`))
		return nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(180759)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180760)
	}
	__antithesis_instrumentation__.Notify(180746)
	if existsErr == nil {
		__antithesis_instrumentation__.Notify(180761)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(180762)
	}
	__antithesis_instrumentation__.Notify(180747)

	var msg string
	display = fmt.Sprintf("%s: initializing certs", c.Name)
	nodes := allNodes(len(c.VMs))
	var ips []string
	if !c.IsLocal() {
		__antithesis_instrumentation__.Notify(180763)
		ips = make([]string, len(nodes))
		if err := c.Parallel(l, "", len(nodes), 0, func(i int) ([]byte, error) {
			__antithesis_instrumentation__.Notify(180764)
			var err error
			ips[i], err = c.GetInternalIP(ctx, nodes[i])
			return nil, errors.Wrapf(err, "IPs")
		}); err != nil {
			__antithesis_instrumentation__.Notify(180765)
			return err
		} else {
			__antithesis_instrumentation__.Notify(180766)
		}
	} else {
		__antithesis_instrumentation__.Notify(180767)
	}
	__antithesis_instrumentation__.Notify(180748)

	if err := c.Parallel(l, display, 1, 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180768)
		sess, err := c.newSession(1)
		if err != nil {
			__antithesis_instrumentation__.Notify(180773)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180774)
		}
		__antithesis_instrumentation__.Notify(180769)
		defer sess.Close()

		var nodeNames []string
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(180775)

			nodeNames = append(nodeNames, "$(hostname)", c.VMs[0].PublicIP)
		} else {
			__antithesis_instrumentation__.Notify(180776)

			nodeNames = append(nodeNames, ips...)
			for i := range c.VMs {
				__antithesis_instrumentation__.Notify(180778)
				nodeNames = append(nodeNames, c.VMs[i].PublicIP)
			}
			__antithesis_instrumentation__.Notify(180777)
			for i := range c.VMs {
				__antithesis_instrumentation__.Notify(180779)
				nodeNames = append(nodeNames, fmt.Sprintf("%s-%04d", c.Name, i+1))

				if c.VMs[i].Provider == aws.ProviderName {
					__antithesis_instrumentation__.Notify(180780)
					nodeNames = append(nodeNames, "ip-"+strings.ReplaceAll(ips[i], ".", "-"))
				} else {
					__antithesis_instrumentation__.Notify(180781)
				}
			}
		}
		__antithesis_instrumentation__.Notify(180770)

		var cmd string
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(180782)
			cmd = fmt.Sprintf(`cd %s ; `, c.localVMDir(1))
		} else {
			__antithesis_instrumentation__.Notify(180783)
		}
		__antithesis_instrumentation__.Notify(180771)
		cmd += fmt.Sprintf(`
rm -fr certs
mkdir -p certs
%[1]s cert create-ca --certs-dir=certs --ca-key=certs/ca.key
%[1]s cert create-client root --certs-dir=certs --ca-key=certs/ca.key
%[1]s cert create-client testuser --certs-dir=certs --ca-key=certs/ca.key
%[1]s cert create-node localhost %[2]s --certs-dir=certs --ca-key=certs/ca.key
tar cvf certs.tar certs
`, cockroachNodeBinary(c, 1), strings.Join(nodeNames, " "))
		if out, err := sess.CombinedOutput(ctx, cmd); err != nil {
			__antithesis_instrumentation__.Notify(180784)
			msg = fmt.Sprintf("%s: %v", out, err)
		} else {
			__antithesis_instrumentation__.Notify(180785)
		}
		__antithesis_instrumentation__.Notify(180772)
		return nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(180786)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180787)
	}
	__antithesis_instrumentation__.Notify(180749)

	if msg != "" {
		__antithesis_instrumentation__.Notify(180788)
		fmt.Fprintln(os.Stderr, msg)
		exit.WithCode(exit.UnspecifiedError())
	} else {
		__antithesis_instrumentation__.Notify(180789)
	}
	__antithesis_instrumentation__.Notify(180750)

	var tmpfileName string
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(180790)
		tmpfileName = os.ExpandEnv(filepath.Join(dir, "certs.tar"))
	} else {
		__antithesis_instrumentation__.Notify(180791)

		tmpfile, err := ioutil.TempFile("", "certs")
		if err != nil {
			__antithesis_instrumentation__.Notify(180795)
			fmt.Fprintln(os.Stderr, err)
			exit.WithCode(exit.UnspecifiedError())
		} else {
			__antithesis_instrumentation__.Notify(180796)
		}
		__antithesis_instrumentation__.Notify(180792)
		_ = tmpfile.Close()
		defer func() {
			__antithesis_instrumentation__.Notify(180797)
			_ = os.Remove(tmpfile.Name())
		}()
		__antithesis_instrumentation__.Notify(180793)

		if err := func() error {
			__antithesis_instrumentation__.Notify(180798)
			return c.scp(fmt.Sprintf("%s@%s:certs.tar", c.user(1), c.Host(1)), tmpfile.Name())
		}(); err != nil {
			__antithesis_instrumentation__.Notify(180799)
			fmt.Fprintln(os.Stderr, err)
			exit.WithCode(exit.UnspecifiedError())
		} else {
			__antithesis_instrumentation__.Notify(180800)
		}
		__antithesis_instrumentation__.Notify(180794)

		tmpfileName = tmpfile.Name()
	}
	__antithesis_instrumentation__.Notify(180751)

	certsTar, err := ioutil.ReadFile(tmpfileName)
	if err != nil {
		__antithesis_instrumentation__.Notify(180801)
		fmt.Fprintln(os.Stderr, err)
		exit.WithCode(exit.UnspecifiedError())
	} else {
		__antithesis_instrumentation__.Notify(180802)
	}
	__antithesis_instrumentation__.Notify(180752)

	display = c.Name + ": distributing certs"
	nodes = nodes[1:]
	return c.Parallel(l, display, len(nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(180803)
		sess, err := c.newSession(nodes[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(180807)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(180808)
		}
		__antithesis_instrumentation__.Notify(180804)
		defer sess.Close()

		sess.SetStdin(bytes.NewReader(certsTar))
		var cmd string
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(180809)
			cmd = fmt.Sprintf(`cd %s ; `, c.localVMDir(nodes[i]))
		} else {
			__antithesis_instrumentation__.Notify(180810)
		}
		__antithesis_instrumentation__.Notify(180805)
		cmd += `tar xf -`
		if out, err := sess.CombinedOutput(ctx, cmd); err != nil {
			__antithesis_instrumentation__.Notify(180811)
			return nil, errors.Wrapf(err, "~ %s\n%s", cmd, out)
		} else {
			__antithesis_instrumentation__.Notify(180812)
		}
		__antithesis_instrumentation__.Notify(180806)
		return nil, nil
	})
}

const progressDone = "=======================================>"
const progressTodo = "----------------------------------------"

func formatProgress(p float64) string {
	__antithesis_instrumentation__.Notify(180813)
	i := int(math.Ceil(float64(len(progressDone)) * (1 - p)))
	if i > len(progressDone) {
		__antithesis_instrumentation__.Notify(180816)
		i = len(progressDone)
	} else {
		__antithesis_instrumentation__.Notify(180817)
	}
	__antithesis_instrumentation__.Notify(180814)
	if i < 0 {
		__antithesis_instrumentation__.Notify(180818)
		i = 0
	} else {
		__antithesis_instrumentation__.Notify(180819)
	}
	__antithesis_instrumentation__.Notify(180815)
	return fmt.Sprintf("[%s%s] %.0f%%", progressDone[i:], progressTodo[:i], 100*p)
}

func (c *SyncedCluster) Put(ctx context.Context, l *logger.Logger, src, dest string) error {
	__antithesis_instrumentation__.Notify(180820)

	var potentialSymlinkPath string
	var err error
	if potentialSymlinkPath, err = filepath.EvalSymlinks(src); err != nil {
		__antithesis_instrumentation__.Notify(180834)
		return err
	} else {
		__antithesis_instrumentation__.Notify(180835)
	}
	__antithesis_instrumentation__.Notify(180821)

	if potentialSymlinkPath != src {
		__antithesis_instrumentation__.Notify(180836)

		var symlinkTargetInfo fs.FileInfo
		if symlinkTargetInfo, err = os.Stat(potentialSymlinkPath); err != nil {
			__antithesis_instrumentation__.Notify(180838)
			return err
		} else {
			__antithesis_instrumentation__.Notify(180839)
		}
		__antithesis_instrumentation__.Notify(180837)
		redColor, resetColor := "\033[31m", "\033[0m"
		l.Printf(redColor + "WARNING: Source file is a symlink." + resetColor)
		l.Printf(redColor+"WARNING: Remote file will inherit the target permissions '%v'."+resetColor, symlinkTargetInfo.Mode())
	} else {
		__antithesis_instrumentation__.Notify(180840)
	}
	__antithesis_instrumentation__.Notify(180822)

	const treeDistFanout = 10

	var detail string
	if !c.IsLocal() {
		__antithesis_instrumentation__.Notify(180841)
		if c.UseTreeDist {
			__antithesis_instrumentation__.Notify(180842)
			detail = " (dist)"
		} else {
			__antithesis_instrumentation__.Notify(180843)
			detail = " (scp)"
		}
	} else {
		__antithesis_instrumentation__.Notify(180844)
	}
	__antithesis_instrumentation__.Notify(180823)
	l.Printf("%s: putting%s %s %s\n", c.Name, detail, src, dest)

	type result struct {
		index int
		err   error
	}

	results := make(chan result, len(c.Nodes))
	lines := make([]string, len(c.Nodes))
	var linesMu syncutil.Mutex
	var wg sync.WaitGroup
	wg.Add(len(c.Nodes))

	sources := make(chan int, len(c.Nodes))
	pushSource := func(i int) {
		__antithesis_instrumentation__.Notify(180845)
		select {
		case sources <- i:
			__antithesis_instrumentation__.Notify(180846)
		default:
			__antithesis_instrumentation__.Notify(180847)
		}
	}
	__antithesis_instrumentation__.Notify(180824)

	if c.UseTreeDist {
		__antithesis_instrumentation__.Notify(180848)

		pushSource(-1)
	} else {
		__antithesis_instrumentation__.Notify(180849)

		for range c.Nodes {
			__antithesis_instrumentation__.Notify(180850)
			pushSource(-1)
		}
	}
	__antithesis_instrumentation__.Notify(180825)

	mkpath := func(i int, dest string) (string, error) {
		__antithesis_instrumentation__.Notify(180851)
		if i == -1 {
			__antithesis_instrumentation__.Notify(180854)
			return src, nil
		} else {
			__antithesis_instrumentation__.Notify(180855)
		}
		__antithesis_instrumentation__.Notify(180852)

		e := expander{
			node: c.Nodes[i],
		}
		dest, err := e.expand(ctx, l, c, dest)
		if err != nil {
			__antithesis_instrumentation__.Notify(180856)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(180857)
		}
		__antithesis_instrumentation__.Notify(180853)
		return fmt.Sprintf("%s@%s:%s", c.user(c.Nodes[i]), c.Host(c.Nodes[i]), dest), nil
	}
	__antithesis_instrumentation__.Notify(180826)

	for i := range c.Nodes {
		__antithesis_instrumentation__.Notify(180858)
		go func(i int, dest string) {
			__antithesis_instrumentation__.Notify(180859)
			defer wg.Done()

			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(180863)

				e := expander{
					node: c.Nodes[i],
				}
				var err error
				dest, err = e.expand(ctx, l, c, dest)
				if err != nil {
					__antithesis_instrumentation__.Notify(180868)
					results <- result{i, err}
					return
				} else {
					__antithesis_instrumentation__.Notify(180869)
				}
				__antithesis_instrumentation__.Notify(180864)
				if _, err := os.Stat(src); err != nil {
					__antithesis_instrumentation__.Notify(180870)
					results <- result{i, err}
					return
				} else {
					__antithesis_instrumentation__.Notify(180871)
				}
				__antithesis_instrumentation__.Notify(180865)
				from, err := filepath.Abs(src)
				if err != nil {
					__antithesis_instrumentation__.Notify(180872)
					results <- result{i, err}
					return
				} else {
					__antithesis_instrumentation__.Notify(180873)
				}
				__antithesis_instrumentation__.Notify(180866)

				var to string
				if filepath.IsAbs(dest) {
					__antithesis_instrumentation__.Notify(180874)
					to = dest
				} else {
					__antithesis_instrumentation__.Notify(180875)
					to = filepath.Join(c.localVMDir(c.Nodes[i]), dest)
				}
				__antithesis_instrumentation__.Notify(180867)

				_ = os.Remove(to)
				results <- result{i, os.Symlink(from, to)}
				return
			} else {
				__antithesis_instrumentation__.Notify(180876)
			}
			__antithesis_instrumentation__.Notify(180860)

			srcIndex := <-sources
			from, err := mkpath(srcIndex, dest)
			if err != nil {
				__antithesis_instrumentation__.Notify(180877)
				results <- result{i, err}
				return
			} else {
				__antithesis_instrumentation__.Notify(180878)
			}
			__antithesis_instrumentation__.Notify(180861)

			to, err := mkpath(i, dest)
			if err != nil {
				__antithesis_instrumentation__.Notify(180879)
				results <- result{i, err}
				return
			} else {
				__antithesis_instrumentation__.Notify(180880)
			}
			__antithesis_instrumentation__.Notify(180862)

			err = c.scp(from, to)
			results <- result{i, err}

			if err != nil {
				__antithesis_instrumentation__.Notify(180881)

				pushSource(srcIndex)
			} else {
				__antithesis_instrumentation__.Notify(180882)

				if srcIndex != -1 {
					__antithesis_instrumentation__.Notify(180884)
					pushSource(srcIndex)
				} else {
					__antithesis_instrumentation__.Notify(180885)
				}
				__antithesis_instrumentation__.Notify(180883)

				for j := 0; j < treeDistFanout; j++ {
					__antithesis_instrumentation__.Notify(180886)
					pushSource(i)
				}
			}
		}(i, dest)
	}
	__antithesis_instrumentation__.Notify(180827)

	go func() {
		__antithesis_instrumentation__.Notify(180887)
		wg.Wait()
		close(results)
	}()
	__antithesis_instrumentation__.Notify(180828)

	var writer ui.Writer
	var ticker *time.Ticker
	if !config.Quiet {
		__antithesis_instrumentation__.Notify(180888)
		ticker = time.NewTicker(100 * time.Millisecond)
	} else {
		__antithesis_instrumentation__.Notify(180889)
		ticker = time.NewTicker(1000 * time.Millisecond)
	}
	__antithesis_instrumentation__.Notify(180829)
	defer ticker.Stop()
	var errOnce sync.Once
	var finalErr error
	setErr := func(e error) {
		__antithesis_instrumentation__.Notify(180890)
		if e != nil {
			__antithesis_instrumentation__.Notify(180891)
			errOnce.Do(func() {
				__antithesis_instrumentation__.Notify(180892)
				finalErr = e
			})
		} else {
			__antithesis_instrumentation__.Notify(180893)
		}
	}
	__antithesis_instrumentation__.Notify(180830)

	var spinner = []string{"|", "/", "-", "\\"}
	spinnerIdx := 0

	for done := false; !done; {
		__antithesis_instrumentation__.Notify(180894)
		select {
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(180896)
			if config.Quiet && func() bool {
				__antithesis_instrumentation__.Notify(180898)
				return l.File == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(180899)
				fmt.Printf(".")
			} else {
				__antithesis_instrumentation__.Notify(180900)
			}
		case r, ok := <-results:
			__antithesis_instrumentation__.Notify(180897)
			done = !ok
			if ok {
				__antithesis_instrumentation__.Notify(180901)
				linesMu.Lock()
				if r.err != nil {
					__antithesis_instrumentation__.Notify(180903)
					setErr(r.err)
					lines[r.index] = r.err.Error()
				} else {
					__antithesis_instrumentation__.Notify(180904)
					lines[r.index] = "done"
				}
				__antithesis_instrumentation__.Notify(180902)
				linesMu.Unlock()
			} else {
				__antithesis_instrumentation__.Notify(180905)
			}
		}
		__antithesis_instrumentation__.Notify(180895)
		if !config.Quiet {
			__antithesis_instrumentation__.Notify(180906)
			linesMu.Lock()
			for i := range lines {
				__antithesis_instrumentation__.Notify(180908)
				fmt.Fprintf(&writer, "  %2d: ", c.Nodes[i])
				if lines[i] != "" {
					__antithesis_instrumentation__.Notify(180910)
					fmt.Fprintf(&writer, "%s", lines[i])
				} else {
					__antithesis_instrumentation__.Notify(180911)
					fmt.Fprintf(&writer, "%s", spinner[spinnerIdx%len(spinner)])
				}
				__antithesis_instrumentation__.Notify(180909)
				fmt.Fprintf(&writer, "\n")
			}
			__antithesis_instrumentation__.Notify(180907)
			linesMu.Unlock()
			_ = writer.Flush(l.Stdout)
			spinnerIdx++
		} else {
			__antithesis_instrumentation__.Notify(180912)
		}
	}
	__antithesis_instrumentation__.Notify(180831)

	if config.Quiet && func() bool {
		__antithesis_instrumentation__.Notify(180913)
		return l.File != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(180914)
		l.Printf("\n")
		linesMu.Lock()
		for i := range lines {
			__antithesis_instrumentation__.Notify(180916)
			l.Printf("  %2d: %s", c.Nodes[i], lines[i])
		}
		__antithesis_instrumentation__.Notify(180915)
		linesMu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(180917)
	}
	__antithesis_instrumentation__.Notify(180832)

	if finalErr != nil {
		__antithesis_instrumentation__.Notify(180918)
		return errors.Wrapf(finalErr, "put %s failed", src)
	} else {
		__antithesis_instrumentation__.Notify(180919)
	}
	__antithesis_instrumentation__.Notify(180833)
	return nil
}

func (c *SyncedCluster) Logs(
	src, dest, user, filter, programFilter string,
	interval time.Duration,
	from, to time.Time,
	out io.Writer,
) error {
	__antithesis_instrumentation__.Notify(180920)
	rsyncNodeLogs := func(ctx context.Context, node Node) error {
		__antithesis_instrumentation__.Notify(180929)
		base := fmt.Sprintf("%d.logs", node)
		local := filepath.Join(dest, base) + "/"
		sshUser := c.user(node)
		rsyncArgs := []string{"-az", "--size-only"}
		var remote string
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(180932)

			localHome := filepath.Dir(c.LogDir(node))
			remote = filepath.Join(localHome, src) + "/"
		} else {
			__antithesis_instrumentation__.Notify(180933)
			logDir := src
			if !filepath.IsAbs(logDir) && func() bool {
				__antithesis_instrumentation__.Notify(180935)
				return user != "" == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(180936)
				return user != sshUser == true
			}() == true {
				__antithesis_instrumentation__.Notify(180937)
				logDir = "~" + user + "/" + logDir
			} else {
				__antithesis_instrumentation__.Notify(180938)
			}
			__antithesis_instrumentation__.Notify(180934)
			remote = fmt.Sprintf("%s@%s:%s/", c.user(node), c.Host(node), logDir)

			rsyncArgs = append(rsyncArgs, "--rsh", "ssh "+
				"-o StrictHostKeyChecking=no "+
				"-o ControlMaster=auto "+
				"-o ControlPath=~/.ssh/%r@%h:%p "+
				"-o UserKnownHostsFile=/dev/null "+
				"-o ControlPersist=2m "+
				strings.Join(sshAuthArgs(), " "))

			if user != "" && func() bool {
				__antithesis_instrumentation__.Notify(180939)
				return user != sshUser == true
			}() == true {
				__antithesis_instrumentation__.Notify(180940)
				rsyncArgs = append(rsyncArgs, "--rsync-path",
					fmt.Sprintf("sudo -u %s rsync", user))
			} else {
				__antithesis_instrumentation__.Notify(180941)
			}
		}
		__antithesis_instrumentation__.Notify(180930)
		rsyncArgs = append(rsyncArgs, remote, local)
		cmd := exec.CommandContext(ctx, "rsync", rsyncArgs...)
		var stderrBuf bytes.Buffer
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderrBuf
		if err := cmd.Run(); err != nil {
			__antithesis_instrumentation__.Notify(180942)
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(180944)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(180945)
			}
			__antithesis_instrumentation__.Notify(180943)
			return errors.Wrapf(err, "failed to rsync from %v to %v:\n%s\n",
				src, dest, stderrBuf.String())
		} else {
			__antithesis_instrumentation__.Notify(180946)
		}
		__antithesis_instrumentation__.Notify(180931)
		return nil
	}
	__antithesis_instrumentation__.Notify(180921)
	rsyncLogs := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(180947)
		g, gctx := errgroup.WithContext(ctx)
		for i := range c.Nodes {
			__antithesis_instrumentation__.Notify(180949)
			node := c.Nodes[i]
			g.Go(func() error {
				__antithesis_instrumentation__.Notify(180950)
				return rsyncNodeLogs(gctx, node)
			})
		}
		__antithesis_instrumentation__.Notify(180948)
		return g.Wait()
	}
	__antithesis_instrumentation__.Notify(180922)
	mergeLogs := func(ctx context.Context, prev, t time.Time) error {
		__antithesis_instrumentation__.Notify(180951)
		cmd := exec.CommandContext(ctx, "cockroach", "debug", "merge-logs",
			dest+"/*/*",
			"--from", prev.Format(time.RFC3339),
			"--to", t.Format(time.RFC3339))
		if filter != "" {
			__antithesis_instrumentation__.Notify(180956)
			cmd.Args = append(cmd.Args, "--filter", filter)
		} else {
			__antithesis_instrumentation__.Notify(180957)
		}
		__antithesis_instrumentation__.Notify(180952)
		if programFilter != "" {
			__antithesis_instrumentation__.Notify(180958)
			cmd.Args = append(cmd.Args, "--program-filter", programFilter)
		} else {
			__antithesis_instrumentation__.Notify(180959)
		}
		__antithesis_instrumentation__.Notify(180953)

		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(180960)
			cmd.Args = append(cmd.Args,
				"--file-pattern", "^(?:.*/)?(?P<id>[0-9]+).*/"+log.FileNamePattern+"$",
				"--prefix", "${id}> ")
		} else {
			__antithesis_instrumentation__.Notify(180961)
		}
		__antithesis_instrumentation__.Notify(180954)
		cmd.Stdout = out
		var errBuf bytes.Buffer
		cmd.Stderr = &errBuf
		if err := cmd.Run(); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(180962)
			return ctx.Err() == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(180963)
			return errors.Wrapf(err, "failed to run cockroach debug merge-logs:\n%v", errBuf.String())
		} else {
			__antithesis_instrumentation__.Notify(180964)
		}
		__antithesis_instrumentation__.Notify(180955)
		return nil
	}
	__antithesis_instrumentation__.Notify(180923)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := os.MkdirAll(dest, 0755); err != nil {
		__antithesis_instrumentation__.Notify(180965)
		return errors.Wrapf(err, "failed to create destination directory")
	} else {
		__antithesis_instrumentation__.Notify(180966)
	}
	__antithesis_instrumentation__.Notify(180924)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer func() { __antithesis_instrumentation__.Notify(180967); signal.Stop(ch); close(ch) }()
	__antithesis_instrumentation__.Notify(180925)
	go func() { __antithesis_instrumentation__.Notify(180968); <-ch; cancel() }()
	__antithesis_instrumentation__.Notify(180926)

	prev := from
	if prev.IsZero() {
		__antithesis_instrumentation__.Notify(180969)
		prev = timeutil.Now().Add(-2 * time.Second).Truncate(time.Microsecond)
	} else {
		__antithesis_instrumentation__.Notify(180970)
	}
	__antithesis_instrumentation__.Notify(180927)
	for to.IsZero() || func() bool {
		__antithesis_instrumentation__.Notify(180971)
		return prev.Before(to) == true
	}() == true {
		__antithesis_instrumentation__.Notify(180972)

		t := timeutil.Now().Add(-1100 * time.Millisecond).Truncate(time.Microsecond)
		if err := rsyncLogs(ctx); err != nil {
			__antithesis_instrumentation__.Notify(180977)
			return errors.Wrapf(err, "failed to sync logs")
		} else {
			__antithesis_instrumentation__.Notify(180978)
		}
		__antithesis_instrumentation__.Notify(180973)
		if !to.IsZero() && func() bool {
			__antithesis_instrumentation__.Notify(180979)
			return t.After(to) == true
		}() == true {
			__antithesis_instrumentation__.Notify(180980)
			t = to
		} else {
			__antithesis_instrumentation__.Notify(180981)
		}
		__antithesis_instrumentation__.Notify(180974)
		if err := mergeLogs(ctx, prev, t); err != nil {
			__antithesis_instrumentation__.Notify(180982)
			return err
		} else {
			__antithesis_instrumentation__.Notify(180983)
		}
		__antithesis_instrumentation__.Notify(180975)
		prev = t
		if !to.IsZero() && func() bool {
			__antithesis_instrumentation__.Notify(180984)
			return !prev.Before(to) == true
		}() == true {
			__antithesis_instrumentation__.Notify(180985)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(180986)
		}
		__antithesis_instrumentation__.Notify(180976)
		select {
		case <-time.After(interval):
			__antithesis_instrumentation__.Notify(180987)
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(180988)
			return nil
		}
	}
	__antithesis_instrumentation__.Notify(180928)
	return nil
}

func (c *SyncedCluster) Get(l *logger.Logger, src, dest string) error {
	__antithesis_instrumentation__.Notify(180989)

	var detail string
	if !c.IsLocal() {
		__antithesis_instrumentation__.Notify(180997)
		detail = " (scp)"
	} else {
		__antithesis_instrumentation__.Notify(180998)
	}
	__antithesis_instrumentation__.Notify(180990)
	l.Printf("%s: getting%s %s %s\n", c.Name, detail, src, dest)

	type result struct {
		index int
		err   error
	}

	var writer ui.Writer
	results := make(chan result, len(c.Nodes))
	lines := make([]string, len(c.Nodes))
	var linesMu syncutil.Mutex

	var wg sync.WaitGroup
	for i := range c.Nodes {
		__antithesis_instrumentation__.Notify(180999)
		wg.Add(1)
		go func(i int) {
			__antithesis_instrumentation__.Notify(181000)
			defer wg.Done()

			src := src
			dest := dest
			if len(c.Nodes) > 1 {
				__antithesis_instrumentation__.Notify(181005)
				base := fmt.Sprintf("%d.%s", c.Nodes[i], filepath.Base(dest))
				dest = filepath.Join(filepath.Dir(dest), base)
			} else {
				__antithesis_instrumentation__.Notify(181006)
			}
			__antithesis_instrumentation__.Notify(181001)

			progress := func(p float64) {
				__antithesis_instrumentation__.Notify(181007)
				linesMu.Lock()
				defer linesMu.Unlock()
				lines[i] = formatProgress(p)
			}
			__antithesis_instrumentation__.Notify(181002)

			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(181008)
				if !filepath.IsAbs(src) {
					__antithesis_instrumentation__.Notify(181012)
					src = filepath.Join(c.localVMDir(c.Nodes[i]), src)
				} else {
					__antithesis_instrumentation__.Notify(181013)
				}
				__antithesis_instrumentation__.Notify(181009)

				var copy func(src, dest string, info os.FileInfo) error
				copy = func(src, dest string, info os.FileInfo) error {
					__antithesis_instrumentation__.Notify(181014)

					mode := info.Mode() | 0444
					if info.IsDir() {
						__antithesis_instrumentation__.Notify(181020)
						if err := os.MkdirAll(dest, mode); err != nil {
							__antithesis_instrumentation__.Notify(181024)
							return err
						} else {
							__antithesis_instrumentation__.Notify(181025)
						}
						__antithesis_instrumentation__.Notify(181021)

						infos, err := ioutil.ReadDir(src)
						if err != nil {
							__antithesis_instrumentation__.Notify(181026)
							return err
						} else {
							__antithesis_instrumentation__.Notify(181027)
						}
						__antithesis_instrumentation__.Notify(181022)

						for _, info := range infos {
							__antithesis_instrumentation__.Notify(181028)
							if err := copy(
								filepath.Join(src, info.Name()),
								filepath.Join(dest, info.Name()),
								info,
							); err != nil {
								__antithesis_instrumentation__.Notify(181029)
								return err
							} else {
								__antithesis_instrumentation__.Notify(181030)
							}
						}
						__antithesis_instrumentation__.Notify(181023)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(181031)
					}
					__antithesis_instrumentation__.Notify(181015)

					if !mode.IsRegular() {
						__antithesis_instrumentation__.Notify(181032)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(181033)
					}
					__antithesis_instrumentation__.Notify(181016)

					out, err := os.Create(dest)
					if err != nil {
						__antithesis_instrumentation__.Notify(181034)
						return err
					} else {
						__antithesis_instrumentation__.Notify(181035)
					}
					__antithesis_instrumentation__.Notify(181017)
					defer out.Close()

					if err := os.Chmod(out.Name(), mode); err != nil {
						__antithesis_instrumentation__.Notify(181036)
						return err
					} else {
						__antithesis_instrumentation__.Notify(181037)
					}
					__antithesis_instrumentation__.Notify(181018)

					in, err := os.Open(src)
					if err != nil {
						__antithesis_instrumentation__.Notify(181038)
						return err
					} else {
						__antithesis_instrumentation__.Notify(181039)
					}
					__antithesis_instrumentation__.Notify(181019)
					defer in.Close()

					p := &ssh.ProgressWriter{
						Writer:   out,
						Done:     0,
						Total:    info.Size(),
						Progress: progress,
					}
					_, err = io.Copy(p, in)
					return err
				}
				__antithesis_instrumentation__.Notify(181010)

				info, err := os.Stat(src)
				if err != nil {
					__antithesis_instrumentation__.Notify(181040)
					results <- result{i, err}
					return
				} else {
					__antithesis_instrumentation__.Notify(181041)
				}
				__antithesis_instrumentation__.Notify(181011)
				err = copy(src, dest, info)
				results <- result{i, err}
				return
			} else {
				__antithesis_instrumentation__.Notify(181042)
			}
			__antithesis_instrumentation__.Notify(181003)

			err := c.scp(fmt.Sprintf("%s@%s:%s", c.user(c.Nodes[0]), c.Host(c.Nodes[i]), src), dest)
			if err == nil {
				__antithesis_instrumentation__.Notify(181043)

				chmod := func(path string, info os.FileInfo, err error) error {
					__antithesis_instrumentation__.Notify(181045)
					if err != nil {
						__antithesis_instrumentation__.Notify(181048)
						return err
					} else {
						__antithesis_instrumentation__.Notify(181049)
					}
					__antithesis_instrumentation__.Notify(181046)
					const oRead = 0004
					if mode := info.Mode(); mode&oRead == 0 {
						__antithesis_instrumentation__.Notify(181050)
						if err := os.Chmod(path, mode|oRead); err != nil {
							__antithesis_instrumentation__.Notify(181051)
							return err
						} else {
							__antithesis_instrumentation__.Notify(181052)
						}
					} else {
						__antithesis_instrumentation__.Notify(181053)
					}
					__antithesis_instrumentation__.Notify(181047)
					return nil
				}
				__antithesis_instrumentation__.Notify(181044)
				err = filepath.Walk(dest, chmod)
			} else {
				__antithesis_instrumentation__.Notify(181054)
			}
			__antithesis_instrumentation__.Notify(181004)

			results <- result{i, err}
		}(i)
	}
	__antithesis_instrumentation__.Notify(180991)

	go func() {
		__antithesis_instrumentation__.Notify(181055)
		wg.Wait()
		close(results)
	}()
	__antithesis_instrumentation__.Notify(180992)

	var ticker *time.Ticker
	if config.Quiet {
		__antithesis_instrumentation__.Notify(181056)
		ticker = time.NewTicker(100 * time.Millisecond)
	} else {
		__antithesis_instrumentation__.Notify(181057)
		ticker = time.NewTicker(1000 * time.Millisecond)
	}
	__antithesis_instrumentation__.Notify(180993)
	defer ticker.Stop()
	haveErr := false

	var spinner = []string{"|", "/", "-", "\\"}
	spinnerIdx := 0

	for done := false; !done; {
		__antithesis_instrumentation__.Notify(181058)
		select {
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(181060)
			if config.Quiet && func() bool {
				__antithesis_instrumentation__.Notify(181062)
				return l.File == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(181063)
				fmt.Printf(".")
			} else {
				__antithesis_instrumentation__.Notify(181064)
			}
		case r, ok := <-results:
			__antithesis_instrumentation__.Notify(181061)
			done = !ok
			if ok {
				__antithesis_instrumentation__.Notify(181065)
				linesMu.Lock()
				if r.err != nil {
					__antithesis_instrumentation__.Notify(181067)
					haveErr = true
					lines[r.index] = r.err.Error()
				} else {
					__antithesis_instrumentation__.Notify(181068)
					lines[r.index] = "done"
				}
				__antithesis_instrumentation__.Notify(181066)
				linesMu.Unlock()
			} else {
				__antithesis_instrumentation__.Notify(181069)
			}
		}
		__antithesis_instrumentation__.Notify(181059)
		if !config.Quiet && func() bool {
			__antithesis_instrumentation__.Notify(181070)
			return l.File == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(181071)
			linesMu.Lock()
			for i := range lines {
				__antithesis_instrumentation__.Notify(181073)
				fmt.Fprintf(&writer, "  %2d: ", c.Nodes[i])
				if lines[i] != "" {
					__antithesis_instrumentation__.Notify(181075)
					fmt.Fprintf(&writer, "%s", lines[i])
				} else {
					__antithesis_instrumentation__.Notify(181076)
					fmt.Fprintf(&writer, "%s", spinner[spinnerIdx%len(spinner)])
				}
				__antithesis_instrumentation__.Notify(181074)
				fmt.Fprintf(&writer, "\n")
			}
			__antithesis_instrumentation__.Notify(181072)
			linesMu.Unlock()
			_ = writer.Flush(l.Stdout)
			spinnerIdx++
		} else {
			__antithesis_instrumentation__.Notify(181077)
		}
	}
	__antithesis_instrumentation__.Notify(180994)

	if config.Quiet && func() bool {
		__antithesis_instrumentation__.Notify(181078)
		return l.File == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(181079)
		l.Printf("\n")
		linesMu.Lock()
		for i := range lines {
			__antithesis_instrumentation__.Notify(181081)
			l.Printf("  %2d: %s", c.Nodes[i], lines[i])
		}
		__antithesis_instrumentation__.Notify(181080)
		linesMu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(181082)
	}
	__antithesis_instrumentation__.Notify(180995)

	if haveErr {
		__antithesis_instrumentation__.Notify(181083)
		return errors.Newf("get %s failed", src)
	} else {
		__antithesis_instrumentation__.Notify(181084)
	}
	__antithesis_instrumentation__.Notify(180996)
	return nil
}

func (c *SyncedCluster) pgurls(
	ctx context.Context, l *logger.Logger, nodes Nodes,
) (map[Node]string, error) {
	__antithesis_instrumentation__.Notify(181085)
	hosts, err := c.pghosts(ctx, l, nodes)
	if err != nil {
		__antithesis_instrumentation__.Notify(181088)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181089)
	}
	__antithesis_instrumentation__.Notify(181086)
	m := make(map[Node]string, len(hosts))
	for node, host := range hosts {
		__antithesis_instrumentation__.Notify(181090)
		m[node] = c.NodeURL(host, c.NodePort(node))
	}
	__antithesis_instrumentation__.Notify(181087)
	return m, nil
}

func (c *SyncedCluster) pghosts(
	ctx context.Context, l *logger.Logger, nodes Nodes,
) (map[Node]string, error) {
	__antithesis_instrumentation__.Notify(181091)
	ips := make([]string, len(nodes))
	if err := c.Parallel(l, "", len(nodes), 0, func(i int) ([]byte, error) {
		__antithesis_instrumentation__.Notify(181094)
		var err error
		ips[i], err = c.GetInternalIP(ctx, nodes[i])
		return nil, errors.Wrapf(err, "pghosts")
	}); err != nil {
		__antithesis_instrumentation__.Notify(181095)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181096)
	}
	__antithesis_instrumentation__.Notify(181092)

	m := make(map[Node]string, len(ips))
	for i, ip := range ips {
		__antithesis_instrumentation__.Notify(181097)
		m[nodes[i]] = ip
	}
	__antithesis_instrumentation__.Notify(181093)
	return m, nil
}

func (c *SyncedCluster) SSH(ctx context.Context, l *logger.Logger, sshArgs, args []string) error {
	__antithesis_instrumentation__.Notify(181098)
	if len(c.Nodes) != 1 && func() bool {
		__antithesis_instrumentation__.Notify(181103)
		return len(args) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(181104)

		sshed, err := maybeSplitScreenSSHITerm2(c)
		if sshed {
			__antithesis_instrumentation__.Notify(181105)
			return err
		} else {
			__antithesis_instrumentation__.Notify(181106)
		}
	} else {
		__antithesis_instrumentation__.Notify(181107)
	}
	__antithesis_instrumentation__.Notify(181099)

	e := expander{
		node: c.Nodes[0],
	}
	var expandedArgs []string
	for _, arg := range args {
		__antithesis_instrumentation__.Notify(181108)
		expandedArg, err := e.expand(ctx, l, c, arg)
		if err != nil {
			__antithesis_instrumentation__.Notify(181110)
			return err
		} else {
			__antithesis_instrumentation__.Notify(181111)
		}
		__antithesis_instrumentation__.Notify(181109)
		expandedArgs = append(expandedArgs, strings.Split(expandedArg, " ")...)
	}
	__antithesis_instrumentation__.Notify(181100)

	var allArgs []string
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(181112)
		allArgs = []string{
			"/bin/bash", "-c",
		}
		cmd := fmt.Sprintf("cd %s ; ", c.localVMDir(c.Nodes[0]))
		if len(args) == 0 {
			__antithesis_instrumentation__.Notify(181115)
			cmd += "/bin/bash "
		} else {
			__antithesis_instrumentation__.Notify(181116)
		}
		__antithesis_instrumentation__.Notify(181113)
		if len(args) > 0 {
			__antithesis_instrumentation__.Notify(181117)
			cmd += fmt.Sprintf("export ROACHPROD=%s ; ", c.roachprodEnvValue(c.Nodes[0]))
			cmd += strings.Join(expandedArgs, " ")
		} else {
			__antithesis_instrumentation__.Notify(181118)
		}
		__antithesis_instrumentation__.Notify(181114)
		allArgs = append(allArgs, cmd)
	} else {
		__antithesis_instrumentation__.Notify(181119)
		allArgs = []string{
			"ssh",
			fmt.Sprintf("%s@%s", c.user(c.Nodes[0]), c.Host(c.Nodes[0])),
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "StrictHostKeyChecking=no",
		}
		allArgs = append(allArgs, sshAuthArgs()...)
		allArgs = append(allArgs, sshArgs...)
		if len(args) > 0 {
			__antithesis_instrumentation__.Notify(181121)
			allArgs = append(allArgs, fmt.Sprintf(
				"export ROACHPROD=%s ;", c.roachprodEnvValue(c.Nodes[0]),
			))
		} else {
			__antithesis_instrumentation__.Notify(181122)
		}
		__antithesis_instrumentation__.Notify(181120)
		allArgs = append(allArgs, expandedArgs...)
	}
	__antithesis_instrumentation__.Notify(181101)

	sshPath, err := exec.LookPath(allArgs[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(181123)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181124)
	}
	__antithesis_instrumentation__.Notify(181102)
	return syscall.Exec(sshPath, allArgs, os.Environ())
}

func (c *SyncedCluster) scp(src, dest string) error {
	__antithesis_instrumentation__.Notify(181125)
	args := []string{
		"scp", "-r", "-C",
		"-o", "StrictHostKeyChecking=no",
	}
	args = append(args, sshAuthArgs()...)
	args = append(args, src, dest)
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		__antithesis_instrumentation__.Notify(181127)
		return errors.Wrapf(err, "~ %s\n%s", strings.Join(args, " "), out)
	} else {
		__antithesis_instrumentation__.Notify(181128)
	}
	__antithesis_instrumentation__.Notify(181126)
	return nil
}

type ParallelResult struct {
	Index int
	Out   []byte
	Err   error
}

func (c *SyncedCluster) Parallel(
	l *logger.Logger, display string, count, concurrency int, fn func(i int) ([]byte, error),
) error {
	__antithesis_instrumentation__.Notify(181129)
	failed, err := c.ParallelE(l, display, count, concurrency, fn)
	if err != nil {
		__antithesis_instrumentation__.Notify(181131)
		sort.Slice(failed, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(181134)
			return failed[i].Index < failed[j].Index
		})
		__antithesis_instrumentation__.Notify(181132)
		for _, f := range failed {
			__antithesis_instrumentation__.Notify(181135)
			fmt.Fprintf(l.Stderr, "%d: %+v: %s\n", f.Index, f.Err, f.Out)
		}
		__antithesis_instrumentation__.Notify(181133)
		return err
	} else {
		__antithesis_instrumentation__.Notify(181136)
	}
	__antithesis_instrumentation__.Notify(181130)
	return nil
}

func (c *SyncedCluster) ParallelE(
	l *logger.Logger, display string, count, concurrency int, fn func(i int) ([]byte, error),
) ([]ParallelResult, error) {
	__antithesis_instrumentation__.Notify(181137)
	if concurrency == 0 || func() bool {
		__antithesis_instrumentation__.Notify(181148)
		return concurrency > count == true
	}() == true {
		__antithesis_instrumentation__.Notify(181149)
		concurrency = count
	} else {
		__antithesis_instrumentation__.Notify(181150)
	}
	__antithesis_instrumentation__.Notify(181138)
	if config.MaxConcurrency > 0 && func() bool {
		__antithesis_instrumentation__.Notify(181151)
		return concurrency > config.MaxConcurrency == true
	}() == true {
		__antithesis_instrumentation__.Notify(181152)
		concurrency = config.MaxConcurrency
	} else {
		__antithesis_instrumentation__.Notify(181153)
	}
	__antithesis_instrumentation__.Notify(181139)

	results := make(chan ParallelResult, count)
	var wg sync.WaitGroup
	wg.Add(count)

	var index int
	startNext := func() {
		__antithesis_instrumentation__.Notify(181154)
		go func(i int) {
			__antithesis_instrumentation__.Notify(181156)
			defer wg.Done()
			out, err := fn(i)
			results <- ParallelResult{i, out, err}
		}(index)
		__antithesis_instrumentation__.Notify(181155)
		index++
	}
	__antithesis_instrumentation__.Notify(181140)

	for index < concurrency {
		__antithesis_instrumentation__.Notify(181157)
		startNext()
	}
	__antithesis_instrumentation__.Notify(181141)

	go func() {
		__antithesis_instrumentation__.Notify(181158)
		wg.Wait()
		close(results)
	}()
	__antithesis_instrumentation__.Notify(181142)

	var writer ui.Writer
	out := l.Stdout
	if display == "" {
		__antithesis_instrumentation__.Notify(181159)
		out = ioutil.Discard
	} else {
		__antithesis_instrumentation__.Notify(181160)
	}
	__antithesis_instrumentation__.Notify(181143)

	var ticker *time.Ticker
	if !config.Quiet {
		__antithesis_instrumentation__.Notify(181161)
		ticker = time.NewTicker(100 * time.Millisecond)
	} else {
		__antithesis_instrumentation__.Notify(181162)
		ticker = time.NewTicker(1000 * time.Millisecond)
		fmt.Fprintf(out, "%s", display)
		if l.File != nil {
			__antithesis_instrumentation__.Notify(181163)
			fmt.Fprintf(out, "\n")
		} else {
			__antithesis_instrumentation__.Notify(181164)
		}
	}
	__antithesis_instrumentation__.Notify(181144)
	defer ticker.Stop()
	complete := make([]bool, count)
	var failed []ParallelResult

	var spinner = []string{"|", "/", "-", "\\"}
	spinnerIdx := 0

	for done := false; !done; {
		__antithesis_instrumentation__.Notify(181165)
		select {
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(181167)
			if config.Quiet && func() bool {
				__antithesis_instrumentation__.Notify(181171)
				return l.File == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(181172)
				fmt.Fprintf(out, ".")
			} else {
				__antithesis_instrumentation__.Notify(181173)
			}
		case r, ok := <-results:
			__antithesis_instrumentation__.Notify(181168)
			if r.Err != nil {
				__antithesis_instrumentation__.Notify(181174)
				failed = append(failed, r)
			} else {
				__antithesis_instrumentation__.Notify(181175)
			}
			__antithesis_instrumentation__.Notify(181169)
			done = !ok
			if ok {
				__antithesis_instrumentation__.Notify(181176)
				complete[r.Index] = true
			} else {
				__antithesis_instrumentation__.Notify(181177)
			}
			__antithesis_instrumentation__.Notify(181170)
			if index < count {
				__antithesis_instrumentation__.Notify(181178)
				startNext()
			} else {
				__antithesis_instrumentation__.Notify(181179)
			}
		}
		__antithesis_instrumentation__.Notify(181166)

		if !config.Quiet && func() bool {
			__antithesis_instrumentation__.Notify(181180)
			return l.File == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(181181)
			fmt.Fprint(&writer, display)
			var n int
			for i := range complete {
				__antithesis_instrumentation__.Notify(181184)
				if complete[i] {
					__antithesis_instrumentation__.Notify(181185)
					n++
				} else {
					__antithesis_instrumentation__.Notify(181186)
				}
			}
			__antithesis_instrumentation__.Notify(181182)
			fmt.Fprintf(&writer, " %d/%d", n, len(complete))
			if !done {
				__antithesis_instrumentation__.Notify(181187)
				fmt.Fprintf(&writer, " %s", spinner[spinnerIdx%len(spinner)])
			} else {
				__antithesis_instrumentation__.Notify(181188)
			}
			__antithesis_instrumentation__.Notify(181183)
			fmt.Fprintf(&writer, "\n")
			_ = writer.Flush(out)
			spinnerIdx++
		} else {
			__antithesis_instrumentation__.Notify(181189)
		}
	}
	__antithesis_instrumentation__.Notify(181145)

	if config.Quiet && func() bool {
		__antithesis_instrumentation__.Notify(181190)
		return l.File == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(181191)
		fmt.Fprintf(out, "\n")
	} else {
		__antithesis_instrumentation__.Notify(181192)
	}
	__antithesis_instrumentation__.Notify(181146)

	if len(failed) > 0 {
		__antithesis_instrumentation__.Notify(181193)
		return failed, errors.New("one or more parallel execution failure")
	} else {
		__antithesis_instrumentation__.Notify(181194)
	}
	__antithesis_instrumentation__.Notify(181147)
	return nil, nil
}

func (c *SyncedCluster) Init(ctx context.Context, l *logger.Logger) error {
	__antithesis_instrumentation__.Notify(181195)

	const firstNodeIdx = 0

	l.Printf("%s: initializing cluster\n", c.Name)
	initOut, err := c.initializeCluster(ctx, firstNodeIdx)
	if err != nil {
		__antithesis_instrumentation__.Notify(181200)
		return errors.WithDetail(err, "install.Init() failed: unable to initialize cluster.")
	} else {
		__antithesis_instrumentation__.Notify(181201)
	}
	__antithesis_instrumentation__.Notify(181196)
	if initOut != "" {
		__antithesis_instrumentation__.Notify(181202)
		l.Printf(initOut)
	} else {
		__antithesis_instrumentation__.Notify(181203)
	}
	__antithesis_instrumentation__.Notify(181197)

	l.Printf("%s: setting cluster settings", c.Name)
	clusterSettingsOut, err := c.setClusterSettings(ctx, l, firstNodeIdx)
	if err != nil {
		__antithesis_instrumentation__.Notify(181204)
		return errors.WithDetail(err, "install.Init() failed: unable to set cluster settings.")
	} else {
		__antithesis_instrumentation__.Notify(181205)
	}
	__antithesis_instrumentation__.Notify(181198)
	if clusterSettingsOut != "" {
		__antithesis_instrumentation__.Notify(181206)
		l.Printf(clusterSettingsOut)
	} else {
		__antithesis_instrumentation__.Notify(181207)
	}
	__antithesis_instrumentation__.Notify(181199)
	return nil
}
