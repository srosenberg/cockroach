package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/startupmigrations"
	"github.com/cockroachdb/errors/oserror"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var manPath string

var genManCmd = &cobra.Command{
	Use:   "man",
	Short: "generate man pages for CockroachDB",
	Long: `This command generates man pages for CockroachDB.

By default, this places man pages into the "man/man1" directory under the
current directory. Use "--path=PATH" to override the output directory. For
example, to install man pages globally on many Unix-like systems,
use "--path=/usr/local/share/man/man1".
`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runGenManCmd),
}

func runGenManCmd(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(32954)
	info := build.GetInfo()
	header := &doc.GenManHeader{
		Section: "1",
		Manual:  "CockroachDB Manual",
		Source:  fmt.Sprintf("CockroachDB %s", info.Tag),
	}

	if !strings.HasSuffix(manPath, string(os.PathSeparator)) {
		__antithesis_instrumentation__.Notify(32958)
		manPath += string(os.PathSeparator)
	} else {
		__antithesis_instrumentation__.Notify(32959)
	}
	__antithesis_instrumentation__.Notify(32955)

	if _, err := os.Stat(manPath); err != nil {
		__antithesis_instrumentation__.Notify(32960)
		if oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(32961)
			if err := os.MkdirAll(manPath, 0755); err != nil {
				__antithesis_instrumentation__.Notify(32962)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32963)
			}
		} else {
			__antithesis_instrumentation__.Notify(32964)
			return err
		}
	} else {
		__antithesis_instrumentation__.Notify(32965)
	}
	__antithesis_instrumentation__.Notify(32956)

	if err := doc.GenManTree(cmd.Root(), header, manPath); err != nil {
		__antithesis_instrumentation__.Notify(32966)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32967)
	}
	__antithesis_instrumentation__.Notify(32957)

	fmt.Println("Generated CockroachDB man pages in", manPath)
	return nil
}

var autoCompletePath string

var genAutocompleteCmd = &cobra.Command{
	Use:   "autocomplete [shell]",
	Short: "generate autocompletion script for CockroachDB",
	Long: `Generate autocompletion script for CockroachDB.

If no arguments are passed, or if 'bash' is passed, a bash completion file is
written to ./cockroach.bash. If 'fish' is passed, a fish completion file
is written to ./cockroach.fish. If 'zsh' is passed, a zsh completion file is written
to ./_cockroach. Use "--out=/path/to/file" to override the output file location.

Note that for the generated file to work on OS X with bash, you'll need to install
Homebrew's bash-completion package (or an equivalent) and follow the post-install
instructions.
`,
	Args:      cobra.OnlyValidArgs,
	ValidArgs: []string{"bash", "zsh", "fish"},
	RunE:      clierrorplus.MaybeDecorateError(runGenAutocompleteCmd),
}

func runGenAutocompleteCmd(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(32968)
	var shell string
	if len(args) > 0 {
		__antithesis_instrumentation__.Notify(32972)
		shell = args[0]
	} else {
		__antithesis_instrumentation__.Notify(32973)
		shell = "bash"
	}
	__antithesis_instrumentation__.Notify(32969)

	var err error
	switch shell {
	case "bash":
		__antithesis_instrumentation__.Notify(32974)
		if autoCompletePath == "" {
			__antithesis_instrumentation__.Notify(32981)
			autoCompletePath = "cockroach.bash"
		} else {
			__antithesis_instrumentation__.Notify(32982)
		}
		__antithesis_instrumentation__.Notify(32975)
		err = cmd.Root().GenBashCompletionFile(autoCompletePath)
	case "fish":
		__antithesis_instrumentation__.Notify(32976)
		if autoCompletePath == "" {
			__antithesis_instrumentation__.Notify(32983)
			autoCompletePath = "cockroach.fish"
		} else {
			__antithesis_instrumentation__.Notify(32984)
		}
		__antithesis_instrumentation__.Notify(32977)
		err = cmd.Root().GenFishCompletionFile(autoCompletePath, true)
	case "zsh":
		__antithesis_instrumentation__.Notify(32978)
		if autoCompletePath == "" {
			__antithesis_instrumentation__.Notify(32985)
			autoCompletePath = "_cockroach"
		} else {
			__antithesis_instrumentation__.Notify(32986)
		}
		__antithesis_instrumentation__.Notify(32979)
		err = cmd.Root().GenZshCompletionFile(autoCompletePath)
	default:
		__antithesis_instrumentation__.Notify(32980)
	}
	__antithesis_instrumentation__.Notify(32970)
	if err != nil {
		__antithesis_instrumentation__.Notify(32987)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32988)
	}
	__antithesis_instrumentation__.Notify(32971)

	fmt.Printf("Generated %s completion file: %s\n", shell, autoCompletePath)
	return nil
}

var aesSize int
var overwriteKey bool

var genEncryptionKeyCmd = &cobra.Command{
	Use:   "encryption-key <key-file>",
	Short: "generate store key for encryption at rest",
	Long: `Generate store key for encryption at rest.

Generates a key suitable for use as a store key for Encryption At Rest.
The resulting key file will be 32 bytes (random key ID) + key_size in bytes.
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(32989)
		encryptionKeyPath := args[0]

		if aesSize != 128 && func() bool {
			__antithesis_instrumentation__.Notify(32997)
			return aesSize != 192 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(32998)
			return aesSize != 256 == true
		}() == true {
			__antithesis_instrumentation__.Notify(32999)
			return fmt.Errorf("store key size should be 128, 192, or 256 bits, got %d", aesSize)
		} else {
			__antithesis_instrumentation__.Notify(33000)
		}
		__antithesis_instrumentation__.Notify(32990)

		kSize := aesSize/8 + 32
		b := make([]byte, kSize)
		if _, err := rand.Read(b); err != nil {
			__antithesis_instrumentation__.Notify(33001)
			return fmt.Errorf("failed to create key with size %d bytes", kSize)
		} else {
			__antithesis_instrumentation__.Notify(33002)
		}
		__antithesis_instrumentation__.Notify(32991)

		openMode := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		if !overwriteKey {
			__antithesis_instrumentation__.Notify(33003)
			openMode |= os.O_EXCL
		} else {
			__antithesis_instrumentation__.Notify(33004)
		}
		__antithesis_instrumentation__.Notify(32992)

		f, err := os.OpenFile(encryptionKeyPath, openMode, 0600)
		if err != nil {
			__antithesis_instrumentation__.Notify(33005)
			return err
		} else {
			__antithesis_instrumentation__.Notify(33006)
		}
		__antithesis_instrumentation__.Notify(32993)
		n, err := f.Write(b)
		if err == nil && func() bool {
			__antithesis_instrumentation__.Notify(33007)
			return n < len(b) == true
		}() == true {
			__antithesis_instrumentation__.Notify(33008)
			err = io.ErrShortWrite
		} else {
			__antithesis_instrumentation__.Notify(33009)
		}
		__antithesis_instrumentation__.Notify(32994)
		if err1 := f.Close(); err == nil {
			__antithesis_instrumentation__.Notify(33010)
			err = err1
		} else {
			__antithesis_instrumentation__.Notify(33011)
		}
		__antithesis_instrumentation__.Notify(32995)

		if err != nil {
			__antithesis_instrumentation__.Notify(33012)
			return err
		} else {
			__antithesis_instrumentation__.Notify(33013)
		}
		__antithesis_instrumentation__.Notify(32996)

		fmt.Printf("successfully created AES-%d key: %s\n", aesSize, encryptionKeyPath)
		return nil
	},
}

var includeReservedSettings bool
var excludeSystemSettings bool

var genSettingsListCmd = &cobra.Command{
	Use:   "settings-list",
	Short: "output a list of available cluster settings",
	Long: `
Output the list of cluster settings known to this binary.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(33014)
		wrapCode := func(s string) string {
			__antithesis_instrumentation__.Notify(33017)
			if sqlExecCtx.TableDisplayFormat == clisqlexec.TableDisplayRawHTML {
				__antithesis_instrumentation__.Notify(33019)
				return fmt.Sprintf("<code>%s</code>", s)
			} else {
				__antithesis_instrumentation__.Notify(33020)
			}
			__antithesis_instrumentation__.Notify(33018)
			return s
		}
		__antithesis_instrumentation__.Notify(33015)

		s := cluster.MakeTestingClusterSettings()
		settings.NewUpdater(&s.SV).ResetRemaining(context.Background())

		var rows [][]string
		for _, name := range settings.Keys(settings.ForSystemTenant) {
			__antithesis_instrumentation__.Notify(33021)
			setting, ok := settings.Lookup(name, settings.LookupForLocalAccess, settings.ForSystemTenant)
			if !ok {
				__antithesis_instrumentation__.Notify(33027)
				panic(fmt.Sprintf("could not find setting %q", name))
			} else {
				__antithesis_instrumentation__.Notify(33028)
			}
			__antithesis_instrumentation__.Notify(33022)

			if excludeSystemSettings && func() bool {
				__antithesis_instrumentation__.Notify(33029)
				return setting.Class() == settings.SystemOnly == true
			}() == true {
				__antithesis_instrumentation__.Notify(33030)
				continue
			} else {
				__antithesis_instrumentation__.Notify(33031)
			}
			__antithesis_instrumentation__.Notify(33023)

			if setting.Visibility() != settings.Public {
				__antithesis_instrumentation__.Notify(33032)

				continue
			} else {
				__antithesis_instrumentation__.Notify(33033)
			}
			__antithesis_instrumentation__.Notify(33024)

			typ, ok := settings.ReadableTypes[setting.Typ()]
			if !ok {
				__antithesis_instrumentation__.Notify(33034)
				panic(fmt.Sprintf("unknown setting type %q", setting.Typ()))
			} else {
				__antithesis_instrumentation__.Notify(33035)
			}
			__antithesis_instrumentation__.Notify(33025)
			var defaultVal string
			if sm, ok := setting.(*settings.VersionSetting); ok {
				__antithesis_instrumentation__.Notify(33036)
				defaultVal = sm.SettingsListDefault()
			} else {
				__antithesis_instrumentation__.Notify(33037)
				defaultVal = setting.String(&s.SV)
				if override, ok := startupmigrations.SettingsDefaultOverrides[name]; ok {
					__antithesis_instrumentation__.Notify(33038)
					defaultVal = override
				} else {
					__antithesis_instrumentation__.Notify(33039)
				}
			}
			__antithesis_instrumentation__.Notify(33026)
			row := []string{wrapCode(name), typ, wrapCode(defaultVal), setting.Description()}
			rows = append(rows, row)
		}
		__antithesis_instrumentation__.Notify(33016)

		sliceIter := clisqlexec.NewRowSliceIter(rows, "dddd")
		cols := []string{"Setting", "Type", "Default", "Description"}
		return sqlExecCtx.PrintQueryOutput(os.Stdout, stderr, cols, sliceIter)
	},
}

var genCmd = &cobra.Command{
	Use:   "gen [command]",
	Short: "generate auxiliary files",
	Long:  "Generate manpages, example shell settings, example databases, etc.",
	RunE:  UsageAndErr,
}

var genCmds = []*cobra.Command{
	genManCmd,
	genAutocompleteCmd,
	genExamplesCmd,
	genHAProxyCmd,
	genSettingsListCmd,
	genEncryptionKeyCmd,
}

func init() {
	genManCmd.PersistentFlags().StringVar(&manPath, "path", "man/man1",
		"path where man pages will be outputted")
	genAutocompleteCmd.PersistentFlags().StringVar(&autoCompletePath, "out", "",
		"path to generated autocomplete file")
	genHAProxyCmd.PersistentFlags().StringVar(&haProxyPath, "out", "haproxy.cfg",
		"path to generated haproxy configuration file")
	varFlag(genHAProxyCmd.Flags(), &haProxyLocality, cliflags.Locality)
	genEncryptionKeyCmd.PersistentFlags().IntVarP(&aesSize, "size", "s", 128,
		"AES key size for encryption at rest (one of: 128, 192, 256)")
	genEncryptionKeyCmd.PersistentFlags().BoolVar(&overwriteKey, "overwrite", false,
		"Overwrite key if it exists")
	genSettingsListCmd.PersistentFlags().BoolVar(&includeReservedSettings, "include-reserved", false,
		"include undocumented 'reserved' settings")
	genSettingsListCmd.PersistentFlags().BoolVar(&excludeSystemSettings, "without-system-only", false,
		"do not list settings only applicable to system tenant")

	genCmd.AddCommand(genCmds...)
}
