package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/x509"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
)

func init() {
	if runtime.GOOS == "windows" {

		skipPermissionChecks = true
	} else {
		skipPermissionChecks = envutil.EnvOrDefaultBool("COCKROACH_SKIP_KEY_PERMISSION_CHECK", false)
	}
}

var skipPermissionChecks bool

type AssetLoader struct {
	ReadDir  func(dirname string) ([]os.FileInfo, error)
	ReadFile func(filename string) ([]byte, error)
	Stat     func(name string) (os.FileInfo, error)
}

var defaultAssetLoader = AssetLoader{
	ReadDir:  ioutil.ReadDir,
	ReadFile: ioutil.ReadFile,
	Stat:     os.Stat,
}

var assetLoaderImpl = defaultAssetLoader

func GetAssetLoader() AssetLoader {
	__antithesis_instrumentation__.Notify(185827)
	return assetLoaderImpl
}

func SetAssetLoader(al AssetLoader) {
	__antithesis_instrumentation__.Notify(185828)
	assetLoaderImpl = al
}

func ResetAssetLoader() {
	__antithesis_instrumentation__.Notify(185829)
	assetLoaderImpl = defaultAssetLoader
}

type PemUsage uint32

const (
	_ PemUsage = iota

	CAPem

	TenantCAPem

	ClientCAPem

	UICAPem

	NodePem

	UIPem

	ClientPem

	TenantPem

	TenantSigningPem

	maxKeyPermissions os.FileMode = 0700

	maxGroupKeyPermissions os.FileMode = 0740

	certExtension = `.crt`
	keyExtension  = `.key`

	defaultCertsDirPerm = 0700
)

func isCA(usage PemUsage) bool {
	__antithesis_instrumentation__.Notify(185830)
	return usage == CAPem || func() bool {
		__antithesis_instrumentation__.Notify(185831)
		return usage == ClientCAPem == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(185832)
		return usage == TenantCAPem == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(185833)
		return usage == UICAPem == true
	}() == true
}

func (p PemUsage) String() string {
	__antithesis_instrumentation__.Notify(185834)
	switch p {
	case CAPem:
		__antithesis_instrumentation__.Notify(185835)
		return "CA"
	case ClientCAPem:
		__antithesis_instrumentation__.Notify(185836)
		return "Client CA"
	case TenantCAPem:
		__antithesis_instrumentation__.Notify(185837)
		return "Tenant Client CA"
	case UICAPem:
		__antithesis_instrumentation__.Notify(185838)
		return "UI CA"
	case NodePem:
		__antithesis_instrumentation__.Notify(185839)
		return "Node"
	case UIPem:
		__antithesis_instrumentation__.Notify(185840)
		return "UI"
	case ClientPem:
		__antithesis_instrumentation__.Notify(185841)
		return "Client"
	case TenantPem:
		__antithesis_instrumentation__.Notify(185842)
		return "Tenant Client"
	default:
		__antithesis_instrumentation__.Notify(185843)
		return "unknown"
	}
}

type CertInfo struct {
	FileUsage PemUsage

	Filename string

	FileContents []byte

	KeyFilename string

	KeyFileContents []byte

	Name string

	ParsedCertificates []*x509.Certificate

	ExpirationTime time.Time

	Error error
}

func isCertificateFile(filename string) bool {
	__antithesis_instrumentation__.Notify(185844)
	return strings.HasSuffix(filename, certExtension)
}

func CertInfoFromFilename(filename string) (*CertInfo, error) {
	__antithesis_instrumentation__.Notify(185845)
	parts := strings.Split(filename, `.`)
	numParts := len(parts)

	if numParts < 2 {
		__antithesis_instrumentation__.Notify(185848)
		return nil, errors.New("not enough parts found")
	} else {
		__antithesis_instrumentation__.Notify(185849)
	}
	__antithesis_instrumentation__.Notify(185846)

	var fileUsage PemUsage
	var name string
	prefix := parts[0]
	switch parts[0] {
	case `ca`:
		__antithesis_instrumentation__.Notify(185850)
		fileUsage = CAPem
		if numParts != 2 {
			__antithesis_instrumentation__.Notify(185860)
			return nil, errors.Errorf("CA certificate filename should match ca%s", certExtension)
		} else {
			__antithesis_instrumentation__.Notify(185861)
		}
	case `ca-client`:
		__antithesis_instrumentation__.Notify(185851)
		fileUsage = ClientCAPem
		if numParts != 2 {
			__antithesis_instrumentation__.Notify(185862)
			return nil, errors.Errorf("client CA certificate filename should match ca-client%s", certExtension)
		} else {
			__antithesis_instrumentation__.Notify(185863)
		}
	case `ca-client-tenant`:
		__antithesis_instrumentation__.Notify(185852)
		fileUsage = TenantCAPem
		if numParts != 2 {
			__antithesis_instrumentation__.Notify(185864)
			return nil, errors.Errorf("tenant CA certificate filename should match ca%s", certExtension)
		} else {
			__antithesis_instrumentation__.Notify(185865)
		}
	case `ca-ui`:
		__antithesis_instrumentation__.Notify(185853)
		fileUsage = UICAPem
		if numParts != 2 {
			__antithesis_instrumentation__.Notify(185866)
			return nil, errors.Errorf("UI CA certificate filename should match ca-ui%s", certExtension)
		} else {
			__antithesis_instrumentation__.Notify(185867)
		}
	case `node`:
		__antithesis_instrumentation__.Notify(185854)
		fileUsage = NodePem
		if numParts != 2 {
			__antithesis_instrumentation__.Notify(185868)
			return nil, errors.Errorf("node certificate filename should match node%s", certExtension)
		} else {
			__antithesis_instrumentation__.Notify(185869)
		}
	case `ui`:
		__antithesis_instrumentation__.Notify(185855)
		fileUsage = UIPem
		if numParts != 2 {
			__antithesis_instrumentation__.Notify(185870)
			return nil, errors.Errorf("UI certificate filename should match ui%s", certExtension)
		} else {
			__antithesis_instrumentation__.Notify(185871)
		}
	case `client`:
		__antithesis_instrumentation__.Notify(185856)
		fileUsage = ClientPem

		name = strings.Join(parts[1:numParts-1], `.`)
		if len(name) == 0 {
			__antithesis_instrumentation__.Notify(185872)
			return nil, errors.Errorf("client certificate filename should match client.<user>%s", certExtension)
		} else {
			__antithesis_instrumentation__.Notify(185873)
		}
	case `client-tenant`:
		__antithesis_instrumentation__.Notify(185857)
		fileUsage = TenantPem

		name = strings.Join(parts[1:numParts-1], `.`)
		if len(name) == 0 {
			__antithesis_instrumentation__.Notify(185874)
			return nil, errors.Errorf("tenant certificate filename should match client-tenant.<tenantid>%s", certExtension)
		} else {
			__antithesis_instrumentation__.Notify(185875)
		}
	case `tenant-signing`:
		__antithesis_instrumentation__.Notify(185858)
		fileUsage = TenantSigningPem

		name = strings.Join(parts[1:numParts-1], `.`)
		if len(name) == 0 {
			__antithesis_instrumentation__.Notify(185876)
			return nil, errors.Errorf("tenant signing certificate filename should match tenant-signing.<tenantid>%s", certExtension)
		} else {
			__antithesis_instrumentation__.Notify(185877)
		}
	default:
		__antithesis_instrumentation__.Notify(185859)
		return nil, errors.Errorf("unknown prefix %q", prefix)
	}
	__antithesis_instrumentation__.Notify(185847)

	return &CertInfo{
		FileUsage: fileUsage,
		Filename:  filename,
		Name:      name,
	}, nil
}

type CertificateLoader struct {
	certsDir             string
	skipPermissionChecks bool
	certificates         []*CertInfo
}

func (cl *CertificateLoader) Certificates() []*CertInfo {
	__antithesis_instrumentation__.Notify(185878)
	return cl.certificates
}

func NewCertificateLoader(certsDir string) *CertificateLoader {
	__antithesis_instrumentation__.Notify(185879)
	return &CertificateLoader{
		certsDir:             certsDir,
		skipPermissionChecks: skipPermissionChecks,
		certificates:         make([]*CertInfo, 0),
	}
}

func (cl *CertificateLoader) MaybeCreateCertsDir() error {
	__antithesis_instrumentation__.Notify(185880)
	dirInfo, err := os.Stat(cl.certsDir)
	if err == nil {
		__antithesis_instrumentation__.Notify(185884)
		if !dirInfo.IsDir() {
			__antithesis_instrumentation__.Notify(185886)
			return errors.Errorf("certs directory %s exists but is not a directory", cl.certsDir)
		} else {
			__antithesis_instrumentation__.Notify(185887)
		}
		__antithesis_instrumentation__.Notify(185885)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(185888)
	}
	__antithesis_instrumentation__.Notify(185881)

	if !oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(185889)
		return makeErrorf(err, "could not stat certs directory %s", cl.certsDir)
	} else {
		__antithesis_instrumentation__.Notify(185890)
	}
	__antithesis_instrumentation__.Notify(185882)

	if err := os.Mkdir(cl.certsDir, defaultCertsDirPerm); err != nil {
		__antithesis_instrumentation__.Notify(185891)
		return makeErrorf(err, "could not create certs directory %s", cl.certsDir)
	} else {
		__antithesis_instrumentation__.Notify(185892)
	}
	__antithesis_instrumentation__.Notify(185883)
	return nil
}

func (cl *CertificateLoader) TestDisablePermissionChecks() {
	__antithesis_instrumentation__.Notify(185893)
	cl.skipPermissionChecks = true
}

func (cl *CertificateLoader) Load() error {
	__antithesis_instrumentation__.Notify(185894)
	fileInfos, err := assetLoaderImpl.ReadDir(cl.certsDir)
	if err != nil {
		__antithesis_instrumentation__.Notify(185898)
		if oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(185900)

			if log.V(3) {
				__antithesis_instrumentation__.Notify(185902)
				log.Infof(context.Background(), "missing certs directory %s", cl.certsDir)
			} else {
				__antithesis_instrumentation__.Notify(185903)
			}
			__antithesis_instrumentation__.Notify(185901)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(185904)
		}
		__antithesis_instrumentation__.Notify(185899)
		return err
	} else {
		__antithesis_instrumentation__.Notify(185905)
	}
	__antithesis_instrumentation__.Notify(185895)

	if log.V(3) {
		__antithesis_instrumentation__.Notify(185906)
		log.Infof(context.Background(), "scanning certs directory %s", cl.certsDir)
	} else {
		__antithesis_instrumentation__.Notify(185907)
	}
	__antithesis_instrumentation__.Notify(185896)

	for _, info := range fileInfos {
		__antithesis_instrumentation__.Notify(185908)
		filename := info.Name()
		fullPath := filepath.Join(cl.certsDir, filename)

		if info.IsDir() {
			__antithesis_instrumentation__.Notify(185914)

			if log.V(3) {
				__antithesis_instrumentation__.Notify(185916)
				log.Infof(context.Background(), "skipping sub-directory %s", fullPath)
			} else {
				__antithesis_instrumentation__.Notify(185917)
			}
			__antithesis_instrumentation__.Notify(185915)
			continue
		} else {
			__antithesis_instrumentation__.Notify(185918)
		}
		__antithesis_instrumentation__.Notify(185909)

		if !isCertificateFile(filename) {
			__antithesis_instrumentation__.Notify(185919)
			if log.V(3) {
				__antithesis_instrumentation__.Notify(185921)
				log.Infof(context.Background(), "skipping non-certificate file %s", filename)
			} else {
				__antithesis_instrumentation__.Notify(185922)
			}
			__antithesis_instrumentation__.Notify(185920)
			continue
		} else {
			__antithesis_instrumentation__.Notify(185923)
		}
		__antithesis_instrumentation__.Notify(185910)

		ci, err := CertInfoFromFilename(filename)
		if err != nil {
			__antithesis_instrumentation__.Notify(185924)
			log.Warningf(context.Background(), "bad filename %s: %v", fullPath, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(185925)
		}
		__antithesis_instrumentation__.Notify(185911)

		fullCertPath := filepath.Join(cl.certsDir, filename)
		certPEMBlock, err := assetLoaderImpl.ReadFile(fullCertPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(185926)
			log.Warningf(context.Background(), "could not read certificate file %s: %v", fullPath, err)
		} else {
			__antithesis_instrumentation__.Notify(185927)
		}
		__antithesis_instrumentation__.Notify(185912)
		ci.FileContents = certPEMBlock

		if err := parseCertificate(ci); err != nil {
			__antithesis_instrumentation__.Notify(185928)
			log.Warningf(context.Background(), "could not parse certificate for %s: %v", fullPath, err)
			ci.Error = err
		} else {
			__antithesis_instrumentation__.Notify(185929)
			if err := cl.findKey(ci); err != nil {
				__antithesis_instrumentation__.Notify(185930)
				log.Warningf(context.Background(), "error finding key for %s: %v", fullPath, err)
				ci.Error = err
			} else {
				__antithesis_instrumentation__.Notify(185931)
				if log.V(3) {
					__antithesis_instrumentation__.Notify(185932)
					log.Infof(context.Background(), "found certificate %s", ci.Filename)
				} else {
					__antithesis_instrumentation__.Notify(185933)
				}
			}
		}
		__antithesis_instrumentation__.Notify(185913)

		cl.certificates = append(cl.certificates, ci)
	}
	__antithesis_instrumentation__.Notify(185897)

	return nil
}

func (cl *CertificateLoader) findKey(ci *CertInfo) error {
	__antithesis_instrumentation__.Notify(185934)
	if isCA(ci.FileUsage) {
		__antithesis_instrumentation__.Notify(185940)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(185941)
	}
	__antithesis_instrumentation__.Notify(185935)

	keyFilename := strings.TrimSuffix(ci.Filename, certExtension) + keyExtension
	fullKeyPath := filepath.Join(cl.certsDir, keyFilename)

	info, err := assetLoaderImpl.Stat(fullKeyPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(185942)
		return errors.Wrapf(err, "could not stat key file %s", fullKeyPath)
	} else {
		__antithesis_instrumentation__.Notify(185943)
	}
	__antithesis_instrumentation__.Notify(185936)

	fileMode := info.Mode()
	if !fileMode.IsRegular() {
		__antithesis_instrumentation__.Notify(185944)
		return errors.Errorf("key file %s is not a regular file", fullKeyPath)
	} else {
		__antithesis_instrumentation__.Notify(185945)
	}
	__antithesis_instrumentation__.Notify(185937)

	if !cl.skipPermissionChecks {
		__antithesis_instrumentation__.Notify(185946)
		aclInfo := sysutil.GetFileACLInfo(info)
		if err = checkFilePermissions(os.Getgid(), fullKeyPath, aclInfo); err != nil {
			__antithesis_instrumentation__.Notify(185947)
			return err
		} else {
			__antithesis_instrumentation__.Notify(185948)
		}
	} else {
		__antithesis_instrumentation__.Notify(185949)
	}
	__antithesis_instrumentation__.Notify(185938)

	keyPEMBlock, err := assetLoaderImpl.ReadFile(fullKeyPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(185950)
		return errors.Wrapf(err, "could not read key file %s", fullKeyPath)
	} else {
		__antithesis_instrumentation__.Notify(185951)
	}
	__antithesis_instrumentation__.Notify(185939)

	ci.KeyFilename = keyFilename
	ci.KeyFileContents = keyPEMBlock
	return nil
}

func parseCertificate(ci *CertInfo) error {
	__antithesis_instrumentation__.Notify(185952)
	if ci.Error != nil {
		__antithesis_instrumentation__.Notify(185958)
		return makeErrorf(ci.Error, "parseCertificate called on bad CertInfo object: %s", ci.Filename)
	} else {
		__antithesis_instrumentation__.Notify(185959)
	}
	__antithesis_instrumentation__.Notify(185953)

	if len(ci.FileContents) == 0 {
		__antithesis_instrumentation__.Notify(185960)
		return errors.Errorf("empty certificate file: %s", ci.Filename)
	} else {
		__antithesis_instrumentation__.Notify(185961)
	}
	__antithesis_instrumentation__.Notify(185954)

	derCerts, err := PEMToCertificates(ci.FileContents)
	if err != nil {
		__antithesis_instrumentation__.Notify(185962)
		return makeErrorf(err, "failed to parse certificate file %s as PEM", ci.Filename)
	} else {
		__antithesis_instrumentation__.Notify(185963)
	}
	__antithesis_instrumentation__.Notify(185955)

	if len(derCerts) == 0 {
		__antithesis_instrumentation__.Notify(185964)
		return errors.Errorf("no certificates found in %s", ci.Filename)
	} else {
		__antithesis_instrumentation__.Notify(185965)
	}
	__antithesis_instrumentation__.Notify(185956)

	certs := make([]*x509.Certificate, len(derCerts))
	var expires time.Time
	for i, c := range derCerts {
		__antithesis_instrumentation__.Notify(185966)
		x509Cert, err := x509.ParseCertificate(c.Bytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(185969)
			return makeErrorf(err, "failed to parse certificate %d in file %s", i, ci.Filename)
		} else {
			__antithesis_instrumentation__.Notify(185970)
		}
		__antithesis_instrumentation__.Notify(185967)

		if i == 0 {
			__antithesis_instrumentation__.Notify(185971)

			if err := validateCockroachCertificate(ci, x509Cert); err != nil {
				__antithesis_instrumentation__.Notify(185973)
				return makeErrorf(err, "failed to validate certificate %d in file %s", i, ci.Filename)
			} else {
				__antithesis_instrumentation__.Notify(185974)
			}
			__antithesis_instrumentation__.Notify(185972)

			expires = x509Cert.NotAfter
		} else {
			__antithesis_instrumentation__.Notify(185975)
		}
		__antithesis_instrumentation__.Notify(185968)
		certs[i] = x509Cert
	}
	__antithesis_instrumentation__.Notify(185957)

	ci.ParsedCertificates = certs
	ci.ExpirationTime = expires
	return nil
}

func validateDualPurposeNodeCert(ci *CertInfo) error {
	__antithesis_instrumentation__.Notify(185976)
	if ci == nil {
		__antithesis_instrumentation__.Notify(185980)
		return errors.Errorf("no node certificate found")
	} else {
		__antithesis_instrumentation__.Notify(185981)
	}
	__antithesis_instrumentation__.Notify(185977)

	if ci.Error != nil {
		__antithesis_instrumentation__.Notify(185982)
		return ci.Error
	} else {
		__antithesis_instrumentation__.Notify(185983)
	}
	__antithesis_instrumentation__.Notify(185978)

	cert := ci.ParsedCertificates[0]
	principals := getCertificatePrincipals(cert)
	if !Contains(principals, NodeUser) {
		__antithesis_instrumentation__.Notify(185984)
		return errors.Errorf("client/server node certificate has principals %q, expected %q",
			principals, NodeUser)
	} else {
		__antithesis_instrumentation__.Notify(185985)
	}
	__antithesis_instrumentation__.Notify(185979)

	return nil
}

func validateCockroachCertificate(ci *CertInfo, cert *x509.Certificate) error {
	__antithesis_instrumentation__.Notify(185986)

	switch ci.FileUsage {
	case NodePem:
		__antithesis_instrumentation__.Notify(185988)

	case ClientPem:
		__antithesis_instrumentation__.Notify(185989)

		principals := getCertificatePrincipals(cert)
		if !Contains(principals, ci.Name) {
			__antithesis_instrumentation__.Notify(185991)
			return errors.Errorf("client certificate has principals %q, expected %q",
				principals, ci.Name)
		} else {
			__antithesis_instrumentation__.Notify(185992)
		}
	default:
		__antithesis_instrumentation__.Notify(185990)
	}
	__antithesis_instrumentation__.Notify(185987)
	return nil
}
