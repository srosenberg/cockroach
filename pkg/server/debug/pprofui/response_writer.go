package pprofui

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io"
	"net/http"
)

type responseBridge struct {
	target     io.Writer
	statusCode int
}

var _ http.ResponseWriter = &responseBridge{}

func (r *responseBridge) Header() http.Header {
	__antithesis_instrumentation__.Notify(190330)
	return http.Header{}
}

func (r *responseBridge) Write(b []byte) (int, error) {
	__antithesis_instrumentation__.Notify(190331)
	return r.target.Write(b)
}

func (r *responseBridge) WriteHeader(statusCode int) {
	__antithesis_instrumentation__.Notify(190332)
	r.statusCode = statusCode
}
