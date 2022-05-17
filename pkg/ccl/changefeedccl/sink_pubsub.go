package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/url"

	"cloud.google.com/go/pubsub"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const credentialsParam = "CREDENTIALS"

const GcpScheme = "gcpubsub"
const gcpScope = "https://www.googleapis.com/auth/pubsub"

const numOfWorkers = 128

func isPubsubSink(u *url.URL) bool {
	__antithesis_instrumentation__.Notify(18558)
	return u.Scheme == GcpScheme
}

type pubsubClient interface {
	init() error
	closeTopics()
	flushTopics()
	sendMessage(content []byte, topic string, key string) error
	sendMessageToAllTopics(content []byte) error
}

type payload struct {
	Key   json.RawMessage `json:"key"`
	Value json.RawMessage `json:"value"`
	Topic string          `json:"topic"`
}

type pubsubMessage struct {
	alloc   kvevent.Alloc
	message payload
	isFlush bool
}

type gcpPubsubClient struct {
	client     *pubsub.Client
	topics     map[string]*pubsub.Topic
	ctx        context.Context
	projectID  string
	region     string
	topicNamer *TopicNamer
	url        sinkURL
}

type pubsubSink struct {
	numWorkers int

	workerCtx   context.Context
	workerGroup ctxgroup.Group

	exitWorkers func()
	eventsChans []chan pubsubMessage

	flushDone chan struct{}

	errChan chan error

	client     pubsubClient
	topicNamer *TopicNamer
}

func getGCPCredentials(ctx context.Context, u sinkURL) (*google.Credentials, error) {
	__antithesis_instrumentation__.Notify(18559)
	const authParam = "AUTH"
	const authSpecified = "specified"
	const authImplicit = "implicit"
	const authDefault = "default"

	var credsJSON []byte
	var creds *google.Credentials
	var err error
	authOption := u.consumeParam(authParam)

	switch authOption {
	case authImplicit:
		__antithesis_instrumentation__.Notify(18560)
		creds, err = google.FindDefaultCredentials(ctx, gcpScope)
		if err != nil {
			__antithesis_instrumentation__.Notify(18567)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(18568)
		}
		__antithesis_instrumentation__.Notify(18561)
		return creds, nil
	case authSpecified:
		__antithesis_instrumentation__.Notify(18562)
		fallthrough
	case authDefault:
		__antithesis_instrumentation__.Notify(18563)
		fallthrough
	default:
		__antithesis_instrumentation__.Notify(18564)
		err := u.decodeBase64(credentialsParam, &credsJSON)
		if err != nil {
			__antithesis_instrumentation__.Notify(18569)
			return nil, errors.Wrap(err, "decoding credentials json")
		} else {
			__antithesis_instrumentation__.Notify(18570)
		}
		__antithesis_instrumentation__.Notify(18565)
		creds, err = google.CredentialsFromJSON(ctx, credsJSON, gcpScope)
		if err != nil {
			__antithesis_instrumentation__.Notify(18571)
			return nil, errors.Wrap(err, "creating credentials")
		} else {
			__antithesis_instrumentation__.Notify(18572)
		}
		__antithesis_instrumentation__.Notify(18566)
		return creds, nil
	}
}

func MakePubsubSink(
	ctx context.Context,
	u *url.URL,
	opts map[string]string,
	targets []jobspb.ChangefeedTargetSpecification,
) (Sink, error) {
	__antithesis_instrumentation__.Notify(18573)

	pubsubURL := sinkURL{URL: u, q: u.Query()}
	pubsubTopicName := pubsubURL.consumeParam(changefeedbase.SinkParamTopicName)

	switch changefeedbase.FormatType(opts[changefeedbase.OptFormat]) {
	case changefeedbase.OptFormatJSON:
		__antithesis_instrumentation__.Notify(18576)
	default:
		__antithesis_instrumentation__.Notify(18577)
		return nil, errors.Errorf(`this sink is incompatible with %s=%s`,
			changefeedbase.OptFormat, opts[changefeedbase.OptFormat])
	}
	__antithesis_instrumentation__.Notify(18574)

	switch changefeedbase.EnvelopeType(opts[changefeedbase.OptEnvelope]) {
	case changefeedbase.OptEnvelopeWrapped:
		__antithesis_instrumentation__.Notify(18578)
	default:
		__antithesis_instrumentation__.Notify(18579)
		return nil, errors.Errorf(`this sink is incompatible with %s=%s`,
			changefeedbase.OptEnvelope, opts[changefeedbase.OptEnvelope])
	}
	__antithesis_instrumentation__.Notify(18575)

	ctx, cancel := context.WithCancel(ctx)
	p := &pubsubSink{
		workerCtx:   ctx,
		numWorkers:  numOfWorkers,
		exitWorkers: cancel,
	}

	switch u.Scheme {
	case GcpScheme:
		__antithesis_instrumentation__.Notify(18580)
		const regionParam = "region"
		projectID := pubsubURL.Host
		region := pubsubURL.consumeParam(regionParam)
		if region == "" {
			__antithesis_instrumentation__.Notify(18584)
			return nil, errors.New("region query parameter not found")
		} else {
			__antithesis_instrumentation__.Notify(18585)
		}
		__antithesis_instrumentation__.Notify(18581)
		tn, err := MakeTopicNamer(targets, WithSingleName(pubsubTopicName))
		if err != nil {
			__antithesis_instrumentation__.Notify(18586)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(18587)
		}
		__antithesis_instrumentation__.Notify(18582)
		g := &gcpPubsubClient{
			topicNamer: tn,
			ctx:        ctx,
			projectID:  projectID,
			region:     gcpEndpointForRegion(region),
			url:        pubsubURL,
		}
		p.client = g
		p.topicNamer = tn
		return p, nil
	default:
		__antithesis_instrumentation__.Notify(18583)
		return nil, errors.Errorf("unknown scheme: %s", u.Scheme)
	}
}

func (p *pubsubSink) Dial() error {
	__antithesis_instrumentation__.Notify(18588)
	p.setupWorkers()
	return p.client.init()
}

func (p *pubsubSink) EmitRow(
	ctx context.Context,
	topic TopicDescriptor,
	key, value []byte,
	updated hlc.Timestamp,
	mvcc hlc.Timestamp,
	alloc kvevent.Alloc,
) error {
	__antithesis_instrumentation__.Notify(18589)
	topicName, err := p.topicNamer.Name(topic)
	if err != nil {
		__antithesis_instrumentation__.Notify(18592)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18593)
	}
	__antithesis_instrumentation__.Notify(18590)
	m := pubsubMessage{
		alloc: alloc, isFlush: false, message: payload{
			Key:   key,
			Value: value,
			Topic: topicName,
		}}

	i := p.workerIndex(key)
	select {

	case <-p.workerCtx.Done():
		__antithesis_instrumentation__.Notify(18594)

		return errors.CombineErrors(p.workerCtx.Err(), p.sinkError())
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(18595)
		return ctx.Err()
	case err := <-p.errChan:
		__antithesis_instrumentation__.Notify(18596)
		return err
	case p.eventsChans[i] <- m:
		__antithesis_instrumentation__.Notify(18597)
	}
	__antithesis_instrumentation__.Notify(18591)
	return nil
}

func (p *pubsubSink) EmitResolvedTimestamp(
	ctx context.Context, encoder Encoder, resolved hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(18598)
	payload, err := encoder.EncodeResolvedTimestamp(ctx, "", resolved)
	if err != nil {
		__antithesis_instrumentation__.Notify(18600)
		return errors.Wrap(err, "encoding resolved timestamp")
	} else {
		__antithesis_instrumentation__.Notify(18601)
	}
	__antithesis_instrumentation__.Notify(18599)

	return p.client.sendMessageToAllTopics(payload)
}

func (p *pubsubSink) Flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(18602)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(18604)
		return ctx.Err()
	case err := <-p.errChan:
		__antithesis_instrumentation__.Notify(18605)
		return err
	default:
		__antithesis_instrumentation__.Notify(18606)
		err := p.flushWorkers()
		if err != nil {
			__antithesis_instrumentation__.Notify(18607)
			return err
		} else {
			__antithesis_instrumentation__.Notify(18608)
		}
	}
	__antithesis_instrumentation__.Notify(18603)

	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(18609)
		return ctx.Err()
	case err := <-p.errChan:
		__antithesis_instrumentation__.Notify(18610)
		return err
	case <-p.flushDone:
		__antithesis_instrumentation__.Notify(18611)
		return p.sinkError()
	}

}

func (p *pubsubSink) Close() error {
	__antithesis_instrumentation__.Notify(18612)
	p.client.closeTopics()
	p.exitWorkers()
	_ = p.workerGroup.Wait()
	if p.errChan != nil {
		__antithesis_instrumentation__.Notify(18616)
		close(p.errChan)
	} else {
		__antithesis_instrumentation__.Notify(18617)
	}
	__antithesis_instrumentation__.Notify(18613)
	if p.flushDone != nil {
		__antithesis_instrumentation__.Notify(18618)
		close(p.flushDone)
	} else {
		__antithesis_instrumentation__.Notify(18619)
	}
	__antithesis_instrumentation__.Notify(18614)
	for i := 0; i < p.numWorkers; i++ {
		__antithesis_instrumentation__.Notify(18620)
		if p.eventsChans[i] != nil {
			__antithesis_instrumentation__.Notify(18621)
			close(p.eventsChans[i])
		} else {
			__antithesis_instrumentation__.Notify(18622)
		}
	}
	__antithesis_instrumentation__.Notify(18615)
	return nil
}

func (p *pubsubSink) Topics() []string {
	__antithesis_instrumentation__.Notify(18623)
	return p.topicNamer.DisplayNamesSlice()
}

func (p *gcpPubsubClient) getTopicClient(name string) (*pubsub.Topic, error) {
	__antithesis_instrumentation__.Notify(18624)
	if topic, ok := p.topics[name]; ok {
		__antithesis_instrumentation__.Notify(18627)
		return topic, nil
	} else {
		__antithesis_instrumentation__.Notify(18628)
	}
	__antithesis_instrumentation__.Notify(18625)
	topic, err := p.openTopic(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(18629)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18630)
	}
	__antithesis_instrumentation__.Notify(18626)
	p.topics[name] = topic
	return topic, nil
}

func (p *pubsubSink) setupWorkers() {
	__antithesis_instrumentation__.Notify(18631)

	p.eventsChans = make([]chan pubsubMessage, p.numWorkers)
	p.workerGroup = ctxgroup.WithContext(p.workerCtx)

	p.errChan = make(chan error, 1)

	p.flushDone = make(chan struct{}, 1)

	for i := 0; i < p.numWorkers; i++ {
		__antithesis_instrumentation__.Notify(18632)

		p.eventsChans[i] = make(chan pubsubMessage)
		j := i
		p.workerGroup.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(18633)
			p.workerLoop(j)
			return nil
		})
	}
}

func (p *pubsubSink) workerLoop(workerIndex int) {
	__antithesis_instrumentation__.Notify(18634)
	for {
		__antithesis_instrumentation__.Notify(18635)
		select {
		case <-p.workerCtx.Done():
			__antithesis_instrumentation__.Notify(18636)
			return
		case msg := <-p.eventsChans[workerIndex]:
			__antithesis_instrumentation__.Notify(18637)
			if msg.isFlush {
				__antithesis_instrumentation__.Notify(18641)

				continue
			} else {
				__antithesis_instrumentation__.Notify(18642)
			}
			__antithesis_instrumentation__.Notify(18638)
			m := msg.message
			b, err := json.Marshal(m)
			if err != nil {
				__antithesis_instrumentation__.Notify(18643)
				p.exitWorkersWithError(err)
			} else {
				__antithesis_instrumentation__.Notify(18644)
			}
			__antithesis_instrumentation__.Notify(18639)
			err = p.client.sendMessage(b, msg.message.Topic, string(msg.message.Key))
			if err != nil {
				__antithesis_instrumentation__.Notify(18645)
				p.exitWorkersWithError(err)
			} else {
				__antithesis_instrumentation__.Notify(18646)
			}
			__antithesis_instrumentation__.Notify(18640)
			msg.alloc.Release(p.workerCtx)
		}
	}
}

func (p *pubsubSink) exitWorkersWithError(err error) {
	__antithesis_instrumentation__.Notify(18647)

	select {
	case p.errChan <- err:
		__antithesis_instrumentation__.Notify(18648)
		p.exitWorkers()
	default:
		__antithesis_instrumentation__.Notify(18649)
	}
}

func (p *pubsubSink) sinkError() error {
	__antithesis_instrumentation__.Notify(18650)
	select {
	case err := <-p.errChan:
		__antithesis_instrumentation__.Notify(18652)
		return err
	default:
		__antithesis_instrumentation__.Notify(18653)
	}
	__antithesis_instrumentation__.Notify(18651)
	return nil
}

func (p *pubsubSink) workerIndex(key []byte) uint32 {
	__antithesis_instrumentation__.Notify(18654)
	return crc32.ChecksumIEEE(key) % uint32(p.numWorkers)
}

func (p *pubsubSink) flushWorkers() error {
	__antithesis_instrumentation__.Notify(18655)
	for i := 0; i < p.numWorkers; i++ {
		__antithesis_instrumentation__.Notify(18657)

		select {
		case <-p.workerCtx.Done():
			__antithesis_instrumentation__.Notify(18658)
			return p.workerCtx.Err()
		case p.eventsChans[i] <- pubsubMessage{isFlush: true}:
			__antithesis_instrumentation__.Notify(18659)
		}
	}
	__antithesis_instrumentation__.Notify(18656)

	p.client.flushTopics()

	select {

	case <-p.workerCtx.Done():
		__antithesis_instrumentation__.Notify(18660)
		return p.workerCtx.Err()
	case p.flushDone <- struct{}{}:
		__antithesis_instrumentation__.Notify(18661)
		return nil
	}
}

func (p *gcpPubsubClient) init() error {
	var client *pubsub.Client
	var err error

	creds, err := getGCPCredentials(p.ctx, p.url)
	if err != nil {
		return err
	}

	client, err = pubsub.NewClient(
		p.ctx,
		p.projectID,
		option.WithCredentials(creds),
		option.WithEndpoint(p.region),
	)

	if err != nil {
		return errors.Wrap(err, "opening client")
	}
	p.client = client
	p.topics = make(map[string]*pubsub.Topic)

	return nil

}

func (p *gcpPubsubClient) openTopic(topicName string) (*pubsub.Topic, error) {
	__antithesis_instrumentation__.Notify(18662)
	t, err := p.client.CreateTopic(p.ctx, topicName)
	if err != nil {
		__antithesis_instrumentation__.Notify(18664)
		if status.Code(err) == codes.AlreadyExists {
			__antithesis_instrumentation__.Notify(18665)
			t = p.client.Topic(topicName)
		} else {
			__antithesis_instrumentation__.Notify(18666)
			return nil, err
		}
	} else {
		__antithesis_instrumentation__.Notify(18667)
	}
	__antithesis_instrumentation__.Notify(18663)
	t.EnableMessageOrdering = true
	return t, nil
}

func (p *gcpPubsubClient) closeTopics() {
	__antithesis_instrumentation__.Notify(18668)
	_ = p.forEachTopic(func(_ string, t *pubsub.Topic) error {
		__antithesis_instrumentation__.Notify(18669)
		t.Stop()
		return nil
	})
}

func (p *gcpPubsubClient) sendMessage(m []byte, topic string, key string) error {
	__antithesis_instrumentation__.Notify(18670)
	t, err := p.getTopicClient(topic)
	if err != nil {
		__antithesis_instrumentation__.Notify(18673)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18674)
	}
	__antithesis_instrumentation__.Notify(18671)
	res := t.Publish(p.ctx, &pubsub.Message{
		Data:        m,
		OrderingKey: key,
	})

	_, err = res.Get(p.ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(18675)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18676)
	}
	__antithesis_instrumentation__.Notify(18672)

	return nil
}

func (p *gcpPubsubClient) sendMessageToAllTopics(m []byte) error {
	__antithesis_instrumentation__.Notify(18677)
	return p.forEachTopic(func(_ string, t *pubsub.Topic) error {
		__antithesis_instrumentation__.Notify(18678)
		res := t.Publish(p.ctx, &pubsub.Message{
			Data: m,
		})
		_, err := res.Get(p.ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(18680)
			return errors.Wrap(err, "emitting resolved timestamp")
		} else {
			__antithesis_instrumentation__.Notify(18681)
		}
		__antithesis_instrumentation__.Notify(18679)
		return nil
	})
}

func (p *gcpPubsubClient) flushTopics() {
	__antithesis_instrumentation__.Notify(18682)
	_ = p.forEachTopic(func(_ string, t *pubsub.Topic) error {
		__antithesis_instrumentation__.Notify(18683)
		t.Flush()
		return nil
	})
}

func (p *gcpPubsubClient) forEachTopic(f func(name string, topicClient *pubsub.Topic) error) error {
	__antithesis_instrumentation__.Notify(18684)
	return p.topicNamer.Each(func(n string) error {
		__antithesis_instrumentation__.Notify(18685)
		t, err := p.getTopicClient(n)
		if err != nil {
			__antithesis_instrumentation__.Notify(18687)
			return err
		} else {
			__antithesis_instrumentation__.Notify(18688)
		}
		__antithesis_instrumentation__.Notify(18686)
		return f(n, t)
	})
}

func gcpEndpointForRegion(region string) string {
	__antithesis_instrumentation__.Notify(18689)
	return fmt.Sprintf("%s-pubsub.googleapis.com:443", region)
}
