package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cmds []*cobra.Command
var quiet bool

func main() {
	__antithesis_instrumentation__.Notify(39997)
	rootCmd := func() *cobra.Command {
		__antithesis_instrumentation__.Notify(39999)
		cmd := &cobra.Command{
			Use:   "docgen",
			Short: "docgen generates documentation for cockroachdb's SQL functions, grammar, and HTTP endpoints",
		}
		cmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress output where possible")
		cmd.AddCommand(cmds...)
		return cmd
	}()
	__antithesis_instrumentation__.Notify(39998)
	if err := rootCmd.Execute(); err != nil {
		__antithesis_instrumentation__.Notify(40000)
		fmt.Println(err)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(40001)
	}
}
