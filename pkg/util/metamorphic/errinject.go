// Copyright 2024 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package metamorphic

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	injectedBlocks = make([]uint32, 1024)
	availableErrs  [21]error
	genericErr     = fmt.Errorf("injected error")
)

const (
	GRPCTimeoutErr             = iota // grpc_util.IsTimeout
	GRPCUnavailableErr                // grpc_util.IsConnectionUnavailable
	GRPCRejectedErr                   // grpc_util.IsConnectionRejected
	GRPCClosedErr                     // grpc_util.IsClosedConnection
	GRPCtxCancelledErr                // grpc_util.IsContextCanceled
	GRPCAuthErr                       // grpc_util.IsAuthError
	GRPCWaitingForInitErr             // grpc_util.IsWaitingForInit
	GRPCRequestDidNotStartErr         // grpc_util.RequestDidNotStart
	GRPCPermissionDeniedErr           // status.Code(err) == codes.PermissionDenied
	GRPCNotFoundErr                   // status.Code(err) == codes.NotFound
	GRPCAlreadyExistsErr              // status.Code(err) == codes.AlreadyExists
	PGOutOfMemoryErr                  // sqlerrors.IsOutOfMemoryError
	PGRelationAlreadyExistsErr        // sqlerrors.IsRelationAlreadyExistsError
	PGUndefinedObjectErr              // pgerror.GetPGCode(err) == pgcode.UndefinedObject
	PGUndefinedFunctionErr            // pgerror.GetPGCode(err) == pgcode.UndefinedFunction
	PGUniqueViolationErr              // pgerror.GetPGCode(err) == pgcode.UniqueViolation
	PGUndefinedColumnErr              // pgerror.GetPGCode(err) == pgcode.UndefinedColumn
	PGInvalidSchemaNameErr            // pgerror.GetPGCode(err) == pgcode.InvalidSchemaName
	PGDuplicateObjectErr              // pgerror.GetPGCode(err) == pgcode.DuplicateObject
	PGUndefinedTableErr               // pgerror.GetPGCode(err) == pgcode.UndefinedTable
	PGInvalidParameterValueErr        // pgerror.GetPGCode(err) == pgcode.InvalidParameterValue
)

func init() {
	availableErrs[GRPCTimeoutErr] = status.Error(codes.DeadlineExceeded, "injected error")
	availableErrs[GRPCUnavailableErr] = status.Error(codes.FailedPrecondition, "injected error")
	availableErrs[GRPCRejectedErr] = status.Error(codes.FailedPrecondition, "injected error")
	availableErrs[GRPCClosedErr] = status.Error(codes.Canceled, "injected error")
	availableErrs[GRPCtxCancelledErr] = status.Error(codes.Canceled, context.Canceled.Error())
	availableErrs[GRPCAuthErr] = status.Error(codes.Unauthenticated, "injected error")
	availableErrs[GRPCWaitingForInitErr] = status.Error(codes.Unavailable, "node waiting for init")
	availableErrs[GRPCRequestDidNotStartErr] = status.Error(codes.Unavailable, "node waiting for init")
	availableErrs[GRPCPermissionDeniedErr] = status.Error(codes.PermissionDenied, "injected error")
	availableErrs[GRPCNotFoundErr] = status.Error(codes.NotFound, "injected error")
	availableErrs[GRPCAlreadyExistsErr] = status.Error(codes.AlreadyExists, "injected error")

	availableErrs[PGOutOfMemoryErr] = pgerror.New(pgcode.OutOfMemory, "injected error")
	availableErrs[PGRelationAlreadyExistsErr] = pgerror.New(pgcode.DuplicateRelation, "injected error")
	availableErrs[PGUndefinedObjectErr] = pgerror.New(pgcode.UndefinedObject, "injected error")
	availableErrs[PGUndefinedFunctionErr] = pgerror.New(pgcode.UndefinedFunction, "injected error")
	availableErrs[PGUniqueViolationErr] = pgerror.New(pgcode.UniqueViolation, "injected error")
	availableErrs[PGUndefinedColumnErr] = pgerror.New(pgcode.UndefinedColumn, "injected error")
	availableErrs[PGInvalidSchemaNameErr] = pgerror.New(pgcode.InvalidSchemaName, "injected error")
	availableErrs[PGDuplicateObjectErr] = pgerror.New(pgcode.DuplicateObject, "injected error")
	availableErrs[PGUndefinedTableErr] = pgerror.New(pgcode.UndefinedTable, "injected error")
	availableErrs[PGInvalidParameterValueErr] = pgerror.New(pgcode.InvalidParameterValue, "injected error")
}

func InjectErr(probability float64, blockId uint32, err *error, errTypes []int) bool {
	rng.Lock()
	defer rng.Unlock()

	if rng.r.Float64() < probability {
		injectedBlocks = append(injectedBlocks, blockId)
		i := len(errTypes)
		if i > 0 {
			// N.B. we pick any of the specified errTypes _and_ genericErr, hence the +1.
			i = rng.r.Intn(i + 1)
		}
		if i < len(errTypes) {
			*err = availableErrs[errTypes[i]]
		} else {
			*err = genericErr
		}
		return true
	}
	return false
}
