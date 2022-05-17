package amazon

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/errors"
)

const awsScheme = "aws"

type awsKMS struct {
	kms                 *kms.KMS
	customerMasterKeyID string
}

var _ cloud.KMS = &awsKMS{}

func init() {
	cloud.RegisterKMSFromURIFactory(MakeAWSKMS, awsScheme)
}

type kmsURIParams struct {
	accessKey string
	secret    string
	tempToken string
	endpoint  string
	region    string
	auth      string
}

func resolveKMSURIParams(kmsURI url.URL) kmsURIParams {
	__antithesis_instrumentation__.Notify(35688)
	params := kmsURIParams{
		accessKey: kmsURI.Query().Get(AWSAccessKeyParam),
		secret:    kmsURI.Query().Get(AWSSecretParam),
		tempToken: kmsURI.Query().Get(AWSTempTokenParam),
		endpoint:  kmsURI.Query().Get(AWSEndpointParam),
		region:    kmsURI.Query().Get(KMSRegionParam),
		auth:      kmsURI.Query().Get(cloud.AuthParam),
	}

	params.secret = strings.Replace(params.secret, " ", "+", -1)
	return params
}

func MakeAWSKMS(uri string, env cloud.KMSEnv) (cloud.KMS, error) {
	__antithesis_instrumentation__.Notify(35689)
	if env.KMSConfig().DisableOutbound {
		__antithesis_instrumentation__.Notify(35696)
		return nil, errors.New("external IO must be enabled to use AWS KMS")
	} else {
		__antithesis_instrumentation__.Notify(35697)
	}
	__antithesis_instrumentation__.Notify(35690)
	kmsURI, err := url.ParseRequestURI(uri)
	if err != nil {
		__antithesis_instrumentation__.Notify(35698)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35699)
	}
	__antithesis_instrumentation__.Notify(35691)

	kmsURIParams := resolveKMSURIParams(*kmsURI)
	region := kmsURIParams.region
	awsConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials(kmsURIParams.accessKey,
			kmsURIParams.secret, kmsURIParams.tempToken),
	}
	if kmsURIParams.endpoint != "" {
		__antithesis_instrumentation__.Notify(35700)
		if env.KMSConfig().DisableHTTP {
			__antithesis_instrumentation__.Notify(35704)
			return nil, errors.New(
				"custom endpoints disallowed for aws kms due to --aws-kms-disable-http flag")
		} else {
			__antithesis_instrumentation__.Notify(35705)
		}
		__antithesis_instrumentation__.Notify(35701)
		awsConfig.Endpoint = &kmsURIParams.endpoint
		if region == "" {
			__antithesis_instrumentation__.Notify(35706)

			region = "default-region"
		} else {
			__antithesis_instrumentation__.Notify(35707)
		}
		__antithesis_instrumentation__.Notify(35702)
		client, err := cloud.MakeHTTPClient(env.ClusterSettings())
		if err != nil {
			__antithesis_instrumentation__.Notify(35708)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(35709)
		}
		__antithesis_instrumentation__.Notify(35703)
		awsConfig.HTTPClient = client
	} else {
		__antithesis_instrumentation__.Notify(35710)
	}
	__antithesis_instrumentation__.Notify(35692)

	opts := session.Options{}
	switch kmsURIParams.auth {
	case "", cloud.AuthParamSpecified:
		__antithesis_instrumentation__.Notify(35711)
		if kmsURIParams.accessKey == "" {
			__antithesis_instrumentation__.Notify(35717)
			return nil, errors.Errorf(
				"%s is set to '%s', but %s is not set",
				cloud.AuthParam,
				cloud.AuthParamSpecified,
				AWSAccessKeyParam,
			)
		} else {
			__antithesis_instrumentation__.Notify(35718)
		}
		__antithesis_instrumentation__.Notify(35712)
		if kmsURIParams.secret == "" {
			__antithesis_instrumentation__.Notify(35719)
			return nil, errors.Errorf(
				"%s is set to '%s', but %s is not set",
				cloud.AuthParam,
				cloud.AuthParamSpecified,
				AWSSecretParam,
			)
		} else {
			__antithesis_instrumentation__.Notify(35720)
		}
		__antithesis_instrumentation__.Notify(35713)
		opts.Config.MergeIn(awsConfig)
	case cloud.AuthParamImplicit:
		__antithesis_instrumentation__.Notify(35714)
		if env.KMSConfig().DisableImplicitCredentials {
			__antithesis_instrumentation__.Notify(35721)
			return nil, errors.New(
				"implicit credentials disallowed for s3 due to --external-io-implicit-credentials flag")
		} else {
			__antithesis_instrumentation__.Notify(35722)
		}
		__antithesis_instrumentation__.Notify(35715)
		opts.SharedConfigState = session.SharedConfigEnable
	default:
		__antithesis_instrumentation__.Notify(35716)
		return nil, errors.Errorf("unsupported value %s for %s", kmsURIParams.auth, cloud.AuthParam)
	}
	__antithesis_instrumentation__.Notify(35693)

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(35723)
		return nil, errors.Wrap(err, "new aws session")
	} else {
		__antithesis_instrumentation__.Notify(35724)
	}
	__antithesis_instrumentation__.Notify(35694)
	if region == "" {
		__antithesis_instrumentation__.Notify(35725)

		return nil, errors.New("could not find the aws kms region")
	} else {
		__antithesis_instrumentation__.Notify(35726)
	}
	__antithesis_instrumentation__.Notify(35695)
	sess.Config.Region = aws.String(region)
	return &awsKMS{
		kms:                 kms.New(sess),
		customerMasterKeyID: strings.TrimPrefix(kmsURI.Path, "/"),
	}, nil
}

func (k *awsKMS) MasterKeyID() (string, error) {
	__antithesis_instrumentation__.Notify(35727)
	return k.customerMasterKeyID, nil
}

func (k *awsKMS) Encrypt(ctx context.Context, data []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(35728)
	encryptInput := &kms.EncryptInput{
		KeyId:     &k.customerMasterKeyID,
		Plaintext: data,
	}

	encryptOutput, err := k.kms.Encrypt(encryptInput)
	if err != nil {
		__antithesis_instrumentation__.Notify(35730)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35731)
	}
	__antithesis_instrumentation__.Notify(35729)

	return encryptOutput.CiphertextBlob, nil
}

func (k *awsKMS) Decrypt(ctx context.Context, data []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(35732)
	decryptInput := &kms.DecryptInput{
		KeyId:          &k.customerMasterKeyID,
		CiphertextBlob: data,
	}

	decryptOutput, err := k.kms.Decrypt(decryptInput)
	if err != nil {
		__antithesis_instrumentation__.Notify(35734)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35735)
	}
	__antithesis_instrumentation__.Notify(35733)

	return decryptOutput.Plaintext, nil
}

func (k *awsKMS) Close() error {
	__antithesis_instrumentation__.Notify(35736)
	return nil
}
