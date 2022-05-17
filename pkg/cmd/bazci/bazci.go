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
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/alessio/shellescape"
	bazelutil "github.com/cockroachdb/cockroach/pkg/build/util"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

const (
	buildSubcmd         = "build"
	runSubcmd           = "run"
	testSubcmd          = "test"
	mergeTestXMLsSubcmd = "merge-test-xmls"
	mungeTestXMLSubcmd  = "munge-test-xml"
)

var (
	artifactsDir    string
	configs         []string
	compilationMode string

	rootCmd = &cobra.Command{
		Use:   "bazci",
		Short: "A glue binary for making Bazel usable in Teamcity",
		Long: `bazci is glue code to make debugging Bazel builds and
tests in Teamcity as painless as possible.`,
		Args: func(cmd *cobra.Command, args []string) error {
			__antithesis_instrumentation__.Notify(37529)
			_, err := parseArgs(args, cmd.ArgsLenAtDash())
			return err
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          bazciImpl,
	}
)

func init() {
	rootCmd.Flags().StringVar(
		&artifactsDir,
		"artifacts_dir",
		"/artifacts",
		"path where artifacts should be staged")
	rootCmd.Flags().StringVar(
		&compilationMode,
		"compilation_mode",
		"dbg",
		"compilation mode to pass down to Bazel (dbg or opt)")
	rootCmd.Flags().StringSliceVar(
		&configs,
		"config",
		[]string{"ci"},
		"list of build configs to apply to bazel calls")
}

type parsedArgs struct {
	subcmd string

	targets []string

	additional []string
}

var errUsage = errors.New("At least 2 arguments required (e.g. `bazci build TARGET`)")

func parseArgs(args []string, argsLenAtDash int) (*parsedArgs, error) {
	__antithesis_instrumentation__.Notify(37530)

	if len(args) < 2 {
		__antithesis_instrumentation__.Notify(37534)
		return nil, errUsage
	} else {
		__antithesis_instrumentation__.Notify(37535)
	}
	__antithesis_instrumentation__.Notify(37531)
	if args[0] != buildSubcmd && func() bool {
		__antithesis_instrumentation__.Notify(37536)
		return args[0] != runSubcmd == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(37537)
		return args[0] != testSubcmd == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(37538)
		return args[0] != mungeTestXMLSubcmd == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(37539)
		return args[0] != mergeTestXMLsSubcmd == true
	}() == true {
		__antithesis_instrumentation__.Notify(37540)
		return nil, errors.Newf("First argument must be `build`, `run`, `test`, `merge-test-xmls`, or `munge-test-xml`; got %v", args[0])
	} else {
		__antithesis_instrumentation__.Notify(37541)
	}
	__antithesis_instrumentation__.Notify(37532)
	var splitLoc int
	if argsLenAtDash < 0 {
		__antithesis_instrumentation__.Notify(37542)

		splitLoc = len(args)
	} else {
		__antithesis_instrumentation__.Notify(37543)
		if argsLenAtDash < 2 {
			__antithesis_instrumentation__.Notify(37544)
			return nil, errUsage
		} else {
			__antithesis_instrumentation__.Notify(37545)
			splitLoc = argsLenAtDash
		}
	}
	__antithesis_instrumentation__.Notify(37533)
	return &parsedArgs{
		subcmd:     args[0],
		targets:    args[1:splitLoc],
		additional: args[splitLoc:],
	}, nil
}

type buildInfo struct {
	binDir string

	testlogsDir string

	goBinaries []string

	cmakeTargets []string

	genruleTargets []string

	tests []string

	transitionTests map[string]string
}

func runBazelReturningStdout(subcmd string, arg ...string) (string, error) {
	__antithesis_instrumentation__.Notify(37546)
	if subcmd != "query" {
		__antithesis_instrumentation__.Notify(37549)
		var configArgs []string

		if subcmd == "cquery" {
			__antithesis_instrumentation__.Notify(37551)
			configArgs = configArgList("test")
		} else {
			__antithesis_instrumentation__.Notify(37552)
			configArgs = configArgList()
		}
		__antithesis_instrumentation__.Notify(37550)
		arg = append(configArgs, arg...)
		arg = append(arg, "-c", compilationMode)
	} else {
		__antithesis_instrumentation__.Notify(37553)
	}
	__antithesis_instrumentation__.Notify(37547)
	arg = append([]string{subcmd}, arg...)
	buf, err := exec.Command("bazel", arg...).Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(37554)
		if len(buf) > 0 {
			__antithesis_instrumentation__.Notify(37557)
			fmt.Printf("COMMAND STDOUT:\n%s\n", string(buf))
		} else {
			__antithesis_instrumentation__.Notify(37558)
		}
		__antithesis_instrumentation__.Notify(37555)
		var cmderr exec.ExitError
		if errors.As(err, &cmderr) && func() bool {
			__antithesis_instrumentation__.Notify(37559)
			return len(cmderr.Stderr) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(37560)
			fmt.Printf("COMMAND STDERR:\n%s\n", string(cmderr.Stderr))
		} else {
			__antithesis_instrumentation__.Notify(37561)
		}
		__antithesis_instrumentation__.Notify(37556)
		fmt.Println("Failed to run Bazel with args: ", arg)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(37562)
	}
	__antithesis_instrumentation__.Notify(37548)
	return strings.TrimSpace(string(buf)), nil
}

func getBuildInfo(args parsedArgs) (buildInfo, error) {
	__antithesis_instrumentation__.Notify(37563)
	if args.subcmd != buildSubcmd && func() bool {
		__antithesis_instrumentation__.Notify(37568)
		return args.subcmd != runSubcmd == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(37569)
		return args.subcmd != testSubcmd == true
	}() == true {
		__antithesis_instrumentation__.Notify(37570)
		return buildInfo{}, errors.Newf("Unexpected subcommand %s. This is a bug!", args.subcmd)
	} else {
		__antithesis_instrumentation__.Notify(37571)
	}
	__antithesis_instrumentation__.Notify(37564)
	binDir, err := runBazelReturningStdout("info", "bazel-bin")
	if err != nil {
		__antithesis_instrumentation__.Notify(37572)
		return buildInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(37573)
	}
	__antithesis_instrumentation__.Notify(37565)
	testlogsDir, err := runBazelReturningStdout("info", "bazel-testlogs")
	if err != nil {
		__antithesis_instrumentation__.Notify(37574)
		return buildInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(37575)
	}
	__antithesis_instrumentation__.Notify(37566)

	ret := buildInfo{
		binDir:          binDir,
		testlogsDir:     testlogsDir,
		transitionTests: make(map[string]string),
	}

	for _, target := range args.targets {
		__antithesis_instrumentation__.Notify(37576)
		output, err := runBazelReturningStdout("query", "--output=label_kind", target)
		if err != nil {
			__antithesis_instrumentation__.Notify(37579)
			return buildInfo{}, err
		} else {
			__antithesis_instrumentation__.Notify(37580)
		}
		__antithesis_instrumentation__.Notify(37577)

		outputSplit := strings.Fields(output)
		if len(outputSplit) != 3 {
			__antithesis_instrumentation__.Notify(37581)
			return buildInfo{}, errors.Newf("Could not parse bazel query output: %v", output)
		} else {
			__antithesis_instrumentation__.Notify(37582)
		}
		__antithesis_instrumentation__.Notify(37578)
		targetKind := outputSplit[0]
		fullTarget := outputSplit[2]

		switch targetKind {
		case "cmake":
			__antithesis_instrumentation__.Notify(37583)
			ret.cmakeTargets = append(ret.cmakeTargets, fullTarget)
		case "genrule", "batch_gen":
			__antithesis_instrumentation__.Notify(37584)
			ret.genruleTargets = append(ret.genruleTargets, fullTarget)
		case "go_binary":
			__antithesis_instrumentation__.Notify(37585)
			ret.goBinaries = append(ret.goBinaries, fullTarget)
		case "go_test":
			__antithesis_instrumentation__.Notify(37586)
			ret.tests = append(ret.tests, fullTarget)
		case "go_transition_test":
			__antithesis_instrumentation__.Notify(37587)

			args := []string{fullTarget, "-c", compilationMode, "--run_under=realpath"}
			runOutput, err := runBazelReturningStdout("run", args...)
			if err != nil {
				__antithesis_instrumentation__.Notify(37594)
				return buildInfo{}, err
			} else {
				__antithesis_instrumentation__.Notify(37595)
			}
			__antithesis_instrumentation__.Notify(37588)
			var binLocation string
			for _, line := range strings.Split(runOutput, "\n") {
				__antithesis_instrumentation__.Notify(37596)
				if strings.HasPrefix(line, "/") {
					__antithesis_instrumentation__.Notify(37597)

					binLocation = strings.TrimSpace(line)
				} else {
					__antithesis_instrumentation__.Notify(37598)
				}
			}
			__antithesis_instrumentation__.Notify(37589)
			componentsBinLocation := strings.Split(binLocation, "/")
			componentsTestlogs := strings.Split(testlogsDir, "/")

			componentsTestlogs[len(componentsTestlogs)-2] = componentsBinLocation[len(componentsTestlogs)-2]
			ret.transitionTests[fullTarget] = strings.Join(componentsTestlogs, "/")
		case "nodejs_test":
			__antithesis_instrumentation__.Notify(37590)
			ret.tests = append(ret.tests, fullTarget)
		case "test_suite":
			__antithesis_instrumentation__.Notify(37591)

			allTests, err := runBazelReturningStdout("query", "tests("+fullTarget+")")
			if err != nil {
				__antithesis_instrumentation__.Notify(37599)
				return buildInfo{}, err
			} else {
				__antithesis_instrumentation__.Notify(37600)
			}
			__antithesis_instrumentation__.Notify(37592)
			ret.tests = append(ret.tests, strings.Fields(allTests)...)
		default:
			__antithesis_instrumentation__.Notify(37593)
			return buildInfo{}, errors.Newf("Got unexpected target kind %v", targetKind)
		}
	}
	__antithesis_instrumentation__.Notify(37567)

	return ret, nil
}

func bazciImpl(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(37601)
	parsedArgs, err := parseArgs(args, cmd.ArgsLenAtDash())
	if err != nil {
		__antithesis_instrumentation__.Notify(37607)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37608)
	}
	__antithesis_instrumentation__.Notify(37602)

	if parsedArgs.subcmd == mungeTestXMLSubcmd {
		__antithesis_instrumentation__.Notify(37609)
		return mungeTestXMLs(*parsedArgs)
	} else {
		__antithesis_instrumentation__.Notify(37610)
	}
	__antithesis_instrumentation__.Notify(37603)
	if parsedArgs.subcmd == mergeTestXMLsSubcmd {
		__antithesis_instrumentation__.Notify(37611)
		return mergeTestXMLs(*parsedArgs)
	} else {
		__antithesis_instrumentation__.Notify(37612)
	}
	__antithesis_instrumentation__.Notify(37604)

	info, err := getBuildInfo(*parsedArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(37613)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37614)
	}
	__antithesis_instrumentation__.Notify(37605)

	completion := make(chan error)
	go func() {
		__antithesis_instrumentation__.Notify(37615)
		processArgs := []string{parsedArgs.subcmd}
		processArgs = append(processArgs, parsedArgs.targets...)
		processArgs = append(processArgs, configArgList()...)
		processArgs = append(processArgs, "-c", compilationMode)
		processArgs = append(processArgs, parsedArgs.additional...)
		fmt.Println("running bazel w/ args: ", shellescape.QuoteCommand(processArgs))
		cmd := exec.Command("bazel", processArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			__antithesis_instrumentation__.Notify(37617)
			completion <- err
			return
		} else {
			__antithesis_instrumentation__.Notify(37618)
		}
		__antithesis_instrumentation__.Notify(37616)
		completion <- cmd.Wait()
	}()
	__antithesis_instrumentation__.Notify(37606)

	return makeWatcher(completion, info).Watch()
}

func mungeTestXMLs(args parsedArgs) error {
	__antithesis_instrumentation__.Notify(37619)
	for _, file := range args.targets {
		__antithesis_instrumentation__.Notify(37621)
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			__antithesis_instrumentation__.Notify(37624)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37625)
		}
		__antithesis_instrumentation__.Notify(37622)
		var buf bytes.Buffer
		err = bazelutil.MungeTestXML(contents, &buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(37626)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37627)
		}
		__antithesis_instrumentation__.Notify(37623)
		err = ioutil.WriteFile(file, buf.Bytes(), 0666)
		if err != nil {
			__antithesis_instrumentation__.Notify(37628)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37629)
		}
	}
	__antithesis_instrumentation__.Notify(37620)
	return nil
}

func mergeTestXMLs(args parsedArgs) error {
	__antithesis_instrumentation__.Notify(37630)
	var xmlsToMerge []bazelutil.TestSuites
	for _, file := range args.targets {
		__antithesis_instrumentation__.Notify(37632)
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			__antithesis_instrumentation__.Notify(37635)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37636)
		}
		__antithesis_instrumentation__.Notify(37633)
		var testSuites bazelutil.TestSuites
		err = xml.Unmarshal(contents, &testSuites)
		if err != nil {
			__antithesis_instrumentation__.Notify(37637)
			return err
		} else {
			__antithesis_instrumentation__.Notify(37638)
		}
		__antithesis_instrumentation__.Notify(37634)
		xmlsToMerge = append(xmlsToMerge, testSuites)
	}
	__antithesis_instrumentation__.Notify(37631)
	return bazelutil.MergeTestXMLs(xmlsToMerge, os.Stdout)
}

func configArgList(exceptions ...string) []string {
	__antithesis_instrumentation__.Notify(37639)
	ret := []string{}
	for _, config := range configs {
		__antithesis_instrumentation__.Notify(37641)
		keep := true
		for _, exception := range exceptions {
			__antithesis_instrumentation__.Notify(37643)
			if config == exception {
				__antithesis_instrumentation__.Notify(37644)
				keep = false
				break
			} else {
				__antithesis_instrumentation__.Notify(37645)
			}
		}
		__antithesis_instrumentation__.Notify(37642)
		if keep {
			__antithesis_instrumentation__.Notify(37646)
			ret = append(ret, "--config="+config)
		} else {
			__antithesis_instrumentation__.Notify(37647)
		}
	}
	__antithesis_instrumentation__.Notify(37640)
	return ret
}

func usingCrossWindowsConfig() bool {
	__antithesis_instrumentation__.Notify(37648)
	for _, config := range configs {
		__antithesis_instrumentation__.Notify(37650)
		if config == "crosswindows" {
			__antithesis_instrumentation__.Notify(37651)
			return true
		} else {
			__antithesis_instrumentation__.Notify(37652)
		}
	}
	__antithesis_instrumentation__.Notify(37649)
	return false
}

func usingCrossDarwinConfig() bool {
	__antithesis_instrumentation__.Notify(37653)
	for _, config := range configs {
		__antithesis_instrumentation__.Notify(37655)
		if config == "crossmacos" {
			__antithesis_instrumentation__.Notify(37656)
			return true
		} else {
			__antithesis_instrumentation__.Notify(37657)
		}
	}
	__antithesis_instrumentation__.Notify(37654)
	return false
}
