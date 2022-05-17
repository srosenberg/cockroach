package result

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Metrics struct {
	LeaseRequestSuccess  int
	LeaseRequestError    int
	LeaseTransferSuccess int
	LeaseTransferError   int
	ResolveCommit        int
	ResolveAbort         int
	ResolvePoison        int
	AddSSTableAsWrites   int
}

func (mt *Metrics) Add(o Metrics) {
	__antithesis_instrumentation__.Notify(97670)
	mt.LeaseRequestSuccess += o.LeaseRequestSuccess
	mt.LeaseRequestError += o.LeaseRequestError
	mt.LeaseTransferSuccess += o.LeaseTransferSuccess
	mt.LeaseTransferError += o.LeaseTransferError
	mt.ResolveCommit += o.ResolveCommit
	mt.ResolveAbort += o.ResolveAbort
	mt.ResolvePoison += o.ResolvePoison
	mt.AddSSTableAsWrites += o.AddSSTableAsWrites
}
