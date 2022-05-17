package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

type kafkaLogAdapter struct {
	ctx context.Context
}

var _ sarama.StdLogger = (*kafkaLogAdapter)(nil)

func (l *kafkaLogAdapter) Print(v ...interface{}) {
	__antithesis_instrumentation__.Notify(18325)
	log.InfofDepth(l.ctx, 1, "", v...)
}
func (l *kafkaLogAdapter) Printf(format string, v ...interface{}) {
	__antithesis_instrumentation__.Notify(18326)
	log.InfofDepth(l.ctx, 1, format, v...)
}
func (l *kafkaLogAdapter) Println(v ...interface{}) {
	__antithesis_instrumentation__.Notify(18327)
	log.InfofDepth(l.ctx, 1, "", v...)
}

func init() {

	ctx := context.Background()
	ctx = logtags.AddTag(ctx, "kafka-producer", nil)
	sarama.Logger = &kafkaLogAdapter{ctx: ctx}

	sarama.MaxRequestSize = math.MaxInt32
}

type kafkaClient interface {
	Partitions(topic string) ([]int32, error)

	RefreshMetadata(topics ...string) error

	Close() error
}

type kafkaSink struct {
	ctx            context.Context
	bootstrapAddrs string
	kafkaCfg       *sarama.Config
	client         kafkaClient
	producer       sarama.AsyncProducer
	topics         *TopicNamer

	lastMetadataRefresh time.Time

	stopWorkerCh chan struct{}
	worker       sync.WaitGroup
	scratch      bufalloc.ByteAllocator
	metrics      *sliMetrics

	mu struct {
		syncutil.Mutex
		inflight int64
		flushErr error
		flushCh  chan struct{}
	}
}

type saramaConfig struct {
	Flush struct {
		Bytes       int          `json:",omitempty"`
		Messages    int          `json:",omitempty"`
		Frequency   jsonDuration `json:",omitempty"`
		MaxMessages int          `json:",omitempty"`
	}

	RequiredAcks string `json:",omitempty"`

	Version string `json:",omitempty"`
}

func (c saramaConfig) Validate() error {
	__antithesis_instrumentation__.Notify(18328)

	if (c.Flush.Bytes > 0 || func() bool {
		__antithesis_instrumentation__.Notify(18330)
		return c.Flush.Messages > 1 == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(18331)
		return c.Flush.Frequency == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(18332)
		return errors.New("Flush.Frequency must be > 0 when Flush.Bytes > 0 or Flush.Messages > 1")
	} else {
		__antithesis_instrumentation__.Notify(18333)
	}
	__antithesis_instrumentation__.Notify(18329)
	return nil
}

func defaultSaramaConfig() *saramaConfig {
	__antithesis_instrumentation__.Notify(18334)
	config := &saramaConfig{}

	config.Flush.Messages = 0
	config.Flush.Frequency = jsonDuration(0)
	config.Flush.Bytes = 0

	config.Flush.MaxMessages = 1000

	return config
}

func (s *kafkaSink) start() {
	__antithesis_instrumentation__.Notify(18335)
	s.stopWorkerCh = make(chan struct{})
	s.worker.Add(1)
	go s.workerLoop()
}

func (s *kafkaSink) Dial() error {
	__antithesis_instrumentation__.Notify(18336)
	client, err := sarama.NewClient(strings.Split(s.bootstrapAddrs, `,`), s.kafkaCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(18339)
		return pgerror.Wrapf(err, pgcode.CannotConnectNow,
			`connecting to kafka: %s`, s.bootstrapAddrs)
	} else {
		__antithesis_instrumentation__.Notify(18340)
	}
	__antithesis_instrumentation__.Notify(18337)
	s.producer, err = sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		__antithesis_instrumentation__.Notify(18341)
		return pgerror.Wrapf(err, pgcode.CannotConnectNow,
			`connecting to kafka: %s`, s.bootstrapAddrs)
	} else {
		__antithesis_instrumentation__.Notify(18342)
	}
	__antithesis_instrumentation__.Notify(18338)
	s.client = client
	s.start()
	return nil
}

func (s *kafkaSink) Close() error {
	__antithesis_instrumentation__.Notify(18343)
	close(s.stopWorkerCh)
	s.worker.Wait()

	_ = s.producer.Close()

	if s.client != nil {
		__antithesis_instrumentation__.Notify(18345)
		return s.client.Close()
	} else {
		__antithesis_instrumentation__.Notify(18346)
	}
	__antithesis_instrumentation__.Notify(18344)
	return nil
}

type messageMetadata struct {
	alloc         kvevent.Alloc
	updateMetrics recordOneMessageCallback
	mvcc          hlc.Timestamp
}

func (s *kafkaSink) EmitRow(
	ctx context.Context,
	topicDescr TopicDescriptor,
	key, value []byte,
	updated, mvcc hlc.Timestamp,
	alloc kvevent.Alloc,
) error {
	__antithesis_instrumentation__.Notify(18347)

	topic, err := s.topics.Name(topicDescr)
	if err != nil {
		__antithesis_instrumentation__.Notify(18349)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18350)
	}
	__antithesis_instrumentation__.Notify(18348)

	msg := &sarama.ProducerMessage{
		Topic:    topic,
		Key:      sarama.ByteEncoder(key),
		Value:    sarama.ByteEncoder(value),
		Metadata: messageMetadata{alloc: alloc, mvcc: mvcc, updateMetrics: s.metrics.recordOneMessage()},
	}
	return s.emitMessage(ctx, msg)
}

func (s *kafkaSink) EmitResolvedTimestamp(
	ctx context.Context, encoder Encoder, resolved hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(18351)
	defer s.metrics.recordResolvedCallback()()

	const metadataRefreshMinDuration = time.Minute
	if timeutil.Since(s.lastMetadataRefresh) > metadataRefreshMinDuration {
		__antithesis_instrumentation__.Notify(18353)
		if err := s.client.RefreshMetadata(s.topics.DisplayNamesSlice()...); err != nil {
			__antithesis_instrumentation__.Notify(18355)
			return err
		} else {
			__antithesis_instrumentation__.Notify(18356)
		}
		__antithesis_instrumentation__.Notify(18354)
		s.lastMetadataRefresh = timeutil.Now()
	} else {
		__antithesis_instrumentation__.Notify(18357)
	}
	__antithesis_instrumentation__.Notify(18352)

	return s.topics.Each(func(topic string) error {
		__antithesis_instrumentation__.Notify(18358)
		payload, err := encoder.EncodeResolvedTimestamp(ctx, topic, resolved)
		if err != nil {
			__antithesis_instrumentation__.Notify(18362)
			return err
		} else {
			__antithesis_instrumentation__.Notify(18363)
		}
		__antithesis_instrumentation__.Notify(18359)
		s.scratch, payload = s.scratch.Copy(payload, 0)

		partitions, err := s.client.Partitions(topic)
		if err != nil {
			__antithesis_instrumentation__.Notify(18364)
			return err
		} else {
			__antithesis_instrumentation__.Notify(18365)
		}
		__antithesis_instrumentation__.Notify(18360)
		for _, partition := range partitions {
			__antithesis_instrumentation__.Notify(18366)
			msg := &sarama.ProducerMessage{
				Topic:     topic,
				Partition: partition,
				Key:       nil,
				Value:     sarama.ByteEncoder(payload),
			}
			if err := s.emitMessage(ctx, msg); err != nil {
				__antithesis_instrumentation__.Notify(18367)
				return err
			} else {
				__antithesis_instrumentation__.Notify(18368)
			}
		}
		__antithesis_instrumentation__.Notify(18361)
		return nil
	})
}

func (s *kafkaSink) Flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(18369)
	defer s.metrics.recordFlushRequestCallback()()

	flushCh := make(chan struct{}, 1)

	s.mu.Lock()
	inflight := s.mu.inflight
	flushErr := s.mu.flushErr
	s.mu.flushErr = nil
	immediateFlush := inflight == 0 || func() bool {
		__antithesis_instrumentation__.Notify(18373)
		return flushErr != nil == true
	}() == true
	if !immediateFlush {
		__antithesis_instrumentation__.Notify(18374)
		s.mu.flushCh = flushCh
	} else {
		__antithesis_instrumentation__.Notify(18375)
	}
	__antithesis_instrumentation__.Notify(18370)
	s.mu.Unlock()

	if immediateFlush {
		__antithesis_instrumentation__.Notify(18376)
		return flushErr
	} else {
		__antithesis_instrumentation__.Notify(18377)
	}
	__antithesis_instrumentation__.Notify(18371)

	if log.V(1) {
		__antithesis_instrumentation__.Notify(18378)
		log.Infof(ctx, "flush waiting for %d inflight messages", inflight)
	} else {
		__antithesis_instrumentation__.Notify(18379)
	}
	__antithesis_instrumentation__.Notify(18372)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(18380)
		return ctx.Err()
	case <-flushCh:
		__antithesis_instrumentation__.Notify(18381)
		s.mu.Lock()
		flushErr := s.mu.flushErr
		s.mu.flushErr = nil
		s.mu.Unlock()
		return flushErr
	}
}

func (s *kafkaSink) startInflightMessage(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(18382)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mu.inflight++
	if log.V(2) {
		__antithesis_instrumentation__.Notify(18384)
		log.Infof(ctx, "emitting %d inflight records to kafka", s.mu.inflight)
	} else {
		__antithesis_instrumentation__.Notify(18385)
	}
	__antithesis_instrumentation__.Notify(18383)
	return nil
}

func (s *kafkaSink) emitMessage(ctx context.Context, msg *sarama.ProducerMessage) error {
	__antithesis_instrumentation__.Notify(18386)
	if err := s.startInflightMessage(ctx); err != nil {
		__antithesis_instrumentation__.Notify(18389)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18390)
	}
	__antithesis_instrumentation__.Notify(18387)

	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(18391)
		return ctx.Err()
	case s.producer.Input() <- msg:
		__antithesis_instrumentation__.Notify(18392)
	}
	__antithesis_instrumentation__.Notify(18388)

	return nil
}

func (s *kafkaSink) workerLoop() {
	__antithesis_instrumentation__.Notify(18393)
	defer s.worker.Done()

	for {
		__antithesis_instrumentation__.Notify(18394)
		var ackMsg *sarama.ProducerMessage
		var ackError error

		select {
		case <-s.stopWorkerCh:
			__antithesis_instrumentation__.Notify(18399)
			return
		case m := <-s.producer.Successes():
			__antithesis_instrumentation__.Notify(18400)
			ackMsg = m
		case err := <-s.producer.Errors():
			__antithesis_instrumentation__.Notify(18401)
			ackMsg, ackError = err.Msg, err.Err
			if ackError != nil {
				__antithesis_instrumentation__.Notify(18402)

				if err.Msg != nil && func() bool {
					__antithesis_instrumentation__.Notify(18403)
					return err.Msg.Key != nil == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(18404)
					return err.Msg.Value != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(18405)
					ackError = errors.Wrapf(ackError,
						"while sending message with key=%s, size=%d",
						err.Msg.Key, err.Msg.Key.Length()+err.Msg.Value.Length())
				} else {
					__antithesis_instrumentation__.Notify(18406)
				}
			} else {
				__antithesis_instrumentation__.Notify(18407)
			}
		}
		__antithesis_instrumentation__.Notify(18395)

		if m, ok := ackMsg.Metadata.(messageMetadata); ok {
			__antithesis_instrumentation__.Notify(18408)
			if ackError == nil {
				__antithesis_instrumentation__.Notify(18410)
				m.updateMetrics(m.mvcc, ackMsg.Key.Length()+ackMsg.Value.Length(), sinkDoesNotCompress)
			} else {
				__antithesis_instrumentation__.Notify(18411)
			}
			__antithesis_instrumentation__.Notify(18409)
			m.alloc.Release(s.ctx)
		} else {
			__antithesis_instrumentation__.Notify(18412)
		}
		__antithesis_instrumentation__.Notify(18396)

		s.mu.Lock()
		s.mu.inflight--
		if s.mu.flushErr == nil && func() bool {
			__antithesis_instrumentation__.Notify(18413)
			return ackError != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(18414)
			s.mu.flushErr = ackError
		} else {
			__antithesis_instrumentation__.Notify(18415)
		}
		__antithesis_instrumentation__.Notify(18397)

		if s.mu.inflight == 0 && func() bool {
			__antithesis_instrumentation__.Notify(18416)
			return s.mu.flushCh != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(18417)
			s.mu.flushCh <- struct{}{}
			s.mu.flushCh = nil
		} else {
			__antithesis_instrumentation__.Notify(18418)
		}
		__antithesis_instrumentation__.Notify(18398)
		s.mu.Unlock()
	}
}

func (s *kafkaSink) Topics() []string {
	__antithesis_instrumentation__.Notify(18419)
	return s.topics.DisplayNamesSlice()
}

type changefeedPartitioner struct {
	hash sarama.Partitioner
}

var _ sarama.Partitioner = &changefeedPartitioner{}
var _ sarama.PartitionerConstructor = newChangefeedPartitioner

func newChangefeedPartitioner(topic string) sarama.Partitioner {
	__antithesis_instrumentation__.Notify(18420)
	return &changefeedPartitioner{
		hash: sarama.NewHashPartitioner(topic),
	}
}

func (p *changefeedPartitioner) RequiresConsistency() bool {
	__antithesis_instrumentation__.Notify(18421)
	return true
}
func (p *changefeedPartitioner) Partition(
	message *sarama.ProducerMessage, numPartitions int32,
) (int32, error) {
	__antithesis_instrumentation__.Notify(18422)
	if message.Key == nil {
		__antithesis_instrumentation__.Notify(18424)
		return message.Partition, nil
	} else {
		__antithesis_instrumentation__.Notify(18425)
	}
	__antithesis_instrumentation__.Notify(18423)
	return p.hash.Partition(message, numPartitions)
}

type jsonDuration time.Duration

func (j *jsonDuration) UnmarshalJSON(b []byte) error {
	__antithesis_instrumentation__.Notify(18426)
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		__antithesis_instrumentation__.Notify(18429)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18430)
	}
	__antithesis_instrumentation__.Notify(18427)
	dur, err := time.ParseDuration(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(18431)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18432)
	}
	__antithesis_instrumentation__.Notify(18428)
	*j = jsonDuration(dur)
	return nil
}

func (c *saramaConfig) Apply(kafka *sarama.Config) error {
	__antithesis_instrumentation__.Notify(18433)

	kafka.Producer.MaxMessageBytes = int(sarama.MaxRequestSize - 1)

	kafka.Producer.Flush.Bytes = c.Flush.Bytes
	kafka.Producer.Flush.Messages = c.Flush.Messages
	kafka.Producer.Flush.Frequency = time.Duration(c.Flush.Frequency)
	kafka.Producer.Flush.MaxMessages = c.Flush.MaxMessages
	if c.Version != "" {
		__antithesis_instrumentation__.Notify(18436)
		parsedVersion, err := sarama.ParseKafkaVersion(c.Version)
		if err != nil {
			__antithesis_instrumentation__.Notify(18438)
			return err
		} else {
			__antithesis_instrumentation__.Notify(18439)
		}
		__antithesis_instrumentation__.Notify(18437)
		kafka.Version = parsedVersion
	} else {
		__antithesis_instrumentation__.Notify(18440)
	}
	__antithesis_instrumentation__.Notify(18434)
	if c.RequiredAcks != "" {
		__antithesis_instrumentation__.Notify(18441)
		parsedAcks, err := parseRequiredAcks(c.RequiredAcks)
		if err != nil {
			__antithesis_instrumentation__.Notify(18443)
			return err
		} else {
			__antithesis_instrumentation__.Notify(18444)
		}
		__antithesis_instrumentation__.Notify(18442)
		kafka.Producer.RequiredAcks = parsedAcks
	} else {
		__antithesis_instrumentation__.Notify(18445)
	}
	__antithesis_instrumentation__.Notify(18435)
	return nil
}

func parseRequiredAcks(a string) (sarama.RequiredAcks, error) {
	__antithesis_instrumentation__.Notify(18446)
	switch a {
	case "0", "NONE":
		__antithesis_instrumentation__.Notify(18447)
		return sarama.NoResponse, nil
	case "1", "ONE":
		__antithesis_instrumentation__.Notify(18448)
		return sarama.WaitForLocal, nil
	case "-1", "ALL":
		__antithesis_instrumentation__.Notify(18449)
		return sarama.WaitForAll, nil
	default:
		__antithesis_instrumentation__.Notify(18450)
		return sarama.WaitForLocal,
			fmt.Errorf(`invalid acks value "%s", must be "NONE"/"0", "ONE"/"1", or "ALL"/"-1"`, a)
	}
}

func getSaramaConfig(opts map[string]string) (config *saramaConfig, err error) {
	__antithesis_instrumentation__.Notify(18451)
	config = defaultSaramaConfig()
	if configStr, haveOverride := opts[changefeedbase.OptKafkaSinkConfig]; haveOverride {
		__antithesis_instrumentation__.Notify(18453)
		err = json.Unmarshal([]byte(configStr), config)
	} else {
		__antithesis_instrumentation__.Notify(18454)
	}
	__antithesis_instrumentation__.Notify(18452)
	return
}

func buildKafkaConfig(u sinkURL, opts map[string]string) (*sarama.Config, error) {
	__antithesis_instrumentation__.Notify(18455)
	dialConfig := struct {
		tlsEnabled    bool
		tlsSkipVerify bool
		caCert        []byte
		clientCert    []byte
		clientKey     []byte
		saslEnabled   bool
		saslHandshake bool
		saslUser      string
		saslPassword  string
		saslMechanism string
	}{}

	if _, err := u.consumeBool(changefeedbase.SinkParamTLSEnabled, &dialConfig.tlsEnabled); err != nil {
		__antithesis_instrumentation__.Notify(18472)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18473)
	}
	__antithesis_instrumentation__.Notify(18456)
	if _, err := u.consumeBool(changefeedbase.SinkParamSkipTLSVerify, &dialConfig.tlsSkipVerify); err != nil {
		__antithesis_instrumentation__.Notify(18474)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18475)
	}
	__antithesis_instrumentation__.Notify(18457)
	if err := u.decodeBase64(changefeedbase.SinkParamCACert, &dialConfig.caCert); err != nil {
		__antithesis_instrumentation__.Notify(18476)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18477)
	}
	__antithesis_instrumentation__.Notify(18458)
	if err := u.decodeBase64(changefeedbase.SinkParamClientCert, &dialConfig.clientCert); err != nil {
		__antithesis_instrumentation__.Notify(18478)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18479)
	}
	__antithesis_instrumentation__.Notify(18459)
	if err := u.decodeBase64(changefeedbase.SinkParamClientKey, &dialConfig.clientKey); err != nil {
		__antithesis_instrumentation__.Notify(18480)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18481)
	}
	__antithesis_instrumentation__.Notify(18460)

	if _, err := u.consumeBool(changefeedbase.SinkParamSASLEnabled, &dialConfig.saslEnabled); err != nil {
		__antithesis_instrumentation__.Notify(18482)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18483)
	}
	__antithesis_instrumentation__.Notify(18461)

	if wasSet, err := u.consumeBool(changefeedbase.SinkParamSASLHandshake, &dialConfig.saslHandshake); !wasSet && func() bool {
		__antithesis_instrumentation__.Notify(18484)
		return err == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(18485)
		dialConfig.saslHandshake = true
	} else {
		__antithesis_instrumentation__.Notify(18486)
		if err != nil {
			__antithesis_instrumentation__.Notify(18488)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(18489)
		}
		__antithesis_instrumentation__.Notify(18487)
		if !dialConfig.saslEnabled {
			__antithesis_instrumentation__.Notify(18490)
			return nil, errors.Errorf(`%s must be enabled to configure SASL handshake behavior`, changefeedbase.SinkParamSASLEnabled)
		} else {
			__antithesis_instrumentation__.Notify(18491)
		}
	}
	__antithesis_instrumentation__.Notify(18462)

	dialConfig.saslMechanism = u.consumeParam(changefeedbase.SinkParamSASLMechanism)
	if dialConfig.saslMechanism != `` && func() bool {
		__antithesis_instrumentation__.Notify(18492)
		return !dialConfig.saslEnabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(18493)
		return nil, errors.Errorf(`%s must be enabled to configure SASL mechanism`, changefeedbase.SinkParamSASLEnabled)
	} else {
		__antithesis_instrumentation__.Notify(18494)
	}
	__antithesis_instrumentation__.Notify(18463)
	if dialConfig.saslMechanism == `` {
		__antithesis_instrumentation__.Notify(18495)
		dialConfig.saslMechanism = sarama.SASLTypePlaintext
	} else {
		__antithesis_instrumentation__.Notify(18496)
	}
	__antithesis_instrumentation__.Notify(18464)
	switch dialConfig.saslMechanism {
	case sarama.SASLTypeSCRAMSHA256, sarama.SASLTypeSCRAMSHA512, sarama.SASLTypePlaintext:
		__antithesis_instrumentation__.Notify(18497)
	default:
		__antithesis_instrumentation__.Notify(18498)
		return nil, errors.Errorf(`param %s must be one of %s, %s, or %s`,
			changefeedbase.SinkParamSASLMechanism,
			sarama.SASLTypeSCRAMSHA256, sarama.SASLTypeSCRAMSHA512, sarama.SASLTypePlaintext)
	}
	__antithesis_instrumentation__.Notify(18465)

	dialConfig.saslUser = u.consumeParam(changefeedbase.SinkParamSASLUser)
	dialConfig.saslPassword = u.consumeParam(changefeedbase.SinkParamSASLPassword)
	if dialConfig.saslEnabled {
		__antithesis_instrumentation__.Notify(18499)
		if dialConfig.saslUser == `` {
			__antithesis_instrumentation__.Notify(18501)
			return nil, errors.Errorf(`%s must be provided when SASL is enabled`, changefeedbase.SinkParamSASLUser)
		} else {
			__antithesis_instrumentation__.Notify(18502)
		}
		__antithesis_instrumentation__.Notify(18500)
		if dialConfig.saslPassword == `` {
			__antithesis_instrumentation__.Notify(18503)
			return nil, errors.Errorf(`%s must be provided when SASL is enabled`, changefeedbase.SinkParamSASLPassword)
		} else {
			__antithesis_instrumentation__.Notify(18504)
		}
	} else {
		__antithesis_instrumentation__.Notify(18505)
		if dialConfig.saslUser != `` {
			__antithesis_instrumentation__.Notify(18507)
			return nil, errors.Errorf(`%s must be enabled if a SASL user is provided`, changefeedbase.SinkParamSASLEnabled)
		} else {
			__antithesis_instrumentation__.Notify(18508)
		}
		__antithesis_instrumentation__.Notify(18506)
		if dialConfig.saslPassword != `` {
			__antithesis_instrumentation__.Notify(18509)
			return nil, errors.Errorf(`%s must be enabled if a SASL password is provided`, changefeedbase.SinkParamSASLEnabled)
		} else {
			__antithesis_instrumentation__.Notify(18510)
		}
	}
	__antithesis_instrumentation__.Notify(18466)

	config := sarama.NewConfig()
	config.ClientID = `CockroachDB`
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = newChangefeedPartitioner

	if dialConfig.tlsEnabled {
		__antithesis_instrumentation__.Notify(18511)
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: dialConfig.tlsSkipVerify,
		}

		if dialConfig.caCert != nil {
			__antithesis_instrumentation__.Notify(18514)
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(dialConfig.caCert)
			config.Net.TLS.Config.RootCAs = caCertPool
		} else {
			__antithesis_instrumentation__.Notify(18515)
		}
		__antithesis_instrumentation__.Notify(18512)

		if dialConfig.clientCert != nil && func() bool {
			__antithesis_instrumentation__.Notify(18516)
			return dialConfig.clientKey == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(18517)
			return nil, errors.Errorf(`%s requires %s to be set`, changefeedbase.SinkParamClientCert, changefeedbase.SinkParamClientKey)
		} else {
			__antithesis_instrumentation__.Notify(18518)
			if dialConfig.clientKey != nil && func() bool {
				__antithesis_instrumentation__.Notify(18519)
				return dialConfig.clientCert == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(18520)
				return nil, errors.Errorf(`%s requires %s to be set`, changefeedbase.SinkParamClientKey, changefeedbase.SinkParamClientCert)
			} else {
				__antithesis_instrumentation__.Notify(18521)
			}
		}
		__antithesis_instrumentation__.Notify(18513)

		if dialConfig.clientCert != nil && func() bool {
			__antithesis_instrumentation__.Notify(18522)
			return dialConfig.clientKey != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(18523)
			cert, err := tls.X509KeyPair(dialConfig.clientCert, dialConfig.clientKey)
			if err != nil {
				__antithesis_instrumentation__.Notify(18525)
				return nil, errors.Wrap(err, `invalid client certificate data provided`)
			} else {
				__antithesis_instrumentation__.Notify(18526)
			}
			__antithesis_instrumentation__.Notify(18524)
			config.Net.TLS.Config.Certificates = []tls.Certificate{cert}
		} else {
			__antithesis_instrumentation__.Notify(18527)
		}
	} else {
		__antithesis_instrumentation__.Notify(18528)
		if dialConfig.caCert != nil {
			__antithesis_instrumentation__.Notify(18530)
			return nil, errors.Errorf(`%s requires %s=true`, changefeedbase.SinkParamCACert, changefeedbase.SinkParamTLSEnabled)
		} else {
			__antithesis_instrumentation__.Notify(18531)
		}
		__antithesis_instrumentation__.Notify(18529)
		if dialConfig.clientCert != nil {
			__antithesis_instrumentation__.Notify(18532)
			return nil, errors.Errorf(`%s requires %s=true`, changefeedbase.SinkParamClientCert, changefeedbase.SinkParamTLSEnabled)
		} else {
			__antithesis_instrumentation__.Notify(18533)
		}
	}
	__antithesis_instrumentation__.Notify(18467)

	if dialConfig.saslEnabled {
		__antithesis_instrumentation__.Notify(18534)
		config.Net.SASL.Enable = true
		config.Net.SASL.Handshake = dialConfig.saslHandshake
		config.Net.SASL.User = dialConfig.saslUser
		config.Net.SASL.Password = dialConfig.saslPassword
		config.Net.SASL.Mechanism = sarama.SASLMechanism(dialConfig.saslMechanism)
		switch config.Net.SASL.Mechanism {
		case sarama.SASLTypeSCRAMSHA512:
			__antithesis_instrumentation__.Notify(18535)
			config.Net.SASL.SCRAMClientGeneratorFunc = sha512ClientGenerator
		case sarama.SASLTypeSCRAMSHA256:
			__antithesis_instrumentation__.Notify(18536)
			config.Net.SASL.SCRAMClientGeneratorFunc = sha256ClientGenerator
		default:
			__antithesis_instrumentation__.Notify(18537)
		}
	} else {
		__antithesis_instrumentation__.Notify(18538)
	}
	__antithesis_instrumentation__.Notify(18468)

	saramaCfg, err := getSaramaConfig(opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(18539)
		return nil, errors.Wrapf(err,
			"failed to parse sarama config; check %s option", changefeedbase.OptKafkaSinkConfig)
	} else {
		__antithesis_instrumentation__.Notify(18540)
	}
	__antithesis_instrumentation__.Notify(18469)

	if err := saramaCfg.Validate(); err != nil {
		__antithesis_instrumentation__.Notify(18541)
		return nil, errors.Wrap(err, "invalid sarama configuration")
	} else {
		__antithesis_instrumentation__.Notify(18542)
	}
	__antithesis_instrumentation__.Notify(18470)

	if err := saramaCfg.Apply(config); err != nil {
		__antithesis_instrumentation__.Notify(18543)
		return nil, errors.Wrap(err, "failed to apply kafka client configuration")
	} else {
		__antithesis_instrumentation__.Notify(18544)
	}
	__antithesis_instrumentation__.Notify(18471)
	return config, nil
}

func makeKafkaSink(
	ctx context.Context,
	u sinkURL,
	targets []jobspb.ChangefeedTargetSpecification,
	opts map[string]string,
	m *sliMetrics,
) (Sink, error) {
	__antithesis_instrumentation__.Notify(18545)
	kafkaTopicPrefix := u.consumeParam(changefeedbase.SinkParamTopicPrefix)
	kafkaTopicName := u.consumeParam(changefeedbase.SinkParamTopicName)
	if schemaTopic := u.consumeParam(changefeedbase.SinkParamSchemaTopic); schemaTopic != `` {
		__antithesis_instrumentation__.Notify(18550)
		return nil, errors.Errorf(`%s is not yet supported`, changefeedbase.SinkParamSchemaTopic)
	} else {
		__antithesis_instrumentation__.Notify(18551)
	}
	__antithesis_instrumentation__.Notify(18546)

	config, err := buildKafkaConfig(u, opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(18552)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18553)
	}
	__antithesis_instrumentation__.Notify(18547)

	topics, err := MakeTopicNamer(
		targets,
		WithPrefix(kafkaTopicPrefix), WithSingleName(kafkaTopicName), WithSanitizeFn(SQLNameToKafkaName))

	if err != nil {
		__antithesis_instrumentation__.Notify(18554)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18555)
	}
	__antithesis_instrumentation__.Notify(18548)

	sink := &kafkaSink{
		ctx:            ctx,
		kafkaCfg:       config,
		bootstrapAddrs: u.Host,
		metrics:        m,
		topics:         topics,
	}

	if unknownParams := u.remainingQueryParams(); len(unknownParams) > 0 {
		__antithesis_instrumentation__.Notify(18556)
		return nil, errors.Errorf(
			`unknown kafka sink query parameters: %s`, strings.Join(unknownParams, ", "))
	} else {
		__antithesis_instrumentation__.Notify(18557)
	}
	__antithesis_instrumentation__.Notify(18549)

	return sink, nil
}
