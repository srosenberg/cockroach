package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "crypto/x509"

func KeyUsageToString(ku x509.KeyUsage) []string {
	__antithesis_instrumentation__.Notify(187379)
	ret := make([]string, 0)
	if ku&x509.KeyUsageDigitalSignature != 0 {
		__antithesis_instrumentation__.Notify(187389)
		ret = append(ret, "DigitalSignature")
	} else {
		__antithesis_instrumentation__.Notify(187390)
	}
	__antithesis_instrumentation__.Notify(187380)
	if ku&x509.KeyUsageContentCommitment != 0 {
		__antithesis_instrumentation__.Notify(187391)
		ret = append(ret, "ContentCommitment")
	} else {
		__antithesis_instrumentation__.Notify(187392)
	}
	__antithesis_instrumentation__.Notify(187381)
	if ku&x509.KeyUsageKeyEncipherment != 0 {
		__antithesis_instrumentation__.Notify(187393)
		ret = append(ret, "KeyEncipherment")
	} else {
		__antithesis_instrumentation__.Notify(187394)
	}
	__antithesis_instrumentation__.Notify(187382)
	if ku&x509.KeyUsageDataEncipherment != 0 {
		__antithesis_instrumentation__.Notify(187395)
		ret = append(ret, "DataEncirpherment")
	} else {
		__antithesis_instrumentation__.Notify(187396)
	}
	__antithesis_instrumentation__.Notify(187383)
	if ku&x509.KeyUsageKeyAgreement != 0 {
		__antithesis_instrumentation__.Notify(187397)
		ret = append(ret, "KeyAgreement")
	} else {
		__antithesis_instrumentation__.Notify(187398)
	}
	__antithesis_instrumentation__.Notify(187384)
	if ku&x509.KeyUsageCertSign != 0 {
		__antithesis_instrumentation__.Notify(187399)
		ret = append(ret, "CertSign")
	} else {
		__antithesis_instrumentation__.Notify(187400)
	}
	__antithesis_instrumentation__.Notify(187385)
	if ku&x509.KeyUsageCRLSign != 0 {
		__antithesis_instrumentation__.Notify(187401)
		ret = append(ret, "CRLSign")
	} else {
		__antithesis_instrumentation__.Notify(187402)
	}
	__antithesis_instrumentation__.Notify(187386)
	if ku&x509.KeyUsageEncipherOnly != 0 {
		__antithesis_instrumentation__.Notify(187403)
		ret = append(ret, "EncipherOnly")
	} else {
		__antithesis_instrumentation__.Notify(187404)
	}
	__antithesis_instrumentation__.Notify(187387)
	if ku&x509.KeyUsageDecipherOnly != 0 {
		__antithesis_instrumentation__.Notify(187405)
		ret = append(ret, "DecipherOnly")
	} else {
		__antithesis_instrumentation__.Notify(187406)
	}
	__antithesis_instrumentation__.Notify(187388)

	return ret
}

func ExtKeyUsageToString(eku x509.ExtKeyUsage) string {
	__antithesis_instrumentation__.Notify(187407)
	switch eku {

	case x509.ExtKeyUsageAny:
		__antithesis_instrumentation__.Notify(187408)
		return "Any"
	case x509.ExtKeyUsageServerAuth:
		__antithesis_instrumentation__.Notify(187409)
		return "ServerAuth"
	case x509.ExtKeyUsageClientAuth:
		__antithesis_instrumentation__.Notify(187410)
		return "ClientAuth"
	case x509.ExtKeyUsageCodeSigning:
		__antithesis_instrumentation__.Notify(187411)
		return "CodeSigning"
	case x509.ExtKeyUsageEmailProtection:
		__antithesis_instrumentation__.Notify(187412)
		return "EmailProtection"
	case x509.ExtKeyUsageIPSECEndSystem:
		__antithesis_instrumentation__.Notify(187413)
		return "IPSECEndSystem"
	case x509.ExtKeyUsageIPSECTunnel:
		__antithesis_instrumentation__.Notify(187414)
		return "IPSECTunnel"
	case x509.ExtKeyUsageIPSECUser:
		__antithesis_instrumentation__.Notify(187415)
		return "IPSECUser"
	case x509.ExtKeyUsageTimeStamping:
		__antithesis_instrumentation__.Notify(187416)
		return "TimeStamping"
	case x509.ExtKeyUsageOCSPSigning:
		__antithesis_instrumentation__.Notify(187417)
		return "OCSPSigning"
	case x509.ExtKeyUsageMicrosoftServerGatedCrypto:
		__antithesis_instrumentation__.Notify(187418)
		return "MicrosoftServerGatedCrypto"
	case x509.ExtKeyUsageNetscapeServerGatedCrypto:
		__antithesis_instrumentation__.Notify(187419)
		return "NetscapeServerGatedCrypto"
	default:
		__antithesis_instrumentation__.Notify(187420)
		return "unknown"
	}
}
