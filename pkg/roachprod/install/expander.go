package install

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/errors"
)

var parameterRe = regexp.MustCompile(`{[^{}]*}`)
var pgURLRe = regexp.MustCompile(`{pgurl(:[-,0-9]+)?}`)
var pgHostRe = regexp.MustCompile(`{pghost(:[-,0-9]+)?}`)
var pgPortRe = regexp.MustCompile(`{pgport(:[-,0-9]+)?}`)
var uiPortRe = regexp.MustCompile(`{uiport(:[-,0-9]+)?}`)
var storeDirRe = regexp.MustCompile(`{store-dir}`)
var logDirRe = regexp.MustCompile(`{log-dir}`)
var certsDirRe = regexp.MustCompile(`{certs-dir}`)

type expander struct {
	node Node

	pgURLs  map[Node]string
	pgHosts map[Node]string
	pgPorts map[Node]string
	uiPorts map[Node]string
}

type expanderFunc func(context.Context, *logger.Logger, *SyncedCluster, string) (expanded string, didExpand bool, err error)

func (e *expander) expand(
	ctx context.Context, l *logger.Logger, c *SyncedCluster, arg string,
) (string, error) {
	__antithesis_instrumentation__.Notify(181517)
	var err error
	s := parameterRe.ReplaceAllStringFunc(arg, func(s string) string {
		__antithesis_instrumentation__.Notify(181520)
		if err != nil {
			__antithesis_instrumentation__.Notify(181523)
			return ""
		} else {
			__antithesis_instrumentation__.Notify(181524)
		}
		__antithesis_instrumentation__.Notify(181521)
		expanders := []expanderFunc{
			e.maybeExpandPgURL,
			e.maybeExpandPgHost,
			e.maybeExpandPgPort,
			e.maybeExpandUIPort,
			e.maybeExpandStoreDir,
			e.maybeExpandLogDir,
			e.maybeExpandCertsDir,
		}
		for _, f := range expanders {
			__antithesis_instrumentation__.Notify(181525)
			v, expanded, fErr := f(ctx, l, c, s)
			if fErr != nil {
				__antithesis_instrumentation__.Notify(181527)
				err = fErr
				return ""
			} else {
				__antithesis_instrumentation__.Notify(181528)
			}
			__antithesis_instrumentation__.Notify(181526)
			if expanded {
				__antithesis_instrumentation__.Notify(181529)
				return v
			} else {
				__antithesis_instrumentation__.Notify(181530)
			}
		}
		__antithesis_instrumentation__.Notify(181522)
		return s
	})
	__antithesis_instrumentation__.Notify(181518)
	if err != nil {
		__antithesis_instrumentation__.Notify(181531)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(181532)
	}
	__antithesis_instrumentation__.Notify(181519)
	return s, nil
}

func (e *expander) maybeExpandMap(
	c *SyncedCluster, m map[Node]string, nodeSpec string,
) (string, error) {
	__antithesis_instrumentation__.Notify(181533)
	if nodeSpec == "" {
		__antithesis_instrumentation__.Notify(181538)
		nodeSpec = "all"
	} else {
		__antithesis_instrumentation__.Notify(181539)
		nodeSpec = nodeSpec[1:]
	}
	__antithesis_instrumentation__.Notify(181534)

	nodes, err := ListNodes(nodeSpec, len(c.VMs))
	if err != nil {
		__antithesis_instrumentation__.Notify(181540)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(181541)
	}
	__antithesis_instrumentation__.Notify(181535)

	var result []string
	for _, node := range nodes {
		__antithesis_instrumentation__.Notify(181542)
		if s, ok := m[node]; ok {
			__antithesis_instrumentation__.Notify(181543)
			result = append(result, s)
		} else {
			__antithesis_instrumentation__.Notify(181544)
		}
	}
	__antithesis_instrumentation__.Notify(181536)
	if len(result) != len(nodes) {
		__antithesis_instrumentation__.Notify(181545)
		return "", errors.Errorf("failed to expand nodes %v, given node map %v", nodes, m)
	} else {
		__antithesis_instrumentation__.Notify(181546)
	}
	__antithesis_instrumentation__.Notify(181537)
	return strings.Join(result, " "), nil
}

func (e *expander) maybeExpandPgURL(
	ctx context.Context, l *logger.Logger, c *SyncedCluster, s string,
) (string, bool, error) {
	__antithesis_instrumentation__.Notify(181547)
	m := pgURLRe.FindStringSubmatch(s)
	if m == nil {
		__antithesis_instrumentation__.Notify(181550)
		return s, false, nil
	} else {
		__antithesis_instrumentation__.Notify(181551)
	}
	__antithesis_instrumentation__.Notify(181548)

	if e.pgURLs == nil {
		__antithesis_instrumentation__.Notify(181552)
		var err error
		e.pgURLs, err = c.pgurls(ctx, l, allNodes(len(c.VMs)))
		if err != nil {
			__antithesis_instrumentation__.Notify(181553)
			return "", false, err
		} else {
			__antithesis_instrumentation__.Notify(181554)
		}
	} else {
		__antithesis_instrumentation__.Notify(181555)
	}
	__antithesis_instrumentation__.Notify(181549)

	s, err := e.maybeExpandMap(c, e.pgURLs, m[1])
	return s, err == nil, err
}

func (e *expander) maybeExpandPgHost(
	ctx context.Context, l *logger.Logger, c *SyncedCluster, s string,
) (string, bool, error) {
	__antithesis_instrumentation__.Notify(181556)
	m := pgHostRe.FindStringSubmatch(s)
	if m == nil {
		__antithesis_instrumentation__.Notify(181559)
		return s, false, nil
	} else {
		__antithesis_instrumentation__.Notify(181560)
	}
	__antithesis_instrumentation__.Notify(181557)

	if e.pgHosts == nil {
		__antithesis_instrumentation__.Notify(181561)
		var err error
		e.pgHosts, err = c.pghosts(ctx, l, allNodes(len(c.VMs)))
		if err != nil {
			__antithesis_instrumentation__.Notify(181562)
			return "", false, err
		} else {
			__antithesis_instrumentation__.Notify(181563)
		}
	} else {
		__antithesis_instrumentation__.Notify(181564)
	}
	__antithesis_instrumentation__.Notify(181558)

	s, err := e.maybeExpandMap(c, e.pgHosts, m[1])
	return s, err == nil, err
}

func (e *expander) maybeExpandPgPort(
	ctx context.Context, l *logger.Logger, c *SyncedCluster, s string,
) (string, bool, error) {
	__antithesis_instrumentation__.Notify(181565)
	m := pgPortRe.FindStringSubmatch(s)
	if m == nil {
		__antithesis_instrumentation__.Notify(181568)
		return s, false, nil
	} else {
		__antithesis_instrumentation__.Notify(181569)
	}
	__antithesis_instrumentation__.Notify(181566)

	if e.pgPorts == nil {
		__antithesis_instrumentation__.Notify(181570)
		e.pgPorts = make(map[Node]string, len(c.VMs))
		for _, node := range allNodes(len(c.VMs)) {
			__antithesis_instrumentation__.Notify(181571)
			e.pgPorts[node] = fmt.Sprint(c.NodePort(node))
		}
	} else {
		__antithesis_instrumentation__.Notify(181572)
	}
	__antithesis_instrumentation__.Notify(181567)

	s, err := e.maybeExpandMap(c, e.pgPorts, m[1])
	return s, err == nil, err
}

func (e *expander) maybeExpandUIPort(
	ctx context.Context, l *logger.Logger, c *SyncedCluster, s string,
) (string, bool, error) {
	__antithesis_instrumentation__.Notify(181573)
	m := uiPortRe.FindStringSubmatch(s)
	if m == nil {
		__antithesis_instrumentation__.Notify(181576)
		return s, false, nil
	} else {
		__antithesis_instrumentation__.Notify(181577)
	}
	__antithesis_instrumentation__.Notify(181574)

	if e.uiPorts == nil {
		__antithesis_instrumentation__.Notify(181578)
		e.uiPorts = make(map[Node]string, len(c.VMs))
		for _, node := range allNodes(len(c.VMs)) {
			__antithesis_instrumentation__.Notify(181579)
			e.uiPorts[node] = fmt.Sprint(c.NodeUIPort(node))
		}
	} else {
		__antithesis_instrumentation__.Notify(181580)
	}
	__antithesis_instrumentation__.Notify(181575)

	s, err := e.maybeExpandMap(c, e.uiPorts, m[1])
	return s, err == nil, err
}

func (e *expander) maybeExpandStoreDir(
	ctx context.Context, l *logger.Logger, c *SyncedCluster, s string,
) (string, bool, error) {
	__antithesis_instrumentation__.Notify(181581)
	if !storeDirRe.MatchString(s) {
		__antithesis_instrumentation__.Notify(181583)
		return s, false, nil
	} else {
		__antithesis_instrumentation__.Notify(181584)
	}
	__antithesis_instrumentation__.Notify(181582)
	return c.NodeDir(e.node, 1), true, nil
}

func (e *expander) maybeExpandLogDir(
	ctx context.Context, l *logger.Logger, c *SyncedCluster, s string,
) (string, bool, error) {
	__antithesis_instrumentation__.Notify(181585)
	if !logDirRe.MatchString(s) {
		__antithesis_instrumentation__.Notify(181587)
		return s, false, nil
	} else {
		__antithesis_instrumentation__.Notify(181588)
	}
	__antithesis_instrumentation__.Notify(181586)
	return c.LogDir(e.node), true, nil
}

func (e *expander) maybeExpandCertsDir(
	ctx context.Context, l *logger.Logger, c *SyncedCluster, s string,
) (string, bool, error) {
	__antithesis_instrumentation__.Notify(181589)
	if !certsDirRe.MatchString(s) {
		__antithesis_instrumentation__.Notify(181591)
		return s, false, nil
	} else {
		__antithesis_instrumentation__.Notify(181592)
	}
	__antithesis_instrumentation__.Notify(181590)
	return c.CertsDir(e.node), true, nil
}
