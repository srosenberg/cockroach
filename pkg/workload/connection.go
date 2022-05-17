package workload

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
)

type ConnFlags struct {
	*pflag.FlagSet
	DBOverride  string
	Concurrency int

	Method string
}

func NewConnFlags(genFlags *Flags) *ConnFlags {
	__antithesis_instrumentation__.Notify(693897)
	c := &ConnFlags{}
	c.FlagSet = pflag.NewFlagSet(`conn`, pflag.ContinueOnError)
	c.StringVar(&c.DBOverride, `db`, ``,
		`Override for the SQL database to use. If empty, defaults to the generator name`)
	c.IntVar(&c.Concurrency, `concurrency`, 2*runtime.GOMAXPROCS(0),
		`Number of concurrent workers`)
	c.StringVar(&c.Method, `method`, `prepare`, `SQL issue method (prepare, noprepare, simple)`)
	genFlags.AddFlagSet(c.FlagSet)
	if genFlags.Meta == nil {
		__antithesis_instrumentation__.Notify(693899)
		genFlags.Meta = make(map[string]FlagMeta)
	} else {
		__antithesis_instrumentation__.Notify(693900)
	}
	__antithesis_instrumentation__.Notify(693898)
	genFlags.Meta[`db`] = FlagMeta{RuntimeOnly: true}
	genFlags.Meta[`concurrency`] = FlagMeta{RuntimeOnly: true}
	genFlags.Meta[`method`] = FlagMeta{RuntimeOnly: true}
	return c
}

func SanitizeUrls(gen Generator, dbOverride string, urls []string) (string, error) {
	__antithesis_instrumentation__.Notify(693901)
	dbName := gen.Meta().Name
	if dbOverride != `` {
		__antithesis_instrumentation__.Notify(693904)
		dbName = dbOverride
	} else {
		__antithesis_instrumentation__.Notify(693905)
	}
	__antithesis_instrumentation__.Notify(693902)
	for i := range urls {
		__antithesis_instrumentation__.Notify(693906)
		parsed, err := url.Parse(urls[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(693909)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(693910)
		}
		__antithesis_instrumentation__.Notify(693907)
		if d := strings.TrimPrefix(parsed.Path, `/`); d != `` && func() bool {
			__antithesis_instrumentation__.Notify(693911)
			return d != dbName == true
		}() == true {
			__antithesis_instrumentation__.Notify(693912)
			return "", fmt.Errorf(`%s specifies database %q, but database %q is expected`,
				urls[i], d, dbName)
		} else {
			__antithesis_instrumentation__.Notify(693913)
		}
		__antithesis_instrumentation__.Notify(693908)
		parsed.Path = dbName

		q := parsed.Query()
		q.Set("application_name", gen.Meta().Name)
		parsed.RawQuery = q.Encode()

		switch parsed.Scheme {
		case "postgres", "postgresql":
			__antithesis_instrumentation__.Notify(693914)
			urls[i] = parsed.String()
		default:
			__antithesis_instrumentation__.Notify(693915)
			return ``, fmt.Errorf(`unsupported scheme: %s`, parsed.Scheme)
		}
	}
	__antithesis_instrumentation__.Notify(693903)
	return dbName, nil
}
