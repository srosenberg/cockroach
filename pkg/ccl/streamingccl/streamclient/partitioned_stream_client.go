package streamclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"net"
	"net/url"

	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streampb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/streaming"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type partitionedStreamClient struct {
	srcDB          *gosql.DB
	urlPlaceholder url.URL

	mu struct {
		syncutil.Mutex

		closed              bool
		activeSubscriptions map[*partitionedStreamSubscription]struct{}
	}
}

func newPartitionedStreamClient(remote *url.URL) (*partitionedStreamClient, error) {
	__antithesis_instrumentation__.Notify(24963)
	db, err := gosql.Open("postgres", remote.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(24965)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(24966)
	}
	__antithesis_instrumentation__.Notify(24964)

	client := &partitionedStreamClient{
		srcDB:          db,
		urlPlaceholder: *remote,
	}
	client.mu.activeSubscriptions = make(map[*partitionedStreamSubscription]struct{})
	return client, nil
}

var _ Client = &partitionedStreamClient{}

func (p *partitionedStreamClient) Create(
	ctx context.Context, tenantID roachpb.TenantID,
) (streaming.StreamID, error) {
	__antithesis_instrumentation__.Notify(24967)
	streamID := streaming.InvalidStreamID

	conn, err := p.srcDB.Conn(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(24970)
		return streamID, err
	} else {
		__antithesis_instrumentation__.Notify(24971)
	}
	__antithesis_instrumentation__.Notify(24968)

	row := conn.QueryRowContext(ctx, `SELECT crdb_internal.start_replication_stream($1)`, tenantID.ToUint64())
	if row.Err() != nil {
		__antithesis_instrumentation__.Notify(24972)
		return streamID, errors.Wrapf(row.Err(), "Error in creating replication stream for tenant %s", tenantID.String())
	} else {
		__antithesis_instrumentation__.Notify(24973)
	}
	__antithesis_instrumentation__.Notify(24969)

	err = row.Scan(&streamID)
	return streamID, err
}

func (p *partitionedStreamClient) Heartbeat(
	ctx context.Context, streamID streaming.StreamID, consumed hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(24974)
	conn, err := p.srcDB.Conn(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(24980)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24981)
	}
	__antithesis_instrumentation__.Notify(24975)

	row := conn.QueryRowContext(ctx,
		`SELECT crdb_internal.replication_stream_progress($1, $2)`, streamID, consumed.String())
	if row.Err() != nil {
		__antithesis_instrumentation__.Notify(24982)
		return errors.Wrapf(row.Err(), "Error in sending heartbeats to replication stream %d", streamID)
	} else {
		__antithesis_instrumentation__.Notify(24983)
	}
	__antithesis_instrumentation__.Notify(24976)

	var rawStatus []byte
	if err := row.Scan(&rawStatus); err != nil {
		__antithesis_instrumentation__.Notify(24984)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24985)
	}
	__antithesis_instrumentation__.Notify(24977)
	var status streampb.StreamReplicationStatus
	if err := protoutil.Unmarshal(rawStatus, &status); err != nil {
		__antithesis_instrumentation__.Notify(24986)
		return err
	} else {
		__antithesis_instrumentation__.Notify(24987)
	}
	__antithesis_instrumentation__.Notify(24978)

	if status.StreamStatus != streampb.StreamReplicationStatus_STREAM_ACTIVE {
		__antithesis_instrumentation__.Notify(24988)
		return streamingccl.NewStreamStatusErr(streamID, status.StreamStatus)
	} else {
		__antithesis_instrumentation__.Notify(24989)
	}
	__antithesis_instrumentation__.Notify(24979)
	return nil
}

func (p *partitionedStreamClient) postgresURL(servingAddr string) (url.URL, error) {
	__antithesis_instrumentation__.Notify(24990)
	host, port, err := net.SplitHostPort(servingAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(24992)
		return url.URL{}, err
	} else {
		__antithesis_instrumentation__.Notify(24993)
	}
	__antithesis_instrumentation__.Notify(24991)
	res := p.urlPlaceholder
	res.Host = net.JoinHostPort(host, port)
	return res, nil
}

func (p *partitionedStreamClient) Plan(
	ctx context.Context, streamID streaming.StreamID,
) (Topology, error) {
	__antithesis_instrumentation__.Notify(24994)
	conn, err := p.srcDB.Conn(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(25000)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25001)
	}
	__antithesis_instrumentation__.Notify(24995)

	row := conn.QueryRowContext(ctx, `SELECT crdb_internal.replication_stream_spec($1)`, streamID)
	if row.Err() != nil {
		__antithesis_instrumentation__.Notify(25002)
		return nil, errors.Wrap(row.Err(), "Error in planning a replication stream")
	} else {
		__antithesis_instrumentation__.Notify(25003)
	}
	__antithesis_instrumentation__.Notify(24996)

	var rawSpec []byte
	if err := row.Scan(&rawSpec); err != nil {
		__antithesis_instrumentation__.Notify(25004)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25005)
	}
	__antithesis_instrumentation__.Notify(24997)
	var spec streampb.ReplicationStreamSpec
	if err := protoutil.Unmarshal(rawSpec, &spec); err != nil {
		__antithesis_instrumentation__.Notify(25006)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25007)
	}
	__antithesis_instrumentation__.Notify(24998)

	topology := Topology{}
	for _, sp := range spec.Partitions {
		__antithesis_instrumentation__.Notify(25008)
		pgURL, err := p.postgresURL(sp.SQLAddress.String())
		if err != nil {
			__antithesis_instrumentation__.Notify(25011)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(25012)
		}
		__antithesis_instrumentation__.Notify(25009)
		rawSpec, err := protoutil.Marshal(sp.PartitionSpec)
		if err != nil {
			__antithesis_instrumentation__.Notify(25013)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(25014)
		}
		__antithesis_instrumentation__.Notify(25010)
		topology = append(topology, PartitionInfo{
			ID:                sp.NodeID.String(),
			SubscriptionToken: SubscriptionToken(rawSpec),
			SrcInstanceID:     int(sp.NodeID),
			SrcAddr:           streamingccl.PartitionAddress(pgURL.String()),
			SrcLocality:       sp.Locality,
		})
	}
	__antithesis_instrumentation__.Notify(24999)
	return topology, nil
}

func (p *partitionedStreamClient) Close() error {
	__antithesis_instrumentation__.Notify(25015)
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.mu.closed {
		__antithesis_instrumentation__.Notify(25018)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(25019)
	}
	__antithesis_instrumentation__.Notify(25016)

	p.mu.closed = true
	for sub := range p.mu.activeSubscriptions {
		__antithesis_instrumentation__.Notify(25020)
		close(sub.closeChan)
		delete(p.mu.activeSubscriptions, sub)
	}
	__antithesis_instrumentation__.Notify(25017)
	return p.srcDB.Close()
}

func (p *partitionedStreamClient) Subscribe(
	ctx context.Context, stream streaming.StreamID, spec SubscriptionToken, checkpoint hlc.Timestamp,
) (Subscription, error) {
	__antithesis_instrumentation__.Notify(25021)
	sps := streampb.StreamPartitionSpec{}
	if err := protoutil.Unmarshal(spec, &sps); err != nil {
		__antithesis_instrumentation__.Notify(25024)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25025)
	}
	__antithesis_instrumentation__.Notify(25022)
	sps.StartFrom = checkpoint

	specBytes, err := protoutil.Marshal(&sps)
	if err != nil {
		__antithesis_instrumentation__.Notify(25026)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25027)
	}
	__antithesis_instrumentation__.Notify(25023)

	res := &partitionedStreamSubscription{
		eventsChan: make(chan streamingccl.Event),
		db:         p.srcDB,
		specBytes:  specBytes,
		streamID:   stream,
		closeChan:  make(chan struct{}),
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mu.activeSubscriptions[res] = struct{}{}
	return res, nil
}

type partitionedStreamSubscription struct {
	err        error
	db         *gosql.DB
	eventsChan chan streamingccl.Event

	closeChan chan struct{}

	streamEvent *streampb.StreamEvent
	specBytes   []byte
	streamID    streaming.StreamID
}

var _ Subscription = (*partitionedStreamSubscription)(nil)

func parseEvent(streamEvent *streampb.StreamEvent) streamingccl.Event {
	__antithesis_instrumentation__.Notify(25028)
	if streamEvent == nil {
		__antithesis_instrumentation__.Notify(25032)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(25033)
	}
	__antithesis_instrumentation__.Notify(25029)

	if streamEvent.Checkpoint != nil {
		__antithesis_instrumentation__.Notify(25034)
		event := streamingccl.MakeCheckpointEvent(streamEvent.Checkpoint.Spans[0].Timestamp)
		streamEvent.Checkpoint = nil
		return event
	} else {
		__antithesis_instrumentation__.Notify(25035)
	}
	__antithesis_instrumentation__.Notify(25030)
	if streamEvent.Batch != nil {
		__antithesis_instrumentation__.Notify(25036)
		event := streamingccl.MakeKVEvent(streamEvent.Batch.KeyValues[0])
		streamEvent.Batch.KeyValues = streamEvent.Batch.KeyValues[1:]
		if len(streamEvent.Batch.KeyValues) == 0 {
			__antithesis_instrumentation__.Notify(25038)
			streamEvent.Batch = nil
		} else {
			__antithesis_instrumentation__.Notify(25039)
		}
		__antithesis_instrumentation__.Notify(25037)
		return event
	} else {
		__antithesis_instrumentation__.Notify(25040)
	}
	__antithesis_instrumentation__.Notify(25031)
	return nil
}

func (p *partitionedStreamSubscription) Subscribe(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(25041)
	defer close(p.eventsChan)
	conn, err := p.db.Conn(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(25046)
		return err
	} else {
		__antithesis_instrumentation__.Notify(25047)
	}
	__antithesis_instrumentation__.Notify(25042)

	_, err = conn.ExecContext(ctx, `SET avoid_buffering = true`)
	if err != nil {
		__antithesis_instrumentation__.Notify(25048)
		return err
	} else {
		__antithesis_instrumentation__.Notify(25049)
	}
	__antithesis_instrumentation__.Notify(25043)
	rows, err := conn.QueryContext(ctx, `SELECT * FROM crdb_internal.stream_partition($1, $2)`,
		p.streamID, p.specBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(25050)
		return err
	} else {
		__antithesis_instrumentation__.Notify(25051)
	}
	__antithesis_instrumentation__.Notify(25044)
	defer rows.Close()

	getNextEvent := func() (streamingccl.Event, error) {
		__antithesis_instrumentation__.Notify(25052)
		if e := parseEvent(p.streamEvent); e != nil {
			__antithesis_instrumentation__.Notify(25057)
			return e, nil
		} else {
			__antithesis_instrumentation__.Notify(25058)
		}
		__antithesis_instrumentation__.Notify(25053)

		if !rows.Next() {
			__antithesis_instrumentation__.Notify(25059)
			if err := rows.Err(); err != nil {
				__antithesis_instrumentation__.Notify(25061)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(25062)
			}
			__antithesis_instrumentation__.Notify(25060)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(25063)
		}
		__antithesis_instrumentation__.Notify(25054)
		var data []byte
		if err := rows.Scan(&data); err != nil {
			__antithesis_instrumentation__.Notify(25064)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(25065)
		}
		__antithesis_instrumentation__.Notify(25055)
		var streamEvent streampb.StreamEvent
		if err := protoutil.Unmarshal(data, &streamEvent); err != nil {
			__antithesis_instrumentation__.Notify(25066)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(25067)
		}
		__antithesis_instrumentation__.Notify(25056)
		p.streamEvent = &streamEvent
		return parseEvent(p.streamEvent), nil
	}
	__antithesis_instrumentation__.Notify(25045)

	for {
		__antithesis_instrumentation__.Notify(25068)
		event, err := getNextEvent()
		if err != nil {
			__antithesis_instrumentation__.Notify(25070)
			p.err = err
			return err
		} else {
			__antithesis_instrumentation__.Notify(25071)
		}
		__antithesis_instrumentation__.Notify(25069)
		select {
		case p.eventsChan <- event:
			__antithesis_instrumentation__.Notify(25072)
		case <-p.closeChan:
			__antithesis_instrumentation__.Notify(25073)

			return nil
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(25074)
			p.err = ctx.Err()
			return p.err
		}
	}
}

func (p *partitionedStreamSubscription) Events() <-chan streamingccl.Event {
	__antithesis_instrumentation__.Notify(25075)
	return p.eventsChan
}

func (p *partitionedStreamSubscription) Err() error {
	__antithesis_instrumentation__.Notify(25076)
	return p.err
}
