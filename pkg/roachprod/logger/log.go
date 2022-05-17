package logger

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	crdblog "github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

const logFlags = log.Lshortfile | log.Ltime | log.LUTC

type Config struct {
	Prefix         string
	Stdout, Stderr io.Writer
}

type loggerOption interface {
	apply(*Config)
}

type logPrefix string

var _ logPrefix

func (p logPrefix) apply(cfg *Config) {
	__antithesis_instrumentation__.Notify(181821)
	cfg.Prefix = string(p)
}

type quietStdoutOption struct {
}

func (quietStdoutOption) apply(cfg *Config) {
	__antithesis_instrumentation__.Notify(181822)
	cfg.Stdout = ioutil.Discard
}

type quietStderrOption struct {
}

func (quietStderrOption) apply(cfg *Config) {
	__antithesis_instrumentation__.Notify(181823)
	cfg.Stderr = ioutil.Discard
}

var QuietStdout quietStdoutOption

var QuietStderr quietStderrOption

type Logger struct {
	path string
	File *os.File

	stdoutL, stderrL *log.Logger

	Stdout, Stderr io.Writer

	mu struct {
		syncutil.Mutex
		closed bool
	}
}

func (cfg *Config) NewLogger(path string) (*Logger, error) {
	__antithesis_instrumentation__.Notify(181824)
	if path == "" {
		__antithesis_instrumentation__.Notify(181830)

		stdout := cfg.Stdout
		if stdout == nil {
			__antithesis_instrumentation__.Notify(181833)
			stdout = os.Stdout
		} else {
			__antithesis_instrumentation__.Notify(181834)
		}
		__antithesis_instrumentation__.Notify(181831)
		stderr := cfg.Stderr
		if stderr == nil {
			__antithesis_instrumentation__.Notify(181835)
			stderr = os.Stderr
		} else {
			__antithesis_instrumentation__.Notify(181836)
		}
		__antithesis_instrumentation__.Notify(181832)
		return &Logger{
			Stdout:  stdout,
			Stderr:  stderr,
			stdoutL: log.New(os.Stdout, cfg.Prefix, logFlags),
			stderrL: log.New(os.Stderr, cfg.Prefix, logFlags),
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(181837)
	}
	__antithesis_instrumentation__.Notify(181825)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		__antithesis_instrumentation__.Notify(181838)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181839)
	}
	__antithesis_instrumentation__.Notify(181826)

	f, err := os.Create(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(181840)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(181841)
	}
	__antithesis_instrumentation__.Notify(181827)

	newWriter := func(w io.Writer) io.Writer {
		__antithesis_instrumentation__.Notify(181842)
		if w == nil {
			__antithesis_instrumentation__.Notify(181844)
			return f
		} else {
			__antithesis_instrumentation__.Notify(181845)
		}
		__antithesis_instrumentation__.Notify(181843)
		return io.MultiWriter(f, w)
	}
	__antithesis_instrumentation__.Notify(181828)

	stdout := newWriter(cfg.Stdout)
	stderr := newWriter(cfg.Stderr)
	stdoutL := log.New(stdout, cfg.Prefix, logFlags)
	var stderrL *log.Logger
	if cfg.Stdout != cfg.Stderr {
		__antithesis_instrumentation__.Notify(181846)
		stderrL = log.New(stderr, cfg.Prefix, logFlags)
	} else {
		__antithesis_instrumentation__.Notify(181847)
		stderrL = stdoutL
	}
	__antithesis_instrumentation__.Notify(181829)
	return &Logger{
		path:    path,
		File:    f,
		Stdout:  stdout,
		Stderr:  stderr,
		stdoutL: stdoutL,
		stderrL: stderrL,
	}, nil
}

type TeeOptType bool

const (
	TeeToStdout TeeOptType = true

	NoTee TeeOptType = false
)

func RootLogger(path string, teeOpt TeeOptType) (*Logger, error) {
	__antithesis_instrumentation__.Notify(181848)
	var stdout, stderr io.Writer
	if teeOpt == TeeToStdout {
		__antithesis_instrumentation__.Notify(181850)
		stdout = os.Stdout
		stderr = os.Stderr
	} else {
		__antithesis_instrumentation__.Notify(181851)
	}
	__antithesis_instrumentation__.Notify(181849)
	cfg := &Config{Stdout: stdout, Stderr: stderr}
	return cfg.NewLogger(path)
}

func (l *Logger) Close() {
	__antithesis_instrumentation__.Notify(181852)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.mu.closed {
		__antithesis_instrumentation__.Notify(181854)
		return
	} else {
		__antithesis_instrumentation__.Notify(181855)
	}
	__antithesis_instrumentation__.Notify(181853)
	l.mu.closed = true
	if l.File != nil {
		__antithesis_instrumentation__.Notify(181856)
		l.File.Close()
		l.File = nil
	} else {
		__antithesis_instrumentation__.Notify(181857)
	}
}

func (l *Logger) Closed() bool {
	__antithesis_instrumentation__.Notify(181858)
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.mu.closed
}

func (l *Logger) ChildLogger(name string, opts ...loggerOption) (*Logger, error) {
	__antithesis_instrumentation__.Notify(181859)

	if l.File == nil {
		__antithesis_instrumentation__.Notify(181863)
		p := name + ": "

		stdoutL := log.New(l.Stdout, p, logFlags)
		var stderrL *log.Logger
		if l.Stdout != l.Stderr {
			__antithesis_instrumentation__.Notify(181865)
			stderrL = log.New(l.Stderr, p, logFlags)
		} else {
			__antithesis_instrumentation__.Notify(181866)
			stderrL = stdoutL
		}
		__antithesis_instrumentation__.Notify(181864)
		return &Logger{
			path:    l.path,
			Stdout:  l.Stdout,
			Stderr:  l.Stderr,
			stdoutL: stdoutL,
			stderrL: stderrL,
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(181867)
	}
	__antithesis_instrumentation__.Notify(181860)

	cfg := &Config{
		Prefix: name + ": ",
		Stdout: l.Stdout,
		Stderr: l.Stderr,
	}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(181868)
		opt.apply(cfg)
	}
	__antithesis_instrumentation__.Notify(181861)

	var path string
	if l.path != "" {
		__antithesis_instrumentation__.Notify(181869)
		path = filepath.Join(filepath.Dir(l.path), name+".log")
	} else {
		__antithesis_instrumentation__.Notify(181870)
	}
	__antithesis_instrumentation__.Notify(181862)
	return cfg.NewLogger(path)
}

func (l *Logger) PrintfCtx(ctx context.Context, f string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(181871)
	l.PrintfCtxDepth(ctx, 2, f, args...)
}

func (l *Logger) Printf(f string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(181872)
	l.PrintfCtxDepth(context.Background(), 2, f, args...)
}

func (l *Logger) PrintfCtxDepth(ctx context.Context, depth int, f string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(181873)
	msg := crdblog.FormatWithContextTags(ctx, f, args...)
	if err := l.stdoutL.Output(depth+1, msg); err != nil {
		__antithesis_instrumentation__.Notify(181874)

		_ = log.Output(depth+1, fmt.Sprintf("failed to log message: %v: %s", err, msg))
	} else {
		__antithesis_instrumentation__.Notify(181875)
	}
}

func (l *Logger) ErrorfCtx(ctx context.Context, f string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(181876)
	l.ErrorfCtxDepth(ctx, 2, f, args...)
}

func (l *Logger) ErrorfCtxDepth(ctx context.Context, depth int, f string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(181877)
	msg := crdblog.FormatWithContextTags(ctx, f, args...)
	if err := l.stderrL.Output(depth+1, msg); err != nil {
		__antithesis_instrumentation__.Notify(181878)

		_ = log.Output(depth+1, fmt.Sprintf("failed to log error: %v: %s", err, msg))
	} else {
		__antithesis_instrumentation__.Notify(181879)
	}
}

func (l *Logger) Errorf(f string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(181880)
	l.ErrorfCtxDepth(context.Background(), 2, f, args...)
}
