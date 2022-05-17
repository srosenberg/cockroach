// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	bazelutil "github.com/cockroachdb/cockroach/pkg/build/util"
	"github.com/cockroachdb/errors"
)

type fileMetadata struct {
	size    int64
	modTime time.Time
}

type Phase int

const (
	initialCachingPhase Phase = iota

	incrementalUpdatePhase

	finalizePhase
)

type watcher struct {
	completion <-chan error
	info       buildInfo

	fileToMeta map[string]fileMetadata

	fileToStaged map[string]struct{}
}

func makeWatcher(completion <-chan error, info buildInfo) *watcher {
	__antithesis_instrumentation__.Notify(37664)
	return &watcher{
		completion:   completion,
		info:         info,
		fileToMeta:   make(map[string]fileMetadata),
		fileToStaged: make(map[string]struct{}),
	}
}

func (w watcher) Watch() error {
	__antithesis_instrumentation__.Notify(37665)

	err := w.stageTestArtifacts(initialCachingPhase)
	if err != nil {
		__antithesis_instrumentation__.Notify(37667)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37668)
	}
	__antithesis_instrumentation__.Notify(37666)

	for {
		__antithesis_instrumentation__.Notify(37669)
		select {
		case buildErr := <-w.completion:
			__antithesis_instrumentation__.Notify(37670)

			testErr := w.stageTestArtifacts(finalizePhase)

			binErr := w.stageBinaryArtifacts()

			if buildErr != nil {
				__antithesis_instrumentation__.Notify(37675)
				return buildErr
			} else {
				__antithesis_instrumentation__.Notify(37676)
			}
			__antithesis_instrumentation__.Notify(37671)
			if testErr != nil {
				__antithesis_instrumentation__.Notify(37677)
				return testErr
			} else {
				__antithesis_instrumentation__.Notify(37678)
			}
			__antithesis_instrumentation__.Notify(37672)
			if binErr != nil {
				__antithesis_instrumentation__.Notify(37679)
				return binErr
			} else {
				__antithesis_instrumentation__.Notify(37680)
			}
			__antithesis_instrumentation__.Notify(37673)
			return nil
		case <-time.After(10 * time.Second):
			__antithesis_instrumentation__.Notify(37674)

			err := w.stageTestArtifacts(incrementalUpdatePhase)
			if err != nil {
				__antithesis_instrumentation__.Notify(37681)
				return err
			} else {
				__antithesis_instrumentation__.Notify(37682)
			}
		}
	}
}

func (w watcher) stageTestArtifacts(phase Phase) error {
	__antithesis_instrumentation__.Notify(37683)
	targetToRelDir := func(target string) string {
		__antithesis_instrumentation__.Notify(37687)
		return strings.ReplaceAll(strings.TrimPrefix(target, "//"), ":", "/")
	}
	__antithesis_instrumentation__.Notify(37684)
	for _, test := range w.info.tests {
		__antithesis_instrumentation__.Notify(37688)

		relDir := targetToRelDir(test)
		for _, tup := range []struct {
			relPath string
			stagefn func(srcContent []byte, outFile io.Writer) error
		}{
			{path.Join(relDir, "test.log"), copyContentTo},
			{path.Join(relDir, "*", "test.log"), copyContentTo},
			{path.Join(relDir, "test.xml"), bazelutil.MungeTestXML},
			{path.Join(relDir, "*", "test.xml"), bazelutil.MungeTestXML},
		} {
			__antithesis_instrumentation__.Notify(37689)
			err := w.maybeStageArtifact(w.info.testlogsDir, tup.relPath, 0644, phase,
				tup.stagefn)
			if err != nil {
				__antithesis_instrumentation__.Notify(37690)
				return err
			} else {
				__antithesis_instrumentation__.Notify(37691)
			}
		}
	}
	__antithesis_instrumentation__.Notify(37685)

	for test, transitionTestLogsDir := range w.info.transitionTests {
		__antithesis_instrumentation__.Notify(37692)
		relDir := targetToRelDir(test)
		for _, tup := range []struct {
			relPath string
			stagefn func(srcContent []byte, outFile io.Writer) error
		}{
			{path.Join(relDir, "test.log"), copyContentTo},
			{path.Join(relDir, "test.xml"), bazelutil.MungeTestXML},
		} {
			__antithesis_instrumentation__.Notify(37693)
			err := w.maybeStageArtifact(transitionTestLogsDir, tup.relPath, 0644, phase, tup.stagefn)
			if err != nil {
				__antithesis_instrumentation__.Notify(37694)
				return err
			} else {
				__antithesis_instrumentation__.Notify(37695)
			}
		}
	}
	__antithesis_instrumentation__.Notify(37686)
	return nil
}

func (w watcher) stageBinaryArtifacts() error {
	__antithesis_instrumentation__.Notify(37696)
	for _, bin := range w.info.goBinaries {
		__antithesis_instrumentation__.Notify(37700)
		relBinPath := bazelutil.OutputOfBinaryRule(bin, usingCrossWindowsConfig())
		err := w.maybeStageArtifact(w.info.binDir, relBinPath, 0755, finalizePhase,
			copyContentTo)
		if err != nil {
			__antithesis_instrumentation__.Notify(37701)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37702)
		}
	}
	__antithesis_instrumentation__.Notify(37697)
	for _, bin := range w.info.genruleTargets {
		__antithesis_instrumentation__.Notify(37703)

		query, err := runBazelReturningStdout("query", "--output=xml", bin)
		if err != nil {
			__antithesis_instrumentation__.Notify(37706)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37707)
		}
		__antithesis_instrumentation__.Notify(37704)
		outs, err := bazelutil.OutputsOfGenrule(bin, query)
		if err != nil {
			__antithesis_instrumentation__.Notify(37708)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37709)
		}
		__antithesis_instrumentation__.Notify(37705)
		for _, relBinPath := range outs {
			__antithesis_instrumentation__.Notify(37710)
			err := w.maybeStageArtifact(w.info.binDir, relBinPath, 0644, finalizePhase, copyContentTo)
			if err != nil {
				__antithesis_instrumentation__.Notify(37711)
				return err
			} else {
				__antithesis_instrumentation__.Notify(37712)
			}
		}
	}
	__antithesis_instrumentation__.Notify(37698)
	for _, bin := range w.info.cmakeTargets {
		__antithesis_instrumentation__.Notify(37713)

		var ext string
		libDir := "lib"
		if usingCrossWindowsConfig() {
			__antithesis_instrumentation__.Notify(37715)
			ext = "dll"

			libDir = "bin"
		} else {
			__antithesis_instrumentation__.Notify(37716)
			if usingCrossDarwinConfig() {
				__antithesis_instrumentation__.Notify(37717)
				ext = "dylib"
			} else {
				__antithesis_instrumentation__.Notify(37718)
				ext = "so"
			}
		}
		__antithesis_instrumentation__.Notify(37714)
		switch bin {
		case "//c-deps:libgeos":
			__antithesis_instrumentation__.Notify(37719)
			for _, relBinPath := range []string{
				fmt.Sprintf("c-deps/libgeos/%s/libgeos_c.%s", libDir, ext),
				fmt.Sprintf("c-deps/libgeos/%s/libgeos.%s", libDir, ext),
			} {
				__antithesis_instrumentation__.Notify(37721)
				err := w.maybeStageArtifact(w.info.binDir, relBinPath, 0644, finalizePhase, copyContentTo)
				if err != nil {
					__antithesis_instrumentation__.Notify(37722)
					return err
				} else {
					__antithesis_instrumentation__.Notify(37723)
				}
			}
		default:
			__antithesis_instrumentation__.Notify(37720)
			return errors.Newf("Unrecognized cmake target %s", bin)
		}
	}
	__antithesis_instrumentation__.Notify(37699)
	return nil
}

func copyContentTo(srcContent []byte, outFile io.Writer) error {
	__antithesis_instrumentation__.Notify(37724)
	_, err := outFile.Write(srcContent)
	return err
}

type cancelableWriter struct {
	filename string
	perm     os.FileMode
	buf      bytes.Buffer
	Canceled bool
}

var _ io.WriteCloser = (*cancelableWriter)(nil)

func (w *cancelableWriter) Write(p []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(37725)
	n, err = w.buf.Write(p)
	if err != nil {
		__antithesis_instrumentation__.Notify(37727)
		w.Canceled = true
	} else {
		__antithesis_instrumentation__.Notify(37728)
	}
	__antithesis_instrumentation__.Notify(37726)
	return
}

func (w *cancelableWriter) Close() error {
	__antithesis_instrumentation__.Notify(37729)
	if !w.Canceled {
		__antithesis_instrumentation__.Notify(37731)
		err := os.MkdirAll(path.Dir(w.filename), 0755)
		if err != nil {
			__antithesis_instrumentation__.Notify(37734)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37735)
		}
		__antithesis_instrumentation__.Notify(37732)
		f, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, w.perm)
		if err != nil {
			__antithesis_instrumentation__.Notify(37736)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37737)
		}
		__antithesis_instrumentation__.Notify(37733)
		defer f.Close()
		_, err = f.Write(w.buf.Bytes())
		return err
	} else {
		__antithesis_instrumentation__.Notify(37738)
	}
	__antithesis_instrumentation__.Notify(37730)
	return nil
}

func (w watcher) maybeStageArtifact(
	rootPath string,
	pattern string,
	perm os.FileMode,
	phase Phase,
	stagefn func(srcContent []byte, outFile io.Writer) error,
) error {
	__antithesis_instrumentation__.Notify(37739)
	stage := func(srcPath, destPath string) error {
		__antithesis_instrumentation__.Notify(37743)
		contents, err := ioutil.ReadFile(srcPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(37747)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37748)
		}
		__antithesis_instrumentation__.Notify(37744)
		writer := &cancelableWriter{filename: destPath, perm: perm}
		err = stagefn(contents, writer)
		if err != nil {
			__antithesis_instrumentation__.Notify(37749)

			writer.Canceled = true

			if phase == finalizePhase {
				__antithesis_instrumentation__.Notify(37751)
				return err
			} else {
				__antithesis_instrumentation__.Notify(37752)
			}
			__antithesis_instrumentation__.Notify(37750)
			log.Printf("WARNING: got error %v trying to stage artifact %s", err,
				destPath)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(37753)
		}
		__antithesis_instrumentation__.Notify(37745)
		err = writer.Close()
		if err != nil {
			__antithesis_instrumentation__.Notify(37754)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37755)
		}
		__antithesis_instrumentation__.Notify(37746)
		var toInsert struct{}
		w.fileToStaged[srcPath] = toInsert
		return nil
	}
	__antithesis_instrumentation__.Notify(37740)

	matches, err := filepath.Glob(path.Join(rootPath, pattern))
	if err != nil {
		__antithesis_instrumentation__.Notify(37756)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37757)
	}
	__antithesis_instrumentation__.Notify(37741)
	for _, srcPath := range matches {
		__antithesis_instrumentation__.Notify(37758)
		relPath, err := filepath.Rel(rootPath, srcPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(37763)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37764)
		}
		__antithesis_instrumentation__.Notify(37759)
		artifactsSubdir := filepath.Base(rootPath)
		if !strings.HasPrefix(artifactsSubdir, "bazel") {
			__antithesis_instrumentation__.Notify(37765)
			artifactsSubdir = "bazel-" + artifactsSubdir
		} else {
			__antithesis_instrumentation__.Notify(37766)
		}
		__antithesis_instrumentation__.Notify(37760)
		destPath := path.Join(artifactsDir, artifactsSubdir, relPath)

		stat, err := os.Stat(srcPath)

		if err != nil {
			__antithesis_instrumentation__.Notify(37767)
			continue
		} else {
			__antithesis_instrumentation__.Notify(37768)
		}
		__antithesis_instrumentation__.Notify(37761)
		meta := fileMetadata{size: stat.Size(), modTime: stat.ModTime()}
		oldMeta, oldMetaOk := w.fileToMeta[srcPath]

		if !oldMetaOk || func() bool {
			__antithesis_instrumentation__.Notify(37769)
			return meta != oldMeta == true
		}() == true {
			__antithesis_instrumentation__.Notify(37770)
			switch phase {
			case initialCachingPhase:
				__antithesis_instrumentation__.Notify(37772)

			case incrementalUpdatePhase, finalizePhase:
				__antithesis_instrumentation__.Notify(37773)
				_, staged := w.fileToStaged[srcPath]

				if staged && func() bool {
					__antithesis_instrumentation__.Notify(37776)
					return !strings.HasSuffix(destPath, ".log") == true
				}() == true {
					__antithesis_instrumentation__.Notify(37777)
					log.Printf("WARNING: re-staging already-staged file at %s",
						destPath)
				} else {
					__antithesis_instrumentation__.Notify(37778)
				}
				__antithesis_instrumentation__.Notify(37774)
				err := stage(srcPath, destPath)
				if err != nil {
					__antithesis_instrumentation__.Notify(37779)
					return err
				} else {
					__antithesis_instrumentation__.Notify(37780)
				}
			default:
				__antithesis_instrumentation__.Notify(37775)
			}
			__antithesis_instrumentation__.Notify(37771)
			w.fileToMeta[srcPath] = meta
		} else {
			__antithesis_instrumentation__.Notify(37781)
		}
		__antithesis_instrumentation__.Notify(37762)
		_, staged := w.fileToStaged[srcPath]

		if !staged && func() bool {
			__antithesis_instrumentation__.Notify(37782)
			return phase == finalizePhase == true
		}() == true {
			__antithesis_instrumentation__.Notify(37783)
			err := stage(srcPath, destPath)
			if err != nil {
				__antithesis_instrumentation__.Notify(37784)
				return err
			} else {
				__antithesis_instrumentation__.Notify(37785)
			}
		} else {
			__antithesis_instrumentation__.Notify(37786)
		}
	}
	__antithesis_instrumentation__.Notify(37742)
	return nil
}
