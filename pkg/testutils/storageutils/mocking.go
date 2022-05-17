package storageutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
)

type raftCmdIDAndIndex struct {
	IDKey kvserverbase.CmdIDKey
	Index int
}

func (r raftCmdIDAndIndex) String() string {
	__antithesis_instrumentation__.Notify(646425)
	return fmt.Sprintf("%s/%d", r.IDKey, r.Index)
}

type ReplayProtectionFilterWrapper struct {
	syncutil.Mutex
	inFlight          singleflight.Group
	processedCommands map[raftCmdIDAndIndex]*roachpb.Error
	filter            kvserverbase.ReplicaCommandFilter
}

func WrapFilterForReplayProtection(
	filter kvserverbase.ReplicaCommandFilter,
) kvserverbase.ReplicaCommandFilter {
	__antithesis_instrumentation__.Notify(646426)
	wrapper := ReplayProtectionFilterWrapper{
		processedCommands: make(map[raftCmdIDAndIndex]*roachpb.Error),
		filter:            filter,
	}
	return wrapper.run
}

func shallowCloneErrorWithTxn(pErr *roachpb.Error) *roachpb.Error {
	__antithesis_instrumentation__.Notify(646427)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(646429)
		pErrCopy := *pErr
		pErrCopy.SetTxn(pErrCopy.GetTxn())
		return &pErrCopy
	} else {
		__antithesis_instrumentation__.Notify(646430)
	}
	__antithesis_instrumentation__.Notify(646428)

	return nil
}

func (c *ReplayProtectionFilterWrapper) run(args kvserverbase.FilterArgs) *roachpb.Error {
	__antithesis_instrumentation__.Notify(646431)
	if !args.InRaftCmd() {
		__antithesis_instrumentation__.Notify(646435)
		return c.filter(args)
	} else {
		__antithesis_instrumentation__.Notify(646436)
	}
	__antithesis_instrumentation__.Notify(646432)

	mapKey := raftCmdIDAndIndex{args.CmdID, args.Index}

	c.Lock()
	if pErr, ok := c.processedCommands[mapKey]; ok {
		__antithesis_instrumentation__.Notify(646437)
		c.Unlock()
		return shallowCloneErrorWithTxn(pErr)
	} else {
		__antithesis_instrumentation__.Notify(646438)
	}
	__antithesis_instrumentation__.Notify(646433)

	resC, _ := c.inFlight.DoChan(mapKey.String(), func() (interface{}, error) {
		__antithesis_instrumentation__.Notify(646439)
		pErr := c.filter(args)

		c.Lock()
		defer c.Unlock()
		c.processedCommands[mapKey] = pErr
		return pErr, nil
	})
	__antithesis_instrumentation__.Notify(646434)
	c.Unlock()

	res := <-resC
	return shallowCloneErrorWithTxn(res.Val.(*roachpb.Error))
}
