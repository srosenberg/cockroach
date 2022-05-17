package serverutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/testutils"
)

func SetClusterSetting(t testutils.TB, c TestClusterInterface, name string, value interface{}) {
	__antithesis_instrumentation__.Notify(645916)
	t.Helper()
	strVal := func() string {
		__antithesis_instrumentation__.Notify(645918)
		switch v := value.(type) {
		case string:
			__antithesis_instrumentation__.Notify(645919)
			return v
		case int, int32, int64:
			__antithesis_instrumentation__.Notify(645920)
			return fmt.Sprintf("%d", v)
		case bool:
			__antithesis_instrumentation__.Notify(645921)
			return strconv.FormatBool(v)
		case float32, float64:
			__antithesis_instrumentation__.Notify(645922)
			return fmt.Sprintf("%f", v)
		case fmt.Stringer:
			__antithesis_instrumentation__.Notify(645923)
			return v.String()
		default:
			__antithesis_instrumentation__.Notify(645924)
			return fmt.Sprintf("%v", value)
		}
	}()
	__antithesis_instrumentation__.Notify(645917)
	query := fmt.Sprintf("SET CLUSTER SETTING %s='%s'", name, strVal)

	for i := 0; i < c.NumServers(); i++ {
		__antithesis_instrumentation__.Notify(645925)
		_, err := c.ServerConn(i).ExecContext(context.Background(), query)
		if err != nil {
			__antithesis_instrumentation__.Notify(645926)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(645927)
		}
	}
}
