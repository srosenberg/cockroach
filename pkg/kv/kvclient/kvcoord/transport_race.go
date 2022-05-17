//go:build race
// +build race

package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

var running int32
var incoming chan *roachpb.BatchRequest

func init() {
	incoming = make(chan *roachpb.BatchRequest, 100)
}

const defaultRaceInterval = 150 * time.Microsecond

func jitter(avgInterval time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(88071)

	if avgInterval < defaultRaceInterval {
		__antithesis_instrumentation__.Notify(88073)
		avgInterval = defaultRaceInterval
	} else {
		__antithesis_instrumentation__.Notify(88074)
	}
	__antithesis_instrumentation__.Notify(88072)
	return time.Duration(rand.Int63n(int64(2 * avgInterval)))
}

type raceTransport struct {
	Transport
}

func (tr raceTransport) SendNext(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(88075)

	requestsCopy := make([]roachpb.RequestUnion, len(ba.Requests))
	for i, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(88078)

		requestsCopy[i] = reflect.Indirect(reflect.ValueOf(ru)).Interface().(roachpb.RequestUnion)
	}
	__antithesis_instrumentation__.Notify(88076)
	ba.Requests = requestsCopy
	select {

	case incoming <- &ba:
		__antithesis_instrumentation__.Notify(88079)
	default:
		__antithesis_instrumentation__.Notify(88080)

	}
	__antithesis_instrumentation__.Notify(88077)
	return tr.Transport.SendNext(ctx, ba)
}

func GRPCTransportFactory(
	opts SendOptions, nodeDialer *nodedialer.Dialer, replicas ReplicaSlice,
) (Transport, error) {
	__antithesis_instrumentation__.Notify(88081)
	if atomic.AddInt32(&running, 1) <= 1 {
		__antithesis_instrumentation__.Notify(88084)
		if err := nodeDialer.Stopper().RunAsyncTask(
			context.TODO(), "transport racer", func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(88085)
				var iters int
				var curIdx int
				defer func() {
					__antithesis_instrumentation__.Notify(88087)
					atomic.StoreInt32(&running, 0)
					log.Infof(
						ctx,
						"transport race promotion: ran %d iterations on up to %d requests",
						iters, curIdx+1,
					)
				}()
				__antithesis_instrumentation__.Notify(88086)

				const size = 1000
				bas := make([]*roachpb.BatchRequest, size)
				encoder := json.NewEncoder(ioutil.Discard)
				for {
					__antithesis_instrumentation__.Notify(88088)
					iters++
					start := timeutil.Now()
					for _, ba := range bas {
						__antithesis_instrumentation__.Notify(88090)
						if ba != nil {
							__antithesis_instrumentation__.Notify(88091)
							if err := encoder.Encode(ba); err != nil {
								__antithesis_instrumentation__.Notify(88092)
								panic(err)
							} else {
								__antithesis_instrumentation__.Notify(88093)
							}
						} else {
							__antithesis_instrumentation__.Notify(88094)
						}
					}
					__antithesis_instrumentation__.Notify(88089)

					jittered := time.After(jitter(timeutil.Since(start)))

					for {
						__antithesis_instrumentation__.Notify(88095)
						select {
						case <-nodeDialer.Stopper().ShouldQuiesce():
							__antithesis_instrumentation__.Notify(88097)
							return
						case ba := <-incoming:
							__antithesis_instrumentation__.Notify(88098)
							bas[curIdx%size] = ba
							curIdx++
							continue
						case <-jittered:
							__antithesis_instrumentation__.Notify(88099)
						}
						__antithesis_instrumentation__.Notify(88096)
						break
					}
				}
			}); err != nil {
			__antithesis_instrumentation__.Notify(88100)

			atomic.StoreInt32(&running, 0)
		} else {
			__antithesis_instrumentation__.Notify(88101)
		}
	} else {
		__antithesis_instrumentation__.Notify(88102)
	}
	__antithesis_instrumentation__.Notify(88082)

	t, err := grpcTransportFactoryImpl(opts, nodeDialer, replicas)
	if err != nil {
		__antithesis_instrumentation__.Notify(88103)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(88104)
	}
	__antithesis_instrumentation__.Notify(88083)
	return &raceTransport{Transport: t}, nil
}
