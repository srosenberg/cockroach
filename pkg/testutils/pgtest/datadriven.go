package pgtest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/datadriven"
	"github.com/jackc/pgproto3/v2"
	"github.com/stretchr/testify/require"
)

func WalkWithRunningServer(t *testing.T, path, addr, user string) {
	__antithesis_instrumentation__.Notify(645679)
	datadriven.Walk(t, path, func(t *testing.T, path string) {
		__antithesis_instrumentation__.Notify(645680)
		RunTest(t, path, addr, user)
	})
}

func WalkWithNewServer(
	t *testing.T, path string, newServer func() (addr, user string, cleanup func()),
) {
	__antithesis_instrumentation__.Notify(645681)
	datadriven.Walk(t, path, func(t *testing.T, path string) {
		__antithesis_instrumentation__.Notify(645682)
		addr, user, cleanup := newServer()
		defer cleanup()
		RunTest(t, path, addr, user)
	})
}

func RunTest(t *testing.T, path, addr, user string) {
	__antithesis_instrumentation__.Notify(645683)
	p, err := NewPGTest(context.Background(), addr, user)

	if err != nil {
		__antithesis_instrumentation__.Notify(645686)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(645687)
	}
	__antithesis_instrumentation__.Notify(645684)
	vars := make(map[string]string)
	datadriven.RunTest(t, path, func(t *testing.T, d *datadriven.TestData) string {
		__antithesis_instrumentation__.Notify(645688)
		switch d.Cmd {
		case "only":
			__antithesis_instrumentation__.Notify(645689)
			if d.HasArg("crdb") && func() bool {
				__antithesis_instrumentation__.Notify(645706)
				return !p.isCockroachDB == true
			}() == true {
				__antithesis_instrumentation__.Notify(645707)
				skip.IgnoreLint(t, "only crdb")
			} else {
				__antithesis_instrumentation__.Notify(645708)
			}
			__antithesis_instrumentation__.Notify(645690)
			if d.HasArg("noncrdb") && func() bool {
				__antithesis_instrumentation__.Notify(645709)
				return p.isCockroachDB == true
			}() == true {
				__antithesis_instrumentation__.Notify(645710)
				skip.IgnoreLint(t, "only non-crdb")
			} else {
				__antithesis_instrumentation__.Notify(645711)
			}
			__antithesis_instrumentation__.Notify(645691)
			return d.Expected

		case "let":
			__antithesis_instrumentation__.Notify(645692)
			require.Len(t, d.CmdArgs, 1, "only one argument permitted for let")
			require.Truef(t, strings.HasPrefix(d.CmdArgs[0].Key, "$"), "let argument must begin with '$'")
			lines := strings.Split(d.Input, "\n")
			require.Len(t, lines, 1, "only one input command permitted for let")
			require.Truef(t, strings.HasPrefix(lines[0], "Query "), "let must use a Query command")
			if err := p.SendOneLine(lines[0]); err != nil {
				__antithesis_instrumentation__.Notify(645712)
				t.Fatalf("%s: send %s: %v", d.Pos, lines[0], err)
			} else {
				__antithesis_instrumentation__.Notify(645713)
			}
			__antithesis_instrumentation__.Notify(645693)
			msgs, err := p.Receive(hasKeepErrMsg(d), &pgproto3.DataRow{}, &pgproto3.ReadyForQuery{})
			if err != nil {
				__antithesis_instrumentation__.Notify(645714)
				t.Fatalf("%s: %+v", d.Pos, err)
			} else {
				__antithesis_instrumentation__.Notify(645715)
			}
			__antithesis_instrumentation__.Notify(645694)
			sawData := false
			for _, msg := range msgs {
				__antithesis_instrumentation__.Notify(645716)
				if dataRow, ok := msg.(*pgproto3.DataRow); ok {
					__antithesis_instrumentation__.Notify(645717)
					require.False(t, sawData, "let Query must return only one row")
					require.Len(t, dataRow.Values, 1, "let Query must return only one column")
					sawData = true
					for _, arg := range d.CmdArgs {
						__antithesis_instrumentation__.Notify(645718)
						vars[arg.Key] = string(dataRow.Values[0])
					}
				} else {
					__antithesis_instrumentation__.Notify(645719)
				}
			}
			__antithesis_instrumentation__.Notify(645695)
			return ""
		case "send":
			__antithesis_instrumentation__.Notify(645696)
			if (d.HasArg("crdb_only") && func() bool {
				__antithesis_instrumentation__.Notify(645720)
				return !p.isCockroachDB == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(645721)
				return (d.HasArg("noncrdb_only") && func() bool {
					__antithesis_instrumentation__.Notify(645722)
					return p.isCockroachDB == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(645723)
				return d.Expected
			} else {
				__antithesis_instrumentation__.Notify(645724)
			}
			__antithesis_instrumentation__.Notify(645697)
			for _, line := range strings.Split(d.Input, "\n") {
				__antithesis_instrumentation__.Notify(645725)
				for k, v := range vars {
					__antithesis_instrumentation__.Notify(645727)
					line = strings.ReplaceAll(line, k, v)
				}
				__antithesis_instrumentation__.Notify(645726)
				if err := p.SendOneLine(line); err != nil {
					__antithesis_instrumentation__.Notify(645728)
					t.Fatalf("%s: send %s: %v", d.Pos, line, err)
				} else {
					__antithesis_instrumentation__.Notify(645729)
				}
			}
			__antithesis_instrumentation__.Notify(645698)
			return ""
		case "receive":
			__antithesis_instrumentation__.Notify(645699)
			if (d.HasArg("crdb_only") && func() bool {
				__antithesis_instrumentation__.Notify(645730)
				return !p.isCockroachDB == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(645731)
				return (d.HasArg("noncrdb_only") && func() bool {
					__antithesis_instrumentation__.Notify(645732)
					return p.isCockroachDB == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(645733)
				return d.Expected
			} else {
				__antithesis_instrumentation__.Notify(645734)
			}
			__antithesis_instrumentation__.Notify(645700)
			until := ParseMessages(d.Input)
			msgs, err := p.Receive(hasKeepErrMsg(d), until...)
			if err != nil {
				__antithesis_instrumentation__.Notify(645735)
				t.Fatalf("%s: %+v", d.Pos, err)
			} else {
				__antithesis_instrumentation__.Notify(645736)
			}
			__antithesis_instrumentation__.Notify(645701)
			return MsgsToJSONWithIgnore(msgs, d)
		case "until":
			__antithesis_instrumentation__.Notify(645702)
			if (d.HasArg("crdb_only") && func() bool {
				__antithesis_instrumentation__.Notify(645737)
				return !p.isCockroachDB == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(645738)
				return (d.HasArg("noncrdb_only") && func() bool {
					__antithesis_instrumentation__.Notify(645739)
					return p.isCockroachDB == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(645740)
				return d.Expected
			} else {
				__antithesis_instrumentation__.Notify(645741)
			}
			__antithesis_instrumentation__.Notify(645703)
			until := ParseMessages(d.Input)
			msgs, err := p.Until(hasKeepErrMsg(d), until...)
			if err != nil {
				__antithesis_instrumentation__.Notify(645742)
				t.Fatalf("%s: %+v", d.Pos, err)
			} else {
				__antithesis_instrumentation__.Notify(645743)
			}
			__antithesis_instrumentation__.Notify(645704)
			return MsgsToJSONWithIgnore(msgs, d)
		default:
			__antithesis_instrumentation__.Notify(645705)
			t.Fatalf("unknown command %s", d.Cmd)
			return ""
		}
	})
	__antithesis_instrumentation__.Notify(645685)
	if err := p.Close(); err != nil {
		__antithesis_instrumentation__.Notify(645744)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(645745)
	}
}

func ParseMessages(s string) []pgproto3.BackendMessage {
	__antithesis_instrumentation__.Notify(645746)
	var msgs []pgproto3.BackendMessage
	for _, typ := range strings.Split(s, "\n") {
		__antithesis_instrumentation__.Notify(645748)
		msgs = append(msgs, toMessage(typ).(pgproto3.BackendMessage))
	}
	__antithesis_instrumentation__.Notify(645747)
	return msgs
}

func hasKeepErrMsg(d *datadriven.TestData) bool {
	__antithesis_instrumentation__.Notify(645749)
	for _, arg := range d.CmdArgs {
		__antithesis_instrumentation__.Notify(645751)
		if arg.Key == "keepErrMessage" {
			__antithesis_instrumentation__.Notify(645752)
			return true
		} else {
			__antithesis_instrumentation__.Notify(645753)
		}
	}
	__antithesis_instrumentation__.Notify(645750)
	return false
}

func MsgsToJSONWithIgnore(msgs []pgproto3.BackendMessage, args *datadriven.TestData) string {
	__antithesis_instrumentation__.Notify(645754)
	ignore := map[string]bool{}
	errs := map[string]string{}
	for _, arg := range args.CmdArgs {
		__antithesis_instrumentation__.Notify(645757)
		switch arg.Key {
		case "keepErrMessage":
			__antithesis_instrumentation__.Notify(645758)
		case "crdb_only":
			__antithesis_instrumentation__.Notify(645759)
		case "noncrdb_only":
			__antithesis_instrumentation__.Notify(645760)
		case "ignore_table_oids":
			__antithesis_instrumentation__.Notify(645761)
			for _, msg := range msgs {
				__antithesis_instrumentation__.Notify(645768)
				if m, ok := msg.(*pgproto3.RowDescription); ok {
					__antithesis_instrumentation__.Notify(645769)
					for i := range m.Fields {
						__antithesis_instrumentation__.Notify(645770)
						m.Fields[i].TableOID = 0
					}
				} else {
					__antithesis_instrumentation__.Notify(645771)
				}
			}
		case "ignore_type_oids":
			__antithesis_instrumentation__.Notify(645762)
			for _, msg := range msgs {
				__antithesis_instrumentation__.Notify(645772)
				if m, ok := msg.(*pgproto3.RowDescription); ok {
					__antithesis_instrumentation__.Notify(645773)
					for i := range m.Fields {
						__antithesis_instrumentation__.Notify(645774)
						m.Fields[i].DataTypeOID = 0
					}
				} else {
					__antithesis_instrumentation__.Notify(645775)
				}
			}
		case "ignore_data_type_sizes":
			__antithesis_instrumentation__.Notify(645763)
			for _, msg := range msgs {
				__antithesis_instrumentation__.Notify(645776)
				if m, ok := msg.(*pgproto3.RowDescription); ok {
					__antithesis_instrumentation__.Notify(645777)
					for i := range m.Fields {
						__antithesis_instrumentation__.Notify(645778)
						m.Fields[i].DataTypeSize = 0
					}
				} else {
					__antithesis_instrumentation__.Notify(645779)
				}
			}
		case "ignore_constraint_name":
			__antithesis_instrumentation__.Notify(645764)
			for _, msg := range msgs {
				__antithesis_instrumentation__.Notify(645780)
				if m, ok := msg.(*pgproto3.ErrorResponse); ok {
					__antithesis_instrumentation__.Notify(645781)
					m.ConstraintName = ""
				} else {
					__antithesis_instrumentation__.Notify(645782)
				}
			}
		case "ignore":
			__antithesis_instrumentation__.Notify(645765)
			for _, typ := range arg.Vals {
				__antithesis_instrumentation__.Notify(645783)
				ignore[fmt.Sprintf("*pgproto3.%s", typ)] = true
			}
		case "mapError":
			__antithesis_instrumentation__.Notify(645766)
			errs[arg.Vals[0]] = arg.Vals[1]
		default:
			__antithesis_instrumentation__.Notify(645767)
			panic(fmt.Errorf("unknown argument: %v", arg))
		}
	}
	__antithesis_instrumentation__.Notify(645755)
	var sb strings.Builder
	enc := json.NewEncoder(&sb)
	for _, msg := range msgs {
		__antithesis_instrumentation__.Notify(645784)
		if ignore[fmt.Sprintf("%T", msg)] {
			__antithesis_instrumentation__.Notify(645786)
			continue
		} else {
			__antithesis_instrumentation__.Notify(645787)
		}
		__antithesis_instrumentation__.Notify(645785)
		if errmsg, ok := msg.(*pgproto3.ErrorResponse); ok {
			__antithesis_instrumentation__.Notify(645788)
			code := errmsg.Code
			if v, ok := errs[code]; ok {
				__antithesis_instrumentation__.Notify(645790)
				code = v
			} else {
				__antithesis_instrumentation__.Notify(645791)
			}
			__antithesis_instrumentation__.Notify(645789)
			if err := enc.Encode(struct {
				Type           string
				Code           string
				Message        string `json:",omitempty"`
				ConstraintName string `json:",omitempty"`
			}{
				Type:           "ErrorResponse",
				Code:           code,
				Message:        errmsg.Message,
				ConstraintName: errmsg.ConstraintName,
			}); err != nil {
				__antithesis_instrumentation__.Notify(645792)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(645793)
			}
		} else {
			__antithesis_instrumentation__.Notify(645794)
			if err := enc.Encode(msg); err != nil {
				__antithesis_instrumentation__.Notify(645795)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(645796)
			}
		}
	}
	__antithesis_instrumentation__.Notify(645756)
	return sb.String()
}

func toMessage(typ string) interface{} {
	__antithesis_instrumentation__.Notify(645797)
	switch typ {
	case "Bind":
		__antithesis_instrumentation__.Notify(645798)
		return &pgproto3.Bind{}
	case "Close":
		__antithesis_instrumentation__.Notify(645799)
		return &pgproto3.Close{}
	case "CommandComplete":
		__antithesis_instrumentation__.Notify(645800)
		return &pgproto3.CommandComplete{}
	case "CopyData":
		__antithesis_instrumentation__.Notify(645801)
		return &pgproto3.CopyData{}
	case "CopyDone":
		__antithesis_instrumentation__.Notify(645802)
		return &pgproto3.CopyDone{}
	case "CopyInResponse":
		__antithesis_instrumentation__.Notify(645803)
		return &pgproto3.CopyInResponse{}
	case "DataRow":
		__antithesis_instrumentation__.Notify(645804)
		return &pgproto3.DataRow{}
	case "Describe":
		__antithesis_instrumentation__.Notify(645805)
		return &pgproto3.Describe{}
	case "ErrorResponse":
		__antithesis_instrumentation__.Notify(645806)
		return &pgproto3.ErrorResponse{}
	case "Execute":
		__antithesis_instrumentation__.Notify(645807)
		return &pgproto3.Execute{}
	case "Parse":
		__antithesis_instrumentation__.Notify(645808)
		return &pgproto3.Parse{}
	case "PortalSuspended":
		__antithesis_instrumentation__.Notify(645809)
		return &pgproto3.PortalSuspended{}
	case "Query":
		__antithesis_instrumentation__.Notify(645810)
		return &pgproto3.Query{}
	case "ReadyForQuery":
		__antithesis_instrumentation__.Notify(645811)
		return &pgproto3.ReadyForQuery{}
	case "Sync":
		__antithesis_instrumentation__.Notify(645812)
		return &pgproto3.Sync{}
	case "ParameterStatus":
		__antithesis_instrumentation__.Notify(645813)
		return &pgproto3.ParameterStatus{}
	case "BindComplete":
		__antithesis_instrumentation__.Notify(645814)
		return &pgproto3.BindComplete{}
	case "ParseComplete":
		__antithesis_instrumentation__.Notify(645815)
		return &pgproto3.ParseComplete{}
	case "RowDescription":
		__antithesis_instrumentation__.Notify(645816)
		return &pgproto3.RowDescription{}
	default:
		__antithesis_instrumentation__.Notify(645817)
		panic(fmt.Errorf("unknown type %q", typ))
	}
}
