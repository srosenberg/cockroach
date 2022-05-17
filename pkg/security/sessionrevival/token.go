package sessionrevival

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/ed25519"
	"time"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	pbtypes "github.com/gogo/protobuf/types"
)

const tokenLifetime = 10 * time.Minute

func CreateSessionRevivalToken(
	cm *security.CertificateManager, user security.SQLUsername,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(187245)
	cert, err := cm.GetTenantSigningCert()
	if err != nil {
		__antithesis_instrumentation__.Notify(187252)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187253)
	}
	__antithesis_instrumentation__.Notify(187246)
	key, err := security.PEMToPrivateKey(cert.KeyFileContents)
	if err != nil {
		__antithesis_instrumentation__.Notify(187254)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187255)
	}
	__antithesis_instrumentation__.Notify(187247)

	now := timeutil.Now()
	issuedAt, err := pbtypes.TimestampProto(now)
	if err != nil {
		__antithesis_instrumentation__.Notify(187256)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187257)
	}
	__antithesis_instrumentation__.Notify(187248)
	expiresAt, err := pbtypes.TimestampProto(now.Add(tokenLifetime))
	if err != nil {
		__antithesis_instrumentation__.Notify(187258)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187259)
	}
	__antithesis_instrumentation__.Notify(187249)

	payload := &sessiondatapb.SessionRevivalToken_Payload{
		User:      user.Normalized(),
		Algorithm: cert.ParsedCertificates[0].PublicKeyAlgorithm.String(),
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}
	payloadBytes, err := protoutil.Marshal(payload)
	if err != nil {
		__antithesis_instrumentation__.Notify(187260)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187261)
	}
	__antithesis_instrumentation__.Notify(187250)

	signature := ed25519.Sign(key.(ed25519.PrivateKey), payloadBytes)

	token := &sessiondatapb.SessionRevivalToken{
		Payload:   payloadBytes,
		Signature: signature,
	}
	tokenBytes, err := protoutil.Marshal(token)
	if err != nil {
		__antithesis_instrumentation__.Notify(187262)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187263)
	}
	__antithesis_instrumentation__.Notify(187251)

	return tokenBytes, nil
}

func ValidateSessionRevivalToken(
	cm *security.CertificateManager, user security.SQLUsername, tokenBytes []byte,
) error {
	__antithesis_instrumentation__.Notify(187264)
	cert, err := cm.GetTenantSigningCert()
	if err != nil {
		__antithesis_instrumentation__.Notify(187270)
		return err
	} else {
		__antithesis_instrumentation__.Notify(187271)
	}
	__antithesis_instrumentation__.Notify(187265)

	token := &sessiondatapb.SessionRevivalToken{}
	payload := &sessiondatapb.SessionRevivalToken_Payload{}
	err = protoutil.Unmarshal(tokenBytes, token)
	if err != nil {
		__antithesis_instrumentation__.Notify(187272)
		return err
	} else {
		__antithesis_instrumentation__.Notify(187273)
	}
	__antithesis_instrumentation__.Notify(187266)
	err = protoutil.Unmarshal(token.Payload, payload)
	if err != nil {
		__antithesis_instrumentation__.Notify(187274)
		return err
	} else {
		__antithesis_instrumentation__.Notify(187275)
	}
	__antithesis_instrumentation__.Notify(187267)
	if err := validatePayloadContents(payload, user); err != nil {
		__antithesis_instrumentation__.Notify(187276)
		return err
	} else {
		__antithesis_instrumentation__.Notify(187277)
	}
	__antithesis_instrumentation__.Notify(187268)
	for _, c := range cert.ParsedCertificates {
		__antithesis_instrumentation__.Notify(187278)
		if err := c.CheckSignature(c.SignatureAlgorithm, token.Payload, token.Signature); err == nil {
			__antithesis_instrumentation__.Notify(187279)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(187280)
		}
	}
	__antithesis_instrumentation__.Notify(187269)
	return errors.New("invalid signature")
}

func validatePayloadContents(
	payload *sessiondatapb.SessionRevivalToken_Payload, user security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(187281)
	issuedAt, err := pbtypes.TimestampFromProto(payload.IssuedAt)
	if err != nil {
		__antithesis_instrumentation__.Notify(187288)
		return err
	} else {
		__antithesis_instrumentation__.Notify(187289)
	}
	__antithesis_instrumentation__.Notify(187282)
	expiresAt, err := pbtypes.TimestampFromProto(payload.ExpiresAt)
	if err != nil {
		__antithesis_instrumentation__.Notify(187290)
		return err
	} else {
		__antithesis_instrumentation__.Notify(187291)
	}
	__antithesis_instrumentation__.Notify(187283)

	now := timeutil.Now()
	if now.Before(issuedAt) {
		__antithesis_instrumentation__.Notify(187292)
		return errors.Errorf("token issue time is in the future (%v)", issuedAt)
	} else {
		__antithesis_instrumentation__.Notify(187293)
	}
	__antithesis_instrumentation__.Notify(187284)
	if now.After(expiresAt) {
		__antithesis_instrumentation__.Notify(187294)
		return errors.Errorf("token expiration time is in the past (%v)", expiresAt)
	} else {
		__antithesis_instrumentation__.Notify(187295)
	}
	__antithesis_instrumentation__.Notify(187285)

	if issuedAt.Add(tokenLifetime + 1*time.Minute).Before(expiresAt) {
		__antithesis_instrumentation__.Notify(187296)
		return errors.Errorf("token expiration time is too far in the future (%v)", expiresAt)
	} else {
		__antithesis_instrumentation__.Notify(187297)
	}
	__antithesis_instrumentation__.Notify(187286)
	if user.Normalized() != payload.User {
		__antithesis_instrumentation__.Notify(187298)
		return errors.Errorf("token is for the wrong user %q, wanted %q", payload.User, user)
	} else {
		__antithesis_instrumentation__.Notify(187299)
	}
	__antithesis_instrumentation__.Notify(187287)
	return nil
}
