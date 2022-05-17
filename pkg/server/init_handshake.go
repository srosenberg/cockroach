package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

const (
	initServiceName     = "temp-init-service"
	caCommonName        = "Cockroach CA"
	defaultInitLifespan = 10 * time.Minute

	trustInitURL     = "/trustInit/"
	deliverBundleURL = "/deliverBundle/"
)

var errInvalidHMAC = errors.New("invalid HMAC signature")

type nodeHostnameAndCA struct {
	HostAddress   string `json:"host_address"`
	CACertificate []byte `json:"ca_certificate"`
	HMAC          []byte `json:"hmac,omitempty"`
}

func createNodeHostnameAndCA(
	hostAddress string, caCert []byte, secretToken []byte,
) (nodeHostnameAndCA, error) {
	__antithesis_instrumentation__.Notify(194041)
	signedMessage := nodeHostnameAndCA{
		HostAddress:   hostAddress,
		CACertificate: caCert,
	}
	h := hmac.New(sha256.New, secretToken)
	h.Write([]byte(hostAddress))
	h.Write(caCert)
	signedMessage.HMAC = h.Sum(nil)

	return signedMessage, nil
}

func (n *nodeHostnameAndCA) validHMAC(secretToken []byte) bool {
	__antithesis_instrumentation__.Notify(194042)
	h := hmac.New(sha256.New, secretToken)
	h.Write([]byte(n.HostAddress))
	h.Write(n.CACertificate)
	expectedMac := h.Sum(nil)
	return hmac.Equal(expectedMac, n.HMAC)
}

type nodeTrustBundle struct {
	Bundle CertificateBundle `json:"certificate_bundle"`

	HMAC []byte `json:"hmac,omitempty"`
}

func (n *nodeTrustBundle) computeHMAC(secretToken []byte) []byte {
	__antithesis_instrumentation__.Notify(194043)
	h := hmac.New(sha256.New, secretToken)
	h.Write(n.Bundle.InterNode.CACertificate)
	h.Write(n.Bundle.InterNode.CAKey)
	h.Write(n.Bundle.UserAuth.CACertificate)
	h.Write(n.Bundle.UserAuth.CAKey)
	h.Write(n.Bundle.SQLService.CACertificate)
	h.Write(n.Bundle.SQLService.CAKey)
	h.Write(n.Bundle.RPCService.CACertificate)
	h.Write(n.Bundle.RPCService.CAKey)
	h.Write(n.Bundle.AdminUIService.CACertificate)
	h.Write(n.Bundle.AdminUIService.CAKey)
	return h.Sum(nil)
}

func (n *nodeTrustBundle) signHMAC(secretToken []byte) {
	__antithesis_instrumentation__.Notify(194044)
	n.HMAC = n.computeHMAC(secretToken)
}

func (n *nodeTrustBundle) validHMAC(secretToken []byte) bool {
	__antithesis_instrumentation__.Notify(194045)
	expectedMac := n.computeHMAC(secretToken)
	return hmac.Equal(expectedMac, n.HMAC)
}

func pemToSignature(caCertPEM []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(194046)
	caCert, _ := pem.Decode(caCertPEM)
	if nil == caCert {
		__antithesis_instrumentation__.Notify(194049)
		return nil, errors.New("failed to parse valid PEM from CACertificate blob")
	} else {
		__antithesis_instrumentation__.Notify(194050)
	}
	__antithesis_instrumentation__.Notify(194047)

	cert, err := x509.ParseCertificate(caCert.Bytes)
	if nil != err {
		__antithesis_instrumentation__.Notify(194051)
		return nil, errors.New("failed to parse valid certificate from CACertificate blob")
	} else {
		__antithesis_instrumentation__.Notify(194052)
	}
	__antithesis_instrumentation__.Notify(194048)

	return cert.Signature, nil
}

func createNodeInitTempCertificates(
	ctx context.Context, hostnames []string, lifespan time.Duration,
) (certs ServiceCertificateBundle, err error) {
	__antithesis_instrumentation__.Notify(194053)
	log.Ops.Infof(ctx, "creating temporary initial certificates for hosts %+v, duration %s", hostnames, lifespan)

	caCtx := logtags.AddTag(ctx, "create-temp-ca", nil)
	caCertPEM, caKeyPEM, err := security.CreateCACertAndKey(caCtx, log.Ops.Infof, lifespan, initServiceName)
	if err != nil {
		__antithesis_instrumentation__.Notify(194056)
		return certs, err
	} else {
		__antithesis_instrumentation__.Notify(194057)
	}
	__antithesis_instrumentation__.Notify(194054)
	serviceCtx := logtags.AddTag(ctx, "create-temp-service", nil)
	serviceCertPEM, serviceKeyPEM, err := security.CreateServiceCertAndKey(
		serviceCtx,
		log.Ops.Infof,
		lifespan,
		security.NodeUser,
		hostnames,
		caCertPEM,
		caKeyPEM,
		false,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(194058)
		return certs, err
	} else {
		__antithesis_instrumentation__.Notify(194059)
	}
	__antithesis_instrumentation__.Notify(194055)
	caCert := pem.EncodeToMemory(caCertPEM)
	caKey := pem.EncodeToMemory(caKeyPEM)
	serviceCert := pem.EncodeToMemory(serviceCertPEM)
	serviceKey := pem.EncodeToMemory(serviceKeyPEM)

	certs = ServiceCertificateBundle{
		CACertificate:   caCert,
		CAKey:           caKey,
		HostCertificate: serviceCert,
		HostKey:         serviceKey,
	}
	return certs, nil
}

func sendBadRequestError(ctx context.Context, err error, w http.ResponseWriter) {
	__antithesis_instrumentation__.Notify(194060)
	http.Error(w, "invalid request message", http.StatusBadRequest)
	log.Ops.Warningf(ctx, "bad request: %s", err)
}

func generateURLForClient(peer string, endpoint string) string {
	__antithesis_instrumentation__.Notify(194061)
	return fmt.Sprintf("https://%s%s", peer, endpoint)
}

type tlsInitHandshaker struct {
	server *http.Server

	token      []byte
	certsDir   string
	listenAddr string

	tempCerts    ServiceCertificateBundle
	trustedPeers chan nodeHostnameAndCA
	finishedInit chan *CertificateBundle
	errors       chan error
	wg           sync.WaitGroup
}

func (t *tlsInitHandshaker) init(ctx context.Context) error {
	serverCert, err := tls.X509KeyPair(t.tempCerts.HostCertificate, t.tempCerts.HostKey)
	if err != nil {
		return err
	}

	certpool := x509.NewCertPool()
	if ok := certpool.AppendCertsFromPEM(t.tempCerts.CACertificate); !ok {
		return errors.New("could not add temp CA certificate to cert pool")
	}
	serviceTLSConf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		RootCAs:      certpool,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(trustInitURL, enhanceHandlerContextWithHTTPClient(ctx, t.onTrustInit))
	mux.HandleFunc(deliverBundleURL, enhanceHandlerContextWithHTTPClient(ctx, t.onDeliverBundle))

	t.server = &http.Server{
		Addr:      t.listenAddr,
		Handler:   mux,
		TLSConfig: serviceTLSConf,
	}
	return nil
}

func enhanceHandlerContextWithHTTPClient(
	baseCtx context.Context, fn func(ctx context.Context, w http.ResponseWriter, req *http.Request),
) http.HandlerFunc {
	__antithesis_instrumentation__.Notify(194062)
	return func(w http.ResponseWriter, req *http.Request) {
		__antithesis_instrumentation__.Notify(194063)
		ctx := logtags.AddTag(baseCtx, "peer", req.RemoteAddr)
		fn(ctx, w, req)
	}
}

func (t *tlsInitHandshaker) onTrustInit(
	ctx context.Context, res http.ResponseWriter, req *http.Request,
) {
	__antithesis_instrumentation__.Notify(194064)
	var challenge nodeHostnameAndCA

	err := json.NewDecoder(req.Body).Decode(&challenge)
	if err != nil {
		__antithesis_instrumentation__.Notify(194068)
		sendBadRequestError(ctx, errors.Wrap(err, "error when unmarshalling challenge"), res)
		return
	} else {
		__antithesis_instrumentation__.Notify(194069)
	}
	__antithesis_instrumentation__.Notify(194065)
	defer req.Body.Close()

	if !challenge.validHMAC(t.token) {
		__antithesis_instrumentation__.Notify(194070)
		sendBadRequestError(ctx, errInvalidHMAC, res)

		select {
		case t.errors <- errInvalidHMAC:
			__antithesis_instrumentation__.Notify(194072)
		default:
			__antithesis_instrumentation__.Notify(194073)
		}
		__antithesis_instrumentation__.Notify(194071)

		return
	} else {
		__antithesis_instrumentation__.Notify(194074)
	}
	__antithesis_instrumentation__.Notify(194066)

	log.Ops.Infof(ctx, "received valid challenge and CA from: %s", challenge.HostAddress)

	t.trustedPeers <- challenge

	ack, err := createNodeHostnameAndCA(t.listenAddr, t.tempCerts.CACertificate, t.token)
	if err != nil {
		__antithesis_instrumentation__.Notify(194075)
		apiV2InternalError(req.Context(), err, res)
		return
	} else {
		__antithesis_instrumentation__.Notify(194076)
	}
	__antithesis_instrumentation__.Notify(194067)
	if err := json.NewEncoder(res).Encode(ack); err != nil {
		__antithesis_instrumentation__.Notify(194077)
		apiV2InternalError(req.Context(), err, res)
		return
	} else {
		__antithesis_instrumentation__.Notify(194078)
	}
}

func (t *tlsInitHandshaker) onDeliverBundle(
	ctx context.Context, res http.ResponseWriter, req *http.Request,
) {
	__antithesis_instrumentation__.Notify(194079)
	bundle := nodeTrustBundle{}
	err := json.NewDecoder(req.Body).Decode(&bundle)
	defer req.Body.Close()
	if err != nil {
		__antithesis_instrumentation__.Notify(194082)
		sendBadRequestError(ctx, errors.Wrap(err, "error when unmarshalling bundle"), res)
		return
	} else {
		__antithesis_instrumentation__.Notify(194083)
	}
	__antithesis_instrumentation__.Notify(194080)
	if !bundle.validHMAC(t.token) {
		__antithesis_instrumentation__.Notify(194084)
		sendBadRequestError(ctx, errors.New("invalid bundle HMAC"), res)
		return
	} else {
		__antithesis_instrumentation__.Notify(194085)
	}
	__antithesis_instrumentation__.Notify(194081)

	log.Ops.Infof(ctx, "received valid cert bundle from trust leader")

	select {
	case t.finishedInit <- &bundle.Bundle:
		__antithesis_instrumentation__.Notify(194086)

	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(194087)
		log.Ops.Warningf(ctx, "context canceled while receiving bundle")
	}
}

func (t *tlsInitHandshaker) startServer(listener net.Listener) error {
	__antithesis_instrumentation__.Notify(194088)
	return t.server.ServeTLS(listener, "", "")
}

func (t *tlsInitHandshaker) stopServer() {
	__antithesis_instrumentation__.Notify(194089)

	_ = t.server.Shutdown(context.Background())
}

func (t *tlsInitHandshaker) getClientForTransport(transport *http.Transport) *http.Client {
	__antithesis_instrumentation__.Notify(194090)
	return &http.Client{
		Timeout:   1 * time.Second,
		Transport: transport,
	}
}

func (t *tlsInitHandshaker) getInsecureClient() *http.Client {
	__antithesis_instrumentation__.Notify(194091)

	clientTransport := http.DefaultTransport.(*http.Transport).Clone()
	clientTransport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return t.getClientForTransport(clientTransport)
}

func (t *tlsInitHandshaker) getClient(rootCAs *x509.CertPool) *http.Client {
	__antithesis_instrumentation__.Notify(194092)

	clientTransport := http.DefaultTransport.(*http.Transport).Clone()
	clientTransport.TLSClientConfig = &tls.Config{
		RootCAs: rootCAs,
	}
	return t.getClientForTransport(clientTransport)
}

func (t *tlsInitHandshaker) getPeerCACert(
	client *http.Client, peerAddress string, selfAddress string,
) (nodeHostnameAndCA, error) {
	__antithesis_instrumentation__.Notify(194093)
	challenge, err := createNodeHostnameAndCA(selfAddress, t.tempCerts.CACertificate, t.token)
	if err != nil {
		__antithesis_instrumentation__.Notify(194099)
		return nodeHostnameAndCA{}, err
	} else {
		__antithesis_instrumentation__.Notify(194100)
	}
	__antithesis_instrumentation__.Notify(194094)

	var body bytes.Buffer
	_ = json.NewEncoder(&body).Encode(challenge)
	res, err := client.Post(generateURLForClient(peerAddress, trustInitURL), "application/json; charset=utf-8", &body)
	if err != nil {
		__antithesis_instrumentation__.Notify(194101)
		return nodeHostnameAndCA{}, err
	} else {
		__antithesis_instrumentation__.Notify(194102)
	}
	__antithesis_instrumentation__.Notify(194095)
	defer res.Body.Close()

	if res.StatusCode != 200 {
		__antithesis_instrumentation__.Notify(194103)
		return nodeHostnameAndCA{}, errors.Errorf("unexpected error returned from peer: HTTP %d", res.StatusCode)
	} else {
		__antithesis_instrumentation__.Notify(194104)
	}
	__antithesis_instrumentation__.Notify(194096)

	var msg nodeHostnameAndCA
	if err := json.NewDecoder(res.Body).Decode(&msg); err != nil {
		__antithesis_instrumentation__.Notify(194105)
		return nodeHostnameAndCA{}, err
	} else {
		__antithesis_instrumentation__.Notify(194106)
	}
	__antithesis_instrumentation__.Notify(194097)
	defer res.Body.Close()

	if !msg.validHMAC(t.token) {
		__antithesis_instrumentation__.Notify(194107)
		return nodeHostnameAndCA{}, errInvalidHMAC
	} else {
		__antithesis_instrumentation__.Notify(194108)
	}
	__antithesis_instrumentation__.Notify(194098)
	return msg, nil
}

func (t *tlsInitHandshaker) runClient(
	ctx context.Context, peerHostname string, selfAddress string,
) {
	__antithesis_instrumentation__.Notify(194109)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	client := t.getInsecureClient()
	defer client.CloseIdleConnections()

	for {
		__antithesis_instrumentation__.Notify(194110)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(194114)
			return
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(194115)
		}
		__antithesis_instrumentation__.Notify(194111)

		peerHostnameAndCa, err := t.getPeerCACert(client, peerHostname, selfAddress)
		if err != nil {
			__antithesis_instrumentation__.Notify(194116)
			log.Ops.Warningf(ctx, "peer CA retrieval error: %v", err)

			select {
			case t.errors <- err:
				__antithesis_instrumentation__.Notify(194119)
			default:
				__antithesis_instrumentation__.Notify(194120)
			}
			__antithesis_instrumentation__.Notify(194117)
			if errors.Is(err, errInvalidHMAC) {
				__antithesis_instrumentation__.Notify(194121)
				return
			} else {
				__antithesis_instrumentation__.Notify(194122)
			}
			__antithesis_instrumentation__.Notify(194118)
			continue
		} else {
			__antithesis_instrumentation__.Notify(194123)
		}
		__antithesis_instrumentation__.Notify(194112)
		select {
		case t.trustedPeers <- peerHostnameAndCa:
			__antithesis_instrumentation__.Notify(194124)
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(194125)
		}
		__antithesis_instrumentation__.Notify(194113)
		return
	}
}

func (t *tlsInitHandshaker) sendBundle(
	ctx context.Context, address string, peerCACert []byte, caBundle nodeTrustBundle,
) (err error) {
	__antithesis_instrumentation__.Notify(194126)
	rootCAs, _ := x509.SystemCertPool()
	rootCAs.AppendCertsFromPEM(peerCACert)

	client := t.getClient(rootCAs)
	defer client.CloseIdleConnections()

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(caBundle); err != nil {
		__antithesis_instrumentation__.Notify(194129)
		return err
	} else {
		__antithesis_instrumentation__.Notify(194130)
	}
	__antithesis_instrumentation__.Notify(194127)

	every := log.Every(time.Second)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var lastError error
	for {
		__antithesis_instrumentation__.Notify(194131)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(194134)
			if lastError != nil {
				__antithesis_instrumentation__.Notify(194137)
				return lastError
			} else {
				__antithesis_instrumentation__.Notify(194138)
			}
			__antithesis_instrumentation__.Notify(194135)
			return errors.Errorf("context canceled before init bundle sent to %s", address)
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(194136)
		}
		__antithesis_instrumentation__.Notify(194132)

		res, err := client.Post(generateURLForClient(address, deliverBundleURL), "application/json; charset=utf-8", bytes.NewReader(body.Bytes()))
		if err == nil {
			__antithesis_instrumentation__.Notify(194139)
			res.Body.Close()
			break
		} else {
			__antithesis_instrumentation__.Notify(194140)
		}
		__antithesis_instrumentation__.Notify(194133)
		lastError = err
		if every.ShouldLog() {
			__antithesis_instrumentation__.Notify(194141)
			log.Ops.Warningf(ctx, "cannot send bundle: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(194142)
		}
	}
	__antithesis_instrumentation__.Notify(194128)

	return nil
}

func initHandshakeHelper(
	ctx context.Context,
	reporter func(string, ...interface{}),
	cfg *base.Config,
	token string,
	numExpectedNodes int,
	peers []string,
	certsDir string,
	listener net.Listener,
) error {
	__antithesis_instrumentation__.Notify(194143)
	if len(token) == 0 {
		__antithesis_instrumentation__.Notify(194154)
		return errors.AssertionFailedf("programming error: token cannot be empty")
	} else {
		__antithesis_instrumentation__.Notify(194155)
	}
	__antithesis_instrumentation__.Notify(194144)
	if numExpectedNodes <= 0 {
		__antithesis_instrumentation__.Notify(194156)
		return errors.AssertionFailedf("programming error: must expect more than 1 node")
	} else {
		__antithesis_instrumentation__.Notify(194157)
	}
	__antithesis_instrumentation__.Notify(194145)
	numExpectedPeers := numExpectedNodes - 1

	addr := listener.Addr()
	var listenHost string
	switch netAddr := addr.(type) {
	case *net.TCPAddr:
		__antithesis_instrumentation__.Notify(194158)
		listenHost = netAddr.IP.String()
	default:
		__antithesis_instrumentation__.Notify(194159)
		return errors.New("unsupported listener protocol: only TCP listeners supported")
	}
	__antithesis_instrumentation__.Notify(194146)
	tempCerts, err := createNodeInitTempCertificates(ctx, []string{listenHost}, defaultInitLifespan)
	if err != nil {
		__antithesis_instrumentation__.Notify(194160)
		return errors.Wrap(err, "failed to create certificates")
	} else {
		__antithesis_instrumentation__.Notify(194161)
	}
	__antithesis_instrumentation__.Notify(194147)

	log.Infof(ctx, "initializing temporary TLS handshake server, listen addr: %s", addr)
	handshaker := &tlsInitHandshaker{
		token:        []byte(token),
		certsDir:     certsDir,
		listenAddr:   addr.String(),
		tempCerts:    tempCerts,
		trustedPeers: make(chan nodeHostnameAndCA, numExpectedPeers),
		errors:       make(chan error, numExpectedPeers*2),
		finishedInit: make(chan *CertificateBundle, 1),
	}
	if err := handshaker.init(ctx); err != nil {
		__antithesis_instrumentation__.Notify(194162)
		return errors.Wrap(err, "error when initializing tls handshaker")
	} else {
		__antithesis_instrumentation__.Notify(194163)
	}
	__antithesis_instrumentation__.Notify(194148)

	defer handshaker.wg.Wait()

	peerCACerts := make(map[string]([]byte))

	if numExpectedPeers > 0 {
		__antithesis_instrumentation__.Notify(194164)
		handshaker.wg.Add(1)
		go func() {
			__antithesis_instrumentation__.Notify(194169)
			defer handshaker.wg.Done()

			log.Ops.Infof(ctx, "starting handshake server")
			defer log.Ops.Infof(ctx, "handshake server stopped")
			if err := handshaker.startServer(listener); !errors.Is(err, http.ErrServerClosed) {
				__antithesis_instrumentation__.Notify(194170)
				log.Ops.Errorf(ctx, "handshake server failed: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(194171)
			}
		}()
		__antithesis_instrumentation__.Notify(194165)

		defer handshaker.stopServer()

		for _, peerAddress := range peers {
			__antithesis_instrumentation__.Notify(194172)
			handshaker.wg.Add(1)
			go func(peerAddress string) {
				__antithesis_instrumentation__.Notify(194173)
				defer handshaker.wg.Done()

				peerCtx := logtags.AddTag(ctx, "peer", log.SafeOperational(peerAddress))
				log.Ops.Infof(peerCtx, "starting handshake client for peer")
				handshaker.runClient(peerCtx, peerAddress, addr.String())
			}(peerAddress)
		}
		__antithesis_instrumentation__.Notify(194166)

		if reporter != nil {
			__antithesis_instrumentation__.Notify(194174)
			reporter("waiting for handshake for %d peers", numExpectedPeers)
		} else {
			__antithesis_instrumentation__.Notify(194175)
		}
		__antithesis_instrumentation__.Notify(194167)

		for len(peerCACerts) < numExpectedPeers {
			__antithesis_instrumentation__.Notify(194176)
			select {
			case p := <-handshaker.trustedPeers:
				__antithesis_instrumentation__.Notify(194177)
				log.Ops.Infof(ctx, "received CA certificate for peer: %s", p.HostAddress)
				if reporter != nil {
					__antithesis_instrumentation__.Notify(194182)
					reporter("trusted peer: %s", p.HostAddress)
				} else {
					__antithesis_instrumentation__.Notify(194183)
				}
				__antithesis_instrumentation__.Notify(194178)
				peerCACerts[p.HostAddress] = p.CACertificate

			case err := <-handshaker.errors:
				__antithesis_instrumentation__.Notify(194179)
				if errors.Is(err, errInvalidHMAC) {
					__antithesis_instrumentation__.Notify(194184)

					log.Ops.Errorf(ctx, "HMAC error from client when connecting to peer: %v", err)
					return errors.New("invalid signature in messages from peers; likely due to token mismatch")
				} else {
					__antithesis_instrumentation__.Notify(194185)
				}
				__antithesis_instrumentation__.Notify(194180)
				log.Ops.Warningf(ctx, "error from client when connecting to peers (retrying): %s", err)

			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(194181)
				return errors.New("context canceled before all peers connected")
			}
		}
		__antithesis_instrumentation__.Notify(194168)
		log.Ops.Infof(ctx, "received response from all peers; choosing trust leader")
	} else {
		__antithesis_instrumentation__.Notify(194186)
	}
	__antithesis_instrumentation__.Notify(194149)

	trustLeader := true
	selfSignature, err := pemToSignature(tempCerts.CACertificate)
	if err != nil {
		__antithesis_instrumentation__.Notify(194187)
		return err
	} else {
		__antithesis_instrumentation__.Notify(194188)
	}
	__antithesis_instrumentation__.Notify(194150)

	for _, peerCertPEM := range peerCACerts {
		__antithesis_instrumentation__.Notify(194189)
		peerSignature, err := pemToSignature(peerCertPEM)
		if err != nil {
			__antithesis_instrumentation__.Notify(194191)
			return err
		} else {
			__antithesis_instrumentation__.Notify(194192)
		}
		__antithesis_instrumentation__.Notify(194190)
		if bytes.Compare(selfSignature, peerSignature) > 0 {
			__antithesis_instrumentation__.Notify(194193)
			trustLeader = false
		} else {
			__antithesis_instrumentation__.Notify(194194)
		}
	}
	__antithesis_instrumentation__.Notify(194151)

	if trustLeader {
		__antithesis_instrumentation__.Notify(194195)
		if reporter != nil {
			__antithesis_instrumentation__.Notify(194199)
			reporter("generating cert bundle for cluster")
		} else {
			__antithesis_instrumentation__.Notify(194200)
		}
		__antithesis_instrumentation__.Notify(194196)
		log.Ops.Infof(ctx, "we are trust leader; initializing certificate bundle")
		leaderCtx := logtags.AddTag(ctx, "trust-leader", nil)

		var b CertificateBundle

		if err := b.InitializeFromConfig(leaderCtx, *cfg); err != nil {
			__antithesis_instrumentation__.Notify(194201)
			return errors.Wrap(err, "error when creating initialization bundle")
		} else {
			__antithesis_instrumentation__.Notify(194202)
		}
		__antithesis_instrumentation__.Notify(194197)

		if numExpectedPeers > 0 {
			__antithesis_instrumentation__.Notify(194203)
			peerInit, err := collectLocalCABundle(cfg.SSLCertsDir)
			if err != nil {
				__antithesis_instrumentation__.Notify(194206)
				return errors.Wrap(err, "error when loading initialization bundle")
			} else {
				__antithesis_instrumentation__.Notify(194207)
			}
			__antithesis_instrumentation__.Notify(194204)

			trustBundle := nodeTrustBundle{Bundle: peerInit}
			trustBundle.signHMAC(handshaker.token)

			if reporter != nil {
				__antithesis_instrumentation__.Notify(194208)
				reporter("sending cert bundle to peers")
			} else {
				__antithesis_instrumentation__.Notify(194209)
			}
			__antithesis_instrumentation__.Notify(194205)

			for p := range peerCACerts {
				__antithesis_instrumentation__.Notify(194210)
				peerCtx := logtags.AddTag(leaderCtx, "peer", p)
				log.Ops.Infof(peerCtx, "delivering bundle to peer")
				if err := handshaker.sendBundle(peerCtx, p, peerCACerts[p], trustBundle); err != nil {
					__antithesis_instrumentation__.Notify(194211)

					return errors.Wrap(err, "error when sending bundle to peers as leader")
				} else {
					__antithesis_instrumentation__.Notify(194212)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(194213)
		}
		__antithesis_instrumentation__.Notify(194198)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(194214)
	}
	__antithesis_instrumentation__.Notify(194152)

	if reporter != nil {
		__antithesis_instrumentation__.Notify(194215)
		reporter("waiting for cert bundle")
	} else {
		__antithesis_instrumentation__.Notify(194216)
	}
	__antithesis_instrumentation__.Notify(194153)
	log.Ops.Infof(ctx, "we are not trust leader; now waiting for bundle from trust leader")

	select {
	case b := <-handshaker.finishedInit:
		__antithesis_instrumentation__.Notify(194217)
		if reporter != nil {
			__antithesis_instrumentation__.Notify(194221)
			reporter("received cert bundle")
		} else {
			__antithesis_instrumentation__.Notify(194222)
		}
		__antithesis_instrumentation__.Notify(194218)

		if b == nil {
			__antithesis_instrumentation__.Notify(194223)
			return errors.New("expected non-nil init bundle to be received from trust leader")
		} else {
			__antithesis_instrumentation__.Notify(194224)
		}
		__antithesis_instrumentation__.Notify(194219)
		log.Ops.Infof(ctx, "received bundle, now initializing node certificate files")
		return b.InitializeNodeFromBundle(ctx, *cfg)

	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(194220)
		return errors.New("context canceled before init bundle received from leader")
	}
}

func InitHandshake(
	ctx context.Context,
	reporter func(string, ...interface{}),
	cfg *base.Config,
	token string,
	numExpectedNodes int,
	peers []string,
	certsDir string,
	listener net.Listener,
) error {
	__antithesis_instrumentation__.Notify(194225)

	return contextutil.RunWithTimeout(ctx, "init handshake", defaultInitLifespan, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(194226)
		ctx = logtags.AddTag(ctx, "init-tls-handshake", nil)
		return initHandshakeHelper(ctx, reporter, cfg, token, numExpectedNodes, peers, certsDir, listener)
	})
}
