package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/http"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errAPIInternalErrorString = "An internal server error has occurred. Please check your CockroachDB logs for more details."

var errAPIInternalError = status.Errorf(
	codes.Internal,
	errAPIInternalErrorString,
)

func apiInternalError(ctx context.Context, err error) error {
	__antithesis_instrumentation__.Notify(189048)
	log.ErrorfDepth(ctx, 1, "%s", err)
	return errAPIInternalError
}

func apiV2InternalError(ctx context.Context, err error, w http.ResponseWriter) {
	__antithesis_instrumentation__.Notify(189049)
	log.ErrorfDepth(ctx, 1, "%s", err)
	http.Error(w, errAPIInternalErrorString, http.StatusInternalServerError)
}
