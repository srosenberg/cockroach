package gcp

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/base64"
	"hash/crc32"
	"net/url"
	"strings"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/errors"
	"google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const gcsScheme = "gs"

type gcsKMS struct {
	kms                 *kms.KeyManagementClient
	customerMasterKeyID string
}

var _ cloud.KMS = &gcsKMS{}

func init() {
	cloud.RegisterKMSFromURIFactory(MakeGCSKMS, gcsScheme)
}

type kmsURIParams struct {
	credentials string
	auth        string
}

func resolveKMSURIParams(kmsURI url.URL) kmsURIParams {
	__antithesis_instrumentation__.Notify(36237)
	params := kmsURIParams{
		credentials: kmsURI.Query().Get(CredentialsParam),
		auth:        kmsURI.Query().Get(cloud.AuthParam),
	}

	return params
}

func MakeGCSKMS(uri string, env cloud.KMSEnv) (cloud.KMS, error) {
	__antithesis_instrumentation__.Notify(36238)
	if env.KMSConfig().DisableOutbound {
		__antithesis_instrumentation__.Notify(36243)
		return nil, errors.New("external IO must be enabled to use GCS KMS")
	} else {
		__antithesis_instrumentation__.Notify(36244)
	}
	__antithesis_instrumentation__.Notify(36239)
	kmsURI, err := url.ParseRequestURI(uri)
	if err != nil {
		__antithesis_instrumentation__.Notify(36245)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36246)
	}
	__antithesis_instrumentation__.Notify(36240)

	kmsURIParams := resolveKMSURIParams(*kmsURI)

	var credentialsOpt []option.ClientOption

	switch kmsURIParams.auth {
	case "", cloud.AuthParamSpecified:
		__antithesis_instrumentation__.Notify(36247)
		if kmsURIParams.credentials == "" {
			__antithesis_instrumentation__.Notify(36252)
			return nil, errors.Errorf(
				"%s is set to '%s', but %s is not set",
				cloud.AuthParam,
				cloud.AuthParamSpecified,
				CredentialsParam,
			)
		} else {
			__antithesis_instrumentation__.Notify(36253)
		}
		__antithesis_instrumentation__.Notify(36248)

		credentialsJSON, err := base64.StdEncoding.DecodeString(kmsURIParams.credentials)
		if err != nil {
			__antithesis_instrumentation__.Notify(36254)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(36255)
		}
		__antithesis_instrumentation__.Notify(36249)

		credentialsOpt = append(credentialsOpt, option.WithCredentialsJSON(credentialsJSON))
	case cloud.AuthParamImplicit:
		__antithesis_instrumentation__.Notify(36250)
		if env.KMSConfig().DisableImplicitCredentials {
			__antithesis_instrumentation__.Notify(36256)
			return nil, errors.New(
				"implicit credentials disallowed for gcs due to --external-io-implicit-credentials flag")
		} else {
			__antithesis_instrumentation__.Notify(36257)
		}

	default:
		__antithesis_instrumentation__.Notify(36251)
		return nil, errors.Errorf("unsupported value %s for %s", kmsURIParams.auth, cloud.AuthParam)
	}
	__antithesis_instrumentation__.Notify(36241)

	ctx := context.Background()

	kmc, err := kms.NewKeyManagementClient(ctx, credentialsOpt...)

	if err != nil {
		__antithesis_instrumentation__.Notify(36258)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36259)
	}
	__antithesis_instrumentation__.Notify(36242)

	cmkID := strings.Split(kmsURI.Path, "/cryptoKeyVersions/")[0]

	return &gcsKMS{
		kms:                 kmc,
		customerMasterKeyID: strings.TrimPrefix(cmkID, "/"),
	}, nil
}

func (k *gcsKMS) MasterKeyID() (string, error) {
	__antithesis_instrumentation__.Notify(36260)
	return k.customerMasterKeyID, nil
}

func (k *gcsKMS) Encrypt(ctx context.Context, data []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(36261)

	crc32c := func(data []byte) uint32 {
		__antithesis_instrumentation__.Notify(36266)
		t := crc32.MakeTable(crc32.Castagnoli)
		return crc32.Checksum(data, t)
	}
	__antithesis_instrumentation__.Notify(36262)
	plaintextCRC32C := crc32c(data)

	encryptInput := &kmspb.EncryptRequest{
		Name:            k.customerMasterKeyID,
		Plaintext:       data,
		PlaintextCrc32C: wrapperspb.Int64(int64(plaintextCRC32C)),
	}

	encryptOutput, err := k.kms.Encrypt(ctx, encryptInput)
	if err != nil {
		__antithesis_instrumentation__.Notify(36267)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36268)
	}
	__antithesis_instrumentation__.Notify(36263)

	if !encryptOutput.VerifiedPlaintextCrc32C {
		__antithesis_instrumentation__.Notify(36269)
		return nil, errors.Errorf("Encrypt: request corrupted in-transit")
	} else {
		__antithesis_instrumentation__.Notify(36270)
	}
	__antithesis_instrumentation__.Notify(36264)
	if int64(crc32c(encryptOutput.Ciphertext)) != encryptOutput.CiphertextCrc32C.Value {
		__antithesis_instrumentation__.Notify(36271)
		return nil, errors.Errorf("Encrypt: response corrupted in-transit")
	} else {
		__antithesis_instrumentation__.Notify(36272)
	}
	__antithesis_instrumentation__.Notify(36265)

	return encryptOutput.Ciphertext, nil
}

func (k *gcsKMS) Decrypt(ctx context.Context, data []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(36273)

	crc32c := func(data []byte) uint32 {
		__antithesis_instrumentation__.Notify(36277)
		t := crc32.MakeTable(crc32.Castagnoli)
		return crc32.Checksum(data, t)
	}
	__antithesis_instrumentation__.Notify(36274)
	ciphertextCRC32C := crc32c(data)

	decryptInput := &kmspb.DecryptRequest{
		Name:             k.customerMasterKeyID,
		Ciphertext:       data,
		CiphertextCrc32C: wrapperspb.Int64(int64(ciphertextCRC32C)),
	}

	decryptOutput, err := k.kms.Decrypt(ctx, decryptInput)
	if err != nil {
		__antithesis_instrumentation__.Notify(36278)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36279)
	}
	__antithesis_instrumentation__.Notify(36275)

	if int64(crc32c(decryptOutput.Plaintext)) != decryptOutput.PlaintextCrc32C.Value {
		__antithesis_instrumentation__.Notify(36280)
		return nil, errors.Errorf("Decrypt: response corrupted in-transit")
	} else {
		__antithesis_instrumentation__.Notify(36281)
	}
	__antithesis_instrumentation__.Notify(36276)

	return decryptOutput.Plaintext, nil
}

func (k *gcsKMS) Close() error {
	__antithesis_instrumentation__.Notify(36282)
	return k.kms.Close()
}
