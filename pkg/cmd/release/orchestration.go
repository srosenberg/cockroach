package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	versionFlag      = "version"
	templatesDirFlag = "template-dir"
	outputDirFlag    = "output-dir"
)

var orchestrationFlags = struct {
	version      string
	templatesDir string
	outputDir    string
}{}

var setOrchestrationVersionCmd = &cobra.Command{
	Use:   "set-orchestration-version",
	Short: "Set orchestration version",
	Long:  "Updates orchestration version under the ./cloud/kubernetes directory",
	RunE:  setOrchestrationVersion,
}

func init() {
	setOrchestrationVersionCmd.Flags().StringVar(&orchestrationFlags.version, versionFlag, "", "cockroachdb version")
	setOrchestrationVersionCmd.Flags().StringVar(&orchestrationFlags.templatesDir, templatesDirFlag, "",
		"orchestration templates directory")
	setOrchestrationVersionCmd.Flags().StringVar(&orchestrationFlags.outputDir, outputDirFlag, "",
		"orchestration directory")
	_ = setOrchestrationVersionCmd.MarkFlagRequired(versionFlag)
	_ = setOrchestrationVersionCmd.MarkFlagRequired(templatesDirFlag)
	_ = setOrchestrationVersionCmd.MarkFlagRequired(outputDirFlag)
}

func setOrchestrationVersion(_ *cobra.Command, _ []string) error {
	__antithesis_instrumentation__.Notify(42807)
	dirInfo, err := os.Stat(orchestrationFlags.templatesDir)
	if err != nil {
		__antithesis_instrumentation__.Notify(42810)
		return fmt.Errorf("cannot stat templates directory: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42811)
	}
	__antithesis_instrumentation__.Notify(42808)
	if !dirInfo.IsDir() {
		__antithesis_instrumentation__.Notify(42812)
		return fmt.Errorf("%s is not a directory", orchestrationFlags.templatesDir)
	} else {
		__antithesis_instrumentation__.Notify(42813)
	}
	__antithesis_instrumentation__.Notify(42809)
	return filepath.Walk(orchestrationFlags.templatesDir, func(filePath string, fileInfo os.FileInfo, e error) error {
		__antithesis_instrumentation__.Notify(42814)
		if e != nil {
			__antithesis_instrumentation__.Notify(42822)
			return e
		} else {
			__antithesis_instrumentation__.Notify(42823)
		}
		__antithesis_instrumentation__.Notify(42815)

		if !fileInfo.Mode().IsRegular() {
			__antithesis_instrumentation__.Notify(42824)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(42825)
		}
		__antithesis_instrumentation__.Notify(42816)

		dir := path.Dir(filePath)
		relDir, err := filepath.Rel(orchestrationFlags.templatesDir, dir)
		if err != nil {
			__antithesis_instrumentation__.Notify(42826)
			return err
		} else {
			__antithesis_instrumentation__.Notify(42827)
		}
		__antithesis_instrumentation__.Notify(42817)
		destDir := filepath.Join(orchestrationFlags.outputDir, relDir)
		destFile := filepath.Join(destDir, fileInfo.Name())
		if err := os.MkdirAll(destDir, 0755); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(42828)
			return !errors.Is(err, os.ErrExist) == true
		}() == true {
			__antithesis_instrumentation__.Notify(42829)
			return err
		} else {
			__antithesis_instrumentation__.Notify(42830)
		}
		__antithesis_instrumentation__.Notify(42818)
		contents, err := ioutil.ReadFile(filePath)
		if err != nil {
			__antithesis_instrumentation__.Notify(42831)
			return err
		} else {
			__antithesis_instrumentation__.Notify(42832)
		}
		__antithesis_instrumentation__.Notify(42819)

		generatedContents := strings.ReplaceAll(string(contents), "@VERSION@", orchestrationFlags.version)
		if strings.HasSuffix(destFile, ".yaml") {
			__antithesis_instrumentation__.Notify(42833)
			generatedContents = fmt.Sprintf("# Generated file, DO NOT EDIT. Source: %s\n", filePath) + generatedContents
		} else {
			__antithesis_instrumentation__.Notify(42834)
		}
		__antithesis_instrumentation__.Notify(42820)
		err = ioutil.WriteFile(destFile, []byte(generatedContents), fileInfo.Mode().Perm())
		if err != nil {
			__antithesis_instrumentation__.Notify(42835)
			return err
		} else {
			__antithesis_instrumentation__.Notify(42836)
		}
		__antithesis_instrumentation__.Notify(42821)
		return nil
	})
}
