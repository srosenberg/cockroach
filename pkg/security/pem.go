package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
)

func WritePEMToFile(path string, mode os.FileMode, overwrite bool, blocks ...*pem.Block) error {
	__antithesis_instrumentation__.Notify(186915)
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if !overwrite {
		__antithesis_instrumentation__.Notify(186919)
		flags |= os.O_EXCL
	} else {
		__antithesis_instrumentation__.Notify(186920)
	}
	__antithesis_instrumentation__.Notify(186916)
	f, err := os.OpenFile(path, flags, mode)
	if err != nil {
		__antithesis_instrumentation__.Notify(186921)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186922)
	}
	__antithesis_instrumentation__.Notify(186917)

	for _, p := range blocks {
		__antithesis_instrumentation__.Notify(186923)
		if err := pem.Encode(f, p); err != nil {
			__antithesis_instrumentation__.Notify(186924)
			return errors.Wrap(err, "could not encode PEM block")
		} else {
			__antithesis_instrumentation__.Notify(186925)
		}
	}
	__antithesis_instrumentation__.Notify(186918)

	return f.Close()
}

func SafeWriteToFile(path string, mode os.FileMode, overwrite bool, contents []byte) error {
	__antithesis_instrumentation__.Notify(186926)
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if !overwrite {
		__antithesis_instrumentation__.Notify(186931)
		flags |= os.O_EXCL
	} else {
		__antithesis_instrumentation__.Notify(186932)
	}
	__antithesis_instrumentation__.Notify(186927)
	f, err := os.OpenFile(path, flags, mode)
	if err != nil {
		__antithesis_instrumentation__.Notify(186933)
		return err
	} else {
		__antithesis_instrumentation__.Notify(186934)
	}
	__antithesis_instrumentation__.Notify(186928)

	n, err := f.Write(contents)
	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(186935)
		return n < len(contents) == true
	}() == true {
		__antithesis_instrumentation__.Notify(186936)
		err = io.ErrShortWrite
	} else {
		__antithesis_instrumentation__.Notify(186937)
	}
	__antithesis_instrumentation__.Notify(186929)
	if err1 := f.Close(); err == nil {
		__antithesis_instrumentation__.Notify(186938)
		err = err1
	} else {
		__antithesis_instrumentation__.Notify(186939)
	}
	__antithesis_instrumentation__.Notify(186930)
	return err
}

func PrivateKeyToPEM(key crypto.PrivateKey) (*pem.Block, error) {
	__antithesis_instrumentation__.Notify(186940)
	switch k := key.(type) {
	case *rsa.PrivateKey:
		__antithesis_instrumentation__.Notify(186941)
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
	case *ecdsa.PrivateKey:
		__antithesis_instrumentation__.Notify(186942)
		bytes, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			__antithesis_instrumentation__.Notify(186947)
			return nil, errors.Wrap(err, "error marshaling ECDSA key")
		} else {
			__antithesis_instrumentation__.Notify(186948)
		}
		__antithesis_instrumentation__.Notify(186943)
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: bytes}, nil
	case ed25519.PrivateKey:
		__antithesis_instrumentation__.Notify(186944)
		bytes, err := x509.MarshalPKCS8PrivateKey(k)
		if err != nil {
			__antithesis_instrumentation__.Notify(186949)
			return nil, errors.Wrap(err, "error marshaling Ed25519 key")
		} else {
			__antithesis_instrumentation__.Notify(186950)
		}
		__antithesis_instrumentation__.Notify(186945)
		return &pem.Block{Type: "PRIVATE KEY", Bytes: bytes}, nil
	default:
		__antithesis_instrumentation__.Notify(186946)
		return nil, errors.Errorf("unknown key type: %v", k)
	}
}

func PrivateKeyToPKCS8(key crypto.PrivateKey) ([]byte, error) {
	__antithesis_instrumentation__.Notify(186951)
	return x509.MarshalPKCS8PrivateKey(key)
}

func PEMToCertificates(contents []byte) ([]*pem.Block, error) {
	__antithesis_instrumentation__.Notify(186952)
	certs := make([]*pem.Block, 0)
	for {
		__antithesis_instrumentation__.Notify(186954)
		var block *pem.Block
		block, contents = pem.Decode(contents)
		if block == nil {
			__antithesis_instrumentation__.Notify(186957)
			break
		} else {
			__antithesis_instrumentation__.Notify(186958)
		}
		__antithesis_instrumentation__.Notify(186955)
		if block.Type != "CERTIFICATE" {
			__antithesis_instrumentation__.Notify(186959)
			return nil, errors.Errorf("block #%d is of type %s, not CERTIFICATE", len(certs), block.Type)
		} else {
			__antithesis_instrumentation__.Notify(186960)
		}
		__antithesis_instrumentation__.Notify(186956)

		certs = append(certs, block)
	}
	__antithesis_instrumentation__.Notify(186953)

	return certs, nil
}

func PEMToPrivateKey(contents []byte) (crypto.PrivateKey, error) {
	__antithesis_instrumentation__.Notify(186961)
	keyBlock, remaining := pem.Decode(contents)
	if keyBlock == nil {
		__antithesis_instrumentation__.Notify(186965)
		return nil, errors.New("no PEM data found")
	} else {
		__antithesis_instrumentation__.Notify(186966)
	}
	__antithesis_instrumentation__.Notify(186962)
	if len(remaining) != 0 {
		__antithesis_instrumentation__.Notify(186967)
		return nil, errors.New("more than one PEM block found")
	} else {
		__antithesis_instrumentation__.Notify(186968)
	}
	__antithesis_instrumentation__.Notify(186963)
	if keyBlock.Type != "PRIVATE KEY" && func() bool {
		__antithesis_instrumentation__.Notify(186969)
		return !strings.HasSuffix(keyBlock.Type, " PRIVATE KEY") == true
	}() == true {
		__antithesis_instrumentation__.Notify(186970)
		return nil, errors.Errorf("PEM block is of type %s", keyBlock.Type)
	} else {
		__antithesis_instrumentation__.Notify(186971)
	}
	__antithesis_instrumentation__.Notify(186964)

	return parsePrivateKey(keyBlock.Bytes)
}

func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	__antithesis_instrumentation__.Notify(186972)
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		__antithesis_instrumentation__.Notify(186976)
		return key, nil
	} else {
		__antithesis_instrumentation__.Notify(186977)
	}
	__antithesis_instrumentation__.Notify(186973)
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		__antithesis_instrumentation__.Notify(186978)
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			__antithesis_instrumentation__.Notify(186979)
			return key, nil
		default:
			__antithesis_instrumentation__.Notify(186980)
			return nil, errors.New("found unknown private key type in PKCS#8 wrapping")
		}
	} else {
		__antithesis_instrumentation__.Notify(186981)
	}
	__antithesis_instrumentation__.Notify(186974)
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		__antithesis_instrumentation__.Notify(186982)
		return key, nil
	} else {
		__antithesis_instrumentation__.Notify(186983)
	}
	__antithesis_instrumentation__.Notify(186975)

	return nil, errors.New("failed to parse private key")
}
