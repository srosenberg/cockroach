package kvserverbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "fmt"

type RaftCommandEncodingVersion byte

const (
	RaftVersionStandard RaftCommandEncodingVersion = 0

	RaftVersionSideloaded RaftCommandEncodingVersion = 1

	RaftCommandIDLen = 8

	RaftCommandPrefixLen = 1 + RaftCommandIDLen

	RaftCommandNoSplitBit = 1 << 7

	RaftCommandNoSplitMask = RaftCommandNoSplitBit - 1
)

func EncodeRaftCommand(
	version RaftCommandEncodingVersion, commandID CmdIDKey, command []byte,
) []byte {
	__antithesis_instrumentation__.Notify(101843)
	b := make([]byte, RaftCommandPrefixLen+len(command))
	EncodeRaftCommandPrefix(b[:RaftCommandPrefixLen], version, commandID)
	copy(b[RaftCommandPrefixLen:], command)
	return b
}

func EncodeRaftCommandPrefix(b []byte, version RaftCommandEncodingVersion, commandID CmdIDKey) {
	__antithesis_instrumentation__.Notify(101844)
	if len(commandID) != RaftCommandIDLen {
		__antithesis_instrumentation__.Notify(101847)
		panic(fmt.Sprintf("invalid command ID length; %d != %d", len(commandID), RaftCommandIDLen))
	} else {
		__antithesis_instrumentation__.Notify(101848)
	}
	__antithesis_instrumentation__.Notify(101845)
	if len(b) != RaftCommandPrefixLen {
		__antithesis_instrumentation__.Notify(101849)
		panic(fmt.Sprintf("invalid command prefix length; %d != %d", len(b), RaftCommandPrefixLen))
	} else {
		__antithesis_instrumentation__.Notify(101850)
	}
	__antithesis_instrumentation__.Notify(101846)
	b[0] = byte(version)
	copy(b[1:], []byte(commandID))
}

func DecodeRaftCommand(data []byte) (CmdIDKey, []byte) {
	__antithesis_instrumentation__.Notify(101851)
	v := RaftCommandEncodingVersion(data[0] & RaftCommandNoSplitMask)
	if v != RaftVersionStandard && func() bool {
		__antithesis_instrumentation__.Notify(101853)
		return v != RaftVersionSideloaded == true
	}() == true {
		__antithesis_instrumentation__.Notify(101854)
		panic(fmt.Sprintf("unknown command encoding version %v", data[0]))
	} else {
		__antithesis_instrumentation__.Notify(101855)
	}
	__antithesis_instrumentation__.Notify(101852)
	return CmdIDKey(data[1 : 1+RaftCommandIDLen]), data[1+RaftCommandIDLen:]
}

func EncodeTestRaftCommand(command []byte, commandID CmdIDKey) []byte {
	__antithesis_instrumentation__.Notify(101856)
	return EncodeRaftCommand(RaftVersionStandard, commandID, command)
}
