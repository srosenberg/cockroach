package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/system"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const (
	applicationTypeJSON = `application/json`
	authorizationHeader = `Authorization`
	defaultConnTimeout  = 3 * time.Second
)

func isWebhookSink(u *url.URL) bool {
	__antithesis_instrumentation__.Notify(18746)
	switch u.Scheme {

	case changefeedbase.SinkSchemeWebhookHTTP, changefeedbase.SinkSchemeWebhookHTTPS:
		__antithesis_instrumentation__.Notify(18747)
		return true
	default:
		__antithesis_instrumentation__.Notify(18748)
		return false
	}
}

type webhookSink struct {
	parallelism int
	retryCfg    retry.Options
	batchCfg    batchConfig
	ts          timeutil.TimeSource

	url        sinkURL
	authHeader string
	client     *httputil.Client

	batchChan chan webhookMessage

	flushDone chan struct{}

	errChan chan error

	workerCtx   context.Context
	workerGroup ctxgroup.Group
	exitWorkers func()
	eventsChans []chan []messagePayload
	metrics     *sliMetrics
}

type webhookSinkPayload struct {
	Payload []json.RawMessage `json:"payload"`
	Length  int               `json:"length"`
}

type encodedPayload struct {
	data     []byte
	alloc    kvevent.Alloc
	emitTime time.Time
	mvcc     hlc.Timestamp
}

func encodePayloadWebhook(messages []messagePayload) (encodedPayload, error) {
	__antithesis_instrumentation__.Notify(18749)
	result := encodedPayload{
		emitTime: timeutil.Now(),
	}

	payload := make([]json.RawMessage, len(messages))
	for i, m := range messages {
		__antithesis_instrumentation__.Notify(18752)
		result.alloc.Merge(&m.alloc)
		payload[i] = m.val
		if m.emitTime.Before(result.emitTime) {
			__antithesis_instrumentation__.Notify(18754)
			result.emitTime = m.emitTime
		} else {
			__antithesis_instrumentation__.Notify(18755)
		}
		__antithesis_instrumentation__.Notify(18753)
		if result.mvcc.IsEmpty() || func() bool {
			__antithesis_instrumentation__.Notify(18756)
			return m.mvcc.Less(result.mvcc) == true
		}() == true {
			__antithesis_instrumentation__.Notify(18757)
			result.mvcc = m.mvcc
		} else {
			__antithesis_instrumentation__.Notify(18758)
		}
	}
	__antithesis_instrumentation__.Notify(18750)

	body := &webhookSinkPayload{
		Payload: payload,
		Length:  len(payload),
	}
	j, err := json.Marshal(body)
	if err != nil {
		__antithesis_instrumentation__.Notify(18759)
		return encodedPayload{}, err
	} else {
		__antithesis_instrumentation__.Notify(18760)
	}
	__antithesis_instrumentation__.Notify(18751)
	result.data = j
	return result, err
}

type messagePayload struct {
	key      []byte
	val      []byte
	alloc    kvevent.Alloc
	emitTime time.Time
	mvcc     hlc.Timestamp
}

type webhookMessage struct {
	flushDone *chan struct{}
	payload   messagePayload
}

type batch struct {
	buffer      []messagePayload
	bufferBytes int
}

func (b *batch) addToBuffer(m messagePayload) {
	__antithesis_instrumentation__.Notify(18761)
	b.bufferBytes += len(m.val)
	b.buffer = append(b.buffer, m)
}

func (b *batch) reset() {
	__antithesis_instrumentation__.Notify(18762)
	b.buffer = b.buffer[:0]
	b.bufferBytes = 0
}

type batchConfig struct {
	Bytes, Messages int          `json:",omitempty"`
	Frequency       jsonDuration `json:",omitempty"`
}

type jsonMaxRetries int

func (j *jsonMaxRetries) UnmarshalJSON(b []byte) error {
	__antithesis_instrumentation__.Notify(18763)
	var i int64

	i, err := strconv.ParseInt(string(b), 10, 64)
	if err == nil {
		__antithesis_instrumentation__.Notify(18765)
		if i <= 0 {
			__antithesis_instrumentation__.Notify(18767)
			return errors.Errorf("max retry count must be a positive integer. use 'inf' for infinite retries.")
		} else {
			__antithesis_instrumentation__.Notify(18768)
		}
		__antithesis_instrumentation__.Notify(18766)
		*j = jsonMaxRetries(i)
	} else {
		__antithesis_instrumentation__.Notify(18769)

		var s string

		if err := json.Unmarshal(b, &s); err != nil {
			__antithesis_instrumentation__.Notify(18771)
			return err
		} else {
			__antithesis_instrumentation__.Notify(18772)
		}
		__antithesis_instrumentation__.Notify(18770)
		if strings.ToLower(s) == "inf" {
			__antithesis_instrumentation__.Notify(18773)

			*j = 0
		} else {
			__antithesis_instrumentation__.Notify(18774)
			if n, err := strconv.Atoi(s); err == nil {
				__antithesis_instrumentation__.Notify(18775)
				*j = jsonMaxRetries(n)
			} else {
				__antithesis_instrumentation__.Notify(18776)
				return errors.Errorf("max retries must be either a positive int or 'inf' for infinite retries.")
			}
		}
	}
	__antithesis_instrumentation__.Notify(18764)
	return nil
}

type retryConfig struct {
	Max     jsonMaxRetries `json:",omitempty"`
	Backoff jsonDuration   `json:",omitempty"`
}

type webhookSinkConfig struct {
	Flush batchConfig `json:",omitempty"`
	Retry retryConfig `json:",omitempty"`
}

func (s *webhookSink) getWebhookSinkConfig(
	opts map[string]string,
) (batchCfg batchConfig, retryCfg retry.Options, err error) {
	__antithesis_instrumentation__.Notify(18777)
	retryCfg = defaultRetryConfig()

	var cfg webhookSinkConfig
	cfg.Retry.Max = jsonMaxRetries(retryCfg.MaxRetries)
	cfg.Retry.Backoff = jsonDuration(retryCfg.InitialBackoff)
	if configStr, ok := opts[changefeedbase.OptWebhookSinkConfig]; ok {
		__antithesis_instrumentation__.Notify(18781)

		if err = json.Unmarshal([]byte(configStr), &cfg); err != nil {
			__antithesis_instrumentation__.Notify(18782)
			return batchCfg, retryCfg, errors.Wrapf(err, "error unmarshalling json")
		} else {
			__antithesis_instrumentation__.Notify(18783)
		}
	} else {
		__antithesis_instrumentation__.Notify(18784)
	}
	__antithesis_instrumentation__.Notify(18778)

	if cfg.Flush.Messages < 0 || func() bool {
		__antithesis_instrumentation__.Notify(18785)
		return cfg.Flush.Bytes < 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(18786)
		return cfg.Flush.Frequency < 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(18787)
		return cfg.Retry.Max < 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(18788)
		return cfg.Retry.Backoff < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(18789)
		return batchCfg, retryCfg, errors.Errorf("invalid option value %s, all config values must be non-negative", changefeedbase.OptWebhookSinkConfig)
	} else {
		__antithesis_instrumentation__.Notify(18790)
	}
	__antithesis_instrumentation__.Notify(18779)

	if (cfg.Flush.Messages > 0 || func() bool {
		__antithesis_instrumentation__.Notify(18791)
		return cfg.Flush.Bytes > 0 == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(18792)
		return cfg.Flush.Frequency == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(18793)
		return batchCfg, retryCfg, errors.Errorf("invalid option value %s, flush frequency is not set, messages may never be sent", changefeedbase.OptWebhookSinkConfig)
	} else {
		__antithesis_instrumentation__.Notify(18794)
	}
	__antithesis_instrumentation__.Notify(18780)

	retryCfg.MaxRetries = int(cfg.Retry.Max)
	retryCfg.InitialBackoff = time.Duration(cfg.Retry.Backoff)
	return cfg.Flush, retryCfg, nil
}

func makeWebhookSink(
	ctx context.Context,
	u sinkURL,
	opts map[string]string,
	parallelism int,
	source timeutil.TimeSource,
	m *sliMetrics,
) (Sink, error) {
	__antithesis_instrumentation__.Notify(18795)
	if u.Scheme != changefeedbase.SinkSchemeWebhookHTTPS {
		__antithesis_instrumentation__.Notify(18805)
		return nil, errors.Errorf(`this sink requires %s`, changefeedbase.SinkSchemeHTTPS)
	} else {
		__antithesis_instrumentation__.Notify(18806)
	}
	__antithesis_instrumentation__.Notify(18796)
	u.Scheme = strings.TrimPrefix(u.Scheme, `webhook-`)

	switch changefeedbase.FormatType(opts[changefeedbase.OptFormat]) {
	case changefeedbase.OptFormatJSON:
		__antithesis_instrumentation__.Notify(18807)

	default:
		__antithesis_instrumentation__.Notify(18808)
		return nil, errors.Errorf(`this sink is incompatible with %s=%s`,
			changefeedbase.OptFormat, opts[changefeedbase.OptFormat])
	}
	__antithesis_instrumentation__.Notify(18797)

	switch changefeedbase.EnvelopeType(opts[changefeedbase.OptEnvelope]) {
	case changefeedbase.OptEnvelopeWrapped:
		__antithesis_instrumentation__.Notify(18809)
	default:
		__antithesis_instrumentation__.Notify(18810)
		return nil, errors.Errorf(`this sink is incompatible with %s=%s`,
			changefeedbase.OptEnvelope, opts[changefeedbase.OptEnvelope])
	}
	__antithesis_instrumentation__.Notify(18798)

	if _, ok := opts[changefeedbase.OptKeyInValue]; !ok {
		__antithesis_instrumentation__.Notify(18811)
		return nil, errors.Errorf(`this sink requires the WITH %s option`, changefeedbase.OptKeyInValue)
	} else {
		__antithesis_instrumentation__.Notify(18812)
	}
	__antithesis_instrumentation__.Notify(18799)

	if _, ok := opts[changefeedbase.OptTopicInValue]; !ok {
		__antithesis_instrumentation__.Notify(18813)
		return nil, errors.Errorf(`this sink requires the WITH %s option`, changefeedbase.OptTopicInValue)
	} else {
		__antithesis_instrumentation__.Notify(18814)
	}
	__antithesis_instrumentation__.Notify(18800)

	var connTimeout time.Duration
	if timeout, ok := opts[changefeedbase.OptWebhookClientTimeout]; ok {
		__antithesis_instrumentation__.Notify(18815)
		var err error
		connTimeout, err = time.ParseDuration(timeout)
		if err != nil {
			__antithesis_instrumentation__.Notify(18816)
			return nil, errors.Wrapf(err, "problem parsing option %s", changefeedbase.OptWebhookClientTimeout)
		} else {
			__antithesis_instrumentation__.Notify(18817)
			if connTimeout <= time.Duration(0) {
				__antithesis_instrumentation__.Notify(18818)
				return nil, fmt.Errorf("option %s must be a positive duration", changefeedbase.OptWebhookClientTimeout)
			} else {
				__antithesis_instrumentation__.Notify(18819)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(18820)
		connTimeout = defaultConnTimeout
	}
	__antithesis_instrumentation__.Notify(18801)

	ctx, cancel := context.WithCancel(ctx)

	sink := &webhookSink{
		workerCtx:   ctx,
		authHeader:  opts[changefeedbase.OptWebhookAuthHeader],
		exitWorkers: cancel,
		parallelism: parallelism,
		ts:          source,
		metrics:     m,
	}

	var err error
	sink.batchCfg, sink.retryCfg, err = sink.getWebhookSinkConfig(opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(18821)
		return nil, errors.Wrapf(err, "error processing option %s", changefeedbase.OptWebhookSinkConfig)
	} else {
		__antithesis_instrumentation__.Notify(18822)
	}
	__antithesis_instrumentation__.Notify(18802)

	sink.client, err = makeWebhookClient(u, connTimeout)
	if err != nil {
		__antithesis_instrumentation__.Notify(18823)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18824)
	}
	__antithesis_instrumentation__.Notify(18803)

	sinkURLParsed, err := url.Parse(u.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(18825)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18826)
	}
	__antithesis_instrumentation__.Notify(18804)
	params := sinkURLParsed.Query()
	params.Del(changefeedbase.SinkParamSkipTLSVerify)
	params.Del(changefeedbase.SinkParamCACert)
	params.Del(changefeedbase.SinkParamClientCert)
	params.Del(changefeedbase.SinkParamClientKey)
	sinkURLParsed.RawQuery = params.Encode()
	sink.url = sinkURL{URL: sinkURLParsed}

	return sink, nil
}

func makeWebhookClient(u sinkURL, timeout time.Duration) (*httputil.Client, error) {
	__antithesis_instrumentation__.Notify(18827)
	client := &httputil.Client{
		Client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{Timeout: timeout}).DialContext,
			},
		},
	}

	dialConfig := struct {
		tlsSkipVerify bool
		caCert        []byte
		clientCert    []byte
		clientKey     []byte
	}{}

	transport := client.Transport.(*http.Transport)

	if _, err := u.consumeBool(changefeedbase.SinkParamSkipTLSVerify, &dialConfig.tlsSkipVerify); err != nil {
		__antithesis_instrumentation__.Notify(18835)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18836)
	}
	__antithesis_instrumentation__.Notify(18828)
	if err := u.decodeBase64(changefeedbase.SinkParamCACert, &dialConfig.caCert); err != nil {
		__antithesis_instrumentation__.Notify(18837)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18838)
	}
	__antithesis_instrumentation__.Notify(18829)
	if err := u.decodeBase64(changefeedbase.SinkParamClientCert, &dialConfig.clientCert); err != nil {
		__antithesis_instrumentation__.Notify(18839)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18840)
	}
	__antithesis_instrumentation__.Notify(18830)
	if err := u.decodeBase64(changefeedbase.SinkParamClientKey, &dialConfig.clientKey); err != nil {
		__antithesis_instrumentation__.Notify(18841)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18842)
	}
	__antithesis_instrumentation__.Notify(18831)

	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: dialConfig.tlsSkipVerify,
	}

	if dialConfig.caCert != nil {
		__antithesis_instrumentation__.Notify(18843)
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			__antithesis_instrumentation__.Notify(18847)
			return nil, errors.Wrap(err, "could not load system root CA pool")
		} else {
			__antithesis_instrumentation__.Notify(18848)
		}
		__antithesis_instrumentation__.Notify(18844)
		if caCertPool == nil {
			__antithesis_instrumentation__.Notify(18849)
			caCertPool = x509.NewCertPool()
		} else {
			__antithesis_instrumentation__.Notify(18850)
		}
		__antithesis_instrumentation__.Notify(18845)
		if !caCertPool.AppendCertsFromPEM(dialConfig.caCert) {
			__antithesis_instrumentation__.Notify(18851)
			return nil, errors.Errorf("failed to parse certificate data:%s", string(dialConfig.caCert))
		} else {
			__antithesis_instrumentation__.Notify(18852)
		}
		__antithesis_instrumentation__.Notify(18846)
		transport.TLSClientConfig.RootCAs = caCertPool
	} else {
		__antithesis_instrumentation__.Notify(18853)
	}
	__antithesis_instrumentation__.Notify(18832)

	if dialConfig.clientCert != nil && func() bool {
		__antithesis_instrumentation__.Notify(18854)
		return dialConfig.clientKey == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(18855)
		return nil, errors.Errorf(`%s requires %s to be set`, changefeedbase.SinkParamClientCert, changefeedbase.SinkParamClientKey)
	} else {
		__antithesis_instrumentation__.Notify(18856)
		if dialConfig.clientKey != nil && func() bool {
			__antithesis_instrumentation__.Notify(18857)
			return dialConfig.clientCert == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(18858)
			return nil, errors.Errorf(`%s requires %s to be set`, changefeedbase.SinkParamClientKey, changefeedbase.SinkParamClientCert)
		} else {
			__antithesis_instrumentation__.Notify(18859)
		}
	}
	__antithesis_instrumentation__.Notify(18833)

	if dialConfig.clientCert != nil && func() bool {
		__antithesis_instrumentation__.Notify(18860)
		return dialConfig.clientKey != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(18861)
		cert, err := tls.X509KeyPair(dialConfig.clientCert, dialConfig.clientKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(18863)
			return nil, errors.Wrap(err, `invalid client certificate data provided`)
		} else {
			__antithesis_instrumentation__.Notify(18864)
		}
		__antithesis_instrumentation__.Notify(18862)
		transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
	} else {
		__antithesis_instrumentation__.Notify(18865)
	}
	__antithesis_instrumentation__.Notify(18834)

	return client, nil
}

func defaultRetryConfig() retry.Options {
	__antithesis_instrumentation__.Notify(18866)
	opts := retry.Options{
		InitialBackoff: 500 * time.Millisecond,
		MaxRetries:     3,
		Multiplier:     2,
	}

	opts.MaxBackoff = opts.InitialBackoff * time.Duration(int(math.Pow(2.0, float64(opts.MaxRetries))))
	return opts
}

func defaultWorkerCount() int {
	__antithesis_instrumentation__.Notify(18867)
	return system.NumCPU()
}

func (s *webhookSink) Dial() error {
	__antithesis_instrumentation__.Notify(18868)
	s.setupWorkers()
	return nil
}

func (s *webhookSink) setupWorkers() {
	__antithesis_instrumentation__.Notify(18869)

	s.eventsChans = make([]chan []messagePayload, s.parallelism)
	s.workerGroup = ctxgroup.WithContext(s.workerCtx)
	s.batchChan = make(chan webhookMessage)

	s.errChan = make(chan error, 1)

	s.flushDone = make(chan struct{})

	s.workerGroup.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(18871)
		s.batchWorker()
		return nil
	})
	__antithesis_instrumentation__.Notify(18870)
	for i := 0; i < s.parallelism; i++ {
		__antithesis_instrumentation__.Notify(18872)
		s.eventsChans[i] = make(chan []messagePayload)
		j := i
		s.workerGroup.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(18873)
			s.workerLoop(j)
			return nil
		})
	}
}

func (s *webhookSink) shouldSendBatch(b batch) bool {
	__antithesis_instrumentation__.Notify(18874)

	switch {

	case s.batchCfg.Messages == 0 && func() bool {
		__antithesis_instrumentation__.Notify(18879)
		return s.batchCfg.Bytes == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(18880)
		return s.batchCfg.Frequency == 0 == true
	}() == true:
		__antithesis_instrumentation__.Notify(18875)
		return true

	case s.batchCfg.Messages > 0 && func() bool {
		__antithesis_instrumentation__.Notify(18881)
		return len(b.buffer) >= s.batchCfg.Messages == true
	}() == true:
		__antithesis_instrumentation__.Notify(18876)
		return true

	case s.batchCfg.Bytes > 0 && func() bool {
		__antithesis_instrumentation__.Notify(18882)
		return b.bufferBytes >= s.batchCfg.Bytes == true
	}() == true:
		__antithesis_instrumentation__.Notify(18877)
		return true
	default:
		__antithesis_instrumentation__.Notify(18878)
		return false
	}
}

func (s *webhookSink) splitAndSendBatch(batch []messagePayload) error {
	__antithesis_instrumentation__.Notify(18883)
	workerBatches := make([][]messagePayload, s.parallelism)
	for _, msg := range batch {
		__antithesis_instrumentation__.Notify(18886)

		i := s.workerIndex(msg.key)
		workerBatches[i] = append(workerBatches[i], msg)
	}
	__antithesis_instrumentation__.Notify(18884)
	for i, workerBatch := range workerBatches {
		__antithesis_instrumentation__.Notify(18887)

		if len(workerBatch) > 0 {
			__antithesis_instrumentation__.Notify(18888)
			select {
			case <-s.workerCtx.Done():
				__antithesis_instrumentation__.Notify(18889)
				return s.workerCtx.Err()
			case s.eventsChans[i] <- workerBatch:
				__antithesis_instrumentation__.Notify(18890)
			}
		} else {
			__antithesis_instrumentation__.Notify(18891)
		}
	}
	__antithesis_instrumentation__.Notify(18885)
	return nil
}

func (s *webhookSink) flushWorkers(done chan struct{}) error {
	__antithesis_instrumentation__.Notify(18892)
	for i := 0; i < len(s.eventsChans); i++ {
		__antithesis_instrumentation__.Notify(18894)

		select {
		case <-s.workerCtx.Done():
			__antithesis_instrumentation__.Notify(18895)
			return s.workerCtx.Err()
		case s.eventsChans[i] <- nil:
			__antithesis_instrumentation__.Notify(18896)
		}
	}
	__antithesis_instrumentation__.Notify(18893)

	select {
	case <-s.workerCtx.Done():
		__antithesis_instrumentation__.Notify(18897)
		return s.workerCtx.Err()
	case done <- struct{}{}:
		__antithesis_instrumentation__.Notify(18898)
		return nil
	}
}

func (s *webhookSink) batchWorker() {
	__antithesis_instrumentation__.Notify(18899)
	var batchTracker batch
	batchTimer := s.ts.NewTimer()
	defer batchTimer.Stop()

	for {
		__antithesis_instrumentation__.Notify(18900)
		select {
		case <-s.workerCtx.Done():
			__antithesis_instrumentation__.Notify(18901)
			return
		case msg := <-s.batchChan:
			__antithesis_instrumentation__.Notify(18902)
			flushRequested := msg.flushDone != nil

			if !flushRequested {
				__antithesis_instrumentation__.Notify(18905)
				batchTracker.addToBuffer(msg.payload)
			} else {
				__antithesis_instrumentation__.Notify(18906)
			}
			__antithesis_instrumentation__.Notify(18903)

			if s.shouldSendBatch(batchTracker) || func() bool {
				__antithesis_instrumentation__.Notify(18907)
				return flushRequested == true
			}() == true {
				__antithesis_instrumentation__.Notify(18908)
				if err := s.splitAndSendBatch(batchTracker.buffer); err != nil {
					__antithesis_instrumentation__.Notify(18910)
					s.exitWorkersWithError(err)
					return
				} else {
					__antithesis_instrumentation__.Notify(18911)
				}
				__antithesis_instrumentation__.Notify(18909)
				batchTracker.reset()

				if flushRequested {
					__antithesis_instrumentation__.Notify(18912)
					if err := s.flushWorkers(*msg.flushDone); err != nil {
						__antithesis_instrumentation__.Notify(18913)
						s.exitWorkersWithError(err)
						return
					} else {
						__antithesis_instrumentation__.Notify(18914)
					}
				} else {
					__antithesis_instrumentation__.Notify(18915)
				}
			} else {
				__antithesis_instrumentation__.Notify(18916)
				if len(batchTracker.buffer) == 1 && func() bool {
					__antithesis_instrumentation__.Notify(18917)
					return time.Duration(s.batchCfg.Frequency) > 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(18918)

					batchTimer.Reset(time.Duration(s.batchCfg.Frequency))
				} else {
					__antithesis_instrumentation__.Notify(18919)
				}
			}

		case <-batchTimer.Ch():
			__antithesis_instrumentation__.Notify(18904)
			batchTimer.MarkRead()
			if len(batchTracker.buffer) > 0 {
				__antithesis_instrumentation__.Notify(18920)
				if err := s.splitAndSendBatch(batchTracker.buffer); err != nil {
					__antithesis_instrumentation__.Notify(18922)
					s.exitWorkersWithError(err)
					return
				} else {
					__antithesis_instrumentation__.Notify(18923)
				}
				__antithesis_instrumentation__.Notify(18921)
				batchTracker.reset()
			} else {
				__antithesis_instrumentation__.Notify(18924)
			}
		}
	}
}

func (s *webhookSink) workerLoop(workerIndex int) {
	__antithesis_instrumentation__.Notify(18925)
	for {
		__antithesis_instrumentation__.Notify(18926)
		select {
		case <-s.workerCtx.Done():
			__antithesis_instrumentation__.Notify(18927)
			return
		case msgs := <-s.eventsChans[workerIndex]:
			__antithesis_instrumentation__.Notify(18928)
			if msgs == nil {
				__antithesis_instrumentation__.Notify(18932)

				continue
			} else {
				__antithesis_instrumentation__.Notify(18933)
			}
			__antithesis_instrumentation__.Notify(18929)

			encoded, err := encodePayloadWebhook(msgs)
			if err != nil {
				__antithesis_instrumentation__.Notify(18934)
				s.exitWorkersWithError(err)
				return
			} else {
				__antithesis_instrumentation__.Notify(18935)
			}
			__antithesis_instrumentation__.Notify(18930)
			if err := s.sendMessageWithRetries(s.workerCtx, encoded.data); err != nil {
				__antithesis_instrumentation__.Notify(18936)
				s.exitWorkersWithError(err)
				return
			} else {
				__antithesis_instrumentation__.Notify(18937)
			}
			__antithesis_instrumentation__.Notify(18931)
			encoded.alloc.Release(s.workerCtx)
			s.metrics.recordEmittedBatch(
				encoded.emitTime, len(msgs), encoded.mvcc, len(encoded.data), sinkDoesNotCompress)
		}
	}
}

func (s *webhookSink) sendMessageWithRetries(ctx context.Context, reqBody []byte) error {
	__antithesis_instrumentation__.Notify(18938)
	requestFunc := func() error {
		__antithesis_instrumentation__.Notify(18940)
		return s.sendMessage(ctx, reqBody)
	}
	__antithesis_instrumentation__.Notify(18939)
	return retry.WithMaxAttempts(ctx, s.retryCfg, s.retryCfg.MaxRetries+1, requestFunc)
}

func (s *webhookSink) sendMessage(ctx context.Context, reqBody []byte) error {
	__antithesis_instrumentation__.Notify(18941)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url.String(), bytes.NewReader(reqBody))
	if err != nil {
		__antithesis_instrumentation__.Notify(18946)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18947)
	}
	__antithesis_instrumentation__.Notify(18942)
	req.Header.Set("Content-Type", applicationTypeJSON)
	if s.authHeader != "" {
		__antithesis_instrumentation__.Notify(18948)
		req.Header.Set(authorizationHeader, s.authHeader)
	} else {
		__antithesis_instrumentation__.Notify(18949)
	}
	__antithesis_instrumentation__.Notify(18943)

	var res *http.Response
	res, err = s.client.Do(req)
	if err != nil {
		__antithesis_instrumentation__.Notify(18950)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18951)
	}
	__antithesis_instrumentation__.Notify(18944)
	defer res.Body.Close()

	if !(res.StatusCode >= http.StatusOK && func() bool {
		__antithesis_instrumentation__.Notify(18952)
		return res.StatusCode < http.StatusMultipleChoices == true
	}() == true) {
		__antithesis_instrumentation__.Notify(18953)
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			__antithesis_instrumentation__.Notify(18955)
			return errors.Wrapf(err, "failed to read body for HTTP response with status: %d", res.StatusCode)
		} else {
			__antithesis_instrumentation__.Notify(18956)
		}
		__antithesis_instrumentation__.Notify(18954)
		return fmt.Errorf("%s: %s", res.Status, string(resBody))
	} else {
		__antithesis_instrumentation__.Notify(18957)
	}
	__antithesis_instrumentation__.Notify(18945)
	return nil
}

func (s *webhookSink) workerIndex(key []byte) uint32 {
	__antithesis_instrumentation__.Notify(18958)
	return crc32.ChecksumIEEE(key) % uint32(s.parallelism)
}

func (s *webhookSink) exitWorkersWithError(err error) {
	__antithesis_instrumentation__.Notify(18959)

	select {
	case s.errChan <- err:
		__antithesis_instrumentation__.Notify(18960)
		s.exitWorkers()
	default:
		__antithesis_instrumentation__.Notify(18961)
	}
}

func (s *webhookSink) sinkError() error {
	__antithesis_instrumentation__.Notify(18962)
	select {
	case err := <-s.errChan:
		__antithesis_instrumentation__.Notify(18963)
		return err
	default:
		__antithesis_instrumentation__.Notify(18964)
		return nil
	}
}

func (s *webhookSink) EmitRow(
	ctx context.Context,
	topic TopicDescriptor,
	key, value []byte,
	updated, mvcc hlc.Timestamp,
	alloc kvevent.Alloc,
) error {
	__antithesis_instrumentation__.Notify(18965)
	select {

	case <-s.workerCtx.Done():
		__antithesis_instrumentation__.Notify(18967)

		return errors.CombineErrors(s.workerCtx.Err(), s.sinkError())
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(18968)
		return ctx.Err()
	case err := <-s.errChan:
		__antithesis_instrumentation__.Notify(18969)
		return err
	case s.batchChan <- webhookMessage{
		payload: messagePayload{
			key:      key,
			val:      value,
			alloc:    alloc,
			emitTime: timeutil.Now(),
			mvcc:     mvcc,
		}}:
		__antithesis_instrumentation__.Notify(18970)
		s.metrics.recordMessageSize(int64(len(key) + len(value)))
	}
	__antithesis_instrumentation__.Notify(18966)
	return nil
}

func (s *webhookSink) EmitResolvedTimestamp(
	ctx context.Context, encoder Encoder, resolved hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(18971)
	defer s.metrics.recordResolvedCallback()()

	payload, err := encoder.EncodeResolvedTimestamp(ctx, "", resolved)
	if err != nil {
		__antithesis_instrumentation__.Notify(18975)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18976)
	}
	__antithesis_instrumentation__.Notify(18972)

	select {

	case <-s.workerCtx.Done():
		__antithesis_instrumentation__.Notify(18977)
		return s.workerCtx.Err()

	case <-s.errChan:
		__antithesis_instrumentation__.Notify(18978)
		return err
	default:
		__antithesis_instrumentation__.Notify(18979)
	}
	__antithesis_instrumentation__.Notify(18973)

	if err := s.sendMessageWithRetries(ctx, payload); err != nil {
		__antithesis_instrumentation__.Notify(18980)
		s.exitWorkersWithError(err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18981)
	}
	__antithesis_instrumentation__.Notify(18974)

	return nil
}

func (s *webhookSink) Flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(18982)
	s.metrics.recordFlushRequestCallback()()

	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(18984)
		return ctx.Err()
	case err := <-s.errChan:
		__antithesis_instrumentation__.Notify(18985)
		return err
	case s.batchChan <- webhookMessage{flushDone: &s.flushDone}:
		__antithesis_instrumentation__.Notify(18986)
	}
	__antithesis_instrumentation__.Notify(18983)

	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(18987)
		return ctx.Err()
	case err := <-s.errChan:
		__antithesis_instrumentation__.Notify(18988)
		return err
	case <-s.flushDone:
		__antithesis_instrumentation__.Notify(18989)
		return s.sinkError()
	}
}

func (s *webhookSink) Close() error {
	__antithesis_instrumentation__.Notify(18990)
	s.exitWorkers()

	_ = s.workerGroup.Wait()
	close(s.batchChan)
	close(s.errChan)
	for _, eventsChan := range s.eventsChans {
		__antithesis_instrumentation__.Notify(18992)
		close(eventsChan)
	}
	__antithesis_instrumentation__.Notify(18991)
	s.client.CloseIdleConnections()
	return nil
}

func redactWebhookAuthHeader(_ string) string {
	__antithesis_instrumentation__.Notify(18993)
	return "redacted"
}
