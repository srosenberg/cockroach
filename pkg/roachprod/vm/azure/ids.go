package azure

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"regexp"

	"github.com/cockroachdb/errors"
)

var azureIDPattern = regexp.MustCompile(
	"/subscriptions/(.+)/resourceGroups/(.+)/providers/(.+?)/(.+?)/(.+)")

type azureID struct {
	provider      string
	resourceGroup string
	resourceName  string
	resourceType  string
	subscription  string
}

func (id azureID) String() string {
	__antithesis_instrumentation__.Notify(183547)
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s/%s/%s",
		id.subscription, id.resourceGroup, id.provider, id.resourceType, id.resourceName)
}

func parseAzureID(id string) (azureID, error) {
	__antithesis_instrumentation__.Notify(183548)
	parts := azureIDPattern.FindStringSubmatch(id)
	if len(parts) == 0 {
		__antithesis_instrumentation__.Notify(183550)
		return azureID{}, errors.Errorf("could not parse Azure ID %q", id)
	} else {
		__antithesis_instrumentation__.Notify(183551)
	}
	__antithesis_instrumentation__.Notify(183549)
	ret := azureID{
		subscription:  parts[1],
		resourceGroup: parts[2],
		provider:      parts[3],
		resourceType:  parts[4],
		resourceName:  parts[5],
	}
	return ret, nil
}
