// uptodate efficiently computes whether an output file is up-to-date with
// regard to its input files.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/MichaelTJones/walk"
	"github.com/cockroachdb/errors/oserror"
	"github.com/spf13/pflag"
)

var debug = pflag.BoolP("debug", "d", false, "debug mode")

func die(err error) {
	__antithesis_instrumentation__.Notify(53212)
	fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	os.Exit(2)
}

func main() {
	__antithesis_instrumentation__.Notify(53213)
	pflag.Usage = func() {
		__antithesis_instrumentation__.Notify(53219)
		fmt.Fprintf(os.Stderr, "usage: %s [-d] OUTPUT INPUT...\n", os.Args[0])
	}
	__antithesis_instrumentation__.Notify(53214)
	pflag.Parse()
	if pflag.NArg() < 2 {
		__antithesis_instrumentation__.Notify(53220)
		pflag.Usage()
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(53221)
	}
	__antithesis_instrumentation__.Notify(53215)
	if !*debug {
		__antithesis_instrumentation__.Notify(53222)
		log.SetOutput(ioutil.Discard)
	} else {
		__antithesis_instrumentation__.Notify(53223)
	}
	__antithesis_instrumentation__.Notify(53216)
	output, inputs := pflag.Arg(0), pflag.Args()[1:]

	fi, err := os.Stat(output)
	if oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(53224)
		log.Printf("output %q is missing", output)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(53225)
		if err != nil {
			__antithesis_instrumentation__.Notify(53226)
			die(err)
		} else {
			__antithesis_instrumentation__.Notify(53227)
		}
	}
	__antithesis_instrumentation__.Notify(53217)
	outputModTime := fi.ModTime()

	for _, input := range inputs {
		__antithesis_instrumentation__.Notify(53228)
		err = walk.Walk(input, func(path string, fi os.FileInfo, err error) error {
			__antithesis_instrumentation__.Notify(53230)
			if err != nil {
				__antithesis_instrumentation__.Notify(53234)
				return err
			} else {
				__antithesis_instrumentation__.Notify(53235)
			}
			__antithesis_instrumentation__.Notify(53231)
			if fi.IsDir() {
				__antithesis_instrumentation__.Notify(53236)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(53237)
			}
			__antithesis_instrumentation__.Notify(53232)
			if !fi.ModTime().Before(outputModTime) {
				__antithesis_instrumentation__.Notify(53238)
				log.Printf("input %q (mtime %s) not older than output %q (mtime %s)",
					path, fi.ModTime(), output, outputModTime)
				os.Exit(1)
			} else {
				__antithesis_instrumentation__.Notify(53239)
			}
			__antithesis_instrumentation__.Notify(53233)
			return nil
		})
		__antithesis_instrumentation__.Notify(53229)
		if err != nil {
			__antithesis_instrumentation__.Notify(53240)
			die(err)
		} else {
			__antithesis_instrumentation__.Notify(53241)
		}
	}
	__antithesis_instrumentation__.Notify(53218)
	log.Printf("all inputs older than output %q (mtime %s)\n", output, outputModTime)
	os.Exit(0)
}
