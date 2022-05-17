package security

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash/crc32"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type joinTokenVersion byte

const (
	joinTokenV0 joinTokenVersion = '0'
)

var errInvalidJoinToken = errors.New("invalid join token")

const (
	joinTokenSecretLen = 16

	JoinTokenExpiration = 30 * time.Minute
)

type JoinToken struct {
	TokenID      uuid.UUID
	SharedSecret []byte
	fingerprint  []byte
}

func GenerateJoinToken(cm *CertificateManager) (JoinToken, error) {
	__antithesis_instrumentation__.Notify(186631)
	var jt JoinToken

	jt.TokenID = uuid.MakeV4()
	r := rand.New(rand.NewSource(timeutil.Now().UnixNano()))
	jt.SharedSecret = randutil.RandBytes(r, joinTokenSecretLen)
	jt.sign(cm.CACert().FileContents)
	return jt, nil
}

func (j *JoinToken) sign(caCert []byte) {
	__antithesis_instrumentation__.Notify(186632)
	signer := hmac.New(sha256.New, j.SharedSecret)
	_, _ = signer.Write(caCert)
	j.fingerprint = signer.Sum(nil)
}

func (j *JoinToken) VerifySignature(caCert []byte) bool {
	__antithesis_instrumentation__.Notify(186633)
	signer := hmac.New(sha256.New, j.SharedSecret)
	_, _ = signer.Write(caCert)
	return hmac.Equal(signer.Sum(nil), j.fingerprint)
}

func (j *JoinToken) UnmarshalText(text []byte) error {
	__antithesis_instrumentation__.Notify(186634)

	switch v := joinTokenVersion(text[0]); v {
	case joinTokenV0:
		__antithesis_instrumentation__.Notify(186636)
		decoder := base64.NewDecoder(base64.URLEncoding, bytes.NewReader(text[1:]))
		decoded, err := ioutil.ReadAll(decoder)
		if err != nil {
			__antithesis_instrumentation__.Notify(186643)
			return err
		} else {
			__antithesis_instrumentation__.Notify(186644)
		}
		__antithesis_instrumentation__.Notify(186637)
		if len(decoded) <= uuid.Size+joinTokenSecretLen+4 {
			__antithesis_instrumentation__.Notify(186645)
			return errInvalidJoinToken
		} else {
			__antithesis_instrumentation__.Notify(186646)
		}
		__antithesis_instrumentation__.Notify(186638)
		expectedCSum := crc32.ChecksumIEEE(decoded[:len(decoded)-4])
		_, cSum, err := encoding.DecodeUint32Ascending(decoded[len(decoded)-4:])
		if err != nil {
			__antithesis_instrumentation__.Notify(186647)
			return err
		} else {
			__antithesis_instrumentation__.Notify(186648)
		}
		__antithesis_instrumentation__.Notify(186639)
		if cSum != expectedCSum {
			__antithesis_instrumentation__.Notify(186649)
			return errInvalidJoinToken
		} else {
			__antithesis_instrumentation__.Notify(186650)
		}
		__antithesis_instrumentation__.Notify(186640)
		if err := j.TokenID.UnmarshalBinary(decoded[:uuid.Size]); err != nil {
			__antithesis_instrumentation__.Notify(186651)
			return err
		} else {
			__antithesis_instrumentation__.Notify(186652)
		}
		__antithesis_instrumentation__.Notify(186641)
		decoded = decoded[uuid.Size:]
		j.SharedSecret = decoded[:joinTokenSecretLen]
		j.fingerprint = decoded[joinTokenSecretLen : len(decoded)-4]
	default:
		__antithesis_instrumentation__.Notify(186642)
		return errInvalidJoinToken
	}
	__antithesis_instrumentation__.Notify(186635)

	return nil
}

func (j *JoinToken) MarshalText() ([]byte, error) {
	__antithesis_instrumentation__.Notify(186653)
	tokenID, err := j.TokenID.MarshalBinary()
	if err != nil {
		__antithesis_instrumentation__.Notify(186658)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186659)
	}
	__antithesis_instrumentation__.Notify(186654)

	if len(j.SharedSecret) != joinTokenSecretLen {
		__antithesis_instrumentation__.Notify(186660)
		return nil, errors.New("join token shared secret not of the right size")
	} else {
		__antithesis_instrumentation__.Notify(186661)
	}
	__antithesis_instrumentation__.Notify(186655)

	var b bytes.Buffer
	token := make([]byte, 0, len(tokenID)+len(j.SharedSecret)+len(j.fingerprint)+4)
	token = append(token, tokenID...)
	token = append(token, j.SharedSecret...)
	token = append(token, j.fingerprint...)

	cSum := crc32.ChecksumIEEE(token)
	token = encoding.EncodeUint32Ascending(token, cSum)
	encoder := base64.NewEncoder(base64.URLEncoding, &b)
	if _, err := encoder.Write(token); err != nil {
		__antithesis_instrumentation__.Notify(186662)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186663)
	}
	__antithesis_instrumentation__.Notify(186656)
	if err := encoder.Close(); err != nil {
		__antithesis_instrumentation__.Notify(186664)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(186665)
	}
	__antithesis_instrumentation__.Notify(186657)

	versionedToken := make([]byte, 0, len(b.Bytes())+1)
	versionedToken = append(versionedToken, byte(joinTokenV0))
	versionedToken = append(versionedToken, b.Bytes()...)
	return versionedToken, nil
}
