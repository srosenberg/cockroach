package install

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/errors"
)

type session interface {
	CombinedOutput(ctx context.Context, cmd string) ([]byte, error)
	Run(ctx context.Context, cmd string) error
	SetWithExitStatus(value bool)
	SetStdin(r io.Reader)
	SetStdout(w io.Writer)
	SetStderr(w io.Writer)
	Start(cmd string) error
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)
	RequestPty() error
	Wait() error
	Close()
}

type remoteSession struct {
	*exec.Cmd
	cancel         func()
	logfile        string
	withExitStatus bool
}

func newRemoteSession(user, host string, logdir string) (*remoteSession, error) {
	__antithesis_instrumentation__.Notify(181651)

	const logfile = ""
	withExitStatus := false
	args := []string{
		user + "@" + host,

		"-q",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking=no",

		"-o", "ServerAliveInterval=60",

		"-o", "ConnectTimeout=5",
	}
	args = append(args, sshAuthArgs()...)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "ssh", args...)
	return &remoteSession{cmd, cancel, logfile, withExitStatus}, nil
}

func (s *remoteSession) errWithDebug(err error) error {
	__antithesis_instrumentation__.Notify(181652)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(181654)
		return s.logfile != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(181655)
		err = errors.Wrapf(err, "ssh verbose log retained in %s", s.logfile)
		s.logfile = ""
	} else {
		__antithesis_instrumentation__.Notify(181656)
	}
	__antithesis_instrumentation__.Notify(181653)
	return err
}

func (s *remoteSession) CombinedOutput(ctx context.Context, cmd string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(181657)
	if s.withExitStatus {
		__antithesis_instrumentation__.Notify(181661)
		cmd = strings.TrimSpace(cmd)
		if !strings.HasSuffix(cmd, ";") {
			__antithesis_instrumentation__.Notify(181663)
			cmd += ";"
		} else {
			__antithesis_instrumentation__.Notify(181664)
		}
		__antithesis_instrumentation__.Notify(181662)
		cmd += "echo -n 'LAST EXIT STATUS: '$?;"
	} else {
		__antithesis_instrumentation__.Notify(181665)
	}
	__antithesis_instrumentation__.Notify(181658)
	s.Cmd.Args = append(s.Cmd.Args, cmd)

	var b []byte
	var err error
	commandFinished := make(chan struct{})

	go func() {
		__antithesis_instrumentation__.Notify(181666)
		b, err = s.Cmd.CombinedOutput()
		err = s.errWithDebug(err)
		close(commandFinished)
	}()
	__antithesis_instrumentation__.Notify(181659)

	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(181667)
		s.Close()
		err = ctx.Err()
		break
	case <-commandFinished:
		__antithesis_instrumentation__.Notify(181668)
		break
	}
	__antithesis_instrumentation__.Notify(181660)
	return b, err
}

func (s *remoteSession) Run(ctx context.Context, cmd string) error {
	__antithesis_instrumentation__.Notify(181669)
	if s.withExitStatus {
		__antithesis_instrumentation__.Notify(181673)
		cmd = strings.TrimSpace(cmd)
		if !strings.HasSuffix(cmd, ";") {
			__antithesis_instrumentation__.Notify(181675)
			cmd += ";"
		} else {
			__antithesis_instrumentation__.Notify(181676)
		}
		__antithesis_instrumentation__.Notify(181674)
		cmd += "echo -n 'LAST EXIT STATUS: '$?;"
	} else {
		__antithesis_instrumentation__.Notify(181677)
	}
	__antithesis_instrumentation__.Notify(181670)
	s.Cmd.Args = append(s.Cmd.Args, cmd)

	var err error
	commandFinished := make(chan struct{})
	go func() {
		__antithesis_instrumentation__.Notify(181678)
		err = s.errWithDebug(s.Cmd.Run())
		close(commandFinished)
	}()
	__antithesis_instrumentation__.Notify(181671)

	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(181679)
		s.Close()
		err = ctx.Err()
		break
	case <-commandFinished:
		__antithesis_instrumentation__.Notify(181680)
		break
	}
	__antithesis_instrumentation__.Notify(181672)
	return err
}

func (s *remoteSession) Start(cmd string) error {
	__antithesis_instrumentation__.Notify(181681)
	if s.withExitStatus {
		__antithesis_instrumentation__.Notify(181683)
		cmd = strings.TrimSpace(cmd)
		if !strings.HasSuffix(cmd, ";") {
			__antithesis_instrumentation__.Notify(181685)
			cmd += ";"
		} else {
			__antithesis_instrumentation__.Notify(181686)
		}
		__antithesis_instrumentation__.Notify(181684)
		cmd += "echo -n 'LAST EXIT STATUS: '$?;"
	} else {
		__antithesis_instrumentation__.Notify(181687)
	}
	__antithesis_instrumentation__.Notify(181682)
	s.Cmd.Args = append(s.Cmd.Args, cmd)
	return s.Cmd.Start()
}

func (s *remoteSession) SetWithExitStatus(value bool) {
	__antithesis_instrumentation__.Notify(181688)
	s.withExitStatus = value
}

func (s *remoteSession) SetStdin(r io.Reader) {
	__antithesis_instrumentation__.Notify(181689)
	s.Stdin = r
}

func (s *remoteSession) SetStdout(w io.Writer) {
	__antithesis_instrumentation__.Notify(181690)
	s.Stdout = w
}

func (s *remoteSession) SetStderr(w io.Writer) {
	__antithesis_instrumentation__.Notify(181691)
	s.Stderr = w
}

func (s *remoteSession) StdoutPipe() (io.Reader, error) {
	__antithesis_instrumentation__.Notify(181692)

	r, err := s.Cmd.StdoutPipe()
	return r, err
}

func (s *remoteSession) StderrPipe() (io.Reader, error) {
	__antithesis_instrumentation__.Notify(181693)

	r, err := s.Cmd.StderrPipe()
	return r, err
}

func (s *remoteSession) RequestPty() error {
	__antithesis_instrumentation__.Notify(181694)
	s.Cmd.Args = append(s.Cmd.Args, "-t")
	return nil
}

func (s *remoteSession) Wait() error {
	__antithesis_instrumentation__.Notify(181695)
	return s.Cmd.Wait()
}

func (s *remoteSession) Close() {
	__antithesis_instrumentation__.Notify(181696)
	s.cancel()
	if s.logfile != "" {
		__antithesis_instrumentation__.Notify(181697)
		_ = os.Remove(s.logfile)
	} else {
		__antithesis_instrumentation__.Notify(181698)
	}
}

type localSession struct {
	*exec.Cmd
	cancel         func()
	withExitStatus bool
}

func newLocalSession() *localSession {
	__antithesis_instrumentation__.Notify(181699)
	ctx, cancel := context.WithCancel(context.Background())
	withExitStatus := false
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c")
	return &localSession{cmd, cancel, withExitStatus}
}

func (s *localSession) CombinedOutput(ctx context.Context, cmd string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(181700)
	if s.withExitStatus {
		__antithesis_instrumentation__.Notify(181704)
		cmd = strings.TrimSpace(cmd)
		if !strings.HasSuffix(cmd, ";") {
			__antithesis_instrumentation__.Notify(181706)
			cmd += ";"
		} else {
			__antithesis_instrumentation__.Notify(181707)
		}
		__antithesis_instrumentation__.Notify(181705)
		cmd += "echo -n 'LAST EXIT STATUS: '$?;"
	} else {
		__antithesis_instrumentation__.Notify(181708)
	}
	__antithesis_instrumentation__.Notify(181701)
	s.Cmd.Args = append(s.Cmd.Args, cmd)

	var b []byte
	var err error
	commandFinished := make(chan struct{})

	go func() {
		__antithesis_instrumentation__.Notify(181709)
		b, err = s.Cmd.CombinedOutput()
		close(commandFinished)
	}()
	__antithesis_instrumentation__.Notify(181702)

	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(181710)
		s.Close()
		err = ctx.Err()
		break
	case <-commandFinished:
		__antithesis_instrumentation__.Notify(181711)
		break
	}
	__antithesis_instrumentation__.Notify(181703)

	return b, err
}

func (s *localSession) Run(ctx context.Context, cmd string) error {
	__antithesis_instrumentation__.Notify(181712)
	if s.withExitStatus {
		__antithesis_instrumentation__.Notify(181716)
		cmd = strings.TrimSpace(cmd)
		if !strings.HasSuffix(cmd, ";") {
			__antithesis_instrumentation__.Notify(181718)
			cmd += ";"
		} else {
			__antithesis_instrumentation__.Notify(181719)
		}
		__antithesis_instrumentation__.Notify(181717)
		cmd += "echo -n 'LAST EXIT STATUS: '$?;"
	} else {
		__antithesis_instrumentation__.Notify(181720)
	}
	__antithesis_instrumentation__.Notify(181713)
	s.Cmd.Args = append(s.Cmd.Args, cmd)

	var err error
	commandFinished := make(chan struct{})
	go func() {
		__antithesis_instrumentation__.Notify(181721)
		err = s.Cmd.Run()
		close(commandFinished)
	}()
	__antithesis_instrumentation__.Notify(181714)

	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(181722)
		s.Close()
		err = ctx.Err()
		break
	case <-commandFinished:
		__antithesis_instrumentation__.Notify(181723)
		break
	}
	__antithesis_instrumentation__.Notify(181715)
	return err
}

func (s *localSession) Start(cmd string) error {
	__antithesis_instrumentation__.Notify(181724)
	if s.withExitStatus {
		__antithesis_instrumentation__.Notify(181726)
		cmd = strings.TrimSpace(cmd)
		if !strings.HasSuffix(cmd, ";") {
			__antithesis_instrumentation__.Notify(181728)
			cmd += ";"
		} else {
			__antithesis_instrumentation__.Notify(181729)
		}
		__antithesis_instrumentation__.Notify(181727)
		cmd += "echo -n 'LAST EXIT STATUS: '$?;"
	} else {
		__antithesis_instrumentation__.Notify(181730)
	}
	__antithesis_instrumentation__.Notify(181725)
	s.Cmd.Args = append(s.Cmd.Args, cmd)
	return s.Cmd.Start()
}

func (s *localSession) SetWithExitStatus(value bool) {
	__antithesis_instrumentation__.Notify(181731)
	s.withExitStatus = value
}

func (s *localSession) SetStdin(r io.Reader) {
	__antithesis_instrumentation__.Notify(181732)
	s.Stdin = r
}

func (s *localSession) SetStdout(w io.Writer) {
	__antithesis_instrumentation__.Notify(181733)
	s.Stdout = w
}

func (s *localSession) SetStderr(w io.Writer) {
	__antithesis_instrumentation__.Notify(181734)
	s.Stderr = w
}

func (s *localSession) StdoutPipe() (io.Reader, error) {
	__antithesis_instrumentation__.Notify(181735)

	r, err := s.Cmd.StdoutPipe()
	return r, err
}

func (s *localSession) StderrPipe() (io.Reader, error) {
	__antithesis_instrumentation__.Notify(181736)

	r, err := s.Cmd.StderrPipe()
	return r, err
}

func (s *localSession) RequestPty() error {
	__antithesis_instrumentation__.Notify(181737)
	return nil
}

func (s *localSession) Wait() error {
	__antithesis_instrumentation__.Notify(181738)
	return s.Cmd.Wait()
}

func (s *localSession) Close() {
	__antithesis_instrumentation__.Notify(181739)
	s.cancel()
}

var sshAuthArgsVal []string
var sshAuthArgsOnce sync.Once

func sshAuthArgs() []string {
	__antithesis_instrumentation__.Notify(181740)
	sshAuthArgsOnce.Do(func() {
		__antithesis_instrumentation__.Notify(181742)
		paths := []string{
			filepath.Join(config.OSUser.HomeDir, ".ssh", "id_ed25519"),
			filepath.Join(config.OSUser.HomeDir, ".ssh", "id_rsa"),
			filepath.Join(config.OSUser.HomeDir, ".ssh", "google_compute_engine"),
		}
		for _, p := range paths {
			__antithesis_instrumentation__.Notify(181743)
			if _, err := os.Stat(p); err == nil {
				__antithesis_instrumentation__.Notify(181744)
				sshAuthArgsVal = append(sshAuthArgsVal, "-i", p)
			} else {
				__antithesis_instrumentation__.Notify(181745)
			}
		}
	})
	__antithesis_instrumentation__.Notify(181741)
	return sshAuthArgsVal
}
