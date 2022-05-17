package docs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/build"
)

var URLBase = "https://www.cockroachlabs.com/docs/" + build.BinaryVersionPrefix()

var URLReleaseNotesBase = fmt.Sprintf("https://www.cockroachlabs.com/docs/releases/%s.0.html",
	build.BinaryVersionPrefix())

func URL(pageName string) string {
	__antithesis_instrumentation__.Notify(58471)
	return URLBase + "/" + pageName
}

func ReleaseNotesURL(pageName string) string {
	__antithesis_instrumentation__.Notify(58472)
	return URLReleaseNotesBase + pageName
}
