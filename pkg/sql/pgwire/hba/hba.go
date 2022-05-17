// Package hba implements an hba.conf parser.
package hba

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
	"github.com/olekukonko/tablewriter"
)

type Conf struct {
	Entries []Entry
}

type Entry struct {
	ConnType ConnType

	Database []String

	User []String

	Address interface{}
	Method  String

	MethodFn     interface{}
	Options      [][2]string
	OptionQuotes []bool

	Input string

	Generated bool
}

type ConnType int

const (
	ConnLocal ConnType = 1 << iota

	ConnHostNoSSL

	ConnHostSSL

	ConnHostAny = ConnHostNoSSL | ConnHostSSL

	ConnAny = ConnHostAny | ConnLocal
)

func (t ConnType) String() string {
	__antithesis_instrumentation__.Notify(559843)
	switch t {
	case ConnLocal:
		__antithesis_instrumentation__.Notify(559844)
		return "local"
	case ConnHostNoSSL:
		__antithesis_instrumentation__.Notify(559845)
		return "hostnossl"
	case ConnHostSSL:
		__antithesis_instrumentation__.Notify(559846)
		return "hostssl"
	case ConnHostAny:
		__antithesis_instrumentation__.Notify(559847)
		return "host"
	default:
		__antithesis_instrumentation__.Notify(559848)
		panic(errors.Newf("unimplemented conn type: %v", int(t)))
	}
}

func (c Conf) String() string {
	__antithesis_instrumentation__.Notify(559849)
	if len(c.Entries) == 0 {
		__antithesis_instrumentation__.Notify(559853)
		return "# (empty configuration)\n"
	} else {
		__antithesis_instrumentation__.Notify(559854)
	}
	__antithesis_instrumentation__.Notify(559850)
	var sb strings.Builder
	sb.WriteString("# Original configuration:\n")
	for _, e := range c.Entries {
		__antithesis_instrumentation__.Notify(559855)
		if e.Generated {
			__antithesis_instrumentation__.Notify(559857)
			continue
		} else {
			__antithesis_instrumentation__.Notify(559858)
		}
		__antithesis_instrumentation__.Notify(559856)
		fmt.Fprintf(&sb, "# %s\n", e.Input)
	}
	__antithesis_instrumentation__.Notify(559851)
	sb.WriteString("#\n# Interpreted configuration:\n")

	table := tablewriter.NewWriter(&sb)
	table.SetAutoWrapText(false)
	table.SetReflowDuringAutoWrap(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetNoWhiteSpace(true)
	table.SetTrimWhiteSpaceAtEOL(true)
	table.SetTablePadding(" ")

	row := []string{"# TYPE", "DATABASE", "USER", "ADDRESS", "METHOD", "OPTIONS"}
	table.Append(row)
	for _, e := range c.Entries {
		__antithesis_instrumentation__.Notify(559859)
		row[0] = e.ConnType.String()
		row[1] = e.DatabaseString()
		row[2] = e.UserString()
		row[3] = e.AddressString()
		row[4] = e.Method.String()
		row[5] = e.OptionsString()
		table.Append(row)
	}
	__antithesis_instrumentation__.Notify(559852)
	table.Render()
	return sb.String()
}

type AnyAddr struct{}

func (AnyAddr) String() string { __antithesis_instrumentation__.Notify(559860); return "all" }

func (h Entry) GetOption(name string) string {
	__antithesis_instrumentation__.Notify(559861)
	var val string
	for _, opt := range h.Options {
		__antithesis_instrumentation__.Notify(559863)
		if opt[0] == name {
			__antithesis_instrumentation__.Notify(559864)

			if val != "" {
				__antithesis_instrumentation__.Notify(559866)
				return ""
			} else {
				__antithesis_instrumentation__.Notify(559867)
			}
			__antithesis_instrumentation__.Notify(559865)
			val = opt[1]
		} else {
			__antithesis_instrumentation__.Notify(559868)
		}
	}
	__antithesis_instrumentation__.Notify(559862)
	return val
}

func (h Entry) Equivalent(other Entry) bool {
	__antithesis_instrumentation__.Notify(559869)
	h.Input = ""
	other.Input = ""
	return reflect.DeepEqual(h, other)
}

func (h Entry) GetOptions(name string) []string {
	__antithesis_instrumentation__.Notify(559870)
	var val []string
	for _, opt := range h.Options {
		__antithesis_instrumentation__.Notify(559872)
		if opt[0] == name {
			__antithesis_instrumentation__.Notify(559873)
			val = append(val, opt[1])
		} else {
			__antithesis_instrumentation__.Notify(559874)
		}
	}
	__antithesis_instrumentation__.Notify(559871)
	return val
}

func (h Entry) ConnTypeMatches(clientConn ConnType) bool {
	__antithesis_instrumentation__.Notify(559875)
	switch clientConn {
	case ConnLocal:
		__antithesis_instrumentation__.Notify(559876)
		return h.ConnType == ConnLocal
	case ConnHostSSL:
		__antithesis_instrumentation__.Notify(559877)

		return h.ConnType&ConnHostSSL != 0
	case ConnHostNoSSL:
		__antithesis_instrumentation__.Notify(559878)

		return h.ConnType&ConnHostNoSSL != 0
	default:
		__antithesis_instrumentation__.Notify(559879)
		panic("unimplemented")
	}
}

func (h Entry) ConnMatches(clientConn ConnType, ip net.IP) (bool, error) {
	__antithesis_instrumentation__.Notify(559880)
	if !h.ConnTypeMatches(clientConn) {
		__antithesis_instrumentation__.Notify(559883)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(559884)
	}
	__antithesis_instrumentation__.Notify(559881)
	if clientConn != ConnLocal {
		__antithesis_instrumentation__.Notify(559885)
		return h.AddressMatches(ip)
	} else {
		__antithesis_instrumentation__.Notify(559886)
	}
	__antithesis_instrumentation__.Notify(559882)
	return true, nil
}

func (h Entry) UserMatches(userName security.SQLUsername) bool {
	__antithesis_instrumentation__.Notify(559887)
	if h.User == nil {
		__antithesis_instrumentation__.Notify(559890)
		return true
	} else {
		__antithesis_instrumentation__.Notify(559891)
	}
	__antithesis_instrumentation__.Notify(559888)
	for _, u := range h.User {
		__antithesis_instrumentation__.Notify(559892)
		if u.Value == userName.Normalized() {
			__antithesis_instrumentation__.Notify(559893)
			return true
		} else {
			__antithesis_instrumentation__.Notify(559894)
		}
	}
	__antithesis_instrumentation__.Notify(559889)
	return false
}

func (h Entry) AddressMatches(addr net.IP) (bool, error) {
	__antithesis_instrumentation__.Notify(559895)
	switch a := h.Address.(type) {
	case AnyAddr:
		__antithesis_instrumentation__.Notify(559896)
		return true, nil
	case *net.IPNet:
		__antithesis_instrumentation__.Notify(559897)
		return a.Contains(addr), nil
	default:
		__antithesis_instrumentation__.Notify(559898)

		return false, errors.Newf("unknown address type: %T", addr)
	}
}

func (h Entry) DatabaseString() string {
	__antithesis_instrumentation__.Notify(559899)
	if h.Database == nil {
		__antithesis_instrumentation__.Notify(559902)
		return "all"
	} else {
		__antithesis_instrumentation__.Notify(559903)
	}
	__antithesis_instrumentation__.Notify(559900)
	var sb strings.Builder
	comma := ""
	for _, s := range h.Database {
		__antithesis_instrumentation__.Notify(559904)
		sb.WriteString(comma)
		sb.WriteString(s.String())
		comma = ","
	}
	__antithesis_instrumentation__.Notify(559901)
	return sb.String()
}

func (h Entry) UserString() string {
	__antithesis_instrumentation__.Notify(559905)
	if h.User == nil {
		__antithesis_instrumentation__.Notify(559908)
		return "all"
	} else {
		__antithesis_instrumentation__.Notify(559909)
	}
	__antithesis_instrumentation__.Notify(559906)
	var sb strings.Builder
	comma := ""
	for _, s := range h.User {
		__antithesis_instrumentation__.Notify(559910)
		sb.WriteString(comma)
		sb.WriteString(s.String())
		comma = ","
	}
	__antithesis_instrumentation__.Notify(559907)
	return sb.String()
}

func (h Entry) AddressString() string {
	__antithesis_instrumentation__.Notify(559911)
	if h.Address == nil {
		__antithesis_instrumentation__.Notify(559913)

		return ""
	} else {
		__antithesis_instrumentation__.Notify(559914)
	}
	__antithesis_instrumentation__.Notify(559912)
	return fmt.Sprintf("%s", h.Address)
}

func (h Entry) OptionsString() string {
	__antithesis_instrumentation__.Notify(559915)
	var sb strings.Builder
	sp := ""
	for i, opt := range h.Options {
		__antithesis_instrumentation__.Notify(559917)
		sb.WriteString(sp)
		sb.WriteString(String{Value: opt[0] + "=" + opt[1], Quoted: h.OptionQuotes[i]}.String())
		sp = " "
	}
	__antithesis_instrumentation__.Notify(559916)
	return sb.String()
}

func (h Entry) String() string {
	__antithesis_instrumentation__.Notify(559918)
	return Conf{Entries: []Entry{h}}.String()
}

type String struct {
	Value  string
	Quoted bool
}

func (s String) String() string {
	__antithesis_instrumentation__.Notify(559919)
	if s.Quoted {
		__antithesis_instrumentation__.Notify(559921)
		return `"` + s.Value + `"`
	} else {
		__antithesis_instrumentation__.Notify(559922)
	}
	__antithesis_instrumentation__.Notify(559920)
	return s.Value
}

func (s String) Empty() bool { __antithesis_instrumentation__.Notify(559923); return s.IsKeyword("") }

func (s String) IsKeyword(v string) bool {
	__antithesis_instrumentation__.Notify(559924)
	return !s.Quoted && func() bool {
		__antithesis_instrumentation__.Notify(559925)
		return s.Value == v == true
	}() == true
}

func ParseAndNormalize(val string) (*Conf, error) {
	__antithesis_instrumentation__.Notify(559926)
	conf, err := Parse(val)
	if err != nil {
		__antithesis_instrumentation__.Notify(559929)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(559930)
	}
	__antithesis_instrumentation__.Notify(559927)

	entries := conf.Entries[:0]
	entriesCopied := false
outer:
	for i := range conf.Entries {
		__antithesis_instrumentation__.Notify(559931)
		entry := conf.Entries[i]

		entry.Database = nil

		if addr, ok := entry.Address.(String); ok && func() bool {
			__antithesis_instrumentation__.Notify(559935)
			return addr.IsKeyword("all") == true
		}() == true {
			__antithesis_instrumentation__.Notify(559936)
			entry.Address = AnyAddr{}
		} else {
			__antithesis_instrumentation__.Notify(559937)
		}
		__antithesis_instrumentation__.Notify(559932)

		for _, iu := range entry.User {
			__antithesis_instrumentation__.Notify(559938)
			if iu.IsKeyword("all") {
				__antithesis_instrumentation__.Notify(559939)
				entry.User = nil
				entries = append(entries, entry)
				continue outer
			} else {
				__antithesis_instrumentation__.Notify(559940)
			}
		}
		__antithesis_instrumentation__.Notify(559933)

		if len(entry.User) != 1 && func() bool {
			__antithesis_instrumentation__.Notify(559941)
			return !entriesCopied == true
		}() == true {
			__antithesis_instrumentation__.Notify(559942)
			entries = append([]Entry(nil), conf.Entries[:len(entries)]...)
			entriesCopied = true
		} else {
			__antithesis_instrumentation__.Notify(559943)
		}
		__antithesis_instrumentation__.Notify(559934)

		allUsers := entry.User
		for userIdx, iu := range allUsers {
			__antithesis_instrumentation__.Notify(559944)
			entry.User = allUsers[userIdx : userIdx+1]
			entry.User[0].Value = tree.Name(iu.Value).Normalize()
			if userIdx > 0 {
				__antithesis_instrumentation__.Notify(559946)
				entry.Generated = true
			} else {
				__antithesis_instrumentation__.Notify(559947)
			}
			__antithesis_instrumentation__.Notify(559945)
			entries = append(entries, entry)
		}
	}
	__antithesis_instrumentation__.Notify(559928)
	conf.Entries = entries
	return conf, nil
}
