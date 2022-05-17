package aws

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
)

const sshPublicKeyFile = "${HOME}/.ssh/id_rsa.pub"

func (p *Provider) sshKeyExists(keyName, region string) (bool, error) {
	__antithesis_instrumentation__.Notify(182998)
	var data struct {
		KeyPairs []struct {
			KeyName string
		}
	}
	args := []string{
		"ec2", "describe-key-pairs",
		"--region", region,
	}
	err := p.runJSONCommand(args, &data)
	if err != nil {
		__antithesis_instrumentation__.Notify(183001)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(183002)
	}
	__antithesis_instrumentation__.Notify(182999)
	for _, keyPair := range data.KeyPairs {
		__antithesis_instrumentation__.Notify(183003)
		if keyPair.KeyName == keyName {
			__antithesis_instrumentation__.Notify(183004)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(183005)
		}
	}
	__antithesis_instrumentation__.Notify(183000)
	return false, nil
}

func (p *Provider) sshKeyImport(keyName, region string) error {
	__antithesis_instrumentation__.Notify(183006)
	_, err := os.Stat(os.ExpandEnv(sshPublicKeyFile))
	if err != nil {
		__antithesis_instrumentation__.Notify(183010)
		if oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(183012)
			return errors.Wrapf(err, "please run ssh-keygen externally to create your %s file", sshPublicKeyFile)
		} else {
			__antithesis_instrumentation__.Notify(183013)
		}
		__antithesis_instrumentation__.Notify(183011)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183014)
	}
	__antithesis_instrumentation__.Notify(183007)

	var data struct {
		KeyName string
	}
	_ = data.KeyName

	user, err := p.FindActiveAccount()
	if err != nil {
		__antithesis_instrumentation__.Notify(183015)
		return err
	} else {
		__antithesis_instrumentation__.Notify(183016)
	}
	__antithesis_instrumentation__.Notify(183008)

	timestamp := timeutil.Now()
	createdAt := timestamp.Format(time.RFC3339)

	IAMUserNameTag := fmt.Sprintf("{Key=IAMUserName,Value=%s}", user)
	createdAtTag := fmt.Sprintf("{Key=CreatedAt,Value=%s}", createdAt)
	tagSpecs := fmt.Sprintf("ResourceType=key-pair,Tags=[%s, %s]", IAMUserNameTag, createdAtTag)

	args := []string{
		"ec2", "import-key-pair",
		"--region", region,
		"--key-name", keyName,
		"--public-key-material", fmt.Sprintf("fileb://%s", sshPublicKeyFile),
		"--tag-specifications", tagSpecs,
	}
	err = p.runJSONCommand(args, &data)

	if err == nil || func() bool {
		__antithesis_instrumentation__.Notify(183017)
		return strings.Contains(err.Error(), "InvalidKeyPair.Duplicate") == true
	}() == true {
		__antithesis_instrumentation__.Notify(183018)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(183019)
	}
	__antithesis_instrumentation__.Notify(183009)
	return err
}

func (p *Provider) sshKeyName() (string, error) {
	__antithesis_instrumentation__.Notify(183020)
	user, err := p.FindActiveAccount()
	if err != nil {
		__antithesis_instrumentation__.Notify(183024)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183025)
	}
	__antithesis_instrumentation__.Notify(183021)

	keyBytes, err := ioutil.ReadFile(os.ExpandEnv(sshPublicKeyFile))
	if err != nil {
		__antithesis_instrumentation__.Notify(183026)
		if oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(183028)
			return "", errors.Wrapf(err, "please run ssh-keygen externally to create your %s file", sshPublicKeyFile)
		} else {
			__antithesis_instrumentation__.Notify(183029)
		}
		__antithesis_instrumentation__.Notify(183027)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183030)
	}
	__antithesis_instrumentation__.Notify(183022)

	hash := sha1.New()
	if _, err := hash.Write(keyBytes); err != nil {
		__antithesis_instrumentation__.Notify(183031)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(183032)
	}
	__antithesis_instrumentation__.Notify(183023)
	hashBytes := hash.Sum(nil)
	hashText := base64.URLEncoding.EncodeToString(hashBytes)

	return fmt.Sprintf("%s-%s", user, hashText), nil
}
