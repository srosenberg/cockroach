package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net"
	"net/http"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/hba"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/identmap"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

const serverHBAConfSetting = "server.host_based_authentication.configuration"

var connAuthConf = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(560090)
	s := settings.RegisterValidatedStringSetting(
		settings.TenantWritable,
		serverHBAConfSetting,
		"host-based authentication configuration to use during connection authentication",
		"",
		checkHBASyntaxBeforeUpdatingSetting,
	)
	s.SetVisibility(settings.Public)
	return s
}()

func loadLocalHBAConfigUponRemoteSettingChange(
	ctx context.Context, server *Server, st *cluster.Settings,
) {
	__antithesis_instrumentation__.Notify(560091)
	val := connAuthConf.Get(&st.SV)

	hbaConfig := DefaultHBAConfig
	if val != "" {
		__antithesis_instrumentation__.Notify(560093)
		var err error
		hbaConfig, err = ParseAndNormalize(val)
		if err != nil {
			__antithesis_instrumentation__.Notify(560094)

			log.Ops.Warningf(ctx, "invalid %s: %v", serverHBAConfSetting, err)
			hbaConfig = DefaultHBAConfig
		} else {
			__antithesis_instrumentation__.Notify(560095)
		}
	} else {
		__antithesis_instrumentation__.Notify(560096)
	}
	__antithesis_instrumentation__.Notify(560092)

	server.auth.Lock()
	defer server.auth.Unlock()
	server.auth.conf = hbaConfig
}

func checkHBASyntaxBeforeUpdatingSetting(values *settings.Values, s string) error {
	__antithesis_instrumentation__.Notify(560097)
	if s == "" {
		__antithesis_instrumentation__.Notify(560102)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(560103)
	}
	__antithesis_instrumentation__.Notify(560098)

	conf, err := hba.Parse(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(560104)
		return err
	} else {
		__antithesis_instrumentation__.Notify(560105)
	}
	__antithesis_instrumentation__.Notify(560099)

	if len(conf.Entries) == 0 {
		__antithesis_instrumentation__.Notify(560106)

		return errors.WithHint(errors.New("no entries"),
			"To use the default configuration, assign the empty string ('').")
	} else {
		__antithesis_instrumentation__.Notify(560107)
	}
	__antithesis_instrumentation__.Notify(560100)

	for _, entry := range conf.Entries {
		__antithesis_instrumentation__.Notify(560108)
		switch entry.ConnType {
		case hba.ConnHostAny:
			__antithesis_instrumentation__.Notify(560113)
		case hba.ConnLocal:
			__antithesis_instrumentation__.Notify(560114)
		case hba.ConnHostSSL, hba.ConnHostNoSSL:
			__antithesis_instrumentation__.Notify(560115)

		default:
			__antithesis_instrumentation__.Notify(560116)
			return unimplemented.Newf("hba-type-"+entry.ConnType.String(),
				"unsupported connection type: %s", entry.ConnType)
		}
		__antithesis_instrumentation__.Notify(560109)
		for _, db := range entry.Database {
			__antithesis_instrumentation__.Notify(560117)
			if !db.IsKeyword("all") {
				__antithesis_instrumentation__.Notify(560118)
				return errors.WithHint(
					unimplemented.New("hba-per-db", "per-database HBA rules are not supported"),
					"Use the special value 'all' (without quotes) to match all databases.")
			} else {
				__antithesis_instrumentation__.Notify(560119)
			}
		}
		__antithesis_instrumentation__.Notify(560110)

		if entry.ConnType != hba.ConnLocal {
			__antithesis_instrumentation__.Notify(560120)

			addrOk := true
			switch t := entry.Address.(type) {
			case *net.IPNet:
				__antithesis_instrumentation__.Notify(560122)
			case hba.AnyAddr:
				__antithesis_instrumentation__.Notify(560123)
			case hba.String:
				__antithesis_instrumentation__.Notify(560124)
				addrOk = t.IsKeyword("all")
			default:
				__antithesis_instrumentation__.Notify(560125)
				addrOk = false
			}
			__antithesis_instrumentation__.Notify(560121)
			if !addrOk {
				__antithesis_instrumentation__.Notify(560126)
				return errors.WithHint(
					unimplemented.New("hba-hostnames", "hostname-based HBA rules are not supported"),
					"List the numeric CIDR notation instead, for example: 127.0.0.1/8.\n"+
						"Alternatively, use 'all' (without quotes) for any IPv4/IPv6 address.")
			} else {
				__antithesis_instrumentation__.Notify(560127)
			}
		} else {
			__antithesis_instrumentation__.Notify(560128)
		}
		__antithesis_instrumentation__.Notify(560111)

		method, ok := hbaAuthMethods[entry.Method.Value]
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(560129)
			return method.fn == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(560130)
			return errors.WithHintf(unimplemented.Newf("hba-method-"+entry.Method.Value,
				"unknown auth method %q", entry.Method.Value),
				"Supported methods: %s", listRegisteredMethods())
		} else {
			__antithesis_instrumentation__.Notify(560131)
		}
		__antithesis_instrumentation__.Notify(560112)

		if check := hbaCheckHBAEntries[entry.Method.Value]; check != nil {
			__antithesis_instrumentation__.Notify(560132)
			if err := check(values, entry); err != nil {
				__antithesis_instrumentation__.Notify(560133)
				return err
			} else {
				__antithesis_instrumentation__.Notify(560134)
			}
		} else {
			__antithesis_instrumentation__.Notify(560135)
		}
	}
	__antithesis_instrumentation__.Notify(560101)
	return nil
}

func ParseAndNormalize(val string) (*hba.Conf, error) {
	__antithesis_instrumentation__.Notify(560136)
	conf, err := hba.ParseAndNormalize(val)
	if err != nil {
		__antithesis_instrumentation__.Notify(560140)
		return conf, err
	} else {
		__antithesis_instrumentation__.Notify(560141)
	}
	__antithesis_instrumentation__.Notify(560137)

	if len(conf.Entries) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(560142)
		return !conf.Entries[0].Equivalent(rootEntry) == true
	}() == true {
		__antithesis_instrumentation__.Notify(560143)
		entries := make([]hba.Entry, 1, len(conf.Entries)+1)
		entries[0] = rootEntry
		entries = append(entries, conf.Entries...)
		conf.Entries = entries
	} else {
		__antithesis_instrumentation__.Notify(560144)
	}
	__antithesis_instrumentation__.Notify(560138)

	for i := range conf.Entries {
		__antithesis_instrumentation__.Notify(560145)
		method := conf.Entries[i].Method.Value
		info, ok := hbaAuthMethods[method]
		if !ok {
			__antithesis_instrumentation__.Notify(560147)

			return nil, errors.Errorf("unknown auth method %s", method)
		} else {
			__antithesis_instrumentation__.Notify(560148)
		}
		__antithesis_instrumentation__.Notify(560146)
		conf.Entries[i].MethodFn = info
	}
	__antithesis_instrumentation__.Notify(560139)

	return conf, nil
}

var insecureEntry = hba.Entry{
	ConnType: hba.ConnHostAny,
	User:     []hba.String{{Value: "all", Quoted: false}},
	Address:  hba.AnyAddr{},
	Method:   hba.String{Value: "--insecure"},
}

var sessionRevivalEntry = hba.Entry{
	ConnType: hba.ConnHostAny,
	User:     []hba.String{{Value: "all", Quoted: false}},
	Address:  hba.AnyAddr{},
	Method:   hba.String{Value: "session_revival_token"},
}

var rootEntry = hba.Entry{
	ConnType: hba.ConnHostAny,
	User:     []hba.String{{Value: security.RootUser, Quoted: false}},
	Address:  hba.AnyAddr{},
	Method:   hba.String{Value: "cert-password"},
	Input:    "host  all root all cert-password # CockroachDB mandatory rule",
}

var DefaultHBAConfig = func() *hba.Conf {
	__antithesis_instrumentation__.Notify(560149)
	loadDefaultMethods()
	conf, err := ParseAndNormalize(`
host  all all  all cert-password # built-in CockroachDB default
local all all      password      # built-in CockroachDB default
`)
	if err != nil {
		__antithesis_instrumentation__.Notify(560151)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(560152)
	}
	__antithesis_instrumentation__.Notify(560150)
	return conf
}()

func (s *Server) GetAuthenticationConfiguration() (*hba.Conf, *identmap.Conf) {
	__antithesis_instrumentation__.Notify(560153)
	s.auth.RLock()
	auth := s.auth.conf
	idMap := s.auth.identityMap
	s.auth.RUnlock()

	if auth == nil {
		__antithesis_instrumentation__.Notify(560156)

		auth = DefaultHBAConfig
	} else {
		__antithesis_instrumentation__.Notify(560157)
	}
	__antithesis_instrumentation__.Notify(560154)
	if idMap == nil {
		__antithesis_instrumentation__.Notify(560158)
		idMap = identmap.Empty()
	} else {
		__antithesis_instrumentation__.Notify(560159)
	}
	__antithesis_instrumentation__.Notify(560155)
	return auth, idMap
}

func RegisterAuthMethod(
	method string, fn AuthMethod, validConnTypes hba.ConnType, checkEntry CheckHBAEntry,
) {
	__antithesis_instrumentation__.Notify(560160)
	hbaAuthMethods[method] = methodInfo{validConnTypes, fn}
	if checkEntry != nil {
		__antithesis_instrumentation__.Notify(560161)
		hbaCheckHBAEntries[method] = checkEntry
	} else {
		__antithesis_instrumentation__.Notify(560162)
	}
}

func listRegisteredMethods() string {
	__antithesis_instrumentation__.Notify(560163)
	methods := make([]string, 0, len(hbaAuthMethods))
	for method := range hbaAuthMethods {
		__antithesis_instrumentation__.Notify(560165)
		methods = append(methods, method)
	}
	__antithesis_instrumentation__.Notify(560164)
	sort.Strings(methods)
	return strings.Join(methods, ", ")
}

var (
	hbaAuthMethods     = map[string]methodInfo{}
	hbaCheckHBAEntries = map[string]CheckHBAEntry{}
)

type methodInfo struct {
	validConnTypes hba.ConnType
	fn             AuthMethod
}

type CheckHBAEntry func(*settings.Values, hba.Entry) error

var NoOptionsAllowed CheckHBAEntry = func(sv *settings.Values, e hba.Entry) error {
	__antithesis_instrumentation__.Notify(560166)
	if len(e.Options) != 0 {
		__antithesis_instrumentation__.Notify(560168)
		return errors.Newf("the HBA method %q does not accept options", e.Method)
	} else {
		__antithesis_instrumentation__.Notify(560169)
	}
	__antithesis_instrumentation__.Notify(560167)
	return nil
}

func chainOptions(opts ...CheckHBAEntry) CheckHBAEntry {
	__antithesis_instrumentation__.Notify(560170)
	return func(values *settings.Values, e hba.Entry) error {
		__antithesis_instrumentation__.Notify(560171)
		for _, o := range opts {
			__antithesis_instrumentation__.Notify(560173)
			if err := o(values, e); err != nil {
				__antithesis_instrumentation__.Notify(560174)
				return err
			} else {
				__antithesis_instrumentation__.Notify(560175)
			}
		}
		__antithesis_instrumentation__.Notify(560172)
		return nil
	}
}

func requireClusterVersion(versionkey clusterversion.Key) CheckHBAEntry {
	__antithesis_instrumentation__.Notify(560176)
	return func(values *settings.Values, e hba.Entry) error {
		__antithesis_instrumentation__.Notify(560177)

		var vh clusterversion.Handle
		if values != nil {
			__antithesis_instrumentation__.Notify(560180)
			vh = values.Opaque().(clusterversion.Handle)
		} else {
			__antithesis_instrumentation__.Notify(560181)
		}
		__antithesis_instrumentation__.Notify(560178)
		if vh != nil && func() bool {
			__antithesis_instrumentation__.Notify(560182)
			return !vh.IsActive(context.TODO(), versionkey) == true
		}() == true {
			__antithesis_instrumentation__.Notify(560183)
			return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
				`HBA authentication method %q requires all nodes to be upgraded to %s`,
				e.Method,
				clusterversion.ByKey(versionkey),
			)
		} else {
			__antithesis_instrumentation__.Notify(560184)
		}
		__antithesis_instrumentation__.Notify(560179)
		return nil
	}
}

func (s *Server) HBADebugFn() http.HandlerFunc {
	__antithesis_instrumentation__.Notify(560185)
	return func(w http.ResponseWriter, req *http.Request) {
		__antithesis_instrumentation__.Notify(560186)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		auth, usernames := s.GetAuthenticationConfiguration()

		_, _ = w.Write([]byte("# Active authentication configuration on this node:\n"))
		_, _ = w.Write([]byte(auth.String()))
		if !usernames.Empty() {
			__antithesis_instrumentation__.Notify(560187)
			_, _ = w.Write([]byte("# Active identity mapping on this node:\n"))
			_, _ = w.Write([]byte(usernames.String()))
		} else {
			__antithesis_instrumentation__.Notify(560188)
		}
	}
}
