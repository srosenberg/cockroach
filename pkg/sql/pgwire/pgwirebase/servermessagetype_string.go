package pgwirebase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(561418)

	var x [1]struct{}
	_ = x[ServerMsgAuth-82]
	_ = x[ServerMsgBackendKeyData-75]
	_ = x[ServerMsgBindComplete-50]
	_ = x[ServerMsgCommandComplete-67]
	_ = x[ServerMsgCloseComplete-51]
	_ = x[ServerMsgCopyInResponse-71]
	_ = x[ServerMsgDataRow-68]
	_ = x[ServerMsgEmptyQuery-73]
	_ = x[ServerMsgErrorResponse-69]
	_ = x[ServerMsgNoticeResponse-78]
	_ = x[ServerMsgNoData-110]
	_ = x[ServerMsgParameterDescription-116]
	_ = x[ServerMsgParameterStatus-83]
	_ = x[ServerMsgParseComplete-49]
	_ = x[ServerMsgPortalSuspended-115]
	_ = x[ServerMsgReady-90]
	_ = x[ServerMsgRowDescription-84]
}

const (
	_ServerMessageType_name_0 = "ServerMsgParseCompleteServerMsgBindCompleteServerMsgCloseComplete"
	_ServerMessageType_name_1 = "ServerMsgCommandCompleteServerMsgDataRowServerMsgErrorResponse"
	_ServerMessageType_name_2 = "ServerMsgCopyInResponse"
	_ServerMessageType_name_3 = "ServerMsgEmptyQuery"
	_ServerMessageType_name_4 = "ServerMsgBackendKeyData"
	_ServerMessageType_name_5 = "ServerMsgNoticeResponse"
	_ServerMessageType_name_6 = "ServerMsgAuthServerMsgParameterStatusServerMsgRowDescription"
	_ServerMessageType_name_7 = "ServerMsgReady"
	_ServerMessageType_name_8 = "ServerMsgNoData"
	_ServerMessageType_name_9 = "ServerMsgPortalSuspendedServerMsgParameterDescription"
)

var (
	_ServerMessageType_index_0 = [...]uint8{0, 22, 43, 65}
	_ServerMessageType_index_1 = [...]uint8{0, 24, 40, 62}
	_ServerMessageType_index_6 = [...]uint8{0, 13, 37, 60}
	_ServerMessageType_index_9 = [...]uint8{0, 24, 53}
)

func (i ServerMessageType) String() string {
	__antithesis_instrumentation__.Notify(561419)
	switch {
	case 49 <= i && func() bool {
		__antithesis_instrumentation__.Notify(561431)
		return i <= 51 == true
	}() == true:
		__antithesis_instrumentation__.Notify(561420)
		i -= 49
		return _ServerMessageType_name_0[_ServerMessageType_index_0[i]:_ServerMessageType_index_0[i+1]]
	case 67 <= i && func() bool {
		__antithesis_instrumentation__.Notify(561432)
		return i <= 69 == true
	}() == true:
		__antithesis_instrumentation__.Notify(561421)
		i -= 67
		return _ServerMessageType_name_1[_ServerMessageType_index_1[i]:_ServerMessageType_index_1[i+1]]
	case i == 71:
		__antithesis_instrumentation__.Notify(561422)
		return _ServerMessageType_name_2
	case i == 73:
		__antithesis_instrumentation__.Notify(561423)
		return _ServerMessageType_name_3
	case i == 75:
		__antithesis_instrumentation__.Notify(561424)
		return _ServerMessageType_name_4
	case i == 78:
		__antithesis_instrumentation__.Notify(561425)
		return _ServerMessageType_name_5
	case 82 <= i && func() bool {
		__antithesis_instrumentation__.Notify(561433)
		return i <= 84 == true
	}() == true:
		__antithesis_instrumentation__.Notify(561426)
		i -= 82
		return _ServerMessageType_name_6[_ServerMessageType_index_6[i]:_ServerMessageType_index_6[i+1]]
	case i == 90:
		__antithesis_instrumentation__.Notify(561427)
		return _ServerMessageType_name_7
	case i == 110:
		__antithesis_instrumentation__.Notify(561428)
		return _ServerMessageType_name_8
	case 115 <= i && func() bool {
		__antithesis_instrumentation__.Notify(561434)
		return i <= 116 == true
	}() == true:
		__antithesis_instrumentation__.Notify(561429)
		i -= 115
		return _ServerMessageType_name_9[_ServerMessageType_index_9[i]:_ServerMessageType_index_9[i+1]]
	default:
		__antithesis_instrumentation__.Notify(561430)
		return "ServerMessageType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
