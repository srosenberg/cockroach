package main

import (
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv/kvnemesis"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/testcluster"
	"github.com/cockroachdb/cockroach/pkg/security/securityassets"
	"github.com/cockroachdb/cockroach/pkg/security/securitytest"
	"unsafe"
	"fmt"
)

// #include <stdint.h>
import "C"

//export LLVMFuzzerTestOneInput
func LLVMFuzzerTestOneInput(data *C.char, size C.size_t) C.int {
	if (size < 3000) {
		return -1
	}
        // TODO(mdempsky): Use unsafe.Slice once golang.org/issue/19367 is accepted.
        s := (*[1<<20]byte)(unsafe.Pointer(data))[:size:size]
	fmt.Println(len(s))
        kvnemesis.Fuzz(s)
        return 0
}

func init() {
	securityassets.SetLoader(securitytest.EmbeddedAssets)
        randutil.SeedForTests()
        serverutils.InitTestServerFactory(server.TestServerFactory)
        serverutils.InitTestClusterFactory(testcluster.TestClusterFactory)
	defer serverutils.TestingSetDefaultTenantSelectionOverride(
		// All the tests in this package are specific to the storage layer.
		base.TestIsSpecificToStorageLayerAndNeedsASystemTenant,
	)()
}


func main() {

}
