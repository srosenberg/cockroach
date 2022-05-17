package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var debugSendKVBatchContext = struct {
	traceFormat string

	traceFile string

	keepCollectedSpans bool
}{}

func setDebugSendKVBatchContextDefaults() {
	__antithesis_instrumentation__.Notify(31649)
	debugSendKVBatchContext.traceFormat = "off"
	debugSendKVBatchContext.traceFile = ""
	debugSendKVBatchContext.keepCollectedSpans = false
}

var debugSendKVBatchCmd = &cobra.Command{
	Use:   "send-kv-batch <jsonfile>",
	Short: "sends a KV BatchRequest via the connected node",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSendKVBatch,
	Long: `
Submits a JSON-encoded roachpb.BatchRequest (file or stdin) to a node which
proxies it into the KV API, and outputs the JSON-encoded roachpb.BatchResponse.
Requires the admin role. The request is logged to the system event log.

This command can modify internal system state. Incorrect use can cause severe
irreversible damage including permanent data loss.

For more information on requests, see roachpb/api.proto. Unknown or invalid
fields will error. Binary fields ([]byte) are base64-encoded. Requests spanning
multiple ranges are wrapped in a transaction.

The following BatchRequest example sets the key "foo" to the value "bar", and
reads the value back (as a base64-encoded roachpb.Value.RawBytes with checksum):

{"requests": [
	{"put": {
		"header": {"key": "Zm9v"},
		"value": {"raw_bytes": "DMEB5ANiYXI="}
	}},
	{"get": {
		"header": {"key": "Zm9v"}
	}}
]}

This would yield the following response:

{"responses": [
	{"put": {
		"header": {}
	}},
	{"get": {
		"header": {"numKeys": "1", "numBytes": "8"},
		"value": {"rawBytes": "DMEB5ANiYXI=", "timestamp": {"wallTime": "1636898055143103168"}}
	}}
]}

To generate JSON requests with Go (see also debug_send_kv_batch_test.go):

func TestSendKVBatchExample(t *testing.T) {
	var ba roachpb.BatchRequest
	ba.Add(roachpb.NewPut(roachpb.Key("foo"), roachpb.MakeValueFromString("bar")))
	ba.Add(roachpb.NewGet(roachpb.Key("foo"), false /* forUpdate */))

	jsonpb := protoutil.JSONPb{}
	jsonProto, err := jsonpb.Marshal(&ba)
	require.NoError(t, err)

	fmt.Println(string(jsonProto))
}
	`,
}

func runSendKVBatch(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(31650)
	enableTracing := false
	fmtJaeger := false
	switch debugSendKVBatchContext.traceFormat {
	case "on", "text":
		__antithesis_instrumentation__.Notify(31662)

		enableTracing = true
		fmtJaeger = false
	case "jaeger":
		__antithesis_instrumentation__.Notify(31663)
		enableTracing = true
		fmtJaeger = true
	case "off":
		__antithesis_instrumentation__.Notify(31664)
	default:
		__antithesis_instrumentation__.Notify(31665)
		return errors.New("unknown --trace value")
	}
	__antithesis_instrumentation__.Notify(31651)
	var traceFile *os.File
	if enableTracing {
		__antithesis_instrumentation__.Notify(31666)
		fileName := debugSendKVBatchContext.traceFile
		if fileName == "" {
			__antithesis_instrumentation__.Notify(31667)

			traceFile = stderr
		} else {
			__antithesis_instrumentation__.Notify(31668)
			var err error
			traceFile, err = os.OpenFile(
				fileName,
				os.O_TRUNC|os.O_CREATE|os.O_WRONLY,

				0600)
			if err != nil {
				__antithesis_instrumentation__.Notify(31670)
				return err
			} else {
				__antithesis_instrumentation__.Notify(31671)
			}
			__antithesis_instrumentation__.Notify(31669)
			defer func() {
				__antithesis_instrumentation__.Notify(31672)
				if err := traceFile.Close(); err != nil {
					__antithesis_instrumentation__.Notify(31673)
					fmt.Fprintf(stderr, "warning: error while closing trace output: %v\n", err)
				} else {
					__antithesis_instrumentation__.Notify(31674)
				}
			}()
		}
	} else {
		__antithesis_instrumentation__.Notify(31675)
	}
	__antithesis_instrumentation__.Notify(31652)

	jsonpb := protoutil.JSONPb{Indent: "  "}

	var baJSON []byte
	var err error
	if len(args) > 0 {
		__antithesis_instrumentation__.Notify(31676)
		baJSON, err = ioutil.ReadFile(args[0])
	} else {
		__antithesis_instrumentation__.Notify(31677)
		baJSON, err = ioutil.ReadAll(os.Stdin)
	}
	__antithesis_instrumentation__.Notify(31653)
	if err != nil {
		__antithesis_instrumentation__.Notify(31678)
		return errors.Wrapf(err, "failed to read input")
	} else {
		__antithesis_instrumentation__.Notify(31679)
	}
	__antithesis_instrumentation__.Notify(31654)

	var ba roachpb.BatchRequest
	if err := jsonpb.Unmarshal(baJSON, &ba); err != nil {
		__antithesis_instrumentation__.Notify(31680)
		return errors.Wrap(err, "invalid JSON")
	} else {
		__antithesis_instrumentation__.Notify(31681)
	}
	__antithesis_instrumentation__.Notify(31655)

	if len(ba.Requests) == 0 {
		__antithesis_instrumentation__.Notify(31682)
		return errors.New("BatchRequest contains no requests")
	} else {
		__antithesis_instrumentation__.Notify(31683)
	}
	__antithesis_instrumentation__.Notify(31656)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conn, _, finish, err := getClientGRPCConn(ctx, serverCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(31684)
		return errors.Wrap(err, "failed to connect to the node")
	} else {
		__antithesis_instrumentation__.Notify(31685)
	}
	__antithesis_instrumentation__.Notify(31657)
	defer finish()
	admin := serverpb.NewAdminClient(conn)

	br, rec, err := sendKVBatchRequestWithTracingOption(ctx, enableTracing, admin, &ba)
	if err != nil {
		__antithesis_instrumentation__.Notify(31686)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31687)
	}
	__antithesis_instrumentation__.Notify(31658)

	if !debugSendKVBatchContext.keepCollectedSpans {
		__antithesis_instrumentation__.Notify(31688)

		br.CollectedSpans = nil
	} else {
		__antithesis_instrumentation__.Notify(31689)
	}
	__antithesis_instrumentation__.Notify(31659)

	brJSON, err := jsonpb.Marshal(br)
	if err != nil {
		__antithesis_instrumentation__.Notify(31690)
		return errors.Wrap(err, "failed to format BatchResponse as JSON")
	} else {
		__antithesis_instrumentation__.Notify(31691)
	}
	__antithesis_instrumentation__.Notify(31660)
	fmt.Println(string(brJSON))

	if enableTracing {
		__antithesis_instrumentation__.Notify(31692)
		out := bufio.NewWriter(traceFile)
		if fmtJaeger {
			__antithesis_instrumentation__.Notify(31694)

			j, err := rec.ToJaegerJSON(ba.Summary(), "", "")
			if err != nil {
				__antithesis_instrumentation__.Notify(31696)
				return err
			} else {
				__antithesis_instrumentation__.Notify(31697)
			}
			__antithesis_instrumentation__.Notify(31695)
			if _, err = fmt.Fprintln(out, j); err != nil {
				__antithesis_instrumentation__.Notify(31698)
				return err
			} else {
				__antithesis_instrumentation__.Notify(31699)
			}
		} else {
			__antithesis_instrumentation__.Notify(31700)
			if _, err = fmt.Fprintln(out, rec); err != nil {
				__antithesis_instrumentation__.Notify(31701)
				return err
			} else {
				__antithesis_instrumentation__.Notify(31702)
			}
		}
		__antithesis_instrumentation__.Notify(31693)
		if err := out.Flush(); err != nil {
			__antithesis_instrumentation__.Notify(31703)
			return err
		} else {
			__antithesis_instrumentation__.Notify(31704)
		}
	} else {
		__antithesis_instrumentation__.Notify(31705)
	}
	__antithesis_instrumentation__.Notify(31661)

	return nil
}

func sendKVBatchRequestWithTracingOption(
	ctx context.Context, verboseTrace bool, admin serverpb.AdminClient, ba *roachpb.BatchRequest,
) (br *roachpb.BatchResponse, rec tracing.Recording, err error) {
	__antithesis_instrumentation__.Notify(31706)
	var sp *tracing.Span
	if verboseTrace {
		__antithesis_instrumentation__.Notify(31709)

		_, sp = tracing.NewTracer().StartSpanCtx(ctx, "debug-send-kv-batch",
			tracing.WithRecording(tracing.RecordingVerbose))
		defer sp.Finish()

		ba.TraceInfo = sp.Meta().ToProto()
	} else {
		__antithesis_instrumentation__.Notify(31710)
	}
	__antithesis_instrumentation__.Notify(31707)

	br, err = admin.SendKVBatch(ctx, ba)

	if sp != nil {
		__antithesis_instrumentation__.Notify(31711)

		sp.ImportRemoteSpans(br.CollectedSpans)

		rec = sp.GetRecording(tracing.RecordingVerbose)
	} else {
		__antithesis_instrumentation__.Notify(31712)
	}
	__antithesis_instrumentation__.Notify(31708)

	return br, rec, errors.Wrap(err, "request failed")
}
