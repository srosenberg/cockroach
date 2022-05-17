package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(21504)

	var x [1]struct{}
	_ = x[codeAuthFailed-1]
	_ = x[codeBackendReadFailed-2]
	_ = x[codeBackendWriteFailed-3]
	_ = x[codeClientReadFailed-4]
	_ = x[codeClientWriteFailed-5]
	_ = x[codeUnexpectedInsecureStartupMessage-6]
	_ = x[codeSNIRoutingFailed-7]
	_ = x[codeUnexpectedStartupMessage-8]
	_ = x[codeParamsRoutingFailed-9]
	_ = x[codeBackendDown-10]
	_ = x[codeBackendRefusedTLS-11]
	_ = x[codeBackendDisconnected-12]
	_ = x[codeClientDisconnected-13]
	_ = x[codeProxyRefusedConnection-14]
	_ = x[codeExpiredClientConnection-15]
	_ = x[codeIdleDisconnect-16]
	_ = x[codeUnavailable-17]
}

const _errorCode_name = "codeAuthFailedcodeBackendReadFailedcodeBackendWriteFailedcodeClientReadFailedcodeClientWriteFailedcodeUnexpectedInsecureStartupMessagecodeSNIRoutingFailedcodeUnexpectedStartupMessagecodeParamsRoutingFailedcodeBackendDowncodeBackendRefusedTLScodeBackendDisconnectedcodeClientDisconnectedcodeProxyRefusedConnectioncodeExpiredClientConnectioncodeIdleDisconnectcodeUnavailable"

var _errorCode_index = [...]uint16{0, 14, 35, 57, 77, 98, 134, 154, 182, 205, 220, 241, 264, 286, 312, 339, 357, 372}

func (i errorCode) String() string {
	__antithesis_instrumentation__.Notify(21505)
	i -= 1
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(21507)
		return i >= errorCode(len(_errorCode_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(21508)
		return "errorCode(" + strconv.FormatInt(int64(i+1), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(21509)
	}
	__antithesis_instrumentation__.Notify(21506)
	return _errorCode_name[_errorCode_index[i]:_errorCode_index[i+1]]
}
