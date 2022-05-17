package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"encoding/base64"
	gohex "encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/cockroachdb/errors"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func runDebugDecodeProto(_ *cobra.Command, _ []string) error {
	__antithesis_instrumentation__.Notify(31800)
	if isatty.IsTerminal(os.Stdin.Fd()) {
		__antithesis_instrumentation__.Notify(31802)
		fmt.Fprintln(stderr,
			`# Reading proto-encoded pieces of data from stdin.
# Press Ctrl+C or Ctrl+D to terminate.`,
		)
	} else {
		__antithesis_instrumentation__.Notify(31803)
	}
	__antithesis_instrumentation__.Notify(31801)
	return streamMap(os.Stdout, os.Stdin,
		func(s string) (bool, string, error) {
			__antithesis_instrumentation__.Notify(31804)
			return tryDecodeValue(s, debugDecodeProtoName, debugDecodeProtoEmitDefaults)
		})
}

func streamMap(out io.Writer, in io.Reader, fn func(string) (bool, string, error)) error {
	__antithesis_instrumentation__.Notify(31805)
	sc := bufio.NewScanner(in)
	sc.Buffer(nil, 128<<20)
	for sc.Scan() {
		__antithesis_instrumentation__.Notify(31807)
		for _, field := range strings.Fields(sc.Text()) {
			__antithesis_instrumentation__.Notify(31809)
			ok, value, err := fn(field)
			if err != nil {
				__antithesis_instrumentation__.Notify(31812)
				fmt.Fprintf(out, "warning:  %v", err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(31813)
			}
			__antithesis_instrumentation__.Notify(31810)
			if !ok {
				__antithesis_instrumentation__.Notify(31814)
				fmt.Fprintf(out, "%s\t", field)

				continue
			} else {
				__antithesis_instrumentation__.Notify(31815)
			}
			__antithesis_instrumentation__.Notify(31811)
			fmt.Fprintf(out, "%s\t", value)
		}
		__antithesis_instrumentation__.Notify(31808)
		fmt.Fprintln(out, "")
	}
	__antithesis_instrumentation__.Notify(31806)
	return sc.Err()
}

func tryDecodeValue(s, protoName string, emitDefaults bool) (ok bool, val string, err error) {
	__antithesis_instrumentation__.Notify(31816)
	bytes, err := gohex.DecodeString(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(31820)
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			__antithesis_instrumentation__.Notify(31822)
			return false, "", nil
		} else {
			__antithesis_instrumentation__.Notify(31823)
		}
		__antithesis_instrumentation__.Notify(31821)
		bytes = b
	} else {
		__antithesis_instrumentation__.Notify(31824)
	}
	__antithesis_instrumentation__.Notify(31817)
	msg, err := protoreflect.DecodeMessage(protoName, bytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(31825)
		return false, "", nil
	} else {
		__antithesis_instrumentation__.Notify(31826)
	}
	__antithesis_instrumentation__.Notify(31818)
	j, err := protoreflect.MessageToJSON(msg, protoreflect.FmtFlags{EmitDefaults: emitDefaults})
	if err != nil {
		__antithesis_instrumentation__.Notify(31827)

		return false, "", errors.Wrapf(err, "while JSON-encoding %#v", msg)
	} else {
		__antithesis_instrumentation__.Notify(31828)
	}
	__antithesis_instrumentation__.Notify(31819)
	return true, j.String(), nil
}
