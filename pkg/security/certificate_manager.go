package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
)

var (
	metaCAExpiration = metric.Metadata{
		Name:        "security.certificate.expiration.ca",
		Help:        "Expiration for the CA certificate. 0 means no certificate or error.",
		Measurement: "Certificate Expiration",
		Unit:        metric.Unit_TIMESTAMP_SEC,
	}
	metaClientCAExpiration = metric.Metadata{
		Name:        "security.certificate.expiration.client-ca",
		Help:        "Expiration for the client CA certificate. 0 means no certificate or error.",
		Measurement: "Certificate Expiration",
		Unit:        metric.Unit_TIMESTAMP_SEC,
	}
	metaUICAExpiration = metric.Metadata{
		Name:        "security.certificate.expiration.ui-ca",
		Help:        "Expiration for the UI CA certificate. 0 means no certificate or error.",
		Measurement: "Certificate Expiration",
		Unit:        metric.Unit_TIMESTAMP_SEC,
	}
	metaNodeExpiration = metric.Metadata{
		Name:        "security.certificate.expiration.node",
		Help:        "Expiration for the node certificate. 0 means no certificate or error.",
		Measurement: "Certificate Expiration",
		Unit:        metric.Unit_TIMESTAMP_SEC,
	}
	metaNodeClientExpiration = metric.Metadata{
		Name:        "security.certificate.expiration.node-client",
		Help:        "Expiration for the node's client certificate. 0 means no certificate or error.",
		Measurement: "Certificate Expiration",
		Unit:        metric.Unit_TIMESTAMP_SEC,
	}
	metaUIExpiration = metric.Metadata{
		Name:        "security.certificate.expiration.ui",
		Help:        "Expiration for the UI certificate. 0 means no certificate or error.",
		Measurement: "Certificate Expiration",
		Unit:        metric.Unit_TIMESTAMP_SEC,
	}

	metaTenantCAExpiration = metric.Metadata{
		Name:        "security.certificate.expiration.ca-client-tenant",
		Help:        "Expiration for the Tenant Client CA certificate. 0 means no certificate or error.",
		Measurement: "Certificate Expiration",
		Unit:        metric.Unit_TIMESTAMP_SEC,
	}
	metaTenantExpiration = metric.Metadata{
		Name:        "security.certificate.expiration.client-tenant",
		Help:        "Expiration for the Tenant Client certificate. 0 means no certificate or error.",
		Measurement: "Certificate Expiration",
		Unit:        metric.Unit_TIMESTAMP_SEC,
	}
)

type CertificateManager struct {
	tenantIdentifier uint64
	CertsLocator

	tlsSettings TLSSettings

	certMetrics CertificateMetrics

	mu syncutil.RWMutex

	initialized bool

	caCert         *CertInfo
	clientCACert   *CertInfo
	uiCACert       *CertInfo
	nodeCert       *CertInfo
	nodeClientCert *CertInfo
	uiCert         *CertInfo
	clientCerts    map[SQLUsername]*CertInfo

	tenantCACert, tenantCert, tenantSigningCert *CertInfo

	serverConfig *tls.Config

	uiServerConfig *tls.Config

	clientConfig *tls.Config

	tenantConfig *tls.Config
}

type CertificateMetrics struct {
	CAExpiration         *metric.Gauge
	ClientCAExpiration   *metric.Gauge
	UICAExpiration       *metric.Gauge
	NodeExpiration       *metric.Gauge
	NodeClientExpiration *metric.Gauge
	UIExpiration         *metric.Gauge
	TenantCAExpiration   *metric.Gauge
	TenantExpiration     *metric.Gauge
}

func makeCertificateManager(
	certsDir string, tlsSettings TLSSettings, opts ...Option,
) *CertificateManager {
	__antithesis_instrumentation__.Notify(185993)
	var o cmOptions
	for _, fn := range opts {
		__antithesis_instrumentation__.Notify(185995)
		fn(&o)
	}
	__antithesis_instrumentation__.Notify(185994)

	return &CertificateManager{
		CertsLocator:     MakeCertsLocator(certsDir),
		tenantIdentifier: o.tenantIdentifier,
		tlsSettings:      tlsSettings,
		certMetrics: CertificateMetrics{
			CAExpiration:         metric.NewGauge(metaCAExpiration),
			ClientCAExpiration:   metric.NewGauge(metaClientCAExpiration),
			UICAExpiration:       metric.NewGauge(metaUICAExpiration),
			NodeExpiration:       metric.NewGauge(metaNodeExpiration),
			NodeClientExpiration: metric.NewGauge(metaNodeClientExpiration),
			UIExpiration:         metric.NewGauge(metaUIExpiration),
			TenantCAExpiration:   metric.NewGauge(metaTenantCAExpiration),
			TenantExpiration:     metric.NewGauge(metaTenantExpiration),
		},
	}
}

type cmOptions struct {
	tenantIdentifier uint64
}

type Option func(*cmOptions)

func ForTenant(tenantIdentifier uint64) Option {
	__antithesis_instrumentation__.Notify(185996)
	return func(opts *cmOptions) {
		__antithesis_instrumentation__.Notify(185997)
		opts.tenantIdentifier = tenantIdentifier
	}
}

func NewCertificateManager(
	certsDir string, tlsSettings TLSSettings, opts ...Option,
) (*CertificateManager, error) {
	__antithesis_instrumentation__.Notify(185998)
	cm := makeCertificateManager(certsDir, tlsSettings, opts...)
	return cm, cm.LoadCertificates()
}

func NewCertificateManagerFirstRun(
	certsDir string, tlsSettings TLSSettings, opts ...Option,
) (*CertificateManager, error) {
	__antithesis_instrumentation__.Notify(185999)
	cm := makeCertificateManager(certsDir, tlsSettings, opts...)
	if err := NewCertificateLoader(cm.certsDir).MaybeCreateCertsDir(); err != nil {
		__antithesis_instrumentation__.Notify(186001)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186002)
	}
	__antithesis_instrumentation__.Notify(186000)

	return cm, cm.LoadCertificates()
}

func (cm *CertificateManager) IsForTenant() bool {
	__antithesis_instrumentation__.Notify(186003)
	return cm.tenantIdentifier != 0
}

func (cm *CertificateManager) Metrics() CertificateMetrics {
	__antithesis_instrumentation__.Notify(186004)
	return cm.certMetrics
}

func (cm *CertificateManager) RegisterSignalHandler(stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(186005)
	ctx := context.Background()
	go func() {
		__antithesis_instrumentation__.Notify(186006)
		ch := sysutil.RefreshSignaledChan()
		for {
			__antithesis_instrumentation__.Notify(186007)
			select {
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(186008)
				return
			case sig := <-ch:
				__antithesis_instrumentation__.Notify(186009)
				log.Ops.Infof(ctx, "received signal %q, triggering certificate reload", sig)
				if err := cm.LoadCertificates(); err != nil {
					__antithesis_instrumentation__.Notify(186010)
					log.Ops.Warningf(ctx, "could not reload certificates: %v", err)
					log.StructuredEvent(ctx, &eventpb.CertsReload{Success: false, ErrorMessage: err.Error()})
				} else {
					__antithesis_instrumentation__.Notify(186011)
					log.StructuredEvent(ctx, &eventpb.CertsReload{Success: true})
				}
			}
		}
	}()
}

type CertsLocator struct {
	certsDir string
}

func MakeCertsLocator(certsDir string) CertsLocator {
	__antithesis_instrumentation__.Notify(186012)
	return CertsLocator{certsDir: os.ExpandEnv(certsDir)}
}

func (cl CertsLocator) CACertPath() string {
	__antithesis_instrumentation__.Notify(186013)
	return filepath.Join(cl.certsDir, CACertFilename())
}

func (cl CertsLocator) FullPath(ci *CertInfo) string {
	__antithesis_instrumentation__.Notify(186014)
	return filepath.Join(cl.certsDir, ci.Filename)
}

func (cl CertsLocator) EnsureCertsDirectory() error {
	__antithesis_instrumentation__.Notify(186015)
	return os.MkdirAll(cl.certsDir, 0700)
}

func CACertFilename() string {
	__antithesis_instrumentation__.Notify(186016)
	return "ca" + certExtension
}

func (cl CertsLocator) CAKeyPath() string {
	__antithesis_instrumentation__.Notify(186017)
	return filepath.Join(cl.certsDir, CAKeyFilename())
}

func CAKeyFilename() string {
	__antithesis_instrumentation__.Notify(186018)
	return "ca" + keyExtension
}

func (cl CertsLocator) TenantCACertPath() string {
	__antithesis_instrumentation__.Notify(186019)
	return filepath.Join(cl.certsDir, TenantCACertFilename())
}

func TenantCACertFilename() string {
	__antithesis_instrumentation__.Notify(186020)
	return "ca-client-tenant" + certExtension
}

func (cl CertsLocator) ClientCACertPath() string {
	__antithesis_instrumentation__.Notify(186021)
	return filepath.Join(cl.certsDir, "ca-client"+certExtension)
}

func (cl CertsLocator) ClientCAKeyPath() string {
	__antithesis_instrumentation__.Notify(186022)
	return filepath.Join(cl.certsDir, "ca-client"+keyExtension)
}

func (cl CertsLocator) ClientNodeCertPath() string {
	__antithesis_instrumentation__.Notify(186023)
	return filepath.Join(cl.certsDir, "client.node"+certExtension)
}

func (cl CertsLocator) ClientNodeKeyPath() string {
	__antithesis_instrumentation__.Notify(186024)
	return filepath.Join(cl.certsDir, "client.node"+keyExtension)
}

func (cl CertsLocator) UICACertPath() string {
	__antithesis_instrumentation__.Notify(186025)
	return filepath.Join(cl.certsDir, "ca-ui"+certExtension)
}

func (cl CertsLocator) UICAKeyPath() string {
	__antithesis_instrumentation__.Notify(186026)
	return filepath.Join(cl.certsDir, "ca-ui"+keyExtension)
}

func (cl CertsLocator) NodeCertPath() string {
	__antithesis_instrumentation__.Notify(186027)
	return filepath.Join(cl.certsDir, NodeCertFilename())
}

func (cl CertsLocator) HasNodeCert() (bool, error) {
	__antithesis_instrumentation__.Notify(186028)
	_, err := os.Stat(cl.NodeCertPath())
	if err != nil {
		__antithesis_instrumentation__.Notify(186030)
		if oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(186032)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(186033)
		}
		__antithesis_instrumentation__.Notify(186031)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(186034)
	}
	__antithesis_instrumentation__.Notify(186029)
	return true, nil
}

func NodeCertFilename() string {
	__antithesis_instrumentation__.Notify(186035)
	return "node" + certExtension
}

func (cl CertsLocator) NodeKeyPath() string {
	__antithesis_instrumentation__.Notify(186036)
	return filepath.Join(cl.certsDir, NodeKeyFilename())
}

func NodeKeyFilename() string {
	__antithesis_instrumentation__.Notify(186037)
	return "node" + keyExtension
}

func (cl CertsLocator) UICertPath() string {
	__antithesis_instrumentation__.Notify(186038)
	return filepath.Join(cl.certsDir, "ui"+certExtension)
}

func (cl CertsLocator) UIKeyPath() string {
	__antithesis_instrumentation__.Notify(186039)
	return filepath.Join(cl.certsDir, "ui"+keyExtension)
}

func (cl CertsLocator) TenantCertPath(tenantIdentifier string) string {
	__antithesis_instrumentation__.Notify(186040)
	return filepath.Join(cl.certsDir, TenantCertFilename(tenantIdentifier))
}

func TenantCertFilename(tenantIdentifier string) string {
	__antithesis_instrumentation__.Notify(186041)
	return "client-tenant." + tenantIdentifier + certExtension
}

func (cl CertsLocator) TenantKeyPath(tenantIdentifier string) string {
	__antithesis_instrumentation__.Notify(186042)
	return filepath.Join(cl.certsDir, TenantKeyFilename(tenantIdentifier))
}

func TenantKeyFilename(tenantIdentifier string) string {
	__antithesis_instrumentation__.Notify(186043)
	return "client-tenant." + tenantIdentifier + keyExtension
}

func (cl CertsLocator) TenantSigningCertPath(tenantIdentifier string) string {
	__antithesis_instrumentation__.Notify(186044)
	return filepath.Join(cl.certsDir, TenantSigningCertFilename(tenantIdentifier))
}

func TenantSigningCertFilename(tenantIdentifier string) string {
	__antithesis_instrumentation__.Notify(186045)
	return "tenant-signing." + tenantIdentifier + certExtension
}

func (cl CertsLocator) TenantSigningKeyPath(tenantIdentifier string) string {
	__antithesis_instrumentation__.Notify(186046)
	return filepath.Join(cl.certsDir, TenantSigningKeyFilename(tenantIdentifier))
}

func TenantSigningKeyFilename(tenantIdentifier string) string {
	__antithesis_instrumentation__.Notify(186047)
	return "tenant-signing." + tenantIdentifier + keyExtension
}

func (cl CertsLocator) ClientCertPath(user SQLUsername) string {
	__antithesis_instrumentation__.Notify(186048)
	return filepath.Join(cl.certsDir, ClientCertFilename(user))
}

func ClientCertFilename(user SQLUsername) string {
	__antithesis_instrumentation__.Notify(186049)
	return "client." + user.Normalized() + certExtension
}

func (cl CertsLocator) ClientKeyPath(user SQLUsername) string {
	__antithesis_instrumentation__.Notify(186050)
	return filepath.Join(cl.certsDir, ClientKeyFilename(user))
}

func ClientKeyFilename(user SQLUsername) string {
	__antithesis_instrumentation__.Notify(186051)
	return "client." + user.Normalized() + keyExtension
}

func (cl CertsLocator) SQLServiceCertPath() string {
	__antithesis_instrumentation__.Notify(186052)
	return filepath.Join(cl.certsDir, SQLServiceCertFilename())
}

func SQLServiceCertFilename() string {
	__antithesis_instrumentation__.Notify(186053)
	return "service.sql" + certExtension
}

func (cl CertsLocator) SQLServiceKeyPath() string {
	__antithesis_instrumentation__.Notify(186054)
	return filepath.Join(cl.certsDir, SQLServiceKeyFilename())
}

func SQLServiceKeyFilename() string {
	__antithesis_instrumentation__.Notify(186055)
	return "service.sql" + keyExtension
}

func (cl CertsLocator) SQLServiceCACertPath() string {
	__antithesis_instrumentation__.Notify(186056)
	return filepath.Join(cl.certsDir, SQLServiceCACertFilename())
}

func SQLServiceCACertFilename() string {
	__antithesis_instrumentation__.Notify(186057)
	return "service.ca.sql" + certExtension
}

func (cl CertsLocator) SQLServiceCAKeyPath() string {
	__antithesis_instrumentation__.Notify(186058)
	return filepath.Join(cl.certsDir, SQLServiceCAKeyFilename())
}

func SQLServiceCAKeyFilename() string {
	__antithesis_instrumentation__.Notify(186059)
	return "service.ca.sql" + keyExtension
}

func (cl CertsLocator) RPCServiceCertPath() string {
	__antithesis_instrumentation__.Notify(186060)
	return filepath.Join(cl.certsDir, RPCServiceCertFilename())
}

func RPCServiceCertFilename() string {
	__antithesis_instrumentation__.Notify(186061)
	return "service.rpc" + certExtension
}

func (cl CertsLocator) RPCServiceKeyPath() string {
	__antithesis_instrumentation__.Notify(186062)
	return filepath.Join(cl.certsDir, RPCServiceKeyFilename())
}

func RPCServiceKeyFilename() string {
	__antithesis_instrumentation__.Notify(186063)
	return "service.rpc" + keyExtension
}

func (cl CertsLocator) RPCServiceCACertPath() string {
	__antithesis_instrumentation__.Notify(186064)
	return filepath.Join(cl.certsDir, RPCServiceCACertFilename())
}

func RPCServiceCACertFilename() string {
	__antithesis_instrumentation__.Notify(186065)
	return "service.ca.rpc" + certExtension
}

func (cl CertsLocator) RPCServiceCAKeyPath() string {
	__antithesis_instrumentation__.Notify(186066)
	return filepath.Join(cl.certsDir, RPCServiceCAKeyFilename())
}

func RPCServiceCAKeyFilename() string {
	__antithesis_instrumentation__.Notify(186067)
	return "service.ca.rpc" + keyExtension
}

func (cm *CertificateManager) CACert() *CertInfo {
	__antithesis_instrumentation__.Notify(186068)
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.caCert
}

func (cm *CertificateManager) ClientCACert() *CertInfo {
	__antithesis_instrumentation__.Notify(186069)
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clientCACert
}

func (cm *CertificateManager) UICACert() *CertInfo {
	__antithesis_instrumentation__.Notify(186070)
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.uiCACert
}

func (cm *CertificateManager) UICert() *CertInfo {
	__antithesis_instrumentation__.Notify(186071)
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.uiCert
}

func checkCertIsValid(cert *CertInfo) error {
	__antithesis_instrumentation__.Notify(186072)
	if cert == nil {
		__antithesis_instrumentation__.Notify(186074)
		return errors.New("not found")
	} else {
		__antithesis_instrumentation__.Notify(186075)
	}
	__antithesis_instrumentation__.Notify(186073)
	return cert.Error
}

func (cm *CertificateManager) NodeCert() *CertInfo {
	__antithesis_instrumentation__.Notify(186076)
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.nodeCert
}

func (cm *CertificateManager) ClientCerts() map[SQLUsername]*CertInfo {
	__antithesis_instrumentation__.Notify(186077)
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clientCerts
}

type Error struct {
	Message string
	Err     error
}

func (e *Error) Error() string {
	__antithesis_instrumentation__.Notify(186078)
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func makeErrorf(err error, format string, args ...interface{}) *Error {
	__antithesis_instrumentation__.Notify(186079)
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

func makeError(err error, s string) *Error {
	__antithesis_instrumentation__.Notify(186080)
	return makeErrorf(err, "%s", s)
}

func (cm *CertificateManager) LoadCertificates() error {
	__antithesis_instrumentation__.Notify(186081)
	cl := NewCertificateLoader(cm.certsDir)
	if err := cl.Load(); err != nil {
		__antithesis_instrumentation__.Notify(186087)
		return makeErrorf(err, "problem loading certs directory %s", cm.certsDir)
	} else {
		__antithesis_instrumentation__.Notify(186088)
	}
	__antithesis_instrumentation__.Notify(186082)

	var caCert, clientCACert, uiCACert, nodeCert, uiCert, nodeClientCert *CertInfo
	var tenantCACert, tenantCert, tenantSigningCert *CertInfo
	clientCerts := make(map[SQLUsername]*CertInfo)
	for _, ci := range cl.Certificates() {
		__antithesis_instrumentation__.Notify(186089)
		switch ci.FileUsage {
		case CAPem:
			__antithesis_instrumentation__.Notify(186090)
			caCert = ci
		case ClientCAPem:
			__antithesis_instrumentation__.Notify(186091)
			clientCACert = ci
		case UICAPem:
			__antithesis_instrumentation__.Notify(186092)
			uiCACert = ci
		case NodePem:
			__antithesis_instrumentation__.Notify(186093)
			nodeCert = ci
		case TenantPem:
			__antithesis_instrumentation__.Notify(186094)

			tenantID, err := strconv.ParseUint(ci.Name, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(186102)
				return errors.Errorf("invalid tenant id %s", ci.Name)
			} else {
				__antithesis_instrumentation__.Notify(186103)
			}
			__antithesis_instrumentation__.Notify(186095)
			if tenantID == cm.tenantIdentifier {
				__antithesis_instrumentation__.Notify(186104)
				tenantCert = ci
			} else {
				__antithesis_instrumentation__.Notify(186105)
			}
		case TenantSigningPem:
			__antithesis_instrumentation__.Notify(186096)

			tenantID, err := strconv.ParseUint(ci.Name, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(186106)
				return errors.Errorf("invalid tenant id %s", ci.Name)
			} else {
				__antithesis_instrumentation__.Notify(186107)
			}
			__antithesis_instrumentation__.Notify(186097)
			if tenantID == cm.tenantIdentifier {
				__antithesis_instrumentation__.Notify(186108)
				tenantSigningCert = ci
			} else {
				__antithesis_instrumentation__.Notify(186109)
			}
		case TenantCAPem:
			__antithesis_instrumentation__.Notify(186098)
			tenantCACert = ci
		case UIPem:
			__antithesis_instrumentation__.Notify(186099)
			uiCert = ci
		case ClientPem:
			__antithesis_instrumentation__.Notify(186100)
			username := MakeSQLUsernameFromPreNormalizedString(ci.Name)
			clientCerts[username] = ci
			if username.IsNodeUser() {
				__antithesis_instrumentation__.Notify(186110)
				nodeClientCert = ci
			} else {
				__antithesis_instrumentation__.Notify(186111)
			}
		default:
			__antithesis_instrumentation__.Notify(186101)
			return errors.Errorf("unsupported certificate %v", ci.Filename)
		}
	}
	__antithesis_instrumentation__.Notify(186083)

	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.initialized {
		__antithesis_instrumentation__.Notify(186112)

		if err := checkCertIsValid(caCert); checkCertIsValid(cm.caCert) == nil && func() bool {
			__antithesis_instrumentation__.Notify(186120)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(186121)
			return makeError(err, "reload would lose valid CA cert")
		} else {
			__antithesis_instrumentation__.Notify(186122)
		}
		__antithesis_instrumentation__.Notify(186113)
		if err := checkCertIsValid(nodeCert); checkCertIsValid(cm.nodeCert) == nil && func() bool {
			__antithesis_instrumentation__.Notify(186123)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(186124)
			return makeError(err, "reload would lose valid node cert")
		} else {
			__antithesis_instrumentation__.Notify(186125)
		}
		__antithesis_instrumentation__.Notify(186114)
		if err := checkCertIsValid(nodeClientCert); checkCertIsValid(cm.nodeClientCert) == nil && func() bool {
			__antithesis_instrumentation__.Notify(186126)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(186127)
			return makeErrorf(err, "reload would lose valid client cert for '%s'", NodeUser)
		} else {
			__antithesis_instrumentation__.Notify(186128)
		}
		__antithesis_instrumentation__.Notify(186115)
		if err := checkCertIsValid(clientCACert); checkCertIsValid(cm.clientCACert) == nil && func() bool {
			__antithesis_instrumentation__.Notify(186129)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(186130)
			return makeError(err, "reload would lose valid CA certificate for client verification")
		} else {
			__antithesis_instrumentation__.Notify(186131)
		}
		__antithesis_instrumentation__.Notify(186116)
		if err := checkCertIsValid(uiCACert); checkCertIsValid(cm.uiCACert) == nil && func() bool {
			__antithesis_instrumentation__.Notify(186132)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(186133)
			return makeError(err, "reload would lose valid CA certificate for UI")
		} else {
			__antithesis_instrumentation__.Notify(186134)
		}
		__antithesis_instrumentation__.Notify(186117)
		if err := checkCertIsValid(uiCert); checkCertIsValid(cm.uiCert) == nil && func() bool {
			__antithesis_instrumentation__.Notify(186135)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(186136)
			return makeError(err, "reload would lose valid UI certificate")
		} else {
			__antithesis_instrumentation__.Notify(186137)
		}
		__antithesis_instrumentation__.Notify(186118)

		if err := checkCertIsValid(tenantCACert); checkCertIsValid(cm.tenantCACert) == nil && func() bool {
			__antithesis_instrumentation__.Notify(186138)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(186139)
			return makeError(err, "reload would lose valid tenant client CA certificate")
		} else {
			__antithesis_instrumentation__.Notify(186140)
		}
		__antithesis_instrumentation__.Notify(186119)
		if err := checkCertIsValid(tenantCert); checkCertIsValid(cm.tenantCert) == nil && func() bool {
			__antithesis_instrumentation__.Notify(186141)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(186142)
			return makeError(err, "reload would lose valid tenant client certificate")
		} else {
			__antithesis_instrumentation__.Notify(186143)
		}
	} else {
		__antithesis_instrumentation__.Notify(186144)
	}
	__antithesis_instrumentation__.Notify(186084)

	if tenantCert == nil && func() bool {
		__antithesis_instrumentation__.Notify(186145)
		return cm.tenantIdentifier != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(186146)
		return makeErrorf(errors.New("tenant client cert not found"), "for %d in %s", cm.tenantIdentifier, cm.certsDir)
	} else {
		__antithesis_instrumentation__.Notify(186147)
	}
	__antithesis_instrumentation__.Notify(186085)

	if nodeClientCert == nil && func() bool {
		__antithesis_instrumentation__.Notify(186148)
		return nodeCert != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(186149)

		if err := validateDualPurposeNodeCert(nodeCert); err != nil {
			__antithesis_instrumentation__.Notify(186150)
			return err
		} else {
			__antithesis_instrumentation__.Notify(186151)
		}
	} else {
		__antithesis_instrumentation__.Notify(186152)
	}
	__antithesis_instrumentation__.Notify(186086)

	cm.caCert = caCert
	cm.clientCACert = clientCACert
	cm.uiCACert = uiCACert

	cm.nodeCert = nodeCert
	cm.nodeClientCert = nodeClientCert
	cm.uiCert = uiCert
	cm.clientCerts = clientCerts

	cm.initialized = true

	cm.serverConfig = nil
	cm.uiServerConfig = nil
	cm.clientConfig = nil

	cm.tenantConfig = nil
	cm.tenantCACert = tenantCACert
	cm.tenantCert = tenantCert
	cm.tenantSigningCert = tenantSigningCert

	cm.updateMetricsLocked()
	return nil
}

func (cm *CertificateManager) updateMetricsLocked() {
	__antithesis_instrumentation__.Notify(186153)
	maybeSetMetric := func(m *metric.Gauge, ci *CertInfo) {
		__antithesis_instrumentation__.Notify(186155)
		if m == nil {
			__antithesis_instrumentation__.Notify(186157)
			return
		} else {
			__antithesis_instrumentation__.Notify(186158)
		}
		__antithesis_instrumentation__.Notify(186156)
		if ci != nil && func() bool {
			__antithesis_instrumentation__.Notify(186159)
			return ci.Error == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(186160)
			m.Update(ci.ExpirationTime.Unix())
		} else {
			__antithesis_instrumentation__.Notify(186161)
			m.Update(0)
		}
	}
	__antithesis_instrumentation__.Notify(186154)

	maybeSetMetric(cm.certMetrics.CAExpiration, cm.caCert)

	maybeSetMetric(cm.certMetrics.ClientCAExpiration, cm.clientCACert)

	maybeSetMetric(cm.certMetrics.UICAExpiration, cm.uiCACert)

	maybeSetMetric(cm.certMetrics.NodeExpiration, cm.nodeCert)

	maybeSetMetric(cm.certMetrics.NodeClientExpiration, cm.nodeClientCert)

	maybeSetMetric(cm.certMetrics.UIExpiration, cm.uiCert)
}

func (cm *CertificateManager) GetServerTLSConfig() (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(186162)
	if _, err := cm.getEmbeddedServerTLSConfig(nil); err != nil {
		__antithesis_instrumentation__.Notify(186164)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186165)
	}
	__antithesis_instrumentation__.Notify(186163)
	return &tls.Config{
		GetConfigForClient: cm.getEmbeddedServerTLSConfig,

		GetCertificate: func(hi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			__antithesis_instrumentation__.Notify(186166)
			return nil, nil
		},
	}, nil
}

func (cm *CertificateManager) getEmbeddedServerTLSConfig(
	_ *tls.ClientHelloInfo,
) (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(186167)
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.serverConfig != nil {
		__antithesis_instrumentation__.Notify(186174)
		return cm.serverConfig, nil
	} else {
		__antithesis_instrumentation__.Notify(186175)
	}
	__antithesis_instrumentation__.Notify(186168)

	ca, err := cm.getCACertLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(186176)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186177)
	}
	__antithesis_instrumentation__.Notify(186169)

	var nodeCert *CertInfo
	if !cm.IsForTenant() {
		__antithesis_instrumentation__.Notify(186178)

		nodeCert, err = cm.getNodeCertLocked()
		if err != nil {
			__antithesis_instrumentation__.Notify(186179)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(186180)
		}
	} else {
		__antithesis_instrumentation__.Notify(186181)

		nodeCert, err = cm.getTenantCertLocked()
		if err != nil {
			__antithesis_instrumentation__.Notify(186182)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(186183)
		}
	}
	__antithesis_instrumentation__.Notify(186170)

	clientCA, err := cm.getClientCACertLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(186184)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186185)
	}
	__antithesis_instrumentation__.Notify(186171)

	tenantCA, err := cm.getTenantCACertLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(186186)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186187)
	}
	__antithesis_instrumentation__.Notify(186172)

	cfg, err := newServerTLSConfig(
		cm.tlsSettings,
		nodeCert.FileContents,
		nodeCert.KeyFileContents,
		ca.FileContents,
		clientCA.FileContents,
		tenantCA.FileContents,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(186188)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186189)
	}
	__antithesis_instrumentation__.Notify(186173)

	cm.serverConfig = cfg
	return cfg, nil
}

func (cm *CertificateManager) GetUIServerTLSConfig() (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(186190)
	if _, err := cm.getEmbeddedUIServerTLSConfig(nil); err != nil {
		__antithesis_instrumentation__.Notify(186192)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186193)
	}
	__antithesis_instrumentation__.Notify(186191)
	return &tls.Config{
		GetConfigForClient: cm.getEmbeddedUIServerTLSConfig,
	}, nil
}

func (cm *CertificateManager) getEmbeddedUIServerTLSConfig(
	_ *tls.ClientHelloInfo,
) (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(186194)
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.uiServerConfig != nil {
		__antithesis_instrumentation__.Notify(186198)
		return cm.uiServerConfig, nil
	} else {
		__antithesis_instrumentation__.Notify(186199)
	}
	__antithesis_instrumentation__.Notify(186195)

	uiCert, err := cm.getUICertLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(186200)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186201)
	}
	__antithesis_instrumentation__.Notify(186196)

	cfg, err := newUIServerTLSConfig(
		cm.tlsSettings,
		uiCert.FileContents,
		uiCert.KeyFileContents)
	if err != nil {
		__antithesis_instrumentation__.Notify(186202)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186203)
	}
	__antithesis_instrumentation__.Notify(186197)

	cm.uiServerConfig = cfg
	return cfg, nil
}

func (cm *CertificateManager) getCACertLocked() (*CertInfo, error) {
	__antithesis_instrumentation__.Notify(186204)
	if err := checkCertIsValid(cm.caCert); err != nil {
		__antithesis_instrumentation__.Notify(186206)
		return nil, makeError(err, "problem with CA certificate")
	} else {
		__antithesis_instrumentation__.Notify(186207)
	}
	__antithesis_instrumentation__.Notify(186205)
	return cm.caCert, nil
}

func (cm *CertificateManager) getClientCACertLocked() (*CertInfo, error) {
	__antithesis_instrumentation__.Notify(186208)
	if cm.clientCACert == nil {
		__antithesis_instrumentation__.Notify(186211)

		return cm.getCACertLocked()
	} else {
		__antithesis_instrumentation__.Notify(186212)
	}
	__antithesis_instrumentation__.Notify(186209)

	if err := checkCertIsValid(cm.clientCACert); err != nil {
		__antithesis_instrumentation__.Notify(186213)
		return nil, makeError(err, "problem with client CA certificate")
	} else {
		__antithesis_instrumentation__.Notify(186214)
	}
	__antithesis_instrumentation__.Notify(186210)
	return cm.clientCACert, nil
}

func (cm *CertificateManager) getUICACertLocked() (*CertInfo, error) {
	__antithesis_instrumentation__.Notify(186215)
	if cm.uiCACert == nil {
		__antithesis_instrumentation__.Notify(186218)

		return cm.getCACertLocked()
	} else {
		__antithesis_instrumentation__.Notify(186219)
	}
	__antithesis_instrumentation__.Notify(186216)

	if err := checkCertIsValid(cm.uiCACert); err != nil {
		__antithesis_instrumentation__.Notify(186220)
		return nil, makeError(err, "problem with UI CA certificate")
	} else {
		__antithesis_instrumentation__.Notify(186221)
	}
	__antithesis_instrumentation__.Notify(186217)
	return cm.uiCACert, nil
}

func (cm *CertificateManager) getNodeCertLocked() (*CertInfo, error) {
	__antithesis_instrumentation__.Notify(186222)
	if err := checkCertIsValid(cm.nodeCert); err != nil {
		__antithesis_instrumentation__.Notify(186224)
		return nil, makeError(err, "problem with node certificate")
	} else {
		__antithesis_instrumentation__.Notify(186225)
	}
	__antithesis_instrumentation__.Notify(186223)
	return cm.nodeCert, nil
}

func (cm *CertificateManager) getUICertLocked() (*CertInfo, error) {
	__antithesis_instrumentation__.Notify(186226)
	if cm.uiCert == nil {
		__antithesis_instrumentation__.Notify(186229)

		if !cm.IsForTenant() {
			__antithesis_instrumentation__.Notify(186231)

			return cm.getNodeCertLocked()
		} else {
			__antithesis_instrumentation__.Notify(186232)
		}
		__antithesis_instrumentation__.Notify(186230)

		return cm.getTenantCertLocked()
	} else {
		__antithesis_instrumentation__.Notify(186233)
	}
	__antithesis_instrumentation__.Notify(186227)
	if err := checkCertIsValid(cm.uiCert); err != nil {
		__antithesis_instrumentation__.Notify(186234)
		return nil, makeError(err, "problem with UI certificate")
	} else {
		__antithesis_instrumentation__.Notify(186235)
	}
	__antithesis_instrumentation__.Notify(186228)
	return cm.uiCert, nil
}

func (cm *CertificateManager) getClientCertLocked(user SQLUsername) (*CertInfo, error) {
	__antithesis_instrumentation__.Notify(186236)
	ci := cm.clientCerts[user]
	if err := checkCertIsValid(ci); err != nil {
		__antithesis_instrumentation__.Notify(186238)
		return nil, makeErrorf(err, "problem with client cert for user %s", user)
	} else {
		__antithesis_instrumentation__.Notify(186239)
	}
	__antithesis_instrumentation__.Notify(186237)

	return ci, nil
}

func (cm *CertificateManager) getNodeClientCertLocked() (*CertInfo, error) {
	__antithesis_instrumentation__.Notify(186240)
	if cm.nodeClientCert == nil {
		__antithesis_instrumentation__.Notify(186243)

		if cm.IsForTenant() {
			__antithesis_instrumentation__.Notify(186245)
			return nil, errors.New("no node client cert for a SQL server")
		} else {
			__antithesis_instrumentation__.Notify(186246)
		}
		__antithesis_instrumentation__.Notify(186244)
		return cm.getNodeCertLocked()
	} else {
		__antithesis_instrumentation__.Notify(186247)
	}
	__antithesis_instrumentation__.Notify(186241)

	if err := checkCertIsValid(cm.nodeClientCert); err != nil {
		__antithesis_instrumentation__.Notify(186248)
		return nil, makeError(err, "problem with node client certificate")
	} else {
		__antithesis_instrumentation__.Notify(186249)
	}
	__antithesis_instrumentation__.Notify(186242)
	return cm.nodeClientCert, nil
}

func (cm *CertificateManager) getTenantCACertLocked() (*CertInfo, error) {
	__antithesis_instrumentation__.Notify(186250)
	if cm.tenantCACert == nil {
		__antithesis_instrumentation__.Notify(186253)
		return cm.getClientCACertLocked()
	} else {
		__antithesis_instrumentation__.Notify(186254)
	}
	__antithesis_instrumentation__.Notify(186251)
	c := cm.tenantCACert
	if err := checkCertIsValid(c); err != nil {
		__antithesis_instrumentation__.Notify(186255)
		return nil, makeError(err, "problem with tenant client CA certificate")
	} else {
		__antithesis_instrumentation__.Notify(186256)
	}
	__antithesis_instrumentation__.Notify(186252)
	return c, nil
}

func (cm *CertificateManager) getTenantCertLocked() (*CertInfo, error) {
	__antithesis_instrumentation__.Notify(186257)
	c := cm.tenantCert
	if err := checkCertIsValid(c); err != nil {
		__antithesis_instrumentation__.Notify(186259)
		return nil, makeError(err, "problem with tenant client certificate")
	} else {
		__antithesis_instrumentation__.Notify(186260)
	}
	__antithesis_instrumentation__.Notify(186258)
	return c, nil
}

func (cm *CertificateManager) GetTenantTLSConfig() (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(186261)
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.tenantConfig != nil {
		__antithesis_instrumentation__.Notify(186267)
		return cm.tenantConfig, nil
	} else {
		__antithesis_instrumentation__.Notify(186268)
	}
	__antithesis_instrumentation__.Notify(186262)

	ca, err := cm.getCACertLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(186269)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186270)
	}
	__antithesis_instrumentation__.Notify(186263)

	caBlob := ca.FileContents

	if cm.tenantCACert != nil {
		__antithesis_instrumentation__.Notify(186271)

		tenantCA, err := cm.getTenantCACertLocked()
		if err == nil {
			__antithesis_instrumentation__.Notify(186272)
			caBlob = AppendCertificatesToBlob(caBlob, tenantCA.FileContents)
		} else {
			__antithesis_instrumentation__.Notify(186273)
		}
	} else {
		__antithesis_instrumentation__.Notify(186274)
	}
	__antithesis_instrumentation__.Notify(186264)

	tenantCert, err := cm.getTenantCertLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(186275)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186276)
	}
	__antithesis_instrumentation__.Notify(186265)

	cfg, err := newClientTLSConfig(
		cm.tlsSettings,
		tenantCert.FileContents,
		tenantCert.KeyFileContents,
		caBlob)
	if err != nil {
		__antithesis_instrumentation__.Notify(186277)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186278)
	}
	__antithesis_instrumentation__.Notify(186266)

	cm.tenantConfig = cfg
	return cfg, nil
}

func (cm *CertificateManager) GetTenantSigningCert() (*CertInfo, error) {
	__antithesis_instrumentation__.Notify(186279)
	cm.mu.Lock()
	defer cm.mu.Unlock()

	c := cm.tenantSigningCert
	if err := checkCertIsValid(c); err != nil {
		__antithesis_instrumentation__.Notify(186281)
		return nil, makeError(err, "problem with tenant signing certificate")
	} else {
		__antithesis_instrumentation__.Notify(186282)
	}
	__antithesis_instrumentation__.Notify(186280)
	return c, nil
}

func (cm *CertificateManager) GetClientTLSConfig(user SQLUsername) (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(186283)
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var ca *CertInfo
	var err error
	if !cm.IsForTenant() {
		__antithesis_instrumentation__.Notify(186289)

		ca, err = cm.getCACertLocked()
		if err != nil {
			__antithesis_instrumentation__.Notify(186290)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(186291)
		}
	} else {
		__antithesis_instrumentation__.Notify(186292)

		ca, err = cm.getTenantCACertLocked()
		if err != nil {
			__antithesis_instrumentation__.Notify(186293)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(186294)
		}
	}
	__antithesis_instrumentation__.Notify(186284)

	if !user.IsNodeUser() {
		__antithesis_instrumentation__.Notify(186295)
		clientCert, err := cm.getClientCertLocked(user)
		if err != nil {
			__antithesis_instrumentation__.Notify(186298)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(186299)
		}
		__antithesis_instrumentation__.Notify(186296)

		cfg, err := newClientTLSConfig(
			cm.tlsSettings,
			clientCert.FileContents,
			clientCert.KeyFileContents,
			ca.FileContents)
		if err != nil {
			__antithesis_instrumentation__.Notify(186300)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(186301)
		}
		__antithesis_instrumentation__.Notify(186297)

		return cfg, nil
	} else {
		__antithesis_instrumentation__.Notify(186302)
	}
	__antithesis_instrumentation__.Notify(186285)

	if cm.clientConfig != nil {
		__antithesis_instrumentation__.Notify(186303)
		return cm.clientConfig, nil
	} else {
		__antithesis_instrumentation__.Notify(186304)
	}
	__antithesis_instrumentation__.Notify(186286)

	clientCert, err := cm.getNodeClientCertLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(186305)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186306)
	}
	__antithesis_instrumentation__.Notify(186287)

	cfg, err := newClientTLSConfig(
		cm.tlsSettings,
		clientCert.FileContents,
		clientCert.KeyFileContents,
		ca.FileContents)
	if err != nil {
		__antithesis_instrumentation__.Notify(186307)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186308)
	}
	__antithesis_instrumentation__.Notify(186288)

	cm.clientConfig = cfg
	return cfg, nil
}

func (cm *CertificateManager) GetUIClientTLSConfig() (*tls.Config, error) {
	__antithesis_instrumentation__.Notify(186309)
	cm.mu.Lock()
	defer cm.mu.Unlock()

	uiCA, err := cm.getUICACertLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(186313)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186314)
	}
	__antithesis_instrumentation__.Notify(186310)

	caBlob := uiCA.FileContents

	if cm.tenantCACert != nil {
		__antithesis_instrumentation__.Notify(186315)

		tenantCA, err := cm.getTenantCACertLocked()
		if err == nil {
			__antithesis_instrumentation__.Notify(186316)
			caBlob = AppendCertificatesToBlob(caBlob, tenantCA.FileContents)
		} else {
			__antithesis_instrumentation__.Notify(186317)
		}
	} else {
		__antithesis_instrumentation__.Notify(186318)
	}
	__antithesis_instrumentation__.Notify(186311)

	cfg, err := newUIClientTLSConfig(cm.tlsSettings, caBlob)
	if err != nil {
		__antithesis_instrumentation__.Notify(186319)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186320)
	}
	__antithesis_instrumentation__.Notify(186312)

	return cfg, nil
}

func (cm *CertificateManager) ListCertificates() ([]*CertInfo, error) {
	__antithesis_instrumentation__.Notify(186321)
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.initialized {
		__antithesis_instrumentation__.Notify(186329)
		return nil, errors.New("certificate manager has not been initialized")
	} else {
		__antithesis_instrumentation__.Notify(186330)
	}
	__antithesis_instrumentation__.Notify(186322)

	ret := make([]*CertInfo, 0, 2+len(cm.clientCerts))
	if cm.caCert != nil {
		__antithesis_instrumentation__.Notify(186331)
		ret = append(ret, cm.caCert)
	} else {
		__antithesis_instrumentation__.Notify(186332)
	}
	__antithesis_instrumentation__.Notify(186323)
	if cm.clientCACert != nil {
		__antithesis_instrumentation__.Notify(186333)
		ret = append(ret, cm.clientCACert)
	} else {
		__antithesis_instrumentation__.Notify(186334)
	}
	__antithesis_instrumentation__.Notify(186324)
	if cm.uiCACert != nil {
		__antithesis_instrumentation__.Notify(186335)
		ret = append(ret, cm.uiCACert)
	} else {
		__antithesis_instrumentation__.Notify(186336)
	}
	__antithesis_instrumentation__.Notify(186325)
	if cm.nodeCert != nil {
		__antithesis_instrumentation__.Notify(186337)
		ret = append(ret, cm.nodeCert)
	} else {
		__antithesis_instrumentation__.Notify(186338)
	}
	__antithesis_instrumentation__.Notify(186326)
	if cm.uiCert != nil {
		__antithesis_instrumentation__.Notify(186339)
		ret = append(ret, cm.uiCert)
	} else {
		__antithesis_instrumentation__.Notify(186340)
	}
	__antithesis_instrumentation__.Notify(186327)
	if cm.clientCerts != nil {
		__antithesis_instrumentation__.Notify(186341)
		for _, cert := range cm.clientCerts {
			__antithesis_instrumentation__.Notify(186342)
			ret = append(ret, cert)
		}
	} else {
		__antithesis_instrumentation__.Notify(186343)
	}
	__antithesis_instrumentation__.Notify(186328)

	return ret, nil
}
