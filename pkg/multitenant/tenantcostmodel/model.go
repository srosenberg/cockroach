package tenantcostmodel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/roachpb"

type RU float64

type Config struct {
	KVReadRequest RU

	KVReadByte RU

	KVWriteRequest RU

	KVWriteByte RU

	PodCPUSecond RU

	PGWireEgressByte RU
}

func (c *Config) KVReadCost(bytes int64) RU {
	__antithesis_instrumentation__.Notify(128759)
	return c.KVReadRequest + RU(bytes)*c.KVReadByte
}

func (c *Config) KVWriteCost(bytes int64) RU {
	__antithesis_instrumentation__.Notify(128760)
	return c.KVWriteRequest + RU(bytes)*c.KVWriteByte
}

func (c *Config) RequestCost(bri RequestInfo) RU {
	__antithesis_instrumentation__.Notify(128761)
	if isWrite, writeBytes := bri.IsWrite(); isWrite {
		__antithesis_instrumentation__.Notify(128763)
		return c.KVWriteCost(writeBytes)
	} else {
		__antithesis_instrumentation__.Notify(128764)
	}
	__antithesis_instrumentation__.Notify(128762)
	return c.KVReadRequest
}

func (c *Config) ResponseCost(bri ResponseInfo) RU {
	__antithesis_instrumentation__.Notify(128765)
	return RU(bri.ReadBytes()) * c.KVReadByte
}

type RequestInfo struct {
	writeBytes int64
}

func MakeRequestInfo(ba *roachpb.BatchRequest) RequestInfo {
	__antithesis_instrumentation__.Notify(128766)
	if !ba.IsWrite() {
		__antithesis_instrumentation__.Notify(128769)
		return RequestInfo{writeBytes: -1}
	} else {
		__antithesis_instrumentation__.Notify(128770)
	}
	__antithesis_instrumentation__.Notify(128767)

	var writeBytes int64
	for i := range ba.Requests {
		__antithesis_instrumentation__.Notify(128771)
		if swr, isSizedWrite := ba.Requests[i].GetInner().(roachpb.SizedWriteRequest); isSizedWrite {
			__antithesis_instrumentation__.Notify(128772)
			writeBytes += swr.WriteBytes()
		} else {
			__antithesis_instrumentation__.Notify(128773)
		}
	}
	__antithesis_instrumentation__.Notify(128768)
	return RequestInfo{writeBytes: writeBytes}
}

func (bri RequestInfo) IsWrite() (isWrite bool, writeBytes int64) {
	__antithesis_instrumentation__.Notify(128774)
	if bri.writeBytes == -1 {
		__antithesis_instrumentation__.Notify(128776)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(128777)
	}
	__antithesis_instrumentation__.Notify(128775)
	return true, bri.writeBytes
}

func TestingRequestInfo(isWrite bool, writeBytes int64) RequestInfo {
	__antithesis_instrumentation__.Notify(128778)
	if !isWrite {
		__antithesis_instrumentation__.Notify(128780)
		return RequestInfo{writeBytes: -1}
	} else {
		__antithesis_instrumentation__.Notify(128781)
	}
	__antithesis_instrumentation__.Notify(128779)
	return RequestInfo{writeBytes: writeBytes}
}

type ResponseInfo struct {
	readBytes int64
}

func MakeResponseInfo(br *roachpb.BatchResponse) ResponseInfo {
	__antithesis_instrumentation__.Notify(128782)
	var readBytes int64
	for _, ru := range br.Responses {
		__antithesis_instrumentation__.Notify(128784)
		readBytes += ru.GetInner().Header().NumBytes
	}
	__antithesis_instrumentation__.Notify(128783)
	return ResponseInfo{readBytes: readBytes}
}

func (bri ResponseInfo) ReadBytes() int64 {
	__antithesis_instrumentation__.Notify(128785)
	return bri.readBytes
}

func TestingResponseInfo(readBytes int64) ResponseInfo {
	__antithesis_instrumentation__.Notify(128786)
	return ResponseInfo{readBytes: readBytes}
}
