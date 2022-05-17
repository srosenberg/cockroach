package cliccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/ccl/baseccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/cliccl/cliflagsccl"
	"github.com/cockroachdb/cockroach/pkg/cli"
	"github.com/spf13/cobra"
)

var storeEncryptionSpecs baseccl.StoreEncryptionSpecList

func init() {
	for _, cmd := range cli.StartCmds {
		cli.VarFlag(cmd.Flags(), &storeEncryptionSpecs, cliflagsccl.EnterpriseEncryption)

		cli.AddPersistentPreRunE(cmd, func(cmd *cobra.Command, _ []string) error {
			return populateStoreSpecsEncryption()
		})
	}
}

func populateStoreSpecsEncryption() error {
	__antithesis_instrumentation__.Notify(19433)
	return baseccl.PopulateStoreSpecWithEncryption(cli.GetServerCfgStores(), storeEncryptionSpecs)
}
