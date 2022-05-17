package oidcccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/hmac"
	crypto_rand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type keyAndSignedToken struct {
	secretKeyCookie    *http.Cookie
	signedTokenEncoded string
}

func newKeyAndSignedToken(keySize int, tokenSize int) (*keyAndSignedToken, error) {
	__antithesis_instrumentation__.Notify(20610)
	secretKey := make([]byte, keySize)
	if _, err := crypto_rand.Read(secretKey); err != nil {
		__antithesis_instrumentation__.Notify(20615)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(20616)
	}
	__antithesis_instrumentation__.Notify(20611)

	token := make([]byte, tokenSize)
	if _, err := crypto_rand.Read(token); err != nil {
		__antithesis_instrumentation__.Notify(20617)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(20618)
	}
	__antithesis_instrumentation__.Notify(20612)

	mac := hmac.New(sha256.New, secretKey)
	_, err := mac.Write(token)
	if err != nil {
		__antithesis_instrumentation__.Notify(20619)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(20620)
	}
	__antithesis_instrumentation__.Notify(20613)

	signedTokenEncoded, err := encodeOIDCState(serverpb.OIDCState{
		Token:    token,
		TokenMAC: mac.Sum(nil),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(20621)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(20622)
	}
	__antithesis_instrumentation__.Notify(20614)

	secretKeyCookie := http.Cookie{
		Name:     secretCookieName,
		Value:    base64.URLEncoding.EncodeToString(secretKey),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	return &keyAndSignedToken{
		&secretKeyCookie,
		signedTokenEncoded,
	}, nil
}

func (kast *keyAndSignedToken) validate() (bool, error) {
	__antithesis_instrumentation__.Notify(20623)
	key, err := base64.URLEncoding.DecodeString(kast.secretKeyCookie.Value)
	if err != nil {
		__antithesis_instrumentation__.Notify(20627)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(20628)
	}
	__antithesis_instrumentation__.Notify(20624)
	mac := hmac.New(sha256.New, key)

	signedToken, err := decodeOIDCState(kast.signedTokenEncoded)
	if err != nil {
		__antithesis_instrumentation__.Notify(20629)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(20630)
	}
	__antithesis_instrumentation__.Notify(20625)

	_, err = mac.Write(signedToken.Token)
	if err != nil {
		__antithesis_instrumentation__.Notify(20631)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(20632)
	}
	__antithesis_instrumentation__.Notify(20626)

	return hmac.Equal(signedToken.TokenMAC, mac.Sum(nil)), nil
}

func encodeOIDCState(statePb serverpb.OIDCState) (string, error) {
	__antithesis_instrumentation__.Notify(20633)
	stateBytes, err := protoutil.Marshal(&statePb)
	if err != nil {
		__antithesis_instrumentation__.Notify(20635)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(20636)
	}
	__antithesis_instrumentation__.Notify(20634)
	return base64.URLEncoding.EncodeToString(stateBytes), nil
}

func decodeOIDCState(encodedState string) (*serverpb.OIDCState, error) {
	__antithesis_instrumentation__.Notify(20637)

	stateBytes, err := base64.URLEncoding.DecodeString(encodedState)
	if err != nil {
		__antithesis_instrumentation__.Notify(20640)
		return nil, errors.Wrap(err, "state could not be decoded")
	} else {
		__antithesis_instrumentation__.Notify(20641)
	}
	__antithesis_instrumentation__.Notify(20638)
	var stateValue serverpb.OIDCState
	if err := protoutil.Unmarshal(stateBytes, &stateValue); err != nil {
		__antithesis_instrumentation__.Notify(20642)
		return nil, errors.Wrap(err, "state could not be unmarshaled")
	} else {
		__antithesis_instrumentation__.Notify(20643)
	}
	__antithesis_instrumentation__.Notify(20639)
	return &stateValue, nil
}
