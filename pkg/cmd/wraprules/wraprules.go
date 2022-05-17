package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func extractGroups(path string) ([]interface{}, error) {
	__antithesis_instrumentation__.Notify(53362)
	var ruleFile struct {
		Groups []interface{}
	}

	file, err := os.Open(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(53367)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(53368)
	}
	__antithesis_instrumentation__.Notify(53363)
	data, err := ioutil.ReadAll(file)
	if err != nil {
		__antithesis_instrumentation__.Notify(53369)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(53370)
	}
	__antithesis_instrumentation__.Notify(53364)
	if err := yaml.UnmarshalStrict(data, &ruleFile); err != nil {
		__antithesis_instrumentation__.Notify(53371)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(53372)
	}
	__antithesis_instrumentation__.Notify(53365)

	if len(ruleFile.Groups) == 0 {
		__antithesis_instrumentation__.Notify(53373)
		return nil, errors.New("did not find a top-level groups entry")
	} else {
		__antithesis_instrumentation__.Notify(53374)
	}
	__antithesis_instrumentation__.Notify(53366)
	return ruleFile.Groups, nil
}

func main() {
	__antithesis_instrumentation__.Notify(53375)
	var outFile string
	rootCmd := &cobra.Command{
		Use:     "wraprules",
		Short:   "wraprules wraps one or more promethus monitoring files into a PrometheusRule object",
		Example: "wraprules -o alert-rules.yaml monitoring/rules/*.rules.yml",
		RunE: func(cmd *cobra.Command, args []string) error {
			__antithesis_instrumentation__.Notify(53377)
			if len(outFile) == 0 {
				__antithesis_instrumentation__.Notify(53382)
				return errors.New("no output file given")
			} else {
				__antithesis_instrumentation__.Notify(53383)
			}
			__antithesis_instrumentation__.Notify(53378)
			if len(args) == 0 {
				__antithesis_instrumentation__.Notify(53384)
				return errors.New("no input file(s) given")
			} else {
				__antithesis_instrumentation__.Notify(53385)
			}
			__antithesis_instrumentation__.Notify(53379)

			var outGroups []interface{}
			for _, path := range args {
				__antithesis_instrumentation__.Notify(53386)
				extracted, err := extractGroups(path)
				if err != nil {
					__antithesis_instrumentation__.Notify(53388)
					return errors.Wrapf(err, "unable to extract from %s", path)
				} else {
					__antithesis_instrumentation__.Notify(53389)
				}
				__antithesis_instrumentation__.Notify(53387)
				outGroups = append(outGroups, extracted...)
			}
			__antithesis_instrumentation__.Notify(53380)

			type bag map[string]interface{}
			output := bag{
				"apiVersion": "monitoring.coreos.com/v1",
				"kind":       "PrometheusRule",
				"metadata": bag{
					"name": "prometheus-cockroachdb-rules",
					"labels": bag{
						"app":        "cockroachdb",
						"prometheus": "cockroachdb",
						"role":       "alert-rules",
					},
				},
				"spec": bag{
					"groups": outGroups,
				},
			}

			outBytes, err := yaml.Marshal(output)
			if err != nil {
				__antithesis_instrumentation__.Notify(53390)
				return err
			} else {
				__antithesis_instrumentation__.Notify(53391)
			}
			__antithesis_instrumentation__.Notify(53381)

			prelude := "# GENERATED FILE - DO NOT EDIT\n"
			outBytes = append([]byte(prelude), outBytes...)

			return ioutil.WriteFile(outFile, outBytes, 0666)
		},
	}
	__antithesis_instrumentation__.Notify(53376)
	rootCmd.Flags().StringVarP(&outFile, "out", "o", "", "The output file")

	if err := rootCmd.Execute(); err != nil {
		__antithesis_instrumentation__.Notify(53392)
		fmt.Println(err)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(53393)
	}
}
