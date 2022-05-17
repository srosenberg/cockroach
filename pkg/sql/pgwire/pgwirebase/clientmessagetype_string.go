package pgwirebase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(560903)

	var x [1]struct{}
	_ = x[ClientMsgBind-66]
	_ = x[ClientMsgClose-67]
	_ = x[ClientMsgCopyData-100]
	_ = x[ClientMsgCopyDone-99]
	_ = x[ClientMsgCopyFail-102]
	_ = x[ClientMsgDescribe-68]
	_ = x[ClientMsgExecute-69]
	_ = x[ClientMsgFlush-72]
	_ = x[ClientMsgParse-80]
	_ = x[ClientMsgPassword-112]
	_ = x[ClientMsgSimpleQuery-81]
	_ = x[ClientMsgSync-83]
	_ = x[ClientMsgTerminate-88]
}

const (
	_ClientMessageType_name_0 = "ClientMsgBindClientMsgCloseClientMsgDescribeClientMsgExecute"
	_ClientMessageType_name_1 = "ClientMsgFlush"
	_ClientMessageType_name_2 = "ClientMsgParseClientMsgSimpleQuery"
	_ClientMessageType_name_3 = "ClientMsgSync"
	_ClientMessageType_name_4 = "ClientMsgTerminate"
	_ClientMessageType_name_5 = "ClientMsgCopyDoneClientMsgCopyData"
	_ClientMessageType_name_6 = "ClientMsgCopyFail"
	_ClientMessageType_name_7 = "ClientMsgPassword"
)

var (
	_ClientMessageType_index_0 = [...]uint8{0, 13, 27, 44, 60}
	_ClientMessageType_index_2 = [...]uint8{0, 14, 34}
	_ClientMessageType_index_5 = [...]uint8{0, 17, 34}
)

func (i ClientMessageType) String() string {
	__antithesis_instrumentation__.Notify(560904)
	switch {
	case 66 <= i && func() bool {
		__antithesis_instrumentation__.Notify(560914)
		return i <= 69 == true
	}() == true:
		__antithesis_instrumentation__.Notify(560905)
		i -= 66
		return _ClientMessageType_name_0[_ClientMessageType_index_0[i]:_ClientMessageType_index_0[i+1]]
	case i == 72:
		__antithesis_instrumentation__.Notify(560906)
		return _ClientMessageType_name_1
	case 80 <= i && func() bool {
		__antithesis_instrumentation__.Notify(560915)
		return i <= 81 == true
	}() == true:
		__antithesis_instrumentation__.Notify(560907)
		i -= 80
		return _ClientMessageType_name_2[_ClientMessageType_index_2[i]:_ClientMessageType_index_2[i+1]]
	case i == 83:
		__antithesis_instrumentation__.Notify(560908)
		return _ClientMessageType_name_3
	case i == 88:
		__antithesis_instrumentation__.Notify(560909)
		return _ClientMessageType_name_4
	case 99 <= i && func() bool {
		__antithesis_instrumentation__.Notify(560916)
		return i <= 100 == true
	}() == true:
		__antithesis_instrumentation__.Notify(560910)
		i -= 99
		return _ClientMessageType_name_5[_ClientMessageType_index_5[i]:_ClientMessageType_index_5[i+1]]
	case i == 102:
		__antithesis_instrumentation__.Notify(560911)
		return _ClientMessageType_name_6
	case i == 112:
		__antithesis_instrumentation__.Notify(560912)
		return _ClientMessageType_name_7
	default:
		__antithesis_instrumentation__.Notify(560913)
		return "ClientMessageType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
