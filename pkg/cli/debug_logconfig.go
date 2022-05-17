package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/util/log/logconfig"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var debugCheckLogConfigCmd = &cobra.Command{
	Use:   "check-log-config",
	Short: "test the log config passed via --log",
	RunE:  runDebugCheckLogConfig,
}

var debugLogChanSel logconfig.ChannelList

func runDebugCheckLogConfig(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(31193)
	if err := setupLogging(context.Background(), cmd,
		true, false); err != nil {
		__antithesis_instrumentation__.Notify(31197)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31198)
	}
	__antithesis_instrumentation__.Notify(31194)
	if cliCtx.ambiguousLogDir {
		__antithesis_instrumentation__.Notify(31199)
		fmt.Fprintf(stderr, "warning: ambiguous configuration, consider overriding the logging directory\n")
	} else {
		__antithesis_instrumentation__.Notify(31200)
	}
	__antithesis_instrumentation__.Notify(31195)

	c := cliCtx.logConfig
	r, err := yaml.Marshal(&c)
	if err != nil {
		__antithesis_instrumentation__.Notify(31201)
		return errors.Wrap(err, "printing configuration")
	} else {
		__antithesis_instrumentation__.Notify(31202)
	}
	__antithesis_instrumentation__.Notify(31196)

	fmt.Println("# configuration after validation:")
	fmt.Println(string(r))

	_, key := c.Export(debugLogChanSel)
	fmt.Println("# graphical diagram URL:")
	fmt.Printf("http://www.plantuml.com/plantuml/uml/%s\n", key)

	return nil
}
