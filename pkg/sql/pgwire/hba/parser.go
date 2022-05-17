package hba

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"net"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
)

type scannedInput struct {
	lines   []hbaLine
	linenos []int
}

type hbaLine struct {
	input  string
	tokens [][]String
}

func Parse(input string) (*Conf, error) {
	__antithesis_instrumentation__.Notify(559948)
	tokens, err := tokenize(input)
	if err != nil {
		__antithesis_instrumentation__.Notify(559951)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(559952)
	}
	__antithesis_instrumentation__.Notify(559949)

	var entries []Entry
	for i, line := range tokens.lines {
		__antithesis_instrumentation__.Notify(559953)
		entry, err := parseHbaLine(line)
		if err != nil {
			__antithesis_instrumentation__.Notify(559955)
			return nil, errors.Wrapf(
				pgerror.WithCandidateCode(err, pgcode.ConfigFile),
				"line %d", tokens.linenos[i])
		} else {
			__antithesis_instrumentation__.Notify(559956)
		}
		__antithesis_instrumentation__.Notify(559954)
		entries = append(entries, entry)
	}
	__antithesis_instrumentation__.Notify(559950)

	return &Conf{Entries: entries}, nil
}

func parseHbaLine(inputLine hbaLine) (entry Entry, err error) {
	__antithesis_instrumentation__.Notify(559957)
	fieldIdx := 0

	entry.Input = inputLine.input
	line := inputLine.tokens

	if len(line[fieldIdx]) > 1 {
		__antithesis_instrumentation__.Notify(559967)
		return entry, errors.WithHint(
			errors.New("multiple values specified for connection type"),
			"Specify exactly one connection type per line.")
	} else {
		__antithesis_instrumentation__.Notify(559968)
	}
	__antithesis_instrumentation__.Notify(559958)
	entry.ConnType, err = ParseConnType(line[fieldIdx][0].Value)
	if err != nil {
		__antithesis_instrumentation__.Notify(559969)
		return entry, err
	} else {
		__antithesis_instrumentation__.Notify(559970)
	}
	__antithesis_instrumentation__.Notify(559959)

	fieldIdx++
	if fieldIdx >= len(line) {
		__antithesis_instrumentation__.Notify(559971)
		return entry, errors.New("end-of-line before database specification")
	} else {
		__antithesis_instrumentation__.Notify(559972)
	}
	__antithesis_instrumentation__.Notify(559960)
	entry.Database = line[fieldIdx]

	fieldIdx++
	if fieldIdx >= len(line) {
		__antithesis_instrumentation__.Notify(559973)
		return entry, errors.New("end-of-line before role specification")
	} else {
		__antithesis_instrumentation__.Notify(559974)
	}
	__antithesis_instrumentation__.Notify(559961)
	entry.User = line[fieldIdx]

	if entry.ConnType != ConnLocal {
		__antithesis_instrumentation__.Notify(559975)
		fieldIdx++
		if fieldIdx >= len(line) {
			__antithesis_instrumentation__.Notify(559978)
			return entry, errors.New("end-of-line before IP address specification")
		} else {
			__antithesis_instrumentation__.Notify(559979)
		}
		__antithesis_instrumentation__.Notify(559976)
		tokens := line[fieldIdx]
		if len(tokens) > 1 {
			__antithesis_instrumentation__.Notify(559980)
			return entry, errors.WithHint(
				errors.New("multiple values specified for host address"),
				"Specify one address range per line.")
		} else {
			__antithesis_instrumentation__.Notify(559981)
		}
		__antithesis_instrumentation__.Notify(559977)
		token := tokens[0]
		switch {
		case token.Value == "":
			__antithesis_instrumentation__.Notify(559982)
			return entry, errors.New("cannot use empty string as address")
		case token.IsKeyword("all"):
			__antithesis_instrumentation__.Notify(559983)
			entry.Address = token
		case token.IsKeyword("samehost"), token.IsKeyword("samenet"):
			__antithesis_instrumentation__.Notify(559984)
			return entry, unimplemented.Newf(
				fmt.Sprintf("hba-net-%s", token.Value),
				"address specification %s is not yet supported", errors.Safe(token.Value))
		default:
			__antithesis_instrumentation__.Notify(559985)

			addr := token.Value
			if strings.Contains(addr, "/") {
				__antithesis_instrumentation__.Notify(559986)
				_, ipnet, err := net.ParseCIDR(addr)
				if err != nil {
					__antithesis_instrumentation__.Notify(559988)
					return entry, err
				} else {
					__antithesis_instrumentation__.Notify(559989)
				}
				__antithesis_instrumentation__.Notify(559987)
				entry.Address = ipnet
			} else {
				__antithesis_instrumentation__.Notify(559990)
				var ip net.IP
				hostname := addr
				if ip = net.ParseIP(addr); ip != nil {
					__antithesis_instrumentation__.Notify(559992)
					hostname = ""
				} else {
					__antithesis_instrumentation__.Notify(559993)
				}
				__antithesis_instrumentation__.Notify(559991)
				if hostname != "" {
					__antithesis_instrumentation__.Notify(559994)
					entry.Address = String{Value: addr, Quoted: token.Quoted}
				} else {
					__antithesis_instrumentation__.Notify(559995)

					fieldIdx++
					if fieldIdx >= len(line) {
						__antithesis_instrumentation__.Notify(560000)
						return entry, errors.WithHint(
							errors.New("end-of-line before netmask specification"),
							"Specify an address range in CIDR notation, or provide a separate netmask.")
					} else {
						__antithesis_instrumentation__.Notify(560001)
					}
					__antithesis_instrumentation__.Notify(559996)
					if len(line[fieldIdx]) > 1 {
						__antithesis_instrumentation__.Notify(560002)
						return entry, errors.New("multiple values specified for netmask")
					} else {
						__antithesis_instrumentation__.Notify(560003)
					}
					__antithesis_instrumentation__.Notify(559997)
					maybeMask := net.ParseIP(line[fieldIdx][0].Value)
					if err := checkMask(maybeMask); err != nil {
						__antithesis_instrumentation__.Notify(560004)
						return entry, errors.Wrapf(err, "invalid IP mask \"%s\"", line[fieldIdx][0].Value)
					} else {
						__antithesis_instrumentation__.Notify(560005)
					}
					__antithesis_instrumentation__.Notify(559998)

					if (maybeMask.To4() == nil) != (ip.To4() == nil) {
						__antithesis_instrumentation__.Notify(560006)
						return entry, errors.Newf("IP address and mask do not match")
					} else {
						__antithesis_instrumentation__.Notify(560007)
					}
					__antithesis_instrumentation__.Notify(559999)
					mask := net.IPMask(maybeMask)
					entry.Address = &net.IPNet{IP: ip.Mask(mask), Mask: mask}
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(560008)
	}
	__antithesis_instrumentation__.Notify(559962)

	fieldIdx++
	if fieldIdx >= len(line) {
		__antithesis_instrumentation__.Notify(560009)
		return entry, errors.New("end-of-line before authentication method")
	} else {
		__antithesis_instrumentation__.Notify(560010)
	}
	__antithesis_instrumentation__.Notify(559963)
	if len(line[fieldIdx]) > 1 {
		__antithesis_instrumentation__.Notify(560011)
		return entry, errors.WithHint(
			errors.New("multiple values specified for authentication method"),
			"Specify exactly one authentication method per line.")
	} else {
		__antithesis_instrumentation__.Notify(560012)
	}
	__antithesis_instrumentation__.Notify(559964)
	entry.Method = line[fieldIdx][0]
	if entry.Method.Value == "" {
		__antithesis_instrumentation__.Notify(560013)
		return entry, errors.New("cannot use empty string as authentication method")
	} else {
		__antithesis_instrumentation__.Notify(560014)
	}
	__antithesis_instrumentation__.Notify(559965)

	for fieldIdx++; fieldIdx < len(line); fieldIdx++ {
		__antithesis_instrumentation__.Notify(560015)
		for _, tok := range line[fieldIdx] {
			__antithesis_instrumentation__.Notify(560016)
			kv := strings.SplitN(tok.Value, "=", 2)
			if len(kv) != 2 {
				__antithesis_instrumentation__.Notify(560018)
				return entry, errors.Newf("authentication option not in name=value format: %s", tok.Value)
			} else {
				__antithesis_instrumentation__.Notify(560019)
			}
			__antithesis_instrumentation__.Notify(560017)
			entry.Options = append(entry.Options, [2]string{kv[0], kv[1]})
			entry.OptionQuotes = append(entry.OptionQuotes, tok.Quoted)
		}
	}
	__antithesis_instrumentation__.Notify(559966)

	return entry, nil
}

func checkMask(maybeMask net.IP) error {
	__antithesis_instrumentation__.Notify(560020)
	if maybeMask == nil {
		__antithesis_instrumentation__.Notify(560027)
		return errors.New("netmask not in IP numeric format")
	} else {
		__antithesis_instrumentation__.Notify(560028)
	}
	__antithesis_instrumentation__.Notify(560021)
	if ip4 := maybeMask.To4(); ip4 != nil {
		__antithesis_instrumentation__.Notify(560029)
		maybeMask = ip4
	} else {
		__antithesis_instrumentation__.Notify(560030)
	}
	__antithesis_instrumentation__.Notify(560022)
	i := 0

	for ; i < len(maybeMask) && func() bool {
		__antithesis_instrumentation__.Notify(560031)
		return maybeMask[i] == '\xff' == true
	}() == true; i++ {
		__antithesis_instrumentation__.Notify(560032)
	}
	__antithesis_instrumentation__.Notify(560023)

	if i < len(maybeMask) {
		__antithesis_instrumentation__.Notify(560033)
		switch maybeMask[i] {
		case 0xff, 0xfe, 0xfc, 0xf8, 0xf0, 0xe0, 0xc0, 0x80:
			__antithesis_instrumentation__.Notify(560034)
			i++
		default:
			__antithesis_instrumentation__.Notify(560035)
		}
	} else {
		__antithesis_instrumentation__.Notify(560036)
	}
	__antithesis_instrumentation__.Notify(560024)

	for ; i < len(maybeMask) && func() bool {
		__antithesis_instrumentation__.Notify(560037)
		return maybeMask[i] == '\x00' == true
	}() == true; i++ {
		__antithesis_instrumentation__.Notify(560038)
	}
	__antithesis_instrumentation__.Notify(560025)

	if i < len(maybeMask) {
		__antithesis_instrumentation__.Notify(560039)
		return errors.New("address is not a mask")
	} else {
		__antithesis_instrumentation__.Notify(560040)
	}
	__antithesis_instrumentation__.Notify(560026)
	return nil
}

func ParseConnType(s string) (ConnType, error) {
	__antithesis_instrumentation__.Notify(560041)
	switch s {
	case "local":
		__antithesis_instrumentation__.Notify(560043)
		return ConnLocal, nil
	case "host":
		__antithesis_instrumentation__.Notify(560044)
		return ConnHostAny, nil
	case "hostssl":
		__antithesis_instrumentation__.Notify(560045)
		return ConnHostSSL, nil
	case "hostnossl":
		__antithesis_instrumentation__.Notify(560046)
		return ConnHostNoSSL, nil
	default:
		__antithesis_instrumentation__.Notify(560047)
	}
	__antithesis_instrumentation__.Notify(560042)
	return 0, errors.Newf("unknown connection type: %q", s)
}
