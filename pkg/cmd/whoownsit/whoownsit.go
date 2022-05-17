// whoownsit looks for OWNERS in the directory and parenting directories
// until it finds an owner for a given file.
//
// Usage: ./whoownsit [<file_or_dir> ...]
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/internal/codeowners"
	"github.com/cockroachdb/cockroach/pkg/internal/reporoot"
)

var walk = flag.Bool("walk", false, "recursively print ownership")
var dirsOnly = flag.Bool("dirs-only", false, "print ownership only for directories")

func main() {
	__antithesis_instrumentation__.Notify(53331)
	flag.Parse()

	codeOwners, err := codeowners.DefaultLoadCodeOwners()
	if err != nil {
		__antithesis_instrumentation__.Notify(53333)
		log.Fatalf("failing to load code owners: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(53334)
	}
	__antithesis_instrumentation__.Notify(53332)

	for _, path := range flag.Args() {
		__antithesis_instrumentation__.Notify(53335)
		if filepath.IsAbs(path) {
			__antithesis_instrumentation__.Notify(53337)
			var err error
			path, err = filepath.Rel(reporoot.Get(), path)
			if err != nil {
				__antithesis_instrumentation__.Notify(53338)
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			} else {
				__antithesis_instrumentation__.Notify(53339)
			}
		} else {
			__antithesis_instrumentation__.Notify(53340)
		}
		__antithesis_instrumentation__.Notify(53336)
		if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			__antithesis_instrumentation__.Notify(53341)
			if err != nil {
				__antithesis_instrumentation__.Notify(53345)
				return err
			} else {
				__antithesis_instrumentation__.Notify(53346)
			}
			__antithesis_instrumentation__.Notify(53342)
			if !*dirsOnly || func() bool {
				__antithesis_instrumentation__.Notify(53347)
				return info.IsDir() == true
			}() == true {
				__antithesis_instrumentation__.Notify(53348)
				matches := codeOwners.Match(path)
				var aliases []string
				for _, match := range matches {
					__antithesis_instrumentation__.Notify(53351)
					aliases = append(aliases, string(match.Name()))
				}
				__antithesis_instrumentation__.Notify(53349)
				if len(aliases) == 0 {
					__antithesis_instrumentation__.Notify(53352)
					aliases = append(aliases, "-")
				} else {
					__antithesis_instrumentation__.Notify(53353)
				}
				__antithesis_instrumentation__.Notify(53350)
				fmt.Println(strings.Join(aliases, ","), " ", path)
			} else {
				__antithesis_instrumentation__.Notify(53354)
			}
			__antithesis_instrumentation__.Notify(53343)
			if !*walk {
				__antithesis_instrumentation__.Notify(53355)
				return filepath.SkipDir
			} else {
				__antithesis_instrumentation__.Notify(53356)
			}
			__antithesis_instrumentation__.Notify(53344)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(53357)
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		} else {
			__antithesis_instrumentation__.Notify(53358)
		}
	}
}
