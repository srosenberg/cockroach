package cliccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/cli/democluster"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

const licenseDefaultURL = "https://register.cockroachdb.com/api/license"

var licenseURL = envutil.EnvOrDefaultString("COCKROACH_DEMO_LICENSE_URL", licenseDefaultURL)

func getLicense(clusterID uuid.UUID) (string, error) {
	__antithesis_instrumentation__.Notify(19410)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", licenseURL, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(19415)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(19416)
	}
	__antithesis_instrumentation__.Notify(19411)

	q := req.URL.Query()

	q.Add("kind", "demo")
	q.Add("version", build.BinaryVersionPrefix())
	q.Add("clusterid", clusterID.String())
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		__antithesis_instrumentation__.Notify(19417)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(19418)
	}
	__antithesis_instrumentation__.Notify(19412)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		__antithesis_instrumentation__.Notify(19419)
		return "", errors.New("unable to connect to licensing endpoint")
	} else {
		__antithesis_instrumentation__.Notify(19420)
	}
	__antithesis_instrumentation__.Notify(19413)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		__antithesis_instrumentation__.Notify(19421)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(19422)
	}
	__antithesis_instrumentation__.Notify(19414)
	return string(bodyBytes), nil
}

func getAndApplyLicense(db *gosql.DB, clusterID uuid.UUID, org string) (bool, error) {
	__antithesis_instrumentation__.Notify(19423)
	license, err := getLicense(clusterID)
	if err != nil {
		__antithesis_instrumentation__.Notify(19427)
		fmt.Fprintf(log.OrigStderr, "\nerror while contacting licensing server:\n%+v\n", err)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(19428)
	}
	__antithesis_instrumentation__.Notify(19424)
	if _, err := db.Exec(`SET CLUSTER SETTING cluster.organization = $1`, org); err != nil {
		__antithesis_instrumentation__.Notify(19429)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(19430)
	}
	__antithesis_instrumentation__.Notify(19425)
	if _, err := db.Exec(`SET CLUSTER SETTING enterprise.license = $1`, license); err != nil {
		__antithesis_instrumentation__.Notify(19431)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(19432)
	}
	__antithesis_instrumentation__.Notify(19426)
	return true, nil
}

func init() {

	democluster.GetAndApplyLicense = getAndApplyLicense
}
