package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo/geos"
)

var (
	APIKey string
)

func init() {
	APIKey = os.Getenv("GEOVIZ_GOOGLE_MAPS_API_KEY")
}

type indexTemplate struct {
	APIKey string
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(40654)
	pkg, err := build.Import("github.com/cockroachdb/cockroach", "", build.FindOnly)
	if err != nil {
		__antithesis_instrumentation__.Notify(40656)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(40657)
	}
	__antithesis_instrumentation__.Notify(40655)
	templates := template.Must(template.ParseFiles(filepath.Join(pkg.Dir, "pkg/cmd/geoviz/templates/index.tmpl.html")))
	if err := templates.ExecuteTemplate(
		w,
		"index.tmpl.html",
		indexTemplate{
			APIKey: APIKey,
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(40658)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		__antithesis_instrumentation__.Notify(40659)
	}
}

func handleLoad(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(40660)
	err := r.ParseForm()
	if err != nil {
		__antithesis_instrumentation__.Notify(40664)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(40665)
	}
	__antithesis_instrumentation__.Notify(40661)

	gviz, err := ImageFromReader(strings.NewReader(r.Form["data"][0]))
	if err != nil {
		__antithesis_instrumentation__.Notify(40666)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(40667)
	}
	__antithesis_instrumentation__.Notify(40662)

	ret, err := json.Marshal(gviz)
	if err != nil {
		__antithesis_instrumentation__.Notify(40668)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(40669)
	}
	__antithesis_instrumentation__.Notify(40663)

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(ret); err != nil {
		__antithesis_instrumentation__.Notify(40670)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(40671)
	}
}

var flagGeoLibsDir = flag.String(
	"geo_libs",
	"/usr/local/lib/cockroach",
	"Location where spatial related libraries can be found.",
)

func main() {
	__antithesis_instrumentation__.Notify(40672)
	flag.Parse()

	if _, err := geos.EnsureInit(geos.EnsureInitErrorDisplayPrivate, *flagGeoLibsDir); err != nil {
		__antithesis_instrumentation__.Notify(40674)
		log.Fatalf("could not initialize GEOS - spatial functions may not be available: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(40675)
	}
	__antithesis_instrumentation__.Notify(40673)

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/load", handleLoad)

	fmt.Printf("running server...\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
