// Package identmap contains the code for parsing a pg_ident.conf file,
// which allows a database operator to create some number of mappings
// between system identities (e.g.: GSSAPI or X.509 principals) and
// database usernames.
package identmap

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/errors"
	"github.com/olekukonko/tablewriter"
)

type Conf struct {
	data map[string][]element

	originalLines []string

	sortedKeys []string
}

func Empty() *Conf {
	__antithesis_instrumentation__.Notify(560196)
	return &Conf{}
}

var linePattern = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)$`)

func From(r io.Reader) (*Conf, error) {
	__antithesis_instrumentation__.Notify(560197)
	ret := &Conf{data: make(map[string][]element)}
	scanner := bufio.NewScanner(r)
	lineNo := 0

	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(560199)
		lineNo++
		line := scanner.Text()
		ret.originalLines = append(ret.originalLines, line)

		line = tidyIdentMapLine(line)
		if line == "" {
			__antithesis_instrumentation__.Notify(560205)
			continue
		} else {
			__antithesis_instrumentation__.Notify(560206)
		}
		__antithesis_instrumentation__.Notify(560200)

		parts := linePattern.FindStringSubmatch(line)
		if len(parts) != 4 {
			__antithesis_instrumentation__.Notify(560207)
			return nil, errors.Errorf("unable to parse line %d: %q", lineNo, line)
		} else {
			__antithesis_instrumentation__.Notify(560208)
		}
		__antithesis_instrumentation__.Notify(560201)
		mapName := parts[1]

		var sysPattern *regexp.Regexp
		var err error
		if sysName := parts[2]; sysName[0] == '/' {
			__antithesis_instrumentation__.Notify(560209)
			sysPattern, err = regexp.Compile(sysName[1:])
		} else {
			__antithesis_instrumentation__.Notify(560210)
			sysPattern, err = regexp.Compile("^" + regexp.QuoteMeta(sysName) + "$")
		}
		__antithesis_instrumentation__.Notify(560202)
		if err != nil {
			__antithesis_instrumentation__.Notify(560211)
			return nil, errors.Wrapf(err, "unable to parse line %d", lineNo)
		} else {
			__antithesis_instrumentation__.Notify(560212)
		}
		__antithesis_instrumentation__.Notify(560203)

		dbUser := parts[3]
		subIdx := strings.Index(dbUser, `\1`)
		if subIdx >= 0 {
			__antithesis_instrumentation__.Notify(560213)
			if sysPattern.NumSubexp() == 0 {
				__antithesis_instrumentation__.Notify(560214)
				return nil, errors.Errorf(
					`saw \1 substitution on line %d, but pattern contains no subexpressions`, lineNo)
			} else {
				__antithesis_instrumentation__.Notify(560215)
			}
		} else {
			__antithesis_instrumentation__.Notify(560216)
		}
		__antithesis_instrumentation__.Notify(560204)

		elt := element{
			dbUser:       dbUser,
			pattern:      sysPattern,
			substituteAt: subIdx,
		}
		if existing, ok := ret.data[mapName]; ok {
			__antithesis_instrumentation__.Notify(560217)
			ret.data[mapName] = append(existing, elt)
		} else {
			__antithesis_instrumentation__.Notify(560218)
			ret.sortedKeys = append(ret.sortedKeys, mapName)
			ret.data[mapName] = []element{elt}
		}
	}
	__antithesis_instrumentation__.Notify(560198)
	sort.Strings(ret.sortedKeys)
	return ret, nil
}

func (c *Conf) Empty() bool {
	__antithesis_instrumentation__.Notify(560219)
	return c.data == nil || func() bool {
		__antithesis_instrumentation__.Notify(560220)
		return len(c.data) == 0 == true
	}() == true
}

func (c *Conf) Map(mapName, systemIdentity string) ([]security.SQLUsername, error) {
	__antithesis_instrumentation__.Notify(560221)
	if c.data == nil {
		__antithesis_instrumentation__.Notify(560225)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(560226)
	}
	__antithesis_instrumentation__.Notify(560222)
	elts := c.data[mapName]
	if elts == nil {
		__antithesis_instrumentation__.Notify(560227)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(560228)
	}
	__antithesis_instrumentation__.Notify(560223)
	var names []security.SQLUsername
	seen := make(map[string]bool)
	for _, elt := range elts {
		__antithesis_instrumentation__.Notify(560229)
		if n := elt.substitute(systemIdentity); n != "" && func() bool {
			__antithesis_instrumentation__.Notify(560230)
			return !seen[n] == true
		}() == true {
			__antithesis_instrumentation__.Notify(560231)

			u, err := security.MakeSQLUsernameFromUserInput(n, security.UsernameValidation)
			if err != nil {
				__antithesis_instrumentation__.Notify(560233)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(560234)
			}
			__antithesis_instrumentation__.Notify(560232)
			names = append(names, u)
			seen[n] = true
		} else {
			__antithesis_instrumentation__.Notify(560235)
		}
	}
	__antithesis_instrumentation__.Notify(560224)
	return names, nil
}

func (c *Conf) String() string {
	__antithesis_instrumentation__.Notify(560236)
	if len(c.data) == 0 {
		__antithesis_instrumentation__.Notify(560240)
		return "# (empty configuration)"
	} else {
		__antithesis_instrumentation__.Notify(560241)
	}
	__antithesis_instrumentation__.Notify(560237)
	var sb strings.Builder
	sb.WriteString("# Original configuration:\n")
	for _, l := range c.originalLines {
		__antithesis_instrumentation__.Notify(560242)
		fmt.Fprintf(&sb, "# %s\n", l)
	}
	__antithesis_instrumentation__.Notify(560238)
	sb.WriteString("# Active configuration:\n")
	table := tablewriter.NewWriter(&sb)
	table.SetAutoWrapText(false)
	table.SetReflowDuringAutoWrap(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetNoWhiteSpace(true)
	table.SetTrimWhiteSpaceAtEOL(true)
	table.SetTablePadding(" ")

	row := []string{"# map-name", "system-username", "database-username", ""}
	table.Append(row)
	for _, k := range c.sortedKeys {
		__antithesis_instrumentation__.Notify(560243)
		row[0] = k
		for _, elt := range c.data[k] {
			__antithesis_instrumentation__.Notify(560244)
			row[1] = elt.pattern.String()
			row[2] = elt.dbUser
			if elt.substituteAt == -1 {
				__antithesis_instrumentation__.Notify(560246)
				row[3] = ""
			} else {
				__antithesis_instrumentation__.Notify(560247)
				row[3] = fmt.Sprintf("# substituteAt=%d", elt.substituteAt)
			}
			__antithesis_instrumentation__.Notify(560245)
			table.Append(row)
		}
	}
	__antithesis_instrumentation__.Notify(560239)
	table.Render()
	return sb.String()
}

type element struct {
	dbUser string

	pattern *regexp.Regexp

	substituteAt int
}

func (e element) substitute(systemUsername string) string {
	__antithesis_instrumentation__.Notify(560248)
	m := e.pattern.FindStringSubmatch(systemUsername)
	if m == nil {
		__antithesis_instrumentation__.Notify(560251)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(560252)
	}
	__antithesis_instrumentation__.Notify(560249)
	if e.substituteAt == -1 {
		__antithesis_instrumentation__.Notify(560253)
		return e.dbUser
	} else {
		__antithesis_instrumentation__.Notify(560254)
	}
	__antithesis_instrumentation__.Notify(560250)
	return e.dbUser[0:e.substituteAt] + m[1] + e.dbUser[e.substituteAt+2:]
}

func tidyIdentMapLine(line string) string {
	__antithesis_instrumentation__.Notify(560255)
	if commentIdx := strings.IndexByte(line, '#'); commentIdx != -1 {
		__antithesis_instrumentation__.Notify(560257)
		line = line[0:commentIdx]
	} else {
		__antithesis_instrumentation__.Notify(560258)
	}
	__antithesis_instrumentation__.Notify(560256)

	return strings.TrimSpace(line)
}
