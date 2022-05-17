//go:build gss
// +build gss

package gssapiccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire"
	"github.com/cockroachdb/errors"
)

// #cgo LDFLAGS: -lgssapi_krb5 -lcom_err -lkrb5 -lkrb5support -ldl -lk5crypto -lresolv
//
// #include <gssapi/gssapi.h>
// #include <stdlib.h>
import "C"

func getGssUser(c pgwire.AuthConn) (connClose func(), gssUser string, _ error) {
	__antithesis_instrumentation__.Notify(19518)
	var (
		majStat, minStat, lminS, gflags C.OM_uint32
		gbuf                            C.gss_buffer_desc
		contextHandle                   C.gss_ctx_id_t  = C.GSS_C_NO_CONTEXT
		acceptorCredHandle              C.gss_cred_id_t = C.GSS_C_NO_CREDENTIAL
		srcName                         C.gss_name_t
		outputToken                     C.gss_buffer_desc
	)

	if err := c.SendAuthRequest(authTypeGSS, nil); err != nil {
		__antithesis_instrumentation__.Notify(19523)
		return nil, "", err
	} else {
		__antithesis_instrumentation__.Notify(19524)
	}
	__antithesis_instrumentation__.Notify(19519)

	connClose = func() {
		__antithesis_instrumentation__.Notify(19525)
		C.gss_delete_sec_context(&lminS, &contextHandle, C.GSS_C_NO_BUFFER)
	}
	__antithesis_instrumentation__.Notify(19520)

	for {
		__antithesis_instrumentation__.Notify(19526)
		token, err := c.GetPwdData()
		if err != nil {
			__antithesis_instrumentation__.Notify(19530)
			return connClose, "", err
		} else {
			__antithesis_instrumentation__.Notify(19531)
		}
		__antithesis_instrumentation__.Notify(19527)

		gbuf.length = C.ulong(len(token))
		gbuf.value = C.CBytes(token)

		majStat = C.gss_accept_sec_context(
			&minStat,
			&contextHandle,
			acceptorCredHandle,
			&gbuf,
			C.GSS_C_NO_CHANNEL_BINDINGS,
			&srcName,
			nil,
			&outputToken,
			&gflags,
			nil,
			nil,
		)
		C.free(gbuf.value)

		if outputToken.length != 0 {
			__antithesis_instrumentation__.Notify(19532)
			outputBytes := C.GoBytes(outputToken.value, C.int(outputToken.length))
			C.gss_release_buffer(&lminS, &outputToken)
			if err := c.SendAuthRequest(authTypeGSSContinue, outputBytes); err != nil {
				__antithesis_instrumentation__.Notify(19533)
				return connClose, "", err
			} else {
				__antithesis_instrumentation__.Notify(19534)
			}
		} else {
			__antithesis_instrumentation__.Notify(19535)
		}
		__antithesis_instrumentation__.Notify(19528)
		if majStat != C.GSS_S_COMPLETE && func() bool {
			__antithesis_instrumentation__.Notify(19536)
			return majStat != C.GSS_S_CONTINUE_NEEDED == true
		}() == true {
			__antithesis_instrumentation__.Notify(19537)
			return connClose, "", gssError("accepting GSS security context failed", majStat, minStat)
		} else {
			__antithesis_instrumentation__.Notify(19538)
		}
		__antithesis_instrumentation__.Notify(19529)
		if majStat != C.GSS_S_CONTINUE_NEEDED {
			__antithesis_instrumentation__.Notify(19539)
			break
		} else {
			__antithesis_instrumentation__.Notify(19540)
		}
	}
	__antithesis_instrumentation__.Notify(19521)

	majStat = C.gss_display_name(&minStat, srcName, &gbuf, nil)
	if majStat != C.GSS_S_COMPLETE {
		__antithesis_instrumentation__.Notify(19541)
		return connClose, "", gssError("retrieving GSS user name failed", majStat, minStat)
	} else {
		__antithesis_instrumentation__.Notify(19542)
	}
	__antithesis_instrumentation__.Notify(19522)
	gssUser = C.GoStringN((*C.char)(gbuf.value), C.int(gbuf.length))
	C.gss_release_buffer(&lminS, &gbuf)

	return connClose, gssUser, nil
}

func gssError(msg string, majStat, minStat C.OM_uint32) error {
	__antithesis_instrumentation__.Notify(19543)
	var (
		gmsg          C.gss_buffer_desc
		lminS, msgCtx C.OM_uint32
	)

	msgCtx = 0
	C.gss_display_status(&lminS, majStat, C.GSS_C_GSS_CODE, C.GSS_C_NO_OID, &msgCtx, &gmsg)
	msgMajor := C.GoString((*C.char)(gmsg.value))
	C.gss_release_buffer(&lminS, &gmsg)

	msgCtx = 0
	C.gss_display_status(&lminS, minStat, C.GSS_C_MECH_CODE, C.GSS_C_NO_OID, &msgCtx, &gmsg)
	msgMinor := C.GoString((*C.char)(gmsg.value))
	C.gss_release_buffer(&lminS, &gmsg)

	return errors.Errorf("%s: %s: %s", msg, msgMajor, msgMinor)
}
