package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

const RemoteNodeID = "remote_node_id"

type nodeProxy struct {
	scheme               string
	rpcContext           *rpc.Context
	parseNodeID          ParseNodeIDFn
	getNodeIDHTTPAddress GetNodeIDHTTPAddressFn
	proxyCache           reverseProxyCache
}

type ParseNodeIDFn func(string) (roachpb.NodeID, bool, error)

type GetNodeIDHTTPAddressFn func(roachpb.NodeID) (*util.UnresolvedAddr, error)

func (np *nodeProxy) nodeProxyHandler(
	w http.ResponseWriter, r *http.Request, next http.HandlerFunc,
) {
	__antithesis_instrumentation__.Notify(194801)
	ctx := r.Context()
	nodeIDString, clearCookieOnError, err := np.getNodeIDFromRequest(r)
	if err != nil {
		__antithesis_instrumentation__.Notify(194805)
		if errors.Is(err, ErrNoNodeID) {
			__antithesis_instrumentation__.Notify(194808)
			next(w, r)
			return
		} else {
			__antithesis_instrumentation__.Notify(194809)
		}
		__antithesis_instrumentation__.Notify(194806)
		if clearCookieOnError {
			__antithesis_instrumentation__.Notify(194810)
			resetCookie(w, r)
		} else {
			__antithesis_instrumentation__.Notify(194811)
		}
		__antithesis_instrumentation__.Notify(194807)
		panic(errors.Wrap(err, "server: unexpected error reading nodeID from request"))
	} else {
		__antithesis_instrumentation__.Notify(194812)
	}
	__antithesis_instrumentation__.Notify(194802)
	nodeID, local, err := np.parseNodeID(nodeIDString)
	if err != nil {
		__antithesis_instrumentation__.Notify(194813)
		httpErr := errors.Wrapf(err, "server: error parsing nodeID from request: %s", nodeIDString)
		log.Errorf(ctx, "%v", httpErr)
		if clearCookieOnError {
			__antithesis_instrumentation__.Notify(194815)
			resetCookie(w, r)
		} else {
			__antithesis_instrumentation__.Notify(194816)
		}
		__antithesis_instrumentation__.Notify(194814)

		http.Error(w, httpErr.Error(), http.StatusBadRequest)
	} else {
		__antithesis_instrumentation__.Notify(194817)
	}
	__antithesis_instrumentation__.Notify(194803)
	if local {
		__antithesis_instrumentation__.Notify(194818)
		next(w, r)
		return
	} else {
		__antithesis_instrumentation__.Notify(194819)
	}
	__antithesis_instrumentation__.Notify(194804)
	np.routeToNode(w, r, clearCookieOnError, nodeID)
}

var ErrNoNodeID = errors.New("http: nodeID not present in request")

func (np *nodeProxy) getNodeIDFromRequest(
	r *http.Request,
) (nodeIDString string, nodeIDFromCookie bool, retErr error) {
	__antithesis_instrumentation__.Notify(194820)
	queryParam := r.URL.Query().Get(RemoteNodeID)
	if queryParam != "" {
		__antithesis_instrumentation__.Notify(194823)
		return queryParam, false, nil
	} else {
		__antithesis_instrumentation__.Notify(194824)
	}
	__antithesis_instrumentation__.Notify(194821)
	cookie, err := r.Cookie(RemoteNodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(194825)
		if errors.Is(err, http.ErrNoCookie) {
			__antithesis_instrumentation__.Notify(194827)
			return "", false, ErrNoNodeID
		} else {
			__antithesis_instrumentation__.Notify(194828)
		}
		__antithesis_instrumentation__.Notify(194826)

		return "", false, errors.Wrapf(err, "server: error decoding node routing cookie")
	} else {
		__antithesis_instrumentation__.Notify(194829)
	}
	__antithesis_instrumentation__.Notify(194822)
	return cookie.Value, true, nil
}

func (np *nodeProxy) routeToNode(
	w http.ResponseWriter, r *http.Request, clearCookieOnError bool, nodeID roachpb.NodeID,
) {
	__antithesis_instrumentation__.Notify(194830)
	addr, err := np.getNodeIDHTTPAddress(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(194833)
		httpErr := errors.Wrapf(err, "unable to get address for n%d", nodeID)
		log.Errorf(r.Context(), "%v", httpErr)

		if clearCookieOnError {
			__antithesis_instrumentation__.Notify(194835)
			resetCookie(w, r)
		} else {
			__antithesis_instrumentation__.Notify(194836)
		}
		__antithesis_instrumentation__.Notify(194834)

		http.Error(w, httpErr.Error(), http.StatusBadRequest)
		return
	} else {
		__antithesis_instrumentation__.Notify(194837)
	}
	__antithesis_instrumentation__.Notify(194831)
	proxy, ok := np.proxyCache.get(*addr)
	if !ok {
		__antithesis_instrumentation__.Notify(194838)
		u := url.URL{
			Scheme: np.scheme,
			Host:   addr.AddressField,
		}
		proxy = httputil.NewSingleHostReverseProxy(&u)

		httpClient, err := np.rpcContext.GetHTTPClient()
		if err != nil {
			__antithesis_instrumentation__.Notify(194840)
			if clearCookieOnError {
				__antithesis_instrumentation__.Notify(194842)
				resetCookie(w, r)
			} else {
				__antithesis_instrumentation__.Notify(194843)
			}
			__antithesis_instrumentation__.Notify(194841)
			panic(errors.Wrapf(err, "server: failed to get httpClient"))
		} else {
			__antithesis_instrumentation__.Notify(194844)
		}
		__antithesis_instrumentation__.Notify(194839)
		proxy.Transport = httpClient.Transport
		np.proxyCache.put(*addr, proxy)
	} else {
		__antithesis_instrumentation__.Notify(194845)
	}
	__antithesis_instrumentation__.Notify(194832)
	proxy.ServeHTTP(w, r)
}

func resetCookie(w http.ResponseWriter, req *http.Request) {
	__antithesis_instrumentation__.Notify(194846)
	if _, err := req.Cookie(RemoteNodeID); err == nil {
		__antithesis_instrumentation__.Notify(194847)
		w.Header().Set("set-cookie", fmt.Sprintf("%s=;expires=Thu, 01 Jan 1970 00:00:01 GMT;path=/", RemoteNodeID))
	} else {
		__antithesis_instrumentation__.Notify(194848)
	}
}

type reverseProxyCache struct {
	mu            syncutil.RWMutex
	proxiesByAddr map[util.UnresolvedAddr]*httputil.ReverseProxy
}

func (c *reverseProxyCache) get(addr util.UnresolvedAddr) (*httputil.ReverseProxy, bool) {
	__antithesis_instrumentation__.Notify(194849)
	c.mu.RLock()
	defer c.mu.RUnlock()
	p, ok := c.proxiesByAddr[addr]
	return p, ok
}

func (c *reverseProxyCache) put(addr util.UnresolvedAddr, proxy *httputil.ReverseProxy) {
	__antithesis_instrumentation__.Notify(194850)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.proxiesByAddr == nil {
		__antithesis_instrumentation__.Notify(194852)
		c.proxiesByAddr = make(map[util.UnresolvedAddr]*httputil.ReverseProxy)
	} else {
		__antithesis_instrumentation__.Notify(194853)
	}
	__antithesis_instrumentation__.Notify(194851)
	c.proxiesByAddr[addr] = proxy
}
