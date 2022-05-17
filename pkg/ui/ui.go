// Package ui embeds the assets for the web UI into the Cockroach binary.
//
// By default, it serves a stub web UI. Linking with distoss or distccl will
// replace the stubs with the OSS UI or the CCL UI, respectively. The exported
// symbols in this package are thus function pointers instead of functions so
// that they can be mutated by init hooks.
package ui

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

var Assets fs.FS

var HaveUI = false

var indexHTMLTemplate = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
	<head>
		<title>Cockroach Console</title>
		<meta charset="UTF-8">
		<link href="favicon.ico" rel="shortcut icon">
	</head>
	<body>
		<div id="react-layout"></div>

		<script>
			window.dataFromServer = {{.}};
		</script>

		<script src="bundle.js" type="text/javascript"></script>
	</body>
</html>
`))

type indexHTMLArgs struct {
	ExperimentalUseLogin bool
	LoginEnabled         bool
	LoggedInUser         *string
	Tag                  string
	Version              string
	NodeID               string
	OIDCAutoLogin        bool
	OIDCLoginEnabled     bool
	OIDCButtonText       string
}

type OIDCUIConf struct {
	ButtonText string
	AutoLogin  bool
	Enabled    bool
}

type OIDCUI interface {
	GetOIDCConf() OIDCUIConf
}

var bareIndexHTML = []byte(fmt.Sprintf(`<!DOCTYPE html>
<title>CockroachDB</title>
Binary built without web UI.
<hr>
<em>%s</em>`, build.GetInfo().Short()))

type Config struct {
	ExperimentalUseLogin bool
	LoginEnabled         bool
	NodeID               *base.NodeIDContainer
	GetUser              func(ctx context.Context) *string
	OIDC                 OIDCUI
}

func Handler(cfg Config) http.Handler {
	__antithesis_instrumentation__.Notify(650206)

	etags := make(map[string]string)

	if HaveUI && func() bool {
		__antithesis_instrumentation__.Notify(650208)
		return Assets != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(650209)

		err := httputil.ComputeEtags(Assets, etags)
		if err != nil {
			__antithesis_instrumentation__.Notify(650210)
			log.Errorf(context.Background(), "Unable to compute asset hashes: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(650211)
		}
	} else {
		__antithesis_instrumentation__.Notify(650212)
	}
	__antithesis_instrumentation__.Notify(650207)

	fileHandlerChain := httputil.EtagHandler(
		etags,
		http.FileServer(
			http.FS(Assets),
		),
	)
	buildInfo := build.GetInfo()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		__antithesis_instrumentation__.Notify(650213)
		if !HaveUI {
			__antithesis_instrumentation__.Notify(650216)
			http.ServeContent(w, r, "index.html", buildInfo.GoTime(), bytes.NewReader(bareIndexHTML))
			return
		} else {
			__antithesis_instrumentation__.Notify(650217)
		}
		__antithesis_instrumentation__.Notify(650214)

		if r.URL.Path != "/" {
			__antithesis_instrumentation__.Notify(650218)
			fileHandlerChain.ServeHTTP(w, r)
			return
		} else {
			__antithesis_instrumentation__.Notify(650219)
		}
		__antithesis_instrumentation__.Notify(650215)

		oidcConf := cfg.OIDC.GetOIDCConf()

		if err := indexHTMLTemplate.Execute(w, indexHTMLArgs{
			ExperimentalUseLogin: cfg.ExperimentalUseLogin,
			LoginEnabled:         cfg.LoginEnabled,
			LoggedInUser:         cfg.GetUser(r.Context()),
			Tag:                  buildInfo.Tag,
			Version:              build.BinaryVersionPrefix(),
			NodeID:               cfg.NodeID.String(),
			OIDCAutoLogin:        oidcConf.AutoLogin,
			OIDCLoginEnabled:     oidcConf.Enabled,
			OIDCButtonText:       oidcConf.ButtonText,
		}); err != nil {
			__antithesis_instrumentation__.Notify(650220)
			err = errors.Wrap(err, "templating index.html")
			http.Error(w, err.Error(), 500)
			log.Errorf(r.Context(), "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(650221)
		}
	})
}
