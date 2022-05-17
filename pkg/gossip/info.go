package gossip

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func (i *Info) expired(now int64) bool {
	__antithesis_instrumentation__.Notify(67786)
	return i.TTLStamp <= now
}

func (i *Info) isFresh(highWaterStamp int64) bool {
	__antithesis_instrumentation__.Notify(67787)
	return i.OrigStamp > highWaterStamp
}

type infoMap map[string]*Info
