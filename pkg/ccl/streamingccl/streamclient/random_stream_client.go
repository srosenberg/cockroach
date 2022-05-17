package streamclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/streaming"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

const (
	RandomStreamSchemaPlaceholder = "CREATE TABLE %s (k INT PRIMARY KEY, v INT)"

	RandomGenScheme = "randomgen"

	ValueRangeKey = "VALUE_RANGE"

	EventFrequency = "EVENT_FREQUENCY"

	KVsPerCheckpoint = "KVS_PER_CHECKPOINT"

	NumPartitions = "NUM_PARTITIONS"

	DupProbability = "DUP_PROBABILITY"

	TenantID = "TENANT_ID"

	IngestionDatabaseID = 50

	IngestionTablePrefix = "foo"
)

var randomStreamClientSingleton = func() *randomStreamClient {
	__antithesis_instrumentation__.Notify(25077)
	c := randomStreamClient{}
	c.mu.tableID = 52
	return &c
}()

func GetRandomStreamClientSingletonForTesting() Client {
	__antithesis_instrumentation__.Notify(25078)
	return randomStreamClientSingleton
}

type InterceptFn func(event streamingccl.Event, spec SubscriptionToken)

type InterceptableStreamClient interface {
	Client

	RegisterInterception(fn InterceptFn)
}

type randomStreamConfig struct {
	valueRange       int
	eventFrequency   time.Duration
	kvsPerCheckpoint int
	numPartitions    int
	dupProbability   float64
	tenantID         roachpb.TenantID
}

func parseRandomStreamConfig(streamURL *url.URL) (randomStreamConfig, error) {
	__antithesis_instrumentation__.Notify(25079)
	c := randomStreamConfig{
		valueRange:       100,
		eventFrequency:   10 * time.Microsecond,
		kvsPerCheckpoint: 100,
		numPartitions:    1,
		dupProbability:   0.5,
		tenantID:         roachpb.SystemTenantID,
	}

	var err error
	if valueRangeStr := streamURL.Query().Get(ValueRangeKey); valueRangeStr != "" {
		__antithesis_instrumentation__.Notify(25086)
		c.valueRange, err = strconv.Atoi(valueRangeStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(25087)
			return c, err
		} else {
			__antithesis_instrumentation__.Notify(25088)
		}
	} else {
		__antithesis_instrumentation__.Notify(25089)
	}
	__antithesis_instrumentation__.Notify(25080)

	if kvFreqStr := streamURL.Query().Get(EventFrequency); kvFreqStr != "" {
		__antithesis_instrumentation__.Notify(25090)
		kvFreq, err := strconv.Atoi(kvFreqStr)
		c.eventFrequency = time.Duration(kvFreq)
		if err != nil {
			__antithesis_instrumentation__.Notify(25091)
			return c, err
		} else {
			__antithesis_instrumentation__.Notify(25092)
		}
	} else {
		__antithesis_instrumentation__.Notify(25093)
	}
	__antithesis_instrumentation__.Notify(25081)

	if kvsPerCheckpointStr := streamURL.Query().Get(KVsPerCheckpoint); kvsPerCheckpointStr != "" {
		__antithesis_instrumentation__.Notify(25094)
		c.kvsPerCheckpoint, err = strconv.Atoi(kvsPerCheckpointStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(25095)
			return c, err
		} else {
			__antithesis_instrumentation__.Notify(25096)
		}
	} else {
		__antithesis_instrumentation__.Notify(25097)
	}
	__antithesis_instrumentation__.Notify(25082)

	if numPartitionsStr := streamURL.Query().Get(NumPartitions); numPartitionsStr != "" {
		__antithesis_instrumentation__.Notify(25098)
		c.numPartitions, err = strconv.Atoi(numPartitionsStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(25099)
			return c, err
		} else {
			__antithesis_instrumentation__.Notify(25100)
		}
	} else {
		__antithesis_instrumentation__.Notify(25101)
	}
	__antithesis_instrumentation__.Notify(25083)

	if dupProbStr := streamURL.Query().Get(DupProbability); dupProbStr != "" {
		__antithesis_instrumentation__.Notify(25102)
		c.dupProbability, err = strconv.ParseFloat(dupProbStr, 32)
		if err != nil {
			__antithesis_instrumentation__.Notify(25103)
			return c, err
		} else {
			__antithesis_instrumentation__.Notify(25104)
		}
	} else {
		__antithesis_instrumentation__.Notify(25105)
	}
	__antithesis_instrumentation__.Notify(25084)

	if tenantIDStr := streamURL.Query().Get(TenantID); tenantIDStr != "" {
		__antithesis_instrumentation__.Notify(25106)
		id, err := strconv.Atoi(tenantIDStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(25108)
			return c, err
		} else {
			__antithesis_instrumentation__.Notify(25109)
		}
		__antithesis_instrumentation__.Notify(25107)
		c.tenantID = roachpb.MakeTenantID(uint64(id))
	} else {
		__antithesis_instrumentation__.Notify(25110)
	}
	__antithesis_instrumentation__.Notify(25085)
	return c, nil
}

func (c randomStreamConfig) URL(table int) string {
	__antithesis_instrumentation__.Notify(25111)
	u := &url.URL{
		Scheme: RandomGenScheme,
		Host:   strconv.Itoa(table),
	}
	q := u.Query()
	q.Add(ValueRangeKey, strconv.Itoa(c.valueRange))
	q.Add(EventFrequency, strconv.Itoa(int(c.eventFrequency)))
	q.Add(KVsPerCheckpoint, strconv.Itoa(c.kvsPerCheckpoint))
	q.Add(NumPartitions, strconv.Itoa(c.numPartitions))
	q.Add(DupProbability, fmt.Sprintf("%f", c.dupProbability))
	q.Add(TenantID, strconv.Itoa(int(c.tenantID.ToUint64())))
	u.RawQuery = q.Encode()
	return u.String()
}

type randomStreamClient struct {
	config randomStreamConfig

	mu struct {
		syncutil.Mutex

		interceptors []func(streamingccl.Event, SubscriptionToken)
		tableID      int
	}
}

var _ Client = &randomStreamClient{}
var _ InterceptableStreamClient = &randomStreamClient{}

func newRandomStreamClient(streamURL *url.URL) (Client, error) {
	__antithesis_instrumentation__.Notify(25112)
	c := randomStreamClientSingleton

	streamConfig, err := parseRandomStreamConfig(streamURL)
	if err != nil {
		__antithesis_instrumentation__.Notify(25114)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25115)
	}
	__antithesis_instrumentation__.Notify(25113)
	c.config = streamConfig
	return c, nil
}

func (m *randomStreamClient) getNextTableID() int {
	__antithesis_instrumentation__.Notify(25116)
	m.mu.Lock()
	defer m.mu.Unlock()
	ret := m.mu.tableID
	m.mu.tableID++
	return ret
}

func (m *randomStreamClient) Plan(ctx context.Context, id streaming.StreamID) (Topology, error) {
	__antithesis_instrumentation__.Notify(25117)
	topology := make(Topology, 0, m.config.numPartitions)
	log.Infof(ctx, "planning random stream for tenant %d", m.config.tenantID)

	for i := 0; i < m.config.numPartitions; i++ {
		__antithesis_instrumentation__.Notify(25119)
		partitionURI := m.config.URL(m.getNextTableID())
		log.Infof(ctx, "planning random stream partition %d for tenant %d: %q", i, m.config.tenantID, partitionURI)

		topology = append(topology,
			PartitionInfo{
				ID:                strconv.Itoa(i),
				SrcAddr:           streamingccl.PartitionAddress(partitionURI),
				SubscriptionToken: []byte(partitionURI),
			})
	}
	__antithesis_instrumentation__.Notify(25118)

	return topology, nil
}

func (m *randomStreamClient) Create(
	ctx context.Context, target roachpb.TenantID,
) (streaming.StreamID, error) {
	__antithesis_instrumentation__.Notify(25120)
	log.Infof(ctx, "creating random stream for tenant %d", target.ToUint64())
	m.config.tenantID = target
	return streaming.StreamID(target.ToUint64()), nil
}

func (m *randomStreamClient) Heartbeat(
	ctx context.Context, ID streaming.StreamID, _ hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(25121)
	return nil
}

func (m *randomStreamClient) getDescriptorAndNamespaceKVForTableID(
	config randomStreamConfig, tableID descpb.ID,
) (*tabledesc.Mutable, []roachpb.KeyValue, error) {
	__antithesis_instrumentation__.Notify(25122)
	tableName := fmt.Sprintf("%s%d", IngestionTablePrefix, tableID)
	testTable, err := sql.CreateTestTableDescriptor(
		context.Background(),
		IngestionDatabaseID,
		tableID,
		fmt.Sprintf(RandomStreamSchemaPlaceholder, tableName),
		catpb.NewBasePrivilegeDescriptor(security.RootUserName()),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(25125)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(25126)
	}
	__antithesis_instrumentation__.Notify(25123)

	codec := keys.MakeSQLCodec(config.tenantID)
	key := catalogkeys.MakePublicObjectNameKey(codec, 50, testTable.Name)
	k := rekey(config.tenantID, key)
	var value roachpb.Value
	value.SetInt(int64(testTable.GetID()))
	value.InitChecksum(k)
	namespaceKV := roachpb.KeyValue{
		Key:   k,
		Value: value,
	}

	descKey := catalogkeys.MakeDescMetadataKey(codec, testTable.GetID())
	descKey = rekey(config.tenantID, descKey)
	descDesc := testTable.DescriptorProto()
	var descValue roachpb.Value
	if err := descValue.SetProto(descDesc); err != nil {
		__antithesis_instrumentation__.Notify(25127)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(25128)
	}
	__antithesis_instrumentation__.Notify(25124)
	descValue.InitChecksum(descKey)
	descKV := roachpb.KeyValue{
		Key:   descKey,
		Value: descValue,
	}

	return testTable, []roachpb.KeyValue{namespaceKV, descKV}, nil
}

func (m *randomStreamClient) Close() error {
	__antithesis_instrumentation__.Notify(25129)
	return nil
}

func (m *randomStreamClient) Subscribe(
	ctx context.Context, stream streaming.StreamID, spec SubscriptionToken, checkpoint hlc.Timestamp,
) (Subscription, error) {
	__antithesis_instrumentation__.Notify(25130)
	partitionURL, err := url.Parse(string(spec))
	if err != nil {
		__antithesis_instrumentation__.Notify(25137)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25138)
	}
	__antithesis_instrumentation__.Notify(25131)
	config, err := parseRandomStreamConfig(partitionURL)
	if err != nil {
		__antithesis_instrumentation__.Notify(25139)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25140)
	}
	__antithesis_instrumentation__.Notify(25132)

	eventCh := make(chan streamingccl.Event)
	now := timeutil.Now()
	startWalltime := timeutil.Unix(0, checkpoint.WallTime)
	if startWalltime.After(now) {
		__antithesis_instrumentation__.Notify(25141)
		panic("cannot start random stream client event stream in the future")
	} else {
		__antithesis_instrumentation__.Notify(25142)
	}
	__antithesis_instrumentation__.Notify(25133)

	var partitionTableID int
	partitionTableID, err = strconv.Atoi(partitionURL.Host)
	if err != nil {
		__antithesis_instrumentation__.Notify(25143)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25144)
	}
	__antithesis_instrumentation__.Notify(25134)
	log.Infof(ctx, "producing kvs for metadata for table %d for tenant %d based on %q", partitionTableID, config.tenantID, spec)
	tableDesc, systemKVs, err := m.getDescriptorAndNamespaceKVForTableID(config, descpb.ID(partitionTableID))
	if err != nil {
		__antithesis_instrumentation__.Notify(25145)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(25146)
	}
	__antithesis_instrumentation__.Notify(25135)
	receiveFn := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(25147)
		defer close(eventCh)

		r := rand.New(rand.NewSource(timeutil.Now().UnixNano()))
		kvInterval := config.eventFrequency

		numKVEventsSinceLastResolved := 0

		rng, _ := randutil.NewPseudoRand()
		var dupKVEvent streamingccl.Event

		for {
			__antithesis_instrumentation__.Notify(25148)
			var event streamingccl.Event
			if numKVEventsSinceLastResolved == config.kvsPerCheckpoint {
				__antithesis_instrumentation__.Notify(25152)

				resolvedTime := timeutil.Now()
				hlcResolvedTime := hlc.Timestamp{WallTime: resolvedTime.UnixNano()}
				event = streamingccl.MakeCheckpointEvent(hlcResolvedTime)
				dupKVEvent = nil

				numKVEventsSinceLastResolved = 0
			} else {
				__antithesis_instrumentation__.Notify(25153)

				if len(systemKVs) > 0 {
					__antithesis_instrumentation__.Notify(25154)
					systemKV := systemKVs[0]
					systemKV.Value.Timestamp = hlc.Timestamp{WallTime: timeutil.Now().UnixNano()}
					event = streamingccl.MakeKVEvent(systemKV)
					systemKVs = systemKVs[1:]
				} else {
					__antithesis_instrumentation__.Notify(25155)
					numKVEventsSinceLastResolved++

					if rng.Float64() < config.dupProbability && func() bool {
						__antithesis_instrumentation__.Notify(25156)
						return dupKVEvent != nil == true
					}() == true {
						__antithesis_instrumentation__.Notify(25157)
						dupKV := dupKVEvent.GetKV()
						event = streamingccl.MakeKVEvent(*dupKV)
					} else {
						__antithesis_instrumentation__.Notify(25158)
						event = streamingccl.MakeKVEvent(makeRandomKey(r, config, tableDesc))
						dupKVEvent = event
					}
				}
			}
			__antithesis_instrumentation__.Notify(25149)

			select {
			case eventCh <- event:
				__antithesis_instrumentation__.Notify(25159)
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(25160)
				return ctx.Err()
			}
			__antithesis_instrumentation__.Notify(25150)

			func() {
				__antithesis_instrumentation__.Notify(25161)
				m.mu.Lock()
				defer m.mu.Unlock()

				if len(m.mu.interceptors) > 0 {
					__antithesis_instrumentation__.Notify(25162)
					for _, interceptor := range m.mu.interceptors {
						__antithesis_instrumentation__.Notify(25163)
						if interceptor != nil {
							__antithesis_instrumentation__.Notify(25164)
							interceptor(event, spec)
						} else {
							__antithesis_instrumentation__.Notify(25165)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(25166)
				}
			}()
			__antithesis_instrumentation__.Notify(25151)

			time.Sleep(kvInterval)
		}
	}
	__antithesis_instrumentation__.Notify(25136)

	return &randomStreamSubscription{
		receiveFn: receiveFn,
		eventCh:   eventCh,
	}, nil
}

type randomStreamSubscription struct {
	receiveFn func(ctx context.Context) error
	eventCh   chan streamingccl.Event
}

func (r *randomStreamSubscription) Subscribe(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(25167)
	return r.receiveFn(ctx)
}

func (r *randomStreamSubscription) Events() <-chan streamingccl.Event {
	__antithesis_instrumentation__.Notify(25168)
	return r.eventCh
}

func (r *randomStreamSubscription) Err() error {
	__antithesis_instrumentation__.Notify(25169)
	return nil
}

func rekey(tenantID roachpb.TenantID, k roachpb.Key) roachpb.Key {
	__antithesis_instrumentation__.Notify(25170)

	tenantPrefix := keys.MakeTenantPrefix(tenantID)
	noTenantPrefix, _, err := keys.DecodeTenantPrefix(k)
	if err != nil {
		__antithesis_instrumentation__.Notify(25172)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(25173)
	}
	__antithesis_instrumentation__.Notify(25171)

	rekeyedKey := append(tenantPrefix, noTenantPrefix...)
	return rekeyedKey
}

func makeRandomKey(
	r *rand.Rand, config randomStreamConfig, tableDesc *tabledesc.Mutable,
) roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(25174)

	keyDatum := tree.NewDInt(tree.DInt(r.Intn(config.valueRange)))

	index := tableDesc.GetPrimaryIndex()

	var colIDToRowIndex catalog.TableColMap
	colIDToRowIndex.Set(index.GetKeyColumnID(0), 0)

	keyPrefix := rowenc.MakeIndexKeyPrefix(keys.SystemSQLCodec, tableDesc.GetID(), index.GetID())
	k, _, err := rowenc.EncodeIndexKey(tableDesc, index, colIDToRowIndex, tree.Datums{keyDatum}, keyPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(25177)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(25178)
	}
	__antithesis_instrumentation__.Notify(25175)
	k = keys.MakeFamilyKey(k, uint32(tableDesc.Families[0].ID))

	k = rekey(config.tenantID, k)

	valueDatum := tree.NewDInt(tree.DInt(r.Intn(config.valueRange)))
	valueBuf, err := valueside.Encode(
		[]byte(nil), valueside.MakeColumnIDDelta(0, tableDesc.Columns[1].ID), valueDatum, []byte(nil))
	if err != nil {
		__antithesis_instrumentation__.Notify(25179)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(25180)
	}
	__antithesis_instrumentation__.Notify(25176)
	var v roachpb.Value
	v.SetTuple(valueBuf)
	v.ClearChecksum()
	v.InitChecksum(k)

	v.Timestamp = hlc.Timestamp{WallTime: timeutil.Now().UnixNano()}

	return roachpb.KeyValue{
		Key:   k,
		Value: v,
	}
}

func (m *randomStreamClient) RegisterInterception(fn InterceptFn) {
	__antithesis_instrumentation__.Notify(25181)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mu.interceptors = append(m.mu.interceptors, fn)
}
